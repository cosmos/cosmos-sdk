package baseapp

import (
	"time"

	"cosmossdk.io/core/comet"
	"cosmossdk.io/core/header"

	abci "github.com/cometbft/cometbft/abci/types"
)

func NewHeaderInfo(height int64, hash []byte, t time.Time, chainID string) header.Info {
	return header.Info{
		Height:  height,
		Hash:    hash,
		Time:    t,
		ChainID: chainID,
	}
}

func NewCometInfo(bg abci.RequestBeginBlock) comet.Info {
	return comet.Info{
		Evidence:          FromABCIEvidence(bg.ByzantineValidators),
		ValidatorsHash:    bg.Hash,
		ProposerAddress:   bg.Header.ProposerAddress,
		DecidedLastCommit: FromABCICommitInfo(bg.LastCommitInfo),
	}
}

// Utils

// FromABCIEvidence converts a CometBFT concrete Evidence type to
// SDK Evidence.
func FromABCIEvidence(e []abci.Misbehavior) []comet.Misbehavior {
	misbehavior := make([]comet.Misbehavior, len(e))

	for i, ev := range e {
		misbehavior[i] = comet.Misbehavior{
			Type:   comet.MisbehaviorType(ev.Type),
			Height: ev.Height,
			Validator: comet.Validator{
				Address: ev.Validator.Address,
				Power:   ev.Validator.Power,
			},
			Time:             ev.Time,
			TotalVotingPower: ev.TotalVotingPower,
		}
	}

	return misbehavior
}

func FromABCICommitInfo(ci abci.CommitInfo) comet.CommitInfo {
	votes := make([]*comet.VoteInfo, len(ci.Votes))
	for i, v := range ci.Votes {
		votes[i] = &comet.VoteInfo{
			Validator: comet.Validator{
				Address: v.Validator.Address,
				Power:   v.Validator.Power,
			},
			SignedLastBlock: v.SignedLastBlock,
		}
	}

	return comet.CommitInfo{
		Round: ci.Round,
		Votes: votes,
	}
}
