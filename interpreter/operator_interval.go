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
	"time"

	"github.com/google/cql/model"
	"github.com/google/cql/result"
	"github.com/google/cql/types"
)

// INTERVAL OPERATORS - https://cql.hl7.org/09-b-cqlreference.html#interval-operators-3

// end of(argument Interval<T>) T
// https://cql.hl7.org/09-b-cqlreference.html#end
func (i *interpreter) evalEnd(m model.IUnaryExpression, intervalObj result.Value) (result.Value, error) {
	return end(intervalObj, &i.evaluationTimestamp)
}

// end returns the upper value of the interval.
// This function wraps the complexities of null inclusive bounds as well as non-inclusive boundary
// calculation via value predecessor functionality.
func end(intervalObj result.Value, evaluationTimestamp *time.Time) (result.Value, error) {
	if result.IsNull(intervalObj) {
		return result.New(nil)
	}
	interval, err := result.ToInterval(intervalObj)
	if err != nil {
		return result.Value{}, err
	}

	if interval.HighInclusive {
		if result.IsNull(interval.High) {
			iType, ok := intervalObj.RuntimeType().(*types.Interval)
			if !ok {
				return result.Value{}, fmt.Errorf("internal error - end got Value that is not an interval type")
			}
			return maxValue(iType.PointType, evaluationTimestamp)
		}
		return interval.High, nil

	}
	if result.IsNull(interval.High) {
		return interval.High, nil
	}
	return predecessor(interval.High, evaluationTimestamp)
}

// start of(argument Interval<T>) T
// https://cql.hl7.org/09-b-cqlreference.html#start
func (i *interpreter) evalStart(m model.IUnaryExpression, intervalObj result.Value) (result.Value, error) {
	return start(intervalObj, &i.evaluationTimestamp)
}

// start returns the lower value of the interval.
// This function wraps the complexities of null inclusive bounds as well as non-inclusive boundary
// calculation via value successor functionality.
func start(intervalObj result.Value, evaluationTimestamp *time.Time) (result.Value, error) {
	if result.IsNull(intervalObj) {
		return result.New(nil)
	}
	interval, err := result.ToInterval(intervalObj)
	if err != nil {
		return result.Value{}, err
	}
	if interval.LowInclusive {
		if result.IsNull(interval.Low) {
			iType, ok := intervalObj.RuntimeType().(*types.Interval)
			if !ok {
				return result.Value{}, fmt.Errorf("internal error - start got Value that is not an interval type")
			}
			return minValue(iType.PointType, evaluationTimestamp)
		}
		return interval.Low, nil
	}
	if result.IsNull(interval.Low) {
		return interval.Low, nil
	}
	return successor(interval.Low, evaluationTimestamp)
}

// op(left DateTime, right Interval<DateTime>) Boolean
// op(left Date, right Interval<Date>) Boolean
// https://cql.hl7.org/09-b-cqlreference.html#after-1
// https://cql.hl7.org/09-b-cqlreference.html#before-1
// https://cql.hl7.org/09-b-cqlreference.html#on-or-after-2
// https://cql.hl7.org/09-b-cqlreference.html#on-or-before-2
// TODO(b/308016038): Once Start and End are properly supported, handle low/high inclusivity.
func (i *interpreter) evalCompareDateTimeInterval(be model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) || result.IsNull(rObj) {
		return result.New(nil)
	}
	l, err := result.ToDateTime(lObj)
	if err != nil {
		return result.Value{}, err
	}

	p, err := precisionFromBinaryExpression(be)
	if err != nil {
		return result.Value{}, err
	}

	allowUnsetPrec := true
	if err := validatePrecisionByType(p, allowUnsetPrec, be.Left().GetResultType()); err != nil {
		return result.Value{}, err
	}

	switch be.(type) {
	case *model.After:
		e, err := end(rObj, &i.evaluationTimestamp)
		if err != nil {
			return result.Value{}, err
		}
		rHigh, err := result.ToDateTime(e)
		if err != nil {
			return result.Value{}, err
		}
		return afterDateTimeWithPrecision(l, rHigh, p)
	case *model.Before:
		s, err := start(rObj, &i.evaluationTimestamp)
		if err != nil {
			return result.Value{}, err
		}
		rLow, err := result.ToDateTime(s)
		if err != nil {
			return result.Value{}, err
		}
		return beforeDateTimeWithPrecision(l, rLow, p)
	case *model.SameOrAfter:
		e, err := end(rObj, &i.evaluationTimestamp)
		if err != nil {
			return result.Value{}, err
		}
		rHigh, err := result.ToDateTime(e)
		if err != nil {
			return result.Value{}, err
		}
		return afterOrEqualDateTimeWithPrecision(l, rHigh, p)
	case *model.SameOrBefore:
		s, err := start(rObj, &i.evaluationTimestamp)
		if err != nil {
			return result.Value{}, err
		}
		rLow, err := result.ToDateTime(s)
		if err != nil {
			return result.Value{}, err
		}
		return beforeOrEqualDateTimeWithPrecision(l, rLow, p)
	}
	return result.Value{}, fmt.Errorf("internal error - unsupported Binary Comparison Expression %v", be)
}

