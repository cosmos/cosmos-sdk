# Governance documentation

*Disclaimer: This is work in progress. Mechanisms are susceptible to change.*

This document describes the high-level architecture of the governance module. The governance module allows bonded Atom holders to vote on proposals on a 1 bonded Atom 1 vote basis.

## Design overview

The governance process is divided in a few steps that are outlined below:

- **Proposal submission:** Proposal is submitted to the blockchain with a deposit 
- **Vote:** Once deposit reaches a certain value (`MinDeposit`), proposal is confirmed and vote opens. Bonded Atom holders can then send `TxGovVote` transactions to vote on the proposal
- If the proposal involves a software upgrade
  - **Signal:** Validator start signaling that they are ready to switch to the new version
  - **Switch:** Once more than 2/3rd validators have signaled their readiness to switch, their software automatically flips to the new version
  
## Proposal submission

### Right to submit a proposal

Any Atom holder, whether bonded or unbonded, can submit proposals by sending a `TxProposal` transaction. Once a proposal is submitted, it is identified by its unique `proposalID`.

### Proposal filter (minimum deposit)

To prevent spam, proposals must be submitted with a `deposit` in Atoms such that `0 < deposit < MinDeposit`. 
Other Atom holders can increase the proposal's deposit by sending a `TxGovDeposit` transaction. 
Once the proposals's deposit reaches `MinDeposit`, it enters voting period. 

### Deposit refund

There are two instances where Atom holders that deposited can claim back their deposit:
- If the proposal is accepted
- If the proposal's deposit does not reach `MinDeposit` for a period longer than `mMxDepositPeriod` (initial value: 2 months). Then the proposal is considered closed and nobody can deposit on it anymore.

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

*Participants* are users that have the right to vote. On the Cosmos Hub, participants are bonded Atom holders. Unbonded Atom holders and other users do not get the right to participate in governance. However, they can submit and deposit on proposals.

### Voting period

Once a proposal reaches `MinDeposit`, it immediately enters `Voting period`. We define `Voting period` as the interval between the moment the vote opens and the moment the vote closes. `Voting period` should always be shorter than `Unbonding period` to prevent double voting. The initial value of `Voting period` is 2 weeks.

### Option set

The option set of a proposal refers to the set of choices a participant can choose from when casting its vote.

The initial option set includes the following options: 
- `Yes`
- `No`
- `NoWithVeto` 
- `Abstain` 

`NoWithVeto` counts as `No` but also adds a `Veto` vote. `Abstain` allows voters to signal that they do not intend to vote in favor or against the proposal but accept the result of the vote. 

*Note: from the UI, for urgent proposals we should maybe add a ‘Not Urgent’ option that casts a `NoWithVeto` vote.*

### Quorum 

Quorum is defined as the minimum percentage of voting power that needs to be casted on a proposal for the result to be valid. 

In the initial version of the governance module, there will be no quorum enforced by the protocol. Participation is ensured via the combination of inheritance and validator's punishment for non-voting.

### Threshold

Threshold is defined as the minimum ratio of `Yes` votes to `No` votes for the proposal to be accepted.

Initially, the threshold is set at 50% with a possibility to veto if more than 1/3rd of votes (excluding `Abstain` votes) are `NoWithVeto` votes. This means that proposals are accepted is the ratio of `Yes` votes to `No` votes at the end of the voting period is superior to 50% and if the number of `NoWithVeto` votes is inferior to 1/3rd of total votes (excluding `Abstain`).

`Urgent` proposals also work with the aforementioned threshold, except there is another condition that can accelerate the acceptance of the proposal. Namely, if the ratio of `Yes` votes to `InitTotalVotingPower` exceeds 2/3, `UrgentProposal` will be immediately accepted, even if the `Voting period` is not finished. `InitTotalVotingPower` is the total voting power of all bonded Atom holders at the moment when the vote opens.

### Inheritance

If a delegator does not vote, it will inherit its validator vote.

