# Distribution

## Overview

Collected fees are pooled globally and divided out passively to validators and
delegators. Each validator has the opportunity to charge commission to the
delegators on the fees collected on behalf of the delegators by the validators.
Fees are paid directly into a global fee pool, and validator proposer-reward
pool. Due to the nature of passive accounting whenever changes to parameters
which affect the rate of fee distribution occurs, withdrawal of fees must also
occur when: 
 
 - withdrawing one must withdrawal the maximum amount they are entitled
   too, leaving nothing in the pool, 
 - bonding, unbonding, or re-delegating tokens to an existing account a
   full withdrawal of the fees must occur (as the rules for lazy accounting
   change), 
 - a validator chooses to change the commission on fees, all accumulated 
   commission fees must be simultaneously withdrawn.

The above scenarios are covered in `triggers.md`.

The distribution mechanism outlines herein is used to lazily distribute the
following between validators and associated delegators:
 - multi-token fees to be socially distributed, 
 - proposer reward pool, 
 - inflated atom provisions, and
 - validator commission on all rewards earned by their delegators stake

Fees are pooled within a global pool, as well as validator specific
proposer-reward pools.  The mechanisms used allow for validators and delegators
to independently and lazily  withdrawn their rewards.  As a part of the lazy
computations adjustment factors must be maintained for each validator and
delegator to determine the true proportion of fees in each pool which they are
entitled too.  Adjustment factors are updated every time a validator or
delegator's voting power changes.  Validators and delegators must withdraw all
fees they are entitled too before they can change their portion of bonded
Atoms. 

## Affect on Staking


Charging commission on Atom provisions while also allowing for Atom-provisions
to be auto-bonded (distributed directly to the validators bonded stake) is
problematic within DPoS.  Fundamentally these two mechnisms are mutually
exclusive. If there are atoms commissions and auto-bonding Atoms, the portion
of Atoms the fee distribution calculation would become very large as the Atom
portion for each delegator would change each block making a withdrawal of fees
for a delegator require a calculation for every single block since the last
withdrawal. In conclusion we can only have atom commission and unbonded atoms
provisions, or bonded atom provisions with no Atom commission, and we elect to
implement the former. Stakeholders wishing to rebond their provisions may elect
to set up a script to periodically withdraw and rebond fees. 

