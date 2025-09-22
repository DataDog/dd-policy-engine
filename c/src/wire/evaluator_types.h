/**
 * Unless explicitly stated otherwise all files in this repository are licensed
 * under the Apache 2.0 License. This product includes software developed at
 * Datadog (https://www.datadoghq.com/).
 *
 * Copyright 2025-Present Datadog, Inc.
 * -----
 * @file wire/evaluation_types.h
 * @brief Internal enum translation layer for the library.
 *
 * @details
 * This header bridges our public-facing enums in `../../include/evaluation_types.h`
 * with the on-the-wire / generated enum values from the FlatBuffers definitions (via `evaluators_reader.h`).
 *
 * It provides static inline helpers to:
 *   - Translate public enum values to the vendor/FlatBuffers integer values
 *   - Translate vendor/FlatBuffers integer values back to public enums
 *   - Sanity check table sizes against enum counts at compile time
 *
 * @note
 * This is **private** library glue — it is not installed, not exported,
 * and must not appear in any public-facing headers.
 * If you are writing code outside the library, include the public header
 * from `../../include/` instead and use only the public enum names.
 *
 * We own the public enum names and values; upstream values may change,
 * and this file is where we adapt to those changes.
 *
 * Guidelines:
 *   - No clever macros in public headers
 *   - Use `_Static_assert` and `-Wswitch-enum` to catch drift early
 *   - Update translation tables and asserts when public or vendor enums change
 *
 * @todo write a script that generate these files
 */
#pragma once

/* PRIVATE header: only used when building the library. Not installed. */

#include <dd/policies/evaluator_types.h>
#include <evaluators_reader.h> /* flatbuffers generated headers */
#include "dd_types.h"          /* dd_ns(...) */

#ifdef __cplusplus
extern "C" {
#endif

static inline dd_ns(CmpTypeSTR_enum_t) dd_strcmp_to_wire(enum plcs_string_comparator v) {
/* Map your public enum -> vendor numeric value. */
#define ENUM_VAL(ID, X) [PLCS_STR_CMP_##ID] = dd_ns(CmpTypeSTR_CMP_##ID),
  static const int map[PLCS_STR_CMP__COUNT] = {PLCS_LIST_STRING_COMPARATORS(ENUM_VAL)};
#undef ENUM_VAL
  _Static_assert(
      PLCS_STR_CMP__COUNT == dd_ns(CmpTypeSTR_CMP_COUNT),
      "update dd_strcmp_to_wire & plcs_string_comparator mappings when you modify CmpTypeSTR"
  );
  return (dd_ns(CmpTypeSTR_enum_t))((unsigned)v < PLCS_STR_CMP__COUNT ? map[v] : -1);
}

static inline dd_ns(CmpTypeNUM_enum_t) dd_numcmp_to_wire(enum plcs_numeric_comparator v) {
#define ENUM_VAL(ID, X) [PLCS_NUM_CMP_##ID] = dd_ns(CmpTypeNUM_CMP_##ID),
  static const int map[PLCS_NUM_CMP__COUNT] = {PLCS_LIST_NUMERIC_COMPARATOR(ENUM_VAL)};
#undef ENUM_VAL
  _Static_assert(
      PLCS_NUM_CMP__COUNT == dd_ns(CmpTypeNUM_CMP_COUNT),
      "update dd_numcmp_to_wire & plcs_numeric_comparator mappings when you modify CmpTypeNUM"
  );
  return (dd_ns(CmpTypeNUM_enum_t))((unsigned)v < PLCS_NUM_CMP__COUNT ? map[v] : -1);
}

static inline dd_ns(StringEvaluators_enum_t) dd_streval_to_wire(enum plcs_string_evaluators v) {
  /* Keep indices aligned with your public enum order. */
#define ENUM_VAL(ID, X) [PLCS_STR_EVAL_##ID] = dd_ns(StringEvaluators_##ID),
  static const int map[PLCS_STR_EVAL__COUNT] = {PLCS_LIST_STRING_EVALUATORS(ENUM_VAL)};
#undef ENUM_VAL
  _Static_assert(
      PLCS_STR_EVAL__COUNT == dd_ns(StringEvaluators_STR_EVAL_COUNT),
      "update dd_streval_to_wire & plcs_string_evaluators mappings when you modify StringEvaluators"
  );
  return (dd_ns(StringEvaluators_enum_t))((unsigned)v < PLCS_STR_EVAL__COUNT ? map[v] : -1);
}

static inline dd_ns(NumericEvaluators_enum_t) dd_numeval_to_wire(enum plcs_numeric_evaluators v) {
#define ENUM_VAL(ID, X) [PLCS_NUM_EVAL_##ID] = dd_ns(NumericEvaluators_##ID),
  static const int map[PLCS_NUM_EVAL__COUNT] = {PLCS_LIST_NUMERIC_EVALUATORS(ENUM_VAL)};
#undef ENUM_VAL
  _Static_assert(
      PLCS_NUM_EVAL__COUNT == dd_ns(NumericEvaluators_NUM_EVAL_COUNT),
      "update dd_numeval_to_wire & plcs_numeric_evaluators mappings when you modify NumericEvaluators"
  );
  return (dd_ns(NumericEvaluators_enum_t))((unsigned)v < PLCS_NUM_EVAL__COUNT ? map[v] : -1);
}

#ifdef __cplusplus
}
#endif
