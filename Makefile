LANGUAGES := c go
SHELL := /bin/bash
SCHEMA_DIR := ./fbs-schema
JSONSCHEMA_OUT_PATH := $(SCHEMA_DIR)/json/

include Makefile.docker

.PHONY: all build build-container test examples clean $(LANGUAGES)

%-all: 
	$(DOCKER_RUN) make -C $(patsubst %-all,%,$@) all
all: build-container $(addsuffix -all,$(LANGUAGES)) examples


%-examples:
	$(DOCKER_RUN) make -C $(patsubst %-examples,%,$@) examples

examples: build-container $(addsuffix -examples,$(LANGUAGES))

%-clean: build-container
	$(DOCKER_RUN) make -C $(patsubst %-clean,%,$@) clean

clean: $(addsuffix -clean,$(LANGUAGES))

generate-jsonschema: build-container
	@$(DOCKER_RUN) echo "[i] generating JSON schema in: $(JSONSCHEMA_OUT_PATH)"
	@$(DOCKER_RUN) flatc --jsonschema -o $(JSONSCHEMA_OUT_PATH) $(SCHEMA_DIR)/policy.fbs

shell:
	DOCKER_EXTRA_OPTS=-i; $(DOCKER_RUN_I) sh

convert-to-json:
	@$(DOCKER_RUN) flatc --json --raw-binary $(SCHEMA_DIR)/policy.fbs -- $(FILE)

build-container:
ifeq ($(IN_DOCKER),false)
	docker build -t $(DOCKER_IMAGE) -f ./Dockerfile.build .
endif
