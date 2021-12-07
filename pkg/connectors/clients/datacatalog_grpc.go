// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	app "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	pb "fybrik.io/fybrik/pkg/connectors/protobuf"

	"emperror.dev/errors"
	datacatalogTaxonomyModels "fybrik.io/fybrik/pkg/taxonomy/model/datacatalog/base"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Ensure that grpcDataCatalog implements the DataCatalog interface
var _ DataCatalog = (*grpcDataCatalog)(nil)

type grpcDataCatalog struct {
	pb.UnimplementedDataCatalogServiceServer

	name       string
	connection *grpc.ClientConn
	client     pb.DataCatalogServiceClient
}

// NewGrpcDataCatalog creates a DataCatalog facade that connects to a GRPC service
// You must call .Close() when you are done using the created instance
func NewGrpcDataCatalog(name string, connectionURL string, connectionTimeout time.Duration) (DataCatalog, error) {
	ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
	defer cancel()
	connection, err := grpc.DialContext(ctx, connectionURL, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("NewGrpcDataCatalog failed when connecting to %s", connectionURL))
	}
	return &grpcDataCatalog{
		name:       name,
		client:     pb.NewDataCatalogServiceClient(connection),
		connection: connection,
	}, nil
}

func (m *grpcDataCatalog) GetAssetInfo(
	in *datacatalogTaxonomyModels.DataCatalogRequest, creds string) (*datacatalogTaxonomyModels.DataCatalogResponse, error) {
	log.Println("open api request received for getting policy decisions: ", *in)
	dataCatalogReq, _ := ConvertDataCatalogOpenAPIReqToGrpcReq(in, creds)
	log.Println("grpc data catalog request to be used for getting asset info: ", dataCatalogReq)
	result, err := m.client.GetDatasetInfo(context.Background(), dataCatalogReq)
	errStatus, _ := status.FromError(err)
	if errStatus.Code() == codes.InvalidArgument {
		return nil, errors.New(app.InvalidAssetID)
	}
	log.Println("GRPC result returned from GetAssetInfo:")
	responseBytes, errJSON := json.MarshalIndent(result, "", "\t")
	if errJSON != nil {
		return nil, fmt.Errorf("error Marshalling External Catalog Connector Response: %v", errJSON)
	}
	log.Print(string(responseBytes))
	dataCatalogResp, err := ConvertDataCatalogGrpcRespToOpenAPIResp(result)
	if err != nil {
		log.Println("Error during conversion to open api response: ", err)
		return nil, err
	}

	res, err := json.MarshalIndent(dataCatalogResp, "", "\t")
	if err != nil {
		log.Println("Error during marshalling data catalog response: ", err)
		return nil, err
	}
	log.Println("Marshalled value of data catalog response: ", string(res))
	return dataCatalogResp, nil
	// return result, err
}

func (m *grpcDataCatalog) RegisterDatasetInfo(ctx context.Context, in *pb.RegisterAssetRequest) (*pb.RegisterAssetResponse, error) {
	result, err := m.client.RegisterDatasetInfo(ctx, in)
	return result, errors.Wrap(err, fmt.Sprintf("register dataset info in %s failed", m.name))
}

func (m *grpcDataCatalog) Close() error {
	return m.connection.Close()
}

func ConvertDataCatalogOpenAPIReqToGrpcReq(in *datacatalogTaxonomyModels.DataCatalogRequest, creds string) (*pb.CatalogDatasetRequest, error) {

	dataCatalogReq := &pb.CatalogDatasetRequest{
		CredentialPath: creds, DatasetId: in.GetAssetID()}
	log.Println("Constructed GRPC data catalog request: ", dataCatalogReq)

	return dataCatalogReq, nil
}

func ConvertDataCatalogGrpcRespToOpenAPIResp(result *pb.CatalogDatasetInfo) (*datacatalogTaxonomyModels.DataCatalogResponse, error) {
	// convert GRPC response to Open Api Response - start
	resourceCols := make([]datacatalogTaxonomyModels.ResourceColumns, 0)

	for colName, compMetaData := range result.GetDetails().Metadata.GetComponentsMetadata() {
		if compMetaData != nil {
			tagsInResp := make(map[string]interface{})
			tagsInResp["tags"] = compMetaData.GetTags()
			rscCol := datacatalogTaxonomyModels.ResourceColumns{
				Name: colName,
				Tags: &tagsInResp}
			resourceCols = append(resourceCols, rscCol)
		}
	}

	datasetTags := make([]interface{}, 0)
	tags := result.GetDetails().Metadata.DatasetTags
	for i := 0; i < len(tags); i++ {
		datasetTags = append(datasetTags, tags[i])
	}
	tagsInResponse := make(map[string]interface{})
	tagsInResponse["tags"] = datasetTags

	resourceMetaData := &datacatalogTaxonomyModels.Resource{
		Name:      result.GetDetails().Name,
		Owner:     &result.GetDetails().DataOwner,
		Geography: &result.GetDetails().Geo,
		Tags:      &tagsInResponse,
		Columns:   &resourceCols,
	}

	var additionalProp map[string]interface{}
	if result.GetDetails().DataStore.Type == pb.DataStore_S3 {
		dsStore := result.GetDetails().GetDataStore().S3
		dataStoreInfo, _ := json.Marshal(dsStore)
		json.Unmarshal(dataStoreInfo, &additionalProp)
	} else if result.GetDetails().DataStore.Type == pb.DataStore_KAFKA {
		dsStore := result.GetDetails().GetDataStore().Kafka
		dataStoreInfo, _ := json.Marshal(dsStore)
		json.Unmarshal(dataStoreInfo, &additionalProp)
	} else if result.GetDetails().DataStore.Type == pb.DataStore_DB2 {
		dsStore := result.GetDetails().GetDataStore().Db2
		dataStoreInfo, _ := json.Marshal(dsStore)
		json.Unmarshal(dataStoreInfo, &additionalProp)
	} else { // DataStore.LOCAL
		additionalProp = make(map[string]interface{})
	}

	connection := datacatalogTaxonomyModels.Connection{
		Name:                 result.GetDetails().DataStore.Name,
		AdditionalProperties: additionalProp,
	}

	details := datacatalogTaxonomyModels.Details{
		Connection: connection,
		DataFormat: &result.GetDetails().DataFormat,
	}
	dataCatalogResp := &datacatalogTaxonomyModels.DataCatalogResponse{
		ResourceMetadata: *resourceMetaData,
		Details:          details,
		Credentials:      result.GetDetails().CredentialsInfo.VaultSecretPath,
	}
	// convert GRPC response to Open Api Response - end

	log.Println("dataCatalogResp in ConvertDataCatalogGrpcRespToOpenAPIResp", dataCatalogResp)

	return dataCatalogResp, nil
}
