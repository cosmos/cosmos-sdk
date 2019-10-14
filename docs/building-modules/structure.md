# Module Specification

This document outlines the recommended structure of Cosmos SDK modules. These
ideas are meant to be applied as suggestions. Application developers are encouraged
to improve upon and contribute to module structure and development design.

## Structure

A typical Cosmos SDK module can be structured as follows:

```shell
x/{module}
├── client
│   ├── cli
│   │   ├── query.go
│   │   └── tx.go
│   └── rest
│       ├── query.go
│       └── tx.go
├── exported
│   └── exported.go
├── internal
│   ├── keeper
│   │   ├── invariants.go
│   │   ├── keeper.go
│   │   ├── ...
│   │   └── querier.go
│   └── types
│       ├── codec.go
│       ├── errors.go
│       ├── events.go
│       ├── expected_keepers.go
│       ├── genesis.go
│       ├── keys.go
│       ├── msgs.go
│       ├── params.go
│       ├── ...
│       └── querier.go
├── abci.go
├── alias.go
├── genesis.go
├── handler.go
├── module.go
├── ...
└── simulation.go
```

- `abci.go`: The module's `BeginBlocker` and `EndBlocker` implementations (if any).
- `alias.go`: The module's exported types, constants, and variables. These are mainly
to improve developer ergonomics by only needing to import a single package. Note,
there is nothing preventing developers from importing other packages from the module
(excluding`internal/`) but it is recommended that `alias.go` have everything
exposed that other modules may need. The majority of the exported values here will
typically come from `internal/` (see below).
- `client/`: The module's CLI and REST client functionality implementation and 
testing.
- `exported/`: The module's exported types -- typically type interfaces. If a module
relies on other module keepers, it is expected to receive them as interface
contracts through the `expected_keepers.go` (which are detailed below) design to
avoid having a direct dependency on the implementing module. However, these
contracts can define methods that operate on and/or return types that are specific
to the contract's implementing module and this is where `exported/` comes into play.
Types defined here allow for `expected_keepers.go` in other modules to define
contracts that use single canonical types. This pattern allows for code to remain
DRY and also alleviates import cycle chaos.
- `genesis.go`: The module's genesis related business logic (e.g. `InitGenesis`).
Note, genesis types are defined in `internal/types`.
- `handler.go`: The module's message handlers.
- `internal/`: The module's internal types and implementations. The purpose of
this package is mainly two fold. First, it signals that this package is not
intended to be used or imported anywhere outside the defining module. Secondly,
it goes hand-in-hand with `alias.go` in that it allows public types and functions
to be used internally while not being exposed outside to the outside world. This
allows for greater freedom of development while maintaining API stability.
  - `internal/types`: The module's type definitions such as messages, `KVStore`
  keys, parameter types, and `expected_keepers.go` contracts.
  - `internal/keeper`: The module's keeper implementation along with any auxiliary
  implementations such as the querier and invariants.
- `module.go`: The module's implementation of the `AppModule` and `AppModuleBasic`
interfaces.
- `simulation.go`: The module's simulation messages and related types (if any).
