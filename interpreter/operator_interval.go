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
	"github.com/google/cql/ucum"
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

// startAndEnd returns the start and end of the interval.
// This function is a helper for calling start() and end() in the same function.
func startAndEnd(intervalObj result.Value, evaluationTimestamp *time.Time) (result.Value, result.Value, error) {
	start, err := start(intervalObj, evaluationTimestamp)
	if err != nil {
		return result.Value{}, result.Value{}, err
	}
	end, err := end(intervalObj, evaluationTimestamp)
	if err != nil {
		return result.Value{}, result.Value{}, err
	}
	return start, end, nil
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

// Overlaps(left Interval<Date>, right Interval<Date>) Boolean
// Overlaps(left Interval<DateTime>, right Interval<DateTime>) Boolean
// https://cql.hl7.org/09-b-cqlreference.html#overlaps
func (i *interpreter) evalOverlapsIntervalDateTimeInterval(be model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
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
	if p != "" {
		return result.Value{}, fmt.Errorf("internal error - overlaps does not yet support precision")
	}

	// Get left interval bounds.
	lObjStart, lObjEnd, err := startAndEnd(lObj, &i.evaluationTimestamp)
	if err != nil {
		return result.Value{}, err
	}
	leftStart, leftEnd, err := applyToValues(lObjStart, lObjEnd, result.ToDateTime)
	if err != nil {
		return result.Value{}, err
	}

	// Get right interval bounds.
	rObjStart, rObjEnd, err := startAndEnd(rObj, &i.evaluationTimestamp)
	if err != nil {
		return result.Value{}, err
	}
	rightStart, rightEnd, err := applyToValues(rObjStart, rObjEnd, result.ToDateTime)
	if err != nil {
		return result.Value{}, err
	}

	// Due to complexity regarding DateTime precision, we will calculate each case separately and
	// return the OR of all results. If any of the cases are true, then the result is true, if any
	// of the cases are null, then the result is null, otherwise the result is false.
	compResults := []result.Value{}
	// Case 1. Left starts during right interval.
	leftStartsDuringRightIntervalValue, err := dateTimeInIntervalWithPrecision(leftStart, rightStart, rightEnd, p)
	if err != nil {
		return result.Value{}, err
	}
	compResults = append(compResults, leftStartsDuringRightIntervalValue)

	// Case 2. Left ends during right interval.
	leftEndsDuringRightIntervalValue, err := dateTimeInIntervalWithPrecision(leftEnd, rightStart, rightEnd, p)
	if err != nil {
		return result.Value{}, err
	}
	compResults = append(compResults, leftEndsDuringRightIntervalValue)

	// Case 3. Right starts during left interval.
	rightStartsDuringLeftIntervalValue, err := dateTimeInIntervalWithPrecision(rightStart, leftStart, leftEnd, p)
	if err != nil {
		return result.Value{}, err
	}
	compResults = append(compResults, rightStartsDuringLeftIntervalValue)

	// Case 4. Right ends during left interval.
	rightEndsDuringLeftIntervalValue, err := dateTimeInIntervalWithPrecision(rightEnd, leftStart, leftEnd, p)
	if err != nil {
		return result.Value{}, err
	}
	compResults = append(compResults, rightEndsDuringLeftIntervalValue)

	trueVal, err := result.New(true)
	if err != nil {
		return result.Value{}, err
	}
	nullVal, err := result.New(nil)
	if err != nil {
		return result.Value{}, err
	}
	if valueInList(trueVal, compResults) {
		return trueVal, nil
	} else if valueInList(nullVal, compResults) {
		return nullVal, nil
	}
	return result.New(false)
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
	s, e, err := startAndEnd(intervalObj, nil)
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

// duringDateTimeWithPrecision returns whether or not the given DateTimeValue is during the given
// low high interval. Returns null in cases where values cannot be compared such as right precision being
// less than left precision.
// All values are expected to be inclusive bounds.
// Return a null value if the comparison cannot be made due to insufficient precision.
func dateTimeInIntervalWithPrecision(a, low, high result.DateTime, p model.DateTimePrecision) (result.Value, error) {
	lowComp, err := compareDateTimeWithPrecision(a, low, p)
	if err != nil {
		return result.Value{}, err
	}
	highComp, err := compareDateTimeWithPrecision(a, high, p)
	if err != nil {
		return result.Value{}, err
	}

	if lowComp == insufficientPrecision || highComp == insufficientPrecision {
		return result.New(nil)
	} else if (lowComp == leftEqualRight || lowComp == leftAfterRight) && (highComp == leftEqualRight || highComp == leftBeforeRight) {
		return result.New(true)
	}
	return result.New(false)
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

// width of(argument Interval<T>) T
// https://cql.hl7.org/09-b-cqlreference.html#width
func evalWidthInterval(m model.IUnaryExpression, intervalObj result.Value) (result.Value, error) {
	if result.IsNull(intervalObj) {
		return result.New(nil)
	}
	interval, err := result.ToInterval(intervalObj)
	if err != nil {
		return result.Value{}, err
	}
	if interval.StaticType.PointType == types.Date || interval.StaticType.PointType == types.DateTime || interval.StaticType.PointType == types.Time {
		return result.Value{}, fmt.Errorf("width operator does not support Date or Time types")
	}
	start, err := start(intervalObj, nil)
	if err != nil {
		return result.Value{}, err
	}
	if result.IsNull(start) {
		return result.New(nil)
	}
	end, err := end(intervalObj, nil)
	if err != nil {
		return result.Value{}, err
	}
	if result.IsNull(end) {
		return result.New(nil)
	}
	switch start.RuntimeType() {
	case types.Decimal:
		startVal, endVal, err := applyToValues(start, end, result.ToFloat64)
		if err != nil {
			return result.Value{}, err
		}
		return result.New(endVal - startVal)
	case types.Integer:
		startVal, endVal, err := applyToValues(start, end, result.ToInt32)
		if err != nil {
			return result.Value{}, err
		}
		return result.New(endVal - startVal)
	case types.Long:
		startVal, endVal, err := applyToValues(start, end, result.ToInt64)
		if err != nil {
			return result.Value{}, err
		}
		return result.New(endVal - startVal)
	case types.Quantity:
		startVal, endVal, err := applyToValues(start, end, result.ToQuantity)
		if err != nil {
			return result.Value{}, err
		}
		// for now naively convery left unit to right unit.
		convertedStartVal, err := ucum.ConvertUnit(startVal.Value, string(startVal.Unit), string(endVal.Unit))
		if err != nil {
			return result.Value{}, err
		}
		return result.New(result.Quantity{Value: endVal.Value - convertedStartVal, Unit: endVal.Unit})
	}
	return result.Value{}, fmt.Errorf("internal error - unsupported point type in evalWidthInterval: %v", start.RuntimeType())
}

// evalExceptIntervalInteger handles except for integer intervals
func (i *interpreter) evalExceptIntervalInteger(leftStart, leftEnd, rightStart, rightEnd result.Value, leftInterval *result.Interval) (result.Value, error) {
	ls, le, rs, re, err := applyToValues4(leftStart, leftEnd, rightStart, rightEnd, result.ToInt32)
	if err != nil {
		return result.Value{}, err
	}

	// No overlap cases
	if le < rs || ls > re {
		return createCleanInterval(leftStart, leftEnd, leftInterval.LowInclusive, leftInterval.HighInclusive, leftInterval.StaticType)
	}

	// Complete overlap - left is completely contained in right
	if rs <= ls && le <= re {
		return result.New(nil)
	}

	// Properly contained check - right is properly contained in left
	if ls < rs && re < le {
		return result.New(nil)
	}

	// Partial overlap cases
	if rs <= ls && ls < re && re < le {
		// Left overlap: return [re+1, le]
		newStart, err := result.New(re + 1)
		if err != nil {
			return result.Value{}, err
		}
		return createCleanInterval(newStart, leftEnd, true, leftInterval.HighInclusive, leftInterval.StaticType)
	}

	if ls < rs && rs < le && le <= re {
		// Right overlap: return [ls, rs-1]
		newEnd, err := result.New(rs - 1)
		if err != nil {
			return result.Value{}, err
		}
		return createCleanInterval(leftStart, newEnd, leftInterval.LowInclusive, true, leftInterval.StaticType)
	}

	return result.New(nil)
}

// evalExceptIntervalDate handles except for date intervals
func (i *interpreter) evalExceptIntervalDate(leftStart, leftEnd, rightStart, rightEnd result.Value, leftInterval *result.Interval) (result.Value, error) {
	ls, le, rs, re, err := applyToValues4(leftStart, leftEnd, rightStart, rightEnd, result.ToDateTime)
	if err != nil {
		return result.Value{}, err
	}

	// No overlap cases
	comp1, err := compareDateTimeWithPrecision(le, rs, model.DAY)
	if err != nil {
		return result.Value{}, err
	}
	comp2, err := compareDateTimeWithPrecision(ls, re, model.DAY)
	if err != nil {
		return result.Value{}, err
	}
	if comp1 == leftBeforeRight || comp2 == leftAfterRight {
		return createCleanInterval(leftStart, leftEnd, leftInterval.LowInclusive, leftInterval.HighInclusive, leftInterval.StaticType)
	}

	// Complete overlap - left is completely contained in right
	comp3, err := compareDateTimeWithPrecision(rs, ls, model.DAY)
	if err != nil {
		return result.Value{}, err
	}
	comp4, err := compareDateTimeWithPrecision(le, re, model.DAY)
	if err != nil {
		return result.Value{}, err
	}
	if (comp3 == leftBeforeRight || comp3 == leftEqualRight) && (comp4 == leftBeforeRight || comp4 == leftEqualRight) {
		return result.New(nil)
	}

	// Properly contained check - right is properly contained in left
	comp5, err := compareDateTimeWithPrecision(ls, rs, model.DAY)
	if err != nil {
		return result.Value{}, err
	}
	comp6, err := compareDateTimeWithPrecision(re, le, model.DAY)
	if err != nil {
		return result.Value{}, err
	}
	if comp5 == leftBeforeRight && comp6 == leftBeforeRight {
		return result.New(nil)
	}

	// TODO: For dates, we need to handle partial overlaps carefully
	return result.New(nil)
}

// Helper function to apply a conversion function to 4 values
func applyToValues4[T any](v1, v2, v3, v4 result.Value, converter func(result.Value) (T, error)) (T, T, T, T, error) {
	var zero T
	r1, err := converter(v1)
	if err != nil {
		return zero, zero, zero, zero, err
	}
	r2, err := converter(v2)
	if err != nil {
		return zero, zero, zero, zero, err
	}
	r3, err := converter(v3)
	if err != nil {
		return zero, zero, zero, zero, err
	}
	r4, err := converter(v4)
	if err != nil {
		return zero, zero, zero, zero, err
	}
	return r1, r2, r3, r4, nil
}

// createCleanInterval creates a new interval without source metadata
func createCleanInterval(low, high result.Value, lowInclusive, highInclusive bool, staticType *types.Interval) (result.Value, error) {
	return result.New(result.Interval{
		Low:           low,
		High:          high,
		LowInclusive:  lowInclusive,
		HighInclusive: highInclusive,
		StaticType:    staticType,
	})
}

// evalExceptIntervalNumeral handles except for numeric intervals (Integer, Long, Decimal)
func evalExceptIntervalNumeral(m model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	// Handle null inputs
	if result.IsNull(lObj) || result.IsNull(rObj) {
		return result.New(nil)
	}

	// Get interval bounds
	leftStart, leftEnd, err := startAndEnd(lObj, nil)
	if err != nil {
		return result.Value{}, err
	}
	rightStart, rightEnd, err := startAndEnd(rObj, nil)
	if err != nil {
		return result.Value{}, err
	}

	// If any bound is null, return null
	if result.IsNull(leftStart) || result.IsNull(leftEnd) || result.IsNull(rightStart) || result.IsNull(rightEnd) {
		return result.New(nil)
	}

	// Get the interval type for creating result
	leftInterval, err := result.ToInterval(lObj)
	if err != nil {
		return result.Value{}, err
	}

	// Dispatch based on point type
	switch leftStart.RuntimeType() {
	case types.Integer:
		ls, le, rs, re, err := applyToValues4(leftStart, leftEnd, rightStart, rightEnd, result.ToInt32)
		if err != nil {
			return result.Value{}, err
		}
		return evalExceptIntervalNumeralLogic(ls, le, rs, re, leftStart, leftEnd, &leftInterval, func(val int32) (result.Value, error) {
			return result.New(val + 1)
		}, func(val int32) (result.Value, error) {
			return result.New(val - 1)
		})
	case types.Long:
		ls, le, rs, re, err := applyToValues4(leftStart, leftEnd, rightStart, rightEnd, result.ToInt64)
		if err != nil {
			return result.Value{}, err
		}
		return evalExceptIntervalNumeralLogic(ls, le, rs, re, leftStart, leftEnd, &leftInterval, func(val int64) (result.Value, error) {
			return result.New(val + 1)
		}, func(val int64) (result.Value, error) {
			return result.New(val - 1)
		})
	case types.Decimal:
		ls, le, rs, re, err := applyToValues4(leftStart, leftEnd, rightStart, rightEnd, result.ToFloat64)
		if err != nil {
			return result.Value{}, err
		}
		const epsilon = 1e-8
		return evalExceptIntervalNumeralLogic(ls, le, rs, re, leftStart, leftEnd, &leftInterval, func(val float64) (result.Value, error) {
			return result.New(val + epsilon)
		}, func(val float64) (result.Value, error) {
			return result.New(val - epsilon)
		})
	default:
		return result.Value{}, fmt.Errorf("unsupported numeric interval point type for except: %v", leftStart.RuntimeType())
	}
}

// evalExceptIntervalNumeralLogic contains the core logic for numeric interval except operations
func evalExceptIntervalNumeralLogic[T comparable](ls, le, rs, re T, leftStart, leftEnd result.Value, leftInterval *result.Interval, successor func(T) (result.Value, error), predecessor func(T) (result.Value, error)) (result.Value, error) {
	// No overlap cases
	if compareValues(le, rs) < 0 || compareValues(ls, re) > 0 {
		return createCleanInterval(leftStart, leftEnd, leftInterval.LowInclusive, leftInterval.HighInclusive, leftInterval.StaticType)
	}

	// Complete overlap - left is completely contained in right
	if compareValues(rs, ls) <= 0 && compareValues(le, re) <= 0 {
		return result.New(nil)
	}

	// Properly contained check - right is properly contained in left
	if compareValues(ls, rs) < 0 && compareValues(re, le) < 0 {
		return result.New(nil)
	}

	// Partial overlap cases
	if compareValues(rs, ls) <= 0 && compareValues(ls, re) < 0 && compareValues(re, le) < 0 {
		// Left overlap: return [re+1, le]
		newStart, err := successor(re)
		if err != nil {
			return result.Value{}, err
		}
		return createCleanInterval(newStart, leftEnd, true, leftInterval.HighInclusive, leftInterval.StaticType)
	}

	if compareValues(ls, rs) < 0 && compareValues(rs, le) < 0 && compareValues(le, re) <= 0 {
		// Right overlap: return [ls, rs-1]
		newEnd, err := predecessor(rs)
		if err != nil {
			return result.Value{}, err
		}
		return createCleanInterval(leftStart, newEnd, leftInterval.LowInclusive, true, leftInterval.StaticType)
	}

	return result.New(nil)
}

// compareValues compares two values of the same type
func compareValues[T comparable](a, b T) int {
	switch any(a).(type) {
	case int32:
		aVal := any(a).(int32)
		bVal := any(b).(int32)
		if aVal < bVal {
			return -1
		} else if aVal > bVal {
			return 1
		}
		return 0
	case int64:
		aVal := any(a).(int64)
		bVal := any(b).(int64)
		if aVal < bVal {
			return -1
		} else if aVal > bVal {
			return 1
		}
		return 0
	case float64:
		aVal := any(a).(float64)
		bVal := any(b).(float64)
		if aVal < bVal {
			return -1
		} else if aVal > bVal {
			return 1
		}
		return 0
	default:
		// For other types, use basic comparison
		if any(a) == any(b) {
			return 0
		}
		// This is a fallback - in practice we should handle all numeric types above
		return -1
	}
}

// evalExceptIntervalQuantity handles except for quantity intervals
func evalExceptIntervalQuantity(m model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	// Handle null inputs
	if result.IsNull(lObj) || result.IsNull(rObj) {
		return result.New(nil)
	}

	// Get interval bounds
	leftStart, leftEnd, err := startAndEnd(lObj, nil)
	if err != nil {
		return result.Value{}, err
	}
	rightStart, rightEnd, err := startAndEnd(rObj, nil)
	if err != nil {
		return result.Value{}, err
	}

	// If any bound is null, return null
	if result.IsNull(leftStart) || result.IsNull(leftEnd) || result.IsNull(rightStart) || result.IsNull(rightEnd) {
		return result.New(nil)
	}

	// Get the interval type for creating result
	leftInterval, err := result.ToInterval(lObj)
	if err != nil {
		return result.Value{}, err
	}

	ls, le, rs, re, err := applyToValues4(leftStart, leftEnd, rightStart, rightEnd, result.ToQuantity)
	if err != nil {
		return result.Value{}, err
	}

	// Check unit compatibility
	if ls.Unit != rs.Unit || le.Unit != re.Unit {
		return result.Value{}, fmt.Errorf("except operator received Quantities with differing unit values")
	}

	// No overlap cases
	if le.Value < rs.Value || ls.Value > re.Value {
		return createCleanInterval(leftStart, leftEnd, leftInterval.LowInclusive, leftInterval.HighInclusive, leftInterval.StaticType)
	}

	// Complete overlap - left is completely contained in right
	if rs.Value <= ls.Value && le.Value <= re.Value {
		return result.New(nil)
	}

	// Properly contained check - right is properly contained in left
	if ls.Value < rs.Value && re.Value < le.Value {
		return result.New(nil)
	}

	// Partial overlap cases
	const epsilon = 1e-8
	if rs.Value <= ls.Value && ls.Value < re.Value && re.Value < le.Value {
		// Left overlap: return [re+epsilon, le]
		newStart, err := result.New(result.Quantity{Value: re.Value + epsilon, Unit: re.Unit})
		if err != nil {
			return result.Value{}, err
		}
		return createCleanInterval(newStart, leftEnd, true, leftInterval.HighInclusive, leftInterval.StaticType)
	}

	if ls.Value < rs.Value && rs.Value < le.Value && le.Value <= re.Value {
		// Right overlap: return [ls, rs-epsilon]
		newEnd, err := result.New(result.Quantity{Value: rs.Value - epsilon, Unit: rs.Unit})
		if err != nil {
			return result.Value{}, err
		}
		return createCleanInterval(leftStart, newEnd, leftInterval.LowInclusive, true, leftInterval.StaticType)
	}

	return result.New(nil)
}

// evalExceptIntervalDateTime handles except for date/datetime/time intervals
func (i *interpreter) evalExceptIntervalDateTime(m model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	// Handle null inputs
	if result.IsNull(lObj) || result.IsNull(rObj) {
		return result.New(nil)
	}

	// Get interval bounds
	leftStart, leftEnd, err := startAndEnd(lObj, &i.evaluationTimestamp)
	if err != nil {
		return result.Value{}, err
	}
	rightStart, rightEnd, err := startAndEnd(rObj, &i.evaluationTimestamp)
	if err != nil {
		return result.Value{}, err
	}

	// If any bound is null, return null
	if result.IsNull(leftStart) || result.IsNull(leftEnd) || result.IsNull(rightStart) || result.IsNull(rightEnd) {
		return result.New(nil)
	}

	// Get the interval type for creating result
	leftInterval, err := result.ToInterval(lObj)
	if err != nil {
		return result.Value{}, err
	}

	// Determine the precision based on the point type
	var precision model.DateTimePrecision
	switch leftStart.RuntimeType() {
	case types.Date:
		precision = model.DAY
	case types.DateTime:
		precision = model.DAY
	case types.Time:
		precision = model.MILLISECOND
	default:
		return result.Value{}, fmt.Errorf("unsupported temporal type for except: %v", leftStart.RuntimeType())
	}

	ls, le, rs, re, err := applyToValues4(leftStart, leftEnd, rightStart, rightEnd, result.ToDateTime)
	if err != nil {
		return result.Value{}, err
	}

	// No overlap cases
	comp1, err := compareDateTimeWithPrecision(le, rs, precision)
	if err != nil {
		return result.Value{}, err
	}
	comp2, err := compareDateTimeWithPrecision(ls, re, precision)
	if err != nil {
		return result.Value{}, err
	}
	if comp1 == leftBeforeRight || comp2 == leftAfterRight {
		return createCleanInterval(leftStart, leftEnd, leftInterval.LowInclusive, leftInterval.HighInclusive, leftInterval.StaticType)
	}

	// Complete overlap - left is completely contained in right
	comp3, err := compareDateTimeWithPrecision(rs, ls, precision)
	if err != nil {
		return result.Value{}, err
	}
	comp4, err := compareDateTimeWithPrecision(le, re, precision)
	if err != nil {
		return result.Value{}, err
	}
	if (comp3 == leftBeforeRight || comp3 == leftEqualRight) && (comp4 == leftBeforeRight || comp4 == leftEqualRight) {
		return result.New(nil)
	}

	// Properly contained check - right is properly contained in left
	comp5, err := compareDateTimeWithPrecision(ls, rs, precision)
	if err != nil {
		return result.Value{}, err
	}
	comp6, err := compareDateTimeWithPrecision(re, le, precision)
	if err != nil {
		return result.Value{}, err
	}
	if comp5 == leftBeforeRight && comp6 == leftBeforeRight {
		return result.New(nil)
	}

	// Partial overlap cases - need to handle date/time arithmetic
	// Left overlap: rs <= ls < re < le -> return [re+1, le]
	comp7, err := compareDateTimeWithPrecision(rs, ls, precision)
	if err != nil {
		return result.Value{}, err
	}
	comp8, err := compareDateTimeWithPrecision(ls, re, precision)
	if err != nil {
		return result.Value{}, err
	}
	comp9, err := compareDateTimeWithPrecision(re, le, precision)
	if err != nil {
		return result.Value{}, err
	}
	if (comp7 == leftBeforeRight || comp7 == leftEqualRight) && comp8 == leftBeforeRight && comp9 == leftBeforeRight {
		// Calculate the successor of re
		newStart, err := successor(rightEnd, &i.evaluationTimestamp)
		if err != nil {
			return result.Value{}, err
		}
		return createCleanInterval(newStart, leftEnd, true, leftInterval.HighInclusive, leftInterval.StaticType)
	}

	// Right overlap: ls < rs < le <= re -> return [ls, rs-1]
	comp10, err := compareDateTimeWithPrecision(ls, rs, precision)
	if err != nil {
		return result.Value{}, err
	}
	comp11, err := compareDateTimeWithPrecision(rs, le, precision)
	if err != nil {
		return result.Value{}, err
	}
	comp12, err := compareDateTimeWithPrecision(le, re, precision)
	if err != nil {
		return result.Value{}, err
	}
	if comp10 == leftBeforeRight && comp11 == leftBeforeRight && (comp12 == leftBeforeRight || comp12 == leftEqualRight) {
		// Calculate the predecessor of rs
		newEnd, err := predecessor(rightStart, &i.evaluationTimestamp)
		if err != nil {
			return result.Value{}, err
		}
		return createCleanInterval(leftStart, newEnd, leftInterval.LowInclusive, true, leftInterval.StaticType)
	}

	// TODO: if we get here, it's a complex case that we can't handle yet
	return result.New(nil)
}
