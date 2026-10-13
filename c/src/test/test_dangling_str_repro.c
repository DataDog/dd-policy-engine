/*
 * Unless explicitly stated otherwise all files in this repository are licensed
 * under the Apache 2.0 License. This product includes software developed at
 * Datadog (https://www.datadoghq.com/).
 *
 * Copyright 2025-Present Datadog, Inc.
 * ----
 * Reproducer for the dangling-string-parameter hypothesis (v0.1.12).
 *
 * Hypothesis: plcs_eval_ctx_set_str_eval_param stores raw pointers — the
 * caller owns the buffer (per the public-header contract). If the caller's
 * buffer is freed or goes out of scope before evaluation, the engine reads
 * dangling memory, producing arch-dependent and potentially deterministic
 * crashes (consistent with the v0.58.1/v0.60.0 launcher arm64 SIGSEGV).
 *
 * The matching fix lives on dmehala/copy-str-params (PRs #8, #29, both
 * closed-not-merged): set_str_eval_param strdup()s and reset/set free
 * the previous copy.
 *
 * These tests fail under -fsanitize=address on main, and pass on the fix
 * branch:
 *
 *   1) repro_set_param_then_free_then_get
 *      Pure API-level UAF: set, free, get. Expect heap-use-after-free
 *      when reading the returned pointer.
 *
 *   2) repro_set_param_then_free_then_evaluate
 *      End-to-end through plcs_evaluate_buffer with a tiny in-memory
 *      policy that exercises the wildcard evaluator on PROCESS_ENVAR.
 *      Expect the UAF inside string_evaluator_wildcard's read of *str.
 *
 * Build with the clang-dev preset (-fsanitize=address,leak,undefined)
 * to observe the failures clearly. Without ASan, glibc may "luckily"
 * keep the freed memory readable and the bug may not surface.
 */
#define _GNU_SOURCE

#include "utest/utest.h"

#include <dd/policies/error_codes.h>
#include <dd/policies/eval_ctx.h>
#include <dd/policies/evaluator_default.h>
#include <dd/policies/policies.h>

#include "actions_builder.h"
#include "eval_ctx.h"
#include "evaluators_builder.h"
#include "evaluators_verifier.h"
#include "flatbuffers_common_builder.h"
#include "nodes_builder.h"
#include "policy.h"            /* plcs_get_policies */
#include "policy_builder.h"
#include "policy_verifier.h"
#include "wire/dd_types.h"

#include <stddef.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

/* -------------------------------------------------------------------------- */
/* Test 1: API-level reproducer                                                */
/* -------------------------------------------------------------------------- */

UTEST(repro_dangling_str, set_param_then_free_then_get) {
  (void)plcs_eval_ctx_init();
  plcs_eval_ctx_reset();

  /* Heap-allocated string the engine will be told to remember. */
  char *envar = strdup("FOO=bar");
  ASSERT_TRUE(envar != NULL);

  int rc = plcs_eval_ctx_set_str_eval_param(PLCS_STR_EVAL_PROCESS_ENVAR, envar);
  ASSERT_EQ(rc, PLCS_ESUCCESS);

  /* Caller frees its buffer. The engine still holds the (now-dangling)
   * pointer — this is the documented contract on main, the very thing
   * dmehala/copy-str-params would change. */
  free(envar);

  /* Read what the engine kept. On main this returns the freed pointer.
   * Touching it triggers ASan heap-use-after-free.
   * On the fix branch the engine returns its own copy and the read is fine. */
  const char *kept = plcs_eval_ctx_get_string_param(PLCS_STR_EVAL_PROCESS_ENVAR);
  ASSERT_TRUE(kept != NULL);

  /* This is the line that should trip ASan on main. */
  size_t len = strlen(kept);
  ASSERT_TRUE(len > 0);

  /* If we made it here under ASan, the fix is in place. */
  plcs_eval_ctx_reset();
}

/* -------------------------------------------------------------------------- */
/* Test 2: End-to-end reproducer through plcs_evaluate_buffer                  */
/* -------------------------------------------------------------------------- */

/* Build a minimal Policies buffer:
 *   Policies { policies: [
 *     Policy { rules: NodeTypeWrapper{ EvaluatorNode{
 *                StrEvaluator(id=PROCESS_ENVAR, cmp=WILDCARD, value="FOO=*")
 *              }}}
 *   ]}
 *
 * The policy has no actions — we only care about the rule evaluation
 * path that consumes the eval_ctx string parameter.
 */
#define BAIL(step)                                                  \
  do {                                                              \
    flatcc_builder_clear(&b);                                       \
    return NULL;                                                    \
  } while (0)

