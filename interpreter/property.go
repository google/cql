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
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
	"unicode"

	d4pb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/datatypes_go_proto"

	"github.com/google/cql/internal/datehelpers"
	"github.com/google/cql/internal/modelinfo"
	"github.com/google/cql/model"
	"github.com/google/cql/result"
	"github.com/google/cql/types"
	"github.com/google/fhir/go/protopath"
	annotations_pb "github.com/google/fhir/go/proto/google/fhir/proto/annotations_go_proto"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// evalProperty evaluates the ELM property expression passed in.
func (i *interpreter) evalProperty(expr *model.Property) (result.Value, error) {
	if expr.Source == nil {
		return result.Value{}, fmt.Errorf("internal error - source must be populated when accessing property %s", expr.Path)
	}
	obj, err := i.evalExpression(expr.Source)
	if err != nil {
		return result.Value{}, err
	}
	if result.IsNull(obj) {
		return result.NewWithSources(nil, expr, obj)
	}
	// TODO(b/315503615): if Element or result type is unset, error in the future.
	subObj, err := i.valueProperty(obj, expr.Path, expr.GetResultType())
	if err != nil {
		return result.Value{}, err
	}
	return subObj.WithSources(expr, obj), nil
}

// valueProperty computes the specified property on the given result.Value.
func (i *interpreter) valueProperty(v result.Value, property string, staticResultType types.IType) (result.Value, error) {
	if property == "" {
		return v, nil
	}

	switch ot := v.GolangValue().(type) {
	case result.Tuple:
		elem, ok := ot.Value[property]
		if !ok {
			// The parser should have already validated that this is a valid property for the Tuple or
			// Class type. If is not set in map return null.
			return result.New(nil)
		}
		return elem, nil
	case result.Named:
		return i.protoProperty(ot, property, staticResultType)
	case result.List:
		return i.listProperty(ot, property, staticResultType)
	case result.Interval:
		switch property {
		case "low":
			return ot.Low, nil
		case "high":
			return ot.High, nil
		case "lowClosed":
			return result.New(ot.LowInclusive)
		case "highClosed":
			return result.New(ot.HighInclusive)
		default:
			return result.Value{}, fmt.Errorf("property %s is not supported on Intervals", property)
		}
	case result.Quantity:
		switch property {
		case "value":
			return result.New(ot.Value)
		case "unit":
			return result.New(string(ot.Unit))
		default:
			return result.Value{}, fmt.Errorf("property %s is not supported on %v", property, types.Quantity)
		}
	case result.Code:
		switch property {
		case "code":
			return result.New(ot.Code)
		case "system":
			return result.New(ot.System)
		case "version":
			return result.New(ot.Version)
		case "display":
			return result.New(ot.Display)
		default:
			return result.Value{}, fmt.Errorf("property %s is not supported on %v", property, types.Code)
		}
	case result.Concept:
		switch property {
		case "codes":
			return result.New(ot.Codes)
		case "display":
			return result.New(ot.Display)
		default:
			return result.Value{}, fmt.Errorf("property %s is not supported on %v", property, types.Concept)
		}
	case result.ValueSet:
		switch property {
		case "id":
			return result.New(ot.ID)
		case "version":
			return result.New(ot.Version)
		default:
			return result.Value{}, fmt.Errorf("property %s is not supported on %v", property, types.ValueSet)
		}
	case result.CodeSystem:
		switch property {
		case "id":
			return result.New(ot.ID)
		case "version":
			return result.New(ot.Version)
		default:
			return result.Value{}, fmt.Errorf("property %s is not supported on %v", property, types.CodeSystem)
		}
		// TODO(b/301606416): Support Ratio and Vocabulary properties.
	default:
		return result.Value{}, fmt.Errorf("unable to eval property %s on unsupported type %v", property, ot)
	}
}

func protoFieldFromJSONName(p proto.Message, jsonProperty string) (string, error) {
	fd := p.ProtoReflect().Descriptor().Fields().ByJSONName(jsonProperty)
	if fd != nil {
		return fd.TextName(), nil
	}
	return "", fmt.Errorf("proto json field name %s not found", jsonProperty)
}

