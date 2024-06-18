# Experimental CQL Playground

This directory holds a very lightweight, experimental web playground to interact
with our CQL Engine. It supports CQL syntax highlighting and a FHIR bundle data
editor.

<img src="https://github.com/google/cql/assets/6299853/9c6df756-fcaf-4862-acc7-d54f9b265ddd" width=400px/>

__We recommend you avoid putting protected health information (PHI) into this
playground. As always, follow your organization's policies with respect to
PHI.__

## Usage

The CQL playground is a Go server that processes CQL evaluation requests and
serves a very minimal frontend. It is meant to be served locally, and serves
on `http://localhost:8080` out of the box.

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
