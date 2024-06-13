// Code generated from Cvlt.g4 by ANTLR 4.13.1. DO NOT EDIT.

package cvlt // Cvlt
import "github.com/antlr4-go/antlr/v4"

// A complete Visitor for a parse tree produced by CvltParser.
type CvltVisitor interface {
	antlr.ParseTreeVisitor

	// Visit a parse tree produced by CvltParser#typeSpecifier.
	VisitTypeSpecifier(ctx *TypeSpecifierContext) interface{}

	// Visit a parse tree produced by CvltParser#namedTypeSpecifier.
	VisitNamedTypeSpecifier(ctx *NamedTypeSpecifierContext) interface{}

	// Visit a parse tree produced by CvltParser#modelIdentifier.
	VisitModelIdentifier(ctx *ModelIdentifierContext) interface{}

	// Visit a parse tree produced by CvltParser#listTypeSpecifier.
	VisitListTypeSpecifier(ctx *ListTypeSpecifierContext) interface{}

	// Visit a parse tree produced by CvltParser#intervalTypeSpecifier.
	VisitIntervalTypeSpecifier(ctx *IntervalTypeSpecifierContext) interface{}

	// Visit a parse tree produced by CvltParser#tupleTypeSpecifier.
	VisitTupleTypeSpecifier(ctx *TupleTypeSpecifierContext) interface{}

	// Visit a parse tree produced by CvltParser#tupleElementDefinition.
	VisitTupleElementDefinition(ctx *TupleElementDefinitionContext) interface{}

	// Visit a parse tree produced by CvltParser#choiceTypeSpecifier.
	VisitChoiceTypeSpecifier(ctx *ChoiceTypeSpecifierContext) interface{}

	// Visit a parse tree produced by CvltParser#literalTerm.
	VisitLiteralTerm(ctx *LiteralTermContext) interface{}

	// Visit a parse tree produced by CvltParser#intervalSelectorTerm.
	VisitIntervalSelectorTerm(ctx *IntervalSelectorTermContext) interface{}

	// Visit a parse tree produced by CvltParser#tupleSelectorTerm.
	VisitTupleSelectorTerm(ctx *TupleSelectorTermContext) interface{}

	// Visit a parse tree produced by CvltParser#instanceSelectorTerm.
	VisitInstanceSelectorTerm(ctx *InstanceSelectorTermContext) interface{}

	// Visit a parse tree produced by CvltParser#listSelectorTerm.
	VisitListSelectorTerm(ctx *ListSelectorTermContext) interface{}

	// Visit a parse tree produced by CvltParser#codeSelectorTerm.
	VisitCodeSelectorTerm(ctx *CodeSelectorTermContext) interface{}

	// Visit a parse tree produced by CvltParser#conceptSelectorTerm.
	VisitConceptSelectorTerm(ctx *ConceptSelectorTermContext) interface{}

	// Visit a parse tree produced by CvltParser#ratio.
	VisitRatio(ctx *RatioContext) interface{}

	// Visit a parse tree produced by CvltParser#booleanLiteral.
	VisitBooleanLiteral(ctx *BooleanLiteralContext) interface{}

	// Visit a parse tree produced by CvltParser#nullLiteral.
	VisitNullLiteral(ctx *NullLiteralContext) interface{}

	// Visit a parse tree produced by CvltParser#stringLiteral.
	VisitStringLiteral(ctx *StringLiteralContext) interface{}

	// Visit a parse tree produced by CvltParser#numberLiteral.
	VisitNumberLiteral(ctx *NumberLiteralContext) interface{}

	// Visit a parse tree produced by CvltParser#longNumberLiteral.
	VisitLongNumberLiteral(ctx *LongNumberLiteralContext) interface{}

	// Visit a parse tree produced by CvltParser#dateTimeLiteral.
	VisitDateTimeLiteral(ctx *DateTimeLiteralContext) interface{}

	// Visit a parse tree produced by CvltParser#dateLiteral.
	VisitDateLiteral(ctx *DateLiteralContext) interface{}

	// Visit a parse tree produced by CvltParser#timeLiteral.
	VisitTimeLiteral(ctx *TimeLiteralContext) interface{}

	// Visit a parse tree produced by CvltParser#quantityLiteral.
	VisitQuantityLiteral(ctx *QuantityLiteralContext) interface{}

	// Visit a parse tree produced by CvltParser#ratioLiteral.
	VisitRatioLiteral(ctx *RatioLiteralContext) interface{}

	// Visit a parse tree produced by CvltParser#intervalSelector.
	VisitIntervalSelector(ctx *IntervalSelectorContext) interface{}

	// Visit a parse tree produced by CvltParser#tupleSelector.
	VisitTupleSelector(ctx *TupleSelectorContext) interface{}

	// Visit a parse tree produced by CvltParser#tupleElementSelector.
	VisitTupleElementSelector(ctx *TupleElementSelectorContext) interface{}

	// Visit a parse tree produced by CvltParser#instanceSelector.
	VisitInstanceSelector(ctx *InstanceSelectorContext) interface{}

	// Visit a parse tree produced by CvltParser#instanceElementSelector.
	VisitInstanceElementSelector(ctx *InstanceElementSelectorContext) interface{}

	// Visit a parse tree produced by CvltParser#listSelector.
	VisitListSelector(ctx *ListSelectorContext) interface{}

	// Visit a parse tree produced by CvltParser#displayClause.
	VisitDisplayClause(ctx *DisplayClauseContext) interface{}

	// Visit a parse tree produced by CvltParser#codeSelector.
	VisitCodeSelector(ctx *CodeSelectorContext) interface{}

	// Visit a parse tree produced by CvltParser#conceptSelector.
	VisitConceptSelector(ctx *ConceptSelectorContext) interface{}

	// Visit a parse tree produced by CvltParser#identifier.
	VisitIdentifier(ctx *IdentifierContext) interface{}

	// Visit a parse tree produced by CvltParser#quantity.
	VisitQuantity(ctx *QuantityContext) interface{}

	// Visit a parse tree produced by CvltParser#unit.
	VisitUnit(ctx *UnitContext) interface{}

	// Visit a parse tree produced by CvltParser#dateTimePrecision.
	VisitDateTimePrecision(ctx *DateTimePrecisionContext) interface{}

	// Visit a parse tree produced by CvltParser#pluralDateTimePrecision.
	VisitPluralDateTimePrecision(ctx *PluralDateTimePrecisionContext) interface{}
}
