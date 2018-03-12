# IBC Spec

*This is a living document and should be edited as the IBC spec and implementation change*

## MVP1

The initial implementation of IBC will include just enough for simple coin transfers between chains, with safety features such as ACK messages being added later.

### IBC Module

```golang
type IBCOutMsg struct {
  IBCTransfer
}

type IBCInMsg struct {
  IBCTransfer
}

type IBCTransfer struct {
  Destination sdk.Address
  Coins       sdk.Coins
}
```

## Relayer

**Packets**
- Connect to 2 Tendermint RPC endpoints
- Query for IBC outgoing `IBCOutMsg` queue (can poll on a certain time interval, or check after each new block, etc)
- For any new `IBCOutMsg`, build `IBCInMsg` and post to destination chain

## CLI

- Load relay process
- Execute `IBCOutMsg`
