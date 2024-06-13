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

// Package reference handles resolving references across CQL libraries and locally within a library
// for the CQL Engine parser and interpreter.
package reference

import (
	"errors"
	"fmt"

	"github.com/google/cql/internal/convert"
	"github.com/google/cql/internal/modelinfo"
	"github.com/google/cql/model"
	"github.com/google/cql/result"
	"github.com/google/cql/types"
)

// Resolver tracks definitions (ExpressionDefs, ParameterDefs, ValueSetDefs...) and aliases across
// CQL libraries and locally within a CQL library. When a definition is created the resolver stores
// a result (for the parser it stores a model.IExpression and for the interpreter a result.Value).
// When a reference to a definition is resolved the result is returned. Resolvers should not be
// shared between the parser and interpreter. A new empty resolver should be passed to the
// interpreter.
type Resolver[T any, F any] struct {
	// defs holds all expression, value sets and parameter definitions in all CQL libraries. funcs
	// holds all built-in functions and all user defined functions in all CQL libraries. The current
	// CQL library does not have access to every definition in defs, or function in funcs, only
	// included libraries which is mapped by includedLibs. Functions and definitions live in separate
	// namespaces, a function can have the same name as an expression definition. Functions can be
	// overloaded, having the same name but different operands.
	defs  map[defKey]exprDef[T]
	funcs map[defKey][]funcDef[F]

	// builtinFuncs holds all CQL built-in functions. builtinFuncs is only used by the parser. The
	// parser coverts all built-in function calls into specific structs in model.go, so the
	// interpreter does not need to resolve any built-in functions. Built-in functions are different
	// from user defined functions because they do not belong to a particular library or have access
	// modifiers.
	builtinFuncs map[string][]convert.Overload[F]

	// aliases, unlike the other maps are not persistent between CQL libraries or even across scopes.
	// Aliases work like a stack and are cleared once we exit the scope in which the alias was
	// defined. Aliases live in the same namespace as definitions.
	aliases []map[aliasKey]T

	// libs holds the qualified identifier of all named libraries that have been parsed.
	libs map[namedLibKey]struct{}

	// includedLibs maps the local identifier of included libraries to their qualified identifier. The
	// following CQL would result in:
	//
	// library measure
	// include helpers called help
	//
	// includedLibs[{localID: "help", includedBy: {qualified: "measure"}}]: {Qualified: "helpers", ...}
	includedLibs map[includeKey]*model.LibraryIdentifier

	// currLib is the current library being parsed.
	currLib libKey

	// unnamedCount is used to generate a unique unnamedLibKey for unnamed libraries.
	unnamedCount int
}

type exprDef[T any] struct {
	isPublic bool
	result   T
}

type funcDef[F any] struct {
	isPublic bool
	isFluent bool
	overload convert.Overload[F]
}

// NewResolver creates a blank resolver with zero global references. Type T is the type saved and
// resolved for definitions. Type F is the type saved and resolved for functions.
func NewResolver[T any, F any]() *Resolver[T, F] {
	return &Resolver[T, F]{
		defs:         make(map[defKey]exprDef[T]),
		funcs:        make(map[defKey][]funcDef[F]),
		builtinFuncs: make(map[string][]convert.Overload[F]),
		aliases:      make([]map[aliasKey]T, 0),
		libs:         make(map[namedLibKey]struct{}),
		includedLibs: make(map[includeKey]*model.LibraryIdentifier),
	}
}

// ClearDefs clears everything except for the built-in functions.
func (r *Resolver[T, F]) ClearDefs() {
	r.defs = make(map[defKey]exprDef[T])
	r.funcs = make(map[defKey][]funcDef[F])
	r.aliases = make([]map[aliasKey]T, 0)
	r.libs = make(map[namedLibKey]struct{})
	r.includedLibs = make(map[includeKey]*model.LibraryIdentifier)
}

// SetCurrentLibrary sets the current library based on the library definition. Either
// SetCurrentLibrary or SetCurrentUnnamed must be called before creating and resolving references.
func (r *Resolver[T, F]) SetCurrentLibrary(m *model.LibraryIdentifier) error {
	l := namedLibKey{qualified: m.Qualified, version: m.Version}
	if _, ok := r.libs[l]; ok {
		return fmt.Errorf("library %s %s already exists", m.Qualified, m.Version)
	}
	r.currLib = l
	r.libs[l] = struct{}{}
	return nil
}

