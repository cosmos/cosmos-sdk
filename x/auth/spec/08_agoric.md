<!--
order: 8
-->

# Agoric

## Intro

Agoric has added a new type of vesting account, made some modifications
to periodic vesting accounts, and has some comments to clarify the
behavior of other vesting accounts.

- Periodic vesting accounts now allow additional grants to be added
to the account after they have been created. The new vesting schedule
is merged with the existing schedule.
- The new account type is clawback vesting, which is like periodic
vesting, but unvested coins may be "clawed back" by the account
which funded the initial grant of coins.  These accounts have
independent schedules for unlocking (being available for transfer)
and vesting (also unavailable for transfer, but also subject to
clawback). Additional grants can be made to an existing account.
Unvested coins may be staked, but staking rewards are subject to
vesting (see details below). Staked (or unbonding) tokens are clawed
back in their staked (unbonding) state.

## Vesting Account Types

### ClawbackVestingAccount

[Snippet available after upstreaming.]

Note that the `vesting_periods` field defines what is locked and subject to
clawback. The `lockup_periods` field defines locking that is not subject to
clawback with the same total amount but a separate schedule. Thus, tokens
might be vested (immune from clawback) but still locked (unavailable for
transfer).

## Vesting Account Specification

### Clawback Vesting Accounts

Works like a Periodic vesting account, except that coins must be both vested
and unlocked in order to be transferred. This allows coins to be vested, but
still not available for transfer. For instance, you can have an account where
the tokens vest monthly over two years, but are locked until 12 months. In
this case, no coins can be transferred until the one year anniversary where
half become transferrable, then one twelfth of the remainder each month
thereafter.

Since the commands to stake and unstake tokens do not specify the character
of the funds to use (i.e. locked, vested, etc.), vesting accounts use a policy
to determine how bonded and unbonding tokens are distributed. To determine
the amount that is available for transfer (the only question most vesting
accounts face), the policy is to maximize the number available for transfer
by maximizing the locked tokens used for delegation. Slashing looks like
tokens which remain forever bonded, and thus reduce the number of actual
bonded and unbonded tokens which are encumbered to prevent transfer. This
is the policy followed by all vesting accounts.

But for clawback accounts, we distinguish between the encumbrance that is
enforced preventing transfer and the right of the funder to retrieve the
unvested amount from the account. The latter is not reduced by slashing,
though slashing might limit the number of tokens which can be retrieved.

Additional grants may be added to an existing `ClawbackVestingAccount` with
their own schedule. Additional grants must come from the same account that
provided the initial grant that created the account.

Staking rewards are automatically added as such an additional grant following
the current vesting schedule, with amounts scaled proportionally. (Staking
rewards are given an immediate unlocking schedule.) The proportion follows
the policy used to determine which tokens may be transferred - staked tokens
prefer to be unvested first.

### Transferring/Sending

We've modified the `x/bank` mechanisms for vesting integration:

- The `LockedCoins()` method takes a `sdk.Context` instead of a `Time`.

## Glossary

- Clawback: removal of unvested tokens from a ClawbackVestingAccount.
- ClawbackVestingAccount: a vesting account specifying separate schedules for
vesting (subject to clawback) and lockup (inability to transfer out of the
account - the encumbrance implemented by the other vesting account types).
