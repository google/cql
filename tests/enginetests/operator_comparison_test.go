// Copyright 2024 Google LLC
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

package enginetests

import (
	"context"
	"strings"
	"testing"

	"github.com/google/cql/interpreter"
	"github.com/google/cql/model"
	"github.com/google/cql/parser"
	"github.com/google/cql/result"
	"github.com/google/cql/types"
	"github.com/google/go-cmp/cmp"
	"github.com/lithammer/dedent"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestEqual(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "null = true",
			cql:  "null = true",
			wantModel: &model.Equal{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						&model.As{
							UnaryExpression: &model.UnaryExpression{
								Expression: model.ResultType(types.Boolean),
								Operand:    model.NewLiteral("null", types.Any),
							},
							AsTypeSpecifier: types.Boolean,
						},
						model.NewLiteral("true", types.Boolean),
					},
					Expression: model.ResultType(types.Boolean),
				},
			},
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "false = null",
			cql:        "false = null",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "true = true",
			cql:        "true = true",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "false = false",
			cql:        "false = false",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "false = true",
			cql:        "false = true",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "true = false",
			cql:        "true = false",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "DateTimes equal",
			cql:        "@2024-02-29T01:20:30.101-07:00 = @2024-02-29T01:20:30.101-07:00",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "DateTimes not equal",
			cql:        "@2024-02-29T01:20:30.101-07:00 = @2028-02-29T01:20:30.101-07:00",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Equal DateTimes until differing precision is null",
			cql:        "@2024-02-29T01:20:30.101-07:00 = @2024-02-29T",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "Dates equal",
			cql:        "@2024-02-29 = @2024-02-29",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Dates not equal",
			cql:        "@2024-02-29 = @2028-02-29",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Equal Dates until differing precision is null",
			cql:        "@2024-02-29 = @2024-02",
			wantResult: newOrFatal(t, nil),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := newFHIRParser(t)
			parsedLibs, err := p.Libraries(context.Background(), wrapInLib(t, tc.cql), parser.Config{})
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantModel, getTESTRESULTModel(t, parsedLibs)); tc.wantModel != nil && diff != "" {
				t.Errorf("Parse diff (-want +got):\n%s", diff)
			}

			results, err := interpreter.Eval(context.Background(), parsedLibs, defaultInterpreterConfig(t, p))
			if err != nil {
				t.Fatalf("Eval returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantResult, getTESTRESULT(t, results), protocmp.Transform()); diff != "" {
				t.Errorf("Eval diff (-want +got)\n%v", diff)
			}
		})
	}
}

func TestNotEqual(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "null != true",
			cql:  "null != true",
			wantModel: &model.Not{
				UnaryExpression: &model.UnaryExpression{
					Expression: model.ResultType(types.Boolean),
					Operand: &model.Equal{
						BinaryExpression: &model.BinaryExpression{
							Operands: []model.IExpression{
								&model.As{
									UnaryExpression: &model.UnaryExpression{
										Expression: model.ResultType(types.Boolean),
										Operand:    model.NewLiteral("null", types.Any),
									},
									AsTypeSpecifier: types.Boolean,
								},
								model.NewLiteral("true", types.Boolean),
							},
							Expression: model.ResultType(types.Boolean),
						},
					},
				},
			},
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "false != null",
			cql:        "false != null",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "true != true",
			cql:        "true != true",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "false != false",
			cql:        "false != false",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "false != true",
			cql:        "false != true",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "true != false",
			cql:        "true != false",
			wantResult: newOrFatal(t, true),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := newFHIRParser(t)
			parsedLibs, err := p.Libraries(context.Background(), wrapInLib(t, tc.cql), parser.Config{})
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantModel, getTESTRESULTModel(t, parsedLibs)); tc.wantModel != nil && diff != "" {
				t.Errorf("Parse diff (-want +got):\n%s", diff)
			}

			results, err := interpreter.Eval(context.Background(), parsedLibs, defaultInterpreterConfig(t, p))
			if err != nil {
				t.Fatalf("Eval returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantResult, getTESTRESULT(t, results), protocmp.Transform()); diff != "" {
				t.Errorf("Eval diff (-want +got)\n%v", diff)
			}
		})
	}
}

