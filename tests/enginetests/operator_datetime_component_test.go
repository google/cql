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

package enginetests

import (
	"context"
	"testing"

	"github.com/google/cql/interpreter"
	"github.com/google/cql/parser"
	"github.com/google/cql/result"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestDateTimeComponentFrom(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantResult result.Value
		wantErr    bool
	}{
		// Year extraction tests
		{
			name:       "Year from Date",
			cql:        "year from @2020-03-15",
			wantResult: newOrFatal(t, int32(2020)),
		},
		{
			name:       "Year from DateTime",
			cql:        "year from @2020-03-15T14:30:25.123",
			wantResult: newOrFatal(t, int32(2020)),
		},
		{
			name:       "Year from Date with year precision",
			cql:        "year from @2023",
			wantResult: newOrFatal(t, int32(2023)),
		},
		
		// Month extraction tests
		{
			name:       "Month from Date",
			cql:        "month from @2020-03-15",
			wantResult: newOrFatal(t, int32(3)),
		},
		{
			name:       "Month from DateTime",
			cql:        "month from @2020-03-15T14:30:25.123",
			wantResult: newOrFatal(t, int32(3)),
		},
		{
			name:       "Month from Date with month precision",
			cql:        "month from @2023-05",
			wantResult: newOrFatal(t, int32(5)),
		},
		
		// Day extraction tests
		{
			name:       "Day from Date",
			cql:        "day from @2020-03-15",
			wantResult: newOrFatal(t, int32(15)),
		},
		{
			name:       "Day from DateTime",
			cql:        "day from @2020-03-15T14:30:25.123",
			wantResult: newOrFatal(t, int32(15)),
		},
		
		// Hour extraction tests
		{
			name:       "Hour from DateTime",
			cql:        "hour from @2020-03-15T14:30:25.123",
			wantResult: newOrFatal(t, int32(14)),
		},
		{
			name:       "Hour from Time",
			cql:        "hour from @T14:30:25.123",
			wantResult: newOrFatal(t, int32(14)),
		},
		
		// Minute extraction tests
		{
			name:       "Minute from DateTime",
			cql:        "minute from @2020-03-15T14:30:25.123",
			wantResult: newOrFatal(t, int32(30)),
		},
		{
			name:       "Minute from Time",
			cql:        "minute from @T14:30:25.123",
			wantResult: newOrFatal(t, int32(30)),
		},
		
		// Second extraction tests
		{
			name:       "Second from DateTime",
			cql:        "second from @2020-03-15T14:30:25.123",
			wantResult: newOrFatal(t, int32(25)),
		},
		{
			name:       "Second from Time",
			cql:        "second from @T14:30:25.123",
			wantResult: newOrFatal(t, int32(25)),
		},
		
		// Millisecond extraction tests
		{
			name:       "Millisecond from DateTime",
			cql:        "millisecond from @2020-03-15T14:30:25.123",
			wantResult: newOrFatal(t, int32(123)),
		},
		{
			name:       "Millisecond from Time",
			cql:        "millisecond from @T14:30:25.123",
			wantResult: newOrFatal(t, int32(123)),
		},
		
		// Null handling tests
		{
			name:       "Year from null",
			cql:        "year from (null as DateTime)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Month from null",
			cql:        "month from (null as Date)",
			wantResult: newOrFatal(t, nil),
		},
		
		// Precision boundary tests - requesting unavailable precision should return null
		{
			name:       "Hour from Date returns null",
			cql:        "hour from @2020-03-15",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Minute from Date returns null",
			cql:        "minute from @2020-03-15",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Second from Date returns null",
			cql:        "second from @2020-03-15",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Millisecond from Date returns null",
			cql:        "millisecond from @2020-03-15",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Year from Time returns null",
			cql:        "year from @T14:30:25.123",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Month from Time returns null",
			cql:        "month from @T14:30:25.123",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Day from Time returns null",
			cql:        "day from @T14:30:25.123",
			wantResult: newOrFatal(t, nil),
		},
		
		// Precision insufficient tests - requesting higher precision than available
		{
			name:       "Month from year-precision Date returns null",
			cql:        "month from @2020",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Day from year-precision Date returns null",
			cql:        "day from @2020",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Day from month-precision Date returns null",
			cql:        "day from @2020-03",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Minute from hour-precision Time returns null",
			cql:        "minute from @T14",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Second from minute-precision Time returns null",
			cql:        "second from @T14:30",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Millisecond from second-precision Time returns null",
			cql:        "millisecond from @T14:30:25",
			wantResult: newOrFatal(t, nil),
		},
		
		// Edge cases
		{
			name:       "Year from leap year",
			cql:        "year from @2020-02-29",
			wantResult: newOrFatal(t, int32(2020)),
		},
		{
			name:       "Month from December",
			cql:        "month from @2020-12-31",
			wantResult: newOrFatal(t, int32(12)),
		},
		{
			name:       "Day from end of month",
			cql:        "day from @2020-12-31",
			wantResult: newOrFatal(t, int32(31)),
		},
		{
			name:       "Hour from midnight",
			cql:        "hour from @2020-01-01T00:00:00",
			wantResult: newOrFatal(t, int32(0)),
		},
		{
			name:       "Hour from end of day",
			cql:        "hour from @2020-01-01T23:59:59",
			wantResult: newOrFatal(t, int32(23)),
		},
		{
			name:       "Minute from start of hour",
			cql:        "minute from @2020-01-01T14:00:00",
			wantResult: newOrFatal(t, int32(0)),
		},
		{
			name:       "Minute from end of hour",
			cql:        "minute from @2020-01-01T14:59:59",
			wantResult: newOrFatal(t, int32(59)),
		},
		{
			name:       "Second from start of minute",
			cql:        "second from @2020-01-01T14:30:00",
			wantResult: newOrFatal(t, int32(0)),
		},
		{
			name:       "Second from end of minute",
			cql:        "second from @2020-01-01T14:30:59",
			wantResult: newOrFatal(t, int32(59)),
		},
		{
			name:       "Millisecond zero",
			cql:        "millisecond from @2020-01-01T14:30:25.000",
			wantResult: newOrFatal(t, int32(0)),
		},
		{
			name:       "Millisecond max",
			cql:        "millisecond from @2020-01-01T14:30:25.999",
			wantResult: newOrFatal(t, int32(999)),
		},
		
		// Timezone offset extraction tests
		{
			name:       "Timezone offset from DateTime with positive offset",
			cql:        "timezoneoffset from @2020-03-15T14:30:25.123+05:00",
			wantResult: newOrFatal(t, 5.0),
		},
		{
			name:       "Timezone offset from DateTime with negative offset",
			cql:        "timezoneoffset from @2020-03-15T14:30:25.123-07:00",
			wantResult: newOrFatal(t, -7.0),
		},
		{
			name:       "Timezone offset from DateTime with UTC",
			cql:        "timezoneoffset from @2020-03-15T14:30:25.123Z",
			wantResult: newOrFatal(t, 0.0),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := newFHIRParser(t)
			parsedLibs, err := p.Libraries(context.Background(), wrapInLib(t, tc.cql), parser.Config{})
			if err != nil {
				if tc.wantErr {
					return
				}
				t.Fatalf("Parse returned unexpected error: %v", err)
			}
			if tc.wantErr {
				t.Fatalf("Parse succeeded, expected error")
			}

			results, err := interpreter.Eval(context.Background(), parsedLibs, defaultInterpreterConfig(t, p))
			if err != nil {
				t.Fatalf("Eval returned unexpected error: %v", err)
			}

			if diff := cmp.Diff(tc.wantResult, getTESTRESULT(t, results), protocmp.Transform()); diff != "" {
				t.Errorf("Eval diff (-want +got)\n%v", diff)
			}
		})
	}
}

