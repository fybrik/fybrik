ROOT_DIR := ../..
include $(ROOT_DIR)/Makefile.env

CHART_NAME=mychart
VERSION=0.7.0
CHART_FILE_NAME=$(CHART_NAME)-$(VERSION).tgz
DATA_DIR?=/tmp

SOURCE := \
	helm.go \
	helm_test.go

TARGET := main

.PHONY: test
test: $(ABSTOOLBIN)/helm
	$(eval TMP1 := $(shell mktemp -d -p $(DATA_DIR)))
	mkdir $(TMP1)/charts
	$(eval TMP := $(TMP1)/charts)
	cd $(TMP) && $(ABSTOOLBIN)/helm create $(CHART_NAME)
	cd $(TMP) && $(ABSTOOLBIN)/helm package $(CHART_NAME) -d $(TMP)  --version $(VERSION)
	cd $(TMP) && $(ABSTOOLBIN)/helm push $(TMP)/$(CHART_FILE_NAME) oci://localhost:5000/fybrik-system/
	TMP_CHART=$(TMP)/$(CHART_NAME) go test -v $(SOURCE)
	rm -rf $(TMP)/$(CHART_FILE_NAME)

.PHONY: build
build:
	go build -gcflags '-N -l' -o $(TARGET) $(SOURCE)

.PHONY: debug
debug:
	dlv exec  ./$(TARGET)
