/*
 * Unless explicitly stated otherwise all files in this repository are licensed
 * under the Apache 2.0 License. This product includes software developed at
 * Datadog (https://www.datadoghq.com/).
 *
 * Copyright 2025-Present Datadog, Inc.
 */
#include <dd/policies/action.h>
#include <dd/policies/error_codes.h>
#include <dd/policies/eval_ctx.h>
#include <dd/policies/evaluator_types.h>
#include <dd/policies/policies.h>
#include <stdio.h>
#include <stdlib.h>

#include "hardcoded-header-reader.h"

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
  (void)policy_description;
  printf("Action: DENY\n");
  printf("Description: '%s' (id: %d)\n", description, action_id);
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
  (void)policy_description;
  printf("Action: ALLOW\n");
  printf("Description: '%s' (id: %d)\n", description, action_id);
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

int main() {
  // Initialize policy evaluation context
  if (plcs_eval_ctx_init() != PLCS_ESUCCESS) {
    fprintf(stderr, "Failed to initialize evaluation context\n");
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
  printf("Evaluating policies from hardcoded buffer (hardcoded_policies)...\n");
  plcs_errors res = plcs_evaluate_buffer(hardcoded_policies, hardcoded_policies_len);

  if (res != PLCS_ESUCCESS) {
    fprintf(stderr, "Failed to evaluate policy buffer\n");
    return EXIT_FAILURE;
  }

  printf("Policy evaluation completed successfully\n");
  return EXIT_SUCCESS;
}
