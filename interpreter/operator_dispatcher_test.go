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

// The dispatcher non error cases are tested by all operator files. This test covers top level
// errors.
func TestDispatcherError(t *testing.T) {
	tests := []struct {
		name    string
		expr    model.IExpression
		wantErr string
	}{
		{
			name: "Unary Expression Unsupported Overload",
			expr: &model.Last{
				UnaryExpression: &model.UnaryExpression{
					Expression: model.ResultType(types.Integer),
					Operand:    model.NewLiteral("4", types.Integer),
				},
			},
			wantErr: "could not resolve Last(System.Integer)",
		},
		{
			name: "Unary Expression Unsupported Overload",
			expr: &model.First{
				UnaryExpression: &model.UnaryExpression{
					Expression: model.ResultType(types.Integer),
					Operand:    model.NewLiteral("false", types.Boolean),
				},
			},
			wantErr: "could not resolve First(System.Boolean)",
		},
		{
			name: "Binary Expression Unsupported Overload",
			expr: &model.Subtract{
				BinaryExpression: &model.BinaryExpression{
					Expression: model.ResultType(types.Integer),
					Operands: []model.IExpression{
						model.NewLiteral("4", types.Integer), model.NewLiteral("Hello", types.String),
					},
				},
			},
			wantErr: "could not resolve Subtract(System.Integer, System.String)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Eval(context.Background(), []*model.Library{wrapInLib(t, test.expr)}, defaultInterpreterConfig(t))
			if err == nil {
				t.Errorf("TestDispatcherError() call to evalLibrary() succeeded, wanted error")
			}
			if !strings.Contains(err.Error(), test.wantErr) {
				t.Errorf("TestDispatcherError() returned unexpected error: %v want: %s", err, test.wantErr)
			}
		})
	}
}
