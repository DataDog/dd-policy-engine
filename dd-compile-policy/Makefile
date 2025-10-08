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

CXX = clang++
CXXFLAGS = -std=c++17 -Os -flto -g0 -fno-exceptions -fno-rtti -ffunction-sections -fdata-sections -fvisibility=hidden -fvisibility-inlines-hidden -DCXXOPTS_NO_EXCEPTIONS -DCXXOPTS_NO_RTTI
LDFLAGS = 
ifeq ($(shell uname), Linux)
	CXXFLAGS += -static -static-libgcc -static-libstdc++
	LDFLAGS += -Wl,--gc-sections -Wl,--strip-all -Wl,--as-needed
endif
INCLUDES = -I$(FLATBUFFERS_DIR)/include -Isrc -Iinclude
LIBS = $(BUILD_DIR)/libflatbuffers.a
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
	mkdir -p include
	wget -O $@ https://raw.githubusercontent.com/jarro2783/cxxopts/v$(CXXOPTS_VERSION)/include/cxxopts.hpp

# Download & extract flatbuffers
$(FLATBUFFERS_TARBALL):
	wget -O $@ https://github.com/google/flatbuffers/archive/refs/tags/v$(FLATBUFFERS_VERSION).tar.gz

$(FLATBUFFERS_DIR): $(FLATBUFFERS_TARBALL)
	tar -xzf $<

# Flatbuffers isn't working too well on musl, so we need to patch it.
# See https://github.com/sartura/flatbuffers/commit/92bd62407329caacd66e92e5bfd2949f2f137bfe
patch_flatbuffers_musl:
	@echo "Patching flatbuffers base.h for musl compatibility..."
	@cd ./$(FLATBUFFERS_DIR) && git apply -p0 $(FLATBUFFERS_PATCH) ; cd -

# Build Flatbuffers static library
$(BUILD_DIR)/libflatbuffers.a: $(FLATBUFFERS_DIR) patch_flatbuffers_musl
	mkdir -p $(BUILD_DIR)
ifeq ($(shell uname), Linux)
	cd $(BUILD_DIR) && cmake -DCMAKE_CXX_FLAGS="-Os -flto -fno-exceptions -fno-rtti -ffunction-sections -fdata-sections -fvisibility=hidden -Wno-stringop-overflow" -DCMAKE_BUILD_TYPE=MinSizeRel -DFLATBUFFERS_BUILD_TESTS=OFF -DFLATBUFFERS_BUILD_FLATC=OFF -DFLATBUFFERS_BUILD_FLATHASH=OFF ..
else
	cd $(BUILD_DIR) && cmake -DCMAKE_CXX_FLAGS="-Os -fno-exceptions -fno-rtti -fvisibility=hidden -Wno-stringop-overflow" -DCMAKE_BUILD_TYPE=MinSizeRel -DFLATBUFFERS_BUILD_TESTS=OFF -DFLATBUFFERS_BUILD_FLATC=OFF -DFLATBUFFERS_BUILD_FLATHASH=OFF ..
endif
	cd $(BUILD_DIR) && cmake --build . --target flatbuffers

flatbuffers: $(BUILD_DIR)/libflatbuffers.a

# Generate Flatbuffers binary schema (bfbs)
$(POLICY_BFBS): $(SCHEMA_PATH)
	flatc --binary --force-defaults --schema -o schema $(SCHEMA_PATH)

# Generate Flatbuffers header file from binary schema
$(POLICY_BFBS_HEADER): $(POLICY_BFBS)
	xxd -i $< > $@

# Build output binary
$(OUT): $(SRC) flatbuffers $(CXXOPTS_HEADER)
	mkdir -p $(MAKEFILE_DIR)/bin
ifeq ($(shell uname), Linux)
	$(CXX) $(CXXFLAGS) $(INCLUDES) $(SRC) $(LIBS) $(LDFLAGS) -o $@
else
	$(CXX) $(CXXFLAGS) $(INCLUDES) $(SRC) $(LIBS) -o $@
endif

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
	rm -rf $(FLATBUFFERS_TARBALL) $(FLATBUFFERS_DIR) $(OUT) bin schema include *.bin *.json $(POLICY_BFBS_HEADER)

fmt:
	@echo "[i] Formatting source files..."
	@clang-format -i $(FILES_TO_FORMAT)

fmt-check:
	@echo -e "[i] Files to format:\n$(FILES_TO_FORMAT)"
	@clang-format --dry-run --Werror $(FILES_TO_FORMAT)
