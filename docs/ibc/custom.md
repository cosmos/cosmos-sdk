<!--
order: 3
-->

# Customization

Learn how to configure your application to use IBC and send data packets to other chains. {synopsis}

This document serves as a guide for developers who want to write their own Inter-blockchain
Communication Protocol (IBC) applications for custom [use-cases](https://github.com/cosmos/ics/blob/master/ibc/4_IBC_USECASES.md).

Due to the modular design of the IBC protocol, IBC
application developers do not need to concern themselves with the low-level details of clients,
connections, and proof verification. Nevertheless a brief explanation of the lower levels of the
stack is given so that application developers may have a high-level understanding of the IBC
protocol. Then the document goes into detail on the abstraction layer most relevant for application
developers (channels and ports), and describes how to define your own custom packets, and
`IBCModule` callbacks.

To have your module interact over IBC you must: bind to a port(s), define your own packet data (and
optionally acknowledgement) structs as well as how to encode/decode them, and implement the
`IBCModule` interface. Below is a more detailed explanation of how to write an IBC application
module correctly.

## Pre-requisites Readings

- [IBC Overview](./overview.md)) {prereq}
- [IBC default integration](./integration.md) {prereq}

## Create a custom IBC application module

<!-- TODO: split content from overview.md -->

### Routing

As mentioned above, modules must implement the IBC module interface (which contains both channel
handshake callbacks and packet handling callbacks). The concrete implementation of this interface
must be registered with the module name as a route on the IBC `Router`.

```go
// app.go
func NewApp(...args) *App {
// ...

// Create static IBC router, add module routes, then set and seal it
ibcRouter := port.NewRouter()

ibcRouter.AddRoute(ibctransfertypes.ModuleName, transferModule)
// Note: moduleCallbacks must implement IBCModule interface
ibcRouter.AddRoute(moduleName, moduleCallbacks)

// Setting Router will finalize all routes by sealing router
// No more routes can be added
app.IBCKeeper.SetRouter(ibcRouter)
```

## Working Example

For a real working example of an IBC application, you can look through the `ibc-transfer` module
which implements everything discussed above.

Here are the useful parts of the module to look at:

[Binding to transfer
port](https://github.com/cosmos/cosmos-sdk/blob/master/x/ibc-transfer/genesis.go)

[Sending transfer
packets](https://github.com/cosmos/cosmos-sdk/blob/master/x/ibc-transfer/keeper/relay.go)

[Implementing IBC
callbacks](https://github.com/cosmos/cosmos-sdk/blob/master/x/ibc-transfer/module.go)

## Next {hide}

Learn about [object-capabilities](./ocap.md) {hide}
