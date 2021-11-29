<!--
order: 7
-->

# Client

## CLI

A user can query and interact with the `gov` module using the CLI.

### Query

The `query` commands allow users to query `gov` state.

```bash
simd query gov --help
```

#### deposit

The `deposit` command allows users to query a deposit for a given proposal from a given depositor.

```bash
simd query gov deposit [proposal-id] [depositer-addr] [flags]
```

Example:

```bash
simd query gov deposit 1 cosmos1..
```

Example Output:

```bash
amount:
- amount: "100"
  denom: stake
depositor: cosmos1..
proposal_id: "1"
```

#### deposits

The `deposits` command allows users to query all deposits for a given proposal.

```bash
simd query gov deposits [proposal-id] [flags]
```

Example:

```bash
simd query gov deposits 1
```

Example Output:

```bash
deposits:
- amount:
  - amount: "100"
    denom: stake
  depositor: cosmos1..
  proposal_id: "1"
pagination:
  next_key: null
  total: "0"
```

#### param

The `param` command allows users to query a given parameter for the `gov` module.

```bash
simd query gov param [param-type] [flags]
```

Example:

```bash
simd query gov param voting
```

Example Output:

```bash
voting_period: "172800000000000"
```

#### params

The `params` command allows users to query all parameters for the `gov` module.

```bash
simd query gov params [flags]
```

Example:

```bash
simd query gov params
```

Example Output:

```bash
deposit_params:
  max_deposit_period: "172800000000000"
  min_deposit:
  - amount: "10000000"
    denom: stake
tally_params:
  quorum: "0.334000000000000000"
  threshold: "0.500000000000000000"
  veto_threshold: "0.334000000000000000"
voting_params:
  voting_period: "172800000000000"
```

#### proposal

The `proposal` command allows users to query a given proposal.

```bash
simd query gov proposal [proposal-id] [flags]
```

Example:

```bash
simd query gov proposal 1
```

Example Output:

```bash
content:
  '@type': /cosmos.gov.v1beta1.TextProposal
  description: testing, testing, 1, 2, 3
  title: Test Proposal
deposit_end_time: "2021-09-17T23:36:18.254995423Z"
final_tally_result:
  abstain: "0"
  "no": "0"
  no_with_veto: "0"
  "yes": "0"
proposal_id: "1"
status: PROPOSAL_STATUS_DEPOSIT_PERIOD
submit_time: "2021-09-15T23:36:18.254995423Z"
total_deposit:
- amount: "100"
  denom: stake
voting_end_time: "0001-01-01T00:00:00Z"
voting_start_time: "0001-01-01T00:00:00Z"
```

#### proposals

The `proposals` command allows users to query all proposals with optional filters.

```bash
simd query gov proposals [flags]
```

Example:

```bash
simd query gov proposals
```

Example Output:

```bash
pagination:
  next_key: null
  total: "1"
proposals:
- content:
    '@type': /cosmos.gov.v1beta1.TextProposal
    description: testing, testing, 1, 2, 3
    title: Test Proposal
  deposit_end_time: "2021-09-17T23:36:18.254995423Z"
  final_tally_result:
    abstain: "0"
    "no": "0"
    no_with_veto: "0"
    "yes": "0"
  proposal_id: "1"
  status: PROPOSAL_STATUS_DEPOSIT_PERIOD
  submit_time: "2021-09-15T23:36:18.254995423Z"
  total_deposit:
  - amount: "100"
    denom: stake
  voting_end_time: "0001-01-01T00:00:00Z"
  voting_start_time: "0001-01-01T00:00:00Z"
```

#### proposer

The `proposer` command allows users to query the proposer for a given proposal.

```bash
simd query gov proposer [proposal-id] [flags]
```

Example:

```bash
simd query gov proposer 1
```

Example Output:

```bash
proposal_id: "1"
proposer: cosmos1r0tllwu5c9dtgwg3wr28lpvf76hg85f5zmh9l2
```

#### tally

The `tally` command allows users to query the tally of a given proposal vote.

```bash
simd query gov tally [proposal-id] [flags]
```

Example:

```bash
simd query gov tally 1
```

Example Output:

```bash
abstain: "0"
"no": "0"
no_with_veto: "0"
"yes": "1"
```

