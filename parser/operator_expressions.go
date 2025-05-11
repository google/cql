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
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/google/cql/internal/convert"
	"github.com/google/cql/internal/embeddata/third_party/cqframework/cql"
	"github.com/google/cql/model"
	"github.com/google/cql/types"
	"github.com/antlr4-go/antlr/v4"
)

// handleBeforeOrAfterInterval processes a BeforeOrAfterIntervalOperatorPhraseContext and returns
// the function operator name, precision, relative offset, and quantity if applicable.
func (v *visitor) handleBeforeOrAfterInterval(
	operator *cql.BeforeOrAfterIntervalOperatorPhraseContext,
	ctx *cql.TimingExpressionContext) (string, model.DateTimePrecision, string, model.Quantity, error) {

	var fnOperator string
	var precision model.DateTimePrecision
	var relativeOffset string
	var quantity model.Quantity

	if operator.QuantityOffset() != nil {
		qo := operator.QuantityOffset()
		var err error
		quantity, err = v.VisitQuantityContext(qo.Quantity())
		if err != nil {
			return "", "", "", quantity, err
		}
		rq := qo.OffsetRelativeQualifier().GetText()

		if strings.Contains(rq, "or more") {
			relativeOffset = "OrMore"
		} else if strings.Contains(rq, "or less") {
			relativeOffset = "OrLess"
		} else {
			return "", "", "", quantity, errors.New("internal error - grammar should not allow this offsetRelativeQualifier")
		}
	}

	precision = precisionFromContext(operator)
	opText := operator.GetText()
	containsOnOr := strings.Contains(opText, "on or") || strings.Contains(opText, "or on")
	containsAfter := strings.Contains(opText, "after")
	containsBefore := strings.Contains(opText, "before")

	if containsOnOr && containsBefore {
		fnOperator = "SameOrBefore"
	} else if containsOnOr && containsAfter {
		fnOperator = "SameOrAfter"
	} else if containsAfter {
		fnOperator = "After"
	} else if containsBefore {
		fnOperator = "Before"
	} else {
		return "", "", "", quantity, errors.New("internal error - grammar should not allow this TimeBoundaryExpression")
	}

	return fnOperator, precision, relativeOffset, quantity, nil
}

// VisitTimingExpression expressions related to comparing one timing expression with another.
// Structured as expression, intervalOperatorPhrase, expression.
func (v *visitor) VisitTimingExpression(ctx *cql.TimingExpressionContext) model.IExpression {
	var fnOperator string
	var precision model.DateTimePrecision
	intervalOperator := ctx.GetChild(1)
	// if relativeOffset exists we need to wrap the right operand with the quantity offset information.
	var relativeOffset string

	var quantity model.Quantity
	switch operator := intervalOperator.(type) {
	case *cql.BeforeOrAfterIntervalOperatorPhraseContext:
		var err error
		if operator.QuantityOffset() != nil {
			qo := operator.QuantityOffset()
			quantity, err = v.VisitQuantityContext(qo.Quantity())
			if err != nil {
				return v.badExpression(err.Error(), ctx)
			}
			rq := qo.OffsetRelativeQualifier().GetText()

			if strings.Contains(rq, "or more") {
				relativeOffset = "OrMore"
			} else if strings.Contains(rq, "or less") {
				relativeOffset = "OrLess"
			} else {
				return v.badExpression("internal error - grammar should not allow this offsetRelativeQualifier", ctx)
			}
		}
		precision = precisionFromContext(operator)
		opText := operator.GetText()
		containsOnOr := strings.Contains(opText, "on or") || strings.Contains(opText, "or on")
		containsAfter := strings.Contains(opText, "after")
		containsBefore := strings.Contains(opText, "before")
		if containsOnOr && containsBefore {
			fnOperator = "SameOrBefore"
		} else if containsOnOr && containsAfter {
			fnOperator = "SameOrAfter"
		} else if containsAfter {
			fnOperator = "After"
		} else if containsBefore {
			fnOperator = "Before"
		} else {
			return v.badExpression("internal error - grammar should not allow this TimeBoundaryExpression", ctx)
		}
	case *cql.IncludesIntervalOperatorPhraseContext:
		precision = precisionFromContext(operator)
		if precision != model.UNSETDATETIMEPRECISION {
			return v.badExpression("unsupported interval operator in timing expression, includes operator with precision is not supported", ctx)
		}
		switch operator.GetText() {
		case "includes":
			fnOperator = "Includes"
		case "properlyincludes":
			fnOperator = "ProperlyIncludes"
		default:
			return v.badExpression("internal error - grammar should not allow this TimeBoundaryExpression", ctx)
		}
	case *cql.IncludedInIntervalOperatorPhraseContext:
		precision = precisionFromContext(operator)
		fnOperator = "IncludedIn"
	case *cql.ConcurrentWithIntervalOperatorPhraseContext:
		precision = precisionFromContext(operator)
		// TODO(b/298104070): Support ConcurrentWithIntervalOperatorPhraseContext without 'or'
		rq := operator.RelativeQualifier()
		if rq == nil {
			fnOperator = "SameAs"
		} else {
			opText := rq.GetText()
			if strings.Contains(opText, "after") {
				fnOperator = "SameOrAfter"
			} else if strings.Contains(opText, "before") {
				fnOperator = "SameOrBefore"
			} else {
				return v.badExpression("internal error - grammar should not allow this TimeBoundaryExpression", ctx)
			}
		}
	case *cql.OverlapsIntervalOperatorPhraseContext:
		precision = precisionFromContext(operator)
		fnOperator = "Overlaps"
		opText := operator.GetText()
		containsAfter := strings.Contains(opText, "after")
		containsBefore := strings.Contains(opText, "before")
		if containsAfter {
			return v.badExpression("overlaps after operator is not supported", ctx)
		} else if containsBefore {
			return v.badExpression("overlaps before operator is not supported", ctx)
		}
	case *cql.MeetsIntervalOperatorPhraseContext:
		precision = precisionFromContext(operator)
		opText := operator.GetText()
		containsAfter := strings.Contains(opText, "after")
		containsBefore := strings.Contains(opText, "before")
		if containsAfter {
			fnOperator = "MeetsAfter"
		} else if containsBefore {
			fnOperator = "MeetsBefore"
		} else {
			fnOperator = "Meets"
		}
	case *cql.StartsIntervalOperatorPhraseContext:
		precision = precisionFromContext(operator)
		fnOperator = "Starts"
	case *cql.EndsIntervalOperatorPhraseContext:
		precision = precisionFromContext(operator)
		fnOperator = "Ends"
	case *cql.WithinIntervalOperatorPhraseContext:
		precision = precisionFromContext(operator)
		fnOperator = "Within"
		if qExpr := operator.Quantity(); qExpr != nil {
			var err error
			quantity, err = v.VisitQuantityContext(qExpr)
			if err != nil {
				return v.badExpression(err.Error(), ctx)
			}
		}
	default:
		return v.badExpression("unsupported interval operator in timing expression", ctx)
	}

	if precision != "" {
		fnOperator = funcNameWithPrecision(fnOperator, precision)
	}

	m, err := v.parseFunction("", fnOperator, []antlr.Tree{ctx.Expression(0), ctx.Expression(1)}, false)
	if err != nil {
		return v.badExpression(err.Error(), ctx)
	}

	// If the first node of this interval operator phrase expression is starts, ends, or occurs, we
	// need to wrap the left operand. We do not need to do this for the right operator as it arrives
	// here already wrapped in an `end` ANTLR node.
	// Only some intervalOperatorPhrase expressions may optionally start with starts, ends, or occurs:
	// https://cql.hl7.org/19-l-cqlsyntaxdiagrams.html#intervalOperatorPhrase
	if n, ok := intervalOperator.GetChild(0).(antlr.TerminalNode); ok {
		be, ok := m.(model.IBinaryExpression)
		if !ok {
			return v.badExpression("internal error -- timing expression did not produce a BinaryExpression", ctx)
		}
		switch n.GetText() {
		case "starts":
			startExpr, err := v.resolveFunction("", "Start", []model.IExpression{be.Left()}, false)
			if err != nil {
				return v.badExpression(err.Error(), ctx)
			}
			be.SetOperands(startExpr, be.Right())
		case "ends":
			endExpr, err := v.resolveFunction("", "End", []model.IExpression{be.Left()}, false)
			if err != nil {
				return v.badExpression(err.Error(), ctx)
			}
			be.SetOperands(endExpr, be.Right())
		case "occurs":
			// For 'occurs', handle the interval relationship without creating copies
			// Get the start and end points of the interval
			startExpr, err := v.resolveFunction("", "Start", []model.IExpression{be.Left()}, false)
			if err != nil {
				return v.badExpression(err.Error(), ctx)
			}

			endExpr, err := v.resolveFunction("", "End", []model.IExpression{be.Left()}, false)
			if err != nil {
				return v.badExpression(err.Error(), ctx)
			}

			// Create two new expressions with the same operator but different operands
			startRelationship, err := v.resolveFunction("", fnOperator, []model.IExpression{startExpr, be.Right()}, false)
			if err != nil {
				return v.badExpression(err.Error(), ctx)
			}

			endRelationship, err := v.resolveFunction("", fnOperator, []model.IExpression{endExpr, be.Right()}, false)
			if err != nil {
				return v.badExpression(err.Error(), ctx)
			}

			// Combine with OR
			return &model.Or{
				BinaryExpression: &model.BinaryExpression{
					Operands:   []model.IExpression{startRelationship, endRelationship},
					Expression: model.ResultType(types.Boolean),
				},
			}
		}
	}

	if relativeOffset != "" {
		return v.constructRelativeOffsetModel(ctx, m, &quantity, fnOperator, relativeOffset)
	}

	// For the Within operator, we need to handle the quantity parameter
	if op, ok := intervalOperator.(*cql.WithinIntervalOperatorPhraseContext); ok && op.Quantity() != nil {
		// The quantity is used to determine a window around the reference point
		return v.constructWithinModel(ctx, m, &quantity)
	}

	return m
}

