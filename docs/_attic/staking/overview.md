//TODO update .rst

# Staking Module

## Overview

The Cosmos Hub is a Tendermint-based Delegated Proof of Stake (DPos) blockchain
system that serves as a backbone of the Cosmos ecosystem. It is operated and
secured by an open and globally decentralized set of validators. Tendermint is
a Byzantine fault-tolerant distributed protocol for consensus among distrusting
parties, in this case the group of validators which produce the blocks for the
Cosmos Hub.  To avoid the nothing-at-stake problem, a validator in Tendermint
needs to lock up coins in a bond deposit. Each bond's atoms are illiquid, they
cannot be transferred - in order to become liquid, they must be unbonded, a
process which will take 3 weeks by default at Cosmos Hub launch. Tendermint
protocol messages are signed by the validator's private key and are therefor
attributable.  Validators acting outside protocol specifications can be made
accountable through punishing by slashing (burning) their bonded Atoms. On the
other hand, validators are rewarded for their service of securing blockchain
network by the inflationary provisions and transactions fees. This incentivizes
correct behavior of the validators and provides the economic security of the
network.

The native token of the Cosmos Hub is called the Atom; becoming a validator of the
Cosmos Hub requires holding Atoms. However, not all Atom holders are validators
of the Cosmos Hub. More precisely, there is a selection process that determines
the validator set as a subset of all validators (Atom holders that
want to become a validator). The other option for Atom holders is to delegate
their atoms to validators, i.e., being a delegator. A delegator is an Atom
holder that has put its Atoms at stake by delegating it to a validator. By bonding
Atoms to secure the network (and taking a risk of being slashed in case of
misbehaviour), a user is rewarded with inflationary provisions and transaction
fees proportional to the amount of its bonded Atoms.  The Cosmos Hub is
designed to efficiently facilitate a small numbers of validators (hundreds),
and large numbers of delegators (tens of thousands).  More precisely, it is the
role of the Staking module of the Cosmos Hub to support various staking
functionality including validator set selection, delegating, bonding and
withdrawing Atoms, and the distribution of inflationary provisions and
transaction fees.

## Basic Terms and Definitions

* Cosmsos Hub - a Tendermint-based Delegated Proof of Stake (DPos)
    blockchain system
* Atom - native token of the Cosmsos Hub
* Atom holder - an entity that holds some amount of Atoms 
* Pool - Global object within the Cosmos Hub which accounts global state
    including the total amount of bonded, unbonding, and unbonded atoms
* Validator Share - Share which a validator holds to represent its portion of
    bonded, unbonding or unbonded atoms in the pool 
* Delegation Share - Shares which a delegation bond holds to represent its
    portion of bonded, unbonding or unbonded shares in a validator
* Bond Atoms - a process of locking Atoms in a delegation share which holds them
    under protocol control. 
* Slash Atoms - the process of burning atoms in the pool and assoiated
    validator shares of a misbehaving validator, (not behaving according to the
    protocol specification). This process devalues the worth of delegation shares 
    of the given validator
* Unbond Shares - Process of retrieving atoms from shares. If the shares are
    bonded the shares must first remain in an inbetween unbonding state for the
    duration of the unbonding period
* Redelegating Shares - Process of redelegating atoms from one validator to
    another. This process is instantaneous, but the redelegated atoms are
    retrospecively slashable if the old validator is found to misbehave for any
    blocks before the redelegation. These atoms are simultaniously slashable
    for any new blocks which the new validator misbehavess
* Validator - entity with atoms which is either actively validating the Tendermint
    protocol (bonded validator) or vying to validate .
* Bonded Validator - a validator whose atoms are currently bonded and liable to
    be slashed. These validators are to be able to sign protocol messages for
    Tendermint consensus. At Cosmos Hub genesis there is a maximum of 100
    bonded validator positions. Only Bonded Validators receive atom provisions
    and fee rewards. 
