// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache 2.0 License. This product includes software developed at
// Datadog (https://www.datadoghq.com/).
//
// Copyright 2025-Present Datadog, Inc.
package schema

import (
	"testing"

	"github.com/DataDog/dd-policy-engine/go/schema/dd/wls"
	"github.com/stretchr/testify/assert"
)

func TestRuntimeLanguageToString(t *testing.T) {
	tests := []struct {
		name     string
		enum     wls.RuntimeLanguage
		expected string
	}{
		{"JVM", wls.RuntimeLanguageJVM, "jvm"},
		{"PYTHON", wls.RuntimeLanguagePYTHON, "python"},
		{"RUBY", wls.RuntimeLanguageRUBY, "ruby"},
		{"DOTNET", wls.RuntimeLanguageDOTNET, "dotnet"},
		{"NODEJS", wls.RuntimeLanguageNODEJS, "nodejs"},
		{"PHP", wls.RuntimeLanguagePHP, "php"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, RuntimeLanguageToString[tt.enum])
		})
	}
}

func TestRuntimeLanguageFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected wls.RuntimeLanguage
		ok       bool
	}{
		{"jvm", "jvm", wls.RuntimeLanguageJVM, true},
		{"java alias", "java", wls.RuntimeLanguageJVM, true},
		{"python", "python", wls.RuntimeLanguagePYTHON, true},
		{"ruby", "ruby", wls.RuntimeLanguageRUBY, true},
		{"dotnet", "dotnet", wls.RuntimeLanguageDOTNET, true},
		{"nodejs", "nodejs", wls.RuntimeLanguageNODEJS, true},
		{"php", "php", wls.RuntimeLanguagePHP, true},
		{"unknown", "go", wls.RuntimeLanguageRUNTIME_LANGUAGE_UNKNOWN, false},
		{"empty", "", wls.RuntimeLanguageRUNTIME_LANGUAGE_UNKNOWN, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, ok := RuntimeLanguageFromString[tt.input]
			assert.Equal(t, tt.ok, ok)
			if ok {
				assert.Equal(t, tt.expected, val)
			}
		})
	}
}

func TestOperatingSystemToString(t *testing.T) {
	tests := []struct {
		name     string
		enum     wls.OperatingSystem
		expected string
	}{
		{"LINUX", wls.OperatingSystemLINUX, "linux"},
		{"WINDOWS", wls.OperatingSystemWINDOWS, "windows"},
		{"MACOS", wls.OperatingSystemMACOS, "macos"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, OperatingSystemToString[tt.enum])
		})
	}
}

func TestOperatingSystemFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected wls.OperatingSystem
		ok       bool
	}{
		{"linux", "linux", wls.OperatingSystemLINUX, true},
		{"windows", "windows", wls.OperatingSystemWINDOWS, true},
		{"macos", "macos", wls.OperatingSystemMACOS, true},
		{"unknown", "freebsd", wls.OperatingSystemOS_UNKNOWN, false},
		{"empty", "", wls.OperatingSystemOS_UNKNOWN, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, ok := OperatingSystemFromString[tt.input]
			assert.Equal(t, tt.ok, ok)
			if ok {
				assert.Equal(t, tt.expected, val)
			}
		})
	}
}

func TestValidateRuntimeLanguage(t *testing.T) {
	assert.NoError(t, ValidateRuntimeLanguage("jvm"))
	assert.NoError(t, ValidateRuntimeLanguage("java"))
	assert.NoError(t, ValidateRuntimeLanguage("python"))
	assert.Error(t, ValidateRuntimeLanguage("go"))
	assert.Error(t, ValidateRuntimeLanguage(""))
}

func TestValidateOperatingSystem(t *testing.T) {
	assert.NoError(t, ValidateOperatingSystem("linux"))
	assert.NoError(t, ValidateOperatingSystem("windows"))
	assert.Error(t, ValidateOperatingSystem("freebsd"))
	assert.Error(t, ValidateOperatingSystem(""))
}
