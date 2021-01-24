// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0
package connector

import (
	"net"

	connectors "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
	kconfig "sigs.k8s.io/controller-runtime/pkg/client/config"
)

func Start(address string) error {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = AddToScheme(scheme)

	client, err := kclient.New(kconfig.GetConfigOrDie(), kclient.Options{Scheme: scheme})
	if err != nil {
		return errors.Wrap(err, "Failed to create client")
	}

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return errors.Wrap(err, "Failed to create a listerning socket")
	}

	server := grpc.NewServer()
	connectors.RegisterDataCatalogServiceServer(server, &DataCatalogService{client})
	connectors.RegisterDataCredentialServiceServer(server, &DataCredentialsService{client})

	if err := server.Serve(listener); err != nil {
		return errors.Wrap(err, "Connector server errored")
	}

	return nil
}
