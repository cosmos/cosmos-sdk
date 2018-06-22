# Implementation (2/2)

## Transactions

### Proposal Submission

Proposals can be submitted by any Atom holder via a `TxGovSubmitProposal` 
transaction.

```go
type TxGovSubmitProposal struct {
  Title           string        //  Title of the proposal
  Description     string        //  Description of the proposal
  Type            ProposalType  //  Type of proposal
  InitialDeposit  int64         //  Initial deposit paid by sender. Must be strictly positive.
}
```

**State modifications:**
* Generate new `proposalID`
* Create new `Proposal`
* Initialise `Proposals` attributes
* Decrease balance of sender by `InitialDeposit`
* If `MinDeposit` is reached:
  * Push `proposalID` in  `ProposalProcessingQueueEnd`
  * Store each validator's voting power in `ValidatorGovInfos`

A `TxGovSubmitProposal` transaction can be handled according to the following 
pseudocode.

```go
// PSEUDOCODE //
// Check if TxGovSubmitProposal is valid. If it is, create proposal //

upon receiving txGovSubmitProposal from sender do
  
  if !correctlyFormatted(txGovSubmitProposal) then 
    // check if proposal is correctly formatted. Includes fee payment.
    throw
  
  initialDeposit = txGovSubmitProposal.InitialDeposit 
  if (initialDeposit <= 0) OR (sender.AtomBalance < initialDeposit) then 
    // InitialDeposit is negative or null OR sender has insufficient funds
    throw
  
  sender.AtomBalance -= initialDeposit
  
  proposalID = generate new proposalID
  proposal = NewProposal()
  
  proposal.Title = txGovSubmitProposal.Title
  proposal.Description = txGovSubmitProposal.Description
  proposal.Type = txGovSubmitProposal.Type
  proposal.TotalDeposit = initialDeposit
  proposal.SubmitBlock = CurrentBlock
  proposal.Deposits.append({initialDeposit, sender})
  proposal.Submitter = sender
  proposal.Votes.Yes = 0
  proposal.Votes.No = 0
  proposal.Votes.NoWithVeto = 0
  proposal.Votes.Abstain = 0
  
  activeProcedure = load(params, 'ActiveProcedure')
  
  if (initialDeposit < activeProcedure.MinDeposit) then  
    // MinDeposit is not reached
    
    proposal.VotingStartBlock = -1
    proposal.InitTotalVotingPower = 0
  
  else  
    // MinDeposit is reached
    
    proposal.VotingStartBlock = CurrentBlock
    proposal.InitTotalVotingPower = TotalVotingPower
    proposal.InitProcedure = activeProcedure
    
    for each validator in CurrentBondedValidators
      // Store voting power of each bonded validator

      validatorGovInfo = new ValidatorGovInfo
      validatorGovInfo.InitVotingPower = validator.VotingPower
      validatorGovInfo.Minus = 0

      store(ValidatorGovInfos, <proposalID | validator.Address>, validatorGovInfo)
    
    ProposalProcessingQueue.push(proposalID)
  
  store(Proposals, proposalID, proposal) // Store proposal in Proposals mapping
  return proposalID
```

### Deposit

Once a proposal is submitted, if 
`Proposal.TotalDeposit < ActiveProcedure.MinDeposit`, Atom holders can send 
`TxGovDeposit` transactions to increase the proposal's deposit.

```go
type TxGovDeposit struct {
  ProposalID    int64   // ID of the proposal
  Deposit       int64   // Number of Atoms to add to the proposal's deposit
}
```

**State modifications:**
* Decrease balance of sender by `deposit`
* Add `deposit` of sender in `proposal.Deposits`
* Increase `proposal.TotalDeposit` by sender's `deposit`
* If `MinDeposit` is reached:
  * Push `proposalID` in  `ProposalProcessingQueueEnd`
  * Store each validator's voting power in `ValidatorGovInfos`

A `TxGovDeposit` transaction has to go through a number of checks to be valid. 
These checks are outlined in the following pseudocode.

```go
// PSEUDOCODE //
// Check if TxGovDeposit is valid. If it is, increase deposit and check if MinDeposit is reached

upon receiving txGovDeposit from sender do
  // check if proposal is correctly formatted. Includes fee payment.
  
  if !correctlyFormatted(txGovDeposit) then  
    throw
  
  proposal = load(Proposals, txGovDeposit.ProposalID)

  if (proposal == nil) then  
    // There is no proposal for this proposalID
    throw
  
  if (txGovDeposit.Deposit <= 0) OR (sender.AtomBalance < txGovDeposit.Deposit)
    // deposit is negative or null OR sender has insufficient funds
    throw
  
  activeProcedure = load(params, 'ActiveProcedure')

  if (proposal.TotalDeposit >= activeProcedure.MinDeposit) then  
    // MinDeposit was reached
    // TODO: shouldnt we do something here ?
    throw
  
  else
    if (CurrentBlock >= proposal.SubmitBlock + activeProcedure.MaxDepositPeriod) then 
      // Maximum deposit period reached
      throw
    
    // sender can deposit
    
    sender.AtomBalance -= txGovDeposit.Deposit

    proposal.Deposits.append({txGovVote.Deposit, sender})
    proposal.TotalDeposit += txGovDeposit.Deposit
    
    if (proposal.TotalDeposit >= activeProcedure.MinDeposit) then  
      // MinDeposit is reached, vote opens
      
      proposal.VotingStartBlock = CurrentBlock
      proposal.InitTotalVotingPower = TotalVotingPower
      proposal.InitProcedure = activeProcedure
      
      for each validator in CurrentBondedValidators
        // Store voting power of each bonded validator

        validatorGovInfo = NewValidatorGovInfo()
        validatorGovInfo.InitVotingPower = validator.VotingPower
        validatorGovInfo.Minus = 0

        store(ValidatorGovInfos, <proposalID | validator.Address>, validatorGovInfo)
      
      ProposalProcessingQueue.push(txGovDeposit.ProposalID)  

    store(Proposals, txGovVote.ProposalID, proposal)
```

