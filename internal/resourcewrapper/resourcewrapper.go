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

// Package resourcewrapper provides helper methods to work with R4 FHIR Resources.
package resourcewrapper

import (
	"fmt"

	"github.com/google/fhir/go/protopath"
	r4pb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/resources/bundle_and_contained_resource_go_proto"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// ResourceWrapper holds helper methods to work with R4 FHIR Contained Resources.
type ResourceWrapper struct {
	Resource *r4pb.ContainedResource
}

// New returns a ResourceWrapper that wraps the R4 ContainedResource.
func New(in *r4pb.ContainedResource) *ResourceWrapper {
	return &ResourceWrapper{
		Resource: in,
	}
}

// ResourceType gets the type of the underlying resource or an error.
func (m *ResourceWrapper) ResourceType() (string, error) {
	if m.Resource == nil {
		return "", fmt.Errorf("resource is nil")
	}
	rMsg, err := m.ResourceMessageField()
	if err != nil {
		return "", err
	}
	return string(rMsg.ProtoReflect().Descriptor().Name()), nil
}

// ResourceID gets the ID of the underlying resource or an error.
func (m *ResourceWrapper) ResourceID() (string, error) {
	msg, err := m.ResourceMessageField()
	if err != nil {
		return "", err
	}
	return protopath.Get[string](msg, protopath.NewPath("id.value"))
}

// ResourceMessageField returns the resource from within the ContainedResource.
func (m *ResourceWrapper) ResourceMessageField() (proto.Message, error) {
	if m.Resource == nil {
		return nil, fmt.Errorf("resource is nil")
	}

	rpb := m.Resource.ProtoReflect()
	oneof := rpb.Descriptor().Oneofs().ByName("oneof_resource")
	if oneof == nil {
		return nil, fmt.Errorf("failed to extract oneof")
	}
	fd := rpb.WhichOneof(oneof)
	if fd == nil {
		return nil, fmt.Errorf("no resource type was populated")
	}
	f := rpb.Get(fd)
	innerMsg, ok := f.Interface().(protoreflect.Message)
	if !ok {
		return nil, fmt.Errorf("inner resource is not a message")
	}
	return innerMsg.Interface(), nil
}
