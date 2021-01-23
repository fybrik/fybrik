// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"time"

	"google.golang.org/grpc"

	app "github.com/ibm/the-mesh-for-data/manager/apis/app/v1alpha1"
	"github.com/ibm/the-mesh-for-data/manager/controllers/app/modules"
	"github.com/ibm/the-mesh-for-data/manager/controllers/utils"
	dc "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
)

// GetConnectionDetails calls the data catalog service
func GetConnectionDetails(req *modules.DataInfo, input *app.M4DApplication) error {
	// Set up a connection to the data catalog interface server.
	conn, err := grpc.Dial(utils.GetDataCatalogServiceAddress(), grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return err
	}
	defer conn.Close()
	c := dc.NewDataCatalogServiceClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	response, err := c.GetDatasetInfo(ctx, &dc.CatalogDatasetRequest{
		AppId:     utils.CreateAppIdentifier(input),
		DatasetId: req.Context.DataSetID,
	})
	if err != nil {
		return err
	}

	req.DataDetails = response.GetDetails().DeepCopy()

	return nil
}

// GetCredentials calls the credentials manager service
// TODO: Choose appropriate catalog connector based on the datacatalog service indicated as part of datasetID
func GetCredentials(req *modules.DataInfo, input *app.M4DApplication) error {
	// Set up a connection to the data catalog interface server.
	conn, err := grpc.Dial(utils.GetCredentialsManagerServiceAddress(), grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return err
	}
	defer conn.Close()
	c := dc.NewDataCredentialServiceClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	dataCredentials, err := c.GetCredentialsInfo(ctx, &dc.DatasetCredentialsRequest{
		DatasetId: req.Context.DataSetID,
		AppId:     utils.CreateAppIdentifier(input)})
	if err != nil {
		return err
	}
	req.Credentials = dataCredentials.DeepCopy()

	return nil
}