func (i *interpreter) protoProperty(source result.Named, property string, staticResultType types.IType) (result.Value, error) {
	// For .value properties on FHIR.dateTime, FHIR.time, FHIR.date, the result type is expected to be
	// a System.DateTime, System.Time, System.Date. This is not how the data is represented in the
	// FHIR proto data model, so we must catch this case and apply manual conversion.

	if property == "value" && (source.RuntimeType.Equal(&types.Named{TypeName: "FHIR.dateTime"}) ||
		source.RuntimeType.Equal(&types.Named{TypeName: "FHIR.time"}) ||
		source.RuntimeType.Equal(&types.Named{TypeName: "FHIR.date"})) {
		return handleDateTimeValueProperty(source.Value, property, i.evaluationTimestamp.Location())
	}

	protoProperty, err := protoFieldFromJSONName(source.Value, property)
	if err != nil {
		return result.Value{}, err
	}
	subAny, err := protopath.Get[any](source.Value, protopath.NewPath(protoProperty))
	if err != nil {
		return result.Value{}, err
	}

	if reflect.ValueOf(subAny).Kind() == reflect.Slice {
		return sliceToValue(reflect.ValueOf(subAny), staticResultType)
	}

	switch s := subAny.(type) {
	case proto.Message:
		return handleProtoValue(s, property, staticResultType, i.modelInfo)
	case protoreflect.Enum:
		return handleEnumValue(s)
	}

	obj, err := result.New(subAny)
	if err != nil {
		return result.Value{}, fmt.Errorf("error at property %s: %w", property, err)
	}
	return obj, nil
}

func handleEnumValue(e protoreflect.Enum) (result.Value, error) {
	// We need to transform this into a System.String.
	// Roughly, we follow the pattern here:
	// https://github.com/google/fhir/blob/358f39a5fae0faa006c9a630ca113f03d191e929/go/jsonformat/marshaller.go#L797-L811
	ed := e.Descriptor()
	ev := ed.Values().ByNumber(e.Number())
	origCode := proto.GetExtension(ev.Options(), annotations_pb.E_FhirOriginalCode).(string)
	if origCode != "" {
		return result.New(origCode)
	}
	rawEnumName := string(ev.Name())
	return result.New(strings.Replace(strings.ToLower(rawEnumName), "_", "-", -1))
}

func handleProtoValue(msg proto.Message, property string, staticResultType types.IType, mi *modelinfo.ModelInfos) (result.Value, error) {
	if !msg.ProtoReflect().IsValid() {
		return result.New(nil)
	}

	// runtimeResultType is the runtime result of the property result. We default to the static type,
	// but in cases (like if the property is on a oneof) this might be updated to a more specific type
	// (e.g. for a property on a oneof, the staticResultType will be a Choice, but the
	// runtimeResultType will be updated to be a specific type based on the set oneof).
	runtimeResultType := staticResultType

	// Check to see if this is a oneof wrapper message, and if so, extract the set oneof type. This is
	// an artifact of how the FHIR protos are generated, that a choice type field like
	// "Observation.value" will have a wrapper message called ValueX and ValueX.choice will contain
	// the actual proto oneof.
	// The protopath library will do this extraction for us, if we access the 'choice' property, which
	// is the field name of all oneof wrapper types in the R4 generated protos.
	// Note if the extension does not exist, this will be false:
	msgReflect := msg.ProtoReflect()
	if proto.GetExtension(msgReflect.Descriptor().Options(), annotations_pb.E_IsChoiceType).(bool) {
		// TODO(b/316960208): When choice types are supported, we can know in advance if a property
		// evaluation will result in a choice type wrapper, and could create a single call
		// to protopath.Get by appending to the protopath.Path.
		oneofValue, err := protopath.Get[any](msg, protopath.NewPath("choice"))
		if err != nil {
			return result.Value{}, err
		}

		// Oneofs should not have repeated fields in FHIR, and we don't know how to compute the runtime
		// type for them. https://hl7.org/fhir/R4/formats.html#choice
		// TODO(b/324240909): support computing the runtime type correctly if needed in the future for
		// other data models or future versions of FHIR.
		if reflect.ValueOf(oneofValue).Kind() == reflect.Slice {
			return result.Value{}, fmt.Errorf("cannot access this oneof proto submessage at property %v because the result %T is a slice, and no repeated elements were expected inside a choice type", property, oneofValue)
		}

		// Since this is a oneof, use the set oneof field to determine the specific named type of the
		// result.
		runtimeResultType, err = fhirOneofRuntimeType(msgReflect, property, mi)
		if err != nil {
			return result.Value{}, err
		}

		var ok bool
		msg, ok = oneofValue.(proto.Message)
		if !ok {
			// TODO(b/316960208): We still expect oneof values to be proto.Message for now. Consider
			// circling back.
			return result.Value{}, fmt.Errorf("cannot access this oneof proto submessage at property %v because %T is not a proto.Message", property, oneofValue)
		}
	}
	namedResultType, ok := runtimeResultType.(*types.Named)
	if !ok {
		return result.Value{}, fmt.Errorf("internal error - handleProtoValue expected runtimeResultType to be a Named type, got: %v", runtimeResultType)
	}
	return result.New(result.Named{Value: msg, RuntimeType: namedResultType})
}

