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
  ClientId        string
  ClientState     *types.Any // proto-packed client state
  ConsensusState  *types.Any // proto-packed consensus state
  Signer          sdk.AccAddress
}
```

This message is expected to fail if:

- `ClientId` is invalid (see naming requirements)
- `ClientState` is empty or invalid
- `ConsensusState` is empty or invalid
- `Signer` is empty
- A light client with the provided id and type already exist

The message creates and stores a light client with the given ID and consensus type,
stores the validator set as the `Commiter` of the given consensus state and stores
both the consensus state and its commitment root (i.e app hash).

### MsgUpdateClient

A light client is updated with a new header using the `MsgUpdateClient`.

```go
type MsgUpdateClient struct {
  ClientId string
  Header   *types.Any // proto-packed header
  Signer   sdk.AccAddress
}
```

This message is expected to fail if:

- `ClientId` is invalid (not alphanumeric or not within 10-20 characters)
- `Header` is empty or invalid
- `Signer` is empty
- A `ClientState` hasn't been created for the given ID
- the header's client type is different from the registered one
- the client is frozen due to misbehaviour and cannot be updated

The message validates the header and updates the consensus state with the new
height, commitment root and validator sets, which are then stored.

### MsgSubmitMisbehaviour

Submit a evidence of light client misbehaviour to freeze the client state and prevent additional packets from being relayed.

```go
type MsgSubmitMisbehaviour struct {
  ClientId     string
  Misbehaviour *types.Any // proto-packed misbehaviour
  Signer       sdk.AccAddress
}
```

This message is expected to fail if:

- `ClientId` is invalid (not alphanumeric or not within 10-20 characters)
- `Misbehaviour` is empty or invalid
- `Signer` is empty
- A `ClientState` hasn't been created for the given ID
- `Misbehaviour` check failed

The message validates the header and updates the consensus state with the new
height, commitment root and validator sets, which are then stored.

## ICS 03 - Connection

### MsgConnectionOpenInit

A connection is initialized on a light client using the `MsgConnectionOpenInit`.

```go
type MsgConnectionOpenInit struct {
	ClientId     string                                       
	ConnectionId string                                        
	Counterparty Counterparty                                  
	Version      string
	Signer       sdk.AccAddress
}
```

This message is expected to fail if:
- `ClientId` is invalid (see naming requirements)
- `ConnectionId` is invalid (see naming requirements)
- `Counterparty` is empty
- 'Version' is not empty and invalid
- `Signer` is empty
- A Client hasn't been created for the given ID
- A Connection for the given ID already exists

The message creates a connection for the given ID with an INIT state.

### MsgConnectionOpenTry

When a counterparty connection is initialized then a connection is initialized on a light client
using the `MsgConnectionOpenTry`.

```go
type MsgConnectionOpenTry struct {
	ClientId                       string
	DesiredConnectionId            string
	CounterpartyChosenConnectionId string
	ClientState                    *types.Any // proto-packed counterparty client
	Counterparty                   Counterparty
	CounterpartyVersions           []string
	ProofHeight                    Height
	ProofInit                      []byte
	ProofClient                    []byte
	ProofConsensus                 []byte
	ConsensusHeight                Height
	Signer                         sdk.AccAddress
}
```

This message is expected to fail if:

- `ClientId` is invalid (see naming requirements)
- `DesiredConnectionId` is invalid (see naming requirements)
- `CounterpartyChosenConnectionId` is not empty and doesn't match `DesiredConnectionId`
- `ClientState` is not a valid client of the executing chain
- `Counterparty` is empty
- `CounterpartyVersions` is empty
- `ProofHeight` is zero
- `ProofInit` is empty
- `ProofClient` is empty
- `ProofConsensus` is empty
- `ConsensusHeight` is zero
- `Signer` is empty
- A Client hasn't been created for the given ID
- A Connection for the given ID already exists
- `ProofInit` does not prove that the counterparty connection is in state INIT
- `ProofClient` does not prove that the counterparty has stored the `ClientState` provided in message
- `ProofConsensus` does not prove that the counterparty has the correct consensus state for this chain

The message creates a connection for the given ID with an TRYOPEN State. The `CounterpartyChosenConnectionID` 
represents the connection ID the counterparty set under `connection.Counterparty.ConnectionId`
to represent the connection ID this chain should use. An empty string indicates the connection
identifier is flexible and gives this chain an opportunity to choose its own identifier.

### MsgConnectionOpenAck

When a counterparty connection is initialized then a connection is opened on a light client 
using the `MsgConnectionOpenAck`.

```go
type MsgConnectionOpenAck struct {
	ConnectionId             string
	CounterpartyConnectionId string 
	Version                  string
	ClientState              *types.Any // proto-packed counterparty client
	ProofHeight              Height
	ProofTry                 []byte
	ProofClient              []byte
	ProofConsensus           []byte
	ConsensusHeight          Height
	Signer                   sdk.AccAddress
}
```

This message is expected to fail if:

- `ConnectionId` is invalid (see naming requirements)
- `CounterpartyConnectionId` is invalid (see naming requirements)
- `Version` is empty
- `ClientState` is not a valid client of the executing chain
- `ProofHeight` is zero
- `ProofTry` is empty
- `ProofClient` is empty
- `ProofConsensus` is empty
- `ConsensusHeight` is zero
- `Signer` is empty
- `ProofTry` does not prove that the counterparty connection is in state TRYOPEN
- `ProofClient` does not prove that the counterparty has stored the `ClientState` provided by message
- `ProofConsensus` does not prove that the counterparty has the correct consensus state for this chain

The message sets the connection state for the given ID to OPEN. `CounterpartyConnectionId`
should be the `ConnectionId` used by the counterparty connection.

### MsgConnectionOpenConfirm

When a counterparty connection is opened then a connection is opened on a light client using
the `MsgConnectionOpenConfirm`.

```go
type MsgConnectionOpenConfirm struct {
  ConnectionId string
	ProofAck     []byte
	ProofHeight  Height
	Signer       sdk.AccAddress
}
```

This message is expected to fail if:

- `ConnectionId` is invalid (see naming requirements)
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
  PortId    string
  ChannelId string
  Channel   Channel
  Signer    sdk.AccAddress
}
```

