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

// Package terminology includes various TerminologyProviders for working with medical terminology.
package terminology

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"path/filepath"
	"strings"

)

var (
	// ErrResourceNotLoaded indicates the resource requested was not loaded.
	ErrResourceNotLoaded = errors.New("resource not loaded")
	// ErrIncorrectResourceType indicated the Resource was not of the desired type.
	ErrIncorrectResourceType = errors.New("incorrect resource type")
	// ErrNotInitialized indicates the terminology provider was not initialized.
	ErrNotInitialized = errors.New("terminology provider not initialized, so no terminology operations can be performed")
	// ErrCircularReference indicates a circular reference was detected in ValueSet compose.
	ErrCircularReference = errors.New("circular reference detected in ValueSet compose")
)

const (
	// codeSystem is the fhir string resourceType for a CodeSystem.
	codeSystem string = "CodeSystem"
	// valueSet is the fhir string resourceType for a ValueSet
	valueSet string = "ValueSet"
)

// NewLocalFHIRProvider returns a new Local FHIR terminology provider initialized with the input
// directory. If multiple ValueSets in the directory have the same ID and Version, the last one seen
// by the LocalFHIR provider will be the one loaded for use.
// TODO(b/297090333): support loading only certain ValueSets into memory, and FHIR versions if needed.
func NewLocalFHIRProvider(dir string) (*LocalFHIRProvider, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	lf := &LocalFHIRProvider{
		codeSystems:       make(map[resourceKey]fhirCodeSystem),
		valueSets:         make(map[resourceKey]fhirValueSet),
		latestCodeSystems: make(map[string]fhirCodeSystem),
		latestValuesets:   make(map[string]fhirValueSet),
		expandedCache:     make(map[resourceKey][]*Code),
	}

	for _, file := range files {
		// Skip non-JSON files, such as BUILD files, READMEs, or sub directories.
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		f, err := os.Open(filepath.Join(dir, file.Name()))
		if err != nil {
			return nil, err
		}
		defer f.Close() // Possible returned error on read is ignored, since it'll be superseeded by the decodeFHIRResource error.
		fr, err := decodeFHIRResource(f)
		if err != nil {
			return nil, err
		}

		switch fr.ResourceType {
		case codeSystem:
			lf.addCodeSystem(fr)
		case valueSet:
			lf.addValueSet(fr)
		}
	}

	return lf, nil
}

// NewInMemoryFHIRProvider returns a new Local FHIR terminology provider initialized with the JSON
// resources. If multiple ValueSets in the directory have the same ID and Version, the last one seen
// by the LocalFHIR provider will be the one loaded for use.
func NewInMemoryFHIRProvider(jsons []string) (*LocalFHIRProvider, error) {
	lf := &LocalFHIRProvider{
		codeSystems:       make(map[resourceKey]fhirCodeSystem),
		valueSets:         make(map[resourceKey]fhirValueSet),
		latestCodeSystems: make(map[string]fhirCodeSystem),
		latestValuesets:   make(map[string]fhirValueSet),
		expandedCache:     make(map[resourceKey][]*Code),
	}

	for _, json := range jsons {
		fr, err := decodeFHIRResource(strings.NewReader(json))
		if err != nil {
			return nil, err
		}

		switch fr.ResourceType {
		case codeSystem:
			lf.addCodeSystem(fr)
		case valueSet:
			lf.addValueSet(fr)
		}
	}

	return lf, nil
}

func (l *LocalFHIRProvider) addCodeSystem(fr *fhirResource) {
	cs := buildFHIRCodeSystem(*fr)
	l.codeSystems[fr.key()] = cs

	latest, ok := l.latestCodeSystems[fr.URL]
	if !ok {
		l.latestCodeSystems[fr.URL] = cs
	} else if fr.Version > latest.Version {
		l.latestCodeSystems[fr.URL] = cs
	}
}

func (l *LocalFHIRProvider) addValueSet(fr *fhirResource) {
	vs := buildFHIRValueSet(*fr)
	l.valueSets[fr.key()] = vs

	latest, ok := l.latestValuesets[fr.URL]
	if !ok {
		l.latestValuesets[fr.URL] = vs
	} else if fr.Version > latest.Version {
		l.latestValuesets[fr.URL] = vs
	}
}

// LocalFHIRProvider is a terminology provider that uses local ValueSet information on the file system.
type LocalFHIRProvider struct {
	codeSystems       map[resourceKey]fhirCodeSystem
	valueSets         map[resourceKey]fhirValueSet
	latestCodeSystems map[string]fhirCodeSystem
	latestValuesets   map[string]fhirValueSet
	expandedCache     map[resourceKey][]*Code
}

