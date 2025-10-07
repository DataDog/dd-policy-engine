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

int main(int argc, char *argv[]) {
  std::string json_str;
  std::string json_file;
  std::string output_path;

  // Parse command line arguments
  cxxopts::Options options("dd-compile-policy",
                           "Compile policy JSON to FlatBuffer binary");

  options.add_options()("input-json", "Input JSON file path",
                        cxxopts::value<std::string>())(
      "output", "Output binary file path",
      cxxopts::value<std::string>())("h,help", "Print usage");

  auto result = options.parse(argc, argv);

  if (result.count("help")) {
    std::cout << options.help() << std::endl;
    return EXIT_SUCCESS;
  }

  if (!result.count("input-json") || !result.count("output")) {
    std::cerr << "Error: Both --input-json and --output are required."
              << std::endl;
    std::cerr << options.help() << std::endl;
    return EXIT_FAILURE;
  }

  json_file = result["input-json"].as<std::string>();
  output_path = result["output"].as<std::string>();

  flatbuffers::Parser parser;

  // Load binary schema (.bfbs)
  if (!parser.Deserialize(reinterpret_cast<const uint8_t *>(schema_policy_bfbs),
                          schema_policy_bfbs_len)) {
    std::cerr << "Failed to parse binary schema\n";
    return EXIT_FAILURE;
  }

  std::cout << json_file << std::endl;
  bool ok = flatbuffers::LoadFile(json_file.c_str(), false, &json_str);
  if (!ok) {
    std::cerr << "failed to open file " << json_file << std::endl;
    return EXIT_FAILURE;
  }

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
