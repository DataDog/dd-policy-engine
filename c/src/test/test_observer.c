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
#include <dd/policies/observer.h>
#include <dd/policies/policies.h>

#include "actions_builder.h"
#include "evaluators_builder.h"
#include "flatbuffers_common_builder.h"
#include "nodes_builder.h"
#include "policy_builder.h"

#include <stdbool.h>
#include <stddef.h>

#define MAX_OBSERVED 16

// Records what the observer saw, in the order nodes were entered.
typedef struct observed {
  size_t policy_enters;
  size_t policy_exits;
  plcs_uuid policy_id;
  int64_t policy_version;
  const char *policy_description;
  plcs_evaluation_result policy_result;

  plcs_evaluation_record records[MAX_OBSERVED];
  size_t len;
  bool overflowed;

  // set by the action handler, to check the policy is explained before it acts
  bool action_saw_policy_exit;
} observed;

static observed g_observed;

static void observe_policy_enter(void *user, plcs_uuid id, int64_t version, const char *description) {
  observed *obs = user;
  ++obs->policy_enters;
  obs->policy_id = id;
  obs->policy_version = version;
  obs->policy_description = description;
}

static void observe_policy_exit(void *user, plcs_evaluation_result result) {
  observed *obs = user;
  ++obs->policy_exits;
  obs->policy_result = result;
}

static size_t observe_node_enter(void *user, int depth) {
  observed *obs = user;
  (void)depth;

  if (obs->len >= MAX_OBSERVED) {
    obs->overflowed = true;
    return MAX_OBSERVED - 1;
  }
  return obs->len++;
}

static void observe_node_exit(void *user, size_t handle, const plcs_evaluation_record *record) {
  observed *obs = user;
  obs->records[handle] = *record;
}

static const plcs_observer g_observer = {
    .policy_enter = observe_policy_enter,
    .policy_exit = observe_policy_exit,
    .node_enter = observe_node_enter,
    .node_exit = observe_node_exit,
    .user = &g_observed,
};

static plcs_errors recording_action(
    plcs_evaluation_result res,
    char *values[],
    size_t value_len,
    const char *description,
    int action_id,
    plcs_uuid policy_id,
    int64_t policy_version,
    const char *policy_description
) {
  (void)res;
  (void)values;
  (void)value_len;
  (void)description;
  (void)action_id;
  (void)policy_id;
  (void)policy_version;
  (void)policy_description;
  g_observed.action_saw_policy_exit = g_observed.policy_exits > 0;
  return PLCS_ESUCCESS;
}

// An OS evaluator node, which the tests drive by setting PLCS_STR_EVAL_OS on the context.
static dd_wls_NodeTypeWrapper_ref_t
build_os_evaluator_node(flatcc_builder_t *b, const char *description, const char *expected) {
  dd_wls_StrEvaluator_ref_t eval = dd_wls_StrEvaluator_create(
      b, dd_wls_StringEvaluators_OS, dd_wls_CmpTypeSTR_CMP_EXACT, flatbuffers_string_create_str(b, expected)
  );
  dd_wls_EvaluatorNode_ref_t evaluator_node = dd_wls_EvaluatorNode_create(
      b, flatbuffers_string_create_str(b, description), dd_wls_EvaluatorType_as_StrEvaluator(eval)
  );
  return dd_wls_NodeTypeWrapper_create(b, dd_wls_NodeType_as_EvaluatorNode(evaluator_node));
}

// An evaluator node holding no evaluator, which is what a malformed policy carries.
static dd_wls_NodeTypeWrapper_ref_t
build_evaluator_node_without_evaluator(flatcc_builder_t *b, const char *description) {
  dd_wls_EvaluatorNode_ref_t evaluator_node =
      dd_wls_EvaluatorNode_create(b, flatbuffers_string_create_str(b, description), dd_wls_EvaluatorType_as_NONE());
  return dd_wls_NodeTypeWrapper_create(b, dd_wls_NodeType_as_EvaluatorNode(evaluator_node));
}

static dd_wls_NodeTypeWrapper_ref_t build_composite(
    flatcc_builder_t *b,
    const char *description,
    dd_wls_BoolOperation_enum_t op,
    const dd_wls_NodeTypeWrapper_ref_t children[],
    size_t children_len
) {
  dd_wls_NodeTypeWrapper_vec_start(b);
  for (size_t ix = 0; ix < children_len; ++ix) {
    dd_wls_NodeTypeWrapper_vec_push(b, children[ix]);
  }
  dd_wls_NodeTypeWrapper_vec_ref_t vec = dd_wls_NodeTypeWrapper_vec_end(b);

  dd_wls_CompositeNode_ref_t composite =
      dd_wls_CompositeNode_create(b, flatbuffers_string_create_str(b, description), op, vec);
  return dd_wls_NodeTypeWrapper_create(b, dd_wls_NodeType_as_CompositeNode(composite));
}

