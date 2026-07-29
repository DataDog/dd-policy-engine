// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache 2.0 License. This product includes software developed at
// Datadog (https://www.datadoghq.com/).
//
// Copyright 2025-Present Datadog, Inc.

package policies

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// schemaDir is the FlatBuffers schema shared with the C engine, relative to this
// package directory (go/policies).
const schemaDir = "../../fbs-schema"

var identRe = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)

// enumMembers extracts the member identifiers of the named FlatBuffers enum from
// a .fbs file. Comments are stripped so identifiers mentioned in documentation
// (e.g. "CMP_WILDCARD" inside an evaluator's comment) are not mistaken for
// members of an unrelated enum.
func enumMembers(t *testing.T, file, enumName string) map[string]bool {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(schemaDir, file))
	if err != nil {
		t.Fatalf("read %s: %v", file, err)
	}
	src := string(raw)

	start := strings.Index(src, "enum "+enumName)
	if start < 0 {
		t.Fatalf("enum %s not found in %s", enumName, file)
	}
	open := strings.IndexByte(src[start:], '{')
	end := strings.IndexByte(src[start:], '}')
	if open < 0 || end < 0 || end < open {
		t.Fatalf("malformed enum %s in %s", enumName, file)
	}
	body := src[start+open+1 : start+end]

	members := map[string]bool{}
	for _, line := range strings.Split(body, "\n") {
		if i := strings.Index(line, "//"); i >= 0 {
			line = line[:i]
		}
		for _, field := range strings.FieldsFunc(line, func(r rune) bool {
			return r == ',' || r == ' ' || r == '\t' || r == '\r'
		}) {
			// Drop an explicit "= 0" assignment (keep the name only).
			if eq := strings.IndexByte(field, '='); eq >= 0 {
				field = field[:eq]
			}
			field = strings.TrimSpace(field)
			if identRe.MatchString(field) {
				members[field] = true
			}
		}
	}
	return members
}

func assertSubset(t *testing.T, kind string, members map[string]bool, recognized []string) {
	t.Helper()
	for _, id := range recognized {
		if !members[id] {
			t.Errorf("%s %q recognized by the dd-wls decoder is absent from the FlatBuffers schema; "+
				"add it to fbs-schema and regenerate", kind, id)
		}
	}
}

// TestDecoderIdentifiersMatchSchema pins every identifier the dd-wls decoder
// recognizes to the FlatBuffers schema shared with the C engine. It reads the
// .fbs source (not the generated JSON schema, which lags the source) so the Go
// decoder and the C engine cannot drift apart silently.
func TestDecoderIdentifiersMatchSchema(t *testing.T) {
	// The decoder accepts any evaluator id generically, but the ids it gives
	// dedicated handling (constant evaluators and label-type ids) must exist in
	// the schema shared with the C engine.
	assertSubset(t, "string evaluator id",
		enumMembers(t, "evaluator_ids.fbs", "StringEvaluators"),
		[]string{
			IDAlwaysTrue, IDAlwaysFalse, IDAlwaysAbstain,
			IDNamespaceName, IDNamespaceLabel, IDPodLabel, IDPodAnnotation, IDContainerLabel,
			IDProcessEnvVar,
		})

	assertSubset(t, "string comparison",
		enumMembers(t, "evaluators.fbs", "CmpTypeSTR"),
		[]string{cmpExact, cmpPrefix, cmpSuffix, cmpContains, cmpWildcard})

	assertSubset(t, "numeric comparison",
		enumMembers(t, "evaluators.fbs", "CmpTypeNUM"),
		[]string{cmpEq, cmpGt, cmpGte, cmpLt, cmpLte})

	assertSubset(t, "boolean operation",
		enumMembers(t, "boolean_operation.fbs", "BoolOperation"),
		[]string{opAnd, opOr, opNot})

	assertSubset(t, "action id",
		enumMembers(t, "action_ids.fbs", "ActionId"),
		[]string{actionInjectAllow, actionInjectDeny, actionEnableSDK, actionEnableProfiler, actionSetEnvVar})
}
