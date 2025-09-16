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
typedef enum plcs_errors {
  DD_ESUCCESS = 0,
  DD_EREGISTER_EVAL_PTR,
  DD_EIX_OVERFLOW,
  DD_ENULL_PTR,
  DD_EINITIZLIED,
  DD_ENO_DATA,
  DD_EUNKNOWN_EVAL_IX,
  DD_EACTIONS_EVAL,
  DD_EUNKNOWN_CMP,
} plcs_errors;
