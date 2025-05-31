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

	"github.com/google/cql/model"
	"github.com/google/cql/result"
	"github.com/google/cql/types"
	"github.com/google/cql/ucum"
)

var decimalStringRegex = regexp.MustCompile(`(\+|\-)?\d+(\.\d+)?`)
var longStringRegex = regexp.MustCompile(`(\+|\-)?\d+`)

// The string should start with a decimal value that may have a prefix of + or -.
// It optionally may also include a unit designation.
// Currently assumes the that inner string has not been escaped.
var quantityStringRegex = regexp.MustCompile(`^([\+|\-]?\d+(?:\.\d+)?){1}\s*('{1}[A-Za-z0-9-]+'{1})?$`)

// TYPE OPERATORS - https://cql.hl7.org/09-b-cqlreference.html#type-operators-1

// as<T>(argument Any) T
// cast as<T>(argument Any) T
// https://cql.hl7.org/09-b-cqlreference.html#as
func (i *interpreter) evalAs(m model.IUnaryExpression, obj result.Value) (result.Value, error) {
	a := m.(*model.As)

	// Null can be cast to anything https://cql.hl7.org/03-developersguide.html#implicit-casting
	if result.IsNull(obj) {
		return result.New(nil)
	}

	// This is a special case, anything can be cast to Any.
	if a.AsTypeSpecifier.Equal(types.Any) {
		return obj, nil
	}

	// At runtime a Choice<Integer, String> will be either Integer or String. So for Choice<Integer,
	// String> As String if the runtime type is a String that will be handled here.
	if obj.RuntimeType().Equal(a.AsTypeSpecifier) {
		return obj, nil
	}
	isSub, err := i.modelInfo.IsSubType(a.AsTypeSpecifier, obj.RuntimeType())
	if err != nil {
		return result.Value{}, err
	}
	if isSub {
		// TODO(b/301606416): The type should probably be changed to AsTypeSpecifier.
		return obj, nil
	}

	// This covers casts to choice types such as Decimal --> Choice<Decimal>. For cases that require a
	// conversion such as Integer --> Choice<Decimal> the parser should have already inserted any
	// necessary conversions so that obj is equal or a subtype of one of the choices
	// As(ToDecimal(operand), Choice<Decimal>).
	aChoice, ok := a.AsTypeSpecifier.(*types.Choice)
	if ok {
		for _, choice := range aChoice.ChoiceTypes {
			if obj.RuntimeType().Equal(choice) {
				return obj, nil
			}

			isSub, err := i.modelInfo.IsSubType(choice, obj.RuntimeType())
			if err != nil {
				return result.Value{}, err
			}
			if isSub {
				return obj, nil
			}
		}
	}

	if a.Strict {
		return result.Value{}, fmt.Errorf("cannot strict cast type %v to type %v", obj.RuntimeType().String(), a.AsTypeSpecifier.String())
	}

	return result.New(nil)
}

// is<T>(argument Any) Boolean
// https://cql.hl7.org/09-b-cqlreference.html#is
// TODO: b/326499348 - We may need to make some updates here or in the subtype resolution for choice
// types, based on how external spec items are clarified.
func (i *interpreter) evalIs(m model.IUnaryExpression, obj result.Value) (result.Value, error) {
	isExpr := m.(*model.Is)

	if obj.RuntimeType().Equal(isExpr.IsTypeSpecifier) {
		return result.New(true)
	}

	isSub, err := i.modelInfo.IsSubType(obj.RuntimeType(), isExpr.IsTypeSpecifier)
	if err != nil {
		return result.Value{}, err
	}
	fmt.Println(isSub)

	return result.New(isSub)
}

