# Distribution

## Overview

This _simple_ distribution mechanism describes a functional way to passively 
distribute rewards between validator and delegators. Note that this mechanism does 
not distribute funds in as precisely as active reward distribution and will therefore
be upgraded in the future. 

The mechanism operates as follows. Collected rewards are pooled globally and
divided out passively to validators and delegators. Each validator has the
opportunity to charge commission to the delegators on the rewards collected on
behalf of the delegators by the validators. Fees are paid directly into a
global reward pool, and validator proposer-reward pool. Due to the nature of
passive accounting, whenever changes to parameters which affect the rate of reward
distribution occurs, withdrawal of rewards must also occur.

 - Whenever withdrawing, one must withdraw the maximum amount they are entitled
   too, leaving nothing in the pool. 
 - Whenever bonding, unbonding, or re-delegating tokens to an existing account, a
   full withdrawal of the rewards must occur (as the rules for lazy accounting
   change).
 - Whenever a validator chooses to change the commission on rewards, all accumulated 
   commission rewards must be simultaneously withdrawn.

The above scenarios are covered in `hooks.md`.

The distribution mechanism outlines herein is used to lazily distribute the
following rewards between validators and associated delegators:
 - multi-token fees to be socially distributed, 
 - proposer reward pool, 
 - inflated atom provisions, and
 - validator commission on all rewards earned by their delegators stake

Fees are pooled within a global pool, as well as validator specific
proposer-reward pools. The mechanisms used allow for validators and delegators
to independently and lazily withdraw their rewards.  

## Shortcomings 

As a part of the lazy computations, each delegator holds an accumulation term
specific to each validator which is used to estimate what their approximate
fair portion of tokens held in the global fee pool is owed to them. 

```
entitlement = delegator-accumulation / all-delegators-accumulation
```

Under the circumstance that there were constant and equal flow of incoming
reward tokens every block, this distribution mechanism would be equal to the
active distribution (distribute individually to all delegators each block).
However this is unrealistic so deviations from the active distribution will
occur based on fluctuations of incoming reward tokens as well as timing of
reward withdrawal by other delegators. 

If you happen to know that incoming rewards are about significantly move up,
you are incentivized to not withdraw until after this event, increasing the
worth of your existing _accum_.

## Affect on Staking

Charging commission on Atom provisions while also allowing for Atom-provisions
to be auto-bonded (distributed directly to the validators bonded stake) is
problematic within DPoS. Fundamentally these two mechanisms are mutually
exclusive. If there are Atom commissions and auto-bonding Atoms, the portion
of Atoms the reward distribution calculation would become very large as the Atom
portion for each delegator would change each block making a withdrawal of rewards
for a delegator require a calculation for every single block since the last
withdrawal. In conclusion, we can only have Atom commission and unbonded atoms
provisions or bonded atom provisions with no Atom commission, and we elect to
implement the former. Stakeholders wishing to rebond their provisions may elect
to set up a script to periodically withdraw and rebond rewards. 
