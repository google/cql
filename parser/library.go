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

package parser

import (
	"fmt"
	"strings"

	"github.com/google/cql/internal/embeddata/third_party/cqframework/cql"
	"github.com/google/cql/internal/modelinfo"
	"github.com/google/cql/internal/reference"
	"github.com/google/cql/model"
	"github.com/google/cql/result"
	"github.com/google/cql/types"
	"github.com/antlr4-go/antlr/v4"
	"slices"
)

// VisitLibrary is the top level visitor for parsing a CQL library.
func (v *visitor) VisitLibrary(ctx cql.ILibraryContext) *model.Library {
	library := &model.Library{
		Identifier: v.LibraryIdentifier(ctx),
	}
	v.makeCurrent(library.Identifier, ctx)

	// TODO: b/325500067 - CodeSystems should be visited before all Codes.
	for _, def := range ctx.AllDefinition() {
		switch d := def.GetChild(0).(type) {
		case *cql.IncludeDefinitionContext:
			library.Includes = append(library.Includes, v.VisitIncludeDefinition(d))
		case *cql.UsingDefinitionContext:
			library.Usings = append(library.Usings, v.VisitUsingDefinition(d))
		case *cql.ParameterDefinitionContext:
			library.Parameters = append(library.Parameters, v.VisitParameterDefinition(d))
		case *cql.CodesystemDefinitionContext:
			library.CodeSystems = append(library.CodeSystems, v.VisitCodeSystemDefinition(d))
		case *cql.ConceptDefinitionContext:
			library.Concepts = append(library.Concepts, v.VisitConceptDefinition(d))
		// TODO: b/326331584 - Support the CodeLiteralSelector format.
		case *cql.CodeDefinitionContext:
			library.Codes = append(library.Codes, v.VisitCodeDefinition(d))
		case *cql.ValuesetDefinitionContext:
			library.Valuesets = append(library.Valuesets, v.VisitValuesetDefinition(d))
		default:
			v.reportError("internal error - unsupported definition", ctx)
			// TODO: b/301606416 - Add support for conceptDefinition.
		}
	}

	statements := &model.Statements{}
	for _, stat := range ctx.AllStatement() {
		switch s := stat.GetChild(0).(type) {
		case *cql.ExpressionDefinitionContext:
			statements.Defs = append(statements.Defs, v.VisitExpressionDefinition(s))
		case *cql.FunctionDefinitionContext:
			statements.Defs = append(statements.Defs, v.VisitFunctionDefinition(s))
		case *cql.ContextDefinitionContext:
			statements.Defs = append(statements.Defs, v.VisitContextDefinition(s))
		default:
			v.reportError("internal error - unsupported statement", ctx)
		}
	}

	if len(statements.Defs) > 0 {
		library.Statements = statements
	}

	return library
}

// LibraryIdentifier is the top level visitor for parsing a CQL libraries unique identifier.
// This function is meant to be called by the parser when creating the includes
// dependency graph.
func (v *visitor) LibraryIdentifier(ctx cql.ILibraryContext) *model.LibraryIdentifier {
	if ctx.LibraryDefinition() != nil {
		return v.VisitLibraryDefinition(ctx.LibraryDefinition())
	}
	return nil
}

// makeCurrent sets the current library within the visitor.
// This function should always be called before Visiting CQL context nodes.
func (v *visitor) makeCurrent(libID *model.LibraryIdentifier, ctx cql.ILibraryContext) {
	if libID == nil {
		// TODO(b/298104070): We should add a warning for unnamed libraries. It is unintuitive that you can have an
		// unnamed library where all definitions are private.
		v.refs.SetCurrentUnnamed()
		return
	}
	if err := v.refs.SetCurrentLibrary(libID); err != nil {
		v.reportError(err.Error(), ctx)
		return
	}
}

// VisitParameter is the top level visitor for parsing parameters passed to the CQL Engine.
func (v *visitor) VisitParameter(ctx cql.ITermContext) model.IExpression {
	switch t := ctx.(type) {
	case *cql.IntervalSelectorTermContext:
		return v.VisitIntervalSelectorTerm(t)
	case *cql.ListSelectorTermContext:
		return v.VisitListSelectorTerm(t)
	case *cql.LiteralTermContext:
		return v.VisitLiteralTerm(t)
	}
	v.reportError("must be a interval, list or literal", ctx)
	return nil
}

