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

package cql_test

import (
	"context"
	"fmt"
	"time"

	log "github.com/golang/glog"
	"github.com/google/cql"
	"github.com/google/cql/result"
	"github.com/google/cql/retriever/local"
	"github.com/lithammer/dedent"
)

// This example demonstrates the CQL API by finding Observations that were effective during a
// measurement period.
func Example() {
	// CQL can run on different data models such as FHIR or QDM. The data model defines the CQL Named
	// types available in retrieves, their properties, subtyping and more. The parser always includes
	// the system data model, but the model info files of other data models can be provided. In this
	// example we use the FHIR data model so we can retrieve FHIR Observations and access their
	// effective and id properties. We currently only support FHIR version 4.0.1 data model.
	//
	// FHIR Helpers is a CQL Library with helper functions to covert between
	fhirDataModel, fhirHelpers, err := cql.FHIRDataModelAndHelpersLib("4.0.1")
	if err != nil {
		log.Fatal(err)
	}

	// In this example we are returning a list of the ID's of Observations that were effective during
	// the measurement period.
	libs := []string{
		dedent.Dedent(`
		library Example version '1.2.3'
		using FHIR version '4.0.1'
		include FHIRHelpers version '4.0.1'
		parameter MeasurementPeriod Interval<DateTime>
		context Patient

		define EffectiveObservations: [Observation] O where O.effective in MeasurementPeriod return O.id.value
		define FirstObservation: First(EffectiveObservations)
		`),
		fhirHelpers,
	}

	// TODO(b/335206660): Golang contexts are not yet properly supported by our engine.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Parameters override the values of the parameters defined in the CQL library. Parameters are a
	// map from the library/parameter name to a string CQL literal. Any valid CQL literal syntax will
	// be accepted, such as 400 or List<Choice<Integer, String>>{1, 'stringParam'}. In this example we
	// override the MeasurementPeriod parameter to an interval between 2017 and 2019.
	parameters := map[result.DefKey]string{
		result.DefKey{
			Library: result.LibKey{Name: "Example", Version: "1.2.3"},
			Name:    "MeasurementPeriod",
		}: "Interval[@2017, @2019]"}

	// Parse will validate the data models and parse the libraries and parameters. ELM (which stands
	// for Expression Logical Model) holds the parsed CQL, ready to be evaluated. Anything in the
	// ParseConfig is optional.
	elm, err := cql.Parse(ctx, libs, cql.ParseConfig{DataModels: [][]byte{fhirDataModel}, Parameters: parameters})
	if err != nil {
		log.Fatalf("Failed to parse: %v", err)
	}

	for _, id := range []string{"PatientID1", "PatientID2"} {
		// The retriever is used by the interpreter to fetch FHIR resources on each CQL
		// retrieve. In this case we are in the `context patient` and call `[Observation]` so the
		// retriever will fetch all Observations for the particular Patient.
		retriever, err := NewRetriever(id)
		if err != nil {
			log.Fatalf("Failed to build retriever: %v", err)
		}

		// Eval executes the ELM (aka parsed CQL) against this particular instantiation of the
		// retriever. Anything in EvalConfig is optional.
		results, err := elm.Eval(ctx, retriever, cql.EvalConfig{})
		if err != nil {
			log.Fatalf("Failed to evaluate: %v", err)
		}

		// The results are stored in maps, and can be accessed via [result.LibKey][Definition]. The CQL
		// string, list, integers... are stored in result.Value and can be converted to a golang value
		// via GolangValue() or by passing the result.Value to a helper like result.ToString. Another
		// option is to use MarshalJSON() to convert the result.Value to json, see the results package
		// for more details.
		cqlObservationID := results[result.LibKey{Name: "Example", Version: "1.2.3"}]["FirstObservation"]

		if result.IsNull(cqlObservationID) {
			fmt.Printf("ID %v: null\n", id)
		} else {
			golangStr, err := result.ToString(cqlObservationID)
			if err != nil {
				log.Fatalf("Failed to get golang string: %v", err)
			}
			fmt.Printf("ID %v: %v\n", id, golangStr)
		}
	}

	// Output:
	// ID PatientID1: null
	// ID PatientID2: Observation2
}

func NewRetriever(patientID string) (*local.Retriever, error) {
	Patient1Bundle := `{
		"resourceType": "Bundle",
		"type": "transaction",
		"entry": [
			{
				"fullUrl": "fullUrl",
				"resource": {
					"resourceType": "Patient",
					"id": "PatientID1",
					"name": [{"given":["John", "Smith"], "family":"Doe"}]}
			},
			{
				"fullUrl": "fullUrl",
				"resource": {
					"resourceType": "Observation",
					"id": "Observation1",
					"effectiveDateTime": "2012-04-02T10:30:10+01:00"
				}
			}
		 ]
	}`
	Patient2Bundle := `{
		"resourceType": "Bundle",
		"type": "transaction",
		"entry": [
			{
				"fullUrl": "fullUrl",
				"resource": {
					"resourceType": "Patient",
					"id": "PatientID2",
					"name": [{"given":["Jane", "Smith"], "family":"Doe"}]}
			},
			{
				"fullUrl": "fullUrl",
				"resource": {
					"resourceType": "Observation",
					"id": "Observation2",
					"effectiveDateTime": "2018-04-02T10:30:10+01:00"
				}
			}
		 ]
	}`
	switch patientID {
	case "PatientID1":
		return local.NewRetrieverFromR4Bundle([]byte(Patient1Bundle))
	case "PatientID2":
		return local.NewRetrieverFromR4Bundle([]byte(Patient2Bundle))
	default:
		return nil, fmt.Errorf("invalid patient id %v", patientID)
	}
}
