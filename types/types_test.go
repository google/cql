// Copyright 2023 Google LLC
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

package types

import (
	"strings"
	"testing"

	ctpb "github.com/google/cql/protos/cql_types_go_proto"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestTypeEqual(t *testing.T) {
	tests := []struct {
		name string
		a    IType
		b    IType
		want bool
	}{
		{
			name: "SystemTypes Equal",
			a:    Integer,
			b:    Integer,
			want: true,
		},
		{
			name: "NamedTypes Equal",
			a:    &Named{TypeName: "TypeA"},
			b:    &Named{TypeName: "TypeA"},
			want: true,
		},
		{
			name: "IntervalTypes Equal",
			a:    &Interval{PointType: Integer},
			b:    &Interval{PointType: Integer},
			want: true,
		},
		{
			name: "IntervalTypes Not Equal",
			a:    &Interval{PointType: Integer},
			b:    &Interval{PointType: &Named{TypeName: "TypeA"}},
			want: false,
		},
		{
			name: "SystemType NamedType Not Equal",
			a:    Integer,
			b:    &Named{TypeName: "TypeA"},
			want: false,
		},
		{
			name: "IntervalType NamedType Not Equal",
			a:    &Interval{PointType: &Named{TypeName: "TypeA"}},
			b:    &Named{TypeName: "TypeA"},
			want: false,
		},
		{
			name: "NamedType Empty Equal",
			a:    &Named{TypeName: ""},
			b:    &Named{TypeName: ""},
			want: true,
		},
		{
			name: "IntervalType Nil Equal",
			a:    &Interval{PointType: nil},
			b:    &Interval{PointType: nil},
			want: true,
		},
		{
			name: "ListType Equal",
			a:    &List{ElementType: Integer},
			b:    &List{ElementType: Integer},
			want: true,
		},
		{
			name: "ListType Not Equal",
			a:    &List{ElementType: Integer},
			b:    &List{ElementType: String},
			want: false,
		},
		{
			name: "ListType Nil Equal",
			a:    &List{ElementType: nil},
			b:    &List{ElementType: nil},
			want: true,
		},
		{
			name: "ChoiceType Equal",
			a:    &Choice{ChoiceTypes: []IType{Integer, &Named{TypeName: "TypeA"}}},
			b:    &Choice{ChoiceTypes: []IType{Integer, &Named{TypeName: "TypeA"}}},
			want: true,
		},
		{
			name: "ChoiceType Equal Different Order",
			a:    &Choice{ChoiceTypes: []IType{Integer, &Named{TypeName: "TypeA"}, Integer, String}},
			b:    &Choice{ChoiceTypes: []IType{String, Integer, &Named{TypeName: "TypeA"}, Integer}},
			want: true,
		},
		{
			name: "ChoiceType Not Equal",
			a:    &Choice{ChoiceTypes: []IType{Integer, &Named{TypeName: "TypeA"}}},
			b:    &Choice{ChoiceTypes: []IType{String, &Named{TypeName: "TypeA"}}},
			want: false,
		},
		{
			name: "ChoiceType Different Lengths Not Equal",
			a:    &Choice{ChoiceTypes: []IType{Integer, &Named{TypeName: "TypeA"}}},
			b:    &Choice{ChoiceTypes: []IType{Integer}},
			want: false,
		},
		{
			name: "ChoiceType Empty Equal",
			a:    &Choice{ChoiceTypes: []IType{}},
			b:    &Choice{ChoiceTypes: []IType{}},
			want: true,
		},
		{
			name: "TupleType Equal",
			a:    &Tuple{ElementTypes: map[string]IType{"apple": Integer, "banana": String}},
			b:    &Tuple{ElementTypes: map[string]IType{"banana": String, "apple": Integer}},
			want: true,
		},
		{
			name: "TupleType Not Equal",
			a:    &Tuple{ElementTypes: map[string]IType{"apple": Integer, "banana": &Named{TypeName: "TypeA"}}},
			b:    &Tuple{ElementTypes: map[string]IType{"banana": String, "apple": Integer}},
			want: false,
		},
		{
			name: "TupleType Different Lengths Not Equal",
			a:    &Tuple{ElementTypes: map[string]IType{"apple": Integer}},
			b:    &Tuple{ElementTypes: map[string]IType{"banana": String, "apple": Integer}},
			want: false,
		},
		{
			name: "TupleType Empty Equal",
			a:    &Tuple{ElementTypes: map[string]IType{}},
			b:    &Tuple{ElementTypes: map[string]IType{}},
			want: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.a.Equal(tc.b); got != tc.want {
				t.Errorf("%v.Equal(%v) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestType_String(t *testing.T) {
	tests := []struct {
		name  string
		given IType
		want  string
	}{
		{
			name:  "System",
			given: Integer,
			want:  "System.Integer",
		},
		{
			name:  "Named",
			given: &Named{TypeName: "TypeA"},
			want:  "Named<TypeA>",
		},
		{
			name:  "Interval",
			given: &Interval{PointType: Integer},
			want:  "Interval<System.Integer>",
		},
		{
			name:  "List<Integer>",
			given: &List{ElementType: Integer},
			want:  "List<System.Integer>",
		},
		{
			name: "List<List<System.Integer>>",
			given: &List{
				ElementType: &List{ElementType: Integer},
			},
			want: "List<List<System.Integer>>",
		},
		{
			name: "List<Interval<System.Integer>>",
			given: &List{
				ElementType: &Interval{PointType: Integer},
			},
			want: "List<Interval<System.Integer>>",
		},
		{
			name:  "Interval<nil>",
			given: &Interval{PointType: nil},
			want:  "Interval<nil>",
		},
		{
			name:  "List<nil>",
			given: &List{ElementType: nil},
			want:  "List<nil>",
		},
		{
			name:  "Choice<System.Integer, System.String>",
			given: &Choice{ChoiceTypes: []IType{Integer, String}},
			want:  "Choice<System.Integer, System.String>",
		},
		{
			name:  "Choice<>",
			given: &Choice{ChoiceTypes: []IType{}},
			want:  "Choice<>",
		},
		{
			name:  "Tuple<System.Integer, System.String>",
			given: &Tuple{ElementTypes: map[string]IType{"apple": Integer, "banana": String}},
			want:  "Tuple<apple: System.Integer, banana: System.String>",
		},
		{
			name:  "Tuple<>",
			given: &Tuple{ElementTypes: map[string]IType{}},
			want:  "Tuple<>",
		},
		{
			name:  "Tuple<nil>",
			given: &Tuple{ElementTypes: nil},
			want:  "Tuple<nil>",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.given.String(); got != tc.want {
				t.Errorf("String() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestNilTypeString(t *testing.T) {
	t.Run("Nil Named", func(t *testing.T) {
		var l *Named
		var i IType
		i = l
		want := "nil Named"
		if got := i.String(); got != want {
			t.Errorf("%v.GetName() = %v, want %v", i, got, want)
		}

		if got := i.Equal(l); got != true {
			t.Errorf("%v.Equal(%v) = %v, want true", i, l, got)
		}
	})

	t.Run("Nil Interval", func(t *testing.T) {
		var l *Interval
		var i IType
		i = l
		want := "nil Interval"
		if got := i.String(); got != want {
			t.Errorf("%v.GetName() = %v, want %v", i, got, want)
		}

		if got := i.Equal(l); got != true {
			t.Errorf("%v.Equal(%v) = %v, want true", i, l, got)
		}
	})

	t.Run("Nil List", func(t *testing.T) {
		var l *List
		var i IType
		i = l
		want := "nil List"
		if got := i.String(); got != want {
			t.Errorf("%v.GetName() = %v, want %v", i, got, want)
		}

		if got := i.Equal(l); got != true {
			t.Errorf("%v.Equal(%v) = %v, want true", i, l, got)
		}
	})

	t.Run("Nil Choice", func(t *testing.T) {
		var l *Choice
		var i IType
		i = l
		want := "nil Choice"
		if got := i.String(); got != want {
			t.Errorf("%v.GetName() = %v, want %v", i, got, want)
		}

		if got := i.Equal(l); got != true {
			t.Errorf("%v.Equal(%v) = %v, want true", i, l, got)
		}
	})

	t.Run("Nil Tuple", func(t *testing.T) {
		var l *Tuple
		var i IType
		i = l
		want := "nil Tuple"
		if got := i.String(); got != want {
			t.Errorf("%v.GetName() = %v, want %v", i, got, want)
		}

		if got := i.Equal(l); got != true {
			t.Errorf("%v.Equal(%v) = %v, want true", i, l, got)
		}
	})
}

func TestType_Name(t *testing.T) {
	tests := []struct {
		name  string
		given IType
		want  string
	}{
		{
			name:  "System",
			given: Integer,
			want:  "System.Integer",
		},
		{
			name:  "Named",
			given: &Named{TypeName: "FHIR.TypeA"},
			want:  "FHIR.TypeA",
		},
		{
			name:  "Interval",
			given: &Interval{PointType: Integer},
			want:  "Interval<System.Integer>",
		},
		{
			name:  "List<Integer>",
			given: &List{ElementType: Integer},
			want:  "List<System.Integer>",
		},
		{
			name: "List<List<FHIR.TypeA>>",
			given: &List{
				ElementType: &List{ElementType: &Named{TypeName: "FHIR.TypeA"}},
			},
			want: "List<List<FHIR.TypeA>>",
		},
		{
			name: "List<Interval<System.Integer>>",
			given: &List{
				ElementType: &Interval{PointType: &Named{TypeName: "FHIR.TypeA"}},
			},
			want: "List<Interval<FHIR.TypeA>>",
		},
		{
			name:  "Choice Types are sorted",
			given: &Choice{ChoiceTypes: []IType{Integer, &Named{TypeName: "FHIR.TypeA"}}},
			want:  "Choice<FHIR.TypeA, System.Integer>",
		},
		{
			name:  "Choice Type single choice",
			given: &Choice{ChoiceTypes: []IType{Integer}},
			want:  "Choice<System.Integer>",
		},
		{
			name:  "Choice is empty",
			given: &Choice{ChoiceTypes: []IType{}},
			want:  "Choice<>",
		},
		{
			name: "Tuple Types are sorted by name",
			given: &Tuple{ElementTypes: map[string]IType{
				"Banana": &Named{TypeName: "FHIR.TypeA"},
				"Apple":  Integer,
			}},
			want: "Tuple { Apple System.Integer, Banana FHIR.TypeA }",
		},
		{
			name: "Tuple Types single element",
			given: &Tuple{ElementTypes: map[string]IType{
				"Apple": Integer,
			}},
			want: "Tuple { Apple System.Integer }",
		},
		{
			name:  "Tuple is empty",
			given: &Tuple{ElementTypes: map[string]IType{}},
			want:  "Tuple { }",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.given.ModelInfoName()
			if err != nil {
				t.Errorf("Name() unexpected err: %v", err)
			}
			if got != tc.want {
				t.Errorf("Name() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestType_NameErrors(t *testing.T) {
	cases := []struct {
		name            string
		nilType         IType
		wantErrContains string
	}{
		{
			name:            "Nil Named",
			nilType:         *new(*Named), // creates a new nil Named pointer
			wantErrContains: "internal error -- unsupported function call on a nil type",
		},
		{
			name:            "Nil List",
			nilType:         *new(*List), // creates a new nil List pointer
			wantErrContains: "internal error -- unsupported function call on a nil type",
		},
		{
			name:            "List<nil>",
			nilType:         &List{}, // creates a new nil List pointer
			wantErrContains: "internal error - nil ElementType for List",
		},
		{
			name:            "Nil Interval",
			nilType:         *new(*Interval), // creates a new nil Interval pointer
			wantErrContains: "internal error -- unsupported function call on a nil type",
		},
		{
			name:            "Interval<nil>",
			nilType:         &Interval{}, // creates a new nil Interval pointer
			wantErrContains: "internal error -- nil PointType for Interval",
		},
		{
			name:            "Nil Choice",
			nilType:         *new(*Choice), // creates a new nil Choice pointer
			wantErrContains: "internal error -- unsupported function call on a nil type",
		},
		{
			name:            "Nil Tuple",
			nilType:         *new(*Tuple), // creates a new nil Tuple pointer
			wantErrContains: "internal error -- unsupported function call on a nil type",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.nilType.ModelInfoName()
			if !strings.Contains(err.Error(), tc.wantErrContains) {
				t.Errorf("Name() unexpected err. got: %v, want error containing %v", err, tc.wantErrContains)
			}
		})
	}
}

func TestToSystem(t *testing.T) {
	cases := []struct {
		input string
		want  IType
	}{
		{input: "Any", want: Any},
		{input: "System.Any", want: Any},
		{input: "Boolean", want: Boolean},
		{input: "System.Boolean", want: Boolean},
		{input: "Integer", want: Integer},
		{input: "System.Integer", want: Integer},
		{input: "Long", want: Long},
		{input: "System.Long", want: Long},
		{input: "Decimal", want: Decimal},
		{input: "System.Decimal", want: Decimal},
		{input: "String", want: String},
		{input: "System.String", want: String},
		{input: "DateTime", want: DateTime},
		{input: "System.DateTime", want: DateTime},
		{input: "Date", want: Date},
		{input: "System.Date", want: Date},
		{input: "Time", want: Time},
		{input: "System.Time", want: Time},
		{input: "Quantity", want: Quantity},
		{input: "System.Quantity", want: Quantity},
		{input: "Ratio", want: Ratio},
		{input: "System.Ratio", want: Ratio},
		{input: "ValueSet", want: ValueSet},
		{input: "System.ValueSet", want: ValueSet},
		{input: "CodeSystem", want: CodeSystem},
		{input: "System.CodeSystem", want: CodeSystem},
		{input: "Vocabulary", want: Vocabulary},
		{input: "System.Vocabulary", want: Vocabulary},
		{input: "Code", want: Code},
		{input: "System.Code", want: Code},
		{input: "Concept", want: Concept},
		{input: "System.Concept", want: Concept},
		{input: "Apple", want: Unset},
		{input: "System.Apple", want: Unset},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := ToSystem(tc.input)
			if !cmp.Equal(got, tc.want) {
				t.Errorf("ToSystem(%v) got: %v, want: %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestToStrings(t *testing.T) {
	tests := []struct {
		name     string
		operands []IType
		want     string
	}{
		{
			name: "Multiple",
			operands: []IType{
				String,
				&Interval{PointType: DateTime},
				&List{ElementType: Integer},
				&Named{TypeName: "Patient"},
			},
			want: "System.String, Interval<System.DateTime>, List<System.Integer>, Named<Patient>",
		},
		{
			name:     "Single",
			operands: []IType{Date},
			want:     "System.Date",
		},
		{
			name:     "Empty",
			operands: []IType{},
			want:     "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ToStrings(tc.operands)
			if got != tc.want {
				t.Errorf("ToStrings() = %v want: %v", got, tc.want)
			}
		})
	}
}

func TestProtoAndBack(t *testing.T) {
	tests := []struct {
		name string
		typ  IType
		want *ctpb.CQLType
	}{
		{
			name: "Any",
			typ:  Any,
			want: &ctpb.CQLType{Type: &ctpb.CQLType_SystemType{SystemType: &ctpb.SystemType{Type: ctpb.SystemType_TYPE_ANY.Enum()}}},
		},
		{
			name: "Boolean",
			typ:  Boolean,
			want: &ctpb.CQLType{Type: &ctpb.CQLType_SystemType{SystemType: &ctpb.SystemType{Type: ctpb.SystemType_TYPE_BOOLEAN.Enum()}}},
		},
		{
			name: "String",
			typ:  String,
			want: &ctpb.CQLType{Type: &ctpb.CQLType_SystemType{SystemType: &ctpb.SystemType{Type: ctpb.SystemType_TYPE_STRING.Enum()}}},
		},
		{
			name: "Integer",
			typ:  Integer,
			want: &ctpb.CQLType{Type: &ctpb.CQLType_SystemType{SystemType: &ctpb.SystemType{Type: ctpb.SystemType_TYPE_INTEGER.Enum()}}},
		},
		{
			name: "Long",
			typ:  Long,
			want: &ctpb.CQLType{Type: &ctpb.CQLType_SystemType{SystemType: &ctpb.SystemType{Type: ctpb.SystemType_TYPE_LONG.Enum()}}},
		},
		{
			name: "Decimal",
			typ:  Decimal,
			want: &ctpb.CQLType{Type: &ctpb.CQLType_SystemType{SystemType: &ctpb.SystemType{Type: ctpb.SystemType_TYPE_DECIMAL.Enum()}}},
		},
		{
			name: "Quantity",
			typ:  Quantity,
			want: &ctpb.CQLType{Type: &ctpb.CQLType_SystemType{SystemType: &ctpb.SystemType{Type: ctpb.SystemType_TYPE_QUANTITY.Enum()}}},
		},
		{
			name: "Ratio",
			typ:  Ratio,
			want: &ctpb.CQLType{Type: &ctpb.CQLType_SystemType{SystemType: &ctpb.SystemType{Type: ctpb.SystemType_TYPE_RATIO.Enum()}}},
		},
		{
			name: "Date",
			typ:  Date,
			want: &ctpb.CQLType{Type: &ctpb.CQLType_SystemType{SystemType: &ctpb.SystemType{Type: ctpb.SystemType_TYPE_DATE.Enum()}}},
		},
		{
			name: "DateTime",
			typ:  DateTime,
			want: &ctpb.CQLType{Type: &ctpb.CQLType_SystemType{SystemType: &ctpb.SystemType{Type: ctpb.SystemType_TYPE_DATE_TIME.Enum()}}},
		},
		{
			name: "Time",
			typ:  Time,
			want: &ctpb.CQLType{Type: &ctpb.CQLType_SystemType{SystemType: &ctpb.SystemType{Type: ctpb.SystemType_TYPE_TIME.Enum()}}},
		},
		{
			name: "CodeSystem",
			typ:  CodeSystem,
			want: &ctpb.CQLType{Type: &ctpb.CQLType_SystemType{SystemType: &ctpb.SystemType{Type: ctpb.SystemType_TYPE_CODE_SYSTEM.Enum()}}},
		},
		{
			name: "ValueSet",
			typ:  ValueSet,
			want: &ctpb.CQLType{Type: &ctpb.CQLType_SystemType{SystemType: &ctpb.SystemType{Type: ctpb.SystemType_TYPE_VALUE_SET.Enum()}}},
		},
		{
			name: "Concept",
			typ:  Concept,
			want: &ctpb.CQLType{Type: &ctpb.CQLType_SystemType{SystemType: &ctpb.SystemType{Type: ctpb.SystemType_TYPE_CONCEPT.Enum()}}},
		},
		{
			name: "Code",
			typ:  Code,
			want: &ctpb.CQLType{Type: &ctpb.CQLType_SystemType{SystemType: &ctpb.SystemType{Type: ctpb.SystemType_TYPE_CODE.Enum()}}},
		},
		{
			name: "Vocabulary",
			typ:  Vocabulary,
			want: &ctpb.CQLType{Type: &ctpb.CQLType_SystemType{SystemType: &ctpb.SystemType{Type: ctpb.SystemType_TYPE_VOCABULARY.Enum()}}},
		},
		{
			name: "Interval",
			typ:  &Interval{PointType: Integer},
			want: &ctpb.CQLType{Type: &ctpb.CQLType_IntervalType{IntervalType: &ctpb.IntervalType{PointType: &ctpb.CQLType{Type: &ctpb.CQLType_SystemType{SystemType: &ctpb.SystemType{Type: ctpb.SystemType_TYPE_INTEGER.Enum()}}}}}},
		},
		{
			name: "List",
			typ:  &List{ElementType: Integer},
			want: &ctpb.CQLType{Type: &ctpb.CQLType_ListType{ListType: &ctpb.ListType{ElementType: &ctpb.CQLType{Type: &ctpb.CQLType_SystemType{SystemType: &ctpb.SystemType{Type: ctpb.SystemType_TYPE_INTEGER.Enum()}}}}}},
		},
		{
			name: "Choice",
			typ:  &Choice{ChoiceTypes: []IType{Integer, String}},
			want: &ctpb.CQLType{Type: &ctpb.CQLType_ChoiceType{ChoiceType: &ctpb.ChoiceType{ChoiceTypes: []*ctpb.CQLType{
				&ctpb.CQLType{Type: &ctpb.CQLType_SystemType{SystemType: &ctpb.SystemType{Type: ctpb.SystemType_TYPE_INTEGER.Enum()}}},
				&ctpb.CQLType{Type: &ctpb.CQLType_SystemType{SystemType: &ctpb.SystemType{Type: ctpb.SystemType_TYPE_STRING.Enum()}}},
			}}}},
		},
		{
			name: "Tuple",
			typ:  &Tuple{ElementTypes: map[string]IType{"apple": Integer, "banana": String}},
			want: &ctpb.CQLType{Type: &ctpb.CQLType_TupleType{TupleType: &ctpb.TupleType{ElementTypes: map[string]*ctpb.CQLType{
				"apple":  &ctpb.CQLType{Type: &ctpb.CQLType_SystemType{SystemType: &ctpb.SystemType{Type: ctpb.SystemType_TYPE_INTEGER.Enum()}}},
				"banana": &ctpb.CQLType{Type: &ctpb.CQLType_SystemType{SystemType: &ctpb.SystemType{Type: ctpb.SystemType_TYPE_STRING.Enum()}}},
			}}}},
		},
		{
			name: "Named",
			typ:  &Named{TypeName: "FHIR.TypeA"},
			want: &ctpb.CQLType{Type: &ctpb.CQLType_NamedType{NamedType: &ctpb.NamedType{TypeName: proto.String("FHIR.TypeA")}}},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := CQLTypeToProto(tc.typ)
			if err != nil {
				t.Errorf("CQLTypeToProto() returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.want, got, protocmp.Transform(), protocmp.SortRepeatedFields(&ctpb.ChoiceType{}, "choice_types")); diff != "" {
				t.Errorf("CQLTypeToProto() returned unexpected diff (-want +got):\n%s", diff)
			}

			gotValue, err := CQLTypeFromProto(tc.want)
			if err != nil {
				t.Errorf("CQLTypeFromProto() returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.typ, gotValue); diff != "" {
				t.Errorf("CQLTypeFromProto() returned unexpected diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMarshalAndUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		declared IType
		want     string
	}{
		{
			name:     "System Type",
			declared: Integer,
			want:     `"System.Integer"`,
		},
		{
			name:     "Named",
			declared: &Named{TypeName: "TypeA"},
			want:     `"TypeA"`,
		},
		{
			name:     "Interval",
			declared: &Interval{PointType: Integer},
			want:     `"Interval<System.Integer>"`,
		},
		{
			name:     "List<Integer>",
			declared: &List{ElementType: Integer},
			want:     `"List<System.Integer>"`,
		},
		{
			name: "List<List<System.Integer>>",
			declared: &List{
				ElementType: &List{ElementType: Integer},
			},
			want: `"List<List<System.Integer>>"`,
		},
		{
			name: "List<Interval<System.Integer>>",
			declared: &List{
				ElementType: &Interval{PointType: Integer},
			},
			want: `"List<Interval<System.Integer>>"`,
		},
		{
			name:     "Interval<nil>",
			declared: &Interval{PointType: nil},
			want:     `"Interval<System.Any>"`,
		},
		{
			name:     "List<nil>",
			declared: &List{ElementType: nil},
			want:     `"List<System.Any>"`,
		},
		{
			name:     "Choice<System.Integer, System.String>",
			declared: &Choice{ChoiceTypes: []IType{Integer, String}},
			want:     `"Choice<System.Integer, System.String>"`,
		},
		{
			name:     "Choice<>",
			declared: &Choice{ChoiceTypes: []IType{}},
			want:     `"Choice<>"`,
		},
		{
			name:     "Tuple<System.Integer, System.String>",
			declared: &Tuple{ElementTypes: map[string]IType{"apple": Integer, "banana": String}},
			want:     `"Tuple { apple System.Integer, banana System.String }"`,
		},
		{
			name:     "Tuple<>",
			declared: &Tuple{ElementTypes: map[string]IType{}},
			want:     `"Tuple { }"`,
		},
		{
			name:     "Tuple<nil>",
			declared: &Tuple{ElementTypes: nil},
			want:     `"Tuple"`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotBytes, err := tc.declared.MarshalJSON()
			if err != nil {
				t.Errorf("json.Marshal(%v) returned unexpected error, %v", string(gotBytes), err)
			}
			if diff := cmp.Diff(tc.want, string(gotBytes)); diff != "" {
				t.Errorf("json.Marshal(%v) returned unexpected diff (-want +got):\n%s", tc.declared, diff)
			}
		})
	}
}
