/*
 * Unit test stubs for eval_ctx module using utest.h
 *
 * These tests exercise basic registration, parameter setting, getters,
 * bounds checking, and a couple of default evaluator sanity checks.
 *
 * Build system is expected to compile this alongside test.c
 * which provides UTEST_MAIN().
 */
#define _GNU_SOURCE
#include "utest/utest.h"

#include <dd/policies/policies.h>

#include "eval_ctx.h"

#include <stddef.h>
#include <stdint.h>
#include <string.h>

/* -------------------------------------------------------------------------- */
/* Dummy evaluators and action for testing                                     */
/* -------------------------------------------------------------------------- */

static plcs_evaluation_result dummy_str_eval(
    const char *policy,
    const plcs_string_comparator cmp,
    const char *ctx,
    const char *description,
    plcs_string_evaluators eval_id
) {
  (void)cmp;
  (void)description;
  (void)eval_id;

  if (!policy || !ctx) {
    return PLCS_EVAL_RESULT_ABSTAIN;
  }
  return (strcmp(policy, ctx) == 0) ? PLCS_EVAL_RESULT_TRUE : PLCS_EVAL_RESULT_FALSE;
}

static plcs_evaluation_result dummy_num_eval(
    const long policy,
    const plcs_numeric_comparator cmp,
    const long ctx,
    const char *description,
    plcs_numeric_evaluators eval_id
) {
  (void)cmp;
  (void)description;
  (void)eval_id;
  return (policy == ctx) ? PLCS_EVAL_RESULT_TRUE : PLCS_EVAL_RESULT_FALSE;
}

static plcs_evaluation_result dummy_unum_eval(
    const unsigned long policy,
    const plcs_numeric_comparator cmp,
    const unsigned long ctx,
    const char *description,
    plcs_numeric_evaluators eval_id
) {
  (void)cmp;
  (void)description;
  (void)eval_id;
  return (policy == ctx) ? PLCS_EVAL_RESULT_TRUE : PLCS_EVAL_RESULT_FALSE;
}

static int g_action_called = 0;
static plcs_errors
dummy_action(plcs_evaluation_result res, char *values[], size_t value_len, const char *description, int action_id) {
  (void)res;
  (void)values;
  (void)value_len;
  (void)description;
  (void)action_id;
  g_action_called++;
  return PLCS_ESUCCESS;
}

/* -------------------------------------------------------------------------- */
/* Tests                                                                       */
/* -------------------------------------------------------------------------- */

UTEST(eval_ctx, eval_ctx_init_double_init) {
  int r1 = plcs_eval_ctx_init();
  int r2 = plcs_eval_ctx_init();
  /* Depending on previous tests, r1 may already be -PLCS_EINITIZLIED. */
  ASSERT_TRUE((r1 == PLCS_ESUCCESS && r2 == PLCS_EINITIZLIED) || (r1 == PLCS_EINITIZLIED && r2 == PLCS_EINITIZLIED));
}

UTEST(eval_ctx, verify_error_handling_string_evaluator_and_param) {
  int rc = plcs_eval_ctx_register_str_evaluator(NULL, PLCS_STR_EVAL_RUNTIME_LANGUAGE);
  ASSERT_EQ(rc, PLCS_EREGISTER_EVAL_PTR);

  rc = plcs_eval_ctx_register_str_evaluator(dummy_str_eval, PLCS_STR_EVAL__COUNT);
  ASSERT_EQ(rc, PLCS_EIX_OVERFLOW);

  const char *param = plcs_eval_ctx_get_string_param(PLCS_STR_EVAL__COUNT);
  ASSERT_TRUE(param == PLCS_STR_NOT_SET);

  int err = plcs_eval_ctx_get_last_error();
  ASSERT_EQ(err, PLCS_EIX_OVERFLOW);
}

