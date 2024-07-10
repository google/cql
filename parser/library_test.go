// Copyright 2023 Google LLC
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
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/google/cql/model"
	"github.com/google/cql/result"
	"github.com/google/cql/types"
	"github.com/google/go-cmp/cmp"
	"github.com/lithammer/dedent"
)

func TestMalformedCQLSingleLibrary(t *testing.T) {
	tests := []struct {
		name        string
		cql         string
		errContains []string
		errCount    int
	}{
		{
			name:        "trival_string",
			cql:         "abcdefg",
			errContains: []string{"extraneous input 'abcdefg'"},
			errCount:    1,
		},
		{
			name: "Code references non-existent CodeSystem",
			cql: dedent.Dedent(`
        library TrivialTest version '1.2.3'
        using FHIR version '4.0.1'
        code c: '1234' from invalid_code_system_id`),
			errContains: []string{"could not resolve the local reference"},
			errCount:    1,
		},
		{
			name: "Code uses id that is not a CodeSystem",
			cql: dedent.Dedent(`
        library TrivialTest version '1.2.3'
        using FHIR version '4.0.1'
				parameter "a number" default 4000
        code c: '1234' from "a number"`),
			errContains: []string{"should be of type System.CodeSystem"},
			errCount:    1,
		},
		{
			name: "no_such_resource",
			cql: dedent.Dedent(`
        library TrivialTest version '1.2.3'
        using FHIR version '4.0.1'
        define TestPatient: ["NoSuchPatient"]
        define TestCondition: ["NoSuchCondition"]`),
			errContains: []string{"NoSuchPatient", "retrieves cannot be", "NoSuchCondition", "retrieves cannot be"},
			errCount:    4,
		},
		{
			name: "no_such_source",
			cql: dedent.Dedent(`
        library TrivialTest version '1.2.3'
        using FHIR version '4.0.1'

				define IsMale: [Patient] P
				  where BogusRef.gender = 'male'`),
			errContains: []string{"BogusRef"},
			errCount:    1,
		},
		{
			// CQL->ELM requires context to be explicitly declared before referencing it,
			// so require the same here.
			name: "missing_context",
			cql: dedent.Dedent(`
        library PatientGender version '1.2.3'
        using FHIR version '4.0.1'
				define gender: Patient.gender`),
			errContains: []string{"Patient"},
			errCount:    1,
		},
		{
			name: "no_such_expression",
			cql: dedent.Dedent(`
        library ReferenceAnotherDefine version '1.2.3'
        using FHIR version '4.0.1'
				define HasSampleObs: exists(NoSuchObs)`),
			errContains: []string{"NoSuchObs"},
			errCount:    1,
		},
		{
			name: "Query_AliasAlreadyExists",
			cql: dedent.Dedent(`
			using FHIR version '4.0.1'
			parameter "P" default 4000
			define X: [Patient] P`),
			errContains: []string{"already exists"},
			errCount:    1,
		},
		{
			name: "ExpressionDefinition_NameAlreadyExists",
			cql: dedent.Dedent(`
			using FHIR version '4.0.1'
			parameter "Population Size" default 4000
			define "Population Size": 4000`),
			errContains: []string{"already exists"},
			errCount:    1,
		},
		{
			name: "ref_invalid_expression",
			cql: dedent.Dedent(`
        library ReferenceAnotherDefine version '1.2.3'
        using FHIR version '4.0.1'
				define BogusDef: [NoSuchResource]
				define HasBogus: exists(BogusDef)`),
			errContains: []string{"NoSuchResource", "retrieves cannot be"},
			errCount:    2,
		},
		{
			name:        "ParameterDefinition_MissingTypeAndDefault",
			cql:         `parameter "Population Size"`,
			errContains: []string{"Parameter definition must include a type or a default, but neither were found"},
			errCount:    1,
		},
		{
			name:        "ParameterDefinition_DefaultAndSpecifiedTypeMismatch",
			cql:         `parameter "Population Name" String default 82`,
			errContains: []string{"Parameter definition specified type String does not match the type of default 82"},
			errCount:    1,
		},
		{
			name: "InvalidFHIRVersionUsingStatement",
			cql: dedent.Dedent(`
			library TrivialTest version '1.2.3'
			using FHIR version '3.0.0'
			`),
			errContains: []string{"FHIR 3.0.0 data model not found"},
			errCount:    1,
		},
		{
			name: "InvalidNonFHIRUsingStatement",
			cql: dedent.Dedent(`
			library TrivialTest version '1.2.3'
			using QUIC version '3.0.0'
			`),
			errContains: []string{"QUIC 3.0.0 data model not found"},
			errCount:    1,
		},
		{
			name: "invalid_FHIR_Version_for_retrieve",
			cql: dedent.Dedent(`
        library TrivialTest version '1.2.3'
        using FHIR version '3.0.0'
        define TestPatient: ["Patient"]`),
			errContains: []string{
				"FHIR 3.0.0 data model not found",
				"using declaration has not been set",
				"retrieves cannot be",
			},
			errCount: 3,
		},
		{
			name: "unsupported_interval_operator",
			cql: dedent.Dedent(`
			library intervalOperatorUnsupported version '1.2.3'
			using FHIR version '4.0.1'

			define "Has coronary heart disease":
				exists (
					[Condition] c
						where c.onset includes start of Interval[@2013-01-01T00:00:00.0, @2014-01-01T00:00:00.0)
				)`),
			errContains: []string{"unsupported interval operator in timing expression"},
			errCount:    1,
		},
		{
			name: "Unexpected Result Type in TimeBoundaryExpression operand",
			cql: `library intervalOperator version '1.2.3'
			using FHIR version '4.0.1'
			define "EndOfInterval":
					end of @2013-01-01T00:00:00.0`,
			errContains: []string{"could not resolve End(System.DateTime)"},
			errCount:    1,
		},
		{
			name: "ExpectedStringGotIdentifier.",
			cql: `library intervalOperator version '1.2.3'
			using FHIR version '4.0.1'
      valueset "My Valueset": "This should be a single-quoted string"`,
			errContains: []string{"expecting STRING"},
			errCount:    1,
		},
		{
			name: "Invalid Expression",
			cql: dedent.Dedent(`
			using FHIR version '4.0.1'
			define "Param": expand 4
				`),
			errContains: []string{"unsupported expression"},
			errCount:    1,
		},
		{
			name: "Using Declaration with Local Identifier",
			cql: dedent.Dedent(`
			using FHIR version '4.0.1' called FIRE
				`),
			errContains: []string{"Using declaration does not support local identifiers"},
			errCount:    1,
		},
		{
			name:        "Code Selector references nonexisistent CodeSystem",
			cql:         "define Foo: Code '132' from cs display 'Severed Leg'",
			errContains: []string{"could not resolve the local reference to cs"},
			errCount:    1,
		},
		{
			name: "Unsupported context",
			cql: dedent.Dedent(`
        library TrivialTest version '1.2.3'
        using FHIR version '4.0.1'
				context Practitioner`),
			errContains: []string{"error -- the CQL engine does not yet support the context"},
			errCount:    1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := newFHIRParser(t).Libraries(context.Background(), []string{test.cql}, Config{})
			if err == nil {
				t.Fatal("Parse succeeded, wanted error")
			}

			var pe *LibraryErrors
			if ok := errors.As(err, &pe); ok {
				for _, ec := range test.errContains {
					if !strings.Contains(pe.Error(), ec) {
						t.Errorf("Returned error (%s) did not contain expected string (%s)",
							err.Error(), test.errContains)
					}
				}

				if len(pe.Errors) != test.errCount {
					t.Errorf("Returned error (%s) had (%d) errors but expected (%d)",
						err.Error(), len(pe.Errors), test.errCount)
				}
			} else {
				t.Errorf("Unexpected test error (%s).", err.Error())
			}
		})
	}
}

func TestMalformedCQLMultipleLibraries(t *testing.T) {
	tests := []struct {
		name        string
		cql         string
		errContains []string
		errCount    int
	}{
		{
			name: "IncludeDef already defined",
			cql: dedent.Dedent(`
        include example.helpers1 version '1.0'
				include example.helpers2 version '1.0' called helpers1`),
			errContains: []string{"already exists"},
			errCount:    1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			libs := []string{
				dedent.Dedent(`
					library example.helpers1 version '1.0'
					define public "public def": 2
					define private "private def": 3 `),
				dedent.Dedent(`
					library example.helpers2 version '1.0'
					define public "public def": 4
					define private "private def": 5 `),
				tc.cql,
			}
			_, err := newFHIRParser(t).Libraries(context.Background(), libs, Config{})
			if err == nil {
				t.Errorf("Parsing succeeded, expected error")
			}

			var pe *LibraryErrors
			if ok := errors.As(err, &pe); ok {
				for _, ec := range tc.errContains {
					s := pe.Error()
					if !strings.Contains(s, ec) {
						t.Errorf("Returned error (%s) did not contain expected string (%s)",
							err.Error(), tc.errContains)
					}
				}

				if len(pe.Errors) != tc.errCount {
					t.Errorf("Returned error (%s) had (%d) errors but expected (%d)",
						err.Error(), len(pe.Errors), tc.errCount)
				}
			} else {
				t.Errorf("Unexpected test error (%s).", err.Error())
			}
		})
	}
}

