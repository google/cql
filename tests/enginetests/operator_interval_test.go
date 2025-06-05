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

func TestEnd(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "High inclusive",
			cql:  "end of Interval[1, 2]",
			wantModel: &model.End{
				UnaryExpression: &model.UnaryExpression{
					Expression: model.ResultType(types.Integer),
					Operand: &model.Interval{
						Low:           model.NewLiteral("1", types.Integer),
						High:          model.NewLiteral("2", types.Integer),
						LowInclusive:  true,
						HighInclusive: true,
						Expression:    model.ResultType(&types.Interval{PointType: types.Integer}),
					},
				},
			},
			wantResult: newOrFatal(t, 2),
		},
		{
			name: "High exclusive null",
			cql:  "end of Interval(5, null)",
			wantModel: &model.End{
				UnaryExpression: &model.UnaryExpression{
					Expression: model.ResultType(types.Integer),
					Operand: &model.Interval{
						Low: model.NewLiteral("5", types.Integer),
						High: &model.As{
							UnaryExpression: &model.UnaryExpression{
								Expression: model.ResultType(types.Integer),
								Operand:    model.NewLiteral("null", types.Any),
							},
							AsTypeSpecifier: types.Integer,
						},
						LowInclusive:  false,
						HighInclusive: false,
						Expression:    model.ResultType(&types.Interval{PointType: types.Integer}),
					},
				},
			},
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Null",
			cql:        "end of null",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "High inclusive null",
			cql:        "end of Interval[2, null]",
			wantResult: newOrFatal(t, int32(2147483647)),
		},
		{
			name:       "High inclusive with all runtime nulls",
			cql:        "end of Interval[null as Integer, null as Integer]",
			wantResult: newOrFatal(t, int32(2147483647)),
		},
		{
			name:       "High exclusive returns predecessor integer",
			cql:        "end of Interval[2, 43)",
			wantResult: newOrFatal(t, int32(42)),
		},
		{
			name:       "High exclusive returns predecessor date",
			cql:        "end of Interval[@2012-01, @2013-01)",
			wantResult: newOrFatal(t, result.Date{Date: time.Date(2012, time.December, 1, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.MONTH}),
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

func TestStart(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Low inclusive",
			cql:  "start of Interval[1, 2]",
			wantModel: &model.Start{
				UnaryExpression: &model.UnaryExpression{
					Expression: model.ResultType(types.Integer),
					Operand: &model.Interval{
						Low:           model.NewLiteral("1", types.Integer),
						High:          model.NewLiteral("2", types.Integer),
						LowInclusive:  true,
						HighInclusive: true,
						Expression:    model.ResultType(&types.Interval{PointType: types.Integer}),
					},
				},
			},
			wantResult: newOrFatal(t, 1),
		},
		{
			name: "Low exclusive null",
			cql:  "start of Interval(null, 2)",
			wantModel: &model.Start{
				UnaryExpression: &model.UnaryExpression{
					Expression: model.ResultType(types.Integer),
					Operand: &model.Interval{
						Low: &model.As{
							UnaryExpression: &model.UnaryExpression{
								Expression: model.ResultType(types.Integer),
								Operand:    model.NewLiteral("null", types.Any),
							},
							AsTypeSpecifier: types.Integer,
						},
						High:          model.NewLiteral("2", types.Integer),
						LowInclusive:  false,
						HighInclusive: false,
						Expression:    model.ResultType(&types.Interval{PointType: types.Integer}),
					},
				},
			},
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Null",
			cql:        "start of null",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Low inclusive null",
			cql:        "start of Interval[null, 2)",
			wantResult: newOrFatal(t, int32(-2147483648)),
		},
		{
			name:       "Low inclusive with all runtime nulls",
			cql:        "start of Interval[null as Integer, null as Integer)",
			wantResult: newOrFatal(t, int32(-2147483648)),
		},
		{
			name:       "Low exclusive returns predecessor integer",
			cql:        "start of Interval(41, 50]",
			wantResult: newOrFatal(t, int32(42)),
		},
		{
			name:       "Low exclusive returns predecessor date",
			cql:        "start of Interval(@2012-11, @2013-01]",
			wantResult: newOrFatal(t, result.Date{Date: time.Date(2012, time.December, 1, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.MONTH}),
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
				t.Errorf("Evaluate diff (-want +got)\n%v", diff)
			}
		})
	}
}

