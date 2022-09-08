TAG := $(shell git for-each-ref --format="%(refname:short)" --sort=-authordate --count=1 refs/tags)

COMMIT := $(shell git rev-list -1 HEAD)

LDFLAGS := -ldflags "-X main.gitCommit=$(COMMIT) -X main.gitTag=$(TAG)"