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

package enginetests

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/cql/interpreter"
	"github.com/google/cql/parser"
	"github.com/google/cql/result"
	"github.com/google/cql/terminology"
	"github.com/google/go-cmp/cmp"
	"github.com/lithammer/dedent"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestExpansionInValueSet(t *testing.T) {
	// Test data - a compose valueset that includes other valuesets
	composeValueSetJSON := `{
		"resourceType": "ValueSet",
		"id": "2.16.840.1.113762.1.4.1032.14",
		"url": "http://cts.nlm.nih.gov/fhir/ValueSet/2.16.840.1.113762.1.4.1032.14",
		"version": "20170526",
		"name": "Cirrhosis",
		"compose": {
			"include": [
				{
					"valueSet": ["http://cts.nlm.nih.gov/fhir/ValueSet/2.16.840.1.113762.1.4.1032.11"]
				},
				{
					"valueSet": ["http://cts.nlm.nih.gov/fhir/ValueSet/2.16.840.1.113762.1.4.1032.12"]
				}
			]
		}
	}`

	// Referenced valueset with expansion
	referencedValueSetJSON := `{
		"resourceType": "ValueSet",
		"id": "2.16.840.1.113762.1.4.1032.11",
		"url": "http://cts.nlm.nih.gov/fhir/ValueSet/2.16.840.1.113762.1.4.1032.11",
		"version": "20170526",
		"name": "Liver Cirrhosis",
		"expansion": {
			"contains": [
				{
					"system": "http://snomed.info/sct",
					"code": "19943007",
					"display": "Cirrhosis of liver"
				},
				{
					"system": "http://hl7.org/fhir/sid/icd-10-cm",
					"code": "K74.60",
					"display": "Unspecified cirrhosis of liver"
				}
			]
		}
	}`

	// Another referenced valueset
	referencedValueSetJSON2 := `{
		"resourceType": "ValueSet",
		"id": "2.16.840.1.113762.1.4.1032.12",
		"url": "http://cts.nlm.nih.gov/fhir/ValueSet/2.16.840.1.113762.1.4.1032.12",
		"version": "20170526",
		"name": "Liver Fibrosis",
		"expansion": {
			"contains": [
				{
					"system": "http://snomed.info/sct",
					"code": "197279005",
					"display": "Fibrosis of liver"
				}
			]
		}
	}`

	// Create terminology provider with expansion support
	terminologyProvider, err := terminology.NewInMemoryFHIRProvider([]string{
		composeValueSetJSON,
		referencedValueSetJSON,
		referencedValueSetJSON2,
	})
	if err != nil {
		t.Fatalf("Failed to create terminology provider: %v", err)
	}

	tests := []struct {
		name string
		cql  string
		want result.Value
	}{
		{
			name: "System.Code in compose ValueSet - should find code",
			cql: `System.Code {
				system: 'http://snomed.info/sct',
				code: '19943007'
			} in "Cirrhosis"`,
			want: newOrFatal(t, true),
		},
		{
			name: "System.Code not in compose ValueSet - should not find code",
			cql: `System.Code {
				system: 'http://snomed.info/sct',
				code: 'NOTFOUND'
			} in "Cirrhosis"`,
			want: newOrFatal(t, false),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testCQL := dedent.Dedent(fmt.Sprintf(`
				library TESTLIB version '1.0.0'
				using FHIR version '4.0.1'
				include FHIRHelpers version '4.0.1'
				valueset "Cirrhosis": 'http://cts.nlm.nih.gov/fhir/ValueSet/2.16.840.1.113762.1.4.1032.14'
				define TESTRESULT: %v`, tc.cql))

			p := newFHIRParser(t)
			parsedLibs, err := p.Libraries(context.Background(), addFHIRHelpersLib(t, testCQL), parser.Config{})
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}

			// Create custom interpreter config with our terminology provider
			config := interpreter.Config{
				DataModels:          p.DataModel(),
				Retriever:           BuildRetriever(t),
				Terminology:         terminologyProvider,
				EvaluationTimestamp: defaultEvalTimestamp,
				ReturnPrivateDefs:   true,
			}

			results, err := interpreter.Eval(context.Background(), parsedLibs, config)
			if err != nil {
				t.Fatalf("Eval returned unexpected error: %v", err)
			}

			got := getTESTRESULT(t, results)
			if diff := cmp.Diff(tc.want, got, protocmp.Transform()); diff != "" {
				t.Errorf("Evaluate diff (-want +got):\n%s", diff)
			}
		})
	}
}