* Delegator - an Atom holder that has bonded Atoms to a validator
* Unbonding period - time required in the unbonding state when unbonding
    shares. Time slashable to old validator after a redelegation. Time for which
    validators can be slashed after an infraction. To provide the requisite
    cryptoeconomic security guarantees, all of these must be equal.
* Atom provisions - The process of increasing the Atom supply.  Atoms are
    periodically created on the Cosmos Hub and issued to bonded Atom holders.
    The goal of inflation is to incentize most of the Atoms in existence to be
    bonded. Atoms are distributed unbonded and using the fee_distribution mechanism
* Transaction fees - transaction fee is a fee that is included in a Cosmsos Hub
    transaction. The fees are collected by the current validator set and 
    distributed among validators and delegators in proportion to their bonded 
    Atom share
* Commission fee - a fee taken from the transaction fees by a validator for 
    their service 

## The pool and the share

At the core of the Staking module is the concept of a pool which denotes a
collection of Atoms contributed by different Atom holders. There are three
pools in the Staking module: the bonded, unbonding, and unbonded pool.  Bonded
Atoms are part of the global bonded pool. If a validator or delegator wants to
unbond its shares, these Shares are moved to the the unbonding pool for the
duration of the unbonding period. From here normally Atoms will be moved
directly into the delegators wallet, however under the situation thatn an
entire validator gets unbonded, the Atoms of the delegations will remain with
the validator and moved to the unbonded pool.  For each pool, the total amount
of bonded, unbonding, or unbonded Atoms are tracked as well as the current
amount of issued pool-shares, the specific holdings of these shares by
validators are tracked in protocol by the validator object. 

A share is a unit of Atom distribution and the value of the share
(share-to-atom exchange rate) can change during system execution. The
share-to-atom exchange rate can be computed as:

`share-to-atom-exchange-rate = size of the pool / ammount of issued shares`

Then for each validator (in a per validator data structure) the protocol keeps
track of the amount of shares the validator owns in a pool. At any point in
time, the exact amount of Atoms a validator has in the pool can be computed as
the number of shares it owns multiplied with the current share-to-atom exchange
rate:

`validator-coins = validator.Shares * share-to-atom-exchange-rate`

The benefit of such accounting of the pool resources is the fact that a
modification to the pool from bonding/unbonding/slashing of Atoms affects only
global data (size of the pool and the number of shares) and not the related
validator data structure, i.e., the data structure of other validators do not
need to be modified. This has the advantage that modifying global data is much
cheaper computationally than modifying data of every validator. Let's explain
this further with several small examples: 

XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
XXX TODO make way less verbose lets use bullet points to describe the example
XXX Also need to update to not include bonded atom provisions all atoms are
XXX   redistributed with the fee pool now

We consider initially 4 validators p1, p2, p3 and p4, and that each validator 
has bonded 10 Atoms to the bonded pool. Furthermore, let's assume that we have 
issued initially 40 shares (note that the initial distribution of the shares, 
i.e., share-to-atom exchange rate can be set to any meaningful value), i.e., 
share-to-atom-ex-rate = 1 atom per share. Then at the global pool level we 
have, the size of the pool is 40 Atoms, and the amount of issued shares is 
equal to 40. And for each validator we store in their corresponding data 
structure that each has 10 shares of the bonded pool. Now lets assume that the 
validator p4 starts process of unbonding of 5 shares. Then the total size of 
the pool is decreased and now it will be 35 shares and the amount of Atoms is 
35 . Note that the only change in other data structures needed is reducing the 
number of shares for a validator p4 from 10 to 5.

Let's consider now the case where a validator p1 wants to bond 15 more atoms to
the pool. Now the size of the pool is 50, and as the exchange rate hasn't 
changed (1 share is still worth 1 Atom), we need to create more shares, i.e. we
now have 50 shares in the pool in total. Validators p2, p3 and p4 still have 
(correspondingly) 10, 10 and 5 shares each worth of 1 atom per share, so we 
don't need to modify anything in their corresponding data structures. But p1 
now has 25 shares, so we update the amount of shares owned by p1 in its 
data structure. Note that apart from the size of the pool that is in Atoms, all
other data structures refer only to shares.

