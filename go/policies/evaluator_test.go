// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache 2.0 License. This product includes software developed at
// Datadog (https://www.datadoghq.com/).
//
// Copyright 2025-Present Datadog, Inc.

package policies

import "testing"

func TestTriStateTruthTables(t *testing.T) {
	all := []Result{ResultFalse, ResultTrue, ResultAbstain}

	andWant := map[[2]Result]Result{
		{ResultFalse, ResultFalse}:     ResultFalse,
		{ResultFalse, ResultTrue}:      ResultFalse,
		{ResultFalse, ResultAbstain}:   ResultFalse,
		{ResultTrue, ResultFalse}:      ResultFalse,
		{ResultTrue, ResultTrue}:       ResultTrue,
		{ResultTrue, ResultAbstain}:    ResultAbstain,
		{ResultAbstain, ResultFalse}:   ResultFalse,
		{ResultAbstain, ResultTrue}:    ResultAbstain,
		{ResultAbstain, ResultAbstain}: ResultAbstain,
	}
	orWant := map[[2]Result]Result{
		{ResultFalse, ResultFalse}:     ResultFalse,
		{ResultFalse, ResultTrue}:      ResultTrue,
		{ResultFalse, ResultAbstain}:   ResultAbstain,
		{ResultTrue, ResultFalse}:      ResultTrue,
		{ResultTrue, ResultTrue}:       ResultTrue,
		{ResultTrue, ResultAbstain}:    ResultTrue,
		{ResultAbstain, ResultFalse}:   ResultAbstain,
		{ResultAbstain, ResultTrue}:    ResultTrue,
		{ResultAbstain, ResultAbstain}: ResultAbstain,
	}
	notWant := map[Result]Result{
		ResultFalse:   ResultTrue,
		ResultTrue:    ResultFalse,
		ResultAbstain: ResultAbstain,
	}

	for _, a := range all {
		for _, b := range all {
			if got := doAnd(a, b); got != andWant[[2]Result{a, b}] {
				t.Errorf("doAnd(%v,%v)=%v want %v", a, b, got, andWant[[2]Result{a, b}])
			}
			if got := doOr(a, b); got != orWant[[2]Result{a, b}] {
				t.Errorf("doOr(%v,%v)=%v want %v", a, b, got, orWant[[2]Result{a, b}])
			}
		}
		if got := doNot(a); got != notWant[a] {
			t.Errorf("doNot(%v)=%v want %v", a, got, notWant[a])
		}
	}
}

// TestWildcardMatch mirrors the wildcard conformance vectors exercised by the C
// engine in c/src/test/test_evaluator.c (string_evaluator_wildcard), so the Go
// matcher stays semantically identical to libpolicies.
func TestWildcardMatch(t *testing.T) {
	cases := []struct {
		pattern, s string
		want       bool
	}{
		// Exact match (no wildcards).
		{"hello", "hello", true},
		{"hello", "world", false},
		{"", "", true},
		{"", "a", false},
		{"a", "", false},

		// '?' matches exactly one character.
		{"?", "a", true},
		{"?", "", false},
		{"?", "ab", false},
		{"a?c", "abc", true},
		{"a?c", "aXc", true},
		{"a?c", "ac", false},
		{"a?c", "abbc", false},
		{"???", "abc", true},
		{"???", "ab", false},
		{"???", "abcd", false},

		// '*' matches zero or more characters.
		{"*", "", true},
		{"*", "anything", true},
		{"a*", "a", true},
		{"a*", "abc", true},
		{"a*", "b", false},
		{"*c", "c", true},
		{"*c", "abc", true},
		{"*c", "abd", false},
		{"a*c", "ac", true},
		{"a*c", "abc", true},
		{"a*c", "aXYZc", true},
		{"a*c", "aXYZd", false},

		// Multiple '*' wildcards.
		{"*foo*", "foo", true},
		{"*foo*", "XXfooYY", true},
		{"*foo*bar*", "foobar", true},
		{"*foo*bar*", "XXfooYYbarZZ", true},
		{"*foo*bar*", "barfoo", false},
		{"*foo*bar*", "fooXXbaz", false},

		// Consecutive '*' collapse to a single '*'.
		{"a***b", "ab", true},
		{"a***b", "aXXXb", true},

		// KEY=*? convention (non-empty RHS only).
		{"FOO=*?", "FOO=", false},
		{"FOO=*?", "FOO=a", true},
		{"FOO=*?", "FOO=ab", true},

		// Mixed '?' and '*'.
		{"a?*", "ab", true},
		{"a?*", "abc", true},
		{"a?*", "a", false},
		{"*?", "a", true},
		{"*?", "", false},

		// Executable path matching.
		{"**/java", "/usr/bin/java", true},
		{"**/java-1.5*/**/java", "/usr/lib/java-1.5.0/bin/java", true},
		{"**/java-1.5*/**/java", "/usr/lib/java-1.8.0/bin/java", false},
		{"**/exe?", "/some/exe2", true},
		{"**/exe?", "/some/other/exeA", true},
		{"**/exe?", "/some/exe", false},
		{"**/exe?", "/some/exe22", false},

		// Argument / suffix matching.
		{"1.*", "1.2.3", true},
		{"1.*", "2.0.0", false},
		{"*csc.dll", "csc.dll", true},
		{"*csc.dll", "/path/to/csc.dll", true},
		{"*csc.dll", "other.dll", false},

		// Pathological patterns: ensure no excessive backtracking.
		{"*a*a*a*a*b", "aaaaaaaaaaac", false},
		{"*a*a*a*a*b", "aaaaaaaaaaab", true},

		// Kubernetes-flavored labels/namespaces.
		{"foo*", "foobar", true},
		{"foo*", "barfoo", false},
		{"*bar", "foobar", true},
		{"k8s-*-svc", "k8s-payments-svc", true},
	}
	for _, c := range cases {
		if got := wildcardMatch(c.pattern, c.s); got != c.want {
			t.Errorf("wildcardMatch(%q,%q)=%v want %v", c.pattern, c.s, got, c.want)
		}
	}
}

