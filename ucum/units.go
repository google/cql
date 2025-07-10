// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ucum

import (
	"strings"
	"sync"
)

const (
	dateYearUnit        = "year"
	dateMonthUnit       = "month"
	dateWeekUnit        = "week"
	dateDayUnit         = "day"
	dateHourUnit        = "hour"
	dateMinuteUnit      = "minute"
	dateSecondUnit      = "second"
	dateMillisecondUnit = "millisecond"
	// oneUnit is the dimensionless unit "1".
	oneUnit = "1"
)

// CQLToUCUMDateUnits maps CQL date units to their UCUM equivalents.
var CQLToUCUMDateUnits = map[string]string{
	"years":        "a_g",
	"year":         "a_g",
	"months":       "mo_g",
	"month":        "mo_g",
	"weeks":        "wk",
	"week":         "wk",
	"days":         "d",
	"day":          "d",
	"hours":        "h",
	"hour":         "h",
	"minutes":      "min",
	"minute":       "min",
	"seconds":      "s",
	"second":       "s",
	"milliseconds": "ms",
	"millisecond":  "ms",
}

// UCUMToCQLDateUnits maps UCUM date units to their CQL equivalents.
var UCUMToCQLDateUnits = map[string]string{
	"a":    "year",
	"a_j":  "year",
	"a_g":  "year",
	"mo":   "month",
	"mo_j": "month",
	"mo_g": "month",
	"wk":   "week",
	"d":    "day",
	"h":    "hour",
	"min":  "minute",
	"s":    "second",
	"ms":   "millisecond",
}

// Common units and their conversion factors relative to a base unit.
var commonUnitFactors = map[string]map[string]float64{
	// Length units (base: meter)
	"m": {
		"cm": 100,
		"mm": 1000,
		"km": 0.001,
		"in": 39.3701,
		"ft": 3.28084,
		"yd": 1.09361,
		"mi": 0.000621371,
	},
	// Mass units (base: gram)
	"g": {
		"mg":      1000,
		"kg":      0.001,
		"lb":      0.00220462,
		"oz":      0.03527396,
		"[oz_av]": 0.03527396,
	},
	// Volume units (base: liter)
	"L": {
		"mL":       1000,
		"dL":       10,
		"cL":       100,
		"kL":       0.001,
		"gal":      0.264172,
		"qt":       1.05669,
		"pt":       2.11338,
		"cup":      4.22675,
		"[foz_us]": 33.814,
	},
	// Time units (base: second)
	"s": {
		"min":  1 / 60.0,
		"h":    1 / 3600.0,
		"d":    1 / 86400.0,
		"wk":   1 / 604800.0,
		"mo_g": 1 / 2592000.0, // Approximate
		"a_g":  1 / 31536000.0,
		"ms":   1000,
	},
	// Clinical/Enzyme units (base: enzyme unit U)
	// 1 U = 1 micromole substrate catalyzed per minute (1 umol/min)
	"U": {
		"mU": 1000,       // milli enzyme units (1 mU = 1 nmol/min)
		"uU": 1000000,    // micro enzyme units (1 uU = 1 pmol/min)
		"nU": 1000000000, // nano enzyme units (1 nU = 1 fmol/min)
		"kU": 0.001,      // kilo enzyme units
	},
	// Clinical osmolality units (base: osmole)
	"osm": {
		"mosm": 1000, // milliosmole
	},
	// Clinical equivalents (base: equivalent)
	"eq": {
		"meq": 1000,    // milliequivalent
		"ueq": 1000000, // microequivalent
	},
}

// Cache of already validated units.
var unitValidityCache = struct {
	sync.RWMutex
	cache map[string]bool
}{
	cache: make(map[string]bool),
}

// Functions for normalizing units and retrieving conversion factors between units.

// getConversionFactor determines the conversion factor between two units.
func getConversionFactor(fromUnit, toUnit string) (bool, float64) {
	// Handle direct conversions within the same base unit.
	for baseUnit, conversions := range commonUnitFactors {
		if ok, factor := getMeasurementConversionFactor(fromUnit, toUnit, baseUnit, conversions); ok {
			return true, factor
		}
	}

	// Since we didn't find a conversion path, check for the special case for UCUM date/time units.
	ok, dateFactor := getDateConversionFactor(fromUnit, toUnit)
	if ok {
		return true, dateFactor
	}

	return false, 0
}

func getMeasurementConversionFactor(fromUnit, toUnit, baseUnit string, conversions map[string]float64) (bool, float64) {
	toFactor, ok := conversions[toUnit]
	if !ok && toUnit != baseUnit {
		return false, 0
	}
	// If the from unit is the base unit, we can return the to factor directly.
	if fromUnit == baseUnit {
		return true, toFactor
	}
	fromFactor, ok := conversions[fromUnit]
	if !ok && fromUnit != baseUnit {
		return false, 0
	}
	if toUnit == baseUnit {
		return true, 1.0 / fromFactor
	}
	// Convert via the base unit: fromUnit -> baseUnit -> toUnit
	return true, toFactor / fromFactor
}

