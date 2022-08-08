package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

// bank message types
const (
	TypeMsgUpdateParams = "update_params"
)

var (
	_ sdk.Msg = &MsgUpdateParams{}
)

// GetSigners returns the signer addresses that are expected to sign the result
// of GetSignBytes.
func (msg MsgUpdateParams) GetSigners() []sdk.AccAddress {
	authority, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{authority}
}

// GetSignBytes returns the raw bytes for a MsgUpdateParams message that
// the expected signer needs to sign.
func (msg MsgUpdateParams) GetSignBytes() []byte {
	return []byte{}
}

// ValidateBasic performs basic MsgUpdateParams message validation.
func (msg MsgUpdateParams) ValidateBasic() error {
	params := tmtypes.ConsensusParamsFromProto(msg.Params)
	return Validate(params)
}
