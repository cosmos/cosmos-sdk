## Transaction & State Machine Interaction Overview

### Conceptual overview

#### States

At any given time, there are any number of validator candidates registered in the state machine.
Each block, the top `n` candidates who are not jailed become *bonded*, meaning that they may propose and vote on blocks.
Validators who are *bonded* are *at stake*, meaning that part or all of their stake is at risk if they commit a protocol fault.

#### Slashing period

In order to mitigate the impact of initially likely categories of non-malicious protocol faults, the Cosmos Hub implements for each validator
a *slashing period*, in which the amount by which a validator can be slashed is capped at the punishment for the worst violation. For example,
if you misconfigure your HSM and double-sign a bunch of old blocks, you'll only be punished for the first double-sign (and then immediately jailed,
so that you have a chance to reconfigure your setup). This will still be quite expensive and desirable to avoid, but slashing periods somewhat blunt
the economic impact of unintentional misconfiguration.

A new slashing period starts whenever a validator is bonded and ends whenever the validator is unbonded (which will happen if the validator is jailed).
The amount of tokens slashed relative to validator power for infractions committed within the slashing period, whenever they are discovered, is capped
at the punishment for the worst infraction (which for the Cosmos Hub at launch will be double-signing a block).

##### ASCII timelines

*Code*

*[*   : timeline start  
*]*   : timeline end  
*<*   : slashing period start  
*>*   : slashing period end  
*C<sub>n</sub>* : infraction `n` committed  
*D<sub>n</sub>* : infraction `n` discovered  
*V<sub>b</sub>* : validator bonded  
*V<sub>u</sub>* : validator unbonded  

*Single infraction*

<----------------->   
[----------C<sub>1</sub>----D<sub>1</sub>,V<sub>u</sub>-----]

A single infraction is committed then later discovered, at which point the validator is unbonded and slashed at the full amount for the infraction.

*Multiple infractions*

<---------------------------->   
[----------C<sub>1</sub>--C<sub>2</sub>---C<sub>3</sub>---D<sub>1</sub>,D<sub>2</sub>,D<sub>3</sub>V<sub>u</sub>-----]

Multiple infractions are committed within a single slashing period then later discovered, at which point the validator is unbonded and slashed for only the worst infraction.

*Multiple infractions after rebonding*


<---------------------------->&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;<-------------->  
[----------C<sub>1</sub>--C<sub>2</sub>---C<sub>3</sub>---D<sub>1</sub>,D<sub>2</sub>,D<sub>3</sub>V<sub>u</sub>---V<sub>b</sub>---C<sub>4</sub>----D<sub>4</sub>,V<sub>u</sub>--]

Multiple infractions are committed within a single slashing period then later discovered, at which point the validator is unbonded and slashed for only the worst infraction.
The validator then unjails themself and rebonds, then commits a fourth infraction - which is discovered and punished at the full amount, since a new slashing period started
when they unjailed and rebonded.

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

If the validator has enough stake to be in the top hundred, they will be automatically rebonded,
and all delegators still delegated to the validator will be rebonded and begin to again collect
provisions and rewards.

### Interactions

In this section we describe the "hooks" - slashing module code that runs when other events happen.

#### Validator Bonded

Upon successful bonding of a validator (a given validator changing from "unbonded" state to "bonded" state,
which may happen on delegation, on unjailing, etc), we create a new `SlashingPeriod` structure for the
now-bonded validator, which `StartHeight` of the current block, `EndHeight` of `0` (sentinel value for not-yet-ended),
and `SlashedSoFar` of `0`:

```golang
onValidatorBonded(address sdk.ValAddress)

  slashingPeriod := SlashingPeriod{
      ValidatorAddr : address,
      StartHeight   : CurrentHeight,
      EndHeight     : 0,    
      SlashedSoFar  : 0,
  }
  setSlashingPeriod(slashingPeriod)
  
  return
```

#### Validator Unbonded

When a validator is unbonded, we update the in-progress `SlashingPeriod` with the current block as the `EndHeight`:

```golang
onValidatorUnbonded(address sdk.ValAddress)

  slashingPeriod = getSlashingPeriod(address, CurrentHeight)
  slashingPeriod.EndHeight = CurrentHeight
  setSlashingPeriod(slashingPeriod)

  return
```

#### Validator Slashed

When a validator is slashed, we look up the appropriate `SlashingPeriod` based on the validator
address and the time of infraction, cap the fraction slashed as `max(SlashFraction, SlashedSoFar)`
(which may be `0`), and update the `SlashingPeriod` with the increased `SlashedSoFar`:

```golang
beforeValidatorSlashed(address sdk.ValAddress, fraction sdk.Rat, infractionHeight int64)
  
  slashingPeriod = getSlashingPeriod(address, infractionHeight)
  totalToSlash = max(slashingPeriod.SlashedSoFar, fraction)
  slashingPeriod.SlashedSoFar = totalToSlash
  setSlashingPeriod(slashingPeriod)

  remainderToSlash = slashingPeriod.SlashedSoFar - totalToSlash
  fraction = remainderToSlash

  continue with slashing
```

##### Safety note

Slashing is capped fractionally per period, but the amount of total bonded stake associated with any given validator can change (by an unbounded amount) over that period.

For example, with MaxFractionSlashedPerPeriod = `0.5`, if a validator is initially slashed at `0.4` near the start of a period when they have 100 steak bonded,
then later slashed at `0.4` when they have `1000` steak bonded, the total amount slashed is just `40 + 100 = 140` (since the latter slash is capped at `0.1`) - 
whereas if they had `1000` steak bonded initially, the total amount slashed would have been `500`.

This means that any slashing events which utilize the slashing period (are capped-per-period) **must** *also* jail the validator when the infraction is discovered.
Otherwise it would be possible for a validator to slash themselves intentionally at a low bond, then increase their bond but no longer be at stake since they would have already hit the `SlashedSoFar` cap.

### State Cleanup

Once no evidence for a given slashing period can possibly be valid (the end time plus the unbonding period is less than the current time),
old slashing periods should be cleaned up. This will be implemented post-launch.
