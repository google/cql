// Copyright 2023 Google LLC
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

// Package types holds a representation of CQL types and related logic. It is used by both the
// parser and interpreter.
package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"google.golang.org/protobuf/proto"

	ctpb "github.com/google/cql/protos/cql_types_go_proto"
)

// IType is an interface implemented by all CQL Type structs.
type IType interface {
	// TODO(b/312172420): Consider renaming IType to Type or types.Specifier.

	// Equal is a strict equal. X.Equal(Y) is true when X and Y are the exact same types.
	Equal(IType) bool

	// String returns a print friendly representation of the type and implements fmt.Stringer.
	String() string

	// ModelInfoName returns the key for this type in the model info.
	//
	// For Named and System types this is the fully qualified name like FHIR.EnrollmentResponseStatus
	// or System.Integer. The name may be split in the XML (namespace="FHIR"
	// name="EnrollmentResponseStatus"), but we use the qualified name.
	//
	// For other types the CQL type specifier syntax is used as the key. For example,
	// Interval<Integer>, Choice<Integer, String> or Tuple { address String }. For Tuple and Choice types
	// we alphabetically sort their inner types.
	ModelInfoName() (string, error)

	// MarshalJSON implements the json.Marshaler interface for the IType.
	MarshalJSON() ([]byte, error)
}

// System represents the primitive types defined by CQL
// (https://cql.hl7.org/09-b-cqlreference.html#types-2).
type System string

const (
	// Unset indicates that the parser did not set this Result Type.
	Unset System = "System.UnsetType"
	// Any is a CQL Any Type. It means that the type could be anything including list, interval,
	// named type from ModelInfo or null.
	Any System = "System.Any"
	// String is a CQL String type.
	String System = "System.String"
	// Integer is a CQL Integer type.
	Integer System = "System.Integer"
	// Decimal is a CQL Decimal type.
	Decimal System = "System.Decimal"
	// Long is a CQL Long type.
	Long System = "System.Long"
	// Quantity is a CQL decimal value and unit pair.
	Quantity System = "System.Quantity"
	// Ratio is the type for ratio - two CQL quantities.
	Ratio System = "System.Ratio"
	// Boolean is a CQL Boolean type.
	Boolean System = "System.Boolean"
	// DateTime is the CQL DateTime type.
	DateTime System = "System.DateTime"
	// Date is the CQL Date type.
	Date System = "System.Date"
	// Time is the CQL Time type.
	Time System = "System.Time"
	// ValueSet is the CQL Valueset type.
	ValueSet System = "System.ValueSet"
	// CodeSystem is a CQL CodeSystem which contains external Code definitions.
	CodeSystem System = "System.CodeSystem"
	// Vocabulary is the CQL Vocabulary type which is the parent type of ValueSet and CodeSystem.
	Vocabulary System = "System.Vocabulary"
	// Code is the CQL System Code type (which is distinct from a FHIR code type).
	Code System = "System.Code"
	// Concept is the CQL System Concept type.
	Concept System = "System.Concept"
)

// ToSystem converts a string to a System type returning System.Unsupported if the string cannot be
// converted to a system type.
func ToSystem(s string) System {
	switch s {
	case "System.Any", "Any":
		return Any
	case "System.String", "String":
		return String
	case "System.Integer", "Integer":
		return Integer
	case "System.Decimal", "Decimal":
		return Decimal
	case "System.Long", "Long":
		return Long
	case "System.Quantity", "Quantity":
		return Quantity
	case "System.Ratio", "Ratio":
		return Ratio
	case "System.Boolean", "Boolean":
		return Boolean
	case "System.DateTime", "DateTime":
		return DateTime
	case "System.Date", "Date":
		return Date
	case "System.Time", "Time":
		return Time
	case "System.ValueSet", "ValueSet":
		return ValueSet
	case "System.CodeSystem", "CodeSystem":
		return CodeSystem
	case "System.Vocabulary", "Vocabulary":
		return Vocabulary
	case "System.Code", "Code":
		return Code
	case "System.Concept", "Concept":
		return Concept
	default:
		return Unset
	}
}

// Equal is a strict equal. X.Equal(Y) is true when X and Y are the exact same types.
func (s System) Equal(a IType) bool {
	aBase, ok := a.(System)
	if !ok {
		return false
	}

	return s == aBase
}

// String returns the model info based name for the type, and implements fmt.Stringer for easy
// printing.
func (s System) String() string {
	return string(s)
}

