// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"log"
	"net"

	"github.com/mesh-for-data/mesh-for-data/manager/controllers/mockup"
	"github.com/mesh-for-data/mesh-for-data/manager/controllers/utils"
	pb "github.com/mesh-for-data/mesh-for-data/pkg/connectors/protobuf"
	"google.golang.org/grpc"
)

const (
	PORT = 8080
)

func main() {
	address := utils.ListeningAddress(PORT)
	log.Printf("starting mock catalog server on address %s", address)

	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("listening error: %v", err)
	}

	server := grpc.NewServer()
	service := mockup.NewTestCatalog()

	pb.RegisterDataCatalogServiceServer(server, service)
	if err := server.Serve(listener); err != nil {
		log.Fatalf("cannot serve mock data catalog: %v", err)
	}
}
