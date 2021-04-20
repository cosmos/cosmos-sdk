<!--
order: 3
-->

# Messages

## MsgSend

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0/proto/cosmos/bank/v1beta1/tx.proto#L19-L28

`handleMsgSend` just runs `inputOutputCoins`.

```go
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
