<!--
order: 2
-->

# State

## Parameters and base types

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

type ProposalType  string

const (
    ProposalTypePlainText       = "Text"
    ProposalTypeSoftwareUpgrade = "SoftwareUpgrade"
)

type ProposalStatus byte


const (
	StatusNil           ProposalStatus = 0x00
    StatusDepositPeriod ProposalStatus = 0x01  // Proposal is submitted. Participants can deposit on it but not vote
    StatusVotingPeriod  ProposalStatus = 0x02  // MinDeposit is reached, participants can vote
    StatusPassed        ProposalStatus = 0x03  // Proposal passed and successfully executed
    StatusRejected      ProposalStatus = 0x04  // Proposal has been rejected
    StatusFailed        ProposalStatus = 0x05  // Proposal passed but failed execution
)
```

## Deposit

```go
  type Deposit struct {
    Amount      sdk.Coins       //  Amount of coins deposited by depositor
    Depositor   crypto.address  //  Address of depositor
  }
```

## ValidatorGovInfo

This type is used in a temp map when tallying

```go
  type ValidatorGovInfo struct {
    Minus     sdk.Dec
    Vote      Vote
  }
```

## Proposals

`Proposal` objects are used to account votes and generally track the proposal's state. They contain `Content` which denotes
what this proposal is about, and other fields, which are the mutable state of
the governance process.

```go
type Proposal struct {
	Content  // Proposal content interface

	ProposalID       uint64
	Status           ProposalStatus  // Status of the Proposal {Pending, Active, Passed, Rejected}
	FinalTallyResult TallyResult     // Result of Tallies

	SubmitTime     time.Time  // Time of the block where TxGovSubmitProposal was included
	DepositEndTime time.Time  // Time that the Proposal would expire if deposit amount isn't met
	TotalDeposit   sdk.Coins  // Current deposit on this proposal. Initial value is set at InitialDeposit

	VotingStartTime time.Time  //  Time of the block where MinDeposit was reached. -1 if MinDeposit is not reached
	VotingEndTime   time.Time  // Time that the VotingPeriod for this proposal will end and votes will be tallied
}
```

```go
type Content interface {
	GetTitle() string
	GetDescription() string
	ProposalRoute() string
	ProposalType() string
	ValidateBasic() sdk.Error
	String() string
}
```

The `Content` on a proposal is an interface which contains the information about
the `Proposal` such as the tile, description, and any notable changes. Also, this
`Content` type can by implemented by any module. The `Content`'s `ProposalRoute`
returns a string which must be used to route the `Content`'s `Handler` in the
governance keeper. This allows the governance keeper to execute proposal logic
implemented by any module. If a proposal passes, the handler is executed. Only
if the handler is successful does the state get persisted and the proposal finally
passes. Otherwise, the proposal is rejected.

```go
type Handler func(ctx sdk.Context, content Content) sdk.Error
```

The `Handler` is responsible for actually executing the proposal and processing
any state changes specified by the proposal. It is executed only if a proposal
passes during `EndBlock`.

We also mention a method to update the tally for a given proposal:

```go
  func (proposal Proposal) updateTally(vote byte, amount sdk.Dec)
```

## Stores

_Stores are KVStores in the multi-store. The key to find the store is the first
parameter in the list_`

We will use one KVStore `Governance` to store two mappings:

- A mapping from `proposalID|'proposal'` to `Proposal`.
- A mapping from `proposalID|'addresses'|address` to `Vote`. This mapping allows
  us to query all addresses that voted on the proposal along with their vote by
  doing a range query on `proposalID:addresses`.

For pseudocode purposes, here are the two function we will use to read or write in stores:

- `load(StoreKey, Key)`: Retrieve item stored at key `Key` in store found at key `StoreKey` in the multistore
- `store(StoreKey, Key, value)`: Write value `Value` at key `Key` in store found at key `StoreKey` in the multistore

## Proposal Processing Queue

**Store:**

- `ProposalProcessingQueue`: A queue `queue[proposalID]` containing all the
  `ProposalIDs` of proposals that reached `MinDeposit`. During each `EndBlock`,
  all the proposals that have reached the end of their voting period are processed.
  To process a finished proposal, the application tallies the votes, computes the
  votes of each validator and checks if every validator in the validator set has
  voted. If the proposal is accepted, deposits are refunded. Finally, the proposal
  content `Handler` is executed.

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
        for each (amount, depositor) in proposal.Deposits
          depositor.AtomBalance += amount

        stateWriter, err := proposal.Handler()
        if err != nil
            // proposal passed but failed during state execution
            proposal.CurrentStatus = ProposalStatusFailed
         else
            // proposal pass and state is persisted
            proposal.CurrentStatus = ProposalStatusAccepted
            stateWriter.save()
      else
        // proposal was rejected
        proposal.CurrentStatus = ProposalStatusRejected

      store(Governance, <proposalID|'proposal'>, proposal)
```
