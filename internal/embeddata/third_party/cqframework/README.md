# CQL grammar

This folder includes the Antlr 4.0 grammar, FHIRHelpers CQL library and ModelInfo files for the CQL language.

## Source
The files are sourced from the [reference implementation repository](https://github.com/cqframework/clinical_quality_language), in the [v1.5.2 tag](https://github.com/cqframework/clinical_quality_language/tree/v1.5.2/Src), which is sourced under [Apache License 2.0](https://github.com/cqframework/clinical_quality_language/blob/v1.5.2/LICENSE).

Note that v1.5.2 tag in this repository does _not_ correspond to the [HL7 CQL
spec v1.5.2](https://cql.hl7.org/history.html). More details in b/325656978, and
we should consider using a later version of the tag.

## Downloading the files
```
# Download files from the v1.5.2 release:
REPOROOT=https://raw.githubusercontent.com/cqframework/clinical_quality_language/v1.5.2
# Download the fhirpath grammar dependency:
curl $REPOROOT/Src/grammar/fhirpath.g4 > fhirpath.g4
# Download the cql grammar and capitalize its name:
rm Cql.g4
echo "// CQL grammar, sourced from $REPOROOT" > Cql.g4
echo '// Capitalized to work with both ANTLR conventions and Go visibility rules.' >> Cql.g4
curl $REPOROOT/Src/grammar/cql.g4 | sed '1 s/^grammar cql;$/grammar Cql;/' >> Cql.g4
# Download modelinfo
curl $REPOROOT/Src/java/quick/src/main/resources/org/hl7/fhir/fhir-modelinfo-4.0.1.xml > fhir-modelinfo-4.0.1.xml
curl $REPOROOT/Src/java/model/src/main/resources/org/hl7/elm/r1/system-modelinfo.xml > system-modelinfo.xml
# Download FHIRHelpers
curl $REPOROOT/Src/java/quick/src/main/resources/org/hl7/fhir/FHIRHelpers-4.0.1.cql > FHIRHelpers-4.0.1.cql
# Download the license file:
curl $REPOROOT/LICENSE > LICENSE
# Download the CVLT grammar dependency:
curl https://https://raw.githubusercontent.com/cqframework/cql-tests-runner/main/cvl/.grammar/cvlt.g4 >> Cvlt.g4
```

## Changes from the source
In order for generated code to work with both ANTLR conventions and Go visibility rules we are capitalizing the cql grammar's name from `cql` to `Cql`.

The system-modelinfo.xml is missing conversion functions. Add the following conversion functions in the system-modelinfo.xml so implicit conversions are supported.

``` xml
<conversionInfo functionName="SYSTEM.ToDecimal" fromType="System.Integer" toType="System.Decimal"/>
<conversionInfo functionName="SYSTEM.ToDecimal" fromType="System.Long" toType="System.Decimal"/>
<conversionInfo functionName="SYSTEM.ToLong" fromType="System.Integer" toType="System.Long"/>
<conversionInfo functionName="SYSTEM.ToDateTime" fromType="System.Date" toType="System.DateTime"/>
<conversionInfo functionName="SYSTEM.ToQuantity" fromType="System.Integer" toType="System.Quantity"/>
<conversionInfo functionName="SYSTEM.ToQuantity" fromType="System.Decimal" toType="System.Quantity"/>
<conversionInfo functionName="SYSTEM.ToConcept" fromType="System.Code" toType="System.Concept"/>
```

Also update ValueSet and Vocabulary to the following, to ensure all properties are supported.
```xml
<ns4:typeInfo xsi:type="ns4:ClassInfo" name="System.Vocabulary" baseType="System.Any">
    <ns4:element name="id" type="System.String"/>
    <ns4:element name="version" type="System.String"/>
    <ns4:element name="name" type="System.String"/>
</ns4:typeInfo>
<ns4:typeInfo xsi:type="ns4:ClassInfo" name="System.ValueSet" baseType="System.Vocabulary">
    <ns4:element name="codesystems">
         <ns4:typeSpecifier xsi:type="ns4:ListTypeSpecifier" elementType="System.CodeSystem"/>
    </ns4:element>
</ns4:typeInfo>
```
