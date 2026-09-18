#!/usr/bin/env python3
# Unless explicitly stated otherwise all files in this repository are licensed
# under the Apache 2.0 License. This product includes software developed at
# Datadog (https://www.datadoghq.com/).
#
# Copyright 2025-Present Datadog, Inc.
"""Injector-package schema fallback for dd-compile-policy.

The Agent pins dd-compile-policy to a specific version, so its baked-in schema
can lag newly added evaluators. datadog-apm-inject updates on a separate
cadence and can drop a fresher schema at a fixed, platform-specific path; the
compiler should prefer it over the one baked into the binary, but an explicit
--schema-file must still win over both.

Usage: injector_schema_fallback_test.py [path-to-dd-compile-policy]
"""

import json
import os
import shutil
import subprocess
import sys
import tempfile

# Mirrors kInjectorSchemaPath in src/compile_policy.cpp.
if os.name == "nt":
    INJECTOR_SCHEMA_PATH = (
        r"C:\ProgramData\Datadog\Installer\packages\datadog-apm-inject\stable\dll\policy.bfbs"
    )
else:
    INJECTOR_SCHEMA_PATH = "/opt/datadog-packages/datadog-apm-inject/stable/inject/policy.bfbs"

BASELINE = {
    "policies": [
        {
            "description": "injector schema fallback baseline",
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


class Runner:
    def __init__(self, binary, work_dir):
        self.binary = binary
        self.work_dir = work_dir
        self.failures = 0

    def ok(self, name, detail):
        print("[ok]   {}: {}".format(name, detail))

    def fail(self, name, detail):
        print("[FAIL] {}".format(name))
        if detail:
            print("       {}".format(detail.strip()))
        self.failures += 1

    def compile(self, name, schema_file=None):
        src = os.path.join(self.work_dir, name + ".json")
        out = os.path.join(self.work_dir, name + ".bin")
        with open(src, "w") as handle:
            handle.write(json.dumps(BASELINE))
        cmd = [self.binary, "--input-file", src, "--output-file", out]
        if schema_file:
            cmd += ["--schema-file", schema_file]
        proc = subprocess.run(
            cmd, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, universal_newlines=True
        )
        return proc.returncode == 0, proc.stdout or "", out

    def assert_compiles(self, name, schema_file, detail):
        ok, diagnostics, _ = self.compile(name, schema_file)
        if ok:
            self.ok(name, detail)
        else:
            self.fail(name, diagnostics)

    def assert_fails(self, name, schema_file, detail):
        ok, diagnostics, _ = self.compile(name, schema_file)
        if ok:
            self.fail(name + ": expected failure, but it compiled", detail)
        else:
            self.ok(name, detail)


def default_binary():
    name = "dd-compile-policy.exe" if os.name == "nt" else "dd-compile-policy"
    here = os.path.dirname(os.path.abspath(__file__))
    return os.path.join(here, os.pardir, "bin", name)


def default_builtin_schema():
    here = os.path.dirname(os.path.abspath(__file__))
    return os.path.join(here, os.pardir, "schema", "policy.bfbs")


def main(argv):
    binary = argv[1] if len(argv) > 1 else os.environ.get("BIN") or default_binary()
    if not os.path.isfile(binary):
        sys.stderr.write(
            "[!] dd-compile-policy not found at {} (run 'make build' first)\n".format(binary)
        )
        return 1

    if os.path.exists(INJECTOR_SCHEMA_PATH):
        print(
            "[i] skipping: {} already exists (looks like a real install, "
            "refusing to touch it)".format(INJECTOR_SCHEMA_PATH)
        )
        return 0

    injector_dir = os.path.dirname(INJECTOR_SCHEMA_PATH)
    try:
        os.makedirs(injector_dir, exist_ok=True)
        with open(INJECTOR_SCHEMA_PATH, "wb") as handle:
            handle.write(b"")
    except OSError as exc:
        print(
            "[i] skipping: cannot write {} ({}); rerun as root/in the build "
            "container to exercise this tier".format(INJECTOR_SCHEMA_PATH, exc)
        )
        try:
            os.remove(INJECTOR_SCHEMA_PATH)
        except OSError:
            pass
        return 0

    good_schema = default_builtin_schema()
    if not os.path.isfile(good_schema):
        sys.stderr.write(
            "[!] built schema not found at {} (run 'make build' first)\n".format(good_schema)
        )
        os.remove(INJECTOR_SCHEMA_PATH)
        return 1

    work_dir = tempfile.mkdtemp(prefix="dd-injector-schema-")
    try:
        run = Runner(binary, work_dir)
        print("[i] dd-compile-policy injector schema fallback ({})".format(binary))

        # An invalid (but present and readable) file at the injector path must
        # be preferred over the baked-in schema when no --schema-file is
        # given: the compile fails, proving the compiler actually tried to
        # deserialize it rather than silently falling back.
        with open(INJECTOR_SCHEMA_PATH, "wb") as handle:
            handle.write(b"not a valid flatbuffers binary schema")
        run.assert_fails(
            "invalid-injector-schema-used-over-builtin",
            None,
            "corrupt injector schema rejected instead of silently using the built-in one",
        )

        # An explicit --schema-file must still win over the injector path,
        # even when the injector path holds something invalid.
        run.assert_compiles(
            "explicit-schema-file-wins-over-injector-path",
            good_schema,
            "explicit --schema-file bypassed the corrupt injector schema",
        )

        # Once the injector path is empty/removed, the baked-in schema is
        # used again (tier 3, unchanged default behavior).
        os.remove(INJECTOR_SCHEMA_PATH)
        run.assert_compiles(
            "falls-back-to-builtin-when-injector-path-absent",
            None,
            "compiled using the schema baked into the binary",
        )

        if run.failures:
            sys.stderr.write(
                "[!] {} injector schema fallback check(s) failed\n".format(run.failures)
            )
            return 1
        print("[i] all injector schema fallback checks passed")
        return 0
    finally:
        shutil.rmtree(work_dir, ignore_errors=True)
        try:
            os.remove(INJECTOR_SCHEMA_PATH)
        except OSError:
            pass


if __name__ == "__main__":
    sys.exit(main(sys.argv))
