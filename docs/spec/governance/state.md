# Implementation (1/2)

## State

### Procedures

`Procedures` define the rule according to which votes are run. There can only 
be one active procedure at any given time. If governance wants to change a 
procedure, either to modify a value or add/remove a parameter, a new procedure 
has to be created and the previous one rendered inactive.

```go

type VoteType byte

const (
    VoteTypeYes         = 0x1
    VoteTypeNo          = 0x2
    VoteTypeNoWithVeto  = 0x3
    VoteTypeAbstain     = 0x4
)

type ProposalType  byte

const (
    ProposalTypePlainText = 0x1
    ProposalTypeSoftwareUpgrade = 0x2

)

type Procedure struct {
  VotingPeriod      int64               //  Length of the voting period. Initial value: 2 weeks
  MinDeposit        int64               //  Minimum deposit for a proposal to enter voting period. 
  ProposalTypes     []string            //  Types available to submitters. {PlainTextProposal, SoftwareUpgradeProposal}
  Threshold         rational.Rational   //  Minimum propotion of Yes votes for proposal to pass. Initial value: 0.5
  Veto              rational.Rational   //  Minimum value of Veto votes to Total votes ratio for proposal to be vetoed. Initial value: 1/3
  MaxDepositPeriod  int64               //  Maximum period for Atom holders to deposit on a proposal. Initial value: 2 months
  GovernancePenalty int64               //  Penalty if validator does not vote
  
  IsActive          bool                //  If true, procedure is active. Only one procedure can have isActive true.
}
```

The current active procedure is stored in a global `params` KVStore.

### Deposit

```go
  type Deposit struct {
    Amount      sdk.Coins       //  Amount of coins deposited by depositer
    Depositer   crypto.address  //  Address of depositer
  }
```

### Votes

```go
  type Votes struct {
    Yes          int64
    No           int64
    NoWithVeto   int64
    Abstain      int64
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
  Submitter             crypto.Address      //  Address of the submitter
  
  VotingStartBlock      int64               //  Height of the block where MinDeposit was reached. -1 if MinDeposit is not reached
  InitTotalVotingPower  int64               //  Total voting power when proposal enters voting period (default 0)
  InitProcedure         Procedure           //  Active Procedure when proposal enters voting period

  Votes                 Votes               //  Total votes for each option
}
```

We also introduce a type `ValidatorGovInfo`

```go
type ValidatorGovInfo struct {
  InitVotingPower     int64   //  Voting power of validator when proposal enters voting period
  Minus               int64   //  Minus of validator, used to compute validator's voting power
}
```

### Stores

*Stores are KVStores in the multistore. The key to find the store is the first parameter in the list*


* `Proposals: int64 => Proposal` maps `proposalID` to the `Proposal`
  `proposalID`
* `Options: <proposalID | voterAddress | Address> => VoteType`: maps to the vote of the `voterAddress` for `proposalID` re its delegation to `Address`.
   Returns 0x0 If `voterAddress` has not voted under this validator.
* `ValidatorGovInfos: <proposalID | Address> => ValidatorGovInfo`: maps to the gov info for the `Address` and `proposalID`.
  Returns `nil` if proposal has not entered voting period or if `address` was not the 
  address of a validator when proposal entered voting period.

For pseudocode purposes, here are the two function we will use to read or write in stores:

* `load(StoreKey, Key)`: Retrieve item stored at key `Key` in store found at key `StoreKey` in the multistore
* `store(StoreKey, Key, value)`: Write value `Value` at key `Key` in store found at key `StoreKey` in the multistore

### Proposal Processing Queue

**Store:**
* `ProposalProcessingQueue`: A queue `queue[proposalID]` containing all the 
  `ProposalIDs` of proposals that reached `MinDeposit`. Each round, the oldest 
  element of `ProposalProcessingQueue` is checked during `BeginBlock` to see if
  `CurrentBlock == VotingStartBlock + InitProcedure.VotingPeriod`. If it is, 
  then the application checks if validators in `InitVotingPowerList` have voted
  and, if not, applies `GovernancePenalty`. If the proposal is accepted, deposits are refunded.
  After that proposal is ejected from `ProposalProcessingQueue` and the next element of the queue is evaluated. 
  Note that if a proposal is accepted under the special condition, 
  its `ProposalID` must be ejected from `ProposalProcessingQueue`.

And the pseudocode for the `ProposalProcessingQueue`:

```go
  in BeginBlock do 
    
    checkProposal()  // First call of the recursive function 
    
    
  // Recursive function. First call in BeginBlock
  func checkProposal()  
    proposalID = ProposalProcessingQueue.Peek()
    if (proposalID == nil)
      return

    proposal = load(Proposals, proposalID) 

    if (proposal.Votes.YesVotes/proposal.InitTotalVotingPower > 2/3)

      // proposal accepted early by super-majority
      // no punishments; refund deposits

      ProposalProcessingQueue.pop()

      var newDeposits []Deposits

      // XXX: why do we need to reset deposits? cant we just clear it ?
      for each (amount, depositer) in proposal.Deposits
        newDeposits.append[{0, depositer}]
        depositer.AtomBalance += amount

      proposal.Deposits = newDeposits
      store(Proposals, proposalID, proposal)

      checkProposal()

    else if (CurrentBlock == proposal.VotingStartBlock + proposal.Procedure.VotingPeriod)

      ProposalProcessingQueue.pop()
      activeProcedure = load(params, 'ActiveProcedure')

      for each validator in CurrentBondedValidators
        validatorGovInfo = load(ValidatorGovInfos, <proposalID | validator.Address>)
        
        if (validatorGovInfo.InitVotingPower != nil)
          // validator was bonded when vote started

          validatorOption = load(Options, <proposalID | validator.Address>)
          if (validatorOption == nil)
            // validator did not vote
            slash validator by activeProcedure.GovernancePenalty


      totalNonAbstain = proposal.Votes.YesVotes + proposal.Votes.NoVotes + proposal.Votes.NoWithVetoVotes
      if( proposal.Votes.YesVotes/totalNonAbstain > 0.5 AND proposal.Votes.NoWithVetoVotes/totalNonAbstain  < 1/3)

        //  proposal was accepted at the end of the voting period
        //  refund deposits (non-voters already punished)

        var newDeposits []Deposits

        for each (amount, depositer) in proposal.Deposits
          newDeposits.append[{0, depositer}]
          depositer.AtomBalance += amount

        proposal.Deposits = newDeposits
        store(Proposals, proposalID, proposal)

        checkProposal()        
```
