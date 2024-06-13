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
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/google/cql/model"
	"github.com/google/cql/result"
	"github.com/google/cql/terminology"
	"github.com/google/cql/types"
	dtpb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/datatypes_go_proto"
	r4pb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/resources/bundle_and_contained_resource_go_proto"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func (i *interpreter) evalExpression(elem model.IExpression) (result.Value, error) {
	switch elem := elem.(type) {
	case *model.Literal:
		return i.evalLiteral(elem)
	case *model.Quantity:
		return i.evalQuantity(elem)
	case *model.Ratio:
		return i.evalRatio(elem)
	case *model.List:
		return i.evalList(elem)
	case *model.Code:
		return i.evalCode(elem)
	case model.IUnaryExpression:
		return i.evalUnaryExpression(elem)
	case model.IBinaryExpression:
		return i.evalBinaryExpression(elem)
	case model.INaryExpression:
		return i.evalNaryExpression(elem)
	case *model.Retrieve:
		return i.evalRetrieve(elem)
	case *model.Property:
		return i.evalProperty(elem)
	case *model.Query:
		return i.evalQuery(elem)
	case *model.QueryLetRef:
		return i.evalQueryLetRef(elem)
	case *model.AliasRef:
		return i.evalAliasRef(elem)
	case *model.CodeSystemRef:
		return i.evalCodeSystemRef(elem)
	case *model.ValuesetRef:
		return i.evalValuesetRef(elem)
	case *model.ParameterRef:
		return i.evalParameterRef(elem)
	case *model.CodeRef:
		return i.evalCodeRef(elem)
	case *model.ConceptRef:
		return i.evalConceptRef(elem)
	case *model.ExpressionRef:
		return i.evalExpressionRef(elem)
	case *model.Interval:
		return i.evalInterval(elem)
	case *model.FunctionRef:
		return i.evalFunctionRef(elem)
	case *model.OperandRef:
		return i.evalOperandRef(elem)
	case *model.Tuple:
		return i.evalTuple(elem)
	case *model.Instance:
		return i.evalInstance(elem)
	case *model.IfThenElse:
		return i.evalIfThenElse(elem)
	case *model.Case:
		return i.evalCase(elem)
	case *model.MaxValue:
		return i.evalMaxValue(elem)
	case *model.MinValue:
		return i.evalMinValue(elem)
	case *model.Message:
		return i.evalMessage(elem)
	default:
		// TODO(b/297089208): Add support for line/col error report.
		return result.Value{}, fmt.Errorf("internal error - unsupported expression")
	}
}

// TODO b/324637028 - Implement all message types
func (i *interpreter) evalMessage(m *model.Message) (result.Value, error) {
	source, err := i.evalExpression(m.Source)
	if err != nil {
		return result.Value{}, err
	}

	cond, err := i.evalExpression(m.Condition)
	if err != nil {
		return result.Value{}, err
	}

	// Whether or not to emit the message value
	condVal, err := result.ToBool(cond)
	if err != nil {
		return result.Value{}, err
	}
	if !condVal {
		return source, nil
	}

	// Emit the message.
	t, err := i.evalExpression(m.Severity)
	if err != nil {
		return result.Value{}, err
	}
	messageType, err := result.ToString(t)
	if err != nil {
		return result.Value{}, err
	}
	severity, err := messageSeverity(messageType)
	if err != nil {
		return result.Value{}, err
	}

	code, err := i.evalExpression(m.Code)
	if err != nil {
		return result.Value{}, err
	}
	codeVal, err := result.ToString(code)
	if err != nil {
		return result.Value{}, err
	}

	message, err := i.evalExpression(m.Message)
	if err != nil {
		return result.Value{}, err
	}
	messageVal, err := result.ToString(message)
	if err != nil {
		return result.Value{}, err
	}

	// TODO b/301606416: Add support for model.TRACE.

	outString := fmt.Sprintf("%s %s: %s", severity, codeVal, messageVal)
	fmt.Println(outString)
	if severity == model.ERROR {
		errMsg := fmt.Sprintf("log error - Message with severity of type `Error` was called with message: %s", outString)
		return source, errors.New(errMsg)
	}
	return source, nil
}

