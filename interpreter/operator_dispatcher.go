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
	"fmt"

	"github.com/google/cql/internal/convert"
	"github.com/google/cql/model"
	"github.com/google/cql/result"
	"github.com/google/cql/types"
)

func (i *interpreter) evalUnaryExpression(m model.IUnaryExpression) (result.Value, error) {
	// Evaluate Operand
	operand, err := i.evalExpression(m.GetOperand())
	if err != nil {
		return result.Value{}, err
	}

	// Match Overload
	overloads, err := i.unaryOverloads(m)
	if err != nil {
		return result.Value{}, err
	}

	evalFunc, err := convert.ExactOverloadMatch[evalUnarySignature]([]types.IType{m.GetOperand().GetResultType()}, overloads, i.modelInfo, m.GetName())
	if err != nil {
		return result.Value{}, err
	}

	// Evaluate the Overload
	res, err := evalFunc(m, operand)
	if err != nil {
		return result.Value{}, err
	}
	return res.WithSources(m, operand), nil
}

func (i *interpreter) evalBinaryExpression(m model.IBinaryExpression) (result.Value, error) {
	// Evaluate Operands
	l, err := i.evalExpression(m.Left())
	if err != nil {
		return result.Value{}, err
	}
	r, err := i.evalExpression(m.Right())
	if err != nil {
		return result.Value{}, err
	}

	// Match Overload
	overloads, err := i.binaryOverloads(m)
	if err != nil {
		return result.Value{}, err
	}

	evalFunc, err := convert.ExactOverloadMatch[evalBinarySignature]([]types.IType{m.Left().GetResultType(), m.Right().GetResultType()}, overloads, i.modelInfo, m.GetName())
	if err != nil {
		return result.Value{}, err
	}

	// Evaluate the Overload
	res, err := evalFunc(m, l, r)
	if err != nil {
		return result.Value{}, err
	}

	return res.WithSources(m, l, r), nil
}

func (i *interpreter) evalNaryExpression(m model.INaryExpression) (result.Value, error) {
	// Evaluate Operands
	evalOps := make([]result.Value, len(m.GetOperands()))
	for idx, operand := range m.GetOperands() {
		operand, err := i.evalExpression(operand)
		if err != nil {
			return result.Value{}, err
		}
		evalOps[idx] = operand
	}

	// Match Overloads
	overloads, err := i.naryOverloads(m)
	if err != nil {
		return result.Value{}, err
	}

	evalFunc, err := convert.ExactOverloadMatch(convert.OperandsToTypes(m.GetOperands()), overloads, i.modelInfo, m.GetName())
	if err != nil {
		return result.Value{}, err
	}

	// Evaluate the Overload
	res, err := evalFunc(m, evalOps)
	if err != nil {
		return result.Value{}, err
	}

	return res.WithSources(m, evalOps...), nil
}

type evalUnarySignature func(model.IUnaryExpression, result.Value) (result.Value, error)
type evalBinarySignature func(model.IBinaryExpression, result.Value, result.Value) (result.Value, error)
type evalNarySignature func(model.INaryExpression, []result.Value) (result.Value, error)