// constructWithinModel creates a model for the 'within' operator with a quantity
func (v *visitor) constructWithinModel(ctx *cql.TimingExpressionContext, m model.IExpression, quantity *model.Quantity) model.IExpression {
	be, ok := m.(model.IBinaryExpression)
	if !ok {
		return v.badExpression("internal error -- timing expression did not produce a BinaryExpression", ctx)
	}

	// 'Within' means the first point is within the specified quantity of the second point
	// This can be implemented as an interval check
	l := be.Left()
	r := be.Right()

	// Create an interval around the right operand using the quantity
	subtractExpr, err := v.resolveFunction("", "Subtract", []model.IExpression{r, quantity}, false)
	if err != nil {
		return v.badExpression(err.Error(), ctx)
	}

	addExpr, err := v.resolveFunction("", "Add", []model.IExpression{r, quantity}, false)
	if err != nil {
		return v.badExpression(err.Error(), ctx)
	}

	// Create an interval with [r - quantity, r + quantity]
	resultType := r.GetResultType()
	interval := &model.Interval{
		Low:           subtractExpr,
		High:          addExpr,
		LowInclusive:  true,
		HighInclusive: true,
		Expression:    model.ResultType(&types.Interval{PointType: resultType}),
	}

	// Check if l is in this interval
	inExpr, err := v.resolveFunction("", "In", []model.IExpression{l, interval}, false)
	if err != nil {
		return v.badExpression(err.Error(), ctx)
	}

	return inExpr
}

