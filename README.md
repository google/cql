<p align="center">
  <h2 align=center>Clinical Quality Language Engine</h2>
  <p align="center">A experimental CQL execution engine for analyzing FHIR healthcare data at
scale<p>
  <p align="center"> 
  <a href="https://godoc.org/github.com/google/cql">
    <img src="https://godoc.org/github.com/google/cql?status.svg" alt="Go Documentation" />
  </a>
</p>



CQL is a domain specific language designed for querying and executing logic on
healthcare data. CQL excels at healthcare data analysis such as defining quality
measures, clinical decision support, cohorts or preparing data for dashboards.
CQL is designed for healthcare with first class support for terminologies, easy
querying of FHIR via FHIRPath, graceful handling of mixed precision or missing
data and built in clinical helper functions. You can find an intro to the CQL
Language at https://cql.hl7.org.

## Features

Notable features of this engine include:

- Built in explainability - for each CQL expression definition we produce a tree
that traces through the data and expressions involved in calculating the final
result
- Built in custom CQL parser, reducing project dependencies and allowing
optimizations between the parser and interpreter
- Benchmarked and optimized to be fast and memory efficient

In addition to the engine this repository has several tools that make it easy to
launch and productionize CQL:

- A scalable Beam job for running CQL over large patient populations
- A CLI for easy configuration
- Integration into Google Cloud Storage

## Limitations

This CQL Engine is **experimental** and not an officially supported Google
Product. The API to call the engine and the format of the results returned are
subject to change. There is limited support of certain parts of the CQL
Language:

- Only the FHIR version 4.0.1 data model is supported
- Only the Patient Context is supported
- Not all system operators are supported
- No support for Quantities with UCUM units
- No support for Interval/List Promotion and Demotion
- No support for related context retrieves
- No support for uncertainties
- No support for importing or exporting ELM

## Getting Started

There are several different ways to use this engine. For quick experimentation
we have a Command Line Interface and REPL. For executing CQL over a large
patient population there is a Beam job. Finally, the CQL Golang Module allows
you to execute CQL by implementing your own connector to a database and
terminology server.

> ⚠️⚠️  **Warning**  ⚠️⚠️
>
> When using these tools with protected health information (PHI), please be sure
to follow your organization's policies with respect to PHI.

## CLI

When intending to run the CQL engine locally over small populations or for quick
experimentation use the CLI located at [cmd/cli](cmd/cli). For documentation
and examples see [cmd/cli/README.md](cmd/cli/README.md).

## Apache Beam Pipeline

The Beam pipeline is recommended when running CQL over large patient populations.
More information and usage examples are documented at
[beam/README.md](beam/README.md).

## Golang Module

The engine can be used via the CQL golang module documented in the
[godoc](https://pkg.go.dev/github.com/google/cql).
The [Retriever interface](retriever/retriever.go) can be implemented to connect
to a custom database or FHIR server. The
[Terminology Provider interface](terminology/provider.go) can be implemented to
connect to a custom Terminology server.

## REPL

For quick experiments with our CQL Engine we have a REPL. More information and
usage examples are documented at [cmd/repl/README.md](cmd/repl/README.md).

## Documentation

If you are interested in how this engine was implemented see
[docs/implementation.md](docs/implementation.md) for an overview of the
codebase.
