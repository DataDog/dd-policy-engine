#!/usr/bin/env python3
# Unless explicitly stated otherwise all files in this repository are licensed
# under the Apache 2.0 License. This product includes software developed at
# Datadog (https://www.datadoghq.com/).
#
# Copyright 2025-Present Datadog, Inc.
"""Forward compatibility of dd-compile-policy.

An older agent ships an older dd-compile-policy, but remote config may hand it a
policy JSON produced against a NEWER schema. The old compiler must not reject the
whole payload just because it sees something it doesn't know about: it should drop
the unknown parts and compile the rest.

Python rather than shell so the same checks run on Linux, macOS and Windows
without depending on a POSIX shell or on which coreutils Git for Windows ships.

Usage: forward_compat_test.py [path-to-dd-compile-policy]
"""

import copy
import json
import os
import shutil
import subprocess
import sys
import tempfile

# A policy touching every table an RC payload reaches: Policies, Policy,
# NodeTypeWrapper, CompositeNode, EvaluatorNode, StrEvaluator, Action.
BASELINE = {
    "policies": [
        {
            "description": "forward compat baseline",
            "rules": {
                "node_type": "CompositeNode",
                "node": {
                    "description": "root",
                    "op": "BOOL_AND",
                    "children": [
                        {
                            "node_type": "EvaluatorNode",
                            "node": {
                                "description": "leaf",
                                "eval_type": "StrEvaluator",
                                "eval": {
                                    "id": "OS",
                                    "cmp": "CMP_EXACT",
                                    "value": "linux",
                                },
                            },
                        }
                    ],
                },
            },
            "actions": [
                {"action": "INJECT_ALLOW", "description": "act", "values": ["x"]}
            ],
            "version": 1,
        }
    ]
}


# Accessors for the tables above, so a case can say where it is injecting without
# repeating the nesting. Each returns a dict inside the document passed in.
def ROOT(doc):
    return doc


def POLICY(doc):
    return doc["policies"][0]


def COMPOSITE(doc):
    return POLICY(doc)["rules"]["node"]


def WRAPPER(doc):
    return COMPOSITE(doc)["children"][0]


def LEAF(doc):
    return WRAPPER(doc)["node"]


def EVAL(doc):
    return LEAF(doc)["eval"]


def ACTION(doc):
    return POLICY(doc)["actions"][0]


class Runner:
    def __init__(self, binary, work_dir):
        self.binary = binary
        self.work_dir = work_dir
        self.failures = 0
        self.baseline_bin = None

    def ok(self, name, detail):
        print("[ok]   {}: {}".format(name, detail))

    def fail(self, name, detail):
        print("[FAIL] {}".format(name))
        if detail:
            print("       {}".format(detail.strip()))
        self.failures += 1

    def compile(self, name, doc):
        """Compile a policy. `doc` is a dict, or a str for deliberately bad JSON.

        Returns (ok, diagnostics, output_path).
        """
        src = os.path.join(self.work_dir, name + ".json")
        out = os.path.join(self.work_dir, name + ".bin")
        text = doc if isinstance(doc, str) else json.dumps(doc)
        with open(src, "w") as handle:
            handle.write(text)
        proc = subprocess.run(
            [self.binary, "--input-file", src, "--output-file", out],
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            universal_newlines=True,
        )
        diagnostics = (proc.stdout or "").replace("Failed to parse JSON: ", "")
        return proc.returncode == 0, diagnostics, out

    # --- unknown fields ----------------------------------------------------
    def assert_field_ignored(self, name, table, member, value):
        """A field a newer schema might add must be dropped.

        Asserting a byte-identical buffer, not just a zero exit status: identical
        output is what proves the field was dropped rather than smuggled into the
        binary the engine later reads. The member is appended, so the known
        fields keep their relative order and the comparison stays meaningful.
        """
        doc = copy.deepcopy(BASELINE)
        table(doc)[member] = value
        ok, diagnostics, out = self.compile(name, doc)
        if not ok:
            self.fail(name + ": compiles", diagnostics)
            return
        if read(out) == read(self.baseline_bin):
            self.ok(name, "ignored, output identical to baseline")
        else:
            self.fail(name, "the unknown field leaked into the buffer")

    # --- unknown enum values ----------------------------------------------
    def assert_enum_degrades(self, name, table, member, future, fallback):
        """An unknown enum value must land on the schema default.

        The engines treat *_UNKNOWN as "abstain", so that is the correct
        degradation for an evaluator or action we cannot run. Proven by compiling
        the same policy with the enum spelled *_UNKNOWN explicitly: a
        byte-identical buffer means the future name reached that value.
        """
        doc = copy.deepcopy(BASELINE)
        table(doc)[member] = future
        ok, diagnostics, out = self.compile(name, doc)
        if not ok:
            self.fail(name + ": compiles", diagnostics)
            return

        reference = copy.deepcopy(BASELINE)
        table(reference)[member] = fallback
        ok, diagnostics, ref_out = self.compile(name + "-ref", reference)
        if not ok:
            self.fail(
                "{}: reference policy with {} compiles".format(name, fallback),
                diagnostics,
            )
            return

        if read(out) == read(ref_out):
            self.ok(name, "degrades to {}".format(fallback))
        else:
            self.fail(name, "the unknown enum value produced something else")

    # --- must still be rejected -------------------------------------------
    def assert_rejected(self, name, doc, note=""):
        ok, _, _ = self.compile(name, doc)
        if ok:
            self.fail(name + ": expected rejection, but it compiled", note)
        else:
            self.ok(name, "rejected as expected")


