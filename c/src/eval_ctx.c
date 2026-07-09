/*
 * Unless explicitly stated otherwise all files in this repository are licensed
 * under the Apache 2.0 License. This product includes software developed at
 * Datadog (https://www.datadoghq.com/).
 *
 * Copyright 2025-Present Datadog, Inc.
 */
#include <dd/policies/error_codes.h>
#include <dd/policies/eval_ctx.h>
#include <dd/policies/policies.h>

#include "eval_ctx.h"

#include <stdbool.h>

static plcs_eval_ctx ctx;
static bool plcs_eval_ctx_initialized = false;

plcs_errors
plcs_eval_ctx_register_str_evaluator(plcs_string_evaluator_function_ptr func_ptr, plcs_string_evaluators ix) {
  if (!func_ptr) {
    return PLCS_EREGISTER_EVAL_PTR;
  }

  if (ix < 0 || ix >= PLCS_STR_EVAL__COUNT) {
    return PLCS_EIX_OVERFLOW;
  }

  ctx.string_evaluators[ix].function_ptr = func_ptr;
  return PLCS_ESUCCESS;
}

plcs_errors
plcs_eval_ctx_register_num_evaluator(plcs_numeric_evaluator_function_ptr func_ptr, plcs_numeric_evaluators ix) {
  if (!func_ptr) {
    return PLCS_EREGISTER_EVAL_PTR;
  }

  if (ix < 0 || ix >= PLCS_NUM_EVAL__COUNT) {
    return PLCS_EIX_OVERFLOW;
  }

  ctx.numeric_evaluators[ix].function_ptr = func_ptr;
  return PLCS_ESUCCESS;
}

plcs_errors
plcs_eval_ctx_register_unum_evaluator(plcs_unumeric_evaluator_function_ptr func_ptr, plcs_numeric_evaluators ix) {
  if (!func_ptr) {
    return PLCS_EREGISTER_EVAL_PTR;
  }

  if (ix < 0 || ix >= PLCS_NUM_EVAL__COUNT) {
    return PLCS_EIX_OVERFLOW;
  }

  ctx.unumeric_evaluators[ix].function_ptr = func_ptr;
  return PLCS_ESUCCESS;
}

plcs_errors plcs_eval_ctx_set_str_eval_param(plcs_string_evaluators ix, const char *value) {
  if (ix < 0 || ix >= PLCS_STR_EVAL__COUNT) {
    return PLCS_EIX_OVERFLOW;
  }

  ctx.string_evaluators[ix].value = value;
  return PLCS_ESUCCESS;
}

plcs_errors plcs_eval_ctx_set_num_eval_param(plcs_numeric_evaluators ix, const long value) {
  if (ix < 0 || ix >= PLCS_NUM_EVAL__COUNT) {
    return PLCS_EIX_OVERFLOW;
  }

  ctx.numeric_evaluators[ix].value = value;
  return PLCS_ESUCCESS;
}

plcs_errors plcs_eval_ctx_set_unum_eval_param(plcs_numeric_evaluators ix, const unsigned long value) {
  if (ix < 0 || ix >= PLCS_NUM_EVAL__COUNT) {
    return PLCS_EIX_OVERFLOW;
  }

  ctx.unumeric_evaluators[ix].value = value;
  return PLCS_ESUCCESS;
}

plcs_errors plcs_eval_ctx_register_action(plcs_action_function_ptr action, plcs_actions ix) {
  if (ix < 0 || ix >= PLCS_ACTIONS__COUNT) {
    return PLCS_EIX_OVERFLOW;
  }

  ctx.actions[ix].function_ptr = action;
  return PLCS_ESUCCESS;
}

void plcs_eval_ctx_set_current_policy_id(const char *policy_id) {
  ctx.current_policy_id = policy_id;
}

plcs_action_function_ptr plcs_eval_ctx_get_action(plcs_actions ix) {
  if (ix < 0 || ix >= PLCS_ACTIONS__COUNT) {
    ctx.error = PLCS_EIX_OVERFLOW;
    return NULL;
  }

  return ctx.actions[ix].function_ptr;
}

plcs_string_evaluator_function_ptr plcs_eval_ctx_get_string_evaluator(plcs_string_evaluators id) {
  if (id < 0 || id >= PLCS_STR_EVAL__COUNT) {
    ctx.error = PLCS_EIX_OVERFLOW;
    return NULL;
  }

  return ctx.string_evaluators[id].function_ptr;
}

const char *plcs_eval_ctx_get_string_param(plcs_string_evaluators id) {
  if (id < 0 || id >= PLCS_STR_EVAL__COUNT) {
    ctx.error = PLCS_EIX_OVERFLOW;
    return PLCS_STR_NOT_SET;
  }

  return ctx.string_evaluators[id].value;
}

plcs_numeric_evaluator_function_ptr plcs_eval_ctx_get_numeric_evaluator(plcs_numeric_evaluators id) {
  if (id < 0 || id >= PLCS_NUM_EVAL__COUNT) {
    ctx.error = PLCS_EIX_OVERFLOW;
    return NULL;
  }

  return ctx.numeric_evaluators[id].function_ptr;
}

