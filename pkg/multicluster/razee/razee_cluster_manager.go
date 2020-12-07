package razee

import (
	"fmt"
	"github.com/IBM/satcon-client-go/client"
	"github.com/IBM/satcon-client-go/client/types"
	"github.com/ghodss/yaml"
	"github.com/go-logr/logr"
	"github.com/ibm/the-mesh-for-data/manager/apis/app/v1alpha1"
	"github.com/ibm/the-mesh-for-data/pkg/multicluster"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	clusterMetadataConfigMapSL string = "/api/v1/namespaces/m4d-system/configmaps/cluster-metadata"
)

var (
	scheme = runtime.NewScheme()
)

//nolint:golint,unused
func init() {
	_ = v1alpha1.AddToScheme(scheme)
}

//nolint:golint,unused
type ClusterManager struct {
	orgId 		string
	con   		client.SatCon
	razeeClient *RazeeClient
	log   		logr.Logger
}

//nolint:golint,unused
func (r *ClusterManager) GetClusters() ([]multicluster.Cluster, error) {
	var clusters []multicluster.Cluster
	razeeClusters, err := r.con.Clusters.ClustersByOrgID(r.orgId)
	if err != nil {
		return nil, err
	}
	for _, c := range razeeClusters {
		configMapJson, err := r.razeeClient.getResourceByKeys(r.orgId, c.ClusterID, clusterMetadataConfigMapSL)
		if err != nil {
			return nil, err
		}
		scheme := runtime.NewScheme()
		clusterMetadataConfigmap := v1.ConfigMap{}
		err = multicluster.Decode(configMapJson, scheme, &clusterMetadataConfigmap)
		if err != nil {
			return nil, err
		}
		cluster := multicluster.Cluster{
			Name: clusterMetadataConfigmap.Data["ClusterName"],
			Metadata: multicluster.ClusterMetadata{
				Region: clusterMetadataConfigmap.Data["Region"],
				Zone:   clusterMetadataConfigmap.Data["Zone"],
			},
		}
		clusters = append(clusters, cluster)
	}
	return clusters, nil
}

//nolint:golint,unused
func createBluePrintSelfLink(namespace string, name string) string {
	return fmt.Sprintf("/apis/app.m4d.ibm.com/v1alpha1/namespaces/%s/blueprints/%s", namespace, name)
}

//nolint:golint,unused
func (r *ClusterManager) GetBlueprint(clusterName string, namespace string, name string) (*v1alpha1.Blueprint, error) {
	selfLink := createBluePrintSelfLink(namespace, name)
	cluster, err := r.razeeClient.getClusterByName(r.orgId, clusterName)
	if err != nil {
		return nil, err
	}
	jsonData, err := r.razeeClient.getResourceByKeys(r.orgId, cluster.ClusterId, selfLink)
	if err != nil {
		return nil, err
	}

	_ = v1alpha1.AddToScheme(scheme)
	blueprint := v1alpha1.Blueprint{}
	err = multicluster.Decode(jsonData, scheme, &blueprint)
	return &blueprint, err
}

//nolint:golint,unused
func getGroupName(cluster string) string {
	return "m4d-" + cluster
}

//nolint:golint,unused
func (r *ClusterManager) CreateBlueprint(cluster string, blueprint *v1alpha1.Blueprint) error {
	groupName := getGroupName(cluster)
	channelName := channelName(cluster, blueprint.Namespace, blueprint.Name)
	version := "0"
	content, err := yaml.Marshal(blueprint)
	if err != nil {
		return err
	}
	razeeClusters, err := r.con.Clusters.ClustersByOrgID(r.orgId)
	if err != nil {
		return err
	}
	var rCluster types.Cluster
	if len(razeeClusters) == 0 {
		err = fmt.Errorf("No clusters found for orgID %v", r.orgId)
		return err
	}
	for _, c := range razeeClusters {
		// Hack until sat-con library is extended with name field
		m := c.Metadata.(map[string]interface{})
		name := fmt.Sprintf("%v", m["name"])
		if name == cluster {
			rCluster = c
		}
	}
	if rCluster.ClusterID == "" {
		err = fmt.Errorf("Cannot find cluster %v", cluster)
		return err
	}

	// check group exists
	groups, err := r.con.Groups.Groups(r.orgId)
	if err != nil {
		return err
	}
	var group *types.Group
	var groupUuid string
	for _, g := range groups {
		if g.Name == groupName {
			group = &g
			groupUuid = g.UUID
		}
	}
	if group == nil {
		addGroup, err := r.con.Groups.AddGroup(r.orgId, groupName)
		if err != nil {
			return err
		}
		groupUuid = addGroup.UUID
	}

	_, err = r.con.Groups.GroupClusters(r.orgId, groupUuid, []string{rCluster.ClusterID})
	if err != nil {
		return err
	}

	// create channel
	channel, err := r.con.Channels.AddChannel(r.orgId, channelName)
	if err != nil {
		return err
	}

	// create channel version
	channelVersion, err := r.con.Versions.AddChannelVersion(r.orgId, channel.UUID, version, []byte(content), "")
	if err != nil {
		return err
	}

	// create subscription
	_, err = r.con.Subscriptions.AddSubscription(r.orgId, channelName, channel.UUID, channelVersion.VersionUUID, []string{groupName})
	if err != nil {
		return err
	}

	r.log.Info("Successfully created subscription! ")
	return nil
}

//nolint:golint,unused
func (r *ClusterManager) UpdateBlueprint(cluster string, blueprint *v1alpha1.Blueprint) error {
	return nil
}

//nolint:golint,unused
func (r *ClusterManager) DeleteBlueprint(cluster string, namespace string, name string) error {
	channelName := channelName(cluster, namespace, name)
	channel, err := r.con.Channels.ChannelByName(r.orgId, channelName)
	if err != nil {
		return err
	}
	for _, s := range channel.Subscriptions {
		subscription, err := r.con.Subscriptions.RemoveSubscription(r.orgId, s.UUID)
		if err != nil {
			return err
		}
		if subscription.Success {
			r.log.Info("Successfully deleted subscription " + subscription.UUID)
		}
	}
	for _, v := range channel.Versions {
		version, err := r.con.Versions.RemoveChannelVersion(r.orgId, v.UUID)
		if err != nil {
			return err
		}
		if version.Success {
			r.log.Info("Successfully deleted version " + version.UUID)
		}
	}

	removeChannel, err := r.con.Channels.RemoveChannel(r.orgId, channel.UUID)
	if err != nil {
		return err
	}
	if removeChannel.Success {
		r.log.Info("Successfully deleted channel " + removeChannel.UUID)
	}
	return nil
}

//nolint:golint,unused
func channelName(cluster string, namespace string, name string) string {
	return "m4d.ibm.com" + "/" + cluster + "/"+  namespace + "/" + name
}

//nolint:golint,unused
func NewRazeeManager(url string, login string, password string, orgId string) multicluster.ClusterManager {
	localAuth := &RazeeLocalRoundTripper{
		url:      url,
		login:    login,
		password: password,
		token:    "",
	}

	con, _ := client.New(url, http.DefaultClient, localAuth)
	razeeClient := NewRazeeLocalClient(url, login, password)
	logger := ctrl.Log.WithName("ClusterManager")

	return &ClusterManager{
		orgId: orgId,
		con:   con,
		razeeClient: razeeClient,
		log:   logger,

	}
}
