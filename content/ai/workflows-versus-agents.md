---
title: "When a workflow is better than an agent"
date: 2026-11-10T00:00:00+03:00
weight: 110
draft: true
description: Implement one task as ordinary code, a fixed model workflow, and an agent, then compare their behavior and complexity.
toc: true
readingTime: true
---

If I already know the steps a program needs to take, what do I gain by letting a model choose them?

The [agents article]({{< relref "ai/ai-agents-and-protocols.md" >}}) separates fixed workflows from model-directed action. I want to put that distinction into code and compare the results on the same task.

*Working draft. The three implementations and comparison results do not exist yet.*

## Choose a bounded task

Use synthetic build logs to produce a triage report with the failed step, supporting lines, and a suggested next action. Include straightforward compiler errors and cases that require looking up a related local note.

Define success before choosing an architecture.

## Build three versions

- Ordinary Go code parses known log patterns and assembles a report.
- A fixed workflow gathers predetermined evidence, asks a model to interpret it, and validates the output.
- An agent chooses among the same permitted inspection tools within a fixed budget.

Give each version access to the same underlying evidence. Record any differences in prompt construction or tool interfaces rather than hiding them.

## Compare what matters

Measure task success and unsupported claims as well as elapsed time and model usage. Describe implementation and debugging effort separately from runtime measurements.

Try an unfamiliar error, a missing note, and an ambiguous log. Examine where predictable code is sufficient and where choosing the next step helps.

## Before publishing

Create the fixtures and all three implementations. Publish failures and measurement conditions. Avoid claiming that either workflows or agents win universally from a small local experiment.
