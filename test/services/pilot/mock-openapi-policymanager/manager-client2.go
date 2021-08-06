package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"

	connectors "fybrik.io/fybrik/pkg/connectors/clients"
	openapiclientmodels "fybrik.io/fybrik/pkg/taxonomy/model/base"
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

	//mainPolicyManagerName := os.Getenv("MAIN_POLICY_MANAGER_NAME")
	mainPolicyManagerName := "OPEN API MANAGER"
	//mainPolicyManagerURL := os.Getenv("MAIN_POLICY_MANAGER_CONNECTOR_URL")
	//mainPolicyManagerURL := "http://v2opaconnector.m4d-system:50050"
	//connectionTimeout, err := getConnectionTimeout()
	// timeOutInSeconds := 120

	timeOutInSecs := getEnv("CONNECTION_TIMEOUT")
	timeOut, err := strconv.Atoi(timeOutInSecs)
	connectionTimeout := time.Duration(timeOut) * time.Second

	// mainPolicyManagerURL := "http://v2opaconnector.m4d-system:50050"
	// policyManager, err := connectors.NewOpenApiPolicyManager(mainPolicyManagerName, mainPolicyManagerURL, connectionTimeout)
	// if err != nil {
	// 	return
	// }

	mainPolicyManagerURL := "opa-connector.fybrik-system:80"
	policyManager, err := connectors.NewGrpcPolicyManager(mainPolicyManagerName, mainPolicyManagerURL, connectionTimeout)
	if err != nil {
		log.Println("returned with error ")
		log.Println("error in policyManager creation: ", err)
		return
	}

	creds := "http://vault.fybrik-system:8200/v1/kubernetes-secrets/wkc-creds?namespace=cp4d"

	// input := []openapiclientmodels.PolicymanagerRequest{*openapiclientmodels.NewPolicymanagerRequest(*openapiclientmodels.NewAction(openapiclientmodels.ActionType("read")), *openapiclientmodels.NewResource("{\"asset_id\": \"0bb3245e-e3ef-40b7-b639-c471bae4966c\", \"catalog_id\": \"503d683f-1d43-4257-a1a3-0ddf5e446ba5\"}", creds))} // []PolicymanagerRequest | input values that need to be considered for filter

	input := openapiclientmodels.NewPolicyManagerRequestWithDefaults()

	reqCtx := make(map[string]interface{})
	reqCtx["intent"] = "Fraud Detection"
	reqCtx["role"] = "Data Scientist"
	input.SetContext(reqCtx)

	action := openapiclientmodels.PolicyManagerRequestAction{}
	action.SetActionType(openapiclientmodels.READ)
	processLocation := "Netherlands"
	action.SetProcessingLocation(processLocation)
	input.SetAction(action)

	//input.SetAction(*openapiclientmodels.NewAction(openapiclientmodels.ActionType("read")))
	// /0fd6ff25-7327-4b55-8ff2-56cc1c934824/asset/5067b64a-67bc-4067-9117-0aff0a9963ea/a
	// input.SetResource(*openapiclientmodels.NewResource("{\"asset_id\": \"0bb3245e-e3ef-40b7-b639-c471bae4966c\", \"catalog_id\": \"503d683f-1d43-4257-a1a3-0ddf5e446ba5\"}"))
	input.SetResource(*openapiclientmodels.NewResource("{\"asset_id\": \"5067b64a-67bc-4067-9117-0aff0a9963ea\", \"catalog_id\": \"0fd6ff25-7327-4b55-8ff2-56cc1c934824\"}"))

	//input.SetRequestContext(openapiclientmodels.RequestContext{})

	// input := openapiclientmodels.PolicymanagerRequest{*openapiclientmodels.NewPolicymanagerRequest(*openapiclientmodels.NewAction(openapiclientmodels.ActionType("read")), *openapiclientmodels.NewResource("{\"asset_id\": \"0bb3245e-e3ef-40b7-b639-c471bae4966c\", \"catalog_id\": \"503d683f-1d43-4257-a1a3-0ddf5e446ba5\"}", creds))} // []PolicymanagerRequest | input values that need to be considered for filter

	log.Println("in manager-client - policy manager request: ", input)
	log.Println("in manager-client - creds: ", creds)

	response, err := policyManager.GetPoliciesDecisions(input, creds)

	bytes, _ := response.MarshalJSON()
	log.Println("in manager-client - Response from `policyManager.GetPoliciesDecisions`: \n", string(bytes))

	var resp openapiclientmodels.PolicyManagerResponse
	err = json.Unmarshal(bytes, &resp)
	log.Println("err: ", err)
	log.Println("resp: ", resp)

	//res2B, _ := json.Marshal(resp)
	res, err := json.MarshalIndent(resp, "", "\t")
	log.Println("err :", err)
	log.Println("marshalled response:", string(res))
}
