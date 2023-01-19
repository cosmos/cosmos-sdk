---
sidebar_position: 1
---

# Depinject

> **DISCLAIMER**: This is a **beta** package. The SDK team is actively working on this feature and we are looking for feedback from the community. Please try it out and let us know what you think.

## Overview

`depinject` is a dependency injection framework for the Cosmos SDK. This module together with `core/appconfig` are meant to simplify the definition of a blockchain by replacing most of `app.go`'s boilerplate code with a configuration file (Go, YAML or JSON).

* [Go Doc](https://pkg.go.dev/cosmossdk.io/depinject)

## Usage

`depinject` includes an expressive and composable [Configuration API](https://pkg.go.dev/github.com/cosmos/cosmos-sdk/depinject#Config).
A core configuration function is `Provide`. The example below demonstrates the registration of free **provider functions** via the `Provide` API.


```go
package main

import (
	"fmt"

	"cosmossdk.io/depinject"
)

type AnotherInt int

func main() {
	var (
	  x int
	  y AnotherInt
	)

	fmt.Printf("Before (%v, %v)\n", x, y)
	depinject.Inject(
		depinject.Provide(
			func() int { return 1 },
			func() AnotherInt { return AnotherInt(2) },
		),
		&x,
		&y,
	)
	fmt.Printf("After (%v, %v)\n", x, y)
}
```

Provider functions form the basis of the dependency tree, they are introspected then their inputs identified as dependencies and outputs as dependants, either for another provider function or state stored outside the DI container, as is the case of `&x` and `&y` above.

### Interface type resolution

`depinject` supports interface types as inputs to provider functions.  In the SDK's case this pattern is used to decouple
`Keeper` dependencies between modules.  For example `x/bank` expects an [AccountKeeper](https://pkg.go.dev/github.com/cosmos/cosmos-sdk/x/bank/types#AccountKeeper) interface as [input to ProvideModule](https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-rc1/x/bank/module.go#L208-L260).

Concretely `SimApp` uses the implementation in `x/auth`, but this design allows for this loose coupling to change.

Given the following types:

```go
package duck

type Duck interface {
	quack()
}

type AlsoDuck interface {
	quack()
}

type Mallard struct{}
type Canvasback struct{}

func (duck Mallard) quack()    {}
func (duck Canvasback) quack() {}

type Pond struct {
	Duck AlsoDuck
}
```

This usage

```go
var pond Pond

depinject.Inject(
  depinject.Provide(
    func() Mallard { return Mallard{} },
    func(duck Duck) Pond {
      return Pond{Duck: duck}
    }),
   &pond)
```

results in an *implicit* binding of `Duck` to `Mallard`.  This works because there is only one implementation of `Duck` in the container.  
However, adding a second provider of `Duck` will result in an error:

```go
var pond Pond

depinject.Inject(
  depinject.Provide(
    func() Mallard { return Mallard{} },
    func() Canvasback { return Canvasback{} },
    func(duck Duck) Pond {
      return Pond{Duck: duck}
    }),
   &pond)
```

A specific binding preference for `Duck` is required.

#### `BindInterface` API

In the above situation registering a binding for a given interface binding may look like

```go
depinject.Inject(
  depinject.Configs(
    depinject.BindInterface(
      "duck.Duck",
      "duck.Mallard"),
     depinject.Provide(
       func() Mallard { return Mallard{} },
       func() Canvasback { return Canvasback{} },
       func(duck Duck) APond {
         return Pond{Duck: duck}
      })),
   &pond)
```

Now `depinject` has enough information to provide `Mallard` as an input to `APond`. 

### Full example in real app

:::warning
When using `depinject.Inject`, the injected types must be pointers.
:::

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-rc1/simapp/app_v2.go#L219-L244
```

## Debugging

Issues with resolving dependencies in the container can be done with logs and [Graphviz](https://graphviz.org) renderings of the container tree.
By default, whenever there is an error, logs will be printed to stderr and a rendering of the dependency graph in Graphviz DOT format will be saved to `debug_container.dot`.

Here is an example Graphviz rendering of a successful build of a dependency graph:
![Graphviz Example](https://raw.githubusercontent.com/cosmos/cosmos-sdk/ff39d243d421442b400befcd959ec3ccd2525154/depinject/testdata/example.svg)

Rectangles represent functions, ovals represent types, rounded rectangles represent modules and the single hexagon
represents the function which called `Build`. Black-colored shapes mark functions and types that were called/resolved
without an error. Gray-colored nodes mark functions and types that could have been called/resolved in the container but
were left unused.

Here is an example Graphviz rendering of a dependency graph build which failed:
![Graphviz Error Example](https://raw.githubusercontent.com/cosmos/cosmos-sdk/ff39d243d421442b400befcd959ec3ccd2525154/depinject/testdata/example_error.svg)

Graphviz DOT files can be converted into SVG's for viewing in a web browser using the `dot` command-line tool, ex:

```txt
dot -Tsvg debug_container.dot > debug_container.svg
```

Many other tools including some IDEs support working with DOT files.
