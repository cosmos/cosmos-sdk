<!--
order: 3
-->

# Messages

In this section we describe the processing of messages for the `slashing` module.

## Unjail

If a validator was automatically unbonded due to downtime and wishes to come back online &
possibly rejoin the bonded set, it must send `MsgUnjail`:

```protobuf
// MsgUnjail is an sdk.Msg used for unjailing a jailed validator, thus returning
// them into the bonded validator set, so they can begin receiving provisions
// and rewards again.
message MsgUnjail {
  string validator_addr = 1;
}
```

Below is a pseudocode of the `MsgSrv/Unjail` RPC:

```go
unjail(tx MsgUnjail)
    validator = getValidator(tx.ValidatorAddr)
    if validator == nil
      fail with "No validator found"

    if getSelfDelegation(validator) == 0
      fail with "validator must self delegate before unjailing"

    if !validator.Jailed
      fail with "Validator not jailed, cannot unjail"

    info = GetValidatorSigningInfo(operator)
    if info.Tombstoned
      fail with "Tombstoned validator cannot be unjailed"
    if block time < info.JailedUntil
      fail with "Validator still jailed, cannot unjail until period has expired"

    validator.Jailed = false
    setValidator(validator)

    return
```

If the validator has enough stake to be in the top `n = MaximumBondedValidators`, it will be automatically rebonded,
and all delegators still delegated to the validator will be rebonded and begin to again collect
provisions and rewards.
