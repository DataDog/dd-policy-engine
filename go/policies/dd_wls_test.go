// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache 2.0 License. This product includes software developed at
// Datadog (https://www.datadoghq.com/).
//
// Copyright 2025-Present Datadog, Inc.

package policies

import "testing"

func mustParse(t *testing.T, raw string) []Policy {
	t.Helper()
	ps, err := ParsePolicies([]byte(raw))
	if err != nil {
		t.Fatalf("ParsePolicies: %v", err)
	}
	return ps
}

// matches reports whether a single policy's rule evaluates to true for ctx.
// Combining the outcomes of several matching policies is the consumer's
// responsibility (see doc.go), so these tests check one policy at a time.
func matches(p Policy, ctx Context) bool {
	return Evaluate(p.Rules, ctx) == ResultTrue
}

func TestParsePodLabelAndActions(t *testing.T) {
	raw := `{
      "policies": [{
        "description": "java for db-user",
        "rules": {
          "node_type": "EvaluatorNode",
          "node": {
            "eval_type": "StrEvaluator",
            "eval": {"id": "POD_LABEL", "cmp": "CMP_EXACT", "value": "app=db-user"}
          }
        },
        "actions": [
          {"action": "INJECT_ALLOW"},
          {"action": "ENABLE_SDK", "values": ["java=latest"]}
        ]
      }]
    }`
	ps := mustParse(t, raw)
	if len(ps) != 1 || ps[0].Name != "java for db-user" {
		t.Fatalf("unexpected parse: %+v", ps)
	}

	if !matches(ps[0], Context{Labels: map[string]map[string]string{IDPodLabel: {"app": "db-user"}}}) {
		t.Fatal("db-user pod should match")
	}
	if out := ps[0].Outcome; !out.Inject || out.TracerVersions["java"] != "latest" {
		t.Fatalf("unexpected outcome: %+v", out)
	}
	if matches(ps[0], Context{Labels: map[string]map[string]string{IDPodLabel: {"app": "other"}}}) {
		t.Fatal("non-matching pod should not match")
	}
}

func TestParseInjectDeny(t *testing.T) {
	raw := `{
      "policies": [{
        "description": "deny app=legacy",
        "rules": {"node_type": "EvaluatorNode", "node": {"eval_type": "StrEvaluator",
          "eval": {"id": "POD_LABEL", "cmp": "CMP_EXACT", "value": "app=legacy"}}},
        "actions": [{"action": "INJECT_DENY"}]
      }]
    }`
	ps := mustParse(t, raw)
	if !matches(ps[0], Context{Labels: map[string]map[string]string{IDPodLabel: {"app": "legacy"}}}) {
		t.Fatalf("policy should match")
	}
	if out := ps[0].Outcome; !out.InjectSet || out.Inject {
		t.Fatalf("matched deny policy must carry an explicit no-inject decision: %+v", out)
	}
}

func TestParseExistenceAndNot(t *testing.T) {
	// tier Exists (CMP_PREFIX "tier=") AND NOT deprecated Exists
	raw := `{
      "policies": [{
        "description": "exists",
        "rules": {
          "node_type": "CompositeNode",
          "node": {
            "op": "BOOL_AND",
            "children": [
              {"node_type": "EvaluatorNode", "node": {"eval_type": "StrEvaluator",
                "eval": {"id": "POD_LABEL", "cmp": "CMP_PREFIX", "value": "tier="}}},
              {"node_type": "CompositeNode", "node": {"op": "BOOL_NOT", "children": [
                {"node_type": "EvaluatorNode", "node": {"eval_type": "StrEvaluator",
                  "eval": {"id": "POD_LABEL", "cmp": "CMP_PREFIX", "value": "deprecated="}}}
              ]}}
            ]
          }
        },
        "actions": [{"action": "INJECT_ALLOW"}]
      }]
    }`
	ps := mustParse(t, raw)

	if !matches(ps[0], Context{Labels: map[string]map[string]string{IDPodLabel: {"tier": "frontend"}}}) {
		t.Errorf("tier present, deprecated absent should match")
	}
	if matches(ps[0], Context{Labels: map[string]map[string]string{IDPodLabel: {}}}) {
		t.Errorf("tier absent should not match")
	}
	if matches(ps[0], Context{Labels: map[string]map[string]string{IDPodLabel: {"tier": "x", "deprecated": "true"}}}) {
		t.Errorf("deprecated present should not match")
	}
}

