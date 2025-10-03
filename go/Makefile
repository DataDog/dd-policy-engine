PREFIX ?= $(CURDIR)

POLICIES_ROOT := $(PREFIX)/..

SCHEMA_DIR := $(POLICIES_ROOT)/fbs-schema
SCHEMA_OUTPUT_DIR := $(PREFIX)/schema
EXAMPLES := examples/example_writer examples/example_generate_c_header_buffer examples/example_json_to_hardcoded_injector_policies

.PHONY=$(EXAMPLES) clean

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
	CGO_ENABLED=0 go build -ldflags="-s -w" ./cmd/... 

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

