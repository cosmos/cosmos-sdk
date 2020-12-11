<!--
order: 3
-->

# State Transitions

The described state transitions assume successful message exection. 

## Create Client

`MsgCreateClient` will initialize and store a `ClientState` and `ConsensusState` in the sub-store
created using a generated client identifier. 

## Update Client

`MsgUpdateClient` will update the `ClientState` and create a new `ConsensusState` for the 
update height.

## Misbehaviour

`MsgSubmitMisbehaviour` will freeze a client.

## Upgrade Client

`MsgUpgradeClient` will upgrade the `ClientState` and `ConsensusState` to the update chain level
parameters and if applicable will update to the new light client implementation. 

## Client Update Proposal

An Update Client Proposal will unfreeze a client and set an updated `ClientState` and a new
`ConsensusState`.

## Connection Open Init

`MsgConnectionOpenInit` will initialize a connection state in INIT.

## Connection Open Try

`MsgConnectionOpenTry` will initialize or update a connection state to be in TRYOPEN.

## Connection Open Ack

`MsgConnectionOpenAck` will update a connection state from INIT or TRYOPEN to be in OPEN.

## Connection Open Confirm

`MsgConnectionOpenAck` will update a connection state from TRYOPEN to OPEN.

## Channel Open Init

`MsgChannelOpenInit` will initialize a channel state in INIT. It will create a channel capability
and set all Send, Receive and Ack Sequences to 1 for the channel. 

## Channel Open Try

`MsgChannelOpenTry` will initialize or update a channel state to be in TRYOPEN. If the channel
is being initialized, It will create a channel capability and set all Send, Receive and Ack 
Sequences to 1 for the channel. 

## Channel Open Ack

`MsgChannelOpenAck` will update the channel state to OPEN. It will set the version and channel 
identifier for its counterparty.

## Channel Open Confirm

`MsgChannelOpenConfirm` will update the channel state to OPEN.

## Channel Close Init

`MsgChannelCloseInit` will update the channel state to CLOSED.

## Channel Close Confirm

`MsgChannelCloseConfirm` will update the channel state to CLOSED.

## Send Packet

A application calling `ChannelKeeper.SendPacket` will incremenet the next sequence send and set
a hash of the packet as the packet commitment. 

## Receive Packet

`MsgRecvPacket` will increment the next sequence receive for ORDERED channels and set a packet 
receipt for UNORDERED channels. 

## Write Acknowledgement

`WriteAcknowledgement` may be executed synchronously during the execution of `MsgRecvPacket` or 
asynchonously by an application module. It writes an acknowledgement to the store.

## Acknowledge Packet

`MsgAcknowledgePacket` deletes the packet commitment and for ORDERED channels increments next
sequences ack. 

## Timeout Packet

`MsgTimeoutPacket` deletes the packet commitment and for ORDERED channels sets the channel state
to CLOSED.

## Timeout Packet on Channel Closure

`MsgTimeoutOnClose` deletes the packet commitment and for ORDERED channels sets the channel state
to CLOSED.
