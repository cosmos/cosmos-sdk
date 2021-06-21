<!-- order: 1 -->

# IBC Overview

Learn what IBC is, its components, and use cases. {synopsis}

## What is the Inter-Blockchain Communication Protocol (IBC)?

This document is a guide for developers who want to write their own IBC apps for custom use cases.

The modular design of the IBC protocol means that IBC app developers do not require in-depth knowledge of the low-level details of clients, connections, and proof verification. This brief explanation of the lower levels of the stack is provided so that app developers can gain a high-level understanding of the IBC protocol.

The abstraction layer details on channels and ports are relevant for app developers. You can define your own custom packets and IBCModule callbacks.

The following requirements must be met for a module to interact over IBC:

- Bind to one or more ports

- Define the packet data

- Define optional acknowledgement structures and methods to encode and decode them

- Implement the IBCModule interface

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

IBC is intended to work in execution environments where modules do not necessarily trust each other. IBC must authenticate module actions on ports and channels so that only modules with the appropriate permissions can use the channels. This security is accomplished using [dynamic capabilities](../architecture/adr-003-dynamic-capability-store.md). Upon binding to a port or creating a channel for a module, IBC returns a dynamic capability that the module must claim to use that port or channel. This binding strategy prevents other modules from using that port or channel since those modules do not own the appropriate capability.

While this explanation is useful background information, IBC modules do not need to interact at all with these lower-level abstractions. The relevant abstraction layer for IBC application developers is that of channels and ports.

Write your IBC applications as self-contained **modules**. A module on one blockchain can communicate with other modules on other blockchains by sending, receiving, and acknowledging packets through channels that are uniquely identified by the `(channelID, portID)` tuple.

A useful analogy is to consider IBC modules as internet apps on a computer. A channel can then be conceptualized as an IP connection, with the IBC portID is like an IP port, and the IBC channelID is like an IP address. A single instance of an IBC module can communicate on the same port with any number of other modules and IBC correctly routes all packets to the relevant module using the `(channelID, portID)` tuple. An IBC module can also communicate with another IBC module over multiple ports by sending each `(portID<->portID)` packet stream on a different unique channel.

### [Ports](https://github.com/cosmos/ibc-go/blob/main/modules/core/05-port)

An IBC module can bind to any number of ports. Each port must be identified by a unique `portID`. Since IBC is designed to be secure with mutually-distrusted modules that operate on the same ledger, binding a port returns the dynamic object capability. To take action on a particular port, for example, to open a channel with its portID, a module must provide the dynamic object capability to the IBC handler. This requirement prevents a malicious module from opening channels with ports it does not own.

IBC modules are responsible for claiming the capability that is returned on `BindPort`.

### [Channels](https://github.com/cosmos/ibc-go/blob/main/modules/core/04-channel)

An IBC channel can be established between two IBC ports. A port is exclusively owned by a single module. IBC packets are sent over channels. Just as IP packets contain the destination IP address, IP port, the source IP address, and source IP port, IBC packets contain the destination portID, channelID, the source portID, and channelID. The IBC packets enable IBC to correctly route the packets to the destination module, while also allowing modules receiving packets to know the sender module.

- A channel can be `ORDERED` so that packets from a sending module must be processed by the receiving module in the order they were sent.

- Recommended, a channel may be `UNORDERED` so that packets from a sending module are processed in the order they arrive, which may not be the order the packets were sent.

Modules may choose which channels they wish to communicate over with. IBC expects modules to implement callbacks that are called during the channel handshake. These callbacks may do custom channel initialization logic. If an error is returned, the channel handshake fails. By returning errors on callbacks, modules can programmatically reject and accept channels.

The channel handshake is a 4-step handshake. Briefly, if a given chain A wants to open a channel with chain B using an already established connection:

1. Chain A sends a `ChanOpenInit` message to signal a channel initialization attempt with chain B.
2. Chain B sends a `ChanOpenTry` message to try opening the channel on chain A.
3. Chain A sends a `ChanOpenAck` message to mark its channel end status as open.
4. Chain B sends a `ChanOpenConfirm` message to mark its channel end status as open.

If all of these actions happen successfully, the channel is open on both sides. At each step in the handshake, the module associated with the `ChannelEnd` executes its callback for that step of the handshake. So on `ChanOpenInit`, the module on chain A has its callback `OnChanOpenInit` executed.

