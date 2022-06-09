# Cosmos SDK Dependency Injection `depinject` Module

## Overview

`depinject` is a dependency injection framework for the Cosmos SDK. This module together with `core/appconfig` are meant
to simplify the definition of a blockchain by replacing most of app.go's boilerplate code with a configuration file (YAML or JSON).

Configuration is provided by a single
[YAML file](https://github.com/cosmos/cosmos-sdk/blob/main/simapp/app.yaml) as specified by
[proto3](https://github.com/cosmos/cosmos-sdk/blob/main/proto/cosmos/auth/module/v1/module.proto) message types.  This
configuration is often passed to [provider functions](https://github.com/cosmos/cosmos-sdk/blob/33c4bac3d3acbb820b97173882fa9feddacb2f5a/runtime/module.go#L106-L131)
whose outputs form the concrete dependency tree.  Type bindings and tree shape are resolved by reflecting over provider
functions, identifying function parameters as inputs (dependencies) and return values as outputs (tree nodes).

## Usage

`depinject` includes an expressive and composable [Configuration API](https://pkg.go.dev/github.com/cosmos/cosmos-sdk/depinject#Config).
A core configuration is `Provide`, for example this code snippet

```go
package main

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/depinject"
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

demonstrates the registration of free **provider functions** via the `Provide` API.  Provider functions form the basis of the
dependency tree, they are introspected then their inputs identified as dependencies and outputs as dependants, either for
another provider function or state stored outside the DI container, as is the case of `&x` and `&y` above.

### Interface type resolution

`depinject` supports interface types as inputs to provider functions.  In the SDK's case this pattern is used to decouple
`Keeper` dependencies between modules.  For example `x/bank` expects an [AccountKeeper] interface as [input to provideModule](https://github.com/cosmos/cosmos-sdk/blob/de343d458aa68c19630177807d6f0e2e6deaf7a9/x/bank/module.go#L224).
Concretely `SimApp` uses the implementation in `x/auth`, but this design allows for this loose coupling to change.

Given the following types

```golang
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

```golang
var pond Pond

depinject.Inject(
  depinject.Provide(
    func() Mallard { return Mallard{} },
    func(duck Duck) Pond {
      return Pond{Duck: duck}
    }),
   &pond)
```

results in an *implicit* binding of `Duck` to `Mallard`.  This works because there is only one implementation of `Duck`
in the container.  However, adding a second provider of `Duck` will result in an error:

```golang
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

#### `Prefer` API

In the above situation registering a preference for a given interface binding may look like

```golang
depinject.Inject(
  depinject.Configs(
  depinject.Prefer(
    "duck.Duck",
    "duck.Mallard"),
   depinject.Provide(
     func() Mallard { return Mallard{} },
     func() Canvasback { return Canvasback{} }),
     func(duck Duck) APond {
       return Pond{Duck: duck}
    }),
   &pond)
```

Now `depinject` has enough information to provide `Mallard` as an input to `APond`. 

### Full example in real app

```go
//ego:embed app.yaml
var appConfigYaml []byte

var appConfig = appconfig.LoadYAML(appConfigYaml)

func NewSimApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	invCheckPeriod uint,
	encodingConfig simappparams.EncodingConfig,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *SimApp {
	app := &SimApp{
		invCheckPeriod: invCheckPeriod,
	}

	var (
		appBuilder *runtime.AppBuilder
		msgServiceRouter *baseapp.MsgServiceRouter
	)

	err := depinject.Inject(AppConfig,
		&appBuilder,
		&app.ParamsKeeper,
		&app.CapabilityKeeper,
		&app.appCodec,
		&app.legacyAmino,
		&app.interfaceRegistry,
		&app.AccountKeeper,
		&app.BankKeeper,
		&app.FeeGrantKeeper,
		&app.StakingKeeper,
		&msgServiceRouter,
	)
	if err != nil {
		panic(err)
	}
...
```

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