/**
 * @file wire/evaluation_result.h
 * @brief Internal translation for plcs_evaluation_result enum.
 *
 * @details
 * This header maps our public-facing `plcs_evaluation_result` enum in
 *   `../../include/evaluation_result.h`
 * to the on-the-wire / generated enum values from the FlatBuffers schema
 * (via `evaluation_result_reader.h`).
 *
 * It provides:
 *   - Static inline helpers to translate public values to vendor/FlatBuffers values
 *   - Static inline helpers to translate vendor/FlatBuffers values back to public values
 *   - Compile-time checks to ensure the mappings remain in sync
 *
 * @note
 * This is **private** library code — it is not installed, not exported,
 * and must not be included in public-facing headers.
 * Outside the library, include the public header from `../../include/`
 * and use only the public enum names and values.
 *
 * Guidelines:
 *   - Keep this mapping aligned with both the public enum and vendor schema
 *   - Use `_Static_assert` and `-Wswitch-enum` to catch drift early
 *   - No clever macros in public headers
 */

#pragma once

#include <dd/policies/evaluation_result.h> /* public-facing enum */
#include <evaluation_result_reader.h>      /* FlatBuffers generated headers */
#include "dd_types.h"                      /* dd_ns(...) */

#ifdef __cplusplus
extern "C" {
#endif

static inline dd_ns(EvaluationResult_enum_t) dd_evalresult_to_wire(enum plcs_evaluation_result v) {
#define ENUM_VAL(ID, X) [PLCS_EVAL_RESULT_##ID] = dd_ns(EvaluationResult_EVAL_RESULT_##ID),
  static const int map[PLCS_EVAL_RESULT__COUNT] = {PLCS_LIST_RESULT(ENUM_VAL)};
#undef ENUM_VAL

  _Static_assert(
      PLCS_EVAL_RESULT__COUNT == dd_ns(EvaluationResult_EVAL_RESULT_COUNT),
      "update dd_evalresult_to_wire & plcs_evaluation_result mappings when EvaluationResult enum changes"
  );
  return (dd_ns(EvaluationResult_enum_t))((unsigned)v < PLCS_EVAL_RESULT__COUNT ? map[v] : -1);
}

#ifdef __cplusplus
}
#endif
