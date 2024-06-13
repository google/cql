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

package datehelpers

import (
	"strings"
	"testing"
	"time"

	"github.com/google/cql/model"
	d4pb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/datatypes_go_proto"
)

func TestParseDate(t *testing.T) {
	evaluationLoc := time.FixedZone("Fixed", 4*60*60)
	tests := []struct {
		name          string
		str           string
		wantTime      time.Time
		wantPrecision model.DateTimePrecision
	}{
		{
			name:          "Year",
			str:           "@2018",
			wantTime:      time.Date(2018, 1, 1, 0, 0, 0, 0, evaluationLoc),
			wantPrecision: model.YEAR,
		},
		{
			name:          "Month",
			str:           "@2018-02",
			wantTime:      time.Date(2018, 2, 1, 0, 0, 0, 0, evaluationLoc),
			wantPrecision: model.MONTH,
		},
		{
			name:          "Day",
			str:           "@2018-02-02",
			wantTime:      time.Date(2018, 2, 2, 0, 0, 0, 0, evaluationLoc),
			wantPrecision: model.DAY,
		},
		{
			name:          "Max date",
			str:           "@9999-12-31",
			wantTime:      time.Date(9999, 12, 31, 0, 0, 0, 0, evaluationLoc),
			wantPrecision: model.DAY,
		},
		{
			name:          "Min date",
			str:           "@0001-01-01",
			wantTime:      time.Date(1, 1, 1, 0, 0, 0, 0, evaluationLoc),
			wantPrecision: model.DAY,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotTime, gotPrecision, err := ParseDate(tc.str, evaluationLoc)
			if err != nil {
				t.Errorf("ParseDate returned unexpected error: %v", err)
			}
			if !gotTime.Equal(tc.wantTime) {
				t.Errorf("ParseDate returned unexpected time: got %v, want %v", gotTime, tc.wantTime)
			}
			if gotPrecision != tc.wantPrecision {
				t.Errorf("ParseDate returned unexpected precision: got %v, want %v", gotPrecision, tc.wantPrecision)
			}
		})
	}
}

func TestParseDate_Error(t *testing.T) {
	evaluationLoc := time.FixedZone("Fixed", 4*60*60)
	tests := []struct {
		name      string
		str       string
		wantError string
	}{
		{
			name:      "Missing @",
			str:       "2018-01-01",
			wantError: "must start with @",
		},
		{
			name:      "Month out of range",
			str:       "@2018-13",
			wantError: `got System.Date @2018-13 but want a layout like @YYYY-MM-DD: month out of range`,
		},
		{
			name:      "Ends with T",
			str:       "@2018-01-01T",
			wantError: `got System.Date @2018-01-01T but want a layout like @YYYY-MM-DD: extra text: "T"`,
		},
		{
			name:      "Does not match format",
			str:       "@2018-01-01T15",
			wantError: `got System.Date @2018-01-01T15 but want a layout like @YYYY-MM-DD: extra text: "T15"`,
		},
		{
			name:      "Also does not match format",
			str:       "@01-01",
			wantError: `got System.Date @01-01 but want a layout like @YYYY-MM-DD`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := ParseDate(tc.str, evaluationLoc)
			if err == nil {
				t.Fatal("ParseDate returned did not return an error")
			}
			if !strings.Contains(err.Error(), tc.wantError) {
				t.Errorf("ParseDate returned unexpected error: got (%v) want (%v)", err, tc.wantError)
			}
		})
	}
}

