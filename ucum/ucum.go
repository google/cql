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

// Package ucum provides an API for validation and conversion of measurement units using the UCUM
// (Unified Code for Units of Measure) standard.
//
// https://unitsofmeasure.org/ucum
// https://cql.hl7.org/02-authorsguide.html#quantities
// The javascript engine also references this package which outlines some of the UCUM syntax:
// https://github.com/LHNCBC/ucum-lhc/blob/master/data/ucum.csv
package ucum

import (
	"fmt"
)

// ConvertUnit converts a value from one unit to another.
func ConvertUnit(fromVal float64, fromUnit, toUnit string) (float64, error) {
	fromUnit, toUnit = normalizeUnit(fromUnit), normalizeUnit(toUnit)

	// If units are identical, no conversion needed. This also handles the special case of empty units
	// (both units are "1").
	if fromUnit == toUnit {
		return fromVal, nil
	}
	if ok, factor := getConversionFactor(fromUnit, toUnit); ok {
		return fromVal * factor, nil
	}

	// A conversion path could not be found.
	return 0, fmt.Errorf("cannot convert from '%s' to '%s'", fromUnit, toUnit)
}

// GetProductOfUnits returns the product of two units.
// The UCUM standard specifies products as a '.' concatenation of the two units.
// https://unitsofmeasure.org/ucum
func GetProductOfUnits(unit1, unit2 string) string {
	unit1, unit2 = normalizeEmptyUnit(unit1), normalizeEmptyUnit(unit2)

	// Special cases when one of the units is "1".
	if unit1 == oneUnit {
		return unit2
	}
	if unit2 == oneUnit {
		return unit1
	}
	// Handle unit squaring; a unit that is multiplied by itself.
	if unit1 == unit2 {
		return fmt.Sprintf("%s2", unit1)
	}
	// For simple units, concatenate with a dot.
	return fmt.Sprintf("%s.%s", unit1, unit2)
}

// GetQuotientOfUnits returns the quotient of two units.
// The UCUM standard specifies quotients as a '/' concatenation of the two units.
// https://unitsofmeasure.org/ucum
func GetQuotientOfUnits(unit1, unit2 string) string {
	unit1, unit2 = normalizeEmptyUnit(unit1), normalizeEmptyUnit(unit2)

	// Handle special cases
	if unit1 == unit2 {
		return oneUnit // Same units cancel out
	}
	if unit2 == oneUnit {
		return unit1 // Dividing by 1 gives original unit
	}
	// For simple units, this is represented with a slash.
	return fmt.Sprintf("%s/%s", unit1, unit2)
}

// ValidateUnit validates if a unit string is valid UCUM syntax.
func ValidateUnit(unit string, allowEmptyUnits bool, allowCQLDateUnits bool) (bool, string) {
	// Handle empty units.
	if unit == "" {
		if allowEmptyUnits {
			return true, ""
		}
		return false, "empty unit is not allowed"
	}

	// Fix empty unit to standard representation.
	if allowEmptyUnits {
		unit = normalizeEmptyUnit(unit)
	}

	// Convert CQL date units to UCUM if needed.
	if allowCQLDateUnits {
		unit = normalizeCQLDateUnit(unit)
	}

	// Check cache first.
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
