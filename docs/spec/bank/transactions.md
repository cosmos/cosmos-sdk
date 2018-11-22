## Transactions

### Common

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

### MsgSend

```golang
type MsgSend struct {
  Inputs  []Input
  Outputs []Output
}
```
