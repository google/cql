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
	"fmt"
	"strings"
	"testing"

	"github.com/google/cql/internal/testhelpers"
	"github.com/google/cql/model"
	"github.com/google/cql/result"
	"github.com/google/cql/retriever/local"
	"github.com/google/cql/terminology"
	"github.com/google/cql/types"
	r4pb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/resources/bundle_and_contained_resource_go_proto"
)

func buildRetriever(t testing.TB) *local.Retriever {
	// TODO(b/300653289): Getting this to parse can be finicky with new lines. Find a more robust way.
	bundle := `{
				"resourceType": "Bundle",
				"type": "transaction",
				"entry": [
					{
						"fullUrl": "fullUrl",
						"resource": {
							"resourceType": "Patient",
							"id": "1",
							"active": true,
							"name": [{"given":["John", "Smith"], "family":"Doe"}]}
					},
					{
						"fullUrl": "fullUrl",
						"resource": {
							"resourceType": "Observation",
							"id": "1",
							"code" : {
								"coding" : [{
									"system" : "http://example.com",
									"code" : "15074-8",
									"display" : "Glucose [Moles/volume] in Blood"
								}]
							}
							}
					},
					{
						"fullUrl": "fullUrl",
						"resource": {
							"resourceType": "Observation",
							"id": "2"}
					},
					{
						"fullUrl": "fullUrl",
						"resource": {
							"resourceType": "Encounter",
							"id": "1"}
					}
				 ]
			}`
	r, err := local.NewRetrieverFromR4Bundle([]byte(bundle))
	if err != nil {
		t.Fatalf("Failed to create retriever: %v", err)
	}
	return r
}

func getTerminologyProvider(t testing.TB) terminology.Provider {
	t.Helper()
	terminologies := []string{
		`{
			"resourceType": "ValueSet",
			"id": "https://example.com/glucose",
			"url": "https://example.com/glucose",
			"version": "1.0.0",
			"expansion": {
				"contains": [
					{ "system": "http://example.com", "code": "15074-8" }
				]
			}
		}`,
	}
	dir := testhelpers.WriteJSONs(t, terminologies)
	tp, err := terminology.NewLocalFHIRProvider(dir)
	if err != nil {
		t.Fatalf("Failed to create terminology provider: %v", err)
	}
	return tp
}

func helperLib(t *testing.T) *model.Library {
	t.Helper()
	// library example.helpers version '1.0'
	// define public "public def" : 2
	// define private "private def": 3
	return &model.Library{
		Identifier: &model.LibraryIdentifier{
			Local:     "helpers",
			Qualified: "example.helpers",
			Version:   "1.0",
		},
		Usings: []*model.Using{
			{
				LocalIdentifier: "FHIR",
				Version:         "4.0.1",
				URI:             "http://hl7.org/fhir",
			},
		},
		Parameters: []*model.ParameterDef{
			{Name: "param true", AccessLevel: model.Public, Element: &model.Element{ResultType: types.Boolean}},
			{Name: "private param false", AccessLevel: model.Private, Element: &model.Element{ResultType: types.Boolean}},
			{Name: "param interval", AccessLevel: model.Public, Element: &model.Element{ResultType: &types.Interval{PointType: types.DateTime}}},
			{
				Name:        "param default",
				AccessLevel: model.Public,
				Default:     &model.Literal{Value: "2", Expression: model.ResultType(types.Integer)},
				Element:     &model.Element{ResultType: types.Integer}},
		},
		Valuesets: []*model.ValuesetDef{
			{
				Name:        "public valueset",
				ID:          "PublicValueset",
				Version:     "1.0",
				AccessLevel: "PUBLIC",
			},
			{
				Name:        "private valueset",
				ID:          "PrivateValueset",
				Version:     "1.0",
				AccessLevel: "PRIVATE",
			},
		},
		CodeSystems: []*model.CodeSystemDef{
			{
				Name:        "public codesystem",
				ID:          "PublicCodeSystem",
				Version:     "1.0",
				AccessLevel: "PUBLIC",
			},
		},
		Statements: &model.Statements{
			Defs: []model.IExpressionDef{
				&model.FunctionDef{
					ExpressionDef: &model.ExpressionDef{
						Name:        "public func",
						Context:     "Patient",
						AccessLevel: "PUBLIC",
						Expression: &model.OperandRef{
							Name:       "A",
							Expression: &model.Expression{Element: &model.Element{ResultType: types.Integer}},
						},
						Element: &model.Element{ResultType: types.Integer},
					},
					Operands: []model.OperandDef{{Name: "A", Expression: &model.Expression{Element: &model.Element{ResultType: types.Integer}}}},
				},
				&model.FunctionDef{
					ExpressionDef: &model.ExpressionDef{
						Name:        "private func",
						Context:     "Patient",
						AccessLevel: "PRIVATE",
						Expression: &model.OperandRef{
							Name:       "A",
							Expression: &model.Expression{Element: &model.Element{ResultType: types.Integer}},
						},
						Element: &model.Element{ResultType: types.Integer},
					},
					Operands: []model.OperandDef{{Name: "A", Expression: &model.Expression{Element: &model.Element{ResultType: types.Integer}}}},
				},
				&model.ExpressionDef{
					Name:        "public def",
					Context:     "Patient",
					AccessLevel: "PUBLIC",
					Expression:  &model.Literal{Value: "2", Expression: &model.Expression{Element: &model.Element{ResultType: types.Integer}}},
					Element:     &model.Element{ResultType: types.Integer},
				},
				&model.ExpressionDef{
					Name:        "private def",
					Context:     "Patient",
					AccessLevel: "PRIVATE",
					Expression:  &model.Literal{Value: "3", Expression: &model.Expression{Element: &model.Element{ResultType: types.Integer}}},
					Element:     &model.Element{ResultType: types.Integer},
				},
			},
		},
	}
}