UTEST(eval_ctx, register_and_get_string_evaluator_and_param) {
  /* Ensure initialized */
  (void)plcs_eval_ctx_init();

  const char *ctx_value = "jvm";
  int rc = plcs_eval_ctx_register_str_evaluator(dummy_str_eval, PLCS_STR_EVAL_RUNTIME_LANGUAGE);
  ASSERT_EQ(rc, PLCS_ESUCCESS);

  rc = plcs_eval_ctx_set_str_eval_param(PLCS_STR_EVAL_RUNTIME_LANGUAGE, ctx_value);
  ASSERT_EQ(rc, PLCS_ESUCCESS);

  plcs_string_evaluator_function_ptr f = plcs_eval_ctx_get_string_evaluator(PLCS_STR_EVAL_RUNTIME_LANGUAGE);
  ASSERT_TRUE(f != NULL);

  const char *param = plcs_eval_ctx_get_string_param(PLCS_STR_EVAL__COUNT);
  ASSERT_TRUE(param == PLCS_STR_NOT_SET);
  int err = plcs_eval_ctx_get_last_error();
  ASSERT_EQ(err, PLCS_EIX_OVERFLOW);

  param = plcs_eval_ctx_get_string_param(PLCS_STR_EVAL_RUNTIME_LANGUAGE);
  ASSERT_TRUE(param != NULL);
  ASSERT_EQ(0, strcmp(param, ctx_value));

  /* Ensure evaluator behaves as expected */
  int r = f("jvm", PLCS_STR_CMP_EXACT, param, "desc", PLCS_STR_EVAL_RUNTIME_LANGUAGE);
  ASSERT_EQ(r, PLCS_EVAL_RESULT_TRUE);

  r = f("python", PLCS_STR_CMP_EXACT, param, "desc", PLCS_STR_EVAL_RUNTIME_LANGUAGE);
  ASSERT_EQ(r, PLCS_EVAL_RESULT_FALSE);
}

UTEST(eval_ctx, verify_error_handling_numeric_evaluator_and_param) {
  int rc = plcs_eval_ctx_register_num_evaluator(NULL, PLCS_NUM_EVAL_JAVA_HEAP);
  ASSERT_EQ(rc, PLCS_EREGISTER_EVAL_PTR);

  rc = plcs_eval_ctx_register_num_evaluator(dummy_num_eval, PLCS_NUM_EVAL__COUNT);
  ASSERT_EQ(rc, PLCS_EIX_OVERFLOW);

  long param = plcs_eval_ctx_get_numeric_param(PLCS_NUM_EVAL__COUNT);
  ASSERT_TRUE(param == PLCS_NUM_NOT_SET);

  int err = plcs_eval_ctx_get_last_error();
  ASSERT_EQ(err, PLCS_EIX_OVERFLOW);
}

UTEST(eval_ctx, register_and_get_numeric_evaluator_and_param) {
  (void)plcs_eval_ctx_init();

  long ctx_value = 42;
  int rc = plcs_eval_ctx_register_num_evaluator(dummy_num_eval, PLCS_NUM_EVAL_JAVA_HEAP);
  ASSERT_EQ(rc, PLCS_ESUCCESS);

  rc = plcs_eval_ctx_register_num_evaluator(NULL, PLCS_NUM_EVAL_JAVA_HEAP);
  ASSERT_EQ(rc, PLCS_EREGISTER_EVAL_PTR);

  rc = plcs_eval_ctx_register_num_evaluator(dummy_num_eval, PLCS_NUM_EVAL__COUNT);
  ASSERT_EQ(rc, PLCS_EIX_OVERFLOW);

  rc = plcs_eval_ctx_set_num_eval_param(PLCS_NUM_EVAL_JAVA_HEAP, ctx_value);
  ASSERT_EQ(rc, PLCS_ESUCCESS);

  rc = plcs_eval_ctx_set_num_eval_param(PLCS_NUM_EVAL__COUNT, ctx_value);
  ASSERT_EQ(rc, PLCS_EIX_OVERFLOW);

  plcs_numeric_evaluator_function_ptr f = plcs_eval_ctx_get_numeric_evaluator(PLCS_NUM_EVAL_JAVA_HEAP);
  ASSERT_TRUE(f != NULL);

  long param = plcs_eval_ctx_get_numeric_param(PLCS_NUM_EVAL_JAVA_HEAP);
  ASSERT_EQ(param, ctx_value);

  int r = f(42, PLCS_NUM_CMP_EQ, param, "desc", PLCS_NUM_EVAL_JAVA_HEAP);
  ASSERT_EQ(r, PLCS_EVAL_RESULT_TRUE);

  r = f(7, PLCS_NUM_CMP_EQ, param, "desc", PLCS_NUM_EVAL_JAVA_HEAP);
  ASSERT_EQ(r, PLCS_EVAL_RESULT_FALSE);
}

