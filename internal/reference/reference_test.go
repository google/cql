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

package reference

import (
	"strings"
	"testing"

	"github.com/google/cql/internal/embeddata"
	"github.com/google/cql/internal/modelinfo"
	"github.com/google/cql/model"
	"github.com/google/cql/result"
	"github.com/google/cql/types"
	"github.com/google/go-cmp/cmp"
)

func TestParserDefAndResolve(t *testing.T) {
	tests := []struct {
		name          string
		want          model.IExpression
		resolverCalls func(*Resolver[model.IExpression, model.IExpression]) model.IExpression
	}{
		{
			name: "Resolve Local Def",
			want: &model.ExpressionRef{Name: "public def"},
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) model.IExpression {
				// This also tests that two expressions named "public def" in two different libraries don't
				// clash.
				d := &Def[model.IExpression]{
					Name:             "public def",
					Result:           &model.ExpressionRef{Name: "public def"},
					IsPublic:         true,
					ValidateIsUnique: true,
				}
				if err := r.Define(d); err != nil {
					t.Errorf("Define(public def) unexpected err: %v", err)
				}
				got, err := r.ResolveLocal("public def")
				if err != nil {
					t.Errorf("ResolveLocal(public def) unexpected err: %v", err)
				}
				return got
			},
		},
		{
			name: "Resolve Built-in Func",
			want: &model.Last{},
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) model.IExpression {
				ops := []model.IExpression{model.NewList([]string{"4"}, types.Integer)}
				opTypes := []types.IType{&types.List{ElementType: types.Any}}
				if err := r.DefineBuiltinFunc("builtin func", opTypes, &model.Last{}); err != nil {
					t.Errorf("DefineBuiltinFunc(builtin func) unexpected err: %v", err)
				}
				got, err := r.ResolveLocalFunc("builtin func", ops, false, newFHIRModelInfo(t))
				if err != nil {
					t.Errorf("ResolveLocalFunc(builtin func) unexpected err: %v", err)
				}
				wantWrappedOperands := []model.IExpression{model.NewList([]string{"4"}, types.Integer)}
				if diff := cmp.Diff(wantWrappedOperands, got.WrappedOperands); diff != "" {
					t.Errorf("ResolveLocalFunc() diff (-want +got):\n%s", diff)
				}
				return got.Result
			},
		},
		{
			name: "Resolve Local Func",
			want: &model.FunctionRef{Name: "public func", Operands: []model.IExpression{}},
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) model.IExpression {
				// This also tests that two funcs named "public func" in two different libraries don't
				// clash.
				f := &Func[model.IExpression]{
					Name:             "public func",
					Operands:         []types.IType{types.Integer},
					Result:           &model.FunctionRef{Name: "public func", Operands: []model.IExpression{}},
					IsPublic:         true,
					IsFluent:         false,
					ValidateIsUnique: true,
				}
				if err := r.DefineFunc(f); err != nil {
					t.Errorf("DefineFunc(public func) unexpected err: %v", err)
				}
				got, err := r.ResolveLocalFunc("public func", []model.IExpression{model.NewLiteral("4", types.Integer)}, false, newFHIRModelInfo(t))
				if err != nil {
					t.Errorf("ResolveLocalFunc(public func) unexpected err: %v", err)
				}
				wantWrappedOperands := []model.IExpression{model.NewLiteral("4", types.Integer)}
				if diff := cmp.Diff(wantWrappedOperands, got.WrappedOperands); diff != "" {
					t.Errorf("ResolveLocalFunc() diff (-want +got):\n%s", diff)
				}
				return got.Result
			},
		},
		{
			name: "Resolve Local Func Based on Operands",
			want: &model.FunctionRef{Name: "public func", Operands: []model.IExpression{}},
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) model.IExpression {
				f := &Func[model.IExpression]{
					Name:             "public func",
					Operands:         []types.IType{},
					Result:           &model.FunctionRef{Name: "public func", Operands: []model.IExpression{}},
					IsPublic:         true,
					IsFluent:         false,
					ValidateIsUnique: true,
				}
				if err := r.DefineFunc(f); err != nil {
					t.Errorf("DefineFunc(public func) unexpected err: %v", err)
				}

				fSameName := &Func[model.IExpression]{
					Name:             "public func",
					Operands:         []types.IType{types.DateTime},
					Result:           &model.FunctionRef{Name: "public func", Operands: []model.IExpression{}},
					IsPublic:         true,
					IsFluent:         false,
					ValidateIsUnique: true,
				}
				if err := r.DefineFunc(fSameName); err != nil {
					t.Errorf("DefineFunc(public func) unexpected err: %v", err)
				}

				got, err := r.ResolveLocalFunc("public func", []model.IExpression{model.NewLiteral("2023-06-30", types.Date)}, false, newFHIRModelInfo(t))
				if err != nil {
					t.Errorf("ResolveLocalFunc(public func) unexpected err: %v", err)
				}

				// This also tests that the operand is wrapped in model.ToDateTime.
				wantWrappedOperands := []model.IExpression{
					&model.ToDateTime{
						UnaryExpression: &model.UnaryExpression{
							Operand:    model.NewLiteral("2023-06-30", types.Date),
							Expression: model.ResultType(types.DateTime),
						},
					},
				}
				if diff := cmp.Diff(wantWrappedOperands, got.WrappedOperands); diff != "" {
					t.Errorf("ResolveLocalFunc() diff (-want +got):\n%s", diff)
				}
				return got.Result
			},
		},
		{
			name: "Resolve Fluent Local Func",
			want: &model.FunctionRef{Name: "is fluent", Operands: []model.IExpression{}},
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) model.IExpression {
				f := &Func[model.IExpression]{
					Name:             "fluent func",
					Operands:         []types.IType{types.DateTime},
					Result:           &model.FunctionRef{Name: "is fluent", Operands: []model.IExpression{}},
					IsPublic:         true,
					IsFluent:         true,
					ValidateIsUnique: true,
				}
				if err := r.DefineFunc(f); err != nil {
					t.Errorf("DefineFunc(fluent func) unexpected err: %v", err)
				}

				// We don't match this one even though it is an exact match because it is not a fluent function.
				fNotFluent := &Func[model.IExpression]{
					Name:             "fluent func",
					Operands:         []types.IType{types.Date},
					Result:           &model.FunctionRef{Name: "not fluent", Operands: []model.IExpression{}},
					IsPublic:         true,
					IsFluent:         false,
					ValidateIsUnique: true,
				}
				if err := r.DefineFunc(fNotFluent); err != nil {
					t.Errorf("DefineFunc(fluent func) unexpected err: %v", err)
				}

				got, err := r.ResolveLocalFunc("fluent func", []model.IExpression{model.NewLiteral("2023-06-30", types.Date)}, true, newFHIRModelInfo(t))
				if err != nil {
					t.Errorf("ResolveLocalFunc(public func) unexpected err: %v", err)
				}
				return got.Result
			},
		},
		{
			name: "Resolve Local Built-in Func based on Operands",
			want: &model.Last{},
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) model.IExpression {
				ops1 := []model.IExpression{model.NewLiteral("4", types.Integer)}
				opTypes1 := []types.IType{types.Integer}
				if err := r.DefineBuiltinFunc("builtin func", opTypes1, &model.Last{}); err != nil {
					t.Errorf("DefineBuiltinFunc(builtin func) unexpected err: %v", err)
				}

				f := &Func[model.IExpression]{
					Name:             "builtin func",
					Operands:         []types.IType{types.String},
					Result:           &model.FunctionRef{},
					IsPublic:         true,
					IsFluent:         false,
					ValidateIsUnique: true,
				}
				if err := r.DefineFunc(f); err != nil {
					t.Errorf("DefineFunc(builtin func) unexpected err: %v", err)
				}

				got, err := r.ResolveLocalFunc("builtin func", ops1, false, newFHIRModelInfo(t))
				if err != nil {
					t.Errorf("ResolveLocalFunc(builtin func) unexpected err: %v", err)
				}
				wantWrappedOperands := []model.IExpression{model.NewLiteral("4", types.Integer)}
				if diff := cmp.Diff(wantWrappedOperands, got.WrappedOperands); diff != "" {
					t.Errorf("ResolveLocalFunc(builtin func) diff (-want +got):\n%s", diff)
				}
				return got.Result
			},
		},
		{
			name: "Resolve Exact Built-in Func",
			want: &model.Last{},
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) model.IExpression {
				opTypes := []types.IType{&types.List{ElementType: types.Any}}
				if err := r.DefineBuiltinFunc("builtin func", opTypes, &model.Last{}); err != nil {
					t.Errorf("DefineBuiltinFunc(builtin func) unexpected err: %v", err)
				}
				got, err := r.ResolveExactLocalFunc("builtin func", opTypes, false, newFHIRModelInfo(t))
				if err != nil {
					t.Errorf("ResolveExactLocalFunc(builtin func) unexpected err: %v", err)
				}
				return got
			},
		},
		{
			name: "Resolve Exact Local Func",
			want: &model.FunctionRef{Name: "public func", Operands: []model.IExpression{}},
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) model.IExpression {
				// This also tests that two funcs named "public func" in two different libraries don't
				// clash.
				f := &Func[model.IExpression]{
					Name:             "public func",
					Operands:         []types.IType{types.Integer},
					Result:           &model.FunctionRef{Name: "public func", Operands: []model.IExpression{}},
					IsPublic:         true,
					IsFluent:         false,
					ValidateIsUnique: true,
				}
				if err := r.DefineFunc(f); err != nil {
					t.Errorf("DefineFunc(public func) unexpected err: %v", err)
				}
				got, err := r.ResolveExactLocalFunc("public func", []types.IType{types.Integer}, false, newFHIRModelInfo(t))
				if err != nil {
					t.Errorf("ResolveExactLocalFunc(public func) unexpected err: %v", err)
				}
				return got
			},
		},
		{
			name: "Resolve Exact Fluent Local Func by Subtype",
			want: &model.FunctionRef{Name: "is fluent", Operands: []model.IExpression{}},
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) model.IExpression {
				f := &Func[model.IExpression]{
					Name:             "fluent func",
					Operands:         []types.IType{types.Vocabulary},
					Result:           &model.FunctionRef{Name: "is fluent", Operands: []model.IExpression{}},
					IsPublic:         true,
					IsFluent:         true,
					ValidateIsUnique: true,
				}
				if err := r.DefineFunc(f); err != nil {
					t.Errorf("DefineFunc(fluent func) unexpected err: %v", err)
				}

				// We don't match this one even though it is an exact match because it is not a fluent function.
				fNotFluent := &Func[model.IExpression]{
					Name:             "fluent func",
					Operands:         []types.IType{types.ValueSet},
					Result:           &model.FunctionRef{Name: "not fluent", Operands: []model.IExpression{}},
					IsPublic:         true,
					IsFluent:         false,
					ValidateIsUnique: true,
				}
				if err := r.DefineFunc(fNotFluent); err != nil {
					t.Errorf("DefineFunc(fluent func) unexpected err: %v", err)
				}

				got, err := r.ResolveExactLocalFunc("fluent func", []types.IType{types.ValueSet}, true, newFHIRModelInfo(t))
				if err != nil {
					t.Errorf("ResolveExactLocalFunc(public func) unexpected err: %v", err)
				}
				return got
			},
		},
		{
			name: "Resolve Global Def",
			want: &model.ExpressionRef{Name: "public def"},
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) model.IExpression {
				got, err := r.ResolveGlobal("helpers", "public def")
				if err != nil {
					t.Errorf("ResolveGlobal(helpers, public def) unexpected err: %v", err)
				}
				return got
			},
		},
		{
			name: "Resolve Global Func",
			want: &model.FunctionRef{Name: "public func", Operands: []model.IExpression{}},
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) model.IExpression {
				ops := []model.IExpression{model.NewLiteral("Apple", types.String)}
				got, err := r.ResolveGlobalFunc("helpers", "public func", ops, false, newFHIRModelInfo(t))
				if err != nil {
					t.Errorf("ResolveGlobalFunc() unexpected err: %v", err)
				}
				wantWrappedOperands := []model.IExpression{model.NewLiteral("Apple", types.String)}
				if diff := cmp.Diff(wantWrappedOperands, got.WrappedOperands); diff != "" {
					t.Errorf("ResolveGlobalFunc() diff (-want +got):\n%s", diff)
				}
				return got.Result
			},
		},
		{
			name: "Resolve Global Fluent Func",
			want: &model.FunctionRef{Name: "public fluent func", Operands: []model.IExpression{}},
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) model.IExpression {
				ops := []model.IExpression{model.NewLiteral("Apple", types.String)}
				got, err := r.ResolveGlobalFunc("helpers", "public fluent func", ops, true, newFHIRModelInfo(t))
				if err != nil {
					t.Errorf("ResolveGlobalFunc() unexpected err: %v", err)
				}
				wantWrappedOperands := []model.IExpression{model.NewLiteral("Apple", types.String)}
				if diff := cmp.Diff(wantWrappedOperands, got.WrappedOperands); diff != "" {
					t.Errorf("ResolveGlobalFunc() diff (-want +got):\n%s", diff)
				}
				return got.Result
			},
		},
		{
			name: "Resolve Exact Global Func",
			want: &model.FunctionRef{Name: "public func", Operands: []model.IExpression{}},
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) model.IExpression {
				opTypes := []types.IType{types.String}
				got, err := r.ResolveExactGlobalFunc("helpers", "public func", opTypes, false, newFHIRModelInfo(t))
				if err != nil {
					t.Errorf("ResolveExactGlobalFunc() unexpected err: %v", err)
				}
				return got
			},
		},
		{
			name: "Resolve Global Fluent Func by Subtype",
			want: &model.FunctionRef{Name: "public fluent func", Operands: []model.IExpression{}},
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) model.IExpression {
				opTypes := []types.IType{types.String}
				got, err := r.ResolveExactGlobalFunc("helpers", "public fluent func", opTypes, true, newFHIRModelInfo(t))
				if err != nil {
					t.Errorf("ResolveExactGlobalFunc() unexpected err: %v", err)
				}
				return got
			},
		},
		{
			name: "Unnamed Def",
			want: &model.ExpressionRef{Name: "public def"},
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) model.IExpression {
				r.SetCurrentUnnamed()
				// Public definitions are stored as private definitions and can still be resolved locally.
				d := &Def[model.IExpression]{
					Name:             "public def",
					Result:           &model.ExpressionRef{Name: "public def"},
					IsPublic:         true,
					ValidateIsUnique: true,
				}
				if err := r.Define(d); err != nil {
					t.Errorf("Define(public def) unexpected err: %v", err)
				}
				got, err := r.ResolveLocal("public def")
				if err != nil {
					t.Errorf("ResolveLocal(public def) unexpected err: %v", err)
				}
				return got
			},
		},
		{
			name: "Unnamed Func",
			want: &model.FunctionRef{Name: "public func", Operands: []model.IExpression{}},
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) model.IExpression {
				r.SetCurrentUnnamed()
				// Public definitions are stored as private definitions and can still be resolved locally.
				f := &Func[model.IExpression]{
					Name:             "public func",
					Operands:         []types.IType{types.Integer},
					Result:           &model.FunctionRef{Name: "public func", Operands: []model.IExpression{}},
					IsPublic:         true,
					IsFluent:         true,
					ValidateIsUnique: true,
				}
				if err := r.DefineFunc(f); err != nil {
					t.Errorf("DefineFunc(public func) unexpected err: %v", err)
				}
				got, err := r.ResolveLocalFunc("public func", []model.IExpression{model.NewLiteral("4", types.Integer)}, false, newFHIRModelInfo(t))
				if err != nil {
					t.Errorf("ResolveLocalFunc(public func) unexpected err: %v", err)
				}
				wantWrappedOperands := []model.IExpression{model.NewLiteral("4", types.Integer)}
				if diff := cmp.Diff(wantWrappedOperands, got.WrappedOperands); diff != "" {
					t.Errorf("ResolveLocalFunc(public func) diff (-want +got):\n%s", diff)
				}
				return got.Result
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// TEST SETUP - PREVIOUS PARSED LIBRARY
			//
			// library example.helpers version '1.0'
			// define public "public def" : ...
			// define private "private def": ...
			// define public function "public func"(a String):...
			// define public fluent function "public fluent func"(a String):...
			// define private function "private func"(b String):...
			r := NewResolver[model.IExpression, model.IExpression]()
			if err := r.SetCurrentLibrary(&model.LibraryIdentifier{
				Local:     "helpers",
				Qualified: "example.helpers",
				Version:   "1.0",
			}); err != nil {
				t.Fatalf("r.SetCurrentLibrary() unexpected err: %v", err)
			}
			buildLibrary(t, r)

			// TEST SETUP - CURRENT LIBRARY
			//
			// library example.measure version '1.0'
			// include example.helpers version '1.0'
			if err := r.SetCurrentLibrary(&model.LibraryIdentifier{
				Local:     "measure",
				Qualified: "example.measure",
				Version:   "1.0",
			}); err != nil {
				t.Fatalf("r.SetCurrentLibrary() unexpected err: %v", err)
			}
			if err := r.IncludeLibrary(&model.LibraryIdentifier{Local: "helpers", Qualified: "example.helpers", Version: "1.0"}, true); err != nil {
				t.Fatalf("r.IncludeLibrary() unexpected err: %v", err)
			}

			got := tc.resolverCalls(r)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("resolverCalls() diff (-want +got):\n%v", diff)
			}
		})
	}
}

