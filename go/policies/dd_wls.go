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
	actionSetEnvVar      = "SET_ENVAR"
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
	ID  string `json:"id"`
	Cmp string `json:"cmp"`
	// Value is a pointer so an omitted/null value is distinguishable from an
	// explicit empty string: the former is rejected for non-constant evaluators
	// (see decodeStrEval), the latter is a legal exact-empty match.
	Value *string `json:"value"`
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
//
// Rule decoding is total (see decodeNodeWrap): an unrecognized or malformed rule
// construct decodes to an ABSTAIN leaf rather than failing, so a policy produced
// by a newer schema stays evaluatable by an older agent. ParsePolicies returns
// an error only for a document that is not valid JSON or that carries a
// malformed value for an action the engine implements.
func ParsePolicies(raw []byte) ([]Policy, error) {
	var doc wlsPolicies
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("invalid policies document: %w", err)
	}

	out := make([]Policy, 0, len(doc.Policies))
	for i, p := range doc.Policies {
		outcome, err := decodeActions(p.Actions)
		if err != nil {
			return nil, fmt.Errorf("policy[%d] %q: %w", i, p.Description, err)
		}
		out = append(out, Policy{
			Name:    p.Description,
			ID:      decodeUUID(p.ID),
			Version: p.Version,
			Rules:   decodeNodeWrap(p.Rules),
			Outcome: outcome,
		})
	}
	return out, nil
}

// decodeNodeWrap decodes one rule node. It is total, mirroring the C engine's
// evaluate_rules, which always yields a tri-state and never fails: any node it
// cannot recognize or parse -- an unknown node_type, a malformed body, or a
// construct from a newer schema -- decodes to an ABSTAIN leaf rather than an
// error. This keeps a newer policy evaluatable by an older agent (forward
// compatibility, since policies can reach agents older than the schema that
// produced them) and stops one unrecognized node from sinking the whole
// document. Genuine correctness of a well-formed policy is enforced upstream
// (the compiler and the cross-engine conformance corpus), not by rejecting at
// runtime.
func decodeNodeWrap(w wlsNodeWrap) *Node {
	switch w.NodeType {
	case nodeComposite:
		var c wlsComposite
		if err := json.Unmarshal(w.Node, &c); err != nil {
			return AlwaysAbstain()
		}
		return decodeComposite(c)
	case nodeEvaluator:
		var e wlsEvaluatorNode
		if err := json.Unmarshal(w.Node, &e); err != nil {
			return AlwaysAbstain()
		}
		return decodeEvaluatorNode(e)
	default:
		return AlwaysAbstain()
	}
}

func decodeComposite(c wlsComposite) *Node {
	children := make([]*Node, 0, len(c.Children))
	for _, child := range c.Children {
		children = append(children, decodeNodeWrap(child))
	}
	switch c.Op {
	case opAnd:
		return &Node{Op: OpAnd, Children: children}
	case opOr:
		return &Node{Op: OpOr, Children: children}
	case opNot:
		// C abstains on a BOOL_NOT without exactly one child.
		if len(children) != 1 {
			return AlwaysAbstain()
		}
		return &Node{Op: OpNot, Children: children}
	default:
		return AlwaysAbstain()
	}
}

// decodeEvaluatorNode dispatches on the wire union member. The Go engine is a
// faithful, generic reimplementation of the C engine: it accepts any evaluator
// id and resolves it against the Context at evaluation time (an id with no
// matching fact abstains, like the C engine's NULL context), so it is not
// restricted to a Kubernetes subset. An unrecognized eval_type or comparator
// (e.g. from a newer schema) abstains rather than failing the parse.
func decodeEvaluatorNode(e wlsEvaluatorNode) *Node {
	switch e.EvalType {
	case evalString:
		var se wlsStrEval
		if err := json.Unmarshal(e.Eval, &se); err != nil {
			return AlwaysAbstain()
		}
		return decodeStrEval(se)
	case evalNumeric:
		var ne wlsNumEval
		if err := json.Unmarshal(e.Eval, &ne); err != nil {
			return AlwaysAbstain()
		}
		cmp, ok := decodeNumCmp(ne.Cmp)
		if !ok {
			return AlwaysAbstain()
		}
		return NumericLeaf(ne.ID, cmp, ne.Value)
	case evalUNumeric:
		var ue wlsUNumEval
		if err := json.Unmarshal(e.Eval, &ue); err != nil {
			return AlwaysAbstain()
		}
		cmp, ok := decodeNumCmp(ue.Cmp)
		if !ok {
			return AlwaysAbstain()
		}
		return UNumericLeaf(ue.ID, cmp, ue.Value)
	default:
		return AlwaysAbstain()
	}
}

