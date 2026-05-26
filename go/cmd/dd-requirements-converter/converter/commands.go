package converter

import (
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
	var matcher wls.CmpTypeSTR
	// no wildcards, return exact match
	if strings.ContainsAny(pattern, "*?") {
		matcher = wls.CmpTypeSTRCMP_WILDCARD
	} else {
		matcher = wls.CmpTypeSTRCMP_EXACT
	}
	strEvaluator := schema.StrEvaluatorCreate(builder, wls.StringEvaluatorsPROCESS_EXE_FULL_PATH, pattern, matcher)
	node := schema.EvaluatorNodeCreate(builder, wls.EvaluatorTypeStrEvaluator, "Path matching: "+pattern, strEvaluator, "")
	return schema.NodeTypeWrapperCreate(builder, node, wls.NodeTypeEvaluatorNode), nil
}
