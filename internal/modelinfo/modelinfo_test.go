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

package modelinfo

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/cql/internal/embeddata"
	"github.com/google/cql/types"
	"github.com/google/go-cmp/cmp"
)

func TestPropertyTypeSpecifier_FHIR(t *testing.T) {
	cases := []struct {
		name       string
		parentType types.IType
		property   string
		want       types.IType
	}{
		{
			name:       "FHIR.Account status",
			parentType: &types.Named{TypeName: "FHIR.Account"},
			property:   "status",
			want:       &types.Named{TypeName: "FHIR.AccountStatus"},
		},
		{
			name:       "FHIR.Account id (comes from base type FHIR.Resource)",
			parentType: &types.Named{TypeName: "FHIR.Account"},
			property:   "id",
			want:       &types.Named{TypeName: "FHIR.id"},
		},
		{
			name:       "List of FHIR.Account status",
			parentType: &types.List{ElementType: &types.Named{TypeName: "FHIR.Account"}},
			property:   "status",
			want:       &types.List{ElementType: &types.Named{TypeName: "FHIR.AccountStatus"}},
		},
		{
			name:       "ListTypeSpecifier FHIR.Account subject",
			parentType: &types.Named{TypeName: "FHIR.Account"},
			property:   "subject",
			want:       &types.List{ElementType: &types.Named{TypeName: "FHIR.Reference"}},
		},
		{
			name:       "Tuple",
			parentType: &types.Tuple{ElementTypes: map[string]types.IType{"foo": types.Integer}},
			property:   "foo",
			want:       types.Integer,
		},
		{
			name:       "System type, FHIR.AccountStatus value",
			parentType: &types.Named{TypeName: "FHIR.AccountStatus"},
			property:   "value",
			want:       types.String,
		},
		{
			name:       "Bundle.Entry.link: ListTypeSpecifier with NamedTypeSpecifier",
			parentType: &types.Named{TypeName: "FHIR.Bundle.Entry"},
			property:   "link",
			want:       &types.List{ElementType: &types.Named{TypeName: "FHIR.Bundle.Link"}},
		},
		{
			name:       "Property on types.Any returns types.Any",
			parentType: types.Any,
			property:   "value",
			want:       types.Any,
		},
		{
			name:       "Interval.low",
			parentType: &types.Interval{PointType: types.Integer},
			property:   "low",
			want:       types.Integer,
		},
		{
			name:       "Interval.high",
			parentType: &types.Interval{PointType: types.Integer},
			property:   "high",
			want:       types.Integer,
		},
		{
			name:       "Interval.lowClosed",
			parentType: &types.Interval{PointType: types.Integer},
			property:   "lowClosed",
			want:       types.Boolean,
		},
		{
			name:       "Interval.highClosed",
			parentType: &types.Interval{PointType: types.Integer},
			property:   "highClosed",
			want:       types.Boolean,
		},
		{
			name:       "Choice Type property result (FHIR.ActivityDefinition.subject)",
			parentType: &types.Named{TypeName: "FHIR.ActivityDefinition"},
			property:   "subject",
			want: &types.Choice{ChoiceTypes: []types.IType{
				&types.Named{TypeName: "FHIR.CodeableConcept"},
				&types.Named{TypeName: "FHIR.Reference"},
			},
			},
		},
		{
			name: "Choice<FHIR.Patient, FHIR.Observation>.id property results in FHIR.id",
			parentType: &types.Choice{ChoiceTypes: []types.IType{
				&types.Named{TypeName: "FHIR.Patient"},
				&types.Named{TypeName: "FHIR.Observation"},
			}},
			property: "id",
			want:     &types.Named{TypeName: "FHIR.id"},
		},
		{
			name: "Choice<FHIR.boolean, FHIR.string>.value property results in Choice<System.Boolean, System.String>",
			parentType: &types.Choice{ChoiceTypes: []types.IType{
				&types.Named{TypeName: "FHIR.boolean"},
				&types.Named{TypeName: "FHIR.string"},
			}},
			property: "value",
			want:     &types.Choice{ChoiceTypes: []types.IType{types.Boolean, types.String}},
		},
		{
			name: "Nested Choice type: Choice<Choice<FHIR.integer FHIR.boolean>, FHIR.string>.value property",
			parentType: &types.Choice{ChoiceTypes: []types.IType{
				&types.Choice{ChoiceTypes: []types.IType{&types.Named{TypeName: "FHIR.integer"}, &types.Named{TypeName: "FHIR.boolean"}}},
				&types.Named{TypeName: "FHIR.string"},
			}},
			property: "value",
			want: &types.Choice{ChoiceTypes: []types.IType{
				&types.Choice{ChoiceTypes: []types.IType{types.Integer, types.Boolean}},
				types.String,
			}},
		},
		// System model info properties:
		{
			name:       "System.ValueSet.codesystems",
			parentType: types.ValueSet,
			property:   "codesystems",
			want:       &types.List{ElementType: types.CodeSystem},
		},
		{
			name:       "System.Code.code",
			parentType: types.Code,
			property:   "code",
			want:       types.String,
		},
		{
			name:       "System.Quantity.value",
			parentType: types.Quantity,
			property:   "value",
			want:       types.Decimal,
		},
		{
			name:       "Subtype inherits properties of parent",
			parentType: types.ValueSet,
			property:   "version",
			want:       types.String,
		},
	}
	modelinfo := newFHIRModelInfo(t)
	modelinfo.SetUsing(Key{Name: "FHIR", Version: "4.0.1"})

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := modelinfo.PropertyTypeSpecifier(tc.parentType, tc.property)
			if err != nil {
				t.Fatalf("PropertyTypeSpecifier(%s, %s) failed unexpectedly: %v", tc.parentType, tc.property, err)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("PropertyTypeSpecifier(%s, %s) diff (-want +got):\n%s", tc.parentType, tc.property, diff)
			}
		})
	}
}