type resourceKey struct {
	URL     string
	Version string
}

func (l *LocalFHIRProvider) findCodeSystem(codeSystemURL, codeSystemVersion string) (fhirCodeSystem, error) {
	var cs fhirCodeSystem
	var ok bool
	
	// Normalize URL to handle http vs https protocol differences
	normalizedURL := normalizeURL(codeSystemURL)
	
	if codeSystemVersion == "" {
		// Try to find the CodeSystem with the normalized URL
		for url, codeSystem := range l.latestCodeSystems {
			if normalizeURL(url) == normalizedURL {
				cs = codeSystem
				ok = true
				break
			}
		}
	} else {
		// Try to find the CodeSystem with the normalized URL and version
		for key, codeSystem := range l.codeSystems {
			if normalizeURL(key.URL) == normalizedURL && key.Version == codeSystemVersion {
				cs = codeSystem
				ok = true
				break
			}
		}
	}

	if !ok {
		return fhirCodeSystem{}, fmt.Errorf("could not find CodeSystem{%s, %s} %w", codeSystemURL, codeSystemVersion, ErrResourceNotLoaded)
	}
	return cs, nil
}

func (l *LocalFHIRProvider) findValueSet(valueSetURL, valueSetVersion string) (fhirValueSet, error) {
	var vs fhirValueSet
	var ok bool
	
	// Normalize URL to handle http vs https protocol differences
	normalizedURL := normalizeURL(valueSetURL)
	
	if valueSetVersion == "" {
		// Try to find the ValueSet with the normalized URL
		for url, valueSet := range l.latestValuesets {
			if normalizeURL(url) == normalizedURL {
				vs = valueSet
				ok = true
				break
			}
		}
	} else {
		// Try to find the ValueSet with the normalized URL and version
		for key, valueSet := range l.valueSets {
			if normalizeURL(key.URL) == normalizedURL && key.Version == valueSetVersion {
				vs = valueSet
				ok = true
				break
			}
		}
	}

	if !ok {
		return fhirValueSet{}, fmt.Errorf("could not find ValueSet{%s, %s} %w", valueSetURL, valueSetVersion, ErrResourceNotLoaded)
	}
	return vs, nil
}

// normalizeURL converts a URL to a protocol-insensitive form by removing http:// or https:// prefix
func normalizeURL(url string) string {
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	return url
}

// AnyInValueSet returns true if any code is contained within the specified Valueset, otherwise
// false. Code.Display is ignored when making this determination. If the valueSetVersion is an empty
// string, this will use the 'latest' value set version based on a simple version string comparison.
// If the resource type of the found resource does not line up return a error
// https://cql.hl7.org/09-b-cqlreference.html#in-valueset
func (l *LocalFHIRProvider) AnyInValueSet(codes []Code, valuesetURL, valuesetVersion string) (bool, error) {
	if l == nil {
		return false, ErrNotInitialized
	}

	// Expand the ValueSet to get all codes
	expandedCodes, err := l.ExpandValueSet(valuesetURL, valuesetVersion)
	if err != nil {
		// The desired ValueSet didn't exist but found a CodeSystem with this key.
		if _, err := l.findCodeSystem(valuesetURL, valuesetVersion); err == nil {
			return false, fmt.Errorf("could not find ValueSet{%s, %s} found CodeSystem instead. %w", valuesetURL, valuesetVersion, ErrIncorrectResourceType)
		}
		return false, err
	}
	
	// Create a map of the expanded codes for efficient lookup
	expandedCodeMap := make(map[codeKey]bool)
	for _, code := range expandedCodes {
		expandedCodeMap[code.key()] = true
	}
	
	// Check if any of the input codes are in the expanded ValueSet
	for _, c := range codes {
		if expandedCodeMap[c.key()] {
			return true, nil
		}
	}
	
	return false, nil
}

