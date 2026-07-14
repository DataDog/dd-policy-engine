package converter

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

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

	versionStr = strings.TrimSpace(versionStr)
	versionCore := versionStr
	if suffix := strings.IndexAny(versionCore, "-+"); suffix >= 0 {
		versionCore = versionCore[:suffix]
	}
	segments := strings.Split(versionCore, ".")
	if len(segments) < 2 {
		return fmt.Errorf("invalid libc version %q: expected at least major.minor", versionStr)
	}
	for _, segment := range segments[:2] {
		if _, err := strconv.ParseUint(segment, 10, 64); err != nil {
			return fmt.Errorf("invalid libc version %q: expected numeric major.minor", versionStr)
		}
	}

	v, err := version.NewVersion(versionStr)
	if err != nil {
		return err
	}

	*rv = RequiredVersion{Version: *v}
	return nil
}

// ConvertToWLS converts a JSONlibc requirement to a WLS DENY policy.
//
// Logic:
//   - supported: true + no version → No policy needed
//   - supported: true + version → DENY if arch+flavor match AND version < min
//   - supported: false + no version → DENY if arch+flavor match
//   - supported: false + version → DENY if arch+flavor match AND version >= min
func (l JSONlibc) ConvertToWLS(builder *flatbuffers.Builder, flavor string) (flatbuffers.UOffsetT, error) {
	// If supported and no version requirement, no policy needed (allowed by default)
	if l.IsSupported && l.RequiredMinVersion == nil {
		return 0, nil
	}

	var nodes []flatbuffers.UOffsetT

	// Normalize architecture string to canonical form (e.g., "arm64" -> "aarch64")
	archEnum, ok := schema.MachineArchitectureFromString[l.Arch]
	if !ok {
		return 0, errors.New("unknown architecture")
	}
	archEval := schema.StrEvaluatorCreate(builder, wls.StringEvaluatorsMACHINE_ARCHITECTURE, schema.MachineArchitectureToString[archEnum], wls.CmpTypeSTRCMP_EXACT)
	archNode := schema.EvaluatorNodeCreate(builder, wls.EvaluatorTypeStrEvaluator, "architecture matching", archEval)
	nodes = append(nodes, schema.NodeTypeWrapperCreate(builder, archNode, wls.NodeTypeEvaluatorNode))

	flavorEval := schema.StrEvaluatorCreate(builder, wls.StringEvaluatorsLIBC_FLAVOR, flavor, wls.CmpTypeSTRCMP_EXACT)
	flavorNode := schema.EvaluatorNodeCreate(builder, wls.EvaluatorTypeStrEvaluator, "flavor matching", flavorEval)
	nodes = append(nodes, schema.NodeTypeWrapperCreate(builder, flavorNode, wls.NodeTypeEvaluatorNode))

	if l.RequiredMinVersion != nil {
		segments := l.RequiredMinVersion.Segments()

		major := segments[0]
		minor := segments[1]
		patch := 0

		if len(segments) > 2 {
			patch = segments[2]
		}

		if l.IsSupported {
			// supported: true + version → DENY if version < min
			// Semver comparison: version < minVersion means:
			// 1. (major < minMajor) OR
			// 2. (major == minMajor AND minor < minMinor) OR
			// 3. (major == minMajor AND minor == minMinor AND patch < minPatch)
			nodes = append(nodes, buildVersionGreaterThan(builder, major, minor, patch))
		} else {
			// supported: false + version → DENY if version >= min
			// Semver comparison: version >= minVersion means:
			// 1. (major > minMajor) OR
			// 2. (major == minMajor AND minor >= minMinor) OR
			// 3. (major == minMajor AND minor == minMinor AND patch >= minPatch)
			nodes = append(nodes, buildVersionLessThanOrEqual(builder, major, minor, patch))
		}
	}

	// combine all evaluators with AND
	composite := schema.CompositeNodeCreate(builder, wls.BoolOperationBOOL_AND, l.Description, nodes)
	return schema.NodeTypeWrapperCreate(builder, composite, wls.NodeTypeCompositeNode), nil
}

