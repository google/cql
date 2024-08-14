// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
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

func (i *interpreter) evalLiteral(l *model.Literal) (result.Value, error) {
	t, ok := l.GetResultType().(types.System)
	if !ok {
		return result.Value{}, fmt.Errorf("Literal type must be a CQL base type, instead got %v", l.GetResultType())
	}
	// TODO(b/301606416): Many of these strconv are not quite correct.
	switch t {
	case types.Integer:
		a, err := strconv.ParseInt(l.Value, 10, 32)
		if err != nil {
			return result.Value{}, err
		}
		return result.NewWithSources(int32(a), l)
	case types.Long:
		value := l.Value
		if strings.HasSuffix(value, "L") {
			value = l.Value[:len(l.Value)-1]
		}
		a, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return result.Value{}, err
		}
		return result.NewWithSources(a, l)
	case types.Decimal:
		d, err := strconv.ParseFloat(l.Value, 64)
		if err != nil {
			return result.Value{}, err
		}
		return result.NewWithSources(d, l)
	case types.Boolean:
		b, err := strconv.ParseBool(l.Value)
		if err != nil {
			return result.Value{}, err
		}
		return result.NewWithSources(b, l)
	case types.String:
		return result.NewWithSources(l.Value, l)
	case types.Date:
		t, p, err := datehelpers.ParseDate(l.Value, i.evaluationTimestamp.Location())
		if err != nil {
			return result.Value{}, err
		}
		return result.NewWithSources(result.Date{Date: t, Precision: p}, l)
	case types.DateTime:
		t, p, err := datehelpers.ParseDateTime(l.Value, i.evaluationTimestamp.Location())
		if err != nil {
			return result.Value{}, err
		}
		return result.NewWithSources(result.DateTime{Date: t, Precision: p}, l)
	case types.Time:
		t, p, err := datehelpers.ParseTime(l.Value, i.evaluationTimestamp.Location())
		if err != nil {
			return result.Value{}, err
		}
		return result.NewWithSources(result.Time{Date: t, Precision: p}, l)
	case types.Any:
		if l.Value == "null" {
			return result.NewWithSources(nil, l)
		}
	}

	// Support other literals.
	return result.Value{}, fmt.Errorf("unsupported literal type %s %v", l.Value, t)
}

func (i *interpreter) evalInterval(l *model.Interval) (result.Value, error) {
	lowObj, err := i.evalExpression(l.Low)
	if err != nil {
		return result.Value{}, err
	}
	highObj, err := i.evalExpression(l.High)
	if err != nil {
		return result.Value{}, err
	}

	iType, ok := l.GetResultType().(*types.Interval)
	if !ok {
		return result.Value{}, fmt.Errorf("internal error -- interval result type should be an interval, got %v", l.GetResultType())
	}

	return result.NewWithSources(result.Interval{
		Low:           lowObj,
		High:          highObj,
		LowInclusive:  l.LowInclusive,
		HighInclusive: l.HighInclusive,
		StaticType:    iType,
	}, l, lowObj, highObj)
}

func (i *interpreter) evalList(l *model.List) (result.Value, error) {
	objs := []result.Value{}
	for index, e := range l.List {
		obj, err := i.evalExpression(e)
		if err != nil {
			return result.Value{}, fmt.Errorf("at index %d: %w", index, err)
		}
		objs = append(objs, obj)
	}
	listResultType, ok := l.GetResultType().(*types.List)
	if !ok {
		return result.Value{}, fmt.Errorf("internal error -- list result type should be a list, got %v", l.GetResultType())
	}
	return result.NewWithSources(result.List{Value: objs, StaticType: listResultType}, l, objs...)
}

func (i *interpreter) evalQuantity(q *model.Quantity) (result.Value, error) {
	qv := result.Quantity{Value: q.Value, Unit: q.Unit}
	return result.NewWithSources(qv, q)
}

func (i *interpreter) evalRatio(r *model.Ratio) (result.Value, error) {
	numerator, err := i.evalQuantity(&r.Numerator)
	if err != nil {
		return result.Value{}, err
	}
	denominator, err := i.evalQuantity(&r.Denominator)
	if err != nil {
		return result.Value{}, err
	}
	rv := result.Ratio{
		Numerator:   numerator.GolangValue().(result.Quantity),
		Denominator: denominator.GolangValue().(result.Quantity),
	}
	return result.NewWithSources(rv, r)
}

func (i *interpreter) evalCode(c *model.Code) (result.Value, error) {
	if c.System == nil {
		return result.Value{}, fmt.Errorf("The CodeSystem for a Code cannot be null, got code: %v", c)
	}
	cs, err := i.evalCodeSystemRef(c.System)
	if err != nil {
		return result.Value{}, err
	}

	csVal, err := result.ToCodeSystem(cs)
	if err != nil {
		return result.Value{}, err
	}
	cv := result.Code{
		Code:    c.Code,
		System:  csVal.ID,
		Version: csVal.Version,
		Display: c.Display,
	}

	return result.NewWithSources(cv, c)
}

