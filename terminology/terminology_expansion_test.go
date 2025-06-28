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

func TestExpandValueSet(t *testing.T) {
	// Create test ValueSets
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

	// Create a ValueSet with a circular reference
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

	// ValueSet with an expansion already included
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

	// Create provider with the test ValueSets
	testValueSets := []string{vsParent, vsChild1, vsChild2, vsGrandchild, vsCircular, vsCircularRef, vsWithExpansion}
	provider, err := NewInMemoryFHIRProvider(testValueSets)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	// Helper function to compare Code slices ignoring order
	compareCodeSlices := func(t *testing.T, got, want []*Code) {
		t.Helper()
		if len(got) != len(want) {
			t.Errorf("Code count mismatch: got %d, want %d", len(got), len(want))
			return
		}

		// Sort both slices by code value for comparison
		sort.Slice(got, func(i, j int) bool { return got[i].Code < got[j].Code })
		sort.Slice(want, func(i, j int) bool { return want[i].Code < want[j].Code })

		for i := range got {
			if !reflect.DeepEqual(got[i], want[i]) {
				t.Errorf("Code mismatch at index %d:\ngot: %+v\nwant: %+v", i, got[i], want[i])
			}
		}
	}

	t.Run("ExpandValueSet with existing expansion", func(t *testing.T) {
		codes, err := provider.ExpandValueSet("http://example.org/fhir/ValueSet/with-expansion", "")
		if err != nil {
			t.Fatalf("ExpandValueSet failed: %v", err)
		}

		want := []*Code{
			{System: "http://example.org/fhir/CodeSystem/expansion", Code: "expansion-code1", Display: "Expansion Code 1"},
			{System: "http://example.org/fhir/CodeSystem/expansion", Code: "expansion-code2", Display: "Expansion Code 2"},
		}

		compareCodeSlices(t, codes, want)
	})

	t.Run("ExpandValueSet with compose that has direct concepts", func(t *testing.T) {
		codes, err := provider.ExpandValueSet("http://example.org/fhir/ValueSet/child1", "")
		if err != nil {
			t.Fatalf("ExpandValueSet failed: %v", err)
		}

		want := []*Code{
			{System: "http://example.org/fhir/CodeSystem/child1", Code: "child1-code1", Display: "Child 1 Code 1"},
			{System: "http://example.org/fhir/CodeSystem/child1", Code: "child1-code2", Display: "Child 1 Code 2"},
		}

		compareCodeSlices(t, codes, want)
	})

	t.Run("ExpandValueSet with nested ValueSet references", func(t *testing.T) {
		codes, err := provider.ExpandValueSet("http://example.org/fhir/ValueSet/parent", "")
		if err != nil {
			t.Fatalf("ExpandValueSet failed: %v", err)
		}

		want := []*Code{
			{System: "http://example.org/fhir/CodeSystem/parent", Code: "parent-code1", Display: "Parent Code 1"},
			{System: "http://example.org/fhir/CodeSystem/child1", Code: "child1-code1", Display: "Child 1 Code 1"},
			{System: "http://example.org/fhir/CodeSystem/child1", Code: "child1-code2", Display: "Child 1 Code 2"},
			{System: "http://example.org/fhir/CodeSystem/child2", Code: "child2-code1", Display: "Child 2 Code 1"},
			{System: "http://example.org/fhir/CodeSystem/grandchild", Code: "grandchild-code1", Display: "Grandchild Code 1"},
		}

		compareCodeSlices(t, codes, want)
	})

	t.Run("ExpandValueSet with circular references", func(t *testing.T) {
		_, err := provider.ExpandValueSet("http://example.org/fhir/ValueSet/circular", "")
		if err == nil {
			t.Errorf("Expected error for circular reference, got nil")
		} else if !errors.Is(err, ErrCircularReference) {
			t.Errorf("Expected circular reference error, got: %v", err)
		}
	})

	t.Run("AnyInValueSet with expanded ValueSet", func(t *testing.T) {
		testCases := []struct {
			name     string
			codes    []Code
			expected bool
		}{
			{
				name: "Code exists in parent",
				codes: []Code{
					{System: "http://example.org/fhir/CodeSystem/parent", Code: "parent-code1"},
				},
				expected: true,
			},
			{
				name: "Code exists in child",
				codes: []Code{
					{System: "http://example.org/fhir/CodeSystem/child1", Code: "child1-code2"},
				},
				expected: true,
			},
			{
				name: "Code exists in grandchild",
				codes: []Code{
					{System: "http://example.org/fhir/CodeSystem/grandchild", Code: "grandchild-code1"},
				},
				expected: true,
			},
			{
				name: "Code does not exist",
				codes: []Code{
					{System: "http://example.org/fhir/CodeSystem/unknown", Code: "unknown-code"},
				},
				expected: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result, err := provider.AnyInValueSet(tc.codes, "http://example.org/fhir/ValueSet/parent", "")
				if err != nil {
					t.Fatalf("AnyInValueSet failed: %v", err)
				}
				if result != tc.expected {
					t.Errorf("AnyInValueSet got %v, want %v", result, tc.expected)
				}
			})
		}
	})

	// Test caching
	t.Run("ExpandValueSet caching", func(t *testing.T) {
		// First call should expand and cache
		_, err := provider.ExpandValueSet("http://example.org/fhir/ValueSet/parent", "")
		if err != nil {
			t.Fatalf("First ExpandValueSet failed: %v", err)
		}

		// Check if it's in the cache
		key := resourceKey{"http://example.org/fhir/ValueSet/parent", ""}
		_, ok := provider.expandedCache[key]
		if !ok {
			t.Errorf("ValueSet expansion not cached as expected")
		}

		// Second call should use the cache
		codes, err := provider.ExpandValueSet("http://example.org/fhir/ValueSet/parent", "")
		if err != nil {
			t.Fatalf("Second ExpandValueSet failed: %v", err)
		}

		want := []*Code{
			{System: "http://example.org/fhir/CodeSystem/parent", Code: "parent-code1", Display: "Parent Code 1"},
			{System: "http://example.org/fhir/CodeSystem/child1", Code: "child1-code1", Display: "Child 1 Code 1"},
			{System: "http://example.org/fhir/CodeSystem/child1", Code: "child1-code2", Display: "Child 1 Code 2"},
			{System: "http://example.org/fhir/CodeSystem/child2", Code: "child2-code1", Display: "Child 2 Code 1"},
			{System: "http://example.org/fhir/CodeSystem/grandchild", Code: "grandchild-code1", Display: "Grandchild Code 1"},
		}

		compareCodeSlices(t, codes, want)
	})
}

