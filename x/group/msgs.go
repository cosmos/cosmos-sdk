package group

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec/legacy"

	proto "github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
	errors "github.com/cosmos/cosmos-sdk/x/group/errors"
	"github.com/cosmos/cosmos-sdk/x/group/internal/math"
)

var _ sdk.Msg = &MsgCreateGroup{}

// Route Implements Msg.
func (m MsgCreateGroup) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgCreateGroup) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgCreateGroup) GetSignBytes() []byte {
	return sdk.MustSortJSON(legacy.Cdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgCreateGroup.
func (m MsgCreateGroup) GetSigners() []sdk.AccAddress {
	admin, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{admin}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgCreateGroup) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		return sdkerrors.Wrap(err, "admin")
	}

	return m.validateMembers()
}

func (m MsgCreateGroup) validateMembers() error {

	err := validateMembers(m.Members)
	if err != nil {
		return err
	}
	return nil
}

func (m Member) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return sdkerrors.Wrap(err, "address")
	}

	if _, err := math.NewNonNegativeDecFromString(m.Weight); err != nil {
		return sdkerrors.Wrap(err, "weight")
	}

	return nil
}

var _ sdk.Msg = &MsgUpdateGroupAdmin{}

// Route Implements Msg.
func (m MsgUpdateGroupAdmin) Route() string {
	return sdk.MsgTypeURL(&m)
}

// Type Implements Msg.
func (m MsgUpdateGroupAdmin) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgUpdateGroupAdmin) GetSignBytes() []byte {
	return sdk.MustSortJSON(legacy.Cdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgUpdateGroupAdmin.
func (m MsgUpdateGroupAdmin) GetSigners() []sdk.AccAddress {
	admin, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{admin}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgUpdateGroupAdmin) ValidateBasic() error {
	if m.GroupId == 0 {
		return sdkerrors.Wrap(errors.ErrEmpty, "group id")
	}

	admin, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		return sdkerrors.Wrap(err, "admin")
	}

	newAdmin, err := sdk.AccAddressFromBech32(m.NewAdmin)
	if err != nil {
		return sdkerrors.Wrap(err, "new admin")
	}

	if admin.Equals(newAdmin) {
		return sdkerrors.Wrap(errors.ErrInvalid, "new and old admin are the same")
	}
	return nil
}

func (m *MsgUpdateGroupAdmin) GetGroupID() uint64 {
	return m.GroupId
}

var _ sdk.Msg = &MsgUpdateGroupMetadata{}

// Route Implements Msg.
func (m MsgUpdateGroupMetadata) Route() string {
	return sdk.MsgTypeURL(&m)
}

// Type Implements Msg.
func (m MsgUpdateGroupMetadata) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgUpdateGroupMetadata) GetSignBytes() []byte {
	return sdk.MustSortJSON(legacy.Cdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgUpdateGroupMetadata.
func (m MsgUpdateGroupMetadata) GetSigners() []sdk.AccAddress {
	admin, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{admin}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgUpdateGroupMetadata) ValidateBasic() error {
	if m.GroupId == 0 {
		return sdkerrors.Wrap(errors.ErrEmpty, "group id")

	}
	_, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		return sdkerrors.Wrap(err, "admin")
	}

	return nil
}

func (m *MsgUpdateGroupMetadata) GetGroupID() uint64 {
	return m.GroupId
}

var _ sdk.Msg = &MsgUpdateGroupMembers{}

// Route Implements Msg.
func (m MsgUpdateGroupMembers) Route() string {
	return sdk.MsgTypeURL(&m)
}

// Type Implements Msg.
func (m MsgUpdateGroupMembers) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgUpdateGroupMembers) GetSignBytes() []byte {
	return sdk.MustSortJSON(legacy.Cdc.MustMarshalJSON(&m))
}

var _ sdk.Msg = &MsgUpdateGroupMembers{}

// GetSigners returns the expected signers for a MsgUpdateGroupMembers.
func (m MsgUpdateGroupMembers) GetSigners() []sdk.AccAddress {
	admin, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{admin}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgUpdateGroupMembers) ValidateBasic() error {
	if m.GroupId == 0 {
		return sdkerrors.Wrap(errors.ErrEmpty, "group id")

	}
	_, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		return sdkerrors.Wrap(err, "admin")
	}

	if len(m.MemberUpdates) == 0 {
		return sdkerrors.Wrap(errors.ErrEmpty, "member updates")
	}
	members := Members{Members: m.MemberUpdates}
	if err := members.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "members")
	}
	return nil
}

func (m *MsgUpdateGroupMembers) GetGroupID() uint64 {
	return m.GroupId
}

var _ sdk.Msg = &MsgCreateGroupWithPolicy{}
var _ types.UnpackInterfacesMessage = MsgCreateGroupWithPolicy{}

