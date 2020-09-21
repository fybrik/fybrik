---
title: "{{ replace .Name "-" " " | title }}"
date: {{ .Date }}
draft: false
weight: 100
---

Focus on the installation steps needed to complete a setup of {{< name >}}. 

In the begning of the document describe the properties of the installation outcome. This allows the reader to quickly ensure that this is the installation he is looking for. For example, provide a figure showing the deployment model, list technical and business requirements such as licenses needed if a commercial component is used, clarify if the setup is suitable for a production environment, etc.

Provide instructions as a sequence of sets to follow. If you notice that you need to describe different flows then consider creating separate setup page instead.