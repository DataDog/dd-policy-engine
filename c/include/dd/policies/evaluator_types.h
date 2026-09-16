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
 *
 * Each entry of the X macro lists below is (ID, wire value, human readable label). The label is
 * what the engine uses to describe a rule that matched, e.g. the PROCESS_EXE evaluator compared
 * with CMP_PREFIX reads as "process executable 'python3.11' is prefixed with 'python'".
 */

/**
 * @brief comparison operators for string evaluators
 */
#define PLCS_LIST_STRING_COMPARATORS(X)                                                                                \
  X(STR_UNKNOWN, 0, "compares to")                                                                                     \
  X(PREFIX, 1, "is prefixed with")                                                                                     \
  X(SUFFIX, 2, "is suffixed with")                                                                                     \
  X(CONTAINS, 3, "contains")                                                                                           \
  X(EXACT, 4, "is")                                                                                                    \
  X(WILDCARD, 5, "matches")

#define ENUM_COMPARATOR_VAL(VAL, IX, LABEL) PLCS_STR_CMP_##VAL = IX,
typedef enum plcs_string_comparator {
  PLCS_LIST_STRING_COMPARATORS(ENUM_COMPARATOR_VAL) PLCS_STR_CMP__COUNT
} plcs_string_comparator;
#undef ENUM_COMPARATOR_VAL

/**
 * @brief comparison operators for numeric (and unsigned numeric) evaluators
 */
#define PLCS_LIST_NUMERIC_COMPARATOR(X)                                                                                \
  X(NUM_UNKNOWN, 0, "compares to")                                                                                     \
  X(EQ, 1, "is")                                                                                                       \
  X(GT, 2, "is greater than")                                                                                          \
  X(GTE, 3, "is at least")                                                                                             \
  X(LT, 4, "is less than")                                                                                             \
  X(LTE, 5, "is at most")

#define ENUM_NUMERIC_VAL(VAL, IX, LABEL) PLCS_NUM_CMP_##VAL = IX,
typedef enum plcs_numeric_comparator {
  PLCS_LIST_NUMERIC_COMPARATOR(ENUM_NUMERIC_VAL) PLCS_NUM_CMP__COUNT
} plcs_numeric_comparator;
#undef ENUM_NUMERIC_VAL

/**
 * @brief string evaluators
 * These represent the supported string evaluators by the policy engine.
 */
