package rpc

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"
)

// ResultBlock defines a wrapper around the standard ResultBlock type overriding various fields.
// nolint: structtag
type ResultBlock struct {
	ctypes.ResultBlock
	BlockMeta BlockMeta `json:"block_meta"`
	Block     Block     `json:"block"`
}

// MarshalJSON implements the json.Marshaler interface. We do this because Amino
// does not respect the JSON stdlib embedding semantics.
func (r ResultBlock) MarshalJSON() ([]byte, error) {
	type resultBlockJSON ResultBlock
	_h := resultBlockJSON(r)

	return json.Marshal(_h)
}

// BlockMeta defines a wrapper around Tendermint's BlockMeta type overriding various fields.
// nolint: structtag
type BlockMeta struct {
	types.BlockMeta
	Header Header `json:"header"` // The block's Header
}

// MarshalJSON implements the json.Marshaler interface. We do this because Amino
// does not respect the JSON stdlib embedding semantics.
func (b BlockMeta) MarshalJSON() ([]byte, error) {
	type blockMetaJSON BlockMeta
	_h := blockMetaJSON(b)

	return json.Marshal(_h)
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

// Block defines a wrapper around Tendermint's Block type overriding various fields.
// nolint: structtag
type Block struct {
	types.Block
	Header     `json:"header"`
	LastCommit Commit `json:"last_commit"`
}

// MarshalJSON implements the json.Marshaler interface. We do this because Amino
// does not respect the JSON stdlib embedding semantics.
func (b Block) MarshalJSON() ([]byte, error) {
	type blockJSON Block
	_h := blockJSON(b)

	return json.Marshal(_h)
}

// Commit defines a wrapper around Tendermint's Commit type overriding various fields.
// nolint: structtag
type Commit struct {
	types.Commit
	Precommits []CommitSig `json:"precommits"`
}

// MarshalJSON implements the json.Marshaler interface. We do this because Amino
// does not respect the JSON stdlib embedding semantics.
func (c Commit) MarshalJSON() ([]byte, error) {
	type commitJSON Commit
	_h := commitJSON(c)

	return json.Marshal(_h)
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
	type commitSigJSON CommitSig
	_h := commitSigJSON(c)

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
		ResultBlock: *res,
		BlockMeta: BlockMeta{
			BlockMeta: *res.BlockMeta,
			Header:    header,
		},
		Block: Block{
			Block:  *res.Block,
			Header: header,
			LastCommit: Commit{
				Commit:     *res.Block.LastCommit,
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
