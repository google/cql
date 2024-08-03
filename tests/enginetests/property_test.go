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

package enginetests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/cql/interpreter"
	"github.com/google/cql/model"
	"github.com/google/cql/parser"
	"github.com/google/cql/result"
	"github.com/google/cql/retriever/local"
	"github.com/google/cql/retriever"
	"github.com/google/cql/types"
	c4pb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/codes_go_proto"
	d4pb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/datatypes_go_proto"
	r4pb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/resources/bundle_and_contained_resource_go_proto"
	r4encounterpb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/resources/encounter_go_proto"
	r4observationpb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/resources/observation_go_proto"
	r4patientpb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/resources/patient_go_proto"
	"github.com/google/go-cmp/cmp"
	"github.com/lithammer/dedent"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestProperty(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		resources  []*r4pb.ContainedResource
		wantModel  model.IExpression
		wantResult result.Value
	}{
		// Literals
		{
			name: "property on null",
			cql:  "define TESTRESULT: null.test",
			wantModel: &model.Property{
				Source:     model.NewLiteral("null", types.Any),
				Path:       "test",
				Expression: model.ResultType(types.Any),
			},
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "property on empty list",
			cql:        "define TESTRESULT: {}.test",
			wantResult: newOrFatal(t, result.List{Value: []result.Value{}, StaticType: &types.List{ElementType: types.Any}}),
		},
		{
			name: "Interval[4, 5].low return 4",
			cql:  "define TESTRESULT: Interval[4, 5].low",
			wantModel: &model.Property{
				Source: &model.Interval{
					Low:           model.NewLiteral("4", types.Integer),
					High:          model.NewLiteral("5", types.Integer),
					LowInclusive:  true,
					HighInclusive: true,
					Expression:    model.ResultType(&types.Interval{PointType: types.Integer}),
				},
				Path:       "low",
				Expression: model.ResultType(types.Integer),
			},
			wantResult: newOrFatal(t, 4),
		},
		{
			name:       "Interval[4, 5].high returns 5",
			cql:        "define TESTRESULT: Interval[4, 5].high",
			wantResult: newOrFatal(t, 5),
		},
		{
			name:       "Interval[4, 5].lowClosed returns true",
			cql:        "define TESTRESULT: Interval[4, 5].lowClosed",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Interval[4, 5].highClosed returns true",
			cql:        "define TESTRESULT: Interval[4, 5].highClosed",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Interval(4, 5).lowClosed returns false",
			cql:        "define TESTRESULT: Interval(4, 5).lowClosed",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Interval(4, 5).highClosed returns false",
			cql:        "define TESTRESULT: Interval(4, 5).highClosed",
			wantResult: newOrFatal(t, false),
		},
		{
			name: "Quantity.unit",
			cql: dedent.Dedent(`
			define Q: 1 month
			define TESTRESULT: Q.unit`),
			wantResult: newOrFatal(t, "month"),
		},
		{
			name: "Code.system",
			cql: dedent.Dedent(`
			codesystem cs: 'https://example.com/cs/diagnosis' version '1.0'
			define C: Code '132' from cs display 'Severed Leg'
			define TESTRESULT: C.system`),
			wantResult: newOrFatal(t, "https://example.com/cs/diagnosis"),
		},
		{
			name: "ValueSet.version",
			cql: dedent.Dedent(`
			valueset vs: 'https://example.com/cs/diagnosis' version '1.0'
			define TESTRESULT: vs.version`),
			wantResult: newOrFatal(t, "1.0"),
		},
		{
			name: "CodeSystem.version",
			cql: dedent.Dedent(`
			codesystem cs: 'https://example.com/cs/diagnosis' version '1.0'
			define TESTRESULT: cs.version`),
			wantResult: newOrFatal(t, "1.0"),
		},
		// TODO(b/301606416): Add tests for concept once concept refs are supported.
		// Tuples and Instance
		{
			name:       "System Instance",
			cql:        "define TESTRESULT: Code{code: 'foo', system: 'bar', display: 'the foo', version: '1.0'}.code",
			wantResult: newOrFatal(t, "foo"),
		},
		{
			name: "FHIR Instance",
			cql: dedent.Dedent(`
			context Patient
			define TESTRESULT: Patient { gender: Patient.gender }.gender`),
			wantResult: newOrFatal(t, result.Named{
				Value:       &r4patientpb.Patient_GenderCode{Value: c4pb.AdministrativeGenderCode_MALE},
				RuntimeType: &types.Named{TypeName: "FHIR.AdministrativeGender"},
			}),
		},
		{
			name:       "Tuple",
			cql:        "define TESTRESULT: Tuple { apple: 'red', banana: 4 }.apple",
			wantResult: newOrFatal(t, "red"),
		},
		{
			name: "Tuple Choice",
			cql: dedent.Dedent(`
			define C: 4 as Choice<Integer, String>
			define TESTRESULT: Tuple { apple : C }.apple`),
			wantResult: newOrFatal(t, 4),
		},
		// FHIR Patient
		{
			name: "protomessage boolean returns boolean proto",
			cql: dedent.Dedent(`
					context Patient
					define TESTRESULT: Patient.active`),
			wantModel: &model.Property{
				Source: &model.ExpressionRef{
					Name:       "Patient",
					Expression: model.ResultType(&types.Named{TypeName: "FHIR.Patient"}),
				},
				Path:       "active",
				Expression: model.ResultType(&types.Named{TypeName: "FHIR.boolean"}),
			},
			wantResult: newOrFatal(t, result.Named{Value: &d4pb.Boolean{Value: true}, RuntimeType: &types.Named{TypeName: "FHIR.boolean"}}),
		},
		{
			name: "property.value on boolean proto returns System.Boolean",
			cql: dedent.Dedent(`
					context Patient
					define TESTRESULT: Patient.active.value`),
			wantResult: newOrFatal(t, true),
		},
		{
			name: "can call nested properties",
			cql: dedent.Dedent(`
					context Patient
					define TESTRESULT: Patient.name.family`),
			wantResult: newOrFatal(t, result.List{
				Value: []result.Value{
					newOrFatal(t, result.Named{Value: &d4pb.String{Value: "FamilyName"}, RuntimeType: &types.Named{TypeName: "FHIR.string"}}),
				},
				StaticType: &types.List{ElementType: &types.Named{TypeName: "FHIR.string"}}}),
		},
		{
			name: "property for enum returns a protomessage",
			cql: dedent.Dedent(`
					context Patient
					define TESTRESULT: Patient.gender`),
			wantResult: newOrFatal(t, result.Named{
				Value:       &r4patientpb.Patient_GenderCode{Value: c4pb.AdministrativeGenderCode_MALE},
				RuntimeType: &types.Named{TypeName: "FHIR.AdministrativeGender"},
			}),
		},
		{
			name: "property on repeated field returns list",
			cql: dedent.Dedent(`
					context Patient
					define TESTRESULT: Patient.name`),
			wantResult: newOrFatal(t, result.List{
				Value: []result.Value{
					newOrFatal(
						t,
						result.Named{
							Value: &d4pb.HumanName{
								Given: []*d4pb.String{
									&d4pb.String{Value: "GivenName"},
								},
								Family: &d4pb.String{Value: "FamilyName"}}, RuntimeType: &types.Named{TypeName: "FHIR.HumanName"}},
					),
				},
				StaticType: &types.List{ElementType: &types.Named{TypeName: "FHIR.HumanName"}},
			}),
		},
		{
			name: "property for unset non-repeated field is null",
			cql: dedent.Dedent(`
					context Patient
					define TESTRESULT: Patient.birthDate`),
			resources:  []*r4pb.ContainedResource{containedFromPatient(&r4patientpb.Patient{})},
			wantResult: newOrFatal(t, nil),
		},
		{
			name: "primitive property.value is null if parent proto is unset",
			cql: dedent.Dedent(`
					context Patient
					define TESTRESULT: Patient.active.value`),
			resources:  []*r4pb.ContainedResource{containedFromPatient(&r4patientpb.Patient{})},
			wantResult: newOrFatal(t, nil),
		},
		{
			name: "property for unset repeated field returns empty list",
			cql: dedent.Dedent(`
					context Patient
					define TESTRESULT: Patient.name`),
			resources:  []*r4pb.ContainedResource{containedFromPatient(&r4patientpb.Patient{})},
			wantResult: newOrFatal(t, result.List{Value: []result.Value{}, StaticType: &types.List{ElementType: &types.Named{TypeName: "FHIR.HumanName"}}}),
		},
		{
			name: "property retrieve on list of protomessages is flattened",
			cql:  "define TESTRESULT: ([Patient]).name.family",
			resources: []*r4pb.ContainedResource{
				containedFromPatient(&r4patientpb.Patient{Name: []*d4pb.HumanName{
					&d4pb.HumanName{Family: &d4pb.String{Value: "John"}},
					&d4pb.HumanName{Family: &d4pb.String{Value: "Jim"}},
				}}),
				containedFromPatient(&r4patientpb.Patient{Name: []*d4pb.HumanName{
					&d4pb.HumanName{Family: &d4pb.String{Value: "Dave"}},
					&d4pb.HumanName{Family: &d4pb.String{Value: "Dan"}},
				}}),
			},
			wantResult: newOrFatal(
				t,
				result.List{
					Value: []result.Value{
						newOrFatal(t, result.Named{Value: &d4pb.String{Value: "John"}, RuntimeType: &types.Named{TypeName: "FHIR.string"}}),
						newOrFatal(t, result.Named{Value: &d4pb.String{Value: "Jim"}, RuntimeType: &types.Named{TypeName: "FHIR.string"}}),
						newOrFatal(t, result.Named{Value: &d4pb.String{Value: "Dave"}, RuntimeType: &types.Named{TypeName: "FHIR.string"}}),
						newOrFatal(t, result.Named{Value: &d4pb.String{Value: "Dan"}, RuntimeType: &types.Named{TypeName: "FHIR.string"}}),
					},
					StaticType: &types.List{
						ElementType: &types.Named{TypeName: "FHIR.string"},
					},
				},
			),
		},
		{
			name: "property retrieve on list of protomessages alternate syntax",
			cql: dedent.Dedent(`
					define PatientRetrieve: [Patient]
					define TESTRESULT: PatientRetrieve.name.family`),
			resources: []*r4pb.ContainedResource{
				containedFromPatient(&r4patientpb.Patient{Name: []*d4pb.HumanName{
					&d4pb.HumanName{Family: &d4pb.String{Value: "John"}},
					&d4pb.HumanName{Family: &d4pb.String{Value: "Jim"}},
				}}),
				containedFromPatient(&r4patientpb.Patient{Name: []*d4pb.HumanName{
					&d4pb.HumanName{Family: &d4pb.String{Value: "Dave"}},
					&d4pb.HumanName{Family: &d4pb.String{Value: "Dan"}},
				}}),
			},
			wantResult: newOrFatal(
				t,
				result.List{
					Value: []result.Value{
						newOrFatal(t, result.Named{Value: &d4pb.String{Value: "John"}, RuntimeType: &types.Named{TypeName: "FHIR.string"}}),
						newOrFatal(t, result.Named{Value: &d4pb.String{Value: "Jim"}, RuntimeType: &types.Named{TypeName: "FHIR.string"}}),
						newOrFatal(t, result.Named{Value: &d4pb.String{Value: "Dave"}, RuntimeType: &types.Named{TypeName: "FHIR.string"}}),
						newOrFatal(t, result.Named{Value: &d4pb.String{Value: "Dan"}, RuntimeType: &types.Named{TypeName: "FHIR.string"}}),
					},
					StaticType: &types.List{
						ElementType: &types.Named{TypeName: "FHIR.string"},
					},
				},
			),
		},
		// Properties on Observations
		{
			name: "unset oneof returns nil",
			cql: dedent.Dedent(`
					define FirstObservation: First([Observation])
					define TESTRESULT: FirstObservation.value`),
			resources:  []*r4pb.ContainedResource{containedFromObservation(&r4observationpb.Observation{})},
			wantResult: newOrFatal(t, nil),
		},
		{
			name: "integer inside oneof returns integer proto",
			cql: dedent.Dedent(`
					define FirstObservation: First([Observation])
					define TESTRESULT: FirstObservation.value`),
			resources: []*r4pb.ContainedResource{
				containedFromObservation(&r4observationpb.Observation{Value: &r4observationpb.Observation_ValueX{Choice: &r4observationpb.Observation_ValueX_Integer{Integer: &d4pb.Integer{Value: 4}}}}),
			},
			wantResult: newOrFatal(t, result.Named{Value: &d4pb.Integer{Value: 4}, RuntimeType: &types.Named{TypeName: "FHIR.integer"}}),
		},
		{
			name: "string inside oneof returns string proto",
			cql: dedent.Dedent(`
					define FirstObservation: First([Observation])
					define TESTRESULT: FirstObservation.value`),
			resources: []*r4pb.ContainedResource{
				containedFromObservation(&r4observationpb.Observation{
					Value: &r4observationpb.Observation_ValueX{Choice: &r4observationpb.Observation_ValueX_StringValue{StringValue: &d4pb.String{Value: "obsValue"}}},
				}),
			},
			wantResult: newOrFatal(t, result.Named{Value: &d4pb.String{Value: "obsValue"}, RuntimeType: &types.Named{TypeName: "FHIR.string"}}),
		},
		{
			name: "FHIR.decimal.value returns a System.Decimal",
			cql: dedent.Dedent(`
					define FirstObservation: First([Observation])
					define TESTRESULT: (FirstObservation.value as FHIR.Quantity).value.value`),
			resources: []*r4pb.ContainedResource{
				containedFromObservation(&r4observationpb.Observation{Value: &r4observationpb.Observation_ValueX{Choice: &r4observationpb.Observation_ValueX_Quantity{Quantity: &d4pb.Quantity{Value: &d4pb.Decimal{Value: "100.1"}}}}}),
			},
			wantResult: newOrFatal(t, 100.1),
		},
		{
			name: "dateTime inside oneof returns dateTime proto",
			cql: dedent.Dedent(`
					define FirstObservation: First([Observation])
					define TESTRESULT: FirstObservation.effective`),
			resources: []*r4pb.ContainedResource{
				containedFromObservation(&r4observationpb.Observation{
					Effective: &r4observationpb.Observation_EffectiveX{
						Choice: &r4observationpb.Observation_EffectiveX_DateTime{DateTime: &d4pb.DateTime{ValueUs: 1711929600000000, Precision: d4pb.DateTime_SECOND, Timezone: "UTC"}},
					},
				}),
			},
			wantResult: newOrFatal(t, result.Named{Value: &d4pb.DateTime{ValueUs: 1711929600000000, Precision: d4pb.DateTime_SECOND, Timezone: "UTC"}, RuntimeType: &types.Named{TypeName: "FHIR.dateTime"}}),
		},
		{
			name: "proto message with capital result type inside oneof",
			cql: dedent.Dedent(`
					define FirstObservation: First([Observation])
					define TESTRESULT: FirstObservation.value`),
			resources: []*r4pb.ContainedResource{
				containedFromObservation(&r4observationpb.Observation{Value: &r4observationpb.Observation_ValueX{Choice: &r4observationpb.Observation_ValueX_SampledData{SampledData: &d4pb.SampledData{Id: &d4pb.String{Value: "myID"}}}}}),
			},
			wantResult: newOrFatal(t, result.Named{
				Value:       &d4pb.SampledData{Id: &d4pb.String{Value: "myID"}},
				RuntimeType: &types.Named{TypeName: "FHIR.SampledData"}, // Note that the result type is set correctly (and is not a Choice type).
			}),
		},
		{
			name: "proto message with lowercase result type inside oneof",
			cql: dedent.Dedent(`
					define FirstObservation: First([Observation])
					define TESTRESULT: FirstObservation.value`),
			resources: []*r4pb.ContainedResource{
				containedFromObservation(&r4observationpb.Observation{Value: &r4observationpb.Observation_ValueX{Choice: &r4observationpb.Observation_ValueX_Time{Time: &d4pb.Time{ValueUs: 1711929600000000}}}}),
			},
			wantResult: newOrFatal(t, result.Named{
				Value:       &d4pb.Time{ValueUs: 1711929600000000},
				RuntimeType: &types.Named{TypeName: "FHIR.time"}, // Note that the result type is set correctly (and is not a Choice type).
			}),
		},
		{
			name: "enums are wrapped",
			cql: dedent.Dedent(`
					define FirstObservation: First([Observation])
					define TESTRESULT: FirstObservation.status`),
			resources: []*r4pb.ContainedResource{
				containedFromObservation(&r4observationpb.Observation{Status: &r4observationpb.Observation_StatusCode{Value: c4pb.ObservationStatusCode_FINAL}}),
			},
			wantResult: newOrFatal(t, result.Named{Value: &r4observationpb.Observation_StatusCode{Value: c4pb.ObservationStatusCode_FINAL}, RuntimeType: &types.Named{TypeName: "FHIR.ObservationStatus"}}),
		},
		{
			name: "enum.value returns string",
			cql: dedent.Dedent(`
					define FirstObservation: First([Observation])
					define TESTRESULT: FirstObservation.status.value`),
			resources: []*r4pb.ContainedResource{
				containedFromObservation(&r4observationpb.Observation{Status: &r4observationpb.Observation_StatusCode{Value: c4pb.ObservationStatusCode_ENTERED_IN_ERROR}}),
			},
			wantResult: newOrFatal(t, "entered-in-error"),
		},
		{
			name: "FHIR.dateTime.value returns System.DateTime",
			cql: dedent.Dedent(`
					define FirstObservation: First([Observation])
					define TESTRESULT: FirstObservation.effective.value`),
			resources: []*r4pb.ContainedResource{
				containedFromObservation(&r4observationpb.Observation{
					Effective: &r4observationpb.Observation_EffectiveX{
						Choice: &r4observationpb.Observation_EffectiveX_DateTime{DateTime: &d4pb.DateTime{ValueUs: 1711929600000000, Precision: d4pb.DateTime_SECOND, Timezone: "UTC"}},
					},
				}),
			},
			wantResult: newOrFatal(t, result.DateTime{Date: time.Date(2024, time.April, 1, 0, 0, 0, 0, time.UTC), Precision: model.SECOND}),
		},
		{
			name: "FHIR.date.value returns System.Date",
			cql: dedent.Dedent(`
					context Patient
					define TESTRESULT: Patient.birthDate.value`),
			resources: []*r4pb.ContainedResource{containedFromPatient(&r4patientpb.Patient{
				Gender:    &r4patientpb.Patient_GenderCode{Value: c4pb.AdministrativeGenderCode_MALE},
				BirthDate: &d4pb.Date{ValueUs: 1711929600000000, Precision: d4pb.Date_DAY, Timezone: "UTC"},
			})},
			wantResult: newOrFatal(t, result.Date{Date: time.Date(2024, time.April, 1, 0, 0, 0, 0, time.FixedZone("", 4*60*60)), Precision: model.DAY}),
		},
		{
			name: "FHIR.dateTime.value returns System.DateTime with microsecond precision mapped to millisecond",
			cql: dedent.Dedent(`
					define FirstObservation: First([Observation])
					define TESTRESULT: FirstObservation.effective.value`),
			resources: []*r4pb.ContainedResource{
				containedFromObservation(&r4observationpb.Observation{
					Effective: &r4observationpb.Observation_EffectiveX{
						Choice: &r4observationpb.Observation_EffectiveX_DateTime{DateTime: &d4pb.DateTime{ValueUs: 1711929600000000, Precision: d4pb.DateTime_MICROSECOND, Timezone: "UTC"}},
					},
				}),
			},
			wantResult: newOrFatal(t, result.DateTime{Date: time.Date(2024, time.April, 1, 0, 0, 0, 0, time.UTC), Precision: model.MILLISECOND}),
		},
		{
			name: "FHIR.dateTime.value returns System.DateTime with correct unknown precision mapping",
			cql: dedent.Dedent(`
					define FirstObservation: First([Observation])
					define TESTRESULT: FirstObservation.effective.value`),
			resources: []*r4pb.ContainedResource{
				containedFromObservation(&r4observationpb.Observation{
					Effective: &r4observationpb.Observation_EffectiveX{
						Choice: &r4observationpb.Observation_EffectiveX_DateTime{DateTime: &d4pb.DateTime{ValueUs: 1711929600000000, Precision: d4pb.DateTime_PRECISION_UNSPECIFIED, Timezone: "UTC"}},
					},
				}),
			},
			wantResult: newOrFatal(t, result.DateTime{Date: time.Date(2024, time.April, 1, 0, 0, 0, 0, time.UTC), Precision: model.UNSETDATETIMEPRECISION}),
		},
		{
			name: "Encounter.class has a different json and proto field name",
			cql: dedent.Dedent(`
					define TESTRESULT: First([Encounter]).class`),
			resources: []*r4pb.ContainedResource{
				&r4pb.ContainedResource{
					OneofResource: &r4pb.ContainedResource_Encounter{
						Encounter: &r4encounterpb.Encounter{
							ClassValue: &d4pb.Coding{Display: &d4pb.String{Value: "Display"}},
						},
					},
				},
			},
			wantResult: newOrFatal(t, result.Named{Value: &d4pb.Coding{Display: &d4pb.String{Value: "Display"}}, RuntimeType: &types.Named{TypeName: "FHIR.Coding"}}),
		},
		{
			name: "Ensure camelCase json properties work correctly: Encounter.serviceType",
			cql: dedent.Dedent(`
					define TESTRESULT: First([Encounter]).serviceType`),
			resources: []*r4pb.ContainedResource{
				&r4pb.ContainedResource{
					OneofResource: &r4pb.ContainedResource_Encounter{
						Encounter: &r4encounterpb.Encounter{
							ServiceType: &d4pb.CodeableConcept{Text: &d4pb.String{Value: "ServiceType"}},
						},
					},
				},
			},
			wantResult: newOrFatal(t, result.Named{Value: &d4pb.CodeableConcept{Text: &d4pb.String{Value: "ServiceType"}}, RuntimeType: &types.Named{TypeName: "FHIR.CodeableConcept"}}),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testCQL := fmt.Sprintf(dedent.Dedent(`
			library TESTLIB version '1.0.0'
			using FHIR version '4.0.1'
			%v`), tc.cql)
			p := newFHIRParser(t)
			parsedLibs, err := p.Libraries(context.Background(), addFHIRHelpersLib(t, testCQL), parser.Config{})
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantModel, getTESTRESULTModel(t, parsedLibs)); tc.wantModel != nil && diff != "" {
				t.Errorf("Parse diff (-want +got):\n%s", diff)
			}

			config := defaultInterpreterConfig(t, p)
			if tc.resources != nil {
				config.Retriever = newRetrieverFromProtosOrFatal(t, tc.resources)
			}
			results, err := interpreter.Eval(context.Background(), parsedLibs, config)
			if err != nil {
				t.Fatalf("Eval returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantResult, getTESTRESULT(t, results), protocmp.Transform()); diff != "" {
				t.Errorf("Eval diff (-want +got)\n%v", diff)
			}

		})
	}
}

func newRetrieverFromProtosOrFatal(t *testing.T, crs []*r4pb.ContainedResource) retriever.Retriever {
	t.Helper()
	bundle := &r4pb.Bundle{}
	for _, cr := range crs {
		bundle.Entry = append(bundle.Entry, &r4pb.Bundle_Entry{Resource: cr})
	}
	ret, err := local.NewRetrieverFromR4BundleProto(bundle)
	if err != nil {
		t.Fatalf("local.NewRetrieverFromR4BundleProto() failed: %v", err)
	}
	return ret
}

func containedFromObservation(o *r4observationpb.Observation) *r4pb.ContainedResource {
	return &r4pb.ContainedResource{
		OneofResource: &r4pb.ContainedResource_Observation{
			Observation: o,
		},
	}
}

func containedFromPatient(p *r4patientpb.Patient) *r4pb.ContainedResource {
	return &r4pb.ContainedResource{
		OneofResource: &r4pb.ContainedResource_Patient{
			Patient: p,
		},
	}
}
