## Transaction & State Machine Interaction Overview

### Transactions

In this section we describe the processing of transactions for the `slashing` module.

#### TxUnjail

If a validator was automatically unbonded due to downtime and wishes to come back online &
possibly rejoin the bonded set, it must send `TxUnjail`:

```golang
type TxUnjail struct {
    ValidatorAddr sdk.AccAddress
}

handleMsgUnjail(tx TxUnjail)

    validator := getValidator(tx.ValidatorAddr)
    if validator == nil
      fail with "No validator found"

    if !validator.Jailed
      fail with "Validator not jailed, cannot unjail"

    info := getValidatorSigningInfo(operator)
    if BlockHeader.Time.Before(info.JailedUntil)
      fail with "Validator still jailed, cannot unjail until period has expired"

    // Update the start height so the validator won't be immediately unbonded again
    info.StartHeight = BlockHeight
    setValidatorSigningInfo(info)

    validator.Jailed = false
    setValidator(validator)

    return
```

If the validater has enough stake to be in the top hundred, they will be automatically rebonded,
and all delegators still delegated to the validator will be rebonded and begin to again collect
provisions and rewards.

### Interactions

#### Validator Bonded

#### Validator Slashed

#### Validator Unjailed

#### Slashing Period Cleanup

Once no evidence for a given slashing period can possibly be valid (the end time plus the unbonding period is less than the current time),
old slashing periods should be cleaned up.
