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

func TestCanConvertQuantity(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "CanConvertQuantity(1 year, null) returns null",
			cql:  "CanConvertQuantity(1 year, null)",
			wantModel: &model.CanConvertQuantity{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						&model.Quantity{
							Value:      1,
							Unit:       "year",
							Expression: model.ResultType(types.Quantity),
						},
						&model.As{
							UnaryExpression: &model.UnaryExpression{
								Operand:    model.NewLiteral("null", types.Any),
								Expression: model.ResultType(types.String),
							},
							AsTypeSpecifier: types.String,
						},
					},
					Expression: model.ResultType(types.Boolean),
				},
			},
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "CanConvertQuantity(null, 'y') returns null",
			cql:        "CanConvertQuantity(null, 'y')",
			wantResult: newOrFatal(t, nil),
		},
		// CanConvertQuantity is unsupported, should return false for all cases.
		{
			name:       "CanConvertQuantity(1 year, 'mo')",
			cql:        "CanConvertQuantity(1 year, 'mo')",
			wantResult: newOrFatal(t, false),
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

func TestDateTimeOperatorBefore(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "True",
			cql:  "@2020-03-01 before day of @2020-03-02",
			wantModel: &model.Before{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("@2020-03-01", types.Date),
						model.NewLiteral("@2020-03-02", types.Date),
					},
					Expression: model.ResultType(types.Boolean),
				},
				Precision: model.DAY,
			},
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "False",
			cql:        "@2020-03-04 before day of @2020-03-02",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Equal dates false",
			cql:        "@2020-03-01 before day of @2020-03-01",
			wantResult: newOrFatal(t, false),
		},
		{
			name: "Left null",
			// Ambiguous match without casting.
			cql:        "(null as Date) before day of @2020-03-02",
			wantResult: newOrFatal(t, nil),
		},
		{
			name: "Right null",
			// Ambiguous match without casting.
			cql:        "@2020 before day of (null as Date)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "No precision",
			cql:        "@2014-01-01 before @2015-01-01",
			wantResult: newOrFatal(t, true),
		},
		// DateTime tests
		{
			name:       "Equal datetimes false",
			cql:        "@2024-02-29T01:20:30.101-07:00 before day of @2024-02-29T01:20:30.101-07:00",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Same Year Different Date at Year Precision is False",
			cql:        "@2024-02-29T01:20:30.101-07:00 before year of @2024-02-28T01:20:30.101-07:00",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Month Precision Before True Even Though Day Precision is Before",
			cql:        "@2024-02-29T01:20:30.101-07:00 before month of @2024-03-31T01:20:30.101-07:00",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Timezone offset taken into account",
			cql:        "@2024-02-19T01:20-04:00 before hour of @2024-02-19T01:20-07:00",
			wantResult: newOrFatal(t, true),
		},
		{
			name: "Timezone not applied for day precision",
			// If timezones were normalized this would be @2024-02-19TZ before day of @2024-02-19TZ
			// which would be false.
			cql:        "@2024-02-18T-04:00 before day of @2024-02-19TZ",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Timezone applied because the left DateTime is hour precision",
			cql:        "@2024-02-18T23-04:00 before day of @2024-02-19TZ",
			wantResult: newOrFatal(t, false),
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

func TestDateTimeOperatorBefore_Error(t *testing.T) {
	tests := []struct {
		name                string
		cql                 string
		wantModel           model.IExpression
		wantEvalErrContains string
	}{
		{
			name: "Invalid Precision Date",
			cql:  "@2020-03-01 before second of @2020-03-02",
			wantModel: &model.Before{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("@2020-03-01", types.Date),
						model.NewLiteral("@2020-03-02", types.Date),
					},
					Expression: model.ResultType(types.Boolean),
				},
				Precision: model.SECOND,
			},
			wantEvalErrContains: "precision must be one of",
		},
		{
			name:                "Invalid Precision DateTime",
			cql:                 "@2024-02-29T01:20:30.101-07:00 before week of @2024-03-31T01:20:30.101-07:00",
			wantEvalErrContains: "precision must be one of",
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
				t.Fatalf("Evaluate Expression expected an error to be returned, got nil instead for cql: %s", tc.cql)
			}
			if !strings.Contains(err.Error(), tc.wantEvalErrContains) {
				t.Errorf("Unexpected evaluation error contents. got: %v, want contains: %v", err.Error(), tc.wantEvalErrContains)
			}
		})
	}
}

