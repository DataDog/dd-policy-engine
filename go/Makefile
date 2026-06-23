PREFIX ?= $(CURDIR)

POLICIES_ROOT := $(PREFIX)/..

SCHEMA_DIR := $(POLICIES_ROOT)/fbs-schema
SCHEMA_OUTPUT_DIR := $(PREFIX)/schema
EXAMPLES := examples/example_writer examples/example_generate_c_header_buffer examples/example_json_to_hardcoded_injector_policies

# C engine (libpolicies) locations used by the cgo cross-engine conformance test.
C_ROOT := $(POLICIES_ROOT)/c
C_SCHEMA_OUTPUT_DIR := $(C_ROOT)/src/schema
LIBPOLICIES := $(C_ROOT)/build/c/libpolicies.a

.PHONY=$(EXAMPLES) clean conformance-cross

all: build

generate-schema-headers:
	@echo "[i] Generating Go schema..."
	@flatc -g -o $(SCHEMA_OUTPUT_DIR) $(SCHEMA_DIR)/*.fbs

$(EXAMPLES): generate-schema-headers
	@echo "[i] Running $@..."
	@cd $@ && go run ./main.go

examples: $(EXAMPLES)

clean:
	@echo "[i] Cleaning up GO artifacts"
	-find $(SCHEMA_OUTPUT_DIR)/dd/wls -type f ! -name 'json.go' -delete
	@rm -rf $(addsuffix /out,$(EXAMPLES))

build: generate-schema-headers generate-antlr-grammar
	go mod tidy
	CGO_ENABLED=0 go build -ldflags="-s -w" -o dd-rules-converter ./cmd/dd-rules-converter
	CGO_ENABLED=0 go build -ldflags="-s -w" -o dd-requirements-converter ./cmd/dd-requirements-converter

test:
	go test ./internal/...
	go test ./cmd/...

# Cross-engine conformance: runs testdata/conformance/vectors.json through both
# the Go engine and the real C engine (libpolicies, via cgo) in one process and
# checks they agree. Requires the C toolchain + flatcc (the build container).
conformance-cross:
	@echo "[i] Generating C schema headers..."
	@flatcc --reader --verifier --builder --common_reader -o $(C_SCHEMA_OUTPUT_DIR) $(SCHEMA_DIR)/*.fbs
	@echo "[i] Building libpolicies.a..."
	@cd $(C_ROOT) && cmake -B build --preset gcc-dev .. && cmake --build build --target dd-policy-engine.static-lib -j
	@echo "[i] Running cross-engine conformance test..."
	@CGO_ENABLED=1 \
		CGO_CFLAGS="-I$(C_ROOT)/include -I$(C_ROOT)/src -I$(C_SCHEMA_OUTPUT_DIR)" \
		CGO_LDFLAGS="$(LIBPOLICIES) -lflatccrt" \
		go test -tags conformance_cgo ./policies/conformance/ -run TestConformanceCrossEngine -v

# Verify by unmarshalling compiled policy to JSON
verify_json:
ifndef BIN
	$(error You must specify BIN=<path_to_bin_file>)
endif
	flatc --json --defaults-json --strict-json --raw-binary -o . $(SCHEMA_DIR)/policy.fbs -- $(BIN)
	@echo "Verification JSON written to $(BIN:.bin=.json)"

fmt-check:
	go vet ./... || true

fmt:
	go fmt ./...

generate-antlr-grammar:
	antlr4 -Dlanguage=Go -package parser -visitor ./internal/parser/QueryLexer.g4
	antlr4 -Dlanguage=Go -package parser -visitor ./internal/parser/QueryParser.g4

