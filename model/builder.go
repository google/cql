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
	"github.com/google/cql/types"
)

// NewLiteral is a helper function to build a model.Literal.
func NewLiteral(value string, t types.IType) *Literal {
	return &Literal{Value: value, Expression: ResultType(t)}
}

// NewInclusiveInterval returns an Interval[NewLiteral(low), NewLiteral(high)] where the low and high are
// literals of type t.
func NewInclusiveInterval(low, high string, t types.IType) *Interval {
	return &Interval{
		Low:           NewLiteral(low, t),
		High:          NewLiteral(high, t),
		LowInclusive:  true,
		HighInclusive: true,
		Expression:    ResultType(&types.Interval{PointType: t}),
	}
}

// NewList returns a List{NewLiteral(elems[0]), NewLiteral(elems[1]), ...} where the elements of the
// list are literals of type t constructed from elems.
func NewList(elems []string, t types.IType) *List {
	l := &List{
		List:       []IExpression{},
		Expression: ResultType(&types.List{ElementType: t}),
	}
	for _, elem := range elems {
		l.List = append(l.List, NewLiteral(elem, t))
	}
	return l
}

// ResultType is a helper function to set the resultType in a model.Element.
func ResultType(t types.IType) *Expression {
	return &Expression{
		Element: &Element{ResultType: t},
	}
}
