import json
import shutil
import subprocess
import tempfile
import unittest
from html.parser import HTMLParser
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]


class FollowUpLinks(HTMLParser):
    def __init__(self):
        super().__init__()
        self.links = []
        self.text = []

    def handle_starttag(self, tag, attrs):
        if tag == "a":
            self.links.extend(value for key, value in attrs if key == "href")

    def handle_data(self, data):
        self.text.append(data)


class FollowUpTests(unittest.TestCase):
    def setUp(self):
        self.directory = tempfile.TemporaryDirectory()
        self.addCleanup(self.directory.cleanup)
        self.site = Path(self.directory.name)
        (self.site / "content/ai").mkdir(parents=True)
        (self.site / "layouts/shortcodes").mkdir(parents=True)
        (self.site / "layouts/_default").mkdir()
        (self.site / "hugo.toml").write_text(
            'baseURL = "https://example.test/"\n'
            'disableKinds = ["taxonomy", "term", "rss", "sitemap"]\n'
        )
        (self.site / "layouts/_default/single.html").write_text("{{ .Content }}")
        shutil.copyfile(
            ROOT / "layouts/shortcodes/follow-ups.html",
            self.site / "layouts/shortcodes/follow-ups.html",
        )

    def page(self, name, body="", **metadata):
        values = {"title": name, "date": "2020-01-01T00:00:00Z", **metadata}
        source = "---\n"
        source += "\n".join(f"{key}: {json.dumps(value)}" for key, value in values.items())
        source += "\n---\n\n" + body + "\n"
        path = self.site / "content/ai" / f"{name}.md"
        path.write_text(source)
        return path

    def parent(self, references):
        return self.page("parent", body="{{< follow-ups >}}", followUps=references)

    def unpublished(self):
        self.page("draft", draft=True)
        self.page("future", date="2027-01-01T00:00:00Z")
        self.page("scheduled", publishDate="2027-01-01T00:00:00Z")
        self.page("expired", expiryDate="2021-01-01T00:00:00Z")
        return ["ai/draft", "ai/future", "ai/scheduled", "ai/expired"]

    def build(self, preview=False):
        command = [
            "hugo",
            "--source",
            str(self.site),
            "--clock",
            "2026-11-10T12:00:00Z",
            "--quiet",
        ]
        if preview:
            command.extend(["--buildDrafts", "--buildFuture", "--buildExpired"])
        result = subprocess.run(command, capture_output=True, text=True)
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        output = (self.site / "public/ai/parent/index.html").read_text()
        parsed = FollowUpLinks()
        parsed.feed(output)
        return output, parsed

    def test_hides_empty_or_unpublished_followups_even_in_previews(self):
        references = self.unpublished()
        for selected in [[], references]:
            self.parent(selected)
            for preview in [False, True]:
                with self.subTest(references=selected, preview=preview):
                    output, parsed = self.build(preview)
                    self.assertNotIn("follow-up-articles", output)
                    self.assertEqual(parsed.links, [])
                    for reference in references:
                        self.assertNotIn(reference.removeprefix("ai/"), output)

    def test_renders_published_titles_and_descriptions_in_declared_order(self):
        hidden = self.unpublished()
        self.page(
            "second",
            title="Second & <careful>",
            description="A response with <bounded> data.",
        )
        self.page("first", title="First article")
        self.parent(["ai/second", *hidden, "ai/first"])
        for preview in [False, True]:
            with self.subTest(preview=preview):
                output, parsed = self.build(preview)
                self.assertIn('<h2 id="follow-up-articles">', output)
                self.assertEqual(parsed.links, ["/ai/second/", "/ai/first/"])
                text = "".join(parsed.text)
                self.assertIn(
                    "Second & <careful>: A response with <bounded> data.", text
                )
                self.assertIn("First article", text)
                self.assertNotIn("<careful>", output)
                self.assertNotIn("<bounded>", output)

    def test_publishing_a_followup_updates_the_unchanged_parent(self):
        parent = self.parent(["ai/next"])
        source = parent.read_text()
        self.page("next", draft=True)
        _, before = self.build()
        self.assertEqual(before.links, [])
        self.page("next", draft=False, title="The next article")
        _, after = self.build()
        self.assertEqual(after.links, ["/ai/next/"])
        self.assertIn("The next article", "".join(after.text))
        self.assertEqual(parent.read_text(), source)


if __name__ == "__main__":
    unittest.main()
