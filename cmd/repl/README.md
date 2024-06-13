# CQL REPL

A REPL CLI for interacting with the CQL engine. This can be a useful tool for quickly iterating while authoring CQL measures.

## Running

To build the program from source run the following from the root of the
repository (note you must have [Go](https://go.dev/dl/) installed):

```bash
go build cmd/repl/repl.go
```

This will build the `repl` binary and write it out in your current directory.
You can then run the binary:

```bash
./repl
```

* If an invalid statement is input, inputting an empty line clears the current
cache
* Typing in `exit` will quit the program.
* To continue a statement across multiple lines type a backslash (\) character
before submitting that line

ex:
```
define multiline_example: \
  @2014 + 1 year
```

Note: While writing multiple statements this way is technically supported
the output ordering of the results is currently undefined.

You may also pass in additional flags to `repl` in order to load some FHIR data
or load an external input CQL library, for example:

```bash
./repl \
  -cql_file="path/to/my/file.cql" \
  -bundle_file="path/to/my/patient_or_resource.json"
```

See the Flags section below for more.

## Flags

**bundle_file** -- Optional

The path to a single bundle FHIR resource (often a patient). This is the
resource file the cql will evaluate against for all executions in the
REPL insance.

**cql_file** -- Optional

The path to a CQL file. The contents of this file are seeded into the current
context and outputs are displayed before the execution of the first REPL loop.

**valuesets_dir** -- Optional

The path to a directory containing json definitions of FHIR valuesets.

## Future Enhancements

* Right now the repl can't tell if an expression is unfinished or incorrect.
`define test:` is an unfinished statement that would be nice if we could
handle separately from invalid statements.
* Ability to run against multiple bundle files. This would require some work
to ensure the output for this were readable.
* Ability to clear the current context and load a different bundle resource
to evaluate against.