func TestEquivalent(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "null ~ true",
			cql:  "null ~ true",
			wantModel: &model.Equivalent{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						&model.As{
							UnaryExpression: &model.UnaryExpression{
								Expression: model.ResultType(types.Boolean),
								Operand:    model.NewLiteral("null", types.Any),
							},
							AsTypeSpecifier: types.Boolean,
						},
						model.NewLiteral("true", types.Boolean),
					},
					Expression: model.ResultType(types.Boolean),
				},
			},
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Boolean nulls are equivalent",
			cql:        "null as Boolean ~ null as Boolean",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "true ~ true",
			cql:        "true ~ true",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "false ~ false",
			cql:        "false ~ false",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "false ~ true",
			cql:        "false ~ true",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "true ~ false",
			cql:        "true ~ false",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "DateTimes equivalent",
			cql:        "@2024-02-29T01:20:30.101-07:00 ~ @2024-02-29T01:20:30.101-07:00",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "DateTimes not equivalent",
			cql:        "@2024-02-29T01:20:30.101-07:00 ~ @2028-02-29T01:20:30.101-07:00",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "DateTime nulls are equivalent",
			cql:        "null as DateTime ~ null as DateTime",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "DateTime equivalent to null is false",
			cql:        "@2024-02-29T01:20:30.101-07:00 ~ null",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Equal DateTimes until differing precision is false",
			cql:        "@2024-02-29T01:20:30.101-07:00 ~ @2024-02-29T",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Dates equivalent",
			cql:        "@2024-02-29 ~ @2024-02-29",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Dates not equivalent",
			cql:        "@2024-02-29 ~ @2028-02-29",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Date nulls are equivalent",
			cql:        "null as Date ~ null as Date",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Date equivalent to null is false",
			cql:        "@2024-02-29 ~ null",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Equal Dates until differing precision is false",
			cql:        "@2024-02-29 ~ @2024-02",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Equivalent integers",
			cql:        "1 ~ 1",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Not equivalent integers",
			cql:        "1 ~ 2",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Integer nulls are equivalent",
			cql:        "null as Integer ~ null as Integer",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Integer equivalent to null is false",
			cql:        "1 ~ null",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Equivalent Long",
			cql:        "1L ~ 1L",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Not equivalent Long",
			cql:        "1L ~ 2L",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Equivalent empty Lists",
			cql:        "{} ~ {}",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Long nulls are equivalent",
			cql:        "null as Long ~ null as Long",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Date equivalent to null is false",
			cql:        "1L ~ null",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Equivalent Lists",
			cql:        "{1, 2} ~ {1, 2}",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Not equivalent Lists",
			cql:        "{1} ~ {2}",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Not equivalent Lists of different length",
			cql:        "{1} ~ {2, 3}",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "{1, 2} ~ null = false",
			cql:        "{1} ~ null",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "null as List<Any> ~ null as List<Any>",
			cql:        "null as List<Any> ~ null as List<Any>",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Equivalent lists with implicit conversions",
			cql:        "{1, 2L} ~ {1L, 2L}",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Equivalent Intervals",
			cql:        "Interval[1, 2] ~ Interval[1, 2]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Not equivalent Intervals",
			cql:        "Interval[1, 4] ~ Interval[1, 2]",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Equivalent intervals with nulls",
			cql:        "Interval(null, 4] ~ Interval(null, 4]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Non equivalent intervals with nulls",
			cql:        "null ~ Interval[1, 4]",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Equivalent null intervals",
			cql:        "null as Interval<Any> ~ null as Interval<Any>",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Equivalent intervals that require implicit conversions",
			cql:        "Interval[1, 4L] ~ Interval[1L, 4L]",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Equivalent Strings",
			cql:        "'a' ~ 'a'",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Equivalent Strings with differing case",
			cql:        "'abc' ~ 'Abc'",
			wantResult: newOrFatal(t, true),
		},
		{
			name: "Equivalent Strings with different whitespace characters",
			// This is "'a b' ~ 'a<TAB>b'"
			cql:        "'a b' ~ 'a\tb'",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Not Equivalent Strings",
			cql:        "'abc' ~ 'zbc'",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Non equivalent Strings with null",
			cql:        "'a' ~ null",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Equivalent null strings",
			cql:        "null as String ~ null as String",
			wantResult: newOrFatal(t, true),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := newFHIRParser(t)
			parsedLibs, err := p.Libraries(context.Background(), wrapInLib(t, tc.cql), parser.Config{})
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantModel, getTESTRESULTModel(t, parsedLibs)); tc.wantModel != nil && diff != "" {
				t.Errorf("Parse diff (-want +got):\n%s", diff)
			}

			results, err := interpreter.Eval(context.Background(), parsedLibs, defaultInterpreterConfig(t, p))
			if err != nil {
				t.Fatalf("Eval returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantResult, getTESTRESULT(t, results), protocmp.Transform()); diff != "" {
				t.Errorf("Eval diff (-want +got)\n%v", diff)
			}
		})
	}
}