// ToDate(argument DateTime) Date
// ToDate(argument Date) Date
// https://cql.hl7.org/09-b-cqlreference.html#todate
func evalToDateDateTime(m model.IUnaryExpression, opObj result.Value) (result.Value, error) {
	if result.IsNull(opObj) {
		return result.New(nil)
	}
	op, err := result.ToDateTime(opObj)
	if err != nil {
		return result.Value{}, err
	}

	newDate := result.Date(op)
	if newDate.Precision != model.DAY && newDate.Precision != model.MONTH && newDate.Precision != model.YEAR {
		// Precision must be more precise than DAY, but we're converting to Date, so set the precision
		// to DAY, which is the max precision for Dates.
		newDate.Precision = model.DAY
	}
	return result.New(newDate)
}

// ToDate(argument String) Date
// https://cql.hl7.org/09-b-cqlreference.html#todate
//
// Converts a ISO-8601 date formatted string to a CQL Date.
func (i *interpreter) evalToDateString(m model.IUnaryExpression, opObj result.Value) (result.Value, error) {
	if result.IsNull(opObj) {
		return result.New(nil)
	}
	op, err := result.ToString(opObj)
	if err != nil {
		return result.Value{}, err
	}

	obj, err := i.stringToDate(op, types.Date)
	if err != nil {
		return result.Value{}, err
	}
	if obj.RuntimeType() != types.Date {
		return result.Value{}, fmt.Errorf("interal error - ToDateString, expected Date got: %v", obj.RuntimeType())
	}

	newDate := obj.GolangValue().(result.Date)
	if newDate.Precision != model.DAY && newDate.Precision != model.MONTH && newDate.Precision != model.YEAR {
		// Precision must be more precise than DAY, but we're converting to Date, so set the precision
		// to DAY, which is the max precision for Dates.
		newDate.Precision = model.DAY
	}
	return result.New(newDate)
}

// ToDateTime(argument Date) DateTime
// ToDateTime(argument DateTime) DateTime
// https://cql.hl7.org/09-b-cqlreference.html#todatetime
//
// Currently timezone will default to UTC, we need to change this when execution timestamp is added.
// Even though the result is a DateTime, time values are considered unspecified, and precision value is not changed.
// TODO(b/303836614) make this function execution timestamp context aware.
func evalToDateTimeDate(m model.IUnaryExpression, opObj result.Value) (result.Value, error) {
	if result.IsNull(opObj) {
		return result.New(nil)
	}

	op, err := result.ToDateTime(opObj)
	if err != nil {
		return result.Value{}, err
	}

	return result.New(op)
}

// ToDateTime(argument String) DateTime
// https://cql.hl7.org/09-b-cqlreference.html#todatetime
//
// Converts a ISO-8601 datetime formatted string to a CQL DateTime.
// Currently timezone will default to UTC, we need to change this when execution timestamp is added.
// Even though the result is a DateTime, time values are considered unspecified, and precision value is not changed.
// TODO(b/303836614) make this function execution timestamp context aware.
func (i *interpreter) evalToDateTimeString(m model.IUnaryExpression, opObj result.Value) (result.Value, error) {
	if result.IsNull(opObj) {
		return result.New(nil)
	}
	op, err := result.ToString(opObj)
	if err != nil {
		return result.Value{}, err
	}

	dt, err := i.stringToDate(op, types.DateTime)
	if err == nil {
		return dt, nil
	}
	// date formatted strings may also be parsed into a datetime.
	d, err := i.stringToDate(op, types.Date)
	if err != nil {
		return result.Value{}, err
	}
	dtv, ok := d.GolangValue().(result.Date)
	if !ok {
		return result.Value{}, fmt.Errorf("internal error - ToDateTimeString, failed to convert string in date format to datetime, %v", op)
	}
	return result.New(result.DateTime(dtv))
}

// ToDecimal(argument Decimal) Decimal
// ToDecimal(argument Long) Decimal
// ToDecimal(argument Integer) Decimal
// https://cql.hl7.org/09-b-cqlreference.html#todecimal
func evalToDecimal(m model.IUnaryExpression, opObj result.Value) (result.Value, error) {
	if result.IsNull(opObj) {
		return result.New(nil)
	}

	switch v := opObj.GolangValue().(type) {
	case int32:
		return result.New(float64(v))
	case int64:
		return result.New(float64(v))
	default:
		op, err := result.ToFloat64(opObj)
		if err != nil {
			return result.Value{}, err
		}
		return result.New(op)
	}
}