func TestParserAliasAndResolve(t *testing.T) {
	// This tests the Aliases, especially that they are appropriately deleted when a scope is exited.
	r := NewResolver[model.IExpression, model.IExpression]()
	if err := r.SetCurrentLibrary(&model.LibraryIdentifier{
		Local:     "measure",
		Qualified: "example.measure",
		Version:   "1.0",
	}); err != nil {
		t.Fatalf("r.SetCurrentLibrary() unexpected err: %v", err)
	}

	// Create Alias P.
	r.EnterScope()
	if err := r.Alias("P", &model.AliasRef{Name: "P"}); err != nil {
		t.Fatalf("Alias(P) unexpected err: %v", err)
	}

	want := &model.AliasRef{Name: "P"}
	got, err := r.ResolveLocal("P")
	if err != nil {
		t.Fatalf("ResolveLocal(P) unexpected err: %v", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ResolveLocal(P) diff (-want +got):\n%v", diff)
	}

	// Create nested Alias O.
	r.EnterScope()
	if err := r.Alias("O", &model.AliasRef{Name: "O"}); err != nil {
		t.Fatalf("Alias(O) unexpected err: %v", err)
	}
	// O is in the inner scope.
	want = &model.AliasRef{Name: "O"}
	got, err = r.ResolveLocal("O")
	if err != nil {
		t.Fatalf("ResolveLocal(O) unexpected err: %v", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ResolveLocal(O) diff (-want +got):\n%v", diff)
	}

	// P still exists in the outer scope.
	want = &model.AliasRef{Name: "P"}
	got, err = r.ResolveLocal("P")
	if err != nil {
		t.Fatalf("ResolveLocal(P) unexpected err: %v", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ResolveLocal(P) diff (-want +got):\n%v", diff)
	}

	// Inner Alias Scope Cleared.
	r.ExitScope()
	// O no longer exists.
	wantError := "could not resolve the local reference"
	_, gotError := r.ResolveLocal("O")
	if gotError == nil {
		t.Fatalf("ResolveLocal(O) expected error got success")
	}
	if !strings.Contains(gotError.Error(), wantError) {
		t.Errorf("Returned error (%s) did not contain (%s)", gotError.Error(), wantError)
	}

	// P still exists in the outer scope.
	want = &model.AliasRef{Name: "P"}
	got, gotError = r.ResolveLocal("P")
	if gotError != nil {
		t.Fatalf("ResolveLocal(P) unexpected err: %v", gotError)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ResolveLocal(P) diff (-want +got):\n%v", diff)
	}

	// Outer Alias Scope Cleared.
	r.ExitScope()
	// P no longer exists.
	wantError = "could not resolve the local reference"
	_, gotError = r.ResolveLocal("P")
	if gotError == nil {
		t.Fatalf("ResolveLocal(P) expected error got success")
	}
	if !strings.Contains(gotError.Error(), wantError) {
		t.Errorf("Returned error (%s) did not contain (%s)", gotError.Error(), wantError)
	}
}

func TestResolveIncludedLibrary(t *testing.T) {
	// TEST SETUP - PREVIOUS PARSED LIBRARY
	//
	// library example.helpers version '1.0'
	r := NewResolver[model.IExpression, model.IExpression]()
	if err := r.SetCurrentLibrary(&model.LibraryIdentifier{
		Local:     "helpers",
		Qualified: "example.helpers",
		Version:   "1.0",
	}); err != nil {
		t.Fatalf("r.SetCurrentLibrary() unexpected err: %v", err)
	}

	// TEST SETUP - CURRENT LIBRARY
	//
	// library example.measure version '1.0'
	// include example.helpers version '1.0'
	if err := r.SetCurrentLibrary(&model.LibraryIdentifier{
		Local:     "measure",
		Qualified: "example.measure",
		Version:   "1.0",
	}); err != nil {
		t.Fatalf("r.SetCurrentLibrary() unexpected err: %v", err)
	}
	if err := r.IncludeLibrary(&model.LibraryIdentifier{Local: "helpers", Qualified: "example.helpers", Version: "1.0"}, true); err != nil {
		t.Fatalf("r.IncludeLibrary() unexpected err: %v", err)
	}

	got := r.ResolveInclude("helpers")

	want := model.LibraryIdentifier{Local: "helpers", Qualified: "example.helpers", Version: "1.0"}
	if *got != want {
		t.Errorf("ResolveInclude(helpers) got %v, want %v", got, want)
	}

}

func TestResolverErrors(t *testing.T) {
	tests := []struct {
		name          string
		resolverCalls func(*Resolver[model.IExpression, model.IExpression]) error
		errContains   string
	}{
		{
			name:        "SetCurrentLibrary Same Name",
			errContains: "already exists",
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) error {
				i := &model.LibraryIdentifier{Local: "helpers", Qualified: "example.helpers", Version: "1.0"}
				return r.SetCurrentLibrary(i)
			},
		},
		{
			name:        "IncludeLibrary Same Name",
			errContains: "already exists",
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) error {
				i := &model.LibraryIdentifier{Local: "helpers", Qualified: "example.helpers", Version: "1.0"}
				return r.IncludeLibrary(i, true)
			},
		},
		{
			name:        "Include and Def Same Name",
			errContains: "already exists",
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) error {
				i := &model.LibraryIdentifier{Local: "public def", Qualified: "example.helpers", Version: "1.0"}
				return r.IncludeLibrary(i, true)
			},
		},
		{
			name:        "Alias and Def Same Name",
			errContains: "already exists",
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) error {
				r.EnterScope()
				defer r.ExitScope()
				return r.Alias("private def", &model.AliasRef{Name: "private def"})
			},
		},
		{
			name:        "Alias Created Without EnterScope",
			errContains: "EnterScope must be called",
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) error {
				return r.Alias("A", &model.AliasRef{Name: "A"})
			},
		},
		{
			name:        "IncludeLibrary Nonexistent Name",
			errContains: "not exist",
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) error {
				i := &model.LibraryIdentifier{Local: "nonexistent", Qualified: "example.nonexistent", Version: "1.0"}
				return r.IncludeLibrary(i, true)
			},
		},
		{
			name:        "Public Def Same Name",
			errContains: "already exist",
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) error {
				d := &Def[model.IExpression]{
					Name:             "public def",
					Result:           &model.ExpressionRef{Name: "public def"},
					IsPublic:         true,
					ValidateIsUnique: true,
				}
				return r.Define(d)
			},
		},
		{
			name:        "Public Func Same Name",
			errContains: "already exist",
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) error {
				f := &Func[model.IExpression]{
					Name:             "public func",
					Operands:         []types.IType{types.String},
					Result:           &model.FunctionRef{Name: "public func", Operands: []model.IExpression{}},
					IsPublic:         false,
					IsFluent:         false,
					ValidateIsUnique: true,
				}
				return r.DefineFunc(f)
			},
		},
		{
			name:        "Private Def Same Name",
			errContains: "already exist",
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) error {
				d := &Def[model.IExpression]{
					Name:             "private def",
					Result:           &model.ExpressionRef{Name: "private def"},
					IsPublic:         false,
					ValidateIsUnique: true,
				}
				return r.Define(d)
			},
		},
		{
			name:        "Private Func Same Name",
			errContains: "already exist",
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) error {
				f := &Func[model.IExpression]{
					Name:             "private func",
					Operands:         []types.IType{types.String},
					Result:           &model.FunctionRef{Name: "private func", Operands: []model.IExpression{}},
					IsPublic:         false,
					IsFluent:         false,
					ValidateIsUnique: true,
				}
				return r.DefineFunc(f)
			},
		},
		{
			name:        "Public Func Same Name As Built-in",
			errContains: "already exist",
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) error {
				ops := []types.IType{types.String}
				if err := r.DefineBuiltinFunc("public func", ops, &model.Last{}); err != nil {
					t.Errorf("DefineBuiltinFunc(public func) unexpected err: %v", err)
				}

				f := &Func[model.IExpression]{
					Name:             "public func",
					Operands:         ops,
					Result:           &model.FunctionRef{},
					IsPublic:         false,
					IsFluent:         false,
					ValidateIsUnique: true,
				}
				return r.DefineFunc(f)
			},
		},
		{
			name:        "ResolveLocal Nonexistent",
			errContains: "could not resolve",
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) error {
				_, err := r.ResolveLocal("nonexistent def")
				return err
			},
		},
		{
			name:        "ResolveLocalFunc Nonexistent",
			errContains: "could not resolve",
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) error {
				_, err := r.ResolveLocalFunc("nonexistent func", []model.IExpression{}, false, newFHIRModelInfo(t))
				return err
			},
		},
		{
			name:        "ResolveLocalFunc Nonexistent Operands",
			errContains: "could not resolve",
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) error {
				_, err := r.ResolveLocalFunc("public func", []model.IExpression{}, false, newFHIRModelInfo(t))
				return err
			},
		},
		{
			name:        "ResolveLocalFunc Not Fluent",
			errContains: "could not resolve",
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) error {
				_, err := r.ResolveLocalFunc("public func", []model.IExpression{model.NewLiteral("Apple", types.String)}, true, newFHIRModelInfo(t))
				return err
			},
		},
		{
			name:        "ResolveExactLocalFunc Nonexistent",
			errContains: "could not resolve",
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) error {
				_, err := r.ResolveExactLocalFunc("nonexistent func", []types.IType{}, false, newFHIRModelInfo(t))
				return err
			},
		},
		{
			name:        "ResolveExactLocalFunc Nonexistent Operands",
			errContains: "could not resolve",
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) error {
				_, err := r.ResolveExactLocalFunc("public func", []types.IType{}, false, newFHIRModelInfo(t))
				return err
			},
		},
		{
			name:        "ResolveExactLocalFunc Not Fluent",
			errContains: "could not resolve",
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) error {
				_, err := r.ResolveExactLocalFunc("public func", []types.IType{types.String}, true, newFHIRModelInfo(t))
				return err
			},
		},
		{
			name:        "ResolveGlobal Nonexistent Def",
			errContains: "resolve the reference",
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) error {
				_, err := r.ResolveGlobal("helpers", "nonexistent def")
				return err
			},
		},
		{
			name:        "ResolveGlobalFunc Nonexistent Def",
			errContains: "could not resolve",
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) error {
				_, err := r.ResolveGlobalFunc("helpers", "nonexistent def", []model.IExpression{}, false, newFHIRModelInfo(t))
				return err
			},
		},
		{
			name:        "ResolveGlobal Nonexistent Library",
			errContains: "resolve the library",
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) error {
				_, err := r.ResolveGlobal("nonexistent lib", "public def")
				return err
			},
		},
		{
			name:        "ResolveGlobalFunc Nonexistent Library",
			errContains: "resolve the library",
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) error {
				_, err := r.ResolveGlobalFunc("nonexistent lib", "public func", []model.IExpression{}, false, newFHIRModelInfo(t))
				return err
			},
		},
		{
			name:        "ResolveGlobalFunc Nonexistent Operands",
			errContains: "could not resolve",
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) error {
				_, err := r.ResolveGlobalFunc("helpers", "public func", []model.IExpression{}, false, newFHIRModelInfo(t))
				return err
			},
		},
		{
			name:        "ResolveGlobalFunc Not Fluent",
			errContains: "could not resolve",
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) error {
				_, err := r.ResolveGlobalFunc("helpers", "public func", []model.IExpression{model.NewLiteral("Apple", types.String)}, true, newFHIRModelInfo(t))
				return err
			},
		},
		{
			name:        "ResolveExactGlobalFunc Nonexistent Library",
			errContains: "resolve the library",
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) error {
				_, err := r.ResolveExactGlobalFunc("nonexistent lib", "public func", []types.IType{}, false, newFHIRModelInfo(t))
				return err
			},
		},
		{
			name:        "ResolveExactGlobalFunc Nonexistent Operands",
			errContains: "could not resolve",
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) error {
				_, err := r.ResolveExactGlobalFunc("helpers", "public func", []types.IType{}, false, newFHIRModelInfo(t))
				return err
			},
		},
		{
			name:        "ResolveExactGlobalFunc Not Fluent",
			errContains: "could not resolve",
			resolverCalls: func(r *Resolver[model.IExpression, model.IExpression]) error {
				_, err := r.ResolveExactGlobalFunc("helpers", "public func", []types.IType{types.String}, true, newFHIRModelInfo(t))
				return err
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// TEST SETUP - PREVIOUS PARSED LIBRARY
			//
			// library example.helpers version '1.0'
			// define public "public def" : ...
			// define private "private def": ...
			// define public function "public func"(a String):...
			// define private function "private func"(b String):...

			r := NewResolver[model.IExpression, model.IExpression]()
			if err := r.SetCurrentLibrary(&model.LibraryIdentifier{
				Local:     "helpers",
				Qualified: "example.helpers",
				Version:   "1.0",
			}); err != nil {
				t.Fatalf("r.SetCurrentLibrary() unexpected err: %v", err)
			}
			buildLibrary(t, r)

			// TEST SETUP - CURRENT LIBRARY
			//
			// library example.measure version '1.0'
			// include example.helpers version '1.0'
			// define public "public def" : ...
			// define private "private def": ...
			// define public function "public func"(a String):...
			// define private function "private func"(b String):...

			if err := r.SetCurrentLibrary(&model.LibraryIdentifier{
				Local:     "measure",
				Qualified: "example.measure",
				Version:   "1.0",
			}); err != nil {
				t.Fatalf("r.SetCurrentLibrary() unexpected err: %v", err)
			}
			if err := r.IncludeLibrary(&model.LibraryIdentifier{Local: "helpers", Qualified: "example.helpers", Version: "1.0"}, true); err != nil {
				t.Fatalf("r.IncludeLibrary() unexpected err: %v", err)
			}
			buildLibrary(t, r)

			gotErr := tc.resolverCalls(r)
			if gotErr == nil {
				t.Fatalf("Resolver did not return an error")
			}
			if !strings.Contains(gotErr.Error(), tc.errContains) {
				t.Errorf("Returned error (%s) did not contain (%s)", gotErr.Error(), tc.errContains)
			}
		})
	}
}

