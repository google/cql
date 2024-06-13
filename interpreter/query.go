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

package interpreter

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/google/cql/model"
	"github.com/google/cql/result"
	"github.com/google/cql/types"
)

// iteration is the aliases for a single iteration of the query. For a CQL query like define Foo:
// from (4) A, ({1, 2, 3}) B iteration may be [A: {A, 4}, B: {B, 1}]. Map is from the alias's name
// to the alias.
type iteration map[string]alias

func (i iteration) Equal(a iteration) bool {
	if len(i) != len(a) {
		return false
	}
	for _, aAlias := range a {
		iAlias, ok := i[aAlias.alias]
		if !ok {
			return false
		}
		// TODO: b/301606416 - when equal is implemented, call that logic from here.
		if !iAlias.obj.Equal(aAlias.obj) {
			return false
		}
	}
	return true
}

type alias struct {
	alias string
	obj   result.Value
}

func (i *interpreter) evalQuery(q *model.Query) (result.Value, error) {
	// The top level scope holds let clauses. Each query iteration defines a nested scope for the
	// source and relationship aliases.
	i.refs.EnterScope()
	defer i.refs.ExitScope()

	iters, sourceObjs, err := i.sourceClause(q.Source)
	if err != nil {
		return result.Value{}, err
	}

	sourceObj, err := i.letClause(q.Let)
	if err != nil {
		return result.Value{}, err
	}
	sourceObjs = append(sourceObjs, sourceObj...)

	for _, relationship := range q.Relationship {
		var err error
		var sourceObj result.Value
		iters, sourceObj, err = i.relationshipClause(iters, relationship)
		if err != nil {
			return result.Value{}, err
		}
		sourceObjs = append(sourceObjs, sourceObj)
	}

	iters, err = i.whereClause(iters, q.Where)
	if err != nil {
		return result.Value{}, err
	}

	// finalVals is the list of values that will be returned by the query.
	finalVals := make([]result.Value, 0, len(iters))

	if q.Aggregate != nil {
		finalVals, err = i.aggregateClause(iters, q.Aggregate)
		if err != nil {
			return result.Value{}, err
		}
	}

	if q.Return != nil {
		finalVals, err = i.returnClause(iters, q.Return)
		if err != nil {
			return result.Value{}, err
		}
	}

	if q.Return == nil && q.Aggregate == nil {
		if len(q.Source) == 1 {
			// If there is no return clause and this was a single source query, unpack the alias.
			for _, iter := range iters {
				for _, alias := range iter {
					finalVals = append(finalVals, alias.obj)
				}
			}
		} else {
			return result.Value{}, errors.New("internal error - multi-source queries must have a return clause, the parser should insert a default one if the user did not write one")
		}
	}

	if q.Sort != nil && len(finalVals) > 0 {
		if sbd, ok := q.Sort.ByItems[0].(*model.SortByDirection); ok {
			err := sortByDirection(finalVals, sbd)
			if err != nil {
				return result.Value{}, err
			}
		} else {
			i.sortByColumn(finalVals, q.Sort.ByItems)
			if err != nil {
				return result.Value{}, err
			}
		}
	}

	// Right now, we expect the query result type to be a List and forward that along.
	returnStaticType, ok := q.GetResultType().(*types.List)
	if ok {
		return result.NewWithSources(result.List{Value: finalVals, StaticType: returnStaticType}, q, sourceObjs...)
	}

	if len(finalVals) == 0 {
		return result.NewWithSources(nil, q, sourceObjs...)
	} else if len(finalVals) == 1 {
		return result.NewWithSources(finalVals[0].GolangValue(), q, sourceObjs...)
	}

	return result.Value{}, fmt.Errorf("internal error - query static result type is %v, but resulted in a list of values", returnStaticType)
}