// ModelInfoName returns the fully qualified type name in the model info convention.
func (s System) ModelInfoName() (string, error) {
	return string(s), nil
}

// Proto returns the proto representation of the system type.
func (s System) Proto() *ctpb.SystemType {
	switch s {
	case Any:
		return &ctpb.SystemType{Type: ctpb.SystemType_TYPE_ANY.Enum()}
	case String:
		return &ctpb.SystemType{Type: ctpb.SystemType_TYPE_STRING.Enum()}
	case Integer:
		return &ctpb.SystemType{Type: ctpb.SystemType_TYPE_INTEGER.Enum()}
	case Decimal:
		return &ctpb.SystemType{Type: ctpb.SystemType_TYPE_DECIMAL.Enum()}
	case Long:
		return &ctpb.SystemType{Type: ctpb.SystemType_TYPE_LONG.Enum()}
	case Quantity:
		return &ctpb.SystemType{Type: ctpb.SystemType_TYPE_QUANTITY.Enum()}
	case Ratio:
		return &ctpb.SystemType{Type: ctpb.SystemType_TYPE_RATIO.Enum()}
	case Boolean:
		return &ctpb.SystemType{Type: ctpb.SystemType_TYPE_BOOLEAN.Enum()}
	case DateTime:
		return &ctpb.SystemType{Type: ctpb.SystemType_TYPE_DATE_TIME.Enum()}
	case Date:
		return &ctpb.SystemType{Type: ctpb.SystemType_TYPE_DATE.Enum()}
	case Time:
		return &ctpb.SystemType{Type: ctpb.SystemType_TYPE_TIME.Enum()}
	case ValueSet:
		return &ctpb.SystemType{Type: ctpb.SystemType_TYPE_VALUE_SET.Enum()}
	case CodeSystem:
		return &ctpb.SystemType{Type: ctpb.SystemType_TYPE_CODE_SYSTEM.Enum()}
	case Vocabulary:
		return &ctpb.SystemType{Type: ctpb.SystemType_TYPE_VOCABULARY.Enum()}
	case Code:
		return &ctpb.SystemType{Type: ctpb.SystemType_TYPE_CODE.Enum()}
	case Concept:
		return &ctpb.SystemType{Type: ctpb.SystemType_TYPE_CONCEPT.Enum()}
	default:
		return &ctpb.SystemType{Type: ctpb.SystemType_TYPE_UNSPECIFIED.Enum()}
	}
}

// SystemFromProto converts a proto to a System type.
func SystemFromProto(pb *ctpb.SystemType) System {
	switch pb.GetType() {
	case ctpb.SystemType_TYPE_ANY:
		return Any
	case ctpb.SystemType_TYPE_STRING:
		return String
	case ctpb.SystemType_TYPE_INTEGER:
		return Integer
	case ctpb.SystemType_TYPE_DECIMAL:
		return Decimal
	case ctpb.SystemType_TYPE_LONG:
		return Long
	case ctpb.SystemType_TYPE_QUANTITY:
		return Quantity
	case ctpb.SystemType_TYPE_RATIO:
		return Ratio
	case ctpb.SystemType_TYPE_BOOLEAN:
		return Boolean
	case ctpb.SystemType_TYPE_DATE_TIME:
		return DateTime
	case ctpb.SystemType_TYPE_DATE:
		return Date
	case ctpb.SystemType_TYPE_TIME:
		return Time
	case ctpb.SystemType_TYPE_VALUE_SET:
		return ValueSet
	case ctpb.SystemType_TYPE_CODE_SYSTEM:
		return CodeSystem
	case ctpb.SystemType_TYPE_VOCABULARY:
		return Vocabulary
	case ctpb.SystemType_TYPE_CODE:
		return Code
	case ctpb.SystemType_TYPE_CONCEPT:
		return Concept
	default:
		return Unset
	}
}

// MarshalJSON implements the json.Marshaler interface for the System type.
func (s System) MarshalJSON() ([]byte, error) {
	return defaultTypeNameJSON(s)
}

// Named defines a single type by name. The name refers to a type defined by ModelInfo.
type Named struct {
	// TypeName is the fully qualified name of the type.
	TypeName string
	// TODO(b/313949948): Named type's name is a qualified identifier, we may wish to also
	// consider storing those in a structured way.
}