// Wraps rules into a single-policy buffer firing INJECT_ALLOW. Pass 0 for a policy without rules.
static void finish_policy_buffer(flatcc_builder_t *b, dd_wls_NodeTypeWrapper_ref_t rules, void **out, size_t *out_len) {
  dd_wls_Action_start(b);
  dd_wls_Action_action_add(b, dd_wls_ActionId_INJECT_ALLOW);
  dd_wls_Action_ref_t action = dd_wls_Action_end(b);
  dd_wls_Action_vec_start(b);
  dd_wls_Action_vec_push(b, action);
  dd_wls_Action_vec_ref_t actions = dd_wls_Action_vec_end(b);

  dd_wls_UUID_t id;
  dd_wls_UUID_assign(&id, 0xAABBCCDDEEFF0011ULL, 0x2233445566778899ULL);

  dd_wls_Policy_start(b);
  dd_wls_Policy_description_add(b, flatbuffers_string_create_str(b, "observed policy"));
  if (rules != 0) {
    dd_wls_Policy_rules_add(b, rules);
  }
  dd_wls_Policy_actions_add(b, actions);
  dd_wls_Policy_id_add(b, &id);
  dd_wls_Policy_version_add(b, 42);
  dd_wls_Policy_ref_t policy = dd_wls_Policy_end(b);

  dd_wls_Policy_vec_start(b);
  dd_wls_Policy_vec_push(b, policy);
  dd_wls_Policies_create_as_root(b, dd_wls_Policy_vec_end(b));

  *out = flatcc_builder_finalize_buffer(b, out_len);
  flatcc_builder_clear(b);
}

// AND(os == linux, NOT(os == windows), OR(os == darwin, os == linux)), which is
// true and reaches every node in the tree.
static void build_mixed_tree_policy(void **out, size_t *out_len) {
  flatcc_builder_t b;
  flatcc_builder_init(&b);

  dd_wls_NodeTypeWrapper_ref_t not_child[] = {build_os_evaluator_node(&b, "not windows", "windows")};
  dd_wls_NodeTypeWrapper_ref_t or_children[] = {
      build_os_evaluator_node(&b, "darwin", "darwin"),
      build_os_evaluator_node(&b, "linux again", "linux"),
  };
  dd_wls_NodeTypeWrapper_ref_t root_children[] = {
      build_os_evaluator_node(&b, "is linux", "linux"),
      build_composite(&b, "not", dd_wls_BoolOperation_BOOL_NOT, not_child, 1),
      build_composite(&b, "or", dd_wls_BoolOperation_BOOL_OR, or_children, 2),
  };

  finish_policy_buffer(&b, build_composite(&b, "root", dd_wls_BoolOperation_BOOL_AND, root_children, 3), out, out_len);
}

// Starts from a clean context watched by `observer`, with the OS the policies below expect.
static void reset_ctx_with_observer(const plcs_observer *observer) {
  (void)plcs_eval_ctx_init();
  plcs_eval_ctx_reset();

  g_observed = (observed){0};
  plcs_eval_ctx_set_observer(observer);
  (void)plcs_eval_ctx_set_str_eval_param(PLCS_STR_EVAL_OS, "linux");
  (void)plcs_eval_ctx_register_action(recording_action, PLCS_ACTION_INJECT_ALLOW);
}

UTEST(observer, reports_every_visited_node_in_tree_order) {
  void *buffer = NULL;
  size_t buffer_len = 0;
  build_mixed_tree_policy(&buffer, &buffer_len);
  ASSERT_TRUE(buffer != NULL);

  reset_ctx_with_observer(&g_observer);
  ASSERT_EQ((int)plcs_evaluate_buffer(buffer, buffer_len), (int)PLCS_ESUCCESS);

  ASSERT_FALSE(g_observed.overflowed);
  ASSERT_EQ(g_observed.len, (size_t)7);

  // pre-order: a node comes before the children it is built from
  const plcs_node_kind expected_kinds[] = {
      PLCS_NODE_AND, PLCS_NODE_STR_EVAL, PLCS_NODE_NOT,      PLCS_NODE_STR_EVAL,
      PLCS_NODE_OR,  PLCS_NODE_STR_EVAL, PLCS_NODE_STR_EVAL,
  };
  const int expected_depths[] = {0, 1, 1, 2, 1, 2, 2};
  const plcs_evaluation_result expected_results[] = {
      PLCS_EVAL_RESULT_TRUE, PLCS_EVAL_RESULT_TRUE,  PLCS_EVAL_RESULT_TRUE, PLCS_EVAL_RESULT_FALSE,
      PLCS_EVAL_RESULT_TRUE, PLCS_EVAL_RESULT_FALSE, PLCS_EVAL_RESULT_TRUE,
  };

  for (size_t ix = 0; ix < g_observed.len; ++ix) {
    ASSERT_EQ((int)g_observed.records[ix].kind, (int)expected_kinds[ix]);
    ASSERT_EQ(g_observed.records[ix].depth, expected_depths[ix]);
    ASSERT_EQ((int)g_observed.records[ix].result, (int)expected_results[ix]);
  }

  flatcc_builder_free(buffer);
}

