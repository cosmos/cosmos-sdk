<!--
order: 1
-->

# Concepts

_Disclaimer: This is work in progress. Mechanisms are susceptible to change._

The governance process is divided in a few steps that are outlined below:

* **Proposal submission:** Proposal is submitted to the blockchain with a
  deposit.
* **Vote:** Once deposit reaches a certain value (`MinDeposit`), proposal is
  confirmed and vote opens. Bonded Atom holders can then send `TxGovVote`
  transactions to vote on the proposal.
* **Execution** After a period of time, the votes are tallied and depending
  on the result, the messages in the proposal will be executed.

## Proposal submission

### Right to submit a proposal

Every account can submit proposals by sending a `MsgSubmitProposal` transaction.
Once a proposal is submitted, it is identified by its unique `proposalID`.

### Proposal Messages

A proposal includes an array of `sdk.Msg`s which are executed automatically if the
proposal passes. The messages are executed by the governance `ModuleAccount` itself. Modules
such as `x/upgrade`, that want to allow certain messages to be executed by governance
only should add a whitelist within the respective msg server, granting the governance
module the right to execute the message once a quorum has been reached. The governance
module uses the `MsgServiceRouter` to check that these messages are correctly constructed
and have a respective path to execute on but do not perform a full validity check.

## Deposit

To prevent spam, proposals must be submitted with a deposit in the coins defined by
the `MinDeposit` param.

When a proposal is submitted, it has to be accompanied with a deposit that must be
strictly positive, but can be inferior to `MinDeposit`. The submitter doesn't need
to pay for the entire deposit on their own. The newly created proposal is stored in
an _inactive proposal queue_ and stays there until its deposit passes the `MinDeposit`.
Other token holders can increase the proposal's deposit by sending a `Deposit`
transaction. If a proposal doesn't pass the `MinDeposit` before the deposit end time
(the time when deposits are no longer accepted), the proposal will be destroyed: the
proposal will be removed from state and the deposit will be burned (see x/gov `EndBlocker`).
When a proposal deposit passes the `MinDeposit` threshold (even during the proposal
submission) before the deposit end time, the proposal will be moved into the
_active proposal queue_ and the voting period will begin.

The deposit is kept in escrow and held by the governance `ModuleAccount` until the
proposal is finalized (passed or rejected).

### Deposit refund and burn

When a proposal is finalized, the coins from the deposit are either refunded or burned
according to the final tally of the proposal:

* If the proposal is approved or rejected but _not_ vetoed, each deposit will be
  automatically refunded to its respective depositor (transferred from the governance
  `ModuleAccount`).
* When the proposal is vetoed with greater than 1/3, deposits will be burned from the
  governance `ModuleAccount` and the proposal information along with its deposit
  information will be removed from state.
* All refunded or burned deposits are removed from the state. Events are issued when
  burning or refunding a deposit.

## Vote

### Participants

_Participants_ are users that have the right to vote on proposals. On the
Cosmos Hub, participants are bonded Atom holders. Unbonded Atom holders and
other users do not get the right to participate in governance. However, they
can submit and deposit on proposals.

Note that some _participants_ can be forbidden to vote on a proposal under a
certain validator if:

* _participant_ bonded or unbonded Atoms to said validator after proposal
  entered voting period.
* _participant_ became validator after proposal entered voting period.

This does not prevent _participant_ to vote with Atoms bonded to other
validators. For example, if a _participant_ bonded some Atoms to validator A
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

* `Yes`
* `No`
* `NoWithVeto`
* `Abstain`

`NoWithVeto` counts as `No` but also adds a `Veto` vote. `Abstain` option
allows voters to signal that they do not intend to vote in favor or against the
proposal but accept the result of the vote.

_Note: from the UI, for urgent proposals we should maybe add a ‘Not Urgent’
option that casts a `NoWithVeto` vote._

### Weighted Votes

