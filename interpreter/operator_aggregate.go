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
	"math"
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
func (i *interpreter) evalCount(_ model.IUnaryExpression, operand result.Value) (result.Value, error) {
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

// Max(argument List<Integer>) Integer
// https://cql.hl7.org/09-b-cqlreference.html#max
func (i *interpreter) evalMaxInteger(m model.IUnaryExpression, operand result.Value) (result.Value, error) {
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

	var maxVal int32
	var foundValue bool
	for _, elem := range l {
		if result.IsNull(elem) {
			continue
		}
		v, err := result.ToInt32(elem)
		if err != nil {
			return result.Value{}, err
		}
		if !foundValue || v > maxVal {
			maxVal = v
			foundValue = true
		}
	}
	if !foundValue {
		return result.New(nil)
	}
	return result.New(maxVal)
}

// Max(argument List<Long>) Long
// https://cql.hl7.org/09-b-cqlreference.html#max
func (i *interpreter) evalMaxLong(m model.IUnaryExpression, operand result.Value) (result.Value, error) {
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

	var maxVal int64
	var foundValue bool
	for _, elem := range l {
		if result.IsNull(elem) {
			continue
		}
		v, err := result.ToInt64(elem)
		if err != nil {
			return result.Value{}, err
		}
		if !foundValue || v > maxVal {
			maxVal = v
			foundValue = true
		}
	}
	if !foundValue {
		return result.New(nil)
	}
	return result.New(maxVal)
}

// Max(argument List<Decimal>) Decimal
// https://cql.hl7.org/09-b-cqlreference.html#max
func (i *interpreter) evalMaxDecimal(m model.IUnaryExpression, operand result.Value) (result.Value, error) {
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

	var maxVal float64
	var foundValue bool
	for _, elem := range l {
		if result.IsNull(elem) {
			continue
		}
		v, err := result.ToFloat64(elem)
		if err != nil {
			return result.Value{}, err
		}
		if !foundValue || v > maxVal {
			maxVal = v
			foundValue = true
		}
	}
	if !foundValue {
		return result.New(nil)
	}
	return result.New(maxVal)
}

// Max(argument List<Quantity>) Quantity
// https://cql.hl7.org/09-b-cqlreference.html#max
func (i *interpreter) evalMaxQuantity(m model.IUnaryExpression, operand result.Value) (result.Value, error) {
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

	var maxVal result.Quantity
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
			maxVal = v
			foundValue = true
			continue
		}
		if maxVal.Unit != v.Unit {
			return result.Value{}, fmt.Errorf("Max(%v) got List of Quantity values with different units which is not supported, got %v and %v", m.GetName(), maxVal.Unit, v.Unit)
		}
		if v.Value > maxVal.Value {
			maxVal = v
		}
	}
	if !foundValue {
		return result.New(nil)
	}
	return result.New(maxVal)
}

// Max(argument List<String>) String
// https://cql.hl7.org/09-b-cqlreference.html#max
func (i *interpreter) evalMaxString(m model.IUnaryExpression, operand result.Value) (result.Value, error) {
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

	var maxVal string
	var foundValue bool
	for _, elem := range l {
		if result.IsNull(elem) {
			continue
		}
		v, err := result.ToString(elem)
		if err != nil {
			return result.Value{}, err
		}
		if !foundValue || v > maxVal {
			maxVal = v
			foundValue = true
		}
	}
	if !foundValue {
		return result.New(nil)
	}
	return result.New(maxVal)
}

// Max(argument List<Time>) Time
// https://cql.hl7.org/09-b-cqlreference.html#max
func (i *interpreter) evalMaxTime(m model.IUnaryExpression, operand result.Value) (result.Value, error) {
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

	var maxVal result.Time
	var foundValue bool
	for _, elem := range l {
		if result.IsNull(elem) {
			continue
		}
		v, err := result.ToTime(elem)
		if err != nil {
			return result.Value{}, err
		}
		if !foundValue {
			maxVal = v
			foundValue = true
			continue
		}
		compareResult, err := compareTime(maxVal, v)
		if err != nil {
			return result.Value{}, err
		}
		if compareResult == leftBeforeRight {
			maxVal = v
		}
	}
	if !foundValue {
		return result.New(nil)
	}
	return result.New(maxVal)
}

// Max(argument List<Date>) Date
// https://cql.hl7.org/09-b-cqlreference.html#max
func (i *interpreter) evalMaxDate(m model.IUnaryExpression, operand result.Value) (result.Value, error) {
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
	return result.New(result.Date(dt))
}

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
	return result.New(dt)
}

