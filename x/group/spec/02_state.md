<!--
order: 2
-->

# State

The `group` module uses the `orm` package which provides table storage with support for
primary keys and secondary indexes. `orm` also defines `Sequence` which is a persistent unique key generator based on a counter that can be used along with `Table`s.

Here's the list of tables and associated sequences and indexes stored as part of the `group` module.

## Group Table

The `groupTable` stores `GroupInfo`: `0x0 | BigEndian(GroupId) -> ProtocolBuffer(GroupInfo)`.

### groupSeq

The value of `groupSeq` is incremented when creating a new group and corresponds to the new `GroupId`: `0x1 | 0x1 -> BigEndian`.

The second `0x1` corresponds to the ORM `sequenceStorageKey`.

### groupByAdminIndex

`groupByAdminIndex` allows to retrieve groups by admin address:
`0x2 | len([]byte(group.Admin)) | []byte(group.Admin) | BigEndian(GroupId) -> []byte()`.

## Group Member Table

The `groupMemberTable` stores `GroupMember`s: `0x10 | BigEndian(GroupId) | []byte(member.Address) -> ProtocolBuffer(GroupMember)`.

The `groupMemberTable` is a primary key table and its `PrimaryKey` is given by
`BigEndian(GroupId) | []byte(member.Address)` which is used by the following indexes.

### groupMemberByGroupIndex

`groupMemberByGroupIndex` allows to retrieve group members by group id:
`0x11 | BigEndian(GroupId) | PrimaryKey -> []byte()`.

### groupMemberByMemberIndex

`groupMemberByMemberIndex` allows to retrieve group members by member address:
`0x12 | len([]byte(member.Address)) | []byte(member.Address) | PrimaryKey -> []byte()`.

## Group Policy Table

The `groupPolicyTable` stores `GroupPolicyInfo`: `0x20 | len([]byte(Address)) | []byte(Address) -> ProtocolBuffer(GroupPolicyInfo)`.

The `groupPolicyTable` is a primary key table and its `PrimaryKey` is given by
`len([]byte(Address)) | []byte(Address)` which is used by the following indexes.

### groupPolicySeq

The value of `groupPolicySeq` is incremented when creating a new group policy and is used to generate the new group policy account `Address`:
`0x21 | 0x1 -> BigEndian`.

The second `0x1` corresponds to the ORM `sequenceStorageKey`.

### groupPolicyByGroupIndex

`groupPolicyByGroupIndex` allows to retrieve group policies by group id:
`0x22 | BigEndian(GroupId) | PrimaryKey -> []byte()`.

### groupPolicyByAdminIndex

`groupPolicyByAdminIndex` allows to retrieve group policies by admin address:
`0x23 | len([]byte(Address)) | []byte(Address) | PrimaryKey -> []byte()`.

## Proposal Table

The `proposalTable` stores `Proposal`s: `0x30 | BigEndian(ProposalId) -> ProtocolBuffer(Proposal)`.

### proposalSeq

The value of `proposalSeq` is incremented when creating a new proposal and corresponds to the new `ProposalId`: `0x31 | 0x1 -> BigEndian`.

The second `0x1` corresponds to the ORM `sequenceStorageKey`.

### proposalByGroupPolicyIndex

`proposalByGroupPolicyIndex` allows to retrieve proposals by group policy account address:
`0x32 | len([]byte(account.Address)) | []byte(account.Address) | BigEndian(ProposalId) -> []byte()`.

### ProposalsByVotingPeriodEndIndex

`proposalsByVotingPeriodEndIndex` allows to retrieve proposals sorted by chronological `voting_period_end`:
`0x33 | sdk.FormatTimeBytes(proposal.VotingPeriodEnd) | BigEndian(ProposalId) -> []byte()`.

This index is used when tallying the proposal votes at the end of the voting period, and for pruning proposals at `VotingPeriodEnd + MaxExecutionPeriod`.

## Vote Table

The `voteTable` stores `Vote`s: `0x40 | BigEndian(ProposalId) | []byte(voter.Address) -> ProtocolBuffer(Vote)`.

The `voteTable` is a primary key table and its `PrimaryKey` is given by
`BigEndian(ProposalId) | []byte(voter.Address)` which is used by the following indexes.

### voteByProposalIndex

`voteByProposalIndex` allows to retrieve votes by proposal id:
`0x41 | BigEndian(ProposalId) | PrimaryKey -> []byte()`.

### voteByVoterIndex

`voteByVoterIndex` allows to retrieve votes by voter address:
`0x42 | len([]byte(voter.Address)) | []byte(voter.Address) | PrimaryKey -> []byte()`.
