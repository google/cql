# How to contribute

We'd love to accept your patches and contributions to this project.

## Before you begin

### Sign our Contributor License Agreement

Contributions to this project must be accompanied by a
[Contributor License Agreement](https://cla.developers.google.com/about) (CLA).
You (or your employer) retain the copyright to your contribution; this simply
gives us permission to use and redistribute your contributions as part of the
project.

If you or your current employer have already signed the Google CLA (even if it
was for a different project), you probably don't need to do it again.

Visit <https://cla.developers.google.com/> to see your current agreements or to
sign a new one.

### Review our community guidelines

This project follows
[Google's Open Source Community Guidelines](https://opensource.google/conduct/).

## Contribution process

### Code reviews

All submissions, including submissions by project members, require review. We
use GitHub pull requests for this purpose. Consult
[GitHub Help](https://help.github.com/articles/about-pull-requests/) for more
information on using pull requests.

## Getting Started

Before you begin, it may be beneficial to check out
[docs/implementation.md](docs/implementation.md). This document should give you
a basic understanding of the different structures that make up the engine, and
how language pieces are defined.

Another good reference is
[https://cql.hl7.org/09-b-cqlreference.html](https://cql.hl7.org/09-b-cqlreference.html)
for getting a good understanding of what behaviors a feature should implement.

### Implementing Your First Issue

A good place to start in the CQL engine repository is by implementing additional
system operators. Operators we do not yet support can be found in the
[exclusions list](tests/spectests/exclusions/exclusions.go) for the CQL
language specification tests. You may also want to check the repository issues
for any particular feature requests or bugs as well.

The [exclusions list](tests/spectests/exclusions/exclusions.go) contains
skip definitions by both test group and test name. Test names marked explicitly
with TODOs may be a great place to start.

### Implementing New Parser Functionality

For the feature that you have selected to work on, you will likely need to start
by implementing some logic to convert parsed ANTLR grammar into an internal
representation.

The [CQL ANTLR Grammar](internal/embeddata/third_party/cqframework/Cql.g4) represents the CQL syntax, which is used to compile CQL into [Go ANTLR nodes](internal/embeddata/third_party/cqframework/cql/).
The parser traverses these ANTLR nodes to build intermediate model structures
for the interpreter to process.

The engine model structures are based on the
[ELM definitions](https://cql.hl7.org/elm.html), and represent compiled CQL.
These models are defined in [model/model.go](model/model.go) while the parsings
of the models is done inside [parser/expressions.go](parser/expressions.go).

Most operators will also need a functional definition mapped to the model as
well, this is done inside [parser/operator.go](parser/operator.go).

### Implementing in the Interpreter

Inside the interpreter you will need to do two things, add a function that takes
in the model and any CQL operands, and map the model to that function in the
operator dispatcher.

For the model to interpreter logic mapping navigate to
[interpreter/operator_dispatcher.go](interpreter/operator_dispatcher.go).
From there you can declare a new interpreter logic function in the relevant
[interpreter/](interpreter/) file.

Please make sure to follow good code hygiene practices and write tests for any
code you write!

### Submitting

Once you are confident in your work feel free to open a pull request against the
repo. Make sure to write a detailed commit message describing what your changes
fix.
