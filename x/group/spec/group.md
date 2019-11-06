# Group Specification

## Types

### `GroupID`

```go
// GroupID is the auto-generated ID of the group
type GroupID uint64 
```


### `Member`

```go
// Member specifies the and power of a group member
type Member struct {
	// The address of a group member. Can be another group or a contract
	Address sdk.AccAddress `json:"address"`
	// The integral weight of this member with respect to other members
	Weight sdk.Int `json:"weight"`
	Description string `json:"description"`
}
```

## Messages

### `MsgCreateGroup`

```go
type MsgCreateGroup struct {
	// The admin of the group is allowed to change the group structure. A group account
	// can own a group in order for the group to be able to manage its own members
	Admin  sdk.AccAddress `json:"admin"`
	// The members of the group and their associated power
	Members []Member `json:"members,omitempty"`
	Description string `json:"description,omitempty"`
}
```

*Returns:* `GroupID` based on an auto-incrementing `uint64`.

### `MsgUpdateGroup

```go
// MsgUpdateGroupMembers updates the members of the group, adding, removing,
// and updating members as needed. To remove an existing member set its Power to 0.
type MsgUpdateGroupMembers struct {
	Admin  sdk.AccAddress `json:"admin"`
	Group  GroupID `json:"group"`
    // NewAdmin sets a new admin for the group. If this is left empty, the
    // current admin is not changed.
	NewAdmin  sdk.AccAddress `json:"new_admin"`
    // Description sets a new description if the string point is non-nil,
    // otherwise the description isn't changed
	Description *string `json:"description,omitempty"`
	MemberUpdates []Member `json:"member_updates,omitempty"`
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
  IterateGroupMembers(group GroupID, fn func (member sdk.AccAddress, weight sdk.Int, description string) (stop bool))
  IterateGroupsByMember(member sdk.Address, fn func (group GroupID) (stop bool))
  IterateGroupsByAdmin(member sdk.Address, fn func (group Group) (stop bool))
  GetGroupDescription(group GroupID) string
  GetTotalWeight(group GroupID) sdk.Int
}
```
