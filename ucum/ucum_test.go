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

package ucum

import (
	"math"
	"testing"
)

// almostEqual checks if two float64 values are approximately equal within a small tolerance
func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-10
}

func TestConvertUnit(t *testing.T) {
	tests := []struct {
		name     string
		fromVal  float64
		fromUnit string
		toUnit   string
		wantVal  float64
	}{
		{"m to cm", 1.0, "m", "cm", 100.0},
		{"cm to m", 100.0, "cm", "m", 1.0},
		{"kg to g", 1.0, "kg", "g", 1000.0},
		{"g to kg", 1000.0, "g", "kg", 1.0},
		{"same units", 1.0, "m", "m", 1.0},
		{"empty units", 1.0, "", "", 1.0}, // "" is treated as "1"

		// Clinical mass units
		{"mg to g", 1000.0, "mg", "g", 1.0},
		{"g to mg", 1.0, "g", "mg", 1000.0},
		{"mg to kg", 1000000.0, "mg", "kg", 1.0},
		{"kg to mg", 1.0, "kg", "mg", 1000000.0},

		// Clinical volume units
		{"mL to L", 1000.0, "mL", "L", 1.0},
		{"L to mL", 1.0, "L", "mL", 1000.0},
		{"mL to dL", 100.0, "mL", "dL", 1.0},
		{"dL to mL", 1.0, "dL", "mL", 100.0},

		// Clinical enzyme units
		{"U to mU", 1.0, "U", "mU", 1000.0},
		{"mU to U", 1000.0, "mU", "U", 1.0},
		{"U to uU", 1.0, "U", "uU", 1000000.0},
		{"uU to U", 1000000.0, "uU", "U", 1.0},
		{"U to nU", 1.0, "U", "nU", 1000000000.0},
		{"nU to U", 1000000000.0, "nU", "U", 1.0},
		{"U to kU", 1000.0, "U", "kU", 1.0},
		{"kU to U", 1.0, "kU", "U", 1000.0},

		// Clinical osmolality units
		{"osm to mosm", 1.0, "osm", "mosm", 1000.0},
		{"mosm to osm", 1000.0, "mosm", "osm", 1.0},

		// Clinical equivalents
		{"eq to meq", 1.0, "eq", "meq", 1000.0},
		{"meq to eq", 1000.0, "meq", "eq", 1.0},
		{"eq to ueq", 1.0, "eq", "ueq", 1000000.0},
		{"ueq to eq", 1000000.0, "ueq", "eq", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := ConvertUnit(tt.fromVal, tt.fromUnit, tt.toUnit)
			if err != nil {
				t.Errorf("ConvertUnit(%v, %q, %q) failed with error = %v", tt.fromVal, tt.fromUnit, tt.toUnit, err)
			}
			if !almostEqual(val, tt.wantVal) {
				t.Errorf("ConvertUnit(%v, %q, %q) val = %v, want %v", tt.fromVal, tt.fromUnit, tt.toUnit, val, tt.wantVal)
			}
		})
	}
}

func TestConvertUnitError(t *testing.T) {
	tests := []struct {
		name       string
		fromVal    float64
		fromUnit   string
		toUnit     string
		wantErrMsg string
	}{
		{"invalid fromUnit", 1.0, "invalid", "m", "cannot convert from 'invalid' to 'm'"},
		{"invalid toUnit", 1.0, "m", "invalid", "cannot convert from 'm' to 'invalid'"},
		{"incompatible units", 1.0, "m", "kg", "cannot convert from 'm' to 'kg'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ConvertUnit(tt.fromVal, tt.fromUnit, tt.toUnit)
			if err == nil {
				t.Errorf("ConvertUnit(%v, %q, %q) succeeded when it should have failed", tt.fromVal, tt.fromUnit, tt.toUnit)
			}
			if err.Error() != tt.wantErrMsg {
				t.Errorf("ConvertUnit(%v, %q, %q) want error message = %v, got %v", tt.fromVal, tt.fromUnit, tt.toUnit, tt.wantErrMsg, err.Error())
			}
		})
	}
}

func TestGetProductOfUnits(t *testing.T) {
	tests := []struct {
		name  string
		unit1 string
		unit2 string
		want  string
	}{
		{"m and m", "m", "m", "m2"},
		{"m and s", "m", "s", "m.s"},
		{"kg and m/s2", "kg", "m/s2", "kg.m/s2"},
		{"unit1 is 1", "1", "m", "m"},
		{"unit2 is 1", "s", "1", "s"},
		{"both units are 1", "1", "1", "1"},
		{"unit1 is empty", "", "m", "m"},
		{"unit2 is empty", "s", "", "s"},
		{"both units are empty", "", "", "1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetProductOfUnits(tt.unit1, tt.unit2)
			if got != tt.want {
				t.Errorf("GetProductOfUnits(%q, %q) = %q, want %q", tt.unit1, tt.unit2, got, tt.want)
			}
		})
	}
}

