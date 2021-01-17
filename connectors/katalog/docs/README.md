# API Reference

Packages:

- [katalog.m4d.ibm.com/v1alpha1](#katalog.m4d.ibm.com/v1alpha1)

# katalog.m4d.ibm.com/v1alpha1

Resource Types:

- [Asset](#asset)




## Asset
<sup><sup>[↩ Parent](#katalog.m4d.ibm.com/v1alpha1 )</sup></sup>








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
      <td>katalog.m4d.ibm.com/v1alpha1</td>
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
        <td></td>
        <td>true</td>
      </tr></tbody>
</table>


### Asset.spec
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
        <td>Asset details</td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#assetspecsecretref">secretRef</a></b></td>
        <td>object</td>
        <td>Reference to a Secret resource holding credentials for this asset</td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#assetspecsecurity">security</a></b></td>
        <td>object</td>
        <td></td>
        <td>true</td>
      </tr></tbody>
</table>


### Asset.spec.details
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
        <td></td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#assetspecdetailsconnection">connection</a></b></td>
        <td>object</td>
        <td>Connection information</td>
        <td>true</td>
      </tr></tbody>
</table>


### Asset.spec.details.connection
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
        <td><b><a href="#assetspecdetailsconnectiondb2">db2</a></b></td>
        <td>object</td>
        <td></td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#assetspecdetailsconnectionkafka">kafka</a></b></td>
        <td>object</td>
        <td></td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#assetspecdetailsconnections3">s3</a></b></td>
        <td>object</td>
        <td>Connection information for S3 compatible object store</td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td> [s3 db2 kafka]</td>
        <td>true</td>
      </tr></tbody>
</table>


### Asset.spec.details.connection.db2
<sup><sup>[↩ Parent](#assetspecdetailsconnection)</sup></sup>





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
        <td></td>
        <td>false</td>
      </tr><tr>
        <td><b>port</b></td>
        <td>string</td>
        <td></td>
        <td>false</td>
      </tr><tr>
        <td><b>ssl</b></td>
        <td>string</td>
        <td></td>
        <td>false</td>
      </tr><tr>
        <td><b>table</b></td>
        <td>string</td>
        <td></td>
        <td>false</td>
      </tr><tr>
        <td><b>url</b></td>
        <td>string</td>
        <td></td>
        <td>false</td>
      </tr></tbody>
</table>


### Asset.spec.details.connection.kafka
<sup><sup>[↩ Parent](#assetspecdetailsconnection)</sup></sup>





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
        <td></td>
        <td>false</td>
      </tr><tr>
        <td><b>key_deserializer</b></td>
        <td>string</td>
        <td></td>
        <td>false</td>
      </tr><tr>
        <td><b>sasl_mechanism</b></td>
        <td>string</td>
        <td></td>
        <td>false</td>
      </tr><tr>
        <td><b>schema_registry</b></td>
        <td>string</td>
        <td></td>
        <td>false</td>
      </tr><tr>
        <td><b>security_protocol</b></td>
        <td>string</td>
        <td></td>
        <td>false</td>
      </tr><tr>
        <td><b>ssl_truststore</b></td>
        <td>string</td>
        <td></td>
        <td>false</td>
      </tr><tr>
        <td><b>ssl_truststore_password</b></td>
        <td>string</td>
        <td></td>
        <td>false</td>
      </tr><tr>
        <td><b>topic_name</b></td>
        <td>string</td>
        <td></td>
        <td>false</td>
      </tr><tr>
        <td><b>value_deserializer</b></td>
        <td>string</td>
        <td></td>
        <td>false</td>
      </tr></tbody>
</table>


### Asset.spec.details.connection.s3
<sup><sup>[↩ Parent](#assetspecdetailsconnection)</sup></sup>



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
        <td></td>
        <td>false</td>
      </tr><tr>
        <td><b>bucket</b></td>
        <td>string</td>
        <td></td>
        <td>true</td>
      </tr><tr>
        <td><b>endpoint</b></td>
        <td>string</td>
        <td></td>
        <td>true</td>
      </tr><tr>
        <td><b>objectKey</b></td>
        <td>string</td>
        <td></td>
        <td>true</td>
      </tr></tbody>
</table>


### Asset.spec.secretRef
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
        <td>Name of the Secret resource (must exist in the same namespace)</td>
        <td>true</td>
      </tr></tbody>
</table>


### Asset.spec.security
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
        <td><b><a href="#assetspecsecuritycomponentsmetadatakey">componentsMetadata</a></b></td>
        <td>map[string]object</td>
        <td>metadata for each component in asset (e.g., column)</td>
        <td>false</td>
      </tr><tr>
        <td><b>geography</b></td>
        <td>string</td>
        <td></td>
        <td>false</td>
      </tr><tr>
        <td><b>namedMetadata</b></td>
        <td>map[string]string</td>
        <td></td>
        <td>false</td>
      </tr><tr>
        <td><b>owner</b></td>
        <td>string</td>
        <td></td>
        <td>false</td>
      </tr><tr>
        <td><b>tags</b></td>
        <td>[]string</td>
        <td>Tags associated with the asset</td>
        <td>false</td>
      </tr></tbody>
</table>


### Asset.spec.security.componentsMetadata[key]
<sup><sup>[↩ Parent](#assetspecsecurity)</sup></sup>





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
        <td></td>
        <td>false</td>
      </tr><tr>
        <td><b>namedMetadata</b></td>
        <td>map[string]string</td>
        <td>Named terms, that exist in Catalog toxonomy and the values for these terms for columns we will have "SchemaDetails" key, that will include technical schema details for this column TODO: Consider create special field for schema outside of metadata</td>
        <td>false</td>
      </tr><tr>
        <td><b>tags</b></td>
        <td>[]string</td>
        <td>Tags - can be any free text added to a component (no taxonomy)</td>
        <td>false</td>
      </tr></tbody>
</table>
