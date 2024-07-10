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

// Package result defines the evaluation results that can be returned by the CQL Engine.
package result

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/cql/model"
	crpb "github.com/google/cql/protos/cql_result_go_proto"
	"google.golang.org/protobuf/proto"
	"github.com/pborman/uuid"
)

// Libraries returns the results of the evaluation of a set of CQL Libraries. The inner
// map[string]Value maps the name of each expression definition to the resulting CQL Value.
// The outer map[LibKey] maps CQL Libraries to the Expression Definitions within the library.
type Libraries map[LibKey]map[string]Value

type cqlLibJSON struct {
	Name    string           `json:"libName"`
	Version string           `json:"libVersion"`
	ExpDefs map[string]Value `json:"expressionDefinitions"`
}

// MarshalJSON returns the CQL Results as a JSON. The JSON will be a list of CQL libraries
// formatted like the following:
//
//	[{
//		'libName': 'TESTLIB',
//		'libVersion': '1.0.0',
//		'expressionDefinitions': {'ExpDef': 3, 'ExpDef2': 4},
//	}, ...],
func (l Libraries) MarshalJSON() ([]byte, error) {
	r := []cqlLibJSON{}
	for k, v := range l {
		r = append(r, cqlLibJSON{
			Name:    k.Name,
			Version: k.Version,
			ExpDefs: v,
		})
	}

	return json.Marshal(r)
}

// Proto converts Libraries to a proto.
func (l Libraries) Proto() (*crpb.Libraries, error) {
	pbLibraries := &crpb.Libraries{
		Libraries: make([]*crpb.Library, 0, len(l)),
	}

	for libKey, lib := range l {
		pbLib := crpb.Library{
			Name:     proto.String(libKey.Name),
			Version:  proto.String(libKey.Version),
			ExprDefs: make(map[string]*crpb.Value, len(lib)),
		}
		for defName, def := range lib {
			pbRes, err := def.Proto()
			if err != nil {
				return nil, err
			}
			pbLib.ExprDefs[defName] = pbRes
		}
		pbLibraries.Libraries = append(pbLibraries.Libraries, &pbLib)
	}
	return pbLibraries, nil
}

// LibrariesFromProto converts a proto to Libraries.
func LibrariesFromProto(pb *crpb.Libraries) (Libraries, error) {
	libraries := Libraries{}
	for _, lib := range pb.Libraries {
		libKey := LibKey{Name: lib.GetName(), Version: lib.GetVersion()}
		libMap := make(map[string]Value, len(lib.ExprDefs))
		for defName, resultpb := range lib.ExprDefs {
			val, err := NewFromProto(resultpb)
			if err != nil {
				return nil, err
			}
			libMap[defName] = val
		}
		libraries[libKey] = libMap
	}
	return libraries, nil
}

// LibKey is the unique identifier for a CQL Library.
type LibKey struct {
	// Name is the fully qualified identifier of the CQL library.
	Name string
	// Version is empty if no version was specified.
	Version string
	// Unnamed libraries do not have a library identifier. Unnamed libraries cannot be referenced, and
	// all definitions are private. Use UnnamedLibKey() to create an unnamed library.
	IsUnnamed bool
}

// UnnamedLibKey returns a LibKey for a library without an identifier. The Name will be "Unnamed
// Library" and the Version will be a random UUID.
func UnnamedLibKey() LibKey {
	return LibKey{Name: "Unnamed Library", Version: uuid.New(), IsUnnamed: true}
}

// LibKeyFromModel is a convenience constructor that returns a LibKey from a
// model.LibraryIdentifier. If the model.LibraryIdentifier is nil, an UnnamedLibKey is returned.
func LibKeyFromModel(lib *model.LibraryIdentifier) LibKey {
	if lib == nil {
		return UnnamedLibKey()
	}
	return LibKey{Name: lib.Qualified, Version: lib.Version, IsUnnamed: false}
}

// Key returns a unique string key representation of the LibKey.
// A space is added between the name and version to avoid naming clashes.
func (l LibKey) Key() string {
	if l.Version == "" {
		return l.Name
	}
	// TODO b/301606416 - Since identifiers can contain spaces, this is not a unique key.
	return l.Name + " " + l.Version
}

// String returns a printable representation of LibKey.
func (l LibKey) String() string {
	if l.IsUnnamed {
		return "Unnamed Library"
	}
	if l.Version == "" {
		return l.Name
	}
	return l.Name + " " + l.Version
}

// DefKey is the unique identifier for a CQL Expression Definition, Parameter or ValueSet.
type DefKey struct {
	Name    string
	Library LibKey
}

// EngineErrorType is the type of error to be set on the EngineError.
type EngineErrorType error

var (
	// ErrLibraryParsing is returned when a library could not be properly parsed.
	ErrLibraryParsing = errors.New("failed to parse library")
	// ErrParameterParsing is returned when a parameter could not be parsed.
	ErrParameterParsing = errors.New("failed to parse parameter")
	// ErrEvaluationError is returned when a runtime error occurs during CQL evaluation.
	ErrEvaluationError = errors.New("failed during CQL evaluation")
)

// EngineError is returned when the CQL Engine fails during parsing or execution.
type EngineError struct {
	Resource string
	ErrType  EngineErrorType
	Err      error
	// TODO b/338270701 - Add a stack trace to the error for the interpreter.
}

// NewEngineError returns a new EngineError with the given resource, engine error type and error.
// err should be the nested error that was returned during parsing or evaluation.
func NewEngineError(resource string, errType EngineErrorType, err error) EngineError {
	return EngineError{Resource: resource, ErrType: errType, Err: err}
}

func (e EngineError) Error() string {
	return fmt.Sprintf("%s: %s, %s", e.ErrType.Error(), e.Resource, e.Err.Error())
}

func (e EngineError) Unwrap() error {
	return e.Err
}
