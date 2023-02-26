# Data Catalog Service - Asset Details

<a name="documentation-for-api-endpoints"></a>
## Documentation for API Endpoints

All URIs are relative to *https://localhost:8080*

Class | Method | HTTP request | Description
------------ | ------------- | ------------- | -------------
*DefaultApi* | [**createAsset**](Apis/DefaultApi.md#createasset) | **POST** /createAsset | This REST API writes data asset information to the data catalog configured in fybrik
*DefaultApi* | [**deleteAsset**](Apis/DefaultApi.md#deleteasset) | **DELETE** /deleteAsset | This REST API deletes data asset
*DefaultApi* | [**getAssetInfo**](Apis/DefaultApi.md#getassetinfo) | **POST** /getAssetInfo | This REST API gets data asset information from the data catalog configured in fybrik for the data sets indicated in FybrikApplication yaml
*DefaultApi* | [**updateAsset**](Apis/DefaultApi.md#updateasset) | **PATCH** /updateAsset | This REST API updates data asset information in the data catalog configured in fybrik


<a name="documentation-for-models"></a>
## Documentation for Models

 - [Connection](Models/Connection.md)
 - [CreateAssetRequest](Models/CreateAssetRequest.md)
 - [CreateAssetResponse](Models/CreateAssetResponse.md)
 - [DeleteAssetRequest](Models/DeleteAssetRequest.md)
 - [DeleteAssetResponse](Models/DeleteAssetResponse.md)
 - [GetAssetRequest](Models/GetAssetRequest.md)
 - [GetAssetResponse](Models/GetAssetResponse.md)
 - [ResourceColumn](Models/ResourceColumn.md)
 - [ResourceDetails](Models/ResourceDetails.md)
 - [ResourceMetadata](Models/ResourceMetadata.md)
 - [UpdateAssetRequest](Models/UpdateAssetRequest.md)
 - [UpdateAssetResponse](Models/UpdateAssetResponse.md)
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
