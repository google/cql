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

package parser

import (
	"fmt"

	"github.com/google/cql/internal/embeddata/third_party/cqframework/cql"
	"github.com/google/cql/internal/reference"
	"github.com/google/cql/model"
	"github.com/google/cql/types"
	"github.com/antlr4-go/antlr/v4"
)

// VisitFunctionDefinition parses a user defined function and saves it in the reference resolver.
func (v *visitor) VisitFunctionDefinition(ctx *cql.FunctionDefinitionContext) *model.FunctionDef {
	fd := &model.FunctionDef{
		ExpressionDef: &model.ExpressionDef{
			Element:     &model.Element{},
			Name:        v.parseIdentifierOrFuntionIdentifier(ctx.IdentifierOrFunctionIdentifier()),
			Context:     v.currentModelContext,
			AccessLevel: v.VisitAccessModifier(ctx.AccessModifier()),
		},
		Operands: []model.OperandDef{},
	}

	if ctx.FluentModifier() != nil {
		fd.Fluent = true
	}

	v.refs.EnterScope()
	defer v.refs.ExitScope()

	for _, ops := range ctx.AllOperandDefinition() {
		op := v.VisitOperandDefinition(ops)

		f := func() model.IExpression {
			return &model.OperandRef{Name: op.Name, Expression: model.ResultType(op.GetResultType())}
		}
		if err := v.refs.Alias(op.Name, f); err != nil {
			v.reportError(err.Error(), ctx)
		}

		fd.Operands = append(fd.Operands, op)
	}

	if ctx.FunctionBody() != nil {
		fd.Expression = v.VisitExpression(ctx.FunctionBody().Expression())

		returnType := v.VisitTypeSpecifier(ctx.TypeSpecifier())
		if returnType != nil && !returnType.Equal(fd.Expression.GetResultType()) {
			v.reportError(fmt.Sprintf("function body return type %v, does not match the specified return %v", fd.Expression.GetResultType(), returnType), ctx)
		}
		fd.ResultType = fd.Expression.GetResultType()
	} else {
		fd.External = true
		returnType := v.VisitTypeSpecifier(ctx.TypeSpecifier())
		if returnType != nil {
			fd.ResultType = returnType
		}
	}

	operandRef := []types.IType{}
	for _, op := range fd.Operands {
		operandRef = append(operandRef, op.GetResultType())
	}

	f := &reference.Func[func() model.IExpression]{
		Name:     fd.Name,
		Operands: operandRef,
		// The Operands are left as nil, they will be set when we parse when this function is called.
		Result: func() model.IExpression {
			return &model.FunctionRef{Name: fd.Name, Operands: nil, Expression: model.ResultType(fd.ResultType)}
		},
		IsPublic:         fd.AccessLevel == model.Public,
		IsFluent:         fd.Fluent,
		ValidateIsUnique: true,
	}
	if err := v.refs.DefineFunc(f); err != nil {
		v.reportError(err.Error(), ctx)
	}

	return fd
}

func (v *visitor) VisitOperandDefinition(ctx cql.IOperandDefinitionContext) model.OperandDef {
	return model.OperandDef{
		Name:       v.parseReferentialIdentifier(ctx.ReferentialIdentifier()),
		Expression: model.ResultType(v.VisitTypeSpecifier(ctx.TypeSpecifier())),
	}
}

// parseQualifiedFunction parses invocations of a global user defined function.
func (v *visitor) parseQualifiedFunction(ctx *cql.QualifiedFunctionContext, libraryName string) model.IExpression {
	name := v.parseIdentifierOrFuntionIdentifier(ctx.IdentifierOrFunctionIdentifier())
	params := []antlr.Tree{}
	if ctx.ParamList() != nil {
		for _, expr := range ctx.ParamList().AllExpression() {
			params = append(params, expr)
		}
	}
	m, err := v.parseFunction(libraryName, name, params, false)
	if err != nil {
		return v.badExpression(err.Error(), ctx)
	}
	return m
}

// VisitFunction parses invocations of a local user defined function or built-in function.
func (v *visitor) VisitFunction(ctx *cql.FunctionContext) model.IExpression {
	name := v.parseReferentialIdentifier(ctx.ReferentialIdentifier())
	params := []antlr.Tree{}
	if ctx.ParamList() != nil {
		for _, expr := range ctx.ParamList().AllExpression() {
			params = append(params, expr)
		}
	}
	m, err := v.parseFunction("", name, params, false)
	if err != nil {
		return v.badExpression(err.Error(), ctx)
	}
	return m
}
