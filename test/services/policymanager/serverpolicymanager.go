// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	mockup "fybrik.io/fybrik/manager/controllers/mockup"
	openapiclientmodels "fybrik.io/fybrik/pkg/taxonomy/model/base"
	"github.com/gin-gonic/gin"
)

const (
	PORT = 50082
)

// func main() {
// 	address := utils.ListeningAddress(PORT)
// 	log.Printf("starting mock policy manager server on address %s", address)

// 	listener, err := net.Listen("tcp", address)
// 	if err != nil {
// 		log.Fatalf("listening error: %v", err)
// 	}

// 	server := grpc.NewServer()
// 	service := &mockup.MockPolicyManager{}

// 	pb.RegisterPolicyManagerServiceServer(server, service)
// 	if err := server.Serve(listener); err != nil {
// 		log.Fatalf("cannot serve mock policy manager: %v", err)
// 	}

// }

var router *gin.Engine

func constructPolicyManagerRequest(inputString string) *openapiclientmodels.PolicyManagerRequest {
	fmt.Println("inconstructPolicymanagerRequest")
	fmt.Println("inputString")
	fmt.Println(inputString)
	var input openapiclientmodels.PolicyManagerRequest
	err := json.Unmarshal([]byte(inputString), &input)
	if err != nil {
		return nil
	}
	fmt.Println("input:", input)
	return &input
}

func main() {
	router = gin.Default()

	router.GET("/getPoliciesDecisions", func(c *gin.Context) {
		input := c.Query("input")
		creds := c.Query("creds")
		policyManagerReq := constructPolicyManagerRequest(input)
		policyManager := &mockup.MockPolicyManager{}
		policyManagerResp, err := policyManager.GetPoliciesDecisions(policyManagerReq, creds)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error in GetPoliciesDecisions!")
			return
		}
		c.JSON(http.StatusOK, policyManagerResp)
	})

	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World!")
	})

	log.Fatal(router.Run(":" + strconv.Itoa(PORT)))
}
