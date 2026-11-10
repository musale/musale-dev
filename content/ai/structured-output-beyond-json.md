---
title: "Structured output: valid JSON isn't enough"
date: 2026-11-10T00:00:00+03:00
weight: 60
draft: true
description: Compare JSON prompting with schema-constrained output, then validate the values before an application uses them.
toc: true
readingTime: true
---

An answer can be valid JSON and still be useless. A priority might contain an unsupported value. A required field might be empty. A well-formed identifier might refer to a record that does not exist.

I want to follow a model response through the same checks I would expect for any other external input.

*Working draft. Runtime support and example behavior need verification before publication.*

## The example to build

Turn a short, synthetic incident report into a triage record with a summary, a priority, and supporting evidence. Include an explicit way to represent insufficient information.

Compare asking for JSON in the prompt with the runtime's supported structured-output mechanism. Keep the requested task and examples the same.

## Follow the validation steps

- Distinguish parseable JSON, schema conformance, and business-rule validity.
- Decode into Go types and handle unknown fields and missing values deliberately.
- Check allowed values and constraints that the schema or runtime does not enforce.
- Explain why constrained output does not establish that the facts are correct.
- Handle refusals, incomplete responses, and validation failures as distinct outcomes.

## Failures to exercise

Use fixtures with an unknown priority, an empty summary, and evidence absent from the source report. Interrupt a response and examine how the client distinguishes partial data from a completed record.

If retries help, give them a fixed budget and preserve the original failure. Do not silently turn an invalid response into an accepted default.

## Before publishing

Verify the chosen runtime's supported schema features. Implement the validator, collect real responses, and document which guarantees come from the runtime and which come from application code.