func TestDefs(t *testing.T) {
	// TEST SETUP - PREVIOUS PARSED LIBRARY
	//
	// define public "public def" : ...
	// define private "private def": ...
	r := NewResolver[result.Value, *model.FunctionDef]()
	r.SetCurrentUnnamed()
	dUnnamed := &Def[result.Value]{
		Name:             "public def",
		Result:           newOrFatal(4, t),
		IsPublic:         true,
		ValidateIsUnique: true,
	}
	if err := r.Define(dUnnamed); err != nil {
		t.Fatalf("r.Define(public def) unexpected err: %v", err)
	}

	dUnnamedPrivate := &Def[result.Value]{
		Name:             "private def",
		Result:           newOrFatal(5, t),
		IsPublic:         false,
		ValidateIsUnique: true,
	}
	if err := r.Define(dUnnamedPrivate); err != nil {
		t.Fatalf("r.Define(private def) unexpected err: %v", err)
	}

	// TEST SETUP - CURRENT LIBRARY
	//
	// library example.measure version '1.0'
	// define public "public def": ...
	// define private "private def": ...
	if err := r.SetCurrentLibrary(&model.LibraryIdentifier{
		Local:     "measure",
		Qualified: "example.measure",
		Version:   "1.0",
	}); err != nil {
		t.Fatalf("r.SetCurrentLibrary() unexpected err: %v", err)
	}
	dPublic := &Def[result.Value]{
		Name:             "public def",
		Result:           newOrFatal(4, t),
		IsPublic:         true,
		ValidateIsUnique: true,
	}
	if err := r.Define(dPublic); err != nil {
		t.Fatalf("r.Define(public def) unexpected err: %v", err)
	}
	dPrivate := &Def[result.Value]{
		Name:             "private def",
		Result:           newOrFatal(5, t),
		IsPublic:         false,
		ValidateIsUnique: true,
	}
	if err := r.Define(dPrivate); err != nil {
		t.Fatalf("r.Define(private def) unexpected err: %v", err)
	}

	t.Run("Public Defs", func(t *testing.T) {
		got, err := r.PublicDefs()
		if err != nil {
			t.Fatalf("r.PublicDefs() unexpected err: %v", err)
		}
		want := map[result.LibKey]map[string]result.Value{
			result.LibKey{Name: "example.measure", Version: "1.0"}: map[string]result.Value{"public def": newOrFatal(4, t)},
		}
		if diff := cmp.Diff(got, want); diff != "" {
			t.Errorf("r.PublicDefs() returned unexpected diff (-got +want):\n%s", diff)
		}
	})

	t.Run("All Defs", func(t *testing.T) {
		got, err := r.PublicAndPrivateDefs()
		if err != nil {
			t.Fatalf("r.Defs() unexpected err: %v", err)
		}
		want := map[result.LibKey]map[string]result.Value{
			result.LibKey{Name: "example.measure", Version: "1.0"}: map[string]result.Value{
				"public def":  newOrFatal(4, t),
				"private def": newOrFatal(5, t),
			},
			result.LibKey{Name: "UnnamedLibrary-0", Version: "1.0"}: map[string]result.Value{
				"public def":  newOrFatal(4, t),
				"private def": newOrFatal(5, t),
			},
		}
		if diff := cmp.Diff(got, want); diff != "" {
			t.Errorf("r.Defs() returned unexpected diff (-got +want):\n%s", diff)
		}
	})
}

