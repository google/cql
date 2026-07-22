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
	"time"

	"github.com/google/cql/interpreter"
	"github.com/google/cql/model"
	"github.com/google/cql/parser"
	"github.com/google/cql/result"
	"github.com/google/cql/types"
	"github.com/google/go-cmp/cmp"
	"github.com/lithammer/dedent"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestQuery(t *testing.T) {
	tests := []struct {
		name                 string
		cql                  string
		wantModel            model.IExpression
		wantResult           result.Value
		wantSourceExpression model.IExpression
		wantSourceValues     []result.Value
	}{
		{
			name: "Without where",
			cql: dedent.Dedent(`
			using FHIR version '4.0.1'
			define TESTRESULT: [Encounter] E`),
			wantModel: &model.Query{
				Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Encounter"}}),
				Source: []*model.AliasedSource{
					{
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
			wantResult: newOrFatal(t, result.List{
				Value: []result.Value{
					newOrFatal(t, result.Named{Value: RetrieveFHIRResource(t, "Encounter", "1"), RuntimeType: &types.Named{TypeName: "FHIR.Encounter"}}),
					newOrFatal(t, result.Named{Value: RetrieveFHIRResource(t, "Encounter", "2"), RuntimeType: &types.Named{TypeName: "FHIR.Encounter"}}),
				},
				StaticType: &types.List{ElementType: &types.Named{TypeName: "FHIR.Encounter"}},
			}),
			wantSourceExpression: &model.Query{
				Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Encounter"}}),
				Source: []*model.AliasedSource{
					{
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
					newOrFatal(t, result.Named{Value: RetrieveFHIRResource(t, "Encounter", "1"), RuntimeType: &types.Named{TypeName: "FHIR.Encounter"}}),
					newOrFatal(t, result.Named{Value: RetrieveFHIRResource(t, "Encounter", "2"), RuntimeType: &types.Named{TypeName: "FHIR.Encounter"}}),
				},
					StaticType: &types.List{ElementType: &types.Named{TypeName: "FHIR.Encounter"}},
				}),
			},
		},
		{
			name: "Let",
			cql:  "define TESTRESULT: ({1, 2, 3}) A let B: 4, C: 5 return A + B + C",
			wantResult: newOrFatal(t, result.List{
				Value: []result.Value{
					newOrFatal(t, 10),
					newOrFatal(t, 11),
					newOrFatal(t, 12),
				},
				StaticType: &types.List{ElementType: types.Integer},
			}),
		},
		{
			name: "Let list is not unpacked",
			cql:  "define TESTRESULT: ({1, 2, 3}) A let B: {4, 5} return A + First(B)",
			wantResult: newOrFatal(t, result.List{
				Value: []result.Value{
					newOrFatal(t, 5),
					newOrFatal(t, 6),
					newOrFatal(t, 7),
				},
				StaticType: &types.List{ElementType: types.Integer},
			}),
		},
		{
			name: "Nested Let",
			cql:  "define TESTRESULT: ({1, 2}) A let B: 3 return (4) C let D: 5 return D + B",
			wantResult: newOrFatal(t, result.List{
				Value: []result.Value{
					newOrFatal(t, 8),
				},
				StaticType: &types.List{ElementType: types.Integer},
			}),
		},
		{
			name: "With where",
			cql: dedent.Dedent(`
			using FHIR version '4.0.1'
			include FHIRHelpers version '4.0.1' called FHIRHelpers

			define TESTRESULT: [Observation] O where O.id = '1'`),
			wantResult: newOrFatal(t, result.List{
				Value: []result.Value{
					newOrFatal(t, result.Named{Value: RetrieveFHIRResource(t, "Observation", "1"), RuntimeType: &types.Named{TypeName: "FHIR.Observation"}}),
				},
				StaticType: &types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}},
			}),
		},
		{
			name: "Where filters everything",
			cql: dedent.Dedent(`
			using FHIR version '4.0.1'
			include FHIRHelpers version '4.0.1' called FHIRHelpers

			define TESTRESULT: [Observation] O where O.id = 'apple'`),
			wantResult: newOrFatal(t, result.List{Value: []result.Value{}, StaticType: &types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}}}),
		},
		{
			name: "Where filters everything by date",
			cql: dedent.Dedent(`
			using FHIR version '4.0.1'
			include FHIRHelpers version '4.0.1' called FHIRHelpers

			define TESTRESULT: [Observation] O where O.effective < @1980-01-01`),
			wantResult: newOrFatal(t, result.List{Value: []result.Value{}, StaticType: &types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}}}),
		},
		{
			name: "Where returns null",
			cql: dedent.Dedent(`
			using FHIR version '4.0.1'
			define TESTRESULT: [Observation] O where null`),
			wantResult: newOrFatal(t, result.List{Value: []result.Value{}, StaticType: &types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}}}),
		},
		{
			name: "Where filters by date",
			cql: dedent.Dedent(`
			using FHIR version '4.0.1'
			include FHIRHelpers version '4.0.1' called FHIRHelpers

			define TESTRESULT: [Encounter] E where start of E.period < @2020-01-01
			`),
			wantResult: newOrFatal(t, result.List{
				Value: []result.Value{
					newOrFatal(t, result.Named{Value: RetrieveFHIRResource(t, "Encounter", "1"), RuntimeType: &types.Named{TypeName: "FHIR.Encounter"}}),
				},
				StaticType: &types.List{ElementType: &types.Named{TypeName: "FHIR.Encounter"}},
			}),
		},
		{
			name: "Where filters by all results",
			cql: dedent.Dedent(`
			using FHIR version '4.0.1'
			include FHIRHelpers version '4.0.1' called FHIRHelpers

			define TESTRESULT: [Encounter] E where start of E.period < @1999-01-01
			`),
			wantResult: newOrFatal(t, result.List{
				Value:      []result.Value{},
				StaticType: &types.List{ElementType: &types.Named{TypeName: "FHIR.Encounter"}},
			}),
		},
		{
			name: "Query retrieves empy list",
			cql: dedent.Dedent(`
			using FHIR version '4.0.1'
			define TESTRESULT: [Procedure] P`),
			wantResult: newOrFatal(t, result.List{Value: []result.Value{}, StaticType: &types.List{ElementType: &types.Named{TypeName: "FHIR.Procedure"}}}),
		},
		{
			name: "Sort descending",
			cql:  "define TESTRESULT: ({@2013-01-02T00:00:00.000Z, @2014-01-02T00:00:00.000Z, @2015-01-02T00:00:00.000Z}) l sort desc",
			wantResult: newOrFatal(t, result.List{Value: []result.Value{
				newOrFatal(t, result.DateTime{Date: time.Date(2015, time.January, 2, 0, 0, 0, 0, time.UTC), Precision: model.MILLISECOND}),
				newOrFatal(t, result.DateTime{Date: time.Date(2014, time.January, 2, 0, 0, 0, 0, time.UTC), Precision: model.MILLISECOND}),
				newOrFatal(t, result.DateTime{Date: time.Date(2013, time.January, 2, 0, 0, 0, 0, time.UTC), Precision: model.MILLISECOND}),
			},
				StaticType: &types.List{ElementType: types.DateTime}}),
		},
		{
			name: "Sort ascending",
			cql:  "define TESTRESULT: ({@2013-01-02T00:00:00.000Z, @2015-01-02T00:00:00.000Z, @2014-01-02T00:00:00.000Z}) l sort ascending",
			wantResult: newOrFatal(t, result.List{Value: []result.Value{
				newOrFatal(t, result.DateTime{Date: time.Date(2013, time.January, 2, 0, 0, 0, 0, time.UTC), Precision: model.MILLISECOND}),
				newOrFatal(t, result.DateTime{Date: time.Date(2014, time.January, 2, 0, 0, 0, 0, time.UTC), Precision: model.MILLISECOND}),
				newOrFatal(t, result.DateTime{Date: time.Date(2015, time.January, 2, 0, 0, 0, 0, time.UTC), Precision: model.MILLISECOND}),
			},
				StaticType: &types.List{ElementType: types.DateTime}}),
		},
		{
			name: "Sort descending date",
			cql:  "define TESTRESULT: ({@2013-01-02, @2014-01-02, @2015-01-02}) l sort desc",
			wantResult: newOrFatal(t, result.List{Value: []result.Value{
				newOrFatal(t, result.Date{Date: time.Date(2015, time.January, 2, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.DAY}),
				newOrFatal(t, result.Date{Date: time.Date(2014, time.January, 2, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.DAY}),
				newOrFatal(t, result.Date{Date: time.Date(2013, time.January, 2, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.DAY}),
			},
				StaticType: &types.List{ElementType: types.Date}}),
		},
		{
			name: "Sort ascending date",
			cql:  "define TESTRESULT: ({@2013-01-02, @2015-01-02, @2014-01-02}) l sort ascending",
			wantResult: newOrFatal(t, result.List{Value: []result.Value{
				newOrFatal(t, result.Date{Date: time.Date(2013, time.January, 2, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.DAY}),
				newOrFatal(t, result.Date{Date: time.Date(2014, time.January, 2, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.DAY}),
				newOrFatal(t, result.Date{Date: time.Date(2015, time.January, 2, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.DAY}),
			},
				StaticType: &types.List{ElementType: types.Date}}),
		},
		{
			name: "Sort descending int",
			cql:  "define TESTRESULT: ({1, 3, 2}) l sort desc",
			wantResult: newOrFatal(t, result.List{Value: []result.Value{
				newOrFatal(t, 3),
				newOrFatal(t, 2),
				newOrFatal(t, 1),
			},
				StaticType: &types.List{ElementType: types.Integer}}),
		},
		{
			name: "Sort ascending int",
			cql:  "define TESTRESULT: ({1, 3, 2}) l sort ascending",
			wantResult: newOrFatal(t, result.List{Value: []result.Value{
				newOrFatal(t, 1),
				newOrFatal(t, 2),
				newOrFatal(t, 3),
			},
				StaticType: &types.List{ElementType: types.Integer}}),
		},
		{
			name: "Sort descending decimal",
			cql:  "define TESTRESULT: ({1.3, 3.2, 2.1}) l sort desc",
			wantResult: newOrFatal(t, result.List{Value: []result.Value{
				newOrFatal(t, 3.2),
				newOrFatal(t, 2.1),
				newOrFatal(t, 1.3),
			},
				StaticType: &types.List{ElementType: types.Decimal}}),
		},
		{
			name: "Sort ascending decimal",
			cql:  "define TESTRESULT: ({1.3, 3.2, 2.1}) l sort ascending",
			wantResult: newOrFatal(t, result.List{Value: []result.Value{
				newOrFatal(t, 1.3),
				newOrFatal(t, 2.1),
				newOrFatal(t, 3.2),
			},
				StaticType: &types.List{ElementType: types.Decimal}}),
		},
		{
			name: "Sort descending string",
			cql:  "define TESTRESULT: ({'apple', 'Bat', 'cat', 'Dog'}) l sort desc",
			wantResult: newOrFatal(t, result.List{Value: []result.Value{
				newOrFatal(t, "cat"),
				newOrFatal(t, "apple"),
				newOrFatal(t, "Dog"),
				newOrFatal(t, "Bat"),
			},
				StaticType: &types.List{ElementType: types.String}}),
		},
		{
			name: "Sort ascending string",
			cql:  "define TESTRESULT: ({'Dog', 'cat', 'Bat', 'apple'}) l sort ascending",
			wantResult: newOrFatal(t, result.List{Value: []result.Value{
				newOrFatal(t, "Bat"),
				newOrFatal(t, "Dog"),
				newOrFatal(t, "apple"),
				newOrFatal(t, "cat"),
			},
				StaticType: &types.List{ElementType: types.String}}),
		},
		{
			name: "Sort by field",
			cql: dedent.Dedent(`
			  using FHIR version '4.0.1'
			  include FHIRHelpers version '4.0.1' called FHIRHelpers
			  define TESTRESULT: [Observation] O sort by effective`),
			wantResult: newOrFatal(t, result.List{
				Value: []result.Value{
					newOrFatal(t, result.Named{Value: RetrieveFHIRResource(t, "Observation", "1"), RuntimeType: &types.Named{TypeName: "FHIR.Observation"}}),
					newOrFatal(t, result.Named{Value: RetrieveFHIRResource(t, "Observation", "2"), RuntimeType: &types.Named{TypeName: "FHIR.Observation"}}),
					newOrFatal(t, result.Named{Value: RetrieveFHIRResource(t, "Observation", "3"), RuntimeType: &types.Named{TypeName: "FHIR.Observation"}}),
				},
				StaticType: &types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}},
			}),
		},
		{
			name: "Sort by field desc",
			cql: dedent.Dedent(`
			  using FHIR version '4.0.1'
			  include FHIRHelpers version '4.0.1' called FHIRHelpers
			  define TESTRESULT: [Observation] O sort by effective desc`),
			wantResult: newOrFatal(t, result.List{
				Value: []result.Value{
					newOrFatal(t, result.Named{Value: RetrieveFHIRResource(t, "Observation", "3"), RuntimeType: &types.Named{TypeName: "FHIR.Observation"}}),
					newOrFatal(t, result.Named{Value: RetrieveFHIRResource(t, "Observation", "2"), RuntimeType: &types.Named{TypeName: "FHIR.Observation"}}),
					newOrFatal(t, result.Named{Value: RetrieveFHIRResource(t, "Observation", "1"), RuntimeType: &types.Named{TypeName: "FHIR.Observation"}}),
				},
				StaticType: &types.List{ElementType: &types.Named{TypeName: "FHIR.Observation"}},
			}),
		},
		// Online test: https://cql-runner.dataphoria.org/
		{
			name: "Sort by expression",
			cql: dedent.Dedent(`
			  using FHIR version '4.0.1'
			  include FHIRHelpers version '4.0.1' called FHIRHelpers
			  define TESTRESULT: [Encounter] E sort by start of period`),
			wantResult: newOrFatal(t, result.List{
				Value: []result.Value{
					newOrFatal(t, result.Named{Value: RetrieveFHIRResource(t, "Encounter", "1"), RuntimeType: &types.Named{TypeName: "FHIR.Encounter"}}),
					newOrFatal(t, result.Named{Value: RetrieveFHIRResource(t, "Encounter", "2"), RuntimeType: &types.Named{TypeName: "FHIR.Encounter"}}),
				},
				StaticType: &types.List{ElementType: &types.Named{TypeName: "FHIR.Encounter"}},
			}),
		},
		{
			name: "Sort by expression expression descending",
			cql: dedent.Dedent(`
			  using FHIR version '4.0.1'
			  include FHIRHelpers version '4.0.1' called FHIRHelpers
			  define TESTRESULT: [Encounter] E sort by start of period desc`),
			wantResult: newOrFatal(t, result.List{
				Value: []result.Value{
					newOrFatal(t, result.Named{Value: RetrieveFHIRResource(t, "Encounter", "2"), RuntimeType: &types.Named{TypeName: "FHIR.Encounter"}}),
					newOrFatal(t, result.Named{Value: RetrieveFHIRResource(t, "Encounter", "1"), RuntimeType: &types.Named{TypeName: "FHIR.Encounter"}}),
				},
				StaticType: &types.List{ElementType: &types.Named{TypeName: "FHIR.Encounter"}},
			}),
		},
		{
			name:       "Aggregate",
			cql:        "define TESTRESULT: ({1, 2, 3, 3, 4}) L aggregate A starting 1: A * L",
			wantResult: newOrFatal(t, 72),
		},
		{
			name:       "Aggregate all",
			cql:        "define TESTRESULT: ({1, 2, 3, 3, 4}) L aggregate all A starting 1: A * L",
			wantResult: newOrFatal(t, 72),
		},
		{
			name:       "Aggregate distinct",
			cql:        "define TESTRESULT: ({1, 2, 3, 3, 4}) L aggregate distinct A starting 1: A * L",
			wantResult: newOrFatal(t, 24),
		},
		{
			name:       "Aggregate no starting expression",
			cql:        "define TESTRESULT: ({1, 2, 3}) L aggregate A : A * L",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Aggregate with multi-source",
			cql:        "define TESTRESULT: from ({1, 2, 3}) B, (4) C aggregate A : A + B + C",
			wantResult: newOrFatal(t, nil),
		},
		{
			name: "List query with all return",
			cql:  "define TESTRESULT: ({2, 2, 3}) l return all (l*2)",
			wantModel: &model.Query{
				Expression: model.ResultType(&types.List{ElementType: types.Integer}),
				Source: []*model.AliasedSource{
					{
						Alias: "l",
						Source: &model.List{
							List: []model.IExpression{
								model.NewLiteral("2", types.Integer),
								model.NewLiteral("2", types.Integer),
								model.NewLiteral("3", types.Integer),
							},
							Expression: model.ResultType(&types.List{ElementType: types.Integer}),
						},
						Expression: model.ResultType(&types.List{ElementType: types.Integer}),
					},
				},
				Return: &model.ReturnClause{
					Element: &model.Element{ResultType: types.Integer},
					Expression: &model.Multiply{
						BinaryExpression: &model.BinaryExpression{
							Expression: model.ResultType(types.Integer),
							Operands: []model.IExpression{
								&model.AliasRef{
									Name:       "l",
									Expression: model.ResultType(types.Integer),
								},
								model.NewLiteral("2", types.Integer),
							},
						},
					},
				},
			},
			wantResult: newOrFatal(t, result.List{
				Value:      []result.Value{newOrFatal(t, 4), newOrFatal(t, 4), newOrFatal(t, 6)},
				StaticType: &types.List{ElementType: types.Integer},
			}),
		},
		{
			name: "List query with distinct return",
			cql:  "define TESTRESULT: ({2, 2, 3}) l return (l*2)",
			wantResult: newOrFatal(t, result.List{
				Value:      []result.Value{newOrFatal(t, 4), newOrFatal(t, 6)},
				StaticType: &types.List{ElementType: types.Integer},
			}),
		},
		{
			name: "List query with where clause and all return",
			cql:  "define TESTRESULT: ({2, 2, 3}) l where l > 2 return all (l*2)",
			wantResult: newOrFatal(t, result.List{
				Value:      []result.Value{newOrFatal(t, 6)},
				StaticType: &types.List{ElementType: types.Integer},
			}),
		},
		{
			name: "Empty list query with return",
			cql:  "define TESTRESULT: (List<Integer>{}) l return all (l*2)",
			wantResult: newOrFatal(t, result.List{
				Value:      []result.Value{},
				StaticType: &types.List{ElementType: types.Integer},
			}),
		},
		{
			name: "List query that has a different result type than the input list",
			cql:  "define TESTRESULT: ({2, 3}) l return all ToDecimal(l)",
			wantResult: newOrFatal(t, result.List{
				Value:      []result.Value{newOrFatal(t, 2.0), newOrFatal(t, 3.0)},
				StaticType: &types.List{ElementType: types.Decimal},
			}),
		},
		{
			name:       "Non list source",
			cql:        "define TESTRESULT: (4) l",
			wantResult: newOrFatal(t, 4),
		},
		{
			name:       "Non list source with return",
			cql:        "define TESTRESULT: (4) l return 'hi'",
			wantResult: newOrFatal(t, "hi"),
		},
		{
			name: "Multiple sources",
			cql:  "define TESTRESULT: from ({2, 3}) A, ({5, 6}) B, (7) C",
			wantResult: newOrFatal(t, result.List{
				Value: []result.Value{
					newOrFatal(t, result.Tuple{
						Value:       map[string]result.Value{"A": newOrFatal(t, 2), "B": newOrFatal(t, 5), "C": newOrFatal(t, 7)},
						RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{"A": types.Integer, "B": types.Integer, "C": types.Integer}},
					}),
					newOrFatal(t, result.Tuple{
						Value:       map[string]result.Value{"A": newOrFatal(t, 2), "B": newOrFatal(t, 6), "C": newOrFatal(t, 7)},
						RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{"A": types.Integer, "B": types.Integer, "C": types.Integer}},
					}),
					newOrFatal(t, result.Tuple{
						Value:       map[string]result.Value{"A": newOrFatal(t, 3), "B": newOrFatal(t, 5), "C": newOrFatal(t, 7)},
						RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{"A": types.Integer, "B": types.Integer, "C": types.Integer}},
					}),
					newOrFatal(t, result.Tuple{
						Value:       map[string]result.Value{"A": newOrFatal(t, 3), "B": newOrFatal(t, 6), "C": newOrFatal(t, 7)},
						RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{"A": types.Integer, "B": types.Integer, "C": types.Integer}},
					}),
				},
				StaticType: &types.List{ElementType: &types.Tuple{ElementTypes: map[string]types.IType{"A": types.Integer, "B": types.Integer, "C": types.Integer}}},
			}),
		},
		{
			name: "Multiple sources all non lists",
			cql:  "define TESTRESULT: from (3) A, (5) B, (7) C",
			wantResult: newOrFatal(t, result.Tuple{
				Value:       map[string]result.Value{"A": newOrFatal(t, 3), "B": newOrFatal(t, 5), "C": newOrFatal(t, 7)},
				RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{"A": types.Integer, "B": types.Integer, "C": types.Integer}},
			}),
		},
		{
			name: "Multiple sources with return",
			cql:  "define TESTRESULT: from ({2, 3}) A, ({5, 6, 7}) B return all A",
			wantResult: newOrFatal(t, result.List{
				Value: []result.Value{
					newOrFatal(t, 2),
					newOrFatal(t, 2),
					newOrFatal(t, 2),
					newOrFatal(t, 3),
					newOrFatal(t, 3),
					newOrFatal(t, 3),
				},
				StaticType: &types.List{ElementType: types.Integer},
			}),
		},
		{
			name: "Multiple sources mixed list",
			cql:  "define TESTRESULT: from ({2, 'hi'}) A, ({5, 6, 7}) B return all A",
			wantResult: newOrFatal(t, result.List{
				Value: []result.Value{
					newOrFatal(t, 2),
					newOrFatal(t, 2),
					newOrFatal(t, 2),
					newOrFatal(t, "hi"),
					newOrFatal(t, "hi"),
					newOrFatal(t, "hi"),
				},
				StaticType: &types.List{ElementType: &types.Choice{ChoiceTypes: []types.IType{types.Integer, types.String}}},
			}),
		},
		{
			name: "Multiple sources mixed list",
			cql:  "define TESTRESULT: from ({2, 'hi'}) A, ({5, 6, 7}) B return all A",
			wantResult: newOrFatal(t, result.List{
				Value: []result.Value{
					newOrFatal(t, 2),
					newOrFatal(t, 2),
					newOrFatal(t, 2),
					newOrFatal(t, "hi"),
					newOrFatal(t, "hi"),
					newOrFatal(t, "hi"),
				},
				StaticType: &types.List{ElementType: &types.Choice{ChoiceTypes: []types.IType{types.Integer, types.String}}},
			}),
		},
		{
			name: "Relationship clause with",
			cql:  "define TESTRESULT: ({1, 2, 3, 4}) A with ({2, 3}) B such that A + B >= 5",
			wantResult: newOrFatal(t, result.List{
				Value: []result.Value{
					newOrFatal(t, 2),
					newOrFatal(t, 3),
					newOrFatal(t, 4),
				},
				StaticType: &types.List{ElementType: types.Integer},
			}),
		},
		{
			name: "Relationship clause without",
			cql:  "define TESTRESULT: ({1, 2, 3, 4}) A without ({2, 3}) B such that A + B >= 5",
			wantResult: newOrFatal(t, result.List{
				Value: []result.Value{
					newOrFatal(t, 1),
					newOrFatal(t, 2),
				},
				StaticType: &types.List{ElementType: types.Integer},
			}),
		},
		{
			name: "Multiple relationship clauses",
			cql: dedent.Dedent(`
			define TESTRESULT: ({1, 2, 3, 4}) A
			with ({2, 3}) B such that A + B >= 5
			with ({2, 3}) C such that A = 4 or A = 1`),
			wantResult: newOrFatal(t, result.List{
				Value: []result.Value{
					newOrFatal(t, 4),
				},
				StaticType: &types.List{ElementType: types.Integer},
			}),
		},
		{
			name: "Relationship clause non list source",
			cql:  "define TESTRESULT: ({1, 2, 3, 4}) A with (2) B such that A + B >= 5",
			wantResult: newOrFatal(t, result.List{
				Value: []result.Value{
					newOrFatal(t, 3),
					newOrFatal(t, 4),
				},
				StaticType: &types.List{ElementType: types.Integer},
			}),
		},
		{
			name: "Relationship clause on multi-source query",
			cql:  "define TESTRESULT: from ({1, 2, 3}) A, ({4, 5}) B with ({2, 3}) C such that A + B + C >= 10",
			wantResult: newOrFatal(t, result.List{
				Value: []result.Value{
					newOrFatal(t, result.Tuple{
						Value:       map[string]result.Value{"A": newOrFatal(t, 2), "B": newOrFatal(t, 5)},
						RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{"A": types.Integer, "B": types.Integer}},
					}),
					newOrFatal(t, result.Tuple{
						Value:       map[string]result.Value{"A": newOrFatal(t, 3), "B": newOrFatal(t, 4)},
						RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{"A": types.Integer, "B": types.Integer}},
					}),
					newOrFatal(t, result.Tuple{
						Value:       map[string]result.Value{"A": newOrFatal(t, 3), "B": newOrFatal(t, 5)},
						RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{"A": types.Integer, "B": types.Integer}},
					}),
				},
				StaticType: &types.List{ElementType: &types.Tuple{ElementTypes: map[string]types.IType{"A": types.Integer, "B": types.Integer}}},
			}),
		},
		{
			// This ensures that properties on null values inside queries are handled correctly.
			name:       "Property on null alias in query",
			cql:        "define TESTRESULT: (null as Code) l return l.code",
			wantResult: newOrFatal(t, nil),
		},
		{
			// Test case for the "could not resolve the local reference to M" bug fix
			name: "Alias reference in where clause",
			cql: dedent.Dedent(`
			using FHIR version '4.0.1'
			include FHIRHelpers version '4.0.1' called FHIRHelpers

			define TESTRESULT: [Encounter] M where M.id = '1'`),
			wantResult: newOrFatal(t, result.List{
				Value: []result.Value{
					newOrFatal(t, result.Named{Value: RetrieveFHIRResource(t, "Encounter", "1"), RuntimeType: &types.Named{TypeName: "FHIR.Encounter"}}),
				},
				StaticType: &types.List{ElementType: &types.Named{TypeName: "FHIR.Encounter"}},
			}),
		},
		{
			// Test case for function alias resolution - reproduces the CDS_Connect_Commons issue
			name: "Function with query alias",
			cql: dedent.Dedent(`
			define function TestFunction(IntList List<Integer>):
			  IntList M
			    where M > 5

			define TESTRESULT: TestFunction({1, 6, 3, 8})`),
			wantResult: newOrFatal(t, result.List{
				Value: []result.Value{
					newOrFatal(t, 6),
					newOrFatal(t, 8),
				},
				StaticType: &types.List{ElementType: types.Integer},
			}),
		},
		{
			// Test case for function with let clause and alias - closer to CDS_Connect_Commons pattern
			name: "Function with let clause and query alias",
			cql: dedent.Dedent(`
			define function TestFunctionWithLet(IntList List<Integer>):
			  IntList M
			    let Threshold: 5
			    where M > Threshold

			define TESTRESULT: TestFunctionWithLet({1, 6, 3, 8})`),
			wantResult: newOrFatal(t, result.List{
				Value: []result.Value{
					newOrFatal(t, 6),
					newOrFatal(t, 8),
				},
				StaticType: &types.List{ElementType: types.Integer},
			}),
		},
		{
			// Test case for multiple function calls with same alias - reproduces the real CDS_Connect_Commons issue
			name: "Multiple functions with same alias name",
			cql: dedent.Dedent(`
			define function ActiveMedicationStatement(MedList List<Integer>):
			  MedList M
			    where M > 5

			define function ActiveMedicationRequest(MedList List<Integer>):
			  MedList M
			    where M > 3

			define TESTRESULT: 
			  exists(ActiveMedicationStatement({1, 6, 3, 8}))
			  or exists(ActiveMedicationRequest({1, 2, 4, 7}))`),
			wantResult: newOrFatal(t, true),
		},
		{
			// Test case that reproduces the exact real-world error with FHIR retrievals and cross-library calls
			name: "Cross library function calls with FHIR retrievals",
			cql: dedent.Dedent(`
			library TestCommons version '1.0.0'
			using FHIR version '4.0.1'

			define function ActiveMedicationStatement(MedList List<MedicationStatement>):
			  MedList M
			    where M.status.value = 'active'

			define function ActiveMedicationRequest(MedList List<MedicationRequest>):
			  MedList M
			    where M.status.value = 'active'
			
			define TESTRESULT:
			  exists(ActiveMedicationStatement([MedicationStatement]))
			  or exists(ActiveMedicationRequest([MedicationRequest]))`),
			wantResult: newOrFatal(t, false), // Both retrievals return empty lists, so exists() returns false
		},
		{
			// Test case for let clause accessing source alias - should fail with current implementation
			name: "Let clause with source alias reference",
			cql: dedent.Dedent(`
			define function TestLetSourceAlias(nums List<Integer>):
			  nums N
			    let threshold: 5
			    where N > threshold

			define TESTRESULT: TestLetSourceAlias({1, 10, 15, 20})`),
			wantResult: newOrFatal(t, result.List{
				Value: []result.Value{
					newOrFatal(t, 10),
					newOrFatal(t, 15),
					newOrFatal(t, 20),
				}, // Values where N > 5
				StaticType: &types.List{ElementType: types.Integer},
			}),
		},
		{
			// Test case for relationship clause with alias access
			name: "Relationship clause with alias access",
			cql: dedent.Dedent(`
			define TESTRESULT: 
			  ({1, 2, 3}) A 
			    with ({4, 5}) B 
			      such that A + B > 6
			    return A`),
			wantResult: newOrFatal(t, result.List{
				Value: []result.Value{
					newOrFatal(t, 2), // 2+4=6, 2+5=7 > 6
					newOrFatal(t, 3), // 3+4=7, 3+5=8 > 6
				},
				StaticType: &types.List{ElementType: types.Integer},
			}),
		},
		{
			// Test case for return clause accessing let variables
			name: "Return clause accessing let variables",
			cql: dedent.Dedent(`
			define function TestReturnLet(nums List<Integer>):
			  nums N
			    let factor: 10
			    return N * factor

			define TESTRESULT: TestReturnLet({1, 2, 3})`),
			wantResult: newOrFatal(t, result.List{
				Value: []result.Value{
					newOrFatal(t, 10),
					newOrFatal(t, 20),
					newOrFatal(t, 30),
				},
				StaticType: &types.List{ElementType: types.Integer},
			}),
		},
		{
			// Test case for aggregate clause with source alias
			name: "Aggregate clause with source alias",
			cql: dedent.Dedent(`
			define function TestAggregateAlias(nums List<Integer>):
			  nums N
			    aggregate A starting 0: A + N

			define TESTRESULT: TestAggregateAlias({1, 2, 3})`),
			wantResult: newOrFatal(t, 6), // Sum of 1+2+3
		},
		{
			// Test case for complex multi-clause with all scope issues
			name: "Complex multi-clause with all scope management",
			cql: dedent.Dedent(`
			define function TestAllClauses(nums List<Integer>):
			  nums N
			    let threshold: 5,
			        multiplier: 2
			    with ({1, 2}) H
			      such that N + H > threshold
			    where N > threshold
			    return N * multiplier

			define TESTRESULT: TestAllClauses({1, 3, 6, 8, 10})`),
			wantResult: newOrFatal(t, result.List{
				Value: []result.Value{
					newOrFatal(t, 12), // 6*2, where 6>5 and 6+H>5 for some H
					newOrFatal(t, 16), // 8*2, where 8>5 and 8+H>5 for some H
					newOrFatal(t, 20), // 10*2, where 10>5 and 10+H>5 for some H
				},
				StaticType: &types.List{ElementType: types.Integer},
			}),
		},
		{
			// Test case that should expose scope management issues - let clause referencing source alias
			name: "Let clause referencing source alias directly",
			cql: dedent.Dedent(`
			define function TestLetWithSourceAlias(nums List<Integer>):
			  nums N
			    let doubled: N * 2
			    where doubled > 10

			define TESTRESULT: TestLetWithSourceAlias({1, 3, 6, 8})`),
			wantResult: newOrFatal(t, result.List{
				Value: []result.Value{
					newOrFatal(t, 6), // 6*2=12 > 10
					newOrFatal(t, 8), // 8*2=16 > 10
				},
				StaticType: &types.List{ElementType: types.Integer},
			}),
		},
		{
			// Test case that reproduces the NCQA_CQLBase error - sort by tuple field
			name: "Sort by tuple field from return clause",
			cql: dedent.Dedent(`
			define TESTRESULT: 
			  ({1, 3, 2}) I
			    return Tuple {
			      value: I,
			      sortKey: I * 10
			    }
			    sort by sortKey asc`),
			wantResult: newOrFatal(t, result.List{
				Value: []result.Value{
					newOrFatal(t, result.Tuple{
						Value: map[string]result.Value{
							"value":   newOrFatal(t, 1),
							"sortKey": newOrFatal(t, 10),
						},
						RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{
							"value":   types.Integer,
							"sortKey": types.Integer,
						}},
					}),
					newOrFatal(t, result.Tuple{
						Value: map[string]result.Value{
							"value":   newOrFatal(t, 2),
							"sortKey": newOrFatal(t, 20),
						},
						RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{
							"value":   types.Integer,
							"sortKey": types.Integer,
						}},
					}),
					newOrFatal(t, result.Tuple{
						Value: map[string]result.Value{
							"value":   newOrFatal(t, 3),
							"sortKey": newOrFatal(t, 30),
						},
						RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{
							"value":   types.Integer,
							"sortKey": types.Integer,
						}},
					}),
				},
				StaticType: &types.List{ElementType: &types.Tuple{ElementTypes: map[string]types.IType{
					"value":   types.Integer,
					"sortKey": types.Integer,
				}}},
			}),
		},
		{
			// Test sorting by integer field in tuple
			name: "Sort by integer field ascending",
			cql: dedent.Dedent(`
			define TESTRESULT: 
			  ({3, 1, 2}) I
			    return Tuple {
			      value: I,
			      intField: I
			    }
			    sort by intField asc`),
			wantResult: newOrFatal(t, result.List{
				Value: []result.Value{
					newOrFatal(t, result.Tuple{
						Value: map[string]result.Value{
							"value":    newOrFatal(t, 1),
							"intField": newOrFatal(t, 1),
						},
						RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{
							"value":    types.Integer,
							"intField": types.Integer,
						}},
					}),
					newOrFatal(t, result.Tuple{
						Value: map[string]result.Value{
							"value":    newOrFatal(t, 2),
							"intField": newOrFatal(t, 2),
						},
						RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{
							"value":    types.Integer,
							"intField": types.Integer,
						}},
					}),
					newOrFatal(t, result.Tuple{
						Value: map[string]result.Value{
							"value":    newOrFatal(t, 3),
							"intField": newOrFatal(t, 3),
						},
						RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{
							"value":    types.Integer,
							"intField": types.Integer,
						}},
					}),
				},
				StaticType: &types.List{ElementType: &types.Tuple{ElementTypes: map[string]types.IType{
					"value":    types.Integer,
					"intField": types.Integer,
				}}},
			}),
		},
		{
			// Test sorting by integer field descending
			name: "Sort by integer field descending",
			cql: dedent.Dedent(`
			define TESTRESULT: 
			  ({1, 3, 2}) I
			    return Tuple {
			      value: I,
			      intField: I
			    }
			    sort by intField desc`),
			wantResult: newOrFatal(t, result.List{
				Value: []result.Value{
					newOrFatal(t, result.Tuple{
						Value: map[string]result.Value{
							"value":    newOrFatal(t, 3),
							"intField": newOrFatal(t, 3),
						},
						RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{
							"value":    types.Integer,
							"intField": types.Integer,
						}},
					}),
					newOrFatal(t, result.Tuple{
						Value: map[string]result.Value{
							"value":    newOrFatal(t, 2),
							"intField": newOrFatal(t, 2),
						},
						RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{
							"value":    types.Integer,
							"intField": types.Integer,
						}},
					}),
					newOrFatal(t, result.Tuple{
						Value: map[string]result.Value{
							"value":    newOrFatal(t, 1),
							"intField": newOrFatal(t, 1),
						},
						RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{
							"value":    types.Integer,
							"intField": types.Integer,
						}},
					}),
				},
				StaticType: &types.List{ElementType: &types.Tuple{ElementTypes: map[string]types.IType{
					"value":    types.Integer,
					"intField": types.Integer,
				}}},
			}),
		},
		{
			// Test sorting by decimal field
			name: "Sort by decimal field ascending",
			cql: dedent.Dedent(`
			define TESTRESULT: 
			  ({3.2, 1.1, 2.5}) D
			    return Tuple {
			      value: D,
			      decField: D
			    }
			    sort by decField asc`),
			wantResult: newOrFatal(t, result.List{
				Value: []result.Value{
					newOrFatal(t, result.Tuple{
						Value: map[string]result.Value{
							"value":    newOrFatal(t, 1.1),
							"decField": newOrFatal(t, 1.1),
						},
						RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{
							"value":    types.Decimal,
							"decField": types.Decimal,
						}},
					}),
					newOrFatal(t, result.Tuple{
						Value: map[string]result.Value{
							"value":    newOrFatal(t, 2.5),
							"decField": newOrFatal(t, 2.5),
						},
						RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{
							"value":    types.Decimal,
							"decField": types.Decimal,
						}},
					}),
					newOrFatal(t, result.Tuple{
						Value: map[string]result.Value{
							"value":    newOrFatal(t, 3.2),
							"decField": newOrFatal(t, 3.2),
						},
						RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{
							"value":    types.Decimal,
							"decField": types.Decimal,
						}},
					}),
				},
				StaticType: &types.List{ElementType: &types.Tuple{ElementTypes: map[string]types.IType{
					"value":    types.Decimal,
					"decField": types.Decimal,
				}}},
			}),
		},
		{
			// Test sorting by string field
			name: "Sort by string field ascending",
			cql: dedent.Dedent(`
			define TESTRESULT: 
			  ({'zebra', 'apple', 'banana'}) S
			    return Tuple {
			      value: S,
			      strField: S
			    }
			    sort by strField asc`),
			wantResult: newOrFatal(t, result.List{
				Value: []result.Value{
					newOrFatal(t, result.Tuple{
						Value: map[string]result.Value{
							"value":    newOrFatal(t, "apple"),
							"strField": newOrFatal(t, "apple"),
						},
						RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{
							"value":    types.String,
							"strField": types.String,
						}},
					}),
					newOrFatal(t, result.Tuple{
						Value: map[string]result.Value{
							"value":    newOrFatal(t, "banana"),
							"strField": newOrFatal(t, "banana"),
						},
						RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{
							"value":    types.String,
							"strField": types.String,
						}},
					}),
					newOrFatal(t, result.Tuple{
						Value: map[string]result.Value{
							"value":    newOrFatal(t, "zebra"),
							"strField": newOrFatal(t, "zebra"),
						},
						RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{
							"value":    types.String,
							"strField": types.String,
						}},
					}),
				},
				StaticType: &types.List{ElementType: &types.Tuple{ElementTypes: map[string]types.IType{
					"value":    types.String,
					"strField": types.String,
				}}},
			}),
		},
		{
			// Test sorting by DateTime field
			name: "Sort by DateTime field descending",
			cql: dedent.Dedent(`
			define TESTRESULT: 
			  ({@2015-01-01T10:00:00.000Z, @2013-01-01T10:00:00.000Z, @2014-01-01T10:00:00.000Z}) DT
			    return Tuple {
			      value: DT,
			      dtField: DT
			    }
			    sort by dtField desc`),
			wantResult: newOrFatal(t, result.List{
				Value: []result.Value{
					newOrFatal(t, result.Tuple{
						Value: map[string]result.Value{
							"value":   newOrFatal(t, result.DateTime{Date: time.Date(2015, time.January, 1, 10, 0, 0, 0, time.UTC), Precision: model.MILLISECOND}),
							"dtField": newOrFatal(t, result.DateTime{Date: time.Date(2015, time.January, 1, 10, 0, 0, 0, time.UTC), Precision: model.MILLISECOND}),
						},
						RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{
							"value":   types.DateTime,
							"dtField": types.DateTime,
						}},
					}),
					newOrFatal(t, result.Tuple{
						Value: map[string]result.Value{
							"value":   newOrFatal(t, result.DateTime{Date: time.Date(2014, time.January, 1, 10, 0, 0, 0, time.UTC), Precision: model.MILLISECOND}),
							"dtField": newOrFatal(t, result.DateTime{Date: time.Date(2014, time.January, 1, 10, 0, 0, 0, time.UTC), Precision: model.MILLISECOND}),
						},
						RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{
							"value":   types.DateTime,
							"dtField": types.DateTime,
						}},
					}),
					newOrFatal(t, result.Tuple{
						Value: map[string]result.Value{
							"value":   newOrFatal(t, result.DateTime{Date: time.Date(2013, time.January, 1, 10, 0, 0, 0, time.UTC), Precision: model.MILLISECOND}),
							"dtField": newOrFatal(t, result.DateTime{Date: time.Date(2013, time.January, 1, 10, 0, 0, 0, time.UTC), Precision: model.MILLISECOND}),
						},
						RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{
							"value":   types.DateTime,
							"dtField": types.DateTime,
						}},
					}),
				},
				StaticType: &types.List{ElementType: &types.Tuple{ElementTypes: map[string]types.IType{
					"value":   types.DateTime,
					"dtField": types.DateTime,
				}}},
			}),
		},
		{
			// Test sorting by Date field
			name: "Sort by Date field ascending",
			cql: dedent.Dedent(`
			define TESTRESULT: 
			  ({@2015-01-01, @2013-01-01, @2014-01-01}) D
			    return Tuple {
			      value: D,
			      dateField: D
			    }
			    sort by dateField asc`),
			wantResult: newOrFatal(t, result.List{
				Value: []result.Value{
					newOrFatal(t, result.Tuple{
						Value: map[string]result.Value{
							"value":     newOrFatal(t, result.Date{Date: time.Date(2013, time.January, 1, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.DAY}),
							"dateField": newOrFatal(t, result.Date{Date: time.Date(2013, time.January, 1, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.DAY}),
						},
						RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{
							"value":     types.Date,
							"dateField": types.Date,
						}},
					}),
					newOrFatal(t, result.Tuple{
						Value: map[string]result.Value{
							"value":     newOrFatal(t, result.Date{Date: time.Date(2014, time.January, 1, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.DAY}),
							"dateField": newOrFatal(t, result.Date{Date: time.Date(2014, time.January, 1, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.DAY}),
						},
						RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{
							"value":     types.Date,
							"dateField": types.Date,
						}},
					}),
					newOrFatal(t, result.Tuple{
						Value: map[string]result.Value{
							"value":     newOrFatal(t, result.Date{Date: time.Date(2015, time.January, 1, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.DAY}),
							"dateField": newOrFatal(t, result.Date{Date: time.Date(2015, time.January, 1, 0, 0, 0, 0, defaultEvalTimestamp.Location()), Precision: model.DAY}),
						},
						RuntimeType: &types.Tuple{ElementTypes: map[string]types.IType{
							"value":     types.Date,
							"dateField": types.Date,
						}},
					}),
				},
				StaticType: &types.List{ElementType: &types.Tuple{ElementTypes: map[string]types.IType{
					"value":     types.Date,
					"dateField": types.Date,
				}}},
			}),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := newFHIRParser(t)
			parsedLibs, err := p.Libraries(context.Background(), addFHIRHelpersLib(t, tc.cql), parser.Config{})
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
			gotResult := getTESTRESULTWithSources(t, results)
			if diff := cmp.Diff(tc.wantSourceExpression, gotResult.SourceExpression(), protocmp.Transform()); tc.wantSourceExpression != nil && diff != "" {
				t.Errorf("Eval SourceExpression diff (-want +got)\n%v", diff)
			}
			if diff := cmp.Diff(tc.wantSourceValues, gotResult.SourceValues(), protocmp.Transform()); tc.wantSourceValues != nil && diff != "" {
				t.Errorf("Eval SourceValues diff (-want +got)\n%v", diff)
			}
		})
	}
}
