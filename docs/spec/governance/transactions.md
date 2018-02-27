# Implementation (2/2)

## Transactions

### Proposal Submission

Proposals can be submitted by any Atom holder via a `TxGovSubmitProposal` 
transaction.

```go
type TxGovSubmitProposal struct {
  Title           string        //  Title of the proposal
  Description     string        //  Description of the proposal
  Type            string        //  Type of proposal. Initial set {PlainTextProposal, SoftwareUpgradeProposal}
  InitialDeposit  int64         //  Initial deposit paid by sender. Must be strictly positive.
}
```

**State modifications:**
* Generate new `proposalID`
* Create new `Proposal`
* Initialise `Proposals` attributes
* Store sender's deposit in `Deposits`
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
  
  else
    if (txGovSubmitProposal.InitialDeposit <= 0) OR (sender.AtomBalance < InitialDeposit) then 
      // InitialDeposit is negative or null OR sender has insufficient funds
      
      throw
    
    else
      sender.AtomBalance -= txGovSubmitProposal.InitialDeposit
      
      proposalID = generate new proposalID
      proposal = NewProposal()
      
      proposal.Title = txGovSubmitProposal.Title
      proposal.Description = txGovSubmitProposal.Description
      proposal.Type = txGovSubmitProposal.Type
      proposal.Deposit = txGovSubmitProposal.InitialDeposit
      proposal.SubmitBlock = CurrentBlock
      
      store(Deposits, <proposalID>:<sender>, txGovSubmitProposal.InitialDeposit)
      activeProcedureNumber = load(Procedures, '0')
      activeProcedure = load(Procedures, activeProcedureNumber)
  
      if (txGovSubmitProposal.InitialDeposit < activeProcedure.MinDeposit) then  
        // MinDeposit is not reached
        
        proposal.VotingStartBlock = -1
        proposal.InitTotalVotingPower = 0
        proposal.InitProcedureNumber = -1
      
      else  
        // MinDeposit is reached
        
        proposal.VotingStartBlock = CurrentBlock
        proposal.InitTotalVotingPower = TotalVotingPower
        proposal.InitProcedureNumber = ActiveProcedureNumber
        
        for each validator in CurrentBondedValidators
          // Store voting power of each bonded validator

          validatorGovInfo = NewValidatorGovInfo()
          validatorGovInfo.InitVotingPower = validator.VotingPower
          validatorGovInfo.Minus = 0

          store(ValidatorGovInfos, <proposalID>:<validator.GovPubKey>, validatorGovInfo)
        
        ProposalProcessingQueue.push(proposalID)
  
      store(Proposals, proposalID, proposal) // Store proposal in Proposals mapping
      return proposalID
```

### Deposit

Once a proposal is submitted, if 
`Proposal.Deposit < ActiveProcedure.MinDeposit`, Atom holders can send 
`TxGovDeposit` transactions to increase the proposal's deposit.

```go
type TxGovDeposit struct {
  ProposalID    int64   // ID of the proposal
  Deposit       int64   // Number of Atoms to add to the proposal's deposit
}
```

**State modifications:**
* Decrease balance of sender by `deposit`
* Initialize or increase `deposit` of sender in `Deposits`
* Increase `proposal.Deposit` by sender's `deposit`
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
  
  else
    proposal = load(Proposals, txGovDeposit.ProposalID)

    if (proposal == nil) then  
      // There is no proposal for this proposalID
      
      throw
    
    else
      if (txGovDeposit.Deposit <= 0) OR (sender.AtomBalance < txGovDeposit.Deposit)
        // deposit is negative or null OR sender has insufficient funds
        
        throw
      
      else        
        activeProcedureNumber = load(Procedures, '0')
        activeProcedure = load(Procedures, activeProcedureNumber)
        if (proposal.Deposit >= activeProcedure.MinDeposit) then  
          // MinDeposit was reached
          
          throw
        
        else
          if (CurrentBlock >= proposal.SubmitBlock + activeProcedure.MaxDepositPeriod) then 
            // Maximum deposit period reached
            
            throw
          
          else
            // sender can deposit
            
            sender.AtomBalance -= txGovDeposit.Deposit
            deposit = load(Deposits, <txGovDeposit.ProposalID>:<sender>)

            if (deposit == nil)
              // sender has never deposited on this proposal 

              store(Deposits, <txGovDeposit.ProposalID>:<sender>, deposit)

            else
              // sender has already deposited on this proposal

              newDeposit = deposit + txGovDeposit.Deposit
              store(Deposits, <txGovDeposit.ProposalID>:<sender>, newDeposit)

            proposal.Deposit += txGovDeposit.Deposit
            
            if (proposal.Deposit >= activeProcedure.MinDeposit) then  
              // MinDeposit is reached, vote opens
              
              proposal.VotingStartBlock = CurrentBlock
              proposal.InitTotalVotingPower = TotalVotingPower
              proposal.InitProcedureNumber = ActiveProcedureNumber
              
              for each validator in CurrentBondedValidators
                // Store voting power of each bonded validator

                validatorGovInfo = NewValidatorGovInfo()
                validatorGovInfo.InitVotingPower = validator.VotingPower
                validatorGovInfo.Minus = 0

                store(ValidatorGovInfos, <proposalID>:<validator.GovPubKey>, validatorGovInfo)
              
              ProposalProcessingQueue.push(txGovDeposit.ProposalID)  
```

