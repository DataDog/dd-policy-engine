# Detect operating system
ifeq ($(OS),Windows_NT)
	DETECTED_OS := Windows
else
	DETECTED_OS := $(shell uname)
endif

LANGUAGES := c go
SHELL := /bin/bash
SCHEMA_DIR := ./fbs-schema
JSONSCHEMA_OUT_PATH := $(SCHEMA_DIR)/json/
COMMIT_SHA ?= $(shell git rev-parse --short HEAD)
DOCKER_REPO ?= datadog/docker-library
BUILD_CONTAINER ?= true

ifneq ($(DETECTED_OS),Windows)
include Makefile.docker
else
# On Windows, set DOCKER_RUN to empty so we can use the same commands as Linux
DOCKER_RUN :=
endif

.PHONY: all build build-container test examples clean $(LANGUAGES)

%-all: 
	$(DOCKER_RUN) make -C $(patsubst %-all,%,$@) all
all: build-container build-dd-compile-policy $(addsuffix -all,$(LANGUAGES)) examples

build-dd-compile-policy:
	$(DOCKER_RUN) make -C dd-compile-policy build

clean-dd-compile-policy:
	$(DOCKER_RUN) make -C dd-compile-policy clean

fmt-dd-compile-policy:
	$(DOCKER_RUN) make -C dd-compile-policy fmt

fmt-dd-compile-policy-check:
	$(DOCKER_RUN) make -C dd-compile-policy fmt-check

%-examples:
	$(DOCKER_RUN) make -C $(patsubst %-examples,%,$@) examples

examples: build-container $(addsuffix -examples,$(LANGUAGES))

%-clean: build-container
	$(DOCKER_RUN) make -C $(patsubst %-clean,%,$@) clean

%-fmt: build-container
	$(DOCKER_RUN) make -C $(patsubst %-fmt,%,$@) fmt

%-fmt-check: build-container
	$(DOCKER_RUN) make -C $(patsubst %-fmt-check,%,$@) fmt-check

clean: $(addsuffix -clean,$(LANGUAGES)) clean-dd-compile-policy

fmt: $(addsuffix -fmt,$(LANGUAGES)) fmt-dd-compile-policy

fmt-check: $(addsuffix -fmt-check,$(LANGUAGES)) fmt-dd-compile-policy-check

generate-jsonschema: build-container
	@$(DOCKER_RUN) echo "[i] generating JSON schema in: $(JSONSCHEMA_OUT_PATH)"
	@$(DOCKER_RUN) flatc --jsonschema -o $(JSONSCHEMA_OUT_PATH) $(SCHEMA_DIR)/policy.fbs

shell:
	DOCKER_EXTRA_OPTS=-i; $(DOCKER_RUN_I) sh

convert-to-json:
	@$(DOCKER_RUN) flatc --json --raw-binary $(SCHEMA_DIR)/policy.fbs -- $(FILE)

build-container:
ifeq ($(IN_DOCKER),false)
ifeq ($(BUILD_CONTAINER),true)
	docker build -t $(DOCKER_IMAGE) -f ./Dockerfile.build .
endif
endif

push-ci-container:
ifeq ($(IN_DOCKER),false)
	docker build --progress=plain --platform linux/amd64 -t $(DOCKER_REPO):dd-policy-engine-$(COMMIT_SHA)-amd64 -f ./Dockerfile.build .
	docker push $(DOCKER_REPO):dd-policy-engine-$(COMMIT_SHA)-amd64
	docker build --progress=plain --platform linux/arm64 -t $(DOCKER_REPO):dd-policy-engine-$(COMMIT_SHA)-arm64 -f ./Dockerfile.build .
	docker push $(DOCKER_REPO):dd-policy-engine-$(COMMIT_SHA)-arm64
	docker buildx imagetools create -t $(DOCKER_REPO):dd-policy-engine-$(COMMIT_SHA) \
			$(DOCKER_REPO):dd-policy-engine-$(COMMIT_SHA)-amd64 \
			$(DOCKER_REPO):dd-policy-engine-$(COMMIT_SHA)-arm64
endif
