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

func TestEvalIfConditional_Error(t *testing.T) {
	tests := []struct {
		name    string
		model   model.IExpression
		wantErr string
	}{
		{
			name: "non boolean condition",
			model: &model.IfThenElse{
				Condition:  model.NewLiteral("2", types.Integer),
				Then:       model.NewLiteral("2", types.Integer),
				Else:       model.NewLiteral("3", types.Integer),
				Expression: model.ResultType(types.Integer),
			},
			wantErr: "cannot convert System.Integer to a boolean",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Eval(context.Background(), []*model.Library{wrapInLib(t, tc.model)}, defaultInterpreterConfig(t))
			if err == nil {
				t.Errorf("Eval succeeded, wanted error")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("Returned error (%s) did not contain expected string (%s)", err, tc.wantErr)
			}
		})
	}
}

func TestCase_Error(t *testing.T) {
	tests := []struct {
		name    string
		expr    model.IExpression
		wantErr string
	}{
		{
			name: "No Comparand When Not Boolean",
			expr: &model.Case{
				Comparand: nil,
				CaseItem: []*model.CaseItem{
					&model.CaseItem{
						When: model.NewLiteral("5", types.Integer),
						Then: model.NewLiteral("6", types.Integer),
					},
				},
				Else:       model.NewLiteral("7", types.Integer),
				Expression: model.ResultType(types.Integer),
			},
			wantErr: "internal error - cannot convert System.Integer to a boolean",
		},
		{
			name: "Comparand Different Type Then When",
			expr: &model.Case{
				Comparand: model.NewLiteral("Apple", types.String),
				CaseItem: []*model.CaseItem{
					&model.CaseItem{
						When: model.NewLiteral("5", types.Integer),
						Then: model.NewLiteral("6", types.Integer),
					},
				},
				Else:       model.NewLiteral("7", types.Integer),
				Expression: model.ResultType(types.Integer),
			},
			wantErr: "internal error - in case expressions the comparand",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Eval(context.Background(), []*model.Library{wrapInLib(t, test.expr)}, defaultInterpreterConfig(t))
			if err == nil {
				t.Errorf("Eval succeeded, wanted error")
			}
			if !strings.Contains(err.Error(), test.wantErr) {
				t.Errorf("Returned error (%s) did not contain expected string (%s)", err, test.wantErr)
			}
		})
	}
}
