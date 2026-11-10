---
title: "Calling an authenticated API without handing credentials to the model"
date: 2026-11-10T00:00:00+03:00
weight: 120
draft: true
description: Build a read-only API tool with application-owned credentials, fixed endpoints, and resource authorization.
toc: true
readingTime: true
---

A tool can read a local file or fetch a record from an API. The second version needs credentials, but that does not mean the model needs to see them.

This follow-up to the local tool-calling article will keep authentication in application code and examine who may access the data behind a successful HTTP response.

*Working draft. The API example and tests are planned.*

## The example to build

Start with a read-only lookup against a controlled example API. The model supplies a record identifier. The application owns the base URL, endpoint, HTTP method, and authentication headers.

Use synthetic records and test credentials, never private production data.

## Questions to cover

- How should the application load credentials without putting them in prompts or traces?
- What can the service credential access, and what may the current user access?
- Which response fields should reach the model?
- How should the client handle redirects, timeouts, rate limits, and expired credentials?
- When would scoped user authorization be more appropriate than a shared API key?

## Experiments before publishing

Test forbidden records, an attempted endpoint override, missing credentials, and upstream errors. Check that logs and model-visible results contain neither credentials nor unnecessary response data.

Keep writes and irreversible operations outside this first API example.
