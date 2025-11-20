/*
 * Unless explicitly stated otherwise all files in this repository are licensed
 * under the Apache 2.0 License. This product includes software developed at
 * Datadog (https://www.datadoghq.com/).
 *
 * Copyright 2025-Present Datadog, Inc.
 */
#include <dd/policies/evaluator_default.h>
#include <inttypes.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include "policy.h"

extern dd_ns(Policy_vec_t) plcs_get_policies(const uint8_t *buffer, size_t size);

void print_node_evaluator(dd_ns(EvaluatorNode_table_t) node) {
  if (!node) {
    return;  // log error?
  }

  dd_ns(EvaluatorType_union_t) evaluator = dd_ns(EvaluatorNode_eval_union)(node);

  switch (evaluator.type) {
    case dd_ns(EvaluatorType_StrEvaluator):
      dd_ns(StrEvaluator_table_t) eval_str = evaluator.value;
      plcs_string_evaluators eval_str_id = dd_ns(StrEvaluator_id)(eval_str);
      printf(
          "[str][%s][%s][%s][value: '%s']\n", dd_wls_CmpTypeSTR_name(dd_ns(StrEvaluator_cmp)(eval_str)),
          dd_wls_StringEvaluators_name(eval_str_id), dd_ns(EvaluatorNode_description)(node),
          dd_ns(StrEvaluator_value)(eval_str)
      );
      break;

    case dd_ns(EvaluatorType_NumEvaluator):
      dd_ns(NumEvaluator_table_t) eval_num = evaluator.value;
      plcs_numeric_evaluators eval_num_id = dd_ns(NumEvaluator_id)(eval_num);
      printf(
          "[num][%s][%s][%s][value: '%" PRId64 "']\n", dd_wls_CmpTypeNUM_name(dd_ns(NumEvaluator_cmp)(eval_num)),
          dd_wls_NumericEvaluators_name(eval_num_id), dd_ns(EvaluatorNode_description)(node),
          dd_ns(NumEvaluator_value)(eval_num)
      );
      break;

    case dd_ns(EvaluatorType_UNumEvaluator):
      dd_ns(UNumEvaluator_table_t) eval_unum = evaluator.value;
      plcs_numeric_evaluators eval_unum_id = dd_ns(UNumEvaluator_id)(eval_unum);
      printf(
          "[unum][%s][%s][%s][value: '%" PRId64 "']\n", dd_wls_CmpTypeNUM_name(dd_ns(UNumEvaluator_cmp)(eval_unum)),
          dd_wls_NumericEvaluators_name(eval_unum_id), dd_ns(EvaluatorNode_description)(node),
          dd_ns(UNumEvaluator_value)(eval_unum)
      );
  }
}

char get_oper_char(dd_wls_BoolOperation_enum_t oper) {
  switch (oper) {
    case dd_wls_BoolOperation_BOOL_OR:
      return '|';
    case dd_wls_BoolOperation_BOOL_AND:
      return '&';
    case dd_wls_BoolOperation_BOOL_NOT:
      return '!';
    default:
      return '?';  // unknown operation
  }
}

size_t print_node_composite(dd_ns(CompositeNode_table_t) node) {
  if (!node) {
    return 0;  // log error?
  }
  dd_ns(NodeTypeWrapper_vec_t) children = dd_ns(CompositeNode_children)(node);
  size_t children_len = children ? dd_ns(NodeTypeWrapper_vec_len)(children) : 0;
  dd_wls_BoolOperation_enum_t oper = dd_ns(CompositeNode_op)(node);
  printf("(%c%c)\n", get_oper_char(oper), get_oper_char(oper));

  return children_len;  // indicate success
}

// Recursive function to print a general tree clearly
void printTree(dd_ns(NodeTypeWrapper_table_t) node, const char *prefix, bool is_last) {
  if (!node)
    return;

  printf("%s", prefix);
  printf(is_last ? "└── " : "├── ");
  // printf("%s\n", root->data);
  size_t children_len = 0;
  dd_ns(NodeTypeWrapper_vec_t) children;

  switch (dd_ns(NodeTypeWrapper_node_type)(node)) {
    case dd_ns(NodeType_EvaluatorNode):
      dd_ns(EvaluatorNode_table_t) evaluator_node = dd_ns(NodeTypeWrapper_node)(node);
      print_node_evaluator(evaluator_node);
      return;
    case dd_ns(NodeType_CompositeNode):
      dd_ns(CompositeNode_table_t) composite_node = dd_ns(NodeTypeWrapper_node)(node);

      children = dd_ns(CompositeNode_children)(dd_ns(NodeTypeWrapper_node)(node));
      children_len = print_node_composite(composite_node);
      break;

    default:
      // error, unknown node type!
      break;
  }

  char newPrefix[1024];
  snprintf(newPrefix, sizeof(newPrefix), "%s%s", prefix, is_last ? "    " : "│   ");

  for (size_t i = 0; i < children_len; i++)
    printTree(dd_ns(NodeTypeWrapper_vec_at)(children, i), newPrefix, i == children_len - 1);
}