// AnyInCodeSystem returns true if any code is contained within the specified CodeSystem, otherwise
// false. Code.Display is ignored when making this determination. If the CodeSystemVersion is an
// empty string, this will use the 'latest' resource version based on a simple version string
// comparison. If the resource type of the found resource does not line up return a error
// https://cql.hl7.org/09-b-cqlreference.html#in-code-system
func (l *LocalFHIRProvider) AnyInCodeSystem(codes []Code, codeSystemURL, codeSystemVersion string) (bool, error) {
	if l == nil {
		return false, ErrNotInitialized
	}

	r, err := l.findCodeSystem(codeSystemURL, codeSystemVersion)
	if err != nil {
		// The desired CodeSystem didn't exist but found a ValueSet with this key.
		if _, err := l.findValueSet(codeSystemURL, codeSystemVersion); err == nil {
			return false, fmt.Errorf("could not find CodeSystem{%s, %s} found ValueSet instead. %w", codeSystemURL, codeSystemVersion, ErrIncorrectResourceType)
		}
		return false, err
	}

	for _, c := range codes {
		// Retrieve the code from the FHIR resource.
		if code := r.code(c.key()); code != nil {
			return true, nil
		}
	}

	return false, nil
}

// ExpandValueSet returns the expanded codes for the provided ValueSet id and version. If the
// valueSetVersion is an empty string, this will use the 'latest' value set version based on a
// simple version string comparison. This method supports compose valuesets with recursive expansion.
func (l *LocalFHIRProvider) ExpandValueSet(valueSetURL, valueSetVersion string) ([]*Code, error) {
	if l == nil {
		return nil, ErrNotInitialized
	}

	key := resourceKey{valueSetURL, valueSetVersion}
	
	// Check cache first
	if cached, exists := l.expandedCache[key]; exists {
		return cached, nil
	}

	vs, err := l.findValueSet(valueSetURL, valueSetVersion)
	if err != nil {
		return nil, err
	}

	// Try to use existing expansion first
	if len(vs.Expansion.Codes) > 0 {
		l.expandedCache[key] = vs.Expansion.Codes
		return vs.Expansion.Codes, nil
	}

	// Expand from compose if no expansion exists
	visited := make(map[string]bool)
	expandedCodes, err := l.expandValueSetInternal(context.Background(), vs, visited)
	if err != nil {
		return nil, err
	}

	// Cache the result
	l.expandedCache[key] = expandedCodes
	return expandedCodes, nil
}

// A base fhirResource that is used to store top level data from parsed json resources. This struct
// exists to perform initial parsing of json resources so we can figure out the type of the resource
// (CodeSystem or ValueSet).
type fhirResource struct {
	ResourceType string `json:"resourceType"`
	URL          string `json:"url"`
	Version      string `json:"version"`
	// Only one of the following fields should be populated for CodeSystems
	Concept   []*Code    `json:"concept"`
	// The following fields are for ValueSets
	Expansion *expansion `json:"expansion"`
	Compose   *compose   `json:"compose"`
}

func (f *fhirResource) key() resourceKey {
	return resourceKey{f.URL, f.Version}
}

func decodeFHIRResource(i io.Reader) (*fhirResource, error) {
	r := fhirResource{}
	if err := json.NewDecoder(i).Decode(&r); err != nil {
		return nil, err
	}
	return &r, nil
}

type fhirValueSet struct {
	ResourceType string `json:"resourceType"`
	URL          string `json:"url"`
	Version      string `json:"version"`
	CodeMap      map[codeKey]*Code
	Expansion    expansion `json:"expansion"`
	Compose      compose   `json:"compose"`
}

func (f *fhirValueSet) code(key codeKey) *Code {
	return f.CodeMap[key]
}

func (f *fhirValueSet) key() resourceKey {
	return resourceKey{f.URL, f.Version}
}

func (f *fhirValueSet) codes() []*Code {
	// If we have expansion codes, return them
	if len(f.Expansion.Codes) > 0 {
		return f.Expansion.Codes
	}
	
	// Otherwise, return codes from the CodeMap (for compose ValueSets)
	// Always return a non-nil slice, even if empty
	codes := make([]*Code, 0, len(f.CodeMap))
	for _, code := range f.CodeMap {
		codes = append(codes, code)
	}
	
	// Ensure we never return nil, always return an empty slice if no codes
	if codes == nil {
		codes = []*Code{}
	}
	return codes
}

type expansion struct {
	Codes []*Code `json:"contains"`
}

// compose represents the compose section of a FHIR ValueSet
type compose struct {
	Include []include `json:"include"`
}

// include represents an include section within compose
type include struct {
	ValueSet []string `json:"valueSet"`
	Concept  []*Code  `json:"concept"`
	System   string   `json:"system"`
	// TODO: Add support for filters if needed
}

type fhirCodeSystem struct {
	ResourceType string `json:"resourceType"`
	URL          string `json:"url"`
	Version      string `json:"version"`
	CodeMap      map[codeKey]*Code
	Concept      []*Code `json:"concept"`
}

