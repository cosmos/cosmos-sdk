<!--
order: 1
-->

# Concepts

> NOTE: if you are not familiar with the IBC terminology and concepts, please read
this [document](https://github.com/cosmos/ics/blob/master/ibc/1_IBC_TERMINOLOGY.md) as prerequisite reading.

## Client Creation, Updates, and Upgrades

IBC clients are on chain light clients. The light client is responsible for verifying
counterparty state. A light client can be created by any user submitting a valid initial 
`ClientState` and `ConsensusState`. The client identifier is auto generated using the
client type and the global client counter appended in the format: `{client-type}-{N}`.
Clients are given a client identifier prefixed store to store their associated client 
state and consensus states. Consensus states are stored using their associated height. 

Clients can be updated by any user submitting a valid `Header`. The client state callback
to `CheckHeaderAndUpdateState` is responsible for verifying the header against previously
stored state. The function should also return the updated client state and consensus state 
if the header is considered a valid update. A light client, such as Tendermint, may have
client specific parameters like `TrustLevel` which must be considered valid in relation
to the `Header`. The update height is not necessarily the lastest height of the light
client. Updates may fill in missing consensus state heights.

Clients may be upgraded. The upgrade should be verified using `VerifyUpgrade`. It is not
a requirement to allow for light client upgrades. For example, the solo machine client 
will simply return an error on `VerifyUpgrade`. Clients which implement upgrades
are expected to account for, but not necessarily support, planned and unplanned upgrades.

## Client Misbehaviour

IBC clients must freeze when the counterparty chain becomes byzantine and 
takes actions that could fool the light client into accepting invalid state 
transitions. Thus, relayers are able to submit Misbehaviour proofs that prove 
that a counterparty chain has signed two Headers for the same height. This 
constitutes misbehaviour as the IBC client could have accepted either header 
as valid. Upon verifying the misbehaviour the IBC client must freeze at that 
height so that any proof verifications for the frozen height or later fail.

Note, there is a difference between the chain-level Misbehaviour that IBC is 
concerned with and the validator-level Evidence that Tendermint is concerned 
with. Tendermint must be able to detect, submit, and punish any evidence of 
individual validators breaking the Tendermint consensus protocol and attempting 
to mount an attack. IBC clients must only act when an attack is successful 
and the chain has successfully forked. In this case, valid Headers submitted 
to the IBC client can no longer be trusted and the client must freeze.

Governance may then choose to override a frozen client and provide the correct, 
canonical Header so that the client can continue operating after the Misbehaviour 
submission.

## ClientUpdateProposal

A governance proposal may be passed to update a specified client with a provided
header. This is useful in unfreezing clients or updating expired clients. Each 
client is expected to implement this functionality. A client may choose to disallow
an update by a governance proposal by returning an error in the client state function
'CheckProposedHeaderAndUpdateState'.

The localhost client cannot be updated by a governance proposal. 

The solo machine client requires the boolean flag 'AllowUpdateAfterProposal' to be set
to true in order to be updated by a proposal. This is set upon client creation and cannot 
be updated later.

The tendermint client has two flags update flags, 'AllowUpdateAfterExpiry' and 
'AllowUpdateAfterMisbehaviour'. The former flag can only be used to unexpire clients. The
latter flag can be used to unfreeze a client and if necessary it will also unexpire the client.
It is advised to let a client expire if it has become frozen before proposing a new header. 
This is to avoid the client from becoming refrozen if the misbehaviour evidence has not 
expired. These boolean flags are set upon client creation and cannot be updated later.

## IBC Client Heights

IBC Client Heights are represented by the struct:

```go
type Height struct {
   RevisionNumber uint64
   RevisionHeight  uint64
}
```

The `RevisionNumber` represents the revision of the chain that the height is representing.
An revision typically represents a continuous, monotonically increasing range of block-heights.
The `RevisionHeight` represents the height of the chain within the given revision.

On any reset of the `RevisionHeight`, for example, when hard-forking a Tendermint chain,
the `RevisionNumber` will get incremented. This allows IBC clients to distinguish between a
block-height `n` of a previous revision of the chain (at revision `p`) and block-height `n` of the current
revision of the chain (at revision `e`).

`Heights` that share the same revision number can be compared by simply comparing their respective `RevisionHeights`.
Heights that do not share the same revision number will only be compared using their respective `RevisionNumbers`.
Thus a height `h` with revision number `e+1` will always be greater than a height `g` with revision number `e`,
**REGARDLESS** of the difference in revision heights.

Ex:

```go
Height{RevisionNumber: 3, RevisionHeight: 0} > Height{RevisionNumber: 2, RevisionHeight: 100000000000}
```

When a Tendermint chain is running a particular revision, relayers can simply submit headers and proofs with the revision number
given by the chain's chainID, and the revision height given by the Tendermint block height. When a chain updates using a hard-fork 
and resets its block-height, it is responsible for updating its chain-id to increment the revision number.
IBC Tendermint clients then verifies the revision number against their `ChainId` and treat the `RevisionHeight` as the Tendermint block-height.

Tendermint chains wishing to use revisions to maintain persistent IBC connections even across height-resetting upgrades must format their chain-ids
in the following manner: `{chainID}-{revision_number}`. On any height-resetting upgrade, the chainID **MUST** be updated with a higher revision number
than the previous value.

Ex:

- Before upgrade ChainID: `gaiamainnet-3`
- After upgrade ChainID: `gaiamainnet-4`

Clients that do not require revisions, such as the solo-machine client, simply hardcode `0` into the revision number whenever they
need to return an IBC height when implementing IBC interfaces and use the `RevisionHeight` exclusively.

Other client-types may implement their own logic to verify the IBC Heights that relayers provide in their `Update`, `Misbehavior`, and
`Verify` functions respectively.

The IBC interfaces expect an `ibcexported.Height` interface, however all clients should use the concrete implementation provided in
`02-client/types` and reproduced above.

## Connection Handshake

The connection handshake occurs in 4 steps as defined in [ICS 03](https://github.com/cosmos/ics/tree/master/spec/ics-003-connection-semantics).

`ConnOpenInit` is the first attempt to initialize a connection on the executing chain. 
The handshake is expected to succeed if the version selected is supported. The connection 
identifier for the counterparty connection must be left empty indicating that the counterparty
must select its own identifier. The connection identifier is auto derived in the format: 
`connection{N}` where N is the next sequence to be used. The counter begins at 0 and increments
by 1. The connection is set and stored in the INIT state upon success.

`ConnOpenTry` is a response to a chain executing `ConnOpenInit`. The executing chain will validate
the chain level parameters the counterparty has stored such as its chainID. The executing chain 
will also verify that if a previous connection exists for the specified connection identifier 
that all the parameters match and its previous state was in INIT. This may occur when both 
chains execute `ConnOpenInit` simultaneously. If the connection does not exist then a connection
identifier is generated in the same format done in `ConnOpenInit`.  The executing chain will verify
that the counterparty created a connection in INIT state. The executing chain will also verify 
The `ClientState` and `ConsensusState` the counterparty stores for the executing chain. The 
executing chain will select a version from the intersection of its supported versions and the 
versions set by the counterparty. The connection is set and stored in the TRYOPEN state upon 
success. 

`ConnOpenAck` may be called on a chain when the counterparty connection has entered TRYOPEN. A
previous connection on the executing chain must exist in either INIT or TRYOPEN. The executing
chain will verify the version the counterparty selected. If the counterparty selected its own 
connection identifier, it will be validated in the basic validation of a `MsgConnOpenAck`. 
The counterparty connection state is verified along with the `ClientState` and `ConsensusState`
stored for the executing chain. The connection is set and stored in the OPEN state upon success.

`ConnOpenConfirm` is a response to a chain executing `ConnOpenAck`. The executing chain's connection
must be in TRYOPEN. The counterparty connection state is verified to be in the OPEN state. The 
connection is set and stored in the OPEN state upon success.

## Connection Version Negotiation

During the handshake procedure for connections a version is agreed
upon between the two parties. This occurs during the first 3 steps of the
handshake.

During `ConnOpenInit`, party A is expected to set all the versions they wish
to support within their connection state. It is expected that this set of
versions is from most preferred to least preferred. This is not a strict
requirement for the SDK implementation of IBC because the party calling
`ConnOpenTry` will greedily select the latest version it supports that the
counterparty supports as well. A specific version can optionally be passed
as `Version` to ensure that the handshake will either complete with that 
version or fail.

During `ConnOpenTry`, party B will select a version from the counterparty's
supported versions. Priority will be placed on the latest supported version.
If a matching version cannot be found an error is returned.

During `ConnOpenAck`, party A will verify that they can support the version
party B selected. If they do not support the selected version an error is
returned. After this step, the connection version is considered agreed upon.


A `Version` is defined as follows:

```go
type Version struct {
	// unique version identifier
	Identifier string 
	// list of features compatible with the specified identifier
	Features []string 
}
```

A version must contain a non empty identifier. Empty feature sets are allowed, but each 
feature must be a non empty string.

::: warning
A set of versions should not contain two versions with the same
identifier, but differing feature sets. This will result in undefined behavior
with regards to version selection in `ConnOpenTry`. Each version in a set of
versions should have a unique version identifier.
:::

## Channel Handshake

The channel handshake occurs in 4 steps as defined in [ICS 04](https://github.com/cosmos/ics/tree/master/spec/ics-004-channel-and-packet-semantics).

`ChanOpenInit` is the first attempt to initialize a channel on top of an existing connection. 
The handshake is expected to succeed if the version selected for the existing connection is a 
supported IBC version. The portID must correspond to a port already binded upon `InitChain`. 
The channel identifier for the counterparty channel must be left empty indicating that the 
counterparty must select its own identifier. The channel identifier is auto derived in the
format: `channel{N}` where N is the next sequence to be used. The channel is set and stored 
in the INIT state upon success. The channel parameters `NextSequenceSend`, `NextSequenceRecv`, 
and `NextSequenceAck` are all set to 1 and a channel capability is created for the given 
portID and channelID path. 

`ChanOpenTry` is a response to a chain executing `ChanOpenInit`. If the executing chain is calling
`ChanOpenTry` after previously executing `ChanOpenInit` then the provided channel parameters must
match the previously selected parameters. If the previous channel does not exist then a channel
identifier is generated in the same format as done in `ChanOpenInit`. The connection the channel 
is created on top of must be an OPEN state and its IBC version must support the desired channel 
type being created (ORDERED, UNORDERED, etc). The executing chain will verify that the channel 
state of the counterparty is in INIT. The executing chain will set and store the channel state 
in TRYOPEN. The channel parameters `NextSequenceSend`, `NextSequenceRecv`, and `NextSequenceAck` 
are all set to 1 and a channel capability is created for the given portID and channelID path only 
if the channel did not previously exist. 

`ChanOpenAck` may be called on a chain when the counterparty channel has entered TRYOPEN. A
previous channel on the executing chain must exist be in either INIT or TRYOPEN state. If the 
counterparty selected its own channel identifier, it will be validated in the basic validation 
of `MsgChanOpenAck`. The executing chain verifies that the counterparty channel state is in 
TRYOPEN. The channel is set and stored in the OPEN state upon success.

`ChanOpenConfirm` is a response to a chain executing `ChanOpenAck`. The executing chain's 
previous channel state must be in TRYOPEN. The executing chain verifies that the counterparty 
channel state is OPEN. The channel is set and stored in the OPEN state upon success.

## Channel Version Negotiation

During the channel handshake procedure a version must be agreed upon between
the two parties. The selection process is largely left to the callers and
the verification of valid versioning must be handled by application developers
in the channel handshake callbacks.

During `ChanOpenInit`, a version string is passed in and set in party A's
channel state.

During `ChanOpenTry`, a version string for party A and for party B are passed
in. The party A version string must match the version string used in
`ChanOpenInit` otherwise channel state verification will fail. The party B
version string could be anything (even different than the proposed one by
party A). However, the proposed version by party B is expected to be fully
supported by party A.

During the `ChanOpenAck` callback, the application module is expected to verify
the version proposed by party B using the `MsgChanOpenAck` `CounterpartyVersion`
field. The application module should throw an error if the version string is
not valid.

In general empty version strings are to be considered valid options for an 
application module.

Application modules may implement their own versioning system, such as semantic
versioning, or they may lean upon the versioning system used for in connection
version negotiation. To use the connection version semantics the application
would simply pass the proto encoded version into each of the handshake calls
and decode the version string into a `Version` instance to do version verification
in the handshake callbacks.

Implementations which do not feel they would benefit from versioning can do
basic string matching using a single compatible version.

## Sending, Receiving, Acknowledging Packets

Terminology:
**Packet Commitment** A hash of the packet stored on the sending chain.
**Packet Receipt** A single bit indicating that a packet has been received. 
Used for timeouts. 
**Acknowledgement** Data written to indicate the result of receiving a packet.
Typically conveying either success or failure of the receive.

A packet may be associated with one of the following states:
- the packet does not exist (ie it has not been sent)
- the packet has been sent but not received (the packet commitment exists on the 
sending chain, but no receipt exists on the receiving chain)
- the packet has been received but not acknowledged (packet commitment exists 
on the sending chain, a receipt exists on the receiving chain, but no acknowledgement
exists on the receiving chain)
- the packet has been acknowledgement but the acknowledgement has not been relayed 
(the packet commitment exists on the sending chain, the receipt and acknowledgement
exist on the receiving chain)
- the packet has completed its life cycle (the packet commitment does not exist on
the sending chain, but a receipt and acknowledgement exist on the receiving chain)

Sending of a packet is initiated by a call to the `ChannelKeeper.SendPacket` 
function by an application module. Packets being sent will be verified for
correctness (core logic only). If the packet is valid, a hash of the packet
will be stored as a packet commitment using the packet sequence in the key.
Packet commitments are stored on the sending chain.

A message should be sent to the receving chain indicating that the packet
has been committed on the sending chain and should be received on the 
receiving chain. The light client on the receiving chain, which verifies
the sending chain's state, should be updated to the lastest sending chain
state if possible. The verification will fail if the latest state of the
light client does not include the packet commitment. The receiving chain
is responsible for verifying that the counterparty set the hash of the 
packet. If verification of the packet to be received is successful, the
receiving chain should store a receipt of the packet and call application
logic if necessary. An acknowledgement may be processed and stored at this time (synchronously)
or at another point in the future (asynchronously). 

Acknowledgements written on the receiving chain may be verified on the 
sending chain. If the sending chain successfully verifies the acknowledgement
then it may delete the packet commitment stored at that sequence. There is
no requirement for acknowledgements to be written. Only the hash of the
acknowledgement is stored on the chain. Application logic may be executed
in conjunction with verifying an acknowledgement. For example, in fungible
cross-chain token transfer, a failed acknowledgement results in locked or
burned funds being refunded. 

Relayers are responsible for reconstructing packets between the sending, 
receiving, and acknowledging of packets. 

IBC applications sending and receiving packets are expected to appropriately
handle data contained within a packet. For example, cross-chain token 
transfers will unmarshal the data into proto definitions representing
a token transfer. 

Future optimizations may allow for storage cleanup. Stored packet 
commitments could be removed from channels which do not write
packet acknowledgements and acknowledgements could be removed
when a packet has completed its life cycle.

## Timing out Packets

A packet may be timed out on the receiving chain if the packet timeout height or timestamp has
been surpassed on the receving chain or the channel has closed. A timed out
packet can only occur if the packet has never been received on the receiving 
chain. ORDERED channels will verify that the packet sequence is greater than 
the `NextSequenceRecv` on the receiving chain. UNORDERED channels will verify 
that the packet receipt has not been written on the receiving chain. A timeout
on channel closure will additionally verify that the counterparty channel has 
been closed. A successful timeout may execute application logic as appropriate.

Both the packet's timeout timestamp and the timeout height must have been 
surpassed on the receiving chain for a timeout to be valid. A timeout timestamp 
or timeout height with a 0 value indicates the timeout field may be ignored. 
Each packet is required to have at least one valid timeout field. 

## Closing Channels

Closing a channel occurs in occurs in 2 handshake steps as defined in [ICS 04](https://github.com/cosmos/ics/tree/master/spec/ics-004-channel-and-packet-semantics).

`ChanCloseInit` will close a channel on the executing chain if the channel exists, it is not 
already closed and the connection it exists upon is OPEN. Channels can only be closed by a 
calling module or in the case of a packet timeout on an ORDERED channel.

`ChanCloseConfirm` is a response to a counterparty channel executing `ChanCloseInit`. The channel
on the executing chain will be closed if the channel exists, the channel is not already closed, 
the connection the channel exists upon is OPEN and the executing chain successfully verifies
that the counterparty channel has been closed.

## Port and Channel Capabilities

## Hostname Validation

Hostname validation is implemented as defined in [ICS 24](https://github.com/cosmos/ics/tree/master/spec/ics-024-host-requirements).

The 24-host sub-module parses and validates identifiers. It also builds 
the key paths used to store IBC related information. 

A valid identifier must conatin only alphanumeric characters or the 
following list of allowed characters: 
".", "\_", "+", "-", "#", "[", "]", "<", ">" 

- Client identifiers must contain between 9 and 64 characters.
- Connection identifiers must contain between 10 and 64 characters.
- Channel identifiers must contain between 10 and 64 characters.
- Port identifiers must contain between 2 and 64 characters.

## Proofs

Proofs for counterparty state validation are provided as bytes. These bytes 
can be unmarshaled into proto definitions as necessary by light clients.
For example, the Tendermint light client will use the bytes as a merkle 
proof where as the solo machine client will unmarshal the proof into
several layers proto definitions used for signature verficiation. 
