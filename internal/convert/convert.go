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

// Package convert is responsible for all things related to implicit conversions. This includes
// inserting necessary implicit conversions into the model at parse time and finding the exact match
// overload at run time in the interpreter.
package convert

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/google/cql/internal/modelinfo"
	"github.com/google/cql/model"
	"github.com/google/cql/types"
)

// ErrAmbiguousMatch is returned when two or more overloads were matched with the same score.
var ErrAmbiguousMatch = errors.New("ambiguous match")

// ErrNoMatch is returned when no overloads were matched.
var ErrNoMatch = errors.New("no matching overloads")

// Overload holds the the declared operands and the result returned if those operands are matched by
// an invocation.
type Overload[F any] struct {
	Operands []types.IType
	// Result is what is returned by OverloadMatch.
	Result F
}

// MatchedOverload returns the result of the OverloadMatch function.
type MatchedOverload[F any] struct {
	// Result is the result of the overload that was matched.
	Result F
	// WrappedOperands are the operands wrapped in all necessary system operators and function refs to
	// convert them to the matched overload.
	WrappedOperands []model.IExpression
}

// OverloadMatch returns MatchedOverload on a match, and an error if there is no match or if the
// match is ambiguous. OverloadMatch returns the least converting match by summing the conversion
// score of each of the arguments. Ambiguous is returned if the least converting match has the same
// conversion score. Example of an ambiguous match:
//
// Declared: Foo(Long), Foo(Decimal)
// Invocation: Foo(Integer)
//
// Name is the function name and is only used for error messages.
// https://cql.hl7.org/03-developersguide.html#function-resolution
// https://cql.hl7.org/03-developersguide.html#conversion-precedence
func OverloadMatch[F any](invoked []model.IExpression, overloads []Overload[F], modelinfo *modelinfo.ModelInfos, name string) (MatchedOverload[F], error) {
	if len(overloads) == 0 {
		return MatchedOverload[F]{}, fmt.Errorf("could not resolve %v(%v): %w", name, OperandsToString(invoked), ErrNoMatch)
	}

	// To create concreteOverloads we go through the overloads, find any that have generics and
	// replace the generics with concrete types.
	concreteOverloads := make([]Overload[F], 0, len(overloads))
	for _, overload := range overloads {
		if isGeneric(overload.Operands) {
			concreteOverload, matched, err := convertGeneric(invoked, overload, modelinfo)
			if err != nil {
				return MatchedOverload[F]{}, fmt.Errorf("%v(%v): %w", name, OperandsToString(invoked), err)
			}
			if matched {
				concreteOverloads = append(concreteOverloads, concreteOverload)
			}
		} else {
			concreteOverloads = append(concreteOverloads, overload)
		}
	}

	ambiguous := false
	minScore := math.MaxInt
	currTypePrecedenceScore := math.MaxInt
	matched := MatchedOverload[F]{WrappedOperands: make([]model.IExpression, len(invoked))}
	for _, overload := range concreteOverloads {
		res, err := operandsImplicitConverter(OperandsToTypes(invoked), overload.Operands, invoked, modelinfo)
		if err != nil {
			return MatchedOverload[F]{}, fmt.Errorf("%v(%v): %w", name, OperandsToString(invoked), err)
		}
		if res.Matched && res.Score == minScore && res.TypePrecedenceScore == currTypePrecedenceScore {
			// The least converting match is now ambiguous
			ambiguous = true
			continue
		}

		if res.Matched && (res.Score < minScore || (res.Score == minScore && res.TypePrecedenceScore < currTypePrecedenceScore)) {
			// A new least converting match
			ambiguous = false
			minScore = res.Score
			currTypePrecedenceScore = res.TypePrecedenceScore
			matched.Result = overload.Result
			// Beware of the shallow copy
			matched.WrappedOperands = res.WrappedOperands
		}
	}
	if ambiguous {
		return matched, fmt.Errorf("%v(%v) %w", name, OperandsToString(invoked), ErrAmbiguousMatch)
	}

	if minScore != math.MaxInt {
		// Matched with conversion to a single overloaded function.
		return matched, nil
	}

	// Build a list of available overloads to provide more helpful error message
	var availableOverloads strings.Builder
	if len(concreteOverloads) > 0 {
		availableOverloads.WriteString(" available overloads: [")
		for i, overload := range concreteOverloads {
			if i > 0 {
				availableOverloads.WriteString(", ")
			}
			availableOverloads.WriteString(fmt.Sprintf("%v(%v)", name, operandsToStringForTypes(overload.Operands)))
		}
		availableOverloads.WriteString("]")
	}

	return MatchedOverload[F]{}, fmt.Errorf("could not resolve %v(%v): %w%v",
		name, OperandsToString(invoked), ErrNoMatch, availableOverloads.String())
}

