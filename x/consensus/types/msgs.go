package types

import (
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

var (
	_ sdk.Msg            = &MsgUpdateParams{}
	_ legacytx.LegacyMsg = &MsgUpdateParams{}
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
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// ValidateBasic performs basic MsgUpdateParams message validation.
func (msg MsgUpdateParams) ValidateBasic() error {
	params := cmttypes.ConsensusParamsFromProto(msg.ToProtoConsensusParams())
	return params.ValidateBasic()
}

func (msg MsgUpdateParams) ToProtoConsensusParams() cmtproto.ConsensusParams {
	return cmtproto.ConsensusParams{
		Block: &cmtproto.BlockParams{
			MaxBytes: msg.Block.MaxBytes,
			MaxGas:   msg.Block.MaxGas,
		},
		Evidence: &cmtproto.EvidenceParams{
			MaxAgeNumBlocks: msg.Evidence.MaxAgeNumBlocks,
			MaxAgeDuration:  msg.Evidence.MaxAgeDuration,
			MaxBytes:        msg.Evidence.MaxBytes,
		},
		Validator: &cmtproto.ValidatorParams{
			PubKeyTypes: msg.Validator.PubKeyTypes,
		},
		Version: cmttypes.DefaultConsensusParams().ToProto().Version, // Version is stored in x/upgrade
	}
}
