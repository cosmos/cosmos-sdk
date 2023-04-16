package types

import (
	"cosmossdk.io/core/blockinfo"
	abci "github.com/cometbft/cometbft/abci/types"
)

var _ blockinfo.Service = (*BlockInfo)(nil)

type BlockInfo struct {
	Height            int64
	Evidence          []blockinfo.Misbehavior
	HeaderHash        []byte
	ValidatorsHash    []byte
	ProposerAddress   []byte
	DecidedLastCommit blockinfo.CommitInfo
}

func NewBlockInfo(bg abci.RequestBeginBlock) *BlockInfo {
	return &BlockInfo{
		Height:            bg.Header.Height,
		Evidence:          FromABCIEvidence(bg.ByzantineValidators),
		HeaderHash:        bg.Hash,
		ValidatorsHash:    bg.Header.NextValidatorsHash,
		ProposerAddress:   bg.Header.ProposerAddress,
		DecidedLastCommit: FromABCICommitInfo(bg.LastCommitInfo),
	}
}

func (bis *BlockInfo) GetHeight() int64 {
	return bis.Height
}

func (bis *BlockInfo) Misbehavior() []blockinfo.Misbehavior {
	return bis.Evidence
}

func (bis *BlockInfo) GetHeaderHash() []byte {
	return bis.HeaderHash
}

func (bis *BlockInfo) GetValidatorsHash() []byte {
	return bis.ValidatorsHash
}

func (bis *BlockInfo) GetProposerAddress() []byte {
	return bis.ProposerAddress
}

func (bis *BlockInfo) GetDecidedLastCommit() blockinfo.CommitInfo {
	return bis.DecidedLastCommit
}

// Utils

// FromABCIEvidence converts a CometBFT concrete Evidence type to
// SDK Evidence.
func FromABCIEvidence(e []abci.Misbehavior) []blockinfo.Misbehavior {
	misbehavior := make([]blockinfo.Misbehavior, len(e))

	for i, ev := range e {
		misbehavior[i] = blockinfo.Misbehavior{
			Type:   blockinfo.MisbehaviorType(ev.Type),
			Height: ev.Height,
			Validator: blockinfo.Validator{
				Address: ev.Validator.Address,
				Power:   ev.Validator.Power,
			},
			Time:             ev.Time,
			TotalVotingPower: ev.TotalVotingPower,
		}
	}

	return misbehavior
}

func FromABCICommitInfo(ci abci.CommitInfo) blockinfo.CommitInfo {
	votes := make([]*blockinfo.VoteInfo, len(ci.Votes))
	for i, v := range ci.Votes {
		votes[i] = &blockinfo.VoteInfo{
			Validator: blockinfo.Validator{
				Address: v.Validator.Address,
				Power:   v.Validator.Power,
			},
			SignedLastBlock: v.SignedLastBlock,
		}
	}

	return blockinfo.CommitInfo{
		Round: ci.Round,
		Votes: votes,
	}
}
