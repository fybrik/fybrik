// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package com.ibm.egeria;

import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.Response;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.Map;
import java.util.HashMap;

import com.datmesh.DataCatalogResponse.CatalogDatasetInfo;
import com.datmesh.DatasetDetailsOuterClass.DatasetDetails;
import com.datmesh.DatasetDetailsOuterClass.DatasetMetadata;
import com.datmesh.DatasetDetailsOuterClass.Db2DataStore;
import com.datmesh.DatasetDetailsOuterClass.KafkaDataStore;
import com.datmesh.DatasetDetailsOuterClass.S3DataStore;
import com.datmesh.DatasetDetailsOuterClass.DataStore.DataStoreType;
import com.datmesh.DatasetDetailsOuterClass.DataComponentMetadata;
import com.datmesh.DatasetDetailsOuterClass.DataStore;

import javax.net.ssl.SSLContext;
import javax.net.ssl.SSLSocketFactory;
import javax.net.ssl.TrustManager;
import javax.net.ssl.X509TrustManager;
import javax.annotation.Nullable;
import javax.net.ssl.HostnameVerifier;
import java.util.concurrent.TimeUnit;
import java.security.KeyManagementException;
import java.security.NoSuchAlgorithmException;
import java.security.cert.CertificateException;
import javax.net.ssl.SSLSession;

import org.apache.commons.lang.exception.ExceptionUtils;

import com.google.gson.JsonArray;
import com.google.gson.JsonObject;
import com.google.gson.JsonParser;
import com.google.gson.JsonSyntaxException;

import java.util.List;
import java.util.ArrayList;

public final class EgeriaClient {
    private static final Logger LOGGER = LoggerFactory.getLogger(EgeriaClient.class.getName());
    public static final String ASSETID_SERVER_NAME = "ServerName";
    public static final String ASSETID_ASSET_GUID = "AssetGuid";

    private String assetIDJson;
    private String serverName;
    private String assetGuid;
    // at this point we make it a constant, this will change and should be passed as
    // credentials for Egeria
    private String userid;

    private EgeriaProperties prop;

    public EgeriaClient(final String assetID)  {
        this.assetIDJson = assetID;
        LOGGER.info("assetID in  EgeriaClient: {}", assetIDJson.replaceAll("[\r\n]", ""));

        JsonObject assetJson = JsonParser.parseString(assetID).getAsJsonObject();
        this.serverName = assetJson.get(ASSETID_SERVER_NAME).getAsString();
        this.assetGuid =  assetJson.get(ASSETID_ASSET_GUID).getAsString();
        // this.assetGuid =  assetJson.get("userid").getAsString();

        this.prop = EgeriaProperties.getEgeriaProperties();

        LOGGER.info("New EgeriaClient created. AssetID={}. serverName={}. assetGUID={}.",
        assetID.replaceAll("[\r\n]", ""), serverName.replaceAll("[\r\n]", ""), assetGuid.replaceAll("[\r\n]", ""));
    }

    public String toString() {
        StringBuilder strbuilder = new StringBuilder();
        strbuilder.append("assetID: ");
        strbuilder.append(assetIDJson);
        return strbuilder.toString();
    }

    public void setEgeriaDefaultUserName(String defaultUserName){
        this.userid = defaultUserName;
    }

