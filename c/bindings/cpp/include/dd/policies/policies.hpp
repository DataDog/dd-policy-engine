/*
 * C++ bindings for `dd-policy-engine`.
 *
 * This header defines the core public API for initializing and interacting with the Policy Engine.
 * It provides mechanisms to set evaluation parameters, register evaluators and actions,
 * and evaluate policy buffers or files.
 */
#pragma once

extern "C" {
#include <dd/policies/action.h>
#include <dd/policies/error_codes.h>
#include <dd/policies/eval_ctx.h>
#include <dd/policies/evaluation_result.h>
#include <dd/policies/evaluator_types.h>
#include <dd/policies/policies.h>
}

#include <array>
#include <cstdint>
#include <functional>
#include <iostream>
#include <optional>
#include <string_view>
#include <vector>

#ifndef _MSC_VER
#include <fstream>
#else
#include <cstdio>
#endif
#include <array>

#include <filesystem>

namespace datadog::policies {

#define ENUM_VAL(ID, X) ID = X,

enum class Result : int { PLCS_LIST_RESULT(ENUM_VAL) };

enum class StringEvaluator : int { PLCS_LIST_STRING_EVALUATORS(ENUM_VAL) COUNT };

enum class NumericEvaluator : int { PLCS_LIST_NUMERIC_EVALUATORS(ENUM_VAL) COUNT };

enum class Action : int { PLCS_LIST_ACTIONS(ENUM_VAL) COUNT };

enum class Error : int { PLCS_LIST_ERRORS(ENUM_VAL) };

#undef ENUM_VAL

using StringEvaluatorFunc = std::function<Result(const char *, const char *, const char *)>;
using NumericEvaluatorFunc = std::function<Result(const long policy, const long ctx, const char *desc)>;
using ActionFunc = std::function<std::optional<Error>(Result, const std::vector<const char *> &, const char *)>;

namespace {

template <typename T, std::size_t N>
constexpr auto make_empty_array() -> std::array<T, N> {
  return {};  // value-initializes elements
}

// NOTE(@dmehala): Temporary workaround because the API doesn't offer the possibility to store additional context.
auto actions_callback = make_empty_array<ActionFunc, PLCS_ACTIONS__COUNT>();
auto string_evaluators_callback = make_empty_array<StringEvaluatorFunc, PLCS_STR_EVAL__COUNT>();
auto numeric_evaluators_callback = make_empty_array<NumericEvaluatorFunc, PLCS_NUM_EVAL__COUNT>();

plcs_evaluation_result on_str_evaluator(
    const char *policy,
    const plcs_string_comparator,
    const char *ctx,
    const char *desc,
    plcs_string_evaluators id
) {
  if (id > PLCS_STR_EVAL__COUNT)
    return PLCS_EVAL_RESULT_ABSTAIN;

  auto callback = string_evaluators_callback[id];
  if (callback == nullptr)
    return PLCS_EVAL_RESULT_ABSTAIN;

  return static_cast<plcs_evaluation_result>(callback(policy, ctx, desc));
}

plcs_evaluation_result on_numeric_evaluator(
    const long policy,
    const plcs_numeric_comparator,
    const long ctx,
    const char *desc,
    plcs_numeric_evaluators id
) {
  if (id > (plcs_numeric_evaluators)PLCS_NUM_CMP__COUNT)
    return PLCS_EVAL_RESULT_ABSTAIN;

  auto callback = numeric_evaluators_callback[id];
  if (callback == nullptr)
    return PLCS_EVAL_RESULT_ABSTAIN;

  return static_cast<plcs_evaluation_result>(callback(policy, ctx, desc));
}

plcs_errors
on_actions(plcs_evaluation_result result, char *values[], size_t values_len, const char *desc, int action_id) {
  if (action_id > PLCS_ACTIONS__COUNT)
    return PLCS_EUNKNOWN_EVAL_IX;

  auto callback = actions_callback[action_id];
  if (callback == nullptr)
    return PLCS_EUNKNOWN_EVAL_IX;

  std::vector<const char *> v(values, values + values_len);

  auto maybe_error = callback(static_cast<Result>(result), v, desc);
  return maybe_error ? static_cast<plcs_errors>(*maybe_error) : PLCS_ESUCCESS;
}

}  // namespace

inline std::string_view to_string(Result res) {
  switch (res) {
    case Result::TRUE:
      return "true";

    case Result::FALSE:
      return "false";

    case Result::ABSTAIN:
      return "abstain";
  }
  std::abort();
}

inline std::string_view to_string(Error err) {
  switch (err) {
    case Error::SUCCESS:
      return "success";
    case Error::REGISTER_EVAL_PTR:
      return "registered null evaluator callback";
    case Error::IX_OVERFLOW:
      return "ix overflow";
    case Error::NULL_PTR:
      return "nullptr";
    case Error::INITIZLIED:
      return "context already initialized";
    case Error::NO_DATA:
      return "no policies to evaluate";
    case Error::UNKNOWN_EVAL_IX:
      return "unknown evaluator";
    case Error::ACTIONS_EVAL:
      return "missing action callback";
    case Error::UNKNOWN_CMP:
      return "unknown comparator";
    case Error::ALLOCATION:
      return "failed to allocate memory";
    case Error::STR_PARAM_EXCEED_MAX_LENGTH:
      return "string parameter exceeds maximum length";
  }
  std::abort();
}

std::ostream &operator<<(std::ostream &os, Error error) {
  return os << to_string(error);
}

std::ostream &operator<<(std::ostream &os, Result res) {
  return os << to_string(res);
}

/// @brief Initializes the policy engine.
///
/// Must be called before using any other API functions.
/// Sets up internal state and prepares the engine for configuration and evaluation.
inline void init() {
  plcs_eval_ctx_init();
}

/// @brief Sets a numeric evaluation parameter.
///
/// @param evaluator The numeric evaluator key.
/// @param value The `long` value to associate with the evaluator.
inline void set_params(NumericEvaluator evaluator, long value) {
  plcs_eval_ctx_set_num_eval_param(static_cast<plcs_numeric_evaluators>(evaluator), value);
}

/// @brief Sets a numeric evaluation parameter.
///
/// @param evaluator The numeric evaluator key.
/// @param value The `unsigned long` value to associate with the evaluator.
inline void set_params(NumericEvaluator evaluator, unsigned long value) {
  plcs_eval_ctx_set_unum_eval_param(static_cast<plcs_numeric_evaluators>(evaluator), value);
}

/// @brief Sets a string evaluation parameter.
///
/// @param evaluator The string evaluator key.
/// @param value The `std::string_view` to associate with the evaluator.
inline void set_params(StringEvaluator evaluator, std::string_view value) {
  // TODO: Make sure value ends with `\0`.
  plcs_eval_ctx_set_str_eval_param(static_cast<plcs_string_evaluators>(evaluator), value.data());
}

/// @brief Registers a string evaluator function.
///
/// @param evaluator The string evaluator key.
/// @param func The evaluator function to register for the key.
inline void register_evaluator(StringEvaluator id, StringEvaluatorFunc cb) {
  string_evaluators_callback[static_cast<size_t>(id)] = cb;
  plcs_eval_ctx_register_str_evaluator(on_str_evaluator, static_cast<plcs_string_evaluators>(id));
}

/// @brief Registers a numeric evaluator function.
///
/// @param evaluator The numeric evaluator key.
/// @param func The evaluator function to register for the key.
inline void register_evaluator(NumericEvaluator id, NumericEvaluatorFunc cb) {
  numeric_evaluators_callback[static_cast<size_t>(id)] = cb;
  plcs_eval_ctx_register_num_evaluator(on_numeric_evaluator, static_cast<plcs_numeric_evaluators>(id));
}

/// @brief Registers a policy action function.
///
/// @param action The action identifier.
/// @param func The function to invoke when this action is triggered.
inline void register_action(Action action, ActionFunc cb) {
  actions_callback[static_cast<size_t>(action)] = cb;
  plcs_eval_ctx_register_action(on_actions, static_cast<plcs_actions>(action));
}

/// @brief Evaluates a policy from a raw buffer.
///
/// @param buffer A vector of bytes representing the serialized policy.
/// @return An optional `Result`. Returns `std::nullopt` if evaluation fails or is invalid.
inline std::optional<Error> evaluate_buffer(const std::vector<uint8_t> &buffer) {
  auto res = plcs_evaluate_buffer((uint8_t *)buffer.data(), buffer.size());
  return res == PLCS_ESUCCESS ? std::nullopt : std::optional{static_cast<Error>(res)};
}

/// @brief Evaluates a policy from a file.
///
/// @param path The filesystem path to the policy file.
/// @return An optional `Result`. Returns `std::nullopt` if file read or evaluation fails.
#ifdef _MSC_VER
// MSVC is broken ffs.
std::optional<Error> evaluate_buffer_from_file(const std::filesystem::path &filepath) {
  FILE *file = NULL;
  fopen_s(&file, filepath.string().c_str(), "rb");
  if (!file) {
    return Error::NO_DATA;
  }

  fseek(file, 0, SEEK_END);
  const auto file_size = ftell(file);
  fseek(file, 0, SEEK_SET);

  std::vector<uint8_t> buffer(file_size);

  const auto read_size = fread(buffer.data(), 1, file_size, file);
  fclose(file);

  if (read_size != file_size) {
    return Error::NO_DATA;
  }

  return evaluate_buffer(buffer);
}
#else
std::optional<Error> evaluate_buffer_from_file(const std::filesystem::path &filepath) {
  std::ifstream f(filepath, std::ios::binary);
  if (!f.is_open()) {
    return Error::NO_DATA;
  }

  std::vector<uint8_t> buffer(std::istreambuf_iterator<char>(f), {});
  return evaluate_buffer(buffer);
}
#endif

}  // namespace datadog::policies
