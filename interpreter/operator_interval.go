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
	"sort"
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


// evalCollapseIntervalDate evaluates collapse for Date intervals
// collapse(argument List<Interval<Date>>) List<Interval<Date>>
// collapse(argument List<Interval<Date>>, per Quantity) List<Interval<Date>>
func (i *interpreter) evalCollapseIntervalDate(m model.INaryExpression, operands []result.Value) (result.Value, error) {
	if len(operands) == 0 {
		return result.Value{}, fmt.Errorf("internal error - Collapse must have at least one operand")
	}

	arg := operands[0]
	// For now, ignore the second argument (per parameter) if present
	// TODO: Implement precision control with the "per" parameter

	// Handle null input
	if result.IsNull(arg) {
		listType, ok := arg.RuntimeType().(*types.List)
		if !ok {
			return result.New(result.List{Value: []result.Value{}, StaticType: &types.List{ElementType: &types.Interval{PointType: types.Date}}})
		}
		return result.New(result.List{Value: []result.Value{}, StaticType: listType})
	}

	// Convert to list of intervals
	intervals, err := result.ToSlice(arg)
	if err != nil {
		return result.Value{}, err
	}

	// Get the type of interval elements
	listType, ok := arg.RuntimeType().(*types.List)
	if !ok {
		return result.Value{}, fmt.Errorf("internal error - evalCollapseIntervalDate got a non-list type")
	}

	if len(intervals) == 0 {
		return result.New(result.List{Value: []result.Value{}, StaticType: listType})
	}

	// Sort the intervals by their start time
	sortedIntervals, err := sortIntervals(intervals, i.evaluationTimestamp)
	if err != nil {
		return result.Value{}, err
	}

	// Use temporal collapse logic for Date intervals
	return collapseTemporalIntervals(sortedIntervals, i.evaluationTimestamp)
}

// evalCollapseIntervalDateTime evaluates collapse for DateTime intervals
// collapse(argument List<Interval<DateTime>>) List<Interval<DateTime>>
// collapse(argument List<Interval<DateTime>>, per Quantity) List<Interval<DateTime>>
func (i *interpreter) evalCollapseIntervalDateTime(m model.INaryExpression, operands []result.Value) (result.Value, error) {
	if len(operands) == 0 {
		return result.Value{}, fmt.Errorf("internal error - Collapse must have at least one operand")
	}

	arg := operands[0]
	// For now, ignore the second argument (per parameter) if present
	// TODO: Implement precision control with the "per" parameter

	// Handle null input
	if result.IsNull(arg) {
		listType, ok := arg.RuntimeType().(*types.List)
		if !ok {
			return result.New(result.List{Value: []result.Value{}, StaticType: &types.List{ElementType: &types.Interval{PointType: types.DateTime}}})
		}
		return result.New(result.List{Value: []result.Value{}, StaticType: listType})
	}

	// Convert to list of intervals
	intervals, err := result.ToSlice(arg)
	if err != nil {
		return result.Value{}, err
	}

	// Get the type of interval elements
	listType, ok := arg.RuntimeType().(*types.List)
	if !ok {
		return result.Value{}, fmt.Errorf("internal error - evalCollapseIntervalDateTime got a non-list type")
	}

	if len(intervals) == 0 {
		return result.New(result.List{Value: []result.Value{}, StaticType: listType})
	}

	// Sort the intervals by their start time
	sortedIntervals, err := sortIntervals(intervals, i.evaluationTimestamp)
	if err != nil {
		return result.Value{}, err
	}

	// Use temporal collapse logic for DateTime intervals
	return collapseTemporalIntervals(sortedIntervals, i.evaluationTimestamp)
}

// evalCollapseIntervalTime evaluates collapse for Time intervals
// collapse(argument List<Interval<Time>>) List<Interval<Time>>
// collapse(argument List<Interval<Time>>, per Quantity) List<Interval<Time>>
func (i *interpreter) evalCollapseIntervalTime(m model.INaryExpression, operands []result.Value) (result.Value, error) {
	if len(operands) == 0 {
		return result.Value{}, fmt.Errorf("internal error - Collapse must have at least one operand")
	}

	arg := operands[0]
	// For now, ignore the second argument (per parameter) if present
	// TODO: Implement precision control with the "per" parameter

	// Handle null input
	if result.IsNull(arg) {
		listType, ok := arg.RuntimeType().(*types.List)
		if !ok {
			return result.New(result.List{Value: []result.Value{}, StaticType: &types.List{ElementType: &types.Interval{PointType: types.Time}}})
		}
		return result.New(result.List{Value: []result.Value{}, StaticType: listType})
	}

	// Convert to list of intervals
	intervals, err := result.ToSlice(arg)
	if err != nil {
		return result.Value{}, err
	}

	// Get the type of interval elements
	listType, ok := arg.RuntimeType().(*types.List)
	if !ok {
		return result.Value{}, fmt.Errorf("internal error - evalCollapseIntervalTime got a non-list type")
	}

	if len(intervals) == 0 {
		return result.New(result.List{Value: []result.Value{}, StaticType: listType})
	}

	// Sort the intervals by their start time
	sortedIntervals, err := sortIntervals(intervals, i.evaluationTimestamp)
	if err != nil {
		return result.Value{}, err
	}

	// Use temporal collapse logic for Time intervals
	return collapseTemporalIntervals(sortedIntervals, i.evaluationTimestamp)
}

