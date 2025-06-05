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

package interpreter

import (
	"fmt"

	"github.com/google/cql/model"
	"github.com/google/cql/result"
	"github.com/google/cql/types"
)

// LIST OPERATORS - https://cql.hl7.org/09-b-cqlreference.html#list-operators-2

// exists(argument List<T>) Boolean
// https://cql.hl7.org/09-b-cqlreference.html#exists
func evalExists(m model.IUnaryExpression, listObj result.Value) (result.Value, error) {
	if result.IsNull(listObj) {
		return result.New(false)
	}
	list, err := result.ToSlice(listObj)
	if err != nil {
		return result.Value{}, err
	}

	if len(list) == 0 {
		return result.New(false)
	}

	for _, elemObj := range list {
		if result.IsNull(elemObj) {
			return result.New(false)
		}
	}

	return result.New(true)
}

// except(argument List<T>, argument List<T>) List<T>
// https://cql.hl7.org/09-b-cqlreference.html#except-1
func evalExcept(m model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) {
		return result.New(nil)
	}
	l, err := result.ToSlice(lObj)
	if err != nil {
		return result.Value{}, err
	}
	// If the right value is null treat it as an empty list.
	var r []result.Value
	if !result.IsNull(rObj) {
		r, err = result.ToSlice(rObj)
		if err != nil {
			return result.Value{}, err
		}
	}
	// create a list of the elements that are in the first list and are not in the second list using
	// the equality operator where each element in the result must be unique.
	var exceptList []result.Value
	for _, elemObj := range l {
		if valueInList(elemObj, r) {
			continue
		}
		if valueInList(elemObj, exceptList) {
			continue
		}
		exceptList = append(exceptList, elemObj)
	}
	return result.New(result.List{
		Value:      exceptList,
		StaticType: lObj.GolangValue().(result.List).StaticType,
	})
}

// flatten(argument List<List<T>>) List<T>
// https://cql.hl7.org/09-b-cqlreference.html#flatten
func evalFlatten(m model.IUnaryExpression, listObj result.Value) (result.Value, error) {
	if result.IsNull(listObj) {
		return result.New(nil)
	}
	list, err := result.ToSlice(listObj)
	if err != nil {
		return result.Value{}, err
	}
	var flattenedList []result.Value
	for _, elemObj := range list {
		if result.IsNull(elemObj) {
			continue
		}
		elemList, err := result.ToSlice(elemObj)
		if err != nil {
			return result.Value{}, err
		}
		flattenedList = append(flattenedList, elemList...)
	}
	return result.New(result.List{
		Value:      flattenedList,
		StaticType: listObj.GolangValue().(result.List).StaticType.ElementType.(*types.List),
	})
}

// in(element T, argument List<T>) Boolean
// https://cql.hl7.org/09-b-cqlreference.html#in-1
func evalInList(m model.IBinaryExpression, lObj, listObj result.Value) (result.Value, error) {
	if result.IsNull(listObj) {
		return result.New(nil)
	}
	r, err := result.ToSlice(listObj)
	if err != nil {
		return result.Value{}, err
	}

	return result.New(valueInList(lObj, r))
}

// included in(element T, argument List<T>) Boolean
// https://cql.hl7.org/09-b-cqlreference.html#included-in-1
// This operator acts as a macro for `in` except in cases where arguments are null.
func evalIncludedIn(m model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) {
		return result.New(nil)
	}
	if result.IsNull(rObj) {
		return result.New(false)
	}
	return evalInList(m, lObj, rObj)
}

// included in(argument List<T>, argument List<T>)
// https://cql.hl7.org/09-b-cqlreference.html#included-in-1
func evalIncludedInList(m model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) || result.IsNull(rObj) {
		return result.New(nil)
	}
	l, err := result.ToSlice(lObj)
	if err != nil {
		return result.Value{}, err
	}
	if len(l) == 0 {
		return result.New(true)
	}

	r, err := result.ToSlice(rObj)
	if err != nil {
		return result.Value{}, err
	}
	for _, elemObj := range l {
		if !valueInList(elemObj, r) {
			return result.New(false)
		}
	}
	return result.New(true)
}