func TestEquivalent_Errors(t *testing.T) {
	tests := []struct {
		name                string
		cql                 string
		wantModel           model.IExpression
		wantEvalErrContains string
	}{
		{
			name:                "Unsupported mixed lists equivalent",
			cql:                 "List<Any>{1, 'str', 1} ~ List<Any>{1, 'str', 1.0}",
			wantEvalErrContains: "unable to match Equivalent overload for elements in a list, this is likely because our engine does not fully support mixed type lists yet",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := newFHIRParser(t)
			parsedLibs, err := p.Libraries(context.Background(), wrapInLib(t, tc.cql), parser.Config{})
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantModel, getTESTRESULTModel(t, parsedLibs)); tc.wantModel != nil && diff != "" {
				t.Errorf("Parse diff (-want +got):\n%s", diff)
			}

			_, err = interpreter.Eval(context.Background(), parsedLibs, defaultInterpreterConfig(t, p))
			if err == nil {
				t.Fatalf("Evaluate Expression expected an error to be returned containing %q, got nil instead", tc.wantEvalErrContains)
			}
			if !strings.Contains(err.Error(), tc.wantEvalErrContains) {
				t.Errorf("Unexpected evaluation error contents got (%v) want (%v)", err.Error(), tc.wantEvalErrContains)
			}
		})
	}
}

func TestNotEquivalent(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "null !~ true",
			cql:  "null !~ true",
			wantModel: &model.Not{
				UnaryExpression: &model.UnaryExpression{
					Expression: model.ResultType(types.Boolean),
					Operand: &model.Equivalent{
						BinaryExpression: &model.BinaryExpression{
							Operands: []model.IExpression{
								&model.As{
									UnaryExpression: &model.UnaryExpression{
										Expression: model.ResultType(types.Boolean),
										Operand:    model.NewLiteral("null", types.Any),
									},
									AsTypeSpecifier: types.Boolean,
								},
								model.NewLiteral("true", types.Boolean),
							},
							Expression: model.ResultType(types.Boolean),
						},
					},
				},
			},
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "true !~ true",
			cql:        "true !~ true",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "false !~ false",
			cql:        "false !~ false",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "false !~ true",
			cql:        "false !~ true",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "true !~ false",
			cql:        "true !~ false",
			wantResult: newOrFatal(t, true),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := newFHIRParser(t)
			parsedLibs, err := p.Libraries(context.Background(), wrapInLib(t, tc.cql), parser.Config{})
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantModel, getTESTRESULTModel(t, parsedLibs)); tc.wantModel != nil && diff != "" {
				t.Errorf("Parse diff (-want +got):\n%s", diff)
			}

			results, err := interpreter.Eval(context.Background(), parsedLibs, defaultInterpreterConfig(t, p))
			if err != nil {
				t.Fatalf("Eval returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantResult, getTESTRESULT(t, results), protocmp.Transform()); diff != "" {
				t.Errorf("Eval diff (-want +got)\n%v", diff)
			}
		})
	}
}

