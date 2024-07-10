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

// Package parser offers a CQL parser that produces an intermediate ELM like data structure for
// evaluation.
package parser

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/cql/internal/embeddata/third_party/cqframework/cql"
	"github.com/google/cql/internal/modelinfo"
	"github.com/google/cql/internal/reference"
	"github.com/google/cql/model"
	"github.com/google/cql/result"
	"github.com/antlr4-go/antlr/v4"
	"gopkg.in/gyuho/goraph.v2"
)

// Config configures the parsing of CQL.
type Config struct {
	// Empty for now, but in the future will contain options like EnableListPromotion.
}

// New returns a new Parser initialized to the data models.
func New(ctx context.Context, dataModels [][]byte) (*Parser, error) {
	p := &Parser{
		refs: reference.NewResolver[func() model.IExpression, func() model.IExpression](),
	}
	mi, err := modelinfo.New(dataModels)
	if err != nil {
		return nil, err
	}
	p.modelInfo = mi

	if err := p.loadSystemOperators(); err != nil {
		return nil, err
	}

	return p, nil
}

// Parser parses CQL library and parameter strings into our intermediate ELM like data structure.
// The parser is responsible for all validation and implicit conversions.
type Parser struct {
	modelInfo *modelinfo.ModelInfos
	refs      *reference.Resolver[func() model.IExpression, func() model.IExpression]
}

// DataModel returns the parsed model info.
func (p *Parser) DataModel() *modelinfo.ModelInfos {
	p.modelInfo.ResetUsing()
	return p.modelInfo
}

// Libraries parses the CQL libraries into a list of model.Library or an error.
// Underlying parsing issues will return a ParsingErrors struct that users can check for and
// report to the user accordingly.
// TODO: b/332337287 - Investigate returning results as a map now that libraries are being sorted.
func (p *Parser) Libraries(ctx context.Context, cqlLibs []string, config Config) ([]*model.Library, error) {
	if cqlLibs == nil || len(cqlLibs) == 0 {
		return nil, result.NewEngineError("", result.ErrLibraryParsing, fmt.Errorf("no CQL libraries were provided"))
	}

	p.refs.ClearDefs()
	sortedLibraries, err := p.topologicalSortLibraries(cqlLibs)
	if err != nil {
		// TODO: b/301606416 Return errors with library name from topological sort.
		return nil, result.NewEngineError("", result.ErrLibraryParsing, err)
	}

	libs := []*model.Library{}
	for _, lexedLib := range sortedLibraries {
		vis := visitor{
			BaseCqlVisitor: &cql.BaseCqlVisitor{},
			errors:         &LibraryErrors{LibKey: lexedLib.key},
			modelInfo:      p.modelInfo,
			refs:           p.refs,
		}
		lib := vis.VisitLibrary(lexedLib.ctx)
		if len(vis.errors.Unwrap()) > 0 {
			return nil, vis.errors
		}
		libs = append(libs, lib)
	}
	return libs, nil
}

type lexedLib struct {
	key result.LibKey
	ctx cql.ILibraryContext
}

