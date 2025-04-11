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

package parser

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/cql/internal/convert"
	"github.com/google/cql/model"
	"github.com/google/cql/types"
	"github.com/antlr4-go/antlr/v4"
)

// parseFunction uses the reference resolver to resolve the function, visits the operands, and sets
// the operands in the model. For some built-in functions it also sets the ResultType. libraryName
// should be an empty string for local functions.
func (v *visitor) parseFunction(libraryName, funcName string, ctxOperands []antlr.Tree, calledFluently bool) (model.IExpression, error) {
	operands := make([]model.IExpression, 0)
	for _, expr := range ctxOperands {
		o := v.VisitExpression(expr)
		operands = append(operands, o)
	}
	return v.resolveFunction(libraryName, funcName, operands, calledFluently)
}

// resolveFunction uses the reference resolver to resolve the function, perform implicit
// conversions, and set the wrapped operands in the model.IExpression. For some built-in functions
// it also sets the ResultType. libraryName should be an empty string for local functions.
// This function takes operands as []model.IExpression, whereas parseFunction takes un parsed antlr
// trees as operands.
func (v *visitor) resolveFunction(libraryName, funcName string, operands []model.IExpression, calledFluently bool) (model.IExpression, error) {
	var resolved *convert.MatchedOverload[func() model.IExpression]
	var err error
	if libraryName != "" {
		resolved, err = v.refs.ResolveGlobalFunc(libraryName, funcName, operands, calledFluently, v.modelInfo)
	} else {
		resolved, err = v.refs.ResolveLocalFunc(funcName, operands, calledFluently, v.modelInfo)
	}
	if err != nil {
		return nil, err
	}

	// Handle special cases such as setting the result type based on an operand.
	r := resolved.Result()
	switch t := r.(type) {
	case *model.Coalesce:
		return v.parseCoalesce(t, resolved.WrappedOperands)
	case *model.Contains:
		// If we reverse the operands we can treat contains as an In.
		contains := t
		r = &model.In{
			Precision:        contains.Precision,
			BinaryExpression: &model.BinaryExpression{Expression: contains.Expression},
		}
		resolved.WrappedOperands[0], resolved.WrappedOperands[1] = resolved.WrappedOperands[1], resolved.WrappedOperands[0]
	case *model.Message:
		if len(resolved.WrappedOperands) != 5 {
			return nil, errors.New("internal error - resolving message function returned incorrect argument")
		}
		t.Source = resolved.WrappedOperands[0]
		t.Condition = resolved.WrappedOperands[1]
		t.Code = resolved.WrappedOperands[2]
		t.Severity = resolved.WrappedOperands[3]
		t.Message = resolved.WrappedOperands[4]
		t.Expression = model.ResultType(resolved.WrappedOperands[0].GetResultType())
	case *model.Except:
		// For Except the left side is the result type.
		t.Expression = model.ResultType(resolved.WrappedOperands[0].GetResultType())
	case *model.Intersect:
		listTypeLeft := resolved.WrappedOperands[0].GetResultType().(*types.List)
		listTypeRight := resolved.WrappedOperands[1].GetResultType().(*types.List)
		listElemType, err := convert.Intersect(listTypeLeft.ElementType, listTypeRight.ElementType)
		if err != nil {
			return nil, err
		}
		t.Expression = model.ResultType(&types.List{ElementType: listElemType})
	case *model.Avg:
		listType := resolved.WrappedOperands[0].GetResultType().(*types.List)
		t.Expression = model.ResultType(listType.ElementType)
	case *model.Flatten:
		nestedListType := resolved.WrappedOperands[0].GetResultType().(*types.List).ElementType.(*types.List)
		t.Expression = model.ResultType(nestedListType)
	case *model.Max:
		listType := resolved.WrappedOperands[0].GetResultType().(*types.List)
		t.Expression = model.ResultType(listType.ElementType)
	case *model.Min:
		listType := resolved.WrappedOperands[0].GetResultType().(*types.List)
		t.Expression = model.ResultType(listType.ElementType)
	case *model.Sum:
		listType := resolved.WrappedOperands[0].GetResultType().(*types.List)
		t.Expression = model.ResultType(listType.ElementType)
	case *model.Skip:
		listType := resolved.WrappedOperands[0].GetResultType().(*types.List)
		t.Expression = model.ResultType(&types.List{ElementType: listType.ElementType})
	case *model.Tail:
		listType := resolved.WrappedOperands[0].GetResultType().(*types.List)
		t.Expression = model.ResultType(&types.List{ElementType: listType.ElementType})
	case *model.Take:
		listType := resolved.WrappedOperands[0].GetResultType().(*types.List)
		t.Expression = model.ResultType(&types.List{ElementType: listType.ElementType})
	case *model.Union:
		listTypeLeft := resolved.WrappedOperands[0].GetResultType().(*types.List)
		listTypeRight := resolved.WrappedOperands[1].GetResultType().(*types.List)
		listElemType, err := convert.DeDuplicate([]types.IType{listTypeLeft.ElementType, listTypeRight.ElementType})
		if err != nil {
			return nil, err
		}
		t.Expression = model.ResultType(&types.List{ElementType: listElemType})
	case *model.End:
		pointType := resolved.WrappedOperands[0].GetResultType().(*types.Interval)
		t.Expression = model.ResultType(pointType.PointType)
	case *model.Start:
		pointType := resolved.WrappedOperands[0].GetResultType().(*types.Interval)
		t.Expression = model.ResultType(pointType.PointType)
	case *model.First:
		// First(List<T>) T is a special case because the ResultType is not known until invocation.
		listType := resolved.WrappedOperands[0].GetResultType().(*types.List)
		t.Expression = model.ResultType(listType.ElementType)
	case *model.Last:
		// Last(List<T>) T is a special case because the ResultType is not known until invocation.
		listType := resolved.WrappedOperands[0].GetResultType().(*types.List)
		t.Expression = model.ResultType(listType.ElementType)
	case *model.Distinct:
		// Distinct(List<T>) List<T> is a special case because the ResultType is not known until invocation.
		listType := resolved.WrappedOperands[0].GetResultType().(*types.List)
		t.Expression = model.ResultType(&types.List{ElementType: listType.ElementType})
	case *model.Indexer:
		switch opType := resolved.WrappedOperands[0].GetResultType().(type) {
		case types.System:
			if opType != types.String {
				return nil, fmt.Errorf("internal error - expected Indexer(String, Integer) overload during parsing, but got Indexer(%v, _)", resolved.WrappedOperands[0].GetResultType())
			}
			t.Expression = model.ResultType(types.String)
		case *types.List:
			t.Expression = model.ResultType(opType.ElementType)
		default:
			return nil, fmt.Errorf("internal error -- upsupported Indexer operand types")
		}
	case *model.Predecessor:
		t.Expression = model.ResultType(resolved.WrappedOperands[0].GetResultType())
	case *model.Successor:
		t.Expression = model.ResultType(resolved.WrappedOperands[0].GetResultType())
	case *model.SingletonFrom:
		// SingletonFrom(List<T>) T is a special case because the ResultType is not known until invocation.
		listType := resolved.WrappedOperands[0].GetResultType().(*types.List)
		t.Expression = model.ResultType(listType.ElementType)
	case *model.CalculateAge:
		// AgeInYears() is a special case as it takes 0 operands but the model.CalculateAge has 1
		// operand, the patient's birthday.
		bday, err := v.patientBirthDateExpression()
		if err != nil {
			return nil, err
		}

		// Currently the FHIR modelinfo Patient Birthday Expression returns System.Date. However, in
		// case in the future it returns something else, try to convert to System.DateTime to match the
		// AgeInYears(DateTime) overload.
		res, err := convert.OperandImplicitConverter(bday.GetResultType(), types.DateTime, bday, v.modelInfo)
		if err != nil {
			return nil, err
		}
		if !res.Matched {
			return nil, fmt.Errorf("internal error - could not implicitly convert the Patient Birthday Expression of type %v to %v", bday.GetResultType(), types.DateTime)
		}

		resolved.WrappedOperands = []model.IExpression{res.WrappedOperand}
	case *model.CalculateAgeAt:
		if len(resolved.WrappedOperands) == 1 {
			// AgeInYearsAt(asOf Date) is a special case as it takes 1 operand but the
			// model.CalculateAgeAt has 2 operand, the patient's birthday.
			bday, err := v.patientBirthDateExpression()
			if err != nil {
				return nil, err
			}

			// For FHIR modelinfo the Patient Birthday Expression should be System.Date so we may need to
			// convert it to DateTime to match the AgeInYearsAt(DateTime, DateTime) overload.
			res, err := convert.OperandImplicitConverter(bday.GetResultType(), resolved.WrappedOperands[0].GetResultType(), bday, v.modelInfo)
			if err != nil {
				return nil, err
			}
			if !res.Matched {
				return nil, fmt.Errorf("internal error - could not implicitly convert the Patient Birthday Expression of type %v to %v", bday.GetResultType(), resolved.WrappedOperands[0].GetResultType())
			}

			// The operands should be AgeInYearsAt(convertedBirthDate)
			resolved.WrappedOperands = []model.IExpression{res.WrappedOperand, resolved.WrappedOperands[0]}
		}
	case *model.Median:
		listType := resolved.WrappedOperands[0].GetResultType().(*types.List)
		t.Expression = model.ResultType(listType.ElementType)
	case *model.PopulationStdDev:
		listType := resolved.WrappedOperands[0].GetResultType().(*types.List)
		t.Expression = model.ResultType(listType.ElementType)
	}

	// Set Operands.
	switch t := r.(type) {
	case model.IUnaryExpression:
		t.SetOperand(resolved.WrappedOperands[0])
	case model.IBinaryExpression:
		t.SetOperands(resolved.WrappedOperands[0], resolved.WrappedOperands[1])
		return r, nil
	case model.INaryExpression:
		t.SetOperands(resolved.WrappedOperands)
	case *model.FunctionRef:
		t.LibraryName = libraryName
		t.Operands = resolved.WrappedOperands
		return r, nil
	case *model.Message:
		// Message is not a function or a *nary expression but extends
		// Expression directly
		return r, nil
	default:
		return nil, errors.New("internal error - resolving function returned an unsupported type")
	}

	return r, nil
}

