TAG := $(shell git describe --tags --abbrev=0)

COMMIT := $(shell git rev-list -1 HEAD)

LDFLAGS := -ldflags "-X main.gitCommit=$(COMMIT) -X main.gitTag=$(TAG)"