// evalCollapseIntervalInteger evaluates collapse for Integer intervals
// collapse(argument List<Interval<Integer>>) List<Interval<Integer>>
// collapse(argument List<Interval<Integer>>, per Quantity) List<Interval<Integer>>
func (i *interpreter) evalCollapseIntervalInteger(m model.INaryExpression, operands []result.Value) (result.Value, error) {
	if len(operands) == 0 {
		return result.Value{}, fmt.Errorf("internal error - Collapse must have at least one operand")
	}

	arg := operands[0]
	// For now, ignore the second argument (per parameter) if present
	// TODO: Implement precision control with the "per" parameter

	// Handle null input
	if result.IsNull(arg) {
		listType, ok := arg.RuntimeType().(*types.List)
		if !ok {
			return result.New(result.List{Value: []result.Value{}, StaticType: &types.List{ElementType: &types.Interval{PointType: types.Integer}}})
		}
		return result.New(result.List{Value: []result.Value{}, StaticType: listType})
	}

	// Convert to list of intervals
	intervals, err := result.ToSlice(arg)
	if err != nil {
		return result.Value{}, err
	}

	// Get the type of interval elements
	listType, ok := arg.RuntimeType().(*types.List)
	if !ok {
		return result.Value{}, fmt.Errorf("internal error - evalCollapseIntervalInteger got a non-list type")
	}

	if len(intervals) == 0 {
		return result.New(result.List{Value: []result.Value{}, StaticType: listType})
	}

	// Sort the intervals by their start time
	sortedIntervals, err := sortIntervals(intervals, i.evaluationTimestamp)
	if err != nil {
		return result.Value{}, err
	}

	// Use numeric collapse logic for Integer intervals
	return collapseNumericIntervals(sortedIntervals, i.evaluationTimestamp)
}

// evalCollapseIntervalDecimal evaluates collapse for Decimal intervals
// collapse(argument List<Interval<Decimal>>) List<Interval<Decimal>>
// collapse(argument List<Interval<Decimal>>, per Quantity) List<Interval<Decimal>>
func (i *interpreter) evalCollapseIntervalDecimal(m model.INaryExpression, operands []result.Value) (result.Value, error) {
	if len(operands) == 0 {
		return result.Value{}, fmt.Errorf("internal error - Collapse must have at least one operand")
	}

	arg := operands[0]
	// For now, ignore the second argument (per parameter) if present
	// TODO: Implement precision control with the "per" parameter

	// Handle null input
	if result.IsNull(arg) {
		listType, ok := arg.RuntimeType().(*types.List)
		if !ok {
			return result.New(result.List{Value: []result.Value{}, StaticType: &types.List{ElementType: &types.Interval{PointType: types.Decimal}}})
		}
		return result.New(result.List{Value: []result.Value{}, StaticType: listType})
	}

	// Convert to list of intervals
	intervals, err := result.ToSlice(arg)
	if err != nil {
		return result.Value{}, err
	}

	// Get the type of interval elements
	listType, ok := arg.RuntimeType().(*types.List)
	if !ok {
		return result.Value{}, fmt.Errorf("internal error - evalCollapseIntervalDecimal got a non-list type")
	}

	if len(intervals) == 0 {
		return result.New(result.List{Value: []result.Value{}, StaticType: listType})
	}

	// Sort the intervals by their start time
	sortedIntervals, err := sortIntervals(intervals, i.evaluationTimestamp)
	if err != nil {
		return result.Value{}, err
	}

	// Use numeric collapse logic for Decimal intervals
	return collapseNumericIntervals(sortedIntervals, i.evaluationTimestamp)
}

