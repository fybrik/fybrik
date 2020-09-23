ROOT_DIR := ..
include $(ROOT_DIR)/Makefile.env

.PHONY: all
all: gen-docs gen-pages

include  $(ROOT_DIR)/.mk/tools.mk

# Proto files

CONNECTORS_PROTO := $(ROOT_DIR)/pkg/connectors/protos

WEBSITE_BASEURL ?= https://ibm.github.io/the-mesh-for-data/
WEBSITE_LOCAL_DIR ?= public

.PHONY: gen-pages
gen-pages: $(TOOLBIN)/hugo
	$(ABSTOOLBIN)/hugo --gc --minify \
		--baseURL $(WEBSITE_BASEURL) \
		--cleanDestinationDir --destination $(WEBSITE_LOCAL_DIR)

.PHONY: gen-docs
gen-docs: gen-connectors-docs gen-api-docs

.PHONY: gen-connectors-docs
gen-connectors-docs: $(TOOLBIN)/protoc $(TOOLBIN)/protoc-gen-docs
	PATH=$(TOOLBIN) protoc -I=$(CONNECTORS_PROTO) \
		--docs_out=mode=html_fragment_with_front_matter:./content/docs/reference/api/generated \
		$(wildcard ./$(CONNECTORS_PROTO)/*.proto)

.PHONY: gen-api-docs
gen-api-docs: $(TOOLBIN)/crd-ref-docs
	$(TOOLBIN)/crd-ref-docs \
	 --source-path=../manager/apis/app \
	 --config=./crd-ref-docs/config.yaml \
	 --templates-dir=./crd-ref-docs/templates/asciidoctor \
	 --renderer=asciidoctor \
	 --output-path=./content/docs/reference/api/generated/app.asciidoc
	sed -i '1i\
	--- \
	title: app.m4d.ibm.com \
	weight: 10 \
	--- \
	' ./content/docs/reference/api/generated/app.asciidoc
	
	$(TOOLBIN)/crd-ref-docs \
	 --source-path=../manager/apis/motion \
	 --config=./crd-ref-docs/config.yaml \
	 --templates-dir=./crd-ref-docs/templates/asciidoctor \
	 --renderer=asciidoctor \
	 --output-path=./content/docs/reference/api/generated/motion.asciidoc
	sed -i '1i\
	--- \
	title: motion.m4d.ibm.com \
	weight: 20 \
	--- \
	' ./content/docs/reference/api/generated/motion.asciidoc
