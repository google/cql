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
	"regexp"
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

// Split(stringToSplit String, separator String) List<String>
// https://cql.hl7.org/09-b-cqlreference.html#split
func (i *interpreter) evalSplit(m model.IBinaryExpression, left, right result.Value) (result.Value, error) {
	if result.IsNull(left) {
		return result.New(nil)
	}
	if result.IsNull(right) {
		return result.New(result.List{Value: []result.Value{left}, StaticType: &types.List{ElementType: types.String}})
	}

	s, sep, err := applyToValues(left, right, result.ToString)
	if err != nil {
		return result.Value{}, err
	}

	splits := strings.Split(s, sep)

	l := result.List{Value: []result.Value{}, StaticType: &types.List{ElementType: types.String}}
	for _, split := range splits {
		item, err := result.New(split)
		if err != nil {
			return result.Value{}, err
		}
		l.Value = append(l.Value, item)
	}
	return result.New(l)
}

// Combine(source List<String>) String
// Combine(source List<String>, separator String) String
// https://cql.hl7.org/09-b-cqlreference.html#combine
func (i *interpreter) evalCombine(m model.INaryExpression, operands []result.Value) (result.Value, error) {
	if len(operands) == 0 {
		// Dispatcher and Parser should prevent this from happening.
		return result.Value{}, fmt.Errorf("internal error - Combine must have at least one operand")
	}

	if result.IsNull(operands[0]) {
		return result.New(nil)
	}

	// Extract string list:
	strList, err := result.ToSlice(operands[0])
	if err != nil {
		return result.Value{}, err
	}
	if len(strList) == 0 {
		return result.New(nil)
	}

	// Extract separator:
	sep := ""
	if len(operands) == 2 {
		if result.IsNull(operands[1]) {
			return result.New(nil)
		}
		sep, err = result.ToString(operands[1])
		if err != nil {
			return result.Value{}, err
		}
	}

	resultBuilder := strings.Builder{}
	for idx, str := range strList {
		if result.IsNull(str) {
			continue
		}
		s, err := result.ToString(str)
		if err != nil {
			return result.Value{}, err
		}
		resultBuilder.WriteString(s)

		// Write the separator unless we are on the last element.
		if idx < len(strList)-1 {
			resultBuilder.WriteString(sep)
		}
	}
	return result.New(resultBuilder.String())
}

// Indexer(argument String, index Integer) String
// https://cql.hl7.org/09-b-cqlreference.html#indexer
// Indexer is also defined for List<T>, see operator_list.go for that implementation.
func (i *interpreter) evalIndexerString(m model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) || result.IsNull(rObj) {
		return result.New(nil)
	}
	str, err := result.ToString(lObj)
	if err != nil {
		return result.Value{}, err
	}
	idx, err := result.ToInt32(rObj)
	if err != nil {
		return result.Value{}, err
	}
	if idx < 0 || idx >= int32(len(str)) {
		return result.New(nil)
	}
	return result.New(string([]rune(str)[idx]))
}

// convert a quantity value to a string
func quantityToString(q result.Quantity) string {
	f := strconv.FormatFloat(q.Value, 'f', -1, 64)
	return fmt.Sprintf("%s '%s'", f, q.Unit)
}

// EndsWith(argument String, suffix String) Boolean
// https://cql.hl7.org/2018May/09-b-cqlreference.html#endswith
func evalEndsWith(m model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) || result.IsNull(rObj) {
		return result.New(nil)
	}
	str, err := result.ToString(lObj)
	if err != nil {
		return result.Value{}, err
	}
	suffix, err := result.ToString(rObj)
	if err != nil {
		return result.Value{}, err
	}
	if int32(len(suffix)) == 0 {
		return result.New(true)
	}
	return result.New(strings.HasSuffix(str, suffix))
}

// LastPositionOf(pattern String, argument String) Integer
// https://cql.hl7.org/09-b-cqlreference.html#lastpositionof
func evalLastPositionOf(m model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) || result.IsNull(rObj) {
		return result.New(nil)
	}
	pattern, err := result.ToString(lObj)
	if err != nil {
		return result.Value{}, err
	}
	argument, err := result.ToString(rObj)
	if err != nil {
		return result.Value{}, err
	}
	return result.New(strings.LastIndex(argument, pattern))
}

