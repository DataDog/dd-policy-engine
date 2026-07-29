#!/usr/bin/env bash
# Unless explicitly stated otherwise all files in this repository are licensed
# under the Apache 2.0 License. This product includes software developed at
# Datadog (https://www.datadoghq.com/).
#
# Copyright 2025-Present Datadog, Inc.
#
# Forward compatibility of dd-compile-policy.
#
# An older agent ships an older dd-compile-policy, but remote config may hand it
# a policy JSON produced against a NEWER schema. The old compiler must not
# reject the whole payload just because it sees something it doesn't know about:
# it should drop the unknown parts and compile the rest.
#
# Usage: forward_compat_test.sh [path-to-dd-compile-policy]

set -u -o pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN="${1:-${BIN:-$SCRIPT_DIR/../bin/dd-compile-policy}}"

if [ ! -x "$BIN" ]; then
  echo "[!] dd-compile-policy not found at $BIN (run 'make build' first)" >&2
  exit 1
fi

WORK_DIR="$(mktemp -d)"
trap 'rm -rf "$WORK_DIR"' EXIT

FAILURES=0

pass() { echo "[ok]   $1"; }
fail() {
  echo "[FAIL] $1"
  shift
  if [ $# -gt 0 ]; then echo "       $*"; fi
  FAILURES=$((FAILURES + 1))
}

# A policy exercising every table an RC payload touches: Policies, Policy,
# NodeTypeWrapper, CompositeNode, EvaluatorNode, StrEvaluator, Action.
# The @NAME@ markers are injection points for "a newer schema added a field
# here"; render() strips the ones a case doesn't use.
read -r -d '' TEMPLATE <<'EOF'
{
  @ROOT@
  "policies": [
    {
      @POLICY@
      "description": "forward compat baseline",
      "rules": {
        "node_type": "CompositeNode",
        "node": {
          @COMPOSITE@
          "description": "root",
          "op": "BOOL_AND",
          "children": [
            {
              "node_type": "EvaluatorNode",
              "node": {
                @LEAF@
                "description": "leaf",
                "eval_type": "StrEvaluator",
                "eval": {
                  @EVAL@
                  "id": "OS",
                  "cmp": "CMP_EXACT",
                  "value": "linux"
                }
              }
            }
          ]
        }
      },
      "actions": [
        {
          @ACTION@
          "action": "INJECT_ALLOW",
          "description": "act",
          "values": ["x"]
        }
      ],
      "version": 1
    }
  ]
}
EOF

MARKERS="ROOT POLICY COMPOSITE LEAF EVAL ACTION"

# render [marker injection] -> the template with `injection` (a JSON member and
# its trailing comma) at `marker`, and every other marker removed.
render() {
  local out="$TEMPLATE" marker
  for marker in $MARKERS; do
    if [ "${1:-}" = "$marker" ]; then
      out="${out//@$marker@/$2}"
    else
      out="${out//@$marker@/}"
    fi
  done
  printf '%s' "$out"
}

# compile <name> <json>: writes $WORK_DIR/<name>.bin, echoes compiler diagnostics
compile() {
  printf '%s' "$2" >"$WORK_DIR/$1.json"
  "$BIN" --input-file "$WORK_DIR/$1.json" --output-file "$WORK_DIR/$1.bin" 2>&1
}

# same_bytes <file> <file>: exact byte comparison of two compiled buffers.
# Uses od rather than cmp because od is coreutils and so is present in
# Git-for-Windows bash, where diffutils (cmp) is not guaranteed. -v keeps od from
# collapsing repeated lines into '*', which would hide a difference.
same_bytes() {
  [ "$(od -An -tx1 -v <"$1")" = "$(od -An -tx1 -v <"$2")" ]
}

echo "[i] dd-compile-policy forward compatibility ($BIN)"

# --- baseline -----------------------------------------------------------------
if out="$(compile base "$(render)")"; then
  pass "baseline policy compiles"
else
  fail "baseline policy compiles" "$out"
  echo "[!] baseline is broken, aborting" >&2
  exit 1
fi

# --- unknown fields must be ignored ------------------------------------------
# The compiler must succeed AND produce a buffer byte-identical to the baseline:
# identical output is what proves the field was dropped rather than smuggled
# into the binary that the engine later reads.
assert_field_ignored() {
  local name="$1" marker="$2" injection="$3"
  if ! out="$(compile "$name" "$(render "$marker" "$injection")")"; then
    fail "$name: compiles" "$out"
    return
  fi
  if same_bytes "$WORK_DIR/base.bin" "$WORK_DIR/$name.bin"; then
    pass "$name: ignored, output identical to baseline"
  else
    fail "$name: output differs from baseline" "the unknown field leaked into the buffer"
  fi
}

assert_field_ignored unknown-field-in-policies-root ROOT '"future_root_field": 42,'
assert_field_ignored unknown-field-in-policy POLICY '"future_policy_field": "hello",'
assert_field_ignored unknown-field-in-composite-node COMPOSITE '"future_node_field": true,'
assert_field_ignored unknown-field-in-evaluator-node LEAF '"future_leaf_field": [1, 2],'
assert_field_ignored unknown-field-in-str-evaluator EVAL '"future_eval_field": { "nested": { "deep": [null] } },'
assert_field_ignored unknown-field-in-action ACTION '"future_action_field": null,'

# --- unknown enum values must degrade to the *_UNKNOWN default ----------------
# The engine treats *_UNKNOWN as "abstain", so falling back to the schema
# default is the correct degradation for an evaluator/action we cannot run.
# Proven by compiling the same policy with the enum spelled *_UNKNOWN
# explicitly: byte-identical output means the future name landed on that value.
assert_unknown_enum_degrades() {
  local name="$1" known="$2" future="$3" fallback="$4" json
  json="$(render)"
  if ! out="$(compile "$name" "${json//$known/$future}")"; then
    fail "$name: compiles" "$out"
    return
  fi
  if ! out="$(compile "$name-ref" "${json//$known/$fallback}")"; then
    fail "$name: reference policy with $fallback compiles" "$out"
    return
  fi
  if same_bytes "$WORK_DIR/$name.bin" "$WORK_DIR/$name-ref.bin"; then
    pass "$name: degrades to $fallback"
  else
    fail "$name: did not degrade to $fallback" "the unknown enum value produced something else"
  fi
}

assert_unknown_enum_degrades unknown-enum-action-id \
  '"INJECT_ALLOW"' '"ACTION_FROM_THE_FUTURE"' '"ACTION_UNKNOWN"'
assert_unknown_enum_degrades unknown-enum-cmp-str \
  '"CMP_EXACT"' '"CMP_FROM_THE_FUTURE"' '"CMP_STR_UNKNOWN"'
assert_unknown_enum_degrades unknown-enum-string-evaluator \
  '"id": "OS"' '"id": "STRING_EVAL_FROM_THE_FUTURE"' '"id": "STRING_EVAL_UNKNOWN"'
assert_unknown_enum_degrades unknown-enum-bool-operation \
  '"BOOL_AND"' '"BOOL_FROM_THE_FUTURE"' '"BOOL_UNKNOWN"'

# --- what is still NOT forward compatible ------------------------------------
assert_rejected() {
  local name="$1" json="$2" note="${3:-}"
  if out="$(compile "$name" "$json")"; then
    fail "$name: expected rejection, but it compiled" "$note"
  else
    pass "$name: rejected as expected"
  fi
}

# Unknown union variants cannot be skipped: FlatBuffers must know the variant to
# know how to read the accompanying table. This pins the limitation, so a newer
# schema must not put a new NodeType/EvaluatorType variant in a policy that
# older agents are expected to load. If unions ever do become forward
# compatible, move these two cases to the tolerated set above.
json="$(render)"
assert_rejected unknown-union-variant-node-type \
  "${json//'"node_type": "EvaluatorNode"'/'"node_type": "FutureNode"'}" \
  "unions became forward compatible - move this case to the tolerated set"
assert_rejected unknown-union-variant-evaluator-type \
  "${json//'"eval_type": "StrEvaluator"'/'"eval_type": "FutureEvaluator"'}" \
  "unions became forward compatible - move this case to the tolerated set"

# --- skipping unknown fields must not disable real validation ----------------
assert_rejected malformed-json '{"policies": [ '
assert_rejected wrong-type-for-known-field \
  "${json//'"version": 1'/'"version": "not-a-number"'}"

# --- summary ------------------------------------------------------------------
if [ "$FAILURES" -ne 0 ]; then
  echo "[!] $FAILURES forward compatibility check(s) failed" >&2
  exit 1
fi
echo "[i] all forward compatibility checks passed"
