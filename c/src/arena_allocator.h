/*
 * Unless explicitly stated otherwise all files in this repository are licensed
 * under the Apache 2.0 License. This product includes software developed at
 * Datadog (https://www.datadoghq.com/).
 *
 * Copyright 2024-Present Datadog, Inc.
 */
#pragma once

#include <stddef.h>
#include <stdint.h>

typedef struct {
  uint8_t *buffer;
  size_t size;
  size_t offset;
} plcs_arena_allocator;

plcs_arena_allocator plcs_arena_new(void *buffer, size_t size);

void plcs_arena_reset(plcs_arena_allocator *arena);

void *plcs_arena_alloc(plcs_arena_allocator *arena, size_t size, size_t alignment);
