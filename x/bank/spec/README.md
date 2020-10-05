# `x/bank`

## Table of Contents

<!-- TOC -->

- **[01. Abstract](#01-abstract)**
- **[02. Concepts](#02-concepts)**
  - [Supply](#supply)
  - [Module Accounts](#module-accounts)
- **[03. State](#03-state)**
- **[04. Keepers](#04-keepers)**
  - [Common Types](#common-types)
  - [BaseKeeper](#basekeeper)
  - [SendKeeper](#sendkeeper)
  - [ViewKeeper](#viewkeeper)
- **[05. Messages](#05-messages)**
  - [MsgSend](#msgsend)
- **[06. Events](#06-events)**
  - [Handlers](#handlers)
- **[07. Parameters](#07-parameters)**

## 01. Abstract

This document specifies the bank module of the Cosmos SDK.

The bank module is responsible for handling multi-asset coin transfers between
accounts and tracking special-case pseudo-transfers which must work differently
with particular kinds of accounts (notably delegating/undelegating for vesting
accounts). It exposes several interfaces with varying capabilities for secure
interaction with other modules which must alter user balances.

In addition, the bank module tracks and provides query support for the total
supply of all assets used in the application.

This module will be used in the Cosmos Hub.

## 02. Concepts

### Supply

The `supply` module:

- passively tracks the total supply of coins within a chain,
- provides a pattern for modules to hold/interact with `Coins`, and
- introduces the invariant check to verify a chain's total supply.

#### Total Supply

The total `Supply` of the network is equal to the sum of all coins from the
account. The total supply is updated every time a `Coin` is minted (eg: as part
of the inflation mechanism) or burned (eg: due to slashing or if a governance
proposal is vetoed).

### Module Accounts

The supply module introduces a new type of `auth.AccountI` interface, called `ModuleAccountI`, which can be used by
modules to allocate tokens and in special cases mint or burn tokens. At a base
level these module accounts are capable of sending/receiving tokens to and from
`auth.AccountI` interfaces and other module accounts. This design replaces previous
alternative designs where, to hold tokens, modules would burn the incoming
tokens from the sender account, and then track those tokens internally. Later,
in order to send tokens, the module would need to effectively mint tokens
within a destination account. The new design removes duplicate logic between
modules to perform this accounting.

The `ModuleAccountI` interface is defined as follows:

```go
// ModuleAccountI defines an account interface for modules that hold tokens in
// an escrow.
type ModuleAccountI interface {
	AccountI                    // same methods as the Account interface

	GetName() string            // name of the module; used to obtain the address
	GetPermissions() []string   // permissions of module account
	HasPermission(string) bool
}
```

> **WARNING!**
> Any module or message handler that allows either direct or indirect sending of funds must explicitly guarantee those funds cannot be sent to module accounts (unless allowed).

The supply `Keeper` also introduces new wrapper functions for the auth `Keeper`
and the bank `Keeper` that are related to `ModuleAccountI`s in order to be able
to:

- Get and set `ModuleAccountI`s by providing the `Name`.
- Send coins from and to other `ModuleAccountI`s or standard `Account`s
  (`BaseAccount` or `VestingAccount`) by passing only the `Name`.
- `Mint` or `Burn` coins for a `ModuleAccountI` (restricted to its permissions).

#### Permissions

Each `ModuleAccountI` has a different set of permissions that provide different
object capabilities to perform certain actions. Permissions need to be
registered upon the creation of the supply `Keeper` so that every time a
`ModuleAccountI` calls the allowed functions, the `Keeper` can lookup the
permissions to that specific account and perform or not the action.

The available permissions are:

- `Minter`: allows for a module to mint a specific amount of coins.
- `Burner`: allows for a module to burn a specific amount of coins.
- `Staking`: allows for a module to delegate and undelegate a specific amount of coins.

## 03. State

The `x/bank` module keeps state of two primary objects, account balances and the
total supply of all balances.

- Balances: `[]byte("balances") | []byte(address) / []byte(balance.Denom) -> ProtocolBuffer(balance)`
- Supply: `0x0 -> ProtocolBuffer(Supply)`

## 04. Keepers

The bank module provides three different exported keeper interfaces which can be passed to other modules which need to read or update account balances. Modules should use the least-permissive interface which provides the functionality they require.

Note that you should always review the `bank` module code to ensure that permissions are limited in the way that you expect.

### Common Types

#### Input

An input of a multiparty transfer

```protobuf
// Input models transaction input.
message Input {
  string   address                        = 1;
  repeated cosmos.base.v1beta1.Coin coins = 2;
}
```

#### Output

An output of a multiparty transfer.

```protobuf
// Output models transaction outputs.
message Output {
  string   address                        = 1;
  repeated cosmos.base.v1beta1.Coin coins = 2;
}
```

### BaseKeeper

The base keeper provides full-permission access: the ability to arbitrary modify any account's balance and mint or burn coins. The `BaseKeeper` struct implements the following `Keeper` interface.

```go
// Keeper defines a module interface that facilitates the transfer of coins
// between accounts.
type Keeper interface {
	SendKeeper

	InitGenesis(sdk.Context, types.GenesisState)
	ExportGenesis(sdk.Context) *types.GenesisState

	GetSupply(ctx sdk.Context) exported.SupplyI
	SetSupply(ctx sdk.Context, supply exported.SupplyI)

	GetDenomMetaData(ctx sdk.Context, denom string) types.Metadata
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
	MarshalSupply(supplyI exported.SupplyI) ([]byte, error)
	UnmarshalSupply(bz []byte) (exported.SupplyI, error)

	types.QueryServer
}
```

### SendKeeper

The send keeper provides access to account balances and the ability to transfer coins between accounts, but not to alter the total supply (mint or burn coins).

```go
// SendKeeper defines a module interface that facilitates the transfer of coins
// between accounts without the possibility of creating coins.
type SendKeeper interface {
	ViewKeeper

	InputOutputCoins(ctx sdk.Context, inputs []types.Input, outputs []types.Output) error
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error

	SubtractCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) error
	AddCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) error

	SetBalance(ctx sdk.Context, addr sdk.AccAddress, balance sdk.Coin) error
	SetBalances(ctx sdk.Context, addr sdk.AccAddress, balances sdk.Coins) error

	GetParams(ctx sdk.Context) types.Params
	SetParams(ctx sdk.Context, params types.Params)

	SendEnabledCoin(ctx sdk.Context, coin sdk.Coin) bool
	SendEnabledCoins(ctx sdk.Context, coins ...sdk.Coin) error

	BlockedAddr(addr sdk.AccAddress) bool
}
```

### ViewKeeper

The view keeper provides read-only access to account balances but no balance alteration functionality. All balance lookups are `O(1)`.

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

## 05. Messages

### MsgSend

```protobuf
// MsgSend represents a message to send coins from one account to another.
message MsgSend {
  string   from_address                    = 1;
  string   to_address                      = 2;
  repeated cosmos.base.v1beta1.Coin amount = 3;
}
```

`handleMsgSend` just runs `inputOutputCoins`.

```
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

## 06. Events

The bank module emits the following events:

### Handlers

#### MsgSend

| Type     | Attribute Key | Attribute Value    |
| -------- | ------------- | ------------------ |
| transfer | recipient     | {recipientAddress} |
| transfer | amount        | {amount}           |
| message  | module        | bank               |
| message  | action        | send               |
| message  | sender        | {senderAddress}    |

#### MsgMultiSend

| Type     | Attribute Key | Attribute Value    |
| -------- | ------------- | ------------------ |
| transfer | recipient     | {recipientAddress} |
| transfer | amount        | {amount}           |
| message  | module        | bank               |
| message  | action        | multisend          |
| message  | sender        | {senderAddress}    |

## 07. Parameters

The bank module contains the following parameters:

| Key                | Type          | Example                            |
| ------------------ | ------------- | ---------------------------------- |
| SendEnabled        | []SendEnabled | [{denom: "stake", enabled: true }] |
| DefaultSendEnabled | bool          | true                               |

The corresponding Protobuf message is:

```protobuf
// Params defines the parameters for the bank module.
message Params {
  option                                    = false;
  repeated SendEnabled send_enabled         = 1;
  bool                 default_send_enabled = 2;
}
```

### SendEnabled

The send enabled parameter is an array of SendEnabled entries mapping coin
denominations to their send_enabled status. Entries in this list take
precedence over the `DefaultSendEnabled` setting.

### DefaultSendEnabled

The default send enabled value controls send transfer capability for all
coin denominations unless specifically included in the array of `SendEnabled`
parameters.
