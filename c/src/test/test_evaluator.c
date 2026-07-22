/*
 * Unless explicitly stated otherwise all files in this repository are licensed
 * under the Apache 2.0 License. This product includes software developed at
 * Datadog (https://www.datadoghq.com/).
 *
 * Copyright 2025-Present Datadog, Inc.
 * ----
 * Unit test stubs for evaluator behavior and boolean composition.
 *
 * These tests use utest.h and exercise:
 *  - evaluate_buffer() basic behavior on NULL/invalid input
 *  - enum-to-string mapping helpers (sanity checks)
 *  - end-to-end evaluation with a generated FlatBuffers header (if available)
 *  - Individual evaluator functions (evaluate_numeric, evaluate_unumeric, node_evaluator)
 *  - Boolean operation functions (DoAnd, DoOr, DoNot, DoOper)
 *  - Composite evaluator function (composite_evaluator)
 *
 * For the end-to-end test, if you have run the Go example
 *   make -C ../go example_generate_c_header_buffer
 * the header will be available at:
 *   policies/go/example_generate_c_header_buffer/out/buffer.h
 * and exposes: `extern const uint8_t hardcoded_policies[];`
 *
 * For full testing of evaluator functions, the following FlatCC builder headers are needed:
 * - evaluators_builder.h (for creating mock StrEvaluator, NumEvaluator, UNumEvaluator)
 * - nodes_builder.h (for creating mock EvaluatorNode, CompositeNode, NodeTypeWrapper)
 * - boolean_operation_builder.h (for creating mock BoolOperation enums)
 *
 * This file intentionally avoids building FlatBuffers in C (we only ship
 * the reader side here). It relies on a generated header when present.
 */
#define _GNU_SOURCE

/* System headers */
#include <stddef.h>
#include <stdint.h>
#include <string.h>

/* External library headers */
#include "utest/utest.h"

/* Project public headers */
#include <dd/policies/error_codes.h>
#include <dd/policies/evaluator_default.h>
#include <dd/policies/policies.h>

/* Project internal headers */
#include "actions_reader.h"
#include "eval_ctx.h"
#include "evaluators_verifier.h"
#include "flatbuffers_common_builder.h"
#include "policy.h"
#include "policy_builder.h"
#include "wire/action.h"
#include "wire/boolean_operation.h"
#include "wire/dd_types.h"
#include "wire/evaluation_result.h"

/* Test-specific headers */
#include "hardcoded_policies.h"

// unexported functions
extern plcs_evaluation_result string_evaluator_exact(const char *eval, const char *param);
extern plcs_evaluation_result string_evaluator_prefix(const char *eval, const char *param);
extern plcs_evaluation_result string_evaluator_suffix(const char *eval, const char *param);
extern plcs_evaluation_result string_evaluator_contains(const char *eval, const char *param);
extern plcs_evaluation_result string_evaluator_wildcard(const char *pattern, const char *str);

extern void plcs_eval_ctx_set_str_eval_error(plcs_string_evaluators ix, plcs_errors error);
extern void plcs_eval_ctx_set_num_eval_error(plcs_numeric_evaluators ix, plcs_errors error);
extern void plcs_eval_ctx_set_unum_eval_error(plcs_numeric_evaluators ix, plcs_errors error);

// Additional extern declarations for evaluator functions
extern plcs_evaluation_result evaluate_string(dd_ns(StrEvaluator_table_t) eval_str, const char *description);
extern plcs_evaluation_result evaluate_numeric(dd_ns(NumEvaluator_table_t) eval_num, const char *description);
extern plcs_evaluation_result evaluate_unumeric(dd_ns(UNumEvaluator_table_t) eval_unum, const char *description);
extern plcs_evaluation_result node_evaluator(dd_ns(EvaluatorNode_table_t) node);
extern plcs_evaluation_result DoAnd(plcs_evaluation_result a, plcs_evaluation_result b);
extern plcs_evaluation_result DoOr(plcs_evaluation_result a, plcs_evaluation_result b);
extern plcs_evaluation_result DoNot(plcs_evaluation_result res);
extern plcs_evaluation_result
DoOper(dd_ns(BoolOperation_enum_t) oper, plcs_evaluation_result a, plcs_evaluation_result b);
extern plcs_evaluation_result composite_evaluator(dd_ns(CompositeNode_table_t) node);

extern void plcs_eval_ctx_reset(void);

/* -------------------------------------------------------------------------- */
/* Helpers                                                                     */
/* -------------------------------------------------------------------------- */

static int g_allow_called = 0;
static int g_deny_called = 0;

static plcs_errors test_action_allow(
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
  g_allow_called++;
  return PLCS_ESUCCESS;
}

static plcs_errors test_action_deny(
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
  g_deny_called++;
  return PLCS_ESUCCESS;
}

/* -------------------------------------------------------------------------- */
/* Tests                                                                       */
/* -------------------------------------------------------------------------- */

UTEST(evaluator, evaluate_buffer_null_returns_no_data) {
  /* Passing NULL buffer yields PLCS_ENO_DATA */
  int rc = plcs_evaluate_buffer(NULL, 0);
  ASSERT_EQ(rc, PLCS_ENO_DATA);
}

UTEST(evaluator, enum_mappings_return_strings) {
  /* These must return non-NULL stable names derived from the FlatBuffers schema */
  ASSERT_TRUE(plcs_string_evaluators_to_string(PLCS_STR_EVAL_COMPONENT) != NULL);
  ASSERT_TRUE(plcs_string_evaluators_to_string(PLCS_STR_EVAL_RUNTIME_LANGUAGE) != NULL);
  ASSERT_TRUE(plcs_numeric_evaluators_to_string(PLCS_NUM_EVAL_JAVA_HEAP) != NULL);
  ASSERT_TRUE(plcs_string_comparator_to_string(PLCS_STR_CMP_EXACT) != NULL);
  ASSERT_TRUE(plcs_numeric_comparator_to_string(PLCS_NUM_CMP_LTE) != NULL);
  ASSERT_TRUE(plcs_evaluation_result_to_string(PLCS_EVAL_RESULT_TRUE) != NULL);
  ASSERT_TRUE(plcs_actions_to_string(PLCS_ACTION_INJECT_DENY) != NULL);
  ASSERT_TRUE(plcs_actions_to_string(PLCS_ACTION_SET_ENVAR) != NULL);
}

