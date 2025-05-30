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
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/cql/internal/convert"
	"github.com/google/cql/internal/datehelpers"
	"github.com/google/cql/internal/embeddata/third_party/cqframework/cql"
	"github.com/google/cql/model"
	"github.com/google/cql/types"
	"github.com/google/cql/ucum"
	"github.com/antlr4-go/antlr/v4"
)

func (v *visitor) VisitExpression(tree antlr.Tree) model.IExpression {
	// Manual dispatch of needed due to
	// https://github.com/antlr/antlr4/issues/2504

	var m model.IExpression
	switch t := tree.(type) {
	case *cql.QualifiedIdentifierExpressionContext:
		m = v.VisitQualifiedIdentifierExpression(t)
	case *cql.RetrieveContext:
		m = v.VisitRetrieve(t)
	case *cql.RetrieveExpressionContext:
		m = v.VisitRetrieveExpression(t)
	case *cql.TypeExpressionContext:
		m = v.VisitTypeExpression(t)
	case *cql.CastExpressionContext:
		m = v.VisitCastExpression(t)
	case *cql.BooleanExpressionContext:
		m = v.VisitBooleanExpressionContext(t)
	case *cql.MembershipExpressionContext:
		m = v.VisitMembershipExpression(t)
	case *cql.BetweenExpressionContext:
		m = v.VisitBetweenExpression(t)
	case *cql.ReferentialIdentifierContext:
		m = v.VisitReferentialIdentifier(t)
	case *cql.FunctionContext:
		m = v.VisitFunction(t)
	case *cql.IfThenElseExpressionTermContext:
		m = v.VisitIfThenElseExpression(t)
	case *cql.IntervalSelectorTermContext:
		m = v.VisitIntervalSelectorTerm(t)
	case *cql.ListSelectorTermContext:
		m = v.VisitListSelectorTerm(t)
	case *cql.CodeSelectorTermContext:
		m = v.VisitCodeSelectorTerm(t)
	case *cql.SimpleStringLiteralContext:
		m = buildLiteral(unquoteString(t.GetText()), types.String)
	case *cql.SimpleNumberLiteralContext:
		m = buildNumberLiteral(t.GetText())
	case *cql.SimpleLiteralContext:
		m = v.VisitSimpleLiteral(t)
	case *cql.LiteralTermContext:
		m = v.VisitLiteralTerm(t)
	case *cql.TimeBoundaryExpressionTermContext:
		m = v.VisitTimeBoundaryExpressionTerm(t)
	case *cql.ExistenceExpressionContext:
		m = v.VisitExistenceExpression(t)
	case *cql.ParenthesizedTermContext:
		m = v.VisitParenthesizedTerm(t)
	case *cql.QuerySourceContext:
		m = v.VisitQuerySource(t)
	case *cql.QueryContext:
		m = v.VisitQuery(t)
	case *cql.EqualityExpressionContext:
		m = v.VisitEqualityExpression(t)
	case *cql.InequalityExpressionContext:
		m = v.VisitInequalityExpression(t)
	case *cql.DifferenceBetweenExpressionContext:
		m = v.VisitDifferenceBetweenExpression(t)
	case *cql.InvocationExpressionTermContext:
		m = v.VisitInvocationExpressionTerm(t)
	case *cql.TimingExpressionContext:
		m = v.VisitTimingExpression(t)
	case *cql.AndExpressionContext:
		m = v.VisitAndExpression(t)
	case *cql.OrExpressionContext:
		m = v.VisitOrExpression(t)
	case *cql.ImpliesExpressionContext:
		m = v.VisitImpliesExpression(t)
	case *cql.InFixSetExpressionContext:
		m = v.VisitInFixSetExpression(t)
	case *cql.AdditionExpressionTermContext:
		m = v.VisitAdditionExpressionTerm(t)
	case *cql.MultiplicationExpressionTermContext:
		m = v.VisitMultiplicationExpressionTerm(t)
	case *cql.PowerExpressionTermContext:
		m = v.VisitPowerExpressionTerm(t)
	case *cql.TimeUnitExpressionTermContext:
		m = v.VisitTimeUnitExpressionTerm(t)
	case *cql.TupleSelectorTermContext:
		m = v.VisitTupleSelectorTerm(t)
	case *cql.InstanceSelectorTermContext:
		m = v.VisitInstanceSelectorTerm(t)
	case *cql.CaseExpressionTermContext:
		m = v.VisitCaseExpressionTerm(t)
	case *cql.NotExpressionContext:
		m = v.VisitNotExpression(t)
	case *cql.PolarityExpressionTermContext:
		m = v.VisitPolarityExpressionTerm(t)
	case *cql.PredecessorExpressionTermContext:
		m = v.VisitPredecessorExpressionTerm(t)
	case *cql.SuccessorExpressionTermContext:
		m = v.VisitSuccessorExpressionTerm(t)
	case *cql.TypeExtentExpressionTermContext:
		m = v.VisitTypeExtentExpressionTermContext(t)
	case *cql.ElementExtractorExpressionTermContext:
		m = v.VisitElementExtractorExpressionTerm(t)
	case *cql.IndexedExpressionTermContext:
		m = v.VisitIndexedExpressionTermContext(t)
	case *cql.AggregateExpressionTermContext:
		m = v.VisitAggregateExpressionTerm(t)
	case *cql.ConversionExpressionTermContext:
		m = v.VisitConversionExpressionTerm(t)
	case *cql.WidthExpressionTermContext:
		m = v.VisitWidthExpressionTerm(t)
	case *cql.SetAggregateExpressionTermContext:
		m = v.VisitSetAggregateExpressionTerm(t)
	case *cql.ConversionExpressionTermContext:
		m = v.VisitConversionExpressionTerm(t)
	case *cql.PointExtractorExpressionTermContext:
		m = v.VisitPointExtractorExpressionTerm(t)
	case *cql.DurationExpressionTermContext:
		m = v.VisitDurationExpressionTerm(t)
	case *cql.DifferenceExpressionTermContext:
		m = v.VisitDifferenceExpressionTerm(t)
	case *cql.DurationBetweenExpressionContext:
		m = v.VisitDurationBetweenExpression(t)

		// All cases that have a single child and recurse to the child are handled below. For example in
		// the CQL grammar the only child of QueryExpression is Query, so QueryExpression can be handled
		// by recursing on it's only child.
	case *cql.StatementContext,
		*cql.DefinitionContext,
		*cql.TermExpressionTermContext,
		*cql.TermExpressionContext,
		*cql.QueryExpressionContext,
		*cql.InvocationTermContext,
		*cql.MemberInvocationContext,
		*cql.FunctionInvocationContext:
		m = v.VisitExpression(tree.GetChild(0))
	default:
		// Line and Column are not available for unsupported expressions.
		pe := &ParsingError{Message: fmt.Sprintf("Internal Error - unsupported expression: %#v", tree)}
		v.errors.Append(pe)
		return invalidExpression{ParsingError: pe, Expression: model.ResultType(types.Any)}
	}

	if m.GetResultType() == types.Unset {
		// Line and Column are not available.
		pe := &ParsingError{Message: fmt.Sprintf("Internal Error - Model Expression ResultType not set: %#v", m)}
		v.errors.Append(pe)
		return invalidExpression{ParsingError: pe, Expression: model.ResultType(types.Any)}
	}

	return m
}

