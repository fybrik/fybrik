// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"os"
	"strings"

	"github.com/ibm/the-mesh-for-data/pkg/multicluster"
	"github.com/ibm/the-mesh-for-data/pkg/multicluster/local"
	"github.com/ibm/the-mesh-for-data/pkg/multicluster/razee"
	"github.com/ibm/the-mesh-for-data/pkg/storage"

	"github.com/ibm/the-mesh-for-data/manager/controllers/motion"

	"github.com/hashicorp/vault/api"
	kruntime "k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	comv1alpha1 "github.com/IBM/dataset-lifecycle-framework/src/dataset-operator/pkg/apis/com/v1alpha1"
	appv1 "github.com/ibm/the-mesh-for-data/manager/apis/app/v1alpha1"
	motionv1 "github.com/ibm/the-mesh-for-data/manager/apis/motion/v1alpha1"
	"github.com/ibm/the-mesh-for-data/manager/controllers/app"
	"github.com/ibm/the-mesh-for-data/manager/controllers/utils"
	"github.com/ibm/the-mesh-for-data/pkg/helm"
	pc "github.com/ibm/the-mesh-for-data/pkg/policy-compiler/policy-compiler"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = kruntime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = motionv1.AddToScheme(scheme)
	_ = appv1.AddToScheme(scheme)
	_ = comv1alpha1.SchemeBuilder.AddToScheme(scheme)
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
	flag.StringVar(&metricsAddr, "metrics-addr", address, "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
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

	if !enableAllControllers && !enableApplicationController && !enableBlueprintController && !enableMotionController {
		setupLog.Info("At least one controller flag must be set!")
		os.Exit(1)
	}

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

	// Initialize ClusterManager
	clusterManager, err := NewClusterManager(mgr)

	if err != nil {
		setupLog.Error(err, "unable to initialize cluster manager")
		os.Exit(1)
	}

	if enableApplicationController || enableAllControllers {
		// Initiate vault client
		vaultClient, errVaultSetup := initVaultConnection()
		if errVaultSetup != nil {
			setupLog.Error(errVaultSetup, "Error setting up vault")
			os.Exit(1)
		}

		// Initialize PolicyCompiler interface
		policyCompiler := pc.NewPolicyCompiler()

		// Initiate the M4DApplication Controller
		applicationController := app.NewM4DApplicationReconciler(mgr, "M4DApplication", vaultClient, policyCompiler, clusterManager, storage.NewProvisionImpl(mgr.GetClient()))
		if err := applicationController.SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "M4DApplication")
			os.Exit(1)
		}
	}

	if enableBlueprintController || enableAllControllers {
		// Initiate the Blueprint Controller
		blueprintController := app.NewBlueprintReconciler(mgr, "Blueprint", new(helm.Impl))
		if err := blueprintController.SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", blueprintController.Name)
			os.Exit(1)
		}
	}

	if enablePlotterController || enableAllControllers {
		app.SetupPlotterController(mgr, clusterManager)
	}

	if enableMotionController || enableAllControllers {
		motion.SetupMotionControllers(mgr)
	}

	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

// init vault client and mount the base directory for storing credentials
func initVaultConnection() (*api.Client, error) {
	token := utils.GetVaultToken()
	if err := utils.MountDatasetVault(token); err != nil {
		return nil, err
	}
	vaultClient, err := utils.InitVault(token)
	if err != nil {
		return nil, err
	}

	// Create and save a vault policy
	path := utils.GetVaultDatasetHome() + "*"
	policy := "path \"" + path + "\"" + " {\n	capabilities = [\"read\"]\n }"
	policyName := "read-dataset-creds"

	setupLog.Info("policyName: " + policyName + "  policy: " + policy)
	if err = utils.WriteVaultPolicy(policyName, policy, vaultClient); err != nil {
		setupLog.Info("      Failed writing policy: " + err.Error())
		return vaultClient, err
	}

	setupLog.Info("Assigning the policy to " + "/role/" + utils.GetSecretProviderRole())
	// Link the policy to the authentication role (configured)
	if err = utils.LinkVaultPolicyToIdentity("/role/"+utils.GetSecretProviderRole(), policyName, vaultClient); err != nil {
		setupLog.Info("Could not create a role " + utils.GetSecretProviderRole() + " : " + err.Error())
		return vaultClient, err
	}
	return vaultClient, nil
}

// This method decides based on the environment variables that are set which
// cluster manager instance should be initiated.
func NewClusterManager(mgr manager.Manager) (multicluster.ClusterManager, error) {
	setupLog := ctrl.Log.WithName("setup")
	if user, razeeLocal := os.LookupEnv("RAZEE_USER"); razeeLocal {
		razeeURL := strings.TrimSpace(os.Getenv("RAZEE_URL"))
		password := strings.TrimSpace(os.Getenv("RAZEE_PASSWORD"))

		setupLog.Info("Using razee local at " + razeeURL)
		return razee.NewRazeeManager(strings.TrimSpace(razeeURL), strings.TrimSpace(user), password)
	} else if apiKey, satConf := os.LookupEnv("IAM_API_KEY"); satConf {
		setupLog.Info("Using IBM Satellite config")
		return razee.NewSatConfManager(strings.TrimSpace(apiKey))
	} else {
		setupLog.Info("Using local cluster manager")
		return local.NewManager(mgr.GetClient(), utils.GetSystemNamespace())
	}
}