func TestParseWildcardPodLabel(t *testing.T) {
	raw := `{
      "policies": [{
        "description": "wildcard service name",
        "rules": {"node_type": "EvaluatorNode", "node": {"eval_type": "StrEvaluator",
          "eval": {"id": "POD_LABEL", "cmp": "CMP_WILDCARD", "value": "app=k8s-*-svc"}}},
        "actions": [{"action": "INJECT_ALLOW"}]
      }]
    }`
	ps := mustParse(t, raw)
	if !matches(ps[0], Context{Labels: map[string]map[string]string{IDPodLabel: {"app": "k8s-payments-svc"}}}) {
		t.Errorf("wildcard label should match k8s-payments-svc")
	}
	if matches(ps[0], Context{Labels: map[string]map[string]string{IDPodLabel: {"app": "other"}}}) {
		t.Errorf("wildcard label should not match other")
	}
}

func TestParseNamespaceNameAndDefault(t *testing.T) {
	raw := `{
      "policies": [
        {
          "description": "ns names",
          "rules": {"node_type": "CompositeNode", "node": {"op": "BOOL_OR", "children": [
            {"node_type": "EvaluatorNode", "node": {"eval_type": "StrEvaluator",
              "eval": {"id": "NAMESPACE_NAME", "cmp": "CMP_EXACT", "value": "payments"}}},
            {"node_type": "EvaluatorNode", "node": {"eval_type": "StrEvaluator",
              "eval": {"id": "NAMESPACE_NAME", "cmp": "CMP_EXACT", "value": "billing"}}}
          ]}},
          "actions": [{"action": "INJECT_ALLOW"}, {"action": "ENABLE_SDK", "values": ["java=latest"]}]
        },
        {
          "description": "default",
          "rules": {"node_type": "EvaluatorNode", "node": {"eval_type": "StrEvaluator",
            "eval": {"id": "ALWAYS_TRUE", "cmp": "CMP_EXACT", "value": ""}}},
          "actions": [{"action": "INJECT_ALLOW"}]
        }
      ]
    }`
	ps := mustParse(t, raw)
	if len(ps) != 2 {
		t.Fatalf("expected 2 policies, got %d", len(ps))
	}

	// billing matches the first (namespace) policy, which enables the java tracer;
	// it does not match the catch-all's job here — that's the consumer's fold.
	billing := Context{Strings: map[string]string{IDNamespaceName: "billing"}}
	if !matches(ps[0], billing) {
		t.Fatal("billing should match the namespace policy")
	}
	if ps[0].Outcome.TracerVersions["java"] != "latest" {
		t.Fatalf("namespace policy should enable java=latest: %+v", ps[0].Outcome)
	}

	// default matches only the catch-all, which injects with no tracer versions.
	def := Context{Strings: map[string]string{IDNamespaceName: "default"}}
	if matches(ps[0], def) {
		t.Fatal("default should not match the namespace policy")
	}
	if !matches(ps[1], def) {
		t.Fatal("default should match the catch-all policy")
	}
	if out := ps[1].Outcome; !out.Inject || len(out.TracerVersions) != 0 {
		t.Fatalf("catch-all should inject with no tracer versions: %+v", out)
	}
}