- If the delegator votes before its validator, it will not inherit from the validator's vote.
- If the delegator votes after its validaotor, it will override its validator vote with its own vote. If the proposal is a `Urgent` proposal, it is possible that the vote will close before delegators have a chance to react and override their validator's vote. This is not a problem, as `Urgent` proposals require more than 2/3rd of the total voting power to pass before the end of the voting period. If more than 2/3rd of validators collude, they can censor the votes of delegators anyway.  

### Validator’s punishment for non-voting

Validators are required to vote on all proposals to ensure that results have legitimacy. Voting is part of validators' directives and failure to do it will result in a penalty. 

If a validator’s address is not in the list of addresses that voted on a proposal and if the vote is closed (i.e. `MinDeposit` was reached and `Voting period` is over), then this validator will automatically be partially slashed of `GovernancePenalty`.

*Note: Need to define values for `GovernancePenalty`*

**Exception:** If a proposal is a `Urgent` proposal and is accepted via the special condition of having a ratio of `Yes` votes to `InitTotalVotingPower` that exceeds 2/3, validators cannot be punished for not having voted on it. That is because the proposal will close as soon as the ratio exceeds 2/3, making it mechanically impossible for some validators to vote on it.

### Governance key and governance address

Validators can make use of an additional slot where they can designate a `Governance PubKey`. By default, a validator's `Governance PubKey` will be the same as its main PubKey. Validators can change this `Governance PubKey` by sending a `Change Governance PubKey` transaction signed by their main `Consensus PubKey`. From there, they will be able to sign vote using the `Governance PrivKey` associated with their `Governance PubKey`. The `Governance PubKey` can be changed at any moment.


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

### Procedures

`Procedures` define the rule according to which votes are run. There can only be one active procedure at any given time. If governance wants to change a procedure, either to modify a value or add/remove a parameter, a new procedure has to be created and the previous one rendered inactive.

```Go
type Procedure struct {
  VotingPeriod      int64              //  Length of the voting period. Initial value: 2 weeks
  MinDeposit        int64              //  Minimum deposit for a proposal to enter voting period. 
  OptionSet         []string            //  Options available to voters. {Yes, No, NoWithVeto, Abstain}
  ProposalTypes     []string            //  Types available to submitters. {PlainTextProposal, SoftwareUpgradeProposal}
  Threshold         int64              //  Minimum value of Yes votes to No votes ratio for proposal to pass. Initial value: 0.5
  Veto              rational.Rational   //  Minimum value of Veto votes to Total votes ratio for proposal to be vetoed. Initial value: 1/3
  MaxDepositPeriod  int64              //  Maximum period for Atom holders to deposit on a proposal. Initial value: 2 months
  GovernancePenalty int64              //  Penalty if validator does not vote
  
  ProcedureNumber   int16              //  Incremented each time a new procedure is created
  IsActive          bool                //  If true, procedure is active. Only one procedure can have isActive true.
}
```

### Proposals

`Proposals` are item to be voted on. They can be submitted by any Atom holder via a `TxGovSubmitProposal` transaction.

```Go
type TxGovSubmitProposal struct {
  Title           string        //  Title of the proposal
  Description     string        //  Description of the proposal
  Type            string        //  Type of proposal. Initial set {PlainTextProposal, SoftwareUpgradeProposal}
  Category        bool          //  false=regular, true=urgent
  InitialDeposit  int64        //  Initial deposit paid by sender. Must be strictly positive.
}

type Proposal struct {
  Title             string              //  Title of the proposal
  Description       string              //  Description of the proposal
  Type              string              //  Type of proposal. Initial set {PlainTextProposal, SoftwareUpgradeProposal}
  Category          bool                //  false=regular, true=urgent
  Deposit           int64              //  Current deposit on this proposal. Initial value is set at InitialDeposit
  SubmitBlock       int64              //  Height of the block where TxGovSubmitProposal was included
  VotingStartBlock  int64              //  Height of the block where MinDeposit was reached. -1 if MinDeposit is not reached.
  Votes             map[string]int64   //  Votes for each option (Yes, No, NoWithVeto, Abstain)
}
```

Each `Proposal` is identified by its unique `proposalID`. 

