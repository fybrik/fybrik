// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"log"
	"time"

	pb "github.com/mesh-for-data/mesh-for-data/pkg/connectors/protobuf"
	openapiclientmodels "github.com/mesh-for-data/mesh-for-data/pkg/taxonomy/model/base"
	"google.golang.org/grpc"
)

var _ PolicyManager = (*grpcPolicyManager)(nil)

type grpcPolicyManager struct {
	name       string
	connection *grpc.ClientConn
	client     pb.PolicyManagerServiceClient
}

// ref: https://sosedoff.com/2014/12/15/generate-random-hex-string-in-go.html
func randomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// NewGrpcPolicyManager creates a PolicyManager facade that connects to a GRPC service
// You must call .Close() when you are done using the created instance
func NewGrpcPolicyManager(name string, connectionURL string, connectionTimeout time.Duration) (PolicyManager, error) {
	// log.Println("in NewGrpcPolicyManager: ")
	// ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
	// defer cancel()
	// log.Println("name: ", name)
	// log.Println("connectionURL: ", connectionURL)
	// connection, err := grpc.DialContext(ctx, connectionURL, grpc.WithInsecure(), grpc.WithBlock())
	// log.Println("connectionTimeout: ", connectionTimeout)
	// if err != nil {
	// 	return nil, errors.Wrap(err, fmt.Sprintf("NewGrpcPolicyManager failed when connecting to %s", connectionURL))
	// }

	serverAddr := flag.String("server_addr", connectionURL, "The server address in the format of host:port")
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())
	log.Println("serverAddr: ", *serverAddr)
	connection, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	//defer conn.Close()

	log.Println("connectionURL: ", connectionURL)
	return &grpcPolicyManager{
		name:       name,
		client:     pb.NewPolicyManagerServiceClient(connection),
		connection: connection,
	}, nil
}

func (m *grpcPolicyManager) GetPoliciesDecisions(in *openapiclientmodels.PolicyManagerRequest, creds string) (*openapiclientmodels.PolicyManagerResponse, error) {

	log.Println("GetPoliciesDecisions: entry")
	appContext := convertOpenApiReqToGrpcReq(in, creds)

	result, _ := m.client.GetPoliciesDecisions(context.Background(), appContext)

	policyManagerResp := convGrpcRespToOpenApiResp(result)
	log.Println("GetPoliciesDecisions: exit")
	return policyManagerResp, nil

	//return policyManagerResp, errors.Wrap(err, fmt.Sprintf("get policies decisions from %s failed", m.name))
	// return result, errors.Wrap(err, fmt.Sprintf("get policies decisions from %s failed", m.name))
}

// func (m *grpcPolicyManager) GetPoliciesDecisions(ctx context.Context, in *pb.ApplicationContext) (*pb.PoliciesDecisions, error) {
// 	result, err := m.client.GetPoliciesDecisions(ctx, in)
// 	return result, errors.Wrap(err, fmt.Sprintf("get policies decisions from %s failed", m.name))
// }

func (m *grpcPolicyManager) Close() error {
	return m.connection.Close()
}
