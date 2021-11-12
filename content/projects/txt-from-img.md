---
title: Testing the Tessaract OCR Project
slug: txt-from-img
tags: [mpesa, asyncio, python]
date: 2019-08-15
description: "[txt-from-img](txt-from-img/) was a gentle introduction to the tessaract library and how to get text from an image in a lambda function."
---

## Motivation

This was a pet project to test out OCR using the Tessaract project. We're currently processing legal documents and some of them are in PDF or image formats. This necessitates an OCR solution which was a good place to start so that the process can be automated. Also, I was looking for a way to make lambda deployments easy and seamless. In the [project](https://github.com/musale/txt-from-img), most of the setup has been done thanks to inspiration from already existing solutions.

## Results

I was able to deploy a lambda function which you can invoke by sending an image which will return the text in the image. Currently, only text in English language can be processed. Other options are possible but are not covered. You can clone the project and test it out on your machine or who knows, get it on one of your pipelines! Enjoy OCR. :camera:
