// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache 2.0 License. This product includes software developed at
// Datadog (https://www.datadoghq.com/).
//
// Copyright 2025-Present Datadog, Inc.

package conformance

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/DataDog/dd-policy-engine/go/policies"
)

// The conformance corpus is the cross-engine contract that guarantees the Go
// evaluator stays semantically identical to the C engine (RFC step 3). Each
// vector is a dd-wls rule tree (the JSON projection of the FlatBuffers schema),
// a set of facts, the expected tri-state result, and optionally the policy
// actions plus the expected decoded Outcome. This Go harness runs every vector
// through the real public API (ParsePolicies + Evaluate); the C harness is
// expected to run the same logical vectors against libpolicies so any drift
// between the two engines is caught by a shared corpus rather than by parallel,
// hand-maintained test suites.
//
// The corpus deliberately covers the semantics that must match exactly:
//   - tri-state AND/OR/NOT composition and short-circuit (ALWAYS_* leaves),
//   - the five string comparators (exact/prefix/suffix/contains/wildcard),
//   - the "KEY=" existence convention and absent-key handling,
//   - ABSTAIN when a fact source is unavailable (unset namespace name).
//
// Vectors with an `actions` array and an `expect_outcome` also exercise action
// decoding (inject decision, tracer versions, SET_ENVAR/ENABLE_PROFILER env-var
// configs). That check is Go-only: C actions are host callbacks with no
// comparable value, so the cross-engine harness asserts eval parity (`expect`)
// alone and ignores the outcome fields.
//
// Comparator edge cases (wildcard backtracking, empty strings) live in the
// policies package's TestWildcardMatch, which mirrors c/src/test/test_evaluator.c
// directly.

const conformanceCorpusDir = "testdata"

type conformanceCorpus struct {
	Vectors []conformanceVector `json:"vectors"`
}

type conformanceVector struct {
	Name   string           `json:"name"`
	Rules  json.RawMessage  `json:"rules"`
	Facts  conformanceFacts `json:"facts"`
	Expect string           `json:"expect"`

	// Actions and ExpectOutcome are optional: when a vector carries them, the Go
	// harness additionally decodes the policy actions and asserts the resulting
	// Outcome. The cross-engine (C) harness ignores both and checks only Expect,
	// because C actions are host callbacks with no comparable engine-level value.
	Actions       json.RawMessage     `json:"actions,omitempty"`
	ExpectOutcome *conformanceOutcome `json:"expect_outcome,omitempty"`
}

// conformanceOutcome is the expected decoded Outcome for a vector's actions.
// TracerConfigs is compared as a name->value map so ordering is irrelevant.
type conformanceOutcome struct {
	Inject         bool              `json:"inject"`
	InjectSet      bool              `json:"inject_set"`
	TracerVersions map[string]string `json:"tracer_versions"`
	TracerConfigs  map[string]string `json:"tracer_configs"`
}

// conformanceFacts mirrors the generic Context: workload facts keyed by
// evaluator id. A missing entry means the source is unavailable (ABSTAIN).
type conformanceFacts struct {
	Strings  map[string]string            `json:"strings"`
	Labels   map[string]map[string]string `json:"labels"`
	Numbers  map[string]int64             `json:"numbers"`
	UNumbers map[string]uint64            `json:"unumbers"`
}

func (f conformanceFacts) toContext() policies.Context {
	return policies.Context{
		Strings:  f.Strings,
		Labels:   f.Labels,
		Numbers:  f.Numbers,
		UNumbers: f.UNumbers,
	}
}

func resultFromName(name string) (policies.Result, bool) {
	switch name {
	case "TRUE":
		return policies.ResultTrue, true
	case "FALSE":
		return policies.ResultFalse, true
	case "ABSTAIN":
		return policies.ResultAbstain, true
	default:
		return policies.ResultAbstain, false
	}
}

func resultName(r policies.Result) string {
	switch r {
	case policies.ResultTrue:
		return "TRUE"
	case policies.ResultFalse:
		return "FALSE"
	default:
		return "ABSTAIN"
	}
}

