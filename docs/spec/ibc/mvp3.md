# IBC Spec

## MVP3

`IBCOpen` is added to open the connection between two chains. Also, `IBCUpdate` is added, making it able to prove the header.

### IBC Module

```golang
type IBCOutMsg struct {
  IBCTransfer
  DestChainID string
}

type IBCInMsg struct {
  IBCTransfer
  Proof             merkle.IAVLProof
  SourceChainID     string
  SourceChainHeight uint64
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
