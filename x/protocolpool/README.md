---
sidebar_position: 1
---

# `x/protocolpool`

## Concepts

Protopool is a module that handle functionality around community pool funds. This provides a separate module account for community pool making it easier to track the pool assets. We no longer track community pool assets in distribution module, but instead in this protocolpool module. Funds are migrated from the distribution module's community pool to protocolpool's module account.

## State Transitions

### FundCommunityPool

FundCommunityPool can be called by any valid account to send funds to the protocolpool module account.

```protobuf
  // FundCommunityPool defines a method to allow an account to directly
  // fund the community pool.
  rpc FundCommunityPool(MsgFundCommunityPool) returns (MsgFundCommunityPoolResponse);
```

### CommunityPoolSpend

CommunityPoolSpend can be called by the module authority (default governance module account) or any account with authorization to spend funds from the protocolpool module account to a receiver address.

```protobuf
  // CommunityPoolSpend defines a governance  operation for sending tokens from
  // the community pool in the x/protocolpool module to another account, which
  // could be the governance module itself. The authority is defined in the
  // keeper.
  rpc CommunityPoolSpend(MsgCommunityPoolSpend) returns (MsgCommunityPoolSpendResponse);
```

## Messages

### MsgFundCommunityPool

This message sends coins directly from the sender to the community pool.

:::tip
If you know the protocolpool module account address, you can directly use bank `send` transaction instead.
::::

```protobuf reference
https://github.com/cosmos/cosmos-sdk/blob/9dd34510e27376005e7e7ff3628eab9dbc8ad6dc/proto/cosmos/protocolpool/v1/tx.proto#L31-L41
```

* The msg will fail if the amount cannot be transferred from the sender to the protocolpool module account.

```go reference
func (k Keeper) FundCommunityPool(ctx context.Context, amount sdk.Coins, sender sdk.AccAddress) error {
	return k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, amount)
}
```

### MsgCommunityPoolSpend

This message distributes funds from the protocolpool module account to the recipient using `DistributeFromFeePool` keeper method.

```protobuf reference
https://github.com/cosmos/cosmos-sdk/blob/9dd34510e27376005e7e7ff3628eab9dbc8ad6dc/proto/cosmos/protocolpool/v1/tx.proto#L46-L59
```

The message will fail under the following conditions:

* The amount cannot be transferred to the recipient from the protocolpool module account.
* The `recipient` address is restricted

```go
func (k Keeper) DistributeFromFeePool(ctx context.Context, amount sdk.Coins, receiveAddr sdk.AccAddress) error {
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiveAddr, amount)
}
```

## Client

It takes the advantage of `AutoCLI`

```go reference
https://github.com/cosmos/cosmos-sdk/blob/9dd34510e27376005e7e7ff3628eab9dbc8ad6dc/x/protocolpool/autocli.go
```
