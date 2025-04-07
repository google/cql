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
	"testing"

	"github.com/google/cql/interpreter"
	"github.com/google/cql/model"
	"github.com/google/cql/parser"
	"github.com/google/cql/result"
	"github.com/google/cql/types"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestConcatenate(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "'a' + 'b'",
			cql:  "'a' + 'b'",
			wantModel: &model.Concatenate{
				NaryExpression: &model.NaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("a", types.String),
						model.NewLiteral("b", types.String),
					},
					Expression: model.ResultType(types.String),
				},
			},
			wantResult: newOrFatal(t, "ab"),
		},
		{
			name:       "'a' + 'b' + 'c'",
			cql:        "'a' + 'b' + 'c'",
			wantResult: newOrFatal(t, "abc"),
		},
		{
			name:       "'a' + null",
			cql:        "'a' + null",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "null + 'a'",
			cql:        "null + 'a'",
			wantResult: newOrFatal(t, nil),
		},
		{
			name: "concatenate with & operator",
			cql:  "'a' & 'b'",
			wantModel: &model.Concatenate{
				NaryExpression: &model.NaryExpression{
					Operands: []model.IExpression{
						&model.Coalesce{
							NaryExpression: &model.NaryExpression{
								Operands:   []model.IExpression{model.NewLiteral("a", types.String), model.NewLiteral("", types.String)},
								Expression: model.ResultType(types.String),
							},
						},
						&model.Coalesce{
							NaryExpression: &model.NaryExpression{
								Operands:   []model.IExpression{model.NewLiteral("b", types.String), model.NewLiteral("", types.String)},
								Expression: model.ResultType(types.String),
							},
						},
					},
					Expression: model.ResultType(types.String),
				},
			},
			wantResult: newOrFatal(t, "ab"),
		},
		{
			name:       "concatenate using & treats null as empty string, when null is second input",
			cql:        "'a' & null",
			wantResult: newOrFatal(t, "a"),
		},
		{
			name:       "concatenate using & treats null as empty string, when null is first input",
			cql:        "null & 'a'",
			wantResult: newOrFatal(t, "a"),
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

func TestToString(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name:       "ToString(true)",
			cql:        "ToString(true)",
			wantResult: newOrFatal(t, "true"),
		},
		{
			name:       "ToString(false)",
			cql:        "ToString(false)",
			wantResult: newOrFatal(t, "false"),
		},
		{
			name:       "ToString(1)",
			cql:        "ToString(1)",
			wantResult: newOrFatal(t, "1"),
		},
		{
			name:       "ToString(-1)",
			cql:        "ToString(-1)",
			wantResult: newOrFatal(t, "-1"),
		},
		{
			name:       "ToString(100000L)",
			cql:        "ToString(100000L)",
			wantResult: newOrFatal(t, "100000"),
		},
		{
			name:       "ToString(-100000L)",
			cql:        "ToString(-100000L)",
			wantResult: newOrFatal(t, "-100000"),
		},
		{
			name:       "ToString(1.42)",
			cql:        "ToString(1.42)",
			wantResult: newOrFatal(t, "1.42"),
		},
		{
			name:       "ToString(-1.42)",
			cql:        "ToString(-1.42)",
			wantResult: newOrFatal(t, "-1.42"),
		},
		{
			name:       "ToString(1 'cm')",
			cql:        "ToString(1 'cm')",
			wantResult: newOrFatal(t, "1 'cm'"),
		},
		{
			name:       "ToString(-1 'cm')",
			cql:        "ToString(-1 'cm')",
			wantResult: newOrFatal(t, "-1 'cm'"),
		},
		{
			name:       "ToString(1'g':0.1'g')",
			cql:        "ToString(1'g':0.1'g')",
			wantResult: newOrFatal(t, "1 'g':0.1 'g'"),
		},
		{
			name:       "ToString(@2022-01-03)",
			cql:        "ToString(@2022-01-03)",
			wantResult: newOrFatal(t, "2022-01-03"),
		},
		{
			name:       "ToString(@2022-01-03T12:00:00Z)",
			cql:        "ToString(@2022-01-03T12:00:00Z)",
			wantResult: newOrFatal(t, "2022-01-03T12:00:00Z"),
		},
		{
			name:       "ToString(DateTime(2022, 1, 3))",
			cql:        "ToString(DateTime(2022, 1, 3))",
			wantResult: newOrFatal(t, "2022-01-03+04:00"),
		},
		{
			name:       "ToString(@T12:01:00)",
			cql:        "ToString(@T12:01:00)",
			wantResult: newOrFatal(t, "12:01:00"),
		},
		{
			name:       "ToString(null as Date)",
			cql:        "ToString(null as Date)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "ToString(null)",
			cql:        "ToString(null)",
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

func TestSplit(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Split('A,B,C', ',')",
			cql:  "Split('A,B,C', ',')",
			wantModel: &model.Split{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("A,B,C", types.String),
						model.NewLiteral(",", types.String),
					},
					Expression: model.ResultType(
						&types.List{ElementType: types.String},
					),
				},
			},
			wantResult: newOrFatal(t, result.List{Value: []result.Value{newOrFatal(t, "A"), newOrFatal(t, "B"), newOrFatal(t, "C")}, StaticType: &types.List{ElementType: types.String}}),
		},
		{
			name:       "Split with stringToSplit=null",
			cql:        "Split(null, ',')",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Split with seperator = null",
			cql:        "Split('test', null)",
			wantResult: newOrFatal(t, result.List{Value: []result.Value{newOrFatal(t, "test")}, StaticType: &types.List{ElementType: types.String}}),
		},
		{
			name:       "Split with seperator not found in stringToSplit",
			cql:        "Split('test abc', ',')",
			wantResult: newOrFatal(t, result.List{Value: []result.Value{newOrFatal(t, "test abc")}, StaticType: &types.List{ElementType: types.String}}),
		},
		{
			name:       "Split with both operands null",
			cql:        "Split(null, null)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Split with both multi-character seperator",
			cql:        "Split('a//b', '//')",
			wantResult: newOrFatal(t, result.List{Value: []result.Value{newOrFatal(t, "a"), newOrFatal(t, "b")}, StaticType: &types.List{ElementType: types.String}}),
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

func TestCombine(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Combine without separator",
			cql:  "Combine({'A', 'B'})",
			wantModel: &model.Combine{
				NaryExpression: &model.NaryExpression{
					Operands: []model.IExpression{
						&model.List{
							Expression: model.ResultType(&types.List{ElementType: types.String}),
							List: []model.IExpression{
								model.NewLiteral("A", types.String),
								model.NewLiteral("B", types.String),
							},
						},
					},
					Expression: model.ResultType(types.String),
				},
			},
			wantResult: newOrFatal(t, "AB"),
		},
		{
			name:       "Combine with separator",
			cql:        "Combine({'A', 'B'}, ':')",
			wantResult: newOrFatal(t, "A:B"),
		},
		{
			name:       "Combine with empty list is null",
			cql:        "Combine({})",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Combine with empty list and non-empty separator is null",
			cql:        "Combine({}, ':')",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Combine with null list is null",
			cql:        "Combine(null, ':')",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Combine with null separator is null",
			cql:        "Combine({'A', 'B'}, null)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Combine skips null elements in input list",
			cql:        "Combine({'A', 'B', null, 'C'})",
			wantResult: newOrFatal(t, "ABC"),
		},
		{
			name:       "Combine with list of nulls",
			cql:        "Combine({null as String, null as String})",
			wantResult: newOrFatal(t, ""),
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

func TestIndexerString(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "[] Indexer",
			cql:  "'abc'[1]",
			wantModel: &model.Indexer{
				BinaryExpression: &model.BinaryExpression{
					Expression: model.ResultType(types.String),
					Operands: []model.IExpression{
						model.NewLiteral("abc", types.String),
						model.NewLiteral("1", types.Integer),
					},
				},
			},
			wantResult: newOrFatal(t, "b"),
		},
		{
			name: "Indexer functional form",
			cql:  "Indexer('abc', 1)",
			wantModel: &model.Indexer{
				BinaryExpression: &model.BinaryExpression{
					Expression: model.ResultType(types.String),
					Operands: []model.IExpression{
						model.NewLiteral("abc", types.String),
						model.NewLiteral("1", types.Integer),
					},
				},
			},
			wantResult: newOrFatal(t, "b"),
		},
		{
			name:       "Indexer with index too large",
			cql:        "'abc'[100]",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Indexer with index smaller than 0",
			cql:        "'abc'[-100]",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Indexer on null",
			cql:        "(null as String)[1]",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Indexer with null index",
			cql:        "'abc'[null as Integer]",
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

func TestEndsWith(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "EndsWithTrue",
			cql:  "EndsWith('apple', 'ple')",
			wantModel: &model.EndsWith{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("apple", types.String),
						model.NewLiteral("ple", types.String),
					},
					Expression: model.ResultType(types.Boolean),
				},
			},
			wantResult: newOrFatal(t, true),
		},
		{
			name: "EndsWithFalse",
			cql: "EndsWith('apple', 'pel')",
			wantResult: newOrFatal(t, false),
		},
		{
			name: "EndsWithBothNull",
			cql: "EndsWith(null, null)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name: "EndsWithRightNull",
			cql: "EndsWith('apple',null)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name: "EndsWithLeftNull",
			cql: "EndsWith(null,'ple')",
			wantResult: newOrFatal(t,nil),
		},
		{
			name: "EndsWithLeftEmpty",
			cql: "EndsWith('','ple')",
			wantResult: newOrFatal(t,false),
		},
		{
			name: "EndsWithRightEmpty",
			cql: "EndsWith('apple','')",
			wantResult: newOrFatal(t,true),
		},
		{
			name: "EndsWithBothEmpty",
			cql: "EndsWith('','')",
			wantResult: newOrFatal(t,true),
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

func TestLengthString(t *testing.T) {
  tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
    {
      name: "Length('ABC') = 3",
      cql: "Length('ABC')",
      wantModel: &model.Length{
        UnaryExpression: &model.UnaryExpression{
          Operand: model.NewLiteral("ABC", types.String),
          Expression: model.ResultType(types.Integer),
        },
      },
      wantResult: newOrFatal(t, 3),
    },
    {
      name: "LengthNullasString",
      cql: "Length(null as String)",
      wantResult: newOrFatal(t, nil),
    },
    {
      name: "LengthBig",
      cql: "Length('How is the weather today')",
      wantResult: newOrFatal(t, 24),
    },
    {
      name: "LengthEmpty",
      cql: "Length('')",
      wantResult: newOrFatal(t,0),
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

func TestLastPositionOf(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "LastPositionOfFound",
			cql: "LastPositionOf('B','ABC')",
			wantModel: &model.LastPositionOf{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("B", types.String),
						model.NewLiteral("ABC", types.String),
					},
					Expression: model.ResultType(types.Integer),
				},
			},
			wantResult: newOrFatal(t, 1),
		},
		{
			name: "LastPositionOfNotFound",
			cql: "LastPositionOf('X', 'ABC')",
			wantResult: newOrFatal(t, -1),
		},
		{
			name: "LastPositionOfFound2",
			cql: "LastPositionOf('B', 'ABCDEDCBA')",
			wantResult: newOrFatal(t, 7),
		},
		{
			name: "LastPositionOfLeftNull",
			cql: "LastPositionOf(null, 'ABC')",
			wantResult: newOrFatal(t, nil),
		},
		{
			name: "LastPositionOfRightNull",
			cql: "LastPositionOf('A', null)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name: "LastPositionOfLong",
			cql: "LastPositionOf('abra','abracadabra')",
			wantResult: newOrFatal(t,7),
		},
		{
			name: "LastPositionEmptyLeft",
			cql: "LastPositionOf('','ABC')",
			wantResult: newOrFatal(t, 3),
		},
		{
			name: "LastPositionOfEmptyRight",
			cql: "LastPositionOf('A','')",
			wantResult: newOrFatal(t, -1),
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

func TestUpper(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Upper",
			cql: "Upper('abc')",
			wantModel: &model.Upper{
				UnaryExpression: &model.UnaryExpression{
					Operand: model.NewLiteral("abc", types.String),
					Expression: model.ResultType(types.String),
				},
			},
				wantResult: newOrFatal(t, "ABC"),
		},
		{
			name: "UpperAlready",
			cql: "Upper('ABC')",
			wantResult: newOrFatal(t, "ABC"),
		},
		{
			name: "UpperNil",
			cql: "Upper(null)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name: "UpperEmpty",
			cql: "Upper('')",
			wantResult: newOrFatal(t, ""),
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