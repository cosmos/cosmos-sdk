# ADR 029: Fee Grant Module

## Changelog

- 2020/08/18: Initial Draft

## Status

Accepted

## Context

In order to make blockchain transactions, the signing account must possess a sufficient balance of the right denomination
in order to pay fees. There are classes of transactions where needing to maintain a wallet with sufficient fees is a
barrier to adoption.

For instance, when proper permissions are setup, someone may temporarily delegate the ability to vote on proposals to
a "burner" account that is stored on a mobile phone with only minimal security.

Other use cases include workers tracking items in a supply chain or farmers submitting field data for analytics
or compliance purposes.

For all of these use cases, UX would be significantly enhanced by obviating the need for these accounts to always
maintain the appropriate fee balance. This is especially true if we wanted to achieve enterprise adoption for something
like supply chain tracking.

While one solution would be to have a service that fills up these accounts automatically with the appropriate fees, a better UX
would be provided by allowing these accounts to pull from a common fee pool account with proper spending limits.
A single pool would reduce the churn of making lots of small "fill up" transactions and also more effectively leverages
the resources of the organization setting up the pool.

## Decision

As a solution we propose a module, `x/feegrant` which allows one account, the "granter" to grant another account, the "grantee"
an allowance to spend the granter's account balance for fees within certain well-defined limits.

Fee allowances are defined by the extensible `FeeAllowanceI` interface:

```go
type FeeAllowanceI {
    // Accept can use fee payment requested as well as timestamp/height of the current block
 	// to determine whether or not to process this. This is checked in
 	// Keeper.UseGrantedFees and the return values should match how it is handled there.
 	//
 	// If it returns an error, the fee payment is rejected, otherwise it is accepted.
 	// The FeeAllowance implementation is expected to update it's internal state
 	// and will be saved again after an acceptance.
 	//
 	// If remove is true (regardless of the error), the FeeAllowance will be deleted from storage
 	// (eg. when it is used up). (See call to RevokeFeeAllowance in Keeper.UseGrantedFees)
 	Accept(fee sdk.Coins, blockTime time.Time, blockHeight int64) (remove bool, err error)
}
```

Two basic fee allowance types, `BasicFeeAllowance` and `PeriodicFeeAllowance` are defined to support known use cases:

```proto
// BasicFeeAllowance implements FeeAllowance with a one-time grant of tokens
// that optionally expires. The delegatee can use up to SpendLimit to cover fees.
message BasicFeeAllowance{
     repeated cosmos_sdk.v1.Coin spend_limit = 1;
     ExpiresAt expiration = 2;
}

// PeriodicFeeAllowance extends FeeAllowance to allow for both a maximum cap,
// as well as a limit per time period.
message PeriodicFeeAllowance{
     BasicFeeAllowance basic = 1;
     Duration period = 2;
     repeated cosmos_sdk.v1.Coin period_spend_limit = 3;
     repeated cosmos_sdk.v1.Coin period_can_spend = 4;

     ExpiresAt period_reset = 5;
}

// Duration is a repeating unit of either clock time or number of blocks.
// This is designed to be added to an ExpiresAt struct.
message Duration{
    google.protobuf.Timestamp clock = 1;
    int64 block = 2;
}

// ExpiresAt is a point in time where something expires.
// It may be *either* block time or block height
message ExpiresAt{
     google.protobuf.Timestamp time = 1;
     int64 height = 2;
 }
```

Allowances can be granted and revoked using `MsgGrantFeeAllowance` and `MsgRevokeFeeAllowance`:

```proto
message MsgGrantFeeAllowance{
     bytes granter = 1;
     bytes grantee = 2;
     google.protobuf.Any allowance = 3;
 }

 // MsgRevokeFeeAllowance removes any existing FeeAllowance from Granter to Grantee.
 message MsgRevokeFeeAllowance{
     bytes granter = 1;
     bytes grantee = 2;
 }
```

In order to use allowances in transactions, we add a new field `fee_payer` to the transaction `Fee` type:
```proto
package cosmos.tx.v1beta1;

message Fee {
  repeated cosmos.base.v1beta1.Coin amount = 1;
  uint64 gas_limit = 2;
  bytes fee_payer = 3;
}
```

`fee_payer` must either be left empty, must equal the first signer (the normal case), or must correspond to an
account which has granted a fee allowance to the first signer.

A new `AnteDecorator` named `DeductGrantedFeeDecorator` will be created in order to process transactions with `fee_payer`
set and correctly deduct fees based on fee allowances.

## Consequences

### Positive

- improved UX for use cases where it is cumbersome to maintain an account balance just for fees

### Negative

### Neutral

- a new field must be added to the transaction `Fee` message and a new `AnteDecorator` must be
created to use it

## References

- Blog article describing initial work: https://medium.com/regen-network/hacking-the-cosmos-cosmwasm-and-key-management-a08b9f561d1b
- Initial public specification: https://gist.github.com/aaronc/b60628017352df5983791cad30babe56
- Original subkeys proposal from B-harvest which influenced this design: https://github.com/cosmos/cosmos-sdk/issues/4480
