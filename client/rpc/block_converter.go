package rpc

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"
)

// Single block (with meta)
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

// Block defines the atomic unit of a Tendermint blockchain.
type Block struct {
	Header     `json:"header"`
	types.Data `json:"data"`
	Evidence   types.EvidenceData `json:"evidence"`
	LastCommit Commit             `json:"last_commit"`
}

// Commit contains the evidence that a block was committed by a set of validators.
// NOTE: Commit is empty for height 1, but never nil.
type Commit struct {
	BlockID    types.BlockID `json:"block_id"`
	Precommits []CommitSig   `json:"precommits"`
}

// CommitSig defines a wrapper around Tendermint's CommitSig type overriding various fields.
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

// ConvertBlockResult allows to convert the given standard ResultBlock into a new ResultBlock having all the
// validator addresses as Bech32 strings instead of HEX ones.
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
