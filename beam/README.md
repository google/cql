# CQL on Beam

The CQL on Beam pipeline is designed for running CQL on large patient
populations. Apache Beam is an open source, unified model for defining both
batch and streaming pipelines. Using the Apache Beam SDKs, you build a program
that defines the pipeline. Then, you execute the pipeline on a specific platform
such as Google Cloud's Dataflow. Apache Beam insulates you from the low-level
details of distributed processing, such as coordinating individual workers,
sharding datasets, and other such tasks. See these
[docs](https://cloud.google.com/dataflow/docs/concepts/beam-programming-model)
for more information on Apache Beam.

**Warning: When using these tools with protected health information (PHI),
please be sure to follow your organization's policies with respect to PHI.**

**Note: There's currently an issue with Beam tests on Windows. Until this is
resolved it is not advised to use the beam job on Windows. See
https://github.com/google/cql/issues/32 for further context.**

## Running

The CQL on Beam pipeline is currently limited to local file system IO. The
pipeline can read FHIR bundles and output NDJSON CQL results to the file system.
Future work will add more IO options, namely reading from FHIR Store and
outputting to BigQuery. Once those IOs are complete the CQL on Beam pipeline can
be run on [Google Cloud's Dataflow](https://cloud.google.com/dataflow/docs/quickstarts/create-pipeline-go).

To build the program from source run the following from the root of the
repository (note you must have [Go](https://go.dev/dl/) installed):

```bash
cd beam
go build -o beam main.go
```

This will build the `beam` binary and write it out in your current directory. You
can then run the binary:

```bash
./beam \
  -cql_dir="path/to/cql/dir/" \
  -fhir_bundle_dir="path/to/bundle/dir/" \
  -fhir_terminology_dir="path/to/terminology/dir/" \
  -ndjson_output_dir="path/to/output/"
```

## Flags

**--cql_dir** Required. The path to a directory containing one or more CQL
files. The engine only reads files ending in a `.cql` suffix. ELM inputs are
not currently supported.

**--evaluation_timestamp** The timestamp to use for evaluating CQL. If not
provided EvaluationTimestamp will default to time.Now() called at the start of
the pipeline.

Example:

```bash
--evaluation_timestamp="@2018-02-02T15:02:03.000-04:00"
```

**--fhir_bundle_dir** Required. The path containing one or more FHIR bundles.
Each file should have one FHIR Bundle containing all of the FHIR resources for a
particular patient.

**--fhir_terminology_dir** Optional. The path to a directory containing json
definitions of FHIR ValueSets.

**--ndjson_output_dir** Required. Output directory that the CQL results will be
written to. The results for each patient are converted to JSON and written as a
line in the NDJSON.

**--return_private_defs** If true will include the output of all private CQL
expression definitions. By default only public definitions are outputted.






