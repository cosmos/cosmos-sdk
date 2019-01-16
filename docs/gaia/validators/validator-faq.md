# Validator FAQ

::: warning Disclaimer
This is work in progress. Mechanisms and values are susceptible to change.
:::

## General Concepts

### What is a validator?

The [Cosmos Hub](/introduction/cosmos-hub.md) is based on [Tendermint](/introduction/tendermint.md), which relies on a set of [validators](/validators/overview.md) to secure the network. The role of validators is to run a full-node and participate in consensus by broadcasting votes which contain cryptographic signatures signed by their private key. Validators commit new blocks in the blockchain and receive revenue in exchange for their work. They must also participate in governance by voting on proposals. Validators are weighted according to their total stake.

### What is 'staking'?

The Cosmos Hub is a public Proof-Of-Stake (PoS) blockchain, meaning that validator's weight is determined by the amount of staking tokens (Atoms) bonded as collateral. These Atoms can be staked directly by the validator or delegated to them by Atom holders.

Any user in the system can declare its intention to become a validator by sending a `create-validator` transaction. From there, they become validators.

The weight (i.e. total stake) of a validator determines wether or not it is an active validator, and also how frequently this node will have to propose a block and how much revenue it will obtain. Initially, only the top 100 validators with the most weight will be active validators. If validators double sign, are frequently offline or do not participate in governance, their staked Atoms (including Atoms of users that delegated to them) can be destroyed, or 'slashed'.

### What is a full-node?

A full-node is a program that fully validates transactions and blocks of a blockchain. It is distinct from a light-node that only processes block headers and a small subset of transactions. Running a full-node requires more resources than a light-node but is necessary in order to be a validator. In practice, running a full-node only implies running a non-compromised and up-to-date version of the software with low network latency and without downtime.

Of course, it is possible and encouraged for any user to run full-nodes even if they do not plan to be validators.

### What is a delegator?

