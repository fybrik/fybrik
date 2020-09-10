// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0


package com.ibm.egeria;

public final class CustomException extends Exception {
    private String errorReason = "";
    private int httpStatusCode = -1;
    private String errorCode = "-1";
    private String errorMessage = "";

    public CustomException(
            final String errReason,
            final int httpStatCode,
            final String errCode,
            final String errMessage) {
        super(errMessage);

        this.errorReason = errReason;
        this.httpStatusCode = httpStatCode;
        this.errorCode = errCode;
        this.errorMessage = errMessage;

    }

    public CustomException(final String errReason, final int httpStatCode, final int errCode, final String errMessage) {
        // this(errReason, httpStatCode, String.valueOf(errCode), errMessage);
        super(errMessage);

        this.errorReason = errReason;
        this.httpStatusCode = httpStatCode;
        this.errorCode = String.valueOf(errCode);
        this.errorMessage = errMessage;
    }

    public CustomException() {
    }

    public CustomException(final String message) {
        super(message);
        this.errorMessage = message;
    }

    public CustomException(final String message,  final int httpStatCode) {
        super(message);
        this.errorMessage = message;
        this.httpStatusCode = httpStatCode;
    }

    public CustomException(final String message, final Throwable cause) {
        super(message, cause);
    }

    public CustomException(final Throwable cause) {
        super(cause);
    }

    @Override
    public boolean equals(final Object obj) {
        if (obj == null) {
            return false;
        }

        if (!CustomException.class.isAssignableFrom(obj.getClass())) {
            return false;
        }

        final CustomException other = (CustomException) obj;
        if ((this.errorReason == null)
                ? other.getErrorReason() != null
                : !this.errorReason.equals(other.getErrorReason())) {
            return false;
        }

        if ((this.errorMessage == null)
                ? other.getErrorMessage() != null
                : !this.errorMessage.equals(other.getErrorMessage())) {
            return false;
        }

        if (this.httpStatusCode != other.getHttpStatusCode()) {
            return false;
        }

        if (!this.errorCode.equals(other.getErrorCode())) {
            return false;
        }
        return true;
    }

    @Override
    public int hashCode() {
        // Ref: https://stackoverflow.com/questions/8180430/how-to-override-equals-method-in-java
        int hash = 3;
        hash = 53 * hash + (this.errorReason != null ? this.errorReason.hashCode() : 0);
        hash = 53 * hash + (this.errorMessage != null ? this.errorMessage.hashCode() : 0);
        hash = 53 * hash + this.httpStatusCode;
        return 53 * hash + this.errorCode.hashCode();
    }

    public CustomException(
                final String message,
                final Throwable cause,
                final boolean enableSuppression,
                final boolean writableStackTrace) {
        super(message, cause, enableSuppression, writableStackTrace);
    }

    @Override
    public String toString() {
        return "{" + " errorReason='" + getErrorReason() + "'" + ", httpStatusCode='" + getHttpStatusCode() + "'"
                + ", errorCode='" + getErrorCode() + "'" + ", errorMessage='" + getErrorMessage() + "'" + "}";
    }

    public String getErrorReason() {
        return this.errorReason;
    }

    public void setErrorReason(final String errReason) {
        this.errorReason = errReason;
    }

    public int getHttpStatusCode() {
        return this.httpStatusCode;
    }

    public void setHttpStatusCode(final int httpStatCode) {
        this.httpStatusCode = httpStatCode;
    }

    public String getErrorCode() {
        return this.errorCode;
    }

    public void setErrorCode(final String errCode) {
        this.errorCode = errCode;
    }

    public String getErrorMessage() {
        return this.errorMessage;
    }

    public void setErrorMessage(final String errMessage) {
        this.errorMessage = errMessage;
    }

}
