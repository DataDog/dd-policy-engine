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
 * @brief represents an action function signature
 *
 * @param res                 The evaluation result determining action behavior.
 * @param values              Array of string values passed to the action.
 * @param value_len           Length of the `values` array.
 * @param description         The action's own description, as authored in the policy.
 * @param action_id           Integer ID of the action (a plcs_actions value).
 * @param policy_id           The id of the policy that produced this action.
 * @param policy_version      The version of the policy that produced this action.
 * @param policy_description  The description of the rule that triggered this action: the
 *                            descriptions of every condition of the matching AND node joined
 *                            with " AND ", or the first matching condition of an OR node.
 *                            Falls back to the policy's own description when no rule matched.
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
    const char *policy_description
);

const char *plcs_actions_to_string(enum plcs_actions action);
