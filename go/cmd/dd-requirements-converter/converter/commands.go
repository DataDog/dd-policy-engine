package converter

import (
	"strings"

	"github.com/DataDog/dd-policy-engine/go/schema"
	"github.com/DataDog/dd-policy-engine/go/schema/dd/wls"

	flatbuffers "github.com/google/flatbuffers/go"
)

// CmdPattern represents a glob pattern for matching executable paths.
type CmdPattern string

// ConvertToWLS converts a glob pattern to a single evaluator node. It is labelled
// with the owning rule's description only when this node is the outermost one the
// rule produces; otherwise it is left unnamed, since the pattern is already
// reported as the node's evaluator and value.
func (c CmdPattern) ConvertToWLS(builder *flatbuffers.Builder, description string) (flatbuffers.UOffsetT, error) {
	pattern := string(c)
	var matcher wls.CmpTypeSTR
	// no wildcards, return exact match
	if strings.ContainsAny(pattern, "*?") {
		matcher = wls.CmpTypeSTRCMP_WILDCARD
	} else {
		matcher = wls.CmpTypeSTRCMP_EXACT
	}
	strEvaluator := schema.StrEvaluatorCreate(builder, wls.StringEvaluatorsPROCESS_EXE_FULL_PATH, pattern, matcher)
	node := schema.EvaluatorNodeCreate(builder, wls.EvaluatorTypeStrEvaluator, description, strEvaluator)
	return schema.NodeTypeWrapperCreate(builder, node, wls.NodeTypeEvaluatorNode), nil
}
