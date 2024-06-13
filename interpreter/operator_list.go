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

	for _, elemObj := range r {
		if lObj.Equal(elemObj) {
			return result.New(true)
		}
	}
	return result.New(false)
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
