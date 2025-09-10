add_library(dd-policy-engine.cpp STATIC)
add_library(dd-policy-engine::cpp ALIAS dd-policy-engine.cpp)

target_compile_features(dd-policy-engine.cpp PUBLIC cxx_std_17)

target_sources(dd-policy-engine.cpp
  PUBLIC
    FILE_SET cpp_public_headers
    TYPE HEADERS
    BASE_DIRS ${CMAKE_CURRENT_SOURCE_DIR}/include
    FILES
      ${CMAKE_CURRENT_SOURCE_DIR}/include/dd/policies/policies.hpp
)

set_target_properties(dd-policy-engine.cpp
  PROPERTIES
    VERIFY_INTERFACE_HEADER_SETS ON
)

target_link_libraries(dd-policy-engine.cpp
  PUBLIC
    dd-policy-engine::obj
)

install(
  TARGETS dd-policy-engine.cpp
  EXPORT ${PROJECT_NAME}-export
  FILE_SET public_headers
  FILE_SET cpp_public_headers
  LIBRARY DESTINATION ${CMAKE_INSTALL_LIBDIR}
  ARCHIVE DESTINATION ${CMAKE_INSTALL_LIBDIR}
  RUNTIME DESTINATION ${CMAKE_INSTALL_LIBDIR}
)

if(DD_POLICY_BUILD_EXAMPLES)
  add_subdirectory(examples)
endif()