    // https://stackoverflow.com/questions/50462157/android-https-urls-are-not-working-in-okhttp3
    private static OkHttpClient generateDefaultOkHttp() {
        OkHttpClient.Builder builder = new OkHttpClient.Builder();
        try {
            // Create a trust manager that does not validate certificate chains
            final TrustManager[] trustAllCerts = new TrustManager[]{
                    new X509TrustManager() {
                        //@SuppressLint("TrustAllX509TrustManager")
                        @Override
                        public void checkClientTrusted(
                            final java.security.cert.X509Certificate[] chain,
                            final String authType) throws CertificateException {
                        }

                        //@SuppressLint("TrustAllX509TrustManager")
                        @Override
                        public void checkServerTrusted(
                            final java.security.cert.X509Certificate[] chain,
                            final String authType) throws CertificateException {
                        }

                        @Override
                        public java.security.cert.X509Certificate[] getAcceptedIssuers() {
                            return new java.security.cert.X509Certificate[]{};
                        }
                    }
            };

            // Install the all-trusting trust manager
            final SSLContext sslContext = SSLContext.getInstance("TLSv1.2");  //"SSL"
            sslContext.init(null, trustAllCerts, new java.security.SecureRandom());
            // Create an ssl socket factory with our all-trusting manager
            final SSLSocketFactory sslSocketFactory = sslContext.getSocketFactory();


            builder.sslSocketFactory(sslSocketFactory, (X509TrustManager) trustAllCerts[0]);
            builder.hostnameVerifier(new HostnameVerifier() {
                //@SuppressLint("BadHostnameVerifier")
                @Override
                public boolean verify(final String hostname,
                                      final SSLSession session) {
                    return true;
                }
            });
        } catch (NoSuchAlgorithmException e) {
            LOGGER.error("SSL context requests algorithm that doesn't exists", e);
        } catch (KeyManagementException e) {
            // e.printStackTrace();
            LOGGER.error("SSL context initiation failed", e);
        }
        builder.connectTimeout(60, TimeUnit.SECONDS)
                .readTimeout(60, TimeUnit.SECONDS)
                .writeTimeout(60, TimeUnit.SECONDS)
                .retryOnConnectionFailure(true);
        return builder.build();
    }

    @Nullable
    private String callEgeriaApi(final String url, final String apiName, final boolean tolerantTo404)
                                                                                throws CustomException, Exception {
        Response response = null;
        try {
            LOGGER.info("Call Egeria API: {}, URL: {}", apiName.replaceAll("[\r\n]", ""), url.replaceAll("[\r\n]", ""));

            // OkHttpClient client = new OkHttpClient();
            OkHttpClient client = generateDefaultOkHttp();
            Request request = new Request.Builder()
                    .url(url)
                    .method("GET", null)
                    .addHeader("content-type", "application/json")
                    .build();
            response = client.newCall(request).execute();
        } catch (Exception e) {
            LOGGER.error("Error calling {} API of EgeriaClient. Error {}. Trace: {}",
                            apiName.replaceAll("[\r\n]", ""),
                            e.toString().replaceAll("[\r\n]", ""),
                            ExceptionUtils.getStackTrace(e).replaceAll("[\r\n]", ""));
            throw e;
        }

        String responseString = null;

        if (response == null) {
            LOGGER.info("Response is NULL calling {} API of EgeriaClient.", apiName.replaceAll("[\r\n]", ""));
            throw new NullPointerException("Response is NULL calling " + apiName);
        }

        if (!response.isSuccessful()) {
            LOGGER.info("Error occurred: Response unsuccessful calling {} API. Response : {}",
                        apiName.replaceAll("[\r\n]", ""),
                        response.toString().replaceAll("[\r\n]", ""));
            throw new CustomException("Response unsuccessful calling " + apiName, response.code());
        }
        try {
            // Get response body
            responseString = response.body().string();

            LOGGER.info("Response from {} API: {}",
                        apiName.replaceAll("[\r\n]", ""),
                        responseString.replaceAll("[\r\n]", ""));

            JsonObject json = JsonParser.parseString(responseString).getAsJsonObject();

            int relatedHttpCode = Integer.parseInt(json.get("relatedHTTPCode").getAsString());

            //in this case internalHttpCode 404 is not an error but wil be translated to toher legitimate value
            if (tolerantTo404 && relatedHttpCode == 404) {
                LOGGER.error("relatedHttpCode is 404 means nothing is found: {}",
                                        String.valueOf(relatedHttpCode).replaceAll("[\r\n]", ""));

                return null;
            }

            if (relatedHttpCode < 200 || relatedHttpCode > 300) {
                LOGGER.error("relatedHttpCode not successfull: {}",
                                        String.valueOf(relatedHttpCode).replaceAll("[\r\n]", ""));

                throw new CustomException("Response unsuccessful calling " + apiName, response.code());
            }
        } catch (IllegalStateException | NullPointerException e) {
            LOGGER.error("Error in parsing response from REST call to {}. Error: {}. Trace: {}",
                            apiName.replaceAll("[\r\n]", ""),
                            e.toString().replaceAll("[\r\n]", ""),
                            ExceptionUtils.getStackTrace(e).replaceAll("[\r\n]", ""));
            throw e;
        }
        return responseString;
    }


