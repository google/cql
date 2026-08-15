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

package transforms

import (
	"context"
	"testing"
	"time"

	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	cbpb "github.com/google/cql/protos/cql_beam_go_proto"
	crpb "github.com/google/cql/protos/cql_result_go_proto"
	"github.com/google/fhir/go/fhirversion"
	"github.com/google/fhir/go/jsonformat"
	bpb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/resources/bundle_and_contained_resource_go_proto"
	"github.com/google/go-cmp/cmp"
	"github.com/lithammer/dedent"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestCQLEvalFn(t *testing.T) {
	tests := []struct {
		name       string
		evalFn     *CQLEvalFn
		input      *bpb.Bundle
		wantResult []*cbpb.BeamResult
		wantError  []*cbpb.BeamError
	}{
		{
			name: "Successful Eval",
			evalFn: &CQLEvalFn{
				CQL: []string{dedent.Dedent(
					`library EvalTest version '1.0'
					using FHIR version '4.0.1'
					valueset "HypertensionVS": 'urn:example:hypertension'
					valueset "DiabetesVS": 'urn:example:diabetes'
					define HasHypertension: exists([Condition: "HypertensionVS"])
					define HasDiabetes: exists([Condition: "DiabetesVS"])`)},
				ValueSets:           valueSets,
				EvaluationTimestamp: time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
			},
			input: parseOrFatal(t, `{
				"resourceType": "Bundle",
				"entry": [
					{
						"resource": {
							"resourceType": "Patient",
							"id": "1"
						}
					},
					{
						"resource": {
							"resourceType": "Condition",
							"id": "1",
							"code": {
							  "coding": [{ "system": "http://example.com", "code": "11111", "display": "Hypertension"}]
							},
							"onsetDateTime" : "2023-10-01"
						}
					},
					{
						"resource": {
							"resourceType": "Condition",
							"id": "2",
							"code": {
							  "coding": [{ "system": "http://example.com", "code": "22222", "display": "Diabetes"}]
							},
							"onsetDateTime" : "2023-11-01"
						}
					}
				 ]
			}`).GetBundle(),
			wantResult: []*cbpb.BeamResult{
				{
					Id:                  proto.String("1"),
					EvaluationTimestamp: timestamppb.New(time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC)),
					Result: &crpb.Libraries{
						Libraries: []*crpb.Library{
							{
								Name:    proto.String("EvalTest"),
								Version: proto.String("1.0"),
								ExprDefs: map[string]*crpb.Value{
									"HasDiabetes": {
										Value: &crpb.Value_BooleanValue{BooleanValue: true},
									},
									"HasHypertension": {
										Value: &crpb.Value_BooleanValue{BooleanValue: true},
									},
									"DiabetesVS": {
										Value: &crpb.Value_ValueSetValue{
											ValueSetValue: &crpb.ValueSet{Id: proto.String("urn:example:diabetes"), Version: proto.String("")},
										},
									},
									"HypertensionVS": {
										Value: &crpb.Value_ValueSetValue{
											ValueSetValue: &crpb.ValueSet{Id: proto.String("urn:example:hypertension"), Version: proto.String("")},
										},
									},
								},
							},
							{
								Name:    proto.String("BeamMetadata"),
								Version: proto.String("1.0.0"),
								ExprDefs: map[string]*crpb.Value{
									"ID": {
										Value: &crpb.Value_StringValue{StringValue: "1"},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "EvalError from Bad ValueSet",
			evalFn: &CQLEvalFn{
				CQL: []string{dedent.Dedent(
					`library "EvalTest" version '1.0'
					using FHIR version '4.0.1'
					valueset "DiabetesVS": 'urn:example:nosuchvalueset'
					define HasDiabetes: exists([Condition: "DiabetesVS"])`)},
				ValueSets:           valueSets,
				EvaluationTimestamp: time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
			},
			input: parseOrFatal(t, `{
				"resourceType": "Bundle",
				"id": "bundle1",
				"entry": [
					{
						"resource": {
							"resourceType": "Patient",
							"id": "1"
						}
					},
					{
						"resource": {
							"resourceType": "Condition",
							"id": "1",
							"code": {
							  "coding": [{ "system": "http://example.com", "code": "11111", "display": "Hypertension"}]
							},
							"onsetDateTime" : "2023-10-01"
						}
					}
				 ]
			}`).GetBundle(),
			wantError: []*cbpb.BeamError{
				{
					ErrorMessage: proto.String("failed during CQL evaluation: EvalTest 1.0, failed to evaluate expression HasDiabetes: could not find ValueSet{urn:example:nosuchvalueset, } resource not loaded"),
					SourceUri:    proto.String("bundle:bundle1"),
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var gotOutput []*cbpb.BeamResult
			var gotError []*cbpb.BeamError
			emitOutput := func(e *cbpb.BeamResult) { gotOutput = append(gotOutput, e) }
			emitError := func(e *cbpb.BeamError) { gotError = append(gotError, e) }

			if err := test.evalFn.Setup(); err != nil {
				t.Fatalf("Setup() failed: %v", err)
			}
			test.evalFn.ProcessElement(context.Background(), test.input, emitOutput, emitError)

			if diff := cmp.Diff(test.wantResult, gotOutput, protocmp.Transform(), protocmp.SortRepeatedFields(&crpb.Libraries{}, "libraries")); diff != "" {
				t.Errorf("ProcessElement() returned diff (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(test.wantError, gotError, protocmp.Transform()); diff != "" {
				t.Errorf("ProcessElement() returned diff (-want +got):\n%s", diff)
			}
		})
	}
}

var valueSets = []string{
	`{
		"resourceType": "ValueSet",
		"url": "urn:example:hypertension",
		"version": "1.0.0",
		"expansion": {
			"contains": [
				{ "system": "http://example.com", "code": "11111" }
			]
		}
	}`,
	`{
		"resourceType": "ValueSet",
		"url": "urn:example:diabetes",
		"version": "1.0.0",
		"expansion": {
			"contains": [
				{ "system": "http://example.com", "code": "22222" }
			]
		}
	}`,
}

func parseOrFatal(t *testing.T, json string) *bpb.ContainedResource {
	t.Helper()
	unmarshaller, err := jsonformat.NewUnmarshallerWithoutValidation("UTC", fhirversion.R4)
	if err != nil {
		t.Fatalf("jsonformat.NewUnmarshallerWithoutValidation() failed: %v", err)
	}
	cr, err := unmarshaller.UnmarshalR4([]byte(json))
	if err != nil {
		t.Fatalf("UnmarshalR4(%q) failed: %v", json, err)
	}

	return cr
}
