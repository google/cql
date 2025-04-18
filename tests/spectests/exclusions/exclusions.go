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
		"CqlAggregateFunctionsTest.xml": XMLTestFileExclusions{
			GroupExcludes: []string{
				// TODO: b/342061715 - unsupported operators.
				"Mode",
				"PopulationVariance",
				"StdDev",
				"Variance",
			},
			NamesExcludes: []string{
				// TODO: b/342061715 - unsupported operators.
				// Only Date and DateTime overloads are supported for max/min.
				"MaxTestInteger",
				"MaxTestString",
				"MaxTestTime",
				"MinTestInteger",
				"MinTestString",
				"MinTestTime",
				// TODO: b/342061783 - Operator is supported but the test assertion uses a rounded value.
				"PopStdDevTest1",
			},
		},
		"CqlAggregateTest.xml": XMLTestFileExclusions{
			GroupExcludes: []string{},
			NamesExcludes: []string{
				// TODO: b/342061715 - unsupported operators.
				"RolledOutIntervals",
			},
		},
		"CqlArithmeticFunctionsTest.xml": XMLTestFileExclusions{
			GroupExcludes: []string{
				// TODO: b/342061715 - unsupported operators.
				"HighBoundary",
				"LowBoundary",
			},
			NamesExcludes: []string{
				// TODO: b/342061783 - Got unexpected result.
				"Subtract2And11D",
				"TruncatedDivide10d1ByNeg3D1Quantity",
				"PrecisionDecimal", // Does not yet support trailing zeros.
				// TODO: b/344002938 - xml test is wrong, asserts with a time zone.
				"DateTimeMinValue",
				"DateTimeMaxValue",
			},
		},
		"CqlComparisonOperatorsTest.xml": XMLTestFileExclusions{
			GroupExcludes: []string{},
			NamesExcludes: []string{
				// TODO: b/342061715 - Unsupported operator.
				"BetweenIntTrue",
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
		"CqlDateTimeOperatorsTest.xml": XMLTestFileExclusions{
			GroupExcludes: []string{
				// TODO: b/342061715 - unsupported operators.
				"Duration",
				// TODO: b/342064491 - runtime error: invalid memory address or nil pointer dereference.
				"SameAs",
			},
			NamesExcludes: []string{
				// TODO: b/342061715 - unsupported operators.
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
				"DateTimeDurationBetweenUncertainInterval",
				"DateTimeDurationBetweenUncertainInterval2",
				"DateTimeDurationBetweenUncertainAdd",
				"DateTimeDurationBetweenUncertainSubtract",
				"DateTimeDurationBetweenUncertainMultiply",
				"DateTimeDurationBetweenUncertainDiv",
				"DateTimeDurationBetweenMonthUncertain",
				"DateTimeDurationBetweenMonthUncertain2",
				"DateTimeDurationBetweenMonthUncertain3",
				"DateTimeDurationBetweenMonthUncertain4",
				"DateTimeDurationBetweenMonthUncertain5",
				"DateTimeDurationBetweenMonthUncertain6",
				"DateTimeDurationBetweenMonthUncertain7",
				"DurationInYears",
				"DurationInWeeks",
				"DurationInWeeks2",
				"DurationInWeeks3",
				"TimeDurationBetweenHour",
				"TimeDurationBetweenHourDiffPrecision",
				"TimeDurationBetweenMinute",
				"TimeDurationBetweenSecond",
				"TimeDurationBetweenMillis",
				"DurationInHoursA",
				"DurationInMinutesA",
				"DurationInDaysA",
				"DurationInHoursAA",
				"DurationInMinutesAA",
				"DurationInDaysAA",
				"DateTimeDifferenceUncertain",
				// TODO: b/343800835 - Error in output date comparison based on execution timestamp logic.
				"DateTimeComponentFromDate",
				// TODO: b/342061783 - Got unexpected result.
				"DateTimeAddLeapYear",
			},
		},
		"CqlIntervalOperatorsTest.xml": XMLTestFileExclusions{
			GroupExcludes: []string{
				// TODO: b/342061715 - unsupported operators.
				"After",
				"Before",
				"Collapse",
				"Expand",
				"Ends",
				"Except",
				"Includes",
				"Included In",
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
				"Width",
			},
			NamesExcludes: []string{
				// TODO: b/342061715 - unsupported operators.
				// Note: overlaps before and after are not supported. but these tests are missing the
				// before/after keyword for Date/Time test cases so they are not excluded.
				"TestOverlapsNull",
				"IntegerIntervalOverlapsTrue",
				"IntegerIntervalOverlapsFalse",
				"DecimalIntervalOverlapsTrue",
				"DecimalIntervalOverlapsFalse",
				"QuantityIntervalOverlapsTrue",
				"QuantityIntervalOverlapsFalse",
				"TestOverlapsBeforeNull",
				"IntegerIntervalOverlapsBeforeTrue",
				"IntegerIntervalOverlapsBeforeFalse",
				"DecimalIntervalOverlapsBeforeTrue",
				"DecimalIntervalOverlapsBeforeFalse",
				"QuantityIntervalOverlapsBeforeTrue",
				"QuantityIntervalOverlapsBeforeFalse",
				"TestOverlapsAfterNull",
				"IntegerIntervalOverlapsAfterTrue",
				"IntegerIntervalOverlapsAfterFalse",
				"DecimalIntervalOverlapsAfterTrue",
				"DecimalIntervalOverlapsAfterFalse",
				"QuantityIntervalOverlapsAfterTrue",
				"QuantityIntervalOverlapsAfterFalse",
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
				"TestNullElement1",
				"TestEqualNull",
				"TestInNullBoundaries",
			},
		},
		"CqlListOperatorsTest.xml": XMLTestFileExclusions{
			GroupExcludes: []string{
				// TODO: b/342061715 - unsupported operators.
				"Descendents",
				"Union",
			},
			NamesExcludes: []string{
				// TODO: b/342061715 - unsupported operator.
				"ContainsNullLeft",
				"In1Null",
				"EquivalentABCAnd123",
				"Equivalent123AndABC",
				"Equivalent123AndString123",
				"EquivalentTimeTrue",
				"EquivalentTimeFalse",
				"IncludesNullLeft",
				"IncludesNullRight",
				"IncludedInNullLeft",
				"IncludedInNullRight",
				"NotEqualABCAnd123",
				"NotEqual123AndABC",
				"NotEqual123AndString123",
				"ProperContainsNullRightFalse",
				"ProperContainsNullRightTrue",
				"ProperContainsTimeNull",
				"ProperInTimeNull",
				"ProperIncludedInEmptyAndEmpty",
				"ProperIncludedInListNullAndListNull",
				"ProperIncludedInDateTimeFalse",
				// TODO: b/342061783 - Got unexpected result.
				"EqualNullNull",
				// TODO: b/346880550 - These test appear to have incorrect assertions.
				"Except23And1234",
				"ProperInNullRightFalse",
				"ProperInNullRightTrue",
				"ProperlyIncludedInNulRight",
				"ProperlyIncludesNullLeft",
				"SkipAll",
				"TailOneElement",
				"TakeNullEmpty",
				"TakeEmpty",
			},
		},
		"CqlQueryTests.xml": XMLTestFileExclusions{
			GroupExcludes: []string{},
			NamesExcludes: []string{},
		},
		"CqlNullologicalOperatorsTest.xml": XMLTestFileExclusions{
			GroupExcludes: []string{},
			NamesExcludes: []string{},
		},
		"CqlStringOperatorsTest.xml": XMLTestFileExclusions{
			GroupExcludes: []string{
				// TODO: b/342061715 - unsupported operators.
				"Matches",
				"PositionOf",
				"ReplaceMatches",
				"StartsWith",
				"Substring",
			},
			NamesExcludes: []string{
				// TODO: b/346880550 - These test appear to have incorrect assertions.
				"DateTimeToString1",
				"DateTimeToString2",
				// The spec test is incorrect, fix pending in
				// https://github.com/cqframework/cql-tests/pull/35.
				"CombineEmptyList",
			},
		},
		"CqlTypesTest.xml": XMLTestFileExclusions{
			GroupExcludes: []string{},
			NamesExcludes: []string{
				// TODO: b/342064012 - Uncertain result.
				"DateTimeUncertain",
				// TODO: b/342061715 - unsupported operators.
				"DateTimeTimeUnspecified",
				// TODO: b/343515613 - fails with unexpected result. Technically not supported.
				"StringUnicodeTest",
				// TODO: b/343515819 - fails with unexpected result.
				"StringTestEscapeQuotes",
			},
		},
		"CqlTypeOperatorsTest.xml": XMLTestFileExclusions{
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
			},
		},
		"ValueLiteralsAndSelectors.xml": XMLTestFileExclusions{
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
