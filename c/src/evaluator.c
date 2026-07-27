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
#include <string.h>
#include "eval_ctx.h"
#include "policy.h"
#include "wire/action.h"
#include "wire/boolean_operation.h"
#include "wire/dd_types.h"
#include "wire/evaluation_result.h"
#define PLCS_MAX_EVAL_DEPTH 64
#define PLCS_MATCHED_RULE_MAX 512

/* Description of the rule that made a policy match, assembled while the tree is
 * evaluated. Each condition is described from the evaluator itself, so it reports the context value it saw:
 * "process executable 'python3.11' is prefixed with 'python'".
 *
 * A node only describes itself when it evaluates to TRUE, and a composite joins
 * whatever its children described with its own operator. Since AND is only TRUE when
 * every child is TRUE, and OR short circuits on the first TRUE child, this yields
 * "every condition" for AND and "the first matching condition" for OR. */
typedef struct {
  char buf[PLCS_MATCHED_RULE_MAX];
  size_t len;
} plcs_matched_rule;

/* Scratch space for one formatted condition. Deliberately one byte longer than the rule
 * buffer, so that a condition too long to format here is also too long for the rule buffer:
 * matched_rule_append is then guaranteed to notice, which keeps truncation handling in one
 * place. */
#define PLCS_CONDITION_MAX (PLCS_MATCHED_RULE_MAX + 1)

// Appends as much of `text` as still fits. A description that does not fit is truncated with "..." at the end
static void matched_rule_append(plcs_matched_rule *rule, const char *text) {
  if (!rule || !text) {
    return;
  }

  static const char ellipsis[] = "...";

  size_t room = sizeof(rule->buf) - rule->len - 1;
  size_t len = strlen(text);

  if (len <= room) {
    memcpy(rule->buf + rule->len, text, len);
    rule->len += len;
    rule->buf[rule->len] = '\0';
    return;
  }

  memcpy(rule->buf + rule->len, text, room);
  rule->len += room;
  rule->buf[rule->len] = '\0';

  size_t marker_len = sizeof(ellipsis) - 1;
  size_t at = rule->len > marker_len ? rule->len - marker_len : 0;
  memcpy(rule->buf + at, ellipsis, rule->len - at);
}

// drops everything appended past `len`; used to undo branches that did not match
static void matched_rule_truncate(plcs_matched_rule *rule, size_t len) {
  if (!rule) {
    return;
  }

  rule->len = len;
  rule->buf[len] = '\0';
}

static size_t matched_rule_len(const plcs_matched_rule *rule) {
  return rule ? rule->len : 0;
}

// human readable labels for the evaluators and comparators
#define PLCS_STR_EVAL_LABEL(ID, IX, LABEL) [PLCS_STR_EVAL_##ID] = LABEL,
#define PLCS_NUM_EVAL_LABEL(ID, IX, LABEL) [PLCS_NUM_EVAL_##ID] = LABEL,
#define PLCS_STR_CMP_LABEL(ID, IX, LABEL) [PLCS_STR_CMP_##ID] = LABEL,
#define PLCS_NUM_CMP_LABEL(ID, IX, LABEL) [PLCS_NUM_CMP_##ID] = LABEL,

static const char *const string_evaluator_labels[PLCS_STR_EVAL__COUNT] = {
    PLCS_LIST_STRING_EVALUATORS(PLCS_STR_EVAL_LABEL)
};
static const char *const numeric_evaluator_labels[PLCS_NUM_EVAL__COUNT] = {
    PLCS_LIST_NUMERIC_EVALUATORS(PLCS_NUM_EVAL_LABEL)
};
static const char *const string_comparator_labels[PLCS_STR_CMP__COUNT] = {
    PLCS_LIST_STRING_COMPARATORS(PLCS_STR_CMP_LABEL)
};
static const char *const numeric_comparator_labels[PLCS_NUM_CMP__COUNT] = {
    PLCS_LIST_NUMERIC_COMPARATOR(PLCS_NUM_CMP_LABEL)
};

#undef PLCS_STR_EVAL_LABEL
#undef PLCS_NUM_EVAL_LABEL
#undef PLCS_STR_CMP_LABEL
#undef PLCS_NUM_CMP_LABEL

static const char *label_at(const char *const labels[], size_t count, unsigned id) {
  const char *label = id < count ? labels[id] : NULL;
  return label ? label : labels[0];
}

static const char *string_evaluator_label(plcs_string_evaluators eval_id) {
  return label_at(string_evaluator_labels, PLCS_STR_EVAL__COUNT, (unsigned)eval_id);
}

static const char *numeric_evaluator_label(plcs_numeric_evaluators eval_id) {
  return label_at(numeric_evaluator_labels, PLCS_NUM_EVAL__COUNT, (unsigned)eval_id);
}

static const char *string_comparator_label(plcs_string_comparator cmp) {
  return label_at(string_comparator_labels, PLCS_STR_CMP__COUNT, (unsigned)cmp);
}