func TestDateTimeOperatorAfter(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "@2020-03-04 after day of @2020-03-02",
			cql:  "@2020-03-04 after day of @2020-03-02",
			wantModel: &model.After{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("@2020-03-04", types.Date),
						model.NewLiteral("@2020-03-02", types.Date),
					},
					Expression: model.ResultType(types.Boolean),
				},
				Precision: model.DAY,
			},
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "@2020-03-01 after day of @2020-03-02 returns false",
			cql:        "@2020-03-01 after day of @2020-03-02",
			wantResult: newOrFatal(t, false),
		},
		{
			name: "null after day of @2020-03-02 returns null",
			// Ambiguous match without casting.
			cql:        "(null as Date) after day of @2020-03-02",
			wantResult: newOrFatal(t, nil),
		},
		{
			name: "@2020 after day of null returns null",
			// Ambiguous match without casting.
			cql:        "@2020 after day of (null as Date)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "@2020-03-01 after day of @2020-03-01 returns false",
			cql:        "@2020-03-01 after day of @2020-03-01",
			wantResult: newOrFatal(t, false),
		},
		// DateTime tests
		{
			name:       "@2024-02-29T01:20:30.101-07:00 after day of @2024-02-29T01:20:30.101-07:00 returns false",
			cql:        "@2024-02-29T01:20:30.101-07:00 after day of @2024-02-29T01:20:30.101-07:00",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Same Year Different Date at Year Precision is False",
			cql:        "@2024-05-29T01:20:30.101-07:00 after year of @2024-03-31T01:20:30.101-07:00",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Month Precision After True Even Though Day Precision is Before",
			cql:        "@2024-05-29T01:20:30.101-07:00 after month of @2024-03-31T01:20:30.101-07:00",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Timezone offset taken into account",
			cql:        "@2024-02-19T01:20+04:00 after hour of @2024-02-19T01:20+07:00",
			wantResult: newOrFatal(t, true),
		},
		{
			name: "Timezone not applied for day precision",
			// If timezones were normalized this would be @2024-02-17TZ after day of @2024-02-17TZ which
			// would be false.
			cql:        "@2024-02-18T+04:00 after day of @2024-02-17TZ",
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
			gotResult := getTESTRESULT(t, results)
			if diff := cmp.Diff(tc.wantResult, gotResult, protocmp.Transform()); diff != "" {
				t.Errorf("Eval diff (-want +got)\n%v", diff)
			}

		})
	}
}