// getDateConversionFactor determines the conversion factor between two date/time units.
func getDateConversionFactor(fromUnit, toUnit string) (bool, float64) {
	if ok, factor := getDateConversionFactorLowerPrecision(fromUnit, toUnit); ok {
		return true, factor
	}
	if ok, factor := getDateConversionFactorLowerPrecision(toUnit, fromUnit); ok {
		return true, 1.0 / factor
	}
	return false, 0
}

// getDateConversionFactorLowerPrecision determines finds a conversion factor between two units when
// the left unit has a higher precision than the right unit.
func getDateConversionFactorLowerPrecision(fromUnit, toUnit string) (bool, float64) {
	fromCQLUnit, fromOk := UCUMToCQLDateUnits[fromUnit]
	if !fromOk {
		return false, 0
	}
	toCQLUnit, toOk := UCUMToCQLDateUnits[toUnit]
	if !toOk {
		return false, 0
	}
	if fromCQLUnit == toCQLUnit {
		return true, 1.0
	}
	switch {
	case fromCQLUnit == dateYearUnit && toCQLUnit == dateMonthUnit:
		return true, 12.0
	case fromCQLUnit == dateYearUnit && toCQLUnit == dateDayUnit:
		return true, 365.25
	case fromCQLUnit == dateMonthUnit && toCQLUnit == dateDayUnit:
		return true, 30.44
	case fromCQLUnit == dateDayUnit && toCQLUnit == dateHourUnit:
		return true, 24.0
	case fromCQLUnit == dateHourUnit && toCQLUnit == dateMinuteUnit:
		return true, 60.0
	case fromCQLUnit == dateHourUnit && toCQLUnit == dateSecondUnit:
		return true, 3600.0
	case fromCQLUnit == dateHourUnit && toCQLUnit == dateMillisecondUnit:
		return true, 3600000.0
	case fromCQLUnit == dateMinuteUnit && toCQLUnit == dateSecondUnit:
		return true, 60.0
	case fromCQLUnit == dateMinuteUnit && toCQLUnit == dateMillisecondUnit:
		return true, 60000.0
	case fromCQLUnit == dateSecondUnit && toCQLUnit == dateMillisecondUnit:
		return true, 1000.0
	default:
		return false, 0
	}
}

// normalizeEmptyUnit handles null/empty units, replacing them with "1" (dimensionless unit).
func normalizeEmptyUnit(unit string) string {
	if unit == "" {
		return oneUnit
	}
	return unit
}

// normalizeCQLDateUnit converts CQL date units to UCUM format.
func normalizeCQLDateUnit(unit string) string {
	if ucumUnit, ok := CQLToUCUMDateUnits[unit]; ok {
		return ucumUnit
	}
	return unit
}

// normalizeUnit applies both empty and date unit fixes.
func normalizeUnit(unit string) string {
	return normalizeCQLDateUnit(normalizeEmptyUnit(unit))
}

// validateUCUMSyntax provides basic validation of UCUM syntax.
func validateUCUMSyntax(unit string) bool {
	// Handle empty unit case
	if unit == "" || unit == oneUnit {
		return true
	}

	// Check if it's a known unit from our conversion tables
	for baseUnit, factors := range commonUnitFactors {
		if unit == baseUnit {
			return true
		}
		for derivedUnit := range factors {
			if unit == derivedUnit {
				return true
			}
		}
	}

	// Check if it's a known CQL date unit in UCUM format
	for _, ucumUnit := range CQLToUCUMDateUnits {
		if unit == ucumUnit {
			return true
		}
	}

	// Basic syntax check for common patterns
	if strings.Contains(unit, "/") {
		// Handle division format (e.g., "m/s")
		parts := strings.Split(unit, "/")
		if len(parts) == 2 {
			return validateUCUMSyntax(parts[0]) && validateUCUMSyntax(parts[1])
		}
	}

	if strings.Contains(unit, ".") {
		// Handle multiplication format (e.g., "N.m")
		parts := strings.Split(unit, ".")
		for _, part := range parts {
			if !validateUCUMSyntax(part) {
				return false
			}
		}
		return true
	}

	// Check for common UCUM suffixes like -2, 2, etc.
	// This is a very simplified check
	if len(unit) > 1 {
		lastChar := unit[len(unit)-1]
		if lastChar >= '0' && lastChar <= '9' {
			baseUnit := unit[:len(unit)-1]
			return validateUCUMSyntax(baseUnit)
		}
	}

	return false
}
