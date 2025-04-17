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
	"google.golang.org/protobuf/testing/protocmp"
)

func TestAllTrue(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "AllTrue({true, true, true})",
			cql:  "AllTrue({true, true, true})",
			wantModel: &model.AllTrue{
				UnaryExpression: &model.UnaryExpression{
					Operand:    model.NewList([]string{"true", "true", "true"}, types.Boolean),
					Expression: model.ResultType(types.Boolean),
				},
			},
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "AllTrue with null input",
			cql:        "AllTrue(null as List<Boolean>)",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "AllTrue({true, true, null})",
			cql:        "AllTrue({true, true, null})",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "AllTrue with empty list",
			cql:        "AllTrue({})",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "AllTrue with all null list",
			cql:        "AllTrue({null, null})",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "AllTrue with false in null list",
			cql:        "AllTrue({null, false})",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "AllTrue with false in true list",
			cql:        "AllTrue({true, false})",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "AllTrue where list contains null false and true",
			cql:        "AllTrue({true, null, false})",
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

func TestAnyTrue(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "AnyTrue({true, true, true})",
			cql:  "AnyTrue({true, true, true})",
			wantModel: &model.AnyTrue{
				UnaryExpression: &model.UnaryExpression{
					Operand:    model.NewList([]string{"true", "true", "true"}, types.Boolean),
					Expression: model.ResultType(types.Boolean),
				},
			},
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "AnyTrue with null input",
			cql:        "AnyTrue(null as List<Boolean>)",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "AnyTrue({true, true, null})",
			cql:        "AnyTrue({true, true, null})",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "AnyTrue with empty list",
			cql:        "AnyTrue({})",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "AnyTrue with all null list",
			cql:        "AnyTrue({null, null})",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "AnyTrue with false in null list",
			cql:        "AnyTrue({null, false})",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "AnyTrue with false in true list",
			cql:        "AnyTrue({false, true})",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "AnyTrue where list contains null false and true",
			cql:        "AnyTrue({false, null, true})",
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

func TestAvg(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Avg({1.0, 2.0, 3.0})",
			cql:  "Avg({1.0, 2.0, 3.0})",
			wantModel: &model.Avg{
				UnaryExpression: &model.UnaryExpression{
					Operand:    model.NewList([]string{"1.0", "2.0", "3.0"}, types.Decimal),
					Expression: model.ResultType(types.Decimal),
				},
			},
			wantResult: newOrFatal(t, 2.0),
		},
		{
			name:       "Avg({1.0, -1.0})",
			cql:        "Avg({1.0, -1.0})",
			wantResult: newOrFatal(t, 0.0),
		},
		{
			name:       "Avg with null input",
			cql:        "Avg(null as List<Decimal>)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Avg({1.0, 2.0, null})",
			cql:        "Avg({1.0, 2.0, null})",
			wantResult: newOrFatal(t, 1.5),
		},
		{
			name:       "Avg with empty list",
			cql:        "Avg({} as List<Decimal>)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Avg with all null decimal list",
			cql:        "Avg({null as Decimal, null as Decimal})",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Avg({2.5 'g', 3.5 'g', null as Quantity})",
			cql:        "Avg({2.5 'g', 3.5 'g', null as Quantity})",
			wantResult: newOrFatal(t, result.Quantity{Value: 3.0, Unit: "g"}),
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

func TestCount(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Count({1, 2, 3})",
			cql:  "Count({1, 2, 3})",
			wantModel: &model.Count{
				UnaryExpression: &model.UnaryExpression{
					Operand:    model.NewList([]string{"1", "2", "3"}, types.Integer),
					Expression: model.ResultType(types.Integer),
				},
			},
			wantResult: newOrFatal(t, 3),
		},
		{
			name:       "Count with null input",
			cql:        "Count(null as List<Integer>)",
			wantResult: newOrFatal(t, 0),
		},
		{
			name:       "Count({1, 2, null})",
			cql:        "Count({1, 2, null})",
			wantResult: newOrFatal(t, 2),
		},
		{
			name:       "Count with empty list",
			cql:        "Count({})",
			wantResult: newOrFatal(t, 0),
		},
		{
			name:       "Count with all null list",
			cql:        "Count({null, null})",
			wantResult: newOrFatal(t, 0),
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

func TestMax(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Max({@2010, @2012, @2011})",
			cql:  "Max({@2010, @2012, @2011})",
			wantModel: &model.Max{
				UnaryExpression: &model.UnaryExpression{
					Operand:    model.NewList([]string{"@2010", "@2012", "@2011"}, types.Date),
					Expression: model.ResultType(types.Date),
				},
			},
			wantResult: newOrFatal(t, result.Date{Date: time.Date(2012, time.January, 01, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.YEAR}),
		},
		{
			name:       "Max with null input",
			cql:        "Max(null as List<Date>)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Max({@2012, @2011, null})",
			cql:        "Max({@2012, @2011, null})",
			wantResult: newOrFatal(t, result.Date{Date: time.Date(2012, time.January, 01, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.YEAR}),
		},
		{
			name:       "Max with empty list",
			cql:        "Max(List<Date>{})",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Max with all null list",
			cql:        "Max({null as Date, null as Date})",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Max({@2014-01-01T01:01:00.000Z, @2014-01-01T01:03:00.000Z, @2014-01-01T01:02:00.000Z})",
			cql:        "Max({@2014-01-01T01:01:00.000Z, @2014-01-01T01:03:00.000Z, @2014-01-01T01:02:00.000Z})",
			wantResult: newOrFatal(t, result.DateTime{Date: time.Date(2014, time.January, 01, 1, 3, 0, 0, time.UTC), Precision: model.MILLISECOND}),
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

func TestMin(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Min({@2012, @2010, @2011})",
			cql:  "Min({@2012, @2010, @2011})",
			wantModel: &model.Min{
				UnaryExpression: &model.UnaryExpression{
					Operand:    model.NewList([]string{"@2012", "@2010", "@2011"}, types.Date),
					Expression: model.ResultType(types.Date),
				},
			},
			wantResult: newOrFatal(t, result.Date{Date: time.Date(2010, time.January, 01, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.YEAR}),
		},
		{
			name:       "Min with null input",
			cql:        "Min(null as List<Date>)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Min({@2012, @2011, null})",
			cql:        "Min({@2012, @2011, null})",
			wantResult: newOrFatal(t, result.Date{Date: time.Date(2011, time.January, 01, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.YEAR}),
		},
		{
			name:       "Min with empty list",
			cql:        "Min(List<Date>{})",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Min with all null list",
			cql:        "Min({null as Date, null as Date})",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Min({@2014-01-01T01:01:00.000Z, @2014-01-01T01:03:00.000Z, @2014-01-01T01:02:00.000Z})",
			cql:        "Min({@2014-01-01T01:01:00.000Z, @2014-01-01T01:03:00.000Z, @2014-01-01T01:02:00.000Z})",
			wantResult: newOrFatal(t, result.DateTime{Date: time.Date(2014, time.January, 01, 1, 1, 0, 0, time.UTC), Precision: model.MILLISECOND}),
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

func TestSum(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Sum({1, 2, 3})",
			cql:  "Sum({1, 2, 3})",
			wantModel: &model.Sum{
				UnaryExpression: &model.UnaryExpression{
					Operand:    model.NewList([]string{"1", "2", "3"}, types.Integer),
					Expression: model.ResultType(types.Integer),
				},
			},
			wantResult: newOrFatal(t, 6),
		},
		{
			name:       "Sum({1, -1})",
			cql:        "Sum({1, -1})",
			wantResult: newOrFatal(t, 0),
		},
		{
			name:       "Sum with null input",
			cql:        "Sum(null as List<Integer>)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Sum({1, 2, null})",
			cql:        "Sum({1, 2, null})",
			wantResult: newOrFatal(t, 3),
		},
		{
			name:       "Sum with empty list",
			cql:        "Sum({} as List<Integer>)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Sum with all null integer list",
			cql:        "Sum({null as Integer, null as Integer})",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Sum({2.1, 3.1})",
			cql:        "Sum({2.1, 3.1})",
			wantResult: newOrFatal(t, 5.2),
		},
		{
			name:       "Sum({100L, 900L})",
			cql:        "Sum({100L, 900L})",
			wantResult: newOrFatal(t, int64(1000)),
		},
		{
			name:       "Sum({2.1 'g', 3.1 'g'})",
			cql:        "Sum({2.1 'g', 3.1 'g'})",
			wantResult: newOrFatal(t, result.Quantity{Value: 5.2, Unit: "g"}),
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

func TestSum_Error(t *testing.T) {
	tests := []struct {
		name            string
		cql             string
		wantModel       model.IExpression
		wantErrContains string
	}{
		{
			name:            "Sum({2.1 'cm', 3.1 'g'})",
			cql:             "Sum({2.1 'cm', 3.1 'g'})",
			wantErrContains: "Quantity values with different units",
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
			if !strings.Contains(err.Error(), tc.wantErrContains) {
				t.Errorf("Eval returned unexpected error: %v, want error containing %q", err, tc.wantErrContains)
			}
		})
	}
}

func TestMedian(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Median({1.5, 2.5, 3.5, 4.5})",
			cql:  "Median({1.5, 2.5, 3.5, 4.5})",
			wantModel: &model.Median{
				UnaryExpression: &model.UnaryExpression{
					Operand:    model.NewList([]string{"1.5", "2.5", "3.5", "4.5"}, types.Decimal),
					Expression: model.ResultType(types.Decimal),
				},
			},
			wantResult: newOrFatal(t, 3.0),
		},
		{
			name:       "Median({1 'cm', 2 'cm', 3 'cm'})",
			cql:        "Median({1 'cm', 2 'cm', 3 'cm'})",
			wantResult: newOrFatal(t, result.Quantity{Value: 2.0, Unit: "cm"}),
		},
		{
			name:       "Median({1.5 'g', 2.5 'g', 3.5 'g', 4.5 'g'})",
			cql:        "Median({1.5 'g', 2.5 'g', 3.5 'g', 4.5 'g'})",
			wantResult: newOrFatal(t, result.Quantity{Value: 3.0, Unit: "g"}),
		},
		{
			name:       "Unordered Quantity list: Median({2.5 'g', 3.5 'g', 1.5 'g', 4.5 'g'})",
			cql:        "Median({2.5 'g', 3.5 'g', 1.5 'g', 4.5 'g'})",
			wantResult: newOrFatal(t, result.Quantity{Value: 3.0, Unit: "g"}),
		},
		{
			name:       "Median({1.0, 2.0, 3.0})",
			cql:        "Median({1.0, 2.0, 3.0})",
			wantResult: newOrFatal(t, 2.0),
		},
		{
			name:       "Median({1.5, 2.5, 3.5, 4.5})",
			cql:        "Median({1.5, 2.5, 3.5, 4.5})",
			wantResult: newOrFatal(t, 3.0),
		},
		{
			name:       "Unordered Decimal list: Median({2.5, 3.5, 1.5, 4.5})",
			cql:        "Median({2.5, 3.5, 1.5, 4.5})",
			wantResult: newOrFatal(t, 3.0),
		},
		{
			name:       "Median(List<Decimal>{})",
			cql:        "Median(List<Decimal>{})",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Median({null as Decimal})",
			cql:        "Median({null as Decimal})",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Median(null as List<Decimal>)",
			cql:        "Median(null as List<Decimal>)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Median(List<Quantity>{})",
			cql:        "Median(List<Quantity>{})",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Median({null as Quantity})",
			cql:        "Median({null as Quantity})",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Median(null as List<Quantity>)",
			cql:        "Median(null as List<Quantity>)",
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

func TestMode(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Mode({1, 2, 3, 2, 1, 2})",
			cql:  "Mode({1, 2, 3, 2, 1, 2})",
			wantModel: &model.Mode{
				UnaryExpression: &model.UnaryExpression{
					Operand:    model.NewList([]string{"1", "2", "3", "2", "1", "2"}, types.Integer),
					Expression: model.ResultType(types.Integer),
				},
			},
			wantResult: newOrFatal(t, 2),
		},
		{
			name:       "Mode({1, 2, 3, 3, 4, 5})",
			cql:        "Mode({1, 2, 3, 3, 4, 5})",
			wantResult: newOrFatal(t, 3),
		},
		{
			name:       "Mode with null input",
			cql:        "Mode(null as List<Integer>)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Mode({1, 2, null, 2})",
			cql:        "Mode({1, 2, null, 2})",
			wantResult: newOrFatal(t, 2),
		},
		{
			name:       "Mode with empty list",
			cql:        "Mode(List<Integer>{})",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Mode with all null list",
			cql:        "Mode({null as Integer, null as Integer})",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Mode({1.5, 2.5, 3.5, 2.5, 1.5, 2.5})",
			cql:        "Mode({1.5, 2.5, 3.5, 2.5, 1.5, 2.5})",
			wantResult: newOrFatal(t, 2.5),
		},
		{
			name:       "Mode({'a', 'b', 'c', 'b', 'a', 'b'})",
			cql:        "Mode({'a', 'b', 'c', 'b', 'a', 'b'})",
			wantResult: newOrFatal(t, "b"),
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
