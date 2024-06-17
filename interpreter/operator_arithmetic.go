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
	"math"
	"reflect"
	"time"

	"github.com/google/cql/model"
	"github.com/google/cql/result"
	"github.com/google/cql/types"
)

// ARITHMETIC OPERATORS - https://cql.hl7.org/09-b-cqlreference.html#arithmetic-operators-4

const (
	// Max/Min values for CQL decimals.
	// In the future we may switch to using math.MaxFloat64.
	maxDecimal = float64(99999999999999999999.99999999)
	minDecimal = float64(-99999999999999999999.99999999)
)

// op(left Integer, right Integer) Integer
// https://cql.hl7.org/09-b-cqlreference.html#add
// https://cql.hl7.org/09-b-cqlreference.html#subtract
// https://cql.hl7.org/09-b-cqlreference.html#multiply
// https://cql.hl7.org/09-b-cqlreference.html#truncated-divide
// https://cql.hl7.org/09-b-cqlreference.html#modulo
func evalArithmeticInteger(m model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) || result.IsNull(rObj) {
		return result.New(nil)
	}
	l, r, err := applyToValues(lObj, rObj, result.ToInt32)
	if err != nil {
		return result.Value{}, err
	}
	return arithmetic(m, l, r)
}

// op(left Long, right Long) Long
// https://cql.hl7.org/09-b-cqlreference.html#add
// https://cql.hl7.org/09-b-cqlreference.html#subtract
// https://cql.hl7.org/09-b-cqlreference.html#multiply
// https://cql.hl7.org/09-b-cqlreference.html#truncated-divide
// https://cql.hl7.org/09-b-cqlreference.html#modulo
func evalArithmeticLong(m model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) || result.IsNull(rObj) {
		return result.New(nil)
	}
	l, r, err := applyToValues(lObj, rObj, result.ToInt64)
	if err != nil {
		return result.Value{}, err
	}
	return arithmetic(m, l, r)
}

// op(left Decimal, right Decimal) Decimal
// https://cql.hl7.org/09-b-cqlreference.html#add
// https://cql.hl7.org/09-b-cqlreference.html#subtract
// https://cql.hl7.org/09-b-cqlreference.html#multiply
// https://cql.hl7.org/09-b-cqlreference.html#truncated-divide
// https://cql.hl7.org/09-b-cqlreference.html#modulo
func evalArithmeticDecimal(m model.IBinaryExpression, lObj result.Value, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) || result.IsNull(rObj) {
		return result.New(nil)
	}
	l, r, err := applyToValues(lObj, rObj, result.ToFloat64)
	if err != nil {
		return result.Value{}, err
	}
	return arithmetic(m, l, r)
}

// op(left Quantity, right Quantity) Quantity
// https://cql.hl7.org/09-b-cqlreference.html#add
// https://cql.hl7.org/09-b-cqlreference.html#subtract
// While the docs for these functions are ambiguous on this topic, performing
// arithmetic on Quantities with a unit of Day or less to a Date/DateTime with
// month or year precision is undefined and should return null. This is because
// of the variability of what the definition of a month is. If a user needs this
// functionality they should use UCUM duration values.
// See: https://cql.hl7.org/09-b-cqlreference.html#equal
func evalArithmeticQuantity(m model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) || result.IsNull(rObj) {
		return result.New(nil)
	}
	l, r, err := applyToValues(lObj, rObj, result.ToQuantity)
	if err != nil {
		return result.Value{}, err
	}
	return arithmeticQuantity(m, l, r)
}

