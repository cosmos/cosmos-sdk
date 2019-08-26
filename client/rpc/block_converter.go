package rpc

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"
)

type ResultBlock struct {
	BlockMeta BlockMeta `json:"block_meta"`
	Block     Block     `json:"block"`
}

type BlockMeta struct {
	BlockID types.BlockID `json:"block_id"` // the block hash and partsethash
	Header  Header        `json:"header"`   // The block's Header
}

// Header defines a wrapper around Tendermint's Header type overriding various fields.
// nolint: structtag
type Header struct {
	// embed original type
	types.Header

	// override fields from original type
	ProposerAddress sdk.ValAddress `json:"proposer_address"`
}

// MarshalJSON implements the json.Marshaler interface. We do this because Amino
// does not respect the JSON stdlib embedding semantics.
func (h Header) MarshalJSON() ([]byte, error) {
	type headerJSON Header
	_h := headerJSON(h)

	return json.Marshal(_h)
}

type Block struct {
	Header     `json:"header"`
	types.Data `json:"data"`
	Evidence   types.EvidenceData `json:"evidence"`
	LastCommit Commit             `json:"last_commit"`
}

type Commit struct {
	BlockID    types.BlockID `json:"block_id"`
	Precommits []CommitSig   `json:"precommits"`
}

// nolint: structtag
type CommitSig struct {
	types.CommitSig
	ValidatorAddress sdk.ValAddress `json:"validator_address"`
}

// MarshalJSON implements the json.Marshaler interface. We do this because Amino
// does not respect the JSON stdlib embedding semantics.
func (c CommitSig) MarshalJSON() ([]byte, error) {
	type headerJSON CommitSig
	_h := headerJSON(c)

	return json.Marshal(_h)
}

func ConvertBlockResult(res *ctypes.ResultBlock) (blockResult ResultBlock) {

	header := Header{
		Header:          res.BlockMeta.Header,
		ProposerAddress: sdk.ValAddress(res.BlockMeta.Header.ProposerAddress),
	}

	return ResultBlock{
		BlockMeta: BlockMeta{
			BlockID: res.BlockMeta.BlockID,
			Header:  header,
		},
		Block: Block{
			Header:   header,
			Data:     res.Block.Data,
			Evidence: res.Block.Evidence,
			LastCommit: Commit{
				BlockID:    res.Block.LastCommit.BlockID,
				Precommits: convertPreCommits(res.Block.LastCommit.Precommits),
			},
		},
	}
}

func convertPreCommits(preCommits []*types.CommitSig) (sigs []CommitSig) {
	for _, commit := range preCommits {
		sig := CommitSig{
			CommitSig:        *commit,
			ValidatorAddress: sdk.ValAddress(commit.ValidatorAddress),
		}

		sigs = append(sigs, sig)
	}

	return sigs
}
