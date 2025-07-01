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

// intersect(left Interval<T>, right Interval<T>) Interval<T>
// https://cql.hl7.org/09-b-cqlreference.html#intersect
// This function is used only for Date, DateTime, and Time intervals
func (i *interpreter) evalIntersectInterval(m model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	// Handle null inputs
	if result.IsNull(lObj) || result.IsNull(rObj) {
		return result.New(nil)
	}

	// Get start and end bounds for both intervals
	lStart, lEnd, err := startAndEnd(lObj, &i.evaluationTimestamp)
	if err != nil {
		return result.Value{}, err
	}
	rStart, rEnd, err := startAndEnd(rObj, &i.evaluationTimestamp)
	if err != nil {
		return result.Value{}, err
	}

	// If any bound is null, return null
	if result.IsNull(lStart) || result.IsNull(lEnd) || result.IsNull(rStart) || result.IsNull(rEnd) {
		return result.New(nil)
	}

	// This function only handles Date, DateTime, and Time intervals
	return i.evalIntersectIntervalDateTime(m, lObj, rObj)
}

// calculateNumeralIntersectionInt32 calculates the intersection of two int32 intervals
func (i *interpreter) calculateNumeralIntersectionInt32(lStart, lEnd, rStart, rEnd int32, lInterval, rInterval result.Interval) (result.Value, error) {
	// Calculate intersection bounds
	intersectionStart := maxInt32(lStart, rStart)
	intersectionEnd := minInt32(lEnd, rEnd)

	// Check if there's no overlap
	if compareNumeral(intersectionStart, intersectionEnd) == leftAfterRight {
		return result.New(nil)
	}

	// For intersection, the result bounds are always inclusive
	// This is because we're using the effective start/end values from startAndEnd()
	// which already account for the original inclusivity
	startInclusive := true
	endInclusive := true

	// Create result values
	startVal, err := result.New(intersectionStart)
	if err != nil {
		return result.Value{}, err
	}
	endVal, err := result.New(intersectionEnd)
	if err != nil {
		return result.Value{}, err
	}

	// Create the intersection interval
	intersectionInterval := result.Interval{
		Low:           startVal,
		High:          endVal,
		LowInclusive:  startInclusive,
		HighInclusive: endInclusive,
		StaticType:    lInterval.StaticType,
	}

	return result.New(intersectionInterval)
}

// calculateNumeralIntersectionInt64 calculates the intersection of two int64 intervals
func (i *interpreter) calculateNumeralIntersectionInt64(lStart, lEnd, rStart, rEnd int64, lInterval, rInterval result.Interval) (result.Value, error) {
	// Calculate intersection bounds
	intersectionStart := maxInt64(lStart, rStart)
	intersectionEnd := minInt64(lEnd, rEnd)

	// Check if there's no overlap
	if compareNumeral(intersectionStart, intersectionEnd) == leftAfterRight {
		return result.New(nil)
	}

	// Calculate inclusivity for intersection bounds
	// Start is inclusive if it matches an inclusive bound from either interval
	startInclusive := true
	if compareNumeral(intersectionStart, lStart) == leftEqualRight {
		startInclusive = lInterval.LowInclusive
	}
	if compareNumeral(intersectionStart, rStart) == leftEqualRight {
		startInclusive = startInclusive && rInterval.LowInclusive
	}

	// End is inclusive if it matches an inclusive bound from either interval
	endInclusive := true
	if compareNumeral(intersectionEnd, lEnd) == leftEqualRight {
		endInclusive = lInterval.HighInclusive
	}
	if compareNumeral(intersectionEnd, rEnd) == leftEqualRight {
		endInclusive = endInclusive && rInterval.HighInclusive
	}

	// Create result values
	startVal, err := result.New(intersectionStart)
	if err != nil {
		return result.Value{}, err
	}
	endVal, err := result.New(intersectionEnd)
	if err != nil {
		return result.Value{}, err
	}

	// Create the intersection interval
	intersectionInterval := result.Interval{
		Low:           startVal,
		High:          endVal,
		LowInclusive:  startInclusive,
		HighInclusive: endInclusive,
		StaticType:    lInterval.StaticType,
	}

	return result.New(intersectionInterval)
}

