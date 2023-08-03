<!--
order: 3
-->

# Messages

In this section we describe the processing of the staking messages and the corresponding updates to the state. All created/modified state objects specified by each message are defined within the [state](./02_state_transitions.md) section.

## MsgCreateValidator

A validator is created using the `MsgCreateValidator` message.
The validator must be created with an initial delegation from the operator.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0/proto/cosmos/staking/v1beta1/tx.proto#L16-L17

+++ https://github.com/cosmos/cosmos-sdk/blob/80b365e60ce2757284c8e91168ad4e28a9171768/proto/cosmos/staking/v1beta1/tx.proto#L71-L89

This message is expected to fail if:

- another validator with this operator address is already registered
- another validator with this pubkey is already registered
- the commission parameters are faulty, namely:
    - `MaxRate` is either > 1 or < 0
    - the initial `Rate` is either negative or > `MaxRate`
    - the initial `MaxChangeRate` is either negative or > `MaxRate`
- the description fields are too large

This message creates and stores the `Validator` object at appropriate indexes.
The validator always starts as unbonded but may be bonded
in the first end-block.

## MsgEditValidator

The `Description`, `CommissionRate` of a validator can be updated using the
`MsgEditValidator` message.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0/proto/cosmos/staking/v1beta1/tx.proto#L19-L20

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0/proto/cosmos/staking/v1beta1/tx.proto#L56-L76

This message is expected to fail if:

- the initial `CommissionRate` is either negative or > `MaxRate`
- the `CommissionRate` has already been updated within the previous 24 hours
- the `CommissionRate` is > `MaxChangeRate`
- the description fields are too large

This message stores the updated `Validator` object.

## MsgDelegate

Within this message the delegator provides coins, and in return receives
some amount of their validator's (newly created) delegator-shares that are
assigned to `Delegation.Shares`.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0/proto/cosmos/staking/v1beta1/tx.proto#L22-L24

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0/proto/cosmos/staking/v1beta1/tx.proto#L81-L90

This message is expected to fail if:

- the validator does not exist
- the `Amount` `Coin` has a denomination different than one defined by `params.BondDenom`
- the exchange rate is invalid, meaning the validator has no tokens (due to slashing) but there are outstanding shares
- the amount delegated is less than the minimum allowed delegation
- the delegator is a liquid staking provider and
    - the delegation exceeds the global liquid staking cap
    - the delegation exceeds the validator liquid staking cap
    - the delegation exceeds the validator bond cap

When this message is processed the following actions occur:

If an existing `Delegation` object for provided addresses does not already
exist then it is created as part of this message otherwise the existing
`Delegation` is updated to include the newly received shares.

The delegator receives newly minted shares at the current exchange rate.
The exchange rate is the number of existing shares in the validator divided by
the number of currently delegated tokens.

The validator is updated in the `ValidatorByPower` index, and the delegation is
tracked in validator object in the `Validators` index.

If the delegator is a liquid staking provider, increment `TotalLiquidStakedTokens`
 and validator's `liquid_shares`.

It is possible to delegate to a jailed validator, the only difference being it
will not be added to the power index until it is unjailed.

![Delegation sequence](../../../docs/uml/svg/delegation_sequence.svg)

## MsgUndelegate

The `MsgUndelegate` message allows delegators to undelegate their tokens from
validator.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0/proto/cosmos/staking/v1beta1/tx.proto#L30-L32

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0/proto/cosmos/staking/v1beta1/tx.proto#L112-L121

This message returns a response containing the completion time of the undelegation:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0/proto/cosmos/staking/v1beta1/tx.proto#L123-L126

This message is expected to fail if:

- the delegation doesn't exist
- the validator doesn't exist
- the delegation has less shares than the ones worth of `Amount`
- existing `UnbondingDelegation` has maximum entries as defined by `params.MaxEntries`
- the `Amount` has a denomination different than one defined by `params.BondDenom`
- the unbonded delegation is a ValidatorBond and the reduction in validator bond would cause the existing liquid delegations to exceed the validator's bond cap

When this message is processed the following actions occur:

