// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package argocd

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"

	app "fybrik.io/fybrik/manager/apis/app/v1beta1"

	"fybrik.io/fybrik/pkg/environment"
	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/multicluster"
	argoclient "fybrik.io/fybrik/pkg/multicluster/argocd/auto-generated/client"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
)

const (
	gitRepoBranch                     = "main"
	appBlueprintPrefix                = "blueprints-"
	blueprintsDirPath                 = "blueprints"
	tmpFileName                       = ".tmp"
	blueprintKind                     = "Blueprint"
	blueprintGroup                    = "app.fybrik.io"
	blueprintVersion                  = "v1beta1"
	gitOrigin                         = "origin"
	failToListClustersErrMsg          = "Failed to list clusters"
	failedToCreateBlueprintErrMsg     = "Failed to create blueprint"
	failedToCreateBlueprintsDirErrMsg = "Failed to create blueprints directory"
	failedToDeleteBlueprintErrMsg     = "Failed to delete blueprint"
	applicationsNamespace             = "argocd"
)

var (
	pushRepoMutex sync.Mutex
	scheme        = runtime.NewScheme()
)

func init() {
	_ = app.AddToScheme(scheme)
}

type gitRepo struct {
	password string
	username string
	url      string
}

// argocdClusterManager for argocd cluster configuration
type argocdClusterManager struct {
	client                     *argoclient.APIClient
	log                        zerolog.Logger
	argoCDAppsGitRepo          gitRepo
	argocdFybrikAppsNamePrefix string
}

// Do Git commit and push commands
func (cm *argocdClusterManager) doGitCommitAndPush(repo *git.Repository, w *git.Worktree, commitMsg string) error {
	// Commits the current staging area to the repository, with the new file
	// just created. We should provide the object.Signature of Author of the
	// commit Since version 5.0.1, we can omit the Author signature, being read
	// from the git config files.
	cm.log.Info().Msg(commitMsg)
	commit, err := w.Commit(commitMsg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "fybrik",
			Email: "fybrik@fybrik",
			When:  time.Now(),
		},
	})
	if err != nil {
		return err
	}

	remote, err := repo.Remote(gitOrigin)
	if err != nil {
		return err
	}

	cm.log.Info().Msg("commit hash " + commit.String())
	po := &git.PushOptions{
		Auth: &githttp.BasicAuth{
			Username: cm.argoCDAppsGitRepo.username,
			Password: cm.argoCDAppsGitRepo.password,
		},
		RemoteName:      gitOrigin,
		RefSpecs:        []config.RefSpec{config.RefSpec("refs/heads/*:refs/heads/*")},
		Progress:        os.Stdout,
		Force:           false,
		InsecureSkipTLS: true,
	}
	cm.log.Info().Msg("do git push " + commitMsg)
	// mutex for the writing operation
	pushRepoMutex.Lock()
	defer pushRepoMutex.Unlock()
	err = remote.Push(po)
	if err != nil {
		return err
	}
	cm.log.Info().Msg("Commit and push is done for " + commitMsg)
	return nil
}

// Create a new folder called "blueprints" in the Git repository.
// This folder includes subfolders for each of the clusters, with each subfolder serving as a container
// for the blueprints created specifically for that cluster
func (cm *argocdClusterManager) createBlueprintsDirIfNotExists() error {
	repoDir, repo, err := cm.cloneGitRepo()
	defer os.RemoveAll(repoDir)
	if err != nil {
		cm.log.Error().Err(err).Msg(failedToCreateBlueprintsDirErrMsg)
		return err
	}
	if _, err := os.Stat(repoDir + "/" + blueprintsDirPath); os.IsNotExist(err) {
		w, err := repo.Worktree()
		if err != nil {
			cm.log.Error().Err(err).Msg(failedToCreateBlueprintsDirErrMsg)
		}

		err = os.MkdirAll(repoDir+"/"+cm.getBlueprintFilePath(), os.ModePerm)
		if err != nil {
			return err
		}
		fullFilename := filepath.Join(repoDir+"/"+cm.getBlueprintFilePath(), tmpFileName)
		_, err = os.Create(fullFilename)
		if err != nil {
			return err
		}

		if err != nil {
			return err
		}
		cm.log.Info().Msg("do git add to create " + cm.getBlueprintFilePath() + " directory")
		_, err = w.Add(cm.getBlueprintFilePath() + tmpFileName)
		if err != nil {
			return err
		}

		clusters, err := cm.GetClusters()
		if err != nil {
			return err
		}
		for _, clusterItem := range clusters {
			cluster := clusterItem.Name
			cm.log.Info().Msg("creating new empty dir for cluster " + cluster)
			err = os.MkdirAll(repoDir+"/"+cm.getBlueprintFilePath()+cluster, os.ModePerm)
			if err != nil {
				return err
			}
			fullFilename := filepath.Join(repoDir+"/"+cm.getBlueprintFilePath()+cluster+"/", tmpFileName)
			_, err = os.Create(fullFilename)
			if err != nil {
				return err
			}
			cm.log.Info().Msg("do git add to create " + cm.getBlueprintFilePath() + cluster + " directory")
			_, err = w.Add(cm.getBlueprintFilePath() + cluster + "/" + tmpFileName)
			if err != nil {
				return err
			}
		}
		err = cm.doGitCommitAndPush(repo, w, "Creating Blueprints folder")
		if err != nil {
			return err
		}
		cm.log.Info().Msg("Successfully created Blueprints folder!")
	}
	return nil
}