// handleDateTimeValueProperty computes the value property for FHIR.date and FHIR.dateTime. It will
// eventually support FHIR.time when support for it is added.
func handleDateTimeValueProperty(sourceProto proto.Message, property string, evaluationLoc *time.Location) (result.Value, error) {
	if property != "value" {
		return result.Value{}, fmt.Errorf("internal error - handleTimeProperty expected a value property, got: %v", property)
	}
	switch v := sourceProto.(type) {
	case *d4pb.Date:
		t, prec, err := datehelpers.ParseFHIRDate(v, evaluationLoc)
		if err != nil {
			return result.Value{}, err
		}
		return result.New(result.Date{
			Date:      t,
			Precision: prec,
		})
	case *d4pb.DateTime:
		t, prec, err := datehelpers.ParseFHIRDateTime(v, evaluationLoc)
		if err != nil {
			return result.Value{}, err
		}
		return result.New(result.DateTime{
			Date:      t,
			Precision: prec,
		})
	}
	// TODO: b/301606416 - add support for FHIR.time.
	return result.Value{}, fmt.Errorf("internal error - handleTimeValueProperty got an unexpected source value type: %T", sourceProto)
}

// fhirOneofRuntimeType computes the runtime type of the property result on the oneofWrapperMsg, by
// inspecting the set oneof field of sourceMsg. This convention may apply to other data models,
// but for now is FHIR specific.
func fhirOneofRuntimeType(oneofWrapperMsg protoreflect.Message, property string, mi *modelinfo.ModelInfos) (*types.Named, error) {
	if oneofWrapperMsg.Descriptor().Oneofs().Len() != 1 {
		// All oneof wrapper types should have only one contained oneof type.
		return nil, fmt.Errorf("internal error - cannot access this proto submessage at property %v because the wrapper oneof type has multiple oneof fields", property)
	}
	oneofField := oneofWrapperMsg.WhichOneof(oneofWrapperMsg.Descriptor().Oneofs().Get(0))
	// fieldName should be similar to the FHIR convention, but without the "nnn" prefix and without
	// the title case described here:
	// https://hl7.org/fhir/R4/formats.html#choice.
	fieldName := oneofField.JSONName()
	// Try to use the JSON name as is, and if it doesn't validate, we may need to capitalize the first
	// letter since some FHIR types are lowercase (FHIR.dateTime) but others are title cased.
	t, err := mi.ToNamed(fieldName)
	if errors.Is(err, modelinfo.ErrTypeNotFound) {
		// Try with title case.
		titledFieldName := []rune(fieldName)
		titledFieldName[0] = unicode.ToUpper(titledFieldName[0])
		return mi.ToNamed(string(titledFieldName))
	}
	return t, err
}

func (i *interpreter) listProperty(l result.List, property string, staticResultType types.IType) (result.Value, error) {
	// The result type should be a list, so let's check that and grab the element type.
	resultListType, ok := staticResultType.(*types.List)
	if !ok {
		return result.Value{}, fmt.Errorf("internal error -- evalPropertyList expects a staticResultType of list, got :%v", staticResultType.String())
	}
	var subList []result.Value
	for idx, elem := range l.Value {
		// To compute a property on a list, we compute the property on each element (elem) in the list
		// and return the combined result list. In cases where the output is a list of lists, the inner
		// lists are later flattened. Because of this, it is possible that the property evaluation for a
		// given element will result in a runtime list but the parser resultListType.ElementType will
		// _not_ be a list for properties that were nested, because the flattening happens after the
		// element property computation. This flattening is defined
		// in https://build.fhir.org/ig/HL7/cql/03-developersguide.html#path-traversal and implemented
		// in the parser static type computation in the internal/modelinfo package.
		//
		// For evalPropertyValue(elem, elemResultType), we want to ensure the passed elemResultType
		// will match the interim runtime type, even in cases where flattening may be later applied. To
		// ensure this, we recompute the property result type using the type helper directly on the list
		// element (elem.property) instead of relying on the parser resultType, where flattening may
		// have been applied.
		//
		// For example, consider the property name.given, where both name and given are repeated:
		// name: [{given: ["a", "b"]}, {given: ["c", "d"]}]
		// Evaluating ".given" should result in ["a", "b", "c", "d"], a flattened list in CQL for the
		// completed computation of evalPropertyList. The parser result type would be List<String>, with
		// the element type being String. However, inside this loop we compute the property
		// ".given" for each input name in the list. When computing a property on a single name element
		// inside this loop (e.g. {given: ["a", "b"]}) the property should result in a runtime slice
		// (["a", "b"]) for each element, so we must ensure we actually pass List<String> for this
		// element result type instead of just String, which would be the parser result type's element
		// type.
		elemResultType, err := i.modelInfo.PropertyTypeSpecifier(elem.RuntimeType(), property)
		if err != nil {
			return result.Value{}, err
		}
		subObj, err := i.valueProperty(elem, property, elemResultType)
		if err != nil {
			return result.Value{}, fmt.Errorf("at index %d: %w", idx, err)
		}

		isSub, err := i.modelInfo.IsSubType(subObj.RuntimeType(), &types.List{ElementType: types.Any})
		if err != nil {
			return result.Value{}, err
		}
		if isSub {
			// When accessing repeated fields such as Patient.name.given we want to return a list of all
			// given's in all names. This flattens the givens into a single list.
			subList = append(subList, subObj.GolangValue().(result.List).Value...)
		} else {
			subList = append(subList, subObj)
		}
	}
	return result.New(result.List{Value: subList, StaticType: resultListType})
}

