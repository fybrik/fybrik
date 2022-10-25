# Discovering an S3 asset through the OpenMetadata UI

This page explains how to discover an existing S3 asset through the OpenMetadata UI. The screenshots refer to `localstack` cloud storage, but the explanations also apply to other S3 services.

Begin by opening your browser to the OpenMetadata UI. If you installed OpenMetadata in your kubernetes cluster in the `open-metadata` namespace, go to http://localhost:8585 after running:
```
kubectl port-forward svc/openmetadata -n open-metadata 8585:8585 &
```

To create a connection to S3 and discovering your CSV asset:

1. Login to OpenMetadata. The default username and password are: `admin` / `admin`
   <br><center><img src="../../static/openmetadata/01_login.jpg" width=600></center>
1. On the left menu, choose `Services`
   <br><center><img src="../../static/openmetadata/02_menu.jpg" width=600></center>
1. Press `Add new Database Service`
   <br><center><img src="../../static/openmetadata/03_add_database_service.jpg" width=600></center>
1. Choose `Datalake` and press `Next`
   <br><center><img src="../../static/openmetadata/04_datalake.jpg" width=600></center>
1. Enter a name for your service, such as `openmetadata-s3`, and press `Next`
   <br><center><img src="../../static/openmetadata/05_service_name.jpg" width=600></center>
1. Enter the connection information. That information includes the `Access Key` and `Secret Key`. The `AWS Region` is mandatory, but it is ignored if you enter an `Endpoint URL`. If your object storage is a local `localstack` deployment, enter its URL (e.g. `http://localstack.fybrik-notebook-sample:4566`). Optionally, you may enter a `Bucket Name`, thereby limiting the discovery process to a single bucket
   <br><center><img src="../../static/openmetadata/06_service_config.jpg" width=600></center>
1. Scroll down and press `Test Connection` to make sure that the credentials you provided are correct. Once you see that the `Connection test was successful`, press `Save`
   <br><center><img src="../../static/openmetadata/07_test_connection.jpg" width=600></center>
1. Choose `Add ingestion`
   <br><center><img src="../../static/openmetadata/08_add_ingestion.jpg" width=600></center>
1. You need not change the ingestion configuration. Press `Next`
   <br><center><img src="../../static/openmetadata/09_add_ingestion_2.jpg" width=600></center>
1. Press `Next`
   <br><center><img src="../../static/openmetadata/10_add_ingestion_3.jpg" width=600></center>
1. Press `Add & Deploy`
   <br><center><img src="../../static/openmetadata/11_add_and_deploy.jpg" width=600></center>
1. The Ingestion Pipeline is created. Press `View Service`
   <br><center><img src="../../static/openmetadata/12_pipeline_created.jpg" width=600></center>
1. Choose the `Ingestions` tab
   <br><center><img src="../../static/openmetadata/13_new_service.jpg" width=600></center>
1. The status of the Ingestion Pipeline might be `Queued`...
   <br><center><img src="../../static/openmetadata/14_ingestion_queued.jpg" width=600></center>
1. ... or `Running`.
   <br><center><img src="../../static/openmetadata/15_ingestion_running.jpg" width=600></center>
1. Wait until the ingestion process has completed successfully, and press the `Explore` tab
   <br><center><img src="../../static/openmetadata/16_ingestion_completed.jpg" width=600></center>
1. Given a list of all OpenMetadata tables, press the table in which you are interested
   <br><center><img src="../../static/openmetadata/17_discovered_table.jpg" width=600></center>
1. You can learn the name that OpenMetadata gave your table by looking at the URL. If, for instance, the URL is `localhost:8585/table/openmetadata-s3.default.demo."PS_20174392719_1491204439457_log.csv"`, then your `assetID` is `openmetadata-s3.default.demo."PS_20174392719_1491204439457_log.csv"`.
   To add tags, press `Add tag`
   <br><center><img src="../../static/openmetadata/18_table_name.jpg" width=600></center>
1. Choose a tag for the dataset, such as `Fybrik.finance`
   <br><center><img src="../../static/openmetadata/19_choose_tag.jpg" width=600></center>
1. Press the check mark
   <br><center><img src="../../static/openmetadata/20_add_tag.jpg" width=600></center>
1. Next, you can add tags to some of the columns
   <br><center><img src="../../static/openmetadata/21_choose_column_tag.jpg" width=600></center>
1. For instance, you may choose `Fybrik.PII` for columns that need to be redacted
   <br><center><img src="../../static/openmetadata/22_add_column_tag.jpg" width=600></center>
1. Finally, press the `Custom Properties` tab
   <br><center><img src="../../static/openmetadata/23_select_custom_properties.jpg" width=600></center>
1. Set the asset properties as needed. For instance, set `connectionType` to `s3`. If the discovered asset is a csv object, set `dataFormat` to `csv`
   <br><center><img src="../../static/openmetadata/24_table_properties.jpg" width=600></center>
You are all set. OpenMetadata has discovered your asset, and you have added tags and metadata values. You can reference this asset using the `asset ID`
