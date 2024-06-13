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

package result

import (
	"errors"
	"testing"
	"time"

	"github.com/google/cql/model"
	"github.com/google/cql/types"
	c4pb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/codes_go_proto"
	r4patientpb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/resources/patient_go_proto"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestToInt32(t *testing.T) {
	tests := []struct {
		name  string
		input Value
		want  int32
	}{
		{
			name:  "Integer",
			input: newOrFatal(t, 4),
			want:  4,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ToInt32(test.input)
			if err != nil {
				t.Fatalf("Int32(%v) failed: %v", test.input, err)
			}
			if got != test.want {
				t.Errorf("Int32(%v) got: %v want: %v", test.input, got, test.want)
			}
		})
	}
}

func TestToInt32Error(t *testing.T) {
	input := newOrFatal(t, 4.0)
	want := "cannot convert"
	_, err := ToInt32(input)
	if err == nil {
		t.Fatalf("Int32(%v) succeeded, want error", input)
	}
	if !errors.Is(err, ErrCannotConvert) {
		t.Errorf("Int32() got error %v want %v", err, want)
	}
}

func TestInt64(t *testing.T) {
	tests := []struct {
		name  string
		input Value
		want  int64
	}{
		{
			name:  "Long",
			input: newOrFatal(t, int64(4)),
			want:  int64(4),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ToInt64(test.input)
			if err != nil {
				t.Fatalf("Int64(%v) failed: %v", test.input, err)
			}
			if got != test.want {
				t.Errorf("Int64(%v) got: %v want: %v", test.input, got, test.want)
			}
		})
	}
}

func TestInt64Error(t *testing.T) {
	input := newOrFatal(t, 4.0)
	want := "cannot convert"
	_, err := ToInt64(input)
	if err == nil {
		t.Fatalf("Int64(%v) succeeded, want error", input)
	}
	if !errors.Is(err, ErrCannotConvert) {
		t.Errorf("Int64() got error %v want %v", err, want)
	}
}

func TestToFloat64(t *testing.T) {
	tests := []struct {
		name  string
		input Value
		want  float64
	}{
		{
			name:  "Decimal",
			input: newOrFatal(t, 4.0),
			want:  4.0,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ToFloat64(test.input)
			if err != nil {
				t.Fatalf("Float64(%v) failed: %v", test.input, err)
			}
			if got != test.want {
				t.Errorf("Float64(%v) got: %v want: %v", test.input, got, test.want)
			}
		})
	}
}

func TestFloat64Error(t *testing.T) {
	input := newOrFatal(t, "hello")
	want := "cannot convert"
	_, err := ToFloat64(input)
	if err == nil {
		t.Fatalf("Float64(%v) succeeded, want error", input)
	}
	if !errors.Is(err, ErrCannotConvert) {
		t.Errorf("Returned error (%s) did not contain expected (%s)", err, want)
	}
}

func TestToQuantity(t *testing.T) {
	tests := []struct {
		name  string
		input Value
		want  Quantity
	}{
		{
			name:  "Quantity",
			input: newOrFatal(t, Quantity{Value: 4.0, Unit: "day"}),
			want:  Quantity{Value: 4.0, Unit: "day"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ToQuantity(test.input)
			if err != nil {
				t.Fatalf("ToQuantity(%v) failed: %v", test.input, err)
			}
			if got != test.want {
				t.Errorf("ToQuantity(%v) got: %v want: %v", test.input, got, test.want)
			}
		})
	}
}

func TestQuantityError(t *testing.T) {
	input := newOrFatal(t, "hello")
	want := "cannot convert"
	_, err := ToQuantity(input)
	if err == nil {
		t.Fatalf("ToQuantity(%v) succeeded, want error", input)
	}
	if !errors.Is(err, ErrCannotConvert) {
		t.Errorf("ToQuantity() got error %v want %v", err, want)
	}
}

func TestToRatio(t *testing.T) {
	tests := []struct {
		name  string
		input Value
		want  Ratio
	}{
		{
			name:  "Ratio",
			input: newOrFatal(t, Ratio{Numerator: Quantity{Value: 4.0, Unit: "day"}, Denominator: Quantity{Value: 5.0, Unit: "day"}}),
			want:  Ratio{Numerator: Quantity{Value: 4.0, Unit: "day"}, Denominator: Quantity{Value: 5.0, Unit: "day"}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ToRatio(test.input)
			if err != nil {
				t.Fatalf("ToRatio(%v) failed: %v", test.input, err)
			}
			if got != test.want {
				t.Errorf("ToRatio(%v) got: %v want: %v", test.input, got, test.want)
			}
		})
	}
}

