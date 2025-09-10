#pragma once

#include <dd/policies/error_codes.h>
#include <dd/policies/evaluation_result.h>
#include <stdlib.h>

/**
 * @brief A mapping between flatbuffers defined enums and a local
 * representation.
 *
 */
typedef enum plcs_actions {
  ACTION_UNKNOWN = 0, // default and fallback value 
  INJECT_DENY = 1,
  INJECT_ALLOW = 2,
  ENABLE_SDK = 3,
  ENABLE_PROFILER = 4,
  SET_ENVAR = 5,
  REEXEC = 6,
  ACTIONS__COUNT
} plcs_actions;

/**
 * @brief represents an action function signature
 *
 */
typedef plcs_errors (*plcs_action_function_ptr)(
    plcs_evaluation_result res,
    char *values[],
    size_t value_len,
    const char *description,
    int action_id
);

const char *plcs_actions_to_string(enum plcs_actions action);
