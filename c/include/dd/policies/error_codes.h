/*
 * Unless explicitly stated otherwise all files in this repository are licensed
 * under the Apache 2.0 License. This product includes software developed at
 * Datadog (https://www.datadoghq.com/).
 *
 * Copyright 2025-Present Datadog, Inc.
 */
#pragma once

/**
 * @brief Error codes for policy evaluation.
 *
 */
#define PLCS_LIST_ERRORS(X)                                                                                            \
  X(SUCCESS, 0)                                                                                                        \
  X(REGISTER_EVAL_PTR, 1)                                                                                              \
  X(IX_OVERFLOW, 2)                                                                                                    \
  X(NULL_PTR, 3)                                                                                                       \
  X(INITIZLIED, 4)                                                                                                     \
  X(NO_DATA, 5)                                                                                                        \
  X(UNKNOWN_EVAL_IX, 6)                                                                                                \
  X(ACTIONS_EVAL, 7)                                                                                                   \
  X(UNKNOWN_CMP, 8)

#define ENUM_VAL(VAL, IX) PLCS_E##VAL = IX,

typedef enum plcs_errors { PLCS_LIST_ERRORS(ENUM_VAL) } plcs_errors;

#undef ENUM_VAL
