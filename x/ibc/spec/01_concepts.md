<!--
order: 1
-->

# Concepts

> NOTE: if you are not familiar with the IBC terminology and concepts, please read
this [document](https://github.com/cosmos/ics/blob/master/ibc/1_IBC_TERMINOLOGY.md) as prerequisite reading.

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