// parseSTRING removes surrounding quotes from a STRING node that was produced using
// a call to `STRING()`. Grammar defined at https://cql.hl7.org/19-l-cqlsyntaxdiagrams.html#STRING.
func parseSTRING(n antlr.TerminalNode) string {
	s := n.GetText()

	// TODO(b/302003569): strings should also be unescaped.
	// CQL escaping rules do not match golang's so `strconv.Unquote()` cannot be used here.
	return s[1 : len(s)-1]
}

func (v *visitor) VisitParenthesizedTerm(ctx *cql.ParenthesizedTermContext) model.IExpression {
	return v.VisitExpression(ctx.GetChild(1))
}

func (v *visitor) VisitTimeUnitExpressionTerm(ctx *cql.TimeUnitExpressionTermContext) model.IExpression {
	// parses statements like: "date from expression"
	// TODO: b/301606416 - Implement time units where left is dateTimePrecision.
	dtc := ctx.GetChild(0).(*cql.DateTimeComponentContext)
	switch component := dtc.GetChild(0).(type) {
	case antlr.TerminalNode:
		dateTimeComponent := component.GetText()
		switch dateTimeComponent {
		case "date":
			return &model.ToDate{
				UnaryExpression: &model.UnaryExpression{
					Operand:    v.VisitExpression(ctx.ExpressionTerm()),
					Expression: model.ResultType(types.Date),
				},
			}
		}
	}
	return v.badExpression(fmt.Sprintf("unsupported date time component conversion (e.g. X in 'X from expression'). got: %s, only %v supported", dtc.GetText(), "date"), ctx)
}

func (v *visitor) VisitTupleSelectorTerm(ctx *cql.TupleSelectorTermContext) model.IExpression {
	tModel := &model.Tuple{}
	tResult := &types.Tuple{ElementTypes: make(map[string]types.IType)}
	for _, tes := range ctx.TupleSelector().AllTupleElementSelector() {
		elem := &model.TupleElement{
			Name:  v.parseReferentialIdentifier(tes.ReferentialIdentifier()),
			Value: v.VisitExpression(tes.Expression()),
		}
		tModel.Elements = append(tModel.Elements, elem)
		tResult.ElementTypes[elem.Name] = elem.Value.GetResultType()
	}
	tModel.Expression = model.ResultType(tResult)
	return tModel
}

func (v *visitor) VisitInstanceSelectorTerm(ctx *cql.InstanceSelectorTermContext) model.IExpression {
	classType := v.VisitNamedTypeSpecifier(ctx.InstanceSelector().NamedTypeSpecifier())
	i := &model.Instance{
		Expression: model.ResultType(classType),
		ClassType:  classType,
	}

	for _, ies := range ctx.InstanceSelector().AllInstanceElementSelector() {
		name := v.parseReferentialIdentifier(ies.ReferentialIdentifier())
		value := v.VisitExpression(ies.Expression())

		// Validate instance element against modelinfo and try to implicitly convert if needed.
		miType, err := v.modelInfo.PropertyTypeSpecifier(classType, name)
		if err != nil {
			return v.badExpression(err.Error(), ctx)
		}

		res, err := convert.OperandImplicitConverter(value.GetResultType(), miType, value, v.modelInfo)
		if err != nil {
			return v.badExpression(err.Error(), ctx)
		}
		if !res.Matched {
			return v.badExpression(fmt.Sprintf("element %q in %v should be implicitly convertible to type %v, but instead received type %v", name, classType, miType, value.GetResultType()), ctx)
		}

		i.Elements = append(i.Elements, &model.InstanceElement{Name: name, Value: res.WrappedOperand})
	}

	return i
}

// TODO: b/324580831 - Possibly add functional definition support for MaxValue and MinValue.
func (v *visitor) VisitTypeExtentExpressionTermContext(ctx *cql.TypeExtentExpressionTermContext) model.IExpression {
	valueType := v.VisitNamedTypeSpecifier(ctx.NamedTypeSpecifier())
	switch valueType {
	case types.Integer,
		types.Long,
		types.Decimal,
		types.Quantity,
		types.Date,
		types.DateTime,
		types.Time:
		// Valid cases.
	default:
		return v.badExpression(fmt.Sprintf("unsupported type for %s expression: %s", ctx.GetText(), valueType.String()), ctx)
	}

	t := ctx.GetText()
	if strings.HasPrefix(t, "maximum") {
		return &model.MaxValue{ValueType: valueType, Expression: model.ResultType(valueType)}
	} else if strings.HasPrefix(t, "minimum") {
		return &model.MinValue{ValueType: valueType, Expression: model.ResultType(valueType)}
	}
	return v.badExpression(fmt.Sprintf("unsupported type extent expression: %v", t), ctx)
}

