include Makefile.env

all: deploy-openmetadata

deploy-openmetadata:
	kubectl create ns ${INSTALLATION_NAMESPACE} || true
	sed s/INSTALLATION_NAMESPACE/${INSTALLATION_NAMESPACE}/g pv1.yaml | kubectl apply -n ${INSTALLATION_NAMESPACE} -f - || true
	sed s/INSTALLATION_NAMESPACE/${INSTALLATION_NAMESPACE}/g pv2.yaml | kubectl apply -n ${INSTALLATION_NAMESPACE} -f - || true
	kubectl create secret generic airflow-mysql-secrets --from-literal=airflow-mysql-password=airflow_pass -n ${INSTALLATION_NAMESPACE} || true
	helm install openmetadata-dependencies open-metadata/openmetadata-dependencies --values values-deps.yaml --wait --timeout 20m -n ${INSTALLATION_NAMESPACE} \
--set mysql.initdbScripts."init_openmetadata_db_scripts\.sql"=${INIT_OPENMETADATA_DB_SCRIPT} \
--set mysql.initdbScripts."init_airflow_db_scripts\.sql"=${INIT_AIRFLOW_DB_SCRIPT} || true

	kubectl create secret generic mysql-secrets --from-literal=openmetadata-mysql-password=openmetadata_password -n ${INSTALLATION_NAMESPACE} || true
	kubectl create secret generic airflow-secrets --from-literal=openmetadata-airflow-password=admin -n ${INSTALLATION_NAMESPACE} || true
	helm install openmetadata open-metadata/openmetadata --wait --timeout 20m -n ${INSTALLATION_NAMESPACE} \
--set global.airflow.host=http://openmetadata-dependencies-web.${INSTALLATION_NAMESPACE}.svc.cluster.local:8080 \
--set global.airflow.openmetadata.serverHostApiUrl="http://openmetadata.${INSTALLATION_NAMESPACE}.svc.cluster.local:8585/api"|| true
