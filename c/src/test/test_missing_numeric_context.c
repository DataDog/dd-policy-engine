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
#include <stdint.h>

typedef struct policy_buffer {
  void *data;
  size_t size;
} policy_buffer;

static dd_wls_NodeTypeWrapper_ref_t build_numeric_rule(flatcc_builder_t *builder, int is_unsigned) {
  flatbuffers_string_ref_t description = flatbuffers_string_create_str(builder, "numeric context");
  dd_wls_EvaluatorNode_ref_t node = 0;

  if (is_unsigned) {
    dd_wls_UNumEvaluator_ref_t evaluator = dd_wls_UNumEvaluator_create(
        builder, dd_wls_NumericEvaluators_JAVA_HEAP, dd_wls_CmpTypeNUM_CMP_LTE, UINT64_C(1024)
    );
    if (evaluator != 0) {
      node = dd_wls_EvaluatorNode_create(builder, description, dd_wls_EvaluatorType_as_UNumEvaluator(evaluator));
    }
  } else {
    dd_wls_NumEvaluator_ref_t evaluator = dd_wls_NumEvaluator_create(
        builder, dd_wls_NumericEvaluators_LIBC_VERSION_MAJOR, dd_wls_CmpTypeNUM_CMP_LTE, INT64_C(2)
    );
    if (evaluator != 0) {
      node = dd_wls_EvaluatorNode_create(builder, description, dd_wls_EvaluatorType_as_NumEvaluator(evaluator));
    }
  }

  return description == 0 || node == 0 ? 0
                                       : dd_wls_NodeTypeWrapper_create(builder, dd_wls_NodeType_as_EvaluatorNode(node));
}

static policy_buffer build_numeric_policy(int is_unsigned) {
  policy_buffer result = {0};
  flatcc_builder_t builder;
  if (flatcc_builder_init(&builder) != 0) {
    return result;
  }

  dd_wls_NodeTypeWrapper_ref_t rule = build_numeric_rule(&builder, is_unsigned);
  flatbuffers_string_vec_ref_t values = flatbuffers_string_vec_create(&builder, NULL, 0);
  flatbuffers_string_ref_t action_description = flatbuffers_string_create_str(&builder, "observe result");
  dd_wls_Action_ref_t action = dd_wls_Action_create(&builder, dd_wls_ActionId_INJECT_DENY, action_description, values);
  dd_wls_Action_vec_ref_t actions = dd_wls_Action_vec_create(&builder, &action, 1);
  flatbuffers_string_ref_t policy_description = flatbuffers_string_create_str(&builder, "missing numeric context");

  if (rule == 0 || values == 0 || action_description == 0 || action == 0 || actions == 0 || policy_description == 0 ||
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
  if (result.data != NULL && dd_wls_Policies_verify_as_root(result.data, result.size) != 0) {
    flatcc_builder_free(result.data);
    result = (policy_buffer){0};
  }

cleanup:
  flatcc_builder_clear(&builder);
  return result;
}

static plcs_evaluation_result observed_result;

static plcs_errors observing_action(
    plcs_evaluation_result result,
    char *values[],
    size_t value_count,
    const char *description,
    int action_id,
    plcs_uuid policy_id,
    int64_t policy_version,
    const char *policy_description,
    const char *rule_description,
    const plcs_matched_rule *matched_rules,
    size_t matched_rules_len
) {
  (void)values;
  (void)value_count;
  (void)description;
  (void)action_id;
  (void)policy_id;
  (void)policy_version;
  (void)policy_description;
  (void)rule_description;
  (void)matched_rules;
  (void)matched_rules_len;
  observed_result = result;
  return PLCS_ESUCCESS;
}

static void assert_missing_numeric_context_abstains(int is_unsigned, int *utest_result) {
  policy_buffer buffer = build_numeric_policy(is_unsigned);
  ASSERT_TRUE(buffer.data != NULL);

  (void)plcs_eval_ctx_init();
  plcs_eval_ctx_reset();
  observed_result = PLCS_EVAL_RESULT_FALSE;
  ASSERT_EQ((int)plcs_eval_ctx_register_action(observing_action, PLCS_ACTION_INJECT_DENY), PLCS_ESUCCESS);

  ASSERT_EQ((int)plcs_evaluate_buffer(buffer.data, buffer.size), PLCS_ESUCCESS);
  flatcc_builder_free(buffer.data);

  ASSERT_EQ((int)observed_result, PLCS_EVAL_RESULT_ABSTAIN);
}

UTEST(policy_evaluator, missing_signed_numeric_context_abstains) {
  assert_missing_numeric_context_abstains(0, utest_result);
}

UTEST(policy_evaluator, missing_unsigned_numeric_context_abstains) {
  assert_missing_numeric_context_abstains(1, utest_result);
}
