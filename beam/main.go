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

// Beam pipeline for computing CQL at scale.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/golang/glog"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/x/beamx"
	"github.com/google/cql/beam/transforms"
	"github.com/apache/beam/sdks/v2/go/pkg/beam"

	// The following import is required for accessing local files.
	"github.com/apache/beam/sdks/v2/go/pkg/beam/io/fileio"
	_ "github.com/apache/beam/sdks/v2/go/pkg/beam/io/filesystem/local"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/io/textio"
)

// TODO(b/317813865): Add input and output options as needed, such as FHIR Store or NDJSON inputs
// and BigQuery outputs.

// flags holds the values of the flags largely to assist in easier testing without having to change
// global variables.
type beamFlags struct {
	CQLDir             string
	FHIRBundleDir      string
	FHIRTerminologyDir string
	EvaluationTimestamp string
	ReturnPrivateDefs   bool
	NDJSONOutputDir     string
}

var flags beamFlags

func init() {
	flag.StringVar(&flags.CQLDir, "cql_dir", "", "(Required) Directory holding one or more CQL files.")
	flag.StringVar(&flags.FHIRBundleDir, "fhir_bundle_dir", "", "(Required) Directory holding FHIR Bundle JSON files, which are used to create a retriever for the CQL engine.")
	flag.StringVar(&flags.FHIRTerminologyDir, "fhir_terminology_dir", "", "(Optional) Directory holding FHIR Valueset JSONs, which are used to create a terminology provider for the CQL engine.")
	flag.StringVar(&flags.EvaluationTimestamp, "evaluation_timestamp", "", "(Optional) The timestamp to use for evaluating CQL. If not provided EvaluationTimestamp will default to time.Now() called at the start of the eval request.")
	flag.BoolVar(&flags.ReturnPrivateDefs, "return_private_defs", false, "(Optional) If true will include the output of all private CQL expression definitions. By default only public definitions are outputted.")
	// TODO b/339070720: Add CQL parameters.
	flag.StringVar(&flags.NDJSONOutputDir, "ndjson_output_dir", "", "(Required) Output directory that the NDJSON files will be written to.")
}

// pipelineConfig holds the validated configuration for the pipeline.
type pipelineConfig struct {
	// TODO: b/339070720 - Instead of parsing on each worker, if we could serialize the cql.ELM struct
	// we could parse once before execution and pass it to each worker.
	CQL                 []string
	FHIRBundleDir       string
	ValueSets           []string
	EvaluationTimestamp time.Time
	ReturnPrivateDefs bool
	NDJSONOutputDir   string
}

func buildPipelineConfig(flags *beamFlags) (*pipelineConfig, error) {
	if flags == nil {
		return nil, fmt.Errorf("flags must not be nil")
	}

	cfg := &pipelineConfig{
		FHIRBundleDir:     flags.FHIRBundleDir,
		ReturnPrivateDefs: flags.ReturnPrivateDefs,
		NDJSONOutputDir:   flags.NDJSONOutputDir,
	}

		if flags.EvaluationTimestamp != "" {
			var err error
			cfg.EvaluationTimestamp, err = time.Parse(time.RFC3339, flags.EvaluationTimestamp)
			if err != nil {
				return nil, fmt.Errorf("evaluation_timestamp must be in RFC3339 format: %v", err)
			}
		} else {
			cfg.EvaluationTimestamp = time.Now()
		}

	if flags.CQLDir == "" {
		return nil, fmt.Errorf("cql_dir must be set")
	}
	if flags.FHIRBundleDir == "" {
		return nil, fmt.Errorf("fhir_bundle_dir must be set")
	}
	if flags.NDJSONOutputDir == "" {
		return nil, fmt.Errorf("ndjson_output_dir must be set")
	}

	var err error
	cfg.CQL, err = readFilesWithSuffix(flags.CQLDir, ".cql")
	if err != nil {
		return nil, err
	}
	if len(cfg.CQL) == 0 {
		return nil, fmt.Errorf("must be at least one CQL file")
	}

	cfg.ValueSets, err = readFilesWithSuffix(flags.FHIRTerminologyDir, ".json")
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// readFilesWithSuffix reads all files from a directory with the given suffix.
func readFilesWithSuffix(dir, allowedFileSuffix string) ([]string, error) {
	if dir == "" {
		return nil, nil
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}
	strs := make([]string, 0, len(files))
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), allowedFileSuffix) {
			continue
		}
		bytes, err := os.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", filepath.Join(dir, file.Name()), err)
		}
		strs = append(strs, string(bytes))
	}
	return strs, nil
}

// buildPipeline uses the config to construct the pipeline. Results and errors are returned for
// tests.
func buildPipeline(s beam.Scope, cfg *pipelineConfig) (results, errors beam.PCollection) {
	matches := fileio.MatchFiles(s, filepath.Join(cfg.FHIRBundleDir, "*.json"))
	files := fileio.ReadMatches(s, matches)
	bundles, loadErrors := beam.ParDo2(s, transforms.FileToBundle, files)

	var evalErrors beam.PCollection
		fn := &transforms.CQLEvalFn{
			CQL:                 cfg.CQL,
			ValueSets:           cfg.ValueSets,
			EvaluationTimestamp: cfg.EvaluationTimestamp,
			ReturnPrivateDefs:   cfg.ReturnPrivateDefs,
		}
		results, evalErrors = beam.ParDo2(s, fn, bundles)

	ndjsonRows, writeErrors := beam.ParDo2(s, transforms.NDJSONSink, results)
	// TODO: b/339070720: Shard the output files.
	textio.Write(s, filepath.Join(cfg.NDJSONOutputDir, "results.ndjson"), ndjsonRows)

	errors = beam.Flatten(s, loadErrors, evalErrors, writeErrors)
	errorRows := beam.ParDo(s, transforms.ErrorsNDJSONSink, errors)
	textio.Write(s, filepath.Join(cfg.NDJSONOutputDir, "errors.ndjson"), errorRows)

	return results, errors
}

func main() {
	flag.Parse()
	beam.Init()

	cfg, err := buildPipelineConfig(&flags)
	if err != nil {
		log.Exit(err)
	}

	p, s := beam.NewPipelineWithRoot()
	_, _ = buildPipeline(s, cfg)

	if err := beamx.Run(context.Background(), p); err != nil {
		log.Exitf("Failed to execute job: %v", err)
	}
}