func TestEquivalentCodes(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		// Code equivalency tests, per https://cql.hl7.org/09-b-cqlreference.html#equivalent-3
		{
			name: `null ~ Code`,
			cql: dedent.Dedent(`
			codesystem cs: 'https://example.com/cs/diagnosis' version '1.0'
			define TESTRESULT: null as Code ~ Code 'code1' from "cs" display 'display1'`),
			wantModel: &model.Equivalent{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						&model.As{
							UnaryExpression: &model.UnaryExpression{
								Expression: model.ResultType(types.Code),
								Operand:    model.NewLiteral("null", types.Any),
							},
							AsTypeSpecifier: types.Code,
						},
						&model.Code{
							System:     &model.CodeSystemRef{Name: "cs", Expression: model.ResultType(types.CodeSystem)},
							Code:       "code1",
							Display:    "display1",
							Expression: model.ResultType(types.Code),
						},
					},
					Expression: model.ResultType(types.Boolean),
				},
			},
			wantResult: newOrFatal(t, false),
		},
		{
			name: `Equivalent codes`,
			cql: dedent.Dedent(`
			codesystem cs: 'https://example.com/cs/diagnosis' version '1.0'
			define TESTRESULT: Code 'code1' from "cs" display 'display1' ~  Code 'code1' from "cs" display 'display1'`),
			wantResult: newOrFatal(t, true),
		},
		{
			name:       `Equivalent codes uses string equivalency for codes`,
			cql:        dedent.Dedent(`define TESTRESULT: Code { system: 'system1', code: '1\t1' } ~  Code { system: 'system1', code: '1 1' }`),
			wantResult: newOrFatal(t, true),
		},
		{
			name: `Equivalent codes uses string equivalency for system`,
			cql: dedent.Dedent(`
			define TESTRESULT: Code { system: 'system 1', code: '1' } ~  Code { system: 'system\t1', code: '1' }`),
			wantResult: newOrFatal(t, true),
		},
		{
			name: `Codes with different displays still true`,
			cql: dedent.Dedent(`
			codesystem cs: 'https://example.com/cs/diagnosis' version '1.0'
			define TESTRESULT: Code 'code1' from "cs" display 'display1' ~  Code 'code1' from "cs" display 'display2'`),
			wantResult: newOrFatal(t, true),
		},
		{
			name: `Codes with different systems`,
			cql: dedent.Dedent(`
			codesystem cs: 'https://example.com/cs/diagnosis' version '1.0'
			codesystem cs2: 'url' version '1.0'
			define TESTRESULT: Code 'code1' from "cs" display 'display1' ~  Code 'code1' from "cs2" display 'display1'`),
			wantResult: newOrFatal(t, false),
		},
		{
			name: `Codes with different codes`,
			cql: dedent.Dedent(`
			codesystem cs: 'https://example.com/cs/diagnosis' version '1.0'
			define TESTRESULT: Code 'code1' from "cs" display 'display1' ~  Code 'code2' from "cs" display 'display1'`),
			wantResult: newOrFatal(t, false),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := newFHIRParser(t)
			parsedLibs, err := p.Libraries(context.Background(), []string{tc.cql}, parser.Config{})
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantModel, getTESTRESULTModel(t, parsedLibs)); tc.wantModel != nil && diff != "" {
				t.Errorf("Parse diff (-want +got):\n%s", diff)
			}

			results, err := interpreter.Eval(context.Background(), parsedLibs, defaultInterpreterConfig(t, p))
			if err != nil {
				t.Fatalf("Eval returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantResult, getTESTRESULT(t, results), protocmp.Transform()); diff != "" {
				t.Errorf("Eval diff (-want +got)\n%v", diff)
			}
		})
	}
}

func TestNotEquivalentCodes(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		// Code equivalency tests, per https://cql.hl7.org/09-b-cqlreference.html#equivalent-3
		{
			name: `Not Equivalent codes`,
			cql: dedent.Dedent(`
			codesystem cs: 'https://example.com/cs/diagnosis' version '1.0'
			code c1: 'code1' from "cs"
			code c2: 'code2' from "cs" display 'display1'
			define TESTRESULT: c1 !~  c2`),
			wantResult: newOrFatal(t, true),
		},
		{
			name: `Not Equivalent on equivalent codes`,
			cql: dedent.Dedent(`
			codesystem cs: 'https://example.com/cs/diagnosis' version '1.0'
			code c1: 'code1' from "cs"
			code c2: 'code1' from "cs" display 'display1'
			define TESTRESULT: c1 !~  c2`),
			wantResult: newOrFatal(t, false),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := newFHIRParser(t)
			parsedLibs, err := p.Libraries(context.Background(), []string{tc.cql}, parser.Config{})
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantModel, getTESTRESULTModel(t, parsedLibs)); tc.wantModel != nil && diff != "" {
				t.Errorf("Parse diff (-want +got):\n%s", diff)
			}

			results, err := interpreter.Eval(context.Background(), parsedLibs, defaultInterpreterConfig(t, p))
			if err != nil {
				t.Fatalf("Eval returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantResult, getTESTRESULT(t, results), protocmp.Transform()); diff != "" {
				t.Errorf("Eval diff (-want +got)\n%v", diff)
			}
		})
	}
}