Additionaly, four lists will be linked to each proposal:
- `DepositorList`: List of addresses that deposited on the proposal with their associated deposit
- `VotersList`: List of addresses that voted **under each validator** with their associated option
- `InitVotingPowerList`: Snapshot of validators' voting power **when proposal enters voting period** (only saves validators whose voting power is >0).
- `MinusesList`: List of minuses for each validator. Used to compute validators' voting power when they cast a vote.

Two final parameters, `InitTotalVotingPower` and `InitProcedureNumber` associated with `proposalID` will be saved when proposal enters voting period.

We also introduce `ProposalProcessingQueue` which lists all the `ProposalIDs` of proposals that reached `MinDeposit` from oldest to newest. Each round, the oldest element of `ProposalProcessingQueue` is checked during `BeginBlock` to see if `CurrentBlock == VotingStartBlock + InitProcedure.VotingPeriod`. If it is, then the application checks if validators in `InitVotingPowerList` have voted and, if not, applies `GovernancePenalty`. After that proposal is ejected from `ProposalProcessingQueue` and the new first element of the queue is evaluated. Note that if a proposal is urgent and accepted under the special condition, its `ProposalID` must be ejected from `ProposalProcessingQueue`.

A `TxGovSubmitProposal` transaction can be handled according to the following pseudocode

```
// PSEUDOCODE //
// Check if TxGovSubmitProposal is valid. If it is, create proposal //

upon receiving txGovSubmitProposal from sender do
  // check if proposal is correctly formatted. Includes fee payment.
  
  if !correctlyFormatted(txGovSubmitProposal) then 
    throw
  
  else
    if (txGovSubmitProposal.InitialDeposit <= 0) OR (sender.AtomBalance < InitialDeposit) then 
      // InitialDeposit is negative or null OR sender has insufficient funds
      
      throw
    
    else
      sender.AtomBalance -= InitialDeposit
      
      proposalID = generate new proposalID
      proposal = create new Proposal from proposalID
      
      proposal.Title = txGovSubmitProposal.Title
      proposal.Description = txGovSubmitProposal.Description
      proposal.Type = txGovSubmitProposal.Type
      proposal.Category = txGovSubmitProposal.Category
      proposal.Deposit = txGovSubmitProposal.InitialDeposit
      proposal.SubmitBlock = CurrentBlock
      
      create depositorsList from proposalID
      initiate deposit of sender in depositorsList at txGovSubmitProposal.InitialDeposit
  
      if (txGovSubmitProposal.InitialDeposit < ActiveProcedure.MinDeposit) then  
        // MinDeposit is not reached
        
        proposal.VotingStartBlock = -1
      
      else  
        // MinDeposit is reached
        
        proposal.VotingStartBlock = CurrentBlock
        
        create  votersList,
                initVotingPowerList,
                minusesList,
                initProcedureNumber,
                initTotalVotingPower  from proposalID
                                    
        snapshot(ActiveProcedure.ProcedureNumber) // Save current procedure number in initProcedureNumber
        snapshot(TotalVotingPower)  // Save total voting power in initTotalVotingPower
        snapshot(ValidatorVotingPower)  // Save validators' voting power in initVotingPowerList
        
        ProposalProcessingQueueEnd++
        ProposalProcessingQueue[ProposalProcessingQueueEnd] = proposalID
  
      return proposalID
```

And the pseudocode for the `ProposalProcessingQueue`:

```
  in BeginBlock do 
    
    checkProposal()  
    
    
    
  func checkProposal()  
    if (ProposalProcessingQueueBeginning == ProposalProcessingQueueEnd)
      return

    else
      retrieve proposalID from ProposalProcessingQueue[ProposalProcessingQueueBeginning]
      retrieve proposal from proposalID
      retrieve initProcedureNumber from proposalID
      retrieve initProcedure from initProcedureNumber

      if (CurrentBlock == proposal.VotingStartBlock + initProcedure.VotingPeriod)
        retrieve initVotingPowerList from proposalID
        retrieve votersList from proposalID
        retrieve validators from initVotingPowerList

        for each validator in validators
          if validator is not in votersList
            slash validator by ActiveProcedure.GovernancePenalty

        ProposalProcessingQueueBeginning++  // ProposalProcessingQueue will have a new element
        checkProposal()

      else
        return          
```

