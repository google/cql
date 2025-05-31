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
	"strings"
	"testing"
	"time"

	"github.com/google/cql/interpreter"
	"github.com/google/cql/model"
	"github.com/google/cql/parser"
	"github.com/google/cql/result"
	"github.com/google/cql/types"
	"github.com/google/go-cmp/cmp"
	"github.com/lithammer/dedent"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestAs(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		// TODO(b/301606416): Add a test case for subtypes (ex Procedure As ImagingProcedure). This is
		// not possible until Named types or Vocabulary type are supported.
		{
			name: "Null",
			cql:  "null as Integer",
			wantModel: &model.As{
				UnaryExpression: &model.UnaryExpression{
					Operand:    model.NewLiteral("null", types.Any),
					Expression: model.ResultType(types.Integer),
				},
				AsTypeSpecifier: types.Integer,
			},
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Integer As Integer",
			cql:        "4 as Integer",
			wantResult: newOrFatal(t, 4),
		},
		{
			name:       "Integer As Any",
			cql:        "4 as Any",
			wantResult: newOrFatal(t, 4),
		},
		{
			name:       "Integer As Choice<String, Integer>",
			cql:        "4 as Choice<String, Integer>",
			wantResult: newOrFatal(t, 4),
		},
		{
			name: "Choice<String, Integer> As Integer",
			cql:  "4 as Choice<String, Integer> as Integer",
			wantModel: &model.As{
				UnaryExpression: &model.UnaryExpression{
					Operand: &model.As{
						UnaryExpression: &model.UnaryExpression{
							Operand:    model.NewLiteral("4", types.Integer),
							Expression: model.ResultType(&types.Choice{ChoiceTypes: []types.IType{types.String, types.Integer}}),
						},
						AsTypeSpecifier: &types.Choice{ChoiceTypes: []types.IType{types.String, types.Integer}},
					},
					Expression: model.ResultType(types.Integer),
				},
				AsTypeSpecifier: types.Integer},
			wantResult: newOrFatal(t, 4),
		},
		{
			name:       "Integer As Decimal Null",
			cql:        "4 as Decimal",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Integer As Choice<String, Decimal> Null",
			cql:        "4 as Choice<String, Decimal>",
			wantResult: newOrFatal(t, nil),
		},
		{
			name: "Strict cast Integer as Any",
			cql:  "cast 4 as Any",
			wantModel: &model.As{
				UnaryExpression: &model.UnaryExpression{
					Operand:    model.NewLiteral("4", types.Integer),
					Expression: model.ResultType(types.Any),
				},
				Strict:          true,
				AsTypeSpecifier: types.Any,
			},
			wantResult: newOrFatal(t, 4),
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

func TestAs_EvalErrors(t *testing.T) {
	tests := []struct {
		name                string
		cql                 string
		wantModel           model.IExpression
		wantEvalErrContains string
	}{
		{
			name: "Strict Integer as Decimal",
			cql:  "cast 4 as Decimal",
			wantModel: &model.As{
				UnaryExpression: &model.UnaryExpression{
					Operand:    model.NewLiteral("4", types.Integer),
					Expression: model.ResultType(types.Decimal),
				},
				Strict:          true,
				AsTypeSpecifier: types.Decimal,
			},
			wantEvalErrContains: "cannot strict cast",
		},
		{
			name:                "Strict Integer As Choice<String, Decimal>",
			cql:                 "cast 4 as Choice<String, Decimal>",
			wantEvalErrContains: "cannot strict cast",
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

			_, err = interpreter.Eval(context.Background(), parsedLibs, defaultInterpreterConfig(t, p))
			if err == nil {
				t.Fatalf("Evaluate Expression expected an error to be returned, got nil instead")
			}
			if !strings.Contains(err.Error(), tc.wantEvalErrContains) {
				t.Errorf("Unexpected evaluation error contents got (%v) want (%v)", err.Error(), tc.wantEvalErrContains)
			}

		})
	}
}

func TestConvertQuantity(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "ConvertQuantity cm to m",
			cql:  "ConvertQuantity(1 'cm', 'm')",
			wantModel: &model.ConvertQuantity{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						&model.Quantity{Value: 1, Unit: "cm", Expression: model.ResultType(types.Quantity)},
						model.NewLiteral("m", types.String),
					},
					Expression: model.ResultType(types.Quantity),
				},
			},
			wantResult: newOrFatal(t, result.Quantity{Value: 0.01, Unit: "m"}),
		},
		{
			name:       "ConvertQuantity m to cm",
			cql:        "ConvertQuantity(1 'm', 'cm')",
			wantResult: newOrFatal(t, result.Quantity{Value: 100, Unit: "cm"}),
		},
		{
			name:       "ConvertQuantity to same unit",
			cql:        "ConvertQuantity(1 'cm', 'cm')",
			wantResult: newOrFatal(t, result.Quantity{Value: 1, Unit: "cm"}),
		},
		{
			name:       "ConvertQuantity Invalid Unit",
			cql:        "ConvertQuantity(1 'cm', 'invalid')",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "ConvertQuantity left null",
			cql:        "ConvertQuantity(null as Quantity, 'm')",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "ConvertQuantity right null",
			cql:        "ConvertQuantity(1 'cm', null as String)",
			wantResult: newOrFatal(t, nil),
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

func TestToDate(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "DateTime",
			cql:  "ToDate(@2024-03-31T01:20:30.101-07:00)",
			wantModel: &model.ToDate{
				UnaryExpression: &model.UnaryExpression{
					Operand:    model.NewLiteral("@2024-03-31T01:20:30.101-07:00", types.DateTime),
					Expression: model.ResultType(types.Date),
				},
			},
			wantResult: newOrFatal(t, result.Date{Date: time.Date(2024, time.March, 31, 1, 20, 30, 101e6, time.FixedZone("", -25200)), Precision: model.DAY}),
		},
		{
			name: "String",
			cql:  "ToDate('2024-03-31')",
			wantModel: &model.ToDate{
				UnaryExpression: &model.UnaryExpression{
					Operand:    model.NewLiteral("2024-03-31", types.String),
					Expression: model.ResultType(types.Date),
				},
			},
			wantResult: newOrFatal(t, result.Date{Date: time.Date(2024, time.March, 31, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.DAY}),
		},
		{
			name:       "Date",
			cql:        "ToDate(@2024-03-31)",
			wantResult: newOrFatal(t, result.Date{Date: time.Date(2024, time.March, 31, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.DAY}),
		},
		{
			name:       "Null",
			cql:        "ToDate(null as String)", // as String necessary to prevent ambiguous match.
			wantResult: newOrFatal(t, nil),
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

func TestToDateTime(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Date",
			cql:  "ToDateTime(@2012-12-02)",
			wantModel: &model.ToDateTime{
				UnaryExpression: &model.UnaryExpression{
					Operand:    model.NewLiteral("@2012-12-02", types.Date),
					Expression: model.ResultType(types.DateTime),
				},
			},
			wantResult: newOrFatal(t, result.DateTime{Date: time.Date(2012, time.December, 2, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.DAY}),
		},
		{
			name:       "Null",
			cql:        "ToDateTime(null as Date)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "String",
			cql:        "ToDateTime('2024-03-31T01:20:30.101-07:00')",
			wantResult: newOrFatal(t, result.DateTime{Date: time.Date(2024, time.March, 31, 1, 20, 30, 101e6, time.FixedZone("", -25200)), Precision: model.MILLISECOND}),
		},
		{
			name:       "DateTime",
			cql:        "ToDateTime(@2024-03-31T01:20:30.101-07:00)",
			wantResult: newOrFatal(t, result.DateTime{Date: time.Date(2024, time.March, 31, 1, 20, 30, 101e6, time.FixedZone("", -25200)), Precision: model.MILLISECOND}),
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

func TestToDecimal(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Null to Decimal",
			cql:  "ToDecimal(null as Integer)",
			wantModel: &model.ToDecimal{
				UnaryExpression: &model.UnaryExpression{
					Operand: &model.As{
						UnaryExpression: &model.UnaryExpression{
							Operand:    model.NewLiteral("null", types.Any),
							Expression: model.ResultType(types.Integer),
						},
						AsTypeSpecifier: types.Integer,
					},
					Expression: model.ResultType(types.Decimal),
				},
			},
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Decimal to Decimal",
			cql:        "ToDecimal(412.2)",
			wantResult: newOrFatal(t, 412.2),
		},
		{
			name:       "Long to Decimal",
			cql:        "ToDecimal(412L)",
			wantResult: newOrFatal(t, 412.0),
		},
		{
			name:       "Integer to Decimal",
			cql:        "ToDecimal(412)",
			wantResult: newOrFatal(t, 412.0),
		},
		{
			name:       "True to Decimal",
			cql:        "ToDecimal(true)",
			wantResult: newOrFatal(t, 1.0),
		},
		{
			name:       "False to Decimal",
			cql:        "ToDecimal(false)",
			wantResult: newOrFatal(t, 0.0),
		},
		{
			name:       "Negative String to Decimal",
			cql:        "ToDecimal('-412.2')",
			wantResult: newOrFatal(t, -412.2),
		},
		{
			name:       "Positive String to Decimal",
			cql:        "ToDecimal('+412')",
			wantResult: newOrFatal(t, 412.0),
		},
		{
			name:       "No Sign String",
			cql:        "ToDecimal('412.00')",
			wantResult: newOrFatal(t, 412.0),
		},
		{
			name:       "Invalid String Additional Character",
			cql:        "ToDecimal('123.4a')",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Invalid String Missing Digit",
			cql:        "ToDecimal('-.4')",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Invalid Empty String",
			cql:        "ToDecimal('')",
			wantResult: newOrFatal(t, nil),
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

func TestToLong(t *testing.T) {
	tests := []struct {
		cql        string
		name       string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Null to Long",
			cql:  "ToLong(null as Integer)",
			wantModel: &model.ToLong{
				UnaryExpression: &model.UnaryExpression{
					Operand: &model.As{
						UnaryExpression: &model.UnaryExpression{
							Operand:    model.NewLiteral("null", types.Any),
							Expression: model.ResultType(types.Integer),
						},
						AsTypeSpecifier: types.Integer,
					},
					Expression: model.ResultType(types.Long),
				},
			},
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Long to Long",
			cql:        "ToLong(412L)",
			wantResult: newOrFatal(t, int64(412)),
		},
		{
			name:       "Integer to Long",
			cql:        "ToLong(412)",
			wantResult: newOrFatal(t, int64(412)),
		},
		{
			name:       "True to Long",
			cql:        "ToLong(true)",
			wantResult: newOrFatal(t, int64(1)),
		},
		{
			name:       "False to Long",
			cql:        "ToLong(false)",
			wantResult: newOrFatal(t, int64(0)),
		},
		{
			name:       "Negative String to Long",
			cql:        "ToLong('-412')",
			wantResult: newOrFatal(t, int64(-412)),
		},
		{
			name:       "Positive String to Long",
			cql:        "ToLong('+412')",
			wantResult: newOrFatal(t, int64(412)),
		},
		{
			name:       "No Sign String",
			cql:        "ToLong('412')",
			wantResult: newOrFatal(t, int64(412)),
		},
		{
			name:       "Invalid String Additional Character",
			cql:        "ToLong('1234a')",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Invalid String Missing Digit",
			cql:        "ToLong('-')",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Invalid Empty String",
			cql:        "ToLong('')",
			wantResult: newOrFatal(t, nil),
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

func TestToQuantity(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Null to Quantity",
			cql:  "ToQuantity(null as Integer)",
			wantModel: &model.ToQuantity{
				UnaryExpression: &model.UnaryExpression{
					Operand: &model.As{
						UnaryExpression: &model.UnaryExpression{
							Operand:    model.NewLiteral("null", types.Any),
							Expression: model.ResultType(types.Integer),
						},
						AsTypeSpecifier: types.Integer,
					},
					Expression: model.ResultType(types.Quantity),
				},
			},
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Integer to Quantity",
			cql:        "ToQuantity(412)",
			wantResult: newOrFatal(t, result.Quantity{Value: 412.0, Unit: "1"}),
		},
		{
			name:       "Decimal to Quantity",
			cql:        "ToQuantity(412.2)",
			wantResult: newOrFatal(t, result.Quantity{Value: 412.2, Unit: "1"}),
		},
		// String ToQuantity tests
		{
			name:       "Negative String to Quantity",
			cql:        "ToQuantity('-412.2')",
			wantResult: newOrFatal(t, result.Quantity{Value: -412.2, Unit: "1"}),
		},
		{
			name:       "Positive String to Quantity",
			cql:        "ToQuantity('+412.2')",
			wantResult: newOrFatal(t, result.Quantity{Value: 412.2, Unit: "1"}),
		},
		{
			name:       "String decimal only to Quantity",
			cql:        "ToQuantity('412.2')",
			wantResult: newOrFatal(t, result.Quantity{Value: 412.2, Unit: "1"}),
		},
		{
			name: "String decimal and unit to Quantity",
			cql:  "ToQuantity('412.2 \\'cm\\'')",
			wantModel: &model.ToQuantity{
				UnaryExpression: &model.UnaryExpression{
					Operand:    model.NewLiteral("412.2 'cm'", types.String),
					Expression: model.ResultType(types.Quantity),
				},
			},
			wantResult: newOrFatal(t, result.Quantity{Value: 412.2, Unit: "cm"}),
		},
		{
			name:       "String decimal and unit no spaces to Quantity",
			cql:        "ToQuantity('412.2\\'cm\\'')",
			wantResult: newOrFatal(t, result.Quantity{Value: 412.2, Unit: "cm"}),
		},
		// Invalid String tests
		{
			name:       "Invalid Empty String to Quantity",
			cql:        "ToQuantity('')",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Invalid String Missing Digit and Unit to Quantity",
			cql:        "ToQuantity('-')",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Invalid String Missing Digit to Quantity",
			cql:        "ToQuantity('-.4')",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Invalid String Missing right quotation to Quantity",
			cql:        "ToQuantity('-1.4 \\'cm')",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Invalid String Missing left quotation to Quantity",
			cql:        "ToQuantity('-1.4 cm\\')",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Invalid String Missing both quotations to Quantity",
			cql:        "ToQuantity('-1.4 cm')",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Invalid String Missing only unit to Quantity",
			cql:        "ToQuantity('\\'cm\\')",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Invalid String unit not valid to Quantity",
			cql:        "ToQuantity('1.0\\'asdf\\'')",
			wantResult: newOrFatal(t, nil),
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

func TestToConcept(t *testing.T) {
	tests := []struct {
		cql        string
		name       string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "null Code",
			cql:  "define TESTRESULT: ToConcept(null as Code)",
			wantModel: &model.ToConcept{
				UnaryExpression: &model.UnaryExpression{
					Operand: &model.As{
						UnaryExpression: &model.UnaryExpression{
							Operand:    model.NewLiteral("null", types.Any),
							Expression: model.ResultType(types.Code),
						},
						AsTypeSpecifier: types.Code,
					},
					Expression: model.ResultType(types.Concept),
				},
			},
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "null List<Code>",
			cql:        "define TESTRESULT: ToConcept(null as List<Code>)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name: "Code no display",
			cql: dedent.Dedent(`
			codesystem cs: 'https://example.com/cs/diagnosis' version '1.0'
			define TESTRESULT: ToConcept(Code '132' from cs)`),
			wantResult: newOrFatal(t, result.Concept{
				Codes: []*result.Code{
					{
						Code:    "132",
						System:  "https://example.com/cs/diagnosis",
						Version: "1.0",
					},
				}}),
		},
		{
			name: "Code with display",
			cql: dedent.Dedent(`
			codesystem cs: 'https://example.com/cs/diagnosis' version '1.0'
			define TESTRESULT: ToConcept(Code '132' from cs display 'Severed Leg')`),
			wantResult: newOrFatal(t, result.Concept{
				Display: "Severed Leg",
				Codes: []*result.Code{
					{
						Code:    "132",
						System:  "https://example.com/cs/diagnosis",
						Version: "1.0",
						Display: "Severed Leg",
					},
				}}),
		},
		{
			name: "List of codes",
			cql: dedent.Dedent(`
			codesystem cs: 'https://example.com/cs/diagnosis' version '1.0'
			define TESTRESULT: ToConcept({Code '132' from cs display 'Severed Leg', Code '444' from cs display 'Burnt Cranium'})`),
			wantResult: newOrFatal(t, result.Concept{
				Codes: []*result.Code{
					{
						Code:    "132",
						System:  "https://example.com/cs/diagnosis",
						Version: "1.0",
						Display: "Severed Leg",
					},
					{
						Code:    "444",
						System:  "https://example.com/cs/diagnosis",
						Version: "1.0",
						Display: "Burnt Cranium",
					},
				}}),
		},
		{
			name:       "ToConcept with null Code",
			cql:        "define TESTRESULT: ToConcept(List<Code>{null})",
			wantResult: newOrFatal(t, result.Concept{Codes: []*result.Code{nil}}),
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

func TestIs(t *testing.T) {
	// Note that cases for "is true", "is null", "is false", are in operator_nullological_test.go
	// since they are considered nullological operators in CQL:
	// https://cql.hl7.org/09-b-cqlreference.html#nullological-operators-3
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "1 is Integer",
			cql:  "1 is Integer",
			wantModel: &model.Is{
				UnaryExpression: &model.UnaryExpression{
					Operand:    model.NewLiteral("1", types.Integer),
					Expression: model.ResultType(types.Boolean),
				},
				IsTypeSpecifier: types.Integer,
			},
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "1 is String",
			cql:        "1 is String",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "1 is Any (Any is subtype of Integer)",
			cql:        "1 is Any",
			wantResult: newOrFatal(t, true),
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