def read(path):
    with open(path, "rb") as handle:
        return handle.read()


def default_binary():
    name = "dd-compile-policy.exe" if os.name == "nt" else "dd-compile-policy"
    here = os.path.dirname(os.path.abspath(__file__))
    return os.path.join(here, os.pardir, "bin", name)


def main(argv):
    binary = argv[1] if len(argv) > 1 else os.environ.get("BIN") or default_binary()
    if not os.path.isfile(binary):
        sys.stderr.write(
            "[!] dd-compile-policy not found at {} (run 'make build' first)\n".format(
                binary
            )
        )
        return 1

    work_dir = tempfile.mkdtemp(prefix="dd-fwdcompat-")
    try:
        run = Runner(binary, work_dir)
        print("[i] dd-compile-policy forward compatibility ({})".format(binary))

        # --- baseline ------------------------------------------------------
        ok, diagnostics, run.baseline_bin = run.compile("base", BASELINE)
        if not ok:
            run.fail("baseline policy compiles", diagnostics)
            sys.stderr.write("[!] baseline is broken, aborting\n")
            return 1
        run.ok("baseline policy compiles", "ok")

        # --- unknown fields must be ignored --------------------------------
        run.assert_field_ignored(
            "unknown-field-in-policies-root", ROOT, "future_root_field", 42
        )
        run.assert_field_ignored(
            "unknown-field-in-policy", POLICY, "future_policy_field", "hello"
        )
        run.assert_field_ignored(
            "unknown-field-in-composite-node", COMPOSITE, "future_node_field", True
        )
        run.assert_field_ignored(
            "unknown-field-in-node-wrapper", WRAPPER, "future_wrapper_field", 1.5
        )
        run.assert_field_ignored(
            "unknown-field-in-evaluator-node", LEAF, "future_leaf_field", [1, 2]
        )
        run.assert_field_ignored(
            "unknown-field-in-str-evaluator",
            EVAL,
            "future_eval_field",
            {"nested": {"deep": [None]}},
        )
        run.assert_field_ignored(
            "unknown-field-in-action", ACTION, "future_action_field", None
        )

        # --- unknown enum values must degrade to *_UNKNOWN -----------------
        run.assert_enum_degrades(
            "unknown-enum-action-id",
            ACTION,
            "action",
            "ACTION_FROM_THE_FUTURE",
            "ACTION_UNKNOWN",
        )
        run.assert_enum_degrades(
            "unknown-enum-cmp-str",
            EVAL,
            "cmp",
            "CMP_FROM_THE_FUTURE",
            "CMP_STR_UNKNOWN",
        )
        run.assert_enum_degrades(
            "unknown-enum-string-evaluator",
            EVAL,
            "id",
            "STRING_EVAL_FROM_THE_FUTURE",
            "STRING_EVAL_UNKNOWN",
        )
        run.assert_enum_degrades(
            "unknown-enum-bool-operation",
            COMPOSITE,
            "op",
            "BOOL_FROM_THE_FUTURE",
            "BOOL_UNKNOWN",
        )

        # --- what is still NOT forward compatible --------------------------
        # Unknown union variants cannot be skipped: FlatBuffers must know the
        # variant to know how to read the accompanying table. This pins the
        # limitation, so a newer schema must not put a new NodeType or
        # EvaluatorType into a policy older agents are expected to load. If
        # unions ever do become forward compatible, move these two cases up to
        # the tolerated set.
        moved = "unions became forward compatible - move this case to the tolerated set"
        doc = copy.deepcopy(BASELINE)
        WRAPPER(doc)["node_type"] = "FutureNode"
        run.assert_rejected("unknown-union-variant-node-type", doc, moved)

        doc = copy.deepcopy(BASELINE)
        LEAF(doc)["eval_type"] = "FutureEvaluator"
        run.assert_rejected("unknown-union-variant-evaluator-type", doc, moved)

        # --- skipping unknown fields must not disable real validation ------
        run.assert_rejected("malformed-json", '{"policies": [ ')

        doc = copy.deepcopy(BASELINE)
        POLICY(doc)["version"] = "not-a-number"
        run.assert_rejected("wrong-type-for-known-field", doc)

        # --- summary -------------------------------------------------------
        if run.failures:
            sys.stderr.write(
                "[!] {} forward compatibility check(s) failed\n".format(run.failures)
            )
            return 1
        print("[i] all forward compatibility checks passed")
        return 0
    finally:
        shutil.rmtree(work_dir, ignore_errors=True)


if __name__ == "__main__":
    sys.exit(main(sys.argv))