func decodeStrEval(e wlsStrEval) *Node {
	switch e.ID {
	case IDAlwaysTrue:
		return AlwaysTrue()
	case IDAlwaysFalse:
		return AlwaysFalse()
	case IDAlwaysAbstain:
		return AlwaysAbstain()
	}
	// A missing/null value decodes to "" under encoding/json, which for
	// CMP_PREFIX/CMP_CONTAINS would match every present fact; the C engine reads a
	// NULL policy string and abstains, so mirror that. An explicit empty string
	// is kept (a legal exact-empty comparison, matching C's non-NULL "").
	if e.Value == nil {
		return AlwaysAbstain()
	}
	cmp, ok := decodeStrCmp(e.Cmp)
	if !ok {
		return AlwaysAbstain()
	}
	if IsLabelID(e.ID) {
		return decodeLabel(e.ID, cmp, *e.Value)
	}
	return StringLeaf(e.ID, cmp, *e.Value)
}

func decodeLabel(id string, cmp StringCmp, raw string) *Node {
	key, value, found := strings.Cut(raw, "=")
	if !found {
		// Malformed label value (no "="): nothing meaningful to compare, abstain.
		return AlwaysAbstain()
	}
	// "key=" with a prefix comparison is the existence convention.
	if cmp == CmpPrefix && value == "" {
		return LabelLeaf(id, key, CmpExists, "")
	}
	return LabelLeaf(id, key, cmp, value)
}

func decodeStrCmp(cmp string) (StringCmp, bool) {
	switch cmp {
	case cmpExact:
		return CmpExact, true
	case cmpPrefix:
		return CmpPrefix, true
	case cmpSuffix:
		return CmpSuffix, true
	case cmpContains:
		return CmpContains, true
	case cmpWildcard:
		return CmpWildcard, true
	default:
		return CmpExact, false
	}
}

func decodeNumCmp(cmp string) (NumericCmp, bool) {
	switch cmp {
	case cmpEq:
		return NumEq, true
	case cmpGt:
		return NumGt, true
	case cmpGte:
		return NumGte, true
	case cmpLt:
		return NumLt, true
	case cmpLte:
		return NumLte, true
	default:
		return NumEq, false
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

func decodeActions(actions []wlsAction) (Outcome, error) {
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
				lang, version, found := strings.Cut(v, "=")
				if !found || lang == "" {
					return Outcome{}, fmt.Errorf("ENABLE_SDK value %q must be encoded as lang=version", v)
				}
				if out.TracerVersions == nil {
					out.TracerVersions = map[string]string{}
				}
				out.TracerVersions[lang] = version
			}
		case actionEnableProfiler:
			out.TracerConfigs = upsertEnvVar(out.TracerConfigs, "DD_PROFILING_ENABLED", "true")
		case actionSetEnvVar:
			for _, v := range a.Values {
				name, value, found := strings.Cut(v, "=")
				if !found || name == "" {
					return Outcome{}, fmt.Errorf("SET_ENVAR value %q must be encoded as NAME=value", v)
				}
				out.TracerConfigs = upsertEnvVar(out.TracerConfigs, name, value)
			}
		default:
			// An action the engine does not implement (e.g. REEXEC, or a future
			// ActionId). Ignore it for forward compatibility, mirroring the C
			// engine, whose perform_actions skips actions with no registered
			// handler rather than failing. This keeps a newer policy evaluatable by
			// an older agent instead of dropping the whole document. Note: this is
			// only about unhandled action *ids*; malformed values of actions we do
			// handle (e.g. ENABLE_SDK, SET_ENVAR) are still rejected above. Add a
			// case here when support for a new action lands.
		}
	}
	return out, nil
}
