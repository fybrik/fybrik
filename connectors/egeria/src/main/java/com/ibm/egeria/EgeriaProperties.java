// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0


package com.ibm.egeria;

public final class EgeriaProperties {
    public static final String EGERIA_SERVER_URL_KEY  = "EGERIA_SERVER_URL";
    public static final String EGERIA_CONNECTOR_PORT_KEY  = "PORT_EGERIA_CONNECTOR";
    public static final String EGERIA_DEFAULT_USERNAME  = "EGERIA_DEFAULT_USERNAME";

    private String egeriaServiseURL;
    private String connectorPort;
    private String egeriaDefaultUserName;

    //Singelton as the properties are set from env. variables and always the same
    private static EgeriaProperties instance;
    public static EgeriaProperties getEgeriaProperties() {
        if (instance == null) {
            instance = new EgeriaProperties();
        }
        return instance;
    }

    private EgeriaProperties() {
        this.egeriaServiseURL = getProperty(EGERIA_SERVER_URL_KEY, true);
        this.connectorPort = getProperty(EGERIA_CONNECTOR_PORT_KEY, false);
        if (this.connectorPort == null) {
            this.connectorPort = EgeriaConnector.EGERIA_DEFAULT_PORT;
        }
        this.egeriaDefaultUserName = getProperty(EGERIA_DEFAULT_USERNAME, false);
    }

    private String getProperty(final String name, final boolean required) {
        String val = System.getenv(name);
        if (val == null && required) {
            throw new AssertionError(name + " not set as environment variable");
        }
        return val;
    }

    public String getEgeriaServiseURL() {
        return egeriaServiseURL;
    }

    public int getConnectorPort() {
        return Integer.parseInt(connectorPort);
    }

    public String getEgeriaDefaultUserName() {
        return egeriaDefaultUserName;
    }
}
