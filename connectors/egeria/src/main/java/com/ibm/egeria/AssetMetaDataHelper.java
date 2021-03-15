// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0


package com.ibm.egeria;

import com.google.gson.Gson;
import java.util.HashMap;


public final class AssetMetaDataHelper {
    private AssetMetaData assetMetaDataObject;

    public String toString() {
        return assetMetaDataObject.toString();
    }

    public AssetMetaDataHelper(final String userJson) throws CustomException {
        Gson gson = new Gson();
        assetMetaDataObject = gson.fromJson(userJson, AssetMetaData.class);
    }

    public String getDataSetID() {
        return assetMetaDataObject.getAsset().getGuid();
    }

    public String getName() {
        return assetMetaDataObject.getAsset().getDisplayName();
    }

    public String getQualifiedName() {
        return assetMetaDataObject.getAsset().getQualifiedName();
    }

    public String getAssetType() {
        // getEncodingStandard() is not supported in Egeria 2.6. Using getElementTypeName() - start
        String dataFormat = assetMetaDataObject.getAsset().getType().getElementTypeName();
        // getEncodingStandard() is not supported in Egeria 2.6. Using getElementTypeName() - end
        if ("CSVFile".equalsIgnoreCase(dataFormat)) { //csv we return in specific format expected by pilot
            return "csv";
        }
        return dataFormat;
    }

    public String getSchemaTypeGuid() {
        if (assetMetaDataObject.getSchemaType() == null) {
            return "";
        }
        return assetMetaDataObject.getSchemaType().getGuid();
    }

    public HashMap<String, String> getDatasetNamedMetaData() {
        HashMap<String, String> namedMetaData = new HashMap<String, String>();

        String description = assetMetaDataObject.getAsset().getDescription();
        namedMetaData.put("description", description);

        return namedMetaData;
    }
}
