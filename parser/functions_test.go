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

package parser

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/cql/model"
	"github.com/google/cql/types"
	"github.com/google/go-cmp/cmp"
	"github.com/lithammer/dedent"
)

func TestFunctionSingleLibrary(t *testing.T) {
	tests := []struct {
		name string
		desc string
		cql  string
		want *model.Library
	}{
		{
			name: "FunctionDef",
			cql: dedent.Dedent(`
			define function "Population"(P Integer):  4
			`),
			want: &model.Library{
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.FunctionDef{
							ExpressionDef: &model.ExpressionDef{
								Name:        "Population",
								AccessLevel: "PUBLIC",
								Expression:  model.NewLiteral("4", types.Integer),
								Element:     &model.Element{ResultType: types.Integer},
							},
							Operands: []model.OperandDef{{Name: "P", Expression: model.ResultType(types.Integer)}},
						},
					},
				},
			},
		},
		{
			name: "FunctionDef with OperandRef",
			cql: dedent.Dedent(`
			define function P(P Integer):  P
			`),
			want: &model.Library{
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.FunctionDef{
							ExpressionDef: &model.ExpressionDef{
								Name:        "P",
								AccessLevel: "PUBLIC",
								Expression:  &model.OperandRef{Name: "P", Expression: model.ResultType(types.Integer)},
								Element:     &model.Element{ResultType: types.Integer},
							},
							Operands: []model.OperandDef{{Name: "P", Expression: model.ResultType(types.Integer)}},
						},
					},
				},
			},
		},
		{
			name: "FunctionDef return",
			cql: dedent.Dedent(`
			define function "Population"() returns Integer:  4
			`),
			want: &model.Library{
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.FunctionDef{
							ExpressionDef: &model.ExpressionDef{
								Name:        "Population",
								AccessLevel: "PUBLIC",
								Expression:  model.NewLiteral("4", types.Integer),
								Element:     &model.Element{ResultType: types.Integer},
							},
							Operands: []model.OperandDef{},
						},
					},
				},
			},
		},
		{
			name: "FunctionDef fluent with access modifier",
			cql: dedent.Dedent(`
			define private fluent function "Population"(): external
			`),
			want: &model.Library{
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.FunctionDef{
							ExpressionDef: &model.ExpressionDef{
								Name:        "Population",
								AccessLevel: "PRIVATE",
								Element:     &model.Element{},
							},
							Operands: []model.OperandDef{},
							External: true,
							Fluent:   true,
						},
					},
				},
			},
		},
		{
			name: "FunctionDef fluent without access modifier",
			cql: dedent.Dedent(`
			define fluent function "Population"(): external
			`),
			want: &model.Library{
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.FunctionDef{
							ExpressionDef: &model.ExpressionDef{
								Name:        "Population",
								AccessLevel: "PUBLIC",
								Element:     &model.Element{},
							},
							Operands: []model.OperandDef{},
							External: true,
							Fluent:   true,
						},
					},
				},
			},
		},
		{
			name: "FunctionDef returns and external",
			cql: dedent.Dedent(`
			define public function "Population"() returns String : external
			`),
			want: &model.Library{
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.FunctionDef{
							ExpressionDef: &model.ExpressionDef{
								Name:        "Population",
								AccessLevel: "PUBLIC",
								Element:     &model.Element{ResultType: types.String},
							},
							Operands: []model.OperandDef{},
							External: true,
						},
					},
				},
			},
		},
		{
			name: "FunctionRef Local",
			cql: dedent.Dedent(`
			define function "Population"(P Integer):  4
			define x: Population(5)
			`),
			want: &model.Library{
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.FunctionDef{
							ExpressionDef: &model.ExpressionDef{
								Name:        "Population",
								AccessLevel: "PUBLIC",
								Expression:  model.NewLiteral("4", types.Integer),
								Element:     &model.Element{ResultType: types.Integer},
							},
							Operands: []model.OperandDef{{Name: "P", Expression: model.ResultType(types.Integer)}},
						},
						&model.ExpressionDef{
							Name: "x",
							Expression: &model.FunctionRef{
								Name: "Population",
								Operands: []model.IExpression{
									model.NewLiteral("5", types.Integer),
								},
								Expression: model.ResultType(types.Integer),
							},
							AccessLevel: "PUBLIC",
							Element:     &model.Element{ResultType: types.Integer},
						},
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parsedLibs, err := newFHIRParser(t).Libraries(context.Background(), []string{test.cql}, Config{})
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(test.want, parsedLibs[0]); diff != "" {
				t.Errorf("%v\nParsing diff (-want +got):\n%s", test.desc, diff)
			}
		})
	}
}

