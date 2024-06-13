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

package modelinfo

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/cql/internal/embeddata"
	"github.com/google/cql/types"
)

// ModelInfos provides methods for interacting with the underlying ModelInfos such as getting type
// specifiers for a property or determining implicit conversions. We currently only support the
// system and one custom ModelInfo. https://cql.hl7.org/03-developersguide.html#multiple-data-models
// is not supported.
type ModelInfos struct {
	// using is the current ModelInfo and corresponds with CQL using declarations. It is nil if no
	// using declaration has been set. Currently only one data model is supported.
	using *Key
	// models are all of the loaded model infos.
	models map[Key]modelInfo
}

// modelInfo holds the parsed data from a ModelInfo XML in an easy to use format. System model info
// is included in every custom modelInfo.
type modelInfo struct {
	// typeMap maps the fully qualified name to a typeInfo. Check the types package
	// ModelInfoName() method for a description of the string key.
	typeMap                      map[string]*TypeInfo
	conversionMap                map[conversionKey]*conversionInfo
	patientBirthDatePropertyName string
	defaultContext               string
	url                          string
	key                          Key
}

// Key is the name and version of a ModeInfo. This is the same name and version as the CQL using
// declaration (ex using FHIR version 4.0.1). Implementation note: this struct needs to be usable as
// a key in a map.
type Key struct {
	// Name represents the name of the model info (ex FHIR).
	Name string
	// Version represents the version of the model info (ex 4.0.1).
	Version string
}

func (k Key) String() string {
	return fmt.Sprintf("%v %v", k.Name, k.Version)
}

type conversionKey struct {
	// Check the types package ModelInfoName() method for a description of fromType and toType.
	fromType string
	toType   string
}

// TypeInfo holds details about a NamedType from a ModelInfo.
type TypeInfo struct {
	// Name is the fully qualified model info type name for this type.
	Name string
	// Properties is a map of property (element) name to the type specifier for it.
	Properties map[string]types.IType
	// BaseType is the fully qualified model info type name for the base (parent) type.
	BaseType string
	// The identifier specifies a unique name for the class that may be independent of the name. In
	// FHIR, this corresponds to the profile identifier.
	Identifier string
	// Retrievable specifies whether the class can be used within a retrieve statement.
	Retrievable bool
	// PrimaryCodePath specifies the path that should be used to perform code filtering when a
	// retrieve does not specify a code path.
	PrimaryCodePath string
}

// ErrTypeNotFound is an error that is returned when a type is not found in the modelinfo.
var ErrTypeNotFound = errors.New("not found in data model")

var errDataModelNotFound = errors.New("data model not found")

var errUsingNotSet = errors.New("using declaration has not been set")

var errPropertyNotFound = errors.New("property not found in data model")

func newErrPropertyNotFound(parentTypeName, property string) error {
	return fmt.Errorf("property %q not found in Parent Type %q %w", property, parentTypeName, errPropertyNotFound)
}

func (m *ModelInfos) typeToModelKey(t types.IType) (modelInfo, Key, error) {
	var key Key
	_, ok := t.(*types.Named)
	if ok {
		if m.using == nil {
			return modelInfo{}, Key{}, fmt.Errorf("cannot use %v %w", t, errUsingNotSet)
		}
		// TODO: b/301606416 - When multiple data models are supported this should map the Named Type
		//   Qualifier --> Using Key.
		key = *m.using
	} else {
		key = Key{Name: "System", Version: "1.0.0"}
	}
	model, ok := m.models[key]
	if !ok {
		return modelInfo{}, Key{}, fmt.Errorf("%v %w", m.using, errDataModelNotFound)
	}
	return model, key, nil
}

