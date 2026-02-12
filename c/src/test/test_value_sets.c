/*
 * Unless explicitly stated otherwise all files in this repository are licensed
 * under the Apache 2.0 License. This product includes software developed at
 * Datadog (https://www.datadoghq.com/).
 *
 * Copyright 2025-Present Datadog, Inc.
 */
#include "utest/utest.h"

#include <dd/policies/value_sets.h>

#include <stddef.h>

/* ---- RuntimeLanguage: enum → string ---- */

UTEST(value_sets, runtime_language_to_string_known) {
  ASSERT_STREQ("jvm", plcs_runtime_language_to_string(PLCS_RUNTIME_LANG_JVM));
  ASSERT_STREQ("python", plcs_runtime_language_to_string(PLCS_RUNTIME_LANG_PYTHON));
  ASSERT_STREQ("ruby", plcs_runtime_language_to_string(PLCS_RUNTIME_LANG_RUBY));
  ASSERT_STREQ("dotnet", plcs_runtime_language_to_string(PLCS_RUNTIME_LANG_DOTNET));
  ASSERT_STREQ("nodejs", plcs_runtime_language_to_string(PLCS_RUNTIME_LANG_NODEJS));
  ASSERT_STREQ("php", plcs_runtime_language_to_string(PLCS_RUNTIME_LANG_PHP));
  ASSERT_STREQ("unknown", plcs_runtime_language_to_string(PLCS_RUNTIME_LANG_UNKNOWN));
}

UTEST(value_sets, runtime_language_to_string_invalid) {
  ASSERT_TRUE(plcs_runtime_language_to_string((enum plcs_runtime_language)99) == NULL);
}

/* ---- RuntimeLanguage: string → enum ---- */

UTEST(value_sets, runtime_language_from_string_known) {
  ASSERT_EQ((int)PLCS_RUNTIME_LANG_JVM, (int)plcs_runtime_language_from_string("jvm"));
  ASSERT_EQ((int)PLCS_RUNTIME_LANG_PYTHON, (int)plcs_runtime_language_from_string("python"));
  ASSERT_EQ((int)PLCS_RUNTIME_LANG_RUBY, (int)plcs_runtime_language_from_string("ruby"));
  ASSERT_EQ((int)PLCS_RUNTIME_LANG_DOTNET, (int)plcs_runtime_language_from_string("dotnet"));
  ASSERT_EQ((int)PLCS_RUNTIME_LANG_NODEJS, (int)plcs_runtime_language_from_string("nodejs"));
  ASSERT_EQ((int)PLCS_RUNTIME_LANG_PHP, (int)plcs_runtime_language_from_string("php"));
}

UTEST(value_sets, runtime_language_from_string_java_alias) {
  ASSERT_EQ((int)PLCS_RUNTIME_LANG_JVM, (int)plcs_runtime_language_from_string("java"));
}

UTEST(value_sets, runtime_language_from_string_unknown) {
  ASSERT_EQ((int)PLCS_RUNTIME_LANG_UNKNOWN, (int)plcs_runtime_language_from_string("go"));
  ASSERT_EQ((int)PLCS_RUNTIME_LANG_UNKNOWN, (int)plcs_runtime_language_from_string(""));
  ASSERT_EQ((int)PLCS_RUNTIME_LANG_UNKNOWN, (int)plcs_runtime_language_from_string(NULL));
}

/* ---- OperatingSystem: enum → string ---- */

UTEST(value_sets, operating_system_to_string_known) {
  ASSERT_STREQ("linux", plcs_operating_system_to_string(PLCS_OS_LINUX));
  ASSERT_STREQ("windows", plcs_operating_system_to_string(PLCS_OS_WINDOWS));
  ASSERT_STREQ("macos", plcs_operating_system_to_string(PLCS_OS_MACOS));
  ASSERT_STREQ("unknown", plcs_operating_system_to_string(PLCS_OS_UNKNOWN));
}

UTEST(value_sets, operating_system_to_string_invalid) {
  ASSERT_TRUE(plcs_operating_system_to_string((enum plcs_operating_system)99) == NULL);
}

/* ---- OperatingSystem: string → enum ---- */

UTEST(value_sets, operating_system_from_string_known) {
  ASSERT_EQ((int)PLCS_OS_LINUX, (int)plcs_operating_system_from_string("linux"));
  ASSERT_EQ((int)PLCS_OS_WINDOWS, (int)plcs_operating_system_from_string("windows"));
  ASSERT_EQ((int)PLCS_OS_MACOS, (int)plcs_operating_system_from_string("macos"));
}

UTEST(value_sets, operating_system_from_string_unknown) {
  ASSERT_EQ((int)PLCS_OS_UNKNOWN, (int)plcs_operating_system_from_string("freebsd"));
  ASSERT_EQ((int)PLCS_OS_UNKNOWN, (int)plcs_operating_system_from_string(""));
  ASSERT_EQ((int)PLCS_OS_UNKNOWN, (int)plcs_operating_system_from_string(NULL));
}
