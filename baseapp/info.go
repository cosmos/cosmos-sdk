package baseapp

import (
	"time"

	"cosmossdk.io/core/comet"
	"cosmossdk.io/core/header"

	abci "github.com/cometbft/cometbft/abci/types"
)

var _ header.Info = (*HeaderInfo)(nil)

// BlockInfo implements the BlockInfo interface for that all consensus engines need to provide
type HeaderInfo struct {
	Height     int64
	HeaderHash []byte
	Time       time.Time
	ChainID    string
}

func NewHeaderInfo(height int64, hash []byte, t time.Time, chainID string) *HeaderInfo {
	return &HeaderInfo{
		Height:     height,
		HeaderHash: hash,
		Time:       t,
		ChainID:    chainID,
	}
}

func (bis *HeaderInfo) GetHeight() int64 {
	return bis.Height
}

func (bis *HeaderInfo) GetTime() time.Time {
	return bis.Time
}

func (bis *HeaderInfo) GetChainID() string {
	return bis.ChainID
}

func (bis *HeaderInfo) GetHeaderHash() []byte {
	return bis.HeaderHash
}

var _ comet.Info = (*CometInfo)(nil)

// CometInfo implements the CometInfo interface
// This is specific to the comet consensus engine
type CometInfo struct {
	Evidence          []comet.Misbehavior
	ValidatorsHash    []byte
	ProposerAddress   []byte
	DecidedLastCommit comet.CommitInfo
}

func NewCometInfo(bg abci.RequestBeginBlock) comet.Info {
	return &CometInfo{
		Evidence:          FromABCIEvidence(bg.ByzantineValidators),
		ValidatorsHash:    bg.Hash,
		ProposerAddress:   bg.Header.ProposerAddress,
		DecidedLastCommit: FromABCICommitInfo(bg.LastCommitInfo),
	}
}

func (cis *CometInfo) GetMisbehavior() []comet.Misbehavior {
	return cis.Evidence
}

func (cis *CometInfo) GetValidatorsHash() []byte {
	return cis.ValidatorsHash
}

func (cis *CometInfo) GetProposerAddress() []byte {
	return cis.ProposerAddress
}

func (cis *CometInfo) GetDecidedLastCommit() comet.CommitInfo {
	return cis.DecidedLastCommit
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
