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

package parser

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/google/cql/model"
	"github.com/google/cql/types"
	"github.com/google/go-cmp/cmp"
	"github.com/lithammer/dedent"
)

func TestQuery(t *testing.T) {
	tests := []struct {
		name string
		desc string
		cql  string
		want model.IExpression
	}{
		{
			name: "Query",
			cql:  dedent.Dedent(`define TESTRESULT: [Observation] o`),
			want: &model.Query{
				Expression: &model.Expression{
					Element: &model.Element{ResultType: &types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}}},
				},
				Source: []*model.AliasedSource{
					{
						Alias: "o",
						Source: &model.Retrieve{
							Expression:   model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}}),
							DataType:     "{http://hl7.org/fhir}Observation",
							TemplateID:   "http://hl7.org/fhir/StructureDefinition/Observation",
							CodeProperty: "code",
						},
						Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}}),
					},
				},
			},
		},
		{
			name: "Query with Filtered Retrieve",
			cql: dedent.Dedent(`
			define TESTRESULT: [Observation: "Blood pressure"] bp`),
			want: &model.Query{
				Expression: &model.Expression{
					Element: &model.Element{ResultType: &types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}}},
				},
				Source: []*model.AliasedSource{
					{
						Alias: "bp",
						Source: &model.Retrieve{
							Expression:   model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}}),
							DataType:     "{http://hl7.org/fhir}Observation",
							TemplateID:   "http://hl7.org/fhir/StructureDefinition/Observation",
							CodeProperty: "code",
							Codes:        &model.ValuesetRef{Name: "Blood pressure", Expression: model.ResultType(types.ValueSet)},
						},
						Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}}),
					},
				},
			},
		},
		{
			name: "Let",
			cql:  dedent.Dedent(`define TESTRESULT: [Observation] O let A: 4, B: 5 return A`),
			want: &model.Query{
				Expression: model.ResultType(&types.List{ElementType: types.Integer}),
				Source: []*model.AliasedSource{
					{
						Alias: "O",
						Source: &model.Retrieve{
							Expression:   model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}}),
							DataType:     "{http://hl7.org/fhir}Observation",
							TemplateID:   "http://hl7.org/fhir/StructureDefinition/Observation",
							CodeProperty: "code",
						},
						Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}}),
					},
				},
				Let: []*model.LetClause{
					{
						Expression: model.NewLiteral("4", types.Integer),
						Identifier: "A",
						Element:    &model.Element{ResultType: types.Integer},
					},
					{
						Expression: model.NewLiteral("5", types.Integer),
						Identifier: "B",
						Element:    &model.Element{ResultType: types.Integer},
					},
				},
				Return: &model.ReturnClause{
					Element:  &model.Element{ResultType: types.Integer},
					Distinct: true,
					Expression: &model.QueryLetRef{
						Name:       "A",
						Expression: model.ResultType(types.Integer),
					},
				},
			},
		},
		{
			name: "Query with relationship",
			cql:  dedent.Dedent(`define TESTRESULT: ({3, 4}) O with ({4, 5}) C such that O = C`),
			want: &model.Query{
				Expression: &model.Expression{
					Element: &model.Element{ResultType: &types.List{ElementType: types.Integer}},
				},
				Source: []*model.AliasedSource{
					{
						Alias:      "O",
						Source:     model.NewList([]string{"3", "4"}, types.Integer),
						Expression: model.ResultType(&types.List{ElementType: types.Integer}),
					},
				},
				Relationship: []model.IRelationshipClause{
					&model.With{
						RelationshipClause: &model.RelationshipClause{
							Element:    &model.Element{ResultType: types.Boolean},
							Expression: model.NewList([]string{"4", "5"}, types.Integer),
							Alias:      "C",
							SuchThat: &model.Equal{
								BinaryExpression: &model.BinaryExpression{
									Expression: model.ResultType(types.Boolean),
									Operands: []model.IExpression{
										&model.AliasRef{Name: "O", Expression: model.ResultType(types.Integer)},
										&model.AliasRef{Name: "C", Expression: model.ResultType(types.Integer)},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Query without relationship",
			cql:  dedent.Dedent(`define TESTRESULT: ({3, 4}) O without ({4, 5}) C such that O = C`),
			want: &model.Query{
				Expression: &model.Expression{
					Element: &model.Element{ResultType: &types.List{ElementType: types.Integer}},
				},
				Source: []*model.AliasedSource{
					{
						Alias:      "O",
						Source:     model.NewList([]string{"3", "4"}, types.Integer),
						Expression: model.ResultType(&types.List{ElementType: types.Integer}),
					},
				},
				Relationship: []model.IRelationshipClause{
					&model.Without{
						RelationshipClause: &model.RelationshipClause{
							Element:    &model.Element{ResultType: types.Boolean},
							Expression: model.NewList([]string{"4", "5"}, types.Integer),
							Alias:      "C",
							SuchThat: &model.Equal{
								BinaryExpression: &model.BinaryExpression{
									Expression: model.ResultType(types.Boolean),
									Operands: []model.IExpression{
										&model.AliasRef{Name: "O", Expression: model.ResultType(types.Integer)},
										&model.AliasRef{Name: "C", Expression: model.ResultType(types.Integer)},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Multi-source",
			cql:  dedent.Dedent(`define TESTRESULT: from [Observation] O, [Patient] P`),
			want: &model.Query{
				Expression: model.ResultType(&types.List{ElementType: &types.Tuple{ElementTypes: map[string]types.IType{"O": &types.Named{TypeName: "FHIR.Observation"}, "P": &types.Named{TypeName: "FHIR.Patient"}}}}),
				Source: []*model.AliasedSource{
					{
						Alias: "O",
						Source: &model.Retrieve{
							Expression:   model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}}),
							DataType:     "{http://hl7.org/fhir}Observation",
							TemplateID:   "http://hl7.org/fhir/StructureDefinition/Observation",
							CodeProperty: "code",
						},
						Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}}),
					},
					{
						Alias: "P",
						Source: &model.Retrieve{
							Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Patient"}}),
							DataType:   "{http://hl7.org/fhir}Patient",
							TemplateID: "http://hl7.org/fhir/StructureDefinition/Patient",
						},
						Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Patient"}}),
					},
				},
				Return: &model.ReturnClause{
					Distinct: true,
					Element:  &model.Element{ResultType: &types.Tuple{ElementTypes: map[string]types.IType{"O": &types.Named{TypeName: "FHIR.Observation"}, "P": &types.Named{TypeName: "FHIR.Patient"}}}},
					Expression: &model.Tuple{
						Expression: model.ResultType(&types.Tuple{ElementTypes: map[string]types.IType{"O": &types.Named{TypeName: "FHIR.Observation"}, "P": &types.Named{TypeName: "FHIR.Patient"}}}),
						Elements: []*model.TupleElement{
							{
								Name:  "O",
								Value: &model.AliasRef{Name: "O", Expression: model.ResultType(&types.Named{TypeName: "FHIR.Observation"})},
							},
							{
								Name:  "P",
								Value: &model.AliasRef{Name: "P", Expression: model.ResultType(&types.Named{TypeName: "FHIR.Patient"})},
							},
						},
					},
				},
			},
		},
		{
			name: "Multi-source on non lists",
			cql:  dedent.Dedent(`define TESTRESULT: from (4) O, ('hi') P`),
			want: &model.Query{
				Expression: model.ResultType(&types.Tuple{ElementTypes: map[string]types.IType{"O": types.Integer, "P": types.String}}),
				Source: []*model.AliasedSource{
					{
						Alias:      "O",
						Source:     model.NewLiteral("4", types.Integer),
						Expression: model.ResultType(types.Integer),
					},
					{
						Alias:      "P",
						Source:     model.NewLiteral("hi", types.String),
						Expression: model.ResultType(types.String),
					},
				},
				Return: &model.ReturnClause{
					Distinct: true,
					Element:  &model.Element{ResultType: &types.Tuple{ElementTypes: map[string]types.IType{"O": types.Integer, "P": types.String}}},
					Expression: &model.Tuple{
						Expression: model.ResultType(&types.Tuple{ElementTypes: map[string]types.IType{"O": types.Integer, "P": types.String}}),
						Elements: []*model.TupleElement{
							{
								Name:  "O",
								Value: &model.AliasRef{Name: "O", Expression: model.ResultType(types.Integer)},
							},
							{
								Name:  "P",
								Value: &model.AliasRef{Name: "P", Expression: model.ResultType(types.String)},
							},
						},
					},
				},
			},
		},
		{
			name: "Multi-source with return",
			cql:  dedent.Dedent(`define TESTRESULT: from ({4, 6}) O, ('hi') P return 5`),
			want: &model.Query{
				Expression: model.ResultType(&types.List{ElementType: types.Integer}),
				Source: []*model.AliasedSource{
					{
						Alias:      "O",
						Source:     model.NewList([]string{"4", "6"}, types.Integer),
						Expression: model.ResultType(&types.List{ElementType: types.Integer}),
					},
					{
						Alias:      "P",
						Source:     model.NewLiteral("hi", types.String),
						Expression: model.ResultType(types.String),
					},
				},
				Return: &model.ReturnClause{
					Distinct:   true,
					Element:    &model.Element{ResultType: types.Integer},
					Expression: model.NewLiteral("5", types.Integer),
				},
			},
		},
		{
			name: "Query with Sort By Direction",
			cql: dedent.Dedent(`
			define TESTRESULT:
				({@2013, @2014, @2015}) dl
				sort desc
			`),
			want: &model.Query{
				Expression: &model.Expression{
					Element: &model.Element{ResultType: &types.List{ElementType: types.Date}},
				},
				Source: []*model.AliasedSource{
					{
						Alias: "dl",
						Source: &model.List{
							Expression: model.ResultType(&types.List{ElementType: types.Date}),
							List: []model.IExpression{
								buildLiteral("@2013", types.Date),
								buildLiteral("@2014", types.Date),
								buildLiteral("@2015", types.Date),
							},
						},
						Expression: model.ResultType(&types.List{ElementType: types.Date}),
					},
				},
				Sort: &model.SortClause{
					ByItems: []model.ISortByItem{
						&model.SortByDirection{SortByItem: &model.SortByItem{Direction: model.DESCENDING}},
					},
				},
			},
		},
		{
			name: "Query with Sort On Column",
			cql: dedent.Dedent(`
			define TESTRESULT:
				[Observation: "Blood pressure"] bp
				sort by effective desc`),
			want: &model.Query{
				Expression: &model.Expression{
					Element: &model.Element{ResultType: &types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}}},
				},
				Source: []*model.AliasedSource{
					{
						Alias: "bp",
						Source: &model.Retrieve{
							Expression:   model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}}),
							DataType:     "{http://hl7.org/fhir}Observation",
							TemplateID:   "http://hl7.org/fhir/StructureDefinition/Observation",
							CodeProperty: "code",
							Codes:        &model.ValuesetRef{Name: "Blood pressure", Expression: model.ResultType(types.ValueSet)},
						},
						Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}}),
					},
				},
				Sort: &model.SortClause{
					ByItems: []model.ISortByItem{
						&model.SortByColumn{
							SortByItem: &model.SortByItem{Direction: model.DESCENDING},
							Path:       "effective",
						},
					},
				},
			},
		},
		{
			name: "Sort by expression",
			cql: dedent.Dedent(`
			  define TESTRESULT: [Encounter] E sort by start of period`),
			want: &model.Query{
				Expression: &model.Expression{
					Element: &model.Element{ResultType: &types.List{ElementType: &types.Named{TypeName: "FHIR.Encounter"}}},
				},
				Source: []*model.AliasedSource{
					{
						Expression: &model.Expression{Element: &model.Element{ResultType: &types.List{ElementType: &types.Named{TypeName: "FHIR.Encounter"}}}},
						Alias:      "E",
						Source: &model.Retrieve{
							Expression:   model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Encounter"}}),
							DataType:     "{http://hl7.org/fhir}Encounter",
							TemplateID:   "http://hl7.org/fhir/StructureDefinition/Encounter",
							CodeProperty: "type",
						},
					},
				},
				Sort: &model.SortClause{
					ByItems: []model.ISortByItem{
						&model.SortByExpression{
							SortByItem: &model.SortByItem{Direction: model.ASCENDING},
							SortExpression: &model.Start{
								UnaryExpression: &model.UnaryExpression{
									Expression: model.ResultType(types.DateTime),
									Operand: &model.FunctionRef{
										Expression:  model.ResultType(&types.Interval{PointType: types.DateTime}),
										Name:        "ToInterval",
										LibraryName: "FHIRHelpers",
										Operands: []model.IExpression{&model.IdentifierRef{
											Expression: model.ResultType(&types.Named{TypeName: "FHIR.Period"}),
											Name:       "period",
										},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Aggregate",
			cql:  "define TESTRESULT: ({1, 2, 3}) N aggregate R starting 1: R * N",
			want: &model.Query{
				Expression: model.ResultType(types.Integer),
				Source: []*model.AliasedSource{
					{
						Alias:      "N",
						Source:     model.NewList([]string{"1", "2", "3"}, types.Integer),
						Expression: model.ResultType(&types.List{ElementType: types.Integer}),
					},
				},
				Aggregate: &model.AggregateClause{
					Element:    &model.Element{ResultType: types.Integer},
					Identifier: "R",
					Starting:   model.NewLiteral("1", types.Integer),
					Distinct:   false,
					Expression: &model.Multiply{
						BinaryExpression: &model.BinaryExpression{
							Expression: model.ResultType(types.Integer),
							Operands: []model.IExpression{
								&model.AliasRef{Name: "R", Expression: model.ResultType(types.Integer)},
								&model.AliasRef{Name: "N", Expression: model.ResultType(types.Integer)},
							},
						},
					},
				},
			},
		},
		{
			name: "Distinct aggregate",
			cql:  "define TESTRESULT: ({1, 2, 3}) N aggregate distinct R starting 'hi': R",
			want: &model.Query{
				Expression: model.ResultType(types.String),
				Source: []*model.AliasedSource{
					{
						Alias:      "N",
						Source:     model.NewList([]string{"1", "2", "3"}, types.Integer),
						Expression: model.ResultType(&types.List{ElementType: types.Integer}),
					},
				},
				Aggregate: &model.AggregateClause{
					Element:    &model.Element{ResultType: types.String},
					Identifier: "R",
					Starting:   model.NewLiteral("hi", types.String),
					Distinct:   true,
					Expression: &model.AliasRef{Name: "R", Expression: model.ResultType(types.String)},
				},
			},
		},
		{
			name: "Aggregate without starting",
			cql:  "define TESTRESULT: ({1, 2, 3}) N aggregate R : R",
			want: &model.Query{
				Expression: model.ResultType(types.Any),
				Source: []*model.AliasedSource{
					{
						Alias:      "N",
						Source:     model.NewList([]string{"1", "2", "3"}, types.Integer),
						Expression: model.ResultType(&types.List{ElementType: types.Integer}),
					},
				},
				Aggregate: &model.AggregateClause{
					Element:    &model.Element{ResultType: types.Any},
					Identifier: "R",
					Starting:   model.NewLiteral("null", types.Any),
					Distinct:   false,
					Expression: &model.AliasRef{Name: "R", Expression: model.ResultType(types.Any)},
				},
			},
		},
		{
			name: "Relationship and Aggregate have the same alias name",
			cql: dedent.Dedent(`
			define TESTRESULT: ({1, 2, 3}) N
			with ({4, 5}) R such that R = N
			aggregate R starting 1: R * N`),
			want: &model.Query{
				Expression: model.ResultType(types.Integer),
				Source: []*model.AliasedSource{
					{
						Alias:      "N",
						Source:     model.NewList([]string{"1", "2", "3"}, types.Integer),
						Expression: model.ResultType(&types.List{ElementType: types.Integer}),
					},
				},
				Relationship: []model.IRelationshipClause{
					&model.With{
						RelationshipClause: &model.RelationshipClause{
							Element:    &model.Element{ResultType: types.Boolean},
							Expression: model.NewList([]string{"4", "5"}, types.Integer),
							Alias:      "R",
							SuchThat: &model.Equal{
								BinaryExpression: &model.BinaryExpression{
									Expression: model.ResultType(types.Boolean),
									Operands: []model.IExpression{
										&model.AliasRef{Name: "R", Expression: model.ResultType(types.Integer)},
										&model.AliasRef{Name: "N", Expression: model.ResultType(types.Integer)},
									},
								},
							},
						},
					},
				},
				Aggregate: &model.AggregateClause{
					Element:    &model.Element{ResultType: types.Integer},
					Identifier: "R",
					Starting:   model.NewLiteral("1", types.Integer),
					Distinct:   false,
					Expression: &model.Multiply{
						BinaryExpression: &model.BinaryExpression{
							Expression: model.ResultType(types.Integer),
							Operands: []model.IExpression{
								&model.AliasRef{Name: "R", Expression: model.ResultType(types.Integer)},
								&model.AliasRef{Name: "N", Expression: model.ResultType(types.Integer)},
							},
						},
					},
				},
			},
		},
		{
			name: "Query with Return",
			cql:  `define TESTRESULT: [Observation] o return o`,
			want: &model.Query{
				Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}}),
				Source: []*model.AliasedSource{
					{
						Alias: "o",
						Source: &model.Retrieve{
							Expression:   model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}}),
							DataType:     "{http://hl7.org/fhir}Observation",
							TemplateID:   "http://hl7.org/fhir/StructureDefinition/Observation",
							CodeProperty: "code",
						},
						Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}}),
					},
				},
				Return: &model.ReturnClause{
					Expression: &model.AliasRef{Name: "o", Expression: model.ResultType(&types.Named{TypeName: "FHIR.Observation"})},
					Element:    &model.Element{ResultType: &types.Named{TypeName: "FHIR.Observation"}},
					Distinct:   true,
				},
			},
		},
		{
			name: "Query start with implicit interval",
			cql:  `define TESTRESULT: [Encounter] E return start of E.period`,
			want: &model.Query{
				Expression: model.ResultType(&types.List{ElementType: types.DateTime}),
				Source: []*model.AliasedSource{
					{
						Alias: "E",
						Source: &model.Retrieve{
							Expression:   model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Encounter"}}),
							DataType:     "{http://hl7.org/fhir}Encounter",
							TemplateID:   "http://hl7.org/fhir/StructureDefinition/Encounter",
							CodeProperty: "type",
						},
						Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Encounter"}}),
					},
				},
				Return: &model.ReturnClause{
					Element: &model.Element{ResultType: types.DateTime},
					Expression: &model.Start{
						UnaryExpression: &model.UnaryExpression{
							Expression: model.ResultType(types.DateTime),
							Operand: &model.FunctionRef{
								Expression:  model.ResultType(&types.Interval{PointType: types.DateTime}),
								Name:        "ToInterval",
								LibraryName: "FHIRHelpers",
								Operands: []model.IExpression{
									&model.Property{
										Expression: model.ResultType(&types.Named{TypeName: "FHIR.Period"}),
										Source:     &model.AliasRef{Name: "E", Expression: model.ResultType(&types.Named{TypeName: "FHIR.Encounter"})},
										Path:       "period",
									},
								},
							},
						},
					},
					Distinct: true,
				},
			},
		},
		{
			name: "Query with All Return",
			cql:  `define TESTRESULT: [Observation] o return all 4`,
			want: &model.Query{
				Expression: model.ResultType(&types.List{ElementType: types.Integer}),
				Source: []*model.AliasedSource{
					{
						Alias: "o",
						Source: &model.Retrieve{
							Expression:   model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}}),
							DataType:     "{http://hl7.org/fhir}Observation",
							TemplateID:   "http://hl7.org/fhir/StructureDefinition/Observation",
							CodeProperty: "code",
						},
						Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}}),
					},
				},
				Return: &model.ReturnClause{
					Expression: model.NewLiteral("4", types.Integer),
					Element:    &model.Element{ResultType: types.Integer},
					Distinct:   false,
				},
			},
		},
		{
			name: "Query with Return, Source is not List",
			cql:  `define TESTRESULT: (5) o return all 4`,
			want: &model.Query{
				Expression: model.ResultType(types.Integer),
				Source: []*model.AliasedSource{
					{
						Alias:      "o",
						Source:     model.NewLiteral("5", types.Integer),
						Expression: model.ResultType(types.Integer),
					},
				},
				Return: &model.ReturnClause{
					Expression: model.NewLiteral("4", types.Integer),
					Element:    &model.Element{ResultType: types.Integer},
					Distinct:   false,
				},
			},
		},
		{
			name: "Query with Where Wrapped FHIRHelpers ToBoolean",
			cql:  `define TESTRESULT: [Patient] P where P.active`,
			want: &model.Query{
				Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Patient"}}),
				Source: []*model.AliasedSource{
					{
						Alias:      "P",
						Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Patient"}}),
						Source: &model.Retrieve{
							DataType:     "{http://hl7.org/fhir}Patient",
							TemplateID:   "http://hl7.org/fhir/StructureDefinition/Patient",
							CodeProperty: "",
							Expression:   model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Patient"}}),
						},
					},
				},
				Where: &model.FunctionRef{
					Expression:  &model.Expression{Element: &model.Element{ResultType: types.Boolean}},
					Name:        "ToBoolean",
					LibraryName: "FHIRHelpers",
					Operands: []model.IExpression{
						&model.Property{
							Source:     &model.AliasRef{Name: "P", Expression: model.ResultType(&types.Named{TypeName: "FHIR.Patient"})},
							Path:       "active",
							Expression: model.ResultType(&types.Named{TypeName: "FHIR.boolean"}),
						},
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cqlLib := dedent.Dedent(fmt.Sprintf(`
			valueset "Blood pressure": 'https://test/file1'
			using FHIR version '4.0.1'
			include FHIRHelpers version '4.0.1' called FHIRHelpers
			context Patient
			%v`, test.cql))

			libs := addFHIRHelpersLib(t, cqlLib)

			parsedLibs, err := newFHIRParser(t).Libraries(context.Background(), libs, Config{})
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(test.want, getTESTRESULTModel(t, parsedLibs)); diff != "" {
				t.Errorf("Parse diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestQuery_Errors(t *testing.T) {
	tests := []struct {
		name        string
		cql         string
		errContains []string
		errCount    int
	}{
		{
			name:        "Multi-source queries must start with from",
			cql:         "[Patient] P, [Observation] O",
			errContains: []string{"for multi-source queries the keyword from is required"},
			errCount:    1,
		},
		{
			name:        "Inclusion clause not implicitly convertible to boolean",
			cql:         "[Patient] P with [Observation] O such that 'not a bool'",
			errContains: []string{"result of a query inclusion clause must be implicitly convertible to a boolean, could not convert System.String to boolean"},
			errCount:    1,
		},
		{
			name:        "Where clause not implicitly convertible to boolean",
			cql:         "[Patient] P where 'not a bool'",
			errContains: []string{"result of a where clause must be implicitly convertible to a boolean, could not convert System.String to boolean"},
			errCount:    1,
		},
		{
			name:        "Source and Let alias have same name",
			cql:         "[Patient] P let P: 4",
			errContains: []string{"alias P already exists"},
			errCount:    1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := newFHIRParser(t).Libraries(context.Background(), wrapInLib(t, test.cql), Config{})
			if err == nil {
				t.Fatalf("Parsing succeeded, expected error")
			}

			var pe *LibraryErrors
			if ok := errors.As(err, &pe); ok {
				for _, ec := range test.errContains {
					if !strings.Contains(pe.Error(), ec) {
						t.Errorf("Returned error (%s) did not contain expected string (%s)",
							pe.Error(), ec)
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