Once a proposal is submitted, if `Proposal.Deposit < ActiveProcedure.MinDeposit`, Atom holders can send `TxGovDeposit` transactions to increase the proposal's deposit.

```Go
type TxGovDeposit struct {
  ProposalID    int64    // ID of the proposal
  Deposit       int64  // Number of Atoms to add to the proposal's deposit
}
```

A `TxGovDeposit` transaction has to go through a number of checks to be valid. These checks are outlined in the following pseudocode.

```
// PSEUDOCODE //
// Check if TxGovDeposit is valid. If it is, increase deposit and check if MinDeposit is reached

upon receiving txGovDeposit from sender do
  // check if proposal is correctly formatted. Includes fee payment.
  
  if !correctlyFormatted(txGovDeposit) then  
    throw
  
  else
    if !exist(txGovDeposit.proposalID) then  
      // There is no proposal for this proposalID
      
      throw
    
    else
      if (txGovDeposit.Deposit <= 0 OR sender.AtomBalance < txGovDeposit.Deposit)
        // deposit is negative or null OR sender has insufficient funds
        
        throw
      
      else
        retrieve proposal from txGovDeposit.ProposalID // retrieve throws if it fails
        
        if (proposal.Deposit >= ActiveProcedure.MinDeposit) then  
          // MinDeposit was reached
          
          throw
        
        else
          if (CurrentBlock >= proposal.SubmitBlock + ActiveProcedure.MaxDepositPeriod) then 
            // Maximum deposit period reached
            
            throw
          
          else
            // sender can deposit
            
            retrieve depositorsList from txGovDeposit.ProposalID
            sender.AtomBalance -= txGovDeposit.Deposit

            if sender is in depositorsList
              increase deposit of sender in depositorsList by txGovDeposit.Deposit

            else
              initialise deposit of sender in depositorsList at txGovDeposit.Deposit
            
            proposal.Deposit += txGovDeposit.Deposit
            
            if (proposal.Deposit >= ActiveProcedure.MinDeposit) then  
              // MinDeposit is reached, vote opens
              
              proposal.VotingStartBlock = CurrentBlock
              
              create  votersList,
                      initVotingPowerList,
                      minusesList,
                      initProcedureNumber,
                      initTotalVotingPower  from proposalID
              
              snapshot(ActiveProcedure.ProcedureNumber) // Save current procedure number in InitProcedureNumber
              snapshot(TotalVotingPower)  // Save total voting power in InitTotalVotingPower
              snapshot(ValidatorVotingPower)  // Save validators' voting power in InitVotingPowerList
              
              ProposalProcessingQueueEnd++  // ProposalProcessingQueue will have a new element
              ProposalProcessingQueue[ProposalProcessingQueueEnd] = txGovDeposit.ProposalID  
```

Finally, if the proposal is accepted or `MinDeposit` was not reached before the end of the `MaximumDepositPeriod`, then Atom holders can send `TxGovClaimDeposit` transaction to claim their deposits.

```Go
  type TxGovClaimDeposit struct {
    ProposalID  int64
  }
```

And the associated pseudocode

