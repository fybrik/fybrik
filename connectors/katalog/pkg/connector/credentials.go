// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0
package connector

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"github.com/ibm/the-mesh-for-data/connectors/katalog/pkg/taxonomy"
	connectors "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type DataCredentialsService struct {
	client kclient.Client
}

func (s *DataCredentialsService) GetCredentialsInfo(ctx context.Context, req *connectors.DatasetCredentialsRequest) (*connectors.DatasetCredentials, error) {
	namespace, name, err := splitNamespacedName(req.DatasetId)
	if err != nil {
		return nil, err
	}

	// Read the secret
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}
	objectKey, err := kclient.ObjectKeyFromObject(secret)
	if err != nil {
		return nil, err
	}
	err = s.client.Get(ctx, objectKey, secret)
	if err != nil {
		return nil, err
	}

	// Get the data fields as strings
	data := map[string]string{}
	for key, value := range secret.Data {
		data[key] = string(value)
	}

	// Decode into Authentication structure
	authn := &taxonomy.Authentication{}
	switch secret.Type {
	case corev1.SecretTypeOpaque:
		err = decodeToStruct(data, authn)
		if err != nil {
			return nil, errors.Wrap(err, "Invalid fields in Secret data")
		}
	case corev1.SecretTypeBasicAuth:
		username := data[corev1.BasicAuthUsernameKey]
		password := data[corev1.BasicAuthPasswordKey]
		authn.Username = &username
		authn.Password = &password
	default:
		// TODO(roee88): add SSHAuth and TLSAuth as in corev1.SecretType
		return nil, fmt.Errorf("unknown secret type %s", secret.Type)
	}

	// Map to current connectors API
	return &connectors.DatasetCredentials{
		DatasetId: req.DatasetId,
		Creds: &connectors.Credentials{
			AccessKey: emptyIfNil(authn.AccessKey),
			SecretKey: emptyIfNil(authn.SecretKey),
			ApiKey:    emptyIfNil(authn.ApiKey),
			Username:  emptyIfNil(authn.Username),
			Password:  emptyIfNil(authn.Password),
		}}, nil
}