// evalCollapseIntervalQuantity evaluates collapse for Quantity intervals
// collapse(argument List<Interval<Quantity>>) List<Interval<Quantity>>
// collapse(argument List<Interval<Quantity>>, per Quantity) List<Interval<Quantity>>
func (i *interpreter) evalCollapseIntervalQuantity(m model.INaryExpression, operands []result.Value) (result.Value, error) {
	if len(operands) == 0 {
		return result.Value{}, fmt.Errorf("internal error - Collapse must have at least one operand")
	}

	arg := operands[0]
	// For now, ignore the second argument (per parameter) if present
	// TODO: Implement precision control with the "per" parameter

	// Handle null input
	if result.IsNull(arg) {
		listType, ok := arg.RuntimeType().(*types.List)
		if !ok {
			return result.New(result.List{Value: []result.Value{}, StaticType: &types.List{ElementType: &types.Interval{PointType: types.Quantity}}})
		}
		return result.New(result.List{Value: []result.Value{}, StaticType: listType})
	}

	// Convert to list of intervals
	intervals, err := result.ToSlice(arg)
	if err != nil {
		return result.Value{}, err
	}

	// Get the type of interval elements
	listType, ok := arg.RuntimeType().(*types.List)
	if !ok {
		return result.Value{}, fmt.Errorf("internal error - evalCollapseIntervalQuantity got a non-list type")
	}

	if len(intervals) == 0 {
		return result.New(result.List{Value: []result.Value{}, StaticType: listType})
	}

	// Sort quantity intervals by start value
	evalTimestamp := i.evaluationTimestamp
	sort.Slice(intervals, func(idx1, idx2 int) bool {
		start1, err1 := start(intervals[idx1], &evalTimestamp)
		start2, err2 := start(intervals[idx2], &evalTimestamp)
		
		// Handle errors or nulls - put them at the beginning
		if err1 != nil || result.IsNull(start1) {
			return true
		}
		if err2 != nil || result.IsNull(start2) {
			return false
		}
		
		// Convert to Quantity and compare
		if qty1, err1 := result.ToQuantity(start1); err1 == nil {
			if qty2, err2 := result.ToQuantity(start2); err2 == nil {
				// Check units match
				if qty1.Unit != qty2.Unit {
					// If units don't match, preserve original order
					return idx1 < idx2
				}
				return qty1.Value < qty2.Value
			}
		}
		
		// Fallback: preserve original order
		return idx1 < idx2
	})

	// Remove null intervals first
	var nonNullIntervals []result.Value
	for _, interval := range intervals {
		if !result.IsNull(interval) {
			nonNullIntervals = append(nonNullIntervals, interval)
		}
	}

	if len(nonNullIntervals) == 0 {
		return result.New(result.List{Value: []result.Value{}, StaticType: listType})
	}

	// Use a simple single-pass algorithm that processes intervals in sorted order
	var collapsedIntervals []result.Value

	// Start with the first interval
	currentInterval := nonNullIntervals[0]

	for idx := 1; idx < len(nonNullIntervals); idx++ {
		nextInterval := nonNullIntervals[idx]

		// Check if current and next intervals can be merged
		canMerge, err := canMergeQuantityIntervals(currentInterval, nextInterval, i.evaluationTimestamp)
		if err != nil {
			return result.Value{}, err
		}

		if canMerge {
			// Merge the intervals
			merged, err := mergeQuantityIntervals(currentInterval, nextInterval, i.evaluationTimestamp)
			if err != nil {
				return result.Value{}, err
			}
			currentInterval = merged
		} else {
			// Can't merge, add current to result and move to next
			cleanCurrentInterval, err := createCleanInterval(currentInterval)
			if err != nil {
				return result.Value{}, err
			}
			collapsedIntervals = append(collapsedIntervals, cleanCurrentInterval)
			currentInterval = nextInterval
		}
	}

	// Add the last interval
	cleanLastInterval, err := createCleanInterval(currentInterval)
	if err != nil {
		return result.Value{}, err
	}
	collapsedIntervals = append(collapsedIntervals, cleanLastInterval)

	// Return the collapsed intervals with the correct list type
	return result.New(result.List{Value: collapsedIntervals, StaticType: listType})
}

// sortIntervals sorts a list of intervals by their start time
func sortIntervals(intervals []result.Value, evaluationTimestamp time.Time) ([]result.Value, error) {
	type sortableInterval struct {
		interval result.Value
		start    result.Value
	}

	var sortable []sortableInterval
	for _, interval := range intervals {
		s, err := start(interval, &evaluationTimestamp)
		if err != nil {
			return nil, err
		}
		sortable = append(sortable, sortableInterval{
			interval: interval,
			start:    s,
		})
	}

	// Sort by start time
	sort.Slice(sortable, func(i, j int) bool {
		// Handle null start times (minimum value)
		if result.IsNull(sortable[i].start) {
			return true
		}
		if result.IsNull(sortable[j].start) {
			return false
		}

		// Try numeric comparison
		leftFloat, leftErr := result.ToFloat64(sortable[i].start)
		rightFloat, rightErr := result.ToFloat64(sortable[j].start)
		if leftErr == nil && rightErr == nil {
			return leftFloat < rightFloat
		}

		// Try datetime comparison
		leftDT, leftErr := result.ToDateTime(sortable[i].start)
		rightDT, rightErr := result.ToDateTime(sortable[j].start)
		if leftErr == nil && rightErr == nil {
			comp, err := compareDateTimeWithPrecision(leftDT, rightDT, "")
			if err == nil && comp != insufficientPrecision {
				return comp == leftBeforeRight
			}
		}

		// If we can't compare, preserve original order
		return i < j
	})

	// Extract sorted intervals
	var result []result.Value
	for _, si := range sortable {
		result = append(result, si.interval)
	}
	return result, nil
}