UTEST(evaluator, test_string_evaluator) {
  /* Test the string evaluator with a simple exact match */
  int res = string_evaluator_exact("test", "test");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);

  /* Test with a mismatch */
  res = string_evaluator_exact("test", "not_test");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_FALSE);

  /* Test with NULL parameters */
  res = string_evaluator_exact(NULL, "test");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_ABSTAIN);
  res = string_evaluator_exact("test", NULL);
  ASSERT_EQ(res, PLCS_EVAL_RESULT_ABSTAIN);

  res = string_evaluator_prefix(NULL, "test");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_ABSTAIN);
  res = string_evaluator_prefix("test", NULL);
  ASSERT_EQ(res, PLCS_EVAL_RESULT_ABSTAIN);

  res = string_evaluator_suffix(NULL, "test");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_ABSTAIN);
  res = string_evaluator_suffix("test", NULL);
  ASSERT_EQ(res, PLCS_EVAL_RESULT_ABSTAIN);
  res = string_evaluator_suffix("long_test", "test");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_FALSE);

  res = string_evaluator_contains(NULL, "test");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_ABSTAIN);
  res = string_evaluator_contains("test", NULL);
  ASSERT_EQ(res, PLCS_EVAL_RESULT_ABSTAIN);
  res = string_evaluator_contains("test", "long_test");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);
  res = string_evaluator_contains("", "long_test");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);
}

UTEST(evaluator, test_string_evaluator_wildcard) {
  /* NULL parameters */
  int res = string_evaluator_wildcard(NULL, "test");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_ABSTAIN);
  res = string_evaluator_wildcard("test", NULL);
  ASSERT_EQ(res, PLCS_EVAL_RESULT_ABSTAIN);
  res = string_evaluator_wildcard(NULL, NULL);
  ASSERT_EQ(res, PLCS_EVAL_RESULT_ABSTAIN);

  /* Exact match (no wildcards) */
  res = string_evaluator_wildcard("hello", "hello");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);
  res = string_evaluator_wildcard("hello", "world");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_FALSE);
  res = string_evaluator_wildcard("", "");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);
  res = string_evaluator_wildcard("", "a");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_FALSE);
  res = string_evaluator_wildcard("a", "");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_FALSE);

  /* ? matches exactly one character */
  res = string_evaluator_wildcard("?", "a");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);
  res = string_evaluator_wildcard("?", "");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_FALSE);
  res = string_evaluator_wildcard("?", "ab");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_FALSE);
  res = string_evaluator_wildcard("a?c", "abc");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);
  res = string_evaluator_wildcard("a?c", "aXc");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);
  res = string_evaluator_wildcard("a?c", "ac");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_FALSE);
  res = string_evaluator_wildcard("a?c", "abbc");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_FALSE);
  res = string_evaluator_wildcard("???", "abc");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);
  res = string_evaluator_wildcard("???", "ab");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_FALSE);
  res = string_evaluator_wildcard("???", "abcd");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_FALSE);

  /* * matches zero or more characters */
  res = string_evaluator_wildcard("*", "");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);
  res = string_evaluator_wildcard("*", "anything");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);
  res = string_evaluator_wildcard("a*", "a");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);
  res = string_evaluator_wildcard("a*", "abc");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);
  res = string_evaluator_wildcard("a*", "b");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_FALSE);
  res = string_evaluator_wildcard("*c", "c");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);
  res = string_evaluator_wildcard("*c", "abc");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);
  res = string_evaluator_wildcard("*c", "abd");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_FALSE);
  res = string_evaluator_wildcard("a*c", "ac");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);
  res = string_evaluator_wildcard("a*c", "abc");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);
  res = string_evaluator_wildcard("a*c", "aXYZc");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);
  res = string_evaluator_wildcard("a*c", "aXYZd");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_FALSE);

  /* Multiple * wildcards */
  res = string_evaluator_wildcard("*foo*", "foo");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);
  res = string_evaluator_wildcard("*foo*", "XXfooYY");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);
  res = string_evaluator_wildcard("*foo*bar*", "foobar");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);
  res = string_evaluator_wildcard("*foo*bar*", "XXfooYYbarZZ");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);
  res = string_evaluator_wildcard("*foo*bar*", "barfoo");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_FALSE);
  res = string_evaluator_wildcard("*foo*bar*", "fooXXbaz");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_FALSE);

  /* Consecutive * treated as single * */
  res = string_evaluator_wildcard("a***b", "ab");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);
  res = string_evaluator_wildcard("a***b", "aXXXb");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);

  /* PROCESS_ENVAR null JSON value: converter emits KEY=*? (non-empty RHS only) */
  res = string_evaluator_wildcard("FOO=*?", "FOO=");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_FALSE);
  res = string_evaluator_wildcard("FOO=*?", "FOO=a");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);
  res = string_evaluator_wildcard("FOO=*?", "FOO=ab");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);

  /* Mixed ? and * */
  res = string_evaluator_wildcard("a?*", "ab");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);
  res = string_evaluator_wildcard("a?*", "abc");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);
  res = string_evaluator_wildcard("a?*", "a");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_FALSE);
  res = string_evaluator_wildcard("*?", "a");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);
  res = string_evaluator_wildcard("*?", "");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_FALSE);

  /* RFC-style patterns: executable path matching */
  res = string_evaluator_wildcard("**/java", "/usr/bin/java");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);
  res = string_evaluator_wildcard("**/java-1.5*/**/java", "/usr/lib/java-1.5.0/bin/java");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);
  res = string_evaluator_wildcard("**/java-1.5*/**/java", "/usr/lib/java-1.8.0/bin/java");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_FALSE);

  /* RFC-style patterns: exe? matching single char suffix */
  res = string_evaluator_wildcard("**/exe?", "/some/exe2");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);
  res = string_evaluator_wildcard("**/exe?", "/some/other/exeA");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);
  res = string_evaluator_wildcard("**/exe?", "/some/exe");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_FALSE);
  res = string_evaluator_wildcard("**/exe?", "/some/exe22");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_FALSE);

  /* RFC-style patterns: argument matching */
  res = string_evaluator_wildcard("1.*", "1.2.3");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);
  res = string_evaluator_wildcard("1.*", "2.0.0");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_FALSE);
  res = string_evaluator_wildcard("*csc.dll", "csc.dll");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);
  res = string_evaluator_wildcard("*csc.dll", "/path/to/csc.dll");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);
  res = string_evaluator_wildcard("*csc.dll", "other.dll");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_FALSE);

  /* Pathological patterns: ensure no excessive backtracking */
  res = string_evaluator_wildcard("*a*a*a*a*b", "aaaaaaaaaaac");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_FALSE);
  res = string_evaluator_wildcard("*a*a*a*a*b", "aaaaaaaaaaab");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);
}

