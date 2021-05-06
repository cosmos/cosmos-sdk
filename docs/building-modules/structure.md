<!--
order: 12
-->

# Recommended Folder Structure

This document outlines the recommended structure of Cosmos SDK modules. These ideas are meant to be applied as suggestions. Application developers are encouraged to improve upon and contribute to module structure and development design. {synopsis}

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
├── keeper
│   ├── genesis.go
│   ├── grpc_query.go
│   ├── hooks.go
│   ├── invariants.go
│   ├── keeper.go
│   ├── keys.go
│   ├── msg_server.go
│   └── querier.go
├── module
│   └── module.go
├── simulation
│   ├── decoder.go
│   ├── genesis.go
│   ├── operations.go
│   └── params.go
├── spec
│   ├── 01_concepts.md
│   ├── 02_state.md
│   ├── 03_messages.md
│   └── 04_events.md
├── {module_name}.pb.go
├── abci.go
├── codec.go
├── errors.go
├── events.go
├── events.pb.go
├── expected_keepers.go
├── genesis.go
├── genesis.pb.go
├── keys.go
├── msgs.go
├── params.go
├── querier.go
├── query.pb.go
└── tx.pb.go
```

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
- `keeper/`: The module's keeper implementation along with any auxiliary
implementations such as the querier and invariants.
- `module/`: The module's implementation of the `AppModule` and `AppModuleBasic`
interfaces.
- `simulation/`: The module's simulation package defines all the required functions
used on the blockchain simulator: randomized genesis state, parameters, weighted
operations, proposal contents and types decoders.
- `spec/`: The module's specification documents.
- `types/`: The module's type definitions such as messages, `KVStore` keys,
parameter types, Protocol Buffer definitions, and `expected_keepers.go` contracts.
- `abci.go`: The module's `BeginBlocker` and `EndBlocker` implementations (if any).

## Next {hide}

Learn about [Errors](./errors.md) {hide}
