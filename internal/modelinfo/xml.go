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

// Package modelinfo provides utilities for working with CQL ModelInfo XML files.
package modelinfo

import (
	"encoding/xml"
)

// modelInfoXML describes the underlying data model that the CQL is running over. It mainly consists
// of TypeInfo with metadata on each type, ConversionInfo which describes conversion between
// ModelInfo types and CQL system types and ContextInfos which describes the available CQL contexts.
// See the ModelInfo schema for the specification: https://cql.hl7.org/elm/schema/modelinfo.xsd
type modelInfoXML struct {
	// Name that will be referenced by using statements in CQL.
	Name string `xml:"name,attr"`
	// Version that will be referenced by using statements in CQL if specified.
	Version                      string `xml:"version,attr"`
	URL                          string `xml:"url,attr"`
	TargetQualifier              string `xml:"targetQualifier,attr"`
	PatientClassName             string `xml:"patientClassName,attr"`
	PatientBirthDatePropertyName string `xml:"patientBirthDatePropertyName,attr"`
	// The default context to be used if the CQL does not specify one (ex Patient).
	DefaultContext  string            `xml:"defaultContext,attr"`
	TypeInfos       []*typeInfoXML    `xml:"typeInfo"`
	ConversionInfos []*conversionInfo `xml:"conversionInfo"`
	ContextInfos    []*contextInfo    `xml:"contextInfo"`
}

// typeInfoXML describes the types in the data model. It describes which properties can be accessed,
// whether this type is retrievable, the name of the type as specified in CQL and more.
type typeInfoXML struct {
	// BaseType creates a hierarchy of types which can be used in places like casting https://cql.hl7.org/03-developersguide.html#casting.
	// System.Any -> FHIR.Resource -> FHIR.DomainResource
	// System.Any -> FHIR.Element -> FHIR.BackboneElement
	BaseType string `xml:"baseType,attr"`
	// Namespace typically just the name of the model. In FHIR it is always set to "FHIR".
	Namespace string `xml:"namespace,attr"`
	// Name specifies the name of the type within the data model, in FHIR this is the name of the
	// resource or for nested types could be something like Account.Coverage.
	Name string `xml:"name,attr"`
	// The identifier specifies a unique name for the class that may be independent of the name. In
	// FHIR, this corresponds to the profile identifier.
	Identifier string `xml:"identifier,attr"`
	// Label specifies the name of the class as it is referenced from CQL.
	Label string `xml:"label,attr"`
	// Retrievable specifies whether the class can be used within a retrieve statement.
	Retrievable bool `xml:"retrievable,attr"`
	// PrimaryCodePath specifies the path that should be used to perform code filtering when a
	// retrieve does not specify a code path.
	PrimaryCodePath string `xml:"primaryCodePath,attr"`
	// XSIType is always set to ClassInfo.
	XSIType string `xml:"http://www.w3.org/2001/XMLSchema-instance type,attr"`
	// Elements holds the CQL properties that can be accessed on the type.
	Elements []*element `xml:"element"`
}

// element is a property that can be accessed on a particular typeInfo. Either the ElementType or
// ElementTypeSpecifier is set.
type element struct {
	// Name is the name that can be accessed using property in CQL.
	Name string `xml:"name,attr"`
	// ElementType is set if this is a single instance of a type. It seems that all ElementTypes have
	// conversionInfo and can be directly converted to a System type. There is no documentation
	// confirming this, but it is true in FHIR.
	ElementType string `xml:"elementType,attr"`
	// Type is a deprecated field replaced by ElementType in newer model info versions. However, in
	// the System model info this is still used. See more here:
	// https://cql.hl7.org/elm/schema/modelinfo.xsd.
	Type string `xml:"type,attr"`
	// ElementTypeSpecifier is set for Lists or Choice types.
	ElementTypeSpecifier *elementTypeSpecifier `xml:"elementTypeSpecifier"`
	// TypeSpecifier a deprecated field replaced by ElementTypeSpecifier. However, it still exists in
	// some model infos like the System model info. See more here:
	// https://cql.hl7.org/elm/schema/modelinfo.xsd.
	TypeSpecifier *elementTypeSpecifier `xml:"typeSpecifier"`
}

// elementTypeSpecifier is used to specify list and choice types for elements.
//
// If a List then either the ElementType will be set if it is a list of types that can be converted
// to System types or ElementTypeSpecifier will be set and hold a namedTypeSpecifier pointing to
// types with their own TypeInfo.
//
// If a Choice then just Choices will be set.
type elementTypeSpecifier struct {
	// Namespace is only set for NamedTypeSpecifiers.
	Namespace string `xml:"namespace,attr"`
	//  ElementType is used for lists of types that have ConversionInfo. It includes the namespace "FHIR.CodeableConcept".
	ElementType string `xml:"elementType,attr"`
	// ElementTypeSpecifier is rarely set, but in FHIR when it is set it is for a List of NamedTypes.
	ElementTypeSpecifier *elementTypeSpecifier `xml:"elementTypeSpecifier"`
	// XSIType for FHIR is one of ListTypeSpecifier or ChoiceTypeSpecifier. If this is the
	// elementTypeSpecifier of an elementTypeSpecifier aka a list of namedTypes then it will be
	// NamedTypeSpecifier.
	XSIType string `xml:"http://www.w3.org/2001/XMLSchema-instance type,attr"`
	// If the XSIType is ChoiceTypeSpecifier, the choice types are specified here.
	Choices []*choice `xml:"choice"`
	// Name is only set if this is a NamedTypeSpecifier, confusingly it is the type.
	Name string `xml:"name,attr"`
}

type choice struct {
	Namespace string `xml:"namespace,attr"`
	// Name confusingly is the type of this choice.
	Name string `xml:"name,attr"`
	// XSIType for FHIR is always NamedTypeSpecifier.
	XSIType string `xml:"http://www.w3.org/2001/XMLSchema-instance type,attr"`
}

type conversionInfo struct {
	FunctionName string `xml:"functionName,attr"`
	// FromType in FHIR is always a FHIR type.
	FromType string `xml:"fromType,attr"`
	// ToType in FHIR is always a System type.
	ToType string `xml:"toType,attr"`
}

type contextInfo struct {
	// Name that will be referenced by context statements within CQL.
	Name             string       `xml:"name,attr"`
	KeyElement       string       `xml:"keyElement,attr"`
	BirthDateElement string       `xml:"birthDateElement,attr"`
	ContextType      *contextType `xml:"contextType"`
}

type contextType struct {
	Namespace string `xml:"namespace,attr"`
	Name      string `xml:"name,attr"`
	// ModelName is deprecated, but here for compatibility.
	ModelName string `xml:"modelName,attr"`
}

// parse loads a CQL data model from a model info XML byte array.
func parse(bytes []byte) (*modelInfoXML, error) {
	info := &modelInfoXML{}
	if err := xml.Unmarshal(bytes, &info); err != nil {
		return nil, err
	}

	// TODO(b/298104070): Any additional invariant checking for the model info content.
	return info, nil
}