// op(left Date, right Quantity) Date
// https://cql.hl7.org/09-b-cqlreference.html#add-1
// https://cql.hl7.org/09-b-cqlreference.html#subtract-1
func evalArithmeticDate(m model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) || result.IsNull(rObj) {
		return result.New(nil)
	}
	r, err := result.ToQuantity(rObj)
	if err != nil {
		return result.Value{}, err
	}

	d, err := result.ToDateTime(lObj)
	if err != nil {
		return result.Value{}, err
	}
	allowUnsetPrec := false // Dates must have a precision
	err = validateDatePrecision(d.Precision, allowUnsetPrec)
	if err != nil {
		return result.Value{}, err
	}
	dtv, err := arithmeticDateTime(m, d, r)
	if err != nil {
		return result.Value{}, err
	}
	return result.New(result.Date{Date: dtv.Date, Precision: dtv.Precision})
}

// op(left DateTime, right Quantity) DateTime
// https://cql.hl7.org/09-b-cqlreference.html#add-1
// https://cql.hl7.org/09-b-cqlreference.html#subtract-1
func evalArithmeticDateTime(m model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) || result.IsNull(rObj) {
		return result.New(nil)
	}
	r, err := result.ToQuantity(rObj)
	if err != nil {
		return result.Value{}, err
	}

	d, err := result.ToDateTime(lObj)
	if err != nil {
		return result.Value{}, err
	}
	allowUnsetPrec := false // DateTimes must have a precision.
	err = validateDateTimePrecision(d.Precision, allowUnsetPrec)
	if err != nil {
		return result.Value{}, err
	}
	dtv, err := arithmeticDateTime(m, d, r)
	if err != nil {
		return result.Value{}, err
	}
	return result.New(dtv)
}

func arithmetic[t float64 | int64 | int32](m model.IBinaryExpression, l, r t) (result.Value, error) {
	switch m.(type) {
	case *model.Add:
		return result.New(l + r)
	case *model.Subtract:
		return result.New(l - r)
	case *model.Multiply:
		return result.New(l * r)
	case *model.TruncatedDivide:
		if r == 0 {
			return result.New(nil)
		}
		// The first int64() truncates any decimal, then t() converts it back to the original type.
		return result.New(t(int64(l / r)))
	case *model.Divide:
		if r == 0 {
			return result.New(nil)
		}
		return result.New(l / r)
	case *model.Modulo:
		return mod(l, r)
	}
	return result.Value{}, fmt.Errorf("internal error - unsupported Binary Arithmetic Expression %v", m)
}

// TODO(b/319156186): Add support for converting quantities between different units.
// TODO(b/319333058): Add support for Date + Quantity arithmetic.
// TODO(b/319525986): Add support for additional arithmetic for Quantities.
func arithmeticQuantity(m model.IBinaryExpression, l, r result.Quantity) (result.Value, error) {
	if l.Unit != r.Unit {
		return result.Value{}, fmt.Errorf("internal error - quantity unit conversion unsupported, got units: %s and %s", l.Unit, r.Unit)
	}
	switch m.(type) {
	case *model.Add:
		return result.New(result.Quantity{Value: l.Value + r.Value, Unit: l.Unit})
	case *model.Subtract:
		return result.New(result.Quantity{Value: l.Value - r.Value, Unit: l.Unit})
	case *model.Multiply:
		return result.Value{}, fmt.Errorf("internal error - quantity multiplication unsupported, got: %v and %v", l, r)
	case *model.TruncatedDivide:
		return result.New(result.Quantity{Value: float64(int64(l.Value / r.Value)), Unit: model.ONEUNIT})
	case *model.Divide:
		return result.New(result.Quantity{Value: l.Value / r.Value, Unit: model.ONEUNIT})
	case *model.Modulo:
		if l.Unit != r.Unit {
		return result.Value{}, fmt.Errorf("internal error - quantity modulo with different units unsupported, got units: %s and %s", l.Unit, r.Unit)
		}
		if r.Value == 0 {
			return result.New(nil)
		}
		return result.New(result.Quantity{Value: math.Mod(l.Value, r.Value), Unit: l.Unit})
	}
	return result.Value{}, fmt.Errorf("internal error - unsupported Binary Arithmetic Expression %v", m)
}

