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

// Package datehelpers provides functions for parsing CQL date, datetime and time strings.
package datehelpers

import (
	"errors"
	"fmt"
	regex "regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/cql/model"
	"github.com/google/cql/types"
	d4pb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/datatypes_go_proto"
)

// Constants for parsing CQL date, datetime and time strings.
var (
	// Date layout constants.
	dateYear  = "2006"
	dateMonth = "2006-01"
	dateDay   = "2006-01-02"

	// DateTime layout constants.
	dateTimeYear             = "2006T"
	dateTimeMonth            = "2006-01T"
	dateTimeDay              = "2006-01-02T"
	dateTimeHour             = "2006-01-02T15"
	dateTimeMinute           = "2006-01-02T15:04"
	dateTimeSecond           = "2006-01-02T15:04:05"
	dateTimeOneMillisecond   = "2006-01-02T15:04:05.0"
	dateTimeTwoMillisecond   = "2006-01-02T15:04:05.00"
	dateTimeThreeMillisecond = "2006-01-02T15:04:05.000"

	// Time layout constants.
	timeHour             = "T15"
	timeMinute           = "T15:04"
	timeSecond           = "T15:04:05"
	timeOneMillisecond   = "T15:04:05.0"
	timeTwoMillisecond   = "T15:04:05.00"
	timeThreeMillisecond = "T15:04:05.000"

	// Timezone constants.
	zuluTZ = "Z"
	tz     = "-07:00"
)

// ErrUnsupportedPrecision is returned when a precision is not supported.
var ErrUnsupportedPrecision = errors.New("unsupported precision")

// ParseDate parses a CQL Date string into a golang time. CQL Dates start with @ and follow a subset
// of ISO-8601.
//
// CQL Dates do not have timezone offsets, but when converting a Date to a DateTime the offset of
// the evaluation timestamp is used. Since all golang times require a location we set all Date
// offset to the offset of the evaluation timestamp.
func ParseDate(rawStr string, evaluationLoc *time.Location) (time.Time, model.DateTimePrecision, error) {
	if evaluationLoc == nil {
		return time.Time{}, model.UNSETDATETIMEPRECISION, fmt.Errorf("internal error - evaluationLoc must be set when calling ParseDate")
	}

	if len(rawStr) == 0 || rawStr[0] != '@' {
		return time.Time{}, model.UNSETDATETIMEPRECISION, fmt.Errorf("internal error - datetime string %v, must start with @", rawStr)
	}
	str := rawStr[1:]

	dates := []struct {
		layout    string
		precision model.DateTimePrecision
	}{
		{layout: dateYear, precision: model.YEAR},
		{layout: dateMonth, precision: model.MONTH},
		{layout: dateDay, precision: model.DAY},
	}

	var err error
	var parsedTime time.Time
	for _, d := range dates {
		parsedTime, err = time.ParseInLocation(d.layout, str, evaluationLoc)
		if err == nil {
			return parsedTime, d.precision, nil
		}
	}

	if parseErr, ok := err.(*time.ParseError); ok {
		return time.Time{}, model.UNSETDATETIMEPRECISION, fmtParsingErr(rawStr, types.Date, "@YYYY-MM-DD", parseErr)
	}
	return time.Time{}, model.UNSETDATETIMEPRECISION, err
}

// ParseDateTime parses a CQL DateTime string into a golang time. CQL Dates start with @ and follow
// a subset of ISO-8601. If rawStr does not include an offset then evaluationLoc will be used.
// Otherwise, the offset in rawStr is used.
func ParseDateTime(rawStr string, evaluationLoc *time.Location) (time.Time, model.DateTimePrecision, error) {
	if evaluationLoc == nil {
		return time.Time{}, model.UNSETDATETIMEPRECISION, fmt.Errorf("internal error - evaluationLoc must be set when calling ParseDateTime")
	}

	if len(rawStr) == 0 || rawStr[0] != '@' {
		return time.Time{}, model.UNSETDATETIMEPRECISION, fmt.Errorf("internal error - datetime string %v, must start with @", rawStr)
	}
	str := rawStr[1:]

	// Since time.ParseInLocation allows any number of fractional seconds no matter the layout, we
	// must manually check.
	re := regex.MustCompile(`\.\d{4}`)
	if re.MatchString(rawStr) {
		return time.Time{}, model.UNSETDATETIMEPRECISION, fmt.Errorf("%v %v can have at most 3 digits of milliseconds precision, want a layout like @YYYY-MM-DDThh:mm:ss.fff(Z|(+/-hh:mm)", types.DateTime, rawStr)
	}

	datetimes := []struct {
		layout    string
		precision model.DateTimePrecision
	}{
		{layout: dateTimeYear, precision: model.YEAR},
		{layout: dateTimeMonth, precision: model.MONTH},
		{layout: dateTimeDay, precision: model.DAY},
		{layout: dateTimeHour, precision: model.HOUR},
		{layout: dateTimeMinute, precision: model.MINUTE},
		// For ParseInLocation, the input may contain a fractional second field immediately after the
		// seconds field, even if the layout does not signify its presence. So, we have to do things in
		// this order.
		{layout: dateTimeOneMillisecond, precision: model.MILLISECOND},
		{layout: dateTimeTwoMillisecond, precision: model.MILLISECOND},
		{layout: dateTimeThreeMillisecond, precision: model.MILLISECOND},
		{layout: dateTimeSecond, precision: model.SECOND},
	}

	var err error
	var parsedTime time.Time
	for _, dt := range datetimes {
		for _, timezone := range []string{zuluTZ, tz, ""} {
			loc := evaluationLoc
			if timezone == zuluTZ {
				loc = time.UTC
			}
			parsedTime, err = time.ParseInLocation(fmt.Sprintf("%v%v", dt.layout, timezone), str, loc)
			if err == nil {
				return parsedTime, dt.precision, nil
			}
		}
	}

	if parseErr, ok := err.(*time.ParseError); ok {
		return time.Time{}, model.UNSETDATETIMEPRECISION, fmtParsingErr(rawStr, types.DateTime, "@YYYY-MM-DDThh:mm:ss.fff(Z|(+/-hh:mm)", parseErr)
	}
	return time.Time{}, model.UNSETDATETIMEPRECISION, err
}