// collapseTemporalIntervals collapses a list of temporal intervals (Date or DateTime)
func collapseTemporalIntervals(intervals []result.Value, evaluationTimestamp time.Time) (result.Value, error) {
	if len(intervals) == 0 {
		return result.New([]result.Value{})
	}

	// Remove null intervals first
	var nonNullIntervals []result.Value
	for _, interval := range intervals {
		if !result.IsNull(interval) {
			nonNullIntervals = append(nonNullIntervals, interval)
		}
	}

	if len(nonNullIntervals) == 0 {
		return result.New([]result.Value{})
	}

	// Sort temporal intervals by start time
	sort.Slice(nonNullIntervals, func(i, j int) bool {
		start1, err1 := start(nonNullIntervals[i], &evaluationTimestamp)
		start2, err2 := start(nonNullIntervals[j], &evaluationTimestamp)

		// Handle errors or nulls - put them at the beginning
		if err1 != nil || result.IsNull(start1) {
			return true
		}
		if err2 != nil || result.IsNull(start2) {
			return false
		}

		// Convert to DateTime and compare
		if dt1, err1 := result.ToDateTime(start1); err1 == nil {
			if dt2, err2 := result.ToDateTime(start2); err2 == nil {
				comp, err := compareDateTimeWithPrecision(dt1, dt2, "")
				if err == nil && comp != insufficientPrecision {
					return comp == leftBeforeRight
				}
			}
		}

		// Fallback: preserve original order
		return i < j
	})

	var collapsedIntervals []result.Value
	currentInterval := nonNullIntervals[0]
	currentEnd, err := end(currentInterval, &evaluationTimestamp)
	if err != nil {
		return result.Value{}, err
	}

	for i := 1; i < len(nonNullIntervals); i++ {
		nextInterval := nonNullIntervals[i]
		nextStart, err := start(nextInterval, &evaluationTimestamp)
		if err != nil {
			return result.Value{}, err
		}
		nextEnd, err := end(nextInterval, &evaluationTimestamp)
		if err != nil {
			return result.Value{}, err
		}

		// If either end point is null, skip comparison
		if result.IsNull(currentEnd) || result.IsNull(nextStart) {
			continue
		}

		// Convert to DateTime for comparison
		currentEndDT, err := result.ToDateTime(currentEnd)
		if err != nil {
			return result.Value{}, err
		}
		nextStartDT, err := result.ToDateTime(nextStart)
		if err != nil {
			return result.Value{}, err
		}

		// Compare with precision
		comparison, err := compareDateTimeWithPrecision(currentEndDT, nextStartDT, "")
		if err != nil {
			return result.Value{}, err
		}

		// Check if intervals are adjacent (for Date intervals, check if end + 1 day = start)
		var adjacent bool = false
		if comparison == leftBeforeRight {
			// Check if they are adjacent (consecutive days for Date intervals)
			if currentEnd.RuntimeType() == types.Date {
				// For dates, check if currentEnd + 1 day = nextStart
				currentEndDate, err := result.ToDateTime(currentEnd)
				if err == nil {
					nextDay := currentEndDate.Date.AddDate(0, 0, 1)
					nextStartDate, err := result.ToDateTime(nextStart)
					if err == nil && nextDay.Equal(nextStartDate.Date) {
						adjacent = true
					}
				}
			}
		}

		// If there's a gap between intervals (and they're not adjacent), start a new interval
		if comparison == leftBeforeRight && !adjacent {
			// Create a clean version of the current interval without source metadata
			cleanCurrentInterval, err := createCleanInterval(currentInterval)
			if err != nil {
				return result.Value{}, err
			}
			collapsedIntervals = append(collapsedIntervals, cleanCurrentInterval)
			currentInterval = nextInterval
			currentEnd = nextEnd
		} else {
			// Intervals overlap or are adjacent, merge them
			currentIntervalValue, err := result.ToInterval(currentInterval)
			if err != nil {
				return result.Value{}, err
			}
			nextIntervalValue, err := result.ToInterval(nextInterval)
			if err != nil {
				return result.Value{}, err
			}

			// Get the earliest start and latest end
			currentStart, err := start(currentInterval, &evaluationTimestamp)
			if err != nil {
				return result.Value{}, err
			}

			var mergedLow, mergedHigh result.Value
			var mergedLowInclusive, mergedHighInclusive bool

			// Determine the lower bound
			if result.IsNull(currentStart) || result.IsNull(nextStart) {
				if result.IsNull(currentStart) {
					mergedLow = nextStart
					mergedLowInclusive = nextIntervalValue.LowInclusive
				} else if result.IsNull(nextStart) {
					mergedLow = currentStart
					mergedLowInclusive = currentIntervalValue.LowInclusive
				}
			} else {
				// Compare the two start times
				currentStartDT, err := result.ToDateTime(currentStart)
				if err != nil {
					return result.Value{}, err
				}
				nextStartDT, err := result.ToDateTime(nextStart)
				if err != nil {
					return result.Value{}, err
				}

				comparison, err := compareDateTimeWithPrecision(currentStartDT, nextStartDT, "")
				if err != nil {
					return result.Value{}, err
				}

				if comparison == leftBeforeRight || comparison == leftEqualRight {
					mergedLow = currentStart
					mergedLowInclusive = currentIntervalValue.LowInclusive
				} else {
					mergedLow = nextStart
					mergedLowInclusive = nextIntervalValue.LowInclusive
				}
			}

			// Determine the upper bound
			if result.IsNull(currentEnd) || result.IsNull(nextEnd) {
				if result.IsNull(currentEnd) {
					mergedHigh = nextEnd
					mergedHighInclusive = nextIntervalValue.HighInclusive
				} else if result.IsNull(nextEnd) {
					mergedHigh = currentEnd
					mergedHighInclusive = currentIntervalValue.HighInclusive
				}
			} else {
				// Compare the two end times
				currentEndDT, err := result.ToDateTime(currentEnd)
				if err != nil {
					return result.Value{}, err
				}
				nextEndDT, err := result.ToDateTime(nextEnd)
				if err != nil {
					return result.Value{}, err
				}

				comparison, err := compareDateTimeWithPrecision(currentEndDT, nextEndDT, "")
				if err != nil {
					return result.Value{}, err
				}

				if comparison == leftAfterRight || comparison == leftEqualRight {
					mergedHigh = currentEnd
					mergedHighInclusive = currentIntervalValue.HighInclusive
				} else {
					mergedHigh = nextEnd
					mergedHighInclusive = nextIntervalValue.HighInclusive
				}
			}

			// Create clean values without source metadata
			cleanLow, err := createCleanValue(mergedLow)
			if err != nil {
				return result.Value{}, err
			}
			cleanHigh, err := createCleanValue(mergedHigh)
			if err != nil {
				return result.Value{}, err
			}

			// Create the merged interval without source metadata
			mergedInterval := result.Interval{
				Low:           cleanLow,
				High:          cleanHigh,
				LowInclusive:  mergedLowInclusive,
				HighInclusive: mergedHighInclusive,
				StaticType:    currentIntervalValue.StaticType,
			}

			mergedIntervalValue, err := result.New(mergedInterval)
			if err != nil {
				return result.Value{}, err
			}

			currentInterval = mergedIntervalValue
			currentEnd = mergedHigh
		}
	}

	// Add the last interval (clean it too)
	cleanLastInterval, err := createCleanInterval(currentInterval)
	if err != nil {
		return result.Value{}, err
	}
	collapsedIntervals = append(collapsedIntervals, cleanLastInterval)

	// Return the collapsed intervals with the correct list type
	if len(intervals) > 0 {
		if firstInterval, err := result.ToInterval(intervals[0]); err == nil {
			return result.New(result.List{Value: collapsedIntervals, StaticType: &types.List{ElementType: &types.Interval{PointType: firstInterval.StaticType.PointType}}})
		}
	}
	return result.New(result.List{Value: collapsedIntervals, StaticType: &types.List{ElementType: &types.Interval{PointType: types.Any}}})
}


