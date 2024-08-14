// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
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

	"github.com/google/cql/interpreter"
	"github.com/google/cql/model"
	"github.com/google/cql/parser"
	"github.com/google/cql/result"
	"github.com/google/cql/types"
	"github.com/google/go-cmp/cmp"
	"github.com/lithammer/dedent"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestCalculateAgeAt(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		// AgeAt()
		// The test patient's birthday is 1950-01-01. Note AgeAt is translated into CalculateAgeAt
		// by the parser, so although the syntax is different much of the interpreter logic is shared.
		{
			name: "AgeAt Date",
			cql:  "AgeInYearsAt(@2023-06-14)",
			wantModel: &model.CalculateAgeAt{
				Precision: model.YEAR,
				BinaryExpression: &model.BinaryExpression{
					Expression: model.ResultType(types.Integer),
					Operands: []model.IExpression{
						&model.Property{
							Expression: model.ResultType(types.Date),
							Source: &model.Property{
								Expression: model.ResultType(&types.Named{TypeName: "FHIR.date"}),
								Source: &model.ExpressionRef{
									Expression: model.ResultType(&types.Named{TypeName: "FHIR.Patient"}),
									Name:       "Patient",
								},
								Path: "birthDate",
							},
							Path: "value",
						},
						model.NewLiteral("@2023-06-14", types.Date),
					},
				},
			},
			wantResult: newOrFatal(t, 73),
		},
		{
			name:       "AgeAt DateTime",
			cql:        "AgeInYearsAt(@2023-06-15T10:01:01.000Z)",
			wantResult: newOrFatal(t, 73),
		},
		// CalculateAgeAt()
		{
			name: "Left Null",
			cql:  "CalculateAgeInYearsAt(null, @2023-06-14)",
			wantModel: &model.CalculateAgeAt{
				Precision: model.YEAR,
				BinaryExpression: &model.BinaryExpression{
					Expression: model.ResultType(types.Integer),
					Operands: []model.IExpression{
						&model.As{
							UnaryExpression: &model.UnaryExpression{
								Expression: model.ResultType(types.Date),
								Operand:    model.NewLiteral("null", types.Any),
							},
							AsTypeSpecifier: types.Date,
						},
						model.NewLiteral("@2023-06-14", types.Date),
					},
				},
			},
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Right Null",
			cql:        "CalculateAgeInYearsAt(@1981-06-15, null)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Years precision Dates",
			cql:        "CalculateAgeInYearsAt(@1981-06-15, @2023-06-14)",
			wantResult: newOrFatal(t, 41),
		},
		{
			name:       "Years precision DateTimes",
			cql:        "CalculateAgeInYearsAt(@1981-06-15, @2023-06-15T10:01:01.000Z)",
			wantResult: newOrFatal(t, 42),
		},
		{
			name:       "Years precision day short",
			cql:        "CalculateAgeInYearsAt(@1981-06-15, @2023-06-14)",
			wantResult: newOrFatal(t, 41),
		},
		{
			name:       "Months precision",
			cql:        "CalculateAgeInMonthsAt(@2022-06-15, @2023-06-14)",
			wantResult: newOrFatal(t, 11),
		},
		{
			name:       "Months precision day short",
			cql:        "CalculateAgeInMonthsAt(@2022-06-15, @2023-06-14)",
			wantResult: newOrFatal(t, 11),
		},
		{
			name:       "Weeks precision",
			cql:        "CalculateAgeInWeeksAt(@2023-06-01, @2023-06-14)",
			wantResult: newOrFatal(t, 1),
		},
		{
			name:       "Days precision",
			cql:        "CalculateAgeInDaysAt(@2023-06-01, @2023-06-15)",
			wantResult: newOrFatal(t, 14),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := newFHIRParser(t)
			parsedLibs, err := p.Libraries(context.Background(), wrapInLib(t, tc.cql), parser.Config{})
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantModel, getTESTRESULTModel(t, parsedLibs)); tc.wantModel != nil && diff != "" {
				t.Errorf("Parse diff (-want +got):\n%s", diff)
			}

			results, err := interpreter.Eval(context.Background(), parsedLibs, defaultInterpreterConfig(t, p))
			if err != nil {
				t.Fatalf("Eval returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantResult, getTESTRESULT(t, results), protocmp.Transform()); diff != "" {
				t.Errorf("Eval diff (-want +got)\n%v", diff)
			}
		})
	}
}