// ToDecimal(argument String) Decimal
// https://cql.hl7.org/09-b-cqlreference.html#todecimal
func evalToDecimalString(m model.IUnaryExpression, opObj result.Value) (result.Value, error) {
	if result.IsNull(opObj) {
		return result.New(nil)
	}
	op, err := result.ToString(opObj)
	if err != nil {
		return result.Value{}, err
	}
	return toDecimalFromString(op)
}

func toDecimalFromString(s string) (result.Value, error) {
	// Check that the string meets the CQL decimal spec requirements.
	found := decimalStringRegex.FindString(s)
	if found == "" || found != s {
		return result.New(nil)
	}

	// ParseFloat works for every string that meets the CQL spec.
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return result.Value{}, err
	}
	return result.New(f)
}

// ToDecimal(argument Boolean) Decimal
// https://cql.hl7.org/09-b-cqlreference.html#todecimal
func evalToDecimalBoolean(m model.IUnaryExpression, opObj result.Value) (result.Value, error) {
	if result.IsNull(opObj) {
		return result.New(nil)
	}
	op, err := result.ToBool(opObj)
	if err != nil {
		return result.Value{}, err
	}
	if op {
		return result.New(1.0)
	}
	return result.New(0.0)
}

// ToLong(argument Long) Long
// ToLong(argument Integer) Long
// https://cql.hl7.org/09-b-cqlreference.html#tolong
func evalToLong(m model.IUnaryExpression, opObj result.Value) (result.Value, error) {
	if result.IsNull(opObj) {
		return result.New(nil)
	}
	switch v := opObj.GolangValue().(type) {
	case int32:
		return result.New(int64(v))
	default:
		op, err := result.ToInt64(opObj)
		if err != nil {
			return result.Value{}, err
		}
		return result.New(op)
	}
}

// ToLong(argument String) Long
// https://cql.hl7.org/09-b-cqlreference.html#tolong
func evalToLongString(m model.IUnaryExpression, opObj result.Value) (result.Value, error) {
	if result.IsNull(opObj) {
		return result.New(nil)
	}
	op, err := result.ToString(opObj)
	if err != nil {
		return result.Value{}, err
	}

	// Check that the string meets the CQL long spec requirements.
	found := longStringRegex.FindString(op)
	if found == "" || found != op {
		return result.New(nil)
	}

	// ParseInt works for every string that meets the CQL spec.
	f, err := strconv.ParseInt(op, 10, 64)
	if err != nil {
		return result.Value{}, err
	}
	return result.New(f)
}

// ToLong(argument Boolean) Long
// https://cql.hl7.org/09-b-cqlreference.html#tolong
func evalToLongBoolean(m model.IUnaryExpression, opObj result.Value) (result.Value, error) {
	if result.IsNull(opObj) {
		return result.New(nil)
	}
	op, err := result.ToBool(opObj)
	if err != nil {
		return result.Value{}, err
	}
	if op {
		return result.New(int64(1))
	}
	return result.New(int64(0))
}

// ToConcept(argument Code) Concept
// https://cql.hl7.org/09-b-cqlreference.html#toconcept
func evalToConceptCode(m model.IUnaryExpression, obj result.Value) (result.Value, error) {
	if result.IsNull(obj) {
		return result.New(nil)
	}
	code, err := result.ToCode(obj)
	if err != nil {
		return result.Value{}, err
	}
	return result.New(result.Concept{Codes: []*result.Code{&code}, Display: code.Display})
}