    private HashMap<String, DataComponentMetadata> callSchemaAttributesAPI(final String schemaTypeGuid)
                                                                            throws CustomException, Exception {
        LOGGER.info("Call Schema Attributes API");
        LOGGER.info("schemaTypeGuid in callSchemaAttributesAPI: {}", schemaTypeGuid.replaceAll("[\r\n]", ""));

        String serviceURLName = "asset-owner";

        //todo: add paging in case of many columns
        String fullUrl = prop.getEgeriaServiseURL()
                            + "/servers/" + serverName
                            + "/open-metadata/common-services/" + serviceURLName
                            + "/connected-asset/users/"
                            + userid + "/assets/schemas/" + schemaTypeGuid
                            + "/schema-attributes?elementStart=0&maxElements=0";

        HashMap<String, DataComponentMetadata> componentsMetadata = new HashMap<String, DataComponentMetadata>();
        try {
            // Get response body
            String responseString = callEgeriaApi(fullUrl, "callSchemaAttributesAPI", true);
            if (responseString == null) {
                return componentsMetadata; //happens if not columns are found attached
            }

            JsonObject shemaJson = JsonParser.parseString(responseString).getAsJsonObject();
            JsonArray componentsLst = shemaJson.get("list").getAsJsonArray();

            for (int i = 0; i < componentsLst.size(); i++) {
                JsonObject componentJson = componentsLst.get(i).getAsJsonObject();
                if (!"TabularColumn".equals(componentJson.get("type").getAsJsonObject()
                                                            .get("elementTypeName").getAsString())) {
                    continue; //we support today only columns components
                }
                String colName = componentJson.get("attributeName").getAsString();
                String guid = componentJson.get("guid").getAsString();
                String pos = componentJson.get("elementPosition").getAsString();

                Map<String, String> namedMetadataForColumn = new HashMap<String, String>();
                namedMetadataForColumn.put("guid", guid);
                namedMetadataForColumn.put("elementPosition", pos);

                DataComponentMetadata columnMetadata = DataComponentMetadata.newBuilder()
                            .setComponentType("column")
                            .putAllNamedMetadata(namedMetadataForColumn)
                            //.addAllTags(columnTags)
                            .build();
                componentsMetadata.put(colName, columnMetadata);
            }

        } catch (IllegalStateException | NullPointerException e) {
            LOGGER.error("Error in parsing response from REST call "
                        + " in callSchemaAttributesAPI()  of EgeriaClient: {}",
                        e.toString().replaceAll("[\r\n]", ""));
            throw e;
        }

        return componentsMetadata;
    }

