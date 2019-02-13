# State

## Pool

The pool tracks the total amounts of tokens (each staking denom is tracked
separately) and their state (bonded or loose). 

Note: `NotBondedTokens` _includes_ both tokens in an `unbonding` state as well
as fully `unbonded` state. 

 - Pool: `0x01 -> amino(pool)`

```golang
type Pool struct {
    NotBondedTokens sdk.Int   // tokens not associated with any bonded validator
    BondedTokens    sdk.Int   // reserve of bonded tokens
}
```

## LastTotalPower

LastTotalPower tracks the total amounts of bonded tokens recorded during the previous 
end block. 

 - LastTotalPower: `0x12 -> amino(sdk.Int)`

## Params

Params is a module-wide configuration structure that stores system parameters
and defines overall functioning of the staking module.

 - Params: `Paramsspace("staking") -> amino(params)`

```golang
type Params struct {
    UnbondingTime time.Duration // time duration of unbonding
    MaxValidators uint16        // maximum number of validators
    MaxEntries    uint16        // max entries for either unbonding delegation or redelegation (per pair/trio)
    BondDenom     string        // bondable coin denomination
}
```

## Validator

Validators objects should be primarily stored and accessed by the
`OperatorAddr`, an SDK validator address for the operator of the validator. Two
additional indices are maintained per validator object in order to fulfill
required lookups for slashing and validator-set updates. A third special index
(`LastValidatorPower`) is also maintained which however remains constant
throughout each block, unlike the first two indices which mirror the validator
records within a block. 

- Validators: `0x21 | OperatorAddr -> amino(validator)`
- ValidatorsByConsAddr: `0x22 | ConsAddr -> OperatorAddr`
- ValidatorsByPower: `0x23 | BigEndian(Tokens) | OperatorAddr -> OperatorAddr`
- LastValidatorsPower: `0x11 OperatorAddr -> amino(Tokens) 

`Validators` is the primary index - it ensures that each operator can have only one
associated validator, where the public key of that validator can change in the
future. Delegators can refer to the immutable operator of the validator, without
concern for the changing public key.

`ValidatorByConsAddr` is an additional index that enables lookups for slashing.
When Tendermint reports evidence, it provides the validator address, so this
map is needed to find the operator. Note that the `ConsAddr` corresponds to the
address which can be derived from the validator's `ConsPubKey`. 

`ValidatorsByPower` is an additional index that provides a sorted list o
potential validators to quickly determine the current active set. Note 
that all validators where `Jailed` is true are not stored within this index.

`LastValidatorsPower` is a special index that provides a historical list of the
last-block's bonded validators. This index remains constant during a block but
is updated during the validator set update process which takes place in [end
block](end_block.md). 

Each validator's state is stored in a `Validator` struct:

```golang
type Validator struct {
    OperatorAddr    sdk.ValAddress // address of the validator's operator; bech encoded in JSON
    ConsPubKey      crypto.PubKey  // Tendermint consensus pubkey of validator
    Jailed          bool           // has the validator been jailed?

    Status          sdk.BondStatus // validator status (bonded/unbonding/unbonded)
    Tokens          sdk.Int        // delegated tokens (incl. self-delegation)
    DelegatorShares sdk.Dec        // total shares issued to a validator's delegators

    Description Description  // description terms for the validator

    // Needed for ordering vals in the by-power key
    UnbondingHeight  int64     // if unbonding, height at which this validator has begun unbonding
    UnbondingMinTime time.Time // if unbonding, min time for the validator to complete unbonding

    Commission Commission // info about the validator's commission
}

type Commission struct {
    Rate          sdk.Dec   // the commission rate charged to delegators
    MaxRate       sdk.Dec   // maximum commission rate which this validator can ever charge
    MaxChangeRate sdk.Dec   // maximum daily increase of the validator commission
    UpdateTime    time.Time // the last time the commission rate was changed
}

type Description struct {
    Moniker  string // name
    Identity string // optional identity signature (ex. UPort or Keybase)
    Website  string // optional website link
    Details  string // optional details
}
```

## Delegation

Delegations are identified by combining `DelegatorAddr` (the address of the delegator)
with the `ValidatorAddr` Delegators are indexed in the store as follows:

- Delegation: ` 0x31 | DelegatorAddr | ValidatorAddr -> amino(delegation)`

Atom holders may delegate coins to validators; under this circumstance their
funds are held in a `Delegation` data structure. It is owned by one
delegator, and is associated with the shares for one validator. The sender of
the transaction is the owner of the bond.

```golang
type Delegation struct {
    DelegatorAddr sdk.AccAddress 
    ValidatorAddr sdk.ValAddress 
    Shares        sdk.Dec        // delegation shares received
}
```

## UnbondingDelegation

Shares in a `Delegation` can be unbonded, but they must for some time exist as
an `UnbondingDelegation`, where shares can be reduced if Byzantine behavior is
detected.

`UnbondingDelegation` are indexed in the store as:

- UnbondingDelegation: ` 0x32 | DelegatorAddr | ValidatorAddr ->
   amino(unbondingDelegation)`
- UnbondingDelegationsFromValidator: ` 0x33 | ValidatorAddr | DelegatorAddr ->
   nil`

The first map here is used in queries, to lookup all unbonding delegations for
a given delegator, while the second map is used in slashing, to lookup all
unbonding delegations associated with a given validator that need to be
slashed.

A UnbondingDelegation object is created every time an unbonding is initiated.

```golang
type UnbondingDelegation struct {
    DelegatorAddr sdk.AccAddress             // delegator
    ValidatorAddr sdk.ValAddress             // validator unbonding from operator addr
    Entries       []UnbondingDelegationEntry // unbonding delegation entries
}

