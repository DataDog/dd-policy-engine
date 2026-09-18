# Detect operating system for platform-specific build logic
ifeq ($(OS),Windows_NT)
	DETECTED_OS := Windows
else
	DETECTED_OS := $(shell uname)
endif

# Default architecture for Windows
CMAKE_ARCH ?= x64

FLATBUFFERS_VERSION = 25.9.23
FLATBUFFERS_TARBALL = flatbuffers-v$(FLATBUFFERS_VERSION).tar.gz
FLATBUFFERS_DIR = flatbuffers-$(FLATBUFFERS_VERSION)
FLATBUFFERS_PATCH = ../flatbuffers.patch # relative to FLATBUFFERS_DIR
FLATBUFFERS_PATCH_FILE = flatbuffers.patch # same file, relative to this Makefile
# Records that the patch applied. Lives OUTSIDE $(FLATBUFFERS_DIR) on purpose: a
# stamp inside it would bump the directory's mtime and re-trigger itself forever.
FLATBUFFERS_PATCH_STAMP = $(FLATBUFFERS_DIR).patched

CXXOPTS_VERSION = 3.2.1
BUILD_DIR = $(FLATBUFFERS_DIR)/build
SRC = src/compile_policy.cpp
OUT = bin/dd-compile-policy
# Include headers and source files for formatting
FILES_TO_FORMAT = $(wildcard ./*/*.cpp ./*/*.hpp ./*/*.c ./*/*.h)


# Default settings (Unix/Linux/macOS)
LIBS = $(BUILD_DIR)/libflatbuffers.a
CXX = g++
CXXFLAGS = -std=c++17 -Os -flto -g0 -fno-exceptions -fno-rtti -ffunction-sections -fdata-sections -fvisibility=hidden -fvisibility-inlines-hidden -DCXXOPTS_NO_EXCEPTIONS -DCXXOPTS_NO_RTTI
LDFLAGS = 
STRIP = strip
CXX_OUTPUT_ARGS = $(LDFLAGS) -o $@
MKDIR = mkdir -p
FLATC = flatc
# Windows ships Python as `python`; `python3` is not reliably on PATH there.
PYTHON = python3

# Windows overrides
ifeq ($(DETECTED_OS),Windows)
	OUT = bin/dd-compile-policy.exe
	LIBS = $(BUILD_DIR)/MinSizeRel/flatbuffers.lib
	CXX = cl.exe
	CXXFLAGS = /std:c++17 /O1 /GL /MT /EHsc /DCXXOPTS_NO_EXCEPTIONS /DCXXOPTS_NO_RTTI  # /MT=static runtime, /GL=optimization
	LDFLAGS = /link /LTCG kernel32.lib user32.lib
	STRIP = echo "Strip not needed on Windows"
	CXX_OUTPUT_ARGS = /Fe$@ $(LDFLAGS)
	MKDIR = powershell -Command "New-Item -ItemType Directory -Force -Path"
	FLATC = $(BUILD_DIR)/MinSizeRel/flatc.exe
	PYTHON = python
endif


# Additional Linux-specific flags
ifeq ($(DETECTED_OS),Linux)
	CXXFLAGS += -static -static-libgcc -static-libstdc++
	LDFLAGS += -Wl,--gc-sections -Wl,--strip-all -Wl,--as-needed
	# Override: Linux puts LDFLAGS before output flag  
	CXX_OUTPUT_ARGS = $(LDFLAGS) -o $@
endif

# Include headers and source files for formatting
FILES_TO_FORMAT = $(wildcard ./*/*.cpp ./*/*.hpp ./*/*.c ./*/*.h)
INCLUDES = -I$(FLATBUFFERS_DIR)/include -Isrc -Iinclude
CXXOPTS_HEADER = include/cxxopts.hpp

# Get the directory of the current Makefile
MAKEFILE_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
SCHEMA_PATH = $(MAKEFILE_DIR)/../fbs-schema/policy.fbs
POLICY_BFBS = schema/policy.bfbs
POLICY_BFBS_HEADER = src/policy_schema_bfbs.h

STRIP = strip
.PHONY: all clean flatbuffers strip_binary test

all: build

build: flatbuffers $(CXXOPTS_HEADER) $(POLICY_BFBS) $(POLICY_BFBS_HEADER) $(OUT) strip_binary

# Download cxxopts header (pinned to version $(CXXOPTS_VERSION))
$(CXXOPTS_HEADER):
	$(MKDIR) include
