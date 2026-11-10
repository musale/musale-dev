---
title: "How do you test something that answers differently each time?"
date: 2026-11-10T00:00:00+03:00
weight: 50
draft: true
description: Build a small evaluation set for a Go model application and distinguish correctness checks from judgments about answer quality.
toc: true
readingTime: true
---

A test that compares an entire model answer with one expected string is easy to write. It can also reject a perfectly useful answer because the wording changed.

I want to test the behavior I need without pretending that every possible answer has a single reference sentence.

*Working draft. The evaluation runner and results still need to be built.*

## Start with the notes assistant

Use the [Go harness]({{< relref "ai/prompting-models-with-go.md" >}}) and a small set of synthetic notes. Include questions it can answer, questions that need a tool, and questions the notes cannot support.

Each case should describe the task, permitted actions, and observable success conditions.

## Separate the checks

- Unit-test parsing and tool policy without a live model.
- Check task properties: the requested fact appears, the source exists, and the application stayed within its call budget.
- Review unsupported claims and usefulness with a written rubric.
- Keep any model-based judge separate from deterministic checks, and inspect its disagreements with human review.

## Compare changes fairly

Run the same cases against the baseline and a changed prompt or model. Record the model identifier, settings, and relevant environment details. Repeat runs to expose variation rather than selecting the nicest answer.

Report counts and denominators alongside latency. A small evaluation set can help catch regressions; it cannot establish that an application works for every user.

## Before publishing

Create the fixtures and runner, define the rubric before measuring, and include failures as well as successes. Keep synthetic evaluation data separate from private conversation traces.
