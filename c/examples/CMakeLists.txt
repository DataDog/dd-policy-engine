# Unless explicitly stated otherwise all files in this repository are licensed
# under the Apache 2.0 License. This product includes software developed at
# Datadog (https://www.datadoghq.com/).
#
# Copyright 2024-Present Datadog, Inc.
add_executable(basic-reader)
target_sources(basic-reader
  PRIVATE
    ${CMAKE_CURRENT_SOURCE_DIR}/basic-reader.c
)
target_link_libraries(basic-reader
  PUBLIC
    dd-policy-engine::static-lib
)

add_executable(hardcoded-header-reader)
target_sources(hardcoded-header-reader
  PRIVATE
    ${CMAKE_CURRENT_SOURCE_DIR}/hardcoded-header-reader.c
)
target_link_libraries(hardcoded-header-reader
  PUBLIC
    dd-policy-engine::static-lib
)

if(LINUX)
  add_executable(mmap-reader)
  target_sources(mmap-reader
    PRIVATE
      ${CMAKE_CURRENT_SOURCE_DIR}/mmap-reader.c
  )
  target_link_libraries(mmap-reader
    PUBLIC
      dd-policy-engine::static-lib
  )
endif()

add_executable(hardcoded_cmdline_evaluator)
target_sources(hardcoded_cmdline_evaluator
  PRIVATE
    ${CMAKE_CURRENT_SOURCE_DIR}/hardcoded_cmdline_evaluator.c
)
target_link_libraries(hardcoded_cmdline_evaluator
  PUBLIC
    dd-policy-engine::static-lib
)
