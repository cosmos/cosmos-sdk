# IBC Spec

*This is a living document and should be edited as the IBC spec and 
implementation change*

## MVP1

The initial implementation of IBC will include just enough for simple coin 
transfers between chains, with safety features such as ACK messages being added 
later.

It is a complete stand-alone module. It includes the commands to send IBC
packets as well as to post them to the destination chain.

### IBC Module

```go
// User facing API

type IBCPacket struct {
    SrcAddr   sdk.Address
    DestAddr  sdk.Address
    Coins     sdk.Coins
    SrcChain  string
    DestChain string
}

// Implements sdk.Msg
type IBCTransferMsg struct {
    IBCPacket
}

// Implements sdk.Msg
type IBCReceiveMsg struct {
    IBCPacket
    Relayer  sdk.Address
    Sequence int64
}

// Internal API

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

```

`egressKey` stores the outgoing `IBCTransfer`s as a list. Its getter takes an 
`EgressKey` and returns the length if `egressKey.Index == -1`, an element if 
`egressKey.Index > 0`.

`ingressKey` stores the latest income `IBCTransfer`'s sequence. It's getter 
takes an `IngressKey`.

## Relayer

**Packets**
- Connect to 2 Tendermint RPC endpoints
- Query for IBC outgoing `IBCOutMsg` queue (can poll on a certain time interval, or check after each new block, etc)
- For any new `IBCOutMsg`, build `IBCInMsg` and post to destination chain

## CLI

- Load relay process
- Execute `IBCOutMsg`
