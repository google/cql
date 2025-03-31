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

package cql_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/cql"
	"github.com/google/cql/model"
	"github.com/google/cql/parser"
	"github.com/google/cql/result"
	"github.com/google/cql/retriever"
	"github.com/google/cql/tests/enginetests"
	"github.com/google/cql/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/lithammer/dedent"
	"google.golang.org/protobuf/testing/protocmp"
)

// CQL Engine tests are for testing the top level CQL Engine API. For detailed testing see the tests
// folder.

func TestCQL(t *testing.T) {
	tests := []struct {
		name                 string
		cql                  []string
		retriever            retriever.Retriever
		parserConfig         cql.ParseConfig
		evalConfig           cql.EvalConfig
		wantResult           result.Value
		wantSourceExpression model.IExpression
		wantSourceValues     []result.Value
	}{
		{
			name: "Simple Query with Retriever",
			cql: []string{dedent.Dedent(`
			library TESTLIB version '1.0.0'
			using FHIR version '4.0.1'
			include FHIRHelpers version '4.0.1'
			context Patient
			define TESTRESULT: [Encounter] E`),
				fhirHelpers(t),
			},
			parserConfig: cql.ParseConfig{
				DataModels: [][]byte{fhirDataModel(t)},
			},
			retriever: enginetests.BuildRetriever(t),
			wantResult: newOrFatal(t, result.List{Value: []result.Value{
				newOrFatal(t, result.Named{Value: enginetests.RetrieveFHIRResource(t, "Encounter", "1"), RuntimeType: &types.Named{TypeName: "FHIR.Encounter"}}),
				newOrFatal(t, result.Named{Value: enginetests.RetrieveFHIRResource(t, "Encounter", "2"), RuntimeType: &types.Named{TypeName: "FHIR.Encounter"}}),
			},
				StaticType: &types.List{ElementType: &types.Named{TypeName: "FHIR.Encounter"}},
			}),
			wantSourceExpression: &model.Query{
				Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Encounter"}}),
				Source: []*model.AliasedSource{
					&model.AliasedSource{
						Alias:      "E",
						Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Encounter"}}),
						Source: &model.Retrieve{
							DataType:     "{http://hl7.org/fhir}Encounter",
							TemplateID:   "http://hl7.org/fhir/StructureDefinition/Encounter",
							CodeProperty: "type",
							Expression:   model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Encounter"}}),
						},
					},
				},
			},
			wantSourceValues: []result.Value{
				newOrFatal(t, result.List{Value: []result.Value{
					newOrFatal(t, result.Named{Value: enginetests.RetrieveFHIRResource(t, "Encounter", "1"), RuntimeType: &types.Named{TypeName: "FHIR.Encounter"}}),
					newOrFatal(t, result.Named{Value: enginetests.RetrieveFHIRResource(t, "Encounter", "2"), RuntimeType: &types.Named{TypeName: "FHIR.Encounter"}}),
				},
					StaticType: &types.List{ElementType: &types.Named{TypeName: "FHIR.Encounter"}},
				}),
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			elm, err := cql.Parse(context.Background(), tc.cql, tc.parserConfig)
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}

			results, err := elm.Eval(context.Background(), tc.retriever, tc.evalConfig)
			if err != nil {
				t.Fatalf("Eval returned unexpected error: %v", err)
			}

			gotResult := getTESTRESULTWithSources(t, results)
			if diff := cmp.Diff(tc.wantResult, gotResult, protocmp.Transform()); diff != "" {
				t.Errorf("Eval diff (-want +got)\n%v", diff)
			}
			if diff := cmp.Diff(tc.wantSourceExpression, gotResult.SourceExpression(), protocmp.Transform()); tc.wantSourceExpression != nil && diff != "" {
				t.Errorf("Eval SourceExpression diff (-want +got)\n%v", diff)
			}
			if diff := cmp.Diff(tc.wantSourceValues, gotResult.SourceValues(), protocmp.Transform()); tc.wantSourceValues != nil && diff != "" {
				t.Errorf("Eval SourceValues diff (-want +got)\n%v", diff)
			}
		})
	}
}

func TestCQL_EvalEmptyTimestamp(t *testing.T) {
	// This test checks that an empty EvalConfig.EvaluationTimestamp is set to time.Now() at the start
	// of the eval request. We do this by sanity checking the result of Now() is within 2 mins of the
	// test assertion time.
	cqlSources := []string{dedent.Dedent(`
	library TESTLIB version '1.0.0'
	define TESTRESULT: Now()`)}
	parserConfig := cql.ParseConfig{}
	evalConfig := cql.EvalConfig{}

	elm, err := cql.Parse(context.Background(), cqlSources, parserConfig)
	if err != nil {
		t.Fatalf("Parse returned unexpected error: %v", err)
	}

	results, err := elm.Eval(context.Background(), nil, evalConfig)
	if err != nil {
		t.Fatalf("Eval returned unexpected error: %v", err)
	}

	gotResult := getTESTRESULTWithSources(t, results)
	dt, err := result.ToDateTime(gotResult)
	if err != nil {
		t.Errorf("returned result is not a DateTime, err: %v", err)
	}

	// Check that the dt is within [time.Now()-2min, time.Now()]
	end := time.Now()
	start := end.Add(-2 * time.Minute)
	if dt.Date.Before(start) || dt.Date.After(end) {
		t.Errorf("returned CQL Now() result is not within 2 mins of time.Now()")
	}
}