// ToConcept(argument List<Code>) Concept
// https://cql.hl7.org/09-b-cqlreference.html#toconcept
func evalToConceptList(m model.IUnaryExpression, obj result.Value) (result.Value, error) {
	if result.IsNull(obj) {
		return result.New(nil)
	}
	codeObjs, err := result.ToSlice(obj)
	if err != nil {
		return result.Value{}, err
	}

	codes := make([]*result.Code, len(codeObjs))
	for i, code := range codeObjs {
		if result.IsNull(code) {
			codes[i] = nil
			continue
		}
		c, err := result.ToCode(code)
		if err != nil {
			return result.Value{}, err
		}
		codes[i] = &c
	}
	return result.New(result.Concept{Codes: codes})
}

// ToQuantity(argument Decimal) Quantity
// ToQuantity(argument Integer) Quantity
// https://cql.hl7.org/09-b-cqlreference.html#toquantity
// TODO: b/323978857 - Implement ToQuantity for Ratio.
// TODO: b/324131050 - Implement ToQuantity for strings.
func evalToQuantity(m model.IUnaryExpression, opObj result.Value) (result.Value, error) {
	if result.IsNull(opObj) {
		return result.New(nil)
	}
	switch t := opObj.GolangValue().(type) {
	case float64:
		return result.New(result.Quantity{Value: t, Unit: model.ONEUNIT})
	case int32:
		return result.New(result.Quantity{Value: float64(t), Unit: model.ONEUNIT})
	default:
		op, err := result.ToQuantity(opObj)
		if err != nil {
			return result.Value{}, err
		}
		return result.New(op)
	}
}

// ToQuantity(argument String) Quantity
// https://cql.hl7.org/09-b-cqlreference.html#toquantity
func evalToQuantityString(m model.IUnaryExpression, opObj result.Value) (result.Value, error) {
	if result.IsNull(opObj) {
		return result.New(nil)
	}
	op, err := result.ToString(opObj)
	if err != nil {
		return result.Value{}, err
	}

	// On valid match FindStringSubmatch returns a list containing:
	// the whole matched text, the captured number, the captured unit text.
	found := quantityStringRegex.FindStringSubmatch(op)
	if len(found) != 3 {
		return result.New(nil)
	}

	// ParseFloat works for every string that meets the CQL spec.
	f, err := strconv.ParseFloat(found[1], 64)
	if err != nil {
		return result.Value{}, err
	}
	unit := "1"
	if len(found[2]) != 0 {
		// trim off quotations
		unit = found[2][1 : len(found[2])-1]
	}
	if valid, _ := ucum.ValidateUnit(unit, true, true); !valid {
		return result.New(nil)
	}
	return result.New(result.Quantity{Value: f, Unit: model.Unit(unit)})
}

// Add an @ symbol to the string so we can use the same parsing logic as engine literals.
func (i *interpreter) stringToDate(input string, inputType types.System) (result.Value, error) {
	return i.evalLiteral(&model.Literal{
		Value:      "@" + unqoteSingle(input),
		Expression: &model.Expression{Element: &model.Element{ResultType: inputType}},
	})
}

// unqoteSingle returns the unquoted version of the string, if it's quoted with single quotes,
// otherwise returns the input string.
func unqoteSingle(str string) string {
	if len(str) > 0 && strings.HasPrefix(str, "'") && strings.HasSuffix(str, "'") {
		return str[1 : len(str)-1]
	}
	return str
}

// ConvertQuantity(argument Quantity, to Unit) Quantity
// https://cql.hl7.org/09-b-cqlreference.html#convertquantity
func evalConvertQuantity(m model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) || result.IsNull(rObj) {
		return result.New(nil)
	}
	quantity, err := result.ToQuantity(lObj)
	if err != nil {
		return result.Value{}, err
	}
	toUnit, err := result.ToString(rObj)
	if err != nil {
		return result.Value{}, err
	}

	// If the units are the same, no conversion needed.
	if quantity.Unit == model.Unit(toUnit) {
		return lObj, nil
	}

	newVal, err := ucum.ConvertUnit(quantity.Value, string(quantity.Unit), toUnit)
	if err != nil {
		return result.New(nil)
	}

	return result.New(result.Quantity{
		Value: newVal,
		Unit:  model.Unit(toUnit),
	})
}
