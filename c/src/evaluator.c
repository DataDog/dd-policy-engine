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
#include "eval_ctx.h"
#include "policy.h"
#include "wire/action.h"
#include "wire/boolean_operation.h"
#include "wire/dd_types.h"
#include "wire/evaluation_result.h"
#define PLCS_MAX_EVAL_DEPTH 64

static char synth_buf[512];

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
      break;

    case dd_ns(BoolOperation_BOOL_OR):
      return DoOr(a, b);
      break;

    case dd_ns(BoolOperation_BOOL_NOT):
      return DoNot(a);
      break;

    default:
      // error unknown result
      return PLCS_EVAL_RESULT_ABSTAIN;
      break;
  }
}

static const char *cmp_op_label(dd_ns(CmpTypeSTR_enum_t) cmp) {
  switch (cmp) {
    case dd_ns(CmpTypeSTR_CMP_EXACT):    return "matches";
    case dd_ns(CmpTypeSTR_CMP_PREFIX):   return "starts with";
    case dd_ns(CmpTypeSTR_CMP_SUFFIX):   return "ends with";
    case dd_ns(CmpTypeSTR_CMP_CONTAINS): return "contains";
    case dd_ns(CmpTypeSTR_CMP_WILDCARD): return "matches (wildcard)";
    default:                             return "?";
  }
}

static const char *str_eval_field_label(dd_ns(StringEvaluators_enum_t) id) {
  switch (id) {
    case dd_ns(StringEvaluators_COMPONENT):                 return "component";
    case dd_ns(StringEvaluators_PROCESS_EXE):               return "executable";
    case dd_ns(StringEvaluators_PROCESS_EXE_FULL_PATH):     return "executable full path";
    case dd_ns(StringEvaluators_PROCESS_BASEDIR_PATH):      return "process base directory";
    case dd_ns(StringEvaluators_PROCESS_ARGV):              return "command-line argument";
    case dd_ns(StringEvaluators_PROCESS_CWD):               return "process working directory";
    case dd_ns(StringEvaluators_RUNTIME_LANGUAGE):          return "language";
    case dd_ns(StringEvaluators_RUNTIME_ENTRY_POINT_FILE):  return "entry point file";
    case dd_ns(StringEvaluators_RUNTIME_ENTRY_POINT_JAR):   return "entry point JAR";
    case dd_ns(StringEvaluators_RUNTIME_ENTRY_POINT_CLASS): return "entry point class";
    case dd_ns(StringEvaluators_RUNTIME_ENTRY_POINT_PACKAGE): return "entry point package";
    case dd_ns(StringEvaluators_RUNTIME_ENTRY_POINT_MODULE): return "entry point module";
    case dd_ns(StringEvaluators_RUNTIME_ENTRY_POINT_SOURCE): return "entry point source file";
    case dd_ns(StringEvaluators_RUNTIME_DOPTION):           return "runtime -D option";
    case dd_ns(StringEvaluators_RUNTIME_VERSION):           return "runtime version";
    case dd_ns(StringEvaluators_LIBC_FLAVOR):               return "libc flavor";
    case dd_ns(StringEvaluators_LIBC_VERSION):              return "libc version";
    case dd_ns(StringEvaluators_MACHINE_ARCHITECTURE):      return "machine architecture";
    case dd_ns(StringEvaluators_HOST_NAME):                 return "host name";
    case dd_ns(StringEvaluators_HOST_IP):                   return "host IP";
    case dd_ns(StringEvaluators_OS):                        return "operating system";
    case dd_ns(StringEvaluators_OS_DISTRO):                 return "OS distribution";
    case dd_ns(StringEvaluators_OS_DISTRO_VERSION):         return "OS distribution version";
    case dd_ns(StringEvaluators_OS_DISTRO_CODENAME):        return "OS distribution codename";
    case dd_ns(StringEvaluators_OS_KERNEL_VERSION):         return "OS kernel version";
    case dd_ns(StringEvaluators_OS_KERNEL_NAME):            return "OS kernel name";
    case dd_ns(StringEvaluators_OS_USER):                   return "OS user";
    case dd_ns(StringEvaluators_OS_USER_GROUP):             return "OS user group";
    case dd_ns(StringEvaluators_CONTAINER_IMAGE):           return "container image";
    case dd_ns(StringEvaluators_CONTAINER_ID):              return "container ID";
    case dd_ns(StringEvaluators_IIS_APPLICATION_POOL):      return "IIS application pool";
    case dd_ns(StringEvaluators_PROCESS_ARGV_0):            return "command-line argument 0";
    case dd_ns(StringEvaluators_PROCESS_ARGV_1):            return "command-line argument 1";
    case dd_ns(StringEvaluators_PROCESS_ARGV_2):            return "command-line argument 2";
    case dd_ns(StringEvaluators_PROCESS_ARGV_3):            return "command-line argument 3";
    case dd_ns(StringEvaluators_PROCESS_ARGV_4):            return "command-line argument 4";
    case dd_ns(StringEvaluators_PROCESS_ARGV_5):            return "command-line argument 5";
    case dd_ns(StringEvaluators_PROCESS_ARGV_N):            return "last command-line argument";
    case dd_ns(StringEvaluators_PROCESS_ARGV_N_2):          return "second to last command-line argument";
    case dd_ns(StringEvaluators_PROCESS_ARGV_N_3):          return "third to last command-line argument";
    case dd_ns(StringEvaluators_PROCESS_ARGV_N_4):          return "fourth to last command-line argument";
    case dd_ns(StringEvaluators_PROCESS_ARGV_N_5):          return "fifth to last command-line argument";
    case dd_ns(StringEvaluators_PROCESS_ARGV_N_6):          return "sixth to last command-line argument";
    case dd_ns(StringEvaluators_PROCESS_ENVAR):             return "process environment variable";
    default:                                                return dd_ns(StringEvaluators_name)(id);
  }
}

