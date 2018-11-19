## Transactions

### MsgSend

#### Input

```golang
type Input struct {
  Address AccAddress
  Coins   Coins
}
```

#### Output

```golang
type Output struct {
  Address AccAddress
  Coins   Coins
}
```

```golang
type MsgSend struct {
  Inputs  []Input
  Outputs []Output
}
```

### MsgIssue

```golang
type MsgIssue struct {
  Banker  AccAddress
  Outputs []Output
}
```