UTEST(evaluator, default_string_eval_wildcard) {
  /* Test wildcard through the plcs_default_string_evaluator dispatch */
  ASSERT_EQ(
      plcs_default_string_evaluator("*.dll", PLCS_STR_CMP_WILDCARD, "test.dll", "d", PLCS_STR_EVAL_COMPONENT),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_TRUE
  );
  ASSERT_EQ(
      plcs_default_string_evaluator("*.dll", PLCS_STR_CMP_WILDCARD, "test.exe", "d", PLCS_STR_EVAL_COMPONENT),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_FALSE
  );
  ASSERT_EQ(
      plcs_default_string_evaluator("he??o", PLCS_STR_CMP_WILDCARD, "hello", "d", PLCS_STR_EVAL_COMPONENT),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_TRUE
  );
  ASSERT_EQ(
      plcs_default_string_evaluator(NULL, PLCS_STR_CMP_WILDCARD, "test", "d", PLCS_STR_EVAL_COMPONENT),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_ABSTAIN
  );
  ASSERT_EQ(
      plcs_default_string_evaluator("test", PLCS_STR_CMP_WILDCARD, NULL, "d", PLCS_STR_EVAL_COMPONENT),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_ABSTAIN
  );
}

UTEST(evaluator, out_of_bounds_eval_error_setters) {
  plcs_eval_ctx_set_str_eval_error(PLCS_STR_EVAL__COUNT, PLCS_EUNKNOWN_EVAL_IX);
  int err = plcs_eval_ctx_get_str_eval_error(PLCS_STR_EVAL__COUNT);
  ASSERT_EQ(err, PLCS_EIX_OVERFLOW);

  plcs_eval_ctx_set_num_eval_error(PLCS_NUM_EVAL__COUNT, PLCS_EUNKNOWN_EVAL_IX);
  err = plcs_eval_ctx_get_num_eval_error(PLCS_NUM_EVAL__COUNT);
  ASSERT_EQ(err, PLCS_EIX_OVERFLOW);

  plcs_eval_ctx_set_unum_eval_error(PLCS_NUM_EVAL__COUNT, PLCS_EUNKNOWN_EVAL_IX);
  err = plcs_eval_ctx_get_unum_eval_error(PLCS_NUM_EVAL__COUNT);
  ASSERT_EQ(err, PLCS_EIX_OVERFLOW);
}

UTEST(evaluator, default_string_eval_sanity) {
  /* Sanity around defaults (no buffer involved) */
  ASSERT_EQ(
      plcs_default_string_evaluator("abc", PLCS_STR_CMP_EXACT, "abc", "d", PLCS_STR_EVAL_COMPONENT),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_TRUE
  );
  ASSERT_EQ(
      plcs_default_string_evaluator("ab", PLCS_STR_CMP_PREFIX, "abc", "d", PLCS_STR_EVAL_COMPONENT),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_TRUE
  );
  ASSERT_EQ(
      plcs_default_string_evaluator("bc", PLCS_STR_CMP_SUFFIX, "abc", "d", PLCS_STR_EVAL_COMPONENT),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_TRUE
  );
  ASSERT_EQ(
      plcs_default_string_evaluator("b", PLCS_STR_CMP_CONTAINS, "abc", "d", PLCS_STR_EVAL_COMPONENT),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_TRUE
  );
  ASSERT_EQ(
      plcs_default_string_evaluator("a?c", PLCS_STR_CMP_WILDCARD, "abc", "d", PLCS_STR_EVAL_COMPONENT),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_TRUE
  );

  /* Abstain on missing data */
  ASSERT_EQ(
      plcs_default_string_evaluator(NULL, PLCS_STR_CMP_EXACT, "abc", "d", PLCS_STR_EVAL_COMPONENT),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_ABSTAIN
  );
  ASSERT_EQ(
      plcs_default_string_evaluator("abc", PLCS_STR_CMP_EXACT, NULL, "d", PLCS_STR_EVAL_COMPONENT),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_ABSTAIN
  );

  ASSERT_EQ((int)evaluate_string(NULL, "d"), PLCS_EVAL_RESULT_ABSTAIN);
}

UTEST(evaluator, default_numeric_eval_sanity) {
  ASSERT_EQ(
      plcs_default_numeric_evaluator(5, PLCS_NUM_CMP_EQ, 5, "d", PLCS_NUM_EVAL_JAVA_HEAP),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_TRUE
  );
  ASSERT_EQ(
      plcs_default_numeric_evaluator(6, PLCS_NUM_CMP_GT, 5, "d", PLCS_NUM_EVAL_JAVA_HEAP),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_TRUE
  );
  ASSERT_EQ(
      plcs_default_numeric_evaluator(5, PLCS_NUM_CMP_GTE, 5, "d", PLCS_NUM_EVAL_JAVA_HEAP),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_TRUE
  );
  ASSERT_EQ(
      plcs_default_numeric_evaluator(4, PLCS_NUM_CMP_LT, 5, "d", PLCS_NUM_EVAL_JAVA_HEAP),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_TRUE
  );
  ASSERT_EQ(
      plcs_default_numeric_evaluator(5, PLCS_NUM_CMP_LTE, 5, "d", PLCS_NUM_EVAL_JAVA_HEAP),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_TRUE
  );
}

UTEST(evaluator, default_unumeric_eval_sanity) {
  ASSERT_EQ(
      plcs_default_unumeric_evaluator(5ul, PLCS_NUM_CMP_EQ, 5ul, "d", PLCS_NUM_EVAL_JAVA_HEAP),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_TRUE
  );
  ASSERT_EQ(
      plcs_default_unumeric_evaluator(6ul, PLCS_NUM_CMP_GT, 5ul, "d", PLCS_NUM_EVAL_JAVA_HEAP),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_TRUE
  );
  ASSERT_EQ(
      plcs_default_unumeric_evaluator(5ul, PLCS_NUM_CMP_GTE, 5ul, "d", PLCS_NUM_EVAL_JAVA_HEAP),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_TRUE
  );
  ASSERT_EQ(
      plcs_default_unumeric_evaluator(4ul, PLCS_NUM_CMP_LT, 5ul, "d", PLCS_NUM_EVAL_JAVA_HEAP),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_TRUE
  );
  ASSERT_EQ(
      plcs_default_unumeric_evaluator(5ul, PLCS_NUM_CMP_LTE, 5ul, "d", PLCS_NUM_EVAL_JAVA_HEAP),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_TRUE
  );
}