func TestMalformedIncludeDependenciesMultipleLibraries(t *testing.T) {
	tests := []struct {
		name    string
		cqlLibs []string
		wantErr string
	}{
		{
			name: "library includes itself",
			cqlLibs: []string{
				dedent.Dedent(`
					library lib1
					include lib1`),
			},
			wantErr: "found circular dependencies",
		},
		{
			name: "library includes non-existent library",
			cqlLibs: []string{
				"include lib1",
			},
			wantErr: "lib1 does not exist",
		},
		{
			name: "library includes non-existent library with version specified",
			cqlLibs: []string{
				"include lib1 version '1.0'",
			},
			wantErr: "lib1 1.0 does not exist",
		},
		{
			name: "repeated library identifier",
			cqlLibs: []string{
				dedent.Dedent(`
					library lib1
					include lib2`),
				dedent.Dedent(`
					library lib1
					include lib3`),
			},
			wantErr: `cql library "lib1" already imported`,
		},
		{
			name: "includes non-existant library",
			cqlLibs: []string{
				dedent.Dedent(`
					library lib1
					include lib2`),
			},
			wantErr: "failed to import library",
		},
		{
			name: "Directly circular include",
			cqlLibs: []string{
				dedent.Dedent(`
					library lib1
					include lib2`),
				dedent.Dedent(`
					library lib2
					include lib1`),
			},
			wantErr: "found circular dependencies",
		},
		{
			name: "Indirectly circular include",
			cqlLibs: []string{
				dedent.Dedent(`
					library lib1
					include lib2`),
				dedent.Dedent(`
					library lib2
					include lib3`),
				dedent.Dedent(`
					library lib3
					include lib1`),
			},
			wantErr: "found circular dependencies",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := newFHIRParser(t).Libraries(context.Background(), tc.cqlLibs, Config{})
			if err == nil {
				t.Fatal("Parse succeeded, wanted error")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("Returned error (%s) did not contain expected string (%s)", err.Error(), tc.wantErr)
			}
		})
	}
}

func TestNoLibraries(t *testing.T) {
	_, err := newFHIRParser(t).Libraries(context.Background(), []string{}, Config{})
	if err == nil {
		t.Fatal("Parse succeeded, wanted error")
	}
	want := "no CQL libraries were provided"
	if !strings.Contains(err.Error(), want) {
		t.Errorf("Returned error (%s) did not contain expected string (%s)", err.Error(), want)
	}
}

