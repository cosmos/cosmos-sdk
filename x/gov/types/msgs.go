package types

import (
	"fmt"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Governance message types and routes
const (
	TypeMsgDeposit        = "deposit"
	TypeMsgVote           = "vote"
	TypeMsgVoteWeighted   = "weighted_vote"
	TypeMsgSubmitProposal = "submit_proposal"
	TypeMsgSignal         = "signal"
)

var (
	_, _, _, _ sdk.Msg                       = &MsgSubmitProposal{}, &MsgDeposit{}, &MsgVote{}, &MsgVoteWeighted{}
	_          types.UnpackInterfacesMessage = &MsgSubmitProposal{}
)

// NewMsgSubmitProposal creates a new MsgSubmitProposal.
//nolint:interfacer
func NewMsgSubmitProposal(messages []sdk.Msg, initialDeposit sdk.Coins, proposer sdk.AccAddress) (*MsgSubmitProposal, error) {
	msgsAny := make([]*types.Any, len(messages))
	for i, msg := range messages {
		any, err := types.NewAnyWithValue(msg)
		if err != nil {
			return &MsgSubmitProposal{}, err
		}

		msgsAny[i] = any
	}

	m := &MsgSubmitProposal{
		Messages:       msgsAny,
		InitialDeposit: initialDeposit,
		Proposer:       proposer.String(),
	}

	return m, nil
}

func (m *MsgSubmitProposal) GetInitialDeposit() sdk.Coins { return m.InitialDeposit }

func (m *MsgSubmitProposal) GetProposer() sdk.AccAddress {
	proposer, _ := sdk.AccAddressFromBech32(m.Proposer)
	return proposer
}

func (m *MsgSubmitProposal) GetMessages() ([]sdk.Msg, error) {
	msgs := make([]sdk.Msg, len(m.Messages))
	for i, msgAny := range m.Messages {
		msg, ok := msgAny.GetCachedValue().(sdk.Msg)
		if !ok {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "messages contains %T which is not a sdk.Msg", msgAny)
		}
		msgs[i] = msg
	}

	return msgs, nil
}

// Deprecated: Content is no longer used. Use Messages instead
func (m *MsgSubmitProposal) GetContent() Content {
	content, ok := m.Content.GetCachedValue().(Content)
	if !ok {
		return nil
	}
	return content
}

func (m *MsgSubmitProposal) SetInitialDeposit(coins sdk.Coins) {
	m.InitialDeposit = coins
}

func (m *MsgSubmitProposal) SetProposer(address fmt.Stringer) {
	m.Proposer = address.String()
}

// Deprecated: Content is no longer used. Use Messages instead
func (m *MsgSubmitProposal) SetContent(content Content) error {
	msg, ok := content.(proto.Message)
	if !ok {
		return fmt.Errorf("can't proto marshal %T", msg)
	}
	any, err := types.NewAnyWithValue(msg)
	if err != nil {
		return err
	}
	m.Content = any
	return nil
}

// Route implements Msg
func (m MsgSubmitProposal) Route() string { return RouterKey }

// Type implements Msg
func (m MsgSubmitProposal) Type() string { return TypeMsgSubmitProposal }

// ValidateBasic implements Msg
func (m MsgSubmitProposal) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Proposer); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, m.Proposer)
	}
	if !m.InitialDeposit.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, m.InitialDeposit.String())
	}
	if m.InitialDeposit.IsAnyNegative() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, m.InitialDeposit.String())
	}
	if m.Messages == nil || len(m.Messages) == 0 {
		return ErrNoProposalMsgs
	}

	if _, err := m.GetMessages(); err != nil {
		return err
	}

	return nil
}

// GetSignBytes implements Msg
func (m MsgSubmitProposal) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements Msg
func (m MsgSubmitProposal) GetSigners() []string {
	return []string{m.Proposer}
}

// String implements the Stringer interface
func (m MsgSubmitProposal) String() string {
	out, _ := yaml.Marshal(m)
	return string(out)
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m MsgSubmitProposal) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var content Content
	return unpacker.UnpackAny(m.Content, &content)
}

// NewMsgDeposit creates a new MsgDeposit instance
//nolint:interfacer
func NewMsgDeposit(depositor sdk.AccAddress, proposalID uint64, amount sdk.Coins) *MsgDeposit {
	return &MsgDeposit{proposalID, depositor.String(), amount}
}

// Route implements Msg
func (msg MsgDeposit) Route() string { return RouterKey }

// Type implements Msg
func (msg MsgDeposit) Type() string { return TypeMsgDeposit }

