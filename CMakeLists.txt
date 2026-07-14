# Unless explicitly stated otherwise all files in this repository are licensed
# under the Apache 2.0 License. This product includes software developed at
# Datadog (https://www.datadoghq.com/).
#
# Copyright 2025-Present Datadog, Inc.
cmake_minimum_required(VERSION 4.0)

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
option(DD_POLICY_ENABLE_SANITIZERS "Enable supported runtime sanitizers" OFF)

if (WIN32)
  option(DD_POLICY_ENABLE_STATIC_CRT "Build libpolicy with static CRT" OFF)
endif ()

if(DD_POLICY_ENABLE_SANITIZERS)
  include(CheckCCompilerFlag)

  if(MSVC)
    set(DD_POLICY_SANITIZER_FLAG "/fsanitize=address")
    set(DD_POLICY_SANITIZER_COMPILE_OPTIONS "${DD_POLICY_SANITIZER_FLAG}")
  elseif(CMAKE_C_COMPILER_ID MATCHES "Clang|GNU")
    # LeakSanitizer is not a supported standalone sanitizer on Apple ARM64.
    if(APPLE)
      set(DD_POLICY_SANITIZERS "address,undefined")
    else()
      set(DD_POLICY_SANITIZERS "address,leak,undefined")
    endif()

    set(DD_POLICY_SANITIZER_FLAG "-fsanitize=${DD_POLICY_SANITIZERS}")
    set(DD_POLICY_SANITIZER_COMPILE_OPTIONS
      -fno-optimize-sibling-calls
      -fno-omit-frame-pointer
      "${DD_POLICY_SANITIZER_FLAG}"
    )
    if(CMAKE_C_COMPILER_ID MATCHES "Clang")
      list(APPEND DD_POLICY_SANITIZER_COMPILE_OPTIONS
        "-fsanitize-ignorelist=${CMAKE_SOURCE_DIR}/.sanitizers.supp"
      )
    endif()
  else()
    message(FATAL_ERROR "Sanitizers are not configured for ${CMAKE_C_COMPILER_ID}")
  endif()

  set(DD_POLICY_SAVED_REQUIRED_LINK_OPTIONS "${CMAKE_REQUIRED_LINK_OPTIONS}")
  set(CMAKE_REQUIRED_LINK_OPTIONS "${DD_POLICY_SANITIZER_FLAG}")
  string(MAKE_C_IDENTIFIER
    "${CMAKE_C_COMPILER_ID}_${DD_POLICY_SANITIZER_FLAG}"
    DD_POLICY_SANITIZER_CHECK_ID
  )
  set(DD_POLICY_SANITIZER_CHECK "DD_POLICY_SANITIZERS_SUPPORTED_${DD_POLICY_SANITIZER_CHECK_ID}")
  check_c_compiler_flag("${DD_POLICY_SANITIZER_FLAG}" ${DD_POLICY_SANITIZER_CHECK})
  set(CMAKE_REQUIRED_LINK_OPTIONS "${DD_POLICY_SAVED_REQUIRED_LINK_OPTIONS}")
  if(NOT ${DD_POLICY_SANITIZER_CHECK})
    message(FATAL_ERROR
      "The ${CMAKE_C_COMPILER_ID} compiler does not support ${DD_POLICY_SANITIZER_FLAG}"
    )
  endif()

  add_compile_options(${DD_POLICY_SANITIZER_COMPILE_OPTIONS})
  add_link_options("${DD_POLICY_SANITIZER_FLAG}")
  message(STATUS "Enabled sanitizers: ${DD_POLICY_SANITIZER_FLAG}")
endif()

# Dependencies
include(FetchContent)
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

