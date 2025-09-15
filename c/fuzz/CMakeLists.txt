# Unless explicitly stated otherwise all files in this repository are licensed
# under the Apache 2.0 License. This product includes software developed at
# Datadog (https://www.datadoghq.com/).
#
# Copyright 2024-Present Datadog, Inc.

add_executable(dd-policy-engine.fuzz main.c)

target_link_libraries(dd-policy-engine.fuzz
  PRIVATE
   dd-policy-engine::static-lib
)
