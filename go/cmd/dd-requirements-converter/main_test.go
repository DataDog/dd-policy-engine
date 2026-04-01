package main

import (
	"encoding/json"
	"testing"

	"github.com/DataDog/dd-policy-engine/go/cmd/dd-requirements-converter/converter"
	"github.com/DataDog/dd-policy-engine/go/schema/dd/wls"

	flatbuffers "github.com/google/flatbuffers/go"
)

// expectedNode represents an expected node in the policy tree
type expectedNode struct {
	nodeType wls.NodeType

	// For CompositeNode
	op       wls.BoolOperation
	children []expectedNode

	// For EvaluatorNode
	evalType wls.EvaluatorType

	// For StrEvaluator
	strField wls.StringEvaluators
	strCmp   wls.CmpTypeSTR
	strValue string

	// For NumEvaluator
	numField wls.NumericEvaluators
	numCmp   wls.CmpTypeNUM
	numValue int64
}

// strEval creates an expected string evaluator (defaults to EXACT comparison)
func strEval(field wls.StringEvaluators, value string, cmp ...wls.CmpTypeSTR) expectedNode {
	cmpType := wls.CmpTypeSTRCMP_EXACT
	if len(cmp) > 0 {
		cmpType = cmp[0]
	}
	return expectedNode{
		nodeType: wls.NodeTypeEvaluatorNode,
		evalType: wls.EvaluatorTypeStrEvaluator,
		strField: field,
		strCmp:   cmpType,
		strValue: value,
	}
}

func numEval(field wls.NumericEvaluators, cmp wls.CmpTypeNUM, value int64) expectedNode {
	return expectedNode{
		nodeType: wls.NodeTypeEvaluatorNode,
		evalType: wls.EvaluatorTypeNumEvaluator,
		numField: field,
		numCmp:   cmp,
		numValue: value,
	}
}

func andNode(children ...expectedNode) expectedNode {
	return expectedNode{
		nodeType: wls.NodeTypeCompositeNode,
		op:       wls.BoolOperationBOOL_AND,
		children: children,
	}
}

func orNode(children ...expectedNode) expectedNode {
	return expectedNode{
		nodeType: wls.NodeTypeCompositeNode,
		op:       wls.BoolOperationBOOL_OR,
		children: children,
	}
}

// Comparison functions to compare actual output against expected structure
func compareNode(t *testing.T, path string, actual *wls.NodeTypeWrapper, expected expectedNode) {
	t.Helper()

	if actual.NodeType() != expected.nodeType {
		t.Errorf("%s: expected node type %s, got %s", path, expected.nodeType.String(), actual.NodeType().String())
		return
	}

	var table flatbuffers.Table
	if !actual.Node(&table) {
		t.Errorf("%s: failed to get node table", path)
		return
	}

	switch expected.nodeType {
	case wls.NodeTypeCompositeNode:
		var composite wls.CompositeNode
		composite.Init(table.Bytes, table.Pos)
		compareCompositeNode(t, path, &composite, expected)

	case wls.NodeTypeEvaluatorNode:
		var evalNode wls.EvaluatorNode
		evalNode.Init(table.Bytes, table.Pos)
		compareEvaluatorNode(t, path, &evalNode, expected)
	}
}

func compareCompositeNode(t *testing.T, path string, actual *wls.CompositeNode, expected expectedNode) {
	t.Helper()

	if actual.Op() != expected.op {
		t.Errorf("%s: expected op %s, got %s", path, expected.op.String(), actual.Op().String())
	}

	if actual.ChildrenLength() != len(expected.children) {
		t.Errorf("%s: expected %d children, got %d", path, len(expected.children), actual.ChildrenLength())
		return
	}

	for i, expectedChild := range expected.children {
		var actualChild wls.NodeTypeWrapper
		if !actual.Children(&actualChild, i) {
			t.Errorf("%s[%d]: failed to get child", path, i)
			continue
		}
		compareNode(t, path+"["+string(rune('0'+i))+"]", &actualChild, expectedChild)
	}
}

