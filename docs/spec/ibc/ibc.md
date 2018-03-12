# IBC Specification

*This is a living document and should be edited as the IBC spec and 
implementation change*

## Engineering Philosophy

The goal is to get the simplest implementation working end-to-end first. Once
the simplest and most insecure and most feature-less use-case is implemented
we can start adding features. Let's get to the end-to-end process first though.


## MVP One

The initial implementation of IBC will include just enough for simple coin 
transfers between chains, with safety features such as ACK messages being added 
later.

### IBC Module

* handles message processing

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

### Relayer

**Packets**
* Connect to 2 Tendermint RPC endpoints
* Query for IBC outgoing `IBCOutMsg` queue (can poll on a certain time 
  interval, or check after each new block, etc)
* For any new `IBCOutMsg`, build `IBCInMsg` and post to destination chain

### CLI

* Load relay process
* Execute `IBCOutMsg`


## MVP2

* `IBCUpdate` is added, making it able to prove the header.

### IBC Module

```golang
type IBCOutMsg struct {
  IBCTransfer
}

type IBCInMsg struct {
  IBCTransfer
  Proof           merkle.IAVLProof
  FromChainID     string
  FromChainHeight uint64
}

// update sync state of other blockchain
type IBCUpdateMsg struct {
  Header tm.Header
  Commit tm.Commit
}

type IBCTransfer struct {
  Destination sdk.Address
  Coins       sdk.Coins
}
```
