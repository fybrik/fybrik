---
hide:
  - toc        # Hide table of contents
---

# API Reference

Packages:

- [app.fybrik.io/v1alpha1](#appfybrikiov1alpha1)
- [katalog.fybrik.io/v1alpha1](#katalogfybrikiov1alpha1)
- [motion.fybrik.io/v1alpha1](#motionfybrikiov1alpha1)

## app.fybrik.io/v1alpha1

Resource Types:

- [Blueprint](#blueprint)

- [FybrikApplication](#fybrikapplication)

- [FybrikModule](#fybrikmodule)

- [FybrikStorageAccount](#fybrikstorageaccount)

- [Plotter](#plotter)




### Blueprint
<sup><sup>[↩ Parent](#appfybrikiov1alpha1 )</sup></sup>






Blueprint is the Schema for the blueprints API

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>app.fybrik.io/v1alpha1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>Blueprint</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#blueprintspec">spec</a></b></td>
        <td>object</td>
        <td>
          BlueprintSpec defines the desired state of Blueprint, which is the runtime environment which provides the Data Scientist's application with secure and governed access to the data requested in the FybrikApplication. The blueprint uses an "argo like" syntax which indicates the components and the flow of data between them as steps TODO: Add an indication of the communication relationships between the components<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#blueprintstatus">status</a></b></td>
        <td>object</td>
        <td>
          BlueprintStatus defines the observed state of Blueprint This includes readiness, error message, and indicators forthe Kubernetes resources owned by the Blueprint for cleanup and status monitoring<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### Blueprint.spec
<sup><sup>[↩ Parent](#blueprint)</sup></sup>



BlueprintSpec defines the desired state of Blueprint, which is the runtime environment which provides the Data Scientist's application with secure and governed access to the data requested in the FybrikApplication. The blueprint uses an "argo like" syntax which indicates the components and the flow of data between them as steps TODO: Add an indication of the communication relationships between the components

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>entrypoint</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#blueprintspecflow">flow</a></b></td>
        <td>object</td>
        <td>
          DataFlow indicates the flow of the data between the components Currently we assume this is linear and thus use steps, but other more complex graphs could be defined as per how it is done in argo workflow<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#blueprintspectemplatesindex">templates</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Blueprint.spec.flow
<sup><sup>[↩ Parent](#blueprintspec)</sup></sup>



DataFlow indicates the flow of the data between the components Currently we assume this is linear and thus use steps, but other more complex graphs could be defined as per how it is done in argo workflow

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#blueprintspecflowstepsindex">steps</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Blueprint.spec.flow.steps[index]
<sup><sup>[↩ Parent](#blueprintspecflow)</sup></sup>



FlowStep is one step indicates an instance of a module in the blueprint, It includes the name of the module template (spec) and the parameters received by the component instance that is initiated by the orchestrator.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#blueprintspecflowstepsindexarguments">arguments</a></b></td>
        <td>object</td>
        <td>
          Arguments are the input parameters for a specific instance of a module.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is the name of the instance of the module. For example, if the application is named "notebook" and an implicitcopy module is deemed necessary.  The FlowStep name would be notebook-implicitcopy.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>template</b></td>
        <td>string</td>
        <td>
          Template is the name of the specification in the Blueprint describing how to instantiate a component indicated by the module.  It is the name of a FybrikModule CRD. For example: implicit-copy-db2wh-to-s3-latest<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Blueprint.spec.flow.steps[index].arguments
<sup><sup>[↩ Parent](#blueprintspecflowstepsindex)</sup></sup>



Arguments are the input parameters for a specific instance of a module.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#blueprintspecflowstepsindexargumentscopy">copy</a></b></td>
        <td>object</td>
        <td>
          CopyArgs are parameters specific to modules that copy data from one data store to another.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#blueprintspecflowstepsindexargumentsreadindex">read</a></b></td>
        <td>[]object</td>
        <td>
          ReadArgs are parameters that are specific to modules that enable an application to read data<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#blueprintspecflowstepsindexargumentswriteindex">write</a></b></td>
        <td>[]object</td>
        <td>
          WriteArgs are parameters that are specific to modules that enable an application to write data<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### Blueprint.spec.flow.steps[index].arguments.copy
<sup><sup>[↩ Parent](#blueprintspecflowstepsindexarguments)</sup></sup>



CopyArgs are parameters specific to modules that copy data from one data store to another.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>transformations</b></td>
        <td>[]object</td>
        <td>
          Transformations are different types of processing that may be done to the data as it is copied.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#blueprintspecflowstepsindexargumentscopydestination">destination</a></b></td>
        <td>object</td>
        <td>
          Destination is the data store to which the data will be copied<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#blueprintspecflowstepsindexargumentscopysource">source</a></b></td>
        <td>object</td>
        <td>
          Source is the where the data currently resides<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Blueprint.spec.flow.steps[index].arguments.copy.destination
<sup><sup>[↩ Parent](#blueprintspecflowstepsindexargumentscopy)</sup></sup>



Destination is the data store to which the data will be copied

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>connection</b></td>
        <td>object</td>
        <td>
          Connection has the relevant details for accesing the data (url, table, ssl, etc.)<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>format</b></td>
        <td>string</td>
        <td>
          Format represents data format (e.g. parquet) as received from catalog connectors<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#blueprintspecflowstepsindexargumentscopydestinationvault">vault</a></b></td>
        <td>object</td>
        <td>
          Holds details for retrieving credentials by the modules from Vault store.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Blueprint.spec.flow.steps[index].arguments.copy.destination.vault
<sup><sup>[↩ Parent](#blueprintspecflowstepsindexargumentscopydestination)</sup></sup>



Holds details for retrieving credentials by the modules from Vault store.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>address</b></td>
        <td>string</td>
        <td>
          Address is Vault address<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>authPath</b></td>
        <td>string</td>
        <td>
          AuthPath is the path to auth method i.e. kubernetes<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>role</b></td>
        <td>string</td>
        <td>
          Role is the Vault role used for retrieving the credentials<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>secretPath</b></td>
        <td>string</td>
        <td>
          SecretPath is the path of the secret holding the Credentials in Vault<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Blueprint.spec.flow.steps[index].arguments.copy.source
<sup><sup>[↩ Parent](#blueprintspecflowstepsindexargumentscopy)</sup></sup>



Source is the where the data currently resides

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>connection</b></td>
        <td>object</td>
        <td>
          Connection has the relevant details for accesing the data (url, table, ssl, etc.)<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>format</b></td>
        <td>string</td>
        <td>
          Format represents data format (e.g. parquet) as received from catalog connectors<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#blueprintspecflowstepsindexargumentscopysourcevault">vault</a></b></td>
        <td>object</td>
        <td>
          Holds details for retrieving credentials by the modules from Vault store.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Blueprint.spec.flow.steps[index].arguments.copy.source.vault
<sup><sup>[↩ Parent](#blueprintspecflowstepsindexargumentscopysource)</sup></sup>



Holds details for retrieving credentials by the modules from Vault store.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>address</b></td>
        <td>string</td>
        <td>
          Address is Vault address<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>authPath</b></td>
        <td>string</td>
        <td>
          AuthPath is the path to auth method i.e. kubernetes<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>role</b></td>
        <td>string</td>
        <td>
          Role is the Vault role used for retrieving the credentials<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>secretPath</b></td>
        <td>string</td>
        <td>
          SecretPath is the path of the secret holding the Credentials in Vault<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Blueprint.spec.flow.steps[index].arguments.read[index]
<sup><sup>[↩ Parent](#blueprintspecflowstepsindexarguments)</sup></sup>



ReadModuleArgs define the input parameters for modules that read data from location A

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>transformations</b></td>
        <td>[]object</td>
        <td>
          Transformations are different types of processing that may be done to the data<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>assetID</b></td>
        <td>string</td>
        <td>
          AssetID identifies the asset to be used for accessing the data when it is ready It is copied from the FybrikApplication resource<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#blueprintspecflowstepsindexargumentsreadindexsource">source</a></b></td>
        <td>object</td>
        <td>
          Source of the read path module<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Blueprint.spec.flow.steps[index].arguments.read[index].source
<sup><sup>[↩ Parent](#blueprintspecflowstepsindexargumentsreadindex)</sup></sup>



Source of the read path module

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>connection</b></td>
        <td>object</td>
        <td>
          Connection has the relevant details for accesing the data (url, table, ssl, etc.)<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>format</b></td>
        <td>string</td>
        <td>
          Format represents data format (e.g. parquet) as received from catalog connectors<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#blueprintspecflowstepsindexargumentsreadindexsourcevault">vault</a></b></td>
        <td>object</td>
        <td>
          Holds details for retrieving credentials by the modules from Vault store.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Blueprint.spec.flow.steps[index].arguments.read[index].source.vault
<sup><sup>[↩ Parent](#blueprintspecflowstepsindexargumentsreadindexsource)</sup></sup>



Holds details for retrieving credentials by the modules from Vault store.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>address</b></td>
        <td>string</td>
        <td>
          Address is Vault address<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>authPath</b></td>
        <td>string</td>
        <td>
          AuthPath is the path to auth method i.e. kubernetes<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>role</b></td>
        <td>string</td>
        <td>
          Role is the Vault role used for retrieving the credentials<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>secretPath</b></td>
        <td>string</td>
        <td>
          SecretPath is the path of the secret holding the Credentials in Vault<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Blueprint.spec.flow.steps[index].arguments.write[index]
<sup><sup>[↩ Parent](#blueprintspecflowstepsindexarguments)</sup></sup>



WriteModuleArgs define the input parameters for modules that write data to location B

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>transformations</b></td>
        <td>[]object</td>
        <td>
          Transformations are different types of processing that may be done to the data as it is written.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#blueprintspecflowstepsindexargumentswriteindexdestination">destination</a></b></td>
        <td>object</td>
        <td>
          Destination is the data store to which the data will be written<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Blueprint.spec.flow.steps[index].arguments.write[index].destination
<sup><sup>[↩ Parent](#blueprintspecflowstepsindexargumentswriteindex)</sup></sup>



Destination is the data store to which the data will be written

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>connection</b></td>
        <td>object</td>
        <td>
          Connection has the relevant details for accesing the data (url, table, ssl, etc.)<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>format</b></td>
        <td>string</td>
        <td>
          Format represents data format (e.g. parquet) as received from catalog connectors<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#blueprintspecflowstepsindexargumentswriteindexdestinationvault">vault</a></b></td>
        <td>object</td>
        <td>
          Holds details for retrieving credentials by the modules from Vault store.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Blueprint.spec.flow.steps[index].arguments.write[index].destination.vault
<sup><sup>[↩ Parent](#blueprintspecflowstepsindexargumentswriteindexdestination)</sup></sup>



Holds details for retrieving credentials by the modules from Vault store.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>address</b></td>
        <td>string</td>
        <td>
          Address is Vault address<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>authPath</b></td>
        <td>string</td>
        <td>
          AuthPath is the path to auth method i.e. kubernetes<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>role</b></td>
        <td>string</td>
        <td>
          Role is the Vault role used for retrieving the credentials<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>secretPath</b></td>
        <td>string</td>
        <td>
          SecretPath is the path of the secret holding the Credentials in Vault<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Blueprint.spec.templates[index]
<sup><sup>[↩ Parent](#blueprintspec)</sup></sup>



ComponentTemplate is a copy of a FybrikModule Custom Resource.  It contains the information necessary to instantiate a component in a FlowStep, which provides the functionality described by the module.  There are 3 different module types.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#blueprintspectemplatesindexchart">chart</a></b></td>
        <td>object</td>
        <td>
          Chart contains the location of the helm chart with info detailing how to deploy<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          Kind of k8s resource<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the template<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Blueprint.spec.templates[index].chart
<sup><sup>[↩ Parent](#blueprintspectemplatesindex)</sup></sup>



Chart contains the location of the helm chart with info detailing how to deploy

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>values</b></td>
        <td>map[string]string</td>
        <td>
          Values to pass to helm chart installation<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of helm chart<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Blueprint.status
<sup><sup>[↩ Parent](#blueprint)</sup></sup>



BlueprintStatus defines the observed state of Blueprint This includes readiness, error message, and indicators forthe Kubernetes resources owned by the Blueprint for cleanup and status monitoring

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          ObservedGeneration is taken from the Blueprint metadata.  This is used to determine during reconcile whether reconcile was called because the desired state changed, or whether status of the allocated resources should be checked.<br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#blueprintstatusobservedstate">observedState</a></b></td>
        <td>object</td>
        <td>
          ObservedState includes information to be reported back to the FybrikApplication resource It includes readiness and error indications, as well as user instructions<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>releases</b></td>
        <td>map[string]integer</td>
        <td>
          Releases map each release to the observed generation of the blueprint containing this release. At the end of reconcile, each release should be mapped to the latest blueprint version or be uninstalled.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### Blueprint.status.observedState
<sup><sup>[↩ Parent](#blueprintstatus)</sup></sup>



ObservedState includes information to be reported back to the FybrikApplication resource It includes readiness and error indications, as well as user instructions

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>dataAccessInstructions</b></td>
        <td>string</td>
        <td>
          DataAccessInstructions indicate how the data user or his application may access the data. Instructions are available upon successful orchestration.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>error</b></td>
        <td>string</td>
        <td>
          Error indicates that there has been an error to orchestrate the modules and provides the error message<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>ready</b></td>
        <td>boolean</td>
        <td>
          Ready represents that the modules have been orchestrated successfully and the data is ready for usage<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>

### FybrikApplication
<sup><sup>[↩ Parent](#appfybrikiov1alpha1 )</sup></sup>






FybrikApplication provides information about the application being used by a Data Scientist, the nature of the processing, and the data sets that the Data Scientist has chosen for processing by the application. The FybrikApplication controller (aka pilot) obtains instructions regarding any governance related changes that must be performed on the data, identifies the modules capable of performing such changes, and finally generates the Blueprint which defines the secure runtime environment and all the components in it.  This runtime environment provides the Data Scientist's application with access to the data requested in a secure manner and without having to provide any credentials for the data sets.  The credentials are obtained automatically by the manager from an external credential management system, which may or may not be part of a data catalog.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>app.fybrik.io/v1alpha1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>FybrikApplication</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#fybrikapplicationspec">spec</a></b></td>
        <td>object</td>
        <td>
          FybrikApplicationSpec defines the desired state of FybrikApplication.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#fybrikapplicationstatus">status</a></b></td>
        <td>object</td>
        <td>
          FybrikApplicationStatus defines the observed state of FybrikApplication.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### FybrikApplication.spec
<sup><sup>[↩ Parent](#fybrikapplication)</sup></sup>



FybrikApplicationSpec defines the desired state of FybrikApplication.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>secretRef</b></td>
        <td>string</td>
        <td>
          SecretRef points to the secret that holds credentials for each system the user has been authenticated with. The secret is deployed in FybrikApplication namespace.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#fybrikapplicationspecselector">selector</a></b></td>
        <td>object</td>
        <td>
          Selector enables to connect the resource to the application Application labels should match the labels in the selector. For some flows the selector may not be used.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>appInfo</b></td>
        <td>map[string]string</td>
        <td>
          AppInfo contains information describing the reasons for the processing that will be done by the Data Scientist's application.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#fybrikapplicationspecdataindex">data</a></b></td>
        <td>[]object</td>
        <td>
          Data contains the identifiers of the data to be used by the Data Scientist's application, and the protocol used to access it and the format expected.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### FybrikApplication.spec.selector
<sup><sup>[↩ Parent](#fybrikapplicationspec)</sup></sup>



Selector enables to connect the resource to the application Application labels should match the labels in the selector. For some flows the selector may not be used.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>clusterName</b></td>
        <td>string</td>
        <td>
          Cluster name<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#fybrikapplicationspecselectorworkloadselector">workloadSelector</a></b></td>
        <td>object</td>
        <td>
          WorkloadSelector enables to connect the resource to the application Application labels should match the labels in the selector.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### FybrikApplication.spec.selector.workloadSelector
<sup><sup>[↩ Parent](#fybrikapplicationspecselector)</sup></sup>



WorkloadSelector enables to connect the resource to the application Application labels should match the labels in the selector.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#fybrikapplicationspecselectorworkloadselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          matchExpressions is a list of label selector requirements. The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### FybrikApplication.spec.selector.workloadSelector.matchExpressions[index]
<sup><sup>[↩ Parent](#fybrikapplicationspecselectorworkloadselector)</sup></sup>



A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### FybrikApplication.spec.data[index]
<sup><sup>[↩ Parent](#fybrikapplicationspec)</sup></sup>



DataContext indicates data set chosen by the Data Scientist to be used by his application, and includes information about the data format and technologies used by the application to access the data.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>catalogService</b></td>
        <td>string</td>
        <td>
          CatalogService represents the catalog service for accessing the requested dataset. If not specified, the enterprise catalog service will be used.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>dataSetID</b></td>
        <td>string</td>
        <td>
          DataSetID is a unique identifier of the dataset chosen from the data catalog for processing by the data user application.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#fybrikapplicationspecdataindexrequirements">requirements</a></b></td>
        <td>object</td>
        <td>
          Requirements from the system<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### FybrikApplication.spec.data[index].requirements
<sup><sup>[↩ Parent](#fybrikapplicationspecdataindex)</sup></sup>



Requirements from the system

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#fybrikapplicationspecdataindexrequirementscopy">copy</a></b></td>
        <td>object</td>
        <td>
          CopyRequrements include the requirements for copying the data<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#fybrikapplicationspecdataindexrequirementsinterface">interface</a></b></td>
        <td>object</td>
        <td>
          Interface indicates the protocol and format expected by the data user<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### FybrikApplication.spec.data[index].requirements.copy
<sup><sup>[↩ Parent](#fybrikapplicationspecdataindexrequirements)</sup></sup>



CopyRequrements include the requirements for copying the data

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#fybrikapplicationspecdataindexrequirementscopycatalog">catalog</a></b></td>
        <td>object</td>
        <td>
          Catalog indicates that the data asset must be cataloged.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>required</b></td>
        <td>boolean</td>
        <td>
          Required indicates that the data must be copied.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### FybrikApplication.spec.data[index].requirements.copy.catalog
<sup><sup>[↩ Parent](#fybrikapplicationspecdataindexrequirementscopy)</sup></sup>



Catalog indicates that the data asset must be cataloged.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>catalogID</b></td>
        <td>string</td>
        <td>
          CatalogID specifies the catalog where the data will be cataloged.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>service</b></td>
        <td>string</td>
        <td>
          CatalogService specifies the datacatalog service that will be used for catalogging the data into.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### FybrikApplication.spec.data[index].requirements.interface
<sup><sup>[↩ Parent](#fybrikapplicationspecdataindexrequirements)</sup></sup>



Interface indicates the protocol and format expected by the data user

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>dataformat</b></td>
        <td>string</td>
        <td>
          DataFormat defines the data format type<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>protocol</b></td>
        <td>string</td>
        <td>
          Protocol defines the interface protocol used for data transactions<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### FybrikApplication.status
<sup><sup>[↩ Parent](#fybrikapplication)</sup></sup>



FybrikApplicationStatus defines the observed state of FybrikApplication.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>catalogedAssets</b></td>
        <td>map[string]string</td>
        <td>
          CatalogedAssets provide the new asset identifiers after being registered in the enterprise catalog It maps the original asset id to the cataloged asset id.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#fybrikapplicationstatusconditionsindex">conditions</a></b></td>
        <td>[]object</td>
        <td>
          Conditions represent the possible error and failure conditions<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>dataAccessInstructions</b></td>
        <td>string</td>
        <td>
          DataAccessInstructions indicate how the data user or his application may access the data. Instructions are available upon successful orchestration.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#fybrikapplicationstatusgenerated">generated</a></b></td>
        <td>object</td>
        <td>
          Generated resource identifier<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          ObservedGeneration is taken from the FybrikApplication metadata.  This is used to determine during reconcile whether reconcile was called because the desired state changed, or whether the Blueprint status changed.<br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#fybrikapplicationstatusprovisionedstoragekey">provisionedStorage</a></b></td>
        <td>map[string]object</td>
        <td>
          ProvisionedStorage maps a dataset (identified by AssetID) to the new provisioned bucket. It allows FybrikApplication controller to manage buckets in case the spec has been modified, an error has occurred, or a delete event has been received. ProvisionedStorage has the information required to register the dataset once the owned plotter resource is ready<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#fybrikapplicationstatusreadendpointsmapkey">readEndpointsMap</a></b></td>
        <td>map[string]object</td>
        <td>
          ReadEndpointsMap maps an datasetID (after parsing from json to a string with dashes) to the endpoint spec from which the asset will be served to the application<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>ready</b></td>
        <td>boolean</td>
        <td>
          Ready is true if a blueprint has been successfully orchestrated<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### FybrikApplication.status.conditions[index]
<sup><sup>[↩ Parent](#fybrikapplicationstatus)</sup></sup>



Condition describes the state of a FybrikApplication at a certain point.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          Message contains the details of the current condition<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>string</td>
        <td>
          Status of the condition: true or false<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          Type of the condition<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### FybrikApplication.status.generated
<sup><sup>[↩ Parent](#fybrikapplicationstatus)</sup></sup>



Generated resource identifier

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>appVersion</b></td>
        <td>integer</td>
        <td>
          Version of FybrikApplication that has generated this resource<br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          Kind of the resource (Blueprint, Plotter)<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the resource<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the resource<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### FybrikApplication.status.provisionedStorage[key]
<sup><sup>[↩ Parent](#fybrikapplicationstatus)</sup></sup>



DatasetDetails contain dataset connection and metadata required to register this dataset in the enterprise catalog

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>datasetRef</b></td>
        <td>string</td>
        <td>
          Reference to a Dataset resource containing the request to provision storage<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>details</b></td>
        <td>object</td>
        <td>
          Dataset information<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>secretRef</b></td>
        <td>string</td>
        <td>
          Reference to a secret where the credentials are stored<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### FybrikApplication.status.readEndpointsMap[key]
<sup><sup>[↩ Parent](#fybrikapplicationstatus)</sup></sup>



EndpointSpec is used both by the module creator and by the status of the fybrikapplication

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>hostname</b></td>
        <td>string</td>
        <td>
          Always equals the release name. Can be omitted.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>scheme</b></td>
        <td>string</td>
        <td>
          For example: http, https, grpc, grpc+tls, jdbc:oracle:thin:@ etc<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>

### FybrikModule
<sup><sup>[↩ Parent](#appfybrikiov1alpha1 )</sup></sup>






FybrikModule is a description of an injectable component. the parameters it requires, as well as the specification of how to instantiate such a component. It is used as metadata only.  There is no status nor reconciliation.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>app.fybrik.io/v1alpha1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>FybrikModule</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#fybrikmodulespec">spec</a></b></td>
        <td>object</td>
        <td>
          FybrikModuleSpec contains the info common to all modules, which are one of the components that process, load, write, audit, monitor the data used by the data scientist's application.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### FybrikModule.spec
<sup><sup>[↩ Parent](#fybrikmodule)</sup></sup>



FybrikModuleSpec contains the info common to all modules, which are one of the components that process, load, write, audit, monitor the data used by the data scientist's application.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#fybrikmodulespecdependenciesindex">dependencies</a></b></td>
        <td>[]object</td>
        <td>
          Other components that must be installed in order for this module to work<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#fybrikmodulespecstatusindicatorsindex">statusIndicators</a></b></td>
        <td>[]object</td>
        <td>
          StatusIndicators allow to check status of a non-standard resource that can not be computed by helm/kstatus<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#fybrikmodulespeccapabilities">capabilities</a></b></td>
        <td>object</td>
        <td>
          Capabilities declares what this module knows how to do and the types of data it knows how to handle<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#fybrikmodulespecchart">chart</a></b></td>
        <td>object</td>
        <td>
          Reference to a Helm chart that allows deployment of the resources required for this module<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>flows</b></td>
        <td>[]enum</td>
        <td>
          Flows is a list of the types of capabilities supported by the module - copy, read, write<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### FybrikModule.spec.dependencies[index]
<sup><sup>[↩ Parent](#fybrikmodulespec)</sup></sup>



Dependency details another component on which this module relies - i.e. a pre-requisit

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is the name of the dependent component<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type provides information used in determining how to instantiate the component<br/>
          <br/>
            <i>Enum</i>: module, connector, feature<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### FybrikModule.spec.statusIndicators[index]
<sup><sup>[↩ Parent](#fybrikmodulespec)</sup></sup>



ResourceStatusIndicator is used to determine the status of an orchestrated resource

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>errorMessage</b></td>
        <td>string</td>
        <td>
          ErrorMessage specifies the resource field to check for an error, e.g. status.errorMsg<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>failureCondition</b></td>
        <td>string</td>
        <td>
          FailureCondition specifies a condition that indicates the resource failure It uses kubernetes label selection syntax (https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/)<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          Kind provides information about the resource kind<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>successCondition</b></td>
        <td>string</td>
        <td>
          SuccessCondition specifies a condition that indicates that the resource is ready It uses kubernetes label selection syntax (https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/)<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### FybrikModule.spec.capabilities
<sup><sup>[↩ Parent](#fybrikmodulespec)</sup></sup>



Capabilities declares what this module knows how to do and the types of data it knows how to handle

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#fybrikmodulespeccapabilitiesactionsindex">actions</a></b></td>
        <td>[]object</td>
        <td>
          Actions are the data transformations that the module supports<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#fybrikmodulespeccapabilitiesapi">api</a></b></td>
        <td>object</td>
        <td>
          API indicates to the application how to access/write the data<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#fybrikmodulespeccapabilitiessupportedinterfacesindex">supportedInterfaces</a></b></td>
        <td>[]object</td>
        <td>
          Copy should have one or more instances in the list, and its content should have source and sink Read should have one or more instances in the list, each with source populated Write should have one or more instances in the list, each with sink populated TODO - In the future if we have a module type that doesn't interface directly with data then this list could be empty<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### FybrikModule.spec.capabilities.actions[index]
<sup><sup>[↩ Parent](#fybrikmodulespeccapabilities)</sup></sup>



SupportedAction declares an action that the module supports (action identifier and its scope)

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>id</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>level</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### FybrikModule.spec.capabilities.api
<sup><sup>[↩ Parent](#fybrikmodulespeccapabilities)</sup></sup>



API indicates to the application how to access/write the data

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>dataformat</b></td>
        <td>string</td>
        <td>
          DataFormat defines the data format type<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#fybrikmodulespeccapabilitiesapiendpoint">endpoint</a></b></td>
        <td>object</td>
        <td>
          EndpointSpec is used both by the module creator and by the status of the fybrikapplication<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>protocol</b></td>
        <td>string</td>
        <td>
          Protocol defines the interface protocol used for data transactions<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### FybrikModule.spec.capabilities.api.endpoint
<sup><sup>[↩ Parent](#fybrikmodulespeccapabilitiesapi)</sup></sup>



EndpointSpec is used both by the module creator and by the status of the fybrikapplication

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>hostname</b></td>
        <td>string</td>
        <td>
          Always equals the release name. Can be omitted.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>scheme</b></td>
        <td>string</td>
        <td>
          For example: http, https, grpc, grpc+tls, jdbc:oracle:thin:@ etc<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### FybrikModule.spec.capabilities.supportedInterfaces[index]
<sup><sup>[↩ Parent](#fybrikmodulespeccapabilities)</sup></sup>



ModuleInOut specifies the protocol and format of the data input and output by the module - if any

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#fybrikmodulespeccapabilitiessupportedinterfacesindexsink">sink</a></b></td>
        <td>object</td>
        <td>
          Sink specifies the output data protocol and format<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#fybrikmodulespeccapabilitiessupportedinterfacesindexsource">source</a></b></td>
        <td>object</td>
        <td>
          Source specifies the input data protocol and format<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>flow</b></td>
        <td>enum</td>
        <td>
          Flow for which this interface is supported<br/>
          <br/>
            <i>Enum</i>: copy, read, write<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### FybrikModule.spec.capabilities.supportedInterfaces[index].sink
<sup><sup>[↩ Parent](#fybrikmodulespeccapabilitiessupportedinterfacesindex)</sup></sup>



Sink specifies the output data protocol and format

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>dataformat</b></td>
        <td>string</td>
        <td>
          DataFormat defines the data format type<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>protocol</b></td>
        <td>string</td>
        <td>
          Protocol defines the interface protocol used for data transactions<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### FybrikModule.spec.capabilities.supportedInterfaces[index].source
<sup><sup>[↩ Parent](#fybrikmodulespeccapabilitiessupportedinterfacesindex)</sup></sup>



Source specifies the input data protocol and format

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>dataformat</b></td>
        <td>string</td>
        <td>
          DataFormat defines the data format type<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>protocol</b></td>
        <td>string</td>
        <td>
          Protocol defines the interface protocol used for data transactions<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### FybrikModule.spec.chart
<sup><sup>[↩ Parent](#fybrikmodulespec)</sup></sup>



Reference to a Helm chart that allows deployment of the resources required for this module

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>values</b></td>
        <td>map[string]string</td>
        <td>
          Values to pass to helm chart installation<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of helm chart<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>

### FybrikStorageAccount
<sup><sup>[↩ Parent](#appfybrikiov1alpha1 )</sup></sup>






FybrikStorageAccount defines a storage account used for copying data. Only S3 based storage is supported. It contains endpoint, region and a reference to the credentials a Owner of the asset is responsible to store the credentials

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>app.fybrik.io/v1alpha1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>FybrikStorageAccount</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#fybrikstorageaccountspec">spec</a></b></td>
        <td>object</td>
        <td>
          FybrikStorageAccountSpec defines the desired state of FybrikStorageAccount<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>object</td>
        <td>
          FybrikStorageAccountStatus defines the observed state of FybrikStorageAccount<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### FybrikStorageAccount.spec
<sup><sup>[↩ Parent](#fybrikstorageaccount)</sup></sup>



FybrikStorageAccountSpec defines the desired state of FybrikStorageAccount

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>endpoint</b></td>
        <td>string</td>
        <td>
          Endpoint<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>regions</b></td>
        <td>[]string</td>
        <td>
          Regions<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>secretRef</b></td>
        <td>string</td>
        <td>
          A name of k8s secret deployed in the control plane. This secret includes secretKey and accessKey credentials for S3 bucket<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>

### Plotter
<sup><sup>[↩ Parent](#appfybrikiov1alpha1 )</sup></sup>






Plotter is the Schema for the plotters API

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>app.fybrik.io/v1alpha1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>Plotter</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#plotterspec">spec</a></b></td>
        <td>object</td>
        <td>
          PlotterSpec defines the desired state of Plotter, which is applied in a multi-clustered environment. Plotter installs the runtime environment (as blueprints running on remote clusters) which provides the Data Scientist's application with secure and governed access to the data requested in the FybrikApplication.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#plotterstatus">status</a></b></td>
        <td>object</td>
        <td>
          PlotterStatus defines the observed state of Plotter This includes readiness, error message, and indicators received from blueprint resources owned by the Plotter for cleanup and status monitoring<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### Plotter.spec
<sup><sup>[↩ Parent](#plotter)</sup></sup>



PlotterSpec defines the desired state of Plotter, which is applied in a multi-clustered environment. Plotter installs the runtime environment (as blueprints running on remote clusters) which provides the Data Scientist's application with secure and governed access to the data requested in the FybrikApplication.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#plotterspecblueprintskey">blueprints</a></b></td>
        <td>map[string]object</td>
        <td>
          Blueprints structure represents remote blueprints mapped by the identifier of a cluster in which they will be running<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#plotterspecselector">selector</a></b></td>
        <td>object</td>
        <td>
          Selector enables to connect the resource to the application Should match the selector of the owner - FybrikApplication CRD.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Plotter.spec.blueprints[key]
<sup><sup>[↩ Parent](#plotterspec)</sup></sup>



BlueprintSpec defines the desired state of Blueprint, which is the runtime environment which provides the Data Scientist's application with secure and governed access to the data requested in the FybrikApplication. The blueprint uses an "argo like" syntax which indicates the components and the flow of data between them as steps TODO: Add an indication of the communication relationships between the components

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>entrypoint</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#plotterspecblueprintskeyflow">flow</a></b></td>
        <td>object</td>
        <td>
          DataFlow indicates the flow of the data between the components Currently we assume this is linear and thus use steps, but other more complex graphs could be defined as per how it is done in argo workflow<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#plotterspecblueprintskeytemplatesindex">templates</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Plotter.spec.blueprints[key].flow
<sup><sup>[↩ Parent](#plotterspecblueprintskey)</sup></sup>



DataFlow indicates the flow of the data between the components Currently we assume this is linear and thus use steps, but other more complex graphs could be defined as per how it is done in argo workflow

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#plotterspecblueprintskeyflowstepsindex">steps</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Plotter.spec.blueprints[key].flow.steps[index]
<sup><sup>[↩ Parent](#plotterspecblueprintskeyflow)</sup></sup>



FlowStep is one step indicates an instance of a module in the blueprint, It includes the name of the module template (spec) and the parameters received by the component instance that is initiated by the orchestrator.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#plotterspecblueprintskeyflowstepsindexarguments">arguments</a></b></td>
        <td>object</td>
        <td>
          Arguments are the input parameters for a specific instance of a module.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is the name of the instance of the module. For example, if the application is named "notebook" and an implicitcopy module is deemed necessary.  The FlowStep name would be notebook-implicitcopy.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>template</b></td>
        <td>string</td>
        <td>
          Template is the name of the specification in the Blueprint describing how to instantiate a component indicated by the module.  It is the name of a FybrikModule CRD. For example: implicit-copy-db2wh-to-s3-latest<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Plotter.spec.blueprints[key].flow.steps[index].arguments
<sup><sup>[↩ Parent](#plotterspecblueprintskeyflowstepsindex)</sup></sup>



Arguments are the input parameters for a specific instance of a module.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#plotterspecblueprintskeyflowstepsindexargumentscopy">copy</a></b></td>
        <td>object</td>
        <td>
          CopyArgs are parameters specific to modules that copy data from one data store to another.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#plotterspecblueprintskeyflowstepsindexargumentsreadindex">read</a></b></td>
        <td>[]object</td>
        <td>
          ReadArgs are parameters that are specific to modules that enable an application to read data<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#plotterspecblueprintskeyflowstepsindexargumentswriteindex">write</a></b></td>
        <td>[]object</td>
        <td>
          WriteArgs are parameters that are specific to modules that enable an application to write data<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### Plotter.spec.blueprints[key].flow.steps[index].arguments.copy
<sup><sup>[↩ Parent](#plotterspecblueprintskeyflowstepsindexarguments)</sup></sup>



CopyArgs are parameters specific to modules that copy data from one data store to another.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>transformations</b></td>
        <td>[]object</td>
        <td>
          Transformations are different types of processing that may be done to the data as it is copied.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#plotterspecblueprintskeyflowstepsindexargumentscopydestination">destination</a></b></td>
        <td>object</td>
        <td>
          Destination is the data store to which the data will be copied<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#plotterspecblueprintskeyflowstepsindexargumentscopysource">source</a></b></td>
        <td>object</td>
        <td>
          Source is the where the data currently resides<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Plotter.spec.blueprints[key].flow.steps[index].arguments.copy.destination
<sup><sup>[↩ Parent](#plotterspecblueprintskeyflowstepsindexargumentscopy)</sup></sup>



Destination is the data store to which the data will be copied

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>connection</b></td>
        <td>object</td>
        <td>
          Connection has the relevant details for accesing the data (url, table, ssl, etc.)<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>format</b></td>
        <td>string</td>
        <td>
          Format represents data format (e.g. parquet) as received from catalog connectors<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#plotterspecblueprintskeyflowstepsindexargumentscopydestinationvault">vault</a></b></td>
        <td>object</td>
        <td>
          Holds details for retrieving credentials by the modules from Vault store.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Plotter.spec.blueprints[key].flow.steps[index].arguments.copy.destination.vault
<sup><sup>[↩ Parent](#plotterspecblueprintskeyflowstepsindexargumentscopydestination)</sup></sup>



Holds details for retrieving credentials by the modules from Vault store.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>address</b></td>
        <td>string</td>
        <td>
          Address is Vault address<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>authPath</b></td>
        <td>string</td>
        <td>
          AuthPath is the path to auth method i.e. kubernetes<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>role</b></td>
        <td>string</td>
        <td>
          Role is the Vault role used for retrieving the credentials<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>secretPath</b></td>
        <td>string</td>
        <td>
          SecretPath is the path of the secret holding the Credentials in Vault<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Plotter.spec.blueprints[key].flow.steps[index].arguments.copy.source
<sup><sup>[↩ Parent](#plotterspecblueprintskeyflowstepsindexargumentscopy)</sup></sup>



Source is the where the data currently resides

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>connection</b></td>
        <td>object</td>
        <td>
          Connection has the relevant details for accesing the data (url, table, ssl, etc.)<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>format</b></td>
        <td>string</td>
        <td>
          Format represents data format (e.g. parquet) as received from catalog connectors<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#plotterspecblueprintskeyflowstepsindexargumentscopysourcevault">vault</a></b></td>
        <td>object</td>
        <td>
          Holds details for retrieving credentials by the modules from Vault store.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Plotter.spec.blueprints[key].flow.steps[index].arguments.copy.source.vault
<sup><sup>[↩ Parent](#plotterspecblueprintskeyflowstepsindexargumentscopysource)</sup></sup>



Holds details for retrieving credentials by the modules from Vault store.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>address</b></td>
        <td>string</td>
        <td>
          Address is Vault address<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>authPath</b></td>
        <td>string</td>
        <td>
          AuthPath is the path to auth method i.e. kubernetes<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>role</b></td>
        <td>string</td>
        <td>
          Role is the Vault role used for retrieving the credentials<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>secretPath</b></td>
        <td>string</td>
        <td>
          SecretPath is the path of the secret holding the Credentials in Vault<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Plotter.spec.blueprints[key].flow.steps[index].arguments.read[index]
<sup><sup>[↩ Parent](#plotterspecblueprintskeyflowstepsindexarguments)</sup></sup>



ReadModuleArgs define the input parameters for modules that read data from location A

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>transformations</b></td>
        <td>[]object</td>
        <td>
          Transformations are different types of processing that may be done to the data<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>assetID</b></td>
        <td>string</td>
        <td>
          AssetID identifies the asset to be used for accessing the data when it is ready It is copied from the FybrikApplication resource<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#plotterspecblueprintskeyflowstepsindexargumentsreadindexsource">source</a></b></td>
        <td>object</td>
        <td>
          Source of the read path module<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Plotter.spec.blueprints[key].flow.steps[index].arguments.read[index].source
<sup><sup>[↩ Parent](#plotterspecblueprintskeyflowstepsindexargumentsreadindex)</sup></sup>



Source of the read path module

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>connection</b></td>
        <td>object</td>
        <td>
          Connection has the relevant details for accesing the data (url, table, ssl, etc.)<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>format</b></td>
        <td>string</td>
        <td>
          Format represents data format (e.g. parquet) as received from catalog connectors<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#plotterspecblueprintskeyflowstepsindexargumentsreadindexsourcevault">vault</a></b></td>
        <td>object</td>
        <td>
          Holds details for retrieving credentials by the modules from Vault store.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Plotter.spec.blueprints[key].flow.steps[index].arguments.read[index].source.vault
<sup><sup>[↩ Parent](#plotterspecblueprintskeyflowstepsindexargumentsreadindexsource)</sup></sup>



Holds details for retrieving credentials by the modules from Vault store.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>address</b></td>
        <td>string</td>
        <td>
          Address is Vault address<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>authPath</b></td>
        <td>string</td>
        <td>
          AuthPath is the path to auth method i.e. kubernetes<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>role</b></td>
        <td>string</td>
        <td>
          Role is the Vault role used for retrieving the credentials<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>secretPath</b></td>
        <td>string</td>
        <td>
          SecretPath is the path of the secret holding the Credentials in Vault<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Plotter.spec.blueprints[key].flow.steps[index].arguments.write[index]
<sup><sup>[↩ Parent](#plotterspecblueprintskeyflowstepsindexarguments)</sup></sup>



WriteModuleArgs define the input parameters for modules that write data to location B

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>transformations</b></td>
        <td>[]object</td>
        <td>
          Transformations are different types of processing that may be done to the data as it is written.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#plotterspecblueprintskeyflowstepsindexargumentswriteindexdestination">destination</a></b></td>
        <td>object</td>
        <td>
          Destination is the data store to which the data will be written<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Plotter.spec.blueprints[key].flow.steps[index].arguments.write[index].destination
<sup><sup>[↩ Parent](#plotterspecblueprintskeyflowstepsindexargumentswriteindex)</sup></sup>



Destination is the data store to which the data will be written

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>connection</b></td>
        <td>object</td>
        <td>
          Connection has the relevant details for accesing the data (url, table, ssl, etc.)<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>format</b></td>
        <td>string</td>
        <td>
          Format represents data format (e.g. parquet) as received from catalog connectors<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#plotterspecblueprintskeyflowstepsindexargumentswriteindexdestinationvault">vault</a></b></td>
        <td>object</td>
        <td>
          Holds details for retrieving credentials by the modules from Vault store.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Plotter.spec.blueprints[key].flow.steps[index].arguments.write[index].destination.vault
<sup><sup>[↩ Parent](#plotterspecblueprintskeyflowstepsindexargumentswriteindexdestination)</sup></sup>



Holds details for retrieving credentials by the modules from Vault store.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>address</b></td>
        <td>string</td>
        <td>
          Address is Vault address<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>authPath</b></td>
        <td>string</td>
        <td>
          AuthPath is the path to auth method i.e. kubernetes<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>role</b></td>
        <td>string</td>
        <td>
          Role is the Vault role used for retrieving the credentials<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>secretPath</b></td>
        <td>string</td>
        <td>
          SecretPath is the path of the secret holding the Credentials in Vault<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Plotter.spec.blueprints[key].templates[index]
<sup><sup>[↩ Parent](#plotterspecblueprintskey)</sup></sup>



ComponentTemplate is a copy of a FybrikModule Custom Resource.  It contains the information necessary to instantiate a component in a FlowStep, which provides the functionality described by the module.  There are 3 different module types.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#plotterspecblueprintskeytemplatesindexchart">chart</a></b></td>
        <td>object</td>
        <td>
          Chart contains the location of the helm chart with info detailing how to deploy<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          Kind of k8s resource<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the template<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Plotter.spec.blueprints[key].templates[index].chart
<sup><sup>[↩ Parent](#plotterspecblueprintskeytemplatesindex)</sup></sup>



Chart contains the location of the helm chart with info detailing how to deploy

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>values</b></td>
        <td>map[string]string</td>
        <td>
          Values to pass to helm chart installation<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of helm chart<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Plotter.spec.selector
<sup><sup>[↩ Parent](#plotterspec)</sup></sup>



Selector enables to connect the resource to the application Should match the selector of the owner - FybrikApplication CRD.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#plotterspecselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          matchExpressions is a list of label selector requirements. The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### Plotter.spec.selector.matchExpressions[index]
<sup><sup>[↩ Parent](#plotterspecselector)</sup></sup>



A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Plotter.status
<sup><sup>[↩ Parent](#plotter)</sup></sup>



PlotterStatus defines the observed state of Plotter This includes readiness, error message, and indicators received from blueprint resources owned by the Plotter for cleanup and status monitoring

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#plotterstatusblueprintskey">blueprints</a></b></td>
        <td>map[string]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          ObservedGeneration is taken from the Plotter metadata.  This is used to determine during reconcile whether reconcile was called because the desired state changed, or whether status of the allocated blueprints should be checked.<br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#plotterstatusobservedstate">observedState</a></b></td>
        <td>object</td>
        <td>
          ObservedState includes information to be reported back to the FybrikApplication resource It includes readiness and error indications, as well as user instructions<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readyTimestamp</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### Plotter.status.blueprints[key]
<sup><sup>[↩ Parent](#plotterstatus)</sup></sup>



MetaBlueprint defines blueprint metadata (name, namespace) and status

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#plotterstatusblueprintskeystatus">status</a></b></td>
        <td>object</td>
        <td>
          BlueprintStatus defines the observed state of Blueprint This includes readiness, error message, and indicators forthe Kubernetes resources owned by the Blueprint for cleanup and status monitoring<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Plotter.status.blueprints[key].status
<sup><sup>[↩ Parent](#plotterstatusblueprintskey)</sup></sup>



BlueprintStatus defines the observed state of Blueprint This includes readiness, error message, and indicators forthe Kubernetes resources owned by the Blueprint for cleanup and status monitoring

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          ObservedGeneration is taken from the Blueprint metadata.  This is used to determine during reconcile whether reconcile was called because the desired state changed, or whether status of the allocated resources should be checked.<br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#plotterstatusblueprintskeystatusobservedstate">observedState</a></b></td>
        <td>object</td>
        <td>
          ObservedState includes information to be reported back to the FybrikApplication resource It includes readiness and error indications, as well as user instructions<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>releases</b></td>
        <td>map[string]integer</td>
        <td>
          Releases map each release to the observed generation of the blueprint containing this release. At the end of reconcile, each release should be mapped to the latest blueprint version or be uninstalled.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### Plotter.status.blueprints[key].status.observedState
<sup><sup>[↩ Parent](#plotterstatusblueprintskeystatus)</sup></sup>



ObservedState includes information to be reported back to the FybrikApplication resource It includes readiness and error indications, as well as user instructions

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>dataAccessInstructions</b></td>
        <td>string</td>
        <td>
          DataAccessInstructions indicate how the data user or his application may access the data. Instructions are available upon successful orchestration.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>error</b></td>
        <td>string</td>
        <td>
          Error indicates that there has been an error to orchestrate the modules and provides the error message<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>ready</b></td>
        <td>boolean</td>
        <td>
          Ready represents that the modules have been orchestrated successfully and the data is ready for usage<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### Plotter.status.observedState
<sup><sup>[↩ Parent](#plotterstatus)</sup></sup>



ObservedState includes information to be reported back to the FybrikApplication resource It includes readiness and error indications, as well as user instructions

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>dataAccessInstructions</b></td>
        <td>string</td>
        <td>
          DataAccessInstructions indicate how the data user or his application may access the data. Instructions are available upon successful orchestration.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>error</b></td>
        <td>string</td>
        <td>
          Error indicates that there has been an error to orchestrate the modules and provides the error message<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>ready</b></td>
        <td>boolean</td>
        <td>
          Ready represents that the modules have been orchestrated successfully and the data is ready for usage<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>

## katalog.fybrik.io/v1alpha1

Resource Types:

- [Asset](#asset)




### Asset
<sup><sup>[↩ Parent](#katalogfybrikiov1alpha1 )</sup></sup>








<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>katalog.fybrik.io/v1alpha1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>Asset</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#assetspec">spec</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Asset.spec
<sup><sup>[↩ Parent](#asset)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#assetspecassetdetails">assetDetails</a></b></td>
        <td>object</td>
        <td>
          Asset details<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#assetspecassetmetadata">assetMetadata</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#assetspecsecretref">secretRef</a></b></td>
        <td>object</td>
        <td>
          Reference to a Secret resource holding credentials for this asset<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Asset.spec.assetDetails
<sup><sup>[↩ Parent](#assetspec)</sup></sup>



Asset details

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>dataFormat</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#assetspecassetdetailsconnection">connection</a></b></td>
        <td>object</td>
        <td>
          Connection information<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Asset.spec.assetDetails.connection
<sup><sup>[↩ Parent](#assetspecassetdetails)</sup></sup>



Connection information

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#assetspecassetdetailsconnectiondb2">db2</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#assetspecassetdetailsconnectionkafka">kafka</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#assetspecassetdetailsconnections3">s3</a></b></td>
        <td>object</td>
        <td>
          Connection information for S3 compatible object store<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          <br/>
          <br/>
            <i>Enum</i>: s3, db2, kafka<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Asset.spec.assetDetails.connection.db2
<sup><sup>[↩ Parent](#assetspecassetdetailsconnection)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>database</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>port</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>ssl</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>table</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>url</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### Asset.spec.assetDetails.connection.kafka
<sup><sup>[↩ Parent](#assetspecassetdetailsconnection)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>bootstrap_servers</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>key_deserializer</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sasl_mechanism</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>schema_registry</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>security_protocol</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>ssl_truststore</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>ssl_truststore_password</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>topic_name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>value_deserializer</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### Asset.spec.assetDetails.connection.s3
<sup><sup>[↩ Parent](#assetspecassetdetailsconnection)</sup></sup>



Connection information for S3 compatible object store

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>region</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>bucket</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>endpoint</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>objectKey</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Asset.spec.assetMetadata
<sup><sup>[↩ Parent](#assetspec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#assetspecassetmetadatacomponentsmetadatakey">componentsMetadata</a></b></td>
        <td>map[string]object</td>
        <td>
          metadata for each component in asset (e.g., column)<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>geography</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namedMetadata</b></td>
        <td>map[string]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>owner</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>tags</b></td>
        <td>[]string</td>
        <td>
          Tags associated with the asset<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### Asset.spec.assetMetadata.componentsMetadata[key]
<sup><sup>[↩ Parent](#assetspecassetmetadata)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>componentType</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namedMetadata</b></td>
        <td>map[string]string</td>
        <td>
          Named terms, that exist in Catalog toxonomy and the values for these terms for columns we will have "SchemaDetails" key, that will include technical schema details for this column TODO: Consider create special field for schema outside of metadata<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>tags</b></td>
        <td>[]string</td>
        <td>
          Tags - can be any free text added to a component (no taxonomy)<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### Asset.spec.secretRef
<sup><sup>[↩ Parent](#assetspec)</sup></sup>



Reference to a Secret resource holding credentials for this asset

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the Secret resource (must exist in the same namespace)<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>

## motion.fybrik.io/v1alpha1

Resource Types:

- [BatchTransfer](#batchtransfer)

- [StreamTransfer](#streamtransfer)




### BatchTransfer
<sup><sup>[↩ Parent](#motionfybrikiov1alpha1 )</sup></sup>






BatchTransfer is the Schema for the batchtransfers API

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>motion.fybrik.io/v1alpha1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>BatchTransfer</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#batchtransferspec">spec</a></b></td>
        <td>object</td>
        <td>
          BatchTransferSpec defines the state of a BatchTransfer. The state includes source/destination specification, a schedule and the means by which data movement is to be conducted. The means is given as a kubernetes job description. In addition, the state also contains a sketch of a transformation instruction. In future releases, the transformation description should be specified in a separate CRD.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#batchtransferstatus">status</a></b></td>
        <td>object</td>
        <td>
          BatchTransferStatus defines the observed state of BatchTransfer This includes a reference to the job that implements the movement as well as the last schedule time. What is missing: Extended status information such as: - number of records moved - technical meta-data<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### BatchTransfer.spec
<sup><sup>[↩ Parent](#batchtransfer)</sup></sup>



BatchTransferSpec defines the state of a BatchTransfer. The state includes source/destination specification, a schedule and the means by which data movement is to be conducted. The means is given as a kubernetes job description. In addition, the state also contains a sketch of a transformation instruction. In future releases, the transformation description should be specified in a separate CRD.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>failedJobHistoryLimit</b></td>
        <td>integer</td>
        <td>
          Maximal number of failed Kubernetes job objects that should be kept. This property will be defaulted by the webhook if not set.<br/>
          <br/>
            <i>Minimum</i>: 0<br/>
            <i>Maximum</i>: 20<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>flowType</b></td>
        <td>enum</td>
        <td>
          Data flow type that specifies if this is a stream or a batch workflow<br/>
          <br/>
            <i>Enum</i>: Batch, Stream<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>image</b></td>
        <td>string</td>
        <td>
          Image that should be used for the actual batch job. This is usually a datamover image. This property will be defaulted by the webhook if not set.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>imagePullPolicy</b></td>
        <td>string</td>
        <td>
          Image pull policy that should be used for the actual job. This property will be defaulted by the webhook if not set.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>maxFailedRetries</b></td>
        <td>integer</td>
        <td>
          Maximal number of failed retries until the batch job should stop trying. This property will be defaulted by the webhook if not set.<br/>
          <br/>
            <i>Minimum</i>: 0<br/>
            <i>Maximum</i>: 10<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>noFinalizer</b></td>
        <td>boolean</td>
        <td>
          If this batch job instance should have a finalizer or not. This property will be defaulted by the webhook if not set.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readDataType</b></td>
        <td>enum</td>
        <td>
          Data type of the data that is read from source (log data or change data)<br/>
          <br/>
            <i>Enum</i>: LogData, ChangeData<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>schedule</b></td>
        <td>string</td>
        <td>
          Cron schedule if this BatchTransfer job should run on a regular schedule. Values are specified like cron job schedules. A good translation to human language can be found here https://crontab.guru/<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>secretProviderRole</b></td>
        <td>string</td>
        <td>
          Secret provider role that should be used for the actual job. This property will be defaulted by the webhook if not set.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>secretProviderURL</b></td>
        <td>string</td>
        <td>
          Secret provider url that should be used for the actual job. This property will be defaulted by the webhook if not set.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#batchtransferspecspark">spark</a></b></td>
        <td>object</td>
        <td>
          Optional Spark configuration for tuning<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>successfulJobHistoryLimit</b></td>
        <td>integer</td>
        <td>
          Maximal number of successful Kubernetes job objects that should be kept. This property will be defaulted by the webhook if not set.<br/>
          <br/>
            <i>Minimum</i>: 0<br/>
            <i>Maximum</i>: 20<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>suspend</b></td>
        <td>boolean</td>
        <td>
          If this batch job instance is run on a schedule the regular schedule can be suspended with this property. This property will be defaulted by the webhook if not set.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#batchtransferspectransformationindex">transformation</a></b></td>
        <td>[]object</td>
        <td>
          Transformations to be applied to the source data before writing to destination<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>writeDataType</b></td>
        <td>enum</td>
        <td>
          Data type of how the data should be written to the target (log data or change data)<br/>
          <br/>
            <i>Enum</i>: LogData, ChangeData<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>writeOperation</b></td>
        <td>enum</td>
        <td>
          Write operation that should be performed when writing (overwrite,append,update) Caution: Some write operations are only available for batch and some only for stream.<br/>
          <br/>
            <i>Enum</i>: Overwrite, Append, Update<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#batchtransferspecdestination">destination</a></b></td>
        <td>object</td>
        <td>
          Destination data store for this batch job<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#batchtransferspecsource">source</a></b></td>
        <td>object</td>
        <td>
          Source data store for this batch job<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### BatchTransfer.spec.spark
<sup><sup>[↩ Parent](#batchtransferspec)</sup></sup>



Optional Spark configuration for tuning

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>appName</b></td>
        <td>string</td>
        <td>
          Name of the transaction. Mainly used for debugging and lineage tracking.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>driverCores</b></td>
        <td>integer</td>
        <td>
          Number of cores that the driver should use<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>driverMemory</b></td>
        <td>integer</td>
        <td>
          Memory that the driver should have<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>executorCores</b></td>
        <td>integer</td>
        <td>
          Number of cores that each executor should have<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>executorMemory</b></td>
        <td>string</td>
        <td>
          Memory that each executor should have<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>image</b></td>
        <td>string</td>
        <td>
          Image to be used for executors<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>imagePullPolicy</b></td>
        <td>string</td>
        <td>
          Image pull policy to be used for executor<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>numExecutors</b></td>
        <td>integer</td>
        <td>
          Number of executors to be started<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>options</b></td>
        <td>map[string]string</td>
        <td>
          Additional options for Spark configuration.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>shufflePartitions</b></td>
        <td>integer</td>
        <td>
          Number of shuffle partitions for Spark<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### BatchTransfer.spec.transformation[index]
<sup><sup>[↩ Parent](#batchtransferspec)</sup></sup>



to be refined...

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>action</b></td>
        <td>enum</td>
        <td>
          Transformation action that should be performed.<br/>
          <br/>
            <i>Enum</i>: RemoveColumns, EncryptColumns, DigestColumns, RedactColumns, SampleRows, FilterRows<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>columns</b></td>
        <td>[]string</td>
        <td>
          Columns that are involved in this action. This property is optional as for some actions no columns have to be specified. E.g. filter is a row based transformation.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the transaction. Mainly used for debugging and lineage tracking.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>options</b></td>
        <td>map[string]string</td>
        <td>
          Additional options for this transformation.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### BatchTransfer.spec.destination
<sup><sup>[↩ Parent](#batchtransferspec)</sup></sup>



Destination data store for this batch job

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#batchtransferspecdestinationcloudant">cloudant</a></b></td>
        <td>object</td>
        <td>
          IBM Cloudant. Needs cloudant legacy credentials.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#batchtransferspecdestinationdatabase">database</a></b></td>
        <td>object</td>
        <td>
          Database data store. For the moment only Db2 is supported.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>description</b></td>
        <td>string</td>
        <td>
          Description of the transfer in human readable form that is displayed in the kubectl get If not provided this will be filled in depending on the datastore that is specified.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#batchtransferspecdestinationkafka">kafka</a></b></td>
        <td>object</td>
        <td>
          Kafka data store. The supposed format within the given Kafka topic is a Confluent compatible format stored as Avro. A schema registry needs to be specified as well.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#batchtransferspecdestinations3">s3</a></b></td>
        <td>object</td>
        <td>
          An object store data store that is compatible with S3. This can be a COS bucket.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### BatchTransfer.spec.destination.cloudant
<sup><sup>[↩ Parent](#batchtransferspecdestination)</sup></sup>



IBM Cloudant. Needs cloudant legacy credentials.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>password</b></td>
        <td>string</td>
        <td>
          Cloudant password. Can be retrieved from vault if specified in vault parameter and is thus optional.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>secretImport</b></td>
        <td>string</td>
        <td>
          Define a secret import definition.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>username</b></td>
        <td>string</td>
        <td>
          Cloudant user. Can be retrieved from vault if specified in vault parameter and is thus optional.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#batchtransferspecdestinationcloudantvault">vault</a></b></td>
        <td>object</td>
        <td>
          Define secrets that are fetched from a Vault instance<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>database</b></td>
        <td>string</td>
        <td>
          Database to be read from/written to<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          Host of cloudant instance<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### BatchTransfer.spec.destination.cloudant.vault
<sup><sup>[↩ Parent](#batchtransferspecdestinationcloudant)</sup></sup>



Define secrets that are fetched from a Vault instance

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>address</b></td>
        <td>string</td>
        <td>
          Address is Vault address<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>authPath</b></td>
        <td>string</td>
        <td>
          AuthPath is the path to auth method i.e. kubernetes<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>role</b></td>
        <td>string</td>
        <td>
          Role is the Vault role used for retrieving the credentials<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>secretPath</b></td>
        <td>string</td>
        <td>
          SecretPath is the path of the secret holding the Credentials in Vault<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### BatchTransfer.spec.destination.database
<sup><sup>[↩ Parent](#batchtransferspecdestination)</sup></sup>



Database data store. For the moment only Db2 is supported.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>password</b></td>
        <td>string</td>
        <td>
          Database password. Can be retrieved from vault if specified in vault parameter and is thus optional.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>secretImport</b></td>
        <td>string</td>
        <td>
          Define a secret import definition.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>user</b></td>
        <td>string</td>
        <td>
          Database user. Can be retrieved from vault if specified in vault parameter and is thus optional.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#batchtransferspecdestinationdatabasevault">vault</a></b></td>
        <td>object</td>
        <td>
          Define secrets that are fetched from a Vault instance<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>db2URL</b></td>
        <td>string</td>
        <td>
          URL to Db2 instance in JDBC format Supported SSL certificates are currently certificates signed with IBM Intermediate CA or cloud signed certificates.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>table</b></td>
        <td>string</td>
        <td>
          Table to be read<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### BatchTransfer.spec.destination.database.vault
<sup><sup>[↩ Parent](#batchtransferspecdestinationdatabase)</sup></sup>



Define secrets that are fetched from a Vault instance

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>address</b></td>
        <td>string</td>
        <td>
          Address is Vault address<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>authPath</b></td>
        <td>string</td>
        <td>
          AuthPath is the path to auth method i.e. kubernetes<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>role</b></td>
        <td>string</td>
        <td>
          Role is the Vault role used for retrieving the credentials<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>secretPath</b></td>
        <td>string</td>
        <td>
          SecretPath is the path of the secret holding the Credentials in Vault<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### BatchTransfer.spec.destination.kafka
<sup><sup>[↩ Parent](#batchtransferspecdestination)</sup></sup>



Kafka data store. The supposed format within the given Kafka topic is a Confluent compatible format stored as Avro. A schema registry needs to be specified as well.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>createSnapshot</b></td>
        <td>boolean</td>
        <td>
          If a snapshot should be created of the topic. Records in Kafka are stored as key-value pairs. Updates/Deletes for the same key are appended to the Kafka topic and the last value for a given key is the valid key in a Snapshot. When this property is true only the last value will be written. If the property is false all values will be written out. As a CDC example: If the property is true a valid snapshot of the log stream will be created. If the property is false the CDC stream will be dumped as is like a change log.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>dataFormat</b></td>
        <td>string</td>
        <td>
          Data format of the objects in S3. e.g. parquet or csv. Please refer to struct for allowed values.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>keyDeserializer</b></td>
        <td>string</td>
        <td>
          Deserializer to be used for the keys of the topic<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>password</b></td>
        <td>string</td>
        <td>
          Kafka user password Can be retrieved from vault if specified in vault parameter and is thus optional.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>saslMechanism</b></td>
        <td>string</td>
        <td>
          SASL Mechanism to be used (e.g. PLAIN or SCRAM-SHA-512) Default SCRAM-SHA-512 will be assumed if not specified<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>secretImport</b></td>
        <td>string</td>
        <td>
          Define a secret import definition.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>securityProtocol</b></td>
        <td>string</td>
        <td>
          Kafka security protocol one of (PLAINTEXT, SASL_PLAINTEXT, SASL_SSL, SSL) Default SASL_SSL will be assumed if not specified<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sslTruststore</b></td>
        <td>string</td>
        <td>
          A truststore or certificate encoded as base64. The format can be JKS or PKCS12. A truststore can be specified like this or in a predefined Kubernetes secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sslTruststoreLocation</b></td>
        <td>string</td>
        <td>
          SSL truststore location.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sslTruststorePassword</b></td>
        <td>string</td>
        <td>
          SSL truststore password.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sslTruststoreSecret</b></td>
        <td>string</td>
        <td>
          Kubernetes secret that contains the SSL truststore. The format can be JKS or PKCS12. A truststore can be specified like this or as<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>user</b></td>
        <td>string</td>
        <td>
          Kafka user name. Can be retrieved from vault if specified in vault parameter and is thus optional.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>valueDeserializer</b></td>
        <td>string</td>
        <td>
          Deserializer to be used for the values of the topic<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#batchtransferspecdestinationkafkavault">vault</a></b></td>
        <td>object</td>
        <td>
          Define secrets that are fetched from a Vault instance<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>kafkaBrokers</b></td>
        <td>string</td>
        <td>
          Kafka broker URLs as a comma separated list.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>kafkaTopic</b></td>
        <td>string</td>
        <td>
          Kafka topic<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>schemaRegistryURL</b></td>
        <td>string</td>
        <td>
          URL to the schema registry. The registry has to be Confluent schema registry compatible.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### BatchTransfer.spec.destination.kafka.vault
<sup><sup>[↩ Parent](#batchtransferspecdestinationkafka)</sup></sup>



Define secrets that are fetched from a Vault instance

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>address</b></td>
        <td>string</td>
        <td>
          Address is Vault address<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>authPath</b></td>
        <td>string</td>
        <td>
          AuthPath is the path to auth method i.e. kubernetes<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>role</b></td>
        <td>string</td>
        <td>
          Role is the Vault role used for retrieving the credentials<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>secretPath</b></td>
        <td>string</td>
        <td>
          SecretPath is the path of the secret holding the Credentials in Vault<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### BatchTransfer.spec.destination.s3
<sup><sup>[↩ Parent](#batchtransferspecdestination)</sup></sup>



An object store data store that is compatible with S3. This can be a COS bucket.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>accessKey</b></td>
        <td>string</td>
        <td>
          Access key of the HMAC credentials that can access the given bucket. Can be retrieved from vault if specified in vault parameter and is thus optional.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>dataFormat</b></td>
        <td>string</td>
        <td>
          Data format of the objects in S3. e.g. parquet or csv. Please refer to struct for allowed values.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>partitionBy</b></td>
        <td>[]string</td>
        <td>
          Partition by partition (for target data stores) Defines the columns to partition the output by for a target data store.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>region</b></td>
        <td>string</td>
        <td>
          Region of S3 service<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>secretImport</b></td>
        <td>string</td>
        <td>
          Define a secret import definition.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>secretKey</b></td>
        <td>string</td>
        <td>
          Secret key of the HMAC credentials that can access the given bucket. Can be retrieved from vault if specified in vault parameter and is thus optional.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#batchtransferspecdestinations3vault">vault</a></b></td>
        <td>object</td>
        <td>
          Define secrets that are fetched from a Vault instance<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>bucket</b></td>
        <td>string</td>
        <td>
          Bucket of S3 service<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>endpoint</b></td>
        <td>string</td>
        <td>
          Endpoint of S3 service<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>objectKey</b></td>
        <td>string</td>
        <td>
          Object key of the object in S3. This is used as a prefix! Thus all objects that have the given objectKey as prefix will be used as input!<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### BatchTransfer.spec.destination.s3.vault
<sup><sup>[↩ Parent](#batchtransferspecdestinations3)</sup></sup>



Define secrets that are fetched from a Vault instance

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>address</b></td>
        <td>string</td>
        <td>
          Address is Vault address<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>authPath</b></td>
        <td>string</td>
        <td>
          AuthPath is the path to auth method i.e. kubernetes<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>role</b></td>
        <td>string</td>
        <td>
          Role is the Vault role used for retrieving the credentials<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>secretPath</b></td>
        <td>string</td>
        <td>
          SecretPath is the path of the secret holding the Credentials in Vault<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### BatchTransfer.spec.source
<sup><sup>[↩ Parent](#batchtransferspec)</sup></sup>



Source data store for this batch job

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#batchtransferspecsourcecloudant">cloudant</a></b></td>
        <td>object</td>
        <td>
          IBM Cloudant. Needs cloudant legacy credentials.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#batchtransferspecsourcedatabase">database</a></b></td>
        <td>object</td>
        <td>
          Database data store. For the moment only Db2 is supported.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>description</b></td>
        <td>string</td>
        <td>
          Description of the transfer in human readable form that is displayed in the kubectl get If not provided this will be filled in depending on the datastore that is specified.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#batchtransferspecsourcekafka">kafka</a></b></td>
        <td>object</td>
        <td>
          Kafka data store. The supposed format within the given Kafka topic is a Confluent compatible format stored as Avro. A schema registry needs to be specified as well.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#batchtransferspecsources3">s3</a></b></td>
        <td>object</td>
        <td>
          An object store data store that is compatible with S3. This can be a COS bucket.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### BatchTransfer.spec.source.cloudant
<sup><sup>[↩ Parent](#batchtransferspecsource)</sup></sup>



IBM Cloudant. Needs cloudant legacy credentials.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>password</b></td>
        <td>string</td>
        <td>
          Cloudant password. Can be retrieved from vault if specified in vault parameter and is thus optional.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>secretImport</b></td>
        <td>string</td>
        <td>
          Define a secret import definition.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>username</b></td>
        <td>string</td>
        <td>
          Cloudant user. Can be retrieved from vault if specified in vault parameter and is thus optional.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#batchtransferspecsourcecloudantvault">vault</a></b></td>
        <td>object</td>
        <td>
          Define secrets that are fetched from a Vault instance<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>database</b></td>
        <td>string</td>
        <td>
          Database to be read from/written to<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          Host of cloudant instance<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### BatchTransfer.spec.source.cloudant.vault
<sup><sup>[↩ Parent](#batchtransferspecsourcecloudant)</sup></sup>



Define secrets that are fetched from a Vault instance

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>address</b></td>
        <td>string</td>
        <td>
          Address is Vault address<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>authPath</b></td>
        <td>string</td>
        <td>
          AuthPath is the path to auth method i.e. kubernetes<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>role</b></td>
        <td>string</td>
        <td>
          Role is the Vault role used for retrieving the credentials<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>secretPath</b></td>
        <td>string</td>
        <td>
          SecretPath is the path of the secret holding the Credentials in Vault<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### BatchTransfer.spec.source.database
<sup><sup>[↩ Parent](#batchtransferspecsource)</sup></sup>



Database data store. For the moment only Db2 is supported.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>password</b></td>
        <td>string</td>
        <td>
          Database password. Can be retrieved from vault if specified in vault parameter and is thus optional.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>secretImport</b></td>
        <td>string</td>
        <td>
          Define a secret import definition.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>user</b></td>
        <td>string</td>
        <td>
          Database user. Can be retrieved from vault if specified in vault parameter and is thus optional.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#batchtransferspecsourcedatabasevault">vault</a></b></td>
        <td>object</td>
        <td>
          Define secrets that are fetched from a Vault instance<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>db2URL</b></td>
        <td>string</td>
        <td>
          URL to Db2 instance in JDBC format Supported SSL certificates are currently certificates signed with IBM Intermediate CA or cloud signed certificates.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>table</b></td>
        <td>string</td>
        <td>
          Table to be read<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### BatchTransfer.spec.source.database.vault
<sup><sup>[↩ Parent](#batchtransferspecsourcedatabase)</sup></sup>



Define secrets that are fetched from a Vault instance

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>address</b></td>
        <td>string</td>
        <td>
          Address is Vault address<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>authPath</b></td>
        <td>string</td>
        <td>
          AuthPath is the path to auth method i.e. kubernetes<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>role</b></td>
        <td>string</td>
        <td>
          Role is the Vault role used for retrieving the credentials<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>secretPath</b></td>
        <td>string</td>
        <td>
          SecretPath is the path of the secret holding the Credentials in Vault<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### BatchTransfer.spec.source.kafka
<sup><sup>[↩ Parent](#batchtransferspecsource)</sup></sup>



Kafka data store. The supposed format within the given Kafka topic is a Confluent compatible format stored as Avro. A schema registry needs to be specified as well.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>createSnapshot</b></td>
        <td>boolean</td>
        <td>
          If a snapshot should be created of the topic. Records in Kafka are stored as key-value pairs. Updates/Deletes for the same key are appended to the Kafka topic and the last value for a given key is the valid key in a Snapshot. When this property is true only the last value will be written. If the property is false all values will be written out. As a CDC example: If the property is true a valid snapshot of the log stream will be created. If the property is false the CDC stream will be dumped as is like a change log.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>dataFormat</b></td>
        <td>string</td>
        <td>
          Data format of the objects in S3. e.g. parquet or csv. Please refer to struct for allowed values.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>keyDeserializer</b></td>
        <td>string</td>
        <td>
          Deserializer to be used for the keys of the topic<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>password</b></td>
        <td>string</td>
        <td>
          Kafka user password Can be retrieved from vault if specified in vault parameter and is thus optional.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>saslMechanism</b></td>
        <td>string</td>
        <td>
          SASL Mechanism to be used (e.g. PLAIN or SCRAM-SHA-512) Default SCRAM-SHA-512 will be assumed if not specified<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>secretImport</b></td>
        <td>string</td>
        <td>
          Define a secret import definition.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>securityProtocol</b></td>
        <td>string</td>
        <td>
          Kafka security protocol one of (PLAINTEXT, SASL_PLAINTEXT, SASL_SSL, SSL) Default SASL_SSL will be assumed if not specified<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sslTruststore</b></td>
        <td>string</td>
        <td>
          A truststore or certificate encoded as base64. The format can be JKS or PKCS12. A truststore can be specified like this or in a predefined Kubernetes secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sslTruststoreLocation</b></td>
        <td>string</td>
        <td>
          SSL truststore location.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sslTruststorePassword</b></td>
        <td>string</td>
        <td>
          SSL truststore password.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sslTruststoreSecret</b></td>
        <td>string</td>
        <td>
          Kubernetes secret that contains the SSL truststore. The format can be JKS or PKCS12. A truststore can be specified like this or as<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>user</b></td>
        <td>string</td>
        <td>
          Kafka user name. Can be retrieved from vault if specified in vault parameter and is thus optional.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>valueDeserializer</b></td>
        <td>string</td>
        <td>
          Deserializer to be used for the values of the topic<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#batchtransferspecsourcekafkavault">vault</a></b></td>
        <td>object</td>
        <td>
          Define secrets that are fetched from a Vault instance<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>kafkaBrokers</b></td>
        <td>string</td>
        <td>
          Kafka broker URLs as a comma separated list.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>kafkaTopic</b></td>
        <td>string</td>
        <td>
          Kafka topic<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>schemaRegistryURL</b></td>
        <td>string</td>
        <td>
          URL to the schema registry. The registry has to be Confluent schema registry compatible.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### BatchTransfer.spec.source.kafka.vault
<sup><sup>[↩ Parent](#batchtransferspecsourcekafka)</sup></sup>



Define secrets that are fetched from a Vault instance

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>address</b></td>
        <td>string</td>
        <td>
          Address is Vault address<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>authPath</b></td>
        <td>string</td>
        <td>
          AuthPath is the path to auth method i.e. kubernetes<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>role</b></td>
        <td>string</td>
        <td>
          Role is the Vault role used for retrieving the credentials<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>secretPath</b></td>
        <td>string</td>
        <td>
          SecretPath is the path of the secret holding the Credentials in Vault<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### BatchTransfer.spec.source.s3
<sup><sup>[↩ Parent](#batchtransferspecsource)</sup></sup>



An object store data store that is compatible with S3. This can be a COS bucket.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>accessKey</b></td>
        <td>string</td>
        <td>
          Access key of the HMAC credentials that can access the given bucket. Can be retrieved from vault if specified in vault parameter and is thus optional.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>dataFormat</b></td>
        <td>string</td>
        <td>
          Data format of the objects in S3. e.g. parquet or csv. Please refer to struct for allowed values.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>partitionBy</b></td>
        <td>[]string</td>
        <td>
          Partition by partition (for target data stores) Defines the columns to partition the output by for a target data store.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>region</b></td>
        <td>string</td>
        <td>
          Region of S3 service<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>secretImport</b></td>
        <td>string</td>
        <td>
          Define a secret import definition.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>secretKey</b></td>
        <td>string</td>
        <td>
          Secret key of the HMAC credentials that can access the given bucket. Can be retrieved from vault if specified in vault parameter and is thus optional.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#batchtransferspecsources3vault">vault</a></b></td>
        <td>object</td>
        <td>
          Define secrets that are fetched from a Vault instance<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>bucket</b></td>
        <td>string</td>
        <td>
          Bucket of S3 service<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>endpoint</b></td>
        <td>string</td>
        <td>
          Endpoint of S3 service<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>objectKey</b></td>
        <td>string</td>
        <td>
          Object key of the object in S3. This is used as a prefix! Thus all objects that have the given objectKey as prefix will be used as input!<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### BatchTransfer.spec.source.s3.vault
<sup><sup>[↩ Parent](#batchtransferspecsources3)</sup></sup>



Define secrets that are fetched from a Vault instance

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>address</b></td>
        <td>string</td>
        <td>
          Address is Vault address<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>authPath</b></td>
        <td>string</td>
        <td>
          AuthPath is the path to auth method i.e. kubernetes<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>role</b></td>
        <td>string</td>
        <td>
          Role is the Vault role used for retrieving the credentials<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>secretPath</b></td>
        <td>string</td>
        <td>
          SecretPath is the path of the secret holding the Credentials in Vault<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### BatchTransfer.status
<sup><sup>[↩ Parent](#batchtransfer)</sup></sup>



BatchTransferStatus defines the observed state of BatchTransfer This includes a reference to the job that implements the movement as well as the last schedule time. What is missing: Extended status information such as: - number of records moved - technical meta-data

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#batchtransferstatusactive">active</a></b></td>
        <td>object</td>
        <td>
          A pointer to the currently running job (or nil)<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>error</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#batchtransferstatuslastcompleted">lastCompleted</a></b></td>
        <td>object</td>
        <td>
          ObjectReference contains enough information to let you inspect or modify the referred object. --- New uses of this type are discouraged because of difficulty describing its usage when embedded in APIs.  1. Ignored fields.  It includes many fields which are not generally honored.  For instance, ResourceVersion and FieldPath are both very rarely valid in actual usage.  2. Invalid usage help.  It is impossible to add specific help for individual usage.  In most embedded usages, there are particular     restrictions like, "must refer only to types A and B" or "UID not honored" or "name must be restricted".     Those cannot be well described when embedded.  3. Inconsistent validation.  Because the usages are different, the validation rules are different by usage, which makes it hard for users to predict what will happen.  4. The fields are both imprecise and overly precise.  Kind is not a precise mapping to a URL. This can produce ambiguity     during interpretation and require a REST mapping.  In most cases, the dependency is on the group,resource tuple     and the version of the actual struct is irrelevant.  5. We cannot easily change it.  Because this type is embedded in many locations, updates to this type     will affect numerous schemas.  Don't make new APIs embed an underspecified API type they do not control. Instead of using this type, create a locally provided and used type that is well-focused on your reference. For example, ServiceReferences for admission registration: https://github.com/kubernetes/api/blob/release-1.17/admissionregistration/v1/types.go#L533 .<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#batchtransferstatuslastfailed">lastFailed</a></b></td>
        <td>object</td>
        <td>
          ObjectReference contains enough information to let you inspect or modify the referred object. --- New uses of this type are discouraged because of difficulty describing its usage when embedded in APIs.  1. Ignored fields.  It includes many fields which are not generally honored.  For instance, ResourceVersion and FieldPath are both very rarely valid in actual usage.  2. Invalid usage help.  It is impossible to add specific help for individual usage.  In most embedded usages, there are particular     restrictions like, "must refer only to types A and B" or "UID not honored" or "name must be restricted".     Those cannot be well described when embedded.  3. Inconsistent validation.  Because the usages are different, the validation rules are different by usage, which makes it hard for users to predict what will happen.  4. The fields are both imprecise and overly precise.  Kind is not a precise mapping to a URL. This can produce ambiguity     during interpretation and require a REST mapping.  In most cases, the dependency is on the group,resource tuple     and the version of the actual struct is irrelevant.  5. We cannot easily change it.  Because this type is embedded in many locations, updates to this type     will affect numerous schemas.  Don't make new APIs embed an underspecified API type they do not control. Instead of using this type, create a locally provided and used type that is well-focused on your reference. For example, ServiceReferences for admission registration: https://github.com/kubernetes/api/blob/release-1.17/admissionregistration/v1/types.go#L533 .<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>lastRecordTime</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>lastScheduleTime</b></td>
        <td>string</td>
        <td>
          Information when was the last time the job was successfully scheduled.<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>lastSuccessTime</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>numRecords</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          <br/>
          <br/>
            <i>Enum</i>: STARTING, RUNNING, SUCCEEDED, FAILED<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### BatchTransfer.status.active
<sup><sup>[↩ Parent](#batchtransferstatus)</sup></sup>



A pointer to the currently running job (or nil)

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>apiVersion</b></td>
        <td>string</td>
        <td>
          API version of the referent.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>fieldPath</b></td>
        <td>string</td>
        <td>
          If referring to a piece of an object instead of an entire object, this string should contain a valid JSON/Go field access statement, such as desiredState.manifest.containers[2]. For example, if the object reference is to a container within a pod, this would take on a value like: "spec.containers{name}" (where "name" refers to the name of the container that triggered the event) or if no container name is specified "spec.containers[2]" (container with index 2 in this pod). This syntax is chosen only to have some well-defined way of referencing a part of an object. TODO: this design is not final and this field is subject to change in the future.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          Kind of the referent. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>resourceVersion</b></td>
        <td>string</td>
        <td>
          Specific resourceVersion to which this reference is made, if any. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>uid</b></td>
        <td>string</td>
        <td>
          UID of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#uids<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### BatchTransfer.status.lastCompleted
<sup><sup>[↩ Parent](#batchtransferstatus)</sup></sup>



ObjectReference contains enough information to let you inspect or modify the referred object. --- New uses of this type are discouraged because of difficulty describing its usage when embedded in APIs.  1. Ignored fields.  It includes many fields which are not generally honored.  For instance, ResourceVersion and FieldPath are both very rarely valid in actual usage.  2. Invalid usage help.  It is impossible to add specific help for individual usage.  In most embedded usages, there are particular     restrictions like, "must refer only to types A and B" or "UID not honored" or "name must be restricted".     Those cannot be well described when embedded.  3. Inconsistent validation.  Because the usages are different, the validation rules are different by usage, which makes it hard for users to predict what will happen.  4. The fields are both imprecise and overly precise.  Kind is not a precise mapping to a URL. This can produce ambiguity     during interpretation and require a REST mapping.  In most cases, the dependency is on the group,resource tuple     and the version of the actual struct is irrelevant.  5. We cannot easily change it.  Because this type is embedded in many locations, updates to this type     will affect numerous schemas.  Don't make new APIs embed an underspecified API type they do not control. Instead of using this type, create a locally provided and used type that is well-focused on your reference. For example, ServiceReferences for admission registration: https://github.com/kubernetes/api/blob/release-1.17/admissionregistration/v1/types.go#L533 .

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>apiVersion</b></td>
        <td>string</td>
        <td>
          API version of the referent.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>fieldPath</b></td>
        <td>string</td>
        <td>
          If referring to a piece of an object instead of an entire object, this string should contain a valid JSON/Go field access statement, such as desiredState.manifest.containers[2]. For example, if the object reference is to a container within a pod, this would take on a value like: "spec.containers{name}" (where "name" refers to the name of the container that triggered the event) or if no container name is specified "spec.containers[2]" (container with index 2 in this pod). This syntax is chosen only to have some well-defined way of referencing a part of an object. TODO: this design is not final and this field is subject to change in the future.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          Kind of the referent. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>resourceVersion</b></td>
        <td>string</td>
        <td>
          Specific resourceVersion to which this reference is made, if any. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>uid</b></td>
        <td>string</td>
        <td>
          UID of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#uids<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### BatchTransfer.status.lastFailed
<sup><sup>[↩ Parent](#batchtransferstatus)</sup></sup>



ObjectReference contains enough information to let you inspect or modify the referred object. --- New uses of this type are discouraged because of difficulty describing its usage when embedded in APIs.  1. Ignored fields.  It includes many fields which are not generally honored.  For instance, ResourceVersion and FieldPath are both very rarely valid in actual usage.  2. Invalid usage help.  It is impossible to add specific help for individual usage.  In most embedded usages, there are particular     restrictions like, "must refer only to types A and B" or "UID not honored" or "name must be restricted".     Those cannot be well described when embedded.  3. Inconsistent validation.  Because the usages are different, the validation rules are different by usage, which makes it hard for users to predict what will happen.  4. The fields are both imprecise and overly precise.  Kind is not a precise mapping to a URL. This can produce ambiguity     during interpretation and require a REST mapping.  In most cases, the dependency is on the group,resource tuple     and the version of the actual struct is irrelevant.  5. We cannot easily change it.  Because this type is embedded in many locations, updates to this type     will affect numerous schemas.  Don't make new APIs embed an underspecified API type they do not control. Instead of using this type, create a locally provided and used type that is well-focused on your reference. For example, ServiceReferences for admission registration: https://github.com/kubernetes/api/blob/release-1.17/admissionregistration/v1/types.go#L533 .

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>apiVersion</b></td>
        <td>string</td>
        <td>
          API version of the referent.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>fieldPath</b></td>
        <td>string</td>
        <td>
          If referring to a piece of an object instead of an entire object, this string should contain a valid JSON/Go field access statement, such as desiredState.manifest.containers[2]. For example, if the object reference is to a container within a pod, this would take on a value like: "spec.containers{name}" (where "name" refers to the name of the container that triggered the event) or if no container name is specified "spec.containers[2]" (container with index 2 in this pod). This syntax is chosen only to have some well-defined way of referencing a part of an object. TODO: this design is not final and this field is subject to change in the future.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          Kind of the referent. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>resourceVersion</b></td>
        <td>string</td>
        <td>
          Specific resourceVersion to which this reference is made, if any. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>uid</b></td>
        <td>string</td>
        <td>
          UID of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#uids<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>

### StreamTransfer
<sup><sup>[↩ Parent](#motionfybrikiov1alpha1 )</sup></sup>






StreamTransfer is the Schema for the streamtransfers API

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>motion.fybrik.io/v1alpha1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>StreamTransfer</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#streamtransferspec">spec</a></b></td>
        <td>object</td>
        <td>
          StreamTransferSpec defines the desired state of StreamTransfer<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#streamtransferstatus">status</a></b></td>
        <td>object</td>
        <td>
          StreamTransferStatus defines the observed state of StreamTransfer<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### StreamTransfer.spec
<sup><sup>[↩ Parent](#streamtransfer)</sup></sup>



StreamTransferSpec defines the desired state of StreamTransfer

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>flowType</b></td>
        <td>enum</td>
        <td>
          Data flow type that specifies if this is a stream or a batch workflow<br/>
          <br/>
            <i>Enum</i>: Batch, Stream<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>image</b></td>
        <td>string</td>
        <td>
          Image that should be used for the actual batch job. This is usually a datamover image. This property will be defaulted by the webhook if not set.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>imagePullPolicy</b></td>
        <td>string</td>
        <td>
          Image pull policy that should be used for the actual job. This property will be defaulted by the webhook if not set.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>noFinalizer</b></td>
        <td>boolean</td>
        <td>
          If this batch job instance should have a finalizer or not. This property will be defaulted by the webhook if not set.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readDataType</b></td>
        <td>enum</td>
        <td>
          Data type of the data that is read from source (log data or change data)<br/>
          <br/>
            <i>Enum</i>: LogData, ChangeData<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>secretProviderRole</b></td>
        <td>string</td>
        <td>
          Secret provider role that should be used for the actual job. This property will be defaulted by the webhook if not set.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>secretProviderURL</b></td>
        <td>string</td>
        <td>
          Secret provider url that should be used for the actual job. This property will be defaulted by the webhook if not set.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>suspend</b></td>
        <td>boolean</td>
        <td>
          If this batch job instance is run on a schedule the regular schedule can be suspended with this property. This property will be defaulted by the webhook if not set.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#streamtransferspectransformationindex">transformation</a></b></td>
        <td>[]object</td>
        <td>
          Transformations to be applied to the source data before writing to destination<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>triggerInterval</b></td>
        <td>string</td>
        <td>
          Interval in which the Micro batches of this stream should be triggered The default is '5 seconds'.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>writeDataType</b></td>
        <td>enum</td>
        <td>
          Data type of how the data should be written to the target (log data or change data)<br/>
          <br/>
            <i>Enum</i>: LogData, ChangeData<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>writeOperation</b></td>
        <td>enum</td>
        <td>
          Write operation that should be performed when writing (overwrite,append,update) Caution: Some write operations are only available for batch and some only for stream.<br/>
          <br/>
            <i>Enum</i>: Overwrite, Append, Update<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#streamtransferspecdestination">destination</a></b></td>
        <td>object</td>
        <td>
          Destination data store for this batch job<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#streamtransferspecsource">source</a></b></td>
        <td>object</td>
        <td>
          Source data store for this batch job<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### StreamTransfer.spec.transformation[index]
<sup><sup>[↩ Parent](#streamtransferspec)</sup></sup>



to be refined...

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>action</b></td>
        <td>enum</td>
        <td>
          Transformation action that should be performed.<br/>
          <br/>
            <i>Enum</i>: RemoveColumns, EncryptColumns, DigestColumns, RedactColumns, SampleRows, FilterRows<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>columns</b></td>
        <td>[]string</td>
        <td>
          Columns that are involved in this action. This property is optional as for some actions no columns have to be specified. E.g. filter is a row based transformation.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the transaction. Mainly used for debugging and lineage tracking.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>options</b></td>
        <td>map[string]string</td>
        <td>
          Additional options for this transformation.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### StreamTransfer.spec.destination
<sup><sup>[↩ Parent](#streamtransferspec)</sup></sup>



Destination data store for this batch job

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#streamtransferspecdestinationcloudant">cloudant</a></b></td>
        <td>object</td>
        <td>
          IBM Cloudant. Needs cloudant legacy credentials.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#streamtransferspecdestinationdatabase">database</a></b></td>
        <td>object</td>
        <td>
          Database data store. For the moment only Db2 is supported.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>description</b></td>
        <td>string</td>
        <td>
          Description of the transfer in human readable form that is displayed in the kubectl get If not provided this will be filled in depending on the datastore that is specified.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#streamtransferspecdestinationkafka">kafka</a></b></td>
        <td>object</td>
        <td>
          Kafka data store. The supposed format within the given Kafka topic is a Confluent compatible format stored as Avro. A schema registry needs to be specified as well.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#streamtransferspecdestinations3">s3</a></b></td>
        <td>object</td>
        <td>
          An object store data store that is compatible with S3. This can be a COS bucket.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### StreamTransfer.spec.destination.cloudant
<sup><sup>[↩ Parent](#streamtransferspecdestination)</sup></sup>



IBM Cloudant. Needs cloudant legacy credentials.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>password</b></td>
        <td>string</td>
        <td>
          Cloudant password. Can be retrieved from vault if specified in vault parameter and is thus optional.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>secretImport</b></td>
        <td>string</td>
        <td>
          Define a secret import definition.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>username</b></td>
        <td>string</td>
        <td>
          Cloudant user. Can be retrieved from vault if specified in vault parameter and is thus optional.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#streamtransferspecdestinationcloudantvault">vault</a></b></td>
        <td>object</td>
        <td>
          Define secrets that are fetched from a Vault instance<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>database</b></td>
        <td>string</td>
        <td>
          Database to be read from/written to<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          Host of cloudant instance<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### StreamTransfer.spec.destination.cloudant.vault
<sup><sup>[↩ Parent](#streamtransferspecdestinationcloudant)</sup></sup>



Define secrets that are fetched from a Vault instance

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>address</b></td>
        <td>string</td>
        <td>
          Address is Vault address<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>authPath</b></td>
        <td>string</td>
        <td>
          AuthPath is the path to auth method i.e. kubernetes<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>role</b></td>
        <td>string</td>
        <td>
          Role is the Vault role used for retrieving the credentials<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>secretPath</b></td>
        <td>string</td>
        <td>
          SecretPath is the path of the secret holding the Credentials in Vault<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### StreamTransfer.spec.destination.database
<sup><sup>[↩ Parent](#streamtransferspecdestination)</sup></sup>



Database data store. For the moment only Db2 is supported.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>password</b></td>
        <td>string</td>
        <td>
          Database password. Can be retrieved from vault if specified in vault parameter and is thus optional.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>secretImport</b></td>
        <td>string</td>
        <td>
          Define a secret import definition.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>user</b></td>
        <td>string</td>
        <td>
          Database user. Can be retrieved from vault if specified in vault parameter and is thus optional.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#streamtransferspecdestinationdatabasevault">vault</a></b></td>
        <td>object</td>
        <td>
          Define secrets that are fetched from a Vault instance<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>db2URL</b></td>
        <td>string</td>
        <td>
          URL to Db2 instance in JDBC format Supported SSL certificates are currently certificates signed with IBM Intermediate CA or cloud signed certificates.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>table</b></td>
        <td>string</td>
        <td>
          Table to be read<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### StreamTransfer.spec.destination.database.vault
<sup><sup>[↩ Parent](#streamtransferspecdestinationdatabase)</sup></sup>



Define secrets that are fetched from a Vault instance

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>address</b></td>
        <td>string</td>
        <td>
          Address is Vault address<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>authPath</b></td>
        <td>string</td>
        <td>
          AuthPath is the path to auth method i.e. kubernetes<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>role</b></td>
        <td>string</td>
        <td>
          Role is the Vault role used for retrieving the credentials<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>secretPath</b></td>
        <td>string</td>
        <td>
          SecretPath is the path of the secret holding the Credentials in Vault<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### StreamTransfer.spec.destination.kafka
<sup><sup>[↩ Parent](#streamtransferspecdestination)</sup></sup>



Kafka data store. The supposed format within the given Kafka topic is a Confluent compatible format stored as Avro. A schema registry needs to be specified as well.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>createSnapshot</b></td>
        <td>boolean</td>
        <td>
          If a snapshot should be created of the topic. Records in Kafka are stored as key-value pairs. Updates/Deletes for the same key are appended to the Kafka topic and the last value for a given key is the valid key in a Snapshot. When this property is true only the last value will be written. If the property is false all values will be written out. As a CDC example: If the property is true a valid snapshot of the log stream will be created. If the property is false the CDC stream will be dumped as is like a change log.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>dataFormat</b></td>
        <td>string</td>
        <td>
          Data format of the objects in S3. e.g. parquet or csv. Please refer to struct for allowed values.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>keyDeserializer</b></td>
        <td>string</td>
        <td>
          Deserializer to be used for the keys of the topic<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>password</b></td>
        <td>string</td>
        <td>
          Kafka user password Can be retrieved from vault if specified in vault parameter and is thus optional.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>saslMechanism</b></td>
        <td>string</td>
        <td>
          SASL Mechanism to be used (e.g. PLAIN or SCRAM-SHA-512) Default SCRAM-SHA-512 will be assumed if not specified<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>secretImport</b></td>
        <td>string</td>
        <td>
          Define a secret import definition.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>securityProtocol</b></td>
        <td>string</td>
        <td>
          Kafka security protocol one of (PLAINTEXT, SASL_PLAINTEXT, SASL_SSL, SSL) Default SASL_SSL will be assumed if not specified<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sslTruststore</b></td>
        <td>string</td>
        <td>
          A truststore or certificate encoded as base64. The format can be JKS or PKCS12. A truststore can be specified like this or in a predefined Kubernetes secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sslTruststoreLocation</b></td>
        <td>string</td>
        <td>
          SSL truststore location.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sslTruststorePassword</b></td>
        <td>string</td>
        <td>
          SSL truststore password.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sslTruststoreSecret</b></td>
        <td>string</td>
        <td>
          Kubernetes secret that contains the SSL truststore. The format can be JKS or PKCS12. A truststore can be specified like this or as<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>user</b></td>
        <td>string</td>
        <td>
          Kafka user name. Can be retrieved from vault if specified in vault parameter and is thus optional.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>valueDeserializer</b></td>
        <td>string</td>
        <td>
          Deserializer to be used for the values of the topic<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#streamtransferspecdestinationkafkavault">vault</a></b></td>
        <td>object</td>
        <td>
          Define secrets that are fetched from a Vault instance<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>kafkaBrokers</b></td>
        <td>string</td>
        <td>
          Kafka broker URLs as a comma separated list.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>kafkaTopic</b></td>
        <td>string</td>
        <td>
          Kafka topic<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>schemaRegistryURL</b></td>
        <td>string</td>
        <td>
          URL to the schema registry. The registry has to be Confluent schema registry compatible.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### StreamTransfer.spec.destination.kafka.vault
<sup><sup>[↩ Parent](#streamtransferspecdestinationkafka)</sup></sup>



Define secrets that are fetched from a Vault instance

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>address</b></td>
        <td>string</td>
        <td>
          Address is Vault address<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>authPath</b></td>
        <td>string</td>
        <td>
          AuthPath is the path to auth method i.e. kubernetes<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>role</b></td>
        <td>string</td>
        <td>
          Role is the Vault role used for retrieving the credentials<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>secretPath</b></td>
        <td>string</td>
        <td>
          SecretPath is the path of the secret holding the Credentials in Vault<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### StreamTransfer.spec.destination.s3
<sup><sup>[↩ Parent](#streamtransferspecdestination)</sup></sup>



An object store data store that is compatible with S3. This can be a COS bucket.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>accessKey</b></td>
        <td>string</td>
        <td>
          Access key of the HMAC credentials that can access the given bucket. Can be retrieved from vault if specified in vault parameter and is thus optional.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>dataFormat</b></td>
        <td>string</td>
        <td>
          Data format of the objects in S3. e.g. parquet or csv. Please refer to struct for allowed values.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>partitionBy</b></td>
        <td>[]string</td>
        <td>
          Partition by partition (for target data stores) Defines the columns to partition the output by for a target data store.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>region</b></td>
        <td>string</td>
        <td>
          Region of S3 service<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>secretImport</b></td>
        <td>string</td>
        <td>
          Define a secret import definition.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>secretKey</b></td>
        <td>string</td>
        <td>
          Secret key of the HMAC credentials that can access the given bucket. Can be retrieved from vault if specified in vault parameter and is thus optional.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#streamtransferspecdestinations3vault">vault</a></b></td>
        <td>object</td>
        <td>
          Define secrets that are fetched from a Vault instance<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>bucket</b></td>
        <td>string</td>
        <td>
          Bucket of S3 service<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>endpoint</b></td>
        <td>string</td>
        <td>
          Endpoint of S3 service<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>objectKey</b></td>
        <td>string</td>
        <td>
          Object key of the object in S3. This is used as a prefix! Thus all objects that have the given objectKey as prefix will be used as input!<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### StreamTransfer.spec.destination.s3.vault
<sup><sup>[↩ Parent](#streamtransferspecdestinations3)</sup></sup>



Define secrets that are fetched from a Vault instance

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>address</b></td>
        <td>string</td>
        <td>
          Address is Vault address<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>authPath</b></td>
        <td>string</td>
        <td>
          AuthPath is the path to auth method i.e. kubernetes<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>role</b></td>
        <td>string</td>
        <td>
          Role is the Vault role used for retrieving the credentials<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>secretPath</b></td>
        <td>string</td>
        <td>
          SecretPath is the path of the secret holding the Credentials in Vault<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### StreamTransfer.spec.source
<sup><sup>[↩ Parent](#streamtransferspec)</sup></sup>



Source data store for this batch job

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#streamtransferspecsourcecloudant">cloudant</a></b></td>
        <td>object</td>
        <td>
          IBM Cloudant. Needs cloudant legacy credentials.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#streamtransferspecsourcedatabase">database</a></b></td>
        <td>object</td>
        <td>
          Database data store. For the moment only Db2 is supported.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>description</b></td>
        <td>string</td>
        <td>
          Description of the transfer in human readable form that is displayed in the kubectl get If not provided this will be filled in depending on the datastore that is specified.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#streamtransferspecsourcekafka">kafka</a></b></td>
        <td>object</td>
        <td>
          Kafka data store. The supposed format within the given Kafka topic is a Confluent compatible format stored as Avro. A schema registry needs to be specified as well.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#streamtransferspecsources3">s3</a></b></td>
        <td>object</td>
        <td>
          An object store data store that is compatible with S3. This can be a COS bucket.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### StreamTransfer.spec.source.cloudant
<sup><sup>[↩ Parent](#streamtransferspecsource)</sup></sup>



IBM Cloudant. Needs cloudant legacy credentials.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>password</b></td>
        <td>string</td>
        <td>
          Cloudant password. Can be retrieved from vault if specified in vault parameter and is thus optional.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>secretImport</b></td>
        <td>string</td>
        <td>
          Define a secret import definition.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>username</b></td>
        <td>string</td>
        <td>
          Cloudant user. Can be retrieved from vault if specified in vault parameter and is thus optional.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#streamtransferspecsourcecloudantvault">vault</a></b></td>
        <td>object</td>
        <td>
          Define secrets that are fetched from a Vault instance<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>database</b></td>
        <td>string</td>
        <td>
          Database to be read from/written to<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          Host of cloudant instance<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### StreamTransfer.spec.source.cloudant.vault
<sup><sup>[↩ Parent](#streamtransferspecsourcecloudant)</sup></sup>



Define secrets that are fetched from a Vault instance

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>address</b></td>
        <td>string</td>
        <td>
          Address is Vault address<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>authPath</b></td>
        <td>string</td>
        <td>
          AuthPath is the path to auth method i.e. kubernetes<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>role</b></td>
        <td>string</td>
        <td>
          Role is the Vault role used for retrieving the credentials<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>secretPath</b></td>
        <td>string</td>
        <td>
          SecretPath is the path of the secret holding the Credentials in Vault<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### StreamTransfer.spec.source.database
<sup><sup>[↩ Parent](#streamtransferspecsource)</sup></sup>



Database data store. For the moment only Db2 is supported.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>password</b></td>
        <td>string</td>
        <td>
          Database password. Can be retrieved from vault if specified in vault parameter and is thus optional.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>secretImport</b></td>
        <td>string</td>
        <td>
          Define a secret import definition.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>user</b></td>
        <td>string</td>
        <td>
          Database user. Can be retrieved from vault if specified in vault parameter and is thus optional.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#streamtransferspecsourcedatabasevault">vault</a></b></td>
        <td>object</td>
        <td>
          Define secrets that are fetched from a Vault instance<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>db2URL</b></td>
        <td>string</td>
        <td>
          URL to Db2 instance in JDBC format Supported SSL certificates are currently certificates signed with IBM Intermediate CA or cloud signed certificates.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>table</b></td>
        <td>string</td>
        <td>
          Table to be read<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### StreamTransfer.spec.source.database.vault
<sup><sup>[↩ Parent](#streamtransferspecsourcedatabase)</sup></sup>



Define secrets that are fetched from a Vault instance

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>address</b></td>
        <td>string</td>
        <td>
          Address is Vault address<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>authPath</b></td>
        <td>string</td>
        <td>
          AuthPath is the path to auth method i.e. kubernetes<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>role</b></td>
        <td>string</td>
        <td>
          Role is the Vault role used for retrieving the credentials<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>secretPath</b></td>
        <td>string</td>
        <td>
          SecretPath is the path of the secret holding the Credentials in Vault<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### StreamTransfer.spec.source.kafka
<sup><sup>[↩ Parent](#streamtransferspecsource)</sup></sup>



Kafka data store. The supposed format within the given Kafka topic is a Confluent compatible format stored as Avro. A schema registry needs to be specified as well.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>createSnapshot</b></td>
        <td>boolean</td>
        <td>
          If a snapshot should be created of the topic. Records in Kafka are stored as key-value pairs. Updates/Deletes for the same key are appended to the Kafka topic and the last value for a given key is the valid key in a Snapshot. When this property is true only the last value will be written. If the property is false all values will be written out. As a CDC example: If the property is true a valid snapshot of the log stream will be created. If the property is false the CDC stream will be dumped as is like a change log.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>dataFormat</b></td>
        <td>string</td>
        <td>
          Data format of the objects in S3. e.g. parquet or csv. Please refer to struct for allowed values.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>keyDeserializer</b></td>
        <td>string</td>
        <td>
          Deserializer to be used for the keys of the topic<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>password</b></td>
        <td>string</td>
        <td>
          Kafka user password Can be retrieved from vault if specified in vault parameter and is thus optional.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>saslMechanism</b></td>
        <td>string</td>
        <td>
          SASL Mechanism to be used (e.g. PLAIN or SCRAM-SHA-512) Default SCRAM-SHA-512 will be assumed if not specified<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>secretImport</b></td>
        <td>string</td>
        <td>
          Define a secret import definition.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>securityProtocol</b></td>
        <td>string</td>
        <td>
          Kafka security protocol one of (PLAINTEXT, SASL_PLAINTEXT, SASL_SSL, SSL) Default SASL_SSL will be assumed if not specified<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sslTruststore</b></td>
        <td>string</td>
        <td>
          A truststore or certificate encoded as base64. The format can be JKS or PKCS12. A truststore can be specified like this or in a predefined Kubernetes secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sslTruststoreLocation</b></td>
        <td>string</td>
        <td>
          SSL truststore location.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sslTruststorePassword</b></td>
        <td>string</td>
        <td>
          SSL truststore password.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sslTruststoreSecret</b></td>
        <td>string</td>
        <td>
          Kubernetes secret that contains the SSL truststore. The format can be JKS or PKCS12. A truststore can be specified like this or as<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>user</b></td>
        <td>string</td>
        <td>
          Kafka user name. Can be retrieved from vault if specified in vault parameter and is thus optional.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>valueDeserializer</b></td>
        <td>string</td>
        <td>
          Deserializer to be used for the values of the topic<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#streamtransferspecsourcekafkavault">vault</a></b></td>
        <td>object</td>
        <td>
          Define secrets that are fetched from a Vault instance<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>kafkaBrokers</b></td>
        <td>string</td>
        <td>
          Kafka broker URLs as a comma separated list.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>kafkaTopic</b></td>
        <td>string</td>
        <td>
          Kafka topic<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>schemaRegistryURL</b></td>
        <td>string</td>
        <td>
          URL to the schema registry. The registry has to be Confluent schema registry compatible.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### StreamTransfer.spec.source.kafka.vault
<sup><sup>[↩ Parent](#streamtransferspecsourcekafka)</sup></sup>



Define secrets that are fetched from a Vault instance

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>address</b></td>
        <td>string</td>
        <td>
          Address is Vault address<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>authPath</b></td>
        <td>string</td>
        <td>
          AuthPath is the path to auth method i.e. kubernetes<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>role</b></td>
        <td>string</td>
        <td>
          Role is the Vault role used for retrieving the credentials<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>secretPath</b></td>
        <td>string</td>
        <td>
          SecretPath is the path of the secret holding the Credentials in Vault<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### StreamTransfer.spec.source.s3
<sup><sup>[↩ Parent](#streamtransferspecsource)</sup></sup>



An object store data store that is compatible with S3. This can be a COS bucket.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>accessKey</b></td>
        <td>string</td>
        <td>
          Access key of the HMAC credentials that can access the given bucket. Can be retrieved from vault if specified in vault parameter and is thus optional.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>dataFormat</b></td>
        <td>string</td>
        <td>
          Data format of the objects in S3. e.g. parquet or csv. Please refer to struct for allowed values.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>partitionBy</b></td>
        <td>[]string</td>
        <td>
          Partition by partition (for target data stores) Defines the columns to partition the output by for a target data store.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>region</b></td>
        <td>string</td>
        <td>
          Region of S3 service<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>secretImport</b></td>
        <td>string</td>
        <td>
          Define a secret import definition.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>secretKey</b></td>
        <td>string</td>
        <td>
          Secret key of the HMAC credentials that can access the given bucket. Can be retrieved from vault if specified in vault parameter and is thus optional.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#streamtransferspecsources3vault">vault</a></b></td>
        <td>object</td>
        <td>
          Define secrets that are fetched from a Vault instance<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>bucket</b></td>
        <td>string</td>
        <td>
          Bucket of S3 service<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>endpoint</b></td>
        <td>string</td>
        <td>
          Endpoint of S3 service<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>objectKey</b></td>
        <td>string</td>
        <td>
          Object key of the object in S3. This is used as a prefix! Thus all objects that have the given objectKey as prefix will be used as input!<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### StreamTransfer.spec.source.s3.vault
<sup><sup>[↩ Parent](#streamtransferspecsources3)</sup></sup>



Define secrets that are fetched from a Vault instance

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>address</b></td>
        <td>string</td>
        <td>
          Address is Vault address<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>authPath</b></td>
        <td>string</td>
        <td>
          AuthPath is the path to auth method i.e. kubernetes<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>role</b></td>
        <td>string</td>
        <td>
          Role is the Vault role used for retrieving the credentials<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>secretPath</b></td>
        <td>string</td>
        <td>
          SecretPath is the path of the secret holding the Credentials in Vault<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### StreamTransfer.status
<sup><sup>[↩ Parent](#streamtransfer)</sup></sup>



StreamTransferStatus defines the observed state of StreamTransfer

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#streamtransferstatusactive">active</a></b></td>
        <td>object</td>
        <td>
          A pointer to the currently running job (or nil)<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>error</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          <br/>
          <br/>
            <i>Enum</i>: STARTING, RUNNING, STOPPED, FAILING<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### StreamTransfer.status.active
<sup><sup>[↩ Parent](#streamtransferstatus)</sup></sup>



A pointer to the currently running job (or nil)

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>apiVersion</b></td>
        <td>string</td>
        <td>
          API version of the referent.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>fieldPath</b></td>
        <td>string</td>
        <td>
          If referring to a piece of an object instead of an entire object, this string should contain a valid JSON/Go field access statement, such as desiredState.manifest.containers[2]. For example, if the object reference is to a container within a pod, this would take on a value like: "spec.containers{name}" (where "name" refers to the name of the container that triggered the event) or if no container name is specified "spec.containers[2]" (container with index 2 in this pod). This syntax is chosen only to have some well-defined way of referencing a part of an object. TODO: this design is not final and this field is subject to change in the future.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          Kind of the referent. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>resourceVersion</b></td>
        <td>string</td>
        <td>
          Specific resourceVersion to which this reference is made, if any. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>uid</b></td>
        <td>string</td>
        <td>
          UID of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#uids<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>
