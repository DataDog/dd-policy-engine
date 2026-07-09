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

func (r JSONRequirements) ConvertToWLS(builder *flatbuffers.Builder) (flatbuffers.UOffsetT, error) {
	var rules []flatbuffers.UOffsetT

	fmt.Printf("Converting %d deny rules\n", len(r.Deny))
	for _, denyRule := range r.Deny {
		denyNode, err := denyRule.ConvertToWLS(builder)
		if err != nil {
			return 0, err
		}
		rules = append(rules, denyNode)
	}

	fmt.Printf("Converting %d glibc requirements\n", len(r.NativeDeps.Glibc))
	for _, glibc := range r.NativeDeps.Glibc {
		glibcNode, err := glibc.ConvertToWLS(builder, "glibc")
		if err != nil {
			return 0, err
		}
		if glibcNode != 0 {
			rules = append(rules, glibcNode)
		}
	}

	fmt.Printf("Converting %d musl requirements\n", len(r.NativeDeps.Musl))
	for _, musl := range r.NativeDeps.Musl {
		muslNode, err := musl.ConvertToWLS(builder, "musl")
		if err != nil {
			return 0, err
		}
		if muslNode != 0 {
			rules = append(rules, muslNode)
		}
	}

	composite := schema.CompositeNodeCreate(builder, wls.BoolOperationBOOL_OR, "requirements", rules)
	compositeNode := schema.NodeTypeWrapperCreate(builder, composite, wls.NodeTypeCompositeNode)

	action := schema.ActionCreate(builder, wls.ActionIdINJECT_DENY, "requirements", nil)
	policy := schema.PolicyCreate(builder, "All requirements", compositeNode, []flatbuffers.UOffsetT{action}, "hardcoded")
	return schema.PoliciesCreate(builder, []flatbuffers.UOffsetT{policy}), nil
}
