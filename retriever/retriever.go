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

// Package retriever defines the interface between the CQL engine and the data source CQL will be
// computed over. Those using the CQL engine must provide an implementation of the Retriever
// Interface.
package retriever

import (
	"context"

	r4pb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/resources/bundle_and_contained_resource_go_proto"
)

// Retriever defines the interface between the CQL engine and the data source CQL will be computed
// over.
type Retriever interface {
	// Retrieve returns all FHIR resources of type fhirResourceType for the patient.
	Retrieve(ctx context.Context, fhirResourceType string) ([]*r4pb.ContainedResource, error)
}