func TestParserTopologicalSortMultipleLibraries(t *testing.T) {
	tests := []struct {
		name    string
		cqlLibs []string
		want    []*model.Library
	}{
		{
			name: "Deps declared in reverse order",
			cqlLibs: []string{
				dedent.Dedent(`
					include lib1
					include lib2`),
				dedent.Dedent(`
				library lib2
				include lib1`),
				"library lib1",
			},
			want: []*model.Library{
				&model.Library{
					Identifier: &model.LibraryIdentifier{Local: "lib1", Qualified: "lib1"},
				},
				&model.Library{
					Identifier: &model.LibraryIdentifier{Local: "lib2", Qualified: "lib2"},
					Includes: []*model.Include{
						{Identifier: &model.LibraryIdentifier{Local: "lib1", Qualified: "lib1"}},
					},
				},
				&model.Library{
					// Unnamed Library
					Includes: []*model.Include{
						{Identifier: &model.LibraryIdentifier{Local: "lib1", Qualified: "lib1"}},
						{Identifier: &model.LibraryIdentifier{Local: "lib2", Qualified: "lib2"}},
					},
				},
			},
		},
		{
			name: "Deps declared in unsorted",
			cqlLibs: []string{
				dedent.Dedent(`
				library lib3
				include lib2
				include lib1`),
				dedent.Dedent(`
				library lib2
				include lib1`),
				"include lib3",
				"library lib1",
			},
			want: []*model.Library{
				&model.Library{
					Identifier: &model.LibraryIdentifier{Local: "lib1", Qualified: "lib1"},
				},
				&model.Library{
					Identifier: &model.LibraryIdentifier{Local: "lib2", Qualified: "lib2"},
					Includes: []*model.Include{
						{Identifier: &model.LibraryIdentifier{Local: "lib1", Qualified: "lib1"}},
					},
				},
				&model.Library{
					Identifier: &model.LibraryIdentifier{Local: "lib3", Qualified: "lib3"},
					Includes: []*model.Include{
						{Identifier: &model.LibraryIdentifier{Local: "lib2", Qualified: "lib2"}},
						{Identifier: &model.LibraryIdentifier{Local: "lib1", Qualified: "lib1"}},
					},
				},
				&model.Library{
					// Unnamed Library
					Includes: []*model.Include{
						{Identifier: &model.LibraryIdentifier{Local: "lib3", Qualified: "lib3"}},
					},
				},
			},
		},
		{
			name: "Two unnamed libraries do not clash",
			cqlLibs: []string{
				"library lib1 version '1.0'",
				"include lib1 version '1.0'",
				"include lib1 version '1.0'",
			},
			want: []*model.Library{
				&model.Library{
					Identifier: &model.LibraryIdentifier{Local: "lib1", Qualified: "lib1", Version: "1.0"},
				},
				&model.Library{
					// Unnamed Library
					Includes: []*model.Include{
						{Identifier: &model.LibraryIdentifier{Local: "lib1", Qualified: "lib1", Version: "1.0"}},
					},
				},
				&model.Library{
					// Unnamed Library
					Includes: []*model.Include{
						{Identifier: &model.LibraryIdentifier{Local: "lib1", Qualified: "lib1", Version: "1.0"}},
					},
				},
			},
		},
		{
			name: "Can include a versioned library without specifying the version",
			cqlLibs: []string{
				"library lib1 version '1.0'",
				"include lib1",
			},
			want: []*model.Library{
				&model.Library{
					Identifier: &model.LibraryIdentifier{Local: "lib1", Qualified: "lib1", Version: "1.0"},
				},
				&model.Library{
					// Unnamed Library
					Includes: []*model.Include{
						{Identifier: &model.LibraryIdentifier{Local: "lib1", Qualified: "lib1", Version: "1.0"}},
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := newFHIRParser(t).Libraries(context.Background(), test.cqlLibs, Config{})
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("Parsing diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParserTopologicalSortMultipleTopLevelLibraries(t *testing.T) {
	tests := []struct {
		name             string
		cqlLibs          []string
		wantTopLevelLibs []model.LibraryIdentifier
	}{
		{
			name: "Simple case with multiple top level libraries",
			cqlLibs: []string{
				dedent.Dedent(`
					library lib3
					include lib2
					include lib1`),
				dedent.Dedent(`
					library measure1
					include lib1
					include lib2
					include lib3`),
				dedent.Dedent(`
					library measure2
					include lib3`),
				"library lib2",
				"library lib1",
			},
			wantTopLevelLibs: []model.LibraryIdentifier{
				{
					Qualified: "measure1",
					Local:     "measure1",
					Version:   "",
				},
				{
					Qualified: "measure2",
					Local:     "measure2",
					Version:   "",
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parsedLibs, err := newFHIRParser(t).Libraries(context.Background(), test.cqlLibs, Config{})
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}

			lastTwoIDs := []model.LibraryIdentifier{
				*parsedLibs[len(parsedLibs)-1].Identifier,
				*parsedLibs[len(parsedLibs)-2].Identifier,
			}
			slices.SortStableFunc(lastTwoIDs, func(a, b model.LibraryIdentifier) int {
				if a.Qualified != b.Qualified {
					return strings.Compare(a.Qualified, b.Qualified)
				}
				return strings.Compare(a.Version, b.Version)
			})
			// Topological sort isn't 100% deterministic so assert the last values are what we want.
			if diff := cmp.Diff(test.wantTopLevelLibs, lastTwoIDs); diff != "" {
				t.Errorf("%v\nLibraries(%v) parsing diff (-want +got):\n%v", test.wantTopLevelLibs, lastTwoIDs, diff)
			}
		})
	}
}

func TestParserTopologicalSortLibrariesIncludeCorrectVersion(t *testing.T) {
	tests := []struct {
		name             string
		cqlLibs          []string
		wantTopLevelLibs []*model.Library
	}{
		{
			name: "Can parse includes with no version specified and multiple versions for the same library",
			cqlLibs: []string{
				"library lib1 version '1.0'",
				"library lib1 version '1.2'",
				dedent.Dedent(`
					library measure1
					include lib1`),
			},
			wantTopLevelLibs: []*model.Library{
				&model.Library{
					Identifier: &model.LibraryIdentifier{
						Qualified: "lib1",
						Local:     "lib1",
						Version:   "1.0",
					},
				},
				&model.Library{
					Identifier: &model.LibraryIdentifier{
						Qualified: "lib1",
						Local:     "lib1",
						Version:   "1.2",
					},
				},
				&model.Library{
					Identifier: &model.LibraryIdentifier{
						Local:     "measure1",
						Qualified: "measure1",
						Version:   "",
					},
					Includes: []*model.Include{
						{
							Identifier: &model.LibraryIdentifier{
								Local:     "lib1",
								Qualified: "lib1",
								Version:   "1.2",
							},
						},
					},
				},
			},
		},
		{
			name: "Can parse includes with no version specified and multiple versions for the same library where one has no version.",
			cqlLibs: []string{
				"library lib1 version '1.0'",
				"library lib1",
				dedent.Dedent(`
					library measure1
					include lib1`),
			},
			wantTopLevelLibs: []*model.Library{
				&model.Library{
					Identifier: &model.LibraryIdentifier{
						Qualified: "lib1",
						Local:     "lib1",
					},
				},
				&model.Library{
					Identifier: &model.LibraryIdentifier{
						Qualified: "lib1",
						Local:     "lib1",
						Version:   "1.0",
					},
				},
				&model.Library{
					Identifier: &model.LibraryIdentifier{
						Local:     "measure1",
						Qualified: "measure1",
						Version:   "",
					},
					Includes: []*model.Include{
						{
							Identifier: &model.LibraryIdentifier{
								Local:     "lib1",
								Qualified: "lib1",
							},
						},
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parsedLibs, err := newFHIRParser(t).Libraries(context.Background(), test.cqlLibs, Config{})
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}

			// Topological sort isn't 100% deterministic so we need to sort the libraries.
			slices.SortStableFunc(parsedLibs, func(a, b *model.Library) int {
				if a.Identifier.Qualified != b.Identifier.Qualified {
					return strings.Compare(a.Identifier.Qualified, b.Identifier.Qualified)
				}
				return strings.Compare(a.Identifier.Version, b.Identifier.Version)
			})
			if diff := cmp.Diff(test.wantTopLevelLibs, parsedLibs); diff != "" {
				t.Errorf("%v\nLibraries(%v) parsing diff (-want +got):\n%v", test.wantTopLevelLibs, parsedLibs, diff)
			}
		})
	}
}

func TestParserSingleLibrary(t *testing.T) {
	tests := []struct {
		name string
		desc string
		cql  string
		want *model.Library
	}{
		{
			name: "LibraryDef with version",
			cql:  "library Example.TrivialTest version '1.2.3'",
			want: &model.Library{
				Identifier: &model.LibraryIdentifier{
					Local:     "TrivialTest",
					Qualified: "Example.TrivialTest",
					Version:   "1.2.3",
				},
			},
		},
		{
			name: "LibraryDef without version",
			cql:  "library TrivialTest",
			want: &model.Library{
				Identifier: &model.LibraryIdentifier{
					Local:     "TrivialTest",
					Qualified: "TrivialTest",
				},
			},
		},
		{
			name: "Context",
			cql: dedent.Dedent(`
        using FHIR version '4.0.1'
				context Patient`),
			want: &model.Library{
				Usings: []*model.Using{
					&model.Using{
						LocalIdentifier: "FHIR",
						Version:         "4.0.1",
						URI:             "http://hl7.org/fhir",
					},
				},
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:        "Patient",
							Context:     "Patient",
							AccessLevel: "PRIVATE",
							Expression: &model.SingletonFrom{
								UnaryExpression: &model.UnaryExpression{
									Operand: &model.Retrieve{
										DataType:   "{http://hl7.org/fhir}Patient",
										TemplateID: "http://hl7.org/fhir/StructureDefinition/Patient",
										Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Patient"}}),
									},
									Expression: model.ResultType(&types.Named{TypeName: "FHIR.Patient"}),
								},
							},
							Element: &model.Element{ResultType: &types.Named{TypeName: "FHIR.Patient"}},
						},
					},
				},
			},
		},
		{
			name: "Retrieve no filter",
			cql: dedent.Dedent(`
        using FHIR version '4.0.1'
        define TestPatient: ["Patient"]`),
			want: &model.Library{
				Usings: []*model.Using{
					&model.Using{
						LocalIdentifier: "FHIR",
						Version:         "4.0.1",
						URI:             "http://hl7.org/fhir",
					},
				},
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:        "TestPatient",
							Context:     "Patient",
							AccessLevel: "PUBLIC",
							Expression: &model.Retrieve{
								DataType:   "{http://hl7.org/fhir}Patient",
								TemplateID: "http://hl7.org/fhir/StructureDefinition/Patient",
								Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Patient"}}),
							},
							Element: &model.Element{ResultType: &types.List{ElementType: &types.Named{TypeName: "FHIR.Patient"}}},
						},
					},
				},
			},
		},
		{
			name: "ValuesetDef",
			cql: dedent.Dedent(`
				codesystem cs: 'https://example.com/codesystem'
				valueset "Diabetes": 'https://example.com/diabetes_value_set'
				valueset "Versioned Diabetes": 'https://example.com/diabetes_value_set' version '1.0.0'
				valueset "CodeSystems Diabetes": 'https://example.com/diabetes_value_set' codesystems { cs }
				public valueset "Public Diabetes": 'https://example.com/diabetes_value_set'
				private valueset "Private Diabetes": 'https://example.com/diabetes_value_set'`),
			want: &model.Library{
				CodeSystems: []*model.CodeSystemDef{
					&model.CodeSystemDef{
						Name:        "cs",
						ID:          "https://example.com/codesystem",
						AccessLevel: "PUBLIC",
						Element:     &model.Element{ResultType: types.CodeSystem},
					},
				},
				Valuesets: []*model.ValuesetDef{
					&model.ValuesetDef{
						Name:        "Diabetes",
						ID:          "https://example.com/diabetes_value_set",
						Element:     &model.Element{ResultType: types.ValueSet},
						AccessLevel: "PUBLIC",
					},
					&model.ValuesetDef{
						Name:        "Versioned Diabetes",
						ID:          "https://example.com/diabetes_value_set",
						Version:     "1.0.0",
						Element:     &model.Element{ResultType: types.ValueSet},
						AccessLevel: "PUBLIC",
					},
					&model.ValuesetDef{
						Name: "CodeSystems Diabetes",
						ID:   "https://example.com/diabetes_value_set",
						CodeSystems: []*model.CodeSystemRef{
							&model.CodeSystemRef{
								Name:       "cs",
								Expression: model.ResultType(types.CodeSystem),
							},
						},
						Element:     &model.Element{ResultType: types.ValueSet},
						AccessLevel: "PUBLIC",
					},
					&model.ValuesetDef{
						Name:        "Public Diabetes",
						ID:          "https://example.com/diabetes_value_set",
						Element:     &model.Element{ResultType: types.ValueSet},
						AccessLevel: "PUBLIC",
					},
					&model.ValuesetDef{
						Name:        "Private Diabetes",
						ID:          "https://example.com/diabetes_value_set",
						Element:     &model.Element{ResultType: types.ValueSet},
						AccessLevel: "PRIVATE",
					},
				},
			},
		},
		{
			name: "CodeSystemDef",
			cql: dedent.Dedent(`
				codesystem cs: 'https://example.com/codesystem'
				codesystem cs2: 'https://example.com/codesystem_2'
				codesystem cs_with_version: 'https://example.com/codesystem_versioned' version '1.0'
				private codesystem cs_private: 'https://example.com/codesystem_private'
				`),
			want: &model.Library{
				CodeSystems: []*model.CodeSystemDef{
					&model.CodeSystemDef{
						Name:        "cs",
						ID:          "https://example.com/codesystem",
						AccessLevel: "PUBLIC",
						Element:     &model.Element{ResultType: types.CodeSystem},
					},
					&model.CodeSystemDef{
						Name:        "cs2",
						ID:          "https://example.com/codesystem_2",
						AccessLevel: "PUBLIC",
						Element:     &model.Element{ResultType: types.CodeSystem},
					},
					&model.CodeSystemDef{
						Name:        "cs_with_version",
						ID:          "https://example.com/codesystem_versioned",
						Version:     "1.0",
						AccessLevel: "PUBLIC",
						Element:     &model.Element{ResultType: types.CodeSystem},
					},
					&model.CodeSystemDef{
						Name:        "cs_private",
						ID:          "https://example.com/codesystem_private",
						AccessLevel: "PRIVATE",
						Element:     &model.Element{ResultType: types.CodeSystem},
					},
				},
			},
		},
		{
			name: "CodeDef",
			cql: dedent.Dedent(`
				codesystem cs_with_version: 'https://example.com/codesystem_versioned' version '1.0'
				code c: '1234' from cs_with_version
				code c_with_display: '12345' from cs_with_version display 'Super Special Display'
				`),
			want: &model.Library{
				CodeSystems: []*model.CodeSystemDef{
					&model.CodeSystemDef{
						Name:        "cs_with_version",
						ID:          "https://example.com/codesystem_versioned",
						Version:     "1.0",
						AccessLevel: "PUBLIC",
						Element:     &model.Element{ResultType: types.CodeSystem},
					},
				},
				Codes: []*model.CodeDef{
					&model.CodeDef{
						Name: "c",
						Code: "1234",
						CodeSystem: &model.CodeSystemRef{
							Name:       "cs_with_version",
							Expression: model.ResultType(types.CodeSystem),
						},
						AccessLevel: "PUBLIC",
						Element:     &model.Element{ResultType: types.Code},
					},
					&model.CodeDef{
						Name: "c_with_display",
						Code: "12345",
						CodeSystem: &model.CodeSystemRef{
							Name:       "cs_with_version",
							Expression: model.ResultType(types.CodeSystem),
						},
						Display:     "Super Special Display",
						AccessLevel: "PUBLIC",
						Element:     &model.Element{ResultType: types.Code},
					},
				},
			},
		},
		{
			name: "ConceptDef",
			cql: dedent.Dedent(`
			codesystem cs: 'https://example.com/codesystem'
			code c: '1234' from cs
			code c2: '456' from cs
			concept con: { c, c2 }
			concept con_with_display: { c } display 'A medical condition'
			private concept pvt_con: { c }
			`),
			want: &model.Library{
				CodeSystems: []*model.CodeSystemDef{
					&model.CodeSystemDef{
						Name:        "cs",
						ID:          "https://example.com/codesystem",
						AccessLevel: "PUBLIC",
						Element:     &model.Element{ResultType: types.CodeSystem},
					},
				},
				Codes: []*model.CodeDef{
					&model.CodeDef{
						Name: "c",
						Code: "1234",
						CodeSystem: &model.CodeSystemRef{
							Name:       "cs",
							Expression: model.ResultType(types.CodeSystem),
						},
						AccessLevel: "PUBLIC",
						Element:     &model.Element{ResultType: types.Code},
					},
					&model.CodeDef{
						Name: "c2",
						Code: "456",
						CodeSystem: &model.CodeSystemRef{
							Name:       "cs",
							Expression: model.ResultType(types.CodeSystem),
						},
						AccessLevel: "PUBLIC",
						Element:     &model.Element{ResultType: types.Code},
					},
				},
				Concepts: []*model.ConceptDef{
					&model.ConceptDef{
						Name: "con",
						Codes: []*model.CodeRef{
							&model.CodeRef{
								Name:       "c",
								Expression: model.ResultType(types.Code),
							},
							&model.CodeRef{
								Name:       "c2",
								Expression: model.ResultType(types.Code),
							},
						},
						AccessLevel: "PUBLIC",
						Element:     &model.Element{ResultType: types.Concept},
					},
					&model.ConceptDef{
						Name: "con_with_display",
						Codes: []*model.CodeRef{
							&model.CodeRef{
								Name:       "c",
								Expression: model.ResultType(types.Code),
							},
						},
						Display:     "A medical condition",
						AccessLevel: "PUBLIC",
						Element:     &model.Element{ResultType: types.Concept},
					},
					&model.ConceptDef{
						Name: "pvt_con",
						Codes: []*model.CodeRef{
							&model.CodeRef{
								Name:       "c",
								Expression: model.ResultType(types.Code),
							},
						},
						AccessLevel: "PRIVATE",
						Element:     &model.Element{ResultType: types.Concept},
					},
				},
			},
		},
		{
			name: "Code Selector",
			cql: dedent.Dedent(`
				codesystem cs: 'https://example.com/codesystem'
				define Foo: Code '132' from cs display 'Severed Leg'
				`),
			want: &model.Library{
				CodeSystems: []*model.CodeSystemDef{
					&model.CodeSystemDef{
						Name:        "cs",
						ID:          "https://example.com/codesystem",
						AccessLevel: "PUBLIC",
						Element:     &model.Element{ResultType: types.CodeSystem},
					},
				},
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:        "Foo",
							AccessLevel: "PUBLIC",
							Element:     &model.Element{ResultType: types.Code},
							Expression: &model.Code{
								Expression: model.ResultType(types.Code),
								System:     &model.CodeSystemRef{Expression: model.ResultType(types.CodeSystem), Name: "cs"},
								Code:       "132",
								Display:    "Severed Leg",
							},
						},
					},
				},
			},
		},
		{
			name: "ExpressionDef with Access Modifier",
			cql: dedent.Dedent(`
        using FHIR version '4.0.1'
				define public Pub: 8
				define private Priv: 4
				`),
			want: &model.Library{
				Usings: []*model.Using{
					&model.Using{
						LocalIdentifier: "FHIR",
						Version:         "4.0.1",
						URI:             "http://hl7.org/fhir",
					},
				},
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:        "Pub",
							Context:     "Patient",
							AccessLevel: "PUBLIC",
							Expression:  model.NewLiteral("8", types.Integer),
							Element:     &model.Element{ResultType: types.Integer},
						},
						&model.ExpressionDef{
							Name:        "Priv",
							Context:     "Patient",
							AccessLevel: "PRIVATE",
							Expression:  model.NewLiteral("4", types.Integer),
							Element:     &model.Element{ResultType: types.Integer},
						},
					},
				},
			},
		},
		{
			name: "Model Used Twice Does Not Overwrite",
			cql: dedent.Dedent(`
        using FHIR version '4.0.1'
				define Res1: First({3})
				define Res2: First({4})
				`),
			want: &model.Library{
				Usings: []*model.Using{
					&model.Using{
						LocalIdentifier: "FHIR",
						Version:         "4.0.1",
						URI:             "http://hl7.org/fhir",
					},
				},
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:        "Res1",
							Context:     "Patient",
							AccessLevel: model.Public,
							Expression: &model.First{
								UnaryExpression: &model.UnaryExpression{
									Operand: &model.List{
										Expression: model.ResultType(&types.List{ElementType: types.Integer}),
										List: []model.IExpression{
											model.NewLiteral("3", types.Integer),
										},
									},
									Expression: model.ResultType(types.Integer),
								},
							},
							Element: &model.Element{ResultType: types.Integer},
						},
						&model.ExpressionDef{
							Name:        "Res2",
							Context:     "Patient",
							AccessLevel: model.Public,
							Expression: &model.First{
								UnaryExpression: &model.UnaryExpression{
									Operand: &model.List{
										Expression: model.ResultType(&types.List{ElementType: types.Integer}),
										List: []model.IExpression{
											model.NewLiteral("4", types.Integer),
										},
									},
									Expression: model.ResultType(types.Integer),
								},
							},
							Element: &model.Element{ResultType: types.Integer},
						},
					},
				},
			},
		},
		{
			name: "ParameterDefinition",
			cql: dedent.Dedent(`
			parameter "Defined Type" Integer
			parameter "Default And Inferred Type" default Interval[@2023-04-01T00:00:00.0, @2024-03-31T00:00:00.0)
			public parameter "Public" Integer
			private parameter "Private" Integer
				`),
			want: &model.Library{
				Identifier: nil,
				Usings:     nil,
				Parameters: []*model.ParameterDef{
					&model.ParameterDef{
						Name:        "Defined Type",
						AccessLevel: "PUBLIC",
						Element:     &model.Element{ResultType: types.Integer},
					},
					&model.ParameterDef{
						Name:        "Default And Inferred Type",
						AccessLevel: "PUBLIC",
						Default: &model.Interval{
							Low:           model.NewLiteral("@2023-04-01T00:00:00.0", types.DateTime),
							High:          model.NewLiteral("@2024-03-31T00:00:00.0", types.DateTime),
							Expression:    model.ResultType(&types.Interval{PointType: types.DateTime}),
							LowInclusive:  true,
							HighInclusive: false,
						},
						Element: &model.Element{ResultType: &types.Interval{PointType: types.DateTime}},
					},
					&model.ParameterDef{
						Name:        "Public",
						AccessLevel: "PUBLIC",
						Element:     &model.Element{ResultType: types.Integer},
					},
					&model.ParameterDef{
						Name:        "Private",
						AccessLevel: "PRIVATE",
						Element:     &model.Element{ResultType: types.Integer},
					},
				},
				Statements: nil,
			},
		},
		{
			name: "KeywordIdentifier Reference",
			cql: dedent.Dedent(`
        define "descending": 32
				define population: descending`),
			want: &model.Library{
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:        "descending",
							AccessLevel: "PUBLIC",
							Expression:  model.NewLiteral("32", types.Integer),
							Element:     &model.Element{ResultType: types.Integer},
						},
						&model.ExpressionDef{
							Name:        "population",
							AccessLevel: "PUBLIC",
							Expression:  &model.ExpressionRef{Name: "descending", Expression: model.ResultType(types.Integer)},
							Element:     &model.Element{ResultType: types.Integer},
						},
					},
				},
			},
		},
		{
			name: "ExpressionReference unquoted references quoted",
			cql: dedent.Dedent(`
        define "age": 32
				define population: age`),
			want: &model.Library{
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:        "age",
							AccessLevel: "PUBLIC",
							Expression:  model.NewLiteral("32", types.Integer),
							Element:     &model.Element{ResultType: types.Integer},
						},
						&model.ExpressionDef{
							Name:        "population",
							AccessLevel: "PUBLIC",
							Expression:  &model.ExpressionRef{Name: "age", Expression: model.ResultType(types.Integer)},
							Element:     &model.Element{ResultType: types.Integer},
						},
					},
				},
			},
		},
		{
			name: "InvocationExpressionTerm local reference",
			cql: dedent.Dedent(`
        using FHIR version '4.0.1'

        define X: [Observation]
				define Y: X`),
			want: &model.Library{
				Usings: []*model.Using{
					&model.Using{
						LocalIdentifier: "FHIR",
						Version:         "4.0.1",
						URI:             "http://hl7.org/fhir",
					},
				},
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:        "X",
							Context:     "Patient",
							AccessLevel: "PUBLIC",
							Expression: &model.Retrieve{
								DataType:     "{http://hl7.org/fhir}Observation",
								TemplateID:   "http://hl7.org/fhir/StructureDefinition/Observation",
								CodeProperty: "code",
								Expression:   model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}}),
							},
							Element: &model.Element{ResultType: &types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}}},
						},
						&model.ExpressionDef{
							Name:        "Y",
							Context:     "Patient",
							AccessLevel: "PUBLIC",
							Expression:  &model.ExpressionRef{Name: "X", Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}})},
							Element:     &model.Element{ResultType: &types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}}},
						},
					},
				},
			},
		},
		{
			name: "InvocationExpressionTerm local reference with property",
			desc: "Also tests nested properties.",
			cql: dedent.Dedent(`
        using FHIR version '4.0.1'

        define X: [Observation]
				define Y: X.category.coding`),
			want: &model.Library{
				Usings: []*model.Using{
					&model.Using{
						LocalIdentifier: "FHIR",
						Version:         "4.0.1",
						URI:             "http://hl7.org/fhir",
					},
				},
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:        "X",
							Context:     "Patient",
							AccessLevel: "PUBLIC",
							Expression: &model.Retrieve{
								DataType:     "{http://hl7.org/fhir}Observation",
								TemplateID:   "http://hl7.org/fhir/StructureDefinition/Observation",
								CodeProperty: "code",
								Expression:   model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}}),
							},
							Element: &model.Element{ResultType: &types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}}},
						},
						&model.ExpressionDef{
							Name:        "Y",
							Context:     "Patient",
							AccessLevel: "PUBLIC",
							Expression: &model.Property{
								Source: &model.Property{
									Source:     &model.ExpressionRef{Name: "X", Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}})},
									Path:       "category",
									Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.CodeableConcept"}}),
								},
								Path:       "coding",
								Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Coding"}}),
							},
							Element: &model.Element{ResultType: &types.List{ElementType: &types.Named{TypeName: "FHIR.Coding"}}},
						},
					},
				},
			},
		},
		{
			name: "QuerySource holds expression",
			desc: "() changes grammar to capture InvocationExpressionTerm not QualifiedIdentifierExpression",
			cql: dedent.Dedent(`
			using FHIR version '4.0.1'

			define X: [Observation]
			define Y: from (X.status) O`),
			want: &model.Library{
				Usings: []*model.Using{
					&model.Using{
						LocalIdentifier: "FHIR",
						Version:         "4.0.1",
						URI:             "http://hl7.org/fhir",
					},
				},
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:        "X",
							Context:     "Patient",
							AccessLevel: "PUBLIC",
							Expression: &model.Retrieve{
								DataType:     "{http://hl7.org/fhir}Observation",
								TemplateID:   "http://hl7.org/fhir/StructureDefinition/Observation",
								CodeProperty: "code",
								Expression:   model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}}),
							},
							Element: &model.Element{ResultType: &types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}}},
						},
						&model.ExpressionDef{
							Name:        "Y",
							Context:     "Patient",
							AccessLevel: "PUBLIC",
							Expression: &model.Query{
								Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.ObservationStatus"}}),
								Source: []*model.AliasedSource{
									&model.AliasedSource{
										Alias: "O",
										Source: &model.Property{
											Source:     &model.ExpressionRef{Name: "X", Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}})},
											Path:       "status",
											Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.ObservationStatus"}}),
										},
										Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.ObservationStatus"}}),
									},
								},
							},
							Element: &model.Element{ResultType: &types.List{ElementType: &types.Named{TypeName: "FHIR.ObservationStatus"}}},
						},
					},
				},
			},
		},
		{
			name: "QualifiedIdentifierExpression local reference",
			cql: dedent.Dedent(`
			using FHIR version '4.0.1'

			define X: [Observation]
			define Y: from X O`),
			want: &model.Library{
				Usings: []*model.Using{
					&model.Using{
						LocalIdentifier: "FHIR",
						Version:         "4.0.1",
						URI:             "http://hl7.org/fhir",
					},
				},
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:        "X",
							Context:     "Patient",
							AccessLevel: "PUBLIC",
							Expression: &model.Retrieve{
								DataType:     "{http://hl7.org/fhir}Observation",
								TemplateID:   "http://hl7.org/fhir/StructureDefinition/Observation",
								CodeProperty: "code",
								Expression:   model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}}),
							},
							Element: &model.Element{ResultType: &types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}}},
						},
						&model.ExpressionDef{
							Name:        "Y",
							Context:     "Patient",
							AccessLevel: "PUBLIC",
							Expression: &model.Query{
								Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}}),
								Source: []*model.AliasedSource{
									&model.AliasedSource{
										Alias:      "O",
										Source:     &model.ExpressionRef{Name: "X", Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}})},
										Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}}),
									},
								},
							},
							Element: &model.Element{ResultType: &types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}}},
						},
					},
				},
			},
		},
		{
			name: "QualifiedIdentifierExpression local reference with multiple properties",
			cql: dedent.Dedent(`
			using FHIR version '4.0.1'
			define X: [Condition] O return O.code.coding.display`),
			want: &model.Library{
				Usings: []*model.Using{
					&model.Using{
						LocalIdentifier: "FHIR",
						Version:         "4.0.1",
						URI:             "http://hl7.org/fhir",
					},
				},
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:        "X",
							Context:     "Patient",
							AccessLevel: "PUBLIC",
							Expression: &model.Query{
								Expression: model.ResultType(&types.List{ElementType: &types.List{ElementType: &types.Named{TypeName: "FHIR.string"}}}),
								Source: []*model.AliasedSource{
									&model.AliasedSource{
										Alias: "O",
										Source: &model.Retrieve{
											DataType:     "{http://hl7.org/fhir}Condition",
											TemplateID:   "http://hl7.org/fhir/StructureDefinition/Condition",
											CodeProperty: "code",
											Expression:   model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Condition"}}),
										},
										Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Condition"}}),
									},
								},
								Return: &model.ReturnClause{
									Distinct: true,
									Element:  &model.Element{ResultType: &types.List{ElementType: &types.Named{TypeName: "FHIR.string"}}},
									Expression: &model.Property{
										Source: &model.Property{
											Source: &model.Property{
												Source: &model.AliasRef{
													Name:       "O",
													Expression: model.ResultType(&types.Named{TypeName: "FHIR.Condition"}),
												},
												Path:       "code",
												Expression: model.ResultType(&types.Named{TypeName: "FHIR.CodeableConcept"}),
											},
											Path:       "coding",
											Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Coding"}}),
										},
										Path:       "display",
										Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.string"}}),
									},
								},
							},
							Element: &model.Element{ResultType: &types.List{ElementType: &types.List{ElementType: &types.Named{TypeName: "FHIR.string"}}}},
						},
					},
				},
			},
		},
		{
			name: "Null",
			cql: dedent.Dedent(`
        using FHIR version '4.0.1'

        define nullValue: null
				define nullFunction: exists("nullValue")
				define nullEqual: "nullValue" = null
				`),
			want: &model.Library{
				Usings: []*model.Using{
					&model.Using{
						LocalIdentifier: "FHIR",
						Version:         "4.0.1",
						URI:             "http://hl7.org/fhir",
					},
				},
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:        "nullValue",
							Context:     "Patient",
							AccessLevel: "PUBLIC",
							Expression:  model.NewLiteral("null", types.Any),
							Element:     &model.Element{ResultType: types.Any}},
						&model.ExpressionDef{
							Name:        "nullFunction",
							Context:     "Patient",
							AccessLevel: "PUBLIC",
							Expression: &model.Exists{
								UnaryExpression: &model.UnaryExpression{
									Operand: &model.As{
										UnaryExpression: &model.UnaryExpression{
											Expression: model.ResultType(&types.List{ElementType: types.Any}),
											Operand:    &model.ExpressionRef{Expression: model.ResultType(types.Any), Name: "nullValue"},
										},
										AsTypeSpecifier: &types.List{ElementType: types.Any},
									},
									Expression: model.ResultType(types.Boolean)},
							},
							Element: &model.Element{ResultType: types.Boolean},
						},
						&model.ExpressionDef{
							Name:        "nullEqual",
							Context:     "Patient",
							AccessLevel: "PUBLIC",
							Expression: &model.Equal{
								BinaryExpression: &model.BinaryExpression{
									Operands: []model.IExpression{
										&model.ExpressionRef{Name: "nullValue", Expression: model.ResultType(types.Any)},
										model.NewLiteral("null", types.Any)},
									Expression: model.ResultType(types.Boolean),
								},
							},
							Element: &model.Element{ResultType: types.Boolean},
						},
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parsedLibs, err := newFHIRParser(t).Libraries(context.Background(), []string{test.cql}, Config{})
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(test.want, parsedLibs[0]); diff != "" {
				t.Errorf("%v\nLibraries(%s) parsing diff (-want +got):\n%s", test.desc, test.cql, diff)
			}
		})
	}
}

