# Governance documentation

*Disclaimer: This is work in progress. Mechanisms are susceptible to change.*

This document describes the high-level architecture of the governance module. The governance module allows bonded Atom holders to vote on proposals on a 1 bonded Atom 1 vote basis.

## Design overview

The governance process is divided in a few steps that are outlined below:

- **Proposal submission:** Proposal is submitted to the blockchain with a deposit 
- **Vote:** Once deposit reaches a certain value (`MinDeposit`), proposal is confirmed and vote opens. Bonded Atom holders can then send `TxGovVote` transactions to vote on the proposal
- If the proposal involves a software upgrade:
  - **Signal:** Validators start signaling that they are ready to switch to the new version
  - **Switch:** Once more than 75% of validators have signaled that they are ready to switch, their software automatically flips to the new version
  
## Proposal submission

### Right to submit a proposal

Any Atom holder, whether bonded or unbonded, can submit proposals by sending a `TxGovProposal` transaction. Once a proposal is submitted, it is identified by its unique `proposalID`.

### Proposal filter (minimum deposit)

To prevent spam, proposals must be submitted with a deposit in Atoms. Voting period will not start as long as the proposal's deposit is smaller than the minimum deposit `MinDeposit`.

When a proposal is submitted, it has to be accompagnied by a deposit that must be strictly positive but can be inferior to `MinDeposit`. Indeed, the submitter need not pay for the entire deposit on its own. If a proposal's deposit is strictly inferior to `MinDeposit`, other Atom holders can increase the proposal's deposit by sending a `TxGovDeposit` transaction. Once the proposals's deposit reaches `MinDeposit`, it enters voting period. 

### Deposit refund

There are two instances where Atom holders that deposited can claim back their deposit:
- If the proposal is accepted
- If the proposal's deposit does not reach `MinDeposit` for a period longer than `MaxDepositPeriod` (initial value: 2 months). Then the proposal is considered closed and nobody can deposit on it anymore.

In such instances, Atom holders that deposited can send a `TxGovClaimDeposit` transaction to retrieve their share of the deposit.

### Proposal types

