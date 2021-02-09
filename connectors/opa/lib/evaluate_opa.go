// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package lib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/hashicorp/go-retryablehttp"
)

func performHTTPReq(standardClient *http.Client, address string, httpMethod string, content string, contentType string) *http.Response {
	reqURL, _ := url.Parse(address)

	reqBody := ioutil.NopCloser(strings.NewReader(content))

	req := &http.Request{
		Method: httpMethod,
		URL:    reqURL,
		Header: map[string][]string{
			"Content-Type": {contentType + "; charset=UTF-8"},
		},
		Body: reqBody,
	}

	log.Println("req in OPAConnector performHTTPReq: ", req)
	log.Println("httpMethod in OPAConnector performHTTPReq: ", httpMethod)
	log.Println("reqURL in OPAConnector performHTTPReq: ", reqURL)
	log.Println("reqBody in OPAConnector performHTTPReq: ", reqBody)
	res, err := standardClient.Do(req)

	if err != nil {
		log.Fatal("Error:", err)
	}
	fmt.Println(httpMethod + " succeeded")

	return res
}

func EvaluatePoliciesOnInput(inputMap map[string]interface{}, opaServerURL string, policyToBeEvaluated string) (string, error) {
	if !strings.HasPrefix(opaServerURL, "http://") {
		opaServerURL = "http://" + opaServerURL + "/"
	}
	if !strings.HasSuffix(opaServerURL, "/") {
		opaServerURL += "/"
	}
	log.Println("using opaServerURL in OPAConnector EvaluatePoliciesOnInput: ", opaServerURL)

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 10
	standardClient := retryClient.HTTPClient // *http.Client

	// input HTTP req
	httpMethod := "POST"
	toPrintBytes, _ := json.MarshalIndent(inputMap, "", "\t")
	inputJSON := "{ \"input\": " + string(toPrintBytes) + " }"
	log.Println("inputJSON")
	log.Println(inputJSON)
	contentType := "application/json"
	log.Println("opaServerURL")
	log.Println(opaServerURL)

	res := performHTTPReq(standardClient, opaServerURL+"v1/data/"+policyToBeEvaluated, httpMethod, inputJSON, contentType)
	data, _ := ioutil.ReadAll(res.Body)
	fmt.Printf("body from input http response: %s\n", data)
	fmt.Printf("status from input http response: %d\n", res.StatusCode)
	res.Body.Close()

	log.Println("responsestring data")
	log.Println(string(data))

	return string(data), nil
}
