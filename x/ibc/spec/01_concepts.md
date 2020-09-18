<!--
order: 1
-->

# Concepts

> NOTE: if you are not familiar with the IBC terminology and concepts, please read
this [document](https://github.com/cosmos/ics/blob/master/ibc/1_IBC_TERMINOLOGY.md) as prerequisite reading.

## IBC Client Heights

IBC Client Heights are represented by the struct:

```go
type Height struct {
   EpochNumber uint64
   EpocHeight  uint64
```

The `EpochNumber` represents the epoch of the chain that the height is representing.
An epoch typically represents a continuous, monotonically increasing range of block-heights.
The `EpochHeight` represents the height of the chain within the given epoch.

On any reset of the `EpochHeight`, for example, when hard-forking a Tendermint chain,
the `EpochNumber` will get incremented. This allows IBC clients to distinguish between a
block-height `n` of a previous version of the chain (at epoch `p`) and block-height `n` of the current
version of the chain (at epoch `e`).

`Heights` that share the same epoch number can be compared by simply comparing their respective `EpochHeights`.
Heights that do not share the same epoch number will only be compared using their respective `EpochNumbers`.
Thus a height `h` with epoch number `e+1` will always be greater than a height `g` with epoch number `e`,
**REGARDLESS** of the difference in epoch heights.

Ex: `Height{EpochNumber: 3, EpochHeight: 0} > Height{EpochNumber: 2, EpochHeight: 100000000000}`

When a Tendermint chain is running a particular version, relayers can simply submit headers and proofs with the epoch number
given by the chain's chainID, and the epoch height given by the Tendermint block height. When a chain updates using a hard-fork 
and resets its block-height, it is responsible for updating its chain-id to increment the epoch number.
IBC Tendermint clients then verifies the epoch number against their `ChainId` and treat the `EpochHeight` as the Tendermint block-height.

Tendermint chains wishing to use epochs to maintain persistent IBC connections even across height-resetting upgrades must format their chain-ids
in the following manner: `{chainID}-{version}`. On any height-resetting upgrade, the chainID **MUST** be updated with a higher epoch number
than the previous value.

Ex:
Before upgrade ChainID: `gaiamainnet-3`
After upgrade ChainID: `gaiamainnet-4`

Clients that do not require epochs, such as the solo-machine client, simply hardcode `0` into the epoch number whenever they 
need to return an IBC height when implementing IBC interfaces and use the `EpochHeight` exclusively.

Other client-types may implement their own logic to verify the IBC Heights that relayers provide in their Update, Misbehavior, and 
Verify functions respectively.

The IBC interfaces expect an `ibcexported.Height` interface, however all clients should use the concrete implementation provided in
`02-client/types` and reproduced above.

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

## Connection Version Negotation

During the handshake procedure for connections a version string is agreed
upon between the two parties. This occurs during the first 3 steps of the
handshake.

During `ConnOpenInit`, party A is expected to set all the versions they wish
to support within their connection state. It is expected that this set of
versions is from most preferred to least preferred. This is not a strict
requirement for the SDK implementation of IBC because the party calling
`ConnOpenTry` will greedily select the latest version it supports that the
counterparty supports as well.

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

## Channel Version Negotation

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

