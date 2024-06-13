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

// Package local is an implementation of the Retriever Interface for the CQL engine. The
// implementation can be initialized from a json FHIR bundle of all the patient's FHIR Resources.
package local

import (
	"context"

	"github.com/google/cql/internal/resourcewrapper"
	"github.com/google/fhir/go/fhirversion"
	"github.com/google/fhir/go/jsonformat"

	r4pb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/resources/bundle_and_contained_resource_go_proto"
)

// Retriever implements the Retriever Interface for the CQL engine.
type Retriever struct {
	resources map[string][]*r4pb.ContainedResource
}

// NewRetrieverFromR4BundleProto initializes a local retriever from a FHIR bundle proto of all
// the patient's FHIR resources.
func NewRetrieverFromR4BundleProto(bundle *r4pb.Bundle) (*Retriever, error) {
	r := &Retriever{resources: make(map[string][]*r4pb.ContainedResource)}
	for _, e := range bundle.GetEntry() {
		rw := resourcewrapper.New(e.GetResource())
		resourceType, err := rw.ResourceType()
		if err != nil {
			return nil, err
		}
		r.resources[resourceType] = append(r.resources[resourceType], rw.Resource)
	}

	return r, nil
}

// NewRetrieverFromR4Bundle initializes a local Retriever from a json R4 FHIR bundle of all the
// patient's FHIR Resources.
func NewRetrieverFromR4Bundle(jsonBundle []byte) (*Retriever, error) {
	unmarshaller, err := jsonformat.NewUnmarshallerWithoutValidation("UTC", fhirversion.R4)
	if err != nil {
		return nil, err
	}
	containedResource, err := unmarshaller.UnmarshalR4(jsonBundle)
	if err != nil {
		return nil, err
	}
	bundle := containedResource.GetBundle()
	return NewRetrieverFromR4BundleProto(bundle)
}

// Retrieve returns all FHIR resources of type fhirResourceType for the patient.
func (r *Retriever) Retrieve(ctx context.Context, fhirResourceType string) ([]*r4pb.ContainedResource, error) {
	if resources, ok := r.resources[fhirResourceType]; ok {
		return resources, nil
	}
	return []*r4pb.ContainedResource{}, nil
}
