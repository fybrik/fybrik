ROOT_DIR := ../..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/tools.mk
 
.PHONY: build
build: $(TOOLBIN)/protoc $(TOOLBIN)/protoc-gen-go $(TOOLBIN)/protoc-gen-go-grpc
	rm -f protobuf/*
	PATH=$(ABSTOOLBIN) protoc -I protos/ \
		--go_out=$(ROOT_DIR)/pkg/connectors/protobuf --go_opt=paths=source_relative \
    	--go-grpc_out=$(ROOT_DIR)/pkg/connectors/protobuf --go-grpc_opt=paths=source_relative \
    	protos/*.proto
