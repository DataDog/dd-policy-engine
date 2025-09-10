#include <dd/policies/policies.hpp>

#include <iostream>

namespace plcs = datadog::policies;

int main(int argc, char *argv[]) {
  if (argc < 2) {
    std::cerr << "wrong syntax: " << argv[0] << " <BUFFER_FILE>\n";
    return 1;
  }

  plcs::init();

  plcs::set_params(plcs::StringEvaluator::PROCESS_EXE, argv[0]);
  plcs::set_params(plcs::StringEvaluator::RUNTIME_LANGUAGE, "cpp");

  plcs::register_evaluator(
      plcs::StringEvaluator::PROCESS_EXE,
      [](const char *policy, const char *ctx, const char *description) {
        std::cout << "\"PROCESS_EXE\" evaluator\n";
        if (policy && ctx && description) {
          std::cout << "  - evaluating: '" << policy << "' with '" << ctx << "'\n"
                    << "  - description: '" << description << "' (id: PROCESS_EXE_PATH)\n";
        }
        return plcs::Result::TRUE;
      }
  );

  plcs::register_action(
      plcs::Action::INJECT_DENY,
      [](plcs::Result result, const std::vector<const char *> &values,
         const char *description) -> std::optional<plcs::Error> {
        std::cout << "INJECT_DENY action\n"
                  << "  - action description: '" << description << "'\n"
                  << "  - evaluation result: " << result << "\n"
                  << "  - values: \n";
        for (auto v : values) {
          std::cout << "    + '" << v << "'\n";
        }
        return std::nullopt;
      }
  );
  plcs::register_action(
      plcs::Action::INJECT_ALLOW,
      [](plcs::Result result, const std::vector<const char *> &values,
         const char *description) -> std::optional<plcs::Error> {
        std::cout << "INJECT_ALLOW action\n"
                  << "  - action description: '" << description << "'\n"
                  << "  - evaluation result: " << result << "\n"
                  << "  - values: \n";
        for (auto v : values) {
          std::cout << "    + '" << v << "'\n";
        }
        return std::nullopt;
      }
  );

  if (auto maybe_errors = plcs::evaluate_buffer_from_file(argv[1]); maybe_errors) {
    std::cout << "Failed to evaluate policy buffer: " << *maybe_errors << "\n";
    return 1;
  }

  return 0;
}
