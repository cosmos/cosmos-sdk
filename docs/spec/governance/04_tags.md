# Tags

The governance module emits the following events/tags:

## EndBlocker

| Key               | Value                                                      |
|-------------------|------------------------------------------------------------|
| `proposal-result` | `proposal-passed`\|`proposal-rejected`\|`proposal-dropped` |

## Handlers

### MsgSubmitProposal

| Key                       | Value                    |
|---------------------------|--------------------------|
| `action`                  | `submit_proposal`        |
| `category`                | `governance`             |
| `sender`                  | {proposerAccountAddress} |
| `proposal-id`             | {proposalID}             |
| `voting-period-start` [0] | {proposalID}             |

* [0] Tag only emitted if the voting period starts during the submission.

### MsgVote

| Key           | Value                 |
|---------------|-----------------------|
| `action`      | `vote`                |
| `category`    | `governance`          |
| `sender`      | {voterAccountAddress} |
| `proposal-id` | {proposalID}          |

### MsgDeposit

| Key           | Value                     |
|---------------|---------------------------|
| `action`      | `deposit`                 |
| `category`    | `governance`              |
| `sender`      | {depositorAccountAddress} |
| `proposal-id` | {proposalID}              |