#### vote

The `vote` command allows users to query a vote for a given proposal.

```bash
simd query gov vote [proposal-id] [voter-addr] [flags]
```

Example:

```bash
simd query gov vote 1 cosmos1..
```

Example Output:

```bash
option: VOTE_OPTION_YES
options:
- option: VOTE_OPTION_YES
  weight: "1.000000000000000000"
proposal_id: "1"
voter: cosmos1..
```

#### votes

The `votes` command allows users to query all votes for a given proposal.

```bash
simd query gov votes [proposal-id] [flags]
```

Example:

```bash
simd query gov votes 1
```

Example Output:

```bash
pagination:
  next_key: null
  total: "0"
votes:
- option: VOTE_OPTION_YES
  options:
  - option: VOTE_OPTION_YES
    weight: "1.000000000000000000"
  proposal_id: "1"
  voter: cosmos1r0tllwu5c9dtgwg3wr28lpvf76hg85f5zmh9l2
```

### Transactions

The `tx` commands allow users to interact with the `gov` module.

```bash
simd tx gov --help
```

#### deposit

The `deposit` command allows users to deposit tokens for a given proposal.

```bash
simd tx gov deposit [proposal-id] [deposit] [flags]
```

Example:

```bash
simd tx gov deposit 1 10000000stake --from cosmos1..
```

#### submit-proposal

The `submit-proposal` command allows users to submit a governance proposal and to optionally include an initial deposit.

```bash
simd tx gov submit-proposal [command] [flags]
```

Example:

```bash
simd tx gov submit-proposal --title="Test Proposal" --description="testing, testing, 1, 2, 3" --type="Text" --deposit="10000000stake" --from cosmos1..
```

Example (`cancel-software-upgrade`):

```bash
simd tx gov submit-proposal cancel-software-upgrade --title="Test Proposal" --description="testing, testing, 1, 2, 3" --deposit="10000000stake" --from cosmos1..
```

Example (`community-pool-spend`):

```bash
simd tx gov submit-proposal community-pool-spend proposal.json --from cosmos1..
```

```json
{
  "title": "Test Proposal",
  "description": "testing, testing, 1, 2, 3",
  "recipient": "cosmos1..",
  "amount": "10000000stake",
  "deposit": "10000000stake"
}
```

Example (`param-change`):

```bash
simd tx gov submit-proposal param-change proposal.json --from cosmos1..
```

```json
{
  "title": "Test Proposal",
  "description": "testing, testing, 1, 2, 3",
  "changes": [
    {
      "subspace": "staking",
      "key": "MaxValidators",
      "value": 100
    }
  ],
  "deposit": "10000000stake"
}
```

Example (`software-upgrade`):

```bash
simd tx gov submit-proposal software-upgrade v2 --title="Test Proposal" --description="testing, testing, 1, 2, 3" --upgrade-height 1000000 --from cosmos1..
```

#### vote

The `vote` command allows users to submit a vote for a given governance proposal.

```bash
simd tx gov vote [command] [flags]
```

Example:

```bash
simd tx gov vote 1 yes --from cosmos1..
```

#### weighted-vote

The `weighted-vote` command allows users to submit a weighted vote for a given governance proposal.

```bash
simd tx gov weighted-vote [proposal-id] [weighted-options]
```

Example:

```bash
simd tx gov weighted-vote 1 yes=0.5,no=0.5 --from cosmos1
```

## gRPC

A user can query the `gov` module using gRPC endpoints.

### Proposal

The `Proposal` endpoint allows users to query a given proposal.

```bash
cosmos.gov.v1beta1.Query/Proposal
```

Example:

```bash
grpcurl -plaintext \
    -d '{"proposal_id":"1"}' \
    localhost:9090 \
    cosmos.gov.v1beta1.Query/Proposal
```

Example Output:

```bash
{
  "proposal": {
    "proposalId": "1",
    "content": {"@type":"/cosmos.gov.v1beta1.TextProposal","description":"testing, testing, 1, 2, 3","title":"Test Proposal"},
    "status": "PROPOSAL_STATUS_VOTING_PERIOD",
    "finalTallyResult": {
      "yes": "0",
      "abstain": "0",
      "no": "0",
      "noWithVeto": "0"
    },
    "submitTime": "2021-09-16T19:40:08.712440474Z",
    "depositEndTime": "2021-09-18T19:40:08.712440474Z",
    "totalDeposit": [
      {
        "denom": "stake",
        "amount": "10000000"
      }
    ],
    "votingStartTime": "2021-09-16T19:40:08.712440474Z",
    "votingEndTime": "2021-09-18T19:40:08.712440474Z"
  }
}
```

### Proposals

The `Proposals` endpoint allows users to query all proposals with optional filters.

```bash
cosmos.gov.v1beta1.Query/Proposals
```

Example:

```bash
grpcurl -plaintext \
    localhost:9090 \
    cosmos.gov.v1beta1.Query/Proposals
```

Example Output:

```bash
{
  "proposals": [
    {
      "proposalId": "1",
      "content": {"@type":"/cosmos.gov.v1beta1.TextProposal","description":"testing, testing, 1, 2, 3","title":"Test Proposal"},
      "status": "PROPOSAL_STATUS_VOTING_PERIOD",
      "finalTallyResult": {
        "yes": "0",
        "abstain": "0",
        "no": "0",
        "noWithVeto": "0"
      },
      "submitTime": "2021-09-16T19:40:08.712440474Z",
      "depositEndTime": "2021-09-18T19:40:08.712440474Z",
      "totalDeposit": [
        {
          "denom": "stake",
          "amount": "10000000"
        }
      ],
      "votingStartTime": "2021-09-16T19:40:08.712440474Z",
      "votingEndTime": "2021-09-18T19:40:08.712440474Z"
    },
    {
      "proposalId": "2",
      "content": {"@type":"/cosmos.upgrade.v1beta1.CancelSoftwareUpgradeProposal","description":"Test Proposal","title":"testing, testing, 1, 2, 3"},
      "status": "PROPOSAL_STATUS_DEPOSIT_PERIOD",
      "finalTallyResult": {
        "yes": "0",
        "abstain": "0",
        "no": "0",
        "noWithVeto": "0"
      },
      "submitTime": "2021-09-17T18:26:57.866854713Z",
      "depositEndTime": "2021-09-19T18:26:57.866854713Z",
      "votingStartTime": "0001-01-01T00:00:00Z",
      "votingEndTime": "0001-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "total": "2"
  }
}
```

### Vote

The `Vote` endpoint allows users to query a vote for a given proposal.

```bash
cosmos.gov.v1beta1.Query/Vote
```

Example:

```bash
grpcurl -plaintext \
    -d '{"proposal_id":"1","voter":"cosmos1.."}' \
    localhost:9090 \
    cosmos.gov.v1beta1.Query/Vote
```

Example Output:

```bash
{
  "vote": {
    "proposalId": "1",
    "voter": "cosmos1..",
    "option": "VOTE_OPTION_YES",
    "options": [
      {
        "option": "VOTE_OPTION_YES",
        "weight": "1000000000000000000"
      }
    ]
  }
}
```

### Votes

The `Votes` endpoint allows users to query all votes for a given proposal.

```bash
cosmos.gov.v1beta1.Query/Votes
```

Example:

```bash
grpcurl -plaintext \
    -d '{"proposal_id":"1"}' \
    localhost:9090 \
    cosmos.gov.v1beta1.Query/Votes
```

Example Output:

```bash
{
  "votes": [
    {
      "proposalId": "1",
      "voter": "cosmos1..",
      "option": "VOTE_OPTION_YES",
      "options": [
        {
          "option": "VOTE_OPTION_YES",
          "weight": "1000000000000000000"
        }
      ]
    }
  ],
  "pagination": {
    "total": "1"
  }
}
```

### Params

The `Params` endpoint allows users to query all parameters for the `gov` module.

<!-- TODO: #10197 Querying governance params outputs nil values -->

```bash
cosmos.gov.v1beta1.Query/Params
```

Example:

```bash
grpcurl -plaintext \
    -d '{"params_type":"voting"}' \
    localhost:9090 \
    cosmos.gov.v1beta1.Query/Params
```

Example Output:

```bash
{
  "votingParams": {
    "votingPeriod": "172800s"
  },
  "depositParams": {
    "maxDepositPeriod": "0s"
  },
  "tallyParams": {
    "quorum": "MA==",
    "threshold": "MA==",
    "vetoThreshold": "MA=="
  }
}
```

