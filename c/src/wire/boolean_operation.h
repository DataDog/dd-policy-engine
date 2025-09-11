/**
 * Unless explicitly stated otherwise all files in this repository are licensed
 * under the Apache 2.0 License. This product includes software developed at
 * Datadog (https://www.datadoghq.com/).
 *
 * Copyright 2024-Present Datadog, Inc.
 * -----
 * @file wire/boolean_operation.h
 * @brief Internal translation for plcs_boolean_operation enum.
 *
 * @details
 * Bridges the public `plcs_boolean_operation` enum (from `../../include/boolean_operation.h`)
 * to the on-the-wire / generated values from the FlatBuffers schema (via
 * `boolean_operation_reader.h`). Provides helpers to translate both ways and a
 * small compile-time check to keep mappings in sync.
 *
 * @note
 * This is **private** library glue — not installed or exported. Do not include
 * it from public headers. External code should include the public header in
 * `../../include/` and use only the public enum names.
 *
 * Guidelines:
 *  - Keep the mapping aligned with both public and vendor enums
 *  - Use `_Static_assert` to guard table size
 *  - Switch without `default:` and cast to the vendor enum type so `-Wswitch-enum`
 *    can catch drift when the schema grows
 */

#pragma once

#include <boolean_operation_reader.h> /* FlatBuffers generated headers */
#include "dd_types.h"                 /* dd_ns(...) */

#ifdef __cplusplus
extern "C" {
#endif

#ifdef __cplusplus
}
#endif
