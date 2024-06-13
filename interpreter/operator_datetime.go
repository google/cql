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
	"cmp"
	"errors"
	"fmt"
	"time"

	"github.com/google/cql/model"
	"github.com/google/cql/result"
	"github.com/google/cql/types"
)

// DATETIME OPERATORS - https://cql.hl7.org/09-b-cqlreference.html#datetime-operators-2

// op(left Date, right Date) Boolean
// https://cql.hl7.org/09-b-cqlreference.html#after
// https://cql.hl7.org/09-b-cqlreference.html#before
// https://cql.hl7.org/09-b-cqlreference.html#same-or-after-1
// https://cql.hl7.org/09-b-cqlreference.html#same-or-before-1
func evalCompareDateWithPrecision(b model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) || result.IsNull(rObj) {
		return result.New(nil)
	}

	p, err := precisionFromBinaryExpression(b)
	if err != nil {
		return result.Value{}, err
	}
	allowUnsetPrec := true
	if err := validateDatePrecision(p, allowUnsetPrec); err != nil {
		return result.Value{}, err
	}
	l, r, err := applyToValues(lObj, rObj, result.ToDateTime)
	if err != nil {
		return result.Value{}, err
	}

	switch b.(type) {
	case *model.After:
		return afterDateTimeWithPrecision(l, r, p)
	case *model.Before:
		return beforeDateTimeWithPrecision(l, r, p)
	case *model.SameOrAfter:
		return afterOrEqualDateTimeWithPrecision(l, r, p)
	case *model.SameOrBefore:
		return beforeOrEqualDateTimeWithPrecision(l, r, p)
	}
	return result.Value{}, fmt.Errorf("internal error - unsupported Binary Comparison Expression %v", b)
}

// op(left DateTime, right DateTime) Boolean
// https://cql.hl7.org/09-b-cqlreference.html#after
// https://cql.hl7.org/09-b-cqlreference.html#before
// https://cql.hl7.org/09-b-cqlreference.html#same-or-after-1
// https://cql.hl7.org/09-b-cqlreference.html#same-or-before-1
func evalCompareDateTimeWithPrecision(b model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) || result.IsNull(rObj) {
		return result.New(nil)
	}
	p, err := precisionFromBinaryExpression(b)
	if err != nil {
		return result.Value{}, err
	}
	allowUnsetPrec := true
	if err := validateDateTimePrecision(p, allowUnsetPrec); err != nil {
		return result.Value{}, err
	}
	l, r, err := applyToValues(lObj, rObj, result.ToDateTime)
	if err != nil {
		return result.Value{}, err
	}

	switch b.(type) {
	case *model.After:
		return afterDateTimeWithPrecision(l, r, p)
	case *model.Before:
		return beforeDateTimeWithPrecision(l, r, p)
	case *model.SameOrAfter:
		return afterOrEqualDateTimeWithPrecision(l, r, p)
	case *model.SameOrBefore:
		return beforeOrEqualDateTimeWithPrecision(l, r, p)
	}
	return result.Value{}, fmt.Errorf("internal error - unsupported Binary Comparison Expression %v", b)
}

func precisionFromBinaryExpression(b model.IBinaryExpression) (model.DateTimePrecision, error) {
	var p model.DateTimePrecision
	switch t := b.(type) {
	case *model.After:
		p = t.Precision
	case *model.Before:
		p = t.Precision
	case *model.SameOrAfter:
		p = t.Precision
	case *model.SameOrBefore:
		p = t.Precision
	default:
		return model.DateTimePrecision(""), fmt.Errorf("internal error - unsupported Binary Comparison Expression %v", b)
	}
	return p, nil
}

// afterDateTime returns whether or not the given DateTimeValue comes after the right DateTimeValue.
// Returns null in cases where values cannot be compared such as right precision being less than
// left precision.
func afterDateTime(l, r result.DateTime) (result.Value, error) {
	compareResult, err := compareDateTime(l, r)
	if err != nil {
		return result.Value{}, err
	}
	switch compareResult {
	case leftBeforeRight:
		return result.New(false)
	case leftEqualRight:
		return result.New(false)
	case leftAfterRight:
		return result.New(true)
	case insufficientPrecision:
		return result.New(nil)
	}
	return result.Value{}, errors.New("internal error - reached the end of timeComparison enum in dateAfter")
}

