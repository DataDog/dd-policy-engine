/*
 * Unless explicitly stated otherwise all files in this repository are licensed
 * under the Apache 2.0 License. This product includes software developed at
 * Datadog (https://www.datadoghq.com/).
 *
 * Copyright 2025-Present Datadog, Inc.
 */
#pragma once

#include <dd/policies/action.h>
#include <dd/policies/error_codes.h>
#include <dd/policies/eval_ctx.h>
#include "wire/evaluator_types.h"

/**
 * @brief String evaluator entry structure.
 */
typedef struct string_evaluator_entry {
  /**< Function pointer to the string evaluator function */
  plcs_string_evaluator_function_ptr function_ptr;
  /**< The value to evaluate against, can be NULL if not set */
  const char *value;
  /**< Error code if the evaluator is not registered or */
  /**< fails !NEEDS TO BE RESET EVERY EVALUATION! */
  plcs_errors error;
} string_evaluator_entry;

typedef struct numeric_evaluator_entry {
  /**< Function pointer to the numeric evaluator function */
  plcs_numeric_evaluator_function_ptr function_ptr;
  /**< The value to evaluate against, can be NUM_NOT_SET if not set */
  long value;
  /**< Error code if the evaluator is not registered or */
  /**< fails !NEEDS TO BE RESET EVERY EVALUATION! */
  plcs_errors error;
} numeric_evaluator_entry;

typedef struct unumeric_evaluator_entry {
  /**< Function pointer to the unumeric evaluator function */
  plcs_unumeric_evaluator_function_ptr function_ptr;
  /**< The value to evaluate against, can be UNUM_NOT_SET if not set */
  unsigned long value;
  /**< Error code if the evaluator is not registered or */
  /**< fails !NEEDS TO BE RESET EVERY EVALUATION! */
  plcs_errors error;
} unumeric_evaluator_entry;

typedef struct action_entry {
  /**< Function pointer to the action function */
  plcs_action_function_ptr function_ptr;
  /**< Error code if the action is not registered or fails !NEEDS */
  plcs_errors error;
} action_entry;

/**
 * @brief The eval_ctx structure represents the evaluators and actions available
 * in the system. It includes:
 * - String evaluators for evaluating string-based policies.
 * - Numeric evaluators for evaluating numeric policies.
 * - Unsigned numeric evaluators for evaluating unsigned numeric policies.
 * - Action function pointers for executing actions based on evaluation results.
 *
 * The eval_ctx is initialized with default evaluators and actions,
 * for debugging purposes you can register custom evaluators and actions.
 * The model is designed to be flexible and extensible, allowing for easy
 * addition of new evaluators and actions.
 *
 * @note The eval_ctx is a mandatory component for policy evaluation.
 * It is used to store the current state of the system and to evaluate policies
 * based on that state. The model is updated as new parameters become available,
 * allowing for dynamic evaluation of policies.
 *
 * For Strings:
 * - Evaluators are registered with a function pointer and an index representing
 * the evaluator type.
 * - Parameters are set using the index and the value to evaluate against.
 * - Parameters containing `STR_NOT_SET` (`NULL`) values are considered as *not
 * set*.
 *
 * For Numerics:
 * - Evaluators are registered with a function pointer and an index representing
 * the evaluator type.
 * - Parameters are set using the index and the numeric value to evaluate
 * against.
 * - Parameters containing `NUM_NOT_SET` (`LONG_MAX`) are considered as *not
 * set*.
 *
 * For Unsigned Numerics:
 * - Evaluators are registered with a function pointer and an index representing
 * the evaluator type.
 * - Parameters are set using the index and the unsigned numeric value to
 * evaluate against.
 * - Parameters containing `UNUM_NOT_SET` (`ULONG_MAX`) are considered as *not
 * set*.
 *
 */
