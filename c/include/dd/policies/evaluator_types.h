#pragma once

#include "evaluation_result.h"

/**
 * @brief these are public represntations of the wire (flatbuffers) enums used in the policy engine
 */

/**
 * @brief comparison operators for string evaluators
 */
#define PLCS_LIST_STRING_COMPARATORS(X)                                                                                \
  X(PREFIX, 0)                                                                                                         \
  X(SUFFIX, 1)                                                                                                         \
  X(CONTAINS, 2)                                                                                                       \
  X(EXACT, 3)

#define ENUM_COMPARATOR_VAL(VAL, IX) PLCS_STR_CMP_##VAL = IX,
typedef enum plcs_string_comparator {
  PLCS_LIST_STRING_COMPARATORS(ENUM_COMPARATOR_VAL) PLCS_STR_CMP__COUNT
} plcs_string_comparator;
#undef ENUM_COMPARATOR_VAL

/**
 * @brief comparison operators for numeric (and unsigned numeric) evaluators
 */
#define PLCS_LIST_NUMERIC_COMPARATOR(X)                                                                                \
  X(EQ, 0)                                                                                                             \
  X(GT, 1)                                                                                                             \
  X(GTE, 2)                                                                                                            \
  X(LT, 3)                                                                                                             \
  X(LTE, 4)

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
  X(COMPONENT, 0)                                                                                                      \
  X(PROCESS_EXE, 1)                                                                                                    \
  X(PROCESS_EXE_FULL_PATH, 2)                                                                                          \
  X(PROCESS_BASEDIR_PATH, 3)                                                                                           \
  X(PROCESS_ARGV, 4)                                                                                                   \
  X(PROCESS_CWD, 5)                                                                                                    \
  X(RUNTIME_LANGUAGE, 6)                                                                                               \
  X(RUNTIME_ENTRY_POINT_FILE, 7)                                                                                       \
  X(RUNTIME_ENTRY_POINT_JAR, 8)                                                                                        \
  X(RUNTIME_ENTRY_POINT_CLASS, 9)                                                                                      \
  X(RUNTIME_ENTRY_POINT_PACKAGE, 10)                                                                                   \
  X(RUNTIME_ENTRY_POINT_MODULE, 11)                                                                                    \
  X(RUNTIME_ENTRY_POINT_SOURCE, 12)                                                                                    \
  X(RUNTIME_DOPTION, 13)                                                                                               \
  X(RUNTIME_VERSION, 14)                                                                                               \
  X(LIBC_FLAVOR, 15)                                                                                                   \
  X(LIBC_VERSION, 16)                                                                                                  \
  X(MACHINE_ARCHITECTURE, 17)                                                                                          \
  X(HOST_NAME, 18)                                                                                                     \
  X(HOST_IP, 19)                                                                                                       \
  X(OS, 20)                                                                                                            \
  X(OS_DISTRO, 21)                                                                                                     \
  X(OS_DISTRO_VERSION, 22)                                                                                             \
  X(OS_DISTRO_CODENAME, 23)                                                                                            \
  X(OS_KERNEL_VERSION, 24)                                                                                             \
  X(OS_KERNEL_NAME, 25)                                                                                                \
  X(OS_USER, 26)                                                                                                       \
  X(OS_USER_GROUP, 27)                                                                                                 \
  X(CONTAINER_IMAGE, 28)                                                                                               \
  X(CONTAINER_ID, 29)                                                                                                  \
  X(ALWAYS_TRUE, 30)                                                                                                   \
  X(ALWAYS_FALSE, 31)                                                                                                  \
  X(ALWAYS_ABSTAIN, 32)

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
  X(JAVA_HEAP, 0)                                                                                                      \
  X(RUNTIME_VERSION_MAJOR, 1)                                                                                          \
  X(RUNTIME_VERSION_MINOR, 2)                                                                                          \
  X(RUNTIME_VERSION_PATCH, 3)                                                                                          \
  X(OS_DISTRO_VERSION_MAJOR, 4)                                                                                        \
  X(OS_DISTRO_VERSION_MINOR, 5)                                                                                        \
  X(OS_DISTRO_VERSION_PATCH, 6)                                                                                        \
  X(OS_KERNEL_VERSION_MAJOR, 7)                                                                                        \
  X(OS_KERNEL_VERSION_MINOR, 8)                                                                                        \
  X(OS_KERNEL_VERSION_PATCH, 9)                                                                                        \
  X(LIBC_VERSION_MAJOR, 10)                                                                                            \
  X(LIBC_VERSION_MINOR, 11)                                                                                            \
  X(LIBC_VERSION_PATCH, 12)

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