func TestRatioError(t *testing.T) {
	input := newOrFatal(t, "hello")
	want := "cannot convert"
	_, err := ToRatio(input)
	if err == nil {
		t.Fatalf("ToRatio(%v) succeeded, want error", input)
	}
	if !errors.Is(err, ErrCannotConvert) {
		t.Errorf("ToRatio() got error %v want %v", err, want)
	}
}

func TestToSlice(t *testing.T) {
	tests := []struct {
		name  string
		input Value
		want  []Value
	}{
		{
			name:  "List",
			input: newOrFatal(t, List{Value: []Value{newOrFatal(t, 4)}}),
			want:  []Value{newOrFatal(t, 4)},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ToSlice(test.input)
			if err != nil {
				t.Fatalf("ToSlice(%v) failed: %v", test.input, err)
			}
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("ToSlice(%v) returned diff (-want +got):\n%s", test.input, diff)
			}
		})
	}
}

func TestToSliceError(t *testing.T) {
	input := newOrFatal(t, 4.0)
	want := "cannot convert"
	_, err := ToSlice(input)
	if err == nil {
		t.Fatalf("ToSlice(%v) succeeded, want error", input)
	}
	if !errors.Is(err, ErrCannotConvert) {
		t.Errorf("ToSlice() got error %v want %v", err, want)
	}
}

func TestToTuple(t *testing.T) {
	tests := []struct {
		name  string
		input Value
		want  map[string]Value
	}{
		{
			name:  "Tuple",
			input: newOrFatal(t, Tuple{Value: map[string]Value{"Apple": newOrFatal(t, 4)}}),
			want:  map[string]Value{"Apple": newOrFatal(t, 4)},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ToTuple(test.input)
			if err != nil {
				t.Fatalf("ToTuple(%v) failed: %v", test.input, err)
			}
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("ToTuple(%v) returned diff (-want +got):\n%s", test.input, diff)
			}
		})
	}
}

func TestToTupleError(t *testing.T) {
	input := newOrFatal(t, 4.0)
	want := "cannot convert"
	_, err := ToTuple(input)
	if err == nil {
		t.Fatalf("ToTuple(%v) succeeded, want error", input)
	}
	if !errors.Is(err, ErrCannotConvert) {
		t.Errorf("ToTuple() got error %v want %v", err, want)
	}
}

func TestToProto(t *testing.T) {
	tests := []struct {
		name  string
		input Value
		want  proto.Message
	}{
		{
			name:  "Proto",
			input: newOrFatal(t, Named{Value: &r4patientpb.Patient_GenderCode{Value: c4pb.AdministrativeGenderCode_MALE}}),
			want:  &r4patientpb.Patient_GenderCode{Value: c4pb.AdministrativeGenderCode_MALE},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ToProto(test.input)
			if err != nil {
				t.Fatalf("ToProto(%v) failed: %v", test.input, err)
			}
			if diff := cmp.Diff(test.want, got, protocmp.Transform()); diff != "" {
				t.Errorf("ToProto(%v) returned diff (-want +got):\n%s", test.input, diff)
			}
		})
	}
}

func TestToProtoError(t *testing.T) {
	input := newOrFatal(t, 4.0)
	want := "cannot convert"
	_, err := ToProto(input)
	if err == nil {
		t.Fatalf("ToProto(%v) succeeded, want error", input)
	}
	if !errors.Is(err, ErrCannotConvert) {
		t.Errorf("ToProto() got error %v want %v", err, want)
	}
}
func TestToDateTime(t *testing.T) {
	tests := []struct {
		name  string
		input Value
		want  DateTime
	}{
		{
			name:  "Date",
			input: newOrFatal(t, Date{Date: time.Date(2023, time.March, 5, 0, 0, 0, 0, time.UTC), Precision: model.DAY}),
			want:  DateTime{Date: time.Date(2023, time.March, 5, 0, 0, 0, 0, time.UTC), Precision: model.DAY},
		},
		{
			name:  "DateTime",
			input: newOrFatal(t, DateTime{Date: time.Date(2024, time.March, 31, 1, 20, 30, 1e8, time.UTC)}),
			want:  DateTime{Date: time.Date(2024, time.March, 31, 1, 20, 30, 1e8, time.UTC)},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ToDateTime(test.input)
			if err != nil {
				t.Fatalf("ToDateTime(%v) failed: %v", test.input, err)
			}
			if got != test.want {
				t.Errorf("ToDateTime(%v) got: %v want: %v", test.input, got, test.want)
			}
		})
	}
}