// Min(argument List<Integer>) Integer
// https://cql.hl7.org/09-b-cqlreference.html#min
func (i *interpreter) evalMinInteger(m model.IUnaryExpression, operand result.Value) (result.Value, error) {
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

	var minVal int32
	var foundValue bool
	for _, elem := range l {
		if result.IsNull(elem) {
			continue
		}
		v, err := result.ToInt32(elem)
		if err != nil {
			return result.Value{}, err
		}
		if !foundValue || v < minVal {
			minVal = v
			foundValue = true
		}
	}
	if !foundValue {
		return result.New(nil)
	}
	return result.New(minVal)
}

// Min(argument List<Long>) Long
// https://cql.hl7.org/09-b-cqlreference.html#min
func (i *interpreter) evalMinLong(m model.IUnaryExpression, operand result.Value) (result.Value, error) {
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

	var minVal int64
	var foundValue bool
	for _, elem := range l {
		if result.IsNull(elem) {
			continue
		}
		v, err := result.ToInt64(elem)
		if err != nil {
			return result.Value{}, err
		}
		if !foundValue || v < minVal {
			minVal = v
			foundValue = true
		}
	}
	if !foundValue {
		return result.New(nil)
	}
	return result.New(minVal)
}

// Min(argument List<Decimal>) Decimal
// https://cql.hl7.org/09-b-cqlreference.html#min
func (i *interpreter) evalMinDecimal(m model.IUnaryExpression, operand result.Value) (result.Value, error) {
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

	var minVal float64
	var foundValue bool
	for _, elem := range l {
		if result.IsNull(elem) {
			continue
		}
		v, err := result.ToFloat64(elem)
		if err != nil {
			return result.Value{}, err
		}
		if !foundValue || v < minVal {
			minVal = v
			foundValue = true
		}
	}
	if !foundValue {
		return result.New(nil)
	}
	return result.New(minVal)
}

// Min(argument List<Quantity>) Quantity
// https://cql.hl7.org/09-b-cqlreference.html#min
func (i *interpreter) evalMinQuantity(m model.IUnaryExpression, operand result.Value) (result.Value, error) {
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

	var minVal result.Quantity
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
			minVal = v
			foundValue = true
			continue
		}
		if minVal.Unit != v.Unit {
			return result.Value{}, fmt.Errorf("Min(%v) got List of Quantity values with different units which is not supported, got %v and %v", m.GetName(), minVal.Unit, v.Unit)
		}
		if v.Value < minVal.Value {
			minVal = v
		}
	}
	if !foundValue {
		return result.New(nil)
	}
	return result.New(minVal)
}

// Min(argument List<String>) String
// https://cql.hl7.org/09-b-cqlreference.html#min
func (i *interpreter) evalMinString(m model.IUnaryExpression, operand result.Value) (result.Value, error) {
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

	var minVal string
	var foundValue bool
	for _, elem := range l {
		if result.IsNull(elem) {
			continue
		}
		v, err := result.ToString(elem)
		if err != nil {
			return result.Value{}, err
		}
		if !foundValue || v < minVal {
			minVal = v
			foundValue = true
		}
	}
	if !foundValue {
		return result.New(nil)
	}
	return result.New(minVal)
}

// Min(argument List<Time>) Time
// https://cql.hl7.org/09-b-cqlreference.html#min
func (i *interpreter) evalMinTime(m model.IUnaryExpression, operand result.Value) (result.Value, error) {
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

	var minVal result.Time
	var foundValue bool
	for _, elem := range l {
		if result.IsNull(elem) {
			continue
		}
		v, err := result.ToTime(elem)
		if err != nil {
			return result.Value{}, err
		}
		if !foundValue {
			minVal = v
			foundValue = true
			continue
		}
		compareResult, err := compareTime(minVal, v)
		if err != nil {
			return result.Value{}, err
		}
		if compareResult == leftAfterRight {
			minVal = v
		}
	}
	if !foundValue {
		return result.New(nil)
	}
	return result.New(minVal)
}

