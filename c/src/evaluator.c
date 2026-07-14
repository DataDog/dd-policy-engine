/*
 * Unless explicitly stated otherwise all files in this repository are licensed
 * under the Apache 2.0 License. This product includes software developed at
 * Datadog (https://www.datadoghq.com/).
 *
 * Copyright 2025-Present Datadog, Inc.
 */
#include <policy_reader.h>

#include <dd/policies/error_codes.h>
#include <dd/policies/evaluator_default.h>
#include <dd/policies/policies.h>

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "eval_ctx.h"
#include "policy.h"
#include "wire/action.h"
#include "wire/boolean_operation.h"
#include "wire/dd_types.h"
#include "wire/evaluation_result.h"
#define PLCS_MAX_EVAL_DEPTH 64

plcs_evaluation_result evaluate_rules(dd_ns(NodeTypeWrapper_table_t) node, int depth);

plcs_evaluation_result evaluate_string(dd_ns(StrEvaluator_table_t) eval_str, const char *description) {
  if (!eval_str) {
    return PLCS_EVAL_RESULT_ABSTAIN;
  }

  plcs_string_evaluators eval_id = dd_ns(StrEvaluator_id)(eval_str);

  const char *param = plcs_eval_ctx_get_string_param(eval_id);

  plcs_string_evaluator_function_ptr eval = plcs_eval_ctx_get_string_evaluator(eval_id);
  if (!eval) {
    eval = plcs_default_string_evaluator;
  }

  // parameter could potentially be NULL, so we check if there was an explicit error
  if (plcs_eval_ctx_peek_last_error() == PLCS_ESUCCESS) {
    return eval(dd_ns(StrEvaluator_value)(eval_str), dd_ns(StrEvaluator_cmp)(eval_str), param, description, eval_id);
  }

  return PLCS_EVAL_RESULT_ABSTAIN;
}

plcs_evaluation_result evaluate_numeric(dd_ns(NumEvaluator_table_t) eval_num, const char *description) {
  if (!eval_num) {
    return PLCS_EVAL_RESULT_ABSTAIN;
  }

  plcs_numeric_evaluators eval_id = dd_ns(NumEvaluator_id)(eval_num);

  const long param = plcs_eval_ctx_get_numeric_param(eval_id);

  if (param == PLCS_NUM_NOT_SET) {
    return PLCS_EVAL_RESULT_ABSTAIN;
  }

  plcs_numeric_evaluator_function_ptr eval = plcs_eval_ctx_get_numeric_evaluator(eval_id);
  if (!eval) {
    eval = plcs_default_numeric_evaluator;
  }

  // parameter could potentially be NULL, so we check if there was an explicit error
  if (plcs_eval_ctx_peek_last_error() == PLCS_ESUCCESS) {
    return eval(dd_ns(NumEvaluator_value)(eval_num), dd_ns(NumEvaluator_cmp)(eval_num), param, description, eval_id);
  }

  return PLCS_EVAL_RESULT_ABSTAIN;
}

plcs_evaluation_result evaluate_unumeric(dd_ns(UNumEvaluator_table_t) eval_unum, const char *description) {
  if (!eval_unum) {
    return PLCS_EVAL_RESULT_ABSTAIN;
  }

  plcs_numeric_evaluators eval_id = dd_ns(UNumEvaluator_id)(eval_unum);

  const unsigned long param = plcs_eval_ctx_get_unumeric_param(eval_id);

  if (param == PLCS_UNUM_NOT_SET) {
    return PLCS_EVAL_RESULT_ABSTAIN;
  }

  plcs_unumeric_evaluator_function_ptr eval = plcs_eval_ctx_get_unumeric_evaluator(eval_id);
  if (!eval) {
    eval = plcs_default_unumeric_evaluator;
  }

  // parameter could potentially be NULL, so we check if there was an explicit error
  if (plcs_eval_ctx_peek_last_error() == PLCS_ESUCCESS) {
    return eval(
        dd_ns(UNumEvaluator_value)(eval_unum), dd_ns(UNumEvaluator_cmp)(eval_unum), param, description, eval_id
    );
  }

  return PLCS_EVAL_RESULT_ABSTAIN;
}