long plcs_eval_ctx_get_numeric_param(plcs_numeric_evaluators id) {
  if (id < 0 || id >= PLCS_NUM_EVAL__COUNT) {
    ctx.error = PLCS_EIX_OVERFLOW;
    return PLCS_NUM_NOT_SET;
  }

  return ctx.numeric_evaluators[id].value;
}

plcs_unumeric_evaluator_function_ptr plcs_eval_ctx_get_unumeric_evaluator(plcs_numeric_evaluators id) {
  if (id < 0 || id >= PLCS_NUM_EVAL__COUNT) {
    ctx.error = PLCS_EIX_OVERFLOW;
    return NULL;
  }

  return ctx.unumeric_evaluators[id].function_ptr;
}

unsigned long plcs_eval_ctx_get_unumeric_param(plcs_numeric_evaluators id) {
  if (id < 0 || id >= PLCS_NUM_EVAL__COUNT) {
    ctx.error = PLCS_EIX_OVERFLOW;
    return PLCS_UNUM_NOT_SET;
  }

  return ctx.unumeric_evaluators[id].value;
}

// TODO: consider implementing it as a stack to preserve error history
void plcs_eval_ctx_set_error(plcs_errors error) {
  ctx.error = error;
}

// TODO: consider implementing it as a stack to preserve error history
void plcs_eval_ctx_set_action_error(plcs_actions ix, plcs_errors error) {
  if (ix >= 0 && ix < PLCS_ACTIONS__COUNT) {
    ctx.actions[ix].error = error;
  }
}

void plcs_eval_ctx_set_str_eval_error(plcs_string_evaluators ix, plcs_errors error) {
  if (ix >= 0 && ix < PLCS_STR_EVAL__COUNT) {
    ctx.string_evaluators[ix].error = error;
  }
}

plcs_errors plcs_eval_ctx_get_str_eval_error(plcs_string_evaluators ix) {
  if (ix >= 0 && ix < PLCS_STR_EVAL__COUNT) {
    return ctx.string_evaluators[ix].error;
  }
  return PLCS_EIX_OVERFLOW;
}

void plcs_eval_ctx_set_num_eval_error(plcs_numeric_evaluators ix, plcs_errors error) {
  if (ix >= 0 && ix < PLCS_NUM_EVAL__COUNT) {
    ctx.numeric_evaluators[ix].error = error;
  }
}

plcs_errors plcs_eval_ctx_get_num_eval_error(plcs_numeric_evaluators ix) {
  if (ix >= 0 && ix < PLCS_NUM_EVAL__COUNT) {
    return ctx.numeric_evaluators[ix].error;
  }
  return PLCS_EIX_OVERFLOW;
}

void plcs_eval_ctx_set_unum_eval_error(plcs_numeric_evaluators ix, plcs_errors error) {
  if (ix >= 0 && ix < PLCS_NUM_EVAL__COUNT) {
    ctx.unumeric_evaluators[ix].error = error;
  }
}

plcs_errors plcs_eval_ctx_get_unum_eval_error(plcs_numeric_evaluators ix) {
  if (ix >= 0 && ix < PLCS_NUM_EVAL__COUNT) {
    return ctx.unumeric_evaluators[ix].error;
  }
  return PLCS_EIX_OVERFLOW;
}

plcs_errors plcs_eval_ctx_peek_last_error(void) {
  return ctx.error;
}

const char *plcs_eval_ctx_get_current_policy_id(void) {
  return ctx.current_policy_id;
}

plcs_errors plcs_eval_ctx_get_last_error(void) {
  plcs_errors error = ctx.error;
  // reset
  ctx.error = PLCS_ESUCCESS;
  return error;
}

void plcs_eval_ctx_reset(void) {
  // Reset all evaluators to NULL and parameters to their 'not set' values
  // Initialize all evaluators to NULL
  for (int i = 0; i < PLCS_STR_EVAL__COUNT; ++i) {
    ctx.string_evaluators[i].error = PLCS_ESUCCESS;
    ctx.string_evaluators[i].function_ptr = NULL;
    ctx.string_evaluators[i].value = PLCS_STR_NOT_SET;
  }

  for (int i = 0; i < PLCS_NUM_EVAL__COUNT; ++i) {
    ctx.numeric_evaluators[i].error = PLCS_ESUCCESS;
    ctx.numeric_evaluators[i].function_ptr = NULL;
    ctx.numeric_evaluators[i].value = PLCS_NUM_NOT_SET;

    ctx.unumeric_evaluators[i].error = PLCS_ESUCCESS;
    ctx.unumeric_evaluators[i].function_ptr = NULL;
    ctx.unumeric_evaluators[i].value = PLCS_NUM_NOT_SET;
  }

  for (int i = 0; i < PLCS_ACTIONS__COUNT; ++i) {
    ctx.actions[i].function_ptr = NULL;
    ctx.actions[i].error = PLCS_ESUCCESS;
  }

  ctx.error = PLCS_ESUCCESS;
  ctx.current_policy_id = PLCS_STR_NOT_SET;
}

plcs_errors plcs_eval_ctx_init(void) {
  if (plcs_eval_ctx_initialized) {
    return PLCS_EINITIZLIED;
  }

  plcs_eval_ctx_reset();

  ctx.error = PLCS_ESUCCESS;

  plcs_eval_ctx_initialized = true;
  return PLCS_ESUCCESS;
}
