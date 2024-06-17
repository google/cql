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

// Package model provides an ELM-like data structure for an intermediate representation of CQL.
package model

import (
	"github.com/google/cql/types"
	"github.com/kylelemons/godebug/pretty"
)

// TODO(b/302631773): We should make this model 1:1 with ELM, utilize the language and terms
// used by ELM's spec and XSD even if we do not automatically generate go code from the XSD.
// As long as this is hand-rolled though, we should comply with go conventions, some of which
// being violated here such as Functions should not have a "get" prefix, and interface names
// should not be `I` prefixed.

// Library represents a base level CQL library, typically from one CQL file.
type Library struct {
	Identifier  *LibraryIdentifier
	Usings      []*Using
	Includes    []*Include
	Parameters  []*ParameterDef
	CodeSystems []*CodeSystemDef
	Concepts    []*ConceptDef
	Valuesets   []*ValuesetDef
	Codes       []*CodeDef
	Statements  *Statements
}

func (l *Library) String() string {
	return pretty.Sprint(l)
}

// IElement is an interface implemented by all CQL Element structs.
type IElement interface {
	Row() int
	Col() int
	GetResultType() types.IType
}

// Element is the base for all CQL nodes.
type Element struct {
	// TODO(b/298104167): Add common row, column.
	ResultType types.IType
}

// Row returns the element's row in the source file.
func (t *Element) Row() int {
	// TODO(b/298104167): Add row information.
	return 0
}

// Col returns the element's column in the source file.
func (t *Element) Col() int {
	// TODO(b/298104167): Add column information.
	return 0
}

// GetResultType returns the type of the result which may be nil if unknown or not yet implemented.
func (t *Element) GetResultType() types.IType {
	if t == nil {
		return types.Unset
	}
	return t.ResultType
}

// DateTimePrecision represents the precision of a DateTimeValue (or soon, TimeValue).
// This is a string and not an integer value so that the default JSON marshaled value is readable
// and useful.
type DateTimePrecision string

const (
	// UNSETDATETIMEPRECISION represents unknown precision.
	UNSETDATETIMEPRECISION DateTimePrecision = ""
	// YEAR represents year precision.
	YEAR DateTimePrecision = "year"
	// MONTH represents month precision.
	MONTH DateTimePrecision = "month"
	// WEEK represents a week precision. Not valid for Date / DateTime values.
	WEEK DateTimePrecision = "week"
	// DAY represents day precision.
	DAY DateTimePrecision = "day"
	// HOUR represents an hour precision.
	HOUR DateTimePrecision = "hour"
	// MINUTE represents a minute precision.
	MINUTE DateTimePrecision = "minute"
	// SECOND represents second precision.
	SECOND DateTimePrecision = "second"
	// MILLISECOND represents millisecond precision.
	MILLISECOND DateTimePrecision = "millisecond"
)

// Unit represents the unit for a QuantityValue.
type Unit string

// TODO(b/319155752) Add support for UCUM values.
const (
	// UNSETUNIT represents unknown unit.
	UNSETUNIT Unit = ""
	// ONEUNIT defines a base unit.
	// This is often the result of dividing quantities with the same unit, canceling the unit out.
	ONEUNIT Unit = "1"
	// YEARUNIT represents year unit.
	YEARUNIT Unit = "year"
	// MONTHUNIT represents month unit.
	MONTHUNIT Unit = "month"
	// WEEKUNIT represents a week unit.
	WEEKUNIT Unit = "week"
	// DAYUNIT represents day unit.
	DAYUNIT Unit = "day"
	// HOURUNIT represents an hour unit.
	HOURUNIT Unit = "hour"
	// MINUTEUNIT represents a minute unit.
	MINUTEUNIT Unit = "minute"
	// SECONDUNIT represents second unit.
	SECONDUNIT Unit = "second"
	// MILLISECONDUNIT represents millisecond unit.
	MILLISECONDUNIT Unit = "millisecond"
)

// AccessLevel defines the access modifier for a definition (ExpressionDef, ParameterDef,
// ValueSetDef). If the user does not specify an access modifier the default is public. If the
// library is unnamed then the interpreter should treat all definitions as private even if the
// access modifier is public.
type AccessLevel string

const (
	// Public means other CQL libraries can access the definition.
	Public AccessLevel = "PUBLIC"
	// Private means only the local CQL libraries can access the definition.
	Private AccessLevel = "PRIVATE"
)

// ValuesetDef is a named valueset definition that references a value set by ID.
type ValuesetDef struct {
	*Element
	Name        string
	ID          string           // 1..1
	Version     string           // 0..1
	CodeSystems []*CodeSystemRef // 0..*
	AccessLevel AccessLevel
}

// CodeSystemDef is a named definition that references an external code system by ID and version.
type CodeSystemDef struct {
	*Element
	Name        string
	ID          string // 1..1
	Version     string // 0..1
	AccessLevel AccessLevel
}

// ConceptDef is a named definition that represents a terminology concept. It is made up of code(s)
// from one or more CodeSystems. At least one code is required.
type ConceptDef struct {
	*Element
	Name        string
	Codes       []*CodeRef // 1..*
	Display     string     // 0..1
	AccessLevel AccessLevel
}

