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

package parser

import (
	"fmt"
	"strings"

	"github.com/google/cql/model"
	"github.com/google/cql/result"
	"github.com/google/cql/types"
	"github.com/antlr4-go/antlr/v4"
)

var _ antlr.ErrorListener = &visitor{}

// LibraryErrors contains a list of CQL parsing errors that occurred within a single library.
type LibraryErrors struct {
	LibKey result.LibKey
	Errors []*ParsingError
}

func (le *LibraryErrors) Error() string {
	msgs := []string{fmt.Sprintf("error(s) in Library %q:", le.LibKey.String())}
	for _, e := range le.Errors {
		msgs = append(msgs, e.Error())
	}
	return strings.Join(msgs, "\n")
}

// Unwrap implements the Go standard errors package Unwrap() function. See
// https://pkg.go.dev/errors.
func (le *LibraryErrors) Unwrap() []error {
	if le == nil {
		return nil
	}
	errs := make([]error, 0, len(le.Errors))
	for _, err := range le.Errors {
		errs = append(errs, err)
	}
	return errs
}

// Append adds the given error to the list of ParsingErrors.
func (le *LibraryErrors) Append(e *ParsingError) {
	le.Errors = append(le.Errors, e)
}

// ParameterErrors contains a list of CQL parsing errors that occurred parsing a single parameter.
type ParameterErrors struct {
	DefKey result.DefKey
	Errors []*ParsingError
}

func (pe *ParameterErrors) Error() string {
	var msgs []string
	for _, e := range pe.Errors {
		msgs = append(msgs, e.Error())
	}
	return strings.Join(msgs, "\n")
}

// Unwrap implements the Go standard errors package Unwrap() function. See
// https://pkg.go.dev/errors.
func (pe *ParameterErrors) Unwrap() []error {
	if pe == nil {
		return nil
	}
	errs := make([]error, 0, len(pe.Errors))
	for _, err := range pe.Errors {
		errs = append(errs, err)
	}
	return errs
}

// Append adds the given error to the list of ParsingErrors.
func (pe *ParameterErrors) Append(e *ParsingError) {
	pe.Errors = append(pe.Errors, e)
}

// ErrorType is the type of parsing error.
type ErrorType string

const (
	// SyntaxError is returned by the lexer when the CQL does not meet the grammar.
	SyntaxError = ErrorType("SyntaxError")
	// ValidationError is returned by the parser when the CQL meets the grammar, but does not meet
	// some other validation rules (like referencing a non-existent expression definition).
	ValidationError = ErrorType("ValidationError")
	// InternalError occurs when the parser errors in an unexpected way. This is not a user error, nor
	// a feature that we purposefully do not support.
	InternalError = ErrorType("InternalError")
	// UnsupportedError is return for CQL language features that are not yet supported.
	UnsupportedError = ErrorType("UnsupportedError")
)

// ErrorSeverity represents different ParsingError severity levels.
type ErrorSeverity string

const (
	// ErrorSeverityInfo is informational.
	ErrorSeverityInfo = ErrorSeverity("Info")
	// ErrorSeverityWarning is a medium severity error.
	ErrorSeverityWarning = ErrorSeverity("Warning")
	// ErrorSeverityError is a high severity error.
	ErrorSeverityError = ErrorSeverity("Error")
)

// ParsingError represents a specific parser error and its location.
type ParsingError struct {
	// High level message about the error.
	Message string
	// Line is the 1-based line number within source file where the error occurred.
	Line int
	// Column is the 0-based column number within source file where the error occurred.
	Column int
	// Type is the type of the error that occurred, such as SyntaxError or InternalError.
	Type ErrorType
	// Severity represents different severity levels.
	Severity ErrorSeverity
	// Cause is an optional, underlying error that caused the parsing error.
	Cause error
}

func (pe *ParsingError) Error() string {
	if pe.Cause != nil {
		return fmt.Sprintf("%d-%d %s: %s", pe.Line, pe.Column, pe.Message, pe.Cause)
	}
	return fmt.Sprintf("%d-%d %s", pe.Line, pe.Column, pe.Message)
}

func (pe *ParsingError) Unwrap() error {
	return pe.Cause
}

// invalidExpression is a placeholder that allows parsing to continue so any additional
// errors can be reported by the parser.
type invalidExpression struct {
	*model.Expression
	ParsingError *ParsingError
}

// badExpression reports a parsing error and returns a placeholder allowing parsing to continue.
func (v visitor) badExpression(msg string, ctx antlr.ParserRuleContext) invalidExpression {
	return invalidExpression{
		ParsingError: v.reportError(msg, ctx),
		Expression:   model.ResultType(types.Any),
	}
}

// invalidTypeSpecifier is a placeholder that allows parsing to continue so any additional
// errors can be reported by the parser.
type invalidTypeSpecifier struct {
	types.System
	ParsingError *ParsingError
}

// badTypeSpecifier reports a parsing error and returns a placeholder allowing parsing to continue.
func (v visitor) badTypeSpecifier(msg string, ctx antlr.ParserRuleContext) invalidTypeSpecifier {
	return invalidTypeSpecifier{
		ParsingError: v.reportError(msg, ctx),
		System:       types.Any,
	}
}

// reportError reports an error within the visitor, and returns the ParsingError that may or may not
// be of use to the caller.
func (v *visitor) reportError(msg string, ctx antlr.ParserRuleContext) *ParsingError {
	pe := &ParsingError{
		Message: msg,
		Line:    ctx.GetStart().GetLine(),
		Column:  ctx.GetStart().GetColumn(),
	}
	v.errors.Append(pe)
	return pe
}

// SyntaxError is called by ANTLR generated code the CQL does not meet the grammar.
func (v visitor) SyntaxError(recognizer antlr.Recognizer, offendingSymbol any, line, column int,
	msg string, e antlr.RecognitionException) {
	v.errors.Append(&ParsingError{Message: msg, Line: line, Column: column})
}

// These Report* functions needed to implement the error listener interface, but there is
// no action to be taken when parsing the CQL grammar so they do nothing.
func (v visitor) ReportAmbiguity(recognizer antlr.Parser, dfa *antlr.DFA, startIndex,
	stopIndex int, exact bool, ambigAlts *antlr.BitSet, configs *antlr.ATNConfigSet) {
	// Intentional
}

func (v visitor) ReportAttemptingFullContext(recognizer antlr.Parser, dfa *antlr.DFA,
	startIndex, stopIndex int, conflictingAlts *antlr.BitSet, configs *antlr.ATNConfigSet) {
	// Intentional
}

func (v visitor) ReportContextSensitivity(recognizer antlr.Parser, dfa *antlr.DFA,
	startIndex, stopIndex, prediction int, configs *antlr.ATNConfigSet) {
	// Intentional
}
