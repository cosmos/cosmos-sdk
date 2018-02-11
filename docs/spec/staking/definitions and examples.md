# Staking Module

## Basic Terms and Definitions

- Cosmsos Hub - a Tendermint-based Proof of Stake blockchain system
- Atom - native token of the Cosmsos Hub
- Atom holder - an entity that holds some amount of Atoms 
- Candidate - an Atom holder that is actively involved in Tendermint blockchain protocol (running Tendermint Full Node
TODO: add link to Full Node definition) and is competing with other candidates to be elected as a Validator 
(TODO: add link to Validator definition))
- Validator - a candidate that is currently selected among a set of candidates to be able to sign protocol messages 
in the Tendermint consensus protocol 
- Delegator - an Atom holder that has bonded any of its Atoms by delegating them to a validator (or a candidate) 
- Bonding Atoms - a process of locking Atoms in a bond deposit (putting Atoms under protocol control).
Atoms are always bonded through a validator (or candidate) process. Bonded atoms can be slashed (burned) in case a 
validator process misbehaves (does not behave according to a protocol specification). Atom holder can regain access 
to its bonded Atoms (if they are not slashed in the meantime), i.e., they can be moved to its account, 
after Unbonding period has expired. 
- Unbonding period - a period of time after which Atom holder gains access to its bonded Atoms (they can be withdrawn 
to a user account) or they can re-delegate  
- Inflationary provisions - inflation is a process of increasing Atom supply. Atoms are being created in the process of 
(Cosmos Hub) blocks creation. Owners of bonded atoms are rewarded for securing network with inflationary provisions 
proportional to it's bonded Atom share.
- Transaction fees - transaction fee is a fee that is included in the Cosmsos Hub transactions. The fees are collected 
by the current validator set and distributed among validators and delegators in proportion to it's bonded Atom share.       
- Commission fee - a fee taken from the transaction fees by a validator for it's service 

## The pool and the share

At the core of the Staking module is the concept of a pool which denotes collection of Atoms contributed by different 
Atom holders. There are two global pools in the Staking module: bonded pool and unbonded pool. Bonded Atoms are part 
of the global bonded pool. On the other side, if a candidate or delegator wants to unbond its Atoms, those Atoms are 
kept in the unbonding pool for a duration of the unbonding period. In the Staking module, a pool is logical concept, 
i.e., there is no pool data structure that would be responsible for managing pool resources. Instead, it is managed 
in a distributed way. More precisely, at the global level, for each pool, we track only the total amount of 
(bonded or unbonded) Atoms and a current amount of issued shares. A share is a unit of Atom distribution and the 
value of the share (share-to-atom exchange rate) is changing during the system execution. The 
share-to-atom exchange rate can be computed as:

`share-to-atom-ex-rate = size of the pool / ammount of issued shares`

Then for each candidate (in a per candidate data structure) we keep track of an amount of shares the candidate is owning 
in a pool. At any point in time, the exact amount of Atoms a candidate has in the pool
can be computed as the number of shares it owns multiplied with the share-to-atom exchange rate:

`candidate-coins = candidate.Shares * share-to-atom-ex-rate`

The benefit of such accounting of the pool resources is the fact that a modification to the pool because of 
bonding/unbonding/slashing/provisioning of atoms affects only global data (size of the pool and the number of shares) 
and the related validator/candidate data structure, i.e., the data structure of other validators do not need to be 
modified. Let's explain this further with several small examples: 

We consider initially 4 validators p1, p2, p3 and p4, and that each validator has bonded 10 Atoms 
to a bonded pool. Furthermore, let's assume that we have issued initially 40 shares (note that the initial distribution 
of the shares, i.e., share-to-atom exchange rate can be set to any meaningful value), i.e., 
share-to-atom-ex-rate = 1 atom per share. Then at the global pool level we have, the size of the pool is 40 Atoms, and
the amount of issued shares is equal to 40. And for each validator we store in their corresponding data structure
that each has 10 shares of the bonded pool. Now lets assume that the validator p4 starts process of unbonding of 5 
shares. Then the total size of the pool is decreased and now it will be 35 shares and the amount of Atoms is 35. 
Note that the only change in other data structures needed is reducing the number of shares for a validator p4 from 10 
to 5.

