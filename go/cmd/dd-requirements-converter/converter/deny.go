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

func normalizeOS(os string) (string, bool) {
	switch os {
	case "windows", "linux":
		return os, true
	case "darwin", "macos":
		return "macos", true
	default:
		return "", false
	}
}

func (d JSONDeny) ConvertToWLS(builder *flatbuffers.Builder) (flatbuffers.UOffsetT, error) {
	var nodes []flatbuffers.UOffsetT

	if d.Os == "" && len(d.Cmds) == 0 && len(d.Args) == 0 && len(d.Envs) == 0 {
		return 0, errors.New("no conditions to match")
	}

	groups := 0
	for _, present := range []bool{d.Os != "", len(d.Cmds) > 0, len(d.Args) > 0, len(d.Envs) > 0} {
		if present {
			groups++
		}
	}
	soleGroupDescription := ""
	if groups == 1 {
		soleGroupDescription = d.Description
	}

	if d.Os != "" {
		os, ok := normalizeOS(d.Os)
		if !ok {
			return 0, errors.New("unknown operating system")
		}

		osEval := schema.StrEvaluatorCreate(builder, wls.StringEvaluatorsOS, os, wls.CmpTypeSTRCMP_EXACT)
		osNode := schema.EvaluatorNodeCreate(builder, wls.EvaluatorTypeStrEvaluator, soleGroupDescription, osEval)
		nodes = append(nodes, schema.NodeTypeWrapperCreate(builder, osNode, wls.NodeTypeEvaluatorNode))
	}

	// convert cmd patterns to evaluator nodes
	var cmdNodes []flatbuffers.UOffsetT
	cmdLeafDescription := ""
	if len(d.Cmds) == 1 {
		cmdLeafDescription = soleGroupDescription
	}
	for _, cmd := range d.Cmds {
		cmdNode, err := cmd.ConvertToWLS(builder, cmdLeafDescription)
		if err != nil {
			return 0, err
		}
		cmdNodes = append(cmdNodes, cmdNode)
	}

	if len(cmdNodes) == 1 {
		nodes = append(nodes, cmdNodes[0])
	} else if len(cmdNodes) > 1 {
		orNode := schema.CompositeNodeCreate(builder, wls.BoolOperationBOOL_OR, soleGroupDescription, cmdNodes)
		nodes = append(nodes, schema.NodeTypeWrapperCreate(builder, orNode, wls.NodeTypeCompositeNode))
	}

	// convert argument lists to evaluator nodes
	var argNodes []flatbuffers.UOffsetT
	argListDescription := ""
	if len(d.Args) == 1 {
		argListDescription = soleGroupDescription
	}
	for _, argumentList := range d.Args {
		argNode, err := argumentList.ConvertToWLS(builder, argListDescription)
		if err != nil {
			return 0, err
		}

		argNodes = append(argNodes, argNode)
	}

	if len(argNodes) == 1 {
		nodes = append(nodes, argNodes[0])
	} else if len(argNodes) > 1 {
		andNode := schema.CompositeNodeCreate(builder, wls.BoolOperationBOOL_AND, soleGroupDescription, argNodes)
		nodes = append(nodes, schema.NodeTypeWrapperCreate(builder, andNode, wls.NodeTypeCompositeNode))
	}

	var envNodes []flatbuffers.UOffsetT
	for key, value := range d.Envs {
		var kv string
		if value == nil {
			kv = key + "=*?"
		} else {
			kv = key + "=" + *value
		}

		var comparator wls.CmpTypeSTR

		if strings.ContainsAny(kv, "*?") {
			comparator = wls.CmpTypeSTRCMP_WILDCARD
		} else {
			comparator = wls.CmpTypeSTRCMP_EXACT
		}

		strEvaluator := schema.StrEvaluatorCreate(builder, wls.StringEvaluatorsPROCESS_ENVAR, kv, comparator)
		envDescription := ""
		if len(d.Envs) == 1 {
			envDescription = soleGroupDescription
		}
		node := schema.EvaluatorNodeCreate(builder, wls.EvaluatorTypeStrEvaluator, envDescription, strEvaluator)
		envNodes = append(envNodes, schema.NodeTypeWrapperCreate(builder, node, wls.NodeTypeEvaluatorNode))
	}

	if len(envNodes) == 1 {
		nodes = append(nodes, envNodes[0])
	} else if len(envNodes) > 1 {
		andNode := schema.CompositeNodeCreate(builder, wls.BoolOperationBOOL_AND, soleGroupDescription, envNodes)
		nodes = append(nodes, schema.NodeTypeWrapperCreate(builder, andNode, wls.NodeTypeCompositeNode))
	}

	var root flatbuffers.UOffsetT
	// if there is only one node, use it directly (it's already a NodeTypeWrapper)
	if len(nodes) == 1 {
		root = nodes[0]
	} else if len(nodes) > 1 {
		andNode := schema.CompositeNodeCreate(builder, wls.BoolOperationBOOL_AND, d.Description, nodes)
		root = schema.NodeTypeWrapperCreate(builder, andNode, wls.NodeTypeCompositeNode)
	}

	return root, nil
}
