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
	"encoding/json"
	"strings"
	"testing"
	"time"

	anypb "google.golang.org/protobuf/types/known/anypb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	datepb "google.golang.org/genproto/googleapis/type/date"
	timeofdaypb "google.golang.org/genproto/googleapis/type/timeofday"

	"github.com/google/cql/internal/datehelpers"
	"github.com/google/cql/model"
	crpb "github.com/google/cql/protos/cql_result_go_proto"
	ctpb "github.com/google/cql/protos/cql_types_go_proto"
	"github.com/google/cql/types"
	d4pb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/datatypes_go_proto"
	r4patientpb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/resources/patient_go_proto"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestEqual(t *testing.T) {
	tests := []struct {
		name      string
		a         Value
		b         Value
		wantEqual bool
	}{
		{
			name:      "equal integers",
			a:         newOrFatal(t, 10),
			b:         newOrFatal(t, 10),
			wantEqual: true,
		},
		{
			name:      "unequal integers",
			a:         newOrFatal(t, 10),
			b:         newOrFatal(t, 20),
			wantEqual: false,
		},
		{
			name:      "equal bool",
			a:         newOrFatal(t, true),
			b:         newOrFatal(t, true),
			wantEqual: true,
		},
		{
			name:      "unequal bool",
			a:         newOrFatal(t, true),
			b:         newOrFatal(t, false),
			wantEqual: false,
		},
		{
			name:      "equal string",
			a:         newOrFatal(t, "hello"),
			b:         newOrFatal(t, "hello"),
			wantEqual: true,
		},
		{
			name:      "equal string",
			a:         newOrFatal(t, "hello"),
			b:         newOrFatal(t, "Hello"),
			wantEqual: false,
		},
		{
			name:      "equal long",
			a:         newOrFatal(t, 10),
			b:         newOrFatal(t, 10),
			wantEqual: true,
		},
		{
			name:      "unequal long",
			a:         newOrFatal(t, 10),
			b:         newOrFatal(t, 20),
			wantEqual: false,
		},
		{
			name:      "equal decimal",
			a:         newOrFatal(t, 10.0000001),
			b:         newOrFatal(t, 10.0000001),
			wantEqual: true,
		},
		{
			name:      "unequal decimal",
			a:         newOrFatal(t, 10.0000001),
			b:         newOrFatal(t, 10.0000002),
			wantEqual: false,
		},
		{
			name:      "equal proto",
			a:         newOrFatal(t, Named{Value: &r4patientpb.Patient{Id: &d4pb.Id{Value: "1"}}, RuntimeType: &types.Named{TypeName: "FHIR.Patient"}}),
			b:         newOrFatal(t, Named{Value: &r4patientpb.Patient{Id: &d4pb.Id{Value: "1"}}, RuntimeType: &types.Named{TypeName: "FHIR.Patient"}}),
			wantEqual: true,
		},
		{
			name:      "unequal proto",
			a:         newOrFatal(t, Named{Value: &r4patientpb.Patient{Id: &d4pb.Id{Value: "1"}}, RuntimeType: &types.Named{TypeName: "FHIR.Patient"}}),
			b:         newOrFatal(t, Named{Value: &r4patientpb.Patient{Id: &d4pb.Id{Value: "2"}}, RuntimeType: &types.Named{TypeName: "FHIR.Patient"}}),
			wantEqual: false,
		},
		{
			name: "equal tuples",
			a: newOrFatal(t, Tuple{
				Value:       map[string]Value{"Apple": newOrFatal(t, 10), "Banana": newOrFatal(t, 20)},
				RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{"Apple": types.Integer, "Banana": types.Integer}},
			}),
			b: newOrFatal(t, Tuple{
				Value:       map[string]Value{"Apple": newOrFatal(t, 10), "Banana": newOrFatal(t, 20)},
				RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{"Apple": types.Integer, "Banana": types.Integer}},
			}),
			wantEqual: true,
		},
		{
			name: "unequal tuples value",
			a: newOrFatal(t, Tuple{
				Value:       map[string]Value{"Apple": newOrFatal(t, 10), "Banana": newOrFatal(t, 20)},
				RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{"Apple": types.Integer, "Banana": types.Integer}},
			}),
			b: newOrFatal(t, Tuple{
				Value:       map[string]Value{"Apple": newOrFatal(t, 20), "Banana": newOrFatal(t, 10)},
				RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{"Apple": types.Integer, "Banana": types.Integer}},
			}),
			wantEqual: false,
		},
		{
			name: "unequal tuples length",
			a: newOrFatal(t, Tuple{
				Value:       map[string]Value{"Apple": newOrFatal(t, 10)},
				RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{"Apple": types.Integer, "Banana": types.Integer}},
			}),
			b: newOrFatal(t, Tuple{
				Value:       map[string]Value{"Apple": newOrFatal(t, 20), "Banana": newOrFatal(t, 10)},
				RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{"Apple": types.Integer, "Banana": types.Integer}},
			}),
			wantEqual: false,
		},
		{
			name: "unequal tuples type",
			a: newOrFatal(t, Tuple{
				Value:       map[string]Value{"Apple": newOrFatal(t, 20), "Banana": newOrFatal(t, 10)},
				RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{"Apple": types.Integer, "Banana": types.Integer}},
			}),
			b: newOrFatal(t, Tuple{
				Value:       map[string]Value{"Apple": newOrFatal(t, 20), "Banana": newOrFatal(t, 10)},
				RuntimeType: &types.Named{TypeName: "FHIR.Fruit"},
			}),
			wantEqual: false,
		},
		{
			name:      "equal list",
			a:         newOrFatal(t, List{Value: []Value{newOrFatal(t, 10), newOrFatal(t, 20)}, StaticType: &types.List{ElementType: types.Integer}}),
			b:         newOrFatal(t, List{Value: []Value{newOrFatal(t, 10), newOrFatal(t, 20)}, StaticType: &types.List{ElementType: types.Integer}}),
			wantEqual: true,
		},
		{
			name:      "unequal list",
			a:         newOrFatal(t, List{Value: []Value{newOrFatal(t, 20), newOrFatal(t, 20)}, StaticType: &types.List{ElementType: types.Integer}}),
			b:         newOrFatal(t, List{Value: []Value{newOrFatal(t, 10), newOrFatal(t, 20)}, StaticType: &types.List{ElementType: types.Integer}}),
			wantEqual: false,
		},
		{
			name:      "unequal list length",
			a:         newOrFatal(t, List{Value: []Value{newOrFatal(t, 20), newOrFatal(t, 20)}, StaticType: &types.List{ElementType: types.Integer}}),
			b:         newOrFatal(t, List{Value: []Value{newOrFatal(t, 10), newOrFatal(t, 20), newOrFatal(t, 20)}, StaticType: &types.List{ElementType: types.Integer}}),
			wantEqual: false,
		},
		{
			name:      "equal null",
			a:         newOrFatal(t, nil),
			b:         newOrFatal(t, nil),
			wantEqual: true,
		},
		{
			name:      "equal quantity",
			a:         newOrFatal(t, Quantity{Value: 1, Unit: model.YEARUNIT}),
			b:         newOrFatal(t, Quantity{Value: 1, Unit: model.YEARUNIT}),
			wantEqual: true,
		},
		{
			name:      "unequal quantity with different value",
			a:         newOrFatal(t, Quantity{Value: 1, Unit: model.YEARUNIT}),
			b:         newOrFatal(t, Quantity{Value: 2, Unit: model.YEARUNIT}),
			wantEqual: false,
		},
		{
			name:      "unequal quantity with different unit",
			a:         newOrFatal(t, Quantity{Value: 1, Unit: model.YEARUNIT}),
			b:         newOrFatal(t, Quantity{Value: 1, Unit: model.MONTHUNIT}),
			wantEqual: false,
		},
		{
			name: "equal ratio",
			a: newOrFatal(t, Ratio{
				Numerator:   Quantity{Value: 1, Unit: model.YEARUNIT},
				Denominator: Quantity{Value: 2, Unit: model.YEARUNIT},
			},
			),
			b: newOrFatal(t, Ratio{
				Numerator:   Quantity{Value: 1, Unit: model.YEARUNIT},
				Denominator: Quantity{Value: 2, Unit: model.YEARUNIT},
			},
			),
			wantEqual: true,
		},
		{
			name: "unequal ratio with different quantity unit",
			a: newOrFatal(t, Ratio{
				Numerator:   Quantity{Value: 1, Unit: model.MONTHUNIT},
				Denominator: Quantity{Value: 2, Unit: model.MONTHUNIT},
			},
			),
			b: newOrFatal(t, Ratio{
				Numerator:   Quantity{Value: 1, Unit: model.YEARUNIT},
				Denominator: Quantity{Value: 2, Unit: model.YEARUNIT},
			},
			),
			wantEqual: false,
		},
		{
			name: "unequal ratio with different quantity values",
			a: newOrFatal(t, Ratio{
				Numerator:   Quantity{Value: 44, Unit: model.YEARUNIT},
				Denominator: Quantity{Value: 55, Unit: model.YEARUNIT},
			},
			),
			b: newOrFatal(t, Ratio{
				Numerator:   Quantity{Value: 1, Unit: model.YEARUNIT},
				Denominator: Quantity{Value: 2, Unit: model.YEARUNIT},
			},
			),
			wantEqual: false,
		},
		{
			name:      "equal valueset",
			a:         newOrFatal(t, ValueSet{ID: "ID", Version: "Version"}),
			b:         newOrFatal(t, ValueSet{ID: "ID", Version: "Version"}),
			wantEqual: true,
		},
		{
			name:      "equal valueset with unsorted but equal codesystem",
			a:         newOrFatal(t, ValueSet{ID: "ID", Version: "Version", CodeSystems: []CodeSystem{{ID: "ID1"}, {ID: "ID2"}}}),
			b:         newOrFatal(t, ValueSet{ID: "ID", Version: "Version", CodeSystems: []CodeSystem{{ID: "ID2"}, {ID: "ID1"}}}),
			wantEqual: true,
		},
		{
			name:      "unequal valueset",
			a:         newOrFatal(t, ValueSet{ID: "ID", Version: "Version"}),
			b:         newOrFatal(t, ValueSet{ID: "ID", Version: "Version2"}),
			wantEqual: false,
		},
		{
			name:      "unequal valueset with unequal codesystem",
			a:         newOrFatal(t, ValueSet{ID: "ID", Version: "Version", CodeSystems: []CodeSystem{{ID: "ID"}}}),
			b:         newOrFatal(t, ValueSet{ID: "ID2", Version: "Version", CodeSystems: []CodeSystem{{ID: "ID", Version: "Version"}}}),
			wantEqual: false,
		},
		{
			name:      "unequal valueset with unequal number of codesystem",
			a:         newOrFatal(t, ValueSet{ID: "ID", Version: "Version", CodeSystems: []CodeSystem{{ID: "ID"}}}),
			b:         newOrFatal(t, ValueSet{ID: "ID2", Version: "Version", CodeSystems: []CodeSystem{{ID: "ID"}, {ID: "ID2"}}}),
			wantEqual: false,
		},
		{
			name:      "equal codesystem",
			a:         newOrFatal(t, CodeSystem{ID: "ID", Version: "Version"}),
			b:         newOrFatal(t, CodeSystem{ID: "ID", Version: "Version"}),
			wantEqual: true,
		},
		{
			name:      "unequal codesystem",
			a:         newOrFatal(t, CodeSystem{ID: "ID", Version: "Version"}),
			b:         newOrFatal(t, CodeSystem{ID: "ID", Version: "Version2"}),
			wantEqual: false,
		},
		{
			name:      "equal code",
			a:         newOrFatal(t, Code{System: "System", Code: "Code"}),
			b:         newOrFatal(t, Code{System: "System", Code: "Code"}),
			wantEqual: true,
		},
		{
			name:      "unequal code",
			a:         newOrFatal(t, Code{System: "System", Code: "Code"}),
			b:         newOrFatal(t, Code{System: "System", Code: "Code2"}),
			wantEqual: false,
		},
		{
			name:      "equal concept",
			a:         newOrFatal(t, Concept{Codes: []*Code{{System: "CodeSystem", Code: "Code"}}, Display: "BO"}),
			b:         newOrFatal(t, Concept{Codes: []*Code{{System: "CodeSystem", Code: "Code"}}, Display: "BO"}),
			wantEqual: true,
		},
		{
			name:      "equal concept with null code",
			a:         newOrFatal(t, Concept{Codes: []*Code{nil, {System: "CodeSystem", Code: "Code"}}, Display: "BO"}),
			b:         newOrFatal(t, Concept{Codes: []*Code{nil, {System: "CodeSystem", Code: "Code"}}, Display: "BO"}),
			wantEqual: true,
		},
		{
			name:      "equal concept with unsorted but equal codes",
			a:         newOrFatal(t, Concept{Codes: []*Code{{System: "CodeSystem", Code: "Code"}, {System: "CodeSystem2", Code: "Code2"}}, Display: "BO"}),
			b:         newOrFatal(t, Concept{Codes: []*Code{{System: "CodeSystem2", Code: "Code2"}, {System: "CodeSystem", Code: "Code"}}, Display: "BO"}),
			wantEqual: true,
		},
		{
			name:      "unequal concept different displays",
			a:         newOrFatal(t, Concept{Codes: []*Code{{System: "CodeSystem", Code: "Code"}}, Display: "BO"}),
			b:         newOrFatal(t, Concept{Codes: []*Code{{System: "CodeSystem", Code: "Code2"}}, Display: "Deoderant"}),
			wantEqual: false,
		},
		{
			name:      "unequal concept",
			a:         newOrFatal(t, Concept{Codes: []*Code{{System: "CodeSystem", Code: "Code"}}, Display: "BO"}),
			b:         newOrFatal(t, Concept{Codes: []*Code{{System: "CodeSystem", Code: "Code"}, {System: "CodeSystem2", Code: "Code2"}}, Display: "BO"}),
			wantEqual: false,
		},
		{
			name:      "unequal concept",
			a:         newOrFatal(t, Concept{Codes: []*Code{{System: "CodeSystem", Code: "Code"}}, Display: "BO"}),
			b:         newOrFatal(t, Concept{Codes: []*Code{{System: "CS", Code: "Co"}}, Display: "BO"}),
			wantEqual: false,
		},
		{
			name:      "unequal different Value types: protoMessage, integer",
			a:         newOrFatal(t, Named{Value: &r4patientpb.Patient{}, RuntimeType: &types.Named{TypeName: "FHIR.Patient"}}),
			b:         newOrFatal(t, 10),
			wantEqual: false,
		},
		{
			name:      "unequal different Value types: null, protoMessage",
			a:         newOrFatal(t, nil),
			b:         newOrFatal(t, Named{Value: &r4patientpb.Patient{}, RuntimeType: &types.Named{TypeName: "FHIR.Patient"}}),
			wantEqual: false,
		},
		{
			name:      "equal Date",
			a:         newOrFatal(t, Date{Date: time.Date(2024, time.March, 1, 0, 0, 0, 0, time.UTC), Precision: model.DAY}),
			b:         newOrFatal(t, Date{Date: time.Date(2024, time.March, 1, 0, 0, 0, 0, time.UTC), Precision: model.DAY}),
			wantEqual: true,
		},
		{
			name:      "Date equal with month granularity",
			a:         newOrFatal(t, Date{Date: time.Date(2024, time.March, 1, 0, 0, 0, 0, time.UTC), Precision: model.MONTH}),
			b:         newOrFatal(t, Date{Date: time.Date(2024, time.March, 1, 0, 0, 0, 0, time.UTC), Precision: model.MONTH}),
			wantEqual: true,
		},
		{
			name:      "uneqal Dates",
			a:         newOrFatal(t, Date{Date: time.Date(2024, time.March, 1, 0, 0, 0, 0, time.UTC), Precision: model.DAY}),
			b:         newOrFatal(t, Date{Date: time.Date(2024, time.March, 5, 0, 0, 0, 0, time.UTC), Precision: model.DAY}),
			wantEqual: false,
		},
		{
			name:      "unequal Date with differing precision granularity",
			a:         newOrFatal(t, Date{Date: time.Date(2024, time.March, 1, 0, 0, 0, 0, time.UTC), Precision: model.DAY}),
			b:         newOrFatal(t, Date{Date: time.Date(2024, time.March, 1, 0, 0, 0, 0, time.UTC), Precision: model.MONTH}),
			wantEqual: false,
		},
		{
			name:      "Date unequal with another type",
			a:         newOrFatal(t, Date{Date: time.Date(2024, time.March, 1, 0, 0, 0, 0, time.UTC), Precision: model.DAY}),
			b:         newOrFatal(t, 10.0000001),
			wantEqual: false,
		},
		{
			name:      "equal DateTimes",
			a:         newOrFatal(t, DateTime{Date: time.Date(2024, time.March, 31, 1, 20, 30, 1e8, time.UTC)}),
			b:         newOrFatal(t, DateTime{Date: time.Date(2024, time.March, 31, 1, 20, 30, 1e8, time.UTC)}),
			wantEqual: true,
		},
		{
			name:      "DateTime equal with year granularity",
			a:         newOrFatal(t, DateTime{Date: time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC), Precision: model.YEAR}),
			b:         newOrFatal(t, DateTime{Date: time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC), Precision: model.YEAR}),
			wantEqual: true,
		},
		{
			name:      "unequal DateTimes",
			a:         newOrFatal(t, DateTime{Date: time.Date(2024, time.March, 31, 0, 2, 3, 1e8, time.UTC)}),
			b:         newOrFatal(t, DateTime{Date: time.Date(2024, time.March, 31, 1, 20, 30, 1e8, time.UTC)}),
			wantEqual: false,
		},
		{
			name:      "unequal DateTime with differing precision granularity",
			a:         newOrFatal(t, DateTime{Date: time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC), Precision: model.YEAR}),
			b:         newOrFatal(t, DateTime{Date: time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC), Precision: model.MONTH}),
			wantEqual: false,
		},
		{
			name:      "DateTime unequal with another type",
			a:         newOrFatal(t, DateTime{Date: time.Date(2024, time.March, 31, 0, 2, 3, 1e8, time.UTC)}),
			b:         newOrFatal(t, 10),
			wantEqual: false,
		},
		{
			name:      "equal Times",
			a:         newOrFatal(t, Time{Date: time.Date(0, time.January, 1, 1, 20, 30, 1e8, time.UTC)}),
			b:         newOrFatal(t, Time{Date: time.Date(0, time.January, 1, 1, 20, 30, 1e8, time.UTC)}),
			wantEqual: true,
		},
		{
			name:      "Times equal with hour granularity",
			a:         newOrFatal(t, Time{Date: time.Date(0, time.January, 1, 4, 0, 0, 0, time.UTC), Precision: model.HOUR}),
			b:         newOrFatal(t, Time{Date: time.Date(0, time.January, 1, 4, 0, 0, 0, time.UTC), Precision: model.HOUR}),
			wantEqual: true,
		},
		{
			name:      "unequal Times",
			a:         newOrFatal(t, Time{Date: time.Date(0, time.January, 1, 0, 2, 3, 1e8, time.UTC)}),
			b:         newOrFatal(t, Time{Date: time.Date(0, time.January, 1, 1, 20, 30, 1e8, time.UTC)}),
			wantEqual: false,
		},
		{
			name:      "unequal Times with differing precision granularity",
			a:         newOrFatal(t, Time{Date: time.Date(0, time.January, 1, 2, 0, 0, 0, time.UTC), Precision: model.HOUR}),
			b:         newOrFatal(t, Time{Date: time.Date(0, time.January, 1, 2, 0, 0, 0, time.UTC), Precision: model.MINUTE}),
			wantEqual: false,
		},
		{
			name:      "DateTime unequal with another type",
			a:         newOrFatal(t, Time{Date: time.Date(0, time.January, 1, 0, 2, 3, 1e8, time.UTC)}),
			b:         newOrFatal(t, 10),
			wantEqual: false,
		},
		{
			name: "equal Intervals",
			a: newOrFatal(t, Interval{
				Low:           newOrFatal(t, 10),
				High:          newOrFatal(t, 20),
				LowInclusive:  true,
				HighInclusive: true,
				StaticType:    &types.Interval{PointType: types.Integer},
			}),
			b: newOrFatal(t, Interval{
				Low:           newOrFatal(t, 10),
				High:          newOrFatal(t, 20),
				LowInclusive:  true,
				HighInclusive: true,
				StaticType:    &types.Interval{PointType: types.Integer},
			}),
			wantEqual: true,
		},
		{
			name: "unequal Interval Low Value",
			a: newOrFatal(t, Interval{
				Low:           newOrFatal(t, 15),
				High:          newOrFatal(t, 20),
				LowInclusive:  true,
				HighInclusive: true,
				StaticType:    &types.Interval{PointType: types.Integer},
			}),
			b: newOrFatal(t, Interval{
				Low:           newOrFatal(t, 10),
				High:          newOrFatal(t, 20),
				LowInclusive:  true,
				HighInclusive: true,
				StaticType:    &types.Interval{PointType: types.Integer},
			}),
			wantEqual: false,
		},
		{
			name: "unequal Interval High Value",
			a: newOrFatal(t, Interval{
				Low:           newOrFatal(t, 10),
				High:          newOrFatal(t, 25),
				LowInclusive:  true,
				HighInclusive: true,
				StaticType:    &types.Interval{PointType: types.Integer},
			}),
			b: newOrFatal(t, Interval{
				Low:           newOrFatal(t, 10),
				High:          newOrFatal(t, 20),
				LowInclusive:  true,
				HighInclusive: true,
				StaticType:    &types.Interval{PointType: types.Integer},
			}),
			wantEqual: false,
		},
		{
			name: "unequal Interval LowInclusive",
			a: newOrFatal(t, Interval{
				Low:           newOrFatal(t, 10),
				High:          newOrFatal(t, 20),
				LowInclusive:  true,
				HighInclusive: true,
				StaticType:    &types.Interval{PointType: types.Integer},
			}),
			b: newOrFatal(t, Interval{
				Low:           newOrFatal(t, 10),
				High:          newOrFatal(t, 20),
				LowInclusive:  false,
				HighInclusive: true,
				StaticType:    &types.Interval{PointType: types.Integer},
			}),
			wantEqual: false,
		},
		{
			name: "unequal Interval HighInclusive",
			a: newOrFatal(t, Interval{
				Low:           newOrFatal(t, 10),
				High:          newOrFatal(t, 20),
				LowInclusive:  true,
				HighInclusive: true,
				StaticType:    &types.Interval{PointType: types.Integer},
			}),
			b: newOrFatal(t, Interval{
				Low:           newOrFatal(t, 10),
				High:          newOrFatal(t, 20),
				LowInclusive:  true,
				HighInclusive: false,
				StaticType:    &types.Interval{PointType: types.Integer},
			}),
			wantEqual: false,
		},
		{
			name: "unequal Interval value Types",
			a: newOrFatal(t, Interval{
				Low:           newOrFatal(t, 10),
				High:          newOrFatal(t, 20),
				LowInclusive:  true,
				HighInclusive: true,
				StaticType:    &types.Interval{PointType: types.Long},
			}),
			b: newOrFatal(t, Interval{
				Low:           newOrFatal(t, 10),
				High:          newOrFatal(t, 20),
				LowInclusive:  true,
				HighInclusive: false,
				StaticType:    &types.Interval{PointType: types.Integer},
			}),
			wantEqual: false,
		},
		{
			name: "unequal Value Types, with Interval",
			a: newOrFatal(t, Interval{
				Low:           newOrFatal(t, 10),
				High:          newOrFatal(t, 20),
				LowInclusive:  true,
				HighInclusive: false,
				StaticType:    &types.Interval{PointType: types.Integer},
			}),
			b:         newOrFatal(t, 10),
			wantEqual: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.a.Equal(tc.b) != tc.wantEqual {
				t.Errorf("Equal(%v, %v) = %v, want %v", tc.a, tc.b, tc.a, tc.wantEqual)
			}
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  Value
	}{
		{
			name:  "nil",
			input: nil,
			want:  Value{goValue: nil, runtimeType: types.Any},
		},
		{
			name:  "quantity",
			input: Quantity{Value: 1, Unit: model.DAYUNIT},
			want:  Value{goValue: Quantity{Value: 1, Unit: model.DAYUNIT}, runtimeType: types.Quantity},
		},
		{
			name:  "ratio",
			input: Ratio{Numerator: Quantity{Value: 1, Unit: model.DAYUNIT}, Denominator: Quantity{Value: 2, Unit: model.DAYUNIT}},
			want:  Value{goValue: Ratio{Numerator: Quantity{Value: 1, Unit: model.DAYUNIT}, Denominator: Quantity{Value: 2, Unit: model.DAYUNIT}}, runtimeType: types.Ratio},
		},
		{
			name:  "bool",
			input: true,
			want:  Value{goValue: true, runtimeType: types.Boolean},
		},
		{
			name:  "string",
			input: "hello",
			want:  Value{goValue: "hello", runtimeType: types.String},
		},
		{
			name:  "int",
			input: 1,
			want:  Value{goValue: int32(1), runtimeType: types.Integer},
		},
		{
			name:  "long",
			input: int64(1),
			want:  Value{goValue: int64(1), runtimeType: types.Long},
		},
		{
			name:  "decimal",
			input: 1.1,
			want:  Value{goValue: 1.1, runtimeType: types.Decimal},
		},
		{
			name:  "valueset",
			input: ValueSet{ID: "ID", Version: "Version"},
			want:  Value{goValue: ValueSet{ID: "ID", Version: "Version"}, runtimeType: types.ValueSet},
		},
		{
			name:  "codesystem",
			input: CodeSystem{ID: "ID", Version: "Version"},
			want:  Value{goValue: CodeSystem{ID: "ID", Version: "Version"}, runtimeType: types.CodeSystem},
		},
		{
			name:  "code",
			input: Code{System: "System", Code: "Code"},
			want:  Value{goValue: Code{System: "System", Code: "Code"}, runtimeType: types.Code},
		},
		{
			name:  "concept",
			input: Concept{Codes: []*Code{{System: "System", Code: "Code"}}, Display: "A disease"},
			want:  Value{goValue: Concept{Codes: []*Code{{System: "System", Code: "Code"}}, Display: "A disease"}, runtimeType: types.Concept},
		},
		{
			name: "tuple",
			input: Tuple{
				Value:       map[string]Value{"Apple": newOrFatal(t, 10), "Banana": newOrFatal(t, 20)},
				RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{"Apple": types.Integer, "Banana": types.Integer}},
			},
			want: Value{
				goValue: Tuple{
					Value:       map[string]Value{"Apple": newOrFatal(t, 10), "Banana": newOrFatal(t, 20)},
					RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{"Apple": types.Integer, "Banana": types.Integer}},
				},
				runtimeType: &types.Tuple{ElementTypes: map[string]types.IType{"Apple": types.Integer, "Banana": types.Integer}}},
		},
		{
			name:  "list of integer Values",
			input: List{Value: []Value{newOrFatal(t, 1)}, StaticType: &types.List{ElementType: types.Integer}},
			want:  Value{goValue: List{Value: []Value{newOrFatal(t, 1)}, StaticType: &types.List{ElementType: types.Integer}}, runtimeType: &types.List{ElementType: types.Integer}},
		},
		{
			name:  "list of integer literals",
			input: List{Value: []Value{newOrFatal(t, 1), newOrFatal(t, 2)}, StaticType: &types.List{ElementType: types.Integer}},
			want:  Value{goValue: List{Value: []Value{newOrFatal(t, 1), newOrFatal(t, 2)}, StaticType: &types.List{ElementType: types.Integer}}, runtimeType: &types.List{ElementType: types.Integer}},
		},
		{
			name:  "protoValue",
			input: Named{Value: &r4patientpb.Patient{Id: &d4pb.Id{Value: "1"}}, RuntimeType: &types.Named{TypeName: "FHIR.Patient"}},
			want: Value{
				goValue:     Named{Value: &r4patientpb.Patient{Id: &d4pb.Id{Value: "1"}}, RuntimeType: &types.Named{TypeName: "FHIR.Patient"}},
				runtimeType: &types.Named{TypeName: "FHIR.Patient"},
			},
		},
		{
			name: "DateValue",
			input: Date{
				Date:      time.Date(2023, time.March, 5, 0, 0, 0, 0, time.UTC),
				Precision: model.DAY,
			},
			want: Value{
				goValue: Date{
					Date:      time.Date(2023, time.March, 5, 0, 0, 0, 0, time.UTC),
					Precision: model.DAY,
				},
				runtimeType: types.Date,
			},
		},
		{
			name: "DateTimeValue",
			input: DateTime{
				Date:      time.Date(2024, time.March, 31, 1, 20, 30, 1e8, time.UTC),
				Precision: model.SECOND,
			},
			want: Value{
				goValue: DateTime{
					Date:      time.Date(2024, time.March, 31, 1, 20, 30, 1e8, time.UTC),
					Precision: model.SECOND,
				},
				runtimeType: types.DateTime,
			},
		},
		{
			name: "TimeValue",
			input: Time{
				Date:      time.Date(0, time.January, 1, 1, 20, 30, 1e8, time.FixedZone("Fixed", 4*60*60)),
				Precision: model.SECOND,
			},
			want: Value{
				goValue: Time{
					Date:      time.Date(0, time.January, 1, 1, 20, 30, 1e8, time.FixedZone("Fixed", 4*60*60)),
					Precision: model.SECOND,
				},
				runtimeType: types.Time,
			},
		},
		{
			name: "IntervalValue",
			input: Interval{
				Low:           newOrFatal(t, 10),
				High:          newOrFatal(t, 20),
				LowInclusive:  true,
				HighInclusive: true,
				StaticType:    &types.Interval{PointType: types.Integer},
			},
			want: Value{
				goValue: Interval{
					Low:           newOrFatal(t, 10),
					High:          newOrFatal(t, 20),
					LowInclusive:  true,
					HighInclusive: true,
					StaticType:    &types.Interval{PointType: types.Integer},
				},
				runtimeType: nil,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := New(tc.input)
			if err != nil {
				t.Errorf("New(%v) returned unexpected error, %v", tc.input, err)
			}
			if diff := cmp.Diff(tc.want, got, protocmp.Transform()); diff != "" {
				t.Errorf("New(%v) returned unexpected diff (-want +got):\n%s", tc.input, diff)
			}
		})
	}
}

