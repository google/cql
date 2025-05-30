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
	"testing"
	"time"

	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	cbpb "github.com/google/cql/protos/cql_beam_go_proto"
	crpb "github.com/google/cql/protos/cql_result_go_proto"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/proto"
)

func TestNDJSONSink(t *testing.T) {
	tests := []struct {
		name          string
		outputs       []*cbpb.BeamResult
		wantValueRows []string
		wantErrors    []*cbpb.BeamError
	}{
		{
			name: "Successful Rows",
			outputs: []*cbpb.BeamResult{
				{
					Id:                  proto.String("1"),
					EvaluationTimestamp: timestamppb.New(time.Date(2023, time.November, 1, 1, 20, 30, 1e8, time.UTC)),
					Result: &crpb.Libraries{
						Libraries: []*crpb.Library{
							{
								Name:    proto.String("TESTLIB"),
								Version: proto.String("1.0.0"),
								ExprDefs: map[string]*crpb.Value{
									"HasDiabetes": {
										Value: &crpb.Value_BooleanValue{BooleanValue: false},
									},
									"HasHypertension": {
										Value: &crpb.Value_BooleanValue{BooleanValue: false},
									},
								},
							},
						},
					},
				},
				{
					Id:                  proto.String("2"),
					EvaluationTimestamp: timestamppb.New(time.Date(2023, time.December, 2, 1, 20, 30, 1e8, time.UTC)),
					Result: &crpb.Libraries{
						Libraries: []*crpb.Library{
							{
								Name:    proto.String("TESTLIB"),
								Version: proto.String("1.0.0"),
								ExprDefs: map[string]*crpb.Value{
									"HasDiabetes": {
										Value: &crpb.Value_BooleanValue{BooleanValue: true},
									},
									"HasHypertension": {
										Value: &crpb.Value_BooleanValue{BooleanValue: true},
									},
								},
							},
						},
					},
				},
			},
			wantValueRows: []string{
				"{\"EvaluationTimestamp\":\"2023-11-01T01:20:30Z\",\"ID\":\"1\",\"Result\":[{\"libName\":\"TESTLIB\",\"libVersion\":\"1.0.0\",\"expressionDefinitions\":{\"HasDiabetes\":{\"@type\":\"System.Boolean\",\"value\":false},\"HasHypertension\":{\"@type\":\"System.Boolean\",\"value\":false}}}]}\n",
				"{\"EvaluationTimestamp\":\"2023-12-02T01:20:30Z\",\"ID\":\"2\",\"Result\":[{\"libName\":\"TESTLIB\",\"libVersion\":\"1.0.0\",\"expressionDefinitions\":{\"HasDiabetes\":{\"@type\":\"System.Boolean\",\"value\":true},\"HasHypertension\":{\"@type\":\"System.Boolean\",\"value\":true}}}]}\n",
			},
		},
	}
	for _, test := range tests {
		var gotValueRows []string
		var gotErrors []*cbpb.BeamError
		emitValue := func(val string) { gotValueRows = append(gotValueRows, val) }
		emitError := func(err *cbpb.BeamError) { gotErrors = append(gotErrors, err) }
		for _, output := range test.outputs {
			NDJSONSink(context.Background(), output, emitValue, emitError)
		}

		if diff := cmp.Diff(test.wantValueRows, gotValueRows, cmpopts.SortSlices(func(a, b string) bool { return a < b })); diff != "" {
			t.Errorf("NDJSONSink() returned value diff (-want +got)\n%v", diff)
		}
		if diff := cmp.Diff(test.wantErrors, gotErrors); diff != "" {
			t.Errorf("NDJSONSink() returned error diff (-want +got):\n%s", diff)
		}
	}
}
