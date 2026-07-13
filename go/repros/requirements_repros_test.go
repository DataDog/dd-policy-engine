//go:build repro

package repros

import (
	"encoding/json"
	"os"
	"os/exec"
	"testing"

	"github.com/DataDog/dd-policy-engine/go/cmd/dd-requirements-converter/converter"
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
	var requirement converter.JSONlibc
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

	matched := evaluateNode(t, root, runtimeContext{
		strings: map[wls.StringEvaluators]string{
			wls.StringEvaluatorsMACHINE_ARCHITECTURE: "x86_64",
			wls.StringEvaluatorsLIBC_FLAVOR:          "glibc",
		},
		numerics: map[wls.NumericEvaluators]int64{
			wls.NumericEvaluatorsLIBC_VERSION_MAJOR: 2,
			wls.NumericEvaluatorsLIBC_VERSION_MINOR: 30,
			wls.NumericEvaluatorsLIBC_VERSION_PATCH: 4,
		},
	})

	if matched {
		t.Fatal("2.30.4 matched an unsupported >= 2.30.5 rule")
	}
}

func TestInvalidDenyOSCannotBroadenRule(t *testing.T) {
	builder := flatbuffers.NewBuilder(128)
	offset, err := (converter.JSONDeny{
		Os:   "freebsd",
		Cmds: []converter.CmdPattern{"/usr/bin/curl"},
	}).ConvertToWLS(builder)
	if err == nil {
		t.Fatalf("invalid OS was dropped, broadening the command deny rule at offset %d", offset)
	}
}

func TestDarwinDenyMatchesCanonicalMacOS(t *testing.T) {
	builder := flatbuffers.NewBuilder(128)
	offset, err := (converter.JSONDeny{Os: "darwin"}).ConvertToWLS(builder)
	if err != nil {
		t.Fatal(err)
	}
	builder.Finish(offset)
	root := wls.GetRootAsNodeTypeWrapper(builder.FinishedBytes(), 0)

	if !evaluateNode(t, root, runtimeContext{
		strings: map[wls.StringEvaluators]string{wls.StringEvaluatorsOS: "macos"},
	}) {
		t.Fatal("darwin requirement did not match the canonical macos runtime value")
	}
}

func TestEmptyLibcVersionReturnsError(t *testing.T) {
	const helper = "DD_REPRO_EMPTY_LIBC_VERSION"
	if os.Getenv(helper) == "1" {
		var requirement converter.JSONlibc
		if err := json.Unmarshal([]byte(`{"arch":"x64","supported":true,"min":""}`), &requirement); err != nil {
			os.Exit(2)
		}
		_, err := requirement.ConvertToWLS(flatbuffers.NewBuilder(128), "glibc")
		if err != nil {
			os.Exit(0)
		}
		os.Exit(3)
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestEmptyLibcVersionReturnsError$")
	cmd.Env = append(os.Environ(), helper+"=1")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("empty version terminated the converter process: %v\n%s", err, output)
	}
}
