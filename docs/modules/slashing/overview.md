## Conceptual overview

### States

At any given time, there are any number of validators registered in the state machine.
Each block, the top `n = MaximumBondedValidators` validators who are not jailed become *bonded*, meaning that they may propose and vote on blocks.
Validators who are *bonded* are *at stake*, meaning that part or all of their stake and their delegators' stake is at risk if they commit a protocol fault.

### Slashing period

In order to mitigate the impact of initially likely categories of non-malicious protocol faults, the Cosmos Hub implements for each validator
a *slashing period*, in which the amount by which a validator can be slashed is capped at the punishment for the worst violation. For example,
if you misconfigure your HSM and double-sign a bunch of old blocks, you'll only be punished for the first double-sign (and then immediately jailed,
so that you have a chance to reconfigure your setup). This will still be quite expensive and desirable to avoid, but slashing periods somewhat blunt
the economic impact of unintentional misconfiguration.

Unlike the unbonding period, the slashing period doesn't have a fixed length. A new slashing period starts whenever a validator is bonded and ends
whenever the validator is unbonded (which will happen if the validator is jailed). The amount of tokens slashed relative to validator power for infractions
committed within the slashing period, whenever they are discovered, is capped at the punishment for the worst infraction
(which for the Cosmos Hub at launch will be double-signing a block).

#### ASCII timelines

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

<--------------------------->   
[----------C<sub>1</sub>--C<sub>2</sub>---C<sub>3</sub>---D<sub>1</sub>,D<sub>2</sub>,D<sub>3</sub>V<sub>u</sub>-----]

Multiple infractions are committed within a single slashing period then later discovered, at which point the validator is unbonded and slashed for only the worst infraction.

*Multiple infractions after rebonding*


<--------------------------->&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;<------------->  
[----------C<sub>1</sub>--C<sub>2</sub>---C<sub>3</sub>---D<sub>1</sub>,D<sub>2</sub>,D<sub>3</sub>V<sub>u</sub>---V<sub>b</sub>---C<sub>4</sub>----D<sub>4</sub>,V<sub>u</sub>--]

Multiple infractions are committed within a single slashing period then later discovered, at which point the validator is unbonded and slashed for only the worst infraction.
The validator then unjails themself and rebonds, then commits a fourth infraction - which is discovered and punished at the full amount, since a new slashing period started
when they unjailed and rebonded.

### Safety note

Slashing is capped fractionally per period, but the amount of total bonded stake associated with any given validator can change (by an unbounded amount) over that period.

For example, with MaxFractionSlashedPerPeriod = `0.5`, if a validator is initially slashed at `0.4` near the start of a period when they have 100 stake bonded,
then later slashed at `0.4` when they have `1000` stake bonded, the total amount slashed is just `40 + 100 = 140` (since the latter slash is capped at `0.1`) - 
whereas if they had `1000` stake bonded initially, the first offense would have been slashed for `400` stake and the total amount slashed would have been `400 + 100 = 500`.

This means that any slashing events which utilize the slashing period (are capped-per-period) **must also** jail the validator when the infraction is discovered.
Otherwise it would be possible for a validator to slash themselves intentionally at a low bond, then increase their bond but no longer be at stake since they would have already hit the `SlashedSoFar` cap.
