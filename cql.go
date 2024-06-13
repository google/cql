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

// Package cql provides tools for parsing and evaluating CQL.
package cql

import (
	"context"
	"fmt"
	"time"

	"github.com/google/cql/internal/embeddata"
	"github.com/google/cql/internal/modelinfo"
	"github.com/google/cql/interpreter"
	"github.com/google/cql/model"
	"github.com/google/cql/parser"
	"github.com/google/cql/result"
	"github.com/google/cql/retriever"
	"github.com/google/cql/terminology"
)

// ParseConfig configures the parsing of CQL to our internal ELM like data structure.
type ParseConfig struct {
	// DataModels are the xml model info files of the data models that will be used by the parser and
	// interpreter. The system model info is included by default. DataModels are optional and could be
	// nil in which case the CQL can only use the system data model.
	DataModels [][]byte

	// Parameters map between the parameters DefKey and a CQL literal. The DefKey specifies the
	// library and the parameters name. The CQL Literal cannot be an expression definition, valueset
	// or other CQL construct. It cannot reference other definitions or call functions. It is parsed
	// at Term in the CQL grammar. Examples of parameters could be 100, 'string',
	// Interval[@2013-01-01T00:00:00.0, @2014-01-01T00:00:00.0) or {1, 2}. Parameters are optional and
	// could be nil.
	Parameters map[result.DefKey]string
}

// Parse parses CQL libraries into our internal ELM like data structure, which can then be
// evaluated.
// Errors returned by Parse will always be a result.EngineError.
func Parse(ctx context.Context, libs []string, config ParseConfig) (*ELM, error) {
	p, err := parser.New(ctx, config.DataModels)
	if err != nil {
		return nil, err
	}
	parsedLibs, err := p.Libraries(ctx, libs, parser.Config{})
	if err != nil {
		return nil, err
	}
	parsedParams, err := p.Parameters(ctx, config.Parameters, parser.Config{})
	if err != nil {
		return nil, err
	}

	return &ELM{
		dataModels:   p.DataModel(),
		parsedParams: parsedParams,
		parsedLibs:   parsedLibs,
	}, nil
}

// EvalConfig configures the interpreter to evaluate ELM to final CQL Results.
type EvalConfig struct {
	// Terminology is the interface through which the interpreter connects to terminology servers. If
	// the CQL being evaluated does not require a terminology server this can be left nil. To connect
	// to a terminology server you will need to implement the terminology.Provider interface, or use
	// one of the included terminology providers. See the terminology package for more details.
	Terminology terminology.Provider

	// EvaluationTimestamp is the time at which the eval request will be executed. The timestamp is
	// used by CQL system operators like Today() and Now(). If not provided EvaluationTimestamp will
	// default to time.Now() called at the start of the eval request.
	EvaluationTimestamp time.Time

	// ReturnPrivateDefs if true will return all private definitions in result.Libraries. By default
	// only public definitions are returned.
	ReturnPrivateDefs bool
}

// Eval executes the parsed CQL against the retriever. The retriever is the interface through which
// the interpreter retrieves external data. So if for example executing the parsed CQL against a
// list of patients, Eval can be called once for each patient with a retriever initialized to
// retrieve data for that patient. To connect to a particular data source you will need to implement
// the retriever.Retriever interface, or use one of the included retrievers. See the retriever
// package for more details. The retriever can be nil if the CQL does not fetch external data. Eval
// should not be called from multiple goroutines on a single *ELM.
// Errors returned by Eval will always be a result.EngineError.
func (e *ELM) Eval(ctx context.Context, retriever retriever.Retriever, config EvalConfig) (result.Libraries, error) {
	c := interpreter.Config{
		DataModels:          e.dataModels,
		Parameters:          e.parsedParams,
		Retriever:           retriever,
		Terminology:         config.Terminology,
		EvaluationTimestamp: config.EvaluationTimestamp,
		ReturnPrivateDefs:   config.ReturnPrivateDefs,
	}

	return interpreter.Eval(ctx, e.parsedLibs, c)
}

// ELM is the parsed CQL, ready to be evaluated.
type ELM struct {
	dataModels   *modelinfo.ModelInfos
	parsedParams map[result.DefKey]model.IExpression
	parsedLibs   []*model.Library
}

// FHIRDataModelAndHelpersLib returns the model info xml file for a FHIR data model and the
// FHIRHelpers CQL library. Currently only version 4.0.1 is supported.
func FHIRDataModelAndHelpersLib(version string) (fhirDM []byte, fhirHelpers string, err error) {
	fhirDM, err = FHIRDataModel(version)
	if err != nil {
		return
	}
	fhirHelpers, err = FHIRHelpersLib(version)
	if err != nil {
		return
	}
	return
}

// FHIRHelpersLib returns the FHIRHelpers CQL library. Currently only version 4.0.1 is supported.
func FHIRHelpersLib(version string) (string, error) {
	if version != "4.0.1" {
		return "", fmt.Errorf("FHIRHelpersLib only supports version 4.0.1, got: %v", version)
	}
	fhirHelpers, err := embeddata.FHIRHelpers.ReadFile("third_party/cqframework/FHIRHelpers-4.0.1.cql")
	if err != nil {
		return "", fmt.Errorf("internal error - could not read FHIRHelpers-4.0.1.cql: %w", err)
	}
	return string(fhirHelpers), nil
}

// FHIRDataModel returns the model info xml file for a FHIR data model. Currently only version 4.0.1
// is supported.
func FHIRDataModel(version string) ([]byte, error) {
	if version != "4.0.1" {
		return nil, fmt.Errorf("FHIRDataModel only supports version 4.0.1, got: %v", version)
	}
	fhirMI, err := embeddata.ModelInfos.ReadFile("third_party/cqframework/fhir-modelinfo-4.0.1.xml")
	if err != nil {
		return nil, fmt.Errorf("internal error - could not read fhir-modelinfo-4.0.1.xml: %w", err)
	}
	return fhirMI, nil
}
