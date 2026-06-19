// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache 2.0 License. This product includes software developed at
// Datadog (https://www.datadoghq.com/).
//
// Copyright 2025-Present Datadog, Inc.

// Package policies is a self-contained, dependency-free policy model and
// tri-state evaluator written in pure Go.
//
// It mirrors the semantics of the dd-policy-engine C evaluator (TRUE / FALSE /
// ABSTAIN over an AND/OR/NOT tree of leaf evaluators), so consumers such as the
// Datadog Cluster Agent can evaluate policies natively without CGO while staying
// semantically identical to the host injector's C engine.
//
// Policies are produced either by parsing a dd-wls document (ParsePolicies), the
// JSON projection of the FlatBuffers policy schema shared with the C engine, or
// by building rule trees programmatically with the exported node constructors
// (And, Or, Not, Leaf, ...). The latter lets a caller lower a friendlier surface
// (for example Kubernetes "targets") into the policy model without this package
// needing any knowledge of that surface.
package policies
