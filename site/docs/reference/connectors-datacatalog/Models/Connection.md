# Connection
Details of connection types supported for accessing data stores. Not all are necessarily supported by fybrik storage allocation mechanism used to store temporary/persistent datasets.
## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**db2** | [db2](../Models/db2.md) |  | [optional] [default: null]
**fybrik-arrow-flight** | [fybrik-arrow-flight](../Models/fybrik-arrow-flight.md) |  | [optional] [default: null]
**google-sheets** | [google-sheets](../Models/google-sheets.md) |  | [optional] [default: null]
**kafka** | [kafka](../Models/kafka.md) |  | [optional] [default: null]
**mysql** | [mysql](../Models/mysql.md) |  | [optional] [default: null]
**name** | String | Name of the connection type to the data source | [default: null]
**postgres** | [postgres](../Models/postgres.md) |  | [optional] [default: null]
**s3** | [s3](../Models/s3.md) |  | [optional] [default: null]
**us-census** | [us-census](../Models/us-census.md) |  | [optional] [default: null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to API-Specification]](../README.md)

