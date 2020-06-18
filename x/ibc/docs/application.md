## Application Writer's Guide

This document serves as a guide for developers who want to write their own IBC applications. Due to the modular design of the IBC protocol, IBC application developers do not need to concern themselves with the low-level details of clients, connections, and proof verification. Nevertheless a brief explanation of the lower levels of the stack is given so that application developers may have a high-level understanding of the IBC protocol. Then the document goes into detail on the abstraction layer most relevant for application developers (channels and ports), and describes how to define your own custom packets, and IBCModule callbacks.

### Core IBC Overview

**[Clients](../02-client)**: IBC Clients are light clients (identified by a unique client-id) that track the consensus states of other blockchains, along with the proof spec necessary to properly verify proofs against the client's consensus state. A client may be associated with any number of connections.

**[Connections](../03-connection)**: Connections encapsulate two ConnectionEnd objects on two seperate blockchains. Each ConnectionEnd is associated with a client of the other blockchain (ie counterparty blockchain). The connection handshake is responsible for verifying that the light clients on each chain are correct for their respective counterparties. Connections, once established, are responsible for facilitation all cross-chain verification of IBC state. A connection may be associated with any number of channels.

**[Proofs](../23-commitment) and [Paths](../24-host)**: In IBC, blockchains do not directly pass messages to each other over the network. Instead, to communicate, a blockchain will commit some state to a specifically defined path reserved for a specific message type and a specific counterparty (perhaps storing a specific connectionEnd as part of a handshake, or a packet intended to be relayed to a module on the counterparty chain). A relayer process monitors for updates to these paths, and will relay messages, by submitting the data stored under the path along with a proof to the counterparty chain. The paths that all IBC implementations must use for committing IBC messages is defined in [ICS-24](https://github.com/cosmos/ics/tree/master/spec/ics-024-host-requirements) and the proof format that all implementations must be able to produce and verify is defined in this [ICS-23 implementation](https://github.com/confio/ics23).

**Capabilities**: TODO

### Channels and Ports

While the above is useful background information, IBC modules do not need to interact at all with these lower-level abstractions. The relevant abstraction layer for IBC application developers is that of channels and ports. IBC applications should be written as self-contained __modules__. A module on one blockchain can thus communicate with other modules on other blockchains by sending, receiving and acknowledging packets through channels, which are uniquely identified by the `(channelID, portID)` tuple. A useful analogy is to consider IBC modules as internet applications on a computer. A channel can then be conceptualized as an IP connection, with the IBC portID being analogous to a IP port and the IBC channelID being analogous to an IP address. Thus, a single instance of an IBC module may communicate on the same port with any number of other modules and and IBC will correctly route all packets to the relevant module using the (channelID, portID tuple). An IBC module may also communicate with another IBC module over multiple ports, with each `(portID<->portID)` packet stream being sent on a different unique channel.

**[Ports](../05-port)**: An IBC module may bind to any number of ports. Each port must be identified by a unique `portID`. Since IBC is designed to be secure with mutually-distrusted modules operating on the same ledger, binding a port will return a dynamic object capability. In order to take action on a particular port (eg open a channel with its portID), a module must provide the dynamic object capability to the IBC handler. This prevents a malicious module from opening channels with ports it does not own. Thus, IBC modules are responsible for claiming the capability that is returned on `BindPort`. Currently, ports must be bound on app initialization. A module may bind to ports in `InitGenesis` like so:

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

**[Channels](../04-channel)**: An IBC channel can be established between 2 IBC ports. Currently, a port is exclusively owned by a single module. IBC packets are sent over channels. Just as IP packets contain the destination IP address and IP port as well as the the source IP address and source IP port, IBC packets will contain the destination portID and channelID as well as the source portID and channelID. This enables IBC to correctly route packets to the destination module, while also allowing modules receiving packets to know the sender module.

Modules may choose which channels they wish to communicate over with, thus IBC expects modules to implement callbacks that are called during the channel handshake. These callbacks may do custom channel initialization logic, if any return an error, the channel handshake will fail. Thus, by returning errors on callbacks, modules can programatically reject and accept channels.

The SDK expects all IBC modules to implement the interface `IBCModule`, defined [here](../05-port/types/module.go). This interface contains all of the callbacks IBC expects modules to implement. This section will describe the callbacks that are called during channel handshake execution.

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

**Packets**:

// TODO