// PropertyTypeSpecifier attempts to return the TypeSpecifier for the property on the parent type
// passed in. The passed property must be only one level deep (e.g. property1 instead of
// property1.property2).
// The input type can be a List, in which case the FHIR Path traversal rules for properties will
// be applied to determine the type: https://build.fhir.org/ig/HL7/cql/03-developersguide.html#path-traversal.
// Currently for top level resource types like FHIR.Patient, they need to be built outside of this
// helper based on knowledge of what resources were retrieved.
func (m *ModelInfos) PropertyTypeSpecifier(parentType types.IType, property string) (types.IType, error) {
	model, key, err := m.typeToModelKey(parentType)
	if err != nil {
		return nil, err
	}
	if strings.Contains(property, ".") {
		return nil, fmt.Errorf("internal error - property passed to PropertyTypeSpecifier should not contain \".\" only a single component of the property should be passed at a time")
	}
	if parentType == nil {
		return nil, fmt.Errorf("internal error - cannot compute the property type of a nil parent type")
	}
	if parentType.Equal(types.Any) {
		// The return type must be Any as well.
		return types.Any, nil
	}

	switch parentTypeSpecifier := parentType.(type) {
	case *types.Tuple:
		childTS, ok := parentTypeSpecifier.ElementTypes[property]
		if !ok {
			return nil, fmt.Errorf("%v does not have property %q", parentTypeSpecifier, property)
		}
		return childTS, nil
	case *types.List:
		innerSpecifier, err := m.PropertyTypeSpecifier(parentTypeSpecifier.ElementType, property)
		if err != nil {
			return nil, err
		}
		isSub, err := m.IsSubType(innerSpecifier, &types.List{ElementType: types.Any})
		if err != nil {
			return nil, err
		}
		if isSub {
			// Since the child is a list and the parent is a list, we don't wrap this in two lists to
			// follow the flattening seen in
			// https://build.fhir.org/ig/HL7/cql/03-developersguide.html#path-traversal.
			return innerSpecifier, nil
		}
		return &types.List{ElementType: innerSpecifier}, nil
	case *types.Named, types.System:
		parentTypeName, err := parentTypeSpecifier.ModelInfoName()
		if err != nil {
			return nil, err
		}
		parent, ok := model.typeMap[parentTypeName]
		if !ok {
			return nil, fmt.Errorf("parentNamedType %q not found in data model %v", parentTypeName, key)
		}
		childTS, ok := parent.Properties[property]
		if !ok {
			// Try to check the property on the base type:
			baseTypes, err := m.BaseTypes(parentType)
			if err != nil {
				return nil, err
			}
			for _, b := range baseTypes {
				ts, err := m.PropertyTypeSpecifier(b, property)
				if errors.Is(err, errPropertyNotFound) {
					continue
				} else if err != nil {
					return nil, err
				}
				return ts, err
			}
			// Otherwise, property not found error:
			return nil, newErrPropertyNotFound(parent.Name, property)
		}
		return childTS, nil
	case *types.Interval:
		// Note these are not defined in the system modelinfo, so the logic is spelled out here to
		// keep it all in one place.
		switch property {
		case "low", "high":
			return parentTypeSpecifier.PointType, nil
		case "lowClosed", "highClosed":
			return types.Boolean, nil
		default:
			return nil, fmt.Errorf("invalid property on interval. got: %v, want: low, high, lowClosed, highClosed %w", property, errPropertyNotFound)
		}
	case *types.Choice:
		// Check the property on each choice, and keep track of the unique valid result types.
		validPropertyTypes := []types.IType{}
		for _, ct := range parentTypeSpecifier.ChoiceTypes {
			propType, err := m.PropertyTypeSpecifier(ct, property)
			if errors.Is(err, errPropertyNotFound) {
				continue
			}
			if err != nil {
				return nil, err
			}
			if !containsType(validPropertyTypes, propType) {
				validPropertyTypes = append(validPropertyTypes, propType)
			}
		}
		switch len(validPropertyTypes) {
		case 0:
			return nil, newErrPropertyNotFound(parentTypeSpecifier.String(), property)
		case 1:
			return validPropertyTypes[0], nil
		default:
			return &types.Choice{ChoiceTypes: validPropertyTypes}, nil
		}

	default:
		return nil, fmt.Errorf("internal error (PropertyTypeSpecifier) - parentType %v is not a NamedTypeSpecifier, ListTypeSpecifier, IntervalTypeSpecifier, or System type specifier", parentType)
	}
}

