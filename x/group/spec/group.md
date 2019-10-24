# Group Specification

## Types

### `Member`

```go
// Member specifies the and power of a group member
type Member struct {
	// The address of a group member. Can be another group or a contract
	Address sdk.AccAddress `json:"address"`
	// The integral power of this member with respect to other members
	Power sdk.Int `json:"power"`
}
```

### `DecisionProtocol`

```go
// DecisionProtocol allows for flexibility in decision policy based both on
// weights (the tally of yes, no, abstain, and veto votes) and votingDuration -
// the amount of time that has been allowed for voting. A zero votingDuration
// means this proposal is being executed as basically a single multi-sig
// transaction with no voting window. This may be okay for some DecisionProtocol's.
// Other protocols may stipulate that at least a few hours or days of voting
// must occur to pass a proposal.
type DecisionProtocol interface {
	Allow(tally Tally, totalPower sdk.Int, votingDuration time.Duration)
}
```

### `Tally`

```go
type Tally struct {
	YesCount sdk.Int
	NoCount sdk.Int
	AbstainCount sdk.Int
	VetoCount sdk.Int
}
```

## Messages

### `MsgCreateGroup`

```go
type MsgCreateGroup struct {
	// The Owner of the group is allowed to change the group structure. A group account
	// can own a group in order for the group to be able to manage its own members
	Owner  sdk.AccAddress `json:"owner"`
	// The members of the group and their associated power
	Members []Member `json:"members,omitempty"`
	Description string `json:"Description,omitempty"`
    DecisionProtocol DecisionProtocol `json:"decision_protocol"`
}
```

*Returns:* `sdk.AccAddress` based on an auto-incrementing `uint64`.

### `MsgGroupExec`

```go
// MsgGroupExec executed the provided messages using the groups account if the
// provided signers pass the group's DecisionProtocol. This is essentially a
// basic multi-signature execution method.
type MsgGroupExec struct {
    Group sdk.AccAddress `json:"group"`
    Signers []sdk.AccAddress `json:"signers"`
    Msgs []sdk.Msg `json:"msgs"`
    // Policy specifies the policy by which this propose will pass or fail.
    // If set to DefaultGroupPolicy, the group's root DecisionProtocol is used. If
    // it is set to another GroupPolicyID for this group, that policy's
    // DecisionProtocol and its Capability grants will be used.
	Policy   GroupPolicyID `json:"policy"`
}
```

### `MsgUpdateGroupMembers`

```go
// MsgUpdateGroupMembers updates the members of the group, adding, removing,
// and updating members as needed. To remove an existing member set its Power to 0.
type MsgUpdateGroupMembers struct {
	Owner  sdk.AccAddress `json:"owner"`
	Group  sdk.AccAddress `json:"group"`
	Members []Member `json:"members,omitempty"`
}
```

### `MsgUpdateGroupOwner`

```go
type MsgUpdateGroupOwner struct {
	Owner  sdk.AccAddress `json:"owner"`
	Group  sdk.AccAddress `json:"group"`
	NewOwner  sdk.AccAddress `json:"new_owner"`
}
```

### `MsgUpdateGroupDecisionProtocol`

```go
type MsgUpdateGroupDecisionProtocol struct {
	Owner  sdk.AccAddress `json:"owner"`
	Group  sdk.AccAddress `json:"group"`
    DecisionProtocol DecisionProtocol `json:"decision_protocol"`
}
```

### `MsgUpdateGroupDescription`

```go
type MsgUpdateGroupDescription struct {
	Owner  sdk.AccAddress `json:"owner"`
	Group  sdk.AccAddress `json:"group"`
    DecisionProtocol DecisionProtocol `json:"decision_protocol"`
}
```

## Keeper

### Constructor: ` NewKeeper(groupStoreKey sdk.StoreKey, cdc *codec.Codec, accountKeeper auth.AccountKeeper, dispatcher msg_delegation.Keeper)`

The group keeper gets a reference to the `auth.AccountKeeper` in order to create
accounts for new groups, and a reference to the `msg_delegation.Keeper` in order
to authorize messages send back to the router.

### Query Methods

```go
type GroupKeeper interface {
  IterateGroupsByMember(member sdk.Address, fn func (group sdk.AccAddress) (stop bool))
  IterateGroupsByOwner(member sdk.Address, fn func (group sdk.AccAddress) (stop bool))
  GetGroupDescription(group sdk.AccAddress) string
  GetTotalPower(group sdk.AccAddress) sdk.Int
  GetDecisionProtocol(group sdk.AccAddress) DecisionProtocol
}
```