type convertedOperands struct {
	// Matched is true if the invoked types can be converted to the declared types.
	Matched bool
	// Score is the least converting score of the conversion based on the CQL conversion precedence.
	// The score of each operand are added together. Multiple conversions on the same operand are not
	// taken into account. https://cql.hl7.org/03-developersguide.html#conversion-precedence
	Score int
	// TypePrecedenceScore is the score based on the declared type category precedence. Lower is
	// better. Used as a tie-breaker when Score is equal between overloads.
	TypePrecedenceScore int
	// WrappedOperands are the operands wrapped in all necessary system operators and function refs to
	// convert them.
	WrappedOperands []model.IExpression
}

// operandsImplicitConverter may be called with nil opsToWrap if the caller only cares about the
// score.
func operandsImplicitConverter(invokedTypes []types.IType, declaredTypes []types.IType, opsToWrap []model.IExpression, tHelper *modelinfo.ModelInfos) (convertedOperands, error) {
	if len(invokedTypes) != len(declaredTypes) {
		return convertedOperands{Matched: false}, nil
	}

	if opsToWrap == nil {
		// A slice of nil used as placeholder model.IExpressions.
		opsToWrap = make([]model.IExpression, len(invokedTypes))
	}

	results := convertedOperands{Matched: true, Score: 0, WrappedOperands: make([]model.IExpression, len(invokedTypes))}
	for i := range invokedTypes {
		result, err := OperandImplicitConverter(invokedTypes[i], declaredTypes[i], opsToWrap[i], tHelper)
		if err != nil {
			return convertedOperands{}, err
		}
		if !result.Matched {
			return convertedOperands{Matched: false}, nil
		}
		results.Score += result.Score
		results.TypePrecedenceScore += result.TypePrecedenceScore
		results.WrappedOperands[i] = result.WrappedOperand
	}
	return results, nil
}

// ConvertedOperand is the result of OperandImplicitConverter.
type ConvertedOperand struct {
	// Matched is true if the invoked type can be converted to the declared type.
	Matched bool
	// Score is the least converting score of the conversion based on the CQL conversion precedence.
	// The score does not take into account multiple conversions.
	// https://cql.hl7.org/03-developersguide.html#conversion-precedence
	Score int
	// TypePrecedenceScore is the score based on the declared type category precedence. Lower is better.
	// Used as a tie-breaker when Score is equal between overloads.
	TypePrecedenceScore int
	// WrappedOperand is the operand wrapped in all necessary system operators and function refs to
	// convert it.
	WrappedOperand model.IExpression
}