func (i *interpreter) evalRetrieve(expr *model.Retrieve) (result.Value, error) {
	if i.retriever == nil {
		return result.Value{}, fmt.Errorf("retriever was not set")
	}
	url, err := i.modelInfo.URL()
	if err != nil {
		return result.Value{}, err
	}
	name := strings.Split(expr.DataType, fmt.Sprintf("{%s}", url))
	if len(name) != 2 {
		return result.Value{}, fmt.Errorf("Resource datatype (%s) did not contain the library uri (%s)", expr.DataType, url)
	}
	got, err := i.retriever.Retrieve(context.Background(), name[1])
	if err != nil {
		return result.Value{}, err
	}

	// Assume the retrieve result type should be a list:
	listResultType, ok := expr.ResultType.(*types.List)
	if !ok {
		return result.Value{}, fmt.Errorf("internal error - retrieve result type should be a list, got %v", expr.ResultType)
	}
	elemResultType, ok := listResultType.ElementType.(*types.Named)
	if !ok {
		return result.Value{}, fmt.Errorf("internal error - retrieve result type should be a list of named types, got %v", listResultType)
	}

	l := []result.Value{}
	for _, c := range got {
		r, err := unwrapContained(c)
		if err != nil {
			return result.Value{}, err
		}
		msg, err := result.New(result.Named{Value: r, RuntimeType: elemResultType})
		if err != nil {
			return result.Value{}, err
		}

		if expr.Codes != nil {
			// We must try to filter on the codes provided.
			if expr.CodeProperty == "" {
				return result.Value{}, fmt.Errorf("code property must be populated when filtering on codes")
			}
			propertyType, err := i.modelInfo.PropertyTypeSpecifier(msg.RuntimeType(), expr.CodeProperty)
			if err != nil {
				return result.Value{}, err
			}
			cc, err := i.valueProperty(msg, expr.CodeProperty, propertyType)
			if err != nil {
				return result.Value{}, err
			}
			// If this isn't a codeableConcept, this will result in an error.
			in, err := i.inValueSet(cc, expr.Codes)
			if err != nil {
				return result.Value{}, err
			}
			if in {
				l = append(l, msg)
			}
		} else {
			// If no code filtering, always add to the result set.
			l = append(l, msg)
		}

	}
	// TODO(b/311222838): Currently only adding matched items as support,
	// but should confirm this meets use case needs.
	return result.NewWithSources(result.List{Value: l, StaticType: listResultType}, expr, l...)
}

func (i *interpreter) inValueSet(codeableConcept result.Value, codes model.IExpression) (bool, error) {
	if result.IsNull(codeableConcept) {
		return false, nil
	}

	vr, ok := codes.(*model.ValuesetRef)
	if !ok {
		return false, fmt.Errorf("only ValueSet references are currently supported for valueset filtering")
	}

	protoVal, ok := codeableConcept.GolangValue().(result.Named)
	if !ok {
		return false, fmt.Errorf("internal error -- inValueSet: the input Value must be a result.Named. got: %s", reflect.ValueOf(codeableConcept).Type())
	}
	ccPB, ok := protoVal.Value.(*dtpb.CodeableConcept)
	if !ok {
		return false, fmt.Errorf("internal error -- the input proto Value must be a *dtpb.CodeableConcept type. got: %s", reflect.ValueOf(codeableConcept).Type())
	}

	vs, err := i.evalValuesetRef(vr)
	if err != nil {
		return false, err
	}

	vsv, ok := vs.GolangValue().(result.ValueSet)
	if !ok {
		return false, fmt.Errorf("internal error - expected a ValueSetValue instead got %v", reflect.ValueOf(vs.GolangValue()).Type())
	}

	for _, coding := range ccPB.GetCoding() {
		// TODO: b/331447080 - Convert to using system operators for evaluating valueset membership.
		in, err := i.terminologyProvider.AnyInValueSet([]terminology.Code{{System: coding.GetSystem().Value, Code: coding.GetCode().Value}}, vsv.ID, vsv.Version)
		if err != nil {
			return false, err
		}
		if in {
			return true, nil
		}
	}

	return false, nil
}