func (v *visitor) VisitLibraryDefinition(ctx cql.ILibraryDefinitionContext) *model.LibraryIdentifier {
	qID := v.VisitQualifiedIdentifier(ctx.QualifiedIdentifier())
	return &model.LibraryIdentifier{
		Version:   v.VisitVersionSpecifier(ctx.VersionSpecifier()),
		Local:     qID[len(qID)-1],
		Qualified: strings.Join(qID, "."),
	}
}

// VisitIncludeDefinition returns the model for an include statement. It should be called as part of
// parsing the entire library. If you only need to parse the include statements, use
// LibraryIncludedIdentifiers.
func (v *visitor) VisitIncludeDefinition(ctx *cql.IncludeDefinitionContext) *model.Include {
	qID := v.VisitQualifiedIdentifier(ctx.QualifiedIdentifier())
	i := &model.Include{
		Identifier: &model.LibraryIdentifier{
			Qualified: strings.Join(qID, "."),
			Version:   v.VisitVersionSpecifier(ctx.VersionSpecifier()),
		},
	}

	if ctx.LocalIdentifier() != nil {
		i.Identifier.Local = v.VisitIdentifier(ctx.LocalIdentifier().Identifier())
	} else {
		i.Identifier.Local = qID[len(qID)-1]
	}

	if err := v.refs.IncludeLibrary(i.Identifier, true); err != nil {
		v.reportError(err.Error(), ctx)
	}
	return i
}

// LibraryIncludedIdentifiers only parses the include statements returning their LibKeys. It is
// meant to be called by the parser when creating the includes dependency graph.
// LibraryIncludedIdentifiers does not add to the reference resolver.
func (v *visitor) LibraryIncludedIdentifiers(ctx cql.ILibraryContext) []result.LibKey {
	includes := []result.LibKey{}
	for _, def := range ctx.AllDefinition() {
		includeDef, ok := def.GetChild(0).(*cql.IncludeDefinitionContext)
		if ok {
			qID := v.VisitQualifiedIdentifier(includeDef.QualifiedIdentifier())
			key := result.LibKey{
				Name:    strings.Join(qID, "."),
				Version: v.VisitVersionSpecifier(includeDef.VersionSpecifier()),
			}
			includes = append(includes, key)
		}
	}
	return includes
}

func (v *visitor) VisitUsingDefinition(ctx *cql.UsingDefinitionContext) *model.Using {
	using := &model.Using{}

	if ctx.LocalIdentifier() != nil {
		v.reportError(fmt.Sprintf("Using declaration does not support local identifiers but received %v", v.VisitIdentifier(ctx.LocalIdentifier().Identifier())), ctx)
	}

	qID := v.VisitQualifiedIdentifier(ctx.QualifiedIdentifier())
	using.LocalIdentifier = strings.Join(qID, ".")
	if ctx.VersionSpecifier() != nil {
		using.Version = v.VisitVersionSpecifier(ctx.VersionSpecifier())
	}

	key := modelinfo.Key{Name: using.LocalIdentifier, Version: using.Version}
	if err := v.modelInfo.SetUsing(key); err != nil {
		v.reportError(err.Error(), ctx)
		return using
	}

	url, err := v.modelInfo.URL()
	if err != nil {
		v.reportError(err.Error(), ctx)
		return using
	}
	using.URI = url
	if v.currentModelContext == "" {
		v.currentModelContext, err = v.modelInfo.DefaultContext()
		if err != nil {
			v.reportError(err.Error(), ctx)
			return using
		}
		if v.currentModelContext == "" {
			// Fall back to Patient as the default if not provided.
			v.currentModelContext = "Patient"
		}
	}

	return using
}

