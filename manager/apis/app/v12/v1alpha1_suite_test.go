// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package v12

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var cfg *rest.Config
var c client.Client
var testEnv *envtest.Environment

func TestMain(m *testing.M) {
	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	if os.Getenv("USE_EXISTING_CONTROLLER") == "true" {
		fmt.Printf("apis/app/v12: Using existing environment; don't load CRDs. \n")
		useexistingcluster := true
		testEnv = &envtest.Environment{
			UseExistingCluster: &useexistingcluster,
		}
	} else {
		fmt.Printf("apis/app/v12: Using fake environment; so set path to CRDs so they are installed. \n")
		testEnv = &envtest.Environment{
			CRDDirectoryPaths: []string{
				filepath.Join(path, "..", "..", "..", "..", "charts", "fybrik-crd", "templates"),
			},
			ErrorIfCRDPathMissing: true,
		}
	}

	err = SchemeBuilder.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Fatal(err)
	}

	if cfg, err = testEnv.Start(); err != nil {
		log.Fatal(err)
	}

	if c, err = client.New(cfg, client.Options{Scheme: scheme.Scheme}); err != nil {
		log.Fatal(err)
	}

	code := m.Run()
	_ = testEnv.Stop()
	os.Exit(code)
}
