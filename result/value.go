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
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	anypb "google.golang.org/protobuf/types/known/anypb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	datepb "google.golang.org/genproto/googleapis/type/date"
	timeofdaypb "google.golang.org/genproto/googleapis/type/timeofday"

	"github.com/google/cql/internal/datehelpers"
	"github.com/google/cql/model"
	crpb "github.com/google/cql/protos/cql_result_go_proto"
	"github.com/google/cql/types"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Value is a CQL Value evaluated by the interpreter.
type Value struct {
	goValue     any
	runtimeType types.IType
	sourceExpr  model.IExpression
	sourceVals  []Value
}

// GolangValue returns the underlying Golang value representing the CQL value. Specifically:
// CQL Null returns Golang nil
// CQL Boolean returns Golang bool
// CQL String returns Golang string
// CQL Integer returns Golang int32
// CQL Long returns Golang int64
// CQL Decimal returns Golang float64
// CQL Quantity returns Golang Quantity struct
// CQL Ratio returns Golang Ratio struct
// CQL Date returns Golang Date struct
// CQL DateTime returns Golang DateTime struct
// CQL Time returns Golang Time struct
// CQL Interval returns Golang Interval struct
// CQL List returns Golang List struct
// CQL Tuple returns Golang Tuple struct
// CQL Named (a type defined in the data model) returns Golang Proto struct
// CQL CodeSystem returns Golang CodeSystem struct
// CQL ValueSet returns Golang ValueSet struct
// CQL Concept returns Golang Concept struct
// CQL Code returns Goland Code struct
//
// You can call GolangValue() and type switch to handle values. Alternatively, if you know that the
// result will be a specific type such as an int32, it is recommended to use the result.ToInt32()
// helper function.
func (v Value) GolangValue() any { return v.goValue }

// RuntimeType returns the type used by the Is system operator
// https://cql.hl7.org/09-b-cqlreference.html#is. This may be different than the type statically
// determined by the Parser. For example, if the Parser statically determines the type to be
// Choice<String, Integer> the runtime type will be the actual type during evaluation, either
// Integer, String or Null. In some cases where a runtime is not known (for example an empty list,
// or an interval with where low and high are nulls) this will fall back to the static type.
func (v Value) RuntimeType() types.IType {
	switch t := v.goValue.(type) {
	case Interval:
		return inferIntervalType(t)
	case List:
		return inferListType(t.Value, t.StaticType)
	default:
		return v.runtimeType
	}
}

// SourceExpression is the CQL expression that created this value. For instance, if the returned
// result is from the CQL expression "a < b", the source expression will be the `model.Less` struct.
func (v Value) SourceExpression() model.IExpression { return v.sourceExpr }

// SourceValues returns the underlying values that were used by the SourceExpression to compute the
// returned value. The ordering of source values is not guaranteed to have any meaning, although
// expressions that produce them should attempt to preserve order when it does have meaning. For
// instance, for the value returned by "a < b", the source values are `a` and `b`.
//
// Source Values will have their own sources, creating a recursive tree structure that allows users
// to trace through the tree of expressions and values used to create it.
func (v Value) SourceValues() []Value { return v.sourceVals }

// For simple types, we can just marshal the value and type.
// More complex representations are handled in marshalJSON() functions of specific types.
type simpleJSONMessage struct {
	Type  json.RawMessage `json:"@type"`
	Value any             `json:"value"`
}

// customJSONMarshaler is an interface for types that need to marshal their own JSON representation.
// I.E. types that are not simple types.
type customJSONMarshaler interface {
	// marshalJSON accepts a bytes array of the type string and returns the JSON representation.
	marshalJSON(json.RawMessage) ([]byte, error)
}

// MarshalJSON returns the value as a JSON string.
// Uses CQL-Serialization spec as a template:
// https://github.com/cqframework/clinical_quality_language/wiki/CQL-Serialization
func (v Value) MarshalJSON() ([]byte, error) {
	rt, err := v.RuntimeType().MarshalJSON()
	if err != nil {
		return nil, err
	}

	// TODO: b/301606416 - Vocabulary support.
	switch gv := v.goValue.(type) {
	case customJSONMarshaler:
		return gv.marshalJSON(rt)
	case bool, float64, int32, int64, string, nil:
		return json.Marshal(simpleJSONMessage{
			Value: gv,
			Type:  rt,
		})
	case Date:
		date, err := datehelpers.DateString(gv.Date, gv.Precision)
		if err != nil {
			return nil, err
		}
		return json.Marshal(simpleJSONMessage{
			Type:  rt,
			Value: date,
		})
	case DateTime:
		dt, err := datehelpers.DateTimeString(gv.Date, gv.Precision)
		if err != nil {
			return nil, err
		}
		return json.Marshal(simpleJSONMessage{
			Type:  rt,
			Value: dt,
		})
	case Time:
		t, err := datehelpers.TimeString(gv.Date, gv.Precision)
		if err != nil {
			return nil, err
		}
		return json.Marshal(simpleJSONMessage{
			Type:  rt,
			Value: t,
		})
	case List:
		// Lists don't embed the type so they can be directly marshalled.
		return json.Marshal(gv.Value)
	case Tuple:
		// Tuples don't embed the type so they can be directly marshalled.
		return json.Marshal(gv.Value)
	default:
		return nil, fmt.Errorf("tried to marshal unsupported type %T, %w", gv, errUnsupportedType)
	}
}