static const char *numeric_comparator_label(plcs_numeric_comparator cmp) {
  return label_at(numeric_comparator_labels, PLCS_NUM_CMP__COUNT, (unsigned)cmp);
}

// appends to matched rule with description like "process executable 'python3.11' is prefixed with 'python'", or
// "runtime language is 'java'" */
static void matched_rule_describe_string(plcs_matched_rule *rule, dd_ns(StrEvaluator_table_t) eval_str) {
  plcs_string_evaluators eval_id = (plcs_string_evaluators)dd_ns(StrEvaluator_id)(eval_str);
  plcs_string_comparator cmp = (plcs_string_comparator)dd_ns(StrEvaluator_cmp)(eval_str);
  const char *param = plcs_eval_ctx_get_string_param(eval_id);
  const char *value = dd_ns(StrEvaluator_value)(eval_str);
  char text[PLCS_CONDITION_MAX];

  param = param ? param : "";
  value = value ? value : "";

  if (cmp == PLCS_STR_CMP_EXACT && strcmp(param, value) == 0) {
    snprintf(text, sizeof(text), "%s is '%s'", string_evaluator_label(eval_id), value);
  } else {
    snprintf(
        text, sizeof(text), "%s '%s' %s '%s'", string_evaluator_label(eval_id), param, string_comparator_label(cmp),
        value
    );
  }

  matched_rule_append(rule, text);
}

// appends to matched rule with description like "java heap 512 is greater than 256", or "runtime major version is 21"
// */
static void matched_rule_describe_numeric(plcs_matched_rule *rule, dd_ns(NumEvaluator_table_t) eval_num) {
  plcs_numeric_evaluators eval_id = (plcs_numeric_evaluators)dd_ns(NumEvaluator_id)(eval_num);
  plcs_numeric_comparator cmp = (plcs_numeric_comparator)dd_ns(NumEvaluator_cmp)(eval_num);
  long param = plcs_eval_ctx_get_numeric_param(eval_id);
  long value = (long)dd_ns(NumEvaluator_value)(eval_num);
  char text[PLCS_CONDITION_MAX];

  if (cmp == PLCS_NUM_CMP_EQ && param == value) {
    snprintf(text, sizeof(text), "%s is %ld", numeric_evaluator_label(eval_id), value);
  } else {
    snprintf(
        text, sizeof(text), "%s %ld %s %ld", numeric_evaluator_label(eval_id), param, numeric_comparator_label(cmp),
        value
    );
  }

  matched_rule_append(rule, text);
}

static void matched_rule_describe_unumeric(plcs_matched_rule *rule, dd_ns(UNumEvaluator_table_t) eval_unum) {
  plcs_numeric_evaluators eval_id = (plcs_numeric_evaluators)dd_ns(UNumEvaluator_id)(eval_unum);
  plcs_numeric_comparator cmp = (plcs_numeric_comparator)dd_ns(UNumEvaluator_cmp)(eval_unum);
  unsigned long param = plcs_eval_ctx_get_unumeric_param(eval_id);
  unsigned long value = (unsigned long)dd_ns(UNumEvaluator_value)(eval_unum);
  char text[PLCS_CONDITION_MAX];

  if (cmp == PLCS_NUM_CMP_EQ && param == value) {
    snprintf(text, sizeof(text), "%s is %lu", numeric_evaluator_label(eval_id), value);
  } else {
    snprintf(
        text, sizeof(text), "%s %lu %s %lu", numeric_evaluator_label(eval_id), param, numeric_comparator_label(cmp),
        value
    );
  }

  matched_rule_append(rule, text);
}

/* Describes the condition a leaf node checked, along with the context value it saw. */
static void matched_rule_describe_node(plcs_matched_rule *rule, dd_ns(EvaluatorNode_table_t) node) {
  dd_ns(EvaluatorType_union_t) evaluator = dd_ns(EvaluatorNode_eval_union)(node);

  switch (evaluator.type) {
    case dd_ns(EvaluatorType_StrEvaluator):
      matched_rule_describe_string(rule, evaluator.value);
      break;

    case dd_ns(EvaluatorType_NumEvaluator):
      matched_rule_describe_numeric(rule, evaluator.value);
      break;

    case dd_ns(EvaluatorType_UNumEvaluator):
      matched_rule_describe_unumeric(rule, evaluator.value);
      break;
  }
}

plcs_evaluation_result evaluate_rules(dd_ns(NodeTypeWrapper_table_t) node, int depth, plcs_matched_rule *rule);

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

