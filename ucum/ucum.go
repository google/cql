// Copyright 2025 Google LLC
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

// Package ucum provides UCUM (Unified Code for Units of Measure) support for the CQL engine.
package ucum

import (
	"fmt"
	"strings"
	"sync"
)

// CQLToUCUMDateUnits maps CQL date units to their UCUM equivalents
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

// UCUMToCQLDateUnits maps UCUM date units to their CQL equivalents
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

// Common units and their conversion factors relative to a base unit
var commonUnitFactors = map[string]map[string]float64{
	// Length units (base: meter)
	"m": {
		"cm":  100,
		"mm":  1000,
		"km":  0.001,
		"in":  39.3701,
		"ft":  3.28084,
		"yd":  1.09361,
		"mi":  0.000621371,
	},
	// Mass units (base: gram)
	"g": {
		"mg":   1000,
		"kg":   0.001,
		"lb":   0.00220462,
		"oz":   0.03527396,
		"[oz_av]": 0.03527396,
	},
	// Volume units (base: liter)
	"L": {
		"mL":  1000,
		"dL":  10,
		"cL":  100,
		"kL":  0.001,
		"gal": 0.264172,
		"qt":  1.05669,
		"pt":  2.11338,
		"cup": 4.22675,
		"[foz_us]": 33.814,
	},
	// Time units (base: second)
	"s": {
		"min":  1/60.0,
		"h":    1/3600.0,
		"d":    1/86400.0,
		"wk":   1/604800.0,
		"mo_g": 1/2592000.0, // Approximate
		"a_g":  1/31536000.0,
		"ms":   1000,
	},
}

// Cache for unit validity
var unitValidityCache = struct {
	sync.RWMutex
	cache map[string]bool
}{
	cache: make(map[string]bool),
}

// CheckUnit validates if a unit string is valid UCUM syntax
func CheckUnit(unit string, allowEmptyUnits bool, allowCQLDateUnits bool) (bool, string) {
	// Handle empty units
	if unit == "" {
		if allowEmptyUnits {
			return true, ""
		}
		return false, "empty unit is not allowed"
	}
	
	// Fix empty unit to standard representation
	if allowEmptyUnits {
		unit = FixEmptyUnit(unit)
	}
	
	// Convert CQL date units to UCUM if needed
	if allowCQLDateUnits {
		unit = FixCQLDateUnit(unit)
	}
	
	// Check cache first
	unitValidityCache.RLock()
	valid, found := unitValidityCache.cache[unit]
	unitValidityCache.RUnlock()
	
	if found {
		if valid {
			return true, ""
		}
		return false, fmt.Sprintf("Invalid UCUM unit: '%s'", unit)
	}
	
	// Basic validation: check if this is a known unit or a simple unit
	valid = validateUCUMSyntax(unit)
	
	// Add to cache
	unitValidityCache.Lock()
	unitValidityCache.cache[unit] = valid
	unitValidityCache.Unlock()
	
	if !valid {
		return false, fmt.Sprintf("Invalid UCUM unit: '%s'", unit)
	}
	return true, ""
}