// SetCurrentUnnamed should be called if the CQL library does not have a library definition. All
// definitions in unnamed libraries are private.
func (r *Resolver[T, F]) SetCurrentUnnamed() {
	l := unnamedLibKey{unnamedID: r.unnamedCount}
	r.currLib = l
	r.unnamedCount++
}

// IncludeLibrary should be called for each include statement in the CQL library. IncludeLibrary
// must be called before a reference to that library is resolved. ValidateIsUnique validates this
// include is unique. It is turned off by the
// interpreter to improve performance.
func (r *Resolver[T, F]) IncludeLibrary(m *model.LibraryIdentifier, validateIsUnique bool) error {
	if validateIsUnique {
		if err := r.isLocallyUnique(m.Local); err != nil {
			return err
		}
	}

	lib := namedLibKey{qualified: m.Qualified, version: m.Version}
	if _, ok := r.libs[lib]; !ok {
		return fmt.Errorf("library %s %s was included, but does not exist", m.Qualified, m.Version)
	}

	r.includedLibs[includeKey{localID: m.Local, includedBy: r.currLib}] = m
	return nil
}

// ResolveInclude takes the local name of an included library and returns the fully qualified
// identifier or nil if this local name does not exist.
func (r *Resolver[T, F]) ResolveInclude(name string) *model.LibraryIdentifier {
	iKey := includeKey{localID: name, includedBy: r.currLib}
	if i, ok := r.includedLibs[iKey]; ok {
		return i
	}
	return nil
}

// Def holds the information needed to define a definition.
type Def[T any] struct {
	Name     string
	Result   T
	IsPublic bool
	// ValidateIsUnique validates this definition name is unique. It is turned off by the interpreter
	// to improve performance.
	ValidateIsUnique bool
}

// Define creates a new definition returning an error if the name already exists. Calling
// ResolveLocal with the same name will return the stored type t. Names must be unique within the
// CQL library. Names must be unique regardless of type, for example a ValueSet and Parameter cannot
// have the same name.
func (r *Resolver[T, F]) Define(d *Def[T]) error {
	if d.ValidateIsUnique {
		if err := r.isLocallyUnique(d.Name); err != nil {
			return err
		}
	}

	_, isUnamed := r.currLib.(unnamedLibKey)
	r.defs[defKey{r.currLib, d.Name}] = exprDef[T]{isPublic: d.IsPublic && !isUnamed, result: d.Result}
	return nil
}

// Func holds the information needed to define a function.
type Func[F any] struct {
	Name     string
	Operands []types.IType
	Result   F
	IsPublic bool
	IsFluent bool
	// ValidateIsUnique validates this function name + signature is unique. It is turned off by the
	// interpreter to improve performance.
	ValidateIsUnique bool
}

// DefineFunc creates a new user defined function returning an error if the function signature
// already exists. Calling ResolveLocalFunc with the same name and operands will return the stored
// type f. Functions can be overloaded with the same name, but must have a unique combination of
// name and operands.
func (r *Resolver[T, F]) DefineFunc(f *Func[F]) error {
	if f.ValidateIsUnique {
		if err := r.isFuncLocallyUnique(f.Name, f.Operands); err != nil {
			return err
		}
	}

	dKey := defKey{r.currLib, f.Name}
	_, isUnamed := r.currLib.(unnamedLibKey)
	fDef := funcDef[F]{
		isPublic: f.IsPublic && !isUnamed,
		isFluent: f.IsFluent,
		overload: convert.Overload[F]{Operands: f.Operands, Result: f.Result},
	}
	r.funcs[dKey] = append(r.funcs[dKey], fDef)
	return nil
}

