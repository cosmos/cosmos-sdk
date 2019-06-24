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
│   |   ├── query.go
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
exposed that other modules may need.