func TestGetQuotientOfUnits(t *testing.T) {
	tests := []struct {
		name  string
		unit1 string
		unit2 string
		want  string
	}{
		{"m and s", "m", "s", "m/s"},
		{"m and m", "m", "m", "1"},
		{"m/s and s", "m/s", "s", "m/s/s"}, // This might need more sophisticated simplification logic
		{"unit1 is 1", "1", "s", "1/s"},
		{"unit2 is 1", "m", "1", "m"},
		{"both units are 1", "1", "1", "1"},
		{"unit1 is empty", "", "s", "1/s"},
		{"unit2 is empty", "m", "", "m"},
		{"both units are empty", "", "", "1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetQuotientOfUnits(tt.unit1, tt.unit2)
			if got != tt.want {
				t.Errorf("GetQuotientOfUnits(%q, %q) = %q, want %q", tt.unit1, tt.unit2, got, tt.want)
			}
		})
	}
}

func TestValidateUnit(t *testing.T) {
	tests := []struct {
		name              string
		unit              string
		allowEmptyUnits   bool
		allowCQLDateUnits bool
	}{
		{"valid unit", "m", false, false},
		{"valid unit with empty allowed", "m", true, false},
		{"valid unit with cql date units allowed", "m", false, true},
		{"empty unit when allowed", "", true, false},
		{"cql date unit year when allowed", "year", false, true},
		{"cql date unit month when allowed", "month", false, true},
		{"cql date unit day when allowed", "day", false, true},
		{"cql date unit hour when allowed", "hour", false, true},
		{"cql date unit minute when allowed", "minute", false, true},
		{"cql date unit second when allowed", "second", false, true},
		{"cql date unit millisecond when allowed", "millisecond", false, true},

		// Clinical units
		{"clinical mass unit mg", "mg", false, false},
		{"clinical volume unit mL", "mL", false, false},
		{"clinical enzyme unit U", "U", false, false},
		{"clinical enzyme unit mU", "mU", false, false},
		{"clinical enzyme unit uU", "uU", false, false},
		{"clinical enzyme unit nU", "nU", false, false},
		{"clinical enzyme unit kU", "kU", false, false},
		{"clinical osmolality unit osm", "osm", false, false},
		{"clinical osmolality unit mosm", "mosm", false, false},
		{"clinical equivalent unit eq", "eq", false, false},
		{"clinical equivalent unit meq", "meq", false, false},
		{"clinical equivalent unit ueq", "ueq", false, false},

		// Common clinical compound units
		{"compound unit mg/dL", "mg/dL", false, false},
		{"compound unit mg/mL", "mg/mL", false, false},
		{"compound unit U/L", "U/L", false, false},
		{"compound unit mU/mL", "mU/mL", false, false},
		{"compound unit mg/L", "mg/L", false, false},
		{"compound unit mL/h", "mL/h", false, false},
		{"compound unit mg/h", "mg/h", false, false},
		{"compound unit U/mL", "U/mL", false, false},
		{"compound unit mosm/L", "mosm/L", false, false},
		{"compound unit meq/L", "meq/L", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, message := ValidateUnit(tt.unit, tt.allowEmptyUnits, tt.allowCQLDateUnits)
			if !valid {
				t.Errorf("CheckUnit(%q, %v, %v) failed when it should have succeeded, got message = %v", tt.unit, tt.allowEmptyUnits, tt.allowCQLDateUnits, message)
			}
		})
	}
}

func TestValidateUnit_Error(t *testing.T) {
	tests := []struct {
		name              string
		unit              string
		allowEmptyUnits   bool
		allowCQLDateUnits bool
		wantMessage       string
	}{
		{"invalid unit", "invalid", false, false, "Invalid UCUM unit: 'invalid'"},
		{"empty unit when not allowed", "", false, false, "empty unit is not allowed"},
		{"cql date unit year when not allowed", "year", false, false, "Invalid UCUM unit: 'year'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, message := ValidateUnit(tt.unit, tt.allowEmptyUnits, tt.allowCQLDateUnits)
			if valid {
				t.Errorf("CheckUnit(%q, %v, %v) succeeded when it should have failed, want message = %v", tt.unit, tt.allowEmptyUnits, tt.allowCQLDateUnits, tt.wantMessage)
			}
			if message != tt.wantMessage {
				t.Errorf("CheckUnit(%q, %v, %v) message = %q, want %q", tt.unit, tt.allowEmptyUnits, tt.allowCQLDateUnits, message, tt.wantMessage)
			}
		})
	}
}
