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

package model

import (
	"testing"

	"github.com/google/cql/types"
	"github.com/google/go-cmp/cmp"
)

func TestBinaryExpression(t *testing.T) {
	cases := []struct {
		name      string
		exp       *BinaryExpression
		wantName  string
		wantLeft  IExpression
		wantRight IExpression
	}{
		{
			name: "Simple",
			exp: &BinaryExpression{
				Operands: []IExpression{
					&Literal{Value: "10"},
					&Literal{Value: "20"},
				},
			},
			wantName:  "test",
			wantLeft:  &Literal{Value: "10"},
			wantRight: &Literal{Value: "20"},
		},
		{
			name:      "Missing all operands",
			exp:       &BinaryExpression{},
			wantName:  "test",
			wantLeft:  nil,
			wantRight: nil,
		},
		{
			name: "Missing one operand",
			exp: &BinaryExpression{
				Operands: []IExpression{
					&Literal{Value: "10"},
				},
			},
			wantName:  "test",
			wantLeft:  &Literal{Value: "10"},
			wantRight: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !cmp.Equal(tc.exp.Left(), tc.wantLeft) {
				t.Errorf("GetLeft() = %v, want %v", tc.exp.Left(), tc.wantLeft)
			}
			if !cmp.Equal(tc.exp.Right(), tc.wantRight) {
				t.Errorf("GetRight() = %v, want %v", tc.exp.Right(), tc.wantRight)
			}
		})
	}
}

func TestNilTypeSpecifier(t *testing.T) {
	t.Run("Nil Expression", func(t *testing.T) {
		l := Literal{
			Expression: nil,
			Value:      "10",
		}
		if got := l.GetResultType(); got != types.Unset {
			t.Errorf("%v.GetResultType() = %v, want types.Unsupported", l, got)
		}
	})

	t.Run("Nil Element", func(t *testing.T) {
		l := Literal{
			Expression: &Expression{Element: nil},
			Value:      "10",
		}
		if got := l.GetResultType(); got != types.Unset {
			t.Errorf("%v.GetResultType() = %v, want types.Unsupported", l, got)
		}
	})
}