func TestPropertyTypeSpecifier_FHIRErrors(t *testing.T) {
	cases := []struct {
		name            string
		parentType      types.IType
		property        string
		wantErrContains string
	}{
		{
			name:            "Missing parent type",
			parentType:      &types.Named{TypeName: "FHIR.MissingFakeType"},
			property:        "status",
			wantErrContains: "parentNamedType \"FHIR.MissingFakeType\" not found in data model",
		},
		{
			name:            "Missing property type",
			parentType:      &types.Named{TypeName: "FHIR.Account"},
			property:        "fake",
			wantErrContains: "property \"fake\" not found in Parent Type \"FHIR.Account\"",
		},
		{
			name:            "Deep nested property",
			parentType:      &types.Named{TypeName: "FHIR.Account"},
			property:        "property1.property2",
			wantErrContains: "internal error - property passed to PropertyTypeSpecifier should not contain \".\" only a single component of the property should be passed at a time",
		},
		{
			name:            "unsupported property on interval",
			parentType:      &types.Interval{PointType: types.Integer},
			property:        "fake",
			wantErrContains: "invalid property on interval. got: fake, want: low, high, lowClosed, highClosed",
		},
		{
			name:            "unsupported property on System.Quantity",
			parentType:      types.Quantity,
			property:        "fake",
			wantErrContains: "property \"fake\" not found in Parent Type \"System.Quantity\"",
		},
		{
			name:            "unsupported property on Choice<System.Quantity>",
			parentType:      &types.Choice{ChoiceTypes: []types.IType{types.Quantity}},
			property:        "fake",
			wantErrContains: "property \"fake\" not found in Parent Type \"Choice<System.Quantity>\"",
		},
		{
			name:            "Tuple missing property",
			parentType:      &types.Tuple{ElementTypes: map[string]types.IType{"foo": types.Integer}},
			property:        "bar",
			wantErrContains: `Tuple<foo: System.Integer> does not have property "bar"`,
		},
	}
	modelinfo := newFHIRModelInfo(t)
	modelinfo.SetUsing(Key{Name: "FHIR", Version: "4.0.1"})

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := modelinfo.PropertyTypeSpecifier(tc.parentType, tc.property)
			if !strings.Contains(err.Error(), tc.wantErrContains) {
				t.Errorf("PropertyTypeSpecifier(%s, %s) returned unexpected error: got: %v, want contains: %s", tc.parentType, tc.property, err, tc.wantErrContains)
			}
		})
	}
}