// CodeDef is a named definition that references an external code from a CodeSystem by ID.
type CodeDef struct {
	*Element
	Name        string
	Code        string         // 1..1
	CodeSystem  *CodeSystemRef // 0..1
	Display     string         // 0..1
	AccessLevel AccessLevel
}

// ParameterDef is a top-level statement that defines a named CQL parameter.
type ParameterDef struct {
	*Element
	Name        string
	Default     IExpression
	AccessLevel AccessLevel
}

// LibraryIdentifier for the library definition. This matches up with the ELM VersionedIdentifier
// (https://cql.hl7.org/04-logicalspecification.html#versionedidentifier). If nil then this is an
// unnamed library.
type LibraryIdentifier struct {
	*Element
	Local     string
	Qualified string // The full identifier of the library.
	Version   string
}

// Using defines a Using directive in CQL.
type Using struct {
	*Element
	LocalIdentifier string
	// URI is the URL specified at the top of the modelinfo, for FHIR "http://hl7.org/fhir".
	URI     string
	Version string
}

// Include defines an Include library statement in CQL.
type Include struct {
	*Element
	Identifier *LibraryIdentifier
}

// Statements is a collection of expression and function definitions, similar to the ELM structure.
type Statements struct {
	Defs []IExpressionDef
}

// IExpressionDef is implemented by both ExpressionDef and FunctionDef.
type IExpressionDef interface {
	IElement
	GetName() string
	GetContext() string
	GetExpression() IExpression
	GetAccessLevel() AccessLevel
}

// ExpressionDef is a top-level named definition of a CQL expression.
type ExpressionDef struct {
	*Element
	Name        string
	Context     string
	Expression  IExpression
	AccessLevel AccessLevel
}

// GetName returns the name of the definition.
func (e *ExpressionDef) GetName() string { return e.Name }

// GetContext returns the context of the definition.
func (e *ExpressionDef) GetContext() string { return e.Context }

// GetExpression returns the expression of the definition.
func (e *ExpressionDef) GetExpression() IExpression { return e.Expression }

// GetAccessLevel returns the access level of the definition.
func (e *ExpressionDef) GetAccessLevel() AccessLevel { return e.AccessLevel }

// FunctionDef represents a user defined function. All CQL built-in functions have their own struct
// defined below.
type FunctionDef struct {
	// The body of the function is represented by the Expression field in the ExpressionDef. The
	// return type is the ResultType set in the Element.
	*ExpressionDef
	Operands []OperandDef
	Fluent   bool
	// External functions do not have a function body.
	External bool
}

// OperandDef defines an operand for a user defined function.
type OperandDef struct {
	// The type of the operand is the ResultType set in the Element.
	*Expression
	Name string
}

// All items below are for CQL expressions. All CQL expressions embed the base Expression struct
// and implement the IExpression interface. This allows for dynamic trees and sequences of
// expression types, matching CQL's structure.

// IExpression is an interface implemented by all CQL Expression structs
type IExpression interface {
	IElement
	isExpression()
}

// Expression is a base type containing common metadata for all CQL expression types.
type Expression struct {
	*Element
}

func (e *Expression) isExpression() {}

// GetResultType returns the type of the result which may be nil if unknown or not yet implemented.
func (e *Expression) GetResultType() types.IType {
	if e == nil {
		return types.Unset
	}
	return e.Element.GetResultType()
}

// Literal represents a CQL literal.
type Literal struct {
	*Expression
	Value string
}

// An Interval expression.
type Interval struct {
	*Expression
	Low  IExpression
	High IExpression

	// Either LowClosedExpression or LowInclusive should be set.
	LowClosedExpression IExpression
	LowInclusive        bool

	// Either HighClosedExpression or HighInclusive should be set.
	HighClosedExpression IExpression
	HighInclusive        bool
}

// Quantity is an expression representation of a clinical quantity.
// https://cql.hl7.org/04-logicalspecification.html#quantity
type Quantity struct {
	*Expression
	Value float64
	Unit  Unit
}

// A Ratio is an expression that expresses a ratio between two Quantities.
// https://cql.hl7.org/04-logicalspecification.html#ratio
type Ratio struct {
	*Expression
	Numerator   Quantity
	Denominator Quantity
}

// A List expression.
type List struct {
	*Expression
	List []IExpression
}

// Code is a literal code selector.
type Code struct {
	*Expression
	System  *CodeSystemRef
	Code    string
	Display string
}

// Tuple represents a tuple (aka Structured Value), see
// https://cql.hl7.org/04-logicalspecification.html#tuple
type Tuple struct {
	*Expression
	Elements []*TupleElement
}

// TupleElement is an element in a CQL Tuple.
type TupleElement struct {
	Name  string
	Value IExpression
}

// Instance represents an instance of a Class (aka Named Structured Value), see
// https://cql.hl7.org/04-logicalspecification.html#instance
type Instance struct {
	*Expression
	ClassType types.IType
	Elements  []*InstanceElement
}

// InstanceElement is an element in a CQL structure Instance.
type InstanceElement struct {
	Name  string
	Value IExpression
}

