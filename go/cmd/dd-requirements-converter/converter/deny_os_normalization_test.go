package converter

import (
	"testing"

	"github.com/DataDog/dd-policy-engine/go/schema/dd/wls"

	flatbuffers "github.com/google/flatbuffers/go"
)

func convertedOSValue(t *testing.T, input string) string {
	t.Helper()

	builder := flatbuffers.NewBuilder(128)
	offset, err := (JSONDeny{Os: input}).ConvertToWLS(builder)
	if err != nil {
		t.Fatal(err)
	}
	builder.Finish(offset)

	root := wls.GetRootAsNodeTypeWrapper(builder.FinishedBytes(), 0)
	if root.NodeType() != wls.NodeTypeEvaluatorNode {
		t.Fatalf("unexpected root node type %s", root.NodeType())
	}

	var table flatbuffers.Table
	if !root.Node(&table) {
		t.Fatal("root node has no value")
	}
	var node wls.EvaluatorNode
	node.Init(table.Bytes, table.Pos)
	if node.EvalType() != wls.EvaluatorTypeStrEvaluator || !node.Eval(&table) {
		t.Fatalf("unexpected evaluator type %s", node.EvalType())
	}

	var evaluator wls.StrEvaluator
	evaluator.Init(table.Bytes, table.Pos)
	if evaluator.Id() != wls.StringEvaluatorsOS {
		t.Fatalf("unexpected string evaluator %s", evaluator.Id())
	}
	return string(evaluator.Value())
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

func TestDenyOSUsesCanonicalMacOS(t *testing.T) {
	for _, input := range []string{"darwin", "macos"} {
		t.Run(input, func(t *testing.T) {
			if got := convertedOSValue(t, input); got != "macos" {
				t.Fatalf("converted OS = %q, want macos", got)
			}
		})
	}
}
