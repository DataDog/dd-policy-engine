// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache 2.0 License. This product includes software developed at
// Datadog (https://www.datadoghq.com/).
//
// Copyright 2025-Present Datadog, Inc.

package policies

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strings"
)

// dd-wls JSON identifiers, matching the policy.schema.json enums (the JSON
// projection of the FlatBuffers policy schema shared with the C engine).
const (
	nodeEvaluator = "EvaluatorNode"
	nodeComposite = "CompositeNode"

	evalString   = "StrEvaluator"
	evalNumeric  = "NumEvaluator"
	evalUNumeric = "UNumEvaluator"

	opAnd = "BOOL_AND"
	opOr  = "BOOL_OR"
	opNot = "BOOL_NOT"

	cmpExact    = "CMP_EXACT"
	cmpPrefix   = "CMP_PREFIX"
	cmpSuffix   = "CMP_SUFFIX"
	cmpContains = "CMP_CONTAINS"
	cmpWildcard = "CMP_WILDCARD"

	cmpEq  = "CMP_EQ"
	cmpGt  = "CMP_GT"
	cmpGte = "CMP_GTE"
	cmpLt  = "CMP_LT"
	cmpLte = "CMP_LTE"

	actionInjectAllow    = "INJECT_ALLOW"
	actionInjectDeny     = "INJECT_DENY"
	actionEnableSDK      = "ENABLE_SDK"
	actionEnableProfiler = "ENABLE_PROFILER"
)

type wlsPolicies struct {
	Policies []wlsPolicy `json:"policies"`
}

type wlsPolicy struct {
	Description string      `json:"description"`
	Rules       wlsNodeWrap `json:"rules"`
	Actions     []wlsAction `json:"actions"`
	ID          *wlsUUID    `json:"id"`
	Version     int64       `json:"version"`
}

// wlsUUID is the dd-wls 128-bit identifier, split into two unsigned longs
// because FlatBuffers cannot represent fixed-size byte arrays in Go.
type wlsUUID struct {
	Hi uint64 `json:"hi"`
	Lo uint64 `json:"lo"`
}

type wlsNodeWrap struct {
	NodeType string          `json:"node_type"`
	Node     json.RawMessage `json:"node"`
}

type wlsComposite struct {
	Description string        `json:"description"`
	Op          string        `json:"op"`
	Children    []wlsNodeWrap `json:"children"`
}

type wlsEvaluatorNode struct {
	Description string          `json:"description"`
	EvalType    string          `json:"eval_type"`
	Eval        json.RawMessage `json:"eval"`
}

type wlsStrEval struct {
	ID    string `json:"id"`
	Cmp   string `json:"cmp"`
	Value string `json:"value"`
}

type wlsNumEval struct {
	ID    string `json:"id"`
	Cmp   string `json:"cmp"`
	Value int64  `json:"value"`
}

type wlsUNumEval struct {
	ID    string `json:"id"`
	Cmp   string `json:"cmp"`
	Value uint64 `json:"value"`
}

type wlsAction struct {
	Action      string   `json:"action"`
	Description string   `json:"description"`
	Values      []string `json:"values"`
}

// ParsePolicies decodes a dd-wls policies document into the native policy model.
// Label evaluators encode their key as "key=value" in the StrEvaluator value; a
// CMP_PREFIX on "key=" (empty value part) is decoded as an existence check.
func ParsePolicies(raw []byte) ([]Policy, error) {
	var doc wlsPolicies
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("invalid policies document: %w", err)
	}

	out := make([]Policy, 0, len(doc.Policies))
	for i, p := range doc.Policies {
		rules, err := decodeNodeWrap(p.Rules)
		if err != nil {
			return nil, fmt.Errorf("policy[%d] %q: %w", i, p.Description, err)
		}
		out = append(out, Policy{
			Name:    p.Description,
			ID:      decodeUUID(p.ID),
			Version: p.Version,
			Rules:   rules,
			Outcome: decodeActions(p.Actions),
		})
	}
	return out, nil
}

func decodeNodeWrap(w wlsNodeWrap) (*Node, error) {
	switch w.NodeType {
	case nodeComposite:
		var c wlsComposite
		if err := json.Unmarshal(w.Node, &c); err != nil {
			return nil, fmt.Errorf("invalid composite node: %w", err)
		}
		return decodeComposite(c)
	case nodeEvaluator:
		var e wlsEvaluatorNode
		if err := json.Unmarshal(w.Node, &e); err != nil {
			return nil, fmt.Errorf("invalid evaluator node: %w", err)
		}
		return decodeEvaluatorNode(e)
	default:
		return nil, fmt.Errorf("unsupported node_type %q", w.NodeType)
	}
}

func decodeComposite(c wlsComposite) (*Node, error) {
	children := make([]*Node, 0, len(c.Children))
	for _, child := range c.Children {
		n, err := decodeNodeWrap(child)
		if err != nil {
			return nil, err
		}
		children = append(children, n)
	}
	switch c.Op {
	case opAnd:
		return &Node{Op: OpAnd, Children: children}, nil
	case opOr:
		return &Node{Op: OpOr, Children: children}, nil
	case opNot:
		if len(children) != 1 {
			return nil, fmt.Errorf("BOOL_NOT requires exactly one child, got %d", len(children))
		}
		return &Node{Op: OpNot, Children: children}, nil
	default:
		return nil, fmt.Errorf("unsupported boolean operation %q", c.Op)
	}
}