// calculateNumeralIntersectionFloat64 calculates the intersection of two float64 intervals
func (i *interpreter) calculateNumeralIntersectionFloat64(lStart, lEnd, rStart, rEnd float64, lInterval, rInterval result.Interval) (result.Value, error) {
	// Calculate intersection bounds
	intersectionStart := maxFloat64(lStart, rStart)
	intersectionEnd := minFloat64(lEnd, rEnd)

	// Check if there's no overlap
	if compareNumeral(intersectionStart, intersectionEnd) == leftAfterRight {
		return result.New(nil)
	}

	// Create intersection interval with appropriate inclusivity
	startInclusive := (compareNumeral(intersectionStart, lStart) == leftEqualRight && lInterval.LowInclusive) ||
		(compareNumeral(intersectionStart, rStart) == leftEqualRight && rInterval.LowInclusive) ||
		(compareNumeral(intersectionStart, lStart) == leftAfterRight && compareNumeral(intersectionStart, lEnd) == leftBeforeRight) ||
		(compareNumeral(intersectionStart, rStart) == leftAfterRight && compareNumeral(intersectionStart, rEnd) == leftBeforeRight)

	endInclusive := (compareNumeral(intersectionEnd, lEnd) == leftEqualRight && lInterval.HighInclusive) ||
		(compareNumeral(intersectionEnd, rEnd) == leftEqualRight && rInterval.HighInclusive) ||
		(compareNumeral(intersectionEnd, lStart) == leftAfterRight && compareNumeral(intersectionEnd, lEnd) == leftBeforeRight) ||
		(compareNumeral(intersectionEnd, rStart) == leftAfterRight && compareNumeral(intersectionEnd, rEnd) == leftBeforeRight)

	// Create result values
	startVal, err := result.New(intersectionStart)
	if err != nil {
		return result.Value{}, err
	}
	endVal, err := result.New(intersectionEnd)
	if err != nil {
		return result.Value{}, err
	}

	// Create the intersection interval
	intersectionInterval := result.Interval{
		Low:           startVal,
		High:          endVal,
		LowInclusive:  startInclusive,
		HighInclusive: endInclusive,
		StaticType:    lInterval.StaticType,
	}

	return result.New(intersectionInterval)
}

// calculateNumeralIntersectionQuantity calculates the intersection of two Quantity intervals
func (i *interpreter) calculateNumeralIntersectionQuantity(lStart, lEnd, rStart, rEnd result.Quantity, lInterval, rInterval result.Interval) (result.Value, error) {
	// Calculate intersection bounds
	intersectionStart := maxFloat64(lStart.Value, rStart.Value)
	intersectionEnd := minFloat64(lEnd.Value, rEnd.Value)

	// Check if there's no overlap
	if compareNumeral(intersectionStart, intersectionEnd) == leftAfterRight {
		return result.New(nil)
	}

	// Calculate inclusivity for intersection bounds
	// Start is inclusive if it matches an inclusive bound from either interval
	startInclusive := true
	if compareNumeral(intersectionStart, lStart.Value) == leftEqualRight {
		startInclusive = lInterval.LowInclusive
	}
	if compareNumeral(intersectionStart, rStart.Value) == leftEqualRight {
		startInclusive = startInclusive && rInterval.LowInclusive
	}

	// End is inclusive if it matches an inclusive bound from either interval
	endInclusive := true
	if compareNumeral(intersectionEnd, lEnd.Value) == leftEqualRight {
		endInclusive = lInterval.HighInclusive
	}
	if compareNumeral(intersectionEnd, rEnd.Value) == leftEqualRight {
		endInclusive = endInclusive && rInterval.HighInclusive
	}

	// Create result values with Quantity type
	startVal, err := result.New(result.Quantity{Value: intersectionStart, Unit: lStart.Unit})
	if err != nil {
		return result.Value{}, err
	}
	endVal, err := result.New(result.Quantity{Value: intersectionEnd, Unit: lStart.Unit})
	if err != nil {
		return result.Value{}, err
	}

	// Create the intersection interval
	intersectionInterval := result.Interval{
		Low:           startVal,
		High:          endVal,
		LowInclusive:  startInclusive,
		HighInclusive: endInclusive,
		StaticType:    lInterval.StaticType,
	}

	return result.New(intersectionInterval)
}