func (v *visitor) VisitIfThenElseExpression(ctx *cql.IfThenElseExpressionTermContext) model.IExpression {
	// Children are ordered as: if, expr, then, expr, else, expr
	cnd := v.VisitExpression(ctx.Expression(0))
	inferredCnd, err := convert.OperandImplicitConverter(cnd.GetResultType(), types.Boolean, cnd, v.modelInfo)
	if err != nil {
		return v.badExpression(err.Error(), ctx)
	}
	if !inferredCnd.Matched {
		return v.badExpression(fmt.Sprintf("could not implicitly convert %v to a %v", cnd.GetResultType(), types.Boolean), ctx)
	}
	thn := v.VisitExpression(ctx.Expression(1))
	els := v.VisitExpression(ctx.Expression(2))
	i, err := convert.InferMixed([]model.IExpression{thn, els}, v.modelInfo)
	if err != nil {
		return v.badExpression(err.Error(), ctx)
	}

	return &model.IfThenElse{
		Condition:  inferredCnd.WrappedOperand,
		Then:       i.WrappedOperands[0],
		Else:       i.WrappedOperands[1],
		Expression: model.ResultType(i.UniformType),
	}
}

func (v *visitor) VisitSimpleLiteral(ctx *cql.SimpleLiteralContext) model.IExpression {
	switch ctx.GetChild(0).(type) {
	case *cql.SimpleStringLiteralContext:
		return buildLiteral(ctx.GetText(), types.String)
	case *cql.SimpleNumberLiteralContext:
		return buildNumberLiteral(ctx.GetText())
	default:
		return v.badExpression("internal error - grammar should never let us reach this point of VisitSimpleLiteral", ctx)
	}
}

func (v *visitor) VisitLiteralTerm(ctx *cql.LiteralTermContext) model.IExpression {
	val := ctx.GetText()
	switch t := ctx.GetChild(0).(type) {
	case *cql.RatioLiteralContext:
		return v.VisitRatioLiteral(t)
	// Quantities are a special literal term that evaluate to an expression.
	case *cql.QuantityLiteralContext:
		q, err := v.VisitQuantityContext(t.Quantity())
		if err != nil {
			return v.badExpression(err.Error(), ctx)
		}
		return &q
	case *cql.NumberLiteralContext:
		return buildNumberLiteral(val)
	case *cql.StringLiteralContext:
		return buildLiteral(unquoteString(val), types.String)
	case *cql.BooleanLiteralContext:
		return buildLiteral(val, types.Boolean)
	case *cql.LongNumberLiteralContext:
		return buildLiteral(val, types.Long)
	case *cql.NullLiteralContext:
		return buildLiteral(val, types.Any)
	case *cql.DateTimeLiteralContext:
		_, _, err := datehelpers.ParseDateTime(val, time.UTC)
		if err != nil {
			return v.badExpression(err.Error(), ctx)
		}
		return buildLiteral(val, types.DateTime)
	case *cql.DateLiteralContext:
		_, _, err := datehelpers.ParseDate(val, time.UTC)
		if err != nil {
			return v.badExpression(err.Error(), ctx)
		}
		return buildLiteral(val, types.Date)
	case *cql.TimeLiteralContext:
		_, _, err := datehelpers.ParseTime(val, time.UTC)
		if err != nil {
			return v.badExpression(err.Error(), ctx)
		}
		return buildLiteral(val, types.Time)
	default:
		return v.badExpression("internal error - grammar should never let us reach this point of VisitLiteralTerm", ctx)
	}
}

func buildNumberLiteral(val string) *model.Literal {
	t := types.Integer
	if strings.Contains(val, ".") {
		t = types.Decimal
	}
	return buildLiteral(val, t)
}

func buildLiteral(val string, t types.System) *model.Literal {
	return &model.Literal{Value: val, Expression: model.ResultType(t)}
}

func (v *visitor) VisitRatioLiteral(ctx *cql.RatioLiteralContext) model.IExpression {
	quantities := ctx.Ratio().AllQuantity()
	if len(quantities) != 2 {
		return v.badExpression("", ctx)
	}

	numerator, err := v.VisitQuantityContext(quantities[0])
	if err != nil {
		return v.badExpression(fmt.Sprintf("internal error - unable to parse ratio literal, got invalid numerator, %s with error: %s", quantities[0], err.Error()), ctx)
	}
	denominator, err := v.VisitQuantityContext(quantities[1])
	if err != nil {
		return v.badExpression(fmt.Sprintf("internal error - unable to parse ratio literal, got invalid denominator, %s with error: %s", quantities[1], err.Error()), ctx)
	}

	return &model.Ratio{Numerator: numerator, Denominator: denominator, Expression: model.ResultType(types.Ratio)}
}

// VisitQuantityContext handles quantity literals in CQL grammar
func (v *visitor) VisitQuantityContext(ctx cql.IQuantityContext) (model.Quantity, error) {
	numberContext := ctx.NUMBER()
	unitContext := ctx.Unit()

	d, err := strconv.ParseFloat(numberContext.GetText(), 64)
	if err != nil {
		return model.Quantity{}, fmt.Errorf("internal error - unable to parse quantity numeral err: %v, value: %s", err, numberContext.GetText())
	}

	if unitContext == nil {
		return model.Quantity{Value: d, Unit: model.ONEUNIT, Expression: model.ResultType(types.Quantity)}, nil
	}
	if unitContext.DateTimePrecision() != nil {
		rs := unitContext.DateTimePrecision().GetText()
		u := stringToTimeUnit(rs)
		if u == model.UNSETUNIT {
			return model.Quantity{}, fmt.Errorf("internal error - invalid quantity unit when parsing quantity, got: %s", unitContext.GetText())
		}
		return model.Quantity{Value: d, Unit: u, Expression: model.ResultType(types.Quantity)}, nil
	}
	if unitContext.PluralDateTimePrecision() != nil {
		rs := pluralToSingularDateTimePrecision(unitContext.PluralDateTimePrecision().GetText())
		u := stringToTimeUnit(rs)
		if u == model.UNSETUNIT {
			return model.Quantity{}, fmt.Errorf("internal error - invalid quantity unit when parsing quantity, got: %s", unitContext.GetText())
		}
		return model.Quantity{Value: d, Unit: u, Expression: model.ResultType(types.Quantity)}, nil
	}
	
	// Use UCUM validation for unit strings
	unitStr := parseSTRING(unitContext.STRING())
	valid, msg := ucum.CheckUnit(unitStr, true, true)
	if !valid {
		// Just log a warning and continue - don't block parsing
		fmt.Printf("Warning: %s\n", msg)
	}
	return model.Quantity{Value: d, Unit: model.Unit(unitStr), Expression: model.ResultType(types.Quantity)}, nil
}

