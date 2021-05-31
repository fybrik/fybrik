// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package policycompilerbl

import (
	"context"
	"log"
	"time"

	pb "github.com/mesh-for-data/mesh-for-data/pkg/connectors/protobuf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type IPolicyManagerConnector interface {
	GetEnforcementActions(*pb.ApplicationContext) (*pb.PoliciesDecisions, error)
}

type PolicyManagerConnector struct {
	IPolicyManagerConnector

	name             string
	address          string
	timeOutInSeconds int
}

func NewPolicyManagerConnector(name, address string, timeOutInSeconds int) PolicyManagerConnector {
	return PolicyManagerConnector{name: name, address: address, timeOutInSeconds: timeOutInSeconds}
}

func (pm PolicyManagerConnector) GetEnforcementActions(appContext *pb.ApplicationContext) (*pb.PoliciesDecisions, error) {
	log.Println("********************  GetEnforcementActions in policy manager ***************************")
	log.Println(appContext)
	log.Println("*********************************************************************")
	log.Println("policy manager attempting to connect to address: ", pm.address)
	log.Println("policy manager attempting with timeout: ", pm.timeOutInSeconds)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(pm.timeOutInSeconds)*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, pm.address, grpc.WithInsecure(), grpc.WithBlock())

	if err != nil {
		log.Printf("Connection to Connector failed: %v", err)
		errStatus, _ := status.FromError(err)
		log.Println("Message: ", errStatus.Message())
		log.Println("Code: ", errStatus.Code())

		return nil, err
	}
	defer conn.Close()

	c := pb.NewPolicyManagerServiceClient(conn)
	policiesDecisions, err := c.GetPoliciesDecisions(ctx, appContext)

	if err != nil {
		log.Printf("Error calling Connetcor GRPC: %v", err)
		errStatus, _ := status.FromError(err)
		log.Println("Message: ", errStatus.Message())
		log.Println("Code: ", errStatus.Code())

		return nil, err
	}
	log.Println("******************** Policy Decisions Fetched ***************************")
	log.Println(policiesDecisions)
	log.Println("*********************************************************************")
	return policiesDecisions, nil
}
