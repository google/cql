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
	"fmt"
	"strings"
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

func TestLocalFunctions(t *testing.T) {
	// TODO(b/311222838): Add source check to additional expressions to improve coverage.
	tests := []struct {
		name       string
		cql        string
		wantModels []model.IExpressionDef
		wantResult result.Value
	}{
		{
			name: "Local Function no operands",
			cql: dedent.Dedent(`
			define function FuncNoOperand(): 1
			define TESTRESULT: FuncNoOperand()`),
			wantModels: []model.IExpressionDef{
				&model.FunctionDef{
					Operands: []model.OperandDef{},
					ExpressionDef: &model.ExpressionDef{
						Name:        "FuncNoOperand",
						Expression:  model.NewLiteral("1", types.Integer),
						Context:     "Patient",
						AccessLevel: model.Public,
						Element:     &model.Element{ResultType: types.Integer},
					},
				},
				&model.ExpressionDef{
					Name: "TESTRESULT",
					Expression: &model.FunctionRef{
						Name:       "FuncNoOperand",
						Operands:   []model.IExpression{},
						Expression: model.ResultType(types.Integer),
					},
					Context:     "Patient",
					AccessLevel: model.Public,
					Element:     &model.Element{ResultType: types.Integer},
				},
			},
			wantResult: newOrFatal(t, 1),
		},
		{
			name: "Local Function with OperandRef",
			cql: dedent.Dedent(`
			define function FuncWithOperand(a Integer): a + 1
			define TESTRESULT: FuncWithOperand(1)`),
			wantModels: []model.IExpressionDef{
				&model.FunctionDef{
					Operands: []model.OperandDef{
						model.OperandDef{Name: "a", Expression: model.ResultType(types.Integer)},
					},
					ExpressionDef: &model.ExpressionDef{
						Name: "FuncWithOperand",
						Expression: &model.Add{
							BinaryExpression: &model.BinaryExpression{
								Expression: model.ResultType(types.Integer),
								Operands: []model.IExpression{
									&model.OperandRef{Name: "a", Expression: model.ResultType(types.Integer)},
									model.NewLiteral("1", types.Integer),
								},
							},
						},
						Context:     "Patient",
						AccessLevel: model.Public,
						Element:     &model.Element{ResultType: types.Integer},
					},
				},
				&model.ExpressionDef{
					Name: "TESTRESULT",
					Expression: &model.FunctionRef{
						Name:       "FuncWithOperand",
						Expression: model.ResultType(types.Integer),
						Operands:   []model.IExpression{model.NewLiteral("1", types.Integer)},
					},
					AccessLevel: model.Public,
					Context:     "Patient",
					Element:     &model.Element{ResultType: types.Integer},
				},
			},
			wantResult: newOrFatal(t, 2),
		},
		{
			name: "Fluent Function",
			cql: dedent.Dedent(`
			define fluent function FluentFunc(a Integer, b String): a + 1
			define TESTRESULT: 1.FluentFunc('apple')`),
			wantModels: []model.IExpressionDef{
				&model.FunctionDef{
					Operands: []model.OperandDef{
						model.OperandDef{Name: "a", Expression: model.ResultType(types.Integer)},
						model.OperandDef{Name: "b", Expression: model.ResultType(types.String)},
					},
					Fluent: true,
					ExpressionDef: &model.ExpressionDef{
						Name: "FluentFunc",
						Expression: &model.Add{
							BinaryExpression: &model.BinaryExpression{
								Expression: model.ResultType(types.Integer),
								Operands: []model.IExpression{
									&model.OperandRef{Name: "a", Expression: model.ResultType(types.Integer)},
									model.NewLiteral("1", types.Integer),
								},
							},
						},
						Context:     "Patient",
						AccessLevel: model.Public,
						Element:     &model.Element{ResultType: types.Integer},
					},
				},
				&model.ExpressionDef{
					Name: "TESTRESULT",
					Expression: &model.FunctionRef{
						Name:       "FluentFunc",
						Expression: model.ResultType(types.Integer),
						Operands: []model.IExpression{
							model.NewLiteral("1", types.Integer),
							model.NewLiteral("apple", types.String),
						},
					},
					AccessLevel: model.Public,
					Context:     "Patient",
					Element:     &model.Element{ResultType: types.Integer},
				},
			},
			wantResult: newOrFatal(t, 2),
		},
		{
			name: "Fluent function called on property",
			cql: dedent.Dedent(`
			define fluent function FluentFunc(a Boolean): a
			define TESTRESULT: [Patient] P return P.active.FluentFunc()`),
			wantResult: newOrFatal(t, result.List{
				Value:      []result.Value{newOrFatal(t, true)},
				StaticType: &types.List{ElementType: types.Boolean}}),
		},
		{
			name: "System operators can be called fluently",
			cql: dedent.Dedent(`
			define TESTRESULT: 4.Add(4)`),
			wantResult: newOrFatal(t, 8),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testCQL := dedent.Dedent(fmt.Sprintf(`
			library TESTLIB version '1.0.0'
			using FHIR version '4.0.1'
			include FHIRHelpers version '4.0.1'
			%v`, tc.cql))
			p := newFHIRParser(t)
			parsedLibs, err := p.Libraries(context.Background(), addFHIRHelpersLib(t, testCQL), parser.Config{})
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantModels, getTESTLIBModel(t, parsedLibs).Statements.Defs); tc.wantModels != nil && diff != "" {
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

func TestGlobalFunctions(t *testing.T) {
	tests := []struct {
		name       string
		cqlLibs    []string
		wantModels []model.IExpressionDef
		wantResult result.Value
	}{
		{
			name: "Global FunctionRef",
			cqlLibs: []string{
				dedent.Dedent(`
					library CQL_Helpers_Library version '1'
					define function PublicFunc(a Integer): a + 1
					define private function PrivateFunc(b Integer): b - 1
					`),
				dedent.Dedent(`
					library TESTLIB version '1.0.0'
					using FHIR version '4.0.1'
					include CQL_Helpers_Library version '1' called helpers
					define TESTRESULT: helpers.PublicFunc(1)`),
			},
			wantModels: []model.IExpressionDef{
				&model.ExpressionDef{
					Name: "TESTRESULT",
					Expression: &model.FunctionRef{
						Name:        "PublicFunc",
						LibraryName: "helpers",
						Expression:  model.ResultType(types.Integer),
						Operands:    []model.IExpression{model.NewLiteral("1", types.Integer)},
					},
					Context:     "Patient",
					AccessLevel: model.Public,
					Element:     &model.Element{ResultType: types.Integer},
				},
			},
			wantResult: newOrFatal(t, 2),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := newFHIRParser(t)
			parsedLibs, err := p.Libraries(context.Background(), tc.cqlLibs, parser.Config{})
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantModels, getTESTLIBModel(t, parsedLibs).Statements.Defs); tc.wantModels != nil && diff != "" {
				t.Errorf("Parse Expression diff (-want +got):\n%s", diff)
			}

			results, err := interpreter.Eval(context.Background(), parsedLibs, defaultInterpreterConfig(t, p))
			if err != nil {
				t.Fatalf("Eval returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantResult, getTESTRESULT(t, results), protocmp.Transform()); diff != "" {
				t.Errorf("Eval returned diff (-want +got)\n%v", diff)
			}

		})
	}
}

func TestFailingFunctions(t *testing.T) {
	tests := []struct {
		name                string
		cql                 string
		wantModels          []model.IExpressionDef
		wantEvalErrContains string
	}{
		{
			name: "External functions are not supported",
			cql: dedent.Dedent(`
			define function ExternalFunc(a Integer) returns Integer : external
			define ExternalFuncRef: ExternalFunc(4)`),
			wantModels: []model.IExpressionDef{
				&model.FunctionDef{
					Operands: []model.OperandDef{
						model.OperandDef{Name: "a", Expression: model.ResultType(types.Integer)},
					},
					External: true,
					ExpressionDef: &model.ExpressionDef{
						Name:        "ExternalFunc",
						Context:     "Patient",
						AccessLevel: model.Public,
						Element:     &model.Element{ResultType: types.Integer},
					},
				},
				&model.ExpressionDef{
					Name: "ExternalFuncRef",
					Expression: &model.FunctionRef{
						Name:       "ExternalFunc",
						Expression: model.ResultType(types.Integer),
						Operands: []model.IExpression{
							model.NewLiteral("4", types.Integer),
						},
					},
					Context:     "Patient",
					AccessLevel: model.Public,
					Element:     &model.Element{ResultType: types.Integer},
				},
			},
			wantEvalErrContains: "external functions are not supported",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testCQL := dedent.Dedent(fmt.Sprintf(`
			library TESTLIB version '1.0.0'
			using FHIR version '4.0.1'
			include FHIRHelpers version '4.0.1'
			%v`, tc.cql))
			p := newFHIRParser(t)
			parsedLibs, err := p.Libraries(context.Background(), addFHIRHelpersLib(t, testCQL), parser.Config{})
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantModels, getTESTLIBModel(t, parsedLibs).Statements.Defs); tc.wantModels != nil && diff != "" {
				t.Errorf("Parse Expression diff (-want +got):\n%s", diff)
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
