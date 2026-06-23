// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache 2.0 License. This product includes software developed at
// Datadog (https://www.datadoghq.com/).
//
// Copyright 2025-Present Datadog, Inc.

package conformance

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/DataDog/dd-policy-engine/go/policies"
)

// The conformance corpus is the cross-engine contract that guarantees the Go
// evaluator stays semantically identical to the C engine (RFC step 3). Each
// vector is a dd-wls rule tree (the JSON projection of the FlatBuffers schema),
// a set of facts, and the expected tri-state result. This Go harness runs every
// vector through the real public API (ParsePolicies + Evaluate); the C harness
// is expected to run the same logical vectors against libpolicies so any drift
// between the two engines is caught by a shared corpus rather than by parallel,
// hand-maintained test suites.
//
// The corpus deliberately covers the semantics that must match exactly:
//   - tri-state AND/OR/NOT composition and short-circuit (ALWAYS_* leaves),
//   - the five string comparators (exact/prefix/suffix/contains/wildcard),
//   - the "KEY=" existence convention and absent-key handling,
//   - ABSTAIN when a fact source is unavailable (unset namespace name).
//
// Comparator edge cases (wildcard backtracking, empty strings) live in the
// policies package's TestWildcardMatch, which mirrors c/src/test/test_evaluator.c
// directly.

const conformanceCorpusPath = "testdata/vectors.json"

type conformanceCorpus struct {
	Vectors []conformanceVector `json:"vectors"`
}

type conformanceVector struct {
	Name   string           `json:"name"`
	Rules  json.RawMessage  `json:"rules"`
	Facts  conformanceFacts `json:"facts"`
	Expect string           `json:"expect"`
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

// wrapVectorAsDocument embeds a vector's rule tree into a minimal dd-wls
// policies document so it can be parsed through the real ParsePolicies path
// instead of a test-only decoder.
func wrapVectorAsDocument(name string, rules json.RawMessage) ([]byte, error) {
	doc := struct {
		Policies []struct {
			Description string          `json:"description"`
			Rules       json.RawMessage `json:"rules"`
			Actions     []struct{}      `json:"actions"`
		} `json:"policies"`
	}{
		Policies: []struct {
			Description string          `json:"description"`
			Rules       json.RawMessage `json:"rules"`
			Actions     []struct{}      `json:"actions"`
		}{
			{Description: name, Rules: rules, Actions: []struct{}{}},
		},
	}
	return json.Marshal(doc)
}

func loadCorpus(t *testing.T) conformanceCorpus {
	t.Helper()
	raw, err := os.ReadFile(conformanceCorpusPath)
	if err != nil {
		t.Fatalf("read corpus: %v", err)
	}
	var corpus conformanceCorpus
	if err := json.Unmarshal(raw, &corpus); err != nil {
		t.Fatalf("parse corpus: %v", err)
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

			doc, err := wrapVectorAsDocument(v.Name, v.Rules)
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
		})
	}
}
