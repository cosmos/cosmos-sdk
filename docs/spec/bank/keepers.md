## Keepers

The bank module provides three different exported keeper interfaces which can be passed to other modules which need to read or update account balances. Modules should use the least-permissive interface which provides the functionality they require.

### BaseKeeper

The base keeper provides full-permission access: the ability to arbitrary modify any account's balance and mint or burn coins.

```golang
type BaseKeeper interface {
  SetCoins(ctx Context, addr AccAddress, amt Coins) error
  SubtractCoins(ctx Context, addr AccAddress, amt Coins) (Coins, Tags, error)
  AddCoins(ctx Context, addr AccAddress, amt Coins) (Coins, Tags, error)
}
```

### SendKeeper

The send keeper provides access to account balances and the ability to transfer coins between accounts, but not to alter the total supply (mint or burn coins).

```golang
type SendKeeper interface {
  SendCoins(ctx Content, from AccAddress, to AccAddress, amt Coins) (Tags, error)
  InputOutputCoins(ctx Context, inputs []Input, outputs []Output) (Tags, error)
}
```

### ViewKeeper

The view keeper provides read-only access to account balances but no balance alteration functionality. All balance lookups are `O(1)`.

```golang
type ViewKeeper interface {
  GetCoins(ctx Context, addr AccAddress) Coins
  HasCoins(ctx Context, addr AccAddress, amt Coins) bool
}
```
