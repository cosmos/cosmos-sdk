# Distribution

## Overview

Collected fees are pooled globally and divided out passively to validators and
delegators. Each validator has the opportunity to charge commission to the
delegators on the fees collected on behalf of the delegators by the validators.
Fees are paid directly into a global fee pool, and validator proposer-reward
pool. Due to the nature of of passive accounting whenever changes to parameters
which affect the rate of fee distribution occurs, withdrawal of fees must also
occur. 
 
 - when withdrawing one must withdrawal the maximum amount they are entitled
   too, leaving nothing in the pool, 
 - when bonding, unbonding, or re-delegating tokens to an existing account a
   full withdrawal of the fees must occur (as the rules for lazy accounting
   change), 
 - when a validator chooses to change the commission on fees, all accumulated 
   commission fees must be simultaneously withdrawn.

The above scenarios are covered is `triggers.md`.

The distribution mechanism outlines herein is used to lazily distribute the
following between validators and associated delegators:
 - Multi-token fees to be socially distributed 
 - Proposer reward pool
 - Inflated atom provisions
 - Validator commission on delegators 

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

Commission on Atom provisions and having autobonded-atom-provisions have been
shown to be  mutually exclusive. Fundamentally if there are atoms commissions
and autobonding, the portion of atoms the fee distribution calculation would
become very large as the atom portion for each delegator would change each
block making a withdrawal of fees for a delegator require a calculation for
every single block since the last withdrawal. In conclusion we can only have atom
commission and unbonded atoms provisions, or bonded atom provisions and no atom
commission. 

