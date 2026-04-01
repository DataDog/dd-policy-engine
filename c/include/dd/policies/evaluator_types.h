/*
 * Unless explicitly stated otherwise all files in this repository are licensed
 * under the Apache 2.0 License. This product includes software developed at
 * Datadog (https://www.datadoghq.com/).
 *
 * Copyright 2025-Present Datadog, Inc.
 */
#pragma once

#include "evaluation_result.h"

/**
 * @brief these are public represntations of the wire (flatbuffers) enums used in the policy engine
 */

/**
 * @brief comparison operators for string evaluators
 */
#define PLCS_LIST_STRING_COMPARATORS(X)                                                                                \
  X(STR_UNKNOWN, 0)                                                                                                    \
  X(PREFIX, 1)                                                                                                         \
  X(SUFFIX, 2)                                                                                                         \
  X(CONTAINS, 3)                                                                                                       \
  X(EXACT, 4)                                                                                                          \
  X(WILDCARD, 5)

#define ENUM_COMPARATOR_VAL(VAL, IX) PLCS_STR_CMP_##VAL = IX,
typedef enum plcs_string_comparator {
  PLCS_LIST_STRING_COMPARATORS(ENUM_COMPARATOR_VAL) PLCS_STR_CMP__COUNT
} plcs_string_comparator;
#undef ENUM_COMPARATOR_VAL

/**
 * @brief comparison operators for numeric (and unsigned numeric) evaluators
 */
#define PLCS_LIST_NUMERIC_COMPARATOR(X)                                                                                \
  X(NUM_UNKNOWN, 0)                                                                                                    \
  X(EQ, 1)                                                                                                             \
  X(GT, 2)                                                                                                             \
  X(GTE, 3)                                                                                                            \
  X(LT, 4)                                                                                                             \
  X(LTE, 5)

#define ENUM_NUMERIC_VAL(VAL, IX) PLCS_NUM_CMP_##VAL = IX,
typedef enum plcs_numeric_comparator {
  PLCS_LIST_NUMERIC_COMPARATOR(ENUM_NUMERIC_VAL) PLCS_NUM_CMP__COUNT
} plcs_numeric_comparator;
#undef ENUM_NUMERIC_VAL

/**
 * @brief string evaluators
 * These represent the supported string evaluators by the policy engine.
 */
#define PLCS_LIST_STRING_EVALUATORS(X)                                                                                 \
  X(STRING_EVAL_UNKNOWN, 0)                                                                                            \
  X(COMPONENT, 1)                                                                                                      \
  X(PROCESS_EXE, 2)                                                                                                    \
  X(PROCESS_EXE_FULL_PATH, 3)                                                                                          \
  X(PROCESS_BASEDIR_PATH, 4)                                                                                           \
  X(PROCESS_ARGV, 5)                                                                                                   \
  X(PROCESS_CWD, 6)                                                                                                    \
  X(RUNTIME_LANGUAGE, 7)                                                                                               \
  X(RUNTIME_ENTRY_POINT_FILE, 8)                                                                                       \
  X(RUNTIME_ENTRY_POINT_JAR, 9)                                                                                        \
  X(RUNTIME_ENTRY_POINT_CLASS, 10)                                                                                     \
  X(RUNTIME_ENTRY_POINT_PACKAGE, 11)                                                                                   \
  X(RUNTIME_ENTRY_POINT_MODULE, 12)                                                                                    \
  X(RUNTIME_ENTRY_POINT_SOURCE, 13)                                                                                    \
  X(RUNTIME_DOPTION, 14)                                                                                               \
  X(RUNTIME_VERSION, 15)                                                                                               \
  X(LIBC_FLAVOR, 16)                                                                                                   \
  X(LIBC_VERSION, 17)                                                                                                  \
  X(MACHINE_ARCHITECTURE, 18)                                                                                          \
  X(HOST_NAME, 19)                                                                                                     \
  X(HOST_IP, 20)                                                                                                       \
  X(OS, 21)                                                                                                            \
  X(OS_DISTRO, 22)                                                                                                     \
  X(OS_DISTRO_VERSION, 23)                                                                                             \
  X(OS_DISTRO_CODENAME, 24)                                                                                            \
  X(OS_KERNEL_VERSION, 25)                                                                                             \
  X(OS_KERNEL_NAME, 26)                                                                                                \
  X(OS_USER, 27)                                                                                                       \
  X(OS_USER_GROUP, 28)                                                                                                 \
  X(CONTAINER_IMAGE, 29)                                                                                               \
  X(CONTAINER_ID, 30)                                                                                                  \
  X(ALWAYS_TRUE, 31)                                                                                                   \
  X(ALWAYS_FALSE, 32)                                                                                                  \
  X(ALWAYS_ABSTAIN, 33)                                                                                                \
  X(IIS_APPLICATION_POOL, 34)                                                                                          \
  X(PROCESS_ARGV_0, 35)                                                                                                \
  X(PROCESS_ARGV_1, 36)                                                                                                \
  X(PROCESS_ARGV_2, 37)                                                                                                \
  X(PROCESS_ARGV_3, 38)                                                                                                \
  X(PROCESS_ARGV_4, 39)                                                                                                \
  X(PROCESS_ARGV_5, 40)                                                                                                \
  X(PROCESS_ARGV_N, 41)                                                                                                \
  X(PROCESS_ARGV_N_2, 42)                                                                                 \
  X(PROCESS_ARGV_N_3, 43)                                                                                 \
  X(PROCESS_ARGV_N_4, 44)                                                                                 \
  X(PROCESS_ARGV_N_5, 45)                                                                                 \
  X(PROCESS_ARGV_N_6, 46)                                                                                 \
  X(PROCESS_ENVAR, 47)

