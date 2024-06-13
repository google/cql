# Overview

This document provides an overview of the CQL Engine codebase. If you just want to use the engine, check out the [getting started](../README.md) section.

**[parser](../parser)** - The parser validates and converts CQL strings into an Abstract Syntax Tree called Expression Model Language [ELM](https://cql.hl7.org/elm.html).\
**[interpreter](../interpreter)** - The interpreter evaluates the ELM using a retriever and terminology provider and returns the results.\
**[model](../model)** - Model is our internal representation of ELM. Model is produced by the parser and evaluated by the interpreter.\
**[retriever](../retriever)** - The retriever is an interface called by the interpreter to retrieve data from external databases. There are implementations to connect to a local FHIR bundle or a FHIR bundle in GCP's GCS. Developers can implement their own retrievers to connect to other databases.\
**[terminology](../terminology)** - Terminology provider is an interface called by the interpreter to fetch value sets from terminology servers. There are implementations to connect to local FHIR value sets. Developers can provide their own implementation to connect to other terminology servers.\
**[result](../result)** - Result is our internal representation of CQL values. They are evaluated and returned by the interpreter.\
**[types](../types)** - Types holds a representation of the CQL type system.\
**[internal/modelinfo](../internal/modelinfo)** - Modelinfo parses model info XML files and provides high level functions like `IsSubType(child, base types.IType)` to interact with the data models.\
**[internal/reference](../internal/reference)** - The reference resolver is shared by both the parser and the interpreter. It is responsible for storing and resolving references to expression definitions, aliases, function signatures, system operators, parameters... and more both within and across CQL libraries.\
**[internal/convert](../internal/convert)** - Convert handles overload matching with implicit conversions.

# Parser

The [Parser package](../parser), parses CQL strings into Expression Model Language [ELM](https://cql.hl7.org/elm.html) that the [Interpreter package](../interpreter) can then evaluate. The [Model package](../model) holds our ELM like data structure. The Model package is almost completely one to one with ELM, with a few exceptions to make it work better in Golang. For example, in ELM [Binary Expression](https://cql.hl7.org/04-logicalspecification.html#binaryexpression) inherits from Operator Expression which inherits from Expression. Golang is not an object oriented language. Deep hierarchies are not useful so we dropped Operator Expressions. As much as possible we try to keep the Model package one to one with ELM so in the future we can easily import and export ELM directly.

The parser uses ANTLR to implement a visitor pattern over the [CQL grammar](../internal/embeddata/cqframework/Cql.g4). Each VisitXXX is responsible for taking the ANTLR context and outputting a piece of the model.go tree. The CQL grammar is made up of both the Cql.g4 and fhirpath.g4 files.

The parser is responsible for all validation. ANTLR and the grammar do some validation work, for example ANTLR would error when parsing `@20155-01-30` since it does not meet grammar's DATEFORMAT. But much of the work of validation needs to be implemented. For example, `@2015-01-99` meets the grammar but should fail since there are not 99 days in January. Errors that occur are not immediately returned. Instead we return a placeholder model.go (usually via v.badExpression()) and continue parsing. In the end a list of all errors as ParsingErrors are returned via the top level API.

The parser is also responsible for overload matching and implicit conversions. Take for example, `1 + 4.5`. The [Add system operator](https://cql.hl7.org/09-b-cqlreference.html#add) defines the `+(left Decimal, right Decimal) Decimal` overload among others. The Parser uses the [conversion precedence](https://cql.hl7.org/03-developersguide.html#conversion-precedence) to score each of the Add overloads. If there is no matching overload an error is returned. If two overloads tie for the minimum score an ambiguous error is returned. Otherwise, the minimum scoring overload is returned with any necessary implicit conversions inserted. In this case a `model.ToDecimal` would be inserted to convert the 1 to 1.0. The Parser inserts all necessary model.ToXXX, model.As and FHIRHelpers.ToXXX from the conversion info in the model info.

CQL supports parameters which can override the parameters in a Library. Parameters have restrictions like not referencing expression definitions and being computable at "compile-time". To support parameters the parser takes parameters as strings and parses them starting at Term in the CQL grammar. Starting at Term in the grammar restricts the parameters to only being CQL Literals or selectors like `Interval[@2013-01-01, @2014-01-01)`. The parser does not support parameters like `1 + 2` which according to the CQL spec should be allowed.

In the parser, we initialize the reference resolver with `reference.Resolver[func() model.IExpression, func() model.IExpression]` This is because the reference resolver should return copies of the resolved model structs. The easiest way to accomplish this in Golang was to return a function that can generate a new struct.

## Example - Parsing System Operator ToDate
CQL System Operators are any unary, binary, ternary or nary expression that inherit from [operator expression](https://cql.hl7.org/04-logicalspecification.html#operatorexpression). This includes operators like `+` and system functions like `ToDate()`. System Operators do not include things like Query or Parameters. System operators are described in detail in the [CQL reference](https://cql.hl7.org/09-b-cqlreference.html). System Operators have overloads that need to be supported. Taking [Less()](https://cql.hl7.org/09-b-cqlreference.html#less) as an example:

```
<(left Integer, right Integer) Boolean
<(left Long, right Long) Boolean
<(left Decimal, right Decimal) Boolean
<(left Quantity, right Quantity) Boolean
<(left Date, right Date) Boolean
<(left DateTime, right DateTime) Boolean
<(left Time, right Time) Boolean
<(left String, right String) Boolean
```

To add support for parsing a system operator add a struct to the `loadSystemOperators()` function in [operators.go](../parser/operators.go) file.

```go
{
	name: "Less",
	operands: [][]types.IType{
		[]types.IType{types.Integer, types.Integer},
		[]types.IType{types.Long, types.Long},
		[]types.IType{types.Decimal, types.Decimal},
		[]types.IType{types.Quantity, types.Quantity},
		[]types.IType{types.Date, types.Date},
		[]types.IType{types.DateTime, types.DateTime},
		[]types.IType{types.Time, types.Time},
		[]types.IType{types.String, types.String},
	},
	model: func() model.IExpression {
		return &model.Less{
			BinaryExpression: &model.BinaryExpression{
				Expression: model.ResultType(types.Boolean),
			},
		}
	},
},
```

`loadSystemOperators()` adds `Less(Date)` and other overloads to the [reference resolver](../internal/reference/reference.go). When we call `ResolveLocalFunc("Less", operands...)` the reference resolver calls [convert.OverloadMatch()](../internal/convert/convert.go) with all the Less overloads that have been loaded. `convert.OverloadMatch` returns the least converting match by summing the score of each of the operands  according to the [conversion precedence](https://cql.hl7.org/03-developersguide.html#conversion-precedence).

Let's take a look at what would happen if we called `Less()` with an operand of `Named<FHIR.date>`. Many conversions are hard coded, but some are defined in the data model. The [modelinfo package](../internal/modelinfo) parses model info xml files saving all named types, their properties and their implicit conversions. When we call `modelinfo.IsImplicitlyConvertible(from, to types.IType)` it returns `FHIRHelpers.ToDate` which converts the data model's `Named<FHIR.date>` to a `System.Date`, matching our overload. `convert.OverloadMatch` finds this is the least converting match with a score of 5 (Implicit Conversion To Simple Type) and returns the Operands wrapped in a `model.FunctionRef`.

```go
&model.FunctionRef{
	Expression:  model.ResultType(types.Date),
	Name:        "ToDate",
	LibraryName: "FHIRHelpers",
	Operands: []model.IExpression{...Original Operand...},
},
```

System operators like less can be called in CQL as a function `Less(1, 2)` or non functional syntax `1 < 2`. The functional syntax is defined in the [translation semantics](https://cql.hl7.org/06-translationsemantics.html#functions). In cases like these ensure that less is added to the `loadSystemOperators()`. Then implement a visitor to handle the non function call grammar. The visitor should call `v.parseFunction()` which validates that the correct type was passed to the system operator.

```go
func (v *visitor) VisitInequalityExpression(ctx *cql.InequalityExpressionContext) model.IExpression {
	name := ctx.GetChild(1).(antlr.TerminalNode).GetText()
	var m model.IExpression
	var err error
	switch name {
	case "<":
		m, err = v.parseFunction("", "Less", []antlr.Tree{ctx.Expression(0), ctx.Expression(1)}, false)
	case ">":
		m, err = v.parseFunction("", "Greater", []antlr.Tree{ctx.Expression(0), ctx.Expression(1)}, false)
	case "<=":
		m, err = v.parseFunction("", "LessOrEqual", []antlr.Tree{ctx.Expression(0), ctx.Expression(1)}, false)
	case ">=":
		m, err = v.parseFunction("", "GreaterOrEqual", []antlr.Tree{ctx.Expression(0), ctx.Expression(1)}, false)
	default:
		return v.badExpression("internal error - grammar should not allow this InequalityExpression", ctx)
	}
	if err != nil {
		return v.badExpression(err.Error(), ctx)
	}

	return m
}
```

Finally, a small quirk, for system operators that can be called in multiple ways, such as exists, the function call syntax often flows through `VisitParenthesizedTerm` and not `VisitFunction` as you may expect.

# Interpreter

The interpreter traverses the ELM like tree represented by the [Model package](../model) to evaluate the final results. The [Result package](../result) holds our representation of CQL values (ex CQL Integer or CQL Tuple) and helper functions to convert between CQL Values and Golang Values (ex ToInt32 which converts a CQL Integer to a Golang int32).

For performance reasons the interpreter does very little validation and assumes that it is passed the correct model. For example, in the parser the reference package validates that the function signatures are unique. If you define `Foo(a Integer)` twice you will get an error. In the interpreter we assume that the model.go is correct and do not check that function signatures are unique. The interpreter also does not do any implicit conversions. The parser should have inserted any necessary operators or calls to FHIRHelper conversion functions such that types of the CQL Values exactly match the overloads of the system operators. ELM that may be valid according to the ELM spec, if translated directly to our internal model would not necessarily meet the validation requirements of our interpreter. In the future if we support ELM, the ELM to model.go mapping will need to do additional validation ensure it meets the assumptions of our interpreter.

For each resulting CQL Value our interpreter builds a tree that can be used for debugging or explainability. Each CQL Value stores a Source Expression holding the expression that was used to calculate the Value and Source Values with the Values used by the Source Expression. For example, if we return a CQL Value of 9 resulting from `4 + 5`, then the Source Expression would be `model.Add` and the Source Values would be CQL Value 4 and 5. Since Source Values can also have their own Source Expression and Source Values a tree showing what went into calculating each CQL result is built.

## Example - Evaluating Add

The Operator Dispatcher is the core framework in the Interpreter that handles matching the correct overload of a CQL System Operator. Taking the [Add operator](https://cql.hl7.org/09-b-cqlreference.html#add) as an example. There are many overloads that need to be supported.

```
+(left Integer, right Integer) Integer
+(left Long, right Long) Long
+(left Decimal, right Decimal) Decimal
+(left Quantity, right Quantity) Quantity
+(left Date, right Quantity) Date
+(left DateTime, right Quantity) DateTime
+(left Time, right Quantity) Time
```

To add support for `+(left Decimal, right Decimal) Decimal` we first register the overload in [operator_dispatcher.go](../interpreter/operator_dispatcher.go). The Operator Dispatcher will now call `evalArithmeticDecimal` when we receive a model.Add with operands of type Decimal. The parser will insert all of the conversions necessary to exactly match one of the overloads.

``` go
case *model.Add, *model.Subtract, *model.Multiply, *model.TruncatedDivide, *model.Modulo:
  return []types.Overload[evalBinarySignature]{
    {
      Operands: []types.IType{types.Integer, types.Integer},
      Result:   evalArithmeticInteger,
    },
    {
      Operands: []types.IType{types.Long, types.Long},
      Result:   evalArithmeticLong,
    },
    {
      Operands: []types.IType{types.Decimal, types.Decimal},
      Result:   evalArithmeticDecimal,
    },
  }, true, nil
```

Now that the overload is registered we need to implement `evalArithmeticDecimal`. The golang files are organized following the layout of the CQL Reference so we add the implementation to [operator_arithmetic.go](../internal/operator_arithmetic.go). Every overload implementation starts with handling the null cases. We can then get the golang floats by calling `result.ToFloat64()`.

``` go
func evalArithmeticDecimal(m model.IBinaryExpression, lVal result.Value, rVal result.Value) (result.Value, error) {
	if result.IsNull(lVal) || result.IsNull(rVal) {
		return result.New(nil)
	}
	l, err := result.ToFloat64(lVal)
	if err != nil {
		return nil, err
	}
	r, err := result.ToFloat64(rVal)
	if err != nil {
		return nil, err
	}

	return arithmetic(m, l, r)
}
```

Code sharing between overloads should generally happen in helper functions that take golang values like `arithmetic[n float64 | int64 | int32](m model.IBinaryExpression, l, r n) (result.Value, error)`. There should be explicit overloads `evalArithmeticInteger`, `evalArithmeticLong`, `evalArithmeticDecimal`... that handle the logic specific to that overload before calling the shared helper. In some special cases like `Last(List<T>) T` there can be one generic overload.

# Testing

- The Parser is tested independently with unit tests. Every *.go file in the Parser package is accompanied with a *_test.go file.
- Most of the "unit tests" for the interpreter are actually integration tests located in [tests/enginetests](../tests/enginetests). This was chosen over traditional unit tests, since an update to the parser meant all the model.go inputs to the interpreter unit tests would also need to be updated. Over time this led to the interpreter unit tests becoming out of date with what the parser produced. Enginetests start with CQL strings instead of model.go, but are otherwise treated as unit tests.
- [Largetests](../tests/largetests) are end-to-end tests, designed for complex CQL and complex input data. LargeTests also holds our benchmarks.
- [Spectests](../tests/spectests) are external CQL tests imported from https://github.com/cqframework/cql-tests.