func TestFailingEvalSingleLibrary(t *testing.T) {
	tests := []struct {
		name        string
		tree        *model.Library
		errContains string
	}{
		{
			name: "ExpressionRef not found local",
			tree: &model.Library{
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:       "Param",
							Context:    "Patient",
							Expression: &model.ExpressionRef{Name: "Non existent"},
						},
					},
				},
			},
			errContains: "could not resolve",
		},
		{
			name: "property no scope or source",
			tree: &model.Library{
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:    "Param",
							Context: "Patient",
							Expression: &model.Property{
								Path: "active",
							},
						},
					},
				},
			},
			errContains: "source must be populated",
		},
		{
			name: "Message returns log error",
			tree: &model.Library{
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:        "Param",
							Context:     "Patient",
							AccessLevel: "PUBLIC",
							Expression: &model.Message{
								Source:     model.NewLiteral("1.2", types.Decimal),
								Condition:  model.NewLiteral("true", types.Boolean),
								Code:       model.NewLiteral("100", types.String),
								Severity:   model.NewLiteral("Error", types.String),
								Message:    model.NewLiteral("Test Message", types.String),
								Expression: model.ResultType(types.Decimal),
							},
						},
					},
				},
			},
			errContains: "log error",
		},
		{
			name: "Query without source",
			tree: &model.Library{
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:        "Param",
							Context:     "Patient",
							AccessLevel: "PUBLIC",
							Expression: &model.Query{
								Source: []*model.AliasedSource{},
							},
						},
					},
				},
			},
			errContains: "query must have",
		},
		{
			name: "Where is not boolean",
			tree: &model.Library{
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:        "Param",
							Context:     "Patient",
							AccessLevel: "PUBLIC",
							Expression: &model.Query{
								Source: []*model.AliasedSource{
									{
										Alias: "O",
										Source: &model.List{
											List: []model.IExpression{
												&model.Literal{Value: "true", Expression: &model.Expression{Element: &model.Element{ResultType: types.Boolean}}},
											},
											Expression: model.ResultType(&types.List{ElementType: types.Boolean}),
										},
									},
								},
								Where: &model.Literal{Value: "3", Expression: &model.Expression{Element: &model.Element{ResultType: types.Integer}}},
							},
						},
					},
				},
			},
			errContains: "where clause of a query",
		},
		{
			name: "Failed retrieve with mismatched URI",
			tree: &model.Library{
				Usings: []*model.Using{{URI: "http://hl7.org/fhir", Version: "4.0.1", LocalIdentifier: "FHIR"}},
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:    "Param",
							Context: "Patient",
							Expression: &model.Retrieve{
								DataType:   "{http://random}Patient",
								TemplateID: "http://hl7.org/fhir/StructureDefinition/Patient",
								Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Patient"}}),
							},
						},
					},
				},
			},
			errContains: "Resource datatype ({http://random}Patient) did not contain the library uri (http://hl7.org/fhir)",
		},
		{
			name: "Incorrect Using Local Identifier",
			tree: &model.Library{
				Usings: []*model.Using{{URI: "http://hl7.org/fhir", Version: "4.0.1", LocalIdentifier: "FIRE"}},
			},
			errContains: "FIRE 4.0.1 data model not found",
		},
		{
			name: "Incorrect Using Version",
			tree: &model.Library{
				Usings: []*model.Using{{URI: "http://hl7.org/fhir", Version: "4.0.2", LocalIdentifier: "FHIR"}},
			},
			errContains: "FHIR 4.0.2 data model not found",
		},
		{
			name: "First unsupported literal",
			tree: &model.Library{
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:    "Param",
							Context: "Patient",
							Expression: &model.UnaryExpression{
								Operand: &model.Literal{
									Value:      "false",
									Expression: &model.Expression{Element: &model.Element{ResultType: types.Boolean}},
								},
							},
						},
					},
				},
			},
			errContains: "internal error - unsupported expression",
		},
		{
			name: "Retrieve Observations without ValuesetRef",
			tree: &model.Library{
				Usings:    []*model.Using{{URI: "http://hl7.org/fhir", Version: "4.0.1", LocalIdentifier: "FHIR"}},
				Valuesets: []*model.ValuesetDef{{Name: "Test Glucose", ID: "https://example.com/glucose", Version: "1.0.0"}},
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:    "Param",
							Context: "Patient",
							Expression: &model.Retrieve{
								CodeProperty: "code",
								Codes:        &model.UnaryExpression{}, // something that's not a ValuesetRef.
								DataType:     "{http://hl7.org/fhir}Observation",
								TemplateID:   "http://hl7.org/fhir/StructureDefinition/Observation",
								Expression:   model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}}),
							},
						},
					},
				},
			},
			errContains: "only ValueSet references are currently supported for valueset filtering",
		},
		{
			name: "Retrieve Observations with incorrect CodeProperty",
			tree: &model.Library{
				Usings:    []*model.Using{{URI: "http://hl7.org/fhir", Version: "4.0.1", LocalIdentifier: "FHIR"}},
				Valuesets: []*model.ValuesetDef{{Name: "Test Glucose", ID: "https://example.com/glucose", Version: "1.0.0"}},
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:    "Param",
							Context: "Patient",
							Expression: &model.Retrieve{
								CodeProperty: "id", // this isn't the code property.
								Codes:        &model.ValuesetRef{Name: "Test Glucose", Expression: model.ResultType(types.ValueSet)},
								DataType:     "{http://hl7.org/fhir}Observation",
								TemplateID:   "http://hl7.org/fhir/StructureDefinition/Observation",
								Expression:   model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Patient"}}),
							},
						},
					},
				},
			},
			errContains: "input proto Value must be a *dtpb.CodeableConcept type",
		},
		{
			name: "Retrieve Observations with missing CodeProperty",
			tree: &model.Library{
				Usings:    []*model.Using{{URI: "http://hl7.org/fhir", Version: "4.0.1", LocalIdentifier: "FHIR"}},
				Valuesets: []*model.ValuesetDef{{Name: "Test Glucose", ID: "https://example.com/glucose", Version: "1.0.0"}},
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:    "Param",
							Context: "Patient",
							Expression: &model.Retrieve{
								CodeProperty: "", // empty
								Codes:        &model.ValuesetRef{Name: "Test Glucose", Expression: model.ResultType(types.ValueSet)},
								DataType:     "{http://hl7.org/fhir}Observation",
								TemplateID:   "http://hl7.org/fhir/StructureDefinition/Observation",
								Expression:   model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}}),
							},
						},
					},
				},
			},
			errContains: "code property must be populated when filtering on codes",
		},
		{
			name: "Retrieve Observations with missing ValueSet",
			tree: &model.Library{
				Usings:    []*model.Using{{URI: "http://hl7.org/fhir", Version: "4.0.1", LocalIdentifier: "FHIR"}},
				Valuesets: []*model.ValuesetDef{{Name: "Test Glucose", ID: "https://example.com/glucose", Version: "1.0.0"}},
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:    "Param",
							Context: "Patient",
							Expression: &model.Retrieve{
								CodeProperty: "code",
								Codes:        &model.ValuesetRef{Name: "Something Missing!", Expression: model.ResultType(types.ValueSet)},
								DataType:     "{http://hl7.org/fhir}Observation",
								TemplateID:   "http://hl7.org/fhir/StructureDefinition/Observation",
								Expression:   model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}}),
							},
						},
					},
				},
			},
			errContains: "could not resolve the local reference",
		},
		{
			name: "List with invalid element",
			tree: &model.Library{
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:    "Param",
							Context: "Patient",
							Expression: &model.List{
								Expression: &model.Expression{Element: &model.Element{ResultType: &types.List{ElementType: types.Integer}}},
								List: []model.IExpression{
									// Property is missing scope and source
									&model.Property{
										Path: "active",
									},
								},
							},
						},
					},
				},
			},
			errContains: "at index 0",
		},
		{
			name: "Invalid Date Literal",
			tree: &model.Library{
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:    "Param",
							Context: "Patient",
							Expression: &model.Literal{
								Value:      "@2024-03-3",
								Expression: model.ResultType(types.Date),
							},
						},
					},
				},
			},
			errContains: "got System.Date @2024-03-3 but want a layout like @YYYY-MM-DD",
		},
		{
			name: "Invalid DateTime Literal",
			tree: &model.Library{
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:    "Param",
							Context: "Patient",
							Expression: &model.Literal{
								Value:      "@20288",
								Expression: model.ResultType(types.DateTime),
							},
						},
					},
				},
			},
			errContains: "got System.DateTime @20288 but want a layout like @YYYY-MM-DDThh:mm:ss.fff(Z|(+/-hh:mm)",
		},
		{
			name: "Invalid Time Literal",
			tree: &model.Library{
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:    "Param",
							Context: "Patient",
							Expression: &model.Literal{
								Value:      "@T30",
								Expression: model.ResultType(types.Time),
							},
						},
					},
				},
			},
			errContains: "got System.Time @T30 but want a layout like @Thh:mm:ss.fff: hour out of range",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.tree.Identifier = &model.LibraryIdentifier{Qualified: "Highly.Qualified", Version: "1.0"}
			_, err := Eval(context.Background(), []*model.Library{test.tree}, defaultInterpreterConfig(t))
			if err == nil {
				t.Errorf("Eval Library(%s) = nil, want error", test.name)
			}
			if err != nil && !strings.Contains(err.Error(), test.errContains) {
				t.Errorf("Returned error (%s) did not contain expected string (%s)", err, test.errContains)
			}
		})
	}
}