UTEST(eval_ctx, verify_error_handling_unumeric_evaluator_and_param) {
  int rc = plcs_eval_ctx_register_unum_evaluator(NULL, PLCS_NUM_EVAL_JAVA_HEAP);
  ASSERT_EQ(rc, PLCS_EREGISTER_EVAL_PTR);

  rc = plcs_eval_ctx_register_unum_evaluator(dummy_unum_eval, PLCS_NUM_EVAL__COUNT);
  ASSERT_EQ(rc, PLCS_EIX_OVERFLOW);

  unsigned long param = plcs_eval_ctx_get_unumeric_param(PLCS_NUM_EVAL__COUNT);
  ASSERT_TRUE(param == PLCS_UNUM_NOT_SET);

  int err = plcs_eval_ctx_get_last_error();
  ASSERT_EQ(err, PLCS_EIX_OVERFLOW);
}

UTEST(eval_ctx, register_and_get_unumeric_evaluator_and_param) {
  (void)plcs_eval_ctx_init();

  unsigned long ctx_value = 7ul;
  int rc = plcs_eval_ctx_register_unum_evaluator(dummy_unum_eval, PLCS_NUM_EVAL_RUNTIME_VERSION_MAJOR);
  ASSERT_EQ(rc, PLCS_ESUCCESS);

  rc = plcs_eval_ctx_register_unum_evaluator(NULL, PLCS_NUM_EVAL_RUNTIME_VERSION_MAJOR);
  ASSERT_EQ(rc, PLCS_EREGISTER_EVAL_PTR);

  rc = plcs_eval_ctx_register_unum_evaluator(dummy_unum_eval, PLCS_NUM_EVAL__COUNT);
  ASSERT_EQ(rc, PLCS_EIX_OVERFLOW);

  rc = plcs_eval_ctx_set_unum_eval_param(PLCS_NUM_EVAL_RUNTIME_VERSION_MAJOR, ctx_value);
  ASSERT_EQ(rc, PLCS_ESUCCESS);

  rc = plcs_eval_ctx_set_unum_eval_param(PLCS_NUM_EVAL__COUNT, ctx_value);
  ASSERT_EQ(rc, PLCS_EIX_OVERFLOW);

  plcs_unumeric_evaluator_function_ptr f = plcs_eval_ctx_get_unumeric_evaluator(PLCS_NUM_EVAL_RUNTIME_VERSION_MAJOR);
  ASSERT_TRUE(f != NULL);

  unsigned long param = plcs_eval_ctx_get_unumeric_param(PLCS_NUM_EVAL_RUNTIME_VERSION_MAJOR);
  ASSERT_EQ(param, ctx_value);

  int r = f(7ul, PLCS_NUM_CMP_EQ, param, "desc", PLCS_NUM_EVAL_RUNTIME_VERSION_MAJOR);
  ASSERT_EQ(r, PLCS_EVAL_RESULT_TRUE);

  r = f(8ul, PLCS_NUM_CMP_EQ, param, "desc", PLCS_NUM_EVAL_RUNTIME_VERSION_MAJOR);
  ASSERT_EQ(r, PLCS_EVAL_RESULT_FALSE);
}

UTEST(eval_ctx, register_and_invoke_action_pointer) {
  (void)plcs_eval_ctx_init();

  int rc = plcs_eval_ctx_register_action(dummy_action, PLCS_ACTION_INJECT_ALLOW);
  ASSERT_EQ(rc, PLCS_ESUCCESS);

  rc = plcs_eval_ctx_register_action(dummy_action, PLCS_ACTIONS__COUNT);
  ASSERT_EQ(rc, PLCS_EIX_OVERFLOW);

  plcs_action_function_ptr act = plcs_eval_ctx_get_action(PLCS_ACTION_INJECT_ALLOW);
  ASSERT_TRUE(act != NULL);

  g_action_called = 0;
  char *vals[] = {(char *)"v1", (char *)"v2"};
  rc = act(PLCS_EVAL_RESULT_TRUE, vals, 2, "desc", PLCS_ACTION_INJECT_ALLOW);
  ASSERT_EQ(rc, PLCS_ESUCCESS);
  ASSERT_EQ(g_action_called, 1);
}