// DefineBuiltinFunc creates a new built-in function, returning an error if the function signature
// already exists. All built in functions must be defined before CQL libraries are parsed. Only the
// Parser defines built-in functions.
// TODO(b/301606416): Refactor DefineBuiltinFunc into the initialization of the reference resolver.
func (r *Resolver[T, F]) DefineBuiltinFunc(name string, operands []types.IType, f F) error {
	if overloads, ok := r.builtinFuncs[name]; ok {
		for _, overload := range overloads {
			if exactMatch(operands, overload.Operands) {
				return fmt.Errorf("internal error - built-in CQL function %v(%v) already exists", name, types.ToStrings(operands))
			}
		}
	}

	r.builtinFuncs[name] = append(r.builtinFuncs[name], convert.Overload[F]{Operands: operands, Result: f})
	return nil
}

// ResolveGlobal resolves a reference to a definition in an included CQL library.
func (r *Resolver[T, F]) ResolveGlobal(libName string, defName string) (T, error) {
	iKey := includeKey{localID: libName, includedBy: r.currLib}
	qKey, ok := r.includedLibs[iKey]
	if !ok {
		return zero[T](), fmt.Errorf("could not resolve the library name %s", libName)
	}

	dKey := defKey{namedLibKey{qualified: qKey.Qualified, version: qKey.Version}, defName}
	a, ok := r.defs[dKey]
	if !ok {
		return zero[T](), fmt.Errorf("could not resolve the reference to %s.%s", libName, defName)
	}
	if !a.isPublic {
		return zero[T](), fmt.Errorf("%s.%s is not public", libName, defName)
	}

	return a.result, nil
}

// ResolveGlobalFunc resolves a reference to a user defined function in an included CQL library.
func (r *Resolver[T, F]) ResolveGlobalFunc(libName string, defName string, operands []model.IExpression, calledFluently bool, modelInfo *modelinfo.ModelInfos) (*convert.MatchedOverload[F], error) {
	iKey := includeKey{localID: libName, includedBy: r.currLib}
	qKey, ok := r.includedLibs[iKey]
	if !ok {
		return nil, fmt.Errorf("could not resolve the library name %s", libName)
	}

	dKey := defKey{namedLibKey{qualified: qKey.Qualified, version: qKey.Version}, defName}
	var overloads []convert.Overload[F]
	if fDefs, ok := r.funcs[dKey]; ok {
		// Filter overloads that are not public or fluent before calling OverloadMatch.
		for _, fDef := range fDefs {
			if fDef.isPublic {
				if !calledFluently || (calledFluently && fDef.isFluent) {
					overloads = append(overloads, fDef.overload)
				}
			}
		}
	}

	ref, err := convert.OverloadMatch(operands, overloads, modelInfo, fmt.Sprintf("%v.%v", libName, defName))
	if err != nil {
		return nil, err
	}
	return &ref, nil
}

// ResolveExactGlobalFunc resolves a reference to a user defined function in an included CQL library
// without any implicit conversions.
func (r *Resolver[T, F]) ResolveExactGlobalFunc(libName string, defName string, operands []types.IType, calledFluently bool, modelInfo *modelinfo.ModelInfos) (F, error) {
	iKey := includeKey{localID: libName, includedBy: r.currLib}
	qKey, ok := r.includedLibs[iKey]
	if !ok {
		return zero[F](), fmt.Errorf("could not resolve the library name %s", libName)
	}

	dKey := defKey{namedLibKey{qualified: qKey.Qualified, version: qKey.Version}, defName}
	var overloads []convert.Overload[F]
	if fDefs, ok := r.funcs[dKey]; ok {
		// Filter overloads that are not public or fluent before calling ExactOverloadMatch.
		for _, fDef := range fDefs {
			if fDef.isPublic {
				if !calledFluently || (calledFluently && fDef.isFluent) {
					overloads = append(overloads, fDef.overload)
				}
			}
		}
	}

	ref, err := convert.ExactOverloadMatch(operands, overloads, modelInfo, fmt.Sprintf("%v.%v", libName, defName))
	if err != nil {
		return zero[F](), err
	}
	return ref, nil
}