### Claim deposit

Finally, if the proposal is accepted or `MinDeposit` was not reached before the 
end of the `MaximumDepositPeriod`, then Atom holders can send 
`TxGovClaimDeposit` transaction to claim their deposits.

```go
  type TxGovClaimDeposit struct {
    ProposalID  int64
  }
```

**State modifications:**
If conditions are met, reimburse the deposit, i.e.
* Increase `AtomBalance` of sender by `deposit`
* Set `deposit` of sender in `DepositorsList` to 0

And the associated pseudocode.

```go
  // PSEUDOCODE //
  /* Check if TxGovClaimDeposit is valid. If vote never started and MaxDepositPeriod is reached or if vote started and        proposal was accepted, return deposit */
  
  upon receiving txGovClaimDeposit from sender do
    // check if proposal is correctly formatted. Includes fee payment.    
    
    if !correctlyFormatted(txGovClaimDeposit) then  
      throw
      
    else 
      proposal = load(Proposals, txGovDeposit.ProposalID)

      if (proposal == nil) then  
        // There is no proposal for this proposalID
        
        throw
        
      else 
        deposit = load(Deposits, <txGovClaimDeposit.ProposalID>:<sender>)
        
        if (deposit == nil)
          // sender has not deposited on this proposal

          throw
          
        else          
          if (deposit <= 0)
            // deposit has already been claimed
            
            throw
            
          else            
            if (proposal.VotingStartBlock <= 0)
            // Vote never started
              
              activeProcedureNumber = load(Procedures, '0')
              activeProcedure = load(Procedures, ActiveProcedureNumber)
              if (CurrentBlock <= proposal.SubmitBlock + activeProcedure.MaxDepositPeriod)
                // MaxDepositPeriod is not reached
                
                throw
                
              else
                //  MaxDepositPeriod is reached 
                //  Set sender's deposit to 0 and refund

                store(Deposits, <txGovClaimDeposit.ProposalID>:<sender>, 0)
                sender.AtomBalance += deposit
                
            else
              // Vote started
                
              initProcedure = load(Procedures, proposal.InitProcedureNumber)
              yesVotes = load(Votes, <txGovClaimDeposi.ProposalIDt>:<'Yes'>)
              noVotes = load(Votes, <txGovClaimDeposit.ProposalID>:<'No'>)
              noWithVetoVotes = load(Votes, <txGovClaimDeposit.ProposalID>:<'NoWithVeto'>)
              
              if  (yesVotes/proposal.InitTotalVotingPower >= 2/3) OR ((CurrentBlock > proposal.VotingStartBlock + initProcedure.VotingPeriod) AND (noWithVetoVotes/(yesVotes+noVotes+noWithVetoVotes)) < 1/3) AND (yesVotes/(yesVotes+noVotes+noWithVetoVotes)) > 1/2)) then
                
                // Proposal was accepted either because
                // pecial condition was met OR
                // Voting period ended and vote satisfies threshold
                
                store(Deposits, <txGovClaimDeposit.ProposalID>:<sender>, 0)
                sender.AtomBalance += deposit
```

### Vote

Once `ActiveProcedure.MinDeposit` is reached, voting period starts. From there,
bonded Atom holders are able to send `TxGovVote` transactions to cast their 
vote on the proposal.

```go
  type TxGovVote struct {
    ProposalID           int64           //  proposalID of the proposal
    Option               string          //  option from OptionSet chosen by the voter
    ValidatorPubKey      crypto.PubKey   //  PubKey of the validator voter wants to tie its vote to
  }
```

**State modifications:**
* If sender is not a validator and validator has not voted, initialize or 
  increase minus of validator by sender's `voting power`
* If sender is not a validator and validator has voted, decrease 
  `Votes['validatorOption']` by sender's `voting power`
* If sender is not a validator, increase `Votes['txGovVote.Option']` 
  by sender's `voting power`
* If sender is a validator, increase `Votes['txGovVote.Option']` by 
  validator's `InitialVotingPower - minus` (`minus` can be equal to 0)

