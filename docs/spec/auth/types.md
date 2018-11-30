## Types

### Account

#### Account Interface

```golang
type Account interface {
  GetAddress() AccAddress
  SetAddress(AccAddress)

  GetPubKey() PubKey
  SetPubKey(PubKey)

  GetAccountNumber() uint64
  SetAccountNumber(uint64)

  GetSequence() uint64
  SetSequence(uint64)

  GetCoins() Coins
  SetCoins(Coins)
}
```

#### Base Account

```golang
type BaseAccount struct {
  Address       AccAddress
  Coins         Coins
  PubKey        PubKey
  AccountNumber uint64
  Sequence      uint64
}
```
