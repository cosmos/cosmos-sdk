<!--
order: 1
-->

# Concepts

## States

At any given time, there are any number of validators registered in the state
machine. Each block, the top `MaxValidators` (defined by `x/staking`) validators
who are not jailed become _bonded_, meaning that they may propose and vote on
blocks. Validators who are _bonded_ are _at stake_, meaning that part or all of
their stake and their delegators' stake is at risk if they commit a protocol fault.

For each of these validators we keep a `ValidatorSigningInfo` record that contains
information partaining to validator's liveness and other infraction related
attributes.

## Tombstone Caps

In order to mitigate the impact of initially likely categories of non-malicious
protocol faults, the Cosmos Hub implements for each validator
a _tombstone_ cap, which only allows a validator to be slashed once for a double
sign fault. For example, if you misconfigure your HSM and double-sign a bunch of
old blocks, you'll only be punished for the first double-sign (and then immediately tombstombed). This will still be quite expensive and desirable to avoid, but tombstone caps
somewhat blunt the economic impact of unintentional misconfiguration.

Liveness faults do not have caps, as they can't stack upon each other. Liveness bugs are "detected" as soon as the infraction occurs, and the validators are immediately put in jail, so it is not possible for them to commit multiple liveness faults without unjailing in between.

## Infraction Timelines

To illustrate how the `x/slashing` module handles submitted evidence through
Tendermint consensus, consider the following examples:

**Definitions**:

_[_ : timeline start  
_]_ : timeline end  
_C<sub>n</sub>_ : infraction `n` committed  
_D<sub>n</sub>_ : infraction `n` discovered  
_V<sub>b</sub>_ : validator bonded  
_V<sub>u</sub>_ : validator unbonded

### Single Double Sign Infraction

<----------------->
[----------C<sub>1</sub>----D<sub>1</sub>,V<sub>u</sub>-----]

A single infraction is committed then later discovered, at which point the
validator is unbonded and slashed at the full amount for the infraction.

### Multiple Double Sign Infractions

<--------------------------->
[----------C<sub>1</sub>--C<sub>2</sub>---C<sub>3</sub>---D<sub>1</sub>,D<sub>2</sub>,D<sub>3</sub>V<sub>u</sub>-----]

Multiple infractions are committed and then later discovered, at which point the
validator is jailed and slashed for only one infraction. Because the validator
is also tombstoned, they can not rejoin the validator set.
