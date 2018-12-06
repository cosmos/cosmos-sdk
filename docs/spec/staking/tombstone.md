# Staking Tombstone

## Abstract

In the current implementation of the `stake` module, when the consensus engine informs the state machine of a validator's consensus fault, the validator is partially slashed, and put into a "jail period", a period of time in which they are not allowed to rejoin the validator set.  However, because of the nature of consensus faults and ABCI, there can be a delay between an infraction occuring, and evidence of the infraction reaching the state machine (this is one of the primary reasons for the existence of the unbonding period).

Part of the design of the `stake` module is that delegators should be slashed for the infractions that happened during blocks that they were delegated to the offending validator, however, they should not be slashed for infractions that their voting power did not contribute to.

Thus, if the sequence of events is:
1. Validator A commits Infraction 1
2. Delegator X delegates to Validator A
3. Evidence for Infraction A reaches state machine
Delegator X should not be slashed.

However, in the current system, because delegator can cause 

In order to think about these cases, we should list the "events" that we have to consider, and then how to deal with each of the cases that happen.



Given the 4 following events:
1 - Validator A Infraction
2 - Delegator X Bonds to Validator A
3 - Validator A Evidence

For this reason, if a validator's 




<------------------------------------------------------------>
    A       B         C        D        E





Note: The tombstone concept, only applies to byzantine faults reported over ABCI.  For slashable offenses tracked by the state machine (such as liveness faults), as there is not a delay between infraction and slashing, the current jail system is adequate.