func (i *interpreter) evalTuple(in *model.Tuple) (result.Value, error) {
	tuple := result.Tuple{
		Value:       make(map[string]result.Value),
		RuntimeType: in.GetResultType(),
	}

	for _, elem := range in.Elements {
		obj, err := i.evalExpression(elem.Value)
		if err != nil {
			return result.Value{}, err
		}
		tuple.Value[elem.Name] = obj
	}

	return result.NewWithSources(tuple, in)
}

func (i *interpreter) evalInstance(in *model.Instance) (result.Value, error) {
	elems := make(map[string]result.Value)
	for _, elem := range in.Elements {
		obj, err := i.evalExpression(elem.Value)
		if err != nil {
			return result.Value{}, err
		}
		elems[elem.Name] = obj
	}

	switch in.ClassType {
	case types.Quantity:
		qv := result.Quantity{}
		// Value
		valueObj, ok := elems["value"]
		if ok {
			value, err := result.ToFloat64(valueObj)
			if err != nil {
				return result.Value{}, err
			}
			qv.Value = value
		}

		// Unit
		unitObj, ok := elems["unit"]
		if ok {
			unit, err := result.ToString(unitObj)
			if err != nil {
				return result.Value{}, err
			}
			qv.Unit = model.Unit(unit)
		}

		return result.New(qv)
	case types.Code:
		cv := result.Code{}
		// Code
		codeObj, ok := elems["code"]
		if ok {
			code, err := result.ToString(codeObj)
			if err != nil {
				return result.Value{}, err
			}
			cv.Code = code
		}

		// System
		systemObj, ok := elems["system"]
		if ok {
			system, err := result.ToString(systemObj)
			if err != nil {
				return result.Value{}, err
			}
			cv.System = system
		}

		// Version
		versionObj, ok := elems["version"]
		if ok {
			version, err := result.ToString(versionObj)
			if err != nil {
				return result.Value{}, err
			}
			cv.Version = version
		}

		// Display
		displayObj, ok := elems["display"]
		if ok {
			display, err := result.ToString(displayObj)
			if err != nil {
				return result.Value{}, err
			}
			cv.Display = display
		}

		return result.New(cv)
	case types.CodeSystem:
		csv := result.CodeSystem{}

		// ID
		idObj, ok := elems["id"]
		if ok {
			id, err := result.ToString(idObj)
			if err != nil {
				return result.Value{}, err
			}
			csv.ID = id
		}

		// Version
		versionObj, ok := elems["version"]
		if ok {
			version, err := result.ToString(versionObj)
			if err != nil {
				return result.Value{}, err
			}
			csv.Version = version
		}

		return result.New(csv)
	case types.Concept:
		cv := result.Concept{}

		// Codes
		codeObj, ok := elems["codes"]
		if ok {
			codeObjs, err := result.ToSlice(codeObj)
			if err != nil {
				return result.Value{}, err
			}
			codes := make([]*result.Code, len(codeObjs))
			for i, codeObj := range codeObjs {
				if result.IsNull(codeObj) {
					codes[i] = nil
					continue
				}
				code, err := result.ToCode(codeObj)
				if err != nil {
					return result.Value{}, err
				}
				codes[i] = &code
			}
			cv.Codes = codes
		}

		// Display
		displayObj, ok := elems["display"]
		if ok {
			display, err := result.ToString(displayObj)
			if err != nil {
				return result.Value{}, err
			}
			cv.Display = display
		}

		return result.New(cv)
	case types.ValueSet:
		vsv := result.ValueSet{}

		// ID
		idObj, ok := elems["id"]
		if ok {
			id, err := result.ToString(idObj)
			if err != nil {
				return result.Value{}, err
			}
			vsv.ID = id
		}

		// Version
		versionObj, ok := elems["version"]
		if ok {
			version, err := result.ToString(versionObj)
			if err != nil {
				return result.Value{}, err
			}
			vsv.Version = version
		}

		// CodeSystem
		codeSysObj, ok := elems["codesystems"]
		if ok {
			codeSysObjs, err := result.ToSlice(codeSysObj)
			if err != nil {
				return result.Value{}, err
			}
			var codeSystems []result.CodeSystem
			for _, codeSysObj := range codeSysObjs {
				codeSys, err := result.ToCodeSystem(codeSysObj)
				if err != nil {
					return result.Value{}, err
				}
				codeSystems = append(codeSystems, codeSys)
			}
			vsv.CodeSystems = codeSystems
		}

		return result.New(vsv)
	}

	// This is a Named type not a System type.
	tuple := result.Tuple{
		Value:       make(map[string]result.Value),
		RuntimeType: in.ClassType,
	}

	for _, elem := range in.Elements {
		obj, err := i.evalExpression(elem.Value)
		if err != nil {
			return result.Value{}, err
		}
		// The parser should have already validated that the name and value of each element matches
		// model info for the class type.
		tuple.Value[elem.Name] = obj
	}

	return result.NewWithSources(tuple, in)
}
