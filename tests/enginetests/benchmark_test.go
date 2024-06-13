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
	"testing"

	"github.com/google/cql/interpreter"
	"github.com/google/cql/parser"
	"github.com/google/cql/result"
)

// forceBenchResult ensures the results are used so the compiler does not optimize away the
// EvalLibraries function.
var forceBenchResult result.Libraries

func BenchmarkInterpreter(b *testing.B) {
	benchmarks := []struct {
		name string
		cql  string
	}{
		{
			name: "Addition",
			cql:  "1 + 2",
		},
	}

	for _, bc := range benchmarks {
		p := newFHIRParser(b)
		parsedLibs, err := p.Libraries(context.Background(), wrapInLib(b, bc.cql), parser.Config{})
		if err != nil {
			b.Fatalf("Parse Libraries returned unexpected error: %v", err)
		}

		config := interpreter.Config{
			DataModels:          p.DataModel(),
			Retriever:           BuildRetriever(b),
			Terminology:         buildTerminologyProvider(b),
			EvaluationTimestamp: defaultEvalTimestamp,
			ReturnPrivateDefs:   true}

		b.Run(bc.name, func(b *testing.B) {
			var force result.Libraries
			for n := 0; n < b.N; n++ {
				force, err = interpreter.Eval(context.Background(), parsedLibs, config)
				if err != nil {
					b.Fatalf("Eval returned unexpected error: %v", err)
				}
			}
			forceBenchResult = force
			b.ReportAllocs()
		})
	}
}
