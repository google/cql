// Copyright 2024 Google LLC
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

	"github.com/google/cql/internal/convert"
	"github.com/google/cql/internal/embeddata/third_party/cqframework/cql"
	"github.com/google/cql/model"
	"github.com/google/cql/types"
	"github.com/antlr4-go/antlr/v4"
)

func (v *visitor) VisitQuery(ctx *cql.QueryContext) model.IExpression {
	// Top level scope for the main query source aliases.
	v.refs.EnterScope()
	defer v.refs.ExitScope()

	q := &model.Query{}
	var err error

	q, err = v.parseSourceClause(ctx.SourceClause(), q)
	if err != nil {
		return v.badExpression(err.Error(), ctx.SourceClause())
	}

	q, err = v.parseLetClause(ctx.LetClause(), q)
	if err != nil {
		return v.badExpression(err.Error(), ctx.LetClause())
	}

	for _, inc := range ctx.AllQueryInclusionClause() {
		q, err = v.parseIncusionClause(inc, q)
		if err != nil {
			return v.badExpression(err.Error(), inc)
		}
	}

	q, err = v.parseWhereClause(ctx.WhereClause(), q)
	if err != nil {
		return v.badExpression(err.Error(), ctx.WhereClause())
	}

	q, err = v.parseSortClause(ctx.SortClause(), q)
	if err != nil {
		return v.badExpression(err.Error(), ctx.SortClause())
	}

	q, err = v.parseAggregateClause(ctx.AggregateClause(), q)
	if err != nil {
		return v.badExpression(err.Error(), ctx.AggregateClause())
	}

	if ctx.AggregateClause() == nil {
		// parseReturnClauseAndSetResultType could parse or in the case of multi-source queries inserts
		// a return clause. It also sets the result type, even if there is no return clause.
		q, err = v.parseReturnClauseAndSetResultType(ctx.ReturnClause(), q)
		if err != nil {
			return v.badExpression(err.Error(), ctx.ReturnClause())
		}
	}
	return q
}

func (v *visitor) VisitQuerySource(ctx *cql.QuerySourceContext) model.IExpression {
	if ctx.Retrieve() != nil {
		return v.VisitExpression(ctx.Retrieve())
	}
	if ctx.QualifiedIdentifierExpression() != nil {
		return v.VisitExpression(ctx.QualifiedIdentifierExpression())
	}
	if ctx.Expression() != nil {
		return v.VisitExpression(ctx.Expression())
	}
	return v.badExpression("internal error - the grammar should prevent us from landing here", ctx)
}

func (v *visitor) parseSourceClause(sc cql.ISourceClauseContext, q *model.Query) (*model.Query, error) {
	// If the keyword from is used in the query then it will start with a TerminalNode.
	_, hasFrom := sc.GetChild(0).(antlr.TerminalNode)
	if len(sc.AllAliasedQuerySource()) > 1 && !hasFrom {
		return nil, fmt.Errorf("for multi-source queries the keyword from is required")
	}

	for _, aqs := range sc.AllAliasedQuerySource() {
		// Aliases are defined in the top level scope of the query.
		as, err := v.parseAliasedQuerySource(aqs)
		if err != nil {
			return nil, err
		}
		q.Source = append(q.Source, as)
	}

	return q, nil
}

func (v *visitor) parseLetClause(lc cql.ILetClauseContext, q *model.Query) (*model.Query, error) {
	if lc == nil {
		return q, nil
	}

	for _, let := range lc.AllLetClauseItem() {
		l := &model.LetClause{
			Expression: v.VisitExpression(let.Expression()),
			Identifier: v.VisitIdentifier(let.Identifier()),
		}
		l.Element = &model.Element{ResultType: l.Expression.GetResultType()}

		f := func() model.IExpression {
			return &model.QueryLetRef{Name: l.Identifier, Expression: model.ResultType(l.GetResultType())}
		}
		// Aliases are defined in the top level scope of the query.
		if err := v.refs.Alias(l.Identifier, f); err != nil {
			return nil, err
		}

		q.Let = append(q.Let, l)
	}
	return q, nil
}