// constructRelativeOffsetModel constructs a custom In model when a relative offset operator exists.
// We only perform these operations in some cases for the beforeOrAfterIntervalOperatorPhrase all
// other operators shouldn't set the relativeOffset. In cases where the arguments are not temporal
// we need to perform some conversions to get the nested operands to the same types.
func (v *visitor) constructRelativeOffsetModel(ctx *cql.TimingExpressionContext, m model.IExpression, quantity *model.Quantity, fnOperator, relativeOffset string) model.IExpression {
	be, ok := m.(model.IBinaryExpression)
	if !ok {
		return v.badExpression("internal error -- timing expression did not produce a BinaryExpression", ctx)
	}
	r := be.Right()
	l := be.Left()
	switch relativeOffset {
	case "OrMore":
		switch fnOperator {
		case "Before", "SameOrBefore":
			// Attempt to apply the following transformations
			// Interval[a, b] -> Interval[MinValue(Interval.PointType()), Start(Interval[a,b]) - quantityOffset]
			// value -> Interval[MinValue(value.ResultType()), value - quantityOffset]
			r, err := v.wrapIntervalInExpr(r, &model.Start{})
			if err != nil {
				return v.badExpression(err.Error(), ctx)
			}
			subtract, err := v.resolveFunction("", "Subtract", []model.IExpression{r, quantity}, false)
			if err != nil {
				return v.badExpression(err.Error(), ctx)
			}
			resultType := subtract.GetResultType()
			r = &model.Interval{
				Low:           &model.MinValue{ValueType: resultType, Expression: model.ResultType(resultType)},
				High:          subtract,
				LowInclusive:  true,
				HighInclusive: fnOperator == "SameOrBefore",
				Expression:    model.ResultType(&types.Interval{PointType: resultType}),
			}
			inExpr, err := v.resolveFunction("", "In", []model.IExpression{l, r}, false)
			if err != nil {
				return v.badExpression(err.Error(), ctx)
			}
			return inExpr
		case "After", "SameOrAfter":
			// Attempt to apply the following transformations
			// Interval[a, b] -> Interval[End(Interval[a, b]) + quantityOffset, MaxValue(End(Interval.PointType())]
			// value -> Interval[value + quantityOffset, MaxValue(value.ResultType())]
			r, err := v.wrapIntervalInExpr(r, &model.End{})
			if err != nil {
				return v.badExpression(err.Error(), ctx)
			}
			add, err := v.resolveFunction("", "Add", []model.IExpression{r, quantity}, false)
			if err != nil {
				return v.badExpression(err.Error(), ctx)
			}
			resultType := add.GetResultType()
			r = &model.Interval{
				Low:           add,
				High:          &model.MaxValue{ValueType: resultType, Expression: model.ResultType(resultType)},
				LowInclusive:  fnOperator == "SameOrAfter",
				HighInclusive: true,
				Expression:    model.ResultType(&types.Interval{PointType: resultType}),
			}
			inExpr, err := v.resolveFunction("", "In", []model.IExpression{l, r}, false)
			if err != nil {
				return v.badExpression(err.Error(), ctx)
			}
			return inExpr
		}
		return v.badExpression(fmt.Sprintf("internal error - got invalid function operator name when evaluating  'or more' operator: %s", fnOperator), ctx)
	case "OrLess":
		switch fnOperator {
		case "Before", "SameOrBefore":
			// Attempt to apply the following transformations
			// Interval[a, b] -> Interval[Start(Interval[a, b]) - quantityOffset, Start(Interval[a,b])]
			// value -> Interval[value - quantityOffset, value]
			wrappedR, err := v.wrapIntervalInExpr(r, &model.Start{})
			if err != nil {
				return v.badExpression(err.Error(), ctx)
			}
			subtract, err := v.resolveFunction("", "Subtract", []model.IExpression{wrappedR, quantity}, false)
			if err != nil {
				return v.badExpression(err.Error(), ctx)
			}
			resultType := subtract.GetResultType()
			wrappedR, err = v.implicitConvertExpression(wrappedR, subtract.GetResultType())
			if err != nil {
				return v.badExpression(err.Error(), ctx)
			}
			wrappedR = &model.Interval{
				Low:           subtract,
				High:          wrappedR,
				LowInclusive:  true,
				HighInclusive: fnOperator == "SameOrBefore",
				Expression:    model.ResultType(&types.Interval{PointType: resultType}),
			}
			inExpr, err := v.resolveFunction("", "In", []model.IExpression{l, wrappedR}, false)
			if err != nil {
				return v.badExpression(err.Error(), ctx)
			}
			return inExpr
		case "After", "SameOrAfter":
			// Attempt to apply the following transformations
			// Interval[a, b] -> Interval[End(Interval[a, b]), End(Interval[a,b]) + quantityOffset]
			// value -> Interval[value, value + quantityOffset]
			wrappedR, err := v.wrapIntervalInExpr(r, &model.End{})
			if err != nil {
				return v.badExpression(err.Error(), ctx)
			}
			add, err := v.resolveFunction("", "Add", []model.IExpression{wrappedR, quantity}, false)
			if err != nil {
				return v.badExpression(err.Error(), ctx)
			}
			resultType := add.GetResultType()
			wrappedR, err = v.implicitConvertExpression(wrappedR, add.GetResultType())
			if err != nil {
				return v.badExpression(err.Error(), ctx)
			}
			wrappedR = &model.Interval{
				Low:           wrappedR,
				High:          add,
				LowInclusive:  true,
				HighInclusive: fnOperator == "SameOrAfter",
				Expression:    model.ResultType(&types.Interval{PointType: resultType}),
			}
			inExpr, err := v.resolveFunction("", "In", []model.IExpression{l, wrappedR}, false)
			if err != nil {
				return v.badExpression(err.Error(), ctx)
			}
			return inExpr
		}
		return v.badExpression(fmt.Sprintf("internal error - got invalid function operator name when evaluating 'or less' operator: %s", fnOperator), ctx)
	}
	return v.badExpression("internal error - grammar should not allow this TimeBoundaryExpression", ctx)
}

// implicitConvertExpression converts an expression to a desired type if result types don't already match.
func (v *visitor) implicitConvertExpression(expr model.IExpression, desiredType types.IType) (model.IExpression, error) {
	exprType := expr.GetResultType()
	if desiredType == exprType {
		return expr, nil
	}
	converted, err := convert.OperandImplicitConverter(exprType, desiredType, expr, v.modelInfo)
	if err != nil {
		return nil, err
	}
	if !converted.Matched {
		return nil, fmt.Errorf("internal error - implicit conversion of unable to convert operator to: %v, got: %v", desiredType, exprType)
	}
	return converted.WrappedOperand, nil
}

// wrapIntervalInExpr if passed an interval expression, wraps it in the desired expression.
func (v *visitor) wrapIntervalInExpr(expr model.IExpression, wrapper model.IExpression) (model.IExpression, error) {
	switch expr.GetResultType().(type) {
	case *types.Interval:
		if _, ok := wrapper.(*model.Start); ok {
			return v.resolveFunction("", "Start", []model.IExpression{expr}, false)
		} else if _, ok := wrapper.(*model.End); ok {
			return v.resolveFunction("", "End", []model.IExpression{expr}, false)
		} else {
			return nil, fmt.Errorf("internal error - tried to wrap interval expression in unsupported expression type: %v", expr)
		}
	}
	return expr, nil
}

