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

	"github.com/google/cql/internal/embeddata"
	"github.com/google/cql/internal/modelinfo"
	"github.com/google/cql/internal/reference"
	"github.com/google/cql/model"
	"github.com/google/cql/result"
	"github.com/google/cql/types"
	r4patientpb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/resources/patient_go_proto"
)

func TestEvalPropertyValue_Errors(t *testing.T) {
	tests := []struct {
		name            string
		property        string
		value           result.Value
		resultType      types.IType
		wantErrContains string
	}{
		{
			name:     "protomessage invalid property",
			property: "apple",
			value:    newOrFatal(t, result.Named{Value: &r4patientpb.Patient{}, RuntimeType: &types.Named{TypeName: "FHIR.Patient"}}),
		},
		{
			name:     "integer",
			property: "name",
			value:    newOrFatal(t, 4),
		},
		{
			name:     "boolean",
			property: "name",
			value:    newOrFatal(t, true),
		},
		{
			name:     "string",
			property: "name",
			value:    newOrFatal(t, "hello"),
		},
		{
			name:     "long",
			property: "name",
			value:    newOrFatal(t, 10),
		},
		{
			name:     "decimal",
			property: "name",
			value:    newOrFatal(t, 10.001),
		},
		{
			name:     "valueset",
			property: "name",
			value:    newOrFatal(t, result.ValueSet{ID: "ID", Version: "Version"}),
		},
		{
			name:     "list of integers",
			property: "name",
			value: newOrFatal(t, result.List{
				Value: []result.Value{
					newOrFatal(t, 4),
					newOrFatal(t, 5),
				},
				StaticType: &types.List{ElementType: types.Integer},
			}),
			resultType: &types.List{ElementType: types.Integer},
		},
		{
			name:     "interval invalid property",
			property: "invalid",
			value: newOrFatal(t, result.Interval{
				Low:           newOrFatal(t, 4),
				High:          newOrFatal(t, 5),
				LowInclusive:  false,
				HighInclusive: true,
				StaticType:    &types.Interval{PointType: types.Integer},
			}),
			wantErrContains: "property invalid is not supported on Intervals",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			i := &interpreter{
				refs:                reference.NewResolver[result.Value, *model.FunctionDef](),
				modelInfo:           newFHIRModelInfo(t),
				evaluationTimestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			}
			_, err := i.valueProperty(tc.value, tc.property, tc.resultType)
			if err == nil {
				t.Errorf("evalPropertyValue(%q) succeeded, want error", tc.property)
			}
			if tc.wantErrContains != "" && !strings.Contains(err.Error(), tc.wantErrContains) {
				t.Errorf("evalPropertyValue(%s) error did not contain expected string. got: %v, want: %v", tc.property, err.Error(), tc.wantErrContains)
			}
		})
	}
}

func TestEvalPropertyValue(t *testing.T) {
	tests := []struct {
		name       string
		property   string
		value      result.Value
		resultType types.IType
		wantValue  result.Value
	}{
		{
			name:      "property on null input value",
			property:  "apple",
			value:     newOrFatal(t, nil),
			wantValue: newOrFatal(t, nil),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			i := &interpreter{
				refs:                reference.NewResolver[result.Value, *model.FunctionDef](),
				modelInfo:           newFHIRModelInfo(t),
				evaluationTimestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			}
			got, err := i.valueProperty(tc.value, tc.property, tc.resultType)
			if err != nil {
				t.Errorf("evalPropertyValue(%q) unexpected error: %v", tc.property, err)
			}
			if !got.Equal(tc.wantValue) {
				t.Errorf("evalPropertyValue(%q) = %v, want %v", tc.property, got, tc.wantValue)
			}
		})
	}
}

func newFHIRModelInfo(t *testing.T) *modelinfo.ModelInfos {
	t.Helper()
	fhirMIBytes, err := embeddata.ModelInfos.ReadFile("third_party/cqframework/fhir-modelinfo-4.0.1.xml")
	if err != nil {
		t.Fatalf("Reading embedded file %s failed unexpectedly: %v", "third_party/cqframework/fhir-modelinfo-4.0.1.xml", err)
	}

	m, err := modelinfo.New([][]byte{fhirMIBytes})
	if err != nil {
		t.Fatalf("New modelinfo unexpected error: %v", err)
	}
	m.SetUsing(modelinfo.Key{Name: "FHIR", Version: "4.0.1"})
	return m
}
