// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package lib

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	clients "fybrik.io/fybrik/pkg/connectors/datacatalog/clients"
	datacatalogTaxonomyModels "fybrik.io/fybrik/pkg/taxonomy/model/datacatalog/base"
	policymanagerTaxonomyModels "fybrik.io/fybrik/pkg/taxonomy/model/policymanager/base"
)

type OpaReader struct {
	opaServerURL string
	opaClient    *http.Client
	dataCatalog  *clients.DataCatalog
}

func NewOpaReader(opasrvurl string, client *http.Client, dataCatalog *clients.DataCatalog) *OpaReader {
	return &OpaReader{opaServerURL: opasrvurl, opaClient: client, dataCatalog: dataCatalog}
}

func (r *OpaReader) updatePolicyManagerRequestWithResourceInfo(in *policymanagerTaxonomyModels.PolicyManagerRequest, catalogMetadata *datacatalogTaxonomyModels.DataCatalogResponse) (*policymanagerTaxonomyModels.PolicyManagerRequest, error) {
	// just printing - start
	responseBytes, errJSON := json.MarshalIndent(catalogMetadata.ResourceMetadata, "", "\t")
	if errJSON != nil {
		return nil, fmt.Errorf("error Marshalling catalogMetadata in updatePolicyManagerRequestWithResourceInfo2: %v", errJSON)
	}
	log.Print("catalogMetadata.ResourceMetadata after MarshalIndent in updatePolicyManagerRequestWithResourceInfo2:" + string(responseBytes))
	// just printing - end

	err := json.Unmarshal(responseBytes, &in.Resource)
	if err != nil {
		return nil, fmt.Errorf("error UnMarshalling in updatePolicyManagerRequestWithResourceInfo2: %v", err)
	}

	// just printing - start
	responseBytes, errJSON = json.MarshalIndent(in, "", "\t")
	if errJSON != nil {
		return nil, fmt.Errorf("error Marshalling taxonomymodels.PolicyManagerRequest in updatePolicyManagerRequestWithResourceInfo2: %v", errJSON)
	}
	log.Print("returning updated taxonomymodels.PolicyManagerRequest in updatePolicyManagerRequestWithResourceInfo2:" + string(responseBytes))
	// just printing - end

	return in, nil
}

func (r *OpaReader) GetOPADecisions(in *policymanagerTaxonomyModels.PolicyManagerRequest, creds string, policyToBeEvaluated string) (policymanagerTaxonomyModels.PolicyManagerResponse, error) {
	datasetID := (in.GetResource()).Name
	objToSend := datacatalogTaxonomyModels.DataCatalogRequest{AssetID: datasetID, OperationType: datacatalogTaxonomyModels.READ}

	info, err := (*r.dataCatalog).GetAssetInfo(&objToSend, creds)
	// info, err := (*r.DataCatalog).GetDatasetInfo(context.Background(), objToSend)
	if err != nil {
		return policymanagerTaxonomyModels.PolicyManagerResponse{}, err
	}

	log.Printf("Received Response from External Catalog Connector for  dataSetID: %s\n", datasetID)
	log.Printf("Response received from External Catalog Connector is given below:")
	responseBytes, errJSON := json.MarshalIndent(info, "", "\t")
	if errJSON != nil {
		return policymanagerTaxonomyModels.PolicyManagerResponse{}, fmt.Errorf("error Marshalling External Catalog Connector Response: %v", errJSON)
	}
	log.Print(string(responseBytes))

	in, _ = r.updatePolicyManagerRequestWithResourceInfo(in, info)

	b, err := json.Marshal(*in)
	if err != nil {
		fmt.Println(err)
		return policymanagerTaxonomyModels.PolicyManagerResponse{}, fmt.Errorf("error during marshal in GetOPADecisions: %v", err)
	}
	inputJSON := "{ \"input\": " + string(b) + " }"
	fmt.Println("updated stringified policy manager request in GetOPADecisions", inputJSON)

	opaEval, err := EvaluatePoliciesOnInput(inputJSON, r.opaServerURL, policyToBeEvaluated, r.opaClient)
	if err != nil {
		log.Printf("error in EvaluatePoliciesOnInput : %v", err)
		return policymanagerTaxonomyModels.PolicyManagerResponse{}, fmt.Errorf("error in EvaluatePoliciesOnInput : %v", err)
	}
	log.Println("OPA Eval : " + opaEval)

	policyManagerResponse := new(policymanagerTaxonomyModels.PolicyManagerResponse)
	err = json.Unmarshal([]byte(opaEval), &policyManagerResponse)
	if err != nil {
		return policymanagerTaxonomyModels.PolicyManagerResponse{}, fmt.Errorf("error in GetOPADecisions during unmarshalling OPA response to Policy Manager Response : %v", err)
	}
	log.Println("unmarshalled policyManagerResp in GetOPADecisions:", policyManagerResponse)

	res, err := json.MarshalIndent(policyManagerResponse, "", "\t")
	if err != nil {
		return policymanagerTaxonomyModels.PolicyManagerResponse{}, fmt.Errorf("error in GetOPADecisions during MarshalIndent Policy Manager Response : %v", err)
	}
	log.Println("Marshalled PolicyManagerResponse from OPA response in GetOPADecisions:", string(res))

	return *policyManagerResponse, nil
}
