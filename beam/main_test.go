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

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	cbpb "github.com/google/cql/protos/cql_beam_go_proto"
	crpb "github.com/google/cql/protos/cql_result_go_proto"
	"github.com/apache/beam/sdks/v2/go/pkg/beam"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/testing/ptest"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/lithammer/dedent"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestPipeline(t *testing.T) {
	_, _, fhirBundleDir := directorySetup(t, cqlLibs, valueSets, fhirBundles)

	tests := []struct {
		name       string
		cfg        *pipelineConfig
		wantOutput []*cbpb.BeamResult
		wantError  []*cbpb.BeamError
	}{
		{
			name: "Successful CQL Eval",
			cfg: &pipelineConfig{
				CQL: []string{dedent.Dedent(
					`library EvalTest version '1.0'
					using FHIR version '4.0.1'
					valueset "DiabetesVS": 'https://example.com/vs/glucose'
					define HasDiabetes: exists([Condition: "DiabetesVS"])
					`,
				)},
				ValueSets:           valueSets,
				FHIRBundleDir:       fhirBundleDir,
				NDJSONOutputDir:     t.TempDir(),
				EvaluationTimestamp: time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
			},
			wantOutput: []*cbpb.BeamResult{
				{
					Id:                  proto.String("1"),
					EvaluationTimestamp: timestamppb.New(time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC)),
					Result: &crpb.Libraries{
						Libraries: []*crpb.Library{
							{
								Name:    proto.String("BeamMetadata"),
								Version: proto.String("1.0.0"),
								ExprDefs: map[string]*crpb.Value{
									"ID": {
										Value: &crpb.Value_StringValue{StringValue: "1"},
									},
								},
							},
							{
								Name:    proto.String("EvalTest"),
								Version: proto.String("1.0"),
								ExprDefs: map[string]*crpb.Value{
									"HasDiabetes": {
										Value: &crpb.Value_BooleanValue{BooleanValue: true},
									},
									"DiabetesVS": {
										Value: &crpb.Value_ValueSetValue{
											ValueSetValue: &crpb.ValueSet{Id: proto.String("https://example.com/vs/glucose"), Version: proto.String("")},
										},
									},
								},
							},
						},
					},
				},
			},
			wantError: []*cbpb.BeamError{},
		},
		{
			name: "CQL Eval Error",
			cfg: &pipelineConfig{
				CQL: []string{dedent.Dedent(
					`library "EvalTest" version '1.0'
					using FHIR version '4.0.1'
					valueset "DiabetesVS": 'urn:example:nosuchvalueset'
					define HasDiabetes: exists([Condition: "DiabetesVS"])`)},
				ValueSets:           valueSets,
				FHIRBundleDir:       fhirBundleDir,
				NDJSONOutputDir:     t.TempDir(),
				EvaluationTimestamp: time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
			},
			wantOutput: []*cbpb.BeamResult{},
			wantError: []*cbpb.BeamError{
				{
					ErrorMessage: proto.String("failed during CQL evaluation: EvalTest 1.0, failed to evaluate expression HasDiabetes: could not find ValueSet{urn:example:nosuchvalueset, } resource not loaded"),
					SourceUri:    proto.String("bundle:bundle1"),
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if runtime.GOOS == "windows" {
				// https://github.com/google/cql/issues/32
				// For some reason the bundle file is not being closed properly on Windows.
				t.Skip("Skipping test on Windows due to io error")
			}
			p, s := beam.NewPipelineWithRoot()
			result, errors := buildPipeline(s, test.cfg)
			beam.ParDo0(s, diffEvalResults, beam.Impulse(s), beam.SideInput{Input: beam.CreateList(s, test.wantOutput)}, beam.SideInput{Input: result})
			beam.ParDo0(s, diffEvalErrors, beam.Impulse(s), beam.SideInput{Input: beam.CreateList(s, test.wantError)}, beam.SideInput{Input: errors})
			if err := ptest.Run(p); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func diffEvalResults(_ []byte, iterWant, iterGot func(**cbpb.BeamResult) bool) error {
	var got, want []*cbpb.BeamResult
	var v *cbpb.BeamResult
	for iterGot(&v) {
		got = append(got, v)
	}
	for iterWant(&v) {
		want = append(want, v)
	}

	sortOutputs := func(a, b *cbpb.BeamResult) bool {
		if a.GetEvaluationTimestamp().GetSeconds() != b.GetEvaluationTimestamp().GetSeconds() {
			return a.GetEvaluationTimestamp().GetSeconds() < b.GetEvaluationTimestamp().GetSeconds()
		}
		return false
	}

	if diff := cmp.Diff(want, got, cmpopts.SortSlices(sortOutputs), protocmp.Transform(), protocmp.SortRepeatedFields(&crpb.Libraries{}, "libraries")); diff != "" {
		return fmt.Errorf("BeamResult unexpected differences (-want +got):\n%s", diff)
	}
	return nil
}

func diffEvalErrors(_ []byte, iterWant, iterGot func(**cbpb.BeamError) bool) error {
	var got, want []*cbpb.BeamError
	var v *cbpb.BeamError
	for iterGot(&v) {
		got = append(got, v)
	}
	for iterWant(&v) {
		want = append(want, v)
	}

	if diff := cmp.Diff(want, got, protocmp.Transform()); diff != "" {
		return fmt.Errorf("BeamError unexpected differences (-want +got):\n%s", diff)
	}
	return nil
}

func TestBuildConfig(t *testing.T) {
	cqlDir, terminologyDir, _ := directorySetup(t, cqlLibs, valueSets, fhirBundles)

	tests := []struct {
		name  string
		flags *beamFlags
		want  *pipelineConfig
	}{
		{
			name: "with evaluation timestamp",
			flags: &beamFlags{
				CQLDir:              cqlDir,
				FHIRTerminologyDir:  terminologyDir,
				FHIRBundleDir:       "fhirBundleDir",
				EvaluationTimestamp: "2024-01-01T00:00:00Z",
				ReturnPrivateDefs:   true,
				NDJSONOutputDir:     "ndjsonOutputDir",
			},
			want: &pipelineConfig{
				CQL:                 cqlLibs,
				ValueSets:           valueSets,
				FHIRBundleDir:       "fhirBundleDir",
				EvaluationTimestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				ReturnPrivateDefs:   true,
				NDJSONOutputDir:     "ndjsonOutputDir",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := buildPipelineConfig(test.flags)
			if err != nil {
				t.Fatalf("buildConfig() failed: %v", err)
			}
			if diff := cmp.Diff(got, test.want); diff != "" {
				t.Errorf("buildConfig() unexpected diff (-got +want):\n %s", diff)
			}
		})
	}
}

func TestBuildConfig_Failure(t *testing.T) {
	cqlDir, _, fhirBundleDir := directorySetup(t, cqlLibs, valueSets, fhirBundles)

	tests := []struct {
		name      string
		flags     *beamFlags
		wantError string
	}{
		{
			name:      "cql_dir not set",
			flags:     &beamFlags{},
			wantError: "cql_dir must be set",
		},
		{
			name: "fhir_bundle_dir not set",
			flags: &beamFlags{
				CQLDir: cqlDir,
			},
			wantError: "fhir_bundle_dir must be set",
		},
		{
			name: "ndjson_output_dir not set",
			flags: &beamFlags{
				CQLDir:        cqlDir,
				FHIRBundleDir: fhirBundleDir,
			},
			wantError: "ndjson_output_dir must be set",
		},
		{
			name: "invalid cql_dir",
			flags: &beamFlags{
				CQLDir:          "baddir",
				FHIRBundleDir:   fhirBundleDir,
				NDJSONOutputDir: "output",
			},
			wantError: "failed to read directory baddir",
		},
		{
			name: "invalid terminology_dir",
			flags: &beamFlags{
				CQLDir:             cqlDir,
				FHIRBundleDir:      fhirBundleDir,
				NDJSONOutputDir:    "output",
				FHIRTerminologyDir: "baddir",
			},
			wantError: "failed to read directory baddir",
		},
		{
			name: "invalid evaluation timestamp",
			flags: &beamFlags{
				EvaluationTimestamp: "invalid",
			},
			wantError: "evaluation_timestamp must be in RFC3339 format",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := buildPipelineConfig(test.flags)
			if err == nil {
				t.Fatalf("buildConfig() succeeded, want error")
			}
			if !strings.Contains(err.Error(), test.wantError) {
				t.Errorf("Unexpected error contents (%v) want (%v)", err.Error(), test.wantError)
			}
		})
	}
}

func directorySetup(t *testing.T, cqlLibs []string, valueSets []string, fhirBundles []string) (string, string, string) {
	t.Helper()
	tempDir := t.TempDir()

	cqlDir := filepath.Join(tempDir, "cqlDir")
	err := os.Mkdir(cqlDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create directory %s: %v", cqlDir, err)
	}
	for i, cql := range cqlLibs {
		err = os.WriteFile(filepath.Join(cqlDir, fmt.Sprintf("cql-%d.cql", i)), []byte(cql), 0644)
		if err != nil {
			t.Fatalf("Failed to write file %s: %v", filepath.Join(cqlDir, cql), err)
		}
	}
	// Write a non .CQL file to test that we don't read it.
	err = os.WriteFile(filepath.Join(cqlDir, "not-cql.txt"), []byte("Should not read this"), 0644)
	if err != nil {
		t.Fatalf("Failed to write file %s: %v", filepath.Join(cqlDir, "not-cql.txt"), err)
	}

	terminologyDir := filepath.Join(tempDir, "terminologyDir")
	err = os.Mkdir(terminologyDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create directory %s: %v", terminologyDir, err)
	}
	for i, vs := range valueSets {
		err = os.WriteFile(filepath.Join(terminologyDir, fmt.Sprintf("vs-%d.json", i)), []byte(vs), 0644)
		if err != nil {
			t.Fatalf("Failed to write file %s: %v", filepath.Join(terminologyDir, vs), err)
		}
	}
	// Write a non .json file to test that we don't read it.
	err = os.WriteFile(filepath.Join(terminologyDir, "not-json.txt"), []byte("Should not read this"), 0644)
	if err != nil {
		t.Fatalf("Failed to write file %s: %v", filepath.Join(terminologyDir, "not-json.txt"), err)
	}

	fhirBundleDir := filepath.Join(tempDir, "fhirBundleDir")
	err = os.Mkdir(fhirBundleDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create directory %s: %v", fhirBundleDir, err)
	}
	for i, fb := range fhirBundles {
		err = os.WriteFile(filepath.Join(fhirBundleDir, fmt.Sprintf("bundle-%d.json", i)), []byte(fb), 0644)
		if err != nil {
			t.Fatalf("Failed to write file %s: %v", filepath.Join(fhirBundleDir, fb), err)
		}
	}
	// Write a non .json file to test that we don't read it.
	err = os.WriteFile(filepath.Join(fhirBundleDir, "not-json.txt"), []byte("Should not read this"), 0644)
	if err != nil {
		t.Fatalf("Failed to write file %s: %v", filepath.Join(fhirBundleDir, "not-json.txt"), err)
	}

	return cqlDir, terminologyDir, fhirBundleDir
}

var cqlLibs = []string{
	`library lib1`,
	`library lib2`,
}

var valueSets = []string{
	`{
		"resourceType": "ValueSet",
		"url": "https://example.com/vs/blood_pressure",
		"version": "1.0.0",
		"expansion": {
			"contains": [
				{ "system": "https://example.com/system", "code": "12345" }
			]
		}
	}`,
	`{
		"resourceType": "ValueSet",
		"url": "https://example.com/vs/glucose",
		"version": "1.0.0",
		"expansion": {
			"contains": [
				{ "system": "https://example.com/system", "code": "54321" }
			]
		}
	}`,
}

var fhirBundles = []string{
	`{
		"resourceType": "Bundle",
		"id": "bundle1",
		"entry": [
			{
				"resource": {
					"resourceType": "Patient",
					"id": "1"
				}
			},
			{
				"resource": {
					"resourceType": "Condition",
					"id": "1",
					"code": {
						"coding": [{ "system": "https://example.com/system", "code": "12345", "display": "Hypertension"}]
					},
					"onsetDateTime" : "2023-10-01"
				}
			},
			{
				"resource": {
					"resourceType": "Condition",
					"id": "2",
					"code": {
						"coding": [{ "system": "https://example.com/system", "code": "54321", "display": "Diabetes"}]
					},
					"onsetDateTime" : "2024-01-02"
				}
			}
		 ]
	}`,
}

func TestMain(m *testing.M) {
	ptest.MainWithDefault(m, "direct")
}