This message is expected to fail if:

- `PortId` is invalid (see naming requirements)
- `ChannelId` is invalid (see naming requirements)
- `Channel` is empty
- `Signer` is empty
- A Channel End exists for the given Channel ID and Port ID

The message creates a channel on chain A with an INIT state for the given Channel ID
and Port ID.

### MsgChannelOpenTry

A channel handshake initialization attempt is acknowledged by a chain B using
the `MsgChannelOpenTry` message.

```go
type MsgChannelOpenTry struct {
	PortId                      string    
	DesiredChannelId            string   
	CounterpartyChosenChannelId string 
	Channel                     Channel 
	CounterpartyVersion         string 
	ProofInit                   []byte
	ProofHeight                 Height
	Signer                      sdk.AccAddress 
}
```

This message is expected to fail if:

- `PortId` is invalid (see naming requirements)
- `DesiredChannelId` is invalid (see naming requirements)
- `CounterpartyChosenChannelId` is not empty and not equal to `ChannelId`
- `Channel` is empty
- `CounterpartyVersion` is empty
- `ProofInit` is empty
- `ProofHeight` is zero
- `Signer` is empty
- A Channel End exists for the given Channel and Port ID
- `ProofInit` does not prove that the counterparty's Channel state is in INIT

The message creates a channel on chain B with an TRYOPEN state for the given Channel ID 
and Port ID. The `CounterpartyChosenChannelId` represents the channel ID the counterparty set under
`connection.Counterparty.ChannelId` to represent the channel ID this chain should use.
An empty string indicates the channel identifier is flexible and gives this chain an
opportunity to choose its own identifier.


### MsgChannelOpenAck

A channel handshake is opened by a chain A using the `MsgChannelOpenAck` message.

```go
type MsgChannelOpenAck struct {
	PortId                string
	ChannelId             string
	CounterpartyChannelId string 
	CounterpartyVersion   string
	ProofTry              []byte
	ProofHeight           Height
	Signer                sdk.AccAddress
}
```

This message is expected to fail if:

- `PortId` is invalid (see naming requirements)
- `ChannelId` is invalid (see naming requirements)
- `CounterpartyChannelId` is invalid (see naming requirements)
- `CounterpartyVersion` is empty
- `ProofTry` is empty
- `ProofHeight` is zero
- `Signer` is empty
- `ProofTry` does not prove that the counterparty's Channel state is in TRYOPEN

The message sets a channel on chain A to state OPEN for the given Channel ID and Port ID.
`CounterpartyChannelId` should be the `ChannelId` used by the counterparty channel.