// sourceClause returns the iterations the query should be executed on and the source values. For
// multi-source queries the iterations are the cartesian product. For example,
// define Foo: from (4) A, ({1, 2, 3}) B will return
// [{A: {A, 4}, B: {B, 1}}, {A: {A, 4}, B: {B, 2}}, {A: {A, 4}, B: {B, 3}}].
func (i *interpreter) sourceClause(s []*model.AliasedSource) ([]iteration, []result.Value, error) {
	if s == nil || len(s) == 0 {
		return nil, nil, fmt.Errorf("internal error - query must have at least one source")
	}

	sourceObjs := make([]result.Value, 0, len(s))

	// For a multi-source query like (4) A, ({1, 2, 3}) B aliases will be
	// [[{A, 4}], [{B, 1}, {B, 2}, {B, 3}]].
	aliases := [][]alias{}
	for _, source := range s {
		a := []alias{}
		obj, err := i.evalExpression(source.Source)
		if err != nil {
			return nil, nil, err
		}
		sourceObjs = append(sourceObjs, obj)
		l, err := result.ToSlice(obj)
		if err == nil {
			// The source is a list so unpack.
			for _, obj := range l {
				a = append(a, alias{alias: source.Alias, obj: obj})
			}
		} else if errors.Is(err, result.ErrCannotConvert) {
			// The source is not a list, append directly.
			a = append(a, alias{alias: source.Alias, obj: obj})
		} else {
			return nil, nil, err
		}
		aliases = append(aliases, a)
	}
	return cartesianProduct(aliases), sourceObjs, nil
}

// cartesianProduct converts [[{A, 4}], [{B, 1}, {B, 2}, {B, 3}]] into the cartesian product
// [{A: {A, 4}, B: {B, 1}}, {A: {A, 4}, B: {B, 2}}, {A: {A, 4}, B: {B, 3}}].
func cartesianProduct(aliases [][]alias) []iteration {
	if len(aliases) == 1 {
		cartIters := []iteration{}
		for _, a := range aliases[0] {

			cartIters = append(cartIters, iteration{a.alias: a})
		}
		return cartIters
	}

	subCartIters := cartesianProduct(aliases[1:])
	cartIters := []iteration{}
	for _, a := range aliases[0] {
		for _, b := range subCartIters {
			cartIter := iteration{a.alias: a}
			for k, v := range b {
				cartIter[k] = v
			}
			cartIters = append(cartIters, cartIter)
		}
	}
	return cartIters
}

func (i *interpreter) letClause(m []*model.LetClause) ([]result.Value, error) {
	sourceObjs := make([]result.Value, 0, len(m))
	for _, letClause := range m {
		obj, err := i.evalExpression(letClause.Expression)
		if err != nil {
			return nil, err
		}
		sourceObjs = append(sourceObjs, obj)

		if err := i.refs.Alias(letClause.Identifier, obj); err != nil {
			return nil, err
		}
	}

	return sourceObjs, nil
}

func (i *interpreter) relationshipClause(iters []iteration, m model.IRelationshipClause) ([]iteration, result.Value, error) {
	var relAliasName string
	var relAliasSource model.IExpression
	var suchThat model.IExpression
	var with bool

	switch t := m.(type) {
	case *model.With:
		relAliasName = t.Alias
		relAliasSource = t.Expression
		suchThat = t.SuchThat
		with = true
	case *model.Without:
		relAliasName = t.Alias
		relAliasSource = t.Expression
		suchThat = t.SuchThat
		with = false
	default:
		return nil, result.Value{}, fmt.Errorf("internal error - there should only be a with or without relationship clause, got: %T", m)
	}

	relSourceObj, err := i.evalExpression(relAliasSource)
	if err != nil {
		return nil, result.Value{}, err
	}

	var relIters []result.Value
	l, err := result.ToSlice(relSourceObj)
	if err == nil {
		relIters = l
	} else if errors.Is(err, result.ErrCannotConvert) {
		relIters = []result.Value{relSourceObj}
	} else {
		return nil, result.Value{}, err
	}

	filteredIters := []iteration{}
	for _, iter := range iters {
		for _, relIter := range relIters {
			i.refs.EnterScope()
			// Define the alias for the relationship.
			if err := i.refs.Alias(relAliasName, relIter); err != nil {
				return nil, result.Value{}, err
			}
			// Define the aliases for the query iterations.
			for _, alias := range iter {
				if err := i.refs.Alias(alias.alias, alias.obj); err != nil {
					return nil, result.Value{}, err
				}
			}

			filter, err := i.evalExpression(suchThat)
			if err != nil {
				return nil, result.Value{}, err
			}

			if !result.IsNull(filter) && !filter.RuntimeType().Equal(types.Boolean) {
				return nil, result.Value{}, fmt.Errorf("internal error - such that clause of a query must evaluate to a boolean or null, instead got %v", filter.RuntimeType())
			}
			if (with && filter.GolangValue() == true) || (!with && filter.GolangValue() == false) {
				// We found a relationship where such that expression evaluated to true. Save this iter in
				// filteredIters and break.
				i.refs.ExitScope()
				filteredIters = append(filteredIters, iter)
				break
			}
			i.refs.ExitScope()
		}
	}
	return filteredIters, relSourceObj, nil
}