// Proto converts Value to a proto. The source expression and source values are dropped.
func (v Value) Proto() (*crpb.Value, error) {
	pbValue := &crpb.Value{}
	switch t := v.goValue.(type) {
	case nil:
		// TODO(b/301606416): Consider supporting typed nulls.
		return pbValue, nil
	case bool:
		pbValue.Value = &crpb.Value_BooleanValue{BooleanValue: t}
	case string:
		pbValue.Value = &crpb.Value_StringValue{StringValue: t}
	case int32:
		pbValue.Value = &crpb.Value_IntegerValue{IntegerValue: t}
	case int64:
		pbValue.Value = &crpb.Value_LongValue{LongValue: t}
	case float64:
		pbValue.Value = &crpb.Value_DecimalValue{DecimalValue: t}
	case Quantity:
		pbValue.Value = &crpb.Value_QuantityValue{QuantityValue: t.Proto()}
	case Ratio:
		pbValue.Value = &crpb.Value_RatioValue{RatioValue: t.Proto()}
	case Date:
		pb, err := t.Proto()
		if err != nil {
			return nil, err
		}
		pbValue.Value = &crpb.Value_DateValue{DateValue: pb}
	case DateTime:
		pb, err := t.Proto()
		if err != nil {
			return nil, err
		}
		pbValue.Value = &crpb.Value_DateTimeValue{DateTimeValue: pb}
	case Time:
		pb, err := t.Proto()
		if err != nil {
			return nil, err
		}
		pbValue.Value = &crpb.Value_TimeValue{TimeValue: pb}
	case Interval:
		pb, err := t.Proto()
		if err != nil {
			return nil, err
		}
		pbValue.Value = &crpb.Value_IntervalValue{IntervalValue: pb}
	case List:
		pb, err := t.Proto()
		if err != nil {
			return nil, err
		}
		pbValue.Value = &crpb.Value_ListValue{ListValue: pb}
	case Named:
		pb, err := t.Proto()
		if err != nil {
			return nil, err
		}
		pbValue.Value = &crpb.Value_NamedValue{NamedValue: pb}
	case Tuple:
		pb, err := t.Proto()
		if err != nil {
			return nil, err
		}
		pbValue.Value = &crpb.Value_TupleValue{TupleValue: pb}
	case CodeSystem:
		pbValue.Value = &crpb.Value_CodeSystemValue{CodeSystemValue: t.Proto()}
	case ValueSet:
		pbValue.Value = &crpb.Value_ValueSetValue{ValueSetValue: t.Proto()}
	case Concept:
		pbValue.Value = &crpb.Value_ConceptValue{ConceptValue: t.Proto()}
	case Code:
		pbValue.Value = &crpb.Value_CodeValue{CodeValue: t.Proto()}
	}
	return pbValue, nil
}

// Equal is our custom implementation of equality used primarily by cmp.Diff in tests. This is not
// CQL equality. Equal only compares the GolangValue and RuntimeType, ignoring SourceExpression and
// SourceValues.
func (v Value) Equal(a Value) bool {
	if !v.RuntimeType().Equal(a.RuntimeType()) {
		return false
	}

	switch t := v.goValue.(type) {
	case Date:
		vDate, ok := a.GolangValue().(Date)
		if !ok {
			return false
		}
		return t.Equal(vDate)
	case DateTime:
		vDateTime, ok := a.GolangValue().(DateTime)
		if !ok {
			return false
		}
		return t.Equal(vDateTime)
	case Time:
		vTime, ok := a.GolangValue().(Time)
		if !ok {
			return false
		}
		return t.Equal(vTime)
	case Interval:
		vInterval, ok := a.GolangValue().(Interval)
		if !ok {
			return false
		}
		return t.Equal(vInterval)
	case List:
		vList, ok := a.GolangValue().(List)
		if !ok {
			return false
		}
		return t.Equal(vList)
	case Tuple:
		vTuple, ok := a.GolangValue().(Tuple)
		if !ok {
			return false
		}
		return t.Equal(vTuple)
	case Named:
		vProto, ok := a.GolangValue().(Named)
		if !ok {
			return false
		}
		return t.Equal(vProto)
	case ValueSet:
		vValueSet, ok := a.GolangValue().(ValueSet)
		if !ok {
			return false
		}
		return t.Equal(vValueSet)
	case Concept:
		vConcept, ok := a.GolangValue().(Concept)
		if !ok {
			return false
		}
		return t.Equal(vConcept)
	default:
		return v.GolangValue() == a.GolangValue()
	}
}

var errUnsupportedType = errors.New("unsupported type")

