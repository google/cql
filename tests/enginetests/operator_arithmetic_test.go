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
	"math"
	"strings"
	"testing"
	"time"

	"github.com/google/cql/interpreter"
	"github.com/google/cql/model"
	"github.com/google/cql/parser"
	"github.com/google/cql/result"
	"github.com/google/cql/types"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestAbs(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Integer",
			cql:  "Abs(-4)",
			wantModel: &model.Abs{
				UnaryExpression: &model.UnaryExpression{
					Operand: &model.Negate{
						UnaryExpression: &model.UnaryExpression{
							Operand:    model.NewLiteral("4", types.Integer),
							Expression: model.ResultType(types.Integer),
						},
					},
					Expression: model.ResultType(types.Integer),
				},
			},
			wantResult: newOrFatal(t, 4),
		},
		{
			name:       "Positive Integer",
			cql:        "Abs(2)",
			wantResult: newOrFatal(t, 2),
		},
		{
			name:       "Minimum Integer",
			cql:        "Abs(-2147483648)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Long",
			cql:        "Abs(-4L)",
			wantResult: newOrFatal(t, int64(4)),
		},
		{
			name:       "Positive Long",
			cql:        "Abs(2L)",
			wantResult: newOrFatal(t, int64(2)),
		},
		{
			name:       "Minimum Long",
			cql:        "Abs(-9223372036854775808L)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Decimal",
			cql:        "Abs(-1.0)",
			wantResult: newOrFatal(t, 1.0),
		},
		{
			name:       "Positive Decimal",
			cql:        "Abs(1.0)",
			wantResult: newOrFatal(t, 1.0),
		},
		{
			name:       "Minimum Decimal",
			cql:        "Abs(-99999999999999999999.99999999)",
			wantResult: newOrFatal(t, float64(99999999999999999999.99999999)),
		},
		{
			name:       "Quantity",
			cql:        "Abs(-1.0 'day')",
			wantResult: newOrFatal(t, result.Quantity{Value: 1.0, Unit: model.DAYUNIT}),
		},
		{
			name:       "Positive Quantity",
			cql:        "Abs(1.0 'day')",
			wantResult: newOrFatal(t, result.Quantity{Value: 1.0, Unit: model.DAYUNIT}),
		},
		{
			name:       "Quantity",
			cql:        "Abs(-99999999999999999999.99999999 'day')",
			wantResult: newOrFatal(t, result.Quantity{Value: 99999999999999999999.99999999, Unit: model.DAYUNIT}),
		},
		{
			name:       "Null",
			cql:        "Abs(null as Integer)",
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

func TestCeiling(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Decimal",
			cql:  "Ceiling(41.1)",
			wantModel: &model.Ceiling{
				UnaryExpression: &model.UnaryExpression{
					Operand:    model.NewLiteral("41.1", types.Decimal),
					Expression: model.ResultType(types.Integer),
				},
			},
			wantResult: newOrFatal(t, 42),
		},
		{
			name:       "Negative",
			cql:        "Ceiling(-2.1)",
			wantResult: newOrFatal(t, -2),
		},
		{
			name:       "Zero",
			cql:        "Ceiling(0.0)",
			wantResult: newOrFatal(t, 0),
		},
		{
			name:       "Integer",
			cql:        "Ceiling(2)",
			wantResult: newOrFatal(t, 2),
		},
		{
			name:       "Minimum Integer",
			cql:        "Ceiling(-2147483648)",
			wantResult: newOrFatal(t, -2147483648),
		},
		{
			name:       "Minimum Decimal out of range",
			cql:        "Ceiling(-99999999999999999999.99999999)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Maximum Decimal out of range",
			cql:        "Ceiling(99999999999999999999.99999999)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Just less than min int32",
			cql:        "Ceiling(-2147483648.5)",
			wantResult: newOrFatal(t, math.MinInt32),
		},
		{
			name:       "More than one less than min int32",
			cql:        "Ceiling(-2147483649.5)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Just more than max int32",
			cql:        "Ceiling(2147483647.5)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "equal to min int32",
			cql:        "Ceiling(-2147483648.0)",
			wantResult: newOrFatal(t, math.MinInt32),
		},
		{
			name:       "equal to max int32",
			cql:        "Ceiling(2147483647.0)",
			wantResult: newOrFatal(t, math.MaxInt32),
		},
		{
			name:       "Null",
			cql:        "Ceiling(null as Decimal)",
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

func TestExp(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Integer",
			cql:  "Exp(4)",
			wantModel: &model.Exp{
				UnaryExpression: &model.UnaryExpression{
					Operand: &model.ToDecimal{
						UnaryExpression: &model.UnaryExpression{
							Operand:    model.NewLiteral("4", types.Integer),
							Expression: model.ResultType(types.Decimal),
						},
					},
					Expression: model.ResultType(types.Decimal),
				},
			},
			wantResult: newOrFatal(t, 54.598150033144236),
		},
		{
			name:       "Positive Integer",
			cql:        "Exp(2)",
			wantResult: newOrFatal(t, 7.38905609893065),
		},
		{
			name:       "Minimum Integer",
			cql:        "Exp(-2147483648)",
			wantResult: newOrFatal(t, 0.0),
		},
		{
			name:       "Long",
			cql:        "Exp(-4L)",
			wantResult: newOrFatal(t, 0.01831563888873418),
		},
		{
			name:       "Positive Long",
			cql:        "Exp(2L)",
			wantResult: newOrFatal(t, 7.38905609893065),
		},
		{
			name:       "Minimum Long",
			cql:        "Exp(-9223372036854775808L)",
			wantResult: newOrFatal(t, 0.0),
		},
		{
			name:       "Decimal zero",
			cql:        "Exp(0.0)",
			wantResult: newOrFatal(t, 1.0),
		},
		{
			name:       "Decimal negative one",
			cql:        "Exp(-1.0)",
			wantResult: newOrFatal(t, 0.36787944117144233),
		},
		{
			name:       "Positive Decimal one",
			cql:        "Exp(1.0)",
			wantResult: newOrFatal(t, 2.718281828459045),
		},
		{
			name:       "Minimum Decimal",
			cql:        "Exp(-99999999999999999999.99999999)",
			wantResult: newOrFatal(t, 0.0),
		},
		{
			name:       "Null",
			cql:        "Exp(null as Decimal)",
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

func TestFloor(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Decimal",
			cql:  "Floor(42.1)",
			wantModel: &model.Floor{
				UnaryExpression: &model.UnaryExpression{
					Operand:    model.NewLiteral("42.1", types.Decimal),
					Expression: model.ResultType(types.Integer),
				},
			},
			wantResult: newOrFatal(t, 42),
		},
		{
			name:       "Negative",
			cql:        "Floor(-2.1)",
			wantResult: newOrFatal(t, -3),
		},
		{
			name:       "Zero",
			cql:        "Floor(0.0)",
			wantResult: newOrFatal(t, 0),
		},
		{
			name:       "Integer",
			cql:        "Floor(2)",
			wantResult: newOrFatal(t, 2),
		},
		{
			name:       "Minimum Integer",
			cql:        "Floor(-2147483648)",
			wantResult: newOrFatal(t, -2147483648),
		},
		{
			name:       "Minimum Decimal out of range",
			cql:        "Floor(-99999999999999999999.99999999)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Maximum Decimal out of range",
			cql:        "Floor(99999999999999999999.99999999)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Just less than min int32",
			cql:        "Floor(-2147483648.5)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Just more than max int32",
			cql:        "Floor(2147483647.5)",
			wantResult: newOrFatal(t, math.MaxInt32),
		},
		{
			name:       "More than one more than max int32",
			cql:        "Floor(2147483648.5)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "equal to min int32",
			cql:        "Floor(-2147483648.0)",
			wantResult: newOrFatal(t, math.MinInt32),
		},
		{
			name:       "equal to max int32",
			cql:        "Floor(2147483647.0)",
			wantResult: newOrFatal(t, math.MaxInt32),
		},
		{
			name:       "Null",
			cql:        "Floor(null as Decimal)",
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

func TestLn(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Decimal",
			cql:  "Ln(1.0)",
			wantModel: &model.Ln{
				UnaryExpression: &model.UnaryExpression{
					Operand:    model.NewLiteral("1.0", types.Decimal),
					Expression: model.ResultType(types.Decimal),
				},
			},
			wantResult: newOrFatal(t, 0.0),
		},
		{
			name:       "Negative",
			cql:        "Ln(-2.1)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Zero",
			cql:        "Ln(0.0)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Ten",
			cql:        "Ln(10.0)",
			wantResult: newOrFatal(t, 2.302585092994046),
		},
		{
			name:       "Integer",
			cql:        "Ln(1)",
			wantResult: newOrFatal(t, 0.0),
		},
		{
			name:       "Minimum Integer",
			cql:        "Ln(-2147483648)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Maximum Integer",
			cql:        "Ln(2147483647.0)",
			wantResult: newOrFatal(t, 21.487562596892644),
		},
		{
			name:       "Null",
			cql:        "Ln(null as Decimal)",
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

func TestLog(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Decimal",
			cql:  "Log(1.0, 10.0)",
			wantModel: &model.Log{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("1.0", types.Decimal),
						model.NewLiteral("10.0", types.Decimal),
					},
					Expression: model.ResultType(types.Decimal),
				},
			},
			wantResult: newOrFatal(t, 0.0),
		},
		{
			name:       "Negative value",
			cql:        "Log(-2.1, 10.0)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Negative base",
			cql:        "Log(2.1, -10.0)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Zero value",
			cql:        "Log(0.0, 10.0)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Zero base",
			cql:        "Log(2.1, 0.0)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Logarithm of 0.125 with base 2",
			cql:        "Log(0.125, 2.0)",
			wantResult: newOrFatal(t, -3.0),
		},
		{
			name:       "Integer of 16 with base 2",
			cql:        "Log(16, 2)",
			wantResult: newOrFatal(t, 4.0),
		},
		{
			name:       "Minimum Integer value",
			cql:        "Log(-2147483648, 10)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Maximum Integer value",
			cql:        "Round(Log(2147483647, 10), 3)",
			wantResult: newOrFatal(t, 9.332),
		},
		{
			name:       "Null value",
			cql:        "Log(null as Decimal, 10)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Null base",
			cql:        "Log(1.0, null as Decimal)",
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

func TestPrecision(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Decimal",
			cql:  "Precision(@2014)",
			wantModel: &model.Precision{
				UnaryExpression: &model.UnaryExpression{
					Operand:    model.NewLiteral("@2014", types.Date),
					Expression: model.ResultType(types.Integer),
				},
			},
			wantResult: newOrFatal(t, 4),
		},
		{
			name:       "Date Year",
			cql:        "Precision(@2014)",
			wantResult: newOrFatal(t, 4),
		},
		{
			name:       "Date Month",
			cql:        "Precision(@2014-02)",
			wantResult: newOrFatal(t, 6),
		},
		{
			name:       "Date Day",
			cql:        "Precision(@2014-01-01)",
			wantResult: newOrFatal(t, 8),
		},
		{
			name:       "DateTime Hour",
			cql:        "Precision(@2014-01-01T10)",
			wantResult: newOrFatal(t, 10),
		},
		{
			name:       "DateTime Minute",
			cql:        "Precision(@2014-01-01T10:10)",
			wantResult: newOrFatal(t, 12),
		},
		{
			name:       "DateTime Second",
			cql:        "Precision(@2014-01-01T10:10:30)",
			wantResult: newOrFatal(t, 14),
		},
		{
			name:       "DateTime Millisecond",
			cql:        "Precision(@2014-01-01T10:10:30.000Z)",
			wantResult: newOrFatal(t, 17),
		},
		{
			name:       "Time Hour",
			cql:        "Precision(@T10)",
			wantResult: newOrFatal(t, 2),
		},
		{
			name:       "Time Millisecond",
			cql:        "Precision(@T01:01:00.000)",
			wantResult: newOrFatal(t, 9),
		},
		{
			name:       "Time Millisecond",
			cql:        "Precision(@2014-01-02T01:01:00.000Z)",
			wantResult: newOrFatal(t, 17),
		},
		{
			name:       "Null",
			cql:        "Precision(null as Date)",
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

func TestAdd(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name:       "Integers",
			cql:        "1 + 2",
			wantResult: newOrFatal(t, 3),
		},
		{
			name: "Longs",
			cql:  "1L + 2",
			wantModel: &model.Add{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("1L", types.Long),
						&model.ToLong{
							UnaryExpression: &model.UnaryExpression{
								Expression: model.ResultType(types.Long),
								Operand:    model.NewLiteral("2", types.Integer),
							},
						},
					},
					Expression: model.ResultType(types.Long),
				},
			},
			wantResult: newOrFatal(t, int64(3)),
		},
		{
			name:       "Decimals",
			cql:        "1.5 + 2",
			wantResult: newOrFatal(t, 3.5),
		},
		{
			name:       "Quantity",
			cql:        "11 'day' + 9 'day'",
			wantResult: newOrFatal(t, result.Quantity{Value: 20, Unit: model.DAYUNIT}),
		},
		{
			name:       "Quantity via class instances",
			cql:        "Quantity{value: 11, unit: 'day'} + 9 'day'",
			wantResult: newOrFatal(t, result.Quantity{Value: 20, Unit: model.DAYUNIT}),
		},
		// Tests for Date and Quantity
		// TODO(b/301606416): Add more tests for DateTime + Quantity
		{
			name:       "Date Quantity",
			cql:        "@2014 + 1 'year'",
			wantResult: newOrFatal(t, result.Date{Date: time.Date(2015, time.January, 1, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.YEAR}),
		},
		{
			name: "Date month precision add year",
			cql:  "@2014-01 + 1 'year'",
			wantModel: &model.Add{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("@2014-01", types.Date),
						&model.Quantity{Value: 1, Unit: model.YEARUNIT, Expression: model.ResultType(types.Quantity)},
					},
					Expression: model.ResultType(types.Date),
				},
			},
			wantResult: newOrFatal(t, result.Date{Date: time.Date(2015, time.January, 1, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.MONTH}),
		},
		{
			name:       "Date year precision add month",
			cql:        "@2014 + 12 'month'",
			wantResult: newOrFatal(t, result.Date{Date: time.Date(2015, time.January, 1, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.YEAR}),
		},
		{
			name:       "Date day precision add week to day precision",
			cql:        "@2014-01-01 + 2 'week'",
			wantResult: newOrFatal(t, result.Date{Date: time.Date(2014, time.January, 15, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.DAY}),
		},
		{
			name:       "Date year precision add 1.6 year truncates to 1",
			cql:        "@2014-01-01 + 1.6 'year'",
			wantResult: newOrFatal(t, result.Date{Date: time.Date(2015, time.January, 1, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.DAY}),
		},
		{
			name:       "Date year precision add 6 months truncates to 0",
			cql:        "@2014-01-01 + 6 'month'",
			wantResult: newOrFatal(t, result.Date{Date: time.Date(2014, time.July, 1, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.DAY}),
		},
		{
			name:       "Date time millisecond precision add 1.6 second does not truncate",
			cql:        "@2014-01-01T00:00:00.000Z + 1.6 'second'",
			wantResult: newOrFatal(t, result.DateTime{Date: time.Date(2014, time.January, 1, 0, 0, 1, 600_000_000, time.UTC), Precision: "millisecond"}),
		},
		// Tests for Nulls
		{
			name: "Integer Null",
			cql:  "1 + null",
			wantModel: &model.Add{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("1", types.Integer),
						&model.As{
							UnaryExpression: &model.UnaryExpression{
								Expression: model.ResultType(types.Integer),
								Operand:    model.NewLiteral("null", types.Any),
							},
							AsTypeSpecifier: types.Integer,
						},
					},
					Expression: model.ResultType(types.Integer),
				},
			},
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Null Integer",
			cql:        "null + 2",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Long Null",
			cql:        "null + 2L",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Null Long",
			cql:        "1L + null",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Decimal Null",
			cql:        "1.5 + null",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Null Decimal",
			cql:        "null + 1.5",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Quantity Null",
			cql:        "11 'day' + null",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Date Null",
			cql:        "@2014 + null",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "DateTime Null",
			cql:        "@2014-01-01T00:00:00.000Z + null",
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

func TestAdd_EvalErrors(t *testing.T) {
	tests := []struct {
		name                string
		cql                 string
		wantModel           model.IExpression
		wantEvalErrContains string
	}{
		{
			name:                "Date month precision add day returns conversion error",
			cql:                 "@2014-01 + 1 'day'",
			wantEvalErrContains: "invalid unit conversion",
		},
		{
			name:                "Date month precision add week returns conversion error",
			cql:                 "@2014-01 + 1 'week'",
			wantEvalErrContains: "cannot convert from week to a higher precision",
		},
		{
			name:                "Date month precision add minutes returns conversion error",
			cql:                 "@2014-01 + 1 'minute'",
			wantEvalErrContains: "invalid unit conversion",
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

func TestSubtract(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name:       "Integers",
			cql:        "1 - 2",
			wantResult: newOrFatal(t, -1),
		},
		{
			name: "Longs",
			cql:  "1L - 2",
			wantModel: &model.Subtract{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("1L", types.Long),
						&model.ToLong{
							UnaryExpression: &model.UnaryExpression{
								Expression: model.ResultType(types.Long),
								Operand:    model.NewLiteral("2", types.Integer),
							},
						},
					},
					Expression: model.ResultType(types.Long),
				},
			},
			wantResult: newOrFatal(t, int64(-1)),
		},
		{
			name:       "Decimals",
			cql:        "1 - 2.0",
			wantResult: newOrFatal(t, -1.0),
		},
		{
			name:       "Quantity",
			cql:        "10.1 'day' - 1.1 'day'",
			wantResult: newOrFatal(t, result.Quantity{Value: 9, Unit: model.DAYUNIT}),
		},
		// Tests for Date and Quantity
		// TODO(b/301606416): Add more tests for DateTime + Quantity
		{
			name:       "Date Quantity",
			cql:        "@2014 - 1 'year'",
			wantResult: newOrFatal(t, result.Date{Date: time.Date(2013, time.January, 1, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.YEAR}),
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
				t.Errorf("Parse  diff (-want +got):\n%s", diff)
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

func TestMultiply(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name:       "Integers",
			cql:        "2 * 4",
			wantResult: newOrFatal(t, 8),
		},
		{
			name:       "Longs",
			cql:        "2 * 3L",
			wantResult: newOrFatal(t, int64(6)),
		},
		{
			name: "Decimals",
			cql:  "2L * 2.0",
			wantModel: &model.Multiply{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						&model.ToDecimal{
							UnaryExpression: &model.UnaryExpression{
								Expression: model.ResultType(types.Decimal),
								Operand:    model.NewLiteral("2L", types.Long),
							},
						},
						model.NewLiteral("2.0", types.Decimal),
					},
					Expression: model.ResultType(types.Decimal),
				},
			},
			wantResult: newOrFatal(t, 4.0),
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

func TestRound(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name:       "Simple",
			cql:        "Round(42.101)",
			wantResult: newOrFatal(t, 42.0),
		},
		{
			name:       "Negative decimal",
			cql:        "Round(-101.42)",
			wantResult: newOrFatal(t, -101.0),
		},
		{
			name: "Integers",
			cql:  "Round(2)",
			wantModel: &model.Round{
				NaryExpression: &model.NaryExpression{
					Operands: []model.IExpression{
						&model.ToDecimal{
							UnaryExpression: &model.UnaryExpression{
								Operand:    model.NewLiteral("2", types.Integer),
								Expression: model.ResultType(types.Decimal),
							},
						},
					},
					Expression: model.ResultType(types.Decimal),
				},
			},
			wantResult: newOrFatal(t, 2.0),
		},
		{
			name:       "Negative, round up",
			cql:        "Round(-0.5)",
			wantResult: newOrFatal(t, 0.0),
		},
		{
			name:       "Negative, round down",
			cql:        "Round(-0.6)",
			wantResult: newOrFatal(t, -1.0),
		},
		{
			name:       "Zero",
			cql:        "Round(0.0)",
			wantResult: newOrFatal(t, 0.0),
		},
		{
			name:       "Null",
			cql:        "Round(null as Decimal)",
			wantResult: newOrFatal(t, nil),
		},
		// With precision
		{
			name:       "Simple with precision",
			cql:        "Round(42.101, 1)",
			wantResult: newOrFatal(t, 42.1),
		},
		{
			name:       "Negative decimal with precision round up",
			cql:        "Round(-101.45, 1)",
			wantResult: newOrFatal(t, -101.4),
		},
		{
			name:       "Negative decimal with precision round down",
			cql:        "Round(-101.46, 1)",
			wantResult: newOrFatal(t, -101.5),
		},
		{
			name:       "Precision is 0",
			cql:        "Round(2.123, 0)",
			wantResult: newOrFatal(t, 2.0),
		},
		{
			name:       "Precision is null",
			cql:        "Round(2.123, null)",
			wantResult: newOrFatal(t, 2.0),
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

func TestRound_EvalErrors(t *testing.T) {
	tests := []struct {
		name                string
		cql                 string
		wantModel           model.IExpression
		wantEvalErrContains string
	}{
		{
			name:                "Round with a negative precision",
			cql:                 "Round(2.123, -1)",
			wantEvalErrContains: "precision must be non-negative",
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

func TestTruncate(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name:       "Simple",
			cql:        "Truncate(42.101)",
			wantResult: newOrFatal(t, 42),
		},
		{
			name:       "Negative decimal",
			cql:        "Truncate(-101.42)",
			wantResult: newOrFatal(t, -101),
		},
		{
			name: "Integers",
			cql:  "Truncate(2)",
			wantModel: &model.Truncate{
				UnaryExpression: &model.UnaryExpression{
					Operand: &model.ToDecimal{
						UnaryExpression: &model.UnaryExpression{
							Operand:    model.NewLiteral("2", types.Integer),
							Expression: model.ResultType(types.Decimal),
						},
					},
					Expression: model.ResultType(types.Integer),
				},
			},
			wantResult: newOrFatal(t, 2),
		},
		{
			name:       "Zero",
			cql:        "Truncate(0.0)",
			wantResult: newOrFatal(t, 0),
		},
		{
			name:       "Null",
			cql:        "Truncate(null as Decimal)",
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

func TestTruncatedDivide(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Integers",
			cql:  "2 div 4",
			wantModel: &model.TruncatedDivide{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("2", types.Integer),
						model.NewLiteral("4", types.Integer),
					},
					Expression: model.ResultType(types.Integer),
				},
			},
			wantResult: newOrFatal(t, 0),
		},
		{
			name:       "Longs",
			cql:        "2 div 4L",
			wantResult: newOrFatal(t, int64(0)),
		},
		{
			name:       "Decimals",
			cql:        "5 div 2.0",
			wantResult: newOrFatal(t, 2.0),
		},
		{
			name:       "Quantity",
			cql:        "10.1 day div 1.1 day",
			wantResult: newOrFatal(t, result.Quantity{Value: 9.0, Unit: model.ONEUNIT}),
		},
		{
			name:       "Divide by zero",
			cql:        "10 div 0",
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

func TestDivide(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Integers",
			cql:  "2 / 4",
			wantModel: &model.Divide{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						&model.ToDecimal{
							UnaryExpression: &model.UnaryExpression{
								Expression: model.ResultType(types.Decimal),
								Operand:    model.NewLiteral("2", types.Integer),
							},
						},
						&model.ToDecimal{
							UnaryExpression: &model.UnaryExpression{
								Expression: model.ResultType(types.Decimal),
								Operand:    model.NewLiteral("4", types.Integer),
							},
						},
					},
					Expression: model.ResultType(types.Decimal),
				},
			},
			wantResult: newOrFatal(t, 0.5),
		},
		{
			name:       "Longs",
			cql:        "2 / 4L",
			wantResult: newOrFatal(t, 0.5),
		},
		{
			name:       "Decimals",
			cql:        "5 / 2.0",
			wantResult: newOrFatal(t, 2.5),
		},
		{
			name:       "Quantity",
			cql:        "5.0 day / 2.0 day",
			wantResult: newOrFatal(t, result.Quantity{Value: 2.5, Unit: model.ONEUNIT}),
		},
		{
			name:       "Divide by zero",
			cql:        "10 / 0",
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

func TestMod(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name:       "Integers Zero",
			cql:        "1000 mod 200",
			wantResult: newOrFatal(t, 0),
		},
		{
			name: "Integers",
			cql:  "5 mod 2",
			wantModel: &model.Modulo{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("5", types.Integer),
						model.NewLiteral("2", types.Integer),
					},
					Expression: model.ResultType(types.Integer),
				},
			},
			wantResult: newOrFatal(t, 1),
		},
		{
			name:       "Longs Zero",
			cql:        "1000L mod 200L",
			wantResult: newOrFatal(t, int64(0)),
		},
		{
			name:       "Longs",
			cql:        "5L mod 2L",
			wantResult: newOrFatal(t, int64(1)),
		},
		{
			name:       "Decimals",
			cql:        "10.1111 mod 2.1111",
			wantResult: newOrFatal(t, 1.6667000000000005),
		},
		{
			name:       "Another Decimals",
			cql:        "2.1111 mod 10.1111",
			wantResult: newOrFatal(t, 2.1111),
		},
		{
			name:       "Quantity",
			cql:        "10 'm' mod 2 'm'",
			wantResult: newOrFatal(t, result.Quantity{Value: 0, Unit: "m"}),
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

func TestPower(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name:       "Integers Zero",
			cql:        "0 ^ 0",
			wantResult: newOrFatal(t, 1),
		},
		{
			name: "Integers",
			cql:  "5 ^ 2",
			wantModel: &model.Power{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("5", types.Integer),
						model.NewLiteral("2", types.Integer),
					},
					Expression: model.ResultType(types.Integer),
				},
			},
			wantResult: newOrFatal(t, 25),
		},
		{
			name:       "Left arg negative",
			cql:        "-5 ^ 2",
			wantResult: newOrFatal(t, 25),
		},
		{
			name:       "Right arg negative",
			cql:        "5 ^ -2",
			wantResult: newOrFatal(t, 1.0/25.0),
		},
		{
			name:       "Right arg negative decimal",
			cql:        "25 ^ -0.5",
			wantResult: newOrFatal(t, .2),
		},
		{
			name:       "Right arg negative long",
			cql:        "5 ^ -2L",
			wantResult: newOrFatal(t, 1.0/25.0),
		},
		{
			name:       "Longs Zero",
			cql:        "2L ^ 2L",
			wantResult: newOrFatal(t, int64(4)),
		},
		{
			name:       "Longs",
			cql:        "5L ^ 2L",
			wantResult: newOrFatal(t, int64(25)),
		},
		{
			name:       "Decimals",
			cql:        "2.5 ^ 2.0",
			wantResult: newOrFatal(t, 6.25),
		},
		{
			name:       "Functional Syntax",
			cql:        "Power(2, 2)",
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

func TestMaximum(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name:       "maximum Integer",
			cql:        "maximum Integer",
			wantModel:  &model.MaxValue{ValueType: types.Integer, Expression: model.ResultType(types.Integer)},
			wantResult: newOrFatal(t, int32(2147483647)),
		},
		{
			name:       "maximum Long",
			cql:        "maximum Long",
			wantResult: newOrFatal(t, int64(9223372036854775807)),
		},
		{
			name:       "maximum Decimal",
			cql:        "maximum Decimal",
			wantResult: newOrFatal(t, float64(99999999999999999999.99999999)),
		},
		{
			name:       "maximum Date",
			cql:        "maximum Date",
			wantResult: newOrFatal(t, result.Date{Date: time.Date(9999, 12, 31, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.DAY}),
		},
		{
			name:       "maximum DateTime",
			cql:        "maximum DateTime",
			wantResult: newOrFatal(t, result.DateTime{Date: time.Date(9999, 12, 31, 23, 59, 59, 999, defaultEvalTimestamp.Location()), Precision: model.MILLISECOND}),
		},
		{
			name:       "maximum Time",
			cql:        "maximum Time",
			wantResult: newOrFatal(t, result.Time{Date: time.Date(0, time.January, 1, 23, 59, 59, 999000000, defaultEvalTimestamp.Location()), Precision: model.MILLISECOND}),
		},
		{
			name:       "maximum Quantity",
			cql:        "maximum Quantity",
			wantResult: newOrFatal(t, result.Quantity{Value: float64(99999999999999999999.99999999), Unit: "1"}),
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

func TestMinimum(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name:       "minimum Integer",
			cql:        "minimum Integer",
			wantModel:  &model.MinValue{ValueType: types.Integer, Expression: model.ResultType(types.Integer)},
			wantResult: newOrFatal(t, int32(-2147483648)),
		},
		{
			name:       "minimum Long",
			cql:        "minimum Long",
			wantResult: newOrFatal(t, int64(-9223372036854775808)),
		},
		{
			name:       "minimum Decimal",
			cql:        "minimum Decimal",
			wantResult: newOrFatal(t, float64(-99999999999999999999.99999999)),
		},
		{
			name:       "minimum Date",
			cql:        "minimum Date",
			wantResult: newOrFatal(t, result.Date{Date: time.Date(1, 1, 1, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.DAY}),
		},
		{
			name:       "minimum DateTime",
			cql:        "minimum DateTime",
			wantResult: newOrFatal(t, result.DateTime{Date: time.Date(1, 1, 1, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.MILLISECOND}),
		},
		{
			name:       "minimum Time",
			cql:        "minimum Time",
			wantResult: newOrFatal(t, result.Time{Date: time.Date(0, time.January, 1, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.MILLISECOND}),
		},
		{
			name:       "minimum Quantity",
			cql:        "minimum Quantity",
			wantResult: newOrFatal(t, result.Quantity{Value: float64(-99999999999999999999.99999999), Unit: "1"}),
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

func TestPredecessor(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "predecessor of 1",
			cql:  "predecessor of 1",
			wantModel: &model.Predecessor{
				UnaryExpression: &model.UnaryExpression{
					Operand:    model.NewLiteral("1", types.Integer),
					Expression: model.ResultType(types.Integer),
				},
			},
			wantResult: newOrFatal(t, 0),
		},
		{
			name:       "predecessor of 1000000000000L",
			cql:        "predecessor of 1000000000000L",
			wantResult: newOrFatal(t, int64(999999999999)),
		},
		{
			name:       "predecessor of 1.0",
			cql:        "predecessor of 1.0",
			wantResult: newOrFatal(t, float64(0.99999999)),
		},
		{
			name:       "predecessor of @2024-01-02",
			cql:        "predecessor of @2024-01-02",
			wantResult: newOrFatal(t, result.Date{Date: time.Date(2024, 1, 1, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.DAY}),
		},
		{
			name:       "predecessor of @2024-01-01T00:00:00.001Z",
			cql:        "predecessor of @2024-01-01T00:00:00.001Z",
			wantResult: newOrFatal(t, result.DateTime{Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Precision: model.MILLISECOND}),
		},
		{
			name:       "predecessor of @T12:00:00.000",
			cql:        "predecessor of @T12:00:00.000",
			wantResult: newOrFatal(t, result.Time{Date: time.Date(0, time.January, 1, 11, 59, 59, 999000000, defaultEvalTimestamp.Location()), Precision: model.MILLISECOND}),
		},
		{
			name:       "predecessor of 1.0'cm'",
			cql:        "predecessor of 1.0'cm'",
			wantResult: newOrFatal(t, result.Quantity{Value: float64(0.99999999), Unit: "cm"}),
		},
		{
			name:       "predecessor of (4 as Choice<Integer, String>)",
			cql:        "predecessor of (4 as Choice<Integer, String>)",
			wantResult: newOrFatal(t, int32(3)),
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

func TestSuccessor(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "successor of 1",
			cql:  "successor of 1",
			wantModel: &model.Successor{
				UnaryExpression: &model.UnaryExpression{
					Operand:    model.NewLiteral("1", types.Integer),
					Expression: model.ResultType(types.Integer),
				},
			},
			wantResult: newOrFatal(t, 2),
		},
		{
			name:       "successor of 1000000000000L",
			cql:        "successor of 1000000000000L",
			wantResult: newOrFatal(t, int64(1000000000001)),
		},
		{
			name:       "successor of 1.0",
			cql:        "successor of 1.0",
			wantResult: newOrFatal(t, float64(1.00000001)),
		},
		{
			name:       "successor of @2024-01-01",
			cql:        "successor of @2024-01-01",
			wantResult: newOrFatal(t, result.Date{Date: time.Date(2024, 1, 2, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.DAY}),
		},
		{
			name:       "successor of @2024-01-01T00:00:00.999Z",
			cql:        "successor of @2024-01-01T00:00:00.999Z",
			wantResult: newOrFatal(t, result.DateTime{Date: time.Date(2024, 1, 1, 0, 0, 1, 0, time.UTC), Precision: model.MILLISECOND}),
		},
		{
			name:       "successor of @T11:59:59.999",
			cql:        "successor of @T11:59:59.999",
			wantResult: newOrFatal(t, result.Time{Date: time.Date(0, time.January, 1, 12, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.MILLISECOND}),
		},
		{
			name:       "successor of 1.0'cm'",
			cql:        "successor of 1.0'cm'",
			wantResult: newOrFatal(t, result.Quantity{Value: float64(1.00000001), Unit: "cm"}),
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

func TestSuccessorPredecessor_EvalErrors(t *testing.T) {
	tests := []struct {
		name                string
		cql                 string
		wantModel           model.IExpression
		wantEvalErrContains string
	}{
		{
			name:                "successor of max value",
			cql:                 "successor of maximum Integer",
			wantEvalErrContains: "tried to compute successor for value that is already a max",
		},
		// Date tests
		{
			name:                "successor of maximum Date",
			cql:                 "successor of maximum Date",
			wantEvalErrContains: "tried to compute successor for value that is already a max",
		},
		{
			name:                "successor of max date for year precision",
			cql:                 "successor of @9999",
			wantEvalErrContains: "tried to compute successor for System.Date that is already a max",
		},
		{
			name:                "predecessor of minimum Date",
			cql:                 "predecessor of minimum Date",
			wantEvalErrContains: "tried to compute predecessor for value that is already a min",
		},
		{
			name:                "predecessor of min date for year precision",
			cql:                 "predecessor of @0001",
			wantEvalErrContains: "tried to compute predecessor for System.Date that is already a min",
		},
		// DateTime tests
		{
			name:                "successor of maximum DateTime",
			cql:                 "successor of maximum DateTime",
			wantEvalErrContains: "tried to compute successor for value that is already a max",
		},
		{
			name:                "successor of max date for day precision",
			cql:                 "successor of @9999-12-31T",
			wantEvalErrContains: "tried to compute successor for System.DateTime that is already a max",
		},
		{
			name:                "predecessor of minimum DateTime",
			cql:                 "predecessor of minimum DateTime",
			wantEvalErrContains: "tried to compute predecessor for value that is already a min",
		},
		{
			name:                "predecessor of min date for day precision",
			cql:                 "predecessor of @0001-01-01T",
			wantEvalErrContains: "tried to compute predecessor for System.DateTime that is already a min",
		},
		// Time tests
		{
			name:                "successor of maximum Time",
			cql:                 "successor of maximum Time",
			wantEvalErrContains: "tried to compute successor for value that is already a max",
		},
		{
			name:                "successor of max time for minute precision",
			cql:                 "successor of @T23:59",
			wantEvalErrContains: "tried to compute successor for System.Time that is already a max",
		},
		{
			name:                "predecessor of minimum Time",
			cql:                 "predecessor of minimum Time",
			wantEvalErrContains: "tried to compute predecessor for value that is already a min",
		},
		{
			name:                "predecessor of min time for minute precision",
			cql:                 "predecessor of @T00:00",
			wantEvalErrContains: "tried to compute predecessor for System.Time that is already a min",
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

func TestNegate(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Integer",
			cql:  "-4",
			wantModel: &model.Negate{
				UnaryExpression: &model.UnaryExpression{
					Operand:    model.NewLiteral("4", types.Integer),
					Expression: model.ResultType(types.Integer),
				},
			},
			wantResult: newOrFatal(t, -4),
		},
		{
			name:       "Minimum Integer",
			cql:        "-2147483648",
			wantResult: newOrFatal(t, int32(math.MinInt32)),
		},
		{
			name:       "Long",
			cql:        "-4L",
			wantResult: newOrFatal(t, int64(-4)),
		},
		{
			name:       "Minimum Long",
			cql:        "-9223372036854775808L",
			wantResult: newOrFatal(t, int64(math.MinInt64)),
		},
		{
			name:       "Decimal",
			cql:        "-1.0",
			wantResult: newOrFatal(t, float64(-1.0)),
		},
		{
			name:       "Minimum Decimal",
			cql:        "-99999999999999999999.99999999",
			wantResult: newOrFatal(t, float64(-99999999999999999999.99999999)),
		},
		{
			name:       "Quantity",
			cql:        "-1.0 'day'",
			wantResult: newOrFatal(t, result.Quantity{Value: -1.0, Unit: model.DAYUNIT}),
		},
		{
			name:       "Null",
			cql:        "-(null as Integer)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "negate minimum Integer",
			cql:        "Negate(minimum Integer)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "negate minimum Long",
			cql:        "Negate(minimum Long)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "negate minimum Decimal",
			cql:        "Negate(minimum Decimal)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "negate minimum Quantity",
			cql:        "Negate(minimum Quantity)",
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
