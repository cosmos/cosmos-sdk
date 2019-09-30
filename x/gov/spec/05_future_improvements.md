# Future Improvements

The current documentation only describes the minimum viable product for the 
governance module. Future improvements may include:

* **`BountyProposals`:** If accepted, a `BountyProposal` creates an open 
  bounty. The `BountyProposal` specifies how many Atoms will be given upon
  completion. These Atoms will be taken from the `reserve pool`. After a 
  `BountyProposal` is accepted by governance, anybody can submit a 
  `SoftwareUpgradeProposal` with the code to claim the bounty. Note that once a 
  `BountyProposal` is accepted, the corresponding funds in the `reserve pool` 
  are locked so that payment can always be honored. In order to link a 
  `SoftwareUpgradeProposal` to an open bounty, the submitter of the 
  `SoftwareUpgradeProposal` will use the `Proposal.LinkedProposal` attribute. 
  If a `SoftwareUpgradeProposal` linked to an open bounty is accepted by 
  governance, the funds that were reserved are automatically transferred to the
  submitter.
* **Complex delegation:** Delegators could choose other representatives than 
  their validators. Ultimately, the chain of representatives would always end 
  up to a validator, but delegators could inherit the vote of their chosen 
  representative before they inherit the vote of their validator. In other 
  words, they would only inherit the vote of their validator if their other 
  appointed representative did not vote.
* **Better process for proposal review:** There would be two parts to 
  `proposal.Deposit`, one for anti-spam (same as in MVP) and an other one to 
  reward third party auditors.