// createCleanInterval creates a new interval result.Value without source metadata
func createCleanInterval(intervalVal result.Value) (result.Value, error) {
	if result.IsNull(intervalVal) {
		return intervalVal, nil
	}

	interval, err := result.ToInterval(intervalVal)
	if err != nil {
		return result.Value{}, err
	}

	// Create clean low and high values
	cleanLow, err := createCleanValue(interval.Low)
	if err != nil {
		return result.Value{}, err
	}
	cleanHigh, err := createCleanValue(interval.High)
	if err != nil {
		return result.Value{}, err
	}

	// Create a new clean interval
	cleanInterval := result.Interval{
		Low:           cleanLow,
		High:          cleanHigh,
		LowInclusive:  interval.LowInclusive,
		HighInclusive: interval.HighInclusive,
		StaticType:    interval.StaticType,
	}

	return result.New(cleanInterval)
}

// createCleanValue creates a new result.Value without source metadata
func createCleanValue(val result.Value) (result.Value, error) {
	if result.IsNull(val) {
		return val, nil
	}

	// Extract the core Go value and create a new result.Value without source metadata
	switch val.RuntimeType() {
	case types.Date:
		date, err := result.ToDateTime(val)
		if err != nil {
			return result.Value{}, err
		}
		// Create a new date in UTC to match test expectations
		utcDate := time.Date(date.Date.Year(), date.Date.Month(), date.Date.Day(), 0, 0, 0, 0, time.UTC)
		return result.New(result.Date{Date: utcDate, Precision: date.Precision})
	case types.DateTime:
		dateTime, err := result.ToDateTime(val)
		if err != nil {
			return result.Value{}, err
		}
		return result.New(dateTime)
	case types.Time:
		time, err := result.ToTime(val)
		if err != nil {
			return result.Value{}, err
		}
		return result.New(time)
	case types.Integer:
		intVal, err := result.ToInt32(val)
		if err != nil {
			return result.Value{}, err
		}
		return result.New(intVal)
	case types.Long:
		longVal, err := result.ToInt64(val)
		if err != nil {
			return result.Value{}, err
		}
		return result.New(longVal)
	case types.Decimal:
		decimalVal, err := result.ToFloat64(val)
		if err != nil {
			return result.Value{}, err
		}
		return result.New(decimalVal)
	case types.Quantity:
		quantityVal, err := result.ToQuantity(val)
		if err != nil {
			return result.Value{}, err
		}
		return result.New(quantityVal)
	default:
		// For other types, just return the original value
		return val, nil
	}
}

