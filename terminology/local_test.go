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

package terminology_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/cql/internal/testhelpers"
	"github.com/google/cql/terminology"
	"github.com/google/go-cmp/cmp"
)

var testJSONResources = []string{`
			{
				"resourceType": "ValueSet",
				"id": "https://test/file1",
				"url": "https://test/file1",
				"version": "1.0.0",
				"expansion": {
					"contains": [
						{ "system": "system1", "code": "1" },
						{ "system": "system1", "code": "2" },
						{ "system": "system2", "code": "3" }
					]
				}
			}
	`,
	`
			{
				"resourceType": "ValueSet",
				"id": "https://test/file2",
				"url": "https://test/file2",
				"version": "2.0.0",
				"expansion": {
					"contains": [
						{ "system": "system3", "code": "4", "display": "four" },
						{ "system": "system3", "code": "5" },
						{ "system": "system4", "code": "6" }
					]
				}
			}
	`,
	// Should override version 1.0.0
	`
			{
				"resourceType": "ValueSet",
				"id": "https://test/file1",
				"url": "https://test/file1",
				"version": "2.0.0",
				"expansion": {
					"contains": [
						{ "system": "system1", "code": "1v2" },
						{ "system": "system1", "code": "2v2" },
						{ "system": "system2", "code": "3v2" }
					]
				}
			}
	`,
	`
			{
				"resourceType": "CodeSystem",
				"id": "https://test/file3",
				"url": "https://test/file3",
				"version": "1.0.0",
				"concept": [
				{
					"code": "sn",
					"definition": "The sniffles"
				},
				{
					"code": "sr",
					"display": "SRT",
					"definition": "A sore throat"
				}
			]
			}
	`,
	// Should override version 1.0.0
	`
			{
				"resourceType": "CodeSystem",
				"id": "https://test/file3",
				"url": "https://test/file3",
				"version": "3.0.0",
				"concept": [
				{
					"code": "snfl",
					"definition": "The sniffles"
				},
				{
					"code": "sre-thrt",
					"display": "SRT",
					"definition": "A sore throat"
				}
			]
			}
	`,
	// The following ValueSet and CodeSystem have the same URL and version.
	`
			{
				"resourceType": "ValueSet",
				"id": "https://test/file4",
				"url": "https://test/file4",
				"version": "4.0.0",
				"expansion": {
					"contains": [
						{ "system": "system1", "code": "1" },
						{ "system": "system1", "code": "2" },
						{ "system": "system2", "code": "3" }
					]
				}
			}
	`,
	`
			{
				"resourceType": "CodeSystem",
				"id": "https://test/file4",
				"url": "https://test/file4",
				"version": "4.0.0",
				"concept": [
				{
					"code": "snfl",
					"definition": "The sniffles"
				},
				{ "code": "1" },
				{ "code": "2" },
				{ "code": "3" }
			]
			}
	`,
	// empty valueset
	`
			{
				"resourceType": "ValueSet",
				"id": "https://test/emptyVS",
				"url": "https://test/emptyVS",
				"version": "1.0.0"
			}
	`,
	// empty codesystem
	`
			{
				"resourceType": "CodeSystem",
				"id": "https://test/emptyCS",
				"url": "https://test/emptyCS",
				"version": "1.0.0"
			}
	`,
}

// writeTestResources writes the standard test FHIR resources to disk, and returns the temporary dir
// where they've been written.
func writeTestResources(t *testing.T) string {
	dir := testhelpers.WriteJSONs(t, testJSONResources)

	// Write additional file to ensure the local provider skips it.
	readme := []byte("This directory contains ValueSet JSON files.\n")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), readme, 0644); err != nil {
		t.Fatalf("Unable to write test README: %v", err)
	}
	return dir
}

func testExpandValueSet(t *testing.T, lf *terminology.LocalFHIRProvider) {
	cases := []struct {
		name            string
		valueSetURL     string
		valueSetVersion string
		wantCodes       []*terminology.Code
		wantErr         error
	}{
		{
			name:            "ValueSet https://test/file2",
			valueSetURL:     "https://test/file2",
			valueSetVersion: "2.0.0",
			wantCodes: []*terminology.Code{
				{System: "system3", Code: "4", Display: "four"},
				{System: "system3", Code: "5"},
				{System: "system4", Code: "6"},
			},
		},
		{
			name:            "ValueSet https://test/file1",
			valueSetURL:     "https://test/file1",
			valueSetVersion: "1.0.0",
			wantCodes: []*terminology.Code{
				{System: "system1", Code: "1"},
				{System: "system1", Code: "2"},
				{System: "system2", Code: "3"},
			},
		},
		{
			name:            "ValueSet https://test/file1",
			valueSetURL:     "https://test/file1",
			valueSetVersion: "2.0.0",
			wantCodes: []*terminology.Code{
				{System: "system1", Code: "1v2"},
				{System: "system1", Code: "2v2"},
				{System: "system2", Code: "3v2"},
			},
		},
		{
			// Valueset no version specified
			name:        "ValueSet https://test/file1",
			valueSetURL: "https://test/file1",
			wantCodes: []*terminology.Code{
				{System: "system1", Code: "1v2"},
				{System: "system1", Code: "2v2"},
				{System: "system2", Code: "3v2"},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			codes, err := lf.ExpandValueSet(tc.valueSetURL, tc.valueSetVersion)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("Expand(%v, %v) unexpected error. got: %v, want: %v", tc.valueSetURL, tc.valueSetVersion, err, tc.wantErr)
			}
			if diff := cmp.Diff(tc.wantCodes, codes); diff != "" {
				t.Errorf("Expand(%v, %v) returned diff (-want +got):\n%s", tc.valueSetURL, tc.valueSetVersion, diff)
			}
		})
	}
}

