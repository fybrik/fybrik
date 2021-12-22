// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"fybrik.io/fybrik/manager/controllers/mockup"
	datacatalogTaxonomyModels "fybrik.io/fybrik/pkg/taxonomy/model/datacatalog/base"
	"github.com/gin-gonic/gin"
)

const (
	PORT = 8080
)

var router *gin.Engine

func main() {
	router = gin.Default()

	router.POST("/getAssetInfo", func(c *gin.Context) {
		creds := ""
		if values := c.Request.Header["X-Request-DataCatalog-Cred"]; len(values) > 0 {
			creds = values[0]
		}
		log.Println("creds extracted from POST request in mockup data catalog:", creds)
		input, _ := ioutil.ReadAll(c.Request.Body)
		log.Println("input extracted from POST request body in mockup data catalog:", string(input))

		var dataCatalogReq datacatalogTaxonomyModels.DataCatalogRequest
		err := json.Unmarshal(input, &dataCatalogReq)
		if err != nil {
			return
		}
		dataCatalog := mockup.NewTestCatalog()
		dataCatalogResp, err := dataCatalog.GetAssetInfo(&dataCatalogReq, creds)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error in getAssetInfo!")
			return
		}
		c.JSON(http.StatusOK, dataCatalogResp)
	})

	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Serving REST APIs as part of data catalog stub")
	})

	log.Fatal(router.Run(":" + strconv.Itoa(PORT)))
}
