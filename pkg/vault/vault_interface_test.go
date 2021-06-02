// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package vault

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	utils "github.com/mesh-for-data/mesh-for-data/manager/controllers/utils"
	"github.com/stretchr/testify/assert"
)

func Log(t *testing.T, label string, err error) {
	if err == nil {
		err = fmt.Errorf("succeeded")
	}
	t.Logf("%s: %s", label, err)
}

func TestCredentialManagerInterface(t *testing.T) {
	var err error
	vaultToken := "dummyToken"
	// TODO add environment variables to test against a real vault instance
	os.Setenv("RUN_WITHOUT_VAULT", "1")
	conn, err := InitConnection(utils.GetVaultAddress(), vaultToken)
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
	err = conn.AddSecret("m4d/test/123/456", credentials)
	assert.Nil(t, err)
	Log(t, "Add secret", err)
	readCredentials, errGet := conn.GetSecret("m4d/test/123/456")
	assert.Nil(t, errGet)
	Log(t, "Get secret", errGet)
	writtenCredentials, _ := json.Marshal(credentials)
	assert.Equal(t, string(writtenCredentials), readCredentials, "Read "+readCredentials+" instead of "+string(writtenCredentials))
	err = conn.DeleteSecret("m4d/test/123/456")
	assert.Nil(t, err)
	Log(t, "Delete secret", err)
	_, err = conn.GetSecret("m4d/test/123/456")
	assert.NotNil(t, err)
}