func TestNew_Error(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		wantErr string
	}{
		{
			name: "Date unsupported precision",
			input: Date{
				Date:      time.Date(2024, time.March, 1, 2, 3, 0, 0, time.UTC),
				Precision: model.SECOND,
			},
			wantErr: datehelpers.ErrUnsupportedPrecision.Error(),
		},
		{
			name: "DateTime unsupported precision",
			input: DateTime{
				Date:      time.Date(2024, time.March, 1, 2, 3, 0, 0, time.UTC),
				Precision: model.WEEK,
			},
			wantErr: datehelpers.ErrUnsupportedPrecision.Error(),
		},
		{
			name: "Time unsupported precision",
			input: Time{
				Date:      time.Date(0, time.January, 1, 2, 3, 0, 0, time.UTC),
				Precision: model.DAY,
			},
			wantErr: datehelpers.ErrUnsupportedPrecision.Error(),
		},
		{
			name: "Time unsupported precision",
			input: Time{
				Date:      time.Date(2024, time.January, 1, 2, 3, 0, 0, time.UTC),
				Precision: model.MINUTE,
			},
			wantErr: "Time must be Year 0000, Month 01, Day 01",
		},
		{
			name:    "CodeSystem missing ID",
			input:   CodeSystem{},
			wantErr: "System.CodeSystem must have an ID",
		},
		{
			name:    "Concept must specify codes",
			input:   Concept{},
			wantErr: "System.Concept must specify the codes field",
		},
		{
			name:    "ValueSet missing ID",
			input:   ValueSet{},
			wantErr: "System.ValueSet must have an ID",
		},
		{
			name:    "unsupported type",
			input:   map[string]string{"test": "test"},
			wantErr: errUnsupportedType.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := New(tc.input)
			if err == nil {
				t.Fatalf("New(%v) succeeded, want error", tc.input)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("Returned error (%s) did not contain expected (%s)", err, tc.wantErr)
			}
		})
	}
}

