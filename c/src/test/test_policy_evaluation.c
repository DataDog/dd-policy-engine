/*
 * Unless explicitly stated otherwise all files in this repository are licensed
 * under the Apache 2.0 License. This product includes software developed at
 * Datadog (https://www.datadoghq.com/).
 *
 * Copyright 2025-Present Datadog, Inc.
 * ----
 * End-to-end policy evaluation regression tests.
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
#include <stdlib.h>
#include <string.h>

typedef struct policy_buffer {
  void *data;
  size_t size;
} policy_buffer;

static policy_buffer
finalize_policies(flatcc_builder_t *builder, const dd_wls_Policy_ref_t *policies, size_t policy_count) {
  policy_buffer result = {0};
  dd_wls_Policy_vec_ref_t policies_vec = dd_wls_Policy_vec_create(builder, policies, policy_count);
  if (policies_vec == 0) {
    return result;
  }
  if (dd_wls_Policies_start_as_root(builder) != 0) {
    return result;
  }
  if (dd_wls_Policies_policies_add(builder, policies_vec) != 0) {
    return result;
  }
  if (dd_wls_Policies_end_as_root(builder) == 0) {
    return result;
  }

  result.data = flatcc_builder_finalize_buffer(builder, &result.size);
  if (result.data != NULL && dd_wls_Policies_verify_as_root(result.data, result.size) != 0) {
    flatcc_builder_free(result.data);
    result = (policy_buffer){0};
  }
  return result;
}

static dd_wls_Action_ref_t
create_action(flatcc_builder_t *builder, dd_wls_ActionId_enum_t id, flatbuffers_string_vec_ref_t values) {
  flatbuffers_string_ref_t description = flatbuffers_string_create_str(builder, "test action");
  return description == 0 ? 0 : dd_wls_Action_create(builder, id, description, values);
}

static dd_wls_Policy_ref_t create_policy(
    flatcc_builder_t *builder,
    dd_wls_NodeTypeWrapper_ref_t rules,
    const dd_wls_Action_ref_t *actions,
    size_t action_count
) {
  dd_wls_Action_vec_ref_t actions_vec = dd_wls_Action_vec_create(builder, actions, action_count);
  flatbuffers_string_ref_t description = flatbuffers_string_create_str(builder, "test policy");
  flatbuffers_string_ref_t id = flatbuffers_string_create_str(builder, "test");
  if (actions_vec == 0 || description == 0 || id == 0) {
    return 0;
  }
  if (dd_wls_Policy_start(builder) != 0) {
    return 0;
  }
  if (dd_wls_Policy_description_add(builder, description) != 0) {
    return 0;
  }
  if (rules != 0 && dd_wls_Policy_rules_add(builder, rules) != 0) {
    return 0;
  }
  if (dd_wls_Policy_actions_add(builder, actions_vec) != 0) {
    return 0;
  }
  if (dd_wls_Policy_id_add(builder, id) != 0) {
    return 0;
  }
  return dd_wls_Policy_end(builder);
}

static policy_buffer finish_single_policy(
    flatcc_builder_t *builder,
    dd_wls_NodeTypeWrapper_ref_t rules,
    const dd_wls_ActionId_enum_t *ids,
    size_t action_count,
    size_t value_count
) {
  policy_buffer result = {0};
  dd_wls_Action_ref_t *actions = calloc(action_count, sizeof(*actions));
  flatbuffers_string_ref_t *values = calloc(value_count, sizeof(*values));
  if ((action_count != 0 && actions == NULL) || (value_count != 0 && values == NULL)) {
    goto cleanup;
  }

  for (size_t i = 0; i < value_count; ++i) {
    values[i] = flatbuffers_string_create_str(builder, "value");
    if (values[i] == 0) {
      goto cleanup;
    }
  }
  flatbuffers_string_vec_ref_t values_vec = flatbuffers_string_vec_create(builder, values, value_count);
  if (values_vec == 0) {
    goto cleanup;
  }

  for (size_t i = 0; i < action_count; ++i) {
    actions[i] = create_action(builder, ids[i], values_vec);
    if (actions[i] == 0) {
      goto cleanup;
    }
  }

  dd_wls_Policy_ref_t policy = create_policy(builder, rules, actions, action_count);
  if (policy != 0) {
    result = finalize_policies(builder, &policy, 1);
  }

cleanup:
  free(values);
  free(actions);
  flatcc_builder_clear(builder);
  return result;
}

static policy_buffer build_action_policy(const dd_wls_ActionId_enum_t *ids, size_t action_count, size_t value_count) {
  policy_buffer result = {0};
  flatcc_builder_t builder;
  if (flatcc_builder_init(&builder) != 0) {
    return result;
  }
  return finish_single_policy(&builder, 0, ids, action_count, value_count);
}

static dd_wls_NodeTypeWrapper_ref_t
build_string_rule(flatcc_builder_t *builder, dd_wls_StringEvaluators_enum_t id, const char *value) {
  flatbuffers_string_ref_t description = flatbuffers_string_create_str(builder, "string test");
  flatbuffers_string_ref_t expected = flatbuffers_string_create_str(builder, value);
  dd_wls_StrEvaluator_ref_t evaluator = dd_wls_StrEvaluator_create(builder, id, dd_wls_CmpTypeSTR_CMP_EXACT, expected);
  if (description == 0 || expected == 0 || evaluator == 0) {
    return 0;
  }
  dd_wls_EvaluatorNode_ref_t node =
      dd_wls_EvaluatorNode_create(builder, description, dd_wls_EvaluatorType_as_StrEvaluator(evaluator));
  return node == 0 ? 0 : dd_wls_NodeTypeWrapper_create(builder, dd_wls_NodeType_as_EvaluatorNode(node));
}

static dd_wls_NodeTypeWrapper_ref_t build_numeric_rule(
    flatcc_builder_t *builder,
    dd_wls_NumericEvaluators_enum_t id,
    dd_wls_CmpTypeNUM_enum_t comparator,
    int64_t value
) {
  flatbuffers_string_ref_t description = flatbuffers_string_create_str(builder, "numeric test");
  dd_wls_NumEvaluator_ref_t evaluator = dd_wls_NumEvaluator_create(builder, id, comparator, value);
  if (description == 0 || evaluator == 0) {
    return 0;
  }
  dd_wls_EvaluatorNode_ref_t node =
      dd_wls_EvaluatorNode_create(builder, description, dd_wls_EvaluatorType_as_NumEvaluator(evaluator));
  return node == 0 ? 0 : dd_wls_NodeTypeWrapper_create(builder, dd_wls_NodeType_as_EvaluatorNode(node));
}

static dd_wls_NodeTypeWrapper_ref_t build_unumeric_rule(
    flatcc_builder_t *builder,
    dd_wls_NumericEvaluators_enum_t id,
    dd_wls_CmpTypeNUM_enum_t comparator,
    uint64_t value
) {
  flatbuffers_string_ref_t description = flatbuffers_string_create_str(builder, "unsigned numeric test");
  dd_wls_UNumEvaluator_ref_t evaluator = dd_wls_UNumEvaluator_create(builder, id, comparator, value);
  if (description == 0 || evaluator == 0) {
    return 0;
  }
  dd_wls_EvaluatorNode_ref_t node =
      dd_wls_EvaluatorNode_create(builder, description, dd_wls_EvaluatorType_as_UNumEvaluator(evaluator));
  return node == 0 ? 0 : dd_wls_NodeTypeWrapper_create(builder, dd_wls_NodeType_as_EvaluatorNode(node));
}

static policy_buffer build_numeric_policy(int is_unsigned) {
  policy_buffer result = {0};
  flatcc_builder_t builder;
  if (flatcc_builder_init(&builder) != 0) {
    return result;
  }

  dd_wls_NodeTypeWrapper_ref_t rule;
  if (is_unsigned) {
    rule = build_unumeric_rule(&builder, dd_wls_NumericEvaluators_JAVA_HEAP, dd_wls_CmpTypeNUM_CMP_LTE, UINT64_C(1024));
  } else {
    rule = build_numeric_rule(
        &builder, dd_wls_NumericEvaluators_LIBC_VERSION_MAJOR, dd_wls_CmpTypeNUM_CMP_LTE, INT64_C(2)
    );
  }
  if (rule == 0) {
    flatcc_builder_clear(&builder);
    return result;
  }
  const dd_wls_ActionId_enum_t action = dd_wls_ActionId_INJECT_DENY;
  return finish_single_policy(&builder, rule, &action, 1, 0);
}

static policy_buffer build_invalid_action_then_valid_deny(void) {
  policy_buffer result = {0};
  flatcc_builder_t builder;
  if (flatcc_builder_init(&builder) != 0) {
    return result;
  }

  flatbuffers_string_vec_ref_t values = flatbuffers_string_vec_create(&builder, NULL, 0);
  dd_wls_Action_ref_t invalid_action = create_action(&builder, (dd_wls_ActionId_enum_t)-1, values);
  dd_wls_Policy_ref_t invalid_policy = create_policy(&builder, 0, &invalid_action, 1);

  dd_wls_NodeTypeWrapper_ref_t valid_rule = build_string_rule(&builder, dd_wls_StringEvaluators_OS, "linux");
  dd_wls_Action_ref_t deny_action = create_action(&builder, dd_wls_ActionId_INJECT_DENY, values);
  dd_wls_Policy_ref_t valid_policy = create_policy(&builder, valid_rule, &deny_action, 1);
  if (values == 0 || invalid_action == 0 || invalid_policy == 0 || valid_rule == 0 || deny_action == 0 ||
      valid_policy == 0) {
    goto cleanup;
  }

  dd_wls_Policy_ref_t policies[] = {invalid_policy, valid_policy};
  result = finalize_policies(&builder, policies, 2);

cleanup:
  flatcc_builder_clear(&builder);
  return result;
}

static policy_buffer build_invalid_boolean_policy(void) {
  policy_buffer result = {0};
  flatcc_builder_t builder;
  if (flatcc_builder_init(&builder) != 0) {
    return result;
  }

  flatbuffers_string_ref_t description = flatbuffers_string_create_str(&builder, "invalid boolean operator");
  dd_wls_NodeTypeWrapper_vec_ref_t children = dd_wls_NodeTypeWrapper_vec_create(&builder, NULL, 0);
  dd_wls_CompositeNode_ref_t composite =
      dd_wls_CompositeNode_create(&builder, description, dd_wls_BoolOperation_BOOL_COUNT, children);
  dd_wls_NodeTypeWrapper_ref_t rule =
      composite == 0 ? 0 : dd_wls_NodeTypeWrapper_create(&builder, dd_wls_NodeType_as_CompositeNode(composite));
  if (description == 0 || children == 0 || rule == 0) {
    flatcc_builder_clear(&builder);
    return result;
  }

  const dd_wls_ActionId_enum_t action = dd_wls_ActionId_INJECT_DENY;
  return finish_single_policy(&builder, rule, &action, 1, 0);
}

static size_t observed_action_calls;
static int observed_evaluation_result;

static void reset_observations(void) {
  observed_action_calls = 0;
  observed_evaluation_result = PLCS_EVAL_RESULT_FALSE;
}

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

static plcs_errors mutating_action(
    plcs_evaluation_result result,
    char *values[],
    size_t value_count,
    const char *description,
    int action_id
) {
  (void)result;
  (void)description;
  (void)action_id;
  if (value_count != 0) {
    values[0][0] = 'X';
  }
  return PLCS_ESUCCESS;
}

UTEST(policy_actions, later_success_does_not_erase_failure) {
  const dd_wls_ActionId_enum_t ids[] = {dd_wls_ActionId_INJECT_DENY, dd_wls_ActionId_INJECT_ALLOW};
  policy_buffer buffer = build_action_policy(ids, 2, 0);
  ASSERT_TRUE(buffer.data != NULL);

  (void)plcs_eval_ctx_init();
  plcs_eval_ctx_reset();
  reset_observations();
  ASSERT_EQ((int)plcs_eval_ctx_register_action(failing_action, PLCS_ACTION_INJECT_DENY), PLCS_ESUCCESS);
  ASSERT_EQ((int)plcs_eval_ctx_register_action(observing_action, PLCS_ACTION_INJECT_ALLOW), PLCS_ESUCCESS);

  int result = plcs_evaluate_buffer(buffer.data, buffer.size);
  flatcc_builder_free(buffer.data);

  ASSERT_EQ(observed_action_calls, (size_t)1);
  ASSERT_EQ(result, PLCS_EACTIONS_EVAL);
}

UTEST(policy_actions, const_policy_buffer_is_not_mutated_by_action_values) {
  const dd_wls_ActionId_enum_t action = dd_wls_ActionId_INJECT_DENY;
  policy_buffer buffer = build_action_policy(&action, 1, 1);
  ASSERT_TRUE(buffer.data != NULL);
  void *original = malloc(buffer.size);
  ASSERT_TRUE(original != NULL);
  memcpy(original, buffer.data, buffer.size);

  (void)plcs_eval_ctx_init();
  plcs_eval_ctx_reset();
  ASSERT_EQ((int)plcs_eval_ctx_register_action(mutating_action, PLCS_ACTION_INJECT_DENY), PLCS_ESUCCESS);

  ASSERT_EQ((int)plcs_evaluate_buffer(buffer.data, buffer.size), PLCS_ESUCCESS);
  int unchanged = memcmp(original, buffer.data, buffer.size) == 0;
  free(original);
  flatcc_builder_free(buffer.data);

  /* plcs_evaluate_buffer accepts a const buffer; action values must not provide a mutable alias into it. */
  ASSERT_TRUE(unchanged);
}

