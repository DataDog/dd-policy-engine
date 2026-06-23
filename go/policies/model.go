// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache 2.0 License. This product includes software developed at
// Datadog (https://www.datadoghq.com/).
//
// Copyright 2025-Present Datadog, Inc.

package policies

// Result is the tri-state outcome of evaluating a rule node, mirroring the
// dd-policy-engine C engine.
type Result uint8

const (
	// ResultFalse means the node evaluated to false.
	ResultFalse Result = iota
	// ResultTrue means the node evaluated to true.
	ResultTrue
	// ResultAbstain means the node could not produce a decision (e.g. an
	// unknown evaluator, or a fact source unavailable in this environment).
	ResultAbstain
)

// BoolOp is the boolean operator of a composite node.
type BoolOp uint8

const (
	// OpAnd is logical AND over the children.
	OpAnd BoolOp = iota
	// OpOr is logical OR over the children.
	OpOr
	// OpNot is logical NOT over a single child.
	OpNot
)

// EvalKind selects the value family of a leaf evaluator, mirroring the three
// evaluator tables of the FlatBuffers schema (StrEvaluator, NumEvaluator,
// UNumEvaluator).
type EvalKind uint8

const (
	// EvalString is a string evaluator: it compares StrValue against the
	// workload's string fact for ID using StrCmp.
	EvalString EvalKind = iota
	// EvalNumeric is a signed numeric evaluator: it compares NumValue against
	// the workload's int64 fact for ID using NumCmp.
	EvalNumeric
	// EvalUNumeric is an unsigned numeric evaluator: it compares UNumValue
	// against the workload's uint64 fact for ID using NumCmp.
	EvalUNumeric
)

// StringCmp is the comparison applied by a string evaluator between its value
// and the fact read from the workload.
type StringCmp uint8

const (
	// CmpExact is string equality.
	CmpExact StringCmp = iota
	// CmpPrefix is true when the fact starts with the evaluator value.
	CmpPrefix
	// CmpSuffix is true when the fact ends with the evaluator value.
	CmpSuffix
	// CmpContains is true when the fact contains the evaluator value.
	CmpContains
	// CmpWildcard is glob matching (* and ?) of the fact against the value.
	CmpWildcard
	// CmpExists is true when the keyed label fact is present, ignoring the
	// value. It is not a wire comparator: the dd-wls "KEY=" + CMP_PREFIX
	// existence convention is decoded into it.
	CmpExists
)

// NumericCmp is the comparison applied by a numeric evaluator. Mirroring the C
// engine, the comparison reads "evaluator value <cmp> workload value" (e.g.
// NumGt is true when the evaluator value is greater than the workload value).
type NumericCmp uint8

const (
	// NumEq is ==.
	NumEq NumericCmp = iota
	// NumGt is >.
	NumGt
	// NumGte is >=.
	NumGte
	// NumLt is <.
	NumLt
	// NumLte is <=.
	NumLte
)

// Well-known evaluator id names. The engine accepts any id name (it is resolved
// generically against the Context), but these are the ones with dedicated
// handling or commonly used by callers.
const (
	// IDAlwaysTrue, IDAlwaysFalse and IDAlwaysAbstain are constant evaluators
	// that ignore the Context and return a fixed result.
	IDAlwaysTrue    = "ALWAYS_TRUE"
	IDAlwaysFalse   = "ALWAYS_FALSE"
	IDAlwaysAbstain = "ALWAYS_ABSTAIN"

	// IDNamespaceName is the workload namespace name (a plain string fact).
	IDNamespaceName = "NAMESPACE_NAME"

	// Label-type ids: their value follows the "KEY=VALUE" convention and is
	// resolved against a key->value map in the Context.
	IDNamespaceLabel = "NAMESPACE_LABEL"
	IDPodLabel       = "POD_LABEL"
	IDPodAnnotation  = "POD_ANNOTATION"
	IDContainerLabel = "CONTAINER_LABEL"
)

// labelIDs is the set of string evaluator ids that carry the "KEY=VALUE"
// convention and are resolved against a key->value map instead of a plain
// string. This is the one place the Go engine intentionally improves on the C
// engine, which holds a single string per evaluator id (no native multi-label
// lookup); the dynamic-key map is the documented Go enhancement.
var labelIDs = map[string]struct{}{
	IDNamespaceLabel: {},
	IDPodLabel:       {},
	IDPodAnnotation:  {},
	IDContainerLabel: {},
}

// IsLabelID reports whether id uses the label "KEY=VALUE" convention.
func IsLabelID(id string) bool {
	_, ok := labelIDs[id]
	return ok
}

// Node is a node in the rule tree. It is either a composite node (Eval == nil,
// using Op and Children) or a leaf (Eval != nil).
type Node struct {
	Op       BoolOp
	Children []*Node
	Eval     *Evaluator
}