// afterDateTimeWithPrecision returns whether or not the given DateTimeValue comes after the right
// DateTimeValue up to the given precision. Returns null in cases where values cannot be compared
// such as right precision being less than left precision.
func afterDateTimeWithPrecision(l, r result.DateTime, p model.DateTimePrecision) (result.Value, error) {
	compareResult, err := compareDateTimeWithPrecision(l, r, p)
	if err != nil {
		return result.Value{}, err
	}
	switch compareResult {
	case leftBeforeRight:
		return result.New(false)
	case leftEqualRight:
		return result.New(false)
	case leftAfterRight:
		return result.New(true)
	case insufficientPrecision:
		return result.New(nil)
	}
	return result.Value{}, errors.New("internal error - reached the end of timeComparison enum in dateAfter")
}

func afterOrEqualDateTime(l, r result.DateTime) (result.Value, error) {
	compareResult, err := compareDateTime(l, r)
	if err != nil {
		return result.Value{}, err
	}
	switch compareResult {
	case leftBeforeRight:
		return result.New(false)
	case leftEqualRight:
		return result.New(true)
	case leftAfterRight:
		return result.New(true)
	case insufficientPrecision:
		return result.New(nil)
	}
	return result.Value{}, errors.New("internal error - reached the end of timeComparison enum in dateAfter")
}

// afterOrEqualDateTimeWithPrecision returns whether or not the given DateTimeValue is on or after
// the right DateTimeValue up to the given precision. Returns null in cases where values cannot be
// compared such as right precision being less than left precision.
func afterOrEqualDateTimeWithPrecision(l, r result.DateTime, p model.DateTimePrecision) (result.Value, error) {
	compareResult, err := compareDateTimeWithPrecision(l, r, p)
	if err != nil {
		return result.Value{}, err
	}
	switch compareResult {
	case leftBeforeRight:
		return result.New(false)
	case leftEqualRight:
		return result.New(true)
	case leftAfterRight:
		return result.New(true)
	case insufficientPrecision:
		return result.New(nil)
	}
	return result.Value{}, errors.New("internal error - reached the end of timeComparison enum in dateAfter")
}

// beforeDateTime returns whether or not the given DateTimeValue comes before the right
// DateTimeValue. Returns null in cases where values cannot be compared such as right precision
// being less than left precision.
func beforeDateTime(l, r result.DateTime) (result.Value, error) {
	compareResult, err := compareDateTime(l, r)
	if err != nil {
		return result.Value{}, err
	}
	switch compareResult {
	case leftBeforeRight:
		return result.New(true)
	case leftEqualRight:
		return result.New(false)
	case leftAfterRight:
		return result.New(false)
	case insufficientPrecision:
		return result.New(nil)
	}
	return result.Value{}, errors.New("internal error - reached the end of timeComparison enum in dateTimeBefore")
}

func beforeOrEqualDateTime(l, r result.DateTime) (result.Value, error) {
	compareResult, err := compareDateTime(l, r)
	if err != nil {
		return result.Value{}, err
	}
	switch compareResult {
	case leftBeforeRight:
		return result.New(true)
	case leftEqualRight:
		return result.New(true)
	case leftAfterRight:
		return result.New(false)
	case insufficientPrecision:
		return result.New(nil)
	}
	return result.Value{}, errors.New("internal error - reached the end of timeComparison enum in dateTimeBefore")
}

// beforeDateTimeWithPrecision returns whether or not the given DateTimeValue comes before the right
// DateTimeValue up to the given precision. Returns null in cases where values cannot be compared
// such as right precision being less than left precision.
func beforeDateTimeWithPrecision(l, r result.DateTime, p model.DateTimePrecision) (result.Value, error) {
	compareResult, err := compareDateTimeWithPrecision(l, r, p)
	if err != nil {
		return result.Value{}, err
	}
	switch compareResult {
	case leftBeforeRight:
		return result.New(true)
	case leftEqualRight:
		return result.New(false)
	case leftAfterRight:
		return result.New(false)
	case insufficientPrecision:
		return result.New(nil)
	}
	return result.Value{}, errors.New("internal error - reached the end of timeComparison enum in dateTimeBefore")
}

// beforeOrEqualDateTimeWithPrecision returns whether or not the given DateTimeValue is on or before
// the right DateTimeValue up to the given precision. Returns null in cases where values cannot be
// compared such as right precision being less than left precision.
func beforeOrEqualDateTimeWithPrecision(l, r result.DateTime, p model.DateTimePrecision) (result.Value, error) {
	compareResult, err := compareDateTimeWithPrecision(l, r, p)
	if err != nil {
		return result.Value{}, err
	}
	switch compareResult {
	case leftBeforeRight:
		return result.New(true)
	case leftEqualRight:
		return result.New(true)
	case leftAfterRight:
		return result.New(false)
	case insufficientPrecision:
		return result.New(nil)
	}
	return result.Value{}, errors.New("internal error - reached the end of timeComparison enum in dateTimeBefore")
}