func (v *visitor) VisitParameterDefinition(ctx cql.IParameterDefinitionContext) *model.ParameterDef {
	p := &model.ParameterDef{
		Name:        v.VisitIdentifier(ctx.Identifier()),
		AccessLevel: v.VisitAccessModifier(ctx.AccessModifier()),
		Element:     &model.Element{},
	}

	if ctx.TypeSpecifier() == nil && ctx.Expression() == nil {
		v.reportError("Parameter definition must include a type or a default, but neither were found", ctx)
	}

	if ctx.TypeSpecifier() != nil {
		p.Element.ResultType = v.VisitTypeSpecifier(ctx.TypeSpecifier())
	}

	// If a default value for the parameter was provided it is stored in the expression.
	if ctx.Expression() != nil {
		p.Default = v.VisitExpression(ctx.Expression())

		if ctx.TypeSpecifier() != nil && !p.Element.ResultType.Equal(p.Default.GetResultType()) {
			// If the specified and default type do not match, report error and use the default type.
			v.reportError(fmt.Sprintf("Parameter definition specified type %s does not match the type of default %s", ctx.TypeSpecifier().GetText(), ctx.Expression().GetText()), ctx)
		}
		// If the default is set the TypeSpecifier is optional, and the ResultType should be inferred
		// from the default.
		p.Element.ResultType = p.Default.GetResultType()
	}

	f := func() model.IExpression {
		return &model.ParameterRef{Name: p.Name, Expression: model.ResultType(p.GetResultType())}
	}
	d := &reference.Def[func() model.IExpression]{
		Name:             p.Name,
		Result:           f,
		IsPublic:         p.AccessLevel == model.Public,
		ValidateIsUnique: true,
	}
	if err := v.refs.Define(d); err != nil {
		v.reportError(err.Error(), ctx)
	}
	return p
}

func (v *visitor) VisitValuesetDefinition(ctx *cql.ValuesetDefinitionContext) *model.ValuesetDef {
	vd := &model.ValuesetDef{
		Name:        v.VisitIdentifier(ctx.Identifier()),
		ID:          parseSTRING(ctx.ValuesetId().STRING()),
		AccessLevel: v.VisitAccessModifier(ctx.AccessModifier()),
		Version:     v.VisitVersionSpecifier(ctx.VersionSpecifier()),
		Element:     &model.Element{ResultType: types.ValueSet},
	}

	if codeSystems := ctx.Codesystems(); codeSystems != nil {
		for _, cs := range codeSystems.AllCodesystemIdentifier() {
			csr := v.parseCodeSystemIdentifier(cs)
			vd.CodeSystems = append(vd.CodeSystems, csr)
		}
	}

	d := &reference.Def[func() model.IExpression]{
		Name: vd.Name,
		Result: func() model.IExpression {
			return &model.ValuesetRef{Name: vd.Name, Expression: model.ResultType(types.ValueSet)}
		},
		IsPublic:         vd.AccessLevel == model.Public,
		ValidateIsUnique: true,
	}
	if err := v.refs.Define(d); err != nil {
		v.reportError(err.Error(), ctx)
	}
	return vd
}

func (v *visitor) VisitCodeSystemDefinition(ctx *cql.CodesystemDefinitionContext) *model.CodeSystemDef {
	cs := &model.CodeSystemDef{
		Name:        v.VisitIdentifier(ctx.Identifier()),
		ID:          parseSTRING(ctx.CodesystemId().STRING()),
		Version:     v.VisitVersionSpecifier(ctx.VersionSpecifier()),
		AccessLevel: v.VisitAccessModifier(ctx.AccessModifier()),
		Element:     &model.Element{ResultType: types.CodeSystem},
	}

	d := &reference.Def[func() model.IExpression]{
		Name: cs.Name,
		Result: func() model.IExpression {
			return &model.CodeSystemRef{Name: cs.Name, Expression: model.ResultType(types.CodeSystem)}
		},
		IsPublic:         cs.AccessLevel == model.Public,
		ValidateIsUnique: true,
	}
	if err := v.refs.Define(d); err != nil {
		v.reportError(err.Error(), ctx)
	}
	return cs
}

