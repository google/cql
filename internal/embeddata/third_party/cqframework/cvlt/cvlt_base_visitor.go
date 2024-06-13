// Code generated from Cvlt.g4 by ANTLR 4.13.1. DO NOT EDIT.

package cvlt // Cvlt
import "github.com/antlr4-go/antlr/v4"

type BaseCvltVisitor struct {
	*antlr.BaseParseTreeVisitor
}

func (v *BaseCvltVisitor) VisitTypeSpecifier(ctx *TypeSpecifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCvltVisitor) VisitNamedTypeSpecifier(ctx *NamedTypeSpecifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCvltVisitor) VisitModelIdentifier(ctx *ModelIdentifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCvltVisitor) VisitListTypeSpecifier(ctx *ListTypeSpecifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCvltVisitor) VisitIntervalTypeSpecifier(ctx *IntervalTypeSpecifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCvltVisitor) VisitTupleTypeSpecifier(ctx *TupleTypeSpecifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCvltVisitor) VisitTupleElementDefinition(ctx *TupleElementDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCvltVisitor) VisitChoiceTypeSpecifier(ctx *ChoiceTypeSpecifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCvltVisitor) VisitLiteralTerm(ctx *LiteralTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCvltVisitor) VisitIntervalSelectorTerm(ctx *IntervalSelectorTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCvltVisitor) VisitTupleSelectorTerm(ctx *TupleSelectorTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCvltVisitor) VisitInstanceSelectorTerm(ctx *InstanceSelectorTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCvltVisitor) VisitListSelectorTerm(ctx *ListSelectorTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCvltVisitor) VisitCodeSelectorTerm(ctx *CodeSelectorTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCvltVisitor) VisitConceptSelectorTerm(ctx *ConceptSelectorTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCvltVisitor) VisitRatio(ctx *RatioContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCvltVisitor) VisitBooleanLiteral(ctx *BooleanLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCvltVisitor) VisitNullLiteral(ctx *NullLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCvltVisitor) VisitStringLiteral(ctx *StringLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCvltVisitor) VisitNumberLiteral(ctx *NumberLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCvltVisitor) VisitLongNumberLiteral(ctx *LongNumberLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCvltVisitor) VisitDateTimeLiteral(ctx *DateTimeLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCvltVisitor) VisitDateLiteral(ctx *DateLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCvltVisitor) VisitTimeLiteral(ctx *TimeLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCvltVisitor) VisitQuantityLiteral(ctx *QuantityLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCvltVisitor) VisitRatioLiteral(ctx *RatioLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCvltVisitor) VisitIntervalSelector(ctx *IntervalSelectorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCvltVisitor) VisitTupleSelector(ctx *TupleSelectorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCvltVisitor) VisitTupleElementSelector(ctx *TupleElementSelectorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCvltVisitor) VisitInstanceSelector(ctx *InstanceSelectorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCvltVisitor) VisitInstanceElementSelector(ctx *InstanceElementSelectorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCvltVisitor) VisitListSelector(ctx *ListSelectorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCvltVisitor) VisitDisplayClause(ctx *DisplayClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCvltVisitor) VisitCodeSelector(ctx *CodeSelectorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCvltVisitor) VisitConceptSelector(ctx *ConceptSelectorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCvltVisitor) VisitIdentifier(ctx *IdentifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCvltVisitor) VisitQuantity(ctx *QuantityContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCvltVisitor) VisitUnit(ctx *UnitContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCvltVisitor) VisitDateTimePrecision(ctx *DateTimePrecisionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCvltVisitor) VisitPluralDateTimePrecision(ctx *PluralDateTimePrecisionContext) interface{} {
	return v.VisitChildren(ctx)
}
