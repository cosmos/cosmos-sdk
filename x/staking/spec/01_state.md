<!--
order: 1
-->

# State

## LastTotalPower

LastTotalPower tracks the total amounts of bonded tokens recorded during the previous end block.
Store entries prefixed with "Last" must remain unchanged until EndBlock.

- LastTotalPower: `0x12 -> ProtocolBuffer(sdk.Int)`

## Params

Params is a module-wide configuration structure that stores system parameters
and defines overall functioning of the staking module.

- Params: `Paramsspace("staking") -> legacy_amino(params)`

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.1/proto/cosmos/staking/v1beta1/staking.proto#L230-L241

## Validator

Validators can have one of three statuses

- `Unbonded`: The validator is not in the active set. They cannot sign blocks and do not earn
  rewards. They can receive delegations.
- `Bonded`": Once the validator receives sufficient bonded tokens they automtically join the
  active set during [`EndBlock`](./05_end_block.md#validator-set-changes) and their status is updated to `Bonded`.
  They are signing blocks and receiving rewards. They can receive further delegations.
  They can be slashed for misbehavior. Delegators to this validator who unbond their delegation
  must wait the duration of the UnbondingTime, a chain-specific param, during which time
  they are still slashable for offences of the source validator if those offences were committed
  during the period of time that the tokens were bonded.
- `Unbonding`: When a validator leaves the active set, either by choice or due to slashing, jailing or
  tombstoning, an unbonding of all their delegations begins. All delegations must then wait the UnbondingTime
  before their tokens are moved to their accounts from the `BondedPool`.

Validators objects should be primarily stored and accessed by the
`OperatorAddr`, an SDK validator address for the operator of the validator. Two
additional indices are maintained per validator object in order to fulfill
required lookups for slashing and validator-set updates. A third special index
(`LastValidatorPower`) is also maintained which however remains constant
throughout each block, unlike the first two indices which mirror the validator
records within a block.

- Validators: `0x21 | OperatorAddr -> ProtocolBuffer(validator)`
- ValidatorsByConsAddr: `0x22 | ConsAddr -> OperatorAddr`
- ValidatorsByPower: `0x23 | BigEndian(ConsensusPower) | OperatorAddr -> OperatorAddr`
- LastValidatorsPower: `0x11 OperatorAddr -> ProtocolBuffer(ConsensusPower)`

`Validators` is the primary index - it ensures that each operator can have only one
associated validator, where the public key of that validator can change in the
future. Delegators can refer to the immutable operator of the validator, without
concern for the changing public key.

`ValidatorByConsAddr` is an additional index that enables lookups for slashing.
When Tendermint reports evidence, it provides the validator address, so this
map is needed to find the operator. Note that the `ConsAddr` corresponds to the
address which can be derived from the validator's `ConsPubKey`.

`ValidatorsByPower` is an additional index that provides a sorted list of
potential validators to quickly determine the current active set. Here
ConsensusPower is validator.Tokens/10^6 by default. Note that all validators
where `Jailed` is true are not stored within this index.

`LastValidatorsPower` is a special index that provides a historical list of the
last-block's bonded validators. This index remains constant during a block but
is updated during the validator set update process which takes place in [`EndBlock`](./05_end_block.md).

Each validator's state is stored in a `Validator` struct:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0/proto/cosmos/staking/v1beta1/staking.proto#L65-L99

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0/proto/cosmos/staking/v1beta1/staking.proto#L24-L63

## Delegation

Delegations are identified by combining `DelegatorAddr` (the address of the delegator)
with the `ValidatorAddr` Delegators are indexed in the store as follows:

- Delegation: `0x31 | DelegatorAddr | ValidatorAddr -> ProtocolBuffer(delegation)`

Stake holders may delegate coins to validators; under this circumstance their
funds are held in a `Delegation` data structure. It is owned by one
delegator, and is associated with the shares for one validator. The sender of
the transaction is the owner of the bond.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0/proto/cosmos/staking/v1beta1/staking.proto#L159-L170

### Delegator Shares

When one Delegates tokens to a Validator they are issued a number of delegator shares based on a
dynamic exchange rate, calculated as follows from the total number of tokens delegated to the
validator and the number of shares issued so far:

`Shares per Token = validator.TotalShares() / validator.Tokens()`

Only the number of shares received is stored on the DelegationEntry. When a delegator then
Undelegates, the token amount they receive is calculated from the number of shares they currently
hold and the inverse exchange rate:

`Tokens per Share = validator.Tokens() / validatorShares()`

These `Shares` are simply an accounting mechanism. They are not a fungible asset. The reason for
this mechanism is to simplify the accounting around slashing. Rather than iteratively slashing the
tokens of every delegation entry, instead the Validators total bonded tokens can be slashed,
effectively reducing the value of each issued delegator share.

## UnbondingDelegation

Shares in a `Delegation` can be unbonded, but they must for some time exist as
an `UnbondingDelegation`, where shares can be reduced if Byzantine behavior is
detected.

`UnbondingDelegation` are indexed in the store as:

- UnbondingDelegation: `0x32 | DelegatorAddr | ValidatorAddr -> ProtocolBuffer(unbondingDelegation)`
- UnbondingDelegationsFromValidator: `0x33 | ValidatorAddr | DelegatorAddr -> nil`

The first map here is used in queries, to lookup all unbonding delegations for
a given delegator, while the second map is used in slashing, to lookup all
unbonding delegations associated with a given validator that need to be
slashed.

A UnbondingDelegation object is created every time an unbonding is initiated.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0/proto/cosmos/staking/v1beta1/staking.proto#L172-L198

## Redelegation

The bonded tokens worth of a `Delegation` may be instantly redelegated from a
source validator to a different validator (destination validator). However when
this occurs they must be tracked in a `Redelegation` object, whereby their
shares can be slashed if their tokens have contributed to a Byzantine fault
committed by the source validator.

`Redelegation` are indexed in the store as:

- Redelegations: `0x34 | DelegatorAddr | ValidatorSrcAddr | ValidatorDstAddr -> ProtocolBuffer(redelegation)`
- RedelegationsBySrc: `0x35 | ValidatorSrcAddr | ValidatorDstAddr | DelegatorAddr -> nil`
- RedelegationsByDst: `0x36 | ValidatorDstAddr | ValidatorSrcAddr | DelegatorAddr -> nil`

The first map here is used for queries, to lookup all redelegations for a given
delegator. The second map is used for slashing based on the `ValidatorSrcAddr`,
while the third map is for slashing based on the `ValidatorDstAddr`.

A redelegation object is created every time a redelegation occurs. To prevent
"redelegation hopping" redelegations may not occur under the situation that:

- the (re)delegator already has another immature redelegation in progress
  with a destination to a validator (let's call it `Validator X`)
- and, the (re)delegator is attempting to create a _new_ redelegation
  where the source validator for this new redelegation is `Validator X`.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0/proto/cosmos/staking/v1beta1/staking.proto#L200-L228

## Queues

All queues objects are sorted by timestamp. The time used within any queue is
first rounded to the nearest nanosecond then sorted. The sortable time format
used is a slight modification of the RFC3339Nano and uses the the format string
`"2006-01-02T15:04:05.000000000"`. Notably this format:

- right pads all zeros
- drops the time zone info (uses UTC)

In all cases, the stored timestamp represents the maturation time of the queue
element.

### UnbondingDelegationQueue

For the purpose of tracking progress of unbonding delegations the unbonding
delegations queue is kept.

- UnbondingDelegation: `0x41 | format(time) -> []DVPair`

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0/proto/cosmos/staking/v1beta1/staking.proto#L123-L133

### RedelegationQueue

For the purpose of tracking progress of redelegations the redelegation queue is
kept.

- RedelegationQueue: `0x42 | format(time) -> []DVVTriplet`

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0/proto/cosmos/staking/v1beta1/staking.proto#L140-L152

### ValidatorQueue

For the purpose of tracking progress of unbonding validators the validator
queue is kept.

- ValidatorQueueTime: `0x43 | format(time) -> []sdk.ValAddress`

The stored object as each key is an array of validator operator addresses from
which the validator object can be accessed. Typically it is expected that only
a single validator record will be associated with a given timestamp however it is possible
that multiple validators exist in the queue at the same location.

## HistoricalInfo

HistoricalInfo objects are stored and pruned at each block such that the staking keeper persists
the `n` most recent historical info defined by staking module parameter: `HistoricalEntries`.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0/proto/cosmos/staking/v1beta1/staking.proto#L15-L22

At each BeginBlock, the staking keeper will persist the current Header and the Validators that committed
the current block in a `HistoricalInfo` object. The Validators are sorted on their address to ensure that
they are in a determisnistic order.
The oldest HistoricalEntries will be pruned to ensure that there only exist the parameter-defined number of
historical entries.
