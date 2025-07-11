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

package terminology

import (
	"errors"
	"reflect"
	"sort"
	"testing"
)

// Helper function to compare Code slices ignoring order
func compareCodeSlices(t *testing.T, got, want []*Code) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("Code count mismatch: got %d, want %d", len(got), len(want))
		return
	}

	sort.Slice(got, func(i, j int) bool { return got[i].Code < got[j].Code })
	sort.Slice(want, func(i, j int) bool { return want[i].Code < want[j].Code })

	for i := range got {
		if !reflect.DeepEqual(got[i], want[i]) {
			t.Errorf("Code mismatch at index %d:\ngot: %+v\nwant: %+v", i, got[i], want[i])
		}
	}
}

// Main functionality test data
func getMainTestValueSets() []string {
	vsParent := `{
		"resourceType": "ValueSet",
		"url": "http://example.org/fhir/ValueSet/parent",
		"version": "1.0.0",
		"compose": {
			"include": [
				{
					"valueSet": ["http://example.org/fhir/ValueSet/child1", "http://example.org/fhir/ValueSet/child2"]
				},
				{
					"system": "http://example.org/fhir/CodeSystem/parent",
					"concept": [
						{
							"code": "parent-code1",
							"display": "Parent Code 1"
						}
					]
				}
			]
		}
	}`

	vsChild1 := `{
		"resourceType": "ValueSet",
		"url": "http://example.org/fhir/ValueSet/child1",
		"version": "1.0.0",
		"compose": {
			"include": [
				{
					"system": "http://example.org/fhir/CodeSystem/child1",
					"concept": [
						{
							"code": "child1-code1",
							"display": "Child 1 Code 1"
						},
						{
							"code": "child1-code2",
							"display": "Child 1 Code 2"
						}
					]
				}
			]
		}
	}`

	vsChild2 := `{
		"resourceType": "ValueSet",
		"url": "http://example.org/fhir/ValueSet/child2",
		"version": "1.0.0",
		"compose": {
			"include": [
				{
					"valueSet": ["http://example.org/fhir/ValueSet/grandchild"]
				},
				{
					"system": "http://example.org/fhir/CodeSystem/child2",
					"concept": [
						{
							"code": "child2-code1",
							"display": "Child 2 Code 1"
						}
					]
				}
			]
		}
	}`

	vsGrandchild := `{
		"resourceType": "ValueSet",
		"url": "http://example.org/fhir/ValueSet/grandchild",
		"version": "1.0.0",
		"compose": {
			"include": [
				{
					"system": "http://example.org/fhir/CodeSystem/grandchild",
					"concept": [
						{
							"code": "grandchild-code1",
							"display": "Grandchild Code 1"
						}
					]
				}
			]
		}
	}`

	vsCircular := `{
		"resourceType": "ValueSet",
		"url": "http://example.org/fhir/ValueSet/circular",
		"version": "1.0.0",
		"compose": {
			"include": [
				{
					"valueSet": ["http://example.org/fhir/ValueSet/circular-ref"]
				}
			]
		}
	}`

	vsCircularRef := `{
		"resourceType": "ValueSet",
		"url": "http://example.org/fhir/ValueSet/circular-ref",
		"version": "1.0.0",
		"compose": {
			"include": [
				{
					"valueSet": ["http://example.org/fhir/ValueSet/circular"]
				}
			]
		}
	}`

	vsWithExpansion := `{
		"resourceType": "ValueSet",
		"url": "http://example.org/fhir/ValueSet/with-expansion",
		"version": "1.0.0",
		"expansion": {
			"contains": [
				{
					"system": "http://example.org/fhir/CodeSystem/expansion",
					"code": "expansion-code1",
					"display": "Expansion Code 1"
				},
				{
					"system": "http://example.org/fhir/CodeSystem/expansion",
					"code": "expansion-code2",
					"display": "Expansion Code 2"
				}
			]
		}
	}`

	return []string{vsParent, vsChild1, vsChild2, vsGrandchild, vsCircular, vsCircularRef, vsWithExpansion}
}

