<!--
order: 4
-->

# Events

The governance module emits the following events:

## EndBlocker

| Type              | Attribute Key   | Attribute Value  |
| ----------------- | --------------- | ---------------- |
| inactive_proposal | proposal_id     | {proposalID}     |
| inactive_proposal | proposal_result | {proposalResult} |
| active_proposal   | proposal_id     | {proposalID}     |
| active_proposal   | proposal_result | {proposalResult} |

## Handlers

### MsgSubmitProposal

| Type                | Attribute Key       | Attribute Value |
| ------------------- | ------------------- | --------------- |
| submit_proposal     | proposal_id         | {proposalID}    |
| submit_proposal [0] | voting_period_start | {proposalID}    |
| proposal_deposit    | amount              | {depositAmount} |
| proposal_deposit    | proposal_id         | {proposalID}    |
| message             | module              | governance      |
| message             | action              | submit_proposal |
| message             | sender              | {senderAddress} |

- [0] Event only emitted if the voting period starts during the submission.

### MsgVote

| Type          | Attribute Key | Attribute Value |
| ------------- | ------------- | --------------- |
| proposal_vote | option        | {voteOption}    |
| proposal_vote | proposal_id   | {proposalID}    |
| message       | module        | governance      |
| message       | action        | vote            |
| message       | sender        | {senderAddress} |

### MsgDeposit

| Type                 | Attribute Key       | Attribute Value |
| -------------------- | ------------------- | --------------- |
| proposal_deposit     | amount              | {depositAmount} |
| proposal_deposit     | proposal_id         | {proposalID}    |
| proposal_deposit [0] | voting_period_start | {proposalID}    |
| message              | module              | governance      |
| message              | action              | deposit         |
| message              | sender              | {senderAddress} |

- [0] Event only emitted if the voting period starts during the submission.
