/*
 * Unless explicitly stated otherwise all files in this repository are licensed
 * under the Apache 2.0 License. This product includes software developed at
 * Datadog (https://www.datadoghq.com/).
 *
 * Copyright 2024-Present Datadog, Inc.
 */
#pragma once
#include <flatcc/flatcc_flatbuffers.h>

/**
 * @brief These are flatbuffers helpers as advised by the flatbuffers
 * documentation.
 */

/**
 * @brief Translates an enum value to a new namespace.
 *
 */
#define TRANSLATE_RENAME_ENUM(orig, new, ns) new = ns##_##orig
#define TRANSLATE_ENUM(name, ns) TRANSLATE_RENAME_ENUM(name, name, ns)

#undef dd_ns
/**
 * @brief A namespace wrapper for flatbuffers types.
 *
 */
#define dd_ns(x) FLATBUFFERS_WRAP_NAMESPACE(dd_wls, x)
