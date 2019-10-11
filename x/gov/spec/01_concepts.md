# Concepts

*Disclaimer: This is work in progress. Mechanisms are susceptible to change.*

The governance process is divided in a few steps that are outlined below:

* **Proposal submission:** Proposal is submitted to the blockchain with a 
  deposit.
* **Vote:** Once deposit reaches a certain value (`MinDeposit`), proposal is 
  confirmed and vote opens. Bonded Atom holders can then send `TxGovVote` 
  transactions to vote on the proposal.
* If the proposal involves a software upgrade:
  * **Signal:** Validators start signaling that they are ready to switch to the
    new version.
  * **Switch:** Once more than 75% of validators have signaled that they are 
    ready to switch, their software automatically flips to the new version.

## Proposal submission

### Right to submit a proposal

Any Atom holder, whether bonded or unbonded, can submit proposals by sending a 
`TxGovProposal` transaction. Once a proposal is submitted, it is identified by 
its unique `proposalID`.

### Proposal types

In the initial version of the governance module, there are two types of 
proposal:
* `PlainTextProposal` All the proposals that do not involve a modification of 
  the source code go under this type. For example, an opinion poll would use a 
  proposal of type `PlainTextProposal`.
* `SoftwareUpgradeProposal`. If accepted, validators are expected to update 
  their software in accordance with the proposal. They must do so by following 
  a 2-steps process described in the [Software Upgrade](#software-upgrade) 
  section below. Software upgrade roadmap may be discussed and agreed on via 
  `PlainTextProposals`, but actual software upgrades must be performed via 
  `SoftwareUpgradeProposals`.

Other modules may expand upon the governance module by implementing their own
proposal types and handlers. These types are registered and processed through the
governance module (eg. `ParamChangeProposal`), which then execute the respective
module's proposal handler when a proposal passes. This custom handler may perform
arbitrary state changes.

## Deposit

To prevent spam, proposals must be submitted with a deposit in the coins defined in the `MinDeposit` param. The voting period will not start until the proposal's deposit equals `MinDeposit`.

When a proposal is submitted, it has to be accompanied by a deposit that must be strictly positive, but can be inferior to `MinDeposit`. The submitter doesn't need to pay for the entire deposit on their own. If a proposal's deposit is inferior to `MinDeposit`, other token holders can increase the proposal's deposit by sending a `Deposit` transaction. The deposit is kept in an escrow in the governance `ModuleAccount` until the proposal is finalized (passed or rejected).

Once the proposal's deposit reaches `MinDeposit`, it enters voting period. If proposal's deposit does not reach `MinDeposit` before `MaxDepositPeriod`, proposal closes and nobody can deposit on it anymore.

### Deposit refund and burn

When a the a proposal finalized, the coins from the deposit are either refunded or burned, according to the final tally of the proposal:

* If the proposal is approved or if it's rejected but _not_ vetoed, deposits will automatically be refunded to their respective depositor (transferred from the governance `ModuleAccount`).
* When the proposal is vetoed with a supermajority, deposits be burned from the governance `ModuleAccount`.

## Vote

### Participants

*Participants* are users that have the right to vote on proposals. On the 
Cosmos Hub, participants are bonded Atom holders. Unbonded Atom holders and 
other users do not get the right to participate in governance. However, they 
can submit and deposit on proposals.

Note that some *participants* can be forbidden to vote on a proposal under a 
certain validator if:
* *participant* bonded or unbonded Atoms to said validator after proposal 
  entered voting period.
* *participant* became validator after proposal entered voting period.

This does not prevent *participant* to vote with Atoms bonded to other 
validators. For example, if a *participant* bonded some Atoms to validator A 
before a proposal entered voting period and other Atoms to validator B after 
proposal entered voting period, only the vote under validator B will be 
forbidden.

### Voting period

Once a proposal reaches `MinDeposit`, it immediately enters `Voting period`. We
define `Voting period` as the interval between the moment the vote opens and
the moment the vote closes. `Voting period` should always be shorter than 
`Unbonding period` to prevent double voting. The initial value of 
`Voting period` is 2 weeks.

### Option set

The option set of a proposal refers to the set of choices a participant can 
choose from when casting its vote.

The initial option set includes the following options: 
- `Yes`
- `No`
- `NoWithVeto` 
- `Abstain` 

`NoWithVeto` counts as `No` but also adds a `Veto` vote. `Abstain` option 
allows voters to signal that they do not intend to vote in favor or against the
proposal but accept the result of the vote. 

*Note: from the UI, for urgent proposals we should maybe add a ‘Not Urgent’ 
option that casts a `NoWithVeto` vote.*

### Quorum 

Quorum is defined as the minimum percentage of voting power that needs to be 
casted on a proposal for the result to be valid. 

### Threshold

Threshold is defined as the minimum proportion of `Yes` votes (excluding 
`Abstain` votes) for the proposal to be accepted.

Initially, the threshold is set at 50% with a possibility to veto if more than
1/3rd of votes (excluding `Abstain` votes) are `NoWithVeto` votes. This means 
that proposals are accepted if the proportion of `Yes` votes (excluding 
`Abstain` votes) at the end of the voting period is superior to 50% and if the 
proportion of `NoWithVeto` votes is inferior to 1/3 (excluding `Abstain` 
votes).

Proposals can be accepted before the end of the voting period if they meet a special condition. Namely, if the ratio of `Yes` votes to `InitTotalVotingPower`exceeds 2:3, the proposal will be immediately accepted, even if the `Voting period` is not finished. `InitTotalVotingPower` is the total voting power of all bonded Atom holders at the moment when the vote opens. 
This condition exists so that the network can react quickly in case of urgency.

### Inheritance

If a delegator does not vote, it will inherit its validator vote.

* If the delegator votes before its validator, it will not inherit from the 
  validator's vote.
* If the delegator votes after its validator, it will override its validator 
  vote with its own. If the proposal is urgent, it is possible 
  that the vote will close before delegators have a chance to react and 
  override their validator's vote. This is not a problem, as proposals require more than 2/3rd of the total voting power to pass before the end of the voting period. If more than 2/3rd of validators collude, they can censor the votes of delegators anyway.

### Validator’s punishment for non-voting

At present, validators are not punished for failing to vote.

### Governance address

Later, we may add permissioned keys that could only sign txs from certain modules. For the MVP, the `Governance address` will be the main validator address generated at account creation. This address corresponds to a different PrivKey than the Tendermint PrivKey which is responsible for signing consensus messages. Validators thus do not have to sign governance transactions with the sensitive Tendermint PrivKey.

## Software Upgrade

If proposals are of type `SoftwareUpgradeProposal`, then nodes need to upgrade 
their software to the new version that was voted. This process is divided in 
two steps.

### Signal

After a `SoftwareUpgradeProposal` is accepted, validators are expected to 
download and install the new version of the software while continuing to run 
the previous version. Once a validator has downloaded and installed the 
upgrade, it will start signaling to the network that it is ready to switch by 
including the proposal's `proposalID` in its *precommits*.(*Note: Confirmation 
that we want it in the precommit?*)

Note: There is only one signal slot per *precommit*. If several 
`SoftwareUpgradeProposals` are accepted in a short timeframe, a pipeline will 
form and they will be implemented one after the other in the order that they 
were accepted.

### Switch

Once a block contains more than 2/3rd *precommits* where a common 
`SoftwareUpgradeProposal` is signaled, all the nodes (including validator 
nodes, non-validating full nodes and light-nodes) are expected to switch to the
new version of the software. 

*Note: Not clear how the flip is handled programmatically*