#define ENUM_STR_CMP_EVAL(ID, IX) PLCS_STR_EVAL_##ID = IX,
typedef enum plcs_string_evaluators {
  PLCS_LIST_STRING_EVALUATORS(ENUM_STR_CMP_EVAL) PLCS_STR_EVAL__COUNT
} plcs_string_evaluators;
#undef ENUM_STR_CMP_EVAL

/**
 * @brief numeric evaluators
 * These represent the supported numeric evaluators by the policy engine.
 */
#define PLCS_LIST_NUMERIC_EVALUATORS(X)                                                                                \
  X(NUMERIC_EVAL_UNKNOWN, 0)                                                                                           \
  X(JAVA_HEAP, 1)                                                                                                      \
  X(RUNTIME_VERSION_MAJOR, 2)                                                                                          \
  X(RUNTIME_VERSION_MINOR, 3)                                                                                          \
  X(RUNTIME_VERSION_PATCH, 4)                                                                                          \
  X(OS_DISTRO_VERSION_MAJOR, 5)                                                                                        \
  X(OS_DISTRO_VERSION_MINOR, 6)                                                                                        \
  X(OS_DISTRO_VERSION_PATCH, 7)                                                                                        \
  X(OS_KERNEL_VERSION_MAJOR, 8)                                                                                        \
  X(OS_KERNEL_VERSION_MINOR, 9)                                                                                        \
  X(OS_KERNEL_VERSION_PATCH, 10)                                                                                       \
  X(LIBC_VERSION_MAJOR, 11)                                                                                            \
  X(LIBC_VERSION_MINOR, 12)                                                                                            \
  X(LIBC_VERSION_PATCH, 13)

#define ENUM_STR_EVAL(ID, IX) PLCS_NUM_EVAL_##ID = IX,
typedef enum plcs_numeric_evaluators {
  PLCS_LIST_NUMERIC_EVALUATORS(ENUM_STR_EVAL) PLCS_NUM_EVAL__COUNT
} plcs_numeric_evaluators;
#undef ENUM_STR_EVAL

/**
 * @brief A signature for string evaluator functions.
 *
 */
typedef plcs_evaluation_result (*plcs_string_evaluator_function_ptr)(
    const char *policy,
    const plcs_string_comparator cmp,
    const char *ctx,
    const char *description,
    plcs_string_evaluators eval_id
);

/**
 * @brief A signature for numeric evaluator functions.
 *
 */
typedef plcs_evaluation_result (*plcs_numeric_evaluator_function_ptr)(
    const long policy,
    const plcs_numeric_comparator cmp,
    const long ctx,
    const char *description,
    plcs_numeric_evaluators eval_id
);

/**
 * @brief A signature for unsigned numeric evaluator functions.
 *
 */
typedef plcs_evaluation_result (*plcs_unumeric_evaluator_function_ptr)(
    const unsigned long policy,
    const plcs_numeric_comparator cmp,
    const unsigned long ctx,
    const char *description,
    plcs_numeric_evaluators eval_id
);

/**
 * @brief Converts a plcs_string_evaluators enum to a string representation.
 * @param eval_id The plcs_string_evaluators enum value.
 * @return A string representation of the evaluator.
 */
const char *plcs_string_evaluators_to_string(enum plcs_string_evaluators eval_id);

/**
 * @brief Converts a plcs_numeric_evaluators enum to a string representation.
 * @param eval_id The plcs_numeric_evaluators enum value.
 * @return A string representation of the evaluator.
 */
const char *plcs_numeric_evaluators_to_string(enum plcs_numeric_evaluators eval_id);

/**
 * @brief Converts a plcs_string_comparator enum to a string representation.
 * @param cmp The plcs_string_comparator enum value.
 * @return A string representation of the comparator.
 */
const char *plcs_string_comparator_to_string(enum plcs_string_comparator cmp);

/**
 * @brief Converts a plcs_numeric_comparator enum to a string representation.
 * @param cmp The plcs_numeric_comparator enum value.
 * @return A string representation of the comparator.
 */
const char *plcs_numeric_comparator_to_string(enum plcs_numeric_comparator cmp);
