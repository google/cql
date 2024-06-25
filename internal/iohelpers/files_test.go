// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package iohelpers

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/bulk_fhir_tools/testhelpers"
)

const testBucketName = "bucketName"

func TestFilesWithSuffix(t *testing.T) {
	tests := []struct {
		name       string
		files      []string
		wantSuffix string
		want       []string
	}{
		{
			name:       "only valid",
			files:      []string{"f2.json", "out.json", "result.json"},
			wantSuffix: "json",
			want:       []string{"f2.json", "out.json", "result.json"},
		},
		{
			name:       "no files in dir",
			files:      []string{},
			wantSuffix: "json",
			want:       []string{},
		},
		{
			name:       "invalid and valid",
			files:      []string{"result.json", "result.tmp", "result.txt"},
			wantSuffix: "json",
			want:       []string{"result.json"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			// Write out test files.
			for i, f := range tc.files {
				if i < len(tc.want) {
					tc.want[i] = filepath.Join(dir, tc.want[i])
				}
				filePath := filepath.Join(dir, f)
				err := os.WriteFile(filePath, []byte("Hello World"), 0644)
				if err != nil {
					t.Fatalf("Failed to write file: %v", err)
				}
			}

			got, err := FilesWithSuffix(context.Background(), dir, tc.wantSuffix, nil)
			if err != nil {
				t.Fatalf("Failed to get files: %v", err)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("FilesWithSuffix() returned an unexpected diff (-want +got): %v", diff)
			}
		})
	}
}

func TestFilesWithSuffixGCS(t *testing.T) {
	tests := []struct {
		name       string
		files      []string
		wantSuffix string
		want       []string
	}{
		{
			name:       "only valid",
			files:      []string{"f2.json", "out.json", "result.json"},
			wantSuffix: "json",
			want: []string{
				gcsPath(t, "dir/f2.json"),
				gcsPath(t, "dir/out.json"),
				gcsPath(t, "dir/result.json"),
			},
		},
		{
			name:       "no files in dir",
			files:      []string{},
			wantSuffix: "json",
			want:       []string{},
		},
		{
			name:       "invalid and valid",
			files:      []string{"result.json", "result.tmp", "result.txt"},
			wantSuffix: "json",
			want:       []string{gcsPath(t, "dir/result.json")},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gcsServer := testhelpers.NewGCSServer(t)
			for _, f := range tc.files {
				gcsServer.AddObject(testBucketName, "dir/"+f, gcsObject(t, "Hello World"))
			}

			got, err := FilesWithSuffix(context.Background(), "gs://"+testBucketName+"/dir", tc.wantSuffix, &IOConfig{GCSEndpoint: gcsServer.URL()})
			if err != nil {
				t.Fatalf("Failed to get files: %v", err)
			}
			sort.Strings(got)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("FilesWithSuffix() returned an unexpected diff (-want +got): %v", diff)
			}
		})
	}
}

func TestReadFile(t *testing.T) {
	d := t.TempDir()
	filePath := filepath.Join(d, "result.json")
	want := `Hello World`
	err := os.WriteFile(filePath, []byte(want), 0644)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	gotBytes, err := ReadFile(context.Background(), filePath, nil)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if diff := cmp.Diff(want, string(gotBytes)); diff != "" {
		t.Errorf("ReadFile() returned an unexpected diff (-want +got): %v", diff)
	}
}

func TestReadFileGCS(t *testing.T) {
	gcsServer := testhelpers.NewGCSServer(t)
	want := `Hello World`
	gcsServer.AddObject(testBucketName, "result.json", gcsObject(t, want))

	gotBytes, err := ReadFile(context.Background(), gcsPath(t, "result.json"), &IOConfig{GCSEndpoint: gcsServer.URL()})
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if diff := cmp.Diff(want, string(gotBytes)); diff != "" {
		t.Errorf("ReadFile() returned an unexpected diff (-want +got): %v", diff)
	}
}

func TestReadFileGCS_Error(t *testing.T) {
	gcsServer := testhelpers.NewGCSServer(t)

	_, err := ReadFile(context.Background(), gcsPath(t, "result.json"), &IOConfig{GCSEndpoint: gcsServer.URL()})
	if err == nil {
		t.Fatal("Should have failed to read non-existent file", err)
	}
}

func TestWriteFile(t *testing.T) {
	want := `Hello World`
	dir := t.TempDir()
	fileName := "result.json"
	filePath := filepath.Join(dir, fileName)
	err := WriteFile(context.Background(), dir, fileName, []byte(want), nil)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	gotBytes, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if diff := cmp.Diff(want, string(gotBytes)); diff != "" {
		t.Errorf("WriteFile() returned an unexpected diff (-want +got): %v", diff)
	}
}

func TestWriteFileGCS(t *testing.T) {
	gcsServer := testhelpers.NewGCSServer(t)
	want := `Hello World`
	fileName := "result.json"
	filePath := gcsPath(t, fileName)
	err := WriteFile(context.Background(), "gs://"+testBucketName, fileName, []byte(want), &IOConfig{GCSEndpoint: gcsServer.URL()})
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	gcsObject, ok := gcsServer.GetObject(testBucketName, fileName)
	if !ok {
		t.Fatalf("Failed to read gcs file: %s", filePath)
	}
	if diff := cmp.Diff(want, string(gcsObject.Data)); diff != "" {
		t.Errorf("WriteFile() returned an unexpected diff (-want +got): %v", diff)
	}
}

// gcsPath returns the full GCS path for a given suffixPath.
// Since these are not real file paths, we don't need to use filepath.Join.
func gcsPath(t *testing.T, suffixPath string) string {
	t.Helper()
	return "gs://" + testBucketName + "/" + suffixPath
}

func gcsObject(t *testing.T, content string) testhelpers.GCSObjectEntry {
	t.Helper()
	return testhelpers.GCSObjectEntry{
		Data: []byte(content),
	}
}
