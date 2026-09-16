/**
 * Unless explicitly stated otherwise all files in this repository are licensed
 * under the Apache 2.0 License. This product includes software developed at
 * Datadog (https://www.datadoghq.com/).
 *
 * Copyright 2025-Present Datadog, Inc.
 * ----
 * @file demo_policies.c
 * @brief Small example showing how to register evaluators and actions, then evaluate a policy buffer.
 *
 * @details
 * This example:
 *   - Implements two action handlers: `ACTION_INJECT_DENY` and `ACTION_INJECT_ALLOW`
 *     (they match the public `plcs_action_function_ptr` signature and just print what they receive).
 *   - Implements one string evaluator: `EVALUATOR_RUNTIME_LANGUAGE`
 *     (it matches `plcs_string_evaluator_function_ptr` and always returns `PLCS_EVAL_RESULT_TRUE` for demo).
 *   - Reads a policy file into memory, initializes the evaluation context, wires everything up,
 *     and calls `plcs_evaluate_buffer(...)`.
 *
 * **Flow**
 *   1. Read policy bytes from disk (`read_file_contents`).
 *   2. Initialize the eval context (`plcs_eval_ctx_init`).
 *   3. Register evaluator + params:
 *        - `REGISTER_STR_EVAL_PARAM(STR_EVAL_RUNTIME_LANGUAGE, EVALUATOR_RUNTIME_LANGUAGE, "jvm")`
 *        - `plcs_eval_ctx_set_str_eval_param(STR_EVAL_PROCESS_EXE_FULL_PATH, "/some/path/to/runtime")`
 *   4. Register action handlers:
 *        - `plcs_eval_ctx_register_action(ACTION_INJECT_DENY, INJECT_DENY)`
 *        - `plcs_eval_ctx_register_action(ACTION_INJECT_ALLOW, INJECT_ALLOW)`
 *   5. Evaluate the buffer (`plcs_evaluate_buffer`), print results, exit.
 *
 * @usage
 * @code
 *   make examples
 *   ./basic-reader path/to/policy.fb
 * @endcode
 *
 * @note This is a minimal, print-only demo. Real handlers would enforce policy
 *       (deny/allow, set env vars, etc.). Error handling is kept simple.
 *
 * @see policies/action.h
 * @see policies/error_codes.h
 * @see policies/eval_ctx.h
 * @see policies/policies.h
 * @see policies/evaluator_types.h
 */

#include "file_oper.h"

#include <dd/policies/action.h>
#include <dd/policies/error_codes.h>
#include <dd/policies/eval_ctx.h>
#include <dd/policies/evaluator_types.h>
#include <dd/policies/policies.h>
#include <stdio.h>
#include <stdlib.h>

// Demo action handler for INJECT_DENY action
plcs_errors ACTION_INJECT_DENY(
    plcs_evaluation_result res,
    char *values[],
    size_t value_len,
    const char *description,
    int action_id,
    plcs_uuid policy_id,
    int64_t policy_version,
    const char *policy_description
) {
  (void)policy_id;
  (void)policy_version;
  printf("Action: DENY\n");
  printf("Description: '%s' (id: %d)\n", description, action_id);
  printf("Matched rule: '%s'\n", policy_description);
  printf("Result: %s\n", res == PLCS_EVAL_RESULT_FALSE ? "false" : res == PLCS_EVAL_RESULT_TRUE ? "true" : "dont-care");

  for (size_t ix = 0; ix < value_len; ++ix) {
    printf("Value[%zu]: '%s'\n", ix, values[ix]);
  }
  return PLCS_ESUCCESS;
}

// Demo action handler for INJECT_ALLOW action
plcs_errors ACTION_INJECT_ALLOW(
    plcs_evaluation_result res,
    char *values[],
    size_t value_len,
    const char *description,
    int action_id,
    plcs_uuid policy_id,
    int64_t policy_version,
    const char *policy_description
) {
  (void)policy_id;
  (void)policy_version;
  printf("Action: ALLOW\n");
  printf("Description: '%s' (id: %d)\n", description, action_id);
  printf("Matched rule: '%s'\n", policy_description);
  printf("Result: %s\n", res == PLCS_EVAL_RESULT_FALSE ? "false" : res == PLCS_EVAL_RESULT_TRUE ? "true" : "dont-care");

  for (size_t ix = 0; ix < value_len; ++ix) {
    printf("Value[%zu]: '%s'\n", ix, values[ix]);
  }
  return PLCS_ESUCCESS;
}

// Demo evaluator for runtime language detection
plcs_evaluation_result EVALUATOR_RUNTIME_LANGUAGE(
    const char *policy,
    const plcs_string_comparator cmp,
    const char *ctx,
    const char *description,
    plcs_string_evaluators eval_id
) {
  printf("Evaluator: Runtime Language\n");
  if (policy && ctx && description) {
    printf("Policy: '%s'\n", policy);
    printf("Context: '%s'\n", ctx);
    printf("Comparator: %d\n", cmp);
    printf("Description: '%s' (id: %d)\n", description, eval_id);
  }
  return PLCS_EVAL_RESULT_TRUE;
}

int main(int argc, char *argv[]) {
  if (argc < 2) {
    fprintf(stderr, "Usage: %s <path_to_policy_file>\n", argv[0]);
    return EXIT_FAILURE;
  }

  // Read policy file
  size_t buffer_size;
  uint8_t *buffer = read_file_contents(argv[1], &buffer_size);
  if (!buffer) {
    return EXIT_FAILURE;
  }
  printf("Successfully read %zu bytes from '%s'\n", buffer_size, argv[1]);

  // Initialize policy evaluation context
  if (plcs_eval_ctx_init() != PLCS_ESUCCESS) {
    fprintf(stderr, "Failed to initialize evaluation context\n");
    free(buffer);
    return EXIT_FAILURE;
  }
  printf("Evaluation context initialized\n");

  // Register evaluators and set parameters
  REGISTER_STR_EVAL_PARAM(PLCS_STR_EVAL_RUNTIME_LANGUAGE, EVALUATOR_RUNTIME_LANGUAGE, "jvm");
  plcs_eval_ctx_set_str_eval_param(PLCS_STR_EVAL_PROCESS_EXE_FULL_PATH, "/some/path/to/runtime");

  // Register action handlers
  plcs_eval_ctx_register_action(ACTION_INJECT_DENY, PLCS_ACTION_INJECT_DENY);
  plcs_eval_ctx_register_action(ACTION_INJECT_ALLOW, PLCS_ACTION_INJECT_ALLOW);

  // Evaluate policy
  printf("Evaluating policies...\n");
  plcs_errors res = plcs_evaluate_buffer(buffer, buffer_size);
  free(buffer);

  if (res != PLCS_ESUCCESS) {
    fprintf(stderr, "Failed to evaluate policy buffer\n");
    return EXIT_FAILURE;
  }

  printf("Policy evaluation completed successfully\n");
  return EXIT_SUCCESS;
}