func (i *interpreter) whereClause(iters []iteration, where model.IExpression) ([]iteration, error) {
	if where == nil {
		return iters, nil
	}

	var filteredIters []iteration
	for _, iter := range iters {
		i.refs.EnterScope()
		for _, alias := range iter {
			if err := i.refs.Alias(alias.alias, alias.obj); err != nil {
				return nil, err
			}
		}
		filter, err := i.evalExpression(where)
		if err != nil {
			return nil, err
		}

		if !result.IsNull(filter) && !filter.RuntimeType().Equal(types.Boolean) {
			return nil, fmt.Errorf("internal error - where clause of a query must evaluate to a boolean or null, instead got %v", filter.RuntimeType())
		}
		if filter.GolangValue() == true {
			filteredIters = append(filteredIters, iter)
		}
		i.refs.ExitScope()
	}
	return filteredIters, nil
}

func (i *interpreter) aggregateClause(iters []iteration, aggregateClause *model.AggregateClause) ([]result.Value, error) {
	var filteredIters []iteration
	if aggregateClause.Distinct {
		for _, iter := range iters {
			filteredIters = appendIfIterDistinct(filteredIters, iter)
		}
	} else {
		filteredIters = iters
	}

	aggregateObj, err := i.evalExpression(aggregateClause.Starting)
	if err != nil {
		return nil, err
	}

	for _, iter := range filteredIters {
		i.refs.EnterScope()

		if err := i.refs.Alias(aggregateClause.Identifier, aggregateObj); err != nil {
			i.refs.ExitScope()
			return nil, err
		}

		for _, alias := range iter {
			if err := i.refs.Alias(alias.alias, alias.obj); err != nil {
				i.refs.ExitScope()
				return nil, err
			}
		}

		aggregateObj, err = i.evalExpression(aggregateClause.Expression)
		if err != nil {
			i.refs.ExitScope()
			return nil, err
		}

		i.refs.ExitScope()
	}
	return []result.Value{aggregateObj}, nil
}

func (i *interpreter) returnClause(iters []iteration, returnClause *model.ReturnClause) ([]result.Value, error) {
	returnObjs := make([]result.Value, 0, len(iters))
	for _, iter := range iters {
		i.refs.EnterScope()
		for _, alias := range iter {
			if err := i.refs.Alias(alias.alias, alias.obj); err != nil {
				return nil, err
			}
		}
		retObj, err := i.evalExpression(returnClause.Expression)
		if err != nil {
			return nil, err
		}
		if returnClause.Distinct {
			returnObjs = appendIfDistinct(returnObjs, retObj)
		} else {
			returnObjs = append(returnObjs, retObj)
		}
		i.refs.ExitScope()
	}
	return returnObjs, nil
}

func appendIfIterDistinct(distinctIters []iteration, maybeDistinct iteration) []iteration {
	for _, distinct := range distinctIters {
		if distinct.Equal(maybeDistinct) {
			return distinctIters
		}
	}
	return append(distinctIters, maybeDistinct)
}

func appendIfDistinct(objs []result.Value, obj result.Value) []result.Value {
	for _, o := range objs {
		// TODO: b/327612359 - when distinct is implemented, call that logic from here.
		if o.Equal(obj) {
			return objs
		}
	}
	return append(objs, obj)
}

