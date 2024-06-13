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

package interpreter

import (
	"strings"
	"testing"
	"time"

	"github.com/google/cql/model"
	"github.com/google/cql/result"
)

func TestCompareWithPrecision(t *testing.T) {
	tests := []struct {
		name      string
		l         result.DateTime
		r         result.DateTime
		precision model.DateTimePrecision
		want      comparison
	}{
		{
			name: "same dates with year precision returns 0",
			l: result.DateTime{
				Date:      time.Date(2024, time.March, 31, 1, 20, 30, 1e8, time.UTC),
				Precision: model.SECOND,
			},
			r: result.DateTime{
				Date:      time.Date(2024, time.March, 31, 1, 20, 30, 1e8, time.UTC),
				Precision: model.SECOND,
			},
			precision: model.YEAR,
			want:      leftEqualRight,
		},
		{
			name: "two different dates with same year and year precision returns 0",
			l: result.DateTime{
				Date:      time.Date(2024, time.February, 5, 1, 20, 30, 1e8, time.UTC),
				Precision: model.SECOND,
			},
			r: result.DateTime{
				Date:      time.Date(2024, time.March, 22, 1, 20, 30, 1e8, time.UTC),
				Precision: model.SECOND,
			},
			precision: model.YEAR,
			want:      leftEqualRight,
		},
		{
			name:      "same dates with month precision returns 0",
			precision: model.MONTH,
			l: result.DateTime{
				Date:      time.Date(2024, time.March, 31, 1, 20, 30, 1e8, time.UTC),
				Precision: model.SECOND,
			},
			r: result.DateTime{
				Date:      time.Date(2024, time.March, 31, 1, 20, 30, 1e8, time.UTC),
				Precision: model.SECOND,
			},
			want: leftEqualRight,
		},
		{
			name:      "left month is lower with month precision returns -1",
			precision: model.MONTH,
			l: result.DateTime{
				Date:      time.Date(2024, time.February, 22, 1, 20, 30, 1e8, time.UTC),
				Precision: model.SECOND,
			},
			r: result.DateTime{
				Date:      time.Date(2024, time.March, 22, 1, 20, 30, 1e8, time.UTC),
				Precision: model.SECOND,
			},
			want: leftBeforeRight,
		},
		{
			name:      "left month is higher with month precision returns 1",
			precision: model.MONTH,
			l: result.DateTime{
				Date:      time.Date(2024, time.September, 31, 1, 20, 30, 1e8, time.UTC),
				Precision: model.SECOND,
			},
			r: result.DateTime{
				Date:      time.Date(2024, time.March, 31, 1, 20, 30, 1e8, time.UTC),
				Precision: model.SECOND,
			},
			want: leftAfterRight,
		},
		{
			name:      "left year is greater but same month at month precision returns 1",
			precision: model.MONTH,
			l: result.DateTime{
				Date:      time.Date(2030, time.March, 31, 1, 20, 30, 1e8, time.UTC),
				Precision: model.SECOND,
			},
			r: result.DateTime{
				Date:      time.Date(2024, time.March, 31, 1, 20, 30, 1e8, time.UTC),
				Precision: model.SECOND,
			},
			want: leftAfterRight,
		},
		{
			name:      "dates are the same, l has higher precision and should return precision error",
			precision: model.DAY,
			l: result.DateTime{
				Date:      time.Date(2024, time.March, 31, 1, 20, 30, 1e8, time.UTC),
				Precision: model.MONTH,
			},
			r: result.DateTime{
				Date:      time.Date(2024, time.March, 31, 1, 20, 30, 1e8, time.UTC),
				Precision: model.SECOND,
			},
			want: insufficientPrecision,
		},
		{
			name:      "leftAfterRight: unset precision uses finest precision of l and r (model.DAY)",
			precision: model.UNSETDATETIMEPRECISION,
			l: result.DateTime{
				Date:      time.Date(2024, time.April, 31, 1, 20, 30, 1e8, time.UTC),
				Precision: model.MONTH,
			},
			r: result.DateTime{
				Date:      time.Date(2024, time.March, 31, 1, 20, 30, 1e8, time.UTC),
				Precision: model.DAY,
			},
			want: leftAfterRight,
		},
		{
			name:      "leftEqualRight: unset precision uses finest precision of l and r (model.DAY)",
			precision: model.UNSETDATETIMEPRECISION,
			l: result.DateTime{
				Date:      time.Date(2024, time.March, 31, 1, 20, 30, 1e8, time.UTC),
				Precision: model.DAY,
			},
			r: result.DateTime{
				Date:      time.Date(2024, time.March, 31, 1, 20, 30, 1e8, time.UTC),
				Precision: model.DAY,
			},
			want: leftEqualRight,
		},
		{
			name:      "leftBeforeRight: unset precision uses finest precision of l and r (model.DAY)",
			precision: model.UNSETDATETIMEPRECISION,
			l: result.DateTime{
				Date:      time.Date(2024, time.January, 31, 1, 20, 30, 1e8, time.UTC),
				Precision: model.DAY,
			},
			r: result.DateTime{
				Date:      time.Date(2024, time.March, 31, 1, 20, 30, 1e8, time.UTC),
				Precision: model.MONTH,
			},
			want: leftBeforeRight,
		},
		{
			name:      "insufficientPrecision: unset precision uses finest precision of l and r (model.DAY)",
			precision: model.UNSETDATETIMEPRECISION,
			l: result.DateTime{
				Date:      time.Date(2024, time.March, 31, 1, 20, 30, 1e8, time.UTC),
				Precision: model.MONTH,
			},
			r: result.DateTime{
				Date:      time.Date(2024, time.March, 31, 1, 20, 30, 1e8, time.UTC),
				Precision: model.DAY,
			},
			want: insufficientPrecision,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := compareDateTimeWithPrecision(test.l, test.r, test.precision)
			if err != nil {
				t.Errorf("compareDateTimeWithPrecision(%v, %v, %v) returned unexpected error: %v", test.l, test.r, test.precision, err)
			}
			if got != test.want {
				t.Errorf("compareDateTimeWithPrecision(%v, %v, %v) = %v, want %v", test.l, test.r, test.precision, got, test.want)
			}
		})
	}
}

