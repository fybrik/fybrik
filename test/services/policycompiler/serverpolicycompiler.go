// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"log"
	"net"

	mockup "github.com/ibm/the-mesh-for-data/manager/controllers/mockup"
	"github.com/ibm/the-mesh-for-data/manager/controllers/utils"
	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedPolicyManagerServiceServer
}

func (s *server) GetPoliciesDecisions(ctx context.Context, in *pb.ApplicationContext) (*pb.PoliciesDecisions, error) {
	policyCompiler := &mockup.MockPolicyCompiler{}
	return policyCompiler.GetPoliciesDecisions(in)
}

func main() {
	address := utils.ListeningAddress(50082)
	log.Printf("Starting mock policy compiler server on address " + address)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Error in listening: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterPolicyManagerServiceServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Error in service: %v", err)
	}
}