UTEST(policy_evaluator, missing_signed_numeric_context_abstains) {
  policy_buffer buffer = build_numeric_policy(0);
  ASSERT_TRUE(buffer.data != NULL);

  (void)plcs_eval_ctx_init();
  plcs_eval_ctx_reset();
  reset_observations();
  ASSERT_EQ((int)plcs_eval_ctx_register_action(observing_action, PLCS_ACTION_INJECT_DENY), PLCS_ESUCCESS);

  ASSERT_EQ((int)plcs_evaluate_buffer(buffer.data, buffer.size), PLCS_ESUCCESS);
  flatcc_builder_free(buffer.data);

  /* CMP_LTE 2 models an unsupported libc range; missing input must not match it. */
  ASSERT_EQ(observed_action_calls, (size_t)1);
  ASSERT_EQ(observed_evaluation_result, PLCS_EVAL_RESULT_ABSTAIN);
}

UTEST(policy_evaluator, missing_unsigned_numeric_context_abstains) {
  policy_buffer buffer = build_numeric_policy(1);
  ASSERT_TRUE(buffer.data != NULL);

  (void)plcs_eval_ctx_init();
  plcs_eval_ctx_reset();
  reset_observations();
  ASSERT_EQ((int)plcs_eval_ctx_register_action(observing_action, PLCS_ACTION_INJECT_DENY), PLCS_ESUCCESS);

  ASSERT_EQ((int)plcs_evaluate_buffer(buffer.data, buffer.size), PLCS_ESUCCESS);
  flatcc_builder_free(buffer.data);

  /* The unsigned evaluator must also abstain for its reset sentinel. */
  ASSERT_EQ(observed_action_calls, (size_t)1);
  ASSERT_EQ(observed_evaluation_result, PLCS_EVAL_RESULT_ABSTAIN);
}

