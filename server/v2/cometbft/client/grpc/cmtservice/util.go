package cmtservice

import (
	"cosmossdk.io/core/address"
	cmtprototypes "github.com/cometbft/cometbft/proto/tendermint/types"
)

// convertHeader converts CometBFT header to sdk header
func convertHeader(h cmtprototypes.Header, consAddrCdc address.Codec) Header {

	propAddress, err := consAddrCdc.BytesToString(h.ProposerAddress)
	if err != nil {
		panic(err)
	}

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
		ProposerAddress:    propAddress,
	}
}

// convertBlock converts CometBFT block to sdk block
func convertBlock(cmtblock *cmtprototypes.Block, consAddrCdc address.Codec) *Block {
	b := new(Block)

	b.Header = convertHeader(cmtblock.Header, consAddrCdc)
	b.LastCommit = cmtblock.LastCommit
	b.Data = cmtblock.Data
	b.Evidence = cmtblock.Evidence

	return b
}