func TestParseUUID(t *testing.T) {
	raw := `{
      "policies": [{
        "description": "with id",
        "id": {"hi": 10, "lo": 11},
        "version": 7,
        "rules": {"node_type": "EvaluatorNode", "node": {"eval_type": "StrEvaluator",
          "eval": {"id": "ALWAYS_TRUE", "cmp": "CMP_EXACT", "value": ""}}},
        "actions": [{"action": "INJECT_ALLOW"}]
      }]
    }`
	ps := mustParse(t, raw)
	if ps[0].ID != "00000000-0000-000a-0000-00000000000b" {
		t.Errorf("unexpected UUID: %q", ps[0].ID)
	}
	if ps[0].Version != 7 {
		t.Errorf("unexpected version: %d", ps[0].Version)
	}

	// No id field => empty ID.
	noID := mustParse(t, `{"policies":[{"description":"x","rules":{"node_type":"EvaluatorNode","node":{"eval_type":"StrEvaluator","eval":{"id":"ALWAYS_TRUE"}}},"actions":[{"action":"INJECT_ALLOW"}]}]}`)
	if noID[0].ID != "" {
		t.Errorf("expected empty ID, got %q", noID[0].ID)
	}
}

// TestEnableProfilerDeduped checks that a single policy decodes
// DD_PROFILING_ENABLED at most once even when it lists ENABLE_PROFILER more than
// once. Deduping across several matching policies is the consumer's fold and
// lives outside this engine.
func TestEnableProfilerDeduped(t *testing.T) {
	countEnv := func(configs []EnvVar, name string) int {
		n := 0
		for _, c := range configs {
			if c.Name == name {
				n++
			}
		}
		return n
	}

	ps := mustParse(t, `{"policies":[{"rules":{"node_type":"EvaluatorNode","node":{"eval_type":"StrEvaluator","eval":{"id":"ALWAYS_TRUE"}}},"actions":[{"action":"ENABLE_PROFILER"},{"action":"ENABLE_PROFILER"}]}]}`)
	if n := countEnv(ps[0].Outcome.TracerConfigs, "DD_PROFILING_ENABLED"); n != 1 {
		t.Fatalf("profiler enabled twice in one policy: got %d DD_PROFILING_ENABLED entries, want 1", n)
	}
}

// TestDecodeActionsIgnoresUnhandledActionIDs documents that an action the engine
// does not implement (SET_ENVAR, REEXEC, or a future ActionId) is ignored rather
// than failing the parse, mirroring the C engine and keeping a newer policy
// forward-compatible with an older evaluator. Actions the engine does handle in
// the same policy still apply.
func TestDecodeActionsIgnoresUnhandledActionIDs(t *testing.T) {
	// REEXEC exists in the ActionId schema but the engine does not implement it;
	// it is ignored (forward compatibility) rather than failing the parse, and a
	// handled action in the same policy still applies.
	raw := `{
      "policies": [{
        "description": "handled + unhandled actions",
        "rules": {"node_type": "EvaluatorNode", "node": {"eval_type": "StrEvaluator",
          "eval": {"id": "ALWAYS_TRUE"}}},
        "actions": [
          {"action": "INJECT_ALLOW"},
          {"action": "REEXEC"}
        ]
      }]
    }`
	ps := mustParse(t, raw)
	if out := ps[0].Outcome; !out.InjectSet || !out.Inject {
		t.Fatalf("handled action must still apply alongside an ignored one: %+v", out)
	}
}

// TestDecodeSetEnvVar checks SET_ENVAR values are parsed as NAME=value entries
// into Outcome.TracerConfigs (the same place ENABLE_PROFILER writes env vars),
// splitting on the first '=' so values may themselves contain '='. Malformed
// entries are rejected, like ENABLE_SDK.
func TestDecodeSetEnvVar(t *testing.T) {
	ps := mustParse(t, `{"policies":[{"rules":{"node_type":"EvaluatorNode","node":{"eval_type":"StrEvaluator","eval":{"id":"ALWAYS_TRUE"}}},"actions":[{"action":"SET_ENVAR","values":["DD_FOO=bar","DD_BAZ=a=b"]}]}]}`)
	got := map[string]string{}
	for _, c := range ps[0].Outcome.TracerConfigs {
		got[c.Name] = c.Value
	}
	if got["DD_FOO"] != "bar" || got["DD_BAZ"] != "a=b" {
		t.Fatalf("SET_ENVAR not parsed into TracerConfigs: %+v", ps[0].Outcome.TracerConfigs)
	}

	// Missing '=' is rejected.
	if _, err := ParsePolicies([]byte(`{"policies":[{"rules":{"node_type":"EvaluatorNode","node":{"eval_type":"StrEvaluator","eval":{"id":"ALWAYS_TRUE"}}},"actions":[{"action":"SET_ENVAR","values":["DD_FOO"]}]}]}`)); err == nil {
		t.Fatal("expected error for SET_ENVAR value without '='")
	}
}

