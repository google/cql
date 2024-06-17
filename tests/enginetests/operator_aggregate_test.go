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
