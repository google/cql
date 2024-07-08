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
	"cmp"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/google/cql/internal/convert"
	"github.com/google/cql/model"
	"github.com/google/cql/result"
	"github.com/google/cql/types"
)

// COMPARISON OPERATORS - https://cql.hl7.org/09-b-cqlreference.html#comparison-operators-4

// =<T>(left T, right T) Boolean
// https://cql.hl7.org/09-b-cqlreference.html#equal
func (i *interpreter) evalEqual(_ model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	// TODO: b/327612471 - Revisit Equal to make sure it is correctly implemented for all types.
	if result.IsNull(lObj) || result.IsNull(rObj) {
		return result.New(nil)
	}
	return result.New(lObj.Equal(rObj))
}

// =(left DateTime, right DateTime) Boolean
// =(left Date, right Date) Boolean
// https://cql.hl7.org/09-b-cqlreference.html#equal
func evalEqualDateTime(_ model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) || result.IsNull(rObj) {
		return result.New(nil)
	}
	lVal, rVal, err := applyToValues(lObj, rObj, result.ToDateTime)
	if err != nil {
		return result.Value{}, err
	}
	comp, err := compareDateTime(lVal, rVal)
	if err != nil {
		return result.Value{}, err
	}
	switch comp {
	case leftEqualRight:
		return result.New(true)
	case insufficientPrecision:
		return result.New(nil)
	default:
		return result.New(false)
	}
}

// ~(left Boolean, right Boolean) Boolean
// ~(left Integer, right Integer) Boolean
// ~(left Long, right Long) Boolean
// All equivalent overloads should be resilient to a nil model.
// https://cql.hl7.org/09-b-cqlreference.html#equivalent
func evalEquivalentSimpleType(_ model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	// TODO(b/301606416): Revisit Equivalent to make sure it is correctly implemented for all types.
	if result.IsNull(lObj) && result.IsNull(rObj) {
		return result.New(true)
	}
	return result.New(lObj.Equal(rObj))
}

// evalEquivalentValue applies the CQL equivalent operator to the passed Values.
func (i *interpreter) evalEquivalentValue(lObj, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) && result.IsNull(rObj) {
		return result.New(true)
	}
	if result.IsNull(lObj) != result.IsNull(rObj) {
		return result.New(false)
	}
	overloads, err := i.binaryOverloads(&model.Equivalent{})
	if err != nil {
		return result.Value{}, err
	}
	// We are able to use RuntimeType instead of static type since all null cases are handled above,
	// and there are no Choice type overloads defined.
	opTypes := []types.IType{lObj.RuntimeType(), rObj.RuntimeType()}
	innerEquivalentFunc, err := convert.ExactOverloadMatch(opTypes, overloads, i.modelInfo, "Equivalent")
	if err != nil {
		return result.Value{}, err
	}
	// All equivalent overloads should be resilient to a nil model.
	return innerEquivalentFunc(nil, lObj, rObj)
}

// equivalentGolang attempts to apply the CQL equivalent operator to the passed Golang values,
// and returns a Golang bool.
// It first attempts to convert them to result.Value by calling result.New, then calls
// evalEquivalentValue.
func (i *interpreter) equivalentGolang(lObj, rObj any) (bool, error) {
	lVal, rVal, err := convertToValues(lObj, rObj)
	if err != nil {
		return false, err
	}
	equi, err := i.evalEquivalentValue(lVal, rVal)
	if err != nil {
		return false, err
	}
	equiBool, err := result.ToBool(equi)
	if err != nil {
		return false, err
	}
	return equiBool, nil
}

func convertToValues(l, r any) (result.Value, result.Value, error) {
	lVal, err := result.New(l)
	if err != nil {
		return result.Value{}, result.Value{}, err
	}
	rVal, err := result.New(r)
	if err != nil {
		return result.Value{}, result.Value{}, err
	}
	return lVal, rVal, nil
}

// ~(left List<T>, right List<T>) Boolean
// All equivalent overloads should be resilient to a nil model.
// https://cql.hl7.org/09-b-cqlreference.html#equivalent-2
func (i *interpreter) evalEquivalentList(_ model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) && result.IsNull(rObj) {
		return result.New(true)
	}
	if result.IsNull(lObj) != result.IsNull(rObj) {
		return result.New(false)
	}
	lList, err := result.ToSlice(lObj)
	if err != nil {
		return result.Value{}, err
	}
	rList, err := result.ToSlice(rObj)
	if err != nil {
		return result.Value{}, err
	}

	if len(lList) != len(rList) {
		return result.New(false)
	}

	// TODO: b/301606416 - For a non-mixed list, one optimization could be to compute the equivalent
	// overload to call once instead of computing it every time inside evalEquivalentValue.
	for idx := range lList {
		// TODO: b/326277425 - Properly support mixed lists. For mixed lists (including List<Any> with
		// mixed element types), the parser may not be able to apply all valid implicit conversions so
		// we may need to consider applying them in the interpreter. For now, an element comparison with
		// types that are not exact matches or subtypes will result in an Equivalent overload matching
		// error. In the future we may consider checking if elements are convertible and returning false
		// instead of an error, but we can address this in the future.
		equi, err := i.evalEquivalentValue(lList[idx], rList[idx])
		if errors.Is(err, convert.ErrNoMatch) {
			return result.Value{}, fmt.Errorf("unable to match Equivalent overload for elements in a list, this is likely because our engine does not fully support mixed type lists yet: %w", err)
		}
		if err != nil {
			return result.Value{}, err
		}
		// All equivalent overloads should return true or false, so we expect a non-null boolean here.
		isEquivalent, err := result.ToBool(equi)
		if err != nil {
			return result.Value{}, err
		}
		if !isEquivalent {
			return result.New(false)
		}
	}
	return result.New(true)
}