// TestParseAllowsExplicitEmptyStringValue checks that an explicit empty value is
// legal (distinct from an omitted one, which is rejected): it decodes to a real
// exact-empty comparison rather than a broad match.
func TestParseAllowsExplicitEmptyStringValue(t *testing.T) {
	ps := mustParse(t, `{"policies":[{"rules":{"node_type":"EvaluatorNode","node":{"eval_type":"StrEvaluator","eval":{"id":"NAMESPACE_NAME","cmp":"CMP_EXACT","value":""}}}}]}`)
	if !matches(ps[0], Context{Strings: map[string]string{IDNamespaceName: ""}}) {
		t.Error("exact-empty should match an empty namespace name")
	}
	if matches(ps[0], Context{Strings: map[string]string{IDNamespaceName: "x"}}) {
		t.Error("exact-empty should not match a non-empty namespace name")
	}
}

// TestBackwardCompatDeepNestingDoesNotCrash checks that a config far deeper than
// maxEvalDepth is parsed without overflowing the stack: the decoder caps its
// recursion at maxEvalDepth and abstains beyond it (the same result eval would
// give), rather than relying on encoding/json's internal nesting limit.
func TestBackwardCompatDeepNestingDoesNotCrash(t *testing.T) {
	nested := func(depth int) string {
		s := `{"node_type":"EvaluatorNode","node":{"eval_type":"StrEvaluator","eval":{"id":"ALWAYS_TRUE"}}}`
		for range depth {
			s = `{"node_type":"CompositeNode","node":{"op":"BOOL_NOT","children":[` + s + `]}}`
		}
		return `{"policies":[{"rules":` + s + `,"actions":[]}]}`
	}
	ps := mustParse(t, nested(maxEvalDepth*20))
	if got := Evaluate(ps[0].Rules, Context{}); got != ResultAbstain {
		t.Fatalf("deep tree should abstain past the depth limit, got %v", got)
	}
}

// TestBackwardCompatIgnoresUnknownFields checks that fields a newer schema may
// add (at the document, policy, node, evaluator, or action level) are ignored
// rather than failing the parse, so an older agent keeps working on a newer
// config.
func TestBackwardCompatIgnoresUnknownFields(t *testing.T) {
	raw := `{"new_top":1,"policies":[{"description":"x","new_pol":true,
	  "rules":{"node_type":"EvaluatorNode","new_node":9,"node":{"eval_type":"StrEvaluator","new_e":2,"eval":{"id":"NAMESPACE_NAME","cmp":"CMP_EXACT","value":"p","new_leaf":5}}},
	  "actions":[{"action":"INJECT_ALLOW","new_act":7}]}]}`
	ps := mustParse(t, raw)
	if got := Evaluate(ps[0].Rules, Context{Strings: map[string]string{IDNamespaceName: "p"}}); got != ResultTrue {
		t.Fatalf("unknown fields should be ignored; want TRUE, got %v", got)
	}
	if !ps[0].Outcome.Inject {
		t.Fatalf("INJECT_ALLOW should still apply alongside unknown action fields")
	}
}

