---
title: "{{ replace .Name "-" " " | title }}"
date: {{ .Date }}
draft: false
weight: 100
---

Component documents describe the internals of a significant {{< name >}} component. Their target audiance are readers with high level of understanding of the project that want to deepen their understanding on a specific aspect, consider contributing to the project, or wish to understand the status and roadmap of a specific component.

Component documents should give some context about the component:
1. Link to relevant architecture sections
1. Link or provide an overview of the requirements addressed by the component
1. Link or provide an overview of the functionality provided by the component

Component documents should provide technical details:
1. Link to relevant code
1. Provide an overview of the component's internals
1. Describe the status of the component

Component documents should provide a roadmap:
1. Aspects that need significant imporovement
1. Unimplemented features that are required for the component to be consider complete
1. Unimplemented features that the community accepted as planned features for the component 
1. Rough timelines
1. Emphasis on things that require signigicant contributions from the community 


