# CQL CLI

A CLI tool for interacting with the CQL engine to run measures against FHIR
input data to output structured results to disk.

The CLI is primarily designed for executing single patient or small population
workflows. For larger population sets the scalable beam job is a more
appropriate solution.

**Warning: When using these tools with protected health information (PHI),
please be sure to follow your organization's policies with respect to PHI.**

## Running

To build the program from source run the following from the root of the
repository (note you must have [Go](https://go.dev/dl/) installed):

```bash
go build cmd/cli/cli.go
```

This will build the `cli` binary and write it out in your current directory. You
can then run the binary:

```bash
./cli \
  -cql_dir="path/to/cql/dir/" \
  -fhir_bundle_dir="path/to/bundle/dir/" \
  -fhir_terminology_dir="path/to/terminology/dir/" \
  -json_output_dir="path/to/output/"
```

## Flags

**--cql_dir** -- Required. The path to a directory containing one or more CQL
files. The engine only reads files ending in a `.cql` suffix. ELM inputs are
not currently supported.

**--execution_timestamp_override** -- Optional. When set overrides the default
evaluation timestamp for the engine. This can be used to run CQL at a given
point in time. The value should be formatted as a CQL DateTime. If not provided
the engine will use the current system DateTime instead.

Example:

```bash
--execution_timestamp_override="@2018-02-02T15:02:03.000-04:00"
```

**--fhir_bundle_dir** -- Optional. The path containing one or more FHIR bundles.
Each of those bundles will cause one evaluation of the input CQL libraries
results of which will each directly map to outputs.

Note: Each file in the bundle directory is expected to be one bundle per file.

**--fhir_terminology_dir** -- Optional. The path to a directory containing json
definitions of FHIR ValueSets.

**--fhir_parameters_file** -- Optional. A file path to a JSON file containing
FHIR Parameters which will be used as inputs to the CQL execution environment.

**--parameters** -- Optional. A comma separated list of parameters which will be
used as inputs to the CQL execution environment. These values are passed as raw
golang strings to the parsings stage. This provides some limitations for more
complicated input value types. For such cases it is recommended to use the flag
`--fhir_parameters_file` instead.

Example:

```bash
--parameters=”aString='string value',integerValue=2,a id with spaces='value'”
```

Note: If both `--parameters` and `--fhir_parameters_file` are provided. FHIR
Parameters will be loaded first. Any parameters in the `--parameters` flag
will override FHIR Parameters with the same name value.

**--json_output_dir** -- Optional. A directory for outputting structured json
results. Each successful run of an input bundle file will result in one output
result file. If no input bundles are supplied the result will be a single
`result.json` file.

Note: The output json structure is currently a custom format and is subject to
change.

**--return_private_defs** -- Optional. When set will have the CQL engine return
both private and public definitions in the CQL results. By default only public
definitions are emitted.

**-V** -- Optional. Outputs the engine version as well as the CQL version to the
terminal. This flag overrides all other behaviors, so no CQL execution will take
place.
