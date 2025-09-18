#pragma once

/**
 * @brief defines a new type for boolean results, similar to optional where
 * EVAL_RESULT_ABSTAIN defines a 'dont-care' (optional) state
 *
 */
#define PLCS_LIST_RESULT(X)                                                                                            \
  X(TRUE, 0)                                                                                                           \
  X(FALSE, 1)                                                                                                          \
  X(ABSTAIN, 2)

#define ENUM_VAL(ID, IX) PLCS_EVAL_RESULT_##ID = IX,
typedef enum plcs_evaluation_result { PLCS_LIST_RESULT(ENUM_VAL) PLCS_EVAL_RESULT__COUNT } plcs_evaluation_result;
#undef ENUM_VAL

const char *plcs_evaluation_result_to_string(enum plcs_evaluation_result res);
