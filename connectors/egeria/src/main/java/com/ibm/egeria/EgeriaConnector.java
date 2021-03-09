// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0


package com.ibm.egeria;

import io.grpc.Server;
import io.grpc.ServerBuilder;
import io.grpc.stub.StreamObserver;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.datmesh.DataCatalogResponse.CatalogDatasetInfo;
import com.datmesh.DataCatalogRequest.CatalogDatasetRequest;

import com.datmesh.DataCatalogServiceGrpc;

import org.apache.commons.lang.exception.ExceptionUtils;
import org.json.simple.parser.ParseException;

// using standard GRPC error codes for sending error information to M4D manager
import io.grpc.Status;

// for getting http error codes
import java.net.HttpURLConnection;

/**
 * Server that manages startup/shutdown of a {@code Greeter} server.
 */
public final class EgeriaConnector {
  private static final Logger LOGGER = LoggerFactory.getLogger(EgeriaConnector.class.getName());

  public static final String EGERIA_DEFAULT_PORT = "50084"; //synched with egeria_connector.yaml

  /* The port on which the server should run */
  private int port;
  private Server server;
  private String egeriaDefaultUserName;

  public String toString() {
      StringBuilder strbuilder = new StringBuilder();
      strbuilder.append("port: ");
      strbuilder.append(String.valueOf(port));
      strbuilder.append("egeriaDefaultUserName: ");
      strbuilder.append(egeriaDefaultUserName);
      return strbuilder.toString();
  }

  public void setEgeriaDefaultUserName(final String egeriaDefaultUserName) {
    this.egeriaDefaultUserName = egeriaDefaultUserName;
  }

  public String getEgeriaDefaultUserName() {
    return this.egeriaDefaultUserName;
  }

  public void setPort(final int prt) {
    this.port = prt;
  }

  public int getPort() {
    return this.port;
  }

  private String getProperty(final String name) {
    String val = System.getenv(name);
    if (val == null) {
        throw new AssertionError(name + " not set as environment variable");
    }
    return val;
  }

  private void start() throws Exception {
    server = ServerBuilder.forPort(port)
                          .addService(new DataCatalogImpl())
                          .build().start();
    LOGGER.info("Server started, listening on {}", String.valueOf(port));
    Runtime.getRuntime().addShutdownHook(new Thread() {
      @Override
      public void run() {
        // Use stderr here since the logger may have been reset by its JVM shutdown hook.
        System.err.println("*** shutting down gRPC server since JVM is shutting down");
        EgeriaConnector.this.stop();
        System.err.println("*** server shut down");
      }
    });
  }

  private void stop() {
    if (server != null) {
      server.shutdown();
    }
  }

  /**
   * Await termination on the main thread since the grpc library uses daemon
   * threads.
   */
  private void blockUntilShutdown() throws InterruptedException {
    if (server != null) {
      server.awaitTermination();
    }
  }


  private class DataCatalogImpl extends DataCatalogServiceGrpc.DataCatalogServiceImplBase {

    @Override
    public void getDatasetInfo(
                  final CatalogDatasetRequest req,
                  final StreamObserver<CatalogDatasetInfo> responseObserver) {

      String datasetID = req.getDatasetId();

      CatalogDatasetInfo reply = null;

      try {
        // now creating a new object of EgeriaClient for a particular request
        // instead of using a static object. This will enable specific handling of a
        // particular request and may also help when we test scenarios where multiple
        // simultaneous requests can come to EgeriaConnector.
        EgeriaClient egeriaClient = new EgeriaClient(datasetID);
        egeriaClient.setEgeriaDefaultUserName(egeriaDefaultUserName);
        reply = egeriaClient.getCatalogDatasetInfo();
        LOGGER.info(
          "Reply Got from EgeriaClient: {}. Sending the Catalog Dataset Info response. ",
          reply.toString().replaceAll("[\r\n]", ""));

        responseObserver.onNext(reply);
        responseObserver.onCompleted();
      } catch (CustomException e) {
        if (e.getHttpStatusCode() == HttpURLConnection.HTTP_NOT_FOUND) {
          responseObserver.onError(Status.INVALID_ARGUMENT.withDescription(e.toString()).asRuntimeException());
        } else if (e.getHttpStatusCode() == HttpURLConnection.HTTP_UNAVAILABLE) {
          responseObserver.onError(Status.UNAVAILABLE.withDescription(e.toString()).asRuntimeException());
        } else if (e.getHttpStatusCode() == HttpURLConnection.HTTP_UNAUTHORIZED) {
          responseObserver.onError(Status.PERMISSION_DENIED.withDescription(e.toString()).asRuntimeException());
        } else {
          responseObserver.onError(
              Status.INTERNAL.withDescription("Internal Error in Catalog Connector (Java Server): " + e)
                  .asRuntimeException());
        }
      } catch (ParseException e) {
        LOGGER.error(
          "Exception during parse of assetID. AssetID = {}. Error: {}, Trace: {}",
                                                          datasetID.replaceAll("[\r\n]", ""),
                                                          e.toString().replaceAll("[\r\n]", ""),
                                                          ExceptionUtils.getStackTrace(e).replaceAll("[\r\n]", ""));
      } catch (Exception e) {
        LOGGER.error(
            "exception in getDatasetInfo of CatalogConnectorImpl: {}, Trace: {}",
            e.toString().replaceAll("[\r\n]", ""), ExceptionUtils.getStackTrace(e).replaceAll("[\r\n]", ""));
        responseObserver.onError(
            Status.INTERNAL.withDescription(
                  "Internal Error in Catalog Connector (Java Server): "
                  + e.getMessage()).asRuntimeException());
      }
    }
  }

  /**
   * Main launches the server from the command line.
   */
  public static void main(final String[] args) throws Exception {
    //properties set from the environment variables
    EgeriaProperties properties = EgeriaProperties.getEgeriaProperties();
    int port = -1;
    String egeriaDefaultUserName = null;

    try {
      port = properties.getConnectorPort();
      egeriaDefaultUserName = properties.getEgeriaDefaultUserName();
    } catch (Exception e) {
      LOGGER.info("Error during parsing {} env variable.", EgeriaProperties.EGERIA_CONNECTOR_PORT_KEY);
      throw e;
    }

    final EgeriaConnector server = new EgeriaConnector();
    try {
      server.setPort(port);
      server.setEgeriaDefaultUserName(egeriaDefaultUserName);
      LOGGER.info("Using port number {} to start the EgeriaConnector Server",
                  String.valueOf(port));

      server.start();
      server.blockUntilShutdown();
    } catch (Exception e) {
        LOGGER.error("Exception in  "
        + "in main() of EgeriaConnector : {}",
        e.toString().replaceAll("[\r\n]", ""));
        throw e;
    }
  }
}