func compareEvaluatorNode(t *testing.T, path string, actual *wls.EvaluatorNode, expected expectedNode) {
	t.Helper()

	if actual.EvalType() != expected.evalType {
		t.Errorf("%s: expected eval type %s, got %s", path, expected.evalType.String(), actual.EvalType().String())
		return
	}

	var evalTable flatbuffers.Table
	if !actual.Eval(&evalTable) {
		t.Errorf("%s: failed to get evaluator table", path)
		return
	}

	switch expected.evalType {
	case wls.EvaluatorTypeStrEvaluator:
		var strEval wls.StrEvaluator
		strEval.Init(evalTable.Bytes, evalTable.Pos)

		if strEval.Id() != expected.strField {
			t.Errorf("%s: expected str field %s, got %s", path, expected.strField.String(), strEval.Id().String())
		}
		if strEval.Cmp() != expected.strCmp {
			t.Errorf("%s: expected str cmp %s, got %s", path, expected.strCmp.String(), strEval.Cmp().String())
		}
		if string(strEval.Value()) != expected.strValue {
			t.Errorf("%s: expected str value '%s', got '%s'", path, expected.strValue, string(strEval.Value()))
		}

	case wls.EvaluatorTypeNumEvaluator:
		var numEval wls.NumEvaluator
		numEval.Init(evalTable.Bytes, evalTable.Pos)

		if numEval.Id() != expected.numField {
			t.Errorf("%s: expected num field %s, got %s", path, expected.numField.String(), numEval.Id().String())
		}
		if numEval.Cmp() != expected.numCmp {
			t.Errorf("%s: expected num cmp %s, got %s", path, expected.numCmp.String(), numEval.Cmp().String())
		}
		if numEval.Value() != expected.numValue {
			t.Errorf("%s: expected num value %d, got %d", path, expected.numValue, numEval.Value())
		}
	}
}

