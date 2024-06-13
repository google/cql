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

// evalIfThenElse handles evaluation and overload matching for the if then else statement.
// When the `if` condition is or evaluates to null, the else condition should be returned.
// https://cql.hl7.org/03-developersguide.html#conditional-expressions
func (i *interpreter) evalIfThenElse(ite *model.IfThenElse) (result.Value, error) {
	cndObj, err := i.evalExpression(ite.Condition)
	if err != nil {
		return result.Value{}, err
	}

	if result.IsNull(cndObj) {
		return i.evalExpression(ite.Else)
	}
	condResult, err := result.ToBool(cndObj)
	if err != nil {
		return result.Value{}, err
	}

	if condResult {
		return i.evalExpression(ite.Then)
	}
	return i.evalExpression(ite.Else)
}

// evalCase handles case expressions.
// https://cql.hl7.org/03-developersguide.html#conditional-expressions
func (i *interpreter) evalCase(c *model.Case) (result.Value, error) {
	if c.Comparand == nil {
		// No Comparand - CaseItem.Whens are booleans.
		for _, caseItem := range c.CaseItem {
			whenObj, err := i.evalExpression(caseItem.When)
			if err != nil {
				return result.Value{}, err
			}
			if result.IsNull(whenObj) {
				// Null is treated as false.
				continue
			}
			when, err := result.ToBool(whenObj)
			if err != nil {
				return result.Value{}, err
			}
			if when {
				return i.evalExpression(caseItem.Then)
			}
		}
		return i.evalExpression(c.Else)
	}

	// Comparand - evaluate Equal(Comparand, CaseItem.When).
	comparandObj, err := i.evalExpression(c.Comparand)
	if err != nil {
		return result.Value{}, err
	}
	if result.IsNull(comparandObj) {
		return i.evalExpression(c.Else)
	}

	for _, caseItem := range c.CaseItem {
		whenObj, err := i.evalExpression(caseItem.When)
		if err != nil {
			return result.Value{}, err
		}
		if result.IsNull(whenObj) {
			continue
		}
		if !comparandObj.RuntimeType().Equal(whenObj.RuntimeType()) {
			return result.Value{}, fmt.Errorf("internal error - in case expressions the comparand and case must be the same type, got comparand: %v case: %v", comparandObj.RuntimeType(), whenObj.RuntimeType())
		}
		eqObj, err := i.evalEqual(nil, comparandObj, whenObj)
		if err != nil {
			return result.Value{}, err
		}
		eq, err := result.ToBool(eqObj)
		if err != nil {
			return result.Value{}, err
		}
		if eq {
			return i.evalExpression(caseItem.Then)
		}
	}
	return i.evalExpression(c.Else)
}