// includes(argument List<T>, element T) Boolean
// https://cql.hl7.org/09-b-cqlreference.html#includes-1
func evalIncludes(m model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) || result.IsNull(rObj) {
		return result.New(nil)
	}
	l, err := result.ToSlice(lObj)
	if err != nil {
		return result.Value{}, err
	}

	return result.New(valueInList(rObj, l))
}

// includes(argument List<T>, argument List<T>) Boolean
// https://cql.hl7.org/09-b-cqlreference.html#includes-1
func evalIncludesList(m model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) || result.IsNull(rObj) {
		return result.New(nil)
	}
	l, err := result.ToSlice(lObj)
	if err != nil {
		return result.Value{}, err
	}
	r, err := result.ToSlice(rObj)
	if err != nil {
		return result.Value{}, err
	}
	for _, elemObj := range r {
		if !valueInList(elemObj, l) {
			return result.New(false)
		}
	}
	return result.New(true)
}

// intersect(argument List<T>, argument List<T>) List<T>
// https://cql.hl7.org/09-b-cqlreference.html#intersect
func evalIntersect(m model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) || result.IsNull(rObj) {
		return result.New(nil)
	}
	l, err := result.ToSlice(lObj)
	if err != nil {
		return result.Value{}, err
	}
	r, err := result.ToSlice(rObj)
	if err != nil {
		return result.Value{}, err
	}
	// create a list of the intersection of the two lists using the equality operator where
	// each element in the result must be unique.
	var intersection []result.Value
	for _, elemObj := range l {
		if !valueInList(elemObj, r) {
			continue
		}
		if valueInList(elemObj, intersection) {
			continue
		}
		intersection = append(intersection, elemObj)
	}
	return result.New(result.List{
		Value:      intersection,
		StaticType: lObj.GolangValue().(result.List).StaticType,
	})
}

// Distinct(argument List<T>) List<T>
// https://cql.hl7.org/09-b-cqlreference.html#distinct
// In the future we should make result.Value hashable so we can use a map instead of a list to
// check for duplicates.
func evalDistinct(m model.IUnaryExpression, listObj result.Value) (result.Value, error) {
	if result.IsNull(listObj) {
		return result.New(nil)
	}
	list, err := result.ToSlice(listObj)
	if err != nil {
		return result.Value{}, err
	}

	var distinctList []result.Value
	for _, elemObj := range list {
		if !valueInList(elemObj, distinctList) {
			distinctList = append(distinctList, elemObj)
		}
	}
	return result.New(result.List{
		Value:      distinctList,
		StaticType: listObj.GolangValue().(result.List).StaticType,
	})
}

// First(argument List<T>) T
// https://cql.hl7.org/09-b-cqlreference.html#first
func evalFirst(m model.IUnaryExpression, listObj result.Value) (result.Value, error) {
	if result.IsNull(listObj) {
		return result.New(nil)
	}
	list, err := result.ToSlice(listObj)
	if err != nil {
		return result.Value{}, err
	}

	if len(list) == 0 {
		return result.New(nil)
	}
	return list[0], nil
}

// Last(argument List<T>) T
// https://cql.hl7.org/09-b-cqlreference.html#last
func evalLast(m model.IUnaryExpression, listObj result.Value) (result.Value, error) {
	if result.IsNull(listObj) {
		return result.New(nil)
	}
	list, err := result.ToSlice(listObj)
	if err != nil {
		return result.Value{}, err
	}

	if len(list) == 0 {
		return result.New(nil)
	}
	return list[len(list)-1], nil
}