// VisitUnit handles unit context objects in CQL grammar
func (v *visitor) VisitUnit(ctx cql.IUnitContext) (model.Unit, error) {
	if ctx == nil {
		return model.UNSETUNIT, nil
	}
	
	if ctx.DateTimePrecision() != nil {
		rs := ctx.DateTimePrecision().GetText()
		u := stringToTimeUnit(rs)
		if u == model.UNSETUNIT {
			return model.UNSETUNIT, fmt.Errorf("invalid date time precision unit: %s", rs)
		}
		return u, nil
	}
	
	if ctx.PluralDateTimePrecision() != nil {
		rs := pluralToSingularDateTimePrecision(ctx.PluralDateTimePrecision().GetText())
		u := stringToTimeUnit(rs)
		if u == model.UNSETUNIT {
			return model.UNSETUNIT, fmt.Errorf("invalid plural date time precision unit: %s", rs)
		}
		return u, nil
	}
	
	if ctx.STRING() != nil {
		unitStr := parseSTRING(ctx.STRING())
		// Validate the unit through the UCUM package
		valid, msg := ucum.CheckUnit(unitStr, true, true)
		if !valid {
			// Just log a warning and continue - don't block parsing
			fmt.Printf("Warning: %s\n", msg)
		}
		return model.Unit(unitStr), nil
	}
	
	return model.UNSETUNIT, fmt.Errorf("unsupported unit context type")
}

// VisitReferentialIdentifier handles ReferentialIdentifiers. ReferentialIdentifiers are used
// throughout the CQL grammar as either a reference to a local definition or a definition in another
// CQL library. A reference to another library (ex libraryName.defName) will commonly be represented
// as (referentialIdentifier.referentialIdenfitier) in the CQL grammar. VisitReferentialIdentifier
// only handles local references returning returning model.XXXRef. If this is the local identifier
// of an included library then an error is returned as that should be handled in the calling
// visitor.
func (v *visitor) VisitReferentialIdentifier(ctx cql.IReferentialIdentifierContext) model.IExpression {
	name := v.parseReferentialIdentifier(ctx)

	if v.refs.HasScopedStruct() {
		sourceFn, err := v.refs.ScopedStruct()
		if err != nil {
			return v.badExpression(err.Error(), ctx)
		}

		// If the query source has the expected property, return the identifier ref. Otherwise
		// fall through to the resolution logic below.
		source := sourceFn()
		elementType := source.GetResultType().(*types.List).ElementType

		ptype, err := v.modelInfo.PropertyTypeSpecifier(elementType, name)
		if err == nil {
			return &model.IdentifierRef{
				Name:       name,
				Expression: model.ResultType(ptype),
			}
		}
	}

	if i := v.refs.ResolveInclude(name); i != nil {
		return v.badExpression(fmt.Sprintf("internal error - referential identifier %v is a local identifier to an included library", name), ctx)
	}

	modelFunc, err := v.refs.ResolveLocal(name)
	if err != nil {
		return v.badExpression(err.Error(), ctx)
	}
	return modelFunc()
}

func (v *visitor) parseReferentialIdentifier(ctx cql.IReferentialIdentifierContext) string {
	if ctx.Identifier() != nil {
		return v.VisitIdentifier(ctx.Identifier())
	}
	// Unsure why the CQL grammar splits this into Identifier/KeywordIdentifier.
	return ctx.KeywordIdentifier().GetText()
}

func (v *visitor) parseIdentifierOrFuntionIdentifier(ctx cql.IIdentifierOrFunctionIdentifierContext) string {
	if ctx.Identifier() != nil {
		return v.VisitIdentifier(ctx.Identifier())
	}
	// Unsure why the CQL grammar splits this into Identifier/FunctionIdentifier.
	return ctx.FunctionIdentifier().GetText()
}

// VisitQualifiedIdentifierExpression handles QualifiedIdentifierExpressions, which are references
// to expressions that are defined either locally or in an included CQL library.
func (v *visitor) VisitQualifiedIdentifierExpression(ctx *cql.QualifiedIdentifierExpressionContext) model.IExpression {
	// Parse the series of identifiers.
	ids := []string{}
	for _, q := range ctx.AllQualifierExpression() {
		ids = append(ids, v.parseReferentialIdentifier(q.ReferentialIdentifier()))
	}
	ids = append(ids, v.parseReferentialIdentifier(ctx.ReferentialIdentifier()))

	if len(ids) == 1 {
		// This must be a reference to a local identifier.
		modelFunc, err := v.refs.ResolveLocal(ids[0])
		if err != nil {
			return v.badExpression(err.Error(), ctx)
		}
		return modelFunc()
	}

	var ref model.IExpression
	var i int
	var err error
	lib := v.refs.ResolveInclude(ids[0])
	if lib != nil {
		// This is a reference to a global expression where ids[0] is the included library identifier
		// and ids[1] is the expression definition in the included library.
		i = 2
		ref, err = v.resolveGlobalRef(lib.Local, v.parseReferentialIdentifier(ctx.ReferentialIdentifier()))
		if err != nil {
			return v.badExpression(err.Error(), ctx)
		}
	} else {
		// ids[0] is a reference to a local identifier.
		i = 1
		modelFunc, err := v.refs.ResolveLocal(ids[0])
		if err != nil {
			return v.badExpression(err.Error(), ctx)
		}
		ref = modelFunc()
	}

	// Any remaining identifiers are properties. Repeatedly wrap in model.Property for the remaining
	// identifiers.
	for _, id := range ids[i:] {
		p := &model.Property{
			Source: ref,
			Path:   id,
		}

		if p.Source.GetResultType() != nil {
			propertyType, err := v.modelInfo.PropertyTypeSpecifier(p.Source.GetResultType(), p.Path)
			if err != nil {
				return v.badExpression(err.Error(), ctx)
			}
			p.Expression = model.ResultType(propertyType)
		}
		ref = p
	}
	return ref
}

func (v *visitor) resolveGlobalRef(libName, defName string) (model.IExpression, error) {
	modelFunc, err := v.refs.ResolveGlobal(libName, defName)
	if err != nil {
		return nil, err
	}

	switch typedM := modelFunc().(type) {
	case *model.CodeRef:
		typedM.LibraryName = libName
		return typedM, nil
	case *model.CodeSystemRef:
		typedM.LibraryName = libName
		return typedM, nil
	case *model.ConceptRef:
		typedM.LibraryName = libName
		return typedM, nil
	case *model.ParameterRef:
		typedM.LibraryName = libName
		return typedM, nil
	case *model.ExpressionRef:
		typedM.LibraryName = libName
		return typedM, nil
	case *model.ValuesetRef:
		typedM.LibraryName = libName
		return typedM, nil
	}

	return nil, fmt.Errorf("internal error - global reference %s.%s is not a supported reference type", libName, defName)
}