Finally, let's consider what happens when new Atoms are created and added to 
the pool due to inflation. Let's assume that the inflation rate is 10 percent 
and that it is applied to the current state of the pool. This means that 5 
Atoms are created and added to the pool and that each validator now 
proportionally increase it's Atom count. Let's analyse how this change is 
reflected in the data structures. First, the size of the pool is increased and 
is now 55 atoms. As a share of each validator in the pool hasn't changed, this 
means that the total number of shares stay the same (50) and that the amount of
shares of each validator stays the same (correspondingly 25, 10, 10, 5). But 
the exchange rate has changed and each share is now worth 55/50 Atoms per 
share, so each validator has effectively increased amount of Atoms it has.  So 
validators now have (correspondingly) 55/2, 55/5, 55/5 and 55/10 Atoms. 

The concepts of the pool and its shares is at the core of the accounting in the 
Staking module. It is used for managing the global pools (such as bonding and 
unbonding pool), but also for distribution of Atoms between validator and its 
delegators (we will explain this in section X).

#### Delegator shares

A validator is, depending on its status, contributing Atoms to either the
unbonding or unbonded pool - the validator in turn holds some amount of pool
shares.  Not all of a validator's Atoms (and respective shares) are necessarily
owned by the validator, some may be owned by delegators to that validator. The
mechanism for distribution of Atoms (and shares) between a validator and its
delegators is based on a notion of delegator shares. More precisely, every
validator is issuing (local) delegator shares
(`Validator.IssuedDelegatorShares`) that represents some portion of global
shares managed by the validator (`Validator.GlobalStakeShares`). The principle
behind managing delegator shares is the same as described in [Section](#The
pool and the share). We now illustrate it with an example.

XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
XXX TODO make way less verbose lets use bullet points to describe the example
XXX Also need to update to not include bonded atom provisions all atoms are
XXX   redistributed with the fee pool now

Let's consider 4 validators p1, p2, p3 and p4, and assume that each validator 
has bonded 10 Atoms to the bonded pool. Furthermore, let's assume that we have 
issued initially 40 global shares, i.e., that 
`share-to-atom-exchange-rate = 1 atom per share`. So we will set 
`GlobalState.BondedPool = 40` and `GlobalState.BondedShares = 40` and in the 
Validator data structure of each validator `Validator.GlobalStakeShares = 10`. 
Furthermore, each validator issued 10 delegator shares which are initially 
owned by itself, i.e., `Validator.IssuedDelegatorShares = 10`, where 
`delegator-share-to-global-share-ex-rate = 1 global share per delegator share`.
Now lets assume that a delegator d1 delegates 5 atoms to a validator p1 and 
consider what are the updates we need to make to the data structures. First, 
`GlobalState.BondedPool = 45` and `GlobalState.BondedShares = 45`. Then, for 
validator p1 we have `Validator.GlobalStakeShares = 15`, but we also need to 
issue also additional delegator shares, i.e., 
`Validator.IssuedDelegatorShares = 15` as the delegator d1 now owns 5 delegator
shares of validator p1, where each delegator share is worth 1 global shares, 
i.e, 1 Atom. Lets see now what happens after 5 new Atoms are created due to 
inflation. In that case, we only need to update `GlobalState.BondedPool` which 
is now equal to 50 Atoms as created Atoms are added to the bonded pool. Note 
that the amount of global and delegator shares stay the same but they are now 
worth more as share-to-atom-exchange-rate is now worth 50/45 Atoms per share. 
Therefore, a delegator d1 now owns:

`delegatorCoins = 5 (delegator shares) * 1 (delegator-share-to-global-share-ex-rate) * 50/45 (share-to-atom-ex-rate) = 5.55 Atoms`  