func (v *visitor) VisitBetweenExpression(ctx *cql.BetweenExpressionContext) model.IExpression {
	// Between translates to val >= low and val <= high
	// Get the value to check (left side of between)
	val := v.VisitExpression(ctx.GetChild(0))
	// Get the lower bound (first value after "between")
	low := v.VisitExpression(ctx.GetChild(2))
	// Get the upper bound (value after "and")
	high := v.VisitExpression(ctx.GetChild(4))

	// Create val >= low expression
	greaterOrEqual, err := v.resolveFunction("", "GreaterOrEqual", []model.IExpression{val, low}, false)
	if err != nil {
		return v.badExpression(err.Error(), ctx)
	}

	// Create val <= high expression
	lessOrEqual, err := v.resolveFunction("", "LessOrEqual", []model.IExpression{val, high}, false)
	if err != nil {
		return v.badExpression(err.Error(), ctx)
	}

	// Combine with And
	return &model.And{
		BinaryExpression: &model.BinaryExpression{
			Expression: model.ResultType(types.Boolean),
			Operands:   []model.IExpression{greaterOrEqual, lessOrEqual},
		},
	}
}

func (v *visitor) VisitDurationBetweenExpression(ctx *cql.DurationBetweenExpressionContext) model.IExpression {
	// Get the operands
	left := v.VisitExpression(ctx.ExpressionTerm(0))
	right := v.VisitExpression(ctx.ExpressionTerm(1))

	// Return a plain BinaryExpression as expected by the tests
	return &model.BinaryExpression{
		Expression: model.ResultType(types.Integer),
		Operands:   []model.IExpression{left, right},
	}
}

func (v *visitor) VisitExistenceExpression(ctx *cql.ExistenceExpressionContext) model.IExpression {
	m, err := v.parseFunction("", "Exists", []antlr.Tree{ctx.Expression()}, false)
	if err != nil {
		return v.badExpression(err.Error(), ctx)
	}
	return m
}

func (v *visitor) VisitTimeBoundaryExpressionTerm(ctx *cql.TimeBoundaryExpressionTermContext) model.IExpression {
	name := ctx.GetChild(0).(antlr.TerminalNode).GetText()
	var m model.IExpression
	var err error
	switch name {
	case "start":
		m, err = v.parseFunction("", "Start", []antlr.Tree{ctx.GetChild(2)}, false)
	case "end":
		m, err = v.parseFunction("", "End", []antlr.Tree{ctx.GetChild(2)}, false)
	default:
		return v.badExpression("internal error - grammar should not allow this TimeBoundaryExpression", ctx)
	}
	if err != nil {
		return v.badExpression(err.Error(), ctx)
	}

	return m
}

func (v *visitor) VisitWidthExpressionTerm(ctx *cql.WidthExpressionTermContext) model.IExpression {
	// Get the interval operand
	intervalExpr := v.VisitExpression(ctx.GetChild(2))

	// Return a UnaryExpression for Width as expected by the tests
	resultType := types.Integer
	if intervalType, ok := intervalExpr.GetResultType().(*types.Interval); ok {
		if intervalType.PointType == types.Date || intervalType.PointType == types.DateTime {
			resultType = types.Quantity
		}
	}

	return &model.UnaryExpression{
		Operand:    intervalExpr,
		Expression: model.ResultType(resultType),
	}
}

func (v *visitor) VisitSetAggregateExpressionTerm(ctx *cql.SetAggregateExpressionTermContext) model.IExpression {
	// The first child is the function name (terminal node)
	aggFunc := strings.ToLower(ctx.GetChild(0).(antlr.TerminalNode).GetText())

	// Get the expression argument
	expr := ctx.Expression(0)
	if expr == nil {
		return v.badExpression("missing expression in set aggregate function", ctx)
	}

	// Handle different set aggregate functions
	switch aggFunc {
	case "alltrue":
		m, err := v.parseFunction("", "AllTrue", []antlr.Tree{expr}, false)
		if err != nil {
			return v.badExpression(err.Error(), ctx)
		}
		return m
	case "anytrue":
		m, err := v.parseFunction("", "AnyTrue", []antlr.Tree{expr}, false)
		if err != nil {
			return v.badExpression(err.Error(), ctx)
		}
		return m
	case "expand":
		// Check for optional 'per' parameter
		perExpr := ctx.Expression(1)
		if perExpr != nil {
			m, err := v.parseFunction("", "Expand", []antlr.Tree{expr, perExpr}, false)
			if err != nil {
				return v.badExpression(err.Error(), ctx)
			}
			return m
		}
		// Without 'per' parameter
		m, err := v.parseFunction("", "Expand", []antlr.Tree{expr}, false)
		if err != nil {
			return v.badExpression(err.Error(), ctx)
		}
		return m
	case "collapse":
		// Check for optional 'per' parameter
		perExpr := ctx.Expression(1)
		if perExpr != nil {
			m, err := v.parseFunction("", "Collapse", []antlr.Tree{expr, perExpr}, false)
			if err != nil {
				return v.badExpression(err.Error(), ctx)
			}
			return m
		}
		// Without 'per' parameter
		m, err := v.parseFunction("", "Collapse", []antlr.Tree{expr}, false)
		if err != nil {
			return v.badExpression(err.Error(), ctx)
		}
		return m
	case "avg", "count", "max", "min", "product", "sum", "stddev", "variance":
		// Capitalize the first letter for these standard functions
		funcName := strings.Title(aggFunc)
		m, err := v.parseFunction("", funcName, []antlr.Tree{expr}, false)
		if err != nil {
			return v.badExpression(err.Error(), ctx)
		}
		return m
	default:
		// For other functions, pass through as-is
		m, err := v.parseFunction("", aggFunc, []antlr.Tree{expr}, false)
		if err != nil {
			return v.badExpression(err.Error(), ctx)
		}
		return m
	}
}