// Evaluator is a leaf condition. It reads the workload fact identified by ID
// from the Context and compares it to the evaluator's value. Kind selects which
// value/comparator/fact family is used.
type Evaluator struct {
	Kind EvalKind
	// ID is the canonical wire evaluator id name (e.g. "NAMESPACE_NAME",
	// "RUNTIME_LANGUAGE", "JAVA_HEAP").
	ID string
	// Key is the label key, set only for label-type string evaluators.
	Key string

	// StrCmp / StrValue are used when Kind == EvalString.
	StrCmp   StringCmp
	StrValue string

	// NumCmp is used when Kind == EvalNumeric or EvalUNumeric.
	NumCmp NumericCmp
	// NumValue is used when Kind == EvalNumeric.
	NumValue int64
	// UNumValue is used when Kind == EvalUNumeric.
	UNumValue uint64
}

// EnvVar is a tracer configuration environment variable returned by a matched
// policy.
type EnvVar struct {
	Name  string
	Value string
}

// Outcome is the configuration applied when a policy matches.
type Outcome struct {
	// Inject reports whether a matched workload should be instrumented. It is
	// true for an allow decision and false for a deny. Inject is only
	// meaningful when InjectSet is true.
	Inject bool
	// InjectSet reports whether the policy carried an explicit inject decision
	// (INJECT_ALLOW or INJECT_DENY). It lets Decide fold multiple matching
	// policies without a policy that omits an inject action silently flipping
	// the decision to deny.
	InjectSet bool
	// TracerVersions maps a tracer name to the version to inject.
	TracerVersions map[string]string
	// TracerConfigs are extra environment variables added alongside the tracer.
	TracerConfigs []EnvVar
}

// Policy pairs a rule tree with the outcome applied when the rule is true.
type Policy struct {
	Name string
	// ID is the canonical UUID of the policy, when the document carries one.
	// It is empty otherwise.
	ID      string
	Version int64
	Rules   *Node
	Outcome Outcome
}

// StringLeaf builds a string evaluator leaf reading the workload string fact
// identified by id and comparing it to value with cmp.
func StringLeaf(id string, cmp StringCmp, value string) *Node {
	return &Node{Eval: &Evaluator{Kind: EvalString, ID: id, StrCmp: cmp, StrValue: value}}
}

// LabelLeaf builds a label-type string evaluator leaf reading label key from
// the workload's key->value map for id and comparing it to value with cmp. Use
// CmpExists to test presence regardless of value.
func LabelLeaf(id, key string, cmp StringCmp, value string) *Node {
	return &Node{Eval: &Evaluator{Kind: EvalString, ID: id, Key: key, StrCmp: cmp, StrValue: value}}
}

// NumericLeaf builds a signed numeric evaluator leaf reading the workload int64
// fact identified by id and comparing it to value with cmp.
func NumericLeaf(id string, cmp NumericCmp, value int64) *Node {
	return &Node{Eval: &Evaluator{Kind: EvalNumeric, ID: id, NumCmp: cmp, NumValue: value}}
}

// UNumericLeaf builds an unsigned numeric evaluator leaf reading the workload
// uint64 fact identified by id and comparing it to value with cmp.
func UNumericLeaf(id string, cmp NumericCmp, value uint64) *Node {
	return &Node{Eval: &Evaluator{Kind: EvalUNumeric, ID: id, NumCmp: cmp, UNumValue: value}}
}

// AlwaysTrue builds a leaf node that always evaluates to ResultTrue.
func AlwaysTrue() *Node { return StringLeaf(IDAlwaysTrue, CmpExact, "") }

// AlwaysFalse builds a leaf node that always evaluates to ResultFalse.
func AlwaysFalse() *Node { return StringLeaf(IDAlwaysFalse, CmpExact, "") }

// AlwaysAbstain builds a leaf node that always evaluates to ResultAbstain.
func AlwaysAbstain() *Node { return StringLeaf(IDAlwaysAbstain, CmpExact, "") }

// And combines conds with a logical AND. An empty list matches everything
// (AlwaysTrue) and a single condition is returned as-is.
func And(conds []*Node) *Node {
	switch len(conds) {
	case 0:
		return AlwaysTrue()
	case 1:
		return conds[0]
	default:
		return &Node{Op: OpAnd, Children: conds}
	}
}

// Or combines conds with a logical OR. An empty list matches nothing
// (AlwaysFalse) and a single condition is returned as-is.
func Or(conds []*Node) *Node {
	switch len(conds) {
	case 0:
		return AlwaysFalse()
	case 1:
		return conds[0]
	default:
		return &Node{Op: OpOr, Children: conds}
	}
}

// Not negates a single node.
func Not(n *Node) *Node { return &Node{Op: OpNot, Children: []*Node{n}} }
