package group

import (
	"github.com/cosmos/gogoproto/proto"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	"github.com/cosmos/cosmos-sdk/x/group/codec"
)

var (
	_ sdk.Msg            = &MsgCreateGroup{}
	_ legacytx.LegacyMsg = &MsgCreateGroup{}
)

// GetSignBytes Implements Msg.
func (m MsgCreateGroup) GetSignBytes() []byte {
	return sdk.MustSortJSON(codec.ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgCreateGroup.
func (m MsgCreateGroup) GetSigners() []sdk.AccAddress {
	admin := sdk.MustAccAddressFromBech32(m.Admin)

	return []sdk.AccAddress{admin}
}

var (
	_ sdk.Msg            = &MsgUpdateGroupAdmin{}
	_ legacytx.LegacyMsg = &MsgUpdateGroupAdmin{}
)

// GetSignBytes Implements Msg.
func (m MsgUpdateGroupAdmin) GetSignBytes() []byte {
	return sdk.MustSortJSON(codec.ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgUpdateGroupAdmin.
func (m MsgUpdateGroupAdmin) GetSigners() []sdk.AccAddress {
	admin := sdk.MustAccAddressFromBech32(m.Admin)

	return []sdk.AccAddress{admin}
}

// GetGroupID gets the group id of the MsgUpdateGroupAdmin.
func (m *MsgUpdateGroupAdmin) GetGroupID() uint64 {
	return m.GroupId
}

var (
	_ sdk.Msg            = &MsgUpdateGroupMetadata{}
	_ legacytx.LegacyMsg = &MsgUpdateGroupMetadata{}
)

// GetSignBytes Implements Msg.
func (m MsgUpdateGroupMetadata) GetSignBytes() []byte {
	return sdk.MustSortJSON(codec.ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgUpdateGroupMetadata.
func (m MsgUpdateGroupMetadata) GetSigners() []sdk.AccAddress {
	admin := sdk.MustAccAddressFromBech32(m.Admin)

	return []sdk.AccAddress{admin}
}

// GetGroupID gets the group id of the MsgUpdateGroupMetadata.
func (m *MsgUpdateGroupMetadata) GetGroupID() uint64 {
	return m.GroupId
}

var (
	_ sdk.Msg            = &MsgUpdateGroupMembers{}
	_ legacytx.LegacyMsg = &MsgUpdateGroupMembers{}
)

// GetSignBytes Implements Msg.
func (m MsgUpdateGroupMembers) GetSignBytes() []byte {
	return sdk.MustSortJSON(codec.ModuleCdc.MustMarshalJSON(&m))
}

var _ sdk.Msg = &MsgUpdateGroupMembers{}

// GetSigners returns the expected signers for a MsgUpdateGroupMembers.
func (m MsgUpdateGroupMembers) GetSigners() []sdk.AccAddress {
	admin := sdk.MustAccAddressFromBech32(m.Admin)

	return []sdk.AccAddress{admin}
}

// GetGroupID gets the group id of the MsgUpdateGroupMembers.
func (m *MsgUpdateGroupMembers) GetGroupID() uint64 {
	return m.GroupId
}

var (
	_ sdk.Msg            = &MsgCreateGroupWithPolicy{}
	_ legacytx.LegacyMsg = &MsgCreateGroupWithPolicy{}

	_ types.UnpackInterfacesMessage = MsgCreateGroupWithPolicy{}
)

// NewMsgCreateGroupWithPolicy creates a new MsgCreateGroupWithPolicy.
func NewMsgCreateGroupWithPolicy(admin string, members []MemberRequest, groupMetadata, groupPolicyMetadata string, groupPolicyAsAdmin bool, decisionPolicy DecisionPolicy) (*MsgCreateGroupWithPolicy, error) {
	m := &MsgCreateGroupWithPolicy{
		Admin:               admin,
		Members:             members,
		GroupMetadata:       groupMetadata,
		GroupPolicyMetadata: groupPolicyMetadata,
		GroupPolicyAsAdmin:  groupPolicyAsAdmin,
	}
	err := m.SetDecisionPolicy(decisionPolicy)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// GetDecisionPolicy gets the decision policy of MsgCreateGroupWithPolicy.
func (m *MsgCreateGroupWithPolicy) GetDecisionPolicy() (DecisionPolicy, error) {
	decisionPolicy, ok := m.DecisionPolicy.GetCachedValue().(DecisionPolicy)
	if !ok {
		return nil, sdkerrors.ErrInvalidType.Wrapf("expected %T, got %T", (DecisionPolicy)(nil), m.DecisionPolicy.GetCachedValue())
	}
	return decisionPolicy, nil
}

// SetDecisionPolicy sets the decision policy for MsgCreateGroupWithPolicy.
func (m *MsgCreateGroupWithPolicy) SetDecisionPolicy(decisionPolicy DecisionPolicy) error {
	any, err := types.NewAnyWithValue(decisionPolicy)
	if err != nil {
		return err
	}
	m.DecisionPolicy = any
	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m MsgCreateGroupWithPolicy) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var decisionPolicy DecisionPolicy
	return unpacker.UnpackAny(m.DecisionPolicy, &decisionPolicy)
}

// GetSignBytes Implements Msg.
func (m MsgCreateGroupWithPolicy) GetSignBytes() []byte {
	return sdk.MustSortJSON(codec.ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgCreateGroupWithPolicy.
func (m MsgCreateGroupWithPolicy) GetSigners() []sdk.AccAddress {
	admin := sdk.MustAccAddressFromBech32(m.Admin)
	return []sdk.AccAddress{admin}
}

var (
	_ sdk.Msg            = &MsgCreateGroupPolicy{}
	_ legacytx.LegacyMsg = &MsgCreateGroupPolicy{}
)

// GetSignBytes Implements Msg.
func (m MsgCreateGroupPolicy) GetSignBytes() []byte {
	return sdk.MustSortJSON(codec.ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgCreateGroupPolicy.
func (m MsgCreateGroupPolicy) GetSigners() []sdk.AccAddress {
	admin := sdk.MustAccAddressFromBech32(m.Admin)
	return []sdk.AccAddress{admin}
}

var (
	_ sdk.Msg            = &MsgUpdateGroupPolicyAdmin{}
	_ legacytx.LegacyMsg = &MsgUpdateGroupPolicyAdmin{}
)

// GetSignBytes Implements Msg.
func (m MsgUpdateGroupPolicyAdmin) GetSignBytes() []byte {
	return sdk.MustSortJSON(codec.ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgUpdateGroupPolicyAdmin.
func (m MsgUpdateGroupPolicyAdmin) GetSigners() []sdk.AccAddress {
	admin := sdk.MustAccAddressFromBech32(m.Admin)

	return []sdk.AccAddress{admin}
}

var (
	_ sdk.Msg            = &MsgUpdateGroupPolicyDecisionPolicy{}
	_ legacytx.LegacyMsg = &MsgUpdateGroupPolicyDecisionPolicy{}

	_ types.UnpackInterfacesMessage = MsgUpdateGroupPolicyDecisionPolicy{}
)

// NewMsgUpdateGroupPolicyDecisionPolicy creates a new MsgUpdateGroupPolicyDecisionPolicy.
func NewMsgUpdateGroupPolicyDecisionPolicy(admin, address sdk.AccAddress, decisionPolicy DecisionPolicy) (*MsgUpdateGroupPolicyDecisionPolicy, error) {
	m := &MsgUpdateGroupPolicyDecisionPolicy{
		Admin:              admin.String(),
		GroupPolicyAddress: address.String(),
	}
	err := m.SetDecisionPolicy(decisionPolicy)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// SetDecisionPolicy sets the decision policy for MsgUpdateGroupPolicyDecisionPolicy.
func (m *MsgUpdateGroupPolicyDecisionPolicy) SetDecisionPolicy(decisionPolicy DecisionPolicy) error {
	msg, ok := decisionPolicy.(proto.Message)
	if !ok {
		return sdkerrors.ErrInvalidType.Wrapf("can't proto marshal %T", msg)
	}
	any, err := types.NewAnyWithValue(msg)
	if err != nil {
		return err
	}
	m.DecisionPolicy = any
	return nil
}

// GetSignBytes Implements Msg.
func (m MsgUpdateGroupPolicyDecisionPolicy) GetSignBytes() []byte {
	return sdk.MustSortJSON(codec.ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgUpdateGroupPolicyDecisionPolicy.
func (m MsgUpdateGroupPolicyDecisionPolicy) GetSigners() []sdk.AccAddress {
	admin := sdk.MustAccAddressFromBech32(m.Admin)

	return []sdk.AccAddress{admin}
}

// GetDecisionPolicy gets the decision policy of MsgUpdateGroupPolicyDecisionPolicy.
func (m *MsgUpdateGroupPolicyDecisionPolicy) GetDecisionPolicy() (DecisionPolicy, error) {
	decisionPolicy, ok := m.DecisionPolicy.GetCachedValue().(DecisionPolicy)
	if !ok {
		return nil, sdkerrors.ErrInvalidType.Wrapf("expected %T, got %T", (DecisionPolicy)(nil), m.DecisionPolicy.GetCachedValue())
	}

	return decisionPolicy, nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m MsgUpdateGroupPolicyDecisionPolicy) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var decisionPolicy DecisionPolicy
	return unpacker.UnpackAny(m.DecisionPolicy, &decisionPolicy)
}

var (
	_ sdk.Msg            = &MsgUpdateGroupPolicyMetadata{}
	_ legacytx.LegacyMsg = &MsgUpdateGroupPolicyMetadata{}
)

// GetSignBytes Implements Msg.
func (m MsgUpdateGroupPolicyMetadata) GetSignBytes() []byte {
	return sdk.MustSortJSON(codec.ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgUpdateGroupPolicyMetadata.
func (m MsgUpdateGroupPolicyMetadata) GetSigners() []sdk.AccAddress {
	admin := sdk.MustAccAddressFromBech32(m.Admin)

	return []sdk.AccAddress{admin}
}

var (
	_ sdk.Msg            = &MsgCreateGroupPolicy{}
	_ legacytx.LegacyMsg = &MsgCreateGroupPolicy{}

	_ types.UnpackInterfacesMessage = MsgCreateGroupPolicy{}
)

// NewMsgCreateGroupPolicy creates a new MsgCreateGroupPolicy.
func NewMsgCreateGroupPolicy(admin sdk.AccAddress, group uint64, metadata string, decisionPolicy DecisionPolicy) (*MsgCreateGroupPolicy, error) {
	m := &MsgCreateGroupPolicy{
		Admin:    admin.String(),
		GroupId:  group,
		Metadata: metadata,
	}
	err := m.SetDecisionPolicy(decisionPolicy)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// GetAdmin gets the admin of MsgCreateGroupPolicy.
func (m *MsgCreateGroupPolicy) GetAdmin() string {
	return m.Admin
}

// GetGroupID gets the group id of MsgCreateGroupPolicy.
func (m *MsgCreateGroupPolicy) GetGroupID() uint64 {
	return m.GroupId
}

// GetMetadata gets the metadata of MsgCreateGroupPolicy.
func (m *MsgCreateGroupPolicy) GetMetadata() string {
	return m.Metadata
}

// GetDecisionPolicy gets the decision policy of MsgCreateGroupPolicy.
func (m *MsgCreateGroupPolicy) GetDecisionPolicy() (DecisionPolicy, error) {
	decisionPolicy, ok := m.DecisionPolicy.GetCachedValue().(DecisionPolicy)
	if !ok {
		return nil, sdkerrors.ErrInvalidType.Wrapf("expected %T, got %T", (DecisionPolicy)(nil), m.DecisionPolicy.GetCachedValue())
	}
	return decisionPolicy, nil
}

// SetDecisionPolicy sets the decision policy of MsgCreateGroupPolicy.
func (m *MsgCreateGroupPolicy) SetDecisionPolicy(decisionPolicy DecisionPolicy) error {
	any, err := types.NewAnyWithValue(decisionPolicy)
	if err != nil {
		return err
	}
	m.DecisionPolicy = any
	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m MsgCreateGroupPolicy) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var decisionPolicy DecisionPolicy
	return unpacker.UnpackAny(m.DecisionPolicy, &decisionPolicy)
}

var (
	_ sdk.Msg            = &MsgSubmitProposal{}
	_ legacytx.LegacyMsg = &MsgSubmitProposal{}
)

// NewMsgSubmitProposal creates a new MsgSubmitProposal.
func NewMsgSubmitProposal(address string, proposers []string, msgs []sdk.Msg, metadata string, exec Exec, title, summary string) (*MsgSubmitProposal, error) {
	m := &MsgSubmitProposal{
		GroupPolicyAddress: address,
		Proposers:          proposers,
		Metadata:           metadata,
		Exec:               exec,
		Title:              title,
		Summary:            summary,
	}
	err := m.SetMsgs(msgs)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// GetSignBytes Implements Msg.
func (m MsgSubmitProposal) GetSignBytes() []byte {
	return sdk.MustSortJSON(codec.ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgSubmitProposal.
func (m MsgSubmitProposal) GetSigners() []sdk.AccAddress {
	addrs, err := m.getProposerAccAddresses()
	if err != nil {
		panic(err)
	}

	return addrs
}

// getProposerAccAddresses returns the proposers as `[]sdk.AccAddress`.
func (m *MsgSubmitProposal) getProposerAccAddresses() ([]sdk.AccAddress, error) {
	addrs := make([]sdk.AccAddress, len(m.Proposers))
	for i, proposer := range m.Proposers {
		addr, err := sdk.AccAddressFromBech32(proposer)
		if err != nil {
			return nil, errorsmod.Wrap(err, "proposers")
		}
		addrs[i] = addr
	}

	return addrs, nil
}

// SetMsgs packs msgs into Any's
func (m *MsgSubmitProposal) SetMsgs(msgs []sdk.Msg) error {
	anys, err := tx.SetMsgs(msgs)
	if err != nil {
		return err
	}
	m.Messages = anys
	return nil
}

// GetMsgs unpacks m.Messages Any's into sdk.Msg's
func (m MsgSubmitProposal) GetMsgs() ([]sdk.Msg, error) {
	return tx.GetMsgs(m.Messages, "proposal")
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m MsgSubmitProposal) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	return tx.UnpackInterfaces(unpacker, m.Messages)
}

var (
	_ sdk.Msg            = &MsgWithdrawProposal{}
	_ legacytx.LegacyMsg = &MsgWithdrawProposal{}
)

// GetSignBytes Implements Msg.
func (m MsgWithdrawProposal) GetSignBytes() []byte {
	return sdk.MustSortJSON(codec.ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgWithdrawProposal.
func (m MsgWithdrawProposal) GetSigners() []sdk.AccAddress {
	admin := sdk.MustAccAddressFromBech32(m.Address)

	return []sdk.AccAddress{admin}
}

var (
	_ sdk.Msg            = &MsgVote{}
	_ legacytx.LegacyMsg = &MsgVote{}
)

// GetSignBytes Implements Msg.
func (m MsgVote) GetSignBytes() []byte {
	return sdk.MustSortJSON(codec.ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgVote.
func (m MsgVote) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.Voter)

	return []sdk.AccAddress{addr}
}

var (
	_ sdk.Msg            = &MsgExec{}
	_ legacytx.LegacyMsg = &MsgExec{}
)

// GetSignBytes Implements Msg.
func (m MsgExec) GetSignBytes() []byte {
	return sdk.MustSortJSON(codec.ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgExec.
func (m MsgExec) GetSigners() []sdk.AccAddress {
	signer := sdk.MustAccAddressFromBech32(m.Executor)

	return []sdk.AccAddress{signer}
}

var (
	_ sdk.Msg            = &MsgLeaveGroup{}
	_ legacytx.LegacyMsg = &MsgLeaveGroup{}
)

// GetSignBytes Implements Msg
func (m MsgLeaveGroup) GetSignBytes() []byte {
	return sdk.MustSortJSON(codec.ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgLeaveGroup
func (m MsgLeaveGroup) GetSigners() []sdk.AccAddress {
	signer := sdk.MustAccAddressFromBech32(m.Address)

	return []sdk.AccAddress{signer}
}