// New converts Golang values to CQL values. This function should be used when creating values from
// call sites where the supporting sources are not know, and to be added with the WithSources()
// function later. Call sites with the needed sources are encouraged to use NewWithSources below.
// Specifically:
// Golang bool converts to CQL Boolean
// Golang string converts to CQL String
// Golang int32 converts to CQL Integer
// Golang int64 converts to CQL Long
// Golang float64 converts to CQL Decimal
// Golang Quantity struct converts to CQL Quantity
// Golang Ratio struct converts to CQL Ratio
// Golang Date struct converts to CQL Date
// Golang DateTime struct converts to CQL DateTime
// Golang Time struct converts to CQL Time
// Golang Interval struct converts to CQL Interval
// Golang []Value converts to CQL List
// Golang map[string]Value converts to CQL Tuple
// Golang proto.Message (a type defined in the data model) converts to CQL Named
// Golang CodeSystem struct converts to CQL CodeSystem
// Golang ValueSet struct converts to CQL ValueSet
// Golang Concept struct converts to CQL Concept
// Golang Code struct converts to CQL Code
func New(val any) (Value, error) {
	if val == nil {
		return Value{runtimeType: types.Any, goValue: nil}, nil
	}
	switch v := val.(type) {
	case int:
		return Value{runtimeType: types.Integer, goValue: int32(v)}, nil
	case int32:
		return Value{runtimeType: types.Integer, goValue: v}, nil
	case int64:
		return Value{runtimeType: types.Long, goValue: v}, nil
	case float64:
		return Value{runtimeType: types.Decimal, goValue: v}, nil
	case Quantity:
		return Value{runtimeType: types.Quantity, goValue: v}, nil
	case Ratio:
		return Value{runtimeType: types.Ratio, goValue: v}, nil
	case bool:
		return Value{runtimeType: types.Boolean, goValue: v}, nil
	case string:
		return Value{runtimeType: types.String, goValue: v}, nil
	case Date:
		switch v.Precision {
		case model.YEAR, model.MONTH, model.DAY, model.UNSETDATETIMEPRECISION:
			return Value{runtimeType: types.Date, goValue: v}, nil
		}
		return Value{}, fmt.Errorf("unsupported precision in Date with value %v %w", v.Precision, datehelpers.ErrUnsupportedPrecision)
	case DateTime:
		switch v.Precision {
		case model.YEAR,
			model.MONTH,
			model.DAY,
			model.HOUR,
			model.MINUTE,
			model.SECOND,
			model.MILLISECOND,
			model.UNSETDATETIMEPRECISION:
			return Value{runtimeType: types.DateTime, goValue: v}, nil
		}
		return Value{}, fmt.Errorf("unsupported precision in DateTime with value %v %w", v.Precision, datehelpers.ErrUnsupportedPrecision)
	case Time:
		switch v.Precision {
		case model.HOUR, model.MINUTE, model.SECOND, model.MILLISECOND, model.UNSETDATETIMEPRECISION:
			if v.Date.Year() != 0 || v.Date.Month() != 1 || v.Date.Day() != 1 {
				return Value{}, fmt.Errorf("internal error - Time must be Year 0000, Month 01, Day 01, instead got %v", v.Date)
			}
			return Value{runtimeType: types.Time, goValue: v}, nil
		}
		return Value{}, fmt.Errorf("unsupported precision in Time with value %v %w", v.Precision, datehelpers.ErrUnsupportedPrecision)
	case Interval:
		// RuntimeType is not set here because it is inferred at RuntimeType() is called.
		return Value{goValue: v}, nil
	case List:
		// RuntimeType is not set here because it is inferred when RuntimeType() is called.
		return Value{goValue: v}, nil
	case Named:
		return Value{runtimeType: v.RuntimeType, goValue: v}, nil
	case Tuple:
		return Value{runtimeType: v.RuntimeType, goValue: v}, nil
	case CodeSystem:
		if v.ID == "" {
			return Value{}, fmt.Errorf("%v must have an ID", types.CodeSystem)
		}
		return Value{runtimeType: types.CodeSystem, goValue: v}, nil
	case Concept:
		if len(v.Codes) == 0 {
			return Value{}, fmt.Errorf("%v must have at least one %v", types.Concept, types.Code)
		}
		return Value{runtimeType: types.Concept, goValue: v}, nil
	case ValueSet:
		if v.ID == "" {
			return Value{}, fmt.Errorf("%v must have an ID", types.ValueSet)
		}
		return Value{runtimeType: types.ValueSet, goValue: v}, nil
	case Code:
		if v.Code == "" {
			return Value{}, fmt.Errorf("%v must have a Code", types.Code)
		}
		return Value{runtimeType: types.Code, goValue: v}, nil
	default:
		return Value{}, fmt.Errorf("%T %w", v, errUnsupportedType)
	}
}

// NewWithSources converts Golang values to CQL values when the sources are known. See New()
// function for full documentation.
func NewWithSources(val any, sourceExp model.IExpression, sourceObjs ...Value) (Value, error) {
	o, err := New(val)
	if err != nil {
		return Value{}, err
	}
	return o.WithSources(sourceExp, sourceObjs...), nil
}

// WithSources returns a version of the value with the given sources. This function has
// the following semantics to ensure all child values and expressions are recursively preserved
// as values propagate through the evaluation tree:
//
// First, if the value already has sources, this creates a copy of that value with the newly
// provided sources, so the original and its sources are preserved. Therefore an value with
// existing sources is never mutated and can be safely stored or reused across many consuming
// expressions if needed by the engine implementation.
//
// Second, if a caller does not explicitly provide a new set of source values, this function will
// use the existing value this is invoked on as the source. For instance, function implementations
// can do this to propagate a trace up the call stack by simply calling
// `valueToReturn.WithSources(theFunctionExpression)` prior to returning.
func (v Value) WithSources(sourceExp model.IExpression, sourceObjs ...Value) Value {
	if v.sourceExpr == nil {
		v.sourceExpr = sourceExp
		v.sourceVals = sourceObjs
		return v
	}

	// TODO b/301606416: This does not make a copy of val for lists, tuples and proto types. This is
	// ok since we currently don't mutate Values after they are created.
	if len(sourceObjs) == 0 {
		return Value{runtimeType: v.runtimeType, goValue: v.goValue, sourceExpr: sourceExp, sourceVals: []Value{v}}
	}
	return Value{runtimeType: v.runtimeType, goValue: v.goValue, sourceExpr: sourceExp, sourceVals: sourceObjs}
}