In the initial version of the governance module, there are two types of proposal:
- `PlainTextProposal`. All the proposals that do not involve a modification of the source code go under this type. For example, an opinion poll would use a proposal of type `PlainTextProposal`
- `SoftwareUpgradeProposal`. If accepted, validators are expected to update their software in accordance with the proposal. They must do so by following a 2-steps process described in the [Software Upgrade](#software-upgrade) section below. Software upgrade roadmap may be discussed and agreed on via `PlainTextProposals`, but actual software upgrades must be performed via `SoftwareUpgradeProposals`.

### Proposal categories

There are two categories of proposal:
- `Regular`
- `Urgent`

These two categories are strictly identical except that `Urgent` proposals can be accepted faster if a certain condition is met. For more information, see [Threshold](#threshold) section.

## Vote

### Participants

*Participants* are users that have the right to vote on proposals. On the Cosmos Hub, participants are bonded Atom holders. Unbonded Atom holders and other users do not get the right to participate in governance. However, they can submit and deposit on proposals.

Note that some *participants* can be forbidden to vote on a proposal under a certain validator if:
- *participant* bonded or unbonded Atoms to said validator after proposal entered voting period
- *participant* became validator after proposal entered voting period

This does not prevent *participant* to vote with Atoms bonded to other validators. For example, if a *participant* bonded some Atoms to validator A before a proposal entered voting period and other Atoms to validator B after proposal entered voting period, only the vote under validator B will be forbidden.

### Voting period

Once a proposal reaches `MinDeposit`, it immediately enters `Voting period`. We define `Voting period` as the interval between the moment the vote opens and the moment the vote closes. `Voting period` should always be shorter than `Unbonding period` to prevent double voting. The initial value of `Voting period` is 2 weeks.

### Option set

The option set of a proposal refers to the set of choices a participant can choose from when casting its vote.

The initial option set includes the following options: 
- `Yes`
- `No`
- `NoWithVeto` 
- `Abstain` 

`NoWithVeto` counts as `No` but also adds a `Veto` vote. `Abstain` option allows voters to signal that they do not intend to vote in favor or against the proposal but accept the result of the vote. 

*Note: from the UI, for urgent proposals we should maybe add a ‘Not Urgent’ option that casts a `NoWithVeto` vote.*

### Quorum 

Quorum is defined as the minimum percentage of voting power that needs to be casted on a proposal for the result to be valid. 

In the initial version of the governance module, there will be no quorum enforced by the protocol. Participation is ensured via the combination of inheritance and validator's punishment for non-voting.

### Threshold

Threshold is defined as the minimum proportion of `Yes` votes (excluding `Abstain` votes) for the proposal to be accepted.

Initially, the threshold is set at 50% with a possibility to veto if more than 1/3rd of votes (excluding `Abstain` votes) are `NoWithVeto` votes. This means that proposals are accepted if the proportion of `Yes` votes (excluding `Abstain` votes) at the end of the voting period is superior to 50% and if the proportion of `NoWithVeto` votes is inferior to 1/3 (excluding `Abstain` votes).

`Urgent` proposals also work with the aforementioned threshold, except there is another condition that can accelerate the acceptance of the proposal. Namely, if the ratio of `Yes` votes to `InitTotalVotingPower` exceeds 2:3, `UrgentProposal` will be immediately accepted, even if the `Voting period` is not finished. `InitTotalVotingPower` is the total voting power of all bonded Atom holders at the moment when the vote opens.

### Inheritance

If a delegator does not vote, it will inherit its validator vote.

- If the delegator votes before its validator, it will not inherit from the validator's vote.
- If the delegator votes after its validator, it will override its validator vote with its own. If the proposal is a `Urgent` proposal, it is possible that the vote will close before delegators have a chance to react and override their validator's vote. This is not a problem, as `Urgent` proposals require more than 2/3rd of the total voting power to pass before the end of the voting period. If more than 2/3rd of validators collude, they can censor the votes of delegators anyway.  

### Validator’s punishment for non-voting

Validators are required to vote on all proposals to ensure that results have legitimacy. Voting is part of validators' directives and failure to do it will result in a penalty. 

If a validator’s address is not in the list of addresses that voted on a proposal and the vote is closed (i.e. `MinDeposit` was reached and `Voting period` is over), then the validator will automatically be partially slashed of `GovernancePenalty`.

*Note: Need to define values for `GovernancePenalty`*

**Exception:** If a proposal is a `Urgent` proposal and is accepted via the special condition of having a ratio of `Yes` votes to `InitTotalVotingPower` that exceeds 2:3, validators cannot be punished for not having voted on it. That is because the proposal will close as soon as the ratio exceeds 2:3, making it mechanically impossible for some validators to vote on it.

### Governance key and governance address

Validators can make use of a slot where they can designate a `Governance PubKey`. By default, a validator's `Governance PubKey` will be the same as its main PubKey. Validators can change this `Governance PubKey` by sending a `Change Governance PubKey` transaction signed by their main `Consensus PrivKey`. From there, they will be able to sign votes using the `Governance PrivKey` associated with their `Governance PubKey`. The `Governance PubKey` can be changed at any moment.


## Software Upgrade

If proposals are of type `SoftwareUpgradeProposal`, then nodes need to upgrade their software to the new version that was voted. This process is divided in two steps.

### Signal

After a `SoftwareUpgradeProposal` is accepted, validators are expected to download and install the new version of the software while continuing to run the previous version. Once a validator has downloaded and installed the upgrade, it will start signaling to the network that it is ready to switch by including the proposal's `proposalID` in its *precommits*.(*Note: Confirmation that we want it in the precommit?*)

Note: There is only one signal slot per *precommit*. If several `SoftwareUpgradeProposals` are accepted in a short timeframe, a pipeline will form and they will be implemented one after the other in the order that they were accepted.

### Switch

Once a block contains more than 2/3rd *precommits* where a common `SoftwareUpgradeProposal` is signaled, all the nodes (including validator nodes, non-validating full nodes and light-nodes) are expected to switch to the new version of the software. 

*Note: Not clear how the flip is handled programatically*


## Implementation

*Disclaimer: This is a suggestion. Only structs and pseudocode. Actual logic and implementation might widely differ*

### State

#### Procedures

`Procedures` define the rule according to which votes are run. There can only be one active procedure at any given time. If governance wants to change a procedure, either to modify a value or add/remove a parameter, a new procedure has to be created and the previous one rendered inactive.

```Go
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

**Store**:
- `Procedures`: a mapping `map[int16]Procedure` of procedures indexed by their `ProcedureNumber`
- `ActiveProcedureNumber`: returns current procedure number

#### Proposals

`Proposals` are item to be voted on. 

```Go
type Proposal struct {
  Title                 string              //  Title of the proposal
  Description           string              //  Description of the proposal
  Type                  string              //  Type of proposal. Initial set {PlainTextProposal, SoftwareUpgradeProposal}
  Category              bool                //  false=regular, true=urgent
  Deposit               int64               //  Current deposit on this proposal. Initial value is set at InitialDeposit
  SubmitBlock           int64               //  Height of the block where TxGovSubmitProposal was included
  
  VotingStartBlock      int64               //  Height of the block where MinDeposit was reached. -1 if MinDeposit is not reached
  InitTotalVotingPower  int64               //  Total voting power when proposal enters voting period (default 0)
  InitProcedureNumber   int16               //  Procedure number of the active procedure when proposal enters voting period (default -1)
  Votes                 map[string]int64    //  Votes for each option (Yes, No, NoWithVeto, Abstain)
}
```

We also introduce a type `ValidatorGovInfo`

```Go
type ValidatorGovInfo struct {
  InitVotingPower     int64   //  Voting power of validator when proposal enters voting period
  Minus               int64   //  Minus of validator, used to compute validator's voting power
}
```

**Store:**

- `Proposals`: A mapping `map[int64]Proposal` of proposals indexed by their `proposalID`
- `Deposits`: A mapping `map[[]byte]int64` of deposits indexed by `<proposalID>:<depositorPubKey>` as `[]byte`. Given a `proposalID` and a `PubKey`, returns deposit (`nil` if `PubKey` has not deposited on the proposal)
- `Options`: A mapping `map[[]byte]string` of options indexed by `<proposalID>:<voterPubKey>:<validatorPubKey>` as `[]byte`. Given a `proposalID`, a `PubKey` and a validator's `PubKey`, returns option chosen by this `PubKey` for this validator (`nil` if `PubKey` has not voted under this validator)
- `ValidatorGovInfos`: A mapping `map[[]byte]ValidatorGovInfo` of validator's governance infos indexed by `<proposalID>:<validatorGovPubKey>`. Returns `nil` if proposal has not entered voting period or if `PubKey` was not the governance public key of a validator when proposal entered voting period.


#### Proposal Processing Queue

**Store:**
- `ProposalProcessingQueue`: A queue `queue[proposalID]` containing all the `ProposalIDs` of proposals that reached `MinDeposit`. Each round, the oldest element of `ProposalProcessingQueue` is checked during `BeginBlock` to see if `CurrentBlock == VotingStartBlock + InitProcedure.VotingPeriod`. If it is, then the application checks if validators in `InitVotingPowerList` have voted and, if not, applies `GovernancePenalty`. After that proposal is ejected from `ProposalProcessingQueue` and the next element of the queue is evaluated. Note that if a proposal is urgent and accepted under the special condition, its `ProposalID` must be ejected from `ProposalProcessingQueue`.

And the pseudocode for the `ProposalProcessingQueue`:

```
  in BeginBlock do 
    
    checkProposal()  // First call of the recursive function 
    
    
  // Recursive function. First call in BeginBlock
  func checkProposal()  
    if (ProposalProcessingQueue.Peek() == nil)
      return

    else
      proposalID = ProposalProcessingQueue.Peek()
      proposal = load(store, Proposals, proposalID)
      initProcedure = load(store, Procedures, proposal.InitProcedureNumber)

      if (proposal.Category AND proposal.Votes['Yes']/proposal.InitTotalVotingPower >= 2/3)

        // proposal was urgent and accepted under the special condition
        // no punishment

        ProposalProcessingQueue.pop()
        checkProposal()

      else if (CurrentBlock == proposal.VotingStartBlock + initProcedure.VotingPeriod)

        activeProcedure = load(store, Procedures, ActiveProcedureNumber)

        for each validator in CurrentBondedValidators
          validatorGovInfo = load(store, ValidatorGovInfos, validator.GovPubKey)
          
          if (validatorGovInfo.InitVotingPower != nil)
            // validator was bonded when vote started

            validatorOption = load(store, Options, validator.GovPubKey)
            if (validatorOption == nil)
              // validator did not vote
              slash validator by activeProcedure.GovernancePenalty

        ProposalProcessingQueue.pop()
        checkProposal()        
```


### Transactions

#### Proposal Submission

Proposals can be submitted by any Atom holder via a `TxGovSubmitProposal` transaction.

```Go
type TxGovSubmitProposal struct {
  Title           string        //  Title of the proposal
  Description     string        //  Description of the proposal
  Type            string        //  Type of proposal. Initial set {PlainTextProposal, SoftwareUpgradeProposal}
  Category        bool          //  false=regular, true=urgent
  InitialDeposit  int64        //  Initial deposit paid by sender. Must be strictly positive.
}
```

**State modifications:**
- Generate new `proposalID`
- Create new `Proposal`
- Initialise `Proposals` attributes
- Store sender's deposit in `Deposits`
- Decrease balance of sender by `InitialDeposit`
- If `MinDeposit` is reached:
  - Push `proposalID` in  `ProposalProcessingQueueEnd`
  - Store each validator's voting power in `ValidatorGovInfos`

A `TxGovSubmitProposal` transaction can be handled according to the following pseudocode

```
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
      proposal.Category = txGovSubmitProposal.Category
      proposal.Deposit = txGovSubmitProposal.InitialDeposit
      proposal.SubmitBlock = CurrentBlock
      
      store(Deposits, <proposalID>:<sender>, txGovSubmitProposal.InitialDeposit)
      activeProcedure = load(store, Procedures, ActiveProcedureNumber)
  
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



#### Deposit

Once a proposal is submitted, if `Proposal.Deposit < ActiveProcedure.MinDeposit`, Atom holders can send `TxGovDeposit` transactions to increase the proposal's deposit.

```Go
type TxGovDeposit struct {
  ProposalID    int64   // ID of the proposal
  Deposit       int64   // Number of Atoms to add to the proposal's deposit
}
```

**State modifications:**
- Decrease balance of sender by `deposit`
- Initialize or increase `deposit` of sender in `Deposits`
- Increase `proposal.Deposit` by sender's `deposit`
- If `MinDeposit` is reached:
  - Push `proposalID` in  `ProposalProcessingQueueEnd`
  - Store each validator's voting power in `ValidatorGovInfos`

A `TxGovDeposit` transaction has to go through a number of checks to be valid. These checks are outlined in the following pseudocode.

```
// PSEUDOCODE //
// Check if TxGovDeposit is valid. If it is, increase deposit and check if MinDeposit is reached

upon receiving txGovDeposit from sender do
  // check if proposal is correctly formatted. Includes fee payment.
  
  if !correctlyFormatted(txGovDeposit) then  
    throw
  
  else
    proposal = load(store, Proposals, txGovDeposit.ProposalID)

    if (proposal == nil) then  
      // There is no proposal for this proposalID
      
      throw
    
    else
      if (txGovDeposit.Deposit <= 0) OR (sender.AtomBalance < txGovDeposit.Deposit)
        // deposit is negative or null OR sender has insufficient funds
        
        throw
      
      else        
        activeProcedure = load(store, Procedures, ActiveProcedureNumber)
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
            deposit = load(store, Deposits, <txGovDeposit.ProposalID>:<sender>)

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

#### Claiming deposit

Finally, if the proposal is accepted or `MinDeposit` was not reached before the end of the `MaximumDepositPeriod`, then Atom holders can send `TxGovClaimDeposit` transaction to claim their deposits.

```Go
  type TxGovClaimDeposit struct {
    ProposalID  int64
  }
```

**State modifications:**
If conditions are met, reimburse the deposit, i.e.
- Increase `AtomBalance` of sender by `deposit`
- Set `deposit` of sender in `DepositorsList` to 0

And the associated pseudocode

```
  // PSEUDOCODE //
  /* Check if TxGovClaimDeposit is valid. If vote never started and MaxDepositPeriod is reached or if vote started and        proposal was accepted, return deposit */
  
  upon receiving txGovClaimDeposit from sender do
    // check if proposal is correctly formatted. Includes fee payment.    
    
    if !correctlyFormatted(txGovClaimDeposit) then  
      throw
      
    else 
      proposal = load(store, Proposals, txGovDeposit.ProposalID)

      if (proposal == nil) then  
        // There is no proposal for this proposalID
        
        throw
        
      else 
        deposit = load(store, Deposits, <txGovClaimDeposit.ProposalID>:<sender>)
        
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
              
              activeProcedure = load(store, Procedures, ActiveProcedureNumber)
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
                
              initProcedure = load(store, Procedures, proposal.InitProcedureNumber)
              
              if  (proposal.Category AND proposal.Votes['Yes']/proposal.InitTotalVotingPower >= 2/3) OR
                  ((CurrentBlock > proposal.VotingStartBlock + initProcedure.VotingPeriod) AND (proposal.Votes['NoWithVeto']/(proposal.Votes['Yes']+proposal.Votes['No']+proposal.Votes['NoWithVeto']) < 1/3) AND (proposal.Votes['Yes']/(proposal.Votes['Yes']+proposal.Votes['No']+proposal.Votes['NoWithVeto']) > 1/2)) then
                
                // Proposal was accepted either because
                // Proposal was urgent and special condition was met
                // Voting period ended and vote satisfies threshold
                
                store(Deposits, <txGovClaimDeposit.ProposalID>:<sender>, 0)
                sender.AtomBalance += deposit

```


#### Vote

Once `ActiveProcedure.MinDeposit` is reached, voting period starts. From there, bonded Atom holders are able to send `TxGovVote` transactions to cast their vote on the proposal.

```Go
  type TxGovVote struct {
    ProposalID           int64           //  proposalID of the proposal
    Option               string          //  option from OptionSet chosen by the voter
    ValidatorPubKey      crypto.PubKey   //  PubKey of the validator voter wants to tie its vote to
  }
```

**State modifications:**
- If sender is not a validator and validator has not voted, initialize or increase minus of validator by sender's `voting power`
- If sender is not a validator and validator has voted, decrease `proposal.Votes['validatorOption']` by sender's `voting power`
- If sender is not a validator, increase `[proposal.Votes['txGovVote.Option']` by sender's `voting power`
- If sender is a validator, increase `proposal.Votes['txGovVote.Option']` by validator's `InitialVotingPower - minus` (`minus` can be equal to 0)

Votes need to be tied to a validator in order to compute validator's voting power. If a delegator is bonded to multiple validators, it will have to send one transaction per validator (the UI should facilitate this so that multiple transactions can be sent in one "vote flow"). 
If the sender is the validator itself, then it will input its own GovernancePubKey as `ValidatorPubKey`



Next is a pseudocode proposal of the way `TxGovVote` transactions can be handled:

```
  // PSEUDOCODE //
  // Check if TxGovVote is valid. If it is, count vote//
  
  upon receiving txGovVote from sender do
    // check if proposal is correctly formatted. Includes fee payment.    
    
    if !correctlyFormatted(txGovDeposit) then  
      throw
    
    else   
      proposal = load(store, Proposals, txGovDeposit.ProposalID)

      if (proposal == nil) then  
        // There is no proposal for this proposalID
        
        throw
      
      else
        initProcedure = load(store, Procedures, proposal.InitProcedureNumber) // get procedure that was active when vote opened
        validator = load(store, Validators, txGovVote.ValidatorPubKey)
      
        if  !initProcedure.OptionSet.includes(txGovVote.Option) OR 
            (validator == nil) then 
         
          // Throws if
          // Option is not in Option Set of procedure that was active when vote opened OR if
          // ValidatorPubKey is not the GovPubKey of a current validator
          
          throw
          
        else
           option = load(store, Options, <txGovVote.ProposalID>:<sender>:<txGovVote.ValidatorPubKey>)

           if (option != nil)
            // sender has already voted with the Atoms bonded to ValidatorPubKey

            throw

           else
            if  (proposal.VotingStartBlock < 0) OR  
                (CurrentBlock > proposal.VotingStartBlock + initProcedure.VotingPeriod) OR 
                (proposal.VotingStartBlock < lastBondingBlock(sender, txGovVote.ValidatorPubKey) OR   
                (proposal.VotingStartBlock < lastUnbondingBlock(sender, txGovVote.ValidatorPubKey) OR   
                (proposal.Category AND proposal.Votes['Yes']/proposal.InitTotalVotingPower >= 2/3) then   

                // Throws if
                // Vote has not started OR if
                // Vote had ended OR if
                // sender bonded Atoms to ValidatorPubKey after start of vote OR if
                // sender unbonded Atoms from ValidatorPubKey after start of vote OR if
                // proposal is urgent and special condition is met, i.e. proposal is accepted and closed

              throw     

            else
              validatorGovInfo = load(store, ValidatorGovInfos, <txGovVote.ProposalID>:<validator.ValidatorGovPubKey>)

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
                    validatorOption = load(store, Options, <txGovVote.ProposalID>:<sender>:<txGovVote.ValidatorPubKey)

                    if (validatorOption == nil)
                      // Validator has not voted already

                      validatorGovInfo.Minus += sender.bondedAmounTo(txGovVote.ValidatorPubKey)
                      store(ValidatorGovInfos, <txGovVote.ProposalID>:<validator.ValidatorGovPubKey>, validatorGovInfo)

                    else
                      // Validator has already voted
                      // Reduce votes of option chosen by validator by sender's bonded Amount

                      proposal.Votes['validatorOption'] -= sender.bondedAmountTo(txGovVote.ValidatorPubKey)

                    // increase votes of option chosen by sender by bonded Amount
                    proposal.Votes['txGovVote.Option'] += sender.bondedAmountTo(txGovVote.ValidatorPubKey)

                else 
                  // sender is the Governance PubKey of the validator whose main PubKey is txGovVote.ValidatorPubKey
                  // i.e. sender == validator
      
                  proposal.Votes['txGovVote.Option'] += (validatorGovInfo.InitVotingPower - validatorGovInfo.Minus)

                           
```


## Future improvements (not in scope for MVP)

The current documentation only describes the minimum viable product for the governance module. Future improvements may include:

- **`BountyProposals`:** If accepted, a `BountyProposal` creates an open bounty. The `BountyProposal` specifies how many Atoms will be given upon completion. These Atoms will be taken from the `reserve pool`. After a `BountyProposal` is accepted by governance, anybody can submit a `SoftwareUpgradeProposal` with the code to claim the bounty. Note that once a `BountyProposal` is accepted, the corresponding funds in the `reserve pool` are locked so that payment can always be honored. In order to link a `SoftwareUpgradeProposal` to an open bounty, the submitter of the `SoftwareUpgradeProposal` will use the `Proposal.LinkedProposal` attribute. If a `SoftwareUpgradeProposal` linked to an open bounty is accepted by governance, the funds that were reserved are automatically transferred to the submitter.
- **Complex delegation:** Delegators could choose other representatives than their validators. Ultimately, the chain of representatives would always end up to a validator, but delegators could inherit the vote of their chosen representative before they inherit the vote of their validator. In other words, they would only inherit the vote of their validator if their other appointed representative did not vote.
- **`ParameterProposals` and `WhitelistProposals`:** These proposals would automatically change pre-defined parameters and whitelists. Upon acceptance, these proposals would not require validators to do the signal and switch process.
- **Better process for proposal review:** There would be two parts to `proposal.Deposit`, one for anti-spam (same as in MVP) and an other one to reward third party auditors.
