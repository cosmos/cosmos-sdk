# Implementation (1/2)

## State

### Parameters and base types

`Parameters` define the rules according to which votes are run. There can only
be one active parameter set at any given time. If governance wants to change a
parameter set, either to modify a value or add/remove a parameter field, a new
parameter set has to be created and the previous one rendered inactive.

```go
type DepositParams struct {
  MinDeposit        sdk.Coins  //  Minimum deposit for a proposal to enter voting period.
  MaxDepositPeriod  time.Time  //  Maximum period for Atom holders to deposit on a proposal. Initial value: 2 months
}
```

```go
type VotingParams struct {
  VotingPeriod      time.Time  //  Length of the voting period. Initial value: 2 weeks
}
```

```go
type TallyParams struct {
  Quorum            sdk.Dec  //  Minimum percentage of stake that needs to vote for a proposal to be considered valid
  Threshold         sdk.Dec  //  Minimum proportion of Yes votes for proposal to pass. Initial value: 0.5
  Veto              sdk.Dec  //  Minimum proportion of Veto votes to Total votes ratio for proposal to be vetoed. Initial value: 1/3
  GovernancePenalty sdk.Dec  //  Penalty if validator does not vote
}
```

Parameters are stored in a global `GlobalParams` KVStore.

Additionally, we introduce some basic types:

```go

type Vote byte

const (
    VoteYes         = 0x1
    VoteNo          = 0x2
    VoteNoWithVeto  = 0x3
    VoteAbstain     = 0x4
)

type ProposalType  byte

const (
    ProposalTypePlainText       = 0x1 // Plain text proposals
    ProposalTypeSoftwareUpgrade = 0x2 // Text proposal inducing a software upgrade
)

type ProposalStatus byte


const (
    ProposalStatusOpen      = 0x1   // Proposal is submitted. Participants can deposit on it but not vote
    ProposalStatusActive    = 0x2   // MinDeposit is reachhed, participants can vote
    ProposalStatusAccepted  = 0x3   // Proposal has been accepted
    ProposalStatusRejected  = 0x4   // Proposal has been rejected
    ProposalStatusClosed   = 0x5   // Proposal never reached MinDeposit
)
```

### Deposit

```go
  type Deposit struct {
    Amount      sdk.Coins       //  Amount of coins deposited by depositor
    Depositor   crypto.address  //  Address of depositor
  }
```

### ValidatorGovInfo

This type is used in a temp map when tallying

```go
  type ValidatorGovInfo struct {
    Minus     sdk.Dec
    Vote      Vote
  }
```

### Proposals

`Proposals` are an item to be voted on.

```go
type Proposal struct {
  Title                 string              //  Title of the proposal
  Description           string              //  Description of the proposal
  Type                  ProposalType        //  Type of proposal. Initial set {PlainTextProposal, SoftwareUpgradeProposal}
  TotalDeposit          sdk.Coins           //  Current deposit on this proposal. Initial value is set at InitialDeposit
  Deposits              []Deposit           //  List of deposits on the proposal
  SubmitTime            time.Time           //  Time of the block where TxGovSubmitProposal was included
  DepositEndTime        time.Time           //  Time that the DepositPeriod of a proposal would expire
  Submitter             sdk.AccAddress      //  Address of the submitter

  VotingStartTime       time.Time           //  Time of the block where MinDeposit was reached. time.Time{} if MinDeposit is not reached
  VotingEndTime         time.Time           //  Time of the block that the VotingPeriod for a proposal will end.
  CurrentStatus         ProposalStatus      //  Current status of the proposal

  YesVotes              sdk.Dec
  NoVotes               sdk.Dec
  NoWithVetoVotes       sdk.Dec
  AbstainVotes          sdk.Dec
}
```

We also mention a method to update the tally for a given proposal:

```go
  func (proposal Proposal) updateTally(vote byte, amount sdk.Dec)
```

### Stores

*Stores are KVStores in the multistore. The key to find the store is the first parameter in the list*`

We will use one KVStore `Governance` to store two mappings:

* A mapping from `proposalID|'proposal'` to `Proposal`
* A mapping from `proposalID|'addresses'|address` to `Vote`. This mapping allows us to query all addresses that voted on the proposal along with their vote by doing a range query on `proposalID:addresses`


For pseudocode purposes, here are the two function we will use to read or write in stores:

* `load(StoreKey, Key)`: Retrieve item stored at key `Key` in store found at key `StoreKey` in the multistore
* `store(StoreKey, Key, value)`: Write value `Value` at key `Key` in store found at key `StoreKey` in the multistore

### Proposal Processing Queue

**Store:**
* `ProposalProcessingQueue`: A queue `queue[proposalID]` containing all the
  `ProposalIDs` of proposals that reached `MinDeposit`. Each `EndBlock`, all the proposals
  that have reached the end of their voting period are processed.
  To process a finished proposal, the application tallies the votes, compute the votes of
  each validator and checks if every validator in the valdiator set have voted.
  If the proposal is accepted, deposits are refunded.

And the pseudocode for the `ProposalProcessingQueue`:

```go
  in EndBlock do

    for finishedProposalID in GetAllFinishedProposalIDs(block.Time)
      proposal = load(Governance, <proposalID|'proposal'>) // proposal is a const key

      validators = Keeper.getAllValidators()
      tmpValMap := map(sdk.AccAddress)ValidatorGovInfo

      // Initiate mapping at 0. This is the amount of shares of the validator's vote that will be overridden by their delegator's votes
      for each validator in validators
        tmpValMap(validator.OperatorAddr).Minus = 0

      // Tally
      voterIterator = rangeQuery(Governance, <proposalID|'addresses'>) //return all the addresses that voted on the proposal
      for each (voterAddress, vote) in voterIterator
        delegations = stakingKeeper.getDelegations(voterAddress) // get all delegations for current voter

        for each delegation in delegations
          // make sure delegation.Shares does NOT include shares being unbonded
          tmpValMap(delegation.ValidatorAddr).Minus += delegation.Shares
          proposal.updateTally(vote, delegation.Shares)

        _, isVal = stakingKeeper.getValidator(voterAddress)
        if (isVal)
          tmpValMap(voterAddress).Vote = vote

      tallyingParam = load(GlobalParams, 'TallyingParam')

      // Update tally if validator voted they voted
      for each validator in validators
        if tmpValMap(validator).HasVoted
          proposal.updateTally(tmpValMap(validator).Vote, (validator.TotalShares - tmpValMap(validator).Minus))



      // Check if proposal is accepted or rejected
      totalNonAbstain := proposal.YesVotes + proposal.NoVotes + proposal.NoWithVetoVotes
      if (proposal.Votes.YesVotes/totalNonAbstain > tallyingParam.Threshold AND proposal.Votes.NoWithVetoVotes/totalNonAbstain  < tallyingParam.Veto)
        //  proposal was accepted at the end of the voting period
        //  refund deposits (non-voters already punished)
        proposal.CurrentStatus = ProposalStatusAccepted
        for each (amount, depositor) in proposal.Deposits
          depositor.AtomBalance += amount

      else
        // proposal was rejected
        proposal.CurrentStatus = ProposalStatusRejected

      store(Governance, <proposalID|'proposal'>, proposal)
```
