/*
 * Unless explicitly stated otherwise all files in this repository are licensed
 * under the Apache 2.0 License. This product includes software developed at
 * Datadog (https://www.datadoghq.com/).
 *
 * Copyright 2025-Present Datadog, Inc.
 */
#pragma once

#include <dd/policies/error_codes.h>
#include <dd/policies/evaluation_result.h>
#include <dd/policies/evaluator_types.h>
#include <stdint.h>
#include <stdlib.h>

/**
 * @brief A mapping between flatbuffers defined enums and a local
 * representation.
 *
 */
#define PLCS_LIST_ACTIONS(X)                                                                                           \
  X(ACTION_UNKNOWN, 0)                                                                                                 \
  X(INJECT_DENY, 1)                                                                                                    \
  X(INJECT_ALLOW, 2)                                                                                                   \
  X(ENABLE_SDK, 3)                                                                                                     \
  X(ENABLE_PROFILER, 4)                                                                                                \
  X(SET_ENVAR, 5)                                                                                                      \
  X(REEXEC, 6)

#define ENUM_VAL(ID, IX) PLCS_ACTION_##ID = IX,
typedef enum plcs_actions { PLCS_LIST_ACTIONS(ENUM_VAL) PLCS_ACTIONS__COUNT } plcs_actions;
#undef ENUM_VAL

/**
 * @brief A 128-bit UUID, represented as two 64-bit halves.
 */
typedef struct {
  uint64_t hi;
  uint64_t lo;
} plcs_uuid;

/**
 * @brief The maximum number of matched rules reported to an action.
 *
 * A policy tree may contain more leaves than this; any matches beyond the limit
 * are dropped rather than reported, so an action must treat the reported list as
 * potentially truncated.
 */
#define PLCS_MATCHED_RULES_MAX 64

/**
 * @brief Selects the active member of the value unions in plcs_matched_rule.
 */
typedef enum plcs_rule_value_kind {
  /** The leaf is a string evaluator: read `str`. */
  PLCS_RULE_VALUE_STR = 0,
  /** The leaf is a signed numeric evaluator: read `num`. */
  PLCS_RULE_VALUE_NUM,
  /** The leaf is an unsigned numeric evaluator: read `unum`. */
  PLCS_RULE_VALUE_UNUM,
} plcs_rule_value_kind;

/**
 * @brief A single leaf condition that evaluated TRUE during a policy evaluation.
 *
 * @note Leaf nodes carry no useful description of their own in practice (policy
 * generation leaves them unset), so the human-readable text for a decision comes
 * from the action and policy descriptions instead.
 *
 * @note All pointers borrow memory owned by the policy buffer and the evaluation
 * context. They are only valid for the duration of the action call - copy anything
 * that must outlive it.
 */
typedef struct plcs_matched_rule {
  /** Which member of `policy_value` and `process_value` is set. */
  plcs_rule_value_kind kind;
  /** A plcs_string_evaluators value when `kind` is PLCS_RULE_VALUE_STR, a
   * plcs_numeric_evaluators value otherwise. */
  int evaluator_id;
  /** A plcs_string_comparator value when `kind` is PLCS_RULE_VALUE_STR, a
   * plcs_numeric_comparator value otherwise. */
  int comparator;
  /** The value the policy compared against. */
  union {
    const char *str;
    long num;
    unsigned long unum;
  } policy_value;
  /** The value read from the evaluation context (the local/process value). */
  union {
    const char *str;
    long num;
    unsigned long unum;
  } process_value;
} plcs_matched_rule;

/**
 * @brief represents an action function signature
 *
 * @param res                 The evaluation result determining action behavior.
 * @param values              Array of string values passed to the action.
 * @param value_len           Length of the `values` array.
 * @param description         The action's own description, as authored in the policy.
 * @param action_id           Integer ID of the action (a plcs_actions value).
 * @param policy_id           The id of the policy that produced this action.
 * @param policy_version      The version of the policy that produced this action.
 * @param policy_description  The description of the policy that produced this action.
 * @param rule_description    The description of the policy's root rule node. In practice
 *                            this restates the policy description, but it is reported
 *                            separately in case a producer sets it more specifically.
 * @param matched_rules       The leaf conditions that evaluated TRUE, in evaluation
 *                            order. Borrowed, and only valid for this call.
 * @param matched_rules_len   Length of the `matched_rules` array, capped at
 *                            PLCS_MATCHED_RULES_MAX.
 *
 */
typedef plcs_errors (*plcs_action_function_ptr)(
    plcs_evaluation_result res,
    char *values[],
    size_t value_len,
    const char *description,
    int action_id,
    plcs_uuid policy_id,
    int64_t policy_version,
    const char *policy_description,
    const char *rule_description,
    const plcs_matched_rule *matched_rules,
    size_t matched_rules_len
);

const char *plcs_actions_to_string(enum plcs_actions action);