// Min(argument List<Date>) Date
// https://cql.hl7.org/09-b-cqlreference.html#min
func (i *interpreter) evalMinDate(m model.IUnaryExpression, operand result.Value) (result.Value, error) {
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
	return result.New(result.Date(dt))
}

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

	values := make([]float64, 0, len(l))
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
	for idx, elem := range l {
		if result.IsNull(elem) {
			continue
		}
		v, err := result.ToQuantity(elem)
		if err != nil {
			return result.Value{}, err
		}
		// We only support List<Quantity> where all the elements have the exact same unit, since we
		// do not support mixed unit Quantity math in our engine yet.
		if idx == 0 {
			unit = v.Unit
		}
		if unit != v.Unit {
			// TODO: b/342061715 - technically we should treat '' unit and '1' unit as the same, but
			// for now we don't (and we should apply this globally).
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

// compareTime compares two Time values and returns the comparison result.
// Similar to compareDateTime but for Time values.
func compareTime(left, right result.Time) (comparison, error) {
	// Compare hour
	if left.Date.Hour() > right.Date.Hour() {
		return leftAfterRight, nil
	}
	if left.Date.Hour() < right.Date.Hour() {
		return leftBeforeRight, nil
	}

	// Compare minute
	if left.Date.Minute() > right.Date.Minute() {
		return leftAfterRight, nil
	}
	if left.Date.Minute() < right.Date.Minute() {
		return leftBeforeRight, nil
	}

	// Compare second
	if left.Date.Second() > right.Date.Second() {
		return leftAfterRight, nil
	}
	if left.Date.Second() < right.Date.Second() {
		return leftBeforeRight, nil
	}

	// Compare millisecond (nanoseconds / 1000000)
	leftMs := left.Date.Nanosecond() / 1000000
	rightMs := right.Date.Nanosecond() / 1000000
	if leftMs > rightMs {
		return leftAfterRight, nil
	}
	if leftMs < rightMs {
		return leftBeforeRight, nil
	}

	// If all components are equal
	return leftEqualRight, nil
}

// PopulationStdDev(argument List<Decimal>) Decimal
// sqrt(sum((v - mean)^2) / count)
// https://cql.hl7.org/09-b-cqlreference.html#population-stddev
func (i *interpreter) evalPopulationStdDevDecimal(m model.IUnaryExpression, operand result.Value) (result.Value, error) {
	if result.IsNull(operand) {
		return result.New(nil)
	}
	l, err := result.ToSlice(operand)
	if err != nil {
		return result.Value{}, err
	}

	countValue, err := i.evalCount(m, operand)
	if err != nil {
		return result.Value{}, err
	}
	if result.IsNull(countValue) {
		return result.New(nil)
	}
	count, err := result.ToInt32(countValue)
	if err != nil {
		return result.Value{}, err
	}
	if count == 0 {
		return result.New(nil)
	}
	meanValue, err := i.evalAvg(m, operand)
	if err != nil {
		return result.Value{}, err
	}
	if result.IsNull(meanValue) {
		return result.New(nil)
	}
	mean, err := result.ToFloat64(meanValue)
	if err != nil {
		return result.Value{}, err
	}
	var sum float64
	for _, elem := range l {
		if result.IsNull(elem) {
			continue
		}
		v, err := result.ToFloat64(elem)
		if err != nil {
			return result.Value{}, err
		}
		sum += (v - mean) * (v - mean)
	}
	// Round to 8 decimal places to match CQL expected precision
	stdDev := math.Sqrt(sum / float64(count))
	roundedStdDev := math.Round(stdDev*100000000) / 100000000
	return result.New(roundedStdDev)
}

// PopulationStdDev(argument List<Quantity>) Quantity
// sqrt(sum((v - mean)^2) / count)
// https://cql.hl7.org/09-b-cqlreference.html#population-stddev
func (i *interpreter) evalPopulationStdDevQuantity(m model.IUnaryExpression, operand result.Value) (result.Value, error) {
	if result.IsNull(operand) {
		return result.New(nil)
	}
	l, err := result.ToSlice(operand)
	if err != nil {
		return result.Value{}, err
	}

	countValue, err := i.evalCount(m, operand)
	if err != nil {
		return result.Value{}, err
	}
	if result.IsNull(countValue) {
		return result.New(nil)
	}
	count, err := result.ToInt32(countValue)
	if err != nil {
		return result.Value{}, err
	}
	if count == 0 {
		return result.New(nil)
	}
	meanValue, err := i.evalAvg(m, operand)
	if err != nil {
		return result.Value{}, err
	}
	if result.IsNull(meanValue) {
		return result.New(nil)
	}
	mean, err := result.ToQuantity(meanValue)
	if err != nil {
		return result.Value{}, err
	}
	var sum float64
	for _, elem := range l {
		if result.IsNull(elem) {
			continue
		}
		v, err := result.ToQuantity(elem)
		if err != nil {
			return result.Value{}, err
		}
		if v.Unit != mean.Unit {
			return result.Value{}, fmt.Errorf("PopulationStdDev(List<Quantity>) operand has different units which is not supported, got %v and %v", v.Unit, mean.Unit)
		}
		sum += (v.Value - mean.Value) * (v.Value - mean.Value)
	}
	return result.New(result.Quantity{Value: math.Sqrt(sum / float64(count)), Unit: mean.Unit})
}

// StdDev(argument List<Decimal>) Decimal
// sqrt(sum((v - mean)^2) / (count - 1))
// https://cql.hl7.org/09-b-cqlreference.html#stddev
func (i *interpreter) evalStdDevDecimal(m model.IUnaryExpression, operand result.Value) (result.Value, error) {
	if result.IsNull(operand) {
		return result.New(nil)
	}
	l, err := result.ToSlice(operand)
	if err != nil {
		return result.Value{}, err
	}

	countValue, err := i.evalCount(m, operand)
	if err != nil {
		return result.Value{}, err
	}
	if result.IsNull(countValue) {
		return result.New(nil)
	}
	count, err := result.ToInt32(countValue)
	if err != nil {
		return result.Value{}, err
	}
	if count <= 1 {
		return result.New(nil)
	}
	meanValue, err := i.evalAvg(m, operand)
	if err != nil {
		return result.Value{}, err
	}
	if result.IsNull(meanValue) {
		return result.New(nil)
	}
	mean, err := result.ToFloat64(meanValue)
	if err != nil {
		return result.Value{}, err
	}
	var sum float64
	for _, elem := range l {
		if result.IsNull(elem) {
			continue
		}
		v, err := result.ToFloat64(elem)
		if err != nil {
			return result.Value{}, err
		}
		sum += (v - mean) * (v - mean)
	}
	// Round to 8 decimal places to match CQL expected precision
	stdDev := math.Sqrt(sum / float64(count-1))
	roundedStdDev := math.Round(stdDev*100000000) / 100000000
	return result.New(roundedStdDev)
}

// StdDev(argument List<Quantity>) Quantity
// sqrt(sum((v - mean)^2) / (count - 1))
// https://cql.hl7.org/09-b-cqlreference.html#stddev
func (i *interpreter) evalStdDevQuantity(m model.IUnaryExpression, operand result.Value) (result.Value, error) {
	if result.IsNull(operand) {
		return result.New(nil)
	}
	l, err := result.ToSlice(operand)
	if err != nil {
		return result.Value{}, err
	}

	countValue, err := i.evalCount(m, operand)
	if err != nil {
		return result.Value{}, err
	}
	if result.IsNull(countValue) {
		return result.New(nil)
	}
	count, err := result.ToInt32(countValue)
	if err != nil {
		return result.Value{}, err
	}
	if count <= 1 {
		return result.New(nil)
	}
	meanValue, err := i.evalAvg(m, operand)
	if err != nil {
		return result.Value{}, err
	}
	if result.IsNull(meanValue) {
		return result.New(nil)
	}
	mean, err := result.ToQuantity(meanValue)
	if err != nil {
		return result.Value{}, err
	}
	var sum float64
	for _, elem := range l {
		if result.IsNull(elem) {
			continue
		}
		v, err := result.ToQuantity(elem)
		if err != nil {
			return result.Value{}, err
		}
		if v.Unit != mean.Unit {
			return result.Value{}, fmt.Errorf("StdDev(List<Quantity>) operand has different units which is not supported, got %v and %v", v.Unit, mean.Unit)
		}
		sum += (v.Value - mean.Value) * (v.Value - mean.Value)
	}
	
	// Round to 8 decimal places to match CQL expected precision
	stdDev := math.Sqrt(sum / float64(count-1))
	roundedStdDev := math.Round(stdDev*100000000) / 100000000
	return result.New(result.Quantity{Value: roundedStdDev, Unit: mean.Unit})
}

// Variance(argument List<Decimal>) Decimal
// sum((v - mean)^2) / (count - 1)
// https://cql.hl7.org/09-b-cqlreference.html#variance
func (i *interpreter) evalVarianceDecimal(m model.IUnaryExpression, operand result.Value) (result.Value, error) {
	if result.IsNull(operand) {
		return result.New(nil)
	}
	l, err := result.ToSlice(operand)
	if err != nil {
		return result.Value{}, err
	}

	countValue, err := i.evalCount(m, operand)
	if err != nil {
		return result.Value{}, err
	}
	if result.IsNull(countValue) {
		return result.New(nil)
	}
	count, err := result.ToInt32(countValue)
	if err != nil {
		return result.Value{}, err
	}
	if count <= 1 {
		return result.New(nil)
	}
	meanValue, err := i.evalAvg(m, operand)
	if err != nil {
		return result.Value{}, err
	}
	if result.IsNull(meanValue) {
		return result.New(nil)
	}
	mean, err := result.ToFloat64(meanValue)
	if err != nil {
		return result.Value{}, err
	}
	var sum float64
	for _, elem := range l {
		if result.IsNull(elem) {
			continue
		}
		v, err := result.ToFloat64(elem)
		if err != nil {
			return result.Value{}, err
		}
		sum += (v - mean) * (v - mean)
	}
	return result.New(sum / float64(count-1))
}

// Variance(argument List<Quantity>) Quantity
// sum((v - mean)^2) / (count - 1)
// https://cql.hl7.org/09-b-cqlreference.html#variance
func (i *interpreter) evalVarianceQuantity(m model.IUnaryExpression, operand result.Value) (result.Value, error) {
	if result.IsNull(operand) {
		return result.New(nil)
	}
	l, err := result.ToSlice(operand)
	if err != nil {
		return result.Value{}, err
	}

	countValue, err := i.evalCount(m, operand)
	if err != nil {
		return result.Value{}, err
	}
	if result.IsNull(countValue) {
		return result.New(nil)
	}
	count, err := result.ToInt32(countValue)
	if err != nil {
		return result.Value{}, err
	}
	if count <= 1 {
		return result.New(nil)
	}
	meanValue, err := i.evalAvg(m, operand)
	if err != nil {
		return result.Value{}, err
	}
	if result.IsNull(meanValue) {
		return result.New(nil)
	}
	mean, err := result.ToQuantity(meanValue)
	if err != nil {
		return result.Value{}, err
	}
	var sum float64
	for _, elem := range l {
		if result.IsNull(elem) {
			continue
		}
		v, err := result.ToQuantity(elem)
		if err != nil {
			return result.Value{}, err
		}
		if v.Unit != mean.Unit {
			return result.Value{}, fmt.Errorf("Variance(List<Quantity>) operand has different units which is not supported, got %v and %v", v.Unit, mean.Unit)
		}
		sum += (v.Value - mean.Value) * (v.Value - mean.Value)
	}
	return result.New(result.Quantity{Value: sum / float64(count-1), Unit: mean.Unit})
}

// Mode(argument List<Decimal>) Decimal
// https://cql.hl7.org/09-b-cqlreference.html#mode
func (i *interpreter) evalModeDecimal(_ model.IUnaryExpression, operand result.Value) (result.Value, error) {
	if result.IsNull(operand) {
		return result.New(nil)
	}
	l, err := result.ToSlice(operand)
	if err != nil {
		return result.Value{}, err
	}

	// Count occurrences of each value
	counts := make(map[float64]int)
	for _, elem := range l {
		if result.IsNull(elem) {
			continue
		}
		v, err := result.ToFloat64(elem)
		if err != nil {
			return result.Value{}, err
		}
		counts[v]++
	}

	if len(counts) == 0 {
		return result.New(nil)
	}

	// Find the most frequent value
	var mode float64
	maxCount := 0
	for value, count := range counts {
		if count > maxCount {
			maxCount = count
			mode = value
		}
	}

	return result.New(mode)
}

// Mode(argument List<Quantity>) Quantity
// https://cql.hl7.org/09-b-cqlreference.html#mode
func (i *interpreter) evalModeQuantity(_ model.IUnaryExpression, operand result.Value) (result.Value, error) {
	if result.IsNull(operand) {
		return result.New(nil)
	}
	l, err := result.ToSlice(operand)
	if err != nil {
		return result.Value{}, err
	}

	// Count occurrences of each value
	type quantityKey struct {
		value float64
		unit  model.Unit
	}
	counts := make(map[quantityKey]int)
	var unit model.Unit
	var hasUnit bool

	for _, elem := range l {
		if result.IsNull(elem) {
			continue
		}
		v, err := result.ToQuantity(elem)
		if err != nil {
			return result.Value{}, err
		}

		if !hasUnit {
			unit = v.Unit
			hasUnit = true
		} else if v.Unit != unit {
			return result.Value{}, fmt.Errorf("Mode(List<Quantity>) operand has different units which is not supported, got %v and %v", v.Unit, unit)
		}

		key := quantityKey{value: v.Value, unit: v.Unit}
		counts[key]++
	}

	if len(counts) == 0 {
		return result.New(nil)
	}

	// Find the most frequent value
	var modeKey quantityKey
	maxCount := 0
	for key, count := range counts {
		if count > maxCount {
			maxCount = count
			modeKey = key
		}
	}

	return result.New(result.Quantity{Value: modeKey.value, Unit: modeKey.unit})
}

// Mode(argument List<String>) String
// https://cql.hl7.org/09-b-cqlreference.html#mode
func (i *interpreter) evalModeString(_ model.IUnaryExpression, operand result.Value) (result.Value, error) {
	if result.IsNull(operand) {
		return result.New(nil)
	}
	l, err := result.ToSlice(operand)
	if err != nil {
		return result.Value{}, err
	}

	// Count occurrences of each value
	counts := make(map[string]int)
	for _, elem := range l {
		if result.IsNull(elem) {
			continue
		}
		v, err := result.ToString(elem)
		if err != nil {
			return result.Value{}, err
		}
		counts[v]++
	}

	if len(counts) == 0 {
		return result.New(nil)
	}

	// Find the most frequent value
	var mode string
	maxCount := 0
	for value, count := range counts {
		if count > maxCount {
			maxCount = count
			mode = value
		}
	}

	return result.New(mode)
}

// Mode(argument List<Integer>) Integer
// https://cql.hl7.org/09-b-cqlreference.html#mode
func (i *interpreter) evalModeInteger(_ model.IUnaryExpression, operand result.Value) (result.Value, error) {
	if result.IsNull(operand) {
		return result.New(nil)
	}
	l, err := result.ToSlice(operand)
	if err != nil {
		return result.Value{}, err
	}

	// Count occurrences of each value
	counts := make(map[int32]int)
	for _, elem := range l {
		if result.IsNull(elem) {
			continue
		}
		v, err := result.ToInt32(elem)
		if err != nil {
			return result.Value{}, err
		}
		counts[v]++
	}

	if len(counts) == 0 {
		return result.New(nil)
	}

	// Find the most frequent value
	var mode int32
	maxCount := 0
	for value, count := range counts {
		if count > maxCount {
			maxCount = count
			mode = value
		}
	}

	return result.New(mode)
}

// Mode(argument List<Long>) Long
// https://cql.hl7.org/09-b-cqlreference.html#mode
func (i *interpreter) evalModeLong(_ model.IUnaryExpression, operand result.Value) (result.Value, error) {
	if result.IsNull(operand) {
		return result.New(nil)
	}
	l, err := result.ToSlice(operand)
	if err != nil {
		return result.Value{}, err
	}

	// Count occurrences of each value
	counts := make(map[int64]int)
	for _, elem := range l {
		if result.IsNull(elem) {
			continue
		}
		v, err := result.ToInt64(elem)
		if err != nil {
			return result.Value{}, err
		}
		counts[v]++
	}

	if len(counts) == 0 {
		return result.New(nil)
	}

	// Find the most frequent value
	var mode int64
	maxCount := 0
	for value, count := range counts {
		if count > maxCount {
			maxCount = count
			mode = value
		}
	}

	return result.New(mode)
}

// Mode(argument List<Date>) Date
// https://cql.hl7.org/09-b-cqlreference.html#mode
func (i *interpreter) evalModeDate(_ model.IUnaryExpression, operand result.Value) (result.Value, error) {
	if result.IsNull(operand) {
		return result.New(nil)
	}
	l, err := result.ToSlice(operand)
	if err != nil {
		return result.Value{}, err
	}

	// Count occurrences of each value
	counts := make(map[string]int)
	dateMap := make(map[string]result.Date)

	for _, elem := range l {
		if result.IsNull(elem) {
			continue
		}
		v, err := result.ToDate(elem)
		if err != nil {
			return result.Value{}, err
		}

		// Format date as string to use as map key
		key := fmt.Sprintf("%d-%02d-%02d", v.Date.Year(), v.Date.Month(), v.Date.Day())
		counts[key]++
		dateMap[key] = v
	}

	if len(counts) == 0 {
		return result.New(nil)
	}

	// Find the most frequent value
	var modeKey string
	maxCount := 0
	for key, count := range counts {
		if count > maxCount {
			maxCount = count
			modeKey = key
		}
	}

	return result.New(dateMap[modeKey])
}

// Mode(argument List<DateTime>) DateTime
// https://cql.hl7.org/09-b-cqlreference.html#mode
func (i *interpreter) evalModeDateTime(_ model.IUnaryExpression, operand result.Value) (result.Value, error) {
	if result.IsNull(operand) {
		return result.New(nil)
	}
	l, err := result.ToSlice(operand)
	if err != nil {
		return result.Value{}, err
	}

	// Count occurrences of each value
	counts := make(map[string]int)
	dtMap := make(map[string]result.DateTime)

	for _, elem := range l {
		if result.IsNull(elem) {
			continue
		}
		v, err := result.ToDateTime(elem)
		if err != nil {
			return result.Value{}, err
		}

		// Format datetime as string to use as map key
		key := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d.%03d",
			v.Date.Year(), v.Date.Month(), v.Date.Day(),
			v.Date.Hour(), v.Date.Minute(), v.Date.Second(),
			v.Date.Nanosecond()/1000000)
		counts[key]++
		dtMap[key] = v
	}

	if len(counts) == 0 {
		return result.New(nil)
	}

	// Find the most frequent value
	var modeKey string
	maxCount := 0
	for key, count := range counts {
		if count > maxCount {
			maxCount = count
			modeKey = key
		}
	}

	return result.New(dtMap[modeKey])
}

// Mode(argument List<Time>) Time
// https://cql.hl7.org/09-b-cqlreference.html#mode
func (i *interpreter) evalModeTime(_ model.IUnaryExpression, operand result.Value) (result.Value, error) {
	if result.IsNull(operand) {
		return result.New(nil)
	}
	l, err := result.ToSlice(operand)
	if err != nil {
		return result.Value{}, err
	}

	// Count occurrences of each value
	counts := make(map[string]int)
	timeMap := make(map[string]result.Time)

	for _, elem := range l {
		if result.IsNull(elem) {
			continue
		}
		v, err := result.ToTime(elem)
		if err != nil {
			return result.Value{}, err
		}

		// Format time as string to use as map key
		key := fmt.Sprintf("%02d:%02d:%02d.%03d",
			v.Date.Hour(), v.Date.Minute(), v.Date.Second(),
			v.Date.Nanosecond()/1000000)
		counts[key]++
		timeMap[key] = v
	}

	if len(counts) == 0 {
		return result.New(nil)
	}

	// Find the most frequent value
	var modeKey string
	maxCount := 0
	for key, count := range counts {
		if count > maxCount {
			maxCount = count
			modeKey = key
		}
	}

	return result.New(timeMap[modeKey])
}

// PopulationVariance(argument List<Decimal>) Decimal
// sum((v - mean)^2) / count
// https://cql.hl7.org/09-b-cqlreference.html#population-variance
func (i *interpreter) evalPopulationVarianceDecimal(m model.IUnaryExpression, operand result.Value) (result.Value, error) {
	if result.IsNull(operand) {
		return result.New(nil)
	}
	l, err := result.ToSlice(operand)
	if err != nil {
		return result.Value{}, err
	}

	countValue, err := i.evalCount(m, operand)
	if err != nil {
		return result.Value{}, err
	}
	if result.IsNull(countValue) {
		return result.New(nil)
	}
	count, err := result.ToInt32(countValue)
	if err != nil {
		return result.Value{}, err
	}
	if count == 0 {
		return result.New(nil)
	}
	meanValue, err := i.evalAvg(m, operand)
	if err != nil {
		return result.Value{}, err
	}
	if result.IsNull(meanValue) {
		return result.New(nil)
	}
	mean, err := result.ToFloat64(meanValue)
	if err != nil {
		return result.Value{}, err
	}
	var sum float64
	for _, elem := range l {
		if result.IsNull(elem) {
			continue
		}
		v, err := result.ToFloat64(elem)
		if err != nil {
			return result.Value{}, err
		}
		sum += (v - mean) * (v - mean)
	}
	return result.New(sum / float64(count))
}

// PopulationVariance(argument List<Quantity>) Quantity
// sum((v - mean)^2) / count
// https://cql.hl7.org/09-b-cqlreference.html#population-variance
func (i *interpreter) evalPopulationVarianceQuantity(m model.IUnaryExpression, operand result.Value) (result.Value, error) {
	if result.IsNull(operand) {
		return result.New(nil)
	}
	l, err := result.ToSlice(operand)
	if err != nil {
		return result.Value{}, err
	}

	countValue, err := i.evalCount(m, operand)
	if err != nil {
		return result.Value{}, err
	}
	if result.IsNull(countValue) {
		return result.New(nil)
	}
	count, err := result.ToInt32(countValue)
	if err != nil {
		return result.Value{}, err
	}
	if count == 0 {
		return result.New(nil)
	}
	meanValue, err := i.evalAvg(m, operand)
	if err != nil {
		return result.Value{}, err
	}
	if result.IsNull(meanValue) {
		return result.New(nil)
	}
	mean, err := result.ToQuantity(meanValue)
	if err != nil {
		return result.Value{}, err
	}
	var sum float64
	for _, elem := range l {
		if result.IsNull(elem) {
			continue
		}
		v, err := result.ToQuantity(elem)
		if err != nil {
			return result.Value{}, err
		}
		if v.Unit != mean.Unit {
			return result.Value{}, fmt.Errorf("PopulationVariance(List<Quantity>) operand has different units which is not supported, got %v and %v", v.Unit, mean.Unit)
		}
		sum += (v.Value - mean.Value) * (v.Value - mean.Value)
	}
	return result.New(result.Quantity{Value: sum / float64(count), Unit: mean.Unit})
}

// GeometricMean(argument List<Decimal>) Decimal
// Power(Product(X), 1 / Count(X))
// https://cql.hl7.org/09-b-cqlreference.html#geometricmean
func (i *interpreter) evalGeometricMeanDecimal(m model.IUnaryExpression, operand result.Value) (result.Value, error) {
	if result.IsNull(operand) {
		return result.New(nil)
	}
	l, err := result.ToSlice(operand)
	if err != nil {
		return result.Value{}, err
	}

	// Collect non-null values for product calculation
	var values []float64
	for _, elem := range l {
		if result.IsNull(elem) {
			continue
		}
		v, err := result.ToFloat64(elem)
		if err != nil {
			return result.Value{}, err
		}
		if v <= 0 {
			return result.Value{}, fmt.Errorf("GeometricMean(%v) operand contains non-positive value %v which is not supported", m.GetName(), v)
		}
		values = append(values, v)
	}

	if len(values) == 0 {
		return result.New(nil)
	}

	// Calculate product
	product := 1.0
	for _, v := range values {
		product *= v
	}

	// Calculate nth root (Power(product, 1/count))
	power := 1.0 / float64(len(values))
	geometricMean := math.Pow(product, power)
	
	// Round to 8 decimal places to match CQL expected precision
	roundedMean := math.Round(geometricMean*100000000) / 100000000
	return result.New(roundedMean)
}

// GeometricMean(argument List<Quantity>) Quantity
// Power(Product(X), 1 / Count(X))
// https://cql.hl7.org/09-b-cqlreference.html#geometricmean
func (i *interpreter) evalGeometricMeanQuantity(m model.IUnaryExpression, operand result.Value) (result.Value, error) {
	if result.IsNull(operand) {
		return result.New(nil)
	}
	l, err := result.ToSlice(operand)
	if err != nil {
		return result.Value{}, err
	}

	// Track unit and values
	var values []float64
	var unit model.Unit

	for idx, elem := range l {
		if result.IsNull(elem) {
			continue
		}
		v, err := result.ToQuantity(elem)
		if err != nil {
			return result.Value{}, err
		}

		// Set initial unit or validate consistency
		if idx == 0 || len(values) == 0 {
			unit = v.Unit
		} else if v.Unit != unit {
			return result.Value{}, fmt.Errorf("GeometricMean(%v) Quantity operand has different units which is not supported, got %v and %v", m.GetName(), unit, v.Unit)
		}

		if v.Value <= 0 {
			return result.Value{}, fmt.Errorf("GeometricMean(%v) operand contains non-positive value %v which is not supported", m.GetName(), v.Value)
		}

		values = append(values, v.Value)
	}

	if len(values) == 0 {
		return result.New(nil)
	}

	// Calculate product
	product := 1.0
	for _, v := range values {
		product *= v
	}

	// Calculate nth root (Power(product, 1/count))
	power := 1.0 / float64(len(values))
	geometricMean := math.Pow(product, power)
	
	// Round to 8 decimal places to match CQL expected precision
	roundedMean := math.Round(geometricMean*100000000) / 100000000
	return result.New(result.Quantity{Value: roundedMean, Unit: unit})
}

// Product(argument List<Integer>) Integer
// Product(argument List<Long>) Long
// Product(argument List<Decimal>) Decimal
// Product(argument List<Quantity>) Quantity
// https://cql.hl7.org/09-b-cqlreference.html#product
func (i *interpreter) evalProduct(m model.IUnaryExpression, operand result.Value) (result.Value, error) {
	if result.IsNull(operand) {
		return result.New(nil)
	}
	l, err := result.ToSlice(operand)
	if err != nil {
		return result.Value{}, err
	}

	lType, ok := operand.RuntimeType().(*types.List)
	if !ok {
		return result.Value{}, fmt.Errorf("Product(%v) operand is not a list", m.GetName())
	}
	switch lType.ElementType {
	case types.Any:
		// Special case for handling lists that contain only null runtime values.
		return result.New(nil)
	case types.Decimal:
		var product float64 = 1.0
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
			product *= v
		}
		if !foundValue {
			return result.New(nil)
		}
		return result.New(product)
	case types.Integer:
		var product int32 = 1
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
			product *= v
		}
		if !foundValue {
			return result.New(nil)
		}
		return result.New(product)
	case types.Long:
		var product int64 = 1
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
			product *= v
		}
		if !foundValue {
			return result.New(nil)
		}
		return result.New(product)
	case types.Quantity:
		var product result.Quantity
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
				product = result.Quantity{Value: 1, Unit: v.Unit}
			}
			if product.Unit != v.Unit {
				return result.Value{}, fmt.Errorf("Product(%v) got List of Quantity values with different units which is not supported, got %v and %v", m.GetName(), product.Unit, v.Unit)
			}
			product.Value *= v.Value
		}
		if !foundValue {
			return result.New(nil)
		}
		return result.New(product)
	default:
		return result.Value{}, fmt.Errorf("Product(%v) operand is not a list of Integer, Long, Decimal, or Quantity", m.GetName())
	}
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
