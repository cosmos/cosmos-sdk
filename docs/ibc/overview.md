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

This section describes the IBC components and links to the repos.

### [Clients](https://github.com/cosmos/ibc-go/blob/main/modules/core/02-client)

IBC clients are light clients that are identified by a unique client id. IBC clients track the consensus states of other blockchains and the proof specs of those blockchains that are required to properly verify proofs against the client's consensus state. A client can be associated with any number of connections to multiple chains. The supported IBC clients are:

- [Solo Machine light client](https://github.com/cosmos/ibc-go/blob/main/modules/light-clients/06-solomachine): devices such as phones, browsers, or laptops.
- [Tendermint light client](https://github.com/cosmos/ibc-go/blob/main/modules/light-clients/07-tendermint): The default for Cosmos SDK-based chains.
- [Localhost (loopback) client](https://github.com/cosmos/ibc-go/blob/main/modules/light-clients/09-localhost): Useful for testing, simulation, and relaying packets to modules on the same application.

### [Connections](https://github.com/cosmos/ibc-go/blob/main/modules/core/03-connection)

Connections encapsulate two `ConnectionEnd` objects on two separate blockchains. Each `ConnectionEnd` is associated with a client of the other blockchain (the counterparty blockchain). The connection handshake is responsible for verifying that the light clients on each chain are correct for their respective counterparties. Connections, once established, are responsible for facilitating all cross-chain verification of IBC state. A connection can be associated with any number of channels.

### [Proofs](https://github.com/cosmos/ibc-go/blob/main/modules/core/23-commitment) and [Paths](https://github.com/cosmos/ibc-go/blob/main/modules/core/24-host)

In IBC, blockchains do not directly pass messages to each other over the network.

- To communicate, a blockchain commits some state to a precisely defined path reserved for a specific message type and a specific counterparty. For example, a blockchain that stores a specific connectionEnd as part of a handshake or a packet intended to be relayed to a module on the counterparty chain.

- A relayer process monitors for updates to these paths and relays messages by submitting the data stored under the path along with a proof of that data to the counterparty chain.

- The paths that all IBC implementations must support for committing IBC messages are defined in [ICS-24 host requirements](https://github.com/cosmos/ics/tree/master/spec/core/ics-024-host-requirements).

- The proof format that all implementations must produce and verify is defined in [ICS-23 implementation](https://github.com/confio/ics23).

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

### [Ports](https://github.com/cosmos/cosmos-sdk/tree/release/v0.42.x/x/ibc/core/05-port)

An IBC module may bind to any number of ports. Each port must be identified by a unique `portID`.
Since IBC is designed to be secure with mutually-distrusted modules operating on the same ledger,
binding a port will return a dynamic object capability. In order to take action on a particular port
(eg open a channel with its portID), a module must provide the dynamic object capability to the IBC
handler. This prevents a malicious module from opening channels with ports it does not own. Thus,
IBC modules are responsible for claiming the capability that is returned on `BindPort`.

### [Channels](https://github.com/cosmos/cosmos-sdk/tree/release/v0.42.x/x/ibc/core/04-channel)

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

### [Packets](https://github.com/cosmos/cosmos-sdk/tree/release/v0.42.x/x/ibc/core/04-channel)

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

### [Receipts and Timeouts](https://github.com/cosmos/cosmos-sdk/tree/release/v0.42.x/x/ibc/core/04-channel)

Since IBC works over a distributed network and relies on potentially faulty relayers to relay messages between ledgers, 
IBC must handle the case where a packet does not get sent to its destination in a timely manner or at all. Thus, packets must 
specify a timeout height or timeout timestamp after which a packet can no longer be successfully received on the destination chain.

If the timeout does get reached, then a proof of packet timeout can be submitted to the original chain which can then perform 
application-specific logic to timeout the packet, perhaps by rolling back the packet send changes (refunding senders any locked funds, etc).

In ORDERED channels, a timeout of a single packet in the channel will cause the channel to close. If packet sequence `n` times out, 
then no packet at sequence `k > n` can be successfully received without violating the contract of ORDERED channels that packets are processed in the order that they are sent. Since ORDERED channels enforce this invariant, a proof that sequence `n` hasn't been received on the destination chain by packet `n`'s specified timeout is sufficient to timeout packet `n` and close the channel.

In the UNORDERED case, packets may be received in any order. Thus, IBC will write a packet receipt for each sequence it has received in the UNORDERED channel. This receipt contains no information, it is simply a marker intended to signify that the UNORDERED channel has received a packet at the specified sequence. To timeout a packet on an UNORDERED channel, one must provide a proof that a packet receipt does not exist for the packet's sequence by the specified timeout. Of course, timing out a packet on an UNORDERED channel will simply trigger the application specific timeout logic for that packet, and will not close the channel.

For this reason, most modules should use UNORDERED channels as they require less liveness guarantees to function effectively for users of that channel.

### [Acknowledgements](https://github.com/cosmos/cosmos-sdk/tree/release/v0.42.x/x/ibc/core/04-channel)

Modules may also choose to write application-specific acknowledgements upon processing a packet. This may either be done synchronously on `OnRecvPacket`, if the module processes packets as soon as they are received from IBC module. Or they may be done asynchronously if module processes packets at some later point after receiving the packet.

Regardless, this acknowledgement data is opaque to IBC much like the packet `Data` and will be treated by IBC as a simple byte string `[]byte`. It is incumbent on receiver modules to encode their acknowledgemnet in such a way that the sender module can decode it correctly. This should be decided through version negotiation during the channel handshake.

The acknowledgement may encode whether the packet processing succeeded or failed, along with additional information that will allow the sender module to take appropriate action.

Once the acknowledgement has been written by the receiving chain, a relayer will relay the acknowledgement back to the original sender module which will then execute application-specific acknowledgment logic using the contents of the acknowledgement. This may involve rolling back packet-send changes in the case of a failed acknowledgement (refunding senders).

Once an acknowledgement is received successfully on the original sender the chain, the IBC module deletes the corresponding packet commitment as it is no longer needed.

## Further Readings and Specs

If you want to learn more about IBC, check the following specifications:

* [IBC specification overview](https://github.com/cosmos/ics/blob/master/ibc/README.md)
* [IBC SDK specification](../../x/ibc/spec/README.md)

## Next {hide}

Learn about how to [integrate](./integration.md) IBC to your application {hide}