// A MessageSeverity determines the type of the message and how it will be processed.
type MessageSeverity string

const (
	// UNSETMESSAGESEVERITY denotes a message severity that shouldn't be allowed.
	UNSETMESSAGESEVERITY MessageSeverity = ""
	// TRACE denotes a message that should be printed with trace information.
	TRACE MessageSeverity = "Trace"
	// MESSAGE denotes a simple message that should be printed.
	MESSAGE MessageSeverity = "Message"
	// WARNING denotes a message that should log a warning to users.
	WARNING MessageSeverity = "Warning"
	// ERROR denotes an error message that should also halt execution.
	ERROR MessageSeverity = "Error"
)

// Message is a CQL expression that represents a message, which is the equivalent of print in most
// other languages.
// https://cql.hl7.org/04-logicalspecification.html#message
type Message struct {
	*Expression
	Source    IExpression
	Condition IExpression
	Code      IExpression
	Severity  IExpression
	Message   IExpression
}

// A SortDirection determines what ordering to use for a query if sorting is enabled.
type SortDirection string

const (
	// UNSETSORTDIRECTION denotes a sort direction that shouldn't be allowed.
	UNSETSORTDIRECTION SortDirection = ""
	// ASCENDING denotes query sorting from smallest to largest values.
	ASCENDING SortDirection = "ASCENDING"
	// DESCENDING denotes query sorting from largest to smallest values.
	DESCENDING SortDirection = "DESCENDING"
)

// A Query expression.
type Query struct {
	*Expression
	Source       []*AliasedSource
	Let          []*LetClause
	Relationship []IRelationshipClause
	Where        IExpression
	Sort         *SortClause
	Aggregate    *AggregateClause // Only aggregate or Return can be populated, not both.
	Return       *ReturnClause
}

// LetClause is https://cql.hl7.org/04-logicalspecification.html#letclause.
type LetClause struct {
	*Element
	Expression IExpression
	Identifier string
}

// IRelationshipClause is an interface that all With and Without meet.
type IRelationshipClause interface {
	IElement
	isRelationshipClause()
}

// RelationshipClause for a Query expression.
type RelationshipClause struct {
	*Element
	// Expression is the source of the inclusion clause.
	Expression IExpression
	Alias      string
	SuchThat   IExpression
}

func (c *RelationshipClause) isRelationshipClause() {}

// With is https://cql.hl7.org/04-logicalspecification.html#with.
type With struct{ *RelationshipClause }

// Without is https://cql.hl7.org/04-logicalspecification.html#without.
type Without struct{ *RelationshipClause }

// SortClause for a Query expression.
type SortClause struct {
	*Element
	ByItems []ISortByItem
}

// AggregateClause for a Query expression.
type AggregateClause struct {
	*Element
	Expression IExpression
	// Starting is the starting value of the aggregate variable. It is always set. If the user does
	// not set it the parser will insert a null literal.
	Starting IExpression
	// Identifier is the alias for the aggregate variable.
	Identifier string
	Distinct   bool
}

// ReturnClause for a Query expression.
type ReturnClause struct {
	*Element
	Expression IExpression
	Distinct   bool
}

// ISortByItem defines one or more items that a query can be sorted by.
// Follows format outlined in https://cql.hl7.org/elm/schema/expression.xsd.
type ISortByItem interface {
	IElement
	isSortByItem()
}

// SortByItem is the base abstract type for all query types.
type SortByItem struct {
	*Element
	Direction SortDirection
}

// SortByDirection enables sorting non-tuple values by direction
type SortByDirection struct {
	*SortByItem
}

func (c *SortByDirection) isSortByItem() {}

// SortByColumn enables sorting by a given column and direction.
type SortByColumn struct {
	*SortByItem
	Path string
}

func (c *SortByColumn) isSortByItem() {}

// AliasedSource is a query source with an alias.
type AliasedSource struct {
	*Expression
	Alias  string
	Source IExpression
}

// Property gets a property from an expression. In ELM if the expression is an AliasRef then Scope
// is set instead of Source. In our model Source is set to AliasRef; there is no Scope.
type Property struct {
	*Expression
	Source IExpression
	Path   string
}

// A Retrieve expression.
type Retrieve struct {
	*Expression
	// TODO(b/312172420): Changing DataType to a named type would make life much easier.
	DataType     string
	TemplateID   string
	CodeProperty string
	// Codes is an expression that returns a list of code values.
	Codes IExpression
}

// Case is a conditional case expression https://cql.hl7.org/04-logicalspecification.html#case.
type Case struct {
	*Expression
	// If comparand is provided it is compared against each When in the CaseItems. The CaseItems are
	// expected to be of the same type or implicitly convertible to the same type as the Comparand. If
	// the comparand is not provided then each When must have resultType boolean.
	Comparand IExpression
	CaseItem  []*CaseItem
	// Else must always be provided.
	Else IExpression
}

// CaseItem is a single case item in a Case expression.
type CaseItem struct {
	*Element
	When IExpression
	Then IExpression
}

// IfThenElse Elm expression from https://cql.hl7.org/04-logicalspecification.html#if
type IfThenElse struct {
	*Expression
	Condition IExpression
	Then      IExpression
	Else      IExpression
}

