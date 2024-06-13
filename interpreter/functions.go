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

package interpreter

import (
	"fmt"

	"github.com/google/cql/model"
	"github.com/google/cql/result"
	"github.com/google/cql/types"
)

func (i *interpreter) evalFunctionRef(f *model.FunctionRef) (result.Value, error) {
	// Evaluate the operands
	ops := []result.Value{}
	opTypes := []types.IType{}
	for _, exp := range f.Operands {
		op, err := i.evalExpression(exp)
		if err != nil {
			return result.Value{}, err
		}
		ops = append(ops, op)
		opTypes = append(opTypes, exp.GetResultType())
	}

	// Resolve the function
	var resolved *model.FunctionDef
	var err error
	if f.LibraryName != "" {
		resolved, err = i.refs.ResolveExactGlobalFunc(f.LibraryName, f.Name, opTypes, false, i.modelInfo)
	} else {
		resolved, err = i.refs.ResolveExactLocalFunc(f.Name, opTypes, false, i.modelInfo)
	}
	if err != nil {
		return result.Value{}, err
	}

	// Evaluate the function
	if resolved.External {
		return result.Value{}, fmt.Errorf("function %v is external, but external functions are not supported", f.Name)
	}
	i.refs.EnterScope()
	defer i.refs.ExitScope()
	for j, op := range ops {
		if err := i.refs.Alias(resolved.Operands[j].Name, op); err != nil {
			return result.Value{}, err
		}
	}
	// TODO(b/301606416): Verify that the type of the result of the evaluated function matches the
	// return type in model.FunctionDef.
	r, err := i.evalExpression(resolved.Expression)
	if err != nil {
		return result.Value{}, err
	}
	// TODO(b/311222838): This currently add only the function expression to the resulting expression,
	// since function parameters would be attached as operands in sub-expressions. We should
	// determine whether this is sufficiently for real explainability workloads.
	return r.WithSources(f), nil
}