type UnbondingDelegationEntry struct {
    CreationHeight int64     // height which the unbonding took place
    CompletionTime time.Time // unix time for unbonding completion
    InitialBalance sdk.Coin  // atoms initially scheduled to receive at completion
    Balance        sdk.Coin  // atoms to receive at completion
}
```

## Redelegation

The bonded tokens worth of a `Delegation` may be instantly redelegated from a
source validator to a different validator (destination validator). However when
this occurs they must be tracked in a `Redelegation` object, whereby their
shares can be slashed if their tokens have contributed to a Byzantine fault
committed by the source validator. 

`Redelegation` are indexed in the store as:

 - Redelegations: `0x34 | DelegatorAddr | ValidatorSrcAddr | ValidatorDstAddr -> amino(redelegation)`
 - RedelegationsBySrc: `0x35 | ValidatorSrcAddr | ValidatorDstAddr | DelegatorAddr -> nil`
 - RedelegationsByDst: `0x36 | ValidatorDstAddr | ValidatorSrcAddr | DelegatorAddr -> nil`

The first map here is used for queries, to lookup all redelegations for a given
delegator. The second map is used for slashing based on the `ValidatorSrcAddr`,
while the third map is for slashing based on the `ValidatorDstAddr`.

A redelegation object is created every time a redelegation occurs.  To prevent
"redelegation hopping" redelegations may not occure under the situation that:
 - the (re)delegator already has another unmature redelegation in progress
   with a destination to a validator (let's call it `Validator X`)
 - and, the (re)delegator is attempting to create a _new_ redelegation 
   where the source validator for this new redelegation is `Validator-X`. 

```golang
type Redelegation struct {
    DelegatorAddr    sdk.AccAddress      // delegator
    ValidatorSrcAddr sdk.ValAddress      // validator redelegation source operator addr
    ValidatorDstAddr sdk.ValAddress      // validator redelegation destination operator addr
    Entries          []RedelegationEntry // redelegation entries
}

type RedelegationEntry struct {
    CreationHeight int64     // height which the redelegation took place
    CompletionTime time.Time // unix time for redelegation completion
    InitialBalance sdk.Coin  // initial balance when redelegation started
    Balance        sdk.Coin  // current balance (current value held in destination validator)
    SharesDst      sdk.Dec   // amount of destination-validator shares created by redelegation
}
```

## Queues

All queues objects are sorted by timestamp. The time used within any queue is
first rounded to the nearest nanosecond then sorted. The sortable time format
used is a slight modification of the RFC3339Nano and uses the the format string
`"2006-01-02T15:04:05.000000000"`. Noteably This format: 

 - right pads all zeros
 - drops the time zone info (uses UTC) 

In all cases, the stored timestamp represents the maturation time of the queue
element. 

### UnbondingDelegationQueue

For the purpose of tracking progress of unbonding delegations the unbonding
delegations queue is kept. 

- UnbondingDelegation: `0x41 | format(time) -> []DVPair`

```
type DVPair struct {
	DelegatorAddr sdk.AccAddress
	ValidatorAddr sdk.ValAddress
}
```

### RedelegationQueue

For the purpose of tracking progress of redelegations the redelegation queue is
kept. 

- UnbondingDelegation: `0x42 | format(time) -> []DVVTriplet`

```
type DVVTriplet struct {
	DelegatorAddr    sdk.AccAddress
	ValidatorSrcAddr sdk.ValAddress
	ValidatorDstAddr sdk.ValAddress
}
```

### ValidatorQueue

For the purpose of tracking progress of unbonding validators the validator
queue is kept. 

- ValidatorQueueTime: `0x43 | format(time) -> []sdk.ValAddress`

The stored object as each key is an array of validator operator addresses from
which the validator object can be accessed.  Typically it is expected that only
a single validator record will be associated with a given timestamp however it is possible
that multiple validators exist in the queue at the same location.
