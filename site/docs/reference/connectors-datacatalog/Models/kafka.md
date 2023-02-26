# kafka
Connection information for accessing a kafka topic
## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**bootstrap\_servers** | String | A comma-separated list of host/port pairs to use for establishing the initial connection to the Kafka cluster | [default: null]
**key\_deserializer** | String | Deserializer to be used for the keys of the topic | [optional] [default: null]
**sasl\_mechanism** | String | SASL Mechanism to be used (e.g. PLAIN or SCRAM-SHA-512) | [optional] [default: SCRAM-SHA-512]
**schema\_registry** | String | Host/port to connect the schema registry server | [optional] [default: null]
**security\_protocol** | String | Kafka security protocol one of (PLAINTEXT, SASL_PLAINTEXT, SASL_SSL, SSL) | [optional] [default: SASL_SSL]
**ssl\_truststore** | String | A truststore or certificate encoded as base64. The format can be JKS or PKCS12. | [optional] [default: null]
**topic\_name** | String | Name of the Kafka topic | [default: null]
**value\_deserializer** | String | Deserializer to be used for the values of the topic | [optional] [default: null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to API-Specification]](../README.md)