// loadSystemOperators defines all CQL System Operators in the reference resolver. The operands
// are not set here, but are instead set when we parse the function invocation in VisitFunction. For
// some System Operators like Last(List<T>) T we also set the return type in VisitFunction as the
// return type is not known until the function is invoked.
func (p *Parser) loadSystemOperators() error {
	systemOperators := []struct {
		// name is the function name.
		name string
		// operands holds all of the overloads for this function name. For example:
		// Foo(A Integer, B Integer)
		// Foo(A Integer, B Long)
		// Operands: [[Integer, Integer], [Integer, Long]]
		operands [][]types.IType
		model    func() model.IExpression
	}{
		// LOGICAL OPERATORS - https://cql.hl7.org/09-b-cqlreference.html#logical-operators-3
		{
			name: "And",
			operands: [][]types.IType{
				{types.Boolean, types.Boolean},
			},
			model: func() model.IExpression {
				return &model.And{
					BinaryExpression: &model.BinaryExpression{
						Expression: model.ResultType(types.Boolean),
					},
				}
			},
		},
		{
			name: "Implies",
			operands: [][]types.IType{
				{types.Boolean, types.Boolean},
			},
			model: func() model.IExpression {
				return &model.Implies{
					BinaryExpression: &model.BinaryExpression{
						Expression: model.ResultType(types.Boolean),
					},
				}
			},
		},
		{
			name: "Not",
			operands: [][]types.IType{
				{types.Boolean},
			},
			model: func() model.IExpression {
				return &model.Not{
					UnaryExpression: &model.UnaryExpression{
						Expression: model.ResultType(types.Boolean),
					},
				}
			},
		},
		{
			name: "Or",
			operands: [][]types.IType{
				{types.Boolean, types.Boolean},
			},
			model: func() model.IExpression {
				return &model.Or{
					BinaryExpression: &model.BinaryExpression{
						Expression: model.ResultType(types.Boolean),
					},
				}
			},
		},
		{
			name: "Xor",
			operands: [][]types.IType{
				{types.Boolean, types.Boolean},
			},
			model: func() model.IExpression {
				return &model.XOr{
					BinaryExpression: &model.BinaryExpression{
						Expression: model.ResultType(types.Boolean),
					},
				}
			},
		},
		// TYPE OPERATORS - https://cql.hl7.org/09-b-cqlreference.html#type-operators-1
		{
			name: "CanConvertQuantity",
			operands: [][]types.IType{
				{types.Quantity, types.String},
			},
			model: func() model.IExpression {
				return &model.CanConvertQuantity{
					BinaryExpression: &model.BinaryExpression{
						Expression: model.ResultType(types.Boolean),
					},
				}
			},
		},
		{
			name: "ToBoolean",
			operands: [][]types.IType{
				{types.Boolean},
				{types.Decimal},
				{types.Long},
				{types.Integer},
				{types.String}},
			model: func() model.IExpression {
				return &model.ToBoolean{
					UnaryExpression: &model.UnaryExpression{
						Expression: model.ResultType(types.Boolean),
					},
				}
			},
		},
		{
			name: "ToDateTime",
			operands: [][]types.IType{
				{types.DateTime},
				{types.Date},
				{types.String}},
			model: func() model.IExpression {
				return &model.ToDateTime{
					UnaryExpression: &model.UnaryExpression{
						Expression: model.ResultType(types.DateTime),
					},
				}
			},
		},
		{
			name: "ToDate",
			operands: [][]types.IType{
				{types.Date},
				{types.DateTime},
				{types.String}},
			model: func() model.IExpression {
				return &model.ToDate{
					UnaryExpression: &model.UnaryExpression{
						Expression: model.ResultType(types.Date),
					},
				}
			},
		},
		{
			name: "ToDecimal",
			operands: [][]types.IType{
				{types.Decimal},
				{types.Long},
				{types.Integer},
				{types.String},
				{types.Boolean}},
			model: func() model.IExpression {
				return &model.ToDecimal{
					UnaryExpression: &model.UnaryExpression{
						Expression: model.ResultType(types.Decimal),
					},
				}
			},
		},
		{
			name: "ToLong",
			operands: [][]types.IType{
				{types.Long},
				{types.Integer},
				{types.String},
				{types.Boolean}},
			model: func() model.IExpression {
				return &model.ToLong{
					UnaryExpression: &model.UnaryExpression{
						Expression: model.ResultType(types.Long),
					},
				}
			},
		},
		{
			name: "ToInteger",
			operands: [][]types.IType{
				{types.Integer},
				{types.Long},
				{types.String},
				{types.Boolean}},
			model: func() model.IExpression {
				return &model.ToInteger{
					UnaryExpression: &model.UnaryExpression{
						Expression: model.ResultType(types.Integer),
					},
				}
			},
		},
		{
			name: "ToQuantity",
			operands: [][]types.IType{
				{types.Quantity},
				{types.Decimal},
				{types.Integer},
				{types.Ratio},
				{types.String}},
			model: func() model.IExpression {
				return &model.ToQuantity{
					UnaryExpression: &model.UnaryExpression{
						Expression: model.ResultType(types.Quantity),
					},
				}
			},
		},
		{
			name: "ToConcept",
			operands: [][]types.IType{
				{types.Code},
				{&types.List{ElementType: types.Code}}},
			model: func() model.IExpression {
				return &model.ToConcept{
					UnaryExpression: &model.UnaryExpression{
						Expression: model.ResultType(types.Concept),
					},
				}
			},
		},
		{
			name: "ToString",
			operands: [][]types.IType{
				{types.Any},
				{types.String},
				{types.Integer},
				{types.Long},
				{types.Decimal},
				{types.Quantity},
				{types.Ratio},
				{types.Date},
				{types.DateTime},
				{types.Time},
				{types.Boolean}},
			model: func() model.IExpression {
				return &model.ToString{
					UnaryExpression: &model.UnaryExpression{
						Expression: model.ResultType(types.String),
					},
				}
			},
		},
		{
			name: "ToTime",
			operands: [][]types.IType{
				{types.Time},
				{types.String}},
			model: func() model.IExpression {
				return &model.ToTime{
					UnaryExpression: &model.UnaryExpression{
						Expression: model.ResultType(types.Time),
					},
				}
			},
		},
		// NULLOGICAL OPERATORS - https://cql.hl7.org/09-b-cqlreference.html#nullological-operators-3
		{
			name: "Coalesce",
			operands: [][]types.IType{
				{convert.GenericType, convert.GenericType},
				{convert.GenericType, convert.GenericType, convert.GenericType},
				{convert.GenericType, convert.GenericType, convert.GenericType, convert.GenericType},
				{convert.GenericType, convert.GenericType, convert.GenericType, convert.GenericType, convert.GenericType},
				{&types.List{ElementType: types.Any}},
			},
			model: func() model.IExpression {
				return &model.Coalesce{
					NaryExpression: &model.NaryExpression{},
				}
			},
		},
		{
			name: "IsNull",
			operands: [][]types.IType{
				{types.Any},
			},
			model: func() model.IExpression {
				return &model.IsNull{
					UnaryExpression: &model.UnaryExpression{
						Expression: model.ResultType(types.Boolean),
					},
				}
			},
		},
		{
			name: "IsFalse",
			operands: [][]types.IType{
				{types.Boolean},
			},
			model: func() model.IExpression {
				return &model.IsFalse{
					UnaryExpression: &model.UnaryExpression{
						Expression: model.ResultType(types.Boolean),
					},
				}
			},
		},
		{
			name: "IsTrue",
			operands: [][]types.IType{
				{types.Boolean},
			},
			model: func() model.IExpression {
				return &model.IsTrue{
					UnaryExpression: &model.UnaryExpression{
						Expression: model.ResultType(types.Boolean),
					},
				}
			},
		},
		// COMPARISON OPERATORS - https://cql.hl7.org/09-b-cqlreference.html#comparison-operators-4
		{
			name: "Equal",
			operands: [][]types.IType{
				{convert.GenericType, convert.GenericType},
			},
			model: func() model.IExpression {
				return &model.Equal{
					BinaryExpression: &model.BinaryExpression{
						Expression: model.ResultType(types.Boolean),
					},
				}
			},
		},
		{
			name: "Equivalent",
			operands: [][]types.IType{
				{convert.GenericType, convert.GenericType},
				// The following overloads come from
				// CLINICAL OPERATORS - https://cql.hl7.org/09-b-cqlreference.html#equivalent-3
				{types.Concept, types.Code},
			},
			model: func() model.IExpression {
				return &model.Equivalent{
					BinaryExpression: &model.BinaryExpression{
						Expression: model.ResultType(types.Boolean),
					},
				}
			},
		},
		{
			name: "Less",
			operands: [][]types.IType{
				[]types.IType{types.Integer, types.Integer},
				[]types.IType{types.Long, types.Long},
				[]types.IType{types.Decimal, types.Decimal},
				[]types.IType{types.Quantity, types.Quantity},
				[]types.IType{types.Date, types.Date},
				[]types.IType{types.DateTime, types.DateTime},
				[]types.IType{types.Time, types.Time},
				[]types.IType{types.String, types.String},
			},
			model: func() model.IExpression {
				return &model.Less{
					BinaryExpression: &model.BinaryExpression{
						Expression: model.ResultType(types.Boolean),
					},
				}
			},
		},
		{
			name: "Greater",
			operands: [][]types.IType{
				[]types.IType{types.Integer, types.Integer},
				[]types.IType{types.Long, types.Long},
				[]types.IType{types.Decimal, types.Decimal},
				[]types.IType{types.Quantity, types.Quantity},
				[]types.IType{types.Date, types.Date},
				[]types.IType{types.DateTime, types.DateTime},
				[]types.IType{types.Time, types.Time},
				[]types.IType{types.String, types.String},
			},
			model: func() model.IExpression {
				return &model.Greater{
					BinaryExpression: &model.BinaryExpression{
						Expression: model.ResultType(types.Boolean),
					},
				}
			},
		},
		{
			name: "LessOrEqual",
			operands: [][]types.IType{
				[]types.IType{types.Integer, types.Integer},
				[]types.IType{types.Long, types.Long},
				[]types.IType{types.Decimal, types.Decimal},
				[]types.IType{types.Quantity, types.Quantity},
				[]types.IType{types.Date, types.Date},
				[]types.IType{types.DateTime, types.DateTime},
				[]types.IType{types.Time, types.Time},
				[]types.IType{types.String, types.String},
			},
			model: func() model.IExpression {
				return &model.LessOrEqual{
					BinaryExpression: &model.BinaryExpression{
						Expression: model.ResultType(types.Boolean),
					},
				}
			},
		},
		{
			name: "GreaterOrEqual",
			operands: [][]types.IType{
				[]types.IType{types.Integer, types.Integer},
				[]types.IType{types.Long, types.Long},
				[]types.IType{types.Decimal, types.Decimal},
				[]types.IType{types.Quantity, types.Quantity},
				[]types.IType{types.Date, types.Date},
				[]types.IType{types.DateTime, types.DateTime},
				[]types.IType{types.Time, types.Time},
				[]types.IType{types.String, types.String},
			},
			model: func() model.IExpression {
				return &model.GreaterOrEqual{
					BinaryExpression: &model.BinaryExpression{
						Expression: model.ResultType(types.Boolean),
					},
				}
			},
		},
		// ARITHMETIC OPERATORS - https://cql.hl7.org/09-b-cqlreference.html#arithmetic-operators-4
		{
			name:     "Abs",
			operands: [][]types.IType{{types.Decimal}},
			model:    absModel(types.Decimal),
		},
		{
			name:     "Abs",
			operands: [][]types.IType{{types.Integer}},
			model:    absModel(types.Integer),
		},
		{
			name:     "Abs",
			operands: [][]types.IType{{types.Long}},
			model:    absModel(types.Long),
		},
		{
			name:     "Abs",
			operands: [][]types.IType{{types.Quantity}},
			model:    absModel(types.Quantity),
		},
		{
			name:     "Add",
			operands: [][]types.IType{{types.Integer, types.Integer}},
			model:    addModel(types.Integer),
		},
		{
			name:     "Add",
			operands: [][]types.IType{{types.Long, types.Long}},
			model:    addModel(types.Long),
		},
		{
			name:     "Add",
			operands: [][]types.IType{{types.Decimal, types.Decimal}},
			model:    addModel(types.Decimal),
		},
		{
			name:     "Add",
			operands: [][]types.IType{{types.Quantity, types.Quantity}},
			model:    addModel(types.Quantity),
		},
		{
			name:     "Ceiling",
			operands: [][]types.IType{{types.Decimal}},
			model: func() model.IExpression {
				return &model.Ceiling{
					UnaryExpression: &model.UnaryExpression{
						Expression: model.ResultType(types.Integer),
					},
				}
			},
		},
		{
			name:     "Exp",
			operands: [][]types.IType{{types.Decimal}},
			model: func() model.IExpression {
				return &model.Exp{
					UnaryExpression: &model.UnaryExpression{
						Expression: model.ResultType(types.Decimal),
					},
				}
			},
		},
		// TODO: b/301606416 - Add support for Exp with Quantities, current behavior is ambiguous.
		{
			name:     "Floor",
			operands: [][]types.IType{{types.Decimal}},
			model: func() model.IExpression {
				return &model.Floor{
					UnaryExpression: &model.UnaryExpression{
						Expression: model.ResultType(types.Integer),
					},
				}
			},
		},
		{
			name:     "Ln",
			operands: [][]types.IType{{types.Decimal}},
			model: func() model.IExpression {
				return &model.Ln{
					UnaryExpression: &model.UnaryExpression{
						Expression: model.ResultType(types.Decimal),
					},
				}
			},
		},
		{
			name:     "Log",
			operands: [][]types.IType{{types.Decimal, types.Decimal}},
			model: func() model.IExpression {
				return &model.Log{
					BinaryExpression: &model.BinaryExpression{
						Expression: model.ResultType(types.Decimal),
					},
				}
			},
		},
		{
			name:     "Negate",
			operands: [][]types.IType{{types.Integer}},
			model:    negateModel(types.Integer),
		},
		{
			name:     "Negate",
			operands: [][]types.IType{{types.Long}},
			model:    negateModel(types.Long),
		},
		{
			name:     "Negate",
			operands: [][]types.IType{{types.Decimal}},
			model:    negateModel(types.Decimal),
		},
		{
			name:     "Negate",
			operands: [][]types.IType{{types.Quantity}},
			model:    negateModel(types.Quantity),
		},
		{
			name:     "Precision",
			operands: [][]types.IType{{types.Date}},
			model:    precisionModel(),
		},
		{
			name:     "Precision",
			operands: [][]types.IType{{types.DateTime}},
			model:    precisionModel(),
		},
		{
			name:     "Precision",
			operands: [][]types.IType{{types.Time}},
			model:    precisionModel(),
		},
		{
			name:     "Subtract",
			operands: [][]types.IType{{types.Integer, types.Integer}},
			model:    subtractModel(types.Integer),
		},
		{
			name:     "Subtract",
			operands: [][]types.IType{{types.Long, types.Long}},
			model:    subtractModel(types.Long),
		},
		{
			name:     "Subtract",
			operands: [][]types.IType{{types.Decimal, types.Decimal}},
			model:    subtractModel(types.Decimal),
		},
		{
			name:     "Subtract",
			operands: [][]types.IType{{types.Quantity, types.Quantity}},
			model:    subtractModel(types.Quantity),
		},
		{
			name:     "Multiply",
			operands: [][]types.IType{{types.Integer, types.Integer}},
			model:    multiplyModel(types.Integer),
		},
		{
			name:     "Multiply",
			operands: [][]types.IType{{types.Long, types.Long}},
			model:    multiplyModel(types.Long),
		},
		{
			name:     "Multiply",
			operands: [][]types.IType{{types.Decimal, types.Decimal}},
			model:    multiplyModel(types.Decimal),
		},
		{
			name:     "Multiply",
			operands: [][]types.IType{{types.Quantity, types.Quantity}},
			model:    multiplyModel(types.Quantity),
		},
		{
			name:     "Divide",
			operands: [][]types.IType{{types.Decimal, types.Decimal}},
			model:    divideModel(types.Decimal),
		},
		{
			name:     "Divide",
			operands: [][]types.IType{{types.Quantity, types.Quantity}},
			model:    divideModel(types.Quantity),
		},
		{
			name:     "Truncate",
			operands: [][]types.IType{{types.Decimal}},
			model: func() model.IExpression {
				return &model.Truncate{
					UnaryExpression: &model.UnaryExpression{
						Expression: model.ResultType(types.Integer),
					},
				}
			},
		},
		{
			name:     "TruncatedDivide",
			operands: [][]types.IType{{types.Integer, types.Integer}},
			model:    truncatedDivideModel(types.Integer),
		},
		{
			name:     "TruncatedDivide",
			operands: [][]types.IType{{types.Long, types.Long}},
			model:    truncatedDivideModel(types.Long),
		},
		{
			name:     "TruncatedDivide",
			operands: [][]types.IType{{types.Decimal, types.Decimal}},
			model:    truncatedDivideModel(types.Decimal),
		},
		{
			name:     "TruncatedDivide",
			operands: [][]types.IType{{types.Quantity, types.Quantity}},
			model:    truncatedDivideModel(types.Quantity),
		},
		{
			name:     "Modulo",
			operands: [][]types.IType{{types.Integer, types.Integer}},
			model:    modModel(types.Integer),
		},
		{
			name:     "Modulo",
			operands: [][]types.IType{{types.Long, types.Long}},
			model:    modModel(types.Long),
		},
		{
			name:     "Modulo",
			operands: [][]types.IType{{types.Decimal, types.Decimal}},
			model:    modModel(types.Decimal),
		},
		{
			name:     "Modulo",
			operands: [][]types.IType{{types.Quantity, types.Quantity}},
			model:    modModel(types.Quantity),
		},
		{
			name:     "Power",
			operands: [][]types.IType{{types.Integer, types.Integer}},
			model:    powerModel(types.Integer),
		},
		{
			name:     "Power",
			operands: [][]types.IType{{types.Long, types.Long}},
			model:    powerModel(types.Long),
		},
		{
			name:     "Power",
			operands: [][]types.IType{{types.Decimal, types.Decimal}},
			model:    powerModel(types.Decimal),
		},
		{
			name: "Predecessor",
			operands: [][]types.IType{
				{types.Integer},
				{types.Long},
				{types.Decimal},
				{types.Quantity},
				{types.Date},
				{types.DateTime},
				{types.Time},
			},
			model: func() model.IExpression {
				return &model.Predecessor{
					UnaryExpression: &model.UnaryExpression{},
				}
			},
		},
		{
			name:     "Round",
			operands: [][]types.IType{{types.Decimal, types.Integer}},
			model: func() model.IExpression {
				return &model.Round{
					NaryExpression: &model.NaryExpression{
						Expression: model.ResultType(types.Decimal),
					},
				}
			},
		},
		{
			name:     "Round",
			operands: [][]types.IType{{types.Decimal}},
			model: func() model.IExpression {
				return &model.Round{
					NaryExpression: &model.NaryExpression{
						Expression: model.ResultType(types.Decimal),
					},
				}
			},
		},
		{
			name: "Successor",
			operands: [][]types.IType{
				{types.Integer},
				{types.Long},
				{types.Decimal},
				{types.Quantity},
				{types.Date},
				{types.DateTime},
				{types.Time},
			},
			model: func() model.IExpression {
				return &model.Successor{
					UnaryExpression: &model.UnaryExpression{},
				}
			},
		},
		// STRING OPERATORS - https://cql.hl7.org/09-b-cqlreference.html#string-operators-3
		// It is not clear whether Add(String, String) is a supported system operator overload.
		{
			name: "Add",
			operands: [][]types.IType{
				{types.String, types.String},
			},
			model: func() model.IExpression {
				return &model.Concatenate{
					NaryExpression: &model.NaryExpression{
						Expression: model.ResultType(types.String),
					},
				}
			},
		},
		{
			name: "Concatenate",
			operands: [][]types.IType{
				{types.String, types.String},
			},
			model: func() model.IExpression {
				return &model.Concatenate{
					NaryExpression: &model.NaryExpression{
						Expression: model.ResultType(types.String),
					},
				}
			},
		},
		{
			name: "EndsWith",
			operands: [][]types.IType{
				{types.String, types.String},
			},
			model: func() model.IExpression {
				return &model.EndsWith{
					BinaryExpression: &model.BinaryExpression{
						Expression: model.ResultType(types.Boolean),
					},
				}
			},
		},
		{
			name: "LastPositionOf",
			operands: [][]types.IType{
				{types.String, types.String},
			},
			model: func() model.IExpression {
				return &model.LastPositionOf{
					BinaryExpression: &model.BinaryExpression{
						Expression: model.ResultType(types.Integer),
					},
				}
			},
		},
		{
			name: "Upper",
			operands: [][]types.IType{
				{types.String},
			},
			model: func() model.IExpression {
				return &model.Upper{
					UnaryExpression: &model.UnaryExpression{
						Expression: model.ResultType(types.String),
					},
				}
			},
		},
		{
			name: "Lower",
			operands: [][]types.IType{
				{types.String},
			},
			model: func() model.IExpression {
				return &model.Lower{
					UnaryExpression: &model.UnaryExpression{
						Expression: model.ResultType(types.String),
					},
				}
			},
		},
		{
			name: "Split",
			operands: [][]types.IType{
				{types.String, types.String},
			},
			model: func() model.IExpression {
				return &model.Split{
					BinaryExpression: &model.BinaryExpression{
						Expression: model.ResultType(&types.List{ElementType: types.String}),
					},
				}
			},
		},
		{
			name: "Combine",
			operands: [][]types.IType{
				{&types.List{ElementType: types.String}},
				{&types.List{ElementType: types.String}, types.String},
			},
			model: func() model.IExpression {
				return &model.Combine{
					NaryExpression: &model.NaryExpression{
						Expression: model.ResultType(types.String),
					},
				}
			},
		},
		{
			name: "Indexer",
			operands: [][]types.IType{
				{types.String, types.Integer},
				{&types.List{ElementType: types.Any}, types.Integer},
			},
			model: func() model.IExpression {
				return &model.Indexer{
					// The result type is set in the resolveFunction().
					BinaryExpression: &model.BinaryExpression{},
				}
			},
		},
		// DATE AND TIME OPERATORS - https://cql.hl7.org/09-b-cqlreference.html#datetime-operators-2
		{
			name:     "Add",
			operands: [][]types.IType{{types.Date, types.Quantity}},
			model:    addModel(types.Date),
		},
		{
			name:     "Add",
			operands: [][]types.IType{{types.DateTime, types.Quantity}},
			model:    addModel(types.DateTime),
		},
		{
			name:     "Add",
			operands: [][]types.IType{{types.Time, types.Quantity}},
			model:    addModel(types.Time),
		},
		{
			name: "After",
			// See generatePrecisionTimingOverloads() for more overloads.
			operands: [][]types.IType{
				[]types.IType{types.Date, types.Date},
				[]types.IType{types.DateTime, types.DateTime},
				[]types.IType{types.Time, types.Time},
			},
			model: func() model.IExpression {
				return &model.After{
					BinaryExpression: &model.BinaryExpression{
						Expression: model.ResultType(types.Boolean),
					},
				}
			},
		},
		{
			name: "Before",
			// See generatePrecisionTimingOverloads() for more overloads.
			operands: [][]types.IType{
				[]types.IType{types.Date, types.Date},
				[]types.IType{types.DateTime, types.DateTime},
				[]types.IType{types.Time, types.Time},
			},
			model: func() model.IExpression {
				return &model.Before{
					BinaryExpression: &model.BinaryExpression{
						Expression: model.ResultType(types.Boolean),
					},
				}
			},
		},
		{
			name: "Date",
			operands: [][]types.IType{
				[]types.IType{types.Integer},
				[]types.IType{types.Integer, types.Integer},
				[]types.IType{types.Integer, types.Integer, types.Integer},
			},
			model: func() model.IExpression {
				return &model.Date{
					NaryExpression: &model.NaryExpression{
						Expression: model.ResultType(types.Date),
					},
				}
			},
		},
		{
			name: "DateTime",
			operands: [][]types.IType{
				[]types.IType{types.Integer},
				[]types.IType{types.Integer, types.Integer},
				[]types.IType{types.Integer, types.Integer, types.Integer},
				[]types.IType{types.Integer, types.Integer, types.Integer, types.Integer},
				[]types.IType{types.Integer, types.Integer, types.Integer, types.Integer, types.Integer},
				[]types.IType{types.Integer, types.Integer, types.Integer, types.Integer, types.Integer, types.Integer},
				[]types.IType{types.Integer, types.Integer, types.Integer, types.Integer, types.Integer, types.Integer, types.Integer},
				[]types.IType{types.Integer, types.Integer, types.Integer, types.Integer, types.Integer, types.Integer, types.Integer, types.Decimal},
			},
			model: func() model.IExpression {
				return &model.DateTime{
					NaryExpression: &model.NaryExpression{
						Expression: model.ResultType(types.DateTime),
					},
				}
			},
		},
		{
			name:     "Now",
			operands: [][]types.IType{{}},
			model: func() model.IExpression {
				return &model.Now{
					NaryExpression: &model.NaryExpression{
						Expression: model.ResultType(types.DateTime),
					},
				}
			},
		},
		{
			name:     "now",
			operands: [][]types.IType{{}},
			model: func() model.IExpression {
				return &model.Now{
					NaryExpression: &model.NaryExpression{
						Expression: model.ResultType(types.DateTime),
					},
				}
			},
		},
		{
			name:     "TimeOfDay",
			operands: [][]types.IType{{}},
			model: func() model.IExpression {
				return &model.TimeOfDay{
					NaryExpression: &model.NaryExpression{
						Expression: model.ResultType(types.Time),
					},
				}
			},
		},
		{
			name: "SameOrAfter",
			// See generatePrecisionTimingOverloads() for more overloads.
			operands: [][]types.IType{
				[]types.IType{types.Date, types.Date},
				[]types.IType{types.DateTime, types.DateTime},
				[]types.IType{types.Time, types.Time},
			},
			model: func() model.IExpression {
				return &model.SameOrAfter{
					BinaryExpression: &model.BinaryExpression{
						Expression: model.ResultType(types.Boolean),
					},
				}
			},
		},
		{
			name: "SameOrBefore",
			// See generatePrecisionTimingOverloads() for more overloads.
			operands: [][]types.IType{
				[]types.IType{types.Date, types.Date},
				[]types.IType{types.DateTime, types.DateTime},
				[]types.IType{types.Time, types.Time},
			},
			model: func() model.IExpression {
				return &model.SameOrBefore{
					BinaryExpression: &model.BinaryExpression{
						Expression: model.ResultType(types.Boolean),
					},
				}
			},
		},
		{
			name:     "Subtract",
			operands: [][]types.IType{{types.Date, types.Quantity}},
			model:    subtractModel(types.Date),
		},
		{
			name:     "Subtract",
			operands: [][]types.IType{{types.DateTime, types.Quantity}},
			model:    subtractModel(types.DateTime),
		},
		{
			name:     "Subtract",
			operands: [][]types.IType{{types.Time, types.Quantity}},
			model:    subtractModel(types.Time),
		},
		{
			name: "Time",
			operands: [][]types.IType{
				[]types.IType{types.Integer},
				[]types.IType{types.Integer, types.Integer},
				[]types.IType{types.Integer, types.Integer, types.Integer},
				[]types.IType{types.Integer, types.Integer, types.Integer, types.Integer},
			},
			model: func() model.IExpression {
				return &model.Time{
					NaryExpression: &model.NaryExpression{
						Expression: model.ResultType(types.Time),
					},
				}
			},
		},
		{
			name:     "Today",
			operands: [][]types.IType{{}},
			model: func() model.IExpression {
				return &model.Today{
					NaryExpression: &model.NaryExpression{
						Expression: model.ResultType(types.Date),
					},
				}
			},
		},
		// INTERVAL OPERATORS - https://cql.hl7.org/09-b-cqlreference.html#interval-operators-3
		{
			name: "After",
			// See generatePrecisionTimingOverloads() for more overloads.
			operands: comparableIntervalOverloads,
			model: func() model.IExpression {
				return &model.After{
					BinaryExpression: &model.BinaryExpression{
						Expression: model.ResultType(types.Boolean),
					},
				}
			},
		},
		{
			name:     "Before",
			operands: comparableIntervalOverloads,
			model: func() model.IExpression {
				return &model.Before{
					BinaryExpression: &model.BinaryExpression{
						Expression: model.ResultType(types.Boolean),
					},
				}
			},
		},
		{
			name: "Contains",
			// Contains is a macro for the In operator but with the operands reversed.
			// We convert to that model in resolveFunctions() above.
			operands: [][]types.IType{
				{convert.GenericList, convert.GenericType},
				// TODO(b/301606416): Add support for ContainsYears, ContainsDays...
				[]types.IType{&types.Interval{PointType: types.Integer}, types.Integer},
				[]types.IType{&types.Interval{PointType: types.Long}, types.Long},
				[]types.IType{&types.Interval{PointType: types.Decimal}, types.Decimal},
				[]types.IType{&types.Interval{PointType: types.Quantity}, types.Quantity},
				[]types.IType{&types.Interval{PointType: types.String}, types.String},
				[]types.IType{&types.Interval{PointType: types.Date}, types.Date},
				[]types.IType{&types.Interval{PointType: types.DateTime}, types.DateTime},
				[]types.IType{&types.Interval{PointType: types.Time}, types.Time},
			},
			model: func() model.IExpression {
				return &model.Contains{
					BinaryExpression: &model.BinaryExpression{
						Expression: model.ResultType(types.Boolean),
					},
				}
			},
		},
		{
			name: "End",
			operands: [][]types.IType{
				{&types.Interval{PointType: types.Any}},
			},
			model: func() model.IExpression {
				return &model.End{
					UnaryExpression: &model.UnaryExpression{},
				}
			},
		},
		{
			name: "In",
			operands: [][]types.IType{
				{convert.GenericType, convert.GenericList},
				[]types.IType{types.Integer, &types.Interval{PointType: types.Integer}},
				[]types.IType{types.Long, &types.Interval{PointType: types.Long}},
				[]types.IType{types.Decimal, &types.Interval{PointType: types.Decimal}},
				[]types.IType{types.Quantity, &types.Interval{PointType: types.Quantity}},
				[]types.IType{types.String, &types.Interval{PointType: types.String}},
				[]types.IType{types.Date, &types.Interval{PointType: types.Date}},
				[]types.IType{types.DateTime, &types.Interval{PointType: types.DateTime}},
				[]types.IType{types.Time, &types.Interval{PointType: types.Time}},
			},
			model: func() model.IExpression {
				return &model.In{
					BinaryExpression: &model.BinaryExpression{
						Expression: model.ResultType(types.Boolean),
					},
				}
			},
		},
		{
			name: "InYears",
			operands: [][]types.IType{
				[]types.IType{types.Date, &types.Interval{PointType: types.Date}},
				[]types.IType{types.DateTime, &types.Interval{PointType: types.DateTime}},
			},
			model: inModel(model.YEAR),
		},
		{
			name: "InMonths",
			operands: [][]types.IType{
				[]types.IType{types.Date, &types.Interval{PointType: types.Date}},
				[]types.IType{types.DateTime, &types.Interval{PointType: types.DateTime}},
			},
			model: inModel(model.MONTH),
		},
		{
			name: "InDays",
			operands: [][]types.IType{
				[]types.IType{types.Date, &types.Interval{PointType: types.Date}},
				[]types.IType{types.DateTime, &types.Interval{PointType: types.DateTime}},
			},
			model: inModel(model.DAY),
		},
		{
			name: "InHours",
			operands: [][]types.IType{
				[]types.IType{types.DateTime, &types.Interval{PointType: types.DateTime}},
				[]types.IType{types.Time, &types.Interval{PointType: types.Time}},
			},
			model: inModel(model.HOUR),
		},
		{
			name: "InMinutes",
			operands: [][]types.IType{
				[]types.IType{types.DateTime, &types.Interval{PointType: types.DateTime}},
				[]types.IType{types.Time, &types.Interval{PointType: types.Time}},
			},
			model: inModel(model.MINUTE),
		},
		{
			name: "InSeconds",
			operands: [][]types.IType{
				[]types.IType{types.DateTime, &types.Interval{PointType: types.DateTime}},
				[]types.IType{types.Time, &types.Interval{PointType: types.Time}},
			},
			model: inModel(model.SECOND),
		},
		{
			name: "InMilliseconds",
			operands: [][]types.IType{
				[]types.IType{types.DateTime, &types.Interval{PointType: types.DateTime}},
				[]types.IType{types.Time, &types.Interval{PointType: types.Time}},
			},
			model: inModel(model.MILLISECOND),
		},
		{
			name: "IncludedIn",
			operands: [][]types.IType{
				// op (left Interval<T>, right Interval<T>) Boolean
				[]types.IType{&types.Interval{PointType: types.Integer}, &types.Interval{PointType: types.Integer}},
				[]types.IType{&types.Interval{PointType: types.Long}, &types.Interval{PointType: types.Long}},
				[]types.IType{&types.Interval{PointType: types.Decimal}, &types.Interval{PointType: types.Decimal}},
				[]types.IType{&types.Interval{PointType: types.Quantity}, &types.Interval{PointType: types.Quantity}},
				[]types.IType{&types.Interval{PointType: types.String}, &types.Interval{PointType: types.String}},
				[]types.IType{&types.Interval{PointType: types.Date}, &types.Interval{PointType: types.Date}},
				[]types.IType{&types.Interval{PointType: types.DateTime}, &types.Interval{PointType: types.DateTime}},
				[]types.IType{&types.Interval{PointType: types.Time}, &types.Interval{PointType: types.Time}},
			},
			model: func() model.IExpression {
				return &model.IncludedIn{
					BinaryExpression: &model.BinaryExpression{
						Expression: model.ResultType(types.Boolean),
					},
				}
			},
		},
		// Included in for point type overloads is a macro for the In operator.
		{
			name: "IncludedIn",
			operands: [][]types.IType{
				// op (left T, right Interval<T>) Boolean
				[]types.IType{types.Integer, &types.Interval{PointType: types.Integer}},
				[]types.IType{types.Long, &types.Interval{PointType: types.Long}},
				[]types.IType{types.Decimal, &types.Interval{PointType: types.Decimal}},
				[]types.IType{types.Quantity, &types.Interval{PointType: types.Quantity}},
				[]types.IType{types.String, &types.Interval{PointType: types.String}},
				[]types.IType{types.Date, &types.Interval{PointType: types.Date}},
				[]types.IType{types.DateTime, &types.Interval{PointType: types.DateTime}},
				[]types.IType{types.Time, &types.Interval{PointType: types.Time}},
			},
			model: func() model.IExpression {
				return &model.In{
					BinaryExpression: &model.BinaryExpression{
						Expression: model.ResultType(types.Boolean),
					},
				}
			},
		},
		{
			name: "IncludedInYears",
			operands: [][]types.IType{
				[]types.IType{&types.Interval{PointType: types.Date}, &types.Interval{PointType: types.Date}},
				[]types.IType{&types.Interval{PointType: types.DateTime}, &types.Interval{PointType: types.DateTime}},
			},
			model: includedInModel(model.YEAR),
		},
		{
			name: "IncludedInYears",
			operands: [][]types.IType{
				[]types.IType{types.Date, &types.Interval{PointType: types.Date}},
				[]types.IType{types.DateTime, &types.Interval{PointType: types.DateTime}},
			},
			model: inModel(model.YEAR),
		},
		{
			name: "IncludedInMonths",
			operands: [][]types.IType{
				[]types.IType{&types.Interval{PointType: types.Date}, &types.Interval{PointType: types.Date}},
				[]types.IType{&types.Interval{PointType: types.DateTime}, &types.Interval{PointType: types.DateTime}},
			},
			model: includedInModel(model.MONTH),
		},
		{
			name: "IncludedInMonths",
			operands: [][]types.IType{
				[]types.IType{types.Date, &types.Interval{PointType: types.Date}},
				[]types.IType{types.DateTime, &types.Interval{PointType: types.DateTime}},
			},
			model: inModel(model.MONTH),
		},
		{
			name: "IncludedInDays",
			operands: [][]types.IType{
				[]types.IType{&types.Interval{PointType: types.Date}, &types.Interval{PointType: types.Date}},
				[]types.IType{&types.Interval{PointType: types.DateTime}, &types.Interval{PointType: types.DateTime}},
			},
			model: includedInModel(model.DAY),
		},
		{
			name: "IncludedInDays",
			operands: [][]types.IType{
				[]types.IType{types.Date, &types.Interval{PointType: types.Date}},
				[]types.IType{types.DateTime, &types.Interval{PointType: types.DateTime}},
			},
			model: inModel(model.DAY),
		},
		{
			name: "IncludedInHours",
			operands: [][]types.IType{
				[]types.IType{&types.Interval{PointType: types.DateTime}, &types.Interval{PointType: types.DateTime}},
				[]types.IType{&types.Interval{PointType: types.Time}, &types.Interval{PointType: types.Time}},
			},
			model: includedInModel(model.HOUR),
		},
		{
			name: "IncludedInHours",
			operands: [][]types.IType{
				[]types.IType{types.DateTime, &types.Interval{PointType: types.DateTime}},
				[]types.IType{types.Time, &types.Interval{PointType: types.Time}},
			},
			model: inModel(model.HOUR),
		},
		{
			name: "IncludedInMinutes",
			operands: [][]types.IType{
				[]types.IType{&types.Interval{PointType: types.DateTime}, &types.Interval{PointType: types.DateTime}},
				[]types.IType{&types.Interval{PointType: types.Time}, &types.Interval{PointType: types.Time}},
			},
			model: includedInModel(model.MINUTE),
		},
		{
			name: "IncludedInMinutes",
			operands: [][]types.IType{
				[]types.IType{types.DateTime, &types.Interval{PointType: types.DateTime}},
				[]types.IType{types.Time, &types.Interval{PointType: types.Time}},
			},
			model: inModel(model.MINUTE),
		},
		{
			name: "IncludedInSeconds",
			operands: [][]types.IType{
				[]types.IType{&types.Interval{PointType: types.DateTime}, &types.Interval{PointType: types.DateTime}},
				[]types.IType{&types.Interval{PointType: types.Time}, &types.Interval{PointType: types.Time}},
			},
			model: includedInModel(model.SECOND),
		},
		{
			name: "IncludedInSeconds",
			operands: [][]types.IType{
				[]types.IType{types.DateTime, &types.Interval{PointType: types.DateTime}},
				[]types.IType{types.Time, &types.Interval{PointType: types.Time}},
			},
			model: inModel(model.SECOND),
		},
		{
			name: "IncludedInMilliseconds",
			operands: [][]types.IType{
				[]types.IType{&types.Interval{PointType: types.DateTime}, &types.Interval{PointType: types.DateTime}},
				[]types.IType{&types.Interval{PointType: types.Time}, &types.Interval{PointType: types.Time}},
			},
			model: includedInModel(model.MILLISECOND),
		},
		{
			name: "IncludedInMilliseconds",
			operands: [][]types.IType{
				[]types.IType{types.DateTime, &types.Interval{PointType: types.DateTime}},
				[]types.IType{types.Time, &types.Interval{PointType: types.Time}},
			},
			model: inModel(model.MILLISECOND),
		},
		{
			name: "Overlaps",
			operands: [][]types.IType{
				[]types.IType{&types.Interval{PointType: types.Date}, &types.Interval{PointType: types.Date}},
				[]types.IType{&types.Interval{PointType: types.DateTime}, &types.Interval{PointType: types.DateTime}},
			},
			model: func() model.IExpression {
				return &model.Overlaps{
					BinaryExpression: &model.BinaryExpression{
						Expression: model.ResultType(types.Boolean),
					},
				}
			},
		},
		{
			name: "SameOrAfter",
			// See generatePrecisionTimingOverloads() for more overloads.
			operands: comparableIntervalOverloads,
			model: func() model.IExpression {
				return &model.SameOrAfter{
					BinaryExpression: &model.BinaryExpression{
						Expression: model.ResultType(types.Boolean),
					},
				}
			},
		},
		{
			name: "SameOrBefore",
			// See generatePrecisionTimingOverloads() for more overloads.
			operands: comparableIntervalOverloads,
			model: func() model.IExpression {
				return &model.SameOrBefore{
					BinaryExpression: &model.BinaryExpression{
						Expression: model.ResultType(types.Boolean),
					},
				}
			},
		},
		{
			name: "Start",
			operands: [][]types.IType{
				{&types.Interval{PointType: types.Any}},
			},
			model: func() model.IExpression {
				return &model.Start{
					UnaryExpression: &model.UnaryExpression{},
				}
			},
		},
		// LIST OPERATORS - https://cql.hl7.org/09-b-cqlreference.html#list-operators-2
		{
			name:     "Except",
			operands: [][]types.IType{{&types.List{ElementType: types.Any}, &types.List{ElementType: types.Any}}},
			model: func() model.IExpression {
				return &model.Except{
					BinaryExpression: &model.BinaryExpression{},
				}
			},
		},
		{
			name: "Exists",
			operands: [][]types.IType{
				{&types.List{ElementType: types.Any}}},
			model: func() model.IExpression {
				return &model.Exists{
					UnaryExpression: &model.UnaryExpression{
						Expression: model.ResultType(types.Boolean),
					},
				}
			},
		},
		{
			name: "First",
			operands: [][]types.IType{
				{&types.List{ElementType: types.Any}}},
			model: func() model.IExpression {
				return &model.First{
					UnaryExpression: &model.UnaryExpression{},
				}
			},
		},
		{
			name: "Flatten",
			operands: [][]types.IType{
				{&types.List{ElementType: types.Any}}},
			model: func() model.IExpression {
				return &model.Flatten{
					UnaryExpression: &model.UnaryExpression{},
				}
			},
		},
		{
			name: "Distinct",
			operands: [][]types.IType{
				{&types.List{ElementType: types.Any}}},
			model: func() model.IExpression {
				return &model.Distinct{
					UnaryExpression: &model.UnaryExpression{},
				}
			},
		},
		{
			name: "Includes",
			operands: [][]types.IType{
				{convert.GenericList, convert.GenericType},
				{convert.GenericList, convert.GenericList},
			},
			model: func() model.IExpression {
				return &model.Includes{
					BinaryExpression: &model.BinaryExpression{
						Expression: model.ResultType(types.Boolean),
					},
				}
			},
		},
		{
			name:     "Intersect",
			operands: [][]types.IType{{&types.List{ElementType: types.Any}, &types.List{ElementType: types.Any}}},
			model: func() model.IExpression {
				return &model.Intersect{
					BinaryExpression: &model.BinaryExpression{},
				}
			},
		},
		{
			name: "IndexOf",
			operands: [][]types.IType{
				{&types.List{ElementType: types.Any}, types.Any},
			},
			model: func() model.IExpression {
				return &model.IndexOf{
					BinaryExpression: &model.BinaryExpression{
						Expression: model.ResultType(types.Integer),
					},
				}
			},
		},
		{
			name: "ProperlyIncludes",
			operands: [][]types.IType{
				{convert.GenericList, convert.GenericType},
				{convert.GenericList, convert.GenericList},
			},
			model: func() model.IExpression {
				return &model.ProperlyIncludes{
					BinaryExpression: &model.BinaryExpression{
						Expression: model.ResultType(types.Boolean),
					},
				}
			},
		},
		{
			name: "Last",
			operands: [][]types.IType{
				{&types.List{ElementType: types.Any}}},
			model: func() model.IExpression {
				return &model.Last{
					UnaryExpression: &model.UnaryExpression{},
				}
			},
		},
		{
			name: "Length",
			operands: [][]types.IType{
				{types.String},
				{&types.List{ElementType: types.Any}}},
			model: func() model.IExpression {
				return &model.Length{
					UnaryExpression: &model.UnaryExpression{
						Expression: model.ResultType(types.Integer),
					},
				}
			},
		},
		{
			name: "SingletonFrom",
			operands: [][]types.IType{
				{&types.List{ElementType: types.Any}}},
			model: func() model.IExpression {
				return &model.SingletonFrom{
					UnaryExpression: &model.UnaryExpression{},
				}
			},
		},
		{
			name: "Skip",
			operands: [][]types.IType{
				{&types.List{ElementType: types.Any}, types.Integer},
			},
			model: func() model.IExpression {
				return &model.Skip{
					BinaryExpression: &model.BinaryExpression{},
				}
			},
		},
		{
			name: "Tail",
			operands: [][]types.IType{
				{&types.List{ElementType: types.Any}}},
			model: func() model.IExpression {
				return &model.Tail{
					UnaryExpression: &model.UnaryExpression{},
				}
			},
		},
		{
			name: "Take",
			operands: [][]types.IType{
				{&types.List{ElementType: types.Any}, types.Integer},
			},
			model: func() model.IExpression {
				return &model.Take{
					BinaryExpression: &model.BinaryExpression{},
				}
			},
		},
		{
			name:     "Union",
			operands: [][]types.IType{{&types.List{ElementType: types.Any}, &types.List{ElementType: types.Any}}},
			model: func() model.IExpression {
				return &model.Union{
					BinaryExpression: &model.BinaryExpression{},
				}
			},
		},
		// AGGREGATE FUNCTIONS - https://cql.hl7.org/09-b-cqlreference.html#aggregate-functions
		{
			name: "AllTrue",
			operands: [][]types.IType{
				{&types.List{ElementType: types.Boolean}},
			},
			model: func() model.IExpression {
				return &model.AllTrue{
					UnaryExpression: &model.UnaryExpression{
						Expression: model.ResultType(types.Boolean),
					},
				}
			},
		},
		{
			name: "AnyTrue",
			operands: [][]types.IType{
				{&types.List{ElementType: types.Boolean}},
			},
			model: func() model.IExpression {
				return &model.AnyTrue{
					UnaryExpression: &model.UnaryExpression{
						Expression: model.ResultType(types.Boolean),
					},
				}
			},
		},
		{
			name: "Avg",
			operands: [][]types.IType{
				{&types.List{ElementType: types.Decimal}},
				{&types.List{ElementType: types.Quantity}},
			},
			model: func() model.IExpression {
				return &model.Avg{
					UnaryExpression: &model.UnaryExpression{},
				}
			},
		},
		{
			name: "Count",
			operands: [][]types.IType{
				{&types.List{ElementType: types.Any}},
			},
			model: func() model.IExpression {
				return &model.Count{
					UnaryExpression: &model.UnaryExpression{
						Expression: model.ResultType(types.Integer),
					},
				}
			},
		},
		{
			name: "Max",
			operands: [][]types.IType{
				{&types.List{ElementType: types.Date}},
				{&types.List{ElementType: types.DateTime}},
			},
			model: func() model.IExpression {
				return &model.Max{
					UnaryExpression: &model.UnaryExpression{},
				}
			},
		},
		{
			name: "Min",
			operands: [][]types.IType{
				{&types.List{ElementType: types.Date}},
				{&types.List{ElementType: types.DateTime}},
			},
			model: func() model.IExpression {
				return &model.Min{
					UnaryExpression: &model.UnaryExpression{},
				}
			},
		},
		{
			name: "Sum",
			operands: [][]types.IType{
				{&types.List{ElementType: types.Decimal}},
				{&types.List{ElementType: types.Integer}},
				{&types.List{ElementType: types.Long}},
				{&types.List{ElementType: types.Quantity}},
			},
			model: func() model.IExpression {
				return &model.Sum{
					UnaryExpression: &model.UnaryExpression{},
				}
			},
		},
		// CLINICAL OPERATORS - https://cql.hl7.org/09-b-cqlreference.html#clinical-operators-3
		{
			name:     "AgeInYears",
			operands: [][]types.IType{[]types.IType{}},
			model:    calculateAgeModel(model.YEAR),
		},
		{
			name:     "AgeInMonths",
			operands: [][]types.IType{[]types.IType{}},
			model:    calculateAgeModel(model.MONTH),
		},
		{
			name:     "AgeInWeeks",
			operands: [][]types.IType{[]types.IType{}},
			model:    calculateAgeModel(model.WEEK),
		},
		{
			name:     "AgeInDays",
			operands: [][]types.IType{[]types.IType{}},
			model:    calculateAgeModel(model.DAY),
		},
		{
			name:     "AgeInHours",
			operands: [][]types.IType{[]types.IType{}},
			model:    calculateAgeModel(model.HOUR),
		},
		{
			name:     "AgeInMinutes",
			operands: [][]types.IType{[]types.IType{}},
			model:    calculateAgeModel(model.MINUTE),
		},
		{
			name:     "AgeInSeconds",
			operands: [][]types.IType{[]types.IType{}},
			model:    calculateAgeModel(model.SECOND),
		},
		{
			name: "AgeInYearsAt",
			operands: [][]types.IType{
				{types.Date},
				{types.DateTime},
			},
			model: calculateAgeAtModel(model.YEAR),
		},
		{
			name: "AgeInMonthsAt",
			operands: [][]types.IType{
				{types.Date},
				{types.DateTime},
			},
			model: calculateAgeAtModel(model.MONTH),
		},
		{
			name: "AgeInWeeksAt",
			operands: [][]types.IType{
				{types.Date},
				{types.DateTime},
			},
			model: calculateAgeAtModel(model.WEEK),
		},
		{
			name: "AgeInDaysAt",
			operands: [][]types.IType{
				{types.Date},
				{types.DateTime},
			},
			model: calculateAgeAtModel(model.DAY),
		},
		{
			name: "AgeInHoursAt",
			operands: [][]types.IType{
				{types.Date},
				{types.DateTime},
			},
			model: calculateAgeAtModel(model.HOUR),
		},
		{
			name: "AgeInMinutesAt",
			operands: [][]types.IType{
				{types.Date},
				{types.DateTime},
			},
			model: calculateAgeAtModel(model.MINUTE),
		},
		{
			name: "AgeInSecondsAt",
			operands: [][]types.IType{
				{types.Date},
				{types.DateTime},
			},
			model: calculateAgeAtModel(model.SECOND),
		},
		{
			name: "CalculateAgeInYears",
			operands: [][]types.IType{
				{types.Date},
				{types.DateTime},
			},
			model: calculateAgeModel(model.YEAR),
		},
		{
			name: "CalculateAgeInMonths",
			operands: [][]types.IType{
				{types.Date},
				{types.DateTime},
			},
			model: calculateAgeModel(model.MONTH),
		},
		{
			name: "CalculateAgeInWeeks",
			operands: [][]types.IType{
				{types.Date},
				{types.DateTime},
			},
			model: calculateAgeModel(model.WEEK),
		},
		{
			name: "CalculateAgeInDays",
			operands: [][]types.IType{
				{types.Date},
				{types.DateTime},
			},
			model: calculateAgeModel(model.DAY),
		},
		{
			name: "CalculateAgeInHours",
			operands: [][]types.IType{
				{types.Date},
				{types.DateTime}},
			model: calculateAgeModel(model.HOUR),
		},
		{
			name: "CalculateAgeInMinutes",
			operands: [][]types.IType{
				{types.Date},
				{types.DateTime},
			},
			model: calculateAgeModel(model.MINUTE),
		},
		{
			name: "CalculateAgeInSeconds",
			operands: [][]types.IType{
				{types.Date},
				{types.DateTime},
			},
			model: calculateAgeModel(model.SECOND),
		},
		{
			name: "CalculateAgeInYearsAt",
			operands: [][]types.IType{
				{types.Date, types.Date},
				{types.DateTime, types.DateTime},
			},
			model: calculateAgeAtModel(model.YEAR),
		},
		{
			name: "CalculateAgeInMonthsAt",
			operands: [][]types.IType{
				{types.Date, types.Date},
				{types.DateTime, types.DateTime},
			},
			model: calculateAgeAtModel(model.MONTH),
		},
		{
			name: "CalculateAgeInWeeksAt",
			operands: [][]types.IType{
				{types.Date, types.Date},
				{types.DateTime, types.DateTime},
			},
			model: calculateAgeAtModel(model.WEEK),
		},
		{
			name: "CalculateAgeInDaysAt",
			operands: [][]types.IType{
				{types.Date, types.Date},
				{types.DateTime, types.DateTime},
			},
			model: calculateAgeAtModel(model.DAY),
		},
		{
			name: "CalculateAgeInHoursAt",
			operands: [][]types.IType{
				{types.DateTime, types.DateTime},
			},
			model: calculateAgeAtModel(model.HOUR),
		},
		{
			name: "CalculateAgeInMinutesAt",
			operands: [][]types.IType{
				{types.DateTime, types.DateTime},
			},
			model: calculateAgeAtModel(model.MINUTE),
		},
		{
			name: "CalculateAgeInSecondsAt",
			operands: [][]types.IType{
				{types.DateTime, types.DateTime},
			},
			model: calculateAgeAtModel(model.SECOND),
		},
		{
			name: "InCodeSystem",
			operands: [][]types.IType{
				{types.Code, types.CodeSystem},
				{&types.List{ElementType: types.Code}, types.CodeSystem},
				{types.Concept, types.CodeSystem},
				{&types.List{ElementType: types.Concept}, types.CodeSystem},
			},
			model: func() model.IExpression {
				return &model.InCodeSystem{
					BinaryExpression: &model.BinaryExpression{
						Expression: model.ResultType(types.Boolean),
					},
				}
			},
		},
		{
			name: "InValueSet",
			operands: [][]types.IType{
				{types.Code, types.ValueSet},
				{&types.List{ElementType: types.Code}, types.ValueSet},
				{types.Concept, types.ValueSet},
				{&types.List{ElementType: types.Concept}, types.ValueSet},
			},
			model: func() model.IExpression {
				return &model.InValueSet{
					BinaryExpression: &model.BinaryExpression{
						Expression: model.ResultType(types.Boolean),
					},
				}
			},
		},
		{
			// ERRORS AND MESSAGING - https://cql.hl7.org/09-b-cqlreference.html#errors-and-messaging
			// The ELM for `Message` is states that all arguments besides the Source are optional.
			// However, the Docs and the Java engine seem to expect message to only support the 5
			// argument overload. For now we are only supporting the 5 argument variant but we can
			// re-address this at a later date.
			name: "Message",
			operands: [][]types.IType{
				{types.Any, types.Boolean, types.String, types.String, types.String},
			},
			model: func() model.IExpression {
				return &model.Message{}
			},
		},
		{
			name: "Median",
			operands: [][]types.IType{
				{&types.List{ElementType: types.Decimal}},
				{&types.List{ElementType: types.Quantity}},
			},
			model: func() model.IExpression {
				return &model.Median{
					UnaryExpression: &model.UnaryExpression{},
				}
			},
		},
		{
			name: "PopulationStdDev",
			operands: [][]types.IType{
				{&types.List{ElementType: types.Decimal}},
				{&types.List{ElementType: types.Quantity}},
			},
			model: func() model.IExpression {
				return &model.PopulationStdDev{
					UnaryExpression: &model.UnaryExpression{},
				}
			},
		},
	}

	for _, b := range systemOperators {
		for _, operand := range b.operands {
			if err := p.refs.DefineBuiltinFunc(b.name, operand, b.model); err != nil {
				return err
			}
		}
	}

	if err := p.generatePrecisionTimingOverloads(); err != nil {
		return err
	}

	if err := p.generateDifferenceBetweenOverloads(); err != nil {
		return err
	}

	return nil
}

