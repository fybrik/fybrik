{% set arrowFlightRelease = arrow_flight_module_version(FybrikRelease,arrowFlight) %}
{% set currentRelease = fybrik_version(FybrikRelease) %}
{% set currentImageTag = fybrik_image_version(FybrikRelease) %}

# FybrikModule chaining sample

This sample shows how to implement a use case where, based on the data source and governance policies, the Fybrik manager determines that it must deploy two FybrikModules to allow a workload access to a dataset. One FybrikModule handles reading the data and the second does the data transformation. Data is passed between the FybrikModules without writing to intermediate storage.

The data read in this example is the `userdata` dataset, a Parquet file found in https://github.com/Teradata/kylo/blob/master/samples/sample-data/parquet/userdata2.parquet. Two FybrikModules are available for use by the Fybrik control plane: the [arrow-flight-module](https://github.com/fybrik/arrow-flight-module) and the [airbyte-module](https://github.com/fybrik/airbyte-module). Only the airbyte-module can give read access to the dataset. However, it does not have any data transformation capabilities. Therefore, to satisfy constraints, the Fybrik manager must deploy both modules: the airbyte module for reading the dataset, and the arrow-flight-module for transforming the dataset based on the governance policies.

To recreate this scenario, you will need a copy of the Fybrik repository (`git clone https://github.com/fybrik/fybrik.git`), and a copy of the airbyte-module repository (`git clone https://github.com/fybrik/airbyte-module.git`). Set the following environment variables: FYBRIK_DIR for the path of the `fybrik` directory, and AIRBYTE_MODULE_DIR for the path of the `airbyte-module` directory.

