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
	"testing"

	"github.com/google/cql/interpreter"
	"github.com/google/cql/model"
	"github.com/google/cql/parser"
	"github.com/google/cql/result"
	"github.com/google/cql/types"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestLogicOperators(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		// And
		{
			name: "true and true",
			cql:  "true and true",
			wantModel: &model.And{
				BinaryExpression: &model.BinaryExpression{
					Expression: model.ResultType(types.Boolean),
					Operands: []model.IExpression{
						model.NewLiteral("true", types.Boolean),
						model.NewLiteral("true", types.Boolean),
					},
				},
			},
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "true and false",
			cql:        "true and false",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "true and null",
			cql:        "true and null",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "null and true",
			cql:        "null and true",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "false and true",
			cql:        "false and true",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "false and false",
			cql:        "false and false",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "false and null",
			cql:        "false and null",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "null and false",
			cql:        "null and false",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "null and null",
			cql:        "null and null",
			wantResult: newOrFatal(t, nil),
		},
		// Or
		{
			name: "true or true",
			cql:  "true or true",
			wantModel: &model.Or{
				BinaryExpression: &model.BinaryExpression{
					Expression: model.ResultType(types.Boolean),
					Operands: []model.IExpression{
						model.NewLiteral("true", types.Boolean),
						model.NewLiteral("true", types.Boolean),
					},
				},
			},
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "true or false",
			cql:        "true or false",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "true or null",
			cql:        "true or null",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "null or true",
			cql:        "null or true",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "false or true",
			cql:        "false or true",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "false or false",
			cql:        "false or false",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "false or null",
			cql:        "false or null",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "null or false",
			cql:        "null or false",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "null or null",
			cql:        "null or null",
			wantResult: newOrFatal(t, nil),
		},
		// Xor
		{
			name: "true xor true",
			cql:  "true xor true",
			wantModel: &model.XOr{
				BinaryExpression: &model.BinaryExpression{
					Expression: model.ResultType(types.Boolean),
					Operands: []model.IExpression{
						model.NewLiteral("true", types.Boolean),
						model.NewLiteral("true", types.Boolean),
					},
				},
			},
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "true xor false",
			cql:        "true xor false",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "true xor null",
			cql:        "true xor null",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "null xor true",
			cql:        "null xor true",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "false xor true",
			cql:        "false xor true",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "false xor false",
			cql:        "false xor false",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "false xor null",
			cql:        "false xor null",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "null xor false",
			cql:        "null xor false",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "null xor null",
			cql:        "null xor null",
			wantResult: newOrFatal(t, nil),
		},
		// Implies
		{
			name: "true implies true",
			cql:  "true implies true",
			wantModel: &model.Implies{
				BinaryExpression: &model.BinaryExpression{
					Expression: model.ResultType(types.Boolean),
					Operands: []model.IExpression{
						model.NewLiteral("true", types.Boolean),
						model.NewLiteral("true", types.Boolean),
					},
				},
			},
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "true implies false",
			cql:        "true implies false",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "true implies null",
			cql:        "true implies null",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "null implies true",
			cql:        "null implies true",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "false implies true",
			cql:        "false implies true",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "false implies false",
			cql:        "false implies false",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "false implies null",
			cql:        "false implies null",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "null implies false",
			cql:        "null implies false",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "null implies null",
			cql:        "null implies null",
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

func TestNot(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "not true",
			cql:  "not true",
			wantModel: &model.Not{
				UnaryExpression: &model.UnaryExpression{
					Operand:    model.NewLiteral("true", types.Boolean),
					Expression: model.ResultType(types.Boolean),
				},
			},
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "not false",
			cql:        "not false",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "not null",
			cql:        "not null",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Not(true) functional form",
			cql:        "Not(true)",
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
