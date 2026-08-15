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
	}{
		// DateTime component extraction
		{
			name:       "Year from DateTime",
			cql:        "year from @2023-05-15T14:30:25.100",
			wantResult: newOrFatal(t, int32(2023)),
		},
		{
			name:       "Month from DateTime",
			cql:        "month from @2023-05-15T14:30:25.100",
			wantResult: newOrFatal(t, int32(5)),
		},
		{
			name:       "Day from DateTime",
			cql:        "day from @2023-05-15T14:30:25.100",
			wantResult: newOrFatal(t, int32(15)),
		},
		{
			name:       "Hour from DateTime",
			cql:        "hour from @2023-05-15T14:30:25.100",
			wantResult: newOrFatal(t, int32(14)),
		},
		{
			name:       "Minute from DateTime",
			cql:        "minute from @2023-05-15T14:30:25.100",
			wantResult: newOrFatal(t, int32(30)),
		},
		{
			name:       "Second from DateTime",
			cql:        "second from @2023-05-15T14:30:25.100",
			wantResult: newOrFatal(t, int32(25)),
		},
		{
			name:       "Millisecond from DateTime",
			cql:        "millisecond from @2023-05-15T14:30:25.100",
			wantResult: newOrFatal(t, int32(100)),
		},
		// Date component extraction
		{
			name:       "Year from Date",
			cql:        "year from @2023-05-15",
			wantResult: newOrFatal(t, int32(2023)),
		},
		{
			name:       "Month from Date",
			cql:        "month from @2023-05-15",
			wantResult: newOrFatal(t, int32(5)),
		},
		{
			name:       "Day from Date",
			cql:        "day from @2023-05-15",
			wantResult: newOrFatal(t, int32(15)),
		},
		// Insufficient precision cases for DateTime
		{
			name:       "Hour from DateTime with insufficient precision",
			cql:        "hour from @2023-05-15",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Minute from DateTime with insufficient precision",
			cql:        "minute from @2023-05",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Second from DateTime with insufficient precision",
			cql:        "second from @2023",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Millisecond from DateTime with insufficient precision",
			cql:        "millisecond from @2023-05-15T14:30",
			wantResult: newOrFatal(t, nil),
		},
		// Null cases
		{
			name:       "Year from null DateTime",
			cql:        "year from (null as DateTime)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Month from null Date",
			cql:        "month from (null as Date)",
			wantResult: newOrFatal(t, nil),
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

func TestTimeComponentFrom(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantResult result.Value
	}{
		// Time component extraction
		{
			name:       "Hour from Time",
			cql:        "hour from @T14:30:25.100",
			wantResult: newOrFatal(t, int32(14)),
		},
		{
			name:       "Minute from Time",
			cql:        "minute from @T14:30:25.100",
			wantResult: newOrFatal(t, int32(30)),
		},
		{
			name:       "Second from Time",
			cql:        "second from @T14:30:25.100",
			wantResult: newOrFatal(t, int32(25)),
		},
		{
			name:       "Millisecond from Time",
			cql:        "millisecond from @T14:30:25.100",
			wantResult: newOrFatal(t, int32(100)),
		},
		// Insufficient precision cases for Time
		{
			name:       "Minute from Time with insufficient precision",
			cql:        "minute from @T14",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Second from Time with insufficient precision",
			cql:        "second from @T14:30",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Millisecond from Time with insufficient precision",
			cql:        "millisecond from @T14:30:25",
			wantResult: newOrFatal(t, nil),
		},
		// Edge cases
		{
			name:       "Hour from midnight Time",
			cql:        "hour from @T00:00:00.000",
			wantResult: newOrFatal(t, int32(0)),
		},
		{
			name:       "Hour from late evening Time",
			cql:        "hour from @T23:59:59.999",
			wantResult: newOrFatal(t, int32(23)),
		},
		{
			name:       "Millisecond from Time with zero milliseconds",
			cql:        "millisecond from @T14:30:25.000",
			wantResult: newOrFatal(t, int32(0)),
		},
		// Null cases
		{
			name:       "Hour from null Time",
			cql:        "hour from (null as Time)",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Minute from null Time",
			cql:        "minute from (null as Time)",
			wantResult: newOrFatal(t, nil),
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

func TestMixedComponentExtraction(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantResult result.Value
	}{
		// Test that the parser correctly chooses DateTime vs Time component extraction
		{
			name:       "Hour from DateTime vs Time - DateTime",
			cql:        "hour from @2023-05-15T14:30:25.100",
			wantResult: newOrFatal(t, int32(14)),
		},
		{
			name:       "Hour from DateTime vs Time - Time",
			cql:        "hour from @T14:30:25.100",
			wantResult: newOrFatal(t, int32(14)),
		},
		{
			name:       "Minute from DateTime vs Time - DateTime",
			cql:        "minute from @2023-05-15T14:30:25.100",
			wantResult: newOrFatal(t, int32(30)),
		},
		{
			name:       "Minute from DateTime vs Time - Time",
			cql:        "minute from @T14:30:25.100",
			wantResult: newOrFatal(t, int32(30)),
		},
		// Test expressions with calculations
		{
			name:       "Hour from calculated DateTime",
			cql:        "hour from (@2023-05-15T14:30:25.100 + 2 hours)",
			wantResult: newOrFatal(t, int32(16)),
		},
		{
			name:       "Day from calculated Date",
			cql:        "day from (@2023-05-15 + 10 days)",
			wantResult: newOrFatal(t, int32(25)),
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
