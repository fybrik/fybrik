# Storage Manager Service

<a name="documentation-for-api-endpoints"></a>
## Documentation for API Endpoints

All URIs are relative to *https://localhost:8082*

Class | Method | HTTP request | Description
------------ | ------------- | ------------- | -------------
*DefaultApi* | [**allocateStorage**](Apis/DefaultApi.md#allocatestorage) | **POST** /allocateStorage | This REST API allocates storage based on the storage account selected by Fybrik
*DefaultApi* | [**deleteStorage**](Apis/DefaultApi.md#deletestorage) | **DELETE** /deleteStorage | This REST API deletes allocated storage
*DefaultApi* | [**getSupportedStorageTypes**](Apis/DefaultApi.md#getsupportedstoragetypes) | **POST** /getSupportedStorageTypes | This REST API returns a list of supported storage types


<a name="documentation-for-models"></a>
## Documentation for Models

 - [AllocateStorageRequest](Models/AllocateStorageRequest.md)
 - [AllocateStorageResponse](Models/AllocateStorageResponse.md)
 - [ApplicationDetails](Models/ApplicationDetails.md)
 - [ConfigOptions](Models/ConfigOptions.md)
 - [Connection](Models/Connection.md)
 - [DatasetDetails](Models/DatasetDetails.md)
 - [DeleteStorageRequest](Models/DeleteStorageRequest.md)
 - [GetSupportedStorageTypesResponse](Models/GetSupportedStorageTypesResponse.md)
 - [Options](Models/Options.md)
 - [SecretRef](Models/SecretRef.md)
 - [db2](Models/db2.md)
 - [fybrik-arrow-flight](Models/fybrik-arrow-flight.md)
 - [google-sheets](Models/google-sheets.md)
 - [https](Models/https.md)
 - [kafka](Models/kafka.md)
 - [mysql](Models/mysql.md)
 - [postgres](Models/postgres.md)
 - [s3](Models/s3.md)


<a name="documentation-for-authorization"></a>
## Documentation for Authorization

All endpoints do not require authorization.
