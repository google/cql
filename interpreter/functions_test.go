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

package interpreter

import (
	"context"
	"strings"
	"testing"

	"github.com/google/cql/model"
	"github.com/google/cql/types"
)

// Keep: These functions produce parse errors if ran via enginetests.
func TestFailingMultipleLibraries(t *testing.T) {
	tests := []struct {
		name        string
		tree        *model.Library
		errContains string
	}{
		{
			name: "Local FunctionRef Not Found",
			tree: &model.Library{
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:       "Param",
							Context:    "Patient",
							Expression: &model.FunctionRef{Name: "Non existent", Operands: []model.IExpression{}},
						},
					},
				},
			},
			errContains: "could not resolve",
		},
		{
			name: "OperandRef Not Found",
			tree: &model.Library{
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.FunctionDef{
							ExpressionDef: &model.ExpressionDef{
								Name:    "FuncName",
								Context: "Patient",
								Expression: &model.OperandRef{
									Name:       "B",
									Expression: &model.Expression{Element: &model.Element{ResultType: types.Integer}},
								},
								Element: &model.Element{ResultType: types.Integer},
							},
							Operands: []model.OperandDef{model.OperandDef{Name: "A", Expression: &model.Expression{Element: &model.Element{ResultType: types.Integer}}}},
						},
						&model.ExpressionDef{
							Name:    "Param",
							Context: "Patient",
							Expression: &model.FunctionRef{
								Name:     "FuncName",
								Operands: []model.IExpression{&model.Literal{Value: "4", Expression: &model.Expression{Element: &model.Element{ResultType: types.Integer}}}},
							},
						},
					},
				},
			},
			errContains: "could not resolve",
		},
		{
			name: "Global FunctionRef Private Not Found",
			tree: &model.Library{
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:    "Param",
							Context: "Patient",
							Expression: &model.FunctionRef{
								Name:        "private func",
								LibraryName: "helpers",
								Operands:    []model.IExpression{&model.Literal{Value: "4", Expression: &model.Expression{Element: &model.Element{ResultType: types.Integer}}}}},
						},
					},
				},
			},
			errContains: "could not resolve",
		},
		{
			name: "Global FunctionRef Library Name Not Found",
			tree: &model.Library{
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:       "Param",
							Context:    "Patient",
							Expression: &model.FunctionRef{Name: "private func", LibraryName: "Non existent", Operands: []model.IExpression{}},
						},
					},
				},
			},
			errContains: "could not resolve",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Eval(context.Background(), []*model.Library{test.tree, helperLib(t)}, defaultInterpreterConfig(t))
			if err == nil {
				t.Errorf("Eval Library(%s) = nil, want error", test.name)
			}
			if !strings.Contains(err.Error(), test.errContains) {
				t.Errorf("Returned error (%s) did not contain expected string (%s)", err, test.errContains)
			}
		})
	}
}
