include(FetchContent)

set(FLATCC_TEST OFF CACHE BOOL "" FORCE)
set(FLATCC_CXX_TEST OFF CACHE BOOL "" FORCE)
set(FLATCC_DEBUG_CLANG_SANITIZE OFF CACHE BOOL "" FORCE)
set(FLATCC_PORTABLE OFF CACHE BOOL "" FORCE)
set(FLATCC_ALLOW_WERROR OFF CACHE BOOL "" FORCE)
set(FLATCC_RTONLY ON CACHE BOOL "" FORCE)
set(FLATCC_INSTALL ON CACHE BOOL "" FORCE)

set(CMAKE_POLICY_VERSION_MINIMUM "3.5")
cmake_policy(SET CMP0175 OLD)

# Temporarily suppres INFO messages from `flatc`. Keep a backup of our C flags because
# `flatcc` updates global C flags.
set(_OLD_C_FLAGS ${CMAKE_C_FLAGS})
set(_OLD_LOG_LEVEL ${CMAKE_MESSAGE_LOG_LEVEL})
set(CMAKE_MESSAGE_LOG_LEVEL WARNING)

FetchContent_Declare(
  flatcc
  GIT_REPOSITORY "https://github.com/dvidelabs/flatcc.git"
  GIT_TAG 47af7e601f511e80bcb85f28adf06af27c6a6b00
  FIND_PACKAGE_ARGS NAMES flatcc
  SYSTEM
  EXCLUDE_FROM_ALL
)

FetchContent_MakeAvailable(flatcc)

# Reset C flags and log level.
set(CMAKE_C_FLAGS ${_OLD_C_FLAGS})
set(CMAKE_MESSAGE_LOG_LEVEL ${_OLD_LOG_LEVEL})

if(NOT TARGET flatccrt)
  message(FATAL_ERROR "flatcrt target was not imported")
endif()

# Fixing flatccrt target
get_target_property(FLATCC_INCLUDE_DIRS flatccrt INTERFACE_INCLUDE_DIRECTORIES)
set_target_properties(flatccrt PROPERTIES 
    INTERFACE_INCLUDE_DIRECTORIES ""
)
target_include_directories(flatccrt INTERFACE 
    $<BUILD_INTERFACE:${flatcc_SOURCE_DIR}/include>
    $<INSTALL_INTERFACE:include>
)

add_library(flatcc::crt ALIAS flatccrt)
