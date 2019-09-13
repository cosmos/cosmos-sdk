# Messages

In this section we describe the processing of the staking messages and the corresponding updates to the state. All created/modified state objects specified by each message are defined within the [state](./02_state.md) section.

## MsgCreateValidator

A validator is created using the `MsgCreateValidator` message, if `AcceptAllValidators` is set to `true`.

```go
type MsgCreateValidator struct {
    Description    Description
    ValidatorAddr  sdk.ValAddress
    PubKey         crypto.PubKey
}
```

This message is expected to fail if:

- another validator with this operator address is already registered
- another validator with this pubkey is already registered
- the description fields are too large
- `AcceptAllValidators` is set to `false` <!-- TODO: check if we can make it automatic so the user doesn't have be aware of the param. -->

This message creates and stores the `Validator` object at appropriate indexes.
The validator always starts as unbonded but may be bonded in the first end-block.

## MsgEditValidator

The `Description` of a validator can be updated using the `MsgEditCandidacy`.

```go
type MsgEditCandidacy struct {
    Description     Description
    ValidatorAddr   sdk.ValAddress
}
```

This message is expected to fail if:

- the description fields are too large

This message stores the updated `Validator` object.

## MsgProposeCreateValidator

A validator is created using the `MsgProposeCreateValidator` message, if `AcceptAllValidators` is set to `false`.

```go
type MsgProposeCreateValidator struct {
	Title       string                   // title of the validator
	Description string                  // description of validator
	Validator   NewValidatorCreatation // validator details
}

type NewValidatorCreatation struct {
	Description      stakingtypes.Description  // description of validator
	ValidatorAddress sdk.ValAddress           // validator address
	PubKey           crypto.PubKey           // public key of the validator
}

```

This message is expected to fail if:

- another validator with this operator address is already registered
- another validator with this pubkey is already registered
- the description fields are too large
- if `AcceptAllValidators` is set to `true`

This message creates and stores the `Validator` object at appropriate indexes.
The validator always starts as unbonded but may be bonded in the first end-block.

## MsgProposeNewWeight

```go
type MsgProposeNewWeight struct {
	Title       string                    // title of the validator
	Description string                   // description of validator
	Validator   ValidatorNewWeight // validator of which the increase is proposed for
}

type ValidatorNewWeight struct {
	ValidatorAddress sdk.ValAddress     // validator address
	PubKey           crypto.PubKey     // public key of the validator
	NewWeight        sdk.Int          // new weight
}
```

This message is expected to fail if:

- `IncreaseWeight` is set to false, if the proposed weight is greater than the current weight.

This message stores the updated `Validator` object.

## MsgBeginUnbonding

The begin unbonding message allows validator to remove themselves from the validator set.

```go
type MsgBeginUnbonding struct {
  ValidatorAddr sdk.ValAddress
}
```

This message is expected to fail if:

- the validator doesn't exist
- existing `UnbondingDelegation` has maximum entries as defined by `params.MaxEntries`

When this message is processed the following actions occur:

- The validator will be removed from the validator set after the predefined `UnbondingTime` has passed
