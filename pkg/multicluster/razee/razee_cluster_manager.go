// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

//nolint:govet,revive
package razee

import (
	"fmt"
	"strconv"
	"strings"

	"emperror.dev/errors"
	"github.com/IBM/satcon-client-go/client"
	"github.com/IBM/satcon-client-go/client/auth/apikey"
	"github.com/IBM/satcon-client-go/client/auth/iam"
	"github.com/IBM/satcon-client-go/client/auth/local"
	"github.com/IBM/satcon-client-go/client/types"
	"github.com/ghodss/yaml"
	"github.com/rs/zerolog"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	app "fybrik.io/fybrik/manager/apis/app/v1beta1"
	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/multicluster"
)

const (
	clusterMetadataConfigMapSL = "/api/v1/namespaces/fybrik-system/configmaps/cluster-metadata"
	endPointURL                = "https://config.satellite.cloud.ibm.com/graphql"
	bluePrintSelfLink          = "/apis/app.fybrik.io/v1beta1/namespaces/%s/blueprints/%s"
	channelNameTemplate        = "fybrik.io-%s-%s"
	groupNameTemplate          = "fybrik-%s"
	clusterGroupKey            = "clusterGroup"
	clusterKey                 = "cluster"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = app.AddToScheme(scheme)
}

type razeeClusterManager struct {
	orgID        string
	clusterGroup string
	con          client.SatCon
	log          zerolog.Logger
}

func (r *razeeClusterManager) IsMultiClusterSetup() bool {
	return true
}

func (r *razeeClusterManager) GetClusters() ([]multicluster.Cluster, error) {
	var clusters []multicluster.Cluster
	var razeeClusters []types.Cluster
	var err error
	if r.clusterGroup != "" {
		var group *types.Group
		r.log.Info().Str(clusterGroupKey, r.clusterGroup).Msg("Using clusterGroup to fetch cluster info")
		group, err = r.con.Groups.GroupByName(r.orgID, r.clusterGroup)
		if err != nil {
			return nil, err
		}
		razeeClusters = group.Clusters
	} else {
		r.log.Info().Msg("Using all clusters in organization as reference clusters.")
		razeeClusters, err = r.con.Clusters.ClustersByOrgID(r.orgID)
		if err != nil {
			return nil, err
		}
	}

	for _, c := range razeeClusters {
		resourceContent, err := r.con.Resources.ResourceContent(r.orgID, c.ClusterID, clusterMetadataConfigMapSL)
		if err != nil {
			r.log.Error().Err(err).Str(clusterKey, c.Name).Msg("Could not fetch cluster information")
			return nil, err
		}
		// If no content for the resource was found the cluster is not part of Fybrik or not installed
		// correctly. Fybrik should ignore those clusters and continue.
		if resourceContent == nil {
			r.log.Error().Err(err).Str(clusterKey, c.Name).Msg("Resource content returned is nil! Skipping cluster")
			continue
		}
		cmcm := corev1.ConfigMap{}
		err = multicluster.Decode(resourceContent.Content, scheme, &cmcm)
		if err != nil {
			return nil, err
		}
		cluster := multicluster.CreateCluster(cmcm)
		clusters = append(clusters, cluster)
	}
	return clusters, nil
}

func createBluePrintSelfLink(namespace, name string) string {
	return fmt.Sprintf(bluePrintSelfLink, namespace, name)
}

func (r *razeeClusterManager) GetBlueprint(clusterName, namespace, name string) (*app.Blueprint, error) {
	selfLink := createBluePrintSelfLink(namespace, name)
	cluster, err := r.con.Clusters.ClusterByName(r.orgID, clusterName)
	if err != nil {
		return nil, err
	}
	jsonData, err := r.con.Resources.ResourceContent(r.orgID, cluster.ClusterID, selfLink)
	log := r.log.With().Str(clusterKey, clusterName).Str(logging.NAME, name).Str(logging.NAMESPACE, namespace).Logger()
	if err != nil {
		log.Error().Err(err).Msg("Error while fetching resource content of blueprint")
		return nil, err
	}
	if jsonData == nil {
		log.Warn().Msg("Could not get any resource data")
		return nil, nil
	}
	log.Debug().Str("Blueprint data", jsonData.Content)

	if jsonData.Content == "" {
		log.Warn().Msg("Retrieved empty data")
		return nil, nil
	}

	blueprint := app.Blueprint{}
	err = multicluster.Decode(jsonData.Content, scheme, &blueprint)
	if blueprint.Namespace == "" {
		log.Warn().Msg("Retrieved an empty blueprint")
		return nil, nil
	}
	return &blueprint, err
}

func getGroupName(cluster string) string {
	return fmt.Sprintf(groupNameTemplate, cluster)
}

type Collection struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []metav1.Object `json:"items" protobuf:"bytes,2,rep,name=items"`
}