// ~(left String, right String) Boolean
// https://cql.hl7.org/09-b-cqlreference.html#equivalent
func evalEquivalentString(_ model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) && result.IsNull(rObj) {
		return result.New(true)
	}
	if result.IsNull(lObj) != result.IsNull(rObj) {
		return result.New(false)
	}
	lStr, rStr, err := applyToValues(lObj, rObj, result.ToString)
	if err != nil {
		return result.Value{}, err
	}
	return result.New(equivalentString(lStr) == equivalentString(rStr))
}

// equivalentString converts all characters to lowercase, and normalizes all whitespace to a single
// space for equivalent string comparison.
func equivalentString(input string) string {
	var out strings.Builder
	for _, elem := range input {
		if unicode.IsSpace(elem) {
			out.WriteString(" ")
		} else {
			out.WriteRune(unicode.ToLower(elem))
		}
	}
	return out.String()
}

// ~(left Interval<T>, right Interval<T>) Boolean
// https://cql.hl7.org/09-b-cqlreference.html#equivalent-1
func (i *interpreter) evalEquivalentInterval(_ model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) && result.IsNull(rObj) {
		return result.New(true)
	}
	if result.IsNull(lObj) != result.IsNull(rObj) {
		return result.New(false)
	}

	// Check to see if start and end points of the interval are equivalent.
	startL, err := start(lObj, &i.evaluationTimestamp)
	if err != nil {
		return result.Value{}, err
	}
	startR, err := start(rObj, &i.evaluationTimestamp)
	if err != nil {
		return result.Value{}, err
	}
	endL, err := end(lObj, &i.evaluationTimestamp)
	if err != nil {
		return result.Value{}, err
	}
	endR, err := end(rObj, &i.evaluationTimestamp)
	if err != nil {
		return result.Value{}, err
	}

	startEqui, err := i.evalEquivalentValue(startL, startR)
	if err != nil {
		return result.Value{}, err
	}
	endEqui, err := i.evalEquivalentValue(endL, endR)
	if err != nil {
		return result.Value{}, err
	}

	// All equivalent overloads should return true or false, so we expect non-null booleans here.
	startEquiBool, endEquiBool, err := applyToValues(startEqui, endEqui, result.ToBool)
	if err != nil {
		return result.Value{}, err
	}
	if startEquiBool && endEquiBool {
		return result.New(true)
	}
	return result.New(false)
}

// ~(left DateTime, right DateTime) Boolean
// ~(left Date, right Date) Boolean
// All equivalent overloads should be resilient to a nil model.
// https://cql.hl7.org/09-b-cqlreference.html#equivalent
func evalEquivalentDateTime(_ model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) && result.IsNull(rObj) {
		return result.New(true)
	}
	if result.IsNull(lObj) != result.IsNull(rObj) {
		return result.New(false)
	}
	lVal, rVal, err := applyToValues(lObj, rObj, result.ToDateTime)
	if err != nil {
		return result.Value{}, err
	}
	comp, err := compareDateTime(lVal, rVal)
	if err != nil {
		return result.Value{}, err
	}
	switch comp {
	case leftEqualRight:
		return result.New(true)
	case insufficientPrecision:
		return result.New(false)
	default:
		return result.New(false)
	}
}

// ~(left Concept, right Code) Boolean
// https://cql.hl7.org/09-b-cqlreference.html#equivalent-3
// Some Equivalent overloads are categorized in the clinical operator section, like this one, but
// are included in operator_comparison.go to keep all equivalent overloads together.
func (i *interpreter) evalEquivalentConceptCode(b model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) && result.IsNull(rObj) {
		return result.New(true)
	}
	if result.IsNull(lObj) != result.IsNull(rObj) {
		return result.New(false)
	}

	con, err := result.ToConcept(lObj)
	if err != nil {
		return result.Value{}, err
	}

	// Sanity check right hand type.
	_, err = result.ToCode(rObj)
	if err != nil {
		return result.Value{}, err
	}

	for _, conCode := range con.Codes {
		conCodeObj, err := result.New(conCode)
		if err != nil {
			return result.Value{}, err
		}
		equi, err := i.evalEquivalentValue(conCodeObj, rObj)
		if err != nil {
			return result.Value{}, err
		}
		equiBool, err := result.ToBool(equi)
		if err != nil {
			return result.Value{}, err
		}
		if equiBool {
			return result.New(true)
		}
	}
	return result.New(false)
}

