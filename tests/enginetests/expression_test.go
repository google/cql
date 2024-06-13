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
	d4pb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/datatypes_go_proto"
	"github.com/google/go-cmp/cmp"
	"github.com/lithammer/dedent"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestExpressions(t *testing.T) {
	// TestExpressions suite is for miscellaneous expressions that aren't system operators and don't
	// have enough test cases to warrant their own test suite.
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		// Message
		{
			name: "Message returns source value",
			cql:  "Message(1.2, true, 'Code 100', 'Message', 'Test Message')",
			wantModel: &model.Message{
				Expression: model.ResultType(types.Decimal),
				Source:     model.NewLiteral("1.2", types.Decimal),
				Condition:  model.NewLiteral("true", types.Boolean),
				Code:       model.NewLiteral("Code 100", types.String),
				Severity:   model.NewLiteral("Message", types.String),
				Message:    model.NewLiteral("Test Message", types.String),
			},
			wantResult: newOrFatal(t, 1.2),
		},
		{
			name:       "Message returns source false condition",
			cql:        "Message(1.2, false, 'Code 100', 'Severity', 'Test Message')",
			wantResult: newOrFatal(t, 1.2),
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

func TestRetrieves(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Retrieve multiple FHIR resources",
			cql:  "define TESTRESULT: [Observation]",
			wantModel: &model.Retrieve{
				DataType:     "{http://hl7.org/fhir}Observation",
				TemplateID:   "http://hl7.org/fhir/StructureDefinition/Observation",
				CodeProperty: "code",
				Expression:   model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}}),
			},
			wantResult: newOrFatal(t, result.List{
				Value: []result.Value{
					newOrFatal(t, result.Named{Value: RetrieveFHIRResource(t, "Observation", "1"), RuntimeType: &types.Named{TypeName: "FHIR.Observation"}}),
					newOrFatal(t, result.Named{Value: RetrieveFHIRResource(t, "Observation", "2"), RuntimeType: &types.Named{TypeName: "FHIR.Observation"}}),
					newOrFatal(t, result.Named{Value: RetrieveFHIRResource(t, "Observation", "3"), RuntimeType: &types.Named{TypeName: "FHIR.Observation"}}),
				},
				StaticType: &types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}},
			}),
		},
		{
			name: "Retrieve filtered by valueset",
			cql: dedent.Dedent(`
			valueset GlucoseVS: 'https://example.com/vs/glucose'
			define TESTRESULT: [Observation: GlucoseVS]`),
			wantResult: newOrFatal(t, result.List{
				Value: []result.Value{
					newOrFatal(t, result.Named{Value: RetrieveFHIRResource(t, "Observation", "2"), RuntimeType: &types.Named{TypeName: "FHIR.Observation"}}),
				},
				StaticType: &types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}},
			}),
		},
		{
			name:       "Retrieve returns empty list",
			cql:        "define TESTRESULT: [Binary]",
			wantResult: newOrFatal(t, result.List{Value: []result.Value{}, StaticType: &types.List{ElementType: &types.Named{TypeName: "FHIR.Binary"}}}),
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

func TestLocalReferences(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Expression Ref",
			cql: dedent.Dedent(`
			define Foo: 4
			define TESTRESULT: Foo`),
			wantModel: &model.ExpressionRef{
				Name:       "Foo",
				Expression: model.ResultType(types.Integer),
			},
			wantResult: newOrFatal(t, 4),
		},
		{
			name: "Property on Expression Ref",
			cql: dedent.Dedent(`
			define Foo: [Patient] P
			define TESTRESULT: Foo.active`),
			wantResult: newOrFatal(t, result.List{
				Value:      []result.Value{newOrFatal(t, result.Named{Value: &d4pb.Boolean{Value: true}, RuntimeType: &types.Named{TypeName: "FHIR.boolean"}})},
				StaticType: &types.List{ElementType: &types.Named{TypeName: "FHIR.boolean"}}}),
		},
		{
			name: "Parameter Ref",
			cql: dedent.Dedent(`
			parameter Foo default 4
			define TESTRESULT: Foo`),
			wantResult: newOrFatal(t, 4),
		},
		{
			name: "Property on Alias Ref",
			cql:  "define TESTRESULT: [Patient] P return P.active",
			wantResult: newOrFatal(t, result.List{
				Value:      []result.Value{newOrFatal(t, result.Named{Value: &d4pb.Boolean{Value: true}, RuntimeType: &types.Named{TypeName: "FHIR.boolean"}})},
				StaticType: &types.List{ElementType: &types.Named{TypeName: "FHIR.boolean"}}}),
		},
		{
			name: "Proto converts to date for or less operator",
			cql: dedent.Dedent(`
			include FHIRHelpers version '4.0.1' called FHIRHelpers
			context Patient
			define TESTRESULT: Patient.birthDate 1 year or less on or before end of Interval[@2019, @2022]`),
			wantResult: newOrFatal(t, false),
		},
		{
			name: "ValueSet Ref",
			cql: dedent.Dedent(`
			valueset VS: 'https://example.com/vs/glucose' version '1.0'
			define TESTRESULT: VS`),
			wantResult: newOrFatal(t, result.ValueSet{ID: "https://example.com/vs/glucose", Version: "1.0"}),
		},
		{
			name: "CodeSystem Ref",
			cql: dedent.Dedent(`
			codesystem CS: 'https://example.com/cs/diagnosis' version '1.0'
			define TESTRESULT: CS`),
			wantResult: newOrFatal(t, result.CodeSystem{ID: "https://example.com/cs/diagnosis", Version: "1.0"}),
		},
		{
			name: "Code Ref",
			cql: dedent.Dedent(`
			codesystem CS: 'https://example.com/cs/diagnosis'
			code C: '1234' from CS display 'Display'
			define TESTRESULT: C`),
			wantResult: newOrFatal(t, result.Code{Code: "1234", System: "https://example.com/cs/diagnosis", Display: "Display"}),
		},
		{
			name: "Concept Ref",
			cql: dedent.Dedent(`
			codesystem CS: 'https://example.com/cs/diagnosis'
			code C: '1234' from CS
			concept Foo: {C} display 'Concept Display'
			define TESTRESULT: Foo`),
			wantResult: newOrFatal(t, result.Concept{
				Codes:   []result.Code{result.Code{Code: "1234", System: "https://example.com/cs/diagnosis"}},
				Display: "Concept Display",
			}),
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

func TestGlobalReferences(t *testing.T) {
	tests := []struct {
		name       string
		cqlLibs    []string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Global Expression Ref",
			cqlLibs: []string{
				dedent.Dedent(`
					library CQL_Helpers_Library version '1'
					define Foo: 4
					`),
				dedent.Dedent(`
					library TESTLIB version '1.0.0'
					include CQL_Helpers_Library version '1' called helpers
					define TESTRESULT: helpers.Foo`),
			},
			wantModel: &model.ExpressionRef{
				LibraryName: "helpers",
				Name:        "Foo",
				Expression:  model.ResultType(types.Integer),
			},
			wantResult: newOrFatal(t, 4),
		},
		{
			name: "Global Parameter Ref",
			cqlLibs: []string{
				dedent.Dedent(`
					library CQL_Helpers_Library version '1'
					parameter Foo default 4
					`),
				dedent.Dedent(`
					library TESTLIB version '1.0.0'
					include CQL_Helpers_Library version '1' called helpers
					define TESTRESULT: helpers.Foo`),
			},
			wantResult: newOrFatal(t, 4),
		},
		{
			name: "Global ValueSet Ref",
			cqlLibs: []string{
				dedent.Dedent(`
					library CQL_Helpers_Library version '1'
					valueset Foo: 'https://example.com/cs/diagnosis' version '1.0'
					`),
				dedent.Dedent(`
					library TESTLIB version '1.0.0'
					include CQL_Helpers_Library version '1' called helpers
					define TESTRESULT: helpers.Foo`),
			},
			wantResult: newOrFatal(t, result.ValueSet{ID: "https://example.com/cs/diagnosis", Version: "1.0"}),
		},
		{
			name: "Global CodeSystem Ref",
			cqlLibs: []string{
				dedent.Dedent(`
					library CQL_Helpers_Library version '1'
					codesystem Foo: 'https://example.com/cs/diagnosis' version '1.0'
					`),
				dedent.Dedent(`
					library TESTLIB version '1.0.0'
					include CQL_Helpers_Library version '1' called helpers
					define TESTRESULT: helpers.Foo`),
			},
			wantResult: newOrFatal(t, result.CodeSystem{ID: "https://example.com/cs/diagnosis", Version: "1.0"}),
		},
		{
			name: "Global Code Ref",
			cqlLibs: []string{
				dedent.Dedent(`
					library CQL_Helpers_Library version '1'
					codesystem CS: 'https://example.com/cs/diagnosis'
					code Foo: '1234' from CS
					`),
				dedent.Dedent(`
					library TESTLIB version '1.0.0'
					include CQL_Helpers_Library version '1' called helpers
					define TESTRESULT: helpers.Foo`),
			},
			wantResult: newOrFatal(t, result.Code{Code: "1234", System: "https://example.com/cs/diagnosis"}),
		},
		{
			name: "Global Concept Ref",
			cqlLibs: []string{
				dedent.Dedent(`
					library CQL_Helpers_Library version '1'
					codesystem CS: 'https://example.com/cs/diagnosis'
					code C: '1234' from CS
					concept Foo: {C} display 'Concept Display'
					`),
				dedent.Dedent(`
					library TESTLIB version '1.0.0'
					include CQL_Helpers_Library version '1' called helpers
					define TESTRESULT: helpers.Foo`),
			},
			wantResult: newOrFatal(t, result.Concept{
				Codes:   []result.Code{result.Code{Code: "1234", System: "https://example.com/cs/diagnosis"}},
				Display: "Concept Display",
			}),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := newFHIRParser(t)
			parsedLibs, err := p.Libraries(context.Background(), tc.cqlLibs, parser.Config{})
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