UTEST(observer, evaluator_nodes_report_both_compared_values) {
  void *buffer = NULL;
  size_t buffer_len = 0;
  build_mixed_tree_policy(&buffer, &buffer_len);
  ASSERT_TRUE(buffer != NULL);

  reset_ctx_with_observer(&g_observer);
  ASSERT_EQ((int)plcs_evaluate_buffer(buffer, buffer_len), (int)PLCS_ESUCCESS);
  ASSERT_EQ(g_observed.len, (size_t)7);

  const plcs_evaluation_record *evaluator = &g_observed.records[1];
  ASSERT_EQ((int)evaluator->kind, (int)PLCS_NODE_STR_EVAL);
  ASSERT_EQ(evaluator->evaluator_id, (int)PLCS_STR_EVAL_OS);
  ASSERT_EQ(evaluator->comparator, (int)PLCS_STR_CMP_EXACT);
  ASSERT_STREQ(evaluator->description, "is linux");
  ASSERT_STREQ(evaluator->policy_value.str, "linux");
  ASSERT_STREQ(evaluator->process_value.str, "linux");

  // a composite compares nothing, so it only carries its own description
  const plcs_evaluation_record *composite = &g_observed.records[0];
  ASSERT_EQ((int)composite->kind, (int)PLCS_NODE_AND);
  ASSERT_STREQ(composite->description, "root");
  ASSERT_EQ(composite->evaluator_id, 0);
  ASSERT_EQ(composite->comparator, 0);

  flatcc_builder_free(buffer);
}

// A failed parse leaves BOOL_UNKNOWN behind, and a malformed policy can hold an
// evaluator node with no evaluator. The evaluation abstains on both, so neither is
// reported as one of the kinds it knows how to evaluate.
UTEST(observer, reports_nodes_it_cannot_interpret_as_unknown) {
  flatcc_builder_t b;
  flatcc_builder_init(&b);

  dd_wls_NodeTypeWrapper_ref_t children[] = {build_evaluator_node_without_evaluator(&b, "no evaluator")};
  void *buffer = NULL;
  size_t buffer_len = 0;
  finish_policy_buffer(
      &b, build_composite(&b, "unparsed operator", dd_wls_BoolOperation_BOOL_UNKNOWN, children, 1), &buffer, &buffer_len
  );
  ASSERT_TRUE(buffer != NULL);

  reset_ctx_with_observer(&g_observer);
  ASSERT_EQ((int)plcs_evaluate_buffer(buffer, buffer_len), (int)PLCS_ESUCCESS);

  ASSERT_EQ(g_observed.len, (size_t)2);
  ASSERT_EQ((int)g_observed.records[0].kind, (int)PLCS_NODE_UNKNOWN);
  ASSERT_EQ((int)g_observed.records[1].kind, (int)PLCS_NODE_UNKNOWN);
  ASSERT_EQ((int)g_observed.records[0].result, (int)PLCS_EVAL_RESULT_ABSTAIN);
  ASSERT_EQ((int)g_observed.records[1].result, (int)PLCS_EVAL_RESULT_ABSTAIN);

  // the description belongs to the node, not to what it holds, so it still comes through
  ASSERT_STREQ(g_observed.records[0].description, "unparsed operator");
  ASSERT_STREQ(g_observed.records[1].description, "no evaluator");

  flatcc_builder_free(buffer);
}

UTEST(observer, frames_the_policy_it_is_reporting_on) {
  void *buffer = NULL;
  size_t buffer_len = 0;
  build_mixed_tree_policy(&buffer, &buffer_len);
  ASSERT_TRUE(buffer != NULL);

  reset_ctx_with_observer(&g_observer);
  ASSERT_EQ((int)plcs_evaluate_buffer(buffer, buffer_len), (int)PLCS_ESUCCESS);

  ASSERT_EQ(g_observed.policy_enters, (size_t)1);
  ASSERT_EQ(g_observed.policy_exits, (size_t)1);
  ASSERT_STREQ(g_observed.policy_description, "observed policy");
  ASSERT_EQ(g_observed.policy_version, (int64_t)42);
  ASSERT_EQ(g_observed.policy_id.hi, (uint64_t)0xAABBCCDDEEFF0011ULL);
  ASSERT_EQ(g_observed.policy_id.lo, (uint64_t)0x2233445566778899ULL);
  ASSERT_EQ((int)g_observed.policy_result, (int)PLCS_EVAL_RESULT_TRUE);

  // the explanation is complete before anything acts on it
  ASSERT_TRUE(g_observed.action_saw_policy_exit);

  flatcc_builder_free(buffer);
}

