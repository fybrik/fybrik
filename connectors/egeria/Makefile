ROOT_DIR := ../..
include $(ROOT_DIR)/Makefile.env
include ${ROOT_DIR}/pkg/policy-compiler/policy-compiler.env

DOCKER_NAME = egr-connector
include $(ROOT_DIR)/hack/make-rules/docker.mk

.PHONY: build
build:
	mvn clean package -DskipTests

docker-build: build

.PHONY: test
test:
	mvn clean test

.PHONY: clean
clean:
	mvn clean && rm -f output-egeriaconnector.log

.PHONY: run
run:
	nohup java -cp target/egeriaconnector-1.0-SNAPSHOT-jar-with-dependencies.jar:src/lib/* com.ibm.egeria.EgeriaConnector > output-egeriaconnector.log &  2>&1 ; echo "$$!" >> pids.txt; sleep 5

.PHONY: terminate
terminate:
	kill -9 $$( cat pids.txt ); rm -f pids.txt

IMAGE := egr-connector

kustomization.yaml: kustomization.yaml.in
	cp $< $@

.PHONY: deploy
deploy: kustomization.yaml
	kustomize edit set image $(IMAGE)=${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/$(IMAGE):${DOCKER_TAGNAME}
	kustomize build . | kubectl apply -f -

.PHONY: undeploy
undeploy: kustomization.yaml
	kustomize edit set image $(IMAGE)=${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/$(IMAGE):${DOCKER_TAGNAME}
	kustomize build . | kubectl delete -f -

.PHONY: charts
charts: kustomization.yaml
	cp egr_connector.yaml ${ROOT_DIR}/charts/m4d/templates/.
