// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"github.com/mesh-for-data/mesh-for-data/pkg/multicluster"
	"github.com/mesh-for-data/mesh-for-data/pkg/multicluster/local"
	"github.com/mesh-for-data/mesh-for-data/pkg/multicluster/razee"
	"github.com/mesh-for-data/mesh-for-data/pkg/storage"

	"github.com/mesh-for-data/mesh-for-data/manager/controllers/motion"

	kruntime "k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	comv1alpha1 "github.com/datashim-io/datashim/src/dataset-operator/pkg/apis/com/v1alpha1"
	appv1 "github.com/mesh-for-data/mesh-for-data/manager/apis/app/v1alpha1"
	motionv1 "github.com/mesh-for-data/mesh-for-data/manager/apis/motion/v1alpha1"
	"github.com/mesh-for-data/mesh-for-data/manager/controllers/app"
	"github.com/mesh-for-data/mesh-for-data/manager/controllers/utils"
	"github.com/mesh-for-data/mesh-for-data/pkg/helm"
	pc "github.com/mesh-for-data/mesh-for-data/pkg/policy-compiler/policy-compiler"
	kapps "k8s.io/api/apps/v1"
	kbatch "k8s.io/api/batch/v1"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = kruntime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = motionv1.AddToScheme(scheme)
	_ = appv1.AddToScheme(scheme)
	_ = comv1alpha1.SchemeBuilder.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = kbatch.AddToScheme(scheme)
	_ = kapps.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

// This component starts all the controllers of the CRDs of the manager.
// This includes the following components:
// - application-controller
// - blueprint-contoller
// - movement-controller
func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var enableApplicationController bool
	var enableBlueprintController bool
	var enablePlotterController bool
	var enableMotionController bool
	var enableAllControllers bool
	var namespace string
	address := utils.ListeningAddress(8085)
	flag.StringVar(&metricsAddr, "metrics-bind-addr", address, "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&enableApplicationController, "enable-application-controller", false,
		"Enable application controller of the manager. This manages CRDs of type M4DApplication.")
	flag.BoolVar(&enableBlueprintController, "enable-blueprint-controller", false,
		"Enable blueprint controller of the manager. This manages CRDs of type Blueprint.")
	flag.BoolVar(&enablePlotterController, "enable-plotter-controller", false,
		"Enable plotter controller of the manager. This manages CRDs of type Plotter.")
	flag.BoolVar(&enableMotionController, "enable-motion-controller", false,
		"Enable motion controller of the manager. This manages CRDs of type BatchTransfer or StreamTransfer.")
	flag.BoolVar(&enableAllControllers, "enable-all-controllers", false,
		"Enables all controllers.")
	flag.StringVar(&namespace, "namespace", "", "The namespace to which this controller manager is limited.")
	flag.Parse()

	if enableAllControllers {
		enableApplicationController = true
		enableBlueprintController = true
		enablePlotterController = true
		enableMotionController = true
	}

	if !enableApplicationController && !enablePlotterController && !enableBlueprintController && !enableMotionController {
		setupLog.Info("At least one controller flag must be set!")
		os.Exit(1)
	}

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	setupLog.Info("creating manager")
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		Namespace:          namespace,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "m4d-operator-leader-election",
		Port:               9443,
	})

	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Initialize ClusterManager
	setupLog.Info("creating cluster manager")
	var clusterManager multicluster.ClusterManager
	if enableApplicationController || enablePlotterController {
		clusterManager, err = NewClusterManager(mgr)
		if err != nil {
			setupLog.Error(err, "unable to initialize cluster manager")
			os.Exit(1)
		}
	}

	if enableApplicationController {
		setupLog.Info("creating M4DApplication controller")

		// Initialize PolicyCompiler interface
		policyCompiler := pc.NewPolicyCompiler()

		// Initiate the M4DApplication Controller
		applicationController, err := app.NewM4DApplicationReconciler(mgr, "M4DApplication", policyCompiler, clusterManager, storage.NewProvisionImpl(mgr.GetClient()))
		if err != nil {
			setupLog.Error(err, "unable to create controller")
			os.Exit(1)
		}
		if err := applicationController.SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "M4DApplication")
			os.Exit(1)
		}
	}

	if enablePlotterController {
		// Initiate the Plotter Controller
		setupLog.Info("creating Plotter controller")
		plotterController := app.NewPlotterReconciler(mgr, "Plotter", clusterManager)
		if err := plotterController.SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", plotterController.Name)
			os.Exit(1)
		}
	}

	if enableBlueprintController {
		// Initiate the Blueprint Controller
		setupLog.Info("creating Blueprint controller")
		blueprintController := app.NewBlueprintReconciler(mgr, "Blueprint", new(helm.Impl))
		if err := blueprintController.SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", blueprintController.Name)
			os.Exit(1)
		}
	}

	if enableMotionController {
		setupLog.Info("creating motion controllers")
		motion.SetupMotionControllers(mgr)
	}

	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

// NewClusterManager decides based on the environment variables that are set which
// cluster manager instance should be initiated.
func NewClusterManager(mgr manager.Manager) (multicluster.ClusterManager, error) {
	setupLog := ctrl.Log.WithName("setup")
	multiClusterGroup := os.Getenv("MULTICLUSTER_GROUP")
	if user, razeeLocal := os.LookupEnv("RAZEE_USER"); razeeLocal {
		razeeURL := strings.TrimSpace(os.Getenv("RAZEE_URL"))
		password := strings.TrimSpace(os.Getenv("RAZEE_PASSWORD"))

		setupLog.Info("Using razee local at " + razeeURL)
		return razee.NewRazeeLocalManager(strings.TrimSpace(razeeURL), strings.TrimSpace(user), password, multiClusterGroup)
	} else if apiKey, satConf := os.LookupEnv("IAM_API_KEY"); satConf {
		setupLog.Info("Using IBM Satellite config")
		return razee.NewSatConfManager(strings.TrimSpace(apiKey), multiClusterGroup)
	} else if apiKey, razeeOauth := os.LookupEnv("API_KEY"); razeeOauth {
		setupLog.Info("Using Razee oauth")

		razeeURL := strings.TrimSpace(os.Getenv("RAZEE_URL"))
		return razee.NewRazeeOAuthManager(strings.TrimSpace(razeeURL), strings.TrimSpace(apiKey), multiClusterGroup)
	} else {
		setupLog.Info("Using local cluster manager")
		return local.NewManager(mgr.GetClient(), utils.GetSystemNamespace())
	}
}