// evalIntersectIntervalDateTime handles intersection for date/time interval types
func (i *interpreter) evalIntersectIntervalDateTime(m model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	// Get start and end bounds for both intervals
	lStart, lEnd, err := startAndEnd(lObj, &i.evaluationTimestamp)
	if err != nil {
		return result.Value{}, err
	}
	rStart, rEnd, err := startAndEnd(rObj, &i.evaluationTimestamp)
	if err != nil {
		return result.Value{}, err
	}

	// If any bound is null, return null
	if result.IsNull(lStart) || result.IsNull(lEnd) || result.IsNull(rStart) || result.IsNull(rEnd) {
		return result.New(nil)
	}

	// Get interval metadata for result construction
	lInterval, _ := result.ToInterval(lObj)

	// Handle different date/time types
	switch lInterval.StaticType.PointType {
	case types.Date:
		return i.evalIntersectIntervalDate(lStart, lEnd, rStart, rEnd, lInterval)
	case types.DateTime:
		return i.evalIntersectIntervalDateTimeType(lStart, lEnd, rStart, rEnd, lInterval)
	case types.Time:
		return i.evalIntersectIntervalTime(lStart, lEnd, rStart, rEnd, lInterval)
	default:
		return result.Value{}, fmt.Errorf("internal error - unsupported date/time type in evalIntersectIntervalDateTime: %v", lInterval.StaticType.PointType)
	}
}

// evalIntersectIntervalDate handles intersection for Date intervals
func (i *interpreter) evalIntersectIntervalDate(lStart, lEnd, rStart, rEnd result.Value, lInterval result.Interval) (result.Value, error) {
	// Convert to DateTime for comparison but preserve Date type in result
	lStartDT, lEndDT, err := applyToValues(lStart, lEnd, result.ToDateTime)
	if err != nil {
		return result.Value{}, err
	}
	rStartDT, rEndDT, err := applyToValues(rStart, rEnd, result.ToDateTime)
	if err != nil {
		return result.Value{}, err
	}

	// Calculate intersection bounds using time.Time comparison
	var intersectionStartDT, intersectionEndDT result.DateTime
	var intersectionStart, intersectionEnd result.Value

	if lStartDT.Date.After(rStartDT.Date) {
		intersectionStartDT = lStartDT
		intersectionStart = lStart
	} else {
		intersectionStartDT = rStartDT
		intersectionStart = rStart
	}

	if lEndDT.Date.Before(rEndDT.Date) {
		intersectionEndDT = lEndDT
		intersectionEnd = lEnd
	} else {
		intersectionEndDT = rEndDT
		intersectionEnd = rEnd
	}

	// Check if there's no overlap
	if intersectionStartDT.Date.After(intersectionEndDT.Date) {
		return result.New(nil)
	}

	// Determine inclusivity for intersection bounds
	startInclusive := (intersectionStartDT.Date.Equal(lStartDT.Date) && lInterval.LowInclusive) ||
		(intersectionStartDT.Date.Equal(rStartDT.Date) && lInterval.LowInclusive) ||
		(intersectionStartDT.Date.After(lStartDT.Date) && intersectionStartDT.Date.Before(lEndDT.Date)) ||
		(intersectionStartDT.Date.After(rStartDT.Date) && intersectionStartDT.Date.Before(rEndDT.Date))

	endInclusive := (intersectionEndDT.Date.Equal(lEndDT.Date) && lInterval.HighInclusive) ||
		(intersectionEndDT.Date.Equal(rEndDT.Date) && lInterval.HighInclusive) ||
		(intersectionEndDT.Date.After(lStartDT.Date) && intersectionEndDT.Date.Before(lEndDT.Date)) ||
		(intersectionEndDT.Date.After(rStartDT.Date) && intersectionEndDT.Date.Before(rEndDT.Date))

	// Create the intersection interval with Date values
	intersectionInterval := result.Interval{
		Low:           intersectionStart,
		High:          intersectionEnd,
		LowInclusive:  startInclusive,
		HighInclusive: endInclusive,
		StaticType:    lInterval.StaticType,
	}

	return result.New(intersectionInterval)
}

