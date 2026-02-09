//go:build ignore

package main

import (
	"fmt"
	"os"
	"strings"

	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/DataDog/dd-policy-engine/go/schema/dd/wls"
)

func printNode(wrapper *wls.NodeTypeWrapper, indent string) {
	nodeType := wrapper.NodeType()
	fmt.Printf("%sNodeType: %s\n", indent, nodeType.String())

	switch nodeType {
	case wls.NodeTypeEvaluatorNode:
		var table flatbuffers.Table
		if wrapper.Node(&table) {
			var evalNode wls.EvaluatorNode
			evalNode.Init(table.Bytes, table.Pos)
			
			fmt.Printf("%s  Description: %s\n", indent, string(evalNode.Description()))
			fmt.Printf("%s  EvalType: %s\n", indent, evalNode.EvalType().String())
			
			// If it's a string evaluator, print details
			if evalNode.EvalType() == wls.EvaluatorTypeStrEvaluator {
				var evalTable flatbuffers.Table
				if evalNode.Eval(&evalTable) {
					var strEval wls.StrEvaluator
					strEval.Init(evalTable.Bytes, evalTable.Pos)
					fmt.Printf("%s  Field: %s\n", indent, strEval.Id().String())
					fmt.Printf("%s  Compare: %s\n", indent, strEval.Cmp().String())
					fmt.Printf("%s  Value: %s\n", indent, string(strEval.Value()))
				}
			} else if evalNode.EvalType() == wls.EvaluatorTypeNumEvaluator {
				var evalTable flatbuffers.Table
				if evalNode.Eval(&evalTable) {
					var numEval wls.NumEvaluator
					numEval.Init(evalTable.Bytes, evalTable.Pos)
					fmt.Printf("%s  Field: %s\n", indent, numEval.Id().String())
					fmt.Printf("%s  Compare: %s\n", indent, numEval.Cmp().String())
					fmt.Printf("%s  Value: %d\n", indent, numEval.Value())
				}
			}
		}

	case wls.NodeTypeCompositeNode:
		var table flatbuffers.Table
		if wrapper.Node(&table) {
			var composite wls.CompositeNode
			composite.Init(table.Bytes, table.Pos)
			
			fmt.Printf("%s  Description: %s\n", indent, string(composite.Description()))
			fmt.Printf("%s  Operation: %s\n", indent, composite.Op().String())
			fmt.Printf("%s  Children (%d):\n", indent, composite.ChildrenLength())
			
			for i := 0; i < composite.ChildrenLength(); i++ {
				var child wls.NodeTypeWrapper
				if composite.Children(&child, i) {
					fmt.Printf("%s  [%d]:\n", indent, i)
					printNode(&child, indent+"    ")
				}
			}
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run deserialize_fb.go <file.bin>")
		os.Exit(1)
	}

	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	policies := wls.GetRootAsPolicies(data, 0)
	fmt.Printf("File: %s (%d bytes)\n", os.Args[1], len(data))
	fmt.Printf("Number of policies: %d\n", policies.PoliciesLength())
	fmt.Println(strings.Repeat("=", 60))

	for i := 0; i < policies.PoliciesLength(); i++ {
		var policy wls.Policy
		if policies.Policies(&policy, i) {
			fmt.Printf("\nPolicy %d: %s\n", i, string(policy.Description()))
			fmt.Println(strings.Repeat("-", 40))

			// Print actions
			fmt.Printf("Actions (%d):\n", policy.ActionsLength())
			for j := 0; j < policy.ActionsLength(); j++ {
				var action wls.Action
				if policy.Actions(&action, j) {
					fmt.Printf("  [%d] %s: %s\n", j, action.Action().String(), string(action.Description()))
				}
			}

			// Print rules tree
			fmt.Println("\nRules:")
			if rules := policy.Rules(nil); rules != nil {
				printNode(rules, "  ")
			} else {
				fmt.Println("  (none)")
			}
		}
	}
}

