// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fixtures

import (
	"context"
	"fmt"
	"os"
	"time"

	log "github.com/golang/glog"
	"github.com/google/cql"
	"github.com/google/cql/result"
	"github.com/google/cql/retriever/local"
)

// Example_radiologyAnomalyDetection demonstrates using CQL for AI-assisted clinical decision support in radiology.
func Example_radiologyAnomalyDetection() {
	// Load FHIR data model and helpers
	fhirDataModel, fhirHelpers, err := cql.FHIRDataModelAndHelpersLib("4.0.1")
	if err != nil {
		log.Fatal(err)
	}

	// Read the CQL library from file
	cqlContent, err := os.ReadFile("radiology_anomaly_detection.cql")
	if err != nil {
		log.Fatalf("Failed to read CQL file: %v", err)
	}

	// Read the FHIR bundle from file
	bundleData, err := os.ReadFile("radiology_sample_bundle.json")
	if err != nil {
		log.Fatalf("Failed to read bundle file: %v", err)
	}

	libs := []string{string(cqlContent), fhirHelpers}

	// Parse the libraries with FHIR data model
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	elm, err := cql.Parse(ctx, libs, cql.ParseConfig{DataModels: [][]byte{fhirDataModel}})
	if err != nil {
		log.Fatalf("Failed to parse radiology CQL: %v", err)
	}

	// Use local retriever with the sample bundle
	retriever, err := local.NewRetrieverFromR4Bundle(bundleData)
	if err != nil {
		log.Fatalf("Failed to build retriever: %v", err)
	}

	// Evaluate the CQL
	results, err := elm.Eval(ctx, retriever, cql.EvalConfig{})
	if err != nil {
		log.Fatalf("Failed to evaluate: %v", err)
	}

	// Print the final decision
	decision := results[result.LibKey{Name: "RadiologyAnomalyDetection", Version: "1.0.0"}]["FinalDecision"]
	if result.IsNull(decision) {
		fmt.Println("FinalDecision: null")
	} else {
		str, err := result.ToString(decision)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("FinalDecision: %s\n", str)
	}

	// Output:
	// FinalDecision: Prioritize for Urgent AI-Assisted Review
}
