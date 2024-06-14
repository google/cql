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

// Package enginetests holds the most of the unit tests for the CQL Engine parser and interpreter.
package enginetests

import (
	"context"
	"embed"
	"fmt"
	"testing"
	"time"

	"github.com/google/cql/internal/embeddata"
	"github.com/google/cql/internal/resourcewrapper"
	"github.com/google/cql/interpreter"
	"github.com/google/cql/model"
	"github.com/google/cql/parser"
	"github.com/google/cql/result"
	"github.com/google/cql/retriever/local"
	"github.com/google/cql/terminology"
	"github.com/google/fhir/go/fhirversion"
	"github.com/google/fhir/go/jsonformat"
	"github.com/lithammer/dedent"
	"google.golang.org/protobuf/proto"
)

//go:embed testdata/patient_bundle.json
var patientBundle []byte

//go:embed testdata/terminology/*.json
var terminologyDir embed.FS

var defaultEvalTimestamp = time.Date(2024, 1, 1, 0, 0, 0, 0, time.FixedZone("Fixed", 4*60*60))

// BuildRetriever returns a new retriever.Retriever from the embedded patient_bundle.json.
func BuildRetriever(t testing.TB) *local.Retriever {
	t.Helper()
	ret, err := local.NewRetrieverFromR4Bundle(patientBundle)
	if err != nil {
		t.Fatalf("Failed to create retriever: %v", err)
	}
	return ret
}

// RetrieveFHIRResource retrieves the FHIR resource with the given type and ID from the embedded
// patient_bundle.json.
func RetrieveFHIRResource(t testing.TB, resourceType string, resourceID string) proto.Message {
	t.Helper()
	unmarshaller, err := jsonformat.NewUnmarshallerWithoutValidation("UTC", fhirversion.R4)
	if err != nil {
		t.Fatalf("Failed to retrieve FHIR Resource: %v", err)
	}
	containedResource, err := unmarshaller.UnmarshalR4(patientBundle)
	if err != nil {
		t.Fatalf("Failed to retrieve FHIR Resource: %v", err)
	}

	for _, e := range containedResource.GetBundle().GetEntry() {
		rw := resourcewrapper.New(e.GetResource())
		rt, err := rw.ResourceType()
		if err != nil {
			t.Fatalf("Failed to retrieve FHIR Resource: %v", err)
		}
		rid, err := rw.ResourceID()
		if err != nil {
			t.Fatalf("Failed to retrieve FHIR Resource: %v", err)
		}

		if rt == resourceType && rid == resourceID {
			msg, err := rw.ResourceMessageField()
			if err != nil {
				t.Fatalf("Failed to retrieve FHIR Resource: %v", err)
			}
			return msg
		}
	}
	t.Fatalf("Failed to retrieve FHIR Resource: could not find resource with type %v and id %v", resourceType, resourceID)
	return nil
}

func buildTerminologyProvider(t testing.TB) terminology.Provider {
	t.Helper()
	files, err := terminologyDir.ReadDir("testdata/terminology")
	if err != nil {
		t.Fatalf("Failed to read terminology folder: %v", err)
	}

	var jsons []string
	for _, f := range files {
		json, err := terminologyDir.ReadFile("testdata/terminology/" + f.Name())
		if err != nil {
			t.Fatalf("Failed to read terminology file: %v", err)
		}
		jsons = append(jsons, string(json))
	}

	tp, err := terminology.NewInMemoryFHIRProvider(jsons)
	if err != nil {
		t.Fatalf("Failed to create terminology provider: %v", err)
	}
	return tp
}

func newFHIRParser(t testing.TB) *parser.Parser {
	t.Helper()
	fhirMI, err := embeddata.ModelInfos.ReadFile("third_party/cqframework/fhir-modelinfo-4.0.1.xml")
	if err != nil {
		t.Fatalf("internal error - could not read fhir-modelinfo-4.0.1.xml: %v", err)
	}
	p, err := parser.New(context.Background(), [][]byte{fhirMI})
	if err != nil {
		t.Fatal("Could not create Parser: ", err)
	}
	return p
}

