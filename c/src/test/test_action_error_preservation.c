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
#include "policy_builder.h"
#include "policy_verifier.h"

#include <stddef.h>

typedef struct policy_buffer {
  void *data;
  size_t size;
} policy_buffer;

static policy_buffer build_action_policy(void) {
  policy_buffer result = {0};
  flatcc_builder_t builder;
  if (flatcc_builder_init(&builder) != 0) {
    return result;
  }

  flatbuffers_string_vec_ref_t values = flatbuffers_string_vec_create(&builder, NULL, 0);
  flatbuffers_string_ref_t first_description = flatbuffers_string_create_str(&builder, "failing action");
  flatbuffers_string_ref_t second_description = flatbuffers_string_create_str(&builder, "successful action");
  dd_wls_Action_ref_t actions[] = {
      dd_wls_Action_create(&builder, dd_wls_ActionId_INJECT_DENY, first_description, values),
      dd_wls_Action_create(&builder, dd_wls_ActionId_INJECT_ALLOW, second_description, values),
  };
  dd_wls_Action_vec_ref_t actions_vec = dd_wls_Action_vec_create(&builder, actions, 2);
  flatbuffers_string_ref_t policy_description = flatbuffers_string_create_str(&builder, "action ordering");

  if (values == 0 || first_description == 0 || second_description == 0 || actions[0] == 0 || actions[1] == 0 ||
      actions_vec == 0 || policy_description == 0 || dd_wls_Policy_start(&builder) != 0 ||
      dd_wls_Policy_description_add(&builder, policy_description) != 0 ||
      dd_wls_Policy_actions_add(&builder, actions_vec) != 0) {
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

static plcs_errors successful_action(
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
  return PLCS_ESUCCESS;
}

UTEST(policy_actions, later_success_does_not_erase_failure) {
  policy_buffer buffer = build_action_policy();
  ASSERT_TRUE(buffer.data != NULL);

  (void)plcs_eval_ctx_init();
  plcs_eval_ctx_reset();
  ASSERT_EQ((int)plcs_eval_ctx_register_action(failing_action, PLCS_ACTION_INJECT_DENY), PLCS_ESUCCESS);
  ASSERT_EQ((int)plcs_eval_ctx_register_action(successful_action, PLCS_ACTION_INJECT_ALLOW), PLCS_ESUCCESS);

  int result = plcs_evaluate_buffer(buffer.data, buffer.size);
  flatcc_builder_free(buffer.data);

  ASSERT_EQ(result, PLCS_EACTIONS_EVAL);
}
