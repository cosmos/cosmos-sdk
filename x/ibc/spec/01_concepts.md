<!--
order: 1
-->

# Concepts

> NOTE: if you are not familiar with the IBC terminology and concepts, please read
this [document](https://github.com/cosmos/ics/blob/master/ibc/1_IBC_TERMINOLOGY.md) as prerequisite reading.

### Connection Version Negotation

During the handshake procedure for connections a version string is agreed upon between the two parties.
This occurs during the first 3 steps of the handshake.

During OpenInit, party A is expected to set all the versions they wish to support within their connection state.

During OpenTry, party B will select a version from the counterparty's supported versions. 
Priority will be placed on the latest supported version.
If a matching version cannot be found an error is returned.

During OpenAck, party A will verify that they can support the version party B selected.
If they do not support the selected version an error is returned.
After this step, the connection version is considered agreed upon.