UTEST(policy_validation, invalid_action_cannot_disable_later_deny_policy) {
  policy_buffer buffer = build_invalid_action_then_valid_deny();
  ASSERT_TRUE(buffer.data != NULL);

  (void)plcs_eval_ctx_init();
  plcs_eval_ctx_reset();
  reset_observations();
  ASSERT_EQ((int)plcs_eval_ctx_set_str_eval_param(PLCS_STR_EVAL_OS, "linux"), PLCS_ESUCCESS);
  ASSERT_EQ((int)plcs_eval_ctx_register_action(observing_action, PLCS_ACTION_INJECT_DENY), PLCS_ESUCCESS);

  /* The buffer passes the generated verifier despite the out-of-range ActionId. */
  ASSERT_EQ(dd_wls_Policies_verify_as_root(buffer.data, buffer.size), 0);
  int result = plcs_evaluate_buffer(buffer.data, buffer.size);
  flatcc_builder_free(buffer.data);

  ASSERT_EQ(result, PLCS_EIX_OVERFLOW);
  ASSERT_EQ(observed_action_calls, (size_t)1);
  ASSERT_EQ(observed_evaluation_result, PLCS_EVAL_RESULT_TRUE);
}

UTEST(policy_validation, invalid_boolean_operator_is_rejected) {
  policy_buffer buffer = build_invalid_boolean_policy();
  ASSERT_TRUE(buffer.data != NULL);

  (void)plcs_eval_ctx_init();
  plcs_eval_ctx_reset();
  reset_observations();
  ASSERT_EQ((int)plcs_eval_ctx_register_action(observing_action, PLCS_ACTION_INJECT_DENY), PLCS_ESUCCESS);

  ASSERT_EQ(dd_wls_Policies_verify_as_root(buffer.data, buffer.size), 0);
  int result = plcs_evaluate_buffer(buffer.data, buffer.size);
  flatcc_builder_free(buffer.data);

  /* BOOL_COUNT passes FlatBuffers verification but is not an executable operator. */
  ASSERT_TRUE(
      observed_evaluation_result >= PLCS_EVAL_RESULT_TRUE && observed_evaluation_result <= PLCS_EVAL_RESULT_ABSTAIN
  );
  ASSERT_EQ(observed_action_calls, (size_t)0);
  ASSERT_EQ(result, PLCS_EUNKNOWN_CMP);
}