// op(left Interval<DateTime>, right Interval<DateTime>) Boolean
// op(left Interval<Date>, right Interval<Date>) Boolean
// https://cql.hl7.org/09-b-cqlreference.html#after-1
// https://cql.hl7.org/09-b-cqlreference.html#before-1
// https://cql.hl7.org/09-b-cqlreference.html#on-or-after-2
// https://cql.hl7.org/09-b-cqlreference.html#on-or-before-2
// TODO(b/308016038): Once Start and End are properly supported, handle low/high inclusivity.
func (i *interpreter) evalCompareIntervalDateTimeInterval(be model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) || result.IsNull(rObj) {
		return result.New(nil)
	}
	p, err := precisionFromBinaryExpression(be)
	if err != nil {
		return result.Value{}, err
	}

	iType, ok := be.Left().GetResultType().(*types.Interval)
	if !ok {
		return result.Value{}, fmt.Errorf("internal error - evalCompareIntervalDateTimeInterval got Value that is not an interval type")
	}
	pointType := iType.PointType
	allowUnsetPrec := true
	if err := validatePrecisionByType(p, allowUnsetPrec, pointType); err != nil {
		return result.Value{}, err
	}

	switch be.(type) {
	case *model.After:
		// lObj starts after the rObj ends
		lObjStart, err := start(lObj, &i.evaluationTimestamp)
		if err != nil {
			return result.Value{}, err
		}
		rObjEnd, err := end(rObj, &i.evaluationTimestamp)
		if err != nil {
			return result.Value{}, err
		}
		lStart, rEnd, err := applyToValues(lObjStart, rObjEnd, result.ToDateTime)
		if err != nil {
			return result.Value{}, err
		}
		return afterDateTimeWithPrecision(lStart, rEnd, p)
	case *model.Before:
		// lObj ends before rObj ond one starts
		lObjEnd, err := end(lObj, &i.evaluationTimestamp)
		if err != nil {
			return result.Value{}, err
		}
		rObjStart, err := start(rObj, &i.evaluationTimestamp)
		if err != nil {
			return result.Value{}, err
		}
		lEnd, rStart, err := applyToValues(lObjEnd, rObjStart, result.ToDateTime)
		if err != nil {
			return result.Value{}, err
		}
		return beforeDateTimeWithPrecision(lEnd, rStart, p)
	case *model.SameOrAfter:
		// lObj starts on or after the rObj ends
		lObjStart, err := start(lObj, &i.evaluationTimestamp)
		if err != nil {
			return result.Value{}, err
		}
		rObjEnd, err := end(rObj, &i.evaluationTimestamp)
		if err != nil {
			return result.Value{}, err
		}
		lStart, rEnd, err := applyToValues(lObjStart, rObjEnd, result.ToDateTime)
		if err != nil {
			return result.Value{}, err
		}
		return afterOrEqualDateTimeWithPrecision(lStart, rEnd, p)
	case *model.SameOrBefore:
		// lObj ends on or before rObj one starts
		lObjEnd, err := end(lObj, &i.evaluationTimestamp)
		if err != nil {
			return result.Value{}, err
		}
		rObjStart, err := start(rObj, &i.evaluationTimestamp)
		if err != nil {
			return result.Value{}, err
		}
		lEnd, rStart, err := applyToValues(lObjEnd, rObjStart, result.ToDateTime)
		if err != nil {
			return result.Value{}, err
		}
		return beforeOrEqualDateTimeWithPrecision(lEnd, rStart, p)
	}
	return result.Value{}, fmt.Errorf("internal error - unsupported Binary Comparison Expression in evalCompareIntervalDateTimeInterval: %v", be)
}