func TestFailingEvalMultipleLibraries(t *testing.T) {
	tests := []struct {
		name        string
		tree        *model.Library
		errContains string
	}{
		{
			name: "ExpressionRef global private",
			tree: &model.Library{
				Identifier: &model.LibraryIdentifier{Qualified: "example.measure", Version: "1.0"},
				Includes: []*model.Include{
					{Identifier: &model.LibraryIdentifier{Local: "helpers", Qualified: "example.helpers", Version: "1.0"}},
				},
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:       "Param",
							Context:    "Patient",
							Expression: &model.ExpressionRef{Name: "private def", LibraryName: "helpers", Expression: model.ResultType(types.Integer)},
							Element:    &model.Element{ResultType: types.Integer},
						},
					},
				},
			},
			errContains: "helpers.private def is not public",
		},
		{
			name: "ExpressionRef library does not exist",
			tree: &model.Library{
				Identifier: &model.LibraryIdentifier{Qualified: "example.measure", Version: "1.0"},
				Includes: []*model.Include{
					{Identifier: &model.LibraryIdentifier{Local: "helpers", Qualified: "example.helpers", Version: "1.0"}},
				},
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:       "Param",
							Context:    "Patient",
							Expression: &model.ExpressionRef{Name: "public def", LibraryName: "non existent"},
						},
					},
				},
			},
			errContains: "resolve the library name",
		},
		{
			name: "ParameterRef global private",
			tree: &model.Library{
				Identifier: &model.LibraryIdentifier{Qualified: "example.measure", Version: "1.0"},
				Includes: []*model.Include{
					{Identifier: &model.LibraryIdentifier{Local: "helpers", Qualified: "example.helpers", Version: "1.0"}},
				},
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:       "Param",
							Context:    "Patient",
							Expression: &model.ParameterRef{Name: "private param false", LibraryName: "helpers", Expression: model.ResultType(types.Boolean)},
						},
					},
				},
			},
			errContains: "helpers.private param false is not public",
		},
		{
			name: "ParameterRef library does not exist",
			tree: &model.Library{
				Identifier: &model.LibraryIdentifier{Qualified: "example.measure", Version: "1.0"},
				Includes: []*model.Include{
					{Identifier: &model.LibraryIdentifier{Local: "helpers", Qualified: "example.helpers", Version: "1.0"}},
				},
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:       "Param",
							Context:    "Patient",
							Expression: &model.ParameterRef{Name: "param true", LibraryName: "non existent"},
						},
					},
				},
			},
			errContains: "resolve the library name",
		},
		{
			name: "ValuesetRef global private",
			tree: &model.Library{
				Identifier: &model.LibraryIdentifier{Qualified: "example.measure", Version: "1.0"},
				Includes: []*model.Include{
					{Identifier: &model.LibraryIdentifier{Local: "helpers", Qualified: "example.helpers", Version: "1.0"}},
				},
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:       "Param",
							Context:    "Patient",
							Expression: &model.ValuesetRef{Name: "private valueset", LibraryName: "helpers", Expression: model.ResultType(types.ValueSet)},
						},
					},
				},
			},
			errContains: "helpers.private valueset is not public",
		},
		{
			name: "ValuesetRef library does not exist",
			tree: &model.Library{
				Identifier: &model.LibraryIdentifier{Qualified: "example.measure", Version: "1.0"},
				Includes: []*model.Include{
					{Identifier: &model.LibraryIdentifier{Local: "helpers", Qualified: "example.helpers", Version: "1.0"}},
				},
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.ExpressionDef{
							Name:       "Param",
							Context:    "Patient",
							Expression: &model.ValuesetRef{Name: "public valuset", LibraryName: "non existent", Expression: model.ResultType(types.ValueSet)},
						},
					},
				},
			},
			errContains: "resolve the library name",
		},
		{
			name: "Unnamed lib with parameter defs",
			tree: &model.Library{
				Parameters: []*model.ParameterDef{
					{
						Name:        "param in unnamed",
						AccessLevel: model.Public,
					},
				},
			},
			errContains: "unnamed libraries cannot have parameters",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Eval(context.Background(), []*model.Library{helperLib(t), test.tree}, defaultInterpreterConfig(t))
			if err == nil {
				t.Errorf("Eval Library(%s) = nil, want error", test.name)
			}
			if err != nil && !strings.Contains(err.Error(), test.errContains) {
				t.Errorf("Returned error (%s) did not contain expected string (%s)", err, test.errContains)
			}
		})
	}
}

