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
	"errors"
	"fmt"
	"math"

	"github.com/google/cql/internal/modelinfo"
	"github.com/google/cql/model"
	"github.com/google/cql/types"
)

// Infered is the result of the InferMixedType function.
type Infered struct {
	// PuntedToChoice is true if all types could not be implicitly converted to a uniform type and we
	// instead converted to a Choice type.
	PuntedToChoice bool
	// UniformType is the type that all invoked types were converted to.
	UniformType types.IType
	// WrappedOperands are the invoked operands wrapped in all necessary system operators and function
	// refs to convert them.
	WrappedOperands []model.IExpression
}

// InferMixed wraps the invoked operands in all necessary system operators and function refs to
// convert them to the same type. As a last resort, InferMixed will convert all operands to a
// Choice types consisting of the set of types of the operands. The least converting path is chosen
// and on ambiguous paths any of the least converting paths is chosen. This function is used to
// infer the type of mixed lists and case expressions in the parser.
//
// [4, 4.5] --> [ToDecimal(4), 4.5]
// ['str', 4] --> [As{'str', Choice<String, Integer>}, As{4, Choice<String, Integer>}]
func InferMixed(invoked []model.IExpression, modelinfo *modelinfo.ModelInfos) (Infered, error) {
	return inferMixedType(OperandsToTypes(invoked), invoked, modelinfo)
}

// inferMixed may be called with nil opsToWrap if the caller only cares about the uniform type.
func inferMixedType(invokedTypes []types.IType, opsToWrap []model.IExpression, modelinfo *modelinfo.ModelInfos) (Infered, error) {
	if len(invokedTypes) == 0 {
		return Infered{PuntedToChoice: false, UniformType: types.Any, WrappedOperands: []model.IExpression{}}, nil
	}

	if opsToWrap == nil {
		// A slice of nil used as placeholder model.IExpressions.
		opsToWrap = make([]model.IExpression, len(invokedTypes))
	}

	allTypesAny := true
	for _, t := range invokedTypes {
		if !t.Equal(types.Any) {
			allTypesAny = false
			break
		}
	}
	if allTypesAny {
		// This special case is necessary because we skip trying the uniform type of Any below. If we
		// are passed all nulls, then the desired behaviour is to return uniform type of Any.
		return Infered{PuntedToChoice: false, UniformType: types.Any, WrappedOperands: opsToWrap}, nil
	}

	minScore := math.MaxInt
	matched := []model.IExpression{}
	var matchedType types.IType
	for _, t := range invokedTypes {
		if t.Equal(types.Any) {
			// All types are a subtype of Any, and can be converted to Any. However, when inferring a
			// uniform type we prefer to use compatible (score 3) over subtype (score 2) and cast the null
			// of type Any to a concrete type. So skip the Any overload.
			continue
		}
		// TODO(b/301606416): Right now we just check the types in invoked as possible UniformTypes, but
		// there may be a UniformType that requires converting all invoked types. For example, if A and
		// B are subtypes of C then invoked (A, B) should return UniformType C. This does create a
		// problem because every type can be converted to Any which would override the intended
		// behaviour of punting to a Choice type below.
		possibleType := make([]types.IType, len(invokedTypes))
		for i := range possibleType {
			possibleType[i] = t
		}

		res, err := operandsImplicitConverter(invokedTypes, possibleType, opsToWrap, modelinfo)
		if err != nil {
			return Infered{}, fmt.Errorf("while inferring mixed type: %w", err)
		}

		if res.Matched && res.Score < minScore {
			// A new least converting match
			minScore = res.Score

			// Beware of the shallow copy
			matched = res.WrappedOperands
			matchedType = t
		}
	}

	// I am not sure if this is the correct behaviour, but if ambiguous we just take the last chosen
	// conversion. I cannot think of a case that would result in ambiguous other than when all types
	// are equal.
	if minScore != math.MaxInt {
		// Matched with conversion to a single type.
		return Infered{PuntedToChoice: false, UniformType: matchedType, WrappedOperands: matched}, nil
	}

	// All types in invoked could not be converted to a uniform type. Convert them to a uniform choice
	// type instead.
	choiceType, err := DeDuplicate(invokedTypes)
	if err != nil {
		return Infered{}, err
	}

	for _, o := range opsToWrap {
		wrapped := &model.As{
			UnaryExpression: &model.UnaryExpression{
				Operand:    o,
				Expression: model.ResultType(choiceType),
			},
			AsTypeSpecifier: choiceType,
			Strict:          false,
		}
		matched = append(matched, wrapped)
	}

	return Infered{PuntedToChoice: true, UniformType: choiceType, WrappedOperands: matched}, nil
}

