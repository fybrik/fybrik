ROOT_DIR := ../../..
include $(ROOT_DIR)/Makefile.env

.PHONY: test
test: *.go
	go test -v
	export CSP_PATH=$(TOOLBIN)/fzn-or-tools
	go test -v

DEBUG := ./debug.out

.PHONY: debug
debug: $(TOOLBIN)/dlv $(DEBUG)
	$(ABSTOOLBIN)/dlv exec $(DEBUG)

$(DEBUG): *.go
	go test -c -o $(DEBUG)
