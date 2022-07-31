---
hide:
  - toc        # Hide table of contents
---

# API Reference

Packages:

- [app.fybrik.io/v1](#appfybrikiov1)
- [katalog.fybrik.io/v1alpha1](#katalogfybrikiov1alpha1)

## app.fybrik.io/v1

Resource Types:

- [Blueprint](#blueprint)

- [FybrikApplication](#fybrikapplication)

- [FybrikModule](#fybrikmodule)

- [FybrikStorageAccount](#fybrikstorageaccount)

- [Plotter](#plotter)




### Blueprint
<sup><sup>[↩ Parent](#appfybrikiov1 )</sup></sup>






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
      <td>app.fybrik.io/v1</td>
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
          BlueprintSpec defines the desired state of Blueprint, which defines the components of the workload's data path that run in a particular cluster. In a single cluster environment there is one blueprint per workload (FybrikApplication). In a multi-cluster environment there is one Blueprint per cluster per workload (FybrikApplication).<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#blueprintstatus">status</a></b></td>
        <td>object</td>
        <td>
          BlueprintStatus defines the observed state of Blueprint This includes readiness, error message, and indicators for the Kubernetes resources owned by the Blueprint for cleanup and status monitoring<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### Blueprint.spec
<sup><sup>[↩ Parent](#blueprint)</sup></sup>



BlueprintSpec defines the desired state of Blueprint, which defines the components of the workload's data path that run in a particular cluster. In a single cluster environment there is one blueprint per workload (FybrikApplication). In a multi-cluster environment there is one Blueprint per cluster per workload (FybrikApplication).

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
        <td><b>cluster</b></td>
        <td>string</td>
        <td>
          Cluster indicates the cluster on which the Blueprint runs<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#blueprintspecmoduleskey">modules</a></b></td>
        <td>map[string]object</td>
        <td>
          Modules is a map which contains modules that indicate the data path components that run in this cluster The map key is moduleInstanceName which is the unique name for the deployed instance related to this workload<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>modulesNamespace</b></td>
        <td>string</td>
        <td>
          ModulesNamespace is the namespace where modules should be allocated<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#blueprintspecapplication">application</a></b></td>
        <td>object</td>
        <td>
          ApplicationContext is a context of the origin FybrikApplication (labels, properties, etc.)<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### Blueprint.spec.modules[key]
<sup><sup>[↩ Parent](#blueprintspec)</sup></sup>



BlueprintModule is a copy of a FybrikModule Custom Resource.  It contains the information necessary to instantiate a datapath component, including the parameters relevant for the particular workload.

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
        <td><b><a href="#blueprintspecmoduleskeychart">chart</a></b></td>
        <td>object</td>
        <td>
          Chart contains the location of the helm chart with info detailing how to deploy<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the FybrikModule on which this is based<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#blueprintspecmoduleskeyarguments">arguments</a></b></td>
        <td>object</td>
        <td>
          Arguments are the input parameters for a specific instance of a module.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>assetIds</b></td>
        <td>[]string</td>
        <td>
          assetIDs indicate the assets processed by this module.  Included so we can track asset status as well as module status in the future.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### Blueprint.spec.modules[key].chart
<sup><sup>[↩ Parent](#blueprintspecmoduleskey)</sup></sup>



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
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of helm chart<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>chartPullSecret</b></td>
        <td>string</td>
        <td>
          Name of secret containing helm registry credentials<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>map[string]string</td>
        <td>
          Values to pass to helm chart installation<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### Blueprint.spec.modules[key].arguments
<sup><sup>[↩ Parent](#blueprintspecmoduleskey)</sup></sup>



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
        <td><b><a href="#blueprintspecmoduleskeyargumentsassetsindex">assets</a></b></td>
        <td>[]object</td>
        <td>
          Assets define asset related arguments, such as data source, transformations, etc.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### Blueprint.spec.modules[key].arguments.assets[index]
<sup><sup>[↩ Parent](#blueprintspecmoduleskeyarguments)</sup></sup>



AssetContext defines the input parameters for modules that access an asset

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
        <td><b>assetID</b></td>
        <td>string</td>
        <td>
          AssetID identifies the asset to be used for accessing the data when it is ready It is copied from the FybrikApplication resource<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>capability</b></td>
        <td>string</td>
        <td>
          Capability of the module<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#blueprintspecmoduleskeyargumentsassetsindexargsindex">args</a></b></td>
        <td>[]object</td>
        <td>
          List of datastores associated with the asset<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#blueprintspecmoduleskeyargumentsassetsindextransformationsindex">transformations</a></b></td>
        <td>[]object</td>
        <td>
          Transformations are different types of processing that may be done to the data as it is copied.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### Blueprint.spec.modules[key].arguments.assets[index].args[index]
<sup><sup>[↩ Parent](#blueprintspecmoduleskeyargumentsassetsindex)</sup></sup>



DataStore contains the details for accesing the data that are sent by catalog connectors Credentials for accesing the data are stored in Vault, in the location represented by Vault property.

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
        <td><b><a href="#blueprintspecmoduleskeyargumentsassetsindexargsindexconnection">connection</a></b></td>
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
        <td>false</td>
      </tr><tr>
        <td><b><a href="#blueprintspecmoduleskeyargumentsassetsindexargsindexvaultkey">vault</a></b></td>
        <td>map[string]object</td>
        <td>
          Holds details for retrieving credentials by the modules from Vault store. It is a map so that different credentials can be stored for the different DataFlow operations.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### Blueprint.spec.modules[key].arguments.assets[index].args[index].connection
<sup><sup>[↩ Parent](#blueprintspecmoduleskeyargumentsassetsindexargsindex)</sup></sup>



Connection has the relevant details for accesing the data (url, table, ssl, etc.)

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
      </tr></tbody>
</table>


#### Blueprint.spec.modules[key].arguments.assets[index].args[index].vault[key]
<sup><sup>[↩ Parent](#blueprintspecmoduleskeyargumentsassetsindexargsindex)</sup></sup>



Holds details for retrieving credentials from Vault store.

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


#### Blueprint.spec.modules[key].arguments.assets[index].transformations[index]
<sup><sup>[↩ Parent](#blueprintspecmoduleskeyargumentsassetsindex)</sup></sup>





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
      </tr></tbody>
</table>


#### Blueprint.spec.application
<sup><sup>[↩ Parent](#blueprintspec)</sup></sup>



ApplicationContext is a context of the origin FybrikApplication (labels, properties, etc.)

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
        <td><b>context</b></td>
        <td>object</td>
        <td>
          Application context such as intent, role, etc.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#blueprintspecapplicationselector">selector</a></b></td>
        <td>object</td>
        <td>
          Application selector is used to identify the user workload. It is obtained from FybrikApplication spec.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### Blueprint.spec.application.selector
<sup><sup>[↩ Parent](#blueprintspecapplication)</sup></sup>



Application selector is used to identify the user workload. It is obtained from FybrikApplication spec.

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
        <td><b><a href="#blueprintspecapplicationselectormatchexpressionsindex">matchExpressions</a></b></td>
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


#### Blueprint.spec.application.selector.matchExpressions[index]
<sup><sup>[↩ Parent](#blueprintspecapplicationselector)</sup></sup>



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
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### Blueprint.status
<sup><sup>[↩ Parent](#blueprint)</sup></sup>



BlueprintStatus defines the observed state of Blueprint This includes readiness, error message, and indicators for the Kubernetes resources owned by the Blueprint for cleanup and status monitoring

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
        <td><b><a href="#blueprintstatusmoduleskey">modules</a></b></td>
        <td>map[string]object</td>
        <td>
          ModulesState is a map which holds the status of each module its key is the moduleInstanceName which is the unique name for the deployed instance related to this workload<br/>
        </td>
        <td>false</td>
      </tr><tr>
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


#### Blueprint.status.modules[key]
<sup><sup>[↩ Parent](#blueprintstatus)</sup></sup>



ObservedState represents a part of the generated Blueprint/Plotter resource status that allows update of FybrikApplication status

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
<sup><sup>[↩ Parent](#appfybrikiov1 )</sup></sup>






FybrikApplication provides information about the application whose data is being operated on, the nature of the processing, and the data sets chosen for processing by the application. The FybrikApplication controller obtains instructions regarding any governance related changes that must be performed on the data, identifies the modules capable of performing such changes, and finally generates the Plotter which defines the secure runtime environment and all the components in it.  This runtime environment provides the application with access to the data requested in a secure manner and without having to provide any credentials for the data sets.  The credentials are obtained automatically by the manager from the credential management system.

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
      <td>app.fybrik.io/v1</td>
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
          FybrikApplicationSpec defines data flows needed by the application, the purpose and other contextual information about the application.<br/>
        </td>
        <td>true</td>
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



FybrikApplicationSpec defines data flows needed by the application, the purpose and other contextual information about the application.

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
        <td><b>appInfo</b></td>
        <td>object</td>
        <td>
          AppInfo contains information describing the reasons for the processing that will be done by the application.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#fybrikapplicationspecdataindex">data</a></b></td>
        <td>[]object</td>
        <td>
          Data contains the identifiers of the data to be used by the Data Scientist's application, and the protocol used to access it and the format expected.<br/>
        </td>
        <td>true</td>
      </tr><tr>
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
          Selector enables to connect the resource to the application Application labels should match the labels in the selector.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### FybrikApplication.spec.data[index]
<sup><sup>[↩ Parent](#fybrikapplicationspec)</sup></sup>



DataContext indicates data set being processed by the workload and includes information about the data format and technologies used to access the data.

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
        <td><b>dataSetID</b></td>
        <td>string</td>
        <td>
          DataSetID is a unique identifier of the dataset chosen from the data catalog. For data catalogs that support multiple sub-catalogs, it includes the catalog id and the dataset id. When writing a new dataset it is the name provided by the user or workload generating it.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#fybrikapplicationspecdataindexrequirements">requirements</a></b></td>
        <td>object</td>
        <td>
          Requirements from the system<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>flow</b></td>
        <td>enum</td>
        <td>
          Flows indicates what is being done with the particular dataset - ex: read, write, copy (ingest), delete This is optional for the purpose of backward compatibility. If nothing is provided, read is assumed.<br/>
          <br/>
            <i>Enum</i>: read, write, delete, copy<br/>
        </td>
        <td>false</td>
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
        <td><b><a href="#fybrikapplicationspecdataindexrequirementsflowparams">flowParams</a></b></td>
        <td>object</td>
        <td>
          FlowParams include the requirements for particular data flows<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#fybrikapplicationspecdataindexrequirementsinterface">interface</a></b></td>
        <td>object</td>
        <td>
          Interface indicates the protocol and format expected by the data user<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### FybrikApplication.spec.data[index].requirements.flowParams
<sup><sup>[↩ Parent](#fybrikapplicationspecdataindexrequirements)</sup></sup>



FlowParams include the requirements for particular data flows

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
        <td><b>catalog</b></td>
        <td>string</td>
        <td>
          Catalog indicates that the data asset must be cataloged, and in which catalog to register it<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>isNewDataSet</b></td>
        <td>boolean</td>
        <td>
          IsNewDataSet if true indicates that the DataContext.DataSetID is user provided and not a full catalog / dataset ID. Relevant when writing. A unique ID from the catalog will be provided in the FybrikApplication Status after a new catalog entry is created.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#fybrikapplicationspecdataindexrequirementsflowparamsmetadata">metadata</a></b></td>
        <td>object</td>
        <td>
          Source asset metadata like asset name, owner, geography, etc Relevant when writing new asset.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>storageEstimate</b></td>
        <td>integer</td>
        <td>
          Storage estimate indicates the estimated amount of storage in MB, GB, TB required when writing new data.<br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### FybrikApplication.spec.data[index].requirements.flowParams.metadata
<sup><sup>[↩ Parent](#fybrikapplicationspecdataindexrequirementsflowparams)</sup></sup>



Source asset metadata like asset name, owner, geography, etc Relevant when writing new asset.

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
        <td><b><a href="#fybrikapplicationspecdataindexrequirementsflowparamsmetadatacolumnsindex">columns</a></b></td>
        <td>[]object</td>
        <td>
          Columns associated with the asset<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>geography</b></td>
        <td>string</td>
        <td>
          Geography of the resource<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the resource<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>owner</b></td>
        <td>string</td>
        <td>
          Owner of the resource<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>tags</b></td>
        <td>object</td>
        <td>
          Tags associated with the asset<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### FybrikApplication.spec.data[index].requirements.flowParams.metadata.columns[index]
<sup><sup>[↩ Parent](#fybrikapplicationspecdataindexrequirementsflowparamsmetadata)</sup></sup>



ResourceColumn represents a column in a tabular resource

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
          Name of the column<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>tags</b></td>
        <td>object</td>
        <td>
          Tags associated with the column<br/>
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
        <td><b>protocol</b></td>
        <td>string</td>
        <td>
          Protocol defines the interface protocol used for data transactions<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>dataformat</b></td>
        <td>string</td>
        <td>
          DataFormat defines the data format type<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### FybrikApplication.spec.selector
<sup><sup>[↩ Parent](#fybrikapplicationspec)</sup></sup>



Selector enables to connect the resource to the application Application labels should match the labels in the selector.

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
        <td><b><a href="#fybrikapplicationspecselectorworkloadselector">workloadSelector</a></b></td>
        <td>object</td>
        <td>
          WorkloadSelector enables to connect the resource to the application Application labels should match the labels in the selector.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>clusterName</b></td>
        <td>string</td>
        <td>
          Cluster name<br/>
        </td>
        <td>false</td>
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
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.<br/>
        </td>
        <td>false</td>
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
        <td><b><a href="#fybrikapplicationstatusassetstateskey">assetStates</a></b></td>
        <td>map[string]object</td>
        <td>
          AssetStates provides a status per asset<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>errorMessage</b></td>
        <td>string</td>
        <td>
          ErrorMessage indicates that an error has happened during the reconcile, unrelated to a specific asset<br/>
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
        <td><b>ready</b></td>
        <td>boolean</td>
        <td>
          Ready is true if all specified assets are either ready to be used or are denied access.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>validApplication</b></td>
        <td>string</td>
        <td>
          ValidApplication indicates whether the FybrikApplication is valid given the defined taxonomy<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>validatedGeneration</b></td>
        <td>integer</td>
        <td>
          ValidatedGeneration is the version of the FyrbikApplication that has been validated with the taxonomy defined.<br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### FybrikApplication.status.assetStates[key]
<sup><sup>[↩ Parent](#fybrikapplicationstatus)</sup></sup>



AssetState defines the observed state of an asset

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
        <td><b>catalogedAsset</b></td>
        <td>string</td>
        <td>
          CatalogedAsset provides a new asset identifier after being registered in the enterprise catalog<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#fybrikapplicationstatusassetstateskeyconditionsindex">conditions</a></b></td>
        <td>[]object</td>
        <td>
          Conditions indicate the asset state (Ready, Deny, Error)<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#fybrikapplicationstatusassetstateskeyendpoint">endpoint</a></b></td>
        <td>object</td>
        <td>
          Endpoint provides the endpoint spec from which the asset will be served to the application<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### FybrikApplication.status.assetStates[key].conditions[index]
<sup><sup>[↩ Parent](#fybrikapplicationstatusassetstateskey)</sup></sup>



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
        <td><b>type</b></td>
        <td>string</td>
        <td>
          Type of the condition<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          Message contains the details of the current condition<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          ObservedGeneration is the version of the resource for which the condition has been evaluated<br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          Status of the condition, one of (`True`, `False`, `Unknown`).<br/>
          <br/>
            <i>Enum</i>: True, False, Unknown<br/>
            <i>Default</i>: Unknown<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### FybrikApplication.status.assetStates[key].endpoint
<sup><sup>[↩ Parent](#fybrikapplicationstatusassetstateskey)</sup></sup>



Endpoint provides the endpoint spec from which the asset will be served to the application

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



DatasetDetails holds details of the provisioned storage

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
        <td><b><a href="#fybrikapplicationstatusprovisionedstoragekeydetails">details</a></b></td>
        <td>object</td>
        <td>
          Dataset information<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#fybrikapplicationstatusprovisionedstoragekeyresourcemetadata">resourceMetadata</a></b></td>
        <td>object</td>
        <td>
          Resource Metadata<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#fybrikapplicationstatusprovisionedstoragekeysecretref">secretRef</a></b></td>
        <td>object</td>
        <td>
          Reference to a secret where the credentials are stored<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### FybrikApplication.status.provisionedStorage[key].details
<sup><sup>[↩ Parent](#fybrikapplicationstatusprovisionedstoragekey)</sup></sup>



Dataset information

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
        <td><b><a href="#fybrikapplicationstatusprovisionedstoragekeydetailsconnection">connection</a></b></td>
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
        <td>false</td>
      </tr><tr>
        <td><b><a href="#fybrikapplicationstatusprovisionedstoragekeydetailsvaultkey">vault</a></b></td>
        <td>map[string]object</td>
        <td>
          Holds details for retrieving credentials by the modules from Vault store. It is a map so that different credentials can be stored for the different DataFlow operations.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### FybrikApplication.status.provisionedStorage[key].details.connection
<sup><sup>[↩ Parent](#fybrikapplicationstatusprovisionedstoragekeydetails)</sup></sup>



Connection has the relevant details for accesing the data (url, table, ssl, etc.)

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
      </tr></tbody>
</table>


#### FybrikApplication.status.provisionedStorage[key].details.vault[key]
<sup><sup>[↩ Parent](#fybrikapplicationstatusprovisionedstoragekeydetails)</sup></sup>



Holds details for retrieving credentials from Vault store.

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


#### FybrikApplication.status.provisionedStorage[key].resourceMetadata
<sup><sup>[↩ Parent](#fybrikapplicationstatusprovisionedstoragekey)</sup></sup>



Resource Metadata

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
        <td><b><a href="#fybrikapplicationstatusprovisionedstoragekeyresourcemetadatacolumnsindex">columns</a></b></td>
        <td>[]object</td>
        <td>
          Columns associated with the asset<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>geography</b></td>
        <td>string</td>
        <td>
          Geography of the resource<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the resource<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>owner</b></td>
        <td>string</td>
        <td>
          Owner of the resource<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>tags</b></td>
        <td>object</td>
        <td>
          Tags associated with the asset<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### FybrikApplication.status.provisionedStorage[key].resourceMetadata.columns[index]
<sup><sup>[↩ Parent](#fybrikapplicationstatusprovisionedstoragekeyresourcemetadata)</sup></sup>



ResourceColumn represents a column in a tabular resource

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
          Name of the column<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>tags</b></td>
        <td>object</td>
        <td>
          Tags associated with the column<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### FybrikApplication.status.provisionedStorage[key].secretRef
<sup><sup>[↩ Parent](#fybrikapplicationstatusprovisionedstoragekey)</sup></sup>



Reference to a secret where the credentials are stored

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
          Secret name<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Secret Namespace<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>

### FybrikModule
<sup><sup>[↩ Parent](#appfybrikiov1 )</sup></sup>






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
      <td>app.fybrik.io/v1</td>
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
      </tr><tr>
        <td><b><a href="#fybrikmodulestatus">status</a></b></td>
        <td>object</td>
        <td>
          FybrikModuleStatus defines the observed state of FybrikModule.<br/>
        </td>
        <td>false</td>
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
        <td><b><a href="#fybrikmodulespeccapabilitiesindex">capabilities</a></b></td>
        <td>[]object</td>
        <td>
          Capabilities declares what this module knows how to do and the types of data it knows how to handle The key to the map is a CapabilityType string<br/>
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
        <td><b>type</b></td>
        <td>string</td>
        <td>
          May be one of service, config or plugin Service: Means that the control plane deploys the component that performs the capability Config: Another pre-installed service performs the capability and the module deployed configures it for the particular workload or dataset Plugin: Indicates that this module performs a capability as part of another service or module rather than as a stand-alone module<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#fybrikmodulespecdependenciesindex">dependencies</a></b></td>
        <td>[]object</td>
        <td>
          Other components that must be installed in order for this module to work<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>description</b></td>
        <td>string</td>
        <td>
          An explanation of what this module does<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>pluginType</b></td>
        <td>string</td>
        <td>
          Plugin type indicates the plugin technology used to invoke the capabilities Ex: vault, fybrik-wasm... Should be provided if type is plugin<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#fybrikmodulespecstatusindicatorsindex">statusIndicators</a></b></td>
        <td>[]object</td>
        <td>
          StatusIndicators allow to check status of a non-standard resource that can not be computed by helm/kstatus<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### FybrikModule.spec.capabilities[index]
<sup><sup>[↩ Parent](#fybrikmodulespec)</sup></sup>



Capability declares what this module knows how to do and the types of data it knows how to handle

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
        <td><b>capability</b></td>
        <td>string</td>
        <td>
          Capability declares what this module knows how to do - ex: read, write, transform...<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#fybrikmodulespeccapabilitiesindexactionsindex">actions</a></b></td>
        <td>[]object</td>
        <td>
          Actions are the data transformations that the module supports<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#fybrikmodulespeccapabilitiesindexapi">api</a></b></td>
        <td>object</td>
        <td>
          API indicates to the application how to access the capabilities provided by the module<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#fybrikmodulespeccapabilitiesindexpluginsindex">plugins</a></b></td>
        <td>[]object</td>
        <td>
          Plugins enable the module to add libraries to perform actions rather than implementing them by itself<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>scope</b></td>
        <td>enum</td>
        <td>
          Scope indicates at what level the capability is used: workload, asset, cluster If not indicated it is assumed to be asset<br/>
          <br/>
            <i>Enum</i>: asset, workload, cluster<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#fybrikmodulespeccapabilitiesindexsupportedinterfacesindex">supportedInterfaces</a></b></td>
        <td>[]object</td>
        <td>
          Copy should have one or more instances in the list, and its content should have source and sink Read should have one or more instances in the list, each with source populated Write should have one or more instances in the list, each with sink populated This field may not be required if not handling data<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### FybrikModule.spec.capabilities[index].actions[index]
<sup><sup>[↩ Parent](#fybrikmodulespeccapabilitiesindex)</sup></sup>





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
          Unique name of an action supported by the module<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### FybrikModule.spec.capabilities[index].api
<sup><sup>[↩ Parent](#fybrikmodulespeccapabilitiesindex)</sup></sup>



API indicates to the application how to access the capabilities provided by the module

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
        <td><b><a href="#fybrikmodulespeccapabilitiesindexapiconnection">connection</a></b></td>
        <td>object</td>
        <td>
          Connection information<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>dataFormat</b></td>
        <td>string</td>
        <td>
          Data format<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### FybrikModule.spec.capabilities[index].api.connection
<sup><sup>[↩ Parent](#fybrikmodulespeccapabilitiesindexapi)</sup></sup>



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
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### FybrikModule.spec.capabilities[index].plugins[index]
<sup><sup>[↩ Parent](#fybrikmodulespeccapabilitiesindex)</sup></sup>





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
          DataFormat indicates the format of data the plugin knows how to process<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>pluginType</b></td>
        <td>string</td>
        <td>
          PluginType indicates the technology used for the module and the plugin to interact The values supported should come from the module taxonomy Examples of such mechanisms are vault plugins, wasm, etc<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### FybrikModule.spec.capabilities[index].supportedInterfaces[index]
<sup><sup>[↩ Parent](#fybrikmodulespeccapabilitiesindex)</sup></sup>



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
        <td><b><a href="#fybrikmodulespeccapabilitiesindexsupportedinterfacesindexsink">sink</a></b></td>
        <td>object</td>
        <td>
          Sink specifies the output data protocol and format<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#fybrikmodulespeccapabilitiesindexsupportedinterfacesindexsource">source</a></b></td>
        <td>object</td>
        <td>
          Source specifies the input data protocol and format<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### FybrikModule.spec.capabilities[index].supportedInterfaces[index].sink
<sup><sup>[↩ Parent](#fybrikmodulespeccapabilitiesindexsupportedinterfacesindex)</sup></sup>



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
        <td><b>protocol</b></td>
        <td>string</td>
        <td>
          Protocol defines the interface protocol used for data transactions<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>dataformat</b></td>
        <td>string</td>
        <td>
          DataFormat defines the data format type<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### FybrikModule.spec.capabilities[index].supportedInterfaces[index].source
<sup><sup>[↩ Parent](#fybrikmodulespeccapabilitiesindexsupportedinterfacesindex)</sup></sup>



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
        <td><b>protocol</b></td>
        <td>string</td>
        <td>
          Protocol defines the interface protocol used for data transactions<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>dataformat</b></td>
        <td>string</td>
        <td>
          DataFormat defines the data format type<br/>
        </td>
        <td>false</td>
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
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of helm chart<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>chartPullSecret</b></td>
        <td>string</td>
        <td>
          Name of secret containing helm registry credentials<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>map[string]string</td>
        <td>
          Values to pass to helm chart installation<br/>
        </td>
        <td>false</td>
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
      </tr><tr>
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
      </tr></tbody>
</table>


#### FybrikModule.status
<sup><sup>[↩ Parent](#fybrikmodule)</sup></sup>



FybrikModuleStatus defines the observed state of FybrikModule.

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
        <td><b><a href="#fybrikmodulestatusconditionsindex">conditions</a></b></td>
        <td>[]object</td>
        <td>
          Conditions indicate the module states with respect to validation<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### FybrikModule.status.conditions[index]
<sup><sup>[↩ Parent](#fybrikmodulestatus)</sup></sup>



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
        <td><b>type</b></td>
        <td>string</td>
        <td>
          Type of the condition<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          Message contains the details of the current condition<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          ObservedGeneration is the version of the resource for which the condition has been evaluated<br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          Status of the condition, one of (`True`, `False`, `Unknown`).<br/>
          <br/>
            <i>Enum</i>: True, False, Unknown<br/>
            <i>Default</i>: Unknown<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>

### FybrikStorageAccount
<sup><sup>[↩ Parent](#appfybrikiov1 )</sup></sup>






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
      <td>app.fybrik.io/v1</td>
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
        <td>true</td>
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
          Endpoint for accessing the data<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>id</b></td>
        <td>string</td>
        <td>
          Identification of a storage account<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>region</b></td>
        <td>string</td>
        <td>
          Storage region<br/>
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
<sup><sup>[↩ Parent](#appfybrikiov1 )</sup></sup>






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
      <td>app.fybrik.io/v1</td>
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
          PlotterSpec defines the desired state of Plotter, which is applied in a multi-clustered environment. Plotter declares what needs to be installed and where (as blueprints running on remote clusters) which provides the Data Scientist's application with secure and governed access to the data requested in the FybrikApplication.<br/>
        </td>
        <td>true</td>
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



PlotterSpec defines the desired state of Plotter, which is applied in a multi-clustered environment. Plotter declares what needs to be installed and where (as blueprints running on remote clusters) which provides the Data Scientist's application with secure and governed access to the data requested in the FybrikApplication.

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
        <td><b><a href="#plotterspecassetskey">assets</a></b></td>
        <td>map[string]object</td>
        <td>
          Assets is a map holding information about the assets The key is the assetID<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#plotterspecflowsindex">flows</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>modulesNamespace</b></td>
        <td>string</td>
        <td>
          ModulesNamespace is the namespace where modules should be allocated<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#plotterspectemplateskey">templates</a></b></td>
        <td>map[string]object</td>
        <td>
          Templates is a map holding the templates used in this plotter steps The key is the template name<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>appInfo</b></td>
        <td>object</td>
        <td>
          Application context to be transferred to the modules<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#plotterspecappselector">appSelector</a></b></td>
        <td>object</td>
        <td>
          Selector enables to connect the resource to the application Application labels should match the labels in the selector. For some flows the selector may not be used.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### Plotter.spec.assets[key]
<sup><sup>[↩ Parent](#plotterspec)</sup></sup>



AssetDetails is a list of assets used in the fybrikapplication. In addition to assets declared in fybrikapplication, AssetDetails list also contains assets that are allocated by the control-plane in order to serve fybrikapplication

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
        <td><b><a href="#plotterspecassetskeyassetdetails">assetDetails</a></b></td>
        <td>object</td>
        <td>
          DataStore contains the details for accesing the data that are sent by catalog connectors Credentials for accesing the data are stored in Vault, in the location represented by Vault property.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>advertisedAssetId</b></td>
        <td>string</td>
        <td>
          AdvertisedAssetID links this asset to asset from fybrikapplication and is used by user facing services<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### Plotter.spec.assets[key].assetDetails
<sup><sup>[↩ Parent](#plotterspecassetskey)</sup></sup>



DataStore contains the details for accesing the data that are sent by catalog connectors Credentials for accesing the data are stored in Vault, in the location represented by Vault property.

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
        <td><b><a href="#plotterspecassetskeyassetdetailsconnection">connection</a></b></td>
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
        <td>false</td>
      </tr><tr>
        <td><b><a href="#plotterspecassetskeyassetdetailsvaultkey">vault</a></b></td>
        <td>map[string]object</td>
        <td>
          Holds details for retrieving credentials by the modules from Vault store. It is a map so that different credentials can be stored for the different DataFlow operations.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### Plotter.spec.assets[key].assetDetails.connection
<sup><sup>[↩ Parent](#plotterspecassetskeyassetdetails)</sup></sup>



Connection has the relevant details for accesing the data (url, table, ssl, etc.)

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
      </tr></tbody>
</table>


#### Plotter.spec.assets[key].assetDetails.vault[key]
<sup><sup>[↩ Parent](#plotterspecassetskeyassetdetails)</sup></sup>



Holds details for retrieving credentials from Vault store.

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


#### Plotter.spec.flows[index]
<sup><sup>[↩ Parent](#plotterspec)</sup></sup>



Flows is the list of data flows driven from fybrikapplication: Each element in the list holds the flow of the data requested in fybrikapplication.

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
        <td><b>assetId</b></td>
        <td>string</td>
        <td>
          AssetID indicates the data set being used in this data flow<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>flowType</b></td>
        <td>enum</td>
        <td>
          Type of the flow (e.g. read)<br/>
          <br/>
            <i>Enum</i>: read, write, delete, copy<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the flow<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#plotterspecflowsindexsubflowsindex">subFlows</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Plotter.spec.flows[index].subFlows[index]
<sup><sup>[↩ Parent](#plotterspecflowsindex)</sup></sup>



Subflows is a list of data flows which are originated from the same data asset but are triggered differently (e.g., one upon init trigger and one upon workload trigger)

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
          Type of the flow (e.g. read)<br/>
          <br/>
            <i>Enum</i>: read, write, delete, copy<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the SubFlow<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#plotterspecflowsindexsubflowsindexstepsindexindex">steps</a></b></td>
        <td>[][]object</td>
        <td>
          Steps defines a series of sequential/parallel data flow steps The first dimension represents parallel data flows. The second sequential components within the same parallel data flow.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>triggers</b></td>
        <td>[]enum</td>
        <td>
          Triggers<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Plotter.spec.flows[index].subFlows[index].steps[index][index]
<sup><sup>[↩ Parent](#plotterspecflowsindexsubflowsindex)</sup></sup>



DataFlowStep contains details on a single data flow step

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
        <td><b>cluster</b></td>
        <td>string</td>
        <td>
          Name of the cluster this step is executed on<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the step<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>template</b></td>
        <td>string</td>
        <td>
          Template is the name of the template to execute the step The full details of the template can be extracted from Plotter.spec.templates list field.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#plotterspecflowsindexsubflowsindexstepsindexindexparameters">parameters</a></b></td>
        <td>object</td>
        <td>
          Step parameters TODO why not flatten the parameters into this data flow step<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### Plotter.spec.flows[index].subFlows[index].steps[index][index].parameters
<sup><sup>[↩ Parent](#plotterspecflowsindexsubflowsindexstepsindexindex)</sup></sup>



Step parameters TODO why not flatten the parameters into this data flow step

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
        <td><b><a href="#plotterspecflowsindexsubflowsindexstepsindexindexparametersactionindex">action</a></b></td>
        <td>[]object</td>
        <td>
          Actions are the data transformations that the module supports<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#plotterspecflowsindexsubflowsindexstepsindexindexparametersapi">api</a></b></td>
        <td>object</td>
        <td>
          ResourceDetails includes asset connection details<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#plotterspecflowsindexsubflowsindexstepsindexindexparametersargsindex">args</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### Plotter.spec.flows[index].subFlows[index].steps[index][index].parameters.action[index]
<sup><sup>[↩ Parent](#plotterspecflowsindexsubflowsindexstepsindexindexparameters)</sup></sup>





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
      </tr></tbody>
</table>


#### Plotter.spec.flows[index].subFlows[index].steps[index][index].parameters.api
<sup><sup>[↩ Parent](#plotterspecflowsindexsubflowsindexstepsindexindexparameters)</sup></sup>



ResourceDetails includes asset connection details

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
        <td><b><a href="#plotterspecflowsindexsubflowsindexstepsindexindexparametersapiconnection">connection</a></b></td>
        <td>object</td>
        <td>
          Connection information<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>dataFormat</b></td>
        <td>string</td>
        <td>
          Data format<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### Plotter.spec.flows[index].subFlows[index].steps[index][index].parameters.api.connection
<sup><sup>[↩ Parent](#plotterspecflowsindexsubflowsindexstepsindexindexparametersapi)</sup></sup>



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
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Plotter.spec.flows[index].subFlows[index].steps[index][index].parameters.args[index]
<sup><sup>[↩ Parent](#plotterspecflowsindexsubflowsindexstepsindexindexparameters)</sup></sup>



StepArgument describes a step: it could be assetID or an endpoint of another step

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
        <td><b><a href="#plotterspecflowsindexsubflowsindexstepsindexindexparametersargsindexapi">api</a></b></td>
        <td>object</td>
        <td>
          API holds information for accessing a module instance<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>assetId</b></td>
        <td>string</td>
        <td>
          AssetID identifies the source asset of this step<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### Plotter.spec.flows[index].subFlows[index].steps[index][index].parameters.args[index].api
<sup><sup>[↩ Parent](#plotterspecflowsindexsubflowsindexstepsindexindexparametersargsindex)</sup></sup>



API holds information for accessing a module instance

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
        <td><b><a href="#plotterspecflowsindexsubflowsindexstepsindexindexparametersargsindexapiconnection">connection</a></b></td>
        <td>object</td>
        <td>
          Connection information<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>dataFormat</b></td>
        <td>string</td>
        <td>
          Data format<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### Plotter.spec.flows[index].subFlows[index].steps[index][index].parameters.args[index].api.connection
<sup><sup>[↩ Parent](#plotterspecflowsindexsubflowsindexstepsindexindexparametersargsindexapi)</sup></sup>



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
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Plotter.spec.templates[key]
<sup><sup>[↩ Parent](#plotterspec)</sup></sup>



Template contains basic information about the required modules to serve the fybrikapplication e.g., the module helm chart name.

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
        <td><b><a href="#plotterspectemplateskeymodulesindex">modules</a></b></td>
        <td>[]object</td>
        <td>
          Modules is a list of dependent modules. e.g., if a plugin module is used then the service module is used in should appear first in the modules list of the same template. If the modules list contains more than one module, the first module in the list is referred to as the "primary module" of which all the parameters to this template are sent to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the template<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### Plotter.spec.templates[key].modules[index]
<sup><sup>[↩ Parent](#plotterspectemplateskey)</sup></sup>



ModuleInfo is a copy of FybrikModule Custom Resource.  It contains information to instantiate resource of type FybrikModule.

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
        <td><b>capability</b></td>
        <td>string</td>
        <td>
          Module capability<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#plotterspectemplateskeymodulesindexchart">chart</a></b></td>
        <td>object</td>
        <td>
          Chart contains the information needed to use helm to install the capability<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the module<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          May be one of service, config or plugin Service: Means that the control plane deploys the component that performs the capability Config: Another pre-installed service performs the capability and the module deployed configures it for the particular workload or dataset Plugin: Indicates that this module performs a capability as part of another service or module rather than as a stand-alone module<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>scope</b></td>
        <td>enum</td>
        <td>
          Scope indicates at what level the capability is used: workload, asset, cluster If not indicated it is assumed to be asset<br/>
          <br/>
            <i>Enum</i>: asset, workload, cluster<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### Plotter.spec.templates[key].modules[index].chart
<sup><sup>[↩ Parent](#plotterspectemplateskeymodulesindex)</sup></sup>



Chart contains the information needed to use helm to install the capability

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
          Name of helm chart<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>chartPullSecret</b></td>
        <td>string</td>
        <td>
          Name of secret containing helm registry credentials<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>map[string]string</td>
        <td>
          Values to pass to helm chart installation<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### Plotter.spec.appSelector
<sup><sup>[↩ Parent](#plotterspec)</sup></sup>



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
        <td><b><a href="#plotterspecappselectorworkloadselector">workloadSelector</a></b></td>
        <td>object</td>
        <td>
          WorkloadSelector enables to connect the resource to the application Application labels should match the labels in the selector.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>clusterName</b></td>
        <td>string</td>
        <td>
          Cluster name<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### Plotter.spec.appSelector.workloadSelector
<sup><sup>[↩ Parent](#plotterspecappselector)</sup></sup>



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
        <td><b><a href="#plotterspecappselectorworkloadselectormatchexpressionsindex">matchExpressions</a></b></td>
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


#### Plotter.spec.appSelector.workloadSelector.matchExpressions[index]
<sup><sup>[↩ Parent](#plotterspecappselectorworkloadselector)</sup></sup>



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
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.<br/>
        </td>
        <td>false</td>
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
        <td><b><a href="#plotterstatusassetskey">assets</a></b></td>
        <td>map[string]object</td>
        <td>
          Assets is a map containing the status per asset. The key of this map is assetId<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#plotterstatusblueprintskey">blueprints</a></b></td>
        <td>map[string]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#plotterstatusconditionsindex">conditions</a></b></td>
        <td>[]object</td>
        <td>
          Conditions represent the possible error and failure conditions<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#plotterstatusflowskey">flows</a></b></td>
        <td>map[string]object</td>
        <td>
          Flows is a map containing the status for each flow the key is the flow name<br/>
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


#### Plotter.status.assets[key]
<sup><sup>[↩ Parent](#plotterstatus)</sup></sup>



ObservedState represents a part of the generated Blueprint/Plotter resource status that allows update of FybrikApplication status

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
          BlueprintStatus defines the observed state of Blueprint This includes readiness, error message, and indicators for the Kubernetes resources owned by the Blueprint for cleanup and status monitoring<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Plotter.status.blueprints[key].status
<sup><sup>[↩ Parent](#plotterstatusblueprintskey)</sup></sup>



BlueprintStatus defines the observed state of Blueprint This includes readiness, error message, and indicators for the Kubernetes resources owned by the Blueprint for cleanup and status monitoring

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
        <td><b><a href="#plotterstatusblueprintskeystatusmoduleskey">modules</a></b></td>
        <td>map[string]object</td>
        <td>
          ModulesState is a map which holds the status of each module its key is the moduleInstanceName which is the unique name for the deployed instance related to this workload<br/>
        </td>
        <td>false</td>
      </tr><tr>
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


#### Plotter.status.blueprints[key].status.modules[key]
<sup><sup>[↩ Parent](#plotterstatusblueprintskeystatus)</sup></sup>



ObservedState represents a part of the generated Blueprint/Plotter resource status that allows update of FybrikApplication status

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


#### Plotter.status.conditions[index]
<sup><sup>[↩ Parent](#plotterstatus)</sup></sup>



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
        <td><b>type</b></td>
        <td>string</td>
        <td>
          Type of the condition<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          Message contains the details of the current condition<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          ObservedGeneration is the version of the resource for which the condition has been evaluated<br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          Status of the condition, one of (`True`, `False`, `Unknown`).<br/>
          <br/>
            <i>Enum</i>: True, False, Unknown<br/>
            <i>Default</i>: Unknown<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### Plotter.status.flows[key]
<sup><sup>[↩ Parent](#plotterstatus)</sup></sup>



FlowStatus includes information to be reported back to the FybrikApplication resource It holds the status per data flow

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
        <td><b><a href="#plotterstatusflowskeysubflowskey">subFlows</a></b></td>
        <td>map[string]object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#plotterstatusflowskeystatus">status</a></b></td>
        <td>object</td>
        <td>
          ObservedState includes information about the current flow It includes readiness and error indications, as well as user instructions<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### Plotter.status.flows[key].subFlows[key]
<sup><sup>[↩ Parent](#plotterstatusflowskey)</sup></sup>



ObservedState represents a part of the generated Blueprint/Plotter resource status that allows update of FybrikApplication status

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


#### Plotter.status.flows[key].status
<sup><sup>[↩ Parent](#plotterstatusflowskey)</sup></sup>



ObservedState includes information about the current flow It includes readiness and error indications, as well as user instructions

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






Asset defines an asset in the catalog

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
        <td><b><a href="#assetspecdetails">details</a></b></td>
        <td>object</td>
        <td>
          Asset details<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#assetspecmetadata">metadata</a></b></td>
        <td>object</td>
        <td>
          Asset metadata<br/>
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


#### Asset.spec.details
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
        <td><b><a href="#assetspecdetailsconnection">connection</a></b></td>
        <td>object</td>
        <td>
          Connection information<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>dataFormat</b></td>
        <td>string</td>
        <td>
          Data format<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### Asset.spec.details.connection
<sup><sup>[↩ Parent](#assetspecdetails)</sup></sup>



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
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


#### Asset.spec.metadata
<sup><sup>[↩ Parent](#assetspec)</sup></sup>



Asset metadata

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
        <td><b><a href="#assetspecmetadatacolumnsindex">columns</a></b></td>
        <td>[]object</td>
        <td>
          Columns associated with the asset<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>geography</b></td>
        <td>string</td>
        <td>
          Geography of the resource<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the resource<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>owner</b></td>
        <td>string</td>
        <td>
          Owner of the resource<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>tags</b></td>
        <td>object</td>
        <td>
          Tags associated with the asset<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


#### Asset.spec.metadata.columns[index]
<sup><sup>[↩ Parent](#assetspecmetadata)</sup></sup>



ResourceColumn represents a column in a tabular resource

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
          Name of the column<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>tags</b></td>
        <td>object</td>
        <td>
          Tags associated with the column<br/>
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
          Name of the Secret resource<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the Secret resource. If it is empty then the asset namespace is used.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>
