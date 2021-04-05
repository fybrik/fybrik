// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package taxonomy

/*
{
	# 	"name": "<name>"
	# 	"destination": "<destination>",
	# 	"processing_geography": "<processing_geography>",
	# 	"purpose": "<purpose>",
	# 	"role": "<role>",
	# 	"type": "<access type>",
	# 	"details": {
	# 		"data_format": "<data_format>",
	# 		"data_store": {
	# 			"name": "<datastore name>"
	# 		},
	# 		"geo": "<geo>",
	# 		"metadata": {
	# 			"components_metadata": {
	# 				"<column name1>": {
	# 					"component_type": "column",
	# 					"named_metadata": {
	# 						"type": "length=10.0,nullable=true,type=date,scale=0.0,signed=false"
	# 					}
	# 				},
	# 				"<column name2>": {
	# 					"component_type": "column",
	# 					"named_metadata": {
	# 						"type": "length=3.0,nullable=true,type=char,scale=0.0,signed=false"
	# 					},
	# 					"tags": ["<tag1>", "<tag2>"]
	# 				}
	# 			},
	# 			"dataset_named_metadata": {
	# 				"<term1 name>": "<term1 value>",
	# 				"<term2 name>": "<term2 value>"
	# 			},
	# 			"dataset_tags": [
	# 				"<tag1>",
	# 				"<tag2>"
	# 			]
	# 		},
	# 	}
	# }
*/
/*
type Requester struct {
	Intent string `json:"intent"`
	Role   string `json:"role"`
}

type Action struct {
	ActionType         string `json:"actionType"`
	ProcessingLocation string `json:"processingLocation,optional"`
}

type PolicyReq struct {
	TheRequester Requester `json:"requester"`
	TheAction    Action    `json:"action"`
}

var (
	taxonomyConfigmapName = "taxonomy-configmap"
	catalogTaxName        = "data_catalog_schema.json"
	policyTaxName         = "policy_manager_schema.json"
	moduleTaxName         = "module_schema.json"

	actionGood    = Action{ActionType: "read", ProcessingLocation: "Netherlands"}
	requesterGood = Requester{Intent: "Fraud Detection", Role: "Data Scientist"}
	policyReqGood = PolicyReq{TheRequester: requesterGood, TheAction: actionGood}
	//	policyReqBad  = "{\"requester\": {\"role\": \"Data Scientist\"}, \"action_type\": {\"type\": \"blabla\"}}"

//	test2 = "{\"apiVersion\": \"app.m4d.ibm.com/v1alpha1\",\"kind\": \"M4DApplication\",\"metadata\": {\"name\": \"unittest-read\"},\"spec\": {\"selector\": {\"workloadSelector\": {\"matchLabels\":{\"app\": \"notebook\"}}},\"appInfo\": {\"purpose\": \"fraud-detection\",\"role\": \"Security\"}, \"data\": [{\"dataSetID\": \"123\",\"requirements\": { \"interface\": {\"protocol\": \"s3\",\"dataformat\": \"parquet\"}}}]}}"
)

func validateJSON(t *testing.T, jsonData string, taxonomyfile string, expectedValid bool) {
	path, err := filepath.Abs(taxonomyfile)
	assert.Nil(t, err)

	taxonomyLoader := gojsonschema.NewReferenceLoader("file://" + path)
	documentLoader := gojsonschema.NewStringLoader(jsonData)
	result, err := gojsonschema.Validate(taxonomyLoader, documentLoader)
	assert.Nil(t, err)

	fmt.Printf("Valid document: %t\n", result.Valid())

	if expectedValid {
		assert.True(t, result.Valid())
	} else {
		fmt.Printf("The document is not valid.  Discrepencies: \n")
		for _, disc := range result.Errors() {
			fmt.Printf("- %s\n", disc)
		}
		assert.False(t, result.Valid())
	}

}

func TestTaxonomy(t *testing.T) {
	jsonPolicyReqGood, err := json.Marshal(policyReqGood)
	validateJSON(t, toString(jsonPolicyReqGood), policyTaxName, true)
	//	validateJSON(t, policyReqBad, policyTaxName, false)
}
*/
