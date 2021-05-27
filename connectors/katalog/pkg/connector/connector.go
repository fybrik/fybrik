// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0
package connector

import (
	"net"

	connectors "github.com/mesh-for-data/mesh-for-data/pkg/connectors/protobuf"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/runtime"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
	kconfig "sigs.k8s.io/controller-runtime/pkg/client/config"
)

func Start(address string) error {
	scheme := runtime.NewScheme()
	_ = AddToScheme(scheme)

	client, err := kclient.New(kconfig.GetConfigOrDie(), kclient.Options{Scheme: scheme})
	if err != nil {
		return errors.Wrap(err, "failed to create client")
	}

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return errors.Wrap(err, "failed to create a listerning socket")
	}

	server := grpc.NewServer()
	connectors.RegisterDataCatalogServiceServer(server, &DataCatalogService{client: client})

	if err := server.Serve(listener); err != nil {
		return errors.Wrap(err, "connector server errored")
	}

	return nil
}
