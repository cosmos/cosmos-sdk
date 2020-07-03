<!--
order: 1
-->

# Concepts

> NOTE: if you are not familiar with the IBC terminology and concepts, please read
this [document](https://github.com/cosmos/ics/blob/master/ibc/1_IBC_TERMINOLOGY.md) as prerequisite reading.

### Connection Version Negotation

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
- there should be no space between the comma separting the identifier and the
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
