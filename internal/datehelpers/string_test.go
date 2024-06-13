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

package datehelpers

import (
	"testing"
	"time"

	"github.com/google/cql/model"
)

func TestDateString(t *testing.T) {
	tests := []struct {
		name      string
		d         time.Time
		precision model.DateTimePrecision
		want      string
	}{
		{
			name:      "year precision",
			d:         time.Date(2024, time.March, 31, 0, 0, 0, 0, time.UTC),
			precision: model.YEAR,
			want:      "@2024",
		},
		{
			name:      "month precision",
			d:         time.Date(2024, time.March, 31, 0, 0, 0, 0, time.UTC),
			precision: model.MONTH,
			want:      "@2024-03",
		},
		{
			name:      "day precision",
			d:         time.Date(2024, time.March, 31, 0, 0, 0, 0, time.UTC),
			precision: model.DAY,
			want:      "@2024-03-31",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := DateString(tc.d, tc.precision)
			if err != nil {
				t.Errorf("DateString(%v) returned unexpected error: %v", tc.d, err)
			}
			if got != tc.want {
				t.Errorf("DateString(%v) = %v, want %v", tc.d, got, tc.want)
			}
		})
	}
}

func TestDateTimeString(t *testing.T) {
	tests := []struct {
		name      string
		dt        time.Time
		precision model.DateTimePrecision
		want      string
	}{
		{
			name:      "hour precision",
			dt:        time.Date(2024, time.March, 31, 1, 20, 30, 1e8, time.UTC),
			precision: model.HOUR,
			want:      "@2024-03-31T01Z",
		},
		{
			name:      "minute precision",
			dt:        time.Date(2024, time.March, 31, 1, 20, 30, 1e8, time.UTC),
			precision: model.MINUTE,
			want:      "@2024-03-31T01:20Z",
		},
		{
			name:      "second precision",
			dt:        time.Date(2024, time.March, 31, 1, 20, 30, 1e8, time.UTC),
			precision: model.SECOND,
			want:      "@2024-03-31T01:20:30Z",
		},
		{
			name:      "millisecond precision",
			dt:        time.Date(2024, time.March, 31, 1, 20, 30, 1e8, time.UTC),
			precision: model.MILLISECOND,
			want:      "@2024-03-31T01:20:30.100Z",
		},
		{
			name:      "non-utc timezone",
			dt:        time.Date(2024, time.March, 31, 1, 20, 30, 1e8, time.FixedZone("Fixed", 4*60*60)),
			precision: model.MILLISECOND,
			want:      "@2024-03-31T01:20:30.100+04:00",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := DateTimeString(tc.dt, tc.precision)
			if err != nil {
				t.Errorf("DateTimeString(%v) returned unexpected error: %v", tc.dt, err)
			}
			if got != tc.want {
				t.Errorf("DateTimeString(%v) = %v, want %v", tc.dt, got, tc.want)
			}
		})
	}
}

func TestTimeString(t *testing.T) {
	tests := []struct {
		name      string
		t         time.Time
		precision model.DateTimePrecision
		want      string
	}{
		{
			name:      "hour precision",
			t:         time.Date(0, time.January, 1, 1, 20, 30, 1e8, time.UTC),
			precision: model.HOUR,
			want:      "T01",
		},
		{
			name:      "minute precision",
			t:         time.Date(0, time.January, 1, 1, 20, 30, 1e8, time.UTC),
			precision: model.MINUTE,
			want:      "T01:20",
		},
		{
			name:      "second precision",
			t:         time.Date(0, time.January, 1, 1, 20, 30, 1e8, time.UTC),
			precision: model.SECOND,
			want:      "T01:20:30",
		},
		{
			name:      "millisecond precision",
			t:         time.Date(0, time.January, 1, 1, 20, 30, 1e8, time.UTC),
			precision: model.MILLISECOND,
			want:      "T01:20:30.100",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := TimeString(tc.t, tc.precision)
			if err != nil {
				t.Errorf("TimeString(%v) returned unexpected error: %v", tc.t, err)
			}
			if got != tc.want {
				t.Errorf("TimeString(%v) = %v, want %v", tc.t, got, tc.want)
			}
		})
	}
}
