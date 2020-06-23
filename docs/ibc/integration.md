<!--
order: 2
-->

# Integration

Learn how to integrate IBC to your application and send data packets to other chains. {synopsis}

This document outlines the required steps to integrate and configure the [IBC
module](https://github.com/cosmos/cosmos-sdk/tree/master/x/ibc) to your Cosmos SDK application and
send fungible token transfers to other chains.

## Integrating the IBC module

```go
// app.go
//
type App struct {
  // baseapp, keys and subspaces definitions

  // other keepers
  // ...
  IBCKeeper        *ibckeeper.Keeper // IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
  EvidenceKeeper   evidencekeeper.Keeper // required to set up the client misbehaviour route
  TransferKeeper   ibctransferkeeper.Keeper // for cross-chain fungible token transfers

  // make scoped keepers public for test purposes
  ScopedIBCKeeper      capabilitykeeper.ScopedKeeper
  ScopedTransferKeeper capabilitykeeper.ScopedKeeper

  /// ...
  /// module and simulation manager definitions
}
```

### Configure the capabilities' `ScopedKeeper`

### Register the IBC router

IBC needs to know which module is bound to which port so that it can route packets to the
appropriate module and call the appropriate callback. The port to module name mapping is handled by
IBC's port `Keeper`. However, the mapping from module name to the relevant callbacks is accomplished by
the [`port.Router`](https://github.com/cosmos/cosmos-sdk/tree/master/x/ibc//05-port/types/router.go).

As mentioned above, modules must implement the IBC module interface (which contains both channel
handshake callbacks and packet handling callbacks). The concrete implementation of this interface
must be registered with the module name as a route on the IBC `Router`.

Currently, the `Router` is static so it must be initialized and set correctly on app initialization.
Once the `Router` has been set, no new routes can be added.

```go
// app.go

// Create static IBC router, add module routes, then set and seal it
ibcRouter := port.NewRouter()

// Note: moduleCallbacks must implement IBCModule interface
ibcRouter.AddRoute(moduleName, moduleCallbacks)

// Setting Router will finalize all routes by sealing router
// No more routes can be added
app.IBCKeeper.SetRouter(ibcRouter)
```

Adding the module routes allows the IBC handler to call the appropriate callback when processing a
channel handshake or a packet.

### Recap

To sumarize, the steps required to add the IBC module to your application are:

- Define `IBCKeeper` and `ScopedKeeper` field on your application
- Initialize the keepers
- Register the IBC router
- Register evidence types

Learn about how to create [custom IBC modules](./custom.md) for your application {hide}