// arithmeticDateTime performs arithmetic operations for Date, Quantity values.
// When performing arithmetic over differing precisions, only whole values up to
// the given Date or DateTime's precision should be added.
func arithmeticDateTime(m model.IBinaryExpression, l result.DateTime, r result.Quantity) (result.DateTime, error) {
	var sign int64
	switch m.(type) {
	case *model.Add:
		sign = 1
	case *model.Subtract:
		sign = -1
	default:
		return result.DateTime{}, fmt.Errorf("internal error - unsupported Binary Arithmetic Expression %v", m)
	}
	cq, err := convertQuantityUpToPrecision(r, l.Precision)
	if err != nil {
		return result.DateTime{}, err
	}

	switch cq.Unit {
	case model.YEARUNIT:
		return result.DateTime{Date: l.Date.AddDate(int(sign)*int(cq.Value), 0, 0), Precision: l.Precision}, nil
	case model.MONTHUNIT:
		return result.DateTime{Date: l.Date.AddDate(0, int(sign)*int(cq.Value), 0), Precision: l.Precision}, nil
	case model.WEEKUNIT:
		// Weeks need to be converted to days before they can be operated on.
		return result.DateTime{Date: l.Date.AddDate(0, 0, int(sign)*int(cq.Value*7)), Precision: l.Precision}, nil
	case model.DAYUNIT:
		return result.DateTime{Date: l.Date.AddDate(0, 0, int(sign)*int(cq.Value)), Precision: l.Precision}, nil
	case model.HOURUNIT:
		d := time.Hour * time.Duration(sign*int64(cq.Value))
		return result.DateTime{Date: l.Date.Add(d), Precision: l.Precision}, nil
	case model.MINUTEUNIT:
		d := time.Minute * time.Duration(sign*int64(cq.Value))
		return result.DateTime{Date: l.Date.Add(d), Precision: l.Precision}, nil
	// Seconds and Milliseconds shouldn't be truncated so we have to convert them to
	// nanoseconds manually.
	case model.SECONDUNIT:
		d := time.Duration(sign * int64(cq.Value*1000*1000*1000))
		return result.DateTime{Date: l.Date.Add(d), Precision: l.Precision}, nil
	case model.MILLISECONDUNIT:
		d := time.Duration(sign * int64(cq.Value*1000*1000))
		return result.DateTime{Date: l.Date.Add(d), Precision: l.Precision}, nil
	}
	return result.DateTime{}, fmt.Errorf("internal error - unsupported quantity unit %v in arithmetic operation", cq.Unit)
}

// Truncate(arg Decimal) Integer
// https://cql.hl7.org/09-b-cqlreference.html#truncate
func evalTruncate(_ model.IUnaryExpression, decimalVal result.Value) (result.Value, error) {
	if result.IsNull(decimalVal) {
		return result.New(nil)
	}
	d, err := result.ToFloat64(decimalVal)
	if err != nil {
		return result.Value{}, err
	}
	return result.New(int32(d))
}

// -(argument Integer) Integer
// https://cql.hl7.org/09-b-cqlreference.html#negate
func evalNegateInteger(m model.IUnaryExpression, obj result.Value) (result.Value, error) {
	if result.IsNull(obj) {
		return result.New(nil)
	}
	val, err := result.ToInt32(obj)
	if err != nil {
		return result.Value{}, err
	}
	min, err := minValue(types.Integer, nil)
	if err != nil {
		return result.Value{}, err
	}
	if obj.Equal(min) {
		return result.New(nil)
	}
	return result.New(-val)
}

// -(argument Long) Long
// https://cql.hl7.org/09-b-cqlreference.html#negate
func evalNegateLong(m model.IUnaryExpression, obj result.Value) (result.Value, error) {
	if result.IsNull(obj) {
		return result.New(nil)
	}
	val, err := result.ToInt64(obj)
	if err != nil {
		return result.Value{}, err
	}
	min, err := minValue(types.Long, nil)
	if err != nil {
		return result.Value{}, err
	}
	if obj.Equal(min) {
		return result.New(nil)
	}
	return result.New(-val)
}

