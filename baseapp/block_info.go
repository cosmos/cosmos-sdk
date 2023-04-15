package baseapp

import abci "github.com/cometbft/cometbft/abci/types"

// ======================================================
//  BlockInfoService
// ======================================================

// BlockInfoService is the service that runtime will provide to modules which need Comet block information.
type BlockInfoService interface {
	GetHeight() int64                // GetHeight returns the height of the block
	Misbehavior() []abci.Misbehavior // Misbehavior returns the misbehavior of the block
	GetHeaderHash() []byte           // GetHeaderHash returns the hash of the block header
	// GetValidatorsHash returns the hash of the validators
	// For Comet, it is the hash of the next validator set
	GetValidatorsHash() []byte
	GetProposerAddress() []byte            // GetProposerAddress returns the address of the block proposer
	GetDecidedLastCommit() abci.CommitInfo // GetDecidedLastCommit returns the last commit info
}

var _ BlockInfoService = (*BlockInfo)(nil)

type BlockInfo struct {
	Height            int64
	Evidence          []abci.Misbehavior
	HeaderHash        []byte
	ValidatorsHash    []byte
	ProposerAddress   []byte
	DecidedLastCommit abci.CommitInfo
}

func NewBlockInfo(bg abci.RequestBeginBlock) *BlockInfo {
	return &BlockInfo{
		Height:            bg.Header.Height,
		Evidence:          bg.ByzantineValidators,
		HeaderHash:        bg.Hash,
		ValidatorsHash:    bg.Header.NextValidatorsHash,
		ProposerAddress:   bg.Header.ProposerAddress,
		DecidedLastCommit: bg.LastCommitInfo,
	}
}

func (bis *BlockInfo) GetHeight() int64 {
	return bis.Height
}

func (bis *BlockInfo) Misbehavior() []abci.Misbehavior {
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

func (bis *BlockInfo) GetDecidedLastCommit() abci.CommitInfo {
	return bis.DecidedLastCommit
}