// MaxValue ELM expression from https://cql.hl7.org/04-logicalspecification.html#maxvalue
type MaxValue struct {
	*Expression
	ValueType types.IType
}

// MinValue ELM expression from https://cql.hl7.org/04-logicalspecification.html#minvalue
type MinValue struct {
	*Expression
	ValueType types.IType
}

// IUnaryExpression is an interface that all Unary expressions meet.
type IUnaryExpression interface {
	IExpression
	GetName() string
	GetOperand() IExpression
	SetOperand(IExpression)
	// To differentiate IUnaryExpression from other interfaces like IBinaryExpression.
	isUnaryExpression()
}

// UnaryExpression is a CQL expression that has one operand. The ELM representation may have
// additional operands.
type UnaryExpression struct {
	*Expression
	Operand IExpression
}

// GetOperand returns the unary expression's operand.
func (a *UnaryExpression) GetOperand() IExpression { return a.Operand }

// SetOperand sets the unary expression's operand.
func (a *UnaryExpression) SetOperand(operand IExpression) { a.Operand = operand }

func (a *UnaryExpression) isUnaryExpression() {}

// TODO(b/297089208): eventually consider moving all UnaryExpressions into their own file for
// organization.

// As is https://cql.hl7.org/09-b-cqlreference.html#as.
type As struct {
	*UnaryExpression
	AsTypeSpecifier types.IType
	Strict          bool
}

var _ IUnaryExpression = &As{}

// Is is https://cql.hl7.org/04-logicalspecification.html#is.
type Is struct {
	*UnaryExpression
	IsTypeSpecifier types.IType
}

var _ IUnaryExpression = &Is{}

// Negate is https://cql.hl7.org/04-logicalspecification.html#negate.
type Negate struct{ *UnaryExpression }

var _ IUnaryExpression = &Negate{}

// Truncate ELM expression https://cql.hl7.org/04-logicalspecification.html#truncate
type Truncate struct{ *UnaryExpression }

var _ IUnaryExpression = &Truncate{}

// Exists is https://cql.hl7.org/04-logicalspecification.html#exists.
type Exists struct{ *UnaryExpression }

var _ IUnaryExpression = &Exists{}

// Not is https://cql.hl7.org/04-logicalspecification.html#not.
type Not struct{ *UnaryExpression }

var _ IUnaryExpression = &Not{}

// First ELM expression from https://cql.hl7.org/04-logicalspecification.html#first
type First struct {
	*UnaryExpression
	// TODO(b/301606416): Support the orderBy parameter.
}

// Last ELM expression from https://cql.hl7.org/04-logicalspecification.html#last.
type Last struct {
	*UnaryExpression
	// TODO(b/301606416): Support the orderBy parameter.
}

var _ IUnaryExpression = &Last{}

// SingletonFrom is https://cql.hl7.org/04-logicalspecification.html#singletonfrom.
type SingletonFrom struct{ *UnaryExpression }

var _ IUnaryExpression = &SingletonFrom{}

// Start is https://cql.hl7.org/04-logicalspecification.html#start.
type Start struct{ *UnaryExpression }

var _ IUnaryExpression = &Start{}

// End is https://cql.hl7.org/04-logicalspecification.html#end.
type End struct{ *UnaryExpression }

var _ IUnaryExpression = &End{}

// Predecessor ELM expression from https://cql.hl7.org/04-logicalspecification.html#predecessor.
type Predecessor struct{ *UnaryExpression }

var _ IUnaryExpression = &Predecessor{}

// Successor ELM expression from https://cql.hl7.org/04-logicalspecification.html#successor.
type Successor struct{ *UnaryExpression }

var _ IUnaryExpression = &Successor{}

// IsNull is https://cql.hl7.org/04-logicalspecification.html#isnull.
type IsNull struct{ *UnaryExpression }

var _ IUnaryExpression = &IsNull{}

// IsFalse is https://cql.hl7.org/04-logicalspecification.html#isfalse.
type IsFalse struct{ *UnaryExpression }

var _ IUnaryExpression = &IsFalse{}

// IsTrue is https://cql.hl7.org/04-logicalspecification.html#istrue.
type IsTrue struct{ *UnaryExpression }

var _ IUnaryExpression = &IsTrue{}

// ToBoolean ELM expression from https://cql.hl7.org/09-b-cqlreference.html#toboolean.
type ToBoolean struct{ *UnaryExpression }

var _ IUnaryExpression = &ToBoolean{}

// ToDateTime ELM expression from https://cql.hl7.org/04-logicalspecification.html#todatetime
type ToDateTime struct{ *UnaryExpression }

var _ IUnaryExpression = &ToDateTime{}

// ToDate ELM expression from https://cql.hl7.org/04-logicalspecification.html#todate.
type ToDate struct{ *UnaryExpression }

var _ IUnaryExpression = &ToDate{}

// ToDecimal ELM expression from https://cql.hl7.org/04-logicalspecification.html#todecimal.
type ToDecimal struct{ *UnaryExpression }

var _ IUnaryExpression = &ToDecimal{}

// ToLong ELM expression from https://cql.hl7.org/04-logicalspecification.html#tolong.
type ToLong struct{ *UnaryExpression }

