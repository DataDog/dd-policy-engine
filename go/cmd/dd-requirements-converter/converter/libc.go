package converter

import (
	"encoding/json"
	"errors"
	"fmt"
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
		versionText := strings.TrimSpace(l.RequiredMinVersion.Original())
		if suffix := strings.IndexAny(versionText, "-+"); suffix >= 0 {
			versionText = versionText[:suffix]
		}
		segments := l.RequiredMinVersion.Segments()
		if strings.Count(versionText, ".") < 1 || len(segments) < 2 {
			return 0, fmt.Errorf("invalid libc version %q: expected at least major.minor", versionText)
		}

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
			nodes = append(nodes, buildRuntimeVersionBelow(builder, major, minor, patch))
		} else {
			// supported: false + version → DENY if version >= min
			// Semver comparison: version >= minVersion means:
			// 1. (major > minMajor) OR
			// 2. (major == minMajor AND minor >= minMinor) OR
			// 3. (major == minMajor AND minor == minMinor AND patch >= minPatch)
			nodes = append(nodes, buildRuntimeVersionAtLeast(builder, major, minor, patch))
		}
	}

	// combine all evaluators with AND
	composite := schema.CompositeNodeCreate(builder, wls.BoolOperationBOOL_AND, l.Description, nodes)
	return schema.NodeTypeWrapperCreate(builder, composite, wls.NodeTypeCompositeNode), nil
}

// Numeric evaluators place the policy threshold on the left side of comparisons.
func buildVersionEvaluator(
	builder *flatbuffers.Builder,
	id wls.NumericEvaluators,
	value int,
	comparator wls.CmpTypeNUM,
	description string,
) flatbuffers.UOffsetT {
	evaluator := schema.NumEvaluatorCreate(builder, id, int64(value), comparator)
	node := schema.EvaluatorNodeCreate(builder, wls.EvaluatorTypeNumEvaluator, description, evaluator)
	return schema.NodeTypeWrapperCreate(builder, node, wls.NodeTypeEvaluatorNode)
}

func buildVersionComposite(
	builder *flatbuffers.Builder,
	operator wls.BoolOperation,
	description string,
	children ...flatbuffers.UOffsetT,
) flatbuffers.UOffsetT {
	composite := schema.CompositeNodeCreate(builder, operator, description, children)
	return schema.NodeTypeWrapperCreate(builder, composite, wls.NodeTypeCompositeNode)
}

func buildRuntimeVersionAtLeast(builder *flatbuffers.Builder, major, minor, patch int) flatbuffers.UOffsetT {
	// case 1: runtime major > minMajor
	case1 := buildVersionEvaluator(
		builder, wls.NumericEvaluatorsLIBC_VERSION_MAJOR, major, wls.CmpTypeNUMCMP_LT, "major version <",
	)

	// major == minMajor (used in Case 2 and Case 3)
	majorEq := buildVersionEvaluator(
		builder, wls.NumericEvaluatorsLIBC_VERSION_MAJOR, major, wls.CmpTypeNUMCMP_EQ, "major version ==",
	)

	// case 2: runtime major == minMajor AND runtime minor > minMinor
	minorLt := buildVersionEvaluator(
		builder, wls.NumericEvaluatorsLIBC_VERSION_MINOR, minor, wls.CmpTypeNUMCMP_LT, "minor version <",
	)
	case2 := buildVersionComposite(builder, wls.BoolOperationBOOL_AND, "major == && minor <", majorEq, minorLt)

	// case 3: runtime major == minMajor AND runtime minor == minMinor AND runtime patch >= minPatch
	minorEq := buildVersionEvaluator(
		builder, wls.NumericEvaluatorsLIBC_VERSION_MINOR, minor, wls.CmpTypeNUMCMP_EQ, "minor version ==",
	)
	patchLte := buildVersionEvaluator(
		builder, wls.NumericEvaluatorsLIBC_VERSION_PATCH, patch, wls.CmpTypeNUMCMP_LTE, "patch version <=",
	)
	case3 := buildVersionComposite(
		builder, wls.BoolOperationBOOL_AND, "major == && minor == && patch <=", majorEq, minorEq, patchLte,
	)

	// combine all cases with OR
	return buildVersionComposite(builder, wls.BoolOperationBOOL_OR, "version <= min", case1, case2, case3)
}

func buildRuntimeVersionBelow(builder *flatbuffers.Builder, major, minor, patch int) flatbuffers.UOffsetT {
	// case 1: runtime major < minMajor
	case1 := buildVersionEvaluator(
		builder, wls.NumericEvaluatorsLIBC_VERSION_MAJOR, major, wls.CmpTypeNUMCMP_GT, "major version >",
	)

	// major == minMajor (used in Case 2 and Case 3)
	majorEq := buildVersionEvaluator(
		builder, wls.NumericEvaluatorsLIBC_VERSION_MAJOR, major, wls.CmpTypeNUMCMP_EQ, "major version ==",
	)

	// case 2: runtime major == minMajor AND runtime minor < minMinor
	minorGt := buildVersionEvaluator(
		builder, wls.NumericEvaluatorsLIBC_VERSION_MINOR, minor, wls.CmpTypeNUMCMP_GT, "minor version >",
	)
	case2 := buildVersionComposite(builder, wls.BoolOperationBOOL_AND, "major == && minor >", majorEq, minorGt)

	// case 3: runtime major == minMajor AND runtime minor == minMinor AND runtime patch < minPatch
	minorEq := buildVersionEvaluator(
		builder, wls.NumericEvaluatorsLIBC_VERSION_MINOR, minor, wls.CmpTypeNUMCMP_EQ, "minor version ==",
	)
	patchGt := buildVersionEvaluator(
		builder, wls.NumericEvaluatorsLIBC_VERSION_PATCH, patch, wls.CmpTypeNUMCMP_GT, "patch version >",
	)
	case3 := buildVersionComposite(
		builder, wls.BoolOperationBOOL_AND, "major == && minor == && patch >", majorEq, minorEq, patchGt,
	)

	// combine all cases with OR
	return buildVersionComposite(builder, wls.BoolOperationBOOL_OR, "version > min", case1, case2, case3)
}