func TestParseDateTime(t *testing.T) {
	evaluationLoc := time.FixedZone("Fixed", 9*60*60)
	tests := []struct {
		name          string
		str           string
		wantTime      time.Time
		wantPrecision model.DateTimePrecision
	}{
		{
			name:          "Year",
			str:           "@2018T",
			wantTime:      time.Date(2018, 1, 1, 0, 0, 0, 0, evaluationLoc),
			wantPrecision: model.YEAR,
		},
		{
			name:          "Month",
			str:           "@2018-02T",
			wantTime:      time.Date(2018, 2, 1, 0, 0, 0, 0, evaluationLoc),
			wantPrecision: model.MONTH,
		},
		{
			name:          "Day",
			str:           "@2018-02-02T",
			wantTime:      time.Date(2018, 2, 2, 0, 0, 0, 0, evaluationLoc),
			wantPrecision: model.DAY,
		},
		{
			name:          "Hour",
			str:           "@2018-02-02T15",
			wantTime:      time.Date(2018, 2, 2, 15, 0, 0, 0, evaluationLoc),
			wantPrecision: model.HOUR,
		},
		{
			name:          "Minute",
			str:           "@2018-02-02T15:02",
			wantTime:      time.Date(2018, 2, 2, 15, 2, 0, 0, evaluationLoc),
			wantPrecision: model.MINUTE,
		},
		{
			name:          "Second",
			str:           "@2018-02-02T15:02:03",
			wantTime:      time.Date(2018, 2, 2, 15, 2, 3, 0, evaluationLoc),
			wantPrecision: model.SECOND,
		},
		{
			name:          "One Digit Millisecond",
			str:           "@2018-02-02T15:02:03.1",
			wantTime:      time.Date(2018, 2, 2, 15, 2, 3, 100000000, evaluationLoc),
			wantPrecision: model.MILLISECOND,
		},
		{
			name:          "Two Digit Millisecond",
			str:           "@2018-02-02T15:02:03.12",
			wantTime:      time.Date(2018, 2, 2, 15, 2, 3, 120000000, evaluationLoc),
			wantPrecision: model.MILLISECOND,
		},
		{
			name:          "Three Digit Millisecond",
			str:           "@2018-02-02T15:02:03.123",
			wantTime:      time.Date(2018, 2, 2, 15, 2, 3, 123000000, evaluationLoc),
			wantPrecision: model.MILLISECOND,
		},
		{
			name:          "Year with timezone",
			str:           "@2018T-04:00",
			wantTime:      time.Date(2018, 1, 1, 0, 0, 0, 0, time.FixedZone("-04:00", -4*60*60)),
			wantPrecision: model.YEAR,
		},
		{
			name:          "Month with timezone",
			str:           "@2018-02T-04:00",
			wantTime:      time.Date(2018, 2, 1, 0, 0, 0, 0, time.FixedZone("-04:00", -4*60*60)),
			wantPrecision: model.MONTH,
		},
		{
			name:          "Day with timezone",
			str:           "@2018-02-02T-04:00",
			wantTime:      time.Date(2018, 2, 2, 0, 0, 0, 0, time.FixedZone("-04:00", -4*60*60)),
			wantPrecision: model.DAY,
		},
		{
			name:          "Hour with timezone",
			str:           "@2018-02-02T15-04:00",
			wantTime:      time.Date(2018, 2, 2, 15, 0, 0, 0, time.FixedZone("-04:00", -4*60*60)),
			wantPrecision: model.HOUR,
		},
		{
			name:          "Minute with timezone",
			str:           "@2018-02-02T15:02-04:00",
			wantTime:      time.Date(2018, 2, 2, 15, 2, 0, 0, time.FixedZone("-04:00", -4*60*60)),
			wantPrecision: model.MINUTE,
		},

		{
			name:          "Second with timezone",
			str:           "@2018-02-02T15:02:03-04:00",
			wantTime:      time.Date(2018, 2, 2, 15, 2, 3, 0, time.FixedZone("-04:00", -4*60*60)),
			wantPrecision: model.SECOND,
		},
		{
			name:          "Millisecond with timezone",
			str:           "@2018-02-02T15:02:03.004-04:00",
			wantTime:      time.Date(2018, 2, 2, 15, 2, 3, 4000000, time.FixedZone("-04:00", -4*60*60)),
			wantPrecision: model.MILLISECOND,
		},
		{
			name:          "Year with zulu timezone",
			str:           "@2018TZ",
			wantTime:      time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC),
			wantPrecision: model.YEAR,
		},
		{
			name:          "Month with zulu timezone",
			str:           "@2018-02TZ",
			wantTime:      time.Date(2018, 2, 1, 0, 0, 0, 0, time.UTC),
			wantPrecision: model.MONTH,
		},
		{
			name:          "Day with zulu timezone",
			str:           "@2018-02-02TZ",
			wantTime:      time.Date(2018, 2, 2, 0, 0, 0, 0, time.UTC),
			wantPrecision: model.DAY,
		},
		{
			name:          "Hour with zulu timezone",
			str:           "@2018-02-02T15Z",
			wantTime:      time.Date(2018, 2, 2, 15, 0, 0, 0, time.UTC),
			wantPrecision: model.HOUR,
		},
		{
			name:          "Minute with zulu timezone",
			str:           "@2018-02-02T15:02Z",
			wantTime:      time.Date(2018, 2, 2, 15, 2, 0, 0, time.UTC),
			wantPrecision: model.MINUTE,
		},
		{
			name:          "Second with zulu timezone",
			str:           "@2018-02-02T15:02:03Z",
			wantTime:      time.Date(2018, 2, 2, 15, 2, 3, 0, time.UTC),
			wantPrecision: model.SECOND,
		},
		{
			name:          "Millisecond with zulu timezone",
			str:           "@2018-02-02T15:02:03.004Z",
			wantTime:      time.Date(2018, 2, 2, 15, 2, 3, 4000000, time.UTC),
			wantPrecision: model.MILLISECOND,
		},
		{
			name:          "Max datetime",
			str:           "@9999-12-31T23:59:59.999",
			wantTime:      time.Date(9999, 12, 31, 23, 59, 59, 999000000, evaluationLoc),
			wantPrecision: model.MILLISECOND,
		},
		{
			name:          "Min datetime",
			str:           "@0001-01-01T00:00:00.0",
			wantTime:      time.Date(1, 1, 1, 0, 0, 0, 0, evaluationLoc),
			wantPrecision: model.MILLISECOND,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotTime, gotPrecision, err := ParseDateTime(tc.str, evaluationLoc)
			if err != nil {
				t.Errorf("ParseDateTime returned unexpected error: %v", err)
			}
			if !gotTime.Equal(tc.wantTime) {
				t.Errorf("ParseDateTime returned unexpected time: got %v, want %v", gotTime, tc.wantTime)
			}
			if gotPrecision != tc.wantPrecision {
				t.Errorf("ParseDateTime returned unexpected precision: got %v, want %v", gotPrecision, tc.wantPrecision)
			}
		})
	}
}

