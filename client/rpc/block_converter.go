package rpc

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"
	"github.com/tendermint/tendermint/version"
)

// ResultBlock represents a single block with metadata
type ResultBlock struct {
	BlockMeta BlockMeta `json:"block_meta"`
	Block     Block     `json:"block"`
}

// BlockMeta contains meta information about a block - namely, it's ID and Header.
type BlockMeta struct {
	BlockID types.BlockID `json:"block_id"` // the block hash and partsethash
	Header  Header        `json:"header"`   // The block's Header
}

// Header defines a wrapper around Tendermint's Header type overriding various fields.
// nolint: structtag
type Header struct {
	// embed original type
	types.Header

	// override fields so json.Marshal will marshal in accordance with amino JSON format
	Version         Consensus      `json:"version"`
	Height          int64          `json:"height,string"`
	NumTxs          int64          `json:"num_txs,string"`
	TotalTxs        int64          `json:"total_txs,string"`
	LastBlockID     BlockID        `json:"last_block_id"`
	ProposerAddress sdk.ValAddress `json:"proposer_address"`
}

// MarshalJSON implements the json.Marshaler interface. We do this because Amino
// does not respect the JSON stdlib embedding semantics.
func (h Header) MarshalJSON() ([]byte, error) {
	type headerJSON Header
	return json.Marshal(headerJSON(h))
}

// BlockID defines a wrapper around Tendermint's BlockID type overriding various fields.
// nolint: structtag
type BlockID struct {
	// embed original type
	types.BlockID

	// override fields from original type
	PartsHeader PartSetHeader `json:"parts"`
}

// MarshalJSON implements the json.Marshaler interface. We do this because Amino
// does not respect the JSON stdlib embedding semantics.
func (b BlockID) MarshalJSON() ([]byte, error) {
	type blockIDJSON BlockID
	return json.Marshal(blockIDJSON(b))
}

// PartSetHeader defines a wrapper around Tendermint's PartSetHeader type overriding various fields.
// nolint: structtag
type PartSetHeader struct {
	// embed original type
	types.PartSetHeader

	// override fields from original type
	Total int `json:"total,string"`
}

// MarshalJSON implements the json.Marshaler interface. We do this because Amino
// does not respect the JSON stdlib embedding semantics.
func (b PartSetHeader) MarshalJSON() ([]byte, error) {
	type partSetHeadJSON PartSetHeader
	return json.Marshal(partSetHeadJSON(b))
}

// Consensus defines a wrapper around Tendermint's Consensus type overriding various fields.
// nolint: structtag
type Consensus struct {
	// embed original type
	version.Consensus

	// override fields so json.Marshal will marshal in accordance with amino JSON format
	App   uint64 `json:"app,string"`
	Block uint64 `json:"block,string"`
}

// MarshalJSON implements the json.Marshaler interface. We do this because Amino
// does not respect the JSON stdlib embedding semantics.
func (b Consensus) MarshalJSON() ([]byte, error) {
	type consensusJSON Consensus
	return json.Marshal(consensusJSON(b))
}

// Block defines the atomic unit of a Tendermint blockchain.
// nolint: structtag.
type Block struct {
	// embed original type
	types.Block

	// override fields
	Header     Header `json:"header"`
	LastCommit Commit `json:"last_commit"`
}

// MarshalJSON implements the json.Marshaler interface. We do this because Amino
// does not respect the JSON stdlib embedding semantics.
func (b Block) MarshalJSON() ([]byte, error) {
	type blockJSON Block
	return json.Marshal(blockJSON(b))
}

// Commit defines a wrapper around Tendermint's Commit type overriding various fields.
// nolint: structtag.
type Commit struct {
	// embed original type
	*types.Commit

	// override fields
	BlockID    BlockID     `json:"block_id"`
	Precommits []CommitSig `json:"precommits"`
}

// CommitSig defines a wrapper around Tendermint's CommitSig type overriding various fields.
// nolint: structtag
type CommitSig struct {
	// embed original type
	types.CommitSig

	// override fields so json.Marshal will marshal in accordance with amino JSON format
	Height           int64          `json:"height,string"`
	Round            int            `json:"round,string"`
	BlockID          BlockID        `json:"block_id"`
	ValidatorAddress sdk.ValAddress `json:"validator_address"`
	ValidatorIndex   int            `json:"validator_index,string"`
}

// MarshalJSON implements the json.Marshaler interface. We do this because Amino
// does not respect the JSON stdlib embedding semantics.
func (c CommitSig) MarshalJSON() ([]byte, error) {
	type commitSigJSON CommitSig
	return json.Marshal(commitSigJSON(c))
}

// ConvertBlockResult allows to convert the given standard ResultBlock into a new ResultBlock having all the
// validator addresses as Bech32 strings instead of HEX ones.
func ConvertBlockResult(res *ctypes.ResultBlock) (blockResult *ResultBlock) {

	if res == nil {
		return nil
	}

	header := Header{
		Header: res.BlockMeta.Header,

		Version: Consensus{
			Consensus: res.BlockMeta.Header.Version,
			App:       res.BlockMeta.Header.Version.App.Uint64(),
			Block:     res.BlockMeta.Header.Version.Block.Uint64(),
		},
		Height:   res.BlockMeta.Header.Height,
		NumTxs:   res.BlockMeta.Header.NumTxs,
		TotalTxs: res.BlockMeta.Header.TotalTxs,
		LastBlockID: BlockID{
			BlockID: res.BlockMeta.Header.LastBlockID,
			PartsHeader: PartSetHeader{
				PartSetHeader: res.BlockMeta.Header.LastBlockID.PartsHeader,
				Total:         res.BlockMeta.Header.LastBlockID.PartsHeader.Total,
			},
		},

		ProposerAddress: sdk.ValAddress(res.BlockMeta.Header.ProposerAddress),
	}

	return &ResultBlock{
		BlockMeta: BlockMeta{
			BlockID: res.BlockMeta.BlockID,
			Header:  header,
		},
		Block: Block{
			Header: header,
			LastCommit: Commit{
				Commit: res.Block.LastCommit,
				BlockID: BlockID{
					BlockID: res.Block.LastCommit.BlockID,
					PartsHeader: PartSetHeader{
						PartSetHeader: res.Block.LastCommit.BlockID.PartsHeader,
						Total:         res.Block.LastCommit.BlockID.PartsHeader.Total,
					},
				},
				Precommits: convertPreCommits(res.Block.LastCommit.Precommits),
			},
		},
	}
}

func convertPreCommits(preCommits []*types.CommitSig) (sigs []CommitSig) {
	for _, commit := range preCommits {
		sig := CommitSig{
			CommitSig: *commit,
			Height:    commit.Height,
			Round:     commit.Round,
			BlockID: BlockID{
				BlockID: commit.BlockID,
				PartsHeader: PartSetHeader{
					PartSetHeader: commit.BlockID.PartsHeader,
					Total:         commit.BlockID.PartsHeader.Total,
				},
			},
			ValidatorAddress: sdk.ValAddress(commit.ValidatorAddress),
			ValidatorIndex:   commit.ValidatorIndex,
		}

		sigs = append(sigs, sig)
	}

	return sigs
}
