PREFIX ?= $(CURDIR)

AR      ?= ar
RANLIB  ?= ranlib
LIBFLATCCRT ?= /root/flatcc/lib/libflatccrt.a

EXAMPLES_DIR := $(PREFIX)/examples
TEST_DIR := $(PREFIX)/src/test
TEST_FILES := $(wildcard $(TEST_DIR)/*.c)
TEST_BIN := $(TEST_DIR)/policies-tests
COVERAGE_DIR := $(PREFIX)/coverage

SOURCES := $(filter-out $(TEST_FILES), $(wildcard $(PREFIX)/src/*.c))
EXAMPLES := basic-reader hardcoded-header-reader mmap-reader hardcoded_cmdline_evaluator
TEST_BIN := $(TEST_DIR)/policies-tests
FILES_TO_CLEAN += $(OBJS) $(SCHEMA_OUTPUT_DIR) lib/ $(addprefix $(EXAMPLES_DIR)/, $(EXAMPLES)) *.html *.css *.gcov $(PREFIX)/src/*.gcda $(PREFIX)/src/*.gcno $(PREFIX)/src/test/*.gcda $(PREFIX)/src/test/*.gcno $(COVERAGE_DIR) $(TEST_BIN)
FILES_TO_FORMAT += $(shell find . -type f \( -name "*.c" -o -name "*.h" \) )
LIB_SRC += $(SOURCES)
OBJS += $(SOURCES:.c=.o)

POLICIES_ROOT := $(PREFIX)/..

SCHEMA_DIR := $(POLICIES_ROOT)/fbs-schema
SCHEMA_OUTPUT_DIR := $(PREFIX)/src/schema

CFLAGS ?= \
	-std=c23 \
	-fPIC \
	-fvisibility=hidden \
	-pedantic-errors \
	-Werror -Wall -Wextra \
	-fstrict-aliasing \
	-funsigned-char \
	-fwrapv \
	-fno-stack-protector \
	-Wconversion \
	-Wcast-align \
	-Wno-sign-conversion \
	-Wmissing-format-attribute \
	-Wpointer-arith \
	-Wformat-nonliteral \
	-Winit-self \
	-Wwrite-strings \
	-Wshadow \
	-Wparentheses \
	-Wsign-conversion \
	-Wenum-conversion \
	-Wswitch-enum \
	-Wvla

# add includes anyway
CFLAGS += -I$(PREFIX)/include \
	-I$(SCHEMA_OUTPUT_DIR)

LDFLAGS += \
	-L$(PREFIX)/lib/ \
	-lpolicies

.PHONY: all clean fmt test-bin $(SCHEMA_OUTPUT_DIR) lib/libpolicies.a $(EXAMPLES) coverage

all: $(SCHEMA_OUTPUT_DIR) $(OBJS) lib/libpolicies.a debugger

%.o: %.c
	@$(CC) $(CFLAGS)  -c $< -o $@

$(SCHEMA_OUTPUT_DIR): $(SCHEMA_DIR)/*.fbs
	@echo "[i] Generating C schema headers..."
	@mkdir -p $(SCHEMA_OUTPUT_DIR)
	@flatcc --reader --verifier --builder --common_reader -o $(SCHEMA_OUTPUT_DIR) $(SCHEMA_DIR)/*.fbs

go-prerequisites:
	@echo "[i] Checking prerequisites..."
	@make -C ../go/ examples
	@command -v clang-format >/dev/null 2>&1 || { echo >&2 "clang-format is required but not installed. Aborting."; exit 1; }

examples: go-prerequisites $(SCHEMA_OUTPUT_DIR) lib/libpolicies.a
	@echo "[i] Building examples..."
	@$(MAKE) -C $(EXAMPLES_DIR)

clean:
	@echo "[i] Cleaning up..."
	@rm -rf $(FILES_TO_CLEAN)
	@$(MAKE) -C $(EXAMPLES_DIR) clean
	@echo "[i] done cleaning"

lib/libpolicies.a: $(OBJS) 
	@mkdir -p lib/.flatccrt_objs
	@rm -rf lib/.flatccrt_objs/* $@
	@echo "[i] Bundling flatcc runtime: $(LIBFLATCCRT)"
	@cd lib/.flatccrt_objs && $(AR) x $(LIBFLATCCRT)
	@$(AR) qc $@ $(OBJS) lib/.flatccrt_objs/*.o
	@$(RANLIB) $@
	@echo "[i] done linking policies library (runtime included)"


fmt:
	@echo "[i] Formatting source files..."
	@clang-format -i $(FILES_TO_FORMAT)

fmt-check:
	@echo -e "[i] Files to format:\n$(FILES_TO_FORMAT)"
	@clang-format --dry-run --Werror $(FILES_TO_FORMAT)

debugger: $(SCHEMA_OUTPUT_DIR) lib/libpolicies.a
	@mkdir -p lib
	@echo "[i] Building debugger binary..."
	@$(CC) $(CFLAGS) $(PREFIX)/src/debugger/basic-debugger.c -L$(PREFIX)/lib/ -lpolicies -o lib/basic-debugger
	@echo "[i] Debugger binary built: lib/basic-debugger"

# Unit tests (utest.h)
# Build all test translation units in src/test and link with object files.
test: $(SCHEMA_OUTPUT_DIR) $(OBJS) test-build
	@echo "[i] Running unit tests..."
	@$(TEST_BIN)

test-build: $(SCHEMA_OUTPUT_DIR) $(OBJS) lib/libpolicies.a
	@echo "[i] Building unit tests..."
	$(CC) $(CFLAGS) -I$(PREFIX)/src/test $(wildcard $(PREFIX)/src/test/*.c) -L$(PREFIX)/lib/ -lpolicies -o $(TEST_BIN)

# Simple coverage target
coverage: CFLAGS += --coverage -fprofile-arcs -ftest-coverage
coverage: LDFLAGS += -lgcov
coverage: clean $(SCHEMA_OUTPUT_DIR) $(OBJS) lib/libpolicies.a test-build
	@echo "[i] Running tests with coverage..."
	@$(TEST_BIN)
	@echo "[i] Coverage report: (GCOV_PREFIX: $(GCOV_PREFIX))"
	@gcov src/eval_ctx.c | grep -E "(File|Lines executed)" | grep -v "Creating" | head -2
	@gcov src/evaluator.c | grep -E "(File|Lines executed)" | grep -v "Creating" | head -2
	@gcov src/evaluator_default.c | grep -E "(File|Lines executed)" | grep -v "Creating" | head -2
	@gcov src/policy.c | grep -E "(File|Lines executed)" | grep -v "Creating" | head -2
	@echo ""
	@echo "Uncovered functions:"
	@gcov -f src/eval_ctx.c src/evaluator.c src/evaluator_default.c src/policy.c | grep "function.*0.00%" || echo "All functions covered"
	@gcovr -r . --exclude 'src/test/*' --exclude 'src/schema/*' --html --html-details -o coverage.html
	@mkdir -p $(COVERAGE_DIR)
	@mv *.html $(COVERAGE_DIR)
	@mv *.css $(COVERAGE_DIR)