// OperandImplicitConverter wraps the operand in any system operators or FHIRHelper function
// references needed to convert the operand from invokedType to declaredType. For example, if going
// from Integer --> Decimal the operand will be wrapped in ToDecimal(operand).
// operandImplicitConverter may apply multiple conversions, and always returns the least converting
// path.
//
// OperandImplicitConverter may be called with a nil opToWrap if the caller only cares about the
// score.
//
// Implementation note, invokedType may not be the same as opToWrap.GetResultType().The two
// diverge on some recursive calls. opToWrap.GetResultType() should not be used in the implementation
// of operandImplicitConverter.
func OperandImplicitConverter(invokedType types.IType, declaredType types.IType, opToWrap model.IExpression, mi *modelinfo.ModelInfos) (ConvertedOperand, error) {
	if invokedType == types.Unset {
		return ConvertedOperand{}, fmt.Errorf("internal error - invokedType is %v", invokedType)
	}
	if declaredType == types.Unset {
		return ConvertedOperand{}, fmt.Errorf("internal error - declaredType is %v", declaredType)
	}

	// Try different conversion paths minConverted will keep track of the lowest scoring conversion.
	declaredTypePrecedence, err := getTypeCategoryPrecedence(declaredType)
	if err != nil {
		return ConvertedOperand{}, err
	}
	minConverted := ConvertedOperand{Score: math.MaxInt, TypePrecedenceScore: declaredTypePrecedence}

	// EXACT MATCH
	if invokedType.Equal(declaredType) {
		return ConvertedOperand{Matched: true, Score: 0, TypePrecedenceScore: declaredTypePrecedence, WrappedOperand: opToWrap}, nil
	}

	// SUBTYPE
	isSub, err := mi.IsSubType(invokedType, declaredType)
	if err != nil {
		return ConvertedOperand{}, err
	}
	if isSub {
		// No conversion wrapper is needed, the interpreter will handle subtypes.
		minConverted = ConvertedOperand{Matched: true, Score: 1, TypePrecedenceScore: declaredTypePrecedence, WrappedOperand: opToWrap}
	}

	// All types can be converted from invoked --> Any --> declared. However that leads to incorrect
	// conversions such as String --> Any --> Integer, so BaseTypes does not return Any.
	baseTypes, err := mi.BaseTypes(invokedType)
	if err != nil {
		return ConvertedOperand{}, err
	}
	for _, baseType := range baseTypes {
		r, err := OperandImplicitConverter(baseType, declaredType, opToWrap, mi)
		if err != nil {
			return ConvertedOperand{}, err
		}
		if r.Matched {
			// Increment score by one since we applied a subtype before recursively calling
			// OperandImplicitConverter.
			r.Score++
			if r.Score < minConverted.Score {
				minConverted = r
			}
		}
	}

	// COMPATIBLE/NULL
	// https://cql.hl7.org/03-developersguide.html#casting
	//
	// This is not described well in the CQL spec but, you can pass null literals to any function.
	// Null has a type of Any, and is documented as being compatible with all types (Conversion
	// Precedence step 3).
	// Ex Any --> Decimal, implicitly calls: As(operand, Decimal)
	if invokedType.Equal(types.Any) {
		wrapped := &model.As{
			UnaryExpression: &model.UnaryExpression{
				Operand:    opToWrap,
				Expression: model.ResultType(declaredType),
			},
			AsTypeSpecifier: declaredType,
			Strict:          false,
		}
		if 2 < minConverted.Score {
			minConverted = ConvertedOperand{Matched: true, Score: 2, TypePrecedenceScore: declaredTypePrecedence, WrappedOperand: wrapped}
		}
	}

	// CAST - invokedType is a Choice type
	// Ex Choice<Integer> --> Decimal   ToDecimal(As(operand, Integer))
	if invokedChoice, ok := invokedType.(*types.Choice); ok {
		for _, choiceType := range invokedChoice.ChoiceTypes {
			choiceWrapped := &model.As{
				UnaryExpression: &model.UnaryExpression{
					Operand:    opToWrap,
					Expression: model.ResultType(choiceType),
				},
				AsTypeSpecifier: choiceType,
				Strict:          false,
			}
			r, err := OperandImplicitConverter(choiceType, declaredType, choiceWrapped, mi)
			if err != nil {
				return ConvertedOperand{}, err
			}
			if r.Matched {
				r.Score += 3
				if r.Score < minConverted.Score {
					minConverted = r
				}
			}
		}
	}

	// CAST - declaredType is a Choice type
	// Ex Integer --> Choice<Decimal>   As(ToDecimal(operand), Choice<Decimal>)
	if declaredChoice, ok := declaredType.(*types.Choice); ok {
		for _, choiceType := range declaredChoice.ChoiceTypes {
			r, err := OperandImplicitConverter(invokedType, choiceType, opToWrap, mi)
			if err != nil {
				return ConvertedOperand{}, err
			}
			if r.Matched {
				wrapped := &model.As{
					UnaryExpression: &model.UnaryExpression{
						Operand:    r.WrappedOperand,
						Expression: model.ResultType(declaredType),
					},
					AsTypeSpecifier: declaredType,
					Strict:          false,
				}
				if 3 < minConverted.Score {
					minConverted = ConvertedOperand{Matched: true, Score: 3, TypePrecedenceScore: declaredTypePrecedence, WrappedOperand: wrapped}
				}
			}
		}
	}

	// IMPLICIT CONVERSION
	res, err := mi.IsImplicitlyConvertible(invokedType, declaredType)
	if err != nil {
		return ConvertedOperand{}, err
	}

	_, invokedIsSystem := invokedType.(types.System)
	if res.IsConvertible && invokedIsSystem {
		// Ex Integer --> Decimal   ToDecimal(operand)
		wrapped, err := wrapSystemImplicitConversion(res.Library, res.Function, opToWrap)
		if err != nil {
			return ConvertedOperand{}, err
		}

		score := implicitConversionScore(declaredType)
		if score < minConverted.Score {
			minConverted = ConvertedOperand{Matched: true, Score: score, TypePrecedenceScore: declaredTypePrecedence, WrappedOperand: wrapped}
		}
	}

	if res.IsConvertible {
		// IMPLICIT CONVERSION TO CLASS TYPE
		// EX FHIR.date --> System.Date   FHIRHelpers.ToDate(operand)
		wrapped := &model.FunctionRef{
			LibraryName: res.Library,
			Name:        res.Function,
			Operands:    []model.IExpression{opToWrap},
			Expression:  model.ResultType(res.OutputType),
		}

		score := implicitConversionScore(declaredType)
		if score < minConverted.Score {
			minConverted = ConvertedOperand{Matched: true, Score: score, TypePrecedenceScore: declaredTypePrecedence, WrappedOperand: wrapped}
		}
	}

	// IMPLICIT CONVERSION TO CLASS TYPE - Intervals and Lists
	switch i := invokedType.(type) {
	// Ex Interval<Integer> --> Interval<Decimal>   Interval[ToDecimal(operand.Low), ToDecimal(operand.High), operand.lowClosed, operand.highClosed]
	case *types.Interval:
		d, ok := declaredType.(*types.Interval)
		if !ok {
			break
		}

		low := &model.Property{Source: opToWrap, Path: "low", Expression: model.ResultType(i.PointType)}
		high := &model.Property{Source: opToWrap, Path: "high", Expression: model.ResultType(i.PointType)}
		rLow, err := OperandImplicitConverter(i.PointType, d.PointType, low, mi)
		if err != nil {
			return ConvertedOperand{}, err
		}
		if !rLow.Matched {
			break
		}
		rHigh, err := OperandImplicitConverter(i.PointType, d.PointType, high, mi)
		if err != nil {
			return ConvertedOperand{}, err
		}
		if !rHigh.Matched {
			break
		}
		wrapped := &model.Interval{
			Low:  rLow.WrappedOperand,
			High: rHigh.WrappedOperand,
			// Since operand could be any CQL expression that resolves to an interval we use the lowClosed
			// and highClosed properties to forwards the bounds of the interval.
			LowClosedExpression:  &model.Property{Source: opToWrap, Path: "lowClosed", Expression: model.ResultType(types.Boolean)},
			HighClosedExpression: &model.Property{Source: opToWrap, Path: "highClosed", Expression: model.ResultType(types.Boolean)},
			Expression:           model.ResultType(d),
		}
		if 5 < minConverted.Score {
			minConverted = ConvertedOperand{Matched: true, Score: 5, TypePrecedenceScore: declaredTypePrecedence, WrappedOperand: wrapped}
		}

	// Ex List<Integer> --> List<Decimal>   [operand] X return ToDecimal(X)
	case *types.List:
		d, ok := declaredType.(*types.List)
		if !ok {
			break
		}
		ref := &model.AliasRef{Name: "X", Expression: model.ResultType(i.ElementType)}
		r, err := OperandImplicitConverter(i.ElementType, d.ElementType, ref, mi)
		if err != nil {
			return ConvertedOperand{}, err
		}
		if !r.Matched {
			break
		}
		wrapped := &model.Query{
			Source: []*model.AliasedSource{&model.AliasedSource{
				Alias:      "X",
				Source:     opToWrap,
				Expression: model.ResultType(i),
			}},
			Return: &model.ReturnClause{
				Expression: r.WrappedOperand,
				Distinct:   false,
				Element:    &model.Element{ResultType: d.ElementType}},
			Expression: model.ResultType(declaredType),
		}
		if 5 < minConverted.Score {
			minConverted = ConvertedOperand{Matched: true, Score: 5, TypePrecedenceScore: declaredTypePrecedence, WrappedOperand: wrapped}
		}
	}

	// TODO(b/301606416): Add List Demotion (T -> List<T>) and Interval Demotion (T -> Interval<T>).

	if minConverted.Matched {
		return minConverted, nil
	}
	return ConvertedOperand{Matched: false}, nil
}