```
  // PSEUDOCODE //
  /* Check if TxGovClaimDeposit is valid. If vote never started and MaxDepositPeriod is reached or if vote started and        proposal was accepted, return deposit */
  
  upon receiving txGovClaimDeposit from sender do
    // check if proposal is correctly formatted. Includes fee payment.    
    
    if !correctlyFormatted(txGovClaimDeposit) then  
      throw
      
    else 
      if !exists(txGovClaimDeposit.ProposalID) then
        // There is no proposal for this proposalID
        
        throw
        
      else 
        retrieve depositorsList from txGovClaimDeposit.ProposalID
        
        
        if sender is not in depositorsList then
          throw
          
        else
          retrieve deposit from sender in depositorsList 
          
          if deposit <= 0
            // deposit has already been claimed
            
            throw
            
          else
            retrieve proposal from txGovClaimDeposit.ProposalID
            
            if proposal.VotingStartBlock <= 0
            // Vote never started
            
              if (CurrentBlock <= proposal.SubmitBlock + ActiveProcedure.MaxDepositPeriod)
                // MaxDepositPeriod is not reached
                
                throw
                
              else
                //  MaxDepositPeriod is reached 
                
                set deposit of sender in depositorsList to 0
                sender.AtomBalance += deposit
                
            else
              // Vote started
              
              retrieve initTotalVotingPower from txGovClaimDeposit.ProposalID
              retrieve initProcedureNumber from txGovClaimDeposit.ProposalID    
              retrieve initProcedure from initProcedureNumber // get procedure that was active when vote opened
              
              if  (proposal.Category AND proposal.Votes['Yes']/initTotalVotingPower >= 2/3) OR
                  ((CurrentBlock > proposal.VotingStartBlock + initProcedure.VotingPeriod) AND (proposal.Votes['NoWithVeto']/(proposal.Votes['Yes']+proposal.Votes['No']+proposal.Votes['NoWithVeto']) < 1/3) AND           (proposal.Votes['Yes']/(proposal.Votes['Yes']+proposal.Votes['No']+proposal.Votes['NoWithVeto']) > 1/2)) then
                
                // Proposal was accepted either because
                // Proposal was urgent and special condition was met
                // Voting period ended and vote satisfies threshold
                
                set deposit of sender in depositorsList to 0
                sender.AtomBalance += deposit
                
              else
                throw
```


### Vote

Once `ActiveProcedure.MinDeposit` is reached, voting period starts. From there, bonded Atom holders are able to send `TxGovVote` transactions to cast their vote on the proposal.