// evalIntersectIntervalDateTimeType handles intersection for DateTime intervals
func (i *interpreter) evalIntersectIntervalDateTimeType(lStart, lEnd, rStart, rEnd result.Value, lInterval result.Interval) (result.Value, error) {
	// Convert to DateTime for comparison
	lStartDT, lEndDT, err := applyToValues(lStart, lEnd, result.ToDateTime)
	if err != nil {
		return result.Value{}, err
	}
	rStartDT, rEndDT, err := applyToValues(rStart, rEnd, result.ToDateTime)
	if err != nil {
		return result.Value{}, err
	}

	// Calculate intersection bounds using time.Time comparison
	var intersectionStart, intersectionEnd result.DateTime
	if lStartDT.Date.After(rStartDT.Date) {
		intersectionStart = lStartDT
	} else {
		intersectionStart = rStartDT
	}

	if lEndDT.Date.Before(rEndDT.Date) {
		intersectionEnd = lEndDT
	} else {
		intersectionEnd = rEndDT
	}

	// Check if there's no overlap
	if intersectionStart.Date.After(intersectionEnd.Date) {
		return result.New(nil)
	}

	// Determine inclusivity for intersection bounds
	startInclusive := (intersectionStart.Date.Equal(lStartDT.Date) && lInterval.LowInclusive) ||
		(intersectionStart.Date.Equal(rStartDT.Date) && lInterval.LowInclusive) ||
		(intersectionStart.Date.After(lStartDT.Date) && intersectionStart.Date.Before(lEndDT.Date)) ||
		(intersectionStart.Date.After(rStartDT.Date) && intersectionStart.Date.Before(rEndDT.Date))

	endInclusive := (intersectionEnd.Date.Equal(lEndDT.Date) && lInterval.HighInclusive) ||
		(intersectionEnd.Date.Equal(rEndDT.Date) && lInterval.HighInclusive) ||
		(intersectionEnd.Date.After(lStartDT.Date) && intersectionEnd.Date.Before(lEndDT.Date)) ||
		(intersectionEnd.Date.After(rStartDT.Date) && intersectionEnd.Date.Before(rEndDT.Date))

	// Create result values
	startVal, err := result.New(intersectionStart)
	if err != nil {
		return result.Value{}, err
	}
	endVal, err := result.New(intersectionEnd)
	if err != nil {
		return result.Value{}, err
	}

	// Create the intersection interval
	intersectionInterval := result.Interval{
		Low:           startVal,
		High:          endVal,
		LowInclusive:  startInclusive,
		HighInclusive: endInclusive,
		StaticType:    lInterval.StaticType,
	}

	return result.New(intersectionInterval)
}

