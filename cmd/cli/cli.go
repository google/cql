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

// A CLI for interacting with the CQL engine.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/cql"
	"github.com/google/cql/internal/datehelpers"
	"github.com/google/cql/internal/iohelpers"
	"github.com/google/cql/result"
	"github.com/google/cql/retriever/local"
	"github.com/google/cql/terminology"
	"github.com/google/fhir/go/fhirversion"
	"github.com/google/fhir/go/jsonformat"
	"github.com/google/bulk_fhir_tools/gcs"
)

type cliConfig struct {
	CQLDir                     string
	ExecutionTimestampOverride string
	FHIRBundleDir              string
	FHIRTerminologyDir         string
	FHIRParametersFile         string
	GCPProject                 string
	Parameters                 string
	ReturnPrivateDefs          bool
	JSONOutputDir              string
	Version                    bool

	// Should not be set directly by a flag.
	gcsEndpoint string
}

func (cfg *cliConfig) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&cfg.CQLDir, "cql_dir", "", "(Required) Directory holding 1 or more CQL files.")
	fs.StringVar(
		&cfg.ExecutionTimestampOverride,
		"execution_timestamp_override",
		"",
		"(Optional) A DateTime to use for overriding the default execution timestamp of the CQL engine. The value of should match the format of a CQL DateTime. If the value provided doesn't contain a timezone utc the default will be UTC. If not supplied the engine will use the current DateTime. Example: @2024-01-01T00:00:00Z",
	)
	fs.StringVar(&cfg.FHIRBundleDir, "fhir_bundle_dir", "", "(Optional) Directory holding FHIR Bundle JSON files.")
	fs.StringVar(&cfg.FHIRTerminologyDir, "fhir_terminology_dir", "", "(Optional) Directory holding FHIR Valueset JSONs.")
	fs.StringVar(&cfg.FHIRParametersFile, "fhir_parameters_file", "", "(Optional) A JSON file holding FHIR Parameters to use during CQL execution. Currently only supports R4.")
	fs.StringVar(&cfg.Parameters, "parameters", "", "(Optional) A comma separated list of parameters to pass to the CQL execution. Example: --parameters=\"aString='string value',integerValue=2\"")
	fs.StringVar(&cfg.GCPProject, "gcp_project", "", "(Optional) The GCP project to use when reading from or writing to GCS.")

	// Output flags.
	fs.BoolVar(&cfg.ReturnPrivateDefs, "return_private_defs", false, "(Optional) If true, will include the output of all private CQL expression definitions. By default only public definitions are outputted. This should only be used for debugging purposes.")
	fs.StringVar(&cfg.JSONOutputDir, "json_output_dir", "", "(Optional) Directory in which to output each evaluation result as a JSON file. If not supplied will output in the current directory.")

	// See: https://cql.hl7.org/history.html for CQL versions.
	fs.BoolVar(&cfg.Version, "V", false, "(Optional) Prints the current version of the CQL engine and CQL version.")

	// Set the default gcs endpoint to the public endpoint, only override if running in a test
	// environment.
	cfg.gcsEndpoint = gcs.DefaultCloudStorageEndpoint
}

const usageMessage = "The CLI for the golang CQL engine."

var errMissingFlag = errors.New("missing required flag")

// The config which is populated by the CLI input flags.
var config cliConfig

func init() {
	config.RegisterFlags(flag.CommandLine)
	defaultUsage := flag.Usage
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, usageMessage)
		defaultUsage()
	}
}

func main() {
	flag.Parse()
	ctx := context.Background()
	if err := mainWrapper(ctx, config); err != nil {
		log.Fatalf("CQL CLI failed with an error: %v", err)
	}
}