func (i *interpreter) unaryOverloads(m model.IUnaryExpression) ([]convert.Overload[evalUnarySignature], error) {
	switch m.(type) {
	case *model.Abs:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{types.Decimal},
				Result:   evalAbsDecimal,
			},
			{
				Operands: []types.IType{types.Integer},
				Result:   evalAbsInteger,
			},
			{
				Operands: []types.IType{types.Long},
				Result:   evalAbsLong,
			},
			{
				Operands: []types.IType{types.Quantity},
				Result:   evalAbsQuantity,
			},
		}, nil
	case *model.Ceiling:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{types.Decimal},
				Result:   evalCeiling,
			},
		}, nil
	case *model.Exp:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{types.Decimal},
				Result:   evalExpDecimal,
			},
		}, nil
	case *model.Floor:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{types.Decimal},
				Result:   evalFloor,
			},
		}, nil
	case *model.Ln:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{types.Decimal},
				Result:   evalLn,
			},
		}, nil
	case *model.Precision:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{types.Date},
				Result:   evalPrecisionDateTime,
			},
			{
				Operands: []types.IType{types.DateTime},
				Result:   evalPrecisionDateTime,
			},
			{
				Operands: []types.IType{types.Time},
				Result:   evalPrecisionTime,
			},
		}, nil
	case *model.Exists:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{&types.List{ElementType: types.Any}},
				Result:   evalExists,
			},
		}, nil
	case *model.Distinct:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{&types.List{ElementType: types.Any}},
				Result:   evalDistinct,
			},
		}, nil
	case *model.First:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{&types.List{ElementType: types.Any}},
				Result:   evalFirst,
			},
		}, nil
	case *model.Last:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{&types.List{ElementType: types.Any}},
				Result:   evalLast,
			},
		}, nil
	case *model.Length:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{types.String},
				Result: evalLengthString,
			},
			{
				Operands: []types.IType{&types.List{ElementType: types.Any}},
				Result:   evalLengthList,
			},
		}, nil
	case *model.As:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{types.Any},
				Result:   i.evalAs,
			},
		}, nil
	case *model.ToDateTime:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{types.DateTime},
				Result:   evalToDateTimeDate,
			},
			{
				Operands: []types.IType{types.Date},
				Result:   evalToDateTimeDate,
			},
			{
				Operands: []types.IType{types.String},
				Result:   i.evalToDateTimeString,
			},
		}, nil
	case *model.ToDate:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{types.Date},
				Result:   evalToDateDateTime,
			},
			{
				Operands: []types.IType{types.DateTime},
				Result:   evalToDateDateTime,
			},
			{
				Operands: []types.IType{types.String},
				Result:   i.evalToDateString,
			},
		}, nil
	case *model.ToDecimal:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{types.Decimal},
				Result:   evalToDecimal,
			},
			{
				Operands: []types.IType{types.Long},
				Result:   evalToDecimal,
			},
			{
				Operands: []types.IType{types.Integer},
				Result:   evalToDecimal,
			},
			{
				Operands: []types.IType{types.String},
				Result:   evalToDecimalString,
			},
			{
				Operands: []types.IType{types.Boolean},
				Result:   evalToDecimalBoolean,
			},
		}, nil
	case *model.ToLong:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{types.Long},
				Result:   evalToLong,
			},
			{
				Operands: []types.IType{types.Integer},
				Result:   evalToLong,
			},
			{
				Operands: []types.IType{types.String},
				Result:   evalToLongString,
			},
			{
				Operands: []types.IType{types.Boolean},
				Result:   evalToLongBoolean,
			},
		}, nil
	case *model.ToQuantity:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{types.Decimal},
				Result:   evalToQuantity,
			},
			{
				Operands: []types.IType{types.Integer},
				Result:   evalToQuantity,
			},
			{
				Operands: []types.IType{types.String},
				Result:   evalToQuantityString,
			},
		}, nil
	case *model.ToConcept:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{types.Code},
				Result:   evalToConceptCode,
			},
			{
				Operands: []types.IType{&types.List{ElementType: types.Code}},
				Result:   evalToConceptList,
			},
		}, nil
	case *model.ToString:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{types.Any},
				Result:   evalToString,
			},
			{
				Operands: []types.IType{types.Boolean},
				Result:   evalToString,
			},
			{
				Operands: []types.IType{types.Integer},
				Result:   evalToString,
			},
			{
				Operands: []types.IType{types.Long},
				Result:   evalToString,
			},
			{
				Operands: []types.IType{types.Decimal},
				Result:   evalToString,
			},
			{
				Operands: []types.IType{types.Quantity},
				Result:   evalToString,
			},
			{
				Operands: []types.IType{types.Ratio},
				Result:   evalToString,
			},
			{
				Operands: []types.IType{types.Date},
				Result:   evalToString,
			},
			{
				Operands: []types.IType{types.DateTime},
				Result:   evalToString,
			},
			{
				Operands: []types.IType{types.Time},
				Result:   evalToString,
			},
		}, nil
	case *model.End:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{&types.Interval{PointType: types.Any}},
				Result:   i.evalEnd,
			},
		}, nil
	case *model.Start:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{&types.Interval{PointType: types.Any}},
				Result:   i.evalStart,
			},
		}, nil
	case *model.SingletonFrom:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{&types.List{ElementType: types.Any}},
				Result:   evalSingletonFrom,
			},
		}, nil
	case *model.Is:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{types.Any},
				Result:   i.evalIs,
			},
		}, nil
	case *model.IsNull:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{types.Any},
				Result:   evalIsNull,
			},
		}, nil
	case *model.IsTrue:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{types.Any},
				Result:   evalIsTrue,
			},
		}, nil
	case *model.IsFalse:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{types.Any},
				Result:   evalIsFalse,
			},
		}, nil
	case *model.Not:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{types.Boolean},
				Result:   evalNot,
			},
		}, nil
	case *model.Negate:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{types.Integer},
				Result:   evalNegateInteger,
			},
			{
				Operands: []types.IType{types.Long},
				Result:   evalNegateLong,
			},
			{
				Operands: []types.IType{types.Decimal},
				Result:   evalNegateDecimal,
			},
			{
				Operands: []types.IType{types.Quantity},
				Result:   evalNegateQuantity,
			},
		}, nil
	case *model.Truncate:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{types.Decimal},
				Result:   evalTruncate,
			},
		}, nil
	case *model.Predecessor:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{types.Integer},
				Result:   i.evalPredecessor,
			},
			{
				Operands: []types.IType{types.Long},
				Result:   i.evalPredecessor,
			},
			{
				Operands: []types.IType{types.Decimal},
				Result:   i.evalPredecessor,
			},
			{
				Operands: []types.IType{types.Quantity},
				Result:   i.evalPredecessor,
			},
			{
				Operands: []types.IType{types.Date},
				Result:   i.evalPredecessor,
			},
			{
				Operands: []types.IType{types.DateTime},
				Result:   i.evalPredecessor,
			},
			{
				Operands: []types.IType{types.Time},
				Result:   i.evalPredecessor,
			},
		}, nil
	case *model.Successor:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{types.Integer},
				Result:   i.evalSuccessor,
			},
			{
				Operands: []types.IType{types.Long},
				Result:   i.evalSuccessor,
			},
			{
				Operands: []types.IType{types.Decimal},
				Result:   i.evalSuccessor,
			},
			{
				Operands: []types.IType{types.Quantity},
				Result:   i.evalSuccessor,
			},
			{
				Operands: []types.IType{types.Date},
				Result:   i.evalSuccessor,
			},
			{
				Operands: []types.IType{types.DateTime},
				Result:   i.evalSuccessor,
			},
			{
				Operands: []types.IType{types.Time},
				Result:   i.evalSuccessor,
			},
		}, nil
	case *model.AllTrue:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{&types.List{ElementType: types.Boolean}},
				Result:   i.evalAllTrue,
			},
		}, nil
	case *model.AnyTrue:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{&types.List{ElementType: types.Boolean}},
				Result:   i.evalAnyTrue,
			},
		}, nil
	case *model.Avg:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{&types.List{ElementType: types.Decimal}},
				Result:   i.evalAvg,
			},
			{
				Operands: []types.IType{&types.List{ElementType: types.Quantity}},
				Result:   i.evalAvg,
			},
		}, nil
	case *model.Count:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{&types.List{ElementType: types.Any}},
				Result:   i.evalCount,
			},
		}, nil
	case *model.Max:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{&types.List{ElementType: types.Date}},
				Result:   i.evalMaxDateTime,
			},
			{
				Operands: []types.IType{&types.List{ElementType: types.DateTime}},
				Result:   i.evalMaxDateTime,
			},
		}, nil
	case *model.Min:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{&types.List{ElementType: types.Date}},
				Result:   i.evalMinDateTime,
			},
			{
				Operands: []types.IType{&types.List{ElementType: types.DateTime}},
				Result:   i.evalMinDateTime,
			},
		}, nil
	case *model.Sum:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{&types.List{ElementType: types.Decimal}},
				Result:   i.evalSum,
			},
			{
				Operands: []types.IType{&types.List{ElementType: types.Integer}},
				Result:   i.evalSum,
			},
			{
				Operands: []types.IType{&types.List{ElementType: types.Long}},
				Result:   i.evalSum,
			},
			{
				Operands: []types.IType{&types.List{ElementType: types.Quantity}},
				Result:   i.evalSum,
			},
		}, nil
	case *model.Median:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{&types.List{ElementType: types.Decimal}},
				Result:   i.evalMedianDecimal,
			},
			{
				Operands: []types.IType{&types.List{ElementType: types.Quantity}},
				Result:   i.evalMedianQuantity,
			},
		}, nil
	case *model.PopulationStdDev:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{&types.List{ElementType: types.Decimal}},
				Result:   i.evalPopulationStdDevDecimal,
			},
			{
				Operands: []types.IType{&types.List{ElementType: types.Quantity}},
				Result:   i.evalPopulationStdDevQuantity,
			},
		}, nil
	case *model.Flatten:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{&types.List{ElementType: &types.List{ElementType: types.Any}}},
				Result:   evalFlatten,
			},
		}, nil
	case *model.Tail:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{&types.List{ElementType: types.Any}},
				Result:   evalTail,
			},
		}, nil
	case *model.Upper:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{types.String},
				Result: evalUpper,
			},
		}, nil
	case *model.Lower:
		return []convert.Overload[evalUnarySignature]{
			{
				Operands: []types.IType{types.String},
				Result: evalLower,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported Unary Expression %v", m.GetName())
	}
}