func wrapSystemImplicitConversion(library string, function string, operand model.IExpression) (model.IExpression, error) {
	if library != "SYSTEM" {
		return nil, fmt.Errorf("internal error - could not find wrapper for %v %v", library, function)
	}
	switch function {
	case "ToDecimal":
		return &model.ToDecimal{UnaryExpression: &model.UnaryExpression{Operand: operand, Expression: model.ResultType(types.Decimal)}}, nil
	case "ToLong":
		return &model.ToLong{UnaryExpression: &model.UnaryExpression{Operand: operand, Expression: model.ResultType(types.Long)}}, nil
	case "ToDateTime":
		return &model.ToDateTime{UnaryExpression: &model.UnaryExpression{Operand: operand, Expression: model.ResultType(types.DateTime)}}, nil
	case "ToQuantity":
		return &model.ToQuantity{UnaryExpression: &model.UnaryExpression{Operand: operand, Expression: model.ResultType(types.Quantity)}}, nil
	case "ToConcept":
		return &model.ToConcept{UnaryExpression: &model.UnaryExpression{Operand: operand, Expression: model.ResultType(types.Concept)}}, nil
	}
	return nil, fmt.Errorf("internal error - could not find wrapper for %v %v", library, function)
}

// getTypeCategoryPrecedence returns a score based on the type category precedence, lower score
// indicates a higher precedence.
// This function is used as a tie breaker when two conversion paths have the same conversion
// score.
func getTypeCategoryPrecedence(t types.IType) (int, error) {
	switch t.(type) {
	case types.System:
		return 1, nil
	case *types.Tuple:
		return 2, nil
	case *types.Named:
		return 3, nil
	case *types.Interval:
		return 3, nil
	case *types.List:
		return 4, nil
	case *types.Choice:
		return 5, nil
	default:
		return 0, fmt.Errorf("internal error - could not find type category precedence for %v", t)
	}
}