ifeq ($(DETECTED_OS),Windows)
	powershell -Command "Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/jarro2783/cxxopts/v$(CXXOPTS_VERSION)/include/cxxopts.hpp' -OutFile '$@'"
else
	wget -O $@ "https://raw.githubusercontent.com/jarro2783/cxxopts/v$(CXXOPTS_VERSION)/include/cxxopts.hpp"
endif

# Download & extract flatbuffers
$(FLATBUFFERS_TARBALL):
ifeq ($(DETECTED_OS),Windows)
	powershell -Command "Invoke-WebRequest -Uri 'https://github.com/google/flatbuffers/archive/refs/tags/v$(FLATBUFFERS_VERSION).tar.gz' -OutFile '$@'"
else
	wget -O $@ "https://github.com/google/flatbuffers/archive/refs/tags/v$(FLATBUFFERS_VERSION).tar.gz"
endif

# Extract and patch FlatBuffers as one step, so the sources are never usable
# without the patch. $(FLATBUFFERS_PATCH) carries two unrelated changes:
#   1. musl compatibility, so the Alpine/Linux build works. See
#      https://github.com/sartura/flatbuffers/commit/92bd62407329caacd66e92e5bfd2949f2f137bfe
#   2. forward compatibility: an unknown enum value in the input JSON falls back
#      to 0 (the *_UNKNOWN default) instead of failing the whole parse, so this
#      compiler can still compile a policy produced against a NEWER schema.
# Only (1) is musl-specific. (2) is needed on every platform, so this must not be
# gated on the OS: doing that left the Windows and macOS binaries rejecting
# unknown enum values outright. tests/forward_compat_test.py covers it.
# git apply is not prefixed with '-' on purpose - a patch that stops applying
# (after a FlatBuffers bump, say) has to fail the build rather than silently
# yield a binary that is not forward compatible.
# --ignore-whitespace is needed on Windows: the tarball sources are LF, while
# core.autocrlf=true checks the patch itself out as CRLF, so without it every
# context line mismatches. It does not weaken the check above - a patch that no
# longer matches the sources still fails.
#
# The build depends on this stamp rather than on $(FLATBUFFERS_DIR), because an
# extracted directory is not evidence that it was patched. Once anything is
# created inside it (the cmake build/ tree, say) its mtime overtakes the
# tarball's, so a directory target would read as up to date forever - upgrading an
# already-built Windows or macOS checkout would silently keep linking the
# unpatched library and keep rejecting unknown enum values. Re-extracting inside
# this recipe also makes it self-healing in both directions: whatever state the
# tree was left in, tar restores pristine sources before git apply runs, so this
# never has to guess whether the patch is already applied.
$(FLATBUFFERS_PATCH_STAMP): $(FLATBUFFERS_TARBALL) $(FLATBUFFERS_PATCH_FILE)
	tar -xzf $(FLATBUFFERS_TARBALL)
	cd $(FLATBUFFERS_DIR) && git apply -p0 --ignore-whitespace $(FLATBUFFERS_PATCH)
	echo patched > $@

# Build FlatBuffers static library for Windows
# Uses CMAKE_MSVC_RUNTIME_LIBRARY=MultiThreaded for static linking (/MT)
ifeq ($(DETECTED_OS),Windows)
$(BUILD_DIR)/MinSizeRel/flatbuffers.lib: $(FLATBUFFERS_PATCH_STAMP)
	$(MKDIR) $(BUILD_DIR)
	cd $(BUILD_DIR) && cmake -A x64 -DCMAKE_BUILD_TYPE=MinSizeRel -DCMAKE_MSVC_RUNTIME_LIBRARY=MultiThreaded -DFLATBUFFERS_BUILD_TESTS=OFF -DFLATBUFFERS_BUILD_FLATC=ON -DFLATBUFFERS_BUILD_FLATHASH=OFF ..
	cd $(BUILD_DIR) && cmake --build . --target flatbuffers --config MinSizeRel
	cd $(BUILD_DIR) && cmake --build . --target flatc --config MinSizeRel

flatbuffers: $(BUILD_DIR)/MinSizeRel/flatbuffers.lib
else
$(BUILD_DIR)/libflatbuffers.a: $(FLATBUFFERS_PATCH_STAMP)
	$(MKDIR) $(BUILD_DIR)