// NewFromProto converts a proto to a Value. The source expression and source values are dropped.
func NewFromProto(pb *crpb.Value) (Value, error) {
	switch t := pb.GetValue().(type) {
	case nil:
		return New(nil)
	case *crpb.Value_BooleanValue:
		return New(t.BooleanValue)
	case *crpb.Value_StringValue:
		return New(t.StringValue)
	case *crpb.Value_IntegerValue:
		return New(t.IntegerValue)
	case *crpb.Value_LongValue:
		return New(t.LongValue)
	case *crpb.Value_DecimalValue:
		return New(t.DecimalValue)
	case *crpb.Value_QuantityValue:
		return New(QuantityFromProto(t.QuantityValue))
	case *crpb.Value_RatioValue:
		return New(RatioFromProto(t.RatioValue))
	case *crpb.Value_DateValue:
		v, err := DateFromProto(t.DateValue)
		if err != nil {
			return Value{}, err
		}
		return New(v)
	case *crpb.Value_DateTimeValue:
		v, err := DateTimeFromProto(t.DateTimeValue)
		if err != nil {
			return Value{}, err
		}
		return New(v)
	case *crpb.Value_TimeValue:
		v, err := TimeFromProto(t.TimeValue)
		if err != nil {
			return Value{}, err
		}
		return New(v)
	case *crpb.Value_IntervalValue:
		v, err := IntervalFromProto(t.IntervalValue)
		if err != nil {
			return Value{}, err
		}
		return New(v)
	case *crpb.Value_ListValue:
		v, err := ListFromProto(t.ListValue)
		if err != nil {
			return Value{}, err
		}
		return New(v)
	case *crpb.Value_TupleValue:
		v, err := TupleFromProto(t.TupleValue)
		if err != nil {
			return Value{}, err
		}
		return New(v)
	case *crpb.Value_NamedValue:
		v, err := NamedFromProto(t.NamedValue)
		if err != nil {
			return Value{}, err
		}
		return New(v)
	case *crpb.Value_CodeSystemValue:
		return New(CodeSystemFromProto(t.CodeSystemValue))
	case *crpb.Value_ValueSetValue:
		return New(ValueSetFromProto(t.ValueSetValue))
	case *crpb.Value_ConceptValue:
		return New(ConceptFromProto(t.ConceptValue))
	case *crpb.Value_CodeValue:
		return New(CodeFromProto(t.CodeValue))
	default:
		return Value{}, fmt.Errorf("%T %w", pb, errUnsupportedType)
	}
}

// Quantity represents a decimal value with an associated unit string.
type Quantity struct {
	Value float64
	Unit  model.Unit
}

// Proto converts Quantity to a proto.
func (q Quantity) Proto() *crpb.Quantity {
	return &crpb.Quantity{
		Value: proto.Float64(q.Value),
		Unit:  proto.String(string(q.Unit)),
	}
}

// QuantityFromProto converts a proto to a Quantity.
func QuantityFromProto(pb *crpb.Quantity) Quantity {
	return Quantity{Value: pb.GetValue(), Unit: model.Unit(pb.GetUnit())}
}

func (q Quantity) marshalJSON(t json.RawMessage) ([]byte, error) {
	return json.Marshal(struct {
		Type  json.RawMessage `json:"@type"`
		Value float64         `json:"value"`
		Unit  string          `json:"unit"`
	}{
		Type:  t,
		Value: q.Value,
		Unit:  string(q.Unit),
	})
}

// Ratio represents a ratio of two quantities.
type Ratio struct {
	Numerator   Quantity
	Denominator Quantity
}

// RatioFromProto converts a proto to a Ratio.
func RatioFromProto(pb *crpb.Ratio) Ratio {
	return Ratio{Numerator: QuantityFromProto(pb.Numerator), Denominator: QuantityFromProto(pb.Denominator)}
}

// Proto converts Ratio to a proto.
func (r Ratio) Proto() *crpb.Ratio {
	return &crpb.Ratio{
		Numerator:   r.Numerator.Proto(),
		Denominator: r.Denominator.Proto(),
	}
}

func (r Ratio) marshalJSON(t json.RawMessage) ([]byte, error) {
	quantityType, err := types.Quantity.MarshalJSON()
	if err != nil {
		return nil, err
	}
	marshalledNumerator, err := r.Numerator.marshalJSON(quantityType)
	if err != nil {
		return nil, err
	}
	marshalledDenominator, err := r.Denominator.marshalJSON(quantityType)
	if err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		Type        json.RawMessage `json:"@type"`
		Numerator   json.RawMessage `json:"numerator"`
		Denominator json.RawMessage `json:"denominator"`
	}{
		Type:        t,
		Numerator:   marshalledNumerator,
		Denominator: marshalledDenominator,
	})
}

// Date is the Golang representation of a CQL Date. CQL Dates do not have timezone offsets, but
// Golang time.Time requires a location. The time.Time should always have the offset of the
// evaluation timestamp. The precision will be is Year, Month or Day.
type Date DateTime

// Equal returns true if this Date matches the provided one, otherwise false.
func (d Date) Equal(v Date) bool {
	return DateTime(d).Equal(DateTime(v))
}

// Proto converts Date to a proto.
func (d Date) Proto() (*crpb.Date, error) {
	pbCQLDate := &crpb.Date{}
	pbDate := &datepb.Date{}
	switch d.Precision {
	case model.YEAR:
		pbDate.Year = int32(d.Date.Year())
		pbCQLDate.Precision = crpb.Date_PRECISION_YEAR.Enum()
	case model.MONTH:
		pbDate.Year = int32(d.Date.Year())
		pbDate.Month = int32(d.Date.Month())
		pbCQLDate.Precision = crpb.Date_PRECISION_MONTH.Enum()
	case model.DAY:
		pbDate.Year = int32(d.Date.Year())
		pbDate.Month = int32(d.Date.Month())
		pbDate.Day = int32(d.Date.Day())
		pbCQLDate.Precision = crpb.Date_PRECISION_DAY.Enum()
	default:
		return nil, fmt.Errorf("unsupported precision in Date with value %v %w", d.Precision, datehelpers.ErrUnsupportedPrecision)
	}

	pbCQLDate.Date = pbDate
	return pbCQLDate, nil
}

// DateFromProto converts a proto to a Date.
func DateFromProto(pb *crpb.Date) (Date, error) {
	var modelPrecision model.DateTimePrecision
	switch pb.GetPrecision() {
	case crpb.Date_PRECISION_YEAR:
		modelPrecision = model.YEAR
	case crpb.Date_PRECISION_MONTH:
		modelPrecision = model.MONTH
	case crpb.Date_PRECISION_DAY:
		modelPrecision = model.DAY
	default:
		return Date{}, fmt.Errorf("unsupported precision converting proto to Date  %v %w", pb.GetPrecision(), datehelpers.ErrUnsupportedPrecision)
	}
	return Date{Date: time.Date(int(pb.Date.GetYear()), time.Month(pb.Date.GetMonth()), int(pb.Date.GetDay()), 0, 0, 0, 0, time.UTC), Precision: modelPrecision}, nil
}

