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

func wildcardMatchToEvaluators(builder *flatbuffers.Builder, pattern string) (flatbuffers.UOffsetT, error) {
	if !strings.Contains(pattern, "*") {
		strEvaluator := schema.StrEvaluatorCreate(builder, wls.StringEvaluatorsPROCESS_ARGV, pattern, wls.CmpTypeSTRCMP_EXACT)
		node := schema.EvaluatorNodeCreate(builder, wls.EvaluatorTypeStrEvaluator, "Argument matching: "+pattern, strEvaluator)
		return schema.NodeTypeWrapperCreate(builder, node, wls.NodeTypeEvaluatorNode), nil
	}

	var nodes []flatbuffers.UOffsetT

	// split by *
	parts := strings.Split(pattern, "*")

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

		strEvaluator := schema.StrEvaluatorCreate(builder, wls.StringEvaluatorsPROCESS_ARGV, part, cmp)
		node := schema.EvaluatorNodeCreate(builder, wls.EvaluatorTypeStrEvaluator, "Argument matching: "+part, strEvaluator)
		nodes = append(nodes, schema.NodeTypeWrapperCreate(builder, node, wls.NodeTypeEvaluatorNode))
	}

	if len(nodes) == 1 {
		return nodes[0], nil
	}

	andNode := schema.CompositeNodeCreate(builder, wls.BoolOperationBOOL_AND, "Match argument pattern: "+pattern, nodes)
	return schema.NodeTypeWrapperCreate(builder, andNode, wls.NodeTypeCompositeNode), nil
}

func (a ArgumentList) ConvertToWLS(builder *flatbuffers.Builder) (flatbuffers.UOffsetT, error) {
	if len(a.Arguments) == 0 {
		return 0, nil
	}

	var nodes []flatbuffers.UOffsetT

	for _, argument := range a.Arguments {
		if argument == "" {
			continue
		}

		argNode, err := wildcardMatchToEvaluators(builder, argument)
		if err != nil {
			return 0, err
		}

		nodes = append(nodes, argNode)
	}

	if len(nodes) == 1 {
		return nodes[0], nil
	}

	andNode := schema.CompositeNodeCreate(builder, wls.BoolOperationBOOL_AND, "Match argument pattern: "+strings.Join(a.Arguments, " "), nodes)
	return schema.NodeTypeWrapperCreate(builder, andNode, wls.NodeTypeCompositeNode), nil
}
