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

package result

import (
	"testing"

	crpb "github.com/google/cql/protos/cql_result_go_proto"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestLibraries_MarshalJSON(t *testing.T) {
	tests := []struct {
		name         string
		unmarshalled Libraries
		want         string
	}{
		{
			name:         "Libraries",
			unmarshalled: Libraries{LibKey{Name: "Highly.Qualified", Version: "1.0"}: map[string]Value{"DefName": newOrFatal(t, 1)}},
			want:         `[{"libName":"Highly.Qualified","libVersion":"1.0","expressionDefinitions":{"DefName":{"@type":"System.Integer","value":1}}}]`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.unmarshalled.MarshalJSON()
			if err != nil {
				t.Fatalf("Json marshalling failed %v", err)
			}
			if diff := cmp.Diff(tc.want, string(got)); diff != "" {
				t.Errorf("json.Marshal() returned unexpected diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestLibraries_ProtoAndBack(t *testing.T) {
	tests := []struct {
		name      string
		libraries Libraries
		wantProto *crpb.Libraries
	}{
		{
			name: "Multiple expression definitions",
			libraries: Libraries{
				LibKey{Name: "TESTLIB", Version: "1.0.0"}: map[string]Value{
					"TESTRESULT":  newOrFatal(t, 1),
					"TESTRESULT2": newOrFatal(t, 2),
				},
			},
			wantProto: &crpb.Libraries{
				Libraries: []*crpb.Library{
					{
						Name:    proto.String("TESTLIB"),
						Version: proto.String("1.0.0"),
						ExprDefs: map[string]*crpb.Value{
							"TESTRESULT": {
								Value: &crpb.Value_IntegerValue{IntegerValue: 1},
							},
							"TESTRESULT2": {
								Value: &crpb.Value_IntegerValue{IntegerValue: 2},
							},
						},
					},
				},
			},
		},
		{
			name: "Multiple libraries",
			libraries: Libraries{
				LibKey{Name: "TESTLIB", Version: "1.0.0"}: map[string]Value{"TESTRESULT": newOrFatal(t, 1)},
				LibKey{Name: "TESTLIB", Version: "2.0.0"}: map[string]Value{"TESTRESULT": newOrFatal(t, 4)},
			},
			wantProto: &crpb.Libraries{
				Libraries: []*crpb.Library{
					{
						Name:    proto.String("TESTLIB"),
						Version: proto.String("1.0.0"),
						ExprDefs: map[string]*crpb.Value{
							"TESTRESULT": {
								Value: &crpb.Value_IntegerValue{IntegerValue: 1},
							},
						},
					},
					{
						Name:    proto.String("TESTLIB"),
						Version: proto.String("2.0.0"),
						ExprDefs: map[string]*crpb.Value{
							"TESTRESULT": {
								Value: &crpb.Value_IntegerValue{IntegerValue: 4},
							},
						},
					},
				},
			},
		},
		{
			name:      "Empty Library",
			libraries: Libraries{},
			wantProto: &crpb.Libraries{Libraries: []*crpb.Library{}},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotProto, err := tc.libraries.Proto()
			if err != nil {
				t.Errorf("Proto() returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantProto, gotProto, protocmp.Transform(), protocmp.SortRepeatedFields(&crpb.Libraries{}, "libraries")); diff != "" {
				t.Errorf("Proto() returned unexpected diff (-want +got):\n%s", diff)
			}

			gotLib, err := LibrariesFromProto(gotProto)
			if err != nil {
				t.Errorf("LibrariesFromProto() returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.libraries, gotLib); diff != "" {
				t.Errorf("LibrariesFromProto() returned unexpected diff (-want +got):\n%s", diff)
			}
		})
	}
}