func TestLocalFHIR_Expand(t *testing.T) {
	testdir := writeTestResources(t)

	lf, err := terminology.NewLocalFHIRProvider(testdir)
	if err != nil {
		t.Fatalf("NewLocalFHIRProvider(%v) unexpected error: %v", testdir, err)
	}

	testExpandValueSet(t, lf)

}

func TestInMemoryFHIR_Expand(t *testing.T) {
	imf, err := terminology.NewInMemoryFHIRProvider(testJSONResources)
	if err != nil {
		t.Fatalf("NewInMemoryFHIRProvider(%v) unexpected error: %v", testJSONResources, err)
	}

	testExpandValueSet(t, imf)
}

func testResourceInCodeSystem(t *testing.T, lf *terminology.LocalFHIRProvider) {
	cases := []struct {
		name    string
		URL     string
		Version string
		Codes   []terminology.Code
		wantIn  bool
	}{
		{
			name:    "Code in CodeSystem https://test/file3",
			URL:     "https://test/file3",
			Version: "3.0.0",
			Codes:   []terminology.Code{{Code: "snfl", System: "https://test/file3"}},
			wantIn:  true,
		},
		{
			name:    "One Code in CodeSystem https://test/file3",
			URL:     "https://test/file3",
			Version: "3.0.0",
			Codes: []terminology.Code{
				{Code: "asthma", System: "https://test/file3"},
				{Code: "snfl", System: "https://test/file3"},
			},
			wantIn: true,
		},
		{
			name:    "Code not in CodeSystem https://test/file3",
			URL:     "https://test/file3",
			Version: "3.0.0",
			Codes:   []terminology.Code{{Code: "asthma", System: "https://test/file3"}},
			wantIn:  false,
		},
		{
			name:   "Code in CodeSystem https://test/file3 latest",
			URL:    "https://test/file3",
			Codes:  []terminology.Code{{Code: "snfl", System: "https://test/file3"}},
			wantIn: true,
		},
		{
			name:   "Code not in CodeSystem https://test/file3 latest",
			URL:    "https://test/file3",
			Codes:  []terminology.Code{{Code: "sn", System: "https://test/file3"}},
			wantIn: false,
		},
		{
			name:    "Code in CodeSystem https://test/file4 when ValueSet with same key exists",
			URL:     "https://test/file4",
			Version: "4.0.0",
			Codes:   []terminology.Code{{Code: "1", System: "https://test/file4"}},
			wantIn:  true,
		},
		{
			name:   "Code not in Empty CodeSystem https://test/emptyCS",
			URL:    "https://test/emptyCS",
			Codes:  []terminology.Code{{Code: "1", System: "https://test/emptyCS"}},
			wantIn: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in, err := lf.AnyInCodeSystem(tc.Codes, tc.URL, tc.Version)
			if err != nil {
				t.Errorf("In(%v, %v, %v) returned unexpected error: %v", tc.Codes, tc.URL, tc.Version, err)
			}
			if !cmp.Equal(tc.wantIn, in) {
				t.Errorf("In(%v, %v, %v) incorrect. got: %v, want: %v ", tc.Codes, tc.URL, tc.Version, in, tc.wantIn)
			}
		})
	}
}