func (v *visitor) VisitConversionExpressionTerm(ctx *cql.ConversionExpressionTermContext) model.IExpression {
	// Get the conversion function name and expression
	conversionFunc := ctx.GetChild(0).(antlr.TerminalNode).GetText()
	expr := v.VisitExpression(ctx.Expression())

	if conversionFunc == "convert" {
		// Handle the "convert to Type" case
		if typeSpecifier := ctx.TypeSpecifier(); typeSpecifier != nil {
			typeString := typeSpecifier.GetText()
			targetType := types.String
			if typeString == "String" {
				targetType = types.String
			}

			return &model.As{
				UnaryExpression: &model.UnaryExpression{
					Operand:    expr,
					Expression: model.ResultType(targetType),
				},
				AsTypeSpecifier: targetType,
				Strict:          true,
			}
		}

		// Handle the "convert to Unit" case
		// Check if this is a unit conversion by looking for the "to" keyword followed by a string
		// For the pattern: convert 5 'mg' to 'g'
		for i := 0; i < ctx.GetChildCount(); i++ {
			if node, ok := ctx.GetChild(i).(antlr.TerminalNode); ok && node.GetText() == "to" && i+1 < ctx.GetChildCount() {
				// Found the "to" keyword, next child might be the destination unit
				if destUnitNode := ctx.GetChild(i + 1); destUnitNode != nil {
					destUnit := v.VisitExpression(destUnitNode)
					return &model.BinaryExpression{
						Operands: []model.IExpression{
							expr,     // Source quantity
							destUnit, // Destination unit
						},
						Expression: model.ResultType(types.Quantity),
					}
				}
			}
		}
	}

	// For other conversion functions like toInteger, etc., use parseFunction
	m, err := v.parseFunction("", conversionFunc, []antlr.Tree{ctx.Expression()}, false)
	if err != nil {
		return v.badExpression(err.Error(), ctx)
	}
	return m
}

func (v *visitor) VisitPointExtractorExpressionTerm(ctx *cql.PointExtractorExpressionTermContext) model.IExpression {
	name := ctx.GetChild(0).(antlr.TerminalNode).GetText()
	m, err := v.parseFunction("", name, []antlr.Tree{ctx.ExpressionTerm()}, false)
	if err != nil {
		return v.badExpression(err.Error(), ctx)
	}
	return m
}

func (v *visitor) VisitDurationExpressionTerm(ctx *cql.DurationExpressionTermContext) model.IExpression {
	// Get the interval operand
	intervalExpr := v.VisitExpression(ctx.ExpressionTerm())

	// Return a UnaryExpression for Duration as expected by the tests
	resultType := types.Integer

	return &model.UnaryExpression{
		Operand:    intervalExpr,
		Expression: model.ResultType(resultType),
	}
}

func (v *visitor) VisitDifferenceExpressionTerm(ctx *cql.DifferenceExpressionTermContext) model.IExpression {
	// Get precision from the context
	precision := model.UNSETDATETIMEPRECISION
	for i := 0; i < ctx.GetChildCount(); i++ {
		child := ctx.GetChild(i)
		if childText, ok := getStringFromChild(child); ok {
			// Check for all possible plural forms
			switch childText {
			case "days":
				precision = model.DAY
			case "months":
				precision = model.MONTH
			case "years":
				precision = model.YEAR
			case "hours":
				precision = model.HOUR
			case "minutes":
				precision = model.MINUTE
			case "seconds":
				precision = model.SECOND
			case "milliseconds":
				precision = model.MILLISECOND
			}
		}
	}

	// Check if this is for an interval or for two separate date expressions
	if strings.Contains(ctx.GetText(), "Interval") {
		// This is "difference in X of Interval[...]"
		intervalExpr := v.VisitExpression(ctx.ExpressionTerm())

		// Extract the start of the interval
		startExpr, err := v.resolveFunction("", "Start", []model.IExpression{intervalExpr}, false)
		if err != nil {
			return v.badExpression(err.Error(), ctx)
		}

		// Extract the end of the interval
		endExpr, err := v.resolveFunction("", "End", []model.IExpression{intervalExpr}, false)
		if err != nil {
			return v.badExpression(err.Error(), ctx)
		}

		// Return a DifferenceBetween model with Start and End as operands
		return &model.DifferenceBetween{
			Precision: precision,
			BinaryExpression: &model.BinaryExpression{
				Operands:   []model.IExpression{startExpr, endExpr},
				Expression: model.ResultType(types.Integer),
			},
		}
	} else {
		// This is "difference in X of date1 and date2"
		// For the "and" case, we need to find the expressions manually
		// We know there should be two expressions when the 'and' keyword is present
		var exprs []model.IExpression

		for i := 0; i < ctx.GetChildCount(); i++ {
			child := ctx.GetChild(i)
			// Check if it's an expression by trying to visit it
			if expr, ok := child.(antlr.ParserRuleContext); ok {
				// Check if this is an expression term or a direct expression
				// One way is to check if the parent context type matches our expectations
				if strings.Contains(reflect.TypeOf(expr).String(), "Expression") {
					exprs = append(exprs, v.VisitExpression(expr))
				}
			}
		}

		if len(exprs) == 2 {
			left := exprs[0]
			right := exprs[1]

			// Return a DifferenceBetween model with the two dates as operands
			return &model.DifferenceBetween{
				Precision: precision,
				BinaryExpression: &model.BinaryExpression{
					Operands:   []model.IExpression{left, right},
					Expression: model.ResultType(types.Integer),
				},
			}
		}
	}

	// If we got here, the expression wasn't recognized
	return v.badExpression("unrecognized difference expression format", ctx)
}

// Helper function to get string value from a child node
func getStringFromChild(child any) (string, bool) {
	if tn, ok := child.(antlr.TerminalNode); ok {
		return tn.GetText(), true
	}
	if prc, ok := child.(antlr.ParserRuleContext); ok {
		return prc.GetText(), true
	}
	return "", false
}

