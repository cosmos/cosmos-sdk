## Types

### Account

Accounts are exposed internally as an interface, and stored internally as
either a base account or vesting account. Module clients wishing to add more
account types may do so.

#### Account Interface

The account interface exposes methods to read and write standard account information.
Note that all of these methods operate on an account struct confirming to the interface
- in order to write the account to the store, the account keeper will need to be used.

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

A base account is the simplest and most common account type, which just stores all requisite
fields directly in a struct.

```golang
type BaseAccount struct {
  Address       AccAddress
  Coins         Coins
  PubKey        PubKey
  AccountNumber uint64
  Sequence      uint64
}
```