var comparableIntervalOverloads = [][]types.IType{
	// op (left Interval<T>, right Interval<T>) Boolean
	[]types.IType{&types.Interval{PointType: types.Integer}, &types.Interval{PointType: types.Integer}},
	[]types.IType{&types.Interval{PointType: types.Long}, &types.Interval{PointType: types.Long}},
	[]types.IType{&types.Interval{PointType: types.Decimal}, &types.Interval{PointType: types.Decimal}},
	[]types.IType{&types.Interval{PointType: types.Quantity}, &types.Interval{PointType: types.Quantity}},
	[]types.IType{&types.Interval{PointType: types.String}, &types.Interval{PointType: types.String}},
	[]types.IType{&types.Interval{PointType: types.Date}, &types.Interval{PointType: types.Date}},
	[]types.IType{&types.Interval{PointType: types.DateTime}, &types.Interval{PointType: types.DateTime}},
	[]types.IType{&types.Interval{PointType: types.Time}, &types.Interval{PointType: types.Time}},
	// op (left T, right Interval<T>) Boolean
	[]types.IType{types.Integer, &types.Interval{PointType: types.Integer}},
	[]types.IType{types.Long, &types.Interval{PointType: types.Long}},
	[]types.IType{types.Decimal, &types.Interval{PointType: types.Decimal}},
	[]types.IType{types.Quantity, &types.Interval{PointType: types.Quantity}},
	[]types.IType{types.String, &types.Interval{PointType: types.String}},
	[]types.IType{types.Date, &types.Interval{PointType: types.Date}},
	[]types.IType{types.DateTime, &types.Interval{PointType: types.DateTime}},
	[]types.IType{types.Time, &types.Interval{PointType: types.Time}},
	// op (left Interval<T>, right T) Boolean
	[]types.IType{&types.Interval{PointType: types.Integer}, types.Integer},
	[]types.IType{&types.Interval{PointType: types.Long}, types.Long},
	[]types.IType{&types.Interval{PointType: types.Decimal}, types.Decimal},
	[]types.IType{&types.Interval{PointType: types.Quantity}, types.Quantity},
	[]types.IType{&types.Interval{PointType: types.String}, types.String},
	[]types.IType{&types.Interval{PointType: types.Date}, types.Date},
	[]types.IType{&types.Interval{PointType: types.DateTime}, types.DateTime},
	[]types.IType{&types.Interval{PointType: types.Time}, types.Time},
}

