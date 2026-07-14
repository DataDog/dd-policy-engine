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
#include "flatbuffers_common_builder.h"
#include "nodes_builder.h"
#include "policy_builder.h"
#include "policy_verifier.h"

#include <stdbool.h>
#include <stddef.h>

#define TEST_EXCESSIVE_NESTING_DEPTH 65

typedef struct policy_buffer {
  void *data;
  size_t size;
} policy_buffer;

plcs_errors evaluate_policy(dd_wls_Policy_table_t policy);

static policy_buffer build_invalid_boolean_policy(size_t nesting_depth, bool verify) {
  policy_buffer result = {0};
  flatcc_builder_t builder;
  if (flatcc_builder_init(&builder) != 0) {
    return result;
  }

  flatbuffers_string_ref_t rule_description = flatbuffers_string_create_str(&builder, "invalid boolean operator");
  dd_wls_NodeTypeWrapper_vec_ref_t children = dd_wls_NodeTypeWrapper_vec_create(&builder, NULL, 0);
  dd_wls_CompositeNode_ref_t composite =
      dd_wls_CompositeNode_create(&builder, rule_description, dd_wls_BoolOperation_BOOL_COUNT, children);
  dd_wls_NodeTypeWrapper_ref_t rule =
      composite == 0 ? 0 : dd_wls_NodeTypeWrapper_create(&builder, dd_wls_NodeType_as_CompositeNode(composite));

  if (rule_description == 0 || children == 0 || composite == 0 || rule == 0) {
    goto cleanup;
  }

  for (size_t depth = 0; depth < nesting_depth; ++depth) {
    children = dd_wls_NodeTypeWrapper_vec_create(&builder, &rule, 1);
    composite = dd_wls_CompositeNode_create(&builder, rule_description, dd_wls_BoolOperation_BOOL_AND, children);
    rule = composite == 0 ? 0 : dd_wls_NodeTypeWrapper_create(&builder, dd_wls_NodeType_as_CompositeNode(composite));
    if (children == 0 || composite == 0 || rule == 0) {
      goto cleanup;
    }
  }

  flatbuffers_string_vec_ref_t values = flatbuffers_string_vec_create(&builder, NULL, 0);
  flatbuffers_string_ref_t action_description = flatbuffers_string_create_str(&builder, "must not run");
  dd_wls_Action_ref_t action = dd_wls_Action_create(&builder, dd_wls_ActionId_INJECT_DENY, action_description, values);
  dd_wls_Action_vec_ref_t actions = dd_wls_Action_vec_create(&builder, &action, 1);
  flatbuffers_string_ref_t policy_description = flatbuffers_string_create_str(&builder, "malformed policy");

  if (values == 0 || action_description == 0 || action == 0 || actions == 0 || policy_description == 0 ||
      dd_wls_Policy_start(&builder) != 0 || dd_wls_Policy_description_add(&builder, policy_description) != 0 ||
      dd_wls_Policy_rules_add(&builder, rule) != 0 || dd_wls_Policy_actions_add(&builder, actions) != 0) {
    goto cleanup;
  }

  dd_wls_Policy_ref_t policy = dd_wls_Policy_end(&builder);
  dd_wls_Policy_vec_ref_t policies = dd_wls_Policy_vec_create(&builder, &policy, 1);
  if (policy == 0 || policies == 0 || dd_wls_Policies_start_as_root(&builder) != 0 ||
      dd_wls_Policies_policies_add(&builder, policies) != 0 || dd_wls_Policies_end_as_root(&builder) == 0) {
    goto cleanup;
  }

  result.data = flatcc_builder_finalize_buffer(&builder, &result.size);
  if (verify && result.data != NULL && dd_wls_Policies_verify_as_root(result.data, result.size) != 0) {
    flatcc_builder_free(result.data);
    result = (policy_buffer){0};
  }

cleanup:
  flatcc_builder_clear(&builder);
  return result;
}

static size_t observed_calls;

static plcs_errors observing_action(
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
  ++observed_calls;
  return PLCS_ESUCCESS;
}

UTEST(policy_validation, invalid_boolean_operator_is_rejected) {
  policy_buffer buffer = build_invalid_boolean_policy(0, true);
  ASSERT_TRUE(buffer.data != NULL);

  (void)plcs_eval_ctx_init();
  plcs_eval_ctx_reset();
  observed_calls = 0;
  ASSERT_EQ((int)plcs_eval_ctx_register_action(observing_action, PLCS_ACTION_INJECT_DENY), PLCS_ESUCCESS);

  ASSERT_EQ(dd_wls_Policies_verify_as_root(buffer.data, buffer.size), 0);
  int result = plcs_evaluate_buffer(buffer.data, buffer.size);
  flatcc_builder_free(buffer.data);

  ASSERT_EQ(observed_calls, (size_t)0);
  ASSERT_EQ(result, PLCS_EUNKNOWN_CMP);
}

UTEST(policy_validation, excessive_nesting_is_rejected_before_actions) {
  // The public buffer path rejects this depth in FlatCC verification. Exercise
  // evaluate_policy directly to preserve its defense-in-depth guarantee.
  policy_buffer buffer = build_invalid_boolean_policy(TEST_EXCESSIVE_NESTING_DEPTH, false);
  ASSERT_TRUE(buffer.data != NULL);

  dd_wls_Policies_table_t root = dd_wls_Policies_as_root(buffer.data);
  dd_wls_Policy_vec_t policies = dd_wls_Policies_policies(root);
  ASSERT_TRUE(policies != NULL);
  ASSERT_EQ(dd_wls_Policy_vec_len(policies), (size_t)1);

  (void)plcs_eval_ctx_init();
  plcs_eval_ctx_reset();
  observed_calls = 0;
  ASSERT_EQ((int)plcs_eval_ctx_register_action(observing_action, PLCS_ACTION_INJECT_DENY), PLCS_ESUCCESS);

  int result = evaluate_policy(dd_wls_Policy_vec_at(policies, 0));
  flatcc_builder_free(buffer.data);

  ASSERT_EQ(observed_calls, (size_t)0);
  ASSERT_EQ(result, PLCS_EUNKNOWN_CMP);
}
