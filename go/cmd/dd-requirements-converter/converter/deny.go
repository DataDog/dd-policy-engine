package converter

import (
	"errors"

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

	if d.Os == "" && len(d.Cmds) == 0 && len(d.Args) == 0 && len(d.Envs) == 0 {
		return 0, errors.New("no conditions to match")
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

	var envNodes []flatbuffers.UOffsetT
	for key, value := range d.Envs {
		var kv string
		if value == nil {
			kv = key + "="
		} else {
			kv = key + "=" + *value
		}

		strEvaluator := schema.StrEvaluatorCreate(builder, wls.StringEvaluatorsPROCESS_ENVAR, kv, wls.CmpTypeSTRCMP_EXACT)
		node := schema.EvaluatorNodeCreate(builder, wls.EvaluatorTypeStrEvaluator, "Environment variable matching: "+kv, strEvaluator)
		envNodes = append(envNodes, schema.NodeTypeWrapperCreate(builder, node, wls.NodeTypeEvaluatorNode))
	}

	if len(envNodes) == 1 {
		nodes = append(nodes, envNodes[0])
	} else if len(envNodes) > 1 {
		andNode := schema.CompositeNodeCreate(builder, wls.BoolOperationBOOL_AND, "Match all environment variable patterns", envNodes)
		nodes = append(nodes, schema.NodeTypeWrapperCreate(builder, andNode, wls.NodeTypeCompositeNode))
	}

	action := schema.ActionCreate(builder, wls.ActionIdINJECT_DENY, d.Description, nil)

	var root flatbuffers.UOffsetT
	// if there is only one node, return a policy with one EvaluatorNode and the deny action
	if len(nodes) == 1 {
		root = schema.NodeTypeWrapperCreate(builder, nodes[0], wls.NodeTypeEvaluatorNode)
	} else if len(nodes) > 1 {
		andNode := schema.CompositeNodeCreate(builder, wls.BoolOperationBOOL_AND, d.Description, nodes)
		root = schema.NodeTypeWrapperCreate(builder, andNode, wls.NodeTypeCompositeNode)
	}

	return schema.PolicyCreate(builder, d.Description, root, []flatbuffers.UOffsetT{action}), nil
}