var _ IUnaryExpression = &ToLong{}

// ToInteger ELM expression from https://cql.hl7.org/09-b-cqlreference.html#tointeger.
type ToInteger struct{ *UnaryExpression }

var _ IUnaryExpression = &ToInteger{}

// ToQuantity ELM expression from https://cql.hl7.org/04-logicalspecification.html#toquantity.
type ToQuantity struct{ *UnaryExpression }

var _ IUnaryExpression = &ToQuantity{}

// ToConcept ELM expression from https://cql.hl7.org/09-b-cqlreference.html#toconcept.
type ToConcept struct{ *UnaryExpression }

var _ IUnaryExpression = &ToConcept{}

// ToString ELM expression from https://cql.hl7.org/09-b-cqlreference.html#tostring.
type ToString struct{ *UnaryExpression }

var _ IUnaryExpression = &ToString{}

// ToTime ELM expression from https://cql.hl7.org/09-b-cqlreference.html#totime.
type ToTime struct{ *UnaryExpression }

var _ IUnaryExpression = &ToTime{}

// AllTrue ELM expression from https://cql.hl7.org/04-logicalspecification.html#alltrue.
// TODO: b/347346351 - In ELM it's modeled as an AggregateExpression, but for now we model it as an
// UnaryExpression since there is no way to set the AggregateExpression's "path" property for CQL as
// far as we can tell.
type AllTrue struct{ *UnaryExpression }

var _ IUnaryExpression = &AllTrue{}

// Count ELM expression from https://cql.hl7.org/09-b-cqlreference.html#count.
// TODO: b/347346351 - In ELM it's modeled as an AggregateExpression, but for now we model it as an
// UnaryExpression since there is no way to set the AggregateExpression's "path" property for CQL as
// far as we can tell.
type Count struct{ *UnaryExpression }

var _ IUnaryExpression = &Count{}

// CalculateAge CQL expression type
type CalculateAge struct {
	*UnaryExpression
	Precision DateTimePrecision
}

// BinaryExpression is a CQL expression that has two operands. The ELM representation may have
// additional operands (ex BinaryExpressionWithPrecision).
type BinaryExpression struct {
	*Expression
	Operands []IExpression
}

// Left returns the Left expression (first operand) of the BinaryExpression. If not present,
// returns nil.
func (b *BinaryExpression) Left() IExpression {
	if len(b.Operands) < 1 {
		return nil
	}
	return b.Operands[0]
}

// Right returns the Right expression (second operand) of the BinaryExpression. If not present,
// returns nil.
func (b *BinaryExpression) Right() IExpression {
	if len(b.Operands) < 2 {
		return nil
	}
	return b.Operands[1]
}

// SetOperands sets the BinaryExpression's operands.
func (b *BinaryExpression) SetOperands(left, right IExpression) {
	b.Operands = []IExpression{left, right}
}

func (b *BinaryExpression) isBinaryExpression() {}

// IBinaryExpression is an interface that all Binary Expressions meet.
// Get prefixes are used below so as to not conflict with struct property names.
type IBinaryExpression interface {
	IExpression
	// GetName returns the name of the BinaryExpression.
	GetName() string
	// Left returns the Left expression (first operand) of the BinaryExpression. If not present,
	// returns nil.
	Left() IExpression
	// Right returns the Right expression (second operand) of the BinaryExpression. If not present,
	// returns nil.
	Right() IExpression
	SetOperands(left, right IExpression)
	// To differentiate IBinaryExpression from other interfaces like INaryExpression.
	isBinaryExpression()
}

// CanConvertQuantity ELM Expression from https://cql.hl7.org/04-logicalspecification.html#canconvertquantity.
type CanConvertQuantity struct{ *BinaryExpression }

var _ IBinaryExpression = &CanConvertQuantity{}

// Equal ELM Expression from https://cql.hl7.org/04-logicalspecification.html#equal.
type Equal struct{ *BinaryExpression }

var _ IBinaryExpression = &Equal{}

// Equivalent ELM Expression from https://cql.hl7.org/04-logicalspecification.html#equivalent.
type Equivalent struct{ *BinaryExpression }

var _ IBinaryExpression = &Equivalent{}

// Less ELM Expression https://cql.hl7.org/04-logicalspecification.html#less
type Less struct{ *BinaryExpression }

var _ IBinaryExpression = &Less{}

// Greater ELM Expression https://cql.hl7.org/04-logicalspecification.html#greater
type Greater struct{ *BinaryExpression }

var _ IBinaryExpression = &Greater{}

// LessOrEqual ELM Expression https://cql.hl7.org/04-logicalspecification.html#lessorequal
type LessOrEqual struct{ *BinaryExpression }

var _ IBinaryExpression = &LessOrEqual{}

// GreaterOrEqual ELM Expression https://cql.hl7.org/04-logicalspecification.html#greaterorequal
type GreaterOrEqual struct{ *BinaryExpression }

var _ IBinaryExpression = &GreaterOrEqual{}

// And is https://cql.hl7.org/04-logicalspecification.html#and.
type And struct{ *BinaryExpression }

