<!--
order: 1
-->

# Concepts

## Grant

`Grant` is stored in the KVStore to record a grant with full context. Every grant will contain `granter`, `grantee` and what kind of `allowance` is granted. `granter` is an account address who is giving permission to `grantee` (the beneficiary account address) to pay for some or all of `grantee`'s transaction fees. `allowance` defines what kind of fee allowance (`BasicAllowance` or `PeriodicAllowance`, see below) is granted to `grantee`. `allowance` accepts an interface which implements `FeeAllowanceI`, encoded as `Any` type. There can be only one existing fee grant allowed for a `grantee` and `granter`, self grants are not allowed.

+++ https://github.com/cosmos/cosmos-sdk/blob/691032b8be0f7539ec99f8882caecefc51f33d1f/proto/cosmos/feegrant/v1beta1/feegrant.proto#L75-L81

`FeeAllowanceI` looks like:

+++ https://github.com/cosmos/cosmos-sdk/blob/691032b8be0f7539ec99f8882caecefc51f33d1f/x/feegrant/fees.go#L9-L32

## Fee Allowance types

There are two types of fee allowances present at the moment:

- `BasicAllowance`
- `PeriodicAllowance`

## BasicAllowance

`BasicAllowance` is permission for `grantee` to use fee from a `granter`'s account. If any of the `spend_limit` or `expiration` reaches its limit, the grant will be removed from the state.

+++ https://github.com/cosmos/cosmos-sdk/blob/691032b8be0f7539ec99f8882caecefc51f33d1f/proto/cosmos/feegrant/v1beta1/feegrant.proto#L13-L26

- `spend_limit` is the limit of coins that are allowed to be used from the `granter` account. If it is empty, it assumes there's no spend limit, `grantee` can use any number of available tokens from `granter` account address before the expiration.

- `expiration` specifies an optional time when this allowance expires. If the value is left empty, there is no expiry for the grant.

- When a grant is created with empty values for `spend_limit` and `expiration`, it is still a valid grant. It won't restrict the `grantee` to use any number of tokens from `granter` and it won't have any expiration. The only way to restrict the `grantee` is by revoking the grant.

## PeriodicAllowance

`PeriodicAllowance` is a repeating fee allowance for the mentioned period, we can mention when the grant can expire as well as when a period can reset. We can also define the maximum number of coins that can be used in a mentioned period of time.

+++ https://github.com/cosmos/cosmos-sdk/blob/691032b8be0f7539ec99f8882caecefc51f33d1f/proto/cosmos/feegrant/v1beta1/feegrant.proto#L28-L73

- `basic` is the instance of `BasicAllowance` which is optional for periodic fee allowance. If empty, the grant will have no `expiration` and no `spend_limit`.

- `period` is the specific period of time, after each period passes, `period_spend_limit` will be reset.

- `period_spend_limit` specifies the maximum number of coins that can be spent in the period.

- `period_can_spend` is the number of coins left to be spent before the period_reset time.

- `period_reset` keeps track of when a next period reset should happen.

## FeeGranter flag

`feegrant` module introduces a `FeeGranter` flag for CLI for the sake of executing transactions with fee granter. When this flag is set, `clientCtx` will append the granter account address for transactions generated through CLI.

+++ https://github.com/cosmos/cosmos-sdk/blob/d97e7907f176777ed8a464006d360bb3e1a223e4/client/cmd.go#L224-L235

+++ https://github.com/cosmos/cosmos-sdk/blob/d97e7907f176777ed8a464006d360bb3e1a223e4/client/tx/tx.go#L120

+++ https://github.com/cosmos/cosmos-sdk/blob/d97e7907f176777ed8a464006d360bb3e1a223e4/x/auth/tx/builder.go#L268-L277

+++ https://github.com/cosmos/cosmos-sdk/blob/d97e7907f176777ed8a464006d360bb3e1a223e4/proto/cosmos/tx/v1beta1/tx.proto#L160-L181

Example cmd:

```go
./simd tx gov submit-proposal --title="Test Proposal" --description="My awesome proposal" --type="Text" --from validator-key --fee-granter=cosmos1xh44hxt7spr67hqaa7nyx5gnutrz5fraw6grxn --chain-id=testnet --fees="10stake"
```

## Granted Fee Deductions

Fees are deducted from grants in the `x/auth` ante handler. To learn more about how ante handlers work, read the [Auth Module AnteHandlers Guide](../../auth/spec/03_antehandlers.md).

## Gas

In order to prevent DoS attacks, using a filtered `x/feegrant` incurs gas. The SDK must assure that the `grantee`'s transactions all conform to the filter set by the `granter`. The SDK does this by iterating over the allowed messages in the filter and charging 10 gas per filtered message. The SDK will then iterate over the messages being sent by the `grantee` to ensure the messages adhere to the filter, also charging 10 gas per message. The SDK will stop iterating and fail the transaction if it finds a message that does not conform to the filter.

**WARNING**: The gas is charged against the granted allowance. Ensure your messages conform to the filter, if any, before sending transactions using your allowance.
