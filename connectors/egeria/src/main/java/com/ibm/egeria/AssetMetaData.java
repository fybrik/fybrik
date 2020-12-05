// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0


package com.ibm.egeria;

import com.google.gson.annotations.Expose;
import com.google.gson.annotations.SerializedName;
import org.apache.commons.lang3.builder.EqualsBuilder;
import org.apache.commons.lang3.builder.HashCodeBuilder;
import org.apache.commons.lang3.builder.ToStringBuilder;

import com.ibm.generatedsources.asset.Asset;
import com.ibm.generatedsources.asset.SchemaType;
// import com.ibm.generatedsources.asset.Attachment;

public class AssetMetaData {

    @SerializedName("noteLogsCount")
    @Expose
    private Integer noteLogsCount;
    @SerializedName("likeCount")
    @Expose
    private Integer likeCount;
    @SerializedName("certificationCount")
    @Expose
    private Integer certificationCount;
    @SerializedName("licenseCount")
    @Expose
    private Integer licenseCount;
    @SerializedName("relatedHTTPCode")
    @Expose
    private Integer relatedHTTPCode;
    @SerializedName("informalTagCount")
    @Expose
    private Integer informalTagCount;
    @SerializedName("knownLocationsCount")
    @Expose
    private Integer knownLocationsCount;
    @SerializedName("relatedMediaReferenceCount")
    @Expose
    private Integer relatedMediaReferenceCount;
    @SerializedName("asset")
    @Expose
    private Asset asset;
    @SerializedName("class")
    @Expose
    private String _class;
    @SerializedName("commentCount")
    @Expose
    private Integer commentCount;
    @SerializedName("externalReferencesCount")
    @Expose
    private Integer externalReferencesCount;
    @SerializedName("externalIdentifierCount")
    @Expose
    private Integer externalIdentifierCount;
    @SerializedName("connectionCount")
    @Expose
    private Integer connectionCount;
    @SerializedName("schemaType")
    @Expose
    private SchemaType schemaType;
    @SerializedName("relatedAssetCount")
    @Expose
    private Integer relatedAssetCount;
    @SerializedName("ratingsCount")
    @Expose
    private Integer ratingsCount;

    public Integer getNoteLogsCount() {
        return noteLogsCount;
    }

    public void setNoteLogsCount(Integer noteLogsCount) {
        this.noteLogsCount = noteLogsCount;
    }

    public AssetMetaData withNoteLogsCount(Integer noteLogsCount) {
        this.noteLogsCount = noteLogsCount;
        return this;
    }

    public Integer getLikeCount() {
        return likeCount;
    }

    public void setLikeCount(Integer likeCount) {
        this.likeCount = likeCount;
    }

    public AssetMetaData withLikeCount(Integer likeCount) {
        this.likeCount = likeCount;
        return this;
    }

    public Integer getCertificationCount() {
        return certificationCount;
    }

    public void setCertificationCount(Integer certificationCount) {
        this.certificationCount = certificationCount;
    }

    public AssetMetaData withCertificationCount(Integer certificationCount) {
        this.certificationCount = certificationCount;
        return this;
    }

    public Integer getLicenseCount() {
        return licenseCount;
    }

    public void setLicenseCount(Integer licenseCount) {
        this.licenseCount = licenseCount;
    }

    public AssetMetaData withLicenseCount(Integer licenseCount) {
        this.licenseCount = licenseCount;
        return this;
    }

    public Integer getRelatedHTTPCode() {
        return relatedHTTPCode;
    }

    public void setRelatedHTTPCode(Integer relatedHTTPCode) {
        this.relatedHTTPCode = relatedHTTPCode;
    }

    public AssetMetaData withRelatedHTTPCode(Integer relatedHTTPCode) {
        this.relatedHTTPCode = relatedHTTPCode;
        return this;
    }

    public Integer getInformalTagCount() {
        return informalTagCount;
    }

    public void setInformalTagCount(Integer informalTagCount) {
        this.informalTagCount = informalTagCount;
    }

    public AssetMetaData withInformalTagCount(Integer informalTagCount) {
        this.informalTagCount = informalTagCount;
        return this;
    }

    public Integer getKnownLocationsCount() {
        return knownLocationsCount;
    }

    public void setKnownLocationsCount(Integer knownLocationsCount) {
        this.knownLocationsCount = knownLocationsCount;
    }

    public AssetMetaData withKnownLocationsCount(Integer knownLocationsCount) {
        this.knownLocationsCount = knownLocationsCount;
        return this;
    }

    public Integer getRelatedMediaReferenceCount() {
        return relatedMediaReferenceCount;
    }

    public void setRelatedMediaReferenceCount(Integer relatedMediaReferenceCount) {
        this.relatedMediaReferenceCount = relatedMediaReferenceCount;
    }

    public AssetMetaData withRelatedMediaReferenceCount(Integer relatedMediaReferenceCount) {
        this.relatedMediaReferenceCount = relatedMediaReferenceCount;
        return this;
    }

    public Asset getAsset() {
        return asset;
    }

    public void setAsset(Asset asset) {
        this.asset = asset;
    }