UTEST(evaluator, test_conversion_evalresult_to_wire) {
  ASSERT_EQ(dd_evalresult_to_wire(PLCS_EVAL_RESULT_TRUE), dd_ns(EvaluationResult_EVAL_RESULT_TRUE));
  ASSERT_EQ(dd_evalresult_to_wire(PLCS_EVAL_RESULT_FALSE), dd_ns(EvaluationResult_EVAL_RESULT_FALSE));
  ASSERT_EQ(dd_evalresult_to_wire(PLCS_EVAL_RESULT_ABSTAIN), dd_ns(EvaluationResult_EVAL_RESULT_ABSTAIN));
  ASSERT_EQ(dd_evalresult_to_wire(PLCS_EVAL_RESULT__COUNT), -1);
}

/* -------------------------------------------------------------------------- */
/* Tests for evaluate_string.                                                 */
/* -------------------------------------------------------------------------- */

UTEST(evaluator, test_evaluate_string_null_input) {
  /* Test with NULL evaluator */
  int res = evaluate_string(NULL, "test description");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_ABSTAIN);
}

UTEST(evaluator, test_evaluate_string_basic_functionality) {
  /* Initialize context for numeric evaluation */
  int rc = plcs_eval_ctx_init();
  ASSERT_TRUE(rc == PLCS_ESUCCESS || rc == PLCS_EINITIZLIED);

  /* Set a string parameter for testing */
  rc = plcs_eval_ctx_set_str_eval_param(PLCS_STR_EVAL_HOST_IP, "1.2.3.4");
  ASSERT_EQ(rc, PLCS_ESUCCESS);

  rc = plcs_eval_ctx_register_str_evaluator(plcs_default_string_evaluator, PLCS_STR_EVAL_HOST_IP);
  ASSERT_EQ(rc, PLCS_ESUCCESS);

  /* Mocking a flatbuffer object */
  flatcc_builder_t b;
  size_t sz;
  flatcc_builder_init(&b);

  dd_wls_StrEvaluator_create_as_root(
      &b, dd_wls_StringEvaluators_HOST_IP, dd_wls_CmpTypeSTR_CMP_EXACT, flatbuffers_string_create_str(&b, "1.2.3.4")
  );

  void *buf = flatcc_builder_finalize_buffer(&b, &sz);
  ASSERT_TRUE(dd_wls_StrEvaluator_verify_as_root(buf, sz) == 0);
  dd_wls_StrEvaluator_table_t eval = dd_wls_StrEvaluator_as_root(buf);

  int res = evaluate_string(eval, "d");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);

  res = evaluate_string(NULL, "d");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_ABSTAIN);

  // reset all ctx:
  plcs_eval_ctx_reset();
  res = evaluate_string(eval, "d");
  // shouldn't be any value
  ASSERT_EQ(res, PLCS_EVAL_RESULT_ABSTAIN);
  flatcc_builder_free(buf);
  flatcc_builder_reset(&b);
}

/* -------------------------------------------------------------------------- */
/* Tests for evaluate_numeric                                                  */
/* -------------------------------------------------------------------------- */

UTEST(evaluator, test_evaluate_numeric_null_input) {
  /* Test with NULL evaluator */
  int res = evaluate_numeric(NULL, "test description");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_ABSTAIN);
}

UTEST(evaluator, test_evaluate_numeric_basic_functionality) {
  /* Initialize context for numeric evaluation */
  int rc = plcs_eval_ctx_init();
  ASSERT_TRUE(rc == PLCS_ESUCCESS || rc == PLCS_EINITIZLIED);

  /* Set a numeric parameter for testing */
  rc = plcs_eval_ctx_set_num_eval_param(PLCS_NUM_EVAL_JAVA_HEAP, 100);
  ASSERT_EQ(rc, PLCS_ESUCCESS);
  /* Mocking a flatbuffer object */
  flatcc_builder_t b;
  size_t sz;
  flatcc_builder_init(&b);
  dd_wls_NumEvaluator_create_as_root(&b, dd_wls_NumericEvaluators_JAVA_HEAP, dd_wls_CmpTypeNUM_CMP_EQ, 100);
  void *buf = flatcc_builder_finalize_buffer(&b, &sz);
  ASSERT_TRUE(dd_wls_NumEvaluator_verify_as_root(buf, sz) == 0);
  dd_wls_NumEvaluator_table_t eval = dd_wls_NumEvaluator_as_root(buf);
  int res = evaluate_numeric(eval, "d");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);
  res = evaluate_numeric(NULL, "d");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_ABSTAIN);
  // force reset all ctx (init will return because of it's implementation)
  plcs_eval_ctx_reset();
  res = evaluate_numeric(eval, "d");
  // shouldn't be any value
  ASSERT_EQ(res, PLCS_EVAL_RESULT_FALSE);
  flatcc_builder_free(buf);
  flatcc_builder_reset(&b);
}

/* -------------------------------------------------------------------------- */
/* Tests for evaluate_unumeric                                                 */
/* -------------------------------------------------------------------------- */

UTEST(evaluator, test_evaluate_unumeric_null_input) {
  /* Test with NULL evaluator */
  int res = evaluate_unumeric(NULL, "test description");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_ABSTAIN);
}

UTEST(evaluator, test_evaluate_unumeric_basic_functionality) {
  /* Mock a UNumEvaluator with basic values */
  /* Note: This test requires proper mocking of FlatCC objects */

  /* Initialize context for unumeric evaluation */
  int rc = plcs_eval_ctx_init();
  ASSERT_TRUE(rc == PLCS_ESUCCESS || rc == PLCS_EINITIZLIED);

  /* Set an unumeric parameter for testing */
  rc = plcs_eval_ctx_set_unum_eval_param(PLCS_NUM_EVAL_JAVA_HEAP, 100);
  ASSERT_EQ(rc, PLCS_ESUCCESS);

  /* Mocking a flatbuffer object */
  flatcc_builder_t b;
  size_t sz;
  flatcc_builder_init(&b);

  dd_wls_UNumEvaluator_create_as_root(&b, dd_wls_NumericEvaluators_JAVA_HEAP, dd_wls_CmpTypeNUM_CMP_EQ, 100);

  void *buf = flatcc_builder_finalize_buffer(&b, &sz);
  ASSERT_TRUE(dd_wls_UNumEvaluator_verify_as_root(buf, sz) == 0);
  dd_wls_UNumEvaluator_table_t eval = dd_wls_UNumEvaluator_as_root(buf);

  int res = evaluate_unumeric(eval, "d");
  flatcc_builder_free(buf);
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);

  res = evaluate_unumeric(NULL, "d");
  ASSERT_EQ(res, PLCS_EVAL_RESULT_ABSTAIN);
  flatcc_builder_reset(&b);
}

/* -------------------------------------------------------------------------- */
/* Tests for node_evaluator                                                   */
/* -------------------------------------------------------------------------- */