ifeq ($(DETECTED_OS),Linux)
	cd $(BUILD_DIR) && cmake -DCMAKE_CXX_FLAGS="-Os -flto -fno-exceptions -fno-rtti -ffunction-sections -fdata-sections -fvisibility=hidden -Wno-stringop-overflow" -DCMAKE_BUILD_TYPE=MinSizeRel -DFLATBUFFERS_BUILD_TESTS=OFF -DFLATBUFFERS_BUILD_FLATC=OFF -DFLATBUFFERS_BUILD_FLATHASH=OFF ..
else
	cd $(BUILD_DIR) && cmake -DCMAKE_CXX_FLAGS="-Os -fno-exceptions -fno-rtti -fvisibility=hidden -Wno-stringop-overflow" -DCMAKE_BUILD_TYPE=MinSizeRel -DFLATBUFFERS_BUILD_TESTS=OFF -DFLATBUFFERS_BUILD_FLATC=OFF -DFLATBUFFERS_BUILD_FLATHASH=OFF ..
endif
	cd $(BUILD_DIR) && cmake --build . --target flatbuffers

flatbuffers: $(BUILD_DIR)/libflatbuffers.a
endif

# Generate Flatbuffers binary schema (bfbs)
$(POLICY_BFBS): $(SCHEMA_PATH) flatbuffers
	$(MKDIR) schema
	$(FLATC) --binary --force-defaults --schema -o schema $(SCHEMA_PATH)

# Generate C header file from binary schema
# On Unix: use xxd tool to convert binary to hex array  
# On Windows: use PowerShell script (xxd not available)
$(POLICY_BFBS_HEADER): $(POLICY_BFBS)
ifeq ($(DETECTED_OS),Windows)
	powershell -ExecutionPolicy Bypass -File convert_binary.ps1 "$(POLICY_BFBS)" "$(POLICY_BFBS_HEADER)"
else
	xxd -i $< > $@
endif

# Build output binary
# $(SRC) #includes $(POLICY_BFBS_HEADER), so it has to be a prerequisite here and
# not only of `build`: prerequisites are unordered, so relying on `build` listing
# both left this racy under -j and broke any target that asked for $(OUT) alone.
$(OUT): $(SRC) flatbuffers $(CXXOPTS_HEADER) $(POLICY_BFBS_HEADER)
	$(MKDIR) $(MAKEFILE_DIR)/bin
	$(CXX) $(CXXFLAGS) $(INCLUDES) $(SRC) $(LIBS) $(CXX_OUTPUT_ARGS)

# Prerequisite on $(OUT), not just ordering within `build`: without it -j runs the
# strip before the link and fails on a missing binary.
strip_binary: $(OUT)
ifneq ($(DETECTED_OS),Windows)
ifeq (, $(shell which $(STRIP)))
	@echo "strip not found, skipping binary strip"
else
	$(STRIP) $(OUT)
endif
endif


# Run the compiler's tests against the built binary.
# Depends on `build`, not $(OUT), so this works as a standalone command on a clean
# checkout and exercises the same artifact CI ships.
test: build
	@$(PYTHON) tests/forward_compat_test.py $(OUT)
	@$(PYTHON) tests/injector_schema_fallback_test.py $(OUT)

# Verify by unmarshalling compiled policy to JSON
verify_json:
ifndef BIN
	$(error You must specify BIN=<path_to_bin_file>)
endif
	$(FLATC) --json --strict-json --raw-binary -o . $(SCHEMA_PATH) -- $(BIN)
	@echo "Verification JSON written to $(BIN:.bin=.json)"

# Clean all build artifacts
clean:
ifeq ($(DETECTED_OS),Windows)
	powershell -Command "Remove-Item -Path '$(FLATBUFFERS_TARBALL)', '$(FLATBUFFERS_DIR)', '$(FLATBUFFERS_PATCH_STAMP)', '$(OUT)', 'bin', 'schema', 'include', '*.bin', '*.json', '$(POLICY_BFBS_HEADER)' -Recurse -Force -ErrorAction SilentlyContinue"
else
	rm -rf $(FLATBUFFERS_TARBALL) $(FLATBUFFERS_DIR) $(FLATBUFFERS_PATCH_STAMP) $(OUT) bin schema include *.bin *.json $(POLICY_BFBS_HEADER)
endif

fmt:
	@echo "[i] Formatting source files..."
	@clang-format -i $(FILES_TO_FORMAT)

fmt-check:
	@echo -e "[i] Files to format:\n$(FILES_TO_FORMAT)"
	@clang-format --dry-run --Werror $(FILES_TO_FORMAT)
