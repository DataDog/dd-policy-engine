/*
 * Unless explicitly stated otherwise all files in this repository are licensed
 * under the Apache 2.0 License. This product includes software developed at
 * Datadog (https://www.datadoghq.com/).
 *
 * Copyright 2025-Present Datadog, Inc.
 * ----
 * Opt-in regression repros. Enable with DD_POLICY_BUILD_REPROS=ON.
 *
 * These tests assert the intended behavior and therefore fail on a revision
 * where the corresponding bug is present.
 */
#define _GNU_SOURCE

#include "utest/utest.h"

#include <dd/policies/error_codes.h>
#include <dd/policies/eval_ctx.h>
#include <dd/policies/policies.h>

#include "actions_builder.h"
#include "evaluators_builder.h"
#include "flatbuffers_common_builder.h"
#include "nodes_builder.h"
#include "policy_builder.h"
#include "policy_verifier.h"

#include <stddef.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

typedef struct repro_buffer {
  void *data;
  size_t size;
} repro_buffer;

static repro_buffer finish_policy(
    flatcc_builder_t *builder,
    dd_wls_NodeTypeWrapper_ref_t rules,
    const dd_wls_ActionId_enum_t *ids,
    size_t action_count
) {
  repro_buffer result = {0};
  dd_wls_Action_ref_t *actions = calloc(action_count, sizeof(*actions));
  if (action_count != 0 && actions == NULL) {
    goto cleanup;
  }

  flatbuffers_string_vec_ref_t values_vec = flatbuffers_string_vec_create(builder, NULL, 0);
  if (values_vec == 0) {
    goto cleanup;
  }

  for (size_t i = 0; i < action_count; ++i) {
    flatbuffers_string_ref_t description = flatbuffers_string_create_str(builder, "repro action");
    if (description == 0) {
      goto cleanup;
    }
    actions[i] = dd_wls_Action_create(builder, ids[i], description, values_vec);
    if (actions[i] == 0) {
      goto cleanup;
    }
  }

  dd_wls_Action_vec_ref_t actions_vec = dd_wls_Action_vec_create(builder, actions, action_count);
  if (actions_vec == 0) {
    goto cleanup;
  }

  flatbuffers_string_ref_t policy_description = flatbuffers_string_create_str(builder, "repro policy");
  flatbuffers_string_ref_t policy_id = flatbuffers_string_create_str(builder, "repro");
  if (policy_description == 0 || policy_id == 0) {
    goto cleanup;
  }

  if (dd_wls_Policy_start(builder) != 0 || dd_wls_Policy_description_add(builder, policy_description) != 0 ||
      (rules != 0 && dd_wls_Policy_rules_add(builder, rules) != 0) ||
      dd_wls_Policy_actions_add(builder, actions_vec) != 0 || dd_wls_Policy_id_add(builder, policy_id) != 0) {
    goto cleanup;
  }
  dd_wls_Policy_ref_t policy = dd_wls_Policy_end(builder);
  if (policy == 0) {
    goto cleanup;
  }

  dd_wls_Policy_ref_t policies[] = {policy};
  dd_wls_Policy_vec_ref_t policies_vec = dd_wls_Policy_vec_create(builder, policies, 1);
  if (policies_vec == 0 || dd_wls_Policies_start_as_root(builder) != 0 ||
      dd_wls_Policies_policies_add(builder, policies_vec) != 0 || dd_wls_Policies_end_as_root(builder) == 0) {
    goto cleanup;
  }

  result.data = flatcc_builder_finalize_buffer(builder, &result.size);
  if (result.data != NULL && dd_wls_Policies_verify_as_root(result.data, result.size) != 0) {
    flatcc_builder_free(result.data);
    result = (repro_buffer){0};
  }

cleanup:
  free(actions);
  flatcc_builder_clear(builder);
  return result;
}

static repro_buffer build_action_policy(const dd_wls_ActionId_enum_t *ids, size_t action_count) {
  repro_buffer result = {0};
  flatcc_builder_t builder;
  if (flatcc_builder_init(&builder) != 0) {
    return result;
  }
  return finish_policy(&builder, 0, ids, action_count);
}

static dd_wls_NodeTypeWrapper_ref_t build_numeric_rule(
    flatcc_builder_t *builder,
    dd_wls_NumericEvaluators_enum_t id,
    dd_wls_CmpTypeNUM_enum_t comparator,
    int64_t value
) {
  flatbuffers_string_ref_t description = flatbuffers_string_create_str(builder, "numeric repro");
  dd_wls_NumEvaluator_ref_t evaluator = dd_wls_NumEvaluator_create(builder, id, comparator, value);
  if (description == 0 || evaluator == 0) {
    return 0;
  }
  dd_wls_EvaluatorNode_ref_t node =
      dd_wls_EvaluatorNode_create(builder, description, dd_wls_EvaluatorType_as_NumEvaluator(evaluator));
  return node == 0 ? 0 : dd_wls_NodeTypeWrapper_create(builder, dd_wls_NodeType_as_EvaluatorNode(node));
}

