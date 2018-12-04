## State

### Accounts

Accounts contain authentication information for a uniquely identified external user of an SDK blockchain,
including public key, address, and account number / sequence number for replay protection. For efficiency,
since account balances must also be fetched to pay fees, account structs also store the balance of a user
as `sdk.Coins`.

Accounts are exposed internally as an interface, and stored internally as
either a base account or vesting account. Module clients wishing to add more
account types may do so.

#### Interface

- `0x01 | Address -> amino(account)`

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

#### BaseAccount

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

#### VestingAccount

See [Vesting](vesting.md).