func containsType(types []types.IType, arg types.IType) bool {
	for _, t := range types {
		if t.Equal(arg) {
			return true
		}
	}
	return false
}

// Convertible is the result of the IsImplicitlyConvertible function.
type Convertible struct {
	IsConvertible bool
	// Library and Function name of the function to call to do the conversion
	// ex FHIRHelpers.ToString.
	Library  string
	Function string
}

// IsImplicitlyConvertible uses model info conversionInfo to determine if one type can be converted
// to another. If the `from` type is convertible to the `to` type, this function will return the
// library and function name to call to do the conversion.
func (m *ModelInfos) IsImplicitlyConvertible(from, to types.IType) (Convertible, error) {
	model, _, err := m.typeToModelKey(from)
	if err != nil {
		return Convertible{}, err
	}
	fromType, err := from.ModelInfoName()
	if err != nil {
		return Convertible{}, err
	}
	toType, err := to.ModelInfoName()
	if err != nil {
		return Convertible{}, err
	}
	ci, ok := model.conversionMap[conversionKey{fromType: fromType, toType: toType}]
	if !ok {
		return Convertible{}, nil
	}

	splitNames := strings.Split(ci.FunctionName, ".")
	if len(splitNames) != 2 {
		return Convertible{}, fmt.Errorf("invalid conversion function name %v", ci.FunctionName)
	}
	return Convertible{IsConvertible: true, Library: splitNames[0], Function: splitNames[1]}, nil
}

// BaseTypes returns all of the BaseTypes (aka Parents) of a type excluding Any (or for nested types
// List<Any>).
func (m *ModelInfos) BaseTypes(child types.IType) ([]types.IType, error) {
	model, _, err := m.typeToModelKey(child)
	if err != nil {
		return nil, err
	}

	if child == nil {
		return []types.IType{}, fmt.Errorf("internal error - child type cannot be nil, got: %v", child)
	}

	if child.Equal(types.Any) {
		return []types.IType{}, nil
	}

	switch c := child.(type) {
	case *types.Tuple:
		// It is too expensive to compute all subtypes of a tuple, and the way BaseTypes is used we
		// don't need to. BaseTypes are used for multi step conversions invoked --> sub --> declared,
		// but Tuples will never undergo this type of multistep conversion.
		return []types.IType{}, nil
	case *types.List:
		baseType, err := m.BaseTypes(c.ElementType)
		if err != nil {
			return []types.IType{}, err
		}
		listBaseTypes := []types.IType{}
		for _, b := range baseType {
			listBaseTypes = append(listBaseTypes, &types.List{ElementType: b})
		}
		return listBaseTypes, nil
	case *types.Interval:
		baseType, err := m.BaseTypes(c.PointType)
		if err != nil {
			return []types.IType{}, err
		}
		intervalBaseTypes := []types.IType{}
		for _, b := range baseType {
			intervalBaseTypes = append(intervalBaseTypes, &types.Interval{PointType: b})
		}
		return intervalBaseTypes, nil
	case *types.Choice:
		// TODO(b/301606416): Unclear when this should be true for choice types. Should we check
		// each individual component type? For example is Choice<System.Any> a parent of
		// Choice<System.Integer, System.String>?
		return []types.IType{}, nil
	}

	name, err := child.ModelInfoName()
	if err != nil {
		return []types.IType{}, err
	}
	baseTypes := []types.IType{}
	for depth := 0; ; depth++ {
		tin, ok := model.typeMap[name]
		if !ok {
			return []types.IType{}, fmt.Errorf("%v not found in the data model", child)
		}
		if tin.BaseType == "System.Any" {
			break
		}
		baseTypes = append(baseTypes, typeSpecifierFromElementType(tin.BaseType))
		name = tin.BaseType

		if depth > 100000 {
			return []types.IType{}, fmt.Errorf("internal error - subtype depth exceeded 100000 for %v", child)
		}
	}
	return baseTypes, nil
}

