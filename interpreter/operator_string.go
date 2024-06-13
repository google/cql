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
	"strings"

	"github.com/google/cql/model"
	"github.com/google/cql/result"
)

// +(left String, right String) String
// &(left String, right String) String
// https://cql.hl7.org/09-b-cqlreference.html#concatenate
func evalConcatenate(m model.INaryExpression, operands []result.Value) (result.Value, error) {
	retStr := strings.Builder{}
	for _, operand := range operands {
		if result.IsNull(operand) {
			// Return null if any operand is null.
			// If the & operator is used, nulls must be treated as empty strings. The parser handles this
			// by inserting a Coalesce operation with an empty string over the operands.
			return result.New(nil)
		}
		opStr, err := result.ToString(operand)
		if err != nil {
			return result.Value{}, err
		}
		retStr.WriteString(opStr)
	}
	return result.New(retStr.String())
}
