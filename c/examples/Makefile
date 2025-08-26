PREFIX ?= $(CURDIR)

EXAMPLES := basic-reader hardcoded-header-reader mmap-reader hardcoded_cmdline_evaluator
FILES_TO_CLEAN += $(EXAMPLES)

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
CFLAGS += -I$(PREFIX)/../include

# for examples
LDFLAGS += \
	-L$(PREFIX)/../lib/ \
	-lpolicies

.PHONY: all clean $(EXAMPLES)

all: $(EXAMPLES)

$(EXAMPLES): %: 
	$(CC) $(CFLAGS) $@.c $(LDFLAGS) -o $@

clean:
	@echo "[i] Cleaning up..."
	@rm -rf $(FILES_TO_CLEAN)
	@echo "[i] done cleaning"