func TestDateTimeOperatorDifferenceBetween(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "difference in years between @2020 and @2022",
			cql:  "difference in years between @2020 and @2022",
			wantModel: &model.DifferenceBetween{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("@2020", types.Date),
						model.NewLiteral("@2022", types.Date),
					},
					Expression: model.ResultType(types.Integer),
				},
				Precision: model.YEAR,
			},
			wantResult: newOrFatal(t, 2),
		},
		{
			name:       "difference in years between @2022 and @2020",
			cql:        "difference in years between @2022 and @2020",
			wantResult: newOrFatal(t, -2),
		},
		{
			name:       "difference in years between null and @2022 returns null",
			cql:        "difference in years between null  and @2022",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "difference in years between @2020 and null returns null",
			cql:        "difference in years between @2020 and null",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "difference in months between @2020-10 and @2022-02",
			cql:        "difference in months between @2020-10 and @2022-02",
			wantResult: newOrFatal(t, 16),
		},
		{
			name:       "difference in weeks from monday to saturday returns zero",
			cql:        "difference in weeks between @2023-11-20 and @2023-11-25",
			wantResult: newOrFatal(t, 0),
		},
		{
			name:       "difference in weeks from monday to sunday returns 1",
			cql:        "difference in weeks between @2023-11-20 and @2023-11-26",
			wantResult: newOrFatal(t, 1),
		},
		{
			name:       "difference in weeks from saturday to saturday returns 1",
			cql:        "difference in weeks between @2023-11-25 and @2023-12-02",
			wantResult: newOrFatal(t, 1),
		},
		{
			name:       "difference in weeks from saturday to saturday a week later returns 2",
			cql:        "difference in weeks between @2023-11-25 and @2023-12-03",
			wantResult: newOrFatal(t, 2),
		},
		{
			name:       "difference in days from saturday to monday returns 2",
			cql:        "difference in days between @2023-11-25 and @2023-11-27",
			wantResult: newOrFatal(t, 2),
		},
		// DateTime tests
		{
			name:       "difference in years between @2022-02-22T01:20:30.101-07:00 and @2024-02-22T01:20:30.101-07:00",
			cql:        "difference in years between @2022-02-22T01:20:30.101-07:00 and @2024-02-22T01:20:30.101-07:00",
			wantResult: newOrFatal(t, 2),
		},
		{
			name:       "difference in hours between @2014-01-01T01:01:00.000Z and @2014-01-02T01:01:00.000Z",
			cql:        "difference in hours between @2014-01-01T01:01:00.000Z and @2014-01-02T01:01:00.000Z",
			wantResult: newOrFatal(t, 24),
		},
		{
			name:       "difference in months between @2014-01-01T01:01:00.000Z and @2014-02",
			cql:        "difference in months between @2014-01-01T01:01:00.000Z and @2014-02",
			wantResult: newOrFatal(t, 1),
		},
		{
			name:       "difference in milliseconds between @2022-02-22T01:20:30.101-07:00 and @2022-02-22T01:20:30.105-07:00",
			cql:        "difference in milliseconds between @2022-02-22T01:20:30.101-07:00 and @2022-02-22T01:20:30.105-07:00",
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
			gotResult := getTESTRESULT(t, results)
			if diff := cmp.Diff(tc.wantResult, gotResult, protocmp.Transform()); diff != "" {
				t.Errorf("Eval diff (-want +got)\n%v", diff)
			}
		})
	}
}

func TestDateTimeOperatorDifferencebetween_Error(t *testing.T) {
	tests := []struct {
		name                string
		cql                 string
		wantModel           model.IExpression
		wantEvalErrContains string
	}{
		{
			name: "difference in months between 2014 and 2016 return error invalid precision",
			cql:  "difference in months between @2014 and @2016",
			wantModel: &model.DifferenceBetween{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("@2014", types.Date),
						model.NewLiteral("@2016", types.Date),
					},
					Expression: model.ResultType(types.Integer),
				},
				Precision: model.MONTH,
			},
			wantEvalErrContains: "difference between specified a precision greater than argument precision",
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
				t.Fatalf("Evaluate Expression expected an error to be returned, got nil instead for cql: %s", tc.cql)
			}
			if !strings.Contains(err.Error(), tc.wantEvalErrContains) {
				t.Errorf("Unexpected evaluation error contents. got: %v, want contains: %v", err.Error(), tc.wantEvalErrContains)
			}
		})
	}
}

