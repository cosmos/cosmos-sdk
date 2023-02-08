package v1beta1

import (
	"fmt"

	"github.com/cosmos/gogoproto/proto"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Governance message types and routes
const (
	TypeMsgDeposit        = "deposit"
	TypeMsgVote           = "vote"
	TypeMsgVoteWeighted   = "weighted_vote"
	TypeMsgSubmitProposal = "submit_proposal"
)

var (
	_, _, _, _ sdk.Msg = &MsgSubmitProposal{}, &MsgDeposit{}, &MsgVote{}, &MsgVoteWeighted{}

	_ codectypes.UnpackInterfacesMessage = &MsgSubmitProposal{}
)

// NewMsgSubmitProposal creates a new MsgSubmitProposal.
func NewMsgSubmitProposal(content Content, initialDeposit sdk.Coins, proposer sdk.AccAddress) (*MsgSubmitProposal, error) {
	m := &MsgSubmitProposal{
		InitialDeposit: initialDeposit,
		Proposer:       proposer.String(),
	}
	err := m.SetContent(content)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// GetInitialDeposit returns the initial deposit of MsgSubmitProposal.
func (m *MsgSubmitProposal) GetInitialDeposit() sdk.Coins { return m.InitialDeposit }

// GetContent returns the content of MsgSubmitProposal.
func (m *MsgSubmitProposal) GetContent() Content {
	content, ok := m.Content.GetCachedValue().(Content)
	if !ok {
		return nil
	}
	return content
}

// SetInitialDeposit sets the given initial deposit for MsgSubmitProposal.
func (m *MsgSubmitProposal) SetInitialDeposit(coins sdk.Coins) {
	m.InitialDeposit = coins
}

// SetProposer sets the given proposer address for MsgSubmitProposal.
func (m *MsgSubmitProposal) SetProposer(address fmt.Stringer) {
	m.Proposer = address.String()
}

// SetContent sets the content for MsgSubmitProposal.
func (m *MsgSubmitProposal) SetContent(content Content) error {
	msg, ok := content.(proto.Message)
	if !ok {
		return fmt.Errorf("can't proto marshal %T", msg)
	}
	any, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return err
	}
	m.Content = any
	return nil
}

// ValidateBasic implements the sdk.Msg interface.
func (m MsgSubmitProposal) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Proposer); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid proposer address: %s", err)
	}
	if !m.InitialDeposit.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, m.InitialDeposit.String())
	}
	if m.InitialDeposit.IsAnyNegative() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, m.InitialDeposit.String())
	}

	for _, coin := range m.InitialDeposit {
		if coin.Denom != "inj" {
			continue
		}
		minDepositAmount, ok := sdk.NewIntFromString("50000000000000000000") // 50 INJ
		if !ok {
			// should never happen, just defensive programming
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "invalid minDepositAmount amount: %s", minDepositAmount.String())
		}
		if coin.Amount.LT(minDepositAmount) {
			return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, "proposals require a minDepositAmount of at least: %s", minDepositAmount.String())
		}
	}

	content := m.GetContent()
	if content == nil {
		return sdkerrors.Wrap(types.ErrInvalidProposalContent, "missing content")
	}
	if !IsValidProposalType(content.ProposalType()) {
		return sdkerrors.Wrap(types.ErrInvalidProposalType, content.ProposalType())
	}
	if err := content.ValidateBasic(); err != nil {
		return err
	}

	return nil
}

// String implements the Stringer interface
func (m MsgSubmitProposal) String() string {
	out, _ := yaml.Marshal(m)
	return string(out)
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m MsgSubmitProposal) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var content Content
	return unpacker.UnpackAny(m.Content, &content)
}

// NewMsgDeposit creates a new MsgDeposit instance
func NewMsgDeposit(depositor sdk.AccAddress, proposalID uint64, amount sdk.Coins) *MsgDeposit {
	return &MsgDeposit{proposalID, depositor.String(), amount}
}

// NewMsgVote creates a message to cast a vote on an active proposal
func NewMsgVote(voter sdk.AccAddress, proposalID uint64, option VoteOption) *MsgVote {
	return &MsgVote{proposalID, voter.String(), option}
}

// NewMsgVoteWeighted creates a message to cast a vote on an active proposal.
func NewMsgVoteWeighted(voter sdk.AccAddress, proposalID uint64, options WeightedVoteOptions) *MsgVoteWeighted {
	return &MsgVoteWeighted{proposalID, voter.String(), options}
}
