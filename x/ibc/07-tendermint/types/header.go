package types

import (
	"bytes"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
)

var _ clientexported.Header = Header{}

// Header defines the Tendermint consensus Header
type Header struct {
	tmtypes.SignedHeader `json:"signed_header" yaml:"signed_header"` // contains the commitment root
	ValidatorSet         *tmtypes.ValidatorSet                       `json:"validator_set" yaml:"validator_set"`
}

// ClientType defines that the Header is a Tendermint consensus algorithm
func (h Header) ClientType() clientexported.ClientType {
	return clientexported.Tendermint
}

// ConsensusState returns the consensus state associated with the header
func (h Header) ConsensusState() ConsensusState {
	return ConsensusState{
		Height:       uint64(h.Height),
		Timestamp:    h.Time,
		Root:         commitmenttypes.NewMerkleRoot(h.AppHash),
		ValidatorSet: h.ValidatorSet,
	}
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
		return sdkerrors.Wrap(err, "header failed basic validation")
	}
	if h.ValidatorSet == nil {
		return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "validator set is nil")
	}
	if !bytes.Equal(h.ValidatorsHash, h.ValidatorSet.Hash()) {
		return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "validator set does not match hash")
	}
	return nil
}

// ToABCIHeader parses the header to an ABCI header type.
// NOTE: only for testing use.
func (h Header) ToABCIHeader() abci.Header {
	return tmtypes.TM2PB.Header(h.SignedHeader.Header)
}
