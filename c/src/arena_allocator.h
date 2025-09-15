#pragma once

#include <stddef.h>
#include <stdint.h>

typedef struct {
  uint8_t *buffer;
  size_t size;
  size_t offset;
} plcs_arena_allocator;

plcs_arena_allocator plcs_arena_init(void *buffer, size_t size);

void plcs_arena_clear(plcs_arena_allocator *arena);

void *arena_alloc_aligned(plcs_arena_allocator *arena, size_t size, size_t alignment);
