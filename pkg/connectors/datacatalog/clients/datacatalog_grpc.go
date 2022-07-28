// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"emperror.dev/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	app "fybrik.io/fybrik/manager/apis/app/v12"
	pb "fybrik.io/fybrik/pkg/connectors/protobuf"
	"fybrik.io/fybrik/pkg/model/datacatalog"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/serde"
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
func NewGrpcDataCatalog(name, connectionURL string, connectionTimeout time.Duration) (DataCatalog, error) {
	ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
	defer cancel()
	connection, err := grpc.DialContext(ctx, connectionURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("NewGrpcDataCatalog failed when connecting to %s", connectionURL))
	}
	return &grpcDataCatalog{
		name:       name,
		client:     pb.NewDataCatalogServiceClient(connection),
		connection: connection,
	}, nil
}

func (m *grpcDataCatalog) GetAssetInfo(in *datacatalog.GetAssetRequest, creds string) (*datacatalog.GetAssetResponse, error) {
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
	if responseBytes == nil {
		return nil, fmt.Errorf("data asset does not exist")
	}
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
}

func (m *grpcDataCatalog) RegisterDatasetInfo(ctx context.Context, in *pb.RegisterAssetRequest) (*pb.RegisterAssetResponse, error) {
	result, err := m.client.RegisterDatasetInfo(ctx, in)
	return result, errors.Wrap(err, fmt.Sprintf("register dataset info in %s failed", m.name))
}

func (m *grpcDataCatalog) Close() error {
	return m.connection.Close()
}

func ConvertDataCatalogOpenAPIReqToGrpcReq(in *datacatalog.GetAssetRequest, creds string) (*pb.CatalogDatasetRequest, error) {
	dataCatalogReq := &pb.CatalogDatasetRequest{
		CredentialPath: creds, DatasetId: string(in.AssetID)}
	log.Println("Constructed GRPC data catalog request: ", dataCatalogReq)

	return dataCatalogReq, nil
}

