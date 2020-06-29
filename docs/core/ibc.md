<!--
order: 9
-->

# IBC

This document serves as a guide for developers who want to write their own IBC applications. Due to the modular design of the IBC protocol, IBC application developers do not need to concern themselves with the low-level details of clients, connections, and proof verification. Nevertheless a brief explanation of the lower levels of the stack is given so that application developers may have a high-level understanding of the IBC protocol. Then the document goes into detail on the abstraction layer most relevant for application developers (channels and ports), and describes how to define your own custom packets, and `IBCModule` callbacks.

To have your module interact over IBC you must: bind to a port(s), define your own packet data (and optionally acknowledgement) structs as well as how to encode/decode them, and implement the `IBCModule` interface. Below is a more detailed explanation of how to write an IBC application module correctly.

## Pre-requisites Readings

- [IBC SDK specification](../../modules/ibc) {prereq}

## Core IBC Overview

**[Clients](https://github.com/cosmos/cosmos-sdk/tree/master/x/ibc/02-client)**: IBC Clients are light clients (identified by a unique client-id) that track the consensus states of other blockchains, along with the proof spec necessary to properly verify proofs against the client's consensus state. A client may be associated with any number of connections.

**[Connections](https://github.com/cosmos/cosmos-sdk/tree/master/x/ibc/03-connection)**: Connections encapsulate two `ConnectionEnd` objects on two seperate blockchains. Each `ConnectionEnd` is associated with a client of the other blockchain (ie counterparty blockchain). The connection handshake is responsible for verifying that the light clients on each chain are correct for their respective counterparties. Connections, once established, are responsible for facilitating all cross-chain verification of IBC state. A connection may be associated with any number of channels.

**[Proofs](https://github.com/cosmos/cosmos-sdk/tree/master/x/ibc/23-commitment) and [Paths](https://github.com/cosmos/cosmos-sdk/tree/master/x/ibc/24-host)**: In IBC, blockchains do not directly pass messages to each other over the network. Instead, to communicate, a blockchain will commit some state to a specifically defined path reserved for a specific message type and a specific counterparty (perhaps storing a specific connectionEnd as part of a handshake, or a packet intended to be relayed to a module on the counterparty chain). A relayer process monitors for updates to these paths, and will relay messages, by submitting the data stored under the path along with a proof to the counterparty chain. The paths that all IBC implementations must use for committing IBC messages is defined in [ICS-24](https://github.com/cosmos/ics/tree/master/spec/ics-024-host-requirements) and the proof format that all implementations must be able to produce and verify is defined in this [ICS-23 implementation](https://github.com/confio/ics23).

**[Capabilities](https://github.com/cosmos/cosmos-sdk/tree/master/x/capability)**: IBC is intended to work in execution environements where modules do not necessarily trust each other. Thus IBC must authenticate module actions on ports and channels so that only modules with the appropriate permissions can use them. This is accomplished using dynamic capabilities ([ADR](../architecture/adr-003-dynamic-capability-store.md)). Upon binding to a port or creating a channel for a module, IBC will return a dynamic capability that the module must claim in order to use that port or channel. This prevents other modules from using that port or channel since they will not own the appropriate capability. For information on the object capability model, look [here](./ocap.md)

### Channels and Ports

While the above is useful background information, IBC modules do not need to interact at all with these lower-level abstractions. The relevant abstraction layer for IBC application developers is that of channels and ports. IBC applications should be written as self-contained **modules**. A module on one blockchain can thus communicate with other modules on other blockchains by sending, receiving and acknowledging packets through channels, which are uniquely identified by the `(channelID, portID)` tuple. A useful analogy is to consider IBC modules as internet applications on a computer. A channel can then be conceptualized as an IP connection, with the IBC portID being analogous to a IP port and the IBC channelID being analogous to an IP address. Thus, a single instance of an IBC module may communicate on the same port with any number of other modules and IBC will correctly route all packets to the relevant module using the `(channelID, portID)` tuple. An IBC module may also communicate with another IBC module over multiple ports, with each `(portID<->portID)` packet stream being sent on a different unique channel.

#### [Ports](https://github.com/cosmos/cosmos-sdk/tree/master/x/ibc/05-port)

An IBC module may bind to any number of ports. Each port must be identified by a unique `portID`. Since IBC is designed to be secure with mutually-distrusted modules operating on the same ledger, binding a port will return a dynamic object capability. In order to take action on a particular port (eg open a channel with its portID), a module must provide the dynamic object capability to the IBC handler. This prevents a malicious module from opening channels with ports it does not own. Thus, IBC modules are responsible for claiming the capability that is returned on `BindPort`. Currently, ports must be bound on app initialization. A module may bind to ports in `InitGenesis` like so:

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

#### [Channels](https://github.com/cosmos/cosmos-sdk/tree/master/x/ibc/04-channel)

An IBC channel can be established between 2 IBC ports. Currently, a port is exclusively owned by a single module. IBC packets are sent over channels. Just as IP packets contain the destination IP address and IP port as well as the source IP address and source IP port, IBC packets will contain the destination portID and channelID as well as the source portID and channelID. This enables IBC to correctly route packets to the destination module, while also allowing modules receiving packets to know the sender module.

A channel may be `ORDERED`, in which case, packets from a sending module must be processed by the receiving module in the order they were sent. Or a channel may be `UNORDERED`, in which case packets from a sending module are processed in the order they arrive (may not be the order they were sent).

Modules may choose which channels they wish to communicate over with, thus IBC expects modules to implement callbacks that are called during the channel handshake. These callbacks may do custom channel initialization logic, if any return an error, the channel handshake will fail. Thus, by returning errors on callbacks, modules can programatically reject and accept channels.

The SDK expects all IBC modules to implement the interface `IBCModule`, defined [here](https://github.com/cosmos/cosmos-sdk/tree/master/x/ibc/05-port/types/module.go). This interface contains all of the callbacks IBC expects modules to implement. This section will describe the callbacks that are called during channel handshake execution.

The channel handshake is a 4 step handshake. Briefly, if chainA wants to open a channel with chainB using an already established connection, chainA will do `ChanOpenInit`, chainB will do `chanOpenTry`, chainA will do `ChanOpenAck`, chainB will do `chanOpenConfirm`. If all this happens successfully, the channel will be open on both sides. At each step in the handshake, the module associated with the `channelEnd` will have it's callback executed for that step of the handshake. So on `chanOpenInit`, the module on chainA will have its callback `OnChanOpenInit` executed.

Just as ports came with dynamic capabilites, channel initialization will return a dynamic capability that the module **must** claim so that they can pass in a capability to authenticate channel actions like sending packets. The channel capability is passed into the callback on the first parts of the handshake; either `OnChanOpenInit` on the initializing chain or `OnChanOpenTry` on the other chain.

Here are the channel handshake callbacks that modules are expected to implement:

```go
// Called by IBC Handler on MsgOpenInit
func (k Keeper) OnChanOpenInit(ctx sdk.Context,
    order channeltypes.Order,
    connectionHops []string,
    portID string,
    channelID string,
    channelCap *capabilitytypes.Capability,
    counterParty channeltypes.Counterparty,
    version string,
) error {
    // OpenInit must claim the channelCapability that IBC passes into the callback
    k.scopedKeeper.ClaimCapability(ctx, channelCap)

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
    // OpenInit must claim the channelCapability that IBC passes into the callback
    k.scopedKeeper.ClaimCapability(ctx, channelCap)

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

The channel closing handshake will also invoke module callbacks that can return errors to abort the closing handshake. Closing a channel is a 2-step handshake, the initiating chain calls `ChanCloseInit` and the finalizing chain calls `ChanCloseConfirm`.

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

#### Packets

Modules communicate with each other by sending packets over IBC channels. As mentioned above, all IBC packets contain the destination `portID` and `channelID` along with the source `portID` and `channelID`, this allows modules to know the sender module of a given packet. IBC packets also contain a sequence to optionally enforce ordering. IBC packets also contain a `TimeoutTimestamp` and `TimeoutHeight`, which when non-zero, will determine the deadline before which the receiving module must process a packet. If the timeout passes without the packet being successfully received, the sending module can timeout the packet and take appropriate actions.

Modules send custom application data to each other inside the `Data []byte` field of the IBC packet. Thus, packet data is completely opaque to IBC handlers. It is incumbent on a sender module to encode their application-specific packet information into the `Data` field of packets, and the receiver module to decode that `Data` back to the original application data.

Thus, modules connected by a channel must agree on what application data they are sending over the channel, as well as how they will encode/decode it. This process is not specified by IBC as it is up to each application module to determine how to implement this agreement. However, for most applications this will happen as a version negotiation during the channel handshake. While more complex version negotiation is possible to implement inside the channel opening handshake, a very simple version negotation is implemented in the [ibc-transfer module](https://github.com/cosmos/cosmos-sdk/tree/master/x/ibc-transfer/module.go).

Thus a module must define a custom packet data structure, along with a well-defined way to encode and decode it to and from `[]byte`.

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

A module receiving a packet must decode the `PacketData` into a structure it expects so that it can act on it.

```go
// Receiving custom application packet data (in OnRecvPacket)
packetData := DecodePacketData(packet.Data)
// handle received custom packet data
```

#### Acknowledgements

Modules may optionally commit an acknowledgement upon receiving and processing a packet. This acknowledgement can then be relayed back to the original sender chain, which can take action depending on the contents of the acknowledgement.

Just as packet data was opaque to IBC, acknowledgements are similarly opaque. Modules must pass and receive acknowledegments with the IBC modules as byte strings.

Thus, modules must agree on how to encode/decode acknowledgements. The process of creating an acknowledgement struct along with encoding and decoding it, is very similar to the packet data example above.

#### Packet Flow Handling

Just as IBC expected modules to implement callbacks for channel handshakes, IBC also expects modules to implement callbacks for handling the packet flow through a channel.

Once a module A and module B are connected to each other, relayers can start relaying packets and acknowledgements back and forth on the channel. The packet flow diagram is [here](https://github.com/cosmos/ics/blob/master/spec/ics-004-channel-and-packet-semantics/packet-state-machine.png). Briefly, a successful packet flow works as follows: module A sends a packet through the IBC module, the packet is received by module B, if module B writes an acknowledgement of the packet then module A will process the acknowledgement. If the packet is not successfully received before the timeout, then module A processes the packet's timeout.

**Sending Packets**: Modules do not send packets through callbacks, since the modules initiate the action of sending packets to the IBC module, as opposed to other parts of the packet flow where msgs sent to the IBC module must trigger execution on the port-bound module through the use of callbacks. Thus, to send a packet a module simply needs to call `SendPacket` on the `IBCChannelKeeper`.

```go
// retrieve the dynamic capability for this channel
channelCap := scopedKeeper.GetCapability(ctx, channelCapName)
// Sending custom application packet data
data := EncodePacketData(customPacketData)
packet.Data = data
// Send packet to IBC, authenticating with channelCap
IBCChannelKeeper.SendPacket(ctx, channelCap, packet)
```

Note: In order to prevent modules from sending packets on channels they do not own, IBC expects modules to pass in the correct channel capability for the packet's source channel.

**Receiving Packets**: To handle receiving packets, the module must implement the `OnRecvPacket` callback. This gets invoked by the IBC module after the packet has been proved valid and correctly processed by the IBC keepers. Thus, the `OnRecvPacket` callback only needs to worry about making the appropriate state changes given the packet data without worrying about whether the packet is valid or not.

Modules may optionally return an acknowledgement as a byte string and return it to the IBC handler. The IBC handler will then commit this acknowledgment of the packet so that a relayer may relay the acknowledgement back to the sender module.

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
`OnRecvPacket` should **only** return an error if we want the entire receive packet execution (including the IBC handling) to be reverted. This will allow the packet to be replayed in the case that some mistake in the relaying caused the packet processing to fail.

If some application-level error happened while processing the packet data, in most cases, we will not want the packet processing to revert. Instead, we may want to encode this failure into the acknowledgement and finish processing the packet. This will ensure the packet cannot be replayed, and will also allow the sender module to potentially remediate the situation upon receiving the acknowledgement. An example of this technique is in the ibc-transfer module's [OnRecvPacket](https://github.com/cosmos/cosmos-sdk/tree/master/x/ibc-transfer/module.go).
:::

**Acknowledging Packets**: If the receiving module writes an ackowledgement while processing the packet, a relayer can relay back the acknowledgement to the sender module. The sender module can then process the acknowledgement using the `OnAcknowledgementPacket` callback. The contents of the acknowledgement is entirely upto the modules on the channel (just like the packet data); however, it may often contain information on whether the packet was successfully received and processed along with some additional data that could be useful for remediation if the packet processing failed.

Since the modules are responsible for agreeing on an encoding/decoding standard for packet data and acknowledgements, IBC will pass in the acknowledgements as `[]byte` to this callback. The callback is responsible for decoding the acknowledgment and processing it.

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

**Timeout Packets**: If the timout for a packet is reached before the packet is successfully received, the receiving chain can no longer process it. Thus, the sending chain must process the timout using `OnTimeoutPacket` to handle this situation. Again the IBC module will verify that the timeout is indeed valid, so our module only needs to implement the state machine logic for what to do once a timeout is reached and the packet can no longer be received.

```go
OnTimeoutPacket(
    ctx sdk.Context,
    packet channeltypes.Packet,
) (*sdk.Result, error) {
    // do custom timeout logic
}
```

#### Registering Module with the IBC Router

IBC needs to know which module is bound to which port so that it can route packets to the appropriate module and call the appropriate callback. The port to module name mapping is handled by IBC's portKeeper. However, the mapping from module name to the relevant callbacks is accomplished by the [port.Router](https://github.com/cosmos/cosmos-sdk/tree/master/x/ibc//05-port/types/router.go).

As mentioned above, modules must implement the IBC module interface (which contains both channel handshake callbacks and packet handling callbacks). The concrete implementation of this interface must be registered with the module name as a route on the IBC Router.

Currently, the Router is static so it must be initialized and set correctly on **app initialization. Once the Router has been set, no new routes can be added.

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

Adding the module routes allows the IBC handler to call the appropriate callback when processing a channel handshake or a packet.

#### Working Example

For a real working example of an IBC application, you can look through the `ibc-transfer` module which implements everything discussed above.

Here are the useful parts of the module to look at:

[Binding to transfer port](https://github.com/cosmos/cosmos-sdk/blob/master/x/ibc-transfer/genesis.go)

[Sending transfer packets](https://github.com/cosmos/cosmos-sdk/blob/master/x/ibc-transfer/keeper/relay.go)

[Implementing IBC callbacks](https://github.com/cosmos/cosmos-sdk/blob/master/x/ibc-transfer/module.go)
