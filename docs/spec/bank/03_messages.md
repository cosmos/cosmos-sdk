# Messages

## MsgSend

```golang
type MsgSend struct {
  Inputs  []Input
  Outputs []Output
}
```

`handleMsgSend` just runs `inputOutputCoins`.

```
handleMsgSend(msg MsgSend)
  inputSum = 0
  for input in inputs
    inputSum += input.Amount
  outputSum = 0
  for output in outputs
    outputSum += output.Amount
  if inputSum != outputSum:
    fail with "input/output amount mismatch"

  return inputOutputCoins(msg.Inputs, msg.Outputs)
```