func (v *visitor) parseIncusionClause(inc cql.IQueryInclusionClauseContext, q *model.Query) (*model.Query, error) {
	v.refs.EnterScope()
	defer v.refs.ExitScope()

	var aqs cql.IAliasedQuerySourceContext
	var exp cql.IExpressionContext
	var with bool
	if inc.WithClause() != nil {
		with = true
		aqs = inc.WithClause().AliasedQuerySource()
		exp = inc.WithClause().Expression()
	} else {
		with = false
		aqs = inc.WithoutClause().AliasedQuerySource()
		exp = inc.WithoutClause().Expression()
	}

	aqsModel, err := v.parseAliasedQuerySource(aqs)
	if err != nil {
		return nil, err
	}

	expModel := v.VisitExpression(exp)
	res, err := convert.OperandImplicitConverter(expModel.GetResultType(), types.Boolean, expModel, v.modelInfo)
	if err != nil {
		return nil, err
	}
	if !res.Matched {
		return nil, fmt.Errorf("result of a query inclusion clause must be implicitly convertible to a boolean, could not convert %v to boolean", expModel.GetResultType())
	}

	rClause := &model.RelationshipClause{
		Element:    &model.Element{ResultType: types.Boolean},
		Expression: aqsModel.Source,
		Alias:      aqsModel.Alias,
		SuchThat:   res.WrappedOperand,
	}

	if with {
		q.Relationship = append(q.Relationship, &model.With{RelationshipClause: rClause})
	} else {
		q.Relationship = append(q.Relationship, &model.Without{RelationshipClause: rClause})
	}
	return q, nil
}

func (v *visitor) parseWhereClause(wc cql.IWhereClauseContext, q *model.Query) (*model.Query, error) {
	if wc == nil {
		return q, nil
	}

	wExp := v.VisitExpression(wc.Expression())

	res, err := convert.OperandImplicitConverter(wExp.GetResultType(), types.Boolean, wExp, v.modelInfo)
	if err != nil {
		return nil, err
	}
	if !res.Matched {
		return nil, fmt.Errorf("result of a where clause must be implicitly convertible to a boolean, could not convert %v to boolean", wExp.GetResultType())
	}
	q.Where = res.WrappedOperand
	return q, nil
}

func (v *visitor) parseSortClause(sc cql.ISortClauseContext, q *model.Query) (*model.Query, error) {
	// TODO(b/316961394): Add check for sortability for CQL query sort columns.
	if sc == nil {
		return q, nil
	}

	var sortByItems []model.ISortByItem
	if sbd, found := maybeGetChildNode[*cql.SortDirectionContext](sc.GetChildren(), nil); found {
		sortDir, err := parseSortDirection(sbd.GetText())
		if err != nil {
			return nil, err
		}
		sortByItems = []model.ISortByItem{
			&model.SortByDirection{
				SortByItem: &model.SortByItem{
					Direction: sortDir,
				},
			},
		}
	} else if sbi, found := maybeGetChildNode[*cql.SortByItemContext](sc.GetChildren(), nil); found {
		// Sort direction is optional in the "sort by" clause, and defaults to ascending.
		var sortText string = "ascending"
		if sbi.SortDirection() != nil {
			sortText = sbi.SortDirection().GetText()
		}
		sortDir, err := parseSortDirection(sortText)
		if err != nil {
			return nil, err
		}

		v.refs.EnterStructScope(func() model.IExpression { return q.Source[0] })
		defer v.refs.ExitStructScope()

		// Set sort context flag to allow forward references in sort expressions
		v.inSortContext = true
		sortExpr := v.VisitExpression(sbi.ExpressionTerm())
		v.inSortContext = false

		switch t := sortExpr.(type) {
		case *model.IdentifierRef:
			sortByItems = []model.ISortByItem{
				&model.SortByColumn{
					SortByItem: &model.SortByItem{
						Direction: sortDir,
					},
					Path: t.Name,
				},
			}
		default:
			sortByItems = []model.ISortByItem{
				&model.SortByExpression{
					SortByItem: &model.SortByItem{
						Direction: sortDir,
					},
					SortExpression: t,
				},
			}
		}

		// TODO(b/317402356): Add static type checking for column paths.
	} else {
		return nil, errors.New("item or direction to sort by was not found")
	}

	q.Sort = &model.SortClause{ByItems: sortByItems}
	return q, nil
}

func parseSortDirection(s string) (model.SortDirection, error) {
	switch s {
	case "ascending", "asc":
		return model.ASCENDING, nil
	case "descending", "desc":
		return model.DESCENDING, nil
	}
	return model.UNSETSORTDIRECTION, fmt.Errorf("unsupported sort direction, expected asc, ascending, desc or descending, got: %s", s)
}