UTEST(observer, skips_nodes_a_short_circuit_never_reached) {
  flatcc_builder_t b;
  flatcc_builder_init(&b);
  dd_wls_NodeTypeWrapper_ref_t children[] = {
      build_os_evaluator_node(&b, "is windows", "windows"),
      build_os_evaluator_node(&b, "never evaluated", "linux"),
  };
  void *buffer = NULL;
  size_t buffer_len = 0;
  finish_policy_buffer(
      &b, build_composite(&b, "root", dd_wls_BoolOperation_BOOL_AND, children, 2), &buffer, &buffer_len
  );
  ASSERT_TRUE(buffer != NULL);

  reset_ctx_with_observer(&g_observer);
  ASSERT_EQ((int)plcs_evaluate_buffer(buffer, buffer_len), (int)PLCS_ESUCCESS);

  // the AND gave up on the first false child, so the second one is not reported
  ASSERT_EQ(g_observed.len, (size_t)2);
  ASSERT_EQ((int)g_observed.records[0].kind, (int)PLCS_NODE_AND);
  ASSERT_EQ((int)g_observed.records[0].result, (int)PLCS_EVAL_RESULT_FALSE);
  ASSERT_STREQ(g_observed.records[1].description, "is windows");
  ASSERT_EQ((int)g_observed.records[1].result, (int)PLCS_EVAL_RESULT_FALSE);

  flatcc_builder_free(buffer);
}

UTEST(observer, frames_a_policy_that_has_no_rules) {
  flatcc_builder_t b;
  flatcc_builder_init(&b);
  void *buffer = NULL;
  size_t buffer_len = 0;
  finish_policy_buffer(&b, 0, &buffer, &buffer_len);
  ASSERT_TRUE(buffer != NULL);

  reset_ctx_with_observer(&g_observer);
  ASSERT_EQ((int)plcs_evaluate_buffer(buffer, buffer_len), (int)PLCS_ESUCCESS);

  // no node is ever entered, so the framing callbacks are the only report
  ASSERT_EQ(g_observed.len, (size_t)0);
  ASSERT_EQ(g_observed.policy_enters, (size_t)1);
  ASSERT_EQ(g_observed.policy_exits, (size_t)1);
  ASSERT_EQ((int)g_observed.policy_result, (int)PLCS_EVAL_RESULT_ABSTAIN);

  flatcc_builder_free(buffer);
}

UTEST(observer, evaluation_is_unchanged_without_an_observer) {
  void *buffer = NULL;
  size_t buffer_len = 0;
  build_mixed_tree_policy(&buffer, &buffer_len);
  ASSERT_TRUE(buffer != NULL);

  reset_ctx_with_observer(NULL);
  ASSERT_TRUE(plcs_eval_ctx_get_observer() == NULL);
  ASSERT_EQ((int)plcs_evaluate_buffer(buffer, buffer_len), (int)PLCS_ESUCCESS);

  ASSERT_EQ(g_observed.len, (size_t)0);
  ASSERT_EQ(g_observed.policy_enters, (size_t)0);

  flatcc_builder_free(buffer);
}

UTEST(observer, a_partially_filled_observer_is_fine) {
  void *buffer = NULL;
  size_t buffer_len = 0;
  build_mixed_tree_policy(&buffer, &buffer_len);
  ASSERT_TRUE(buffer != NULL);

  // only interested in results, so nothing tracks handles
  static const plcs_observer results_only = {.node_exit = observe_node_exit, .user = &g_observed};
  reset_ctx_with_observer(&results_only);
  ASSERT_EQ((int)plcs_evaluate_buffer(buffer, buffer_len), (int)PLCS_ESUCCESS);

  // without node_enter every record lands on handle 0, which is the last one out
  ASSERT_EQ(g_observed.len, (size_t)0);
  ASSERT_EQ((int)g_observed.records[0].kind, (int)PLCS_NODE_AND);
  ASSERT_EQ((int)g_observed.records[0].result, (int)PLCS_EVAL_RESULT_TRUE);

  flatcc_builder_free(buffer);
}

UTEST(observer, is_dropped_by_a_context_reset) {
  reset_ctx_with_observer(&g_observer);
  ASSERT_TRUE(plcs_eval_ctx_get_observer() == &g_observer);

  plcs_eval_ctx_reset();
  ASSERT_TRUE(plcs_eval_ctx_get_observer() == NULL);
}