### Deposit

The `Deposit` endpoint allows users to query a deposit for a given proposal from a given depositor.

```bash
cosmos.gov.v1beta1.Query/Deposit
```

Example:

```bash
grpcurl -plaintext \
    '{"proposal_id":"1","depositor":"cosmos1.."}' \
    localhost:9090 \
    cosmos.gov.v1beta1.Query/Deposit
```

Example Output:

```bash
{
  "deposit": {
    "proposalId": "1",
    "depositor": "cosmos1..",
    "amount": [
      {
        "denom": "stake",
        "amount": "10000000"
      }
    ]
  }
}
```

### deposits

The `Deposits` endpoint allows users to query all deposits for a given proposal.

```bash
cosmos.gov.v1beta1.Query/Deposits
```

Example:

```bash
grpcurl -plaintext \
    -d '{"proposal_id":"1"}' \
    localhost:9090 \
    cosmos.gov.v1beta1.Query/Deposits
```

Example Output:

```bash
{
  "deposits": [
    {
      "proposalId": "1",
      "depositor": "cosmos1..",
      "amount": [
        {
          "denom": "stake",
          "amount": "10000000"
        }
      ]
    }
  ],
  "pagination": {
    "total": "1"
  }
}
```

### TallyResult

The `TallyResult` endpoint allows users to query the tally of a given proposal.

```bash
cosmos.gov.v1beta1.Query/TallyResult
```

Example:

```bash
grpcurl -plaintext \
    -d '{"proposal_id":"1"}' \
    localhost:9090 \
    cosmos.gov.v1beta1.Query/TallyResult
```

Example Output:

```bash
{
  "tally": {
    "yes": "1000000",
    "abstain": "0",
    "no": "0",
    "noWithVeto": "0"
  }
}
```

## REST

A user can query the `gov` module using REST endpoints.

### proposal

The `proposals` endpoint allows users to query a given proposal.

```bash
/cosmos/gov/v1beta1/proposals/{proposal_id}
```

Example:

```bash
curl localhost:1317/cosmos/gov/v1beta1/proposals/1
```

Example Output:

```bash
{
  "proposal": {
    "proposal_id": "1",
    "content": {
      "@type": "/cosmos.gov.v1beta1.TextProposal",
      "title": "Test Proposal",
      "description": "testing, testing, 1, 2, 3"
    },
    "status": "PROPOSAL_STATUS_VOTING_PERIOD",
    "final_tally_result": {
      "yes": "0",
      "abstain": "0",
      "no": "0",
      "no_with_veto": "0"
    },
    "submit_time": "2021-09-16T19:40:08.712440474Z",
    "deposit_end_time": "2021-09-18T19:40:08.712440474Z",
    "total_deposit": [
      {
        "denom": "stake",
        "amount": "10000000"
      }
    ],
    "voting_start_time": "2021-09-16T19:40:08.712440474Z",
    "voting_end_time": "2021-09-18T19:40:08.712440474Z"
  }
}
```

### proposals

The `proposals` endpoint also allows users to query all proposals with optional filters.

```bash
/cosmos/gov/v1beta1/proposals
```

Example:

```bash
curl localhost:1317/cosmos/gov/v1beta1/proposals
```

Example Output:

```bash
{
  "proposals": [
    {
      "proposal_id": "1",
      "content": {
        "@type": "/cosmos.gov.v1beta1.TextProposal",
        "title": "Test Proposal",
        "description": "testing, testing, 1, 2, 3"
      },
      "status": "PROPOSAL_STATUS_VOTING_PERIOD",
      "final_tally_result": {
        "yes": "0",
        "abstain": "0",
        "no": "0",
        "no_with_veto": "0"
      },
      "submit_time": "2021-09-16T19:40:08.712440474Z",
      "deposit_end_time": "2021-09-18T19:40:08.712440474Z",
      "total_deposit": [
        {
          "denom": "stake",
          "amount": "10000000"
        }
      ],
      "voting_start_time": "2021-09-16T19:40:08.712440474Z",
      "voting_end_time": "2021-09-18T19:40:08.712440474Z"
    },
    {
      "proposal_id": "2",
      "content": {
        "@type": "/cosmos.upgrade.v1beta1.CancelSoftwareUpgradeProposal",
        "title": "Test Proposal",
        "description": "testing, testing, 1, 2, 3"
      },
      "status": "PROPOSAL_STATUS_DEPOSIT_PERIOD",
      "final_tally_result": {
        "yes": "0",
        "abstain": "0",
        "no": "0",
        "no_with_veto": "0"
      },
      "submit_time": "2021-09-17T18:26:57.866854713Z",
      "deposit_end_time": "2021-09-19T18:26:57.866854713Z",
      "total_deposit": [
      ],
      "voting_start_time": "0001-01-01T00:00:00Z",
      "voting_end_time": "0001-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "2"
  }
}
```

