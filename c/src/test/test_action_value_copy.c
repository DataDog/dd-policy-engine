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
#include <string.h>

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

  flatbuffers_string_ref_t value = flatbuffers_string_create_str(&builder, "value");
  flatbuffers_string_vec_ref_t values = flatbuffers_string_vec_create(&builder, &value, 1);
  flatbuffers_string_ref_t action_description = flatbuffers_string_create_str(&builder, "immutable action values");
  dd_wls_Action_ref_t action = dd_wls_Action_create(&builder, dd_wls_ActionId_INJECT_DENY, action_description, values);
  dd_wls_Action_vec_ref_t actions = dd_wls_Action_vec_create(&builder, &action, 1);
  flatbuffers_string_ref_t policy_description = flatbuffers_string_create_str(&builder, "const action values");

  if (value == 0 || values == 0 || action_description == 0 || action == 0 || actions == 0 || policy_description == 0 ||
      dd_wls_Policy_start(&builder) != 0 || dd_wls_Policy_description_add(&builder, policy_description) != 0 ||
      dd_wls_Policy_actions_add(&builder, actions) != 0) {
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

static int received_expected_value;

static plcs_errors inspect_action_values(
    plcs_evaluation_result result,
    const char *values[],
    size_t value_count,
    const char *description,
    int action_id
) {
  (void)result;
  (void)description;
  (void)action_id;
  received_expected_value = value_count == 1 && values[0] != NULL && strcmp(values[0], "value") == 0;
  return PLCS_ESUCCESS;
}

UTEST(policy_actions, callbacks_receive_immutable_values) {
  policy_buffer buffer = build_action_policy();
  ASSERT_TRUE(buffer.data != NULL);

  (void)plcs_eval_ctx_init();
  plcs_eval_ctx_reset();
  received_expected_value = 0;
  ASSERT_EQ((int)plcs_eval_ctx_register_action(inspect_action_values, PLCS_ACTION_INJECT_DENY), PLCS_ESUCCESS);

  ASSERT_EQ((int)plcs_evaluate_buffer(buffer.data, buffer.size), PLCS_ESUCCESS);
  flatcc_builder_free(buffer.data);

  ASSERT_TRUE(received_expected_value);
}
