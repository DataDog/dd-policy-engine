package converter

import (
	"errors"
	"strings"

	"github.com/DataDog/dd-policy-engine/go/schema"
	"github.com/DataDog/dd-policy-engine/go/schema/dd/wls"

	flatbuffers "github.com/google/flatbuffers/go"
)

type JSONDeny struct {
	Id          string             `json:"id"`
	Description string             `json:"description"`
	Os          string             `json:"os"`
	Cmds        []CmdPattern       `json:"cmds"`
	Args        []ArgumentList     `json:"args"`
	Envs        map[string]*string `json:"envars"`
}

func isValidOS(os string) bool {
	return os == "windows" || os == "linux" || os == "darwin"
}

func (d JSONDeny) ConvertToWLS(builder *flatbuffers.Builder) (flatbuffers.UOffsetT, error) {
	var nodes []flatbuffers.UOffsetT

	if d.Os == "" && len(d.Cmds) == 0 && len(d.Args) == 0 && len(d.EnVars) == 0 {
		return 0, nil
	}

	if d.Os != "" && isValidOS(d.Os) {
		osEval := schema.StrEvaluatorCreate(builder, wls.StringEvaluatorsOS, d.Os, wls.CmpTypeSTRCMP_EXACT)
		osNode := schema.EvaluatorNodeCreate(builder, wls.EvaluatorTypeStrEvaluator, "OS matching", osEval)
		nodes = append(nodes, schema.NodeTypeWrapperCreate(builder, osNode, wls.NodeTypeEvaluatorNode))
	}

	// convert cmd patterns to evaluator nodes
	var cmdNodes []flatbuffers.UOffsetT
	for _, cmd := range d.Cmds {
		cmdNode, err := cmd.ConvertToWLS(builder)
		if err != nil {
			return 0, err
		}
		cmdNodes = append(cmdNodes, cmdNode)
	}

	if len(cmdNodes) == 1 {
		nodes = append(nodes, cmdNodes[0])
	} else if len(cmdNodes) > 1 {
		orNode := schema.CompositeNodeCreate(builder, wls.BoolOperationBOOL_OR, "Match any cmd pattern", cmdNodes)
		nodes = append(nodes, schema.NodeTypeWrapperCreate(builder, orNode, wls.NodeTypeCompositeNode))
	}

	// convert argument lists to evaluator nodes
	var argNodes []flatbuffers.UOffsetT
	for _, argumentList := range d.Args {
		argNode, err := argumentList.ConvertToWLS(builder)
		if err != nil {
			return 0, err
		}

		argNodes = append(argNodes, argNode)
	}

	if len(argNodes) == 1 {
		nodes = append(nodes, argNodes[0])
	} else if len(argNodes) > 1 {
		andNode := schema.CompositeNodeCreate(builder, wls.BoolOperationBOOL_AND, "Match all argument patterns", argNodes)
		nodes = append(nodes, schema.NodeTypeWrapperCreate(builder, andNode, wls.NodeTypeCompositeNode))
	}

	// if there is only one node, return it
	if len(nodes) == 1 {
		return nodes[0], nil
	}

	// if there are multiple nodes, combine them with AND
	rootNode := schema.CompositeNodeCreate(builder, wls.BoolOperationBOOL_AND, d.Description, nodes)
	return schema.NodeTypeWrapperCreate(builder, rootNode, wls.NodeTypeCompositeNode), nil
}
