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

package parser

import (
	"errors"
	"testing"

	"github.com/google/cql/result"
)

var ErrorBadFile = errors.New("bad file")

func TestLibraryErrors(t *testing.T) {
	// Create a LibraryError
	libErr := &LibraryErrors{LibKey: result.LibKey{Name: "TESTLIB", Version: "1.0.0"}}
	libErr.Append(&ParsingError{
		Message:  "first parsing error",
		Line:     1,
		Column:   1,
		Type:     InternalError,
		Severity: ErrorSeverityError,
	})
	libErr.Append(&ParsingError{
		Message:  "second parsing error",
		Line:     2,
		Column:   2,
		Type:     InternalError,
		Severity: ErrorSeverityError,
		Cause:    ErrorBadFile,
	})

	// Make the library error a generic error
	var gotErr error = libErr

	// Can access LibraryErrors with errors.As
	var wantLibErr *LibraryErrors
	ok := errors.As(gotErr, &wantLibErr)
	if !ok {
		t.Errorf("errors.As(*LibraryErrors) = false, want true")
	}

	// Unwraps the first parsing error
	var wantParsingErr *ParsingError
	ok = errors.As(gotErr, &wantParsingErr)
	if !ok {
		t.Errorf("errors.As(*ParsingErrors) = false, want true")
	}
	if wantParsingErr.Message != "first parsing error" {
		t.Errorf("wantParsingErr.Message = %q, want %q", wantParsingErr.Message, "first parsing error")
	}

	// Is unwraps all errors in the error tree
	ok = errors.Is(gotErr, ErrorBadFile)
	if !ok {
		t.Errorf("errors.Is(gotErr, ErrorBadFile) = false, want true")
	}

	wantErrorString := `error(s) in Library "TESTLIB 1.0.0":
1-1 first parsing error
2-2 second parsing error: bad file`
	if gotErr.Error() != wantErrorString {
		t.Errorf("gotErr.Error() = %q, want %q", gotErr.Error(), wantErrorString)
	}
}

func TestParameterErrors(t *testing.T) {
	// Create a ParameterError
	paramErr := &ParameterErrors{DefKey: result.DefKey{Name: "TESTDEF", Library: result.LibKey{Name: "TESTLIB", Version: "1.0.0"}}}
	paramErr.Append(&ParsingError{
		Message:  "first parsing error",
		Line:     1,
		Column:   1,
		Type:     InternalError,
		Severity: ErrorSeverityError,
		Cause:    errors.New("first cause"),
	})
	paramErr.Append(&ParsingError{
		Message:  "second parsing error",
		Line:     2,
		Column:   2,
		Type:     InternalError,
		Severity: ErrorSeverityError,
		Cause:    ErrorBadFile,
	})

	// Make the parameter error a generic error
	var gotErr error = paramErr

	// Can access ParameterErrors with errors.As
	var wantParamErr *ParameterErrors
	ok := errors.As(gotErr, &wantParamErr)
	if !ok {
		t.Errorf("errors.As(*ParameterErrors) = false, want true")
	}

	// Unwraps the first parsing error
	var wantParsingErr *ParsingError
	ok = errors.As(gotErr, &wantParsingErr)
	if !ok {
		t.Errorf("errors.As(*ParsingErrors) = false, want true")
	}
	if wantParsingErr.Message != "first parsing error" {
		t.Errorf("wantParsingErr.Message = %q, want %q", wantParsingErr.Message, "first parsing error")
	}

	// Is unwraps all errors in the error tree
	ok = errors.Is(gotErr, ErrorBadFile)
	if !ok {
		t.Errorf("errors.Is(gotErr, ErrorBadFile) = false, want true")
	}

	wantErrorString := `1-1 first parsing error: first cause
2-2 second parsing error: bad file`
	if gotErr.Error() != wantErrorString {
		t.Errorf("gotErr.Error() = %q, want %q", gotErr.Error(), wantErrorString)
	}
}
