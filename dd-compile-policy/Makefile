# Detect operating system for platform-specific build logic
ifeq ($(OS),Windows_NT)
	DETECTED_OS := Windows
else
	DETECTED_OS := $(shell uname)
endif

# Default architecture for Windows cross-compilation
# Can be overridden: make build CMAKE_ARCH=arm64
CMAKE_ARCH ?= x64

FLATBUFFERS_VERSION = 25.9.23
FLATBUFFERS_TARBALL = flatbuffers-v$(FLATBUFFERS_VERSION).tar.gz
FLATBUFFERS_DIR = flatbuffers-$(FLATBUFFERS_VERSION)
FLATBUFFERS_PATCH = ../flatbuffers.patch # relative to FLATBUFFERS_DIR

CXXOPTS_VERSION = 3.2.1
BUILD_DIR = $(FLATBUFFERS_DIR)/build
SRC = src/compile_policy.cpp

# Platform-specific compiler and tool settings
ifeq ($(DETECTED_OS),Windows)
	OUT = bin/dd-compile-policy.exe
	LIBS = $(BUILD_DIR)/MinSizeRel/flatbuffers.lib
	CXX = cl.exe                      # Microsoft Visual C++ compiler
	# Windows compiler flags:
	# /MT = static runtime (no DLL dependencies)
	# /EHsc = enable C++ exceptions  
	# /GL = whole program optimization
	CXXFLAGS = /std:c++17 /O1 /GL /MT /EHsc /DCXXOPTS_NO_EXCEPTIONS /DCXXOPTS_NO_RTTI
	LDFLAGS = /link /LTCG kernel32.lib user32.lib
	STRIP = echo "Strip not needed on Windows"
	# Windows uses PowerShell for cross-platform commands
	MKDIR = powershell -Command "New-Item -ItemType Directory -Force -Path"
	WGET = powershell -Command "Invoke-WebRequest -Uri"
	WGET_OUT = -OutFile
else
	OUT = bin/dd-compile-policy
	LIBS = $(BUILD_DIR)/libflatbuffers.a
	CXX = g++
	CXXFLAGS = -std=c++17 -Os -flto -g0 -fno-exceptions -fno-rtti -ffunction-sections -fdata-sections -fvisibility=hidden -fvisibility-inlines-hidden -DCXXOPTS_NO_EXCEPTIONS -DCXXOPTS_NO_RTTI
	LDFLAGS = 
	STRIP = strip
	MKDIR = mkdir -p
	WGET = wget -O
	WGET_OUT =
endif

# Additional Linux-specific flags
ifeq ($(DETECTED_OS),Linux)
	CXXFLAGS += -static -static-libgcc -static-libstdc++
	LDFLAGS += -Wl,--gc-sections -Wl,--strip-all -Wl,--as-needed
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
	$(WGET) "https://raw.githubusercontent.com/jarro2783/cxxopts/v$(CXXOPTS_VERSION)/include/cxxopts.hpp" $(WGET_OUT) $@
else
	$(WGET) $@ https://raw.githubusercontent.com/jarro2783/cxxopts/v$(CXXOPTS_VERSION)/include/cxxopts.hpp
endif

# Download & extract flatbuffers
$(FLATBUFFERS_TARBALL):
ifeq ($(DETECTED_OS),Windows)
	$(WGET) "https://github.com/google/flatbuffers/archive/refs/tags/v$(FLATBUFFERS_VERSION).tar.gz" $(WGET_OUT) $@
else
	$(WGET) $@ https://github.com/google/flatbuffers/archive/refs/tags/v$(FLATBUFFERS_VERSION).tar.gz
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
ifeq ($(CMAKE_ARCH),arm64)
	# ARM64 cross-compilation: Build library for ARM64, tool for x64 host
	# This is needed because ARM64 executables can't run on x64 Windows runners
	# 1. Build FlatBuffers library for ARM64 target architecture  
	cd $(BUILD_DIR) && cmake -A ARM64 -DCMAKE_BUILD_TYPE=MinSizeRel -DCMAKE_MSVC_RUNTIME_LIBRARY=MultiThreaded -DFLATBUFFERS_BUILD_TESTS=OFF -DFLATBUFFERS_BUILD_FLATC=OFF -DFLATBUFFERS_BUILD_FLATHASH=OFF ..
	cd $(BUILD_DIR) && cmake --build . --target flatbuffers --config MinSizeRel
	# 2. Build flatc.exe tool for x64 host (so it can run during build)
	$(MKDIR) $(BUILD_DIR)_host
	cd $(BUILD_DIR)_host && cmake -A x64 -DCMAKE_BUILD_TYPE=MinSizeRel -DCMAKE_MSVC_RUNTIME_LIBRARY=MultiThreaded -DFLATBUFFERS_BUILD_TESTS=OFF -DFLATBUFFERS_BUILD_FLATC=ON -DFLATBUFFERS_BUILD_FLATHASH=OFF ..
	cd $(BUILD_DIR)_host && cmake --build . --target flatc --config MinSizeRel
	# 3. Copy x64 flatc.exe to expected location for schema generation
	powershell -Command "Copy-Item '$(BUILD_DIR)_host/MinSizeRel/flatc.exe' '$(BUILD_DIR)/MinSizeRel/flatc.exe'"
