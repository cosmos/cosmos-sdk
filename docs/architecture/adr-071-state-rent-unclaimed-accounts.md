# ADR 071: State Rent for Unclaimed Accounts

## Changelog

* Feb 5, 2024: Initial Draft (@alexanderbez)

## Status

DRAFT

## Abstract

We propose a mechanism for state rent to be applied to unclaimed accounts/balances.
This mechanism can be extended to existing accounts/balances as well with some minor
revisions.

## Context

As of SDK 0.50.x (Eden), including all previous versions, when an existing account
sends tokens to a new account, i.e. an account that does not exist on-chain yet,
the new account is created automatically via `SetAccount` as a byproduct of sending
tokens.

This might seem inconsequential, but it has a few major implications. The primary
being that users can frontrun account creation by possibly knowing the address
ahead of time and creating the account before the sender sends tokens. Another
implication is that this can drastically bloat state depending on the parameters
of the chain (e.g. min fees).

With the advent of [#19188](https://github.com/cosmos/cosmos-sdk/pull/19188), the
execution flow is changed such that the account is not created automatically. Instead,
the balance is set on the recipient's address. Then as soon as the recipient sends
their first transaction, the account is created.

However, even with this improvement in execution flow, there are still DoS and
state bloat implications. E.g. iterating over a large balance array can drastically
degrade the performance of a chain in `BeginBlock` and `EndBlock`.

Thus, we propose a mechanism for charging state rent on yet-to-be-claimed accounts/balances.
This mechanism can further be extended and adapted to charge state rent for existing
objects on-chain.

## Decision

We propose to adopt a similar mechanism to Solana's [state rent algorithm](https://docs.solanalabs.com/implemented-proposals/rent). Specifically, we propose that unclaimed account balance's be charged some
dynamic rent, set by governance and controlled by on-chain parameters every epoch.

When a user sends tokens to a non-existent account, only that account's balance
will exist until the recipient sends their first transaction. All non-existent
account balance's will be indexed using a collections map, indexed by the amount
and address.

We define some threshold `MinBalance`, such that all balances for non-existent
accounts that fall below this threshold, will be charged a dynamic rent fee every
configurable epoch, `E`. If the non-existent account's balance is no longer
sufficient to cover the rent or the balance reaches zero, the balance will be
permanently deleted from state. In other words, the recipient of the balance must
send their first transaction prior to being delinquent, otherwise the recipient
looses their funds.

The collected rent can then be sent to the community pool, burned, or used in some
combination of the two. Note, rent will be charged per unique denomination in the
unclaimed account's balance.

We define the collection index as follows, such that we only iterate over balances
that are below the threshold `MinBalance`:

```go
var (
  // ...

	NewAccountByAmountPrefix = collections.NewPrefix(0)
)

type Keeper struct {
  // ...

  NewAccKeySet collections.KeySet[collections.Triple[sdkmath.Int, string, []byte]] // <balance, denom, address>
}

func NewKeeper(...) {
  k := Keeper{
    // ...

    NewAccKeySet: collections.NewKeySet(
      schemaBuilder,
      NewAccountByAmountPrefix,
      "new_account_by_amount",
      collections.TripleKeyCodec(collections.IntKey, collections.StringKey, collections.BytesKey),
    ),
  }

  // ...
}
```

We can then define iteration as follows:

```go
k.NewAccKeySet.Walk(ctx, nil, func(key collections.Triple[sdkmath.Int, string, []byte]) (bool, error) {
  amount := key.K1()
  if amount.LT(MinBalance) {
    // Charge rent...
  }
})
```

We propose that we introduce the index and iteration logic to reside in the new
`x/accounts` module, where the configured epoch module has access to the `x/accounts`
module to trigger the state rent execution. However, the actual state rent logic
resides within `x/accounts`, or perhaps a new module altogether.

## Alternatives

There are many models for which we can design a sufficient state rent solution,
see [Further Discussions](#further-discussions) below for an example. However,
in the context of unclaimed account balances, we believe the solution laid out is
simple and effective to achieve the desired outcome of minimizing state bloat.

## Consequences

### Positive

* Provides an economic incentive model for state bloat of unclaimed account balances
  can be minimized.
* We can further expand upon the aforementioned proposal to design a more generalized
  approach of state rent going beyond just unclaimed account balances, such as existing
  accounts that have a balance below some threshold.

### Neutral

* Requires usage of an epoch or epoch-like module.
* If the list of unclaimed account balances that have a balance below the chain's
  threshold, the epoch iteration could potentially introduce some performance
  degradation.

## Further Discussions

Given that the proposed design is slightly based on Solana's state rent mechanism,
we can take the opportunity to discuss if there are inefficiencies that we could
improve upon from Solan's model.

Specifically, for Solana's case, developers mainly chose fully rent-exempt states,
and as these were tied to the native coin's price, they became prohibitively
expensive. Additionally, the need to update every rent-incurring account, a
process that involved scanning the entire state at least once per epoch, and
writing to many accounts once an epoch, which made the system inefficient. This
inefficiency, coupled with low developer engagement, led to the abandonment of
the rent concept in favor of exclusively rent-exempt account allocations.

However, some, if not all, of these points may not apply to our use case.

## References

* https://github.com/cosmos/cosmos-sdk/pull/19188
* https://docs.solanalabs.com/implemented-proposals/rent