typedef struct plcs_eval_ctx {
  /**< EVALUATORS */
  /**< (a simple map evaluator id (enum):func_ptr) */
  string_evaluator_entry string_evaluators[PLCS_STR_EVAL__COUNT];

  /**< (a simple map evaluator id (enum):func_ptr) */
  numeric_evaluator_entry numeric_evaluators[PLCS_NUM_EVAL__COUNT];

  /**< (a simple map evaluator id (enum):func_ptr) */
  unumeric_evaluator_entry unumeric_evaluators[PLCS_NUM_EVAL__COUNT];

  /**< (a simple map action id (enum):func_ptr) */
  action_entry actions[PLCS_ACTIONS__COUNT];

  /**< TODO: consider implementing this as a stack to preserve history of errors */
  plcs_errors error;

  /**
   * @brief Description of the top-level OR child that returned TRUE during the
   * most recent policy evaluation. NULL if no OR-level match occurred. Set by
   * composite_evaluator at depth 0; read via plcs_eval_ctx_get_matched_description().
   *
   * For dd-requirements-converter policies, the tree looks something like:
   *
   *   Policy("All requirements")
   *   └── (||)                          <- depth 0, root; composite_evaluator checks depth == 0
   *       ├── (&&) "Ignore npm CLI"     <- top-level OR child; description captured on match
   *       │   └── [PROCESS_ARGV_1 ~= *\/npm]
   *       ├── (&&) "Ignore yarn"
   *       │   └── [PROCESS_ARGV_1 ~= *\/yarn]
   *       └── (&&) "glibc arm >= 2.24"
   *           ├── [MACHINE_ARCHITECTURE == arm]
   *           ├── [LIBC_FLAVOR == glibc]
   *           └── (||)                  <- depth 2; never triggers capture
   *               └── ...
   */
  const char *matched_description;

  /**
   * @brief Description of the most recent EvaluatorNode that returned TRUE.
   * Overwritten on each TRUE leaf result during evaluation. Used by
   * capture_matched_description to append specifics to a composite rule
   * description (e.g. "Ignore the yarn CLI" + "Argument matching: *\/yarn.js"
   * → "Ignore the yarn CLI (Argument matching: *\/yarn.js)").
   * Reset to NULL at the start of each evaluate_policy call.
   */
  const char *matched_evaluator_description;

} plcs_eval_ctx;

/**
 * @brief Used to set an error in the evaluation context
 * @param error plcs_errors enum
 */
void plcs_eval_ctx_set_error(plcs_errors error);

/**
 * @brief Sets an error code for an action
 * @param ix An action ID from plcs_actions enum
 * @param error plcs_errors enum
 */
void plcs_eval_ctx_set_action_error(plcs_actions ix, plcs_errors error);

/**
 * @brief Sets an error code for a string evaluator
 * @param ix A plcs_string_evaluators enum ID
 * @param error plcs_errors enum
 */
void plcs_eval_ctx_set_str_eval_error(plcs_string_evaluators ix, plcs_errors error);

/**
 * @brief Sets an error code for a numeric evaluator
 * @param id A plcs_numeric_evaluators enum ID
 * @param error plcs_errors enum
 */
void plcs_eval_ctx_set_num_eval_error(plcs_numeric_evaluators id, plcs_errors error);

/**
 * @brief Sets an error code for an unsigned numeric evaluator
 * @param ix A plcs_numeric_evaluators enum ID
 * @param error plcs_errors enum
 */
void plcs_eval_ctx_set_unum_eval_error(plcs_numeric_evaluators ix, plcs_errors error);

/**
 * @brief Records the description of the matched OR child for the current policy evaluation.
 *
 * Internal — called by composite_evaluator when a top-level OR child returns TRUE.
 * The pointer is stored verbatim (no copy) and must remain valid until the next
 * plcs_eval_ctx_reset() or plcs_evaluate_buffer() call — in practice, a string
 * inside the FlatBuffer being evaluated.
 *
 * @param desc Description string, or NULL to clear.
 */
void plcs_eval_ctx_set_matched_description(const char *desc);

/**
 * @brief Combines a composite node description with the most-recently-matched
 * evaluator description and stores the result as the matched description.
 *
 * Formats as "<rule_desc> (<eval_desc>)", e.g.
 * "Ignore the yarn CLI (Argument matching: *\/yarn.js)".
 * The result is written into a static internal buffer valid until the next
 * evaluate_policy call.
 */
void plcs_eval_ctx_set_matched_description_combined(const char *rule_desc, const char *eval_desc);

/**
 * @brief Records the description of the most recent EvaluatorNode that returned TRUE.
 * Called by node_evaluator on a TRUE result.
 */
void plcs_eval_ctx_set_matched_evaluator_description(const char *desc);

/**
 * @brief Returns the description of the most recent EvaluatorNode that returned TRUE,
 * or NULL if none has fired yet in this evaluation cycle.
 */
const char *plcs_eval_ctx_get_matched_evaluator_description(void);