func TestIntervalBefore(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		// Date, Date overloads:
		{
			name: "Date before interval",
			cql:  "@2020-03-04 before day of Interval[@2020-03-05, @2020-03-07]",
			wantModel: &model.Before{
				BinaryExpression: &model.BinaryExpression{
					Expression: model.ResultType(types.Boolean),
					Operands: []model.IExpression{
						model.NewLiteral("@2020-03-04", types.Date),
						&model.Interval{
							Expression:    model.ResultType(&types.Interval{PointType: types.Date}),
							Low:           model.NewLiteral("@2020-03-05", types.Date),
							High:          model.NewLiteral("@2020-03-07", types.Date),
							LowInclusive:  true,
							HighInclusive: true,
						},
					},
				},
				Precision: model.DAY,
			},
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Date before null Interval<Date>",
			cql:        "@2024-02-28 before null as Interval<Date>",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "null Date before Interval<Date>",
			cql:        "null as Date before Interval[@2024-02-29, @2024-03-29]",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Date before interval without precision",
			cql:        "@2020-03-04 before Interval[@2020-03-05, @2020-03-07]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Date day matches interval start",
			cql:        "@2020-03-05 before day of Interval[@2020-03-05, @2020-03-07]",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Date Interval right null",
			cql:        "@2020-03-04 before day of Interval[@2020-03-05, null]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Date Interval left null",
			cql:        "@2020-03-04 before day of Interval[null, @2020-03-07]",
			wantResult: newOrFatal(t, false),
		},
		// DateTime, DateTime overloads:
		{
			name:       "DateTime before null Interval<DateTime>",
			cql:        "@2024-02-28T01:20:30.101-07:00 before null as Interval<DateTime>",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "null DateTime before Interval<DateTime>",
			cql:        "null as DateTime before Interval[@2024-02-29T01:20:30.101-07:00, @2024-03-29T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "DateTime before interval",
			cql:        "@2024-02-28T01:20:30.101-07:00 before day of Interval[@2024-02-29T01:20:30.101-07:00, @2024-03-29T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "DateTime before interval without precision",
			cql:        "@2024-02-28T01:20:30.101-07:00 before Interval[@2024-02-29T01:20:30.101-07:00, @2024-03-29T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "DateTime matches closed interval start",
			cql:        "@2024-02-29T01:20:30.101-07:00 before day of Interval[@2024-02-29T01:20:30.101-07:00, @2024-03-29T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "DateTime Interval right null",
			cql:        "@2020-03-04T before day of Interval[@2020-03-05T, null]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "DateTime Interval left null",
			cql:        "@2020-03-04T before day of Interval[null, @2020-03-07T]",
			wantResult: newOrFatal(t, false),
		},
		// Interval<DateTime>, Interval<DateTime> overloads:
		{
			name:       "Interval<DateTime> before null as Interval<DateTime>",
			cql:        "Interval[@2024-01-25T01:20:30.101-07:00, @2024-01-29T01:20:30.101-07:00] before null as Interval<DateTime>",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "null as Interval<DateTime> before Interval<DateTime>",
			cql:        "null as Interval<DateTime> before Interval[@2024-01-25T01:20:30.101-07:00, @2024-01-29T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Interval<DateTime> before Interval<DateTime> at seconds precision",
			cql:        "Interval[@2024-01-25T01:20:30.101-07:00, @2024-02-29T01:20:10.101-07:00] before second of Interval[@2024-02-29T01:20:30.101-07:00, @2024-03-29T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Interval<DateTime> before Interval<DateTime> without precision",
			cql:        "Interval[@2024-01-25T01:20:30.101-07:00, @2024-01-29T01:20:30.101-07:00] before Interval[@2024-02-29T01:20:30.101-07:00, @2024-03-29T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Left Interval<DateTime> end matches closed right interval start",
			cql:        "Interval[@2024-01-25T01:20:30.101-07:00, @2024-02-29T01:20:30.101-07:00] before Interval[@2024-02-29T01:20:30.101-07:00, @2024-03-29T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Left Interval<DateTime> end matches open right interval start",
			cql:        "Interval[@2024-01-25T01:20:30.101-07:00, @2024-02-29T01:20:30.101-07:00] before Interval(@2024-02-29T01:20:30.101-07:00, @2024-03-29T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Left Interval<DateTime> before right Interval with null end",
			cql:        "Interval[@2024-01-25T01:20:30.101-07:00, @2024-02-28T01:20:30.101-07:00] before Interval[@2024-02-29T01:20:30.101-07:00, null]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Left Interval<DateTime> before right Interval with null start",
			cql:        "Interval[@2024-01-25T01:20:30.101-07:00, @2024-02-28T01:20:30.101-07:00] before Interval[null, @2024-02-29T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, false),
		},
		// Sanity check some Interval<Date>, Interval<Date> overloads
		{
			name:       "Interval<Date> before Interval<Date> at day precision",
			cql:        "Interval[@2024-01-25, @2024-02-28] before day of Interval[@2024-02-29, @2024-03-29]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Interval<Date> before Interval<Date> without precision",
			cql:        "Interval[@2024-01-25, @2024-02-28] before Interval[@2024-02-29, @2024-03-29]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Left Interval<DateTime> end matches closed right interval start",
			cql:        "Interval[@2024-01-25, @2024-02-29] before Interval[@2024-02-29, @2024-03-29]",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Left Interval<DateTime> end matches open right interval start",
			cql:        "Interval[@2024-01-25, @2024-02-29] before Interval(@2024-02-29, @2024-03-29]",
			wantResult: newOrFatal(t, true),
		},
		// starts or ends before syntax:
		{
			name:       "Interval<Date> starts before another",
			cql:        "Interval[@2024, @2026] starts before Interval[@2027, @2027]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Interval<Date> does not start before another",
			cql:        "Interval[@2024, @2026] starts before Interval[@2024, @2027]",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "null Interval<Date> starts before Interval<Date>",
			cql:        "null as Interval<Date> starts before Interval[@2024, @2027]",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "null low bound starts before Interval<Date>",
			cql:        "Interval[null, @2027] starts before Interval[@2024, @2027]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Interval<Date> ends after another",
			cql:        "Interval[@2024, @2028] ends after Interval[@2027, @2027]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Interval<Date> does not end after another",
			cql:        "Interval[@2024, @2026] ends after Interval[@2024, @2027]",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "null Interval<Date> ends after Interval<Date>",
			cql:        "null as Interval<Date> ends after Interval[@2024, @2027]",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "null high bound ends after Interval<Date>",
			cql:        "Interval[@2024, null] ends after Interval[@2024, @2027]",
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

func TestIntervalAfter(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		// Date, Date overloads:
		{
			name: "Day precision true",
			cql:  "@2020-03-09 after day of Interval[@2020-03-05, @2020-03-07]",
			wantModel: &model.After{
				BinaryExpression: &model.BinaryExpression{
					Expression: model.ResultType(types.Boolean),
					Operands: []model.IExpression{
						model.NewLiteral("@2020-03-09", types.Date),
						&model.Interval{
							Expression:    model.ResultType(&types.Interval{PointType: types.Date}),
							Low:           model.NewLiteral("@2020-03-05", types.Date),
							High:          model.NewLiteral("@2020-03-07", types.Date),
							LowInclusive:  true,
							HighInclusive: true,
						},
					},
				},
				Precision: model.DAY,
			},
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Date after null Interval<Date>",
			cql:        "@2024-02-28 after null as Interval<Date>",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "null Date after Interval<Date>",
			cql:        "null as Date after Interval[@2024-02-29, @2024-03-29]",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Same day is false",
			cql:        "@2020-03-07 after day of Interval[@2020-03-05, @2020-03-07]",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Interval right null",
			cql:        "@2020-03-09 after day of Interval[@2020-03-05, null]",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Interval left null",
			cql:        "@2020-03-09 after day of Interval[null, @2020-03-07]",
			wantResult: newOrFatal(t, true),
		},
		// DateTime, DateTime overloads:
		{
			name:       "DateTime after null Interval<DateTime>",
			cql:        "@2024-02-28T01:20:30.101-07:00 after null as Interval<DateTime>",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "null DateTime after Interval<DateTime>",
			cql:        "null as DateTime after Interval[@2024-02-29T01:20:30.101-07:00, @2024-03-29T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "DateTime after interval",
			cql:        "@2024-04-28T01:20:30.101-07:00 after day of Interval[@2024-02-28T01:20:30.101-07:00, @2024-02-29T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "DateTime after interval without precision",
			cql:        "@2024-04-28T01:20:30.101-07:00 after Interval[@2024-02-29T01:20:30.101-07:00, @2024-03-29T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "DateTime equals end of closed interval",
			cql:        "@2024-03-28T01:20:30.101-07:00 after day of Interval[@2024-02-28T01:20:30.101-07:00, @2024-03-28T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "DateTime Interval right null",
			cql:        "@2024-02-28T01:20:30.101-07:00 after day of Interval[@2024-01-28T01:20:30.101-07:00, null]",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "DateTime after Interval with left null",
			cql:        "@2024-02-29T01:20:30.101-07:00 after day of Interval[null, @2024-02-28T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, true),
		},
		// Interval<DateTime>, Interval<DateTime> overloads:
		{
			name:       "Interval<DateTime> after null as Interval<DateTime>",
			cql:        "Interval[@2024-01-25T01:20:30.101-07:00, @2024-01-29T01:20:30.101-07:00] after null as Interval<DateTime>",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "null as Interval<DateTime> after Interval<DateTime>",
			cql:        "null as Interval<DateTime> after Interval[@2024-01-25T01:20:30.101-07:00, @2024-01-29T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Interval<DateTime> after Interval<DateTime> at seconds precision",
			cql:        "Interval[@2024-02-10T01:20:45.101-07:00, @2024-02-29T01:20:10.101-07:00] after second of Interval[@2024-01-25T01:20:30.101-07:00, @2024-02-10T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Interval<DateTime> after Interval<DateTime> with no precision",
			cql:        "Interval[@2024-02-12T01:20:30.101-07:00, @2024-02-29T01:20:10.101-07:00] after Interval[@2024-01-25T01:20:30.101-07:00, @2024-02-10T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Left Interval<DateTime> start matches closed right Interval end",
			cql:        "Interval[@2024-01-25T01:20:30.101-07:00, @2024-02-29T01:20:30.101-07:00] after day of Interval[@2024-01-20T01:20:30.101-07:00, @2024-01-25T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Left Interval<DateTime> start matches open right Interval end",
			cql:        "Interval[@2024-01-25T01:20:30.101-07:00, @2024-02-29T01:20:30.101-07:00] after Interval[@2024-01-20T01:20:30.101-07:00, @2024-01-25T01:20:30.101-07:00)",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Left Interval<DateTime> after right Interval with null end",
			cql:        "Interval[@2024-01-25T01:20:30.101-07:00, @2024-02-28T01:20:30.101-07:00] after Interval[@2024-02-29T01:20:30.101-07:00, null]",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Left Interval<DateTime> after right Interval with null start",
			cql:        "Interval[@2024-04-25T01:20:30.101-07:00, @2024-05-28T01:20:30.101-07:00] after Interval[null, @2024-02-29T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, true),
		},
		// Sanity check some Interval<Date>, Interval<Date> overloads
		{
			name:       "Interval<Date> after Interval<Date> at day precision",
			cql:        "Interval[@2024-02-29, @2024-03-29] after day of Interval[@2024-01-29, @2024-02-28]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Interval<Date> after Interval<Date> without precision",
			cql:        "Interval[@2024-01-25, @2024-02-28] after Interval[@2024-02-29, @2024-03-29]",
			wantResult: newOrFatal(t, false),
		},
		// starts or ends after syntax:
		{
			name:       "Interval<Date> starts after another",
			cql:        "Interval[@2028, @2029] starts after Interval[@2027, @2027]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Interval<Date> does not start after another",
			cql:        "Interval[@2021, @2026] starts after Interval[@2024, @2027]",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "null Interval<Date> starts after Interval<Date>",
			cql:        "null as Interval<Date> starts after Interval[@2024, @2027]",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "null low bound starts after Interval<Date>",
			cql:        "Interval[null, @2027] starts after Interval[@2024, @2027]",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Interval<Date> ends after another",
			cql:        "Interval[@2024, @2028] ends after Interval[@2027, @2027]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Interval<Date> does not end after another",
			cql:        "Interval[@2024, @2026] ends after Interval[@2024, @2027]",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "null Interval<Date> ends after Interval<Date>",
			cql:        "null as Interval<Date> ends after Interval[@2024, @2027]",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "null high bound ends after Interval<Date>",
			cql:        "Interval[@2024, null] ends after Interval[@2024, @2027]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Interval<DateTime> starts after another",
			cql:        "Interval[@2028T, @2029T] starts after Interval[@2027T, @2027T]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Interval<DateTime> does not start after another",
			cql:        "Interval[@2021T, @2026T] starts after Interval[@2024T, @2027T]",
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

func TestIntervalSameOrBefore(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Day precision true",
			cql:  "@2020-03-04 on or before day of Interval[@2020-03-05, @2020-03-07]",
			wantModel: &model.SameOrBefore{
				BinaryExpression: &model.BinaryExpression{
					Expression: model.ResultType(types.Boolean),
					Operands: []model.IExpression{
						model.NewLiteral("@2020-03-04", types.Date),
						&model.Interval{
							Expression:    model.ResultType(&types.Interval{PointType: types.Date}),
							Low:           model.NewLiteral("@2020-03-05", types.Date),
							High:          model.NewLiteral("@2020-03-07", types.Date),
							LowInclusive:  true,
							HighInclusive: true,
						},
					},
				},
				Precision: model.DAY,
			},
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Date on or before null as Interval<Date>",
			cql:        "@2020-03-04 on or before null as Interval<Date>",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Same day is true",
			cql:        "@2020-03-05 on or before day of Interval[@2020-03-05, @2020-03-07]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Interval right null",
			cql:        "@2020-03-04 on or before day of Interval[@2020-03-05, null]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Interval left null",
			cql:        "@2020-03-04 on or before day of Interval[null, @2020-03-07]",
			wantResult: newOrFatal(t, false),
		},
		{
			name: "Interval left null with minimum Date",
			// TODO(b/329322517): swap to minimum date literal when we can represent it with functional
			// syntax (does not currently parse).
			cql:        "@0001-01-01 on or before day of Interval[null, @2020-03-07]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "DateTime before interval",
			cql:        "@2024-02-28T01:20:30.101-07:00 on or before day of Interval[@2024-02-29T01:20:30.101-07:00, @2024-03-29T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "DateTime before interval without precision",
			cql:        "@2024-02-28T01:20:30.101-07:00 on or before Interval[@2024-02-29T01:20:30.101-07:00, @2024-03-29T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "DateTime matches closed interval start",
			cql:        "@2024-02-29T01:20:30.101-07:00 on or before day of Interval[@2024-02-29T01:20:30.101-07:00, @2024-03-29T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "DateTime Interval right null",
			cql:        "@2020-03-04T on or before day of Interval[@2020-03-05T, null]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "DateTime Interval left null",
			cql:        "@2020-03-04T on or before day of Interval[null, @2020-03-07T]",
			wantResult: newOrFatal(t, false),
		},
		{
			name: "DateTime Interval left null with minimum DateTime",
			// TODO(b/329322517): swap to minimum date literal when we can represent it with functional
			// syntax (does not currently parse).
			cql:        "minimum DateTime on or before day of Interval[null, @2020-03-07T]",
			wantResult: newOrFatal(t, true),
		},
		// Interval<DateTime>, Interval<DateTime> overloads:
		{
			name:       "Interval<DateTime> on or before null as Interval<DateTime>",
			cql:        "Interval[@2024-01-25T01:20:30.101-07:00, @2024-01-29T01:20:30.101-07:00] on or before null as Interval<DateTime>",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "null as Interval<DateTime> on or before Interval<DateTime>",
			cql:        "null as Interval<DateTime> on or before Interval[@2024-01-25T01:20:30.101-07:00, @2024-01-29T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Interval<DateTime> on or before Interval<DateTime> at seconds precision",
			cql:        "Interval[@2024-01-25T01:20:30.101-07:00, @2024-02-29T01:20:10.101-07:00] on or before second of Interval[@2024-02-29T01:20:30.101-07:00, @2024-03-29T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Interval<DateTime> on or before Interval<DateTime> without precision",
			cql:        "Interval[@2024-01-25T01:20:30.101-07:00, @2024-01-29T01:20:30.101-07:00] on or before Interval[@2024-02-29T01:20:30.101-07:00, @2024-03-29T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, true),
		},
		{
			// For "on or before" this is true, whereas for "before" this is false
			name:       "Left Interval<DateTime> end equals closed right interval start",
			cql:        "Interval[@2024-01-25T01:20:30.101-07:00, @2024-02-29T01:20:30.101-07:00] on or before Interval[@2024-02-29T01:20:30.101-07:00, @2024-03-29T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Left Interval<DateTime> end equals open right interval start",
			cql:        "Interval[@2024-01-25T01:20:30.101-07:00, @2024-02-29T01:20:30.101-07:00] on or before Interval(@2024-02-29T01:20:30.101-07:00, @2024-03-29T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, true),
		},
		// Sanity check some Interval<Date>, Interval<Date> overloads
		{
			name:       "Left Interval<Date> end equals closed right interval start",
			cql:        "Interval[@2024-01-25, @2024-02-28] on or before Interval[@2024-02-28, @2024-03-29]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Interval<Date> before Interval<Date> at day precision",
			cql:        "Interval[@2024-01-25, @2024-02-28] on or before day of Interval[@2024-02-29, @2024-03-29]",
			wantResult: newOrFatal(t, true),
		},
		// starts or ends before syntax with Dates:
		{
			name: "Interval<Date> starts on or before another",
			// For "on or before" this is true, whereas for "before" this is false
			cql:        "Interval[@2024, @2026] starts on or before Interval[@2024, @2027]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Interval<Date> does not start on or before another",
			cql:        "Interval[@2025, @2026] starts on or before Interval[@2024, @2027]",
			wantResult: newOrFatal(t, false),
		},
		{
			name: "Interval<Date> ends on or before another",
			// For "on or before" this is true, whereas for "before" this is false
			cql:        "Interval[@2024, @2028] ends on or before Interval[@2028, @2029]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Interval<Date> does not end on or before another",
			cql:        "Interval[@2025, @2026] ends on or before Interval[@2024, @2027]",
			wantResult: newOrFatal(t, false),
		},
		{
			name: "Interval<DateTime> starts on or before another",
			// For "on or before" this is true, whereas for "before" this is false
			cql:        "Interval[@2024T, @2026T] starts on or before Interval[@2024T, @2027T]",
			wantResult: newOrFatal(t, true),
		},
		{
			name: "Interval<DateTime> ends on or before another",
			// For "on or before" this is true, whereas for "before" this is false
			cql:        "Interval[@2024T, @2028T] ends on or before Interval[@2028T, @2029T]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Interval<Date> starts on or before another with open bounds",
			cql:        "Interval[@2024, @2026] starts on or before Interval(@2024, @2027)",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Interval<Date> starts on or before another with open bounds both sides",
			cql:        "Interval(@2024, @2026] starts on or before Interval(@2024, @2027)",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Interval<Date> does not start on or before another with open bounds",
			cql:        "Interval(@2024, @2026] starts on or before Interval(@2023, @2027)",
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

func TestIntervalSameOrAfter(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Day precision true",
			cql:  "@2020-03-09 on or after day of Interval[@2020-03-05, @2020-03-07]",
			wantModel: &model.SameOrAfter{
				BinaryExpression: &model.BinaryExpression{
					Expression: model.ResultType(types.Boolean),
					Operands: []model.IExpression{
						model.NewLiteral("@2020-03-09", types.Date),
						&model.Interval{
							Expression:    model.ResultType(&types.Interval{PointType: types.Date}),
							Low:           model.NewLiteral("@2020-03-05", types.Date),
							High:          model.NewLiteral("@2020-03-07", types.Date),
							LowInclusive:  true,
							HighInclusive: true,
						},
					},
				},
				Precision: model.DAY,
			},
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Date on or after null as Interval<Date>",
			cql:        "@2020-03-04 on or after null as Interval<Date>",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Same day is true",
			cql:        "@2020-03-07 on or after day of Interval[@2020-03-05, @2020-03-07]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Interval right null",
			cql:        "@2020-03-09 on or after day of Interval[@2020-03-05, null]",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Interval right null with max date",
			cql:        "@9999-12-31 on or after day of Interval[@2020-03-05, null]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Interval left null",
			cql:        "@2020-03-09 on or after day of Interval[null, @2020-03-07]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "DateTime after interval",
			cql:        "@2024-04-28T01:20:30.101-07:00 on or after day of Interval[@2024-02-28T01:20:30.101-07:00, @2024-02-29T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "DateTime after interval without precision",
			cql:        "@2024-04-28T01:20:30.101-07:00 on or after Interval[@2024-02-29T01:20:30.101-07:00, @2024-03-29T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "DateTime equals end of closed interval",
			cql:        "@2024-03-28T01:20:30.101-07:00 on or after day of Interval[@2024-02-28T01:20:30.101-07:00, @2024-03-28T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "DateTime Interval right null",
			cql:        "@2024-02-28T01:20:30.101-07:00 on or after day of Interval[@2024-01-28T01:20:30.101-07:00, null]",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Interval right null with max DateTime",
			cql:        "maximum DateTime on or after day of Interval[@2020-03-05T, null]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "DateTime after Interval with left null",
			cql:        "@2024-02-29T01:20:30.101-07:00 on or after day of Interval[null, @2024-02-28T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, true),
		},
		// Interval<DateTime>, Interval<DateTime> overloads:
		{
			name:       "Interval<DateTime> on or after null as Interval<DateTime>",
			cql:        "Interval[@2024-01-25T01:20:30.101-07:00, @2024-01-29T01:20:30.101-07:00] on or after null as Interval<DateTime>",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "null as Interval<DateTime> on or after Interval<DateTime>",
			cql:        "null as Interval<DateTime> on or after Interval[@2024-01-25T01:20:30.101-07:00, @2024-01-29T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Interval<DateTime> on or after Interval<DateTime> at seconds precision",
			cql:        "Interval[@2024-02-10T01:20:45.101-07:00, @2024-02-29T01:20:10.101-07:00] on or after second of Interval[@2024-01-25T01:20:30.101-07:00, @2024-02-10T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Interval<DateTime> on or after Interval<DateTime> with no precision",
			cql:        "Interval[@2024-02-12T01:20:30.101-07:00, @2024-02-29T01:20:10.101-07:00] on or after Interval[@2024-01-25T01:20:30.101-07:00, @2024-02-10T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, true),
		},
		{
			// 'on or after' returns true here, whereas 'after' would return false
			name:       "Left Interval<DateTime> start matches closed right Interval end",
			cql:        "Interval[@2024-01-25T01:20:30.101-07:00, @2024-02-29T01:20:30.101-07:00] on or after Interval[@2024-01-20T01:20:30.101-07:00, @2024-01-25T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Left Interval<DateTime> start matches open right Interval end",
			cql:        "Interval[@2024-01-25T01:20:30.101-07:00, @2024-02-29T01:20:30.101-07:00] after Interval[@2024-01-20T01:20:30.101-07:00, @2024-01-25T01:20:30.101-07:00)",
			wantResult: newOrFatal(t, true),
		},
		// Sanity check some Interval<Date>, Interval<Date> overloads
		{
			name:       "Interval<Date> on or after Interval<Date> at day precision",
			cql:        "Interval[@2024-02-29, @2024-03-29] on or after day of Interval[@2024-01-29, @2024-02-28]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Interval<Date> after Interval<Date> without precision",
			cql:        "Interval[@2024-01-25, @2024-02-28] on or after Interval[@2024-02-29, @2024-03-29]",
			wantResult: newOrFatal(t, false),
		},
		// starts or ends after syntax:
		{
			name: "Interval<Date> starts on or after another",
			// 'on or after' returns true here, whereas 'after' would return false
			cql:        "Interval[@2029, @2030] starts on or after Interval[@2028, @2029]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Interval<Date> does not start on or after another",
			cql:        "Interval[@2021, @2026] starts on or after Interval[@2024, @2027]",
			wantResult: newOrFatal(t, false),
		},
		{
			name: "Interval<Date> ends on or after another",
			// 'on or after' returns true here, whereas 'after' would return false
			cql:        "Interval[@2024, @2028] ends on or after Interval[@2027, @2028]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Interval<Date> does not end after another",
			cql:        "Interval[@2024, @2026] ends on or after Interval[@2024, @2027]",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Interval<Date> starts on or after another with open bounds",
			cql:        "Interval[@2027, @2028] starts on or after Interval(@2024, @2027)",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Interval<Date> starts on or after another with open bounds both sides",
			cql:        "Interval(@2027, @2028] starts on or after Interval(@2024, @2027)",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Interval<Date> does not start on or after another with open bounds",
			cql:        "Interval(@2024, @2026] starts on or after Interval(@2023, @2027)",
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

// TestIntervalRelativeOffsetBefore test cases where 'or more' or 'or less' operator appears.
func TestIntervalRelativeOffsetBefore(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Interval ends 1 year or less on or before end of another",
			cql:  "Interval[@2015, @2020] ends 1 year or less on or before end of Interval[@2019, @2022]",
			wantModel: &model.In{
				BinaryExpression: &model.BinaryExpression{
					Expression: model.ResultType(types.Boolean),
					Operands: []model.IExpression{
						&model.End{
							UnaryExpression: &model.UnaryExpression{
								Operand:    model.NewInclusiveInterval("@2015", "@2020", types.Date),
								Expression: model.ResultType(types.Date),
							},
						},
						&model.Interval{
							Expression: model.ResultType(&types.Interval{PointType: types.Date}),
							Low: &model.Subtract{
								BinaryExpression: &model.BinaryExpression{
									Operands: []model.IExpression{
										&model.End{
											UnaryExpression: &model.UnaryExpression{
												Operand:    model.NewInclusiveInterval("@2019", "@2022", types.Date),
												Expression: model.ResultType(types.Date),
											},
										},
										&model.Quantity{Value: 1, Unit: "year", Expression: model.ResultType(types.Quantity)},
									},
									Expression: model.ResultType(types.Date),
								},
							},
							High: &model.End{
								UnaryExpression: &model.UnaryExpression{
									Operand:    model.NewInclusiveInterval("@2019", "@2022", types.Date),
									Expression: model.ResultType(types.Date),
								},
							},
							LowInclusive:  true,
							HighInclusive: true,
						},
					},
				},
			},
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Interval ends 1 year or less on or before end of another",
			cql:        "Interval[@2015, @2021] ends 1 year or less on or before end of Interval[@2019, @2022]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Date 1 year or more on or before start of Interval",
			cql:        "@2015 1 year or more on or before start of Interval[@2019, @2022]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Date 1 year or more on or before end of Interval",
			cql:        "@2030 1 year or more on or after end of Interval[@2019, @2022]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Date 1 year or more on or after end of Interval",
			cql:        "@2022 1 year or more on or after end of Interval[@2019, @2022]",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Interval<integer> ends 1'1' or less on or before end of another",
			cql:        "Interval[1, 3] ends 1 '1' or less on or before end of Interval[3, 4]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Interval<integer> ends 1'1' or less on or after end of another",
			cql:        "Interval[5, 8] starts 1 '1' or less on or after end of Interval[3, 4]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Interval<integer> ends 1'1' or more on or before start of another",
			cql:        "Interval[0, 1] ends 1 '1' or more on or before start of Interval[3, 4]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Interval<integer> starts 1'1' or more on or after start of another",
			cql:        "Interval[6, 8] starts 1 '1' or more on or after start of Interval[3, 4]",
			wantResult: newOrFatal(t, true),
		},
		// Overload matching tests that the or less operator properly applies conversions on manually
		// inserted models in the parser.
		{
			name:       "Interval 1 year or less on or before Interval",
			cql:        "Interval[@2015, @2020] ends 1 year or less on or before Interval[@2019, @2022]",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Interval<Date> 1 year or less on or before DateTime",
			cql:        "Interval[@2015, @2016] ends 1 year or less on or before @2019-06-06T",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Date 1 year or less on or before DateTime",
			cql:        "@2016 1 year or less on or before @2019-06-06T",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "DateTime 1 year or less on or before Date",
			cql:        "@2016-06-06T 1 year or less on or before @2019",
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

func TestIntervalIn(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Point is null date",
			cql:  "null in year of Interval[@2020, @2022]",
			wantModel: &model.In{
				BinaryExpression: &model.BinaryExpression{
					Expression: model.ResultType(types.Boolean),
					Operands: []model.IExpression{
						&model.As{
							UnaryExpression: &model.UnaryExpression{
								Expression: model.ResultType(types.Date),
								Operand:    model.NewLiteral("null", types.Any),
							},
							AsTypeSpecifier: types.Date,
						},
						model.NewInclusiveInterval("@2020", "@2022", types.Date),
					},
				},
				Precision: model.YEAR,
			},
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Point is null datetime",
			cql:        "null in year of Interval[@2024-03-31T00:00:00.000Z, @2024-03-31T00:00:00.000Z]",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "@2020 in null",
			cql:        "@2020 in null",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "On inclusive bound date",
			cql:        "@2020-03 in month of Interval[@2020-03-25, @2022-04)",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "On inclusive bound datetime",
			cql:        "@2024-03-31T00:00:00.000Z in month of Interval[@2024-03-31T00:00:00.000Z, @2025-03-31T00:00:00.000Z)",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "On exclusive bound date",
			cql:        "@2020-03 in month of Interval(@2020-03-25, @2022-04)",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "On exclusive bound datetime",
			cql:        "@2024-03-31T00:00:00.000Z in month of Interval(@2024-03-31T00:00:00.000Z, @2025-03-31T00:00:00.000Z)",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "On exclusive and inclusive bound",
			cql:        "@2020-03 in month of Interval(@2020-03, @2022-03]",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Insufficient precision date",
			cql:        "@2020-03 in day of Interval[@2020-03-25, @2020-04]",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Insufficient precision datetime",
			cql:        "@2024-03 in day of Interval[@2024-03-28T00:00:00.000Z, @2024-03-31T00:00:00.000Z]",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Insufficient precision but for sure false",
			cql:        "@2028-03 in day of Interval[@2020-03-25, @2020-04]",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Null inclusive bound is true",
			cql:        "@2020-03 in month of Interval[null, @2022-04)",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Null inclusive bound but this is for sure false",
			cql:        "@2025-03 in month of Interval[null, @2022-04)",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Null exclusive bound is null",
			cql:        "@2021-03 in month of Interval(null, @2022-04)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Null exclusive bound but this is for sure false",
			cql:        "@2025-03 in month of Interval(null, @2022-04)",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "No in operator precision: On inclusive bound date",
			cql:        "@2020-03-25 in Interval[@2020-03-25, @2022-04)",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "No in operator precision: On inclusive bound datetime",
			cql:        "@2024-03-31T00:00:00.000Z in Interval[@2024-03-31T00:00:00.000Z, @2025-03-31T00:00:00.000Z)",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "No in operator precision: On exclusive bound date",
			cql:        "@2020-03-25 in Interval(@2020-03-25, @2022-04)",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "No in operator precision: On exclusive bound datetime",
			cql:        "@2024-03-31T00:00:00.000Z in Interval(@2024-03-31T00:00:00.000Z, @2025-03-31T00:00:00.000Z)",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "No in operator precision with differing operand precision",
			cql:        "@2020-03 in Interval[@2020-03-25, @2022-04-25)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "integer in interval",
			cql:        "42 in Interval[0, 100]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "integer not in interval",
			cql:        "42 in Interval[0, 25]",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "integer in interval on bounds",
			cql:        "25 in Interval[0, 25]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "integer in interval on bounds not inclusive",
			cql:        "25 in Interval[0, 25)",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "integer before interval bounds not inclusive",
			cql:        "24 in Interval[0, 25)",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "integer in interval on bounds",
			cql:        "25 in Interval[0, 25)",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "integer in interval on bounds",
			cql:        "25 in Interval[0, null)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Double in interval on upper bounds",
			cql:        "1.5 in Interval[1.0, 1.5]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Long not in interval",
			cql:        "0L in Interval[1L, 2L]",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Quantity in interval on bounds",
			cql:        "1'cm' in Interval[1'cm', 2'cm')",
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
				t.Errorf("Evaluate diff (-want +got)\n%v", diff)
			}
		})
	}
}

func TestIntervalIncludedIn(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		// TODO: b/331225778 - Null support handling for Included In operator
		{
			name: "On inclusive bound date",
			cql:  "@2020-03 included in month of Interval[@2020-03-25, @2022-04)",
			// Model should implicitly be converted to model.In when left operator is point type.
			wantModel: &model.In{
				BinaryExpression: &model.BinaryExpression{
					Expression: model.ResultType(types.Boolean),
					Operands: []model.IExpression{
						model.NewLiteral("@2020-03", types.Date),
						&model.Interval{
							Low:           model.NewLiteral("@2020-03-25", types.Date),
							High:          model.NewLiteral("@2022-04", types.Date),
							LowInclusive:  true,
							HighInclusive: false,
							Expression:    model.ResultType(&types.Interval{PointType: types.Date}),
						},
					},
				},
				Precision: model.MONTH,
			},
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "On right arg null",
			cql:        "@2020-03 included in month of null",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "During syntax, inclusive bound date",
			cql:        "@2020-03 during month of Interval[@2020-03-25, @2022-04)",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "On inclusive bound datetime",
			cql:        "@2024-03-31T00:00:00.000Z included in month of Interval[@2024-03-31T00:00:00.000Z, @2025-03-31T00:00:00.000Z)",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "On exclusive bound date",
			cql:        "@2020-03 included in month of Interval(@2020-03-25, @2022-04)",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "On exclusive bound datetime",
			cql:        "@2024-03-31T00:00:00.000Z included in month of Interval(@2024-03-31T00:00:00.000Z, @2025-03-31T00:00:00.000Z)",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "On inclusive bound datetime second precision",
			cql:        "@2024-03-31T00:00:00.000Z included in second of Interval[@2024-03-31T00:00:00.000Z, @2024-03-31T00:00:05.000Z]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "On exclusive and inclusive bound",
			cql:        "@2020-03 included in month of Interval(@2020-03, @2022-03]",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Insufficient precision date",
			cql:        "@2020-03 included in day of Interval[@2020-03-25, @2020-04]",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Insufficient precision datetime",
			cql:        "@2024-03 included in day of Interval[@2024-03-28T00:00:00.000Z, @2024-03-31T00:00:00.000Z]",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Insufficient precision but for sure false",
			cql:        "@2028-03 included in day of Interval[@2020-03-25, @2020-04]",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Null inclusive bound is true",
			cql:        "@2020-03 included in month of Interval[null, @2022-04)",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Null inclusive bound but this is for sure false",
			cql:        "@2025-03 included in month of Interval[null, @2022-04)",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Null exclusive bound is null",
			cql:        "@2021-03 included in month of Interval(null, @2022-04)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Null exclusive bound but this is for sure false",
			cql:        "@2025-03 included in month of Interval(null, @2022-04)",
			wantResult: newOrFatal(t, false),
		},
		// No precision
		{
			name:       "No included in operator precision: On inclusive bound date",
			cql:        "@2020-03-25 included in Interval[@2020-03-25, @2022-04)",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "No included in operator precision: On exclusive bound datetime",
			cql:        "@2024-03-31T00:00:00.000Z included in Interval(@2024-03-31T00:00:00.000Z, @2025-03-31T00:00:00.000Z)",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "No included in operator precision with differing operand precision",
			cql:        "@2020-03 included in Interval[@2020-03-25, @2022-04-25)",
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

func TestIntervalOverlaps(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		// Interval<DateTime>, Interval<DateTime> overloads:
		{
			name:       "Interval<DateTime> overlaps null as Interval<DateTime>",
			cql:        "Interval[@2024-01-25T01:20:30.101-07:00, @2024-01-29T01:20:30.101-07:00] overlaps null as Interval<DateTime>",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "null as Interval<DateTime> overlaps Interval<DateTime>",
			cql:        "null as Interval<DateTime> overlaps Interval[@2024-01-25T01:20:30.101-07:00, @2024-01-29T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Left ends during right",
			cql:        "Interval[@2024-01-25T01:20:30.101-07:00, @2024-03-02T01:20:30.101-07:00] overlaps Interval[@2024-02-29T01:20:30.101-07:00, @2024-03-29T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Left starts during right",
			cql:        "Interval[@2024-03-25T01:20:30.101-07:00, @2024-06-02T01:20:30.101-07:00] overlaps Interval[@2024-02-29T01:20:30.101-07:00, @2024-03-29T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Left contains right, or starts before right and ends after right",
			cql:        "Interval[@2024-01-25T01:20:30.101-07:00, @2024-06-02T01:20:30.101-07:00] overlaps Interval[@2024-02-29T01:20:30.101-07:00, @2024-03-29T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Right contains left, or starts after right and ends before right",
			cql:        "Interval[@2024-02-29T01:20:30.101-07:00, @2024-03-29T01:20:30.101-07:00] overlaps Interval[@2024-01-25T01:20:30.101-07:00, @2024-06-02T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Left strictly before right",
			cql:        "Interval[@2024-01-25T01:20:30.101-07:00, @2024-01-29T01:20:30.101-07:00] overlaps Interval[@2024-02-29T01:20:30.101-07:00, @2024-03-29T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Left strictly after right",
			cql:        "Interval[@2024-02-29T01:20:30.101-07:00, @2024-03-29T01:20:30.101-07:00] overlaps Interval[@2024-01-25T01:20:30.101-07:00, @2024-01-29T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Left Interval<DateTime> end matches closed right interval start",
			cql:        "Interval[@2024-01-25T01:20:30.101-07:00, @2024-02-29T01:20:30.101-07:00] overlaps Interval[@2024-02-29T01:20:30.101-07:00, @2024-03-29T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Left Interval<DateTime> end matches open right interval start",
			cql:        "Interval[@2024-01-25T01:20:30.101-07:00, @2024-02-29T01:20:30.101-07:00] overlaps Interval(@2024-02-29T01:20:30.101-07:00, @2024-03-29T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Left Interval<DateTime> overlaps right Interval with null end",
			cql:        "Interval[@2024-01-25T01:20:30.101-07:00, @2024-03-28T01:20:30.101-07:00] overlaps Interval[@2024-02-29T01:20:30.101-07:00, null]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Left Interval<DateTime> overlaps right Interval with null start",
			cql:        "Interval[@2024-01-25T01:20:30.101-07:00, @2024-02-28T01:20:30.101-07:00] overlaps Interval[null, @2024-02-29T01:20:30.101-07:00]",
			wantResult: newOrFatal(t, true),
		},
		// Sanity check some Interval<Date>, Interval<Date> overloads
		{
			name:       "Left strictly before right",
			cql:        "Interval[@2024-01-25, @2024-02-28] overlaps Interval[@2024-02-29, @2024-03-29]",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Left strictly after right",
			cql:        "Interval[@2024-02-29, @2024-03-29] overlaps Interval[@2024-01-25, @2024-02-28]",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Left Interval<DateTime> end matches closed right interval start",
			cql:        "Interval[@2024-01-25, @2024-02-29] overlaps Interval[@2024-02-29, @2024-03-29]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Left Interval<DateTime> end matches open right interval start",
			cql:        "Interval[@2024-01-25, @2024-02-29] overlaps Interval(@2024-02-29, @2024-03-29]",
			wantResult: newOrFatal(t, false),
		},
		// mixed precision tests
		{
			name:       "Left ends during right but insufficient precision to determine overlap",
			cql:        "Interval[@2024-01-25, @2024-02-28] overlaps Interval[@2024-02, @2024-03-29]",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Left starts during right but insufficient precision to determine overlap",
			cql:        "Interval[@2024-02-28, @2024-03-29] overlaps Interval[@2024-01-25, @2024-02]",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Left starts and ends during right uncertain period",
			cql:        "Interval[@2025-01-25, @2025-02-28] overlaps Interval[@2024-02, @2025]",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Right starts during uncertain period and ends before left ends",
			cql:        "Interval[@2024, @2025-02] overlaps Interval[@2024-02, @2025-01]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Right starts during left and ends during uncertain period",
			cql:        "Interval[@2024-02, @2025] overlaps Interval[@2024-03, @2025-02]",
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

func TestIntervalContains(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		// TODO: b/331225778 - Null support handling for Contains operator
		{
			name: "On inclusive bound date",
			cql:  "Interval[@2020-03-25, @2022-04] contains month of @2020-03",
			wantModel: &model.In{
				BinaryExpression: &model.BinaryExpression{
					Expression: model.ResultType(types.Boolean),
					Operands: []model.IExpression{
						model.NewLiteral("@2020-03", types.Date),
						&model.Interval{
							Low:           model.NewLiteral("@2020-03-25", types.Date),
							High:          model.NewLiteral("@2022-04", types.Date),
							LowInclusive:  true,
							HighInclusive: true,
							Expression:    model.ResultType(&types.Interval{PointType: types.Date}),
						},
					},
				},
				Precision: model.MONTH,
			},
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Point arg null datetime",
			cql:        "Interval[@2024-03-31T00:00:00.000Z, @2024-03-31T00:00:00.000Z] contains null",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "On inclusive bound datetime",
			cql:        "Interval[@2024-03-31T00:00:00.000Z, @2025-03-31T00:00:00.000Z) contains month of @2024-03-31T00:00:00.000Z",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "On exclusive bound date",
			cql:        "Interval(@2020-03-25, @2022-04) contains month of @2020-03",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "On exclusive bound datetime",
			cql:        "Interval(@2024-03-31T00:00:00.000Z, @2025-03-31T00:00:00.000Z) contains month of @2024-03-31T00:00:00.000Z",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "On inclusive bound datetime second precision",
			cql:        "Interval[@2024-03-31T00:00:00.000Z, @2024-03-31T00:00:05.000Z] contains second of @2024-03-31T00:00:00.000Z",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "On exclusive and inclusive bound",
			cql:        "Interval(@2020-03, @2022-03] contains month of @2020-03",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Insufficient precision date",
			cql:        "Interval[@2020-03-25, @2020-04] contains day of @2020-03",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Insufficient precision datetime",
			cql:        "Interval[@2024-03-28T00:00:00.000Z, @2024-03-31T00:00:00.000Z] contains day of @2024-03",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Insufficient precision but for sure false",
			cql:        "Interval[@2028-03-25, @2020-04] contains day of @2020-03",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Null inclusive bound is true",
			cql:        "Interval[null, @2022-04) contains month of @2020-03",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Null inclusive bound but this is for sure false",
			cql:        "Interval[null, @2022-04) contains month of @2025-03",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Null exclusive bound is null",
			cql:        "Interval(null, @2022-04) contains month of @2021-03",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Null exclusive bound but this is for sure false",
			cql:        "Interval(null, @2022-04) contains month of @2025-03",
			wantResult: newOrFatal(t, false),
		},
		// No precision
		{
			name:       "No included in operator precision: On inclusive bound date",
			cql:        "Interval[@2020-03-25, @2022-04) contains day of @2020-03-25",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "No included in operator precision: On exclusive bound datetime",
			cql:        "Interval(@2024-03-31T00:00:00.000Z, @2025-03-31T00:00:00.000Z) contains day of @2024-03-31T00:00:00.000Z",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "No included in operator precision with differing operand precision",
			cql:        "Interval[@2020-03-25, @2022-04-25) contains @2020-03",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "interval contains integer",
			cql:        "Interval[0, 100] contains 42",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Interval does not contain integer",
			cql:        "Interval[0, 25] contains 42",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Interval contains integer on bounds",
			cql:        "Interval[0, 25] contains 25",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Interval on bounds not inclusive integer",
			cql:        "Interval[0, 25) contains 25",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Integer before interval bounds not inclusive",
			cql:        "Interval[0, 25) contains 24",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Interval not contains integer with null exlusive bounds",
			cql:        "Interval[0, null) contains 25",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Interval contains double on upper bounds",
			cql:        "Interval[1.0, 1.5] contains 1.5",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Interval not contains long",
			cql:        "Interval[1L, 2L] contains 0L",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Interval contains quantity on bounds",
			cql:        "Interval[1'cm', 2'cm') contains 1'cm'",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Functional syntax",
			cql:        "Contains(Interval[0, 100], 42)",
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

func TestIntervalWidth(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "width of inclusive interval",
			cql:  "width of Interval[1, 4]",
			wantModel: &model.Width{
				UnaryExpression: &model.UnaryExpression{
					Operand: &model.Interval{
						Low:           model.NewLiteral("1", types.Integer),
						High:          model.NewLiteral("4", types.Integer),
						LowInclusive:  true,
						HighInclusive: true,
						Expression:    model.ResultType(&types.Interval{PointType: types.Integer}),
					},
					Expression: model.ResultType(types.Integer),
				},
			},
			wantResult: newOrFatal(t, 3),
		},
		{
			name:       "width of interval non-inclusive bounds",
			cql:        "width of Interval(1, 4)",
			wantResult: newOrFatal(t, 1),
		},
		{
			name:       "width of interval with null bounds",
			cql:        "width of Interval(null, 4)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "width of interval with null bounds",
			cql:        "width of Interval(1, null)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "width of interval with null bounds",
			cql:        "width of Interval(null as Integer, null as Integer)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "width of interval with decimal",
			cql:        "Round(width of Interval[1.0, 4.0], 1)",
			wantResult: newOrFatal(t, 3.0),
		},
		{
			name:       "width of interval with long",
			cql:        "width of Interval[1L, 4L]",
			wantResult: newOrFatal(t, int64(3)),
		},
		{
			name:       "width of interval with quantity",
			cql:        "width of Interval[1'cm', 4'cm']",
			wantResult: newOrFatal(t, result.Quantity{Value: 3, Unit: "cm"}),
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

func TestComparison_Error(t *testing.T) {
	tests := []struct {
		name                string
		cql                 string
		wantModel           model.IExpression
		wantEvalErrContains string
	}{
		{
			name:                "Quantity in interval with not matching lower unit",
			cql:                 "1'cm' in Interval[1'm', 2'cm']",
			wantEvalErrContains: "in operator recieved Quantities with differing unit values",
		},
		{
			name:                "Quantity in interval with not matching lower unit",
			cql:                 "1'cm' in Interval[1'cm', 2'm']",
			wantEvalErrContains: "in operator recieved Quantities with differing unit values",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := newFHIRParser(t)
			parsedLibs, err := p.Libraries(context.Background(), wrapInLib(t, tc.cql), parser.Config{})
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}

			_, err = interpreter.Eval(context.Background(), parsedLibs, defaultInterpreterConfig(t, p))
			if err == nil {
				t.Fatal("Eval succeeded, wanted error")
			}
			if !strings.Contains(err.Error(), tc.wantEvalErrContains) {
				t.Errorf("Unexpected evaluation error contents. got (%v), want contains (%v)", err.Error(), tc.wantEvalErrContains)
			}
		})
	}
}
