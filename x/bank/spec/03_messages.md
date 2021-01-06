<!--
order: 3
-->

# Messages

## MsgSend

```go
// MsgSend represents a message to send coins from one account to another.
message MsgSend {
  string   from_address                    = 1;
  string   to_address                      = 2;
  repeated cosmos.base.v1beta1.Coin amount = 3;
}
```

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
