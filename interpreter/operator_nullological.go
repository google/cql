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

// Coalesce<T>(argument1 T, argument2 T) T
// Coalesce<T>(argument1 T, argument2 T, argument3 T) T
// Coalesce<T>(argument1 T, argument2 T, argument3 T, argument4 T) T
// Coalesce<T>(argument1 T, argument2 T, argument3 T, argument4 T, argument5 T) T
// https://cql.hl7.org/09-b-cqlreference.html#coalesce
func evalCoalesce(m model.INaryExpression, objs []result.Value) (result.Value, error) {
	for _, obj := range objs {
		if !result.IsNull(obj) {
			return obj, nil
		}
	}
	return result.New(nil)
}

// Coalesce<T>(arguments List<T>) T
// https://cql.hl7.org/09-b-cqlreference.html#coalesce
func evalCoalesceList(m model.INaryExpression, objs []result.Value) (result.Value, error) {
	if len(objs) != 1 {
		return result.Value{}, fmt.Errorf("coalesce list overload expected exactly 1 list argument, got: %v", len(objs))
	}
	list, err := result.ToSlice(objs[0])
	if err != nil {
		return result.Value{}, fmt.Errorf("coalesce list overload unable to convert argument to List, err: %w", err)
	}

	for _, obj := range list {
		if !result.IsNull(obj) {
			return obj, nil
		}
	}
	return result.New(nil)
}

// is null(argument Any) Boolean
// https://cql.hl7.org/09-b-cqlreference.html#isnull
func evalIsNull(m model.IUnaryExpression, obj result.Value) (result.Value, error) {
	return result.New(result.IsNull(obj))
}

// is true(argument Boolean) Boolean
// https://cql.hl7.org/09-b-cqlreference.html#istrue
func evalIsTrue(m model.IUnaryExpression, obj result.Value) (result.Value, error) {
	if result.IsNull(obj) {
		return result.New(false)
	}
	objVal, err := result.ToBool(obj)
	if err != nil {
		return result.Value{}, err
	}
	return result.New(objVal == true)
}

// is false(argument Boolean) Boolean
// https://cql.hl7.org/09-b-cqlreference.html#isfalse
func evalIsFalse(m model.IUnaryExpression, obj result.Value) (result.Value, error) {
	if result.IsNull(obj) {
		return result.New(false)
	}
	objVal, err := result.ToBool(obj)
	if err != nil {
		return result.Value{}, err
	}
	return result.New(objVal == false)
}