- validator's `DelegatorShares` and the delegation's `Shares` are both reduced by the message `SharesAmount`
- calculate the token worth of the shares remove that amount tokens held within the validator
- with those removed tokens, if the validator is:
    - `Bonded` - add them to an entry in `UnbondingDelegation` (create `UnbondingDelegation` if it doesn't exist) with a completion time a full unbonding period from the current time. Update pool shares to reduce BondedTokens and increase NotBondedTokens by token worth of the shares.
    - `Unbonding` - add them to an entry in `UnbondingDelegation` (create `UnbondingDelegation` if it doesn't exist) with the same completion time as the validator (`UnbondingMinTime`).
    - `Unbonded` - then send the coins the message `DelegatorAddr`
- if there are no more `Shares` in the delegation, then the delegation object is removed from the store
- if the delegator is a liquid staking provider, decrement `TotalLiquidStakedTokens` and validator's `liquid_shares`

![Unbond sequence](../../../docs/uml/svg/unbond_sequence.svg)

## MsgBeginRedelegate

The redelegation command allows delegators to instantly switch validators. Once
the unbonding period has passed, the redelegation is automatically completed in
the EndBlocker.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0/proto/cosmos/staking/v1beta1/tx.proto#L26-L28

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0/proto/cosmos/staking/v1beta1/tx.proto#L95-L105

This message returns a response containing the completion time of the redelegation:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0/proto/cosmos/staking/v1beta1/tx.proto#L107-L110

This message is expected to fail if:

- the delegation doesn't exist
- the source or destination validators don't exist
- the delegation has less shares than the ones worth of `Amount`
- the source validator has a receiving redelegation which is not matured (aka. the redelegation may be transitive)
- existing `Redelegation` has maximum entries as defined by `params.MaxEntries`
- the `Amount` `Coin` has a denomination different than one defined by `params.BondDenom`
- if the delegation is a `ValidatorBond` and the reduction in validator bond would cause the existing liquid delegation to exceed the cap
- if delegator is a liquid staking provider and the delegation exceeds the global liquid staking cap, the validator liquid staking cap or the validator bond cap

When this message is processed the following actions occur:

- the source validator's `DelegatorShares` and the delegations `Shares` are both reduced by the message `SharesAmount`
- calculate the token worth of the shares remove that amount tokens held within the source validator.
- if the source validator is:
    - `Bonded` - add an entry to the `Redelegation` (create `Redelegation` if it doesn't exist) with a completion time a full unbonding period from the current time. Update pool shares to reduce BondedTokens and increase NotBondedTokens by token worth of the shares (this may be effectively reversed in the next step however).
    - `Unbonding` - add an entry to the `Redelegation` (create `Redelegation` if it doesn't exist) with the same completion time as the validator (`UnbondingMinTime`).
    - `Unbonded` - no action required in this step
- Delegate the token worth to the destination validator, possibly moving tokens back to the bonded state.
- if there are no more `Shares` in the source delegation, then the source delegation object is removed from the store
- if delegator is a liquid staking provider, decrement src validator's `liquid_shares` and increment dest validator's `liquid_shares`

![Begin redelegation sequence](../../../docs/uml/svg/begin_redelegation_sequence.svg)

## MsgCancelUnbondingDelegation

The `MsgCancelUnbondingDelegation` message allows delegators to cancel the `unbondingDelegation` entry and deleagate back to a previous validator.

+++ hhttps://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/proto/cosmos/staking/v1beta1/tx.proto#L36-L40

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/proto/cosmos/staking/v1beta1/tx.proto#L146-L165

This message is expected to fail if:

* the `unbondingDelegation` entry is already processed.
* the `cancel unbonding delegation` amount is greater than the `unbondingDelegation` entry balance.
* the `cancel unbonding delegation` height doesn't exists in the `unbondingDelegationQueue` of the delegator.

When this message is processed the following actions occur:

* if the `unbondingDelegation` Entry balance is zero 
    * in this condition `unbondingDelegation` entry will be removed from `unbondingDelegationQueue`.
    * otherwise `unbondingDelegationQueue` will be updated with new `unbondingDelegation` entry balance and initial balance
* the validator's `DelegatorShares` and the delegation's `Shares` are both increased by the message `Amount`.

### `MsgTokenizeShares`

The `MsgTokenizeShares` message is used to create tokenize delegated tokens. At execution, the specified amount of delegation disappear from the account and share tokens are provided. Share tokens are denominated in the validator and record id of the underlying delegation.

A user may tokenize some or all of their delegation.

They will receive shares with the denom of `cosmosvaloper1xxxx/5` where 5 is the record id for the validator operator.

MsgTokenizeSharesResponse provides the number of tokens generated and their denom.

This message is expected to fail if:
- the delegation is a `ValidatorBond` (`ValidatorBond` cannot be tokenized).
- the the sender is NOT a liquid staking provider and tokenized shares would exceed the global liquid staking cap, the validator liquid staking cap, or the validator bond cap
- the account has `TokenizeSharesLock` enabled.
- the account is a `VestingAccount`. Users will have to move vested tokens to a new account and endure the unbonding period. We view this as an acceptable tradeoff vs. the complex book keeping required to track vested tokens.
- the delegator has zero delegation.

When this message is processed the following actions occur:
- At execution, the specified amount of delegation disappear from the account and share tokens are provided.