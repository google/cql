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

package gcsretriever

import (
	"context"
	"testing"

	"github.com/google/cql/retriever/local"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
	"github.com/google/bulk_fhir_tools/testhelpers"
)

func TestGCSRetrieve(t *testing.T) {
	var bucketID = "TestBucket"
	var fileName = "TestFile"
	var bundle = `{
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
			}`

	server := testhelpers.NewGCSServer(t)
	server.AddObject(bucketID, fileName, testhelpers.GCSObjectEntry{
		Data: []byte(bundle),
	})

	path := "gs://" + bucketID + "/" + fileName
	gcsRetriever, err := New(path, server.URL())
	if err != nil {
		t.Fatalf("failed to create GCSRetriever: %v", err)
	}

	gcsResult, err := gcsRetriever.Retrieve(context.Background(), "Patient")
	if err != nil {
		t.Fatalf("failed to retrieve Patient from gcs retriever: %v", err)
	}

	localRetriever, err := local.NewRetrieverFromR4Bundle([]byte(bundle))
	if err != nil {
		t.Fatalf("failed to create localRetriever: %v", err)
	}
	localResult, err := localRetriever.Retrieve(context.Background(), "Patient")
	if err != nil {
		t.Fatalf("failed to retrieve Patient from local retriever: %v", err)
	}

	if diff := cmp.Diff(gcsResult, localResult, protocmp.Transform()); diff != "" {
		t.Errorf("retrieved Patient from gcs retriever diff from local retriever (-want +got):\n%s", diff)
	}

}