func TestIsImplicitlyConvertible(t *testing.T) {
	cases := []struct {
		name string
		from types.IType
		to   types.IType
		want Convertible
	}{
		{
			name: "System.Integer -> System.Decimal",
			from: types.Integer,
			to:   types.Decimal,
			want: Convertible{
				IsConvertible: true,
				Library:       "SYSTEM",
				Function:      "ToDecimal",
			},
		},
		{
			name: "System.Date -> System.DateTime",
			from: types.Date,
			to:   types.DateTime,
			want: Convertible{
				IsConvertible: true,
				Library:       "SYSTEM",
				Function:      "ToDateTime",
			},
		},
		{
			name: "FHIR.Coding -> System.Code",
			from: &types.Named{TypeName: "FHIR.Coding"},
			to:   types.Code,
			want: Convertible{
				IsConvertible: true,
				Library:       "FHIRHelpers",
				Function:      "ToCode",
			},
		},
		{
			name: "FHIR.Period -> Interval<System.DateTime>",
			from: &types.Named{TypeName: "FHIR.Period"},
			to:   &types.Interval{PointType: types.DateTime},
			want: Convertible{
				IsConvertible: true,
				Library:       "FHIRHelpers",
				Function:      "ToInterval",
			},
		},
	}

	modelinfo := newFHIRModelInfo(t)
	modelinfo.SetUsing(Key{Name: "FHIR", Version: "4.0.1"})
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := modelinfo.IsImplicitlyConvertible(tc.from, tc.to)
			if err != nil {
				t.Fatalf("IsImplicitlyConvertible(%s, %s) failed unexpectedly: %v", tc.from, tc.to, err)
			}
			if got != tc.want {
				t.Errorf("IsImplicitlyConvertible(%s, %s) got: %v, want: %v", tc.from, tc.to, got, tc.want)
			}
		})
	}
}

func TestBaseTypes(t *testing.T) {
	cases := []struct {
		name  string
		child types.IType
		want  []types.IType
	}{
		{
			name:  "FHIR Modelinfo",
			child: &types.Named{TypeName: "FHIR.Patient"},
			want: []types.IType{
				&types.Named{TypeName: "FHIR.DomainResource"},
				&types.Named{TypeName: "FHIR.Resource"},
			},
		},
		{
			name:  "Another FHIR Modelinfo",
			child: &types.Named{TypeName: "FHIR.id"},
			want: []types.IType{
				&types.Named{TypeName: "FHIR.string"},
				&types.Named{TypeName: "FHIR.Element"},
			},
		},
		{
			name:  "Interval",
			child: &types.Interval{PointType: types.ValueSet},
			want:  []types.IType{&types.Interval{PointType: types.Vocabulary}},
		},
		{
			name:  "List",
			child: &types.List{ElementType: types.ValueSet},
			want:  []types.IType{&types.List{ElementType: types.Vocabulary}},
		},
		{
			name:  "Any",
			child: types.Any,
			want:  []types.IType{},
		},
		{
			name:  "Choice",
			child: &types.Choice{ChoiceTypes: []types.IType{types.Integer, types.String}},
			want:  []types.IType{},
		},
		{
			name:  "Tuple",
			child: &types.Tuple{ElementTypes: map[string]types.IType{"foo": types.ValueSet}},
			want:  []types.IType{},
		},
		{
			name:  "List<Tuple>",
			child: &types.List{ElementType: &types.Tuple{ElementTypes: map[string]types.IType{"foo": types.ValueSet}}},
			want:  []types.IType{},
		},
	}

	modelinfo := newFHIRModelInfo(t)
	modelinfo.SetUsing(Key{Name: "FHIR", Version: "4.0.1"})
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := modelinfo.BaseTypes(tc.child)
			if err != nil {
				t.Fatalf("BaseTypes(%v) failed unexpectedly: %v", tc.child, err)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("BaseTypes(%v) diff (-want +got):\n%s", tc.child, diff)
			}
		})
	}
}