func TestEquivalentConceptCode(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		// (Concept, Code) equivalency tests, per
		// https://cql.hl7.org/09-b-cqlreference.html#equivalent-3
		{
			name: `Equivalent Concept and Code`,
			cql:  "define TESTRESULT: Equivalent(Concept { codes: { Code { system: 'http://example.com', code: '1' } } }, Code { system: 'http://example.com', code: '1' })",
			wantModel: &model.Equivalent{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						&model.Instance{
							Expression: model.ResultType(types.Concept),
							ClassType:  types.Concept,
							Elements: []*model.InstanceElement{
								&model.InstanceElement{
									Name: "codes",
									Value: &model.List{
										Expression: model.ResultType(&types.List{ElementType: types.Code}),
										List: []model.IExpression{
											&model.Instance{
												Expression: model.ResultType(types.Code),
												ClassType:  types.Code,
												Elements: []*model.InstanceElement{
													&model.InstanceElement{Name: "system", Value: model.NewLiteral("http://example.com", types.String)},
													&model.InstanceElement{Name: "code", Value: model.NewLiteral("1", types.String)},
												},
											},
										},
									},
								},
							},
						},
						&model.Instance{
							Expression: model.ResultType(types.Code),
							ClassType:  types.Code,
							Elements: []*model.InstanceElement{
								&model.InstanceElement{Name: "system", Value: model.NewLiteral("http://example.com", types.String)},
								&model.InstanceElement{Name: "code", Value: model.NewLiteral("1", types.String)},
							},
						},
					},
					Expression: model.ResultType(types.Boolean),
				},
			},
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Equivalent with ~ operator",
			cql:        "define TESTRESULT: Concept { codes: { Code { system: 'http://example.com', code: '1' } } } ~ Code { system: 'http://example.com', code: '1' }",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Equivalent where Concept has multiple codes",
			cql:        "define TESTRESULT: Equivalent(Concept { codes: { Code { system: 'http://example.com', code: '1' }, Code { system: 'http://example.com', code: '2' } } }, Code { system: 'http://example.com', code: '1' })",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Equivalent uses string equivalency for code comparison",
			cql:        "define TESTRESULT: Equivalent(Concept { codes: { Code { system: 'http://example.com', code: '1 1' } } }, Code { system: 'http://example.com', code: '1\t1' })",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Not Equivalent",
			cql:        "define TESTRESULT: Equivalent(Concept { codes: { Code { system: 'http://example.com', code: '1' } } }, Code { system: 'http://example.com', code: '2' })",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Equivalent(null, null)",
			cql:        "define TESTRESULT: Equivalent(null as Concept, null as Code)",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "Equivalent(null, Code)",
			cql:        "define TESTRESULT: Equivalent(null as Concept, Code { system: 'http://example.com', code: '1' })",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "Equivalent(Concept, null)",
			cql:        "define TESTRESULT: Equivalent(Concept { codes: { Code { system: 'http://example.com', code: '1' } } }, null as Code)",
			wantResult: newOrFatal(t, false),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := newFHIRParser(t)
			parsedLibs, err := p.Libraries(context.Background(), []string{tc.cql}, parser.Config{})
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantModel, getTESTRESULTModel(t, parsedLibs)); tc.wantModel != nil && diff != "" {
				t.Errorf("Parse diff (-want +got):\n%s", diff)
			}

			results, err := interpreter.Eval(context.Background(), parsedLibs, defaultInterpreterConfig(t, p))
			if err != nil {
				t.Fatalf("Eval returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantResult, getTESTRESULT(t, results), protocmp.Transform()); diff != "" {
				t.Errorf("Eval diff (-want +got)\n%v", diff)
			}
		})
	}
}

