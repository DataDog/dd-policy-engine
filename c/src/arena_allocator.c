#include "arena_allocator.h"

#include <string.h>

plcs_arena_allocator plcs_arena_init(void *buffer, size_t size) {
  plcs_arena_allocator arena;
  arena.buffer = (uint8_t *)buffer;
  arena.size = size;
  arena.offset = 0;

  return arena;
}

void plcs_arena_clear(plcs_arena_allocator *arena) {
  memset(arena->buffer, 0, arena->size);
}

void *arena_alloc_aligned(plcs_arena_allocator *arena, size_t size, size_t alignment) {
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