func TestBaseTypes_Errors(t *testing.T) {
	cases := []struct {
		name            string
		childType       types.IType
		wantErrContains string
	}{
		{
			name:            "nil child type",
			childType:       nil,
			wantErrContains: "cannot be nil",
		},
		{
			name:            "Unknown child type",
			childType:       &types.Named{TypeName: "unknown"},
			wantErrContains: "Named<unknown> not found in the data model",
		},
	}

	modelinfo := newFHIRModelInfo(t)
	modelinfo.SetUsing(Key{Name: "FHIR", Version: "4.0.1"})
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := modelinfo.BaseTypes(tc.childType)
			if err == nil {
				t.Fatalf("BaseTypes(%v) succeeded, expected (%v)", tc.childType, tc.wantErrContains)
			}
			if !strings.Contains(err.Error(), tc.wantErrContains) {
				t.Fatalf("BaseTypes(%v) unexpected error. got: %v, want error contains: %v", tc.childType, err, tc.wantErrContains)
			}
		})
	}
}

func TestIsSubType(t *testing.T) {
	cases := []struct {
		name      string
		childType types.IType
		baseType  types.IType
		want      bool
	}{
		{
			name:      "FHIR.DomainResource is child of FHIR.Resource (FHIR.Resource -> FHIR.DomainResource)",
			childType: &types.Named{TypeName: "FHIR.DomainResource"},
			baseType:  &types.Named{TypeName: "FHIR.Resource"},
			want:      true,
		},
		{
			name:      "FHIR.Resource is not a child of FHIR.DomainResource (FHIR.Resource -> FHIR.DomainResource)",
			childType: &types.Named{TypeName: "FHIR.Resource"},
			baseType:  &types.Named{TypeName: "FHIR.DomainResource"},
			want:      false,
		},
		{
			name:      "FHIR.Patient is child of FHIR.Resource (FHIR.Resource -> FHIR.DomainResource -> FHIR.Patient)",
			childType: &types.Named{TypeName: "FHIR.Patient"},
			baseType:  &types.Named{TypeName: "FHIR.Resource"},
			want:      true,
		},
		{
			name:      "FHIR.id is child of FHIR.string (FHIR.id -> FHIR.string)",
			childType: &types.Named{TypeName: "FHIR.id"},
			baseType:  &types.Named{TypeName: "FHIR.string"},
			want:      true,
		},
		{
			name:      "Interval<FHIR.Patient> is child of Interval<FHIR.Resource> (FHIR.Resource -> FHIR.DomainResource -> FHIR.Patient)",
			childType: &types.Interval{PointType: &types.Named{TypeName: "FHIR.Patient"}},
			baseType:  &types.Interval{PointType: &types.Named{TypeName: "FHIR.Resource"}},
			want:      true,
		},
		{
			name:      "List<FHIR.Patient> is child of List<FHIR.Resource> (FHIR.Resource -> FHIR.DomainResource -> FHIR.Patient)",
			childType: &types.List{ElementType: &types.Named{TypeName: "FHIR.Patient"}},
			baseType:  &types.List{ElementType: &types.Named{TypeName: "FHIR.Resource"}},
			want:      true,
		},
		{
			name:      "FHIR.Patient is child of System.Any (System.Any -> FHIR.Resource -> FHIR.DomainResource -> FHIR.Patient)",
			childType: &types.Named{TypeName: "FHIR.Patient"},
			baseType:  types.Any,
			want:      true,
		},
		{
			name:      "System.Any is parent of ChoiceType",
			childType: &types.Choice{ChoiceTypes: []types.IType{types.Integer}},
			baseType:  types.Any,
			want:      true,
		},
		{
			name:      "System.Any is parent of System.Valueset",
			childType: types.ValueSet,
			baseType:  types.Any,
			want:      true,
		},
		{
			name:      "System.Any is parent of System.Integer",
			childType: types.Integer,
			baseType:  types.Any,
			want:      true,
		},
		{
			name:      "System.Any is parent of Interval",
			childType: &types.Interval{PointType: types.Integer},
			baseType:  types.Any,
			want:      true,
		},
		{
			name:      "System.Any is parent of Tuple",
			childType: &types.Tuple{ElementTypes: map[string]types.IType{"foo": types.ValueSet}},
			baseType:  types.Any,
			want:      true,
		},
		{
			name:      "System.Any is not a parent",
			childType: types.Any,
			baseType:  types.String,
			want:      false,
		},
		{
			name:      "System.Vocabulary is a parent of System.Valueset",
			childType: types.ValueSet,
			baseType:  types.Vocabulary,
			want:      true,
		},
		{
			name:      "System.String is not a parent of ChoiceType",
			childType: &types.Choice{ChoiceTypes: []types.IType{types.Integer}},
			baseType:  types.String,
			want:      false,
		},
		{
			name:      "Tuple is Subtype of Tuple",
			childType: &types.Tuple{ElementTypes: map[string]types.IType{"foo": types.ValueSet, "bar": types.String}},
			baseType:  &types.Tuple{ElementTypes: map[string]types.IType{"foo": types.Vocabulary, "bar": types.String}},
			want:      true,
		},
		{
			name:      "List<Tuple> is Subtype of List<Tuple>",
			childType: &types.List{ElementType: &types.Tuple{ElementTypes: map[string]types.IType{"foo": types.ValueSet, "bar": types.String}}},
			baseType:  &types.List{ElementType: &types.Tuple{ElementTypes: map[string]types.IType{"foo": types.Vocabulary, "bar": types.String}}},
			want:      true,
		},
		{
			name:      "Tuple is not Subtype unequal number of elements",
			childType: &types.Tuple{ElementTypes: map[string]types.IType{"foo": types.ValueSet}},
			baseType:  &types.Tuple{ElementTypes: map[string]types.IType{"foo": types.Vocabulary, "bar": types.String}},
			want:      false,
		},
		{
			name:      "Tuple is not Subtype non subtype elements",
			childType: &types.Tuple{ElementTypes: map[string]types.IType{"foo": types.ValueSet, "bar": types.String}},
			baseType:  &types.Tuple{ElementTypes: map[string]types.IType{"foo": types.Vocabulary, "bar": types.Integer}},
			want:      false,
		},
		{
			name:      "Tuple is not Subtype of other type",
			childType: &types.Tuple{ElementTypes: map[string]types.IType{"foo": types.ValueSet, "bar": types.String}},
			baseType:  &types.Named{TypeName: "FHIR.Patient"},
			want:      false,
		},
	}

	modelinfo := newFHIRModelInfo(t)
	modelinfo.SetUsing(Key{Name: "FHIR", Version: "4.0.1"})
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := modelinfo.IsSubType(tc.childType, tc.baseType)
			if err != nil {
				t.Fatalf("IsSubType(%v, %v) failed unexpectedly: %v", tc.childType, tc.baseType, err)
			}
			if got != tc.want {
				t.Errorf("IsSubType(%v, %v) got: %v, want: %v", tc.childType, tc.baseType, got, tc.want)
			}
		})
	}
}

