# IBC Core Implementation Specification

## State Structure

### ConnID

// TODO: %s ConnID ChainID
`ConnID` is the local name for a connected chain. When a conn is established between two chains, each of them subjectively name the other. Is is done by storing a mapping from the `ConnID` to the `ChainConfig`. 

* Proposal1: IP-style format

A `ConnID` is length 4 of byte array. 

NOTE: due to the state structure, the length of `ConnID`s should not be same with the `PortID`s'. 

There can be a nameservice for `ConnID`s, providing human-readable aliases.

* Proposal2: domain-style format

A `ConnID` is variadic length of alphanumeric slice, where its length is prefixed.

There can be a nameservice for `ConnID`s, where the users can bid for the names.

`ConnID`s does not have to be uniform on the network. Naming is subjective on the chains.

### PortID

A `PortID` is an independent queue space. Typically separated port will be allocated to separated modules, but a module can also have more than one port(for example, each contract in Ethermint module could have their own port). 

A `PortID` is length 2(static)/8(extended) of byte array. 

A static `PortID` is used when the number of `PortID`s is known at the compile time and does not grow over the execution. For example, the bank module will take a single static `PortID`. 0x0000 is reserved for IBC module, 0xFF__ are reserved for extended `PortID`s.

Extended `PortID`s are used when a module is expected to take unknown numbers of ports during the execution. Last 6 bytes of the id can be allocated dynamically. An extended `PortID` should start with 0xFF.

Extended `PortID`s does not have their own `PortConfig`s, but inherits the `PortConfig` of the `PortID` of its first two bytes.

`PortID`s may be uniform on the network. When a `PortID` is occupied by different modules on the multiple chains, it is possible that they cannot establish a proper connection for those modules(unless there is an alternative port reserved for the modules).

### PortType

A `PortType` denotes the functionality it provides. For example in bidirectional connection between two chains, a receipt is returned when the packet is failed. Since they are exchanged on a same port but they should use their own sequencing, we assign 0x00 for packets and 0x01 for receipts to differenciate them. 

`PortType`s should be uniform on the network. The meaning of the `PortType`s cannot be different between chains and should be specified on the protocol.

### Queue

A queue is defined by the pair of (`ConnID`, `PortID`, `PortType`). The prefix for this queue is simple concatenation of these three byte arrays. Under this, a bigendian encoded `uint64` is suffixed and the actual packets is stored there. 

### Visual

```
CCCC => ConnConfig{...}
PP => PortConfig{...}
CCCCPPT => Incoming Sequence
CCCCPPT00000000 => Packet{...}
CCCCFFPPPPPPT => Incoming Sequence
CCCCFFPPPPPPT00000000 => Packet{...}

where

C = ConnID
P = PortID
F = PortID 0xFF
0 = Sequence

each character represents single byte
```

All of the above values should be able to verify from the other chain with lightclient, and should be under a same root prefix.

## Keeper

x/ibc/keeper.go

```go
type Keeper struct {
    root store.Value
    conn func(ConnID) conn
    port func(PortID) port
}
```

`Keeper.root` stores `root`. It basically works as a magic number indicating the root point in the state for IBC, along with the additional chain constant, such as unbonding period and protocol version.

It does not have to be immutable, since it is only been read once at `conn.Listen()` on the other chain.

## Connection

x/ibc/keeper.go

Connection is defined as the following struct:

```go
type conn struct {
    config store.Value
    
    status store.Value
    commits store.Indexer
    owner store.Value
}
```

### Config

x/ibc/conn.go

`conn.Config` stores the `ConnConfig` to identify the other chain that it is receiving messages from. 

```go
type ConnConfig struct {
    ROT lite.FullCommit
    ConnID ConnID
    RootKeyPath string
}
```

`ROT` it the root-of-trust fullcommit to start the lightclient process. 

`ConnID` is the ID used by the other chain to identify this chain. It is used for constructing valid keypath for merkle proof.

`RootKeyPath` is the merkle keypath for the other chain's `Keeper.root`. Any other key-value pairs that this chain have to verify is under the root keypath on a fixed scheme.

### Status

x/ibc/conn.go

```go
const (
    ConnIdle Status = iota
    ConnSpeak
    ConnListen
    ConnSpeakSafe
    ConnListenSafe
)
```

#### ConnIdle

`ConnIdle` is the default state of conn. It means that there is no message, configuration, or lightclient headers for its `ConnID`. 

The internal logic can call `conn.speak()` to transit the state from `ConnIdle` to `ConnSpeak`.

#### ConnSpeak

`ConnSpeak` is the state where the chain can send messages to the `ConnID`. However, it does not guarantee that the other chain can receive this message, nor receiving message from this chain. It is still useful if the chain only wants to broadcast message, but if the chain wants to establish a bidirectional communication, it should call `conn.listen()` to transit the state to `ConnListen`

#### ConnListen

`ConnListen` is the state where the chain can receive message from the other chain. `conn.listen()` is called with the argument of `ChainConfig`, which contains a root-of-trust commit that can be used for header tracking.

The message can be sent back and forth, but this state does not guarantee that the other chain is also tracking this chain. An attacker could be involved in the initial config registeration, shadowing a chain from another. 

To establish bidirectional communication permissionlessly, the chain can call `conn.speakSafe()` to verify that the other chain has a valid `ChainConfig` which contains a root-of-trust commit from this chain. The chain can also call `conn.forceSafe()` to bypass the handshaking logic, if there is a trusted source who verifies that fact(e.g. single account registered `ConnConfig` on both side).

#### ConnSpeakSafe

`ConnSpeakSafe` is the state where the chain can assure that the message it is sending will be received at the other chain. It first verifies the merkle proof of the chain config on the other chain so it is following this chain properly. It then send a `PacketConfirmSafe` via the queue so the other side can acknowledge about it.

#### ConnListenSafe

`ConnListenSafe` is the state wher e the other chain approved that the other chain's state has been transit to `ConnSpeakSafe`.

`conn.listenSafe` is automatically called by the ibc handler when a `PacketConfirmSafe` is received. 

Advanced queue manipulation, such as cleanup or timeout, is recommended to be done in this state.

### Commits

x/ibc/provider.go

`conn.commits` is a simple `uint64 => lite.FullCommit` mapping. It is wrapped by `provider`, a temporal object implementing `lite.Provider`. 

When calling `conn.listen()`, its root-of-trust commit is also stored in the `conn.commits` so the `lite.DynamicVerifier` can start tracking from there.

### Owner

x/ibc/conn.go
// TODO

`conn.owner` stores a `ConnUser` who called the `conn.speak()`. When there is a further state transition, the logic automatically checks whether the caller has equal or more permission then the registered `owner`. 

## Port

x/ibc/keeper.go

`port` is defined as the following struct:

```go
type port struct {
    config store.Value
    queue func(ConnID, PortType) queue
}
```

### Config

x/ibc/port.go

```go

```

// TODO

`PortConfig` is checked by the other chain at `port.ready()`. When the ports are not compatible with each other(occupied by different logics, version not compatible, etc.), the safe connection cannot be established and the application logic should manage to resolve it. 

## Queue

x/ibc/keeper/go

`queue `is defined as the following struct:

```go
type queue struct {
    outgoing store.Queue
    incoming store.Value
    
    status store.Value
}
```

### Outgoing

Outgoing packets are stored in `queue.outgoing`. 

### Incoming

### Status
