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
