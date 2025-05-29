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

package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/google/cql/result"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/bulk_fhir_tools/testhelpers"
)

const testBucketName = "bucketName"

func TestCLI(t *testing.T) {
	tests := []struct {
		name                       string
		cql                        string
		fhirBundle                 string
		fhirTerminology            string
		fhirParameters             string
		returnPrivateDefs          bool
		executionTimestampOverride string
		wantTestResult             string
	}{
		{
			name: "Simple CLI call with most flags set",
			cql: `
			library TESTLIB
			define TESTRESULT: true`,
			fhirBundle:      `{"resourceType": "Bundle", "id": "example", "entry": []}`,
			fhirTerminology: `{"resourceType": "ValueSet", "id": "https://test/emptyVS", "url": "https://test/emptyVS"}`,
			fhirParameters:  `{"resourceType": "Parameters", "id": "example", "parameter": []}`,
			wantTestResult:  `{"@type": "System.Boolean", "value": true}`,
		},
		{
			name: "ReturnPrivateDefs is set and returned",
			cql: `
			library TESTLIB
			define private TESTRESULT: true`,
			fhirBundle:        `{"resourceType": "Bundle", "id": "example", "entry": []}`,
			returnPrivateDefs: true,
			wantTestResult:    `{"@type": "System.Boolean", "value": true}`,
		},
		{
			name: "Can override execution timestamp",
			cql: `
			library TESTLIB
			define TESTRESULT: Now()`,
			fhirBundle:                 `{"resourceType": "Bundle", "id": "example", "entry": []}`,
			executionTimestampOverride: "@2018-02-02T15:02:03.000-04:00",
			wantTestResult:             `{"@type": "System.DateTime","value": "@2018-02-02T15:02:03.000-04:00"}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create temp directories for each of the file based flags.
			testDirCfg := defaultCLIConfig(t)
			bundleFileName := "test_bundle.json"
			var bundleFilePath string
			if tc.fhirBundle != "" {
				bundleFilePath = filepath.Join(testDirCfg.FHIRBundleDir, bundleFileName)
				if runtime.GOOS == "windows" {
					// We need to add an extra escape for windows to work because the file path is interpolated
					// into a JSON string value.
					bundleFilePath = strings.ReplaceAll(bundleFilePath, "\\", "\\\\")
				}
			}
			// Fill test directories with test file content.
			writeLocalFileWithContent(t, filepath.Join(testDirCfg.CQLDir, "test_code.cql"), tc.cql)
			writeLocalFileWithContent(t, filepath.Join(testDirCfg.FHIRBundleDir, bundleFileName), tc.fhirBundle)
			writeLocalFileWithContent(t, filepath.Join(testDirCfg.FHIRTerminologyDir, "terminology.json"), tc.fhirTerminology)
			// need to not always create this file
			writeLocalFileWithContent(t, testDirCfg.FHIRParametersFile, tc.fhirParameters)

			// Setup the CLI config.
			cfg := cliConfig{
				CQLDir:                     testDirCfg.CQLDir,
				JSONOutputDir:              testDirCfg.JSONOutputDir,
				ReturnPrivateDefs:          tc.returnPrivateDefs,
				ExecutionTimestampOverride: tc.executionTimestampOverride,
			}
			if tc.fhirTerminology != "" {
				cfg.FHIRTerminologyDir = testDirCfg.FHIRTerminologyDir
			}
			if tc.fhirBundle != "" {
				cfg.FHIRBundleDir = testDirCfg.FHIRBundleDir
			}
			if tc.fhirParameters != "" {
				cfg.FHIRParametersFile = testDirCfg.FHIRParametersFile
			}

			if err := mainWrapper(context.Background(), cfg); err != nil {
				t.Errorf("mainWrapper() returned an unexpected error: %v", err)
			}

			// Read and validate the output file.
			entries, err := os.ReadDir(testDirCfg.JSONOutputDir)
			if err != nil {
				t.Errorf("os.ReadDir() returned an unexpected error: %v", err)
			}
			if len(entries) != 1 {
				t.Errorf("os.ReadDir() expected %v entries got: %v", 1, len(entries))
			}
			resultBytes, err := os.ReadFile(path.Join(testDirCfg.JSONOutputDir, entries[0].Name()))
			if err != nil {
				t.Errorf("os.ReadFile() returned an unexpected error: %v", err)
			}
			gotResult := string(normalizeJSON(t, resultBytes))
			// all tests should have a bundle source and a tc.want which will be populated here.
			wantResult := string(normalizeJSON(t, []byte(fmt.Sprintf(`{
				"bundleSource": "%s",
				"evalResults": [
					{
						"expressionDefinitions": {
							"TESTRESULT": %s
						},
						"libName": "TESTLIB",
						"libVersion": ""
					}
				]
			}`, bundleFilePath, tc.wantTestResult))))
			if diff := cmp.Diff(wantResult, gotResult); diff != "" {
				t.Errorf("mainWrapper() returned an unexpected diff (-want +got): %v", diff)
			}
		})
	}
}

func TestCLIWithGCS(t *testing.T) {
	cql := `
	library TESTLIB
	define TESTRESULT: true`
	parametersFile := "fhir_parameters/parameters.json"
	cfg, gcsServer := defaultGCSConfig(t)
	cfg.FHIRParametersFile = gcsPath(t, parametersFile)
	gcsServer.AddObject(testBucketName, "cql/test_code.cql", gcsObject(t, cql))
	gcsServer.AddObject(testBucketName, "fhir_bundle/test_bundle.json", gcsObject(t, `{"resourceType": "Bundle", "id": "example", "entry": []}`))
	gcsServer.AddObject(testBucketName, "fhir_terminology/terminology.json", gcsObject(t, `{"resourceType": "ValueSet", "id": "https://test/emptyVS", "url": "https://test/emptyVS"}`))
	gcsServer.AddObject(testBucketName, parametersFile, gcsObject(t, `{"resourceType": "Parameters", "id": "example", "parameter": []}`))

	if err := mainWrapper(context.Background(), cfg); err != nil {
		t.Errorf("mainWrapper() returned an unexpected error: %v", err)
	}

	entry, found := gcsServer.GetObject(testBucketName, "json_output/test_bundle.json")
	if !found {
		t.Errorf("mainWrapper() did not write the expected output file")
	}
	gotResult := string(normalizeJSON(t, entry.Data))
	wantResult := string(normalizeJSON(t, []byte(`{
		"bundleSource": "gs://bucketName/fhir_bundle/test_bundle.json",
		"evalResults": [
			{
				"expressionDefinitions": {
					"TESTRESULT": {
        		"@type": "System.Boolean",
        		"value": true
					}
				},
				"libName": "TESTLIB",
				"libVersion": ""
			}
		]
	}`)))
	if diff := cmp.Diff(wantResult, gotResult); diff != "" {
		t.Errorf("mainWrapper() returned an unexpected diff (-want +got): %v", diff)
	}
}

func TestVersionOverridesCQLExecution(t *testing.T) {
	// Create a temp directory for each of the file based flags.
	cqlDir := t.TempDir()
	jsonOutputDir := t.TempDir()
	// Create simple example file content.
	simpleCQL := []byte(`
	library TESTLIB
	define TESTRESULT: true`)
	// Write the files to their respective temp directories.
	if err := os.WriteFile(path.Join(cqlDir, "test_code.cql"), simpleCQL, 0644); err != nil {
		t.Fatalf("os.WriteFile() returned an unexpected error: %v", err)
	}
	// Setup the CLI config.
	cfg := cliConfig{
		CQLDir:        cqlDir,
		JSONOutputDir: jsonOutputDir,
		Version:       true,
	}

	if err := mainWrapper(context.Background(), cfg); err != nil {
		t.Errorf("mainWrapper() returned an unexpected error: %v", err)
	}
	entries, err := os.ReadDir(jsonOutputDir)
	if err != nil {
		t.Errorf("os.ReadDir() returned an unexpected error: %v", err)
	}
	// CQL engine should not have written any output.
	if len(entries) != 0 {
		t.Errorf("os.ReadDir() expected no entries got: %v", len(entries))
	}
}

func TestParseFHIRParameters(t *testing.T) {
	tests := []struct {
		name           string
		parametersText string
		want           map[result.DefKey]string
	}{
		{
			name: "empty",
			parametersText: `{
				"resourceType": "Parameters",
				"id": "example",
				"parameter": []
			}`,
			want: map[result.DefKey]string{},
		},
		{
			name: "boolean",
			parametersText: `{
				"resourceType": "Parameters",
				"id": "example",
				"parameter": [
					{
						"name": "boolean value",
						"valueBoolean": true
					}
				]
			}`,
			want: map[result.DefKey]string{
				{Name: "boolean value"}: "true",
			},
		},
		{
			name: "decimal",
			parametersText: `{
				"resourceType": "Parameters",
				"id": "example",
				"parameter": [
					{
						"name": "decimal value",
						"valueDecimal": 1.1
					}
				]
			}`,
			want: map[result.DefKey]string{
				{Name: "decimal value"}: "1.1",
			},
		},
		{
			name: "integer",
			parametersText: `{
				"resourceType": "Parameters",
				"id": "example",
				"parameter": [
					{
						"name": "integer value",
						"valueInteger": 42
					}
				]
			}`,
			want: map[result.DefKey]string{
				{Name: "integer value"}: "42",
			},
		},
		{
			name: "positive integer",
			parametersText: `{
				"resourceType": "Parameters",
				"id": "example",
				"parameter": [
					{
						"name": "positive integer value",
						"valuePositiveInt": 42
					}
				]
			}`,
			want: map[result.DefKey]string{
				{Name: "positive integer value"}: "42",
			},
		},
		{
			name: "unsigned integer",
			parametersText: `{
				"resourceType": "Parameters",
				"id": "example",
				"parameter": [
					{
						"name": "unsigned integer value",
						"valueUnsignedInt": 43
					}
				]
			}`,
			want: map[result.DefKey]string{
				{Name: "unsigned integer value"}: "43",
			},
		},
		// This currently isn't particularly useful but we will need to support it properly in the future.
		{
			name: "code",
			parametersText: `{
				"resourceType": "Parameters",
				"id": "example",
				"parameter": [
					{
						"name": "code value",
						"valueCode": "value: 'code_value'"
					}
				]
			}`,
			want: map[result.DefKey]string{
				{Name: "code value"}: "value: 'code_value'",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseFHIRParameters([]byte(tc.parametersText))
			if err != nil {
				t.Errorf("parseFHIRParameters() with text %s returned an unexpected error: %v", string(tc.parametersText), err)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("parseFHIRParameters() with text %s returned an unexpected diff (-want +got): %v", string(tc.parametersText), diff)
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	testDirs := defaultCLIConfig(t)

	tests := []struct {
		name string
		args []string
		cfg  cliConfig
	}{
		{
			name: "Simple config with all flags set",
			cfg: cliConfig{
				Parameters:         "aString='string value'",
				CQLDir:             testDirs.CQLDir,
				FHIRBundleDir:      testDirs.FHIRBundleDir,
				FHIRTerminologyDir: testDirs.FHIRTerminologyDir,
				FHIRParametersFile: testDirs.FHIRParametersFile,
				JSONOutputDir:      testDirs.JSONOutputDir,
			},
		},
		{
			name: "Only valid cqlDir is required",
			cfg: cliConfig{
				CQLDir: testDirs.CQLDir,
			},
		},
		{
			name: "valid terminologyDir",
			cfg: cliConfig{
				CQLDir:             testDirs.CQLDir,
				FHIRTerminologyDir: testDirs.FHIRTerminologyDir,
			},
		},
		{
			name: "valid jsonOutputDir",
			cfg: cliConfig{
				CQLDir:        testDirs.CQLDir,
				JSONOutputDir: testDirs.JSONOutputDir,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := validateConfig(context.Background(), &tc.cfg); err != nil {
				t.Errorf("validateFlags() with %v returned an unexpected error %v", tc.cfg, err)
			}
		})
	}
}

func TestValidateConfigError(t *testing.T) {
	tests := []struct {
		name    string
		cfg     cliConfig
		wantErr error
	}{
		{
			name:    "cqlDir is required",
			cfg:     cliConfig{},
			wantErr: errMissingFlag,
		},
		{
			name: "bundleDir invalid path",
			cfg: cliConfig{
				CQLDir:        t.TempDir(),
				FHIRBundleDir: "/bad/path",
			},
			wantErr: fs.ErrNotExist,
		},
		{
			name: "terminologyDir invalid path",
			cfg: cliConfig{
				CQLDir:             t.TempDir(),
				FHIRTerminologyDir: "/bad/path",
			},
			wantErr: fs.ErrNotExist,
		},
		{
			name: "jsonOutputDir invalid path",
			cfg: cliConfig{
				CQLDir:        t.TempDir(),
				JSONOutputDir: "/bad/path",
			},
			wantErr: fs.ErrNotExist,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := validateConfig(context.Background(), &tc.cfg); !errors.Is(err, tc.wantErr) {
				t.Errorf("validateFlags() with %v returned unexpected error wanted %s, got %s", tc.cfg, tc.wantErr.Error(), err.Error())
			}
		})
	}
}

func TestParseFlags(t *testing.T) {
	testDirs := defaultCLIConfig(t)

	tests := []struct {
		name string
		args []string
		want cliConfig
	}{
		{
			name: "Simple config with all flags set",
			args: []string{
				`--parameters=aString='string value'`,
				"--cql_dir=" + testDirs.CQLDir,
				"--fhir_bundle_dir=" + testDirs.FHIRBundleDir,
				"--fhir_terminology_dir=" + testDirs.FHIRTerminologyDir,
				"--fhir_parameters_file=" + testDirs.FHIRParametersFile,
				"--json_output_dir=" + testDirs.JSONOutputDir,
			},
			want: cliConfig{
				Parameters:         "aString='string value'",
				CQLDir:             testDirs.CQLDir,
				FHIRBundleDir:      testDirs.FHIRBundleDir,
				FHIRTerminologyDir: testDirs.FHIRTerminologyDir,
				FHIRParametersFile: testDirs.FHIRParametersFile,
				JSONOutputDir:      testDirs.JSONOutputDir,
				gcsEndpoint:        "https://storage.googleapis.com/",
			},
		},
		{
			name: "No flags set",
			args: []string{},
			want: cliConfig{
				gcsEndpoint: "https://storage.googleapis.com/",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fs := flag.NewFlagSet("test_flagset", flag.PanicOnError)
			var cfg cliConfig
			cfg.RegisterFlags(fs)

			if err := fs.Parse(tc.args); err != nil {
				t.Errorf("fs.Parse(%v) returned an unexpected error: %v", tc.args, err)
			}
			if diff := cmp.Diff(tc.want, cfg, cmpopts.IgnoreFields(cliConfig{}, "gcsEndpoint")); diff != "" {
				t.Errorf("After fs.Parse(%v) got an unexpected diff (-want +got): %v", tc.args, diff)
			}
		})
	}
}

func defaultCLIConfig(t *testing.T) cliConfig {
	t.Helper()
	return cliConfig{
		CQLDir:             t.TempDir(),
		FHIRBundleDir:      t.TempDir(),
		FHIRTerminologyDir: t.TempDir(),
		FHIRParametersFile: filepath.Join(t.TempDir(), "parameters.json"),
		JSONOutputDir:      t.TempDir(),
	}
}

func defaultGCSConfig(t *testing.T) (cliConfig, *testhelpers.GCSServer) {
	t.Helper()
	gcsServer := testhelpers.NewGCSServer(t)
	return cliConfig{
		CQLDir:             gcsPath(t, "cql"),
		FHIRBundleDir:      gcsPath(t, "fhir_bundle"),
		FHIRTerminologyDir: gcsPath(t, "fhir_terminology"),
		GCPProject:         "test-project",
		JSONOutputDir:      gcsPath(t, "json_output"),
		gcsEndpoint:        gcsServer.URL(),
	}, gcsServer
}

func gcsPath(t *testing.T, suffixPath string) string {
	t.Helper()
	return "gs://" + path.Join(testBucketName, suffixPath)
}

func gcsObject(t *testing.T, content string) testhelpers.GCSObjectEntry {
	t.Helper()
	return testhelpers.GCSObjectEntry{
		Data: []byte(content),
	}
}

func normalizeJSON(t *testing.T, b []byte) []byte {
	t.Helper()
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatalf("json.Unmarshal() returned an unexpected error: %v", err)
	}
	outBytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("json.Marshal() returned an unexpected error: %v", err)
	}
	return outBytes
}

// Only write to a local file if content is not empty.
func writeLocalFileWithContent(t *testing.T, filePath, content string) {
	t.Helper()
	if content == "" {
		return
	}
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("os.WriteFile() returned an unexpected error: %v", err)
	}
}

// newOrFatal returns a new result.Value or calls fatal on error.
func newOrFatal(t testing.TB, a any) result.Value {
	t.Helper()
	o, err := result.New(a)
	if err != nil {
		t.Fatalf("New(%v) returned unexpected error: %v", a, err)
	}
	return o
}