// CanConvertQuantity(left Quantity, right String) Boolean
// https://cql.hl7.org/09-b-cqlreference.html#canconvertquantity
// Returns whether or not a Quantity can be converted into the given unit string.
// This is not a required function to implement, and for the time being we are
// choosing to not implement unit conversion so this function always returns false.
func evalCanConvertQuantity(b model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) || result.IsNull(rObj) {
		return result.New(nil)
	}
	return result.New(false)
}

// difference in _precision_ between(left Date, right Date) Integer
// https://cql.hl7.org/09-b-cqlreference.html#difference
// Returns the number of boundaries crossed between two dates.
func evalDifferenceBetweenDate(b model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	m := b.(*model.DifferenceBetween)
	p := model.DateTimePrecision(m.Precision)
	if err := validatePrecision(p, []model.DateTimePrecision{model.YEAR, model.MONTH, model.WEEK, model.DAY}); err != nil {
		return result.Value{}, err
	}

	if result.IsNull(lObj) || result.IsNull(rObj) {
		return result.New(nil)
	}

	l, r, err := applyToValues(lObj, rObj, result.ToDateTime)
	if err != nil {
		return result.Value{}, err
	}
	return dateTimeDifference(l, r, p)
}

// difference in _precision_ between(left DateTime, right DateTime) Integer
// https://cql.hl7.org/09-b-cqlreference.html#difference
// Returns the number of boundaries crossed between two datetimes.
func evalDifferenceBetweenDateTime(b model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	m := b.(*model.DifferenceBetween)
	p := model.DateTimePrecision(m.Precision)
	if err := validatePrecision(p, []model.DateTimePrecision{model.YEAR, model.MONTH, model.WEEK, model.DAY, model.HOUR, model.MINUTE, model.SECOND, model.MILLISECOND}); err != nil {
		return result.Value{}, err
	}

	if result.IsNull(lObj) || result.IsNull(rObj) {
		return result.New(nil)
	}

	l, r, err := applyToValues(lObj, rObj, result.ToDateTime)
	if err != nil {
		return result.Value{}, err
	}
	return dateTimeDifference(l, r, p)
}

// Now() DateTime
// https://cql.hl7.org/09-b-cqlreference.html#now
// Returns the evaluation timestamp value in DateTime format.
func (i *interpreter) evalNow(n model.INaryExpression, _ []result.Value) (result.Value, error) {
	return result.New(result.DateTime{Date: i.evaluationTimestamp, Precision: model.MILLISECOND})
}

// Date(year Integer) Date
// Date(year Integer, month Integer) Date
// Date(year Integer, month Integer, day Integer) Date
// https://cql.hl7.org/09-b-cqlreference.html#date-1
func (i *interpreter) evalDate(n model.INaryExpression, objs []result.Value) (result.Value, error) {
	if result.IsNull(objs[0]) {
		return result.Value{}, fmt.Errorf("in Date %v cannot be null", model.YEAR)
	}

	var dateVals []int
	foundNull := false
	precisions := []model.DateTimePrecision{model.YEAR, model.MONTH, model.DAY}
	for i := range objs {
		if result.IsNull(objs[i]) {
			foundNull = true
			continue
		}
		if foundNull {
			return result.Value{}, fmt.Errorf("when constructing Date precision %v had value %v, even though a higher precision was null", precisions[i], objs[i].GolangValue())
		}
		v, err := result.ToInt32(objs[i])
		if err != nil {
			return result.Value{}, err
		}
		dateVals = append(dateVals, int(v))
	}

	switch len(dateVals) {
	case 1:
		t := time.Date(dateVals[0], 1, 1, 0, 0, 0, 0, time.UTC)
		if err := validateDateTime(dateVals, t); err != nil {
			return result.Value{}, err
		}
		return result.New(result.Date{Date: t, Precision: model.YEAR})
	case 2:
		t := time.Date(dateVals[0], time.Month(dateVals[1]), 1, 0, 0, 0, 0, time.UTC)
		if err := validateDateTime(dateVals, t); err != nil {
			return result.Value{}, err
		}
		return result.New(result.Date{Date: t, Precision: model.MONTH})
	case 3:
		t := time.Date(dateVals[0], time.Month(dateVals[1]), dateVals[2], 0, 0, 0, 0, time.UTC)
		if err := validateDateTime(dateVals, t); err != nil {
			return result.Value{}, err
		}
		return result.New(result.Date{Date: t, Precision: model.DAY})
	default:
		return result.Value{}, errors.New("internal error - should never receive Date with more than 3 arguments")
	}
}