func TestNewWithSources(t *testing.T) {
	defaultSourceObs := []Value{newOrFatal(t, "PLACEHOLDER")}
	defaultSourceExpr := &model.Add{}
	tests := []struct {
		name  string
		input any
		want  Value
	}{
		{
			name:  "nil",
			input: nil,
			want:  Value{sourceExpr: defaultSourceExpr, sourceVals: defaultSourceObs, goValue: nil, runtimeType: types.Any},
		},
		{
			name:  "bool",
			input: true,
			want:  Value{sourceExpr: defaultSourceExpr, sourceVals: defaultSourceObs, goValue: true, runtimeType: types.Boolean},
		},
		{
			name:  "string",
			input: "hello",
			want:  Value{sourceExpr: defaultSourceExpr, sourceVals: defaultSourceObs, goValue: "hello", runtimeType: types.String},
		},
		{
			name:  "int",
			input: 1,
			want:  Value{sourceExpr: defaultSourceExpr, sourceVals: defaultSourceObs, goValue: int32(1), runtimeType: types.Integer},
		},
		{
			name:  "long",
			input: int64(1),
			want:  Value{sourceExpr: defaultSourceExpr, sourceVals: defaultSourceObs, goValue: int64(1), runtimeType: types.Long},
		},
		{
			name:  "decimal",
			input: 1.1,
			want:  Value{sourceExpr: defaultSourceExpr, sourceVals: defaultSourceObs, goValue: 1.1, runtimeType: types.Decimal},
		},
		{
			name:  "quantity",
			input: Quantity{Value: 1, Unit: model.DAYUNIT},
			want:  Value{sourceExpr: defaultSourceExpr, sourceVals: defaultSourceObs, goValue: Quantity{Value: 1, Unit: model.DAYUNIT}, runtimeType: types.Quantity},
		},
		{
			name:  "ratio",
			input: Ratio{Numerator: Quantity{Value: 1, Unit: model.DAYUNIT}, Denominator: Quantity{Value: 2, Unit: model.DAYUNIT}},
			want: Value{
				goValue:     Ratio{Numerator: Quantity{Value: 1, Unit: model.DAYUNIT}, Denominator: Quantity{Value: 2, Unit: model.DAYUNIT}},
				runtimeType: types.Ratio,
				sourceExpr:  defaultSourceExpr,
				sourceVals:  defaultSourceObs,
			},
		},
		{
			name:  "valueset",
			input: ValueSet{ID: "ID", Version: "Version"},
			want:  Value{sourceExpr: defaultSourceExpr, sourceVals: defaultSourceObs, goValue: ValueSet{ID: "ID", Version: "Version"}, runtimeType: types.ValueSet},
		},
		{
			name:  "codesystem",
			input: CodeSystem{ID: "ID", Version: "Version"},
			want:  Value{sourceExpr: defaultSourceExpr, sourceVals: defaultSourceObs, goValue: CodeSystem{ID: "ID", Version: "Version"}, runtimeType: types.CodeSystem},
		},
		{
			name:  "concept",
			input: Concept{Codes: []*Code{{System: "System", Code: "Code"}}, Display: "A disease"},
			want:  Value{sourceExpr: defaultSourceExpr, sourceVals: defaultSourceObs, goValue: Concept{Codes: []*Code{{System: "System", Code: "Code"}}, Display: "A disease"}, runtimeType: types.Concept},
		},
		{
			name:  "concept with null codes",
			input: Concept{Codes: []*Code{nil}},
			want:  Value{sourceExpr: defaultSourceExpr, sourceVals: defaultSourceObs, goValue: Concept{Codes: []*Code{nil}}, runtimeType: types.Concept},
		},
		{
			name:  "list of integer Values",
			input: List{Value: []Value{newOrFatal(t, 1)}, StaticType: &types.List{ElementType: types.Integer}},
			want:  Value{sourceExpr: defaultSourceExpr, sourceVals: defaultSourceObs, goValue: List{Value: []Value{newOrFatal(t, 1)}, StaticType: &types.List{ElementType: types.Integer}}, runtimeType: &types.List{ElementType: types.Integer}},
		},
		{
			name:  "ProtoValue",
			input: Named{Value: &r4patientpb.Patient{Id: &d4pb.Id{Value: "1"}}, RuntimeType: &types.Named{TypeName: "FHIR.Patient"}},
			want: Value{
				sourceExpr:  defaultSourceExpr,
				sourceVals:  defaultSourceObs,
				goValue:     Named{Value: &r4patientpb.Patient{Id: &d4pb.Id{Value: "1"}}, RuntimeType: &types.Named{TypeName: "FHIR.Patient"}},
				runtimeType: &types.Named{TypeName: "FHIR.Patient"},
			},
		},
		{
			name: "tuple",
			input: Tuple{
				Value:       map[string]Value{"Apple": newOrFatal(t, 10), "Banana": newOrFatal(t, 20)},
				RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{"Apple": types.Integer, "Banana": types.Integer}},
			},
			want: Value{
				sourceExpr: defaultSourceExpr,
				sourceVals: defaultSourceObs,
				goValue: Tuple{
					Value:       map[string]Value{"Apple": newOrFatal(t, 10), "Banana": newOrFatal(t, 20)},
					RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{"Apple": types.Integer, "Banana": types.Integer}},
				},
				runtimeType: &types.Tuple{ElementTypes: map[string]types.IType{"Apple": types.Integer, "Banana": types.Integer}},
			},
		},
		{
			name:  "ListValue",
			input: List{Value: []Value{newOrFatal(t, 1)}, StaticType: &types.List{ElementType: types.Integer}},
			want: Value{
				sourceExpr:  defaultSourceExpr,
				sourceVals:  defaultSourceObs,
				goValue:     List{Value: []Value{newOrFatal(t, 1)}, StaticType: &types.List{ElementType: types.Integer}},
				runtimeType: nil,
			},
		},
		{
			name: "DateValue",
			input: Date{
				Date:      time.Date(2023, time.March, 5, 0, 0, 0, 0, time.UTC),
				Precision: model.DAY,
			},
			want: Value{
				sourceExpr: defaultSourceExpr,
				sourceVals: defaultSourceObs,
				goValue: Date{
					Date:      time.Date(2023, time.March, 5, 0, 0, 0, 0, time.UTC),
					Precision: model.DAY,
				},
				runtimeType: types.Date,
			},
		},
		{
			name: "DateTimeValue",
			input: DateTime{
				Date:      time.Date(2024, time.March, 31, 1, 20, 30, 1e8, time.UTC),
				Precision: model.SECOND,
			},
			want: Value{
				sourceExpr: defaultSourceExpr,
				sourceVals: defaultSourceObs,
				goValue: DateTime{
					Date:      time.Date(2024, time.March, 31, 1, 20, 30, 1e8, time.UTC),
					Precision: model.SECOND,
				},
				runtimeType: types.DateTime,
			},
		},
		{
			name: "TimeValue",
			input: Time{
				Date:      time.Date(0, time.January, 1, 1, 20, 30, 1e8, time.FixedZone("Fixed", 4*60*60)),
				Precision: model.SECOND,
			},
			want: Value{
				sourceExpr: defaultSourceExpr,
				sourceVals: defaultSourceObs,
				goValue: Time{
					Date:      time.Date(0, time.January, 1, 1, 20, 30, 1e8, time.FixedZone("Fixed", 4*60*60)),
					Precision: model.SECOND,
				},
				runtimeType: types.Time,
			},
		},
		{
			name: "IntervalValue",
			input: Interval{
				Low:           newOrFatal(t, 10),
				High:          newOrFatal(t, 20),
				LowInclusive:  true,
				HighInclusive: true,
				StaticType:    &types.Interval{PointType: types.Integer},
			},
			want: Value{
				sourceExpr: defaultSourceExpr,
				sourceVals: defaultSourceObs,
				goValue: Interval{
					Low:           newOrFatal(t, 10),
					High:          newOrFatal(t, 20),
					LowInclusive:  true,
					HighInclusive: true,
					StaticType:    &types.Interval{PointType: types.Integer},
				},
				runtimeType: &types.Interval{PointType: types.Integer},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewWithSources(tc.input, defaultSourceExpr, defaultSourceObs...)
			if err != nil {
				t.Errorf("NewWithSources(%v) returned unexpected error, %v", tc.input, err)
			}
			if diff := cmp.Diff(tc.want, got, protocmp.Transform()); diff != "" {
				t.Errorf("NewWithSources(%v) returned unexpected diff (-want +got):\n%s", tc.input, diff)
			}

			// Test adding sources with a specific source value makes it visible.
			wrappedExpr := &model.Subtract{}
			wrappedSourceObj := newOrFatal(t, "Wrapper")
			wrappedWithSourceObj := got.WithSources(wrappedExpr, wrappedSourceObj)
			if diff := cmp.Diff(tc.want.GolangValue(), wrappedWithSourceObj.GolangValue(), protocmp.Transform()); diff != "" {
				t.Errorf("Wrapped value for %v did not match the original (-want +got):\n%s", tc.input, diff)
			}
			if diff := cmp.Diff(wrappedExpr, wrappedWithSourceObj.SourceExpression()); diff != "" {
				t.Errorf("Wrapped source expression for %v returned unexpected diff (-want +got):\n%s", tc.input, diff)
			}
			if diff := cmp.Diff([]Value{wrappedSourceObj}, wrappedWithSourceObj.SourceValues()); diff != "" {
				t.Errorf("Wrapped source values for %v returned unexpected diff (-want +got):\n%s", tc.input, diff)
			}

			// Test adding sources without a new source value keeps the existing source value.

			wrappedWithoutSourceObj := got.WithSources(wrappedExpr)
			if diff := cmp.Diff(tc.want.GolangValue(), wrappedWithoutSourceObj.GolangValue(), protocmp.Transform()); diff != "" {
				t.Errorf("Wrapped without new source value for %v did not match the original (-want +got):\n%s", tc.input, diff)
			}
			if diff := cmp.Diff(wrappedExpr, wrappedWithoutSourceObj.SourceExpression()); diff != "" {
				t.Errorf("Wrapped source expression without new source for %v returned unexpected diff (-want +got):\n%s", tc.input, diff)
			}
			if diff := cmp.Diff([]Value{got}, wrappedWithoutSourceObj.SourceValues()); diff != "" {
				t.Errorf("Wrapped source values without new source for %v returned unexpected diff (-want +got):\n%s", tc.input, diff)
			}

			// TODO b/301606416: Add tests where we modify the value of the original Value.
			if diff := cmp.Diff(tc.want, got, protocmp.Transform()); diff != "" {
				t.Errorf("Original Value for (%v) should not be modified after wrapping but was: (-want +got):\n%s", tc.input, diff)
			}
		})
	}
}

