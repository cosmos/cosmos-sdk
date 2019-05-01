# Validator FAQ

::: warning Disclaimer
This is work in progress. Mechanisms and values are susceptible to change.
:::

## General Concepts

### What is a validator?

The [Cosmos Hub](../what-is-gaia.md) is based on [Tendermint](https://tendermint.com/docs/introduction/what-is-tendermint.html), which relies on a set of validators to secure the network. The role of validators is to run a full-node and participate in consensus by broadcasting votes which contain cryptographic signatures signed by their private key. Validators commit new blocks in the blockchain and receive revenue in exchange for their work. They must also participate in governance by voting on proposals. Validators are weighted according to their total stake.

### What is 'staking'?

The Cosmos Hub is a public Proof-Of-Stake (PoS) blockchain, meaning that the weight of validators is determined by the amount of staking tokens (Atoms) bonded as collateral. These Atoms can be self-delegated directly by the validator or delegated to them by other Atom holders.

Any user in the system can declare their intention to become a validator by sending a `create-validator` transaction. From there, they become validator candidates.

The weight (i.e. voting power) of a validator determines whether or not they are an active validator. Initially, only the top 100 validators with the most voting power will be active validators. 

### What is a full-node?

A full-node is a program that fully validates transactions and blocks of a blockchain. It is distinct from a light-node that only processes block headers and a small subset of transactions. Running a full-node requires more resources than a light-node but is necessary in order to be a validator. In practice, running a full-node only implies running a non-compromised and up-to-date version of the software with low network latency and without downtime.

Of course, it is possible and encouraged for users to run full-nodes even if they do not plan to be validators.

### What is a delegator?

Delegators are Atom holders who cannot, or do not want to run a validator themselves. Atom holders can delegate Atoms to a validator and obtain a part of their revenue in exchange (for more detail on how revenue is distributed, see [**What is the incentive to stake?**](#what-is-the-incentive-to-stake?) and [**What are validators commission?**](#what-are-validators-commission?) sections below).

Because they share revenue with their validators, delegators also share risks. Should a validator misbehave, each of their delegators will be partially slashed in proportion to their delegated stake. This is why delegators should perform due diligence on validators before delegating, as well as spreading their stake over multiple validators.

Delegators play a critical role in the system, as they are responsible for choosing validators. Being a delegator is not a passive role: Delegators should actively monitor the actions of their validators and participate in governance. For more, read the [delegator's faq](https://cosmos.network/resources/delegators).

## Becoming a Validator

### How to become a validator?

Any participant in the network can signal that they want to become a validator by sending a `create-validator` transaction, where they must fill out the following parameters:

* **Validator's `PubKey`:** The private key associated with this Tendermint `PubKey` is used to sign _prevotes_ and _precommits_. 
* **Validator's Address:** Application level address. This is the address used to identify your validator publicly. The private key associated with this address is used to delegate, unbond, claim rewards, and participate in governance.
* **Validator's name (moniker)**
* **Validator's website (Optional)**
* **Validator's description (Optional)**
* **Initial commission rate**: The commission rate on block rewards and fees charged to delegators.
* **Maximum commission:** The maximum commission rate which this validator  can charge. This parameter cannot be changed after `create-validator` is processed.
* **Commission max change rate:** The maximum daily increase of the validator  commission. This parameter cannot be changed after `create-validator` is processed.
* **Minimum self-delegation:** Minimum amount of Atoms the validator needs to have bonded at all time. If the validator's self-delegated stake falls below this limit, their entire staking pool will unbond.

Once a validator is created, Atom holders can delegate atoms to them, effectively adding stake to their pool. The total stake of an address is the combination of Atoms bonded by delegators and Atoms self-bonded by the entity which designated themselves.

Out of all validator candidates that signaled themselves, the 100 with the most total stake are the ones who are designated as validators. They become **validators** If a validator's total stake falls below the top 100 then that validator loses their validator privileges: they don't participate in consensus and generate rewards any more. Over time, the maximum number of validators will increase, according to the following schedule (*Note: this schedule can be changed by governance*):

* **Year 0:** 100
* **Year 1:** 113
* **Year 2:** 127
* **Year 3:** 144
* **Year 4:** 163
* **Year 5:** 184
* **Year 6:** 208
* **Year 7:** 235
* **Year 8:** 265
* **Year 9:** 300
* **Year 10:** 300

## Testnet

### How can I join the testnet?

The Testnet is a great environment to test your validator setup before launch.

We view testnet participation as a great way to signal to the community that you are ready and able to operate a validator. You can find all relevant information about the testnet [here](../join-testnet.md) and [here](https://github.com/cosmos/testnets).

### What are the different types of keys?

In short, there are two types of keys:

- **Tendermint Key**: This is a unique key used to sign consensus votes. 
    + It is associated with a public key `cosmosvalconspub` (Get this value with `gaiad tendermint show-validator`)
    + It is generated when the node is created with gaiad init.
    
- **Application key**: This key is created from `gaiacli` and used to sign transactions. Application keys are associated with a public key prefixed by `cosmospub` and an address prefixed by `cosmos`. Both are derived from account keys generated by `gaiacli keys add`.

Note: A validator's operator key is directly tied to an application key, but
uses reserved prefixes solely for this purpose: `cosmosvaloper` and `cosmosvaloperpub`

### What are the different states a validator can be in?

After a validator is created with a `create-validator` transaction, they can be in three states:

- `in validator set`: Validator is in the active set and participates in consensus. Validator is earning rewards and can be slashed for misbehaviour.
- `jailed`: Validator misbehaved and is in jail, i.e. outisde of the validator set. If the jailing is due to being offline for too long, the validator can send an `unjail` transaction in order to re-enter the validator set. If the jailing is due to double signing, the validator cannot unjail.  
- `unbonded`: Validator is not in the active set, and therefore not signing blocs. Validator cannot be slashed, and does not earn any reward. It is still possible to delegate Atoms to this validator. Un-delegating from an `unbonded` validator is immediate.


### What is 'self-delegation'? How can I increase my 'self-delegation'?

Self-delegation is delegation from a validator to themselves. This amount can be increases by sending a `delegate` transaction from your validator's `application` application key.

### Is there a minimum amount of Atoms that must be delegated to be an active (=bonded) validator?

The minimum is `1 atom`. 

### How will delegators choose their validators?

Delegators are free to choose validators according to their own subjective criteria. This said, criteria anticipated to be important include:

* **Amount of self-delegated Atoms:** Number of Atoms a validator self-delegated to themselves. A validator with a higher amount of self-delegated Atoms has more skin in the game, making them more liable for their actions.
* **Amount of delegated Atoms:** Total number of Atoms delegated to a validator. A high voting power shows that the community trusts this validator, but it also means that this validator is a bigger target for hackers. Bigger validators also decrease the decentralisation of the network.
* **Commission rate:** Commission applied on revenue by validators before it is distributed to their delegators. 
* **Track record:** Delegators will likely look at the track record of the validators they plan to delegate to. This includes seniority, past votes on proposals, historical average uptime and how often the node was compromised.

Apart from these criteria, there will be a possibility for validators to signal a website address to complete their resume. Validators will need to build reputation one way or another to attract delegators. For example, it would be a good practice for validators to have their setup audited by third parties. Note though, that the Tendermint team will not approve or conduct any audit themselves. For more on due diligence, see [this blog post](https://medium.com/@interchain_io/3d0faf10ce6f)

## Responsibilities

### Do validators need to be publicly identified?

No, they do not. Each delegator will value validators based on their own criteria. Validators will be able to register a website address when they nominate themselves so that they can advertise their operation as they see fit. Some delegators may prefer a website that clearly displays the team operating the validator and their resume, while others might prefer anonymous validators with positive track records.

### What are the responsibilities of a validator?

Validators have two main responsibilities:

* **Be able to constantly run a correct version of the software:**Validators need to make sure that their servers are always online and their private keys are not compromised.
* **Actively participate in governance:** Validators are required to vote on every proposal.

Additionally, validators are expected to be active members of the community. They should always be up-to-date with the current state of the ecosystem so that they can easily adapt to any change.

### What does 'participate in governance' entail?

Validators and delegators on the Cosmos Hub can vote on proposals to change operational parameters (such as the block gas limit), coordinate upgrades, or make a decision on any given matter. 

Validators play a special role in the governance system. Being the pillars of the system, they are required to vote on every proposal. It is especially important since delegators who do not vote will inherit the vote of their validator.

### What does staking imply?

Staking Atoms can be thought of as a safety deposit on validation activities. When a validator or a delegator wants to retrieve part or all of their deposit, they send an `unbonding` transaction. Then, Atoms undergo a **3 weeks unbonding period** during which they are liable to being slashed for potential misbehaviors committed by the validator before the unbonding process started.

Validators, and by association delegators, receive block rewards, fees, and have the right to participate in governance. If a validator misbehaves, a certain portion of their total stake is slashed. This means that every delegator that bonded Atoms to this validator gets penalized in proportion to their bonded stake. Delegators are therefore incentivized to delegate to validators that they anticipate will function safely.

### Can a validator run away with their delegators' Atoms?

By delegating to a validator, a user delegates voting power. The more voting power a validator have, the more weight they have in the consensus and governance processes. This does not mean that the validator has custody of their delegators' Atoms. **By no means can a validator run away with its delegator's funds**.

Even though delegated funds cannot be stolen by their validators, delegators are still liable if their validators misbehave. 

### How often will a validator be chosen to propose the next block? Does it go up with the quantity of bonded Atoms?

The validator that is selected to propose the next block is called proposer. Each proposer is selected deterministically, and the frequency of being chosen is proportional to the voting power (i.e. amount of bonded Atoms) of the validator. For example, if the total bonded stake across all validators is 100 Atoms and a validator's total stake is 10 Atoms, then this validator will proposer ~10% of the blocks.

### Will validators of the Cosmos Hub ever be required to validate other zones in the Cosmos ecosystem?

Yes, they will. If governance decides so, validators of the Cosmos hub may be required to validate additional zones in the Cosmos ecosystem. 

## Incentives

### What is the incentive to stake?

Each member of a validator's staking pool earns different types of revenue:

* **Block rewards:** Native tokens of applications run by validators (e.g. Atoms on the Cosmos Hub) are inflated to produce block provisions. These provisions exist to incentivize Atom holders to bond their stake, as non-bonded Atom will be diluted over time.
* **Transaction fees:** The Cosmos Hub maintains a whitelist of token that are accepted as fee payment. The initial fee token is the `atom`.

This total revenue is divided among validators' staking pools according to each validator's weight. Then, within each validator's staking pool the revenue is divided among delegators in proportion to each delegator's stake. A commission on delegators' revenue is applied by the validator before it is distributed.

### What is the incentive to run a validator ?

Validators earn proportionally more revenue than their delegators because of commissions.

Validators also play a major role in governance. If a delegator does not vote, they inherit the vote from their validator. This gives validators a major responsibility in the ecosystem.

### What are validators commission?

Revenue received by a validator's pool is split between the validator and their delegators. The validator can apply a commission on the part of the revenue that goes to their delegators. This commission is set as a percentage. Each validator is free to set their initial commission, maximum daily commission change rate and maximum commission. The Cosmos Hub enforces the parameter that each validator sets. Only the commission rate can change after the validator is created. 

### How are block rewards distributed?

Block rewards are distributed proportionally to all validators relative to their voting power. This means that even though each validator gains atoms with each reward, all validators will maintain equal weight over time.

Let us take an example where we have 10 validators with equal voting power and a commission rate of 1%. Let us also assume that the reward for a block is 1000 Atoms and that each validator has 20% of self-bonded Atoms. These tokens do not go directly to the proposer. Instead, they are evenly spread among validators. So now each validator's pool has 100 Atoms. These 100 Atoms will be distributed according to each participant's stake:

* Commission: `100*80%*1% = 0.8 Atoms`
* Validator gets: `100\*20% + Commission = 20.8 Atoms`
* All delegators get: `100\*80% - Commission = 79.2 Atoms`

Then, each delegator can claim their part of the 79.2 Atoms in proportion to their stake in the validator's staking pool. 

### How are fees distributed?

Fees are similarly distributed with the exception that the block proposer can get a bonus on the fees of the block they propose if they include more than the strict minimum of required precommits.

When a validator is selected to propose the next block, they must include at least 2/3 precommits of the previous block. However, there is an incentive to include more than 2/3 precommits in the form of a bonus. The bonus is linear: it ranges from 1% if the proposer includes 2/3rd precommits (minimum for the block to be valid) to 5% if the proposer includes 100% precommits. Of course the proposer should not wait too long or other validators may timeout and move on to the next proposer. As such, validators have to find a balance between wait-time to get the most signatures and risk of losing out on proposing the next block. This mechanism aims to incentivize non-empty block proposals, better networking between validators as well as to mitigate censorship.

Let's take a concrete example to illustrate the aforementioned concept. In this example, there are 10 validators with equal stake. Each of them applies a 1% commission rate and has 20% of self-delegated Atoms. Now comes a successful block that collects a total of 1025.51020408 Atoms in fees.

First, a 2% tax is applied. The corresponding Atoms go to the reserve pool. Reserve pool's funds can be allocated through governance to fund bounties and upgrades.

* `2% * 1025.51020408 = 20.51020408` Atoms go to the reserve pool.

1005 Atoms now remain. Let's assume that the proposer included 100% of the signatures in its block. It thus obtains the full bonus of 5%.

We have to solve this simple equation to find the reward R for each validator:

`9*R + R + R*5% = 1005 ⇔ R = 1005/10.05 = 100`

* For the proposer validator:
  * The pool obtains `R + R * 5%`: 105 Atoms
  * Commission: `105 * 80% * 1%` = 0.84 Atoms
  * Validator's reward: `105 * 20% + Commission` = 21.84 Atoms
  * Delegators' rewards: `105 * 80% - Commission` = 83.16 Atoms (each delegator will be able to claim its portion of these rewards in proportion to their stake)
* For each non-proposer validator:
  * The pool obtains R: 100 Atoms
  * Commission: `100 * 80% * 1%` = 0.8 Atoms
  * Validator's reward: `100 * 20% + Commission` = 20.8 Atoms
  * Delegators' rewards: `100 * 80% - Commission` = 79.2 Atoms (each delegator will be able to claim their portion of these rewards in proportion to their stake)

### What are the slashing conditions?

If a validator misbehaves, their delegated stake will be partially slashed. There are currently two faults that can result in slashing of funds for a validator and their delegators:

* **Double signing:** If someone reports on chain A that a validator signed two blocks at the same height on chain A and chain B, and if chain A and chain B share a common ancestor, then this validator will get slashed by 5% on chain A.
* **Downtime:** If a validator misses more than 95% of the last 10.000 blocks, they will get slashed by 0.01%. 

### Do validators need to self-delegate Atoms?

Yes, they do need to self-delegate at least `1 atom`. Even though there is no obligation for validators to self-delegate more than `1 atom`, delegators should want their validator to have more self-delegated Atoms in their staking pool. In other words, validators should have skin in the game.

In order for delegators to have some guarantee about how much skin-in-the-game their validator has, the latter can signal a minimum amount of self-delegated Atoms. If a validator's self-delegation goes below the limit that it predefined, this validator and all of its delegators will unbond.

### How to prevent concentration of stake in the hands of a few top validators?

For now the community is expected to behave in a smart and self-preserving way. When a mining pool in Bitcoin gets too much mining power the community usually stops contributing to that pool. The Cosmos Hub will rely on the same effect initially. Other mechanisms are in place to smoothen this process as much as possible:

* **Penalty-free re-delegation:** This is to allow delegators to easily switch from one validator to another, in order to reduce validator stickiness.
* **UI warning:** Users will be warned by Cosmos Voyager if they want to delegate to a validator that already has a significant amount of staking power.

## Technical Requirements

### What are hardware requirements?

Validators should expect to provision one or more data center locations with redundant power, networking, firewalls, HSMs and servers.

We expect that a modest level of hardware specifications will be needed initially and that they might rise as network use increases. Participating in the testnet is the best way to learn more.

### What are software requirements?

In addition to running a Cosmos Hub node, validators should develop monitoring, alerting and management solutions.

### What are bandwidth requirements?

The Cosmos network has the capacity for very high throughput relative to chains like Ethereum or Bitcoin.

We recommend that the data center nodes only connect to trusted full-nodes in the cloud or other validators that know each other socially. This relieves the data center node from the burden of mitigating denial-of-service attacks.

Ultimately, as the network becomes more heavily used, multigigabyte per day bandwidth is very realistic.

### What does running a validator imply in terms of logistics?

A successful validator operation will require the efforts of multiple highly skilled individuals and continuous operational attention. This will be considerably more involved than running a bitcoin miner for instance.

### How to handle key management?

Validators should expect to run an HSM that supports ed25519 keys. Here are potential options:

* YubiHSM 2
* Ledger Nano S
* Ledger BOLOS SGX enclave
* Thales nShield support

The Tendermint team does not recommend one solution above the other. The community is encouraged to bolster the effort to improve HSMs and the security of key management.

### What can validators expect in terms of operations?

Running effective operation is the key to avoiding unexpectedly unbonding or being slashed. This includes being able to respond to attacks, outages, as well as to maintain security and isolation in your data center.

### What are the maintenance requirements?

Validators should expect to perform regular software updates to accommodate upgrades and bug fixes. There will inevitably be issues with the network early in its bootstrapping phase that will require substantial vigilance.

### How can validators protect themselves from denial-of-service attacks?

Denial-of-service attacks occur when an attacker sends a flood of internet traffic to an IP address to prevent the server at the IP address from connecting to the internet.

An attacker scans the network, tries to learn the IP address of various validator nodes and disconnect them from communication by flooding them with traffic.

One recommended way to mitigate these risks is for validators to carefully structure their network topology in a so-called sentry node architecture.

Validator nodes should only connect to full-nodes they trust because they operate them themselves or are run by other validators they know socially. A validator node will typically run in a data center. Most data centers provide direct links the networks of major cloud providers. The validator can use those links to connect to sentry nodes in the cloud. This shifts the burden of denial-of-service from the validator's node directly to its sentry nodes, and may require new sentry nodes be spun up or activated to mitigate attacks on existing ones.

Sentry nodes can be quickly spun up or change their IP addresses. Because the links to the sentry nodes are in private IP space, an internet based attacked cannot disturb them directly. This will ensure validator block proposals and votes always make it to the rest of the network.

It is expected that good operating procedures on that part of validators will completely mitigate these threats.

For more on sentry node architecture, see [this](https://forum.cosmos.network/t/sentry-node-architecture-overview/454).
