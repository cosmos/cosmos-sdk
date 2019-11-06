# Group Policy Specification

## Types

### `DecisionPolicy`

```go
// DecisionPolicy allows for flexibility in decision policy based both on
// weights (the tally of yes, no, abstain, and veto votes) and votingDuration -
// the amount of time that has been allowed for voting. A zero votingDuration
// means this proposal is being executed as basically a single multi-sig
// transaction with no voting window. This may be okay for some DecisionPolicy's.
// Other policies may stipulate that at least a few hours or days of voting
// must occur to pass a proposal.
type DecisionPolicy interface {
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

### `MsgCreateGroupAccount
```go
// MsgCreateGroupPolicy creates a new group policy with the provided
// DecisionPolicy. A group policy allows groups to set different decision
// policys for different types of action.
type MsgCreateGroupAccount struct {
	Admin          sdk.AccAddress `json:"admin"`
	Group          GroupID `json:"group"`
	DecisionPolicy DecisionPolicy `json:"decision_policy"`
	Description string `json:"Description,omitempty"`
}
```

*Returns:*  a new auto-generated `sdk.AccAddress` for this group account.

### `MsgGroupExec`

```go
// MsgGroupExec executed the provided messages using the groups account if the
// provided signers pass the group's DecisionPolicy. This is essentially a
// basic multi-signature execution method.
type MsgGroupExec struct {
    GroupAccount sdk.AccAddress `json:"group_account"`
    Signers []sdk.AccAddress `json:"signers"`
    Msgs []sdk.Msg `json:"msgs"`
}
```

### Update messages

```go
type MsgUpdateGroupAccount struct {
	Admin  sdk.AccAddress `json:"admin"`
	GroupAccount  sdk.AccAddress `json:"group_account"`
    // NewAdmin sets a new admin for the group. If this is left empty, the
    // current admin is not changed.
	NewAdmin  sdk.AccAddress `json:"new_admin"`
    // Description sets a new description if the string point is non-nil,
    // otherwise the description isn't changed
	Description *string `json:"description,omitempty"`
    // DecisionPolicy changes the GroupAccount DecisionPolicy, if left to nil,
    // the DecisionPolicy isn't changed
    DecisionPolicy DecisionPolicy `json:"decision_policy"`
}

```