static inline void capture_matched_description(dd_ns(NodeTypeWrapper_table_t) node) {
  if (!node) {
    return;
  }
  // A more specific description from a deeper node takes priority.
  const char *existing = plcs_eval_ctx_get_matched_description();
  if (existing && existing[0] != '\0') {
    return;
  }
  const char *description = NULL;
  switch (dd_ns(NodeTypeWrapper_node_type)(node)) {
    case dd_ns(NodeType_CompositeNode): {
      dd_ns(CompositeNode_table_t) comp = dd_ns(NodeTypeWrapper_node)(node);
      description = dd_ns(CompositeNode_description)(comp);
      if ((!description || description[0] == '\0') &&
          dd_ns(CompositeNode_op)(comp) == dd_ns(BoolOperation_BOOL_NOT)) {
        dd_ns(NodeTypeWrapper_vec_t) children = dd_ns(CompositeNode_children)(comp);
        if (children && dd_ns(NodeTypeWrapper_vec_len)(children) == 1) {
          capture_matched_description(dd_ns(NodeTypeWrapper_vec_at)(children, 0));
        }
        return;
      }
      break;
    }
    case dd_ns(NodeType_EvaluatorNode): {
      dd_ns(EvaluatorNode_table_t) eval_node = dd_ns(NodeTypeWrapper_node)(node);
      description = dd_ns(EvaluatorNode_description)(eval_node);
      if (!description || description[0] == '\0') {
        dd_ns(EvaluatorType_union_t) eval = dd_ns(EvaluatorNode_eval_union)(eval_node);
        if (eval.type == dd_ns(EvaluatorType_StrEvaluator)) {
          dd_ns(StrEvaluator_table_t) str_eval = eval.value;
          const char *field = str_eval_field_label(dd_ns(StrEvaluator_id)(str_eval));
          const char *op    = cmp_op_label(dd_ns(StrEvaluator_cmp)(str_eval));
          const char *value = dd_ns(StrEvaluator_value)(str_eval);
          snprintf(
              synth_buf, sizeof(synth_buf),
              "%s %s '%s'", field ? field : "?", op, value ? value : ""
          );
          plcs_eval_ctx_set_matched_description(synth_buf);
          return;
        }
      }
      break;
    }
  }
  if (description && description[0] != '\0') {
    plcs_eval_ctx_set_matched_description(description);
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
      res = PLCS_EVAL_RESULT_ABSTAIN;
      break;

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
      break;
  }

  // keep iterating recursively over the tree
  dd_ns(NodeTypeWrapper_table_t) last_true_child = NULL;
  for (size_t ix = 0; ix < children_len; ++ix) {
    dd_ns(NodeTypeWrapper_table_t) child = dd_ns(NodeTypeWrapper_vec_at)(children, ix);
    plcs_evaluation_result child_res = evaluate_rules(child, depth + 1);
    res = DoOper(oper, res, child_res);

    // short circuit OR: capture the child that caused TRUE
    if (oper == dd_ns(BoolOperation_BOOL_OR) && res == PLCS_EVAL_RESULT_TRUE) {
      capture_matched_description(child);
      return res;
    }

    // track last TRUE child for AND description capture
    if (oper == dd_ns(BoolOperation_BOOL_AND) && child_res == PLCS_EVAL_RESULT_TRUE) {
      last_true_child = child;
    }

    // short circuit AND
    if (oper == dd_ns(BoolOperation_BOOL_AND) && res == PLCS_EVAL_RESULT_FALSE) {
      return res;
    }
  }

  // AND succeeded: capture description of last TRUE child (no-op if OR already set one)
  if (oper == dd_ns(BoolOperation_BOOL_AND) && last_true_child) {
    capture_matched_description(last_true_child);
  }

  return res;
}

