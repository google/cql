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

package terminology

// Code represents a CQL or ELM "Code", which is equivalent to a FHIR Coding.
type Code struct {
	// Code is the code value. This is repetitive, but matches the ELM/CQL and FHIR naming.
	Code string `json:"code"`
	// System is the coding system id.
	System string `json:"system"`
	// Display is an optional display string that represents this code.
	Display string `json:"display"`
}

// key returns the codingKey that uniquely identifies this Code.
func (c *Code) key() codeKey {
	return codeKey{Value: c.Code, System: c.System}
}

// codeKey contains code value and coding system information that uniquely identifies a Code.
type codeKey struct {
	// Value is the code value of this code.
	Value string
	// System is the coding system id.
	System string
}
