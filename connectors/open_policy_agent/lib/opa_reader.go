// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package lib

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	openapiclientmodels "fybrik.io/fybrik/pkg/taxonomy/model/base"
)

type OpaReader struct {
	opaServerURL string
	opaClient    *http.Client
}

func NewOpaReader(opasrvurl string, client *http.Client) *OpaReader {
	return &OpaReader{opaServerURL: opasrvurl, opaClient: client}
}

func (r *OpaReader) updatePolicyManagerRequestWithResourceInfo(in *openapiclientmodels.PolicyManagerRequest, catalogMetadata map[string]interface{}) (*openapiclientmodels.PolicyManagerRequest, error) {
	responseBytes, errJSON := json.MarshalIndent(catalogMetadata, "", "\t")
	if errJSON != nil {
		return nil, fmt.Errorf("error Marshalling External Catalog Connector Response: %v", errJSON)
	}

	var catalogJSON interface{}
	err := json.Unmarshal(responseBytes, &catalogJSON)
	if err != nil {
		return nil, fmt.Errorf("error UnMarshalling WKC Catalog Connector Response: %v", err)
	}
	if main, ok := catalogJSON.(map[string]interface{}); ok {
		if details, ok := main["details"].(map[string]interface{}); ok {
			if metadata, ok := details["metadata"].(map[string]interface{}); ok {
				if datasetTags, ok := metadata["dataset_tags"].([]interface{}); ok {
					tagArr := make([]string, 0)
					for i := 0; i < len(datasetTags); i++ {
						tagVal := datasetTags[i].(string)
						tagArr = append(tagArr, tagVal)
					}
					log.Println("tagArr: ", tagArr)

					tagVal := make(map[string]interface{})
					for i := 0; i < len(tagArr); i++ {
						splitStr := strings.Split(tagArr[i], " = ")
						// residency = Turkey
						tagVal[splitStr[0]] = splitStr[1]
					}
					log.Println("tagVal: ", tagVal)
					resource := in.GetResource()
					(&resource).SetTags(tagVal)
					in.SetResource(resource)
					log.Println("in.GetResource().GetTags(): ", (&resource).GetTags())
				}
				if componentsMetadata, ok := metadata["components_metadata"].(map[string]interface{}); ok {
					listofcols := []string{}
					listoftags := [][]string{}
					lstOfValueTags := []string{}
					for key, val := range componentsMetadata {
						log.Println("key :", key)
						log.Println("val :", val)
						listofcols = append(listofcols, key)

						if columnsMetadata, ok := val.(map[string]interface{}); ok {
							if tagsList, ok := columnsMetadata["tags"].([]interface{}); ok {
								for _, tagElem := range tagsList {
									lstOfValueTags = append(lstOfValueTags, tagElem.(string))
								}
								listoftags = append(listoftags, lstOfValueTags)
							} else {
								lstOfValueTags = []string{}
								listoftags = append(listoftags, lstOfValueTags)
							}
						}
					}
					log.Println("******** listofcols : *******", listofcols)
					log.Println("******** listoftags: *******", listoftags)

					cols := []openapiclientmodels.ResourceColumns{}

					var newcol *openapiclientmodels.ResourceColumns
					numOfCols := len(listofcols)
					numOfTags := 0
					for i := 0; i < numOfCols; i++ {
						newcol = new(openapiclientmodels.ResourceColumns)
						newcol.SetName(listofcols[i])
						numOfTags = len(listoftags[i])
						if numOfTags > 0 {
							q := make(map[string]interface{})
							for j := 0; j < len(listoftags[i]); j++ {
								q[listoftags[i][j]] = "true"
							}
							log.Println("set tags of col:", listofcols[i], " to:", q)
							newcol.SetTags(q)
						}
						cols = append(cols, *newcol)
					}
					log.Println("******** cols : *******")
					log.Println("cols=", cols)
					for i := 0; i < numOfCols; i++ {
						log.Println("cols=", cols[i].GetName())
						log.Println("cols=", cols[i].GetTags())
					}
					log.Println("******** in before: *******", *in)
					log.Println("******** res before: *******", in.Resource)
					res := in.Resource
					(&res).SetColumns(cols)
					in.SetResource(res)
					log.Println("******** res after: *******", res)
					log.Println("******** in after: *******", *in)

					log.Println("******** updated policy manager resp object : *******")
					b, err := json.Marshal(*in)
					if err != nil {
						fmt.Println(err)
						return nil, err
					}
					fmt.Println("stringified policy manager request", string(b))
					log.Println("******** updated policy manager resp object end : *******")
				}
			}
		}
	}
	return in, nil
}

