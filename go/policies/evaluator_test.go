// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache 2.0 License. This product includes software developed at
// Datadog (https://www.datadoghq.com/).
//
// Copyright 2025-Present Datadog, Inc.

package policies

import (
	"math"
	"testing"
)

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
	if got := Evaluate(&Node{Op: OpAnd}, Context{}); got != ResultTrue {
		t.Errorf("empty AND = %v want ResultTrue", got)
	}
	if got := Evaluate(&Node{Op: OpOr}, Context{}); got != ResultFalse {
		t.Errorf("empty OR = %v want ResultFalse", got)
	}
	if got := Evaluate(Not(&Node{Op: OpAnd}), Context{}); got != ResultFalse {
		t.Errorf("NOT(empty AND) = %v want ResultFalse", got)
	}
}

func TestEvaluateLeafLabels(t *testing.T) {
	ctx := Context{
		Strings: map[string]string{IDNamespaceName: "payments"},
		Labels: map[string]map[string]string{
			IDPodLabel:       {"app": "web", "tier": "frontend"},
			IDNamespaceLabel: {},
		},
	}
	tests := []struct {
		name string
		node *Node
		want Result
	}{
		{"pod label match", LabelLeaf(IDPodLabel, "app", CmpExact, "web"), ResultTrue},
		{"pod label mismatch", LabelLeaf(IDPodLabel, "app", CmpExact, "db"), ResultFalse},
		{"pod label absent is false", LabelLeaf(IDPodLabel, "missing", CmpExact, "x"), ResultFalse},
		{"exists present", LabelLeaf(IDPodLabel, "tier", CmpExists, ""), ResultTrue},
		{"exists absent", LabelLeaf(IDPodLabel, "missing", CmpExists, ""), ResultFalse},
		{"namespace name match", StringLeaf(IDNamespaceName, CmpExact, "payments"), ResultTrue},
		{"namespace name mismatch", StringLeaf(IDNamespaceName, CmpExact, "billing"), ResultFalse},
		{"namespace label key absent is false", LabelLeaf(IDNamespaceLabel, "team", CmpExact, "x"), ResultFalse},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := Evaluate(tc.node, ctx); got != tc.want {
				t.Errorf("got %v want %v", got, tc.want)
			}
		})
	}
}

// TestSourceUnavailableAbstains checks that a fact source absent from the
// Context abstains rather than comparing a zero value, mirroring the C engine
// returning ABSTAIN on a NULL context. A present-but-empty source still
// evaluates concretely.
func TestSourceUnavailableAbstains(t *testing.T) {
	// String source unavailable.
	n := StringLeaf(IDNamespaceName, CmpExact, "payments")
	if got := Evaluate(n, Context{}); got != ResultAbstain {
		t.Errorf("unavailable namespace name = %v want ResultAbstain", got)
	}
	if got := Evaluate(n, Context{Strings: map[string]string{IDNamespaceName: ""}}); got != ResultFalse {
		t.Errorf("present-but-empty namespace name vs \"payments\" = %v want ResultFalse", got)
	}

	// Label source unavailable vs present-but-missing key.
	lbl := LabelLeaf(IDPodLabel, "app", CmpExact, "web")
	if got := Evaluate(lbl, Context{}); got != ResultAbstain {
		t.Errorf("unavailable pod labels = %v want ResultAbstain", got)
	}
	if got := Evaluate(lbl, Context{Labels: map[string]map[string]string{IDPodLabel: {}}}); got != ResultFalse {
		t.Errorf("present-but-missing pod label = %v want ResultFalse", got)
	}

	// Numeric source unavailable.
	num := NumericLeaf("JAVA_HEAP", NumEq, 100)
	if got := Evaluate(num, Context{}); got != ResultAbstain {
		t.Errorf("unavailable numeric fact = %v want ResultAbstain", got)
	}
}

