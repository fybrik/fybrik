// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"

	pmclient "fybrik.io/fybrik/pkg/connectors/policymanager/clients"
	taxonomymodels "fybrik.io/fybrik/pkg/taxonomy/model/policymanager/base"
)

func getEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatalf("Env Variable %v not defined", key)
	}
	log.Printf("Env. variable extracted: %s - %s\n", key, value)
	return value
}

func main() {
	mainPolicyManagerName := "OPEN API MANAGER"

	timeOutInSecs := getEnv("CONNECTION_TIMEOUT")
	timeOut, _ := strconv.Atoi(timeOutInSecs)
	connectionTimeout := time.Duration(timeOut) * time.Second

	mainPolicyManagerURL := "http://opa-connector.fybrik-system:80"
	log.Println("mainPolicyManagerURL set to :", mainPolicyManagerURL)
	policyManager, err := pmclient.NewOpenAPIPolicyManager(mainPolicyManagerName, mainPolicyManagerURL, connectionTimeout)
	if err != nil {
		return
	}

	creds := "http://vault.fybrik-system:8200/v1/kubernetes-secrets/<SECRET-NAME>?namespace=<NAMESPACE>"
	input := taxonomymodels.NewPolicyManagerRequestWithDefaults()

	reqCtx := make(map[string]interface{})
	reqCtx["intent"] = "Fraud Detection"
	reqCtx["role"] = "Data Scientist"
	// reqCtx["role"] = "Business Analyst"
	input.SetContext(reqCtx)

	action := taxonomymodels.PolicyManagerRequestAction{}
	action.SetActionType(taxonomymodels.READ)
	processLocation := "Netherlands"
	action.SetProcessingLocation(processLocation)
	input.SetAction(action)

	input.SetResource(*taxonomymodels.NewResource("{\"asset_id\": \"5067b64a-67bc-4067-9117-0aff0a9963ea\", \"catalog_id\": \"0fd6ff25-7327-4b55-8ff2-56cc1c934824\"}"))

	log.Println("in manager-client - policy manager request: ", input)
	log.Println("in manager-client - creds: ", creds)

	response, _ := policyManager.GetPoliciesDecisions(input, creds)

	bytes, _ := response.MarshalJSON()
	log.Println("in manager-client - Response from `policyManager.GetPoliciesDecisions`: \n", string(bytes))

	var resp taxonomymodels.PolicyManagerResponse
	err = json.Unmarshal(bytes, &resp)
	log.Println("err: ", err)
	log.Println("resp: ", resp)

	res, err := json.MarshalIndent(resp, "", "\t")
	log.Println("err :", err)
	log.Println("marshalled response:", string(res))
}
