// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache 2.0 License. This product includes software developed at
// Datadog (https://www.datadoghq.com/).
//
// Copyright 2025-Present Datadog, Inc.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/DataDog/dd-policy-engine/go/cmd/dd-requirements-converter/converter"

	flatbuffers "github.com/google/flatbuffers/go"
)

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

	bytes, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("Failed to read input file: %v", err)
	}

	var requirements converter.JSONRequirements
	if err := json.Unmarshal(bytes, &requirements); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
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
