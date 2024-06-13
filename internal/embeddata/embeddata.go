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

// Package embeddata holds embedded data required by the production CQL engine. Test data embeds
// can be found in the testdata package.
package embeddata

import "embed"

// ModelInfos contain embedded model info files, specifically fhir-modelinfo-4.0.1.xml and
// system-modelinfo.xml.
//
//go:embed third_party/cqframework/fhir-modelinfo-4.0.1.xml third_party/cqframework/system-modelinfo.xml
var ModelInfos embed.FS

// FHIRHelpers contains the embedded FHIRHelpers-4.0.1.cql file.
//
//go:embed third_party/cqframework/FHIRHelpers-4.0.1.cql
var FHIRHelpers embed.FS