// IndexOf(argument List<T>, element T) Integer
// https://cql.hl7.org/09-b-cqlreference.html#indexof
func evalIndexOf(m model.IBinaryExpression, listObj, valueObj result.Value) (result.Value, error) {
	if result.IsNull(listObj) || result.IsNull(valueObj) {
		return result.New(nil)
	}
	list, err := result.ToSlice(listObj)
	if err != nil {
		return result.Value{}, err
	}
	if len(list) == 0 {
		return result.New(int32(-1))
	}

	for i, elemObj := range list {
		if valueObj.Equal(elemObj) {
			return result.New(int32(i))
		}
	}
	return result.New(int32(-1))
}

// Length(argument List<T>) Integer
// https://cql.hl7.org/09-b-cqlreference.html#length-1
func evalLengthList(m model.IUnaryExpression, listObj result.Value) (result.Value, error) {
	if result.IsNull(listObj) {
		return result.New(int32(0))
	}
	list, err := result.ToSlice(listObj)
	if err != nil {
		return result.Value{}, err
	}
	return result.New(int32(len(list)))
}

// properly includes(element Targument List<T>, element T) Boolean
// https://cql.hl7.org/09-b-cqlreference.html#properly-includes-1
func evalProperlyIncludes(m model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) {
		return result.New(nil)
	}
	l, err := result.ToSlice(lObj)
	if err != nil {
		return result.Value{}, err
	}

	if len(l) <= 1 {
		return result.New(false)
	}
	return result.New(valueInList(rObj, l))
}

// properly includes(argument List<T>, argument List<T>) Boolean
// https://cql.hl7.org/09-b-cqlreference.html#properly-includes-1
func evalProperlyIncludesList(m model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) || result.IsNull(rObj) {
		return result.New(nil)
	}
	l, err := result.ToSlice(lObj)
	if err != nil {
		return result.Value{}, err
	}
	r, err := result.ToSlice(rObj)
	if err != nil {
		return result.Value{}, err
	}
	if len(l) <= len(r) {
		return result.New(false)
	}
	for _, elemObj := range r {
		if !valueInList(elemObj, l) {
			return result.New(false)
		}
	}
	return result.New(true)
}

// properly included in(element T, argument List<T>) Boolean
// https://cql.hl7.org/09-b-cqlreference.html#properly-included-in-1
// Note: The docs are a bit ambiguous on this point, but the point type for this operator has the
// order of the arguments reversed from properly includes for the same overload.
func evalProperlyIncludedIn(m model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	return evalProperlyIncludes(m, rObj, lObj)
}

// properly included in(argument List<T>, argument List<T>) Boolean
// https://cql.hl7.org/09-b-cqlreference.html#properly-included-in-1
// Note: Similar to the point overload, the order of the arguments is reversed from properly
// includes for the same overload.
func evalProperlyIncludedInList(m model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	return evalProperlyIncludesList(m, rObj, lObj)
}

// singleton from(argument List<T>) T
// https://cql.hl7.org/09-b-cqlreference.html#singleton-from
func evalSingletonFrom(m model.IUnaryExpression, listObj result.Value) (result.Value, error) {
	if result.IsNull(listObj) {
		return result.New(nil)
	}

	list, err := result.ToSlice(listObj)
	if err != nil {
		return result.Value{}, err
	}

	switch len(list) {
	case 0:
		return result.New(nil)
	case 1:
		return list[0], nil
	default:
		return result.Value{}, fmt.Errorf("singleton from requires a list of length 0 or 1, but got length %d", len(list))
	}
}

// Skip(argument List<T>, index Integer) List<T>
// https://cql.hl7.org/09-b-cqlreference.html#skip
func evalSkip(m model.IBinaryExpression, listObj, indexObj result.Value) (result.Value, error) {
	if result.IsNull(listObj) {
		return result.New(nil)
	}
	if result.IsNull(indexObj) {
		return result.New(nil)
	}
	list, err := result.ToSlice(listObj)
	if err != nil {
		return result.Value{}, err
	}
	index, err := result.ToInt32(indexObj)
	if err != nil {
		return result.Value{}, err
	}

	staticType := listObj.GolangValue().(result.List).StaticType
	// If the index is out of bounds, return an empty list.
	if index < 0 || index >= int32(len(list)) {
		return result.New(result.List{StaticType: staticType})
	}
	return result.New(result.List{Value: list[index:], StaticType: staticType})
}

