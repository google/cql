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
}

type resourceKey struct {
	URL     string
	Version string
}

func (l *LocalFHIRProvider) findCodeSystem(codeSystemURL, codeSystemVersion string) (fhirCodeSystem, error) {
	var vs fhirCodeSystem
	var ok bool
	if codeSystemVersion == "" {
		vs, ok = l.latestCodeSystems[codeSystemURL]
	} else {
		vs, ok = l.codeSystems[resourceKey{codeSystemURL, codeSystemVersion}]
	}

	if !ok {
		return fhirCodeSystem{}, fmt.Errorf("could not find CodeSystem{%s, %s} %w", codeSystemURL, codeSystemVersion, ErrResourceNotLoaded)
	}
	return vs, nil
}

func (l *LocalFHIRProvider) findValueSet(valueSetURL, valueSetVersion string) (fhirValueSet, error) {
	var vs fhirValueSet
	var ok bool
	if valueSetVersion == "" {
		vs, ok = l.latestValuesets[valueSetURL]
	} else {
		vs, ok = l.valueSets[resourceKey{valueSetURL, valueSetVersion}]
	}

	if !ok {
		return fhirValueSet{}, fmt.Errorf("could not find ValueSet{%s, %s} %w", valueSetURL, valueSetVersion, ErrResourceNotLoaded)
	}
	return vs, nil
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

	r, err := l.findValueSet(valuesetURL, valuesetVersion)
	if err != nil {
		// The desired ValueSet didn't exist but found a CodeSystem with this key.
		if _, err := l.findCodeSystem(valuesetURL, valuesetVersion); err == nil {
			return false, fmt.Errorf("could not find ValueSet{%s, %s} found CodeSystem instead. %w", valuesetURL, valuesetVersion, ErrIncorrectResourceType)
		}
		return false, err
	}

	for _, c := range codes {
		foundCode := r.code(c.key())
		if foundCode != nil {
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
// simple version string comparison.
func (l *LocalFHIRProvider) ExpandValueSet(valueSetURL, valueSetVersion string) ([]*Code, error) {
	if l == nil {
		return nil, ErrNotInitialized
	}

	r, err := l.findValueSet(valueSetURL, valueSetVersion)
	if err != nil {
		return nil, err
	}

	return r.codes(), nil
}

// A base fhirResource that is used to store top level data from parsed json resources. This struct
// exists to perform initial parsing of json resources so we can figure out the type of the resource
// (CodeSystem or ValueSet).
type fhirResource struct {
	ResourceType string `json:"resourceType"`
	URL          string `json:"url"`
	Version      string `json:"version"`
	// Only one of the following two fields should be populated
	Concept   []*Code    `json:"concept"`
	Expansion *expansion `json:"expansion"`
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
}

func (f *fhirValueSet) code(key codeKey) *Code {
	return f.CodeMap[key]
}

func (f *fhirValueSet) key() resourceKey {
	return resourceKey{f.URL, f.Version}
}

func (f *fhirValueSet) codes() []*Code {
	return f.Expansion.Codes
}

type expansion struct {
	Codes []*Code `json:"contains"`
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

	for _, c := range vs.Expansion.Codes {
		vs.CodeMap[c.key()] = c
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
