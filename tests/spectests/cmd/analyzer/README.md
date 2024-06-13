# CQL Test Analyzer

A simple CLI tool for analyzing the current coverage of the golang CQL engine
of the external XML tests.

// TODO: b/346997754 - set this up as a github action.

## Build & Run

```
go build cql/tests/spectests/cmd/analyzer/analyzer.go

./analyzer
```

This will output totals and per stat coverage of the XML stats.