// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package connector

import "github.com/gin-gonic/gin"

// NewRouter returns a new router.
func NewRouter(controller *ConnectorController) *gin.Engine {
	router := gin.Default()
	router.POST("/getAssetInfo", controller.getAssetInfo)
	return router
}