// TestEvaluateNumeric exercises signed and unsigned numeric evaluators,
// mirroring plcs_default_numeric_evaluator ("evaluator value <cmp> workload").
func TestEvaluateNumeric(t *testing.T) {
	numCtx := Context{Numbers: map[string]int64{"RUNTIME_VERSION_MAJOR": 17}}
	numCases := []struct {
		name string
		node *Node
		want Result
	}{
		{"eq true", NumericLeaf("RUNTIME_VERSION_MAJOR", NumEq, 17), ResultTrue},
		{"eq false", NumericLeaf("RUNTIME_VERSION_MAJOR", NumEq, 11), ResultFalse},
		{"gt true (policy>workload)", NumericLeaf("RUNTIME_VERSION_MAJOR", NumGt, 21), ResultTrue},
		{"gt false", NumericLeaf("RUNTIME_VERSION_MAJOR", NumGt, 8), ResultFalse},
		{"lte true", NumericLeaf("RUNTIME_VERSION_MAJOR", NumLte, 17), ResultTrue},
		{"lt false equal", NumericLeaf("RUNTIME_VERSION_MAJOR", NumLt, 17), ResultFalse},
	}
	for _, tc := range numCases {
		t.Run("num/"+tc.name, func(t *testing.T) {
			if got := Evaluate(tc.node, numCtx); got != tc.want {
				t.Errorf("got %v want %v", got, tc.want)
			}
		})
	}

	uCtx := Context{UNumbers: map[string]uint64{"JAVA_HEAP": 1024}}
	uCases := []struct {
		name string
		node *Node
		want Result
	}{
		{"eq true", UNumericLeaf("JAVA_HEAP", NumEq, 1024), ResultTrue},
		{"gte true", UNumericLeaf("JAVA_HEAP", NumGte, 2048), ResultTrue},
		{"lt true", UNumericLeaf("JAVA_HEAP", NumLt, 512), ResultTrue},
		{"lt false", UNumericLeaf("JAVA_HEAP", NumLt, 4096), ResultFalse},
	}
	for _, tc := range uCases {
		t.Run("unum/"+tc.name, func(t *testing.T) {
			if got := Evaluate(tc.node, uCtx); got != tc.want {
				t.Errorf("got %v want %v", got, tc.want)
			}
		})
	}

	// A signed evaluator does not read the unsigned registry and vice versa.
	if got := Evaluate(NumericLeaf("JAVA_HEAP", NumEq, 1024), uCtx); got != ResultAbstain {
		t.Errorf("signed evaluator reading unsigned-only fact = %v want ResultAbstain", got)
	}
}

// TestNumericSentinelAbstains checks that a fact equal to the C engine's not-set
// sentinel (math.MaxInt64 / math.MaxUint64) abstains rather than being compared,
// matching evaluate_numeric/evaluate_unumeric (PLCS_NUM_NOT_SET/PLCS_UNUM_NOT_SET).
// The comparisons below would be TRUE if the sentinel were compared, so ABSTAIN
// proves the sentinel short-circuits.
func TestNumericSentinelAbstains(t *testing.T) {
	if got := Evaluate(NumericLeaf("RUNTIME_VERSION_MAJOR", NumLt, 0),
		Context{Numbers: map[string]int64{"RUNTIME_VERSION_MAJOR": math.MaxInt64}}); got != ResultAbstain {
		t.Errorf("signed sentinel fact: got %v want ResultAbstain", got)
	}
	if got := Evaluate(UNumericLeaf("JAVA_HEAP", NumLt, 0),
		Context{UNumbers: map[string]uint64{"JAVA_HEAP": math.MaxUint64}}); got != ResultAbstain {
		t.Errorf("unsigned sentinel fact: got %v want ResultAbstain", got)
	}
	// One below the sentinel is an ordinary, compared value.
	if got := Evaluate(NumericLeaf("RUNTIME_VERSION_MAJOR", NumLt, 0),
		Context{Numbers: map[string]int64{"RUNTIME_VERSION_MAJOR": math.MaxInt64 - 1}}); got != ResultTrue {
		t.Errorf("near-sentinel signed fact: got %v want ResultTrue", got)
	}
	if got := Evaluate(UNumericLeaf("JAVA_HEAP", NumLt, 0),
		Context{UNumbers: map[string]uint64{"JAVA_HEAP": math.MaxUint64 - 1}}); got != ResultTrue {
		t.Errorf("near-sentinel unsigned fact: got %v want ResultTrue", got)
	}
}

// TestEvaluateGenericStringID checks that an arbitrary, non-Kubernetes string
// evaluator id (here a host process fact) is supported generically.
func TestEvaluateGenericStringID(t *testing.T) {
	ctx := Context{Strings: map[string]string{"RUNTIME_LANGUAGE": "java"}}
	if got := Evaluate(StringLeaf("RUNTIME_LANGUAGE", CmpExact, "java"), ctx); got != ResultTrue {
		t.Errorf("runtime language match = %v want ResultTrue", got)
	}
	if got := Evaluate(StringLeaf("RUNTIME_LANGUAGE", CmpExact, "python"), ctx); got != ResultFalse {
		t.Errorf("runtime language mismatch = %v want ResultFalse", got)
	}
}

// TestEvaluateDepthLimit checks that a rule tree deeper than maxEvalDepth
// abstains instead of recursing, mirroring PLCS_MAX_EVAL_DEPTH in the C engine.
func TestEvaluateDepthLimit(t *testing.T) {
	// Wrap a TRUE leaf in NOT a number of times. Each NOT adds one level; an
	// even count would normally yield TRUE.
	nest := func(levels int) *Node {
		n := AlwaysTrue()
		for i := 0; i < levels; i++ {
			n = Not(n)
		}
		return n
	}
	// Within the limit: 2 NOTs over TRUE stays TRUE.
	if got := Evaluate(nest(2), Context{}); got != ResultTrue {
		t.Errorf("shallow tree = %v want ResultTrue", got)
	}
	// Beyond the limit: the deepest nodes abstain, so the whole tree abstains.
	if got := Evaluate(nest(maxEvalDepth+10), Context{}); got != ResultAbstain {
		t.Errorf("over-deep tree = %v want ResultAbstain", got)
	}
}
