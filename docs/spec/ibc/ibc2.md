# IBC Spec

## MVP2

`IBCUpdate` is added, making it able to prove the header.

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