func TestJSONlibc_ConvertToWLS(t *testing.T) {
	tests := []struct {
		name         string
		inputJSON    string
		flavor       string
		expectNil    bool
		expectedRoot expectedNode
	}{
		{
			name:      "supported with no version - no policy needed",
			inputJSON: `{"arch": "x64", "supported": true}`,
			flavor:    "glibc",
			expectNil: true,
		},
		{
			// JSON: {"arch": "x64", "supported": true, "min": "2.17"}
			// Policy: DENY if arch=x86_64 AND flavor=glibc AND version < 2.17
			name:      "supported with version - deny if version < min",
			inputJSON: `{"arch": "x64", "supported": true, "min": "2.17"}`,
			flavor:    "glibc",
			expectNil: false,
			expectedRoot: andNode(
				strEval(wls.StringEvaluatorsMACHINE_ARCHITECTURE, "x86_64"),
				strEval(wls.StringEvaluatorsLIBC_FLAVOR, "glibc"),
				orNode( // version < 2.17
					numEval(wls.NumericEvaluatorsLIBC_VERSION_MAJOR, wls.CmpTypeNUMCMP_GT, 2),
					andNode( // major == 2 AND minor < 17
						numEval(wls.NumericEvaluatorsLIBC_VERSION_MAJOR, wls.CmpTypeNUMCMP_EQ, 2),
						numEval(wls.NumericEvaluatorsLIBC_VERSION_MINOR, wls.CmpTypeNUMCMP_GT, 17),
					),
					andNode( // major == 2 AND minor == 17 AND patch < 0
						numEval(wls.NumericEvaluatorsLIBC_VERSION_MAJOR, wls.CmpTypeNUMCMP_EQ, 2),
						numEval(wls.NumericEvaluatorsLIBC_VERSION_MINOR, wls.CmpTypeNUMCMP_EQ, 17),
						numEval(wls.NumericEvaluatorsLIBC_VERSION_PATCH, wls.CmpTypeNUMCMP_GT, 0),
					),
				),
			),
		},
		{
			// JSON: {"arch": "arm64", "supported": false}
			// Policy: DENY if arch=aarch64 AND flavor=musl (entirely unsupported)
			name:      "unsupported with no version - deny all matching arch+flavor",
			inputJSON: `{"arch": "arm64", "supported": false}`,
			flavor:    "musl",
			expectNil: false,
			expectedRoot: andNode(
				strEval(wls.StringEvaluatorsMACHINE_ARCHITECTURE, "aarch64"),
				strEval(wls.StringEvaluatorsLIBC_FLAVOR, "musl"),
			),
		},
		{
			// 32-bit x86: JSON "x86" → policy operand "x86"
			name:      "32-bit x86 architecture",
			inputJSON: `{"arch": "x86", "supported": false}`,
			flavor:    "glibc",
			expectNil: false,
			expectedRoot: andNode(
				strEval(wls.StringEvaluatorsMACHINE_ARCHITECTURE, "x86"),
				strEval(wls.StringEvaluatorsLIBC_FLAVOR, "glibc"),
			),
		},
		{
			// 32-bit ARM: JSON "arm" → policy operand "arm"
			name:      "32-bit ARM architecture",
			inputJSON: `{"arch": "arm", "supported": false}`,
			flavor:    "musl",
			expectNil: false,
			expectedRoot: andNode(
				strEval(wls.StringEvaluatorsMACHINE_ARCHITECTURE, "arm"),
				strEval(wls.StringEvaluatorsLIBC_FLAVOR, "musl"),
			),
		},
		{
			// JSON: {"arch": "x64", "supported": false, "min": "2.30"}
			// Policy: DENY if arch=x86_64 AND flavor=glibc AND version >= 2.30
			name:      "unsupported with version - deny if version >= min",
			inputJSON: `{"arch": "x64", "supported": false, "min": "2.30"}`,
			flavor:    "glibc",
			expectNil: false,
			expectedRoot: andNode(
				strEval(wls.StringEvaluatorsMACHINE_ARCHITECTURE, "x86_64"),
				strEval(wls.StringEvaluatorsLIBC_FLAVOR, "glibc"),
				orNode( // version >= 2.30
					numEval(wls.NumericEvaluatorsLIBC_VERSION_MAJOR, wls.CmpTypeNUMCMP_LT, 2),
					andNode( // major == 2 AND minor > 30
						numEval(wls.NumericEvaluatorsLIBC_VERSION_MAJOR, wls.CmpTypeNUMCMP_EQ, 2),
						numEval(wls.NumericEvaluatorsLIBC_VERSION_MINOR, wls.CmpTypeNUMCMP_LTE, 30),
					),
					andNode( // major == 2 AND minor == 30 AND patch >= 0
						numEval(wls.NumericEvaluatorsLIBC_VERSION_MAJOR, wls.CmpTypeNUMCMP_EQ, 2),
						numEval(wls.NumericEvaluatorsLIBC_VERSION_MINOR, wls.CmpTypeNUMCMP_EQ, 30),
						numEval(wls.NumericEvaluatorsLIBC_VERSION_PATCH, wls.CmpTypeNUMCMP_LTE, 0),
					),
				),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var libc converter.JSONlibc
			if err := json.Unmarshal([]byte(tt.inputJSON), &libc); err != nil {
				t.Fatalf("Failed to unmarshal JSON: %v", err)
			}

			builder := flatbuffers.NewBuilder(1024)
			offset, err := libc.ConvertToWLS(builder, tt.flavor)
			if err != nil {
				t.Fatalf("ConvertToWLS failed: %v", err)
			}

			if tt.expectNil {
				if offset != 0 {
					t.Errorf("Expected no node (offset=0), got offset=%d", offset)
				}
				return
			}

			if offset == 0 {
				t.Fatal("Expected a node, got offset=0")
			}

			builder.Finish(offset)
			wrapper := wls.GetRootAsNodeTypeWrapper(builder.FinishedBytes(), 0)

			compareNode(t, "root", wrapper, tt.expectedRoot)
		})
	}
}