// Or is https://cql.hl7.org/04-logicalspecification.html#or
type Or struct{ *BinaryExpression }

// XOr is https://cql.hl7.org/04-logicalspecification.html#xor
type XOr struct{ *BinaryExpression }

// Implies is https://cql.hl7.org/04-logicalspecification.html#implies
type Implies struct{ *BinaryExpression }

// Add ELM Expression https://cql.hl7.org/04-logicalspecification.html#add
type Add struct{ *BinaryExpression }

// Subtract ELM Expression https://cql.hl7.org/04-logicalspecification.html#subtract
type Subtract struct{ *BinaryExpression }

// Multiply ELM Expression https://cql.hl7.org/04-logicalspecification.html#multiply
type Multiply struct{ *BinaryExpression }

// Divide ELM Expression https://cql.hl7.org/04-logicalspecification.html#divide
type Divide struct{ *BinaryExpression }

// Modulo ELM Expression https://cql.hl7.org/04-logicalspecification.html#modulo
type Modulo struct{ *BinaryExpression }

// TruncatedDivide ELM Expression https://cql.hl7.org/04-logicalspecification.html#truncateddivide
type TruncatedDivide struct{ *BinaryExpression }

// Except ELM Expression https://cql.hl7.org/04-logicalspecification.html#except
// Except is a nary expression but we are only supporting two operands.
type Except struct{ *BinaryExpression }

// Intersect ELM Expression https://cql.hl7.org/04-logicalspecification.html#intersect
// Intersect is a nary expression but we are only supporting two operands.
type Intersect struct{ *BinaryExpression }

// Union ELM Expression https://cql.hl7.org/04-logicalspecification.html#union
// Union is a nary expression but we are only supporting two operands.
type Union struct{ *BinaryExpression }

// BinaryExpressionWithPrecision represents a BinaryExpression with a precision property.
type BinaryExpressionWithPrecision struct {
	*BinaryExpression
	// Precision returns the precision of this BinaryExpression. It must be one of the following:
	// https://cql.hl7.org/19-l-cqlsyntaxdiagrams.html#dateTimePrecision.
	Precision DateTimePrecision
}

// Before ELM expression from https://cql.hl7.org/04-logicalspecification.html#before.
type Before BinaryExpressionWithPrecision

var _ IBinaryExpression = &Before{}

// After ELM expression from https://cql.hl7.org/04-logicalspecification.html#after.
type After BinaryExpressionWithPrecision

// SameOrBefore ELM expression from https://cql.hl7.org/04-logicalspecification.html#sameorbefore.
type SameOrBefore BinaryExpressionWithPrecision

// SameOrAfter ELM expression from https://cql.hl7.org/04-logicalspecification.html#sameorafter.
type SameOrAfter BinaryExpressionWithPrecision

// DifferenceBetween ELM expression from https://cql.hl7.org/04-logicalspecification.html#differencebetween.
type DifferenceBetween BinaryExpressionWithPrecision

// In ELM expression from https://cql.hl7.org/04-logicalspecification.html#in.
type In BinaryExpressionWithPrecision

// IncludedIn ELM expression from https://cql.hl7.org/04-logicalspecification.html#included-in.
type IncludedIn BinaryExpressionWithPrecision

// InCodeSystem is https://cql.hl7.org/09-b-cqlreference.html#in-codesystem.
// This is not technically 1:1 with the ELM definition. The ELM defines Code, CodeSystem and
// CodeSystemExpression arguments, the last being seemingly impossible to set for for now we're
// treating this as a binary expression.
type InCodeSystem struct{ *BinaryExpression }

// InValueSet is https://cql.hl7.org/09-b-cqlreference.html#in-valueset.
// This is not technically 1:1 with the ELM definition. The ELM defines Code, ValueSet and
// ValueSetExpression arguments, the last being seemingly impossible to set for for now we're
// treating this as a binary expression.
type InValueSet struct{ *BinaryExpression }

// Contains ELM expression from https://cql.hl7.org/04-logicalspecification.html#contains.
type Contains BinaryExpressionWithPrecision

// CalculateAgeAt ELM expression from https://cql.hl7.org/04-logicalspecification.html#calculateageat.
type CalculateAgeAt BinaryExpressionWithPrecision

// INaryExpression is an interface that Expressions with any number of operands meet.
type INaryExpression interface {
	IExpression
	GetName() string
	GetOperands() []IExpression
	SetOperands([]IExpression)
	// To differentiate INaryExpression from other interfaces like IBinaryExpression.
	isNaryExpression()
}

// NaryExpression that takes any number of operands including zero. The ELM representation may have
// additional operands.
type NaryExpression struct {
	*Expression
	Operands []IExpression
}

// GetOperands returns the operands of the NaryExpression.
func (n *NaryExpression) GetOperands() []IExpression {
	return n.Operands
}

// SetOperands sets the NaryExpression's operands.
func (n *NaryExpression) SetOperands(ops []IExpression) {
	n.Operands = ops
}

func (n *NaryExpression) isNaryExpression() {}

// Coalesce is https://cql.hl7.org/04-logicalspecification.html#coalesce.
type Coalesce struct{ *NaryExpression }