plcs_evaluation_result node_evaluator(dd_ns(EvaluatorNode_table_t) node) {
  plcs_evaluation_result result = PLCS_EVAL_RESULT_ABSTAIN;
  if (!node) {
    return result;  // log error?
  }

  dd_ns(EvaluatorType_union_t) evaluator = dd_ns(EvaluatorNode_eval_union)(node);

  switch (evaluator.type) {
    case dd_ns(EvaluatorType_StrEvaluator):
      return evaluate_string(evaluator.value, dd_ns(EvaluatorNode_description)(node));

    case dd_ns(EvaluatorType_NumEvaluator):
      return evaluate_numeric(evaluator.value, dd_ns(EvaluatorNode_description)(node));

    case dd_ns(EvaluatorType_UNumEvaluator):
      return evaluate_unumeric(evaluator.value, dd_ns(EvaluatorNode_description)(node));
  }

  plcs_eval_ctx_set_error(PLCS_EUNKNOWN_EVAL_IX);
  return result;
}

plcs_evaluation_result DoAnd(plcs_evaluation_result a, plcs_evaluation_result b) {
  // 0 & {0, 1, x} -> 0
  if (a == PLCS_EVAL_RESULT_FALSE || b == PLCS_EVAL_RESULT_FALSE) {
    return PLCS_EVAL_RESULT_FALSE;
  }

  // {1} & {1} -> a & b
  if (a != PLCS_EVAL_RESULT_ABSTAIN && b != PLCS_EVAL_RESULT_ABSTAIN) {
    return PLCS_EVAL_RESULT_TRUE;
  }

  // {1, x} & {x} -> x
  return PLCS_EVAL_RESULT_ABSTAIN;
}

plcs_evaluation_result DoOr(plcs_evaluation_result a, plcs_evaluation_result b) {
  // {1} | {0, 1, x} -> 1
  if (a == PLCS_EVAL_RESULT_TRUE || b == PLCS_EVAL_RESULT_TRUE) {
    return PLCS_EVAL_RESULT_TRUE;
  }

  // {0, 1} | {0, 1} -> a | b
  if (a != PLCS_EVAL_RESULT_ABSTAIN && b != PLCS_EVAL_RESULT_ABSTAIN) {
    return PLCS_EVAL_RESULT_FALSE;
  }

  // {0, x} | {x} -> x
  return PLCS_EVAL_RESULT_ABSTAIN;
}

plcs_evaluation_result DoNot(plcs_evaluation_result res) {
  return res == PLCS_EVAL_RESULT_ABSTAIN ? PLCS_EVAL_RESULT_ABSTAIN : !res;
}

plcs_evaluation_result DoOper(dd_ns(BoolOperation_enum_t) oper, plcs_evaluation_result a, plcs_evaluation_result b) {
  switch (oper) {
    case dd_ns(BoolOperation_BOOL_AND):
      return DoAnd(a, b);

    case dd_ns(BoolOperation_BOOL_OR):
      return DoOr(a, b);

    case dd_ns(BoolOperation_BOOL_NOT):
      return DoNot(a);

    case dd_ns(BoolOperation_BOOL_UNKNOWN):
    case dd_ns(BoolOperation_BOOL_COUNT):
    default:
      return PLCS_EVAL_RESULT_ABSTAIN;
  }
}

plcs_evaluation_result composite_evaluator(dd_ns(CompositeNode_table_t) node, int depth) {
  if (!node) {
    return PLCS_EVAL_RESULT_ABSTAIN;
  }

  dd_ns(NodeTypeWrapper_vec_t) children = dd_ns(CompositeNode_children)(node);
  size_t children_len = children ? dd_ns(NodeTypeWrapper_vec_len)(children) : 0;
  dd_ns(BoolOperation_enum_t) oper = dd_ns(CompositeNode_op)(node);

  plcs_evaluation_result res;

  switch (oper) {
    case dd_ns(BoolOperation_BOOL_UNKNOWN):
      return PLCS_EVAL_RESULT_ABSTAIN;

    case dd_ns(BoolOperation_BOOL_OR):
      res = PLCS_EVAL_RESULT_FALSE;
      break;

    case dd_ns(BoolOperation_BOOL_AND):
      res = PLCS_EVAL_RESULT_TRUE;
      break;

    case dd_ns(BoolOperation_BOOL_NOT):
      // CAN ONLY HAVE ONE CHILD!
      // otherwise this is a non valid boolean operation
      if (children_len != 1) {
        // log error
        return PLCS_EVAL_RESULT_ABSTAIN;
      }
      return DoNot(evaluate_rules(dd_ns(NodeTypeWrapper_vec_at)(children, 0), depth + 1));

    case dd_ns(BoolOperation_BOOL_COUNT):
    default:
      return PLCS_EVAL_RESULT_ABSTAIN;
  }

  // keep iterating recursively over the tree
  for (size_t ix = 0; ix < children_len; ++ix) {
    res = DoOper(oper, res, evaluate_rules(dd_ns(NodeTypeWrapper_vec_at)(children, ix), depth + 1));

    // short circuit
    if (oper == dd_ns(BoolOperation_BOOL_OR) && res == PLCS_EVAL_RESULT_TRUE) {
      return res;
    }

    // short circuit
    if (oper == dd_ns(BoolOperation_BOOL_AND) && res == PLCS_EVAL_RESULT_FALSE) {
      return res;
    }
  }

  return res;
}

