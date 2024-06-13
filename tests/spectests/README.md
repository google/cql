# XML tests

Uses XML test files imported from https://github.com/cqframework/cql-tests.

XML tests constructs a CQL expression for each test in the following manner from
the input XML file:

```
define <group.Name>_<test.Name>:
  (<expression>) ~ <output>
```

The CQL strings are then evaluated and checked to see if they return `true`.
If `true`, then the test passes otherwise the test fails.

The input XML file is expected to match the specifications defined in
 `cql/tests/spectests/cqltests/testSchema.xsd`.

## How to generate `model.go`

* Run
```
git clone https://github.com/GoComply/xsd2go.git
```
* Change the directory
```
cd go_modules/xsd2go/cli
```
* Run
```
go run ./gocomply_xsd2go convert path/to/testSchema.xsd ./scap path/to/output/dir
```

## Coverage Stats

To output coverage stats of these XML tests navigate to
[cmd/analyzer](cmd/analyzer) for a CLI analysis tool.