func (v *visitor) VisitConceptDefinition(ctx *cql.ConceptDefinitionContext) *model.ConceptDef {
	var display string
	if d := ctx.DisplayClause(); d != nil {
		display = parseSTRING(d.STRING())
	}

	var codes []*model.CodeRef
	for _, codeID := range ctx.AllCodeIdentifier() {
		codes = append(codes, v.VisitCodeIdentifier(codeID))
	}

	c := &model.ConceptDef{
		Name:        v.VisitIdentifier(ctx.Identifier()),
		Codes:       codes,
		Display:     display,
		AccessLevel: v.VisitAccessModifier(ctx.AccessModifier()),
		Element:     &model.Element{ResultType: types.Concept},
	}

	d := &reference.Def[func() model.IExpression]{
		Name: c.Name,
		Result: func() model.IExpression {
			return &model.ConceptRef{Name: c.Name, Expression: model.ResultType(types.Concept)}
		},
		IsPublic:         c.AccessLevel == model.Public,
		ValidateIsUnique: true,
	}
	if err := v.refs.Define(d); err != nil {
		v.reportError(err.Error(), ctx)
	}
	return c
}

func (v *visitor) VisitCodeDefinition(ctx *cql.CodeDefinitionContext) *model.CodeDef {
	c := &model.CodeDef{
		Name:        v.VisitIdentifier(ctx.Identifier()),
		Code:        parseSTRING(ctx.CodeId().STRING()),
		CodeSystem:  v.parseCodeSystemIdentifier(ctx.CodesystemIdentifier()),
		AccessLevel: v.VisitAccessModifier(ctx.AccessModifier()),
		Element:     &model.Element{ResultType: types.Code},
	}

	if d := ctx.DisplayClause(); d != nil {
		c.Display = parseSTRING(d.STRING())
	}

	def := &reference.Def[func() model.IExpression]{
		Name: c.Name,
		Result: func() model.IExpression {
			return &model.CodeRef{Name: c.Name, Expression: model.ResultType(types.Code)}
		},
		IsPublic:         c.AccessLevel == model.Public,
		ValidateIsUnique: true,
	}
	if err := v.refs.Define(def); err != nil {
		v.reportError(err.Error(), ctx)
	}
	return c
}

func (v *visitor) parseCodeSystemIdentifier(ctx cql.ICodesystemIdentifierContext) *model.CodeSystemRef {
	csr := &model.CodeSystemRef{}
	csi := v.VisitIdentifier(ctx.Identifier())
	var libID string
	var csExpr func() model.IExpression
	var err error
	if ctx.LibraryIdentifier() != nil {
		libID = v.VisitIdentifier(ctx.LibraryIdentifier().Identifier())
		csExpr, err = v.refs.ResolveGlobal(libID, csi)
	} else {
		csExpr, err = v.refs.ResolveLocal(csi)
	}
	if err != nil {
		v.reportError(err.Error(), ctx)
		return csr
	}

	csr, ok := csExpr().(*model.CodeSystemRef)
	if !ok {
		fullID := csi
		if libID != "" {
			fullID = libID + "." + csi
		}
		v.reportError(fmt.Sprintf("%v should be of type %v but instead got %v", fullID, types.CodeSystem, csExpr().GetResultType()), ctx)
	}
	return csr
}

func (v *visitor) VisitCodeIdentifier(ctx cql.ICodeIdentifierContext) *model.CodeRef {
	codeRef := &model.CodeRef{}
	cID := v.VisitIdentifier(ctx.Identifier())
	var libID string
	var codeExpr func() model.IExpression
	var err error
	if ctx.LibraryIdentifier() != nil {
		libID = v.VisitIdentifier(ctx.LibraryIdentifier().Identifier())
		codeExpr, err = v.refs.ResolveGlobal(libID, cID)
	} else {
		codeExpr, err = v.refs.ResolveLocal(cID)
	}
	if err != nil {
		v.reportError(err.Error(), ctx)
		return codeRef
	}

	codeRef, ok := codeExpr().(*model.CodeRef)
	if !ok {
		fullID := cID
		if libID != "" {
			fullID = libID + "." + cID
		}
		v.reportError(fmt.Sprintf("expected to find CodeRef for identifier %s, got %v", fullID, codeExpr()), ctx)
	}
	return codeRef
}