func NewArgoCDClusterManager(connectionURL, user, password, gitRepoURL, gitRepoUser, gitRepoPassword,
	argocdFybrikAppsNamePrefix string) (multicluster.ClusterManager, error) {
	logger := logging.LogInit(logging.SETUP, "ArgoCDManager")
	tlsConfig := &tls.Config{
		// FIXME: handle TLS properly
		// #nosec G402
		InsecureSkipVerify: true,
	}
	retryClient := retryablehttp.NewClient()
	retryClient.HTTPClient.Transport = &http.Transport{TLSClientConfig: tlsConfig}

	configuration := argoclient.Configuration{
		Servers: argoclient.ServerConfigurations{
			argoclient.ServerConfiguration{
				URL:         connectionURL,
				Description: "Endpoint URL",
			},
		},
		HTTPClient: retryClient.StandardClient(),
	}

	// https://argo-cd.readthedocs.io/en/stable/developer-guide/api-docs/#authorization
	apiClient := argoclient.NewAPIClient(&configuration)

	sessionReq := argoclient.SessionSessionCreateRequest{
		Username: &user,
		Password: &password,
	}

	sess := apiClient.SessionServiceApi.SessionServiceCreate(context.Background())
	sess = sess.Body(sessionReq)
	sessionResp, httpResp, err := apiClient.SessionServiceApi.SessionServiceCreateExecute(sess)
	defer httpResp.Body.Close()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get bearer token")
		return nil, err
	}
	if httpResp.StatusCode != http.StatusOK {
		logger.Error().Msg("Failed get bearer token: http status code is " + strconv.Itoa(httpResp.StatusCode))
		return nil, errors.New("http status code is " + strconv.Itoa(httpResp.StatusCode))
	}

	token := fmt.Sprintf("Bearer %s", *sessionResp.Token)
	configuration.DefaultHeader = map[string]string{"Authorization": token}

	logger.Info().Msg("Bearer token successfully fetched")

	logger.Info().Msg("Initializing ArgoCD cluster manager")

	cm := argocdClusterManager{
		client: argoclient.NewAPIClient(&configuration),
		log:    logger,
		argoCDAppsGitRepo: gitRepo{
			password: gitRepoPassword,
			username: gitRepoUser,
			url:      gitRepoURL,
		},
		argocdFybrikAppsNamePrefix: argocdFybrikAppsNamePrefix,
	}
	err = cm.createBlueprintsDirIfNotExists()
	if err != nil {
		return nil, err
	}

	return &cm, nil
}

func (cm *argocdClusterManager) packClusterConfigMap(params map[string]string) *v1.ConfigMap {
	configMap := v1.ConfigMap{}
	configMap.Data = params
	return &configMap
}

