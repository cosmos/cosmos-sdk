package rpc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"
	"github.com/tendermint/tendermint/version"
	"sync"
	"time"
)

type CosmosResultBlock struct {
	BlockMeta CosmosBlockMeta `json:"block_meta"`
	Block     CosmosBlock     `json:"block"`
}

type CosmosBlockMeta struct {
	BlockID types.BlockID `json:"block_id"` // the block hash and partsethash
	Header  CosmosHeader  `json:"header"`   // The block's Header
}

type CosmosHeader struct {
	// basic block info
	Version  version.Consensus `json:"version"`
	ChainID  string            `json:"chain_id"`
	Height   int64             `json:"height"`
	Time     time.Time         `json:"time"`
	NumTxs   int64             `json:"num_txs"`
	TotalTxs int64             `json:"total_txs"`

	// prev block info
	LastBlockID types.BlockID `json:"last_block_id"`

	// hashes of block data
	LastCommitHash cmn.HexBytes `json:"last_commit_hash"` // commit from validators from the last block
	DataHash       cmn.HexBytes `json:"data_hash"`        // transactions

	// hashes from the app output from the prev block
	ValidatorsHash     cmn.HexBytes `json:"validators_hash"`      // validators for the current block
	NextValidatorsHash cmn.HexBytes `json:"next_validators_hash"` // validators for the next block
	ConsensusHash      cmn.HexBytes `json:"consensus_hash"`       // consensus params for current block
	AppHash            cmn.HexBytes `json:"app_hash"`             // state after txs from the previous block
	LastResultsHash    cmn.HexBytes `json:"last_results_hash"`    // root hash of all results from the txs from the previous block

	// consensus info
	EvidenceHash    cmn.HexBytes   `json:"evidence_hash"`    // evidence included in the block
	ProposerAddress sdk.ValAddress `json:"proposer_address"` // original proposer of the block
}

type CosmosBlock struct {
	mtx          sync.Mutex
	CosmosHeader `json:"header"`
	types.Data   `json:"data"`
	Evidence     types.EvidenceData `json:"evidence"`
	LastCommit   CosmosCommit       `json:"last_commit"`
}

type CosmosCommit struct {
	BlockID    types.BlockID     `json:"block_id"`
	Precommits []CosmosCommitSig `json:"precommits"`
}

type CosmosCommitSig struct {
	Type             types.SignedMsgType `json:"type"`
	Height           int64               `json:"height"`
	Round            int                 `json:"round"`
	BlockID          types.BlockID       `json:"block_id"` // zero if vote is nil.
	Timestamp        time.Time           `json:"timestamp"`
	ValidatorAddress sdk.ValAddress      `json:"validator_address"`
	ValidatorIndex   int                 `json:"validator_index"`
	Signature        []byte              `json:"signature"`
}

func ConvertBlockResult(res *ctypes.ResultBlock) (blockResult CosmosResultBlock) {

	// header
	header := CosmosHeader{
		Version:  res.BlockMeta.Header.Version,
		ChainID:  res.BlockMeta.Header.ChainID,
		Height:   res.BlockMeta.Header.Height,
		Time:     res.BlockMeta.Header.Time,
		NumTxs:   res.BlockMeta.Header.NumTxs,
		TotalTxs: res.BlockMeta.Header.TotalTxs,

		LastBlockID: res.BlockMeta.Header.LastBlockID,

		LastCommitHash: res.BlockMeta.Header.LastCommitHash,
		DataHash:       res.BlockMeta.Header.DataHash,

		ValidatorsHash:     res.BlockMeta.Header.ValidatorsHash,
		NextValidatorsHash: res.BlockMeta.Header.NextValidatorsHash,
		ConsensusHash:      res.BlockMeta.Header.ConsensusHash,
		AppHash:            res.BlockMeta.Header.AppHash,
		LastResultsHash:    res.BlockMeta.Header.LastResultsHash,

		EvidenceHash:    res.BlockMeta.Header.EvidenceHash,
		ProposerAddress: sdk.ValAddress(res.BlockMeta.Header.ProposerAddress),
	}

	// meta
	blockMeta := CosmosBlockMeta{
		BlockID: res.BlockMeta.BlockID,
		Header:  header,
	}

	// commit
	commit := CosmosCommit{
		BlockID:    res.Block.LastCommit.BlockID,
		Precommits: convertPreCommits(res.Block.LastCommit.Precommits),
	}

	// block
	block := CosmosBlock{
		CosmosHeader: header,
		Data:         res.Block.Data,
		Evidence:     res.Block.Evidence,
		LastCommit:   commit,
	}

	// blockResult
	blockResult = CosmosResultBlock{
		BlockMeta: blockMeta,
		Block:     block,
	}

	return blockResult
}

func convertPreCommits(preCommits []*types.CommitSig) (sigs []CosmosCommitSig) {
	for _, commit := range preCommits {
		sig := CosmosCommitSig{
			Type:             commit.Type,
			Height:           commit.Height,
			Round:            commit.Round,
			BlockID:          commit.BlockID,
			Timestamp:        commit.Timestamp,
			ValidatorAddress: sdk.ValAddress(commit.ValidatorAddress),
			ValidatorIndex:   commit.ValidatorIndex,
			Signature:        commit.Signature,
		}

		sigs = append(sigs, sig)
	}

	return sigs
}
