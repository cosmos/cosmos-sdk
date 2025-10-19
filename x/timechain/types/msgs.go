// SPDX-License-Identifier: MPL-2.0
// Copyright © 2025 Timechain-Arweave-LunCoSim Contributors

package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"cosmossdk.io/x/auth/migrations/legacytx"
)

var (
	_ sdk.Msg = &MsgProposeSlot{}
	_ sdk.Msg = &MsgConfirmSlot{}
	_ sdk.Msg = &MsgRelayEvent{}
	_ legacytx.LegacyMsg = &MsgProposeSlot{}
	_ legacytx.LegacyMsg = &MsgConfirmSlot{}
	_ legacytx.LegacyMsg = &MsgRelayEvent{}
)

// MsgProposeSlot
func NewMsgProposeSlot(creator sdk.AccAddress, slot uint64, vdfOutput []byte, payloadHash []byte) *MsgProposeSlot {
	return &MsgProposeSlot{
		Creator:      creator.String(),
		Slot:         slot,
		VdfOutput:    vdfOutput,
		PayloadHash:  payloadHash,
	}
}

func (msg *MsgProposeSlot) Route() string {
	return RouterKey
}

func (msg *MsgProposeSlot) Type() string {
	return "ProposeSlot"
}

func (msg *MsgProposeSlot) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgProposeSlot) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgProposeSlot) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	return err
}

// MsgConfirmSlot
func NewMsgConfirmSlot(creator sdk.AccAddress, slotId uint64, validator sdk.ValAddress, sig []byte) *MsgConfirmSlot {
	return &MsgConfirmSlot{
		Creator:   creator.String(),
		SlotId:    slotId,
		Validator: validator.String(),
		Sig:       sig,
	}
}

func (msg *MsgConfirmSlot) Route() string {
	return RouterKey
}

func (msg *MsgConfirmSlot) Type() string {
	return "ConfirmSlot"
}

func (msg *MsgConfirmSlot) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgConfirmSlot) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgConfirmSlot) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	return err
}

// MsgRelayEvent
func NewMsgRelayEvent(creator sdk.AccAddress, event []byte, tssSig []byte) *MsgRelayEvent {
	return &MsgRelayEvent{
		Creator: creator.String(),
		Event:   event,
		TssSig:  tssSig,
	}
}

func (msg *MsgRelayEvent) Route() string {
	return RouterKey
}

func (msg *MsgRelayEvent) Type() string {
	return "RelayEvent"
}

func (msg *MsgRelayEvent) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgRelayEvent) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRelayEvent) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	return err
}
