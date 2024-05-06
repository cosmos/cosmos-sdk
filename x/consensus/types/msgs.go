package types

import (
	"errors"

	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v1"
	cmttypes "github.com/cometbft/cometbft/types"
)

func (msg MsgUpdateParams) ToProtoConsensusParams() (cmtproto.ConsensusParams, error) {
	if msg.Evidence == nil || msg.Block == nil || msg.Validator == nil {
		return cmtproto.ConsensusParams{}, errors.New("all parameters must be present")
	}

	cp := cmtproto.ConsensusParams{
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
	if msg.Feature != nil && msg.Feature.VoteExtensionsEnableHeight != nil && msg.Feature.PbtsEnableHeight != nil {
		cp.Feature = &cmtproto.FeatureParams{
			VoteExtensionsEnableHeight: msg.Feature.VoteExtensionsEnableHeight,
			PbtsEnableHeight:           msg.Feature.PbtsEnableHeight,
		}
	}
	if msg.Synchrony != nil {
		cp.Synchrony = &cmtproto.SynchronyParams{
			Precision:    msg.Synchrony.Precision,
			MessageDelay: msg.Synchrony.MessageDelay,
		}
	}

	return cp, nil
}