Delegators are Atom holders who cannot, or do not want to run validator operations themselves. Through [Cosmos Voyager](/getting-started/voyager.md), a user can delegate Atoms to a validator and obtain a part of its revenue in exchange (for more detail on how revenue is distributed, see **What is the incentive to stake?** and **What is a validator's commission?** sections below).

Because they share revenue with their validators, delegators also share responsibility. Should a validator misbehave, each of its delegators will be partially slashed in proportion to their stake. This is why delegators should perform due diligence on validators before delegating, as well as spreading their stake over multiple validators.

Delegators play a critical role in the system, as they are responsible for choosing validators. Being a delegator is not a passive role: Delegators should actively monitor the actions of their validators and participate in governance.

## Becoming a Validator

### How to become a validator?

Any participant in the network can signal that they want to become a validator by sending a `create-validator` transaction, where they must fill out the following parameters:

* Validator's PubKey:  The private key associated with PubKey is used to sign _prevotes_ and _precommits_. This way, validators can have different accounts for validating and holding liquid funds.
* Validator's Address: Application level address. This is the address used to identify your validator publicly. The private key associated with this address is used to bond, unbond, claim rewards, and participate in governance (in MVP only).
* Validator's name (moniker)
* Validator's website (Optional)
* Validator's description (Optional)
* Initial commission rate: The commission rate on block provisions, block rewards and fees charged to delegators
* Maximum commission: The maximum commission rate which this validator  can charge
* Commission change rate: The maximum daily increase of the validator  commission
* Minimum self-bond amount: Minimum amount of Atoms the validator need to have bonded at all time. If the validator's self-bonded stake falls below this limit, its entire staking pool will unbond.
* Initial self-bond amount: Initial amount of Atoms the validator wants to self-bond

Once a validator is created, Atom holders can delegate atoms to it, effectively adding stake to this pool. The total stake of an address is the combination of Atoms bonded by delegators and Atoms self-bonded by the entity which designated itself.

Out of all validators that signaled themselves, the 100 with the most stake are the ones who are designated as validators. They become **bonded validators** If a validator's total stake falls below the top 100 then that validator loses its validator privileges, it enters **unbonding mode** and, eventually, becomes **unbonded** . Over time, the maximum number of validators will increase, according to a predefined schedule:

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

We view testnet participation as a great way to signal to the community that you are ready and able to operate a validator. You can find all relevant information about the testnet [here](https://github.com/cosmos/cosmos-sdk/tree/develop/cmd/gaia/testnets) and [here](https://github.com/cosmos/testnets).

### What are the different types of keys?

In short, there are two types of keys:

* **Tendermint Key**: This is a unique key used to sign block hashes. It is associated with a public key `cosmosvalconspub`.
  * Generated when the node is created with gaiad init.
  * Get this value with `gaiad tendermint show-validator`
    e.g. `cosmosvalconspub1zcjduc3qcyj09qc03elte23zwshdx92jm6ce88fgc90rtqhjx8v0608qh5ssp0w94c`

* **Application keys**: These keys are created from the application and used to sign transactions. As a validator, you will probably use one key to sign staking-related transactions, and another key to sign governance-related transactions. Application keys are associated with a public key `cosmospub` and an address `cosmos`. Both are derived from account keys generated by `gaiacli keys add`.
  * Note: A validator's operator key is directly tied to an application key, but
  uses reserved prefixes solely for this purpose: `cosmosvaloper` and `cosmosvaloperpub`

### What are the different states a validator can be in?

After a validator is created with a `create-validator` transaction, it can be in three states:

- `bonded`: Validator is in the active set and participates in consensus. Validator is earning rewards and can be slashed for misbehaviour.
- `unbonding`: Validator is not in the active set and does not participate in consensus. Validator is not earning rewards, but can still be slashed for misbehaviour. This is a transition state from `bonded` to `unbonded`. If validator does not send a `rebond` transaction while in `unbonding` mode, it will take three weeks for the state transition to complete. 
- `unbonded`: Validator is not in the active set, and therefore not signing blocs. Validator cannot be slashed, and does not earn any reward. It is still possible to delegate Atoms to this validator. Un-delegating from an `unbonded` validator is immediate.

Delegators have the same state as their validator. 

*Note that delegation are not necessarily bonded. Atoms can be delegated and bonded, delegated and unbonding, delegated and unbonded, or liquid*


### What is 'self-bond'? How can I increase my 'self-bond'?

### Is there a faucet?

If you want to obtain coins for the testnet, you can do so by using [this faucet](https://gaia.faucetcosmos.network/)

### Is there a minimum amount of Atoms that must be staked to be an active (=bonded) validator?

There is no minimum. The top 100 validators with the highest total stake (where total stake = self-bonded stake + delegators stake) are the active validators.

### How will delegators choose their validators?

Delegators are free to choose validators according to their own subjective criteria. This said, criteria anticipated to be important include:

* **Amount of self-bonded Atoms:** Number of Atoms a validator self-bonded to its staking pool. A validator with higher amount of self-bonded Atoms has more skin in the game, making it more liable for its actions.
* **Amount of delegated Atoms:** Total number of Atoms delegated to a validator. A high stake shows that the community trusts this validator, but it also means that this validator is a bigger target for hackers. Indeed, hackers are incentivized to hack bigger validators as they receive a reward proportionate to the stake of the validator they can prove to have compromised. Validators are expected to become less and less attractive as their amount of delegated Atoms grows.
* **Commission rate:** Commission applied on revenue by validators before it is distributed to their delegators
* **Track record:** Delegators will likely look at the track record of the validators they plan to delegate to. This includes seniority, past votes on proposals, historical average uptime and how often the node was compromised.

Apart from these criteria that will be displayed in Cosmos Voyager, there will be a possibility for validators to signal a website address to complete their resume. Validators will need to build reputation one way or another to attract delegators. For example, it would be a good practice for validators to have their setup audited by third parties. Note though, that the Tendermint team will not approve or conduct any audit itself. For more on due diligence, see [this blog post](https://medium.com/@interchain_io/3d0faf10ce6f)

## Responsibilites

### Do validators need to be publicly identified?

No, they do not. Each delegator will value validators based on their own criteria. Validators will be able (and are advised) to register a website address when they nominate themselves so that they can advertise their operation as they see fit. Some delegators may prefer a website that clearly displays the team running the validator and their resume, while others might prefer anonymous validators with positive track records. Most likely both identified and anonymous validators will coexist in the validator set.

### What are the responsiblities of a validator?

Validators have two main responsibilities:

* **Be able to constantly run a correct version of the software:** validators need to make sure that their servers are always online and their private keys are not compromised.
* **Actively participate in governance:** validators are required to vote on every proposal.

Additionally, validators are expected to be active members of the community. They should always be up-to-date with the current state of the ecosystem so that they can easily adapt to any change.

### What does 'participate in governance' entail?

Validators and delegators on the Cosmos Hub can vote on proposals to change operational parameters (such as the block gas limit), coordinate upgrades, as well as vote on amendments to the human-readable constitution that govern the Cosmos Hub.

Validators play a special role in the governance system. Being the pillars of the system, they are required to vote on every proposal. It is especially important since delegators who do not vote will inherit the vote of their validator. Each time a validator does not vote on a proposal, it will get slashed by a minimal amount.

### What does staking imply?

Staking Atoms can be thought of as a safety deposit on validation activities. When a validator or a delegator wants to retrieve part or all of their deposit, they send an unbonding transaction. Then, Atoms undergo a _three weeks unbonding period_ during which they are liable to being slashed for potential misbehaviors committed by the validator before the unbonding process started.

Validators, and by association delegators, receive block provisions, block rewards, fee rewards, and the right to participate in governance. If a validator misbehaves, a certain portion of its total stake is slashed (the severity of the penalty depends on the type of misbehavior). This means that every user that bonded Atoms to this validator gets penalized in proportion to its stake. Delegators are therefore incentivized to delegate to validators that they anticipate will function safely.

### Can a validator run away with its delegators' Atoms?

By delegating to a validator, a user delegates staking power. The more staking power a validator has, the more weight it has in the consensus and governance processes. This does not mean that the validator has custody of its delegators' Atoms. _By no means can a validator run away with its delegator's funds_.

Even though delegated funds cannot be stolen by their validators, delegators are still liable if their validators misbehave. In such case, each delegators' stake will be partially slashed in proportion to their relative stake.

### How often will a validator be chosen to propose the next block? Does it go up with the quantity of Atoms staked?

The validator that is selected to propose the next block is called proposer. Each proposer is selected deterministically, and the frequency of being chosen is equal to the relative total stake (where total stake = self-bonded stake + delegators stake) of the validator. For example, if the total bonded stake across all validators is 100 Atoms and a validator's total stake is 10 Atoms, then this validator will be chosen 10% of the time as the next proposer.

### Will validators of the Cosmos Hub ever be required to validate other zones in the Cosmos ecosystem?

Yes, they will. Initially, validators of the Cosmos hub will also validate the first public Ethermint zone. If governance decides so, validators of the Cosmos hub may be required to validate additional zones in the Cosmos ecosystem. As the case with the Ethermint Zone, for each additional zone compensation is to be provided in the form of block rewards and transaction fees.

## Incentives

### What is the incentive to stake?

Each member of a validator's staking pool earns different types of revenue:

* **Block provisions:** Native tokens of applications run by validators (e.g. Atoms on the Cosmos Hub) are inflated to produce block provisions. These provisions exist to incentivize Atom holders to bond their stake, as non-bonded Atom will be diluted over time.
* **Block rewards:** For the Ethermint zone, block rewards are paid in Photons. Initial distribution of Photons will be hard spooned from Ethereum. This means Photons will be emitted 1:1 to Ether.
* **Transaction fees:** The Cosmos Hub maintains a whitelist of token that are accepted as fee payment.

This total revenue is divided among validators' staking pools according to each validator's weight. Then, within each validator's staking pool the revenue is divided among delegators in proportion to each delegator's stake. Note that a commission on delegators' revenue is applied by the validator before it is distributed.

### What is the incentive to run a validator ?

Validators earn proportionally more revenue than their delegators because of commissions.

Validators also play a major role in governance. If a delegator does not vote, it inherits the vote from its validator. This gives validators a major responsibility in the ecosystem.

### What is a validator's commission?

Revenue received by a validator's pool is split between the validator and its delegators. The validator can apply a commission on the part of the revenue that goes to its delegators. This commission is set as a percentage. Each validator is free to set its initial commission, maximum daily commission change rate and maximum commission. The Cosmos Hub enforces the parameter that each validator sets. These parameters can only be defined when initially declaring candidacy, and may only be constrained further after being declared.

### How are block provisions distributed?

Block provisions are distributed proportionally to all validators relative to their total stake. This means that even though each validator gains atoms with each provision, all validators will still maintain equal weight.

Let us take an example where we have 10 validators with equal staking power and a commission rate of 1%. Let us also assume that the provision for a block is 1000 Atoms and that each validator has 20% of self-bonded Atoms. These tokens do not go directly to the proposer. Instead, they are evenly spread among validators. So now each validator's pool has 100 Atoms. These 100 Atoms will be distributed according to each participant's stake:

* Commission: `100*80%*1% = 0.8 Atoms`
* Validator gets: `100\*20% + Commission = 20.8 Atoms`
* All delegators get: `100\*80% - Commission = 79.2 Atoms`

Then, each delegator can claim its part of the 79.2 Atoms in proportion to their stake in the validator's staking pool. Note that the validator's commission is not applied on block provisions. Note that block rewards (paid in Photons) are distributed according to the same mechanism.

### How are fees distributed?

Fees are similarly distributed with the exception that the block proposer can get a bonus on the fees of the block it proposes if it includes more than the strict minimum of required precommits.

When a validator is selected to propose the next block, it must include at least 2/3 precommits for the previous block in the form of validator signatures. However, there is an incentive to include more than 2/3 precommits in the form of a bonus. The bonus is linear: it ranges from 1% if the proposer includes 2/3rd precommits (minimum for the block to be valid) to 5% if the proposer includes 100% precommits. Of course the proposer should not wait too long or other validators may timeout and move on to the next proposer. As such, validators have to find a balance between wait-time to get the most signatures and risk of losing out on proposing the next block. This mechanism aims to incentivize non-empty block proposals, better networking between validators as well as to mitigate censorship.

Let's take a concrete example to illustrate the aforementioned concept. In this example, there are 10 validators with equal stake. Each of them applies a 1% commission and has 20% of self-bonded Atoms. Now comes a successful block that collects a total of 1025.51020408 Atoms in fees.

First, a 2% tax is applied. The corresponding Atoms go to the reserve pool. Reserve pool's funds can be allocated through governance to fund bounties and upgrades.

* `2% \* 1025.51020408 = 20.51020408` Atoms go to the reserve pool.

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
  * Delegators' rewards: `100 * 80% - Commission` = 79.2 Atoms (each delegator will be able to claim its portion of these rewards in proportion to their stake)

### What are the slashing conditions?

If a validator misbehaves, its bonded stake along with its delegators' stake and will be slashed. The severity of the punishment depends on the type of fault. There are 3 main faults that can result in slashing of funds for a validator and its delegators:

* **Double signing:** If someone reports on chain A that a validator signed two blocks at the same height on chain A and chain B, and if chain A and chain B share a common ancestor, then this validator will get slashed on chain A
* **Unavailability:** If a validator's signature has not been included in the last X blocks, the validator will get slashed by a marginal amount proportional to X. If X is above a certain limit Y, then the validator will get unbonded
* **Non-voting:** If a validator did not vote on a proposal, its stake will receive a minor slash.

Note that even if a validator does not intentionally misbehave, it can still be slashed if its node crashes, looses connectivity, gets DDOSed, or if its private key is compromised.

### Do validators need to self-bond Atoms?

No, they do not. A validators total stake is equal to the sum of its own self-bonded stake and of its delegated stake. This means that a validator can compensate its low amount of self-bonded stake by attracting more delegators. This is why reputation is very important for validators.

Even though there is no obligation for validators to self-bond Atoms, delegators should want their validator to have self-bonded Atoms in their staking pool. In other words, validators should have skin in the game.

In order for delegators to have some guarantee about how much skin-in-the-game their validator has, the latter can signal a minimum amount of self-bonded Atoms. If a validator's self-bond goes below the limit that it predefined, this validator and all of its delegators will unbond.

### How to prevent concentration of stake in the hands of a few top validators?

For now the community is expected to behave in a smart and self-preserving way. When a mining pool in Bitcoin gets too much mining power the community usually stops contributing to that pool. The Cosmos Hub will rely on the same effect initially. In the future, other mechanisms will be deployed to smoothen this process as much as possible:

* **Penalty-free re-delegation:** This is to allow delegators to easily switch from one validator to another, in order to reduce validator stickiness.
* **Hack bounty:** This is an incentive for the community to hack validators. There will be bounties proportionate to the size of the validator, so that a validator becomes a bigger target as its stake grows.
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
