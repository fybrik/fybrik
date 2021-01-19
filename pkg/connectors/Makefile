ROOT_DIR := ../..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/tools.mk
 
########### build targets ###########

.PHONY: protos-build
protos-build: $(TOOLBIN)/protoc $(TOOLBIN)/protoc-gen-go
	rm -f protobuf/*
	PATH=$(ABSTOOLBIN) protoc -I protos/ protos/*.proto --go_opt=paths=source_relative --go_out=plugins=grpc:$(ROOT_DIR)/pkg/connectors/protobuf

.PHONY: build
build: protos-build