// Edge case test data
func getEdgeCaseTestValueSets() []string {
	vsEmptyCompose := `{
		"resourceType": "ValueSet",
		"url": "http://example.org/fhir/ValueSet/empty-compose",
		"version": "1.0.0",
		"compose": {
			"include": []
		}
	}`

	vsMissingRef := `{
		"resourceType": "ValueSet",
		"url": "http://example.org/fhir/ValueSet/missing-ref",
		"version": "1.0.0",
		"compose": {
			"include": [
				{
					"valueSet": ["http://example.org/fhir/ValueSet/nonexistent"]
				}
			]
		}
	}`

	vsNoComposeOrExpansion := `{
		"resourceType": "ValueSet",
		"url": "http://example.org/fhir/ValueSet/no-compose-expansion",
		"version": "1.0.0"
	}`

	return []string{vsEmptyCompose, vsMissingRef, vsNoComposeOrExpansion}
}

// System filter test data
func getSystemFilterTestValueSets() []string {
	vsSystemFilter := `{
		"resourceType": "ValueSet",
		"url": "http://example.org/fhir/ValueSet/system-filter",
		"version": "1.0.0",
		"compose": {
			"include": [
				{
					"system": "http://example.org/fhir/CodeSystem/test"
				}
			]
		}
	}`

	vsSystemWithFilter := `{
		"resourceType": "ValueSet",
		"url": "http://example.org/fhir/ValueSet/system-with-filter",
		"version": "1.0.0",
		"compose": {
			"include": [
				{
					"system": "http://example.org/fhir/CodeSystem/test",
					"filter": [
						{
							"property": "concept",
							"op": "is-a",
							"value": "parent-concept"
						}
					]
				}
			]
		}
	}`

	return []string{vsSystemFilter, vsSystemWithFilter}
}

func TestExpandValueSet(t *testing.T) {
	tests := []struct {
		name          string
		valueSetURL   string
		version       string
		expectedCodes []*Code
	}{
		{
			name:        "existing expansion",
			valueSetURL: "http://example.org/fhir/ValueSet/with-expansion",
			version:     "",
			expectedCodes: []*Code{
				{System: "http://example.org/fhir/CodeSystem/expansion", Code: "expansion-code1", Display: "Expansion Code 1"},
				{System: "http://example.org/fhir/CodeSystem/expansion", Code: "expansion-code2", Display: "Expansion Code 2"},
			},
		},
		{
			name:        "compose with direct concepts",
			valueSetURL: "http://example.org/fhir/ValueSet/child1",
			version:     "",
			expectedCodes: []*Code{
				{System: "http://example.org/fhir/CodeSystem/child1", Code: "child1-code1", Display: "Child 1 Code 1"},
				{System: "http://example.org/fhir/CodeSystem/child1", Code: "child1-code2", Display: "Child 1 Code 2"},
			},
		},
		{
			name:        "nested ValueSet references",
			valueSetURL: "http://example.org/fhir/ValueSet/parent",
			version:     "",
			expectedCodes: []*Code{
				{System: "http://example.org/fhir/CodeSystem/parent", Code: "parent-code1", Display: "Parent Code 1"},
				{System: "http://example.org/fhir/CodeSystem/child1", Code: "child1-code1", Display: "Child 1 Code 1"},
				{System: "http://example.org/fhir/CodeSystem/child1", Code: "child1-code2", Display: "Child 1 Code 2"},
				{System: "http://example.org/fhir/CodeSystem/child2", Code: "child2-code1", Display: "Child 2 Code 1"},
				{System: "http://example.org/fhir/CodeSystem/grandchild", Code: "grandchild-code1", Display: "Grandchild Code 1"},
			},
		},
	}

	provider, err := NewInMemoryFHIRProvider(getMainTestValueSets())
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			codes, err := provider.ExpandValueSet(tt.valueSetURL, tt.version)
			if err != nil {
				t.Fatalf("ExpandValueSet failed: %v", err)
			}
			
			compareCodeSlices(t, codes, tt.expectedCodes)
		})
	}
}

