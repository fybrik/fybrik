package vaultconnbl

import (
	"fmt"
	"log"
	"net"
	"os"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/http"
	"github.com/hashicorp/vault/vault"

	tu "github.com/ibm/the-mesh-for-data/connectors/vault/testutil"
	vltutils "github.com/ibm/the-mesh-for-data/connectors/vault/vault_utils"
	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
)

// TestVaultConnectorNormalRun tests the execution of GetCredentialsInfo with dummy mock id and
// dummy credentials stored in the runtime test instance of vault
func TestVaultConnectorNormalRun(t *testing.T) {
	ln, client := createTestVault(t)
	defer ln.Close()

	srv := &Server{}
	appID := "mock-appID"
	//datasetID := "{\"ServerName\":\"cocoMDS3\",\"AssetGuid\":\"f6d9bf8c-dd37-4747-bca9-8ca9d1a5bb8f\"}"
	datasetID := "mock-datasetID"
	objToSendForCredential := &pb.DatasetCredentialsRequest{AppId: appID, DatasetId: datasetID}

	config := vltutils.VaultConfig{
		Token:   "token",
		Address: "address",
	}
	connection := vltutils.VaultConnection{
		Config: config,
	}

	data := map[string]interface{}{
		"dataset_id":  datasetID,
		"credentials": "my_egeria_credentials_test",
	}

	_, err := client.Logical().Write("secret/"+datasetID, data)
	if err != nil {
		t.Errorf("Error writing credentials from vault for " + datasetID + ":" + err.Error())
	}

	connection.Client = client
	userVaultPath := tu.GetEnvironment()
	var vault vltutils.VaultConnection
	vault = connection
	credentialsInfo, _ := srv.GetCredentialsInfo(objToSendForCredential, vault, userVaultPath)
	log.Println("credentialsInfo")
	log.Println(credentialsInfo)

	expectedCredentials := tu.GetExpectedVaultCredentials(objToSendForCredential)
	tu.EnsureDeepEqualCredentials(t, credentialsInfo, expectedCredentials)
}

func createTestVault(t *testing.T) (net.Listener, *api.Client) {
	t.Helper()

	// Create an in-memory, unsealed core (the "backend", if you will).
	core, keyShares, rootToken := vault.TestCoreUnsealed(t)
	_ = keyShares

	// Start an HTTP server for the core.
	ln, addr := http.TestServer(t, core)

	// Create a client that talks to the server, initially authenticating with
	// the root token.
	conf := api.DefaultConfig()
	conf.Address = addr

	client, err := api.NewClient(conf)
	if err != nil {
		t.Fatal(err)
	}
	client.SetToken(rootToken)

	return ln, client
}

func TestMain(m *testing.M) {
	fmt.Println("TestMain function called = vault_connector_test ")
	tu.EnvValues["USER_VAULT_PATH"] = "secret"

	code := m.Run()
	fmt.Println("TestMain function called after Run = vault_connector_test ")
	os.Exit(code)
}
