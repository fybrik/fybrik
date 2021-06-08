# Protocol Documentation
<a name="top"></a>



<a name="credentials.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## credentials.proto


 <!-- end services -->


<a name="connectors.Credentials"></a>

### Credentials



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| access_key | [string](#string) |  | access credential for the bucket where the asset is stored |
| secret_key | [string](#string) |  |  |
| username | [string](#string) |  |  |
| password | [string](#string) |  |  |
| api_key | [string](#string) |  | api key assigned to the bucket in which the asset is stored |
| resource_instance_id | [string](#string) |  | resource instance id for the bucket |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->



<a name="data_catalog_response.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## data_catalog_response.proto


 <!-- end services -->


<a name="connectors.CatalogDatasetInfo"></a>

### CatalogDatasetInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| dataset_id | [string](#string) |  |  |
| details | [DatasetDetails](#connectors.DatasetDetails) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->



<a name="policy_manager_response.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## policy_manager_response.proto


 <!-- end services -->


<a name="connectors.ComponentVersion"></a>

### ComponentVersion



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| id | [string](#string) |  |  |
| version | [string](#string) |  |  |






<a name="connectors.DatasetDecision"></a>

### DatasetDecision



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| dataset | [DatasetIdentifier](#connectors.DatasetIdentifier) |  |  |
| decisions | [OperationDecision](#connectors.OperationDecision) | repeated |  |






<a name="connectors.EnforcementAction"></a>

### EnforcementAction



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| id | [string](#string) |  |  |
| level | [EnforcementAction.EnforcementActionLevel](#connectors.EnforcementAction.EnforcementActionLevel) |  |  |
| args | [EnforcementAction.ArgsEntry](#connectors.EnforcementAction.ArgsEntry) | repeated |  |






<a name="connectors.EnforcementAction.ArgsEntry"></a>

### EnforcementAction.ArgsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="connectors.OperationDecision"></a>

### OperationDecision



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| operation | [AccessOperation](#connectors.AccessOperation) |  |  |
| enforcement_actions | [EnforcementAction](#connectors.EnforcementAction) | repeated |  |
| used_policies | [Policy](#connectors.Policy) | repeated |  |






<a name="connectors.PoliciesDecisions"></a>

### PoliciesDecisions



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| component_versions | [ComponentVersion](#connectors.ComponentVersion) | repeated |  |
| dataset_decisions | [DatasetDecision](#connectors.DatasetDecision) | repeated | one per dataset |
| general_decisions | [OperationDecision](#connectors.OperationDecision) | repeated |  |






<a name="connectors.Policy"></a>

### Policy



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| name | [string](#string) |  |  |
| description | [string](#string) |  |  |
| type | [string](#string) |  |  |
| hierarchy | [string](#string) | repeated |  |





 <!-- end messages -->


<a name="connectors.EnforcementAction.EnforcementActionLevel"></a>

### EnforcementAction.EnforcementActionLevel


| Name | Number | Description |
| ---- | ------ | ----------- |
| UNKNOWN | 0 |  |
| DATASET | 1 |  |
| COLUMN | 2 |  |
| ROW | 3 |  |
| CELL | 4 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->



<a name="data_catalog_request.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## data_catalog_request.proto


 <!-- end services -->


<a name="connectors.CatalogDatasetRequest"></a>

### CatalogDatasetRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| credential_path | [string](#string) |  | link to vault plugin for reading k8s secret with user credentials |
| dataset_id | [string](#string) |  | identifier of asset - always needed. JSON expected. Interpreted by the Connector, can contain any additional information as part of JSON |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->



<a name="policy_manager_request.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## policy_manager_request.proto


 <!-- end services -->


<a name="connectors.AccessOperation"></a>

### AccessOperation



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| type | [AccessOperation.AccessType](#connectors.AccessOperation.AccessType) |  |  |
| destination | [string](#string) |  | Destination for transfer or write. |






<a name="connectors.ApplicationContext"></a>

### ApplicationContext



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| credential_path | [string](#string) |  | link to vault plugin for reading k8s secret with user credentials |
| app_info | [ApplicationDetails](#connectors.ApplicationDetails) |  |  |
| datasets | [DatasetContext](#connectors.DatasetContext) | repeated |  |
| general_operations | [AccessOperation](#connectors.AccessOperation) | repeated |  |






<a name="connectors.ApplicationDetails"></a>

### ApplicationDetails



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| processing_geography | [string](#string) |  |  |
| properties | [ApplicationDetails.PropertiesEntry](#connectors.ApplicationDetails.PropertiesEntry) | repeated |  |






<a name="connectors.ApplicationDetails.PropertiesEntry"></a>

### ApplicationDetails.PropertiesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="connectors.DatasetContext"></a>

### DatasetContext



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| dataset | [DatasetIdentifier](#connectors.DatasetIdentifier) |  |  |
| operation | [AccessOperation](#connectors.AccessOperation) |  |  |






<a name="connectors.DatasetIdentifier"></a>

### DatasetIdentifier



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| dataset_id | [string](#string) |  | identifier of asset - always needed. JSON expected. Interpreted by the Connector, can contain any additional information as part of JSON |





 <!-- end messages -->


<a name="connectors.AccessOperation.AccessType"></a>

### AccessOperation.AccessType


| Name | Number | Description |
| ---- | ------ | ----------- |
| UNKNOWN | 0 |  |
| READ | 1 |  |
| COPY | 2 |  |
| WRITE | 3 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->



<a name="data_catalog_service.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## data_catalog_service.proto



<a name="connectors.DataCatalogService"></a>

### DataCatalogService


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| GetDatasetInfo | [CatalogDatasetRequest](#connectors.CatalogDatasetRequest) | [CatalogDatasetInfo](#connectors.CatalogDatasetInfo) |  |
| RegisterDatasetInfo | [RegisterAssetRequest](#connectors.RegisterAssetRequest) | [RegisterAssetResponse](#connectors.RegisterAssetResponse) |  |

 <!-- end services -->

 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->



<a name="policy_manager_service.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## policy_manager_service.proto



<a name="connectors.PolicyManagerService"></a>

### PolicyManagerService


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| GetPoliciesDecisions | [ApplicationContext](#connectors.ApplicationContext) | [PoliciesDecisions](#connectors.PoliciesDecisions) |  |

 <!-- end services -->

 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->



<a name="register_asset_response.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## register_asset_response.proto


 <!-- end services -->


<a name="connectors.RegisterAssetResponse"></a>

### RegisterAssetResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| asset_id | [string](#string) |  | Returns the id of the new asset registered in a catalog |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->



<a name="dataset_details.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## dataset_details.proto


 <!-- end services -->


<a name="connectors.CredentialsInfo"></a>

### CredentialsInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| vault_secret_path | [string](#string) |  | the path to Vault secret which is used to retrive the dataset credentials from the catalog. |






<a name="connectors.DataComponentMetadata"></a>

### DataComponentMetadata



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| component_type | [string](#string) |  | e.g., column |
| named_metadata | [DataComponentMetadata.NamedMetadataEntry](#connectors.DataComponentMetadata.NamedMetadataEntry) | repeated | Named terms, that exist in Catalog toxonomy and the values for these terms for columns we will have "SchemaDetails" key, that will include technical schema details for this column TODO: Consider create special field for schema outside of metadata |
| tags | [string](#string) | repeated | Tags - can be any free text added to a component (no taxonomy) |






<a name="connectors.DataComponentMetadata.NamedMetadataEntry"></a>

### DataComponentMetadata.NamedMetadataEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="connectors.DataStore"></a>

### DataStore



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| type | [DataStore.DataStoreType](#connectors.DataStore.DataStoreType) |  |  |
| name | [string](#string) |  | for auditing and readability. Can be same as location type or can have more info if availble from catalog |
| db2 | [Db2DataStore](#connectors.Db2DataStore) |  | oneof location { // should have been oneof but for technical rasons, a problem to translate it to JSON, we remove the oneof for now should have been local, db2, s3 without "location" but had a problem to compile it in proto - collision with proto name DataLocationDb2 |
| s3 | [S3DataStore](#connectors.S3DataStore) |  |  |
| kafka | [KafkaDataStore](#connectors.KafkaDataStore) |  |  |






<a name="connectors.DatasetDetails"></a>

### DatasetDetails



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | name in Catalog |
| data_owner | [string](#string) |  | information on the owner of data asset - can have different formats for different catalogs |
| data_store | [DataStore](#connectors.DataStore) |  | All info about the data store |
| data_format | [string](#string) |  |  |
| geo | [string](#string) |  | geography location where data resides (if this information available) |
| metadata | [DatasetMetadata](#connectors.DatasetMetadata) |  | LocationType locationType = 10; //publicCloud/privateCloud etc. Should be filled later when we understand better if we have a closed set of values and how they are used. |
| credentials_info | [CredentialsInfo](#connectors.CredentialsInfo) |  | information about how to retrive dataset credentials from the catalog. |






<a name="connectors.DatasetMetadata"></a>

### DatasetMetadata



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| dataset_named_metadata | [DatasetMetadata.DatasetNamedMetadataEntry](#connectors.DatasetMetadata.DatasetNamedMetadataEntry) | repeated |  |
| dataset_tags | [string](#string) | repeated | Tags - can be any free text added to a component (no taxonomy) |
| components_metadata | [DatasetMetadata.ComponentsMetadataEntry](#connectors.DatasetMetadata.ComponentsMetadataEntry) | repeated | metadata for each component in asset. In tabular data each column is a component, then we will have: column name -> column metadata |






<a name="connectors.DatasetMetadata.ComponentsMetadataEntry"></a>

### DatasetMetadata.ComponentsMetadataEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [DataComponentMetadata](#connectors.DataComponentMetadata) |  |  |






<a name="connectors.DatasetMetadata.DatasetNamedMetadataEntry"></a>

### DatasetMetadata.DatasetNamedMetadataEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="connectors.Db2DataStore"></a>

### Db2DataStore



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| url | [string](#string) |  |  |
| database | [string](#string) |  |  |
| table | [string](#string) |  | reformat to SCHEMA.TABLE struct |
| port | [string](#string) |  |  |
| ssl | [string](#string) |  | Note that bool value if set to "false" does not appear in the struct at all |






<a name="connectors.KafkaDataStore"></a>

### KafkaDataStore



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| topic_name | [string](#string) |  |  |
| bootstrap_servers | [string](#string) |  |  |
| schema_registry | [string](#string) |  |  |
| key_deserializer | [string](#string) |  |  |
| value_deserializer | [string](#string) |  |  |
| security_protocol | [string](#string) |  |  |
| sasl_mechanism | [string](#string) |  |  |
| ssl_truststore | [string](#string) |  |  |
| ssl_truststore_password | [string](#string) |  |  |






<a name="connectors.S3DataStore"></a>

### S3DataStore



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| endpoint | [string](#string) |  |  |
| bucket | [string](#string) |  |  |
| object_key | [string](#string) |  | can be object name or the prefix for dataset |
| region | [string](#string) |  | WKC does not return it, it will stay empty in our case!!! |





 <!-- end messages -->


<a name="connectors.DataStore.DataStoreType"></a>

### DataStore.DataStoreType


| Name | Number | Description |
| ---- | ------ | ----------- |
| UNKNOWN | 0 |  |
| LOCAL | 1 |  |
| S3 | 2 |  |
| DB2 | 3 |  |
| KAFKA | 4 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->



<a name="register_asset_request.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## register_asset_request.proto


 <!-- end services -->


<a name="connectors.RegisterAssetRequest"></a>

### RegisterAssetRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| creds | [Credentials](#connectors.Credentials) |  |  |
| dataset_details | [DatasetDetails](#connectors.DatasetDetails) |  |  |
| destination_catalog_id | [string](#string) |  |  |
| credential_path | [string](#string) |  | link to vault plugin for reading k8s secret with user credentials |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |
