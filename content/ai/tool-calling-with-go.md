---
title: "Tool calling without giving the model the keys"
date: 2026-11-10T00:00:00+03:00
weight: 40
draft: true
description: Extend a Go model harness with tool validation, explicit permissions, and bounded execution.
toc: true
readingTime: true
---

The model asks to read a file. That sounds simple until the file lives outside the workspace, the arguments are malformed, or the request would change something on disk.

In the [small Go harness]({{< relref "ai/prompting-models-with-go.md" >}}), we already restrict reads to one file and limit the number of calls. I want to make the boundary between a model's request and the application's permission to act easier to inspect.

*Working draft. The extensions and experiments below are planned, not implemented.*

## The example to build

Extend the local notes assistant with `list_notes`, a restricted `read_note`, and `propose_note`.

`propose_note` produces a proposed change; it does not save it. The application displays the destination and exact content for a person to approve. Only application code can save the approved proposal, within the permitted workspace. Refusing or canceling approval leaves the files unchanged.

Keep the tools local to the filesystem. The model still runs through the existing local HTTP client, but these tools do not call external APIs, execute shell commands, or use MCP.

## Follow one request through the application

- Inspect the tool description, the model's arguments, and the result returned to it.
- Separate decoding, input validation, authorization, and execution.
- Reject unknown tools and unexpected fields. Check paths in the application, not only in the advertised schema.
- Bind approval to the exact proposed operation. A changed proposal needs new approval.
- Recheck the destination and relevant file state before saving, and prevent an approval from being reused.
- Bound calls, elapsed time, and result size. Explain which operations actually honor cancellation.

## Requests to test

Try a permitted read, a missing note, a path outside the allowed directory, and arguments with the wrong types. Test the dispatcher directly with fixtures so a cooperative model cannot hide a missing check.

Also put misleading instructions in a note. Observe whether the model follows them, while checking separately that the application still enforces its permissions.

For writes, test approval, rejection, and cancellation. Change the proposal or destination file after review and verify that the application does not apply an outdated approval. Check that the model cannot approve its own proposal or save it through a different tool.

## Follow-up articles

These are separate topics, outside this article's local tool example:

- [An authenticated API tool]({{< relref "ai/authenticated-api-tools-with-go.md" >}}): start with a read-only endpoint. Keep credentials in the application, fix the allowed endpoint and method, and distinguish service authentication from the user's permission to access a resource.
- [Exposing the tools through MCP]({{< relref "ai/notes-tools-over-mcp.md" >}}): reuse the tool behavior and explain what the protocol changes. Connecting through MCP should not silently grant broader permissions.
- [Restricted command execution]({{< relref "ai/restricted-command-tools.md" >}}): examine isolation, restricted capabilities, and resource limits as a separate exercise. The local notes example does not need arbitrary command execution.
- [An email-to-GitHub collaboration assistant]({{< relref "ai/email-github-collaboration-assistant.md" >}}): read selected emails, inspect approved repositories, recommend matches privately, and prepare a response draft. Treat incoming email as untrusted input and enforce disclosure rules for private projects. Sending email or inviting collaborators would require separate authorization and approval.

## Before publishing

Add runnable examples and tests, record real traces, and describe the limits of filesystem confinement. Include the approval-based `propose_note` flow; leave authenticated APIs and the other follow-up topics for their own articles. Avoid presenting a prompt instruction or a timeout as a complete security boundary.