plcs_evaluation_result evaluate_rules(dd_ns(NodeTypeWrapper_table_t) node, int depth) {
  if (depth > PLCS_MAX_EVAL_DEPTH) {
    return PLCS_EVAL_RESULT_ABSTAIN;
  }

  switch (dd_ns(NodeTypeWrapper_node_type)(node)) {
    case dd_ns(NodeType_EvaluatorNode):
      return node_evaluator(dd_ns(NodeTypeWrapper_node)(node));

    case dd_ns(NodeType_CompositeNode):
      return composite_evaluator(dd_ns(NodeTypeWrapper_node)(node), depth);

    default:
      return PLCS_EVAL_RESULT_ABSTAIN;
  }
}

static void free_action_values(char *values[], size_t values_len) {
  for (size_t ix = 0; ix < values_len; ++ix) {
    free(values[ix]);
  }
}

static plcs_errors copy_action_values(flatbuffers_string_vec_t source_values, char *values[], size_t values_len) {
  for (size_t ix = 0; ix < values_len; ++ix) {
    const char *value = flatbuffers_string_vec_at(source_values, ix);
    if (!value) {
      values[ix] = NULL;
      continue;
    }

    size_t value_len = strlen(value) + 1;
    values[ix] = malloc(value_len);
    if (!values[ix]) {
      return PLCS_EACTIONS_EVAL;
    }
    memcpy(values[ix], value, value_len);
  }

  return PLCS_ESUCCESS;
}

static inline plcs_errors perform_actions(plcs_evaluation_result eval_res, dd_ns(Action_vec_t) actions_vec) {
  plcs_errors first_error = PLCS_ESUCCESS;

  size_t actions_len = dd_ns(Action_vec_len)(actions_vec);
  for (size_t ix = 0; ix < actions_len; ++ix) {
    dd_ns(Action_table_t) action = dd_ns(Action_vec_at)(actions_vec, ix);
    int action_id = dd_ns(Action_action)(action);
    plcs_action_function_ptr action_function = plcs_eval_ctx_get_action(action_id);
    if (!action_function) {
      continue;
    }

    flatbuffers_string_vec_t source_values = dd_ns(Action_values)(action);
    size_t values_len = flatbuffers_vec_len(source_values);
    if (values_len >= (size_t)PLCS_ACTION_VALUES_MAX) {
      if (first_error == PLCS_ESUCCESS) {
        first_error = PLCS_EIX_OVERFLOW;
      }
      continue;
    }

    char *values[PLCS_ACTION_VALUES_MAX] = {0};
    plcs_errors action_result = copy_action_values(source_values, values, values_len);
    if (action_result == PLCS_ESUCCESS) {
      // Keep the allocated pointers intact for cleanup if the callback replaces
      // entries in its pointer array.
      char *callback_values[PLCS_ACTION_VALUES_MAX] = {0};
      memcpy(callback_values, values, values_len * sizeof(*values));
      action_result =
          action_function(eval_res, callback_values, values_len, dd_ns(Action_description)(action), action_id);
    }
    free_action_values(values, values_len);

    plcs_eval_ctx_set_action_error(action_id, action_result);
    if (first_error == PLCS_ESUCCESS && action_result != PLCS_ESUCCESS) {
      first_error = action_result;
    }
  }

  return first_error;
}

