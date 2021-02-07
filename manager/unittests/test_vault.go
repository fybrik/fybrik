// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"log"
	"os"

	utils "github.com/ibm/the-mesh-for-data/manager/controllers/utils"
)

func main() {
	log.Printf("Testing vault lib dataset credentials")

	token := os.Getenv("VAULT_TOKEN")
	err := utils.MountDatasetVault(token)
	if err != nil {
		log.Fatalf("Error mounting dataset key value provider: %v", err)
		return
	}

	vaultClient, err := utils.InitVault(token)
	if err != nil {
		log.Fatalf("Error connecting to vault: %v", err)
		return
	}

	if vaultClient == nil {
		log.Fatalf("Null vaultClient")
		return
	}

	// Put credentials in vault
	credentials := map[string]interface{}{
		"username": "datasetuser1",
		"password": "myfavoritepassword",
	}

	secretaddr, err := utils.AddToVault("123/456", credentials, vaultClient)
	if err != nil {
		log.Fatalf(err.Error())
		return
	}
	log.Println("Credentials written.  Link to credentials in vault: " + secretaddr)

	readCredentials, err := utils.GetFromVault(secretaddr, vaultClient)
	if err != nil {
		log.Fatalf(err.Error())
		return
	}
	log.Println("Read credentials from vault: " + readCredentials)

	writtenCredentials, _ := json.Marshal(credentials)
	if readCredentials == string(writtenCredentials) {
		log.Println("Credentials read match those that were written!")
	} else {
		log.Println("Credentials read DO NOT match those that were written!")
		log.Println("   Written: " + string(writtenCredentials))
		log.Println("   Read: " + readCredentials)
	}

	badPath := utils.GetVaultDatasetHome() + "abc/def"
	log.Println("Now see what happens if we try to read something that does not exist: " + badPath)
	badCredentials, err := utils.GetFromVault(badPath, vaultClient)
	if err != nil {
		log.Println("   Received error AS EXPECTED: " + err.Error())
	} else {
		log.Println("   Hmm, we should have gotten an error when reading from " + badPath)
		log.Println("      Instead we got credentials: " + badCredentials)
	}

	// Testing for user credentials
	log.Println("\nTesting vault policy functions")
	err = utils.MountUserVault(token)
	if err != nil {
		log.Fatalf("Error mounting user key value provider: %v", err)
		return
	}

	// Put credentials in vault
	credentials2 := map[string]interface{}{
		"username": "systemXuser1",
		"password": "systemXpassword",
	}

	secretaddr2, err := utils.AddUserCredentialsToVault("namespace1", "computeX", "datacatalogY", credentials2, vaultClient)
	if err != nil {
		log.Fatalf(err.Error())
		return
	}
	log.Println("User Credentials written.  Link to credentials in vault: " + secretaddr2)

	readCredentials2, err := utils.GetFromVault(secretaddr2, vaultClient)
	if err != nil {
		log.Fatalf(err.Error())
		return
	}
	log.Println("Read user credentials from vault: " + readCredentials2)

	writtenCredentials2, _ := json.Marshal(credentials2)
	if readCredentials2 == string(writtenCredentials2) {
		log.Println("Credentials read match those that were written!")
	} else {
		log.Println("Credentials read DO NOT match those that were written!")
		log.Println("   Written: " + string(writtenCredentials2))
		log.Println("   Read: " + readCredentials2)
	}

	err = utils.DeleteFromVault(secretaddr2, vaultClient)
	if err != nil {
		log.Fatalf(err.Error())
		return
	}
	log.Println("Deleted user credentials from vault: " + secretaddr2)

	// Testing for vault authentication and policies
	log.Println("\nTesting vault policy functions")

	// Create and save a vault policy
	path := "/test-identity"
	policy := "path \"" + path + "\"" + " {\n	capabilities = [\"read\"]\n }"
	policyName := "testpolicy"
	log.Println("   Vault policy: " + policy)

	err = utils.WriteVaultPolicy(policyName, policy, vaultClient)
	if err != nil {
		log.Println("      Failed writing policy: " + err.Error())
	} else {
		log.Println("      Succeeded writing policy")
	}

	/*  Should be done during cluster config 	*/
	err = utils.InitVaultAuth(vaultClient)
	if err != nil {
		log.Println("      Failed initializing vault authentication: " + err.Error())
	} else {
		log.Println("      Succeeded initializing vault authentication")
	}

	identity := "/role/demo"
	err = utils.LinkVaultPolicyToIdentity(identity, policyName, vaultClient)
	if err != nil {
		log.Println("      Failed adding policy to identity " + identity + ": " + err.Error())
	} else {
		log.Println("      Succeeded adding policy to identity")
	}

	err = utils.RemoveVaultPolicyFromIdentity(identity, policyName, vaultClient)
	if err != nil {
		log.Println("      Failed removing policy from identity " + identity + ": " + err.Error())
	} else {
		log.Println("      Succeeded removing policy from identity")
	}

	err = utils.DeleteVaultPolicy("testpolicy", vaultClient)
	if err != nil {
		log.Println("      Failed deleting policy: " + err.Error())
	} else {
		log.Println("      Succeeded deleting policy")
	}
}
