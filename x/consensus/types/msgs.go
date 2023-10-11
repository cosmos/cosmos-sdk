package types

import (
<<<<<<< HEAD
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtypes "github.com/cometbft/cometbft/types"
=======
	"errors"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"
>>>>>>> ed14ec03b (chore: check for nil params (#18041))

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

const (
	TypeMsgUpdateParams = "update_params"
)

<<<<<<< HEAD
var _ legacytx.LegacyMsg = &MsgUpdateParams{}

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

func (msg MsgUpdateParams) Route() string {
	return sdk.MsgTypeURL(&msg)
}

func (msg MsgUpdateParams) Type() string {
	return sdk.MsgTypeURL(&msg)
}

// ValidateBasic performs basic MsgUpdateParams message validation.
func (msg MsgUpdateParams) ValidateBasic() error {
	params := tmtypes.ConsensusParamsFromProto(msg.ToProtoConsensusParams())
	return params.ValidateBasic()
}

func (msg MsgUpdateParams) ToProtoConsensusParams() tmproto.ConsensusParams {
	return tmproto.ConsensusParams{
		Block: &tmproto.BlockParams{
=======
func (msg MsgUpdateParams) ToProtoConsensusParams() (cmtproto.ConsensusParams, error) {
	if msg.Evidence == nil || msg.Block == nil || msg.Validator == nil {
		return cmtproto.ConsensusParams{}, errors.New("all parameters must be present")
	}

	cp := cmtproto.ConsensusParams{
		Block: &cmtproto.BlockParams{
>>>>>>> ed14ec03b (chore: check for nil params (#18041))
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
		Version: tmtypes.DefaultConsensusParams().ToProto().Version, // Version is stored in x/upgrade
	}
<<<<<<< HEAD
=======

	if msg.Abci != nil {
		cp.Abci = &cmtproto.ABCIParams{
			VoteExtensionsEnableHeight: msg.Abci.VoteExtensionsEnableHeight,
		}
	}

	return cp, nil
>>>>>>> ed14ec03b (chore: check for nil params (#18041))
}
