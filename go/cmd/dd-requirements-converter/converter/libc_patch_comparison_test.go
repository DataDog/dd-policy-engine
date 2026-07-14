package converter

import (
	"encoding/json"
	"testing"

	"github.com/DataDog/dd-policy-engine/go/schema/dd/wls"

	flatbuffers "github.com/google/flatbuffers/go"
)

type libcRuntimeContext struct {
	strings  map[wls.StringEvaluators]string
	numerics map[wls.NumericEvaluators]int64
}

func evaluateLibcNode(t *testing.T, wrapper *wls.NodeTypeWrapper, ctx libcRuntimeContext) bool {
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
			childResult := evaluateLibcNode(t, &child, ctx)
			switch node.Op() {
			case wls.BoolOperationBOOL_AND:
				result = result && childResult
			case wls.BoolOperationBOOL_OR:
				result = result || childResult
			default:
				t.Fatalf("unsupported boolean operator %s", node.Op())
			}
		}
		return result
	}

	t.Fatalf("unsupported node type %s", wrapper.NodeType())
	return false
}

func convertLibcRequirement(t *testing.T, supported bool) *wls.NodeTypeWrapper {
	t.Helper()

	var requirement JSONlibc
	input := `{"arch":"x64","supported":false,"min":"2.30.5"}`
	if supported {
		input = `{"arch":"x64","supported":true,"min":"2.30.5"}`
	}
	if err := json.Unmarshal([]byte(input), &requirement); err != nil {
		t.Fatal(err)
	}

	builder := flatbuffers.NewBuilder(1024)
	offset, err := requirement.ConvertToWLS(builder, "glibc")
	if err != nil {
		t.Fatal(err)
	}
	builder.Finish(offset)
	return wls.GetRootAsNodeTypeWrapper(builder.FinishedBytes(), 0)
}

func TestLibcPatchThresholdOrdering(t *testing.T) {
	tests := []struct {
		name                   string
		supported              bool
		major, minor, patch    int64
		wantGeneratedDenyMatch bool
	}{
		{name: "unsupported/earlier patch", major: 2, minor: 30, patch: 4, wantGeneratedDenyMatch: false},
		{name: "unsupported/exact patch", major: 2, minor: 30, patch: 5, wantGeneratedDenyMatch: true},
		{name: "unsupported/later patch", major: 2, minor: 30, patch: 6, wantGeneratedDenyMatch: true},
		{name: "unsupported/later minor", major: 2, minor: 31, patch: 0, wantGeneratedDenyMatch: true},
		{name: "supported/earlier patch", supported: true, major: 2, minor: 30, patch: 4, wantGeneratedDenyMatch: true},
		{name: "supported/exact patch", supported: true, major: 2, minor: 30, patch: 5, wantGeneratedDenyMatch: false},
		{name: "supported/later patch", supported: true, major: 2, minor: 30, patch: 6, wantGeneratedDenyMatch: false},
		{name: "supported/earlier minor", supported: true, major: 2, minor: 29, patch: 9, wantGeneratedDenyMatch: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := convertLibcRequirement(t, tt.supported)
			matched := evaluateLibcNode(t, root, libcRuntimeContext{
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

			if matched != tt.wantGeneratedDenyMatch {
				t.Fatalf("deny match = %t, want %t", matched, tt.wantGeneratedDenyMatch)
			}
		})
	}
}