func TestToDateTimeError(t *testing.T) {
	input := newOrFatal(t, 4.0)
	want := "cannot convert"
	_, err := ToDateTime(input)
	if err == nil {
		t.Fatalf("ToDateTime(%v) succeeded, want error", input)
	}
	if !errors.Is(err, ErrCannotConvert) {
		t.Errorf("ToDateTime() got error %v want %v", err, want)
	}
}

func TestToInterval(t *testing.T) {
	tests := []struct {
		name  string
		input Value
		want  Interval
	}{
		{
			name: "Interval",
			input: newOrFatal(t, Interval{
				Low:           newOrFatal(t, 4),
				High:          newOrFatal(t, 5),
				LowInclusive:  true,
				HighInclusive: true,
				StaticType:    &types.Interval{PointType: types.Integer},
			}),
			want: Interval{
				Low:           newOrFatal(t, 4),
				High:          newOrFatal(t, 5),
				LowInclusive:  true,
				HighInclusive: true,
				StaticType:    &types.Interval{PointType: types.Integer},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ToInterval(test.input)
			if err != nil {
				t.Fatalf("ToInterval(%v) failed: %v", test.input, err)
			}
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("ToInterval(%v) returned diff (-want +got):\n%s", test.input, diff)
			}
		})
	}
}

func TestIntervalError(t *testing.T) {
	input := newOrFatal(t, 4.0)
	want := "cannot convert"
	_, err := ToInterval(input)
	if err == nil {
		t.Fatalf("ToInterval(%v) succeeded, want error", input)
	}
	if !errors.Is(err, ErrCannotConvert) {
		t.Errorf("ToInterval() got error %v want %v", err, want)
	}
}

func TestToCode(t *testing.T) {
	tests := []struct {
		name  string
		input Value
		want  Code
	}{
		{
			name:  "ToCode",
			input: newOrFatal(t, Code{System: "foo", Code: "bar", Display: "the foo", Version: "1.0"}),
			want:  Code{System: "foo", Code: "bar", Display: "the foo", Version: "1.0"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ToCode(test.input)
			if err != nil {
				t.Fatalf("ToCode(%v) failed: %v", test.input, err)
			}
			if got != test.want {
				t.Errorf("ToCode(%v) got: %v want: %v", test.input, got, test.want)
			}
		})
	}
}

func TestToCodeError(t *testing.T) {
	input := newOrFatal(t, 4.0)
	want := "cannot convert"
	_, err := ToCode(input)
	if err == nil {
		t.Fatalf("ToCode(%v) succeeded, want error", input)
	}
	if !errors.Is(err, ErrCannotConvert) {
		t.Errorf("ToCode() got error %v want %v", err, want)
	}
}

func TestToCodeSystem(t *testing.T) {
	tests := []struct {
		name  string
		input Value
		want  CodeSystem
	}{
		{
			name:  "CodeSystem",
			input: newOrFatal(t, CodeSystem{ID: "example.com", Version: "1.0"}),
			want:  CodeSystem{ID: "example.com", Version: "1.0"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ToCodeSystem(test.input)
			if err != nil {
				t.Fatalf("ToCodeSystem(%v) failed: %v", test.input, err)
			}
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("ToCodeSystem(%v) returned diff (-want +got):\n%s", test.input, diff)
			}
		})
	}
}

func TestToCodeSystemError(t *testing.T) {
	input := newOrFatal(t, 4.0)
	want := "cannot convert"
	_, err := ToCodeSystem(input)
	if err == nil {
		t.Fatalf("ToCodeSystem(%v) succeeded, want error", input)
	}
	if !errors.Is(err, ErrCannotConvert) {
		t.Errorf("ToCodeSystem() got error %v want %v", err, want)
	}
}

func TestToValueSet(t *testing.T) {
	test := struct {
		name  string
		input Value
		want  ValueSet
	}{
		name:  "ValueSet",
		input: newOrFatal(t, ValueSet{ID: "example.com", Version: "1.0"}),
		want:  ValueSet{ID: "example.com", Version: "1.0"},
	}
	t.Run(test.name, func(t *testing.T) {
		got, err := ToValueSet(test.input)
		if err != nil {
			t.Fatalf("ToValueSet(%v) failed: %v", test.input, err)
		}
		if diff := cmp.Diff(test.want, got); diff != "" {
			t.Errorf("ToValueSet(%v) returned diff (-want +got):\n%s", test.input, diff)
		}
	})
}

func TestToValueSetError(t *testing.T) {
	input := newOrFatal(t, 4.0)
	want := "cannot convert"
	_, err := ToValueSet(input)
	if err == nil {
		t.Fatalf("ToValueSet(%v) succeeded, want error", input)
	}
	if !errors.Is(err, ErrCannotConvert) {
		t.Errorf("ToValueSet() got error %v want %v", err, want)
	}
}