[ADR-037](https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-037-gov-split-vote.md) introduces the weighted vote feature which allows a staker to split their votes into several voting options. For example, it could use 70% of its voting power to vote Yes and 30% of its voting power to vote No.

Often times the entity owning that address might not be a single individual. For example, a company might have different stakeholders who want to vote differently, and so it makes sense to allow them to split their voting power. Currently, it is not possible for them to do "passthrough voting" and giving their users voting rights over their tokens. However, with this system, exchanges can poll their users for voting preferences, and then vote on-chain proportionally to the results of the poll.

To represent weighted vote on chain, we use the following Protobuf message.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/proto/cosmos/gov/v1beta1/gov.proto#L33-L43

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/proto/cosmos/gov/v1beta1/gov.proto#L136-L150

For a weighted vote to be valid, the `options` field must not contain duplicate vote options, and the sum of weights of all options must be equal to 1.

### Quorum

Quorum is defined as the minimum percentage of voting power that needs to be
cast on a proposal for the result to be valid.

### Threshold

Threshold is defined as the minimum proportion of `Yes` votes (excluding
`Abstain` votes) for the proposal to be accepted.

Initially, the threshold is set at 50% of `Yes` votes, excluding `Abstain`
votes. A possibility to veto exists if more than 1/3rd of all votes are
`NoWithVeto` votes.  Note, both of these values are derived from the `TallyParams`
on-chain parameter, which is modifiable by governance.
This means that proposals are accepted iff:

* There exist bonded tokens.
* Quorum has been achieved.
* The proportion of `Abstain` votes is inferior to 1/1.
* The proportion of `NoWithVeto` votes is inferior to 1/3, including
  `Abstain` votes.
* The proportion of `Yes` votes, excluding `Abstain` votes, at the end of
  the voting period is superior to 1/2.

### Inheritance

If a delegator does not vote, it will inherit its validator vote.

* If the delegator votes before its validator, it will not inherit from the
  validator's vote.
* If the delegator votes after its validator, it will override its validator
  vote with its own. If the proposal is urgent, it is possible
  that the vote will close before delegators have a chance to react and
  override their validator's vote. This is not a problem, as proposals require more than 2/3rd of the total voting power to pass before the end of the voting period. Because as little as 1/3 + 1 validation power could collude to censor transactions, non-collusion is already assumed for ranges exceeding this threshold.

### Validator’s punishment for non-voting

At present, validators are not punished for failing to vote.

### Governance address

Later, we may add permissioned keys that could only sign txs from certain modules. For the MVP, the `Governance address` will be the main validator address generated at account creation. This address corresponds to a different PrivKey than the Tendermint PrivKey which is responsible for signing consensus messages. Validators thus do not have to sign governance transactions with the sensitive Tendermint PrivKey.

## Software Upgrade

If proposals are of type `SoftwareUpgradeProposal`, then nodes need to upgrade
their software to the new version that was voted. This process is divided into
two steps:

### Signal

After a `SoftwareUpgradeProposal` is accepted, validators are expected to
download and install the new version of the software while continuing to run
the previous version. Once a validator has downloaded and installed the
upgrade, it will start signaling to the network that it is ready to switch by
including the proposal's `proposalID` in its _precommits_.(_Note: Confirmation
that we want it in the precommit?_)

Note: There is only one signal slot per _precommit_. If several
`SoftwareUpgradeProposals` are accepted in a short timeframe, a pipeline will
form and they will be implemented one after the other in the order that they
were accepted.

### Switch

Once a block contains more than 2/3rd _precommits_ where a common
`SoftwareUpgradeProposal` is signaled, all the nodes (including validator
nodes, non-validating full nodes and light-nodes) are expected to switch to the
new version of the software.

Validators and full nodes can use an automation tool, such as [Cosmovisor](https://github.com/cosmos/cosmos-sdk/blob/main/cosmovisor/README.md), for automatically switching version of the chain.