// VisitQualifiedIdentifier handles QualifiedIdentifiers, which are only used as the full qualified
// library name in library definitions or includes.
func (v *visitor) VisitQualifiedIdentifier(ctx cql.IQualifiedIdentifierContext) []string {
	var ids []string
	for _, c := range ctx.AllQualifier() {
		id := v.VisitIdentifier(c.Identifier())
		ids = append(ids, id)
	}
	ids = append(ids, v.VisitIdentifier(ctx.Identifier()))
	return ids
}

func (v *visitor) VisitIntervalSelectorTerm(ctx *cql.IntervalSelectorTermContext) model.IExpression {
	ictx := ctx.GetChild(0).(*cql.IntervalSelectorContext)
	l := v.VisitExpression(ictx.Expression(0))
	h := v.VisitExpression(ictx.Expression(1))

	// These are the supported interval types, which are based on the overloads for Successor and
	// Predecessor operators. Low and high must be implicitly convertible to one of these types or an
	// error is returned.
	declared := [][]types.IType{
		{types.Integer, types.Integer},
		{types.Long, types.Long},
		{types.Decimal, types.Decimal},
		{types.Quantity, types.Quantity},
		{types.Date, types.Date},
		{types.DateTime, types.DateTime},
		{types.Time, types.Time}}
	overloads := []convert.Overload[func() *model.Interval]{}
	for _, o := range declared {
		overload := convert.Overload[func() *model.Interval]{
			Operands: o,
			Result: func() *model.Interval {
				return &model.Interval{}
			},
		}
		overloads = append(overloads, overload)
	}

	matched, err := convert.OverloadMatch([]model.IExpression{l, h}, overloads, v.modelInfo, "Interval")
	if err != nil {
		return v.badExpression(err.Error(), ctx)
	}

	interval := matched.Result()
	interval.Low = matched.WrappedOperands[0]
	interval.High = matched.WrappedOperands[1]
	interval.Expression = model.ResultType(&types.Interval{PointType: interval.Low.GetResultType()})

	if ictx.GetChild(1).(*antlr.TerminalNodeImpl).GetText() == "[" {
		interval.LowInclusive = true
	}
	if ictx.GetChild(ictx.GetChildCount()-1).(*antlr.TerminalNodeImpl).GetText() == "]" {
		interval.HighInclusive = true
	}

	return interval
}

func (v *visitor) VisitListSelectorTerm(ctx *cql.ListSelectorTermContext) model.IExpression {
	lctx := ctx.GetChild(0).(*cql.ListSelectorContext)

	var listElemType types.IType
	if lctx.TypeSpecifier() != nil {
		listElemType = v.VisitTypeSpecifier(lctx.TypeSpecifier())
	}

	var elemExp []model.IExpression
	for _, exp := range lctx.AllExpression() {
		elemExp = append(elemExp, v.VisitExpression(exp))
	}

	// Empty list
	if len(elemExp) == 0 {
		if lctx.TypeSpecifier() == nil {
			listElemType = types.Any
		}
		return &model.List{List: []model.IExpression{}, Expression: model.ResultType(&types.List{ElementType: listElemType})}
	}

	// There is a List Type Specifier
	if lctx.TypeSpecifier() != nil {
		m := &model.List{Expression: model.ResultType(&types.List{ElementType: listElemType})}
		for _, exp := range elemExp {
			// Attempt to convert each list element to the type specifier type.
			converted, err := convert.OperandImplicitConverter(exp.GetResultType(), listElemType, exp, v.modelInfo)
			if err != nil {
				return v.badExpression(fmt.Sprintf("internal error - converting the list element to the List type specifier: %v", err), lctx)
			}
			if !converted.Matched {
				return v.badExpression(fmt.Sprintf("unable to convert list element (%v) to the declared List type specifier element type (%v)", exp.GetResultType(), listElemType), lctx)
			}
			m.List = append(m.List, converted.WrappedOperand)
		}
		return m
	}

	// No List Type Specifier, try to convert all the elements to a common type, otherwise punt to
	// Choice.
	inferred, err := convert.InferMixed(elemExp, v.modelInfo)
	if err != nil {
		return v.badExpression(fmt.Sprintf("internal error - inferring the uniform type of a list literal: %v", err), lctx)
	}
	return &model.List{
		Expression: model.ResultType(&types.List{ElementType: inferred.UniformType}),
		List:       inferred.WrappedOperands,
	}
}

func (v *visitor) VisitCodeSelectorTerm(ctx *cql.CodeSelectorTermContext) *model.Code {
	c := &model.Code{
		Code:       parseSTRING(ctx.CodeSelector().STRING()),
		System:     v.parseCodeSystemIdentifier(ctx.CodeSelector().CodesystemIdentifier()),
		Expression: model.ResultType(types.Code),
	}
	if ctx.CodeSelector().DisplayClause() != nil {
		c.Display = parseSTRING(ctx.CodeSelector().DisplayClause().STRING())
	}
	return c
}

func (v *visitor) VisitTypeSpecifier(ctx cql.ITypeSpecifierContext) types.IType {
	if ctx == nil {
		return nil
	}

	switch {
	case ctx.NamedTypeSpecifier() != nil:
		return v.VisitNamedTypeSpecifier(ctx.NamedTypeSpecifier())
	case ctx.ListTypeSpecifier() != nil:
		return v.VisitListTypeSpecifier(ctx.ListTypeSpecifier())
	case ctx.IntervalTypeSpecifier() != nil:
		return v.VisitIntervalTypeSpecifier(ctx.IntervalTypeSpecifier())
	case ctx.ChoiceTypeSpecifier() != nil:
		return v.VisitChoiceTypeSpecifier(ctx.ChoiceTypeSpecifier())
	case ctx.TupleTypeSpecifier() != nil:
		return v.VisitTupleTypeSpecifier(ctx.TupleTypeSpecifier())
	default:
		return v.badTypeSpecifier("internal error - grammar should never let us reach this point of VisitTypeSpecifier", ctx)
	}
}

