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

// Package exclusions contains the test group and test name exclusions for the XML tests.
package exclusions

// Exclusions for the XML tests.

// XMLTestFileExclusions contains the test group and test name exclusions for a given XML test file.
type XMLTestFileExclusions struct {
	GroupExcludes []string
	NamesExcludes []string
}

// XMLTestFileExclusionDefinitions returns all of the test group and test name exclusions. A TODO
// should be included for each set of skipped tests.
func XMLTestFileExclusionDefinitions() map[string]XMLTestFileExclusions {
	return map[string]XMLTestFileExclusions{
		"CqlAggregateFunctionsTest.xml": {
			GroupExcludes: []string{},
			NamesExcludes: []string{
				// TODO: b/344002938 - xml test seems wrong, asserts ml when it should be mL.
				"SumTestQuantity",
			},
		},
		"CqlAggregateTest.xml": {
			GroupExcludes: []string{},
			NamesExcludes: []string{
				// TODO: b/342061715 - unsupported operators.
				"RolledOutIntervals",
			},
		},
		"CqlArithmeticFunctionsTest.xml": {
			GroupExcludes: []string{
				// TODO: b/342061715 - unsupported operators.
				"HighBoundary",
				"LowBoundary",
			},
			NamesExcludes: []string{
				// TODO: b/342061606 - Unit conversion.
				"TruncatedDivide10By5DQuantity",
				"TruncatedDivide414By206DQuantity",
				// TODO: b/342061783 - Got unexpected result.
				"Subtract2And11D",
				"TruncatedDivide10d1ByNeg3D1Quantity",
				"PrecisionDecimal", // Does not yet support trailing zeros.
				// TODO: b/344002938 - xml test is wrong, asserts with a time zone.
				"DateTimeMinValue",
				"DateTimeMaxValue",
			},
		},
		"CqlComparisonOperatorsTest.xml": {
			GroupExcludes: []string{},
			NamesExcludes: []string{
				// TODO: b/342061715 - Unsupported operator.
				"DateTimeDayCompare",
				"TimeGreaterTrue",
				"TimeGreaterFalse",
				"TimeGreaterEqTrue",
				"TimeGreaterEqTrue2",
				"TimeGreaterEqFalse",
				"TimeLessTrue",
				"TimeLessFalse",
				"TimeLessEqTrue",
				"TimeLessEqTrue2",
				"TimeLessEqFalse",
				"EquivTupleJohnJohn",
				"EquivTupleJohnJohnWithNulls",
				"EquivTupleJohnJane",
				"EquivTupleJohn1John2",
				"EquivTime10A10A",
				"EquivTime10A10P",
				// TODO: b/342061783 - Got unexpected result.
				"TupleEqJohn1John1WithNullName",
				"TupleNotEqJohn1John1WithNullName",
			},
		},
		"CqlDateTimeOperatorsTest.xml": {
			GroupExcludes: []string{
				// TODO: b/342061715 - unsupported operators.
				"DateTimeComponentFrom",
				// TODO: b/342064491 - runtime error: invalid memory address or nil pointer dereference.
				"SameAs",
			},
			NamesExcludes: []string{
				// TODO: b/342061715 - unsupported operators.
				"DateTimeAddYearInWeeks",
				"DateAdd2YearsAsMonths",
				"DateAdd2YearsAsMonthsRem1",
				"DateAdd1Year",
				"DateTimeSubtractYearInWeeks",
				"DateSubtract2YearsAsMonths",
				"DateSubtract2YearsAsMonthsRem1",
				"DateSubtract1Year",
				"DateTimeComponentFromYear",
				"DateTimeComponentFromMonth",
				"DateTimeComponentFromMonthMinBoundary",
				"DateTimeComponentFromDay",
				"DateTimeComponentFromHour",
				"DateTimeComponentFromMinute",
				"DateTimeComponentFromSecond",
				"DateTimeComponentFromMillisecond",
				"DateTimeComponentFromTimezone",
				"TimeComponentFromHour",
				"TimeComponentFromMinute",
				"TimeComponentFromSecond",
				"TimeComponentFromMilli",
				"TimeAdd5Hours",
				"TimeAdd1Minute",
				"TimeAdd1Second",
				"TimeAdd1Millisecond",
				"TimeAdd5Hours1Minute",
				"TimeAdd5hoursByMinute",
				"TimeAfterHourTrue",
				"TimeAfterHourFalse",
				"TimeAfterMinuteTrue",
				"TimeAfterMinuteFalse",
				"TimeAfterSecondTrue",
				"TimeAfterSecondFalse",
				"TimeAfterMillisecondTrue",
				"TimeAfterMillisecondFalse",
				"TimeAfterTimeCstor",
				"TimeBeforeHourTrue",
				"TimeBeforeHourFalse",
				"TimeBeforeMinuteTrue",
				"TimeBeforeMinuteFalse",
				"TimeBeforeSecondTrue",
				"TimeBeforeSecondFalse",
				"TimeBeforeMillisecondTrue",
				"TimeBeforeMillisecondFalse",
				"TimeDifferenceHour",
				"TimeDifferenceMinute",
				"TimeDifferenceSecond",
				"TimeDifferenceMillis",
				"TimeSameOrAfterHourTrue1",
				"TimeSameOrAfterHourTrue2",
				"TimeSameOrAfterHourFalse",
				"TimeSameOrAfterMinuteTrue1",
				"TimeSameOrAfterMinuteTrue2",
				"TimeSameOrAfterMinuteFalse",
				"TimeSameOrAfterSecondTrue1",
				"TimeSameOrAfterSecondTrue2",
				"TimeSameOrAfterSecondFalse",
				"TimeSameOrAfterMillisTrue1",
				"TimeSameOrAfterMillisTrue2",
				"TimeSameOrAfterMillisFalse",
				"TimeSameOrBeforeHourTrue1",
				"TimeSameOrBeforeHourTrue2",
				"TimeSameOrBeforeHourFalse",
				"TimeSameOrBeforeMinuteTrue1",
				"TimeSameOrBeforeMinuteFalse0",
				"TimeSameOrBeforeMinuteFalse",
				"TimeSameOrBeforeSecondTrue1",
				"TimeSameOrBeforeSecondFalse0",
				"TimeSameOrBeforeSecondFalse",
				"TimeSameOrBeforeMillisTrue1",
				"TimeSameOrBeforeMillisFalse0",
				"TimeSameOrBeforeMillisFalse",
				"TimeSubtract5Hours",
				"TimeSubtract1Minute",
				"TimeSubtract1Second",
				"TimeSubtract1Millisecond",
				"TimeSubtract5Hours1Minute",
				"TimeSubtract5hoursByMinute",
				// TODO: b/342064803 - Invalid unit conversion.
				"DateTimeAdd2YearsByDays",
				"DateTimeAdd2YearsByDaysRem5Days",
				// TODO: b/342064012 - Uncertain result.
				"DateAdd33Days",
				"DateSubtract33Days",
				"DateTimeDurationBetweenUncertainInterval",
				"DateTimeDurationBetweenUncertainInterval2",
				"DateTimeDurationBetweenUncertainAdd",
				"DateTimeDurationBetweenUncertainSubtract",
				"DateTimeDurationBetweenUncertainMultiply",
				"DateTimeDurationBetweenUncertainDivIf",
				"DateTimeDurationBetweenMonthUncertain",
				"DateTimeDurationBetweenMonthUncertain2",
				"DateTimeDurationBetweenMonthUncertain3",
				"DateTimeDurationBetweenMonthUncertain4",
				"DateTimeDurationBetweenMonthUncertain5",
				"DateTimeDurationBetweenMonthUncertain6",
				"DateTimeDurationBetweenMonthUncertain7",
				"DurationInWeeks",
				"DurationInWeeks2",
				"DateTimeSubtract1YearInSeconds",
				"TimeDurationBetweenHour",
				"TimeDurationBetweenHourDiffPrecision",
				"TimeDurationBetweenHourDiffPrecision2",
				"TimeDurationBetweenMinute",
				"TimeDurationBetweenSecond",
				"TimeDurationBetweenMillis",
				"DurationInHoursA",
				"DurationInMinutesA", 
				"DurationInDaysA",
				"DurationInDaysAA",
				"DateTimeDifferenceUncertain",
				// TODO: b/343800835 - Error in output date comparison based on execution timestamp logic.
				"DateTimeComponentFromDate",
				// TODO: b/342061783 - Got unexpected result.
				"DateTimeAddLeapYear",
			},
		},
		"CqlIntervalOperatorsTest.xml": {
			GroupExcludes: []string{
				// TODO: b/342061715 - unsupported operators.
				"After",
				"Before",
				"Collapse",
				"Expand",
				"Ends",
				"Except",
				"Includes",
				"Intersect",
				"Meets",
				"MeetsBefore",
				"MeetsAfter",
				"PointFrom",
				"ProperContains",
				"ProperIn",
				"ProperlyIncludes",
				"ProperlyIncludedIn",
				"Starts",
				"Union",
			},
			NamesExcludes: []string{
				// TODO: b/342061715 - unsupported operators.
				// Note: overlaps before and after are not supported. but these tests are missing the
				// before/after keyword for Date/Time test cases so they are not excluded.
				"TestOverlapsNull",
				"IntegerIntervalOverlapsTrue",
				"IntegerIntervalOverlapsTrue2",
				"IntegerIntervalOverlapsTrue3",
				"IntegerIntervalOverlapsFalse",
				"DecimalIntervalOverlapsTrue",
				"DecimalIntervalOverlapsFalse",
				"QuantityIntervalOverlapsTrue",
				"QuantityIntervalOverlapsFalse",
				"TestOverlapsBeforeNull",
				"IntegerIntervalOverlapsBeforeTrue",
				"IntegerIntervalOverlapsBeforeFalse",
				"IntegerIntervalOverlapsBeforeFalse2",
				"DecimalIntervalOverlapsBeforeTrue",
				"DecimalIntervalOverlapsBeforeFalse",
				"QuantityIntervalOverlapsBeforeTrue",
				"QuantityIntervalOverlapsBeforeFalse",
				"TestOverlapsAfterNull",
				"IntegerIntervalOverlapsAfterTrue",
				"IntegerIntervalOverlapsAfterFalse",
				"IntegerIntervalOverlapsAfterFalse2",
				"DecimalIntervalOverlapsAfterTrue",
				"DecimalIntervalOverlapsAfterFalse",
				"QuantityIntervalOverlapsAfterTrue",
				"QuantityIntervalOverlapsAfterFalse",
				"DecimalIntervalIncludedInTrue",
				"DecimalIntervalIncludedInFalse",
				"IntegerIntervalIncludedInTrue",
				"IntegerIntervalIncludedInFalse",
				"QuantityIntervalIncludedInTrue",
				"QuantityIntervalIncludedInFalse",
				"DateTimeIncludedInTrue",
				"DateTimeIncludedInFalse",
				"TimeIncludedInTrue",
				"TimeIncludedInFalse",
				"DateTimeIncludedInNull",
				"DateTimeIncludedInPrecisionTrue",
				"DateTimeIncludedInPrecisionNull",
				"TimeOverlapsAfterTrue",
				"TimeOverlapsAfterFalse",
				"TimeOverlapsBeforeTrue",
				"TimeOverlapsBeforeFalse",
				"TimeOverlapsTrue",
				"TimeOverlapsFalse",
				"TimeContainsFalse",
				"TimeContainsTrue",
				// TODO: b/342061783 - Got unexpected result.
				"TimeInTrue",
				"TimeInFalse",
				"TimeInNull",
				"Issue32Interval",
				"TimeEquivalentTrue",
				"TimeEquivalentFalse",
				"TestOnOrAfterDateTrue",
				"TestOnOrAfterTimeTrue",
				"TestOnOrAfterTimeFalse",
				"TestOnOrAfterIntegerTrue",
				"TestOnOrAfterDecimalFalse",
				"TestOnOrAfterQuantityTrue",
				"TestOnOrBeforeDateTrue",
				"TestOnOrBeforeTimeTrue",
				"TestOnOrBeforeTimeFalse",
				"TestOnOrBeforeIntegerTrue",
				"TestOnOrBeforeDecimalFalse",
				"TestOnOrBeforeQuantityTrue",
				// TODO: b/342064453 - Ambiguous match.
				"TestEqualNull",
				"TestInNullBoundaries",
			},
		},
		"CqlListOperatorsTest.xml": {
			GroupExcludes: []string{
				// TODO: b/342061715 - unsupported operators.
				"Descendents",
			},
			NamesExcludes: []string{
				// TODO: b/342061715 - unsupported operator.
				"ContainsNullLeft",
				"EqualABCAnd123",
				"Equal123AndABC",
				"Equal123AndString123",
				"EquivalentABCAnd123",
				"Equivalent123AndABC",
				"Equivalent123AndString123",
				"EquivalentTimeTrue",
				"EquivalentTimeFalse",
				// In this case the converter still can't tell if null should be converted to List<Decimal>
				// or List<Integer>.
				"IncludesNullLeft",
				"IncludedInNullRight",
				"NotEqualABCAnd123",
				"NotEqual123AndABC",
				"NotEqual123AndString123",
				// TODO: b/342061783 - Got unexpected result.
				"quantityList",
				"ProperContainsTimeNull",
				"ProperInTimeNull",
				// TODO: b/346880550 - These test appear to have incorrect assertions.
				"Except23And1234",
				"ProperlyIncludedInNulRight",
				"ProperlyIncludesNullLeft",
				"SkipAll",
				"TailOneElement",
				"TakeNullEmpty",
				"TakeEmpty",
			},
		},
		"CqlOverloadMatching.xml": {
			GroupExcludes: []string{
				// TODO: b/342061783 - xml tests don't yet support library tags.
				"OverloadMatching",
			},
			NamesExcludes: []string{},
		},
		"CqlQueryTests.xml": {
			GroupExcludes: []string{},
			NamesExcludes: []string{},
		},
		"CqlNullologicalOperatorsTest.xml": {
			GroupExcludes: []string{},
			NamesExcludes: []string{},
		},
		"CqlStringOperatorsTest.xml": {
			GroupExcludes: []string{
				// TODO: b/342061715 - unsupported operators.
			},
			NamesExcludes: []string{
				// TODO: b/346880550 - These test appear to have incorrect assertions.
				"DateTimeToString1",
				"DateTimeToString2",
			},
		},
		"CqlTypesTest.xml": {
			GroupExcludes: []string{},
			NamesExcludes: []string{
				// TODO: b/342064012 - Uncertain result.
				"DateTimeUncertain",
				// TODO: b/342061715 - unsupported operators.
				"DateTimeTimeUnspecified",
				// TODO: b/343515613 - fails with unexpected result. Technically not supported.
				"StringUnicodeTest",
				// TODO: b/343515819 - fails with unexpected result.
				"QuantityTest",
				"QuantityTest2",
				"StringTestEscapeQuotes",
			},
		},
		"CqlTypeOperatorsTest.xml": {
			GroupExcludes: []string{
				// TODO: b/342061715 - unsupported operators.
				"Convert",
				"ToBoolean",
				"ToConcept",
				"ToInteger",
				"ToTime",
			},
			NamesExcludes: []string{
				// TODO: b/342061715 - unsupported operators.
				"ToDateTimeTimeUnspecified",
				// TODO: b/343515613 - fails with unexpected result. Technically not supported.
				"ToDateTimeMalformed",
			},
		},
		"ValueLiteralsAndSelectors.xml": {
			GroupExcludes: []string{},
			NamesExcludes: []string{
				// TODO: b/342061715 - unsupported operator. These return a decimal
				// value at runtime but expects an integer. This operator should probably return a choice type.
				"DecimalNegOneStep",
				"DecimalNegTenStep",
			},
		},
	}
}
