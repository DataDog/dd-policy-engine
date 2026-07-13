// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache 2.0 License. This product includes software developed at
// Datadog (https://www.datadoghq.com/).
//
// Copyright 2025-Present Datadog, Inc.
package schema

import (
	"fmt"

	"github.com/DataDog/dd-policy-engine/go/schema/dd/wls"
)

// RuntimeLanguageToString maps RuntimeLanguage enum values to their canonical
// lowercase string representation used in policy evaluation.
var RuntimeLanguageToString = map[wls.RuntimeLanguage]string{
	wls.RuntimeLanguageJVM:    "jvm",
	wls.RuntimeLanguagePYTHON: "python",
	wls.RuntimeLanguageRUBY:   "ruby",
	wls.RuntimeLanguageDOTNET: "dotnet",
	wls.RuntimeLanguageNODEJS: "nodejs",
	wls.RuntimeLanguagePHP:    "php",
}

// RuntimeLanguageFromString maps canonical lowercase strings to RuntimeLanguage
// enum values. Use this to validate and convert user-provided runtime strings.
var RuntimeLanguageFromString = map[string]wls.RuntimeLanguage{
	"jvm":    wls.RuntimeLanguageJVM,
	"java":   wls.RuntimeLanguageJVM,
	"python": wls.RuntimeLanguagePYTHON,
	"ruby":   wls.RuntimeLanguageRUBY,
	"dotnet": wls.RuntimeLanguageDOTNET,
	"nodejs": wls.RuntimeLanguageNODEJS,
	"php":    wls.RuntimeLanguagePHP,
}

// OperatingSystemToString maps OperatingSystem enum values to their canonical
// lowercase string representation used in policy evaluation.
var OperatingSystemToString = map[wls.OperatingSystem]string{
	wls.OperatingSystemLINUX:   "linux",
	wls.OperatingSystemWINDOWS: "windows",
	wls.OperatingSystemMACOS:   "macos",
}

// OperatingSystemFromString maps canonical lowercase strings to OperatingSystem
// enum values. Use this to validate and convert user-provided OS strings.
var OperatingSystemFromString = map[string]wls.OperatingSystem{
	"linux":   wls.OperatingSystemLINUX,
	"windows": wls.OperatingSystemWINDOWS,
	"macos":   wls.OperatingSystemMACOS,
}

// MachineArchitectureToString maps MachineArchitecture enum values to their canonical
// string representation used in policy evaluation (matching kernel-reported values).
var MachineArchitectureToString = map[wls.MachineArchitecture]string{
	wls.MachineArchitectureAARCH64: "aarch64",
	wls.MachineArchitectureX86_64:  "x86_64",
	wls.MachineArchitectureARM:     "arm",
	wls.MachineArchitectureX86:     "x86",
}

// MachineArchitectureFromString maps various architecture strings to MachineArchitecture
// enum values. This handles the arm64/aarch64 and x64/x86_64/amd64 aliasing.
var MachineArchitectureFromString = map[string]wls.MachineArchitecture{
	"aarch64": wls.MachineArchitectureAARCH64,
	"arm64":   wls.MachineArchitectureAARCH64, // alias
	"x86_64":  wls.MachineArchitectureX86_64,
	"x64":     wls.MachineArchitectureX86_64, // alias
	"amd64":   wls.MachineArchitectureX86_64, // alias
	"arm":     wls.MachineArchitectureARM,
	"x86":     wls.MachineArchitectureX86,
}

// ValidateMachineArchitecture checks whether the given string is a recognized
// machine architecture value.
func ValidateMachineArchitecture(s string) error {
	if _, ok := MachineArchitectureFromString[s]; !ok {
		return fmt.Errorf("unknown machine architecture: %q", s)
	}
	return nil
}

// ValidateRuntimeLanguage checks whether the given string is a recognized
// runtime language value.
func ValidateRuntimeLanguage(s string) error {
	if _, ok := RuntimeLanguageFromString[s]; !ok {
		return fmt.Errorf("unknown runtime language: %q", s)
	}
	return nil
}

// ValidateOperatingSystem checks whether the given string is a recognized
// operating system value.
func ValidateOperatingSystem(s string) error {
	if _, ok := OperatingSystemFromString[s]; !ok {
		return fmt.Errorf("unknown operating system: %q", s)
	}
	return nil
}
