---
title: "Streaming model responses in Go"
date: 2026-11-10T00:00:00+03:00
weight: 90
draft: true
description: Handle partial model responses, cancellation, broken connections, and slow consumers in a Go streaming client.
toc: true
readingTime: true
---

Seeing the first part of an answer arrive feels different from waiting for the whole response. It also gives the application more states to handle: an unfinished sentence, a canceled request, or a connection that ends before completion.

The [existing Go example]({{< relref "ai/prompting-models-with-go.md" >}}) already streams text. I want to examine its behavior when the connection or consumer stops cooperating.

*Working draft. The fault-injection examples and measurements still need implementation.*

## Keep the wire format visible

Explain the selected endpoint's streaming format and distinguish application events from HTTP reads. A network read does not necessarily contain one complete JSON object, token, or printable character.

Keep this article about text streaming. The existing example deliberately does not combine streaming with tool calls.

## Build a controllable test server

Use Go's HTTP test tools to split events across writes, delay output, send malformed data, and close without a completion event.

Test cancellation before the first event and midway through an answer. Slow down the output writer to examine backpressure without introducing an unbounded queue.

## Measure the wait

Record time to the first user-visible text and total completion time. Distinguish warm and cold runs, and explain why visible improvement does not necessarily mean faster total generation.

Decide how partial answers appear after failure. Retrying after showing text may repeat output; restarting is an application decision, not a transparent repair.

## Before publishing

Verify endpoint completion semantics, add deterministic transport tests, and record actual timings without presenting one machine's results as a general benchmark.