plcs_evaluation_result evaluate_rules(dd_ns(NodeTypeWrapper_table_t) node, int depth) {
  if (depth > PLCS_MAX_EVAL_DEPTH) {
    return PLCS_EVAL_RESULT_ABSTAIN;
  }

  switch (dd_ns(NodeTypeWrapper_node_type)(node)) {
    case dd_ns(NodeType_EvaluatorNode):
      dd_ns(EvaluatorNode_table_t) evaluator_node = dd_ns(NodeTypeWrapper_node)(node);
      return node_evaluator(evaluator_node);
      break;

    case dd_ns(NodeType_CompositeNode):
      dd_ns(CompositeNode_table_t) composite_node = dd_ns(NodeTypeWrapper_node)(node);
      return composite_evaluator(composite_node, depth);
      break;

    default:
      // error, unknown node type!
      break;
  }

  // log error
  return PLCS_EVAL_RESULT_ABSTAIN;
}

static inline plcs_errors perform_actions(plcs_evaluation_result eval_res, dd_ns(Action_vec_t) actions_vec) {
  plcs_errors res = PLCS_ESUCCESS;

  // iterate
  size_t len = dd_ns(Action_vec_len)(actions_vec);
  for (size_t ix = 0; ix < len; ++ix) {
    dd_ns(Action_table_t) action = dd_ns(Action_vec_at)(actions_vec, ix);
    int action_id = dd_ns(Action_action)(action);
    if (action_id >= dd_ns(ActionId_ACTIONS_COUNT) || !plcs_eval_ctx_get_action(action_id)) {
      continue;
    }
    size_t values_len = flatbuffers_vec_len(dd_ns(Action_values(action)));
    // something went wrong, we need to bail
    if (values_len >= (size_t)PLCS_ACTION_VALUES_MAX) {
      res = PLCS_EIX_OVERFLOW;
      break;
    }
    char *values[PLCS_ACTION_VALUES_MAX];
    for (size_t v_ix = 0; v_ix < values_len; ++v_ix) {
      values[v_ix] = (char *)flatbuffers_string_vec_at(dd_ns(Action_values(action)), v_ix);
    }
    plcs_action_function_ptr action_function = plcs_eval_ctx_get_action(action_id);
    if (action_function) {
      res = action_function(eval_res, values, values_len, dd_ns(Action_description)(action), action_id);
    } else {
      res = PLCS_EACTIONS_EVAL;
    }
  }

  return res;
}

plcs_errors evaluate_policy(dd_ns(Policy_table_t) policy) {
  // extract actions
  dd_ns(Action_vec_t) actions = dd_ns(Policy_actions)(policy);

  // extract rules
  dd_ns(NodeTypeWrapper_table_t) rules = dd_ns(Policy_rules)(policy);

  // reset matched description
  plcs_eval_ctx_set_matched_description(NULL);

  // evaluate rules if they exist, otherwise return EVAL_RESULT_ABSTAIN
  plcs_evaluation_result eval_res = rules ? evaluate_rules(rules, 0) : PLCS_EVAL_RESULT_ABSTAIN;

  if (eval_res == PLCS_EVAL_RESULT_TRUE &&
      (plcs_eval_ctx_get_matched_description() == NULL || plcs_eval_ctx_get_matched_description()[0] == '\0')) {
    capture_matched_description(rules);
  }

  // Prepend [skip]/[allow] prefix from the first action when no prefix is present yet.
  const char *desc = plcs_eval_ctx_get_matched_description();
  if (desc && desc[0] != '\0' && desc[0] != '[' &&
      actions && dd_ns(Action_vec_len)(actions) > 0) {
    dd_ns(Action_table_t) first_action = dd_ns(Action_vec_at)(actions, 0);
    if (first_action) {
      const char *prefix =
          dd_ns(Action_action)(first_action) == dd_ns(ActionId_INJECT_DENY) ? "[skip] " : "[allow] ";
      char temp[512];
      snprintf(temp, sizeof(temp), "%s%s", prefix, desc);
      snprintf(synth_buf, sizeof(synth_buf), "%s", temp);
      plcs_eval_ctx_set_matched_description(synth_buf);
    }
  }

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
  plcs_errors total_errors = 0;
  for (size_t ix = 0; ix < policies_count; ++ix) {
    dd_ns(Policy_table_t) policy = dd_ns(Policy_vec_at)(policies, ix);
    if (!policy) {
      // not necessarily an error, could be empty policy
      continue;
    }
    plcs_errors res = evaluate_policy(policy);
    // success is 0, errors are > 0, if total_errors is > 0, it means there was
    // an error
    // TODO: track these errors using an errono style map in the eval_ctx
    total_errors += res;
  }

  return total_errors;
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