// DateTime is the Golang representation of a CQL DateTime. The time.Time may have different
// offsets. The precision will be anything from Year to Millisecond.
type DateTime struct {
	Date      time.Time
	Precision model.DateTimePrecision
}

// Proto converts DateTime to a proto.
func (d DateTime) Proto() (*crpb.DateTime, error) {
	pbCQLDateTime := &crpb.DateTime{}
	switch d.Precision {
	case model.YEAR:
		pbCQLDateTime.Precision = crpb.DateTime_PRECISION_YEAR.Enum()
	case model.MONTH:
		pbCQLDateTime.Precision = crpb.DateTime_PRECISION_MONTH.Enum()
	case model.DAY:
		pbCQLDateTime.Precision = crpb.DateTime_PRECISION_DAY.Enum()
	case model.HOUR:
		pbCQLDateTime.Precision = crpb.DateTime_PRECISION_HOUR.Enum()
	case model.MINUTE:
		pbCQLDateTime.Precision = crpb.DateTime_PRECISION_MINUTE.Enum()
	case model.SECOND:
		pbCQLDateTime.Precision = crpb.DateTime_PRECISION_SECOND.Enum()
	case model.MILLISECOND:
		pbCQLDateTime.Precision = crpb.DateTime_PRECISION_MILLISECOND.Enum()
	default:
		return nil, fmt.Errorf("unsupported precision in DateTime with value %v %w", d.Precision, datehelpers.ErrUnsupportedPrecision)
	}

	pbCQLDateTime.Date = timestamppb.New(d.Date)
	return pbCQLDateTime, nil
}

// DateTimeFromProto converts a proto to a DateTime.
func DateTimeFromProto(pb *crpb.DateTime) (DateTime, error) {
	var modelPrecision model.DateTimePrecision
	switch pb.GetPrecision() {
	case crpb.DateTime_PRECISION_YEAR:
		modelPrecision = model.YEAR
	case crpb.DateTime_PRECISION_MONTH:
		modelPrecision = model.MONTH
	case crpb.DateTime_PRECISION_DAY:
		modelPrecision = model.DAY
	case crpb.DateTime_PRECISION_HOUR:
		modelPrecision = model.HOUR
	case crpb.DateTime_PRECISION_MINUTE:
		modelPrecision = model.MINUTE
	case crpb.DateTime_PRECISION_SECOND:
		modelPrecision = model.SECOND
	case crpb.DateTime_PRECISION_MILLISECOND:
		modelPrecision = model.MILLISECOND
	default:
		return DateTime{}, fmt.Errorf("unsupported precision converting Proto to DateTime %v %w", pb.GetPrecision(), datehelpers.ErrUnsupportedPrecision)
	}
	return DateTime{Date: pb.Date.AsTime(), Precision: modelPrecision}, nil
}

// Equal returns true if this DateTime matches the provided one, otherwise false.
func (d DateTime) Equal(v DateTime) bool {
	if !d.Date.Equal(v.Date) {
		return false
	}
	if d.Precision != v.Precision {
		return false
	}
	return true
}

// Time is the Golang representation of a CQL Time. CQL Times do not have year, month, days or a
// timezone but Golang time.Time does. We use the date 0000-01-01 and timezone UTC for all golang
// time.Time. The precision will be between Hour and Millisecond.
type Time DateTime

// Equal returns true if this Time matches the provided one, otherwise false.
func (t Time) Equal(v Time) bool {
	return DateTime(t).Equal(DateTime(v))
}

// Proto converts Time to a proto.
func (t Time) Proto() (*crpb.Time, error) {
	pbCQLTime := &crpb.Time{}
	pbTime := &timeofdaypb.TimeOfDay{}
	switch t.Precision {
	case model.HOUR:
		pbTime.Hours = int32(t.Date.Hour())
		pbCQLTime.Precision = crpb.Time_PRECISION_HOUR.Enum()
	case model.MINUTE:
		pbTime.Hours = int32(t.Date.Hour())
		pbTime.Minutes = int32(t.Date.Minute())
		pbCQLTime.Precision = crpb.Time_PRECISION_MINUTE.Enum()
	case model.SECOND:
		pbTime.Hours = int32(t.Date.Hour())
		pbTime.Minutes = int32(t.Date.Minute())
		pbTime.Seconds = int32(t.Date.Second())
		pbCQLTime.Precision = crpb.Time_PRECISION_SECOND.Enum()
	case model.MILLISECOND:
		pbTime.Hours = int32(t.Date.Hour())
		pbTime.Minutes = int32(t.Date.Minute())
		pbTime.Seconds = int32(t.Date.Second())
		pbTime.Nanos = int32(t.Date.Nanosecond())
		pbCQLTime.Precision = crpb.Time_PRECISION_MILLISECOND.Enum()
	default:
		return nil, fmt.Errorf("unsupported precision converting proto to Time %v %w", t.Precision, datehelpers.ErrUnsupportedPrecision)
	}

	pbCQLTime.Date = pbTime
	return pbCQLTime, nil
}

// TimeFromProto converts a proto to a Time.
func TimeFromProto(pb *crpb.Time) (Time, error) {
	var modelPrecision model.DateTimePrecision
	switch pb.GetPrecision() {
	case crpb.Time_PRECISION_HOUR:
		modelPrecision = model.HOUR
	case crpb.Time_PRECISION_MINUTE:
		modelPrecision = model.MINUTE
	case crpb.Time_PRECISION_SECOND:
		modelPrecision = model.SECOND
	case crpb.Time_PRECISION_MILLISECOND:
		modelPrecision = model.MILLISECOND
	default:
		return Time{}, fmt.Errorf("unsupported precision in Time with value %v %w", pb.GetPrecision(), datehelpers.ErrUnsupportedPrecision)
	}
	pbDate := pb.GetDate()
	return Time{Date: time.Date(0, time.January, 1, int(pbDate.GetHours()), int(pbDate.GetMinutes()), int(pbDate.GetSeconds()), int(pbDate.GetNanos()), time.UTC), Precision: modelPrecision}, nil
}