// DateTime(year Integer) DateTime
// DateTime(year Integer, month Integer) DateTime
// DateTime(year Integer, month Integer, day Integer) DateTime
// DateTime(year Integer, month Integer, day Integer, hour Integer) DateTime
// DateTime(year Integer, month Integer, day Integer, hour Integer, minute Integer) DateTime
// DateTime(year Integer, month Integer, day Integer, hour Integer, minute Integer, second Integer) DateTime
// DateTime(year Integer, month Integer, day Integer, hour Integer, minute Integer, second Integer, millisecond Integer) DateTime
// DateTime(year Integer, month Integer, day Integer, hour Integer, minute Integer, second Integer, millisecond Integer, timezoneOffset Decimal) DateTime
// https://cql.hl7.org/09-b-cqlreference.html#datetime-1
func (i *interpreter) evalDateTime(n model.INaryExpression, objs []result.Value) (result.Value, error) {
	allNull := true
	for _, obj := range objs {
		if !result.IsNull(obj) {
			allNull = false
			break
		}
	}
	if allNull {
		return result.New(nil)
	}

	loc := i.evaluationTimestamp.Location()
	if len(objs) == 8 {
		v, err := result.ToFloat64(objs[7])
		if err != nil {
			return result.Value{}, err
		}
		if v > 14 || v < -14 {
			return result.Value{}, fmt.Errorf("timezone offset %v is out of range", v)
		}
		// int() will truncate timezones with greater than second precision.
		loc = time.FixedZone(fmt.Sprintf("%v", v), int(v*60.0*60.0))
		objs = objs[:7]
	}

	var dateVals []int
	isNull := false
	precisions := []model.DateTimePrecision{model.YEAR, model.MONTH, model.DAY, model.HOUR, model.MINUTE, model.SECOND, model.MILLISECOND}
	for i := range objs {
		if result.IsNull(objs[i]) {
			isNull = true
			continue
		}
		if isNull {
			return result.Value{}, fmt.Errorf("when constructing DateTime precision %v had value %v, even though a higher precision was null", precisions[i], objs[i])
		}
		v, err := result.ToInt32(objs[i])
		if err != nil {
			return result.Value{}, err
		}
		dateVals = append(dateVals, int(v))
	}

	switch len(dateVals) {
	case 1:
		t := time.Date(dateVals[0], 1, 1, 0, 0, 0, 0, loc)
		if err := validateDateTime(dateVals, t); err != nil {
			return result.Value{}, err
		}
		return result.New(result.DateTime{Date: t, Precision: model.YEAR})
	case 2:
		t := time.Date(dateVals[0], time.Month(dateVals[1]), 1, 0, 0, 0, 0, loc)
		return result.New(result.DateTime{Date: t, Precision: model.MONTH})
	case 3:
		t := time.Date(dateVals[0], time.Month(dateVals[1]), dateVals[2], 0, 0, 0, 0, loc)
		if err := validateDateTime(dateVals, t); err != nil {
			return result.Value{}, err
		}
		return result.New(result.DateTime{Date: t, Precision: model.DAY})
	case 4:
		t := time.Date(dateVals[0], time.Month(dateVals[1]), dateVals[2], dateVals[3], 0, 0, 0, loc)
		if err := validateDateTime(dateVals, t); err != nil {
			return result.Value{}, err
		}
		return result.New(result.DateTime{Date: t, Precision: model.HOUR})
	case 5:
		t := time.Date(dateVals[0], time.Month(dateVals[1]), dateVals[2], dateVals[3], dateVals[4], 0, 0, loc)
		if err := validateDateTime(dateVals, t); err != nil {
			return result.Value{}, err
		}
		return result.New(result.DateTime{Date: t, Precision: model.MINUTE})
	case 6:
		t := time.Date(dateVals[0], time.Month(dateVals[1]), dateVals[2], dateVals[3], dateVals[4], dateVals[5], 0, loc)
		if err := validateDateTime(dateVals, t); err != nil {
			return result.Value{}, err
		}
		return result.New(result.DateTime{Date: t, Precision: model.SECOND})
	case 7:
		t := time.Date(dateVals[0], time.Month(dateVals[1]), dateVals[2], dateVals[3], dateVals[4], dateVals[5], dateVals[6]*int(time.Millisecond/time.Nanosecond), loc)
		if err := validateDateTime(dateVals, t); err != nil {
			return result.Value{}, err
		}
		return result.New(result.DateTime{Date: t, Precision: model.MILLISECOND})
	default:
		return result.Value{}, errors.New("internal error - should never receive DateTime with more than 8 arguments")
	}
}

