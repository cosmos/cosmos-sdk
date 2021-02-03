<!--
order: 1
-->

# Concepts

## FeeAllowanceGrant

`FeeAllowanceGrant` is stored in the KVStore to record a grant with full context. Every grant will contain `granter`, `grantee` and what kind of `allowance` is granted. `granter` is an account address who is giving permissoin to `grantee`(another account address) to use fees, where as `grantee` is an account address of beneficiary. and `allowance` is what kind of fee allowance(`BasicFeeAllowance` or `PeriodicFeeAllowance`) is granted to grantee. `allowance` can accepts an interface which implements `FeeAllowanceI` as `Any` type. There can be only one existing feegrant allowed for a `grantee` and `granter`, self grant not allowed.

+++ https://github.com/cosmos/cosmos-sdk/blob/master/proto/cosmos/feegrant/v1beta1/feegrant.proto#L75-L81

`FeeAllowanceI` looks like: 

+++ https://github.com/cosmos/cosmos-sdk/blob/master/x/feegrant/types/fees.go#L9-L32

## Fee Allowance types
There are two types of fee allowances present at the moment
- `BasicFeeAllowance`
- `PeriodicFeeAllowance`

## BasicFeeAllowance

`BasicFeeAllowance` is one time permission for `grantee` to use fee from a `granter`'s account. if any of the `spend_limit` or `expiration` reached the grant will be removed from the state.
 

+++ https://github.com/cosmos/cosmos-sdk/blob/master/proto/cosmos/feegrant/v1beta1/feegrant.proto#L13-L26

- `spend_limit` is a limit of coins that are allowed to use from the `granter` account. If no value mentioned assumes as no limit for coins, `grantee` can use any number of available tokens from `granter` account address before the expiration.

- `expiration` is time of when the grant can be expire. If the value left empty there is no expiry for the grant.

- whenever a grant created with the empty values of `spend_limit` and `expiration` still it is valid grant. It won't restrict the `grantee` to use any number of tokens from `granter` and no expiration. The only way to restrict the `grantee` is revoking the grant. 

## PeriodicFeeAllowance

`PeriodicFeeAllowance` is a repeating fee allowance for the mentioned period, we can mention when the grant can expire as well as when a period can reset. We can also mention how many maximum of coins can be used in a mentioned period of time.

+++ https://github.com/cosmos/cosmos-sdk/blob/master/proto/cosmos/feegrant/v1beta1/feegrant.proto#L28-L73

- `basic` is the instance of `BasicFeeAllowance` which is optional for periodic fee allowance. if empty grant will have no `expiration` and no `spend_limit`

- `period` is the specific period of time or blocks, after period crossed `period_spend_limit` will be reset. 

- `period_spend_limit` keeps track of how many coins left in the period.

- `period_can_spend` specifies max coins can be used in every period.

- `period_reset` keeps track of when a next period reset should happen.

## FeeAccount flag

`feegrant` module will introduce a `FeeAccount` flag for cli for the sake executing transactions with fee granter, when this flag set `clientCtx` will append the granter account address for transaction generated through cli.

```go 
message Fee {
  repeated cosmos.base.v1beta1.Coin amount = 1;
  uint64 gas_limit = 2;
  string payer = 3;
  string granter = 4;
}
```

Example cmd:
```go
./simd tx gov submit-proposal --title="Test Proposal" --description="My awesome proposal" --type="Text" --from validator-key --fee-account=cosmos1xh44hxt7spr67hqaa7nyx5gnutrz5fraw6grxn --chain-id=testnet --fees="10stake"
```

## DeductGrantedFeeDecorator

`feegrant` module also adds a `DeductGrantedFeeDecorator` ante handler. Whenever a transaction is being executed with `granter` field set, then this ante handler will check whether `payer` and `granter` has proper fee allowance grant in state. If it exists the fees will be deducted from the `granter`'s account address. If the `granter` field isn't set then this ante handler works as normal fee deductor.