// OperandsToString returns a print friendly representation of the operands.
func OperandsToString(operands []model.IExpression) string {
	var stringOperands strings.Builder
	for i, operand := range operands {
		if i > 0 {
			stringOperands.WriteString(", ")
		}
		if operand == nil || operand.GetResultType() == nil {
			stringOperands.WriteString("nil")
		} else {
			stringOperands.WriteString(operand.GetResultType().String())
		}
	}
	return stringOperands.String()
}

// operandsToStringForTypes returns a string representation of type operands.
// This is similar to OperandsToString but works on IType slices instead of IExpression slices.
func operandsToStringForTypes(operands []types.IType) string {
	var stringOperands strings.Builder
	for i, operand := range operands {
		if i > 0 {
			stringOperands.WriteString(", ")
		}
		if operand == nil {
			stringOperands.WriteString("null")
		} else {
			stringOperands.WriteString(operand.String())
		}
	}
	return stringOperands.String()
}

// OperandsToTypes returns the types of the ResultType of the operands.
func OperandsToTypes(operands []model.IExpression) []types.IType {
	var types []types.IType
	for _, operand := range operands {
		types = append(types, operand.GetResultType())
	}
	return types
}

func implicitConversionScore(t types.IType) int {
	switch t {
	case types.String, types.Integer, types.Long, types.Decimal, types.Boolean, types.Date, types.DateTime, types.Time:
		// Simple types.
		return 4
	default:
		// Everything else is a class type.
		return 5
	}
}

// Generic types are used to define generic overloads, the overloads with T in
// https://cql.hl7.org/09-b-cqlreference.html. The only place generics are used is to define system
// operators in the Parser. The ResultType in model should never be a Generic type. The interpreter
// should never deal with generic types.
type Generic string

const (
	// GenericType represents a generic CQL type, shown as T in the CQL reference. Never nest a
	// GenericType in a real type (ex List<GenericType>). Instead use the GenericList below.
	GenericType Generic = "GenericType"
	// GenericInterval represents a generic interval type, shown as Interval<T> in the CQL reference.
	GenericInterval Generic = "GenericInterval"
	// GenericList represents a generic list type, shown as List<T> in the CQL reference.
	GenericList Generic = "GenericList"
)

// Equal is a strict equal. X.Equal(Y) is true when X and Y are the exact same types.
func (s Generic) Equal(a types.IType) bool {
	aBase, ok := a.(Generic)
	if !ok {
		return false
	}
	return s == aBase
}