// DeDuplicate finds a minimal choice type given a list of types. Duplicates are removed and
// Choice types are recursively flattened. No implicit conversions are applied. Ex:
// [Integer, String, Choice<Integer, Quantity>, String] will return Choice<Integer, Quantity, String>
func DeDuplicate(ts []types.IType) (types.IType, error) {
	if ts == nil || len(ts) == 0 {
		return nil, errors.New("internal error - empty or nil list of types passed to DeDuplicate")
	}

	var flatTypes []types.IType
	for _, t := range ts {
		ts, err := flattenChoices(t, 0)
		if err != nil {
			return nil, err
		}
		flatTypes = append(flatTypes, ts...)
	}

	choiceType := &types.Choice{ChoiceTypes: []types.IType{}}
	for _, t := range flatTypes {
		if !containsType(choiceType.ChoiceTypes, t) {
			choiceType.ChoiceTypes = append(choiceType.ChoiceTypes, t)
		}
	}

	if len(choiceType.ChoiceTypes) == 1 {
		// Choice type is not needed.
		return choiceType.ChoiceTypes[0], nil
	}

	return choiceType, nil
}

// Intersect finds the intersection of two types. Choice types are flattened and no implicit
// conversions are applied. Ex Integer, Choice<Integer, Quantity> will return Integer.
func Intersect(left types.IType, right types.IType) (types.IType, error) {
	if left.Equal(right) {
		return left, nil
	}

	flatTypesLeft, err := flattenChoices(left, 0)
	if err != nil {
		return nil, err
	}
	flatTypesRight, err := flattenChoices(right, 0)
	if err != nil {
		return nil, err
	}

	choiceType := &types.Choice{ChoiceTypes: []types.IType{}}
	for _, t := range flatTypesLeft {
		if containsType(flatTypesRight, t) && !containsType(choiceType.ChoiceTypes, t) {
			choiceType.ChoiceTypes = append(choiceType.ChoiceTypes, t)
		}
	}

	if len(choiceType.ChoiceTypes) == 1 {
		// Choice type is not needed.
		return choiceType.ChoiceTypes[0], nil
	}

	if len(choiceType.ChoiceTypes) == 0 {
		return nil, fmt.Errorf("no common types between %v and %v", left, right)
	}

	return choiceType, nil
}

func flattenChoices(t types.IType, recursion int) ([]types.IType, error) {
	if recursion > 100000 {
		return nil, fmt.Errorf("internal error - nested choice recursion limit exceeded")
	}
	choiceT, ok := t.(*types.Choice)
	if !ok {
		// Not a choice type so end recursion.
		return []types.IType{t}, nil
	}

	var flatTypes []types.IType
	for _, t := range choiceT.ChoiceTypes {
		ts, err := flattenChoices(t, recursion+1)
		if err != nil {
			return nil, err
		}
		flatTypes = append(flatTypes, ts...)
	}
	return flatTypes, nil
}

func containsType(types []types.IType, arg types.IType) bool {
	for _, t := range types {
		if t.Equal(arg) {
			return true
		}
	}
	return false
}