func TestEmptyCompositeShortCircuit(t *testing.T) {
	if got := Evaluate(&Node{Op: OpAnd}, Facts{}); got != ResultTrue {
		t.Errorf("empty AND = %v want ResultTrue", got)
	}
	if got := Evaluate(&Node{Op: OpOr}, Facts{}); got != ResultFalse {
		t.Errorf("empty OR = %v want ResultFalse", got)
	}
	if got := Evaluate(Not(&Node{Op: OpAnd}), Facts{}); got != ResultFalse {
		t.Errorf("NOT(empty AND) = %v want ResultFalse", got)
	}
}

func TestEvaluateLeafLabels(t *testing.T) {
	f := Facts{
		NamespaceName: "payments",
		PodLabels:     map[string]string{"app": "web", "tier": "frontend"},
	}
	tests := []struct {
		name string
		node *Node
		want Result
	}{
		{"pod label match", Leaf(SourcePodLabel, "app", CmpExact, "web"), ResultTrue},
		{"pod label mismatch", Leaf(SourcePodLabel, "app", CmpExact, "db"), ResultFalse},
		{"pod label absent is false", Leaf(SourcePodLabel, "missing", CmpExact, "x"), ResultFalse},
		{"exists present", Leaf(SourcePodLabel, "tier", CmpExists, ""), ResultTrue},
		{"exists absent", Leaf(SourcePodLabel, "missing", CmpExists, ""), ResultFalse},
		{"namespace name match", Leaf(SourceNamespaceName, "", CmpExact, "payments"), ResultTrue},
		{"namespace name mismatch", Leaf(SourceNamespaceName, "", CmpExact, "billing"), ResultFalse},
		{"namespace label source absent", Leaf(SourceNamespaceLabel, "team", CmpExact, "x"), ResultFalse},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := Evaluate(tc.node, f); got != tc.want {
				t.Errorf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestDecideFirstMatchWins(t *testing.T) {
	ps := []Policy{
		{
			Name:    "java",
			Rules:   Leaf(SourcePodLabel, "app", CmpExact, "db"),
			Outcome: Outcome{TracerVersions: map[string]string{"java": "1"}},
		},
		{
			Name:    "catch-all",
			Rules:   AlwaysTrue(),
			Outcome: Outcome{TracerVersions: map[string]string{"php": "2"}},
		},
	}

	out, ok := Decide(ps, Facts{PodLabels: map[string]string{"app": "db"}})
	if !ok || out.TracerVersions["java"] != "1" {
		t.Fatalf("expected java policy, got %+v ok=%v", out, ok)
	}

	out, ok = Decide(ps, Facts{PodLabels: map[string]string{"app": "web"}})
	if !ok || out.TracerVersions["php"] != "2" {
		t.Fatalf("expected catch-all policy, got %+v ok=%v", out, ok)
	}

	_, ok = Decide(ps[:1], Facts{PodLabels: map[string]string{"app": "web"}})
	if ok {
		t.Fatalf("expected no match")
	}
}
