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
  Signer          sdk.AccAddress
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
  Signer   sdk.AccAddress
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

### MsgConnectionOpenInit

A connection is initialized on a light client using the `MsgConnectionOpenInit`.

```go
type MsgConnectionOpenInit struct {
	ClientID     string                                       
	ConnectionID string                                        
	Counterparty Counterparty                                  
	Signer       sdk.AccAddress
}
```

This message is expected to fail if:
- `ClientID` is invalid (see naming requirements)
- `ConnectionID` is invalid (see naming requirements)
- `Counterparty` is empty
- `Signer` is empty
- A Client hasn't been created for the given ID
- A Connection for the given ID already exists

The message creates a connection for the given ID with an INIT state.

### MsgConnectionOpenTry

When a counterparty connection is initialized then a connection is initialized on a light client
using the `MsgConnectionOpenTry`.

```go
type MsgConnectionOpenTry struct {
	ClientID             string       
	ConnectionID         string      
	Counterparty         Counterparty 
	CounterpartyVersions []string     
	ProofInit            []byte 
	ProofHeight          uint64 
	ProofConsensus       []byte   
	ConsensusHeight      uint64  
	Signer               sdk.AccAddress 
}
```

This message is expected to fail if:
- `ClientID` is invalid (see naming requirements)
- `ConnectionID` is invalid (see naming requirements)
- `Counterparty` is empty
- `CounterpartyVersions` is empty 
- `ProofInit` is empty
- `ProofHeight` is zero
- `ProofConsensus` is empty
- `ConsensusHeight` is zero
- `Signer` is empty
- A Client hasn't been created for the given ID
- A Connection for the given ID already exists
- `ProofInit` does not prove that the counterparty connection is in state INIT
- `ProofConsensus` does not prove that the counterparty has the correct consensus state for this chain

The message creates a connection for the given ID with an INIT State.

### MsgConnectionOpenAck

When a counterparty connection is initialized then a connection is opened on a light client 
using the `MsgConnectionOpenAck`.

```go
type MsgConnectionOpenAck struct {
	ConnectionID    string 
	Version         string 
	ProofTry        []byte 
	ProofHeight     uint64 
	ProofConsensus  []byte      
	ConsensusHeight uint64     
	Signer          sdk.AccAddress 
}
```

This message is expected to fail if:
- `ConnectionID` is invalid (see naming requirements)
- `Version` is empty
- `ProofTry` is empty
- `ProofHeight` is zero
- `ProofConsensus` is empty
- `ConsensusHeight` is zero
- `Signer` is empty
- `ProofTry` does not prove that the counterparty connection is in state TRYOPEN
- `ProofConsensus` does not prove that the counterparty has the correct consensus state for this chain

The message sets the connection state for the given ID to OPEN.

### MsgConnectionOpenConfirm

When a counterparty connection is opened then a connection is opened on a light client using
the `MsgConnectionOpenConfirm`.

```go
type MsgConnectionOpenConfirm struct {
	ConnectionID string 
	ProofAck     []byte   
	ProofHeight  uint64    
	Signer       sdk.AccAddress 
}
```

This message is expected to fail if:
- `ConnectionID` is invalid (see naming requirements)
- `ProofAck` is empty
- `ProofHeight` is zero
- `Signer` is empty
- A Connection with the given ID does not exist
- `ProofAck` does not prove that the counterparty connection is in state OPEN

The message sets the connection state for the given ID to OPEN.

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