func TestInValueSetAndCodeSystem(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		// ValueSet tests
		{
			name: "Code In ValueSet",
			cql: dedent.Dedent(`
			valueset VS: 'https://example.com/vs/glucose' version '1.0.0'
			codesystem CS: 'https://example.com/cs/diagnosis' version '1.0.0'
			code C: 'gluc' from CS
			define TESTRESULT: C in VS`),
			wantResult: newOrFatal(t, true),
		},
		{
			name: "Code In unversioned ValueSet",
			cql: dedent.Dedent(`
			valueset VS: 'https://example.com/vs/glucose'
			codesystem CS: 'https://example.com/cs/diagnosis' version '1.0.0'
			code C: 'gluc' from CS
			define TESTRESULT: C in VS`),
			wantResult: newOrFatal(t, true),
		},
		{
			name: "Code Not In ValueSet",
			cql: dedent.Dedent(`
			valueset VS: 'https://example.com/vs/glucose' version '1.0.0'
			codesystem CS: 'https://example.com/cs/diagnosis' version '1.0.0'
			code C: 'NotInValueSet' from CS
			define TESTRESULT: C in VS`),
			wantResult: newOrFatal(t, false),
		},
		{
			name: "One Code from List<Code> In unversioned ValueSet",
			cql: dedent.Dedent(`
			valueset VS: 'https://example.com/vs/glucose'
			codesystem CS: 'https://example.com/cs/diagnosis' version '1.0.0'
			code ExistsCode: 'gluc' from CS
			code NonexistantCode: 'NotInValueSet' from CS
			define TESTRESULT: { NonexistantCode, ExistsCode } in VS`),
			wantResult: newOrFatal(t, true),
		},
		{
			name: "List<Code> Not In unversioned ValueSet",
			cql: dedent.Dedent(`
			valueset VS: 'https://example.com/vs/glucose'
			codesystem CS: 'https://example.com/cs/diagnosis' version '1.0.0'
			code NonexistantCode: 'NotInValueSet' from CS
			code NonexistantCode2: 'NotInValueSet2' from CS
			define TESTRESULT: { NonexistantCode, NonexistantCode2 } in VS`),
			wantResult: newOrFatal(t, false),
		},
		{
			name: "One Code from Concept In unversioned ValueSet",
			cql: dedent.Dedent(`
			valueset VS: 'https://example.com/vs/glucose'
			codesystem CS: 'https://example.com/cs/diagnosis' version '1.0.0'
			code ExistsCode: 'gluc' from CS
			code NonexistantCode: 'NotInValueSet' from CS
			concept Con: { NonexistantCode, ExistsCode }
			define TESTRESULT: Con in VS`),
			wantResult: newOrFatal(t, true),
		},
		{
			name: "Concept Not In unversioned ValueSet",
			cql: dedent.Dedent(`
			valueset VS: 'https://example.com/vs/glucose'
			codesystem CS: 'https://example.com/cs/diagnosis' version '1.0.0'
			code NonexistantCode: 'NotInValueSet' from CS
			code NonexistantCode2: 'NotInValueSet2' from CS
			concept Con: { NonexistantCode, NonexistantCode2 }
			define TESTRESULT: Con in VS`),
			wantResult: newOrFatal(t, false),
		},
		{
			name: "One Code from List<Concept> In unversioned ValueSet",
			cql: dedent.Dedent(`
			valueset VS: 'https://example.com/vs/glucose'
			codesystem CS: 'https://example.com/cs/diagnosis' version '1.0.0'
			code ExistsCode: 'gluc' from CS
			code NonexistantCode: 'NotInValueSet' from CS
			concept ConWithValidCode: { ExistsCode }
			concept ConNoValidCode: { NonexistantCode }
			define TESTRESULT: { ConNoValidCode, ConWithValidCode } in VS`),
			wantResult: newOrFatal(t, true),
		},
		{
			name: "List<Concept> Not In unversioned ValueSet",
			cql: dedent.Dedent(`
			valueset VS: 'https://example.com/vs/glucose'
			codesystem CS: 'https://example.com/cs/diagnosis' version '1.0.0'
			code NonexistantCode: 'NotInValueSet' from CS
			code NonexistantCode2: 'NotInValueSet2' from CS
			concept ConNoValidCode: { NonexistantCode }
			concept ConNoValidCode2: { NonexistantCode2 }
			define TESTRESULT: { ConNoValidCode, ConNoValidCode2 } in VS`),
			wantResult: newOrFatal(t, false),
		},
		// CodeSystem tests
		{
			name: "Code In Code System",
			cql: dedent.Dedent(`
			codesystem CS: 'https://example.com/cs/diagnosis' version '1.0.0'
			code C: 'snfl' from CS
			define TESTRESULT: C in CS`),
			wantResult: newOrFatal(t, true),
		},
		{
			name: "Code Not In Code System",
			cql: dedent.Dedent(`
			codesystem CS: 'https://example.com/cs/diagnosis' version '1.0.0'
			code C: 'NotInCodeSystem' from CS
			define TESTRESULT: C in CS`),
			wantResult: newOrFatal(t, false),
		},
		{
			name: "One Code from List<Code> In Code System",
			cql: dedent.Dedent(`
			codesystem CS: 'https://example.com/cs/diagnosis' version '1.0.0'
			code ExistsCode: 'snfl' from CS
			code NonexistantCode: 'NotInCodeSystem' from CS
			define TESTRESULT: { NonexistantCode, ExistsCode } in CS`),
			wantResult: newOrFatal(t, true),
		},
		{
			name: "List<Code> Not In Code System",
			cql: dedent.Dedent(`
			codesystem CS: 'https://example.com/cs/diagnosis' version '1.0.0'
			code NonexistantCode: 'NotInCodeSystem' from CS
			code NonexistantCode2: 'NotInCodeSystem2' from CS
			define TESTRESULT: { NonexistantCode, NonexistantCode2 } in CS`),
			wantResult: newOrFatal(t, false),
		},
		{
			name: "One Code from Concept In Code System",
			cql: dedent.Dedent(`
			codesystem CS: 'https://example.com/cs/diagnosis' version '1.0.0'
			code ExistsCode: 'snfl' from CS
			code NonexistantCode: 'NotInCodeSystem' from CS
			concept Con: { NonexistantCode, ExistsCode }
			define TESTRESULT: Con in CS`),
			wantResult: newOrFatal(t, true),
		},
		{
			name: "Concept Not In Code System",
			cql: dedent.Dedent(`
			codesystem CS: 'https://example.com/cs/diagnosis' version '1.0.0'
			code NonexistantCode: 'NotInCodeSystem' from CS
			code NonexistantCode2: 'NotInCodeSystem2' from CS
			concept Con: { NonexistantCode, NonexistantCode2 }
			define TESTRESULT: Con in CS`),
			wantResult: newOrFatal(t, false),
		},
		{
			name: "One Code from List<Concept> In Code System",
			cql: dedent.Dedent(`
			codesystem CS: 'https://example.com/cs/diagnosis' version '1.0.0'
			code ExistsCode: 'snfl' from CS
			code NonexistantCode: 'NotInCodeSystem' from CS
			concept ConWithValidCode: { ExistsCode }
			concept ConNoValidCode: { NonexistantCode }
			define TESTRESULT: { ConNoValidCode, ConWithValidCode } in CS`),
			wantResult: newOrFatal(t, true),
		},
		{
			name: "Concept Not In Code System",
			cql: dedent.Dedent(`
			codesystem CS: 'https://example.com/cs/diagnosis' version '1.0.0'
			code NonexistantCode: 'NotInCodeSystem' from CS
			code NonexistantCode2: 'NotInCodeSystem2' from CS
			concept ConNoValidCode: { NonexistantCode }
			concept ConNoValidCode2: { NonexistantCode2 }
			define TESTRESULT: { ConNoValidCode, ConNoValidCode2 } in CS`),
			wantResult: newOrFatal(t, false),
		},
		{
			name: "One Code from Concept that contains nulls In Code System",
			cql: dedent.Dedent(`
			codesystem CS: 'https://example.com/cs/diagnosis' version '1.0.0'
			code ExistsCode: 'snfl' from CS
			define Con: Concept{ codes: { null as Code, ExistsCode, null as Code } }
			define TESTRESULT: Con in CS`),
			wantResult: newOrFatal(t, true),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testCQL := dedent.Dedent(fmt.Sprintf(`
				library TESTLIB version '1.0.0'
				using FHIR version '4.0.1'
				%v`, tc.cql))
			p := newFHIRParser(t)
			parsedLibs, err := p.Libraries(context.Background(), addFHIRHelpersLib(t, testCQL), parser.Config{})
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantModel, getTESTRESULTModel(t, parsedLibs)); tc.wantModel != nil && diff != "" {
				t.Errorf("Parse diff (-want +got):\n%s", diff)
			}

			results, err := interpreter.Eval(context.Background(), parsedLibs, defaultInterpreterConfig(t, p))
			if err != nil {
				t.Fatalf("Eval returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantResult, getTESTRESULT(t, results), protocmp.Transform()); diff != "" {
				t.Errorf("Eval diff (-want +got)\n%v", diff)
			}

		})
	}
}
