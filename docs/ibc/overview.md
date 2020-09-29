<!--
order: 1
-->

# Overview

Learn what IBC is, its components and use cases. {synopsis}

## What is the Interblockchain Communication Protocol (IBC)?

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

## Components Overview

### [Clients](https://github.com/cosmos/cosmos-sdk/tree/master/x/ibc/02-client)

IBC Clients are light clients (identified by a unique client-id) that track the consensus states of
other blockchains, along with the proof spec necessary to properly verify proofs against the
client's consensus state. A client may be associated with any number of connections to multiple
chains. The supported IBC clients are:

* [Solo Machine light client](https://github.com/cosmos/cosmos-sdk/tree/master/x/ibc/light-clients/solomachine): devices such as phones, browsers, or laptops.
* [Tendermint light client](https://github.com/cosmos/cosmos-sdk/tree/master/x/ibc/07-tendermint): The default for SDK-based chains,
* [Localhost (loopback) client](https://github.com/cosmos/cosmos-sdk/tree/master/x/ibc/09-localhost): Useful for
testing, simulation and relaying packets to modules on the same application.

### [Connections](https://github.com/cosmos/cosmos-sdk/tree/master/x/ibc/03-connection)

Connections encapsulate two `ConnectionEnd` objects on two seperate blockchains. Each
`ConnectionEnd` is associated with a client of the other blockchain (ie counterparty blockchain).
The connection handshake is responsible for verifying that the light clients on each chain are
correct for their respective counterparties. Connections, once established, are responsible for
facilitation all cross-chain verification of IBC state. A connection may be associated with any
number of channels.

### [Proofs](https://github.com/cosmos/cosmos-sdk/tree/master/x/ibc/23-commitment) and [Paths](https://github.com/cosmos/cosmos-sdk/tree/master/x/ibc/24-host)
  
In IBC, blockchains do not directly pass messages to each other over the network. Instead, to
communicate, a blockchain will commit some state to a specifically defined path reserved for a
specific message type and a specific counterparty (perhaps storing a specific connectionEnd as part
of a handshake, or a packet intended to be relayed to a module on the counterparty chain). A relayer
process monitors for updates to these paths, and will relay messages, by submitting the data stored
under the path along with a proof to the counterparty chain. The paths that all IBC implementations
must use for committing IBC messages is defined in
[ICS-24](https://github.com/cosmos/ics/tree/master/spec/ics-024-host-requirements) and the proof
format that all implementations must be able to produce and verify is defined in this [ICS-23 implementation](https://github.com/confio/ics23).

### [Capabilities](./ocap.md)

IBC is intended to work in execution environements where modules do not necessarily trust each
other. Thus IBC must authenticate module actions on ports and channels so that only modules with the
appropriate permissions can use them. This is accomplished using [dynamic
capabilities](../architecture/adr-003-dynamic-capability-store.md). Upon binding to a port or
creating a channel for a module, IBC will return a dynamic capability that the module must claim in
order to use that port or channel. This prevents other modules from using that port or channel since
they will not own the appropriate capability.

While the above is useful background information, IBC modules do not need to interact at all with
these lower-level abstractions. The relevant abstraction layer for IBC application developers is
that of channels and ports. IBC applications should be written as self-contained **modules**. A
module on one blockchain can thus communicate with other modules on other blockchains by sending,
receiving and acknowledging packets through channels, which are uniquely identified by the
`(channelID, portID)` tuple. A useful analogy is to consider IBC modules as internet applications on
a computer. A channel can then be conceptualized as an IP connection, with the IBC portID being
analogous to a IP port and the IBC channelID being analogous to an IP address. Thus, a single
instance of an IBC module may communicate on the same port with any number of other modules and and
IBC will correctly route all packets to the relevant module using the (channelID, portID tuple). An
IBC module may also communicate with another IBC module over multiple ports, with each
`(portID<->portID)` packet stream being sent on a different unique channel.

### [Ports](https://github.com/cosmos/cosmos-sdk/tree/master/x/ibc/05-port)

An IBC module may bind to any number of ports. Each port must be identified by a unique `portID`.
Since IBC is designed to be secure with mutually-distrusted modules operating on the same ledger,
binding a port will return a dynamic object capability. In order to take action on a particular port
(eg open a channel with its portID), a module must provide the dynamic object capability to the IBC
handler. This prevents a malicious module from opening channels with ports it does not own. Thus,
IBC modules are responsible for claiming the capability that is returned on `BindPort`.

### [Channels](https://github.com/cosmos/cosmos-sdk/tree/master/x/ibc/04-channel)

An IBC channel can be established between 2 IBC ports. Currently, a port is exclusively owned by a
single module. IBC packets are sent over channels. Just as IP packets contain the destination IP
address and IP port as well as the source IP address and source IP port, IBC packets will contain
the destination portID and channelID as well as the source portID and channelID. This enables IBC to
correctly route packets to the destination module, while also allowing modules receiving packets to
know the sender module.

A channel may be `ORDERED`, in which case, packets from a sending module must be processed by the
receiving module in the order they were sent. Or a channel may be `UNORDERED`, in which case packets
from a sending module are processed in the order they arrive (may not be the order they were sent).

Modules may choose which channels they wish to communicate over with, thus IBC expects modules to
implement callbacks that are called during the channel handshake. These callbacks may do custom
channel initialization logic, if any return an error, the channel handshake will fail. Thus, by
returning errors on callbacks, modules can programatically reject and accept channels.

The channel handshake is a 4 step handshake. Briefly, if a given chain A wants to open a channel with
chain B using an already established connection:

1. chain A sends a `ChanOpenInit` message to signal a channel initialization attempt with chain B.
2. chain B sends a `ChanOpenTry` message to try opening the channel on chain A.
3. chain A sends a `ChanOpenAck` message to mark its channel end status as open.
4. chain B sends a `ChanOpenConfirm` message to mark its channel end status as open.

If all this happens successfully, the channel will be open on both sides. At each step in the handshake, the module
associated with the `ChannelEnd` will have it's callback executed for that step of the handshake. So
on `ChanOpenInit`, the module on chain A will have its callback `OnChanOpenInit` executed.

Just as ports came with dynamic capabilites, channel initialization will return a dynamic capability
that the module **must** claim so that they can pass in a capability to authenticate channel actions
like sending packets. The channel capability is passed into the callback on the first parts of the
handshake; either `OnChanOpenInit` on the initializing chain or `OnChanOpenTry` on the other chain.

### [Packets](https://github.com/cosmos/cosmos-sdk/tree/master/x/ibc/04-channel)

Modules communicate with each other by sending packets over IBC channels. As mentioned above, all
IBC packets contain the destination `portID` and `channelID` along with the source `portID` and
`channelID`, this allows modules to know the sender module of a given packet. IBC packets also
contain a sequence to optionally enforce ordering. IBC packets also contain a `TimeoutTimestamp` and
`TimeoutHeight`, which when non-zero, will determine the deadline before which the receiving module
must process a packet. If the timeout passes without the packet being successfully received, the
sending module can timeout the packet and take appropriate actions.

Modules send custom application data to each other inside the `Data []byte` field of the IBC packet.
Thus, packet data is completely opaque to IBC handlers. It is incumbent on a sender module to encode
their application-specific packet information into the `Data` field of packets, and the receiver
module to decode that `Data` back to the original application data.

## Further Readings and Specs

If you want to learn more about IBC, check the following specifications:

* [IBC specification overview](https://github.com/cosmos/ics/blob/master/ibc/README.md)
* [IBC SDK specification](../../modules/ibc)

## Next {hide}

Learn about how to [integrate](./integration.md) IBC to your application {hide}
