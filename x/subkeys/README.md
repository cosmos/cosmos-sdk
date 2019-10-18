## Fee Delegation Work

Much of this is ported from an older implementation of fee delegation from aaronc.

To take a look at this (which had many changes to auth as well, which has been heavily refactored upstream):

```
git remote add keys https://github.com/cosmos-cg-key-management/cosmos-sdk.git
git fetch keys
git diff --compact-summary f1b08b85f 980d713f
```

This is based on https://gist.github.com/aaronc/b60628017352df5983791cad30babe56#fee-delegation

In particular the following parts:

--------------------

The `delegation` module also allows for fee delegation via some
changes to the `AnteHandler` and `StdTx`. The behavior is similar
to that described above for `Msg` delegations except using
the interface `FeeAllowance` instead of `Capability`:

```go
// FeeAllowance defines a permission for one account to use another account's balance
// to pay fees
type FeeAllowance interface {
	// Accept checks whether this allowance allows the provided fees to be spent,
	// and optionally updates the allowance or deletes it entirely
	Accept(fee sdk.Coins, block abci.Header) (allow bool, updated FeeAllowance, delete bool)
}
```

An example `FeeAllowance` could be created that simply sets a `SpendLimit`:

```go
type BasicFeeAllowance struct {
	// SpendLimit specifies the maximum amount of tokens that can be spent
	// by this capability and will be updated as tokens are spent. If it is
	// empty, there is no spend limit and any amount of coins can be spent.
	SpendLimit sdk.Coins
}

func (cap BasicFeeAllowance) Accept(fee sdk.Coins, block abci.Header) (allow bool, updated FeeAllowance, delete bool) {
	left, invalid := cap.SpendLimit.SafeSub(fee)
	if invalid {
		return false, nil, false
	}
	if left.IsZero() {
		return true, nil, true
	}
	return true, BasicFeeAllowance{SpendLimit: left}, false
}

```

Other `FeeAllowance` types could be created such as a daily spend limit.

## `StdTx` and `AnteHandler` changes

In order to support delegated fees `StdTx` and the `AnteHandler` needed to be changed.

The field `FeeAccount` was added to `StdTx`.

```go
type StdTx struct {
	Msgs       []sdk.Msg      `json:"msg"`
	Fee        StdFee         `json:"fee"`
	Signatures []StdSignature `json:"signatures"`
	Memo       string         `json:"memo"`
	// FeeAccount is an optional account that fees can be spent from if such
	// delegation is enabled
	FeeAccount sdk.AccAddress `json:"fee_account"`
}
```

An interface `FeeDelegationHandler` (which is implemented by the `delegation` module) was created and a parameter for it was added to the default `AnteHandler`:

```go
type FeeDelegationHandler interface {
	// AllowDelegatedFees checks if the grantee can use the granter's account to spend the specified fees, updating
	// any fee allowance in accordance with the provided fees
	AllowDelegatedFees(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, fee sdk.Coins) bool
}

// NewAnteHandler returns an AnteHandler that checks and increments sequence
// numbers, checks signatures & account numbers, and deducts fees from the first
// signer.
func NewAnteHandler(ak AccountKeeper, fck FeeCollectionKeeper, feeDelegationHandler FeeDelegationHandler, sigGasConsumer SignatureVerificationGasConsumer) sdk.AnteHandler {
```

Basically if someone sets `FeeAccount` on `StdTx`, the `AnteHandler` will call into the `delegation` module via its `FeeDelegationHandler` and check if the tx's fees have been delegated by that `FeeAccount` to the key actually signing the transaction.

## Core `FeeAllowance` types

```go
type BasicFeeAllowance struct {
	// SpendLimit specifies the maximum amount of tokens that can be spent
	// by this capability and will be updated as tokens are spent. If it is
	// empty, there is no spend limit and any amount of coins can be spent.
	SpendLimit sdk.Coins
	// Expiration specifies an optional time when this allowance expires
	Expiration time.Time
}

type PeriodicFeeAllowance struct {
	BasicFeeAllowance
	// Period specifies the time duration in which PeriodSpendLimit coins can
	// be spent before that allowance is reset
	Period time.Duration
	// PeriodSpendLimit specifies the maximum number of coins that can be spent
	// in the Period
	PeriodSpendLimit sdk.Coins
	// PeriodCanSpend is the number of coins left to be spend before the PeriodReset time
	PeriodCanSpend sdk.Coins
	// PeriodReset is the time at which this period resets and a new one begins,
	// it is calculated from the start time of the first transaction after the
	// last period ended
	PeriodReset time.Time
}
```