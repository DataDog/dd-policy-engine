add_executable(dd-policy-engine.unit-tests)
target_sources(dd-policy-engine.unit-tests
  PRIVATE
    ${CMAKE_CURRENT_SOURCE_DIR}/test_actions.c
    ${CMAKE_CURRENT_SOURCE_DIR}/test_eval_ctx.c
    ${CMAKE_CURRENT_SOURCE_DIR}/test_evaluator.c
    ${CMAKE_CURRENT_SOURCE_DIR}/test_main.c
    ${CMAKE_CURRENT_SOURCE_DIR}/test_policy_reader.c
)

target_include_directories(dd-policy-engine.unit-tests
  PRIVATE
    ${CMAKE_BINARY_DIR}/c
    ${CMAKE_SOURCE_DIR}/c/src
)

target_link_libraries(dd-policy-engine.unit-tests
  PUBLIC
    dd-policy-engine::obj
)

add_test(
  NAME dd-policy-engine.unit-tests 
  COMMAND dd-policy-engine.unit-tests
)