// Length(argument String) Integer
// https://cql.hl7.org/09-b-cqlreference.html#length
func evalLengthString(m model.IUnaryExpression, stringObj result.Value) (result.Value, error) {
	if result.IsNull(stringObj) {
		return result.New(nil)
	}
	string, err := result.ToString(stringObj)
	if err != nil {
		return result.Value{}, err
	}

	return result.New(int32(len(string)))
}

// Upper(argument String) String
// https://cql.hl7.org/09-b-cqlreference.html#Upper
func evalUpper(m model.IUnaryExpression, strObj result.Value) (result.Value, error) {
	if result.IsNull(strObj) {
		return result.New(nil)
	}
	str, err := result.ToString(strObj)
	if err != nil {
		return result.Value{}, err
	}
	return result.New(strings.ToUpper(str))
}

// Lower(argument String) String
// https://cql.hl7.org/09-b-cqlreference.html#lower
func evalLower(m model.IUnaryExpression, stringObj result.Value) (result.Value, error) {
	if result.IsNull(stringObj) {
		return result.New(nil)
	}
	str, err := result.ToString(stringObj)
	if err != nil {
		return result.Value{}, err
	}
	return result.New(strings.ToLower(str))
}

// PositionOf(pattern String, argument String) Integer
// https://cql.hl7.org/09-b-cqlreference.html#positionof
func evalPositionOf(m model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) || result.IsNull(rObj) {
		return result.New(nil)
	}
	pattern, err := result.ToString(lObj)
	if err != nil {
		return result.Value{}, err
	}
	argument, err := result.ToString(rObj)
	if err != nil {
		return result.Value{}, err
	}
	return result.New(strings.Index(argument, pattern))
}

// StartsWith(argument String, prefix String) Boolean
// https://cql.hl7.org/09-b-cqlreference.html#startswith
func evalStartsWith(m model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) || result.IsNull(rObj) {
		return result.New(nil)
	}
	argument, err := result.ToString(lObj)
	if err != nil {
		return result.Value{}, err
	}
	prefix, err := result.ToString(rObj)
	if err != nil {
		return result.Value{}, err
	}
	return result.New(strings.HasPrefix(argument, prefix))
}

// ReplaceMatches(argument String, pattern String, substitution String) String
// https://cql.hl7.org/09-b-cqlreference.html#replacematches
func evalReplaceMatches(m model.INaryExpression, operands []result.Value) (result.Value, error) {

	transformSubstring := func(cqlSubst string) string {
	// This helper function checks and converts PCRE escapes to Go escapes to ensure regexp package compatibility.
		var goSubst strings.Builder
		runes := []rune(cqlSubst)
		i := 0
		for i < len(runes) {
			if runes[i] == '\\' {
				if i+1 < len(runes) {
					switch runes[i+1] {
					case '$':
						goSubst.WriteString("$$") // PCRE \$ -> Go $$
						i++
					case '\\':
						goSubst.WriteRune('\\') // PCRE \\ -> Go \
						i++
					// Add other PCRE escapes if needed (e.g., \n, \t)
					// and convert them to actual characters or Go's equivalents if any.
					// Go's Expand treats \n literally, so you'd convert to actual newline.
					default:
						goSubst.WriteRune('\\') // Keep other backslashed chars
						goSubst.WriteRune(runes[i+1])
						i++
					}
				} else {
					goSubst.WriteRune('\\') // Trailing backslash
				}
			} else if runes[i] == '$' {
				// Check if it's not followed by a character that would make it a Go literal $ (e.g. another $)
				// or if it's a group $1, ${name} which are compatible.
				// For simplicity, assuming $ not part of \$ or $$ is a group or literal $ that Go handles.
				goSubst.WriteRune('$')
			} else {
				goSubst.WriteRune(runes[i])
			}
			i++
		}
		return goSubst.String()
	}


	if result.IsNull(operands[0]) || result.IsNull(operands[1]) || result.IsNull(operands[2]) {
		return result.New(nil)
	}
	argumentStr, err := result.ToString(operands[0])
	if err != nil {
		return result.Value{}, err
	}
	patternStr, err := result.ToString(operands[1])
	if err != nil {
		return result.Value{}, err
	}
	substitutionStr, err := result.ToString(operands[2])
	if err != nil {
		return result.Value{}, err
	}
	patternRe, err := regexp.Compile(patternStr)
	if err != nil {
		return result.New(false)
	}

	substitutionStr = transformSubstring(substitutionStr)

	return result.New(patternRe.ReplaceAllString(argumentStr, substitutionStr))
}