type InvalidRetriever struct{}

func (i *InvalidRetriever) Retrieve(_ context.Context, _ string) ([]*r4pb.ContainedResource, error) {
	return []*r4pb.ContainedResource{{}}, nil
}

func TestInvalidContainedResource(t *testing.T) {
	tree := &model.Library{
		Usings: []*model.Using{{URI: "http://hl7.org/fhir", Version: "4.0.1", LocalIdentifier: "FHIR"}},
		Statements: &model.Statements{
			Defs: []model.IExpressionDef{
				&model.ExpressionDef{
					Name:    "Param",
					Context: "Patient",
					Expression: &model.Retrieve{
						DataType:   "{http://hl7.org/fhir}Patient",
						TemplateID: "http://hl7.org/fhir/StructureDefinition/Patient",
						Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Patient"}}),
					},
				},
			},
		},
	}
	config := defaultInterpreterConfig(t)
	config.Retriever = &InvalidRetriever{}
	_, err := Eval(context.Background(), []*model.Library{tree}, config)
	if err == nil {
		t.Errorf("Eval Library() = nil, want error")
	}
	if !strings.Contains(err.Error(), "no resource type was populated") {
		t.Errorf("Returned error (%s) did not contain expected string (%s)", err, "no resource type was populated")
	}
}

