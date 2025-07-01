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
	"strings"
	"testing"

	"github.com/google/cql/model"
	"github.com/google/cql/types"
	"github.com/google/go-cmp/cmp"
	"github.com/lithammer/dedent"
)

func TestOperatorExpressions(t *testing.T) {
	tests := []struct {
		name string
		desc string
		cql  string
		want model.IExpression
	}{
		{
			name: "Concatenate &",
			cql:  "'Hi' & 'Hello'",
			want: &model.Concatenate{
				NaryExpression: &model.NaryExpression{
					Operands: []model.IExpression{
						&model.Coalesce{
							NaryExpression: &model.NaryExpression{
								Operands:   []model.IExpression{model.NewLiteral("Hi", types.String), model.NewLiteral("", types.String)},
								Expression: model.ResultType(types.String),
							},
						},
						&model.Coalesce{
							NaryExpression: &model.NaryExpression{
								Operands:   []model.IExpression{model.NewLiteral("Hello", types.String), model.NewLiteral("", types.String)},
								Expression: model.ResultType(types.String),
							},
						},
					},
					Expression: model.ResultType(types.String),
				},
			},
		},
		{
			name: "Concatenate +",
			cql:  "'Hi' + 'Hello'",
			want: &model.Concatenate{
				NaryExpression: &model.NaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("Hi", types.String),
						model.NewLiteral("Hello", types.String),
					},
					Expression: model.ResultType(types.String),
				},
			},
		},
		{
			name: "Arithmetic Addition",
			cql:  "1 + 2",
			want: &model.Add{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("1", types.Integer),
						model.NewLiteral("2", types.Integer),
					},
					Expression: model.ResultType(types.Integer),
				},
			},
		},
		{
			name: "Arithmetic Subtraction",
			cql:  "1 - 2",
			want: &model.Subtract{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("1", types.Integer),
						model.NewLiteral("2", types.Integer),
					},
					Expression: model.ResultType(types.Integer),
				},
			},
		},
		{
			name: "Arithmetic Multiplication",
			cql:  "1 * 2",
			want: &model.Multiply{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("1", types.Integer),
						model.NewLiteral("2", types.Integer),
					},
					Expression: model.ResultType(types.Integer),
				},
			},
		},
		{
			name: "Arithmetic Modulo with different types",
			cql:  "40L mod 3",
			want: &model.Modulo{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("40L", types.Long),
						&model.ToLong{
							UnaryExpression: &model.UnaryExpression{
								Operand:    model.NewLiteral("3", types.Integer),
								Expression: model.ResultType(types.Long),
							},
						},
					},
					Expression: model.ResultType(types.Long),
				},
			},
		},
		{
			name: "Arithmetic Power with different types",
			cql:  "4.0 ^ 2",
			want: &model.Power{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("4.0", types.Decimal),
						&model.ToDecimal{
							UnaryExpression: &model.UnaryExpression{
								Operand:    model.NewLiteral("2", types.Integer),
								Expression: model.ResultType(types.Decimal),
							},
						},
					},
					Expression: model.ResultType(types.Decimal),
				},
			},
		},
		{
			name: "Arithmetic Truncated Divide",
			cql:  "40 div 3",
			want: &model.TruncatedDivide{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("40", types.Integer),
						model.NewLiteral("3", types.Integer),
					},
					Expression: model.ResultType(types.Integer),
				},
			},
		},
		{
			name: "Arithmetic Divide Integer by Decimal",
			cql:  "40 / 3.1234567",
			want: &model.Divide{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						&model.ToDecimal{
							UnaryExpression: &model.UnaryExpression{
								Operand:    model.NewLiteral("40", types.Integer),
								Expression: model.ResultType(types.Decimal),
							},
						},
						model.NewLiteral("3.1234567", types.Decimal),
					},
					Expression: model.ResultType(types.Decimal),
				},
			},
		},
		{
			name: "Arithmetic Order Of Operations",
			cql:  "1 * 3 + 5",
			want: &model.Add{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						&model.Multiply{
							BinaryExpression: &model.BinaryExpression{
								Operands: []model.IExpression{
									model.NewLiteral("1", types.Integer),
									model.NewLiteral("3", types.Integer),
								},
								Expression: model.ResultType(types.Integer),
							},
						},
						model.NewLiteral("5", types.Integer),
					},
					Expression: model.ResultType(types.Integer),
				},
			},
		},
		{
			name: "Arithmetic Order Of Operations 2",
			cql:  "1 + 3 * 5",
			want: &model.Add{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("1", types.Integer),
						&model.Multiply{
							BinaryExpression: &model.BinaryExpression{
								Operands: []model.IExpression{
									model.NewLiteral("3", types.Integer),
									model.NewLiteral("5", types.Integer),
								},
								Expression: model.ResultType(types.Integer),
							},
						},
					},
					Expression: model.ResultType(types.Integer),
				},
			},
		},
		{
			name: "Negate",
			cql:  "-4",
			want: &model.Negate{
				UnaryExpression: &model.UnaryExpression{
					Operand:    model.NewLiteral("4", types.Integer),
					Expression: model.ResultType(types.Integer),
				},
			},
		},
		{
			name: "+ does not negate",
			cql:  "+4",
			want: model.NewLiteral("4", types.Integer),
		},
		{
			name: "Timing Expression with temporal Interval Operator (before)",
			cql:  "Interval[@2013-01-01T00:00:00.0, @2014-01-01T00:00:00.0) before start of Interval[@2013-01-01T00:00:00.0, @2014-01-01T00:00:00.0)",
			want: &model.Before{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						&model.Interval{
							Low:           model.NewLiteral("@2013-01-01T00:00:00.0", types.DateTime),
							High:          model.NewLiteral("@2014-01-01T00:00:00.0", types.DateTime),
							Expression:    model.ResultType(&types.Interval{PointType: types.DateTime}),
							LowInclusive:  true,
							HighInclusive: false,
						},
						&model.Start{
							UnaryExpression: &model.UnaryExpression{
								Operand: &model.Interval{
									Low:           model.NewLiteral("@2013-01-01T00:00:00.0", types.DateTime),
									High:          model.NewLiteral("@2014-01-01T00:00:00.0", types.DateTime),
									Expression:    model.ResultType(&types.Interval{PointType: types.DateTime}),
									LowInclusive:  true,
									HighInclusive: false,
								},
								Expression: model.ResultType(types.DateTime),
							},
						},
					},
					Expression: model.ResultType(types.Boolean),
				},
			},
		},
		{
			name: "TimingExpression On Or Before",
			cql:  "Interval[@2013-01-01T00:00:00.0, @2014-01-01T00:00:00.0) on or before start of Interval[@2013-01-01T00:00:00.0, @2014-01-01T00:00:00.0)",
			want: &model.SameOrBefore{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						&model.Interval{
							Low:           model.NewLiteral("@2013-01-01T00:00:00.0", types.DateTime),
							High:          model.NewLiteral("@2014-01-01T00:00:00.0", types.DateTime),
							Expression:    model.ResultType(&types.Interval{PointType: types.DateTime}),
							LowInclusive:  true,
							HighInclusive: false,
						},
						&model.Start{
							UnaryExpression: &model.UnaryExpression{
								Operand: &model.Interval{
									Low:           model.NewLiteral("@2013-01-01T00:00:00.0", types.DateTime),
									High:          model.NewLiteral("@2014-01-01T00:00:00.0", types.DateTime),
									Expression:    model.ResultType(&types.Interval{PointType: types.DateTime}),
									LowInclusive:  true,
									HighInclusive: false,
								},
								Expression: model.ResultType(types.DateTime),
							},
						},
					},
					Expression: model.ResultType(types.Boolean),
				},
			},
		},
		{
			name: "TimingExpression Before Or On Start Of Interval",
			cql:  "Interval[@2013-01-01T00:00:00.0, @2014-01-01T00:00:00.0) before or on start of Interval[@2013-01-01T00:00:00.0, @2014-01-01T00:00:00.0)",
			want: &model.SameOrBefore{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						&model.Interval{
							Low:           model.NewLiteral("@2013-01-01T00:00:00.0", types.DateTime),
							High:          model.NewLiteral("@2014-01-01T00:00:00.0", types.DateTime),
							Expression:    model.ResultType(&types.Interval{PointType: types.DateTime}),
							LowInclusive:  true,
							HighInclusive: false,
						},
						&model.Start{
							UnaryExpression: &model.UnaryExpression{
								Operand: &model.Interval{
									Low:           model.NewLiteral("@2013-01-01T00:00:00.0", types.DateTime),
									High:          model.NewLiteral("@2014-01-01T00:00:00.0", types.DateTime),
									Expression:    model.ResultType(&types.Interval{PointType: types.DateTime}),
									LowInclusive:  true,
									HighInclusive: false,
								},
								Expression: model.ResultType(types.DateTime),
							},
						},
					},
					Expression: model.ResultType(types.Boolean),
				},
			},
		},
		{
			name: "TimingExpression After Start Of Interval",
			cql:  "Interval[@2013-01-01T00:00:00.0, @2014-01-01T00:00:00.0) after start of Interval[@2013-01-01T00:00:00.0, @2014-01-01T00:00:00.0)",
			want: &model.After{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						&model.Interval{
							Low:           model.NewLiteral("@2013-01-01T00:00:00.0", types.DateTime),
							High:          model.NewLiteral("@2014-01-01T00:00:00.0", types.DateTime),
							Expression:    model.ResultType(&types.Interval{PointType: types.DateTime}),
							LowInclusive:  true,
							HighInclusive: false,
						},
						&model.Start{
							UnaryExpression: &model.UnaryExpression{
								Operand: &model.Interval{
									Low:           model.NewLiteral("@2013-01-01T00:00:00.0", types.DateTime),
									High:          model.NewLiteral("@2014-01-01T00:00:00.0", types.DateTime),
									Expression:    model.ResultType(&types.Interval{PointType: types.DateTime}),
									LowInclusive:  true,
									HighInclusive: false,
								},
								Expression: model.ResultType(types.DateTime),
							},
						},
					},
					Expression: model.ResultType(types.Boolean),
				},
			},
		},
		{
			name: "TimingExpression After Or On Start Of Interval",
			cql:  "Interval[@2013-01-01T00:00:00.0, @2014-01-01T00:00:00.0) after or on start of Interval[@2013-01-01T00:00:00.0, @2014-01-01T00:00:00.0)",
			want: &model.SameOrAfter{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						&model.Interval{
							Low:           model.NewLiteral("@2013-01-01T00:00:00.0", types.DateTime),
							High:          model.NewLiteral("@2014-01-01T00:00:00.0", types.DateTime),
							Expression:    model.ResultType(&types.Interval{PointType: types.DateTime}),
							LowInclusive:  true,
							HighInclusive: false,
						},
						&model.Start{
							UnaryExpression: &model.UnaryExpression{
								Operand: &model.Interval{
									Low:           model.NewLiteral("@2013-01-01T00:00:00.0", types.DateTime),
									High:          model.NewLiteral("@2014-01-01T00:00:00.0", types.DateTime),
									LowInclusive:  true,
									HighInclusive: false,
									Expression:    model.ResultType(&types.Interval{PointType: types.DateTime}),
								},
								Expression: model.ResultType(types.DateTime),
							},
						},
					},
					Expression: model.ResultType(types.Boolean),
				},
			},
		},
		{
			name: "TimingExpression On Or After Start Of Interval",
			cql:  "Interval[@2013-01-01T00:00:00.0, @2014-01-01T00:00:00.0) on or after start of Interval[@2013-01-01T00:00:00.0, @2014-01-01T00:00:00.0)",
			want: &model.SameOrAfter{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						&model.Interval{
							Low:           model.NewLiteral("@2013-01-01T00:00:00.0", types.DateTime),
							High:          model.NewLiteral("@2014-01-01T00:00:00.0", types.DateTime),
							Expression:    model.ResultType(&types.Interval{PointType: types.DateTime}),
							LowInclusive:  true,
							HighInclusive: false,
						},
						&model.Start{
							UnaryExpression: &model.UnaryExpression{
								Operand: &model.Interval{
									Low:           model.NewLiteral("@2013-01-01T00:00:00.0", types.DateTime),
									High:          model.NewLiteral("@2014-01-01T00:00:00.0", types.DateTime),
									Expression:    model.ResultType(&types.Interval{PointType: types.DateTime}),
									LowInclusive:  true,
									HighInclusive: false,
								},
								Expression: model.ResultType(types.DateTime),
							},
						},
					},
					Expression: model.ResultType(types.Boolean),
				},
			},
		},
		{
			name: "TimingExpression On Or After Start Of Interval With Precision",
			cql:  "Interval[@2013-01-01T00:00:00.0, @2014-01-01T00:00:00.0) on or after year of start of Interval[@2013-01-01T00:00:00.0, @2014-01-01T00:00:00.0)",
			want: &model.SameOrAfter{
				Precision: model.YEAR,
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						&model.Interval{
							Low:           model.NewLiteral("@2013-01-01T00:00:00.0", types.DateTime),
							High:          model.NewLiteral("@2014-01-01T00:00:00.0", types.DateTime),
							Expression:    model.ResultType(&types.Interval{PointType: types.DateTime}),
							LowInclusive:  true,
							HighInclusive: false,
						},
						&model.Start{
							UnaryExpression: &model.UnaryExpression{
								Operand: &model.Interval{
									Low:           model.NewLiteral("@2013-01-01T00:00:00.0", types.DateTime),
									High:          model.NewLiteral("@2014-01-01T00:00:00.0", types.DateTime),
									Expression:    model.ResultType(&types.Interval{PointType: types.DateTime}),
									LowInclusive:  true,
									HighInclusive: false,
								},
								Expression: model.ResultType(types.DateTime),
							},
						},
					},
					Expression: model.ResultType(types.Boolean),
				},
			},
		},
		{
			name: "BeforeOrAfterIntervalOperatorPhraseContext Date 1 Year Or Less On Or Before Date",
			cql:  "@2020 1 year or less on or before @2022",
			want: &model.In{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("@2020", types.Date),
						&model.Interval{
							Low: &model.Subtract{
								BinaryExpression: &model.BinaryExpression{
									Operands: []model.IExpression{
										model.NewLiteral("@2022", types.Date),
										&model.Quantity{Value: 1, Unit: "year", Expression: model.ResultType(types.Quantity)},
									},
									Expression: model.ResultType(types.Date),
								},
							},
							High:          model.NewLiteral("@2022", types.Date),
							Expression:    model.ResultType(&types.Interval{PointType: types.Date}),
							LowInclusive:  true,
							HighInclusive: true,
						},
					},
					Expression: model.ResultType(types.Boolean),
				},
			},
		},
		{
			name: "BeforeOrAfterIntervalOperatorPhraseContext Interval Starts 1 Year Or Less On Or Before End Of Interval",
			cql:  "Interval[@2022, @2024] starts 1 year or less on or after end of Interval[@2020, @2022]",
			want: &model.In{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						&model.Start{
							UnaryExpression: &model.UnaryExpression{
								Operand:    model.NewInclusiveInterval("@2022", "@2024", types.Date),
								Expression: model.ResultType(types.Date),
							},
						},
						&model.Interval{
							Low: &model.End{
								UnaryExpression: &model.UnaryExpression{
									Operand:    model.NewInclusiveInterval("@2020", "@2022", types.Date),
									Expression: model.ResultType(types.Date),
								},
							},
							High: &model.Add{
								BinaryExpression: &model.BinaryExpression{
									Operands: []model.IExpression{
										&model.End{
											UnaryExpression: &model.UnaryExpression{
												Operand:    model.NewInclusiveInterval("@2020", "@2022", types.Date),
												Expression: model.ResultType(types.Date),
											},
										},
										&model.Quantity{Value: 1, Unit: "year", Expression: model.ResultType(types.Quantity)},
									},
									Expression: model.ResultType(types.Date),
								},
							},
							Expression:    model.ResultType(&types.Interval{PointType: types.Date}),
							LowInclusive:  true,
							HighInclusive: true,
						},
					},
					Expression: model.ResultType(types.Boolean),
				},
			},
		},
		{
			name: "BeforeOrAfterIntervalOperatorPhraseContext Date 1 Year Or More On Or After End Of Interval",
			cql:  "@2022 1 year or more on or after end of Interval[@2020, @2022]",
			want: &model.In{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("@2022", types.Date),
						&model.Interval{
							Low: &model.Add{
								BinaryExpression: &model.BinaryExpression{
									Operands: []model.IExpression{
										&model.End{
											UnaryExpression: &model.UnaryExpression{
												Operand:    model.NewInclusiveInterval("@2020", "@2022", types.Date),
												Expression: model.ResultType(types.Date),
											},
										},
										&model.Quantity{Value: 1, Unit: "year", Expression: model.ResultType(types.Quantity)},
									},
									Expression: model.ResultType(types.Date),
								},
							},
							High:          &model.MaxValue{ValueType: types.Date, Expression: model.ResultType(types.Date)},
							Expression:    model.ResultType(&types.Interval{PointType: types.Date}),
							LowInclusive:  true,
							HighInclusive: true,
						},
					},
					Expression: model.ResultType(types.Boolean),
				},
			},
		},
		{
			name: "BeforeOrAfterIntervalOperatorPhraseContext Date 1 Year Or More On Or Before Start Of Interval",
			cql:  "@2022 1 year or more on or before start of Interval[@2020, @2022]",
			want: &model.In{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("@2022", types.Date),
						&model.Interval{
							Low: &model.MinValue{ValueType: types.Date, Expression: model.ResultType(types.Date)},
							High: &model.Subtract{
								BinaryExpression: &model.BinaryExpression{
									Operands: []model.IExpression{
										&model.Start{
											UnaryExpression: &model.UnaryExpression{
												Operand:    model.NewInclusiveInterval("@2020", "@2022", types.Date),
												Expression: model.ResultType(types.Date),
											},
										},
										&model.Quantity{Value: 1, Unit: "year", Expression: model.ResultType(types.Quantity)},
									},
									Expression: model.ResultType(types.Date),
								},
							},
							Expression:    model.ResultType(&types.Interval{PointType: types.Date}),
							LowInclusive:  true,
							HighInclusive: true,
						},
					},
					Expression: model.ResultType(types.Boolean),
				},
			},
		},
		{
			name: "Difference in Months Between @2014-01-01 and @2014-02-01",
			cql:  "difference in months between @2014-01-01 and @2014-02-01",
			want: &model.DifferenceBetween{
				Precision: model.MONTH,
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("@2014-01-01", types.Date),
						model.NewLiteral("@2014-02-01", types.Date),
					},
					Expression: model.ResultType(types.Integer),
				},
			},
		},
		{
			name: "1 included in Interval[0, 5] returns in expresssion",
			cql:  "1 included in Interval[0, 5]",
			// Since left operator is of point type, model should be implicitly converted to model.In.
			want: &model.In{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("1", types.Integer),
						&model.Interval{
							Low:           model.NewLiteral("0", types.Integer),
							High:          model.NewLiteral("5", types.Integer),
							Expression:    model.ResultType(&types.Interval{PointType: types.Integer}),
							LowInclusive:  true,
							HighInclusive: true,
						},
					},
					Expression: model.ResultType(types.Boolean),
				},
			},
		},
		{
			name: "Interval[@2012, @2014] during Interval[@2010, @2020]",
			cql:  "Interval[@2012, @2014] during Interval[@2010, @2020]",
			want: &model.IncludedIn{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						&model.Interval{
							Low:           model.NewLiteral("@2012", types.Date),
							High:          model.NewLiteral("@2014", types.Date),
							Expression:    model.ResultType(&types.Interval{PointType: types.Date}),
							LowInclusive:  true,
							HighInclusive: true,
						},
						&model.Interval{
							Low:           model.NewLiteral("@2010", types.Date),
							High:          model.NewLiteral("@2020", types.Date),
							Expression:    model.ResultType(&types.Interval{PointType: types.Date}),
							LowInclusive:  true,
							HighInclusive: true,
						},
					},
					Expression: model.ResultType(types.Boolean),
				},
			},
		},
		{
			name: "IsNull",
			cql:  "null is null",
			want: &model.IsNull{
				UnaryExpression: &model.UnaryExpression{
					Operand:    model.NewLiteral("null", types.Any),
					Expression: model.ResultType(types.Boolean),
				},
			},
		},
		{
			name: "IsFalse",
			cql:  "null is false",
			want: &model.IsFalse{
				UnaryExpression: &model.UnaryExpression{
					Operand: &model.As{
						UnaryExpression: &model.UnaryExpression{
							Expression: model.ResultType(types.Boolean),
							Operand:    model.NewLiteral("null", types.Any),
						},
						AsTypeSpecifier: types.Boolean,
					},
					Expression: model.ResultType(types.Boolean),
				},
			},
		},
		{
			name: "IsTrue",
			cql:  "null is true",
			want: &model.IsTrue{
				UnaryExpression: &model.UnaryExpression{
					Operand: &model.As{
						UnaryExpression: &model.UnaryExpression{
							Expression: model.ResultType(types.Boolean),
							Operand:    model.NewLiteral("null", types.Any),
						},
						AsTypeSpecifier: types.Boolean,
					},
					Expression: model.ResultType(types.Boolean),
				},
			},
		},
		{
			name: "IsTrue with Not",
			cql:  "null is not true",
			want: &model.Not{
				UnaryExpression: &model.UnaryExpression{
					Expression: model.ResultType(types.Boolean),
					Operand: &model.IsTrue{
						UnaryExpression: &model.UnaryExpression{
							Operand: &model.As{
								UnaryExpression: &model.UnaryExpression{
									Expression: model.ResultType(types.Boolean),
									Operand:    model.NewLiteral("null", types.Any),
								},
								AsTypeSpecifier: types.Boolean,
							},
							Expression: model.ResultType(types.Boolean),
						},
					},
				},
			},
		},
		{
			name: "Less",
			cql:  "15 < 10",
			want: &model.Less{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("15", types.Integer),
						model.NewLiteral("10", types.Integer),
					},
					Expression: model.ResultType(types.Boolean),
				},
			},
		},
		{
			name: "Greater",
			cql:  "15 > 10",
			want: &model.Greater{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("15", types.Integer),
						model.NewLiteral("10", types.Integer),
					},
					Expression: model.ResultType(types.Boolean),
				},
			},
		},
		{
			name: "LessOrEqual",
			cql:  "15 <= 10",
			want: &model.LessOrEqual{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("15", types.Integer),
						model.NewLiteral("10", types.Integer),
					},
					Expression: model.ResultType(types.Boolean),
				},
			},
		},
		{
			name: "GreaterOrEqual",
			cql:  "15 >= 10",
			want: &model.GreaterOrEqual{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("15", types.Integer),
						model.NewLiteral("10", types.Integer),
					},
					Expression: model.ResultType(types.Boolean),
				},
			},
		},
		{
			name: "Equal",
			cql:  "1 = 1",
			want: &model.Equal{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("1", types.Integer),
						model.NewLiteral("1", types.Integer),
					},
					Expression: model.ResultType(types.Boolean),
				},
			},
		},
		{
			name: "Equivalent",
			cql:  "1 ~ 1",
			want: &model.Equivalent{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("1", types.Integer),
						model.NewLiteral("1", types.Integer),
					},
					Expression: model.ResultType(types.Boolean),
				},
			},
		},
		{
			name: "TimeBoundaryExpression End Of Interval",
			cql:  "end of Interval[@2013-01-01T00:00:00.0, @2014-01-01T00:00:00.0)",
			want: &model.End{
				UnaryExpression: &model.UnaryExpression{
					Operand: &model.Interval{
						Low:           model.NewLiteral("@2013-01-01T00:00:00.0", types.DateTime),
						High:          model.NewLiteral("@2014-01-01T00:00:00.0", types.DateTime),
						Expression:    model.ResultType(&types.Interval{PointType: types.DateTime}),
						LowInclusive:  true,
						HighInclusive: false,
					},
					Expression: model.ResultType(types.DateTime),
				},
			},
		},
		{
			name: "MembershipExpression In With Precision",
			cql:  "@2013-01-01T00:00:00.0 in year of Interval[@2013-01-01T00:00:00.0, @2014-01-01T00:00:00.0)",
			want: &model.In{
				Precision: model.YEAR,
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("@2013-01-01T00:00:00.0", types.DateTime),
						&model.Interval{
							Low:           model.NewLiteral("@2013-01-01T00:00:00.0", types.DateTime),
							High:          model.NewLiteral("@2014-01-01T00:00:00.0", types.DateTime),
							Expression:    model.ResultType(&types.Interval{PointType: types.DateTime}),
							LowInclusive:  true,
							HighInclusive: false,
						},
					},
					Expression: model.ResultType(types.Boolean),
				},
			},
		},
		{
			name: "MembershipExpression In Without Precision",
			cql:  "@2013-01-01T00:00:00.0 in Interval[@2013-01-01T00:00:00.0, @2014-01-01T00:00:00.0)",
			want: &model.In{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("@2013-01-01T00:00:00.0", types.DateTime),
						&model.Interval{
							Low:           model.NewLiteral("@2013-01-01T00:00:00.0", types.DateTime),
							High:          model.NewLiteral("@2014-01-01T00:00:00.0", types.DateTime),
							Expression:    model.ResultType(&types.Interval{PointType: types.DateTime}),
							LowInclusive:  true,
							HighInclusive: false,
						},
					},
					Expression: model.ResultType(types.Boolean),
				},
			},
		},
		{
			name: "MembershipExpression Contains With Precision",
			cql:  "Interval[@2013-01-01T00:00:00.0, @2014-01-01T00:00:00.0) contains year of @2013-01-01T00:00:00.0",
			want: &model.In{
				Precision: model.YEAR,
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("@2013-01-01T00:00:00.0", types.DateTime),
						&model.Interval{
							Low:           model.NewLiteral("@2013-01-01T00:00:00.0", types.DateTime),
							High:          model.NewLiteral("@2014-01-01T00:00:00.0", types.DateTime),
							Expression:    model.ResultType(&types.Interval{PointType: types.DateTime}),
							LowInclusive:  true,
							HighInclusive: false,
						},
					},
					Expression: model.ResultType(types.Boolean),
				},
			},
		},
		{
			name: "MembershipExpression Contains Without Precision",
			cql:  "Interval[@2013-01-01T00:00:00.0, @2014-01-01T00:00:00.0) contains @2013-01-01T00:00:00.0",
			want: &model.In{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("@2013-01-01T00:00:00.0", types.DateTime),
						&model.Interval{
							Low:           model.NewLiteral("@2013-01-01T00:00:00.0", types.DateTime),
							High:          model.NewLiteral("@2014-01-01T00:00:00.0", types.DateTime),
							Expression:    model.ResultType(&types.Interval{PointType: types.DateTime}),
							LowInclusive:  true,
							HighInclusive: false,
						},
					},
					Expression: model.ResultType(types.Boolean),
				},
			},
		},
		{
			name: "not",
			cql:  "not true",
			want: &model.Not{
				UnaryExpression: &model.UnaryExpression{
					Operand:    model.NewLiteral("true", types.Boolean),
					Expression: model.ResultType(types.Boolean),
				},
			},
		},
		{
			name: "And",
			cql:  "true and false",
			want: &model.And{
				BinaryExpression: &model.BinaryExpression{
					Expression: model.ResultType(types.Boolean),
					Operands: []model.IExpression{
						model.NewLiteral("true", types.Boolean),
						model.NewLiteral("false", types.Boolean),
					},
				},
			},
		},
		{
			name: "Or",
			cql:  "false or true",
			want: &model.Or{
				BinaryExpression: &model.BinaryExpression{
					Expression: model.ResultType(types.Boolean),
					Operands: []model.IExpression{
						model.NewLiteral("false", types.Boolean),
						model.NewLiteral("true", types.Boolean),
					},
				},
			},
		},
		{
			name: "XOr",
			cql:  "true xor false",
			want: &model.XOr{
				BinaryExpression: &model.BinaryExpression{
					Expression: model.ResultType(types.Boolean),
					Operands: []model.IExpression{
						model.NewLiteral("true", types.Boolean),
						model.NewLiteral("false", types.Boolean),
					},
				},
			},
		},
		{
			name: "Implies",
			cql:  "false implies true",
			want: &model.Implies{
				BinaryExpression: &model.BinaryExpression{
					Expression: model.ResultType(types.Boolean),
					Operands: []model.IExpression{
						model.NewLiteral("false", types.Boolean),
						model.NewLiteral("true", types.Boolean),
					},
				},
			},
		},
		{
			name: "Except",
			cql:  "{1, 4} except {'hi'}",
			want: &model.Except{
				BinaryExpression: &model.BinaryExpression{
					Expression: model.ResultType(&types.List{ElementType: types.Integer}),
					Operands: []model.IExpression{
						model.NewList([]string{"1", "4"}, types.Integer),
						model.NewList([]string{"hi"}, types.String),
					},
				},
			},
		},
		{
			name: "Intersect",
			cql:  "{1, 4} intersect {1}",
			want: &model.Intersect{
				BinaryExpression: &model.BinaryExpression{
					Expression: model.ResultType(&types.List{ElementType: types.Integer}),
					Operands: []model.IExpression{
						model.NewList([]string{"1", "4"}, types.Integer),
						model.NewList([]string{"1"}, types.Integer),
					},
				},
			},
		},
		{
			name: "Union",
			cql:  "{1, 4} union {1}",
			want: &model.Union{
				BinaryExpression: &model.BinaryExpression{
					Expression: model.ResultType(&types.List{ElementType: types.Integer}),
					Operands: []model.IExpression{
						model.NewList([]string{"1", "4"}, types.Integer),
						model.NewList([]string{"1"}, types.Integer),
					},
				},
			},
		},
		{
			name: "Union | syntax",
			cql:  "{1} | {1}",
			want: &model.Union{
				BinaryExpression: &model.BinaryExpression{
					Expression: model.ResultType(&types.List{ElementType: types.Integer}),
					Operands: []model.IExpression{
						model.NewList([]string{"1"}, types.Integer),
						model.NewList([]string{"1"}, types.Integer),
					},
				},
			},
		},
		{
			name: "SingletonFrom non-functional syntax",
			cql:  "singleton from {1}",
			want: &model.SingletonFrom{
				UnaryExpression: &model.UnaryExpression{
					Operand: &model.List{
						Expression: model.ResultType(&types.List{ElementType: types.Integer}),
						List: []model.IExpression{
							model.NewLiteral("1", types.Integer),
						},
					},
					Expression: model.ResultType(types.Integer),
				},
			},
		},
		{
			name: "Indexer [] syntax for List<T>",
			cql:  "{1, 2, 3}[0]",
			want: &model.Indexer{
				BinaryExpression: &model.BinaryExpression{
					Expression: model.ResultType(types.Integer),
					Operands: []model.IExpression{
						model.NewList([]string{"1", "2", "3"}, types.Integer),
						model.NewLiteral("0", types.Integer),
					},
				},
			},
		},
		{
			name: "Indexer [] syntax for String",
			cql:  "'abc'[0]",
			want: &model.Indexer{
				BinaryExpression: &model.BinaryExpression{
					Expression: model.ResultType(types.String),
					Operands: []model.IExpression{
						model.NewLiteral("abc", types.String),
						model.NewLiteral("0", types.Integer),
					},
				},
			},
		},
		{
			name: "Between Expression with Integers",
			cql:  "5 between 1 and 10",
			want: &model.And{
				BinaryExpression: &model.BinaryExpression{
					Expression: model.ResultType(types.Boolean),
					Operands: []model.IExpression{
						&model.GreaterOrEqual{
							BinaryExpression: &model.BinaryExpression{
								Expression: model.ResultType(types.Boolean),
								Operands: []model.IExpression{
									model.NewLiteral("5", types.Integer),
									model.NewLiteral("1", types.Integer),
								},
							},
						},
						&model.LessOrEqual{
							BinaryExpression: &model.BinaryExpression{
								Expression: model.ResultType(types.Boolean),
								Operands: []model.IExpression{
									model.NewLiteral("5", types.Integer),
									model.NewLiteral("10", types.Integer),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Between Expression with Dates",
			cql:  "@2015 between @2010 and @2020",
			want: &model.And{
				BinaryExpression: &model.BinaryExpression{
					Expression: model.ResultType(types.Boolean),
					Operands: []model.IExpression{
						&model.GreaterOrEqual{
							BinaryExpression: &model.BinaryExpression{
								Expression: model.ResultType(types.Boolean),
								Operands: []model.IExpression{
									model.NewLiteral("@2015", types.Date),
									model.NewLiteral("@2010", types.Date),
								},
							},
						},
						&model.LessOrEqual{
							BinaryExpression: &model.BinaryExpression{
								Expression: model.ResultType(types.Boolean),
								Operands: []model.IExpression{
									model.NewLiteral("@2015", types.Date),
									model.NewLiteral("@2020", types.Date),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Width Expression Term",
			cql:  "width of Interval[1, 10]",
			want: &model.Width{
				UnaryExpression: &model.UnaryExpression{
					Operand: &model.Interval{
						Low:           model.NewLiteral("1", types.Integer),
						High:          model.NewLiteral("10", types.Integer),
						Expression:    model.ResultType(&types.Interval{PointType: types.Integer}),
						LowInclusive:  true,
						HighInclusive: true,
					},
					Expression: model.ResultType(types.Integer),
				},
			},
		},
		{
			name: "Conversion Expression To Type",
			cql:  "convert 5 to String",
			want: &model.As{
				UnaryExpression: &model.UnaryExpression{
					Operand:    model.NewLiteral("5", types.Integer),
					Expression: model.ResultType(types.String),
				},
				AsTypeSpecifier: types.String,
				Strict:          true,
			},
		},
		{
			name: "Point Extractor Start Expression",
			cql:  "start of Interval[@2010-01-01, @2020-01-01]",
			want: &model.Start{
				UnaryExpression: &model.UnaryExpression{
					Operand: &model.Interval{
						Low:           model.NewLiteral("@2010-01-01", types.Date),
						High:          model.NewLiteral("@2020-01-01", types.Date),
						Expression:    model.ResultType(&types.Interval{PointType: types.Date}),
						LowInclusive:  true,
						HighInclusive: true,
					},
					Expression: model.ResultType(types.Date),
				},
			},
		},
		{
			name: "Point Extractor End Expression",
			cql:  "end of Interval[@2010-01-01, @2020-01-01]",
			want: &model.End{
				UnaryExpression: &model.UnaryExpression{
					Operand: &model.Interval{
						Low:           model.NewLiteral("@2010-01-01", types.Date),
						High:          model.NewLiteral("@2020-01-01", types.Date),
						Expression:    model.ResultType(&types.Interval{PointType: types.Date}),
						LowInclusive:  true,
						HighInclusive: true,
					},
					Expression: model.ResultType(types.Date),
				},
			},
		},
		{
			name: "Duration Expression Term",
			cql:  "duration in days of Interval[@2010-01-01, @2020-01-01]",
			want: &model.Duration{
				Precision: model.DAY,
				UnaryExpression: &model.UnaryExpression{
					Operand: &model.Interval{
						Low:           model.NewLiteral("@2010-01-01", types.Date),
						High:          model.NewLiteral("@2020-01-01", types.Date),
						Expression:    model.ResultType(&types.Interval{PointType: types.Date}),
						LowInclusive:  true,
						HighInclusive: true,
					},
					Expression: model.ResultType(types.Integer),
				},
			},
		},
		{
			name: "Duration Expression Term in Years",
			cql:  "duration in years of Interval[@2010-01-01, @2020-01-01]",
			want: &model.Duration{
				Precision: model.YEAR,
				UnaryExpression: &model.UnaryExpression{
					Operand: &model.Interval{
						Low:           model.NewLiteral("@2010-01-01", types.Date),
						High:          model.NewLiteral("@2020-01-01", types.Date),
						Expression:    model.ResultType(&types.Interval{PointType: types.Date}),
						LowInclusive:  true,
						HighInclusive: true,
					},
					Expression: model.ResultType(types.Integer),
				},
			},
		},
		{
			name: "Difference in days between two dates",
			cql:  "difference in days between @2010-01-01 and @2020-01-01",
			want: &model.DifferenceBetween{
				Precision: model.DAY,
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("@2010-01-01", types.Date),
						model.NewLiteral("@2020-01-01", types.Date),
					},
					Expression: model.ResultType(types.Integer),
				},
			},
		},
		{
			name: "Difference in days of interval",
			cql:  "difference in days of Interval[@2010-01-01, @2020-01-01]",
			want: &model.DifferenceBetween{
				Precision: model.DAY,
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						&model.Start{
							UnaryExpression: &model.UnaryExpression{
								Operand: &model.Interval{
									Low:           model.NewLiteral("@2010-01-01", types.Date),
									High:          model.NewLiteral("@2020-01-01", types.Date),
									Expression:    model.ResultType(&types.Interval{PointType: types.Date}),
									LowInclusive:  true,
									HighInclusive: true,
								},
								Expression: model.ResultType(types.Date),
							},
						},
						&model.End{
							UnaryExpression: &model.UnaryExpression{
								Operand: &model.Interval{
									Low:           model.NewLiteral("@2010-01-01", types.Date),
									High:          model.NewLiteral("@2020-01-01", types.Date),
									Expression:    model.ResultType(&types.Interval{PointType: types.Date}),
									LowInclusive:  true,
									HighInclusive: true,
								},
								Expression: model.ResultType(types.Date),
							},
						},
					},
					Expression: model.ResultType(types.Integer),
				},
			},
		},
		{
			name: "Duration Between Expression",
			cql:  "duration in days between @2010-01-01 and @2020-01-01",
			want: &model.DurationBetween{
				Precision: model.DAY,
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("@2010-01-01", types.Date),
						model.NewLiteral("@2020-01-01", types.Date),
					},
					Expression: model.ResultType(types.Integer),
				},
			},
		},
		{
			name: "Duration Between Expression in Months",
			cql:  "duration in months between @2010-01-01 and @2020-01-01",
			want: &model.DurationBetween{
				Precision: model.MONTH,
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("@2010-01-01", types.Date),
						model.NewLiteral("@2020-01-01", types.Date),
					},
					Expression: model.ResultType(types.Integer),
				},
			},
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

func TestOperatorExpressions_Errors(t *testing.T) {
	tests := []struct {
		name        string
		cql         string
		errContains []string
		errCount    int
	}{
		{
			name:        "invalid addition overload",
			cql:         `2 + @2024`,
			errContains: []string{"could not resolve Add(System.Integer, System.Date)"},
			errCount:    1,
		},
		{
			name:        "invalid division overload",
			cql:         `2 / @2024`,
			errContains: []string{"could not resolve Divide(System.Integer, System.Date)"},
			errCount:    1,
		},
		{
			name:        "intersect with no common types",
			cql:         `{3} intersect {@2024}`,
			errContains: []string{"no common types between System.Integer and System.Date"},
			errCount:    1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := newFHIRParser(t).Libraries(context.Background(), wrapInLib(t, test.cql), Config{})
			if err == nil {
				t.Fatal("Parsing succeeded, expected error")
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

func TestClinicalOperatorExpressions(t *testing.T) {
	tests := []struct {
		name string
		cql  string
		want model.IExpression
	}{
		{
			name: "Code in CodeSystem",
			cql: dedent.Dedent(`
			codesystem cs: 'ExampleCodeSystem'
			code c: '1234' from cs
			using FHIR version '4.0.1'
			context Patient
			define TESTRESULT: c in cs`),
			want: &model.InCodeSystem{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						&model.CodeRef{Name: "c", Expression: model.ResultType(types.Code)},
						&model.CodeSystemRef{Name: "cs", Expression: model.ResultType(types.CodeSystem)},
					},
					Expression: &model.Expression{Element: &model.Element{ResultType: types.Boolean}},
				},
			},
		},
		{
			name: "InCodeSystem(Code, CodeSystemRef)",
			cql: dedent.Dedent(`
			codesystem cs: 'ExampleCodeSystem'
			code c: '1234' from cs
			using FHIR version '4.0.1'
			context Patient
			define TESTRESULT: InCodeSystem(c, cs)`),
			want: &model.InCodeSystem{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						&model.CodeRef{Name: "c", Expression: model.ResultType(types.Code)},
						&model.CodeSystemRef{Name: "cs", Expression: model.ResultType(types.CodeSystem)},
					},
					Expression: &model.Expression{Element: &model.Element{ResultType: types.Boolean}},
				},
			},
		},
		{
			name: "Code in ValueSet",
			cql: dedent.Dedent(`
			codesystem cs: 'ExampleCodeSystem'
			valueset vs: 'ExampleValueset'
			code c: '1234' from cs
			using FHIR version '4.0.1'
			context Patient
			define TESTRESULT: c in vs`),
			want: &model.InValueSet{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						&model.CodeRef{Name: "c", Expression: model.ResultType(types.Code)},
						&model.ValuesetRef{Name: "vs", Expression: model.ResultType(types.ValueSet)},
					},
					Expression: &model.Expression{Element: &model.Element{ResultType: types.Boolean}},
				},
			},
		},
		{
			name: "InValueSet(Code, ValueSetRef)",
			cql: dedent.Dedent(`
			codesystem cs: 'ExampleCodeSystem'
			valueset vs: 'ExampleValueset'
			code c: '1234' from cs
			using FHIR version '4.0.1'
			context Patient
			define TESTRESULT: InValueSet(c, vs)`),
			want: &model.InValueSet{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						&model.CodeRef{Name: "c", Expression: model.ResultType(types.Code)},
						&model.ValuesetRef{Name: "vs", Expression: model.ResultType(types.ValueSet)},
					},
					Expression: &model.Expression{Element: &model.Element{ResultType: types.Boolean}},
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
			if diff := cmp.Diff(test.want, getTESTRESULTModel(t, parsedLibs)); diff != "" {
				t.Errorf("Parse diff (-want +got):\n%s", diff)
			}
		})
	}
}