// IsSubType returns true if the child type has the base type (parent) anywhere in it's type
// hierarchy. Returns errors on nil types.IType, since it cannot be determined what the hierarchy
// is.
func (m *ModelInfos) IsSubType(child, base types.IType) (bool, error) {
	model, _, err := m.typeToModelKey(child)
	if err != nil {
		return false, err
	}

	if child == nil || base == nil {
		return false, fmt.Errorf("internal error - child or base type cannot be nil, got: %v, %v", child, base)
	}

	if base.Equal(types.Any) {
		return true, nil
	}

	if child.Equal(types.Any) {
		return false, nil
	}

	switch c := child.(type) {
	case *types.Tuple:
		// Tuples are not defined in modelinfo, so we cannot check modelinfo for subtyping information, we
		// need to calculate it.
		b, isTuple := base.(*types.Tuple)
		if !isTuple {
			return false, nil
		}
		if len(c.ElementTypes) != len(b.ElementTypes) {
			return false, nil
		}

		for childName, childType := range c.ElementTypes {
			baseType, ok := b.ElementTypes[childName]
			if !ok {
				return false, nil
			}
			isSub, err := m.IsSubType(childType, baseType)
			if err != nil {
				return false, err
			}
			if childType.Equal(baseType) || isSub {
				continue
			}
			return false, nil
		}
		return true, nil
	case *types.List:
		baseList, ok := base.(*types.List)
		if !ok {
			return false, nil
		}
		return m.IsSubType(c.ElementType, baseList.ElementType)
	case *types.Interval:
		baseInt, ok := base.(*types.Interval)
		if !ok {
			return false, nil
		}
		return m.IsSubType(c.PointType, baseInt.PointType)
	case *types.Choice:
		// TODO(b/301606416): Unclear when this should be true for choice types. Should we check
		// each individual component type? For example is Choice<System.Any> a parent of
		// Choice<System.Integer, System.String>?
		if base.Equal(types.Any) {
			return true, nil
		}
		return false, nil
	}

	// For all other types (Named, System):
	cName, err := child.ModelInfoName()
	if err != nil {
		return false, err
	}
	tin, ok := model.typeMap[cName]
	if !ok {
		return false, fmt.Errorf("child type %q not found in model info", child.String())
	}

	// TODO(b/301606416): Revisit what happens if a base type is a choice type. For now, return
	// false since the spec is unclear on choice type hierarchies.
	if _, ok := base.(*types.Choice); ok {
		return false, nil
	}

	pName, err := base.ModelInfoName()
	if err != nil {
		return false, err
	}
	if tin.BaseType == pName {
		return true, nil
	}

	anyModelName, err := types.Any.ModelInfoName()
	if err != nil {
		return false, err
	}
	if tin.BaseType == anyModelName {
		return false, nil
	}
	return m.IsSubType(typeSpecifierFromElementType(tin.BaseType), base)
}

// SetUsing corresponds to a CQL using declaration.
func (m *ModelInfos) SetUsing(key Key) error {
	if m.using != nil && key != *m.using {
		return fmt.Errorf("only one data model at a time is currently supported, but got %v and %v", m.using, key)
	}

	if _, ok := m.models[key]; !ok {
		return fmt.Errorf("%v %w", key, errDataModelNotFound)
	}

	m.using = &key
	return nil
}

// ResetUsing resets the using declaration to the system model info key.
func (m *ModelInfos) ResetUsing() {
	m.using = nil
}

// PatientBirthDatePropertyName returns the PatientBirthDatePropertyName field from the custom
// ModelInfo.
func (m *ModelInfos) PatientBirthDatePropertyName() (string, error) {
	if m.using == nil {
		return "", errUsingNotSet
	}
	model, ok := m.models[*m.using]
	if !ok {
		return "", fmt.Errorf("%v %w", m.using, errDataModelNotFound)
	}

	return model.patientBirthDatePropertyName, nil
}

// URL returns the URL field from the custom ModelInfo.
func (m *ModelInfos) URL() (string, error) {
	if m.using == nil {
		return "", errUsingNotSet
	}
	model, ok := m.models[*m.using]
	if !ok {
		return "", fmt.Errorf("%v %w", m.using, errDataModelNotFound)
	}
	return model.url, nil
}

