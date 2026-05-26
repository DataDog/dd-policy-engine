package main

import (
  "sort"
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

type Rule struct {
	Description string `toml:"description"`
  Instrument bool `toml:"instrument"`
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

func run(data string) ([]byte, error) {
	var rules Rules
	meta, err := toml.Decode(data, &rules)
	if err != nil {
		log.Printf("failed to decode rules: %v", err)
    return nil, err
	}

  // Sort to create rules in a deterministic order to ensure the generated flatbuffer is consistent for the same rule file.
  rulesId := make([]string, 0, len(rules))
	for k := range rules {
		rulesId = append(rulesId, k)
	}
	sort.Strings(rulesId)

	builder := NewPolicyBuilder()

	n_rules := len(rules)
	log.Printf("Found %d rules", n_rules)

	i := 0
	for _, id := range rulesId {
		i += 1
    rule := rules[id]
    log.Printf("[%d/%d] Processing rule %q", i, n_rules, id)
    log.Printf("  - Validating...")
    err = validateRule(meta, id)
    if err != nil {
      log.Printf("  Error: %s", err)
      return nil, err
    }

		log.Printf("  - Parsing...")
		ruleAst, err := parser.Parse(rule.Expression)
		if err != nil {
			log.Printf("  Parsing error: %s", err)
      return nil, err
		}

		if err = addRule(builder, id, rule.Description, ruleAst, rule.Instrument); err != nil {
			log.Printf("  Error: %s", err)
      return nil, err
		}
	}

	policies := schema.PoliciesCreate(builder.builder, builder.offsets)
	builder.builder.Finish(policies)
	buffer := builder.builder.FinishedBytes()
  return buffer, nil
}

func main() {
	ruleFile := flag.String("rules", "", "TOML rule file")
	outputFile := flag.String("output", "policy.fb", "Location of the generated policy. Default to ./policy.fb")
	flag.Parse()

	if *ruleFile == "" {
		fmt.Fprintln(os.Stderr, "error: -rules flag is required")
		flag.Usage()
	}

  log.Printf("Reading %q", *ruleFile)
  fileContent, err := os.ReadFile(*ruleFile)
	if err != nil {
		log.Fatalf("failed to access rules file %q: %v", *ruleFile, err)
	}

  buffer, err := run(string(fileContent))
  if err != nil {
    os.Exit(1)
  }

	err = os.WriteFile(*outputFile, buffer, 0644)
	if err != nil {
		log.Fatalf("Failed to write buffer to file: %v", err)
	}
	log.Printf("Wrote %d bytes to: %s", len(buffer), *outputFile)

  log.Printf("Setting permissions on: %s", *outputFile)
  err = SetPermissions(*outputFile)
  if err != nil {
    log.Fatalf("Failed to set permissions to file: %v", err)
  }

	os.Exit(0)
}

func validateRule(meta toml.MetaData, ruleId string) error {
  requiredFields := []string{ "expression", "instrument" }

  for _, fieldName := range requiredFields {
    if !meta.IsDefined(ruleId, fieldName) {
      return fmt.Errorf("mandatory field %q is missing from rule %q", fieldName, ruleId)
    }
  }

  return nil
}

func createStrEvaluatorNode(builder *flatbuffers.Builder, evaluatorId wls.StringEvaluators, value string, cmpp wls.CmpTypeSTR, description string) flatbuffers.UOffsetT {
	evaluator := schema.StrEvaluatorCreate(builder, evaluatorId, value, cmpp)
	node := schema.EvaluatorNodeCreate(builder, wls.EvaluatorTypeStrEvaluator, description, evaluator, "")
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
	case "process.executable":
		return createStrEvaluatorNode(builder, wls.StringEvaluatorsPROCESS_EXE, node.Value, cmp, "Check process is "+node.Value), nil
	case "dotnet.dll":
		return createStrEvaluatorNode(builder, wls.StringEvaluatorsRUNTIME_ENTRY_POINT_FILE, node.Value, cmp, "Check process DLL is "+node.Value), nil
	case "runtime.language":
		return createStrEvaluatorNode(builder, wls.StringEvaluatorsRUNTIME_LANGUAGE, node.Value, cmp, "Check runtime is "+node.Value), nil
	case "iis.application_pool":
		return createStrEvaluatorNode(builder, wls.StringEvaluatorsIIS_APPLICATION_POOL, node.Value, cmp, "Check IIS application pool is "+node.Value), nil
	}

	return 0, fmt.Errorf("unsupported field \"%s\"", node.Field)
}

func createConditionalNode(builder *flatbuffers.Builder, oper wls.BoolOperation, description string, nodes []flatbuffers.UOffsetT) flatbuffers.UOffsetT {
	nodeRoot := schema.CompositeNodeCreate(builder, oper, description, nodes, "")
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

func addRule(builder *PolicyBuilder, id string, description string, ast parser.AST, enableInstrumentation bool) error {
	var nodes []flatbuffers.UOffsetT

	// Create tree
	if err := processAST(&nodes, builder.builder, ast.Node); err != nil {
		return err
	}

	// Create root and action nodes
  actionKind := wls.ActionIdINJECT_DENY
  if (enableInstrumentation) {
    actionKind = wls.ActionIdINJECT_ALLOW
  }

	nodeRoot := createConditionalNode(builder.builder, wls.BoolOperationBOOL_AND, id, nodes)
	action := schema.ActionCreate(builder.builder, actionKind, id, []string{})

	builder.offsets = append(builder.offsets, schema.PolicyCreate(builder.builder, description, nodeRoot, []flatbuffers.UOffsetT{action}))
	return nil
}
