---
sidebar_position: 1
---

# `x/protocolpool`

## Concepts

`x/protocolpool` is a supplemental Cosmos SDK module that handles functionality for community pool funds. The module provides a separate module account for the community pool making it easier to track the pool assets. Starting with v0.53 of the Cosmos SDK, community funds can be tracked using this module instead of the `x/distribution` module. Funds are migrated from the `x/distribution` module's community pool to `x/protocolpool`'s module account automatically.

This module is `supplemental`; it is not required to run a Cosmos SDK chain. `x/protocolpool` enhances the community pool functionality provided by `x/distribution` and enables custom modules to further extend the community pool.

Note: _as long as an external commmunity pool keeper (here, `x/protocolpool`) is wired in DI configs, `x/distribution` will automatically use it for its external pool._

## Usage Limitations

The following `x/distribution` handlers will now return an error when the `protocolpool` module is used with `x/distribution`:

**QueryService**

- `CommunityPool`

**MsgService**

- `CommunityPoolSpend`
- `FundCommunityPool`

If you have services that rely on this functionality from `x/distribution`, please update them to use the `x/protocolpool` equivalents.

## State Transitions

### FundCommunityPool

FundCommunityPool can be called by any valid account to send funds to the `x/protocolpool` module account.

```protobuf
  // FundCommunityPool defines a method to allow an account to directly
  // fund the community pool.
  rpc FundCommunityPool(MsgFundCommunityPool) returns (MsgFundCommunityPoolResponse);
```

### CommunityPoolSpend

CommunityPoolSpend can be called by the module authority (default governance module account) or any account with authorization to spend funds from the `x/protocolpool` module account to a receiver address.

```protobuf
  // CommunityPoolSpend defines a governance  operation for sending tokens from
  // the community pool in the x/protocolpool module to another account, which
  // could be the governance module itself. The authority is defined in the
  // keeper.
  rpc CommunityPoolSpend(MsgCommunityPoolSpend) returns (MsgCommunityPoolSpendResponse);
```

### CreateContinuousFund

CreateContinuousFund is a message used to initiate a continuous fund for a specific recipient. The proposed percentage of funds will be distributed only on withdraw request for the recipient. The fund distribution continues until expiry time is reached or continuous fund request is canceled.
NOTE:  This feature is designed to work with the SDK's default bond denom. 

```protobuf
  // CreateContinuousFund defines a method to distribute a percentage of funds to an address continuously.
  // This ContinuousFund can be indefinite or run until a given expiry time.
  // Funds come from validator block rewards from x/distribution, but may also come from
  // any user who funds the ProtocolPoolEscrow module account directly through x/bank.
  rpc CreateContinuousFund(MsgCreateContinuousFund) returns (MsgCreateContinuousFundResponse);
```

### CancelContinuousFund

CancelContinuousFund is a message used to cancel an existing continuous fund proposal for a specific recipient. Cancelling a continuous fund stops further distribution of funds, and the state object is removed from storage.

```protobuf
  // CancelContinuousFund defines a method for cancelling continuous fund.
  rpc CancelContinuousFund(MsgCancelContinuousFund) returns (MsgCancelContinuousFundResponse);
```

## Messages

### MsgFundCommunityPool

This message sends coins directly from the sender to the community pool.

:::tip
If you know the `x/protocolpool` module account address, you can directly use bank `send` transaction instead.
:::

```protobuf reference
https://github.com/cosmos/cosmos-sdk/blob/release/v0.53.x/proto/cosmos/protocolpool/v1/tx.proto#L43-L53
```

* The msg will fail if the amount cannot be transferred from the sender to the `x/protocolpool` module account.

```go
func (k Keeper) FundCommunityPool(ctx context.Context, amount sdk.Coins, sender sdk.AccAddress) error {
	return k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, amount)
}
```

### MsgCommunityPoolSpend

This message distributes funds from the `x/protocolpool` module account to the recipient using `DistributeFromCommunityPool` keeper method.

```protobuf reference
https://github.com/cosmos/cosmos-sdk/blob/release/v0.53.x/proto/cosmos/protocolpool/v1/tx.proto#L58-L69
```

The message will fail under the following conditions:

* The amount cannot be transferred to the recipient from the `x/protocolpool` module account.
* The `recipient` address is restricted

```go
func (k Keeper) DistributeFromCommunityPool(ctx context.Context, amount sdk.Coins, receiveAddr sdk.AccAddress) error {
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiveAddr, amount)
}
```

### MsgCreateContinuousFund

This message is used to create a continuous fund for a specific recipient. The proposed percentage of funds will be distributed only on withdraw request for the recipient. This fund distribution continues until expiry time is reached or continuous fund request is canceled.

```protobuf reference
https://github.com/cosmos/cosmos-sdk/blob/release/v0.53.x/proto/cosmos/protocolpool/v1/tx.proto#L114-L130
```

The message will fail under the following conditions:

- The recipient address is empty or restricted.
- The percentage is zero/negative/greater than one.
- The Expiry time is less than the current block time.

:::warning
If two continuous fund proposals to the same address are created, the previous ContinuousFund will be updated with the new ContinuousFund.
:::

```go reference
https://github.com/cosmos/cosmos-sdk/blob/release/v0.53.x/x/protocolpool/keeper/msg_server.go#L103-L166
```

### MsgCancelContinuousFund

This message is used to cancel an existing continuous fund proposal for a specific recipient. Once canceled, the continuous fund will no longer distribute funds at each begin block, and the state object will be removed. 

```protobuf reference
https://github.com/cosmos/cosmos-sdk/blob/release/v0.53.x/x/protocolpool/proto/cosmos/protocolpool/v1/tx.proto#L136-L161
```

The message will fail under the following conditions:

- The recipient address is empty or restricted.
- The ContinuousFund for the recipient does not exist.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/release/v0.53.x/x/protocolpool/keeper/msg_server.go#L188-L226
```

## Client

It takes the advantage of `AutoCLI`

```go reference
https://github.com/cosmos/cosmos-sdk/blob/release/v0.53.x/x/protocolpool/autocli.go
```