// Tail(argument List<T>) List<T>
// https://cql.hl7.org/09-b-cqlreference.html#tail
func evalTail(m model.IUnaryExpression, listObj result.Value) (result.Value, error) {
	if result.IsNull(listObj) {
		return result.New(nil)
	}
	list, err := result.ToSlice(listObj)
	if err != nil {
		return result.Value{}, err
	}
	staticType := listObj.GolangValue().(result.List).StaticType
	// If the list is empty, return an empty list.
	if len(list) == 0 {
		return result.New(result.List{StaticType: staticType})
	}
	return result.New(result.List{Value: list[1:], StaticType: staticType})
}

// Take(argument List<T>, number Integer) List<T>
// https://cql.hl7.org/09-b-cqlreference.html#take
// Removes the last n elements from the list.
func evalTake(m model.IBinaryExpression, listObj, numberObj result.Value) (result.Value, error) {
	if result.IsNull(listObj) {
		return result.New(nil)
	}
	list, err := result.ToSlice(listObj)
	if err != nil {
		return result.Value{}, err
	}
	staticType := listObj.GolangValue().(result.List).StaticType
	if result.IsNull(numberObj) {
		return result.New(result.List{StaticType: staticType})
	}
	number, err := result.ToInt32(numberObj)
	if err != nil {
		return result.Value{}, err
	}
	if number <= 0 {
		return result.New(result.List{StaticType: staticType})
	}
	if number > int32(len(list)) {
		return result.New(result.List{Value: list, StaticType: staticType})
	}
	return result.New(result.List{Value: list[:number], StaticType: staticType})
}

// Indexer(argument List<T>, index Integer) T
// [](argument List<T>, index Integer) T
// https://cql.hl7.org/09-b-cqlreference.html#indexer-1
// Indexer is also defined for String, see operator_string.go for that implementation.
func (i *interpreter) evalIndexerList(m model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) || result.IsNull(rObj) {
		return result.New(nil)
	}
	list, err := result.ToSlice(lObj)
	if err != nil {
		return result.Value{}, err
	}
	idx, err := result.ToInt32(rObj)
	if err != nil {
		return result.Value{}, err
	}
	if idx < 0 || idx >= int32(len(list)) {
		return result.New(nil)
	}
	return list[idx], nil
}

// Union(left List<T>, right List<T>) List<T>
// https://cql.hl7.org/09-b-cqlreference.html#union-1
func evalUnion(m model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	staticType := &types.List{ElementType: types.Any}
	var l []result.Value
	var r []result.Value
	var err error
	if !result.IsNull(lObj) {
		staticType = lObj.GolangValue().(result.List).StaticType
		l, err = result.ToSlice(lObj)
		if err != nil {
			return result.Value{}, err
		}
	}
	if !result.IsNull(rObj) {
		staticType = rObj.GolangValue().(result.List).StaticType
		r, err = result.ToSlice(rObj)
		if err != nil {
			return result.Value{}, err
		}
	}
	var unionList []result.Value
	for _, elemObj := range l {
		unionList = append(unionList, elemObj)
	}
	for _, elemObj := range r {
		if !valueInList(elemObj, l) {
			unionList = append(unionList, elemObj)
		}
	}
	return result.New(result.List{
		Value:      unionList,
		StaticType: staticType,
	})
}

// valueInList returns true if the value is in the list using equality scemantics.
func valueInList(value result.Value, list []result.Value) bool {
	for _, elemObj := range list {
		if value.Equal(elemObj) {
			return true
		}
	}
	return false
}
