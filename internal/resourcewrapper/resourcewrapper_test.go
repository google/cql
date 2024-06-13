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

package resourcewrapper

import (
	"testing"

	r4pb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/resources/bundle_and_contained_resource_go_proto"
	r4patientpb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/resources/patient_go_proto"
)

func TestGetType(t *testing.T) {
	tests := []struct {
		name      string
		resource  *ResourceWrapper
		wantType  string
		wantError bool
	}{
		{
			name:      "R4 Patient",
			resource:  New(&r4pb.ContainedResource{OneofResource: &r4pb.ContainedResource_Patient{Patient: &r4patientpb.Patient{}}}),
			wantType:  "Patient",
			wantError: false,
		},
		{
			name:      "empty resource",
			resource:  New(&r4pb.ContainedResource{}),
			wantType:  "",
			wantError: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotType, gotError := tc.resource.ResourceType()
			if gotType != tc.wantType {
				t.Errorf("GetType() returned unexpected type = %q, want %q", gotType, tc.wantType)
			}
			if tc.wantError {
				if gotError == nil {
					t.Errorf("GetType() expected an error but got %v", gotError)
				}
			} else {
				if gotError != nil {
					t.Errorf("GetType() returned unexpected error = %v", gotError)
				}
			}
		})
	}
}
