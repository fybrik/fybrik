package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"

	connectors "github.com/mesh-for-data/mesh-for-data/pkg/connectors/clients"
	openapiclientmodels "github.com/mesh-for-data/mesh-for-data/pkg/taxonomy/model/base"
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

	mainPolicyManagerURL := "opa-connector.m4d-system:80"
	policyManager, err := connectors.NewGrpcPolicyManager(mainPolicyManagerName, mainPolicyManagerURL, connectionTimeout)
	if err != nil {
		log.Println("returned with error ")
		log.Println("error in policyManager creation  %v", err)
		return
	}

	creds := "http://vault.m4d-system:8200/v1/kubernetes-secrets/wkc-creds?namespace=cp4d"

	// input := []openapiclientmodels.PolicymanagerRequest{*openapiclientmodels.NewPolicymanagerRequest(*openapiclientmodels.NewAction(openapiclientmodels.ActionType("read")), *openapiclientmodels.NewResource("{\"asset_id\": \"0bb3245e-e3ef-40b7-b639-c471bae4966c\", \"catalog_id\": \"503d683f-1d43-4257-a1a3-0ddf5e446ba5\"}", creds))} // []PolicymanagerRequest | input values that need to be considered for filter

	input := openapiclientmodels.NewPolicyManagerRequestWithDefaults()

	//reqCtx := openapiclientmodels.NewRequestContextWithDefaults()
	context := make(map[string]interface{})
	//reqCtx.SetIntent(openapiclientmodels.FRAUD_DETECTION)
	//reqCtx.SetRole(openapiclientmodels.DATA_SCIENTIST)
	context["Intent"] = "Fraud Detection"
	context["Role"] = "Data Scientist"
	input.SetContext(context)

	action := openapiclientmodels.NewPolicyManagerRequestActionWithDefaults()
	action.SetActionType(openapiclientmodels.READ)
	action.SetProcessingLocation("Netherlands")
	input.SetAction(*action)

	//input.SetAction(*openapiclientmodels.NewAction(openapiclientmodels.ActionType("read")))
	input.SetResource(*openapiclientmodels.NewResource("{\"asset_id\": \"0bb3245e-e3ef-40b7-b639-c471bae4966c\", \"catalog_id\": \"503d683f-1d43-4257-a1a3-0ddf5e446ba5\"}"))
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
