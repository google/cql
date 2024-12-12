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

// valueInList returns true if the value is in the list using equality scemantics.
func valueInList(value result.Value, list []result.Value) bool {
	for _, elemObj := range list {
		if value.Equal(elemObj) {
			return true
		}
	}
	return false
}