func TestCmdPattern_ConvertToWLS(t *testing.T) {
	tests := []struct {
		name         string
		pattern      string
		expectedRoot expectedNode
	}{
		{
			// Exact match: "/usr/bin/python"
			// → StrEvaluator(PROCESS_EXE_FULL_PATH, EXACT, "/usr/bin/python")
			name:         "exact path match",
			pattern:      "/usr/bin/python",
			expectedRoot: strEval(wls.StringEvaluatorsPROCESS_EXE_FULL_PATH, "/usr/bin/python", wls.CmpTypeSTRCMP_EXACT),
		},
		{
			// Wildcard suffix: "/usr/bin/*"
			// → StrEvaluator(PROCESS_EXE_FULL_PATH, WILDCARD, "/usr/bin/")
			name:         "wildcard suffix - prefix match",
			pattern:      "/usr/bin/*",
			expectedRoot: strEval(wls.StringEvaluatorsPROCESS_EXE_FULL_PATH, "/usr/bin/*", wls.CmpTypeSTRCMP_WILDCARD),
		},
		{
			// Wildcard prefix: "**/python"
			// → StrEvaluator(PROCESS_EXE_FULL_PATH, WILDCARD, "**/python")
			name:         "wildcard prefix - suffix match",
			pattern:      "**/python",
			expectedRoot: strEval(wls.StringEvaluatorsPROCESS_EXE_FULL_PATH, "**/python", wls.CmpTypeSTRCMP_WILDCARD),
		},
		{
			// Middle wildcard: "/usr/*/python"
			// → StrEvaluator(PROCESS_EXE_FULL_PATH, WILDCARD "/usr/*/python")
			name:         "middle wildcard",
			pattern:      "/usr/*/python",
			expectedRoot: strEval(wls.StringEvaluatorsPROCESS_EXE_FULL_PATH, "/usr/*/python", wls.CmpTypeSTRCMP_WILDCARD),
		},
		{
			// Double wildcard both ends: "**/bin/*"
			// → StrEvaluator(PROCESS_EXE_FULL_PATH, WILDCARD, "**/bin/*")
			name:         "wildcards both ends - contains match",
			pattern:      "**/bin/*",
			expectedRoot: strEval(wls.StringEvaluatorsPROCESS_EXE_FULL_PATH, "**/bin/*", wls.CmpTypeSTRCMP_WILDCARD),
		},
		{
			// Wildcards both ends: "**/bin/**/python*"
			// → StrEvaluator(PROCESS_EXE_FULL_PATH, WILDCARD, "**/bin/**/python*")
			name:         "wildcards both ends - multiple contains matches",
			pattern:      "**/bin/**/python*",
			expectedRoot: strEval(wls.StringEvaluatorsPROCESS_EXE_FULL_PATH, "**/bin/**/python*", wls.CmpTypeSTRCMP_WILDCARD),
		},
		{
			// Wildcard suffix: "/usr/b?n"
			// → StrEvaluator(PROCESS_EXE_FULL_PATH, WILDCARD, "/usr/b?n")
			name:         "wildcard suffix - prefix match",
			pattern:      "/usr/b?n",
			expectedRoot: strEval(wls.StringEvaluatorsPROCESS_EXE_FULL_PATH, "/usr/b?n", wls.CmpTypeSTRCMP_WILDCARD),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := converter.CmdPattern(tt.pattern)

			builder := flatbuffers.NewBuilder(1024)
			offset, err := cmd.ConvertToWLS(builder)
			if err != nil {
				t.Fatalf("ConvertToWLS failed: %v", err)
			}

			builder.Finish(offset)
			wrapper := wls.GetRootAsNodeTypeWrapper(builder.FinishedBytes(), 0)

			compareNode(t, "root", wrapper, tt.expectedRoot)
		})
	}
}