    private List<String> callTagsAPI() throws CustomException, Exception {
        LOGGER.info("Call Tags API");

        //todo: add paging in case of many tags
        String fullUrl = prop.getEgeriaServiseURL()
                        + "/servers/" + serverName
                        + "/open-metadata/common-services/asset-consumer/connected-asset/users/"
                        + userid + "/assets/" + assetGuid
                        + "/informal-tags?elementStart=0&maxElements=10";

        List<String> allTagsRelatedToAsset = new ArrayList<String>();
        try {
            // Get response body
            String responseString = callEgeriaApi(fullUrl, "callTagsAPI", true);
            if (responseString == null) {
                return allTagsRelatedToAsset; //no informal tags were found attached
            }

            JsonObject tagsJson = JsonParser.parseString(responseString).getAsJsonObject();
            JsonArray componentsLst = tagsJson.get("list").getAsJsonArray();
            for (int i = 0; i < componentsLst.size(); i++) {
                JsonObject tagJson = componentsLst.get(i).getAsJsonObject();
                if (!"InformalTag".equals(tagJson.get("type").getAsJsonObject()
                                                            .get("elementTypeName").getAsString())) {
                    continue; //we support today only InformalTags
                }
                String tagName = tagJson.get("name").getAsString();
                allTagsRelatedToAsset.add(tagName);
            }
            LOGGER.info("allTagsRelatedToAsset : {}",
                        allTagsRelatedToAsset.toString().replaceAll("[\r\n]", ""));
        } catch (IllegalStateException | NullPointerException e) {
            LOGGER.error("Error in parsing response from REST call "
                        + " in callTagsAPI()  of EgeriaClient: {}",
                        e.toString().replaceAll("[\r\n]", ""));
            throw e;
        }
        return allTagsRelatedToAsset;
    }

    private DataStore getAssetStore(final String qualifiedName) throws CustomException {
        DataStore dataStore = null;
        try {
            JsonObject storeJson = JsonParser.parseString(qualifiedName).getAsJsonObject();
            String storeName = storeJson.get("data_store").getAsString();
            switch (storeName) {
                case "DB2":
                    Db2DataStore storeDB2 = Db2DataStore.newBuilder()
                                        .setUrl(storeJson.get("url").getAsString())
                                        .setDatabase(storeJson.get("database").getAsString())
                                        .setSsl(storeJson.get("ssl").getAsString())
                                        .setPort(storeJson.get("port").getAsString())
                                        .setTable(storeJson.get("table").getAsString())
                                        .build();

                    return DataStore.newBuilder().setType(DataStoreType.DB2)
                                        .setName("DB2")
                                        .setDb2(storeDB2).build();
                case "S3":
                    S3DataStore storeS3 = S3DataStore.newBuilder()
                                        .setBucket(storeJson.get("bucket").getAsString())
                                        .setEndpoint(storeJson.get("endpoint").getAsString())
                                        .setObjectKey(storeJson.get("object_key").getAsString())
                                        .setRegion(storeJson.get("region").getAsString())
                                        .build();

                    return DataStore.newBuilder().setType(DataStoreType.S3)
                                        .setName("object store")
                                        .setS3(storeS3).build();
                case "Kafka":
                    KafkaDataStore store = KafkaDataStore.newBuilder()
                    .setTopicName(storeJson.get("topic_name").toString())
                    .setBootstrapServers(storeJson.get("bootstrap_servers").toString())
                    .setSchemaRegistry(storeJson.get("schema_registry").toString())
                    .setKeyDeserializer(storeJson.get("key_deserializer").toString())
                    .setValueDeserializer(storeJson.get("value_deserializer").toString())
                    .setSecurityProtocol(storeJson.get("security_protocol").toString())
                    .setSaslMechanism(storeJson.get("sasl_mechanism").toString())
                    .setSslTruststore(storeJson.get("ssl_truststore").toString())
                    .setSslTruststorePassword(storeJson.get("ssl_truststore_password").toString())
                    .build();

                    return DataStore.newBuilder().setType(DataStoreType.KAFKA)
                        .setName("kafka")
                        .setKafka(store).build();

                default:
                    LOGGER.info("Unknown data store. Json parsed correctly but data store not recognized: {}",
                            storeName.replaceAll("[\r\n]", ""));
                    throw new CustomException("Unknown data store: " + storeName);
            }
        } catch (JsonSyntaxException e) { //thrown in case it is not a legitimate JSON structure
            //in this case treat asset as LOCAL
            dataStore = DataStore.newBuilder()
                                .setType(DataStoreType.LOCAL)
                                .setName(qualifiedName)
                                .build();
        }
        LOGGER.info("DataStore parsed: {}", dataStore.toString().replaceAll("[\r\n]", ""));
        return dataStore;
    }

