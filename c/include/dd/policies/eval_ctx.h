/*
 * Unless explicitly stated otherwise all files in this repository are licensed
 * under the Apache 2.0 License. This product includes software developed at
 * Datadog (https://www.datadoghq.com/).
 *
 * Copyright 2025-Present Datadog, Inc.
 */
#pragma once

/**
 *
 * The context model represents the local system or component and includes:
 * 1. Function pointers for supported evaluators (string, numeric, unsigned
 * numeric).
 * 2. Parameters associated with each evaluator type (string, numeric, unsigned
 * numeric).
 * 3. Function pointers for supported actions.
 *
 * Since evaluation may occur at various stages of the application lifecycle,
 * the model should be updated promptly as new parameters become available.
 *
 */

#include "action.h"
#include "error_codes.h"
#include "evaluator_types.h"

#include <limits.h>
#include <stdint.h>

#define PLCS_STR_NOT_SET NULL
#define PLCS_NUM_NOT_SET LONG_MAX
#define PLCS_UNUM_NOT_SET ULONG_MAX

/**
 * @brief Maximum length (excluding the null terminator) of a policy
 * description retained by the eval context. Longer descriptions are
 * truncated when stored via plcs_eval_ctx_set_policy_description.
 */
#define PLCS_POLICY_DESCRIPTION_MAX_LEN 255

/**
 * @brief opaque struct, defined in src/eval_ctx.h
 */
typedef struct plcs_eval_ctx plcs_eval_ctx;

/**
 * @brief A 128-bit UUID, represented as two 64-bit halves.
 */
typedef struct {
  uint64_t hi;
  uint64_t lo;
} plcs_uuid;

/**
 * @brief Registers a new string evaluator.
 *
 * @param func_ptr a function pointer to the string evaluator function.
 * @param ix a number representing one of the plcs_string_evaluators (up to
 * STR_EVAL__COUNT)
 * @return int on success DD_ESUCCESS(0), on error > 0 plcs_errors code
 */
plcs_errors
plcs_eval_ctx_register_str_evaluator(plcs_string_evaluator_function_ptr func_ptr, plcs_string_evaluators ix);

/**
 * @brief Registers a new numeric evaluator.
 *
 * @param func_ptr a function pointer to the numeric evaluator function.
 * @param ix a number representing one of the plcs_numeric_evaluators (up to
 * NUM_EVAL__COUNT)
 * @return int on success DD_ESUCCESS(0), on error > 0 plcs_errors code
 */
plcs_errors
plcs_eval_ctx_register_num_evaluator(plcs_numeric_evaluator_function_ptr func_ptr, plcs_numeric_evaluators ix);

/**
 * @brief Registers a new unsigned numeric evaluator.
 *
 * @param func_ptr a function pointer to the unsigned numeric evaluator
 * function.
 * @param ix a number representing one of the UnumericEvaluators (up to
 * UNUM_EVAL__COUNT)
 * @return int on success DD_ESUCCESS(0), on error > 0 plcs_errors code
 */
plcs_errors
plcs_eval_ctx_register_unum_evaluator(plcs_unumeric_evaluator_function_ptr func_ptr, plcs_numeric_evaluators ix);

/**
 * @brief Registers a new action.
 *
 * @param action a function pointer to the action function.
 * @param ix a number representing one of the plcs_actions (up to ACTIONS__COUNT)
 * @return int on success DD_ESUCCESS(0), on error > 0 plcs_errors code
 */
plcs_errors plcs_eval_ctx_register_action(plcs_action_function_ptr action, plcs_actions ix);

/**
 * @brief Sets the local string parameter for string evaluator ix
 *
 * @param value a string representing the local value to evaluate against *NOTE*
 * caller responsible for freeing.
 * @param ix a number representing one of the plcs_string_evaluators (up to
 * STR_EVAL__COUNT)
 * @return int on success DD_ESUCCESS(0), on error > 0 plcs_errors code
 */
plcs_errors plcs_eval_ctx_set_str_eval_param(plcs_string_evaluators ix, const char *value);

/**
 * @brief Sets the local numeric parameter for numeric evaluator ix
 *
 * @param ix a number representing one of the plcs_numeric_evaluators (up to
 * NUM_EVAL__COUNT)
 * @param value the local numeric value to evaluate against
 * @return int on success DD_ESUCCESS(0), on error > 0 plcs_errors code
 */
plcs_errors plcs_eval_ctx_set_num_eval_param(plcs_numeric_evaluators ix, const long value);

/**
 * @brief Sets the local unsigned numeric parameter for unsigned numeric
 * evaluator ix
 *
 * @param ix a number representing one of the UnumericEvaluators (up to
 * UNUM_EVAL__COUNT)
 * @param value the local unsigned numeric value to evaluate against
 * @return int on success DD_ESUCCESS(0), on error > 0 plcs_errors code
 */
plcs_errors plcs_eval_ctx_set_unum_eval_param(plcs_numeric_evaluators ix, const unsigned long value);

/**
 * @brief Sets the id of the policy currently loaded in the context.
 * @param id the policy id as a plcs_uuid.
 * @return int on success DD_ESUCCESS(0), on error > 0 plcs_errors code
 */
plcs_errors plcs_eval_ctx_set_policy_id(plcs_uuid id);

/**
 * @brief Sets the version of the policy currently loaded in the context.
 * @param version the policy version as an int64_t.
 * @return int on success DD_ESUCCESS(0), on error > 0 plcs_errors code
 */
plcs_errors plcs_eval_ctx_set_policy_version(int64_t version);