func TestIsSubType_FHIRErrors(t *testing.T) {
	cases := []struct {
		name            string
		childType       types.IType
		baseType        types.IType
		wantErrContains string
	}{
		{
			name:            "nil child type",
			childType:       nil,
			baseType:        &types.Named{TypeName: "FHIR.Resource"},
			wantErrContains: "cannot be nil",
		},
		{
			name:            "nil parent type",
			childType:       &types.Named{TypeName: "FHIR.Resource"},
			baseType:        nil,
			wantErrContains: "cannot be nil",
		},
	}

	modelinfo := newFHIRModelInfo(t)
	modelinfo.SetUsing(Key{Name: "FHIR", Version: "4.0.1"})
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := modelinfo.IsSubType(tc.childType, tc.baseType)
			if err == nil {
				t.Fatalf("IsSubType(%v, %v) succeeded, expected (%v)", tc.childType, tc.baseType, tc.wantErrContains)
			}
			if !strings.Contains(err.Error(), tc.wantErrContains) {
				t.Fatalf("IsSubType(%v, %v) unexpected error. got: %v, want error contains: %v", tc.childType, tc.baseType, err, tc.wantErrContains)
			}
		})
	}
}

func TestToNamed(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  *types.Named
	}{
		{
			name:  "Unqualified",
			input: "Patient",
			want:  &types.Named{TypeName: "FHIR.Patient"},
		},
		{
			name:  "Qualified",
			input: "FHIR.Patient",
			want:  &types.Named{TypeName: "FHIR.Patient"},
		},
		{
			name:  "Unqualified Multi Level",
			input: "Account.Coverage",
			want:  &types.Named{TypeName: "FHIR.Account.Coverage"},
		},
		{
			name:  "Qualified Multi Level",
			input: "FHIR.Account.Coverage",
			want:  &types.Named{TypeName: "FHIR.Account.Coverage"},
		},
	}

	modelinfo := newFHIRModelInfo(t)
	modelinfo.SetUsing(Key{Name: "FHIR", Version: "4.0.1"})
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := modelinfo.ToNamed(tc.input)
			if err != nil {
				t.Fatalf("ToNamed() failed unexpectedly: %v", err)
			}
			if !got.Equal(tc.want) {
				t.Errorf("ToNamed() got: %v, want: %v", got, tc.want)
			}
		})
	}
}

