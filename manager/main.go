// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"os"
	"strconv"
	"strings"
	"time"

	"emperror.dev/errors"
	corev1 "k8s.io/api/core/v1"

	connectors "github.com/mesh-for-data/mesh-for-data/pkg/connectors/clients"
	"github.com/mesh-for-data/mesh-for-data/pkg/multicluster"
	"github.com/mesh-for-data/mesh-for-data/pkg/multicluster/local"
	"github.com/mesh-for-data/mesh-for-data/pkg/multicluster/razee"
	"github.com/mesh-for-data/mesh-for-data/pkg/storage"

	"github.com/mesh-for-data/mesh-for-data/manager/controllers/motion"

	"k8s.io/apimachinery/pkg/fields"
	kruntime "k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	appv1 "github.com/mesh-for-data/mesh-for-data/manager/apis/app/v1alpha1"
	motionv1 "github.com/mesh-for-data/mesh-for-data/manager/apis/motion/v1alpha1"
	"github.com/mesh-for-data/mesh-for-data/manager/controllers/app"
	"github.com/mesh-for-data/mesh-for-data/manager/controllers/utils"
	"github.com/mesh-for-data/mesh-for-data/pkg/helm"
	kapps "k8s.io/api/apps/v1"
	kbatch "k8s.io/api/batch/v1"
)

var (
	scheme   = kruntime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = motionv1.AddToScheme(scheme)
	_ = appv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = kbatch.AddToScheme(scheme)
	_ = kapps.AddToScheme(scheme)
}

func run(namespace string, metricsAddr string, enableLeaderElection bool,
	enableApplicationController, enableBlueprintController, enablePlotterController, enableMotionController bool) int {
	setupLog.Info("creating manager")
	systemNamespaceSelector := fields.SelectorFromSet(fields.Set{"metadata.namespace": utils.GetSystemNamespace()})
	workerNamespaceSelector := fields.SelectorFromSet(fields.Set{"metadata.namespace": "m4d-blueprints"})
	selectorsByObject := cache.SelectorsByObject{
		&appv1.Plotter{}:                {Field: systemNamespaceSelector},
		&appv1.M4DModule{}:              {Field: systemNamespaceSelector},
		&appv1.M4DStorageAccount{}:      {Field: systemNamespaceSelector},
		&appv1.Blueprint{}:              {Field: workerNamespaceSelector},
		&motionv1.BatchTransfer{}:       {Field: workerNamespaceSelector},
		&motionv1.StreamTransfer{}:      {Field: workerNamespaceSelector},
		&kbatch.Job{}:                   {Field: workerNamespaceSelector},
		&corev1.Secret{}:                {Field: workerNamespaceSelector},
		&corev1.Pod{}:                   {Field: workerNamespaceSelector},
		&corev1.PersistentVolumeClaim{}: {Field: workerNamespaceSelector},
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		Namespace:          namespace,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "m4d-operator-leader-election",
		Port:               9443,
		NewCache:           cache.BuilderWithOptions(cache.Options{SelectorsByObject: selectorsByObject}),
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		return 1
	}

	// Initialize ClusterManager
	setupLog.Info("creating cluster manager")
	var clusterManager multicluster.ClusterManager
	if enableApplicationController || enablePlotterController {
		clusterManager, err = newClusterManager(mgr)
		if err != nil {
			setupLog.Error(err, "unable to initialize cluster manager")
			return 1
		}
	}

	if enableApplicationController {
		setupLog.Info("creating M4DApplication controller")

		// Initialize PolicyManager interface
		policyManager, err := newPolicyManager()
		if err != nil {
			setupLog.Error(err, "unable to create policy manager facade", "controller", "M4DApplication")
			return 1
		}
		defer func() {
			if err := policyManager.Close(); err != nil {
				setupLog.Error(err, "unable to close policy manager facade", "controller", "M4DApplication")
			}
		}()

		// Initialize DataCatalog interface
		catalog, err := newDataCatalog()
		if err != nil {
			setupLog.Error(err, "unable to create data catalog facade", "controller", "M4DApplication")
			return 1
		}
		defer func() {
			if err := catalog.Close(); err != nil {
				setupLog.Error(err, "unable to close data catalog facade", "controller", "M4DApplication")
			}
		}()

		// Initiate the M4DApplication Controller
		applicationController := app.NewM4DApplicationReconciler(mgr, "M4DApplication", policyManager, catalog, clusterManager, storage.NewProvisionImpl(mgr.GetClient()))
		if err := applicationController.SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "M4DApplication")
			return 1
		}
		if os.Getenv("ENABLE_WEBHOOKS") != "false" {
			if err := (&appv1.M4DApplication{}).SetupWebhookWithManager(mgr); err != nil {
				setupLog.Error(err, "unable to create webhook", "webhook", "M4DApplication")
				return 1
			}
		}
	}

	if enablePlotterController {
		// Initiate the Plotter Controller
		setupLog.Info("creating Plotter controller")
		plotterController := app.NewPlotterReconciler(mgr, "Plotter", clusterManager)
		if err := plotterController.SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", plotterController.Name)
			return 1
		}
	}

	if enableBlueprintController {
		// Initiate the Blueprint Controller
		setupLog.Info("creating Blueprint controller")
		blueprintController := app.NewBlueprintReconciler(mgr, "Blueprint", new(helm.Impl))
		if err := blueprintController.SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", blueprintController.Name)
			return 1
		}
	}

	if enableMotionController {
		setupLog.Info("creating motion controllers")
		if err := motion.SetupMotionControllers(mgr); err != nil {
			setupLog.Error(err, "unable to setup motion controllers")
			return 1
		}
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		return 1
	}

	return 0
}

