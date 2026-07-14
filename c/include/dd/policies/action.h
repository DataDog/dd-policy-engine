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
 * @brief represents an action function signature
 *
 */
typedef plcs_errors (*plcs_action_function_ptr)(
    plcs_evaluation_result res,
    const char *values[],
    size_t value_len,
    const char *description,
    int action_id
);

const char *plcs_actions_to_string(enum plcs_actions action);
