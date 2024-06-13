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

package transforms

import (
	"context"
	"reflect"
	"time"

	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	"github.com/google/cql"
	cbpb "github.com/google/cql/protos/cql_beam_go_proto"
	"github.com/google/cql/result"
	"github.com/google/cql/retriever/local"
	"github.com/google/cql/terminology"
	bpb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/resources/bundle_and_contained_resource_go_proto"
	"github.com/apache/beam/sdks/v2/go/pkg/beam"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/register"
	"google.golang.org/protobuf/proto"
)

const counterPrefix = "beam_cql"

var (
	bundleCount      = beam.NewCounter(counterPrefix, "fhir_bundles")
	bundleErrorCount = beam.NewCounter(counterPrefix, "fhir_bundle_read_errors")
	eventCount       = beam.NewCounter(counterPrefix, "events")
	errCount         = beam.NewCounter(counterPrefix, "errors")
)

func init() {
	register.DoFn4x1[context.Context, *bpb.Bundle, func(*cbpb.BeamResult), func(*cbpb.BeamError), error](&CQLEvalFn{})
	beam.RegisterType(reflect.TypeOf((*cbpb.BeamResult)(nil)))
	beam.RegisterType(reflect.TypeOf((*cbpb.BeamError)(nil)))
}

// BeamMetadata produces the ID when running the CQL.
const BeamMetadata = `
library BeamMetadata version '1.0.0'
using FHIR version '4.0.1'
context Patient
define ID: Patient.id.value
`

// CQLEvalFn is a DoFn that parses and evaluates CQL.
type CQLEvalFn struct {
	// Only exported fields are serialized.
	CQL                 []string
	ValueSets           []string
	EvaluationTimestamp time.Time
	ReturnPrivateDefs   bool
	elm                 *cql.ELM
	terminology         terminology.Provider
}

// Setup parses the CQL and initializes the terminology provider.
func (fn *CQLEvalFn) Setup() error {
	var err error
	fhirDM, err := cql.FHIRDataModel("4.0.1")
	if err != nil {
		return err
	}
	fn.elm, err = cql.Parse(context.Background(), append(fn.CQL, BeamMetadata), cql.ParseConfig{DataModels: [][]byte{fhirDM}})
	if err != nil {
		return err
	}
	fn.terminology, err = terminology.NewInMemoryFHIRProvider(fn.ValueSets)
	return err
}

func (fn *CQLEvalFn) ProcessElement(ctx context.Context, bundle *bpb.Bundle, emit func(*cbpb.BeamResult), emitError func(*cbpb.BeamError)) error {
	retriever, err := local.NewRetrieverFromR4BundleProto(bundle)
	if err != nil {
		errCount.Inc(ctx, 1)
		emitError(&cbpb.BeamError{
			ErrorMessage: proto.String(err.Error()),
			SourceUri:    proto.String(sourceURI(bundle)),
		})
		return err
	}

	res, err := fn.elm.Eval(ctx, retriever, cql.EvalConfig{Terminology: fn.terminology, EvaluationTimestamp: fn.EvaluationTimestamp, ReturnPrivateDefs: fn.ReturnPrivateDefs})
	if err != nil {
		errCount.Inc(ctx, 1)
		emitError(&cbpb.BeamError{
			ErrorMessage: proto.String(err.Error()),
			SourceUri:    proto.String(sourceURI(bundle)),
		})
		return nil
	}

	var patientID string
	p := res[result.LibKey{Name: "BeamMetadata", Version: "1.0.0"}]["ID"]
	if !result.IsNull(p) {
		patientID, err = result.ToString(p)
		if err != nil {
			errCount.Inc(ctx, 1)
			emitError(&cbpb.BeamError{
				ErrorMessage: proto.String(err.Error()),
				SourceUri:    proto.String(sourceURI(bundle)),
			})
			return nil
		}
	}

	pbResult, err := res.Proto()
	if err != nil {
		errCount.Inc(ctx, 1)
		emitError(&cbpb.BeamError{
			ErrorMessage: proto.String(err.Error()),
			SourceUri:    proto.String(sourceURI(bundle)),
		})
		return nil
	}

	evalRes := &cbpb.BeamResult{
		Id:                  proto.String(patientID),
		EvaluationTimestamp: timestamppb.New(fn.EvaluationTimestamp),
		Result:              pbResult,
	}
	emit(evalRes)
	return nil
}

func sourceURI(bundle *bpb.Bundle) string {
	// TODO(b/317813865): Fall back to different source identifiers like the first patient id
	// if the bundle has no id.
	if bundle.GetId().GetValue() != "" {
		return "bundle:" + bundle.GetId().GetValue()
	}
	return ""
}