// ParseTime parses a CQL Time string into a golang time. CQL Time start with @ and roughly follow
// ISO-8601.
func ParseTime(rawStr string, evaluationLoc *time.Location) (time.Time, model.DateTimePrecision, error) {
	if len(rawStr) == 0 || rawStr[0] != '@' {
		return time.Time{}, model.UNSETDATETIMEPRECISION, fmt.Errorf("internal error - datetime string %v, must start with @", rawStr)
	}
	str := rawStr[1:]

	// Since time.ParseInLocation allows any number of fractional seconds no matter the layout, we
	// must manually check.
	re := regex.MustCompile(`\.\d{4}`)
	if re.MatchString(rawStr) {
		return time.Time{}, model.UNSETDATETIMEPRECISION, fmt.Errorf("%v %v can have at most 3 digits of milliseconds precision, want a layout like @Thh:mm:ss.fff", types.Time, rawStr)
	}

	times := []struct {
		layout    string
		precision model.DateTimePrecision
	}{
		{layout: timeHour, precision: model.HOUR},
		{layout: timeMinute, precision: model.MINUTE},
		// For ParseInLocation, the input may contain a fractional second field immediately after the
		// seconds field, even if the layout does not signify its presence. So, we have to do things in
		// this order.
		{layout: timeOneMillisecond, precision: model.MILLISECOND},
		{layout: timeTwoMillisecond, precision: model.MILLISECOND},
		{layout: timeThreeMillisecond, precision: model.MILLISECOND},
		{layout: timeSecond, precision: model.SECOND},
	}

	var err error
	var parsedTime time.Time
	for _, t := range times {
		parsedTime, err = time.ParseInLocation(t.layout, str, evaluationLoc)
		if err == nil {
			return parsedTime, t.precision, nil
		}
	}

	if parseErr, ok := err.(*time.ParseError); ok {
		return time.Time{}, model.UNSETDATETIMEPRECISION, fmtParsingErr(rawStr, types.Time, "@Thh:mm:ss.fff", parseErr)
	}
	return time.Time{}, model.UNSETDATETIMEPRECISION, err
}

// TODO: b/341120071 - if we refactor the FHIR proto library to expose the date/dateTime helpers,
// we can consider using them here to help. Specifically we can consider using SerializeDate to
// transform back to a FHIR string which might be easier to process, and include the FHIR proto
// roundtrip logic.

// ParseFHIRDate parses a FHIR Date proto into a golang time. Similar to other helpers in
// this package, if the proto does not have a timezone set then evaluationLoc will be used.
//
// To match ParseDate, we take the year, month, and day values from the FHIR proto (in the FHIR
// proto's timezone) and create a new time.Time at 0:00:00 in the evaluationLoc timezone, which is
// the default timezone we attach to all CQL Dates in our codebase.
func ParseFHIRDate(d *d4pb.Date, evaluationLoc *time.Location) (time.Time, model.DateTimePrecision, error) {
	secs := d.GetValueUs() / 1e6 // Integer division
	// time.Unix parses the Date into the local timezone, which is not what we want. What we want is
	// to get the Date (year, month, day) values in the original timezone of the proto, and then
	// attach the evaluationLoc timezone to it at 0:00:00.
	t := time.Unix(secs, 0)
	// if tz is not set then use the evaluationLoc. It's unclear if it's even possible for this to
	// be unset.
	loc := evaluationLoc
	if tz := d.GetTimezone(); tz != "" {
		var err error
		loc, err = getLocation(tz)
		if err != nil {
			return time.Time{}, model.UNSETDATETIMEPRECISION, fmt.Errorf("error loading timezone from FHIR %w", err)
		}
	}
	t = t.In(loc)
	// t is now in its original timezone. We're going to grab the day, month, and year and create
	// a new time in the evaluation request timezone.
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, evaluationLoc)

	return t, datePrecisionFromProto(d.GetPrecision()), nil
}

