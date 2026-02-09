package converter

import (
	"fmt"

	"github.com/DataDog/dd-policy-engine/go/schema"

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
	var policies []flatbuffers.UOffsetT

	fmt.Printf("Converting %d deny rules:\n", len(r.Deny))
	for i, denyRule := range r.Deny {
		fmt.Printf("- deny #%d\n", i)
		denyNode, err := denyRule.ConvertToWLS(builder)
		if err != nil {
			return 0, err
		}
		if denyNode != 0 {
			denyAction := schema.ActionCreate(builder, wls.ActionIdINJECT_DENY, denyRule.Description)
			denyPolicy := schema.PolicyCreate(builder, denyRule.Description, denyNode, []flatbuffers.UOffsetT{denyAction})
			policies = append(policies, denyPolicy)
		}
	}

	fmt.Printf("Converting %d glibc requirements:\n", len(r.NativeDeps.Glibc))
	for i, glibc := range r.NativeDeps.Glibc {
		glibc, err := glibc.ConvertToWLS(builder, "glibc")
		if err != nil {
			return 0, err
		}
		policies = append(policies, glibc)
	}

	fmt.Printf("Converting %d musl requirements:\n", len(r.NativeDeps.Musl))
	for i, musl := range r.NativeDeps.Musl {
		musl, err := musl.ConvertToWLS(builder, "musl")
		if err != nil {
			return 0, err
		}
		policies = append(policies, musl)
	}

	return schema.PoliciesCreate(builder, policies), nil
}
