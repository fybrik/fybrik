---
title: Preview website locally 
description: Explains how to locally build, test, serve, and preview the website.
weight: 10
keywords: [contribute, serve, Docker, Hugo, build]
---

{{< warning >}}
This website is only tested with Hugo extended version 0.70.0
{{</ warning >}}


# Requirements

- Install [Hugo extended](https://gohugo.io/getting-started/installing/)
- Install [asciidoctor](https://asciidoctor.org/docs/install-toolchain/) 
- both are available in homebrew (`brew install hugo asciidoctor`)

# Usage

- Run `make gen-docs` to generate documentation pa  ges from the project APIs (protos and CRDs)
- Run `make server` and browse http://localhost:1313/ to preview the website locally

The website is refreshed as you make content changes.