// Time(hour Integer) Time
// Time(hour Integer, minute Integer) Time
// Time(hour Integer, minute Integer, second Integer) Time
// Time(hour Integer, minute Integer, second Integer, millisecond Integer) Time
// https://cql.hl7.org/09-b-cqlreference.html#time-1
func (i *interpreter) evalTime(n model.INaryExpression, objs []result.Value) (result.Value, error) {
	if result.IsNull(objs[0]) {
		return result.Value{}, fmt.Errorf("in Time %v cannot be null", model.HOUR)
	}

	var timeVals []int
	foundNull := false
	precisions := []model.DateTimePrecision{model.HOUR, model.MINUTE, model.SECOND, model.MILLISECOND}
	for i := range objs {
		if result.IsNull(objs[i]) {
			foundNull = true
			continue
		}
		if foundNull {
			return result.Value{}, fmt.Errorf("when constructing Time precision %v had value %v, even though a higher precision was null", precisions[i], objs[i].GolangValue())
		}
		v, err := result.ToInt32(objs[i])
		if err != nil {
			return result.Value{}, err
		}
		timeVals = append(timeVals, int(v))
	}

	switch len(timeVals) {
	case 1:
		t := time.Date(0, 1, 1, timeVals[0], 0, 0, 0, i.evaluationTimestamp.Location())
		if err := validateTime(timeVals, t); err != nil {
			return result.Value{}, err
		}
		return result.New(result.Time{Date: t, Precision: model.HOUR})
	case 2:
		t := time.Date(0, 1, 1, timeVals[0], timeVals[1], 0, 0, i.evaluationTimestamp.Location())
		if err := validateTime(timeVals, t); err != nil {
			return result.Value{}, err
		}
		return result.New(result.Time{Date: t, Precision: model.MINUTE})
	case 3:
		t := time.Date(0, 1, 1, timeVals[0], timeVals[1], timeVals[2], 0, i.evaluationTimestamp.Location())
		if err := validateTime(timeVals, t); err != nil {
			return result.Value{}, err
		}
		return result.New(result.Time{Date: t, Precision: model.SECOND})
	case 4:
		t := time.Date(0, 1, 1, timeVals[0], timeVals[1], timeVals[2], timeVals[3]*int(time.Millisecond/time.Nanosecond), i.evaluationTimestamp.Location())
		if err := validateTime(timeVals, t); err != nil {
			return result.Value{}, err
		}
		return result.New(result.Time{Date: t, Precision: model.MILLISECOND})
	default:
		return result.Value{}, errors.New("internal error - should never receive Time with more than 4 arguments")
	}
}

// Golang time.Date() values may be outside their usual ranges and will be normalized. This is not
// the desired CQL behaviour so we check if they have been normalized and return an error.
func validateDateTime(dateVals []int, t time.Time) error {
	if dateVals[0] < 1 || dateVals[0] > 9999 {
		return fmt.Errorf("%v %v is out of range", model.YEAR, dateVals[0])
	}

	for i := range dateVals {
		switch i {
		case 0:
			if t.Year() != dateVals[0] {
				return fmt.Errorf("%v %v is out of range", model.MONTH, dateVals[1])
			}
		case 1:
			if int(t.Month()) != dateVals[1] {
				return fmt.Errorf("%v %v is out of range", model.DAY, dateVals[2])
			}
		case 2:
			if t.Day() != dateVals[2] {
				return fmt.Errorf("%v %v is out of range", model.HOUR, dateVals[3])
			}
		case 3:
			if t.Hour() != dateVals[3] {
				return fmt.Errorf("%v %v is out of range", model.MINUTE, dateVals[4])
			}
		case 4:
			if t.Minute() != dateVals[4] {
				return fmt.Errorf("%v %v is out of range", model.SECOND, dateVals[5])
			}
		case 5:
			if t.Second() != dateVals[5] {
				return fmt.Errorf("%v %v is out of range", model.MILLISECOND, dateVals[6])
			}
		}
	}
	return nil
}

