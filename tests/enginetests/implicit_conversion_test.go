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
	"github.com/google/cql/model"
	"github.com/google/cql/parser"
	"github.com/google/cql/result"
	"github.com/google/cql/types"
	"github.com/google/go-cmp/cmp"
	"github.com/lithammer/dedent"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestImplicitConversions(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "Exact match over subtype",
			cql: dedent.Dedent(`
			valueset vs: 'url' version '1.0'
			define function Foo(a Vocabulary): a.id
			define function Foo(a ValueSet): a.version
			define TESTRESULT: Foo(vs)`),
			wantResult: newOrFatal(t, "1.0"),
		},
		{
			name: "Tuple subtype",
			cql: dedent.Dedent(`
			valueset vs: 'url' version '1.0'
			define function Foo(a Tuple{ apple Vocabulary }): a.apple.id
			define TESTRESULT: Foo(Tuple { apple: vs})`),
			wantResult: newOrFatal(t, "url"),
		},
		{
			name: "Cast choice",
			cql: dedent.Dedent(`
			context Patient
			define function Foo(a String): a
			define TESTRESULT: Foo('hi' as Choice<String, Integer>)`),
			wantResult: newOrFatal(t, "hi"),
		},
		{
			name: "Cast choice as null",
			cql: dedent.Dedent(`
			context Patient
			define function Foo(a String): a
			define TESTRESULT: Foo(4 as Choice<String, Integer>)`),
			wantResult: newOrFatal(t, nil),
		},
		{
			name: "Implicit conversion FHIR.Boolean to System.Boolean",
			cql: dedent.Dedent(`
			context Patient
			define function Foo(a Boolean): a
			define TESTRESULT: Foo(Patient.active)`),
			wantResult: newOrFatal(t, true),
		},
		{
			name: "Nested implicit conversion List<FHIR.Boolean> to List<System.Boolean>",
			cql: dedent.Dedent(`
			context Patient
			define function Foo(a List<Boolean>): First(a)
			define TESTRESULT: Foo({Patient.active})`),
			wantResult: newOrFatal(t, true),
		},
		{
			name: "Class instance subtype allowed property",
			cql: dedent.Dedent(`
			valueset vs: 'url' version '1.0'
			define function Foo(a Vocabulary): a.id
			define TESTRESULT: Foo(vs)`),
			wantResult: newOrFatal(t, "url"),
		},
		{
			name:       "Generic system operator implicitly converts to same type",
			cql:        "define TESTRESULT: 4 = 4.5",
			wantResult: newOrFatal(t, false),
		},
		{
			name: "Mixed list converts to same type",
			cql:  "define TESTRESULT: {4, 4.5}",
			wantResult: newOrFatal(t, result.List{
				Value:      []result.Value{newOrFatal(t, 4.0), newOrFatal(t, 4.5)},
				StaticType: &types.List{ElementType: types.Decimal}}),
		},
		{
			name: "Exact match is not fluent",
			cql: dedent.Dedent(`
			define fluent function Foo(a Decimal): a + 4
			define function Foo(a Integer): a + 3
			define TESTRESULT: 4.Foo()`),
			wantResult: newOrFatal(t, 8.0),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testCQL := fmt.Sprintf(dedent.Dedent(`
			library TESTLIB version '1.0.0'
			using FHIR version '4.0.1'
			include FHIRHelpers version '4.0.1'
			%v`), tc.cql)
			p := newFHIRParser(t)
			parsedLibs, err := p.Libraries(context.Background(), addFHIRHelpersLib(t, testCQL), parser.Config{})
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantModel, getTESTRESULTModel(t, parsedLibs)); tc.wantModel != nil && diff != "" {
				t.Errorf("Parse diff (-want +got):\n%s", diff)
			}

			results, err := interpreter.Eval(context.Background(), parsedLibs, defaultInterpreterConfig(t, p))
			if err != nil {
				t.Fatalf("Eval returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantResult, getTESTRESULT(t, results), protocmp.Transform()); diff != "" {
				t.Errorf("Eval diff (-want +got)\n%v", diff)
			}
		})
	}
}