// ParseFHIRDateTime parses a FHIR DateTime proto into a golang time. Similar to other helpers in
// this package, if the proto does not have a timezone set then evaluationLoc will be used.
func ParseFHIRDateTime(d *d4pb.DateTime, evaluationLoc *time.Location) (time.Time, model.DateTimePrecision, error) {
	if evaluationLoc == nil {
		return time.Time{}, model.UNSETDATETIMEPRECISION, fmt.Errorf("internal error - evaluationLoc must be set when calling ParseFHIRDateTime")
	}
	secs := d.GetValueUs() / 1e6 // Integer division
	usecRemainder := d.GetValueUs() % 1e6
	nsRemainder := usecRemainder * 1e3
	t := time.Unix(secs, nsRemainder)

	// If the proto has a timezone set then we use it, otherwise we use the evaluationLoc. It's
	// unclear if it's even possible for the timezone to be unset.
	loc := evaluationLoc
	var err error
	if tz := d.GetTimezone(); tz != "" {
		loc, err = getLocation(tz)
		if err != nil {
			return time.Time{}, model.UNSETDATETIMEPRECISION, fmt.Errorf("error loading timezone from FHIR %w", err)
		}
	}

	return t.In(loc), dateTimePrecisionFromProto(d.GetPrecision()), nil
}

func datePrecisionFromProto(p d4pb.Date_Precision) model.DateTimePrecision {
	switch p {
	case d4pb.Date_YEAR:
		return model.YEAR
	case d4pb.Date_MONTH:
		return model.MONTH
	case d4pb.Date_DAY:
		return model.DAY
	}
	return model.UNSETDATETIMEPRECISION
}

func dateTimePrecisionFromProto(p d4pb.DateTime_Precision) model.DateTimePrecision {
	switch p {
	case d4pb.DateTime_YEAR:
		return model.YEAR
	case d4pb.DateTime_MONTH:
		return model.MONTH
	case d4pb.DateTime_DAY:
		return model.DAY
	case d4pb.DateTime_SECOND:
		return model.SECOND
	case d4pb.DateTime_MILLISECOND:
		return model.MILLISECOND
	// FHIR datetimes can have microsecond precision, since CQL doesn't support this we map it to millisecond.
	case d4pb.DateTime_MICROSECOND:
		return model.MILLISECOND
	}
	return model.UNSETDATETIMEPRECISION
}

// getLocation and offsetToSeconds are copied from FHIR Proto:
// https://github.com/google/fhir/blob/5ae1b8d319bce275c16457c1f3c321804c202488/go/jsonformat/internal/jsonpbhelper/fhirutil.go#L500
// TODO: b/341120071 - we should refactor FHIR proto so we can depend on their time helpers
// directly.

// getLocation parses tz as an IANA location or a UTC offset.
func getLocation(tz string) (*time.Location, error) {
	if tz == "UTC" {
		return time.UTC, nil
	}
	l, err := time.LoadLocation(tz)
	if err != nil {
		offset, err := offsetToSeconds(tz)
		if err != nil {
			return nil, err
		}
		return time.FixedZone(tz, offset), nil
	}
	return l, nil
}

func offsetToSeconds(offset string) (int, error) {
	if offset == "" || offset == "UTC" {
		return 0, nil
	}
	sign := offset[0]
	if sign != '+' && sign != '-' {
		return 0, fmt.Errorf("invalid timezone offset: %v", offset)
	}
	arr := strings.Split(offset[1:], ":")
	if len(arr) != 2 {
		return 0, fmt.Errorf("invalid timezone offset: %v", offset)
	}
	hour, err := strconv.Atoi(arr[0])
	if err != nil {
		return 0, fmt.Errorf("invalid hour in timezone offset %v: %v", offset, err)
	}
	minute, err := strconv.Atoi(arr[1])
	if err != nil {
		return 0, fmt.Errorf("invalid minute in timezone offset %v: %v", offset, err)
	}
	if sign == '-' {
		return -hour*3600 - minute*60, nil
	}
	return hour*3600 + minute*60, nil
}

func fmtParsingErr(rawStr string, t types.IType, layout string, e *time.ParseError) error {
	return fmt.Errorf("got %v %v but want a layout like %v%v", t, rawStr, layout, e.Message)
}