// Main entry point starts manager and controllers
func main() {
	var namespace string
	var metricsAddr string
	var enableLeaderElection bool
	var enableApplicationController bool
	var enableBlueprintController bool
	var enablePlotterController bool
	var enableMotionController bool
	var enableAllControllers bool
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

	os.Exit(run(namespace, metricsAddr, enableLeaderElection,
		enableApplicationController, enableBlueprintController, enablePlotterController, enableMotionController))
}

func newDataCatalog() (connectors.DataCatalog, error) {
	connectionTimeout, err := getConnectionTimeout()
	if err != nil {
		return nil, err
	}
	providerName := os.Getenv("CATALOG_PROVIDER_NAME")
	connectorURL := os.Getenv("CATALOG_CONNECTOR_URL")
	connector, err := connectors.NewGrpcDataCatalog(providerName, connectorURL, connectionTimeout)
	setupLog.Info("setting data catalog client", "Name", providerName, "URL", connectorURL, "Timeout", connectionTimeout)
	if err != nil {
		return nil, err
	}
	return connector, nil
}

func newPolicyManager() (connectors.PolicyManager, error) {
	connectionTimeout, err := getConnectionTimeout()
	if err != nil {
		return nil, err
	}

	mainPolicyManagerName := os.Getenv("MAIN_POLICY_MANAGER_NAME")
	mainPolicyManagerURL := os.Getenv("MAIN_POLICY_MANAGER_CONNECTOR_URL")
	setupLog.Info("setting main policy manager client", "Name", mainPolicyManagerName, "URL", mainPolicyManagerURL, "Timeout", connectionTimeout)
	policyManager, err := connectors.NewGrpcPolicyManager(mainPolicyManagerName, mainPolicyManagerURL, connectionTimeout)
	if err != nil {
		return nil, err
	}

	useExtensionPolicyManager, err := strconv.ParseBool(os.Getenv("USE_EXTENSIONPOLICY_MANAGER"))
	if useExtensionPolicyManager && err == nil {
		extensionPolicyManagerName := os.Getenv("EXTENSIONS_POLICY_MANAGER_NAME")
		extensionPolicyManagerURL := os.Getenv("EXTENSIONS_POLICY_MANAGER_CONNECTOR_URL")
		setupLog.Info("setting extension policy manager client", "Name", extensionPolicyManagerName, "URL", extensionPolicyManagerURL, "Timeout", connectionTimeout)
		extensionPolicyManager, err := connectors.NewGrpcPolicyManager(extensionPolicyManagerName, extensionPolicyManagerURL, connectionTimeout)
		if err != nil {
			return nil, err
		}
		policyManager = connectors.NewMultiPolicyManager(policyManager, extensionPolicyManager)
	}

	return policyManager, nil
}

// newClusterManager decides based on the environment variables that are set which
// cluster manager instance should be initiated.
func newClusterManager(mgr manager.Manager) (multicluster.ClusterManager, error) {
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

func getConnectionTimeout() (time.Duration, error) {
	connectionTimeout := os.Getenv("CONNECTION_TIMEOUT")
	timeOutInSeconds, err := strconv.Atoi(connectionTimeout)
	if err != nil {
		return 0, errors.Wrap(err, "Atoi conversion of CONNECTION_TIMEOUT failed")
	}
	return time.Duration(timeOutInSeconds) * time.Second, nil
}