func TestArgumentList_ConvertToWLS(t *testing.T) {
	tests := []struct {
		name         string
		inputJSON    string
		expectedRoot expectedNode
	}{
		{
			// Single exact argument: {"args": ["-version"]}
			// → StrEvaluator(PROCESS_ARGV, EXACT, "-version")
			name:         "single exact argument",
			inputJSON:    `{"args": ["-version"]}`,
			expectedRoot: strEval(wls.StringEvaluatorsPROCESS_ARGV, "-version"),
		},
		{
			// Argument at position: {"args": ["-flag"], "position": 1}
			// → StrEvaluator(PROCESS_ARGV_1, EXACT, "-flag")
			name:         "argument at specific position",
			inputJSON:    `{"args": ["-flag"], "position": 1}`,
			expectedRoot: strEval(wls.StringEvaluatorsPROCESS_ARGV_1, "-flag"),
		},
		{
			// Multiple arguments: {"args": ["-a", "-b"], "position": 1}
			// → AND(ARGV_1 == "-a", ARGV_2 == "-b")
			name:      "multiple arguments at consecutive positions",
			inputJSON: `{"args": ["-a", "-b"], "position": 1}`,
			expectedRoot: andNode(
				strEval(wls.StringEvaluatorsPROCESS_ARGV_1, "-a"),
				strEval(wls.StringEvaluatorsPROCESS_ARGV_2, "-b"),
			),
		},
		{
			name:         "wildcard in argument",
			inputJSON:    `{"args": ["--config=*"]}`,
			expectedRoot: strEval(wls.StringEvaluatorsPROCESS_ARGV, "--config=*", wls.CmpTypeSTRCMP_WILDCARD),
		},
		{
			name:         "last argument position with wildcard",
			inputJSON:    `{"args": ["*.txt"], "position": -1}`,
			expectedRoot: strEval(wls.StringEvaluatorsPROCESS_ARGV_N, "*.txt", wls.CmpTypeSTRCMP_WILDCARD),
		},
		{
			name:         "question mark wildcard in argument",
			inputJSON:    `{"args": ["file-?.log"]}`,
			expectedRoot: strEval(wls.StringEvaluatorsPROCESS_ARGV, "file-?.log", wls.CmpTypeSTRCMP_WILDCARD),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var argList converter.ArgumentList
			if err := json.Unmarshal([]byte(tt.inputJSON), &argList); err != nil {
				t.Fatalf("Failed to unmarshal JSON: %v", err)
			}

			builder := flatbuffers.NewBuilder(1024)
			offset, err := argList.ConvertToWLS(builder)
			if err != nil {
				t.Fatalf("ConvertToWLS failed: %v", err)
			}

			builder.Finish(offset)
			wrapper := wls.GetRootAsNodeTypeWrapper(builder.FinishedBytes(), 0)

			compareNode(t, "root", wrapper, tt.expectedRoot)
		})
	}
}

