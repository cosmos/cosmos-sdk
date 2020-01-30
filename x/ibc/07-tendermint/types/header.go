package types

import (
	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
)

var _ clientexported.Header = Header{}

// Header defines the Tendermint consensus Header
type Header struct {
	tmtypes.SignedHeader                       // contains the commitment root
	ValidatorSet         *tmtypes.ValidatorSet `json:"validator_set" yaml:"validator_set"`
}

// ClientType defines that the Header is a Tendermint consensus algorithm
func (h Header) ClientType() clientexported.ClientType {
	return clientexported.Tendermint
}

// GetHeight returns the current height
//
// NOTE: also referred as `sequence`
func (h Header) GetHeight() uint64 {
	return uint64(h.Height)
}

// ValidateBasic calls the SignedHeader ValidateBasic function
// and checks that validatorsets are not nil
func (h Header) ValidateBasic(chainID string) error {
	if err := h.SignedHeader.ValidateBasic(chainID); err != nil {
		return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, err.Error())
	}
	if h.ValidatorSet == nil {
		return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "validator set is nil")
	}
	return nil
}

// ToABCIHeader parses the header to an ABCI header type.
// NOTE: only for testing use.
func (h Header) ToABCIHeader() abci.Header {
	return abci.Header{
		Version: abci.Version{
			App:   h.Version.App.Uint64(),
			Block: h.Version.Block.Uint64(),
		},
		ChainID: h.ChainID,
		Height:  h.Height,
		Time:    h.Time,
		LastBlockId: abci.BlockID{
			Hash: h.LastBlockID.Hash,
			PartsHeader: abci.PartSetHeader{
				Total: int32(h.LastBlockID.PartsHeader.Total),
				Hash:  h.LastBlockID.PartsHeader.Hash,
			},
		},
		LastCommitHash:     h.LastCommitHash,
		DataHash:           h.DataHash,
		ValidatorsHash:     h.ValidatorsHash,
		NextValidatorsHash: h.NextValidatorsHash,
		ConsensusHash:      h.ConsensusHash,
		AppHash:            h.AppHash,
		LastResultsHash:    h.LastResultsHash,
		EvidenceHash:       h.EvidenceHash,
		ProposerAddress:    h.ProposerAddress,
	}
}