func newOrFatal(a any, t *testing.T) result.Value {
	o, err := result.New(a)
	if err != nil {
		t.Fatalf("New(%v) returned unexpected error: %v", a, err)
	}
	return o
}

func buildLibrary(t *testing.T, r *Resolver[model.IExpression, model.IExpression]) {
	// define public "public def" : ...
	// define private "private def": ...
	// define public function "public func"(a String):...
	// define private function "private func"(b String):...

	t.Helper()
	dPublic := &Def[model.IExpression]{
		Name:             "public def",
		Result:           &model.ExpressionRef{Name: "public def"},
		IsPublic:         true,
		ValidateIsUnique: true,
	}
	if err := r.Define(dPublic); err != nil {
		t.Fatalf("r.Define(public def) unexpected err: %v", err)
	}
	dPrivate := &Def[model.IExpression]{
		Name:             "private def",
		Result:           &model.ExpressionRef{Name: "private def"},
		IsPublic:         false,
		ValidateIsUnique: true,
	}
	if err := r.Define(dPrivate); err != nil {
		t.Fatalf("r.Define(private def) unexpected err: %v", err)
	}
	ops := []types.IType{types.String}
	fPublic := &Func[model.IExpression]{
		Name:             "public func",
		Operands:         ops,
		Result:           &model.FunctionRef{Name: "public func", Operands: []model.IExpression{}},
		IsPublic:         true,
		IsFluent:         false,
		ValidateIsUnique: true,
	}
	if err := r.DefineFunc(fPublic); err != nil {
		t.Fatalf("r.Define(public def) unexpected err: %v", err)
	}
	fPublicFluent := &Func[model.IExpression]{
		Name:             "public fluent func",
		Operands:         ops,
		Result:           &model.FunctionRef{Name: "public fluent func", Operands: []model.IExpression{}},
		IsPublic:         true,
		IsFluent:         true,
		ValidateIsUnique: true,
	}
	if err := r.DefineFunc(fPublicFluent); err != nil {
		t.Fatalf("r.Define(public fluent def) unexpected err: %v", err)
	}
	fPrivate := &Func[model.IExpression]{
		Name:             "private func",
		Operands:         ops,
		Result:           &model.FunctionRef{Name: "private func", Operands: []model.IExpression{}},
		IsPublic:         false,
		IsFluent:         false,
		ValidateIsUnique: true,
	}
	if err := r.DefineFunc(fPrivate); err != nil {
		t.Fatalf("r.Define(private def) unexpected err: %v", err)
	}
}

func newFHIRModelInfo(t *testing.T) *modelinfo.ModelInfos {
	t.Helper()
	fhirMIBytes, err := embeddata.ModelInfos.ReadFile("third_party/cqframework/fhir-modelinfo-4.0.1.xml")
	if err != nil {
		t.Fatalf("Reading embedded file %s failed unexpectedly: %v", "third_party/cqframework/fhir-modelinfo-4.0.1.xml", err)
	}

	m, err := modelinfo.New([][]byte{fhirMIBytes})
	if err != nil {
		t.Fatalf("New modelinfo unexpected error: %v", err)
	}
	m.SetUsing(modelinfo.Key{Name: "FHIR", Version: "4.0.1"})
	return m
}
