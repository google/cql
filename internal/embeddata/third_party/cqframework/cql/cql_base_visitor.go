// Code generated from Cql.g4 by ANTLR 4.13.1. DO NOT EDIT.

package cql // Cql
import "github.com/antlr4-go/antlr/v4"


type BaseCqlVisitor struct {
	*antlr.BaseParseTreeVisitor
}

func (v *BaseCqlVisitor) VisitDefinition(ctx *DefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitLibrary(ctx *LibraryContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitLibraryDefinition(ctx *LibraryDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitUsingDefinition(ctx *UsingDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitIncludeDefinition(ctx *IncludeDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitLocalIdentifier(ctx *LocalIdentifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitAccessModifier(ctx *AccessModifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitParameterDefinition(ctx *ParameterDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitCodesystemDefinition(ctx *CodesystemDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitValuesetDefinition(ctx *ValuesetDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitCodesystems(ctx *CodesystemsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitCodesystemIdentifier(ctx *CodesystemIdentifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitLibraryIdentifier(ctx *LibraryIdentifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitCodeDefinition(ctx *CodeDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitConceptDefinition(ctx *ConceptDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitCodeIdentifier(ctx *CodeIdentifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitCodesystemId(ctx *CodesystemIdContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitValuesetId(ctx *ValuesetIdContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitVersionSpecifier(ctx *VersionSpecifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitCodeId(ctx *CodeIdContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitTypeSpecifier(ctx *TypeSpecifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitNamedTypeSpecifier(ctx *NamedTypeSpecifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitModelIdentifier(ctx *ModelIdentifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitListTypeSpecifier(ctx *ListTypeSpecifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitIntervalTypeSpecifier(ctx *IntervalTypeSpecifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitTupleTypeSpecifier(ctx *TupleTypeSpecifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitTupleElementDefinition(ctx *TupleElementDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitChoiceTypeSpecifier(ctx *ChoiceTypeSpecifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitStatement(ctx *StatementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitExpressionDefinition(ctx *ExpressionDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitContextDefinition(ctx *ContextDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitFluentModifier(ctx *FluentModifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitFunctionDefinition(ctx *FunctionDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitOperandDefinition(ctx *OperandDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitFunctionBody(ctx *FunctionBodyContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitQuerySource(ctx *QuerySourceContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitAliasedQuerySource(ctx *AliasedQuerySourceContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitAlias(ctx *AliasContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitQueryInclusionClause(ctx *QueryInclusionClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitWithClause(ctx *WithClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitWithoutClause(ctx *WithoutClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitRetrieve(ctx *RetrieveContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitContextIdentifier(ctx *ContextIdentifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitCodePath(ctx *CodePathContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitCodeComparator(ctx *CodeComparatorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitTerminology(ctx *TerminologyContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitQualifier(ctx *QualifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitQuery(ctx *QueryContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitSourceClause(ctx *SourceClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitLetClause(ctx *LetClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitLetClauseItem(ctx *LetClauseItemContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitWhereClause(ctx *WhereClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitReturnClause(ctx *ReturnClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitAggregateClause(ctx *AggregateClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitStartingClause(ctx *StartingClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitSortClause(ctx *SortClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitSortDirection(ctx *SortDirectionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitSortByItem(ctx *SortByItemContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitQualifiedIdentifier(ctx *QualifiedIdentifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitQualifiedIdentifierExpression(ctx *QualifiedIdentifierExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitQualifierExpression(ctx *QualifierExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitSimplePathIndexer(ctx *SimplePathIndexerContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitSimplePathQualifiedIdentifier(ctx *SimplePathQualifiedIdentifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitSimplePathReferentialIdentifier(ctx *SimplePathReferentialIdentifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitSimpleStringLiteral(ctx *SimpleStringLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitSimpleNumberLiteral(ctx *SimpleNumberLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitDurationBetweenExpression(ctx *DurationBetweenExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitInFixSetExpression(ctx *InFixSetExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitRetrieveExpression(ctx *RetrieveExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitTimingExpression(ctx *TimingExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitQueryExpression(ctx *QueryExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitNotExpression(ctx *NotExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitBooleanExpression(ctx *BooleanExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitOrExpression(ctx *OrExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitCastExpression(ctx *CastExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitAndExpression(ctx *AndExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitBetweenExpression(ctx *BetweenExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitMembershipExpression(ctx *MembershipExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitDifferenceBetweenExpression(ctx *DifferenceBetweenExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitInequalityExpression(ctx *InequalityExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitEqualityExpression(ctx *EqualityExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitExistenceExpression(ctx *ExistenceExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitImpliesExpression(ctx *ImpliesExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitTermExpression(ctx *TermExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitTypeExpression(ctx *TypeExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitDateTimePrecision(ctx *DateTimePrecisionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitDateTimeComponent(ctx *DateTimeComponentContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitPluralDateTimePrecision(ctx *PluralDateTimePrecisionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitAdditionExpressionTerm(ctx *AdditionExpressionTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitIndexedExpressionTerm(ctx *IndexedExpressionTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitWidthExpressionTerm(ctx *WidthExpressionTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitSetAggregateExpressionTerm(ctx *SetAggregateExpressionTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitTimeUnitExpressionTerm(ctx *TimeUnitExpressionTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitIfThenElseExpressionTerm(ctx *IfThenElseExpressionTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitTimeBoundaryExpressionTerm(ctx *TimeBoundaryExpressionTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitElementExtractorExpressionTerm(ctx *ElementExtractorExpressionTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitConversionExpressionTerm(ctx *ConversionExpressionTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitTypeExtentExpressionTerm(ctx *TypeExtentExpressionTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitPredecessorExpressionTerm(ctx *PredecessorExpressionTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitPointExtractorExpressionTerm(ctx *PointExtractorExpressionTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitMultiplicationExpressionTerm(ctx *MultiplicationExpressionTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitAggregateExpressionTerm(ctx *AggregateExpressionTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitDurationExpressionTerm(ctx *DurationExpressionTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitDifferenceExpressionTerm(ctx *DifferenceExpressionTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitCaseExpressionTerm(ctx *CaseExpressionTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitPowerExpressionTerm(ctx *PowerExpressionTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitSuccessorExpressionTerm(ctx *SuccessorExpressionTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitPolarityExpressionTerm(ctx *PolarityExpressionTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitTermExpressionTerm(ctx *TermExpressionTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitInvocationExpressionTerm(ctx *InvocationExpressionTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitCaseExpressionItem(ctx *CaseExpressionItemContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitDateTimePrecisionSpecifier(ctx *DateTimePrecisionSpecifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitRelativeQualifier(ctx *RelativeQualifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitOffsetRelativeQualifier(ctx *OffsetRelativeQualifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitExclusiveRelativeQualifier(ctx *ExclusiveRelativeQualifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitQuantityOffset(ctx *QuantityOffsetContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitTemporalRelationship(ctx *TemporalRelationshipContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitConcurrentWithIntervalOperatorPhrase(ctx *ConcurrentWithIntervalOperatorPhraseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitIncludesIntervalOperatorPhrase(ctx *IncludesIntervalOperatorPhraseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitIncludedInIntervalOperatorPhrase(ctx *IncludedInIntervalOperatorPhraseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitBeforeOrAfterIntervalOperatorPhrase(ctx *BeforeOrAfterIntervalOperatorPhraseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitWithinIntervalOperatorPhrase(ctx *WithinIntervalOperatorPhraseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitMeetsIntervalOperatorPhrase(ctx *MeetsIntervalOperatorPhraseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitOverlapsIntervalOperatorPhrase(ctx *OverlapsIntervalOperatorPhraseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitStartsIntervalOperatorPhrase(ctx *StartsIntervalOperatorPhraseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitEndsIntervalOperatorPhrase(ctx *EndsIntervalOperatorPhraseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitInvocationTerm(ctx *InvocationTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitLiteralTerm(ctx *LiteralTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitExternalConstantTerm(ctx *ExternalConstantTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitIntervalSelectorTerm(ctx *IntervalSelectorTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitTupleSelectorTerm(ctx *TupleSelectorTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitInstanceSelectorTerm(ctx *InstanceSelectorTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitListSelectorTerm(ctx *ListSelectorTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitCodeSelectorTerm(ctx *CodeSelectorTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitConceptSelectorTerm(ctx *ConceptSelectorTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitParenthesizedTerm(ctx *ParenthesizedTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitQualifiedMemberInvocation(ctx *QualifiedMemberInvocationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitQualifiedFunctionInvocation(ctx *QualifiedFunctionInvocationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitQualifiedFunction(ctx *QualifiedFunctionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitMemberInvocation(ctx *MemberInvocationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitFunctionInvocation(ctx *FunctionInvocationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitThisInvocation(ctx *ThisInvocationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitIndexInvocation(ctx *IndexInvocationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitTotalInvocation(ctx *TotalInvocationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitFunction(ctx *FunctionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitRatio(ctx *RatioContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitBooleanLiteral(ctx *BooleanLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitNullLiteral(ctx *NullLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitStringLiteral(ctx *StringLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitNumberLiteral(ctx *NumberLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitLongNumberLiteral(ctx *LongNumberLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitDateTimeLiteral(ctx *DateTimeLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitDateLiteral(ctx *DateLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitTimeLiteral(ctx *TimeLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitQuantityLiteral(ctx *QuantityLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitRatioLiteral(ctx *RatioLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitIntervalSelector(ctx *IntervalSelectorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitTupleSelector(ctx *TupleSelectorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitTupleElementSelector(ctx *TupleElementSelectorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitInstanceSelector(ctx *InstanceSelectorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitInstanceElementSelector(ctx *InstanceElementSelectorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitListSelector(ctx *ListSelectorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitDisplayClause(ctx *DisplayClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitCodeSelector(ctx *CodeSelectorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitConceptSelector(ctx *ConceptSelectorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitKeyword(ctx *KeywordContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitReservedWord(ctx *ReservedWordContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitKeywordIdentifier(ctx *KeywordIdentifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitObsoleteIdentifier(ctx *ObsoleteIdentifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitFunctionIdentifier(ctx *FunctionIdentifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitTypeNameIdentifier(ctx *TypeNameIdentifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitReferentialIdentifier(ctx *ReferentialIdentifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitReferentialOrTypeNameIdentifier(ctx *ReferentialOrTypeNameIdentifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitIdentifierOrFunctionIdentifier(ctx *IdentifierOrFunctionIdentifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitIdentifier(ctx *IdentifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitExternalConstant(ctx *ExternalConstantContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitParamList(ctx *ParamListContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitQuantity(ctx *QuantityContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlVisitor) VisitUnit(ctx *UnitContext) interface{} {
	return v.VisitChildren(ctx)
}