// DefaultContext returns the default context of the custom ModelInfo. To be used if the CQL does
// not specify one. This is actually not set in the FHIR 4.0.1 ModelInfo.
func (m *ModelInfos) DefaultContext() (string, error) {
	if m.using == nil {
		return "", errUsingNotSet
	}
	model, ok := m.models[*m.using]
	if !ok {
		return "", fmt.Errorf("%v %w", m.using, errDataModelNotFound)
	}
	return model.defaultContext, nil
}

// ToNamed converts a string into a NamedType. The string may be qualified (FHIR.Patient) or
// unqualified (Patient) and a Named type with the qualified name will be returned. ToNamed
// validates that the type is in the custom ModelInfo set by the using declaration. System types
// should not be passed to this function and will throw an error.
func (m *ModelInfos) ToNamed(str string) (*types.Named, error) {
	if m.using == nil {
		return nil, errUsingNotSet
	}
	model, ok := m.models[*m.using]
	if !ok {
		return nil, fmt.Errorf("%v %w", m.using, errDataModelNotFound)
	}

	if str == "" {
		return nil, fmt.Errorf("received an empty type, which is invalid")
	}

	// If t is not qualified with key.Name (ex FHIR), then qualify it.
	qualifiedStr := str
	if !strings.HasPrefix(str, m.using.Name+".") {
		qualifiedStr = fmt.Sprintf("%s.%s", m.using.Name, str)
	}

	if _, ok := model.typeMap[qualifiedStr]; !ok {
		return nil, fmt.Errorf("type %v %w %v", str, ErrTypeNotFound, m.using)
	}
	return &types.Named{TypeName: qualifiedStr}, nil
}

// NamedTypeInfo returns the TypeInfo for the given Named type.
func (m *ModelInfos) NamedTypeInfo(t *types.Named) (*TypeInfo, error) {
	model, _, err := m.typeToModelKey(t)
	if err != nil {
		return nil, err
	}

	tInfo, ok := model.typeMap[t.TypeName]
	if !ok {
		return nil, fmt.Errorf("invalid type %v", t)
	}
	return tInfo, nil
}

// New creates a new ModelInfos. The byte array of all the custom ModelInfo should be passed in. System
// ModelInfo is always loaded by default and does not need to be passed in.
func New(modelInfoBytes [][]byte) (*ModelInfos, error) {
	if len(modelInfoBytes) > 1 {
		return nil, fmt.Errorf("only one data model is currently supported")
	}

	// Load system model info by default.
	sysMIBytes, err := embeddata.ModelInfos.ReadFile("third_party/cqframework/system-modelinfo.xml")
	if err != nil {
		return nil, err
	}
	modelInfoBytes = append(modelInfoBytes, sysMIBytes)

	modelInfos := &ModelInfos{
		models: make(map[Key]modelInfo, len(modelInfoBytes)),
		using:  nil,
	}

	for _, miBytes := range modelInfoBytes {
		miXML, err := parse(miBytes)
		if err != nil {
			return nil, err
		}
		mi, err := load(miXML)
		if err != nil {
			return nil, err
		}
		modelInfos.models[mi.key] = *mi
		if mi.key != (Key{Name: "FHIR", Version: "4.0.1"}) && mi.key != (Key{Name: "System", Version: "1.0.0"}) {
			return nil, fmt.Errorf("only FHIR 4.0.1 data model is supported")
		}
	}

	return modelInfos, nil
}

// load parses the XML into a usable data structure.
func load(miXML *modelInfoXML) (*modelInfo, error) {
	mi := &modelInfo{
		patientBirthDatePropertyName: miXML.PatientBirthDatePropertyName,
		url:                          miXML.URL,
		defaultContext:               miXML.DefaultContext,
		key:                          Key{Name: miXML.Name, Version: miXML.Version},
		typeMap:                      make(map[string]*TypeInfo),
		conversionMap:                make(map[conversionKey]*conversionInfo),
	}

	for _, ti := range miXML.TypeInfos {
		err := loadTypeInfo(mi, ti)
		if err != nil {
			return nil, err
		}
	}

	for _, ci := range miXML.ConversionInfos {
		mi.conversionMap[conversionKey{fromType: ci.FromType, toType: ci.ToType}] = ci
	}

	return mi, nil
}

