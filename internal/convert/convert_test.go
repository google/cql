// Copyright 2023 Google LLC
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

package convert

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/cql/internal/embeddata"
	"github.com/google/cql/internal/modelinfo"
	"github.com/google/cql/model"
	"github.com/google/cql/types"
	"github.com/google/go-cmp/cmp"
)

func TestOverloadMatch(t *testing.T) {
	tests := []struct {
		name      string
		invoked   []model.IExpression
		overloads []Overload[string]
		wantRes   MatchedOverload[string]
	}{
		{
			name: "Multiple Operands",
			invoked: []model.IExpression{
				model.NewLiteral("String", types.String),
				model.NewInclusiveInterval("@2020-03-04", "@2020-03-05", types.Date),
				model.NewList([]string{"4", "5"}, types.Integer),
			},
			overloads: []Overload[string]{
				Overload[string]{
					Result:   "Too Short",
					Operands: []types.IType{types.String, &types.Interval{PointType: types.Date}},
				},
				Overload[string]{
					Result: "Too Long",
					Operands: []types.IType{
						types.String,
						&types.Interval{PointType: types.Date},
						&types.List{ElementType: types.Integer},
						&types.Named{TypeName: "Patient"},
					},
				},
				Overload[string]{
					Result: "Just Right",
					Operands: []types.IType{
						types.String,
						&types.Interval{PointType: types.Date},
						&types.List{ElementType: types.Integer},
					},
				},
			},
			wantRes: MatchedOverload[string]{
				Result: "Just Right",
				WrappedOperands: []model.IExpression{
					model.NewLiteral("String", types.String),
					model.NewInclusiveInterval("@2020-03-04", "@2020-03-05", types.Date),
					model.NewList([]string{"4", "5"}, types.Integer),
				},
			},
		},
		{
			name: "Single Operand",
			invoked: []model.IExpression{
				model.NewLiteral("String", types.String),
			},
			overloads: []Overload[string]{
				Overload[string]{
					Result:   "Just Right",
					Operands: []types.IType{types.String},
				},
				Overload[string]{
					Result:   "Wrong Type",
					Operands: []types.IType{types.Integer},
				},
			},
			wantRes: MatchedOverload[string]{
				Result: "Just Right",
				WrappedOperands: []model.IExpression{
					model.NewLiteral("String", types.String),
				},
			},
		},
		{
			name:    "No Operands",
			invoked: []model.IExpression{},
			overloads: []Overload[string]{
				Overload[string]{
					Result:   "Too Long",
					Operands: []types.IType{types.Integer},
				},
				Overload[string]{
					Result:   "Just Right",
					Operands: []types.IType{},
				},
			},
			wantRes: MatchedOverload[string]{
				Result:          "Just Right",
				WrappedOperands: []model.IExpression{},
			},
		},
		{
			name:    "Multiple Conversions",
			invoked: []model.IExpression{model.NewLiteral("4", types.Integer), model.NewLiteral("@2020-03-05", types.Date)},
			overloads: []Overload[string]{
				Overload[string]{
					Result:   "One Simple Conversion",
					Operands: []types.IType{types.Decimal, types.Date},
				},
				Overload[string]{
					Result:   "Two Simple Conversion",
					Operands: []types.IType{types.Decimal, types.DateTime},
				},
			},
			wantRes: MatchedOverload[string]{
				Result: "One Simple Conversion",
				WrappedOperands: []model.IExpression{
					&model.ToDecimal{
						UnaryExpression: &model.UnaryExpression{
							Operand:    model.NewLiteral("4", types.Integer),
							Expression: model.ResultType(types.Decimal),
						},
					},
					model.NewLiteral("@2020-03-05", types.Date),
				},
			},
		},
		{
			name:    "Ambiguous Follwed By Exact Match",
			invoked: []model.IExpression{model.NewLiteral("4", types.Integer), model.NewLiteral("@2020-03-05", types.Date)},
			overloads: []Overload[string]{
				Overload[string]{
					Result:   "Ambiguous 1",
					Operands: []types.IType{types.Long, types.DateTime},
				},
				Overload[string]{
					Result:   "Ambiguous 2",
					Operands: []types.IType{types.Decimal, types.DateTime},
				},
				Overload[string]{
					Result:   "Exact Match",
					Operands: []types.IType{types.Integer, types.Date},
				},
			},
			wantRes: MatchedOverload[string]{
				Result: "Exact Match",
				WrappedOperands: []model.IExpression{
					model.NewLiteral("4", types.Integer),
					model.NewLiteral("@2020-03-05", types.Date),
				},
			},
		},
		{
			name: "Generics",
			invoked: []model.IExpression{
				model.NewLiteral("@2020-03-05", types.Date),
				model.NewInclusiveInterval("@2020-03-05", "@2020-03-05", types.DateTime),
				model.NewList([]string{"@2020-03-05", "@2020-03-05"}, types.DateTime),
				model.NewLiteral("Apples", types.String),
			},
			overloads: []Overload[string]{
				Overload[string]{
					Result:   "Too Short",
					Operands: []types.IType{GenericType, GenericInterval},
				},
				Overload[string]{
					Result: "Too Long",
					Operands: []types.IType{
						GenericType,
						GenericInterval,
						GenericList,
						types.String,
						types.String,
					},
				},
				Overload[string]{
					Result: "Wrong Type",
					Operands: []types.IType{
						GenericType,
						GenericInterval,
						GenericList,
						&types.Named{TypeName: "Patient"},
					},
				},
				Overload[string]{
					Result: "Just Right",
					Operands: []types.IType{
						GenericType,
						GenericInterval,
						GenericList,
						types.String,
					},
				},
			},
			wantRes: MatchedOverload[string]{
				Result: "Just Right",
				WrappedOperands: []model.IExpression{
					&model.ToDateTime{
						UnaryExpression: &model.UnaryExpression{
							Operand:    model.NewLiteral("@2020-03-05", types.Date),
							Expression: model.ResultType(types.DateTime),
						},
					},
					model.NewInclusiveInterval("@2020-03-05", "@2020-03-05", types.DateTime),
					model.NewList([]string{"@2020-03-05", "@2020-03-05"}, types.DateTime),
					model.NewLiteral("Apples", types.String),
				},
			},
		},
		// TODO(b/312172420): Add tests where generics need a list promotion and interval promotion once
		// supported in the conversion precedence.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			modelinfo := newFHIRModelInfo(t)
			gotRes, err := OverloadMatch(tc.invoked, tc.overloads, modelinfo, "Name")
			if err != nil {
				t.Fatalf("overloadMatch() unexpected err: %v", err)
			}
			if diff := cmp.Diff(tc.wantRes, gotRes); diff != "" {
				t.Errorf("overloadMatch() diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestOverloadMatch_Error(t *testing.T) {
	tests := []struct {
		name        string
		invoked     []model.IExpression
		overloads   []Overload[string]
		errContains string
		errIs       error // optional
	}{
		{
			name: "Single No Match",
			invoked: []model.IExpression{
				model.NewLiteral("String", types.String),
			},
			overloads: []Overload[string]{
				Overload[string]{
					Result:   "Wrong Type",
					Operands: []types.IType{types.Integer},
				},
				Overload[string]{
					Result:   "Too Short",
					Operands: []types.IType{},
				},
				Overload[string]{
					Result:   "Too Long",
					Operands: []types.IType{types.String, types.String},
				},
			},
			errContains: "could not resolve",
			errIs:       ErrNoMatch,
		},
		{
			name: "Multiple No Match",
			invoked: []model.IExpression{
				model.NewLiteral("String", types.String),
				model.NewInclusiveInterval("@2020-03-04", "@2020-03-05", types.Date),
				model.NewList([]string{"4", "5"}, types.Integer),
			},
			overloads: []Overload[string]{
				Overload[string]{
					Result:   "Too Short",
					Operands: []types.IType{types.String, &types.Interval{PointType: types.DateTime}},
				},
				Overload[string]{
					Result: "Too Long",
					Operands: []types.IType{
						types.String,
						&types.Interval{PointType: types.DateTime},
						&types.List{ElementType: types.Integer},
						&types.Named{TypeName: "Patient"},
					},
				},
				Overload[string]{
					Result: "Wrong Type",
					Operands: []types.IType{
						types.Integer,
						&types.Interval{PointType: types.DateTime},
						&types.List{ElementType: types.Integer},
					},
				},
			},
			errContains: "could not resolve",
			errIs:       ErrNoMatch,
		},
		{
			name:    "Ambiguous Match",
			invoked: []model.IExpression{model.NewLiteral("String", types.String), model.NewLiteral("String", types.String)},
			overloads: []Overload[string]{
				Overload[string]{
					Result:   "Ambiguous 1",
					Operands: []types.IType{types.Any, types.String},
				},
				Overload[string]{
					Result:   "Ambiguous 2",
					Operands: []types.IType{types.Any, types.String},
				},
			},
			errContains: "ambiguous",
			errIs:       ErrAmbiguousMatch,
		},
		{
			name: "Generic No Match",
			invoked: []model.IExpression{
				model.NewLiteral("String", types.String),
				model.NewInclusiveInterval("@2020-03-04", "@2020-03-05", types.Date),
				model.NewList([]string{"4", "5"}, types.Integer),
			},
			overloads: []Overload[string]{
				Overload[string]{
					Result:   "Too Short",
					Operands: []types.IType{GenericType, GenericInterval},
				},
				Overload[string]{
					Result: "Too Long",
					Operands: []types.IType{
						GenericType,
						GenericInterval,
						GenericList,
						&types.Named{TypeName: "Patient"},
					},
				},
				Overload[string]{
					Result: "No Uniform Type",
					Operands: []types.IType{
						GenericType,
						GenericInterval,
						GenericList,
					},
				},
			},
			errContains: "could not resolve",
			errIs:       ErrNoMatch,
		},
		{
			name:    "Ambiguous Generic",
			invoked: []model.IExpression{model.NewLiteral("String", types.String), model.NewLiteral("String", types.String)},
			overloads: []Overload[string]{
				Overload[string]{
					Result:   "Ambiguous 1",
					Operands: []types.IType{GenericType, types.String},
				},
				Overload[string]{
					Result:   "Ambiguous 2",
					Operands: []types.IType{types.String, GenericType},
				},
			},
			errContains: "ambiguous",
		},
		{
			name:    "Nil ResultType",
			invoked: []model.IExpression{&model.Literal{Value: "Apple"}},
			overloads: []Overload[string]{
				Overload[string]{
					Result:   "Overload",
					Operands: []types.IType{types.String},
				},
			},
			errContains: "internal error - invokedType is",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			modelinfo := newFHIRModelInfo(t)
			_, err := OverloadMatch(tc.invoked, tc.overloads, modelinfo, "Name")
			if err == nil {
				t.Fatalf("overloadMatch() did not return an error")
			}
			if !strings.Contains(err.Error(), tc.errContains) {
				t.Errorf("Returned error (%s) did not contain expected string (%s)", err.Error(), tc.errContains)
			}
			if tc.errIs != nil && !errors.Is(err, tc.errIs) {
				t.Errorf("returned error %v that is not errors.Is() the expected err: %v", err, tc.errIs)
			}
		})
	}
}

func TestOperandImplicitConverter(t *testing.T) {
	tests := []struct {
		name         string
		invokedType  types.IType
		declaredType types.IType
		want         ConvertedOperand
	}{
		{
			name:         "Exact Match",
			invokedType:  &types.Interval{PointType: types.DateTime},
			declaredType: &types.Interval{PointType: types.DateTime},
			want:         ConvertedOperand{Matched: true, Score: 0, WrappedOperand: model.NewLiteral("operand", types.String)},
		},
		{
			name:         "SubType",
			invokedType:  &types.Interval{PointType: types.DateTime},
			declaredType: types.Any,
			want:         ConvertedOperand{Matched: true, Score: 1, WrappedOperand: model.NewLiteral("operand", types.String)},
		},
		{
			name:         "Tuple SubType",
			invokedType:  &types.Tuple{ElementTypes: map[string]types.IType{"foo": types.ValueSet, "bar": types.String}},
			declaredType: &types.Tuple{ElementTypes: map[string]types.IType{"foo": types.Vocabulary, "bar": types.String}},
			want:         ConvertedOperand{Matched: true, Score: 1, WrappedOperand: model.NewLiteral("operand", types.String)},
		},
		{
			name:         "List<Tuple> SubType",
			invokedType:  &types.List{ElementType: &types.Tuple{ElementTypes: map[string]types.IType{"foo": types.ValueSet, "bar": types.String}}},
			declaredType: &types.List{ElementType: &types.Tuple{ElementTypes: map[string]types.IType{"foo": types.Vocabulary, "bar": types.String}}},
			want:         ConvertedOperand{Matched: true, Score: 1, WrappedOperand: model.NewLiteral("operand", types.String)},
		},
		{
			name:         "Compatible aka Null",
			invokedType:  types.Any,
			declaredType: types.Decimal,
			want: ConvertedOperand{
				Matched: true,
				Score:   2,
				WrappedOperand: &model.As{
					UnaryExpression: &model.UnaryExpression{
						Operand:    model.NewLiteral("operand", types.String),
						Expression: model.ResultType(types.Decimal),
					},
					AsTypeSpecifier: types.Decimal,
					Strict:          false,
				},
			},
		},
		{
			name: "Choice Cast to Concrete Type",
			invokedType: &types.Choice{
				ChoiceTypes: []types.IType{
					&types.Interval{PointType: types.DateTime},
					&types.Interval{PointType: types.Date},
				},
			},
			declaredType: &types.Interval{PointType: types.DateTime},
			want: ConvertedOperand{
				Matched: true,
				Score:   3,
				WrappedOperand: &model.As{
					UnaryExpression: &model.UnaryExpression{
						Operand:    model.NewLiteral("operand", types.String),
						Expression: model.ResultType(&types.Interval{PointType: types.DateTime}),
					},
					AsTypeSpecifier: &types.Interval{PointType: types.DateTime},
					Strict:          false,
				},
			},
		},
		{
			name:        "Concrete Type Cast to Choice",
			invokedType: &types.Interval{PointType: types.DateTime},
			declaredType: &types.Choice{
				ChoiceTypes: []types.IType{
					&types.Interval{PointType: types.DateTime},
					&types.Interval{PointType: types.Date},
				},
			},
			want: ConvertedOperand{
				Matched: true,
				Score:   3,
				WrappedOperand: &model.As{
					UnaryExpression: &model.UnaryExpression{
						Operand: model.NewLiteral("operand", types.String),
						Expression: model.ResultType(&types.Choice{
							ChoiceTypes: []types.IType{
								&types.Interval{PointType: types.DateTime},
								&types.Interval{PointType: types.Date},
							},
						}),
					},
					AsTypeSpecifier: &types.Choice{
						ChoiceTypes: []types.IType{
							&types.Interval{PointType: types.DateTime},
							&types.Interval{PointType: types.Date},
						},
					},
					Strict: false,
				},
			},
		},
		{
			name: "Choice to Choice",
			invokedType: &types.Choice{
				ChoiceTypes: []types.IType{
					types.String,
					types.Integer,
				},
			},
			declaredType: &types.Choice{
				ChoiceTypes: []types.IType{
					&types.Interval{PointType: types.Integer},
					types.Decimal,
				},
			},
			want: ConvertedOperand{
				Matched: true,
				Score:   3,
				WrappedOperand: &model.As{
					AsTypeSpecifier: &types.Choice{ChoiceTypes: []types.IType{&types.Interval{PointType: types.Integer}, types.Decimal}},
					Strict:          false,
					UnaryExpression: &model.UnaryExpression{
						Expression: model.ResultType(&types.Choice{ChoiceTypes: []types.IType{&types.Interval{PointType: types.Integer}, types.Decimal}}),
						Operand: &model.ToDecimal{
							UnaryExpression: &model.UnaryExpression{
								Expression: model.ResultType(types.Decimal),
								Operand: &model.As{
									UnaryExpression: &model.UnaryExpression{
										Operand:    model.NewLiteral("operand", types.String),
										Expression: model.ResultType(types.Integer),
									},
									AsTypeSpecifier: types.Integer,
									Strict:          false,
								},
							},
						},
					},
				},
			},
		},
		{
			name:         "FHIR ModelInfo Converison to Simple",
			invokedType:  &types.Named{TypeName: "FHIR.date"},
			declaredType: types.Date,
			want: ConvertedOperand{
				Matched: true,
				Score:   4,
				WrappedOperand: &model.FunctionRef{
					LibraryName: "FHIRHelpers",
					Name:        "ToDate",
					Expression:  model.ResultType(types.Date),
					Operands:    []model.IExpression{model.NewLiteral("operand", types.String)},
				},
			},
		},
		{
			name:         "FHIR ModelInfo Conversion to Class",
			invokedType:  &types.Named{TypeName: "FHIR.Period"},
			declaredType: &types.Interval{PointType: types.DateTime},
			want: ConvertedOperand{
				Matched: true,
				Score:   5,
				WrappedOperand: &model.FunctionRef{
					LibraryName: "FHIRHelpers",
					Name:        "ToInterval",
					Expression:  model.ResultType(&types.Interval{PointType: types.DateTime}),
					Operands:    []model.IExpression{model.NewLiteral("operand", types.String)},
				},
			},
		},
		{
			name:         "System ModelInfo Conversion to Class",
			invokedType:  types.Integer,
			declaredType: types.Quantity,
			want: ConvertedOperand{
				Matched: true,
				Score:   5,
				WrappedOperand: &model.ToQuantity{
					UnaryExpression: &model.UnaryExpression{
						Operand:    model.NewLiteral("operand", types.String),
						Expression: model.ResultType(types.Quantity),
					},
				},
			},
		},
		{
			name:         "Simple Conversion",
			invokedType:  types.Date,
			declaredType: types.DateTime,
			want: ConvertedOperand{
				Matched: true,
				Score:   4,
				WrappedOperand: &model.ToDateTime{
					UnaryExpression: &model.UnaryExpression{
						Operand:    model.NewLiteral("operand", types.String),
						Expression: model.ResultType(types.DateTime),
					},
				},
			},
		},
		{
			name:         "Interval Conversion",
			invokedType:  &types.Interval{PointType: types.Decimal},
			declaredType: &types.Interval{PointType: types.Quantity},
			want: ConvertedOperand{
				Matched: true,
				Score:   5,
				WrappedOperand: &model.Interval{
					Expression:           model.ResultType(&types.Interval{PointType: types.Quantity}),
					LowClosedExpression:  &model.Property{Source: model.NewLiteral("operand", types.String), Path: "lowClosed", Expression: model.ResultType(types.Boolean)},
					HighClosedExpression: &model.Property{Source: model.NewLiteral("operand", types.String), Path: "highClosed", Expression: model.ResultType(types.Boolean)},
					Low: &model.ToQuantity{
						UnaryExpression: &model.UnaryExpression{
							Expression: model.ResultType(types.Quantity),
							Operand: &model.Property{
								Path:       "low",
								Expression: model.ResultType(types.Decimal),
								Source:     model.NewLiteral("operand", types.String),
							},
						},
					},
					High: &model.ToQuantity{
						UnaryExpression: &model.UnaryExpression{
							Expression: model.ResultType(types.Quantity),
							Operand: &model.Property{
								Path:       "high",
								Expression: model.ResultType(types.Decimal),
								Source:     model.NewLiteral("operand", types.String),
							},
						},
					},
				},
			},
		},
		{
			name:         "List Conversion",
			invokedType:  &types.List{ElementType: types.Integer},
			declaredType: &types.List{ElementType: types.Long},
			want: ConvertedOperand{
				Matched: true,
				Score:   5,
				WrappedOperand: &model.Query{
					Source: []*model.AliasedSource{
						&model.AliasedSource{
							Alias:      "X",
							Source:     model.NewLiteral("operand", types.String),
							Expression: model.ResultType(&types.List{ElementType: types.Integer}),
						},
					},
					Return: &model.ReturnClause{
						Element: &model.Element{ResultType: types.Long},
						Expression: &model.ToLong{
							UnaryExpression: &model.UnaryExpression{
								Expression: model.ResultType(types.Long),
								Operand:    &model.AliasRef{Name: "X", Expression: model.ResultType(types.Integer)},
							},
						},
						Distinct: false,
					},
					Expression: model.ResultType(&types.List{ElementType: types.Long}),
				},
			},
		},
		{
			name:         "Multiple Conversions - Subtype and FHIR ModelInfo Converison to Simple",
			invokedType:  &types.Named{TypeName: "FHIR.id"},
			declaredType: types.String,
			want: ConvertedOperand{
				Matched: true,
				Score:   5,
				WrappedOperand: &model.FunctionRef{
					Expression:  model.ResultType(types.String),
					Name:        "ToString",
					LibraryName: "FHIRHelpers",
					Operands:    []model.IExpression{model.NewLiteral("operand", types.String)},
				}},
		},
		{
			name: "Multiple Conversions - FHIR Cast and Implicit Conversion to Simple",
			invokedType: &types.Choice{
				ChoiceTypes: []types.IType{
					&types.Named{TypeName: "FHIR.Quantity"},
					&types.Named{TypeName: "FHIR.CodeableConcept"},
					&types.Named{TypeName: "FHIR.string"},
					&types.Named{TypeName: "FHIR.boolean"},
					&types.Named{TypeName: "FHIR.integer"},
					&types.Named{TypeName: "FHIR.Range"},
					&types.Named{TypeName: "FHIR.Ratio"},
					&types.Named{TypeName: "FHIR.SampledData"},
					&types.Named{TypeName: "FHIR.time"},
					&types.Named{TypeName: "FHIR.dateTime"},
					&types.Named{TypeName: "FHIR.Period"},
				},
			},
			declaredType: types.Boolean,
			want: ConvertedOperand{
				Matched: true,
				Score:   7,
				WrappedOperand: &model.FunctionRef{
					LibraryName: "FHIRHelpers",
					Name:        "ToBoolean",
					Expression:  model.ResultType(types.Boolean),
					Operands: []model.IExpression{
						&model.As{
							UnaryExpression: &model.UnaryExpression{
								Operand:    model.NewLiteral("operand", types.String),
								Expression: model.ResultType(&types.Named{TypeName: "FHIR.boolean"}),
							},
							AsTypeSpecifier: &types.Named{TypeName: "FHIR.boolean"},
							Strict:          false,
						},
					},
				},
			},
		},
		{
			name: "Multiple Conversions - Cast and Implicit Conversion to Class",
			invokedType: &types.Choice{
				ChoiceTypes: []types.IType{
					&types.Interval{PointType: types.Date},
				},
			},
			declaredType: &types.Interval{PointType: types.DateTime},
			want: ConvertedOperand{
				Matched: true,
				Score:   8,
				WrappedOperand: &model.Interval{
					Expression: model.ResultType(&types.Interval{PointType: types.DateTime}),
					LowClosedExpression: &model.Property{
						Source: &model.As{
							UnaryExpression: &model.UnaryExpression{
								Operand:    model.NewLiteral("operand", types.String),
								Expression: model.ResultType(&types.Interval{PointType: types.Date}),
							},
							AsTypeSpecifier: &types.Interval{PointType: types.Date},
							Strict:          false,
						},
						Path:       "lowClosed",
						Expression: model.ResultType(types.Boolean),
					},
					HighClosedExpression: &model.Property{
						Source: &model.As{
							UnaryExpression: &model.UnaryExpression{
								Operand:    model.NewLiteral("operand", types.String),
								Expression: model.ResultType(&types.Interval{PointType: types.Date}),
							},
							AsTypeSpecifier: &types.Interval{PointType: types.Date},
							Strict:          false,
						},
						Path:       "highClosed",
						Expression: model.ResultType(types.Boolean),
					},
					Low: &model.ToDateTime{
						UnaryExpression: &model.UnaryExpression{
							Expression: model.ResultType(types.DateTime),
							Operand: &model.Property{
								Path:       "low",
								Expression: model.ResultType(types.Date),
								Source: &model.As{
									UnaryExpression: &model.UnaryExpression{
										Operand:    model.NewLiteral("operand", types.String),
										Expression: model.ResultType(&types.Interval{PointType: types.Date}),
									},
									AsTypeSpecifier: &types.Interval{PointType: types.Date},
									Strict:          false,
								},
							},
						},
					},
					High: &model.ToDateTime{
						UnaryExpression: &model.UnaryExpression{
							Expression: model.ResultType(types.DateTime),
							Operand: &model.Property{
								Path:       "high",
								Expression: model.ResultType(types.Date),
								Source: &model.As{
									UnaryExpression: &model.UnaryExpression{
										Operand:    model.NewLiteral("operand", types.String),
										Expression: model.ResultType(&types.Interval{PointType: types.Date}),
									},
									AsTypeSpecifier: &types.Interval{PointType: types.Date},
									Strict:          false,
								},
							},
						},
					},
				},
			},
		},
		{
			name:         "Invalid Simple Conversion",
			invokedType:  types.Integer,
			declaredType: types.DateTime,
			want:         ConvertedOperand{Matched: false},
		},
		{
			name:         "Invalid Choice to Concrete",
			invokedType:  &types.Choice{ChoiceTypes: []types.IType{types.Date, types.String}},
			declaredType: types.Integer,
			want:         ConvertedOperand{Matched: false},
		},
		{
			name:         "Invalid Concrete to Choice",
			invokedType:  &types.Interval{PointType: types.Date},
			declaredType: &types.Choice{ChoiceTypes: []types.IType{types.Integer, types.String}},
			want:         ConvertedOperand{Matched: false},
		},
		{
			name: "Invalid Choice to Choice",
			invokedType: &types.Choice{
				ChoiceTypes: []types.IType{
					types.Date,
					types.String,
				},
			},
			declaredType: &types.Choice{
				ChoiceTypes: []types.IType{
					types.Integer,
					&types.Interval{PointType: types.String},
				},
			},
			want: ConvertedOperand{Matched: false},
		},
		{
			name:         "Invalid FHIR ModelInfo Conversion",
			invokedType:  &types.Named{TypeName: "FHIR.date"},
			declaredType: types.Integer,
			want:         ConvertedOperand{Matched: false},
		},
		{
			name:         "Invalid Interval Conversion",
			invokedType:  &types.Interval{PointType: types.String},
			declaredType: &types.Interval{PointType: types.Integer},
			want:         ConvertedOperand{Matched: false},
		},
		{
			name:         "Invalid List Conversion",
			invokedType:  &types.List{ElementType: &types.Interval{PointType: types.String}},
			declaredType: &types.List{ElementType: types.String},
			want:         ConvertedOperand{Matched: false},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			modelinfo := newFHIRModelInfo(t)
			got, err := OperandImplicitConverter(tc.invokedType, tc.declaredType, model.NewLiteral("operand", types.String), modelinfo)
			if err != nil {
				t.Fatalf("operandImplicitConverter() unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("operandImplicitConverter() diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestOperandImplicitConverter_NilOperands(t *testing.T) {
	modelinfo := newFHIRModelInfo(t)
	got, err := OperandImplicitConverter(types.Date, types.DateTime, nil, modelinfo)
	if err != nil {
		t.Fatalf("operandImplicitConverter() unexpected error: %v", err)
	}

	want := ConvertedOperand{
		Matched: true,
		Score:   4,
		WrappedOperand: &model.ToDateTime{
			UnaryExpression: &model.UnaryExpression{
				// This is nil which is ok because the caller does not care about the wrapped operand only
				// the score.
				Operand:    nil,
				Expression: model.ResultType(types.DateTime),
			},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("operandImplicitConverter() diff (-want +got):\n%s", diff)
	}
}

func TestOperandImplicitConverter_Error(t *testing.T) {
	tests := []struct {
		name         string
		invokedType  types.IType
		declaredType types.IType
		errContains  string
	}{
		{
			name:         "invokedType Unsupported",
			invokedType:  types.Unset,
			declaredType: types.Integer,
			errContains:  "internal error - invokedType is",
		},
		{
			name:         "declaredType Unsupported",
			invokedType:  types.Integer,
			declaredType: types.Unset,
			errContains:  "internal error - declaredType is",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			modelinfo := newFHIRModelInfo(t)
			_, err := OperandImplicitConverter(tc.invokedType, tc.declaredType, model.NewLiteral("operand", types.String), modelinfo)
			if err == nil {
				t.Fatalf("OperandImplicitConverter() did not return an error")
			}
			if !strings.Contains(err.Error(), tc.errContains) {
				t.Errorf("Returned error (%s) did not contain expected string (%s)", err.Error(), tc.errContains)
			}
		})
	}
}

func TestOperandsToString(t *testing.T) {
	tests := []struct {
		name     string
		operands []model.IExpression
		want     string
	}{
		{
			name: "Multiple",
			operands: []model.IExpression{
				model.NewLiteral("String", types.String),
				model.NewInclusiveInterval("@2020-03-04", "@2020-03-05", types.Date),
				model.NewList([]string{"4", "5"}, types.Integer),
			},
			want: "System.String, Interval<System.Date>, List<System.Integer>",
		},
		{
			name:     "Single",
			operands: []model.IExpression{model.NewLiteral("String", types.String)},
			want:     "System.String",
		},
		{
			name:     "Empty",
			operands: []model.IExpression{},
			want:     "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := OperandsToString(tc.operands)
			if got != tc.want {
				t.Errorf("OperandsToString() = %v want: %v", got, tc.want)
			}
		})
	}
}

func newFHIRModelInfo(t *testing.T) *modelinfo.ModelInfos {
	t.Helper()
	rawFHIRMI, err := embeddata.ModelInfos.ReadFile("third_party/cqframework/fhir-modelinfo-4.0.1.xml")
	if err != nil {
		t.Fatalf("Reading embedded file %s failed unexpectedly: %v", "third_party/cqframework/fhir-modelinfo-4.0.1.xml", err)
	}

	mi, err := modelinfo.New([][]byte{rawFHIRMI})
	if err != nil {
		t.Fatalf("modelinfo.New() unexpected error: %v", err)
	}
	mi.SetUsing(modelinfo.Key{Name: "FHIR", Version: "4.0.1"})
	return mi
}
