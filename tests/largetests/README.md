# CQL Engine Large End-to-End Tests

These tests are set up to run like fixture or screenshot tests, where we want coverage and regression detection over large test outputs. For example, all of the expression definition results of a large CQL measure.

To add a new test case:

1. Add your CQL, FHIR Bundle, and Valueset files to the `tests/` directory in
   the pattern laid out below.
2. Update `buildAllTestCases` in large_test.go to add your test case, for example:

```
{
  Name:       "example - Most recent systolic bp with a valid status",
  CQLFile:    "tests/example/main.cql",
  BundleFile: "tests/example/data.json",
  WantFile:   "tests/example/output.json",
},
```
3. The first time the test is run it will fail and print out the interpreter output in json. The printed output can be copied and pasted into `output.json` and the test should pass.

Simply running `TestLarge` will run all of the tests, the same as how normal Go tests operate.

Preferred test data layout:

* `valuesets/`: add your Valueset JSONs here, one JSON file per valueset. All
  valuesets are always loaded for every test.
* `tests/`
  * `your_test_name/`
      * `main.cql`: the main CQL to be executed. Multiple libraries not supported
      yet.
      * `data.json`: A FHIR bundle of the available resources for the test
      execution.
      * `output.json`: The expected output of evaluating `main.cql` over
      `data.json`.
  * `your_test_with_subtests/`: if you want to run single CQL over multiple data
    bundles:
      * `main.cql`
      * `subtest1/`
          * `data.json`
          * `output.json`
