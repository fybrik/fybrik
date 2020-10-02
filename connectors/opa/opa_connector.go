// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	opabl "github.com/ibm/the-mesh-for-data/connectors/opa/opaconn-bl"
	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
	"google.golang.org/grpc"
)

var opaServerURL = ""

const defaultPort = "50082" //synched with opa_connector.yaml

type server struct {
	pb.UnimplementedPolicyManagerServiceServer
	opaReader *opabl.Server
}

func getEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatalf("Env Variable %v not defined", key)
	}
	log.Printf("Env. variable extracted: %s - %s\n", key, value)
	return value
}

func getEnvWithDefault(key string, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Printf("Env. variable not found, defualt value used: %s - %s\n", key, defaultValue)
		return defaultValue
	}

	log.Printf("Env. variable extracted: %s - %s\n", key, value)
	return value
}

func (s *server) GetPoliciesDecisions(ctx context.Context, in *pb.ApplicationContext) (*pb.PoliciesDecisions, error) {

	log.Println("*********************************Received ApplicationContext *****************************")
	log.Println(in)
	log.Println("******************************************************************************************")
	catalogConnectorAddress := getEnv("CATALOG_CONNECTOR_URL")

	timeOutInSecs := getEnv("CONNECTION_TIMEOUT")
	timeOut, err := strconv.Atoi(timeOutInSecs)

	if err != nil {
		log.Printf("Atoi conversion of timeOutinseconds failed: %v", err)
		return nil, fmt.Errorf("Atoi conversion of timeOutinseconds failed: %v", err)
	}

	eval, err := s.opaReader.GetPoliciesDecisions(in, catalogConnectorAddress, timeOut)
	jsonOutput, _ := json.MarshalIndent(eval, "", "\t")
	log.Println("Received evaluation : " + string(jsonOutput))
	log.Println("err:", err)

	return eval, err
}

func main() {
	port := getEnvWithDefault("PORT_OPA_CONNECTOR", defaultPort)
	opaServerURL = getEnv("OPA_SERVER_URL") // set global variable

	log.Println("OPA_SERVER_URL env variable in OPAConnector: ", opaServerURL)
	log.Println("Using port to start go opa connector : ", port)

	log.Printf("Server starts listening on port %v", port)
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Error in listening: %v", err)
	}
	s := grpc.NewServer()
	srv := &server{opaReader: opabl.NewServer(opaServerURL)}
	pb.RegisterPolicyManagerServiceServer(s, srv)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Error in service: %v", err)
	}
}
