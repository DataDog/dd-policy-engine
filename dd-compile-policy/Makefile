# Include headers and source files for formatting
FILES_TO_FORMAT = $(wildcard ./*/*.cc ./*/*.cpp ./*/*.hpp ./*/*.c ./*/*.h)

# Get the directory of the current Makefile
MAKEFILE_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
SCHEMA_PATH = $(MAKEFILE_DIR)/../fbs-schema/policy.fbs
POLICY_BFBS = schema/policy.bfbs
POLICY_BFBS_HEADER = src/policy_schema_bfbs.h

STRIP = strip

CMAKE_PRESET := clang-dev
BUILD_DIR = build
OUT = $(BUILD_DIR)/dd-compile-policy/dd-compile-policy

.PHONY: all clean strip_binary

all: build

build: dd-compile-policy strip_binary

dd-compile-policy: $(POLICY_BFBS_HEADER)
	cmake -B $(BUILD_DIR) --preset $(CMAKE_PRESET) ..
	cmake --build $(BUILD_DIR) --target dd-policy-engine.compile-policy -j

# Generate Flatbuffers binary schema (bfbs)
$(POLICY_BFBS): $(SCHEMA_PATH)
	flatc --binary --force-defaults --schema -o schema $(SCHEMA_PATH)

# Generate Flatbuffers header file from binary schema
$(POLICY_BFBS_HEADER): $(POLICY_BFBS)
	xxd -i $< > $@

strip_binary:
ifeq (, $(shell which $(STRIP)))
	@echo "strip not found, skipping binary strip"
else
	$(STRIP) $(OUT)
endif

# Verify by unmarshalling compiled policy to JSON
verify_json:
ifndef BIN
	$(error You must specify BIN=<path_to_bin_file>)
endif
	flatc --json --strict-json --raw-binary -o . $(SCHEMA_PATH) -- $(BIN)
	@echo "Verification JSON written to $(BIN:.bin=.json)"

clean:
	rm -rf $(BUILD_DIR)

fmt:
	@echo "[i] Formatting source files..."
	@clang-format -i $(FILES_TO_FORMAT)

fmt-check:
	@echo -e "[i] Files to format:\n$(FILES_TO_FORMAT)"
	@clang-format --dry-run --Werror $(FILES_TO_FORMAT)