// -(argument Decimal) Decimal
// https://cql.hl7.org/09-b-cqlreference.html#negate
func evalNegateDecimal(m model.IUnaryExpression, obj result.Value) (result.Value, error) {
	if result.IsNull(obj) {
		return result.New(nil)
	}
	val, err := result.ToFloat64(obj)
	if err != nil {
		return result.Value{}, err
	}
	min, err := minValue(types.Decimal, nil)
	if err != nil {
		return result.Value{}, err
	}
	if obj.Equal(min) {
		return result.New(nil)
	}
	return result.New(-val)
}

// -(argument Quantity) Quantity
// https://cql.hl7.org/09-b-cqlreference.html#negate
func evalNegateQuantity(m model.IUnaryExpression, obj result.Value) (result.Value, error) {
	if result.IsNull(obj) {
		return result.New(nil)
	}
	val, err := result.ToQuantity(obj)
	if err != nil {
		return result.Value{}, err
	}
	min, err := minValue(types.Quantity, nil)
	if err != nil {
		return result.Value{}, err
	}
	if obj.Equal(min) {
		return result.New(nil)
	}
	val.Value = -val.Value
	return result.New(val)
}

// predecessor of<T>(obj T) T
// https://cql.hl7.org/09-b-cqlreference.html#predecessor
func (i *interpreter) evalPredecessor(m model.IUnaryExpression, obj result.Value) (result.Value, error) {
	return predecessor(obj, &i.evaluationTimestamp)
}

// successor of<T>(obj T) T
// https://cql.hl7.org/09-b-cqlreference.html#successor
func (i *interpreter) evalSuccessor(m model.IUnaryExpression, obj result.Value) (result.Value, error) {
	return successor(obj, &i.evaluationTimestamp)
}

func predecessor(obj result.Value, evaluationTimestamp *time.Time) (result.Value, error) {
	if result.IsNull(obj) {
		return result.New(nil)
	}
	minVal, err := minValue(obj.RuntimeType(), evaluationTimestamp)
	if err != nil {
		return result.Value{}, err
	}
	if obj.Equal(minVal) {
		return result.Value{}, fmt.Errorf("tried to compute predecessor for value that is already a min value, %v", obj.GolangValue())
	}

	switch t := obj.RuntimeType(); t {
	case types.Integer:
		i, err := result.ToInt32(obj)
		if err != nil {
			return result.Value{}, err
		}
		return result.New(i - 1)
	case types.Long:
		l, err := result.ToInt64(obj)
		if err != nil {
			return result.Value{}, err
		}
		return result.New(l - 1)
	case types.Decimal:
		d, err := result.ToFloat64(obj)
		if err != nil {
			return result.Value{}, err
		}
		return result.New(d - 0.00000001)
	case types.Quantity:
		// TODO: b/329707836 -  Determine under what cases quantities should be incremented by whole numbers in stead of decimals.
		q, err := result.ToQuantity(obj)
		if err != nil {
			return result.Value{}, err
		}
		return result.New(result.Quantity{Value: q.Value - 0.00000001, Unit: q.Unit})
	case types.Date, types.DateTime, types.Time:
		return dateTimePredecessor(obj, evaluationTimestamp)
	default:
		return result.Value{}, fmt.Errorf("internal error - unsupported type %v", t)
	}
}

