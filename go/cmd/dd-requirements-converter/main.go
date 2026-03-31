// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache 2.0 License. This product includes software developed at
// Datadog (https://www.datadoghq.com/).
//
// Copyright 2025-Present Datadog, Inc.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/DataDog/dd-policy-engine/go/cmd/dd-requirements-converter/converter"

	flatbuffers "github.com/google/flatbuffers/go"
)

// parseRequirementsJSON decodes requirements document bytes the same way as the CLI: rejects
// empty/whitespace-only input and invalid JSON before conversion.
func parseRequirementsJSON(raw []byte) (converter.JSONRequirements, error) {
	var req converter.JSONRequirements
	if len(bytes.TrimSpace(raw)) == 0 {
		return req, fmt.Errorf("requirements input is empty or whitespace-only")
	}
	if err := json.Unmarshal(raw, &req); err != nil {
		return req, fmt.Errorf("invalid requirements JSON: %w", err)
	}
	if req.Version != 1 {
		return req, fmt.Errorf("requirements.bin version must be 1, got %d", req.Version)
	}
	return req, nil
}

func main() {
	inputFile := flag.String("input-file", "", "Input JSON file path (required)")
	outputFile := flag.String("output-file", "", "Output binary file path (required)")
	flag.Parse()

	if *inputFile == "" || *outputFile == "" {
		fmt.Fprintf(os.Stderr, "Error: --input-file and --output-file are required.\n\n")
		flag.Usage()
		os.Exit(1)
	}

	file, err := os.Open(*inputFile)
	if err != nil {
		log.Fatalf("Failed to open input file %s: %v", *inputFile, err)
	}
	defer file.Close()

	raw, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("failed to read input file %s: %v", *inputFile, err)
	}

	requirements, err := parseRequirementsJSON(raw)
	if err != nil {
		log.Fatalf("invalid requirements in %s: %v", *inputFile, err)
	}

	builder := flatbuffers.NewBuilder(1024)

	policies, err := requirements.ConvertToWLS(builder)
	if err != nil {
		log.Fatalf("Failed to convert policies to WLS: %v", err)
	}

	builder.Finish(policies)
	buffer := builder.FinishedBytes()

	if err := os.WriteFile(*outputFile, buffer, 0644); err != nil {
		log.Fatalf("Failed to write output: %v", err)
	}
	fmt.Printf("Wrote %d bytes to %s\n", len(buffer), *outputFile)
}