    public CatalogDatasetInfo getCatalogDatasetInfo() throws CustomException, Exception {
        LOGGER.info("Call Asset API");
        LOGGER.info("userid in Egeria: " + userid);

        String fullUrl = prop.getEgeriaServiseURL()
                        + "/servers/" + serverName
                        + "/open-metadata/common-services/asset-owner/"
                        + "connected-asset/users/" + userid + "/assets/"
                        + assetGuid;

        String userAssetJson = callEgeriaApi(fullUrl, "GetAssetAPI", false);

        AssetMetaDataHelper assetMetaDataHelper = new AssetMetaDataHelper(userAssetJson);
        LOGGER.info("assetMetaDataHelper : {}", assetMetaDataHelper.toString().replaceAll("[\r\n]", ""));

        String schemaTypeGuid = assetMetaDataHelper.getSchemaTypeGuid();
        HashMap<String, DataComponentMetadata> componentsMetadata = 
                                                callSchemaAttributesAPI(schemaTypeGuid);
        LOGGER.info("listOfColumns in getCatalogDatasetInfo: {}",
                                                    componentsMetadata.toString().replaceAll("[\r\n]", ""));

        List<String> allTagsRelatedToAsset = callTagsAPI();
        LOGGER.info("allTagsRelatedToAsset in getCatalogDatasetInfo: {}",
                                        allTagsRelatedToAsset.toString().replaceAll("[\r\n]", ""));

        // start building response object - egeria
        HashMap<String, String> namedMetadataForDatast = assetMetaDataHelper.getDatasetNamedMetaData();
        DatasetMetadata metadata = DatasetMetadata.newBuilder()
                .addAllDatasetTags(allTagsRelatedToAsset)
                .putAllDatasetNamedMetadata(namedMetadataForDatast)
                .putAllComponentsMetadata(componentsMetadata)
                .build();

        DatasetDetails datasetDetails = null;
        // data owner is not supported in Egeria 2.6 now.
        String dataowner = "";
        // data owner is not supported in Egeria 2.6 now.

        String name = assetMetaDataHelper.getName();

        String typeOfAsset = assetMetaDataHelper.getAssetType(); //maybe will be in additional properties
        if (typeOfAsset == null){
            typeOfAsset = "";
        }

        //it can contain a direct link to the file or a json with remote object
        String qualifiedName = assetMetaDataHelper.getQualifiedName();
        // fix for https://github.com/IBM/the-mesh-for-data/issues/122 - start
        JsonObject storeJson = JsonParser.parseString(qualifiedName).getAsJsonObject();
        String geo = storeJson.get("data_location").getAsString();
        // fix for https://github.com/IBM/the-mesh-for-data/issues/122 - end
        DataStore dataStore = getAssetStore(qualifiedName);

        datasetDetails = DatasetDetails.newBuilder()
                        .setName(name)
                        .setDataOwner(dataowner)
                        .setMetadata(metadata)
                        .setDataStore(dataStore)
                        .setGeo(geo)
                        .setDataFormat(typeOfAsset)
                        .build();


        LOGGER.info("datasetDetails in getCatalogDatasetInfo: {}",
                    datasetDetails.toString().replaceAll("[\r\n]", ""));
        CatalogDatasetInfo  info = CatalogDatasetInfo.newBuilder()
                                    .setDatasetId(assetIDJson)
                                    .setDetails(datasetDetails)
                                    .build();
        LOGGER.info("CatalogDatasetInfo in getCatalogDatasetInfo: {}",
                    info.toString().replaceAll("[\r\n]", ""));
        return CatalogDatasetInfo.newBuilder()
               .setDatasetId(assetIDJson)
               .setDetails(datasetDetails)
               .build();
    }
}