// NewMsgCreateGroupWithPolicy creates a new MsgCreateGroupWithPolicy.
func NewMsgCreateGroupWithPolicy(admin string, members []Member, group_metadata []byte, group_policy_metadata []byte, groupPolicyAsAdmin bool, decisionPolicy DecisionPolicy) (*MsgCreateGroupWithPolicy, error) {
	m := &MsgCreateGroupWithPolicy{
		Admin:               admin,
		Members:             members,
		GroupMetadata:       group_metadata,
		GroupPolicyMetadata: group_policy_metadata,
		GroupPolicyAsAdmin:  groupPolicyAsAdmin,
	}
	err := m.SetDecisionPolicy(decisionPolicy)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (m *MsgCreateGroupWithPolicy) GetDecisionPolicy() DecisionPolicy {
	decisionPolicy, ok := m.DecisionPolicy.GetCachedValue().(DecisionPolicy)
	if !ok {
		return nil
	}
	return decisionPolicy
}

func (m *MsgCreateGroupWithPolicy) SetDecisionPolicy(decisionPolicy DecisionPolicy) error {
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

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m MsgCreateGroupWithPolicy) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var decisionPolicy DecisionPolicy
	return unpacker.UnpackAny(m.DecisionPolicy, &decisionPolicy)
}

// Route Implements Msg.
func (m MsgCreateGroupWithPolicy) Route() string {
	return sdk.MsgTypeURL(&m)
}

// Type Implements Msg.
func (m MsgCreateGroupWithPolicy) Type() string {
	return sdk.MsgTypeURL(&m)
}

// GetSignBytes Implements Msg.
func (m MsgCreateGroupWithPolicy) GetSignBytes() []byte {
	return sdk.MustSortJSON(legacy.Cdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgCreateGroupWithPolicy.
func (m MsgCreateGroupWithPolicy) GetSigners() []sdk.AccAddress {
	admin, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{admin}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgCreateGroupWithPolicy) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		return sdkerrors.Wrap(err, "admin")
	}
	policy := m.GetDecisionPolicy()
	if policy == nil {
		return sdkerrors.Wrap(errors.ErrEmpty, "decision policy")
	}
	if err := policy.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "decision policy")
	}

	return m.validateMembers()
}

func (m MsgCreateGroupWithPolicy) validateMembers() error {

	err := validateMembers(m.Members)
	if err != nil {
		return err
	}
	return nil
}

var _ sdk.Msg = &MsgCreateGroupPolicy{}

// Route Implements Msg.
func (m MsgCreateGroupPolicy) Route() string {
	return sdk.MsgTypeURL(&m)
}

// Type Implements Msg.
func (m MsgCreateGroupPolicy) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgCreateGroupPolicy) GetSignBytes() []byte {
	return sdk.MustSortJSON(legacy.Cdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgCreateGroupPolicy.
func (m MsgCreateGroupPolicy) GetSigners() []sdk.AccAddress {
	admin, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{admin}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgCreateGroupPolicy) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		return sdkerrors.Wrap(err, "admin")
	}
	if m.GroupId == 0 {
		return sdkerrors.Wrap(errors.ErrEmpty, "group id")
	}

	policy := m.GetDecisionPolicy()
	if policy == nil {
		return sdkerrors.Wrap(errors.ErrEmpty, "decision policy")
	}

	if err := policy.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "decision policy")
	}
	return nil
}

var _ sdk.Msg = &MsgUpdateGroupPolicyAdmin{}

// Route Implements Msg.
func (m MsgUpdateGroupPolicyAdmin) Route() string {
	return sdk.MsgTypeURL(&m)
}

// Type Implements Msg.
func (m MsgUpdateGroupPolicyAdmin) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgUpdateGroupPolicyAdmin) GetSignBytes() []byte {
	return sdk.MustSortJSON(legacy.Cdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgUpdateGroupPolicyAdmin.
func (m MsgUpdateGroupPolicyAdmin) GetSigners() []sdk.AccAddress {
	admin, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{admin}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgUpdateGroupPolicyAdmin) ValidateBasic() error {
	admin, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		return sdkerrors.Wrap(err, "admin")
	}

	newAdmin, err := sdk.AccAddressFromBech32(m.NewAdmin)
	if err != nil {
		return sdkerrors.Wrap(err, "new admin")
	}

	_, err = sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return sdkerrors.Wrap(err, "group policy")
	}

	if admin.Equals(newAdmin) {
		return sdkerrors.Wrap(errors.ErrInvalid, "new and old admin are same")
	}
	return nil
}

var _ sdk.Msg = &MsgUpdateGroupPolicyDecisionPolicy{}
var _ types.UnpackInterfacesMessage = MsgUpdateGroupPolicyDecisionPolicy{}

