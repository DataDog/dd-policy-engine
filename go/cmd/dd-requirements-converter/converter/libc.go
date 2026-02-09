package converter

import (
	"log"
	"encoding/json"

	"github.com/DataDog/dd-policy-engine/go/schema"
	"github.com/DataDog/dd-policy-engine/go/schema/dd/wls"
	"github.com/hashicorp/go-version"

	flatbuffers "github.com/google/flatbuffers/go"
)

type JSONlibc struct {
	Arch               string           `json:"arch"`
	Description        string           `json:"description"`
	IsSupported        bool             `json:"supported"`
	RequiredMinVersion *RequiredVersion `json:"min"`
}

type RequiredVersion struct {
	version.Version
}

// UnmarshalJSON Implement UnmarshalJSON to enforce allowed values
func (rv *RequiredVersion) UnmarshalJSON(data []byte) error {
	var versionStr string
	if err := json.Unmarshal(data, &versionStr); err != nil {
		return err
	}

	if versionStr == "" {
		return nil
	}

	v, err := version.NewVersion(versionStr)
	if err != nil {
		return err
	}

	*rv = RequiredVersion{Version: *v}
	return nil
}

func getActionId(supported bool) wls.ActionId {
	if supported {
		return wls.ActionIdINJECT_ALLOW
	}
	return wls.ActionIdINJECT_DENY
}

// ConvertToWLS converts a JSONGlibc requirement to a WLS policy
func (l JSONlibc) ConvertToWLS(builder *flatbuffers.Builder, flavor string) (flatbuffers.UOffsetT, error) {
	var nodes []flatbuffers.UOffsetT

	archEval := schema.StrEvaluatorCreate(builder, wls.StringEvaluatorsMACHINE_ARCHITECTURE, l.Arch, wls.CmpTypeSTRCMP_EXACT)
	archNode := schema.EvaluatorNodeCreate(builder, wls.EvaluatorTypeStrEvaluator, "architecture matching", archEval)
	nodes = append(nodes, schema.NodeTypeWrapperCreate(builder, archNode, wls.NodeTypeEvaluatorNode))

	flavorEval := schema.StrEvaluatorCreate(builder, wls.StringEvaluatorsLIBC_FLAVOR, flavor, wls.CmpTypeSTRCMP_EXACT)
	flavorNode := schema.EvaluatorNodeCreate(builder, wls.EvaluatorTypeStrEvaluator, "flavor matching", flavorEval)
	nodes = append(nodes, schema.NodeTypeWrapperCreate(builder, flavorNode, wls.NodeTypeEvaluatorNode))

	if l.RequiredMinVersion != nil {
		segments := l.RequiredMinVersion.Segments()

		if len(segments) < 2 {
			log.Fatalf("Invalid version: %v", l.RequiredMinVersion)
		}

		major := segments[0]
		minor := segments[1]
		patch := 0

		if len(segments) > 2 {
			patch = segments[2]
		}

		majorEval := schema.NumEvaluatorCreate(builder, wls.NumericEvaluatorsLIBC_VERSION_MAJOR, int64(major), wls.CmpTypeNUMCMP_GTE)
		majorNode := schema.EvaluatorNodeCreate(builder, wls.EvaluatorTypeNumEvaluator, "major version matching", majorEval)
		nodes = append(nodes, schema.NodeTypeWrapperCreate(builder, majorNode, wls.NodeTypeEvaluatorNode))

		minorEval := schema.NumEvaluatorCreate(builder, wls.NumericEvaluatorsLIBC_VERSION_MINOR, int64(minor), wls.CmpTypeNUMCMP_GTE)
		minorNode := schema.EvaluatorNodeCreate(builder, wls.EvaluatorTypeNumEvaluator, "minor version matching", minorEval)
		nodes = append(nodes, schema.NodeTypeWrapperCreate(builder, minorNode, wls.NodeTypeEvaluatorNode))

		patchEval := schema.NumEvaluatorCreate(builder, wls.NumericEvaluatorsLIBC_VERSION_PATCH, int64(patch), wls.CmpTypeNUMCMP_GTE)
		patchNode := schema.EvaluatorNodeCreate(builder, wls.EvaluatorTypeNumEvaluator, "patch version matching", patchEval)
		nodes = append(nodes, schema.NodeTypeWrapperCreate(builder, patchNode, wls.NodeTypeEvaluatorNode))
	}

	// Combine all evaluators with AND (arch AND libc AND version must all match)
	var libcNode flatbuffers.UOffsetT
	composite := schema.CompositeNodeCreate(builder, wls.BoolOperationBOOL_AND, l.Description, nodes)
	libcNode = schema.NodeTypeWrapperCreate(builder, composite, wls.NodeTypeCompositeNode)

	// Determine action based on supported flag
	actionId := getActionId(l.IsSupported)

	action := schema.ActionCreate(builder, actionId, l.Description, nil)

	// Create and return the policy
	return schema.PolicyCreate(builder, l.Description, libcNode, []flatbuffers.UOffsetT{action}), nil
}
