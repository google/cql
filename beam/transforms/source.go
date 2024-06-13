// Copyright 2024 Google LLC
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

package transforms

import (
	"context"
	"fmt"

	cbpb "github.com/google/cql/protos/cql_beam_go_proto"
	"github.com/google/fhir/go/fhirversion"
	"github.com/google/fhir/go/jsonformat"
	bpb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/resources/bundle_and_contained_resource_go_proto"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/io/fileio"
	"google.golang.org/protobuf/proto"
)

// FileToBundle returns a collection of FHIR R4 bundles and a collection of `ProcessingError` protos
// for files that could not be parsed into a bundle.
func FileToBundle(ctx context.Context, file fileio.ReadableFile, emitBundle func(*bpb.Bundle), emitError func(*cbpb.BeamError)) {
	emitErr := func(err error) {
		bundleErrorCount.Inc(ctx, 1)
		emitError(&cbpb.BeamError{ErrorMessage: proto.String(err.Error())})
	}

	data, err := file.Read(ctx)
	if err != nil {
		emitErr(err)
		return
	}

	unmarshaller, err := jsonformat.NewUnmarshallerWithoutValidation("UTC", fhirversion.R4)
	if err != nil {
		emitErr(err)
		return
	}

	p, err := unmarshaller.Unmarshal(data)
	if err != nil {
		emitErr(err)
		return
	}

	b := p.(*bpb.ContainedResource).GetBundle()
	if b != nil {
		emitBundle(b)
	} else {
		emitErr(fmt.Errorf("no bundle found in file: %s", file.Metadata.Path))
	}
}