### voter vote

The `votes` endpoint allows users to query a vote for a given proposal.

```bash
/cosmos/gov/v1beta1/proposals/{proposal_id}/votes/{voter}
```

Example:

```bash
curl localhost:1317/cosmos/gov/v1beta1/proposals/1/votes/cosmos1..
```

Example Output:

```bash
{
  "vote": {
    "proposal_id": "1",
    "voter": "cosmos1..",
    "option": "VOTE_OPTION_YES",
    "options": [
      {
        "option": "VOTE_OPTION_YES",
        "weight": "1.000000000000000000"
      }
    ]
  }
}
```

### votes

The `votes` endpoint allows users to query all votes for a given proposal.

```bash
/cosmos/gov/v1beta1/proposals/{proposal_id}/votes
```

Example:

```bash
curl localhost:1317/cosmos/gov/v1beta1/proposals/1/votes
```

Example Output:

```bash
{
  "votes": [
    {
      "proposal_id": "1",
      "voter": "cosmos1..",
      "option": "VOTE_OPTION_YES",
      "options": [
        {
          "option": "VOTE_OPTION_YES",
          "weight": "1.000000000000000000"
        }
      ]
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "1"
  }
}
```

### params

The `params` endpoint allows users to query all parameters for the `gov` module.

<!-- TODO: #10197 Querying governance params outputs nil values -->

```bash
/cosmos/gov/v1beta1/params/{params_type}
```

Example:

```bash
curl localhost:1317/cosmos/gov/v1beta1/params/voting
```

Example Output:

```bash
{
  "voting_params": {
    "voting_period": "172800s"
  },
  "deposit_params": {
    "min_deposit": [
    ],
    "max_deposit_period": "0s"
  },
  "tally_params": {
    "quorum": "0.000000000000000000",
    "threshold": "0.000000000000000000",
    "veto_threshold": "0.000000000000000000"
  }
}
```

### deposits

The `deposits` endpoint allows users to query a deposit for a given proposal from a given depositor.

```bash
/cosmos/gov/v1beta1/proposals/{proposal_id}/deposits/{depositor}
```

Example:

```bash
curl localhost:1317/cosmos/gov/v1beta1/proposals/1/deposits/cosmos1..
```

Example Output:

```bash
{
  "deposit": {
    "proposal_id": "1",
    "depositor": "cosmos1..",
    "amount": [
      {
        "denom": "stake",
        "amount": "10000000"
      }
    ]
  }
}
```

### proposal deposits

The `deposits` endpoint allows users to query all deposits for a given proposal.

```bash
/cosmos/gov/v1beta1/proposals/{proposal_id}/deposits
```

Example:

```bash
curl localhost:1317/cosmos/gov/v1beta1/proposals/1/deposits
```

Example Output:

```bash
{
  "deposits": [
    {
      "proposal_id": "1",
      "depositor": "cosmos1..",
      "amount": [
        {
          "denom": "stake",
          "amount": "10000000"
        }
      ]
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "1"
  }
}
```

### tally

The `tally` endpoint allows users to query the tally of a given proposal.

```bash
/cosmos/gov/v1beta1/proposals/{proposal_id}/tally
```

Example:

```bash
curl localhost:1317/cosmos/gov/v1beta1/proposals/1/tally
```

Example Output:

```bash
{
  "tally": {
    "yes": "1000000",
    "abstain": "0",
    "no": "0",
    "no_with_veto": "0"
  }
}
```
