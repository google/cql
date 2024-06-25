<p align="center">
  <h2 align=center>Clinical Quality Language Engine</h2>
  <p align="center">An <b>experimental</b> CQL execution engine for analyzing FHIR healthcare data at
scale</p>
  <p align="center">
    <a href="https://github.com/google/cql/actions"><img src="https://github.com/google/cql/workflows/go_test/badge.svg" alt="GitHub Actions Build Status"/></a>
    <a href="https://godoc.org/github.com/google/cql"><img src="https://godoc.org/github.com/google/cql?status.svg" alt="Go Documentation"/></a>
  </p>
</p>

CQL is a domain specific language designed for querying and executing logic on
healthcare data. CQL excels at healthcare data analysis such as defining quality
measures, clinical decision support, cohorts or preparing data for dashboards.
CQL is designed for healthcare with first class support for terminologies, easy
querying of FHIR via FHIRPath, graceful handling of mixed precision or missing
data and built in clinical helper functions. You can find an intro to the CQL
Language at https://cql.hl7.org.

This repository contains an experimental CQL execution engine in Go, along
with various supporting tools and ecosystem connectors. See the
[Getting Started](#getting-started) section to get up and running quickly.

<div align="center">
  <img width="600" src="https://github.com/google/cql/assets/6299853/f11cbde5-9a44-41ea-847d-1de20e327306"/>
  <p><i>An example CQL snippet.</i></p>
</div>

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
- No support for Interval/List Promotion and Demotion
- No support for related context retrieves
- No support for uncertainties
- No support for importing or exporting ELM
- No support for Quantity unit conversion

## Getting Started

**⚠️ Warning: When using these tools with protected health information (PHI), please be sure
to follow your organization's policies with respect to PHI. ⚠️**

There are several different ways to use this CQL engine, and get up and running quickly.
Click on the links below for details:

* [__Web Playground__](cmd/cqlplay/README.md): A local interactive web playground
  in the browser, to quickly experiment with our CQL engine. Includes CQL syntax
  highlighting and a basic FHIR editor.
* [__CLI__](cmd/cli/README.md): A command-line interface to run this CQL engine
  over small populations or for quick experimentation.
* [__Apache Beam__](beam/README.md): The Beam pipeline is recommended when running CQL over
  large patient populations.
* [__REPL__](cmd/repl/README.md): An interactive command line REPL for quick CQL explorations and experiments.
* [__Golang Module__](https://pkg.go.dev/github.com/google/cql): The CQL execution engine can be
  used as a Go library via the [CQL golang module](https://pkg.go.dev/github.com/google/cql).
  The [Retriever interface](retriever/retriever.go) can be implemented to connect
  to a custom database or FHIR server. The
  [Terminology Provider interface](terminology/provider.go) can be implemented to
  connect to a custom Terminology server.

## Documentation

If you are interested in how this engine was implemented see
[docs/implementation.md](docs/implementation.md) for an overview of the
codebase.

## Contributors

We would like to recognize those who were significant contributors to the
 [initial squash commit](https://github.com/google/cql/commit/bf9849f80b57acea42612a1808d4461bb8412f93) of this project:

**[Kai Bailey](https://github.com/kai-bailey)**, **[Suyash Kumar](https://github.com/suyashkumar)**,  **[Evan Gordon](https://github.com/evan-gordon)**, [Ryan Brush](https://github.com/rbrush),  [Lisa Yin](https://github.com/lisayin), Other Googlers

These are in order of number of changes. Bolded contributors sent
over 100 changes each to the initial commit.

Tech Lead: [Suyash Kumar](https://github.com/suyashkumar) <br />
Product Manager: [Chris Grenz](https://github.com/chrisgrenz) <br />
Eng Manager: [Ed Nanale](https://github.com/enanale)

Contributors since the initial squash commit can be seen in the [contributors tab](https://github.com/google/cql/graphs/contributors).

Thank you to all contributors!