void print_actions(dd_ns(Policy_table_t) policy) {
  dd_ns(Action_vec_t) actions = dd_ns(Policy_actions)(policy);
  size_t actions_len = actions ? dd_ns(Action_vec_len)(actions) : 0;

  printf("  actions[%zu]\n", actions_len);

  for (size_t i = 0; i < actions_len; i++) {
    dd_ns(Action_table_t) act = dd_ns(Action_vec_at)(actions, i);
    if (!act) {
      continue;
    }

    // action ID
    dd_wls_ActionId_enum_t aid = dd_ns(Action_action)(act);
    const char *aid_name = dd_wls_ActionId_name(aid);

    // description
    const char *desc = dd_ns(Action_description)(act);
    if (!desc) {
      desc = "";
    }

    printf("    - id=%s desc='%s'\n", aid_name, desc);

    // values: vector<string>
    flatbuffers_string_vec_t vals = dd_ns(Action_values)(act);
    size_t vlen = vals ? flatbuffers_string_vec_len(vals) : 0;

    if (vlen > 0) {
      printf("      values[%zu]: ", vlen);
      for (size_t j = 0; j < vlen; j++) {
        const char *v = flatbuffers_string_vec_at(vals, j);
        printf("%s'%s'", j == 0 ? "" : ", ", v ? v : "");
      }
      printf("\n");
    }
  }
}

void print_policies(const uint8_t *buffer, size_t size) {
  dd_ns(Policy_vec_t) policies = plcs_get_policies(buffer, size);
  if (!policies) {
    // not necessarily an error, could be empty policies
    return;
  }

  size_t policies_count = dd_ns(Policy_vec_len)(policies);
  for (size_t ix = 0; ix < policies_count; ++ix) {
    printf("\n---------------[%zu]---------------\n", ix);
    dd_ns(Policy_table_t) policy = dd_ns(Policy_vec_at)(policies, ix);
    if (!policy) {
      // not necessarily an error, could be empty policy
      continue;
    }

    // 🔎 NEW: show policy description + version
    const char *p_desc = dd_ns(Policy_description)(policy);
    int64_t version = dd_ns(Policy_version)(policy);
    printf("Policy[%zu]: desc='%s' version=%" PRId64 "\n", ix, p_desc ? p_desc : "", version);

    dd_ns(NodeTypeWrapper_table_t) rules = dd_ns(Policy_rules)(policy);
    printTree(rules, "", true);

    // 🔎 NEW: show actions for this policy
    print_actions(policy);

    printf("---------------[%zu]---------------\n\n", ix);
  }
}

static uint8_t *read_file_contents(const char *filepath, size_t *out_size) {
  FILE *file = fopen(filepath, "rb");
  if (!file) {
    perror("Error opening file");
    return NULL;
  }

  // Get file size
  if (fseek(file, 0, SEEK_END) != 0) {
    perror("Error seeking file");
    fclose(file);
    return NULL;
  }

  long file_size = ftell(file);
  if (file_size < 0) {
    perror("Error getting file size");
    fclose(file);
    return NULL;
  }

  if (fseek(file, 0, SEEK_SET) != 0) {
    perror("Error seeking to start of file");
    fclose(file);
    return NULL;
  }

  // Allocate buffer
  uint8_t *buffer = (uint8_t *)malloc((size_t)file_size);
  if (!buffer) {
    perror("Memory allocation failed");
    fclose(file);
    return NULL;
  }

  // Read file
  size_t bytes_read = fread(buffer, 1, (size_t)file_size, file);
  fclose(file);

  if (bytes_read != (size_t)file_size) {
    perror("Error reading file");
    free(buffer);
    return NULL;
  }

  *out_size = bytes_read;
  return buffer;
}

int main(int argc, char *argv[]) {
  if (argc < 2) {
    fprintf(stderr, "Usage: %s <path_to_policy_file>\n", argv[0]);
    return EXIT_FAILURE;
  }

  // Read policy file
  size_t buffer_size;
  uint8_t *buffer = read_file_contents(argv[1], &buffer_size);
  if (!buffer) {
    return EXIT_FAILURE;
  }
  printf("Successfully read %zu bytes from '%s'\n", buffer_size, argv[1]);

  print_policies(buffer, buffer_size);
  free(buffer);

  printf("Policy evaluation completed successfully\n");
  return EXIT_SUCCESS;
}