// TestParseProcessEnvVar checks PROCESS_ENVAR is decoded as a keyed KEY=VALUE
// evaluator (like a label), so several env-var conditions AND'd together -- as
// the requirements converter's deny rules emit -- resolve against independent
// keys in Context.Labels. A single plain-string fact could not satisfy two
// different variables, and callers supplying env vars via Context.Labels would
// otherwise abstain.
func TestParseProcessEnvVar(t *testing.T) {
	// deny.go emits AND(PROCESS_ENVAR "FOO=1", PROCESS_ENVAR "BAR=2").
	raw := `{"policies":[{"description":"deny FOO=1 and BAR=2","rules":{"node_type":"CompositeNode","node":{"op":"BOOL_AND","children":[
	  {"node_type":"EvaluatorNode","node":{"eval_type":"StrEvaluator","eval":{"id":"PROCESS_ENVAR","cmp":"CMP_EXACT","value":"FOO=1"}}},
	  {"node_type":"EvaluatorNode","node":{"eval_type":"StrEvaluator","eval":{"id":"PROCESS_ENVAR","cmp":"CMP_EXACT","value":"BAR=2"}}}
	]}},"actions":[{"action":"INJECT_DENY"}]}]}`
	ps := mustParse(t, raw)

	both := Context{Labels: map[string]map[string]string{IDProcessEnvVar: {"FOO": "1", "BAR": "2"}}}
	if got := Evaluate(ps[0].Rules, both); got != ResultTrue {
		t.Fatalf("both env vars present: want TRUE, got %v", got)
	}
	if got := Evaluate(ps[0].Rules, Context{Labels: map[string]map[string]string{IDProcessEnvVar: {"FOO": "1"}}}); got == ResultTrue {
		t.Fatalf("BAR absent: must not be TRUE, got %v", got)
	}

	// deny.go's "exists with any non-empty value" form: NAME=*? with CMP_WILDCARD.
	exists := mustParse(t, `{"policies":[{"rules":{"node_type":"EvaluatorNode","node":{"eval_type":"StrEvaluator","eval":{"id":"PROCESS_ENVAR","cmp":"CMP_WILDCARD","value":"FOO=*?"}}}}]}`)
	if got := Evaluate(exists[0].Rules, Context{Labels: map[string]map[string]string{IDProcessEnvVar: {"FOO": "bar"}}}); got != ResultTrue {
		t.Fatalf("FOO set non-empty: want TRUE, got %v", got)
	}
	if got := Evaluate(exists[0].Rules, Context{Labels: map[string]map[string]string{IDProcessEnvVar: {"FOO": ""}}}); got == ResultTrue {
		t.Fatalf("FOO empty: *? requires >=1 char, must not be TRUE, got %v", got)
	}
}

// TestParsePoliciesIsolatesBadPolicy checks that a single undecodable policy is
// skipped rather than dropping the whole document: a newer config with one
// policy an older agent can't decode must not leave it with zero policies. The
// good policy still loads, and the error reports the skipped one.
func TestParsePoliciesIsolatesBadPolicy(t *testing.T) {
	raw := `{"policies":[
	  {"description":"bad","rules":{"node_type":"EvaluatorNode","node":{"eval_type":"StrEvaluator","eval":{"id":"ALWAYS_TRUE"}}},"actions":[{"action":"ENABLE_SDK","values":["java"]}]},
	  {"description":"good","rules":{"node_type":"EvaluatorNode","node":{"eval_type":"StrEvaluator","eval":{"id":"ALWAYS_TRUE"}}},"actions":[{"action":"INJECT_ALLOW"}]}
	]}`
	ps, err := ParsePolicies([]byte(raw))
	if err == nil {
		t.Fatal("expected a non-nil error reporting the skipped policy")
	}
	if len(ps) != 1 || ps[0].Name != "good" {
		t.Fatalf("expected only the good policy to load, got %+v", ps)
	}
	if !ps[0].Outcome.Inject {
		t.Fatal("the surviving good policy should still apply its INJECT_ALLOW")
	}
}

