<!--
order: 1
-->

# Concepts

> NOTE: if you are not familiar with the IBC terminology and concepts, please read
this [document](https://github.com/cosmos/ics/blob/master/ibc/1_IBC_TERMINOLOGY.md) as prerequisite reading.

## Client Creation, Updates, and Upgrades

IBC clients are on chain light clients. The light client is responsible for verifying
counterparty state. A light client can be created by any user submitting a client
identifier and a valid initial `ClientState` and `ConsensusState`. The client identifier
must not already be used. Clients are given a client identifier prefixed store to
store their associated client state and consensus states. Consensus states are 
stored using its associated height. 

Clients can be updated by any user submitting a valid `Header`. The client state callback
to `CheckHeaderAndUpdateState` is responsible for verifying the header against previously
stored state. The function should also return the updated client state and consensus state 
if the header is considered a valid update. A light client, such as Tendermint, may have
client specific parameters like `TrustLevel` which must be considered valid in relation
to the `Header`.

Clients may be upgraded. The upgrade should be verified using `VerifyUpgrade`. It is not
a requirement to allow for light client upgrades. For example, the solo machine client 
will simply return an error on `VerifyUpgrade`.

## Client Misbehaviour

IBC clients must freeze when the counterparty chain becomes malicious and 
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
   VersionNumber uint64
   VersionHeight  uint64
}
```

The `VersionNumber` represents the version of the chain that the height is representing.
An version typically represents a continuous, monotonically increasing range of block-heights.
The `VersionHeight` represents the height of the chain within the given version.

On any reset of the `VersionHeight`, for example, when hard-forking a Tendermint chain,
the `VersionNumber` will get incremented. This allows IBC clients to distinguish between a
block-height `n` of a previous version of the chain (at version `p`) and block-height `n` of the current
version of the chain (at version `e`).

`Heights` that share the same version number can be compared by simply comparing their respective `VersionHeights`.
Heights that do not share the same version number will only be compared using their respective `VersionNumbers`.
Thus a height `h` with version number `e+1` will always be greater than a height `g` with version number `e`,
**REGARDLESS** of the difference in version heights.

Ex:

```go
Height{VersionNumber: 3, VersionHeight: 0} > Height{VersionNumber: 2, VersionHeight: 100000000000}
```

When a Tendermint chain is running a particular version, relayers can simply submit headers and proofs with the version number
given by the chain's chainID, and the version height given by the Tendermint block height. When a chain updates using a hard-fork 
and resets its block-height, it is responsible for updating its chain-id to increment the version number.
IBC Tendermint clients then verifies the version number against their `ChainId` and treat the `VersionHeight` as the Tendermint block-height.

Tendermint chains wishing to use versions to maintain persistent IBC connections even across height-resetting upgrades must format their chain-ids
in the following manner: `{chainID}-{version_number}`. On any height-resetting upgrade, the chainID **MUST** be updated with a higher version number
than the previous value.

Ex:

- Before upgrade ChainID: `gaiamainnet-3`
- After upgrade ChainID: `gaiamainnet-4`

Clients that do not require versions, such as the solo-machine client, simply hardcode `0` into the version number whenever they
need to return an IBC height when implementing IBC interfaces and use the `VersionHeight` exclusively.

Other client-types may implement their own logic to verify the IBC Heights that relayers provide in their `Update`, `Misbehavior`, and
`Verify` functions respectively.

The IBC interfaces expect an `ibcexported.Height` interface, however all clients should use the concrete implementation provided in
`02-client/types` and reproduced above.

## Connection Handshake

## Connection Version Negotiation

During the handshake procedure for connections a version string is agreed
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

A valid connection version is considered to be in the following format:
`(version-identifier,[feature-0,feature-1])`

- the version tuple must be enclosed in parentheses
- the feature set must be enclosed in brackets
- there should be no space between the comma separating the identifier and the
  feature set
- the version identifier must no contain any commas
- each feature must not contain any commas
- each feature must be separated by commas

::: warning
A set of versions should not contain two versions with the same
identifier, but differing feature sets. This will result in undefined behavior
with regards to version selection in `ConnOpenTry`. Each version in a set of
versions should have a unique version identifier.
:::

## Channel Handshake

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

A packet may be associated with one of the following states:
- packet does not exist (ie it has not been sent)
- packet has been sent but not received (the packet commitment exists on the 
sending chain, but no receipt exists on the receiving chain)
- packet has been received but not acknowledged (packet commitment exists 
on the sending chain, a receipt exists on the receiving chain, but no acknowledgement
exists on the receiving chain)
- packet has been acknowledgement but the acknowledgement has not been relayed 
(the packet commitment exists on the sending chain, the receipt and acknowledgement
exist on the receiving chain)
- packet has completed its life cycle (the packet commitment does not exist on
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
state if possible. The verification will fail if the lastest state of the
light client does not include the packet commitment. The receiving chain
is responsible for verifying that the counterparty set the hash of the 
packet. If verification of the packet to be received is successful, the
receiving chain should store a receipt of the packet and call application
logic if necessary. An acknowledgement may be stored at this time or at 
another point in the future aynchronously. 

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

Future optimizations may allow for storage cleanup of stored packet 
commitments and acknowledgements that no longer provide any usefulness. 
This may be from packets that have completed their life cycles or from
channels which do not require written acknowledgements. 

## Timing out Packets

A packet may be timed out on the receiving chain if the packet timeouts have
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

## Port and Channel Capabilities

## Hostname Validation

## Proofs
