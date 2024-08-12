// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package interpreter

import (
	"fmt"
	"sort"

	"github.com/google/cql/model"
	"github.com/google/cql/result"
	"github.com/google/cql/types"
)

// AGGREGATE FUNCTIONS - https://cql.hl7.org/09-b-cqlreference.html#aggregate-functions

// AllTrue(argument List<Boolean>) Boolean
// https://cql.hl7.org/09-b-cqlreference.html#alltrue
func (i *interpreter) evalAllTrue(m model.IUnaryExpression, operand result.Value) (result.Value, error) {
	if result.IsNull(operand) {
		return result.New(true)
	}
	l, err := result.ToSlice(operand)
	if err != nil {
		return result.Value{}, err
	}
	for _, elem := range l {
		if result.IsNull(elem) {
			continue
		}
		bv, err := result.ToBool(elem)
		if err != nil {
			return result.Value{}, err
		}
		if !bv {
			return result.New(false)
		}
	}
	return result.New(true)
}

// AnyTrue(argument List<Boolean>) Boolean
// https://cql.hl7.org/09-b-cqlreference.html#anytrue
func (i *interpreter) evalAnyTrue(m model.IUnaryExpression, operand result.Value) (result.Value, error) {
	if result.IsNull(operand) {
		return result.New(false)
	}
	l, err := result.ToSlice(operand)
	if err != nil {
		return result.Value{}, err
	}
	for _, elem := range l {
		if result.IsNull(elem) {
			continue
		}
		bv, err := result.ToBool(elem)
		if err != nil {
			return result.Value{}, err
		}
		if bv {
			return result.New(true)
		}
	}
	return result.New(false)
}

// Avg(argument List<Decimal>) Decimal
// Avg(argument List<Quantity>) Quantity
// https://cql.hl7.org/09-b-cqlreference.html#avg
func (i *interpreter) evalAvg(m model.IUnaryExpression, operand result.Value) (result.Value, error) {
	if result.IsNull(operand) {
		return result.New(nil)
	}
	l, err := result.ToSlice(operand)
	if err != nil {
		return result.Value{}, err
	}
	lType, ok := operand.RuntimeType().(*types.List)
	if !ok {
		return result.Value{}, fmt.Errorf("Avg(%v) operand is not a list", m.GetName())
	}
	switch lType.ElementType {
	case types.Any:
		// Special case for handling lists that contain only null runtime values.
		return result.New(nil)
	case types.Decimal:
		var sum, count float64
		for _, elem := range l {
			if result.IsNull(elem) {
				continue
			}
			v, err := result.ToFloat64(elem)
			if err != nil {
				return result.Value{}, err
			}
			count++
			sum += v
		}
		if count == 0 {
			return result.New(nil)
		}
		return result.New(sum / count)
	case types.Quantity:
		// Keep a running sum of found quantity values and then divide by the count at the end.
		var resultQuantity *result.Quantity
		var count float64
		for _, elem := range l {
			if result.IsNull(elem) {
				continue
			}
			v, err := result.ToQuantity(elem)
			if err != nil {
				return result.Value{}, err
			}
			if resultQuantity == nil {
				resultQuantity = &result.Quantity{Value: 0, Unit: v.Unit}
			}
			if resultQuantity.Unit != v.Unit {
				return result.Value{}, fmt.Errorf("Avg(%v) Quantity operand has different units which is not supported, got %v and %v", m.GetName(), resultQuantity.Unit, v.Unit)
			}
			count++
			resultQuantity.Value += v.Value
		}
		if resultQuantity == nil {
			return result.New(nil)
		}
		resultQuantity.Value /= count
		return result.New(*resultQuantity)
	default:
		return result.Value{}, fmt.Errorf("Avg(%v) operand is not a list of Decimal or Quantity", m.GetName())
	}
}

// Count(argument List<T>) Integer
// https://cql.hl7.org/09-b-cqlreference.html#count
func (i *interpreter) evalCount(m model.IUnaryExpression, operand result.Value) (result.Value, error) {
	if result.IsNull(operand) {
		return result.New(0)
	}
	l, err := result.ToSlice(operand)
	if err != nil {
		return result.Value{}, err
	}
	count := 0
	for _, elem := range l {
		if !result.IsNull(elem) {
			count++
		}
	}
	return result.New(count)
}

