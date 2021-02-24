// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package vault

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"github.com/stretchr/testify/assert"
	utils "github.com/ibm/the-mesh-for-data/manager/controllers/utils"
)

func Log(t *testing.T, label string, err error) {
	if err == nil {
		err = fmt.Errorf("succeeded")
	}
	t.Logf("%s: %s", label, err)
}


func TestHelmCache(t *testing.T) {
	var err error
	os.Setenv("RUN_WITHOUT_VAULT","0")
	conn,err := InitConnection(utils.GetVaultAddress(),utils.GetVaultToken())
	assert.Nil(t, err)
	Log(t, "init vault", err)
	err = conn.Mount("v1/sys/mounts/m4d/test")
	assert.Nil(t, err)
	Log(t, "mount", err)
	// Put credentials in vault
	credentials := map[string]interface{}{
		"username": "datasetuser1",
		"password": "myfavoritepassword",
	}
	err = conn.AddSecret("m4d/test/123/456",credentials)
	assert.Nil(t, err)
	Log(t, "Add secret", err)
	readCredentials :=""
	readCredentials,err = conn.GetSecret("m4d/test/123/456")
	assert.Nil(t, err)
	Log(t, "Get secret", err)
	writtenCredentials, _ := json.Marshal(credentials)
	assert.Equal(t,string(writtenCredentials),readCredentials,"Read " + readCredentials + " instead of " + string(writtenCredentials))
	err = conn.DeleteSecret("m4d/test/123/456")
	assert.Nil(t, err)
	Log(t, "Delete secret", err)
	readCredentials,err = conn.GetSecret("m4d/test/123/456")
	assert.NotNil(t, err)
}
