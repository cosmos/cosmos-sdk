# Implementation (1/2)

## State

### Procedures and base types

`Procedures` define the rule according to which votes are run. There can only 
be one active procedure at any given time. If governance wants to change a 
procedure, either to modify a value or add/remove a parameter, a new procedure 
has to be created and the previous one rendered inactive.


```go
type Procedure struct {
  VotingPeriod      int64               //  Length of the voting period. Initial value: 2 weeks
  MinDeposit        sdk.Coins           //  Minimum deposit for a proposal to enter voting period. 
  Threshold         rational.Rational   //  Minimum propotion of Yes votes for proposal to pass. Initial value: 0.5
  Veto              rational.Rational   //  Minimum value of Veto votes to Total votes ratio for proposal to be vetoed. Initial value: 1/3
  MaxDepositPeriod  int64               //  Maximum period for Atom holders to deposit on a proposal. Initial value: 2 months
  GovernancePenalty sdk.Rat               //  Penalty if validator does not vote
  
  IsActive          bool                //  If true, procedure is active. Only one procedure can have isActive true.
}
```

The current active procedure is stored in a global `params` KVStore.

And some basic types:

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
    ProposalTypePlainText       = 0x1
    ProposalTypeSoftwareUpgrade = 0x2
)

type ProposalStatus byte

const (
    ProposalStatusOpen      = 0x1   // Proposal is submitted. Participants can deposit on it but not vote
    ProposalStatusActive    = 0x2   // MinDeposit is reachhed, participants can vote
    ProposalStatusAccepted  = 0x3   // Proposal has been accepted
    ProposalStatusRejected  = 0x4   // Proposal has been rejected
)
```

### Deposit

```go
  type Deposit struct {
    Amount      sdk.Coins       //  Amount of coins deposited by depositer
    Depositer   crypto.address  //  Address of depositer
  }
```

### ValidatorGovInfo

This type is used in a temp map when tallying 

```go
  type ValidatorGovInfo struct {
    Minus     sdk.Rat
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
  SubmitBlock           int64               //  Height of the block where TxGovSubmitProposal was included
  Submitter             sdk.Address      //  Address of the submitter
  
  VotingStartBlock      int64               //  Height of the block where MinDeposit was reached. -1 if MinDeposit is not reached
  InitProcedure         Procedure           //  Active Procedure when proposal enters voting period
  CurrentStatus         ProposalStatus      //  Current status of the proposal

  YesVotes              sdk.Rat
  NoVotes               sdk.Rat
  NoWithVetoVotes       sdk.Rat
  AbstainVotes          sdk.Rat
}
```

We also mention a method to update the tally for a given proposal:

```go
  func (proposal Proposal) updateTally(vote byte, amount sdk.Rat)
```

### Stores

*Stores are KVStores in the multistore. The key to find the store is the first parameter in the list*`

We will use one KVStore `Governance` to store two mappings:

* A mapping from `proposalID` to `Proposal`
* A mapping from `proposalID:addresses:address` to `Vote`. This mapping allows us to query all addresses that voted on the proposal along with their vote by doing a range query on `proposalID:addresses`


For pseudocode purposes, here are the two function we will use to read or write in stores:

* `load(StoreKey, Key)`: Retrieve item stored at key `Key` in store found at key `StoreKey` in the multistore
* `store(StoreKey, Key, value)`: Write value `Value` at key `Key` in store found at key `StoreKey` in the multistore

### Proposal Processing Queue

**Store:**
* `ProposalProcessingQueue`: A queue `queue[proposalID]` containing all the 
  `ProposalIDs` of proposals that reached `MinDeposit`. Each round, the oldest 
  element of `ProposalProcessingQueue` is checked during `BeginBlock` to see if
  `CurrentBlock == VotingStartBlock + InitProcedure.VotingPeriod`. If it is, 
  then the application tallies the votes, compute the votes of each validator and checks if every validator in the valdiator set have voted
  and, if not, applies `GovernancePenalty`. If the proposal is accepted, deposits are refunded.
  After that proposal is ejected from `ProposalProcessingQueue` and the next element of the queue is evaluated. 

And the pseudocode for the `ProposalProcessingQueue`:

```go
  in EndBlock do 
    
    checkProposal()  // First call of the recursive function 
    
    
  // Recursive function. First call in BeginBlock
  func checkProposal()  
    proposalID = ProposalProcessingQueue.Peek()
    if (proposalID == nil)
      return

    proposal = load(Governance, proposalID) 

    if (CurrentBlock == proposal.VotingStartBlock + proposal.Procedure.VotingPeriod && proposal.CurrentStatus == ProposalStatusActive)

    // End of voting period, tally

      ProposalProcessingQueue.pop()

      validators = stakeKeeper.getAllValidators()
      tmpValMap := map(sdk.Address)ValidatorGovInfo

      // Initiate mapping at 0. Validators that remain at 0 at the end of tally will be punished
      for each validator in validators
        tmpValMap(validator).Minus = 0
      
      voterIterator = rangeQuery(Governance, <proposalID|addresses>) //return all the addresses that voted on the proposal

      // Tally
      for each (voterAddress, vote) in voterIterator
        delegations = stakeKeeper.getDelegations(voterAddress) // get all delegations for current voter

        for each delegation in delegations
          tmpValMap(delegation.ValidatorAddr).Minus += delegation.Shares
          proposal.updateTally(vote, delegation.Shares)

        _, isVal = stakeKeeper.getValidator(voterAddress)
        if (isVal)
        tmpValMap(voterAddress).Vote = vote

      // Slash validators that did not vote, or update tally if they voted
      for each validator in validators
        if (!tmpValMap(validator).HasVoted)
          slash validator by proposal.Procedure.GovernancePenalty
        else
          proposal.updateTally(tmpValMap(validator).Vote, (validator.TotalShares - tmpValMap(validator).Minus))

      // Check if proposal is accepted or rejected
      totalNonAbstain := proposal.YesVotes + proposal.NoVotes + proposal.NoWithVetoVotes
      if (proposal.Votes.YesVotes/totalNonAbstain > proposal.InitProcedure.Threshold AND proposal.Votes.NoWithVetoVotes/totalNonAbstain  < proposal.InitProcedure.Veto)
        //  proposal was accepted at the end of the voting period
        //  refund deposits (non-voters already punished)
        proposal.CurrentStatus = ProposalStatusAccepted
        for each (amount, depositer) in proposal.Deposits
          depositer.AtomBalance += amount

      else 
        // proposal was rejected
        proposal.CurrentStatus = ProposalStatusRejected

      store(Governance, proposalID, proposal)
      checkProposal()        
```
