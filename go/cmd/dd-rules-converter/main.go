package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/DataDog/dd-policy-engine/go/internal/parser"
	"github.com/DataDog/dd-policy-engine/go/schema"
	wls "github.com/DataDog/dd-policy-engine/go/schema/dd/wls"
	flatbuffers "github.com/google/flatbuffers/go"
)

func writeBufferToFile(buffer []byte, fileName string) {
	err := os.WriteFile(fileName, buffer, 0644)
	if err != nil {
		log.Fatalf("Failed to write buffer to file: %v", err)
	}
	fmt.Printf("Wrote %d bytes to: %s\n", len(buffer), fileName)
}

type Rule struct {
	Description string `toml:"description"`
	Expression  string `toml:"expression"`
}

// Config is a map of rule names to their Rule definitions
type Rules map[string]Rule

type PolicyBuilder struct {
	builder *flatbuffers.Builder
	offsets []flatbuffers.UOffsetT
}

func NewPolicyBuilder() *PolicyBuilder {
	return &PolicyBuilder{
		builder: flatbuffers.NewBuilder(1024),
	}
}

func main() {
	rulesFile := flag.String("rules", "", "TOML rule files")
	outputFile := flag.String("output", "policy.fb", "Location of the generated policy")
	flag.Parse()

	if *rulesFile == "" {
		fmt.Fprintln(os.Stderr, "error: -rules flag is required")
		flag.Usage()
	}

	if _, err := os.Stat(*rulesFile); err != nil {
		log.Fatalf("failed to access rules file %q: %v", *rulesFile, err)
	}

	var rules Rules
	_, err := toml.DecodeFile(*rulesFile, &rules)
	if err != nil {
		log.Fatalf("failed to decode rules file %q: %v", *rulesFile, err)
	}

	builder := NewPolicyBuilder()

	n_rules := len(rules)
	fmt.Printf("Found %d rules defined in %s\n", n_rules, *rulesFile)

	i := 0
	for id, rule := range rules {
		i += 1
		fmt.Printf("[%d/%d] Parsing rule \"%s\" \n", i, n_rules, id)
		ruleAst, err := parser.Parse(rule.Expression)
		if err != nil {
			fmt.Printf("Parsing error: \n%s\n", err)
			os.Exit(1)
		}

		if err = addRule(builder, id, rule.Description, ruleAst); err != nil {
			fmt.Printf("Error: \n%s", err)
			os.Exit(1)
		}
	}

	policies := schema.PoliciesCreate(builder.builder, builder.offsets)
	builder.builder.Finish(policies)
	buffer := builder.builder.FinishedBytes()
	writeBufferToFile(buffer, *outputFile)
	os.Exit(0)
}

func createStrEvaluatorNode(builder *flatbuffers.Builder, evaluatorId wls.StringEvaluators, value string, cmpp wls.CmpTypeSTR, description string) flatbuffers.UOffsetT {
	evaluator := schema.StrEvaluatorCreate(builder, evaluatorId, value, cmpp)
	node := schema.EvaluatorNodeCreate(builder, wls.EvaluatorTypeStrEvaluator, description, evaluator)
	return schema.NodeTypeWrapperCreate(builder, node, wls.NodeTypeEvaluatorNode)
}

func toStrComparator(cmp parser.Comparator) wls.CmpTypeSTR {
	switch cmp {
	case parser.Exact:
		return wls.CmpTypeSTRCMP_EXACT
	case parser.Prefix:
		return wls.CmpTypeSTRCMP_PREFIX
	case parser.Suffix:
		return wls.CmpTypeSTRCMP_SUFFIX
	case parser.Contains:
		return wls.CmpTypeSTRCMP_CONTAINS
	}
	panic("unsupported comparator")
}

func createNode(builder *flatbuffers.Builder, node *parser.TermNode) (flatbuffers.UOffsetT, error) {
	cmp := toStrComparator(node.Comparator)

	switch node.Field {
	case "process.name":
		return createStrEvaluatorNode(builder, wls.StringEvaluatorsPROCESS_EXE, node.Value, cmp, "Check process is "+node.Value), nil
	case "process.dll":
		return createStrEvaluatorNode(builder, wls.StringEvaluatorsRUNTIME_ENTRY_POINT_FILE, node.Value, cmp, "Check process DLL is "+node.Value), nil
	case "runtime":
		return createStrEvaluatorNode(builder, wls.StringEvaluatorsRUNTIME_LANGUAGE, node.Value, cmp, "Check runtime is "+node.Value), nil
	case "iis-app-pool":
		return createStrEvaluatorNode(builder, wls.StringEvaluatorsIIS_APPLICATION_POOL, node.Value, cmp, "Check IIS application pool is "+node.Value), nil
	}

	return 0, fmt.Errorf("unsupported field \"%s\"", node.Field)
}

func createConditionalNode(builder *flatbuffers.Builder, oper wls.BoolOperation, description string, nodes []flatbuffers.UOffsetT) flatbuffers.UOffsetT {
	nodeRoot := schema.CompositeNodeCreate(builder, oper, description, nodes)
	return schema.NodeTypeWrapperCreate(builder, nodeRoot, wls.NodeTypeCompositeNode)
}

func processAST(nodes *[]flatbuffers.UOffsetT, builder *flatbuffers.Builder, node parser.Node) error {
	switch n := node.(type) {
	case *parser.TermNode:
		if leafNode, err := createNode(builder, n); err != nil {
			return err
		} else {
			*nodes = append(*nodes, leafNode)
			return nil
		}

	case *parser.BooleanNode:
		var nn []flatbuffers.UOffsetT
		err := processAST(&nn, builder, n.Left)
		if err != nil {
			return err
		}

		err = processAST(&nn, builder, n.Right)
		if err != nil {
			return err
		}

		var operator wls.BoolOperation
		switch n.Kind {
		case parser.Or:
			operator = wls.BoolOperationBOOL_OR
		case parser.And:
			operator = wls.BoolOperationBOOL_AND
		default:
			return fmt.Errorf("unsupported operator \"%s\"", n.Kind)

		}

		andNode := createConditionalNode(builder, operator, "", nn)
		*nodes = append(*nodes, andNode)
		return nil

	case *parser.UnaryBooleanNode:
		var nn []flatbuffers.UOffsetT
		err := processAST(&nn, builder, n.Expr)
		if err != nil {
			return err
		}

		var operator wls.BoolOperation
		switch n.Kind {
		case parser.Not:
			operator = wls.BoolOperationBOOL_NOT
		default:
			return fmt.Errorf("unsupported operator \"%s\"", n.Kind)
		}

		compNode := createConditionalNode(builder, operator, "", nn)
		*nodes = append(*nodes, compNode)
		return nil

	}

	return errors.New("unexpected node type")
}

func addRule(builder *PolicyBuilder, id string, description string, ast parser.AST) error {
	var nodes []flatbuffers.UOffsetT

	// Create tree
	if err := processAST(&nodes, builder.builder, ast.Node); err != nil {
		fmt.Printf("Error: %s\n", err)
		return err
	}

	// Create root and action nodes
	nodeRoot := createConditionalNode(builder.builder, wls.BoolOperationBOOL_AND, id, nodes)
	action := schema.ActionCreate(builder.builder, wls.ActionIdINJECT_ALLOW, id, []string{})

	builder.offsets = append(builder.offsets, schema.PolicyCreate(builder.builder, description, nodeRoot, []flatbuffers.UOffsetT{action}))
	return nil
}