func (v *visitor) VisitNamedTypeSpecifier(ctx cql.INamedTypeSpecifierContext) types.IType {
	// Construct the string type.
	var ids []string
	for _, c := range ctx.AllQualifier() {
		id := v.VisitIdentifier(c.Identifier())
		ids = append(ids, id)
	}
	ref := ctx.ReferentialOrTypeNameIdentifier()
	if ref.TypeNameIdentifier() != nil {
		ids = append(ids, ref.TypeNameIdentifier().GetText())
	} else if ref.ReferentialIdentifier().Identifier() != nil {
		ids = append(ids, v.VisitIdentifier(ref.ReferentialIdentifier().Identifier()))
	} else if ref.ReferentialIdentifier().KeywordIdentifier() != nil {
		ids = append(ids, ref.ReferentialIdentifier().KeywordIdentifier().GetText())
	}

	strType := strings.Join(ids, ".")

	// Convert the string type to an types.IType and validate it is a correct type.
	sys := types.ToSystem(strType)
	if !sys.Equal(types.Unset) {
		return sys
	}

	named, err := v.modelInfo.ToNamed(strType)
	if err != nil {
		return v.badTypeSpecifier(err.Error(), ctx)
	}
	return named
}

func (v *visitor) VisitListTypeSpecifier(ctx cql.IListTypeSpecifierContext) *types.List {
	return &types.List{
		ElementType: v.VisitTypeSpecifier(ctx.TypeSpecifier()),
	}
}

func (v *visitor) VisitIntervalTypeSpecifier(ctx cql.IIntervalTypeSpecifierContext) *types.Interval {
	return &types.Interval{
		PointType: v.VisitTypeSpecifier(ctx.TypeSpecifier()),
	}
}

func (v *visitor) VisitChoiceTypeSpecifier(ctx cql.IChoiceTypeSpecifierContext) *types.Choice {
	c := &types.Choice{ChoiceTypes: []types.IType{}}
	for _, t := range ctx.AllTypeSpecifier() {
		c.ChoiceTypes = append(c.ChoiceTypes, v.VisitTypeSpecifier(t))
	}
	return c
}

func (v *visitor) VisitTupleTypeSpecifier(ctx cql.ITupleTypeSpecifierContext) *types.Tuple {
	t := &types.Tuple{ElementTypes: map[string]types.IType{}}
	for _, elem := range ctx.AllTupleElementDefinition() {
		name := v.parseReferentialIdentifier(elem.ReferentialIdentifier())
		t.ElementTypes[name] = v.VisitTypeSpecifier(elem.TypeSpecifier())
	}
	return t
}

func (v *visitor) VisitVersionSpecifier(ctx cql.IVersionSpecifierContext) string {
	if ctx == nil {
		return ""
	}
	return parseSTRING(ctx.STRING())
}

func (v *visitor) VisitCaseExpressionTerm(ctx *cql.CaseExpressionTermContext) model.IExpression {
	caseModel := &model.Case{}

	for _, ctxCaseItem := range ctx.AllCaseExpressionItem() {
		caseItem := &model.CaseItem{
			When: v.VisitExpression(ctxCaseItem.Expression(0)),
			Then: v.VisitExpression(ctxCaseItem.Expression(1)),
		}
		caseModel.CaseItem = append(caseModel.CaseItem, caseItem)
	}

	expr := ctx.AllExpression()
	var err error
	if len(expr) == 1 {
		// There is no comparand
		caseModel.Else = v.VisitExpression(expr[0])
		caseModel, err = v.booleanWhen(caseModel)
		if err != nil {
			return v.badExpression(err.Error(), ctx)
		}
	} else {
		// There is a comparand
		caseModel.Comparand = v.VisitExpression(expr[0])
		caseModel.Else = v.VisitExpression(expr[1])
		caseModel, err = v.uniformWhen(caseModel)
		if err != nil {
			return v.badExpression(err.Error(), ctx)
		}
	}

	// The CaseItem.Whens are wrapped in necessary conversions. Now we need to wrap CaseItem.Then and
	// Else.
	mixed := make([]model.IExpression, 0, len(caseModel.CaseItem))
	for _, caseItem := range caseModel.CaseItem {
		mixed = append(mixed, caseItem.Then)
	}
	mixed = append(mixed, caseModel.Else)

	uniform, err := convert.InferMixed(mixed, v.modelInfo)
	if err != nil {
		return v.badExpression(err.Error(), ctx)
	}

	for i := 0; i < len(caseModel.CaseItem); i++ {
		caseModel.CaseItem[i].Then = uniform.WrappedOperands[i]
	}
	caseModel.Else = uniform.WrappedOperands[len(uniform.WrappedOperands)-1]
	caseModel.Expression = model.ResultType(uniform.UniformType)

	return caseModel
}

// booleanWhen is for when there is no comparand. Each CaseItem.When must evaluate to a Boolean or
// something implicitly convertible to a Boolean. booleanWhen wraps each CaseItem.When in the
// necessary conversion or returns an error if a CaseItem.When is not convertible.
func (v *visitor) booleanWhen(caseModel *model.Case) (*model.Case, error) {
	for i := 0; i < len(caseModel.CaseItem); i++ {
		res, err := convert.OperandImplicitConverter(caseModel.CaseItem[i].When.GetResultType(), types.Boolean, caseModel.CaseItem[i].When, v.modelInfo)
		if err != nil {
			return nil, err
		}
		if !res.Matched {
			return nil, fmt.Errorf("could not implicitly convert %v to a %v", caseModel.CaseItem[i].When.GetResultType(), types.Boolean)
		}
		caseModel.CaseItem[i].When = res.WrappedOperand
	}

	return caseModel, nil
}

// uniformWhen is for when the comparand will be compared against each CaseItem.When. The comparand
// and all CaseItem.When must be the same type, or implicitly convertible to the same type.
// uniformWhen wraps all CaseItem.When in implicit conversion so all CaseItem.When are a uniform
// type. If there is no uniform type an error is returned.
func (v *visitor) uniformWhen(caseModel *model.Case) (*model.Case, error) {
	mixed := make([]model.IExpression, 0, len(caseModel.CaseItem))
	for _, caseItem := range caseModel.CaseItem {
		mixed = append(mixed, caseItem.When)
	}
	mixed = append(mixed, caseModel.Comparand)

	uniform, err := convert.InferMixed(mixed, v.modelInfo)
	if err != nil {
		return nil, err
	}

	if uniform.PuntedToChoice {
		return nil, fmt.Errorf("could not implicitly convert then comparand %v and cases %v to the same type", caseModel.Comparand.GetResultType(), convert.OperandsToString(mixed[0:len(mixed)-1]))
	}

	for i := 0; i < len(caseModel.CaseItem); i++ {
		caseModel.CaseItem[i].When = uniform.WrappedOperands[i]
	}
	caseModel.Comparand = uniform.WrappedOperands[len(uniform.WrappedOperands)-1]

	return caseModel, nil
}

