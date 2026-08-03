// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache 2.0 License. This product includes software developed at
// Datadog (https://www.datadoghq.com/).
//
// Copyright 2025-Present Datadog, Inc.

//go:build conformance_cgo

// Cross-engine half of the conformance suite: it runs every vector of
// testdata/vectors.json through BOTH engines in a single process and asserts
// they agree:
//
//   - the Go engine, via the public ParsePolicies + Evaluate path, and
//   - the real C engine (libpolicies.a), linked through cgo (see cgo.go), fed
//     the exact same vector serialized to a FlatBuffers NodeTypeWrapper buffer
//     (the wire form the C engine consumes in production).
//
// Gated behind the "conformance_cgo" build tag; run with `make -C go
// conformance-cross`. The portable, Go-only corpus test lives in corpus_test.go
// and runs everywhere.

package conformance

import (
	"testing"

	"github.com/DataDog/dd-policy-engine/go/policies"
)

// TestConformanceCrossEngine runs every corpus vector through the Go engine and
// the C engine and fails on any disagreement between them or with the expected
// result. It is the executable form of the cross-engine contract.
func TestConformanceCrossEngine(t *testing.T) {
	corpus := loadCorpus(t)

	for _, v := range corpus.Vectors {
		t.Run(v.Name, func(t *testing.T) {
			if _, ok := resultFromName(v.Expect); !ok {
				t.Fatalf("vector %q has invalid expect %q", v.Name, v.Expect)
			}

			// Go engine, via the real public API. Actions are included so the
			// document parses like production, but only the tri-state eval result
			// is compared across engines (C has no comparable action outcome).
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
			goName := resultName(policies.Evaluate(ps[0].Rules, v.Facts.toContext()))

			// C engine, fed the same vector serialized to FlatBuffers.
			buf, err := buildRulesBuffer(v.Rules)
			if err != nil {
				t.Fatalf("build rules buffer: %v", err)
			}
			cName, err := cEvaluateBuffer(buf, v.Facts.Strings, v.Facts.Labels, v.Facts.Numbers, v.Facts.UNumbers)
			if err != nil {
				t.Fatalf("C evaluate: %v", err)
			}

			if goName != cName {
				t.Errorf("engine divergence: Go=%s C=%s (expected %s)", goName, cName, v.Expect)
			}
			if goName != v.Expect {
				t.Errorf("Go engine: got %s, want %s", goName, v.Expect)
			}
			if cName != v.Expect {
				t.Errorf("C engine: got %s, want %s", cName, v.Expect)
			}
		})
	}
}

// TestBuildRulesBufferRejectsUnknownComparator ensures the C-buffer builder is as
// strict as the Go decoder (policies.ParsePolicies -> decodeStrCmp): an unknown
// wire identifier must error rather than silently serialize the enum's zero value
// (CMP_STR_UNKNOWN). Otherwise a typo in the shared corpus could mask, rather
// than surface, a real cross-engine divergence.
func TestBuildRulesBufferRejectsUnknownComparator(t *testing.T) {
	rules := []byte(`{
      "node_type": "EvaluatorNode",
      "node": {
        "eval_type": "StrEvaluator",
        "eval": {"id": "NAMESPACE_NAME", "cmp": "CMP_BOGUS", "value": "x"}
      }
    }`)
	if _, err := buildRulesBuffer(rules); err == nil {
		t.Fatalf("expected buildRulesBuffer to reject unknown comparator %q, got nil error", "CMP_BOGUS")
	}
}