// Interval is the Golang representation of a CQL Interval.
type Interval struct {
	Low           Value
	High          Value
	LowInclusive  bool
	HighInclusive bool
	// StaticType is used for the RuntimeType() of the interval when the interval contains
	// only runtime nulls (meaning the runtime type cannot be reliably inferred).
	StaticType *types.Interval // Field not exported.
}

// Equal returns true if this Interval matches the provided one, otherwise false.
func (i Interval) Equal(v Interval) bool {
	if !i.StaticType.Equal(v.StaticType) {
		return false
	}
	if !i.Low.Equal(v.Low) || !i.High.Equal(v.High) || i.LowInclusive != v.LowInclusive || i.HighInclusive != v.HighInclusive || !i.StaticType.Equal(v.StaticType) {
		return false
	}
	return true
}

// Proto converts Interval to a proto.
func (i Interval) Proto() (*crpb.Interval, error) {
	typepb, err := i.StaticType.Proto()
	if err != nil {
		return nil, err
	}
	pbInterval := &crpb.Interval{StaticType: typepb}

	pbInterval.Low, err = i.Low.Proto()
	if err != nil {
		return nil, err
	}
	pbInterval.High, err = i.High.Proto()
	if err != nil {
		return nil, err
	}
	pbInterval.LowInclusive = proto.Bool(i.LowInclusive)
	pbInterval.HighInclusive = proto.Bool(i.HighInclusive)
	return pbInterval, nil
}

// IntervalFromProto converts a proto to an Interval.
func IntervalFromProto(pb *crpb.Interval) (Interval, error) {
	typ, err := types.IntervalFromProto(pb.GetStaticType())
	if err != nil {
		return Interval{}, err
	}

	low, err := NewFromProto(pb.Low)
	if err != nil {
		return Interval{}, err
	}

	high, err := NewFromProto(pb.High)
	if err != nil {
		return Interval{}, err
	}

	return Interval{Low: low, High: high, LowInclusive: pb.GetLowInclusive(), HighInclusive: pb.GetHighInclusive(), StaticType: typ}, nil
}

func inferIntervalType(i Interval) types.IType {
	if !IsNull(i.Low) {
		return &types.Interval{PointType: i.Low.RuntimeType()}
	}
	if !IsNull(i.High) {
		return &types.Interval{PointType: i.High.RuntimeType()}
	}
	// Fallback to static type
	return i.StaticType
}

func (i Interval) marshalJSON(t json.RawMessage) ([]byte, error) {
	low, err := i.Low.MarshalJSON()
	if err != nil {
		return nil, err
	}
	high, err := i.High.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		Type          json.RawMessage `json:"@type"`
		Low           json.RawMessage `json:"low"`
		High          json.RawMessage `json:"high"`
		LowInclusive  bool            `json:"lowClosed"`
		HighInclusive bool            `json:"highClosed"`
	}{
		Type:          t,
		Low:           low,
		High:          high,
		LowInclusive:  i.LowInclusive,
		HighInclusive: i.HighInclusive,
	})
}

// List is the Golang representation of a CQL List.
type List struct {
	Value []Value
	// StaticType is used for the RuntimeType() of the list when the list is empty.
	StaticType *types.List
}

// Proto converts List to a proto.
func (l List) Proto() (*crpb.List, error) {
	typepb, err := l.StaticType.Proto()
	if err != nil {
		return nil, err
	}
	pbList := &crpb.List{StaticType: typepb}
	for _, v := range l.Value {
		pb, err := v.Proto()
		if err != nil {
			return nil, err
		}
		pbList.Value = append(pbList.Value, pb)
	}
	return pbList, nil
}

// ListFromProto converts a proto to a List.
func ListFromProto(pb *crpb.List) (List, error) {
	l := List{Value: make([]Value, 0, len(pb.Value))}

	typ, err := types.ListFromProto(pb.GetStaticType())
	if err != nil {
		return List{}, err
	}
	l.StaticType = typ

	for _, pbv := range pb.Value {
		v, err := NewFromProto(pbv)
		if err != nil {
			return List{}, err
		}
		l.Value = append(l.Value, v)
	}
	return l, nil
}

// Equal returns true if this List matches the provided one, otherwise false.
func (l List) Equal(v List) bool {
	if !l.StaticType.Equal(v.StaticType) {
		return false
	}
	if len(l.Value) != len(v.Value) {
		return false
	}
	for idx, obj := range l.Value {
		if !obj.Equal(v.Value[idx]) {
			return false
		}
	}
	return true
}

func inferListType(l []Value, staticType types.IType) types.IType {
	// The parser should have already done type inference and conversions according to
	// https://cql.hl7.org/03-developersguide.html#literals-and-selectors, if necessary for a List
	// literal without a type specifier.
	//
	// At runtime, we simply return the runtime type of the first element, or fall back to the
	// static type if the list is empty.
	// TODO(b/326277425): support mixed lists that may have a choice result type.
	if len(l) == 0 {
		// Because we fall back to a static type, this might be a choice type, even though mixed lists
		// are not fully supported yet.
		return staticType
	}
	return &types.List{ElementType: l[0].RuntimeType()}
}

// Named is the Golang representation fo a CQL Class (aka a CQL Named Structured Value). This could
// be any type defined in the Data Model like a FHIR.Encounter.
type Named struct {
	Value proto.Message
	// RuntimeType is the runtime type of this proto message value. Often times this is just the
	// same as the static named type, but in some cases (e.g. Choice types) the caller should resolve
	// this to the specific runtime type.
	RuntimeType *types.Named
}

