include(FetchContent)

set(FLATBUFFERS_BUILD_TESTS OFF CACHE BOOL "" FORCE)
set(FLATBUFFERS_INSTALL OFF CACHE BOOL "" FORCE)
set(FLATBUFFERS_BUILD_FLATC OFF CACHE BOOL "" FORCE)
set(FLATBUFFERS_LIBCXX_WITH_CLANG OFF CACHE BOOL "" FORCE)

# Flatbuffers isn't working too well on musl, so we need to patch it.
# See https://github.com/sartura/flatbuffers/commit/92bd62407329caacd66e92e5bfd2949f2f137bfe
FetchContent_Declare(
  flatbuffers
  GIT_REPOSITORY https://github.com/google/flatbuffers.git
  GIT_TAG v25.9.23
  PATCH_COMMAND git apply ${CMAKE_CURRENT_SOURCE_DIR}/flatbuffers.patch
  EXCLUDE_FROM_ALL
)

FetchContent_MakeAvailable(flatbuffers)

if(NOT TARGET flatbuffers)
  message(FATAL_ERROR "Required target flatbuffers was not imported")
endif()

if(NOT TARGET flatbuffers::flatbuffers)
  add_library(flatbuffers::flatbuffers ALIAS flatbuffers)
endif()

set(CXXOPTS_VERSION 3.2.1)

function(cxxopts_getversion version_arg)
  message(STATUS "cxxopts version ${CXXOPTS_VERSION}")
  set(${version_arg} ${CXXOPTS_VERSION} PARENT_SCOPE)
endfunction()

# Cxxopts
FetchContent_Declare(
  cxxopts
  GIT_REPOSITORY https://github.com/jarro2783/cxxopts.git
  GIT_TAG v${CXXOPTS_VERSION}
  EXCLUDE_FROM_ALL
  FIND_PACKAGE_ARGS NAMES cxxopts ${CXXOPTS_VERSION}
)

FetchContent_MakeAvailable(cxxopts)

if(NOT TARGET cxxopts::cxxopts)
  message(FATAL_ERROR "Required target cxxopts::cxxopts was not imported")
endif()

add_executable(dd-policy-engine.compile-policy)

target_compile_features(dd-policy-engine.compile-policy
  PUBLIC
    cxx_std_17
)

target_sources(dd-policy-engine.compile-policy
  PRIVATE
    src/compile_policy.cc
)

target_link_libraries(dd-policy-engine.compile-policy
  PRIVATE
    cxxopts::cxxopts
    flatbuffers::flatbuffers
)

set_target_properties(dd-policy-engine.compile-policy
  PROPERTIES 
    OUTPUT_NAME "dd-compile-policy"
)
