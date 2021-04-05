ROOT_DIR := ../..
include $(ROOT_DIR)/Makefile.env

SOURCE := \
	helm.go \
	helm_test.go

TARGET := main

.PHONY: test
test:
	go test -v $(SOURCE)

.PHONY: build
build:
	go build -gcflags '-N -l' -o $(TARGET) $(SOURCE)

.PHONY: debug
debug:
	dlv exec  ./$(TARGET)