func successor(obj result.Value, evaluationTimestamp *time.Time) (result.Value, error) {
	if result.IsNull(obj) {
		return result.New(nil)
	}
	maxVal, err := maxValue(obj.RuntimeType(), evaluationTimestamp)
	if err != nil {
		return result.Value{}, err
	}
	if obj.Equal(maxVal) {
		return result.Value{}, fmt.Errorf("tried to compute successor for value that is already a max value, %v", obj.GolangValue())
	}

	switch t := obj.RuntimeType(); t {
	case types.Integer:
		i, err := result.ToInt32(obj)
		if err != nil {
			return result.Value{}, err
		}
		return result.New(i + 1)
	case types.Long:
		l, err := result.ToInt64(obj)
		if err != nil {
			return result.Value{}, err
		}
		return result.New(l + 1)
	case types.Decimal:
		d, err := result.ToFloat64(obj)
		if err != nil {
			return result.Value{}, err
		}
		return result.New(d + 0.00000001)
	case types.Quantity:
		// TODO: b/329707836 -  Determine under what cases quantities should be incremented by whole numbers in stead of decimals.
		q, err := result.ToQuantity(obj)
		if err != nil {
			return result.Value{}, err
		}
		return result.New(result.Quantity{Value: q.Value + 0.00000001, Unit: q.Unit})
	case types.Date, types.DateTime, types.Time:
		return dateTimeSuccessor(obj, evaluationTimestamp)
	default:
		return result.Value{}, fmt.Errorf("internal error - unsupported type %v", t)
	}
}

// dateTimePredecessor computes the predecessor for a date time value.
// Returns error for unsupported types, precisions or values out of range.
func dateTimePredecessor(dt result.Value, evaluationTimestamp *time.Time) (result.Value, error) {
	t := dt.RuntimeType()
	switch t {
	case types.Date, types.DateTime, types.Time:
		// Valid types
	default:
		return result.Value{}, fmt.Errorf("internal error - unsupported type %v for date time predecessor", t)
	}

	minVal, err := minValue(dt.RuntimeType(), evaluationTimestamp)
	if err != nil {
		return result.Value{}, err
	}
	d, err := result.ToDateTime(dt)
	if err != nil {
		return result.Value{}, err
	}
	minDt, err := result.ToDateTime(minVal)
	if err != nil {
		return result.Value{}, err
	}

	// Check if the Value is the minimum at its precision.
	minDt.Precision = d.Precision
	if cmpResult, err := compareDateTime(d, minDt); err != nil {
		return result.Value{}, err
	} else if cmpResult == leftEqualRight || cmpResult == leftBeforeRight {
		return result.Value{}, fmt.Errorf("tried to compute predecessor for %s that is already a min value for it's precision, %v", t, dt.GolangValue())
	}

	var predecessorVal result.DateTime
	switch d.Precision {
	case model.YEAR:
		predecessorVal = result.DateTime{Date: d.Date.AddDate(-1, 0, 0), Precision: d.Precision}
	case model.MONTH:
		predecessorVal = result.DateTime{Date: d.Date.AddDate(0, -1, 0), Precision: d.Precision}
	case model.DAY:
		predecessorVal = result.DateTime{Date: d.Date.AddDate(0, 0, -1), Precision: d.Precision}
	case model.HOUR:
		predecessorVal = result.DateTime{Date: d.Date.Add(-time.Hour), Precision: d.Precision}
	case model.MINUTE:
		predecessorVal = result.DateTime{Date: d.Date.Add(-time.Minute), Precision: d.Precision}
	case model.SECOND:
		predecessorVal = result.DateTime{Date: d.Date.Add(-time.Second), Precision: d.Precision}
	case model.MILLISECOND:
		predecessorVal = result.DateTime{Date: d.Date.Add(-time.Millisecond), Precision: d.Precision}
	default:
		return result.Value{}, fmt.Errorf("internal error - unsupported precision %v in %s predecessor", d.Precision, t)
	}

	switch dt.RuntimeType() {
	case types.Date:
		return result.New(result.Date{Date: predecessorVal.Date, Precision: predecessorVal.Precision})
	case types.DateTime:
		return result.New(result.DateTime{Date: predecessorVal.Date, Precision: predecessorVal.Precision})
	case types.Time:
		return result.New(result.Time{Date: predecessorVal.Date, Precision: predecessorVal.Precision})
	default:
		return result.Value{}, fmt.Errorf("internal error - unsupported type %v for date time predecessor", t)
	}
}

