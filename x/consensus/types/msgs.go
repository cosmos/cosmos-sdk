package types

import (
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// bank message types
const (
	TypeMsgUpdateParams = "update_params"
)

var _ sdk.Msg = &MsgUpdateParams{}

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
	params := tmtypes.ConsensusParamsFromProto(msg.ToProtoConsensusParams())
	return Validate(params)
}

func (msg MsgUpdateParams) ToProtoConsensusParams() tmproto.ConsensusParams {
	return tmproto.ConsensusParams{
		Block: &tmproto.BlockParams{
			MaxBytes: msg.Block.MaxBytes,
			MaxGas:   msg.Block.MaxGas,
		},
		Evidence: &tmproto.EvidenceParams{
			MaxAgeNumBlocks: msg.Evidence.MaxAgeNumBlocks,
			MaxAgeDuration:  msg.Evidence.MaxAgeDuration,
			MaxBytes:        msg.Evidence.MaxBytes,
		},
		Validator: &tmproto.ValidatorParams{
			PubKeyTypes: msg.Validator.PubKeyTypes,
		},
		Version: tmtypes.DefaultConsensusParams().ToProto().Version,
	}
}
