/*
 * Unless explicitly stated otherwise all files in this repository are licensed
 * under the Apache 2.0 License. This product includes software developed at
 * Datadog (https://www.datadoghq.com/).
 *
 * Copyright 2025-Present Datadog, Inc.
 */

#include <dd/policies/value_sets.h>
#include <string.h>

typedef struct {
  enum plcs_runtime_language value;
  const char *str;
} runtime_language_entry;

typedef struct {
  enum plcs_operating_system value;
  const char *str;
} operating_system_entry;

#define RUNTIME_LANG_ENTRY(ID, IX, STR) {PLCS_##ID, STR},
static const runtime_language_entry runtime_language_table[] = {PLCS_LIST_RUNTIME_LANGUAGES(RUNTIME_LANG_ENTRY)};
#undef RUNTIME_LANG_ENTRY

#define OS_ENTRY(ID, IX, STR) {PLCS_##ID, STR},
static const operating_system_entry operating_system_table[] = {PLCS_LIST_OPERATING_SYSTEMS(OS_ENTRY)};
#undef OS_ENTRY

static const size_t runtime_language_count = sizeof(runtime_language_table) / sizeof(runtime_language_table[0]);
static const size_t operating_system_count = sizeof(operating_system_table) / sizeof(operating_system_table[0]);

const char *plcs_runtime_language_to_string(enum plcs_runtime_language v) {
  for (size_t i = 0; i < runtime_language_count; i++) {
    if (runtime_language_table[i].value == v) {
      return runtime_language_table[i].str;
    }
  }
  return NULL;
}

enum plcs_runtime_language plcs_runtime_language_from_string(const char *s) {
  if (s == NULL) {
    return PLCS_RUNTIME_LANG_UNKNOWN;
  }
  for (size_t i = 0; i < runtime_language_count; i++) {
    if (strcmp(runtime_language_table[i].str, s) == 0) {
      return runtime_language_table[i].value;
    }
  }
  if (strcmp(s, "java") == 0) {
    return PLCS_RUNTIME_LANG_JVM;
  }
  return PLCS_RUNTIME_LANG_UNKNOWN;
}

const char *plcs_operating_system_to_string(enum plcs_operating_system v) {
  for (size_t i = 0; i < operating_system_count; i++) {
    if (operating_system_table[i].value == v) {
      return operating_system_table[i].str;
    }
  }
  return NULL;
}

enum plcs_operating_system plcs_operating_system_from_string(const char *s) {
  if (s == NULL) {
    return PLCS_OS_UNKNOWN;
  }
  for (size_t i = 0; i < operating_system_count; i++) {
    if (strcmp(operating_system_table[i].str, s) == 0) {
      return operating_system_table[i].value;
    }
  }
  return PLCS_OS_UNKNOWN;
}