func TestCQL_ParseErrors(t *testing.T) {
	tests := []struct {
		name         string
		cql          []string
		retriever    retriever.Retriever
		parserConfig cql.ParseConfig
		wantErr      error
	}{
		{
			name: "Invalid named library",
			cql: []string{dedent.Dedent(`
			library TESTLIB version '1.0.0'
			using FHIR version '4.0.1'
			context Patient
			define TESTRESULT: [Encounter] E`)},
			retriever: enginetests.BuildRetriever(t),
			wantErr: &parser.LibraryErrors{
				LibKey: result.LibKey{Name: "TESTLIB", IsUnnamed: false /* Version is ignored*/},
				Errors: []*parser.ParsingError{
					{Message: "FHIR 4.0.1 data model not found", Line: 3, Column: 0},
					{Message: "using declaration has not been set", Line: 4, Column: 0},
					{Message: "using declaration has not been set", Line: 5, Column: 20},
					{Message: "retrieves cannot be performed on type System.Any", Line: 5, Column: 19},
				},
			},
		},
		{
			name:      "Invalid unnamed library",
			cql:       []string{"define Foo: 1 + 'a'"},
			retriever: enginetests.BuildRetriever(t),
			wantErr: &parser.LibraryErrors{
				LibKey: result.LibKey{Name: "Unnamed Library", IsUnnamed: true /* Version is ignored*/},
				Errors: []*parser.ParsingError{
					{
						Message: "could not resolve Add(System.Integer, System.String): no matching overloads\nAvailable overloads:\n  Add(System.Integer, System.Integer)\n  Add(System.Long, System.Long)\n  Add(System.Decimal, System.Decimal)\n  Add(System.Quantity, System.Quantity)\n  Add(System.String, System.String)\n  Add(System.Date, System.Quantity)\n  Add(System.DateTime, System.Quantity)\n  Add(System.Time, System.Quantity)\n",
						Line:    1,
						Column:  12,
					},
				},
			},
		},
		{
			name: "Invalid parameter",
			cql: []string{dedent.Dedent(`
			library TESTLIB version '1.0.0'
			using FHIR version '4.0.1'
			context Patient
			define TESTRESULT: [Encounter] E`),
				fhirHelpers(t),
			},
			retriever: enginetests.BuildRetriever(t),
			parserConfig: cql.ParseConfig{
				Parameters: map[result.DefKey]string{
					result.DefKey{Name: "param name", Library: result.LibKey{Name: "TESTLIB", Version: "1.0.0"}}: "invalid value",
				},
				DataModels: [][]byte{fhirDataModel(t)},
			},
			wantErr: &parser.ParameterErrors{
				DefKey: result.DefKey{
					Name:    "param name",
					Library: result.LibKey{Name: "TESTLIB", Version: "1.0.0"},
				},
				Errors: []*parser.ParsingError{{Message: "must be a single literal"}},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := cql.Parse(context.Background(), tc.cql, tc.parserConfig)
			if err == nil {
				t.Fatalf("Parse succeeded, expected error")
			}
			if diff := cmp.Diff(tc.wantErr, err, cmpopts.IgnoreFields(parser.LibraryErrors{}, "LibKey.Version")); diff != "" {
				t.Errorf("Parse returned err: %v, want %v (-want, +got): %v", err, tc.wantErr, diff)
			}
		})
	}
}

func TestCQL_EvalErrors(t *testing.T) {
	tests := []struct {
		name         string
		cql          []string
		retriever    retriever.Retriever
		parserConfig cql.ParseConfig
		evalConfig   cql.EvalConfig
	}{
		{
			name: "Retriever nil error",
			cql: []string{dedent.Dedent(`
			library TESTLIB version '1.0.0'
			using FHIR version '4.0.1'
			include FHIRHelpers version '4.0.1'
			context Patient
			define TESTRESULT: [Encounter] E`),
				fhirHelpers(t),
			},
			parserConfig: cql.ParseConfig{
				DataModels: [][]byte{fhirDataModel(t)},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			elm, err := cql.Parse(context.Background(), tc.cql, tc.parserConfig)
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}

			_, err = elm.Eval(context.Background(), tc.retriever, tc.evalConfig)
			if err == nil {
				t.Fatalf("Eval succeeded, expected error")
			}
			engErr, ok := err.(result.EngineError)
			if !ok {
				t.Fatalf("Returned error (%s) was not a result.EngineError", err)
			}
			if !errors.Is(engErr.ErrType, result.ErrEvaluationError) {
				t.Errorf("Returned error (%s) was not a result.ErrEvaluationError error", err)
			}
		})
	}
}