// ResolveLocal resolves a reference to a definition in the current CQL library.
func (r *Resolver[T, F]) ResolveLocal(name string) (T, error) {
	dKey := defKey{r.currLib, name}
	if a, ok := r.defs[dKey]; ok {
		return a.result, nil
	}

	aKey := aliasKey{r.currLib, name}
	if a, ok := r.findAlias(aKey); ok {
		return a, nil
	}

	return zero[T](), fmt.Errorf("could not resolve the local reference to %s", name)
}

// ResolveLocalFunc resolves a reference to a user defined or built-in function in the current CQL
// library.
func (r *Resolver[T, F]) ResolveLocalFunc(name string, operands []model.IExpression, calledFluently bool, modelInfo *modelinfo.ModelInfos) (*convert.MatchedOverload[F], error) {
	overloads := make([]convert.Overload[F], 0)
	if overs, ok := r.builtinFuncs[name]; ok {
		overloads = append(overloads, overs...)
	}

	fDefs, ok := r.funcs[defKey{r.currLib, name}]
	if ok {
		// Filter overloads that are fluent before calling OverloadMatch.
		for _, fDef := range fDefs {
			if !calledFluently || (calledFluently && fDef.isFluent) {
				overloads = append(overloads, fDef.overload)
			}
		}
	}

	ref, err := convert.OverloadMatch(operands, overloads, modelInfo, name)
	if err != nil {
		return nil, err
	}
	return &ref, nil
}

// ResolveExactLocalFunc resolves a reference to a user defined function in the current CQL library
// without any implicit conversions.
func (r *Resolver[T, F]) ResolveExactLocalFunc(name string, operands []types.IType, calledFluently bool, modelInfo *modelinfo.ModelInfos) (F, error) {
	overloads := make([]convert.Overload[F], 0)
	if overs, ok := r.builtinFuncs[name]; ok {
		overloads = append(overloads, overs...)
	}

	fDefs, ok := r.funcs[defKey{r.currLib, name}]
	if ok {
		// Filter overloads that are fluent before calling ExactOverloadMatch.
		for _, fDef := range fDefs {
			if !calledFluently || (calledFluently && fDef.isFluent) {
				overloads = append(overloads, fDef.overload)
			}
		}
	}

	ref, err := convert.ExactOverloadMatch(operands, overloads, modelInfo, name)
	if err != nil {
		return zero[F](), err
	}
	return ref, nil
}

// EnterScope starts a new scope for aliases. EndScope should be called to remove all aliases in
// this scope.
func (r *Resolver[T, F]) EnterScope() {
	r.aliases = append(r.aliases, make(map[aliasKey]T))
}

// ExitScope clears any aliases created since the last call to EnterScope.
func (r *Resolver[T, F]) ExitScope() {
	if len(r.aliases) > 0 {
		r.aliases = r.aliases[:len(r.aliases)-1]
	}
}

// Alias creates a new alias within the current scope. When EndScope is called all aliases in the
// scope will be removed. Calling ResolveLocal with the same name will return the stored type t.
// Names must be unique within the CQL library.
func (r *Resolver[T, F]) Alias(name string, a T) error {
	if len(r.aliases) == 0 {
		return errors.New("internal error - EnterScope must be called before creating an alias")
	}
	if err := r.isLocallyUnique(name); err != nil {
		return err
	}
	aKey := aliasKey{r.currLib, name}
	r.aliases[len(r.aliases)-1][aKey] = a
	return nil
}

// PublicDefs returns the public definitions stored in the reference resolver.
func (r *Resolver[T, F]) PublicDefs() (map[result.LibKey]map[string]T, error) {
	pDefs := make(map[result.LibKey]map[string]T)
	for k, v := range r.defs {
		if v.isPublic {
			namedK, ok := k.library.(namedLibKey)
			if !ok {
				return nil, fmt.Errorf("internal error - %v is not a namedLibKey", k.library)
			}
			lKey := result.LibKey{Name: namedK.qualified, Version: namedK.version}
			if _, ok := pDefs[lKey]; !ok {
				pDefs[lKey] = make(map[string]T)
			}
			pDefs[lKey][k.name] = v.result
		}
	}
	return pDefs, nil
}