UTEST(eval_ctx, bounds_checks_and_error_reporting) {
  (void)plcs_eval_ctx_init();

  /* Out of range register attempts should overflow */
  int rc = plcs_eval_ctx_register_str_evaluator(dummy_str_eval, (plcs_string_evaluators)PLCS_STR_EVAL__COUNT);
  ASSERT_EQ(rc, PLCS_EIX_OVERFLOW);

  rc = plcs_eval_ctx_set_str_eval_param((plcs_string_evaluators)PLCS_STR_EVAL__COUNT, "x");
  ASSERT_EQ(rc, PLCS_EIX_OVERFLOW);

  /* Getter with OOB should set last error to PLCS_EIX_OVERFLOW */
  plcs_string_evaluator_function_ptr f =
      plcs_eval_ctx_get_string_evaluator((plcs_string_evaluators)PLCS_STR_EVAL__COUNT);
  ASSERT_TRUE(f == NULL);
  int err = plcs_eval_ctx_get_last_error();
  ASSERT_EQ(err, PLCS_EIX_OVERFLOW);

  /* Same for numeric */
  plcs_numeric_evaluator_function_ptr nf =
      plcs_eval_ctx_get_numeric_evaluator((plcs_numeric_evaluators)PLCS_NUM_EVAL__COUNT);
  ASSERT_TRUE(nf == NULL);
  err = plcs_eval_ctx_get_last_error();
  ASSERT_EQ(err, PLCS_EIX_OVERFLOW);

  /* And unsigned numeric */
  plcs_unumeric_evaluator_function_ptr unf =
      plcs_eval_ctx_get_unumeric_evaluator((plcs_numeric_evaluators)PLCS_NUM_EVAL__COUNT);
  ASSERT_TRUE(unf == NULL);
  err = plcs_eval_ctx_get_last_error();
  ASSERT_EQ(err, PLCS_EIX_OVERFLOW);

  /* And actions */
  plcs_action_function_ptr act = plcs_eval_ctx_get_action((plcs_actions)PLCS_ACTIONS__COUNT);
  ASSERT_TRUE(act == NULL);
  err = plcs_eval_ctx_get_last_error();
  ASSERT_EQ(err, PLCS_EIX_OVERFLOW);
}

UTEST(eval_ctx, last_error_set_and_get) {
  (void)plcs_eval_ctx_init();

  plcs_eval_ctx_set_error(PLCS_EUNKNOWN_CMP);
  ASSERT_EQ(plcs_eval_ctx_peek_last_error(), (plcs_errors)PLCS_EUNKNOWN_CMP);
  ASSERT_EQ(plcs_eval_ctx_get_last_error(), (plcs_errors)PLCS_EUNKNOWN_CMP);
  /* After get_last_error(), it resets to success */
  ASSERT_EQ(plcs_eval_ctx_peek_last_error(), (plcs_errors)PLCS_ESUCCESS);
}

UTEST(eval_ctx, set_error_out_of_bound) {
  plcs_eval_ctx_set_action_error(PLCS_ACTIONS__COUNT, 0);
  int err = plcs_eval_ctx_get_last_error();
  ASSERT_EQ(err, PLCS_ESUCCESS);
}

/* -------------------------------------------------------------------------- */
/* Default evaluator sanity checks                                             */
/* -------------------------------------------------------------------------- */

UTEST(default_evaluators, string_comparators) {
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
      plcs_default_string_evaluator("abc", PLCS_STR_CMP_EXACT, "abcd", "d", PLCS_STR_EVAL_COMPONENT),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_FALSE
  );
  ASSERT_EQ(
      plcs_default_string_evaluator("ac", PLCS_STR_CMP_PREFIX, "abc", "d", PLCS_STR_EVAL_COMPONENT),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_FALSE
  );
  ASSERT_EQ(
      plcs_default_string_evaluator("ac", PLCS_STR_CMP_SUFFIX, "abc", "d", PLCS_STR_EVAL_COMPONENT),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_FALSE
  );
  ASSERT_EQ(
      plcs_default_string_evaluator("z", PLCS_STR_CMP_CONTAINS, "abc", "d", PLCS_STR_EVAL_COMPONENT),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_FALSE
  );

  /* ABSTAIN on missing data */
  ASSERT_EQ(
      plcs_default_string_evaluator(NULL, PLCS_STR_CMP_EXACT, "abc", "d", PLCS_STR_EVAL_COMPONENT),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_ABSTAIN
  );
  ASSERT_EQ(
      plcs_default_string_evaluator("abc", PLCS_STR_CMP_EXACT, NULL, "d", PLCS_STR_EVAL_COMPONENT),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_ABSTAIN
  );

  /* ERRORS */
  ASSERT_EQ(
      plcs_default_string_evaluator("abc", PLCS_STR_CMP__COUNT, "a", "d", PLCS_STR_EVAL_COMPONENT),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_ABSTAIN
  );
  ASSERT_EQ(
      plcs_default_string_evaluator("abc", PLCS_STR_CMP__COUNT + 1, "a", "d", PLCS_STR_EVAL_COMPONENT),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_ABSTAIN
  );
  int err = plcs_eval_ctx_get_str_eval_error(PLCS_STR_EVAL_COMPONENT);
  ASSERT_EQ(err, PLCS_EUNKNOWN_CMP);

  err = plcs_eval_ctx_get_str_eval_error(PLCS_STR_EVAL__COUNT);
  ASSERT_EQ(err, PLCS_EIX_OVERFLOW);
}

