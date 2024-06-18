// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package interpreter

import (
	"github.com/google/cql/model"
	"github.com/google/cql/result"
)

// AGGREGATE FUNCTIONS - https://cql.hl7.org/09-b-cqlreference.html#aggregate-functions

// AllTrue(argument List<Boolean>) Boolean
// https://cql.hl7.org/09-b-cqlreference.html#alltrue
func (i *interpreter) evalAllTrue(m model.IUnaryExpression, operand result.Value) (result.Value, error) {
	if result.IsNull(operand) {
		return result.New(true)
	}
	l, err := result.ToSlice(operand)
	if err != nil {
		return result.Value{}, err
	}
	for _, elem := range l {
		if result.IsNull(elem) {
			continue
		}
		bv, err := result.ToBool(elem)
		if err != nil {
			return result.Value{}, err
		}
		if !bv {
			return result.New(false)
		}
	}
	return result.New(true)
}

// AnyTrue(argument List<Boolean>) Boolean
// https://cql.hl7.org/09-b-cqlreference.html#anytrue
func (i *interpreter) evalAnyTrue(m model.IUnaryExpression, operand result.Value) (result.Value, error) {
	if result.IsNull(operand) {
		return result.New(false)
	}
	l, err := result.ToSlice(operand)
	if err != nil {
		return result.Value{}, err
	}
	for _, elem := range l {
		if result.IsNull(elem) {
			continue
		}
		bv, err := result.ToBool(elem)
		if err != nil {
			return result.Value{}, err
		}
		if bv {
			return result.New(true)
		}
	}
	return result.New(false)
}

// Count(argument List<T>) Integer
// https://cql.hl7.org/09-b-cqlreference.html#count
func (i *interpreter) evalCount(m model.IUnaryExpression, operand result.Value) (result.Value, error) {
	if result.IsNull(operand) {
		return result.New(0)
	}
	l, err := result.ToSlice(operand)
	if err != nil {
		return result.Value{}, err
	}
	count := 0
	for _, elem := range l {
		if !result.IsNull(elem) {
			count++
		}
	}
	return result.New(count)
}