### Vote

Once `ActiveProcedure.MinDeposit` is reached, voting period starts. From there,
bonded Atom holders are able to send `TxGovVote` transactions to cast their 
vote on the proposal.

```go
  type TxGovVote struct {
    ProposalID           int64           //  proposalID of the proposal
    Option               string          //  option chosen by the voter
    ValidatorAddress      crypto.address //  Address of the validator voter wants to tie its vote to
  }
```

**State modifications:**
* If sender is not a validator and validator has not voted, initialize or 
  increase minus of validator by sender's `voting power`
* If sender is not a validator and validator has voted, decrease 
  votes of `validatorOption` by sender's `voting power`
* If sender is not a validator, increase votes of `txGovVote.Option`
  by sender's `voting power`
* If sender is a validator, increase votes of `txGovVote.Option` by 
  validator's `InitVotingPower - minus` (`minus` can be equal to 0)

Votes need to be tied to a validator in order to compute validator's voting 
power. If a delegator is bonded to multiple validators, it will have to send 
one transaction per validator (the UI should facilitate this so that multiple 
transactions can be sent in one "vote flow"). If the sender is the validator 
itself, then it will input its own address as `Address`

Next is a pseudocode proposal of the way `TxGovVote` transactions are 
handled:

```go
  // PSEUDOCODE //
  // Check if TxGovVote is valid. If it is, count vote//
  
  upon receiving txGovVote from sender do
    // check if proposal is correctly formatted. Includes fee payment.    
    
    if !correctlyFormatted(txGovDeposit) then  
      throw
    
    proposal = load(Proposals, txGovDeposit.ProposalID)

    if (proposal == nil) then  
      // There is no proposal for this proposalID
      throw
    
    validator = load(CurrentValidators, txGovVote.Address)
        if (validator == nil) then 
         
          // Throws if
          // ValidatorAddress is not the address of a current validator
          
          throw
          
        else
           option = load(Options, <txGovVote.ProposalID>:<sender>:<txGovVote.ValidatorAddress>)

    if (option != nil)
     // sender has already voted with the Atoms bonded to Address
     throw

    if  (proposal.VotingStartBlock < 0) OR  
        (CurrentBlock > proposal.VotingStartBlock + proposal.InitProcedure.VotingPeriod) OR 
        (proposal.VotingStartBlock < lastBondingBlock(sender, txGovVote.Address) OR   
        (proposal.VotingStartBlock < lastUnbondingBlock(sender, txGovVote.Address) OR   
        (proposal.Votes.YesVotes/proposal.InitTotalVotingPower >= 2/3) then   

        // Throws if
        // Vote has not started OR if
        // Vote had ended OR if
        // sender bonded Atoms to Address after start of vote OR if
        // sender unbonded Atoms from Address after start of vote OR if
        // special condition is met, i.e. proposal is accepted and closed

        throw     

    validatorGovInfo = load(ValidatorGovInfos, <txGovVote.ProposalID>:<validator.Address>)

    if (validatorGovInfo == nil)
      // validator became validator after proposal entered voting period 
      throw

    // sender can vote, check if sender == validator and store sender's option in Options
    
    store(Options, <txGovVote.ProposalID>:<sender>:<txGovVote.Address>, txGovVote.Option)

    if (sender != validator.address)
      // Here, sender is not the Address of the validator whose Address is txGovVote.Address

      if sender does not have bonded Atoms to txGovVote.Address then
        // check in Staking module
        throw

      validatorOption = load(Options, <txGovVote.ProposalID>:<sender>:<txGovVote.Address>)

      if (validatorOption == nil)
        // Validator has not voted already

        validatorGovInfo.Minus += sender.bondedAmounTo(txGovVote.Address)
        store(ValidatorGovInfos, <txGovVote.ProposalID>:<validator.Address>, validatorGovInfo)

      else
        // Validator has already voted
        // Reduce votes of option chosen by validator by sender's bonded Amount

        proposal.Votes.validatorOption -= sender.bondedAmountTo(txGovVote.Address)

      // increase votes of option chosen by sender by bonded Amount

      senderOption = txGovVote.Option
      propoal.Votes.senderOption -= sender.bondedAmountTo(txGovVote.Address)

      store(Proposals, txGovVote.ProposalID, proposal)
        

    else 
      // sender is the address of the validator whose main Address is txGovVote.Address
      // i.e. sender == validator

      proposal.Votes.validatorOption += (validatorGovInfo.InitVotingPower - validatorGovInfo.Minus)

      store(Proposals, txGovVote.ProposalID, proposal)
```