UTEST(evaluator, test_node_evaluator_null_input) {
  /* Test with NULL node */
  int res = node_evaluator(NULL);
  ASSERT_EQ(res, PLCS_EVAL_RESULT_ABSTAIN);
}

UTEST(evaluator, test_node_evaluator_basic_functionality) {
  /* Mock an EvaluatorNode with basic values */
  /* Note: This test requires proper mocking of FlatCC objects */

#include <stdio.h>
  printf("hi?!\n");
  /* Initialize context for evaluation */
  int rc = plcs_eval_ctx_init();
  ASSERT_TRUE(rc == PLCS_ESUCCESS || rc == PLCS_EINITIZLIED);

  flatcc_builder_t b;
  size_t sz;
  flatcc_builder_init(&b);
  rc = plcs_eval_ctx_init();
  ASSERT_TRUE(rc == PLCS_ESUCCESS || rc == PLCS_EINITIZLIED);

  rc = plcs_eval_ctx_set_str_eval_param(PLCS_STR_EVAL_RUNTIME_ENTRY_POINT_JAR, "test.jar");
  ASSERT_EQ(rc, PLCS_ESUCCESS);
  // str eval
  dd_wls_StrEvaluator_ref_t str = dd_wls_StrEvaluator_create(
      &b, dd_wls_StringEvaluators_RUNTIME_ENTRY_POINT_JAR, dd_wls_CmpTypeSTR_CMP_EXACT,
      flatbuffers_string_create_str(&b, "test.jar")
  );
  dd_wls_EvaluatorNode_create_as_root(&b, str, dd_wls_EvaluatorType_as_StrEvaluator(str));
  void *buf = flatcc_builder_finalize_buffer(&b, &sz);
  dd_wls_EvaluatorNode_table_t eval = dd_wls_EvaluatorNode_as_root(buf);

  rc = node_evaluator(eval);
  flatcc_builder_free(buf);
  flatcc_builder_clear(&b);
  ASSERT_EQ(rc, PLCS_EVAL_RESULT_TRUE);

  /* Set a numeric parameter for testing */
  flatcc_builder_init(&b);
  rc = plcs_eval_ctx_init();
  ASSERT_TRUE(rc == PLCS_ESUCCESS || rc == PLCS_EINITIZLIED);
  rc = plcs_eval_ctx_set_num_eval_param(PLCS_NUM_EVAL_JAVA_HEAP, 100);
  ASSERT_EQ(rc, PLCS_ESUCCESS);

  // num eval
  dd_wls_NumEvaluator_ref_t num =
      dd_wls_NumEvaluator_create(&b, dd_wls_NumericEvaluators_JAVA_HEAP, dd_wls_CmpTypeNUM_CMP_EQ, 100);

  dd_wls_EvaluatorNode_create_as_root(&b, num, dd_wls_EvaluatorType_as_NumEvaluator(num));

  buf = flatcc_builder_finalize_buffer(&b, &sz);
  eval = dd_wls_EvaluatorNode_as_root(buf);

  rc = node_evaluator(eval);
  flatcc_builder_free(buf);
  flatcc_builder_clear(&b);
  ASSERT_EQ(rc, PLCS_EVAL_RESULT_TRUE);

  /* Set a numeric parameter for testing */
  flatcc_builder_init(&b);
  rc = plcs_eval_ctx_init();
  rc = plcs_eval_ctx_set_unum_eval_param(PLCS_NUM_EVAL_RUNTIME_VERSION_MINOR, 4);
  ASSERT_EQ(rc, PLCS_ESUCCESS);
  // unum eval
  dd_wls_UNumEvaluator_ref_t unum =
      dd_wls_UNumEvaluator_create(&b, dd_wls_NumericEvaluators_RUNTIME_VERSION_MINOR, dd_wls_CmpTypeNUM_CMP_EQ, 4);

  dd_wls_EvaluatorNode_create_as_root(&b, unum, dd_wls_EvaluatorType_as_UNumEvaluator(unum));

  buf = flatcc_builder_finalize_buffer(&b, &sz);
  eval = dd_wls_EvaluatorNode_as_root(buf);

  rc = node_evaluator(eval);
  flatcc_builder_free(buf);
  ASSERT_EQ(rc, PLCS_EVAL_RESULT_TRUE);
}

/* -------------------------------------------------------------------------- */
/* Tests for DoAnd function                                                    */
/* -------------------------------------------------------------------------- */

UTEST(evaluator, test_DoAnd_basic_operations) {
  /* Test AND logic with various combinations */

  /* TRUE & anything = anything */
  ASSERT_EQ((int)DoAnd(PLCS_EVAL_RESULT_TRUE, PLCS_EVAL_RESULT_TRUE), PLCS_EVAL_RESULT_TRUE);
  ASSERT_EQ((int)DoAnd(PLCS_EVAL_RESULT_TRUE, PLCS_EVAL_RESULT_FALSE), PLCS_EVAL_RESULT_FALSE);
  ASSERT_EQ((int)DoAnd(PLCS_EVAL_RESULT_TRUE, PLCS_EVAL_RESULT_ABSTAIN), PLCS_EVAL_RESULT_ABSTAIN);
  ASSERT_EQ((int)DoAnd(PLCS_EVAL_RESULT_FALSE, PLCS_EVAL_RESULT_TRUE), PLCS_EVAL_RESULT_FALSE);
  ASSERT_EQ((int)DoAnd(PLCS_EVAL_RESULT_ABSTAIN, PLCS_EVAL_RESULT_TRUE), PLCS_EVAL_RESULT_ABSTAIN);

  /* FALSE & FALSE = FALSE */
  ASSERT_EQ((int)DoAnd(PLCS_EVAL_RESULT_FALSE, PLCS_EVAL_RESULT_FALSE), PLCS_EVAL_RESULT_FALSE);

  /* FALSE & ABSTAIN = FALSE */
  ASSERT_EQ((int)DoAnd(PLCS_EVAL_RESULT_FALSE, PLCS_EVAL_RESULT_ABSTAIN), PLCS_EVAL_RESULT_FALSE);
  ASSERT_EQ((int)DoAnd(PLCS_EVAL_RESULT_ABSTAIN, PLCS_EVAL_RESULT_FALSE), PLCS_EVAL_RESULT_FALSE);

  /* ABSTAIN & ABSTAIN = ABSTAIN */
  ASSERT_EQ((int)DoAnd(PLCS_EVAL_RESULT_ABSTAIN, PLCS_EVAL_RESULT_ABSTAIN), PLCS_EVAL_RESULT_ABSTAIN);
}

/* -------------------------------------------------------------------------- */
/* Tests for DoOr function                                                    */
/* -------------------------------------------------------------------------- */