// Equal is a strict equal. X.Equal(Y) is true when X and Y are the exact same types.
func (n *Named) Equal(a IType) bool {
	aName, ok := a.(*Named)
	if !ok {
		return false
	}

	// n or aName are unknown or unsupported.
	if n == nil || aName == nil {
		return n == aName
	}

	return aName.TypeName == n.TypeName
}

// String returns the model info based name for the type, and implements fmt.Stringer for easy
// printing.
func (n *Named) String() string {
	if n == nil {
		return "nil Named"
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "Named<%s>", string(n.TypeName))
	return sb.String()
}

// ModelInfoName returns the fully qualified type name in the model info convention.
func (n *Named) ModelInfoName() (string, error) {
	if n == nil {
		return "", errTypeNil
	}
	return n.TypeName, nil
}

// Proto returns the proto representation of the named type.
func (n *Named) Proto() (*ctpb.NamedType, error) {
	if n == nil {
		return nil, errTypeNil
	}
	return &ctpb.NamedType{TypeName: proto.String(n.TypeName)}, nil
}

// NamedFromProto converts a proto to a Named type.
func NamedFromProto(pb *ctpb.NamedType) (*Named, error) {
	if pb == nil {
		return nil, errTypeNil
	}
	return &Named{TypeName: pb.GetTypeName()}, nil
}

// MarshalJSON implements the json.Marshaler interface for the Named type.
func (n Named) MarshalJSON() ([]byte, error) {
	return defaultTypeNameJSON(&n)
}

// Interval defines the type for an interval.
type Interval struct {
	PointType IType
}

// Equal is a strict equal. X.Equal(Y) is true when X and Y are the exact same types.
func (i *Interval) Equal(a IType) bool {
	aInterval, ok := a.(*Interval)
	if !ok {
		return false
	}

	if i == nil || aInterval == nil {
		// TODO(b/301606416): Add a test to cover this case.
		return i == aInterval
	}
	if i.PointType == nil || aInterval.PointType == nil {
		return i.PointType == aInterval.PointType
	}

	return i.PointType.Equal(aInterval.PointType)
}

// String returns the model info based name for the type, and implements fmt.Stringer for easy
// printing.
func (i *Interval) String() string {
	if i == nil {
		return "nil Interval"
	}
	if i.PointType == nil {
		return "Interval<nil>"
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "Interval<%s>", string(i.PointType.String()))
	return sb.String()
}

// ModelInfoName returns name as the CQL interval type specifier.
func (i *Interval) ModelInfoName() (string, error) {
	if i == nil {
		return "", errTypeNil
	}
	if i.PointType == nil {
		return "", fmt.Errorf("internal error -- nil PointType for Interval")
	}
	it, err := i.PointType.ModelInfoName()
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "Interval<%s>", it)
	return sb.String(), nil
}

// Proto returns the proto representation of the interval type.
func (i *Interval) Proto() (*ctpb.IntervalType, error) {
	if i == nil {
		return nil, errTypeNil
	}
	if i.PointType == nil {
		return nil, fmt.Errorf("internal error -- nil PointType for Interval")
	}

	pointTypePB, err := CQLTypeToProto(i.PointType)
	if err != nil {
		return nil, err
	}

	return &ctpb.IntervalType{PointType: pointTypePB}, nil
}

// IntervalFromProto converts a proto to an Interval type.
func IntervalFromProto(pb *ctpb.IntervalType) (*Interval, error) {
	if pb == nil {
		return nil, errTypeNil
	}
	pointType, err := CQLTypeFromProto(pb.GetPointType())
	if err != nil {
		return nil, err
	}
	return &Interval{PointType: pointType}, nil
}

// MarshalJSON implements the json.Marshaler interface for the Interval type.
func (i Interval) MarshalJSON() ([]byte, error) {
	if i.PointType == nil {
		return []byte(`"Interval<` + Any.String() + `>"`), nil
	}
	return defaultTypeNameJSON(&i)
}

// List defines the type for a list.
type List struct {
	// The type of the elements in the list.
	ElementType IType
}

// Equal is a strict equal. X.Equal(Y) is true when X and Y are the exact same types.
func (l *List) Equal(a IType) bool {
	aList, ok := a.(*List)
	if !ok {
		return false
	}

	if l == nil || aList == nil {
		return l == aList
	}
	if l.ElementType == nil || aList.ElementType == nil {
		return l.ElementType == aList.ElementType
	}

	return l.ElementType.Equal(aList.ElementType)
}