func (r *OpaReader) GetOPADecisions(in *openapiclientmodels.PolicyManagerRequest, creds string, catalogReader *CatalogReader, policyToBeEvaluated string) (openapiclientmodels.PolicyManagerResponse, error) {
	datasetsMetadata, err := catalogReader.GetDatasetsMetadataFromCatalog(in, creds)
	if err != nil {
		return openapiclientmodels.PolicyManagerResponse{}, err
	}
	datasetID := in.GetResource().Name
	metadata := datasetsMetadata[datasetID]

	inputMap, ok := metadata.(map[string]interface{})
	if !ok {
		return openapiclientmodels.PolicyManagerResponse{}, fmt.Errorf("error in unmarshalling dataset metadata (datasetID = %s): %v", datasetID, err)
	}

	in, _ = r.updatePolicyManagerRequestWithResourceInfo(in, inputMap)

	b, err := json.Marshal(*in)
	if err != nil {
		fmt.Println(err)
		return openapiclientmodels.PolicyManagerResponse{}, fmt.Errorf("error during marshal in GetOPADecisions: %v", err)
	}
	inputJSON := "{ \"input\": " + string(b) + " }"
	fmt.Println("updated stringified policy manager request in GetOPADecisions", inputJSON)

	opaEval, err := EvaluatePoliciesOnInput(inputJSON, r.opaServerURL, policyToBeEvaluated, r.opaClient)
	if err != nil {
		log.Printf("error in EvaluatePoliciesOnInput : %v", err)
		return openapiclientmodels.PolicyManagerResponse{}, fmt.Errorf("error in EvaluatePoliciesOnInput : %v", err)
	}
	log.Println("OPA Eval : " + opaEval)

	policyManagerResponse := new(openapiclientmodels.PolicyManagerResponse)
	err = json.Unmarshal([]byte(opaEval), &policyManagerResponse)
	if err != nil {
		return openapiclientmodels.PolicyManagerResponse{}, fmt.Errorf("error in GetOPADecisions during unmarshalling OPA response to Policy Manager Response : %v", err)
	}
	log.Println("unmarshalled policyManagerResp in GetOPADecisions:", policyManagerResponse)

	res, err := json.MarshalIndent(policyManagerResponse, "", "\t")
	if err != nil {
		return openapiclientmodels.PolicyManagerResponse{}, fmt.Errorf("error in GetOPADecisions during MarshalIndent Policy Manager Response : %v", err)
	}
	log.Println("Marshalled PolicyManagerResponse from OPA response in GetOPADecisions:", string(res))

	return *policyManagerResponse, nil
}

// func buildNewEnforcementAction(transformAction interface{}) (*openapiclientmodels.ActionOnColumns, bool) {
// 	log.Println("transformAction", transformAction)
// 	var actionOnColumns = new(openapiclientmodels.ActionOnColumns)
// 	if result1, ok := transformAction.(map[string]interface{}); ok {
// 		log.Println("transformAction type :", reflect.TypeOf(result1))
// 		log.Println("result1[\"action\"].(string) :", result1["action"].(map[string]interface{}))
// 		if result, ok := result1["action"].(map[string]interface{}); ok {
// 			res1 := result["name"].(string)
// 			switch res1 {
// 			case string(openapiclientmodels.REMOVE_COLUMN):
// 				actionOnColumns.SetName(openapiclientmodels.REMOVE_COLUMN)
// 				log.Println("Name:", openapiclientmodels.REMOVE_COLUMN)