UTEST(evaluator, test_DoOr_basic_operations) {
  /* Test OR logic with various combinations */

  /* TRUE | anything = TRUE */
  ASSERT_EQ((int)DoOr(PLCS_EVAL_RESULT_TRUE, PLCS_EVAL_RESULT_TRUE), PLCS_EVAL_RESULT_TRUE);
  ASSERT_EQ((int)DoOr(PLCS_EVAL_RESULT_TRUE, PLCS_EVAL_RESULT_FALSE), PLCS_EVAL_RESULT_TRUE);
  ASSERT_EQ((int)DoOr(PLCS_EVAL_RESULT_TRUE, PLCS_EVAL_RESULT_ABSTAIN), PLCS_EVAL_RESULT_TRUE);
  ASSERT_EQ((int)DoOr(PLCS_EVAL_RESULT_FALSE, PLCS_EVAL_RESULT_TRUE), PLCS_EVAL_RESULT_TRUE);
  ASSERT_EQ((int)DoOr(PLCS_EVAL_RESULT_ABSTAIN, PLCS_EVAL_RESULT_TRUE), PLCS_EVAL_RESULT_TRUE);

  /* FALSE | FALSE = FALSE */
  ASSERT_EQ((int)DoOr(PLCS_EVAL_RESULT_FALSE, PLCS_EVAL_RESULT_FALSE), PLCS_EVAL_RESULT_FALSE);

  /* FALSE | ABSTAIN = ABSTAIN */
  ASSERT_EQ((int)DoOr(PLCS_EVAL_RESULT_FALSE, PLCS_EVAL_RESULT_ABSTAIN), PLCS_EVAL_RESULT_ABSTAIN);
  ASSERT_EQ((int)DoOr(PLCS_EVAL_RESULT_ABSTAIN, PLCS_EVAL_RESULT_FALSE), PLCS_EVAL_RESULT_ABSTAIN);

  /* ABSTAIN | ABSTAIN = ABSTAIN */
  ASSERT_EQ((int)DoOr(PLCS_EVAL_RESULT_ABSTAIN, PLCS_EVAL_RESULT_ABSTAIN), PLCS_EVAL_RESULT_ABSTAIN);
}

/* -------------------------------------------------------------------------- */
/* Tests for DoNot function                                                   */
/* -------------------------------------------------------------------------- */

UTEST(evaluator, test_DoNot_basic_operations) {
  /* Test NOT logic with various inputs */

  /* NOT TRUE = FALSE */
  ASSERT_EQ((int)DoNot(PLCS_EVAL_RESULT_TRUE), PLCS_EVAL_RESULT_FALSE);

  /* NOT FALSE = TRUE */
  ASSERT_EQ((int)DoNot(PLCS_EVAL_RESULT_FALSE), PLCS_EVAL_RESULT_TRUE);

  /* NOT ABSTAIN = ABSTAIN (preserved) */
  ASSERT_EQ((int)DoNot(PLCS_EVAL_RESULT_ABSTAIN), PLCS_EVAL_RESULT_ABSTAIN);
}

/* -------------------------------------------------------------------------- */
/* Tests for DoOper function                                                  */
/* -------------------------------------------------------------------------- */

UTEST(evaluator, test_DoOper_basic_operations) {
  /* Test DoOper with various boolean operations */

  /* Test AND operation */
  int res = DoOper(dd_ns(BoolOperation_BOOL_AND), PLCS_EVAL_RESULT_TRUE, PLCS_EVAL_RESULT_TRUE);
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);

  res = DoOper(dd_ns(BoolOperation_BOOL_AND), PLCS_EVAL_RESULT_TRUE, PLCS_EVAL_RESULT_FALSE);
  ASSERT_EQ(res, PLCS_EVAL_RESULT_FALSE);

  res = DoOper(dd_ns(BoolOperation_BOOL_AND), PLCS_EVAL_RESULT_FALSE, PLCS_EVAL_RESULT_TRUE);
  ASSERT_EQ(res, PLCS_EVAL_RESULT_FALSE);

  /* Test OR operation */
  res = DoOper(dd_ns(BoolOperation_BOOL_OR), PLCS_EVAL_RESULT_TRUE, PLCS_EVAL_RESULT_FALSE);
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);

  res = DoOper(dd_ns(BoolOperation_BOOL_OR), PLCS_EVAL_RESULT_FALSE, PLCS_EVAL_RESULT_FALSE);
  ASSERT_EQ(res, PLCS_EVAL_RESULT_FALSE);

  /* Test NOT operation (second parameter ignored for NOT) */
  res = DoOper(dd_ns(BoolOperation_BOOL_NOT), PLCS_EVAL_RESULT_TRUE, PLCS_EVAL_RESULT_FALSE);
  ASSERT_EQ(res, PLCS_EVAL_RESULT_FALSE);

  res = DoOper(dd_ns(BoolOperation_BOOL_NOT), PLCS_EVAL_RESULT_FALSE, PLCS_EVAL_RESULT_TRUE);
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);

  res = DoOper(dd_ns(BoolOperation_BOOL_NOT), PLCS_EVAL_RESULT_ABSTAIN, PLCS_EVAL_RESULT_TRUE);
  ASSERT_EQ(res, PLCS_EVAL_RESULT_ABSTAIN);

  /* Test unknown operation */
  res = DoOper(99, PLCS_EVAL_RESULT_TRUE, PLCS_EVAL_RESULT_TRUE);
  ASSERT_EQ(res, PLCS_EVAL_RESULT_ABSTAIN);
}

/* -------------------------------------------------------------------------- */
/* Tests for composite_evaluator                                              */
/* -------------------------------------------------------------------------- */

UTEST(evaluator, test_composite_evaluator_null_input) {
  /* Test with NULL node */
  int res = composite_evaluator(NULL);
  ASSERT_EQ(res, PLCS_EVAL_RESULT_ABSTAIN);
}

UTEST(evaluator, test_composite_evaluator_basic_functionality) {
  /* Mock a CompositeNode with basic values */
  /* Note: This test requires proper mocking of FlatCC objects */

  /* Initialize context for evaluation */
  int rc = plcs_eval_ctx_init();
  ASSERT_TRUE(rc == PLCS_ESUCCESS || rc == PLCS_EINITIZLIED);

  /* Note: Full testing would require creating mock FlatCC objects */
  /* This test demonstrates the test structure for when headers are available */

  /* TODO: When FlatCC builder headers are available, create mock objects like:
   * - CompositeNode with op=BOOL_AND, children=[mock_evaluator_node]
   * - Test composite_evaluator() with the mock object
   * - Verify it applies boolean operations correctly
   */
}

