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
	"fmt"
	"strings"
	"testing"

	"github.com/google/cql/internal/embeddata"
	"github.com/google/cql/model"
	"github.com/google/cql/types"
	"github.com/google/go-cmp/cmp"
	"github.com/lithammer/dedent"
)

func TestParserExpressions_Errors(t *testing.T) {
	tests := []struct {
		name        string
		cql         string
		errContains []string
		errCount    int
	}{
		{
			name:        "Invalid CQL",
			cql:         "abcdefg",
			errContains: []string{"could not resolve", "abcdefg"},
			errCount:    1,
		},
		{
			name:        "Interval Selector High And Low Type Mismatch",
			cql:         `Interval[3.0, @2015-01-01)`,
			errContains: []string{"could not resolve Interval(System.Decimal, System.Date)"},
			errCount:    1,
		},
		{
			name:        "List Selector With Type Specifier No Matching Types",
			cql:         `List<Long>{'Hello', 5}`,
			errContains: []string{"unable to convert list element (System.String) to the declared List type specifier element type (System.Long)"},
			errCount:    1,
		},
		{
			name:        "Instance Selector not implicitly convertible",
			cql:         "Quantity {value: 'wrong type', unit: 'mg'}",
			errContains: []string{`element "value" in System.Quantity should be implicitly convertible to type System.Decimal, but instead received type System.String`},
			errCount:    1,
		},
		{
			name:        "Invalid Date",
			cql:         "@2015-99",
			errContains: []string{`want a layout like @YYYY-MM-DD`},
			errCount:    1,
		},
		{
			name:        "Invalid DateTime",
			cql:         "@2015-21-13T99",
			errContains: []string{`want a layout like @YYYY-MM-DDThh:mm:ss.fff(Z|(+/-hh:mm)`},
			errCount:    1,
		},
		{
			name:        "Invalid Time",
			cql:         "@T99",
			errContains: []string{`want a layout like @Thh:mm:ss.fff`},
			errCount:    1,
		},
		{
			name:        "Instance Selector incorrect field",
			cql:         "Quantity {bogusfield: 'wrong type', unit: 'mg'}",
			errContains: []string{`property "bogusfield" not found in Parent Type "System.Quantity" property not found in data model`},
			errCount:    1,
		},
		{
			name:        "Resource not retrievable",
			cql:         `[ObservationStatus]`,
			errContains: []string{"tried to retrieve type Named<FHIR.ObservationStatus>, but this type is not retrievable"},
			errCount:    1,
		},
		{
			name:        "Retrieve unknown type",
			cql:         `[randomtype]`,
			errContains: []string{"not found in data model", "retrieves cannot be"},
			errCount:    2,
		},
		{
			name: "Case No Comparand When Not Boolean",
			cql: `
			case
			when 'Apple' then 4
			else 5
			end`,
			errContains: []string{"could not implicitly convert System.String to a System.Boolean"},
			errCount:    1,
		},
		{
			name: "Case with Comparand not Convertible",
			cql: `
			case 5
			when 'Apple' then 4
			else 5
			end`,
			errContains: []string{"could not implicitly convert then comparand System.Integer and cases System.String to the same type"},
			errCount:    1,
		},
		{
			name:        "If then statement with non boolean condition",
			cql:         `if 42 then 3 else 4`,
			errContains: []string{"could not implicitly convert"},
			errCount:    1,
		},
		{
			name:        "Where Clause Not Implicitly Convertible To Boolean",
			cql:         `({1, 2, 3}) P where P`,
			errContains: []string{"result of a where clause"},
			errCount:    1,
		},
		{
			name:        "maximum String",
			cql:         "maximum String",
			errContains: []string{"unsupported type for maximumString"},
			errCount:    1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := newFHIRParser(t).Libraries(context.Background(), wrapInLib(t, test.cql), Config{})

			if err == nil {
				t.Fatal("Parse succeeded, wanted error")
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

func TestParserExpressions(t *testing.T) {
	tests := []struct {
		name string
		desc string
		cql  string
		want model.IExpression
	}{
		{
			name: "Case No Comparand",
			cql: `
			case
			when true then 4
			else 5
			end`,
			want: &model.Case{
				Comparand: nil,
				CaseItem: []*model.CaseItem{
					&model.CaseItem{
						When: model.NewLiteral("true", types.Boolean),
						Then: model.NewLiteral("4", types.Integer),
					},
				},
				Else:       model.NewLiteral("5", types.Integer),
				Expression: model.ResultType(types.Integer),
			},
		},
		{
			name: "Case With Comparand",
			cql: `
			case 4
			when 5 then 6
			else 7
			end`,
			want: &model.Case{
				Comparand: model.NewLiteral("4", types.Integer),
				CaseItem: []*model.CaseItem{
					&model.CaseItem{
						When: model.NewLiteral("5", types.Integer),
						Then: model.NewLiteral("6", types.Integer),
					},
				},
				Else:       model.NewLiteral("7", types.Integer),
				Expression: model.ResultType(types.Integer),
			},
		},
		{
			name: "Case With Comparand Implictly Converted",
			cql: `
			case 5.2
			when 5 then 6
			else 7
			end`,
			want: &model.Case{
				Comparand: model.NewLiteral("5.2", types.Decimal),
				CaseItem: []*model.CaseItem{
					&model.CaseItem{
						When: &model.ToDecimal{
							UnaryExpression: &model.UnaryExpression{
								Operand:    model.NewLiteral("5", types.Integer),
								Expression: model.ResultType(types.Decimal),
							},
						},
						Then: model.NewLiteral("6", types.Integer),
					},
				},
				Else:       model.NewLiteral("7", types.Integer),
				Expression: model.ResultType(types.Integer),
			},
		},
		{
			name: "Case Then Implicitly Converted",
			cql: `
			case 4
			when 5 then 6.0
			when 7 then 8
			else 9L
			end`,
			want: &model.Case{
				Comparand: model.NewLiteral("4", types.Integer),
				CaseItem: []*model.CaseItem{
					&model.CaseItem{
						When: model.NewLiteral("5", types.Integer),
						Then: model.NewLiteral("6.0", types.Decimal),
					},
					&model.CaseItem{
						When: model.NewLiteral("7", types.Integer),
						Then: &model.ToDecimal{
							UnaryExpression: &model.UnaryExpression{
								Operand:    model.NewLiteral("8", types.Integer),
								Expression: model.ResultType(types.Decimal),
							},
						},
					},
				},
				Else: &model.ToDecimal{
					UnaryExpression: &model.UnaryExpression{
						Operand:    model.NewLiteral("9L", types.Long),
						Expression: model.ResultType(types.Decimal),
					},
				},
				Expression: model.ResultType(types.Decimal),
			},
		},
		{
			name: "Case Then Implicitly Converted to Choice",
			cql: `
			case 4
			when 5 then 6.0
			when 7 then 'Apple'
			else 9L
			end`,
			want: &model.Case{
				Comparand: model.NewLiteral("4", types.Integer),
				CaseItem: []*model.CaseItem{
					&model.CaseItem{
						When: model.NewLiteral("5", types.Integer),
						Then: &model.As{
							UnaryExpression: &model.UnaryExpression{
								Operand:    model.NewLiteral("6.0", types.Decimal),
								Expression: model.ResultType(&types.Choice{ChoiceTypes: []types.IType{types.Decimal, types.String, types.Long}}),
							},
							AsTypeSpecifier: &types.Choice{ChoiceTypes: []types.IType{types.Decimal, types.String, types.Long}},
						},
					},
					&model.CaseItem{
						When: model.NewLiteral("7", types.Integer),
						Then: &model.As{
							UnaryExpression: &model.UnaryExpression{
								Operand:    model.NewLiteral("Apple", types.String),
								Expression: model.ResultType(&types.Choice{ChoiceTypes: []types.IType{types.Decimal, types.String, types.Long}}),
							},
							AsTypeSpecifier: &types.Choice{ChoiceTypes: []types.IType{types.Decimal, types.String, types.Long}},
						},
					},
				},
				Else: &model.As{
					UnaryExpression: &model.UnaryExpression{
						Operand:    model.NewLiteral("9L", types.Long),
						Expression: model.ResultType(&types.Choice{ChoiceTypes: []types.IType{types.Decimal, types.String, types.Long}}),
					},
					AsTypeSpecifier: &types.Choice{ChoiceTypes: []types.IType{types.Decimal, types.String, types.Long}},
				},
				Expression: model.ResultType(&types.Choice{ChoiceTypes: []types.IType{types.Decimal, types.String, types.Long}}),
			},
		},
		{
			name: "As",
			cql:  "15 as String",
			want: &model.As{
				UnaryExpression: &model.UnaryExpression{
					Operand:    model.NewLiteral("15", types.Integer),
					Expression: model.ResultType(types.String),
				},
				AsTypeSpecifier: types.String,
				Strict:          false,
			},
		},
		{
			name: "Cast As",
			cql:  "cast 15 as String",
			want: &model.As{
				UnaryExpression: &model.UnaryExpression{
					Operand:    model.NewLiteral("15", types.Integer),
					Expression: model.ResultType(types.String),
				},
				AsTypeSpecifier: types.String,
				Strict:          true,
			},
		},
		{
			name: "Is",
			cql:  "15 is String",
			want: &model.Is{
				UnaryExpression: &model.UnaryExpression{
					Operand:    model.NewLiteral("15", types.Integer),
					Expression: model.ResultType(types.Boolean),
				},
				IsTypeSpecifier: types.String,
			},
		},
		{
			name: "If Then Else",
			cql:  "if 1 = 2 then 3 else 4",
			want: &model.IfThenElse{
				Condition: &model.Equal{
					BinaryExpression: &model.BinaryExpression{
						Operands: []model.IExpression{
							model.NewLiteral("1", types.Integer),
							model.NewLiteral("2", types.Integer),
						},
						Expression: model.ResultType(types.Boolean),
					},
				},
				Then:       model.NewLiteral("3", types.Integer),
				Else:       model.NewLiteral("4", types.Integer),
				Expression: model.ResultType(types.Integer),
			},
		},
		{
			name: "If Then Else with implicit result types",
			cql:  "if 1 = 2 then 3.0 else 4",
			want: &model.IfThenElse{
				Condition: &model.Equal{
					BinaryExpression: &model.BinaryExpression{
						Operands: []model.IExpression{
							model.NewLiteral("1", types.Integer),
							model.NewLiteral("2", types.Integer),
						},
						Expression: model.ResultType(types.Boolean),
					},
				},
				Then: model.NewLiteral("3.0", types.Decimal),
				Else: &model.ToDecimal{
					UnaryExpression: &model.UnaryExpression{
						Operand:    model.NewLiteral("4", types.Integer),
						Expression: model.ResultType(types.Decimal),
					},
				},
				Expression: model.ResultType(types.Decimal),
			},
		},
		{
			name: "predecessor of 1",
			cql:  "predecessor of 1",
			want: &model.Predecessor{
				UnaryExpression: &model.UnaryExpression{
					Operand:    model.NewLiteral("1", types.Integer),
					Expression: model.ResultType(types.Integer),
				},
			},
		},
		{
			name: "successor of @2023",
			cql:  "successor of @2023",
			want: &model.Successor{
				UnaryExpression: &model.UnaryExpression{
					Operand:    model.NewLiteral("@2023", types.Date),
					Expression: model.ResultType(types.Date),
				},
			},
		},
		{
			name: "Interval Low Inclusive High Exlusive",
			cql:  "Interval[10, 20)",
			want: &model.Interval{
				Expression: &model.Expression{
					Element: &model.Element{ResultType: &types.Interval{PointType: types.Integer}},
				},
				Low:           model.NewLiteral("10", types.Integer),
				High:          model.NewLiteral("20", types.Integer),
				LowInclusive:  true,
				HighInclusive: false,
			},
		},
		{
			name: "Interval Low Inclusive High Inclusive",
			cql:  "Interval[10, 20]",
			want: &model.Interval{
				Expression:    model.ResultType(&types.Interval{PointType: types.Integer}),
				Low:           model.NewLiteral("10", types.Integer),
				High:          model.NewLiteral("20", types.Integer),
				LowInclusive:  true,
				HighInclusive: true,
			},
		},
		{
			name: "Interval Low Exclusive High Exclusive",
			cql:  "Interval(10, 20)",
			want: &model.Interval{
				Expression:    model.ResultType(&types.Interval{PointType: types.Integer}),
				Low:           model.NewLiteral("10", types.Integer),
				High:          model.NewLiteral("20", types.Integer),
				LowInclusive:  false,
				HighInclusive: false,
			},
		},
		{
			name: "Interval Low Exclusive High Inclusive",
			cql:  "Interval(10, 20]",
			want: &model.Interval{
				Expression:    model.ResultType(&types.Interval{PointType: types.Integer}),
				Low:           model.NewLiteral("10", types.Integer),
				High:          model.NewLiteral("20", types.Integer),
				LowInclusive:  false,
				HighInclusive: true,
			},
		},
		{
			name: "Interval Type Implicitly Converted",
			cql:  "Interval(10, null]",
			want: &model.Interval{
				Expression: model.ResultType(&types.Interval{PointType: types.Integer}),
				Low:        model.NewLiteral("10", types.Integer),
				High: &model.As{
					UnaryExpression: &model.UnaryExpression{
						Operand:    model.NewLiteral("null", types.Any),
						Expression: model.ResultType(types.Integer),
					},
					AsTypeSpecifier: types.Integer,
				},
				LowInclusive:  false,
				HighInclusive: true,
			},
		},
		{
			name: "Interval low property",
			cql:  "Interval(10, 20].low",
			want: &model.Property{
				Source: &model.Interval{
					Expression:    model.ResultType(&types.Interval{PointType: types.Integer}),
					Low:           model.NewLiteral("10", types.Integer),
					High:          model.NewLiteral("20", types.Integer),
					LowInclusive:  false,
					HighInclusive: true,
				},
				Path:       "low",
				Expression: model.ResultType(types.Integer),
			},
		},
		{
			name: "Interval high property",
			cql:  "Interval(10, 20].high",
			want: &model.Property{
				Source: &model.Interval{
					Expression:    model.ResultType(&types.Interval{PointType: types.Integer}),
					Low:           model.NewLiteral("10", types.Integer),
					High:          model.NewLiteral("20", types.Integer),
					LowInclusive:  false,
					HighInclusive: true,
				},
				Path:       "high",
				Expression: model.ResultType(types.Integer),
			},
		},
		{
			name: "Interval highClosed property",
			cql:  "Interval(10, 20].highClosed",
			want: &model.Property{
				Source: &model.Interval{
					Expression:    model.ResultType(&types.Interval{PointType: types.Integer}),
					Low:           model.NewLiteral("10", types.Integer),
					High:          model.NewLiteral("20", types.Integer),
					LowInclusive:  false,
					HighInclusive: true,
				},
				Path:       "highClosed",
				Expression: model.ResultType(types.Boolean),
			},
		},
		{
			name: "Interval lowClosed property",
			cql:  "Interval(10, 20].lowClosed",
			want: &model.Property{
				Source: &model.Interval{
					Expression:    model.ResultType(&types.Interval{PointType: types.Integer}),
					Low:           model.NewLiteral("10", types.Integer),
					High:          model.NewLiteral("20", types.Integer),
					LowInclusive:  false,
					HighInclusive: true,
				},
				Path:       "lowClosed",
				Expression: model.ResultType(types.Boolean),
			},
		},
		{
			name: "Quoted Property",
			cql:  `Interval(10, 20]."lowClosed"`,
			want: &model.Property{
				Source: &model.Interval{
					Expression:    model.ResultType(&types.Interval{PointType: types.Integer}),
					Low:           model.NewLiteral("10", types.Integer),
					High:          model.NewLiteral("20", types.Integer),
					LowInclusive:  false,
					HighInclusive: true,
				},
				Path:       "lowClosed",
				Expression: model.ResultType(types.Boolean),
			},
		},
		{
			name: "Quantity value property",
			cql:  "6'gm/cm3'.value",
			want: &model.Property{
				Source: &model.Quantity{
					Value:      6,
					Unit:       "gm/cm3",
					Expression: model.ResultType(types.Quantity),
				},
				Path:       "value",
				Expression: model.ResultType(types.Decimal),
			},
		},
		{
			name: "Tuple property",
			cql:  "Tuple{foo: 4}.foo",
			want: &model.Property{
				Source: &model.Tuple{
					Elements:   []*model.TupleElement{&model.TupleElement{Name: "foo", Value: model.NewLiteral("4", types.Integer)}},
					Expression: model.ResultType(&types.Tuple{ElementTypes: map[string]types.IType{"foo": types.Integer}}),
				},
				Path:       "foo",
				Expression: model.ResultType(types.Integer),
			},
		},
		{
			name: "ListSelector No Type Specifier",
			cql:  "{'final', 'amended', 'corrected'}",
			want: &model.List{
				Expression: model.ResultType(&types.List{ElementType: types.String}),
				List: []model.IExpression{
					model.NewLiteral("final", types.String),
					model.NewLiteral("amended", types.String),
					model.NewLiteral("corrected", types.String),
				},
			},
		},
		{
			name: "ListSelector No Type Specifier with implicit conversions",
			cql:  "{10L, 1L, 20, 30}",
			want: &model.List{
				Expression: model.ResultType(&types.List{ElementType: types.Long}),
				List: []model.IExpression{
					model.NewLiteral("10L", types.Long),
					model.NewLiteral("1L", types.Long),
					&model.ToLong{
						UnaryExpression: &model.UnaryExpression{
							Operand:    model.NewLiteral("20", types.Integer),
							Expression: model.ResultType(types.Long),
						},
					},
					&model.ToLong{
						UnaryExpression: &model.UnaryExpression{
							Operand:    model.NewLiteral("30", types.Integer),
							Expression: model.ResultType(types.Long),
						},
					},
				},
			},
		},
		{
			name: "ListSelector with Type Specifier",
			cql:  "List<Integer>{1, 2, 3}",
			want: &model.List{
				Expression: model.ResultType(&types.List{ElementType: types.Integer}),
				List: []model.IExpression{
					model.NewLiteral("1", types.Integer),
					model.NewLiteral("2", types.Integer),
					model.NewLiteral("3", types.Integer),
				},
			},
		},
		{
			name: "ListSelector With Type Specifier requiring implicit conversions",
			cql:  "List<Long>{1, 2}",
			want: &model.List{
				Expression: model.ResultType(&types.List{ElementType: types.Long}),
				List: []model.IExpression{
					&model.ToLong{
						UnaryExpression: &model.UnaryExpression{
							Operand:    model.NewLiteral("1", types.Integer),
							Expression: model.ResultType(types.Long),
						},
					},
					&model.ToLong{
						UnaryExpression: &model.UnaryExpression{
							Operand:    model.NewLiteral("2", types.Integer),
							Expression: model.ResultType(types.Long),
						},
					},
				},
			},
		},
		{
			name: "ListSelector Empty List No TypeSpecifier",
			cql:  "{}",
			want: &model.List{
				Expression: model.ResultType(&types.List{ElementType: types.Any}),
				List:       []model.IExpression{},
			},
		},
		{
			name: "ListSelector Empty with TypeSpecifier",
			cql:  "List<Integer>{}",
			want: &model.List{
				Expression: model.ResultType(&types.List{ElementType: types.Integer}),
				List:       []model.IExpression{},
			},
		},
		{
			name: "ListSelector mixed",
			cql:  "{1, 'hi'}",
			want: &model.List{
				Expression: model.ResultType(&types.List{ElementType: &types.Choice{ChoiceTypes: []types.IType{types.Integer, types.String}}}),
				List: []model.IExpression{
					&model.As{
						UnaryExpression: &model.UnaryExpression{
							Expression: model.ResultType(&types.Choice{ChoiceTypes: []types.IType{types.Integer, types.String}}),
							Operand:    model.NewLiteral("1", types.Integer),
						},
						AsTypeSpecifier: &types.Choice{ChoiceTypes: []types.IType{types.Integer, types.String}},
					},
					&model.As{
						UnaryExpression: &model.UnaryExpression{
							Expression: model.ResultType(&types.Choice{ChoiceTypes: []types.IType{types.Integer, types.String}}),
							Operand:    model.NewLiteral("hi", types.String),
						},
						AsTypeSpecifier: &types.Choice{ChoiceTypes: []types.IType{types.Integer, types.String}},
					},
				},
			},
		},
		{
			name: "Literal Int",
			cql:  "10",
			want: model.NewLiteral("10", types.Integer),
		},
		{
			name: "Literal Decimal",
			cql:  "10.1",
			want: model.NewLiteral("10.1", types.Decimal),
		},
		{
			name: "Literal String",
			cql:  "'str'",
			want: model.NewLiteral("str", types.String),
		},
		{
			name: "Literal Bool",
			cql:  "false",
			want: model.NewLiteral("false", types.Boolean),
		},
		{
			name: "Literal Long",
			cql:  "1000000L",
			want: model.NewLiteral("1000000L", types.Long),
		},
		{
			name: "Literal DateTime",
			cql:  "@2023-01-25T14:30:14.559",
			want: model.NewLiteral("@2023-01-25T14:30:14.559", types.DateTime),
		},
		{
			name: "Literal Date",
			cql:  "@2023-01-25",
			want: model.NewLiteral("@2023-01-25", types.Date),
		},
		{
			name: "Literal Time",
			cql:  "@T10:00:00.0",
			want: model.NewLiteral("@T10:00:00.0", types.Time),
		},
		{
			name: "Literal Quantity",
			cql:  "6'gm/cm3'",
			want: &model.Quantity{Value: 6, Unit: "gm/cm3", Expression: model.ResultType(types.Quantity)},
		},
		{
			name: "Literal Quantity temporal",
			cql:  "1 year",
			want: &model.Quantity{Value: 1, Unit: "year", Expression: model.ResultType(types.Quantity)},
		},
		{
			name: "Literal Quantity temporal plural",
			cql:  "6 years",
			want: &model.Quantity{Value: 6, Unit: "year", Expression: model.ResultType(types.Quantity)},
		},
		{
			name: "Literal Ratio",
			cql:  "1 'cm':2 'cm'",
			want: &model.Ratio{
				Numerator:   model.Quantity{Value: 1, Unit: "cm", Expression: model.ResultType(types.Quantity)},
				Denominator: model.Quantity{Value: 2, Unit: "cm", Expression: model.ResultType(types.Quantity)},
				Expression:  model.ResultType(types.Ratio),
			},
		},
		{
			name: "Time Unit Expression, 'date from'",
			cql:  dedent.Dedent(`date from @2013-01-01T00:00:00.0`),
			want: &model.ToDate{
				UnaryExpression: &model.UnaryExpression{
					Operand:    model.NewLiteral("@2013-01-01T00:00:00.0", types.DateTime),
					Expression: model.ResultType(types.Date),
				},
			},
		},
		{
			name: "Tuple Selector",
			cql:  "Tuple{code: 'foo', id: 4}",
			want: &model.Tuple{
				Expression: model.ResultType(&types.Tuple{ElementTypes: map[string]types.IType{"code": types.String, "id": types.Integer}}),
				Elements: []*model.TupleElement{
					&model.TupleElement{Name: "code", Value: model.NewLiteral("foo", types.String)},
					&model.TupleElement{Name: "id", Value: model.NewLiteral("4", types.Integer)},
				},
			},
		},
		{
			name: "Instance Selector",
			cql:  "Code{code: 'foo', system: 'bar', display: 'the foo', version: '1.0'}",
			want: &model.Instance{
				Expression: model.ResultType(types.Code),
				ClassType:  types.Code,
				Elements: []*model.InstanceElement{
					&model.InstanceElement{Name: "code", Value: model.NewLiteral("foo", types.String)},
					&model.InstanceElement{Name: "system", Value: model.NewLiteral("bar", types.String)},
					&model.InstanceElement{Name: "display", Value: model.NewLiteral("the foo", types.String)},
					&model.InstanceElement{Name: "version", Value: model.NewLiteral("1.0", types.String)},
				},
			},
		},
		{
			name: "Instance Selector with implicit conversion",
			cql:  "Quantity {value: 4, unit: 'mg'}",
			want: &model.Instance{
				Expression: model.ResultType(types.Quantity),
				ClassType:  types.Quantity,
				Elements: []*model.InstanceElement{
					&model.InstanceElement{
						Name: "value",
						Value: &model.ToDecimal{
							UnaryExpression: &model.UnaryExpression{
								Operand:    model.NewLiteral("4", types.Integer),
								Expression: model.ResultType(types.Decimal),
							},
						},
					},
					&model.InstanceElement{Name: "unit", Value: model.NewLiteral("mg", types.String)},
				},
			},
		},
		{
			name: "NamedTypeSpecifier Quoted",
			cql:  `List<"System.String">{}`,
			want: &model.List{
				Expression: model.ResultType(&types.List{ElementType: types.String}),
				List:       []model.IExpression{},
			},
		},
		{
			name: "NamedTypeSpecifier Qualfied and Quoted",
			cql:  `List<"System".String>{}`,
			want: &model.List{
				Expression: model.ResultType(&types.List{ElementType: types.String}),
				List:       []model.IExpression{},
			},
		},
		{
			name: "NamedTypeSpecifier Qualfied and Unquoted",
			cql:  `List<System.String>{}`,
			want: &model.List{
				Expression: model.ResultType(&types.List{ElementType: types.String}),
				List:       []model.IExpression{},
			},
		},
		{
			name: "NamedTypeSpecifier FHIR",
			cql:  `List<Patient>{}`,
			want: &model.List{
				Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Patient"}}),
				List:       []model.IExpression{},
			},
		},
		{
			name: "maximum Integer",
			cql:  "maximum Integer",
			want: &model.MaxValue{ValueType: types.Integer, Expression: model.ResultType(types.Integer)},
		},
		{
			name: "minimum Decimal",
			cql:  "minimum Decimal",
			want: &model.MinValue{ValueType: types.Decimal, Expression: model.ResultType(types.Decimal)},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parsedLibs, err := newFHIRParser(t).Libraries(context.Background(), wrapInLib(t, test.cql), Config{})
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(test.want, getTESTRESULTModel(t, parsedLibs)); diff != "" {
				t.Errorf("Parse diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParserExpressions_SingleLibrary(t *testing.T) {
	tests := []struct {
		name string
		desc string
		cql  string
		want *model.Library
	}{
		{
			name: "Choice Type Specifier",
			cql:  `define function "Population"(P Choice<Integer, String>):  4`,
			want: &model.Library{
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.FunctionDef{
							ExpressionDef: &model.ExpressionDef{
								Name:        "Population",
								AccessLevel: "PUBLIC",
								Expression:  model.NewLiteral("4", types.Integer),
								Element:     &model.Element{ResultType: types.Integer},
							},
							Operands: []model.OperandDef{{Name: "P", Expression: model.ResultType(&types.Choice{ChoiceTypes: []types.IType{types.Integer, types.String}})}},
						},
					},
				},
			},
		},
		{
			name: "Tuple Type Specifier",
			cql:  `define function "Population"(P Tuple{apple Integer, banana String}):  4`,
			want: &model.Library{
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.FunctionDef{
							ExpressionDef: &model.ExpressionDef{
								Name:        "Population",
								AccessLevel: "PUBLIC",
								Expression:  model.NewLiteral("4", types.Integer),
								Element:     &model.Element{ResultType: types.Integer},
							},
							Operands: []model.OperandDef{{Name: "P", Expression: model.ResultType(&types.Tuple{ElementTypes: map[string]types.IType{"apple": types.Integer, "banana": types.String}})}},
						},
					},
				},
			},
		},
		{
			name: "List Type Specifier",
			cql:  `define function "Population"(P List<Integer>):  4`,
			want: &model.Library{
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.FunctionDef{
							ExpressionDef: &model.ExpressionDef{
								Name:        "Population",
								AccessLevel: "PUBLIC",
								Expression:  model.NewLiteral("4", types.Integer),
								Element:     &model.Element{ResultType: types.Integer},
							},
							Operands: []model.OperandDef{{Name: "P", Expression: model.ResultType(&types.List{ElementType: types.Integer})}},
						},
					},
				},
			},
		},
		{
			name: "Interval Type Specifier",
			cql:  `define function "Population"(P Interval<Integer>):  4`,
			want: &model.Library{
				Statements: &model.Statements{
					Defs: []model.IExpressionDef{
						&model.FunctionDef{
							ExpressionDef: &model.ExpressionDef{
								Name:        "Population",
								AccessLevel: "PUBLIC",
								Expression:  model.NewLiteral("4", types.Integer),
								Element:     &model.Element{ResultType: types.Integer},
							},
							Operands: []model.OperandDef{{Name: "P", Expression: model.ResultType(&types.Interval{PointType: types.Integer})}},
						},
					},
				},
			},
		},
		{
			name: "Nested Alias Scopes",
			cql: dedent.Dedent(`
        using FHIR version '4.0.1'
        define function ToConcept(concept FHIR.CodeableConcept):
							System.Concept {
									codes: concept.coding C return C,
									display: concept.text.value
							}`),
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
						&model.FunctionDef{
							ExpressionDef: &model.ExpressionDef{
								Name:        "ToConcept",
								AccessLevel: "PUBLIC",
								Context:     "Patient",
								Element:     &model.Element{ResultType: types.Concept},
								Expression: &model.Instance{
									Expression: &model.Expression{Element: &model.Element{ResultType: types.Concept}},
									ClassType:  types.Concept,
									Elements: []*model.InstanceElement{
										&model.InstanceElement{
											Name: "codes",
											// We need to convert "concept.coding C return C" which returns a
											// List<FHIR.Coding> to List<Coding> to match the codes element in the Concept
											// structured type. To do that the parser inserts "(concept.coding C return C)
											// X return all FHIRHelper.ToCode(X)"
											Value: &model.Query{
												Expression: model.ResultType(&types.List{ElementType: types.Code}),
												Source: []*model.AliasedSource{
													&model.AliasedSource{
														Alias:      "X",
														Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Coding"}}),
														// This is the "concept.coding C return C" query you see in the CQL which returns List<FHIR.Coding>
														Source: &model.Query{
															Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Coding"}}),
															Source: []*model.AliasedSource{
																&model.AliasedSource{
																	Alias:      "C",
																	Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Coding"}}),
																	Source: &model.Property{
																		Source: &model.OperandRef{
																			Name:       "concept",
																			Expression: model.ResultType(&types.Named{TypeName: "FHIR.CodeableConcept"}),
																		},
																		Path:       "coding",
																		Expression: model.ResultType(&types.List{ElementType: &types.Named{TypeName: "FHIR.Coding"}}),
																	},
																},
															},
															Return: &model.ReturnClause{
																Element: &model.Element{ResultType: &types.Named{TypeName: "FHIR.Coding"}},
																Expression: &model.AliasRef{
																	Name:       "C",
																	Expression: model.ResultType(&types.Named{TypeName: "FHIR.Coding"})},
																Distinct: true,
															},
														},
													},
												},
												Return: &model.ReturnClause{
													Element: &model.Element{ResultType: types.Code},
													Expression: &model.FunctionRef{
														LibraryName: "FHIRHelpers",
														Name:        "ToCode",
														Operands: []model.IExpression{&model.AliasRef{
															Name:       "X",
															Expression: model.ResultType(&types.Named{TypeName: "FHIR.Coding"})}},
														Expression: model.ResultType(types.Code),
													},
													Distinct: false,
												},
											},
										},
										&model.InstanceElement{
											Name: "display",
											Value: &model.Property{
												Path:       "value",
												Expression: model.ResultType(types.String),
												Source: &model.Property{
													Path: "text",
													Source: &model.OperandRef{
														Name:       "concept",
														Expression: model.ResultType(&types.Named{TypeName: "FHIR.CodeableConcept"}),
													},
													Expression: model.ResultType(&types.Named{TypeName: "FHIR.string"}),
												},
											},
										},
									},
								},
							},
							Operands: []model.OperandDef{{Name: "concept", Expression: model.ResultType(&types.Named{TypeName: "FHIR.CodeableConcept"})}},
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
				t.Errorf("Parse diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestIdentifiers(t *testing.T) {
	// These are very targeted tests for various ways to write identifiers in CQL,
	// following https://cql.hl7.org/03-developersguide.html#identifiers and
	// https://cql.hl7.org/19-l-cqlsyntaxdiagrams.html#identifier.
	//
	// Some use cases are not yet implemented in the parser and would be skipped.

	tests := []struct {
		// The identifier as it would appear in an authored CQL library.
		given string
		// The parsed name of the identifier as it should appear in the parsed model,
		// which should be stripped from quoting and unescaped.
		// If set to `expectError` then the given identifier is expected to cause
		// `ParsingErrors` to be returned.
		want string

		// True if this case is expected to return a ParseError
		expectError bool

		// True to skip not-yet-implemented cases
		skip bool
	}{
		// Examples from https://cql.hl7.org/03-developersguide.html#identifiers.
		// Simple identifiers:
		{given: "Foo", want: "Foo"},
		{given: "Foo1", want: "Foo1"},
		{given: "_Foo", want: "_Foo"}, // discouraged, but allowed
		{given: "foo", want: "foo"},
		// Delimited identifiers:
		{given: "`Diagnosis`", want: "Diagnosis"},
		{given: "`Encounter, Performed`", want: "Encounter, Performed"},
		// Quoted identifiers:
		{given: `"Diagnosis"`, want: "Diagnosis"},
		{given: `"Encounter, Performed"`, want: "Encounter, Performed"},
		// Delimited identifiers are unescaped:
		{given: "`with\\'squote`", want: `with'squote`},
		{given: "`with\\\"dquote`", want: `with"dquote`},
		{given: "`with\\`backtick`", want: "with`backtick"},
		{given: "`with\\rCR`", want: "with\rCR"},
		{given: "`with\\nLF`", want: "with\nLF"},
		{given: "`with\\tTAB`", want: "with\tTAB"},
		{given: "`with\\fFF`", want: "with\fFF"},
		{given: "`with\\\\backslash`", want: `with\backslash`},
		{given: "`with\\u0020unicode`", want: "with unicode", skip: true},
		// Quoted identifiers are unescaped:
		{given: "\"with\\'squote\"", want: `with'squote`},
		{given: "\"with\\\"dquote\"", want: `with"dquote`},
		{given: "\"with\\`backtick\"", want: "with`backtick"},
		{given: "\"with\\rCR\"", want: "with\rCR"},
		{given: "`with\\nLF`", want: "with\nLF"},
		{given: "\"with\\tTAB\"", want: "with\tTAB"},
		{given: "`with\\fFF`", want: "with\fFF"},
		{given: "\"with\\\\backslash\"", want: `with\backslash`},
		{given: "\"with\\u0020unicode\"", want: "with unicode", skip: true},
		// More valid simple identifiers:
		{given: "fOO", want: "fOO"},
		{given: "CamelCase", want: "CamelCase"},
		{given: "d1g1ts", want: "d1g1ts"},
		{given: "numb34", want: "numb34"},
		{given: "under_score", want: "under_score"},
		{given: "underscore_", want: "underscore_"},
		// Additional valid quoted and delimited examples. Those are not pretty, but are allowed:
		{given: `"with.dot"`, want: "with.dot"},   //  "with-dot"  -> with-dot
		{given: `"with-dash"`, want: "with-dash"}, // "with-dash" -> with-dash
		{given: `" "`, want: " "},                 //  ""  -> a space
		// Invalid simple identifiers which would require quoting or delimiting to be used:
		{given: "0Foo", expectError: true},
		{given: " ", expectError: true},
		{given: "with-dash", expectError: true},
		{given: "unicode\\u0030", expectError: true},
		{given: "'singlequote'", expectError: true},
	}

	for _, tc := range tests {
		t.Run(tc.given, func(t *testing.T) {
			if tc.skip {
				t.Skipf("not implemented yet")
			}

			cql := fmt.Sprintf("define %s: 4", tc.given)
			libs, err := newFHIRParser(t).Libraries(context.Background(), []string{cql}, Config{})
			if tc.expectError {
				if err == nil {
					t.Fatal("Libraries() expected parse error but got none")
				}
				var pe *LibraryErrors
				if ok := errors.As(err, &pe); !ok {
					t.Fatal("Libraries() expected parse error: ", err)
				}
				return
			}

			gotName := libs[0].Statements.Defs[0].GetName()
			if gotName != tc.want {
				t.Errorf("Extracting name from parsed Library(%s) = [%v], want [%s]", cql, gotName, tc.want)
			}
		})

	}
}

func TestIdentifiableExpressions(t *testing.T) {
	// TestIdentifiableExpressions tests that all locations that handle identifiers properly call
	// VisitIdentifier. For detailed testing of VisitIdentifier see the TestIdentifiers suite. Some
	// use cases are not yet implemented in the parser and would be skipped.

	identifiers := []struct {
		// The identifier as it would appear in an authored CQL library.
		given string
		// The parsed name of the identifier as it should appear in the parsed model,
		// which should be stripped from quoting and unescaped.
		// If set to `expectError` then the given identifier is expected to cause
		// `ParsingErrors` to be returned.
		want string

		// True if this case is expected to return a ParseError
		expectError bool

		// True to skip not-yet-implemented cases
		skip bool
	}{
		// Examples from https://cql.hl7.org/03-developersguide.html#identifiers.
		// Simple identifier:
		{given: "Foo", want: "Foo"},
		// Quoted identifier:
		{given: `"Diagnosis"`, want: "Diagnosis"},
		// Delimited identifiers are unescaped:
		{given: "`with\\'squote`", want: `with'squote`},
		// Quoted identifiers are unescaped:
		{given: "\"with\\'squote\"", want: `with'squote`},
		// Invalid simple identifiers which would require quoting or delimiting to be used:
		{given: "'singlequote'", expectError: true},
	}

	identifiableExpressions := map[string]struct {
		// CQL template with placeholders for a supplied identifier.
		cql string

		// Extract the parsed identifier name from the parsed library.
		gotExtractor func(*model.Library) string

		// True to skip not-yet-implemented cases
		skip bool
	}{
		"parameterDefinition": {
			cql: "parameter %s Integer",
			gotExtractor: func(got *model.Library) string {
				return got.Parameters[0].Name
			},
		},
		"valuesetDefinition": {
			cql: "valueset %s: 'TRIVIAL'",
			gotExtractor: func(got *model.Library) string {
				return got.Valuesets[0].Name
			},
		},
		"expressionDefinition": {
			cql: "define %s: ''",
			gotExtractor: func(got *model.Library) string {
				return got.Statements.Defs[0].GetName()
			},
		},
		"from:": {
			cql: dedent.Dedent(`
			define %[1]s: ''
			define TRIVIAL: from %[1]s TRIVIAL return TRIVIAL`),
			gotExtractor: func(got *model.Library) string {
				return got.Statements.Defs[1].GetExpression().(*model.Query).Source[0].Source.(*model.ExpressionRef).Name
			},
		},
		"from-alias:": {
			cql: dedent.Dedent(`
			define OTHER: ''
			define TRIVIAL: from OTHER %s`),
			gotExtractor: func(got *model.Library) string {
				return got.Statements.Defs[1].GetExpression().(*model.Query).Source[0].Alias
			},
		},
		"from-return:": {
			cql: dedent.Dedent(`
			define OTHER: ''
			define TRIVIAL: from OTHER TRIVIAL return %s`),
			skip: true,
		},
		"from-aggregate:": {
			cql: dedent.Dedent(`
			define OTHER: ''
			define TRIVIAL: from OTHER TRIVIAL aggregate %s: 0`),
			skip: true,
		},
		"externalConstant": {
			cql:  "define TRIVIAL: %%%s",
			skip: true,
		},
		"let": {
			cql: dedent.Dedent(`
			define OTHER: ''
			define TRIVIAL: from OTHER O let %[1]s :'' return %[1]s`),
			skip: true,
		},
		"localIdentifier": {
			cql:  "include TRIVIAL called %s",
			skip: true,
		},
		"functionDefinition": {
			cql: "define function %s () : external",
			gotExtractor: func(got *model.Library) string {
				return got.Statements.Defs[0].GetName()
			},
		},
		"functionDefinition-op-type": {
			cql:  "define function TRIVIAL (TRIVIAL %s) : external",
			skip: true,
		},
		"functionDefinition-op-name": {
			cql: "define function TRIVIAL (%[1]s Integer): %[1]s+1",
			gotExtractor: func(got *model.Library) string {
				return got.Statements.Defs[0].(*model.FunctionDef).Operands[0].Name
			},
		},
		"codeDefinition": {
			cql:  "code %s: 'TRIVIAL' from TRIVIAL",
			skip: true,
		},
		"codeDefinition-from-lib": {
			cql:  "code TRIVIAL: 'TRIVIAL' from %s.TRIVIAL",
			skip: true,
		},
		"codeDefinition-from-id": {
			cql:  "code TRIVIAL: 'TRIVIAL' from TRIVIAL.%s",
			skip: true,
		},
		"conceptDefinition": {
			cql:  "concept %s: { TRIVIAL }",
			skip: true,
		},
		"conceptDefinition-code": {
			cql:  "concept TRIVIAL: { %s }",
			skip: true,
		},
		"codesystemDefinition": {
			cql: "codesystem %s: 'TRIVIAL'",
			gotExtractor: func(got *model.Library) string {
				return got.CodeSystems[0].Name
			},
		},
	}

	for _, identifier := range identifiers {
		for exprName, test := range identifiableExpressions {
			t.Run(identifier.given+"_"+exprName, func(t *testing.T) {
				t.Parallel()
				if test.skip || identifier.skip {
					t.Skipf("not implemented yet")
				}

				cql := fmt.Sprintf(test.cql, identifier.given)
				libs, err := newFHIRParser(t).Libraries(context.Background(), []string{cql}, Config{})
				if identifier.expectError {
					if err == nil {
						t.Fatal("Libraries() expected parse error but got none")
					}
					var pe *LibraryErrors
					ok := errors.As(err, &pe)
					if !ok {
						t.Fatal("Libraries() expected parse error: ", err)
					}
					return
				}

				gotName := test.gotExtractor(libs[0])

				if gotName != identifier.want {
					t.Errorf("Extracting name from parsed Library(%s) = [%v], want [%s]", cql, gotName, identifier.want)
				}
			})
		}
	}
}

func newFHIRParser(t testing.TB) *Parser {
	t.Helper()
	fhirMI, err := embeddata.ModelInfos.ReadFile("third_party/cqframework/fhir-modelinfo-4.0.1.xml")
	if err != nil {
		t.Fatalf("internal error - could not read fhir-modelinfo-4.0.1.xml: %v", err)
	}
	p, err := New(context.Background(), [][]byte{fhirMI})
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