// Golang time.Date() values may be outside their usual ranges and will be normalized. This is not
// the desired CQL behaviour so we check if they have been normalized and return an error.
func validateTime(dateVals []int, t time.Time) error {
	for i := range dateVals {
		switch i {
		case 0:
			if t.Day() != 1 {
				return fmt.Errorf("%v %v is out of range", model.HOUR, dateVals[0])
			}
		case 1:
			if t.Hour() != dateVals[0] {
				return fmt.Errorf("%v %v is out of range", model.MINUTE, dateVals[1])
			}
		case 2:
			if t.Minute() != dateVals[1] {
				return fmt.Errorf("%v %v is out of range", model.SECOND, dateVals[2])
			}
		case 3:
			if t.Second() != dateVals[2] {
				return fmt.Errorf("%v %v is out of range", model.MILLISECOND, dateVals[3])
			}
		}
	}
	return nil
}

// TimeOfDay() Time
// https://cql.hl7.org/09-b-cqlreference.html#timeofday
// Returns the time of the evaluation timestamp value as a Time value.
// TODO: b/346805860 - Enforce execution timestamp have millisecond precision.
func (i *interpreter) evalTimeOfDay(_ model.INaryExpression, _ []result.Value) (result.Value, error) {
	t := time.Date(0, time.January, 1, i.evaluationTimestamp.Hour(), i.evaluationTimestamp.Minute(), i.evaluationTimestamp.Second(), i.evaluationTimestamp.Nanosecond(), i.evaluationTimestamp.Location())
	return result.New(result.Time{Date: t, Precision: model.MILLISECOND})
}

// Today() Date
// https://cql.hl7.org/09-b-cqlreference.html#today
// Returns the evaluation timestamp value in Date format.
func (i *interpreter) evalToday(n model.INaryExpression, _ []result.Value) (result.Value, error) {
	year, month, day := i.evaluationTimestamp.Date()
	return result.New(result.Date{
		Date:      time.Date(year, month, day, 0, 0, 0, 0, time.UTC),
		Precision: model.DAY,
	})
}

// dateTimeDifference returns the difference at the desired precision for two Go time values.
// Left value can be greater than right value, in such cases a negative value should be returned.
// TODO b/318386749 - Add uncertainty logic once uncertainties are implemented.
func dateTimeDifference(l, r result.DateTime, opPrecision model.DateTimePrecision) (result.Value, error) {
	if !precisionGreaterOrEqual(opPrecision, l.Precision) || !precisionGreaterOrEqual(opPrecision, r.Precision) {
		// TODO b/318386749 - precisionGreaterOrEqual is a temporary check to ensure the precision of
		// "difference in _precision_ between" is greater than the precision of the
		// l, r dateTimeValues. In the future we need to support all cases by returning
		// an uncertainty.
		return result.Value{}, fmt.Errorf("difference between specified a precision greater than argument precision got, %s, and %s, wanted %v", l.Precision, r.Precision, opPrecision)
	}
	left, right := l.Date, r.Date

	switch opPrecision {
	case model.YEAR:
		return result.New(right.Year() - left.Year())
	case model.MONTH:
		return result.New(12*(right.Year()-left.Year()) + int((right.Month())) - int(left.Month()))
	case model.WEEK:
		// Weekly borders crossed are number of times a Sunday boundary has been crossed.
		// TODO(b/301606416): Weeks do not correctly support negative values.
		diffInDays := int(right.Sub(left).Hours() / 24)
		leftDaysSinceSunday, rightDaysSinceSunday := int(left.Weekday()), int(right.Weekday())
		if diffInDays < 7 && rightDaysSinceSunday < leftDaysSinceSunday {
			return result.New(1)
		} else if diffInDays < 7 {
			return result.New(0)
		}
		// There is at least one week here. Remove the left side days until Sunday and add a week to account for that.
		// From there the number of remaining weeks are only whole seven day weeks.
		return result.New(int((diffInDays-(7-leftDaysSinceSunday))/7) + 1)
	case model.DAY:
		// This logic is to ensure we are only counting by day boundaries crossed.
		epoch := time.UnixMilli(0)
		leftDaysSinceEpoch := int(left.Sub(epoch).Hours() / 24)
		rightDaysSinceEpoch := int(right.Sub(epoch).Hours() / 24)
		return result.New(rightDaysSinceEpoch - leftDaysSinceEpoch)
	case model.HOUR:
		return result.New(int(right.Sub(left).Hours()))
	case model.MINUTE:
		return result.New(int(right.Sub(left).Minutes()))
	case model.SECOND:
		// TODO(b/301606416): According to the spec seconds and milliseconds should be combined and
		// compared as a decimal. It is not clear what this means, but this implementation may be
		// incorrect.
		return result.New(int(right.Sub(left).Seconds()))
	case model.MILLISECOND:
		return result.New(int(right.Sub(left).Milliseconds()))
	default:
		return result.Value{}, fmt.Errorf("unsupported precision for dateTimeDifference: %v", opPrecision)
	}
}