Let's consider now the case where a validator p1 wants to bond 15 more atoms to the pool. Now the size of the pool
is 50, and as the exchange rate hasn't changed (1 share is still worth 1 Atom), we need to create more shares, 
i.e. we now have 50 shares in the pool in total.
Validators p2, p3 and p4 still have (correspondingly) 10, 10 and 5 shares each worth of 1 atom per share, so we 
don't need to modify anything in their corresponding data structures. But p1 now has 25 shares, so we update the amount
of shares owned by the p1 in its data structure. Note that apart from the size of the pool that is in Atoms, all other
data structures refer only to shares.

Finally, let's consider what happens when new Atoms are created and added to the pool due to inflation. Let's assume that 
the inflation rate is 10 percent and that it is applied to the current state of the pool. This means that 5 Atoms are 
created and added to the pool and that each validator now proportionally increase it's Atom count. Let's analyse how this 
change is reflected in the data structures. First, the size of the pool is increased and is now 55 atoms. As a share of 
each validator in the pool hasn't changed, this means that the total number of shares stay the same (50) and that the 
amount of shares of each validator stays the same (correspondingly 25, 10, 10, 5). But the exchange rate has changed and 
each share is now worth 55/50 Atoms per share, so each validator has effectively increased amount of Atoms it has. 
So validators now have (correspondingly) 55/2, 55/5, 55/5 and 55/10 Atoms. 

The concepts of the pool and its shares is at the core of the accounting in the Staking module. It is used for managing 
the global pools (such as bonding and unbonding pool), but also for distribution of Atoms between validator and its 
delegators (we will explain this in section X).         

#### Delegator shares

A candidate is, depending on it's status, contributing Atoms to either the bonded or unbonded pool, and in return gets  
some amount of (global) pool shares. Note that not all those Atoms (and respective shares) are owned by the candidate 
as some Atoms could be delegated to a candidate. The mechanism for distribution of Atoms (and shares) between a 
candidate and it's delegators is based on a notion of delegator shares. More precisely, every candidate is issuing 
(local) delegator shares (`Candidate.IssuedDelegatorShares`) that represents some portion of global shares 
managed by the candidate (`Candidate.GlobalStakeShares`). The principle behind managing delegator shares is the same 
as described in [Section](#The pool and the share). We now illustrate it with an example.

Lets consider 4 validators p1, p2, p3 and p4, and assume that each validator has bonded 10 Atoms to a bonded pool. 
Furthermore, lets assume that we have issued initially 40 global shares, i.e., that `share-to-atom-ex-rate = 1 atom per 
share`. So we will `GlobalState.BondedPool = 40` and `GlobalState.BondedShares = 40` and in the Candidate data structure 
of each validator `Candidate.GlobalStakeShares = 10`. Furthermore, each validator issued 10 delegator 
shares which are initially owned by itself, i.e., `Candidate.IssuedDelegatorShares = 10`, where 
`delegator-share-to-global-share-ex-rate = 1 global share per delegator share`.
Now lets assume that a delegator d1 delegates 5 atoms to a validator p1 and consider what are the updates we need 
to make to the data structures. First, `GlobalState.BondedPool = 45` and `GlobalState.BondedShares = 45`. Then,
for validator p1 we have `Candidate.GlobalStakeShares = 15`, but we also need to issue also additional delegator shares, 
i.e., `Candidate.IssuedDelegatorShares = 15` as the delegator d1 now owns 5 delegator shares of validator p1, where
each delegator share is worth 1 global shares, i.e, 1 Atom. Lets see now what happens after 5 new Atoms are created due
to inflation. In that case, we only need to update `GlobalState.BondedPool` which is now equal to 50 Atoms as created
Atoms are added to the bonded pool. Note that the amount of global and delegator shares stay the same but they are now 
worth more as share-to-atom-ex-rate is now worth 50/45 Atoms per share. Therefore, a delegator d1 now owns 

`delegatorCoins = 10 (delegator shares) * 1 (delegator-share-to-global-share-ex-rate) * 50/45 (share-to-atom-ex-rate) = 100/9 Atoms`  
   





