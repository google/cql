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

package modelinfo

import (
	"testing"

	"github.com/google/cql/internal/embeddata"
	"github.com/google/go-cmp/cmp"
	"github.com/lithammer/dedent"
)

func TestLoadModelInfo(t *testing.T) {
	tests := []struct {
		desc string
		xml  string
		want *modelInfoXML
	}{
		{
			desc: "header_only",
			xml: dedent.Dedent(`
        <modelInfo xmlns="urn:hl7-org:elm-modelinfo:r1"
				           name="FHIR" version="4.0.1" url="http://hl7.org/fhir"
									 targetQualifier="fhir" patientClassName="FHIR.Patient"
									 patientBirthDatePropertyName="birthDate.value">
        </modelInfo>`),
			want: &modelInfoXML{
				Name:                         "FHIR",
				Version:                      "4.0.1",
				URL:                          "http://hl7.org/fhir",
				TargetQualifier:              "fhir",
				PatientClassName:             "FHIR.Patient",
				PatientBirthDatePropertyName: "birthDate.value",
			},
		},
		{
			desc: "minimal_resource_test",
			xml: dedent.Dedent(`
        <modelInfo xmlns="urn:hl7-org:elm-modelinfo:r1"
									 xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
				           name="FHIR" version="4.0.1" url="http://hl7.org/fhir"
									 targetQualifier="fhir" patientClassName="FHIR.Patient"
									 patientBirthDatePropertyName="birthDate.value">
          <typeInfo baseType="FHIR.DomainResource"
      		    namespace="FHIR" name="FakeCondition"
      			  identifier="http://hl7.org/fhir/StructureDefinition/FakeCondition"
      			  label="FakeCondition" retrievable="true" primaryCodePath="fakeCode"
      			  xsi:type="ClassInfo">
            <element name="clinicalStatus" elementType="FHIR.CodeableConcept"/>
            <element name="fakeCode" elementType="FHIR.CodeableConcept"/>
						<element name="event">
							<elementTypeSpecifier elementType="FHIR.CodeableConcept" xsi:type="ListTypeSpecifier"/>
						</element>
						<element name="link">
							<elementTypeSpecifier xsi:type="ListTypeSpecifier">
									<elementTypeSpecifier namespace="FHIR" name="Bundle.Link" xsi:type="NamedTypeSpecifier"/>
							</elementTypeSpecifier>
						</element>
						<element name="subject">
							<elementTypeSpecifier xsi:type="ChoiceTypeSpecifier">
									<choice namespace="FHIR" name="CodeableConcept" xsi:type="NamedTypeSpecifier"/>
									<choice namespace="FHIR" name="Reference" xsi:type="NamedTypeSpecifier"/>
							</elementTypeSpecifier>
						</element>
      		</typeInfo>
					<typeInfo baseType="FHIR.Element" namespace="FHIR" name="AggregationMode" retrievable="false" xsi:type="ClassInfo">
      			<element name="value" elementType="System.String"/>
   				</typeInfo>
					<conversionInfo functionName="FHIRHelpers.ToCode" fromType="FHIR.Coding" toType="System.Code"/>
					<conversionInfo functionName="FHIRHelpers.ToConcept" fromType="FHIR.CodeableConcept" toType="System.Concept"/>
					<contextInfo name="Patient" keyElement="id" birthDateElement="birthDate.value">
      			<contextType namespace="FHIR" name="Patient"/>
   				</contextInfo>
					<contextInfo name="Encounter" keyElement="id">
							<contextType namespace="FHIR" name="Encounter" modelName="modelName"/>
					</contextInfo>
        </modelInfo>`),
			want: &modelInfoXML{
				Name:                         "FHIR",
				Version:                      "4.0.1",
				URL:                          "http://hl7.org/fhir",
				TargetQualifier:              "fhir",
				PatientClassName:             "FHIR.Patient",
				PatientBirthDatePropertyName: "birthDate.value",
				TypeInfos: []*typeInfoXML{
					{
						BaseType:        "FHIR.DomainResource",
						Name:            "FakeCondition",
						Namespace:       "FHIR",
						Identifier:      "http://hl7.org/fhir/StructureDefinition/FakeCondition",
						PrimaryCodePath: "fakeCode",
						XSIType:         "ClassInfo",
						Label:           "FakeCondition",
						Retrievable:     true,
						Elements: []*element{
							{Name: "clinicalStatus", ElementType: "FHIR.CodeableConcept"},
							{Name: "fakeCode", ElementType: "FHIR.CodeableConcept"},
							{
								Name: "event",
								ElementTypeSpecifier: &elementTypeSpecifier{
									ElementType: "FHIR.CodeableConcept",
									XSIType:     "ListTypeSpecifier"},
							},
							{
								Name: "link",
								ElementTypeSpecifier: &elementTypeSpecifier{
									ElementTypeSpecifier: &elementTypeSpecifier{Namespace: "FHIR", Name: "Bundle.Link", XSIType: "NamedTypeSpecifier"},
									XSIType:              "ListTypeSpecifier"},
							},
							{
								Name: "subject",
								ElementTypeSpecifier: &elementTypeSpecifier{
									Choices: []*choice{
										{
											Namespace: "FHIR",
											Name:      "CodeableConcept",
											XSIType:   "NamedTypeSpecifier",
										},
										{
											Namespace: "FHIR",
											Name:      "Reference",
											XSIType:   "NamedTypeSpecifier",
										},
									},
									XSIType: "ChoiceTypeSpecifier"},
							},
						},
					},
					{
						BaseType:    "FHIR.Element",
						Name:        "AggregationMode",
						Namespace:   "FHIR",
						XSIType:     "ClassInfo",
						Retrievable: false,
						Elements: []*element{
							{Name: "value", ElementType: "System.String"},
						},
					},
				},
				ConversionInfos: []*conversionInfo{
					{
						FunctionName: "FHIRHelpers.ToCode",
						FromType:     "FHIR.Coding",
						ToType:       "System.Code",
					},
					{
						FunctionName: "FHIRHelpers.ToConcept",
						FromType:     "FHIR.CodeableConcept",
						ToType:       "System.Concept",
					},
				},
				ContextInfos: []*contextInfo{
					{
						Name:             "Patient",
						KeyElement:       "id",
						BirthDateElement: "birthDate.value",
						ContextType:      &contextType{Namespace: "FHIR", Name: "Patient"},
					},
					{
						Name:        "Encounter",
						KeyElement:  "id",
						ContextType: &contextType{Namespace: "FHIR", Name: "Encounter", ModelName: "modelName"},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got, err := parse([]byte(test.xml))
			if err != nil {
				t.Fatalf("LoadModel failed: %v", err)
			}
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("LoadModel() diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestLoadModelInfo_SanityCheckEmbeddedModelInfos(t *testing.T) {
	// This is a sanity check test, checking that loading in the embedded model info files
	// work without returning an error.
	cases := []struct {
		name string
		path string
	}{
		{
			name: "FHIR 4.0.1",
			path: "third_party/cqframework/fhir-modelinfo-4.0.1.xml",
		},
		{
			name: "System ModelInfo",
			path: "third_party/cqframework/system-modelinfo.xml",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := embeddata.ModelInfos.ReadFile(tc.path)
			if err != nil {
				t.Fatalf("Reading embedded file %s failed unexpectedly: %v", tc.path, err)
			}
			if _, err := parse(data); err != nil {
				t.Fatalf("loadModel(%s) failed unexpectedly: %v", tc.path, err)
			}
		})
	}
}