static plcs_errors validate_rules(dd_ns(NodeTypeWrapper_table_t) node, int depth) {
  if (!node || depth > PLCS_MAX_EVAL_DEPTH) {
    return PLCS_ESUCCESS;
  }

  if (dd_ns(NodeTypeWrapper_node_type)(node) != dd_ns(NodeType_CompositeNode)) {
    return PLCS_ESUCCESS;
  }

  dd_ns(CompositeNode_table_t) composite = dd_ns(NodeTypeWrapper_node)(node);
  if (!composite) {
    return PLCS_ESUCCESS;
  }

  dd_ns(BoolOperation_enum_t) oper = dd_ns(CompositeNode_op)(composite);
  dd_ns(NodeTypeWrapper_vec_t) children = dd_ns(CompositeNode_children)(composite);
  size_t children_len = children ? dd_ns(NodeTypeWrapper_vec_len)(children) : 0;

  switch (oper) {
    case dd_ns(BoolOperation_BOOL_UNKNOWN):
    case dd_ns(BoolOperation_BOOL_AND):
    case dd_ns(BoolOperation_BOOL_OR):
    case dd_ns(BoolOperation_BOOL_NOT):
      break;

    case dd_ns(BoolOperation_BOOL_COUNT):
    default:
      return PLCS_EUNKNOWN_CMP;
  }

  for (size_t ix = 0; ix < children_len; ++ix) {
    plcs_errors validation_result = validate_rules(dd_ns(NodeTypeWrapper_vec_at)(children, ix), depth + 1);
    if (validation_result != PLCS_ESUCCESS) {
      return validation_result;
    }
  }

  return PLCS_ESUCCESS;
}

static plcs_errors validate_actions(dd_ns(Action_vec_t) actions) {
  size_t actions_len = actions ? dd_ns(Action_vec_len)(actions) : 0;
  for (size_t ix = 0; ix < actions_len; ++ix) {
    int action_id = dd_ns(Action_action)(dd_ns(Action_vec_at)(actions, ix));
    if (action_id < 0 || action_id >= dd_ns(ActionId_ACTIONS_COUNT)) {
      return PLCS_EIX_OVERFLOW;
    }
  }

  return PLCS_ESUCCESS;
}

plcs_errors evaluate_policy(dd_ns(Policy_table_t) policy) {
  // extract actions
  dd_ns(Action_vec_t) actions = dd_ns(Policy_actions)(policy);

  // extract rules
  dd_ns(NodeTypeWrapper_table_t) rules = dd_ns(Policy_rules)(policy);

  plcs_errors validation_result = validate_actions(actions);
  if (validation_result == PLCS_ESUCCESS) {
    validation_result = validate_rules(rules, 0);
  }
  if (validation_result != PLCS_ESUCCESS) {
    return validation_result;
  }

  // // evaluate rules if they exist, otherwise return EVAL_RESULT_ABSTAIN
  plcs_evaluation_result eval_res = rules ? evaluate_rules(rules, 0) : PLCS_EVAL_RESULT_ABSTAIN;

  // perform actions given evaluation result
  return perform_actions(eval_res, actions);
}

plcs_errors plcs_evaluate_buffer(const uint8_t *buffer, size_t size) {
  dd_ns(Policy_vec_t) policies = plcs_get_policies(buffer, size);
  if (!policies) {
    // not necessarily an error, could be empty policies
    return PLCS_ENO_DATA;
  }

  size_t policies_count = dd_ns(Policy_vec_len)(policies);
  plcs_errors first_error = PLCS_ESUCCESS;
  for (size_t ix = 0; ix < policies_count; ++ix) {
    dd_ns(Policy_table_t) policy = dd_ns(Policy_vec_at)(policies, ix);
    if (!policy) {
      // not necessarily an error, could be empty policy
      continue;
    }
    plcs_errors policy_result = evaluate_policy(policy);
    if (first_error == PLCS_ESUCCESS && policy_result != PLCS_ESUCCESS) {
      first_error = policy_result;
    }
  }

  return first_error;
}

const char *plcs_string_evaluators_to_string(enum plcs_string_evaluators v) {
  return dd_ns(StringEvaluators_name)(dd_streval_to_wire(v));
}

const char *plcs_numeric_evaluators_to_string(enum plcs_numeric_evaluators v) {
  return dd_ns(NumericEvaluators_name)(dd_numeval_to_wire(v));
}

const char *plcs_string_comparator_to_string(enum plcs_string_comparator v) {
  return dd_ns(CmpTypeSTR_name)(dd_strcmp_to_wire(v));
}

const char *plcs_numeric_comparator_to_string(enum plcs_numeric_comparator v) {
  return dd_ns(CmpTypeNUM_name)(dd_numcmp_to_wire(v));
}

const char *plcs_evaluation_result_to_string(enum plcs_evaluation_result res) {
  return dd_ns(EvaluationResult_name)(dd_evalresult_to_wire(res));
}

const char *plcs_actions_to_string(enum plcs_actions action) {
  return dd_ns(ActionId_name)(dd_action_to_wire(action));
}