// buildVersionLessThan creates a node that matches when version < major.minor.patch
func buildVersionLessThanOrEqual(builder *flatbuffers.Builder, major, minor, patch int) flatbuffers.UOffsetT {
	// case 1: major < minMajor
	majorLtEval := schema.NumEvaluatorCreate(builder, wls.NumericEvaluatorsLIBC_VERSION_MAJOR, int64(major), wls.CmpTypeNUMCMP_LT)
	majorLtNode := schema.EvaluatorNodeCreate(builder, wls.EvaluatorTypeNumEvaluator, "major version <", majorLtEval)
	case1 := schema.NodeTypeWrapperCreate(builder, majorLtNode, wls.NodeTypeEvaluatorNode)

	// major == minMajor (used in Case 2 and Case 3)
	majorEqEval := schema.NumEvaluatorCreate(builder, wls.NumericEvaluatorsLIBC_VERSION_MAJOR, int64(major), wls.CmpTypeNUMCMP_EQ)
	majorEqNode := schema.EvaluatorNodeCreate(builder, wls.EvaluatorTypeNumEvaluator, "major version ==", majorEqEval)
	majorEqWrapper := schema.NodeTypeWrapperCreate(builder, majorEqNode, wls.NodeTypeEvaluatorNode)

	// case 2: major == minMajor AND minor < minMinor
	minorLteEval := schema.NumEvaluatorCreate(builder, wls.NumericEvaluatorsLIBC_VERSION_MINOR, int64(minor), wls.CmpTypeNUMCMP_LTE)
	minorLteNode := schema.EvaluatorNodeCreate(builder, wls.EvaluatorTypeNumEvaluator, "minor version <=", minorLteEval)
	minorLteWrapper := schema.NodeTypeWrapperCreate(builder, minorLteNode, wls.NodeTypeEvaluatorNode)

	case2And := schema.CompositeNodeCreate(builder, wls.BoolOperationBOOL_AND, "major == && minor <=", []flatbuffers.UOffsetT{majorEqWrapper, minorLteWrapper})
	case2 := schema.NodeTypeWrapperCreate(builder, case2And, wls.NodeTypeCompositeNode)

	// case 3: major == minMajor AND minor == minMinor AND patch < minPatch
	minorEqEval := schema.NumEvaluatorCreate(builder, wls.NumericEvaluatorsLIBC_VERSION_MINOR, int64(minor), wls.CmpTypeNUMCMP_EQ)
	minorEqNode := schema.EvaluatorNodeCreate(builder, wls.EvaluatorTypeNumEvaluator, "minor version ==", minorEqEval)
	minorEqWrapper := schema.NodeTypeWrapperCreate(builder, minorEqNode, wls.NodeTypeEvaluatorNode)

	patchLteEval := schema.NumEvaluatorCreate(builder, wls.NumericEvaluatorsLIBC_VERSION_PATCH, int64(patch), wls.CmpTypeNUMCMP_LTE)
	patchLteNode := schema.EvaluatorNodeCreate(builder, wls.EvaluatorTypeNumEvaluator, "patch version <=", patchLteEval)
	patchLteWrapper := schema.NodeTypeWrapperCreate(builder, patchLteNode, wls.NodeTypeEvaluatorNode)

	case3And := schema.CompositeNodeCreate(builder, wls.BoolOperationBOOL_AND, "major == && minor == && patch <=", []flatbuffers.UOffsetT{majorEqWrapper, minorEqWrapper, patchLteWrapper})
	case3 := schema.NodeTypeWrapperCreate(builder, case3And, wls.NodeTypeCompositeNode)

	// combine all cases with OR
	versionOr := schema.CompositeNodeCreate(builder, wls.BoolOperationBOOL_OR, "version <= min", []flatbuffers.UOffsetT{case1, case2, case3})
	return schema.NodeTypeWrapperCreate(builder, versionOr, wls.NodeTypeCompositeNode)
}

// buildVersionGreaterOrEqual creates a node that matches when version >= major.minor.patch
func buildVersionGreaterThan(builder *flatbuffers.Builder, major, minor, patch int) flatbuffers.UOffsetT {
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

	patchGtEval := schema.NumEvaluatorCreate(builder, wls.NumericEvaluatorsLIBC_VERSION_PATCH, int64(patch), wls.CmpTypeNUMCMP_GT)
	patchGtNode := schema.EvaluatorNodeCreate(builder, wls.EvaluatorTypeNumEvaluator, "patch version >", patchGtEval)
	patchGtWrapper := schema.NodeTypeWrapperCreate(builder, patchGtNode, wls.NodeTypeEvaluatorNode)

	case3And := schema.CompositeNodeCreate(builder, wls.BoolOperationBOOL_AND, "major == && minor == && patch >", []flatbuffers.UOffsetT{majorEqWrapper, minorEqWrapper, patchGtWrapper})
	case3 := schema.NodeTypeWrapperCreate(builder, case3And, wls.NodeTypeCompositeNode)

	// combine all cases with OR
	versionOr := schema.CompositeNodeCreate(builder, wls.BoolOperationBOOL_OR, "version > min", []flatbuffers.UOffsetT{case1, case2, case3})
	return schema.NodeTypeWrapperCreate(builder, versionOr, wls.NodeTypeCompositeNode)
}