func TestMarshalJSON(t *testing.T) {
	tests := []struct {
		name         string
		unmarshalled Value
		want         string
	}{
		// Simple types.
		{
			name:         "nil",
			unmarshalled: newOrFatal(t, nil),
			want:         `{"@type":"System.Any","value":null}`,
		},
		{
			name:         "Int",
			unmarshalled: newOrFatal(t, 1),
			want:         `{"@type":"System.Integer","value":1}`,
		},
		{
			name:         "Long",
			unmarshalled: newOrFatal(t, int64(1)),
			want:         `{"@type":"System.Long","value":1}`,
		},
		{
			name:         "Decimal",
			unmarshalled: newOrFatal(t, 4.5),
			want:         `{"@type":"System.Decimal","value":4.5}`,
		},
		{
			name:         "String",
			unmarshalled: newOrFatal(t, "hello"),
			want:         `{"@type":"System.String","value":"hello"}`,
		},
		{
			name:         "Bool",
			unmarshalled: newOrFatal(t, true),
			want:         `{"@type":"System.Boolean","value":true}`,
		},
		{
			name:         "Quantity",
			unmarshalled: newOrFatal(t, Quantity{Value: 1, Unit: model.YEARUNIT}),
			want:         `{"@type":"System.Quantity","value":1,"unit":"year"}`,
		},
		{
			name: "Ratio",
			unmarshalled: newOrFatal(t,
				Ratio{Numerator: Quantity{Value: 1, Unit: model.YEARUNIT}, Denominator: Quantity{Value: 2, Unit: model.YEARUNIT}}),
			want: `{"@type":"System.Ratio","numerator":{"@type":"System.Quantity","value":1,"unit":"year"},"denominator":{"@type":"System.Quantity","value":2,"unit":"year"}}`,
		},
		{
			name:         "Code",
			unmarshalled: newOrFatal(t, Code{System: "foo", Code: "bar", Display: "the foo", Version: "1.0"}),
			want:         `{"@type":"System.Code","code":"bar","display":"the foo","system":"foo","version":"1.0"}`,
		},
		{
			name:         "Valueset",
			unmarshalled: newOrFatal(t, ValueSet{ID: "ID", Version: "Version"}),
			want:         `{"@type":"System.ValueSet","id":"ID","version":"Version"}`,
		},
		{
			name:         "CodeSystem",
			unmarshalled: newOrFatal(t, CodeSystem{ID: "ID", Version: "Version"}),
			want:         `{"@type":"System.CodeSystem","id":"ID","version":"Version"}`,
		},
		{
			name:         "Concept",
			unmarshalled: newOrFatal(t, Concept{Codes: []*Code{{System: "foo", Code: "bar", Version: "1.0"}}, Display: "A disease"}),
			want:         `{"@type":"System.Concept","codes":[{"@type":"System.Code","code":"bar","system":"foo","version":"1.0"}],"display":"A disease"}`,
		},
		{
			name: "Date",
			unmarshalled: newOrFatal(t, Date{
				Date:      time.Date(2024, time.March, 31, 0, 0, 0, 0, time.UTC),
				Precision: model.DAY,
			}),
			want: `{"@type":"System.Date","value":"@2024-03-31"}`,
		},
		{
			name: "DateTime",
			unmarshalled: newOrFatal(t, DateTime{
				Date:      time.Date(2024, time.March, 31, 1, 20, 30, 1e8, time.UTC),
				Precision: model.SECOND,
			}),
			want: `{"@type":"System.DateTime","value":"@2024-03-31T01:20:30Z"}`,
		},
		{
			name: "Time with UTC TimeZone",
			unmarshalled: newOrFatal(t, Time{
				Date:      time.Date(0, time.January, 1, 1, 20, 30, 1e8, time.UTC),
				Precision: model.SECOND,
			}),
			want: `{"@type":"System.Time","value":"T01:20:30"}`,
		},
		{
			name: "Time with TimeZone",
			unmarshalled: newOrFatal(t, Time{
				Date:      time.Date(0, time.January, 1, 1, 20, 30, 1e8, time.FixedZone("Fixed", 4*60*60)),
				Precision: model.MILLISECOND,
			}),
			want: `{"@type":"System.Time","value":"T01:20:30.100"}`,
		},
		// Complex types.
		{
			name: "Interval",
			unmarshalled: newOrFatal(t, Interval{
				Low:           newOrFatal(t, 10),
				High:          newOrFatal(t, 20),
				LowInclusive:  true,
				HighInclusive: true,
				StaticType:    &types.Interval{PointType: types.Integer},
			}),
			want: `{"@type":"Interval\u003cSystem.Integer\u003e","low":{"@type":"System.Integer","value":10},"high":{"@type":"System.Integer","value":20},"lowClosed":true,"highClosed":true}`,
		},
		{
			name: "Tuple",
			unmarshalled: newOrFatal(t, Tuple{
				Value:       map[string]Value{"Apple": newOrFatal(t, 10), "Banana": newOrFatal(t, 20)},
				RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{"Apple": types.Integer, "Banana": types.Integer}},
			}),
			want: `{"Apple":{"@type":"System.Integer","value":10},"Banana":{"@type":"System.Integer","value":20}}`,
		},
		{
			name:         "List",
			unmarshalled: newOrFatal(t, List{Value: []Value{newOrFatal(t, 3), newOrFatal(t, 4)}, StaticType: &types.List{ElementType: types.Integer}}),
			want:         `[{"@type":"System.Integer","value":3},{"@type":"System.Integer","value":4}]`,
		},
		{
			name:         "Proto",
			unmarshalled: newOrFatal(t, Named{Value: &r4patientpb.Patient{Active: &d4pb.Boolean{Value: true}}, RuntimeType: &types.Named{TypeName: "FHIR.Patient"}}),
			want:         `{"@type":"FHIR.Patient","value":{"active":{"value":true}}}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			got, err := json.Marshal(tc.unmarshalled)
			if err != nil {
				t.Fatalf("Json marshalling failed %v", err)
			}
			if diff := cmp.Diff(tc.want, string(got)); diff != "" {
				t.Errorf("json.Marshal() returned unexpected diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRuntimeType(t *testing.T) {
	cases := []struct {
		name            string
		input           Value
		wantRuntimeType types.IType
	}{
		{
			name: "Empty list falls back to static type",
			input: newOrFatal(
				t,
				List{
					Value:      []Value{},
					StaticType: &types.List{ElementType: types.Integer},
				},
			),
			wantRuntimeType: &types.List{ElementType: types.Integer},
		},
		{
			name: "List runtime type is inferred for a non-empty list",
			input: newOrFatal(
				t,
				List{
					Value:      []Value{newOrFatal(t, 3), newOrFatal(t, 4)},
					StaticType: &types.List{ElementType: &types.Choice{ChoiceTypes: []types.IType{types.Integer, types.String}}}},
			),
			wantRuntimeType: &types.List{ElementType: types.Integer},
		},
		{
			name: "List runtime type is inferred from first non-null value",
			input: newOrFatal(
				t,
				List{
					Value:      []Value{newOrFatal(t, nil), newOrFatal(t, nil), newOrFatal(t, 4)},
					StaticType: &types.List{ElementType: &types.Choice{ChoiceTypes: []types.IType{types.Integer}}}},
			),
			wantRuntimeType: &types.List{ElementType: types.Integer},
		},
		{
			name: "Interval with two nulls falls back to static type",
			input: newOrFatal(
				t,
				Interval{
					Low:        newOrFatal(t, nil),
					High:       newOrFatal(t, nil),
					StaticType: &types.Interval{PointType: types.Integer}},
			),
			wantRuntimeType: &types.Interval{PointType: types.Integer},
		},
		{
			name: "Interval runtime type inferred for non-null values",
			input: newOrFatal(
				t,
				Interval{
					Low:        newOrFatal(t, 1),
					High:       newOrFatal(t, nil),
					StaticType: &types.Interval{PointType: &types.Choice{ChoiceTypes: []types.IType{types.Integer, types.Date}}}},
			),
			wantRuntimeType: &types.Interval{PointType: types.Integer},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.input.RuntimeType()
			if !got.Equal(tc.wantRuntimeType) {
				t.Errorf("%v RuntimeType() = %v, want %v", tc.input, got, tc.wantRuntimeType)
			}
		})
	}
}

func TestProtoAndBack(t *testing.T) {
	tests := []struct {
		name      string
		value     Value
		wantProto *crpb.Value
		// wantSameValue is true if the value is expected to be the same as the original value.
		wantSameValue bool
		// If wantSameValue is false, wantDifferentValue is the value that is expected.
		wantDifferentValue Value
	}{
		{
			name:          "Null",
			value:         newOrFatal(t, nil),
			wantProto:     &crpb.Value{},
			wantSameValue: true,
		},
		{
			name:  "Boolean",
			value: newOrFatal(t, true),
			wantProto: &crpb.Value{
				Value: &crpb.Value_BooleanValue{BooleanValue: true},
			},
			wantSameValue: true,
		},
		{
			name:  "String",
			value: newOrFatal(t, "hello"),
			wantProto: &crpb.Value{
				Value: &crpb.Value_StringValue{StringValue: "hello"},
			},
			wantDifferentValue: newOrFatal(t, "hello"),
		},
		{
			name:  "Integer",
			value: newOrFatal(t, 1),
			wantProto: &crpb.Value{
				Value: &crpb.Value_IntegerValue{IntegerValue: 1},
			},
			wantSameValue: true,
		},
		{
			name:  "Long",
			value: newOrFatal(t, int64(1)),
			wantProto: &crpb.Value{
				Value: &crpb.Value_LongValue{LongValue: 1},
			},
			wantSameValue: true,
		},
		{
			name:  "Decimal",
			value: newOrFatal(t, 1.1),
			wantProto: &crpb.Value{
				Value: &crpb.Value_DecimalValue{DecimalValue: 1.1},
			},
			wantSameValue: true,
		},
		{
			name:  "Quantity",
			value: newOrFatal(t, Quantity{Value: 1, Unit: model.YEARUNIT}),
			wantProto: &crpb.Value{
				Value: &crpb.Value_QuantityValue{
					QuantityValue: &crpb.Quantity{
						Value: proto.Float64(1),
						Unit:  proto.String(string(model.YEARUNIT)),
					},
				},
			},
			wantSameValue: true,
		},
		{
			name:  "Ratio",
			value: newOrFatal(t, Ratio{Numerator: Quantity{Value: 1, Unit: model.YEARUNIT}, Denominator: Quantity{Value: 2, Unit: model.YEARUNIT}}),
			wantProto: &crpb.Value{
				Value: &crpb.Value_RatioValue{
					RatioValue: &crpb.Ratio{
						Numerator: &crpb.Quantity{
							Value: proto.Float64(1),
							Unit:  proto.String(string(model.YEARUNIT)),
						},
						Denominator: &crpb.Quantity{
							Value: proto.Float64(2),
							Unit:  proto.String(string(model.YEARUNIT)),
						},
					},
				},
			},
			wantSameValue: true,
		},
		{
			name:  "Date",
			value: newOrFatal(t, Date{Date: time.Date(2024, time.March, 31, 1, 20, 30, 1e8, time.UTC), Precision: model.DAY}),
			wantProto: &crpb.Value{
				Value: &crpb.Value_DateValue{
					DateValue: &crpb.Date{
						Date: &datepb.Date{
							Year:  2024,
							Month: 3,
							Day:   31,
						},
						Precision: crpb.Date_PRECISION_DAY.Enum(),
					},
				},
			},
			// Hours and lower are dropped due to precision.
			wantDifferentValue: newOrFatal(t, Date{Date: time.Date(2024, time.March, 31, 0, 0, 0, 0, time.UTC), Precision: model.DAY}),
		},
		{
			name:  "DateTime",
			value: newOrFatal(t, DateTime{Date: time.Date(2024, time.March, 31, 1, 20, 30, 1e8, time.UTC), Precision: model.SECOND}),
			wantProto: &crpb.Value{
				Value: &crpb.Value_DateTimeValue{
					DateTimeValue: &crpb.DateTime{
						Date:      timestamppb.New(time.Date(2024, time.March, 31, 1, 20, 30, 1e8, time.UTC)),
						Precision: crpb.DateTime_PRECISION_SECOND.Enum(),
					},
				},
			},
			wantSameValue: true,
		},
		{
			name:  "Time",
			value: newOrFatal(t, Time{Date: time.Date(0, time.January, 1, 1, 20, 30, 1e8, time.UTC), Precision: model.SECOND}),
			wantProto: &crpb.Value{
				Value: &crpb.Value_TimeValue{
					TimeValue: &crpb.Time{
						Date: &timeofdaypb.TimeOfDay{
							Hours:   1,
							Minutes: 20,
							Seconds: 30,
						},
						Precision: crpb.Time_PRECISION_SECOND.Enum(),
					},
				},
			},
			// Seconds are dropped due to precision.
			wantDifferentValue: newOrFatal(t, Time{Date: time.Date(0, time.January, 1, 1, 20, 30, 0, time.UTC), Precision: model.SECOND}),
		},
		{
			name:  "Interval",
			value: newOrFatal(t, Interval{Low: newOrFatal(t, 1), High: newOrFatal(t, 2), LowInclusive: true, HighInclusive: true, StaticType: &types.Interval{PointType: types.Integer}}),
			wantProto: &crpb.Value{
				Value: &crpb.Value_IntervalValue{
					IntervalValue: &crpb.Interval{
						Low:           &crpb.Value{Value: &crpb.Value_IntegerValue{IntegerValue: 1}},
						High:          &crpb.Value{Value: &crpb.Value_IntegerValue{IntegerValue: 2}},
						LowInclusive:  proto.Bool(true),
						HighInclusive: proto.Bool(true),
						StaticType:    &ctpb.IntervalType{PointType: &ctpb.CQLType{Type: &ctpb.CQLType_SystemType{SystemType: &ctpb.SystemType{Type: ctpb.SystemType_TYPE_INTEGER.Enum()}}}},
					},
				},
			},
			wantSameValue: true,
		},
		{
			name:  "List",
			value: newOrFatal(t, List{Value: []Value{newOrFatal(t, 1), newOrFatal(t, 2)}, StaticType: &types.List{ElementType: types.Integer}}),
			wantProto: &crpb.Value{
				Value: &crpb.Value_ListValue{
					ListValue: &crpb.List{
						Value: []*crpb.Value{
							{Value: &crpb.Value_IntegerValue{IntegerValue: 1}},
							{Value: &crpb.Value_IntegerValue{IntegerValue: 2}},
						},
						StaticType: &ctpb.ListType{ElementType: &ctpb.CQLType{Type: &ctpb.CQLType_SystemType{SystemType: &ctpb.SystemType{Type: ctpb.SystemType_TYPE_INTEGER.Enum()}}}},
					},
				},
			},
			wantSameValue: true,
		},
		{
			name: "Tuple",
			value: newOrFatal(t, Tuple{
				Value:       map[string]Value{"Apple": newOrFatal(t, 10), "Banana": newOrFatal(t, 20)},
				RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{"Apple": types.Integer, "Banana": types.Integer}},
			}),
			wantProto: &crpb.Value{
				Value: &crpb.Value_TupleValue{
					TupleValue: &crpb.Tuple{
						Value: map[string]*crpb.Value{
							"Apple":  {Value: &crpb.Value_IntegerValue{IntegerValue: 10}},
							"Banana": {Value: &crpb.Value_IntegerValue{IntegerValue: 20}},
						},
						RuntimeType: &crpb.Tuple_TupleType{TupleType: &ctpb.TupleType{ElementTypes: map[string]*ctpb.CQLType{
							"Apple":  {Type: &ctpb.CQLType_SystemType{SystemType: &ctpb.SystemType{Type: ctpb.SystemType_TYPE_INTEGER.Enum()}}},
							"Banana": {Type: &ctpb.CQLType_SystemType{SystemType: &ctpb.SystemType{Type: ctpb.SystemType_TYPE_INTEGER.Enum()}}},
						}}},
					},
				},
			},
			wantSameValue: true,
		},
		{
			name:  "CodeSystem",
			value: newOrFatal(t, CodeSystem{ID: "ID", Version: "Version"}),
			wantProto: &crpb.Value{
				Value: &crpb.Value_CodeSystemValue{
					CodeSystemValue: &crpb.CodeSystem{
						Id:      proto.String("ID"),
						Version: proto.String("Version"),
					},
				},
			},
			wantSameValue: true,
		},
		{
			name:  "ValueSet",
			value: newOrFatal(t, ValueSet{ID: "ID", Version: "Version", CodeSystems: []CodeSystem{{ID: "CSID", Version: "CSVersion"}}}),
			wantProto: &crpb.Value{
				Value: &crpb.Value_ValueSetValue{
					ValueSetValue: &crpb.ValueSet{
						Id:      proto.String("ID"),
						Version: proto.String("Version"),
						CodeSystems: []*crpb.CodeSystem{
							{
								Id:      proto.String("CSID"),
								Version: proto.String("CSVersion"),
							},
						},
					},
				},
			},
			wantSameValue: true,
		},
		{
			name:  "Concept",
			value: newOrFatal(t, Concept{Codes: []*Code{{System: "System", Code: "Code"}}, Display: "A disease"}),
			wantProto: &crpb.Value{
				Value: &crpb.Value_ConceptValue{
					ConceptValue: &crpb.Concept{
						Codes: []*crpb.Code{
							{
								System:  proto.String("System"),
								Code:    proto.String("Code"),
								Display: proto.String(""),
								Version: proto.String(""),
							},
						},
						Display: proto.String("A disease"),
					},
				},
			},
			wantSameValue: true,
		},
		{
			name:  "Code",
			value: newOrFatal(t, Code{System: "System", Code: "Code", Version: "Version", Display: "A disease"}),
			wantProto: &crpb.Value{
				Value: &crpb.Value_CodeValue{
					CodeValue: &crpb.Code{
						System:  proto.String("System"),
						Code:    proto.String("Code"),
						Version: proto.String("Version"),
						Display: proto.String("A disease"),
					},
				},
			},
			wantSameValue: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotProto, err := tc.value.Proto()
			if err != nil {
				t.Errorf("Proto() returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantProto, gotProto, protocmp.Transform()); diff != "" {
				t.Errorf("Proto() returned unexpected diff (-want +got):\n%s", diff)
			}

			gotValue, err := NewFromProto(gotProto)
			if err != nil {
				t.Errorf("NewFromProto() returned unexpected error: %v", err)
			}

			wantValue := gotValue
			if !tc.wantSameValue {
				wantValue = tc.wantDifferentValue
			}
			if diff := cmp.Diff(wantValue, gotValue); diff != "" {
				t.Errorf("NewFromProto() returned unexpected diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestProtoAndBack_Named(t *testing.T) {
	value := newOrFatal(t, Named{Value: &r4patientpb.Patient{Id: &d4pb.Id{Value: "1"}}, RuntimeType: &types.Named{TypeName: "FHIR.Patient"}})
	wantProto := &crpb.Value{
		Value: &crpb.Value_NamedValue{
			NamedValue: &crpb.Named{
				Value:       anyProtoOrFatal(t, &r4patientpb.Patient{Id: &d4pb.Id{Value: "1"}}),
				RuntimeType: &ctpb.NamedType{TypeName: proto.String("FHIR.Patient")},
			},
		},
	}

	gotProto, err := value.Proto()
	if err != nil {
		t.Errorf("Proto() returned unexpected error: %v", err)
	}
	if diff := cmp.Diff(wantProto, gotProto, protocmp.Transform()); diff != "" {
		t.Errorf("Proto() returned unexpected diff (-want +got):\n%s", diff)
	}

	_, err = NewFromProto(gotProto)
	if err == nil {
		t.Errorf("NewFromProto() succeeded, want error")
	}
}

func newOrFatal(t *testing.T, a any) Value {
	t.Helper()
	o, err := New(a)
	if err != nil {
		t.Fatalf("New(%v) returned unexpected error: %v", a, err)
	}
	return o
}

func anyProtoOrFatal(t *testing.T, a proto.Message) *anypb.Any {
	t.Helper()
	o, err := anypb.New(a)
	if err != nil {
		t.Fatalf("anypb.New(%v) returned unexpected error: %v", a, err)
	}
	return o
}