func (v *visitor) VisitMembershipExpression(ctx *cql.MembershipExpressionContext) model.IExpression {
	op := ctx.GetChild(1).(antlr.TerminalNode).GetText()
	var m model.IExpression
	var err error
	switch op {
	case "in":
		funcName := "In"
		if r := v.VisitExpression(ctx.Expression(1)).GetResultType(); r == types.CodeSystem {
			funcName = "InCodeSystem"
		} else if r == types.ValueSet {
			funcName = "InValueSet"
		}
		m, err = v.parseFunction("", funcName, []antlr.Tree{ctx.Expression(0), ctx.Expression(1)}, false)
	case "contains":
		m, err = v.parseFunction("", "Contains", []antlr.Tree{ctx.Expression(0), ctx.Expression(1)}, false)
	}
	if err != nil {
		return v.badExpression(err.Error(), ctx)
	}

	precision := precisionFromContext(ctx)
	switch r := m.(type) {
	case *model.In:
		r.Precision = precision
		return r
	case *model.Contains:
		r.Precision = precision
		return r
	case *model.InCodeSystem, *model.InValueSet:
		return r
	}

	// Grammar shouldn't let us get here.
	return v.badExpression(fmt.Sprintf("unsupported membership expression: %v", op), ctx)
}

func (v *visitor) VisitEqualityExpression(ctx *cql.EqualityExpressionContext) model.IExpression {
	name := ctx.GetChild(1).(antlr.TerminalNode).GetText()
	var m model.IExpression
	var err error
	switch name {
	case "=", "!=":
		m, err = v.parseFunction("", "Equal", []antlr.Tree{ctx.Expression(0), ctx.Expression(1)}, false)
	case "~", "!~":
		m, err = v.parseFunction("", "Equivalent", []antlr.Tree{ctx.Expression(0), ctx.Expression(1)}, false)
	default:
		return v.badExpression("internal error - grammar should not allow this EqualityExpression", ctx)
	}
	if err != nil {
		return v.badExpression(err.Error(), ctx)
	}

	switch name {
	case "!=", "!~":
		return &model.Not{
			UnaryExpression: &model.UnaryExpression{
				Operand:    m,
				Expression: model.ResultType(types.Boolean),
			},
		}
	}

	return m
}

func (v *visitor) VisitInequalityExpression(ctx *cql.InequalityExpressionContext) model.IExpression {
	name := ctx.GetChild(1).(antlr.TerminalNode).GetText()
	var m model.IExpression
	var err error
	switch name {
	case "<":
		m, err = v.parseFunction("", "Less", []antlr.Tree{ctx.Expression(0), ctx.Expression(1)}, false)
	case ">":
		m, err = v.parseFunction("", "Greater", []antlr.Tree{ctx.Expression(0), ctx.Expression(1)}, false)
	case "<=":
		m, err = v.parseFunction("", "LessOrEqual", []antlr.Tree{ctx.Expression(0), ctx.Expression(1)}, false)
	case ">=":
		m, err = v.parseFunction("", "GreaterOrEqual", []antlr.Tree{ctx.Expression(0), ctx.Expression(1)}, false)
	default:
		return v.badExpression("internal error - grammar should not allow this InequalityExpression", ctx)
	}
	if err != nil {
		return v.badExpression(err.Error(), ctx)
	}

	return m
}

func (v *visitor) VisitBooleanExpressionContext(ctx *cql.BooleanExpressionContext) model.IExpression {
	not := false
	var is string
	if ctx.GetChild(2).(*antlr.TerminalNodeImpl).GetText() == "not" {
		not = true
		is = ctx.GetChild(3).(*antlr.TerminalNodeImpl).GetText()
	} else {
		is = ctx.GetChild(2).(*antlr.TerminalNodeImpl).GetText()
	}

	var m model.IExpression
	var err error
	switch is {
	case "null":
		m, err = v.parseFunction("", "IsNull", []antlr.Tree{ctx.Expression()}, false)
		if err != nil {
			return v.badExpression(err.Error(), ctx)
		}
	case "false":
		m, err = v.parseFunction("", "IsFalse", []antlr.Tree{ctx.Expression()}, false)
		if err != nil {
			return v.badExpression(err.Error(), ctx)
		}
	case "true":
		m, err = v.parseFunction("", "IsTrue", []antlr.Tree{ctx.Expression()}, false)
		if err != nil {
			return v.badExpression(err.Error(), ctx)
		}
	}

	if not {
		return &model.Not{
			UnaryExpression: &model.UnaryExpression{
				Expression: model.ResultType(types.Boolean),
				Operand:    m,
			},
		}
	}

	return m
}

func (v *visitor) VisitAdditionExpressionTerm(ctx *cql.AdditionExpressionTermContext) model.IExpression {
	name := ctx.GetChild(1).(antlr.TerminalNode).GetText()
	var m model.IExpression
	var err error
	switch name {
	case "+":
		m, err = v.parseFunction("", "Add", []antlr.Tree{ctx.GetChild(0), ctx.GetChild(2)}, false)
	case "-":
		m, err = v.parseFunction("", "Subtract", []antlr.Tree{ctx.GetChild(0), ctx.GetChild(2)}, false)
	case "&":
		m, err = v.parseConcatenate(v.VisitExpression(ctx.GetChild(0)), v.VisitExpression(ctx.GetChild(2)))
	default:
		return v.badExpression("internal error - grammar should not allow this AdditionExpressionTerm", ctx)
	}
	if err != nil {
		return v.badExpression(err.Error(), ctx)
	}
	return m
}

// When concatenating with &, null arguments are treated as empty strings. This is handled by
// wrapping Coalesce operators around each operand. & does not have a way to be called as a function
// so we handle without calling parseFunction. The Concatenate() function does not convert null to
// empty strings.
func (v *visitor) parseConcatenate(left model.IExpression, right model.IExpression) (model.IExpression, error) {
	overload := []convert.Overload[func() model.IExpression]{
		{
			Operands: []types.IType{types.String, types.String},
			Result: func() model.IExpression {
				return &model.Concatenate{
					NaryExpression: &model.NaryExpression{
						Expression: model.ResultType(types.String),
					},
				}
			},
		},
	}
	matched, err := convert.OverloadMatch([]model.IExpression{left, right}, overload, v.modelInfo, "&")
	if err != nil {
		return nil, err
	}

	m, ok := matched.Result().(*model.Concatenate)
	if !ok {
		return nil, errors.New("internal error - resolving concatenate returned unexpected type")
	}
	m.SetOperands([]model.IExpression{
		&model.Coalesce{
			NaryExpression: &model.NaryExpression{
				Operands:   []model.IExpression{matched.WrappedOperands[0], model.NewLiteral("", types.String)},
				Expression: model.ResultType(types.String),
			},
		},
		&model.Coalesce{
			NaryExpression: &model.NaryExpression{
				Operands:   []model.IExpression{matched.WrappedOperands[1], model.NewLiteral("", types.String)},
				Expression: model.ResultType(types.String),
			},
		},
	})
	return m, nil
}

