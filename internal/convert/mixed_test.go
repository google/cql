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

package convert

import (
	"strings"
	"testing"

	"github.com/google/cql/model"
	"github.com/google/cql/types"
	"github.com/google/go-cmp/cmp"
)

func TestInferMixedType(t *testing.T) {
	tests := []struct {
		name    string
		invoked []model.IExpression
		want    Infered
	}{
		{
			name: "Multiple Operands Compatible",
			invoked: []model.IExpression{
				model.NewLiteral("4", types.Integer),
				model.NewLiteral("4.4", types.Decimal),
				model.NewLiteral("5", types.Long)},
			want: Infered{
				PuntedToChoice: false,
				UniformType:    types.Decimal,
				WrappedOperands: []model.IExpression{
					&model.ToDecimal{
						UnaryExpression: &model.UnaryExpression{
							Operand:    model.NewLiteral("4", types.Integer),
							Expression: model.ResultType(types.Decimal),
						},
					},
					model.NewLiteral("4.4", types.Decimal),
					&model.ToDecimal{
						UnaryExpression: &model.UnaryExpression{
							Operand:    model.NewLiteral("5", types.Long),
							Expression: model.ResultType(types.Decimal),
						},
					},
				},
			},
		},
		{
			name: "Multiple Operands Compatible Same Type",
			invoked: []model.IExpression{
				model.NewLiteral("4", types.Integer),
				model.NewLiteral("5", types.Integer)},
			want: Infered{
				PuntedToChoice: false,
				UniformType:    types.Integer,
				WrappedOperands: []model.IExpression{
					model.NewLiteral("4", types.Integer),
					model.NewLiteral("5", types.Integer),
				},
			},
		},
		{
			name: "Any triggers Compatible not Subtype",
			invoked: []model.IExpression{
				model.NewLiteral("4", types.Integer),
				model.NewLiteral("null", types.Any)},
			want: Infered{
				PuntedToChoice: false,
				UniformType:    types.Integer,
				WrappedOperands: []model.IExpression{
					model.NewLiteral("4", types.Integer),
					&model.As{
						UnaryExpression: &model.UnaryExpression{
							Operand:    model.NewLiteral("null", types.Any),
							Expression: model.ResultType(types.Integer),
						},
						AsTypeSpecifier: types.Integer,
						Strict:          false,
					},
				},
			},
		},
		{
			name: "All Any",
			invoked: []model.IExpression{
				model.NewLiteral("null", types.Any),
				model.NewLiteral("null", types.Any),
			},
			want: Infered{
				PuntedToChoice: false,
				UniformType:    types.Any,
				WrappedOperands: []model.IExpression{
					model.NewLiteral("null", types.Any),
					model.NewLiteral("null", types.Any),
				},
			},
		},
		{
			name: "Single Operand Compatible",
			invoked: []model.IExpression{
				model.NewLiteral("string", types.String),
			},
			want: Infered{
				PuntedToChoice: false,
				UniformType:    types.String,
				WrappedOperands: []model.IExpression{
					model.NewLiteral("string", types.String),
				},
			},
		},
		{
			name:    "No Operands",
			invoked: []model.IExpression{},
			want: Infered{
				PuntedToChoice:  false,
				UniformType:     types.Any,
				WrappedOperands: []model.IExpression{},
			},
		},
		{
			name: "Multiple Operands Incompatible",
			invoked: []model.IExpression{
				model.NewLiteral("String", types.String),
				// This tests that it is Choice<String, List<Integer>> not Choice<String, String, List<Integer>>
				model.NewLiteral("String2", types.String),
				model.NewList([]string{"4", "5"}, types.Integer),
			},
			want: Infered{
				PuntedToChoice: true,
				UniformType:    &types.Choice{ChoiceTypes: []types.IType{types.String, &types.List{ElementType: types.Integer}}},
				WrappedOperands: []model.IExpression{
					&model.As{
						UnaryExpression: &model.UnaryExpression{
							Operand:    model.NewLiteral("String", types.String),
							Expression: model.ResultType(&types.Choice{ChoiceTypes: []types.IType{types.String, &types.List{ElementType: types.Integer}}}),
						},
						AsTypeSpecifier: &types.Choice{ChoiceTypes: []types.IType{types.String, &types.List{ElementType: types.Integer}}},
						Strict:          false,
					},
					&model.As{
						UnaryExpression: &model.UnaryExpression{
							Operand:    model.NewLiteral("String2", types.String),
							Expression: model.ResultType(&types.Choice{ChoiceTypes: []types.IType{types.String, &types.List{ElementType: types.Integer}}}),
						},
						AsTypeSpecifier: &types.Choice{ChoiceTypes: []types.IType{types.String, &types.List{ElementType: types.Integer}}},
						Strict:          false,
					},
					&model.As{
						UnaryExpression: &model.UnaryExpression{
							Operand:    model.NewList([]string{"4", "5"}, types.Integer),
							Expression: model.ResultType(&types.Choice{ChoiceTypes: []types.IType{types.String, &types.List{ElementType: types.Integer}}}),
						},
						AsTypeSpecifier: &types.Choice{ChoiceTypes: []types.IType{types.String, &types.List{ElementType: types.Integer}}},
						Strict:          false,
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			modelinfo := newFHIRModelInfo(t)
			got, err := InferMixed(tc.invoked, modelinfo)
			if err != nil {
				t.Fatalf("InferMixed() unexpected err: %v", err)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("InferMixed() diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDeDuplicate(t *testing.T) {
	tests := []struct {
		name  string
		input []types.IType
		want  types.IType
	}{
		{
			name:  "Duplicates removed",
			input: []types.IType{types.Integer, types.String, types.Integer},
			want:  &types.Choice{ChoiceTypes: []types.IType{types.Integer, types.String}},
		},
		{
			name:  "Choice flattened with duplicates removed",
			input: []types.IType{types.Integer, types.String, &types.Choice{ChoiceTypes: []types.IType{types.Integer, &types.List{ElementType: types.Integer}, &types.Choice{ChoiceTypes: []types.IType{types.Quantity}}}}},
			want:  &types.Choice{ChoiceTypes: []types.IType{types.Integer, types.String, &types.List{ElementType: types.Integer}, types.Quantity}},
		},
		{
			name:  "Choice is not needed",
			input: []types.IType{types.Integer, types.Integer},
			want:  types.Integer,
		},
		{
			name:  "No implicit conversions applied",
			input: []types.IType{types.Decimal, types.Integer},
			want:  &types.Choice{ChoiceTypes: []types.IType{types.Decimal, types.Integer}},
		},
	}

	for _, tc := range tests {
		got, err := DeDuplicate(tc.input)
		if err != nil {
			t.Errorf("DeDuplicate(%v) returned an unexpected error: %v", tc.input, err)
			continue
		}

		if diff := cmp.Diff(tc.want, got); diff != "" {
			t.Errorf("DeDuplicate(%v) returned an unexpected diff (-want +got): %v", tc.input, diff)
		}
	}
}

func TestIntersect(t *testing.T) {
	tests := []struct {
		name  string
		left  types.IType
		right types.IType
		want  types.IType
	}{
		{
			name:  "Equal types",
			left:  types.Integer,
			right: types.Integer,
			want:  types.Integer,
		},
		{
			name:  "Choice flattened",
			left:  &types.Choice{ChoiceTypes: []types.IType{types.Integer, types.String, &types.List{ElementType: types.Integer}, types.Quantity}},
			right: &types.Choice{ChoiceTypes: []types.IType{types.Integer, &types.List{ElementType: types.Integer}, &types.Choice{ChoiceTypes: []types.IType{types.Quantity}}}},
			want:  &types.Choice{ChoiceTypes: []types.IType{types.Integer, &types.List{ElementType: types.Integer}, types.Quantity}},
		},
		{
			name:  "Choice is not needed",
			left:  &types.Choice{ChoiceTypes: []types.IType{types.Integer}},
			right: types.Integer,
			want:  types.Integer,
		},
	}

	for _, tc := range tests {
		got, err := Intersect(tc.left, tc.right)
		if err != nil {
			t.Errorf("Intersect(%v, %v) returned an unexpected error: %v", tc.left, tc.right, err)
			continue
		}

		if diff := cmp.Diff(tc.want, got); diff != "" {
			t.Errorf("Intersect(%v, %v) returned an unexpected diff (-want +got): %v", tc.left, tc.right, diff)
		}
	}
}

func TestIntersect_Errors(t *testing.T) {
	tests := []struct {
		name    string
		left    types.IType
		right   types.IType
		wantErr string
	}{
		{
			name:    "No common types",
			left:    types.Integer,
			right:   types.String,
			wantErr: "no common types between",
		},
	}

	for _, tc := range tests {
		_, gotErr := Intersect(tc.left, tc.right)
		if gotErr == nil {
			t.Fatalf("Intersect succeeded, but expected an error")
		}
		if !strings.Contains(gotErr.Error(), tc.wantErr) {
			t.Errorf("Unexpected evaluation error contents. got (%v) want (%v)", gotErr.Error(), tc.wantErr)
		}
	}
}
