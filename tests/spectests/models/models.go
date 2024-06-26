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

// Code generated by https://github.com/gocomply/xsd2go; DO NOT EDIT.
// Models for http://hl7.org/fhirpath/tests
package models

import (
	"encoding/xml"
)

// Element
type Tests struct {
	XMLName xml.Name `xml:"tests"`

	Name string `xml:"name,attr"`

	Version string `xml:"version,attr"`

	Description string `xml:"description,attr"`

	Reference string `xml:"reference,attr"`

	Notes string `xml:"notes"`

	Group []Group `xml:"group"`
}

// XSD ComplexType declarations

type Group struct {
	XMLName xml.Name

	Name string `xml:"name,attr"`

	Version string `xml:"version,attr"`

	Description string `xml:"description,attr"`

	Reference string `xml:"reference,attr"`

	Notes string `xml:"notes"`

	Test []Test `xml:"test"`
}

type Test struct {
	XMLName xml.Name

	Name string `xml:"name,attr"`

	Version string `xml:"version,attr"`

	Description string `xml:"description,attr"`

	Reference string `xml:"reference,attr"`

	Inputfile string `xml:"inputfile,attr"`

	Predicate bool `xml:"predicate,attr"`

	Mode ModeType `xml:"mode,attr"`

	Ordered bool `xml:"ordered,attr"`

	CheckOrderedFunctions bool `xml:"checkOrderedFunctions,attr"`

	Expression Expression `xml:"expression"`

	Output []Output `xml:"output"`

	Notes string `xml:"notes"`
}

type Expression struct {
	XMLName xml.Name

	Invalid InvalidType `xml:"invalid,attr"`

	Text string `xml:",chardata"`
}

type Output struct {
	XMLName xml.Name

	Type OutputType `xml:"type,attr,omitempty"`

	Text string `xml:",chardata"`
}

// XSD SimpleType declarations

type OutputType string

const OutputTypeBoolean OutputType = "boolean"

const OutputTypeCode OutputType = "code"

const OutputTypeDate OutputType = "date"

const OutputTypeDatetime OutputType = "dateTime"

const OutputTypeDecimal OutputType = "decimal"

const OutputTypeInteger OutputType = "integer"

const OutputTypeLong OutputType = "long"

const OutputTypeQuantity OutputType = "quantity"

const OutputTypeString OutputType = "string"

const OutputTypeTime OutputType = "time"

type InvalidType string

const InvalidTypeFalse InvalidType = "false"

const InvalidTypeSemantic InvalidType = "semantic"

const InvalidTypeTrue InvalidType = "true"

type ModeType string

const ModeTypeStrict ModeType = "strict"

const ModeTypeLoose ModeType = "loose"