func (f *fhirCodeSystem) key() resourceKey {
	return resourceKey{f.URL, f.Version}
}

func (f *fhirCodeSystem) codes() []*Code {
	return f.Concept
}

// Retrieve a Code from the CodeSystem
func (f *fhirCodeSystem) code(key codeKey) *Code {
	if key.System != f.URL {
		return nil
	}
	// CodeSystems don't have System set on their codes.
	key.System = ""
	return f.CodeMap[key]
}

func buildFHIRValueSet(fr fhirResource) fhirValueSet {
	vs := fhirValueSet{
		ResourceType: fr.ResourceType,
		URL:          fr.URL,
		Version:      fr.Version,
		CodeMap:      make(map[codeKey]*Code),
	}
	if fr.Expansion != nil {
		vs.Expansion = *fr.Expansion
	}
	if fr.Compose != nil {
		vs.Compose = *fr.Compose
	}

	// Add codes from expansion (pre-expanded ValueSets)
	for _, c := range vs.Expansion.Codes {
		vs.CodeMap[c.key()] = c
	}

	// Add codes from compose (compose ValueSets)
	for _, include := range vs.Compose.Include {
		for _, concept := range include.Concept {
			// Set the system for the concept if it's not already set
			if concept.System == "" && include.System != "" {
				concept.System = include.System
			}
			vs.CodeMap[concept.key()] = concept
		}
	}

	return vs
}

func buildFHIRCodeSystem(fr fhirResource) fhirCodeSystem {
	cs := fhirCodeSystem{
		ResourceType: fr.ResourceType,
		URL:          fr.URL,
		Version:      fr.Version,
		Concept:      fr.Concept,
		CodeMap:      make(map[codeKey]*Code),
	}

	for _, c := range cs.Concept {
		cs.CodeMap[c.key()] = c
	}
	return cs
}

// expandValueSetInternal recursively expands a valueset from its compose definition
func (l *LocalFHIRProvider) expandValueSetInternal(ctx context.Context, vs fhirValueSet, visited map[string]bool) ([]*Code, error) {
	allCodes := make([]*Code, 0)
	
	// Check for circular reference
	vsKey := normalizeURL(vs.URL)
	if vs.Version != "" {
		vsKey = vs.URL + "|" + vs.Version
	}
	
	if visited[vsKey] {
		return nil, fmt.Errorf("circular reference detected for ValueSet %s: %w", vsKey, ErrCircularReference)
	}
	
	// Mark this ValueSet as being processed
	visited[vsKey] = true
	defer func() {
		// Remove from visited when done processing
		delete(visited, vsKey)
	}()
	
	// Parse compose from the raw JSON if not already parsed
	compose, err := l.parseCompose(vs)
	if err != nil {
		return nil, err
	}

	for _, include := range compose.Include {
		// Check for context cancellation
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		// Handle include.valueSet (recursive expansion)
		for _, vsRef := range include.ValueSet {
			refCodes, err := l.expandValueSetWithVisited(vsRef, "", visited)
			if err != nil {
				return nil, fmt.Errorf("failed to expand referenced valueset %s: %w", vsRef, err)
			}
			allCodes = append(allCodes, refCodes...)
		}

		// Handle include.concept (direct codes)
		for _, concept := range include.Concept {
			if concept != nil {
				// Set the system if not already set
				if concept.System == "" && include.System != "" {
					concept.System = include.System
				}
				allCodes = append(allCodes, concept)
			}
		}

		// TODO: Handle include.system with filters if needed
		// This would require expanding all codes from a code system
	}

	return allCodes, nil
}

// expandValueSetWithVisited expands a ValueSet while tracking visited ValueSets to prevent circular references
func (l *LocalFHIRProvider) expandValueSetWithVisited(valueSetURL, valueSetVersion string, visited map[string]bool) ([]*Code, error) {
	vs, err := l.findValueSet(valueSetURL, valueSetVersion)
	if err != nil {
		return nil, err
	}

	// Try to use existing expansion first
	if len(vs.Expansion.Codes) > 0 {
		return vs.Expansion.Codes, nil
	}

	// Expand from compose if no expansion exists
	return l.expandValueSetInternal(context.Background(), vs, visited)
}

// parseCompose extracts compose information from a valueset
func (l *LocalFHIRProvider) parseCompose(vs fhirValueSet) (compose, error) {
	// For now, return the compose data that should be available in the valueset
	// This will be properly implemented when we enhance the base parsing
	return vs.Compose, nil
}