func TestGreater(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "2 > 1",
			cql:  "2 > 1",
			wantModel: &model.Greater{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("2", types.Integer),
						model.NewLiteral("1", types.Integer),
					},
					Expression: model.ResultType(types.Boolean),
				},
			},
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "1 > 2",
			cql:        "1 > 2",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "1 > 1",
			cql:        "1 > 1",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "1 > null",
			cql:        "1 > null",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "2L > 1L",
			cql:        "2L > 1L",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "1L > 2L",
			cql:        "1L > 2L",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "1L > 1L",
			cql:        "1L > 1L",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "1L > null",
			cql:        "1L > null",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "1.1 > 1.0",
			cql:        "1.1 > 1.0",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "1.0 > 2.0",
			cql:        "1.0 > 2.0",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "1.0 > 1.0",
			cql:        "1.0 > 1.0",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "2.0 > null",
			cql:        "2.0 > null",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "@2020 > @2019",
			cql:        "@2020 > @2019",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "@2019 > @2020",
			cql:        "@2019 > @2020",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "@2019 > @2019",
			cql:        "@2019 > @2019",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "@2019 > null",
			cql:        "@2019 > null",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "@2020-01 > @2020 left has greater precision",
			cql:        "@2020-01 > @2020",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "@2020-01-02T02 > @2020-01-02T01",
			cql:        "@2020-01-02T02:01:00.000Z > @2020-01-02T01:01:00.000Z",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "@2020-01-02T01 > @2020-01-02T02",
			cql:        "@2020-01-02T01:01:00.000Z > @2020-01-02T02:01:00.000Z",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "@2020-01-02T02 > @2020-01-02T02",
			cql:        "@2020-01-02T02:01:00.000Z > @2020-01-02T02:01:00.000Z",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "@2020-01-02T01 > null",
			cql:        "@2020-01-02T01:01:00.000Z > null",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "'ab' > 'aa'",
			cql:        "'ab' > 'aa'",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "'aa' > 'ab'",
			cql:        "'aa' > 'ab'",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "'aa' > 'aa'",
			cql:        "'aa' > 'aa'",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "'ab' > null",
			cql:        "'ab' > null",
			wantResult: newOrFatal(t, nil),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := newFHIRParser(t)
			parsedLibs, err := p.Libraries(context.Background(), wrapInLib(t, tc.cql), parser.Config{})
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantModel, getTESTRESULTModel(t, parsedLibs)); tc.wantModel != nil && diff != "" {
				t.Errorf("Parse diff (-want +got):\n%s", diff)
			}

			results, err := interpreter.Eval(context.Background(), parsedLibs, defaultInterpreterConfig(t, p))
			if err != nil {
				t.Fatalf("Eval returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantResult, getTESTRESULT(t, results), protocmp.Transform()); diff != "" {
				t.Errorf("Eval diff (-want +got)\n%v", diff)
			}
		})
	}
}