plcs_evaluation_result composite_evaluator(dd_ns(CompositeNode_table_t) node, int depth, plcs_matched_rule *rule) {
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
      dd_ns(NodeTypeWrapper_table_t) negated = dd_ns(NodeTypeWrapper_vec_at)(children, 0);
      res = DoNot(evaluate_rules(negated, depth + 1, NULL));
      // the child is what did *not* match, so describe it negated. A negated composite has
      // no single condition to point at, so it is left to the policy description instead.
      if (res == PLCS_EVAL_RESULT_TRUE && dd_ns(NodeTypeWrapper_node_type)(negated) == dd_ns(NodeType_EvaluatorNode)) {
        matched_rule_append(rule, "NOT (");
        matched_rule_describe_node(rule, dd_ns(NodeTypeWrapper_node)(negated));
        matched_rule_append(rule, ")");
      }
      return res;
  }

  const char *separator = oper == dd_ns(BoolOperation_BOOL_AND) ? " AND " : " OR ";
  const size_t rule_start = matched_rule_len(rule);

  // keep iterating recursively over the tree
  for (size_t ix = 0; ix < children_len; ++ix) {
    const size_t before_separator = matched_rule_len(rule);
    if (before_separator > rule_start) {
      matched_rule_append(rule, separator);
    }
    const size_t before_child = matched_rule_len(rule);

    res = DoOper(oper, res, evaluate_rules(dd_ns(NodeTypeWrapper_vec_at)(children, ix), depth + 1, rule));

    // the child had nothing to describe, so drop the separator we added for it
    if (matched_rule_len(rule) == before_child) {
      matched_rule_truncate(rule, before_separator);
    }

    // short circuit
    if (oper == dd_ns(BoolOperation_BOOL_OR) && res == PLCS_EVAL_RESULT_TRUE) {
      break;
    }

    // short circuit
    if (oper == dd_ns(BoolOperation_BOOL_AND) && res == PLCS_EVAL_RESULT_FALSE) {
      break;
    }
  }

  // this branch did not match, so it is not part of the rule that triggered
  if (res != PLCS_EVAL_RESULT_TRUE) {
    matched_rule_truncate(rule, rule_start);
  }

  return res;
}

plcs_evaluation_result evaluate_rules(dd_ns(NodeTypeWrapper_table_t) node, int depth, plcs_matched_rule *rule) {
  if (depth > PLCS_MAX_EVAL_DEPTH) {
    return PLCS_EVAL_RESULT_ABSTAIN;
  }

  switch (dd_ns(NodeTypeWrapper_node_type)(node)) {
    case dd_ns(NodeType_EvaluatorNode):
      dd_ns(EvaluatorNode_table_t) evaluator_node = dd_ns(NodeTypeWrapper_node)(node);
      plcs_evaluation_result res = node_evaluator(evaluator_node);
      if (res == PLCS_EVAL_RESULT_TRUE) {
        matched_rule_describe_node(rule, evaluator_node);
      }
      return res;
      break;

    case dd_ns(NodeType_CompositeNode):
      dd_ns(CompositeNode_table_t) composite_node = dd_ns(NodeTypeWrapper_node)(node);
      return composite_evaluator(composite_node, depth, rule);
      break;

    default:
      // error, unknown node type!
      break;
  }

  // log error
  return PLCS_EVAL_RESULT_ABSTAIN;
}

static inline plcs_errors perform_actions(
    plcs_evaluation_result eval_res,
    dd_ns(Action_vec_t) actions_vec,
    plcs_uuid policy_id,
    int64_t policy_version,
    const char *policy_description
) {
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
      res = action_function(
          eval_res, values, values_len, dd_ns(Action_description)(action), action_id, policy_id, policy_version,
          policy_description
      );
      plcs_eval_ctx_set_action_error(action_id, res);
    } else {
      res = PLCS_EACTIONS_EVAL;
    }
  }

  return res;
}

plcs_errors evaluate_policy(dd_ns(Policy_table_t) policy) {
  // extract actions
  dd_ns(Action_vec_t) actions = dd_ns(Policy_actions)(policy);

  dd_wls_UUID_struct_t raw_policy_id = dd_ns(Policy_id)(policy);
  plcs_uuid policy_id = raw_policy_id
                            ? (plcs_uuid){.hi = dd_wls_UUID_hi(raw_policy_id), .lo = dd_wls_UUID_lo(raw_policy_id)}
                            : (plcs_uuid){0};

  // extract rules
  dd_ns(NodeTypeWrapper_table_t) rules = dd_ns(Policy_rules)(policy);

  // // evaluate rules if they exist, otherwise return EVAL_RESULT_ABSTAIN
  plcs_matched_rule matched_rule = {.buf = {'\0'}, .len = 0};
  plcs_evaluation_result eval_res = rules ? evaluate_rules(rules, 0, &matched_rule) : PLCS_EVAL_RESULT_ABSTAIN;

  // description describes the rule that triggered the actions, while falling
  // back to the policy's own description when nothing matched
  const char *description = matched_rule.len > 0 ? matched_rule.buf : dd_ns(Policy_description)(policy);

  // perform actions given evaluation result
  return perform_actions(eval_res, actions, policy_id, dd_ns(Policy_version)(policy), description);
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