else
	# x64 build: Everything can be built for the same architecture
	cd $(BUILD_DIR) && cmake -A x64 -DCMAKE_BUILD_TYPE=MinSizeRel -DCMAKE_MSVC_RUNTIME_LIBRARY=MultiThreaded -DFLATBUFFERS_BUILD_TESTS=OFF -DFLATBUFFERS_BUILD_FLATC=ON -DFLATBUFFERS_BUILD_FLATHASH=OFF ..
	cd $(BUILD_DIR) && cmake --build . --target flatbuffers --config MinSizeRel
	cd $(BUILD_DIR) && cmake --build . --target flatc --config MinSizeRel
endif

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
ifeq ($(DETECTED_OS),Windows)
	$(BUILD_DIR)/MinSizeRel/flatc.exe --binary --force-defaults --schema -o schema $(SCHEMA_PATH)
else
	flatc --binary --force-defaults --schema -o schema $(SCHEMA_PATH)
endif

# Generate C header file from binary schema
# On Unix: use xxd tool to convert binary to hex array  
# On Windows: use PowerShell script (xxd not available)
$(POLICY_BFBS_HEADER): $(POLICY_BFBS)
ifeq ($(DETECTED_OS),Windows)
	# Create temporary PowerShell script to convert binary to C header
	# This mimics the behavior of 'xxd -i' on Unix systems
	@echo '$$bytes = Get-Content "$(POLICY_BFBS)" -Encoding Byte' > convert_binary.ps1
	@echo '$$hexArray = $$bytes | ForEach-Object { "0x{0:x2}" -f $$_ }' >> convert_binary.ps1
	@echo '$$hexString = $$hexArray -join ", "' >> convert_binary.ps1  
	@echo '$$header = "unsigned char schema_policy_bfbs[] = { " + $$hexString + " };"' >> convert_binary.ps1
	@echo '$$length = "unsigned int schema_policy_bfbs_len = " + $$bytes.Length + ";"' >> convert_binary.ps1
	@echo '$$header + [System.Environment]::NewLine + $$length | Out-File "$(POLICY_BFBS_HEADER)" -Encoding ascii' >> convert_binary.ps1
	powershell -ExecutionPolicy Bypass -File convert_binary.ps1
	powershell -Command "Remove-Item convert_binary.ps1"
else
	xxd -i $< > $@
endif

# Build output binary
$(OUT): $(SRC) flatbuffers $(CXXOPTS_HEADER)
	$(MKDIR) $(MAKEFILE_DIR)/bin
ifeq ($(DETECTED_OS),Windows)
	$(CXX) $(CXXFLAGS) $(INCLUDES) $(SRC) $(LIBS) /Fe$@ $(LDFLAGS)
else
ifeq ($(DETECTED_OS),Linux)
	$(CXX) $(CXXFLAGS) $(INCLUDES) $(SRC) $(LIBS) $(LDFLAGS) -o $@
else
	$(CXX) $(CXXFLAGS) $(INCLUDES) $(SRC) $(LIBS) -o $@
endif
endif

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
ifeq ($(DETECTED_OS),Windows)
	$(BUILD_DIR)/MinSizeRel/flatc.exe --json --strict-json --raw-binary -o . $(SCHEMA_PATH) -- $(BIN)
else
	flatc --json --strict-json --raw-binary -o . $(SCHEMA_PATH) -- $(BIN)
endif
	@echo "Verification JSON written to $(BIN:.bin=.json)"

# Clean all build artifacts
# Windows: Include host build directory used for ARM64 cross-compilation
clean:
ifeq ($(DETECTED_OS),Windows)
	powershell -Command "Remove-Item -Path '$(FLATBUFFERS_TARBALL)', '$(FLATBUFFERS_DIR)', '$(BUILD_DIR)_host', '$(OUT)', 'bin', 'schema', 'include', '*.bin', '*.json', '$(POLICY_BFBS_HEADER)' -Recurse -Force -ErrorAction SilentlyContinue"
else
	rm -rf $(FLATBUFFERS_TARBALL) $(FLATBUFFERS_DIR) $(OUT) bin schema include *.bin *.json $(POLICY_BFBS_HEADER)
endif

fmt:
	@echo "[i] Formatting source files..."
	@clang-format -i $(FILES_TO_FORMAT)

fmt-check:
	@echo -e "[i] Files to format:\n$(FILES_TO_FORMAT)"
	@clang-format --dry-run --Werror $(FILES_TO_FORMAT)
