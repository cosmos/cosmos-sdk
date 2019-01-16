## Keepers

The bank module provides three different exported keeper interfaces which can be passed to other modules which need to read or update account balances. Modules should use the least-permissive interface which provides the functionality they require.

Note that you should always review the `bank` module code to ensure that permissions are limited in the way that you expect.

### Common Types

#### Input

An input of a multiparty transfer

```golang
type Input struct {
  Address AccAddress
  Coins   Coins
}
```

#### Output

An output of a multiparty transfer.

```golang
type Output struct {
  Address AccAddress
  Coins   Coins
}
```

### BaseKeeper

The base keeper provides full-permission access: the ability to arbitrary modify any account's balance and mint or burn coins.

```golang
type BaseKeeper interface {
  SetCoins(addr AccAddress, amt Coins)
  SubtractCoins(addr AccAddress, amt Coins)
  AddCoins(addr AccAddress, amt Coins)
  InputOutputCoins(inputs []Input, outputs []Output)
}
```

`setCoins` fetches an account by address, sets the coins on the account, and saves the account.

```
setCoins(addr AccAddress, amt Coins)
  account = accountKeeper.getAccount(addr)
  if account == nil
    fail with "no account found"
  account.Coins = amt
  accountKeeper.setAccount(account)
```

`subtractCoins` fetches the coins of an account, subtracts the provided amount, and saves the account. This decreases the total supply.

```
subtractCoins(addr AccAddress, amt Coins)
  oldCoins = getCoins(addr)
  newCoins = oldCoins - amt
  if newCoins < 0
    fail with "cannot end up with negative coins"
  setCoins(addr, newCoins)
```

`addCoins` fetches the coins of an account, adds the provided amount, and saves the account. This increases the total supply.

```
addCoins(addr AccAddress, amt Coins)
  oldCoins = getCoins(addr)
  newCoins = oldCoins + amt
  setCoins(addr, newCoins)
```

`inputOutputCoins` transfers coins from any number of input accounts to any number of output accounts.

```
inputOutputCoins(inputs []Input, outputs []Output)
  for input in inputs
    subtractCoins(input.Address, input.Coins)
  for output in outputs
    addCoins(output.Address, output.Coins)
```

### SendKeeper

The send keeper provides access to account balances and the ability to transfer coins between accounts, but not to alter the total supply (mint or burn coins).

```golang
type SendKeeper interface {
  SendCoins(from AccAddress, to AccAddress, amt Coins)
}
```

`sendCoins` transfers coins from one account to another.

```
sendCoins(from AccAddress, to AccAddress, amt Coins)
  subtractCoins(from, amt)
  addCoins(to, amt)
```

### ViewKeeper

The view keeper provides read-only access to account balances but no balance alteration functionality. All balance lookups are `O(1)`.

```golang
type ViewKeeper interface {
  GetCoins(addr AccAddress) Coins
  HasCoins(addr AccAddress, amt Coins) bool
}
```

`getCoins` returns the coins associated with an account.

```
getCoins(addr AccAddress)
  account = accountKeeper.getAccount(addr)
  if account == nil
    return Coins{}
  return account.Coins
```

`hasCoins` returns whether or not an account has at least the provided amount of coins.

```
hasCoins(addr AccAddress, amt Coins)
  account = accountKeeper.getAccount(addr)
  coins = getCoins(addr)
  return coins >= amt 
```
