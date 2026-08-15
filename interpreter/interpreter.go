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

// Package interpreter interprets and evaluates the data model produced by the CQL parser.
package interpreter

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/cql/internal/modelinfo"
	"github.com/google/cql/internal/reference"
	"github.com/google/cql/model"
	"github.com/google/cql/result"
	"github.com/google/cql/retriever"
	"github.com/google/cql/terminology"
	"github.com/google/cql/types"
)

// Config configures the evaluation of the CQL.
type Config struct {
	DataModels          *modelinfo.ModelInfos
	Parameters          map[result.DefKey]model.IExpression
	Retriever           retriever.Retriever
	Terminology         terminology.Provider
	EvaluationTimestamp time.Time
	ReturnPrivateDefs   bool
}

// Eval evaluates the intermediate ELM like data structure from our parser.
func Eval(ctx context.Context, libs []*model.Library, config Config) (result.Libraries, error) {
	i := &interpreter{
		refs:                reference.NewResolver[result.Value, *model.FunctionDef](),
		terminologyProvider: config.Terminology,
		retriever:           config.Retriever,
		modelInfo:           config.DataModels,
		evaluationTimestamp: config.EvaluationTimestamp,
	}

	for _, lib := range libs {
		if err := i.evalLibrary(lib, config.Parameters); err != nil {
			return nil, result.NewEngineError(result.LibKeyFromModel(lib.Identifier).String(), result.ErrEvaluationError, err)
		}
	}

	if config.ReturnPrivateDefs {
		return i.refs.PublicAndPrivateDefs()
	}
	return i.refs.PublicDefs()
}

// interpreter takes the intermediate ELM like data structure from the parser and executes it.
type interpreter struct {
	refs                *reference.Resolver[result.Value, *model.FunctionDef]
	retriever           retriever.Retriever
	terminologyProvider terminology.Provider
	modelInfo           *modelinfo.ModelInfos
	evaluationTimestamp time.Time
}

