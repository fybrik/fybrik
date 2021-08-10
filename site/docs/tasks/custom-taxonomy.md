# Using a Custom Taxonomy for Resource Validation

The project defines a set of immutable structural JSON schemas, or "taxonomies" for resources deployed in Fybrik. 
However, since the taxonomy is meant to be configurable, a `taxonomy.json` file is referenced from these schemas for any definition that is customizable.

The `taxonomy.json` file is generated from a base taxonomy and zero or more taxonomy layers:

- The base taxonomy is maintained by the project and includes all of the structural definitions that are subject to customization (e.g.: tags, actions). 

- The taxonomy layers are maintained by users and external systems that add customizations over the base taxonomy (e.g., defining specific tags, actions).


This task describes how to deploy Fybrik with a custom `taxonomy.json` file that is generated with the Taxonomy Compile CLI tool. 

## Taxonomy Compile CLI tool 
A CLI tool for compiling a base taxonomy and zero or more taxonomy layers is provided in our repo. 
The following command can be used from the root directory of our repo to run this tool. 

Usage:
```bash
  go run main.go taxonomy compile --out <outputFile> --base <baseFile> [<layerFile> ...] [--codegen]
```

Flags:

- -b, --base string : File with base taxonomy definitions (required)

- --codegen : Best effort to make output suitable for code generation tools

- -o, --out string : Path for output file (default "taxonomy.json")

This will generate a `taxonomy.json` file with the layers specified. 

## Deploy Fybrik with Custom Taxonomy

To deploy Fybrik with the generated `taxonomy.json` file, follow the [`quickstart guide`](https://fybrik.io/v0.4/get-started/quickstart/) but use the command below instead of `helm install fybrik fybrik-charts/fybrik -n fybrik-system --wait`:

```bash
helm install fybrik fybrik-charts/fybrik -n fybrik-system --wait --set-file taxonomyOverride=taxonomy.json
```
The `--set-file` flag will pass in your custom `taxonomy.json` file to use for taxonomy validation in Fybrik.
If this flag is not provided, Fybrik will use the default `taxonomy.json` file with no layers compiled into it. 