func TestToNamed_Errors(t *testing.T) {
	cases := []struct {
		name            string
		input           string
		wantErrContains string
	}{
		{
			name:            "Unqualified Non Existent",
			input:           "Apple",
			wantErrContains: "type Apple not found",
		},
		{
			name:            "Qualified Non Existent",
			input:           "FHIR.Apple",
			wantErrContains: "type FHIR.Apple not found",
		},
		{
			name:            "System Type",
			input:           "Integer",
			wantErrContains: "type Integer not found",
		},
		{
			name:            "Empty",
			input:           "",
			wantErrContains: "received an empty type, which is invalid",
		},
		{
			name:            "Multi Level Non Existent",
			input:           "Apple.Banana",
			wantErrContains: "type Apple.Banana not found",
		},
	}

	modelinfo := newFHIRModelInfo(t)
	modelinfo.SetUsing(Key{Name: "FHIR", Version: "4.0.1"})
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := modelinfo.ToNamed(tc.input)
			if !strings.Contains(err.Error(), tc.wantErrContains) {
				t.Fatalf("ToNamed() unexpected error. got: %v, want error contains: %v", err, tc.wantErrContains)
			}
		})
	}
}

func TestToNamed_NotFoundError(t *testing.T) {
	modelInfo := newFHIRModelInfo(t)
	modelInfo.SetUsing(Key{Name: "FHIR", Version: "4.0.1"})
	_, err := modelInfo.ToNamed("Apple")
	if !errors.Is(err, ErrTypeNotFound) {
		t.Errorf("ToNamed(%s) unexpected error. got: %v, want error contains: %v", "Apple", err, ErrTypeNotFound)
	}
}

func TestPatientBirthDatePropertyName(t *testing.T) {
	modelinfo := newFHIRModelInfo(t)
	t.Run("PatientBirthDatePropertyName Error", func(t *testing.T) {
		_, err := modelinfo.PatientBirthDatePropertyName()
		if !errors.Is(err, errUsingNotSet) {
			t.Errorf("PatientBirthDatePropertyName() unexpected error. got: %v, want error contains: %v", err, errUsingNotSet)
		}
	})

	t.Run("PatientBirthDatePropertyName Success", func(t *testing.T) {
		modelinfo.SetUsing(Key{Name: "FHIR", Version: "4.0.1"})
		bDay, err := modelinfo.PatientBirthDatePropertyName()
		wantBDay := "birthDate.value"
		if err != nil {
			t.Fatalf("PatientBirthDatePropertyName() failed unexpectedly: %v", err)
		}
		if bDay != wantBDay {
			t.Errorf("PatientBirthDatePropertyName() got: %v, want: %v", bDay, wantBDay)
		}
	})
}

func TestURL(t *testing.T) {
	modelinfo := newFHIRModelInfo(t)

	t.Run("URL Error", func(t *testing.T) {
		_, err := modelinfo.URL()
		if !errors.Is(err, errUsingNotSet) {
			t.Errorf("URL() unexpected error. got: %v, want error contains: %v", err, errUsingNotSet)
		}
	})

	t.Run("URL Success", func(t *testing.T) {
		modelinfo.SetUsing(Key{Name: "FHIR", Version: "4.0.1"})
		url, err := modelinfo.URL()
		wantURL := "http://hl7.org/fhir"
		if err != nil {
			t.Fatalf("URL() failed unexpectedly: %v", err)
		}
		if url != wantURL {
			t.Errorf("URL() got: %v, want: %v", url, wantURL)
		}
	})
}

