/*
 * Unless explicitly stated otherwise all files in this repository are licensed
 * under the Apache 2.0 License. This product includes software developed at
 * Datadog (https://www.datadoghq.com/).
 *
 * Copyright 2025-Present Datadog, Inc.
 */
#pragma once

/**
 * @brief Canonical value sets for runtime languages and operating systems.
 *
 * These enums mirror the FlatBuffers definitions in fbs-schema/value_sets.fbs
 * and provide bidirectional mapping between enum values and lowercase strings.
 */

#define PLCS_LIST_RUNTIME_LANGUAGES(X)                                                                                 \
  X(RUNTIME_LANG_UNKNOWN, 0, "unknown")                                                                                \
  X(RUNTIME_LANG_JVM, 1, "jvm")                                                                                        \
  X(RUNTIME_LANG_PYTHON, 2, "python")                                                                                  \
  X(RUNTIME_LANG_RUBY, 3, "ruby")                                                                                      \
  X(RUNTIME_LANG_DOTNET, 4, "dotnet")                                                                                  \
  X(RUNTIME_LANG_NODEJS, 5, "nodejs")                                                                                  \
  X(RUNTIME_LANG_PHP, 6, "php")

#define ENUM_RUNTIME_LANG_VAL(ID, IX, STR) PLCS_##ID = IX,
typedef enum plcs_runtime_language {
  PLCS_LIST_RUNTIME_LANGUAGES(ENUM_RUNTIME_LANG_VAL) PLCS_RUNTIME_LANG__COUNT
} plcs_runtime_language;
#undef ENUM_RUNTIME_LANG_VAL

#define PLCS_LIST_OPERATING_SYSTEMS(X)                                                                                 \
  X(OS_UNKNOWN, 0, "unknown")                                                                                          \
  X(OS_LINUX, 1, "linux")                                                                                              \
  X(OS_WINDOWS, 2, "windows")                                                                                          \
  X(OS_MACOS, 3, "macos")

#define ENUM_OS_VAL(ID, IX, STR) PLCS_##ID = IX,
typedef enum plcs_operating_system { PLCS_LIST_OPERATING_SYSTEMS(ENUM_OS_VAL) PLCS_OS__COUNT } plcs_operating_system;
#undef ENUM_OS_VAL

/**
 * @brief Converts a plcs_runtime_language enum to its canonical lowercase string.
 * @param v The enum value.
 * @return The string representation, or NULL if invalid.
 */
const char *plcs_runtime_language_to_string(enum plcs_runtime_language v);

/**
 * @brief Converts a canonical lowercase string to a plcs_runtime_language enum.
 * @param s The string to look up.
 * @return The enum value, or PLCS_RUNTIME_LANG_UNKNOWN if not found.
 */
enum plcs_runtime_language plcs_runtime_language_from_string(const char *s);

/**
 * @brief Converts a plcs_operating_system enum to its canonical lowercase string.
 * @param v The enum value.
 * @return The string representation, or NULL if invalid.
 */
const char *plcs_operating_system_to_string(enum plcs_operating_system v);

/**
 * @brief Converts a canonical lowercase string to a plcs_operating_system enum.
 * @param s The string to look up.
 * @return The enum value, or PLCS_OS_UNKNOWN if not found.
 */
enum plcs_operating_system plcs_operating_system_from_string(const char *s);
