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
	"fmt"
	"time"

	"github.com/google/cql/model"
)

// DateString returns a CQL Date string representation of a Date.
func DateString(d time.Time, precision model.DateTimePrecision) (string, error) {
	var s string
	switch precision {
	case model.YEAR:
		s = d.Format(dateYear)
	case model.MONTH:
		s = d.Format(dateMonth)
	case model.DAY:
		s = d.Format(dateDay)
	default:
		return "", fmt.Errorf("unsupported precision in Date with value %v %w", precision, ErrUnsupportedPrecision)
	}
	return "@" + s, nil
}

// DateTimeString returns a CQL DateTime string representation of a DateTime.
func DateTimeString(d time.Time, precision model.DateTimePrecision) (string, error) {
	var dtFormat string
	switch precision {
	case model.YEAR:
		dtFormat = dateTimeYear
	case model.MONTH:
		dtFormat = dateTimeMonth
	case model.DAY:
		dtFormat = dateTimeDay
	case model.HOUR:
		dtFormat = dateTimeHour
	case model.MINUTE:
		dtFormat = dateTimeMinute
	case model.SECOND:
		dtFormat = dateTimeSecond
	case model.MILLISECOND:
		dtFormat = dateTimeThreeMillisecond
	default:
		return "", fmt.Errorf("unsupported precision in Date with value %v %w", precision, ErrUnsupportedPrecision)
	}
	tzFormat := "Z07:00" // uses "Z" for UTC timezones, -07:00 style for others.
	return "@" + d.Format(dtFormat+tzFormat), nil
}

// TimeString returns a CQL Time string representation of a Time.
func TimeString(d time.Time, precision model.DateTimePrecision) (string, error) {
	var tFormat string
	switch precision {
	case model.HOUR:
		tFormat = timeHour
	case model.MINUTE:
		tFormat = timeMinute
	case model.SECOND:
		tFormat = timeSecond
	case model.MILLISECOND:
		tFormat = timeThreeMillisecond
	default:
		return "", fmt.Errorf("unsupported precision in Date with value %v %w", precision, ErrUnsupportedPrecision)
	}
	return d.Format(tFormat), nil
}