func TestParserMultipleLibraries(t *testing.T) {
	tests := []struct {
		name string
		cql  string
		want *model.Library
	}{
		{
			name: "IncludeDef without called",
			cql:  `include example.helpers1 version '1.0'`,
			want: &model.Library{
				Includes: []*model.Include{
					{
						Identifier: &model.LibraryIdentifier{
							Local:     "helpers1",
							Qualified: "example.helpers1",
							Version:   "1.0",
						},
					},
				},
			},
		},
		{
			name: "IncludeDef without version",
			cql:  `include example.helpers2 called Helpers`,
			want: &model.Library{
				Includes: []*model.Include{
					{
						Identifier: &model.LibraryIdentifier{
							Local:     "Helpers",
							Qualified: "example.helpers2",
						},
					},
				},
			},
		},
		{
			name: "IncludeDef with called",
			cql:  `include example.helpers1 version '1.0' called Helpers`,
			want: &model.Library{
				Includes: []*model.Include{
					{
						Identifier: &model.LibraryIdentifier{
							Local:     "Helpers",
							Qualified: "example.helpers1",
							Version:   "1.0",
						},
					},
				},
			},
		},
		{
			name: "IncludeDef multiple includes",
			cql: dedent.Dedent(`
        include example.helpers1 version '1.0' called Helpers
				include example.helpers2 called Helpers2`),
			want: &model.Library{
				Includes: []*model.Include{
					{
						Identifier: &model.LibraryIdentifier{
							Local:     "Helpers",
							Qualified: "example.helpers1",
							Version:   "1.0",
						},
					},
					{
						Identifier: &model.LibraryIdentifier{
							Local:     "Helpers2",
							Qualified: "example.helpers2",
						},
					},
				},
			},
		},
		{
			name: "Global Reference Parameter",
			cql: dedent.Dedent(`
        include example.helpers1 version '1.0' called Helpers
				define X: Helpers."public param"`),
			want: &model.Library{
				Includes: []*model.Include{
					{
						Identifier: &model.LibraryIdentifier{
							Local:     "Helpers",
							Qualified: "example.helpers1",
							Version:   "1.0",
						},
					},
				},
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:        "X",
							AccessLevel: "PUBLIC",
							Expression:  &model.ParameterRef{Name: "public param", LibraryName: "Helpers", Expression: model.ResultType(types.Integer)},
							Element:     &model.Element{ResultType: types.Integer},
						},
					},
				},
			},
		},
		{
			name: "Global Reference CodeSystem",
			cql: dedent.Dedent(`
        include example.helpers1 version '1.0' called Helpers
				define X: Helpers."public codesystem"`),
			want: &model.Library{
				Includes: []*model.Include{
					{
						Identifier: &model.LibraryIdentifier{
							Local:     "Helpers",
							Qualified: "example.helpers1",
							Version:   "1.0",
						},
					},
				},
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:        "X",
							AccessLevel: "PUBLIC",
							Element:     &model.Element{ResultType: types.CodeSystem},
							Expression:  &model.CodeSystemRef{Name: "public codesystem", LibraryName: "Helpers", Expression: model.ResultType(types.CodeSystem)},
						},
					},
				},
			},
		},
		{
			name: "Global Reference ValueSet",
			cql: dedent.Dedent(`
        include example.helpers1 version '1.0' called Helpers
				define X: Helpers."public valueset"`),
			want: &model.Library{
				Includes: []*model.Include{
					{
						Identifier: &model.LibraryIdentifier{
							Local:     "Helpers",
							Qualified: "example.helpers1",
							Version:   "1.0",
						},
					},
				},
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:        "X",
							AccessLevel: "PUBLIC",
							Element:     &model.Element{ResultType: types.ValueSet},
							Expression:  &model.ValuesetRef{Name: "public valueset", LibraryName: "Helpers", Expression: model.ResultType(types.ValueSet)},
						},
					},
				},
			},
		},
		{
			name: "InvocationExpressionTerm Global Reference",
			cql: dedent.Dedent(`
        include example.helpers1 version '1.0' called Helpers
				define X: Helpers."public def"`),
			want: &model.Library{
				Includes: []*model.Include{
					{
						Identifier: &model.LibraryIdentifier{
							Local:     "Helpers",
							Qualified: "example.helpers1",
							Version:   "1.0",
						},
					},
				},
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:        "X",
							AccessLevel: "PUBLIC",
							Expression:  &model.ExpressionRef{Name: "public def", LibraryName: "Helpers", Expression: model.ResultType(types.Integer)},
							Element:     &model.Element{ResultType: types.Integer},
						},
					},
				},
			},
		},
		{
			name: "InvocationExpressionTerm Global Reference with Property",
			cql: dedent.Dedent(`
        include example.helpers1 version '1.0' called Helpers
				define X: Helpers."fhir def".status`),
			want: &model.Library{
				Includes: []*model.Include{
					{
						Identifier: &model.LibraryIdentifier{
							Local:     "Helpers",
							Qualified: "example.helpers1",
							Version:   "1.0",
						},
					},
				},
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:        "X",
							AccessLevel: "PUBLIC",
							Expression: &model.Property{
								Source:     &model.ExpressionRef{Name: "fhir def", LibraryName: "Helpers", Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}})},
								Path:       "status",
								Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.ObservationStatus"}}),
							},
							Element: &model.Element{ResultType: &types.List{ElementType: &types.Named{TypeName: "FHIR.ObservationStatus"}}},
						},
					},
				},
			},
		},
		{
			name: "QualifiedIdentifierExpression Global Reference",
			cql: dedent.Dedent(`
        include example.helpers1 version '1.0' called Helpers
				define X: from Helpers."public def" P`),
			want: &model.Library{
				Includes: []*model.Include{
					{
						Identifier: &model.LibraryIdentifier{
							Local:     "Helpers",
							Qualified: "example.helpers1",
							Version:   "1.0",
						},
					},
				},
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:        "X",
							AccessLevel: "PUBLIC",
							Expression: &model.Query{
								Expression: model.ResultType(types.Integer),
								Source: []*model.AliasedSource{
									&model.AliasedSource{
										Alias:      "P",
										Source:     &model.ExpressionRef{Name: "public def", LibraryName: "Helpers", Expression: model.ResultType(types.Integer)},
										Expression: model.ResultType(types.Integer),
									}},
							},
							Element: &model.Element{ResultType: types.Integer},
						},
					},
				},
			},
		},
		{
			name: "QualifiedIdentifierExpression Global Reference with Property",
			cql: dedent.Dedent(`
        include example.helpers1 version '1.0' called Helpers
				define X: from (4) P return Helpers."public interval".high`),
			want: &model.Library{
				Includes: []*model.Include{
					{
						Identifier: &model.LibraryIdentifier{
							Local:     "Helpers",
							Qualified: "example.helpers1",
							Version:   "1.0",
						},
					},
				},
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:        "X",
							AccessLevel: "PUBLIC",
							Expression: &model.Query{
								Expression: model.ResultType(types.Integer),
								Source: []*model.AliasedSource{
									&model.AliasedSource{
										Alias:      "P",
										Source:     model.NewLiteral("4", types.Integer),
										Expression: model.ResultType(types.Integer),
									},
								},
								Return: &model.ReturnClause{
									Distinct: true,
									Expression: &model.Property{
										Source:     &model.ExpressionRef{Name: "public interval", LibraryName: "Helpers", Expression: model.ResultType(&types.Interval{PointType: types.Integer})},
										Path:       "high",
										Expression: model.ResultType(types.Integer),
									},
									Element: &model.Element{ResultType: types.Integer},
								},
							},
							Element: &model.Element{ResultType: types.Integer},
						},
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			libs := []string{
				dedent.Dedent(`
					library example.helpers1 version '1.0'
					using FHIR version '4.0.1'
					parameter "public param" default 1
					codesystem "public codesystem": 'url-codesystem'
					valueset "public valueset": 'url-valueset'
					context Patient
					define public "public def": 2
					define public "public interval": Interval[1, 2]
					define private "private def": 3
					define public "fhir def": [Observation]`),
				dedent.Dedent(`
					library example.helpers2
					define public "public def": 4
					define private "private def": 5 `),
				test.cql,
			}

			parsedLibs, err := newFHIRParser(t).Libraries(context.Background(), libs, Config{})
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}
			var gotLib *model.Library
			for _, l := range parsedLibs {
				if l.Identifier == nil {
					gotLib = l
					break
				}
			}
			if diff := cmp.Diff(test.want, gotLib); diff != "" {
				t.Errorf("Parse diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParameters(t *testing.T) {
	tests := []struct {
		name         string
		passedParams map[result.DefKey]string
		want         map[result.DefKey]model.IExpression
	}{
		{
			name: "Literal Int",
			passedParams: map[result.DefKey]string{
				result.DefKey{Name: "lit", Library: result.LibKey{Name: "Highly.Qualified", Version: "1.0"}}: "4",
			},
			want: map[result.DefKey]model.IExpression{
				result.DefKey{
					Name:    "lit",
					Library: result.LibKey{Name: "Highly.Qualified", Version: "1.0"}}: model.NewLiteral("4", types.Integer)},
		},
		{
			name: "Literal String",
			passedParams: map[result.DefKey]string{
				result.DefKey{Name: "lit", Library: result.LibKey{Name: "Highly.Qualified", Version: "1.0"}}: "'Hello'",
			},
			want: map[result.DefKey]model.IExpression{
				result.DefKey{
					Name:    "lit",
					Library: result.LibKey{Name: "Highly.Qualified", Version: "1.0"}}: model.NewLiteral("Hello", types.String)},
		},
		{
			name: "List",
			passedParams: map[result.DefKey]string{
				result.DefKey{Name: "list", Library: result.LibKey{Name: "Highly.Qualified", Version: "1.0"}}: "{1, 2}",
			},
			want: map[result.DefKey]model.IExpression{
				result.DefKey{Name: "list", Library: result.LibKey{Name: "Highly.Qualified", Version: "1.0"}}: &model.List{
					Expression: model.ResultType(&types.List{ElementType: types.Integer}),
					List: []model.IExpression{
						model.NewLiteral("1", types.Integer),
						model.NewLiteral("2", types.Integer),
					},
				},
			},
		},
		{
			name: "Interval",
			passedParams: map[result.DefKey]string{
				result.DefKey{Name: "interval", Library: result.LibKey{Name: "Highly.Qualified", Version: "1.0"}}: "Interval[@2013-01-01T00:00:00.0, @2014-01-01T00:00:00.0)",
			},
			want: map[result.DefKey]model.IExpression{
				result.DefKey{
					Name:    "interval",
					Library: result.LibKey{Name: "Highly.Qualified", Version: "1.0"}}: &model.Interval{
					Low:           model.NewLiteral("@2013-01-01T00:00:00.0", types.DateTime),
					High:          model.NewLiteral("@2014-01-01T00:00:00.0", types.DateTime),
					Expression:    model.ResultType(&types.Interval{PointType: types.DateTime}),
					LowInclusive:  true,
					HighInclusive: false,
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := newFHIRParser(t).Parameters(context.Background(), tc.passedParams, Config{})
			if err != nil {
				t.Fatalf("Parse Parameters returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Parse diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMalformedParameters(t *testing.T) {
	tests := []struct {
		name         string
		passedParams map[result.DefKey]string
		errContains  string
	}{
		{
			name: "InvocationTerm",
			passedParams: map[result.DefKey]string{
				result.DefKey{Name: "lit", Library: result.LibKey{Name: "Highly.Qualified", Version: "1.0"}}: "invo",
			},
			errContains: "must be a interval",
		},
		{
			name: "Expression Definition",
			passedParams: map[result.DefKey]string{
				result.DefKey{Name: "lit", Library: result.LibKey{Name: "Highly.Qualified", Version: "1.0"}}: "define population: 4",
			},
			errContains: "must be a single",
		},
		{
			name: "Multiple Literals",
			passedParams: map[result.DefKey]string{
				result.DefKey{Name: "lit", Library: result.LibKey{Name: "Highly.Qualified", Version: "1.0"}}: "4 15",
			},
			errContains: "must be a single",
		},
		{
			name: "No Literals",
			passedParams: map[result.DefKey]string{
				result.DefKey{Name: "lit", Library: result.LibKey{Name: "Highly.Qualified", Version: "1.0"}}: "",
			},
			errContains: "mismatched input",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := newFHIRParser(t).Parameters(context.Background(), tc.passedParams, Config{})
			if err == nil {
				t.Fatalf("Parameters() did not fail parsing")
			}

			if !strings.Contains(err.Error(), tc.errContains) {
				t.Errorf("Returned error (%s) did not contain expected string (%s)", err.Error(), tc.errContains)
			}
		})
	}
}

func TestRealisticCQL(t *testing.T) {
	tests := []struct {
		name string
		cql  string
		want *model.Library
	}{
		{
			name: "M1",
			cql: dedent.Dedent(`
			library Milestones version '1.0.0'
			using FHIR version '4.0.1'
			context Patient

			define Gender: Patient.gender
			`),
			want: &model.Library{
				Identifier: &model.LibraryIdentifier{
					Local:     "Milestones",
					Qualified: "Milestones",
					Version:   "1.0.0",
				},
				Usings: []*model.Using{
					&model.Using{
						LocalIdentifier: "FHIR",
						Version:         "4.0.1",
						URI:             "http://hl7.org/fhir",
					},
				},
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:        "Patient",
							Context:     "Patient",
							AccessLevel: "PRIVATE",
							Expression: &model.SingletonFrom{
								UnaryExpression: &model.UnaryExpression{
									Operand: &model.Retrieve{
										DataType:   "{http://hl7.org/fhir}Patient",
										TemplateID: "http://hl7.org/fhir/StructureDefinition/Patient",
										Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Patient"}}),
									},
									Expression: model.ResultType(&types.Named{TypeName: "FHIR.Patient"}),
								},
							},
							Element: &model.Element{ResultType: &types.Named{TypeName: "FHIR.Patient"}},
						},
						&model.ExpressionDef{
							Name:        "Gender",
							Context:     "Patient",
							AccessLevel: "PUBLIC",
							Expression: &model.Property{
								Source:     &model.ExpressionRef{Name: "Patient", Expression: model.ResultType(&types.Named{TypeName: "FHIR.Patient"})},
								Path:       "gender",
								Expression: model.ResultType(&types.Named{TypeName: "FHIR.AdministrativeGender"}),
							},
							Element: &model.Element{ResultType: &types.Named{TypeName: "FHIR.AdministrativeGender"}},
						},
					},
				},
			},
		},
		{
			name: "AgeInYearsAt",
			cql: dedent.Dedent(`
        library TrivialTest version '1.2.3'
        using FHIR version '4.0.1'
				parameter "Measurement Period" Interval<Date>
				context Patient
				define "Initial Population":
          AgeInYearsAt(end of "Measurement Period") in Interval[45, 65]
				`),
			want: &model.Library{
				Identifier: &model.LibraryIdentifier{
					Local:     "TrivialTest",
					Qualified: "TrivialTest",
					Version:   "1.2.3",
				},
				Usings: []*model.Using{
					&model.Using{
						LocalIdentifier: "FHIR",
						Version:         "4.0.1",
						URI:             "http://hl7.org/fhir",
					},
				},
				Parameters: []*model.ParameterDef{
					&model.ParameterDef{
						Name:        "Measurement Period",
						AccessLevel: "PUBLIC",
						Element:     &model.Element{ResultType: &types.Interval{PointType: types.Date}},
					},
				},
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:        "Patient",
							Context:     "Patient",
							AccessLevel: "PRIVATE",
							Expression: &model.SingletonFrom{
								UnaryExpression: &model.UnaryExpression{
									Operand: &model.Retrieve{

										DataType:   "{http://hl7.org/fhir}Patient",
										TemplateID: "http://hl7.org/fhir/StructureDefinition/Patient",
										Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Patient"}}),
									},
									Expression: model.ResultType(&types.Named{TypeName: "FHIR.Patient"}),
								},
							},
							Element: &model.Element{ResultType: &types.Named{TypeName: "FHIR.Patient"}},
						},
						&model.ExpressionDef{
							Name:        "Initial Population",
							Context:     "Patient",
							AccessLevel: "PUBLIC",
							Expression: &model.In{
								BinaryExpression: &model.BinaryExpression{
									Operands: []model.IExpression{
										&model.CalculateAgeAt{
											// AgeInYears system function converted to a calculation that gets the birthdate
											BinaryExpression: &model.BinaryExpression{
												Expression: &model.Expression{
													Element: &model.Element{ResultType: types.Integer},
												},
												Operands: []model.IExpression{
													&model.Property{
														Expression: model.ResultType(types.Date),
														Source: &model.Property{
															Expression: model.ResultType(&types.Named{TypeName: "FHIR.date"}),
															Source: &model.ExpressionRef{
																Expression: model.ResultType(&types.Named{TypeName: "FHIR.Patient"}),
																Name:       "Patient",
															},
															Path: "birthDate",
														},
														Path: "value",
													},
													&model.End{
														UnaryExpression: &model.UnaryExpression{
															Operand:    &model.ParameterRef{Name: "Measurement Period", Expression: model.ResultType(&types.Interval{PointType: types.Date})},
															Expression: model.ResultType(types.Date),
														},
													},
												},
											},
											Precision: model.YEAR,
										},
										&model.Interval{
											Expression:    model.ResultType(&types.Interval{PointType: types.Integer}),
											Low:           model.NewLiteral("45", types.Integer),
											High:          model.NewLiteral("65", types.Integer),
											LowInclusive:  true,
											HighInclusive: true,
										}},
									Expression: model.ResultType(types.Boolean),
								},
							},
							Element: &model.Element{ResultType: types.Boolean},
						},
					},
				},
			},
		},
		{
			name: "M2_HasCoronaryHeartDisease",
			cql: dedent.Dedent(`
        library Measure version '1.2.3'
        using FHIR version '4.0.1'

				valueset "Coronary arteriosclerosis": 'url-valueset-coronary' version '1.0.0'
				valueset "Blood pressure": 'url-valueset-blood-pressure' version '1.0.0'

				parameter "Measurement Period"
					default Interval[@2023-04-01T00:00:00.0, @2024-03-31T00:00:00.0)

				context Patient

				define "Has coronary heart disease":
					exists (
						[Condition: "Coronary arteriosclerosis"] chd
							where chd.onset as FHIR.dateTime before start of "Measurement Period"
					)

				define "Most recent blood pressure reading":
					Last(
						[Observation: "Blood pressure"] bp
							where bp.status in {'final', 'amended', 'corrected'}
							and bp.effective in "Measurement Period"
							sort by effective desc
					)

				define "Most recent blood pressure reading below 150":
					"Most recent blood pressure reading".value < 150

				define "Initial Population":
					AgeInYearsAt(start of "Measurement Period") > 80

				define "Denominator":
					"Initial Population"
					and "Has coronary heart disease"

				define "Numerator":
					"Initial Population"
					and "Denominator"
					and "Most recent blood pressure reading below 150"
				`),
			want: &model.Library{
				Identifier: &model.LibraryIdentifier{
					Local:     "Measure",
					Qualified: "Measure",
					Version:   "1.2.3",
				},
				Usings: []*model.Using{
					&model.Using{
						LocalIdentifier: "FHIR",
						Version:         "4.0.1",
						URI:             "http://hl7.org/fhir",
					},
				},
				Parameters: []*model.ParameterDef{
					&model.ParameterDef{
						Element:     &model.Element{ResultType: &types.Interval{PointType: types.DateTime}},
						Name:        "Measurement Period",
						AccessLevel: "PUBLIC",
						Default: &model.Interval{
							Low:           model.NewLiteral("@2023-04-01T00:00:00.0", types.DateTime),
							High:          model.NewLiteral("@2024-03-31T00:00:00.0", types.DateTime),
							Expression:    model.ResultType(&types.Interval{PointType: types.DateTime}),
							LowInclusive:  true,
							HighInclusive: false,
						},
					},
				},
				Valuesets: []*model.ValuesetDef{
					{
						Name:        "Coronary arteriosclerosis",
						ID:          "url-valueset-coronary",
						Version:     "1.0.0",
						Element:     &model.Element{ResultType: types.ValueSet},
						AccessLevel: "PUBLIC",
					},
					{
						Name:        "Blood pressure",
						ID:          "url-valueset-blood-pressure",
						Version:     "1.0.0",
						Element:     &model.Element{ResultType: types.ValueSet},
						AccessLevel: "PUBLIC",
					},
				},
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:        "Patient",
							Context:     "Patient",
							AccessLevel: "PRIVATE",
							Expression: &model.SingletonFrom{
								UnaryExpression: &model.UnaryExpression{
									Operand: &model.Retrieve{
										DataType:   "{http://hl7.org/fhir}Patient",
										TemplateID: "http://hl7.org/fhir/StructureDefinition/Patient",
										Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Patient"}}),
									},
									Expression: model.ResultType(&types.Named{TypeName: "FHIR.Patient"}),
								},
							},
							Element: &model.Element{ResultType: &types.Named{TypeName: "FHIR.Patient"}},
						},
						&model.ExpressionDef{
							Name:        "Has coronary heart disease",
							Context:     "Patient",
							AccessLevel: "PUBLIC",
							Expression: &model.Exists{
								UnaryExpression: &model.UnaryExpression{
									Operand: &model.Query{
										Source: []*model.AliasedSource{
											&model.AliasedSource{
												Alias: "chd",
												Source: &model.Retrieve{
													DataType:     "{http://hl7.org/fhir}Condition",
													TemplateID:   "http://hl7.org/fhir/StructureDefinition/Condition",
													CodeProperty: "code",
													Codes:        &model.ValuesetRef{Name: "Coronary arteriosclerosis", Expression: model.ResultType(types.ValueSet)},
													Expression:   model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Condition"}}),
												},
												Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Condition"}}),
											},
										},
										Where: &model.Before{
											BinaryExpression: &model.BinaryExpression{
												Operands: []model.IExpression{
													&model.FunctionRef{
														Expression:  model.ResultType(types.DateTime),
														Name:        "ToDateTime",
														LibraryName: "FHIRHelpers",
														Operands: []model.IExpression{
															&model.As{
																AsTypeSpecifier: &types.Named{TypeName: "FHIR.dateTime"},
																UnaryExpression: &model.UnaryExpression{
																	Expression: model.ResultType(&types.Named{TypeName: "FHIR.dateTime"}),
																	Operand: &model.Property{
																		Source: &model.AliasRef{Name: "chd", Expression: model.ResultType(&types.Named{TypeName: "FHIR.Condition"})},
																		Path:   "onset",
																		Expression: model.ResultType(&types.Choice{
																			ChoiceTypes: []types.IType{
																				&types.Named{TypeName: "FHIR.dateTime"},
																				&types.Named{TypeName: "FHIR.Age"},
																				&types.Named{TypeName: "FHIR.Period"},
																				&types.Named{TypeName: "FHIR.Range"},
																				&types.Named{TypeName: "FHIR.string"},
																			},
																		}),
																	},
																},
															},
														},
													},

													&model.Start{
														UnaryExpression: &model.UnaryExpression{
															Operand:    &model.ParameterRef{Name: "Measurement Period", Expression: model.ResultType(&types.Interval{PointType: types.DateTime})},
															Expression: model.ResultType(types.DateTime),
														},
													},
												},
												Expression: model.ResultType(types.Boolean),
											},
										},
										Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Condition"}}),
									},
									Expression: model.ResultType(types.Boolean),
								},
							},
							Element: &model.Element{ResultType: types.Boolean},
						},
						&model.ExpressionDef{
							Name:        "Most recent blood pressure reading",
							Context:     "Patient",
							AccessLevel: "PUBLIC",
							Element:     &model.Element{ResultType: &types.Named{TypeName: "FHIR.Observation"}},
							Expression: &model.Last{
								UnaryExpression: &model.UnaryExpression{
									Expression: model.ResultType(&types.Named{TypeName: "FHIR.Observation"}),
									Operand: &model.Query{
										Source: []*model.AliasedSource{
											&model.AliasedSource{
												Alias: "bp",
												Source: &model.Retrieve{
													DataType:     "{http://hl7.org/fhir}Observation",
													TemplateID:   "http://hl7.org/fhir/StructureDefinition/Observation",
													CodeProperty: "code",
													Codes: &model.ValuesetRef{
														Name:       "Blood pressure",
														Expression: model.ResultType(types.ValueSet),
													},
													Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}}),
												},
												Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}}),
											},
										},
										Where: &model.And{
											BinaryExpression: &model.BinaryExpression{
												Expression: model.ResultType(types.Boolean),
												Operands: []model.IExpression{
													&model.In{
														BinaryExpression: &model.BinaryExpression{
															Operands: []model.IExpression{
																&model.FunctionRef{
																	Expression:  model.ResultType(types.String),
																	Name:        "ToString",
																	LibraryName: "FHIRHelpers",
																	Operands: []model.IExpression{
																		&model.Property{
																			Source:     &model.AliasRef{Name: "bp", Expression: model.ResultType(&types.Named{TypeName: "FHIR.Observation"})},
																			Path:       "status",
																			Expression: model.ResultType(&types.Named{TypeName: "FHIR.ObservationStatus"}),
																		},
																	},
																},
																&model.List{
																	Expression: model.ResultType(&types.List{ElementType: types.String}),
																	List: []model.IExpression{
																		model.NewLiteral("final", types.String),
																		model.NewLiteral("amended", types.String),
																		model.NewLiteral("corrected", types.String),
																	},
																},
															},
															Expression: model.ResultType(types.Boolean),
														},
													},
													&model.In{
														BinaryExpression: &model.BinaryExpression{
															Operands: []model.IExpression{
																&model.FunctionRef{
																	Expression:  model.ResultType(types.DateTime),
																	Name:        "ToDateTime",
																	LibraryName: "FHIRHelpers",
																	Operands: []model.IExpression{
																		&model.As{
																			UnaryExpression: &model.UnaryExpression{
																				Expression: model.ResultType(&types.Named{TypeName: "FHIR.dateTime"}),
																				Operand: &model.Property{
																					Source: &model.AliasRef{Name: "bp", Expression: model.ResultType(&types.Named{TypeName: "FHIR.Observation"})},
																					Path:   "effective",
																					Expression: model.ResultType(&types.Choice{
																						ChoiceTypes: []types.IType{
																							&types.Named{TypeName: "FHIR.dateTime"},
																							&types.Named{TypeName: "FHIR.Period"},
																							&types.Named{TypeName: "FHIR.Timing"},
																							&types.Named{TypeName: "FHIR.instant"},
																						}}),
																				},
																			},
																			AsTypeSpecifier: &types.Named{TypeName: "FHIR.dateTime"},
																		},
																	},
																},
																&model.ParameterRef{Name: "Measurement Period", Expression: model.ResultType(&types.Interval{PointType: types.DateTime})},
															},
															Expression: model.ResultType(types.Boolean),
														},
													},
												},
											},
										},
										Sort: &model.SortClause{
											ByItems: []model.ISortByItem{
												&model.SortByColumn{
													SortByItem: &model.SortByItem{
														Direction: model.DESCENDING,
													},
													Path: "effective",
												},
											},
										},
										Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}}),
									},
								},
							},
						},
						&model.ExpressionDef{
							Name:        "Most recent blood pressure reading below 150",
							Context:     "Patient",
							AccessLevel: "PUBLIC",
							Element:     &model.Element{ResultType: types.Boolean},
							Expression: &model.Less{
								BinaryExpression: &model.BinaryExpression{
									Operands: []model.IExpression{
										&model.FunctionRef{
											Expression:  model.ResultType(types.Integer),
											Name:        "ToInteger",
											LibraryName: "FHIRHelpers",
											Operands: []model.IExpression{
												&model.As{
													UnaryExpression: &model.UnaryExpression{
														Expression: model.ResultType(&types.Named{TypeName: "FHIR.integer"}),
														Operand: &model.Property{
															Source: &model.ExpressionRef{Name: "Most recent blood pressure reading", Expression: model.ResultType(&types.Named{TypeName: "FHIR.Observation"})},
															Path:   "value",
															Expression: model.ResultType(&types.Choice{
																ChoiceTypes: []types.IType{
																	&types.Named{TypeName: "FHIR.Quantity"},
																	&types.Named{TypeName: "FHIR.CodeableConcept"},
																	&types.Named{TypeName: "FHIR.string"},
																	&types.Named{TypeName: "FHIR.boolean"},
																	&types.Named{TypeName: "FHIR.integer"},
																	&types.Named{TypeName: "FHIR.Range"},
																	&types.Named{TypeName: "FHIR.Ratio"},
																	&types.Named{TypeName: "FHIR.SampledData"},
																	&types.Named{TypeName: "FHIR.time"},
																	&types.Named{TypeName: "FHIR.dateTime"},
																	&types.Named{TypeName: "FHIR.Period"},
																},
															}),
														},
													},
													AsTypeSpecifier: &types.Named{TypeName: "FHIR.integer"},
												},
											},
										},
										model.NewLiteral("150", types.Integer),
									},
									Expression: model.ResultType(types.Boolean),
								},
							},
						},
						&model.ExpressionDef{
							Name:        "Initial Population",
							Context:     "Patient",
							AccessLevel: "PUBLIC",
							Element:     &model.Element{ResultType: types.Boolean},
							Expression: &model.Greater{
								BinaryExpression: &model.BinaryExpression{
									Operands: []model.IExpression{

										&model.CalculateAgeAt{
											BinaryExpression: &model.BinaryExpression{
												Expression: model.ResultType(types.Integer),
												Operands: []model.IExpression{
													&model.ToDateTime{
														UnaryExpression: &model.UnaryExpression{
															Expression: model.ResultType(types.DateTime),
															Operand: &model.Property{
																Expression: model.ResultType(types.Date),
																Source: &model.Property{
																	Expression: model.ResultType(&types.Named{TypeName: "FHIR.date"}),
																	Source: &model.ExpressionRef{
																		Expression: model.ResultType(&types.Named{TypeName: "FHIR.Patient"}),
																		Name:       "Patient",
																	},
																	Path: "birthDate",
																},
																Path: "value",
															},
														},
													},
													&model.Start{
														UnaryExpression: &model.UnaryExpression{
															Operand:    &model.ParameterRef{Name: "Measurement Period", Expression: model.ResultType(&types.Interval{PointType: types.DateTime})},
															Expression: model.ResultType(types.DateTime),
														},
													},
												},
											},
											Precision: model.YEAR,
										},
										model.NewLiteral("80", types.Integer),
									},
									Expression: model.ResultType(types.Boolean),
								},
							},
						},
						&model.ExpressionDef{
							Name:        "Denominator",
							Context:     "Patient",
							AccessLevel: "PUBLIC",
							Element:     &model.Element{ResultType: types.Boolean},
							Expression: &model.And{
								BinaryExpression: &model.BinaryExpression{
									Expression: model.ResultType(types.Boolean),
									Operands: []model.IExpression{
										&model.ExpressionRef{Name: "Initial Population", Expression: model.ResultType(types.Boolean)},
										&model.ExpressionRef{Name: "Has coronary heart disease", Expression: model.ResultType(types.Boolean)},
									},
								},
							},
						},
						&model.ExpressionDef{
							Name:        "Numerator",
							Context:     "Patient",
							AccessLevel: "PUBLIC",
							Element:     &model.Element{ResultType: types.Boolean},
							Expression: &model.And{
								BinaryExpression: &model.BinaryExpression{
									Expression: model.ResultType(types.Boolean),
									Operands: []model.IExpression{
										&model.And{
											BinaryExpression: &model.BinaryExpression{
												Expression: model.ResultType(types.Boolean),
												Operands: []model.IExpression{
													&model.ExpressionRef{Name: "Initial Population", Expression: model.ResultType(types.Boolean)},
													&model.ExpressionRef{Name: "Denominator", Expression: model.ResultType(types.Boolean)},
												},
											},
										},
										&model.ExpressionRef{Name: "Most recent blood pressure reading below 150", Expression: model.ResultType(types.Boolean)},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parsedLibs, err := newFHIRParser(t).Libraries(context.Background(), []string{test.cql}, Config{})
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(test.want, parsedLibs[0]); diff != "" {
				t.Errorf("Parsing diff (-want +got):\n%s", diff)
			}
		})
	}
}
