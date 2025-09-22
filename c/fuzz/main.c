/*
 * Unless explicitly stated otherwise all files in this repository are licensed
 * under the Apache 2.0 License. This product includes software developed at
 * Datadog (https://www.datadoghq.com/).
 *
 * Copyright 2024-Present Datadog, Inc.
 */
#include <dd/policies/policies.h>
#include <stdint.h>

int LLVMFuzzerTestOneInput(const uint8_t *data, size_t size) {
  plcs_evaluate_buffer(data, size);
  return 0;
}
