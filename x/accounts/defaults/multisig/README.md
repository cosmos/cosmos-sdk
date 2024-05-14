# Multisig Accounts


* [State](#state)
  * [Config](#config)
  * [Proposal](#proposal)
  * [Members](#members)
* [Methods](#methods)
    * [MsgInit](#msginit)
    * [MsgUpdateConfig](#msgupdateconfig)
    * [MsgCreateProposal](#msgcreateproposal)
    * [MsgVote](#msgvote)
    * [MsgExecuteProposal](#msgexecuteproposal)

The x/accounts/defaults/multisig module provides the implementation for multisig accounts within the x/accounts module.

## State

The multisig account keeps its members as a map of addresses and weights (`<[]byte, uint64>`), a config struct, a map of proposals and a map of votes.

```go
type Account struct {
	Members  collections.Map[[]byte, uint64]
	Sequence collections.Sequence
	Config   collections.Item[v1.Config]

	addrCodec     address.Codec
	headerService header.Service
	eventService  event.Service

	Proposals collections.Map[uint64, v1.Proposal]
	Votes     collections.Map[collections.Pair[uint64, []byte], int32] // key: proposalID + voter address
}
```

### Config

The config contains the basic rules defining how the multisig will work. All of these fields can be modified afterwards by calling [MsgUpdateConfig](#msgupdateconfig).

```protobuf
message Config {
  int64 threshold = 1;

  int64 quorum = 2;

  // voting_period is the duration in seconds for the voting period.
  int64 voting_period = 3;

  // revote defines if members can change their vote.
  bool revote = 4;

  // early_execution defines if the multisig can be executed before the voting period ends.
  bool early_execution = 5;
}
```

### Proposal

The proposal contains the title, summary, messages and the status of the proposal. The messages are stored as `google.protobuf.Any` to allow for any type of message to be stored.

```protobuf
message Proposal {
  string   title                        = 1;
  string   summary                      = 2;
  repeated google.protobuf.Any messages = 3;

  // if true, the proposal will execute as soon as the quorum is reached (last voter will execute).
  bool execute = 4;

  // voting_period_end will be set by the account when the proposal is created.
  int64 voting_period_end = 5;

  ProposalStatus status = 6;
}
```

### Members

Members are stored as a map of addresses and weights. The weight is used to determine the voting power of the member.


## Methods

### MsgInit

The `MsgInit` message initializes a multisig account with the given members and config.

```protobuf
message MsgInit {
  repeated Member members = 1;
  Config          Config  = 2;
}
```

### MsgUpdateConfig

The `MsgUpdateConfig` message updates the config of the multisig account. Only the members that are changing are required, and if their weight is 0, they are removed. If the config is nil, then it will not be updated.

```protobuf
message MsgUpdateConfig {
  // only the members that are changing are required, if their weight is 0, they are removed.
  repeated Member update_members = 1;

  // not all fields from Config can be changed
  Config Config = 2;
}
```

### MsgCreateProposal

Only a member can create a proposal. The proposal will be stored in the account and the members will be able to vote on it.

If a voting period is not set, the proposal will be created using the voting period from the config. If the proposal has a voting period, it will be used instead.

```protobuf
message MsgCreateProposal {
  Proposal proposal = 1;
}
```

### MsgVote

The `MsgVote` message allows a member to vote on a proposal. The vote can be either `Yes`, `No` or `Abstain`.

```protobuf
message MsgVote {
  uint64 proposal_id = 1;
  VoteOption   vote        = 2;
}

// VoteOption enumerates the valid vote options for a given proposal.
enum VoteOption {
  // VOTE_OPTION_UNSPECIFIED defines a no-op vote option.
  VOTE_OPTION_UNSPECIFIED = 0;
  // VOTE_OPTION_YES defines the yes proposal vote option.
  VOTE_OPTION_YES = 1;
  // VOTE_OPTION_ABSTAIN defines the abstain proposal vote option.
  VOTE_OPTION_ABSTAIN = 2;
  // VOTE_OPTION_NO defines the no proposal vote option.
  VOTE_OPTION_NO = 3;
}
```


### MsgExecuteProposal

The `MsgExecuteProposal` message allows a member to execute a proposal. The proposal must have reached the quorum and the voting period must have ended.

```protobuf
message MsgExecuteProposal {
  uint64 proposal_id = 1;
}
```