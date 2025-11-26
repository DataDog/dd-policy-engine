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
.PHONY: all clean flatbuffers strip_binary

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

$(FLATBUFFERS_DIR): $(FLATBUFFERS_TARBALL)
	tar -xzf $<

# Flatbuffers isn't working too well on musl, so we need to patch it.
# See https://github.com/sartura/flatbuffers/commit/92bd62407329caacd66e92e5bfd2949f2f137bfe
# Apply musl compatibility patch for Alpine Linux builds
# Windows and other platforms don't need this patch
patch_flatbuffers_musl:
ifeq ($(DETECTED_OS),Linux)
	@cd ./$(FLATBUFFERS_DIR) && git apply -p0 $(FLATBUFFERS_PATCH) ; cd -
endif

# Build FlatBuffers static library for Windows
# Uses CMAKE_MSVC_RUNTIME_LIBRARY=MultiThreaded for static linking (/MT)
ifeq ($(DETECTED_OS),Windows)
$(BUILD_DIR)/MinSizeRel/flatbuffers.lib: $(FLATBUFFERS_DIR) patch_flatbuffers_musl
	$(MKDIR) $(BUILD_DIR)
	cd $(BUILD_DIR) && cmake -A x64 -DCMAKE_BUILD_TYPE=MinSizeRel -DCMAKE_MSVC_RUNTIME_LIBRARY=MultiThreaded -DFLATBUFFERS_BUILD_TESTS=OFF -DFLATBUFFERS_BUILD_FLATC=ON -DFLATBUFFERS_BUILD_FLATHASH=OFF ..
	cd $(BUILD_DIR) && cmake --build . --target flatbuffers --config MinSizeRel
	cd $(BUILD_DIR) && cmake --build . --target flatc --config MinSizeRel

flatbuffers: $(BUILD_DIR)/MinSizeRel/flatbuffers.lib
else
$(BUILD_DIR)/libflatbuffers.a: $(FLATBUFFERS_DIR) patch_flatbuffers_musl
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
$(OUT): $(SRC) flatbuffers $(CXXOPTS_HEADER)
	$(MKDIR) $(MAKEFILE_DIR)/bin
	$(CXX) $(CXXFLAGS) $(INCLUDES) $(SRC) $(LIBS) $(CXX_OUTPUT_ARGS)

strip_binary:
ifneq ($(DETECTED_OS),Windows)
ifeq (, $(shell which $(STRIP)))
	@echo "strip not found, skipping binary strip"
else
	$(STRIP) $(OUT)
endif
endif


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
	powershell -Command "Remove-Item -Path '$(FLATBUFFERS_TARBALL)', '$(FLATBUFFERS_DIR)', '$(OUT)', 'bin', 'schema', 'include', '*.bin', '*.json', '$(POLICY_BFBS_HEADER)' -Recurse -Force -ErrorAction SilentlyContinue"
else
	rm -rf $(FLATBUFFERS_TARBALL) $(FLATBUFFERS_DIR) $(OUT) bin schema include *.bin *.json $(POLICY_BFBS_HEADER)
endif

fmt:
	@echo "[i] Formatting source files..."
	@clang-format -i $(FILES_TO_FORMAT)

fmt-check:
	@echo -e "[i] Files to format:\n$(FILES_TO_FORMAT)"
	@clang-format --dry-run --Werror $(FILES_TO_FORMAT)
