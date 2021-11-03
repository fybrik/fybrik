ROOT_DIR := ../..
include $(ROOT_DIR)/Makefile.env
export HELM_EXPERIMENTAL_OCI=1

SOURCE := \
	reference.go \
	helm.go \
	helm_test.go

TARGET := main

.PHONY: test
test: $(ABSTOOLBIN)/helm
	$(eval TMP := $(shell mktemp -d))
	cd $(TMP) && $(ABSTOOLBIN)/helm create mychart
	cd $(TMP) && $(ABSTOOLBIN)/helm package mychart -d $(TMP)  --version 0.7.0
	cd $(TMP) && $(ABSTOOLBIN)/helm push $(TMP)/mychart-0.7.0.tgz oci://localhost:5000/fybrik-system/
	rm -rf $(TMP)/mychart-0.7.0.tgz
	TMP_CHART=$(TMP)/mychart go test -v $(SOURCE)

.PHONY: build
build:
	go build -gcflags '-N -l' -o $(TARGET) $(SOURCE)

.PHONY: debug
debug:
	dlv exec  ./$(TARGET)