func TestExpandValueSetEdgeCases(t *testing.T) {
	// Test empty compose
	vsEmptyCompose := `{
		"resourceType": "ValueSet",
		"url": "http://example.org/fhir/ValueSet/empty-compose",
		"version": "1.0.0",
		"compose": {
			"include": []
		}
	}`

	// Test missing referenced ValueSet
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

	// Test ValueSet with no compose or expansion
	vsNoComposeOrExpansion := `{
		"resourceType": "ValueSet",
		"url": "http://example.org/fhir/ValueSet/no-compose-expansion",
		"version": "1.0.0"
	}`

	testValueSets := []string{vsEmptyCompose, vsMissingRef, vsNoComposeOrExpansion}
	provider, err := NewInMemoryFHIRProvider(testValueSets)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	t.Run("ExpandValueSet with empty compose", func(t *testing.T) {
		codes, err := provider.ExpandValueSet("http://example.org/fhir/ValueSet/empty-compose", "")
		if err != nil {
			t.Fatalf("ExpandValueSet failed: %v", err)
		}
		if len(codes) != 0 {
			t.Errorf("Expected empty result, got %d codes", len(codes))
		}
	})

	t.Run("ExpandValueSet with missing reference", func(t *testing.T) {
		_, err := provider.ExpandValueSet("http://example.org/fhir/ValueSet/missing-ref", "")
		if err == nil {
			t.Errorf("Expected error for missing reference, got nil")
		} else if !errors.Is(err, ErrResourceNotLoaded) {
			t.Errorf("Expected resource not loaded error, got: %v", err)
		}
	})

	t.Run("ExpandValueSet with no compose or expansion", func(t *testing.T) {
		codes, err := provider.ExpandValueSet("http://example.org/fhir/ValueSet/no-compose-expansion", "")
		if err != nil {
			t.Fatalf("ExpandValueSet failed: %v", err)
		}
		if len(codes) != 0 {
			t.Errorf("Expected empty result, got %d codes", len(codes))
		}
	})

	t.Run("ExpandValueSet with nonexistent ValueSet", func(t *testing.T) {
		_, err := provider.ExpandValueSet("http://example.org/fhir/ValueSet/nonexistent", "")
		if err == nil {
			t.Errorf("Expected error for nonexistent ValueSet, got nil")
		} else if !errors.Is(err, ErrResourceNotLoaded) {
			t.Errorf("Expected resource not loaded error, got: %v", err)
		}
	})
}

func TestExpandValueSetWithSystemFilters(t *testing.T) {
	// Test ValueSet with system-based includes (no specific concepts)
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

	// Test ValueSet with system and filter
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

	testValueSets := []string{vsSystemFilter, vsSystemWithFilter}
	provider, err := NewInMemoryFHIRProvider(testValueSets)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	t.Run("ExpandValueSet with system filter (no concepts)", func(t *testing.T) {
		codes, err := provider.ExpandValueSet("http://example.org/fhir/ValueSet/system-filter", "")
		if err != nil {
			t.Fatalf("ExpandValueSet failed: %v", err)
		}
		// Should return empty since we don't have the actual CodeSystem to expand from
		if len(codes) != 0 {
			t.Errorf("Expected empty result for system filter without CodeSystem, got %d codes", len(codes))
		}
	})

	t.Run("ExpandValueSet with system and filter", func(t *testing.T) {
		codes, err := provider.ExpandValueSet("http://example.org/fhir/ValueSet/system-with-filter", "")
		if err != nil {
			t.Fatalf("ExpandValueSet failed: %v", err)
		}
		// Should return empty since we don't support filters yet
		if len(codes) != 0 {
			t.Errorf("Expected empty result for system with filter, got %d codes", len(codes))
		}
	})
}

func TestProviderNotInitialized(t *testing.T) {
	var provider *LocalFHIRProvider

	t.Run("ExpandValueSet on nil provider", func(t *testing.T) {
		_, err := provider.ExpandValueSet("test", "")
		if !errors.Is(err, ErrNotInitialized) {
			t.Errorf("Expected ErrNotInitialized, got: %v", err)
		}
	})

	t.Run("AnyInValueSet on nil provider", func(t *testing.T) {
		_, err := provider.AnyInValueSet([]Code{}, "test", "")
		if !errors.Is(err, ErrNotInitialized) {
			t.Errorf("Expected ErrNotInitialized, got: %v", err)
		}
	})
}