// TODO(b/312172420): Move BinaryOverloads and UnaryOverloads to their own files.
func (i *interpreter) binaryOverloads(m model.IBinaryExpression) ([]convert.Overload[evalBinarySignature], error) {
	switch m.(type) {
	case *model.Add, *model.Subtract:
		return []convert.Overload[evalBinarySignature]{
			{
				Operands: []types.IType{types.Integer, types.Integer},
				Result:   evalArithmeticInteger,
			},
			{
				Operands: []types.IType{types.Long, types.Long},
				Result:   evalArithmeticLong,
			},
			{
				Operands: []types.IType{types.Decimal, types.Decimal},
				Result:   evalArithmeticDecimal,
			},
			{
				Operands: []types.IType{types.Quantity, types.Quantity},
				Result:   evalArithmeticQuantity,
			},
			{
				Operands: []types.IType{types.Date, types.Quantity},
				Result:   evalArithmeticDate,
			},
			{
				Operands: []types.IType{types.DateTime, types.Quantity},
				Result:   evalArithmeticDateTime,
			},
		}, nil
	case *model.Multiply, *model.TruncatedDivide, *model.Modulo:
		return []convert.Overload[evalBinarySignature]{
			{
				Operands: []types.IType{types.Integer, types.Integer},
				Result:   evalArithmeticInteger,
			},
			{
				Operands: []types.IType{types.Long, types.Long},
				Result:   evalArithmeticLong,
			},
			{
				Operands: []types.IType{types.Decimal, types.Decimal},
				Result:   evalArithmeticDecimal,
			},
			{
				Operands: []types.IType{types.Quantity, types.Quantity},
				Result:   evalArithmeticQuantity,
			},
		}, nil
	case *model.Power:
		return []convert.Overload[evalBinarySignature]{
			{
				Operands: []types.IType{types.Integer, types.Integer},
				Result:   evalPower,
			},
			{
				Operands: []types.IType{types.Long, types.Long},
				Result:   evalPower,
			},
			{
				Operands: []types.IType{types.Decimal, types.Decimal},
				Result:   evalPower,
			},
			{
				Operands: []types.IType{types.Quantity, types.Quantity},
				Result:   evalPower,
			},
		}, nil
	case *model.Divide:
		return []convert.Overload[evalBinarySignature]{
			{
				Operands: []types.IType{types.Decimal, types.Decimal},
				Result:   evalArithmeticDecimal,
			},
			{
				Operands: []types.IType{types.Quantity, types.Quantity},
				Result:   evalArithmeticQuantity,
			},
		}, nil
	case *model.Log:
		return []convert.Overload[evalBinarySignature]{
			{
				Operands: []types.IType{types.Decimal, types.Decimal},
				Result:   evalLog,
			},
		}, nil
	case *model.And, *model.Or, *model.XOr, *model.Implies:
		return []convert.Overload[evalBinarySignature]{
			{
				Operands: []types.IType{types.Boolean, types.Boolean},
				Result:   evalLogic,
			},
		}, nil
	case *model.Equal:
		return []convert.Overload[evalBinarySignature]{
			{
				Operands: []types.IType{types.Any, types.Any},
				Result:   i.evalEqual,
			},
			{
				Operands: []types.IType{types.DateTime, types.DateTime},
				Result:   evalEqualDateTime,
			},
			{
				Operands: []types.IType{types.Date, types.Date},
				Result:   evalEqualDateTime,
			},
		}, nil
	case *model.Equivalent:
		// TODO(b/301606416): Expand equivalent support to all types.
		// All equivalent overloads must be resilient to a nil model input.
		return []convert.Overload[evalBinarySignature]{
			{
				Operands: []types.IType{types.Boolean, types.Boolean},
				Result:   evalEquivalentSimpleType,
			},
			{
				Operands: []types.IType{types.Integer, types.Integer},
				Result:   evalEquivalentSimpleType,
			},
			{
				Operands: []types.IType{types.Long, types.Long},
				Result:   evalEquivalentSimpleType,
			},
			{
				Operands: []types.IType{types.String, types.String},
				Result:   evalEquivalentString,
			},
			{
				Operands: []types.IType{types.DateTime, types.DateTime},
				Result:   evalEquivalentDateTime,
			},
			{
				Operands: []types.IType{types.Date, types.Date},
				Result:   evalEquivalentDateTime,
			},
			// The parser will make sure the List<T>, List<T> have correctly matching or converted T.
			{
				Operands: []types.IType{&types.List{ElementType: types.Any}, &types.List{ElementType: types.Any}},
				Result:   i.evalEquivalentList,
			},
			// The parser will make sure the Interval<T>, Interval<T> have correctly matching or converted T.
			{
				Operands: []types.IType{&types.Interval{PointType: types.Any}, &types.Interval{PointType: types.Any}},
				Result:   i.evalEquivalentInterval,
			},
			{
				Operands: []types.IType{types.Concept, types.Code},
				Result:   i.evalEquivalentConceptCode,
			},
			{
				Operands: []types.IType{types.Code, types.Code},
				Result:   i.evalEquivalentCodeCode,
			},
		}, nil
	case *model.Less, *model.LessOrEqual, *model.Greater, *model.GreaterOrEqual:
		return []convert.Overload[evalBinarySignature]{
			{
				Operands: []types.IType{types.Integer, types.Integer},
				Result:   evalCompareInteger,
			},
			{
				Operands: []types.IType{types.Long, types.Long},
				Result:   evalCompareLong,
			},
			{
				Operands: []types.IType{types.Decimal, types.Decimal},
				Result:   evalCompareDecimal,
			},
			{
				Operands: []types.IType{types.String, types.String},
				Result:   evalCompareString,
			},
			{
				Operands: []types.IType{types.Date, types.Date},
				Result:   evalCompareDateTime,
			},
			{
				Operands: []types.IType{types.DateTime, types.DateTime},
				Result:   evalCompareDateTime,
			},
		}, nil
	case *model.After, *model.Before, *model.SameOrAfter, *model.SameOrBefore:
		return []convert.Overload[evalBinarySignature]{
			{
				Operands: []types.IType{types.Date, types.Date},
				Result:   evalCompareDateWithPrecision,
			},
			{
				Operands: []types.IType{types.DateTime, types.DateTime},
				Result:   evalCompareDateTimeWithPrecision,
			},
			{
				Operands: []types.IType{types.Date, &types.Interval{PointType: types.Date}},
				Result:   i.evalCompareDateTimeInterval,
			},
			{
				Operands: []types.IType{types.DateTime, &types.Interval{PointType: types.DateTime}},
				Result:   i.evalCompareDateTimeInterval,
			},
			{
				Operands: []types.IType{&types.Interval{PointType: types.Date}, &types.Interval{PointType: types.Date}},
				Result:   i.evalCompareIntervalDateTimeInterval,
			},
			{
				Operands: []types.IType{&types.Interval{PointType: types.DateTime}, &types.Interval{PointType: types.DateTime}},
				Result:   i.evalCompareIntervalDateTimeInterval,
			},
		}, nil
	case *model.Overlaps:
		return []convert.Overload[evalBinarySignature]{
			{
				Operands: []types.IType{&types.Interval{PointType: types.Date}, &types.Interval{PointType: types.Date}},
				Result:   i.evalOverlapsIntervalDateTimeInterval,
			},
			{
				Operands: []types.IType{&types.Interval{PointType: types.DateTime}, &types.Interval{PointType: types.DateTime}},
				Result:   i.evalOverlapsIntervalDateTimeInterval,
			},
		}, nil
	case *model.CanConvertQuantity:
		return []convert.Overload[evalBinarySignature]{
			{
				Operands: []types.IType{types.Quantity, types.String},
				Result:   evalCanConvertQuantity,
			},
		}, nil
	case *model.DifferenceBetween:
		return []convert.Overload[evalBinarySignature]{
			{
				Operands: []types.IType{types.Date, types.Date},
				Result:   evalDifferenceBetweenDate,
			},
			{
				Operands: []types.IType{types.DateTime, types.DateTime},
				Result:   evalDifferenceBetweenDateTime,
			},
		}, nil
	case *model.In:
		// TODO(b/301606416): Support all other In operator overloads.
		return []convert.Overload[evalBinarySignature]{
			{
				Operands: []types.IType{types.Any, &types.List{ElementType: types.Any}},
				Result:   evalInList,
			},
			{
				Operands: []types.IType{types.Decimal, &types.Interval{PointType: types.Decimal}},
				Result:   evalInIntervalNumeral,
			},
			{
				Operands: []types.IType{types.Long, &types.Interval{PointType: types.Long}},
				Result:   evalInIntervalNumeral,
			},
			{
				Operands: []types.IType{types.Integer, &types.Interval{PointType: types.Integer}},
				Result:   evalInIntervalNumeral,
			},
			{
				Operands: []types.IType{types.Quantity, &types.Interval{PointType: types.Quantity}},
				Result:   evalInIntervalNumeral,
			},
			{
				Operands: []types.IType{types.Date, &types.Interval{PointType: types.Date}},
				Result:   i.evalInIntervalDateTime,
			},
			{
				Operands: []types.IType{types.DateTime, &types.Interval{PointType: types.DateTime}},
				Result:   i.evalInIntervalDateTime,
			},
		}, nil
	case *model.InCodeSystem:
		return []convert.Overload[evalBinarySignature]{
			{
				Operands: []types.IType{types.Code, types.CodeSystem},
				Result:   i.evalInCodeSystem,
			},
			{
				Operands: []types.IType{&types.List{ElementType: types.Code}, types.CodeSystem},
				Result:   i.evalInCodeSystem,
			},
			{
				Operands: []types.IType{types.Concept, types.CodeSystem},
				Result:   i.evalInCodeSystem,
			},
			{
				Operands: []types.IType{&types.List{ElementType: types.Concept}, types.CodeSystem},
				Result:   i.evalInCodeSystem,
			},
		}, nil
	case *model.InValueSet:
		return []convert.Overload[evalBinarySignature]{
			{
				Operands: []types.IType{types.Code, types.ValueSet},
				Result:   i.evalInValueSet,
			},
			{
				Operands: []types.IType{&types.List{ElementType: types.Code}, types.ValueSet},
				Result:   i.evalInValueSet,
			},
			{
				Operands: []types.IType{types.Concept, types.ValueSet},
				Result:   i.evalInValueSet,
			},
			{
				Operands: []types.IType{&types.List{ElementType: types.Concept}, types.ValueSet},
				Result:   i.evalInValueSet,
			},
		}, nil
	case *model.CalculateAgeAt:
		return []convert.Overload[evalBinarySignature]{
			{
				Operands: []types.IType{types.Date, types.Date},
				Result:   evalCalculateAgeAtDate,
			},
			{
				Operands: []types.IType{types.DateTime, types.DateTime},
				Result:   evalCalculateAgeAtDateTime,
			},
		}, nil
	case *model.Split:
		return []convert.Overload[evalBinarySignature]{
			{
				Operands: []types.IType{types.String, types.String},
				Result:   i.evalSplit,
			},
		}, nil
	case *model.Includes:
		return []convert.Overload[evalBinarySignature]{
			{
				Operands: []types.IType{&types.List{ElementType: types.Any}, types.Any},
				Result:   evalIncludes,
			},
			{
				Operands: []types.IType{&types.List{ElementType: types.Any}, &types.List{ElementType: types.Any}},
				Result:   evalIncludesList,
			},
		}, nil
	case *model.Indexer:
		return []convert.Overload[evalBinarySignature]{
			{
				Operands: []types.IType{types.String, types.Integer},
				Result:   i.evalIndexerString,
			},
			{
				Operands: []types.IType{&types.List{ElementType: types.Any}, types.Integer},
				Result:   i.evalIndexerList,
			},
		}, nil
	case *model.IndexOf:
		return []convert.Overload[evalBinarySignature]{
			{
				Operands: []types.IType{&types.List{ElementType: types.Any}, types.Any},
				Result:   evalIndexOf,
			},
		}, nil
	case *model.Except:
		return []convert.Overload[evalBinarySignature]{
			{
				Operands: []types.IType{&types.List{ElementType: types.Any}, &types.List{ElementType: types.Any}},
				Result:   evalExcept,
			},
		}, nil
	case *model.Intersect:
		return []convert.Overload[evalBinarySignature]{
			{
				Operands: []types.IType{&types.List{ElementType: types.Any}, &types.List{ElementType: types.Any}},
				Result:   evalIntersect,
			},
		}, nil
	case *model.ProperlyIncludes:
		return []convert.Overload[evalBinarySignature]{
			{
				Operands: []types.IType{&types.List{ElementType: types.Any}, types.Any},
				Result:   evalProperlyIncludes,
			},
			{
				Operands: []types.IType{&types.List{ElementType: types.Any}, &types.List{ElementType: types.Any}},
				Result:   evalProperlyIncludesList,
			},
		}, nil
	case *model.Skip:
		return []convert.Overload[evalBinarySignature]{
			{
				Operands: []types.IType{&types.List{ElementType: types.Any}, types.Integer},
				Result:   evalSkip,
			},
		}, nil
	case *model.Take:
		return []convert.Overload[evalBinarySignature]{
			{
				Operands: []types.IType{&types.List{ElementType: types.Any}, types.Integer},
				Result:   evalTake,
			},
		}, nil
	case *model.EndsWith:
		return []convert.Overload[evalBinarySignature]{
			{
				Operands: []types.IType{types.String, types.String},
				Result:   evalEndsWith,
			},
		}, nil
	case *model.LastPositionOf:
		return []convert.Overload[evalBinarySignature]{
			{
				Operands: []types.IType{types.String, types.String},
				Result:   evalLastPositionOf,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported Binary Expression %v", m.GetName())
	}
}

func (i *interpreter) naryOverloads(m model.INaryExpression) ([]convert.Overload[evalNarySignature], error) {
	switch m.(type) {
	case *model.Date:
		return []convert.Overload[evalNarySignature]{
			{
				Operands: []types.IType{types.Integer},
				Result:   i.evalDate,
			},
			{
				Operands: []types.IType{types.Integer, types.Integer},
				Result:   i.evalDate,
			},
			{
				Operands: []types.IType{types.Integer, types.Integer, types.Integer},
				Result:   i.evalDate,
			},
		}, nil
	case *model.DateTime:
		return []convert.Overload[evalNarySignature]{
			{
				Operands: []types.IType{types.Integer},
				Result:   i.evalDateTime,
			},
			{
				Operands: []types.IType{types.Integer, types.Integer},
				Result:   i.evalDateTime,
			},
			{
				Operands: []types.IType{types.Integer, types.Integer, types.Integer},
				Result:   i.evalDateTime,
			},
			{
				Operands: []types.IType{types.Integer, types.Integer, types.Integer, types.Integer},
				Result:   i.evalDateTime,
			},
			{
				Operands: []types.IType{types.Integer, types.Integer, types.Integer, types.Integer, types.Integer},
				Result:   i.evalDateTime,
			},
			{
				Operands: []types.IType{types.Integer, types.Integer, types.Integer, types.Integer, types.Integer, types.Integer},
				Result:   i.evalDateTime,
			},
			{
				Operands: []types.IType{types.Integer, types.Integer, types.Integer, types.Integer, types.Integer, types.Integer, types.Integer},
				Result:   i.evalDateTime,
			},
			{
				Operands: []types.IType{types.Integer, types.Integer, types.Integer, types.Integer, types.Integer, types.Integer, types.Integer, types.Decimal},
				Result:   i.evalDateTime,
			},
		}, nil
	case *model.Time:
		return []convert.Overload[evalNarySignature]{
			{
				Operands: []types.IType{types.Integer},
				Result:   i.evalTime,
			},
			{
				Operands: []types.IType{types.Integer, types.Integer},
				Result:   i.evalTime,
			},
			{
				Operands: []types.IType{types.Integer, types.Integer, types.Integer},
				Result:   i.evalTime,
			},
			{
				Operands: []types.IType{types.Integer, types.Integer, types.Integer, types.Integer},
				Result:   i.evalTime,
			},
		}, nil
	case *model.Now:
		return []convert.Overload[evalNarySignature]{
			{
				Operands: []types.IType{},
				Result:   i.evalNow,
			},
		}, nil
	case *model.TimeOfDay:
		return []convert.Overload[evalNarySignature]{
			{
				Operands: []types.IType{},
				Result:   i.evalTimeOfDay,
			},
		}, nil
	case *model.Today:
		return []convert.Overload[evalNarySignature]{
			{
				Operands: []types.IType{},
				Result:   i.evalToday,
			},
		}, nil
	case *model.Coalesce:
		return []convert.Overload[evalNarySignature]{
			{
				Operands: []types.IType{types.Any, types.Any},
				Result:   evalCoalesce,
			},
			{
				Operands: []types.IType{types.Any, types.Any, types.Any},
				Result:   evalCoalesce,
			},
			{
				Operands: []types.IType{types.Any, types.Any, types.Any, types.Any},
				Result:   evalCoalesce,
			},
			{
				Operands: []types.IType{types.Any, types.Any, types.Any, types.Any, types.Any},
				Result:   evalCoalesce,
			},
			{
				Operands: []types.IType{&types.List{ElementType: types.Any}},
				Result:   evalCoalesceList,
			},
		}, nil
	case *model.Concatenate:
		return []convert.Overload[evalNarySignature]{
			{
				Operands: []types.IType{types.String, types.String},
				Result:   evalConcatenate,
			},
		}, nil
	case *model.Combine:
		return []convert.Overload[evalNarySignature]{
			{
				Operands: []types.IType{&types.List{ElementType: types.String}},
				Result:   i.evalCombine,
			},
			{
				Operands: []types.IType{&types.List{ElementType: types.String}, types.String},
				Result:   i.evalCombine,
			},
		}, nil
	case *model.Round:
		return []convert.Overload[evalNarySignature]{
			{
				Operands: []types.IType{types.Decimal},
				Result:   evalRound,
			},
			{
				Operands: []types.IType{types.Decimal, types.Integer},
				Result:   evalRound,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported Nary Expression %v", m.GetName())
	}
}
