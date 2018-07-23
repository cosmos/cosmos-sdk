# Delegator FAQ

## What is a delegator?

People that cannot, or do not want to run [validator](/validators/overview.md) operations, can still participate in the staking process as delegators. Indeed, validators are not chosen based on their own stake but based on their total stake, which is the sum of their own stake and of the stake that is delegated to them. This is an important property, as it makes delegators a safeguard against validators that exhibit bad behavior. If a validator misbehaves, its delegators will move their Atoms away from it, thereby reducing its stake. Eventually, if a validator's stake falls under the top 100 addresses with highest stake, it will exit the validator set.

Delegators share the revenue of their validators, but they also share the risks. In terms of revenue, validators and delegators differ in that validators can apply a commission on the revenue that goes to their delegator before it is distributed. This commission is known to delegators beforehand and can only change according to predefined constraints (see section below). In terms of risk, delegators' Atoms can be slashed if their validator misbehaves. For more, see Risks section.

To become delegators, Atom holders need to send a "Bond transaction" from [Cosmos Voyager](/getting-started/voyager.md) where they specify how many Atoms they want to bond and to which validator. A list of validator candidates will be displayed in Cosmos Voyager. Later, if a delegator wants to unbond part or all of its stake, it needs to send an "Unbond transaction". From there, the delegator will have to wait 3 weeks to retrieve its Atoms.

## Choosing a validator

In order to choose their validators, delegators have access to a range of information directly in Cosmos Voyager.

* Validator's name: Name that was chosen by the validator candidate when it declared candidacy.
* Validator's description: Description that was provided by the validator candidate when it declared candidacy.
* Validator's website: Link to the validator's website.
* Initial commission rate: The commission rate on revenue charged to any delegators (see below for more detail).
* Commission change rate: The maximum daily increase of the validator's commission
* Maximum commission: The maximum commission rate which this validator candidate can charge.
* Minimum self-bond amount: Minimum amount of Atoms the validator candidate need to have bonded at all time. If the validator's self-bonded stake falls below this limit, its entire staking pool (i.e. all its delegators) will unbond. This parameter exists as a safeguard for delegators. Indeed, when a validator misbehaves, part of its total stake gets slashed. This included the validator's own stake as well as its delegators' stake. Thus, a validator with a high amount of self-bonded Atoms has more skin-in-the-game than a validator with a low amount. The minimum self-bond amount parameter guarantees to delegators that a validator will never fall below a certain amount of self-bonded stake, thereby ensuring a minimum level of skin-in-the-game.

## Directives of delegators

Being a delegator is not a passive task. Here are the main directives of a delegator:

* Perform careful due diligence on validators before delegating. If a validator misbehaves, part of its total stake, which includes the stake of its delegators, can be slashed. Delegators should therefore carefully select validators they think will behave correctly.
* Actively monitor their validator after having delegated. Delegators should ensure that the validators they're delegating to behaves correctly, meaning that they have good uptime, do not get hacked and participate in governance. They should also monitor the commission rate that is applied. If a delegator is not satisfied with its validator, it can unbond or switch to another validator.
* Participate in governance. Delegators can and are expected to actively participate in governance. A delegator's voting power is proportional to the size of its stake. If a delegator does not vote, it will inherit the vote of its validator. Delegators therefore act as a counterbalance to their validators.

## Revenue

Validators and delegators earn revenue in exchange for their services. This revenue is given in three forms:

* Block provisions (Atoms): They are paid in newly created Atoms. Block provisions exist to incentivize Atom holders to stake. The yearly inflation rate fluctuates around a target of 2/3 bonded stake. If the total bonded stake is less than 2/3 of the total Atom supply, inflation increases until it reaches 20%. If the total bonded stake is more than 2/3 of the Atom supply, inflation decreases until it reaches 7%. This means that if total bonded stake stays less than 2/3 of the total Atom supply for a prolonged period of time, unbonded Atom holders can expect their Atom value to deflate by 20% per year.
* Block rewards ([Photons](https://blog.cosmos.network/cosmos-fee-token-introducing-the-photon-8a62b2f51aa): They are paid in Photons. Initial distribution of Photons will take the form of a hard spoon of the Ethereum chain. Atom holders will vote on the parameter of this hard spoon, like the date of the snapshot or the initial distribution. Additionally, bonded Atom holders will receive newly created Photons as block rewards. Photons will be distributed at a fixed rate in proportion to each bonded Atom holder's stake. This rate will be decided via governance.
* Transaction fees (various tokens): Each transfer on the Cosmos Hub comes with transactions fees. These fees can be paid in any currency that is whitelisted by the Hub's governance. Fees are distributed to bonded Atom holders in proportion to their stake. The first whitelisted tokens at launch are Atoms and Photons.

## Validator's commission

Each validator's staking pool receives revenue in proportion to its total stake. However, before this revenue is distributed to delegators inside the staking pool, the validator can apply a commission. In other words, delegators have to pay a commission to their validators on the revenue they earn. Let us look at a concrete example:

We consider a validator whose stake (i.e. self-bonded stake + delegated stake) is 10% of the total stake of all validators. This validator has 20% self-bonded stake and applies a commission of 10%. Now let us consider a block with the following revenue:

* 990 Atoms in block provisions
* 10 Photons in block reward
* 10 Atoms and 90 Photons in transaction fees.

This amounts to a total of 1000 Atoms and 100 Photons to be distributed among all staking pools.

Our validator's staking pool represents 10% of the total stake, which means the pool obtains 100 Atoms and 10 Photons. Now let us look at the internal distribution of revenue:

* Commission = `10% * 80% * 100` Atoms + `10% * 80% * 10` Photons = 8 Atoms + 0.8 Photons
* Validator's revenue = `20% * 100` Atoms + `20% * 10` Photons + Commission = 28 Atoms + 2.8 Photons
* Delegators' total revenue = `80% * 100` Atoms + `20% * 10` Photons - Commission = 72 Atoms + 7.2 Photons

Then, each delegator in the staking pool can claim its portion of the delegators' total revenue.

## Risks

Staking Atoms is not free of risk. First, staked Atoms are locked up, and retrieving them requires a 3 week waiting period called unbonding period. Additionally, if a validator misbehaves, a portion of its total stake can be slashed (i.e. destroyed). This includes the stake of their delegators.

There are 3 main slashing conditions:

* Double signing: If someone reports on chain A that a validator signed two blocks at the same height on chain A and chain B, this validator will get slashed on chain A
* Unavailability: If a validator's signature has not been included in the last X blocks, the validator will get slashed by a marginal amount proportional to X. If X is above a certain limit Y, then the validator will get unbonded
* Non-voting: If a validator did not vote on a proposal and once the fault is reported by a someone, its stake will receive a minor slash.

This is why Atom holders should perform careful due diligence on validators before delegating. It is also important that delegators actively monitor the activity of their validators. If a validator behaves suspiciously or is too often offline, delegators can choose to unbond from it or switch to another validator. Delegators can also mitigate risk by distributing their stake across multiple validators.