func validateConfig(ctx context.Context, cfg *cliConfig) error {
	if cfg.CQLDir == "" {
		return fmt.Errorf("%w --cql_dir", errMissingFlag)
	}
	err := validatePath(ctx, cfg.CQLDir, cfg.GCPProject, cfg.gcsEndpoint, "cql_dir")
	if err != nil {
		return err
	}
	if cfg.FHIRBundleDir != "" {
		err := validatePath(ctx, cfg.FHIRBundleDir, cfg.GCPProject, cfg.gcsEndpoint, "fhir_bundle_dir")
		if err != nil {
			return err
		}
	}
	if cfg.FHIRTerminologyDir != "" {
		err := validatePath(ctx, cfg.FHIRTerminologyDir, cfg.GCPProject, cfg.gcsEndpoint, "fhir_terminology_dir")
		if err != nil {
			return err
		}
	}
	if cfg.JSONOutputDir != "" {
		err := validatePath(ctx, cfg.JSONOutputDir, cfg.GCPProject, cfg.gcsEndpoint, "json_output_dir")
		if err != nil {
			return err
		}
	}
	return nil
}

// validatePath validates that the path is a valid GCS path or a local file path.
// flagName is the name of the flag that was used to pass in the path that is being validated.
func validatePath(ctx context.Context, validationPath, gcpProject, gcsEndpoint, flagName string) error {
	if strings.HasPrefix(validationPath, "gs://") {
		bucket, _, err := gcs.PathComponents(validationPath)
		if err != nil {
			return err
		}
		if err := validateGCSBucketInProject(ctx, bucket, gcpProject, gcsEndpoint); err != nil {
			return err
		}
	} else if _, err := os.Stat(validationPath); err != nil {
		return fmt.Errorf("--%s: %w", flagName, err)
	}
	return nil
}

func validateGCSBucketInProject(ctx context.Context, bucket, project, endpoint string) error {
	// Only allow writing to and reading from GCS buckets in the same project.
	// We do this to prevent cases where a user could mis-type a bucket name and end up writing PHI
	// data to a bucket in a different project.
	if project == "" {
		return fmt.Errorf("--gcp_project must be set if you are using a GCS file IO")
	}
	c, err := gcs.NewClient(ctx, bucket, endpoint)
	if err != nil {
		return err
	}
	isInProject, err := c.IsBucketInProject(ctx, project)
	if err != nil {
		return err
	}
	if !isInProject {
		return fmt.Errorf("could not find GCS Bucket %s in the GCP project %s", bucket, project)
	}
	return nil
}

// TODO: b/340361303 - Consider using something like the logger interface in medical claims tools before open sourcing.
func mainWrapper(ctx context.Context, cfg cliConfig) error {
	if cfg.Version {
		fmt.Println("CQL Engine Version: (Beta) 0.0.1, CQL Version: 1.5.2")
		return nil
	}
	if err := validateConfig(ctx, &cfg); err != nil {
		return err
	}
	cqlLibs, err := readCQLLibs(ctx, cfg.CQLDir, &cfg)
	if err != nil {
		return fmt.Errorf("failed to read CQL libraries: %w", err)
	}
	fhirDM, err := cql.FHIRDataModel("4.0.1")
	if err != nil {
		return fmt.Errorf("failed to create FHIR data model: %w", err)
	}
	config := cql.ParseConfig{DataModels: [][]byte{fhirDM}}
	if cfg.FHIRParametersFile != "" {
		parametersText, err := iohelpers.ReadFile(ctx, cfg.FHIRParametersFile, &iohelpers.IOConfig{GCSEndpoint: cfg.gcsEndpoint})
		if err != nil {
			return fmt.Errorf("failed to read FHIR parameters file %s: %w", cfg.FHIRParametersFile, err)
		}
		params, err := parseFHIRParameters(parametersText)
		if err != nil {
			return fmt.Errorf("failed to parse FHIR parameters file %s: %w", cfg.FHIRParametersFile, err)
		}
		config.Parameters = params
	}
	if cfg.Parameters != "" {
		for _, param := range strings.Split(cfg.Parameters, ",") {
			parts := strings.Split(param, "=")
			if len(parts) != 2 {
				return fmt.Errorf("--parameters was passed an invalid input string: %s", param)
			}
			config.Parameters[result.DefKey{Name: parts[0]}] = parts[1]
		}
	}
	elm, err := cql.Parse(ctx, cqlLibs, config)
	if err != nil {
		return fmt.Errorf("failed to parse CQL: %w", err)
	}
	tp, err := maybeGetTerminologyProvider(ctx, cfg.FHIRTerminologyDir, &cfg)
	if err != nil {
		return fmt.Errorf("failed to get terminology: %w", err)
	}

	evalConfig := cql.EvalConfig{
		ReturnPrivateDefs: cfg.ReturnPrivateDefs,
		Terminology:       tp,
	}
	if cfg.ExecutionTimestampOverride != "" {
		t, _, err := datehelpers.ParseDateTime(cfg.ExecutionTimestampOverride, time.UTC)
		if err != nil {
			return fmt.Errorf("failed to parse execution timestamp override to a valid DateTime value: %w", err)
		}
		evalConfig.EvaluationTimestamp = t
	}
	if err = runCQLWithBundleDir(ctx, elm, cfg.FHIRBundleDir, cfg.JSONOutputDir, evalConfig, &cfg); err != nil {
		return fmt.Errorf("failed to run CQL: %w", err)
	}
	return nil
}