func TestGreaterOrEqual(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "2 >= 1",
			cql:  "2 >= 1",
			wantModel: &model.GreaterOrEqual{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("2", types.Integer),
						model.NewLiteral("1", types.Integer),
					},
					Expression: model.ResultType(types.Boolean),
				},
			},
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "1 >= 2",
			cql:        "1 >= 2",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "1 >= 1",
			cql:        "1 >= 1",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "1 >= null",
			cql:        "1 >= null",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "2L >= 1L",
			cql:        "2L >= 1L",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "1L >= 2L",
			cql:        "1L >= 2L",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "1L >= 1L",
			cql:        "1L >= 1L",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "1L >= null",
			cql:        "1L >= null",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "1.1 >= 1.0",
			cql:        "1.1 >= 1.0",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "1.0 >= 2.0",
			cql:        "1.0 >= 2.0",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "1.0 >= 1.0",
			cql:        "1.0 >= 1.0",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "2.0 >= null",
			cql:        "2.0 >= null",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "@2020 >= @2019",
			cql:        "@2020 >= @2019",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "@2019 >= @2020",
			cql:        "@2019 >= @2020",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "@2019 >= @2019",
			cql:        "@2019 >= @2019",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "@2019 >= null",
			cql:        "@2019 >= null",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "@2020-01 >= @2020 left has greater precision",
			cql:        "@2020-01 >= @2020",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "@2020-01-02T02 >= @2020-01-02T01",
			cql:        "@2020-01-02T02:01:00.000Z >= @2020-01-02T01:01:00.000Z",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "@2020-01-02T01 >= @2020-01-02T02",
			cql:        "@2020-01-02T01:01:00.000Z >= @2020-01-02T02:01:00.000Z",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "@2020-01-02T02 >= @2020-01-02T02",
			cql:        "@2020-01-02T02:01:00.000Z >= @2020-01-02T02:01:00.000Z",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "@2020-01-02T01 >= null",
			cql:        "@2020-01-02T01:01:00.000Z >= null",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "'ab' >= 'aa'",
			cql:        "'ab' >= 'aa'",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "'aa' >= 'ab'",
			cql:        "'aa' >= 'ab'",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "'aa' >= 'aa'",
			cql:        "'aa' >= 'aa'",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "'ab' >= null",
			cql:        "'ab' >= null",
			wantResult: newOrFatal(t, nil),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := newFHIRParser(t)
			parsedLibs, err := p.Libraries(context.Background(), wrapInLib(t, tc.cql), parser.Config{})
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantModel, getTESTRESULTModel(t, parsedLibs)); tc.wantModel != nil && diff != "" {
				t.Errorf("Parse diff (-want +got):\n%s", diff)
			}

			results, err := interpreter.Eval(context.Background(), parsedLibs, defaultInterpreterConfig(t, p))
			if err != nil {
				t.Fatalf("Eval returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantResult, getTESTRESULT(t, results), protocmp.Transform()); diff != "" {
				t.Errorf("Eval diff (-want +got)\n%v", diff)
			}
		})
	}
}

func TestLess(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "2 < 1",
			cql:  "2 < 1",
			wantModel: &model.Less{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("2", types.Integer),
						model.NewLiteral("1", types.Integer),
					},
					Expression: model.ResultType(types.Boolean),
				},
			},
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "1 < 2",
			cql:        "1 < 2",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "1 < 1",
			cql:        "1 < 1",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "1 < null",
			cql:        "1 < null",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "2L < 1L",
			cql:        "2L < 1L",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "1L < 2L",
			cql:        "1L < 2L",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "1L < 1L",
			cql:        "1L < 1L",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "1L < null",
			cql:        "1L < null",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "1.1 < 1.0",
			cql:        "1.1 < 1.0",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "1.0 < 2.0",
			cql:        "1.0 < 2.0",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "1.0 < 1.0",
			cql:        "1.0 < 1.0",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "2.0 < null",
			cql:        "2.0 < null",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "@2020 < @2019",
			cql:        "@2020 < @2019",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "@2019 < @2020",
			cql:        "@2019 < @2020",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "@2019 < @2019",
			cql:        "@2019 < @2019",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "@2019 < null",
			cql:        "@2019 < null",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "@2020-01 < @2020 left has greater precision",
			cql:        "@2020-01 < @2020",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "@2020-01-02T02 < @2020-01-02T01",
			cql:        "@2020-01-02T02:01:00.000Z < @2020-01-02T01:01:00.000Z",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "@2020-01-02T01 < @2020-01-02T02",
			cql:        "@2020-01-02T01:01:00.000Z < @2020-01-02T02:01:00.000Z",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "@2020-01-02T02 < @2020-01-02T02",
			cql:        "@2020-01-02T02:01:00.000Z < @2020-01-02T02:01:00.000Z",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "@2020-01-02T01 < null",
			cql:        "@2020-01-02T01:01:00.000Z < null",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "'ab' < 'aa'",
			cql:        "'ab' < 'aa'",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "'aa' < 'ab'",
			cql:        "'aa' < 'ab'",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "'aa' < 'aa'",
			cql:        "'aa' < 'aa'",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "'ab' < null",
			cql:        "'ab' < null",
			wantResult: newOrFatal(t, nil),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := newFHIRParser(t)
			parsedLibs, err := p.Libraries(context.Background(), wrapInLib(t, tc.cql), parser.Config{})
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantModel, getTESTRESULTModel(t, parsedLibs)); tc.wantModel != nil && diff != "" {
				t.Errorf("Parse diff (-want +got):\n%s", diff)
			}

			results, err := interpreter.Eval(context.Background(), parsedLibs, defaultInterpreterConfig(t, p))
			if err != nil {
				t.Fatalf("Eval returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantResult, getTESTRESULT(t, results), protocmp.Transform()); diff != "" {
				t.Errorf("Eval diff (-want +got)\n%v", diff)
			}
		})
	}
}

