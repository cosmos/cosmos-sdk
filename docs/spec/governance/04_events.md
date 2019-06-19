# Tags

The governance module emits the following events/tags:

## EndBlocker

| Type              | Attribute Key   | Attribute Value  |
|-------------------|-----------------|------------------|
| inactive_proposal | proposal_id     | {proposalID}     |
| inactive_proposal | proposal_result | {proposalResult} |
| active_proposal   | proposal_id     | {proposalID}     |
| active_proposal   | proposal_result | {proposalResult} |

## Handlers

### MsgSubmitProposal

| Type                | Attribute Key       | Attribute Value |
|---------------------|---------------------|-----------------|
| submit_proposal     | sender              | {senderAddress} |
| submit_proposal     | proposal_id         | {proposalID}    |
| submit_proposal [0] | voting_period_start | {proposalID}    |
| proposal_deposit    | sender              | {senderAddress} |
| proposal_deposit    | amount              | {depositAmount} |
| proposal_deposit    | proposal_id         | {proposalID}    |
| message             | module              | governance      |
| message             | action              | submit_proposal |

* [0] Event only emitted if the voting period starts during the submission.

### MsgVote

| Type          | Attribute Key | Attribute Value |
|---------------|---------------|-----------------|
| proposal_vote | sender        | {senderAddress} |
| proposal_vote | option        | {voteOption}    |
| proposal_vote | proposal_id   | {proposalID}    |
| message       | module        | governance      |
| message       | action        | vote            |

### MsgDeposit

| Type                 | Attribute Key       | Attribute Value |
|----------------------|---------------------|-----------------|
| proposal_deposit     | sender              | {senderAddress} |
| proposal_deposit     | amount              | {depositAmount} |
| proposal_deposit     | proposal_id         | {proposalID}    |
| proposal_deposit [0] | voting_period_start | {proposalID}    |
| message              | module              | governance      |
| message              | action              | deposit         |

* [0] Event only emitted if the voting period starts during the submission.