// String returns the model info based name for the type, and implements fmt.Stringer for easy
// printing.
func (s Generic) String() string {
	return fmt.Sprintf("Generic.%v", string(s))
}

// ModelInfoName should never be called for Generics.
func (s Generic) ModelInfoName() (string, error) {
	return "", errors.New("Generic type does not have a model info name")
}

// MarshalJSON should never be called for Generics.
func (s Generic) MarshalJSON() ([]byte, error) {
	return nil, errors.New("Generics should not be marshalled")
}

// convertGeneric takes the invoked operands ex (Integer, String, Decimal), a generic overload (T,
// String, T) and returns the least converting concrete overload that still satisfies the the
// generic constraints, in this case (Decimal, String, Decimal). If there is no concrete
// instantiation of this generic overload that will work false is returned.
func convertGeneric[F any](invoked []model.IExpression, genericDeclared Overload[F], mi *modelinfo.ModelInfos) (Overload[F], bool, error) {
	if len(invoked) != len(genericDeclared.Operands) {
		// There is no concrete instantiation of this generic overload that will work, so return false.
		return Overload[F]{}, false, nil
	}

	// genericInvokedTypes holds the types of the invoked operands that need to match with a generic.
	// For example for invoked (Integer, String, Decimal), and genericDeclared (T, String, T) then
	// genericInvokedTypes is (Integer, Decimal). We use inferMixedType to find the least converting T
	// for genericInvokedTypes.
	genericInvokedTypes := make([]types.IType, 0)
	for i := range invoked {
		switch genericDeclared.Operands[i] {
		case GenericType:
			genericInvokedTypes = append(genericInvokedTypes, invoked[i].GetResultType())
		case GenericInterval:
			if interval, ok := invoked[i].GetResultType().(*types.Interval); ok {
				genericInvokedTypes = append(genericInvokedTypes, interval.PointType)
			} else {
				// If we are matching Interval<T> and invoked[i] is not an interval, we will either need to
				// apply interval promotion or check that there's an implicit conversion to Interval<T>.
				// For now, only the interval promotion case is supported, so we add invoked[i] as is to
				// find the least converting path to the T in Interval<T>.
				// TODO(b/333923412): add in full support for the cases in which there's an implicit
				// conversion to an Interval<T>.
				genericInvokedTypes = append(genericInvokedTypes, invoked[i].GetResultType())
			}
		case GenericList:
			if list, ok := invoked[i].GetResultType().(*types.List); ok {
				genericInvokedTypes = append(genericInvokedTypes, list.ElementType)
			} else {
				genericInvokedTypes = append(genericInvokedTypes, invoked[i].GetResultType())
			}
		}
	}

	// TODO(b/301606416): We should convert the generic overload into all least converting concrete
	// overloads, but if there is a tie inferMixedType just randomly picks one.
	inferred, err := inferMixedType(genericInvokedTypes, nil, mi)
	if err != nil {
		return Overload[F]{}, false, err
	}

	if inferred.PuntedToChoice {
		// There is no concrete instantiation of this generic overload that will work, so return false.
		return Overload[F]{}, false, nil
	}

	// concreteOverload is the genericDeclared overload with T replaced by the inferred.UniformType.
	concreteOverload := make([]types.IType, len(genericDeclared.Operands))
	for i := range genericDeclared.Operands {
		switch genericDeclared.Operands[i] {
		case GenericType:
			concreteOverload[i] = inferred.UniformType
		case GenericInterval:
			if _, ok := inferred.UniformType.(*types.Interval); !ok {
				// Since we sometimes use the PointType above we need to rewrap in an Interval.
				concreteOverload[i] = &types.Interval{PointType: inferred.UniformType}
			} else {
				concreteOverload[i] = inferred.UniformType
			}
		case GenericList:
			// Wrap the inferred type T in a List.
			concreteOverload[i] = &types.List{ElementType: inferred.UniformType}
		default:
			concreteOverload[i] = genericDeclared.Operands[i]
		}
	}

	genericDeclared.Operands = concreteOverload
	return genericDeclared, true, nil
}

func isGeneric(operands []types.IType) bool {
	for _, operand := range operands {
		if operand.Equal(GenericType) || operand.Equal(GenericInterval) || operand.Equal(GenericList) {
			return true
		}
	}
	return false
}