func TestExpandValueSetErrors(t *testing.T) {
	tests := []struct {
		name                string
		valueSetURL         string
		version             string
		expectSpecificError error
	}{
		{
			name:                "circular references",
			valueSetURL:         "http://example.org/fhir/ValueSet/circular",
			version:             "",
			expectSpecificError: ErrCircularReference,
		},
	}

	provider, err := NewInMemoryFHIRProvider(getMainTestValueSets())
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := provider.ExpandValueSet(tt.valueSetURL, tt.version)
			if err == nil {
				t.Errorf("Expected error, got nil")
				return
			}
			if tt.expectSpecificError != nil && !errors.Is(err, tt.expectSpecificError) {
				t.Errorf("Expected specific error %v, got: %v", tt.expectSpecificError, err)
			}
		})
	}
}

func TestAnyInValueSet(t *testing.T) {
	tests := []struct {
		name     string
		codes    []Code
		valueSet string
		version  string
		expected bool
	}{
		{
			name: "code exists in parent",
			codes: []Code{
				{System: "http://example.org/fhir/CodeSystem/parent", Code: "parent-code1"},
			},
			valueSet: "http://example.org/fhir/ValueSet/parent",
			version:  "",
			expected: true,
		},
		{
			name: "code exists in child",
			codes: []Code{
				{System: "http://example.org/fhir/CodeSystem/child1", Code: "child1-code2"},
			},
			valueSet: "http://example.org/fhir/ValueSet/parent",
			version:  "",
			expected: true,
		},
		{
			name: "code exists in grandchild",
			codes: []Code{
				{System: "http://example.org/fhir/CodeSystem/grandchild", Code: "grandchild-code1"},
			},
			valueSet: "http://example.org/fhir/ValueSet/parent",
			version:  "",
			expected: true,
		},
		{
			name: "code does not exist",
			codes: []Code{
				{System: "http://example.org/fhir/CodeSystem/unknown", Code: "unknown-code"},
			},
			valueSet: "http://example.org/fhir/ValueSet/parent",
			version:  "",
			expected: false,
		},
	}

	// Create provider with the test ValueSets
	provider, err := NewInMemoryFHIRProvider(getMainTestValueSets())
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := provider.AnyInValueSet(tt.codes, tt.valueSet, tt.version)
			if err != nil {
				t.Fatalf("AnyInValueSet failed: %v", err)
			}
			
			if result != tt.expected {
				t.Errorf("AnyInValueSet got %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestExpandValueSetEdgeCases(t *testing.T) {
	tests := []struct {
		name               string
		valueSetURL        string
		version            string
		expectedCodesCount int
	}{
		{
			name:               "empty compose",
			valueSetURL:        "http://example.org/fhir/ValueSet/empty-compose",
			version:            "",
			expectedCodesCount: 0,
		},
		{
			name:               "no compose or expansion",
			valueSetURL:        "http://example.org/fhir/ValueSet/no-compose-expansion",
			version:            "",
			expectedCodesCount: 0,
		},
	}

	provider, err := NewInMemoryFHIRProvider(getEdgeCaseTestValueSets())
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			codes, err := provider.ExpandValueSet(tt.valueSetURL, tt.version)
			if err != nil {
				t.Fatalf("ExpandValueSet failed: %v", err)
			}
			
			if len(codes) != tt.expectedCodesCount {
				t.Errorf("Expected %d codes, got %d", tt.expectedCodesCount, len(codes))
			}
		})
	}
}

func TestExpandValueSetEdgeCasesErrors(t *testing.T) {
	tests := []struct {
		name                string
		valueSetURL         string
		version             string
		expectSpecificError error
	}{
		{
			name:                "missing reference",
			valueSetURL:         "http://example.org/fhir/ValueSet/missing-ref",
			version:             "",
			expectSpecificError: ErrResourceNotLoaded,
		},
		{
			name:                "nonexistent ValueSet",
			valueSetURL:         "http://example.org/fhir/ValueSet/nonexistent",
			version:             "",
			expectSpecificError: ErrResourceNotLoaded,
		},
	}

	provider, err := NewInMemoryFHIRProvider(getEdgeCaseTestValueSets())
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := provider.ExpandValueSet(tt.valueSetURL, tt.version)
			if err == nil {
				t.Errorf("Expected error, got nil")
				return
			}
			if tt.expectSpecificError != nil && !errors.Is(err, tt.expectSpecificError) {
				t.Errorf("Expected specific error %v, got: %v", tt.expectSpecificError, err)
			}
		})
	}
}

