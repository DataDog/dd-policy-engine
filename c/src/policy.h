/*
 * Unless explicitly stated otherwise all files in this repository are licensed
 * under the Apache 2.0 License. This product includes software developed at
 * Datadog (https://www.datadoghq.com/).
 *
 * Copyright 2024-Present Datadog, Inc.
 */
#pragma once

#include <stdint.h>

#include <policy_reader.h>

#include "wire/dd_types.h"

/**
 * @brief Retrieves a vector of policies from the provided buffer.
 *
 * @param buffer Pointer to the buffer containing a FlatBuffers Policies root
 * object.
 * @return A vector of policies or NULL if the buffer is invalid.
 */
dd_ns(Policy_vec_t) plcs_get_policies(const uint8_t *buffer, size_t size);
