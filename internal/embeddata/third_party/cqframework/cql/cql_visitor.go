// Code generated from Cql.g4 by ANTLR 4.13.1. DO NOT EDIT.

package cql // Cql
import "github.com/antlr4-go/antlr/v4"


// A complete Visitor for a parse tree produced by CqlParser.
type CqlVisitor interface {
	antlr.ParseTreeVisitor

	// Visit a parse tree produced by CqlParser#definition.
	VisitDefinition(ctx *DefinitionContext) interface{}

	// Visit a parse tree produced by CqlParser#library.
	VisitLibrary(ctx *LibraryContext) interface{}

	// Visit a parse tree produced by CqlParser#libraryDefinition.
	VisitLibraryDefinition(ctx *LibraryDefinitionContext) interface{}

	// Visit a parse tree produced by CqlParser#usingDefinition.
	VisitUsingDefinition(ctx *UsingDefinitionContext) interface{}

	// Visit a parse tree produced by CqlParser#includeDefinition.
	VisitIncludeDefinition(ctx *IncludeDefinitionContext) interface{}

	// Visit a parse tree produced by CqlParser#localIdentifier.
	VisitLocalIdentifier(ctx *LocalIdentifierContext) interface{}

	// Visit a parse tree produced by CqlParser#accessModifier.
	VisitAccessModifier(ctx *AccessModifierContext) interface{}

	// Visit a parse tree produced by CqlParser#parameterDefinition.
	VisitParameterDefinition(ctx *ParameterDefinitionContext) interface{}

	// Visit a parse tree produced by CqlParser#codesystemDefinition.
	VisitCodesystemDefinition(ctx *CodesystemDefinitionContext) interface{}

	// Visit a parse tree produced by CqlParser#valuesetDefinition.
	VisitValuesetDefinition(ctx *ValuesetDefinitionContext) interface{}

	// Visit a parse tree produced by CqlParser#codesystems.
	VisitCodesystems(ctx *CodesystemsContext) interface{}

	// Visit a parse tree produced by CqlParser#codesystemIdentifier.
	VisitCodesystemIdentifier(ctx *CodesystemIdentifierContext) interface{}

	// Visit a parse tree produced by CqlParser#libraryIdentifier.
	VisitLibraryIdentifier(ctx *LibraryIdentifierContext) interface{}

	// Visit a parse tree produced by CqlParser#codeDefinition.
	VisitCodeDefinition(ctx *CodeDefinitionContext) interface{}

	// Visit a parse tree produced by CqlParser#conceptDefinition.
	VisitConceptDefinition(ctx *ConceptDefinitionContext) interface{}

	// Visit a parse tree produced by CqlParser#codeIdentifier.
	VisitCodeIdentifier(ctx *CodeIdentifierContext) interface{}

	// Visit a parse tree produced by CqlParser#codesystemId.
	VisitCodesystemId(ctx *CodesystemIdContext) interface{}

	// Visit a parse tree produced by CqlParser#valuesetId.
	VisitValuesetId(ctx *ValuesetIdContext) interface{}

	// Visit a parse tree produced by CqlParser#versionSpecifier.
	VisitVersionSpecifier(ctx *VersionSpecifierContext) interface{}

	// Visit a parse tree produced by CqlParser#codeId.
	VisitCodeId(ctx *CodeIdContext) interface{}

	// Visit a parse tree produced by CqlParser#typeSpecifier.
	VisitTypeSpecifier(ctx *TypeSpecifierContext) interface{}

	// Visit a parse tree produced by CqlParser#namedTypeSpecifier.
	VisitNamedTypeSpecifier(ctx *NamedTypeSpecifierContext) interface{}

	// Visit a parse tree produced by CqlParser#modelIdentifier.
	VisitModelIdentifier(ctx *ModelIdentifierContext) interface{}

	// Visit a parse tree produced by CqlParser#listTypeSpecifier.
	VisitListTypeSpecifier(ctx *ListTypeSpecifierContext) interface{}

	// Visit a parse tree produced by CqlParser#intervalTypeSpecifier.
	VisitIntervalTypeSpecifier(ctx *IntervalTypeSpecifierContext) interface{}

	// Visit a parse tree produced by CqlParser#tupleTypeSpecifier.
	VisitTupleTypeSpecifier(ctx *TupleTypeSpecifierContext) interface{}

	// Visit a parse tree produced by CqlParser#tupleElementDefinition.
	VisitTupleElementDefinition(ctx *TupleElementDefinitionContext) interface{}

	// Visit a parse tree produced by CqlParser#choiceTypeSpecifier.
	VisitChoiceTypeSpecifier(ctx *ChoiceTypeSpecifierContext) interface{}

	// Visit a parse tree produced by CqlParser#statement.
	VisitStatement(ctx *StatementContext) interface{}

	// Visit a parse tree produced by CqlParser#expressionDefinition.
	VisitExpressionDefinition(ctx *ExpressionDefinitionContext) interface{}

	// Visit a parse tree produced by CqlParser#contextDefinition.
	VisitContextDefinition(ctx *ContextDefinitionContext) interface{}

	// Visit a parse tree produced by CqlParser#fluentModifier.
	VisitFluentModifier(ctx *FluentModifierContext) interface{}

	// Visit a parse tree produced by CqlParser#functionDefinition.
	VisitFunctionDefinition(ctx *FunctionDefinitionContext) interface{}

	// Visit a parse tree produced by CqlParser#operandDefinition.
	VisitOperandDefinition(ctx *OperandDefinitionContext) interface{}

	// Visit a parse tree produced by CqlParser#functionBody.
	VisitFunctionBody(ctx *FunctionBodyContext) interface{}

	// Visit a parse tree produced by CqlParser#querySource.
	VisitQuerySource(ctx *QuerySourceContext) interface{}

	// Visit a parse tree produced by CqlParser#aliasedQuerySource.
	VisitAliasedQuerySource(ctx *AliasedQuerySourceContext) interface{}

	// Visit a parse tree produced by CqlParser#alias.
	VisitAlias(ctx *AliasContext) interface{}

	// Visit a parse tree produced by CqlParser#queryInclusionClause.
	VisitQueryInclusionClause(ctx *QueryInclusionClauseContext) interface{}

	// Visit a parse tree produced by CqlParser#withClause.
	VisitWithClause(ctx *WithClauseContext) interface{}

	// Visit a parse tree produced by CqlParser#withoutClause.
	VisitWithoutClause(ctx *WithoutClauseContext) interface{}

	// Visit a parse tree produced by CqlParser#retrieve.
	VisitRetrieve(ctx *RetrieveContext) interface{}

	// Visit a parse tree produced by CqlParser#contextIdentifier.
	VisitContextIdentifier(ctx *ContextIdentifierContext) interface{}

	// Visit a parse tree produced by CqlParser#codePath.
	VisitCodePath(ctx *CodePathContext) interface{}

	// Visit a parse tree produced by CqlParser#codeComparator.
	VisitCodeComparator(ctx *CodeComparatorContext) interface{}

	// Visit a parse tree produced by CqlParser#terminology.
	VisitTerminology(ctx *TerminologyContext) interface{}

	// Visit a parse tree produced by CqlParser#qualifier.
	VisitQualifier(ctx *QualifierContext) interface{}

	// Visit a parse tree produced by CqlParser#query.
	VisitQuery(ctx *QueryContext) interface{}

	// Visit a parse tree produced by CqlParser#sourceClause.
	VisitSourceClause(ctx *SourceClauseContext) interface{}

	// Visit a parse tree produced by CqlParser#letClause.
	VisitLetClause(ctx *LetClauseContext) interface{}

	// Visit a parse tree produced by CqlParser#letClauseItem.
	VisitLetClauseItem(ctx *LetClauseItemContext) interface{}

	// Visit a parse tree produced by CqlParser#whereClause.
	VisitWhereClause(ctx *WhereClauseContext) interface{}

	// Visit a parse tree produced by CqlParser#returnClause.
	VisitReturnClause(ctx *ReturnClauseContext) interface{}

	// Visit a parse tree produced by CqlParser#aggregateClause.
	VisitAggregateClause(ctx *AggregateClauseContext) interface{}

	// Visit a parse tree produced by CqlParser#startingClause.
	VisitStartingClause(ctx *StartingClauseContext) interface{}

	// Visit a parse tree produced by CqlParser#sortClause.
	VisitSortClause(ctx *SortClauseContext) interface{}

	// Visit a parse tree produced by CqlParser#sortDirection.
	VisitSortDirection(ctx *SortDirectionContext) interface{}

	// Visit a parse tree produced by CqlParser#sortByItem.
	VisitSortByItem(ctx *SortByItemContext) interface{}

	// Visit a parse tree produced by CqlParser#qualifiedIdentifier.
	VisitQualifiedIdentifier(ctx *QualifiedIdentifierContext) interface{}

	// Visit a parse tree produced by CqlParser#qualifiedIdentifierExpression.
	VisitQualifiedIdentifierExpression(ctx *QualifiedIdentifierExpressionContext) interface{}

	// Visit a parse tree produced by CqlParser#qualifierExpression.
	VisitQualifierExpression(ctx *QualifierExpressionContext) interface{}

	// Visit a parse tree produced by CqlParser#simplePathIndexer.
	VisitSimplePathIndexer(ctx *SimplePathIndexerContext) interface{}

	// Visit a parse tree produced by CqlParser#simplePathQualifiedIdentifier.
	VisitSimplePathQualifiedIdentifier(ctx *SimplePathQualifiedIdentifierContext) interface{}

	// Visit a parse tree produced by CqlParser#simplePathReferentialIdentifier.
	VisitSimplePathReferentialIdentifier(ctx *SimplePathReferentialIdentifierContext) interface{}

	// Visit a parse tree produced by CqlParser#simpleStringLiteral.
	VisitSimpleStringLiteral(ctx *SimpleStringLiteralContext) interface{}

	// Visit a parse tree produced by CqlParser#simpleNumberLiteral.
	VisitSimpleNumberLiteral(ctx *SimpleNumberLiteralContext) interface{}

	// Visit a parse tree produced by CqlParser#durationBetweenExpression.
	VisitDurationBetweenExpression(ctx *DurationBetweenExpressionContext) interface{}

	// Visit a parse tree produced by CqlParser#inFixSetExpression.
	VisitInFixSetExpression(ctx *InFixSetExpressionContext) interface{}

	// Visit a parse tree produced by CqlParser#retrieveExpression.
	VisitRetrieveExpression(ctx *RetrieveExpressionContext) interface{}

	// Visit a parse tree produced by CqlParser#timingExpression.
	VisitTimingExpression(ctx *TimingExpressionContext) interface{}

	// Visit a parse tree produced by CqlParser#queryExpression.
	VisitQueryExpression(ctx *QueryExpressionContext) interface{}

	// Visit a parse tree produced by CqlParser#notExpression.
	VisitNotExpression(ctx *NotExpressionContext) interface{}

	// Visit a parse tree produced by CqlParser#booleanExpression.
	VisitBooleanExpression(ctx *BooleanExpressionContext) interface{}

	// Visit a parse tree produced by CqlParser#orExpression.
	VisitOrExpression(ctx *OrExpressionContext) interface{}

	// Visit a parse tree produced by CqlParser#castExpression.
	VisitCastExpression(ctx *CastExpressionContext) interface{}

	// Visit a parse tree produced by CqlParser#andExpression.
	VisitAndExpression(ctx *AndExpressionContext) interface{}

	// Visit a parse tree produced by CqlParser#betweenExpression.
	VisitBetweenExpression(ctx *BetweenExpressionContext) interface{}

	// Visit a parse tree produced by CqlParser#membershipExpression.
	VisitMembershipExpression(ctx *MembershipExpressionContext) interface{}

	// Visit a parse tree produced by CqlParser#differenceBetweenExpression.
	VisitDifferenceBetweenExpression(ctx *DifferenceBetweenExpressionContext) interface{}

	// Visit a parse tree produced by CqlParser#inequalityExpression.
	VisitInequalityExpression(ctx *InequalityExpressionContext) interface{}

	// Visit a parse tree produced by CqlParser#equalityExpression.
	VisitEqualityExpression(ctx *EqualityExpressionContext) interface{}

	// Visit a parse tree produced by CqlParser#existenceExpression.
	VisitExistenceExpression(ctx *ExistenceExpressionContext) interface{}

	// Visit a parse tree produced by CqlParser#impliesExpression.
	VisitImpliesExpression(ctx *ImpliesExpressionContext) interface{}

	// Visit a parse tree produced by CqlParser#termExpression.
	VisitTermExpression(ctx *TermExpressionContext) interface{}

	// Visit a parse tree produced by CqlParser#typeExpression.
	VisitTypeExpression(ctx *TypeExpressionContext) interface{}

	// Visit a parse tree produced by CqlParser#dateTimePrecision.
	VisitDateTimePrecision(ctx *DateTimePrecisionContext) interface{}

	// Visit a parse tree produced by CqlParser#dateTimeComponent.
	VisitDateTimeComponent(ctx *DateTimeComponentContext) interface{}

	// Visit a parse tree produced by CqlParser#pluralDateTimePrecision.
	VisitPluralDateTimePrecision(ctx *PluralDateTimePrecisionContext) interface{}

	// Visit a parse tree produced by CqlParser#additionExpressionTerm.
	VisitAdditionExpressionTerm(ctx *AdditionExpressionTermContext) interface{}

	// Visit a parse tree produced by CqlParser#indexedExpressionTerm.
	VisitIndexedExpressionTerm(ctx *IndexedExpressionTermContext) interface{}

	// Visit a parse tree produced by CqlParser#widthExpressionTerm.
	VisitWidthExpressionTerm(ctx *WidthExpressionTermContext) interface{}

	// Visit a parse tree produced by CqlParser#setAggregateExpressionTerm.
	VisitSetAggregateExpressionTerm(ctx *SetAggregateExpressionTermContext) interface{}

	// Visit a parse tree produced by CqlParser#timeUnitExpressionTerm.
	VisitTimeUnitExpressionTerm(ctx *TimeUnitExpressionTermContext) interface{}

	// Visit a parse tree produced by CqlParser#ifThenElseExpressionTerm.
	VisitIfThenElseExpressionTerm(ctx *IfThenElseExpressionTermContext) interface{}

	// Visit a parse tree produced by CqlParser#timeBoundaryExpressionTerm.
	VisitTimeBoundaryExpressionTerm(ctx *TimeBoundaryExpressionTermContext) interface{}

	// Visit a parse tree produced by CqlParser#elementExtractorExpressionTerm.
	VisitElementExtractorExpressionTerm(ctx *ElementExtractorExpressionTermContext) interface{}

	// Visit a parse tree produced by CqlParser#conversionExpressionTerm.
	VisitConversionExpressionTerm(ctx *ConversionExpressionTermContext) interface{}

	// Visit a parse tree produced by CqlParser#typeExtentExpressionTerm.
	VisitTypeExtentExpressionTerm(ctx *TypeExtentExpressionTermContext) interface{}

	// Visit a parse tree produced by CqlParser#predecessorExpressionTerm.
	VisitPredecessorExpressionTerm(ctx *PredecessorExpressionTermContext) interface{}

	// Visit a parse tree produced by CqlParser#pointExtractorExpressionTerm.
	VisitPointExtractorExpressionTerm(ctx *PointExtractorExpressionTermContext) interface{}

	// Visit a parse tree produced by CqlParser#multiplicationExpressionTerm.
	VisitMultiplicationExpressionTerm(ctx *MultiplicationExpressionTermContext) interface{}

	// Visit a parse tree produced by CqlParser#aggregateExpressionTerm.
	VisitAggregateExpressionTerm(ctx *AggregateExpressionTermContext) interface{}

	// Visit a parse tree produced by CqlParser#durationExpressionTerm.
	VisitDurationExpressionTerm(ctx *DurationExpressionTermContext) interface{}

	// Visit a parse tree produced by CqlParser#differenceExpressionTerm.
	VisitDifferenceExpressionTerm(ctx *DifferenceExpressionTermContext) interface{}

	// Visit a parse tree produced by CqlParser#caseExpressionTerm.
	VisitCaseExpressionTerm(ctx *CaseExpressionTermContext) interface{}

	// Visit a parse tree produced by CqlParser#powerExpressionTerm.
	VisitPowerExpressionTerm(ctx *PowerExpressionTermContext) interface{}

	// Visit a parse tree produced by CqlParser#successorExpressionTerm.
	VisitSuccessorExpressionTerm(ctx *SuccessorExpressionTermContext) interface{}

	// Visit a parse tree produced by CqlParser#polarityExpressionTerm.
	VisitPolarityExpressionTerm(ctx *PolarityExpressionTermContext) interface{}

	// Visit a parse tree produced by CqlParser#termExpressionTerm.
	VisitTermExpressionTerm(ctx *TermExpressionTermContext) interface{}

	// Visit a parse tree produced by CqlParser#invocationExpressionTerm.
	VisitInvocationExpressionTerm(ctx *InvocationExpressionTermContext) interface{}

	// Visit a parse tree produced by CqlParser#caseExpressionItem.
	VisitCaseExpressionItem(ctx *CaseExpressionItemContext) interface{}

	// Visit a parse tree produced by CqlParser#dateTimePrecisionSpecifier.
	VisitDateTimePrecisionSpecifier(ctx *DateTimePrecisionSpecifierContext) interface{}

	// Visit a parse tree produced by CqlParser#relativeQualifier.
	VisitRelativeQualifier(ctx *RelativeQualifierContext) interface{}

	// Visit a parse tree produced by CqlParser#offsetRelativeQualifier.
	VisitOffsetRelativeQualifier(ctx *OffsetRelativeQualifierContext) interface{}

	// Visit a parse tree produced by CqlParser#exclusiveRelativeQualifier.
	VisitExclusiveRelativeQualifier(ctx *ExclusiveRelativeQualifierContext) interface{}

	// Visit a parse tree produced by CqlParser#quantityOffset.
	VisitQuantityOffset(ctx *QuantityOffsetContext) interface{}

	// Visit a parse tree produced by CqlParser#temporalRelationship.
	VisitTemporalRelationship(ctx *TemporalRelationshipContext) interface{}

	// Visit a parse tree produced by CqlParser#concurrentWithIntervalOperatorPhrase.
	VisitConcurrentWithIntervalOperatorPhrase(ctx *ConcurrentWithIntervalOperatorPhraseContext) interface{}

	// Visit a parse tree produced by CqlParser#includesIntervalOperatorPhrase.
	VisitIncludesIntervalOperatorPhrase(ctx *IncludesIntervalOperatorPhraseContext) interface{}

	// Visit a parse tree produced by CqlParser#includedInIntervalOperatorPhrase.
	VisitIncludedInIntervalOperatorPhrase(ctx *IncludedInIntervalOperatorPhraseContext) interface{}

	// Visit a parse tree produced by CqlParser#beforeOrAfterIntervalOperatorPhrase.
	VisitBeforeOrAfterIntervalOperatorPhrase(ctx *BeforeOrAfterIntervalOperatorPhraseContext) interface{}

	// Visit a parse tree produced by CqlParser#withinIntervalOperatorPhrase.
	VisitWithinIntervalOperatorPhrase(ctx *WithinIntervalOperatorPhraseContext) interface{}

	// Visit a parse tree produced by CqlParser#meetsIntervalOperatorPhrase.
	VisitMeetsIntervalOperatorPhrase(ctx *MeetsIntervalOperatorPhraseContext) interface{}

	// Visit a parse tree produced by CqlParser#overlapsIntervalOperatorPhrase.
	VisitOverlapsIntervalOperatorPhrase(ctx *OverlapsIntervalOperatorPhraseContext) interface{}

	// Visit a parse tree produced by CqlParser#startsIntervalOperatorPhrase.
	VisitStartsIntervalOperatorPhrase(ctx *StartsIntervalOperatorPhraseContext) interface{}

	// Visit a parse tree produced by CqlParser#endsIntervalOperatorPhrase.
	VisitEndsIntervalOperatorPhrase(ctx *EndsIntervalOperatorPhraseContext) interface{}

	// Visit a parse tree produced by CqlParser#invocationTerm.
	VisitInvocationTerm(ctx *InvocationTermContext) interface{}

	// Visit a parse tree produced by CqlParser#literalTerm.
	VisitLiteralTerm(ctx *LiteralTermContext) interface{}

	// Visit a parse tree produced by CqlParser#externalConstantTerm.
	VisitExternalConstantTerm(ctx *ExternalConstantTermContext) interface{}

	// Visit a parse tree produced by CqlParser#intervalSelectorTerm.
	VisitIntervalSelectorTerm(ctx *IntervalSelectorTermContext) interface{}

	// Visit a parse tree produced by CqlParser#tupleSelectorTerm.
	VisitTupleSelectorTerm(ctx *TupleSelectorTermContext) interface{}

	// Visit a parse tree produced by CqlParser#instanceSelectorTerm.
	VisitInstanceSelectorTerm(ctx *InstanceSelectorTermContext) interface{}

	// Visit a parse tree produced by CqlParser#listSelectorTerm.
	VisitListSelectorTerm(ctx *ListSelectorTermContext) interface{}

	// Visit a parse tree produced by CqlParser#codeSelectorTerm.
	VisitCodeSelectorTerm(ctx *CodeSelectorTermContext) interface{}

	// Visit a parse tree produced by CqlParser#conceptSelectorTerm.
	VisitConceptSelectorTerm(ctx *ConceptSelectorTermContext) interface{}

	// Visit a parse tree produced by CqlParser#parenthesizedTerm.
	VisitParenthesizedTerm(ctx *ParenthesizedTermContext) interface{}

	// Visit a parse tree produced by CqlParser#qualifiedMemberInvocation.
	VisitQualifiedMemberInvocation(ctx *QualifiedMemberInvocationContext) interface{}

	// Visit a parse tree produced by CqlParser#qualifiedFunctionInvocation.
	VisitQualifiedFunctionInvocation(ctx *QualifiedFunctionInvocationContext) interface{}

	// Visit a parse tree produced by CqlParser#qualifiedFunction.
	VisitQualifiedFunction(ctx *QualifiedFunctionContext) interface{}

	// Visit a parse tree produced by CqlParser#memberInvocation.
	VisitMemberInvocation(ctx *MemberInvocationContext) interface{}

	// Visit a parse tree produced by CqlParser#functionInvocation.
	VisitFunctionInvocation(ctx *FunctionInvocationContext) interface{}

	// Visit a parse tree produced by CqlParser#thisInvocation.
	VisitThisInvocation(ctx *ThisInvocationContext) interface{}

	// Visit a parse tree produced by CqlParser#indexInvocation.
	VisitIndexInvocation(ctx *IndexInvocationContext) interface{}

	// Visit a parse tree produced by CqlParser#totalInvocation.
	VisitTotalInvocation(ctx *TotalInvocationContext) interface{}

	// Visit a parse tree produced by CqlParser#function.
	VisitFunction(ctx *FunctionContext) interface{}

	// Visit a parse tree produced by CqlParser#ratio.
	VisitRatio(ctx *RatioContext) interface{}

	// Visit a parse tree produced by CqlParser#booleanLiteral.
	VisitBooleanLiteral(ctx *BooleanLiteralContext) interface{}

	// Visit a parse tree produced by CqlParser#nullLiteral.
	VisitNullLiteral(ctx *NullLiteralContext) interface{}

	// Visit a parse tree produced by CqlParser#stringLiteral.
	VisitStringLiteral(ctx *StringLiteralContext) interface{}

	// Visit a parse tree produced by CqlParser#numberLiteral.
	VisitNumberLiteral(ctx *NumberLiteralContext) interface{}

	// Visit a parse tree produced by CqlParser#longNumberLiteral.
	VisitLongNumberLiteral(ctx *LongNumberLiteralContext) interface{}

	// Visit a parse tree produced by CqlParser#dateTimeLiteral.
	VisitDateTimeLiteral(ctx *DateTimeLiteralContext) interface{}

	// Visit a parse tree produced by CqlParser#dateLiteral.
	VisitDateLiteral(ctx *DateLiteralContext) interface{}

	// Visit a parse tree produced by CqlParser#timeLiteral.
	VisitTimeLiteral(ctx *TimeLiteralContext) interface{}

	// Visit a parse tree produced by CqlParser#quantityLiteral.
	VisitQuantityLiteral(ctx *QuantityLiteralContext) interface{}

	// Visit a parse tree produced by CqlParser#ratioLiteral.
	VisitRatioLiteral(ctx *RatioLiteralContext) interface{}

	// Visit a parse tree produced by CqlParser#intervalSelector.
	VisitIntervalSelector(ctx *IntervalSelectorContext) interface{}

	// Visit a parse tree produced by CqlParser#tupleSelector.
	VisitTupleSelector(ctx *TupleSelectorContext) interface{}

	// Visit a parse tree produced by CqlParser#tupleElementSelector.
	VisitTupleElementSelector(ctx *TupleElementSelectorContext) interface{}

	// Visit a parse tree produced by CqlParser#instanceSelector.
	VisitInstanceSelector(ctx *InstanceSelectorContext) interface{}

	// Visit a parse tree produced by CqlParser#instanceElementSelector.
	VisitInstanceElementSelector(ctx *InstanceElementSelectorContext) interface{}

	// Visit a parse tree produced by CqlParser#listSelector.
	VisitListSelector(ctx *ListSelectorContext) interface{}

	// Visit a parse tree produced by CqlParser#displayClause.
	VisitDisplayClause(ctx *DisplayClauseContext) interface{}

	// Visit a parse tree produced by CqlParser#codeSelector.
	VisitCodeSelector(ctx *CodeSelectorContext) interface{}

	// Visit a parse tree produced by CqlParser#conceptSelector.
	VisitConceptSelector(ctx *ConceptSelectorContext) interface{}

	// Visit a parse tree produced by CqlParser#keyword.
	VisitKeyword(ctx *KeywordContext) interface{}

	// Visit a parse tree produced by CqlParser#reservedWord.
	VisitReservedWord(ctx *ReservedWordContext) interface{}

	// Visit a parse tree produced by CqlParser#keywordIdentifier.
	VisitKeywordIdentifier(ctx *KeywordIdentifierContext) interface{}

	// Visit a parse tree produced by CqlParser#obsoleteIdentifier.
	VisitObsoleteIdentifier(ctx *ObsoleteIdentifierContext) interface{}

	// Visit a parse tree produced by CqlParser#functionIdentifier.
	VisitFunctionIdentifier(ctx *FunctionIdentifierContext) interface{}

	// Visit a parse tree produced by CqlParser#typeNameIdentifier.
	VisitTypeNameIdentifier(ctx *TypeNameIdentifierContext) interface{}

	// Visit a parse tree produced by CqlParser#referentialIdentifier.
	VisitReferentialIdentifier(ctx *ReferentialIdentifierContext) interface{}

	// Visit a parse tree produced by CqlParser#referentialOrTypeNameIdentifier.
	VisitReferentialOrTypeNameIdentifier(ctx *ReferentialOrTypeNameIdentifierContext) interface{}

	// Visit a parse tree produced by CqlParser#identifierOrFunctionIdentifier.
	VisitIdentifierOrFunctionIdentifier(ctx *IdentifierOrFunctionIdentifierContext) interface{}

	// Visit a parse tree produced by CqlParser#identifier.
	VisitIdentifier(ctx *IdentifierContext) interface{}

	// Visit a parse tree produced by CqlParser#externalConstant.
	VisitExternalConstant(ctx *ExternalConstantContext) interface{}

	// Visit a parse tree produced by CqlParser#paramList.
	VisitParamList(ctx *ParamListContext) interface{}

	// Visit a parse tree produced by CqlParser#quantity.
	VisitQuantity(ctx *QuantityContext) interface{}

	// Visit a parse tree produced by CqlParser#unit.
	VisitUnit(ctx *UnitContext) interface{}

}