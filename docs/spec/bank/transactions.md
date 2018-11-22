## Transactions

### MsgSend

```golang
type MsgSend struct {
  Inputs  []Input
  Outputs []Output
}
```

`handleMsgSend` just runs `inputOutputCoins`.

```
handleMsgSend(msg MsgSend)
  return inputOutputCoins(msg.Inputs, msg.Outputs)
```
