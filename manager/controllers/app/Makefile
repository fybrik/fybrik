ROOT_DIR := ../../..
include $(ROOT_DIR)/Makefile.env

.PHONY: test
test: *.go
	go test $(TEST_OPTIONS)

DEBUG := ./debug.out

.PHONY: debug
debug: $(TOOLBIN)/dlv $(DEBUG)
	$(ABSTOOLBIN)/dlv exec $(DEBUG)

$(DEBUG): *.go
	go test $(TEST_OPTIONS) -c -o $(DEBUG)