static dd_wls_NodeTypeWrapper_ref_t
build_string_rule(flatcc_builder_t *builder, dd_wls_StringEvaluators_enum_t id, const char *value) {
  flatbuffers_string_ref_t description = flatbuffers_string_create_str(builder, "string repro");
  flatbuffers_string_ref_t expected = flatbuffers_string_create_str(builder, value);
  dd_wls_StrEvaluator_ref_t evaluator = dd_wls_StrEvaluator_create(builder, id, dd_wls_CmpTypeSTR_CMP_EXACT, expected);
  if (description == 0 || expected == 0 || evaluator == 0) {
    return 0;
  }
  dd_wls_EvaluatorNode_ref_t node =
      dd_wls_EvaluatorNode_create(builder, description, dd_wls_EvaluatorType_as_StrEvaluator(evaluator));
  return node == 0 ? 0 : dd_wls_NodeTypeWrapper_create(builder, dd_wls_NodeType_as_EvaluatorNode(node));
}

static repro_buffer
build_numeric_policy(dd_wls_NumericEvaluators_enum_t id, dd_wls_CmpTypeNUM_enum_t comparator, int64_t value) {
  repro_buffer result = {0};
  flatcc_builder_t builder;
  if (flatcc_builder_init(&builder) != 0) {
    return result;
  }
  dd_wls_NodeTypeWrapper_ref_t rule = build_numeric_rule(&builder, id, comparator, value);
  if (rule == 0) {
    flatcc_builder_clear(&builder);
    return result;
  }
  const dd_wls_ActionId_enum_t action = dd_wls_ActionId_INJECT_DENY;
  return finish_policy(&builder, rule, &action, 1);
}

static repro_buffer build_invalid_then_valid_policy(void) {
  repro_buffer result = {0};
  flatcc_builder_t builder;
  if (flatcc_builder_init(&builder) != 0) {
    return result;
  }

  dd_wls_NodeTypeWrapper_ref_t children[] = {
      build_string_rule(&builder, dd_wls_StringEvaluators_STR_EVAL_COUNT, "invalid"),
      build_string_rule(&builder, dd_wls_StringEvaluators_OS, "linux"),
  };
  if (children[0] == 0 || children[1] == 0) {
    flatcc_builder_clear(&builder);
    return result;
  }

  dd_wls_NodeTypeWrapper_vec_ref_t children_vec = dd_wls_NodeTypeWrapper_vec_create(&builder, children, 2);
  if (children_vec == 0) {
    flatcc_builder_clear(&builder);
    return result;
  }
  flatbuffers_string_ref_t description = flatbuffers_string_create_str(&builder, "invalid OR valid");
  if (description == 0) {
    flatcc_builder_clear(&builder);
    return result;
  }
  dd_wls_CompositeNode_ref_t composite =
      dd_wls_CompositeNode_create(&builder, description, dd_wls_BoolOperation_BOOL_OR, children_vec);
  dd_wls_NodeTypeWrapper_ref_t rule =
      composite == 0 ? 0 : dd_wls_NodeTypeWrapper_create(&builder, dd_wls_NodeType_as_CompositeNode(composite));
  if (rule == 0) {
    flatcc_builder_clear(&builder);
    return result;
  }

  const dd_wls_ActionId_enum_t action = dd_wls_ActionId_INJECT_DENY;
  return finish_policy(&builder, rule, &action, 1);
}

static plcs_errors failing_action(
    plcs_evaluation_result result,
    char *values[],
    size_t value_count,
    const char *description,
    int action_id
) {
  (void)result;
  (void)values;
  (void)value_count;
  (void)description;
  (void)action_id;
  return PLCS_EACTIONS_EVAL;
}

static size_t observed_action_calls;
static plcs_evaluation_result observed_evaluation_result;

static plcs_errors observing_action(
    plcs_evaluation_result result,
    char *values[],
    size_t value_count,
    const char *description,
    int action_id
) {
  (void)values;
  (void)value_count;
  (void)description;
  (void)action_id;
  observed_action_calls++;
  observed_evaluation_result = result;
  return PLCS_ESUCCESS;
}

