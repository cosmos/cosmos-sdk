package v1

import (
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

var (
	_, _, _, _, _ sdk.Msg                            = &MsgSubmitProposal{}, &MsgDeposit{}, &MsgVote{}, &MsgVoteWeighted{}, &MsgExecLegacyContent{}
	_, _          codectypes.UnpackInterfacesMessage = &MsgSubmitProposal{}, &MsgExecLegacyContent{}
)

// NewMsgSubmitProposal creates a new MsgSubmitProposal.
//
//nolint:interfacer
func NewMsgSubmitProposal(messages []sdk.Msg, initialDeposit sdk.Coins, proposer string, metadata string) (*MsgSubmitProposal, error) {
	m := &MsgSubmitProposal{
		InitialDeposit: initialDeposit,
		Proposer:       proposer,
		Metadata:       metadata,
	}

	anys, err := sdktx.SetMsgs(messages)
	if err != nil {
		return nil, err
	}

	m.Messages = anys

	return m, nil
}

func (m *MsgSubmitProposal) GetMsgs() ([]sdk.Msg, error) {
	return sdktx.GetMsgs(m.Messages, "sdk.MsgProposal")
}

// Route implements Msg
func (m MsgSubmitProposal) Route() string { return types.RouterKey }

// Type implements Msg
func (m MsgSubmitProposal) Type() string { return sdk.MsgTypeURL(&m) }

// ValidateBasic implements Msg
func (m MsgSubmitProposal) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Proposer); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid proposer address: %s", err)
	}

	deposit := sdk.NewCoins(m.InitialDeposit...)
	if !deposit.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, deposit.String())
	}

	if deposit.IsAnyNegative() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, deposit.String())
	}

	// Check that either metadata or Msgs length is non nil.
	if len(m.Messages) == 0 && len(m.Metadata) == 0 {
		return sdkerrors.Wrap(types.ErrNoProposalMsgs, "either metadata or Msgs length must be non-nil")
	}

	msgs, err := m.GetMsgs()
	if err != nil {
		return err
	}

	for idx, msg := range msgs {
		if err := msg.ValidateBasic(); err != nil {
			return sdkerrors.Wrap(types.ErrInvalidProposalMsg,
				fmt.Sprintf("msg: %d, err: %s", idx, err.Error()))
		}
	}

	return nil
}

// GetSignBytes implements Msg
func (m MsgSubmitProposal) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements Msg
func (m MsgSubmitProposal) GetSigners() []sdk.AccAddress {
	proposer, _ := sdk.AccAddressFromBech32(m.Proposer)
	return []sdk.AccAddress{proposer}
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m MsgSubmitProposal) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return sdktx.UnpackInterfaces(unpacker, m.Messages)
}

// NewMsgDeposit creates a new MsgDeposit instance
//
//nolint:interfacer
func NewMsgDeposit(depositor sdk.AccAddress, proposalID uint64, amount sdk.Coins) *MsgDeposit {
	return &MsgDeposit{proposalID, depositor.String(), amount}
}

// Route implements Msg
func (msg MsgDeposit) Route() string { return types.RouterKey }

// Type implements Msg
func (msg MsgDeposit) Type() string { return sdk.MsgTypeURL(&msg) }

// ValidateBasic implements Msg
func (msg MsgDeposit) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Depositor); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid depositor address: %s", err)
	}
	amount := sdk.NewCoins(msg.Amount...)
	if !amount.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, amount.String())
	}
	if amount.IsAnyNegative() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, amount.String())
	}

	return nil
}

// GetSignBytes implements Msg
func (msg MsgDeposit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements Msg
func (msg MsgDeposit) GetSigners() []sdk.AccAddress {
	depositor, _ := sdk.AccAddressFromBech32(msg.Depositor)
	return []sdk.AccAddress{depositor}
}

// NewMsgVote creates a message to cast a vote on an active proposal
//
//nolint:interfacer
func NewMsgVote(voter sdk.AccAddress, proposalID uint64, option VoteOption, metadata string) *MsgVote {
	return &MsgVote{proposalID, voter.String(), option, metadata}
}

// Route implements Msg
func (msg MsgVote) Route() string { return types.RouterKey }

// Type implements Msg
func (msg MsgVote) Type() string { return sdk.MsgTypeURL(&msg) }

// ValidateBasic implements Msg
func (msg MsgVote) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Voter); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid voter address: %s", err)
	}
	if !ValidVoteOption(msg.Option) {
		return sdkerrors.Wrap(types.ErrInvalidVote, msg.Option.String())
	}

	return nil
}

// GetSignBytes implements Msg
func (msg MsgVote) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements Msg
func (msg MsgVote) GetSigners() []sdk.AccAddress {
	voter, _ := sdk.AccAddressFromBech32(msg.Voter)
	return []sdk.AccAddress{voter}
}

// NewMsgVoteWeighted creates a message to cast a vote on an active proposal
//
//nolint:interfacer
func NewMsgVoteWeighted(voter sdk.AccAddress, proposalID uint64, options WeightedVoteOptions, metadata string) *MsgVoteWeighted {
	return &MsgVoteWeighted{proposalID, voter.String(), options, metadata}
}

// Route implements Msg
func (msg MsgVoteWeighted) Route() string { return types.RouterKey }

// Type implements Msg
func (msg MsgVoteWeighted) Type() string { return sdk.MsgTypeURL(&msg) }

// ValidateBasic implements Msg
func (msg MsgVoteWeighted) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Voter); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid voter address: %s", err)
	}
	if len(msg.Options) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, WeightedVoteOptions(msg.Options).String())
	}

	totalWeight := sdk.NewDec(0)
	usedOptions := make(map[VoteOption]bool)
	for _, option := range msg.Options {
		if !option.IsValid() {
			return sdkerrors.Wrap(types.ErrInvalidVote, option.String())
		}
		weight, err := sdk.NewDecFromStr(option.Weight)
		if err != nil {
			return sdkerrors.Wrapf(types.ErrInvalidVote, "Invalid weight: %s", err)
		}
		totalWeight = totalWeight.Add(weight)
		if usedOptions[option.Option] {
			return sdkerrors.Wrap(types.ErrInvalidVote, "Duplicated vote option")
		}
		usedOptions[option.Option] = true
	}

	if totalWeight.GT(sdk.NewDec(1)) {
		return sdkerrors.Wrap(types.ErrInvalidVote, "Total weight overflow 1.00")
	}

	if totalWeight.LT(sdk.NewDec(1)) {
		return sdkerrors.Wrap(types.ErrInvalidVote, "Total weight lower than 1.00")
	}

	return nil
}

// GetSignBytes implements Msg
func (msg MsgVoteWeighted) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements Msg
func (msg MsgVoteWeighted) GetSigners() []sdk.AccAddress {
	voter, _ := sdk.AccAddressFromBech32(msg.Voter)
	return []sdk.AccAddress{voter}
}

func NewMsgExecLegacyContent(content *codectypes.Any, authority string) *MsgExecLegacyContent {
	return &MsgExecLegacyContent{
		Content:   content,
		Authority: authority,
	}
}

func (c MsgExecLegacyContent) GetSigners() []sdk.AccAddress {
	authority, _ := sdk.AccAddressFromBech32(c.Authority)
	return []sdk.AccAddress{authority}
}

func (c MsgExecLegacyContent) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(c.Authority)
	if err != nil {
		return err
	}

	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m MsgExecLegacyContent) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var content v1beta1.Content
	return unpacker.UnpackAny(m.Content, &content)
}
