define license_go
	echo && cd $1 && \
	GO111MODULE=on go mod tidy && \
	GO111MODULE=on go mod vendor && \
	GO111MODULE=on go mod verify && \
	($(ABSTOOLBIN)/license_finder || true)
endef

define license_java
	echo && cd $1 && \
	($(ABSTOOLBIN)/license_finder || true)
endef

GO_VERSION:=1.16
CODE_MAINT += go-version
.PHONY: go-version
go-version:
	@(go version | grep -q 'go$(GO_VERSION)\(\.[0-9]*\)\? ') || \
	echo 'WARNING: bad go version to fix run: eval "$$(gimme $(GO_VERSION))"'

CODE_MAINT += fmt
.PHONY: fmt
fmt:
	go fmt ./...

CODE_MAINT += vet
.PHONY: vet
vet:
	go vet ./...

CODE_MAINT += fix
.PHONY: fix
fix:
	go fix ./...

CODE_MAINT += tidy
.PHONY: tidy
tidy:
	go mod tidy

.PHONY: revive
revive: $(TOOLBIN)/revive
	$(TOOLBIN)/revive -config lint-rules.toml -formatter stylish ./...

.PHONY: lint
lint: $(TOOLBIN)/golangci-lint
	$(TOOLBIN)/golangci-lint run

.PHONY: lint-fix
lint-fix: $(TOOLBIN)/golangci-lint
	$(TOOLBIN)/golangci-lint run --fix

.PHONY: lint-updated
lint-updated: $(TOOLBIN)/golangci-lint
	$(TOOLBIN)/golangci-lint run --new-from-rev=HEAD~

.PHONY: lint-godox
lint-godox: $(TOOLBIN)/golangci-lint
	$(TOOLBIN)/golangci-lint run --enable=godox

.PHONY: misspell
misspell: $(TOOLBIN)/misspell
	$(TOOLBIN)/misspell --error ./**

.PHONY: misspell-fix
misspell-fix: $(TOOLBIN)/misspell
	$(TOOLBIN)/misspell -w ./**

.PHONY: goimports-fix
goimports-fix-file: $(TOOLBIN)/goimports
	$(TOOLBIN)/goimports -local fybrik.io/fybrik -w

CODE_MAINT += protos-lint
.PHONY: protos-lint
protos-lint: $(TOOLBIN)/protoc $(TOOLBIN)/protoc-gen-lint
	@for i in $$(find . -name protos -type d); do \
		echo "protoc-gen-lint on $$i/*.proto"; \
		PATH=$(ABSTOOLBIN) protoc -I $$i/ $$i/*.proto --lint_out=sort_imports:$$i; \
	done

.PHONY: verify
verify: $(CODE_MAINT)

.PHONY: verify-lint
verify-lint: $(CODE_MAINT) lint-updated