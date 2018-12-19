# Staking Tombstone

## Abstract

In the current implementation of the `slashing` module, when the consensus engine informs the state machine of a validator's consensus fault, the validator is partially slashed, and put into a "jail period", a period of time in which they are not allowed to rejoin the validator set.  However, because of the nature of consensus faults and ABCI, there can be a delay between an infraction occuring, and evidence of the infraction reaching the state machine (this is one of the primary reasons for the existence of the unbonding period).

In the current system design, once a validator is put in the jail for a consensus fault, after the `JailPeriod` they are allowed to send a transaction to `unjail` themselves, and thus rejoin the validator set.

One of the "design desires" of `slashing` module is that if multiple infractions occur before evidence is executed (and a validator is put in jail), they should only be punished for single worst infraction, but not cumulatively.  For example, if the sequence of events is:
1. Validator A commits Infraction 1 (worth 30% slash)
2. Validator A commits Infraction 2 (worth 40% slash)
3. Validator A commits Infraction 3 (worth 35% slash)
4. Evidence for Infraction 1 reaches state machine (and validator is put in jail)
5. Evidence for Infraction 2 reaches state machine
6. Evidence for Infraction 3 reaches state machine
   
Only Infraction 2 should have it's slash take effect, as it is the highest.  This is done, so that in the case of the compromise of a validator's consensus key, they will only be punished once, even if the hacker double-signs many blocks.  Because, the unjailing has to be done with the validator's operator key, they have a chance to re-secure their consensus key, and then signal that they are ready using their operator key.  We call this period during which we track only the max infraction, the "slashing period".

> Note:  Currently, from what I can tell in the code, while the stake actively delegated to a validator follows the Max slash rule, but stake that is unbonding or redelegating away from a validator gets slashed cumulatively for consecutive faults.  I think this may be a bug that needs to be fixed independently of whether the tombstone is accepted or not.

Once, a validator rejoins by unjailing themselves, we begin a new slashing period; if they commit a new infraction after unjailing, it gets slashed cumulatively on top of the worst infraction from the previous slashing period.

However, while infractions are grouped based off of the slashing periods, because evidence can be submitted up to an `unbondingPeriod` after the infraction, we still have to allow for evidence to be submitted for previous slashing periods.  For example, if the sequence of events is:
1. Validator A commits Infraction 1 (worth 30% slash)
2. Validator A commits Infraction 2 (worth 40% slash)
3. Evidence for Infraction 1 reaches state machine (and Validator A is put in jail)
4. Validator A unjails
   
We are now in a new slashing period, however we still have to keep the door open for the previous infraction, as the evidence for Infraction 2 may still come in. As the number of slashing periods increase, it creates more complexity as we have to keep track of the highest infraction amount for every single slashing period.

> Note:  Currently, according to the `slashing` module spec, a new slashing period is created everytime a validator is unbonded then rebonded.  This should probably be changed to jailed/unjailed, as in the current system, let's say I compromised the key of the rank 100 validator, I could bond my own validator into and out of the validator set many times, in order to create as many slashing periods I want for the validator.  Then I can create infractions for each of the slashing periods I created for the validator, allowing me to get them multiply slashed. I'm not sure if this is how it is implemented in the code, or is just a mistake in the spec.  For the remainder of this, I will assume that we only start a new slashing period when a validator gets unjailed.

The maximum number of slashing periods is the `len(UnbondingPeriod) / len(JailPeriod)`.  The current defaults in Gaia for the `UnbondingPeriod` and `JailPeriod` are 3 weeks and 2 days, respectively.  This means there could potentially be up to 11 slashing periods concurrently being tracked per validator.  If we set the `JailPeriod >= UnbondingPeriod`, we only have to track 1 slashing period (i.e not have to track slashing periods).

Currently, in the jail period implementation, once a validator unjails, all of their delegators who are delegated to them (haven't unbonded / redelegated away), stay with them.  Given that consensus safety faults, are so egregious (way more so than liveness faults), it is probably prudent to have delegators not "auto-rebond" to the validator. Thus, we propose that instead of being put in a "jailed state" after evidence for a consensus safety fault, validators are instead put into a "tombstone state", which means the validator is kicked out of the validator set and not allowed to rejoin.  All of the stake that was delegated to it is put into an unbonding period.  The validator operator can create a new validator if they would like, preferably with a new consensus key (do we need to enforce this?  No rational validator should reuse the same compromised key lol), but they have to "reearn" their delegations back.

Doing this tombstone system and getting rid of the slashing period tracking, will make the `slashing` module way simpler, especially because we can remove the hooks between the `stake` and `slashing` modules.

> Note: The tombstone concept, only applies to byzantine faults reported over ABCI.  For slashable offenses tracked by the state machine (such as liveness faults), as there is not a delay between infraction and slashing, no slashing period tracking is needed. Also, a liveness bug probably isn't so egregious that it mandates force unbonding all delegations, and so the current jail system is adequate.


## Further improvements / Related proposals:

### Single slashing amount

Another optimization that can be made is that if we assume that all ABCI faults for Tendermint consensus are slashed at the same level, we don't have to keep track of "max slash".  Once an ABCI fault happens once, we don't have to worry about comparing potential future ones to find the max.

I believe current Tendermint ABCI faults are primarily:
- Unjustified precommits (double signs)
- Signing a precommit when you're in unbonding phase (can be used to trick light clients)

Do we want to punish these two faults at different levels?  If not, we can enact the above change.  Note:  This change may make sense for current Tendermint consensus, but maybe not for a different consensus algorithm or future versions of Tendermint that may want to punish at different levels (for example, partial slashing).

### Store infractions in state instead of iterating over unbonds/redelegations

Currently, every time evidence of a new fault comes in, we currently iterate over all of the unbonds/redelegations away from a validator to see if the slash affects them or not.  If it does, we decrease the "balance" of the `ubd` or `red`.  However, as the number of unbonds or redelegations can be very high, this might be very expensive.  Instead, we can store evidences for all infractions that happened in the last `Unbonding Period` in state, and then whenever a ubd or red hits maturity, it can check if it needs to be slashed by checking it against the last infraction that happened before they started unbonding/redelegating away.  Because we only need to store the infractions from the last unbon  


<!-- 
First, part of the design of the `stake` module is that delegators should be slashed for the infractions that happened during blocks that they were delegated to the offending validator, however, they should not be slashed for infractions that their voting power did not contribute to.

Thus, if the sequence of events is:
1. Validator A commits Infraction 1
2. Delegator X delegates to Validator A
3. Evidence for Infraction 1 reaches state machine
Delegator X should not be slashed.

Similarly, if the sequence of events is:
1. Delegator X delegates to Validator A
2. Delegator X unbonds from Validator A and begins unbonding period
3. Validator A commits Infraction 1
4. Evidence for Infraction 1 reaches state machine
5. Delegator X finishes unbonding.
Delegator X should not be slashed. -->