type cqlResult struct {
	BundleSource string           `json:"bundleSource,omitempty"`
	EvalResults  result.Libraries `json:"evalResults"`
}

func runCQLWithBundleDir(ctx context.Context, elm *cql.ELM, fhirBundleDir string, outputDir string, evalConfig cql.EvalConfig, cfg *cliConfig) error {
	// If fhirBundleDir is empty run one eval with empty bundle retriever.
	if fhirBundleDir == "" {
		r, err := elm.Eval(ctx, &local.Retriever{}, evalConfig)
		if err != nil {
			return err
		}
		return outputCQLResults(ctx, outputDir, "results.json", cqlResult{EvalResults: r}, cfg)
	}

	bundleFilePaths, err := iohelpers.FilesWithSuffix(ctx, fhirBundleDir, ".json", &iohelpers.IOConfig{GCSEndpoint: cfg.gcsEndpoint})
	if err != nil {
		return err
	}
	if len(bundleFilePaths) == 0 {
		fmt.Printf("no files found in FHIR bundle directory %s, exiting", fhirBundleDir)
		return nil
	}

	// TODO(b/301659936): implement a concurrent version of this.
	for _, filePath := range bundleFilePaths {
		fhirData, err := iohelpers.ReadFile(ctx, filePath, &iohelpers.IOConfig{GCSEndpoint: cfg.gcsEndpoint})
		if err != nil {
			return err
		}
		ret, err := local.NewRetrieverFromR4Bundle(fhirData)
		if err != nil {
			return err
		}
		r, err := elm.Eval(ctx, ret, evalConfig)
		if err != nil {
			return err
		}
		_, fileName := filepath.Split(filePath)
		if err := outputCQLResults(ctx, outputDir, fileName, cqlResult{BundleSource: filePath, EvalResults: r}, cfg); err != nil {
			return err
		}
	}
	return nil
}

// maybeGetTerminologyProvider constructs a ValueSet terminology provider if provided with a valid directory.
func maybeGetTerminologyProvider(ctx context.Context, terminologyDir string, cfg *cliConfig) (terminology.Provider, error) {
	if terminologyDir == "" {
		return nil, nil
	}
	filePaths, err := iohelpers.FilesWithSuffix(ctx, terminologyDir, ".json", &iohelpers.IOConfig{GCSEndpoint: cfg.gcsEndpoint})
	if err != nil {
		return nil, err
	}

	var jsonTerminologyData []string
	for _, filePath := range filePaths {
		b, err := iohelpers.ReadFile(ctx, filePath, &iohelpers.IOConfig{GCSEndpoint: cfg.gcsEndpoint})
		if err != nil {
			return nil, err
		}
		jsonTerminologyData = append(jsonTerminologyData, string(b))
	}
	return terminology.NewInMemoryFHIRProvider(jsonTerminologyData)
}

