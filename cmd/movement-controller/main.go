// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"github.com/ibm/the-mesh-for-data/manager/controllers/motion"
	"github.com/ibm/the-mesh-for-data/manager/controllers/utils"
	"os"

	"sigs.k8s.io/controller-runtime/pkg/manager"

	motionv1 "github.com/ibm/the-mesh-for-data/manager/apis/motion/v1alpha1"
	kruntime "k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = kruntime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = motionv1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

// This manager only manages BatchTransfers and StreamTransfers
// and is based on the controllers that are used for the regular
// manager. It is meant to be used for data transfers only.
func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var namespace string
	address := utils.ListeningAddress(8085)
	flag.StringVar(&metricsAddr, "metrics-addr", address, "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&namespace, "namespace", "", "The namespace to which this controller manager is limited.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	var ctrlOps manager.Options

	if len(namespace) > 0 {
		// manager restricted to a single namespace
		ctrlOps = ctrl.Options{
			Scheme:             scheme,
			Namespace:          namespace,
			MetricsBindAddress: metricsAddr,
			LeaderElection:     enableLeaderElection,
			LeaderElectionID:   "m4d-operator-leader-election",
			Port:               9443,
		}
	} else {
		// manager not restricted to a namespace.
		ctrlOps = ctrl.Options{
			Scheme:             scheme,
			MetricsBindAddress: metricsAddr,
			LeaderElection:     enableLeaderElection,
			LeaderElectionID:   "m4d-operator-leader-election",
			Port:               9443,
		}
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrlOps)

	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	motion.SetupMotionControllers(mgr)

	// +kubebuilder:scaffold:builder

	setupLog.Info("starting movement-controller")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