// evalLibrary takes a library and evaluates all the expressions that it contains.
func (i *interpreter) evalLibrary(lib *model.Library, passedParams map[result.DefKey]model.IExpression) error {
	for _, using := range lib.Usings {
		if err := i.modelInfo.SetUsing(modelinfo.Key{Name: using.LocalIdentifier, Version: using.Version}); err != nil {
			return err
		}
	}

	if passedParams == nil {
		passedParams = make(map[result.DefKey]model.IExpression)
	}

	if lib.Identifier != nil {
		i.refs.SetCurrentLibrary(lib.Identifier)
	} else {
		i.refs.SetCurrentUnnamed()
	}

	err := i.evalParameters(lib.Parameters, lib.Identifier, passedParams)
	if err != nil {
		return err
	}

	// CodeSystems must be evaluated before ValueSets.
	// TODO b/325631219 - Add a test to validate CodeSystems are evaluated before Valuesets.
	for _, cs := range lib.CodeSystems {
		csObj, err := result.New(result.CodeSystem{ID: cs.ID, Version: cs.Version})
		if err != nil {
			return err
		}
		d := &reference.Def[result.Value]{
			Name:             cs.Name,
			Result:           csObj,
			IsPublic:         cs.AccessLevel == model.Public,
			ValidateIsUnique: false,
		}
		if err := i.refs.Define(d); err != nil {
			return err
		}
	}

	for _, vs := range lib.Valuesets {
		var codeSystems []result.CodeSystem
		for _, cs := range vs.CodeSystems {
			csr, err := i.evalCodeSystemRef(cs)
			if err != nil {
				return err
			}
			csVal, err := result.ToCodeSystem(csr)
			if err != nil {
				return err
			}
			codeSystems = append(codeSystems, csVal)
		}
		vObj, err := result.New(result.ValueSet{ID: vs.ID, Version: vs.Version, CodeSystems: codeSystems})
		if err != nil {
			return err
		}
		d := &reference.Def[result.Value]{
			Name:             vs.Name,
			Result:           vObj,
			IsPublic:         vs.AccessLevel == model.Public,
			ValidateIsUnique: false,
		}
		if err := i.refs.Define(d); err != nil {
			return err
		}
	}
	for _, c := range lib.Codes {
		// TODO: b/326332640 - Investigate whether CodeSystem for codes is optional.
		if c.CodeSystem == nil {
			return fmt.Errorf("The CodeSystem for a Code cannot be null, got code: %v", c)
		}
		cs, err := i.evalCodeSystemRef(c.CodeSystem)
		if err != nil {
			return err
		}

		csVal, err := result.ToCodeSystem(cs)
		if err != nil {
			return err
		}
		cv := result.Code{
			Code:    c.Code,
			System:  csVal.ID,
			Version: csVal.Version,
			Display: c.Display,
		}
		cObj, err := result.New(cv)
		if err != nil {
			return err
		}
		d := &reference.Def[result.Value]{
			Name:             c.Name,
			Result:           cObj,
			IsPublic:         c.AccessLevel == model.Public,
			ValidateIsUnique: false,
		}
		if err := i.refs.Define(d); err != nil {
			return err
		}
	}

	for _, c := range lib.Concepts {
		codes := make([]*result.Code, len(c.Codes))
		for index, code := range c.Codes {
			codeRef, err := i.evalCodeRef(code)
			if err != nil {
				return err
			}
			// Note: the CQL grammar does not allow null codes in a concept using the `concept foo: {}`
			// syntax.
			if result.IsNull(codeRef) {
				codes[index] = nil
				continue
			}
			codeVal, err := result.ToCode(codeRef)
			if err != nil {
				return err
			}
			codes[index] = &codeVal
		}
		cObj, err := result.New(result.Concept{Codes: codes, Display: c.Display})
		if err != nil {
			return err
		}
		d := &reference.Def[result.Value]{
			Name:             c.Name,
			Result:           cObj,
			IsPublic:         c.AccessLevel == model.Public,
			ValidateIsUnique: false,
		}
		if err := i.refs.Define(d); err != nil {
			return err
		}
	}

	for _, inc := range lib.Includes {
		if err := i.refs.IncludeLibrary(inc.Identifier, false); err != nil {
			return err
		}
	}

	if lib.Statements != nil {
		for _, s := range lib.Statements.Defs {
			switch t := s.(type) {
			case *model.ExpressionDef:
				res, err := i.evalExpression(s.GetExpression())
				if err != nil {
					return fmt.Errorf("failed to evaluate expression %s: %w", s.GetName(), err)
				}
				d := &reference.Def[result.Value]{
					Name:             s.GetName(),
					Result:           res,
					IsPublic:         s.GetAccessLevel() == model.Public,
					ValidateIsUnique: false,
				}
				if err = i.refs.Define(d); err != nil {
					return err
				}
			case *model.FunctionDef:
				opTypes := []types.IType{}
				for _, op := range t.Operands {
					opTypes = append(opTypes, op.GetResultType())
				}
				f := &reference.Func[*model.FunctionDef]{
					Name:             s.GetName(),
					Operands:         opTypes,
					Result:           t,
					IsPublic:         s.GetAccessLevel() == model.Public,
					IsFluent:         t.Fluent,
					ValidateIsUnique: false,
				}
				if err = i.refs.DefineFunc(f); err != nil {
					return err
				}
			default:
				return errors.New("internal error - unsupported statement type")
			}
		}
	}

	return nil
}

func (i *interpreter) evalParameters(paramDefs []*model.ParameterDef, id *model.LibraryIdentifier, passedParams map[result.DefKey]model.IExpression) error {
	if id == nil && len(paramDefs) > 0 {
		return fmt.Errorf("unnamed libraries cannot have parameters, got %v", paramDefs[0].Name)
	} else if id == nil {
		return nil
	}

	lKey := result.LibKeyFromModel(id)
	for _, param := range paramDefs {
		var err error
		var pObj result.Value
		pModel, ok := passedParams[result.DefKey{Name: param.Name, Library: lKey}]
		if ok {
			// TODO(b/301606416): We are not supporting arbitrary expressions as passed parameters. We
			// should verify that this is a list, interval or literal.
			pObj, err = i.evalExpression(pModel)
			if err != nil {
				return err
			}
		} else if param.Default != nil {
			// TODO(b/301606416): Parameter defaults should be computable at compile time. We should
			// verify and move this computation to parse time.
			pObj, err = i.evalExpression(param.Default)
			if err != nil {
				return err
			}
		} else {
			// TODO(b/301606416): Send a warning to the user that the param was not provided and is therefore null.
			pObj, err = result.New(nil)
			if err != nil {
				return err
			}
		}
		d := &reference.Def[result.Value]{
			Name:             param.Name,
			Result:           pObj,
			IsPublic:         param.AccessLevel == model.Public,
			ValidateIsUnique: false,
		}
		i.refs.Define(d)
	}
	return nil
}
