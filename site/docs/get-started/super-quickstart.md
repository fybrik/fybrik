# Super Quick Start Guide

this guide contains a script that you can run with a single command to see a demo of Fybrik in action.  

this demo demonstartes the following sequence -  

1. A **data operator** boots up Fybrik and registers a data asset of financial data to Fybrik

2. A **data governance officer** defines data policies, such as which columns contain PII (personally identifiable information), and submits them to Fybrik

3. A **data user** submits a request to read the data asset using Fybrik

4. **Fybrik** fetchs the data in an optimal way for cost and performance, and automatically redacts columns according to the data plicies

5. The **data user** can consume the governed data instantly

All your require for this demo is a working bash terminal.  
The demo will make a bin folder at your current directory with all the required dependencies for Fybrik.

## Demo with OpenMetaData data catlog (~25 mintues)
For a demo with the current version of Fybrik using its main data catlog OpenMetaData, you can run the following command.  
Note: booting up Fybrik with openmetadata takes ~25 minutes on most machines, so grab a coffee while you wait!  
Alternatively you can run the [next](#demo-with-katalog-a-data-catalog-stub-5-mintues) script which boots up Fybrik with a data catalog stub.  

```bash
// use doron's method for running the script
```

## Demo with Katalog, a data catalog stub (~5 mintues)
For a demo using an older version of Fybrik, without its main Data Catlog you can use this script, which runs using Katalog, a data catalog stub.  
this is stirclty for demos, testing and evaluation purposes.

```bash
// use doron's method for running the script
```

## Cleanup

To stop local kubernetes cluster booted up on your machine for this demo, and to remove the folder created with the dependencies for Fybrik, run this.  
warning - if you already had a bin folder at your current directory these commands will delete it and its contents

```bash
bin/kind delete clusters kind-fybrik-installation-sample
rm -rf bin 
```