// collapseNumericIntervals collapses a list of numeric intervals (Integer, Decimal, Quantity)
func collapseNumericIntervals(intervals []result.Value, evaluationTimestamp time.Time) (result.Value, error) {
	if len(intervals) == 0 {
		return result.New([]result.Value{})
	}

	// Remove null intervals first
	var nonNullIntervals []result.Value
	for _, interval := range intervals {
		if !result.IsNull(interval) {
			nonNullIntervals = append(nonNullIntervals, interval)
		}
	}

	if len(nonNullIntervals) == 0 {
		return result.New([]result.Value{})
	}

	// Sort numeric intervals by start value
	sort.Slice(nonNullIntervals, func(i, j int) bool {
		start1, err1 := start(nonNullIntervals[i], &evaluationTimestamp)
		start2, err2 := start(nonNullIntervals[j], &evaluationTimestamp)

		// Handle errors or nulls - put them at the beginning
		if err1 != nil || result.IsNull(start1) {
			return true
		}
		if err2 != nil || result.IsNull(start2) {
			return false
		}

		// Try integer comparison first
		if int1, err1 := result.ToInt32(start1); err1 == nil {
			if int2, err2 := result.ToInt32(start2); err2 == nil {
				return int1 < int2
			}
		}

		// Try float comparison for decimals/quantities
		if float1, err1 := result.ToFloat64(start1); err1 == nil {
			if float2, err2 := result.ToFloat64(start2); err2 == nil {
				return float1 < float2
			}
		}

		// Fallback: preserve original order
		return i < j
	})

	// Use a simple single-pass algorithm that processes intervals in sorted order
	var collapsedIntervals []result.Value

	// Start with the first interval
	currentInterval := nonNullIntervals[0]

	for i := 1; i < len(nonNullIntervals); i++ {
		nextInterval := nonNullIntervals[i]

		// Check if current and next intervals can be merged
		canMerge, err := canMergeNumericIntervals(currentInterval, nextInterval, evaluationTimestamp)
		if err != nil {
			return result.Value{}, err
		}

		if canMerge {
			// Merge the intervals
			merged, err := mergeNumericIntervals(currentInterval, nextInterval, evaluationTimestamp)
			if err != nil {
				return result.Value{}, err
			}
			currentInterval = merged
		} else {
			// Can't merge, add current to result and move to next
			cleanCurrentInterval, err := createCleanInterval(currentInterval)
			if err != nil {
				return result.Value{}, err
			}
			collapsedIntervals = append(collapsedIntervals, cleanCurrentInterval)
			currentInterval = nextInterval
		}
	}

	// Add the last interval
	cleanLastInterval, err := createCleanInterval(currentInterval)
	if err != nil {
		return result.Value{}, err
	}
	collapsedIntervals = append(collapsedIntervals, cleanLastInterval)

	// Return the collapsed intervals with the correct list type
	if len(intervals) > 0 {
		if firstInterval, err := result.ToInterval(intervals[0]); err == nil {
			return result.New(result.List{Value: collapsedIntervals, StaticType: &types.List{ElementType: &types.Interval{PointType: firstInterval.StaticType.PointType}}})
		}
	}
	return result.New(result.List{Value: collapsedIntervals, StaticType: &types.List{ElementType: &types.Interval{PointType: types.Any}}})
}

