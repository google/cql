# Engine Tests

Engine Tests holds most of the "unit tests" for the CQL Engine parser and interpreter. The problem with writing typical unit tests for the parser (CQL --> model) and interpreter (model --> result) is the model input to the interpreter tests. When writing hundreds of interpreter unit tests it is easy for developers to miss bugs by writing an invalid model input or a model input that does not match what our parser produces. Instead we decided to have integration tests (CQL --> model --> result) as our primary suite of unit tests. This way the unit tests use a model that our parser actually produces, even as we change and update the parser.

Although the engine tests are end to end, they should target as specific a behaviour as possible just like a typical unit test. For testing large, realistic CQL expressions use the large tests. We still have some parser (CQL --> model) and interpreter (model --> result) unit tests. For example, if a developer wants to tests a model input to the interpreter that our parser will not produce they should write a typical interpreter (model --> result) unit test. But the majority of interpreter test should be written as engine tests to avoid incorrect or out of date models.

The cons of this approach are:

* Typical Golang unit tests are stored in the file next to the code they are testing (foo.go, foo_test.go). This approach breaks that pattern.

* It encourages coupling of our parser to our interpreter, and may mean our interpreter less robust to inputs from other parsers. For now we are treating our interpreter as an internal detail not meant to be called directly.
