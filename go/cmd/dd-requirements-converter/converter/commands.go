package converter

import (
	"errors"
	"strings"

	"github.com/DataDog/dd-policy-engine/go/schema"
	"github.com/DataDog/dd-policy-engine/go/schema/dd/wls"

	flatbuffers "github.com/google/flatbuffers/go"
)

// CmdPattern represents a glob pattern for matching executable paths.
type CmdPattern string

// ConvertToWLS converts a glob pattern to a single evaluator node.
// If the pattern produces multiple evaluators, they are combined with AND.
func (c CmdPattern) ConvertToWLS(builder *flatbuffers.Builder) (flatbuffers.UOffsetT, error) {
	pattern := string(c)

	// no wildcards, return exact match
	if !strings.Contains(pattern, "*") {
		strEvaluator := schema.StrEvaluatorCreate(builder, wls.StringEvaluatorsPROCESS_EXE_FULL_PATH, pattern, wls.CmpTypeSTRCMP_EXACT)
		node := schema.EvaluatorNodeCreate(builder, wls.EvaluatorTypeStrEvaluator, "Path matching: "+pattern, strEvaluator)
		return schema.NodeTypeWrapperCreate(builder, node, wls.NodeTypeEvaluatorNode), nil
	}

	// split by ** and * to collect all literal parts
	// e.g., "**/foo/*/bar" -> ["", "/foo/", "/bar"] -> ["/foo/", "/bar"]
	var parts []string
	for _, segment := range strings.Split(pattern, "**") {
		for _, part := range strings.Split(segment, "*") {
			if part != "" && part != "/" {
				parts = append(parts, part)
			}
		}
	}

	if len(parts) == 0 {
		return 0, errors.New("no parts found in pattern")
	}

	startsWithWildcard := strings.HasPrefix(pattern, "*")
	endsWithWildcard := strings.HasSuffix(pattern, "*")

	var nodes []flatbuffers.UOffsetT
	for i, part := range parts {
		isFirst := i == 0
		isLast := i == len(parts)-1

		// determine comparison type based on position
		var cmp wls.CmpTypeSTR
		switch {
		case isFirst && !startsWithWildcard:
			// first part, pattern doesn't start with * -> PREFIX
			cmp = wls.CmpTypeSTRCMP_PREFIX
		case isLast && !endsWithWildcard:
			// last part, pattern doesn't end with * -> SUFFIX
			cmp = wls.CmpTypeSTRCMP_SUFFIX
		default:
			// middle parts -> CONTAINS
			cmp = wls.CmpTypeSTRCMP_CONTAINS
		}

		strEvaluator := schema.StrEvaluatorCreate(builder, wls.StringEvaluatorsPROCESS_EXE_FULL_PATH, part, cmp)
		node := schema.EvaluatorNodeCreate(builder, wls.EvaluatorTypeStrEvaluator, "Path matching: "+part, strEvaluator)
		nodes = append(nodes, schema.NodeTypeWrapperCreate(builder, node, wls.NodeTypeEvaluatorNode))
	}

	// Return single node or combine multiple with AND
	if len(nodes) == 1 {
		return nodes[0], nil
	}
	andNode := schema.CompositeNodeCreate(builder, wls.BoolOperationBOOL_AND, "Pattern: "+pattern, nodes)
	return schema.NodeTypeWrapperCreate(builder, andNode, wls.NodeTypeCompositeNode), nil
}