func NewMsgUpdateGroupPolicyDecisionPolicyRequest(admin sdk.AccAddress, address sdk.AccAddress, decisionPolicy DecisionPolicy) (*MsgUpdateGroupPolicyDecisionPolicy, error) {
	m := &MsgUpdateGroupPolicyDecisionPolicy{
		Admin:   admin.String(),
		Address: address.String(),
	}
	err := m.SetDecisionPolicy(decisionPolicy)
	if err != nil {
		return nil, err
	}
	return m, nil
}

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

// Route Implements Msg.
func (m MsgUpdateGroupPolicyDecisionPolicy) Route() string {
	return sdk.MsgTypeURL(&m)
}

// Type Implements Msg.
func (m MsgUpdateGroupPolicyDecisionPolicy) Type() string {
	return sdk.MsgTypeURL(&m)
}

// GetSignBytes Implements Msg.
func (m MsgUpdateGroupPolicyDecisionPolicy) GetSignBytes() []byte {
	return sdk.MustSortJSON(legacy.Cdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgUpdateGroupPolicyDecisionPolicy.
func (m MsgUpdateGroupPolicyDecisionPolicy) GetSigners() []sdk.AccAddress {
	admin, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{admin}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgUpdateGroupPolicyDecisionPolicy) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		return sdkerrors.Wrap(err, "admin")
	}

	_, err = sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return sdkerrors.Wrap(err, "group policy")
	}

	policy := m.GetDecisionPolicy()
	if policy == nil {
		return sdkerrors.Wrap(errors.ErrEmpty, "decision policy")
	}

	if err := policy.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "decision policy")
	}

	return nil
}

func (m *MsgUpdateGroupPolicyDecisionPolicy) GetDecisionPolicy() DecisionPolicy {
	decisionPolicy, ok := m.DecisionPolicy.GetCachedValue().(DecisionPolicy)
	if !ok {
		return nil
	}
	return decisionPolicy
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m MsgUpdateGroupPolicyDecisionPolicy) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var decisionPolicy DecisionPolicy
	return unpacker.UnpackAny(m.DecisionPolicy, &decisionPolicy)
}

var _ sdk.Msg = &MsgUpdateGroupPolicyMetadata{}

// Route Implements Msg.
func (m MsgUpdateGroupPolicyMetadata) Route() string {
	return sdk.MsgTypeURL(&m)
}

// Type Implements Msg.
func (m MsgUpdateGroupPolicyMetadata) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgUpdateGroupPolicyMetadata) GetSignBytes() []byte {
	return sdk.MustSortJSON(legacy.Cdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgUpdateGroupPolicyMetadata.
func (m MsgUpdateGroupPolicyMetadata) GetSigners() []sdk.AccAddress {
	admin, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{admin}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgUpdateGroupPolicyMetadata) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		return sdkerrors.Wrap(err, "admin")
	}

	_, err = sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return sdkerrors.Wrap(err, "group policy")
	}

	return nil
}

var _ sdk.Msg = &MsgCreateGroupPolicy{}
var _ types.UnpackInterfacesMessage = MsgCreateGroupPolicy{}

