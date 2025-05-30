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

package local

import (
	"context"
	"testing"

	"github.com/google/fhir/go/fhirversion"
	"github.com/google/fhir/go/jsonformat"
	r4datapb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/datatypes_go_proto"
	r4pb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/resources/bundle_and_contained_resource_go_proto"
	r4patientpb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/resources/patient_go_proto"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestRetrieverFromR4Bundle(t *testing.T) {
	tests := []struct {
		name          string
		bundle        string
		wantResources []*r4pb.ContainedResource
	}{
		{
			name: "Single Patient",
			bundle: `{
				"resourceType": "Bundle",
				"type": "transaction",
				"entry": [
					{
						"fullUrl": "fullUrl",
						"resource": {
							"resourceType": "Patient",
							"id": "1"}
					},
					{
						"fullUrl": "fullUrl",
						"resource": {
							"resourceType": "Encounter",
							"id": "1"}
					},
					{
						"fullUrl": "fullUrl",
						"resource": {
							"resourceType": "Observation",
							"id": "1"}
					}
				 ]
			}`,
			wantResources: []*r4pb.ContainedResource{
				{
					OneofResource: &r4pb.ContainedResource_Patient{
						Patient: &r4patientpb.Patient{Id: &r4datapb.Id{Value: "1"}},
					},
				},
			},
		},
		{
			name: "No Patients Returns Empty Slice",
			bundle: `{
				"resourceType": "Bundle",
				"type": "transaction",
				"entry": [
					{
						"fullUrl": "fullUrl",
						"resource": {
							"resourceType": "Observation",
							"id": "1"}
					}
				 ]
			}`,
			wantResources: []*r4pb.ContainedResource{},
		},
		{
			name: "Multiple Patients",
			bundle: `{
				"resourceType": "Bundle",
				"type": "transaction",
				"entry": [
					{
						"fullUrl": "fullUrl",
						"resource": {
							"resourceType": "Patient",
							"id": "1"}
					},
					{
						"fullUrl": "fullUrl",
						"resource": {
							"resourceType": "Patient",
							"id": "2"}
					},
					{
						"fullUrl": "fullUrl",
						"resource": {
							"resourceType": "Observation",
							"id": "1"}
					}
				 ]
			}`,
			wantResources: []*r4pb.ContainedResource{
				{
					OneofResource: &r4pb.ContainedResource_Patient{
						Patient: &r4patientpb.Patient{Id: &r4datapb.Id{Value: "1"}},
					},
				},
				{
					OneofResource: &r4pb.ContainedResource_Patient{
						Patient: &r4patientpb.Patient{Id: &r4datapb.Id{Value: "2"}},
					},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Test retrieve directly from bundles.
			r, err := NewRetrieverFromR4Bundle([]byte(tc.bundle))
			if err != nil {
				t.Fatalf("NewRetrieverFromR4Bundle() failed: %v", err)
			}
			gotResources, err := r.Retrieve(context.Background(), "Patient")
			if err != nil {
				t.Fatalf("Retrieve(ctx, \"Patient\") got err: %v", err)
			}
			if diff := cmp.Diff(gotResources, tc.wantResources, protocmp.Transform()); diff != "" {
				t.Errorf("Retrieve(ctx, \"Patient\") => %v, want %v, (-got +want): %v", gotResources, tc.wantResources, diff)
			}

			// Test retrieve from protocol buffers.
			unmarshaller, err := jsonformat.NewUnmarshallerWithoutValidation("UTC", fhirversion.R4)
			if err != nil {
				t.Fatalf("jsonformat.NewUnmarshallerWithoutValidation() failed: %v", err)
			}
			cr, err := unmarshaller.UnmarshalR4([]byte(tc.bundle))
			if err != nil {
				t.Fatalf("UnmarshalR4() failed: %v", err)
			}
			r, err = NewRetrieverFromR4BundleProto(cr.GetBundle())
			if err != nil {
				t.Fatalf("NewRetrieverFromR4BundleProto() failed: %v", err)
			}
			gotResources, err = r.Retrieve(context.Background(), "Patient")
			if err != nil {
				t.Fatalf("Retrieve(ctx, \"Patient\") got err: %v", err)
			}
			if diff := cmp.Diff(gotResources, tc.wantResources, protocmp.Transform()); diff != "" {
				t.Errorf("Retrieve(ctx, \"Patient\") => %v, want %v, (-got +want): %v", gotResources, tc.wantResources, diff)
			}
		})
	}
}
