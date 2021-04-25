// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"time"

	"encoding/json"

	"google.golang.org/grpc"

	app "github.com/mesh-for-data/mesh-for-data/manager/apis/app/v1alpha1"
	"github.com/mesh-for-data/mesh-for-data/manager/controllers/utils"
	pb "github.com/mesh-for-data/mesh-for-data/pkg/connectors/protobuf"
	"github.com/mesh-for-data/mesh-for-data/pkg/vault"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// RegisterAsset registers a new asset in the specified catalog
// Input arguments:
// - catalogID: the destination catalog identifier
// - info: connection and credential details
// Returns:
// - an error if happened
// - the new asset identifier
func (r *M4DApplicationReconciler) RegisterAsset(catalogID string, info *app.DatasetDetails, input *app.M4DApplication) (string, error) {
	// Set up a connection to the data catalog interface server.
	conn, err := grpc.Dial(utils.GetDataCatalogServiceAddress(), grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return "", err
	}
	defer conn.Close()
	c := pb.NewDataCatalogServiceClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	datasetDetails := &pb.DatasetDetails{}
	err = info.Details.Into(datasetDetails)
	if err != nil {
		return "", err
	}

	var creds *pb.Credentials
	if creds, err = SecretToCredentials(r.Client, types.NamespacedName{Name: info.SecretRef, Namespace: utils.GetSystemNamespace()}); err != nil {
		return "", err
	}
	var credentialPath string
	if input.Spec.SecretRef != "" {
		credentialPath = utils.GetVaultAddress() + vault.PathForReadingKubeSecret(input.Namespace, input.Spec.SecretRef)
	}

	response, err := c.RegisterDatasetInfo(ctx, &pb.RegisterAssetRequest{
		Creds:                creds,
		DatasetDetails:       datasetDetails,
		DestinationCatalogId: catalogID,
		CredentialPath:       credentialPath,
	})
	if err != nil {
		return "", err
	}
	return response.GetAssetId(), nil
}

var translationMap = map[string]string{
	"accessKeyID":        "access_key",
	"accessKey":          "access_key",
	"secretAccessKey":    "secret_key",
	"SecretKey":          "secret_key",
	"apiKey":             "api_key",
	"resourceInstanceId": "resource_instance_id",
}

// SecretToCredentialMap fetches a secret and converts into a map matching credentials proto
func SecretToCredentialMap(cl client.Client, secretRef types.NamespacedName) (map[string]interface{}, error) {
	// fetch a secret
	secret := &corev1.Secret{}
	if err := cl.Get(context.Background(), secretRef, secret); err != nil {
		return nil, err
	}
	credsMap := make(map[string]interface{})
	for key, val := range secret.Data {
		if translated, found := translationMap[key]; found {
			credsMap[translated] = string(val)
		} else {
			credsMap[key] = string(val)
		}
	}
	return credsMap, nil
}

// SecretToCredentials fetches a secret and constructs Credentials structure
func SecretToCredentials(cl client.Client, secretRef types.NamespacedName) (*pb.Credentials, error) {
	credsMap, err := SecretToCredentialMap(cl, secretRef)
	if err != nil {
		return nil, err
	}
	jsonStr, err := json.Marshal(credsMap)
	if err != nil {
		return nil, err
	}
	var creds pb.Credentials
	if err := json.Unmarshal(jsonStr, &creds); err != nil {
		return nil, err
	}
	return &creds, nil
}