/**
 * @brief Sets the description of the policy currently loaded in the context.
 * The context copies the string into its own fixed-size storage (truncating
 * to PLCS_POLICY_DESCRIPTION_MAX_LEN bytes if needed), so the caller retains
 * ownership of and may free/invalidate `description` immediately after this
 * call returns.
 * @param description a string representing the policy description.
 * @return int on success DD_ESUCCESS(0), on error > 0 plcs_errors code
 */
plcs_errors plcs_eval_ctx_set_policy_description(const char *description);

/**
 * @brief Get a function pointer for a specific evaluator id.
 * @param id plcs_string_evaluators enum.
 * @return NULL on error (also sets ctx.error) or the function ptr.
 */
plcs_string_evaluator_function_ptr plcs_eval_ctx_get_string_evaluator(plcs_string_evaluators id);

/**
 * @brief Get the parameter for a specific evaluator id.
 * @param id plcs_string_evaluators enum.
 * @return STR_NOT_SET on error (also sets ctx.error) or a const char* value.
 */
const char *plcs_eval_ctx_get_string_param(plcs_string_evaluators id);

/**
 * @brief Get a function pointer for a specific evaluator id.
 * @param id plcs_numeric_evaluators enum.
 * @return NULL on error (also sets ctx.error) or the function ptr.
 */
plcs_numeric_evaluator_function_ptr plcs_eval_ctx_get_numeric_evaluator(plcs_numeric_evaluators id);

/**
 * @brief Get the parameter for a specific evaluator id.
 * @param id plcs_numeric_evaluators enum.
 * @return NUM_NOT_SET on error (also sets ctx.error) or a long value.
 */
long plcs_eval_ctx_get_numeric_param(plcs_numeric_evaluators id);

/**
 * @brief Get a function pointer for a specific evaluator id.
 * @param id plcs_numeric_evaluators enum.
 * @return NULL on error (also sets ctx.error) or the function ptr.
 */
plcs_unumeric_evaluator_function_ptr plcs_eval_ctx_get_unumeric_evaluator(plcs_numeric_evaluators id);

/**
 * @brief Get the parameter for a specific evaluator id.
 * @param id plcs_numeric_evaluators enum.
 * @return NUM_NOT_SET on error (also sets ctx.error) or an unsigned long value.
 */
unsigned long plcs_eval_ctx_get_unumeric_param(plcs_numeric_evaluators id);

/**
 * @brief Get a function pointer for a specific action id.
 * @param ix plcs_actions enum.
 * @return NULL on error (also sets ctx.error) or the function ptr.
 */
plcs_action_function_ptr plcs_eval_ctx_get_action(plcs_actions ix);

/**
 * @brief Get the id of the policy currently loaded in the context.
 * @return the policy id as a plcs_uuid.
 */
plcs_uuid plcs_eval_ctx_get_policy_id(void);

/**
 * @brief Get the version of the policy currently loaded in the context.
 * @return the policy version as an int64_t.
 */
int64_t plcs_eval_ctx_get_policy_version(void);

/**
 * @brief Get the description of the policy currently loaded in the context.
 * @return the policy description as a const char*, owned by the context and
 * valid until the next call to plcs_eval_ctx_set_policy_description or
 * plcs_eval_ctx_reset. May be truncated to PLCS_POLICY_DESCRIPTION_MAX_LEN
 * bytes.
 */
const char *plcs_eval_ctx_get_policy_description(void);

/**
 * @brief An accessor for the last error
 * @return the last error as a plcs_errors enum, NOTE: This will RESET the error!
 */
plcs_errors plcs_eval_ctx_get_last_error(void);

/**
 * @brief An accessor for the last error that doesnt reset the error!
 * @return the last error as policies_error enum, NOTE: This WILL NOT RESET the error!
 */
plcs_errors plcs_eval_ctx_peek_last_error(void);

/**
 * @brief Initializes the context model.
 * This function sets all evaluator function pointers to NULL and parameters to
 * their 'not set' values
 * @return a plcs_errors enum indicating success or failure
 */
plcs_errors plcs_eval_ctx_init(void);

/**
 * @brief Resets all the error codes to DD_ESUCESS, must be run after each evaluation.
 * @param
 */
void plcs_eval_ctx_reset(void);

/**
 * @brief Gets the error code for a string evaluator
 * @param ix A plcs_string_evaluators enum ID
 * @return The error code as a plcs_errors enum
 */
plcs_errors plcs_eval_ctx_get_str_eval_error(plcs_string_evaluators ix);

/**
 * @brief Gets the error code for a numeric evaluator
 * @param ix A plcs_numeric_evaluators enum ID
 * @return The error code as a plcs_errors enum
 */
plcs_errors plcs_eval_ctx_get_num_eval_error(plcs_numeric_evaluators ix);

/**
 * @brief Gets the error code for an unsigned numeric evaluator
 * @param ix A plcs_numeric_evaluators enum ID
 * @return The error code as a plcs_errors enum
 */
plcs_errors plcs_eval_ctx_get_unum_eval_error(plcs_numeric_evaluators ix);

/**
 * @def REGISTER_STR_EVAL_PARAM
 * @brief Registers a string evaluation parameter for policy evaluation context.
 *
 * This macro is used to define and register a string parameter that can be
 * evaluated within the policy evaluation context.
 */
#define REGISTER_STR_EVAL_PARAM(id, function_ptr, param)                                                               \
  do {                                                                                                                 \
    plcs_eval_ctx_register_str_evaluator(function_ptr, id);                                                            \
    plcs_eval_ctx_set_str_eval_param(id, param);                                                                       \
  } while (0)
