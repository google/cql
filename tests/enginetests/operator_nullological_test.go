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

func TestCoalesce(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Coalesce(null, 1)",
			cql:  "Coalesce(null, 1)",
			wantModel: &model.Coalesce{
				NaryExpression: &model.NaryExpression{
					Operands: []model.IExpression{
						&model.As{
							UnaryExpression: &model.UnaryExpression{
								Expression: model.ResultType(types.Integer),
								Operand:    model.NewLiteral("null", types.Any),
							},
							AsTypeSpecifier: types.Integer,
						},
						model.NewLiteral("1", types.Integer),
					},
					Expression: model.ResultType(types.Integer),
				},
			},
			wantResult: newOrFatal(t, 1),
		},
		{
			name:       "Coalesce(1, 2)",
			cql:        "Coalesce(1, 2)",
			wantResult: newOrFatal(t, 1),
		},
		{
			name:       "Coalesce(null, null)",
			cql:        "Coalesce(null, null)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Coalesce(null, null, 2)",
			cql:        "Coalesce(null, null, 2)",
			wantResult: newOrFatal(t, 2),
		},
		{
			name:       "Coalesce(null, null, null)",
			cql:        "Coalesce(null, null, null)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Coalesce(null, null, null, 2)",
			cql:        "Coalesce(null, null, null, 2)",
			wantResult: newOrFatal(t, 2),
		},
		{
			name:       "Coalesce(null, null, null, null, null)",
			cql:        "Coalesce(null, null, null, null, null)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Coalesce(null, null, null, null, 2)",
			cql:        "Coalesce(null, null, null, null, 2)",
			wantResult: newOrFatal(t, 2),
		},
		{
			name:       "Coalesce(null, null, null, null, null)",
			cql:        "Coalesce(null, null, null, null, null)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Coalesce({})",
			cql:        "Coalesce({})",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Coalesce({null, 1})",
			cql:        "Coalesce({null, 1})",
			wantResult: newOrFatal(t, 1),
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
				t.Errorf("Parse Expression diff (-want +got):\n%s", diff)
			}

			results, err := interpreter.Eval(context.Background(), parsedLibs, defaultInterpreterConfig(t, p))
			if err != nil {
				t.Fatalf("Eval returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantResult, getTESTRESULT(t, results), protocmp.Transform()); diff != "" {
				t.Errorf("Evaluate Expression returned diff (-want +got)\n%v", diff)
			}
		})
	}
}

func TestIsNullogical(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "1 is null",
			cql:  "1 is null",
			wantModel: &model.IsNull{
				UnaryExpression: &model.UnaryExpression{
					Operand:    model.NewLiteral("1", types.Integer),
					Expression: model.ResultType(types.Boolean),
				},
			},
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "null is null",
			cql:        "null is null",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "IsNull(null)",
			cql:        "IsNull(null)",
			wantResult: newOrFatal(t, true),
		},
		{
			name: "true is true",
			cql:  "true is true",
			wantModel: &model.IsTrue{
				UnaryExpression: &model.UnaryExpression{
					Operand:    model.NewLiteral("true", types.Boolean),
					Expression: model.ResultType(types.Boolean),
				},
			},
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "false is true",
			cql:        "false is true",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "null is true",
			cql:        "null is true",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "IsTrue(true)",
			cql:        "IsTrue(true)",
			wantResult: newOrFatal(t, true),
		},
		{
			name: "true is false",
			cql:  "true is false",
			wantModel: &model.IsFalse{
				UnaryExpression: &model.UnaryExpression{
					Operand:    model.NewLiteral("true", types.Boolean),
					Expression: model.ResultType(types.Boolean),
				},
			},
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "false is false",
			cql:        "false is false",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "null is false",
			cql:        "null is false",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "IsFalse(false)",
			cql:        "IsFalse(false)",
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