```Go
  type TxGovVote struct {
    ProposalID           int64           //  proposalID of the proposal
    Option               string          //  option from OptionSet chosen by the voter
    ValidatorPubKey      crypto.PubKey   //  PubKey of the validator voter wants to tie its vote to
  }
```

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
      if !exists(txGovVote.proposalID) OR  
         
         // Throws if
         // proposalID does not exist
         
        throw
      
      else
        retrieve initProcedureNumber from txGovVote.ProposalID    
        retrieve initProcedure from initProcedureNumber // get procedure that was active when vote opened
      
        if  !initProcedure.OptionSet.includes(txGovVote.Option) OR 
            !isValid(txGovVote.ValidatorPubKey) then 
         
          // Throws if
          // Option is not in Option Set of procedure that was active when vote opened OR if
          // ValidatorPubKey is not the GovPubKey of a current validator
          
          throw
          
        else
           retrieve votersList from txGovVote.ProposalID

           if sender is in votersList under txGovVote.ValidatorPubKey then 
            // sender has already voted with the Atoms bonded to ValidatorPubKey

            throw

           else
            retrieve proposal from txGovVote.ProposalID
            retrieve InitTotalVotingPower from txGovVote.ProposalID

            if  (proposal.VotingStartBlock < 0) OR  
                (CurrentBlock > proposal.VotingStartBlock + initProcedure.VotingPeriod) OR 
                (proposal.VotingStartBlock < lastBondingBlock(sender, txGovVote.ValidatorPubKey) OR   
                (proposal.VotingStartBlock < lastUnbondingBlock(sender, txGovVote.ValidatorPubKey) OR   
                (proposal.Category AND proposal.Votes['Yes']/InitTotalVotingPower >= 2/3) then   

                // Throws if
                // Vote has not started OR if
                // Vote had ended OR if
                // sender bonded Atoms to ValidatorPubKey after start of vote OR if
                // sender unbonded Atoms from ValidatorPubKey after start of vote OR if
                // proposal is urgent and special condition is met, i.e. proposal is accepted and closed

              throw     

            else
              // sender can vote, check if sender == validator and add sender to voter list
              
              add sender to votersList under txGovVote.ValidatorPubKey

              if (sender is not equal to GovPubKey that corresponds to txGovVote.ValidatorPubKey)
                // Here, sender is not the Governance PubKey of the validator whose PubKey is txGovVote.ValidatorPubKey

                if sender does not have bonded Atoms to txGovVote.ValidatorPubKey then
                  throw

                else
                  if txGovVote.ValidatorPubKey is not in votersList under txGovVote.ValidatorPubKey then
                    // Validator has not voted already

                    if exists(MinusesList[txGovVote.ValidatorPubKey]) then
                      // a minus already exists for this validator's PubKey, increase minus
                      // by the amount of Atoms sender has bonded to ValidatorPubKey

                      MinusesList[txGovVote.ValidatorPubKey] += sender.bondedAmountTo(txGovVote.ValidatorPubKey)

                    else
                      // a minus does not already exist for this validator's PubKey, initialise minus
                      // at the amount of Atoms sender has bonded to ValidatorPubKey

                      MinusesList[txGovVote.ValidatorPubKey] = sender.bondedAmountTo(txGovVote.ValidatorPubKey)

                  else
                    // Validator has already voted
                    // Reduce option count chosen by validator by sender's bonded Amount

                    retrieve validatorOption from votersList using txGovVote.ValidatorPubKey
                    proposal.Votes['validatorOption'] -= sender.bondedAmountTo(txGovVote.ValidatorPubKey)

                  // increase Option count chosen by sender by bonded Amount
                  proposal.Votes['txGovVote.Option'] += sender.bondedAmountTo(txGovVote.ValidatorPubKey)

              else 
                // sender is the Governance PubKey of the validator whose main PubKey is txGovVote.ValidatorPubKey
                // i.e. sender == validator

                retrieve initialVotingPower from InitVotingPowerList using txGovVote.ValidatorPubKey
                
                
                if exists(MinusesList[txGovVote.ValidatorPubKey]) then
                  // a minus exists for this validator's PubKey, decrease vote of validator by minus

                  proposal.Votes['txGovVote.Option'] += (initialVotingPower - MinusesList[txGovVote.ValidatorPubKey])

                else
                  // a minus does not exist for this validator's PubKey, validator votes with full voting power

                  proposal.Votes['txGovVote.Option'] += initialVotingPower
                  
              if (proposal.Category AND proposal.Votes['Yes']/InitTotalVotingPower >= 2/3)
                // after vote is counted, if proposal is urgent and special condition is met
                // remove proposalID from ProposalProcessingQueue
                
                remove txGovVote.ProposalID from ProposalProcessingQueue
                Rearrange ProposalProcessingQueue
                ProposalProcessingQueueEnd--            
```


## Future improvements (not in scope for MVP)

The current documentation only describes the minimum viable product for the governance module. Future improvements may include:

- **`BountyProposals`:** If accepted, a `BountyProposal` creates an open bounty. The `BountyProposal` specifies how many Atoms will be given upon completion. These Atoms will be taken from the `reserve pool`. After a `BountyProposal` is accepted by governance, anybody can submit a `SoftwareUpgradeProposal` with the code to claim the bounty. Note that once a `BountyProposal` is accepted, the corresponding funds in the `reserve pool` are locked so that payment can always be honored. In order to link a `SoftwareUpgradeProposal` to an open bounty, the submitter of the `SoftwareUpgradeProposal` will use the `Proposal.LinkedProposal` attribute. If a `SoftwareUpgradeProposal` linked to an open bounty is accepted by governance, the funds that were reserved are automatically transferred to the submitter.
- **Complex delegation:** Delegators could choose other representatives than their validators. Ultimately, the chain of representatives would always end up to a validator, but delegators could inherit the vote of their chosen representative before they inherit the vote of their validator. In other words, they would only inherit the vote of their validator if their other appointed representative did not vote.
- **`ParameterProposals` and `WhitelistProposals`:** These proposals would automatically change pre-defined parameters and whitelists. Upon acceptance, these proposals would not require validators to do the signal and switch process.
- **Better process for proposal review:** There would be two parts to `proposal.Deposit`, one for anti-spam (same as in MVP) and an other one to reward third party auditors.
