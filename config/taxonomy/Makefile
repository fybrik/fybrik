GLANG ?= "go"
all: clean compile gen
compile:
	go run ../../main.go taxonomy compile --base ./base/base.yaml --out ./objects/taxonomy.json ./layers/actions.yaml ./layers/locations.yaml \
		--codegen  
gen:
	mkdir -p out
	~/bin/openapitools/openapi-generator-cli generate -i ./codegen.spec.yaml -g $(GLANG) -o out \
		--additional-properties=legacyDiscriminatorBehavior=false \
		--enable-post-process-file \
		--type-mappings=object=interface{}
clean:
	rm -rf docs
	rm -rf out