UTEST(repro_actions, later_success_does_not_erase_failure) {
  const dd_wls_ActionId_enum_t ids[] = {dd_wls_ActionId_INJECT_DENY, dd_wls_ActionId_INJECT_ALLOW};
  repro_buffer buffer = build_action_policy(ids, 2);
  ASSERT_TRUE(buffer.data != NULL);

  (void)plcs_eval_ctx_init();
  plcs_eval_ctx_reset();
  ASSERT_EQ(plcs_eval_ctx_register_action(failing_action, PLCS_ACTION_INJECT_DENY), PLCS_ESUCCESS);
  ASSERT_EQ(plcs_eval_ctx_register_action(observing_action, PLCS_ACTION_INJECT_ALLOW), PLCS_ESUCCESS);

  plcs_errors result = plcs_evaluate_buffer(buffer.data, buffer.size);
  flatcc_builder_free(buffer.data);

  /* A failed action must make the policy evaluation fail even if a later action succeeds. */
  ASSERT_EQ(result, PLCS_EACTIONS_EVAL);
}

UTEST(repro_actions, unregistered_action_is_an_error) {
  const dd_wls_ActionId_enum_t action = dd_wls_ActionId_INJECT_DENY;
  repro_buffer buffer = build_action_policy(&action, 1);
  ASSERT_TRUE(buffer.data != NULL);

  (void)plcs_eval_ctx_init();
  plcs_eval_ctx_reset();

  plcs_errors result = plcs_evaluate_buffer(buffer.data, buffer.size);
  flatcc_builder_free(buffer.data);

  /* Silently skipping a missing enforcement action makes the policy a no-op. */
  ASSERT_EQ(result, PLCS_EACTIONS_EVAL);
}

UTEST(repro_evaluator, missing_numeric_context_abstains) {
  repro_buffer buffer = build_numeric_policy(dd_wls_NumericEvaluators_LIBC_VERSION_MAJOR, dd_wls_CmpTypeNUM_CMP_LT, 1);
  ASSERT_TRUE(buffer.data != NULL);

  (void)plcs_eval_ctx_init();
  plcs_eval_ctx_reset();
  observed_action_calls = 0;
  observed_evaluation_result = PLCS_EVAL_RESULT_FALSE;
  ASSERT_EQ(plcs_eval_ctx_register_action(observing_action, PLCS_ACTION_INJECT_DENY), PLCS_ESUCCESS);

  ASSERT_EQ(plcs_evaluate_buffer(buffer.data, buffer.size), PLCS_ESUCCESS);
  flatcc_builder_free(buffer.data);

  ASSERT_EQ(observed_action_calls, (size_t)1);
  ASSERT_EQ(observed_evaluation_result, PLCS_EVAL_RESULT_ABSTAIN);
}

UTEST(repro_evaluator, invalid_evaluator_does_not_poison_valid_sibling) {
  repro_buffer buffer = build_invalid_then_valid_policy();
  ASSERT_TRUE(buffer.data != NULL);

  (void)plcs_eval_ctx_init();
  plcs_eval_ctx_reset();
  observed_action_calls = 0;
  observed_evaluation_result = PLCS_EVAL_RESULT_FALSE;
  ASSERT_EQ(plcs_eval_ctx_set_str_eval_param(PLCS_STR_EVAL_OS, "linux"), PLCS_ESUCCESS);
  ASSERT_EQ(plcs_eval_ctx_register_action(observing_action, PLCS_ACTION_INJECT_DENY), PLCS_ESUCCESS);

  ASSERT_EQ(plcs_evaluate_buffer(buffer.data, buffer.size), PLCS_ESUCCESS);
  flatcc_builder_free(buffer.data);

  ASSERT_EQ(observed_action_calls, (size_t)1);
  ASSERT_EQ(observed_evaluation_result, PLCS_EVAL_RESULT_TRUE);
}

UTEST(repro_eval_ctx, set_string_parameter_copies_caller_buffer) {
  (void)plcs_eval_ctx_init();
  plcs_eval_ctx_reset();

  char *value = strdup("FOO=bar");
  ASSERT_TRUE(value != NULL);
  ASSERT_EQ(plcs_eval_ctx_set_str_eval_param(PLCS_STR_EVAL_PROCESS_ENVAR, value), PLCS_ESUCCESS);

  const char *stored = plcs_eval_ctx_get_string_param(PLCS_STR_EVAL_PROCESS_ENVAR);
  ASSERT_TRUE(stored != NULL);
  /* A context-owned copy cannot be the caller's allocation. */
  ASSERT_TRUE(stored != value);
  ASSERT_EQ(strcmp(stored, "FOO=bar"), 0);
  free(value);
}