Just as ports came with dynamic capabilities, channel initialization returns a dynamic capability that the module **must** claim so that they can pass in a capability to authenticate channel actions like sending packets. The channel capability is passed into the callback on the first parts of the handshake: `OnChanOpenInit` on the initializing chain or `OnChanOpenTry` on the other chain.

### [Packets](https://github.com/cosmos/ibc-go/blob/main/modules/core/04-channel)

Modules communicate with each other by sending packets over IBC channels. All IBC packets contain:

- Destination `portID`

- Destination `channelID`

- Source `portID`

- Source `channelID`

  These port and channels allow the modules to know the sender module of a given packet.

- A sequence to optionally enforce ordering

- `TimeoutTimestamp` and `TimeoutHeight`

  When non-zero, these timeout values determine the deadline before which the receiving module must process a packet.

  If the timeout passes without the packet being successfully received, the sending module can timeout the packet and take appropriate actions.

Modules send custom application data to each other inside the `Data []byte` field of the IBC packet. Packet data is completely opaque to IBC handlers. The sender module must encode their application-specific packet information into the `Data` field of packets. The receiver module must decode that `Data` back to the original application data.

### [Receipts and Timeouts](https://github.com/cosmos/ibc-go/blob/main/modules/core/04-channel)

Since IBC works over a distributed network and relies on potentially faulty relayers to relay messages between ledgers, IBC must handle the case where a packet does not get sent to its destination in a timely manner or at all. Packets must specify a timeout height or timeout timestamp after which a packet can no longer be successfully received on the destination chain.

If the timeout is reached, then a proof-of-packet timeout can be submitted to the original chain which can then perform application-specific logic to timeout the packet, perhaps by rolling back the packet send changes (refunding senders any locked funds, and so on).

In ORDERED channels, a timeout of a single packet in the channel closes the channel. If packet sequence `n` times out, then no packet at sequence `k > n` can be successfully received without violating the contract of ORDERED channels that packets are processed in the order that they are sent. Since ORDERED channels enforce this invariant, a proof that sequence `n` hasn't been received on the destination chain by packet `n`'s specified timeout is sufficient to timeout packet `n` and close the channel.

In the UNORDERED case, packets can be received in any order. IBC writes a packet receipt for each sequence it has received in the UNORDERED channel. This receipt contains no information and is simply a marker intended to signify that the UNORDERED channel has received a packet at the specified sequence. To timeout a packet on an UNORDERED channel, proof that a packet receipt does not exist is required for the packet's sequence by the specified timeout. Of course, timing out a packet on an UNORDERED channel triggers the application specific timeout logic for that packet, and does not close the channel.

For this reason, most modules that use UNORDERED channels are recommended as they require less liveness guarantees to function effectively for users of that channel.

### [Acknowledgements](https://github.com/cosmos/ibc-go/blob/main/modules/core/04-channel)

Modules also write application-specific acknowledgements when processing a packet. Acknowledgements can be done:

- Synchronously on `OnRecvPacket` if the module processes packets as soon as they are received from IBC module.

- Asynchronously if module processes packets at some later point after receiving the packet.

This acknowledgement data is opaque to IBC much like the packet `Data` and is treated by IBC as a simple byte string `[]byte`. The receiver modules must encode their acknowledgement so that the sender module can decode it correctly. How the acknowledgement is encoded should be decided through version negotiation during the channel handshake.

The acknowledgement can encode whether the packet processing succeeded or failed, along with additional information that allows the sender module to take appropriate action.

After the acknowledgement has been written by the receiving chain, a relayer relays the acknowledgement back to the original sender module which then executes application-specific acknowledgment logic using the contents of the acknowledgement. This acknowledgement can involve rolling back packet-send changes in the case of a failed acknowledgement (refunding senders).

After an acknowledgement is received successfully on the original sender the chain, the IBC module deletes the corresponding packet commitment as it is no longer needed.

## Further Readings and Specs

To learn more about IBC, check out the following specifications:

- [IBC specs](https://github.com/cosmos/ibc/tree/master/spec)
- [IBC protocol on the Cosmos SDK](https://github.com/cosmos/ibc-go/blob/main/docs/spec.md)

## Next {hide}

Learn about how to [integrate](./integration.md) IBC to your application {hide}
