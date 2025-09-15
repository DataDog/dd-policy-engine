
#include <dd/policies/error_codes.h>
#include <dd/policies/eval_ctx.h>
#include <dd/policies/policies.h>

#include "eval_ctx.h"

#include <stdbool.h>

static plcs_eval_ctx ctx;
static bool plcs_eval_ctx_initialized = false;
const size_t PLCS_MAX_STR_PARAM_LENGTH = 1024;

plcs_errors
plcs_eval_ctx_register_str_evaluator(plcs_string_evaluator_function_ptr func_ptr, plcs_string_evaluators ix) {
  if (!func_ptr) {
    return DD_EREGISTER_EVAL_PTR;
  }

  if (ix >= STR_EVAL__COUNT) {
    return DD_EIX_OVERFLOW;
  }

  ctx.string_evaluators[ix].function_ptr = func_ptr;
  return DD_ESUCCESS;
}

plcs_errors
plcs_eval_ctx_register_num_evaluator(plcs_numeric_evaluator_function_ptr func_ptr, plcs_numeric_evaluators ix) {
  if (!func_ptr) {
    return DD_EREGISTER_EVAL_PTR;
  }

  if (ix >= NUM_EVAL__COUNT) {
    return DD_EIX_OVERFLOW;
  }

  ctx.numeric_evaluators[ix].function_ptr = func_ptr;
  return DD_ESUCCESS;
}

plcs_errors
plcs_eval_ctx_register_unum_evaluator(plcs_unumeric_evaluator_function_ptr func_ptr, plcs_numeric_evaluators ix) {
  if (!func_ptr) {
    return DD_EREGISTER_EVAL_PTR;
  }

  if (ix >= NUM_EVAL__COUNT) {
    return DD_EIX_OVERFLOW;
  }

  ctx.unumeric_evaluators[ix].function_ptr = func_ptr;
  return DD_ESUCCESS;
}

static void *plcs_strdup_with_allocator(plcs_arena_allocator *allocator, const char *str) {
  if (str == NULL)
    return NULL;

  size_t len = strlen(str) + 1;  //< +1 for null terminated character
  void *buffer = arena_alloc_aligned(allocator, len, 1);
  if (buffer == NULL) {
    return NULL;
  }

  memcpy(buffer, str, len);
  return buffer;
}

plcs_errors plcs_eval_ctx_set_str_eval_param(plcs_string_evaluators ix, const char *value) {
  if (ix >= STR_EVAL__COUNT) {
    return DD_EIX_OVERFLOW;
  }

  size_t len = strlen(value) + 1;
  if (len > PLCS_MAX_STR_PARAM_LENGTH) {
    return DD_ESTR_EVAL_EXCEED_MAX_LENGTH;
  }

  void *newstr = plcs_strdup_with_allocator(&ctx.allocator, value);
  if (newstr == NULL) {
  }

  ctx.string_evaluators[ix].value = newstr;
  return DD_ESUCCESS;
}

plcs_errors plcs_eval_ctx_set_num_eval_param(plcs_numeric_evaluators ix, const long value) {
  if (ix >= NUM_EVAL__COUNT) {
    return DD_EIX_OVERFLOW;
  }

  ctx.numeric_evaluators[ix].value = value;
  return DD_ESUCCESS;
}

plcs_errors plcs_eval_ctx_set_unum_eval_param(plcs_numeric_evaluators ix, const unsigned long value) {
  if (ix >= NUM_EVAL__COUNT) {
    return DD_EIX_OVERFLOW;
  }

  ctx.unumeric_evaluators[ix].value = value;
  return DD_ESUCCESS;
}

plcs_errors plcs_eval_ctx_register_action(plcs_action_function_ptr action, plcs_actions ix) {
  if (ix >= ACTIONS__COUNT) {
    return DD_EIX_OVERFLOW;
  }

  ctx.actions[ix].function_ptr = action;
  return DD_ESUCCESS;
}

plcs_action_function_ptr plcs_eval_ctx_get_action(plcs_actions ix) {
  if (ix >= ACTIONS__COUNT) {
    ctx.error = DD_EIX_OVERFLOW;
    return NULL;
  }

  return ctx.actions[ix].function_ptr;
}

plcs_string_evaluator_function_ptr plcs_eval_ctx_get_string_evaluator(plcs_string_evaluators id) {
  if (id >= STR_EVAL__COUNT) {
    ctx.error = DD_EIX_OVERFLOW;
    return NULL;
  }

  return ctx.string_evaluators[id].function_ptr;
}

const char *plcs_eval_ctx_get_string_param(plcs_string_evaluators id) {
  if (id >= STR_EVAL__COUNT) {
    ctx.error = DD_EIX_OVERFLOW;
    return STR_NOT_SET;
  }

  return ctx.string_evaluators[id].value;
}

plcs_numeric_evaluator_function_ptr plcs_eval_ctx_get_numeric_evaluator(plcs_numeric_evaluators id) {
  if (id >= NUM_EVAL__COUNT) {
    ctx.error = DD_EIX_OVERFLOW;
    return NULL;
  }

  return ctx.numeric_evaluators[id].function_ptr;
}