func (v *visitor) VisitInvocationExpressionTerm(ctx *cql.InvocationExpressionTermContext) model.IExpression {
	// InvocationExpressionTerms can either be
	// 1) Referencing a public definition in an included library. In this case the ExpressionTerm is
	//    the local identifier of the included library and the QualifiedInvocation is the identifier
	//    of the definition.
	// 2) Accessing the property of an expression. In this case the ExpressionTerm is either an
	//    expression or an identifier for an expression.
	expr := ctx.ExpressionTerm()

	// Case 1, the ExpressionTerm is the local identifier of an included library.
	lib := v.isIncludedLibrary(expr)
	if lib != nil {
		switch r := ctx.QualifiedInvocation().GetChild(0).(type) {
		case *cql.ReferentialIdentifierContext:
			m, err := v.resolveGlobalRef(lib.Local, v.parseReferentialIdentifier(r))
			if err != nil {
				return v.badExpression(err.Error(), ctx)
			}
			return m
		case *cql.QualifiedFunctionContext:
			return v.parseQualifiedFunction(r, lib.Local)
		}
	}

	switch r := ctx.QualifiedInvocation().GetChild(0).(type) {
	case *cql.ReferentialIdentifierContext:
		// Case 2, accessing the property of an expression.
		p := &model.Property{
			Source: v.VisitExpression(expr),
			Path:   v.parseReferentialIdentifier(r),
		}

		if p.Source.GetResultType() != nil {
			propertyType, err := v.modelInfo.PropertyTypeSpecifier(p.Source.GetResultType(), p.Path)
			if err != nil {
				return v.badExpression(err.Error(), ctx)
			}
			p.Expression = model.ResultType(propertyType)
		}
		return p
	case *cql.QualifiedFunctionContext:
		// Case 3, fluent functions
		name := v.parseIdentifierOrFuntionIdentifier(r.IdentifierOrFunctionIdentifier())

		// Prepend expr as the first argument to the function call.
		params := []antlr.Tree{expr}
		if r.ParamList() != nil {
			for _, expr := range r.ParamList().AllExpression() {
				params = append(params, expr)
			}
		}
		// Note the grammar does not allow qualified fluent functions (ex
		// Patient.active.FHIRHelpers.ToBoolean()) so you can never call a fluent function in another
		// library.
		m, err := v.parseFunction("", name, params, true)
		if err != nil {
			return v.badExpression(fmt.Errorf("%w (may not be a fluent function)", err).Error(), ctx)
		}

		return m
	}

	return v.badExpression("internal error - grammar should never reach this point of InvocationExpressionTerm", ctx)
}

// isIncludedLibrary checks if the this is the local identifier to an included library, returning
// the model.LibraryIdentifier if it is.
func (v *visitor) isIncludedLibrary(expr cql.IExpressionTermContext) *model.LibraryIdentifier {
	invoc, ok := expr.GetChild(0).(*cql.InvocationTermContext)
	if ok {
		memInvoc, ok := invoc.GetChild(0).(*cql.MemberInvocationContext)
		if ok {
			ref, ok := memInvoc.GetChild(0).(*cql.ReferentialIdentifierContext)
			if ok {
				name := v.parseReferentialIdentifier(ref)
				if i := v.refs.ResolveInclude(name); i != nil {
					return i
				}
			}
		}
	}
	return nil
}

// createRetrieve creates a retrieve struct for the given type populated with modelinfo metadata.
// The caller can then add additional filters to it as needed.
func (v *visitor) createRetrieve(resourceType string) (*model.Retrieve, error) {
	namedType, err := v.modelInfo.ToNamed(resourceType)
	if err != nil {
		return nil, err
	}
	tInfo, err := v.modelInfo.NamedTypeInfo(namedType)
	if err != nil {
		return nil, err
	}
	url, err := v.modelInfo.URL()
	if err != nil {
		return nil, err
	}

	if !tInfo.Retrievable {
		return nil, fmt.Errorf("tried to retrieve type %s, but this type is not retrievable", namedType)
	}
	split := strings.Split(resourceType, ".")
	unqualifiedName := split[len(split)-1]
	r := &model.Retrieve{
		DataType:     fmt.Sprintf("{%v}%v", url, unqualifiedName),
		TemplateID:   tInfo.Identifier,
		CodeProperty: tInfo.PrimaryCodePath,
		Expression:   model.ResultType(&types.List{ElementType: namedType}),
	}

	return r, nil
}

func (v *visitor) VisitRetrieve(ctx *cql.RetrieveContext) model.IExpression {
	typ := v.VisitNamedTypeSpecifier(ctx.NamedTypeSpecifier())
	namedType, ok := typ.(*types.Named)
	if !ok {
		return v.badExpression(fmt.Sprintf("retrieves cannot be performed on type %v", typ), ctx)
	}

	r, err := v.createRetrieve(namedType.TypeName)
	if err != nil {
		return v.badExpression(err.Error(), ctx)
	}

	t := ctx.Terminology()
	if t != nil {
		if t.Expression() != nil {
			r.Codes = v.VisitExpression(t.Expression())
		} else if t.QualifiedIdentifierExpression() != nil {
			r.Codes = v.VisitExpression(t.QualifiedIdentifierExpression())
		}
	}

	return r
}

func (v *visitor) VisitRetrieveExpression(ctx *cql.RetrieveExpressionContext) model.IExpression {
	return v.VisitRetrieve(ctx.Retrieve().(*cql.RetrieveContext))
}