    public AssetMetaData withAsset(Asset asset) {
        this.asset = asset;
        return this;
    }

    public String getClass_() {
        return _class;
    }

    public void setClass_(String _class) {
        this._class = _class;
    }

    public AssetMetaData withClass(String _class) {
        this._class = _class;
        return this;
    }

    public Integer getCommentCount() {
        return commentCount;
    }

    public void setCommentCount(Integer commentCount) {
        this.commentCount = commentCount;
    }

    public AssetMetaData withCommentCount(Integer commentCount) {
        this.commentCount = commentCount;
        return this;
    }

    public Integer getExternalReferencesCount() {
        return externalReferencesCount;
    }

    public void setExternalReferencesCount(Integer externalReferencesCount) {
        this.externalReferencesCount = externalReferencesCount;
    }

    public AssetMetaData withExternalReferencesCount(Integer externalReferencesCount) {
        this.externalReferencesCount = externalReferencesCount;
        return this;
    }

    public Integer getExternalIdentifierCount() {
        return externalIdentifierCount;
    }

    public void setExternalIdentifierCount(Integer externalIdentifierCount) {
        this.externalIdentifierCount = externalIdentifierCount;
    }

    public AssetMetaData withExternalIdentifierCount(Integer externalIdentifierCount) {
        this.externalIdentifierCount = externalIdentifierCount;
        return this;
    }

    public Integer getConnectionCount() {
        return connectionCount;
    }

    public void setConnectionCount(Integer connectionCount) {
        this.connectionCount = connectionCount;
    }

    public AssetMetaData withConnectionCount(Integer connectionCount) {
        this.connectionCount = connectionCount;
        return this;
    }

    public SchemaType getSchemaType() {
        return schemaType;
    }

    public void setSchemaType(SchemaType schemaType) {
        this.schemaType = schemaType;
    }

    public AssetMetaData withSchemaType(SchemaType schemaType) {
        this.schemaType = schemaType;
        return this;
    }

    public Integer getRelatedAssetCount() {
        return relatedAssetCount;
    }

    public void setRelatedAssetCount(Integer relatedAssetCount) {
        this.relatedAssetCount = relatedAssetCount;
    }

    public AssetMetaData withRelatedAssetCount(Integer relatedAssetCount) {
        this.relatedAssetCount = relatedAssetCount;
        return this;
    }

    public Integer getRatingsCount() {
        return ratingsCount;
    }

    public void setRatingsCount(Integer ratingsCount) {
        this.ratingsCount = ratingsCount;
    }

    public AssetMetaData withRatingsCount(Integer ratingsCount) {
        this.ratingsCount = ratingsCount;
        return this;
    }

    @Override
    public String toString() {
        return new ToStringBuilder(this).append("noteLogsCount", noteLogsCount).append("likeCount", likeCount).append("certificationCount", certificationCount).append("licenseCount", licenseCount).append("relatedHTTPCode", relatedHTTPCode).append("informalTagCount", informalTagCount).append("knownLocationsCount", knownLocationsCount).append("relatedMediaReferenceCount", relatedMediaReferenceCount).append("asset", asset).append("_class", _class).append("commentCount", commentCount).append("externalReferencesCount", externalReferencesCount).append("externalIdentifierCount", externalIdentifierCount).append("connectionCount", connectionCount).append("schemaType", schemaType).append("relatedAssetCount", relatedAssetCount).append("ratingsCount", ratingsCount).toString();
    }

    @Override
    public int hashCode() {
        return new HashCodeBuilder().append(relatedHTTPCode).append(connectionCount).append(externalIdentifierCount).append(relatedAssetCount).append(likeCount).append(externalReferencesCount).append(noteLogsCount).append(commentCount).append(relatedMediaReferenceCount).append(informalTagCount).append(schemaType).append(ratingsCount).append(knownLocationsCount).append(_class).append(asset).append(certificationCount).append(licenseCount).toHashCode();
    }

    @Override
    public boolean equals(Object other) {
        if (other == this) {
            return true;
        }
        if ((other instanceof AssetMetaData) == false) {
            return false;
        }
        AssetMetaData rhs = ((AssetMetaData) other);
        return new EqualsBuilder().append(relatedHTTPCode, rhs.relatedHTTPCode).append(connectionCount, rhs.connectionCount).append(externalIdentifierCount, rhs.externalIdentifierCount).append(relatedAssetCount, rhs.relatedAssetCount).append(likeCount, rhs.likeCount).append(externalReferencesCount, rhs.externalReferencesCount).append(noteLogsCount, rhs.noteLogsCount).append(commentCount, rhs.commentCount).append(relatedMediaReferenceCount, rhs.relatedMediaReferenceCount).append(informalTagCount, rhs.informalTagCount).append(schemaType, rhs.schemaType).append(ratingsCount, rhs.ratingsCount).append(knownLocationsCount, rhs.knownLocationsCount).append(_class, rhs._class).append(asset, rhs.asset).append(certificationCount, rhs.certificationCount).append(licenseCount, rhs.licenseCount).isEquals();
    }

}
