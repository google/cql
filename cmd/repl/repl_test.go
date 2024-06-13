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
	"strings"
	"testing"
)

func TestValidateFlags(t *testing.T) {
	tests := []struct {
		name            string
		cqlInputText    string
		bundleFileText  string
		valuesetDirText string
	}{
		{
			name:            "No input flags is valid",
			cqlInputText:    "",
			bundleFileText:  "",
			valuesetDirText: "",
		},
		{
			name:            "cql file suffix is valid",
			cqlInputText:    "/tmp/cql.cql",
			bundleFileText:  "",
			valuesetDirText: "",
		},
		{
			name:            "bundle json suffix is valid",
			cqlInputText:    "",
			bundleFileText:  "/tmp/bundle.json",
			valuesetDirText: "",
		},
		{
			name:            "valid directory is valid",
			cqlInputText:    "",
			bundleFileText:  "",
			valuesetDirText: t.TempDir(),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := validateFlags(tc.cqlInputText, tc.bundleFileText, tc.valuesetDirText); err != nil {
				t.Errorf("validateFlags() returned unexpected error: %v", err)
			}
		})
	}
}

func TestValidateFlagsError(t *testing.T) {
	tests := []struct {
		name            string
		cqlInputText    string
		bundleFileText  string
		valuesetDirText string
		wantErr         string
	}{
		{
			name:            "cql file without cql suffix returns error",
			cqlInputText:    "/tmp/cql.txt",
			bundleFileText:  "",
			valuesetDirText: "",
			wantErr:         "--cql_file flag is required to be a valid .cql",
		},
		{
			name:            "bundle json file without json suffix returns error",
			cqlInputText:    "",
			bundleFileText:  "/tmp/bundle.txt",
			valuesetDirText: "",
			wantErr:         "--bundle_file when specified, is required to be a valid json file",
		},
		{
			name:            "invalid directory returns error",
			cqlInputText:    "",
			bundleFileText:  "",
			valuesetDirText: "/my/fake/dir",
			wantErr:         "no such file or directory",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := validateFlags(tc.cqlInputText, tc.bundleFileText, tc.valuesetDirText); !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("validateFlags() returned unexpected error (-want +got):\n%s, %s", tc.wantErr, err.Error())
			}
		})
	}
}
