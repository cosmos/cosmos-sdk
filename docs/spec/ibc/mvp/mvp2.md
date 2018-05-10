# IBC Spec

*This is a living document and should be edited as the IBC spec and implementation change*

## MVP2

IBC module will store its own router for handling custom incoming msgs. `IBCPush` are made for inter-module communication. `IBCRegisterMsg` adds a handler in the router of the module.

### IBC Module

```golang
// User facing API

type Packet struct {
    Data      Payload
    SrcChain  string
    DestChain string
}

type Payload interface {
    Type() string
    ValidateBasic() sdk.Error
}

type TransferPayload struct {
    SrcAddr  sdk.Address
    DestAddr sdk.Address
    Coins    sdk.Coins
}

// Implements sdk.Msg
type IBCTransferMsg struct {
    Packet
}

// Implements sdk.Msg
type IBCReceiveMsg struct {
    Packet
    Relayer  sdk.Address
    Sequence int64
}

// Internal API

type rule struct {
    r string
    f func(sdk.Context, IBCPacket) sdk.Result
}

type Dispatcher struct {
    rules []rule
}

func NewHandler(dispatcher Dispatcher, ibcm IBCMapper) sdk.Handler

type IBCMapper struct {
    ibcKey sdk.StoreKey // IngressKey / EgressKey             => Value
                        // Ingress: Source Chain ID           => last income msg's sequence
                        // Egress: (Dest chain ID, Msg index) => length / indexed msg
}

type IngressKey struct {
    SrcChain string
}

type EgressKey struct {
    DestChain   string
    Index       int64
}

// Used by other modules
func (ibcm IBCMapper) PushPacket(ctx sdk.Context, dest string, payload Payload)
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