func (v *visitor) parseAggregateClause(ac cql.IAggregateClauseContext, q *model.Query) (*model.Query, error) {
	if ac == nil {
		return q, nil
	}

	aModel := &model.AggregateClause{
		Identifier: v.VisitIdentifier(ac.Identifier()),
	}

	if _, ok := ac.GetChild(1).(antlr.TerminalNode); ok {
		aModel.Distinct = ac.GetChild(1).(antlr.TerminalNode).GetText() == "distinct"
	}

	if ac.StartingClause() != nil {
		if ac.StartingClause().SimpleLiteral() != nil {
			aModel.Starting = v.VisitExpression(ac.StartingClause().SimpleLiteral())
		} else if ac.StartingClause().Quantity() != nil {
			aModel.Starting = v.VisitExpression(ac.StartingClause().Quantity())
		} else if ac.StartingClause().Expression() != nil {
			aModel.Starting = v.VisitExpression(ac.StartingClause().Expression())
		} else {
			return nil, fmt.Errorf("internal error - grammar should not allow another StartingClause")
		}
	} else {
		aModel.Starting = model.NewLiteral("null", types.Any)
	}

	// Define an alias for aggregation variable.
	v.refs.EnterScope()
	defer v.refs.ExitScope()
	v.refs.Alias(aModel.Identifier, func() model.IExpression {
		return &model.AliasRef{Name: aModel.Identifier, Expression: model.ResultType(aModel.Starting.GetResultType())}
	})

	aModel.Expression = v.VisitExpression(ac.Expression())
	aModel.Element = &model.Element{ResultType: aModel.Expression.GetResultType()}

	// The result of the query is the result of the last iteration of the aggregate expression.
	q.Expression = model.ResultType(aModel.GetResultType())
	q.Aggregate = aModel

	return q, nil
}

func (v *visitor) parseReturnClauseAndSetResultType(rc cql.IReturnClauseContext, q *model.Query) (*model.Query, error) {
	atLeastOneSourceList := false
	for _, as := range q.Source {
		_, ok := as.GetResultType().(*types.List)
		if ok {
			atLeastOneSourceList = true
			break
		}
	}

	if rc == nil && len(q.Source) > 1 {
		// For multi-source queries if there is no return clase the parser inserts a Tuple selector.
		tModel := &model.Tuple{}
		tType := &types.Tuple{ElementTypes: make(map[string]types.IType)}
		for _, aSource := range q.Source {
			aRef, err := v.refs.ResolveLocal(aSource.Alias)
			if err != nil {
				return nil, err
			}
			tModel.Elements = append(tModel.Elements, &model.TupleElement{Name: aSource.Alias, Value: aRef()})
			tType.ElementTypes[aSource.Alias] = aRef().GetResultType()
		}

		tModel.Expression = model.ResultType(tType)
		q.Return = &model.ReturnClause{
			Distinct:   true,
			Expression: tModel,
			Element:    &model.Element{ResultType: tType},
		}

		q.Expression = model.ResultType(tType)
		if atLeastOneSourceList {
			q.Expression = model.ResultType(&types.List{ElementType: q.Expression.GetResultType()})
		}
		return q, nil
	}

	if rc == nil {
		// No return clause and there is only a single source.
		q.Expression = model.ResultType(q.Source[0].GetResultType())
		return q, nil
	}

	// There is a return clause.
	rModel := &model.ReturnClause{
		Expression: v.VisitExpression(rc.Expression()),
		Distinct:   true,
	}
	rModel.Element = &model.Element{ResultType: rModel.Expression.GetResultType()}
	if rc.GetChildCount() == 3 && rc.GetChild(1).(antlr.TerminalNode).GetText() == "all" {
		rModel.Distinct = false
	}
	q.Return = rModel

	if atLeastOneSourceList {
		q.Expression = model.ResultType(&types.List{ElementType: rModel.GetResultType()})
	} else {
		q.Expression = model.ResultType(rModel.GetResultType())
	}
	return q, nil
}

func (v *visitor) parseAliasedQuerySource(aqs cql.IAliasedQuerySourceContext) (*model.AliasedSource, error) {
	alias := v.VisitIdentifier(aqs.Alias().Identifier())
	source := v.VisitExpression(aqs.QuerySource())
	aqsModel := &model.AliasedSource{
		Alias:      alias,
		Source:     source,
		Expression: model.ResultType(source.GetResultType()),
	}

	// If the AliasedSource is a List, when referencing the AliasRef inside the query it
	// should refer to a single element of the AliasedSource list.
	// For example in [Observation] O Where O.status = 'final' the Aliased reference O when resolved
	// inside O.status should be of type FHIR.Observation instead of List<FHIR.Observation>.
	aliasRefResultType := aqsModel.GetResultType()
	listAliasResultType, ok := aqsModel.GetResultType().(*types.List)
	if ok {
		// If it's a list, we set the alias ref ResultType to the ElementType.
		aliasRefResultType = listAliasResultType.ElementType
	}

	f := func() model.IExpression {
		return &model.AliasRef{Name: alias, Expression: model.ResultType(aliasRefResultType)}
	}
	if err := v.refs.Alias(alias, f); err != nil {
		return nil, err
	}
	return aqsModel, nil
}