// in _precision_ (point Decimal, argument Interval<Decimal>) Boolean
// in _precision_ (point Long, argument Interval<Long>) Boolean
// in _precision_ (point Integer, argument Interval<Integer>) Boolean
// in _precision_ (point Quantity, argument Interval<Quantity>) Boolean
// https://cql.hl7.org/09-b-cqlreference.html#in
// 'Contains' with the left and right args reversed is forwarded here.
func evalInIntervalNumeral(b model.IBinaryExpression, pointObj, intervalObj result.Value) (result.Value, error) {
	if result.IsNull(pointObj) {
		return result.New(nil)
	}
	if result.IsNull(intervalObj) {
		return result.New(false)
	}

	// start and end handles null inclusivity as well as non-inclusive logic.
	s, err := start(intervalObj, nil)
	if err != nil {
		return result.Value{}, err
	}
	e, err := end(intervalObj, nil)
	if err != nil {
		return result.Value{}, err
	}
	// This will only happen for null exclusive bounds.
	// TODO b/335910011 - fix not inclusive but outside of the opposite bounds.
	if result.IsNull(s) || result.IsNull(e) {
		return result.New(nil)
	}

	switch pointObj.RuntimeType() {
	case types.Decimal:
		point, err := result.ToFloat64(pointObj)
		if err != nil {
			return result.Value{}, err
		}
		startVal, endVal, err := applyToValues(s, e, result.ToFloat64)
		if err != nil {
			return result.Value{}, err
		}
		return numeralInInterval(point, startVal, endVal)
	case types.Integer:
		point, err := result.ToInt32(pointObj)
		if err != nil {
			return result.Value{}, err
		}
		startVal, endVal, err := applyToValues(s, e, result.ToInt32)
		if err != nil {
			return result.Value{}, err
		}
		return numeralInInterval(point, startVal, endVal)
	case types.Long:
		point, err := result.ToInt64(pointObj)
		if err != nil {
			return result.Value{}, err
		}
		startVal, endVal, err := applyToValues(s, e, result.ToInt64)
		if err != nil {
			return result.Value{}, err
		}
		return numeralInInterval(point, startVal, endVal)
	case types.Quantity:
		point, err := result.ToQuantity(pointObj)
		if err != nil {
			return result.Value{}, err
		}
		startVal, endVal, err := applyToValues(s, e, result.ToQuantity)
		if err != nil {
			return result.Value{}, err
		}
		if point.Unit != startVal.Unit {
			return result.Value{}, fmt.Errorf("in operator recieved Quantities with differing unit values, unit conversion is not currently supported, got: %v, %v", point.Unit, startVal.Unit)
		}
		if point.Unit != endVal.Unit {
			return result.Value{}, fmt.Errorf("in operator recieved Quantities with differing unit values, unit conversion is not currently supported, got: %v, %v", point.Unit, endVal.Unit)
		}
		return numeralInInterval(point.Value, startVal.Value, endVal.Value)
	default:
		return result.Value{}, fmt.Errorf("internal error - unsupported point type in evalInIntervalNumeral: %v", pointObj.RuntimeType())
	}
}

func numeralInInterval[t float64 | int64 | int32](point, startVal, endVal t) (result.Value, error) {
	// This should only happen in cases such as Interval[1, 1).
	if compareNumeral(startVal, endVal) == leftAfterRight {
		return result.New(false)
	}
	lowCompare := compareNumeral(point, startVal)
	highCompare := compareNumeral(point, endVal)
	return inInterval(lowCompare, highCompare, true, true)
}

