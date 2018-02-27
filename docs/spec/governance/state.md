# Implementation (1/2)

## State

### Procedures

`Procedures` define the rule according to which votes are run. There can only 
be one active procedure at any given time. If governance wants to change a 
procedure, either to modify a value or add/remove a parameter, a new procedure 
has to be created and the previous one rendered inactive.

```go
type Procedure struct {
  VotingPeriod      int64               //  Length of the voting period. Initial value: 2 weeks
  MinDeposit        int64               //  Minimum deposit for a proposal to enter voting period. 
  OptionSet         []string            //  Options available to voters. {Yes, No, NoWithVeto, Abstain}
  ProposalTypes     []string            //  Types available to submitters. {PlainTextProposal, SoftwareUpgradeProposal}
  Threshold         rational.Rational   //  Minimum propotion of Yes votes for proposal to pass. Initial value: 0.5
  Veto              rational.Rational   //  Minimum value of Veto votes to Total votes ratio for proposal to be vetoed. Initial value: 1/3
  MaxDepositPeriod  int64               //  Maximum period for Atom holders to deposit on a proposal. Initial value: 2 months
  GovernancePenalty int64               //  Penalty if validator does not vote
  
  IsActive          bool                //  If true, procedure is active. Only one procedure can have isActive true.
}
```


### Proposals

`Proposals` are item to be voted on. 

```go
type Proposal struct {
  Title                 string              //  Title of the proposal
  Description           string              //  Description of the proposal
  Type                  string              //  Type of proposal. Initial set {PlainTextProposal, SoftwareUpgradeProposal}
  Deposit               int64               //  Current deposit on this proposal. Initial value is set at InitialDeposit
  SubmitBlock           int64               //  Height of the block where TxGovSubmitProposal was included
  
  VotingStartBlock      int64               //  Height of the block where MinDeposit was reached. -1 if MinDeposit is not reached
  InitTotalVotingPower  int64               //  Total voting power when proposal enters voting period (default 0)
  InitProcedureNumber   int16               //  Procedure number of the active procedure when proposal enters voting period (default -1)
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

* `Procedures`: a mapping `map[int16]Procedure` of procedures indexed by their 
  `ProcedureNumber`. First ever procedure is found at index '1'. Index '0' is reserved for parameter `ActiveProcedureNumber` which returns the number of the current procedure.
* `Proposals`: A mapping `map[int64]Proposal` of proposals indexed by their 
  `proposalID`
* `Votes`: A mapping `map[[]byte]int64` of votes indexed by `<proposalID>:<option>` as `[]byte`. Given a `proposalID` and an `option`, returns votes for  that option.
* `Deposits`: A mapping `map[[]byte]int64` of deposits indexed by 
  `<proposalID>:<depositorPubKey>` as `[]byte`. Given a `proposalID` and a 
  `PubKey`, returns deposit (`nil` if `PubKey` has not deposited on the 
  proposal)
* `Options`: A mapping `map[[]byte]string` of options indexed by 
  `<proposalID>:<voterPubKey>:<validatorPubKey>` as `[]byte`. Given a 
  `proposalID`, a `PubKey` and a validator's `PubKey`, returns option chosen by
  this `PubKey` for this validator (`nil` if `PubKey` has not voted under this 
  validator)
* `ValidatorGovInfos`: A mapping `map[[]byte]ValidatorGovInfo` of validator's 
  governance infos indexed by `<proposalID>:<validatorGovPubKey>`. Returns 
  `nil` if proposal has not entered voting period or if `PubKey` was not the 
  governance public key of a validator when proposal entered voting period.

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
  and, if not, applies `GovernancePenalty`. After that proposal is ejected from 
  `ProposalProcessingQueue` and the next element of the queue is evaluated. 
  Note that if a proposal is urgent and accepted under the special condition, 
  its `ProposalID` must be ejected from `ProposalProcessingQueue`.

And the pseudocode for the `ProposalProcessingQueue`:

```go
  in BeginBlock do 
    
    checkProposal()  // First call of the recursive function 
    
    
  // Recursive function. First call in BeginBlock
  func checkProposal()  
    if (ProposalProcessingQueue.Peek() == nil)
      return

    else
      proposalID = ProposalProcessingQueue.Peek()
      proposal = load(Proposals, proposalID) 
      initProcedure = load(Procedures, proposal.InitProcedureNumber)
      yesVotes = load(Votes, <proposalID>:<'Yes'>)

      if (yesVotes/proposal.InitTotalVotingPower >= 2/3)

        // proposal was urgent and accepted under the special condition
        // no punishment

        ProposalProcessingQueue.pop()
        checkProposal()

      else if (CurrentBlock == proposal.VotingStartBlock + initProcedure.VotingPeriod)

        activeProcedureNumber = load(Procedures, '0')
        activeProcedure = load(Procedures, activeProcedureNumber)

        for each validator in CurrentBondedValidators
          validatorGovInfo = load(multistore, ValidatorGovInfos, validator.GovPubKey)
          
          if (validatorGovInfo.InitVotingPower != nil)
            // validator was bonded when vote started

            validatorOption = load(Options, validator.GovPubKey)
            if (validatorOption == nil)
              // validator did not vote
              slash validator by activeProcedure.GovernancePenalty

        ProposalProcessingQueue.pop()
        checkProposal()        
```