// sliceToValue takes a slice of arbitrary Golang values and converts it to a properly typed
// *result.List Value. This means that it will ensure that the Golang elements of the input slice
// are properly converted to result.Values with the right type, based on the expected result type
// of the overall slice (listType). For example, a Golang slice of []*r4pb.HumanName with listType
// of *types.List{ElementType: "FHIR.HumanName"} would be properly converted to a result.List
// with the right type, and all the List elements would be correctly added as
// result.Named{..., Type: types.Named{"FHIR.HumanName"}} with the correct static result type
// annotated.
// TODO(b/315503615): would it be helpful to other callers to have this be in result package and
// exported?
func sliceToValue(v reflect.Value, staticResultType types.IType) (result.Value, error) {
	if v.Kind() != reflect.Slice {
		return result.Value{}, fmt.Errorf("%T: not a slice", v)
	}

	listType, ok := staticResultType.(*types.List)
	if !ok {
		return result.Value{}, fmt.Errorf("internal error -- sliceToValue expects a staticResultType of list, got :%v", staticResultType.String())
	}
	elementType, ok := listType.ElementType.(*types.Named)
	if !ok {
		return result.Value{}, fmt.Errorf("internal error -- sliceToValue expects a staticResultType of list with named elements, got :%v", staticResultType.String())
	}

	l := make([]result.Value, v.Len())
	for i := 0; i < v.Len(); i++ {
		val := v.Index(i).Interface()
		switch typedVal := val.(type) {
		case proto.Message:
			// TODO(b/315503615): apply proto reflection based type determination when that is added,
			// to resolve choice types to their specific runtime type.
			o, err := result.New(result.Named{Value: typedVal, RuntimeType: elementType})
			if err != nil {
				return result.Value{}, fmt.Errorf("unable to create Value at index %d: %w", i, err)
			}
			l[i] = o
			continue
			// TODO: add interval
		}
		if reflect.ValueOf(val).Kind() == reflect.Slice {
			// This means we have a nested list. This would happen when computing "value" property on
			// something like {"value": [[1,2], [3,4]]}.
			// This should be exceedingly rare in FHIR land, but could happen in CQL.
			// We don't support mixed lists yet, so assume the elemType is itself a list
			// and not a choice.
			innerList, ok := listType.ElementType.(*types.List)
			if !ok {
				return result.Value{}, fmt.Errorf("internal error -- sliceToValue got element value of type Slice, so expected it to be a list but got :%v", listType.ElementType)
			}
			o, err := sliceToValue(reflect.ValueOf(val), innerList)
			if err != nil {
				return result.Value{}, fmt.Errorf("unable to create Value at index %d: %w", i, err)
			}
			l[i] = o
			continue
		}

		// Other primitive types:
		o, err := result.New(val)
		if err != nil {
			return result.Value{}, fmt.Errorf("unable to create Value at index %d: %w", i, err)
		}
		l[i] = o
	}
	return result.New(result.List{Value: l, StaticType: listType})
}

func datePrecisionFromProto(p d4pb.Date_Precision) model.DateTimePrecision {
	switch p {
	case d4pb.Date_YEAR:
		return model.YEAR
	case d4pb.Date_MONTH:
		return model.MONTH
	case d4pb.Date_DAY:
		return model.DAY
	}
	return model.UNSETDATETIMEPRECISION
}

func dateTimePrecisionFromProto(p d4pb.DateTime_Precision) model.DateTimePrecision {
	switch p {
	case d4pb.DateTime_YEAR:
		return model.YEAR
	case d4pb.DateTime_MONTH:
		return model.MONTH
	case d4pb.DateTime_DAY:
		return model.DAY
	case d4pb.DateTime_SECOND:
		return model.SECOND
	case d4pb.DateTime_MILLISECOND:
		return model.MILLISECOND
	// FHIR datetimes can have microsecond precision, since CQL doesn't support this we map it to millisecond.
	case d4pb.DateTime_MICROSECOND:
		return model.MILLISECOND
	}
	return model.UNSETDATETIMEPRECISION
}