func TestJSONDeny_ConvertToWLS(t *testing.T) {
	tests := []struct {
		name         string
		inputJSON    string
		expectedRoot expectedNode
	}{
		{
			// OS only: {"os": "linux"}
			// → StrEvaluator(OS, EXACT, "linux")
			name:         "os only",
			inputJSON:    `{"os": "linux", "description": "deny linux"}`,
			expectedRoot: strEval(wls.StringEvaluatorsOS, "linux"),
		},
		{
			// Single cmd: {"cmds": ["/usr/bin/curl"]}
			// → StrEvaluator(PROCESS_EXE_FULL_PATH, EXACT, "/usr/bin/curl")
			name:         "single exact command",
			inputJSON:    `{"cmds": ["/usr/bin/curl"], "description": "deny curl"}`,
			expectedRoot: strEval(wls.StringEvaluatorsPROCESS_EXE_FULL_PATH, "/usr/bin/curl"),
		},
		{
			// Multiple cmds: {"cmds": ["/usr/bin/curl", "/usr/bin/wget"]}
			// → OR(cmd1, cmd2)
			name:      "multiple commands - OR",
			inputJSON: `{"cmds": ["/usr/bin/curl", "/usr/bin/wget"], "description": "deny download tools"}`,
			expectedRoot: orNode(
				strEval(wls.StringEvaluatorsPROCESS_EXE_FULL_PATH, "/usr/bin/curl"),
				strEval(wls.StringEvaluatorsPROCESS_EXE_FULL_PATH, "/usr/bin/wget"),
			),
		},
		{
			// OS + cmd: {"os": "linux", "cmds": ["/bin/rm"]}
			// → AND(OS == "linux", cmd == "/bin/rm")
			name:      "os and command",
			inputJSON: `{"os": "linux", "cmds": ["/bin/rm"], "description": "deny rm on linux"}`,
			expectedRoot: andNode(
				strEval(wls.StringEvaluatorsOS, "linux"),
				strEval(wls.StringEvaluatorsPROCESS_EXE_FULL_PATH, "/bin/rm"),
			),
		},
		{
			// Single env var: {"envars": {"DEBUG": "1"}}
			// → StrEvaluator(PROCESS_ENVAR, EXACT, "DEBUG=1")
			name:         "single environment variable",
			inputJSON:    `{"envars": {"DEBUG": "1"}, "description": "deny debug mode"}`,
			expectedRoot: strEval(wls.StringEvaluatorsPROCESS_ENVAR, "DEBUG=1"),
		},
		{
			name:         "environment variable wildcard asterisk in value",
			inputJSON:    `{"envars": {"PATH": "/usr/*/bin"}, "description": "deny path pattern"}`,
			expectedRoot: strEval(wls.StringEvaluatorsPROCESS_ENVAR, "PATH=/usr/*/bin", wls.CmpTypeSTRCMP_WILDCARD),
		},
		{
			name:         "environment variable wildcard question in value",
			inputJSON:    `{"envars": {"TERM": "xterm-?56color"}, "description": "deny term pattern"}`,
			expectedRoot: strEval(wls.StringEvaluatorsPROCESS_ENVAR, "TERM=xterm-?56color", wls.CmpTypeSTRCMP_WILDCARD),
		},
		{
			// JSON null value → KEY=*? + CMP_WILDCARD ("any non-empty value" for KEY=)
			name:         "environment variable null value matches non-empty only",
			inputJSON:    `{"envars": {"FOO": null}, "description": "deny when FOO set to any non-empty value"}`,
			expectedRoot: strEval(wls.StringEvaluatorsPROCESS_ENVAR, "FOO=*?", wls.CmpTypeSTRCMP_WILDCARD),
		},
		{
			// Single arg: {"args": [{"args": ["-rf"]}]}
			// → StrEvaluator(PROCESS_ARGV, EXACT, "-rf")
			name:         "single argument",
			inputJSON:    `{"args": [{"args": ["-rf"]}], "description": "deny -rf flag"}`,
			expectedRoot: strEval(wls.StringEvaluatorsPROCESS_ARGV, "-rf"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var deny converter.JSONDeny
			if err := json.Unmarshal([]byte(tt.inputJSON), &deny); err != nil {
				t.Fatalf("Failed to unmarshal JSON: %v", err)
			}

			builder := flatbuffers.NewBuilder(1024)
			offset, err := deny.ConvertToWLS(builder)
			if err != nil {
				t.Fatalf("ConvertToWLS failed: %v", err)
			}

			builder.Finish(offset)
			wrapper := wls.GetRootAsNodeTypeWrapper(builder.FinishedBytes(), 0)

			compareNode(t, "root", wrapper, tt.expectedRoot)
		})
	}
}

func TestParseRequirementsJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "incomplete object",
			input:   "{",
			wantErr: true,
		},
		{
			name:    "lone closing brace",
			input:   "}",
			wantErr: true,
		},
		{
			name:    "empty object version not one",
			input:   "{}",
			wantErr: false,
		},
		{
			name:    "version one with empty deny and native_deps",
			input:   `{"version":1,"deny":[],"native_deps":{"glibc":[],"musl":[]}}`,
			wantErr: false,
		},
		{
			name:    "whitespace only",
			input:   "   \n\t  ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := parseRequirementsJSON([]byte(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected parse error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestJSONRequirements_ConvertToWLS(t *testing.T) {
	tests := []struct {
		name              string
		inputJSON         string
		wantUnmarshalErr  bool
		wantVersionErr    bool
		expectedRuleCount int // when conversion runs: rules ORed together in the single policy
	}{
		{
			name:              "version one only",
			inputJSON:         `{"version":1}`,
			expectedRuleCount: 0,
		},
		{
			name:           "version zero",
			inputJSON:      `{"version":0}`,
			wantVersionErr: true,
		},
		{
			name:           "version two",
			inputJSON:      `{"version":2}`,
			wantVersionErr: true,
		},
		{
			name:             "version must be number",
			inputJSON:        `{ "version": "not a number" }`,
			wantUnmarshalErr: true,
		},
		{
			name:           "empty object missing version",
			inputJSON:      `{}`,
			wantVersionErr: true,
		},
		{
			name:              "empty requirements",
			inputJSON:         `{"version":1,"deny":[],"native_deps":{"glibc":[],"musl":[]}}`,
			expectedRuleCount: 0,
		},
		{
			name: "glibc + musl + deny combined",
			inputJSON: `{
				"version": 1,
				"deny": [{"os": "windows", "description": "no windows"}],
				"native_deps": {
					"glibc": [{"arch": "x64", "supported": true, "min": "2.17"}],
					"musl": [{"arch": "arm64", "supported": false}]
				}
			}`,
			expectedRuleCount: 3, // 1 deny + 1 glibc + 1 musl ORed together
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req converter.JSONRequirements
			err := json.Unmarshal([]byte(tt.inputJSON), &req)
			if tt.wantUnmarshalErr {
				if err == nil {
					t.Fatal("expected JSON unmarshal error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unmarshal JSON: %v", err)
			}

			err = validateRequirementsVersion(req)
			if tt.wantVersionErr {
				if err == nil {
					t.Fatal("expected validateRequirementsVersion error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("validateRequirementsVersion: %v", err)
			}

			builder := flatbuffers.NewBuilder(1024)
			offset, err := req.ConvertToWLS(builder)
			if err != nil {
				t.Fatalf("ConvertToWLS failed: %v", err)
			}

			builder.Finish(offset)
			policies := wls.GetRootAsPolicies(builder.FinishedBytes(), 0)

			// Always 1 policy now (all rules ORed together)
			if policies.PoliciesLength() != 1 {
				t.Errorf("Expected 1 policy, got %d", policies.PoliciesLength())
				return
			}

			var policy wls.Policy
			if !policies.Policies(&policy, 0) {
				t.Fatal("Failed to get policy")
			}

			rules := policy.Rules(nil)
			if rules == nil {
				t.Fatal("Expected rules node, got nil")
			}

			// The root should be an OR node with the expected number of children
			if rules.NodeType() != wls.NodeTypeCompositeNode {
				t.Fatalf("Expected composite node, got %s", rules.NodeType().String())
			}

			var table flatbuffers.Table
			if !rules.Node(&table) {
				t.Fatal("Failed to get node table")
			}

			var composite wls.CompositeNode
			composite.Init(table.Bytes, table.Pos)

			if composite.Op() != wls.BoolOperationBOOL_OR {
				t.Errorf("Expected OR operation, got %s", composite.Op().String())
			}

			if composite.ChildrenLength() != tt.expectedRuleCount {
				t.Errorf("Expected %d rules, got %d", tt.expectedRuleCount, composite.ChildrenLength())
			}
		})
	}
}