func sortByDirection(objs []result.Value, sbd *model.SortByDirection) error {
	// Only allow Dates, DateTimes, Integers, Decimals, Longs and Strings for now.
	// TODO(b/316984809): add sorting support for other types and nulls.
	var sortFunc func(a, b result.Value) int
	rt := objs[0].RuntimeType()
	switch rt {
	case types.Integer:
		sortFunc = func(a, b result.Value) int {
			av := a.GolangValue().(int32)
			bv := b.GolangValue().(int32)
			if sbd.SortByItem.Direction == model.DESCENDING {
				return compareNumeralInt(bv, av)
			}
			return compareNumeralInt(av, bv)
		}
	case types.Decimal:
		sortFunc = func(a, b result.Value) int {
			av := a.GolangValue().(float64)
			bv := b.GolangValue().(float64)
			if sbd.SortByItem.Direction == model.DESCENDING {
				return compareNumeralInt(bv, av)
			}
			return compareNumeralInt(av, bv)
		}
	case types.Long:
		sortFunc = func(a, b result.Value) int {
			av := a.GolangValue().(int64)
			bv := b.GolangValue().(int64)
			if sbd.SortByItem.Direction == model.DESCENDING {
				return compareNumeralInt(bv, av)
			}
			return compareNumeralInt(av, bv)
		}
	case types.String:
		sortFunc = func(a, b result.Value) int {
			av := a.GolangValue().(string)
			bv := b.GolangValue().(string)
			if sbd.SortByItem.Direction == model.DESCENDING {
				return strings.Compare(bv, av)
			}
			return strings.Compare(av, bv)
		}
	case types.Date:
		sortFunc = func(a, b result.Value) int {
			av := a.GolangValue().(result.Date).Date
			bv := b.GolangValue().(result.Date).Date
			if sbd.SortByItem.Direction == model.DESCENDING {
				return bv.Compare(av)
			}
			return av.Compare(bv)
		}
	case types.DateTime:
		// TODO: b/301606416 - we should use a precision aware comparison here.
		sortFunc = func(a, b result.Value) int {
			av := a.GolangValue().(result.DateTime).Date
			bv := b.GolangValue().(result.DateTime).Date
			if sbd.SortByItem.Direction == model.DESCENDING {
				return bv.Compare(av)
			}
			return av.Compare(bv)
		}
	default:
		return fmt.Errorf("sort column of a query by direction must evaluate to a Date, DateTime, Integer, Decimal, or Long, instead got %v", objs[0].RuntimeType())
	}
	slices.SortFunc(objs[:], sortFunc)
	return nil
}

// compareNumeralInt returns the integer comparison value of two numeric values.
func compareNumeralInt[t float64 | int64 | int32](left, right t) int {
	switch compareNumeral(left, right) {
	case leftBeforeRight:
		return -1
	case leftEqualRight:
		return 0
	case leftAfterRight:
		return 1
	default:
		panic("internal error - unsupported comparison, this case should never happen")
	}
}

func (i *interpreter) sortByColumn(objs []result.Value, sbis []model.ISortByItem) error {
	// Validate sort column types.
	for _, sortItems := range sbis {
		// TODO(b/316984809): Is this validation in advance necessary? What if other values (beyond
		// objs[0]) have a different runtime type for the property (e.g. if they're a choice type)?
		// Consider validating types inline during the sort instead.
		path := sortItems.(*model.SortByColumn).Path
		propertyType, err := i.modelInfo.PropertyTypeSpecifier(objs[0].RuntimeType(), path)
		if err != nil {
			return err
		}
		columnVal, err := i.valueProperty(objs[0], path, propertyType)
		if err != nil {
			return err
		}
		// Strictly only allow DateTimes for now.
		// TODO(b/316984809): add sorting support for other types.
		if !columnVal.RuntimeType().Equal(types.DateTime) {
			return fmt.Errorf("sort column of a query must evaluate to a date time, instead got %v", columnVal.RuntimeType())
		}
	}

	var sortErr error = nil
	slices.SortFunc(objs[:], func(a, b result.Value) int {
		for _, sortItems := range sbis {
			sortCol := sortItems.(*model.SortByColumn)
			// Passing the static types here is likely unimportant, but we compute it for completeness.
			aType, err := i.modelInfo.PropertyTypeSpecifier(a.RuntimeType(), sortCol.Path)
			if err != nil {
				sortErr = err
				continue
			}
			ap, err := i.valueProperty(a, sortCol.Path, aType)
			if err != nil {
				sortErr = err
				continue
			}
			bType, err := i.modelInfo.PropertyTypeSpecifier(b.RuntimeType(), sortCol.Path)
			if err != nil {
				sortErr = err
				continue
			}
			bp, err := i.valueProperty(b, sortCol.Path, bType)
			if err != nil {
				sortErr = err
				continue
			}
			av := ap.GolangValue().(result.DateTime).Date
			bv := bp.GolangValue().(result.DateTime).Date

			// In the future when we have an implementation of dateTime comparison without precision we should swap to using that.
			// TODO(b/308012659): Implement dateTime comparison that doesn't take a precision.
			if av.Equal(bv) {
				continue
			} else if sortCol.SortByItem.Direction == model.DESCENDING {
				return bv.Compare(av)
			}
			return av.Compare(bv)
		}
		// All columns evaluated to equal so this sort is undefined.
		return 0
	})
	return sortErr
}
