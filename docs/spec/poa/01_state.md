# State

## LastTotalPower

LastTotalPower tracks the total amount of weight recorded during the previous end block.

- LastTotalPower: `0x12 -> amino(sdk.Int)`

## Params

Params is a module-wide configuration structure that stores system parameters
and defines overall functioning of the POA module.

- Params: `Paramsspace("poa") -> amino(params)`

```go
type Params struct {
    UnbondingTime       time.Duration // time duration of unbonding
    MaxValidators       uint16        // maximum number of validators
   // TODO: MaxEntries          uint16        // max entries for either unbonding delegation or redelegation (per pair/trio)
    AcceptAllValidators bool          // Sets the value if a network wants to accept all applicants to be validators
    IncreaseWeight      bool          // Disallow validators to increase there power
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
- ValidatorsByPower: `0x23 | BigEndian(ConsensusPower) | OperatorAddr -> OperatorAddr`
- LastValidatorsPower: `0x11 OperatorAddr -> amino(ConsensusPower)`

`Validators` is the primary index - it ensures that each operator can have only one
associated validator, where the public key of that validator can change in the
future. Delegators can refer to the immutable operator of the validator, without
concern for the changing public key.

`ValidatorByConsAddr` is an additional index that enables lookups for slashing.
When Tendermint reports evidence, it provides the validator address, so this
map is needed to find the operator. Note that the `ConsAddr` corresponds to the
address which can be derived from the validator's `ConsPubKey`.

`ValidatorsByPower` is an additional index that provides a sorted list o
potential validators to quickly determine the current active set. Here
ConsensusPower is validator.Tokens/10^6. Note that all validators where
`Jailed` is true are not stored within this index.

`LastValidatorsPower` is a special index that provides a historical list of the
last-block's bonded validators. This index remains constant during a block but
is updated during the validator set update process which takes place in [`EndBlock`](./04_end_block.md).

Each validator's state is stored in a `Validator` struct:

```go
type Validator struct {
    OperatorAddress         sdk.ValAddress  // address of the validator's operator; bech encoded in JSON
    ConsPubKey              crypto.PubKey   // the consensus public key of the validator; bech encoded in JSON
    Jailed                  bool            // has the validator been jailed from bonded status?
    Status                  sdk.BondStatus  // validator status (bonded/unbonding/unbonded)
    Weight                  sdk.Int         // weight (repuatation) associated with each validator
    Description             Description     // description terms for the validator
    UnbondingHeight         int64           // if unbonding, height at which this validator has begun unbonding
    UnbondingCompletionTime time.Time       // if unbonding, min time for the validator to complete unbonding
}

type Description struct {
    Moniker          string // name
    Identity         string // optional identity signature (ex. UPort or Keybase)
    Website          string // optional website link
    SecurityContact  string // optional email for security contact
    Details          string // optional details
}
```

// TODO: add in section for unbonding from validator set

## Queues

All queues objects are sorted by timestamp. The time used within any queue is
first rounded to the nearest nanosecond then sorted. The sortable time format
used is a slight modification of the RFC3339Nano and uses the the format string
`"2006-01-02T15:04:05.000000000"`. Notably this format:

- right pads all zeros
- drops the time zone info (uses UTC)

In all cases, the stored timestamp represents the maturation time of the queue
element.

### ValidatorQueue

For the purpose of tracking progress of unbonding validators the validator
queue is kept.

- ValidatorQueueTime: `0x43 | format(time) -> []sdk.ValAddress`

The stored object as each key is an array of validator operator addresses from
which the validator object can be accessed. Typically it is expected that only
a single validator record will be associated with a given timestamp however it is possible
that multiple validators exist in the queue at the same location.
