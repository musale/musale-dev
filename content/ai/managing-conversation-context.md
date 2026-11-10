---
title: "What happens when the conversation gets too long?"
date: 2026-11-10T00:00:00+03:00
weight: 70
draft: true
description: Explore conversation budgets, trimming, and summarization, and measure what information each approach loses.
toc: true
readingTime: true
---

The [Go harness]({{< relref "ai/prompting-models-with-go.md" >}}) stops when its history budget is exhausted. That is honest, but it does not help someone who needs to continue a conversation.

What would we lose if we removed old messages? Could a summary keep the important details, or would it quietly change the task?

*Working draft. The context-management experiments have not been run.*

## Build a conversation we can inspect

Create a synthetic conversation with an early constraint, a later correction, and tool results needed by a final question. Keep an unmodified copy so we can compare what each strategy retains.

## Compare three approaches

- Stop with an explicit budget error, as the existing example does.
- Keep recent turns while preserving application instructions and complete tool-call exchanges.
- Summarize older turns, keeping the summary separate from authoritative instructions.

Distinguish a local byte limit from model tokens and the actual context window. Account for tool definitions and reserved output space rather than measuring user-visible text alone.

## Check what survives

Ask about the early constraint and the corrected fact after each reduction. Inspect whether a summary invents details or promotes untrusted tool content into instructions.

Record request size, available token counts, and latency. Treat summarization as another model operation that can fail and consume time.

## Before publishing

Verify how the selected runtime reports context use and handles oversized requests. Add a repeatable conversation fixture and tests that preserve message ordering and tool-call relationships.