// wrapInLib wraps the cql expression in a library and expression definition.
func wrapInLib(t testing.TB, cql string) []string {
	t.Helper()
	cqlLib := dedent.Dedent(fmt.Sprintf(`
	library TESTLIB version '1.0.0'
	using FHIR version '4.0.1'
	context Patient
	define TESTRESULT: %v`, cql))

	return addFHIRHelpersLib(t, cqlLib)
}

func addFHIRHelpersLib(t testing.TB, lib string) []string {
	fhirHelpers, err := embeddata.FHIRHelpers.ReadFile("third_party/cqframework/FHIRHelpers-4.0.1.cql")
	if err != nil {
		t.Fatalf("internal error - could not read FHIRHelpers-4.0.1.cql: %v", err)
	}
	return []string{lib, string(fhirHelpers)}
}

// defaultInterpreterConfig returns an interpreter.Config with the default values used in the
// engine tests.
// The evaluation timestamp is fixed at Jan 1, 2024 +04:00.
func defaultInterpreterConfig(t testing.TB, p *parser.Parser) interpreter.Config {
	return interpreter.Config{
		DataModels:          p.DataModel(),
		Retriever:           BuildRetriever(t),
		Terminology:         buildTerminologyProvider(t),
		EvaluationTimestamp: defaultEvalTimestamp,
		ReturnPrivateDefs:   true}
}

func getTESTLIBModel(t testing.TB, parsedLibs []*model.Library) model.Library {
	t.Helper()
	for _, lib := range parsedLibs {
		if lib.Identifier.Qualified == "TESTLIB" {
			return *lib
		}
	}
	t.Fatalf("Could not find TESTLIB library")
	return model.Library{}
}

// getTESTRESULTModel finds the first TESTRESULT definition in any library and returns the model.
func getTESTRESULTModel(t testing.TB, parsedLibs []*model.Library) model.IExpression {
	t.Helper()

	for _, parsedLib := range parsedLibs {
		if parsedLib.Statements == nil {
			continue
		}
		for _, def := range parsedLib.Statements.Defs {
			if def.GetName() == "TESTRESULT" {
				return def.GetExpression()
			}
		}
	}

	t.Fatalf("Could not find TESTRESULT expression definition")
	return nil
}

// getTESTRESULT finds the first TESTRESULT definition in any library and returns the result without
// sources. The result package implementation of Equal ignores sources and is used by tests, however
// it still shows up in diffs. Using getTESTRESULT strips the sources so it doesn't show up in
// cmp.Diff.
func getTESTRESULT(t testing.TB, resultLibs result.Libraries) result.Value {
	t.Helper()
	res := getTESTRESULTWithSources(t, resultLibs)
	gotWithoutMeta, err := result.New(res.GolangValue())
	if err != nil {
		t.Fatalf("Could not find remove Meta from TESTRESULT, %v", err)
	}
	return gotWithoutMeta
}

// getTESTRESULTWithSources finds the first TESTRESULT definition in any library and returns the
// result with sources.
func getTESTRESULTWithSources(t testing.TB, resultLibs result.Libraries) result.Value {
	t.Helper()

	for _, resultLib := range resultLibs {
		for name, res := range resultLib {
			if name == "TESTRESULT" {
				return res
			}
		}
	}

	t.Fatalf("Could not find TESTRESULT expression definition")
	return result.Value{}
}

func newOrFatal(t testing.TB, a any) result.Value {
	t.Helper()
	o, err := result.New(a)
	if err != nil {
		t.Fatalf("New(%v) returned unexpected error: %v", a, err)
	}
	return o
}

func newWithSourcesOrFatal(t testing.TB, a any, expr model.IExpression, obj []result.Value) result.Value {
	t.Helper()
	o, err := result.NewWithSources(a, expr, obj...)
	if err != nil {
		t.Fatalf("New(%v) returned unexpected error: %v", a, err)
	}
	return o
}
