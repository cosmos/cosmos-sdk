package cometbft

import (
	"context"

	abci "github.com/cometbft/cometbft/api/cometbft/abci/v1"

	"cosmossdk.io/core/comet"
	corecontext "cosmossdk.io/core/context"
)

func contextWithCometInfo(ctx context.Context, info comet.Info) context.Context {
	return context.WithValue(ctx, corecontext.CometInfoKey, info)
}

// toCoreEvidence takes comet evidence and returns sdk evidence
func toCoreEvidence(ev []abci.Misbehavior) []comet.Evidence {
	evidence := make([]comet.Evidence, len(ev))
	for i, e := range ev {
		evidence[i] = comet.Evidence{
			Type:             comet.MisbehaviorType(e.Type),
			Height:           e.Height,
			Time:             e.Time,
			TotalVotingPower: e.TotalVotingPower,
			Validator: comet.Validator{
				Address: e.Validator.Address,
				Power:   e.Validator.Power,
			},
		}
	}
	return evidence
}

// toCoreCommitInfo takes comet commit info and returns sdk commit info
func toCoreCommitInfo(commit abci.CommitInfo) comet.CommitInfo {
	ci := comet.CommitInfo{
		Round: commit.Round,
	}

	for _, v := range commit.Votes {
		ci.Votes = append(ci.Votes, comet.VoteInfo{
			Validator: comet.Validator{
				Address: v.Validator.Address,
				Power:   v.Validator.Power,
			},
			BlockIDFlag: comet.BlockIDFlag(v.BlockIdFlag),
		})
	}
	return ci
}

// toCoreExtendedCommitInfo takes comet extended commit info and returns sdk commit info
func toCoreExtendedCommitInfo(commit abci.ExtendedCommitInfo) comet.CommitInfo {
	ci := comet.CommitInfo{
		Round: commit.Round,
		Votes: make([]comet.VoteInfo, len(commit.Votes)),
	}

	for i, v := range commit.Votes {
		ci.Votes[i] = comet.VoteInfo{
			Validator: comet.Validator{
				Address: v.Validator.Address,
				Power:   v.Validator.Power,
			},
			BlockIDFlag: comet.BlockIDFlag(v.BlockIdFlag),
		}
	}

	return ci
}
