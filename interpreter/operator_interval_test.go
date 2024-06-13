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

func TestIntervalOperatorIn_Error(t *testing.T) {
	tests := []struct {
		name    string
		expr    model.IExpression
		wantErr string
	}{
		{
			name: "Invalid Precision Date",
			expr: &model.In{
				BinaryExpression: &model.BinaryExpression{
					Expression: model.ResultType(types.Boolean),
					Operands: []model.IExpression{
						model.NewLiteral("@2020-03", types.Date),
						model.NewInclusiveInterval("@2020-03-25", "@2020-04", types.Date),
					},
				},
				Precision: model.SECOND,
			},
			wantErr: "precision must be one of",
		},
		{
			name: "Invalid Precision DateTime",
			expr: &model.In{
				BinaryExpression: &model.BinaryExpression{
					Expression: model.ResultType(types.Boolean),
					Operands: []model.IExpression{
						model.NewLiteral("@2024-02-29T01:20:30.101-07:00", types.DateTime),
						model.NewInclusiveInterval("@2024-02-29T01:20:30.101-07:00", "@2024-04-29T01:20:30.101-07:00", types.DateTime),
					},
				},
				Precision: model.WEEK,
			},
			wantErr: "precision must be one of",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Eval(context.Background(), []*model.Library{wrapInLib(t, test.expr)}, defaultInterpreterConfig(t))
			if err == nil {
				t.Errorf("TestIntervalOperatorIn_Error() call to evalLibrary() succeeded, wanted error")
			}
			if !strings.Contains(err.Error(), test.wantErr) {
				t.Errorf("TestIntervalOperatorIn_Error() returned unexpected error: %v want: %s", err, test.wantErr)
			}
		})
	}
}
