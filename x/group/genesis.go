package group

import (
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/group/errors"
	"github.com/cosmos/cosmos-sdk/x/group/internal/math"
)

// NewGenesisState creates a new genesis state with default values.
func NewGenesisState() *GenesisState {
	return &GenesisState{}
}

func (s GenesisState) Validate() error {
	for _, f := range s.Groups {
		groupId := f.GetGroupId()
		if groupId == 0 {
			return sdkerrors.Wrap(errors.ErrEmpty, "group's group id")
		}
		_, err := sdk.AccAddressFromBech32(f.GetAdmin())
		if err != nil {
			return sdkerrors.Wrap(err, "admin")
		}
		if _, err := math.NewNonNegativeDecFromString(f.GetTotalWeight()); err != nil {
			return sdkerrors.Wrap(err, "total weight")
		}
		if f.GetVersion() == 0 {
			return sdkerrors.Wrap(errors.ErrEmpty, "version")
		}
	}

	for _, f := range s.GroupAccounts {
		_, err := sdk.AccAddressFromBech32(f.Admin)
		if err != nil {
			return sdkerrors.Wrap(err, "group account admin")
		}
		_, err = sdk.AccAddressFromBech32(f.Address)
		if err != nil {
			return sdkerrors.Wrap(err, "group account address")
		}
		if f.GroupId == 0 {
			return sdkerrors.Wrap(errors.ErrEmpty, "group account's group id")
		}
		if f.Version == 0 {
			return sdkerrors.Wrap(errors.ErrEmpty, "group account version")
		}
		policy := f.GetDecisionPolicy()
		if policy == nil {
			return sdkerrors.Wrap(errors.ErrEmpty, "group account policy")
		}
		if err := policy.ValidateBasic(); err != nil {
			return sdkerrors.Wrap(err, "group account policy")
		}
	}

	for _, f := range s.GroupMembers {
		if f.GetGroupId() == 0 {
			return sdkerrors.Wrap(errors.ErrEmpty, "group member")
		}
		err := f.GetMember().ValidateBasic()
		if err != nil {
			return sdkerrors.Wrap(err, "group member")
		}
	}

	for _, f := range s.Proposals {
		if f.ProposalId == 0 {
			return sdkerrors.Wrap(errors.ErrEmpty, "proposal id")
		}
		_, err := sdk.AccAddressFromBech32(f.Address)
		if err != nil {
			return sdkerrors.Wrap(err, "proposer group account address")
		}
		if f.GroupVersion == 0 {
			return sdkerrors.Wrap(errors.ErrEmpty, "proposal group version")
		}
		if f.GroupAccountVersion == 0 {
			return sdkerrors.Wrap(errors.ErrEmpty, "proposal group account version")
		}
		_, err = f.VoteState.GetYesCount()
		if err != nil {
			return sdkerrors.Wrap(err, "proposal VoteState yes count")
		}
		_, err = f.VoteState.GetNoCount()
		if err != nil {
			return sdkerrors.Wrap(err, "proposal VoteState no count")
		}
		_, err = f.VoteState.GetAbstainCount()
		if err != nil {
			return sdkerrors.Wrap(err, "proposal VoteState abstain count")
		}
		_, err = f.VoteState.GetVetoCount()
		if err != nil {
			return sdkerrors.Wrap(err, "proposal VoteState veto count")
		}
	}

	for _, f := range s.Votes {
		_, err := sdk.AccAddressFromBech32(f.GetVoter())
		if err != nil {
			return sdkerrors.Wrap(err, "voter")
		}
		if f.GetProposalId() == 0 {
			return sdkerrors.Wrap(errors.ErrEmpty, "voter proposal id")
		}
		if f.GetChoice() == Choice_CHOICE_UNSPECIFIED {
			return sdkerrors.Wrap(errors.ErrEmpty, "voter choice")
		}
		if _, ok := Choice_name[int32(f.GetChoice())]; !ok {
			return sdkerrors.Wrap(errors.ErrInvalid, "choice")
		}
		if f.GetSubmittedAt().IsZero() {
			return sdkerrors.Wrap(errors.ErrEmpty, "submitted at")
		}
	}
	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (s GenesisState) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	for _, g := range s.GroupAccounts {
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
