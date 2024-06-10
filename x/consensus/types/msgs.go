package types

import (
	"errors"

	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v1"
	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/gogoproto/types"
)

func (msg MsgUpdateParams) ToProtoConsensusParams() (cmtproto.ConsensusParams, error) {
	if msg.Evidence == nil || msg.Block == nil || msg.Validator == nil {
		return cmtproto.ConsensusParams{}, errors.New("all parameters must be present")
	}

	if msg.Abci != nil && msg.Feature != nil && msg.Feature.VoteExtensionsEnableHeight != nil {
		return cmtproto.ConsensusParams{}, errors.New("abci in sections Feature and (deprecated) ABCI cannot be used simultaneously")
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
		Version:   cmttypes.DefaultConsensusParams().ToProto().Version, // Version is stored in x/upgrade
		Feature:   &cmtproto.FeatureParams{},
		Synchrony: &cmtproto.SynchronyParams{},
	}

	if msg.Abci != nil {
		cp.Feature.VoteExtensionsEnableHeight = &types.Int64Value{
			Value: msg.Abci.VoteExtensionsEnableHeight,
		}
	}

	if msg.Feature != nil {
		if msg.Feature.VoteExtensionsEnableHeight != nil {
			cp.Feature.VoteExtensionsEnableHeight = &types.Int64Value{
				Value: msg.Feature.GetVoteExtensionsEnableHeight().GetValue(),
			}
		}
		if msg.Feature.PbtsEnableHeight != nil {
			cp.Feature.PbtsEnableHeight = &types.Int64Value{
				Value: msg.Feature.GetPbtsEnableHeight().GetValue(),
			}
		}
	}

	if msg.Synchrony != nil {
		if msg.Synchrony.MessageDelay != nil {
			delay := *msg.Synchrony.MessageDelay
			cp.Synchrony.MessageDelay = &delay
		}
		if msg.Synchrony.Precision != nil {
			precision := *msg.Synchrony.Precision
			cp.Synchrony.Precision = &precision
		}
	}

	return cp, nil
}
