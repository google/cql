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
	"strconv"
	"strings"

	"github.com/google/cql/internal/datehelpers"
	"github.com/google/cql/model"
	"github.com/google/cql/result"
	"github.com/google/cql/types"
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

// ToString(Boolean) String
// ToString(Integer) String
// ToString(Long) String
// ToString(Decimal) String
// ToString(Quantity) String
// ToString(Ratio) String
// ToString(Date) String
// ToString(DateTime) String
// ToString(Time) String
// https://cql.hl7.org/09-b-cqlreference.html#tostring
// In the future we may put this logic directly onto the result.Value interface. Could be useful
// for debugging, and/or the REPL.
func evalToString(_ model.IUnaryExpression, operand result.Value) (result.Value, error) {
	if result.IsNull(operand) {
		return result.New(nil)
	}

	switch operand.RuntimeType() {
	case types.Boolean:
		b, err := result.ToBool(operand)
		if err != nil {
			return result.Value{}, err
		}
		return result.New(strconv.FormatBool(b))
	case types.Integer:
		i, err := result.ToInt32(operand)
		if err != nil {
			return result.Value{}, err
		}
		return result.New(strconv.FormatInt(int64(i), 10))
	case types.Long:
		i, err := result.ToInt64(operand)
		if err != nil {
			return result.Value{}, err
		}
		return result.New(strconv.FormatInt(i, 10))
	case types.Decimal:
		d, err := result.ToFloat64(operand)
		if err != nil {
			return result.Value{}, err
		}
		return result.New(strconv.FormatFloat(d, 'f', -1, 64))
	case types.Quantity:
		q, err := result.ToQuantity(operand)
		if err != nil {
			return result.Value{}, err
		}
		return result.New(quantityToString(q))
	case types.Ratio:
		r, err := result.ToRatio(operand)
		if err != nil {
			return result.Value{}, err
		}
		return result.New(fmt.Sprintf("%s:%s", quantityToString(r.Numerator), quantityToString(r.Denominator)))
	case types.Date:
		d, err := result.ToDateTime(operand)
		if err != nil {
			return result.Value{}, err
		}
		s, err := datehelpers.DateString(d.Date, d.Precision)
		if err != nil {
			return result.Value{}, err
		}
		// Remove the leading '@'
		return result.New(s[1:])
	case types.DateTime:
		d, err := result.ToDateTime(operand)
		if err != nil {
			return result.Value{}, err
		}
		s, err := datehelpers.DateTimeString(d.Date, d.Precision)
		if err != nil {
			return result.Value{}, err
		}
		// Remove the leading '@'
		return result.New(s[1:])
	case types.Time:
		t, err := result.ToDateTime(operand)
		if err != nil {
			return result.Value{}, err
		}
		s, err := datehelpers.TimeString(t.Date, t.Precision)
		if err != nil {
			return result.Value{}, err
		}
		// Remove the leading 'T'
		return result.New(s[1:])
	default:
		return result.Value{}, fmt.Errorf("unsupported operand type for ToString: %v", operand.RuntimeType())
	}
}

// convert a quantity value to a string
func quantityToString(q result.Quantity) string {
	f := strconv.FormatFloat(q.Value, 'f', -1, 64)
	return fmt.Sprintf("%s '%s'", f, q.Unit)
}
