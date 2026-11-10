---
title: "Exposing local notes tools through MCP"
date: 2026-11-10T00:00:00+03:00
weight: 130
draft: true
description: Put an MCP interface around existing notes tools without changing their permissions or approval requirements.
toc: true
readingTime: true
---

Our notes tools do not need MCP to work. A Go function can already validate a request, read a note, and return a result.

The next question is what changes when another application needs to discover and call those tools through a standard interface.

*Working draft. The protocol version, SDK, and example client still need selection and verification.*

## Keep behavior separate from transport

Expose the local notes tools through an MCP server while reusing their validation and permission rules. Compare a direct function call with a protocol request for the same operation.

## Questions to cover

- What do discovery, argument schemas, and tool results look like on the wire?
- Which responsibilities belong to the host, the server, and the underlying tool?
- Where should human approval happen when client and server are separate processes?
- How do cancellation and errors cross the boundary?
- Why does connection to an MCP server not grant unlimited access to its machine?

## Experiments before publishing

Run the same policy tests through the direct dispatcher and the protocol interface. Try unknown tools, disallowed paths, and an unavailable server.

Use one verified SDK and protocol revision throughout. Do not copy approval assumptions from a local CLI into a remote service without defining identity and authorization.