// NewMsgCreateGroupPolicy creates a new MsgCreateGroupPolicy.
func NewMsgCreateGroupPolicy(admin sdk.AccAddress, group uint64, metadata []byte, decisionPolicy DecisionPolicy) (*MsgCreateGroupPolicy, error) {
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

func (m *MsgCreateGroupPolicy) GetAdmin() string {
	return m.Admin
}

func (m *MsgCreateGroupPolicy) GetGroupID() uint64 {
	return m.GroupId
}

func (m *MsgCreateGroupPolicy) GetMetadata() []byte {
	return m.Metadata
}

func (m *MsgCreateGroupPolicy) GetDecisionPolicy() DecisionPolicy {
	decisionPolicy, ok := m.DecisionPolicy.GetCachedValue().(DecisionPolicy)
	if !ok {
		return nil
	}
	return decisionPolicy
}

func (m *MsgCreateGroupPolicy) SetDecisionPolicy(decisionPolicy DecisionPolicy) error {
	msg, ok := decisionPolicy.(proto.Message)
	if !ok {
		return fmt.Errorf("can't proto marshal %T", msg)
	}
	any, err := types.NewAnyWithValue(msg)
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

var _ sdk.Msg = &MsgSubmitProposal{}

// NewMsgSubmitProposalRequest creates a new MsgSubmitProposal.
func NewMsgSubmitProposalRequest(address string, proposers []string, msgs []sdk.Msg, metadata []byte, exec Exec) (*MsgSubmitProposal, error) {
	m := &MsgSubmitProposal{
		Address:   address,
		Proposers: proposers,
		Metadata:  metadata,
		Exec:      exec,
	}
	err := m.SetMsgs(msgs)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// Route Implements Msg.
func (m MsgSubmitProposal) Route() string {
	return sdk.MsgTypeURL(&m)
}

// Type Implements Msg.
func (m MsgSubmitProposal) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgSubmitProposal) GetSignBytes() []byte {
	return sdk.MustSortJSON(legacy.Cdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgSubmitProposal.
func (m MsgSubmitProposal) GetSigners() []sdk.AccAddress {
	addrs := make([]sdk.AccAddress, len(m.Proposers))
	for i, proposer := range m.Proposers {
		addr, err := sdk.AccAddressFromBech32(proposer)
		if err != nil {
			panic(err)
		}
		addrs[i] = addr
	}
	return addrs
}

// ValidateBasic does a sanity check on the provided data
func (m MsgSubmitProposal) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return sdkerrors.Wrap(err, "group policy")
	}

	if len(m.Proposers) == 0 {
		return sdkerrors.Wrap(errors.ErrEmpty, "proposers")
	}
	addrs := make([]sdk.AccAddress, len(m.Proposers))
	for i, proposer := range m.Proposers {
		addr, err := sdk.AccAddressFromBech32(proposer)
		if err != nil {
			return sdkerrors.Wrap(err, "proposers")
		}
		addrs[i] = addr
	}
	if err := AccAddresses(addrs).ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "proposers")
	}

	msgs := m.GetMsgs()
	for i, msg := range msgs {
		if err := msg.ValidateBasic(); err != nil {
			return sdkerrors.Wrapf(err, "msg %d", i)
		}
	}
	return nil
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
func (m MsgSubmitProposal) GetMsgs() []sdk.Msg {
	msgs, err := tx.GetMsgs(m.Messages, "proposal")
	if err != nil {
		panic(err)
	}
	return msgs
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m MsgSubmitProposal) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	return tx.UnpackInterfaces(unpacker, m.Messages)
}

var _ sdk.Msg = &MsgWithdrawProposal{}

// Route Implements Msg.
func (m MsgWithdrawProposal) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgWithdrawProposal) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgWithdrawProposal) GetSignBytes() []byte {
	return sdk.MustSortJSON(legacy.Cdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgWithdrawProposal.
func (m MsgWithdrawProposal) GetSigners() []sdk.AccAddress {
	admin, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{admin}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgWithdrawProposal) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return sdkerrors.Wrap(err, "admin")
	}

	if m.ProposalId == 0 {
		return sdkerrors.Wrap(errors.ErrEmpty, "proposal id")
	}

	return nil
}

var _ sdk.Msg = &MsgVote{}

// Route Implements Msg.
func (m MsgVote) Route() string {
	return sdk.MsgTypeURL(&m)
}

// Type Implements Msg.
func (m MsgVote) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgVote) GetSignBytes() []byte {
	return sdk.MustSortJSON(legacy.Cdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgVote.
func (m MsgVote) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Voter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgVote) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Voter)
	if err != nil {
		return sdkerrors.Wrap(err, "voter")
	}
	if m.ProposalId == 0 {
		return sdkerrors.Wrap(errors.ErrEmpty, "proposal id")
	}
	if m.Option == VOTE_OPTION_UNSPECIFIED {
		return sdkerrors.Wrap(errors.ErrEmpty, "vote option")
	}
	if _, ok := VoteOption_name[int32(m.Option)]; !ok {
		return sdkerrors.Wrap(errors.ErrInvalid, "vote option")
	}
	return nil
}

var _ sdk.Msg = &MsgExec{}

// Route Implements Msg.
func (m MsgExec) Route() string {
	return sdk.MsgTypeURL(&m)
}

// Type Implements Msg.
func (m MsgExec) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgExec) GetSignBytes() []byte {
	return sdk.MustSortJSON(legacy.Cdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgExec.
func (m MsgExec) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(m.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgExec) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Signer)
	if err != nil {
		return sdkerrors.Wrap(err, "signer")
	}
	if m.ProposalId == 0 {
		return sdkerrors.Wrap(errors.ErrEmpty, "proposal id")
	}
	return nil
}

func validateMembers(members []Member) error {
	index := make(map[string]struct{}, len(members))
	for i := range members {
		member := members[i]
		_, err := sdk.AccAddressFromBech32(member.Address)
		if err != nil {
			return sdkerrors.Wrap(err, "address")
		}

		if _, err := math.NewPositiveDecFromString(member.Weight); err != nil {
			return sdkerrors.Wrap(err, "weight")
		}

		addr := member.Address
		if _, exists := index[addr]; exists {
			return sdkerrors.Wrapf(errors.ErrDuplicate, "address: %s", addr)
		}
		index[addr] = struct{}{}
	}

	return nil
}
