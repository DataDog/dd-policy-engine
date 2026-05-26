package converter

import (
	"fmt"

	"github.com/DataDog/dd-policy-engine/go/schema"
	"github.com/DataDog/dd-policy-engine/go/schema/dd/wls"

	flatbuffers "github.com/google/flatbuffers/go"
)

type JSONRequirements struct {
	Schema     string         `json:"$schema,omitempty"`
	Version    int            `json:"version"`
	Deny       []JSONDeny     `json:"deny"`
	NativeDeps JSONNativeDeps `json:"native_deps"`
}

type JSONNativeDeps struct {
	Glibc []JSONlibc `json:"glibc"`
	Musl  []JSONlibc `json:"musl"`
}

// ConvertToWLS produces a single Policy whose rules tree is a top-level OR over
// all source rules — preserving the engine's short-circuit on OR so that
// languages with many rules (Ruby has 251) don't pay an O(n) cost on every
// process startup.
//
// Each rule's root node carries a stable rule_id (see CompositeNode.rule_id /
// EvaluatorNode.rule_id in nodes.fbs). When the top-level OR matches a rule,
// the engine reads that rule_id (see composite_evaluator in
// dd-policy-engine/c/src/evaluator.c) and exposes it via
// plcs_eval_ctx_get_matched_rule_id, so the consuming action callback can
// surface "which rule fired" in telemetry.
func (r JSONRequirements) ConvertToWLS(builder *flatbuffers.Builder) (flatbuffers.UOffsetT, error) {
	var ruleSubtrees []flatbuffers.UOffsetT

	fmt.Printf("Converting %d deny rules\n", len(r.Deny))
	for _, denyRule := range r.Deny {
		if denyRule.Id == "" {
			return 0, fmt.Errorf("deny rule has no id; cannot identify it in telemetry")
		}
		denyNode, err := denyRule.ConvertToWLS(builder, denyRule.Id)
		if err != nil {
			return 0, err
		}
		ruleSubtrees = append(ruleSubtrees, denyNode)
	}

	fmt.Printf("Converting %d glibc requirements\n", len(r.NativeDeps.Glibc))
	for _, glibc := range r.NativeDeps.Glibc {
		glibcNode, err := glibc.ConvertToWLS(builder, "glibc", libcRuleID("glibc", glibc))
		if err != nil {
			return 0, err
		}
		if glibcNode == 0 {
			continue
		}
		ruleSubtrees = append(ruleSubtrees, glibcNode)
	}

	fmt.Printf("Converting %d musl requirements\n", len(r.NativeDeps.Musl))
	for _, musl := range r.NativeDeps.Musl {
		muslNode, err := musl.ConvertToWLS(builder, "musl", libcRuleID("musl", musl))
		if err != nil {
			return 0, err
		}
		if muslNode == 0 {
			continue
		}
		ruleSubtrees = append(ruleSubtrees, muslNode)
	}

	composite := schema.CompositeNodeCreate(builder, wls.BoolOperationBOOL_OR, "requirements", ruleSubtrees, "")
	compositeNode := schema.NodeTypeWrapperCreate(builder, composite, wls.NodeTypeCompositeNode)

	action := schema.ActionCreate(builder, wls.ActionIdINJECT_DENY, "requirements", nil)
	policy := schema.PolicyCreate(builder, "All requirements", compositeNode, []flatbuffers.UOffsetT{action})
	return schema.PoliciesCreate(builder, []flatbuffers.UOffsetT{policy}), nil
}

// libcRuleID synthesizes a stable, bounded-cardinality identifier for a libc
// requirement (JSONlibc has no `id` field). The shape is
// "libc_<flavor>_<arch>[_min_<v>|_unsupported[_above_<v>]]" so downstream
// telemetry can distinguish e.g. "glibc x86_64 < 2.17" from "musl arm64
// entirely unsupported" without parsing the rule body.
func libcRuleID(flavor string, l JSONlibc) string {
	if l.RequiredMinVersion != nil {
		if l.IsSupported {
			return fmt.Sprintf("libc_%s_%s_below_min_%s", flavor, l.Arch, l.RequiredMinVersion.String())
		}
		return fmt.Sprintf("libc_%s_%s_unsupported_at_or_above_%s", flavor, l.Arch, l.RequiredMinVersion.String())
	}
	return fmt.Sprintf("libc_%s_%s_unsupported", flavor, l.Arch)
}