// Get the cluster info.
// The cluster information is retrieved from the Argo CD Application for Fybrik deployment on the cluster.
// To do so, an API call to ApplicationService is done in order to fetch the helm parameters
// of the deployment in the cluster.
// For example, to get the cluster zone the helm parameter cluster.zone
// is retrieved from the Argo CD Application:
//
//	helm:
//	  parameters:
//	    - name: cluster.name
//	      value: kind-control
//	    - name: cluster.zone
//	      value: baggin
//
// TODO: Consider retrieving the cluster info from fybrik cluster-metadata configMap resource
// by using ApplicationServiceGetResource API call
func (cm *argocdClusterManager) getClusterInfo(clusterName string) (multicluster.Cluster, error) {
	var cluster multicluster.Cluster
	req := cm.client.ApplicationServiceApi.ApplicationServiceGet(context.Background(),
		cm.argocdFybrikAppsNamePrefix+"-"+clusterName)
	cm.log.Info().Msg("application name: " + cm.argocdFybrikAppsNamePrefix + "-" + clusterName)
	argocdApplication, httpResp, err := cm.client.ApplicationServiceApi.ApplicationServiceGetExecute(req)
	defer httpResp.Body.Close()
	if err != nil {
		cm.log.Error().Err(err).Msg("Failed to get argocd application")
		return cluster, err
	}
	if httpResp.StatusCode != http.StatusOK {
		cm.log.Error().Msg("Failed to get argocd application: http status code is " + strconv.Itoa(httpResp.StatusCode))
		return cluster, errors.New("http status code is " + strconv.Itoa(httpResp.StatusCode))
	}
	fybrikHelmParams := argocdApplication.GetSpec().Source.Helm.GetParameters()
	var params = make(map[string]string)

	for _, helmParam := range fybrikHelmParams {
		switch helmParam.GetName() {
		case "cluster.region":
			params["Region"] = helmParam.GetValue()
			cm.log.Info().Msg("region: " + helmParam.GetValue())
		case "cluster.zone":
			params["Zone"] = helmParam.GetValue()
			cm.log.Info().Msg("zone: " + helmParam.GetValue())
		case "cluster.name":
			params["ClusterName"] = helmParam.GetValue()
			cm.log.Info().Msg("ClusterName: " + helmParam.GetValue())
		case "cluster.vaultAuthPath":
			params["VaultAuthPath"] = helmParam.GetValue()
			cm.log.Info().Msg("VaultAuthPath: " + helmParam.GetValue())
		}
	}

	if len(params) != 4 {
		cm.log.Error().Err(err).Msg("missing expected cluster info related field in helm params")
		return cluster, errors.New("Failed to get cluster info")
	}

	return multicluster.CreateCluster(*cm.packClusterConfigMap(params)), nil
}

// Clone a git repository into local filesystem
func (cm *argocdClusterManager) cloneGitRepo() (string, *git.Repository, error) {
	cm.log.Info().Msg(cm.argoCDAppsGitRepo.username + cm.argoCDAppsGitRepo.url)
	tmpDir, err := os.MkdirTemp(environment.GetDataDir(), "blueprints-repo")
	if err != nil {
		return "", nil, err
	}
	r, err := git.PlainClone(tmpDir, false, &git.CloneOptions{
		Auth: &githttp.BasicAuth{
			Username: cm.argoCDAppsGitRepo.username,
			Password: cm.argoCDAppsGitRepo.password,
		},
		URL: cm.argoCDAppsGitRepo.url,
	})
	if err != nil {
		cm.log.Error().Err(err).Msg("Failed to clone repo")
		return "", nil, err
	}
	return tmpDir, r, nil
}

// GetClusters returns a list of registered clusters
// To do so an API call to ClusterServiceList is used.
func (cm *argocdClusterManager) GetClusters() ([]multicluster.Cluster, error) {
	cm.log.Info().Msg("list clusters")
	var clusters []multicluster.Cluster
	req := cm.client.ClusterServiceApi.ClusterServiceList(context.Background())

	clustersList, httpResp, err := cm.client.ClusterServiceApi.ClusterServiceListExecute(req)
	defer httpResp.Body.Close()
	if err != nil {
		cm.log.Error().Err(err).Msg(failToListClustersErrMsg)
		return nil, err
	}
	if httpResp.StatusCode != http.StatusOK {
		cm.log.Error().Msg(failToListClustersErrMsg + " http status code is " + strconv.Itoa(httpResp.StatusCode))
		return nil, errors.New("http status code is " + strconv.Itoa(httpResp.StatusCode))
	}
	if !clustersList.HasItems() {
		cm.log.Error().Msg(failToListClustersErrMsg + " no cluster exists")
		return nil, errors.New("no cluster exists")
	}
	for _, clusterItem := range clustersList.GetItems() {
		name := clusterItem.GetName()
		cm.log.Info().Msg("cluster name: " + name)
		cluster, err := cm.getClusterInfo(name)
		if err != nil {
			cm.log.Error().Err(err).Msg(failToListClustersErrMsg)
			return nil, err
		}
		clusters = append(clusters, cluster)
	}

	return clusters, nil
}

func (cm *argocdClusterManager) IsMultiClusterSetup() bool {
	return true
}

func (cm *argocdClusterManager) getBlueprintFileName(cluster, blueprintName, blueprintNamespace string) string {
	return cluster + "-" + blueprintName + "-" + blueprintNamespace + ".yaml"
}
func (cm *argocdClusterManager) getBlueprintFilePath() string {
	return blueprintsDirPath + "/"
}