func (v *visitor) VisitExpressionDefinition(ctx *cql.ExpressionDefinitionContext) *model.ExpressionDef {
	ed := &model.ExpressionDef{
		Name:        v.VisitIdentifier(ctx.Identifier()),
		Context:     v.currentModelContext,
		AccessLevel: v.VisitAccessModifier(ctx.AccessModifier()),
		Expression:  v.VisitExpression(ctx.Expression()),
	}
	expRef := &model.ExpressionRef{Name: ed.Name}

	// Set the return type of the ExpressionDef to the return type of the inner expression, and set
	// the ExpressionRef's result type.
	if ed.Expression.GetResultType() != nil {
		ed.Element = &model.Element{ResultType: ed.Expression.GetResultType()}
		expRef.Expression = model.ResultType(ed.Expression.GetResultType())
	}

	d := &reference.Def[func() model.IExpression]{
		Name: ed.Name,
		Result: func() model.IExpression {
			return expRef
		},
		IsPublic:         ed.AccessLevel == model.Public,
		ValidateIsUnique: true,
	}
	if err := v.refs.Define(d); err != nil {
		v.reportError(err.Error(), ctx)
	}
	return ed
}

func (v *visitor) VisitContextDefinition(ctx *cql.ContextDefinitionContext) *model.ExpressionDef {
	cname := v.VisitIdentifier(ctx.Identifier())
	var ed *model.ExpressionDef

	if err := validateContext(cname); err != nil {
		return &model.ExpressionDef{
			Name:       cname,
			Expression: v.badExpression(err.Error(), ctx),
		}
	}

	r, err := v.createRetrieve(cname)
	if err != nil {
		return &model.ExpressionDef{
			Name:       cname,
			Expression: v.badExpression(err.Error(), ctx),
		}
	}

	sf := &model.SingletonFrom{
		UnaryExpression: &model.UnaryExpression{
			Operand: r,
		},
	}
	if r.GetResultType() != nil {
		switch opType := r.GetResultType().(type) {
		case *types.List:
			sf.UnaryExpression.Expression = model.ResultType(opType.ElementType)
		case types.System:
			if opType == types.Any {
				sf.UnaryExpression.Expression = model.ResultType(types.Any)
			}
			// This should not happen for context definition, and in future will be handled by overload
			// matching for general singleton from.
			return &model.ExpressionDef{
				Name:       cname,
				Expression: v.badExpression(fmt.Sprintf("SingletonFrom expected a List type or null as input, got: %v", opType), ctx),
			}
		default:
			// This should not happen for context definition, and in future will be handled by overload
			// matching for general singleton from.
			return &model.ExpressionDef{
				Name:       cname,
				Expression: v.badExpression(fmt.Sprintf("SingletonFrom expected a List type or null as input, got: %v", opType), ctx),
			}
		}

		ed = &model.ExpressionDef{
			Name:        cname,
			Context:     cname,
			AccessLevel: model.Private,
			Expression:  sf,
			Element:     &model.Element{ResultType: sf.GetResultType()},
		}
	}

	d := &reference.Def[func() model.IExpression]{
		Name: ed.Name,
		Result: func() model.IExpression {
			return &model.ExpressionRef{Name: ed.Name, Expression: model.ResultType(ed.GetResultType())}
		},
		// Context definitions are always private.
		IsPublic:         false,
		ValidateIsUnique: true,
	}
	if err := v.refs.Define(d); err != nil {
		v.reportError(err.Error(), ctx)
	}
	return ed
}

var supportedContexts = []string{"Patient"}

func validateContext(ctx string) error {
	if !slices.Contains(supportedContexts, ctx) {
		return fmt.Errorf("error -- the CQL engine does not yet support the context %q, only %v are supported", ctx, supportedContexts)
	}
	// TODO: b/329250181 - Also validate contexts against model info, when other contexts are
	// supported.
	return nil
}

func (v *visitor) VisitAccessModifier(ctx cql.IAccessModifierContext) model.AccessLevel {
	if ctx == nil {
		return model.Public
	}
	if ctx.GetChild(0).(antlr.TerminalNode).GetText() == "private" {
		return model.Private
	}
	return model.Public
}

type parsingErrors interface {
	Append(e *ParsingError)
	Error() string
	Unwrap() []error
}

type visitor struct {
	*cql.BaseCqlVisitor

	modelInfo *modelinfo.ModelInfos

	// The current model context, e.g "Patient".
	currentModelContext string

	refs *reference.Resolver[func() model.IExpression, func() model.IExpression]

	// Accumulated parsing errors to be returned to the caller.
	errors parsingErrors

	// Track if we're currently parsing within a sort context
	inSortContext bool
}