// topologicalSortLibraries parses the CQL libraries into ANTLR, topologically sorts their
// dependencies and returns a sorted list of lexedLib.
// Note: In cases where a library is included without a version, it attempts to find the latest
// version of the library with a string value comparison. This is a naive approach and may not work
// with non numerical versioning systems.
func (p *Parser) topologicalSortLibraries(cqlLibs []string) ([]lexedLib, error) {
	// lexedLibraries maps graph ID to library that has been lexed.
	lexedLibraries := make(map[string]lexedLib, len(cqlLibs))
	// includeDependencies maps a library to its dependencies.
	includeDependencies := make(map[result.LibKey][]result.LibKey, len(cqlLibs))
	graph := goraph.NewGraph()

	for _, cqlText := range cqlLibs {
		vis := visitor{
			BaseCqlVisitor: &cql.BaseCqlVisitor{},
			errors:         &LibraryErrors{},
			modelInfo:      p.modelInfo,
			refs:           p.refs,
		}

		lex := cql.NewCqlLexer(antlr.NewInputStream(cqlText))
		par := cql.NewCqlParser(antlr.NewCommonTokenStream(lex, 0))

		lex.AddErrorListener(vis)
		par.AddErrorListener(vis)

		libContext := par.Library()
		libKey := result.LibKeyFromModel(vis.LibraryIdentifier(libContext))

		// Return if the lexer found syntax errors.
		if len(vis.errors.Unwrap()) > 0 {
			libErrs := vis.errors.(*LibraryErrors)
			// Need to set the libKey here because it is not included when we create the visitor.
			libErrs.LibKey = libKey
			return nil, libErrs
		}

		lexedLibraries[libKey.Key()] = lexedLib{key: libKey, ctx: libContext}
		includeDependencies[libKey] = vis.LibraryIncludedIdentifiers(libContext)

		if ok := graph.AddNode(goraph.NewNode(libKey.Key())); !ok {
			return nil, fmt.Errorf("cql library %q already imported", libKey.String())
		}
	}
	// Build graph DAG links.
	for libID, deps := range includeDependencies {
		libNode := goraph.NewNode(libID.Key())
		for _, includedID := range deps {
			// If the version is not specified, use the latest version. This mimics the behavior found in
			// the reference resolver.
			if includedID.Version == "" {
				for libKey := range includeDependencies {
					if libKey.Name != includedID.Name {
						continue
					}
					if strings.Compare(includedID.Version, libKey.Version) == -1 {
						includedID = libKey
					}
				}
			}
			includedNode := goraph.NewNode(includedID.Key())
			if err := graph.AddEdge(includedNode.ID(), libNode.ID(), 1); err != nil {
				return nil, fmt.Errorf("failed to import library %q, dependency graph could not resolve with error: %w", includedID, err)
			}
		}
	}
	sortedLibraryIDs, isValidDag := goraph.TopologicalSort(graph)
	if !isValidDag {
		// TODO: b/332600632 - Add which library has circular dependencies to error output.
		return nil, fmt.Errorf("included cql libraries are not valid, found circular dependencies")
	}

	sortedLibs := make([]lexedLib, 0, len(sortedLibraryIDs))
	for _, libID := range sortedLibraryIDs {
		sortedLibs = append(sortedLibs, lexedLibraries[libID.String()])
	}
	return sortedLibs, nil
}

// Parameters parses CQL literals into model.IExpressions. Each param should be a CQL literal, not an
// expression definition, valueset or other CQL construct.
func (p *Parser) Parameters(ctx context.Context, params map[result.DefKey]string, config Config) (map[result.DefKey]model.IExpression, error) {
	if params == nil {
		return nil, nil
	}
	parsedParams := make(map[result.DefKey]model.IExpression, len(params))
	for k, v := range params {
		e, err := p.parameter(k, v)
		if err != nil {
			return nil, err
		}
		parsedParams[k] = e
	}
	return parsedParams, nil
}

// parameter parses an individual CQL literal. The CQL spec does not specify anything beyond that
// the environment passes parameters. We have chosen to take passed parameters as CQL literals.
func (p *Parser) parameter(key result.DefKey, input string) (model.IExpression, error) {
	p.refs.ClearDefs()

	vis := visitor{
		BaseCqlVisitor: &cql.BaseCqlVisitor{},
		errors:         &ParameterErrors{DefKey: key, Errors: []*ParsingError{}},
		modelInfo:      p.modelInfo,
		refs:           p.refs,
	}
	lex := cql.NewCqlLexer(antlr.NewInputStream(input))
	par := cql.NewCqlParser(antlr.NewCommonTokenStream(lex, 0))

	lex.AddErrorListener(vis)
	par.AddErrorListener(vis)

	// We start parsing at the term precedence in the CQL grammar. VisitParameter() will throw an
	// error if the parsed CQL is not a literal.
	t := par.Term()
	if len(vis.errors.Unwrap()) > 0 {
		return nil, vis.errors
	}

	// Remove whitespace and check if the input string is equal to to the parsed string. This ensures
	// there is no extraneous input that the parser did not match.
	if strings.Join(strings.Fields(input), "") != t.GetText() {
		return nil, &ParameterErrors{
			DefKey: key,
			Errors: []*ParsingError{{Message: "must be a single literal"}},
		}
	}

	m := vis.VisitParameter(t)
	if len(vis.errors.Unwrap()) > 0 {
		return nil, vis.errors
	}
	return m, nil
}

// printTree is a debugging function that prints the entire parsed Antlr Tree.
func printTree(vis visitor, input string) {
	lex := cql.NewCqlLexer(antlr.NewInputStream(input))
	par := cql.NewCqlParser(antlr.NewCommonTokenStream(lex, 0))

	lex.AddErrorListener(vis)
	par.AddErrorListener(vis)

	fmt.Printf("Antlr Parsed Tree \n%v\n", antlr.TreesStringTree(par.Library(), par.GetRuleNames(), par))
}