// GetBlueprint returns a blueprint matching the given name, namespace and cluster details
// To do so an API call to ApplicationServiceGetResource is done which returns the actual Blueprint
// resource deployed on the cluster.
func (cm *argocdClusterManager) GetBlueprint(cluster, namespace, name string) (*app.Blueprint, error) {
	cm.log.Info().Msg("Get Blueprint " + " cluster " + cluster + " namespace: " + namespace + " name: " + name)

	req1 := cm.client.ApplicationServiceApi.ApplicationServiceGetResource(context.Background(), appBlueprintPrefix+cluster)
	req1 = req1.ResourceName(name)
	req1 = req1.Kind(blueprintKind)
	req1 = req1.AppNamespace(applicationsNamespace)
	req1 = req1.Group(blueprintGroup)
	req1 = req1.Version(blueprintVersion)
	req1 = req1.Namespace(namespace)
	resp1, httpResp, err := cm.client.ApplicationServiceApi.ApplicationServiceGetResourceExecute(req1)
	defer httpResp.Body.Close()
	if err != nil {
		cm.log.Error().Err(err).Msg("Failed to get application manifest")
		return nil, err
	}
	if httpResp.StatusCode != http.StatusOK {
		cm.log.Error().Msg("Failed to get application manifest: http status code is " + strconv.Itoa(httpResp.StatusCode))
		return nil, errors.New("http status code is " + strconv.Itoa(httpResp.StatusCode))
	}
	cm.log.Info().Msg("print manifest")

	manifest := resp1.GetManifest()
	blueprint := app.Blueprint{}
	cm.log.Info().Msg(manifest)
	err = multicluster.Decode(manifest, scheme, &blueprint)
	if err != nil {
		return nil, err
	}
	if blueprint.Namespace == "" {
		cm.log.Warn().Msg("Retrieved an empty blueprint")
		return nil, nil
	}
	cm.log.Info().Msg("found manifest for " + blueprint.GetName())
	cm.log.Info().Msg("blueprint successfully read " + blueprint.Namespace)
	return &blueprint, nil
}

// CreateBlueprint creates a blueprint resource or updates an existing one
// It does so by pushing the blueprint to the Git repository.
func (cm *argocdClusterManager) CreateBlueprint(cluster string, blueprint *app.Blueprint) error {
	cm.log.Info().Msg("Create Blueprint " + " cluster " + cluster + " namespace: " + blueprint.Namespace + " name: " + blueprint.Name)
	repoDir, repo, err := cm.cloneGitRepo()
	defer os.RemoveAll(repoDir)
	if err != nil {
		cm.log.Error().Err(err).Msg(failedToCreateBlueprintErrMsg)
		return err
	}
	fileName := cm.getBlueprintFileName(cluster, blueprint.Name, blueprint.Namespace)
	w, err := repo.Worktree()
	if err != nil {
		cm.log.Error().Err(err).Msg(failedToCreateBlueprintErrMsg)
	}

	content, err := yaml.Marshal(blueprint)
	if err != nil {
		return err
	}
	cm.log.Info().Msg("fullPath: " + fileName)

	fullFilename := filepath.Join(repoDir+"/"+cm.getBlueprintFilePath()+cluster, fileName)
	err = os.WriteFile(fullFilename, content, 0644)

	if err != nil {
		return err
	}
	cm.log.Info().Msg("do git add of blueprint " + cm.getBlueprintFilePath() + cluster + "/" + fileName)
	_, err = w.Add(cm.getBlueprintFilePath() + cluster + "/" + fileName)
	if err != nil {
		return err
	}

	err = cm.doGitCommitAndPush(repo, w, "Create Blueprint")
	if err != nil {
		return err
	}
	cm.log.Info().Msg("Successfully created blueprint!")
	return nil
}

// UpdateBlueprint updates the given blueprint or creates a new one if it does not exist
func (cm *argocdClusterManager) UpdateBlueprint(cluster string, blueprint *app.Blueprint) error {
	cm.log.Info().Msg("Update Blueprint " + " cluster " + cluster + "namespace: " + blueprint.Namespace + " name: " + blueprint.Name)
	return cm.CreateBlueprint(cluster, blueprint)
}

// DeleteBlueprint deletes the blueprint resource
// It does so by removing the blueprint from the Git repository.
func (cm *argocdClusterManager) DeleteBlueprint(cluster, namespace, name string) error {
	repoDir, repo, err := cm.cloneGitRepo()
	defer os.RemoveAll(repoDir)
	if err != nil {
		cm.log.Error().Err(err).Msg(failedToDeleteBlueprintErrMsg)
		return err
	}
	fileName := cm.getBlueprintFileName(cluster, name, namespace)

	cm.log.Info().Msg("fullPath: " + fileName)

	cm.log.Info().Msg("do git remove of blueprint " + cm.getBlueprintFilePath() + cluster + "/" + fileName)
	w, err := repo.Worktree()
	if err != nil {
		cm.log.Error().Err(err).Msg(failedToDeleteBlueprintErrMsg)
	}

	_, err = w.Remove(cm.getBlueprintFilePath() + cluster + "/" + fileName)
	if err != nil {
		return err
	}

	err = cm.doGitCommitAndPush(repo, w, "Delete Blueprint")
	if err != nil {
		return err
	}
	cm.log.Info().Msg("Successfully deleted blueprint!")
	return nil
}