func TestDateTimeComponentFromPrecisionLogic(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantResult result.Value
	}{
		// Test precision logic with different Date precisions
		{
			name:       "Year from year-precision Date",
			cql:        "year from @2020",
			wantResult: newOrFatal(t, int32(2020)),
		},
		{
			name:       "Year from month-precision Date",
			cql:        "year from @2020-03",
			wantResult: newOrFatal(t, int32(2020)),
		},
		{
			name:       "Year from day-precision Date",
			cql:        "year from @2020-03-15",
			wantResult: newOrFatal(t, int32(2020)),
		},
		{
			name:       "Month from month-precision Date",
			cql:        "month from @2020-03",
			wantResult: newOrFatal(t, int32(3)),
		},
		{
			name:       "Month from day-precision Date",
			cql:        "month from @2020-03-15",
			wantResult: newOrFatal(t, int32(3)),
		},
		{
			name:       "Day from day-precision Date",
			cql:        "day from @2020-03-15",
			wantResult: newOrFatal(t, int32(15)),
		},
		
		// Test precision logic with different Time precisions
		{
			name:       "Hour from hour-precision Time",
			cql:        "hour from @T14",
			wantResult: newOrFatal(t, int32(14)),
		},
		{
			name:       "Hour from minute-precision Time",
			cql:        "hour from @T14:30",
			wantResult: newOrFatal(t, int32(14)),
		},
		{
			name:       "Hour from second-precision Time",
			cql:        "hour from @T14:30:25",
			wantResult: newOrFatal(t, int32(14)),
		},
		{
			name:       "Hour from millisecond-precision Time",
			cql:        "hour from @T14:30:25.123",
			wantResult: newOrFatal(t, int32(14)),
		},
		{
			name:       "Minute from minute-precision Time",
			cql:        "minute from @T14:30",
			wantResult: newOrFatal(t, int32(30)),
		},
		{
			name:       "Minute from second-precision Time",
			cql:        "minute from @T14:30:25",
			wantResult: newOrFatal(t, int32(30)),
		},
		{
			name:       "Minute from millisecond-precision Time",
			cql:        "minute from @T14:30:25.123",
			wantResult: newOrFatal(t, int32(30)),
		},
		{
			name:       "Second from second-precision Time",
			cql:        "second from @T14:30:25",
			wantResult: newOrFatal(t, int32(25)),
		},
		{
			name:       "Second from millisecond-precision Time",
			cql:        "second from @T14:30:25.123",
			wantResult: newOrFatal(t, int32(25)),
		},
		{
			name:       "Millisecond from millisecond-precision Time",
			cql:        "millisecond from @T14:30:25.123",
			wantResult: newOrFatal(t, int32(123)),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := newFHIRParser(t)
			parsedLibs, err := p.Libraries(context.Background(), wrapInLib(t, tc.cql), parser.Config{})
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}

			results, err := interpreter.Eval(context.Background(), parsedLibs, defaultInterpreterConfig(t, p))
			if err != nil {
				t.Fatalf("Eval returned unexpected error: %v", err)
			}

			if diff := cmp.Diff(tc.wantResult, getTESTRESULT(t, results), protocmp.Transform()); diff != "" {
				t.Errorf("Eval diff (-want +got)\n%v", diff)
			}
		})
	}
}
