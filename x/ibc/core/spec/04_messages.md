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
  ClientState     *types.Any // proto-packed client state
  ConsensusState  *types.Any // proto-packed consensus state
  Signer          sdk.AccAddress
}
```

This message is expected to fail if:

- `ClientState` is empty or invalid
- `ConsensusState` is empty or invalid
- `Signer` is empty

The message creates and stores a light client with an initial consensus state using a generated client
identifier.

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
- The client is frozen due to misbehaviour and cannot be updated
- The header fails to provide a valid update for the client

The message validates the header and updates the client state and consensus state for the 
header height.

### MsgUpgradeClient
```go
type MsgUpgradeClient struct {
	ClientId      string 
	ClientState   *types.Any // proto-packed client state
	UpgradeHeight *Height 
	ProofUpgrade  []byte 
	Signer        string 
}
```

This message is expected to fail if:

- `ClientId` is invalid (not alphanumeric or not within 10-20 characters)
- `ClientState` is empty or invalid
- `UpgradeHeight` is empty or zero
- `ProofUpgrade` is empty
- `Signer` is empty
- A `ClientState` hasn't been created for the given ID
- The client is frozen due to misbehaviour and cannot be upgraded
- The upgrade proof fails 

The message upgrades the client state and consensus state upon successful validation of a
chain upgrade. 

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

The message verifies the misbehaviour and freezes the client. 

## ICS 03 - Connection

### MsgConnectionOpenInit

A connection is initialized on a light client using the `MsgConnectionOpenInit`.

```go
type MsgConnectionOpenInit struct {
	ClientId     string                                       
	Counterparty Counterparty                                  
	Version      string
	Signer       sdk.AccAddress
}
```

This message is expected to fail if:
- `ClientId` is invalid (see naming requirements)
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
	PreviousConnectionId           string
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
- `PreviousConnectionId` is not empty and invalid (see naming requirements)
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
- If a previous connection exists but does not match the supplied parameters.
- `ProofInit` does not prove that the counterparty connection is in state INIT
- `ProofClient` does not prove that the counterparty has stored the `ClientState` provided in message
- `ProofConsensus` does not prove that the counterparty has the correct consensus state for this chain

The message creates a connection for a generated connection ID with an TRYOPEN State. If a previous
connection already exists, it updates the connection state from INIT to TRYOPEN.

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
  Channel   Channel
  Signer    sdk.AccAddress
}
```

This message is expected to fail if:

- `PortId` is invalid (see naming requirements)
- `Channel` is empty
- `Signer` is empty
- A Channel End exists for the given Channel ID and Port ID

The message creates a channel on chain A with an INIT state for a generated Channel ID
and Port ID.

### MsgChannelOpenTry

A channel handshake initialization attempt is acknowledged by a chain B using
the `MsgChannelOpenTry` message.

```go
type MsgChannelOpenTry struct {
	PortId                      string    
	PreviousChannelId            string   
	Channel                     Channel 
	CounterpartyVersion         string 
	ProofInit                   []byte
	ProofHeight                 Height
	Signer                      sdk.AccAddress 
}
```

This message is expected to fail if:

- `PortId` is invalid (see naming requirements)
- `PreviousChannelId` is not empty and invalid (see naming requirements)
- `Channel` is empty
- `CounterpartyVersion` is empty
- `ProofInit` is empty
- `ProofHeight` is zero
- `Signer` is empty
- A previous channel exists and does not match the provided parameters.
- `ProofInit` does not prove that the counterparty's Channel state is in INIT

The message creates a channel on chain B with an TRYOPEN state for using a generated Channel ID 
and given Port ID if the previous channel does not already exist. Otherwise it udates the 
previous channel state from INIT to TRYOPEN.


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

The message times out a packet that was sent on chain A and never received on chain B.

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

The message times out a packet that was sent on chain A and never received on chain B.

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

The message acknowledges that the packet sent from chainA was received on chain B.
