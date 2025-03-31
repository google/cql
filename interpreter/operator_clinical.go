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
	"time"

	"github.com/google/cql/model"
	"github.com/google/cql/result"
	"github.com/google/cql/terminology"
	"github.com/google/cql/types"
)

// CLINICAL OPERATORS - https://cql.hl7.org/09-b-cqlreference.html#clinical-operators-3


// CalculateAgeIn[Years|Months|Weeks|Days|Hours|Minutes|Seconds](birthDate Date|DateTime) Integer
// https://cql.hl7.org/09-b-cqlreference.html#calculateage
func evalCalculateAge(u model.IUnaryExpression, birthObj result.Value) (result.Value, error) {
	m := u.(*model.CalculateAge)
	p := model.DateTimePrecision(m.Precision)
	if err := validatePrecision(p, []model.DateTimePrecision{model.YEAR, model.MONTH, model.WEEK, model.DAY, model.HOUR, model.MINUTE, model.SECOND}); err != nil {
		return result.Value{}, err
	}
	if result.IsNull(birthObj) {
		return result.New(nil)
	}

	birth, err := result.ToDateTime(birthObj)
	if err != nil {
		return result.Value{}, err
	}
	
	// Use current time as asOf time
	asOf := result.DateTime{Date: time.Now()}
	return calculateAgeAt(birth, asOf, p)
}

// CalculateAgeIn[Years|Months|Weeks|Days]At(birthDate Date, asOf Date) Integer
// https://cql.hl7.org/09-b-cqlreference.html#calculateageat
func evalCalculateAgeAtDate(b model.IBinaryExpression, birthObj, asOfObj result.Value) (result.Value, error) {
	m := b.(*model.CalculateAgeAt)
	p := model.DateTimePrecision(m.Precision)
	if err := validatePrecision(p, []model.DateTimePrecision{model.YEAR, model.MONTH, model.WEEK, model.DAY}); err != nil {
		return result.Value{}, err
	}
	if result.IsNull(birthObj) || result.IsNull(asOfObj) {
		return result.New(nil)
	}

	birth, asOf, err := applyToValues(birthObj, asOfObj, result.ToDateTime)
	if err != nil {
		return result.Value{}, err
	}
	return calculateAgeAt(birth, asOf, p)
}

// CalculateAgeIn[Years|Months|Weeks|Days|Hours|Minutes|Seconds]At(birthDate DateTime, asOf DateTime) Integer
// https://cql.hl7.org/09-b-cqlreference.html#calculateageat
func evalCalculateAgeAtDateTime(b model.IBinaryExpression, birthObj, asOfObj result.Value) (result.Value, error) {
	m := b.(*model.CalculateAgeAt)
	p := model.DateTimePrecision(m.Precision)
	if err := validatePrecision(p, []model.DateTimePrecision{model.YEAR, model.MONTH, model.WEEK, model.DAY, model.HOUR, model.MINUTE, model.SECOND}); err != nil {
		return result.Value{}, err
	}
	if result.IsNull(birthObj) || result.IsNull(asOfObj) {
		return result.New(nil)
	}

	birth, asOf, err := applyToValues(birthObj, asOfObj, result.ToDateTime)
	if err != nil {
		return result.Value{}, err
	}
	return calculateAgeAt(birth, asOf, p)
}

// calculateAgeAt is the helper to calculate age for Date and DateTime overloads.
func calculateAgeAt(birth, asOf result.DateTime, p model.DateTimePrecision) (result.Value, error) {
	// TODO(b/304349114): CalculateAgeAt is just syntactic sugar over the duration system operator. We
	// should fully implement duration and call the duration helper function from calculateAge.

	if p == model.YEAR {
		if asOf.Date.Month() > birth.Date.Month() ||
			(asOf.Date.Month() == birth.Date.Month() && asOf.Date.Day() >= birth.Date.Day()) {
			return result.New(asOf.Date.Year() - birth.Date.Year())
		}

		return result.New(asOf.Date.Year() - birth.Date.Year() - 1)
	}

	if p == model.MONTH {
		months := 12*(asOf.Date.Year()-birth.Date.Year()) + int((asOf.Date.Month())) - int(birth.Date.Month())
		if asOf.Date.Day() < birth.Date.Day() {
			months--
		}

		return result.New(months)
	}

	if p == model.WEEK {
		return result.New(int(asOf.Date.Sub(birth.Date).Hours() / 24 / 7))
	}

	if p == model.DAY {
		return result.New(int(asOf.Date.Sub(birth.Date).Hours() / 24))
	}

	// TODO(b/304349114): Per https://cql.hl7.org/09-b-cqlreference.html#ageat and
	// the external tests mentioned in b/304349114#comment3, these date-related
	// functions should propagate "uncertainty" ranges if precision is less than
	// a day.
	return result.Value{}, fmt.Errorf("Unsupported CalculateAgeAt precision %v", p)
}

