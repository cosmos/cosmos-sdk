# Module Specification

This document attempts to outline the recommended structure of Cosmos SDK modules.
However, the ideas outlined here are meant to be applied as suggestions. Application
developers are encouraged to improve upon and contribute to module structure and
development design.

## Structure

A typical Cosmos SDK module can be structured as follows:

```shell
x/{module}
├── abci.go
├── alias.go
├── client
│   ├── cli
│   │   ├── query.go
│   │   └── tx.go
│   └── rest
│       ├── query.go
│       └── tx.go
├── exported
│   └── exported.go
├── genesis.go
├── handler.go
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
│       ├── keys.go
│       ├── msgs.go
│       ├── params.go
│       ├── ...
│       └── querier.go
├── module.go
├── ...
└── simulation.go
```

- `abci.go`: The module's `BeginBlocker` and `EndBlocker` implementations (if any).
- `alias.go`: The module's exported types, constants, and variables. These are mainly
to improve developer ergonomics by only needing to import a single package. Note,
there is nothing preventing developers from importing other packages from the module
(not including `internal/`) but it is recommended that `alias.go` have everything
exposed that other modules may need. A majority of the exported values here will
typically come from `internal/` (see below).
- `client/`: The module's CLI and REST client functionality implementation and 
testing.
- `exported/`: The module's exported types -- typically type interfaces. If a module
relies on other module keepers, it is expected to receive them as interface
contracts through `expected_keepers.go` (which are detailed below) to avoid having
a direct dependency on that module. However, these contracts can define methods
that operate on and/or return types that are specific to the contract's implementing
module and this is where `exported/` comes into play. Types defined here allow for
`expected_keepers.go` in other modules to define contracts that use single
canonical types and allows for code to remain DRY.
- `genesis.go`: The module's genesis related business logic (e.g. `InitGenesis`).
- `handler.go`: The module's message handlers.