// decodeEvaluatorNode dispatches on the wire union member. The Go engine is a
// faithful, generic reimplementation of the C engine: it accepts any evaluator
// id and resolves it against the Context at evaluation time (an id with no
// matching fact abstains, like the C engine's NULL context), so it is not
// restricted to a Kubernetes subset.
func decodeEvaluatorNode(e wlsEvaluatorNode) (*Node, error) {
	switch e.EvalType {
	case evalString:
		var se wlsStrEval
		if err := json.Unmarshal(e.Eval, &se); err != nil {
			return nil, fmt.Errorf("invalid string evaluator: %w", err)
		}
		return decodeStrEval(se)
	case evalNumeric:
		var ne wlsNumEval
		if err := json.Unmarshal(e.Eval, &ne); err != nil {
			return nil, fmt.Errorf("invalid numeric evaluator: %w", err)
		}
		cmp, err := decodeNumCmp(ne.Cmp)
		if err != nil {
			return nil, err
		}
		return NumericLeaf(ne.ID, cmp, ne.Value), nil
	case evalUNumeric:
		var ue wlsUNumEval
		if err := json.Unmarshal(e.Eval, &ue); err != nil {
			return nil, fmt.Errorf("invalid unsigned numeric evaluator: %w", err)
		}
		cmp, err := decodeNumCmp(ue.Cmp)
		if err != nil {
			return nil, err
		}
		return UNumericLeaf(ue.ID, cmp, ue.Value), nil
	default:
		return nil, fmt.Errorf("unsupported eval_type %q", e.EvalType)
	}
}

func decodeStrEval(e wlsStrEval) (*Node, error) {
	switch e.ID {
	case IDAlwaysTrue:
		return AlwaysTrue(), nil
	case IDAlwaysFalse:
		return AlwaysFalse(), nil
	case IDAlwaysAbstain:
		return AlwaysAbstain(), nil
	}
	cmp, err := decodeStrCmp(e.Cmp)
	if err != nil {
		return nil, err
	}
	if IsLabelID(e.ID) {
		return decodeLabel(e.ID, cmp, e.Value)
	}
	return StringLeaf(e.ID, cmp, e.Value), nil
}

func decodeLabel(id string, cmp StringCmp, raw string) (*Node, error) {
	key, value, found := strings.Cut(raw, "=")
	if !found {
		return nil, fmt.Errorf("label evaluator value %q must be encoded as key=value", raw)
	}
	// "key=" with a prefix comparison is the existence convention.
	if cmp == CmpPrefix && value == "" {
		return LabelLeaf(id, key, CmpExists, ""), nil
	}
	return LabelLeaf(id, key, cmp, value), nil
}

func decodeStrCmp(cmp string) (StringCmp, error) {
	switch cmp {
	case cmpExact:
		return CmpExact, nil
	case cmpPrefix:
		return CmpPrefix, nil
	case cmpSuffix:
		return CmpSuffix, nil
	case cmpContains:
		return CmpContains, nil
	case cmpWildcard:
		return CmpWildcard, nil
	default:
		return CmpExact, fmt.Errorf("unsupported string comparison %q", cmp)
	}
}

func decodeNumCmp(cmp string) (NumericCmp, error) {
	switch cmp {
	case cmpEq:
		return NumEq, nil
	case cmpGt:
		return NumGt, nil
	case cmpGte:
		return NumGte, nil
	case cmpLt:
		return NumLt, nil
	case cmpLte:
		return NumLte, nil
	default:
		return NumEq, fmt.Errorf("unsupported numeric comparison %q", cmp)
	}
}

// decodeUUID renders the dd-wls hi/lo pair as a canonical UUID string. It
// returns an empty string when the document carries no id.
func decodeUUID(id *wlsUUID) string {
	if id == nil {
		return ""
	}
	var b [16]byte
	binary.BigEndian.PutUint64(b[0:8], id.Hi)
	binary.BigEndian.PutUint64(b[8:16], id.Lo)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func decodeActions(actions []wlsAction) Outcome {
	out := Outcome{}
	for _, a := range actions {
		switch a.Action {
		case actionInjectAllow:
			out.Inject = true
			out.InjectSet = true
		case actionInjectDeny:
			out.Inject = false
			out.InjectSet = true
		case actionEnableSDK:
			for _, v := range a.Values {
				lang, version, _ := strings.Cut(v, "=")
				if out.TracerVersions == nil {
					out.TracerVersions = map[string]string{}
				}
				out.TracerVersions[lang] = version
			}
		case actionEnableProfiler:
			out.TracerConfigs = append(out.TracerConfigs, EnvVar{Name: "DD_PROFILING_ENABLED", Value: "true"})
		}
	}
	return out
}