/* -------------------------------------------------------------------------- */
/* Integration tests for boolean operations                                   */
/* -------------------------------------------------------------------------- */

UTEST(evaluator_integration, test_boolean_operation_truth_table) {
  /* Test comprehensive truth table for boolean operations */

  /* AND truth table */
  ASSERT_EQ(
      (int)DoOper(dd_ns(BoolOperation_BOOL_AND), PLCS_EVAL_RESULT_TRUE, PLCS_EVAL_RESULT_TRUE), PLCS_EVAL_RESULT_TRUE
  );
  ASSERT_EQ(
      (int)DoOper(dd_ns(BoolOperation_BOOL_AND), PLCS_EVAL_RESULT_TRUE, PLCS_EVAL_RESULT_FALSE), PLCS_EVAL_RESULT_FALSE
  );
  ASSERT_EQ(
      (int)DoOper(dd_ns(BoolOperation_BOOL_AND), PLCS_EVAL_RESULT_TRUE, PLCS_EVAL_RESULT_ABSTAIN),
      PLCS_EVAL_RESULT_ABSTAIN
  );
  ASSERT_EQ(
      (int)DoOper(dd_ns(BoolOperation_BOOL_AND), PLCS_EVAL_RESULT_FALSE, PLCS_EVAL_RESULT_TRUE), PLCS_EVAL_RESULT_FALSE
  );
  ASSERT_EQ(
      (int)DoOper(dd_ns(BoolOperation_BOOL_AND), PLCS_EVAL_RESULT_FALSE, PLCS_EVAL_RESULT_FALSE), PLCS_EVAL_RESULT_FALSE
  );
  ASSERT_EQ(
      (int)DoOper(dd_ns(BoolOperation_BOOL_AND), PLCS_EVAL_RESULT_FALSE, PLCS_EVAL_RESULT_ABSTAIN),
      PLCS_EVAL_RESULT_FALSE
  );
  ASSERT_EQ(
      (int)DoOper(dd_ns(BoolOperation_BOOL_AND), PLCS_EVAL_RESULT_ABSTAIN, PLCS_EVAL_RESULT_TRUE),
      PLCS_EVAL_RESULT_ABSTAIN
  );
  ASSERT_EQ(
      (int)DoOper(dd_ns(BoolOperation_BOOL_AND), PLCS_EVAL_RESULT_ABSTAIN, PLCS_EVAL_RESULT_FALSE),
      PLCS_EVAL_RESULT_FALSE
  );
  ASSERT_EQ(
      (int)DoOper(dd_ns(BoolOperation_BOOL_AND), PLCS_EVAL_RESULT_ABSTAIN, PLCS_EVAL_RESULT_ABSTAIN),
      PLCS_EVAL_RESULT_ABSTAIN
  );

  /* OR truth table */
  ASSERT_EQ(
      (int)DoOper(dd_ns(BoolOperation_BOOL_OR), PLCS_EVAL_RESULT_TRUE, PLCS_EVAL_RESULT_TRUE), PLCS_EVAL_RESULT_TRUE
  );
  ASSERT_EQ(
      (int)DoOper(dd_ns(BoolOperation_BOOL_OR), PLCS_EVAL_RESULT_TRUE, PLCS_EVAL_RESULT_FALSE), PLCS_EVAL_RESULT_TRUE
  );
  ASSERT_EQ(
      (int)DoOper(dd_ns(BoolOperation_BOOL_OR), PLCS_EVAL_RESULT_TRUE, PLCS_EVAL_RESULT_ABSTAIN), PLCS_EVAL_RESULT_TRUE
  );
  ASSERT_EQ(
      (int)DoOper(dd_ns(BoolOperation_BOOL_OR), PLCS_EVAL_RESULT_FALSE, PLCS_EVAL_RESULT_TRUE), PLCS_EVAL_RESULT_TRUE
  );
  ASSERT_EQ(
      (int)DoOper(dd_ns(BoolOperation_BOOL_OR), PLCS_EVAL_RESULT_FALSE, PLCS_EVAL_RESULT_FALSE), PLCS_EVAL_RESULT_FALSE
  );
  ASSERT_EQ(
      (int)DoOper(dd_ns(BoolOperation_BOOL_OR), PLCS_EVAL_RESULT_FALSE, PLCS_EVAL_RESULT_ABSTAIN),
      PLCS_EVAL_RESULT_ABSTAIN
  );
  ASSERT_EQ(
      (int)DoOper(dd_ns(BoolOperation_BOOL_OR), PLCS_EVAL_RESULT_ABSTAIN, PLCS_EVAL_RESULT_TRUE), PLCS_EVAL_RESULT_TRUE
  );
  ASSERT_EQ(
      (int)DoOper(dd_ns(BoolOperation_BOOL_OR), PLCS_EVAL_RESULT_ABSTAIN, PLCS_EVAL_RESULT_FALSE),
      PLCS_EVAL_RESULT_ABSTAIN
  );
  ASSERT_EQ(
      (int)DoOper(dd_ns(BoolOperation_BOOL_OR), PLCS_EVAL_RESULT_ABSTAIN, PLCS_EVAL_RESULT_ABSTAIN),
      PLCS_EVAL_RESULT_ABSTAIN
  );
}

/* -------------------------------------------------------------------------- */
/* Additional edge case tests                                                  */
/* -------------------------------------------------------------------------- */

UTEST(evaluator, test_boolean_operations_edge_cases) {
  /* Test edge cases and boundary conditions */

  /* Test with invalid enum values */
  int res = DoOper(99, PLCS_EVAL_RESULT_TRUE, PLCS_EVAL_RESULT_TRUE);
  ASSERT_EQ(res, PLCS_EVAL_RESULT_ABSTAIN);

  res = DoOper(-1, PLCS_EVAL_RESULT_FALSE, PLCS_EVAL_RESULT_FALSE);
  ASSERT_EQ(res, PLCS_EVAL_RESULT_ABSTAIN);

  /* Test with BOOL_COUNT (should be invalid) */
  res = DoOper(dd_ns(BoolOperation_BOOL_COUNT), PLCS_EVAL_RESULT_TRUE, PLCS_EVAL_RESULT_TRUE);
  ASSERT_EQ(res, PLCS_EVAL_RESULT_ABSTAIN);
}