// dateTimeSuccessor computes the successor for a date time value.
// Returns error for unsupported types, precisions or values out of range.
func dateTimeSuccessor(dt result.Value, evaluationTimestamp *time.Time) (result.Value, error) {
	t := dt.RuntimeType()
	switch t {
	case types.Date, types.DateTime, types.Time:
		// Valid types
	default:
		return result.Value{}, fmt.Errorf("internal error - unsupported type %v for date time successor", t)
	}

	maxVal, err := maxValue(dt.RuntimeType(), evaluationTimestamp)
	if err != nil {
		return result.Value{}, err
	}
	d, err := result.ToDateTime(dt)
	if err != nil {
		return result.Value{}, err
	}
	maxDt, err := result.ToDateTime(maxVal)
	if err != nil {
		return result.Value{}, err
	}

	// Check if the Value is the maximum at its precision.
	maxDt.Precision = d.Precision
	if cmpResult, err := compareDateTime(d, maxDt); err != nil {
		return result.Value{}, err
	} else if cmpResult == leftEqualRight || cmpResult == leftAfterRight {
		return result.Value{}, fmt.Errorf("tried to compute successor for %s that is already a max value for it's precision, %v", t, dt.GolangValue())
	}

	var successorVal result.DateTime
	switch d.Precision {
	case model.YEAR:
		successorVal = result.DateTime{Date: d.Date.AddDate(1, 0, 0), Precision: d.Precision}
	case model.MONTH:
		successorVal = result.DateTime{Date: d.Date.AddDate(0, 1, 0), Precision: d.Precision}
	case model.DAY:
		successorVal = result.DateTime{Date: d.Date.AddDate(0, 0, 1), Precision: d.Precision}
	case model.HOUR:
		successorVal = result.DateTime{Date: d.Date.Add(time.Hour), Precision: d.Precision}
	case model.MINUTE:
		successorVal = result.DateTime{Date: d.Date.Add(time.Minute), Precision: d.Precision}
	case model.SECOND:
		successorVal = result.DateTime{Date: d.Date.Add(time.Second), Precision: d.Precision}
	case model.MILLISECOND:
		successorVal = result.DateTime{Date: d.Date.Add(time.Millisecond), Precision: d.Precision}
	default:
		return result.Value{}, fmt.Errorf("internal error - unsupported precision %v in date time successor", d.Precision)
	}

	switch dt.RuntimeType() {
	case types.Date:
		return result.New(result.Date{Date: successorVal.Date, Precision: successorVal.Precision})
	case types.DateTime:
		return result.New(result.DateTime{Date: successorVal.Date, Precision: successorVal.Precision})
	case types.Time:
		return result.New(result.Time{Date: successorVal.Date, Precision: successorVal.Precision})
	default:
		return result.Value{}, fmt.Errorf("internal error - unsupported type %v for date time successor", t)
	}
}

// maximum<T>() T
// https://cql.hl7.org/09-b-cqlreference.html#maximum
func (i *interpreter) evalMaxValue(m *model.MaxValue) (result.Value, error) {
	return maxValue(m.ValueType, &i.evaluationTimestamp)
}

// minimum<T>() T
// https://cql.hl7.org/09-b-cqlreference.html#minimum
func (i *interpreter) evalMinValue(m *model.MinValue) (result.Value, error) {
	return minValue(m.ValueType, &i.evaluationTimestamp)
}