UTEST(default_evaluators, numeric_comparators) {
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

  ASSERT_EQ(
      plcs_default_numeric_evaluator(4, PLCS_NUM_CMP_EQ, 5, "d", PLCS_NUM_EVAL_JAVA_HEAP),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_FALSE
  );
  ASSERT_EQ(
      plcs_default_numeric_evaluator(4, PLCS_NUM_CMP_GT, 5, "d", PLCS_NUM_EVAL_JAVA_HEAP),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_FALSE
  );
  ASSERT_EQ(
      plcs_default_numeric_evaluator(6, PLCS_NUM_CMP_GTE, 7, "d", PLCS_NUM_EVAL_JAVA_HEAP),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_FALSE
  );
  ASSERT_EQ(
      plcs_default_numeric_evaluator(6, PLCS_NUM_CMP_LT, 5, "d", PLCS_NUM_EVAL_JAVA_HEAP),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_FALSE
  );
  ASSERT_EQ(
      plcs_default_numeric_evaluator(6, PLCS_NUM_CMP_LTE, 5, "d", PLCS_NUM_EVAL_JAVA_HEAP),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_FALSE
  );

  /* ERRORS */
  ASSERT_EQ(
      plcs_default_numeric_evaluator(1, PLCS_NUM_CMP__COUNT, 2, "d", PLCS_NUM_EVAL_JAVA_HEAP),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_ABSTAIN
  );
  ASSERT_EQ(
      plcs_default_numeric_evaluator(1, PLCS_NUM_CMP__COUNT + 1, 2, "d", PLCS_NUM_EVAL_JAVA_HEAP),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_ABSTAIN
  );
  int err = plcs_eval_ctx_get_num_eval_error(PLCS_NUM_EVAL_JAVA_HEAP);
  ASSERT_EQ(err, PLCS_EUNKNOWN_CMP);

  err = plcs_eval_ctx_get_num_eval_error(PLCS_NUM_EVAL__COUNT);
  ASSERT_EQ(err, PLCS_EIX_OVERFLOW);
}

UTEST(default_evaluators, unumeric_comparators) {
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

  ASSERT_EQ(
      plcs_default_unumeric_evaluator(4ul, PLCS_NUM_CMP_EQ, 5ul, "d", PLCS_NUM_EVAL_JAVA_HEAP),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_FALSE
  );
  ASSERT_EQ(
      plcs_default_unumeric_evaluator(4ul, PLCS_NUM_CMP_GT, 5ul, "d", PLCS_NUM_EVAL_JAVA_HEAP),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_FALSE
  );
  ASSERT_EQ(
      plcs_default_unumeric_evaluator(6ul, PLCS_NUM_CMP_GTE, 7ul, "d", PLCS_NUM_EVAL_JAVA_HEAP),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_FALSE
  );
  ASSERT_EQ(
      plcs_default_unumeric_evaluator(6ul, PLCS_NUM_CMP_LT, 5ul, "d", PLCS_NUM_EVAL_JAVA_HEAP),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_FALSE
  );
  ASSERT_EQ(
      plcs_default_unumeric_evaluator(6ul, PLCS_NUM_CMP_LTE, 5ul, "d", PLCS_NUM_EVAL_JAVA_HEAP),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_FALSE
  );

  /* ERRORS */
  ASSERT_EQ(
      plcs_default_unumeric_evaluator(1, PLCS_NUM_CMP__COUNT, 2, "d", PLCS_NUM_EVAL_JAVA_HEAP),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_ABSTAIN
  );
  ASSERT_EQ(
      plcs_default_unumeric_evaluator(1, PLCS_NUM_CMP__COUNT + 1, 2, "d", PLCS_NUM_EVAL_JAVA_HEAP),
      (plcs_evaluation_result)PLCS_EVAL_RESULT_ABSTAIN
  );
  int err = plcs_eval_ctx_get_unum_eval_error(PLCS_NUM_EVAL_JAVA_HEAP);
  ASSERT_EQ(err, PLCS_EUNKNOWN_CMP);

  err = plcs_eval_ctx_get_unum_eval_error(PLCS_NUM_EVAL__COUNT);
  ASSERT_EQ(err, PLCS_EIX_OVERFLOW);
}
