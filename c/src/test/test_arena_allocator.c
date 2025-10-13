#include <stdalign.h>
#include <stdint.h>
#include "arena_allocator.h"
#include "utest/utest.h"

UTEST(arena, alloc) {
  plcs_arena_allocator *arena = plcs_arena_new(sizeof(uint8_t) * 8);

  typedef struct {
    char data[7];
  } buf7;
  typedef struct {
    char data[1];
  } buf1;

  buf7 *ptr = plcs_arena_alloc(arena, 7, alignof(uint8_t));
  ASSERT_TRUE(ptr != NULL);

  buf1 *ptr2 = plcs_arena_alloc(arena, 1, alignof(buf1));
  ASSERT_TRUE(ptr2 != NULL);

  buf1 *ptr3 = plcs_arena_alloc(arena, 1, alignof(buf1));
  ASSERT_TRUE(ptr3 == NULL);

  plcs_arena_reset(arena);

  buf7 *ptr4 = plcs_arena_alloc(arena, 7, alignof(uint8_t));
  ASSERT_TRUE(ptr4 != NULL);

  plcs_arena_free(arena);
}

UTEST(arena, zero_size) {
  plcs_arena_allocator *arena = plcs_arena_new(0);
  EXPECT_TRUE(arena == NULL);
}