// Proto converts Named to a proto.
func (n Named) Proto() (*crpb.Named, error) {
	a, err := anypb.New(n.Value)
	if err != nil {
		return nil, err
	}
	typepb, err := n.RuntimeType.Proto()
	if err != nil {
		return nil, err
	}
	return &crpb.Named{
		Value:       a,
		RuntimeType: typepb,
	}, nil
}

// NamedFromProto converts a proto to a Named.
func NamedFromProto(pb *crpb.Named) (Named, error) {
	return Named{}, errors.New("currently do not support converting a proto named type back into a golang value")
}

// Named types aren't called out in the spec yet so we are defining our own representation
// here for now.
func (n Named) marshalJSON(_ json.RawMessage) ([]byte, error) {
	v, err := protojson.Marshal(n.Value)
	if err != nil {
		return nil, err
	}

	return json.Marshal(struct {
		Type  types.IType     `json:"@type"`
		Value json.RawMessage `json:"value"`
	}{
		Type:  n.RuntimeType,
		Value: v,
	})
}

// Equal returns true if this Named matches the provided one, otherwise false.
func (n Named) Equal(v Named) bool {
	if !n.RuntimeType.Equal(v.RuntimeType) {
		return false
	}
	return proto.Equal(n.Value, v.Value)
}

// Tuple is the Golang representation of a CQL Tuple (aka a CQL Structured Value).
type Tuple struct {
	// Value is the map of element name to CQL Value.
	Value map[string]Value
	// RuntimeType could be a tuple type or if this was a Class instance could be the class type
	// (FHIR.Patient, System.Quantity...). For Choice types this should resolve to the specific
	// runtime type.
	RuntimeType types.IType
}

// Proto converts Tuple to a proto.
func (t Tuple) Proto() (*crpb.Tuple, error) {
	pbTuple := &crpb.Tuple{
		Value: make(map[string]*crpb.Value),
	}

	switch typ := t.RuntimeType.(type) {
	case *types.Tuple:
		typepb, err := typ.Proto()
		if err != nil {
			return nil, err
		}
		pbTuple.RuntimeType = &crpb.Tuple_TupleType{TupleType: typepb}
	case *types.Named:
		typepb, err := typ.Proto()
		if err != nil {
			return nil, err
		}
		pbTuple.RuntimeType = &crpb.Tuple_NamedType{NamedType: typepb}
	default:
		return nil, fmt.Errorf("converting to proto found unsupported tuple type %v", t.RuntimeType)
	}

	for k, v := range t.Value {
		pb, err := v.Proto()
		if err != nil {
			return nil, err
		}
		pbTuple.Value[k] = pb
	}
	return pbTuple, nil
}

// TupleFromProto converts a proto to a Tuple.
func TupleFromProto(pb *crpb.Tuple) (Tuple, error) {
	cqlTuple := Tuple{Value: make(map[string]Value)}
	switch pbt := pb.RuntimeType.(type) {
	case *crpb.Tuple_TupleType:
		tupleType, err := types.TupleFromProto(pbt.TupleType)
		if err != nil {
			return Tuple{}, err
		}
		cqlTuple.RuntimeType = tupleType
	case *crpb.Tuple_NamedType:
		namedType, err := types.NamedFromProto(pbt.NamedType)
		if err != nil {
			return Tuple{}, err
		}
		cqlTuple.RuntimeType = namedType
	default:
		return Tuple{}, fmt.Errorf("converting from proto found unsupported tuple type %v", pb.RuntimeType)
	}

	for k, v := range pb.Value {
		v, err := NewFromProto(v)
		if err != nil {
			return Tuple{}, err
		}
		cqlTuple.Value[k] = v
	}
	return cqlTuple, nil
}

// Equal returns true if this Tuple matches the provided one, otherwise false.
func (t Tuple) Equal(vTuple Tuple) bool {
	if !t.RuntimeType.Equal(vTuple.RuntimeType) {
		return false
	}
	if len(t.Value) != len(vTuple.Value) {
		return false
	}
	for k, v := range t.Value {
		if !v.Equal(vTuple.Value[k]) {
			return false
		}
	}
	return true
}

// ValueSet is the Golang representation of a CQL ValueSet.
type ValueSet struct {
	ID      string // 1..1
	Version string // 0..1
	// Unlike the CQL reference we are not including the local name as it is not considered useful.
	CodeSystems []CodeSystem // 0..*
}

// Equal returns true if this ValueSet matches the provided one, otherwise false.
func (v ValueSet) Equal(a ValueSet) bool {
	if v.ID != a.ID ||
		v.Version != a.Version ||
		len(v.CodeSystems) != len(a.CodeSystems) {
		return false
	}
	slices.SortFunc(v.CodeSystems, compareCodeSystem)
	slices.SortFunc(a.CodeSystems, compareCodeSystem)
	for i, c := range a.CodeSystems {
		if c != v.CodeSystems[i] {
			return false
		}
	}
	return true
}

// Proto converts ValueSet to a proto.
func (v ValueSet) Proto() *crpb.ValueSet {
	pbValueSet := &crpb.ValueSet{
		Id:          proto.String(v.ID),
		Version:     proto.String(v.Version),
		CodeSystems: make([]*crpb.CodeSystem, 0, len(v.CodeSystems)),
	}
	for _, c := range v.CodeSystems {
		pbValueSet.CodeSystems = append(pbValueSet.CodeSystems, c.Proto())
	}
	return pbValueSet
}

// ValueSetFromProto converts a proto to a ValueSet.
func ValueSetFromProto(pb *crpb.ValueSet) ValueSet {
	codeSystems := make([]CodeSystem, 0, len(pb.CodeSystems))
	for _, c := range pb.CodeSystems {
		codeSystems = append(codeSystems, CodeSystemFromProto(c))
	}
	return ValueSet{ID: pb.GetId(), Version: pb.GetVersion(), CodeSystems: codeSystems}
}