func (v *visitor) VisitTypeExpression(ctx *cql.TypeExpressionContext) model.IExpression {
	// Although Is and As are unary system operators, they do not need to be included in
	// loadSystemOperators(). This is because there is no overload matching, they work for any operand
	// type.
	if ctx.GetChild(1).(antlr.TerminalNode).GetText() == "is" {
		return &model.Is{
			UnaryExpression: &model.UnaryExpression{
				Operand:    v.VisitExpression(ctx.GetChild(0)),
				Expression: model.ResultType(types.Boolean),
			},
			IsTypeSpecifier: v.VisitTypeSpecifier(ctx.TypeSpecifier()),
		}
	}

	return &model.As{
		UnaryExpression: &model.UnaryExpression{
			Operand:    v.VisitExpression(ctx.Expression()),
			Expression: model.ResultType(v.VisitTypeSpecifier(ctx.TypeSpecifier())),
		},
		AsTypeSpecifier: v.VisitTypeSpecifier(ctx.TypeSpecifier()),
		Strict:          false,
	}
}

func (v *visitor) VisitCastExpression(ctx *cql.CastExpressionContext) model.IExpression {
	// Although As is a unary system operator, it does not need to be included in
	// loadSystemOperators(). This is because there is no overload matching, it works for any operand
	// type.
	asType := v.VisitTypeSpecifier(ctx.TypeSpecifier())
	return &model.As{
		UnaryExpression: &model.UnaryExpression{
			Operand:    v.VisitExpression(ctx.Expression()),
			Expression: model.ResultType(asType),
		},
		AsTypeSpecifier: asType,
		Strict:          true,
	}
}

func (v *visitor) VisitIdentifier(ctx cql.IIdentifierContext) string {
	if ctx.IDENTIFIER() != nil {
		return ctx.IDENTIFIER().GetText()
	}
	if ctx.QUOTEDIDENTIFIER() != nil || ctx.DELIMITEDIDENTIFIER() != nil {
		return unquoteString(ctx.GetText())
	}

	v.reportError("Invalid identifier", ctx)

	return ctx.GetText()
}

// maybeGetChildNode returns the first child of type T within children, otherwise returns the
// default return value and false. defaultReturn should typically be nil if T is a pointer, and
// exists to prevent an allocation inside maybeGetChildNode.
func maybeGetChildNode[T antlr.Tree](children []antlr.Tree, defaultReturn T) (T, bool) {
	for _, c := range children {
		if node, ok := c.(T); ok {
			return node, true
		}
	}
	return defaultReturn, false
}

// PluralToSingularDateTimePrecision converts a plural time precision (years) to singular (year).
func pluralToSingularDateTimePrecision(pluralPrecision string) string {
	return strings.TrimSuffix(pluralPrecision, "s")
}

// precisionFromContext returns the string precision from the given context if available.
// Currently only accepts a cql.DateTimePrecisionSpecifierContext or cql.DateTimePrecisionContext.
// Otherwise returns `UNSETDATETIMEPRECISION` (an empty string).
func precisionFromContext(ctx antlr.ParserRuleContext) model.DateTimePrecision {
	if n, ok := maybeGetChildNode[*cql.DateTimePrecisionSpecifierContext](ctx.GetChildren(), nil); ok {
		return stringToPrecision(n.GetChild(0).(*cql.DateTimePrecisionContext).GetText())
	} else if n, ok := maybeGetChildNode[*cql.DateTimePrecisionContext](ctx.GetChildren(), nil); ok {
		return stringToPrecision(n.GetText())
	}
	return model.UNSETDATETIMEPRECISION
}

func stringToPrecision(s string) model.DateTimePrecision {
	switch s {
	case "year":
		return model.YEAR
	case "month":
		return model.MONTH
	case "week":
		return model.WEEK
	case "day":
		return model.DAY
	case "hour":
		return model.HOUR
	case "minute":
		return model.MINUTE
	case "second":
		return model.SECOND
	case "millisecond":
		return model.MILLISECOND
	}
	return model.UNSETDATETIMEPRECISION
}

func dateTimePrecisions() []model.DateTimePrecision {
	return []model.DateTimePrecision{
		model.YEAR,
		model.MONTH,
		model.WEEK,
		model.DAY,
		model.HOUR,
		model.MINUTE,
		model.SECOND,
		model.MILLISECOND,
	}
}

// funcNameWithPrecision converts a model.DateTimePrecision to a string used in a function name ex
// AfterYears.
func funcNameWithPrecision(name string, p model.DateTimePrecision) string {
	pStr := ""
	switch p {
	case model.YEAR:
		pStr = "Years"
	case model.MONTH:
		pStr = "Months"
	case model.WEEK:
		pStr = "Weeks"
	case model.DAY:
		pStr = "Days"
	case model.HOUR:
		pStr = "Hours"
	case model.MINUTE:
		pStr = "Minutes"
	case model.SECOND:
		pStr = "Seconds"
	case model.MILLISECOND:
		pStr = "Milliseconds"
	}
	return fmt.Sprintf("%s%s", name, pStr)
}

// stringToTimeUnit converts a string to a model.Unit for temporal values.
// TODO(b/319326228): move common date/datetime logic into a temporal package.
func stringToTimeUnit(s string) model.Unit {
	switch s {
	case "year":
		return model.YEARUNIT
	case "month":
		return model.MONTHUNIT
	case "week":
		return model.WEEKUNIT
	case "day":
		return model.DAYUNIT
	case "hour":
		return model.HOURUNIT
	case "minute":
		return model.MINUTEUNIT
	case "second":
		return model.SECONDUNIT
	case "millisecond":
		return model.MILLISECONDUNIT
	}
	return model.UNSETUNIT
}

// unquoteString takes the given CQL string, removes the surrounding ' and unescapes it.
// Escaped to character mapping: https://cql.hl7.org/03-developersguide.html#literals.
// TODO(b/302003569): properly unescaping unicode characters is not yet supported
func unquoteString(s string) string {
	s = s[1 : len(s)-1]
	for i := 0; i < len(s)-1; i++ {
		quoted := s[i : i+2]
		var replace string
		switch quoted {
		case `\'`:
			replace = `'`
		case `\"`:
			replace = `"`
		case "\\`":
			replace = "`"
		case `\r`:
			replace = "\r"
		case `\n`:
			replace = "\n"
		case `\t`:
			replace = "\t"
		case `\f`:
			replace = "\f"
		case `\\`:
			replace = `\`
		}
		if replace != "" {
			s = s[0:i] + replace + s[i+2:]
		}
	}
	return s
}