1. Install Fybrik Prerequisites. Follow the instruction in the Fybrik [Quick Start Guide](https://fybrik.io/dev/get-started/quickstart/). Stop before the "Install control plane" section.

1. Before installing the control plane, we need to customize the [Fybrik taxonomy](https://fybrik.io/dev/tasks/custom-taxonomy/) to define the new connection and interface types. Run:
    ```bash
    cd $FYBRIK_DIR
    git checkout {{ currentRelease|default('master') }}
    go run main.go taxonomy compile --out custom-taxonomy.json --base charts/fybrik/files/taxonomy/taxonomy.json $AIRBYTE_MODULE_DIR/fybrik/fybrik-taxonomy-customize.yaml
    helm install fybrik-crd charts/fybrik-crd -n fybrik-system --wait
    helm install fybrik charts/fybrik --set coordinator.catalog=openmetadata --set openmetadataConnector.openmetadata_endpoint=http://openmetadata.open-metadata:8585/api --set global.tag={{ currentImageTag|default('master') }} -n fybrik-system --wait --set-file taxonomyOverride=custom-taxonomy.json
    ```

1. Install the Airbyte module:
    ```bash
    kubectl apply -f $AIRBYTE_MODULE_DIR/module.yaml -n fybrik-system
    ```

1. Install the arrow-flight module for transformations:
```bash
{% if arrowFlightRelease != 'latest' %}
  kubectl apply -f https://github.com/fybrik/arrow-flight-module/releases/download/{{ arrowFlightRelease }}/module.yaml -n fybrik-system
{% else %}
  kubectl apply -f https://github.com/fybrik/arrow-flight-module/releases/{{ arrowFlightRelease }}/download/module.yaml -n fybrik-system
{% endif %}
```

1. Create a new namespace for the application, and set it as default:
    ```bash
    kubectl create namespace fybrik-airbyte-sample
    kubectl config set-context --current --namespace=fybrik-airbyte-sample
    ```

1. Next, register the data asset itself in the data catalog. We use port-forwarding to send asset creation requests to the OpenMetadata connector.
    ```bash
    kubectl port-forward svc/openmetadata-connector -n fybrik-system 8081:8080 &
    cat << EOF | curl -X POST localhost:8081/createAsset -d @-
    {
      "destinationCatalogID": "openmetadata",
      "destinationAssetID": "userdata",
      "details": {
        "dataFormat": "csv",
        "connection": {
          "name": "file",
          "file": {
            "connector": "airbyte/source-file",
            "dataset_name": "userdata",
            "format": "parquet",
            "url": "https://github.com/Teradata/kylo/raw/master/samples/sample-data/parquet/userdata2.parquet",
            "provider": {
              "storage": "HTTPS"
            }
          }
        }
      },
      "resourceMetadata": {
        "name": "test data",
        "geography": "theshire ",
        "tags": {
          "Purpose.finance": "true"
        },
        "columns": [
          {
            "name": "first_name",
            "tags": {
              "PII.Sensitive": "true"
            }
          },
          {
            "name": "last_name",
            "tags": {
              "PII.Sensitive": "true"
            }
          },
          {
            "name": "birthdate",
            "tags": {
              "PII.Sensitive": "true"
            }
          }
        ]
      }
    }
    EOF
    ```

    The response from the OpenMetadata connector should look like this:
    ```bash
    {"assetID":"openmetadata-file.default.openmetadata.userdata"}
    ```

    The asset is now registered in the catalog. Store the asset ID in a `CATALOGED_ASSET` variable:
    ```bash
    CATALOGED_ASSET="openmetadata-file.default.openmetadata.userdata"
    ```

1. Create the policy to access the asset (we use a policy that requires redactions of `PII.Sensitive` columns):
   ```bash
   kubectl -n fybrik-system create configmap sample-policy --from-file=$AIRBYTE_MODULE_DIR/fybrik/sample-policy-restrictive.rego
   kubectl -n fybrik-system label configmap sample-policy openpolicyagent.org/policy=rego
   while [[ $(kubectl get cm sample-policy -n fybrik-system -o 'jsonpath={.metadata.annotations.openpolicyagent\.org/policy-status}') != '{"status":"ok"}' ]]; do echo "waiting for policy to be applied" && sleep 5; done
   ```

1. Create a [`FybrikApplication`](../reference/crds.md#fybrikapplication) resource to register the workload to the control plane of Fybrik. The value you place in the `dataSetID` field is your asset ID, as explained above.

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: app.fybrik.io/v1beta1
    kind: FybrikApplication
    metadata:
      name: my-app
      labels:
        app: my-app
    spec:
      selector:
        workloadSelector:
          matchLabels:
            app: my-app
      appInfo:
        intent: Fraud Detection
      data:
        - dataSetID: ${CATALOGED_ASSET}
          requirements:
            interface:
              protocol: fybrik-arrow-flight
    EOF
    ```

1. After the FybrikApplication is applied, the Fybrik control plane attempts to create the data path for the application. Fybrik realizes that the Airbyte module can give the application access to the `userdata` dataset, and that the arrow-flight module could provide the redaction transformation. Fybrik deploys both modules in the `fybrik-blueprints` namespace. To verify that the Airbyte module and the arrow-flight module were indeed deployed, run:
   ```bash
   kubectl get pods -n fybrik-blueprints
   ```
   You should see pods with names similar to:
   ```bash
   NAME                                                              READY   STATUS    RESTARTS   AGE
   my-app-fybrik-airbyte-sample-airbyte-module-airbyte-module4kvrq   2/2     Running   0          43s
   my-app-fybrik-airbyte-sample-arrow-flight-module-arrow-flibxsq2   1/1     Running   0          43s
   ```

1. Wait for the FybrikModule pods to be ready by running:
   ```bash
   kubectl wait pod --all --for=condition=ready -n fybrik-blueprints --timeout 10m
   ```

1. To verify that the Airbyte module gives access to the `userdata` dataset, run:
   ```bash
   cd $AIRBYTE_MODULE_DIR/helm/client
   ./deploy_airbyte_module_client_pod.sh
   kubectl exec -it my-shell -n default -- python3 /root/client.py --host my-app-fybrik-airbyte-sample-arrow-flight-module.fybrik-blueprints --port 80 --asset ${CATALOGED_ASSET}
   ```
   You should see the following output:
   ```bash
          registration_dttm      id first_name last_name  email  ...     country birthdate     salary                     title comments
   0    2016-02-03T13:36:39     1.0      XXXXX     XXXXX  XXXXX  ...   Indonesia     XXXXX  140249.37  Senior Financial Analyst         
   1    2016-02-03T00:22:28     2.0      XXXXX     XXXXX  XXXXX  ...       China     XXXXX        NaN                                   
   2    2016-02-03T18:29:04     3.0      XXXXX     XXXXX  XXXXX  ...      France     XXXXX  236219.26                   Teacher         
   3    2016-02-03T13:42:19     4.0      XXXXX     XXXXX  XXXXX  ...      Russia     XXXXX        NaN    Nuclear Power Engineer         
   4    2016-02-03T00:15:29     5.0      XXXXX     XXXXX  XXXXX  ...      France     XXXXX   50210.02             Senior Editor         
   ..                   ...     ...        ...       ...    ...  ...         ...       ...        ...                       ...      ...
   995  2016-02-03T13:36:49   996.0      XXXXX     XXXXX  XXXXX  ...       China     XXXXX  185421.82                                  "
   996  2016-02-03T04:39:01   997.0      XXXXX     XXXXX  XXXXX  ...    Malaysia     XXXXX  279671.68                                   
   997  2016-02-03T00:33:54   998.0      XXXXX     XXXXX  XXXXX  ...      Poland     XXXXX  112275.78                                   
   998  2016-02-03T00:15:08   999.0      XXXXX     XXXXX  XXXXX  ...  Kazakhstan     XXXXX   53564.76        Speech Pathologist         
   999  2016-02-03T00:53:53  1000.0      XXXXX     XXXXX  XXXXX  ...     Nigeria     XXXXX  239858.70                                   
   
   [1000 rows x 13 columns]
   ```

1. Alternatively, one can access the `userdata` dataset from a Jupyter notebook, as described in the [notebook sample](../notebook-read/#read-the-dataset-from-the-notebook). To determine the virtual endpoint from which to access the data set, run:
   ```bash
   CATALOGED_ASSET_MODIFIED=$(echo $CATALOGED_ASSET | sed 's/\./\\\./g')
   ENDPOINT_SCHEME=$(kubectl get fybrikapplication my-app -o jsonpath={.status.assetStates.${CATALOGED_ASSET_MODIFIED}.endpoint.fybrik-arrow-flight.scheme})
   ENDPOINT_HOSTNAME=$(kubectl get fybrikapplication my-app -o jsonpath={.status.assetStates.${CATALOGED_ASSET_MODIFIED}.endpoint.fybrik-arrow-flight.hostname})
   ENDPOINT_PORT=$(kubectl get fybrikapplication my-app -o jsonpath={.status.assetStates.${CATALOGED_ASSET_MODIFIED}.endpoint.fybrik-arrow-flight.port})
   printf "${ENDPOINT_SCHEME}://${ENDPOINT_HOSTNAME}:${ENDPOINT_PORT}"
   ```

   Note that the virtual endpoint determined from the FybrikApplication status points to the arrow-flight transform module, although this is transparent to the user.

   Insert a new notebook cell to install pandas and pyarrow packages:
   ```bash
   %pip install pandas pyarrow
   ```

   Finally, given the endpoint value determined above, insert the following to a new notebook cell:
   ```bash
   import json
   import pyarrow.flight as fl

   # Create a Flight client
   client = fl.connect('<ENDPOINT>')

   # Prepare the request
   request = {
       "asset": "openmetadata-file.default.openmetadata.userdata",
   }

   # Send request and fetch result as a pandas DataFrame
   info = client.get_flight_info(fl.FlightDescriptor.for_command(json.dumps(request)))
   reader: fl.FlightStreamReader = client.do_get(info.endpoints[0].ticket)
   print(reader.read_pandas())
   ```
