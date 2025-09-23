add_executable(dd-policy-engine.cpp-reader)

target_compile_features(dd-policy-engine.cpp-reader PUBLIC cxx_std_17)

target_sources(dd-policy-engine.cpp-reader
  PRIVATE
    ${CMAKE_CURRENT_SOURCE_DIR}/reader.cc
)

target_link_libraries(dd-policy-engine.cpp-reader
  PRIVATE
    dd-policy-engine::cpp
)