// evalIntersectIntervalTime handles intersection for Time intervals
func (i *interpreter) evalIntersectIntervalTime(lStart, lEnd, rStart, rEnd result.Value, lInterval result.Interval) (result.Value, error) {
	// Convert to DateTime for comparison but preserve Time type in result
	lStartDT, lEndDT, err := applyToValues(lStart, lEnd, result.ToDateTime)
	if err != nil {
		return result.Value{}, err
	}
	rStartDT, rEndDT, err := applyToValues(rStart, rEnd, result.ToDateTime)
	if err != nil {
		return result.Value{}, err
	}

	// Calculate intersection bounds using time.Time comparison
	var intersectionStartDT, intersectionEndDT result.DateTime
	var intersectionStart, intersectionEnd result.Value

	if lStartDT.Date.After(rStartDT.Date) {
		intersectionStartDT = lStartDT
		intersectionStart = lStart
	} else {
		intersectionStartDT = rStartDT
		intersectionStart = rStart
	}

	if lEndDT.Date.Before(rEndDT.Date) {
		intersectionEndDT = lEndDT
		intersectionEnd = lEnd
	} else {
		intersectionEndDT = rEndDT
		intersectionEnd = rEnd
	}

	// Check if there's no overlap
	if intersectionStartDT.Date.After(intersectionEndDT.Date) {
		return result.New(nil)
	}

	// Determine inclusivity for intersection bounds
	startInclusive := (intersectionStartDT.Date.Equal(lStartDT.Date) && lInterval.LowInclusive) ||
		(intersectionStartDT.Date.Equal(rStartDT.Date) && lInterval.LowInclusive) ||
		(intersectionStartDT.Date.After(lStartDT.Date) && intersectionStartDT.Date.Before(lEndDT.Date)) ||
		(intersectionStartDT.Date.After(rStartDT.Date) && intersectionStartDT.Date.Before(rEndDT.Date))

	endInclusive := (intersectionEndDT.Date.Equal(lEndDT.Date) && lInterval.HighInclusive) ||
		(intersectionEndDT.Date.Equal(rEndDT.Date) && lInterval.HighInclusive) ||
		(intersectionEndDT.Date.After(lStartDT.Date) && intersectionEndDT.Date.Before(lEndDT.Date)) ||
		(intersectionEndDT.Date.After(rStartDT.Date) && intersectionEndDT.Date.Before(rEndDT.Date))

	// Create the intersection interval with Time values
	intersectionInterval := result.Interval{
		Low:           intersectionStart,
		High:          intersectionEnd,
		LowInclusive:  startInclusive,
		HighInclusive: endInclusive,
		StaticType:    lInterval.StaticType,
	}

	return result.New(intersectionInterval)
}


// Helper functions for min/max calculations
func maxInt32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

