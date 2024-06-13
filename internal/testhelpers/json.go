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

// Package testhelpers is an internal package providing useful test helpers for the CQL engine
// project.
package testhelpers

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// WriteJSONs writes each string in jsons to a JSON file in a temporary test directory, which
// is returned.
func WriteJSONs(t testing.TB, jsons []string) (dir string) {
	t.Helper()
	dir = t.TempDir()
	for i, json := range jsons {
		if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("file_%d.json", i)), []byte(json), 0644); err != nil {
			t.Fatalf("Unable to write test json: %v", err)
		}
	}
	return dir
}