//nolint:funlen,gocyclo
func (r *razeeClusterManager) CreateBlueprint(cluster string, blueprint *app.Blueprint) error {
	groupName := getGroupName(cluster)
	channelName := channelName(cluster, blueprint.Name)
	version := "0"
	logging.LogStructure("Blueprint to create", blueprint, &r.log, zerolog.DebugLevel, false, false)
	content, err := yaml.Marshal(blueprint)
	if err != nil {
		return err
	}

	rCluster, err := r.con.Clusters.ClusterByName(r.orgID, cluster)
	if err != nil {
		return errors.Wrap(err, "error while fetching cluster by name")
	}
	if rCluster == nil {
		return fmt.Errorf("no cluster found for orgID %v and cluster name %v", r.orgID, cluster)
	}

	// check group exists
	group, err := r.con.Groups.GroupByName(r.orgID, groupName)
	if err != nil {
		if err.Error() == "Cannot destructure property 'req_id' of 'context' as it is undefined." {
			r.log.Info().Msg("Group does not exist. Creating group.")
		} else {
			r.log.Error().Err(err).Str("group", groupName).Msg("Error while fetching group by name")
			return err
		}
	}
	var groupUUID string
	if group == nil {
		addGroup, err := r.con.Groups.AddGroup(r.orgID, groupName)
		if err != nil {
			return err
		}
		groupUUID = addGroup.UUID
	} else {
		groupUUID = group.UUID
	}

	_, err = r.con.Groups.GroupClusters(r.orgID, groupUUID, []string{rCluster.ClusterID})
	//nolint:revive
	if err != nil {
		r.log.Error().Err(err).Str("group", groupName).
			Str(clusterKey, rCluster.Name).Str("groupUUID", groupUUID).Msg("Error while creating group")
		return err
	}

	// Check if channel exists
	existingChannel, err := r.con.Channels.ChannelByName(r.orgID, channelName)
	if err != nil {
		if !strings.HasPrefix(err.Error(), "Query channelByName error.") {
			return err
		}
	}
	if existingChannel != nil {
		// Channel already exists. Update channel instead of creating
		r.log.Info().Str("existingChannel", existingChannel.Name).Msg("Channel already exists! Updating channel version...")
		return r.UpdateBlueprint(cluster, blueprint)
	}

	// create channel
	channel, err := r.con.Channels.AddChannel(r.orgID, channelName)
	if err != nil {
		return err
	}

	// create channel version
	channelVersion, err := r.con.Versions.AddChannelVersion(r.orgID, channel.UUID, version, content, "")
	if err != nil {
		// Remove channel if channelVersion could not be created
		removeChannel, channelRemoveErr := r.con.Channels.RemoveChannel(r.orgID, channel.UUID)
		if channelRemoveErr != nil {
			r.log.Error().Err(channelRemoveErr).Msg("Unable to remove channel after error")
		} else if removeChannel.Success {
			r.log.Info().Msg("Rolled back channel version after error")
		}

		return err
	}

	// create subscription
	_, err = r.con.Subscriptions.AddSubscription(r.orgID, channelName, channel.UUID, channelVersion.VersionUUID, []string{groupName})
	if err != nil {
		// Remove channelVersion and channel if the subscription could not be created
		removeChannelVersion, versionRemoveErr := r.con.Versions.RemoveChannelVersion(r.orgID, channelVersion.VersionUUID)
		if versionRemoveErr != nil {
			r.log.Error().Err(versionRemoveErr).Msg("Unable to remove channel version after error")
		} else if removeChannelVersion.Success {
			r.log.Info().Msg("Rolled back channel version after error")
		}
		removeChannel, channelRemoveErr := r.con.Channels.RemoveChannel(r.orgID, channel.UUID)
		if channelRemoveErr != nil {
			r.log.Error().Err(channelRemoveErr).Msg("Unable to remove channel after error")
		} else if removeChannel.Success {
			r.log.Info().Msg("Rolled back channel after error")
		}
		return err
	}

	r.log.Info().Msg("Successfully created subscription!")
	return nil
}

func (r *razeeClusterManager) UpdateBlueprint(cluster string, blueprint *app.Blueprint) error {
	channelName := channelName(cluster, blueprint.Name)

	content, err := yaml.Marshal(blueprint)
	if err != nil {
		return err
	}
	logging.LogStructure("Blueprint to update", blueprint, &r.log, zerolog.DebugLevel, false, false)

	max := 0
	channelInfo, err := r.con.Channels.ChannelByName(r.orgID, channelName)
	if err != nil {
		return fmt.Errorf("cannot fetch channel info for channel '%s'", channelName)
	}
	for _, version := range channelInfo.Versions {
		var v int
		v, err = strconv.Atoi(version.Name)
		if err != nil {
			return fmt.Errorf("cannot parse version name %s", version.Name)
		} else if max < v {
			max = v
		}
	}

	nextVersion := strconv.Itoa(max + 1)

	// There is only one subscription per channel in our use case
	if len(channelInfo.Subscriptions) != 1 {
		return errors.New("found more or less than one subscription")
	}
	subscriptionUUID := channelInfo.Subscriptions[0].UUID

	r.log.Trace().Str("nextVersion", nextVersion).Str("subscriptionUUID", subscriptionUUID).
		Str("channelUUID", channelInfo.UUID).Msg("Creating new channel version")

	// create channel version
	channelVersion, err := r.con.Versions.AddChannelVersion(r.orgID, channelInfo.UUID, nextVersion, content, "")
	if err != nil {
		r.log.Error().Err(err)
		return err
	}

	r.log.Trace().Msg("Updating subscription...")

	// update subscription
	_, err = r.con.Subscriptions.SetSubscription(r.orgID, subscriptionUUID, channelVersion.VersionUUID)
	if err != nil {
		return err
	}

	r.log.Info().Msg("Subscription successfully updated!")

	return nil
}