func TestCQL_MultipleEvals(t *testing.T) {
	tests := []struct {
		name                 string
		cql                  []string
		retriever            retriever.Retriever
		parserConfig         cql.ParseConfig
		evalConfig           cql.EvalConfig
		wantResult           result.Value
		wantSourceExpression model.IExpression
		wantSourceValues     []result.Value
	}{
		{
			name: "Simple Query with Retriever",
			cql: []string{dedent.Dedent(`
			library TESTLIB version '1.0.0'
			using FHIR version '4.0.1'
			include FHIRHelpers version '4.0.1'
			context Patient
			define TESTRESULT: [Encounter] E`),
				fhirHelpers(t),
			},
			parserConfig: cql.ParseConfig{
				DataModels: [][]byte{fhirDataModel(t)},
			},
			retriever: enginetests.BuildRetriever(t),
			wantResult: newOrFatal(t, result.List{
				Value: []result.Value{
					newOrFatal(t, result.Named{Value: enginetests.RetrieveFHIRResource(t, "Encounter", "1"), RuntimeType: &types.Named{TypeName: "FHIR.Encounter"}}),
					newOrFatal(t, result.Named{Value: enginetests.RetrieveFHIRResource(t, "Encounter", "2"), RuntimeType: &types.Named{TypeName: "FHIR.Encounter"}}),
				},
				StaticType: &types.List{ElementType: &types.Named{TypeName: "FHIR.Encounter"}},
			}),
			wantSourceExpression: &model.Query{
				Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Encounter"}}),
				Source: []*model.AliasedSource{
					&model.AliasedSource{
						Alias:      "E",
						Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Encounter"}}),
						Source: &model.Retrieve{
							DataType:     "{http://hl7.org/fhir}Encounter",
							TemplateID:   "http://hl7.org/fhir/StructureDefinition/Encounter",
							CodeProperty: "type",
							Expression:   model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Encounter"}}),
						},
					},
				},
			},
			wantSourceValues: []result.Value{
				newOrFatal(t, result.List{
					Value: []result.Value{
						newOrFatal(t, result.Named{Value: enginetests.RetrieveFHIRResource(t, "Encounter", "1"), RuntimeType: &types.Named{TypeName: "FHIR.Encounter"}}),
						newOrFatal(t, result.Named{Value: enginetests.RetrieveFHIRResource(t, "Encounter", "2"), RuntimeType: &types.Named{TypeName: "FHIR.Encounter"}}),
					},
					StaticType: &types.List{ElementType: &types.Named{TypeName: "FHIR.Encounter"}},
				}),
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			elm, err := cql.Parse(context.Background(), tc.cql, tc.parserConfig)
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}

			for i := 0; i < 3; i++ {
				results, err := elm.Eval(context.Background(), tc.retriever, tc.evalConfig)
				if err != nil {
					t.Fatalf("Eval returned unexpected error: %v", err)
				}

				gotResult := getTESTRESULTWithSources(t, results)
				if diff := cmp.Diff(tc.wantResult, gotResult, protocmp.Transform()); diff != "" {
					t.Errorf("Eval diff (-want +got)\n%v", diff)
				}
				if diff := cmp.Diff(tc.wantSourceExpression, gotResult.SourceExpression(), protocmp.Transform()); tc.wantSourceExpression != nil && diff != "" {
					t.Errorf("Eval SourceExpression diff (-want +got)\n%v", diff)
				}
				if diff := cmp.Diff(tc.wantSourceValues, gotResult.SourceValues(), protocmp.Transform()); tc.wantSourceValues != nil && diff != "" {
					t.Errorf("Eval SourceValues diff (-want +got)\n%v", diff)
				}
			}
		})
	}
}

func fhirDataModel(t testing.TB) []byte {
	t.Helper()
	fdm, err := cql.FHIRDataModel("4.0.1")
	if err != nil {
		t.Fatalf("FHIRDataModel returned unexpected error: %v", err)
	}
	return fdm
}

func fhirHelpers(t testing.TB) string {
	t.Helper()
	fh, err := cql.FHIRHelpersLib("4.0.1")
	if err != nil {
		t.Fatalf("FHIRHelepersLib returned unexpected error: %v", err)
	}
	return fh
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

// newOrFatal returns a new result.Value or calls fatal on error.
func newOrFatal(t testing.TB, a any) result.Value {
	t.Helper()
	o, err := result.New(a)
	if err != nil {
		t.Fatalf("New(%v) returned unexpected error: %v", a, err)
	}
	return o
}