func minInt32(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func maxFloat64(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func minFloat64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// Type-specific intersect functions for dispatcher
func (i *interpreter) evalIntersectIntervalInteger(m model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	// Handle null inputs
	if result.IsNull(lObj) || result.IsNull(rObj) {
		return result.New(nil)
	}

	// Get start and end bounds for both intervals
	lStart, lEnd, err := startAndEnd(lObj, &i.evaluationTimestamp)
	if err != nil {
		return result.Value{}, err
	}
	rStart, rEnd, err := startAndEnd(rObj, &i.evaluationTimestamp)
	if err != nil {
		return result.Value{}, err
	}

	// If any bound is null, return null
	if result.IsNull(lStart) || result.IsNull(lEnd) || result.IsNull(rStart) || result.IsNull(rEnd) {
		return result.New(nil)
	}

	// Get interval metadata for result construction
	lInterval, _ := result.ToInterval(lObj)
	rInterval, _ := result.ToInterval(rObj)

	// Convert to int32 values
	lStartVal, lEndVal, err := applyToValues(lStart, lEnd, result.ToInt32)
	if err != nil {
		return result.Value{}, err
	}
	rStartVal, rEndVal, err := applyToValues(rStart, rEnd, result.ToInt32)
	if err != nil {
		return result.Value{}, err
	}

	return i.calculateNumeralIntersectionInt32(lStartVal, lEndVal, rStartVal, rEndVal, lInterval, rInterval)
}

func (i *interpreter) evalIntersectIntervalLong(m model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	// Handle null inputs
	if result.IsNull(lObj) || result.IsNull(rObj) {
		return result.New(nil)
	}

	// Get start and end bounds for both intervals
	lStart, lEnd, err := startAndEnd(lObj, &i.evaluationTimestamp)
	if err != nil {
		return result.Value{}, err
	}
	rStart, rEnd, err := startAndEnd(rObj, &i.evaluationTimestamp)
	if err != nil {
		return result.Value{}, err
	}

	// If any bound is null, return null
	if result.IsNull(lStart) || result.IsNull(lEnd) || result.IsNull(rStart) || result.IsNull(rEnd) {
		return result.New(nil)
	}

	// Get interval metadata for result construction
	lInterval, _ := result.ToInterval(lObj)
	rInterval, _ := result.ToInterval(rObj)

	// Convert to int64 values
	lStartVal, lEndVal, err := applyToValues(lStart, lEnd, result.ToInt64)
	if err != nil {
		return result.Value{}, err
	}
	rStartVal, rEndVal, err := applyToValues(rStart, rEnd, result.ToInt64)
	if err != nil {
		return result.Value{}, err
	}

	return i.calculateNumeralIntersectionInt64(lStartVal, lEndVal, rStartVal, rEndVal, lInterval, rInterval)
}

func (i *interpreter) evalIntersectIntervalDecimal(m model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	// Handle null inputs
	if result.IsNull(lObj) || result.IsNull(rObj) {
		return result.New(nil)
	}

	// Get start and end bounds for both intervals
	lStart, lEnd, err := startAndEnd(lObj, &i.evaluationTimestamp)
	if err != nil {
		return result.Value{}, err
	}
	rStart, rEnd, err := startAndEnd(rObj, &i.evaluationTimestamp)
	if err != nil {
		return result.Value{}, err
	}

	// If any bound is null, return null
	if result.IsNull(lStart) || result.IsNull(lEnd) || result.IsNull(rStart) || result.IsNull(rEnd) {
		return result.New(nil)
	}

	// Get interval metadata for result construction
	lInterval, _ := result.ToInterval(lObj)
	rInterval, _ := result.ToInterval(rObj)

	// Convert to float64 values
	lStartVal, lEndVal, err := applyToValues(lStart, lEnd, result.ToFloat64)
	if err != nil {
		return result.Value{}, err
	}
	rStartVal, rEndVal, err := applyToValues(rStart, rEnd, result.ToFloat64)
	if err != nil {
		return result.Value{}, err
	}

	return i.calculateNumeralIntersectionFloat64(lStartVal, lEndVal, rStartVal, rEndVal, lInterval, rInterval)
}

func (i *interpreter) evalIntersectIntervalQuantity(m model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	// Handle null inputs
	if result.IsNull(lObj) || result.IsNull(rObj) {
		return result.New(nil)
	}

	// Get start and end bounds for both intervals
	lStart, lEnd, err := startAndEnd(lObj, &i.evaluationTimestamp)
	if err != nil {
		return result.Value{}, err
	}
	rStart, rEnd, err := startAndEnd(rObj, &i.evaluationTimestamp)
	if err != nil {
		return result.Value{}, err
	}

	// If any bound is null, return null
	if result.IsNull(lStart) || result.IsNull(lEnd) || result.IsNull(rStart) || result.IsNull(rEnd) {
		return result.New(nil)
	}

	// Get interval metadata for result construction
	lInterval, _ := result.ToInterval(lObj)
	rInterval, _ := result.ToInterval(rObj)

	// Convert to Quantity values
	lStartVal, lEndVal, err := applyToValues(lStart, lEnd, result.ToQuantity)
	if err != nil {
		return result.Value{}, err
	}
	rStartVal, rEndVal, err := applyToValues(rStart, rEnd, result.ToQuantity)
	if err != nil {
		return result.Value{}, err
	}

	// Check unit compatibility
	if lStartVal.Unit != rStartVal.Unit {
		return result.Value{}, fmt.Errorf("intersect operator received Quantities with differing unit values, unit conversion is not currently supported, got: %v, %v", lStartVal.Unit, rStartVal.Unit)
	}

	return i.calculateNumeralIntersectionQuantity(lStartVal, lEndVal, rStartVal, rEndVal, lInterval, rInterval)
}
