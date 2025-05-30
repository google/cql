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
	"testing"
	"time"

	"github.com/google/cql/interpreter"
	"github.com/google/cql/model"
	"github.com/google/cql/parser"
	"github.com/google/cql/result"
	"github.com/google/cql/types"
	c4pb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/codes_go_proto"
	r4patientpb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/resources/patient_go_proto"
	"github.com/google/go-cmp/cmp"
	"github.com/lithammer/dedent"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestLiteral(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Integer",
			cql:  "1",
			wantModel: &model.Literal{
				Value:      "1",
				Expression: model.ResultType(types.Integer),
			},
			wantResult: newOrFatal(t, 1),
		},
		{
			name:       "Long",
			cql:        "1L",
			wantResult: newOrFatal(t, int64(1)),
		},
		{
			name:       "Decimal",
			cql:        "1.0",
			wantResult: newOrFatal(t, 1.0),
		},
		{
			name:       "Quantity with temporal unit",
			cql:        "1 'month'",
			wantResult: newOrFatal(t, result.Quantity{Value: 1, Unit: model.MONTHUNIT}),
		},
		{
			name:       "Ratio",
			cql:        "1 'cm':2 'cm'",
			wantResult: newOrFatal(t, result.Ratio{Numerator: result.Quantity{Value: 1, Unit: "cm"}, Denominator: result.Quantity{Value: 2, Unit: "cm"}}),
		},
		{
			name:       "Boolean",
			cql:        "true",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "String",
			cql:        "'apple'",
			wantResult: newOrFatal(t, "apple"),
		},
		{
			name:       "Null",
			cql:        "null",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Date with default timezone",
			cql:        "@2024-02-20",
			wantResult: newOrFatal(t, result.Date{Date: time.Date(2024, 2, 20, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.DAY}),
		},
		{
			name:       "DateTime with default timezone",
			cql:        "@2024-03-31T",
			wantResult: newOrFatal(t, result.DateTime{Date: time.Date(2024, time.March, 31, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.DAY}),
		},
		{
			name:       "DateTime Zulu override",
			cql:        "@2024-03-31T01:20:30.101Z",
			wantResult: newOrFatal(t, result.DateTime{Date: time.Date(2024, time.March, 31, 1, 20, 30, 101e6, time.UTC), Precision: model.MILLISECOND}),
		},
		{
			name:       "DateTime TimeZone override",
			cql:        "@2024-03-31T01:20:30.101-07:00",
			wantResult: newOrFatal(t, result.DateTime{Date: time.Date(2024, time.March, 31, 1, 20, 30, 101e6, time.FixedZone("-07:00", -7*60*60)), Precision: model.MILLISECOND}),
		},
		{
			name:       "Time",
			cql:        "@T12",
			wantResult: newOrFatal(t, result.Time{Date: time.Date(0, time.January, 1, 12, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.HOUR}),
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

func TestIntervalSelector(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Integers",
			cql:  "Interval[1, 2)",
			wantModel: &model.Interval{
				Low:           model.NewLiteral("1", types.Integer),
				High:          model.NewLiteral("2", types.Integer),
				LowInclusive:  true,
				HighInclusive: false,
				Expression:    model.ResultType(&types.Interval{PointType: types.Integer}),
			},
			wantResult: newOrFatal(t, result.Interval{
				Low:           newOrFatal(t, 1),
				High:          newOrFatal(t, 2),
				LowInclusive:  true,
				HighInclusive: false,
				StaticType:    &types.Interval{PointType: types.Integer},
			}),
		},
		{
			name: "Longs",
			cql:  "Interval(1L, 3]",
			wantResult: newOrFatal(t, result.Interval{
				Low:           newOrFatal(t, int64(1)),
				High:          newOrFatal(t, int64(3)),
				LowInclusive:  false,
				HighInclusive: true,
				StaticType:    &types.Interval{PointType: types.Long},
			}),
		},
		{
			name: "Decimals",
			cql:  "Interval(1.0, 3]",
			wantResult: newOrFatal(t, result.Interval{
				Low:           newOrFatal(t, 1.0),
				High:          newOrFatal(t, 3.0),
				LowInclusive:  false,
				HighInclusive: true,
				StaticType:    &types.Interval{PointType: types.Decimal},
			}),
		},
		{
			name: "Date",
			cql:  "Interval(@2024-03, @2024-03]",
			wantResult: newOrFatal(t, result.Interval{
				Low:           newOrFatal(t, result.Date{Date: time.Date(2024, 3, 1, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.MONTH}),
				High:          newOrFatal(t, result.Date{Date: time.Date(2024, 3, 1, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.MONTH}),
				LowInclusive:  false,
				HighInclusive: true,
				StaticType:    &types.Interval{PointType: types.Date},
			}),
		},
		{
			name: "DateTime",
			cql:  "Interval(@2024-03-31T01:20:30.101Z, @2024-03-31T01:20:30.101Z]",
			wantResult: newOrFatal(t, result.Interval{
				Low:           newOrFatal(t, result.DateTime{Date: time.Date(2024, time.March, 31, 1, 20, 30, 101e6, time.UTC), Precision: model.MILLISECOND}),
				High:          newOrFatal(t, result.DateTime{Date: time.Date(2024, time.March, 31, 1, 20, 30, 101e6, time.UTC), Precision: model.MILLISECOND}),
				LowInclusive:  false,
				HighInclusive: true,
				StaticType:    &types.Interval{PointType: types.DateTime},
			}),
		},
		{
			name: "Left Null",
			cql:  "Interval[null, 2)",
			wantResult: newOrFatal(t, result.Interval{
				Low:           newOrFatal(t, nil),
				High:          newOrFatal(t, 2),
				LowInclusive:  true,
				HighInclusive: false,
				StaticType:    &types.Interval{PointType: types.Integer},
			}),
		},
		{
			name: "Right Null",
			cql:  "Interval[1, null)",
			wantResult: newOrFatal(t, result.Interval{
				Low:           newOrFatal(t, 1),
				High:          newOrFatal(t, nil),
				LowInclusive:  true,
				HighInclusive: false,
				StaticType:    &types.Interval{PointType: types.Integer},
			}),
		},
		{
			name: "Both null (with static types)",
			cql:  "Interval[null as Integer, null as Integer)",
			wantResult: newOrFatal(t, result.Interval{
				Low:           newOrFatal(t, nil),
				High:          newOrFatal(t, nil),
				LowInclusive:  true,
				HighInclusive: false,
				StaticType:    &types.Interval{PointType: types.Integer},
			}),
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

func TestListSelector(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Non mixed",
			cql:  "{1, 2}",
			wantModel: &model.List{
				Expression: model.ResultType(&types.List{ElementType: types.Integer}),
				List: []model.IExpression{
					model.NewLiteral("1", types.Integer),
					model.NewLiteral("2", types.Integer),
				},
			},
			wantResult: newOrFatal(t, result.List{
				Value:      []result.Value{newOrFatal(t, 1), newOrFatal(t, 2)},
				StaticType: &types.List{ElementType: types.Integer},
			}),
		},
		{
			name: "With type specifier",
			cql:  "List<Decimal>{1, 2.0}",
			wantResult: newOrFatal(t, result.List{
				Value:      []result.Value{newOrFatal(t, 1.0), newOrFatal(t, 2.0)},
				StaticType: &types.List{ElementType: types.Decimal},
			}),
		},
		{
			name: "Mixed implicitly convertible to same type",
			cql:  "{1, 2.0}",
			wantResult: newOrFatal(t, result.List{
				Value:      []result.Value{newOrFatal(t, 1.0), newOrFatal(t, 2.0)},
				StaticType: &types.List{ElementType: types.Decimal},
			}),
		},
		{
			name: "Mixed",
			cql:  "{1, 'hi'}",
			wantResult: newOrFatal(t, result.List{
				Value:      []result.Value{newOrFatal(t, 1), newOrFatal(t, "hi")},
				StaticType: &types.List{ElementType: &types.Choice{ChoiceTypes: []types.IType{types.Integer, types.String}}},
			}),
		},
		{
			name: "Null is converted based on type specifier",
			cql:  "List<Integer>{null}",
			wantModel: &model.List{
				Expression: model.ResultType(&types.List{ElementType: types.Integer}),
				List: []model.IExpression{
					&model.As{
						UnaryExpression: &model.UnaryExpression{
							Operand:    model.NewLiteral("null", types.Any),
							Expression: model.ResultType(types.Integer),
						},
						AsTypeSpecifier: types.Integer,
					},
				},
			},
			wantResult: newOrFatal(t, result.List{
				Value:      []result.Value{newOrFatal(t, nil)},
				StaticType: &types.List{ElementType: types.Integer},
			}),
		},
		{
			name: "Empty",
			cql:  "{}",
			wantResult: newOrFatal(t, result.List{
				Value:      []result.Value{},
				StaticType: &types.List{ElementType: types.Any},
			}),
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

func TestCodeSelectors(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Code Selector",
			cql: dedent.Dedent(`
			codesystem cs: 'https://example.com/cs/diagnosis' version '1.0'
			define TESTRESULT: Code '132' from cs display 'Severed Leg'`),
			wantModel: &model.Code{
				Expression: model.ResultType(types.Code),
				System: &model.CodeSystemRef{
					Name:       "cs",
					Expression: model.ResultType(types.CodeSystem),
				},
				Code:    "132",
				Display: "Severed Leg",
			},
			wantResult: newOrFatal(t,
				result.Code{
					Code:    "132",
					Display: "Severed Leg",
					System:  "https://example.com/cs/diagnosis",
					Version: "1.0",
				}),
		},
		{
			name: "Code Selector no display",
			cql: dedent.Dedent(`
			codesystem cs: 'https://example.com/cs/diagnosis' version '1.0'
			define TESTRESULT: Code '132' from cs`),
			wantResult: newOrFatal(t,
				result.Code{
					Code:    "132",
					System:  "https://example.com/cs/diagnosis",
					Version: "1.0",
				}),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := newFHIRParser(t)
			parsedLibs, err := p.Libraries(context.Background(), []string{tc.cql}, parser.Config{})
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

func TestTupleAndInstanceSelector(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name:       "Quantity Instance",
			cql:        "Quantity{value: 4, unit: 'day' }",
			wantResult: newOrFatal(t, result.Quantity{Value: 4, Unit: "day"}),
		},
		{
			name: "Code Instance",
			cql:  "Code{code: 'foo', system: 'bar', version: '1.0', display: 'severed leg' }",
			wantModel: &model.Instance{
				ClassType: types.Code,
				Elements: []*model.InstanceElement{
					{Name: "code", Value: model.NewLiteral("foo", types.String)},
					{Name: "system", Value: model.NewLiteral("bar", types.String)},
					{Name: "version", Value: model.NewLiteral("1.0", types.String)},
					{Name: "display", Value: model.NewLiteral("severed leg", types.String)},
				},
				Expression: model.ResultType(types.Code),
			},
			wantResult: newOrFatal(t, result.Code{Code: "foo", System: "bar", Display: "severed leg", Version: "1.0"}),
		},
		{
			name:       "CodeSystem Instance",
			cql:        "CodeSystem{id: 'id', version: '1.0' }",
			wantResult: newOrFatal(t, result.CodeSystem{ID: "id", Version: "1.0"}),
		},
		{
			name: "Concept Instance no codes",
			cql:  "Concept{codes: {} }",
			wantResult: newOrFatal(t, result.Concept{
				Codes: []*result.Code{},
			}),
		},
		{
			name: "Concept Instance with null codes",
			cql:  "Concept{codes: { null as Code, null as Code } }",
			wantResult: newOrFatal(t, result.Concept{
				Codes: []*result.Code{nil, nil},
			}),
		},
		{
			name: "Concept Instance",
			cql:  "Concept{codes: {Code{code: 'foo', system: 'bar', version: '1.0' }}, display: 'display' }",
			wantResult: newOrFatal(t, result.Concept{
				Codes:   []*result.Code{{Code: "foo", System: "bar", Version: "1.0"}},
				Display: "display",
			}),
		},
		{
			name: "ValueSet Instance",
			cql:  "ValueSet{id: 'id', version: '1.0', codesystems: List<CodeSystem>{CodeSystem{id: 'id', version: '1.0' }}}",
			wantResult: newOrFatal(t, result.ValueSet{
				ID:          "id",
				Version:     "1.0",
				CodeSystems: []result.CodeSystem{{ID: "id", Version: "1.0"}},
			}),
		},
		{
			name: "FHIR Instance",
			// The test setup wraps the CQL expression in a library which includes context Patient so we
			// can use Patient.gender.
			cql: "Patient { gender: Patient.gender }",
			wantResult: newOrFatal(t, result.Tuple{
				Value: map[string]result.Value{
					"gender": newOrFatal(t, result.Named{
						Value:       &r4patientpb.Patient_GenderCode{Value: c4pb.AdministrativeGenderCode_MALE},
						RuntimeType: &types.Named{TypeName: "FHIR.AdministrativeGender"},
					}),
				},
				RuntimeType: &types.Named{TypeName: "FHIR.Patient"},
			}),
		},
		{
			name: "Tuple",
			cql:  "Tuple { apple: 'red', banana: 4 as Choice<Integer, String> }",
			wantModel: &model.Tuple{
				Elements: []*model.TupleElement{
					{Name: "apple", Value: model.NewLiteral("red", types.String)},
					{
						Name: "banana",
						Value: &model.As{
							UnaryExpression: &model.UnaryExpression{
								Expression: model.ResultType(&types.Choice{ChoiceTypes: []types.IType{types.Integer, types.String}}),
								Operand:    model.NewLiteral("4", types.Integer),
							},
							AsTypeSpecifier: &types.Choice{ChoiceTypes: []types.IType{types.Integer, types.String}},
						},
					},
				},
				Expression: model.ResultType(&types.Tuple{ElementTypes: map[string]types.IType{"apple": types.String, "banana": &types.Choice{ChoiceTypes: []types.IType{types.Integer, types.String}}}}),
			},
			wantResult: newOrFatal(t, result.Tuple{
				Value: map[string]result.Value{
					"apple":  newOrFatal(t, "red"),
					"banana": newOrFatal(t, 4),
				},
				RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{"apple": types.String, "banana": &types.Choice{ChoiceTypes: []types.IType{types.Integer, types.String}}}},
			}),
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
