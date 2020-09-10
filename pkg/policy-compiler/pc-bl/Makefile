.PHONY: pc-bl-test
pc-bl-test:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out

.PHONY: test
test: pc-bl-test