// 				resCols := result["columns"].([]interface{})
// 				log.Println("resCols", resCols)
// 				lstOfCols := []string{}
// 				for i := 0; i < len(resCols); i++ {
// 					lstOfCols = append(lstOfCols, resCols[i].(string))
// 				}
// 				log.Println("lstOfCols", lstOfCols)
// 				actionOnColumns.SetColumns(lstOfCols)

// 				return actionOnColumns, true

// 			case string(openapiclientmodels.ENCRYPT_COLUMN):
// 				actionOnColumns.SetName(openapiclientmodels.ENCRYPT_COLUMN)
// 				log.Println("Name:", openapiclientmodels.ENCRYPT_COLUMN)

// 				resCols := result["columns"].([]interface{})
// 				log.Println("resCols", resCols)
// 				lstOfCols := []string{}
// 				for i := 0; i < len(resCols); i++ {
// 					lstOfCols = append(lstOfCols, resCols[i].(string))
// 				}
// 				log.Println("lstOfCols", lstOfCols)
// 				actionOnColumns.SetColumns(lstOfCols)

// 				return actionOnColumns, true

// 			case string(openapiclientmodels.REDACT_COLUMN):
// 				actionOnColumns.SetName(openapiclientmodels.REDACT_COLUMN)
// 				log.Println("Name:", openapiclientmodels.REDACT_COLUMN)

// 				resCols := result["columns"].([]interface{})
// 				log.Println("resCols", resCols)
// 				lstOfCols := []string{}
// 				for i := 0; i < len(resCols); i++ {
// 					lstOfCols = append(lstOfCols, resCols[i].(string))
// 				}
// 				log.Println("lstOfCols", lstOfCols)
// 				actionOnColumns.SetColumns(lstOfCols)

// 				return actionOnColumns, true

// 			case string(openapiclientmodels.PERIODIC_BLACKOUT):
// 				//if monthlyDaysNum, ok := extractArgument(action["arguments"], "monthly_days_end"); ok {
// 				actionOnColumns.SetName(openapiclientmodels.PERIODIC_BLACKOUT)
// 				log.Println("Name:", openapiclientmodels.PERIODIC_BLACKOUT)

// 				resCols := result["columns"].([]interface{})
// 				log.Println("resCols", resCols)
// 				lstOfCols := []string{}
// 				for i := 0; i < len(resCols); i++ {
// 					lstOfCols = append(lstOfCols, resCols[i].(string))
// 				}
// 				log.Println("lstOfCols", lstOfCols)
// 				actionOnColumns.SetColumns(lstOfCols)

// 				return actionOnColumns, true
// 				//}
// 				//else if yearlyDaysNum, ok := extractArgument(action["arguments"], "yearly_days_end"); ok {
// 				// actionOnColumns.SetName(openapiclientmodels.PERIODIC_BLACKOUT)
// 				// actionOnColumns.SetColumns(result["columns"].([]string))
// 				// return actionOnColumns, true
// 				//}

// 			default:
// 				log.Printf("Unknown Enforcement Action receieved from OPA")
// 			}
// 		}
// 	}
// 	return nil, false
// }

// func extractArgument(arguments interface{}, argName string) (string, bool) {
// 	if argsMap, ok := arguments.(map[string]interface{}); ok {
// 		if value, ok := argsMap[argName].(string); ok {
// 			return value, true
// 		}
// 	}
// 	return "", false
// }

// func buildNewPolicy(usedPolicy interface{}) (*string, bool) {
// 	log.Println("in buildNewPolicy")
// 	if policy, ok := usedPolicy.(map[string]interface{}); ok {
// 		//todo: add other fields that can be returned as part of the policy struct
// 		if description, ok := policy["policy"].(string); ok {
// 			newUsedPolicy := description
// 			return &newUsedPolicy, true
// 		}
// 	}

// 	return nil, false
// }
