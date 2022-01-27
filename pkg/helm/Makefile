ROOT_DIR := ../..
include $(ROOT_DIR)/Makefile.env

# To enable OCI experimental support for Helm versions prior to v3.8.0, HELM_EXPERIMENTAL_OCI is set
export HELM_EXPERIMENTAL_OCI=1

CHART_NAME=mychart
VERSION=0.7.0
CHART_FILE_NAME=$(CHART_NAME)-$(VERSION).tgz

SOURCE := \
	helm.go \
	helm_test.go

TARGET := main

.PHONY: test
test: $(ABSTOOLBIN)/helm
	$(eval TMP := $(shell mktemp -d))
	cd $(TMP) && $(ABSTOOLBIN)/helm create $(CHART_NAME)
	cd $(TMP) && $(ABSTOOLBIN)/helm package $(CHART_NAME) -d $(TMP)  --version $(VERSION)
	cd $(TMP) && $(ABSTOOLBIN)/helm push $(TMP)/$(CHART_FILE_NAME) oci://localhost:5000/fybrik-system/
	rm -rf $(TMP)/$(CHART_FILE_NAME)
	TMP_CHART=$(TMP)/$(CHART_NAME) go test -v $(SOURCE)

.PHONY: build
build:
	go build -gcflags '-N -l' -o $(TARGET) $(SOURCE)

.PHONY: debug
debug:
	dlv exec  ./$(TARGET)