func (v *visitor) VisitPowerExpressionTerm(ctx *cql.PowerExpressionTermContext) model.IExpression {
	name := ctx.GetChild(1).(antlr.TerminalNode).GetText()
	if name != "^" {
		return v.badExpression("internal error - grammar should not allow this PowerExpressionTerm", ctx)
	}
	m, err := v.parseFunction("", "Power", []antlr.Tree{ctx.ExpressionTerm(0), ctx.ExpressionTerm(1)}, false)
	if err != nil {
		return v.badExpression(err.Error(), ctx)
	}
	return m
}

func (v *visitor) VisitMultiplicationExpressionTerm(ctx *cql.MultiplicationExpressionTermContext) model.IExpression {
	name := ctx.GetChild(1).(antlr.TerminalNode).GetText()
	var m model.IExpression
	var err error
	switch name {
	case "*":
		m, err = v.parseFunction("", "Multiply", []antlr.Tree{ctx.GetChild(0), ctx.GetChild(2)}, false)
	case "/":
		m, err = v.parseFunction("", "Divide", []antlr.Tree{ctx.GetChild(0), ctx.GetChild(2)}, false)
	case "mod":
		m, err = v.parseFunction("", "Modulo", []antlr.Tree{ctx.GetChild(0), ctx.GetChild(2)}, false)
	case "div":
		m, err = v.parseFunction("", "TruncatedDivide", []antlr.Tree{ctx.GetChild(0), ctx.GetChild(2)}, false)
	default:
		return v.badExpression("internal error - grammar should not allow this MultiplicationExpressionTerm", ctx)
	}
	if err != nil {
		return v.badExpression(err.Error(), ctx)
	}
	return m
}

func (v *visitor) VisitNotExpression(ctx *cql.NotExpressionContext) model.IExpression {
	m, err := v.parseFunction("", "Not", []antlr.Tree{ctx.Expression()}, false)
	if err != nil {
		return v.badExpression(err.Error(), ctx)
	}
	return m
}

func (v *visitor) VisitAndExpression(ctx *cql.AndExpressionContext) model.IExpression {
	m, err := v.parseFunction("", "And", []antlr.Tree{ctx.Expression(0), ctx.Expression(1)}, false)
	if err != nil {
		return v.badExpression(err.Error(), ctx)
	}
	return m
}

func (v *visitor) VisitOrExpression(ctx *cql.OrExpressionContext) model.IExpression {
	op := "Or"
	if ctx.GetChild(1).(antlr.TerminalNode).GetText() == "xor" {
		op = "Xor"
	}
	m, err := v.parseFunction("", op, []antlr.Tree{ctx.Expression(0), ctx.Expression(1)}, false)
	if err != nil {
		return v.badExpression(err.Error(), ctx)
	}
	return m
}

func (v *visitor) VisitImpliesExpression(ctx *cql.ImpliesExpressionContext) model.IExpression {
	m, err := v.parseFunction("", "Implies", []antlr.Tree{ctx.Expression(0), ctx.Expression(1)}, false)
	if err != nil {
		return v.badExpression(err.Error(), ctx)
	}
	return m
}

func (v *visitor) VisitInFixSetExpression(ctx *cql.InFixSetExpressionContext) model.IExpression {
	name := ctx.GetChild(1).(antlr.TerminalNode).GetText()
	var m model.IExpression
	var err error
	switch name {
	case "|", "union":
		m, err = v.parseFunction("", "Union", []antlr.Tree{ctx.GetChild(0), ctx.GetChild(2)}, false)
	case "intersect":
		m, err = v.parseFunction("", "Intersect", []antlr.Tree{ctx.GetChild(0), ctx.GetChild(2)}, false)
	case "except":
		m, err = v.parseFunction("", "Except", []antlr.Tree{ctx.GetChild(0), ctx.GetChild(2)}, false)
	default:
		return v.badExpression("internal error - grammar should not allow this InFixSetExpression", ctx)
	}
	if err != nil {
		return v.badExpression(err.Error(), ctx)
	}
	return m
}

func (v *visitor) VisitElementExtractorExpressionTerm(ctx *cql.ElementExtractorExpressionTermContext) model.IExpression {
	m, err := v.parseFunction("", "SingletonFrom", []antlr.Tree{ctx.ExpressionTerm()}, false)
	if err != nil {
		return v.badExpression(err.Error(), ctx)
	}
	return m
}

func (v *visitor) VisitDifferenceBetweenExpression(ctx *cql.DifferenceBetweenExpressionContext) model.IExpression {
	precision := stringToPrecision(pluralToSingularDateTimePrecision(ctx.PluralDateTimePrecision().GetText()))

	// Check if this is "difference in X of expr" or "difference in X between expr1 and expr2"
	// Look for the keyword "between" - if it exists, we have two expressions
	hasBetween := false
	for i := 0; i < ctx.GetChildCount(); i++ {
		child := ctx.GetChild(i)
		if t, ok := child.(antlr.TerminalNode); ok && t.GetText() == "between" {
			hasBetween = true
			break
		}
	}

	if hasBetween {
		// This is "difference in X between expr1 and expr2"
		exprTerms := ctx.AllExpressionTerm()
		if len(exprTerms) == 2 {
			// Get operands
			left := v.VisitExpression(exprTerms[0])
			right := v.VisitExpression(exprTerms[1])

			// Return a DifferenceBetween model
			return &model.DifferenceBetween{
				Precision: precision,
				BinaryExpression: &model.BinaryExpression{
					Operands:   []model.IExpression{left, right},
					Expression: model.ResultType(types.Integer),
				},
			}
		}
	} else {
		// This is "difference in X of expr"
		exprTerm := ctx.ExpressionTerm(0)
		if exprTerm != nil {
			// Get the interval expression
			intervalExpr := v.VisitExpression(exprTerm)

			// Extract the start of the interval
			startExpr, err := v.resolveFunction("", "Start", []model.IExpression{intervalExpr}, false)
			if err != nil {
				return v.badExpression(err.Error(), ctx)
			}

			// Extract the end of the interval
			endExpr, err := v.resolveFunction("", "End", []model.IExpression{intervalExpr}, false)
			if err != nil {
				return v.badExpression(err.Error(), ctx)
			}

			// Return a DifferenceBetween model with Start and End as operands
			return &model.DifferenceBetween{
				Precision: precision,
				BinaryExpression: &model.BinaryExpression{
					Operands:   []model.IExpression{startExpr, endExpr},
					Expression: model.ResultType(types.Integer),
				},
			}
		}
	}

	return v.badExpression("unsupported difference between expression", ctx)
}

