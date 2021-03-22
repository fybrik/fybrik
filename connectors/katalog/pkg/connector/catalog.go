// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0
package connector

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"

	"log"

	utils "github.com/ibm/the-mesh-for-data/connectors/katalog/pkg/connector/utils"
	connectors "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// TODO(roee88): This is a temporary implementation of a catalog connector to
// Katalog. It is here to map between Katalog CRDs to the connectors proto
// definitions. Eventually, the connectors proto definitions won't hardcode so
// much and rely on validating against a configured OpenAPI spec instead, making
// most of the code in this file unnecessary.

type DataCatalogService struct {
	connectors.UnimplementedDataCatalogServiceServer

	client kclient.Client
}

func (s *DataCatalogService) GetDatasetInfo(ctx context.Context, req *connectors.CatalogDatasetRequest) (*connectors.CatalogDatasetInfo, error) {
	namespace, name, err := utils.SplitNamespacedName(req.DatasetId)
	if err != nil {
		return nil, err
	}
	log.Printf("In GetDatasetInfo: asset namespace is " + namespace + " asset name is " + name)
	asset, err := getAsset(ctx, s.client, namespace, name)
	if err != nil {
		return nil, err
	}

	datastore, err := buildDataStore(asset)
	if err != nil {
		return nil, err
	}

	log.Printf("In GetDatasetInfo: VaultSecretPath is " + utils.VaultSecretPath(namespace, asset.Spec.SecretRef.Name))
	return &connectors.CatalogDatasetInfo{
		DatasetId: req.DatasetId,
		Details: &connectors.DatasetDetails{
			Name:       req.DatasetId,
			DataOwner:  utils.EmptyIfNil(asset.Spec.AssetMetadata.Owner),
			DataFormat: utils.EmptyIfNil(asset.Spec.AssetDetails.DataFormat),
			Geo:        utils.EmptyIfNil(asset.Spec.AssetMetadata.Geography),
			DataStore:  datastore,
			CredentialsInfo: &connectors.CredentialsInfo{
				VaultSecretPath: utils.VaultSecretPath(namespace, asset.Spec.SecretRef.Name),
			},
			Metadata: buildDatasetMetadata(asset),
		},
	}, nil
}

func buildDatasetMetadata(asset *Asset) *connectors.DatasetMetadata {
	assetMetadata := asset.Spec.AssetMetadata

	var namedMetadata map[string]string
	if assetMetadata.NamedMetadata != nil {
		namedMetadata = assetMetadata.NamedMetadata.AdditionalProperties
	}

	componentsMetadata := map[string]*connectors.DataComponentMetadata{}
	for componentName, componentValue := range assetMetadata.ComponentsMetadata.AdditionalProperties {
		var componentNamedMetadata map[string]string
		if componentValue.NamedMetadata != nil {
			componentNamedMetadata = componentValue.NamedMetadata.AdditionalProperties
		}
		componentsMetadata[componentName] = &connectors.DataComponentMetadata{
			ComponentType: "column",
			Tags:          utils.EmptyArrayIfNil(componentValue.Tags),
			NamedMetadata: componentNamedMetadata,
		}
	}

	return &connectors.DatasetMetadata{
		DatasetTags:          utils.EmptyArrayIfNil(assetMetadata.Tags),
		DatasetNamedMetadata: namedMetadata,
		ComponentsMetadata:   componentsMetadata,
	}
}

func buildDataStore(asset *Asset) (*connectors.DataStore, error) {
	connection := asset.Spec.AssetDetails.Connection
	switch connection.Type {
	case "s3":
		return &connectors.DataStore{
			Type: connectors.DataStore_S3,
			Name: asset.Name,
			S3: &connectors.S3DataStore{
				Endpoint:  connection.S3.Endpoint,
				Bucket:    connection.S3.Bucket,
				ObjectKey: connection.S3.ObjectKey,
				Region:    utils.EmptyIfNil(connection.S3.Region),
			},
		}, nil
	case "kafka":
		return &connectors.DataStore{
			Type: connectors.DataStore_KAFKA,
			Name: asset.Name,
			Kafka: &connectors.KafkaDataStore{
				TopicName:             utils.EmptyIfNil(connection.Kafka.TopicName),
				BootstrapServers:      utils.EmptyIfNil(connection.Kafka.BootstrapServers),
				SchemaRegistry:        utils.EmptyIfNil(connection.Kafka.SchemaRegistry),
				KeyDeserializer:       utils.EmptyIfNil(connection.Kafka.KeyDeserializer),
				ValueDeserializer:     utils.EmptyIfNil(connection.Kafka.ValueDeserializer),
				SecurityProtocol:      utils.EmptyIfNil(connection.Kafka.SecurityProtocol),
				SaslMechanism:         utils.EmptyIfNil(connection.Kafka.SaslMechanism),
				SslTruststore:         utils.EmptyIfNil(connection.Kafka.SslTruststore),
				SslTruststorePassword: utils.EmptyIfNil(connection.Kafka.SslTruststorePassword),
			},
		}, nil
	case "db2":
		return &connectors.DataStore{
			Type: connectors.DataStore_DB2,
			Name: asset.Name,
			Db2: &connectors.Db2DataStore{
				Url:      utils.EmptyIfNil(connection.Db2.Url),
				Database: utils.EmptyIfNil(connection.Db2.Database),
				Table:    utils.EmptyIfNil(connection.Db2.Table),
				Port:     utils.EmptyIfNil(connection.Db2.Port),
				Ssl:      utils.EmptyIfNil(connection.Db2.Ssl),
			},
		}, nil
	default:
		return nil, errors.New("unknown datastore type")
	}
}

func getAsset(ctx context.Context, client kclient.Client, namespace string, name string) (*Asset, error) {
	// Read asset as unstructured
	object := &unstructured.Unstructured{}
	object.SetGroupVersionKind(schema.GroupVersionKind{Group: GroupVersion.Group, Version: GroupVersion.Version, Kind: "Asset"})
	object.SetNamespace(namespace)
	object.SetName(name)

	objectKey, err := kclient.ObjectKeyFromObject(object)
	if err != nil {
		return nil, err
	}

	err = client.Get(ctx, objectKey, object)
	if err != nil {
		return nil, err
	}

	// Decode into an Asset object
	asset := &Asset{}
	bytes, err := object.MarshalJSON()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, asset)
	if err != nil {
		return nil, err
	}

	return asset, nil
}
