
// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache 2.0 License. This product includes software developed at
// Datadog (https://www.datadoghq.com/).
//
// Copyright 2025-Present Datadog, Inc.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Datadog/dd-policy-engine/go/cmd/dd-requirements-converter/converter"

	flatbuffers "github.com/google/flatbuffers/go"
)

func writeBufferToFile(buffer []byte, fileName string, outDir string) {
	binDir := filepath.Join(outDir, "out")

	err := os.MkdirAll(binDir, 0755)
	if err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return
	}

	binPath := filepath.Join(binDir, fileName)
	err = os.WriteFile(binPath, buffer, 0644)
	if err != nil {
		log.Fatalf("Failed to write buffer to file: %v", err)
	}
	fmt.Printf("Wrote %d bytes to: %s\n\n", len(buffer), fileName)
}

func generateCHeader(varName string, data []byte) string {
	var sb strings.Builder
	sb.WriteString("#pragma once\n\n")
	sb.WriteString("#include <stdint.h>\n\n")
	sb.WriteString(fmt.Sprintf("const uint8_t %s[] = {\n", varName))

	for i, b := range data {
		if i%12 == 0 {
			sb.WriteString("  ")
		}
		sb.WriteString(fmt.Sprintf("0x%02X", b))
		if i < len(data)-1 {
			sb.WriteString(", ")
		}
		if (i+1)%12 == 0 {
			sb.WriteString("\n")
		}
	}

	if len(data)%12 != 0 {
		sb.WriteString("\n")
	}

	sb.WriteString("};\n")
	sb.WriteString(fmt.Sprintf("const unsigned int %s_len = %d;\n", varName, len(data)))
	return sb.String()
}

func finalizePolicies(builder *flatbuffers.Builder, policies flatbuffers.UOffsetT, fileName string, outDir string) {

	builder.Finish(policies)
	buffer := builder.FinishedBytes()

	header := generateCHeader("hardcoded_policies", buffer)

	err := os.MkdirAll(outDir, 0755)
	if err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return
	}

	err = os.WriteFile(outDir+fileName+".h", []byte(header), 0644)
	if err != nil {
		log.Fatalf("Failed to write header file: %v", err)
	}

	writeBufferToFile(buffer, fileName+".bin", outDir)
}

func main() {
	// Parse command line arguments
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <input-json-file> <output-header-file>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s skips.json out/hardcoded.h\n", os.Args[0])
		os.Exit(1)
	}

	inputFile := os.Args[1]
	outputHeaderFile := os.Args[2]

	file, err := os.Open(inputFile)
	if err != nil {
		log.Fatalf("Failed to open input file %s: %v", inputFile, err)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("Failed to read input file: %v", err)
	}

	var requirements converter.JSONRequirements
	if err := json.Unmarshal(bytes, &requirements); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	builder := flatbuffers.NewBuilder(1024)

	offset, err := requirements.ConvertToWLS(builder)
	if err != nil {
		log.Fatalf("Failed to convert policies to WLS: %v", err)
	}

	// Extract directory and base filename from output path
	lastSlash := strings.LastIndex(outputHeaderFile, "/")
	var outDir, baseName string
	if lastSlash == -1 {
		outDir = "./"
		baseName = outputHeaderFile
	} else {
		outDir = outputHeaderFile[:lastSlash+1]
		baseName = outputHeaderFile[lastSlash+1:]
	}

	// Remove .h extension if present
	baseName = strings.TrimSuffix(baseName, ".h")

	finalizePolicies(builder, offset, baseName, outDir)

}
