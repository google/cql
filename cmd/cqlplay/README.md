# Experimental CQL Playground

This directory holds a very lightweight, experimental web playground to interact
with our CQL Engine. It supports CQL syntax highlighting and a FHIR bundle data
editor.

The CQL playground is a Go server that processes CQL evaluation requests and
serves a very minimal frontend. It is meant to be served locally, and serves
on `http://localhost:8080` out of the box.

<img src="https://github.com/google/cql/assets/6299853/9c6df756-fcaf-4862-acc7-d54f9b265ddd" width=400px/>

__We recommend you avoid putting protected health information (PHI) into this
playground. As always, follow your organization's policies with respect to
PHI.__

## Usage

### Run locally

Run the following from the root of the repository (note you must have [Go](https://go.dev/dl/) installed):

```sh
go run cmd/cqlplay/main.go
```

Then in your browser, navigate to http://localhost:8080.

If you'd like to build a binary, you can run:

```sh
go build -o cqlplay cmd/cqlplay/main.go
./cqlplay
```

### Quick Start on GitHub Codespaces
If you want to get up and running quickly without any local setup, you can run this playground on GitHub codespaces.

[![Open in GitHub Codespaces](https://github.com/codespaces/badge.svg)](https://codespaces.new/google/cql?quickstart=1)

Once the codespace starts, simply paste the following into the terminal:

```sh
go run cmd/cqlplay/main.go
```

Once the program is running, you'll see a message pop up in the lower right hand corner with a link to the running playground. Click "Open in Browser" to use the playground. That's it!

<img width="455" alt="image of pop up with link to running application" src="https://github.com/google/cql/assets/6299853/13ab862c-251f-43c0-8ff9-0d3349edb5bf">