type comparison int

const (
	unsetComparison comparison = iota
	leftBeforeRight
	leftEqualRight
	leftAfterRight
	insufficientPrecision
	comparedToNull
)

// orderedPrecisions are DateTimePrecisions ordered from least precise to most precise.
var orderedPrecisions = []model.DateTimePrecision{model.YEAR, model.MONTH, model.DAY, model.WEEK, model.HOUR, model.MINUTE, model.SECOND, model.MILLISECOND}

// getFinestPrecision returns the finest (most precise) precision between two DateTimePrecisions. An
// error is returned if any of the precisions are unset.
func getFinestPrecision(l, r model.DateTimePrecision) (model.DateTimePrecision, error) {
	if l == model.UNSETDATETIMEPRECISION || r == model.UNSETDATETIMEPRECISION {
		return model.UNSETDATETIMEPRECISION, fmt.Errorf("internal error -- input to getFinestPrecision must not be unset. got: %v, %v", l, r)
	}

	// Iterating over precisions from least precise to most precise.
	for _, currPrec := range orderedPrecisions {
		// If one precision matches, return the _other_ one, which must be equally or more precise.
		if l == currPrec {
			return r, nil
		}
		if r == currPrec {
			return l, nil
		}
	}
	// We should not get here:
	return model.UNSETDATETIMEPRECISION, fmt.Errorf("internal error -- unable to get finest precision for: %v, %v", l, r)
}

// compareDateTimeWithPrecision returns a comparison of DateTimeValues with the given maximum
// TimePrecision. If either left or right has insufficient precision to determine which is greater
// before reaching maxPrecision then insufficientPrecision is returned.
func compareDateTimeWithPrecision(left, right result.DateTime, maxPrecision model.DateTimePrecision) (comparison, error) {
	if maxPrecision == model.UNSETDATETIMEPRECISION {
		// If precision is unset, proceed until the finest precision specified by either input.
		finestPrecision, err := getFinestPrecision(left.Precision, right.Precision)
		if err != nil {
			return unsetComparison, err
		}
		maxPrecision = finestPrecision
	}

	left = normalizeDateTime(left)
	right = normalizeDateTime(right)

	for _, p := range orderedPrecisions {
		switch p {
		case model.YEAR:
			if r := cmp.Compare(left.Date.Year(), right.Date.Year()); r != 0 {
				return toComparison(r), nil
			}
		case model.MONTH:
			if r := cmp.Compare(left.Date.Month(), right.Date.Month()); r != 0 {
				return toComparison(r), nil
			}
		// Note that week is intentionally skipped because it is not valid for DateTime/Dates.
		case model.DAY:
			if r := cmp.Compare(left.Date.Day(), right.Date.Day()); r != 0 {
				return toComparison(r), nil
			}
		case model.HOUR:
			if r := cmp.Compare(left.Date.Hour(), right.Date.Hour()); r != 0 {
				return toComparison(r), nil
			}
		case model.MINUTE:
			if r := cmp.Compare(left.Date.Minute(), right.Date.Minute()); r != 0 {
				return toComparison(r), nil
			}
		// TODO: b/329321570 - According to the spec, we may need to combine seconds and milliseconds
		// into a decimal, and do the comparison at the seconds precision.
		case model.SECOND:
			if r := cmp.Compare(left.Date.Second(), right.Date.Second()); r != 0 {
				return toComparison(r), nil
			}
		case model.MILLISECOND:
			r := cmp.Compare(left.Date.UnixMilli(), right.Date.UnixMilli())
			return toComparison(r), nil
		}
		if p == maxPrecision {
			// Reached the max required precision, so they are equal.
			return leftEqualRight, nil
		}
		if p == left.Precision || p == right.Precision {
			return insufficientPrecision, nil
		}
	}
	return leftEqualRight, nil
}

// compareDateTime returns a pure comparison of DateTimeValues. If left and right are equal up
// to the precision of only one of the two values insufficientPrecision is returned.
func compareDateTime(left, right result.DateTime) (comparison, error) {
	return compareDateTimeWithPrecision(left, right, model.UNSETDATETIMEPRECISION)
}

func normalizeDateTime(d result.DateTime) result.DateTime {
	if precisionGreaterOrEqual(d.Precision, model.DAY) {
		return d
	}
	d.Date = d.Date.In(time.UTC)
	return d
}

