package tmservice

import (
	tmprototypes "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// convertHeader converts tendermint header to sdk header
func convertHeader(h tmprototypes.Header) Header {
	return Header{
		Version:            h.Version,
		ChainID:            h.ChainID,
		Height:             h.Height,
		Time:               h.Time,
		LastBlockId:        h.LastBlockId,
		ValidatorsHash:     h.ValidatorsHash,
		NextValidatorsHash: h.NextValidatorsHash,
		ConsensusHash:      h.ConsensusHash,
		AppHash:            h.AppHash,
		DataHash:           h.DataHash,
		EvidenceHash:       h.EvidenceHash,
		LastResultsHash:    h.LastResultsHash,
		LastCommitHash:     h.LastCommitHash,
		ProposerAddress:    sdk.ConsAddress(h.ProposerAddress).String(),
	}
}

<<<<<<< HEAD:client/grpc/tmservice/util.go
// convertBlock converts tendermint block to sdk block
func convertBlock(tmblock *tmprototypes.Block) *Block {
=======
// convertBlock converts CometBFT block to sdk block
func convertBlock(cmtblock *cmtprototypes.Block) *Block {
>>>>>>> 07dc5e70e (fix: Change proposer address cast for `sdk_block` conversion (#15243)):client/grpc/cmtservice/util.go
	b := new(Block)

	b.Header = convertHeader(cmtblock.Header)
	b.LastCommit = cmtblock.LastCommit
	b.Data = cmtblock.Data
	b.Evidence = cmtblock.Evidence

	return b
}
