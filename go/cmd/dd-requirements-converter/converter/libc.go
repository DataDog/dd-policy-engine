package converter

import (
	"encoding/json"
	"log"

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

		// Semver comparison: version >= minVersion means:
		// 1. (major > minMajor) OR
		// 2. (major == minMajor AND minor > minMinor) OR
		// 3. (major == minMajor AND minor == minMinor AND patch >= minPatch)

		// case 1: major > minMajor
		majorGtEval := schema.NumEvaluatorCreate(builder, wls.NumericEvaluatorsLIBC_VERSION_MAJOR, int64(major), wls.CmpTypeNUMCMP_GT)
		majorGtNode := schema.EvaluatorNodeCreate(builder, wls.EvaluatorTypeNumEvaluator, "major version >", majorGtEval)
		case1 := schema.NodeTypeWrapperCreate(builder, majorGtNode, wls.NodeTypeEvaluatorNode)

		// major == minMajor (used in Case 2 and Case 3)
		majorEqEval := schema.NumEvaluatorCreate(builder, wls.NumericEvaluatorsLIBC_VERSION_MAJOR, int64(major), wls.CmpTypeNUMCMP_EQ)
		majorEqNode := schema.EvaluatorNodeCreate(builder, wls.EvaluatorTypeNumEvaluator, "major version ==", majorEqEval)
		majorEqWrapper := schema.NodeTypeWrapperCreate(builder, majorEqNode, wls.NodeTypeEvaluatorNode)

		// case 2: major == minMajor AND minor > minMinor
		minorGtEval := schema.NumEvaluatorCreate(builder, wls.NumericEvaluatorsLIBC_VERSION_MINOR, int64(minor), wls.CmpTypeNUMCMP_GT)
		minorGtNode := schema.EvaluatorNodeCreate(builder, wls.EvaluatorTypeNumEvaluator, "minor version >", minorGtEval)
		minorGtWrapper := schema.NodeTypeWrapperCreate(builder, minorGtNode, wls.NodeTypeEvaluatorNode)

		case2And := schema.CompositeNodeCreate(builder, wls.BoolOperationBOOL_AND, "major == && minor >", []flatbuffers.UOffsetT{majorEqWrapper, minorGtWrapper})
		case2 := schema.NodeTypeWrapperCreate(builder, case2And, wls.NodeTypeCompositeNode)

		// case 3: major == minMajor AND minor == minMinor AND patch >= minPatch
		minorEqEval := schema.NumEvaluatorCreate(builder, wls.NumericEvaluatorsLIBC_VERSION_MINOR, int64(minor), wls.CmpTypeNUMCMP_EQ)
		minorEqNode := schema.EvaluatorNodeCreate(builder, wls.EvaluatorTypeNumEvaluator, "minor version ==", minorEqEval)
		minorEqWrapper := schema.NodeTypeWrapperCreate(builder, minorEqNode, wls.NodeTypeEvaluatorNode)

		patchGteEval := schema.NumEvaluatorCreate(builder, wls.NumericEvaluatorsLIBC_VERSION_PATCH, int64(patch), wls.CmpTypeNUMCMP_GTE)
		patchGteNode := schema.EvaluatorNodeCreate(builder, wls.EvaluatorTypeNumEvaluator, "patch version >=", patchGteEval)
		patchGteWrapper := schema.NodeTypeWrapperCreate(builder, patchGteNode, wls.NodeTypeEvaluatorNode)

		case3And := schema.CompositeNodeCreate(builder, wls.BoolOperationBOOL_AND, "major == && minor == && patch >=", []flatbuffers.UOffsetT{majorEqWrapper, minorEqWrapper, patchGteWrapper})
		case3 := schema.NodeTypeWrapperCreate(builder, case3And, wls.NodeTypeCompositeNode)

		// combine all cases with OR
		versionOr := schema.CompositeNodeCreate(builder, wls.BoolOperationBOOL_OR, "check libc version is at least the minimum required version", []flatbuffers.UOffsetT{case1, case2, case3})
		versionNode := schema.NodeTypeWrapperCreate(builder, versionOr, wls.NodeTypeCompositeNode)
		nodes = append(nodes, versionNode)
	}

	// combine all evaluators with AND (arch AND libc flavor AND version must all match)
	var libcNode flatbuffers.UOffsetT
	composite := schema.CompositeNodeCreate(builder, wls.BoolOperationBOOL_AND, l.Description, nodes)
	libcNode = schema.NodeTypeWrapperCreate(builder, composite, wls.NodeTypeCompositeNode)

	// determine action based on supported flag
	actionId := getActionId(l.IsSupported)

	action := schema.ActionCreate(builder, actionId, l.Description, nil)

	return schema.PolicyCreate(builder, l.Description, libcNode, []flatbuffers.UOffsetT{action}), nil
}
