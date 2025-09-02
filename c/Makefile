PREFIX ?= $(CURDIR)
POLICIES_ROOT := $(PREFIX)/..

BUILD_DIR ?= build
TEST_DIR := $(BUILD_DIR)/c/src/test
COVERAGE_DIR := $(PREFIX)/coverage

TEST_BIN := $(TEST_DIR)/dd-policy-engine.unit-tests

EXAMPLES := basic-reader hardcoded-header-reader mmap-reader hardcoded_cmdline_evaluator
FILES_TO_FORMAT += $(shell find . -type f \( -name "*.c" -o -name "*.h" \) )

SCHEMA_DIR := $(POLICIES_ROOT)/fbs-schema
SCHEMA_OUTPUT_DIR := $(PREFIX)/src/schema

# Run `cmake --list-presets` to get the list of all available presets.
CMAKE_PRESET := clang-dev

.PHONY: all clean fmt $(SCHEMA_OUTPUT_DIR) lib/libpolicies.a $(EXAMPLES) coverage

all: clean go-prerequisites configure
	@cmake --build $(BUILD_DIR) -j

$(SCHEMA_OUTPUT_DIR): $(SCHEMA_DIR)/*.fbs
	@echo "[i] Generating C schema headers..."
	@mkdir -p $(SCHEMA_OUTPUT_DIR)
	@flatcc --reader --verifier --builder --common_reader -o $(SCHEMA_OUTPUT_DIR) $(SCHEMA_DIR)/*.fbs

go-prerequisites:
	@echo "[i] Checking prerequisites..."
	@make -C ../go/ examples
	@command -v clang-format >/dev/null 2>&1 || { echo >&2 "clang-format is required but not installed. Aborting."; exit 1; }

configure:
	@echo "[i] Configuring CMake project..."
	@cmake -B $(BUILD_DIR) --preset $(CMAKE_PRESET) ..

examples: go-prerequisites configure
	@echo "[i] Building examples..."
	@cmake --build $(BUILD_DIR) --target $(EXAMPLES)

clean:
	@echo "[i] Cleaning up..."
	@rm -rf $(BUILD_DIR)
	@rm -rf $(COVERAGE_DIR)
	@echo "[i] Done cleaning"

lib/libpolicies.a: configure
	@echo "[i] Building libpolicies.a"
	@cmake --build $(BUILD_DIR) --target dd-policy-engine.static-lib

fmt:
	@echo "[i] Formatting source files..."
	@clang-format -i $(FILES_TO_FORMAT)

fmt-check:
	@echo -e "[i] Files to format:\n$(FILES_TO_FORMAT)"
	@clang-format --dry-run --Werror $(FILES_TO_FORMAT)

debugger: configure
	@echo "[i] Building debugger binary..."
	@cmake --build $(BUILD_DIR) --target dd-policy-engine.debugger
	@echo "[i] Debugger binary built: $(BUILD_DIR)/c/basic-debugger"

# Unit tests (utest.h)
# Build all test translation units in src/test and link with object files.
test: test-build
	@echo "[i] Running unit tests..."
	@ctest --test-dir $(BUILD_DIR) -V

test-build: configure
	@echo "[i] Building unit tests..."
	@cmake --build $(BUILD_DIR) --target dd-policy-engine.unit-tests

# Simple coverage target
coverage: CMAKE_PRESET = clang-coverage
coverage: clean configure test-build
	@echo "[i] Running tests with coverage..."
	@mkdir $(COVERAGE_DIR)
	@$(TEST_BIN)
	@echo "[i] Coverage report:"
	@lcov --quiet --capture --directory $(BUILD_DIR) --output-file $(COVERAGE_DIR)/raw.info
	@lcov --quiet --remove $(COVERAGE_DIR)/raw.info -o $(COVERAGE_DIR)/filtered.info --ignore-errors unused '/usr/*' '/opt/*' 'src/test/*' 'build/*'
	@genhtml --quiet --dark-mode $(COVERAGE_DIR)/filtered.info --output-directory $(COVERAGE_DIR)/report
	@echo "[i] HTML Report: $(COVERAGE_DIR)/report/index.html"
