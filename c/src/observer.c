/*
 * Unless explicitly stated otherwise all files in this repository are licensed
 * under the Apache 2.0 License. This product includes software developed at
 * Datadog (https://www.datadoghq.com/).
 *
 * Copyright 2025-Present Datadog, Inc.
 */
#include "observer.h"

#include "eval_ctx.h"
#include "wire/boolean_operation.h"

// Describes an evaluator node: what it compared, and the two values it compared. The process
// value is read from the context after the evaluation, because an evaluator that
// scans a collection (e.g. one matching any argv element) only knows which value
// matched once it has found one, and reports it back through its context entry.
static plcs_evaluation_record describe_evaluator_node(dd_ns(EvaluatorNode_table_t) node) {
  dd_ns(EvaluatorType_union_t) evaluator = dd_ns(EvaluatorNode_eval_union)(node);
  plcs_evaluation_record record = {.description = dd_ns(EvaluatorNode_description)(node)};

  switch (evaluator.type) {
    case dd_ns(EvaluatorType_StrEvaluator): {
      dd_ns(StrEvaluator_table_t) eval_str = evaluator.value;
      plcs_string_evaluators eval_id = dd_ns(StrEvaluator_id)(eval_str);
      record.kind = PLCS_NODE_STR_EVAL;
      record.evaluator_id = eval_id;
      record.comparator = dd_ns(StrEvaluator_cmp)(eval_str);
      record.policy_value.str = dd_ns(StrEvaluator_value)(eval_str);
      record.process_value.str = plcs_eval_ctx_get_string_param(eval_id);
      break;
    }

    case dd_ns(EvaluatorType_NumEvaluator): {
      dd_ns(NumEvaluator_table_t) eval_num = evaluator.value;
      plcs_numeric_evaluators eval_id = dd_ns(NumEvaluator_id)(eval_num);
      record.kind = PLCS_NODE_NUM_EVAL;
      record.evaluator_id = eval_id;
      record.comparator = dd_ns(NumEvaluator_cmp)(eval_num);
      record.policy_value.num = dd_ns(NumEvaluator_value)(eval_num);
      record.process_value.num = plcs_eval_ctx_get_numeric_param(eval_id);
      break;
    }

    case dd_ns(EvaluatorType_UNumEvaluator): {
      dd_ns(UNumEvaluator_table_t) eval_unum = evaluator.value;
      plcs_numeric_evaluators eval_id = dd_ns(UNumEvaluator_id)(eval_unum);
      record.kind = PLCS_NODE_UNUM_EVAL;
      record.evaluator_id = eval_id;
      record.comparator = dd_ns(UNumEvaluator_cmp)(eval_unum);
      record.policy_value.unum = dd_ns(UNumEvaluator_value)(eval_unum);
      record.process_value.unum = plcs_eval_ctx_get_unumeric_param(eval_id);
      break;
    }

    // No evaluator set, or one added by a newer schema. The evaluation abstains on
    // it, so there is nothing to describe beyond the node being there.
    default:
      record.kind = PLCS_NODE_UNKNOWN;
      break;
  }

  return record;
}

// Describes a composite: which boolean operator it applies.
static plcs_evaluation_record describe_composite_node(dd_ns(CompositeNode_table_t) node) {
  plcs_evaluation_record record = {.description = dd_ns(CompositeNode_description)(node)};

  switch (dd_ns(CompositeNode_op)(node)) {
    case dd_ns(BoolOperation_BOOL_OR):
      record.kind = PLCS_NODE_OR;
      break;

    case dd_ns(BoolOperation_BOOL_NOT):
      record.kind = PLCS_NODE_NOT;
      break;

    case dd_ns(BoolOperation_BOOL_AND):
      record.kind = PLCS_NODE_AND;
      break;

    // BOOL_UNKNOWN is what a failed parse leaves behind, and the evaluation
    // abstains on it rather than treating it as any operator, so it is not
    // reported as one either.
    default:
      record.kind = PLCS_NODE_UNKNOWN;
      break;
  }

  return record;
}

plcs_evaluation_record
plcs_describe_node(dd_ns(NodeTypeWrapper_table_t) node, plcs_evaluation_result result, int depth) {
  const void *inner = dd_ns(NodeTypeWrapper_node)(node);
  plcs_evaluation_record record = dd_ns(NodeTypeWrapper_node_type)(node) == dd_ns(NodeType_EvaluatorNode)
                                      ? describe_evaluator_node(inner)
                                      : describe_composite_node(inner);

  record.result = result;
  record.depth = depth;

  return record;
}