// in(code Code, codesystem CodeSystemRef) Boolean
// in(codes List<Code>, codesystem CodeSystemRef) Boolean
// in(concept Concept, codesystem CodeSystemRef) Boolean
// in(concepts List<Concept>, codesystem CodeSystemRef) Boolean
// https://cql.hl7.org/09-b-cqlreference.html#in-codesystem
// The In operator for list overloads checks if any value is in the CodeSystem.
// TODO: b/327282181 - add support for other In CodeSystem Operators
func (i *interpreter) evalInCodeSystem(b model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) {
		return result.New(false)
	}

	if result.IsNull(rObj) {
		return result.Value{}, fmt.Errorf("in operator for CodeSystems should always resolve to a valid codesystem got: %v", rObj)
	}
	csv, err := result.ToCodeSystem(rObj)
	if err != nil {
		return result.Value{}, err
	}

	termCodes, err := valueToCodes(lObj)
	if err != nil {
		return result.Value{}, err
	}

	in, err := i.terminologyProvider.AnyInCodeSystem(termCodes, csv.ID, csv.Version)
	if err != nil {
		return result.Value{}, err
	}
	return result.New(in)
}

// in(code Code, valueset ValueSetRef) Boolean
// in(codes List<Code>, valueset ValueSetRef) Boolean
// in(concept Concept, valueset ValueSetRef) Boolean
// in(concepts List<Concept>, valueset ValueSetRef) Boolean
// https://cql.hl7.org/09-b-cqlreference.html#in-valueset
// The In operator for list overloads checks if any value is in the ValueSet.
// TODO: b/327281742 - add support for other In ValueSet Operators
func (i *interpreter) evalInValueSet(b model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	if result.IsNull(lObj) {
		return result.New(false)
	}

	if result.IsNull(rObj) {
		return result.Value{}, fmt.Errorf("in operator for Valuesets should always resolve to a valid valueset got: %v", rObj)
	}
	vsv, err := result.ToValueSet(rObj)
	if err != nil {
		return result.Value{}, err
	}

	termCodes, err := valueToCodes(lObj)
	if err != nil {
		return result.Value{}, err
	}

	in, err := i.terminologyProvider.AnyInValueSet(termCodes, vsv.ID, vsv.Version)
	if err != nil {
		return result.Value{}, err
	}
	return result.New(in)
}

// valueToCodes is the helper to convert a value to a list of terminology.Code. Returns an error for
// value types that are not valid clinical values. Currently only supports Code, Concept,
// List<Code>, List<Concept>.
// In cases where null codes can be present, they should be filtered out.
func valueToCodes(o result.Value) ([]terminology.Code, error) {
	var termCodes []terminology.Code
	if rt := o.RuntimeType(); rt.Equal(types.Code) {
		lv, err := result.ToCode(o)
		if err != nil {
			return nil, err
		}
		return []terminology.Code{{System: lv.System, Code: lv.Code}}, nil
	} else if rt.Equal(types.Concept) {
		concept, err := result.ToConcept(o)
		if err != nil {
			return nil, err
		}
		for _, c := range concept.NonNullCodeValues() {
			termCodes = append(termCodes, terminology.Code{System: c.System, Code: c.Code})
		}
		return termCodes, nil
	} else if rt.Equal(&types.List{ElementType: types.Code}) {
		list, err := result.ToSlice(o)
		if err != nil {
			return nil, err
		}

		for _, c := range list {
			code, err := result.ToCode(c)
			if err != nil {
				return nil, err
			}
			termCodes = append(termCodes, terminology.Code{System: code.System, Code: code.Code})
		}
		return termCodes, nil
	} else if rt.Equal(&types.List{ElementType: types.Concept}) {
		list, err := result.ToSlice(o)
		if err != nil {
			return nil, err
		}

		for _, c := range list {
			concept, err := result.ToConcept(c)
			if err != nil {
				return nil, err
			}
			for _, c := range concept.NonNullCodeValues() {
				termCodes = append(termCodes, terminology.Code{System: c.System, Code: c.Code})
			}
		}
		return termCodes, nil
	}
	return nil, fmt.Errorf("unsupported runtime type for clinical in operator, got: %v", o.RuntimeType())
}
