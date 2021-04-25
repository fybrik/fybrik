ROOT_DIR := ../..
DOCKER_NAME = egr-connector

include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/docker.mk

.PHONY: build
build:
	mvn -B clean install -DskipTests

docker-build: build
	docker build . -t ${IMG}

.PHONY: test
test:
	mvn -B clean test

.PHONY: clean
clean:
	mvn clean && rm -f output-egeriaconnector.log

.PHONY: run
run:
	$(info EGERIA_SERVER_URL is $(EGERIA_SERVER_URL))
	nohup java -cp target/egeriaconnector-1.0-SNAPSHOT-jar-with-dependencies.jar:src/lib/* com.ibm.egeria.EgeriaConnector > output-egeriaconnector.log &  2>&1 ; echo "$$!" >> pids.txt; sleep 5

.PHONY: terminate
terminate:
	kill -9 $$( cat pids.txt ); rm -f pids.txt

-include ${ROOT_DIR}/.env
.EXPORT_ALL_VARIABLES:
