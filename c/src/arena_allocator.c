/*
 * Unless explicitly stated otherwise all files in this repository are licensed
 * under the Apache 2.0 License. This product includes software developed at
 * Datadog (https://www.datadoghq.com/).
 *
 * Copyright 2024-Present Datadog, Inc.
 */
#include "arena_allocator.h"

#include <stdlib.h>
#include <string.h>

plcs_arena_allocator *plcs_arena_new(size_t size) {
  if (size == 0) {
    return NULL;
  }

  void *buffer = malloc(size);
  if (buffer == NULL) {
    return NULL;
  }

  plcs_arena_allocator *arena = (plcs_arena_allocator *)malloc(sizeof(plcs_arena_allocator));
  arena->buffer = (uint8_t *)buffer;
  arena->size = size;
  arena->offset = 0;
  return arena;
}

void plcs_arena_free(plcs_arena_allocator *arena) {
  free(arena->buffer);
  arena->buffer = NULL;

  arena->size = 0;
  arena->offset = 0;
}

void plcs_arena_reset(plcs_arena_allocator *arena) {
  arena->offset = 0;
}

void *plcs_arena_alloc(plcs_arena_allocator *arena, size_t size, size_t alignment) {
  uintptr_t current = (uintptr_t)(arena->buffer + arena->offset);
  uintptr_t aligned = (current + (alignment - 1)) & ~(uintptr_t)(alignment - 1);
  size_t new_offset = aligned - (uintptr_t)arena->buffer + size;

  if (new_offset > arena->size) {
    return NULL;
  }

  // Reconstruct the aligned pointer relative to arena->buffer.
  // Although we could return (void *)aligned directly, rebuilding it this way
  // ensures the returned pointer is explicitly derived from the original buffer.
  // This helps memory analysis tools (like Valgrind or AddressSanitizer)
  // understand that the pointer belongs to the allocated region.
  void *ptr = (void *)(arena->buffer + aligned - (uintptr_t)arena->buffer);

  arena->offset = new_offset;
  return ptr;
}