func loadTypeInfo(mi *modelInfo, ti *typeInfoXML) error {
	qualifiedTypeName := ti.Name
	if ti.Namespace != "" {
		qualifiedTypeName = strings.Join([]string{ti.Namespace, ti.Name}, ".")
	}
	tin := &TypeInfo{
		Name:            qualifiedTypeName,
		Properties:      make(map[string]types.IType),
		BaseType:        ti.BaseType,
		Identifier:      ti.Identifier,
		Retrievable:     ti.Retrievable,
		PrimaryCodePath: ti.PrimaryCodePath,
	}

	for _, e := range ti.Elements {
		if e.ElementTypeSpecifier == nil && e.TypeSpecifier == nil {
			// This is a simple ElementType defined elsewhere.
			if e.ElementType != "" {
				tin.Properties[e.Name] = typeSpecifierFromElementType(e.ElementType)
				continue
			}
			// Otherwise use type, which is deprecated but still in use.
			if e.Type != "" {
				tin.Properties[e.Name] = typeSpecifierFromElementType(e.Type)
				continue
			}
			// Unclear what to do, this is an error
			return fmt.Errorf("internal error -- in model info elementTypeSpecifier is nil, and neither elementType or type is set")
		}
		// Otherwise, handle a more specific type specifier:
		specifier := e.TypeSpecifier // deprecated field used as default
		if e.ElementTypeSpecifier != nil {
			specifier = e.ElementTypeSpecifier // override with non-deprecated field if set
		}
		sp, err := buildElementTypeSpecifier(specifier)
		if err != nil {
			return err
		}
		tin.Properties[e.Name] = sp

	}
	if _, ok := mi.typeMap[qualifiedTypeName]; ok {
		// There's a clash with something already in the typeMap. For now, we error.
		return fmt.Errorf("duplicate model info type for type %q", qualifiedTypeName)
	}
	mi.typeMap[qualifiedTypeName] = tin
	return nil
}

func buildElementTypeSpecifier(e *elementTypeSpecifier) (types.IType, error) {
	switch xt := e.XSIType; xt {
	// The ns4 prefixes come from the System model info, for some reason.
	case "NamedTypeSpecifier", "ns4:NamedTypeSpecifier":
		qualifiedTypeName := strings.Join([]string{e.Namespace, e.Name}, ".")
		return &types.Named{TypeName: qualifiedTypeName}, nil
	case "ListTypeSpecifier", "ns4:ListTypeSpecifier":
		if e.ElementTypeSpecifier != nil {
			childTS, err := buildElementTypeSpecifier(e.ElementTypeSpecifier)
			if err != nil {
				return nil, err
			}
			return &types.List{ElementType: childTS}, nil
		}
		return &types.List{ElementType: typeSpecifierFromElementType(e.ElementType)}, nil

	case "ChoiceTypeSpecifier", "ns4:ChoiceTypeSpecifier":
		t := &types.Choice{}
		for _, ct := range e.Choices {
			typeName := ct.Name
			if ct.Namespace != "" {
				typeName = strings.Join([]string{ct.Namespace, ct.Name}, ".")
			}
			t.ChoiceTypes = append(t.ChoiceTypes, typeSpecifierFromElementType(typeName))
		}
		return t, nil
	default:
		return nil, fmt.Errorf("unsupported elementTypeSpecifer in modelInfo. got: %s, want: [ListTypeSpecifier, ChoiceTypeSpecifier]", xt)
	}
}

// typeSpecifierFromElementType returns the type specifier of the passed fully qualified model
// info string type name.
func typeSpecifierFromElementType(modelInfoType string) types.IType {
	if strings.HasPrefix(modelInfoType, "System.") {
		return types.ToSystem(modelInfoType)

	}
	return &types.Named{TypeName: modelInfoType}
}
