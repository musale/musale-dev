---
title: "What changes when an assistant can run a command?"
date: 2026-11-10T00:00:00+03:00
weight: 140
draft: true
description: Examine restricted command execution, process isolation, and resource limits before giving an assistant access to executable tools.
toc: true
readingTime: true
---

A tool that reads one approved note has a small set of possible effects. A command runner can reach much further, especially if it inherits the application's filesystem access and credentials.

This article will examine that difference before adding command execution to a model application.

*Working draft. The isolated execution environment and examples need design and testing.*

## Begin with a narrow operation

Prefer one fixed executable and validated arguments over a general shell interface. Use a disposable workspace with synthetic files and no inherited secrets.

## Boundaries to explain

- Command allowlists and argument validation are different from process isolation.
- Shell interpretation introduces behavior that a direct executable invocation does not require.
- Child processes, filesystem access, and network access need explicit treatment.
- Output limits and cancellation must account for the process tree, not just the initiating request.
- Approval should describe the actual operation and cannot replace isolation.

## Experiments before publishing

Test a prohibited executable, unexpected arguments, excessive output, and a process that exceeds its runtime budget. Verify cleanup and what the process can access.

Choose a concrete isolation mechanism and state its limits. Do not present a timeout, working directory, or container label alone as proof that arbitrary execution is safe.