func TestDateTimeOperatorSameOrAfter(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		// same or after syntax
		{
			name: "@2020-03-04 same day or after @2020-03-02",
			cql:  "@2020-03-04 same day or after @2020-03-02",
			wantModel: &model.SameOrAfter{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("@2020-03-04", types.Date),
						model.NewLiteral("@2020-03-02", types.Date),
					},
					Expression: model.ResultType(types.Boolean),
				},
				Precision: model.DAY,
			},
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "@2020-03-01 same day or after @2020-03-02 returns false",
			cql:        "@2020-03-01 same day or after @2020-03-02",
			wantResult: newOrFatal(t, false),
		},
		// on or after operator
		{
			name: "@2020-03-04 on or after day of @2020-03-02",
			cql:  "@2020-03-04 on or after day of @2020-03-02",
			wantModel: &model.SameOrAfter{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("@2020-03-04", types.Date),
						model.NewLiteral("@2020-03-02", types.Date),
					},
					Expression: model.ResultType(types.Boolean),
				},
				Precision: model.DAY,
			},
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "@2020-03-01 on or after day of @2020-03-02 returns false",
			cql:        "@2020-03-01 on or after day of @2020-03-02",
			wantResult: newOrFatal(t, false),
		},
		{
			name: "null on or after year of @2020 returns null",
			// Ambiguous match without casting.
			cql:        "(null as Date) on or after year of @2020",
			wantResult: newOrFatal(t, nil),
		},
		{
			name: "@2020 on or after year of null returns null",
			// Ambiguous match without casting.
			cql:        "@2020 on or after year of (null as Date)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "@2024-02-29T01:20:30.101-07:00 on or after day of @2024-02-29T01:20:30.101-07:00",
			cql:        "@2024-02-29T01:20:30.101-07:00 on or after day of @2024-02-29T01:20:30.101-07:00",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Same Year Different Date at Year Precision is True",
			cql:        "@2024-03-30T01:20:30.101-07:00 on or after year of @2024-02-28T01:20:30.101-07:00",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Month Precision After True Even Though Day Precision is Before",
			cql:        "@2024-05-29T01:20:30.101-07:00 on or after month of @2024-03-31T01:20:30.101-07:00",
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

func TestDateTimeOperatorSameOrBefore(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		// same or before syntax
		{
			name: "@2020-03-01 same day or before @2020-03-02",
			cql:  "@2020-03-01 same day or before @2020-03-02",
			wantModel: &model.SameOrBefore{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("@2020-03-01", types.Date),
						model.NewLiteral("@2020-03-02", types.Date),
					},
					Expression: model.ResultType(types.Boolean),
				},
				Precision: model.DAY,
			},
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "@2020-03-04 same day or before @2020-03-02 returns false",
			cql:        "@2020-03-04 same day or before @2020-03-02",
			wantResult: newOrFatal(t, false),
		},
		// on or before operator
		{
			name: "@2020-03-01 on or before day of @2020-03-02",
			cql:  "@2020-03-01 on or before day of @2020-03-02",
			wantModel: &model.SameOrBefore{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("@2020-03-01", types.Date),
						model.NewLiteral("@2020-03-02", types.Date),
					},
					Expression: model.ResultType(types.Boolean),
				},
				Precision: model.DAY,
			},
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "@2020-03-04 on or before day of @2020-03-02 returns false",
			cql:        "@2020-03-04 on or before day of @2020-03-02",
			wantResult: newOrFatal(t, false),
		},
		{
			name: "null on or before year of @2020 returns null",
			// Ambiguous match without casting.
			cql:        "(null as Date) on or before year of @2020",
			wantResult: newOrFatal(t, nil),
		},
		{
			name: "@2020 on or before year of null returns null",
			// Ambiguous match without casting.
			cql:        "@2020 on or before year of (null as Date)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "@2024-02-29T01:20:30.101-07:00 on or before day of @2024-02-29T01:20:30.101-07:00",
			cql:        "@2024-02-29T01:20:30.101-07:00 on or before day of @2024-02-29T01:20:30.101-07:00",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Same Year Different Date at Year Precision is True",
			cql:        "@2024-02-29T01:20:30.101-07:00 on or before year of @2024-03-31T01:20:30.101-07:00",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Month Precision Before True Even Though Day Precision is After",
			cql:        "@2024-02-29T01:20:30.101-07:00 on or before month of @2024-03-31T01:20:30.101-07:00",
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

func TestEvaluationTimestamp(t *testing.T) {
	tests := []struct {
		name                string
		cql                 string
		evaluationTimestamp time.Time
		wantModel           model.IExpression
		wantResult          result.Value
	}{
		{
			name:                "Now returns passed evaluation timestamp",
			cql:                 "define TESTRESULT: Now()",
			evaluationTimestamp: time.Date(2024, time.January, 1, 0, 0, 0, 1, time.UTC),
			wantModel: &model.Now{
				NaryExpression: &model.NaryExpression{
					Operands:   []model.IExpression{},
					Expression: model.ResultType(types.DateTime),
				},
			},
			wantResult: newOrFatal(t, result.DateTime{Date: time.Date(2024, time.January, 1, 0, 0, 0, 1, time.UTC), Precision: model.MILLISECOND}),
		},
		{
			name:                "Time returns passed evaluation timestamp time components",
			cql:                 "define TESTRESULT: TimeOfDay()",
			evaluationTimestamp: time.Date(0, time.January, 1, 1, 2, 3, 4, time.UTC),
			wantResult:          newOrFatal(t, result.Time{Date: time.Date(0, time.January, 1, 1, 2, 3, 4, time.UTC), Precision: model.MILLISECOND}),
		},
		{
			name:                "Today truncates time values",
			cql:                 "define TESTRESULT: Today()",
			evaluationTimestamp: time.Date(2024, time.January, 1, 1, 1, 1, 1, time.UTC),
			wantResult:          newOrFatal(t, result.Date{Date: time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC), Precision: model.DAY}),
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
			config.EvaluationTimestamp = tc.evaluationTimestamp
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

func TestDate(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Year",
			cql:  "Date(2014)",
			wantModel: &model.Date{
				NaryExpression: &model.NaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("2014", types.Integer),
					},
					Expression: model.ResultType(types.Date),
				},
			},
			wantResult: newOrFatal(t, result.Date{Date: time.Date(2014, time.January, 1, 0, 0, 0, 0, time.UTC), Precision: model.YEAR}),
		},
		{
			name:       "Month",
			cql:        "Date(2014, 9)",
			wantResult: newOrFatal(t, result.Date{Date: time.Date(2014, time.September, 1, 0, 0, 0, 0, time.UTC), Precision: model.MONTH}),
		},
		{
			name:       "Day",
			cql:        "Date(2014, 9, 4)",
			wantResult: newOrFatal(t, result.Date{Date: time.Date(2014, time.September, 4, 0, 0, 0, 0, time.UTC), Precision: model.DAY}),
		},
		{
			name:       "Functional and string constructors equal",
			cql:        "Date(2014, 9, 4) = @2014-09-04",
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

func TestDateTime(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Year",
			cql:  "DateTime(2014)",
			wantModel: &model.DateTime{
				NaryExpression: &model.NaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("2014", types.Integer),
					},
					Expression: model.ResultType(types.DateTime),
				},
			},
			wantResult: newOrFatal(t, result.DateTime{Date: time.Date(2014, time.January, 1, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.YEAR}),
		},
		{
			name:       "Month",
			cql:        "DateTime(2014, 9)",
			wantResult: newOrFatal(t, result.DateTime{Date: time.Date(2014, time.September, 1, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.MONTH}),
		},
		{
			name:       "Day",
			cql:        "DateTime(2014, 9, 4)",
			wantResult: newOrFatal(t, result.DateTime{Date: time.Date(2014, time.September, 4, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.DAY}),
		},
		{
			name:       "Hour",
			cql:        "DateTime(2014, 9, 4, 12)",
			wantResult: newOrFatal(t, result.DateTime{Date: time.Date(2014, time.September, 4, 12, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.HOUR}),
		},
		{
			name:       "Minute",
			cql:        "DateTime(2014, 9, 4, 12, 30)",
			wantResult: newOrFatal(t, result.DateTime{Date: time.Date(2014, time.September, 4, 12, 30, 0, 0, defaultEvalTimestamp.Location()), Precision: model.MINUTE}),
		},
		{
			name:       "Second",
			cql:        "DateTime(2014, 9, 4, 12, 30, 30)",
			wantResult: newOrFatal(t, result.DateTime{Date: time.Date(2014, time.September, 4, 12, 30, 30, 0, defaultEvalTimestamp.Location()), Precision: model.SECOND}),
		},
		{
			name:       "Millisecond",
			cql:        "DateTime(2014, 9, 4, 12, 30, 30, 100)",
			wantResult: newOrFatal(t, result.DateTime{Date: time.Date(2014, time.September, 4, 12, 30, 30, 100*1000000, defaultEvalTimestamp.Location()), Precision: model.MILLISECOND}),
		},
		{
			name:       "Timezone",
			cql:        "DateTime(2014, 9, 4, 12, 30, 30, 100, -7)",
			wantResult: newOrFatal(t, result.DateTime{Date: time.Date(2014, time.September, 4, 12, 30, 30, 100*1000000, time.FixedZone("-7", -7*60*60)), Precision: model.MILLISECOND}),
		},
		{
			name:       "Functional and string constructors equal",
			cql:        "DateTime(2014, 9, 4, 12, 30, 30, 101) = @2014-09-04T12:30:30.101",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Functional and string constructors equal with timezone",
			cql:        "DateTime(2014, 9, 4, 12, 30, 30, 101, -7) = @2014-09-04T12:30:30.101-07:00",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "All null arguments",
			cql:        "DateTime(null, null, null, null, null, null, null, null)",
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

func TestTime(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Hour",
			cql:  "Time(21)",
			wantModel: &model.Time{
				NaryExpression: &model.NaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("21", types.Integer),
					},
					Expression: model.ResultType(types.Time),
				},
			},
			wantResult: newOrFatal(t, result.Time{Date: time.Date(0, time.January, 1, 21, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.HOUR}),
		},
		{
			name:       "Minute",
			cql:        "Time(12, 30)",
			wantResult: newOrFatal(t, result.Time{Date: time.Date(0, time.January, 1, 12, 30, 0, 0, defaultEvalTimestamp.Location()), Precision: model.MINUTE}),
		},
		{
			name:       "Second",
			cql:        "Time(12, 30, 30)",
			wantResult: newOrFatal(t, result.Time{Date: time.Date(0, time.January, 1, 12, 30, 30, 0, defaultEvalTimestamp.Location()), Precision: model.SECOND}),
		},
		{
			name:       "Millisecond",
			cql:        "Time(12, 30, 30, 100)",
			wantResult: newOrFatal(t, result.Time{Date: time.Date(0, time.January, 1, 12, 30, 30, 100*1000000, defaultEvalTimestamp.Location()), Precision: model.MILLISECOND}),
		},
		{
			name:       "Functional and string constructors equal",
			cql:        "Time(12, 30, 30, 101) = @T12:30:30.101",
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

func TestDateTimeConstructor_Errors(t *testing.T) {
	tests := []struct {
		name      string
		cql       string
		wantModel model.IExpression
		wantErr   string
	}{
		{
			name:    "Date val after null",
			cql:     "Date(2014, null, 15)",
			wantErr: "when constructing Date precision day had value 15, even though a higher precision was null",
		},
		{
			name:    "Date starts with null",
			cql:     "Date(null)",
			wantErr: "in Date year cannot be null",
		},
		{
			name:    "Date year outside range",
			cql:     "Date(-1)",
			wantErr: "year -1 is out of range",
		},
		{
			name:    "Date day outside range",
			cql:     "Date(2014, 5, 32)",
			wantErr: "day 32 is out of range",
		},
		{
			name:    "Time val after null",
			cql:     "Time(12, null, 30)",
			wantErr: "when constructing Time precision second had value 30, even though a higher precision was null",
		},
		{
			name:    "Time starts with null",
			cql:     "Time(null)",
			wantErr: "in Time hour cannot be null",
		},
		{
			name:    "Time hour outside range",
			cql:     "Time(25)",
			wantErr: "hour 25 is out of range",
		},
		{
			name:    "Time millisecond outside range",
			cql:     "Time(12, 30, 30, -100)",
			wantErr: "millisecond -100 is out of range",
		},
		{
			name:    "DateTime year outside range",
			cql:     "DateTime(99999999)",
			wantErr: "year 99999999 is out of range",
		},
		{
			name:    "DateTime millisecond outside range",
			cql:     "DateTime(2014, 9, 4, 12, 30, 30, -101, -7)",
			wantErr: "millisecond -101 is out of range",
		},
		{
			name:    "DateTime timezone our of range",
			cql:     "DateTime(2014, 9, 4, 12, 30, 30, 101, -14.1)",
			wantErr: "timezone offset -14.1 is out of range",
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
				t.Fatal("Evaluate Expression expected an error to be returned, got nil instead")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("Unexpected evaluation error contents. got (%v), want (%v)", err.Error(), tc.wantErr)
			}
		})
	}
}
