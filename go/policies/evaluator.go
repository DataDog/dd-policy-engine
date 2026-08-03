// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache 2.0 License. This product includes software developed at
// Datadog (https://www.datadoghq.com/).
//
// Copyright 2025-Present Datadog, Inc.

package policies

import (
	"math"
	"strings"
)

// maxEvalDepth bounds the rule-tree recursion, mirroring PLCS_MAX_EVAL_DEPTH in
// the C engine (c/src/evaluator.c). A node deeper than this evaluates to
// ResultAbstain instead of recursing further.
const maxEvalDepth = 64

// numNotSet and unumNotSet are the reserved "not set" sentinels: a fact equal to
// one is treated as unavailable (ABSTAIN), never compared. They are fixed to the
// 64-bit maxima to match the FlatBuffers wire value width (NumEvaluator.value is
// `long` and UNumEvaluator.value is `ulong`, both 64-bit) and the C engine's
// PLCS_NUM_NOT_SET / PLCS_UNUM_NOT_SET.
//
// This is a fixed-width 64-bit contract. It matches the C engine exactly on LP64
// platforms (Linux), where C's long/unsigned long are 64-bit so LONG_MAX/ULONG_MAX
// equal these. On LLP64 (MSVC/Windows) C's long is 32-bit, so its sentinels become
// MaxInt32/MaxUint32 and it truncates 64-bit values; C parity there requires the C
// engine to use fixed-width int64_t/uint64_t. Go's facts are int64/uint64 on every
// platform, so the sentinels are deliberately NOT target-specific.
const (
	numNotSet  int64  = math.MaxInt64
	unumNotSet uint64 = math.MaxUint64
)

// Context carries the workload facts a policy is evaluated against, keyed by
// evaluator id. It mirrors the C engine's per-id value registry: a fact that is
// not present means the source is unavailable in this environment and the leaf
// evaluates to ResultAbstain (like the C engine's NULL context), rather than
// comparing against a zero value.
//
// Keyed KEY=VALUE ids -- labels, annotations, and process environment variables
// (see IsLabelID) -- read from Labels (a real key->value map, the Go enhancement
// over the C single-string-per-id model); list ids -- unpositioned PROCESS_ARGV
// (see IsListID) -- read from Lists and match if any element matches; all other
// string ids read from Strings; numeric ids read from Numbers (signed) or
// UNumbers (unsigned) depending on the evaluator kind.
type Context struct {
	Strings map[string]string
	Labels  map[string]map[string]string
	// Lists holds multi-valued string facts (e.g. an unpositioned PROCESS_ARGV's
	// full argv). A list evaluator is true when any element matches; a missing
	// entry is an unavailable source (ABSTAIN), a present-but-empty list is FALSE.
	Lists map[string][]string
	// Numbers and UNumbers hold signed/unsigned numeric facts. math.MaxInt64 and
	// math.MaxUint64 are reserved: they are the C engine's not-set sentinels, so a
	// fact equal to one evaluates to ResultAbstain (as if absent), never compared.
	Numbers  map[string]int64
	UNumbers map[string]uint64
}

// upsertEnvVar sets name=value in configs, replacing any existing entry with the
// same name rather than appending a duplicate (a later value wins). Environment
// variable names are unique by definition, so the same config is never emitted
// twice even when a policy repeats an action that sets it.
func upsertEnvVar(configs []EnvVar, name, value string) []EnvVar {
	for i := range configs {
		if configs[i].Name == name {
			configs[i].Value = value
			return configs
		}
	}
	return append(configs, EnvVar{Name: name, Value: value})
}

// Evaluate walks the rule tree and returns its tri-state result.
func Evaluate(n *Node, ctx Context) Result {
	return evaluate(n, ctx, 0)
}

// evaluate is the depth-tracking core of Evaluate. depth mirrors the C engine:
// the root is evaluated at depth 0 and a node deeper than maxEvalDepth abstains.
func evaluate(n *Node, ctx Context, depth int) Result {
	if depth > maxEvalDepth {
		return ResultAbstain
	}
	if n == nil {
		return ResultAbstain
	}
	if n.Eval != nil {
		return n.Eval.eval(ctx)
	}
	switch n.Op {
	case OpNot:
		if len(n.Children) != 1 {
			return ResultAbstain
		}
		return doNot(evaluate(n.Children[0], ctx, depth+1))
	case OpOr:
		res := ResultFalse
		for _, c := range n.Children {
			res = doOr(res, evaluate(c, ctx, depth+1))
			if res == ResultTrue {
				return res
			}
		}
		return res
	case OpAnd:
		res := ResultTrue
		for _, c := range n.Children {
			res = doAnd(res, evaluate(c, ctx, depth+1))
			if res == ResultFalse {
				return res
			}
		}
		return res
	default:
		return ResultAbstain
	}
}

// doAnd implements tri-state AND: false dominates, abstain is contagious among
// non-false operands.
func doAnd(a, b Result) Result {
	if a == ResultFalse || b == ResultFalse {
		return ResultFalse
	}
	if a != ResultAbstain && b != ResultAbstain {
		return ResultTrue
	}
	return ResultAbstain
}

