//go:build ignore

package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/DataDog/dd-policy-engine/go/schema/dd/wls"
)

var output io.Writer = os.Stdout

func printNode(wrapper *wls.NodeTypeWrapper, indent string) {
	nodeType := wrapper.NodeType()
	fmt.Fprintf(output, "%sNodeType: %s\n", indent, nodeType.String())

	switch nodeType {
	case wls.NodeTypeEvaluatorNode:
		var table flatbuffers.Table
		if wrapper.Node(&table) {
			var evalNode wls.EvaluatorNode
			evalNode.Init(table.Bytes, table.Pos)
			
			fmt.Fprintf(output, "%s  Description: %s\n", indent, string(evalNode.Description()))
			fmt.Fprintf(output, "%s  EvalType: %s\n", indent, evalNode.EvalType().String())
			
			// If it's a string evaluator, print details
			if evalNode.EvalType() == wls.EvaluatorTypeStrEvaluator {
				var evalTable flatbuffers.Table
				if evalNode.Eval(&evalTable) {
					var strEval wls.StrEvaluator
					strEval.Init(evalTable.Bytes, evalTable.Pos)
					fmt.Fprintf(output, "%s  Field: %s\n", indent, strEval.Id().String())
					fmt.Fprintf(output, "%s  Compare: %s\n", indent, strEval.Cmp().String())
					fmt.Fprintf(output, "%s  Value: %s\n", indent, string(strEval.Value()))
				}
			} else if evalNode.EvalType() == wls.EvaluatorTypeNumEvaluator {
				var evalTable flatbuffers.Table
				if evalNode.Eval(&evalTable) {
					var numEval wls.NumEvaluator
					numEval.Init(evalTable.Bytes, evalTable.Pos)
					fmt.Fprintf(output, "%s  Field: %s\n", indent, numEval.Id().String())
					fmt.Fprintf(output, "%s  Compare: %s\n", indent, numEval.Cmp().String())
					fmt.Fprintf(output, "%s  Value: %d\n", indent, numEval.Value())
				}
			}
		}

	case wls.NodeTypeCompositeNode:
		var table flatbuffers.Table
		if wrapper.Node(&table) {
			var composite wls.CompositeNode
			composite.Init(table.Bytes, table.Pos)
			
			fmt.Fprintf(output, "%s  Description: %s\n", indent, string(composite.Description()))
			fmt.Fprintf(output, "%s  Operation: %s\n", indent, composite.Op().String())
			fmt.Fprintf(output, "%s  Children (%d):\n", indent, composite.ChildrenLength())
			
			for i := 0; i < composite.ChildrenLength(); i++ {
				var child wls.NodeTypeWrapper
				if composite.Children(&child, i) {
					fmt.Fprintf(output, "%s  [%d]:\n", indent, i)
					printNode(&child, indent+"    ")
				}
			}
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run deserialize_fb.go <file.bin> [output.txt]")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	
	// Determine output file: use second arg, or derive from input (.bin -> .txt)
	var outputFile string
	if len(os.Args) >= 3 {
		outputFile = os.Args[2]
	} else {
		outputFile = strings.TrimSuffix(inputFile, ".bin") + ".txt"
	}

	data, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Create output file
	f, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()
	output = f

	policies := wls.GetRootAsPolicies(data, 0)
	fmt.Fprintf(output, "File: %s (%d bytes)\n", inputFile, len(data))
	fmt.Fprintf(output, "Number of policies: %d\n", policies.PoliciesLength())
	fmt.Fprintln(output, strings.Repeat("=", 60))

	for i := 0; i < policies.PoliciesLength(); i++ {
		var policy wls.Policy
		if policies.Policies(&policy, i) {
			fmt.Fprintf(output, "\nPolicy %d: %s\n", i, string(policy.Description()))
			fmt.Fprintln(output, strings.Repeat("-", 40))

			// Print actions
			fmt.Fprintf(output, "Actions (%d):\n", policy.ActionsLength())
			for j := 0; j < policy.ActionsLength(); j++ {
				var action wls.Action
				if policy.Actions(&action, j) {
					fmt.Fprintf(output, "  [%d] %s: %s\n", j, action.Action().String(), string(action.Description()))
				}
			}

			// Print rules tree
			fmt.Fprintln(output, "\nRules:")
			if rules := policy.Rules(nil); rules != nil {
				printNode(rules, "  ")
			} else {
				fmt.Fprintln(output, "  (none)")
			}
		}
	}

	fmt.Printf("Output written to: %s\n", outputFile)
}

