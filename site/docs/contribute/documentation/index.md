# Contribute Documentation

The content of this website is the documentation of the project. The documentation is managed in `/site/docs` as markdown files. [MkDocs](https://www.mkdocs.org/) and the [Material](https://squidfunk.github.io/mkdocs-material/) theme are used to generate the website from these markdown files.

Reference pages are auto generated from the source code. Therefore, if you change Kubernetes Custom Resource Definitions or the connectors API then you must add reasonable documentation comments. The rest of the documentation pages are written manually.

Contributing to the documentation is therefore similar to code contribution and follows the same process of using pull requests. However, when writing documentation you must also follow the [formatting](./formatting) and [style](./style) guidelines.

Before opening a pull request, to preview the website locally, you should [install Mkdocs-Material](https://squidfunk.github.io/mkdocs-material/getting-started/) and follow a [few more steps](https://github.com/fybrik/fybrik/blob/master/site/README.md).