package baseapp

import (
	"time"

	coreinfo "cosmossdk.io/core/info"
	abci "github.com/cometbft/cometbft/abci/types"
)

var _ coreinfo.BlockInfo = (*BlockInfo)(nil)

// BlockInfo implements the BlockInfo interface for that all consensus engines need to provide
type BlockInfo struct {
	Height     int64
	HeaderHash []byte
	Time       time.Time
	ChainID    string
}

func NewBlockInfo(height int64, hash []byte, t time.Time, chainID string) *BlockInfo {
	return &BlockInfo{
		Height:     height,
		HeaderHash: hash,
		Time:       t,
		ChainID:    chainID,
	}
}

func (bis *BlockInfo) GetHeight() int64 {
	return bis.Height
}

func (bis *BlockInfo) GetTime() time.Time {
	return bis.Time
}

func (bis *BlockInfo) GetChainID() string {
	return bis.ChainID
}

func (bis *BlockInfo) GetHeaderHash() []byte {
	return bis.HeaderHash
}

var _ coreinfo.CometInfo = (*CometInfo)(nil)

// CometInfo implements the CometInfo interface
// This is specific to the comet consensus engine
type CometInfo struct {
	Evidence          []coreinfo.Misbehavior
	ValidatorsHash    []byte
	ProposerAddress   []byte
	DecidedLastCommit coreinfo.CommitInfo
}

func NewCometInfo(bg abci.RequestBeginBlock) coreinfo.CometInfo {
	return &CometInfo{
		Evidence:          FromABCIEvidence(bg.ByzantineValidators),
		ValidatorsHash:    bg.Hash,
		ProposerAddress:   bg.Header.ProposerAddress,
		DecidedLastCommit: FromABCICommitInfo(bg.LastCommitInfo),
	}
}

func (cis *CometInfo) GetMisbehavior() []coreinfo.Misbehavior {
	return cis.Evidence
}

func (cis *CometInfo) GetValidatorsHash() []byte {
	return cis.ValidatorsHash
}

func (cis *CometInfo) GetProposerAddress() []byte {
	return cis.ProposerAddress
}

func (cis *CometInfo) GetDecidedLastCommit() coreinfo.CommitInfo {
	return cis.DecidedLastCommit
}

// Utils

// FromABCIEvidence converts a CometBFT concrete Evidence type to
// SDK Evidence.
func FromABCIEvidence(e []abci.Misbehavior) []coreinfo.Misbehavior {
	misbehavior := make([]coreinfo.Misbehavior, len(e))

	for i, ev := range e {
		misbehavior[i] = coreinfo.Misbehavior{
			Type:   coreinfo.MisbehaviorType(ev.Type),
			Height: ev.Height,
			Validator: coreinfo.Validator{
				Address: ev.Validator.Address,
				Power:   ev.Validator.Power,
			},
			Time:             ev.Time,
			TotalVotingPower: ev.TotalVotingPower,
		}
	}

	return misbehavior
}

func FromABCICommitInfo(ci abci.CommitInfo) coreinfo.CommitInfo {
	votes := make([]*coreinfo.VoteInfo, len(ci.Votes))
	for i, v := range ci.Votes {
		votes[i] = &coreinfo.VoteInfo{
			Validator: coreinfo.Validator{
				Address: v.Validator.Address,
				Power:   v.Validator.Power,
			},
			SignedLastBlock: v.SignedLastBlock,
		}
	}

	return coreinfo.CommitInfo{
		Round: ci.Round,
		Votes: votes,
	}
}