func testResourceInValueSet(t *testing.T, lf *terminology.LocalFHIRProvider) {
	cases := []struct {
		name    string
		URL     string
		Version string
		Codes   []terminology.Code
		wantIn  bool
	}{
		{
			name:    "Code not in ValueSet https://test/file1",
			URL:     "https://test/file1",
			Version: "1.0.0",
			Codes:   []terminology.Code{{System: "system3", Code: "4"}},
			wantIn:  false,
		},
		{
			name:    "Code in ValueSet https://test/file1",
			URL:     "https://test/file1",
			Version: "1.0.0",
			Codes:   []terminology.Code{{System: "system2", Code: "3"}},
			wantIn:  true,
		},
		{
			name:    "One Code in ValueSet https://test/file1",
			URL:     "https://test/file1",
			Version: "1.0.0",
			Codes: []terminology.Code{
				{System: "system2", Code: "3"},
				{System: "system3", Code: "4"},
			},
			wantIn: true,
		},
		{
			name:    "Code not in ValueSet https://test/file1 v2",
			URL:     "https://test/file1",
			Version: "2.0.0",
			Codes:   []terminology.Code{{System: "system2", Code: "3"}},
			wantIn:  false,
		},
		{
			name:   "Code not in ValueSet https://test/file1 latest",
			URL:    "https://test/file1",
			Codes:  []terminology.Code{{System: "system2", Code: "3"}},
			wantIn: false,
		},
		{
			name:   "Code in ValueSet https://test/file1 latest",
			URL:    "https://test/file1",
			Codes:  []terminology.Code{{System: "system2", Code: "3v2"}},
			wantIn: true,
		},
		{
			name:    "Code in ValueSet https://test/file4 when CodeSystem with same key exists",
			URL:     "https://test/file4",
			Version: "4.0.0",
			Codes:   []terminology.Code{{System: "system1", Code: "1"}},
			wantIn:  true,
		},
		{
			name:   "Code not in Empty ValueSet https://test/emptyVS",
			URL:    "https://test/emptyVS",
			Codes:  []terminology.Code{{System: "system1", Code: "1"}},
			wantIn: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in, err := lf.AnyInValueSet(tc.Codes, tc.URL, tc.Version)
			if err != nil {
				t.Errorf("In(%v, %v, %v) unexpected error. got: %v", tc.Codes, tc.URL, tc.Version, err)
			}
			if !cmp.Equal(tc.wantIn, in) {
				t.Errorf("In(%v, %v, %v) incorrect. got: %v, want: %v ", tc.Codes, tc.URL, tc.Version, in, tc.wantIn)
			}
		})
	}
}

func TestLocalFHIR_In(t *testing.T) {
	testdir := writeTestResources(t)
	lf, err := terminology.NewLocalFHIRProvider(testdir)
	if err != nil {
		t.Fatalf("NewLocalFHIRProvider(%v) unexpected error: %v", testdir, err)
	}
	testResourceInCodeSystem(t, lf)
	testResourceInValueSet(t, lf)
}

func TestInMemoryFHIR_In(t *testing.T) {
	imf, err := terminology.NewInMemoryFHIRProvider(testJSONResources)
	if err != nil {
		t.Fatalf("NewInMemoryFHIRProvider(%v) unexpected error: %v", testJSONResources, err)
	}

	testResourceInCodeSystem(t, imf)
	testResourceInValueSet(t, imf)
}

func testResourceInError(t *testing.T, lf *terminology.LocalFHIRProvider) {
	cases := []struct {
		name         string
		ResourceType string
		URL          string
		Version      string
		Code         terminology.Code
		wantErr      error
	}{
		{
			name:    "ErrResourceNotLoaded_MissingURL",
			URL:     "https://test/file20",
			Version: "1.0.0",
			wantErr: terminology.ErrResourceNotLoaded,
		},
		{
			name:    "ErrResourceNotLoaded_MissingVersion",
			URL:     "https://test/file1",
			Version: "4.0.0",
			wantErr: terminology.ErrResourceNotLoaded,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := lf.AnyInValueSet([]terminology.Code{tc.Code}, tc.URL, tc.Version)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("InValueSet(%v, %v) unexpected error. got: %v, want: %v", tc.URL, tc.Version, err, tc.wantErr)
			}

			_, err = lf.AnyInCodeSystem([]terminology.Code{tc.Code}, tc.URL, tc.Version)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("InCodeSystem(%v, %v) unexpected error. got: %v, want: %v", tc.URL, tc.Version, err, tc.wantErr)
			}
		})
	}
}

func TestLocalFHIR_InError(t *testing.T) {
	testdir := writeTestResources(t)
	lf, err := terminology.NewLocalFHIRProvider(testdir)
	if err != nil {
		t.Fatalf("NewLocalFHIRProvider(%v) unexpected error: %v", testdir, err)
	}
	testResourceInError(t, lf)
}

func TestInMemoryFHIR_InError(t *testing.T) {
	imf, err := terminology.NewInMemoryFHIRProvider(testJSONResources)
	if err != nil {
		t.Fatalf("NewInMemoryFHIRProvider(%v) unexpected error: %v", testJSONResources, err)
	}

	testResourceInError(t, imf)
}

func TestLocalFHIR_NotInitialized(t *testing.T) {
	var tp *terminology.LocalFHIRProvider

	if _, err := tp.AnyInCodeSystem([]terminology.Code{{"", "", ""}}, "", ""); !errors.Is(err, terminology.ErrNotInitialized) {
		t.Errorf("In() on nil provider got unexpected error. got: %v, want: %v", err, terminology.ErrNotInitialized)
	}

	if _, err := tp.AnyInValueSet([]terminology.Code{{"", "", ""}}, "", ""); !errors.Is(err, terminology.ErrNotInitialized) {
		t.Errorf("In() on nil provider got unexpected error. got: %v, want: %v", err, terminology.ErrNotInitialized)
	}
	if _, err := tp.ExpandValueSet("", ""); !errors.Is(err, terminology.ErrNotInitialized) {
		t.Errorf("Expand() on nil provider got unexpected error. got: %v, want: %v", err, terminology.ErrNotInitialized)
	}
}
