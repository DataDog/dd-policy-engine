// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache 2.0 License. This product includes software developed at
// Datadog (https://www.datadoghq.com/).
//
// Copyright 2025-Present Datadog, Inc.

// Package policies is a self-contained, dependency-free policy model and
// tri-state evaluator written in pure Go.
//
// It is a faithful, generic reimplementation of the dd-policy-engine C
// evaluator (TRUE / FALSE / ABSTAIN over an AND/OR/NOT tree of leaf
// evaluators), so consumers can evaluate policies natively without CGO while
// staying semantically identical to the host injector's C engine. It is not
// restricted to any environment: leaf evaluators are identified by their wire
// evaluator id (the full StringEvaluators and NumericEvaluators id space) and
// resolved against a generic Context, mirroring the C engine's per-id value
// registry. An id with no matching fact in the Context abstains, exactly as the
// C engine returns ABSTAIN on a NULL context. The purely mechanical parts of
// the C engine that carry no value in Go (action callbacks, function-pointer
// registration) are not reproduced; each policy carries a decoded Outcome
// instead. Evaluate returns a single policy's tri-state result and its Outcome
// is read directly; combining the outcomes of several matching policies
// (which policy wins, and in what order) is deliberately left to the consumer,
// mirroring how the C host accumulates action side effects across policies
// rather than the engine doing it.
//
// String, signed-numeric and unsigned-numeric evaluators are all supported. The
// one intentional enhancement over C is label-type ids (POD_LABEL,
// NAMESPACE_LABEL, ...), which resolve against a real key->value map instead of
// the C single-string-per-id limitation.
//
// Policies are produced either by parsing a dd-wls document (ParsePolicies), the
// JSON projection of the FlatBuffers policy schema shared with the C engine, or
// by building rule trees programmatically with the exported node constructors
// (And, Or, Not, StringLeaf, LabelLeaf, NumericLeaf, UNumericLeaf, ...). The
// latter lets a caller lower a friendlier surface (for example Kubernetes
// "targets") into the policy model without this package needing any knowledge
// of that surface.
package policies