static void *build_envar_wildcard_policy(size_t *out_size) {
  flatcc_builder_t b;
  if (flatcc_builder_init(&b) != 0) return NULL;

  flatbuffers_string_ref_t pat = flatbuffers_string_create_str(&b, "FOO=*");
  if (!pat) BAIL("string_create_str");

  /* StrEvaluator: PROCESS_ENVAR  WILDCARD  "FOO=*" */
  dd_wls_StrEvaluator_ref_t str_eval = dd_wls_StrEvaluator_create(
      &b, dd_wls_StringEvaluators_PROCESS_ENVAR, dd_wls_CmpTypeSTR_CMP_WILDCARD, pat
  );
  if (!str_eval) BAIL("StrEvaluator_create");

  /* The flatcc *_create helpers reject 0/NULL refs for ALL fields (incl.
   * optional ones), so we provide non-empty placeholder strings instead of
   * 0 wherever a ref is required. */
  flatbuffers_string_ref_t eval_desc = flatbuffers_string_create_str(&b, "");
  if (!eval_desc) BAIL("eval_desc string");
  flatbuffers_string_ref_t pol_desc = flatbuffers_string_create_str(&b, "");
  if (!pol_desc) BAIL("pol_desc string");

  /* EvaluatorNode wrapping the StrEvaluator. */
  dd_wls_EvaluatorNode_ref_t eval_node =
      dd_wls_EvaluatorNode_create(&b, eval_desc, dd_wls_EvaluatorType_as_StrEvaluator(str_eval));
  if (!eval_node) BAIL("EvaluatorNode_create");

  /* NodeTypeWrapper { node = EvaluatorNode } — this is the rules tree. */
  dd_wls_NodeTypeWrapper_ref_t rules =
      dd_wls_NodeTypeWrapper_create(&b, dd_wls_NodeType_as_EvaluatorNode(eval_node));
  if (!rules) BAIL("NodeTypeWrapper_create");

  /* Empty Action vector for actions field. */
  if (dd_wls_Action_vec_start(&b) != 0) BAIL("Action_vec_start");
  dd_wls_Action_vec_ref_t actions_vec = dd_wls_Action_vec_end(&b);
  if (!actions_vec) BAIL("Action_vec_end");

  /* Policy. Policy_create unconditionally dereferences the UUID arg, so we
   * must supply a real (zeroed) UUID even if we don't care about it. */
  dd_wls_UUID_t zero_id = {0, 0};
  dd_wls_Policy_ref_t policy = dd_wls_Policy_create(
      &b, pol_desc, rules, actions_vec, &zero_id, /*version*/ 0
  );
  if (!policy) BAIL("Policy_create");

  /* Vector of policies (size 1). */
  if (dd_wls_Policy_vec_start(&b) != 0) BAIL("Policy_vec_start");
  if (!dd_wls_Policy_vec_push(&b, policy)) BAIL("Policy_vec_push");
  dd_wls_Policy_vec_ref_t policies_vec = dd_wls_Policy_vec_end(&b);
  if (!policies_vec) BAIL("Policy_vec_end");

  /* Policies as root via explicit start/add/end so the policies field
   * definitely lands on the table. (The _create_as_root convenience was
   * silently dropping the field for reasons that didn't need debugging.) */
  if (dd_wls_Policies_start_as_root(&b) != 0) BAIL("Policies_start_as_root");
  if (dd_wls_Policies_policies_add(&b, policies_vec) != 0) BAIL("Policies_policies_add");
  if (!dd_wls_Policies_end_as_root(&b)) BAIL("Policies_end_as_root");

  void *buf = flatcc_builder_finalize_buffer(&b, out_size);
  flatcc_builder_clear(&b);
  return buf;
}

UTEST(repro_dangling_str, set_param_then_free_then_evaluate) {
  (void)plcs_eval_ctx_init();
  plcs_eval_ctx_reset();

  size_t buf_size = 0;
  void *buf = build_envar_wildcard_policy(&buf_size);
  ASSERT_TRUE(buf != NULL);
  ASSERT_TRUE(buf_size > 0);

  /* Sanity: the buffer must round-trip through the verifier and yield a
   * non-empty policies vector. If this fires, the test is a no-op for
   * the UAF (plcs_evaluate_buffer would return ENO_DATA early). */
  ASSERT_EQ(dd_wls_Policies_verify_as_root(buf, buf_size), 0);
  dd_ns(Policy_vec_t) pols = plcs_get_policies((const uint8_t *)buf, buf_size);
  ASSERT_TRUE(pols != NULL);
  ASSERT_EQ(dd_ns(Policy_vec_len)(pols), (size_t)1);

  /* Heap string passed in, then freed before evaluate. */
  char *envar = strdup("FOO=bar");
  ASSERT_TRUE(envar != NULL);

  int rc = plcs_eval_ctx_set_str_eval_param(PLCS_STR_EVAL_PROCESS_ENVAR, envar);
  ASSERT_EQ(rc, PLCS_ESUCCESS);

  free(envar);

  /* On main: the wildcard evaluator dereferences a freed pointer.
   *   string_evaluator_wildcard(pattern="FOO=*", str=<dangling>) -> UAF
   * On dmehala/copy-str-params: the engine has its own copy → clean run.
   *
   * We don't assert on the return value: the point is that ASan must
   * either fire (main) or not (fix branch). */
  (void)plcs_evaluate_buffer((const uint8_t *)buf, buf_size);

  free(buf);
  plcs_eval_ctx_reset();
}
