.PHONY: pc-test
pc-test:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out

.PHONY: test
test: pc-test