func TestExpandValueSetWithSystemFilters(t *testing.T) {
	tests := []struct {
		name               string
		valueSetURL        string
		version            string
		expectedCodesCount int
	}{
		{
			name:               "system filter (no concepts)",
			valueSetURL:        "http://example.org/fhir/ValueSet/system-filter",
			version:            "",
			expectedCodesCount: 0, // Should return empty since we don't have the actual CodeSystem to expand from
		},
		{
			name:               "system with filter",
			valueSetURL:        "http://example.org/fhir/ValueSet/system-with-filter",
			version:            "",
			expectedCodesCount: 0, // Should return empty since we don't support filters yet
		},
	}

	provider, err := NewInMemoryFHIRProvider(getSystemFilterTestValueSets())
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			codes, err := provider.ExpandValueSet(tt.valueSetURL, tt.version)
			if err != nil {
				t.Fatalf("ExpandValueSet failed: %v", err)
			}
			
			if len(codes) != tt.expectedCodesCount {
				t.Errorf("Expected %d codes, got %d", tt.expectedCodesCount, len(codes))
			}
		})
	}
}

func TestExpandValueSetCaching(t *testing.T) {
	provider, err := NewInMemoryFHIRProvider(getMainTestValueSets())
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	valueSetURL := "http://example.org/fhir/ValueSet/parent"
	version := ""

	// First call should expand and cache
	_, err = provider.ExpandValueSet(valueSetURL, version)
	if err != nil {
		t.Fatalf("First ExpandValueSet failed: %v", err)
	}

	// Check if it's in the cache
	key := resourceKey{valueSetURL, version}
	_, ok := provider.expandedCache[key]
	if !ok {
		t.Errorf("ValueSet expansion not cached as expected")
	}

	// Second call should use the cache
	codes, err := provider.ExpandValueSet(valueSetURL, version)
	if err != nil {
		t.Fatalf("Second ExpandValueSet failed: %v", err)
	}

	expectedCodes := []*Code{
		{System: "http://example.org/fhir/CodeSystem/parent", Code: "parent-code1", Display: "Parent Code 1"},
		{System: "http://example.org/fhir/CodeSystem/child1", Code: "child1-code1", Display: "Child 1 Code 1"},
		{System: "http://example.org/fhir/CodeSystem/child1", Code: "child1-code2", Display: "Child 1 Code 2"},
		{System: "http://example.org/fhir/CodeSystem/child2", Code: "child2-code1", Display: "Child 2 Code 1"},
		{System: "http://example.org/fhir/CodeSystem/grandchild", Code: "grandchild-code1", Display: "Grandchild Code 1"},
	}

	compareCodeSlices(t, codes, expectedCodes)
}

func TestProviderNotInitialized(t *testing.T) {
	tests := []struct {
		name        string
		testFunc    func(*LocalFHIRProvider) error
		expectError error
	}{
		{
			name: "ExpandValueSet on nil provider",
			testFunc: func(p *LocalFHIRProvider) error {
				_, err := p.ExpandValueSet("test", "")
				return err
			},
			expectError: ErrNotInitialized,
		},
		{
			name: "AnyInValueSet on nil provider",
			testFunc: func(p *LocalFHIRProvider) error {
				_, err := p.AnyInValueSet([]Code{}, "test", "")
				return err
			},
			expectError: ErrNotInitialized,
		},
	}

	var provider *LocalFHIRProvider

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.testFunc(provider)
			if !errors.Is(err, tt.expectError) {
				t.Errorf("Expected error %v, got: %v", tt.expectError, err)
			}
		})
	}
}