// readCQLLibs reads all CQL (files containing the .cql suffix) files from a directory.
func readCQLLibs(ctx context.Context, dir string, cfg *cliConfig) ([]string, error) {
	filePaths, err := iohelpers.FilesWithSuffix(ctx, dir, ".cql", &iohelpers.IOConfig{GCSEndpoint: cfg.gcsEndpoint})
	if err != nil {
		return nil, err
	}

	cqlLibs := make([]string, 0, len(filePaths))
	for _, filePath := range filePaths {
		var cqlData []byte
		cqlData, err = iohelpers.ReadFile(ctx, filePath, &iohelpers.IOConfig{GCSEndpoint: cfg.gcsEndpoint})
		if err != nil {
			return nil, fmt.Errorf("failed to read CQL file %s: %w", filePath, err)
		}
		cqlLibs = append(cqlLibs, string(cqlData))
	}
	return cqlLibs, nil
}

// parseFHIRParameters takes in JSON bytes containing FHIR parameters and converts them to parameter
// name to value mappings.
// Like other parts of the engine currently only supports FHIR R4.
func parseFHIRParameters(parametersBytes []byte) (map[result.DefKey]string, error) {
	unmarshaller, err := jsonformat.NewUnmarshallerWithoutValidation("UTC", fhirversion.R4)
	if err != nil {
		return nil, err
	}
	containedResource, err := unmarshaller.UnmarshalR4(parametersBytes)
	if err != nil {
		return nil, err
	}
	params := containedResource.GetParameters()
	if params == nil {
		return nil, fmt.Errorf("failed to parse FHIR parameters with text: %s", string(parametersBytes))
	}

	// For now we are doing a simple conversion for a subset of the types.
	// We take those values and convert them to strings which can be parsed by the parser into CQL
	// values. Later we will need to create a parser for FHIR Parameters to/from CQL values.
	// TODO b/337956267 - Create a parser for FHIR Parameters to/from CQL values.
	r := map[result.DefKey]string{}
	for _, param := range params.GetParameter() {
		choice := param.GetValue()
		name := param.GetName().GetValue()
		choice.GetQuantity()

		var val string
		if c := choice.GetStringValue(); c != nil {
			// Since we need to be able to parse these out into a CQL literal, internally we wrap the
			// raw value of FHIR Strings in single quotes.
			val = fmt.Sprintf("'%s'", c.GetValue())
		} else if c := choice.GetBoolean(); c != nil {
			val = strconv.FormatBool(c.GetValue())
		} else if c := choice.GetCode(); c != nil {
			val = c.GetValue()
		} else if c := choice.GetDecimal(); c != nil {
			val = c.GetValue()
		} else if c := choice.GetInteger(); c != nil {
			val = strconv.FormatInt(int64(c.GetValue()), 10)
		} else if c := choice.GetPositiveInt(); c != nil {
			val = strconv.FormatInt(int64(c.GetValue()), 10)
		} else if c := choice.GetUnsignedInt(); c != nil {
			val = strconv.FormatInt(int64(c.GetValue()), 10)
		} else {
			return nil, fmt.Errorf("unsupported FHIR Parameter %s was not of type (string, boolean, code, decimal, integer, positiveInt, or unsignedInt)", name)
		}
		r[result.DefKey{Name: name}] = val
	}
	return r, nil
}

func outputCQLResults(ctx context.Context, path, fileName string, results cqlResult, cfg *cliConfig) error {
	// need to update this for different output types
	jsonResults, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	return iohelpers.WriteFile(ctx, path, fileName, jsonResults, &iohelpers.IOConfig{GCSEndpoint: cfg.gcsEndpoint})
}
