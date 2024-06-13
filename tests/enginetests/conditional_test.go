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
	"github.com/lithammer/dedent"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestIf(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "if 1 = 1 then 2 else 3",
			cql:  "if 1 = 1 then 2 else 3",
			wantModel: &model.IfThenElse{
				Condition: &model.Equal{
					BinaryExpression: &model.BinaryExpression{
						Operands: []model.IExpression{
							model.NewLiteral("1", types.Integer),
							model.NewLiteral("1", types.Integer),
						},
						Expression: model.ResultType(types.Boolean),
					},
				},
				Then:       model.NewLiteral("2", types.Integer),
				Else:       model.NewLiteral("3", types.Integer),
				Expression: model.ResultType(types.Integer),
			},
			wantResult: newOrFatal(t, 2),
		},
		{
			name:       "if 1 = 2 then 2 else 3",
			cql:        "if 1 = 2 then 2 else 3",
			wantResult: newOrFatal(t, 3),
		},
		{
			name:       "if null then 2 else 3",
			cql:        "if null then 2 else 3",
			wantResult: newOrFatal(t, 3),
		},
		// Result expression type inference tests.
		{
			// If the else case can be implicitly converted to the then case it should be wrapped
			// in a conversion operator.
			name: "if 1 = 2 then 2.0 else 3",
			cql:  "if 1 = 2 then 2.0 else 3",
			wantModel: &model.IfThenElse{
				Condition: &model.Equal{
					BinaryExpression: &model.BinaryExpression{
						Operands: []model.IExpression{
							model.NewLiteral("1", types.Integer),
							model.NewLiteral("2", types.Integer),
						},
						Expression: model.ResultType(types.Boolean),
					},
				},
				Then: model.NewLiteral("2.0", types.Decimal),
				Else: &model.ToDecimal{
					UnaryExpression: &model.UnaryExpression{
						Operand:    model.NewLiteral("3", types.Integer),
						Expression: model.ResultType(types.Decimal),
					},
				},
				Expression: model.ResultType(types.Decimal),
			},
			wantResult: newOrFatal(t, 3.0),
		},
		{
			name: "If returns choice type",
			cql:  "if 1 = 2 then 2 else 'hi there!'",
			wantModel: &model.IfThenElse{
				Condition: &model.Equal{
					BinaryExpression: &model.BinaryExpression{
						Operands: []model.IExpression{
							model.NewLiteral("1", types.Integer),
							model.NewLiteral("2", types.Integer),
						},
						Expression: model.ResultType(types.Boolean),
					},
				},
				Then: &model.As{
					UnaryExpression: &model.UnaryExpression{
						Operand:    model.NewLiteral("2", types.Integer),
						Expression: model.ResultType(&types.Choice{ChoiceTypes: []types.IType{types.Integer, types.String}}),
					},
					AsTypeSpecifier: &types.Choice{ChoiceTypes: []types.IType{types.Integer, types.String}},
				},
				Else: &model.As{
					UnaryExpression: &model.UnaryExpression{
						Operand:    model.NewLiteral("hi there!", types.String),
						Expression: model.ResultType(&types.Choice{ChoiceTypes: []types.IType{types.Integer, types.String}}),
					},
					AsTypeSpecifier: &types.Choice{ChoiceTypes: []types.IType{types.Integer, types.String}},
				},
				Expression: model.ResultType(&types.Choice{ChoiceTypes: []types.IType{types.Integer, types.String}}),
			},
			wantResult: newOrFatal(t, "hi there!"),
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

func TestCase(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "No comparand returns then",
			cql: dedent.Dedent(`
			case
			when false then 9
			when true then 4
			else 5
			end`),
			wantModel: &model.Case{
				Comparand: nil,
				CaseItem: []*model.CaseItem{
					&model.CaseItem{
						When: model.NewLiteral("false", types.Boolean),
						Then: model.NewLiteral("9", types.Integer),
					},
					&model.CaseItem{
						When: model.NewLiteral("true", types.Boolean),
						Then: model.NewLiteral("4", types.Integer),
					},
				},
				Else:       model.NewLiteral("5", types.Integer),
				Expression: model.ResultType(types.Integer),
			},
			wantResult: newOrFatal(t, 4),
		},
		{
			name: "No comparand, returns else",
			cql: dedent.Dedent(`
			case
			when false then 9
			when false then 4
			else 5
			end`),
			wantResult: newOrFatal(t, 5),
		},
		{
			name: "No comparand, case with null is skipped",
			cql: dedent.Dedent(`
			case
			when null then 9
			when false then 4
			else 5
			end`),
			wantResult: newOrFatal(t, 5),
		},
		{
			name: "Comparand, returns then case",
			cql: dedent.Dedent(`
			case 4
			when 5 then 9
			when 4 then 6
			else 7
			end`),
			wantResult: newOrFatal(t, 6),
		},
		{
			name: "Comparand, no match returns else",
			cql: dedent.Dedent(`
			case 4
			when 5 then 9
			when 3 then 6
			else 7
			end`),
			wantResult: newOrFatal(t, 7),
		},
		{
			name: "Comparand, null case skipped",
			cql: dedent.Dedent(`
			case 5
			when null then 6
			when 5 then 9
			else 7
			end`),
			wantResult: newOrFatal(t, 9),
		},
		{
			name: "Null comparand, returns else",
			cql: dedent.Dedent(`
			case null
			when null then 6
			when 5 then 9
			else 7
			end`),
			wantResult: newOrFatal(t, 7),
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
