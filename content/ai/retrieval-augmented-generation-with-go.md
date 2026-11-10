---
title: "RAG, built small enough to understand"
date: 2026-11-10T00:00:00+03:00
weight: 80
draft: true
description: Build retrieval over a small local document collection and trace generated answers back to their sources.
toc: true
readingTime: true
---

The model does not know what is in my local notes unless the application supplies it. Sending every note might work for a tiny folder, but it becomes awkward as the collection grows.

I want to retrieve a small amount of relevant text and see whether the answer actually depends on it.

*Working draft. The retrieval example and comparisons are planned.*

## Start with a small collection

Use synthetic project notes with stable document identifiers. Write questions whose answers are present, questions that require more than one passage, and questions the collection cannot answer.

Build a keyword-search baseline before adding embeddings. Keep retrieval and answer generation separate so each can be inspected.

## Follow the data

- Split documents into passages and preserve source identifiers.
- Rank candidate passages and inspect what reaches the prompt.
- Compare keyword search with embedding retrieval on the same questions.
- Ask for citations tied to supplied passages, and check that each cited passage exists.
- Check whether the cited text supports the claim, not just whether a citation was formatted correctly.

## Make retrieval fail visibly

Try ambiguous wording, a stale document, and conflicting passages. Separate a missed retrieval from a model that ignores good evidence.

Treat retrieved text as untrusted input. Retrieval is not an authorization system; a real multi-user application must filter access before supplying documents to a model.

## Before publishing

Choose and verify an embedding runtime, if the baseline justifies adding one. Publish the synthetic collection and expected source passages. Explain the limits of the example without claiming that citations eliminate unsupported answers.
