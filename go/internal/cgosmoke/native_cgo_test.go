//go:build cmake_cgo

// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache 2.0 License. This product includes software developed at
// Datadog (https://www.datadoghq.com/).
//
// Copyright 2025-Present Datadog, Inc.

package cgosmoke

import "testing"

func TestCMakeBuiltPolicyEngineIsLinked(t *testing.T) {
	if got := trueResultName(); got != "EVAL_RESULT_TRUE" {
		t.Fatalf("unexpected native evaluation result name: %q", got)
	}
}