// String returns the model info based name for the type, and implements fmt.Stringer for easy
// printing.
func (l *List) String() string {
	if l == nil {
		return "nil List"
	}
	if l.ElementType == nil {
		return "List<nil>"
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "List<%s>", string(l.ElementType.String()))
	return sb.String()
}

// ModelInfoName returns name as the CQL list type specifier.
func (l *List) ModelInfoName() (string, error) {
	if l == nil {
		return "", errTypeNil
	}
	if l.ElementType == nil {
		return "", fmt.Errorf("internal error - nil ElementType for List")
	}
	et, err := l.ElementType.ModelInfoName()
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "List<%s>", et)
	return sb.String(), nil
}

// Proto returns the proto representation of the list type.
func (l *List) Proto() (*ctpb.ListType, error) {
	if l == nil {
		return nil, errTypeNil
	}
	if l.ElementType == nil {
		return nil, fmt.Errorf("internal error - nil ElementType for List")
	}

	listTypePB, err := CQLTypeToProto(l.ElementType)
	if err != nil {
		return nil, err
	}

	return &ctpb.ListType{ElementType: listTypePB}, nil
}

// ListFromProto converts a proto to a List type.
func ListFromProto(pb *ctpb.ListType) (*List, error) {
	if pb == nil {
		return nil, errTypeNil
	}
	elementType, err := CQLTypeFromProto(pb.GetElementType())
	if err != nil {
		return nil, err
	}
	return &List{ElementType: elementType}, nil
}

// MarshalJSON implements the json.Marshaler interface for the List type.
func (l List) MarshalJSON() ([]byte, error) {
	if l.ElementType == nil {
		return []byte(`"List<` + Any.String() + `>"`), nil
	}
	return defaultTypeNameJSON(&l)
}

// Choice defines the type for a choice type.
type Choice struct {
	ChoiceTypes []IType
}

// Equal is a strict equal. X.Equal(Y) is true when X and Y are the exact same types.
func (c *Choice) Equal(a IType) bool {
	if c == nil || a == nil {
		return c == a
	}

	aChoice, ok := a.(*Choice)
	if !ok {
		return false
	}

	if len(aChoice.ChoiceTypes) != len(c.ChoiceTypes) {
		return false
	}

	// Order of the ChoiceTypes does not matter, so create a copied slice to pop from.
	cChoiceSet := make([]IType, len(c.ChoiceTypes))
	copy(cChoiceSet, c.ChoiceTypes)
	for _, aType := range aChoice.ChoiceTypes {
		for i, cType := range cChoiceSet {
			if cType.Equal(aType) {
				// Pop the index from cChoiceSet.
				cChoiceSet = append(cChoiceSet[:i], cChoiceSet[i+1:]...)
				break
			}
		}
	}
	if len(cChoiceSet) == 0 {
		return true
	}
	return false
}

// String returns the model info based name for the type, and implements fmt.Stringer for easy
// printing.
func (c *Choice) String() string {
	if c == nil {
		return "nil Choice"
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "Choice<%s>", ToStrings(c.ChoiceTypes))
	return sb.String()
}

// ModelInfoName returns name as the CQL choice type specifier with ChoiceTypes sorted.
func (c *Choice) ModelInfoName() (string, error) {
	if c == nil {
		return "", errTypeNil
	}
	if c.ChoiceTypes == nil {
		return "", fmt.Errorf("internal error - nil ChoiceTypes for Choice")
	}

	sortedNames := make([]string, 0, len(c.ChoiceTypes))
	for _, choice := range c.ChoiceTypes {
		name, err := choice.ModelInfoName()
		if err != nil {
			return "", err
		}
		sortedNames = append(sortedNames, name)
	}
	sort.Strings(sortedNames)

	var sb strings.Builder
	fmt.Fprint(&sb, "Choice<")
	for i, n := range sortedNames {
		if i > 0 {
			fmt.Fprint(&sb, ", ")
		}
		fmt.Fprint(&sb, n)
	}
	fmt.Fprint(&sb, ">")
	return sb.String(), nil
}

// Proto returns the proto representation of the choice type.
func (c *Choice) Proto() (*ctpb.ChoiceType, error) {
	if c == nil {
		return nil, errTypeNil
	}
	if c.ChoiceTypes == nil {
		return nil, fmt.Errorf("internal error - nil ChoiceTypes for Choice")
	}

	choicepb := &ctpb.ChoiceType{ChoiceTypes: make([]*ctpb.CQLType, 0, len(c.ChoiceTypes))}
	for _, choiceType := range c.ChoiceTypes {
		choiceTypePB, err := CQLTypeToProto(choiceType)
		if err != nil {
			return nil, err
		}
		choicepb.ChoiceTypes = append(choicepb.ChoiceTypes, choiceTypePB)
	}

	return choicepb, nil
}