func TestLessOrEqual(t *testing.T) {
	tests := []struct {
		name       string
		cql        string
		wantModel  model.IExpression
		wantResult result.Value
	}{
		{
			name: "2 <= 1",
			cql:  "2 <= 1",
			wantModel: &model.LessOrEqual{
				BinaryExpression: &model.BinaryExpression{
					Operands: []model.IExpression{
						model.NewLiteral("2", types.Integer),
						model.NewLiteral("1", types.Integer),
					},
					Expression: model.ResultType(types.Boolean),
				},
			},
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "1 <= 2",
			cql:        "1 <= 2",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "1 <= 1",
			cql:        "1 <= 1",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "1 <= null",
			cql:        "1 <= null",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "2L <= 1L",
			cql:        "2L <= 1L",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "1L <= 2L",
			cql:        "1L <= 2L",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "1L <= 1L",
			cql:        "1L <= 1L",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "1L <= null",
			cql:        "1L <= null",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "1.1 <= 1.0",
			cql:        "1.1 <= 1.0",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "1.0 <= 2.0",
			cql:        "1.0 <= 2.0",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "1.0 <= 1.0",
			cql:        "1.0 <= 1.0",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "2.0 <= null",
			cql:        "2.0 <= null",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "@2020 <= @2019",
			cql:        "@2020 <= @2019",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "@2019 <= @2020",
			cql:        "@2019 <= @2020",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "@2019 <= @2019",
			cql:        "@2019 <= @2019",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "@2019 <= null",
			cql:        "@2019 <= null",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "@2020-01 <= @2020 left has greater precision",
			cql:        "@2020-01 <= @2020",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "@2020-01-02T02 <= @2020-01-02T01",
			cql:        "@2020-01-02T02:01:00.000Z <= @2020-01-02T01:01:00.000Z",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "@2020-01-02T01 <= @2020-01-02T02",
			cql:        "@2020-01-02T01:01:00.000Z <= @2020-01-02T02:01:00.000Z",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "@2020-01-02T02 <= @2020-01-02T02",
			cql:        "@2020-01-02T02:01:00.000Z <= @2020-01-02T02:01:00.000Z",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "@2020-01-02T01 <= null",
			cql:        "@2020-01-02T01:01:00.000Z <= null",
			wantResult: newOrFatal(t, nil),
		},
		{
			name:       "'ab' <= 'aa'",
			cql:        "'ab' <= 'aa'",
			wantResult: newOrFatal(t, false),
		},
		{
			name:       "'aa' <= 'ab'",
			cql:        "'aa' <= 'ab'",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "'aa' <= 'aa'",
			cql:        "'aa' <= 'aa'",
			wantResult: newOrFatal(t, true),
		},
		{
			name:       "'ab' <= null",
			cql:        "'ab' <= null",
			wantResult: newOrFatal(t, nil),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := newFHIRParser(t)
			parsedLibs, err := p.Libraries(context.Background(), wrapInLib(t, tc.cql), parser.Config{})
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantModel, getTESTRESULTModel(t, parsedLibs)); tc.wantModel != nil && diff != "" {
				t.Errorf("Parse diff (-want +got):\n%s", diff)
			}

			results, err := interpreter.Eval(context.Background(), parsedLibs, defaultInterpreterConfig(t, p))
			if err != nil {
				t.Fatalf("Eval returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantResult, getTESTRESULT(t, results), protocmp.Transform()); diff != "" {
				t.Errorf("Eval diff (-want +got)\n%v", diff)
			}
		})
	}
}