// doOr implements tri-state OR: true dominates, abstain is contagious among
// non-true operands.
func doOr(a, b Result) Result {
	if a == ResultTrue || b == ResultTrue {
		return ResultTrue
	}
	if a != ResultAbstain && b != ResultAbstain {
		return ResultFalse
	}
	return ResultAbstain
}

// doNot flips true/false and leaves abstain unchanged.
func doNot(a Result) Result {
	switch a {
	case ResultTrue:
		return ResultFalse
	case ResultFalse:
		return ResultTrue
	default:
		return ResultAbstain
	}
}

func (e *Evaluator) eval(ctx Context) Result {
	switch e.Kind {
	case EvalString:
		return e.evalString(ctx)
	case EvalNumeric:
		v, ok := ctx.Numbers[e.ID]
		if !ok || v == numNotSet {
			// Absent, or equal to the C engine's PLCS_NUM_NOT_SET sentinel, which
			// the C engine treats as unset -> ABSTAIN rather than comparing.
			return ResultAbstain
		}
		return compareNumeric(e.NumCmp, e.NumValue, v)
	case EvalUNumeric:
		v, ok := ctx.UNumbers[e.ID]
		if !ok || v == unumNotSet {
			// Absent, or equal to the C engine's PLCS_UNUM_NOT_SET sentinel.
			return ResultAbstain
		}
		return compareNumeric(e.NumCmp, e.UNumValue, v)
	default:
		return ResultAbstain
	}
}

func (e *Evaluator) evalString(ctx Context) Result {
	switch e.ID {
	case IDAlwaysTrue:
		return ResultTrue
	case IDAlwaysFalse:
		return ResultFalse
	case IDAlwaysAbstain:
		return ResultAbstain
	}
	if IsLabelID(e.ID) {
		labels, ok := ctx.Labels[e.ID]
		if !ok {
			// The label source itself is unavailable in this environment.
			return ResultAbstain
		}
		v, present := labels[e.Key]
		return labelResult(e.StrCmp, e.StrValue, v, present)
	}
	if IsListID(e.ID) {
		values, ok := ctx.Lists[e.ID]
		if !ok {
			// The list source itself is unavailable in this environment.
			return ResultAbstain
		}
		// A list evaluator matches when any element matches (e.g. an unpositioned
		// argument pattern against the full argv); present but no match is false.
		for _, v := range values {
			if compareString(e.StrCmp, e.StrValue, v) {
				return ResultTrue
			}
		}
		return ResultFalse
	}
	v, ok := ctx.Strings[e.ID]
	if !ok {
		// Source unavailable here: mirrors the C engine's NULL context.
		return ResultAbstain
	}
	return boolToResult(compareString(e.StrCmp, e.StrValue, v))
}

// labelResult resolves a label-keyed evaluator against a present label source.
// A missing label key is false for every comparison except CmpExists (which
// reports presence). This matches Kubernetes label-selector semantics once
// composed with the tree operators: In/matchLabels require presence, while
// NotIn/DoesNotExist match absent keys via the surrounding NOT.
func labelResult(cmp StringCmp, want, got string, present bool) Result {
	if cmp == CmpExists {
		return boolToResult(present)
	}
	if !present {
		return ResultFalse
	}
	return boolToResult(compareString(cmp, want, got))
}

func compareString(cmp StringCmp, pattern, value string) bool {
	switch cmp {
	case CmpExact:
		return pattern == value
	case CmpPrefix:
		return strings.HasPrefix(value, pattern)
	case CmpSuffix:
		return strings.HasSuffix(value, pattern)
	case CmpContains:
		return strings.Contains(value, pattern)
	case CmpWildcard:
		return wildcardMatch(pattern, value)
	case CmpExists:
		return true
	default:
		return false
	}
}

// compareNumeric mirrors plcs_default_numeric_evaluator: it computes
// "evaluator value <cmp> workload value".
func compareNumeric[T int64 | uint64](cmp NumericCmp, policy, workload T) Result {
	switch cmp {
	case NumEq:
		return boolToResult(policy == workload)
	case NumGt:
		return boolToResult(policy > workload)
	case NumGte:
		return boolToResult(policy >= workload)
	case NumLt:
		return boolToResult(policy < workload)
	case NumLte:
		return boolToResult(policy <= workload)
	default:
		return ResultAbstain
	}
}

func boolToResult(b bool) Result {
	if b {
		return ResultTrue
	}
	return ResultFalse
}

// wildcardMatch reports whether s matches the glob pattern, where '*' matches
// any run of characters and '?' matches a single character. It is a linear
// two-pointer matcher operating on bytes, sufficient for label/namespace values.
func wildcardMatch(pattern, s string) bool {
	starPat, starStr := -1, -1
	p, i := 0, 0
	for i < len(s) {
		if p < len(pattern) && (pattern[p] == s[i] || pattern[p] == '?') {
			p++
			i++
			continue
		}
		if p < len(pattern) && pattern[p] == '*' {
			for p < len(pattern) && pattern[p] == '*' {
				p++
			}
			if p == len(pattern) {
				return true
			}
			starPat = p
			starStr = i
			continue
		}
		if starPat != -1 {
			p = starPat
			starStr++
			i = starStr
			continue
		}
		return false
	}
	for p < len(pattern) && pattern[p] == '*' {
		p++
	}
	return p == len(pattern)
}
