/*
 * Unless explicitly stated otherwise all files in this repository are licensed
 * under the Apache 2.0 License. This product includes software developed at
 * Datadog (https://www.datadoghq.com/).
 *
 * Copyright 2024-Present Datadog, Inc.
 */
#pragma once

/**
 * @brief defines a new type for boolean results, similar to optional where
 * EVAL_RESULT_ABSTAIN defines a 'dont-care' (optional) state
 *
 */
typedef enum plcs_evaluation_result {
  EVAL_RESULT_TRUE = 0,
  EVAL_RESULT_FALSE = 1,
  EVAL_RESULT_ABSTAIN = 2,
  EVAL_RESULT__COUNT
} plcs_evaluation_result;

const char *plcs_evaluation_result_to_string(enum plcs_evaluation_result res);
