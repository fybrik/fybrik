// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package optimizer

import (
	"testing"

	app "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/pkg/multicluster"
)

func TestBuildModel(t *testing.T) {
	mod := app.FybrikModule{}
	moduleMap := map[string]*app.FybrikModule{"mod1": &mod, "mod2": &mod}
	cluster := multicluster.Cluster{}
	clusters := []multicluster.Cluster{cluster}
	dataInfo := DataInfo{moduleMap, clusters}
	dpCSP := NewDataPathCSP(&dataInfo)
	err := dpCSP.BuildFzModel(3)
	if err != nil {
		t.Errorf("Failed building a CSP model: %s", err)
	}
}
