package converter

import (
	"errors"
	"strings"

	"github.com/DataDog/dd-policy-engine/go/schema"
	"github.com/DataDog/dd-policy-engine/go/schema/dd/wls"

	flatbuffers "github.com/google/flatbuffers/go"
)

type ArgumentList struct {
	Arguments []string `json:"args"`
	Position  *int     `json:"position"`
}

func getArgvEvaluatorForPosition(position *int) wls.StringEvaluators {
	if position == nil {
		return wls.StringEvaluatorsPROCESS_ARGV
	}

	switch *position {
	case 0:
		return wls.StringEvaluatorsPROCESS_ARGV_0
	case 1:
		return wls.StringEvaluatorsPROCESS_ARGV_1
	case 2:
		return wls.StringEvaluatorsPROCESS_ARGV_2
	case 3:
		return wls.StringEvaluatorsPROCESS_ARGV_3
	case 4:
		return wls.StringEvaluatorsPROCESS_ARGV_4
	case 5:
		return wls.StringEvaluatorsPROCESS_ARGV_5
	case -1:
		return wls.StringEvaluatorsPROCESS_ARGV_N
	}

	return wls.StringEvaluatorsPROCESS_ARGV
}

// if the pattern is a single string, return an evaluator node
// if the pattern contains wildcards, return a composite node with parts of the pattern as evaluator nodes
func wildcardMatchToEvaluators(builder *flatbuffers.Builder, pattern string, position *int) (flatbuffers.UOffsetT, error) {
	if pattern == "*" || pattern == "?" {
		return 0, nil
	}

	if !strings.Contains(pattern, "*") && !strings.Contains(pattern, "?") {
		strEvaluator := schema.StrEvaluatorCreate(builder, getArgvEvaluatorForPosition(position), pattern, wls.CmpTypeSTRCMP_EXACT)
		node := schema.EvaluatorNodeCreate(builder, wls.EvaluatorTypeStrEvaluator, "Argument matching: "+pattern, strEvaluator)
		return schema.NodeTypeWrapperCreate(builder, node, wls.NodeTypeEvaluatorNode), nil
	}

	var nodes []flatbuffers.UOffsetT

	// split by * and ? (normalize ? to * so both act as wildcards)
	normalized := strings.ReplaceAll(pattern, "?", "*")
	parts := strings.Split(normalized, "*")

	for i, part := range parts {
		if part == "" {
			continue
		}

		isFirst := i == 0
		isLast := i == len(parts)-1

		var cmp wls.CmpTypeSTR
		switch {
		case isFirst:
			cmp = wls.CmpTypeSTRCMP_PREFIX
		case isLast:
			cmp = wls.CmpTypeSTRCMP_SUFFIX
		default:
			cmp = wls.CmpTypeSTRCMP_CONTAINS
		}

		strEvaluator := schema.StrEvaluatorCreate(builder, getArgvEvaluatorForPosition(position), part, cmp)

		node := schema.EvaluatorNodeCreate(builder, wls.EvaluatorTypeStrEvaluator, "Argument matching for pattern: "+part, strEvaluator)
		nodes = append(nodes, schema.NodeTypeWrapperCreate(builder, node, wls.NodeTypeEvaluatorNode))
	}

	if len(nodes) == 1 {
		return nodes[0], nil
	}

	andNode := schema.CompositeNodeCreate(builder, wls.BoolOperationBOOL_AND, "Match argument pattern: "+pattern, nodes)
	return schema.NodeTypeWrapperCreate(builder, andNode, wls.NodeTypeCompositeNode), nil
}

// return a NodeTypeWrapper with an evaluator node or a composite node with the argument patterns
func (a ArgumentList) ConvertToWLS(builder *flatbuffers.Builder) (flatbuffers.UOffsetT, error) {
	if len(a.Arguments) == 0 {
		return 0, errors.New("no arguments provided")
	}

	// nodes is a list of evaluator nodes or a composite node with the argument patterns
	var nodes []flatbuffers.UOffsetT

	var pos *int
	if a.Position != nil {
		localPos := *a.Position
		pos = &localPos
	}

	for i, argument := range a.Arguments {
		if argument == "" {
			continue
		}

		// increment position for arguments after the first
		if pos != nil && i > 0 {
			(*pos)++
		}

		argNode, err := wildcardMatchToEvaluators(builder, argument, pos)
		if err != nil {
			return 0, err
		}

		// Skip pure wildcards (they match anything)
		if argNode == 0 {
			continue
		}

		nodes = append(nodes, argNode)
	}

	if len(nodes) == 0 {
		return 0, errors.New("pattern has no matchable parts")
	}

	if len(nodes) == 1 {
		return nodes[0], nil
	}

	// combine multiple evaluators for separate arguments
	// e.g. { "args": ["-version", "1.*"], "position": 1 }
	// will be converted to:
	// AND(
	// EvaluatorNode: arg matches "-version" at position 1
	// EvaluatorNode: arg matches "1.*" at position 2
	// )
	andNode := schema.CompositeNodeCreate(builder, wls.BoolOperationBOOL_AND, "Match argument pattern: "+strings.Join(a.Arguments, " "), nodes)
	return schema.NodeTypeWrapperCreate(builder, andNode, wls.NodeTypeCompositeNode), nil
}
