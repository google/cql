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

func TestClinicalOperatorCalculateAgeAt_Error(t *testing.T) {
	ageTests := []struct {
		name    string
		expr    model.IExpression
		wantErr string
	}{
		{
			name: "Invalid Precision Date",
			expr: &model.CalculateAgeAt{
				BinaryExpression: &model.BinaryExpression{
					Expression: model.ResultType(types.Date),
					Operands: []model.IExpression{
						model.NewLiteral("@1981", types.Date),
						model.NewLiteral("@2023-06-14", types.Date),
					},
				},
				Precision: model.MINUTE,
			},
			wantErr: "precision must be one of [year month week day]",
		},
		{
			name: "Invalid Precision DateTime",
			expr: &model.CalculateAgeAt{
				BinaryExpression: &model.BinaryExpression{
					Expression: model.ResultType(types.DateTime),
					Operands: []model.IExpression{
						model.NewLiteral("@1981-06-15T10:01:01.000Z", types.DateTime),
						model.NewLiteral("@2023-06-14T10:01:01.000Z", types.DateTime),
					},
				},
				Precision: model.MILLISECOND,
			},
			wantErr: "precision must be one of [year month week day hour minute second]",
		},
	}
	for _, tc := range ageTests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Eval(context.Background(), []*model.Library{wrapInLib(t, tc.expr)}, defaultInterpreterConfig(t))
			if err == nil {
				t.Errorf("TestClinicalOperatorCalculateAgeAt_Error() call to evalLibrary() succeeded, wanted error")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("TestClinicalOperatorCalculateAgeAt_Error() returned unexpected error: %v want: %s", err, tc.wantErr)
			}
		})
	}
}
