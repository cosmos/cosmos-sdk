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

To have your module interact over IBC you must: bind to a port(s), define your own packet data and acknolwedgement structs as well as how to encode/decode them, and implement the
`IBCModule` interface. Below is a more detailed explanation of how to write an IBC application
module correctly.

## Pre-requisites Readings

- [IBC Overview](./overview.md)) {prereq}
- [IBC default integration](./integration.md) {prereq}

## Create a custom IBC application module

### Implement `IBCModule` Interface and callbacks

The Cosmos SDK expects all IBC modules to implement the [`IBCModule`
interface](https://github.com/cosmos/cosmos-sdk/tree/release/v0.42.x/x/ibc/core/05-port/types/module.go). This
interface contains all of the callbacks IBC expects modules to implement. This section will describe
the callbacks that are called during channel handshake execution.

Here are the channel handshake callbacks that modules are expected to implement:

```go
// Called by IBC Handler on MsgOpenInit
func (k Keeper) OnChanOpenInit(ctx sdk.Context,
    order channeltypes.Order,
    connectionHops []string,
    portID string,
    channelID string,
    channelCap *capabilitytypes.Capability,
    counterparty channeltypes.Counterparty,
    version string,
) error {
    // OpenInit must claim the channelCapability that IBC passes into the callback
    if err := k.ClaimCapability(ctx, chanCap, host.ChannelCapabilityPath(portID, channelID)); err != nil {
			return err
	}

    // ... do custom initialization logic

    // Use above arguments to determine if we want to abort handshake
    // Examples: Abort if order == UNORDERED,
    // Abort if version is unsupported
    err := checkArguments(args)
    return err
}

// Called by IBC Handler on MsgOpenTry
OnChanOpenTry(
    ctx sdk.Context,
    order channeltypes.Order,
    connectionHops []string,
    portID,
    channelID string,
    channelCap *capabilitytypes.Capability,
    counterparty channeltypes.Counterparty,
    version,
    counterpartyVersion string,
) error {
    // Module may have already claimed capability in OnChanOpenInit in the case of crossing hellos
    // (ie chainA and chainB both call ChanOpenInit before one of them calls ChanOpenTry)
    // If the module can already authenticate the capability then the module already owns it so we don't need to claim
    // Otherwise, module does not have channel capability and we must claim it from IBC
    if !k.AuthenticateCapability(ctx, chanCap, host.ChannelCapabilityPath(portID, channelID)) {
        // Only claim channel capability passed back by IBC module if we do not already own it
        if err := k.scopedKeeper.ClaimCapability(ctx, chanCap, host.ChannelCapabilityPath(portID, channelID)); err != nil {
            return err
        }
    }
    
    // ... do custom initialization logic

    // Use above arguments to determine if we want to abort handshake
    err := checkArguments(args)
    return err
}

// Called by IBC Handler on MsgOpenAck
OnChanOpenAck(
    ctx sdk.Context,
    portID,
    channelID string,
    counterpartyVersion string,
) error {
    // ... do custom initialization logic

    // Use above arguments to determine if we want to abort handshake
    err := checkArguments(args)
    return err
}

// Called by IBC Handler on MsgOpenConfirm
OnChanOpenConfirm(
    ctx sdk.Context,
    portID,
    channelID string,
) error {
    // ... do custom initialization logic

    // Use above arguments to determine if we want to abort handshake
    err := checkArguments(args)
    return err
}
```

The channel closing handshake will also invoke module callbacks that can return errors to abort the
closing handshake. Closing a channel is a 2-step handshake, the initiating chain calls
`ChanCloseInit` and the finalizing chain calls `ChanCloseConfirm`.

```go
// Called by IBC Handler on MsgCloseInit
OnChanCloseInit(
    ctx sdk.Context,
    portID,
    channelID string,
) error {
    // ... do custom finalization logic

    // Use above arguments to determine if we want to abort handshake
    err := checkArguments(args)
    return err
}

// Called by IBC Handler on MsgCloseConfirm
OnChanCloseConfirm(
    ctx sdk.Context,
    portID,
    channelID string,
) error {
    // ... do custom finalization logic

    // Use above arguments to determine if we want to abort handshake
    err := checkArguments(args)
    return err
}
```

#### Channel Handshake Version Negotiation

Application modules are expected to verify versioning used during the channel handshake procedure.

* `ChanOpenInit` callback should verify that the `MsgChanOpenInit.Version` is valid
* `ChanOpenTry` callback should verify that the `MsgChanOpenTry.Version` is valid and that `MsgChanOpenTry.CounterpartyVersion` is valid.
* `ChanOpenAck` callback should verify that the `MsgChanOpenAck.CounterpartyVersion` is valid and supported.

Versions must be strings but can implement any versioning structure. If your application plans to
have linear releases then semantic versioning is recommended. If your application plans to release
various features in between major releases then it is advised to use the same versioning scheme
as IBC. This versioning scheme specifies a version identifier and compatible feature set with
that identifier. Valid version selection includes selecting a compatible version identifier with
a subset of features supported by your application for that version. The struct is used for this
scheme can be found in `03-connection/types`.

Since the version type is a string, applications have the ability to do simple version verification
via string matching or they can use the already impelemented versioning system and pass the proto
encoded version into each handhshake call as necessary.

ICS20 currently implements basic string matching with a single supported version.

### Bind Ports

Currently, ports must be bound on app initialization. A module may bind to ports in `InitGenesis`
like so:

```go
func InitGenesis(ctx sdk.Context, keeper keeper.Keeper, state types.GenesisState) {
    // ... other initialization logic

    // Only try to bind to port if it is not already bound, since we may already own
    // port capability from capability InitGenesis
    if !isBound(ctx, state.PortID) {
        // module binds to desired ports on InitChain
        // and claims returned capabilities
        cap1 := keeper.IBCPortKeeper.BindPort(ctx, port1)
        cap2 := keeper.IBCPortKeeper.BindPort(ctx, port2)
        cap3 := keeper.IBCPortKeeper.BindPort(ctx, port3)

        // NOTE: The module's scoped capability keeper must be private
        keeper.scopedKeeper.ClaimCapability(cap1)
        keeper.scopedKeeper.ClaimCapability(cap2)
        keeper.scopedKeeper.ClaimCapability(cap3)
    }

    // ... more initialization logic
}
```

### Custom Packets

Modules connected by a channel must agree on what application data they are sending over the
channel, as well as how they will encode/decode it. This process is not specified by IBC as it is up
to each application module to determine how to implement this agreement. However, for most
applications this will happen as a version negotiation during the channel handshake. While more
complex version negotiation is possible to implement inside the channel opening handshake, a very
simple version negotation is implemented in the [ibc-transfer module](https://github.com/cosmos/cosmos-sdk/tree/release/v0.42.x/x/ibc/applications/transfer/module.go).

Thus, a module must define its a custom packet data structure, along with a well-defined way to
encode and decode it to and from `[]byte`.

```go
// Custom packet data defined in application module
type CustomPacketData struct {
    // Custom fields ...
}

EncodePacketData(packetData CustomPacketData) []byte {
    // encode packetData to bytes
}

DecodePacketData(encoded []byte) (CustomPacketData) {
    // decode from bytes to packet data
}
```

Then a module must encode its packet data before sending it through IBC.

```go
// Sending custom application packet data
data := EncodePacketData(customPacketData)
packet.Data = data
IBCChannelKeeper.SendPacket(ctx, packet)
```

A module receiving a packet must decode the `PacketData` into a structure it expects so that it can
act on it.

```go
// Receiving custom application packet data (in OnRecvPacket)
packetData := DecodePacketData(packet.Data)
// handle received custom packet data
```

#### Packet Flow Handling

Just as IBC expected modules to implement callbacks for channel handshakes, IBC also expects modules
to implement callbacks for handling the packet flow through a channel.

Once a module A and module B are connected to each other, relayers can start relaying packets and
acknowledgements back and forth on the channel.

![IBC packet flow diagram](https://media.githubusercontent.com/media/cosmos/ics/master/spec/ics-004-channel-and-packet-semantics/packet-state-machine.png)

Briefly, a successful packet flow works as follows:

1. module A sends a packet through the IBC module
2. the packet is received by module B
3. if module B writes an acknowledgement of the packet then module A will process the
   acknowledgement
4. if the packet is not successfully received before the timeout, then module A processes the
   packet's timeout.

##### Sending Packets

Modules do not send packets through callbacks, since the modules initiate the action of sending
packets to the IBC module, as opposed to other parts of the packet flow where msgs sent to the IBC
module must trigger execution on the port-bound module through the use of callbacks. Thus, to send a
packet a module simply needs to call `SendPacket` on the `IBCChannelKeeper`.

```go
// retrieve the dynamic capability for this channel
channelCap := scopedKeeper.GetCapability(ctx, channelCapName)
// Sending custom application packet data
data := EncodePacketData(customPacketData)
packet.Data = data
// Send packet to IBC, authenticating with channelCap
IBCChannelKeeper.SendPacket(ctx, channelCap, packet)
```

::: warning
In order to prevent modules from sending packets on channels they do not own, IBC expects
modules to pass in the correct channel capability for the packet's source channel.
:::

##### Receiving Packets

To handle receiving packets, the module must implement the `OnRecvPacket` callback. This gets
invoked by the IBC module after the packet has been proved valid and correctly processed by the IBC
keepers. Thus, the `OnRecvPacket` callback only needs to worry about making the appropriate state
changes given the packet data without worrying about whether the packet is valid or not.

Modules may return an acknowledgement as a byte string and return it to the IBC handler.
The IBC handler will then commit this acknowledgement of the packet so that a relayer may relay the
acknowledgement back to the sender module.

```go
OnRecvPacket(
    ctx sdk.Context,
    packet channeltypes.Packet,
) (res *sdk.Result, ack []byte, abort error) {
    // Decode the packet data
    packetData := DecodePacketData(packet.Data)

    // do application state changes based on packet data
    // and return result, acknowledgement and abortErr
    // Note: abortErr is only not nil if we need to abort the entire receive packet, and allow a replay of the receive.
    // If the application state change failed but we do not want to replay the packet,
    // simply encode this failure with relevant information in ack and return nil error
    res, ack, abortErr := processPacket(ctx, packet, packetData)

    // if we need to abort the entire receive packet, return error
    if abortErr != nil {
        return nil, nil, abortErr
    }

    // Encode the ack since IBC expects acknowledgement bytes
    ackBytes := EncodeAcknowledgement(ack)

    return res, ackBytes, nil
}
```

::: warning
`OnRecvPacket` should **only** return an error if we want the entire receive packet execution
(including the IBC handling) to be reverted. This will allow the packet to be replayed in the case
that some mistake in the relaying caused the packet processing to fail.

If some application-level error happened while processing the packet data, in most cases, we will
not want the packet processing to revert. Instead, we may want to encode this failure into the
acknowledgement and finish processing the packet. This will ensure the packet cannot be replayed,
and will also allow the sender module to potentially remediate the situation upon receiving the
acknowledgement. An example of this technique is in the `ibc-transfer` module's
[`OnRecvPacket`](https://github.com/cosmos/cosmos-sdk/tree/release/v0.42.x/x/ibc/applications/transfer/module.go).
:::

### Acknowledgements

Modules may commit an acknowledgement upon receiving and processing a packet in the case of synchronous packet processing.
In the case where a packet is processed at some later point after the packet has been received (asynchronous execution), the acknowledgement 
will be written once the packet has been processed by the application which may be well after the packet receipt.

NOTE: Most blockchain modules will want to use the synchronous execution model in which the module processes and writes the acknowledgement 
for a packet as soon as it has been received from the IBC module.

This acknowledgement can then be relayed back to the original sender chain, which can take action
depending on the contents of the acknowledgement.

Just as packet data was opaque to IBC, acknowledgements are similarly opaque. Modules must pass and
receive acknowledegments with the IBC modules as byte strings.

Thus, modules must agree on how to encode/decode acknowledgements. The process of creating an
acknowledgement struct along with encoding and decoding it, is very similar to the packet data
example above. [ICS 04](https://github.com/cosmos/ics/tree/master/spec/ics-004-channel-and-packet-semantics#acknowledgement-envelope)
specifies a recommended format for acknowledgements. This acknowledgement type can be imported from
[channel types](https://github.com/cosmos/cosmos-sdk/tree/release/v0.42.x/x/ibc/core/04-channel/types).

While modules may choose arbitrary acknowledgement structs, a default acknowledgement types is provided by IBC [here](https://github.com/cosmos/cosmos-sdk/tree/release/v0.42.x/proto/ibc/core/channel/v1/channel.proto):

```proto
// Acknowledgement is the recommended acknowledgement format to be used by
// app-specific protocols.
// NOTE: The field numbers 21 and 22 were explicitly chosen to avoid accidental
// conflicts with other protobuf message formats used for acknowledgements.
// The first byte of any message with this format will be the non-ASCII values
// `0xaa` (result) or `0xb2` (error). Implemented as defined by ICS:
// https://github.com/cosmos/ics/tree/master/spec/ics-004-channel-and-packet-semantics#acknowledgement-envelope
message Acknowledgement {
  // response contains either a result or an error and must be non-empty
  oneof response {
    bytes  result = 21;
    string error  = 22;
  }
}
```

#### Acknowledging Packets

After a module writes an acknowledgement, a relayer can relay back the acknowledgement to the sender module. The sender module can
then process the acknowledgement using the `OnAcknowledgementPacket` callback. The contents of the
acknowledgement is entirely upto the modules on the channel (just like the packet data); however, it
may often contain information on whether the packet was successfully processed along
with some additional data that could be useful for remediation if the packet processing failed.

Since the modules are responsible for agreeing on an encoding/decoding standard for packet data and
acknowledgements, IBC will pass in the acknowledgements as `[]byte` to this callback. The callback
is responsible for decoding the acknowledgement and processing it.

```go
OnAcknowledgementPacket(
    ctx sdk.Context,
    packet channeltypes.Packet,
    acknowledgement []byte,
) (*sdk.Result, error) {
    // Decode acknowledgement
    ack := DecodeAcknowledgement(acknowledgement)

    // process ack
    res, err := processAck(ack)
    return res, err
}
```

#### Timeout Packets

If the timeout for a packet is reached before the packet is successfully received or the 
counterparty channel end is closed before the packet is successfully received, then the receiving
chain can no longer process it. Thus, the sending chain must process the timeout using
`OnTimeoutPacket` to handle this situation. Again the IBC module will verify that the timeout is
indeed valid, so our module only needs to implement the state machine logic for what to do once a
timeout is reached and the packet can no longer be received.

```go
OnTimeoutPacket(
    ctx sdk.Context,
    packet channeltypes.Packet,
) (*sdk.Result, error) {
    // do custom timeout logic
}
```

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
port](https://github.com/cosmos/cosmos-sdk/tree/release/v0.42.x/x/ibc/applications/transfer/genesis.go)

[Sending transfer
packets](https://github.com/cosmos/cosmos-sdk/tree/release/v0.42.x/x/ibc/applications/transfer/keeper/relay.go)

[Implementing IBC
callbacks](https://github.com/cosmos/cosmos-sdk/tree/release/v0.42.x/x/ibc/applications/transfer/module.go)

## Next {hide}

Learn about [building modules](../building-modules/intro.md) {hide}