// ValidateBasic implements Msg
func (msg MsgDeposit) ValidateBasic() error {
	if msg.Depositor == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.Depositor)
	}
	if !msg.Amount.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.Amount.String())
	}
	if msg.Amount.IsAnyNegative() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.Amount.String())
	}

	return nil
}

// String implements the Stringer interface
func (msg MsgDeposit) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// GetSignBytes implements Msg
func (msg MsgDeposit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements Msg
func (msg MsgDeposit) GetSigners() []string {
	return []string{msg.Depositor}
}

// NewMsgVote creates a message to cast a vote on an active proposal
//nolint:interfacer
func NewMsgVote(voter sdk.AccAddress, proposalID uint64, option VoteOption) *MsgVote {
	return &MsgVote{proposalID, voter.String(), option}
}

// Route implements Msg
func (msg MsgVote) Route() string { return RouterKey }

// Type implements Msg
func (msg MsgVote) Type() string { return TypeMsgVote }

// ValidateBasic implements Msg
func (msg MsgVote) ValidateBasic() error {
	if msg.Voter == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.Voter)
	}

	if !ValidVoteOption(msg.Option) {
		return sdkerrors.Wrap(ErrInvalidVote, msg.Option.String())
	}

	return nil
}

// String implements the Stringer interface
func (msg MsgVote) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// GetSignBytes implements Msg
func (msg MsgVote) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements Msg
func (msg MsgVote) GetSigners() []string {
	return []string{msg.Voter}
}

// NewMsgVoteWeighted creates a message to cast a vote on an active proposal
//nolint:interfacer
func NewMsgVoteWeighted(voter sdk.AccAddress, proposalID uint64, options WeightedVoteOptions) *MsgVoteWeighted {
	return &MsgVoteWeighted{proposalID, voter.String(), options}
}

// Route implements Msg
func (msg MsgVoteWeighted) Route() string { return RouterKey }

// Type implements Msg
func (msg MsgVoteWeighted) Type() string { return TypeMsgVoteWeighted }

// ValidateBasic implements Msg
func (msg MsgVoteWeighted) ValidateBasic() error {
	if msg.Voter == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.Voter)
	}

	if len(msg.Options) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, WeightedVoteOptions(msg.Options).String())
	}

	totalWeight := sdk.NewDec(0)
	usedOptions := make(map[VoteOption]bool)
	for _, option := range msg.Options {
		if !ValidWeightedVoteOption(option) {
			return sdkerrors.Wrap(ErrInvalidVote, option.String())
		}
		totalWeight = totalWeight.Add(option.Weight)
		if usedOptions[option.Option] {
			return sdkerrors.Wrap(ErrInvalidVote, "Duplicated vote option")
		}
		usedOptions[option.Option] = true
	}

	if totalWeight.GT(sdk.NewDec(1)) {
		return sdkerrors.Wrap(ErrInvalidVote, "Total weight overflow 1.00")
	}

	if totalWeight.LT(sdk.NewDec(1)) {
		return sdkerrors.Wrap(ErrInvalidVote, "Total weight lower than 1.00")
	}

	return nil
}

// String implements the Stringer interface
func (msg MsgVoteWeighted) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// GetSignBytes implements Msg
func (msg MsgVoteWeighted) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements Msg
func (msg MsgVoteWeighted) GetSigners() []string {
	return []string{msg.Voter}
}

func NewMsgSignal(title, description string) *MsgSignal {
	return &MsgSignal{title, description}
}

// Route implements Msg
func (msg MsgSignal) Route() string { return RouterKey }

// Type implements Msg
func (msg MsgSignal) Type() string { return TypeMsgSignal }

// String implements the Stringer interface
func (msg MsgSignal) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// ValidateBasic implements Msg
func (msg MsgSignal) ValidateBasic() error {
	if len(strings.TrimSpace(msg.Title)) == 0 {
		return sdkerrors.Wrap(ErrInvalidSignalMsg, "title cannot be blank")
	}
	if len(msg.Title) > MaxTitleLength {
		return sdkerrors.Wrap(ErrInvalidSignalMsg, fmt.Sprintf("title is longer than max length of %d", MaxTitleLength))
	}

	if len(msg.Description) > MaxDescriptionLength {
		return sdkerrors.Wrap(ErrInvalidSignalMsg, fmt.Sprintf("description is longer than max length of %d", MaxDescriptionLength))
	}

	return nil
}

// GetSignBytes implements Msg
func (msg MsgSignal) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements Msg
func (msg *MsgSignal) GetSigners() []string {
	return []string{""}
}