func (r *razeeClusterManager) DeleteBlueprint(cluster, namespace, name string) error {
	channelName := channelName(cluster, name)
	channel, err := r.con.Channels.ChannelByName(r.orgID, channelName)
	if err != nil {
		return err
	}
	for _, s := range channel.Subscriptions {
		subscription, err := r.con.Subscriptions.RemoveSubscription(r.orgID, s.UUID)
		if err != nil {
			return err
		}
		if subscription.Success {
			r.log.Info().Msg("Successfully deleted subscription " + subscription.UUID)
		}
	}
	for _, v := range channel.Versions {
		version, err := r.con.Versions.RemoveChannelVersion(r.orgID, v.UUID)
		if err != nil {
			return err
		}
		if version.Success {
			r.log.Info().Msg("Successfully deleted version " + version.UUID)
		}
	}

	removeChannel, err := r.con.Channels.RemoveChannel(r.orgID, channel.UUID)
	if err != nil {
		return err
	}
	if removeChannel.Success {
		r.log.Info().Msg("Successfully deleted channel " + removeChannel.UUID)
	}
	return nil
}

// The channel name should be per cluster and plotter, so it cannot be based on
// the namespace that is random for every blueprint
func channelName(cluster, name string) string {
	return fmt.Sprintf(channelNameTemplate, cluster, name)
}

// NewRazeeLocalClusterManager creates an instance of Razee based ClusterManager with userName/password authentication
func NewRazeeLocalClusterManager(url, login, password, clusterGroup string) (multicluster.ClusterManager, error) {
	localAuth, err := local.NewClient(url, login, password)
	if err != nil {
		return nil, err
	}
	con, _ := client.New(url, localAuth)
	logger := logging.LogInit(logging.CONTROLLER, "RazeeManager")
	me, err := con.Users.Me()
	if err != nil {
		return nil, err
	}

	if me == nil {
		return nil, errors.New("could not retrieve login information of Razee")
	}

	logger.Info().Str(clusterGroupKey, clusterGroup).Str("orgID", me.OrgId).Msg("Initializing Razee local")

	return &razeeClusterManager{
		orgID:        me.OrgId,
		clusterGroup: clusterGroup,
		con:          con,
		log:          logger,
	}, nil
}

// NewRazeeOAuthClusterManager creates an instance of Razee based ClusterManager with OAuth authentication
func NewRazeeOAuthClusterManager(url, apiKey, clusterGroup string) (multicluster.ClusterManager, error) {
	auth, err := apikey.NewClient(apiKey)
	if err != nil {
		return nil, err
	}
	con, _ := client.New(url, auth)
	logger := logging.LogInit(logging.CONTROLLER, "RazeeManager")
	me, err := con.Users.Me()
	if err != nil {
		return nil, err
	}

	if me == nil {
		return nil, errors.New("could not retrieve login information of Razee")
	}

	logger.Info().Str("orgID", me.OrgId).Str(clusterGroupKey, clusterGroup).Msg("Initializing Razee using oauth")

	return &razeeClusterManager{
		orgID:        me.OrgId,
		clusterGroup: clusterGroup,
		con:          con,
		log:          logger,
	}, nil
}

// NewSatConfClusterManager creates an instance of Razee based ClusterManager with Satellite authentication
func NewSatConfClusterManager(apiKey, clusterGroup string) (multicluster.ClusterManager, error) {
	iamClient, err := iam.NewIAMClient(apiKey, "")
	if err != nil {
		return nil, err
	}
	if iamClient == nil {
		return nil, errors.New("the IAMClient returned nil for IBM Cloud Satellite Config")
	}
	con, err := client.New(endPointURL, iamClient.Client)
	if err != nil {
		return nil, err
	}

	me, err := con.Users.Me()
	if err != nil {
		return nil, err
	}

	if me == nil {
		return nil, errors.New("could not retrieve login information of Razee")
	}

	logger := logging.LogInit(logging.CONTROLLER, "RazeeManager")

	logger.Info().Str(clusterGroupKey, clusterGroup).Str("orgID", me.OrgId).Msg("Initializing Razee with IBM Satellite Config")

	return &razeeClusterManager{
		orgID:        me.OrgId,
		clusterGroup: clusterGroup,
		con:          con,
		log:          logger,
	}, nil
}