//nolint:funlen
func ConvertDataCatalogGrpcRespToOpenAPIResp(result *pb.CatalogDatasetInfo) (*datacatalog.GetAssetResponse, error) {
	// convert GRPC response to Open Api Response - start
	resourceCols := make([]datacatalog.ResourceColumn, 0)

	//nolint:revive // Ignore add-constant to log msg
	for colName, compMetaData := range result.GetDetails().Metadata.GetComponentsMetadata() {
		if compMetaData == nil {
			continue
		}

		rscCol := datacatalog.ResourceColumn{
			Name: colName}
		rsColMap := make(map[string]interface{})
		tags := compMetaData.GetTags()
		for i := 0; i < len(tags); i++ {
			rsColMap[tags[i]] = true
		}

		responseBytes, errJSON := json.MarshalIndent(rsColMap, "", "\t")
		if errJSON != nil {
			return nil, fmt.Errorf("error Marshalling in ConvertDataCatalogGrpcRespToOpenAPIResp: %v", errJSON)
		}
		log.Print("responseBytes after MarshalIndent in ConvertDataCatalogGrpcRespToOpenAPIResp:" + string(responseBytes))

		if err := json.Unmarshal(responseBytes, &rscCol.Tags); err != nil {
			return nil, fmt.Errorf("error UnMarshalling in ConvertDataCatalogGrpcRespToOpenAPIResp: %v", errJSON)
		}

		// just printing - start
		responseBytes, errJSON = json.MarshalIndent(&rscCol, "", "\t")
		if errJSON != nil {
			return nil, fmt.Errorf("error Marshalling in ConvertDataCatalogGrpcRespToOpenAPIResp: %v", errJSON)
		}
		log.Print("responseBytes after MarshalIndent in ConvertDataCatalogGrpcRespToOpenAPIResp:" + string(responseBytes))
		// just printing - end

		resourceCols = append(resourceCols, rscCol)
	}

	tags := result.GetDetails().Metadata.DatasetTags
	tagsInResponse := taxonomy.Tags{}
	tagsInResponse.Items = make(map[string]interface{}, len(tags))
	for i := 0; i < len(tags); i++ {
		tagsInResponse.Items[tags[i]] = true
	}

	resourceMetaData := &datacatalog.ResourceMetadata{
		Name:      result.GetDetails().Name,
		Owner:     result.GetDetails().DataOwner,
		Geography: result.GetDetails().Geo,
		Tags:      &tagsInResponse,
		Columns:   resourceCols,
	}

	additionalProp := make(map[string]interface{})
	var connectionDetails interface{}
	var connectionName string
	var err error
	switch result.GetDetails().DataStore.Type {
	case pb.DataStore_S3:
		dsStore := result.GetDetails().GetDataStore().S3
		dataStoreInfo, _ := json.Marshal(dsStore)
		err = json.Unmarshal(dataStoreInfo, &connectionDetails)
		connectionName = "s3"
		additionalProp[connectionName] = connectionDetails
	case pb.DataStore_KAFKA:
		dsStore := result.GetDetails().GetDataStore().Kafka
		dataStoreInfo, _ := json.Marshal(dsStore)
		err = json.Unmarshal(dataStoreInfo, &connectionDetails)
		connectionName = "kafka"
		additionalProp[connectionName] = connectionDetails
	case pb.DataStore_DB2:
		dsStore := result.GetDetails().GetDataStore().Db2
		dataStoreInfo, _ := json.Marshal(dsStore)
		err = json.Unmarshal(dataStoreInfo, &connectionDetails)
		connectionName = "db2"
		additionalProp[connectionName] = connectionDetails
	default: // DataStore.LOCAL
		// log something
		connectionName = "local"
	}
	if err != nil {
		return nil, errors.New("error during unmarshal of dataStoreInfo")
	}

	connection := taxonomy.Connection{
		Name: taxonomy.ConnectionType(connectionName),
		AdditionalProperties: serde.Properties{
			Items: additionalProp,
		},
	}

	details := datacatalog.ResourceDetails{
		Connection: connection,
		DataFormat: taxonomy.DataFormat(result.Details.DataFormat),
	}
	dataCatalogResp := &datacatalog.GetAssetResponse{
		ResourceMetadata: *resourceMetaData,
		Details:          details,
		Credentials:      result.GetDetails().CredentialsInfo.VaultSecretPath,
	}
	// convert GRPC response to Open Api Response - end

	log.Println("dataCatalogResp in ConvertDataCatalogGrpcRespToOpenAPIResp", dataCatalogResp)

	return dataCatalogResp, nil
}

// just adding this dummy implementation as we are going to remove grpc support soon.
// Then this file will be removed. Till then we provide a dummy implementation.
func (m *grpcDataCatalog) CreateAsset(in *datacatalog.CreateAssetRequest, creds string) (*datacatalog.CreateAssetResponse, error) {
	return &datacatalog.CreateAssetResponse{AssetID: "testAssetID"}, nil
}

// just adding this dummy implementation as we are going to remove grpc support soon.
// Then this file will be removed. Till then we provide a dummy implementation.
func (m *grpcDataCatalog) DeleteAsset(in *datacatalog.DeleteAssetRequest, creds string) (*datacatalog.DeleteAssetResponse, error) {
	return &datacatalog.DeleteAssetResponse{Status: "DeleteAsset not implemented via GRPC"}, nil
}

// just adding this dummy implementation as we are going to remove grpc support soon.
// Then this file will be removed. Till then we provide a dummy implementation.
func (m *grpcDataCatalog) UpdateAsset(in *datacatalog.UpdateAssetRequest, creds string) (*datacatalog.UpdateAssetResponse, error) {
	return &datacatalog.UpdateAssetResponse{Status: "UpdateAsset not implemented via GRPC"}, nil
}
