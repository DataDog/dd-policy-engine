//go:build cmake_cgo

// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache 2.0 License. This product includes software developed at
// Datadog (https://www.datadoghq.com/).
//
// Copyright 2025-Present Datadog, Inc.

package cgosmoke

/*
#include <dd/policies/policies.h>
*/
import "C"

func trueResultName() string {
	return C.GoString(C.plcs_evaluation_result_to_string(C.PLCS_EVAL_RESULT_TRUE))
}