// canMergeNumericIntervals checks if two numeric intervals can be merged (overlap or are adjacent)
func canMergeNumericIntervals(interval1, interval2 result.Value, evaluationTimestamp time.Time) (bool, error) {
	// Get bounds of both intervals
	start1, end1, err := startAndEnd(interval1, &evaluationTimestamp)
	if err != nil {
		return false, err
	}
	start2, end2, err := startAndEnd(interval2, &evaluationTimestamp)
	if err != nil {
		return false, err
	}

	// Handle null bounds
	if result.IsNull(start1) || result.IsNull(end1) || result.IsNull(start2) || result.IsNull(end2) {
		return false, nil
	}

	// Check if intervals overlap or are adjacent
	// For sorted intervals, we only need to check if interval1.end >= interval2.start (with tolerance)

	// Try integer comparison first
	end1Int, end1IntErr := result.ToInt32(end1)
	start2Int, start2IntErr := result.ToInt32(start2)

	if end1IntErr == nil && start2IntErr == nil {
		// Integer intervals: overlapping if end1 >= start2, adjacent if end1 + 1 >= start2
		// This covers both cases: [1,5] and [3,7] (overlapping), [1,5] and [6,10] (adjacent)
		return end1Int+1 >= start2Int, nil
	}

	// For all other numeric types, convert to float64 and use tolerance
	end1Float, end1FloatErr := result.ToFloat64(end1)
	start2Float, start2FloatErr := result.ToFloat64(start2)

	if end1FloatErr == nil && start2FloatErr == nil {
		// Decimal/Long intervals: overlapping or adjacent with small tolerance for precision issues
		const tolerance = 1e-8
		return end1Float >= start2Float-tolerance, nil
	}

	// Try quantity comparison
	end1Qty, end1QtyErr := result.ToQuantity(end1)
	start2Qty, start2QtyErr := result.ToQuantity(start2)

	if end1QtyErr == nil && start2QtyErr == nil {
		// Check units match
		if end1Qty.Unit != start2Qty.Unit {
			return false, fmt.Errorf("cannot merge intervals with different units: %v vs %v", end1Qty.Unit, start2Qty.Unit)
		}
		// Quantity intervals: overlapping if end1 >= start2, adjacent if end1 + 1 >= start2
		// For quantities, we treat them like decimals with a small tolerance
		const tolerance = 1e-8
		canMerge := end1Qty.Value >= start2Qty.Value-tolerance

		return canMerge, nil
	}

	return false, nil
}

// mergeNumericIntervals merges two numeric intervals into one
func mergeNumericIntervals(interval1, interval2 result.Value, evaluationTimestamp time.Time) (result.Value, error) {
	interval1Val, err := result.ToInterval(interval1)
	if err != nil {
		return result.Value{}, err
	}
	interval2Val, err := result.ToInterval(interval2)
	if err != nil {
		return result.Value{}, err
	}

	// Get bounds of both intervals
	start1, end1, err := startAndEnd(interval1, &evaluationTimestamp)
	if err != nil {
		return result.Value{}, err
	}
	start2, end2, err := startAndEnd(interval2, &evaluationTimestamp)
	if err != nil {
		return result.Value{}, err
	}

	// Determine the merged bounds
	var mergedLow, mergedHigh result.Value
	var mergedLowInclusive, mergedHighInclusive bool

	// Find the minimum start
	if result.IsNull(start1) {
		mergedLow = start2
		mergedLowInclusive = interval2Val.LowInclusive
	} else if result.IsNull(start2) {
		mergedLow = start1
		mergedLowInclusive = interval1Val.LowInclusive
	} else {
		// Compare starts
		start1Float, start1FloatErr := result.ToFloat64(start1)
		start2Float, start2FloatErr := result.ToFloat64(start2)

		if start1FloatErr == nil && start2FloatErr == nil {
			if start1Float <= start2Float {
				mergedLow = start1
				mergedLowInclusive = interval1Val.LowInclusive
			} else {
				mergedLow = start2
				mergedLowInclusive = interval2Val.LowInclusive
			}
		} else {
			// Try integer comparison
			start1Int, start1IntErr := result.ToInt32(start1)
			start2Int, start2IntErr := result.ToInt32(start2)

			if start1IntErr == nil && start2IntErr == nil {
				if start1Int <= start2Int {
					mergedLow = start1
					mergedLowInclusive = interval1Val.LowInclusive
				} else {
					mergedLow = start2
					mergedLowInclusive = interval2Val.LowInclusive
				}
			} else {
				// Default to first interval's start
				mergedLow = start1
				mergedLowInclusive = interval1Val.LowInclusive
			}
		}
	}

	// Find the maximum end
	if result.IsNull(end1) {
		mergedHigh = end2
		mergedHighInclusive = interval2Val.HighInclusive
	} else if result.IsNull(end2) {
		mergedHigh = end1
		mergedHighInclusive = interval1Val.HighInclusive
	} else {
		// Compare ends
		end1Float, end1FloatErr := result.ToFloat64(end1)
		end2Float, end2FloatErr := result.ToFloat64(end2)

		if end1FloatErr == nil && end2FloatErr == nil {
			if end1Float >= end2Float {
				mergedHigh = end1
				mergedHighInclusive = interval1Val.HighInclusive
			} else {
				mergedHigh = end2
				mergedHighInclusive = interval2Val.HighInclusive
			}
		} else {
			// Try integer comparison
			end1Int, end1IntErr := result.ToInt32(end1)
			end2Int, end2IntErr := result.ToInt32(end2)

			if end1IntErr == nil && end2IntErr == nil {
				if end1Int >= end2Int {
					mergedHigh = end1
					mergedHighInclusive = interval1Val.HighInclusive
				} else {
					mergedHigh = end2
					mergedHighInclusive = interval2Val.HighInclusive
				}
			} else {
				// Default to first interval's end
				mergedHigh = end1
				mergedHighInclusive = interval1Val.HighInclusive
			}
		}
	}

	// Create clean values without source metadata
	cleanLow, err := createCleanValue(mergedLow)
	if err != nil {
		return result.Value{}, err
	}
	cleanHigh, err := createCleanValue(mergedHigh)
	if err != nil {
		return result.Value{}, err
	}

	// Create the merged interval
	mergedInterval := result.Interval{
		Low:           cleanLow,
		High:          cleanHigh,
		LowInclusive:  mergedLowInclusive,
		HighInclusive: mergedHighInclusive,
		StaticType:    interval1Val.StaticType,
	}

	return result.New(mergedInterval)
}