// ChoiceFromProto converts a proto to a Choice type.
func ChoiceFromProto(pb *ctpb.ChoiceType) (*Choice, error) {
	if pb == nil {
		return nil, errTypeNil
	}
	choice := &Choice{ChoiceTypes: make([]IType, 0, len(pb.GetChoiceTypes()))}
	for _, choiceTypePB := range pb.GetChoiceTypes() {
		choiceType, err := CQLTypeFromProto(choiceTypePB)
		if err != nil {
			return nil, err
		}
		choice.ChoiceTypes = append(choice.ChoiceTypes, choiceType)
	}
	return choice, nil
}

// MarshalJSON implements the json.Marshaler interface for the Choice type.
func (c Choice) MarshalJSON() ([]byte, error) {
	if c.ChoiceTypes == nil {
		return []byte(`"Choice"`), nil
	}
	if len(c.ChoiceTypes) == 0 {
		return []byte(`"Choice<>"`), nil
	}
	return defaultTypeNameJSON(&c)
}

// Tuple defines the type for a tuple (aka Structured Value).
type Tuple struct {
	// ElementTypes is a map from element name to its type.
	ElementTypes map[string]IType
}

// Equal is a strict equal. X.Equal(Y) is true when X and Y are the exact same types.
func (t *Tuple) Equal(a IType) bool {
	if t == nil || a == nil {
		return t == a
	}

	aTuple, ok := a.(*Tuple)
	if !ok {
		return false
	}

	if len(aTuple.ElementTypes) != len(t.ElementTypes) {
		return false
	}

	for tName, tType := range t.ElementTypes {
		aType, ok := aTuple.ElementTypes[tName]
		if !ok {
			return false
		}
		if !aType.Equal(tType) {
			return false
		}
	}

	return true
}

// String returns the model info based name for the type, and implements fmt.Stringer for easy
// printing.
func (t *Tuple) String() string {
	if t == nil {
		return "nil Tuple"
	}
	if t.ElementTypes == nil {
		return "Tuple<nil>"
	}

	// Deterministically order the types by type name.
	elementKeys := make([]string, 0, len(t.ElementTypes))
	for name := range t.ElementTypes {
		elementKeys = append(elementKeys, name)
	}
	sort.Strings(elementKeys)

	var sb strings.Builder
	fmt.Fprint(&sb, "Tuple<")
	i := 0
	for _, name := range elementKeys {
		if i > 0 {
			fmt.Fprint(&sb, ", ")
		}
		if t == nil {
			fmt.Fprintf(&sb, "%s: nil", name)
		} else {
			fmt.Fprintf(&sb, "%s: %s", name, t.ElementTypes[name].String())
		}
		i++
	}
	fmt.Fprint(&sb, ">")
	return sb.String()
}

// ModelInfoName returns name as the CQL tuple type specifier with ElementTypes sorted by name.
func (t *Tuple) ModelInfoName() (string, error) {
	if t == nil {
		return "", errTypeNil
	}
	if t.ElementTypes == nil {
		return "", fmt.Errorf("internal error - nil ElementTypes for Tuple")
	}

	if len(t.ElementTypes) == 0 {
		return "Tuple { }", nil
	}

	sortedNames := make([]string, 0, len(t.ElementTypes))
	for name := range t.ElementTypes {
		sortedNames = append(sortedNames, name)
	}
	sort.Strings(sortedNames)

	var sb strings.Builder
	fmt.Fprint(&sb, "Tuple { ")
	for i, name := range sortedNames {
		if i > 0 {
			fmt.Fprint(&sb, ", ")
		}
		elemType, err := t.ElementTypes[name].ModelInfoName()
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&sb, "%s %s", name, elemType)
	}
	fmt.Fprint(&sb, " }")
	return sb.String(), nil
}

// Proto returns the proto representation of the tuple type.
func (t *Tuple) Proto() (*ctpb.TupleType, error) {
	if t == nil {
		return nil, errTypeNil
	}
	if t.ElementTypes == nil {
		return nil, fmt.Errorf("internal error - nil ElementTypes for Tuple")
	}

	tuplepb := &ctpb.TupleType{ElementTypes: make(map[string]*ctpb.CQLType, len(t.ElementTypes))}
	for elemName, elemType := range t.ElementTypes {
		elemTypePB, err := CQLTypeToProto(elemType)
		if err != nil {
			return nil, err
		}
		tuplepb.ElementTypes[elemName] = elemTypePB
	}

	return tuplepb, nil
}