func TestCompareDateTimeWithPrecision_Errors(t *testing.T) {
	tests := []struct {
		name            string
		l               result.DateTime
		r               result.DateTime
		precision       model.DateTimePrecision
		wantErrContains string
	}{
		{
			name: "unset precision on l DateTimeValue",
			l: result.DateTime{
				Date:      time.Date(2024, time.March, 31, 1, 20, 30, 1e8, time.UTC),
				Precision: model.UNSETDATETIMEPRECISION,
			},
			r: result.DateTime{
				Date:      time.Date(2024, time.March, 31, 1, 20, 30, 1e8, time.UTC),
				Precision: model.SECOND,
			},
			precision:       model.UNSETDATETIMEPRECISION,
			wantErrContains: "internal error -- input to getFinestPrecision must not be unset.",
		},
		{
			name: "unset precision on r DateTimeValue",
			l: result.DateTime{
				Date:      time.Date(2024, time.March, 31, 1, 20, 30, 1e8, time.UTC),
				Precision: model.SECOND,
			},
			r: result.DateTime{
				Date:      time.Date(2024, time.March, 31, 1, 20, 30, 1e8, time.UTC),
				Precision: model.UNSETDATETIMEPRECISION,
			},
			precision:       model.UNSETDATETIMEPRECISION,
			wantErrContains: "internal error -- input to getFinestPrecision must not be unset.",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := compareDateTimeWithPrecision(test.l, test.r, test.precision)
			if !strings.Contains(err.Error(), test.wantErrContains) {
				t.Errorf("compareDateTimeWithPrecision(%v, %v, %v) returned unexpected error: got: %v, want contains: %v", test.l, test.r, test.precision, err, test.wantErrContains)
			}
		})
	}
}