// wrapVectorAsDocument embeds a vector's rule tree (and optional actions) into a
// minimal dd-wls policies document so it can be parsed through the real
// ParsePolicies path instead of a test-only decoder. actions may be nil, in
// which case the policy carries an empty actions list.
func wrapVectorAsDocument(name string, rules, actions json.RawMessage) ([]byte, error) {
	if len(actions) == 0 {
		actions = json.RawMessage("[]")
	}
	doc := struct {
		Policies []struct {
			Description string          `json:"description"`
			Rules       json.RawMessage `json:"rules"`
			Actions     json.RawMessage `json:"actions"`
		} `json:"policies"`
	}{
		Policies: []struct {
			Description string          `json:"description"`
			Rules       json.RawMessage `json:"rules"`
			Actions     json.RawMessage `json:"actions"`
		}{
			{Description: name, Rules: rules, Actions: actions},
		},
	}
	return json.Marshal(doc)
}

// checkOutcome asserts a decoded policy Outcome matches the vector's expectation.
// TracerConfigs is compared order-insensitively as a name->value map.
func checkOutcome(t *testing.T, want conformanceOutcome, got policies.Outcome) {
	t.Helper()
	if got.Inject != want.Inject || got.InjectSet != want.InjectSet {
		t.Errorf("outcome inject: got {inject:%v set:%v} want {inject:%v set:%v}",
			got.Inject, got.InjectSet, want.Inject, want.InjectSet)
	}
	if !sameStringMap(got.TracerVersions, want.TracerVersions) {
		t.Errorf("outcome tracer_versions: got %v want %v", got.TracerVersions, want.TracerVersions)
	}
	gotConfigs := map[string]string{}
	for _, e := range got.TracerConfigs {
		gotConfigs[e.Name] = e.Value
	}
	if !sameStringMap(gotConfigs, want.TracerConfigs) {
		t.Errorf("outcome tracer_configs: got %v want %v", gotConfigs, want.TracerConfigs)
	}
}

// sameStringMap compares two string maps, treating nil and empty as equal.
func sameStringMap(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

// loadCorpus reads every *.json file under testdata/ and merges their vectors,
// so the corpus can be split into focused, readable per-category files instead
// of one monolith. Vector names must be unique across all files.
func loadCorpus(t *testing.T) conformanceCorpus {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(conformanceCorpusDir, "*.json"))
	if err != nil {
		t.Fatalf("glob corpus: %v", err)
	}
	if len(files) == 0 {
		t.Fatalf("no corpus files in %s", conformanceCorpusDir)
	}

	var corpus conformanceCorpus
	seen := map[string]string{} // vector name -> file it first appeared in
	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		var part conformanceCorpus
		if err := json.Unmarshal(raw, &part); err != nil {
			t.Fatalf("parse %s: %v", f, err)
		}
		for _, v := range part.Vectors {
			if prev, dup := seen[v.Name]; dup {
				t.Fatalf("duplicate vector name %q in %s and %s", v.Name, prev, f)
			}
			seen[v.Name] = f
			corpus.Vectors = append(corpus.Vectors, v)
		}
	}
	if len(corpus.Vectors) == 0 {
		t.Fatal("conformance corpus is empty")
	}
	return corpus
}

func TestConformanceCorpus(t *testing.T) {
	corpus := loadCorpus(t)

	for _, v := range corpus.Vectors {
		t.Run(v.Name, func(t *testing.T) {
			want, ok := resultFromName(v.Expect)
			if !ok {
				t.Fatalf("vector %q has invalid expect %q", v.Name, v.Expect)
			}

			doc, err := wrapVectorAsDocument(v.Name, v.Rules, v.Actions)
			if err != nil {
				t.Fatalf("wrap vector: %v", err)
			}
			ps, err := policies.ParsePolicies(doc)
			if err != nil {
				t.Fatalf("ParsePolicies: %v", err)
			}
			if len(ps) != 1 {
				t.Fatalf("expected 1 policy, got %d", len(ps))
			}

			got := policies.Evaluate(ps[0].Rules, v.Facts.toContext())
			if got != want {
				t.Errorf("vector %q: got %s want %s", v.Name, resultName(got), v.Expect)
			}

			// When the vector declares expected actions, also assert the decoded
			// Outcome. This is Go-only; the cross-engine harness checks eval parity.
			if v.ExpectOutcome != nil {
				checkOutcome(t, *v.ExpectOutcome, ps[0].Outcome)
			}
		})
	}
}