func TestDefaultContext(t *testing.T) {
	modelinfo := newFHIRModelInfo(t)

	t.Run("DefaultContext Error", func(t *testing.T) {
		_, err := modelinfo.DefaultContext()
		if !errors.Is(err, errUsingNotSet) {
			t.Errorf("DefaultContext() unexpected error. got: %v, want error contains: %v", err, errUsingNotSet)
		}
	})

	t.Run("DefaultContext Success", func(t *testing.T) {
		modelinfo.SetUsing(Key{Name: "FHIR", Version: "4.0.1"})
		ctx, err := modelinfo.DefaultContext()
		wantCtx := ""
		if err != nil {
			t.Fatalf("DefaultContext() failed unexpectedly: %v", err)
		}
		if ctx != wantCtx {
			t.Errorf("DefaultContext() got: %v, want: %v", ctx, wantCtx)
		}
	})
}

func TestNamedTypeInfo(t *testing.T) {
	modelinfo := newFHIRModelInfo(t)

	t.Run("NamedTypeInfo Error", func(t *testing.T) {
		_, err := modelinfo.NamedTypeInfo(&types.Named{TypeName: "FHIR.Account.Coverage"})
		if !errors.Is(err, errUsingNotSet) {
			t.Errorf("NamedTypeInfo() unexpected error. got: %v, want error contains: %v", err, errUsingNotSet)
		}
	})

	t.Run("NamedTypeInfo Success", func(t *testing.T) {
		modelinfo.SetUsing(Key{Name: "FHIR", Version: "4.0.1"})
		info, err := modelinfo.NamedTypeInfo(&types.Named{TypeName: "FHIR.Account.Coverage"})
		wantInfo := &TypeInfo{
			Name:       "FHIR.Account.Coverage",
			Properties: map[string]types.IType{"coverage": &types.Named{TypeName: "FHIR.Reference"}, "priority": &types.Named{TypeName: "FHIR.positiveInt"}},
			BaseType:   "FHIR.BackboneElement",
		}
		if err != nil {
			t.Fatalf("NamedTypeInfo() failed unexpectedly: %v", err)
		}
		if diff := cmp.Diff(info, wantInfo); diff != "" {
			t.Errorf("NamedTypeInfo() returned unexpected diff (-got +want):\n%s", diff)
		}
	})
}

func TestSetUsing_Error(t *testing.T) {
	t.Run("Data model does not exist", func(t *testing.T) {
		modelinfo := newFHIRModelInfo(t)
		err := modelinfo.SetUsing(Key{Name: "Apple", Version: "4.0.1"})
		wantErr := "Apple 4.0.1 data model not found"
		if !strings.Contains(err.Error(), wantErr) {
			t.Errorf("SetUsing() unexpected error. got (%v), want error contains: (%v)", err, wantErr)
		}
	})
	t.Run("Only one data model allowed", func(t *testing.T) {
		modelinfo := newFHIRModelInfo(t)
		err := modelinfo.SetUsing(Key{Name: "FHIR", Version: "4.0.1"})
		if err != nil {
			t.Fatalf("SetUsing() failed unexpectedly: %v", err)
		}
		err = modelinfo.SetUsing(Key{Name: "Apple", Version: "4.0.1"})
		wantErr := "only one data model at a time is currently supported, but got FHIR 4.0.1 and Apple 4.0.1"
		if !strings.Contains(err.Error(), wantErr) {
			t.Errorf("SetUsing() unexpected error. got (%v), want error contains: (%v)", err, wantErr)
		}
	})
}

func newFHIRModelInfo(t *testing.T) *ModelInfos {
	t.Helper()
	fhirMIBytes, err := embeddata.ModelInfos.ReadFile("third_party/cqframework/fhir-modelinfo-4.0.1.xml")
	if err != nil {
		t.Fatalf("Reading embedded file %s failed unexpectedly: %v", "third_party/cqframework/fhir-modelinfo-4.0.1.xml", err)
	}

	m, err := New([][]byte{fhirMIBytes})
	if err != nil {
		t.Fatalf("New modelinfo unexpected error: %v", err)
	}
	return m
}
