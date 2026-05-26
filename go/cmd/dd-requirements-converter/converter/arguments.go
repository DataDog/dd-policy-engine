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
	case -2:
		return wls.StringEvaluatorsPROCESS_ARGV_N_2
	case -3:
		return wls.StringEvaluatorsPROCESS_ARGV_N_3
	case -4:
		return wls.StringEvaluatorsPROCESS_ARGV_N_4
	case -5:
		return wls.StringEvaluatorsPROCESS_ARGV_N_5
	case -6:
		return wls.StringEvaluatorsPROCESS_ARGV_N_6
	default:
		return wls.StringEvaluatorsPROCESS_ARGV
	}
}

// wildcardMatchToEvaluators returns one StrEvaluator per argument pattern: EXACT if there are no
// glob metacharacters, CMP_WILDCARD if the pattern contains * or ?. A pattern that is only "*" or "?" matches any value and returns offset 0.
func wildcardMatchToEvaluators(builder *flatbuffers.Builder, pattern string, position *int) (flatbuffers.UOffsetT, error) {
	if pattern == "*" || pattern == "?" {
		return 0, nil
	}

	ev := getArgvEvaluatorForPosition(position)
	var cmp wls.CmpTypeSTR
	if strings.ContainsAny(pattern, "*?") {
		cmp = wls.CmpTypeSTRCMP_WILDCARD
	} else {
		cmp = wls.CmpTypeSTRCMP_EXACT
	}

	strEvaluator := schema.StrEvaluatorCreate(builder, ev, pattern, cmp)
	node := schema.EvaluatorNodeCreate(builder, wls.EvaluatorTypeStrEvaluator, "Argument matching: "+pattern, strEvaluator, "")
	return schema.NodeTypeWrapperCreate(builder, node, wls.NodeTypeEvaluatorNode), nil
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
		if pos != nil && i > 0 {
			*pos++
		}

		argNode, err := wildcardMatchToEvaluators(builder, argument, pos)
		if err != nil {
			return 0, err
		}
		if argNode != 0 {
			nodes = append(nodes, argNode)
		}
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
	andNode := schema.CompositeNodeCreate(builder, wls.BoolOperationBOOL_AND, "Match argument pattern: "+strings.Join(a.Arguments, " "), nodes, "")
	return schema.NodeTypeWrapperCreate(builder, andNode, wls.NodeTypeCompositeNode), nil
}
