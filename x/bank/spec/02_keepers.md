<!--
order: 2
-->

# Keepers

The bank module provides these exported keeper interfaces that can be
passed to other modules that read or update account balances. Modules
should use the least-permissive interface that provides the functionality they
require.

Best practices dictate careful review of `bank` module code to ensure that
permissions are limited in the way that you expect.

## Blocklisting Addresses

The `x/bank` module accepts a map of addresses that are considered blocklisted
from directly and explicitly receiving funds through means such as `MsgSend` and
`MsgMultiSend` and direct API calls like `SendCoinsFromModuleToAccount`.

Typically, these addresses are module accounts. If these addresses receive funds
outside the expected rules of the state machine, invariants are likely to be
broken and could result in a halted network.

By providing the `x/bank` module with a blocklisted set of addresses, an error occurs for the operation if a user or client attempts to directly or indirectly send funds to a blocklisted account, for example, by using [IBC](http://docs.cosmos.network/master/ibc/).

## Common Types

### Input

An input of a multiparty transfer

```protobuf
// Input models transaction input.
message Input {
  string   address                        = 1;
  repeated cosmos.base.v1beta1.Coin coins = 2;
}
```

### Output

An output of a multiparty transfer.

```protobuf
// Output models transaction outputs.
message Output {
  string   address                        = 1;
  repeated cosmos.base.v1beta1.Coin coins = 2;
}
```

## BaseKeeper

The base keeper provides full-permission access: the ability to arbitrary modify any account's balance and mint or burn coins.

Restricted permission to mint per module could be achieved by using baseKeeper with `WithMintCoinsRestriction` to give specific restrictions to mint (e.g. only minting certain denom).

```go
// Keeper defines a module interface that facilitates the transfer of coins
// between accounts.
type Keeper interface {
    SendKeeper

    InitGenesis(sdk.Context, *types.GenesisState)
    ExportGenesis(sdk.Context) *types.GenesisState

    GetSupply(ctx sdk.Context, denom string) sdk.Coin
    GetPaginatedTotalSupply(ctx sdk.Context, pagination *query.PageRequest) (sdk.Coins, *query.PageResponse, error)
    IterateTotalSupply(ctx sdk.Context, cb func(sdk.Coin) bool)
    GetDenomMetaData(ctx sdk.Context, denom string) (types.Metadata, bool)
    SetDenomMetaData(ctx sdk.Context, denomMetaData types.Metadata)
    IterateAllDenomMetaData(ctx sdk.Context, cb func(types.Metadata) bool)

    SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
    SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error
    SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
    DelegateCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
    UndelegateCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
    MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
    BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error

    DelegateCoins(ctx sdk.Context, delegatorAddr, moduleAccAddr sdk.AccAddress, amt sdk.Coins) error
    UndelegateCoins(ctx sdk.Context, moduleAccAddr, delegatorAddr sdk.AccAddress, amt sdk.Coins) error

    types.QueryServer
}
```

## SendKeeper

The send keeper provides access to account balances and the ability to transfer coins between
accounts. The send keeper does not alter the total supply (mint or burn coins).

```go
// SendKeeper defines a module interface that facilitates the transfer of coins
// between accounts without the possibility of creating coins.
type SendKeeper interface {
    ViewKeeper

    InputOutputCoins(ctx sdk.Context, inputs []types.Input, outputs []types.Output) error
    SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error

    GetParams(ctx sdk.Context) types.Params
    SetParams(ctx sdk.Context, params types.Params)

    IsSendEnabledCoin(ctx sdk.Context, coin sdk.Coin) bool
    IsSendEnabledCoins(ctx sdk.Context, coins ...sdk.Coin) error

    BlockedAddr(addr sdk.AccAddress) bool
}
```

## ViewKeeper

The view keeper provides read-only access to account balances. The view keeper does not have balance alteration functionality. All balance lookups are `O(1)`.

```go
// ViewKeeper defines a module interface that facilitates read only access to
// account balances.
type ViewKeeper interface {
    ValidateBalance(ctx sdk.Context, addr sdk.AccAddress) error
    HasBalance(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coin) bool

    GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
    GetAccountsBalances(ctx sdk.Context) []types.Balance
    GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
    LockedCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
    SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins

    IterateAccountBalances(ctx sdk.Context, addr sdk.AccAddress, cb func(coin sdk.Coin) (stop bool))
    IterateAllBalances(ctx sdk.Context, cb func(address sdk.AccAddress, coin sdk.Coin) (stop bool))
}
```
