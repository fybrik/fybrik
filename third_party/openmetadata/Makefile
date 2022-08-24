include Makefile.env

all: deploy-openmetadata

deploy-openmetadata:
	kubectl create ns fybrik-system || true
	kubectl apply -f pv1.yaml -n fybrik-system || true
	kubectl apply -f pv2.yaml -n fybrik-system || true
	kubectl create secret generic airflow-mysql-secrets --from-literal=airflow-mysql-password=airflow_pass -n fybrik-system || true
	helm install openmetadata-dependencies open-metadata/openmetadata-dependencies --values values-deps.yaml --wait --timeout 20m -n fybrik-system \
--set mysql.initdbScripts."init_openmetadata_db_scripts\.sql"=${INIT_OPENMETADATA_DB_SCRIPT} \
--set mysql.initdbScripts."init_airflow_db_scripts\.sql"=${INIT_AIRFLOW_DB_SCRIPT} || true

	kubectl create secret generic mysql-secrets --from-literal=openmetadata-mysql-password=openmetadata_password -n fybrik-system || true
	kubectl create secret generic airflow-secrets --from-literal=openmetadata-airflow-password=admin -n fybrik-system || true
	helm install openmetadata open-metadata/openmetadata --wait --timeout 20m -n fybrik-system \
--set global.airflow.host=http://openmetadata-dependencies-web.fybrik-system.svc.cluster.local:8080 \
--set global.airflow.openmetadata.serverHostApiUrl="http://openmetadata.fybrik-system.svc.cluster.local:8585/api"|| true