// Concatenate is https://cql.hl7.org/04-logicalspecification.html#concatenate.
type Concatenate struct{ *NaryExpression }

// Date is the functional syntax to create a Date https://cql.hl7.org/09-b-cqlreference.html#date-1.
type Date struct{ *NaryExpression }

// DateTime is the functional syntax to create a to create a CQL DateTime
// https://cql.hl7.org/09-b-cqlreference.html#datetime-1.
type DateTime struct{ *NaryExpression }

// Now is https://cql.hl7.org/04-logicalspecification.html#now.
// Note: in the future we may implement the OperatorExpression, and should convert this to
// one of those at that point.
type Now struct{ *NaryExpression }

// TimeOfDay is https://cql.hl7.org/04-logicalspecification.html#timeofday
// Note: in the future we may implement the OperatorExpression, and should convert this to
// one of those at that point.
type TimeOfDay struct{ *NaryExpression }

// Time is the functional syntax to create a CQL Time
// https://cql.hl7.org/09-b-cqlreference.html#time-1.
type Time struct{ *NaryExpression }

// Today is https://cql.hl7.org/04-logicalspecification.html#today.
// Note: in the future we may implement the OperatorExpression, and should convert this to
// one of thse at that point.
type Today struct{ *NaryExpression }

// ParameterRef defines a reference to a ParameterDef definition used in CQL expressions.
type ParameterRef struct {
	*Expression
	Name string
	// LibraryName is empty for parameters defined in the local CQL library. Otherwise it is the
	// local identifier of the included library.
	LibraryName string
}

// ValuesetRef defines a reference to a ValuesetDef definition used in CQL expressions.
type ValuesetRef struct {
	*Expression
	Name string
	// LibraryName is empty for valuesets defined in the local CQL library. Otherwise it is the
	// local identifier of the included library.
	LibraryName string
}

// CodeSystemRef defines a reference to a CodeSystemDef definition used in CQL expressions.
type CodeSystemRef struct {
	*Expression
	Name string
	// LibraryName is empty for CodeSystems defined in the local CQL library. Otherwise it is the
	// local identifier of the included library.
	LibraryName string
}

// ConceptRef defines a reference to a ConceptDef definition used in CQL expressions.
type ConceptRef struct {
	*Expression
	Name string
	// LibraryName is empty for Concepts defined in the local CQL library. Otherwise it is the
	// local identifier of the included library.
	LibraryName string
}

// CodeRef defines a reference to a Code definition used in CQL expressions.
type CodeRef struct {
	*Expression
	Name string
	// LibraryName is empty for Codes defined in the local CQL library. Otherwise it is the
	// local identifier of the included library.
	LibraryName string
}

// ExpressionRef defines a reference to a ExpressionDef definition used in CQL expressions.
type ExpressionRef struct {
	*Expression
	Name string
	// LibraryName is empty for expressions defined in the local CQL library. Otherwise it is the
	// local identifier of the included library.
	LibraryName string
}

// AliasRef defines a reference to a source within the scope of a query.
type AliasRef struct {
	*Expression
	Name string
}

// QueryLetRef is similar to an AliasRef except specific to references to let clauses.
type QueryLetRef struct {
	*Expression
	Name string
}

// FunctionRef defines a reference to a user defined function.
type FunctionRef struct {
	*Expression
	Name string
	// LibraryName is empty for expressions defined in the local CQL library. Otherwise it is the
	// local identifier of the included library.
	LibraryName string
	Operands    []IExpression
}

// OperandRef defines a reference to an operand within a function.
type OperandRef struct {
	*Expression
	Name string
}

// UNARY EXPRESSION GETNAME()

// GetName returns the name of the system operator.
func (a *As) GetName() string { return "As" }

// GetName returns the name of the system operator.
func (i *Is) GetName() string { return "Is" }

// GetName returns the name of the system operator.
func (e *Exists) GetName() string { return "Exists" }

// GetName returns the name of the system operator.
func (n *Not) GetName() string { return "Not" }

// GetName returns the name of the system operator.
func (a *Truncate) GetName() string { return "Truncate" }

// GetName returns the name of the system operator.
func (f *First) GetName() string { return "First" }

// GetName returns the name of the system operator.
func (l *Last) GetName() string { return "Last" }

// GetName returns the name of the system operator.
func (s *SingletonFrom) GetName() string { return "As" }

// GetName returns the name of the system operator.
func (a *Start) GetName() string { return "Start" }

// GetName returns the name of the system operator.
func (a *End) GetName() string { return "End" }

// GetName returns the name of the system operator.
func (a *Predecessor) GetName() string { return "Predecessor" }

// GetName returns the name of the system operator.
func (a *Successor) GetName() string { return "Successor" }

// GetName returns the name of the system operator.
func (a *IsNull) GetName() string { return "IsNull" }

// GetName returns the name of the system operator.
func (a *IsFalse) GetName() string { return "IsFalse" }

// GetName returns the name of the system operator.
func (a *IsTrue) GetName() string { return "IsTrue" }

// GetName returns the name of the system operator.
func (a *ToBoolean) GetName() string { return "ToBoolean" }

// GetName returns the name of the system operator.
func (a *ToDateTime) GetName() string { return "ToDateTime" }