// TupleFromProto converts a proto to a Tuple type.
func TupleFromProto(pb *ctpb.TupleType) (*Tuple, error) {
	if pb == nil {
		return nil, errTypeNil
	}
	tuple := &Tuple{ElementTypes: make(map[string]IType, len(pb.GetElementTypes()))}
	for elemName, elemTypePB := range pb.GetElementTypes() {
		elemType, err := CQLTypeFromProto(elemTypePB)
		if err != nil {
			return nil, err
		}
		tuple.ElementTypes[elemName] = elemType
	}
	return tuple, nil
}

// CQLTypeToProto returns the generic CQLType proto of the CQL type.
func CQLTypeToProto(typ IType) (*ctpb.CQLType, error) {
	switch t := typ.(type) {
	case System:
		return &ctpb.CQLType{Type: &ctpb.CQLType_SystemType{SystemType: t.Proto()}}, nil
	case *Named:
		namedpb, err := t.Proto()
		if err != nil {
			return nil, err
		}
		return &ctpb.CQLType{Type: &ctpb.CQLType_NamedType{NamedType: namedpb}}, nil
	case *Interval:
		intervalpb, err := t.Proto()
		if err != nil {
			return nil, err
		}
		return &ctpb.CQLType{Type: &ctpb.CQLType_IntervalType{IntervalType: intervalpb}}, nil
	case *List:
		listpb, err := t.Proto()
		if err != nil {
			return nil, err
		}
		return &ctpb.CQLType{Type: &ctpb.CQLType_ListType{ListType: listpb}}, nil
	case *Choice:
		choicepb, err := t.Proto()
		if err != nil {
			return nil, err
		}
		return &ctpb.CQLType{Type: &ctpb.CQLType_ChoiceType{ChoiceType: choicepb}}, nil
	case *Tuple:
		tuplepb, err := t.Proto()
		if err != nil {
			return nil, err
		}
		return &ctpb.CQLType{Type: &ctpb.CQLType_TupleType{TupleType: tuplepb}}, nil
	default:
		return nil, fmt.Errorf("internal error - unsupported type %v in CQLTypeToProto", t)
	}
}

// CQLTypeFromProto converts the generic CQLType proto to a CQL type.
func CQLTypeFromProto(pb *ctpb.CQLType) (IType, error) {
	switch t := pb.Type.(type) {
	case *ctpb.CQLType_SystemType:
		return SystemFromProto(t.SystemType), nil
	case *ctpb.CQLType_NamedType:
		return NamedFromProto(t.NamedType)
	case *ctpb.CQLType_IntervalType:
		return IntervalFromProto(t.IntervalType)
	case *ctpb.CQLType_ListType:
		return ListFromProto(t.ListType)
	case *ctpb.CQLType_ChoiceType:
		return ChoiceFromProto(t.ChoiceType)
	case *ctpb.CQLType_TupleType:
		return TupleFromProto(t.TupleType)
	default:
		return nil, fmt.Errorf("internal error - unsupported type %v in CQLTypeFromProto", t)
	}
}

// ToStrings returns a print friendly representation of the types.
func ToStrings(ts []IType) string {
	var sb strings.Builder
	for i, t := range ts {
		if i > 0 {
			fmt.Fprint(&sb, ", ")
		}
		if t == nil {
			fmt.Fprint(&sb, "nil")
		} else {
			fmt.Fprint(&sb, t.String())
		}
	}
	return sb.String()
}

// MarshalJSON implements the json.Marshaler interface for the Tuple type.
func (t Tuple) MarshalJSON() ([]byte, error) {
	if t.ElementTypes == nil {
		return json.Marshal("Tuple")
	}
	return defaultTypeNameJSON(&t)
}

// zero is a helper function to return the Zero value of a generic type T.
func zero[T any]() T {
	var zero T
	return zero
}

var errTypeNil = errors.New("internal error -- unsupported function call on a nil type")

func defaultTypeNameJSON(t IType) ([]byte, error) {
	modelInfoName, err := t.ModelInfoName()
	if err != nil {
		return nil, err
	}
	return []byte(`"` + modelInfoName + `"`), nil
}