// toComparison converts the result of cmp.Compare() to comparison.
func toComparison(a int) comparison {
	switch a {
	case -1:
		return leftBeforeRight
	case 0:
		return leftEqualRight
	case 1:
		return leftAfterRight
	}
	return unsetComparison
}

func validateDateTimePrecision(precision model.DateTimePrecision, allowUnset bool) error {
	allowed := []model.DateTimePrecision{model.YEAR, model.MONTH, model.DAY, model.HOUR, model.MINUTE, model.SECOND, model.MILLISECOND}
	if allowUnset {
		allowed = append(allowed, model.UNSETDATETIMEPRECISION)
	}
	return validatePrecision(precision, allowed)
}

func validateDatePrecision(precision model.DateTimePrecision, allowUnset bool) error {
	allowed := []model.DateTimePrecision{model.YEAR, model.MONTH, model.DAY}
	if allowUnset {
		allowed = append(allowed, model.UNSETDATETIMEPRECISION)
	}
	return validatePrecision(precision, allowed)
}

// validatePrecision returns an error if p is not in validPs.
func validatePrecision(p model.DateTimePrecision, validPs []model.DateTimePrecision) error {
	for _, v := range validPs {
		if p == v {
			return nil
		}
	}
	return fmt.Errorf("precision must be one of %v, got %v", validPs, p)
}

func validatePrecisionByType(precision model.DateTimePrecision, allowUnset bool, dateType types.IType) error {
	switch dateType {
	case types.Date:
		return validateDatePrecision(precision, allowUnset)
	case types.DateTime:
		return validateDateTimePrecision(precision, allowUnset)
	default:
		return fmt.Errorf("unsupported type for validatePrecisionByType got: %v, expected: types.Date, types.DateTime", dateType)
	}
}

// precisionGreaterOrEqual returns true if l is of greater or equal precision than r.
func precisionGreaterOrEqual(l, r model.DateTimePrecision) bool {
	if l == r {
		return true
	}
	precisions := []model.DateTimePrecision{model.YEAR, model.MONTH, model.WEEK, model.DAY, model.HOUR, model.MINUTE, model.SECOND, model.MILLISECOND}
	for _, p := range precisions {
		if p == l {
			return true
		}
		if p == r {
			return false
		}
	}
	return false
}

// convertQuantityUpToPrecision converts a QuantityValue to a given precision.
// Returns error for cases where a conversion cannot be performed.
// This function only converts upwards.
func convertQuantityUpToPrecision(q result.Quantity, wantPrecision model.DateTimePrecision) (result.Quantity, error) {
	qp := model.DateTimePrecision(q.Unit)
	qv := q.Value
	// validateDateTimePrecision explicitly does not check for 'week' values so we need to
	// do that manually.
	err := validatePrecision(qp, []model.DateTimePrecision{model.YEAR, model.MONTH, model.WEEK, model.DAY, model.HOUR, model.MINUTE, model.SECOND, model.MILLISECOND})
	if err != nil {
		return result.Quantity{}, err
	}
	if precisionGreaterOrEqual(qp, wantPrecision) {
		return result.Quantity{Value: qv, Unit: model.Unit(qp)}, nil
	}
	if qp == model.WEEK {
		return result.Quantity{}, fmt.Errorf("error: cannot convert from week to a higher precision for Date/DateTime values. want: %v, got: %v", wantPrecision, q.Unit)
	}

	// It's considered an error to convert from days/weeks up to month.
	precisions := []model.DateTimePrecision{model.MILLISECOND, model.SECOND, model.MINUTE, model.HOUR, model.DAY, model.MONTH, model.YEAR}
	foundStartPrecision := false
	// iterate up to the precision of the quantity then start converting.
	for _, p := range precisions {
		if p == qp {
			foundStartPrecision = true
			continue
		}
		if !foundStartPrecision {
			continue
		}

		switch p {
		case model.SECOND:
			qv = qv / 1000
		case model.MINUTE,
			model.HOUR:
			qv = qv / 60
		case model.DAY:
			qv = qv / 24
		case model.MONTH:
			return result.Quantity{}, fmt.Errorf("error: invalid unit conversion, starting precision cannot be converted to be more precise than days. want: %v, got: %v", wantPrecision, q.Unit)
		case model.YEAR:
			qv = qv / 12
		}
		if p == wantPrecision {
			return result.Quantity{Value: qv, Unit: model.Unit(wantPrecision)}, nil
		}
	}
	return result.Quantity{}, fmt.Errorf("error: failed to reach desired precision when adding Date/DateTime to Quantity with precisions want: %v, got: %v", wantPrecision, q.Unit)
}