func TestConvertValuesWith(t *testing.T) {
	tests := []struct {
		name    string
		left    result.Value
		right   result.Value
		fn      func(result.Value) (int32, error)
		wantErr string
	}{
		{
			name:  "left value throws error on error",
			left:  newOrFatal(t, 1),
			right: newOrFatal(t, 2),
			fn: func(o result.Value) (int32, error) {
				if o.GolangValue().(int32) == 1 {
					return 0, fmt.Errorf("error invalid value")
				}
				return o.GolangValue().(int32), nil
			},
			wantErr: "error invalid value",
		},
		{
			name:  "left value throws error on error",
			left:  newOrFatal(t, 2),
			right: newOrFatal(t, 1),
			fn: func(o result.Value) (int32, error) {
				if o.GolangValue().(int32) == 1 {
					return 0, fmt.Errorf("error invalid value")
				}
				return o.GolangValue().(int32), nil
			},
			wantErr: "error invalid value",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := applyToValues(tc.left, tc.right, tc.fn)
			if err == nil {
				t.Errorf("applyToValue() expected error but got none with args: %v %v, wanted: %s, got: %v", tc.left, tc.right, tc.wantErr, err)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("applyToValues() returned unexpected error: %v, wanted: %s", err, tc.wantErr)
			}
		})
	}
}
