package group

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewGenesisState creates a new genesis state with default values.
func NewGenesisState() *GenesisState {
	return &GenesisState{}
}

func (s GenesisState) Validate() error {
	groups := make(map[uint64]GroupInfo)
	groupPolicies := make(map[string]GroupPolicyInfo)
	groupPoliciesWithId := make(map[uint64]GroupPolicyInfo)
	groupMembers := make(map[uint64]GroupMember)
	proposals := make(map[uint64]Proposal)

	for _, g := range s.Groups {
		if err := g.ValidateBasic(); err != nil {
			return sdkerrors.Wrap(err, "Group validation failed")
		}
		groups[g.GroupId] = *g
	}

	for _, g := range s.GroupPolicies {

		// check that group with group policy's GroupId exists
		if _, exists := groups[g.GroupId]; !exists {
			return sdkerrors.Wrap(sdkerrors.ErrNotFound, fmt.Sprintf("group with GroupId %d doesn't exist", g.GroupId))
		}

		if err := g.ValidateBasic(); err != nil {
			return sdkerrors.Wrap(err, "GroupPolicy validation failed")
		}
		groupPolicies[g.Address] = *g
		groupPoliciesWithId[g.GroupId] = *g
	}

	for _, g := range s.GroupMembers {

		// check that group with group member's GroupId exists
		if _, exists := groups[g.GroupId]; !exists {
			return sdkerrors.Wrap(sdkerrors.ErrNotFound, fmt.Sprintf("group member with GroupId %d doesn't exist", g.GroupId))
		}

		if err := g.ValidateBasic(); err != nil {
			return sdkerrors.Wrap(err, "GroupMember validation failed")
		}
		groupMembers[g.GroupId] = *g
	}

	for _, p := range s.Proposals {

		// check that group policy with proposal address exists
		if _, exists := groupPolicies[p.Address]; !exists {
			return sdkerrors.Wrap(sdkerrors.ErrNotFound, fmt.Sprintf("group policy account with address %s doesn't correspond to proposal address", p.Address))
		}

		if err := p.ValidateBasic(); err != nil {
			return sdkerrors.Wrap(err, "Proposal validation failed")
		}
		proposals[p.ProposalId] = *p
	}

	for _, v := range s.Votes {

		if err := v.ValidateBasic(); err != nil {
			return sdkerrors.Wrap(err, "Vote validation failed")
		}

		// check that proposal exists
		if _, exists := proposals[v.ProposalId]; exists {
			proposalAddr := proposals[v.ProposalId].Address
			// check that voter is a group member corresponding to proposal id
			if _, exists := groupPolicies[proposalAddr]; exists {
				groupId := groupPolicies[proposalAddr].GroupId
				if _, exists := groups[groupId]; exists {
					groupMember := groupMembers[groupId]
					if groupMember.Member.Address == v.Voter {
						return nil
					}
					return sdkerrors.Wrap(sdkerrors.ErrNotFound, "voter is not a group member corresponding to proposal id")
				}
				return sdkerrors.Wrap(sdkerrors.ErrNotFound, "voter is not a group member corresponding to proposal id")
			}
			return sdkerrors.Wrap(sdkerrors.ErrNotFound, "voter is not a group member corresponding to proposal id")
		} else {
			return sdkerrors.Wrap(sdkerrors.ErrNotFound, fmt.Sprintf("proposal with ProposalId %d doesn't exist", v.ProposalId))
		}
	}
	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (s GenesisState) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	for _, g := range s.GroupPolicies {
		err := g.UnpackInterfaces(unpacker)
		if err != nil {
			return err
		}
	}
	for _, p := range s.Proposals {
		err := p.UnpackInterfaces(unpacker)
		if err != nil {
			return err
		}
	}
	return nil
}