// PublicAndPrivateDefs should not be used for normal engine execution.
// Returns all public and private definitions, including definitions in unnamed
// libraries. Unnamed libraries are converted to UnnamedLibrary-0 1.0, UnnamedLibrary-1 1.0
// and so on which can clash with any named libraries that happened to be named
// UnnamedLibrary-0. Therefore this should only be used for unit tests and the CQL REPL.
func (r *Resolver[T, F]) PublicAndPrivateDefs() (map[result.LibKey]map[string]T, error) {
	defs := make(map[result.LibKey]map[string]T)
	for k, v := range r.defs {
		var lKey result.LibKey
		switch tk := k.library.(type) {
		case namedLibKey:
			lKey = result.LibKey{Name: tk.qualified, Version: tk.version}
		case unnamedLibKey:
			lKey = result.LibKey{Name: fmt.Sprintf("UnnamedLibrary-%d", tk.unnamedID), Version: "1.0"}
		default:
			return nil, fmt.Errorf("internal error - %v is an unexpected key type", k.library)
		}

		if _, ok := defs[lKey]; !ok {
			defs[lKey] = make(map[string]T)
		}
		defs[lKey][k.name] = v.result
	}
	return defs, nil
}

// isLocallyUnique checks if the name is unique within the current CQL library.
func (r *Resolver[T, F]) isLocallyUnique(name string) error {
	dKey := defKey{r.currLib, name}
	if _, ok := r.defs[dKey]; ok {
		return fmt.Errorf("identifier %v already exists in this CQL library", dKey.name)
	}

	iKey := includeKey{localID: name, includedBy: r.currLib}
	if _, ok := r.includedLibs[iKey]; ok {
		return fmt.Errorf("identifier %v already exists in this CQL library", iKey.localID)
	}

	aKey := aliasKey{r.currLib, name}

	if _, ok := r.findAlias(aKey); ok {
		return fmt.Errorf("alias %v already exists", aKey.name)
	}

	return nil
}

func (r *Resolver[T, F]) isFuncLocallyUnique(name string, operands []types.IType) error {
	if overloads, ok := r.builtinFuncs[name]; ok {
		for _, overload := range overloads {
			if exactMatch(operands, overload.Operands) {
				return fmt.Errorf("built-in CQL function %v(%v) already exists", name, types.ToStrings(operands))
			}
		}
	}

	dKey := defKey{r.currLib, name}
	if overloads, ok := r.funcs[dKey]; ok {
		for _, overload := range overloads {
			if exactMatch(operands, overload.overload.Operands) {
				return fmt.Errorf("function %v(%v) already exists", dKey.name, types.ToStrings(operands))
			}
		}
	}
	return nil
}

func (r *Resolver[T, F]) findAlias(aKey aliasKey) (T, bool) {
	for _, aMap := range r.aliases {
		if t, ok := aMap[aKey]; ok {
			return t, true
		}
	}
	return zero[T](), false
}

func exactMatch(ops1, ops2 []types.IType) bool {
	if len(ops1) != len(ops2) {
		return false
	}
	for i := range ops1 {
		if !ops1[i].Equal(ops2[i]) {
			return false
		}
	}
	return true
}

type libKey interface {
	// libKey is used as the key in maps. Struct that implements it should be comparable.
	isComparableLibKey()
}

func isComparable[_ comparable]() {}

type namedLibKey struct {
	qualified string
	version   string // Empty if no version was specified.
}

func (k namedLibKey) isComparableLibKey() {}

func _[P namedLibKey]() {
	// Enforces at build time that namedLibKey is comparable.
	_ = isComparable[P]
}

// An unnamed library is when there is no library definition in the CQL library ex) `library Test
// version '1'`. All definitions in unnamed libraries are private.
type unnamedLibKey struct {
	unnamedID int
}

func (k unnamedLibKey) isComparableLibKey() {}

func _[P unnamedLibKey]() {
	// Enforces at build time that unnamedLibKey is comparable.
	_ = isComparable[P]
}

type defKey struct {
	library libKey
	name    string
}

type includeKey struct {
	localID string
	// includedBy is the library that is including the library localID.
	includedBy libKey
}

type aliasKey struct {
	library libKey
	name    string
}

// zero is a helper function to return the Zero value of a generic type T.
func zero[T any]() T {
	var zero T
	return zero
}