#define PLCS_LIST_STRING_EVALUATORS(X)                                                                                 \
  X(STRING_EVAL_UNKNOWN, 0, "unknown value")                                                                           \
  X(COMPONENT, 1, "component")                                                                                         \
  X(PROCESS_EXE, 2, "process executable")                                                                              \
  X(PROCESS_EXE_FULL_PATH, 3, "process executable full path")                                                          \
  X(PROCESS_BASEDIR_PATH, 4, "process base directory path")                                                            \
  X(PROCESS_ARGV, 5, "process arguments")                                                                              \
  X(PROCESS_CWD, 6, "process working directory")                                                                       \
  X(RUNTIME_LANGUAGE, 7, "runtime language")                                                                           \
  X(RUNTIME_ENTRY_POINT_FILE, 8, "runtime entry point file")                                                           \
  X(RUNTIME_ENTRY_POINT_JAR, 9, "runtime entry point jar")                                                             \
  X(RUNTIME_ENTRY_POINT_CLASS, 10, "runtime entry point class")                                                        \
  X(RUNTIME_ENTRY_POINT_PACKAGE, 11, "runtime entry point package")                                                    \
  X(RUNTIME_ENTRY_POINT_MODULE, 12, "runtime entry point module")                                                      \
  X(RUNTIME_ENTRY_POINT_SOURCE, 13, "runtime entry point source")                                                      \
  X(RUNTIME_DOPTION, 14, "runtime -D option")                                                                          \
  X(RUNTIME_VERSION, 15, "runtime version")                                                                            \
  X(LIBC_FLAVOR, 16, "libc flavor")                                                                                    \
  X(LIBC_VERSION, 17, "libc version")                                                                                  \
  X(MACHINE_ARCHITECTURE, 18, "machine architecture")                                                                  \
  X(HOST_NAME, 19, "host name")                                                                                        \
  X(HOST_IP, 20, "host IP")                                                                                            \
  X(OS, 21, "operating system")                                                                                        \
  X(OS_DISTRO, 22, "OS distribution")                                                                                  \
  X(OS_DISTRO_VERSION, 23, "OS distribution version")                                                                  \
  X(OS_DISTRO_CODENAME, 24, "OS distribution codename")                                                                \
  X(OS_KERNEL_VERSION, 25, "OS kernel version")                                                                        \
  X(OS_KERNEL_NAME, 26, "OS kernel name")                                                                              \
  X(OS_USER, 27, "OS user")                                                                                            \
  X(OS_USER_GROUP, 28, "OS user group")                                                                                \
  X(CONTAINER_IMAGE, 29, "container image")                                                                            \
  X(CONTAINER_ID, 30, "container id")                                                                                  \
  X(ALWAYS_TRUE, 31, "always true")                                                                                    \
  X(ALWAYS_FALSE, 32, "always false")                                                                                  \
  X(ALWAYS_ABSTAIN, 33, "always abstain")                                                                              \
  X(IIS_APPLICATION_POOL, 34, "IIS application pool")                                                                  \
  X(PROCESS_ARGV_0, 35, "process argument 0")                                                                          \
  X(PROCESS_ARGV_1, 36, "process argument 1")                                                                          \
  X(PROCESS_ARGV_2, 37, "process argument 2")                                                                          \
  X(PROCESS_ARGV_3, 38, "process argument 3")                                                                          \
  X(PROCESS_ARGV_4, 39, "process argument 4")                                                                          \
  X(PROCESS_ARGV_5, 40, "process argument 5")                                                                          \
  X(PROCESS_ARGV_N, 41, "last process argument")                                                                       \
  X(PROCESS_ARGV_N_2, 42, "2nd to last process argument")                                                              \
  X(PROCESS_ARGV_N_3, 43, "3rd to last process argument")                                                              \
  X(PROCESS_ARGV_N_4, 44, "4th to last process argument")                                                              \
  X(PROCESS_ARGV_N_5, 45, "5th to last process argument")                                                              \
  X(PROCESS_ARGV_N_6, 46, "6th to last process argument")                                                              \
  X(PROCESS_ENVAR, 47, "process environment variable")                                                                 \
  X(CONTAINER_IMAGE_TAG, 48, "container image tag")                                                                    \
  X(CONTAINER_IMAGE_DIGEST, 49, "container image digest")                                                              \
  X(CONTAINER_NAME, 50, "container name")                                                                              \
  X(CONTAINER_LABEL, 51, "container label")                                                                            \
  X(NAMESPACE_NAME, 52, "namespace name")                                                                              \
  X(NAMESPACE_LABEL, 53, "namespace label")                                                                            \
  X(POD_LABEL, 54, "pod label")                                                                                        \
  X(POD_ANNOTATION, 55, "pod annotation")

#define ENUM_STR_CMP_EVAL(ID, IX, LABEL) PLCS_STR_EVAL_##ID = IX,
typedef enum plcs_string_evaluators {
  PLCS_LIST_STRING_EVALUATORS(ENUM_STR_CMP_EVAL) PLCS_STR_EVAL__COUNT
} plcs_string_evaluators;
#undef ENUM_STR_CMP_EVAL

/**
 * @brief numeric evaluators
 * These represent the supported numeric evaluators by the policy engine.
 */
#define PLCS_LIST_NUMERIC_EVALUATORS(X)                                                                                \
  X(NUMERIC_EVAL_UNKNOWN, 0, "unknown value")                                                                          \
  X(JAVA_HEAP, 1, "java heap")                                                                                         \
  X(RUNTIME_VERSION_MAJOR, 2, "runtime major version")                                                                 \
  X(RUNTIME_VERSION_MINOR, 3, "runtime minor version")                                                                 \
  X(RUNTIME_VERSION_PATCH, 4, "runtime patch version")                                                                 \
  X(OS_DISTRO_VERSION_MAJOR, 5, "OS distribution major version")                                                       \
  X(OS_DISTRO_VERSION_MINOR, 6, "OS distribution minor version")                                                       \
  X(OS_DISTRO_VERSION_PATCH, 7, "OS distribution patch version")                                                       \
  X(OS_KERNEL_VERSION_MAJOR, 8, "OS kernel major version")                                                             \
  X(OS_KERNEL_VERSION_MINOR, 9, "OS kernel minor version")                                                             \
  X(OS_KERNEL_VERSION_PATCH, 10, "OS kernel patch version")                                                            \
  X(LIBC_VERSION_MAJOR, 11, "libc major version")                                                                      \
  X(LIBC_VERSION_MINOR, 12, "libc minor version")                                                                      \
  X(LIBC_VERSION_PATCH, 13, "libc patch version")

#define ENUM_STR_EVAL(ID, IX, LABEL) PLCS_NUM_EVAL_##ID = IX,
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