// Max(argument List<Date>) Date
// Max(argument List<DateTime>) DateTime
// https://cql.hl7.org/09-b-cqlreference.html#max
func (i *interpreter) evalMaxDateTime(m model.IUnaryExpression, operand result.Value) (result.Value, error) {
	if result.IsNull(operand) {
		return result.New(nil)
	}
	l, err := result.ToSlice(operand)
	if err != nil {
		return result.Value{}, err
	}
	if len(l) == 0 {
		return result.New(nil)
	}
	lType, ok := operand.RuntimeType().(*types.List)
	if !ok {
		return result.Value{}, fmt.Errorf("Max(%v) operand is not a list", m.GetName())
	}
	// Special case for handling lists that contain only null runtime values.
	if lType.ElementType == types.Any {
		return result.New(nil)
	}
	minDtVal, err := minValue(lType.ElementType, &i.evaluationTimestamp)
	if err != nil {
		return result.Value{}, err
	}
	dt, err := result.ToDateTime(minDtVal)
	if err != nil {
		return result.Value{}, err
	}
	for _, elem := range l {
		if result.IsNull(elem) {
			continue
		}
		v, err := result.ToDateTime(elem)
		if err != nil {
			return result.Value{}, err
		}
		compareResult, err := compareDateTime(dt, v)
		if err != nil {
			return result.Value{}, err
		}
		if compareResult == leftBeforeRight {
			dt = v
		}
	}
	if m.GetResultType() == types.Date {
		return result.New(result.Date(dt))
	}
	return result.New(dt)
}

// Min(argument List<Date>) Date
// Min(argument List<DateTime>) DateTime
// https://cql.hl7.org/09-b-cqlreference.html#min
func (i *interpreter) evalMinDateTime(m model.IUnaryExpression, operand result.Value) (result.Value, error) {
	if result.IsNull(operand) {
		return result.New(nil)
	}
	l, err := result.ToSlice(operand)
	if err != nil {
		return result.Value{}, err
	}
	if len(l) == 0 {
		return result.New(nil)
	}
	lType, ok := operand.RuntimeType().(*types.List)
	if !ok {
		return result.Value{}, fmt.Errorf("Min(%v) operand is not a list", m.GetName())
	}
	// Special case for handling lists that contain only null runtime values.
	if lType.ElementType == types.Any {
		return result.New(nil)
	}
	maxDtVal, err := maxValue(lType.ElementType, &i.evaluationTimestamp)
	if err != nil {
		return result.Value{}, err
	}
	dt, err := result.ToDateTime(maxDtVal)
	if err != nil {
		return result.Value{}, err
	}
	for _, elem := range l {
		if result.IsNull(elem) {
			continue
		}
		v, err := result.ToDateTime(elem)
		if err != nil {
			return result.Value{}, err
		}
		compareResult, err := compareDateTime(dt, v)
		if err != nil {
			return result.Value{}, err
		}
		if compareResult == leftAfterRight {
			dt = v
		}
	}
	if m.GetResultType() == types.Date {
		return result.New(result.Date(dt))
	}
	return result.New(dt)
}

// Median(argument List<Decimal>) Decimal
// https://cql.hl7.org/09-b-cqlreference.html#median
func (i *interpreter) evalMedianDecimal(_ model.IUnaryExpression, operand result.Value) (result.Value, error) {
	if result.IsNull(operand) {
		return result.New(nil)
	}

	l, err := result.ToSlice(operand)
	if err != nil {
		return result.Value{}, err
	}

	var values []float64
	for _, elem := range l {
		if result.IsNull(elem) {
			continue
		}
		v, err := result.ToFloat64(elem)
		if err != nil {
			return result.Value{}, err
		}
		values = append(values, v)
	}
	if len(values) == 0 {
		return result.New(nil)
	}

	median := calculateMedianFloat64(values)
	return result.New(median)
}

