# Unless explicitly stated otherwise all files in this repository are licensed
# under the Apache 2.0 License. This product includes software developed at
# Datadog (https://www.datadoghq.com/).
#
# Copyright 2025-Present Datadog, Inc.

# flatcc update globals C and CXX flags. Cache our C and CXX flags to reset later.
set(OLD_CMAKE_C_FLAGS ${CMAKE_C_FLAGS})
set(OLD_CMAKE_CXX_FLAGS ${CMAKE_CXX_FLAGS})

set(FLATCC_TEST OFF CACHE BOOL "" FORCE)
set(FLATCC_CXX_TEST OFF CACHE BOOL "" FORCE)
set(FLATCC_DEBUG_CLANG_SANITIZE OFF CACHE BOOL "" FORCE)
set(FLATCC_REFLECTION OFF CACHE BOOL "" FORCE)
set(FLATCC_ALLOW_WERROR OFF CACHE BOOL "" FORCE)

set(CMAKE_POLICY_VERSION_MINIMUM "3.5")

FetchContent_Declare(
  flatcc
  GIT_REPOSITORY https://github.com/dvidelabs/flatcc.git
  GIT_TAG 47af7e601f511e80bcb85f28adf06af27c6a6b00
  PATCH_COMMAND git apply ${CMAKE_CURRENT_SOURCE_DIR}/cmake/flatcc.patch
  EXCLUDE_FROM_ALL
)

FetchContent_MakeAvailable(flatcc)

# Reset our flags.
set(CMAKE_C_FLAGS ${OLD_CMAKE_C_FLAGS})
set(CMAKE_CXX_FLAGS ${OLD_CMAKE_CXX_FLAGS})

if(NOT TARGET flatccrt-obj)
  message(FATAL_ERROR "Required target flatccrt-obj was not imported")
endif()