### MsgChannelOpenConfirm

A channel handshake is confirmed and opened by a chain B using the `MsgChannelOpenConfirm`
message.

```go
type MsgChannelOpenConfirm struct {
	PortId              string
	ChannelId           string
	ProofAck            []byte
	ProofHeight         Height
	Signer              sdk.AccAddress
}
```

This message is expected to fail if:

- `PortId` is invalid (see naming requirements)
- `ChannelId` is invalid (see naming requirements)
- `ProofAck` is empty
- `ProofHeight` is zero
- `Signer` is empty
- `ProofAck` does not prove that the counterparty's Channel state is in OPEN

The message sets a channel on chain B to state OPEN for the given Channel ID and Port ID.

### MsgChannelCloseInit

A channel is closed on chain A using the `MsgChannelCloseInit`.

```go
type MsgChannelCloseInit struct {
		PortId    string
		ChannelId string  
		Signer    sdk.AccAddress
}
```

This message is expected to fail if:

- `PortId` is invalid (see naming requirements)
- `ChannelId` is invalid (see naming requirements)
- `Signer` is empty
- A Channel for the given Port ID and Channel ID does not exist or is already closed

The message closes a channel on chain A for the given Port ID and Channel ID.

### MsgChannelCloseConfirm

A channel is closed on chain B using the `MsgChannelCloseConfirm`.

```go
type MsgChannelCloseConfirm struct {
	PortId      string
	ChannelId   string
	ProofInit   []byte  
	ProofHeight Height
	Signer      sdk.AccAddress
}
```

This message is expected to fail if:

- `PortId` is invalid (see naming requirements)
- `ChannelId` is invalid (see naming requirements)
- `ProofInit` is empty
- `ProofHeight` is zero
- `Signer` is empty
- A Channel for the given Port ID and Channel ID does not exist or is already closed
- `ProofInit` does not prove that the counterparty set its channel to state CLOSED

The message closes a channel on chain B for the given Port ID and Channel ID.

### MsgRecvPacket

A packet is received on chain B using the `MsgRecvPacket`.

```go
type MsgRecvPacket struct {
    Packet      Packet
    Proof       []byte
    ProofHeight Height
    Signer      sdk.AccAddress
}
```

This message is expected to fail if:

- `Proof` is empty
- `ProofHeight` is zero
- `Signer` is empty
- `Packet` fails basic validation
- `Proof` does not prove that the counterparty sent the `Packet`.

The message receives a packet on chain B.

### MsgTimeout

A packet is timed out on chain A using the `MsgTimeout`.

```go
type MsgTimeout struct {
    Packet           Packet
    Proof            []byte
    ProofHeight      Height
    NextSequenceRecv uint64
    Signer           sdk.AccAddress
}
```

This message is expected to fail if:

- `Proof` is empty
- `ProofHeight` is zero
- `NextSequenceRecv` is zero
- `Signer` is empty
- `Packet` fails basic validation
- `Proof` does not prove that the packet has not been received on the counterparty chain.

The message times out a packet on chain B.

### MsgTimeoutOnClose

A packet is timed out on chain A due to the closure of the channel end on chain B using 
the `MsgTimeoutOnClose`.

```go
type MsgTimeoutOnClose struct {
    Packet           Packet
    Proof            []byte
    ProofClose       []byte
    ProofHeight      Height
    NextSequenceRecv uint64
    Signer           sdk.AccAddress
}
```

This message is expected to fail if:

- `Proof` is empty
- `ProofClose` is empty
- `ProofHeight` is zero
- `NextSequenceRecv` is zero
- `Signer` is empty
- `Packet` fails basic validation
- `Proof` does not prove that the packet has not been received on the counterparty chain.
- `ProofClose` does not prove that the counterparty channel end has been closed.

The message times out a packet on chain B.

### MsgAcknowledgement

A packet is acknowledged on chain A using the `MsgAcknowledgement`.

```go
type MsgAcknowledgement struct {
    Packet          Packet
    Acknowledgement []byte
    Proof           []byte
    ProofHeight     Height
    Signer          sdk.AccAddress
}
```

This message is expected to fail if:

- `Proof` is empty
- `ProofHeight` is zero
- `Signer` is empty
- `Packet` fails basic validation
- `Acknowledgement` is empty
- `Proof` does not prove that the counterparty received the `Packet`.

The message receives a packet on chain A.
