<!--
order: 7
-->

# Client

## CLI

A user can query and interact with the `gov` module using the CLI.

### Query

The `query` commands allow users to query `gov` state.

```
simd query gov --help
```

#### deposit

The `deposit` command allows users to query a deposit for a given proposal from a given depositor.

```
simd query gov deposit [proposal-id] [depositer-addr] [flags]
```

Example:

```
simd query gov deposit 1 cosmos1..
```

Example Output:

```
amount:
- amount: "100"
  denom: stake
depositor: cosmos1..
proposal_id: "1"
```

#### deposits

The `deposits` command allows users to query all deposits for a given proposal.

```
simd query gov deposits [proposal-id] [flags]
```

Example:

```
simd query gov deposits 1
```

Example Output:

```
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

```
simd query gov param [param-type] [flags]
```

Example:

```
simd query gov param voting
```

Example Output:

```
voting_period: "172800000000000"
```

#### params

The `params` command allows users to query all parameters for the `gov` module.

```
simd query gov params [flags]
```

Example:

```
simd query gov params
```

Example Output:

```
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

```
simd query gov proposal [proposal-id] [flags]
```

Example:

```
simd query gov proposal 1
```

Example Output:

```
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

```
simd query gov proposals [flags]
```

Example:

```
simd query gov proposals
```

Example Output:

```
pagination:
  next_key: null
  total: "0"
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

```
simd query gov proposer [proposal-id] [flags]
```

Example:

```
simd query gov proposer 1
```

Example Output:

```
proposal_id: "1"
proposer: cosmos1r0tllwu5c9dtgwg3wr28lpvf76hg85f5zmh9l2
```

#### tally

The `tally` command allows users to query the tally of a given proposal vote.

```
simd query gov tally [proposal-id] [flags]
```

Example:

```
simd query gov tally 1
```

Example Output:

```
abstain: "0"
"no": "0"
no_with_veto: "0"
"yes": "1"
```

#### vote

The `vote` command allows users to query a vote for a given proposal.

```
simd query gov vote [proposal-id] [voter-addr] [flags]
```

Example:

```
simd query gov vote 1 cosmos1..
```

Example Output:

```
option: VOTE_OPTION_YES
options:
- option: VOTE_OPTION_YES
  weight: "1.000000000000000000"
proposal_id: "1"
voter: cosmos1..
```

#### votes

The `votes` command allows users to query all votes for a given proposal.

```
simd query gov votes [proposal-id] [flags]
```

Example:

```
simd query gov votes 1
```

Example Output:

```
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

```
simd tx gov --help
```

#### deposit

The `deposit` command allows users to deposit tokens for a given proposal.

```
simd tx gov deposit [proposal-id] [deposit] [flags]
```

Example:

```
simd tx gov deposit 1 100stake --from cosmos1..
```

#### submit-proposal

The `submit-proposal` command allows users to submit a governance proposal and to optionally include an initial deposit.

```
simd tx gov submit-proposal [command] [flags]
```

Example:

```
simd tx gov submit-proposal --title="Test Proposal" --description="testing, testing, 1, 2, 3" --type="Text" --deposit="100stake" --from cosmos1..
```

Example (`cancel-software-upgrade`):

```
```

Example (`community-pool-spend`):

```
```

Example (`param-change`):

```
```

Example (`software-upgrade`):

```
```

#### vote

The `vote` command allows users to submit a vote for a given governance proposal.

```
simd tx gov vote [command] [flags]
```

Example:

```
simd tx gov vote 1 yes --from cosmos1..
```

#### weighted-vote

The `weighted-vote` command allows users to submit a weighted vote for a given governance proposal.

```
simd tx gov weighted-vote [proposal-id] [weighted-options]
```

Example:

```
simd tx gov weighted-vote 1 yes=0.5,no=0.5 --from cosmos1
```

## gRPC

A user can query the `gov` module using gRPC endpoints.

### Proposal

The `Proposal` endpoint allows users to query...

```
```

Example:

```
```

Example Output:

```
```

### Proposals

The `Proposals` endpoint allows users to query...

```
```

Example:

```
```

Example Output:

```
```

### Vote

The `Vote` endpoint allows users to query...

```
```

Example:

```
```

Example Output:

```
```

### Votes

The `Votes` endpoint allows users to query...

```
```

Example:

```
```

Example Output:

```
```

### Params

The `Params` endpoint allows users to query...

```
```

Example:

```
```

Example Output:

```
```

### Deposit

The `Deposit` endpoint allows users to query...

```
```

Example:

```
```

Example Output:

```
```

### Deposits

The `Deposits` endpoint allows users to query...

```
```

Example:

```
```

Example Output:

```
```

### TallyResult

The `TallyResult` endpoint allows users to query...

```
```

Example:

```
```

Example Output:

```
```
