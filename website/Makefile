ROOT_DIR := ..
include $(ROOT_DIR)/Makefile.env

.PHONY: all
all: gen-docs gen-pages

include  $(ROOT_DIR)/hack/make-rules/tools.mk

# Proto files

CONNECTORS_PROTO := $(ROOT_DIR)/pkg/connectors/protos

WEBSITE_BASEURL ?= https://ibm.github.io/the-mesh-for-data/
WEBSITE_LOCAL_DIR ?= public

.PHONY: server
server: $(TOOLBIN)/hugo
	$(ABSTOOLBIN)/hugo server --gc --minify \
		--cleanDestinationDir --destination $(WEBSITE_LOCAL_DIR)

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
	./crd-ref-docs/frontmatter.sh ./content/docs/reference/api/generated/app.asciidoc app.m4d.ibm.com 10
	$(TOOLBIN)/crd-ref-docs \
	 --source-path=../manager/apis/motion \
	 --config=./crd-ref-docs/config.yaml \
	 --templates-dir=./crd-ref-docs/templates/asciidoctor \
	 --renderer=asciidoctor \
	 --output-path=./content/docs/reference/api/generated/motion.asciidoc
	./crd-ref-docs/frontmatter.sh ./content/docs/reference/api/generated/motion.asciidoc motion.m4d.ibm.com 20
