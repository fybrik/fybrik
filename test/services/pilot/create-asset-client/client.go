package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"fybrik.io/fybrik/pkg/model/datacatalog"
)

type Response struct {
	ID     string `json:"id"`
	Joke   string `json:"joke"`
	Status int    `json:"status"`
}

func main() {
	request := datacatalog.CreateAssetRequest{
		DestinationCatalogID: "test",
	}

	//Encode the data
	postBody, _ := json.Marshal(request)
	responseBody := bytes.NewBuffer(postBody)

	fmt.Println("Calling API...")
	client := &http.Client{}
	req, err := http.NewRequest("POST", "http://katalog-connector:80/createAssetInfo", responseBody)
	if err != nil {
		fmt.Print(err.Error())
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Print(err.Error())
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err.Error())
	}
	var responseObject datacatalog.CreateAssetResponse
	json.Unmarshal(bodyBytes, &responseObject)
	fmt.Printf("API Response as struct %+v\n", responseObject)
}