// Median(argument List<Quantity>) Quantity
// https://cql.hl7.org/09-b-cqlreference.html#median
func (i *interpreter) evalMedianQuantity(_ model.IUnaryExpression, operand result.Value) (result.Value, error) {
	if result.IsNull(operand) {
		return result.New(nil)
	}

	l, err := result.ToSlice(operand)
	if err != nil {
		return result.Value{}, err
	}

	values := make([]float64, 0, len(l))
	var unit model.Unit
	for _, elem := range l {
		if result.IsNull(elem) {
			continue
		}
		v, err := result.ToQuantity(elem)
		if err != nil {
			return result.Value{}, err
		}
		// We only support List<Quantity> where all the elements have the exact same unit, since we do not support
		// mixed unit Quantity math in our engine yet.
		if unit == "" {
			unit = v.Unit
		} else if unit != v.Unit {
			return result.Value{}, fmt.Errorf("Median(List<Quantity>) operand has different units which is not supported, got %v and %v", unit, v.Unit)
		}
		values = append(values, v.Value)
	}
	if len(values) == 0 {
		return result.New(nil)
	}
	median := calculateMedianFloat64(values)
	return result.New(result.Quantity{Value: median, Unit: unit})
}

// calculateMedianFloat64 calculates the median of a slice of float64 values.
// This modifies the values slice in place while sorting it.
func calculateMedianFloat64(values []float64) float64 {
	sort.Float64s(values)
	mid := len(values) / 2
	if len(values)%2 == 0 {
		return (values[mid-1] + values[mid]) / 2
	}
	return values[mid]
}

// Sum(argument List<Decimal>) Decimal
// Sum(argument List<Integer>) Integer
// Sum(argument List<Long>) Long
// Sum(argument List<Quantity>) Quantity
// https://cql.hl7.org/09-b-cqlreference.html#sum
func (i *interpreter) evalSum(m model.IUnaryExpression, operand result.Value) (result.Value, error) {
	if result.IsNull(operand) {
		return result.New(nil)
	}
	l, err := result.ToSlice(operand)
	if err != nil {
		return result.Value{}, err
	}
	lType, ok := operand.RuntimeType().(*types.List)
	if !ok {
		return result.Value{}, fmt.Errorf("Sum(%v) operand is not a list", m.GetName())
	}
	switch lType.ElementType {
	case types.Any:
		// Special case for handling lists that contain only null runtime values.
		return result.New(nil)
	case types.Decimal:
		var sum float64
		var foundValue bool
		for _, elem := range l {
			if result.IsNull(elem) {
				continue
			}
			foundValue = true
			v, err := result.ToFloat64(elem)
			if err != nil {
				return result.Value{}, err
			}
			sum += v
		}
		if !foundValue {
			return result.New(nil)
		}
		return result.New(sum)
	case types.Integer:
		var sum int32
		var foundValue bool
		for _, elem := range l {
			if result.IsNull(elem) {
				continue
			}
			foundValue = true
			v, err := result.ToInt32(elem)
			if err != nil {
				return result.Value{}, err
			}
			sum += v
		}
		if !foundValue {
			return result.New(nil)
		}
		return result.New(sum)
	case types.Long:
		var sum int64
		var foundValue bool
		for _, elem := range l {
			if result.IsNull(elem) {
				continue
			}
			foundValue = true
			v, err := result.ToInt64(elem)
			if err != nil {
				return result.Value{}, err
			}
			sum += v
		}
		if !foundValue {
			return result.New(nil)
		}
		return result.New(sum)
	case types.Quantity:
		var sum result.Quantity
		var foundValue bool
		for _, elem := range l {
			if result.IsNull(elem) {
				continue
			}
			v, err := result.ToQuantity(elem)
			if err != nil {
				return result.Value{}, err
			}
			if !foundValue {
				foundValue = true
				sum = result.Quantity{Value: 0, Unit: v.Unit}
			}
			if sum.Unit != v.Unit {
				return result.Value{}, fmt.Errorf("Sum(%v) got List of Quantity values with different units which is not supported, got %v and %v", m.GetName(), sum.Unit, v.Unit)
			}
			sum.Value += v.Value
		}
		if !foundValue {
			return result.New(nil)
		}
		return result.New(sum)
	default:
		return result.Value{}, fmt.Errorf("Sum(%v) operand is not a list of Decimal, Integer, Long, or Quantity", m.GetName())
	}
}