// TODO: b/301606416 - Need to be able to output ValueSet name.
func (v ValueSet) marshalJSON(runtimeType json.RawMessage) ([]byte, error) {
	var cs []byte
	if len(v.CodeSystems) > 0 {
		var err error
		if cs, err = json.Marshal(v.CodeSystems); err != nil {
			return nil, err
		}
	}

	return json.Marshal(struct {
		Type        json.RawMessage `json:"@type"`
		ID          string          `json:"id"`
		Version     string          `json:"version,omitempty"`
		CodeSystems json.RawMessage `json:"codesystems,omitempty"`
	}{
		Type:        runtimeType,
		ID:          v.ID,
		Version:     v.Version,
		CodeSystems: cs,
	})
}

// CodeSystem is the Golang representation of a CQL CodeSystem.
type CodeSystem struct {
	ID      string // 1..1
	Version string // 0..1
	// Unlike the CQL reference we are not including the local name as it is not considered useful.
}

// TODO: b/301606416 - Need to be able to output CodeSystem name.
func (c CodeSystem) marshalJSON(runtimeType json.RawMessage) ([]byte, error) {
	return json.Marshal(struct {
		Type    json.RawMessage `json:"@type"`
		ID      string          `json:"id"`
		Version string          `json:"version,omitempty"`
	}{
		Type:    runtimeType,
		ID:      c.ID,
		Version: c.Version,
	})
}

func compareCodeSystem(a, b CodeSystem) int {
	if a.ID != b.ID {
		return strings.Compare(a.ID, b.ID)
	}
	return strings.Compare(a.Version, b.Version)
}

// Proto converts CodeSystem to a proto.
func (c CodeSystem) Proto() *crpb.CodeSystem {
	return &crpb.CodeSystem{
		Id:      proto.String(c.ID),
		Version: proto.String(c.Version),
	}
}

// CodeSystemFromProto converts a proto to a CodeSystem.
func CodeSystemFromProto(pb *crpb.CodeSystem) CodeSystem {
	return CodeSystem{ID: pb.GetId(), Version: pb.GetVersion()}
}

// Concept is the Golang representation of a CQL Concept.
type Concept struct {
	Codes   []Code // 1..*
	Display string // 0..1
}

// Equal returns true if this Concept matches the provided one, otherwise false.
func (c Concept) Equal(v Concept) bool {
	if len(c.Codes) != len(v.Codes) || c.Display != v.Display {
		return false
	}
	slices.SortFunc(c.Codes, compareCode)
	slices.SortFunc(v.Codes, compareCode)
	for i, c := range c.Codes {
		if c != v.Codes[i] {
			return false
		}
	}
	return true
}

// Proto converts Concept to a proto.
func (c Concept) Proto() *crpb.Concept {
	pbConcept := &crpb.Concept{
		Display: proto.String(c.Display),
		Codes:   make([]*crpb.Code, 0, len(c.Codes)),
	}
	for _, code := range c.Codes {
		pbConcept.Codes = append(pbConcept.Codes, code.Proto())
	}
	return pbConcept
}

// ConceptFromProto converts a proto to a Concept.
func ConceptFromProto(pb *crpb.Concept) Concept {
	codes := make([]Code, 0, len(pb.Codes))
	for _, c := range pb.Codes {
		codes = append(codes, CodeFromProto(c))
	}
	return Concept{Codes: codes, Display: pb.GetDisplay()}
}

func (c Concept) marshalJSON(runtimeType json.RawMessage) ([]byte, error) {
	codeType, err := types.Code.MarshalJSON()
	if err != nil {
		return nil, err
	}
	var codes []json.RawMessage
	for _, code := range c.Codes {
		code, err := code.marshalJSON(codeType)
		if err != nil {
			return nil, err
		}
		codes = append(codes, code)
	}

	return json.Marshal(struct {
		Type    json.RawMessage   `json:"@type"`
		Codes   []json.RawMessage `json:"codes"`
		Display string            `json:"display,omitempty"`
	}{
		Type:    runtimeType,
		Codes:   codes,
		Display: c.Display,
	})
}

// Code is the Golang representation of a CQL Code.
type Code struct {
	Code    string // 1..1
	Display string // 0..1
	System  string // 0..1
	Version string // 0..1
}

func (c Code) marshalJSON(runtimeType json.RawMessage) ([]byte, error) {
	return json.Marshal(struct {
		Type    json.RawMessage `json:"@type"`
		Code    string          `json:"code"`
		Display string          `json:"display,omitempty"`
		System  string          `json:"system"`
		Version string          `json:"version,omitempty"`
	}{
		Type:    runtimeType,
		Code:    c.Code,
		Display: c.Display,
		System:  c.System,
		Version: c.Version,
	})
}

// compareCode is used for sorting for go Equal() implementation. This is different from CQL
// equality where display is ignored.
func compareCode(a, b Code) int {
	if a.Code != b.Code {
		return strings.Compare(a.Code, b.Code)
	} else if a.System != b.System {
		return strings.Compare(a.System, b.System)
	} else if a.Version != b.Version {
		return strings.Compare(a.Version, b.Version)
	}
	return strings.Compare(a.Display, b.Display)
}

// Proto converts Code to a proto.
func (c Code) Proto() *crpb.Code {
	return &crpb.Code{
		Code:    proto.String(c.Code),
		Display: proto.String(c.Display),
		System:  proto.String(c.System),
		Version: proto.String(c.Version),
	}
}

// CodeFromProto converts a proto to a Code.
func CodeFromProto(pb *crpb.Code) Code {
	return Code{Code: pb.GetCode(), Display: pb.GetDisplay(), System: pb.GetSystem(), Version: pb.GetVersion()}
}