// TestParseErrors covers the only cases ParsePolicies still rejects: a document
// that is not valid JSON, or one carrying a malformed value for an action the
// engine implements. Unrecognized/malformed *rules* no longer error -- they
// abstain (see TestUnrecognizedRuleConstructsAbstain).
func TestParseErrors(t *testing.T) {
	cases := map[string]string{
		"bad json":             `{`,
		"sdk value no equals":  `{"policies":[{"rules":{"node_type":"EvaluatorNode","node":{"eval_type":"StrEvaluator","eval":{"id":"ALWAYS_TRUE"}}},"actions":[{"action":"ENABLE_SDK","values":["java"]}]}]}`,
		"sdk value empty lang": `{"policies":[{"rules":{"node_type":"EvaluatorNode","node":{"eval_type":"StrEvaluator","eval":{"id":"ALWAYS_TRUE"}}},"actions":[{"action":"ENABLE_SDK","values":["=1.2"]}]}]}`,
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := ParsePolicies([]byte(raw)); err == nil {
				t.Errorf("expected error for %q", name)
			}
		})
	}
}

// TestUnrecognizedRuleConstructsAbstain checks that a rule an older agent cannot
// recognize (a newer schema) or a malformed rule decodes to an ABSTAIN leaf
// rather than failing the whole document -- mirroring the C engine's total
// evaluate_rules, so a policy from a newer schema stays evaluatable by an older
// agent. Rich facts are supplied so a non-abstaining (buggy) leaf would return a
// definite TRUE/FALSE, making the ABSTAIN assertion meaningful (e.g. a missing
// value silently treated as "" would match CMP_PREFIX).
func TestUnrecognizedRuleConstructsAbstain(t *testing.T) {
	ctx := Context{
		Strings:  map[string]string{IDNamespaceName: "x"},
		Numbers:  map[string]int64{"RUNTIME_VERSION_MAJOR": 1},
		UNumbers: map[string]uint64{"JAVA_HEAP": 1},
		Labels:   map[string]map[string]string{IDPodLabel: {"app": "x"}},
	}
	cases := map[string]string{
		"unknown node type":    `{"node_type":"Mystery","node":{}}`,
		"unknown boolean op":   `{"node_type":"CompositeNode","node":{"op":"BOOL_XOR","children":[]}}`,
		"not wrong arity":      `{"node_type":"CompositeNode","node":{"op":"BOOL_NOT","children":[{"node_type":"EvaluatorNode","node":{"eval_type":"StrEvaluator","eval":{"id":"ALWAYS_TRUE"}}},{"node_type":"EvaluatorNode","node":{"eval_type":"StrEvaluator","eval":{"id":"ALWAYS_TRUE"}}}]}}`,
		"unknown eval type":    `{"node_type":"EvaluatorNode","node":{"eval_type":"BogusEvaluator","eval":{}}}`,
		"unknown string cmp":   `{"node_type":"EvaluatorNode","node":{"eval_type":"StrEvaluator","eval":{"id":"NAMESPACE_NAME","cmp":"CMP_REGEX","value":"x"}}}`,
		"unknown numeric cmp":  `{"node_type":"EvaluatorNode","node":{"eval_type":"NumEvaluator","eval":{"id":"RUNTIME_VERSION_MAJOR","cmp":"CMP_BOGUS","value":1}}}`,
		"numeric missing cmp":  `{"node_type":"EvaluatorNode","node":{"eval_type":"NumEvaluator","eval":{}}}`,
		"string value omitted": `{"node_type":"EvaluatorNode","node":{"eval_type":"StrEvaluator","eval":{"id":"NAMESPACE_NAME","cmp":"CMP_PREFIX"}}}`,
		"string value null":    `{"node_type":"EvaluatorNode","node":{"eval_type":"StrEvaluator","eval":{"id":"NAMESPACE_NAME","cmp":"CMP_PREFIX","value":null}}}`,
		"label without equals": `{"node_type":"EvaluatorNode","node":{"eval_type":"StrEvaluator","eval":{"id":"POD_LABEL","cmp":"CMP_EXACT","value":"app"}}}`,
		"malformed node body":  `{"node_type":"CompositeNode","node":"not-an-object"}`,
	}
	for name, rules := range cases {
		t.Run(name, func(t *testing.T) {
			ps := mustParse(t, `{"policies":[{"rules":`+rules+`,"actions":[]}]}`)
			if got := Evaluate(ps[0].Rules, ctx); got != ResultAbstain {
				t.Errorf("expected ABSTAIN, got %v", got)
			}
		})
	}
}