// getConversionFactor determines the conversion factor between two units
func getConversionFactor(fromUnit, toUnit string) (float64, bool) {
	// Handle direct conversions within the same base unit
	for baseUnit, conversions := range commonUnitFactors {
		// Check if from unit is the base
		if fromUnit == baseUnit {
			if factor, ok := conversions[toUnit]; ok {
				return factor, true
			}
		}
		
		// Check if to unit is the base
		if toUnit == baseUnit {
			if factor, ok := conversions[fromUnit]; ok {
				// Inverse conversion
				return 1.0 / factor, true
			}
		}
		
		// Check for conversion between two units of the same type
		if fromFactor, fromOk := conversions[fromUnit]; fromOk {
			if toFactor, toOk := conversions[toUnit]; toOk {
				// Convert via the base unit: fromUnit -> baseUnit -> toUnit
				return toFactor / fromFactor, true
			}
		}
	}
	
	// Special case for UCUM date/time units
	// These need more detailed conversions based on UCUM specifications
	if fromCQLUnit, fromOk := UCUMToCQLDateUnits[fromUnit]; fromOk {
		if toCQLUnit, toOk := UCUMToCQLDateUnits[toUnit]; toOk {
			// Convert between date/time units
			switch {
			case fromCQLUnit == "year" && toCQLUnit == "month":
				return 12.0, true
			case fromCQLUnit == "month" && toCQLUnit == "year":
				return 1.0 / 12.0, true
			case fromCQLUnit == "year" && toCQLUnit == "day":
				return 365.25, true
			case fromCQLUnit == "day" && toCQLUnit == "year":
				return 1.0 / 365.25, true
			case fromCQLUnit == "month" && toCQLUnit == "day":
				return 30.44, true
			case fromCQLUnit == "day" && toCQLUnit == "month":
				return 1.0 / 30.44, true
			case fromCQLUnit == "day" && toCQLUnit == "hour":
				return 24.0, true
			case fromCQLUnit == "hour" && toCQLUnit == "day":
				return 1.0 / 24.0, true
			case fromCQLUnit == "hour" && toCQLUnit == "minute":
				return 60.0, true
			case fromCQLUnit == "minute" && toCQLUnit == "hour":
				return 1.0 / 60.0, true
			case fromCQLUnit == "minute" && toCQLUnit == "second":
				return 60.0, true
			case fromCQLUnit == "second" && toCQLUnit == "minute":
				return 1.0 / 60.0, true
			case fromCQLUnit == "second" && toCQLUnit == "millisecond":
				return 1000.0, true
			case fromCQLUnit == "millisecond" && toCQLUnit == "second":
				return 1.0 / 1000.0, true
			}
		}
	}
	
	return 0, false
}

// ConvertUnit converts a value from one unit to another
func ConvertUnit(fromVal float64, fromUnit, toUnit string) (float64, error) {
	// Fix units
	fromUnit = FixUnit(fromUnit)
	toUnit = FixUnit(toUnit)
	
	// If units are identical, no conversion needed
	if fromUnit == toUnit {
		return fromVal, nil
	}
	
	// Special case for empty unit (dimensionless quantity)
	if fromUnit == "1" && toUnit == "1" {
		return fromVal, nil
	}
	
	// For date/time units, handle special conversions
	if factor, ok := getConversionFactor(fromUnit, toUnit); ok {
		return fromVal * factor, nil
	}
	
	// If we get here, we couldn't find a conversion path
	return 0, fmt.Errorf("cannot convert from '%s' to '%s'", fromUnit, toUnit)
}

// GetProductOfUnits returns the product of two units (for multiplication)
func GetProductOfUnits(unit1, unit2 string) string {
	unit1, unit2 = FixEmptyUnit(unit1), FixEmptyUnit(unit2)
	
	// Handle special cases
	if unit1 == "1" {
		return unit2
	}
	if unit2 == "1" {
		return unit1
	}
	
	// Handle unit squaring (when the same unit is multiplied by itself)
	if unit1 == unit2 {
		// Return the unit with a squared notation (unit2)
		return fmt.Sprintf("%s2", unit1)
	}
	
	// Handle special unit combinations based on UCUM rules
	// TODO: Add more special cases as needed
	
	// For simple units, concatenate with a dot
	return fmt.Sprintf("%s.%s", unit1, unit2)
}

// GetQuotientOfUnits returns the quotient of two units (for division)
func GetQuotientOfUnits(unit1, unit2 string) string {
	unit1, unit2 = FixEmptyUnit(unit1), FixEmptyUnit(unit2)
	
	// Handle special cases
	if unit1 == unit2 {
		return "1"  // Same units cancel out
	}
	if unit2 == "1" {
		return unit1 // Dividing by 1 gives original unit
	}
	
	// For simple units, we can represent with a slash
	// This is a simplified implementation
	return fmt.Sprintf("%s/%s", unit1, unit2)
}

// validateUCUMSyntax provides basic validation of UCUM syntax
func validateUCUMSyntax(unit string) bool {
	// Handle empty unit case
	if unit == "" || unit == "1" {
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
	
	// Default to accepting for now to avoid false negatives
	// A more comprehensive implementation would check against a full UCUM database
	return true
}