func (i *interpreter) evalEquivalentCodeCode(b model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) && result.IsNull(rObj) {
		return result.New(true)
	}
	if result.IsNull(lObj) != result.IsNull(rObj) {
		return result.New(false)
	}

	lCode, rCode, err := applyToValues(lObj, rObj, result.ToCode)
	if err != nil {
		return result.Value{}, err
	}

	// Codes are equivalent if the system and codes are equivalent.
	codesEqui, err := i.equivalentGolang(lCode.Code, rCode.Code)
	if err != nil {
		return result.Value{}, err
	}
	if !codesEqui {
		return result.New(false)
	}

	systemsEqui, err := i.equivalentGolang(lCode.System, rCode.System)
	if err != nil {
		return result.Value{}, err
	}
	if !systemsEqui {
		return result.New(false)
	}
	return result.New(true)
}

// op(left Integer, right Integer) Boolean
// https://cql.hl7.org/09-b-cqlreference.html#less
// https://cql.hl7.org/09-b-cqlreference.html#less-or-equal
// https://cql.hl7.org/09-b-cqlreference.html#greater
// https://cql.hl7.org/09-b-cqlreference.html#greater-or-equal
func evalCompareInteger(m model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) || result.IsNull(rObj) {
		return result.New(nil)
	}
	l, r, err := applyToValues(lObj, rObj, result.ToInt32)
	if err != nil {
		return result.Value{}, err
	}
	return compare(m, l, r)
}

// op(left Long, right Long) Boolean
// https://cql.hl7.org/09-b-cqlreference.html#less
// https://cql.hl7.org/09-b-cqlreference.html#less-or-equal
// https://cql.hl7.org/09-b-cqlreference.html#greater
// https://cql.hl7.org/09-b-cqlreference.html#greater-or-equal
func evalCompareLong(m model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) || result.IsNull(rObj) {
		return result.New(nil)
	}
	l, r, err := applyToValues(lObj, rObj, result.ToInt64)
	if err != nil {
		return result.Value{}, err
	}
	return compare(m, l, r)
}

// op(left Decimal, right Decimal) Boolean
// https://cql.hl7.org/09-b-cqlreference.html#less
// https://cql.hl7.org/09-b-cqlreference.html#less-or-equal
// https://cql.hl7.org/09-b-cqlreference.html#greater
// https://cql.hl7.org/09-b-cqlreference.html#greater-or-equal
func evalCompareDecimal(m model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) || result.IsNull(rObj) {
		return result.New(nil)
	}
	l, r, err := applyToValues(lObj, rObj, result.ToFloat64)
	if err != nil {
		return result.Value{}, err
	}
	return compare(m, l, r)
}

// op(left String, right String) Boolean
// https://cql.hl7.org/09-b-cqlreference.html#less
// https://cql.hl7.org/09-b-cqlreference.html#less-or-equal
// https://cql.hl7.org/09-b-cqlreference.html#greater
// https://cql.hl7.org/09-b-cqlreference.html#greater-or-equal
func evalCompareString(m model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) || result.IsNull(rObj) {
		return result.New(nil)
	}
	l, r, err := applyToValues(lObj, rObj, result.ToString)
	if err != nil {
		return result.Value{}, err
	}
	return compare(m, l, r)
}

// op(left DateTime, right DateTime) Boolean
// https://cql.hl7.org/09-b-cqlreference.html#less
// https://cql.hl7.org/09-b-cqlreference.html#less-or-equal
// https://cql.hl7.org/09-b-cqlreference.html#greater
// https://cql.hl7.org/09-b-cqlreference.html#greater-or-equal
func evalCompareDateTime(m model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) || result.IsNull(rObj) {
		return result.New(nil)
	}
	l, r, err := applyToValues(lObj, rObj, result.ToDateTime)
	if err != nil {
		return result.Value{}, err
	}
	switch m.(type) {
	case *model.Less:
		return beforeDateTime(l, r)
	case *model.LessOrEqual:
		return beforeOrEqualDateTime(l, r)
	case *model.Greater:
		return afterDateTime(l, r)
	case *model.GreaterOrEqual:
		return afterOrEqualDateTime(l, r)
	}
	return result.Value{}, fmt.Errorf("internal error - unsupported Binary Comparison Expression %v", m)
}

func compare[n cmp.Ordered](m model.IBinaryExpression, l, r n) (result.Value, error) {
	switch m.(type) {
	case *model.Less:
		return result.New(l < r)
	case *model.LessOrEqual:
		return result.New(l <= r)
	case *model.Greater:
		return result.New(l > r)
	case *model.GreaterOrEqual:
		return result.New(l >= r)
	}
	return result.Value{}, fmt.Errorf("internal error - unsupported Binary Comparison Expression %v", m)
}
