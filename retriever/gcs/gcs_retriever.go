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

// Package gcsretriever is an implementation of the Retriever Interface for the
// CQL Engine that pulls bundles from gcs.
package gcsretriever

import (
	"context"
	"fmt"
	"io"

	"github.com/google/cql/retriever/local"
	r4pb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/resources/bundle_and_contained_resource_go_proto"
	"github.com/google/bulk_fhir_tools/gcs"
)

// Retriever implements the Retriever Interface.
type Retriever struct {
	resources   *local.Retriever
	gcsFile     string
	endpointURL string
}

// New creates a new Retriever that retrieves bundles from GCS.
func New(gcsFile, endpointURL string) (*Retriever, error) {
	r := &Retriever{gcsFile: gcsFile, endpointURL: endpointURL}
	resources, err := loadBundle(gcsFile, endpointURL)
	if err != nil {
		return nil, err
	}
	r.resources = resources
	return r, nil
}

// Retrieve returns all FHIR resources of type fhirResourceType for the patient.
func (r *Retriever) Retrieve(ctx context.Context, fhirResourceType string) ([]*r4pb.ContainedResource, error) {
	return r.resources.Retrieve(ctx, fhirResourceType)
}

func loadBundle(gcsFile string, endpointURL string) (*local.Retriever, error) {
	bucket, object, err := gcs.PathComponents(gcsFile)
	if err != nil {
		return nil, fmt.Errorf("could not parse GCS path %q: %w", gcsFile, err)
	}
	ctx := context.Background()
	client, err := gcs.NewClient(ctx, bucket, endpointURL)
	if err != nil {
		return nil, fmt.Errorf("could not connect to a client %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	rc, err := client.GetFileReader(ctx, object)
	if err != nil {
		return nil, fmt.Errorf("could not access reader for %s/%s: %v", bucket, object, err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll() failed: %w", err)
	}

	return local.NewRetrieverFromR4Bundle(data)
}