// Note: For Date/Time based values, the spec states that an engine can choose to set the timezone
// to UTC for min/max values, we use the evaluation timestamp's timezone.
// We do this because when creating a literal it's also in the evaluation timestamp's timezone, and
// some external tests will fail when these are different.
func maxValue(t types.IType, evaluationTimestamp *time.Time) (result.Value, error) {
	switch t {
	case types.Integer:
		return result.New(int32(math.MaxInt32))
	case types.Long:
		return result.New(int64(math.MaxInt64))
	case types.Decimal:
		return result.New(maxDecimal)
	case types.Quantity:
		return result.New(result.Quantity{Value: maxDecimal, Unit: "1"})
	case types.Date:
		if evaluationTimestamp == nil {
			return result.Value{}, fmt.Errorf("internal error - evaluation timestamp cannot be nil for Date max value")
		}
		return result.New(result.Date{Date: time.Date(9999, 12, 31, 0, 0, 0, 0, evaluationTimestamp.Location()), Precision: model.DAY})
	case types.DateTime:
		if evaluationTimestamp == nil {
			return result.Value{}, fmt.Errorf("internal error - evaluation timestamp cannot be nil for DateTime max value")
		}
		return result.New(result.DateTime{Date: time.Date(9999, 12, 31, 23, 59, 59, 999, evaluationTimestamp.Location()), Precision: model.MILLISECOND})
	case types.Time:
		if evaluationTimestamp == nil {
			return result.Value{}, fmt.Errorf("internal error - evaluation timestamp cannot be nil for Time max value")
		}
		return result.New(result.Time{Date: time.Date(0, time.January, 1, 23, 59, 59, 999000000, evaluationTimestamp.Location()), Precision: model.MILLISECOND})
	default:
		return result.Value{}, fmt.Errorf("unsupported type, cannot compute max value for: %v", t)
	}
}

// Note: For Date/Time based values, the spec states that an engine can choose to set the timezone
// to UTC for min/max values, we use the evaluation timestamp's timezone.
// We do this because when creating a literal it's also in the evaluation timestamp's timezone, and
// some external tests will fail when these are different.
func minValue(t types.IType, evaluationTimestamp *time.Time) (result.Value, error) {
	switch t {
	case types.Integer:
		return result.New(int32(math.MinInt32))
	case types.Long:
		return result.New(int64(math.MinInt64))
	case types.Decimal:
		return result.New(minDecimal)
	case types.Quantity:
		return result.New(result.Quantity{Value: minDecimal, Unit: "1"})
	case types.Date:
		if evaluationTimestamp == nil {
			return result.Value{}, fmt.Errorf("internal error - evaluation timestamp cannot be nil for Date min value")
		}
		return result.New(result.Date{Date: time.Date(1, 1, 1, 0, 0, 0, 0, evaluationTimestamp.Location()), Precision: model.DAY})
	case types.DateTime:
		if evaluationTimestamp == nil {
			return result.Value{}, fmt.Errorf("internal error - evaluation timestamp cannot be nil for DateTime min value")
		}
		return result.New(result.DateTime{Date: time.Date(1, 1, 1, 0, 0, 0, 0, evaluationTimestamp.Location()), Precision: model.MILLISECOND})
	case types.Time:
		if evaluationTimestamp == nil {
			return result.Value{}, fmt.Errorf("internal error - evaluation timestamp cannot be nil for Time min value")
		}
		return result.New(result.Time{Date: time.Date(0, time.January, 1, 0, 0, 0, 0, evaluationTimestamp.Location()), Precision: model.MILLISECOND})
	default:
		return result.Value{}, fmt.Errorf("unsupported type, cannot compute min value for: %v", t)
	}
}

// https://cql.hl7.org/09-b-cqlreference.html#modulo
// According to the spec, "If the result of the modulo cannot be represented, or the right argument
// is 0, the result is null.".
func mod(l, r any) (result.Value, error) {
	// The modulo operator doesn't support floats so we need to call math.Mod.
	switch l.(type) {
	case int32, int:
		rVal := r.(int32)
		if rVal == 0 {
			return result.New(nil)
		}
		return result.New(l.(int32) % r.(int32))
	case int64:
		rVal := r.(int64)
		if rVal == 0 {
			return result.New(nil)
		}
		return result.New(l.(int64) % r.(int64))
	case float64:
		rVal := r.(float64)
		if rVal == 0 {
			return result.New(nil)
		}
		return result.New(math.Mod(l.(float64), r.(float64)))
	}
	return result.Value{}, fmt.Errorf("internal error - mod does not support %v", reflect.TypeOf(l))
}
