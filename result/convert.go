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

package result

import (
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"
)

// ErrCannotConvert is an error that is returned when a conversion cannot be performed.
var ErrCannotConvert = errors.New("internal error - cannot convert")

// IsNull returns true if the provided Value is a null.
func IsNull(v Value) bool {
	return v.GolangValue() == nil
}

// ToBool takes a CQL Boolean and returns the underlying golang value, a bool.
func ToBool(v Value) (bool, error) {
	b, ok := v.GolangValue().(bool)
	if !ok {
		return false, fmt.Errorf("%w %v to a boolean", ErrCannotConvert, v.RuntimeType())
	}
	return b, nil
}

// ToString takes a CQL String and returns the underlying golang value, a string.
func ToString(v Value) (string, error) {
	s, ok := v.GolangValue().(string)
	if !ok {
		return "", fmt.Errorf("%w %v to a string", ErrCannotConvert, v.RuntimeType())
	}
	return s, nil
}

// ToInt32 takes a CQL Integer and returns the underlying golang value, an int32.
func ToInt32(v Value) (int32, error) {
	i, ok := v.GolangValue().(int32)
	if !ok {
		return 0, fmt.Errorf("%w %v to a int32", ErrCannotConvert, v.RuntimeType())
	}
	return i, nil
}

// ToInt64 takes a CQL Long and returns the underlying golang value, an int64.
func ToInt64(o Value) (int64, error) {
	l, ok := o.GolangValue().(int64)
	if !ok {
		return 0, fmt.Errorf("%w %v to a int64", ErrCannotConvert, o.RuntimeType())
	}
	return l, nil
}

// ToFloat64 takes a CQL Float and returns the underlying golang value, a float64.
func ToFloat64(v Value) (float64, error) {
	d, ok := v.GolangValue().(float64)
	if !ok {
		return 0, fmt.Errorf("%w %v to a float64", ErrCannotConvert, v.RuntimeType())
	}
	return d, nil
}

// ToQuantity takes a CQL Quantity and returns the underlying golang value, a Quantity.
func ToQuantity(v Value) (Quantity, error) {
	i, ok := v.GolangValue().(Quantity)
	if !ok {
		return Quantity{}, fmt.Errorf("%w %v to a Quantity", ErrCannotConvert, v.RuntimeType())
	}
	return i, nil
}

// ToRatio takes a CQL Ratio and returns the underlying golang value, a Ratio.
func ToRatio(v Value) (Ratio, error) {
	r, ok := v.GolangValue().(Ratio)
	if !ok {
		return Ratio{}, fmt.Errorf("%w %v to a Ratio", ErrCannotConvert, v.RuntimeType())
	}
	return r, nil
}

// ToDateTime takes a CQL Date, Time or DateTime and returns the underlying golang value, a
// DateTime. Since Date, Time and DateTime share the same underlying golang value DateTime()
// can be used to handle both Date, Time and DateTime generically.
func ToDateTime(v Value) (DateTime, error) {
	switch t := v.GolangValue().(type) {
	case DateTime:
		return t, nil
	case Date:
		return DateTime(t), nil
	case Time:
		return DateTime(t), nil
	default:
		return DateTime{}, fmt.Errorf("%w %v to a DateTime", ErrCannotConvert, v.RuntimeType())
	}
}

// ToInterval takes a CQL Interval and returns the underlying golang value, an Interval.
func ToInterval(v Value) (Interval, error) {
	i, ok := v.GolangValue().(Interval)
	if !ok {
		return Interval{}, fmt.Errorf("%w %v to a Interval", ErrCannotConvert, v.RuntimeType())
	}
	return i, nil
}

// ToSlice takes a CQL Slice and returns the underlying golang value, a []Value.
func ToSlice(v Value) ([]Value, error) {
	l, ok := v.GolangValue().(List)
	if !ok {
		return nil, fmt.Errorf("%w %v to a []Value", ErrCannotConvert, v.RuntimeType())
	}
	return l.Value, nil
}

// ToTuple takes a CQL Tuple and returns the underlying golang value, a map[string]Value.
func ToTuple(v Value) (map[string]Value, error) {
	t, ok := v.GolangValue().(Tuple)
	if !ok {
		return nil, fmt.Errorf("%w %v to a map[string]Value", ErrCannotConvert, v.RuntimeType())
	}
	return t.Value, nil
}

// ToProto takes a CQL Named type and returns the underlying golang value, a proto.Message. Named
// types are any type defined in the data model. The proto.Message is a FHIR Proto. See
// https://github.com/google/fhir for more details.
func ToProto(v Value) (proto.Message, error) {
	t, ok := v.GolangValue().(Named)
	if !ok {
		return nil, fmt.Errorf("%w %v to a proto.Message", ErrCannotConvert, v.RuntimeType())
	}
	return t.Value, nil
}

// ToCodeSystem takes a CQL CodeSystem and returns the underlying golang value, a CodeSystem.
func ToCodeSystem(o Value) (CodeSystem, error) {
	i, ok := o.GolangValue().(CodeSystem)
	if !ok {
		return CodeSystem{}, fmt.Errorf("%w %v to a CodeSystem", ErrCannotConvert, o.RuntimeType())
	}
	return i, nil
}

// ToValueSet takes a CQL ValueSet and returns the underlying golang value, a ValueSet.
func ToValueSet(o Value) (ValueSet, error) {
	i, ok := o.GolangValue().(ValueSet)
	if !ok {
		return ValueSet{}, fmt.Errorf("%w %v to a ValueSet", ErrCannotConvert, o.RuntimeType())
	}
	return i, nil
}

// ToConcept takes a CQL Concept and returns the underlying golang value, a Concept.
func ToConcept(v Value) (Concept, error) {
	c, ok := v.GolangValue().(Concept)
	if !ok {
		return Concept{}, fmt.Errorf("%w %v to a Concept", ErrCannotConvert, v.RuntimeType())
	}
	return c, nil
}

// ToCode takes a CQL Code and returns the underlying golang value, a Code.
func ToCode(v Value) (Code, error) {
	i, ok := v.GolangValue().(Code)
	if !ok {
		return Code{}, fmt.Errorf("%w %v to a Code", ErrCannotConvert, v.RuntimeType())
	}
	return i, nil
}