// canMergeQuantityIntervals checks if two quantity intervals can be merged (overlap or are adjacent)
func canMergeQuantityIntervals(interval1, interval2 result.Value, evaluationTimestamp time.Time) (bool, error) {
	// Get bounds of both intervals
	start1, end1, err := startAndEnd(interval1, &evaluationTimestamp)
	if err != nil {
		return false, err
	}
	start2, end2, err := startAndEnd(interval2, &evaluationTimestamp)
	if err != nil {
		return false, err
	}

	// Handle null bounds
	if result.IsNull(start1) || result.IsNull(end1) || result.IsNull(start2) || result.IsNull(end2) {
		return false, nil
	}

	// Convert to quantities
	end1Qty, err := result.ToQuantity(end1)
	if err != nil {
		return false, err
	}
	start2Qty, err := result.ToQuantity(start2)
	if err != nil {
		return false, err
	}

	// Check units match
	if end1Qty.Unit != start2Qty.Unit {
		return false, fmt.Errorf("cannot merge intervals with different units: %v vs %v", end1Qty.Unit, start2Qty.Unit)
	}

	// For quantities, check if overlapping or adjacent
	// Adjacent means end1 + 1 >= start2, overlapping means end1 >= start2
	canMerge := end1Qty.Value+1 >= start2Qty.Value

	return canMerge, nil
}

// mergeQuantityIntervals merges two quantity intervals into one
func mergeQuantityIntervals(interval1, interval2 result.Value, evaluationTimestamp time.Time) (result.Value, error) {
	interval1Val, err := result.ToInterval(interval1)
	if err != nil {
		return result.Value{}, err
	}
	interval2Val, err := result.ToInterval(interval2)
	if err != nil {
		return result.Value{}, err
	}

	// Get bounds of both intervals
	start1, end1, err := startAndEnd(interval1, &evaluationTimestamp)
	if err != nil {
		return result.Value{}, err
	}
	start2, end2, err := startAndEnd(interval2, &evaluationTimestamp)
	if err != nil {
		return result.Value{}, err
	}

	// Determine the merged bounds
	var mergedLow, mergedHigh result.Value
	var mergedLowInclusive, mergedHighInclusive bool

	// Find the minimum start
	if result.IsNull(start1) {
		mergedLow = start2
		mergedLowInclusive = interval2Val.LowInclusive
	} else if result.IsNull(start2) {
		mergedLow = start1
		mergedLowInclusive = interval1Val.LowInclusive
	} else {
		// Compare starts using quantities
		start1Qty, err := result.ToQuantity(start1)
		if err != nil {
			return result.Value{}, err
		}
		start2Qty, err := result.ToQuantity(start2)
		if err != nil {
			return result.Value{}, err
		}

		if start1Qty.Value <= start2Qty.Value {
			mergedLow = start1
			mergedLowInclusive = interval1Val.LowInclusive
		} else {
			mergedLow = start2
			mergedLowInclusive = interval2Val.LowInclusive
		}
	}

	// Find the maximum end
	if result.IsNull(end1) {
		mergedHigh = end2
		mergedHighInclusive = interval2Val.HighInclusive
	} else if result.IsNull(end2) {
		mergedHigh = end1
		mergedHighInclusive = interval1Val.HighInclusive
	} else {
		// Compare ends using quantities
		end1Qty, err := result.ToQuantity(end1)
		if err != nil {
			return result.Value{}, err
		}
		end2Qty, err := result.ToQuantity(end2)
		if err != nil {
			return result.Value{}, err
		}

		if end1Qty.Value >= end2Qty.Value {
			mergedHigh = end1
			mergedHighInclusive = interval1Val.HighInclusive
		} else {
			mergedHigh = end2
			mergedHighInclusive = interval2Val.HighInclusive
		}
	}

	// Create clean values without source metadata
	cleanLow, err := createCleanValue(mergedLow)
	if err != nil {
		return result.Value{}, err
	}
	cleanHigh, err := createCleanValue(mergedHigh)
	if err != nil {
		return result.Value{}, err
	}

	// Create the merged interval
	mergedInterval := result.Interval{
		Low:           cleanLow,
		High:          cleanHigh,
		LowInclusive:  mergedLowInclusive,
		HighInclusive: mergedHighInclusive,
		StaticType:    interval1Val.StaticType,
	}

	return result.New(mergedInterval)
}
