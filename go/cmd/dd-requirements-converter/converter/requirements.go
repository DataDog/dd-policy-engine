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

func ActionDescription(ruleName string) string {
	if ruleName == "" {
		return "Instrumentation rule is applied."
	}

	return fmt.Sprintf("Instrumentation rule %q is applied.", ruleName)
}

// denyPolicy gives one rule its own policy, so the rule's name is the policy
// description.
func denyPolicy(builder *flatbuffers.Builder, node flatbuffers.UOffsetT, description string) flatbuffers.UOffsetT {
	action := schema.ActionCreate(builder, wls.ActionIdINJECT_DENY, ActionDescription(description), nil)
	return schema.PolicyCreate(builder, description, node, []flatbuffers.UOffsetT{action})
}

func (r JSONRequirements) ConvertToWLS(builder *flatbuffers.Builder) (flatbuffers.UOffsetT, error) {
	var policies []flatbuffers.UOffsetT

	fmt.Printf("Converting %d deny rules\n", len(r.Deny))
	for _, denyRule := range r.Deny {
		denyNode, err := denyRule.ConvertToWLS(builder)
		if err != nil {
			return 0, err
		}
		policies = append(policies, denyPolicy(builder, denyNode, denyRule.Description))
	}

	fmt.Printf("Converting %d glibc requirements\n", len(r.NativeDeps.Glibc))
	for _, glibc := range r.NativeDeps.Glibc {
		glibcNode, err := glibc.ConvertToWLS(builder, "glibc")
		if err != nil {
			return 0, err
		}
		if glibcNode != 0 {
			policies = append(policies, denyPolicy(builder, glibcNode, glibc.RuleDescription("glibc")))
		}
	}

	fmt.Printf("Converting %d musl requirements\n", len(r.NativeDeps.Musl))
	for _, musl := range r.NativeDeps.Musl {
		muslNode, err := musl.ConvertToWLS(builder, "musl")
		if err != nil {
			return 0, err
		}
		if muslNode != 0 {
			policies = append(policies, denyPolicy(builder, muslNode, musl.RuleDescription("musl")))
		}
	}

	return schema.PoliciesCreate(builder, policies), nil
}