func (p *Parser) generatePrecisionTimingOverloads() error {
	overloads := [][]types.IType{
		[]types.IType{types.Date, types.Date},
		[]types.IType{types.DateTime, types.DateTime},
		[]types.IType{types.Time, types.Time},
		// (left Interval<T>, right Interval<T>) Boolean
		[]types.IType{&types.Interval{PointType: types.Date}, &types.Interval{PointType: types.Date}},
		[]types.IType{&types.Interval{PointType: types.DateTime}, &types.Interval{PointType: types.DateTime}},
		[]types.IType{&types.Interval{PointType: types.Time}, &types.Interval{PointType: types.Time}},
		// (left T, right Interval<T>) Boolean
		[]types.IType{types.Date, &types.Interval{PointType: types.Date}},
		[]types.IType{types.DateTime, &types.Interval{PointType: types.DateTime}},
		[]types.IType{types.Time, &types.Interval{PointType: types.Time}},
		// (left Interval<T>, right T) Boolean
		[]types.IType{&types.Interval{PointType: types.Date}, types.Date},
		[]types.IType{&types.Interval{PointType: types.DateTime}, types.DateTime},
		[]types.IType{&types.Interval{PointType: types.Time}, types.Time},
	}

	for _, precision := range dateTimePrecisions() {
		name := funcNameWithPrecision("After", precision)
		for _, overload := range overloads {
			if err := p.refs.DefineBuiltinFunc(name, overload, afterModel(precision)); err != nil {
				return err
			}
		}
	}
	for _, precision := range dateTimePrecisions() {
		name := funcNameWithPrecision("Before", precision)
		for _, overload := range overloads {
			if err := p.refs.DefineBuiltinFunc(name, overload, beforeModel(precision)); err != nil {
				return err
			}
		}
	}
	for _, precision := range dateTimePrecisions() {
		name := funcNameWithPrecision("SameOrAfter", precision)
		for _, overload := range overloads {
			if err := p.refs.DefineBuiltinFunc(name, overload, sameOrAfterModel(precision)); err != nil {
				return err
			}
		}
	}
	for _, precision := range dateTimePrecisions() {
		name := funcNameWithPrecision("SameOrBefore", precision)
		for _, overload := range overloads {
			if err := p.refs.DefineBuiltinFunc(name, overload, sameOrBeforeModel(precision)); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *Parser) generateDifferenceBetweenOverloads() error {
	overloads := [][]types.IType{
		[]types.IType{types.Date, types.Date},
		[]types.IType{types.DateTime, types.DateTime},
		[]types.IType{types.Time, types.Time},
	}

	for _, precision := range dateTimePrecisions() {
		name := funcNameWithPrecision("DifferenceBetween", precision)
		for _, overload := range overloads {
			if err := p.refs.DefineBuiltinFunc(name, overload, differenceBetweenModel(precision)); err != nil {
				return err
			}
		}
	}
	return nil
}

func differenceBetweenModel(precision model.DateTimePrecision) func() model.IExpression {
	return func() model.IExpression {
		return &model.DifferenceBetween{
			BinaryExpression: &model.BinaryExpression{
				Expression: model.ResultType(types.Integer),
			},
			Precision: precision,
		}
	}
}

func afterModel(precision model.DateTimePrecision) func() model.IExpression {
	return func() model.IExpression {
		return &model.After{
			BinaryExpression: &model.BinaryExpression{
				Expression: model.ResultType(types.Boolean),
			},
			Precision: precision,
		}
	}
}

func beforeModel(precision model.DateTimePrecision) func() model.IExpression {
	return func() model.IExpression {
		return &model.Before{
			BinaryExpression: &model.BinaryExpression{
				Expression: model.ResultType(types.Boolean),
			},
			Precision: precision,
		}
	}
}

func sameOrAfterModel(precision model.DateTimePrecision) func() model.IExpression {
	return func() model.IExpression {
		return &model.SameOrAfter{
			BinaryExpression: &model.BinaryExpression{
				Expression: model.ResultType(types.Boolean),
			},
			Precision: precision,
		}
	}
}

func sameOrBeforeModel(precision model.DateTimePrecision) func() model.IExpression {
	return func() model.IExpression {
		return &model.SameOrBefore{
			BinaryExpression: &model.BinaryExpression{
				Expression: model.ResultType(types.Boolean),
			},
			Precision: precision,
		}
	}
}

// Returns an expression containing the patient's birth date property, as defined by the model info.
// For FHIR model info this should return a System Date.
func (v *visitor) patientBirthDateExpression() (model.IExpression, error) {
	pDate, err := v.modelInfo.PatientBirthDatePropertyName()
	if err != nil {
		return nil, err
	}
	// Ideally this should be handled in VisitInvocationExpressionTerm, but this property isn't really
	// CQL, and did not go through the ANTLR parser so we handle it here.
	// The goal is to turn birthDate.value into
	//   &model.Property{
	//      Source: &model.Property{Source: &model.ExpressionRef{"Patient"}, Path: "birthDate"},
	//      Path: "value",
	//   }
	propertyComponents := strings.Split(pDate, ".")

	var source model.IExpression
	// This references a statement created by `context Patient` expressions, which must exist
	// to access in-context patient properties like this.
	sourceFunc, err := v.refs.ResolveLocal("Patient")
	if err != nil {
		return nil, err
	}
	source = sourceFunc()
	for _, component := range propertyComponents {
		propertyType, err := v.modelInfo.PropertyTypeSpecifier(source.GetResultType(), component)
		if err != nil {
			return nil, err
		}
		source = &model.Property{
			Source:     source,
			Path:       component,
			Expression: model.ResultType(propertyType),
		}
	}
	return source, nil
}

func calculateAgeModel(precision model.DateTimePrecision) func() model.IExpression {
	return func() model.IExpression {
		return &model.CalculateAge{
			UnaryExpression: &model.UnaryExpression{
				Expression: model.ResultType(types.Integer),
			},
			Precision: precision,
		}
	}
}

func calculateAgeAtModel(precision model.DateTimePrecision) func() model.IExpression {
	return func() model.IExpression {
		return &model.CalculateAgeAt{
			BinaryExpression: &model.BinaryExpression{
				Expression: model.ResultType(types.Integer),
			},
			Precision: precision,
		}
	}
}

func inModel(precision model.DateTimePrecision) func() model.IExpression {
	return func() model.IExpression {
		return &model.In{
			BinaryExpression: &model.BinaryExpression{
				Expression: model.ResultType(types.Boolean),
			},
			Precision: precision,
		}
	}
}

func includedInModel(precision model.DateTimePrecision) func() model.IExpression {
	return func() model.IExpression {
		return &model.IncludedIn{
			BinaryExpression: &model.BinaryExpression{
				Expression: model.ResultType(types.Boolean),
			},
			Precision: precision,
		}
	}
}

func absModel(resultType types.System) func() model.IExpression {
	return func() model.IExpression {
		return &model.Abs{
			UnaryExpression: &model.UnaryExpression{
				Expression: model.ResultType(resultType),
			},
		}
	}
}

func addModel(resultType types.System) func() model.IExpression {
	return func() model.IExpression {
		return &model.Add{
			BinaryExpression: &model.BinaryExpression{
				Expression: model.ResultType(resultType),
			},
		}
	}
}

func negateModel(resultType types.System) func() model.IExpression {
	return func() model.IExpression {
		return &model.Negate{
			UnaryExpression: &model.UnaryExpression{
				Expression: model.ResultType(resultType),
			},
		}
	}
}

func subtractModel(resultType types.System) func() model.IExpression {
	return func() model.IExpression {
		return &model.Subtract{
			BinaryExpression: &model.BinaryExpression{
				Expression: model.ResultType(resultType),
			},
		}
	}
}

func multiplyModel(resultType types.System) func() model.IExpression {
	return func() model.IExpression {
		return &model.Multiply{
			BinaryExpression: &model.BinaryExpression{
				Expression: model.ResultType(resultType),
			},
		}
	}
}

func divideModel(resultType types.System) func() model.IExpression {
	return func() model.IExpression {
		return &model.Divide{
			BinaryExpression: &model.BinaryExpression{
				Expression: model.ResultType(resultType),
			},
		}
	}
}

func truncatedDivideModel(resultType types.System) func() model.IExpression {
	return func() model.IExpression {
		return &model.TruncatedDivide{
			BinaryExpression: &model.BinaryExpression{
				Expression: model.ResultType(resultType),
			},
		}
	}
}

func modModel(resultType types.System) func() model.IExpression {
	return func() model.IExpression {
		return &model.Modulo{
			BinaryExpression: &model.BinaryExpression{
				Expression: model.ResultType(resultType),
			},
		}
	}
}

func powerModel(resultType types.System) func() model.IExpression {
	return func() model.IExpression {
		return &model.Power{
			BinaryExpression: &model.BinaryExpression{
				Expression: model.ResultType(resultType),
			},
		}
	}
}

func precisionModel() func() model.IExpression {
	return func() model.IExpression {
		return &model.Precision{
			UnaryExpression: &model.UnaryExpression{
				Expression: model.ResultType(types.Integer),
			},
		}
	}
}

func (v *visitor) parseCoalesce(m *model.Coalesce, operands []model.IExpression) (model.IExpression, error) {
	if len(operands) == 1 {
		// This is the list overload.
		lType, ok := operands[0].GetResultType().(*types.List)
		if !ok {
			return nil, fmt.Errorf("internal error - Coalesce() overload with one operand should be of type list")
		}
		m.SetOperands(operands)
		m.Expression = model.ResultType(lType.ElementType)
		return m, nil
	}

	// This does not match the described behaviour of Coalesce in the spec, but based on community
	// discussions it is the correct behaviour.
	res, err := convert.InferMixed(operands, v.modelInfo)
	if err != nil {
		return nil, err
	}
	if res.PuntedToChoice {
		return nil, fmt.Errorf("all operands in coalesce(%v) must be implicitly convertible to the same type", convert.OperandsToString(operands))
	}
	m.SetOperands(res.WrappedOperands)
	m.Expression = model.ResultType(res.UniformType)
	return m, nil
}
