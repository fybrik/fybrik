---
title: "Roadmap"
date: 2020-08-02T14:46:05+03:00
draft: false
weight: 1
---


The project is in very early stages. 

While we haven't set an official roadmap, some planned key features are:
1. Multicluster support
1. Secure comminication between control plane components (using Istio)
1. Secure secret management: ensure secrets are only observed by services that are authorized to view them
1. Isolation: ensure that no service is communication with an unauthorized party
1. Modules for write-path: validate that exported data is in compliance with policies
1. Istio-based blueprint: A typical blueprint to be supported in the near future would be composed of an Istio gateway to police egress traffic from the services running in the blurprint. Istio routing resources applied to the gateway, to expose virtual endpoints for datasets. WASM filters applied to the gateway, to support credential injection and possibly transformations over data.

Look for GitHub issues with the `design` tag for concrete plans (once available).