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

	"github.com/google/cql/types"
)

func TestExactOverloadMatch(t *testing.T) {
	tests := []struct {
		name      string
		invoked   []types.IType
		overloads []Overload[string]
		want      string
	}{
		{
			name:    "Multiple Operands",
			invoked: []types.IType{types.String, &types.Interval{PointType: types.Date}, &types.List{ElementType: types.Integer}},
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
			want: "Just Right",
		},
		{
			name:    "Single Operand",
			invoked: []types.IType{types.String},
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
			want: "Just Right",
		},
		{
			name:    "No Operands",
			invoked: []types.IType{},
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
			want: "Just Right",
		},
		{
			name:    "SubType Is Exact Match",
			invoked: []types.IType{types.Integer, types.Date},
			overloads: []Overload[string]{
				Overload[string]{
					Result:   "SubType",
					Operands: []types.IType{types.Any, types.Date},
				},
				Overload[string]{
					Result:   "One Simple Conversion",
					Operands: []types.IType{types.Decimal, types.Date},
				},
			},
			want: "SubType",
		},
		{
			name:    "Conversions Followed By Exact Match",
			invoked: []types.IType{types.Integer, types.Date},
			overloads: []Overload[string]{
				Overload[string]{
					Result:   "Conversion 1",
					Operands: []types.IType{types.Long, types.DateTime},
				},
				Overload[string]{
					Result:   "Conversion 2",
					Operands: []types.IType{types.Decimal, types.DateTime},
				},
				Overload[string]{
					Result:   "Exact Match",
					Operands: []types.IType{types.Integer, types.Date},
				},
			},
			want: "Exact Match",
		},
		{
			name:    "Exact Match beats SubType match",
			invoked: []types.IType{types.CodeSystem, types.CodeSystem},
			overloads: []Overload[string]{
				Overload[string]{
					Result:   "Exact Match",
					Operands: []types.IType{types.CodeSystem, types.CodeSystem},
				},
				Overload[string]{
					Result:   "SubType",
					Operands: []types.IType{types.Any, types.Any},
				},
			},
			want: "Exact Match",
		},
		// List SubType tests
		{
			name:    "Can disambiguate Any between List<Any> and Any",
			invoked: []types.IType{&types.List{ElementType: types.Any}, types.Any},
			overloads: []Overload[string]{
				Overload[string]{
					Result:   "Exact Match",
					Operands: []types.IType{&types.List{ElementType: types.Any}, types.Any},
				},
				Overload[string]{
					Result:   "SubType",
					Operands: []types.IType{&types.List{ElementType: types.Any}, &types.List{ElementType: types.Any}},
				},
			},
			want: "Exact Match",
		},
		{
			name:    "Can disambiguate List<Any> between List<Any> and Any",
			invoked: []types.IType{&types.List{ElementType: types.Any}, &types.List{ElementType: types.Any}},
			overloads: []Overload[string]{
				Overload[string]{
					Result:   "Exact Match",
					Operands: []types.IType{&types.List{ElementType: types.Any}, &types.List{ElementType: types.Any}},
				},
				Overload[string]{
					Result:   "SubType",
					Operands: []types.IType{&types.List{ElementType: types.Any}, types.Any},
				},
			},
			want: "Exact Match",
		},
		{
			name:    "Can disambiguate three layer deep list of Any from Any",
			invoked: []types.IType{&types.List{ElementType: &types.List{ElementType: &types.List{ElementType: types.Any}}}},
			overloads: []Overload[string]{
				Overload[string]{
					Result:   "Exact Match",
					Operands: []types.IType{&types.List{ElementType: &types.List{ElementType: &types.List{ElementType: types.Any}}}},
				},
				Overload[string]{
					Result:   "SubType",
					Operands: []types.IType{types.Any},
				},
			},
			want: "Exact Match",
		},
		{
			name:    "Chooses the most specific match between two sub types",
			invoked: []types.IType{&types.List{ElementType: types.Integer}},
			overloads: []Overload[string]{
				Overload[string]{
					Result:   "Exact Match",
					Operands: []types.IType{&types.List{ElementType: types.Any}},
				},
				Overload[string]{
					Result:   "SubType",
					Operands: []types.IType{types.Any},
				},
			},
			want: "Exact Match",
		},
		// Interval SubType tests
		{
			name:    "Can disambiguate Any between Interval<Integer> and Any",
			invoked: []types.IType{&types.Interval{PointType: types.Integer}, types.Any},
			overloads: []Overload[string]{
				Overload[string]{
					Result:   "Exact Match",
					Operands: []types.IType{&types.Interval{PointType: types.Integer}, types.Any},
				},
				Overload[string]{
					Result:   "SubType",
					Operands: []types.IType{&types.Interval{PointType: types.Integer}, &types.Interval{PointType: types.Integer}},
				},
			},
			want: "Exact Match",
		},
		{
			name:    "Can disambiguate Interval<Integer> between Interval<Integer> and Any",
			invoked: []types.IType{&types.Interval{PointType: types.Integer}, &types.Interval{PointType: types.Integer}},
			overloads: []Overload[string]{
				Overload[string]{
					Result:   "Exact Match",
					Operands: []types.IType{&types.Interval{PointType: types.Integer}, &types.Interval{PointType: types.Integer}},
				},
				Overload[string]{
					Result:   "SubType",
					Operands: []types.IType{&types.Interval{PointType: types.Integer}, types.Any},
				},
			},
			want: "Exact Match",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			modelInfo := newFHIRModelInfo(t)
			got, err := ExactOverloadMatch(tc.invoked, tc.overloads, modelInfo, "Name")
			if err != nil {
				t.Fatalf("ExactOverloadMatch() unexpected err: %v", err)
			}
			if got != tc.want {
				t.Errorf("ExactOverloadMatch() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestExactOverloadMatch_Error(t *testing.T) {
	tests := []struct {
		name        string
		invoked     []types.IType
		overloads   []Overload[string]
		errContains string
	}{
		{
			name:    "Conversions No Match",
			invoked: []types.IType{types.Integer, types.Date},
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
			errContains: "could not resolve",
		},
		{
			name:    "Single No Match",
			invoked: []types.IType{types.String},
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
		},
		{
			name: "Multiple No Match",
			invoked: []types.IType{
				types.String,
				&types.Interval{PointType: types.Date},
				&types.List{ElementType: types.Integer},
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
		},
		{
			name:    "Ambiguous Match",
			invoked: []types.IType{types.String, types.String},
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
		},
		{
			name:    "Unsupported ResultType",
			invoked: []types.IType{types.Unset},
			overloads: []Overload[string]{
				Overload[string]{
					Result:   "Overload",
					Operands: []types.IType{types.String},
				},
			},
			errContains: "internal error - invokedType",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			modelinfo := newFHIRModelInfo(t)
			_, err := ExactOverloadMatch(tc.invoked, tc.overloads, modelinfo, "Name")
			if err == nil {
				t.Fatalf("ExactOverloadMatch() did not return an error")
			}
			if !strings.Contains(err.Error(), tc.errContains) {
				t.Errorf("Returned error (%s) did not contain expected string (%s)", err.Error(), tc.errContains)
			}
		})
	}
}