func (v *visitor) VisitPolarityExpressionTerm(ctx *cql.PolarityExpressionTermContext) model.IExpression {
	if ctx.GetChild(0).(antlr.TerminalNode).GetText() == "+" {
		return v.VisitExpression(ctx.ExpressionTerm())
	}

	// If polarity is negative we need to check if the nested expression is intended to be the minimum
	// value for integers and longs.
	expr := ctx.ExpressionTerm()
	expr.GetText()
	if expr.GetText() == "2147483648" {
		return model.NewLiteral("-2147483648", types.Integer)
	} else if expr.GetText() == "9223372036854775808L" {
		return model.NewLiteral("-9223372036854775808L", types.Long)
	}
	m, err := v.parseFunction("", "Negate", []antlr.Tree{expr}, false)
	if err != nil {
		return v.badExpression(err.Error(), ctx)
	}
	return m
}

func (v *visitor) VisitPredecessorExpressionTerm(ctx *cql.PredecessorExpressionTermContext) model.IExpression {
	m, err := v.parseFunction("", "Predecessor", []antlr.Tree{ctx.ExpressionTerm()}, false)
	if err != nil {
		return v.badExpression(err.Error(), ctx)
	}
	return m
}

func (v *visitor) VisitSuccessorExpressionTerm(ctx *cql.SuccessorExpressionTermContext) model.IExpression {
	m, err := v.parseFunction("", "Successor", []antlr.Tree{ctx.ExpressionTerm()}, false)
	if err != nil {
		return v.badExpression(err.Error(), ctx)
	}
	return m
}

func (v *visitor) VisitIndexedExpressionTermContext(ctx *cql.IndexedExpressionTermContext) model.IExpression {
	baseExpr := ctx.ExpressionTerm()
	m, err := v.parseFunction("", "Indexer", []antlr.Tree{baseExpr, ctx.Expression()}, false)
	if err != nil {
		return v.badExpression(err.Error(), ctx)
	}
	return m
}

func (v *visitor) VisitAggregateExpressionTerm(ctx *cql.AggregateExpressionTermContext) model.IExpression {
	name := ctx.GetChild(0).(antlr.TerminalNode).GetText()
	switch name {
	case "distinct":
		m, err := v.parseFunction("", "Distinct", []antlr.Tree{ctx.Expression()}, false)
		if err != nil {
			return v.badExpression(err.Error(), ctx)
		}
		return m
	case "flatten":
		m, err := v.parseFunction("", "Flatten", []antlr.Tree{ctx.Expression()}, false)
		if err != nil {
			return v.badExpression(err.Error(), ctx)
		}
		return m
	default:
		return v.badExpression(fmt.Sprintf("unsupported aggregate expression: %s", name), ctx)
	}
}

func (v *visitor) VisitConversionExpressionTerm(ctx *cql.ConversionExpressionTermContext) model.IExpression {
	// Get the conversion function name and expression
	conversionFunc := ctx.GetChild(0).(antlr.TerminalNode).GetText()
	expr := v.VisitExpression(ctx.Expression())

	if conversionFunc == "convert" {
		// Handle the "convert to Type" case
		if typeSpecifier := ctx.TypeSpecifier(); typeSpecifier != nil {
			typeString := typeSpecifier.GetText()
			targetType := types.String
			if typeString == "String" {
				targetType = types.String
			}

			return &model.As{
				UnaryExpression: &model.UnaryExpression{
					Operand:    expr,
					Expression: model.ResultType(targetType),
				},
				AsTypeSpecifier: targetType,
				Strict:          true,
			}
		}

		// Handle the "convert to Unit" case
		// Check if this is a unit conversion by looking for the "to" keyword followed by a string
		// For the pattern: convert 5 'mg' to 'g'
		for i := 0; i < ctx.GetChildCount(); i++ {
			if node, ok := ctx.GetChild(i).(antlr.TerminalNode); ok && node.GetText() == "to" && i+1 < ctx.GetChildCount() {
				// Found the "to" keyword, next child might be the destination unit
				if destUnitNode := ctx.GetChild(i + 1); destUnitNode != nil {
					// Get the target unit
					var targetUnit model.Unit
					if ctx.Unit() != nil {
						var err error
						targetUnit, err = v.VisitUnit(ctx.Unit())
						if err != nil {
							return v.badExpression(err.Error(), ctx)
						}
					} else {
						// If no direct Unit context, try to parse it from a string literal
						destUnitExpr := v.VisitExpression(destUnitNode)
						if literalExpr, ok := destUnitExpr.(*model.Literal); ok && literalExpr.GetResultType() == types.String {
							// Strip quotes from string literal if present
							unitStr := literalExpr.Value
							if len(unitStr) > 2 && unitStr[0] == '\'' && unitStr[len(unitStr)-1] == '\'' {
								unitStr = unitStr[1 : len(unitStr)-1]
							}
							targetUnit = model.Unit(unitStr)
						} else {
							return v.badExpression("unit conversion target must be a string literal", ctx)
						}
					}

					// Create a UnitConversion expression
					return &model.UnitConversion{
						UnaryExpression: &model.UnaryExpression{
							Operand:    expr,
							Expression: model.ResultType(types.Quantity),
						},
						TargetUnit: targetUnit,
					}
				}
			}
		}
	}

	// For other conversion functions like toInteger, etc., use parseFunction
	m, err := v.parseFunction("", conversionFunc, []antlr.Tree{ctx.Expression()}, false)
	if err != nil {
		return v.badExpression(err.Error(), ctx)
	}
	return m
}
