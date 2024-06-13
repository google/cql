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

// Package transforms provides Dofns for Beam jobs processing CQL.
package transforms

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	cbpb "github.com/google/cql/protos/cql_beam_go_proto"
	"github.com/google/cql/result"
	"github.com/apache/beam/sdks/v2/go/pkg/beam"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var (
	ndjsonSinkToProtoErrorCount = beam.NewCounter(counterPrefix, "ndjson_sink_to_proto_errors")
	ndjsonSinkToJSONErrorCount  = beam.NewCounter(counterPrefix, "ndjson_sink_to_json_errors")
)

func quote(s string) string {
	if len(s) == 0 {
		return s
	}
	return strconv.Quote(s)
}

// NDJSONSink marshals BeamResult to JSON and writes it to a NDJSON file.
func NDJSONSink(ctx context.Context, output *cbpb.BeamResult, emitValue func(string), emitError func(*cbpb.BeamError)) {
	libs, err := result.LibrariesFromProto(output.Result)
	if err != nil {
		ndjsonSinkToProtoErrorCount.Inc(ctx, 1)
		emitError(&cbpb.BeamError{ErrorMessage: proto.String(err.Error())})
		return
	}

	evalTime := output.EvaluationTimestamp.AsTime().In(time.UTC)
	jMap := map[string]any{
		"ID":                  output.Id,
		"EvaluationTimestamp": evalTime.Format(time.RFC3339),
		"Result":              libs,
	}

	jResult, err := json.Marshal(jMap)
	if err != nil {
		ndjsonSinkToJSONErrorCount.Inc(ctx, 1)
		emitError(&cbpb.BeamError{ErrorMessage: proto.String(err.Error())})
		return
	}
	emitValue(fmt.Sprintf("%v\n", string(jResult)))
}

// ErrorsNDJSONSink writes processing errors to an NDJSON file for troubleshooting.
func ErrorsNDJSONSink(beamErr *cbpb.BeamError, emitError func(string)) {
	jsonBytes, err := protojson.Marshal(beamErr)
	if err != nil {
		emitError(fmt.Sprintf("Failed to marshal BeamError to JSON: %v", err))
	}
	emitError(fmt.Sprintf("%v\n", string(jsonBytes)))
}
