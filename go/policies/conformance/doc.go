// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache 2.0 License. This product includes software developed at
// Datadog (https://www.datadoghq.com/).
//
// Copyright 2025-Present Datadog, Inc.

// Package conformance holds the cross-engine conformance suite for the Go
// policy evaluator. It is intentionally kept out of the importable "policies"
// package so that the cgo bridge to the C engine (and its FlatBuffers/schema
// dependencies) never leaks into consumers' module graphs: the cgo bridge must
// live in a regular, non-test source file (cgo is not allowed in _test.go
// files), and any import it carries would otherwise become a dependency of the
// public "policies" package.
//
// The suite has two halves sharing one corpus (testdata/vectors.json):
//   - the portable, Go-only runner (corpus_test.go), which runs everywhere;
//   - the cross-engine runner (cross_test.go + cgo.go), gated behind the
//     "conformance_cgo" build tag, which compares the Go engine against the
//     real C engine (libpolicies, via cgo).
package conformance