Votes need to be tied to a validator in order to compute validator's voting 
power. If a delegator is bonded to multiple validators, it will have to send 
one transaction per validator (the UI should facilitate this so that multiple 
transactions can be sent in one "vote flow"). If the sender is the validator 
itself, then it will input its own GovernancePubKey as `ValidatorPubKey`

Next is a pseudocode proposal of the way `TxGovVote` transactions can be 
handled:

```go
  // PSEUDOCODE //
  // Check if TxGovVote is valid. If it is, count vote//
  
  upon receiving txGovVote from sender do
    // check if proposal is correctly formatted. Includes fee payment.    
    
    if !correctlyFormatted(txGovDeposit) then  
      throw
    
    else   
      proposal = load(Proposals, txGovDeposit.ProposalID)

      if (proposal == nil) then  
        // There is no proposal for this proposalID
        
        throw
      
      else
        initProcedure = load(Procedures, proposal.InitProcedureNumber) // get procedure that was active when vote opened
        validator = load(Validators, txGovVote.ValidatorPubKey)
      
        if  !initProcedure.OptionSet.includes(txGovVote.Option) OR 
            (validator == nil) then 
         
          // Throws if
          // Option is not in Option Set of procedure that was active when vote opened OR if
          // ValidatorPubKey is not the GovPubKey of a current validator
          
          throw
          
        else
           option = load(Options, <txGovVote.ProposalID>:<sender>:<txGovVote.ValidatorPubKey>)

           if (option != nil)
            // sender has already voted with the Atoms bonded to ValidatorPubKey

            throw

           else
            yesVotes = load(Votes, <txGovVote.ProposalID>:<'Yes'>)

            if  (proposal.VotingStartBlock < 0) OR  
                (CurrentBlock > proposal.VotingStartBlock + initProcedure.VotingPeriod) OR 
                (proposal.VotingStartBlock < lastBondingBlock(sender, txGovVote.ValidatorPubKey) OR   
                (proposal.VotingStartBlock < lastUnbondingBlock(sender, txGovVote.ValidatorPubKey) OR   
                (yesVotes/proposal.InitTotalVotingPower >= 2/3) then   

                // Throws if
                // Vote has not started OR if
                // Vote had ended OR if
                // sender bonded Atoms to ValidatorPubKey after start of vote OR if
                // sender unbonded Atoms from ValidatorPubKey after start of vote OR if
                // special condition is met, i.e. proposal is accepted and closed

              throw     

            else
              validatorGovInfo = load(ValidatorGovInfos, <txGovVote.ProposalID>:<validator.ValidatorGovPubKey>)

              if (validatorGovInfo == nil)
                // validator became validator after proposal entered voting period 

                throw

              else
                // sender can vote, check if sender == validator and store sender's option in Options
                
                store(Options, <txGovVote.ProposalID>:<sender>:<txGovVote.ValidatorPubKey>, txGovVote.Option)

                if (sender != validator.GovPubKey)
                  // Here, sender is not the Governance PubKey of the validator whose PubKey is txGovVote.ValidatorPubKey

                  if sender does not have bonded Atoms to txGovVote.ValidatorPubKey then
                    // check in Staking module

                    throw

                  else
                    validatorOption = load(Options, <txGovVote.ProposalID>:<sender>:<txGovVote.ValidatorPubKey)

                    if (validatorOption == nil)
                      // Validator has not voted already

                      validatorGovInfo.Minus += sender.bondedAmounTo(txGovVote.ValidatorPubKey)
                      store(ValidatorGovInfos, <txGovVote.ProposalID>:<validator.ValidatorGovPubKey>, validatorGovInfo)

                    else
                      // Validator has already voted
                      // Reduce votes of option chosen by validator by sender's bonded Amount

                      validatorVotes = load(Votes, <txGovVote.ProposalID>:<'validatorOption'>)
                      store(Votes, <txGovVote.ProposalID>:<'validatorOption'>, validatorVotes - sender.bondedAmountTo(txGovVote.ValidatorPubKey))

                    // increase votes of option chosen by sender by bonded Amount
                    optionVotes = load(Votes, <txGovVote.ProposalID>:<txGovVote.Option>)
                    store(Votes,<txGovVote.ProposalID>:<txGovVote.Option>, optionVotes + sender.bondedAmountTo(txGovVote.ValidatorPubKey))
                    

                else 
                  // sender is the Governance PubKey of the validator whose main PubKey is txGovVote.ValidatorPubKey
                  // i.e. sender == validator

                  optionVotes = load(Votes, <txGovVote.ProposalID>:<txGovVote.Option>)
                  store(Votes,<txGovVote.ProposalID>:<txGovVote.Option>, optionVotes + (validatorGovInfo.InitVotingPower - validatorGovInfo.Minus))
```