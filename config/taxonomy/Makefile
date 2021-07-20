ROOT_DIR := ../..
all: clean compile gen

compile-taxonomy-schema:
	go run $(ROOT_DIR)/main.go taxonomy compile --base ./base/base.yaml --out ./objects/taxonomy.json \
		--codegen  

validate:
	go run ../../main.go taxonomy validate ./objects/policy_manager_request.json
	go run ../../main.go taxonomy validate ./objects/policy_manager_response.json