func compareNumeral[t float64 | int64 | int32](left, right t) comparison {
	if left == right {
		return leftEqualRight
	} else if left < right {
		return leftBeforeRight
	}
	return leftAfterRight
}

// in _precision_ (point DateTime, argument Interval<DateTime>) Boolean
// in _precision_ (point Date, argument Interval<Date>) Boolean
// https://cql.hl7.org/09-b-cqlreference.html#in
// 'IncludedIn' with left arg of point type is forwarded here.
// 'Contains' with the left and right args reversed is forwarded here.
func (i *interpreter) evalInIntervalDateTime(b model.IBinaryExpression, pointObj, intervalObj result.Value) (result.Value, error) {
	m := b.(*model.In)
	precision := model.DateTimePrecision(m.Precision)
	allowUnsetPrec := true
	if err := validatePrecisionByType(precision, allowUnsetPrec, m.Left().GetResultType()); err != nil {
		return result.Value{}, err
	}

	if result.IsNull(pointObj) {
		return result.New(nil)
	}
	if result.IsNull(intervalObj) {
		return result.New(false)
	}

	point, err := result.ToDateTime(pointObj)
	if err != nil {
		return result.Value{}, err
	}
	interval, err := result.ToInterval(intervalObj)
	if err != nil {
		return result.Value{}, err
	}

	var lowCompare, highCompare comparison
	s, err := start(intervalObj, &i.evaluationTimestamp)
	if err != nil {
		return result.Value{}, err
	}
	if result.IsNull(s) {
		lowCompare = comparedToNull
	} else {
		low, err := result.ToDateTime(s)
		if err != nil {
			return result.Value{}, err
		}
		lowCompare, err = compareDateTimeWithPrecision(point, low, precision)
		if err != nil {
			return result.Value{}, err
		}
	}

	e, err := end(intervalObj, &i.evaluationTimestamp)
	if err != nil {
		return result.Value{}, err
	}
	if result.IsNull(e) {
		highCompare = comparedToNull
	} else {
		high, err := result.ToDateTime(e)
		if err != nil {
			return result.Value{}, err
		}
		highCompare, err = compareDateTimeWithPrecision(point, high, precision)
		if err != nil {
			return result.Value{}, err
		}
	}

	return inInterval(lowCompare, highCompare, interval.LowInclusive, interval.HighInclusive)
}

func inInterval(lowCompare, highCompare comparison, lowInclusive, highInclusive bool) (result.Value, error) {
	// This includes cases where we know the point is for sure outside the interval such as:
	// 5 in Interval[0, 2] - point is outside the interval
	if lowCompare == leftBeforeRight || highCompare == leftAfterRight {
		return result.New(false)
	}

	// Handles Cases:
	// 3 in Interval[0, 3) - point is on the exclusive bound
	// 3 in Interval[3, 3) - ignores cases like this, the will fall through to null
	if (lowCompare == leftEqualRight && !lowInclusive) && !(highCompare == leftEqualRight && highInclusive) {
		return result.New(false)
	}
	if (highCompare == leftEqualRight && !highInclusive) && !(lowCompare == leftEqualRight && lowInclusive) {
		return result.New(false)
	}

	// This handles three cases:
	// 3 in Interval[0, 5] - point is within the interval
	// 3 in Interval[0, 3] - point is on the boundary but the boundary is inclusive
	if lowCompare == leftAfterRight || (lowInclusive && lowCompare == leftEqualRight) {
		if highCompare == leftBeforeRight || (highInclusive && highCompare == leftEqualRight) {
			return result.New(true)
		}
	}

	// All other cases return null, this includes:
	// * Cases where a start/end bound was null: 3 in Interval(null, 5]
	// * Cases where Dates/DateTimes have insufficient precision for the comparison:
	//   Date(2020) in Interval[Date(2020, 3), Date(2020, 4)]
	return result.New(nil)
}
