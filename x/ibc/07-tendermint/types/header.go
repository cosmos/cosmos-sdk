package types

import (
	"bytes"
	"time"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
)

var _ clientexported.Header = (*Header)(nil)

// ClientType defines that the Header is a Tendermint consensus algorithm
func (h Header) ClientType() clientexported.ClientType {
	return clientexported.Tendermint
}

// ConsensusState returns the consensus state associated with the header
func (h Header) ConsensusState() ConsensusState {
	return ConsensusState{
		Height:       h.GetHeight(),
		Timestamp:    h.GetTime(),
		Root:         commitmenttypes.NewMerkleRoot(h.SignedHeader.Header.AppHash),
		ValidatorSet: h.ValidatorSet,
	}
}

// GetHeight returns the current height
//
// NOTE: also referred as `sequence`
func (h Header) GetHeight() uint64 {
	return uint64(h.SignedHeader.Header.GetHeight())
}

// GetTime returns the signed header's time
func (h Header) GetTime() time.Time {
	return h.SignedHeader.Header.GetTime()
}

// ValidateBasic calls the SignedHeader ValidateBasic function
// and checks that validatorsets are not nil
func (h Header) ValidateBasic(chainID string) error {
	if err := h.SignedHeader.Header.ValidateBasic(chainID); err != nil {
		return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, err.Error())
	}
	if h.ValidatorSet == nil {
		return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "validator set is nil")
	}
	if !bytes.Equal(h.SignedHeader.Header.ValidatorsHash, h.ValidatorSet.Hash()) {
		return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "validator set does not match hash")
	}
	return nil
}