func TestParseDateTime_Error(t *testing.T) {
	evaluationLoc := time.FixedZone("Fixed", 4*60*60)
	tests := []struct {
		name      string
		str       string
		wantError string
	}{
		{
			name:      "Missing @",
			str:       "2018-02-02T15",
			wantError: "must start with @",
		},
		{
			name:      "Hour out of range",
			str:       "@2018-02-02T29",
			wantError: `got System.DateTime @2018-02-02T29 but want a layout like @YYYY-MM-DDThh:mm:ss.fff(Z|(+/-hh:mm): hour out of range`,
		},
		{
			name:      "Time component without day",
			str:       "@2018-01T15",
			wantError: `got System.DateTime @2018-01T15 but want a layout like @YYYY-MM-DDThh:mm:ss.fff(Z|(+/-hh:mm)`,
		},
		{
			name:      "Too precise",
			str:       "@2018-02-02T15:02:03.004523Z",
			wantError: `System.DateTime @2018-02-02T15:02:03.004523Z can have at most 3 digits of milliseconds precision, want a layout like @YYYY-MM-DDThh:mm:ss.fff(Z|(+/-hh:mm)`,
		},
		{
			name:      "Does not match format",
			str:       "@01-01",
			wantError: `got System.DateTime @01-01 but want a layout like @YYYY-MM-DDThh:mm:ss.fff(Z|(+/-hh:mm)`,
		},
		{
			name:      "Invalid timezone offset",
			str:       "@2018-02-02T15:02:03.004-34:00",
			wantError: `got System.DateTime @2018-02-02T15:02:03.004-34:00 but want a layout like @YYYY-MM-DDThh:mm:ss.fff(Z|(+/-hh:mm): extra text: "-34:00"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := ParseDateTime(tc.str, evaluationLoc)
			if err == nil {
				// TODO: This test cases parses with Go 1.22
				// but fails (correctly) with Go 1.23.
				// Until switching to Go 1.23,
				// ignore this test if it parses.
				// This can be removed after we can assume
				// that all users are at Go 1.23.
				if tc.str == "@2018-02-02T15:02:03.004-34:00" {
					t.Skip("skipping will not start failing until Go 1.23")
				}
				t.Fatal("ParseDateTime returned did not return an error")
			}
			if !strings.Contains(err.Error(), tc.wantError) {
				t.Errorf("ParseDateTime returned unexpected error: got (%v) want (%v)", err, tc.wantError)
			}
		})
	}
}

func TestParseTime(t *testing.T) {
	tests := []struct {
		name          string
		str           string
		wantTime      time.Time
		wantPrecision model.DateTimePrecision
	}{
		{
			name:          "Hour",
			str:           "@T15",
			wantTime:      time.Date(0, 1, 1, 15, 0, 0, 0, time.UTC),
			wantPrecision: model.HOUR,
		},
		{
			name:          "Minute",
			str:           "@T15:02",
			wantTime:      time.Date(0, 1, 1, 15, 2, 0, 0, time.UTC),
			wantPrecision: model.MINUTE,
		},
		{
			name:          "Second",
			str:           "@T15:02:03",
			wantTime:      time.Date(0, 1, 1, 15, 2, 3, 0, time.UTC),
			wantPrecision: model.SECOND,
		},
		{
			name:          "Millisecond",
			str:           "@T15:02:03.004",
			wantTime:      time.Date(0, 1, 1, 15, 2, 3, 4000000, time.UTC),
			wantPrecision: model.MILLISECOND,
		},
		{
			name:          "Max time",
			str:           "@T23:59:59.999",
			wantTime:      time.Date(0, 1, 1, 23, 59, 59, 999000000, time.UTC),
			wantPrecision: model.MILLISECOND,
		},
		{
			name:          "Min time",
			str:           "@T00:00:00.0",
			wantTime:      time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC),
			wantPrecision: model.MILLISECOND,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotTime, gotPrecision, err := ParseTime(tc.str, time.UTC)
			if err != nil {
				t.Errorf("ParseTime returned unexpected error: %v", err)
			}
			if !gotTime.Equal(tc.wantTime) {
				t.Errorf("ParseTime returned unexpected time: got %v, want %v", gotTime, tc.wantTime)
			}
			if gotPrecision != tc.wantPrecision {
				t.Errorf("ParseTime returned unexpected precision: got %v, want %v", gotPrecision, tc.wantPrecision)
			}
		})
	}
}

func TestParseTime_Error(t *testing.T) {
	tests := []struct {
		name      string
		str       string
		wantError string
	}{
		{
			name:      "Missing @",
			str:       "T15",
			wantError: "must start with @",
		},
		{
			name:      "Hour out of range",
			str:       "@T29",
			wantError: `got System.Time @T29 but want a layout like @Thh:mm:ss.fff: hour out of range`,
		},
		{
			name:      "Too precise",
			str:       "@T15:02:03.0045",
			wantError: `System.Time @T15:02:03.0045 can have at most 3 digits of milliseconds precision, want a layout like @Thh:mm:ss.fff`,
		},
		{
			name:      "Has a timezone",
			str:       "@T15:02:03.004Z",
			wantError: `got System.Time @T15:02:03.004Z but want a layout like @Thh:mm:ss.fff: extra text: "Z"`,
		},
		{
			name:      "Does not match format",
			str:       "@01-01",
			wantError: `got System.Time @01-01 but want a layout like @Thh:mm:ss.fff`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := ParseTime(tc.str, time.UTC)
			if err == nil {
				t.Fatal("ParseTime returned did not return an error")
			}
			if !strings.Contains(err.Error(), tc.wantError) {
				t.Errorf("ParseTime returned unexpected error: got (%v) want (%v)", err, tc.wantError)
			}
		})
	}
}

func TestParseFHIRDateTime(t *testing.T) {
	tests := []struct {
		name          string
		dateTime      *d4pb.DateTime
		evaluationLoc *time.Location
		wantTime      time.Time
		wantPrecision model.DateTimePrecision
	}{
		{
			name:          "DateTime in UTC",
			dateTime:      &d4pb.DateTime{ValueUs: 1711936984000000, Precision: d4pb.DateTime_SECOND, Timezone: "UTC"},
			evaluationLoc: time.FixedZone("America/Los_Angeles", -7*60*60),
			wantTime:      time.Date(2024, time.April, 1, 2, 3, 4, 0, time.UTC),
			wantPrecision: model.SECOND,
		},
		{
			name:          "DateTime in America/Los_Angeles",
			dateTime:      &d4pb.DateTime{ValueUs: 1711936984000000, Precision: d4pb.DateTime_SECOND, Timezone: "America/Los_Angeles"},
			evaluationLoc: time.FixedZone("America/Los_Angeles", -7*60*60),
			wantTime:      time.Date(2024, time.March, 31, 19, 3, 4, 0, time.FixedZone("America/Los_Angeles", -7*60*60)),
			wantPrecision: model.SECOND,
		},
		{
			name:          "DateTime using -07:00 offset",
			dateTime:      &d4pb.DateTime{ValueUs: 1711936984000000, Precision: d4pb.DateTime_SECOND, Timezone: "-07:00"},
			evaluationLoc: time.FixedZone("America/Los_Angeles", -7*60*60),
			wantTime:      time.Date(2024, time.March, 31, 19, 3, 4, 0, time.FixedZone("-07:00", -7*60*60)),
			wantPrecision: model.SECOND,
		},
		{
			name:          "DateTime using +07:00 offset",
			dateTime:      &d4pb.DateTime{ValueUs: 1711936984000000, Precision: d4pb.DateTime_SECOND, Timezone: "+07:00"},
			evaluationLoc: time.FixedZone("America/Los_Angeles", -7*60*60),
			wantTime:      time.Date(2024, time.April, 1, 9, 3, 4, 0, time.FixedZone("+07:00", +7*60*60)),
			wantPrecision: model.SECOND,
		},
		{
			name:          "DateTime fallback to evaluation location",
			dateTime:      &d4pb.DateTime{ValueUs: 1711936984000000, Precision: d4pb.DateTime_SECOND, Timezone: ""},
			evaluationLoc: time.FixedZone("America/Los_Angeles", -7*60*60),
			wantTime:      time.Date(2024, time.March, 31, 19, 3, 4, 0, time.FixedZone("America/Los_Angeles", -7*60*60)),
			wantPrecision: model.SECOND,
		},
		{
			name:          "DateTime Millisecond precision",
			dateTime:      &d4pb.DateTime{ValueUs: 1711936984500000, Precision: d4pb.DateTime_MILLISECOND, Timezone: "UTC"},
			evaluationLoc: time.FixedZone("America/Los_Angeles", -7*60*60),
			wantTime:      time.Date(2024, time.April, 1, 2, 3, 4, 5e8, time.UTC),
			wantPrecision: model.MILLISECOND,
		},
		{
			name:          "DateTime Microsecond maps to Millisecond precision",
			dateTime:      &d4pb.DateTime{ValueUs: 1711936984500000, Precision: d4pb.DateTime_MICROSECOND, Timezone: "UTC"},
			evaluationLoc: time.FixedZone("America/Los_Angeles", -7*60*60),
			wantTime:      time.Date(2024, time.April, 1, 2, 3, 4, 5e8, time.UTC),
			wantPrecision: model.MILLISECOND,
		},
		{
			name:          "DateTime Day precision",
			dateTime:      &d4pb.DateTime{ValueUs: 1711936984500000, Precision: d4pb.DateTime_DAY, Timezone: "UTC"},
			evaluationLoc: time.FixedZone("America/Los_Angeles", -7*60*60),
			wantTime:      time.Date(2024, time.April, 1, 2, 3, 4, 5e8, time.UTC),
			wantPrecision: model.DAY,
		},
		{
			name:          "DateTime Month precision",
			dateTime:      &d4pb.DateTime{ValueUs: 1711936984500000, Precision: d4pb.DateTime_MONTH, Timezone: "UTC"},
			evaluationLoc: time.FixedZone("America/Los_Angeles", -7*60*60),
			wantTime:      time.Date(2024, time.April, 1, 2, 3, 4, 5e8, time.UTC),
			wantPrecision: model.MONTH,
		},
		{
			name:          "DateTime Year precision",
			dateTime:      &d4pb.DateTime{ValueUs: 1711936984500000, Precision: d4pb.DateTime_YEAR, Timezone: "UTC"},
			evaluationLoc: time.FixedZone("America/Los_Angeles", -7*60*60),
			wantTime:      time.Date(2024, time.April, 1, 2, 3, 4, 5e8, time.UTC),
			wantPrecision: model.YEAR,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotTime, gotPrecision, err := ParseFHIRDateTime(tc.dateTime, tc.evaluationLoc)
			if err != nil {
				t.Errorf("ParseFHIRDateTime returned unexpected error: %v", err)
			}
			if !gotTime.Equal(tc.wantTime) {
				t.Errorf("ParseFHIRDateTime returned unexpected time: got %v, want %v", gotTime, tc.wantTime)
			}
			if gotPrecision != tc.wantPrecision {
				t.Errorf("ParseFHIRDateTime returned unexpected precision: got %v, want %v", gotPrecision, tc.wantPrecision)
			}
		})
	}
}

func TestParseFHIRDateTime_NoEvalTimestampError(t *testing.T) {
	_, _, err := ParseFHIRDateTime(&d4pb.DateTime{ValueUs: 1711929600000000, Precision: d4pb.DateTime_SECOND, Timezone: ""}, nil)
	if err == nil {
		t.Fatal("ParseFHIRDateTime returned did not return an error")
	}
	if !strings.Contains(err.Error(), "internal error - evaluationLoc must be set when calling ParseFHIRDateTime") {
		t.Errorf("ParseFHIRDateTime returned unexpected error: got (%v) want (%v)", err, "evaluation location")
	}
}

func TestParseFHIRDate(t *testing.T) {
	tests := []struct {
		name          string
		date          *d4pb.Date
		evaluationLoc *time.Location
		wantTime      time.Time
		wantPrecision model.DateTimePrecision
	}{
		{
			name:          "Date with Day precision",
			date:          &d4pb.Date{ValueUs: 1711929600000000, Precision: d4pb.Date_DAY, Timezone: "UTC"},
			evaluationLoc: time.FixedZone("America/Los_Angeles", -7*60*60),
			wantTime:      time.Date(2024, time.April, 1, 0, 0, 0, 0, time.FixedZone("America/Los_Angeles", -7*60*60)),
			wantPrecision: model.DAY,
		},
		{
			name:          "Date with offset based timzeone in the proto",
			date:          &d4pb.Date{ValueUs: 1711929600000000, Precision: d4pb.Date_DAY, Timezone: "-07:00"},
			evaluationLoc: time.FixedZone("America/Los_Angeles", -7*60*60),
			// The date corresponding to ValueUS in -7:00, is actually March 31, 2024. In the previous
			// test case, the date in UTC for the same ValueUS timestamp is April 1 already.
			wantTime:      time.Date(2024, time.March, 31, 0, 0, 0, 0, time.FixedZone("America/Los_Angeles", -7*60*60)),
			wantPrecision: model.DAY,
		},
		{
			name:          "Date with missing timezone in proto falls back to evaluationLoc",
			date:          &d4pb.Date{ValueUs: 1711929600000000, Precision: d4pb.Date_DAY, Timezone: ""},
			evaluationLoc: time.FixedZone("America/Los_Angeles", -7*60*60),
			wantTime:      time.Date(2024, time.March, 31, 0, 0, 0, 0, time.FixedZone("America/Los_Angeles", -7*60*60)),
			wantPrecision: model.DAY,
		},
		{
			name:          "Date with Month precision",
			date:          &d4pb.Date{ValueUs: 1711929600000000, Precision: d4pb.Date_MONTH, Timezone: "UTC"},
			evaluationLoc: time.FixedZone("America/Los_Angeles", -7*60*60),
			wantTime:      time.Date(2024, time.April, 1, 0, 0, 0, 0, time.FixedZone("America/Los_Angeles", -7*60*60)),
			wantPrecision: model.MONTH,
		},
		{
			name:          "Date with Year precision",
			date:          &d4pb.Date{ValueUs: 1711929600000000, Precision: d4pb.Date_YEAR, Timezone: "UTC"},
			evaluationLoc: time.FixedZone("America/Los_Angeles", -7*60*60),
			wantTime:      time.Date(2024, time.April, 1, 0, 0, 0, 0, time.FixedZone("America/Los_Angeles", -7*60*60)),
			wantPrecision: model.YEAR,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotTime, gotPrecision, err := ParseFHIRDate(tc.date, tc.evaluationLoc)
			if err != nil {
				t.Errorf("ParseFHIRDate returned unexpected error: %v", err)
			}
			if !gotTime.Equal(tc.wantTime) {
				t.Errorf("ParseFHIRDate returned unexpected time: got %v, want %v", gotTime, tc.wantTime)
			}
			if gotPrecision != tc.wantPrecision {
				t.Errorf("ParseFHIRDate returned unexpected precision: got %v, want %v", gotPrecision, tc.wantPrecision)
			}
		})
	}
}
