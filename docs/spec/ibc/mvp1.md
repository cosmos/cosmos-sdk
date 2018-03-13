# IBC Spec

*This is a living document and should be edited as the IBC spec and implementation change*

## MVP1

The initial implementation of IBC will include just enough for simple coin transfers between chains, with safety features such as ACK messages being added later.

### IBC Module

```golang
// User facing API

type IBCTransferPacket struct {
    DestAddr sdk.Address
    Coins    sdk.Coins
}

type IBCTransferMsg struct {
    IBCTransferPacket
    DestChain string
}

type IBCReceiveMsg struct {
    IBCTransferPacket
    SrcChain string
}

// Internal API

type IBCMapper struct {
    ingressKey sdk.StoreKey // Source Chain ID            => last income msg's sequence
    egressKey  sdk.StoreKey // (Dest chain ID, Msg index) => length / indexed msg
}

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
