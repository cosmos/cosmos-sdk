# IBC Spec

*This is a living document and should be edited as the IBC spec and implementation change*

## MVP2

IBC module will store its own router for handling custom incoming msgs. `IBCPush` are made for inter-module communication. `IBCRegisterMsg` adds a handler in the router of the module.

### IBC Module

```golang
// -------------------
// x/ibc/types

type Packet struct {
    Payload   Payload
    SrcChain  string
    DestChain string
}

type Payload interface {
    Type() string
    ValidateBasic() sdk.Error
}

type Handler func(sdk.Context, Payload) sdk.Result

type Keeper interface {
    Sender(...Payload) Sender
    RegisterHandler(string, Handler)
    Receive(sdk.Context, Packet, int64) sdk.Result
}

type Sender interface {
    Push(sdk.Context, Payload, string)
}

// ------------------
// x/ibc

// Implements sdk.Msg
type ReceiveMsg struct {
    Packet
    Relayer  sdk.Address
    Sequence int64
}

func NewHandler(keeper types.Keeper) sdk.Handler

// ------------------
// x/bank

// Implements ibc.Payload
type SendPayload struct {
    SrcAddr  sdk.Address
    DestAddr sdk.Address
    Coins    sdk.Coins
}

// Implements sdk.Msg
type IBCSendMsg struct {
    DestChain string
    TransferPayload
}

// Internal API

type IngressKey struct {
    SrcChain string
}

type EgressKey struct {
    DestChain   string
    Index       int64
}
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
