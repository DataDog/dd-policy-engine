#include <flatbuffers/flatbuffers.h>
#include <flatbuffers/idl.h>
#include <flatbuffers/minireflect.h>
#include <flatbuffers/reflection.h>
#include <flatbuffers/util.h>

#include <fstream>
#include <iostream>
#include <string>
#include <vector>

#include "cxxopts.hpp"
#include "policy_schema_bfbs.h" // Contains `schema_policy_bfbs` and `schema_policy_bfbs_len`

// dd-compile-policy ships bundled inside the Agent, which pins it
// to a specific version, so its baked-in schema can lag newly added
// string/numeric evaluators. The datadog-apm-inject package updates on a
// separate cadence and can drop a newer schema at this fixed path to unblock
// new evaluators without an Agent/compiler version bump.
#if defined(_WIN32)
constexpr char kInjectorSchemaPath[] =
    "C:\\ProgramData\\Datadog\\Installer\\packages\\datadog-apm-inject\\stable\\dll\\policy.bfbs";
#else
constexpr char kInjectorSchemaPath[] =
    "/opt/datadog-packages/datadog-apm-inject/stable/inject/policy.bfbs";
#endif

int main(int argc, char *argv[]) {
  std::string json_str;
  std::string json_file;
  std::string output_path;

  // Parse command line arguments
  cxxopts::Options options("dd-compile-policy",
                           "Compile policy JSON to FlatBuffer binary");

  options.add_options()("input-file", "Input JSON file path",
                        cxxopts::value<std::string>())(
      "input-string", "Input JSON string", cxxopts::value<std::string>())(
      "output-file", "Output binary file path", cxxopts::value<std::string>())(
      "schema-file",
      "Path to a binary FlatBuffers schema (.bfbs) to compile against, "
      "overriding the built-in policy schema",
      cxxopts::value<std::string>())("h,help", "Print usage");

  auto result = options.parse(argc, argv);

  if (result.count("help")) {
    std::cout << options.help() << std::endl;
    return EXIT_SUCCESS;
  }

  // Validate that exactly one input method is provided
  const bool has_input_file = result.count("input-file");
  const bool has_input_string = result.count("input-string");

  if (!has_input_file && !has_input_string) {
    std::cerr << "Error: Either --input-file or --input-string is required."
              << std::endl;
    std::cerr << options.help() << std::endl;
    return EXIT_FAILURE;
  }

  if (has_input_file && has_input_string) {
    std::cerr << "Error: Cannot specify both --input-file and --input-string. "
                 "Use only one."
              << std::endl;
    std::cerr << options.help() << std::endl;
    return EXIT_FAILURE;
  }

  if (!result.count("output-file")) {
    std::cerr << "Error: --output-file is required." << std::endl;
    std::cerr << options.help() << std::endl;
    return EXIT_FAILURE;
  }

  if (has_input_file) {
    json_file = result["input-file"].as<std::string>();
  }
  if (has_input_string) {
    json_str = result["input-string"].as<std::string>();
  }
  output_path = result["output-file"].as<std::string>();

  flatbuffers::Parser parser;

  // Forward compatibility: a newer remote config payload may carry fields this
  // (older) compiler's schema doesn't know about. Ignore them instead of
  // rejecting the whole policy set, so an old agent still applies the parts it
  // understands. Unknown enum values already degrade to the schema default
  // (`*_UNKNOWN`), which the engine treats as "abstain".
  parser.opts.skip_unexpected_fields_in_json = true;

  // Load binary schema (.bfbs), in order of preference:
  //   1. an explicit --schema-file, if provided (hard error if unreadable);
  //   2. the schema at kInjectorSchemaPath, updated independently of this
  //      binary by the injector (silently skipped if absent/unreadable);
  //   3. the schema built into this binary.
  const uint8_t *schema_bytes =
      reinterpret_cast<const uint8_t *>(schema_policy_bfbs);
  size_t schema_len = schema_policy_bfbs_len;
  std::string schema_file_buf;

  if (result.count("schema-file")) {
    const std::string schema_path = result["schema-file"].as<std::string>();
    if (!flatbuffers::LoadFile(schema_path.c_str(), true, &schema_file_buf)) {
      std::cerr << "failed to open schema file " << schema_path << std::endl;
      return EXIT_FAILURE;
    }
    schema_bytes = reinterpret_cast<const uint8_t *>(schema_file_buf.data());
    schema_len = schema_file_buf.size();
  } else if (flatbuffers::LoadFile(kInjectorSchemaPath, true,
                                    &schema_file_buf)) {
    schema_bytes = reinterpret_cast<const uint8_t *>(schema_file_buf.data());
    schema_len = schema_file_buf.size();
  }

  if (!parser.Deserialize(schema_bytes, schema_len)) {
    std::cerr << "Failed to parse binary schema\n";
    return EXIT_FAILURE;
  }

  // Load JSON content based on input method
  if (has_input_file) {
    const bool ok = flatbuffers::LoadFile(json_file.c_str(), false, &json_str);
    if (!ok) {
      std::cerr << "failed to open file " << json_file << std::endl;
      return EXIT_FAILURE;
    }
  }
  // For input-string, json_str is already set from command line argument

  // Parse JSON string
  if (!parser.ParseJson(json_str.c_str())) {
    std::cerr << "Failed to parse JSON: " << parser.error_ << std::endl;
    return EXIT_FAILURE;
  }

  // Write FlatBuffer binary
  std::ofstream out(output_path, std::ios::binary);
  if (!out) {
    std::cerr << "Failed to open output file: " << output_path << std::endl;
    return EXIT_FAILURE;
  }
  out.write(reinterpret_cast<const char *>(parser.builder_.GetBufferPointer()),
            parser.builder_.GetSize());

  return EXIT_SUCCESS;
}
