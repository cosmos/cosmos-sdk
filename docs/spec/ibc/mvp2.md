# IBC Spec

*This is a living document and should be edited as the IBC spec and implementation change*

## MVP2

IBC module will store its own router for handling custom incoming msgs. `IBCPush` are made for inter-module communication. `IBCRegisterMsg` adds a handler in the router of the module.

### IBC Module

```golang
// User facing API

type IBCTransferData struct {
    SrcAddr  sdk.Address
    DestAddr sdk.Address
    Coins    sdk.Coins
}

// Implements ibc.PacketData
type IBCTransferPacket struct {
    IBCTransferData
}

// Implements ibc.PacketData
type IBCReceivePacket struct {
    IBCTransferData    
}

type Packet struct {
    Data      PacketData
    SrcChain  string    
    DestChain string
}

// Internal API

func NewHandler(dispatcher Dispatcher, ibcm IBCMapper) sdk.Handler

type IBCMapper struct {
    ingressKey sdk.StoreKey // Source Chain ID            => last income msg's sequence
    egressKey  sdk.StoreKey // (Dest chain ID, Msg index) => length / indexed msg
}

type IngressKey struct {
    SourceChain string
}

type EgressKey struct {
    DestChain   string
    Index       int64
}

// Used by other modules
func (ibcm IBCMapper) PushPacket(ctx sdk.Context, dest string, data PacketData)
```

`egressKey` stores the outgoing `IBCTransfer`s as a list. Its getter takes an `EgressKey` and returns the length if `egressKey.Index == -1`, an element if `egressKey.Index > 0`.

`ingressKey` stores the last income `IBCTransfer`'s sequence. Its getter takes an `IngressKey`.

## Relayer

**Packets**
- Connect to 2 Tendermint RPC endpoints
- Query for IBC outgoing `IBCOutMsg` queue (can poll on a certain time interval, or check after each new block, etc)
- For any new `IBCOutMsg`, build `IBCInMsg` and post to destination chain

## CLI

- Load relay process
- Execute `IBCOutMsg`