long plcs_eval_ctx_get_numeric_param(plcs_numeric_evaluators id) {
  if (id >= NUM_EVAL__COUNT) {
    ctx.error = DD_EIX_OVERFLOW;
    return NUM_NOT_SET;
  }

  return ctx.numeric_evaluators[id].value;
}

plcs_unumeric_evaluator_function_ptr plcs_eval_ctx_get_unumeric_evaluator(plcs_numeric_evaluators id) {
  if (id >= NUM_EVAL__COUNT) {
    ctx.error = DD_EIX_OVERFLOW;
    return NULL;
  }

  return ctx.unumeric_evaluators[id].function_ptr;
}

unsigned long plcs_eval_ctx_get_unumeric_param(plcs_numeric_evaluators id) {
  if (id >= NUM_EVAL__COUNT) {
    ctx.error = DD_EIX_OVERFLOW;
    return UNUM_NOT_SET;
  }

  return ctx.unumeric_evaluators[id].value;
}

// TODO: consider implementing it as a stack to preserve error history
void plcs_eval_ctx_set_error(plcs_errors error) {
  ctx.error = error;
}

// TODO: consider implementing it as a stack to preserve error history
void plcs_eval_ctx_set_action_error(plcs_actions ix, plcs_errors error) {
  if (ix < ACTIONS__COUNT) {
    ctx.actions[ix].error = error;
  }
}

void plcs_eval_ctx_set_str_eval_error(plcs_string_evaluators ix, plcs_errors error) {
  if (ix < STR_EVAL__COUNT) {
    ctx.string_evaluators[ix].error = error;
  }
}

plcs_errors plcs_eval_ctx_get_str_eval_error(plcs_string_evaluators ix) {
  if (ix < STR_EVAL__COUNT) {
    return ctx.string_evaluators[ix].error;
  }
  return DD_EIX_OVERFLOW;
}

void plcs_eval_ctx_set_num_eval_error(plcs_numeric_evaluators ix, plcs_errors error) {
  if (ix < NUM_EVAL__COUNT) {
    ctx.numeric_evaluators[ix].error = error;
  }
}

plcs_errors plcs_eval_ctx_get_num_eval_error(plcs_numeric_evaluators ix) {
  if (ix < NUM_EVAL__COUNT) {
    return ctx.numeric_evaluators[ix].error;
  }
  return DD_EIX_OVERFLOW;
}

void plcs_eval_ctx_set_unum_eval_error(plcs_numeric_evaluators ix, plcs_errors error) {
  if (ix < NUM_EVAL__COUNT) {
    ctx.unumeric_evaluators[ix].error = error;
  }
}

plcs_errors plcs_eval_ctx_get_unum_eval_error(plcs_numeric_evaluators ix) {
  if (ix < NUM_EVAL__COUNT) {
    return ctx.unumeric_evaluators[ix].error;
  }
  return DD_EIX_OVERFLOW;
}

plcs_errors plcs_eval_ctx_peek_last_error(void) {
  return ctx.error;
}

plcs_errors plcs_eval_ctx_get_last_error(void) {
  plcs_errors error = ctx.error;
  // reset
  ctx.error = DD_ESUCCESS;
  return error;
}

void plcs_eval_ctx_reset(void) {
  plcs_arena_clear(&ctx.allocator);

  // Reset all evaluators to NULL and parameters to their 'not set' values
  // Initialize all evaluators to NULL
  for (int i = 0; i < STR_EVAL__COUNT; ++i) {
    ctx.string_evaluators[i].error = DD_ESUCCESS;
    ctx.string_evaluators[i].function_ptr = NULL;
    ctx.string_evaluators[i].value = STR_NOT_SET;
  }

  for (int i = 0; i < NUM_EVAL__COUNT; ++i) {
    ctx.numeric_evaluators[i].error = DD_ESUCCESS;
    ctx.numeric_evaluators[i].function_ptr = NULL;
    ctx.numeric_evaluators[i].value = NUM_NOT_SET;

    ctx.unumeric_evaluators[i].error = DD_ESUCCESS;
    ctx.unumeric_evaluators[i].function_ptr = NULL;
    ctx.unumeric_evaluators[i].value = NUM_NOT_SET;
  }

  for (int i = 0; i < ACTIONS__COUNT; ++i) {
    ctx.actions[i].function_ptr = NULL;
    ctx.actions[i].error = DD_ESUCCESS;
  }

  ctx.error = DD_ESUCCESS;
}

#define DD_GET_ARRAY_SIZE(x) (size_t)sizeof(buffer) / sizeof(buffer[0])

plcs_errors plcs_eval_ctx_init(void) {
  if (plcs_eval_ctx_initialized) {
    return DD_EINITIALIZED;
  }

  // Allocation: 32 * 1024B = 32 KiB
  uint8_t *buffer = calloc(STR_EVAL__COUNT, PLCS_MAX_STR_PARAM_LENGTH);
  if (buffer == NULL) {
    return DD_EINITIALIZED;
  }

  ctx.allocator = plcs_arena_init(&buffer, PLCS_MAX_STR_PARAM_LENGTH * STR_EVAL__COUNT);
  plcs_eval_ctx_reset();

  ctx.error = DD_ESUCCESS;

  plcs_eval_ctx_initialized = true;
  return DD_ESUCCESS;
}
