/*
 * Unless explicitly stated otherwise all files in this repository are licensed
 * under the Apache 2.0 License. This product includes software developed at
 * Datadog (https://www.datadoghq.com/).
 *
 * Copyright 2025-Present Datadog, Inc.
 */

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

typedef struct policy_buffer {
  void *data;
  size_t size;
} policy_buffer;

static dd_wls_Action_ref_t create_action(flatcc_builder_t *builder, dd_wls_ActionId_enum_t id) {
  flatbuffers_string_vec_ref_t values = flatbuffers_string_vec_create(builder, NULL, 0);
  flatbuffers_string_ref_t description = flatbuffers_string_create_str(builder, "test action");
  return values == 0 || description == 0 ? 0 : dd_wls_Action_create(builder, id, description, values);
}

static dd_wls_Policy_ref_t create_policy(
    flatcc_builder_t *builder,
    dd_wls_NodeTypeWrapper_ref_t rules,
    dd_wls_Action_ref_t actions[],
    size_t actions_len
) {
  dd_wls_Action_vec_ref_t actions_vec = dd_wls_Action_vec_create(builder, actions, actions_len);
  flatbuffers_string_ref_t description = flatbuffers_string_create_str(builder, "test policy");
  if (actions_vec == 0 || description == 0 || dd_wls_Policy_start(builder) != 0 ||
      dd_wls_Policy_description_add(builder, description) != 0 ||
      (rules != 0 && dd_wls_Policy_rules_add(builder, rules) != 0) ||
      dd_wls_Policy_actions_add(builder, actions_vec) != 0) {
    return 0;
  }
  return dd_wls_Policy_end(builder);
}

static dd_wls_NodeTypeWrapper_ref_t build_linux_rule(flatcc_builder_t *builder) {
  flatbuffers_string_ref_t description = flatbuffers_string_create_str(builder, "linux rule");
  flatbuffers_string_ref_t expected = flatbuffers_string_create_str(builder, "linux");
  dd_wls_StrEvaluator_ref_t evaluator =
      dd_wls_StrEvaluator_create(builder, dd_wls_StringEvaluators_OS, dd_wls_CmpTypeSTR_CMP_EXACT, expected);
  if (description == 0 || expected == 0 || evaluator == 0) {
    return 0;
  }
  dd_wls_EvaluatorNode_ref_t node =
      dd_wls_EvaluatorNode_create(builder, description, dd_wls_EvaluatorType_as_StrEvaluator(evaluator));
  return node == 0 ? 0 : dd_wls_NodeTypeWrapper_create(builder, dd_wls_NodeType_as_EvaluatorNode(node));
}

static policy_buffer build_invalid_actions_then_valid_deny(void) {
  policy_buffer result = {0};
  flatcc_builder_t builder;
  if (flatcc_builder_init(&builder) != 0) {
    return result;
  }

  dd_wls_Action_ref_t invalid_action = create_action(&builder, (dd_wls_ActionId_enum_t)-1);
  dd_wls_Action_ref_t invalid_actions[] = {invalid_action};
  dd_wls_Policy_ref_t invalid_policy = create_policy(&builder, 0, invalid_actions, 1);
  dd_wls_Policy_ref_t second_invalid_policy = create_policy(&builder, 0, invalid_actions, 1);
  dd_wls_NodeTypeWrapper_ref_t valid_rule = build_linux_rule(&builder);
  dd_wls_Action_ref_t unsupported_action = create_action(&builder, dd_wls_ActionId_ACTIONS_COUNT);
  dd_wls_Action_ref_t deny_action = create_action(&builder, dd_wls_ActionId_INJECT_DENY);
  dd_wls_Action_ref_t supported_and_unsupported_actions[] = {unsupported_action, deny_action};
  dd_wls_Policy_ref_t valid_policy = create_policy(&builder, valid_rule, supported_and_unsupported_actions, 2);
  if (invalid_action == 0 || invalid_policy == 0 || second_invalid_policy == 0 || valid_rule == 0 || deny_action == 0 ||
      unsupported_action == 0 || valid_policy == 0) {
    goto cleanup;
  }

  dd_wls_Policy_ref_t policy_refs[] = {invalid_policy, second_invalid_policy, valid_policy};
  dd_wls_Policy_vec_ref_t policies = dd_wls_Policy_vec_create(&builder, policy_refs, 3);
  if (policies == 0 || dd_wls_Policies_start_as_root(&builder) != 0 ||
      dd_wls_Policies_policies_add(&builder, policies) != 0 || dd_wls_Policies_end_as_root(&builder) == 0) {
    goto cleanup;
  }

  result.data = flatcc_builder_finalize_buffer(&builder, &result.size);
  if (result.data != NULL && dd_wls_Policies_verify_as_root(result.data, result.size) != 0) {
    flatcc_builder_free(result.data);
    result = (policy_buffer){0};
  }

cleanup:
  flatcc_builder_clear(&builder);
  return result;
}

static size_t observed_calls;
static plcs_evaluation_result observed_result;

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
  ++observed_calls;
  observed_result = result;
  return PLCS_ESUCCESS;
}

UTEST(policy_validation, unknown_actions_are_skipped_and_valid_action_runs) {
  policy_buffer buffer = build_invalid_actions_then_valid_deny();
  ASSERT_TRUE(buffer.data != NULL);

  (void)plcs_eval_ctx_init();
  plcs_eval_ctx_reset();
  observed_calls = 0;
  observed_result = PLCS_EVAL_RESULT_FALSE;
  ASSERT_EQ((int)plcs_eval_ctx_set_str_eval_param(PLCS_STR_EVAL_OS, "linux"), PLCS_ESUCCESS);
  ASSERT_EQ((int)plcs_eval_ctx_register_action(observing_action, PLCS_ACTION_INJECT_DENY), PLCS_ESUCCESS);

  ASSERT_EQ(dd_wls_Policies_verify_as_root(buffer.data, buffer.size), 0);
  int result = plcs_evaluate_buffer(buffer.data, buffer.size);
  flatcc_builder_free(buffer.data);

  ASSERT_EQ(result, PLCS_ESUCCESS);
  ASSERT_EQ(observed_calls, (size_t)1);
  ASSERT_EQ((int)observed_result, PLCS_EVAL_RESULT_TRUE);
}
