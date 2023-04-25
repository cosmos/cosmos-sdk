package v2

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/group"
)

// GroupInfo is a type alias for group.GroupInfo with the PrimaryKeyFields method
// this is required by legacy orm.
type GroupInfo struct {
	group.GroupInfo
}

func (g GroupInfo) PrimaryKeyFields() []interface{} {
	return []interface{}{g.Id}
}

// GroupPolicyInfo is a type alias for group.GroupPolicyInfo with the PrimaryKeyFields method
// this is required by legacy orm.
type GroupPolicyInfo struct {
	group.GroupPolicyInfo
}

func (g GroupPolicyInfo) PrimaryKeyFields() []interface{} {
	addr := sdk.MustAccAddressFromBech32(g.Address)

	return []interface{}{addr.Bytes()}
}

// Proposal is a type alias for group.Proposal with the PrimaryKeyFields method
// this is required by legacy orm.
type Proposal struct {
	group.Proposal
}

func (g Proposal) PrimaryKeyFields() []interface{} {
	return []interface{}{g.Id}
}

// GroupMember is a type alias for group.GroupMember with the PrimaryKeyFields method
// this is required by legacy orm.
type GroupMember struct {
	group.GroupMember
}

func (g GroupMember) PrimaryKeyFields() []interface{} {
	addr := sdk.MustAccAddressFromBech32(g.Member.Address)

	return []interface{}{g.GroupId, addr.Bytes()}
}

// Vote is a type alias for group.Vote with the PrimaryKeyFields method
// this is required by legacy orm.
type Vote struct {
	group.Vote
}

func (v Vote) PrimaryKeyFields() []interface{} {
	addr := sdk.MustAccAddressFromBech32(v.Voter)

	return []interface{}{v.ProposalId, addr.Bytes()}
}
