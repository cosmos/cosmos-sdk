<!--
order: 4
-->

# Messages

In this section we describe the processing of the IBC messages and the corresponding updates to the state.

## ICS 02 - Client

### MsgCreateClient

A light client is created using the `MsgCreateClient`.

```go
type MsgCreateClient struct {
  ClientID        string
  ClientType      string
  ConsensusState  ConsensusState
  Signer          AccAddress
}
```

This message is expected to fail if:

- `ClientID` is invalid (not alphanumeric or not within 10-20 characters)
- `ClientType` is not registered
- `ConsensusState` is empty
- `Signer` is empty
- A light client with the provided id and type already exist

The message creates and stores a light client with the given ID and consensus type,
stores the validator set as the `Commiter` of the given consensus state and stores
both the consensus state and its commitment root (i.e app hash).

### MsgUpdateClient

A light client is updated with a new header using the `MsgUpdateClient`.

```go
type MsgUpdateClient struct {
  ClientID string
  Header   Header
  Signer   AccAddress
}
```

This message is expected to fail if:

- `ClientID` is invalid (not alphanumeric or not within 10-20 characters)
- `Header` is empty
- `Signer` is empty
- A Client hasn't been created for the given ID
- the header's client type is different from the registered one
- the client is frozen due to misbehaviour and cannot be updated

The message validates the header and updates the consensus state with the new
height, commitment root and validator sets, which are then stored.

## ICS 03 - Connection

## ICS 04 - Channels

### MsgChannelOpenInit

A channel handshake is initiated by a chain A using the `MsgChannelOpenInit`
message.

```go
type MsgChannelOpenInit struct {
  PortID    string
  ChannelID string
  Channel   Channel
  Signer    sdk.AccAddress
}
```

This message is expected to fail if:

## ICS 20 - Fungible Token Transfer
