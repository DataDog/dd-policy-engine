package converter

import (
	"strings"
	"errors"
	"log"

	"github.com/DataDog/dd-policy-engine/go/schema"
	"github.com/DataDog/dd-policy-engine/go/schema/dd/wls"

	flatbuffers "github.com/google/flatbuffers/go"
)

type JSONlibc struct {
	Arch        string `json:"arch"`
	Description string `json:"description"`
	Supported   bool   `json:"supported"`
	Min         string `json:"min"`
}

func getActionId(supported bool) wls.ActionId {
	if supported {
		return wls.ActionIdINJECT_ALLOW
	}
	return wls.ActionIdINJECT_DENY
}

func parseSemver(version string) (int64, int64, int64, error) {
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return 0, 0, 0, errors.New("invalid version format: " + version)
	}

	major, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, 0, errors.New("invalid major version")
	}

	minor, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, 0, errors.New("invalid minor version")
	}

	if len(parts) < 3 {
		return major, minor, 0, nil
	}

	patch, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return 0, 0, 0, errors.New("invalid patch version")
	}
	
	return major, minor, patch, nil
}

// ConvertToWLS converts a JSONGlibc requirement to a WLS policy
func (l JSONlibc) ConvertToWLS(builder *flatbuffers.Builder) (flatbuffers.UOffsetT, error) {
	var nodes []flatbuffers.UOffsetT
	
	archEval := schema.StrEvaluatorCreate(builder, wls.StringEvaluatorsMACHINE_ARCHITECTURE, l.Arch, wls.CmpTypeSTRCMP_EXACT)
	archNode := schema.EvaluatorNodeCreate(builder, wls.EvaluatorTypeStrEvaluator, "architecture matching", archEval)
	nodes = append(nodes, schema.NodeTypeWrapperCreate(builder, archNode, wls.NodeTypeEvaluatorNode))
	
	libcEval := schema.StrEvaluatorCreate(builder, wls.StringEvaluatorsLIBC_FLAVOR, "glibc", wls.CmpTypeSTRCMP_EXACT)
	libcNode := schema.EvaluatorNodeCreate(builder, wls.EvaluatorTypeStrEvaluator, "glic flavor matching", libcEval)
	nodes = append(nodes, schema.NodeTypeWrapperCreate(builder, libcNode, wls.NodeTypeEvaluatorNode))
	
	if l.Min != "" {
		major, minor, patch, err := parseSemver(l.Min)
		if err != nil {
			log.Fatalf("Failed to parse semver: %v", err)
		}
	
		majorEval := schema.NumEvaluatorCreate(builder, wls.NumericEvaluatorsLIBC_VERSION_MAJOR, major, wls.CmpTypeNUMCMP_GTE)
		majorNode := schema.EvaluatorNodeCreate(builder, wls.EvaluatorTypeNumEvaluator, "major version matching", majorEval)
		nodes = append(nodes, schema.NodeTypeWrapperCreate(builder, majorNode, wls.NodeTypeEvaluatorNode))
	
		minorEval := schema.NumEvaluatorCreate(builder, wls.NumericEvaluatorsLIBC_VERSION_MINOR, minor, wls.CmpTypeNUMCMP_GTE)
		minorNode := schema.EvaluatorNodeCreate(builder, wls.EvaluatorTypeNumEvaluator, "minor version matching", minorEval)
		nodes = append(nodes, schema.NodeTypeWrapperCreate(builder, minorNode, wls.NodeTypeEvaluatorNode))
	
		patchEval := schema.NumEvaluatorCreate(builder, wls.NumericEvaluatorsLIBC_VERSION_PATCH, patch, wls.CmpTypeNUMCMP_GTE)
		patchNode := schema.EvaluatorNodeCreate(builder, wls.EvaluatorTypeNumEvaluator, "patch version matching", patchEval)
		nodes = append(nodes, schema.NodeTypeWrapperCreate(builder, patchNode, wls.NodeTypeEvaluatorNode))
	}
	
	// Combine all evaluators with AND (arch AND libc AND version must all match)
	var libcNode flatbuffers.UOffsetT
	composite := schema.CompositeNodeCreate(builder, wls.BoolOperationBOOL_AND, l.Description, nodes)
	libcNode = schema.NodeTypeWrapperCreate(builder, composite, wls.NodeTypeCompositeNode)
	
	// Determine action based on supported flag
	actionId := getActionId(l.Supported)

	action := schema.ActionCreate(builder, actionId, l.Description, []string{l.Arch, l.Min})
	
	// Create and return the policy
	return schema.PolicyCreate(builder, l.Description, libcNode, []flatbuffers.UOffsetT{action}), nil
}