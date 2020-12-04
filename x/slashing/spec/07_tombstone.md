<!--
order: 7
-->

# Staking Tombstone

## Abstract

In the current implementation of the `slashing` module, when the consensus engine
informs the state machine of a validator's consensus fault, the validator is
partially slashed, and put into a "jail period", a period of time in which they
are not allowed to rejoin the validator set. However, because of the nature of
consensus faults and ABCI, there can be a delay between an infraction occurring,
and evidence of the infraction reaching the state machine (this is one of the
primary reasons for the existence of the unbonding period).

> Note: The tombstone concept, only applies to faults that have a delay between
> the infraction occurring and evidence reaching the state machine. For example,
> evidence of a validator double signing may take a while to reach the state machine
> due to unpredictable evidence gossip layer delays and the ability of validators to
> selectively reveal double-signatures (e.g. to infrequently-online light clients).
> Liveness slashing, on the other hand, is detected immediately as soon as the
> infraction occurs, and therefore no slashing period is needed. A validator is
> immediately put into jail period, and they cannot commit another liveness fault
> until they unjail. In the future, there may be other types of byzantine faults
> that have delays (for example, submitting evidence of an invalid proposal as a transaction).
> When implemented, it will have to be decided whether these future types of
> byzantine faults will result in a tombstoning (and if not, the slash amounts
> will not be capped by a slashing period).

In the current system design, once a validator is put in the jail for a consensus
fault, after the `JailPeriod` they are allowed to send a transaction to `unjail`
themselves, and thus rejoin the validator set.

One of the "design desires" of the `slashing` module is that if multiple
infractions occur before evidence is executed (and a validator is put in jail),
they should only be punished for single worst infraction, but not cumulatively.
For example, if the sequence of events is:

1. Validator A commits Infraction 1 (worth 30% slash)
2. Validator A commits Infraction 2 (worth 40% slash)
3. Validator A commits Infraction 3 (worth 35% slash)
4. Evidence for Infraction 1 reaches state machine (and validator is put in jail)
5. Evidence for Infraction 2 reaches state machine
6. Evidence for Infraction 3 reaches state machine

Only Infraction 2 should have its slash take effect, as it is the highest. This
is done, so that in the case of the compromise of a validator's consensus key,
they will only be punished once, even if the hacker double-signs many blocks.
Because, the unjailing has to be done with the validator's operator key, they
have a chance to re-secure their consensus key, and then signal that they are
ready using their operator key. We call this period during which we track only
the max infraction, the "slashing period".

Once, a validator rejoins by unjailing themselves, we begin a new slashing period;
if they commit a new infraction after unjailing, it gets slashed cumulatively on
top of the worst infraction from the previous slashing period.

However, while infractions are grouped based off of the slashing periods, because
evidence can be submitted up to an `unbondingPeriod` after the infraction, we
still have to allow for evidence to be submitted for previous slashing periods.
For example, if the sequence of events is:

1. Validator A commits Infraction 1 (worth 30% slash)
2. Validator A commits Infraction 2 (worth 40% slash)
3. Evidence for Infraction 1 reaches state machine (and Validator A is put in jail)
4. Validator A unjails

We are now in a new slashing period, however we still have to keep the door open
for the previous infraction, as the evidence for Infraction 2 may still come in.
As the number of slashing periods increase, it creates more complexity as we have
to keep track of the highest infraction amount for every single slashing period.

> Note: Currently, according to the `slashing` module spec, a new slashing period
> is created every time a validator is unbonded then rebonded. This should probably
> be changed to jailed/unjailed. See issue [#3205](https://github.com/cosmos/cosmos-sdk/issues/3205)
> for further details. For the remainder of this, I will assume that we only start
> a new slashing period when a validator gets unjailed.

The maximum number of slashing periods is the `len(UnbondingPeriod) / len(JailPeriod)`.
The current defaults in Gaia for the `UnbondingPeriod` and `JailPeriod` are 3 weeks
and 2 days, respectively. This means there could potentially be up to 11 slashing
periods concurrently being tracked per validator. If we set the `JailPeriod >= UnbondingPeriod`,
we only have to track 1 slashing period (i.e not have to track slashing periods).

Currently, in the jail period implementation, once a validator unjails, all of
their delegators who are delegated to them (haven't unbonded / redelegated away),
stay with them. Given that consensus safety faults are so egregious
(way more so than liveness faults), it is probably prudent to have delegators not
"auto-rebond" to the validator. Thus, we propose setting the "jail time" for a
validator who commits a consensus safety fault, to `infinite` (i.e. a tombstone state).
This essentially kicks the validator out of the validator set and does not allow
them to re-enter the validator set. All of their delegators (including the operator themselves)
have to either unbond or redelegate away. The validator operator can create a new
validator if they would like, with a new operator key and consensus key, but they
have to "re-earn" their delegations back. To put the validator in the tombstone
state, we set `DoubleSignJailEndTime` to `time.Unix(253402300800)`, the maximum
time supported by Amino.

Implementing the tombstone system and getting rid of the slashing period tracking
will make the `slashing` module way simpler, especially because we can remove all
of the hooks defined in the `slashing` module consumed by the `staking` module
(the `slashing` module still consumes hooks defined in `staking`).

### Single slashing amount

Another optimization that can be made is that if we assume that all ABCI faults
for Tendermint consensus are slashed at the same level, we don't have to keep
track of "max slash". Once an ABCI fault happens, we don't have to worry about
comparing potential future ones to find the max.

Currently the only Tendermint ABCI fault is:

- Unjustified precommits (double signs)

It is currently planned to include the following fault in the near future:

- Signing a precommit when you're in unbonding phase (needed to make light client bisection safe)

Given that these faults are both attributable byzantine faults, we will likely
want to slash them equally, and thus we can enact the above change.

> Note: This change may make sense for current Tendermint consensus, but maybe
> not for a different consensus algorithm or future versions of Tendermint that
> may want to punish at different levels (for example, partial slashing).
