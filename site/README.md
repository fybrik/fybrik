# Fybrik Website

Holds source files for the project documentation and the website hosting it

## Contribute 

Read the [Contribution guidelines](https://fybrik.io/dev/contribute/documentation/)

## Requirements

- Make
- Python 3.x
- [Material for MkDocs](https://squidfunk.github.io/mkdocs-material/)
    ```bash
    # in some distros the command is pip3
    pip install mkdocs-material
    ```

## Usage

- Run `make generate` to generate documentation pages from the project APIs (protos and CRDs)
- Run `make run` and browse http://localhost:8000/ to preview the website locally
