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

package testhelpers_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/cql/internal/testhelpers"
	"github.com/google/go-cmp/cmp"
)

func TestWriteJSONs(t *testing.T) {
	// Sanity check for WriteJSONs.
	jsons := []string{
		`{"key": "value"}`,
		`random bytes`,
		`{"one": {"two": 3}}`,
	}

	dir := testhelpers.WriteJSONs(t, jsons)

	for i, json := range jsons {
		file := filepath.Join(dir, fmt.Sprintf("file_%d.json", i))
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("os.ReadFile(%v) unexpected err: %v", file, err)
		}
		if !cmp.Equal(string(data), json) {
			t.Errorf("WriteJSONs incorrect file contents for index %d. got: %v, want: %v", i, data, json)
		}
	}

	// Check number of files in dir:
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("os.ReadDir(%v) unexpected err: %v", dir, err)
	}
	if len(files) != len(jsons) {
		t.Fatalf("len of files in output directory = %v, want %v", len(files), len(jsons))
	}

}
