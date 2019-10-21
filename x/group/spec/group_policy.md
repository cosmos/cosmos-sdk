# Group Policy Specification

## Types

### `GroupPolicyID`

```go
type GroupPolicy uint64 
```

The ID of a group policy, generated as an auto-incrementing integer.

## Messages

### `MsgCreateGroupPolicy`
```go
// MsgCreateGroupPolicy creates a new group policy with the provided
// DecisionProtocol. A group policy allows groups to set different decision
// protocols for different types of action.
type MsgCreateGroupPolicy struct {
	Owner          sdk.AccAddress `json:"owner"`
	Group          sdk.AccAddress `json:"group"`
	DecisionProtocol DecisionProtocol `json:"decision_protocol"`
	Description string `json:"Description,omitempty"`
}
```

### `MsgGroupPolicyGrant`

The `Capability` interface is defined in the `msg_delegation` module. All of
the capabilities that can be used for generic msg delegation can be used with
group policies as well.

```go
// MsgGroupPolicyGrant grants the group policy the provided capability
// with the provided expiration.
type MsgGroupPolicyGrant struct {
	Owner          sdk.AccAddress `json:"owner"`
    Policy         GroupPolicyID
    Capability     msg_delegation.Capability    
	Expiration time.Time      `json:"expiration"`
}
```

### `MsgGroupPolicyRevoke`

```go
// MsgGroupPolicyRevoke revokes any capability of the provided msg type that
// has been granted to the provided policy.
type MsgGroupPolicyRevoke struct {
	Owner          sdk.AccAddress `json:"owner"`
    Policy         GroupPolicyID
    MsgType sdk.Msg        `json:"msg_type"`
}
```

### Update messages

```go
type MsgUpdateGroupPolicy struct {
	Owner  sdk.AccAddress `json:"owner"`
    Policy         GroupPolicyID
    DecisionPolicy DecisionPolicy `json:"decision_policy"`
}

type MsgUpdateGroupPolicyDescription struct {
	Owner  sdk.AccAddress `json:"owner"`
    Policy         GroupPolicyID
    DecisionPolicy DecisionPolicy `json:"decision_policy"`
}

type MsgDeleteGroupPolicy struct {
	Owner  sdk.AccAddress `json:"owner"`
    Policy         GroupPolicyID
}
```

## Keeper

### Query Methods

```go
type GroupKeeper interface {
  IteratePoliciesByGroup(group sdk.Address, fn func (policy GroupPolicyID) (stop bool))
  GetPolicyDescription(policy GroupPolicyID) string
  GetPolicyDecisionProtocol(policy GroupPolicyID) DecisionProtocol
}
```