// unwrapContained returns the FHIR resource from within the ContainedResource.
func unwrapContained(r *r4pb.ContainedResource) (proto.Message, error) {
	if r == nil {
		return nil, fmt.Errorf("resource is nil")
	}

	rpb := r.ProtoReflect()
	oneof := rpb.Descriptor().Oneofs().ByName("oneof_resource")
	if oneof == nil {
		return nil, fmt.Errorf("failed to extract oneof")
	}
	fd := rpb.WhichOneof(oneof)
	if fd == nil {
		return nil, fmt.Errorf("no resource type was populated")
	}
	f := rpb.Get(fd)
	innerMsg, ok := f.Interface().(protoreflect.Message)
	if !ok {
		return nil, fmt.Errorf("inner resource is not a message")
	}
	return innerMsg.Interface(), nil
}

func (i *interpreter) evalQueryLetRef(a *model.QueryLetRef) (result.Value, error) {
	return i.refs.ResolveLocal(a.Name)
}

func (i *interpreter) evalAliasRef(a *model.AliasRef) (result.Value, error) {
	return i.refs.ResolveLocal(a.Name)
}

func (i *interpreter) evalOperandRef(a *model.OperandRef) (result.Value, error) {
	return i.refs.ResolveLocal(a.Name)
}

func (i *interpreter) evalCodeSystemRef(expr *model.CodeSystemRef) (result.Value, error) {
	if expr.LibraryName != "" {
		return i.refs.ResolveGlobal(expr.LibraryName, expr.Name)
	}
	return i.refs.ResolveLocal(expr.Name)
}

func (i *interpreter) evalValuesetRef(expr *model.ValuesetRef) (result.Value, error) {
	if expr.LibraryName != "" {
		return i.refs.ResolveGlobal(expr.LibraryName, expr.Name)
	}
	return i.refs.ResolveLocal(expr.Name)
}

func (i *interpreter) evalParameterRef(expr *model.ParameterRef) (result.Value, error) {
	if expr.LibraryName != "" {
		return i.refs.ResolveGlobal(expr.LibraryName, expr.Name)
	}
	return i.refs.ResolveLocal(expr.Name)
}

func (i *interpreter) evalCodeRef(expr *model.CodeRef) (result.Value, error) {
	if expr.LibraryName != "" {
		return i.refs.ResolveGlobal(expr.LibraryName, expr.Name)
	}
	return i.refs.ResolveLocal(expr.Name)
}

func (i *interpreter) evalConceptRef(expr *model.ConceptRef) (result.Value, error) {
	if expr.LibraryName != "" {
		return i.refs.ResolveGlobal(expr.LibraryName, expr.Name)
	}
	return i.refs.ResolveLocal(expr.Name)
}

func (i *interpreter) evalExpressionRef(expr *model.ExpressionRef) (result.Value, error) {
	if expr.LibraryName != "" {
		return i.refs.ResolveGlobal(expr.LibraryName, expr.Name)
	}
	return i.refs.ResolveLocal(expr.Name)
}

// applyToValues is a convenience wrapper that invokes fn on both Values. If an error is returned
// for either invocation, it is returned, otherwise the results are returned.
func applyToValues[T any](l, r result.Value, fn func(result.Value) (T, error)) (T, T, error) {
	lObj, err := fn(l)
	if err != nil {
		return *new(T), *new(T), err
	}
	rObj, err := fn(r)
	if err != nil {
		return *new(T), *new(T), err
	}
	return lObj, rObj, nil
}

func messageSeverity(s string) (model.MessageSeverity, error) {
	switch s {
	case "Error":
		return model.ERROR, nil
	case "Message":
		return model.MESSAGE, nil
	case "Trace":
		return model.TRACE, nil
	case "Warning":
		return model.WARNING, nil
	default:
		return model.UNSETMESSAGESEVERITY, fmt.Errorf("invalid message severity %s", s)
	}
}
