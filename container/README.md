# Cosmos SDK Dependency Injection `container` Module

## Overview

TODO

## Usage

TODO

## Debugging

Issues with resolving dependencies in the container can be done with logs
and [Graphviz](https://graphviz.org) renderings of the container tree. By default, whenever there is an error, logs will
be printed to stderr and a rendering of the dependency graph in Graphviz DOT format will be saved to
`debug_container.dot`.

Here is an example Graphviz rendering of a successful build of a dependency graph:
![Graphviz Example](./testdata/example.svg)

Rectangles represent functions, ovals represent types, rounded rectangles represent modules and the single hexagon
represents the function which called `Build`. Black-colored shapes mark functions and types that were called/resolved
without an error. Gray-colored nodes mark functions and types that could have been called/resolved in the container but
were left unused.

Here is an example Graphviz rendering of a dependency graph build which failed:
![Graphviz Error Example](./testdata/example_error.svg)

Graphviz DOT files can be converted into SVG's for viewing in a web browser using the `dot` command-line tool, ex:
```
> dot -Tsvg debug_container.dot > debug_container.svg
```

Many other tools including some IDEs support working with DOT files.