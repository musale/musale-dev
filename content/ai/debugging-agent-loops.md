---
title: "Debugging an agent that gets stuck"
date: 2026-11-10T00:00:00+03:00
weight: 100
draft: true
description: Trace repeated agent actions, reproduce failures, and distinguish useful retries from loops that need to stop.
toc: true
readingTime: true
---

The assistant reads the same note again. Then it retries a tool that already failed. The process is active, but the task is not getting closer to done.

I want enough evidence to explain a stuck run without guessing from its final answer.

*Working draft. The stuck-run fixtures and trace extensions are planned.*

## Start with an observable loop

Extend the [Go harness's JSONL trace]({{< relref "ai/prompting-models-with-go.md" >}}) with run identifiers, iteration numbers, elapsed time, tool outcomes, and an explicit stopping reason.

Use synthetic notes and controlled tool failures. Do not publish private prompts or raw production traces.

## Reproduce different failures

- Repeat an identical tool request without new information.
- Retry a persistent permission failure as though it were a transient outage.
- Keep asking for a tool result that exceeds the application's output limit.
- Continue planning after the task already has enough information for an answer.

Use recorded model-response fixtures to test the loop's control flow. Live model runs can then show how often those paths occur under the chosen conditions.

## Decide when to stop

Compare request budgets, deadlines, and repeated-action detection. Repetition can be legitimate, so a detector needs context and an explicit policy rather than an unexplained heuristic.

Report whether a run completed, failed, was canceled, or exhausted its budget. Returning the last partial answer should not disguise an incomplete task as success.

## Before publishing

Implement redaction and bounded trace output. Include one annotated failing run, a deterministic reproduction, and evidence that the fix addresses the failure rather than merely changing the prompt.