func TestFunctionMultipleLibraries(t *testing.T) {
	tests := []struct {
		name string
		cql  string
		want *model.Library
	}{
		{
			name: "QualifiedFunction Global Reference",
			cql: dedent.Dedent(`
				library measure version '1.0'
        include example.helpers version '1.0' called Helpers
				define X: Helpers."public func"(5)`),
			want: &model.Library{
				Identifier: &model.LibraryIdentifier{Local: "measure", Qualified: "measure", Version: "1.0"},
				Includes: []*model.Include{
					{
						Identifier: &model.LibraryIdentifier{
							Local:     "Helpers",
							Qualified: "example.helpers",
							Version:   "1.0",
						},
					},
				},
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:        "X",
							AccessLevel: "PUBLIC",
							Expression: &model.FunctionRef{
								Name:        "public func",
								LibraryName: "Helpers",
								Operands: []model.IExpression{
									model.NewLiteral("5", types.Integer),
								},
								Expression: model.ResultType(types.Integer)},
							Element: &model.Element{ResultType: types.Integer},
						},
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cqlLibs := []string{
				dedent.Dedent(`
					library example.helpers version '1.0'
					define public function "public func"(A Integer): 2
					define private function "private func"(A Integer): 3 `),
				test.cql,
			}
			parsedLibs, err := newFHIRParser(t).Libraries(context.Background(), cqlLibs, Config{})
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(test.want, parsedLibs[1]); diff != "" {
				t.Errorf("Parsing diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMalformedFunctionSingleLibrary(t *testing.T) {
	tests := []struct {
		name        string
		cql         string
		errContains []string
		errCount    int
	}{
		{
			name: "FunctionDef_ReturnTypeMismatch",
			cql: dedent.Dedent(`
			define function "Population"() returns String: 4
			`),
			errContains: []string{"function body return type"},
			errCount:    1,
		},
		{
			name: "FunctionDef Already Exists",
			cql: dedent.Dedent(`
			define function A(): 4
			define function A(): 5
			`),
			errContains: []string{"function A() already exists"},
			errCount:    1,
		},
		{
			name: "Matching function is not fluent",
			cql: dedent.Dedent(`
			define function Foo(a Integer): a
			define Bar: 5.Foo()
			`),
			errContains: []string{"could not resolve Foo(System.Integer): no matching overloads (may not be a fluent function)"},
			errCount:    1,
		},
		{
			name: "FunctionRef Local Does Not Exist",
			cql: dedent.Dedent(`
			define X: P()
			`),
			errContains: []string{"could not resolve"},
			errCount:    1,
		},
		{
			name: "OperandDef Same Name",
			cql: dedent.Dedent(`
			define function "Population"(A Integer, A String):  4
			`),
			errContains: []string{"alias A already exists"},
			errCount:    1,
		},
		{
			name: "OperandDef and ExpressionDef Same Name",
			cql: dedent.Dedent(`
			define A: 5
			define function "Population"(A Integer):  4
			`),
			errContains: []string{"identifier A already exists"},
			errCount:    1,
		},
		{
			name: "OperandRef Does Not Exist",
			cql: dedent.Dedent(`
			define function "Population"(A Integer):  B
			`),
			errContains: []string{"could not resolve the local"},
			errCount:    1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := newFHIRParser(t).Libraries(context.Background(), []string{test.cql}, Config{})
			if err == nil {
				t.Fatal("Parsing succeeded, expected error")
			}

			var pe *LibraryErrors
			if ok := errors.As(err, &pe); ok {
				for _, ec := range test.errContains {
					if !strings.Contains(pe.Error(), ec) {
						t.Errorf("Returned error (%s) did not contain expected string (%s)",
							err.Error(), test.errContains)
					}
				}

				if len(pe.Errors) != test.errCount {
					t.Errorf("Returned error (%s) had (%d) errors but expected (%d)",
						err.Error(), len(pe.Errors), test.errCount)
				}
			} else {
				t.Errorf("Unexpected test error (%s).", err.Error())
			}
		})
	}
}

func TestMalformedFunctionMultipleLibraries(t *testing.T) {
	tests := []struct {
		name        string
		cql         string
		errContains []string
		errCount    int
	}{
		{
			name: "QualifiedFunction cannot resolve function name",
			cql: dedent.Dedent(`
				library measure version '1.0'
        include example.helpers version '1.0' called Helpers
				define X: Helpers."private func"(5)`),
			errContains: []string{"could not resolve"},
			errCount:    1,
		},
		{
			name: "QualifiedFunction cannot resolve library name",
			cql: dedent.Dedent(`
			library measure version '1.0'
			include example.helpers version '1.0' called Helpers
			define X: NonExistent."private func"(5)`),
			errContains: []string{
				"could not resolve the local reference to NonExistent",
				"could not resolve private func(System.Any, System.Integer)"},
			errCount: 2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cqlLibs := []string{
				dedent.Dedent(`
					library example.helpers version '1.0'
					define public function "public func"(A Integer): 2
					define private function "private func"(A Integer): 3 `),
				tc.cql,
			}
			_, err := newFHIRParser(t).Libraries(context.Background(), cqlLibs, Config{})
			var pe *LibraryErrors
			if ok := errors.As(err, &pe); ok {
				for _, ec := range tc.errContains {
					if !strings.Contains(pe.Error(), ec) {
						t.Errorf("Returned error (%s) did not contain expected string (%s)",
							err.Error(), tc.errContains)
					}
				}

				if len(pe.Errors) != tc.errCount {
					t.Errorf("Returned error (%s) had (%d) errors but expected (%d)",
						err.Error(), len(pe.Errors), tc.errCount)
				}
			} else {
				t.Errorf("Unexpected test error (%s).", err.Error())
			}
		})
	}
}
