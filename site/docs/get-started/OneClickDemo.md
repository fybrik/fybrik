{% set currentRelease = fybrik_version(FybrikRelease) %}

# One Click Demo

This guide contains a script that you can run with a single command, to see a demo of Fybrik in action.  

The demo demonstrates the following sequence:

1.  A kind Kubernetes cluster is installed and Fybrik and all its dependencies are deployed to the cluster.  
You can also try to do this step by step in the [QuickStart](./quickstart.md) segment.

2.  A **data operator** registers a data asset in a data catalog used by Fybrik, and tags it as financial data.  
You can also try to do this, and the next parts of the demo, step by step in the [notebook-read-sample](../samples/notebook-read.md) segment.


3.  A **data governance officer** defines data policies, such as which columns contain PII (personally identifiable information), and submits them to Fybrik.

4.  A **data user** submits a request to read the data asset using Fybrik.

5.  **Fybrik** fetches the data asset, and automatically redacts columns according to the data policies.

6.  The **data user** can consume the governed data instantly.

## Requirements

1. Linux based OS or MacOS with a working bash terminal.
2. Docker [installed and deployed](https://docs.docker.com/get-docker/). 

The demo will make a bin folder at your current directory with all the other required dependencies for Fybrik.

## The Demo
We recommend trying Fybrik with its main data catalog [OpenMetaData](https://open-metadata.org/), but it takes ~25 minutes on most machines, so grab a coffee while you wait!  
Alternatively you can try Fybrik with Katalog,  a data catalog stub, strictly used for demos, to see a demo that takes ~5 minutes.

=== "Demo with OpenMetaData" 
    ```bash
    
    export FYBRIK_VERSION={{ currentRelease|default('master') }}; curl https://raw.githubusercontent.com/fybrik/fybrik/{{ currentRelease|default('master') }}/samples/OneClickDemo/OneClickDemo-OMD.sh | bash -
    ```

=== "Demo with Katalog" 
    ```bash
    export FYBRIK_VERSION={{ currentRelease|default('master') }}; curl https://raw.githubusercontent.com/fybrik/fybrik/{{ currentRelease|default('master') }}/samples/OneClickDemo/OneClickDemo-Katalog.sh | bash -
    ```

> **NOTE**: At the end of the demo, you will see in your terminal a sample from a table that the data user consumed. one of the columns will display XXXXX instead of values, indicating that it has been automatically redacted due to data policies.

## Cleanup

To stop the local kind kubernetes cluster booted up on your machine in this demo, and to remove the folder created with the dependencies for Fybrik, run this.  

```bash
bin/kind delete cluster --name=kind-fybrik-installation-sample
rm -rf bin 
rm res.out
```

> **WARNING**: If you already had a bin folder at your current directory these commands will delete it and its contents.