UTEST(evaluator, test_boolean_operations_commutativity) {
  /* Test that AND and OR operations are commutative */

  /* AND commutativity */
  ASSERT_EQ(DoAnd(PLCS_EVAL_RESULT_TRUE, PLCS_EVAL_RESULT_FALSE), DoAnd(PLCS_EVAL_RESULT_FALSE, PLCS_EVAL_RESULT_TRUE));
  ASSERT_EQ(
      DoAnd(PLCS_EVAL_RESULT_TRUE, PLCS_EVAL_RESULT_ABSTAIN), DoAnd(PLCS_EVAL_RESULT_ABSTAIN, PLCS_EVAL_RESULT_TRUE)
  );
  ASSERT_EQ(
      DoAnd(PLCS_EVAL_RESULT_FALSE, PLCS_EVAL_RESULT_ABSTAIN), DoAnd(PLCS_EVAL_RESULT_ABSTAIN, PLCS_EVAL_RESULT_FALSE)
  );

  /* OR commutativity */
  ASSERT_EQ(DoOr(PLCS_EVAL_RESULT_TRUE, PLCS_EVAL_RESULT_FALSE), DoOr(PLCS_EVAL_RESULT_FALSE, PLCS_EVAL_RESULT_TRUE));
  ASSERT_EQ(
      DoOr(PLCS_EVAL_RESULT_TRUE, PLCS_EVAL_RESULT_ABSTAIN), DoOr(PLCS_EVAL_RESULT_ABSTAIN, PLCS_EVAL_RESULT_TRUE)
  );
  ASSERT_EQ(
      DoOr(PLCS_EVAL_RESULT_FALSE, PLCS_EVAL_RESULT_ABSTAIN), DoOr(PLCS_EVAL_RESULT_ABSTAIN, PLCS_EVAL_RESULT_FALSE)
  );
}

UTEST(evaluator, test_boolean_operations_associativity) {
  /* Test that AND and OR operations are associative */

  /* AND associativity: (a & b) & c = a & (b & c) */
  plcs_evaluation_result left = DoAnd(DoAnd(PLCS_EVAL_RESULT_TRUE, PLCS_EVAL_RESULT_FALSE), PLCS_EVAL_RESULT_ABSTAIN);
  plcs_evaluation_result right = DoAnd(PLCS_EVAL_RESULT_TRUE, DoAnd(PLCS_EVAL_RESULT_FALSE, PLCS_EVAL_RESULT_ABSTAIN));
  ASSERT_EQ(left, right);

  /* OR associativity: (a | b) | c = a | (b | c) */
  left = DoOr(DoOr(PLCS_EVAL_RESULT_FALSE, PLCS_EVAL_RESULT_TRUE), PLCS_EVAL_RESULT_ABSTAIN);
  right = DoOr(PLCS_EVAL_RESULT_FALSE, DoOr(PLCS_EVAL_RESULT_TRUE, PLCS_EVAL_RESULT_ABSTAIN));
  ASSERT_EQ(left, right);
}

UTEST(evaluator, test_boolean_operations_distributivity) {
  /* Test De Morgan's laws and distributivity */

  /* De Morgan's law: NOT(a AND b) = NOT(a) OR NOT(b) */
  plcs_evaluation_result left = DoNot(DoAnd(PLCS_EVAL_RESULT_TRUE, PLCS_EVAL_RESULT_FALSE));
  plcs_evaluation_result right = DoOr(DoNot(PLCS_EVAL_RESULT_TRUE), DoNot(PLCS_EVAL_RESULT_FALSE));
  ASSERT_EQ(left, right);

  /* De Morgan's law: NOT(a OR b) = NOT(a) AND NOT(b) */
  left = DoNot(DoOr(PLCS_EVAL_RESULT_TRUE, PLCS_EVAL_RESULT_FALSE));
  right = DoAnd(DoNot(PLCS_EVAL_RESULT_TRUE), DoNot(PLCS_EVAL_RESULT_FALSE));
  ASSERT_EQ(left, right);
}

UTEST(evaluator, test_extern_declarations_working) {
  /* Simple test to verify that extern declarations are working */
  /* This test calls the functions to ensure they can be linked */

  /* Test DoAnd with simple values */
  int res = DoAnd(PLCS_EVAL_RESULT_TRUE, PLCS_EVAL_RESULT_TRUE);
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);

  /* Test DoOr with simple values */
  res = DoOr(PLCS_EVAL_RESULT_FALSE, PLCS_EVAL_RESULT_TRUE);
  ASSERT_EQ(res, PLCS_EVAL_RESULT_TRUE);

  /* Test DoNot with simple values */
  res = DoNot(PLCS_EVAL_RESULT_TRUE);
  ASSERT_EQ(res, PLCS_EVAL_RESULT_FALSE);

  /* Test DoOper with AND operation */
  res = DoOper(dd_ns(BoolOperation_BOOL_AND), PLCS_EVAL_RESULT_TRUE, PLCS_EVAL_RESULT_FALSE);
  ASSERT_EQ(res, PLCS_EVAL_RESULT_FALSE);
}

/*
 * End-to-end evaluation test using a generated FlatBuffers header.
 *
 * The generated buffer contains several simple policies of the form:
 *   if RUNTIME_LANGUAGE == "<lang>" then execute action INJECT_DENY
 *
 * We set the runtime language in the evaluation context, register action handlers,
 * and verify that actions are executed (regardless of TRUE/FALSE - the engine
 * passes the plcs_evaluation_result to the action and the action may decide what to do).
 */
UTEST(evaluator_integration, evaluate_generated_header_if_available) {
  /* Initialize context and register actions */
  int rc = plcs_eval_ctx_init();
  /* init; accept already-inited code path as well */
  ASSERT_TRUE(rc == PLCS_ESUCCESS || rc == PLCS_EINITIZLIED);

  g_allow_called = 0;
  g_deny_called = 0;

  int prc = plcs_eval_ctx_register_action(test_action_allow, PLCS_ACTION_INJECT_ALLOW);
  ASSERT_EQ(prc, PLCS_ESUCCESS);
  prc = plcs_eval_ctx_register_action(test_action_deny, PLCS_ACTION_INJECT_DENY);
  ASSERT_EQ(prc, PLCS_ESUCCESS);

  /* Provide context parameter used by policies (runtime language) */
  prc = plcs_eval_ctx_set_str_eval_param(PLCS_STR_EVAL_RUNTIME_LANGUAGE, "jvm");
  ASSERT_EQ(prc, PLCS_ESUCCESS);

  /* Evaluate the embedded buffer */
  int eval_rc = plcs_evaluate_buffer(hardcoded_policies, hardcoded_policies_len);

  /* Non-zero would indicate action failures; we expect success. */
  ASSERT_EQ(eval_rc, PLCS_ESUCCESS);

  /* We expect at least one DENY action to have been invoked. */
  ASSERT_TRUE(g_deny_called >= 1);

#if HAVE_HARDCODED_POLICIES_HEADER
  /* No generated header available; this test is a stub. */
  ASSERT_TRUE(1);
#endif
}