# Top-level developer targets. The Go suite includes a build-tagged cgo smoke
# test that links the exact native archive produced by this CMake build.
if(PROJECT_IS_TOP_LEVEL)
  find_program(DD_POLICY_GO_EXECUTABLE
    NAMES go
    DOC "Go executable used by the go-test target"
  )

  if(DD_POLICY_GO_EXECUTABLE AND NOT MSVC)
    set(DD_POLICY_CGO_CFLAGS
      "-I \"${CMAKE_CURRENT_SOURCE_DIR}/c/include\""
    )
    set(DD_POLICY_GO_TEST_ENV
      "CC=${CMAKE_C_COMPILER}"
      "CGO_ENABLED=1"
      "CGO_CFLAGS=${DD_POLICY_CGO_CFLAGS}"
      "GOCACHE=${CMAKE_CURRENT_BINARY_DIR}/go-build-cache"
    )
    set(DD_POLICY_GO_EXTLDFLAGS
      "$<TARGET_FILE:dd-policy-engine.static-lib>"
    )
    if(DD_POLICY_ENABLE_SANITIZERS)
      string(APPEND DD_POLICY_GO_EXTLDFLAGS
        " ${DD_POLICY_SANITIZER_FLAG}"
      )
    endif()

    file(GLOB DD_POLICY_CGO_SMOKE_SOURCES CONFIGURE_DEPENDS
      "${CMAKE_CURRENT_SOURCE_DIR}/go/internal/cgosmoke/*.go"
    )
    file(GLOB_RECURSE DD_POLICY_PUBLIC_HEADERS CONFIGURE_DEPENDS
      "${CMAKE_CURRENT_SOURCE_DIR}/c/include/*.h"
    )
    set(DD_POLICY_CGO_SMOKE_BINARY
      "${CMAKE_CURRENT_BINARY_DIR}/go-cgo-smoke-$<CONFIG>${CMAKE_EXECUTABLE_SUFFIX}"
    )

    add_custom_command(
      OUTPUT "${DD_POLICY_CGO_SMOKE_BINARY}"
      COMMAND "${CMAKE_COMMAND}" -E make_directory
              "${CMAKE_CURRENT_BINARY_DIR}/go-build-cache"
      COMMAND "${CMAKE_COMMAND}" -E env
              ${DD_POLICY_GO_TEST_ENV}
              "${DD_POLICY_GO_EXECUTABLE}" test -c
              -tags=cmake_cgo
              "-ldflags=-extldflags '${DD_POLICY_GO_EXTLDFLAGS}'"
              -o "${DD_POLICY_CGO_SMOKE_BINARY}"
              ./internal/cgosmoke
      DEPENDS
        dd-policy-engine.static-lib
        ${DD_POLICY_CGO_SMOKE_SOURCES}
        ${DD_POLICY_PUBLIC_HEADERS}
        "${CMAKE_CURRENT_SOURCE_DIR}/go/go.mod"
        "${CMAKE_CURRENT_SOURCE_DIR}/go/go.sum"
      WORKING_DIRECTORY "${CMAKE_CURRENT_SOURCE_DIR}/go"
      COMMENT "Linking the cgo smoke test against the CMake-built policy engine"
      VERBATIM
    )
    add_custom_target(go-cgo-smoke
      DEPENDS "${DD_POLICY_CGO_SMOKE_BINARY}"
    )

    add_custom_target(go-unit-tests
      COMMAND "${CMAKE_COMMAND}" -E env
              "GOCACHE=${CMAKE_CURRENT_BINARY_DIR}/go-build-cache"
              "${DD_POLICY_GO_EXECUTABLE}" test ./...
      WORKING_DIRECTORY "${CMAKE_CURRENT_SOURCE_DIR}/go"
      COMMENT "Running Go tests"
      VERBATIM
    )
    add_custom_target(go-cgo-test
      COMMAND "${DD_POLICY_CGO_SMOKE_BINARY}"
      COMMENT "Running the precompiled cgo smoke test"
      VERBATIM
    )
    add_dependencies(go-cgo-test go-cgo-smoke)

    add_custom_target(go-test)
    add_dependencies(go-test go-unit-tests go-cgo-test)
    message(STATUS "Go test target enabled: ${DD_POLICY_GO_EXECUTABLE}")
  else()
    if(NOT DD_POLICY_GO_EXECUTABLE)
      set(DD_POLICY_GO_TEST_UNAVAILABLE
        "The go-test target requires Go on PATH or DD_POLICY_GO_EXECUTABLE to be set."
      )
    else()
      set(DD_POLICY_GO_TEST_UNAVAILABLE
        "The go-test target requires a GCC-compatible CMake compiler; MSVC archives cannot be linked reliably by cgo."
      )
    endif()
    add_custom_target(go-test
      COMMAND "${CMAKE_COMMAND}" -E echo
              "${DD_POLICY_GO_TEST_UNAVAILABLE}"
      COMMAND "${CMAKE_COMMAND}" -E false
      USES_TERMINAL
      VERBATIM
    )
    message(STATUS "${DD_POLICY_GO_TEST_UNAVAILABLE}")
  endif()

  if(DD_POLICY_BUILD_TESTS)
    add_custom_target(c-test
      COMMAND "${CMAKE_CTEST_COMMAND}"
              --test-dir "${CMAKE_CURRENT_BINARY_DIR}"
              --build-config "$<CONFIG>"
              --output-on-failure
      DEPENDS dd-policy-engine.unit-tests
      COMMENT "Running C policy engine tests"
      USES_TERMINAL
      VERBATIM
    )
  endif()

  add_custom_target(check)
  add_dependencies(check go-test)
  if(TARGET c-test)
    add_dependencies(check c-test)
  endif()
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
