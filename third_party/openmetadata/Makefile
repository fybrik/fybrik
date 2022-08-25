include Makefile.env

all: deploy-openmetadata

deploy-openmetadata:
	kubectl create ns ${INSTALLATION_NAMESPACE} || true
	sed s/INSTALLATION_NAMESPACE/${INSTALLATION_NAMESPACE}/g pv1.yaml | kubectl apply -n ${INSTALLATION_NAMESPACE} -f - || true
	sed s/INSTALLATION_NAMESPACE/${INSTALLATION_NAMESPACE}/g pv2.yaml | kubectl apply -n ${INSTALLATION_NAMESPACE} -f - || true
	kubectl create secret generic airflow-mysql-secrets --from-literal=airflow-mysql-password=${AIRFLOW_MYSQL_PASSWORD} -n ${INSTALLATION_NAMESPACE} || true
	helm install openmetadata-dependencies open-metadata/openmetadata-dependencies --values values-deps.yaml --wait --timeout 20m -n ${INSTALLATION_NAMESPACE} \
--set mysql.initdbScripts."init_openmetadata_db_scripts\.sql"=${INIT_OPENMETADATA_DB_SCRIPT} \
--set mysql.initdbScripts."init_airflow_db_scripts\.sql"=${INIT_AIRFLOW_DB_SCRIPT} \
--set airflow.airflow.users[0].username=${OPENMETADATA_AIRFLOW_USER} \
--set airflow.airflow.users[0].password=${OPENMETADATA_AIRFLOW_PASSWORD} \
--set airflow.airflow.users[0].firstName=${AIRFLOW_ADMIN_FIRST_NAME} \
--set airflow.airflow.users[0].lastName=${AIRFLOW_ADMIN_LAST_NAME} \
--set airflow.airflow.users[0].email=${AIRFLOW_ADMIN_EMAIL} \
--set airflow.externalDatabase.user=${AIRFLOW_MYSQL_USER} \
--set airflow.airflow.config.AIRFLOW__LINEAGE__OPENMETADATA_API_ENDPOINT="${OPENMETADATA_ENDPOINT}" \
--set airflow.externalDatabase.host=${MYSQL_ENDPOINT} \
--set mysql.auth.rootPassword=${MYSQL_ROOT_PASSWORD} || true

	kubectl create secret generic mysql-secrets --from-literal=openmetadata-mysql-password=${OPENMETADATA_MYSQL_PASSWORD} -n ${INSTALLATION_NAMESPACE} || true
	kubectl create secret generic airflow-secrets --from-literal=openmetadata-airflow-password=${OPENMETADATA_AIRFLOW_PASSWORD} -n ${INSTALLATION_NAMESPACE} || true
	helm install openmetadata open-metadata/openmetadata --wait --timeout 20m -n ${INSTALLATION_NAMESPACE} \
--set global.airflow.host=${AIRFLOW_ENDPOINT} \
--set global.airflow.openmetadata.serverHostApiUrl=${OPENMETADATA_ENDPOINT} \
--set global.database.auth.username=${OPENMETADATA_MYSQL_USER} \
--set global.airflow.auth.username=${OPENMETADATA_AIRFLOW_USER} || true
