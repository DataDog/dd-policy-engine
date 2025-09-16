# Unless explicitly stated otherwise all files in this repository are licensed
# under the Apache 2.0 License. This product includes software developed at
# Datadog (https://www.datadoghq.com/).
#
# Copyright 2025-Present Datadog, Inc.
cmake_minimum_required(VERSION 3.28...4.0)

list(APPEND CMAKE_MODULE_PATH "${CMAKE_CURRENT_LIST_DIR}/cmake")

if (WIN32)
  cmake_policy(SET CMP0091 NEW)

  set(WINDOWS_EXPORT_ALL_SYMBOLS ON)

  if (DD_POLICY_ENABLE_STATIC_CRT)
    set(CMAKE_MSVC_RUNTIME_LIBRARY "MultiThreaded$<$<CONFIG:Debug>:Debug>")
    set(CMAKE_C_FLAGS_RELEASE "${CMAKE_C_FLAGS_RELEASE} /MT")
    set(CMAKE_C_FLAGS_DEBUG "${CMAKE_C_FLAGS_DEBUG} /MTd")
  endif ()
endif ()

project(
  dd-policy-engine
  VERSION 0.1.0
  LANGUAGES C CXX
  DESCRIPTION "Policy Engine"
)

option(DD_POLICY_BUILD_TESTS "Enable building tests" ${PROJECT_IS_TOP_LEVEL})
option(DD_POLICY_BUILD_EXAMPLES "Enable building examples" ${PROJECT_IS_TOP_LEVEL})
option(DD_POLICY_BUILD_FUZZERS "Enable building fuzzers" OFF)

if (WIN32)
  option(DD_POLICY_ENABLE_STATIC_CRT "Build libpolicy with static CRT" OFF)
endif ()

# Dependencies
find_package(flatcc REQUIRED)

# C Library
add_subdirectory(c)

if(DD_POLICY_BUILD_TESTS)
  enable_testing()
  add_subdirectory(c/src/test)
endif()

if(DD_POLICY_BUILD_EXAMPLES)
  add_subdirectory(c/examples)
endif()

if(DD_POLICY_BUILD_FUZZERS)
  add_subdirectory(c/fuzz)
endif()

# install
include(GNUInstallDirs)
include(CMakePackageConfigHelpers)

write_basic_package_version_file(
  "${PROJECT_NAME}-config-version.cmake"
  VERSION ${PROJECT_VERSION}
  COMPATIBILITY SameMajorVersion
)

configure_package_config_file(
  "${CMAKE_CURRENT_SOURCE_DIR}/cmake/${PROJECT_NAME}-config.cmake.in"
  ${PROJECT_NAME}-config.cmake
  INSTALL_DESTINATION "${CMAKE_INSTALL_LIBDIR}/cmake/${PROJECT_NAME}"
  NO_CHECK_REQUIRED_COMPONENTS_MACRO
  PATH_VARS CMAKE_INSTALL_INCLUDEDIR
)

install(
  EXPORT dd-policy-engine-targets
  NAMESPACE dd-policy-engine::
  FILE ${PROJECT_NAME}-targets.cmake
  DESTINATION "${CMAKE_INSTALL_LIBDIR}/cmake/${PROJECT_NAME}"
)

install(
  FILES 
    "${CMAKE_CURRENT_BINARY_DIR}/${PROJECT_NAME}-config.cmake"
    "${CMAKE_CURRENT_BINARY_DIR}/${PROJECT_NAME}-config-version.cmake"
  DESTINATION "${CMAKE_INSTALL_LIBDIR}/cmake/${PROJECT_NAME}"
)
