package converter

import (
	"encoding/json"
	"testing"

	"github.com/DataDog/dd-policy-engine/go/schema/dd/wls"

	flatbuffers "github.com/google/flatbuffers/go"
)

type runtimeContext struct {
	strings  map[wls.StringEvaluators]string
	numerics map[wls.NumericEvaluators]int64
}

func evaluateNode(t *testing.T, wrapper *wls.NodeTypeWrapper, ctx runtimeContext) bool {
	t.Helper()

	var table flatbuffers.Table
	if !wrapper.Node(&table) {
		t.Fatal("node wrapper has no value")
	}

	switch wrapper.NodeType() {
	case wls.NodeTypeEvaluatorNode:
		var node wls.EvaluatorNode
		node.Init(table.Bytes, table.Pos)
		if !node.Eval(&table) {
			t.Fatal("evaluator node has no evaluator")
		}

		switch node.EvalType() {
		case wls.EvaluatorTypeStrEvaluator:
			var evaluator wls.StrEvaluator
			evaluator.Init(table.Bytes, table.Pos)
			return string(evaluator.Value()) == ctx.strings[evaluator.Id()]
		case wls.EvaluatorTypeNumEvaluator:
			var evaluator wls.NumEvaluator
			evaluator.Init(table.Bytes, table.Pos)
			policyValue := evaluator.Value()
			contextValue := ctx.numerics[evaluator.Id()]
			switch evaluator.Cmp() {
			case wls.CmpTypeNUMCMP_EQ:
				return policyValue == contextValue
			case wls.CmpTypeNUMCMP_GT:
				return policyValue > contextValue
			case wls.CmpTypeNUMCMP_GTE:
				return policyValue >= contextValue
			case wls.CmpTypeNUMCMP_LT:
				return policyValue < contextValue
			case wls.CmpTypeNUMCMP_LTE:
				return policyValue <= contextValue
			default:
				t.Fatalf("unsupported numeric comparator %s", evaluator.Cmp())
			}
		default:
			t.Fatalf("unsupported evaluator type %s", node.EvalType())
		}

	case wls.NodeTypeCompositeNode:
		var node wls.CompositeNode
		node.Init(table.Bytes, table.Pos)
		result := node.Op() == wls.BoolOperationBOOL_AND
		for i := 0; i < node.ChildrenLength(); i++ {
			var child wls.NodeTypeWrapper
			if !node.Children(&child, i) {
				t.Fatalf("missing child %d", i)
			}
			childResult := evaluateNode(t, &child, ctx)
			if node.Op() == wls.BoolOperationBOOL_AND {
				result = result && childResult
			} else if node.Op() == wls.BoolOperationBOOL_OR {
				result = result || childResult
			} else {
				t.Fatalf("unsupported boolean operator %s", node.Op())
			}
		}
		return result
	}

	t.Fatalf("unsupported node type %s", wrapper.NodeType())
	return false
}

func TestUnsupportedLibcPatchDoesNotDenyEarlierPatch(t *testing.T) {
	var requirement JSONlibc
	if err := json.Unmarshal([]byte(`{"arch":"x64","supported":false,"min":"2.30.5"}`), &requirement); err != nil {
		t.Fatal(err)
	}

	builder := flatbuffers.NewBuilder(1024)
	offset, err := requirement.ConvertToWLS(builder, "glibc")
	if err != nil {
		t.Fatal(err)
	}
	builder.Finish(offset)
	root := wls.GetRootAsNodeTypeWrapper(builder.FinishedBytes(), 0)

	tests := []struct {
		name                string
		major, minor, patch int64
		want                bool
	}{
		{name: "2.30.4", major: 2, minor: 30, patch: 4, want: false},
		{name: "2.30.5", major: 2, minor: 30, patch: 5, want: true},
		{name: "2.30.6", major: 2, minor: 30, patch: 6, want: true},
		{name: "2.31.0", major: 2, minor: 31, patch: 0, want: true},
		{name: "3.0.0", major: 3, minor: 0, patch: 0, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched := evaluateNode(t, root, runtimeContext{
				strings: map[wls.StringEvaluators]string{
					wls.StringEvaluatorsMACHINE_ARCHITECTURE: "x86_64",
					wls.StringEvaluatorsLIBC_FLAVOR:          "glibc",
				},
				numerics: map[wls.NumericEvaluators]int64{
					wls.NumericEvaluatorsLIBC_VERSION_MAJOR: tt.major,
					wls.NumericEvaluatorsLIBC_VERSION_MINOR: tt.minor,
					wls.NumericEvaluatorsLIBC_VERSION_PATCH: tt.patch,
				},
			})

			if matched != tt.want {
				t.Fatalf("match = %t, want %t for unsupported >= 2.30.5", matched, tt.want)
			}
		})
	}
}

func TestInvalidDenyOSCannotBroadenRule(t *testing.T) {
	builder := flatbuffers.NewBuilder(128)
	offset, err := (JSONDeny{
		Os:   "freebsd",
		Cmds: []CmdPattern{"/usr/bin/curl"},
	}).ConvertToWLS(builder)
	if err == nil {
		t.Fatalf("invalid OS was dropped, broadening the command deny rule at offset %d", offset)
	}
}

func TestDarwinDenyMatchesCanonicalMacOS(t *testing.T) {
	for _, input := range []string{"darwin", "macos"} {
		t.Run(input, func(t *testing.T) {
			builder := flatbuffers.NewBuilder(128)
			offset, err := (JSONDeny{Os: input}).ConvertToWLS(builder)
			if err != nil {
				t.Fatal(err)
			}
			builder.Finish(offset)
			root := wls.GetRootAsNodeTypeWrapper(builder.FinishedBytes(), 0)

			if !evaluateNode(t, root, runtimeContext{
				strings: map[wls.StringEvaluators]string{wls.StringEvaluatorsOS: "macos"},
			}) {
				t.Fatalf("%s requirement did not match the canonical macos runtime value", input)
			}
		})
	}
}

func TestEmptyLibcVersionReturnsError(t *testing.T) {
	testCases := []struct {
		name    string
		version string
	}{
		{name: "empty"},
		{name: "2", version: "2"},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var requirement JSONlibc
			if err := json.Unmarshal(
				[]byte(`{"arch":"x64","supported":true,"min":"`+testCase.version+`"}`),
				&requirement,
			); err != nil {
				t.Fatal(err)
			}

			if _, err := requirement.ConvertToWLS(flatbuffers.NewBuilder(128), "glibc"); err == nil {
				t.Fatalf("expected version %q to be rejected", testCase.version)
			}
		})
	}
}