// GetName returns the name of the system operator.
func (a *ToDate) GetName() string { return "ToDate" }

// GetName returns the name of the system operator.
func (a *ToDecimal) GetName() string { return "ToDecimal" }

// GetName returns the name of the system operator.
func (a *ToLong) GetName() string { return "ToLong" }

// GetName returns the name of the system operator.
func (a *ToInteger) GetName() string { return "ToInteger" }

// GetName returns the name of the system operator.
func (a *ToQuantity) GetName() string { return "ToQuantity" }

// GetName returns the name of the system operator.
func (a *ToConcept) GetName() string { return "ToConcept" }

// GetName returns the name of the system operator.
func (a *ToString) GetName() string { return "ToString" }

// GetName returns the name of the system operator.
func (a *ToTime) GetName() string { return "ToTime" }

// GetName returns the name of the system operator.
func (a *CalculateAge) GetName() string { return "CalculateAge" }

// GetName returns the name of the system operator.
func (a *Negate) GetName() string { return "Negate" }

// BINARY EXPRESSION GETNAME()

// GetName returns the name of the system operator.
func (a *CanConvertQuantity) GetName() string { return "CanConvertQuantity" }

// GetName returns the name of the system operator.
func (a *Equal) GetName() string { return "Equal" }

// GetName returns the name of the system operator.
func (a *Equivalent) GetName() string { return "Equivalent" }

// GetName returns the name of the system operator.
func (a *Less) GetName() string { return "Less" }

// GetName returns the name of the system operator.
func (a *Greater) GetName() string { return "Greater" }

// GetName returns the name of the system operator.
func (a *LessOrEqual) GetName() string { return "LessOrEqual" }

// GetName returns the name of the system operator.
func (a *GreaterOrEqual) GetName() string { return "GreaterOrEqual" }

// GetName returns the name of the system operator.
func (a *And) GetName() string { return "And" }

// GetName returns the name of the system operator.
func (a *Or) GetName() string { return "Or" }

// GetName returns the name of the system operator.
func (a *XOr) GetName() string { return "XOr" }

// GetName returns the name of the system operator.
func (a *Implies) GetName() string { return "Implies" }

// GetName returns the name of the system operator.
func (a *Add) GetName() string { return "Add" }

// GetName returns the name of the system operator.
func (a *Subtract) GetName() string { return "Subtract" }

// GetName returns the name of the system operator.
func (a *Multiply) GetName() string { return "Multiply" }

// GetName returns the name of the system operator.
func (a *Divide) GetName() string { return "Divide" }

// GetName returns the name of the system operator.
func (a *Modulo) GetName() string { return "Modulo" }

// GetName returns the name of the system operator.
func (a *TruncatedDivide) GetName() string { return "TruncatedDivide" }

// GetName returns the name of the system operator.
func (a *Before) GetName() string { return "Before" }

// GetName returns the name of the system operator.
func (a *After) GetName() string { return "After" }

// GetName returns the name of the system operator.
func (a *SameOrBefore) GetName() string { return "SameOrBefore" }

// GetName returns the name of the system operator.
func (a *SameOrAfter) GetName() string { return "SameOrAfter" }

// GetName returns the name of the system operator.
func (a *DifferenceBetween) GetName() string { return "DifferenceBetween" }

// GetName returns the name of the system operator.
func (a *In) GetName() string { return "In" }

// GetName returns the name of the system operator.
func (a *IncludedIn) GetName() string { return "IncludedIn" }

// GetName returns the name of the system operator.
func (a *InCodeSystem) GetName() string { return "InCodeSystem" }

// GetName returns the name of the system operator.
func (a *InValueSet) GetName() string { return "InValueSet" }

// GetName returns the name of the system operator.
func (a *Contains) GetName() string { return "Contains" }

// GetName returns the name of the system operator.
func (a *CalculateAgeAt) GetName() string { return "CalculateAgeAt" }

// GetName returns the name of the system operator.
func (a *Except) GetName() string { return "Except" }

// GetName returns the name of the system operator.
func (a *Intersect) GetName() string { return "Intersect" }

// GetName returns the name of the system operator.
func (a *Union) GetName() string { return "Union" }

// NARY EXPRESSION GETNAME()

// GetName returns the name of the system operator.
func (a *Coalesce) GetName() string { return "Coalesce" }

// GetName returns the name of the system operator.
func (a *Concatenate) GetName() string { return "Concatenate" }

// GetName returns the name of the system operator.
func (a *Date) GetName() string { return "Date" }

// GetName returns the name of the system operator.
func (a *DateTime) GetName() string { return "DateTime" }

// GetName returns the name of the system operator.
func (a *Now) GetName() string { return "Now" }

// GetName returns the name of the system operator.
func (a *TimeOfDay) GetName() string { return "TimeOfDay" }

// GetName returns the name of the system operator.
func (a *Time) GetName() string { return "Time" }

// GetName returns the name of the system operator.
func (a *Today) GetName() string { return "Today" }

// GetName returns the name of the system operator.
func (a *AllTrue) GetName() string { return "AllTrue" }

// GetName returns the name of the system operator.
func (c *Count) GetName() string { return "Count" }
