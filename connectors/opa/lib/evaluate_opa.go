// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package lib

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/buger/jsonparser"

	"github.com/tidwall/pretty"
)

func performHTTPReq(standardClient *http.Client, address string, httpMethod string, content string, contentType string) (*http.Response, error) {
	reqURL, err := url.Parse(address)
	if err != nil {
		return nil, err
	}
	if reqURL.Scheme != "http" && reqURL.Scheme != "https" {
		err := errors.New("opa server url scheme should be http or https")
		return nil, err
	}

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
		return nil, err
	}
	log.Println(httpMethod + " succeeded")

	return res, nil
}

func doesOpaHaveUserPoliciesLoaded(responsedata []byte) (string, bool) {
	decisionid, _ := jsonparser.GetString(responsedata, "decision_id")

	log.Printf("decision_id: %s", decisionid)
	if value, _, _, err := jsonparser.Get(responsedata, "result"); err == nil {
		log.Printf("result: %s", value)
	} else {
		log.Printf("Result Key does not exist implying no policies are loaded in opa")
		return decisionid, false
	}
	return decisionid, true
}

func EvaluatePoliciesOnInput(inputJSON string, opaServerURL string, policyToBeEvaluated string, standardClient *http.Client) (string, error) {
	log.Println("using opaServerURL in OPAConnector EvaluatePoliciesOnInput: ", opaServerURL)

	// input HTTP req
	httpMethod := "POST"
	log.Println("inputJSON in pretty print ")
	res1 := pretty.Pretty([]byte(inputJSON))
	log.Println("res = ", string(res1))

	contentType := "application/json"
	log.Println("opaServerURL")
	log.Println(opaServerURL)

	res, err := performHTTPReq(standardClient, opaServerURL+"/v1/data/"+policyToBeEvaluated, httpMethod, inputJSON, contentType)
	if err != nil {
		return "", err
	}
	data, _ := ioutil.ReadAll(res.Body)
	log.Printf("body from input http response: %s\n", data)
	log.Printf("status from input http response: %d\n", res.StatusCode)
	res.Body.Close()

	log.Println("responsestring data")
	log.Println(string(data))

	currentData := string(data)
	decisionid, flag := doesOpaHaveUserPoliciesLoaded(data)
	if !flag {
		// simulating Allow Enforcement Action. No result implies allow.
		currentData = "{\"decision_id\":\"" + decisionid + "\",\"result\": []}"
		log.Println("currentData - modified")
		log.Println(currentData)
	}

	return currentData, nil
}
