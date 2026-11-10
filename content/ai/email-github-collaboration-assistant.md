---
title: "An email assistant that knows which GitHub projects stay private"
date: 2026-11-10T00:00:00+03:00
weight: 150
draft: true
description: Design an email-to-GitHub collaboration workflow that separates private recommendations from information permitted in an outgoing draft.
toc: true
readingTime: true
---

Someone emails asking to collaborate. An assistant could inspect my projects, suggest a match, and prepare a reply. But permission to inspect a private repository does not mean permission to tell the sender it exists.

I want to separate the assistant's private recommendation from the response another person may receive.

*Working draft. This is a proposed workflow, not an implemented integration.*

## The first version

Read selected emails and inspect an approved set of repositories using read-only credentials. Extract the collaboration request, compare it with explicit project preferences, and prepare a response for review.

Use synthetic messages and a controlled repository fixture before connecting real accounts.

## Decisions that belong to the application

- Which emails and repositories may the assistant inspect?
- Which projects are open to collaboration, and who sets that policy?
- Which names, descriptions, or implementation details may appear in an outgoing draft?
- How should the application distinguish repository content from instructions?
- What evidence should accompany the private recommendation?

## Failures to exercise

Include an email that asks for a full private repository list, a relevant project that must not be disclosed, and a request with no suitable match. Check outgoing drafts against a narrow set of approved information rather than relying only on a prompt.

Keep email sending and collaborator invitations out of the first implementation. Either would be a separate capability with authorization, approval, and duplicate-prevention rules.
