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

package interpreter

import (
	"context"
	"testing"
	"time"

	"github.com/google/cql/internal/embeddata"
	"github.com/google/cql/model"
	"github.com/google/cql/parser"
	"github.com/google/cql/result"
)

var defaultEvalTimestamp = time.Date(2024, 1, 1, 0, 0, 0, 0, time.FixedZone("Fixed", 4*60*60))

func defaultInterpreterConfig(t testing.TB) Config {
	t.Helper()
	fhirMI, err := embeddata.ModelInfos.ReadFile("third_party/cqframework/fhir-modelinfo-4.0.1.xml")
	if err != nil {
		t.Fatalf("internal error - could not read fhir-modelinfo-4.0.1.xml: %v", err)
	}
	p, err := parser.New(context.Background(), [][]byte{fhirMI})
	if err != nil {
		t.Fatal("Could not create Parser: ", err)
	}
	return Config{
		DataModels:          p.DataModel(),
		Retriever:           buildRetriever(t),
		Terminology:         getTerminologyProvider(t),
		EvaluationTimestamp: defaultEvalTimestamp,
		ReturnPrivateDefs:   true}
}

func wrapInLib(t *testing.T, expr model.IExpression) *model.Library {
	return &model.Library{
		Identifier: &model.LibraryIdentifier{Qualified: "TESTLIB", Version: "1.0.0"},
		Usings:     []*model.Using{{LocalIdentifier: "FHIR", Version: "4.0.1", URI: "http://hl7.org/fhir"}},
		Statements: &model.Statements{
			Defs: []model.IExpressionDef{
				&model.ExpressionDef{
					Name:       "TESTRESULT",
					Context:    "Patient",
					Expression: expr,
				},
			},
		},
	}
}

func newOrFatal(t *testing.T, a any) result.Value {
	t.Helper()
	o, err := result.New(a)
	if err != nil {
		t.Fatalf("New(%v) returned unexpected error: %v", a, err)
	}
	return o
}
