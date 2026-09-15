/*
 * Unless explicitly stated otherwise all files in this repository are licensed
 * under the Apache 2.0 License. This product includes software developed at
 * Datadog (https://www.datadoghq.com/).
 *
 * Copyright 2025-Present Datadog, Inc.
 */
#pragma once

// for plcs_uuid, which should move to a header of its own
#include <dd/policies/action.h>
#include <dd/policies/evaluation_result.h>

#include <stdint.h>
#include <stdlib.h>

/**
 * @brief What a visited node is, and which value union member to read.
 */
typedef enum plcs_node_kind {
  /** Composite node: all children must hold. */
  PLCS_NODE_AND = 0,
  /** Composite node: any child must hold. */
  PLCS_NODE_OR,
  /** Composite node: its single child must not hold. */
  PLCS_NODE_NOT,
  /** String evaluator node: read `str` from the value unions. */
  PLCS_NODE_STR_EVAL,
  /** Signed numeric evaluator node: read `num` from the value unions. */
  PLCS_NODE_NUM_EVAL,
  /** Unsigned numeric evaluator node: read `unum` from the value unions. */
  PLCS_NODE_UNUM_EVAL,
} plcs_node_kind;

/**
 * @brief One node evaluation, as it happened.
 *
 * The evaluator fields and the value unions are only meaningful for evaluator
 * nodes; they are zeroed for composites.
 *
 * @note All pointers borrow memory owned by the policy buffer and the evaluation
 *       context. They are only valid for the duration of the callback - copy
 *       anything that must outlive it.
 */
typedef struct plcs_evaluation_record {
  /** What the node is. */
  plcs_node_kind kind;
  /** What the node evaluated to. */
  plcs_evaluation_result result;
  /** Distance from the root of the policy's rule tree, which is at 0. */
  int depth;
  /** The node's own description, as authored in the policy. May be empty. */
  const char *description;
  /** A plcs_string_evaluators value when `kind` is PLCS_NODE_STR_EVAL, a
   * plcs_numeric_evaluators value for the other evaluator kinds. */
  int evaluator_id;
  /** A plcs_string_comparator value when `kind` is PLCS_NODE_STR_EVAL, a
   * plcs_numeric_comparator value for the other evaluator kinds. */
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
} plcs_evaluation_record;

/**
 * @brief Watches evaluations walk the policy trees.
 *
 * Register one with plcs_eval_ctx_set_observer() to see every policy evaluated,
 * whether or not it ends up firing an action. Without an observer the engine skips
 * the whole thing, including reading back the values each node compared.
 *
 * Each policy is framed by `policy_enter` and `policy_exit`. In between, its
 * nodes are walked: the engine enters a node, evaluates it, and leaves it, so
 * `node_enter` always has a matching `node_exit` and a node is entered before
 * the children it is built from. Nodes the evaluation never reached, because a
 * parent short-circuited, are not entered.
 *
 * Every callback may be NULL.
 */
typedef struct plcs_observer {
  /** Called before a policy's tree is walked. Everything up to the matching
   * `policy_exit` belongs to this policy. */
  void (*policy_enter)(void *user, plcs_uuid policy_id, int64_t policy_version, const char *description);
  /** Called once the policy has its result, before its actions run. Also called
   * for a policy that has no rules at all, where no node is ever entered. */
  void (*policy_exit)(void *user, plcs_evaluation_result result);
  /** Called on the way into a node, before it is evaluated. Returns a handle
   * that `node_exit` gets back, which is how an observer that stores records in
   * tree order can reserve a place for one whose result is not known yet. */
  size_t (*node_enter)(void *user, int depth);
  /** Called on the way out, with everything now known about the node. */
  void (*node_exit)(void *user, size_t handle, const plcs_evaluation_record *record);
  /** Passed back to every callback untouched. */
  void *user;
} plcs_observer;
